package engine

import (
	"encoding/json"
	"math"
	"testing"
)

func TestRunSummaryResetsAtRespawnWhileRecordsRemain(t *testing.T) {
	w := New(17, Config{Size: 1600, Mode: Classic})
	id := w.AddPlayer("Runner", false, "")
	p := w.Player(id)
	if !p.HasRun || p.RunBorn != 0 || p.RunPeak != SpawnMass || p.Peak != SpawnMass {
		t.Fatal("spawn summary", p)
	}
	for i := range w.Entities {
		if w.Entities[i].Owner == id {
			w.Entities[i].Mass = 100
			w.Sources += 64
		}
	}
	w.Step()
	p = w.Player(id)
	if p.RunPeak < 99 || p.Peak < 99 {
		t.Fatal("growth record missing", p)
	}
	for i := range w.Entities {
		if w.Entities[i].Owner == id {
			w.Sinks += w.Entities[i].Mass
			w.Entities[i].Mass = 0
		}
	}
	w.Step()
	if p.Alive {
		t.Fatal("missing death")
	}
	peak := p.RunPeak
	for range 61 {
		w.Step()
	}
	if p.RunPeak != peak {
		t.Fatal("death record changed")
	}
	if err := w.Respawn(id); err != nil {
		t.Fatal(err)
	}
	p = w.Player(id)
	if p.RunBorn != w.Tick || p.RunPeak != SpawnMass || p.Peak < 99 || p.RunKills != 0 {
		t.Fatal("run reset corrupted record", p)
	}
	if err := w.Valid(); err != nil {
		t.Fatal(err)
	}
}

func empty() *World { return New(7, Config{Size: 1600, Mode: Classic}) }
func place(w *World, id uint32, x, y, mass float64) {
	for i := range w.Entities {
		e := &w.Entities[i]
		if e.Owner == id && e.Kind == Cell {
			w.Sources += mass - e.Mass
			e.Mass = mass
			e.X = x
			e.Y = y
			e.VX = 0
			e.VY = 0
		}
	}
}
func fixture(w *World, kind Kind, x, y, mass float64) {
	w.NextID++
	w.Entities = append(w.Entities, Entity{ID: w.NextID, Kind: kind, X: x, Y: y, Mass: mass})
	w.Sources += mass
}
func invariant(t *testing.T, w *World) {
	t.Helper()
	if err := w.Valid(); err != nil {
		t.Fatal(err)
	}
	if math.Abs(w.TotalMass()+w.Sinks-w.Sources) > 1e-5 {
		t.Fatalf("mass: total=%f sinks=%f sources=%f", w.TotalMass(), w.Sinks, w.Sources)
	}
}

func TestMovementMassAndBounds(t *testing.T) {
	w := empty()
	small := w.AddPlayer("Small", false, "")
	big := w.AddPlayer("Big", false, "")
	place(w, small, 200, 300, 36)
	place(w, big, 200, 1200, 360)
	_ = w.Input(small, 1500, 300, 0)
	_ = w.Input(big, 1500, 1200, 0)
	for range 30 {
		w.Step()
	}
	if w.Cells(small)[0].X-200 <= w.Cells(big)[0].X-200 {
		t.Fatal("larger mass must move slower")
	}
	for range 600 {
		w.Step()
	}
	invariant(t, w)
	for _, e := range w.Entities {
		if e.X < Radius(e.Mass) || e.X > w.Config.Size-Radius(e.Mass) {
			t.Fatal("escaped map")
		}
	}
}
func TestMovementSlowsNearCursorWithoutChangingFarSpeed(t *testing.T) {
	w := empty()
	a := w.AddPlayer("Near", false, "")
	b := w.AddPlayer("Far", false, "")
	place(w, a, 300, 300, 36)
	place(w, b, 300, 900, 36)
	_ = w.Input(a, 310, 300, 0)
	_ = w.Input(b, 1300, 900, 0)
	w.Step()
	near, far := w.Cells(a)[0].X-300, w.Cells(b)[0].X-300
	if near >= far*.3 || math.Abs(far-220.0/30) > .001 {
		t.Fatal("cursor falloff", near, far)
	}
	for range 120 {
		w.Step()
	}
	if w.Cells(a)[0].X > 310.001 || w.Cells(a)[0].X < 309.5 {
		t.Fatal("cursor overshoot or failed convergence", w.Cells(a))
	}
	invariant(t, w)
}
func TestExpeditionMutationsAffectRealMovementFoodAndMerge(t *testing.T) {
	for mutation := 0; mutation < 3; mutation++ {
		w := empty()
		w.Config.Mode = Rogue
		id := w.AddPlayer("Explorer", false, "")
		place(w, id, 800, 800, 36)
		w.Player(id).Mutations[mutation] = 1
		switch mutation {
		case 0:
			_ = w.Input(id, 1500, 800, 0)
			w.Step()
			if math.Abs(w.Cells(id)[0].X-800-220.0/30*1.1) > .001 {
				t.Fatal("speed mutation inactive")
			}
		case 1:
			fixture(w, Food, 800, 800, 3)
			_ = w.Input(id, 800, 800, 0)
			w.Step()
			if math.Abs(w.Mass(id)-39.6) > .001 {
				t.Fatal("food mutation inactive", w.Mass(id))
			}
		case 2:
			place(w, id, 800, 800, 100)
			_ = w.Input(id, 1500, 800, Split)
			w.Step()
			if len(w.Cells(id)) != 2 || w.Cells(id)[0].MergeAt != 301 {
				t.Fatal("cohesion mutation inactive", w.Cells(id))
			}
		}
		invariant(t, w)
		w.Config.Mode = Classic
		if w.Valid() == nil {
			t.Fatal("classic accepted rogue mutation")
		}
	}
}
func TestConsumptionRatioAndContainment(t *testing.T) {
	w := empty()
	a := w.AddPlayer("A", false, "")
	b := w.AddPlayer("B", false, "")
	place(w, a, 800, 800, 100)
	place(w, b, 800, 800, 90)
	_ = w.Input(a, 800, 800, 0)
	_ = w.Input(b, 800, 800, 0)
	w.Step()
	if !w.Player(b).Alive {
		t.Fatal("near-equal mass cannot be eaten")
	}
	place(w, a, 800, 800, 150)
	place(w, b, 800, 800, 90)
	w.Step()
	if w.Player(b).Alive || w.Player(a).Kills != 1 || math.Abs(w.Mass(a)-240) > .001 {
		t.Fatal("consumption and elimination not recorded")
	}
	invariant(t, w)
}
func TestFoodTransfersMassOnce(t *testing.T) {
	w := empty()
	id := w.AddPlayer("A", false, "")
	place(w, id, 800, 800, 36)
	_ = w.Input(id, 800, 800, 0)
	fixture(w, Food, 800, 800, 1)
	w.Step()
	if w.Mass(id) != 37 {
		t.Fatal(w.Mass(id))
	}
	w.Step()
	if w.Mass(id) != 37 {
		t.Fatal("food duplicated")
	}
	invariant(t, w)
}
func TestSplitConservesMassAndHasCooldown(t *testing.T) {
	w := empty()
	id := w.AddPlayer("A", false, "")
	place(w, id, 800, 800, 400)
	_ = w.Input(id, 1400, 800, Split)
	w.Step()
	if len(w.Cells(id)) != 2 || w.Mass(id) != 400 {
		t.Fatal("split mass")
	}
	_ = w.Input(id, 1400, 800, Split)
	w.Step()
	if len(w.Cells(id)) != 2 {
		t.Fatal("cooldown ignored")
	}
	for range 300 {
		_ = w.Input(id, 800, 800, Split)
		w.Step()
		if len(w.Cells(id)) > MaxCells {
			t.Fatal("too many cells")
		}
		invariant(t, w)
	}
}
func TestEjectHasCostAndCooldown(t *testing.T) {
	w := empty()
	id := w.AddPlayer("A", false, "")
	place(w, id, 800, 800, 100)
	_ = w.Input(id, 1400, 800, Eject)
	w.Step()
	if w.Mass(id) != 100-EjectCost {
		t.Fatal(w.Mass(id))
	}
	n := 0
	for _, e := range w.Entities {
		if e.Kind == Ejected {
			n++
			if e.Mass != EjectMass {
				t.Fatal(e.Mass)
			}
		}
	}
	if n != 1 {
		t.Fatal(n)
	}
	_ = w.Input(id, 1400, 800, Eject)
	w.Step()
	if w.Mass(id) != 100-EjectCost {
		t.Fatal("eject cooldown ignored")
	}
	invariant(t, w)
}
func TestOwnCellsMergeOnlyAfterDeadline(t *testing.T) {
	w := empty()
	id := w.AddPlayer("A", false, "")
	place(w, id, 800, 800, 100)
	_ = w.Input(id, 1400, 800, Split)
	w.Step()
	for i := range w.Entities {
		e := &w.Entities[i]
		if e.Owner == id {
			e.X = 800
			e.Y = 800
			e.VX = 0
			e.VY = 0
			e.MergeAt = w.Tick + 2
		}
	}
	_ = w.Input(id, 800, 800, 0)
	w.Step()
	if len(w.Cells(id)) != 2 {
		t.Fatal("early merge")
	}
	w.Step()
	if len(w.Cells(id)) != 1 {
		t.Fatal("merge missing")
	}
	invariant(t, w)
}
func TestVirusExplodesLargeCellWithoutCreatingMass(t *testing.T) {
	w := empty()
	id := w.AddPlayer("A", false, "")
	place(w, id, 800, 800, 200)
	_ = w.Input(id, 800, 800, 0)
	fixture(w, Virus, 800, 800, 100)
	w.Step()
	if len(w.Cells(id)) < 2 || len(w.Cells(id)) > MaxCells || math.Abs(w.Mass(id)-300) > .001 {
		t.Fatal(len(w.Cells(id)), w.Mass(id))
	}
	invariant(t, w)
}
func TestFeedingVirusLaunchesAnother(t *testing.T) {
	w := empty()
	fixture(w, Virus, 800, 800, 100)
	for range 7 {
		fixture(w, Ejected, 800, 800, EjectMass)
		w.Entities[len(w.Entities)-1].VX = 300
	}
	w.Step()
	n := 0
	for _, e := range w.Entities {
		if e.Kind == Virus {
			n++
		}
	}
	if n != 2 {
		t.Fatal("feeding did not create a launched virus", n)
	}
	invariant(t, w)
}
func TestTeamsNeverEatAllies(t *testing.T) {
	w := New(9, Config{Size: 1600, Mode: Teams})
	a := w.AddPlayer("A", false, "")
	b := w.AddPlayer("B", false, "")
	w.Player(b).Team = w.Player(a).Team
	for i := range w.Entities {
		if w.Entities[i].Owner == b {
			w.Entities[i].Team = w.Player(a).Team
		}
	}
	place(w, a, 800, 800, 200)
	place(w, b, 800, 800, 36)
	_ = w.Input(a, 800, 800, 0)
	_ = w.Input(b, 800, 800, 0)
	w.Step()
	if !w.Player(b).Alive {
		t.Fatal("ate teammate")
	}
	invariant(t, w)
}
func TestInvalidInputAndRespawn(t *testing.T) {
	w := empty()
	id := w.AddPlayer("A", false, "")
	if w.Input(id, math.NaN(), 0, 0) == nil || w.Input(id, 0, 0, 128) == nil || w.Input(999, 0, 0, 0) == nil {
		t.Fatal("invalid input accepted")
	}
	if w.Respawn(id) == nil {
		t.Fatal("alive player spawned twice")
	}
	killer := w.AddPlayer("B", false, "")
	place(w, killer, 800, 800, 300)
	place(w, id, 800, 800, 36)
	_ = w.Input(killer, 800, 800, 0)
	_ = w.Input(id, 800, 800, 0)
	w.Step()
	if w.Respawn(id) == nil {
		t.Fatal("respawn cooldown ignored")
	}
	for range 61 {
		w.Step()
	}
	if err := w.Respawn(id); err != nil {
		t.Fatal(err)
	}
	if !w.Player(id).Alive {
		t.Fatal("respawn missing")
	}
	invariant(t, w)
}
func TestObservationCannotSeeOutsideInterestOrChangeRandomState(t *testing.T) {
	w := empty()
	a := w.AddPlayer("A", false, "")
	b := w.AddPlayer("B", true, "cautious")
	place(w, a, 70, 70, 36)
	place(w, b, 1500, 1500, 36)
	before, _ := json.Marshal(w)
	obs := w.Observe(a)
	for _, e := range obs.Entities {
		if e.Owner == b {
			t.Fatal("bot can see through the world")
		}
	}
	after, _ := json.Marshal(w)
	if string(before) != string(after) {
		t.Fatal("observation changed simulation")
	}
}
func TestSeededReplayAndCheckpoint(t *testing.T) {
	a := New(42, Config{Size: 1600, Food: 150, Viruses: 5, Mode: Classic})
	b := New(42, a.Config)
	for _, w := range []*World{a, b} {
		w.AddPlayer("A", false, "")
		w.AddPlayer("B", true, "balanced")
	}
	for tick := range 1000 {
		for _, w := range []*World{a, b} {
			for _, p := range w.Players {
				_ = w.Input(p.ID, float64((tick*17)%1500), float64((tick*13)%1500), uint8(tick%3))
				if !p.Alive {
					_ = w.Respawn(p.ID)
				}
			}
			w.Step()
			invariant(t, w)
		}
		if tick == 500 {
			data, _ := json.Marshal(b)
			b = &World{}
			if err := json.Unmarshal(data, b); err != nil {
				t.Fatal(err)
			}
		}
	}
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	if string(aj) != string(bj) {
		t.Fatal("same inputs/seed diverged across checkpoint")
	}
}

func TestEntityBudgetsBoundEjectionAndVirusGrowth(t *testing.T) {
	w := empty()
	id := w.AddPlayer("A", false, "")
	place(w, id, 800, 800, 100)
	for range MaxEjected {
		fixture(w, Ejected, 100, 100, EjectMass)
	}
	_ = w.Input(id, 1400, 800, Eject)
	w.Step()
	if w.Mass(id) != 100 {
		t.Fatal("charged mass despite exhausted pellet budget")
	}
	invariant(t, w)
	w = empty()
	for range MaxViruses {
		fixture(w, Virus, 100, 100, 100)
	}
	for range 7 {
		fixture(w, Ejected, 100, 100, EjectMass)
	}
	w.Step()
	count := 0
	for _, e := range w.Entities {
		if e.Kind == Virus {
			count++
		}
	}
	if count > MaxViruses {
		t.Fatal("virus budget exceeded", count)
	}
	invariant(t, w)
}

func TestSplitHasUsefulReachAndResponsiveChaining(t *testing.T) {
	w := New(4, Config{Size: 6000, Mode: Classic})
	id := w.AddPlayer("A", false, "")
	place(w, id, 2500, 3000, 800)
	_ = w.Input(id, 5500, 3000, Split)
	w.Step()
	cells := w.Cells(id)
	if math.Abs(cells[1].X-cells[0].X) < Radius(cells[0].Mass)+Radius(cells[1].Mass) {
		t.Fatal("split children overlap at launch")
	}
	for range 14 {
		w.Step()
	}
	cells = w.Cells(id)
	if math.Abs(cells[1].X-cells[0].X) < 300 {
		t.Fatal("split has insufficient tactical reach")
	}
	_ = w.Input(id, 5500, 3000, Split)
	w.Step()
	if len(w.Cells(id)) != 4 {
		t.Fatal("split chain blocked too long")
	}
	invariant(t, w)
}
func TestFoodHasVisibleMassVariety(t *testing.T) {
	w := New(8, DefaultConfig(Classic))
	counts := map[float64]int{}
	for _, e := range w.Entities {
		if e.Kind == Food {
			counts[e.Mass]++
		}
	}
	if len(counts) < 3 || w.Config.Food < 3500 {
		t.Fatal("arena lacks food variety/density", counts, w.Config.Food)
	}
	invariant(t, w)
}
func TestActionsBufferOnlyUntilTheirShortCooldown(t *testing.T) {
	w := empty()
	id := w.AddPlayer("A", false, "")
	place(w, id, 800, 800, 800)
	_ = w.Input(id, 1400, 800, Split)
	w.Step()
	if !w.ActionOccurred() {
		t.Fatal("split feedback missing")
	}
	_ = w.Input(id, 1400, 800, Split)
	w.Step()
	if w.ActionOccurred() {
		t.Fatal("cooldown should still apply")
	}
	for range SplitCooldownTicks {
		w.Step()
	}
	if len(w.Cells(id)) != 4 {
		t.Fatal("buffered split was lost")
	}
	for range 5 {
		w.Step()
	}
	if len(w.Cells(id)) != 4 {
		t.Fatal("buffered action executed repeatedly")
	}
	invariant(t, w)
}

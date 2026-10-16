package expedition

import (
	"encoding/json"
	"github.com/spy-exe/cells/apps/game-server/internal/bot"
	"github.com/spy-exe/cells/apps/game-server/internal/engine"
	"testing"
)

func scene() (*engine.World, uint32, *Run) {
	w := engine.New(17, engine.Config{Size: 3200, Mode: engine.Rogue})
	id := w.AddPlayer("Explorer", false, "")
	return w, id, New(17, w, id)
}
func grant(w *engine.World, id uint32, mass float64) {
	for i := range w.Entities {
		e := &w.Entities[i]
		if e.Kind == engine.Cell && e.Owner == id {
			w.Sources += mass - e.Mass
			e.Mass = mass
			break
		}
	}
}
func TestSectorChoicesVictoryAndReplay(t *testing.T) {
	w, id, r := scene()
	other, otherID, second := scene()
	for stage := 1; stage <= 3; stage++ {
		if r.Stage != stage || r.Phase != "running" {
			t.Fatal(r)
		}
		grant(w, id, 200)
		grant(other, otherID, 200)
		w.Tick = r.Deadline
		other.Tick = second.Deadline
		r.Advance(w)
		second.Advance(other)
		if stage < 3 {
			if r.Phase != "choice" {
				t.Fatal(r)
			}
			rev := r.Revision
			if err := r.Choose(w, "cilia", rev-1); err == nil {
				t.Fatal("stale choice accepted")
			}
			if err := r.Choose(w, "invalid", rev); err == nil {
				t.Fatal("unknown mutation accepted")
			}
			if err := r.Choose(w, "cilia", rev); err != nil {
				t.Fatal(err)
			}
			if err := second.Choose(other, "cilia", second.Revision); err != nil {
				t.Fatal(err)
			}
			if err := r.Choose(w, "cilia", rev); err == nil {
				t.Fatal("duplicate choice applied")
			}
		}
		if err := w.Valid(); err != nil {
			t.Fatal(err)
		}
	}
	if r.Phase != "victory" || w.Player(id).Mutations[0] != 2 {
		t.Fatal(r, w.Player(id))
	}
	a, _ := json.Marshal(w)
	b, _ := json.Marshal(other)
	if string(a) != string(b) {
		t.Fatal("seed replay diverged")
	}
	encoded, _ := json.Marshal(r)
	var loaded Run
	if err := json.Unmarshal(encoded, &loaded); err != nil {
		t.Fatal(err)
	}
	if err := loaded.Validate(w); err != nil {
		t.Fatal(err)
	}
	if loaded.View(w).Phase != "victory" {
		t.Fatal("persisted victory lost")
	}
}
func TestFailureDeathAndIneligibleMutation(t *testing.T) {
	w, id, r := scene()
	if err := r.Choose(w, "digestion", r.Revision); err == nil {
		t.Fatal("mutation before offer")
	}
	w.Tick = r.Deadline
	r.Advance(w)
	if r.Phase != "failed" || r.Reason != "biomass" {
		t.Fatal(r)
	}
	before := r.Revision
	r.Advance(w)
	if r.Revision != before {
		t.Fatal("terminal advanced")
	}
	w, id, r = scene()
	w.Player(id).Alive = false
	r.Advance(w)
	if r.Phase != "fallen" {
		t.Fatal(r)
	}
	r.Stage = 99
	if err := r.Validate(w); err == nil {
		t.Fatal("invalid stage accepted")
	}
}
func TestEveryMutationHasBoundedEffectsAndCannotAlterClassic(t *testing.T) {
	for _, choice := range []string{"cilia", "digestion", "cohesion"} {
		w, id, r := scene()
		grant(w, id, 100)
		w.Tick = r.Deadline
		r.Advance(w)
		if err := r.Choose(w, choice, r.Revision); err != nil {
			t.Fatal(choice, err)
		}
		view := r.View(w)
		if view.Ranks[choice] != 1 || view.SecondsLeft != 25 || view.TargetMass != 100 {
			t.Fatal(view)
		}
		classic := engine.New(17, engine.DefaultConfig(engine.Classic))
		r.Phase = "choice"
		if err := r.Choose(classic, choice, r.Revision); err == nil {
			t.Fatal("classic changed")
		}
	}
}

// Exercise complete ticks, tactical opponents and food collisions, without
// granting mass or jumping time. This is a playability check, not a win-rate claim.
func TestSeededExpeditionIsWinnableThroughRealSimulation(t *testing.T) {
	wins := 0
	for _, seed := range []uint64{17, 42, 91, 123} {
		w := engine.New(seed, engine.DefaultConfig(engine.Rogue))
		id := w.AddPlayer("Explorer", false, "")
		r := New(seed, w, id)
		memories := map[uint32]*bot.Memory{}
		for step := 0; step < 3*SectorTicks+3; step++ {
			if r.Phase == "choice" {
				if err := r.Choose(w, "digestion", r.Revision); err != nil {
					t.Fatal(err)
				}
			}
			if r.Phase != "running" {
				break
			}
			if w.Tick%6 == 0 {
				for _, p := range w.Players {
					if !p.Alive {
						continue
					}
					m := memories[p.ID]
					if m == nil {
						m = &bot.Memory{}
						memories[p.ID] = m
					}
					profile := p.Profile
					if !p.Bot {
						profile = "balanced"
					}
					intent := bot.Decide(w.Observe(p.ID), m, profile)
					_ = w.Input(p.ID, intent.X, intent.Y, intent.Actions)
				}
			}
			w.Step()
			r.Advance(w)
		}
		if err := w.Valid(); err != nil {
			t.Fatal(err)
		}
		t.Logf("seed=%d phase=%s stage=%d peak=%.1f kills=%d", seed, r.Phase, r.Stage, w.Player(id).RunPeak, w.Player(id).RunKills)
		if r.Phase == "victory" {
			wins++
		}
	}
	if wins == 0 {
		t.Fatal("no reference agent can complete the expedition")
	}
}

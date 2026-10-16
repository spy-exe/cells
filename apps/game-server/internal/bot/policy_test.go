package bot

import (
	"github.com/spy-exe/cells/apps/game-server/internal/engine"
	"math"
	"testing"
)

func scene() engine.Observation {
	return engine.Observation{Self: 1, Alive: true, Size: 6000, X: 3000, Y: 3000, View: 1000, Mass: 100, Entities: []engine.Visible{{ID: 2, Owner: 1, Kind: engine.Cell, X: 3000, Y: 3000, Mass: 100}}}
}
func TestFindFood(t *testing.T) {
	o := scene()
	o.Entities = append(o.Entities, engine.Visible{ID: 3, Kind: engine.Food, X: 3100, Y: 3000, Mass: 1})
	d := Decide(o, &Memory{}, "balanced")
	if d.X <= o.X || d.Reason != "forage" {
		t.Fatal(d)
	}
}
func TestFleeThreat(t *testing.T) {
	o := scene()
	o.Entities = append(o.Entities, engine.Visible{ID: 3, Owner: 9, Kind: engine.Cell, X: 3130, Y: 3000, Mass: 300})
	d := Decide(o, &Memory{}, "balanced")
	if d.X >= o.X || d.Reason != "flee" || d.Actions != 0 {
		t.Fatal(d)
	}
}
func TestSplitOnlyWithLandingAdvantage(t *testing.T) {
	o := scene()
	o.Mass = 400
	o.Entities[0].Mass = 400
	o.Entities = append(o.Entities, engine.Visible{ID: 3, Owner: 9, Kind: engine.Cell, X: 3200, Y: 3000, Mass: 60})
	d := Decide(o, &Memory{}, "aggressive")
	if d.Actions&engine.Split == 0 || d.Reason != "split-hunt" {
		t.Fatal(d)
	}
	o.Entities = append(o.Entities, engine.Visible{ID: 4, Owner: 8, Kind: engine.Cell, X: 3250, Y: 3000, Mass: 500})
	d = Decide(o, &Memory{}, "aggressive")
	if d.Actions&engine.Split != 0 {
		t.Fatal("suicidal split", d)
	}
}
func TestAvoidVirusAndUseCover(t *testing.T) {
	o := scene()
	o.Mass = 300
	o.Entities[0].Mass = 300
	o.Entities = append(o.Entities, engine.Visible{ID: 3, Kind: engine.Virus, X: 3100, Y: 3000, Mass: 100}, engine.Visible{ID: 4, Kind: engine.Food, X: 3150, Y: 3000, Mass: 1})
	d := Decide(o, &Memory{}, "cautious")
	if math.Abs(d.Y-3000) < 20 && d.X > 3050 {
		t.Fatal("walked into virus", d)
	}
}
func TestProfilesAndDeterminism(t *testing.T) {
	for _, p := range Profiles {
		a, b := Decide(scene(), &Memory{}, p), Decide(scene(), &Memory{}, p)
		if a != b || math.IsNaN(a.X) || a.X < 0 || a.X > 6000 {
			t.Fatal(p, a, b)
		}
	}
	o := scene()
	o.Alive = false
	if d := Decide(o, &Memory{}, "balanced"); d.Actions != 0 {
		t.Fatal(d)
	}
}
func TestVirusFeedingOpportunity(t *testing.T) {
	o := scene()
	o.Mass = 200
	o.Entities[0].Mass = 200
	o.Entities = append(o.Entities, engine.Visible{ID: 3, Kind: engine.Virus, X: 3250, Y: 3000, Mass: 100}, engine.Visible{ID: 4, Owner: 9, Kind: engine.Cell, X: 3650, Y: 3000, Mass: 600})
	d := Decide(o, &Memory{}, "balanced")
	if d.Reason != "feed-virus" || d.Actions != engine.Eject {
		t.Fatal(d)
	}
}
func TestTeamAndWallAwareness(t *testing.T) {
	o := scene()
	o.X = 40
	o.Entities[0].X = 40
	o.Entities = append(o.Entities, engine.Visible{ID: 3, Owner: 9, Kind: engine.Cell, X: 170, Y: 3000, Mass: 300})
	d := Decide(o, &Memory{}, "cautious")
	if d.X < 40 {
		t.Fatal("fled into boundary", d)
	}
	o.Mode = engine.Teams
	d = Decide(o, &Memory{}, "balanced")
	if d.Reason == "flee" {
		t.Fatal("fled from ally", d)
	}
}
func TestVelocityPredictionAndCooldown(t *testing.T) {
	o := scene()
	o.Mass = 400
	o.Entities[0].Mass = 400
	o.Entities = append(o.Entities, engine.Visible{ID: 3, Owner: 9, Kind: engine.Cell, X: 3200, Y: 3000, Mass: 60})
	m := &Memory{}
	Decide(o, m, "balanced")
	o.Tick = 6
	o.Entities[1].X = 3250
	o.SplitIn = 12
	d := Decide(o, m, "balanced")
	if d.X <= 3250 || d.Actions&engine.Split != 0 {
		t.Fatal("velocity/cooldown", d)
	}
	for _, profile := range []string{"beginner", "cautious"} {
		if d = Decide(o, &Memory{}, profile); d.Actions&engine.Split != 0 {
			t.Fatal(profile, d)
		}
	}
	o.Entities = nil
	if d = Decide(o, m, "balanced"); d.Reason != "observe" {
		t.Fatal("missing own body", d)
	}
}
func TestMemoryExpiresAndSmallCellsUseVirusCover(t *testing.T) {
	o := scene()
	o.Mass = 36
	o.Entities[0].Mass = 36
	o.Entities = append(o.Entities, engine.Visible{ID: 3, Owner: 9, Kind: engine.Cell, X: 3200, Y: 3000, Mass: 200}, engine.Visible{ID: 4, Kind: engine.Virus, X: 3100, Y: 3000, Mass: 100})
	m := &Memory{Seen: map[uint32]Track{999: {Tick: 0}}}
	o.Tick = 180
	d := Decide(o, m, "opportunist")
	if _, exists := m.Seen[999]; exists || d.Reason != "flee" {
		t.Fatal(d, m)
	}
}

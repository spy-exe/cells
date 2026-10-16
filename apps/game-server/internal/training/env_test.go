package training

import (
	"encoding/json"
	"github.com/spy-exe/cells/apps/game-server/internal/engine"
	"math"
	"reflect"
	"testing"
)

func smallConfig() Config {
	c := DefaultConfig()
	c.Agents = 2
	c.Opponents = 2
	c.Horizon = 25
	c.Arena.Food = 60
	c.Arena.Viruses = 2
	return c
}
func TestSeededEpisodeIsFiniteAndReproducible(t *testing.T) {
	a, err := New(71, smallConfig())
	if err != nil {
		t.Fatal(err)
	}
	b, err := New(71, smallConfig())
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 25; i++ {
		x, err := a.StepPolicy("random")
		if err != nil {
			t.Fatal(err)
		}
		y, err := b.StepPolicy("random")
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(x, y) || a.Hash() != b.Hash() {
			t.Fatal("seed/action replay diverged", i)
		}
		for _, s := range x.Agents {
			if len(s.Observation) != 80 || math.IsNaN(s.Reward) {
				t.Fatal("invalid observation/reward")
			}
			for _, v := range s.Observation {
				if v < -1 || v > 1 || math.IsNaN(float64(v)) {
					t.Fatal("outside observation space", v)
				}
			}
		}
		if x.Done {
			return
		}
	}
	t.Fatal("episode never terminated at its horizon")
}
func TestRejectsActionsWithoutChangingAuthoritativeState(t *testing.T) {
	e, err := New(12, smallConfig())
	if err != nil {
		t.Fatal(err)
	}
	before := e.Hash()
	for _, input := range []map[string]Action{
		{"agent_0": {math.NaN(), 0, 0, 0}, "agent_1": {}},
		{"agent_0": {1.1, 0, 0, 0}, "agent_1": {}},
		{"unknown": {}, "agent_1": {}},
		{"agent_0": {}},
	} {
		if _, err = e.Step(input); err == nil {
			t.Fatal("accepted invalid/missing action")
		}
		if e.Hash() != before {
			t.Fatal("invalid action changed the world")
		}
	}
}
func TestRewardCannotProfitFromOwnEjectionOrPassiveInfiniteSurvival(t *testing.T) {
	cycle := Reward(100, 86, 0, false, 1) + Reward(86, 98, 0, false, 2)
	if cycle >= 0 {
		t.Fatal("own feeding loop earns reward", cycle)
	}
	idle := 0.
	for step := 1; step <= 600; step++ {
		idle += Reward(36, 36, 0, false, step)
	}
	if idle >= 0 {
		t.Fatal("passive survival earns net reward", idle)
	}
	if Reward(36, 0, 0, true, 2) >= -1 {
		t.Fatal("death has no meaningful consequence")
	}
	if Reward(36, 72, 1, false, 2) <= Reward(36, 72, 0, false, 2) {
		t.Fatal("elimination is ignored")
	}
}
func TestFeaturesUseOnlyVisibleEntitiesAndDoNotMutateRNG(t *testing.T) {
	w := engine.New(18, engine.Config{Size: 6000, Mode: engine.Arena})
	own := w.AddPlayer("Local", false, "balanced")
	w.AddPlayer("Distant", true, "balanced")
	for i := range w.Entities {
		c := &w.Entities[i]
		if c.Owner == own {
			c.X = 400
			c.Y = 400
		} else {
			c.X = 5600
			c.Y = 5600
		}
		p := w.Player(c.Owner)
		p.TargetX = c.X
		p.TargetY = c.Y
	}
	before, _ := json.Marshal(w)
	first := Encode(w.Observe(own))
	after, _ := json.Marshal(w)
	if string(before) != string(after) {
		t.Fatal("observation changed RNG/state")
	}
	for i := range w.Entities {
		if w.Entities[i].Owner != own {
			w.Entities[i].Mass = 5000
		}
	}
	w.Step()
	if Encode(w.Observe(own)) != first {
		t.Fatal("hidden enemy leaked into local features")
	}
}

// Package training drives the production engine without network, database or rendering.
package training

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/spy-exe/cells/apps/game-server/internal/bot"
	"github.com/spy-exe/cells/apps/game-server/internal/engine"
)

type Action [4]float64
type Config struct {
	Agents, Opponents, Horizon int
	Arena                      engine.Config
}

func DefaultConfig() Config {
	return Config{Agents: 1, Opponents: 7, Horizon: 300, Arena: engine.Config{Size: 3200, Food: 1200, Viruses: 12, Mode: engine.Arena}}
}

type Metrics struct {
	Mass            float64 `json:"mass"`
	Peak            float64 `json:"peak"`
	MassArea        float64 `json:"mass_area"`
	SurvivalSeconds float64 `json:"survival_seconds"`
	Kills           int     `json:"kills"`
	Win             bool    `json:"win"`
}
type AgentResult struct {
	Observation [80]float32 `json:"observation"`
	Reward      float64     `json:"reward"`
	Terminated  bool        `json:"terminated"`
	Truncated   bool        `json:"truncated"`
	Metrics     Metrics     `json:"metrics"`
}
type Result struct {
	Tick   uint64                 `json:"tick"`
	Done   bool                   `json:"done"`
	Agents map[string]AgentResult `json:"agents"`
	Hash   string                 `json:"hash,omitempty"`
}
type Env struct {
	World           *engine.World
	Config          Config
	LastPolicyNanos int64
	ids             []uint32
	names           []string
	memories        map[uint32]*bot.Memory
	areas           map[uint32]float64
	random          uint64
	steps           int
	done            bool
}

func New(seed uint64, cfg Config) (*Env, error) {
	if cfg.Agents < 1 || cfg.Agents > 4 || cfg.Opponents < 0 || cfg.Opponents > 12 || cfg.Horizon < 1 || cfg.Horizon > 1200 {
		return nil, errors.New("invalid bounded episode configuration")
	}
	if cfg.Arena.Size < 800 || cfg.Arena.Size > 20000 || math.IsNaN(cfg.Arena.Size) || cfg.Arena.Food < 0 || cfg.Arena.Food > 3600 || cfg.Arena.Viruses < 0 || cfg.Arena.Viruses > 128 {
		return nil, errors.New("invalid training arena")
	}
	cfg.Arena.Mode = engine.Arena
	e := &Env{World: engine.New(seed, cfg.Arena), Config: cfg, memories: map[uint32]*bot.Memory{}, areas: map[uint32]float64{}, random: seed ^ 0xd1b54a32d192ed03}
	if e.random == 0 {
		e.random = 1
	}
	for i := 0; i < cfg.Agents; i++ {
		name := fmt.Sprintf("agent_%d", i)
		e.names = append(e.names, name)
		e.ids = append(e.ids, e.World.AddPlayer(name, false, "balanced"))
	}
	for i := 0; i < cfg.Opponents; i++ {
		e.World.AddPlayer(fmt.Sprintf("reference_%d", i), true, bot.Profiles[(i+int(seed%uint64(len(bot.Profiles))))%len(bot.Profiles)])
	}
	return e, nil
}
func (e *Env) memory(id uint32) *bot.Memory {
	if e.memories[id] == nil {
		e.memories[id] = &bot.Memory{}
	}
	return e.memories[id]
}
func (e *Env) uniform() float64 {
	x := e.random
	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	e.random = x
	return float64((x*2685821657736338717)>>11) / (1 << 53)
}
func Reward(beforeMass, afterMass float64, kills int, died bool, step int) float64 {
	r := math.Log1p(afterMass) - math.Log1p(beforeMass) + .2*float64(kills) - .001
	if died {
		r -= 1
	} else if step == 100 || step == 300 || step == 500 {
		r += .05
	}
	return r
}
func (e *Env) allDead() bool {
	for _, id := range e.ids {
		if e.World.Player(id).Alive {
			return false
		}
	}
	return true
}
func (e *Env) Snapshot() Result {
	r := Result{Tick: e.World.Tick, Done: e.done, Agents: map[string]AgentResult{}}
	maxMass := 0.
	if e.done {
		for _, p := range e.World.Players {
			if p.Alive {
				maxMass = math.Max(maxMass, e.World.Mass(p.ID))
			}
		}
	}
	for i, id := range e.ids {
		p := e.World.Player(id)
		mass := e.World.Mass(id)
		until := e.World.Tick
		if !p.Alive {
			until = p.DiedAt
		}
		r.Agents[e.names[i]] = AgentResult{Observation: Encode(e.World.Observe(id)), Terminated: !p.Alive, Truncated: e.steps >= e.Config.Horizon && p.Alive, Metrics: Metrics{Mass: mass, Peak: math.Max(p.Peak, mass), MassArea: e.areas[id], SurvivalSeconds: float64(until) / engine.TickRate, Kills: p.Kills, Win: e.done && p.Alive && mass >= maxMass}}
	}
	if e.done {
		r.Hash = e.Hash()
	}
	return r
}
func (e *Env) Hash() string {
	encoded, _ := json.Marshal(e.World)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}
func (e *Env) Step(actions map[string]Action) (Result, error) {
	if e.done {
		return Result{}, errors.New("episode is complete")
	}
	active := 0
	for i, id := range e.ids {
		if !e.World.Player(id).Alive {
			continue
		}
		active++
		a, ok := actions[e.names[i]]
		if !ok {
			return Result{}, errors.New("missing live-agent action")
		}
		for _, v := range a {
			if math.IsNaN(v) || math.IsInf(v, 0) || v < -1 || v > 1 {
				return Result{}, errors.New("action outside normalized space")
			}
		}
	}
	if len(actions) != active {
		return Result{}, errors.New("unexpected agent action")
	}
	before := make(map[uint32]float64, len(e.ids))
	kills := make(map[uint32]int, len(e.ids))
	alive := make(map[uint32]bool, len(e.ids))
	for i, id := range e.ids {
		p := e.World.Player(id)
		before[id] = e.World.Mass(id)
		kills[id] = p.Kills
		alive[id] = p.Alive
		if !p.Alive {
			continue
		}
		a := actions[e.names[i]]
		o := e.World.Observe(id)
		bits := uint8(0)
		if a[2] > .5 {
			bits |= engine.Split
		}
		if a[3] > .5 {
			bits |= engine.Eject
		}
		if err := e.World.Input(id, o.X+a[0]*o.View, o.Y+a[1]*o.View, bits); err != nil {
			return Result{}, err
		}
	}
	e.LastPolicyNanos = 0
	for n := 0; n < 6; n++ {
		if e.World.Tick%6 == 0 {
			for _, p := range e.World.Players {
				if !p.Bot || !p.Alive {
					continue
				}
				d := bot.Decide(e.World.Observe(p.ID), e.memory(p.ID), p.Profile)
				if err := e.World.Input(p.ID, d.X, d.Y, d.Actions); err != nil {
					return Result{}, err
				}
			}
		}
		e.World.Step()
		for _, id := range e.ids {
			e.areas[id] += e.World.Mass(id) / engine.TickRate
		}
		if e.allDead() {
			break
		}
	}
	if err := e.World.Valid(); err != nil {
		return Result{}, fmt.Errorf("invalid simulation: %w", err)
	}
	e.steps++
	e.done = e.steps >= e.Config.Horizon || e.allDead()
	r := e.Snapshot()
	for i, id := range e.ids {
		if alive[id] {
			a := r.Agents[e.names[i]]
			a.Reward = Reward(before[id], a.Metrics.Mass, a.Metrics.Kills-kills[id], a.Terminated, e.steps)
			r.Agents[e.names[i]] = a
		}
	}
	return r, nil
}
func (e *Env) StepPolicy(policy string) (Result, error) {
	if policy != "random" && policy != "heuristic" {
		return Result{}, errors.New("unknown reference policy")
	}
	if e.done {
		return Result{}, errors.New("episode is complete")
	}
	actions := map[string]Action{}
	var nanos int64
	for i, id := range e.ids {
		if !e.World.Player(id).Alive {
			continue
		}
		o := e.World.Observe(id)
		start := time.Now()
		a := Action{}
		if policy == "random" {
			for j := range a {
				a[j] = e.uniform()*2 - 1
			}
		} else {
			d := bot.Decide(o, e.memory(id), "balanced")
			a[0] = clip((d.X - o.X) / o.View)
			a[1] = clip((d.Y - o.Y) / o.View)
			a[2] = -1
			a[3] = -1
			if d.Actions&engine.Split != 0 {
				a[2] = 1
			}
			if d.Actions&engine.Eject != 0 {
				a[3] = 1
			}
		}
		nanos += time.Since(start).Nanoseconds()
		actions[e.names[i]] = a
	}
	r, err := e.Step(actions)
	e.LastPolicyNanos = nanos
	return r, err
}

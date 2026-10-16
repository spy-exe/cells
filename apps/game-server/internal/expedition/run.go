// Package expedition supplies a finite solo rules module over the shared engine.
package expedition

import (
	"errors"
	"fmt"
	"github.com/spy-exe/cells/apps/game-server/internal/engine"
	"math"
)

const SectorTicks = 25 * engine.TickRate

var targets = [3]float64{60, 100, 160}
var mutations = [3]string{"cilia", "digestion", "cohesion"}

type Run struct {
	Seed               uint64
	Player             uint32
	Stage              int
	Phase, Reason      string
	Deadline, Revision uint64
}
type Offer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Detail string `json:"detail"`
}
type View struct {
	Stage       int              `json:"stage"`
	Phase       string           `json:"phase"`
	Reason      string           `json:"reason"`
	Revision    uint64           `json:"revision"`
	SecondsLeft int              `json:"seconds_left"`
	TargetMass  float64          `json:"target_mass"`
	Mass        float64          `json:"mass"`
	Peak        float64          `json:"peak"`
	Kills       int              `json:"kills"`
	SeedLabel   string           `json:"seed_label"`
	Ranks       map[string]uint8 `json:"ranks"`
	Offers      []Offer          `json:"offers"`
}

func New(seed uint64, w *engine.World, player uint32) *Run {
	r := &Run{Seed: seed, Player: player, Stage: 1, Phase: "running", Deadline: w.Tick + SectorTicks, Revision: 1}
	r.wave(w)
	return r
}
func (r *Run) wave(w *engine.World) {
	profiles := []string{"beginner", "balanced", "opportunist", "aggressive"}
	names := []string{"Esporo", "Micélio", "Deriva", "Pólen"}
	for i, name := range names {
		w.AddPlayer(fmt.Sprintf("%s %d", name, r.Stage), true, profiles[min(r.Stage-1, i)])
	}
	if r.Stage == 3 {
		id := w.AddPlayer("Sentinela", true, "aggressive")
		mass := math.Min(500, math.Max(180, w.Mass(r.Player)*1.2))
		for i := range w.Entities {
			e := &w.Entities[i]
			if e.Kind == engine.Cell && e.Owner == id {
				w.Sources += mass - e.Mass
				e.Mass = mass
				radius := engine.Radius(mass)
				e.X = math.Max(radius, math.Min(w.Config.Size-radius, e.X))
				e.Y = math.Max(radius, math.Min(w.Config.Size-radius, e.Y))
				break
			}
		}
	}
}
func (r *Run) Advance(w *engine.World) {
	if r.Phase != "running" {
		return
	}
	p := w.Player(r.Player)
	if p == nil || !p.Alive {
		r.Phase = "fallen"
		r.Reason = "consumed"
		r.Revision++
		return
	}
	if w.Tick < r.Deadline {
		return
	}
	switch {
	case w.Mass(r.Player) < targets[r.Stage-1]:
		r.Phase = "failed"
		r.Reason = "biomass"
	case r.Stage == 3:
		r.Phase = "victory"
	default:
		r.Phase = "choice"
	}
	r.Revision++
}
func (r *Run) Choose(w *engine.World, id string, revision uint64) error {
	p := w.Player(r.Player)
	if w.Config.Mode != engine.Rogue || r.Phase != "choice" || r.Revision != revision || p == nil || !p.Alive {
		return errors.New("mutation is no longer available")
	}
	index := -1
	for i, mutation := range mutations {
		if id == mutation {
			index = i
		}
	}
	if index < 0 || p.Mutations[index] >= 3 {
		return errors.New("invalid mutation")
	}
	p.Mutations[index]++
	r.Stage++
	r.Phase = "running"
	r.Reason = ""
	r.Deadline = w.Tick + SectorTicks
	r.Revision++
	r.wave(w)
	return nil
}
func (r *Run) View(w *engine.World) View {
	v := View{Stage: r.Stage, Phase: r.Phase, Reason: r.Reason, Revision: r.Revision, TargetMass: targets[r.Stage-1], Mass: w.Mass(r.Player), SeedLabel: fmt.Sprintf("%06X", r.Seed&0xFFFFFF), Ranks: map[string]uint8{}, Offers: []Offer{}}
	if w.Tick < r.Deadline {
		v.SecondsLeft = int((r.Deadline - w.Tick + 29) / engine.TickRate)
	}
	if p := w.Player(r.Player); p != nil {
		v.Peak = p.RunPeak
		v.Kills = p.RunKills
		for i, id := range mutations {
			v.Ranks[id] = p.Mutations[i]
		}
	}
	if r.Phase == "choice" {
		v.Offers = []Offer{{"cilia", "Cílios", "+10% de velocidade de movimento."}, {"digestion", "Metabolismo", "+20% de massa recebida dos alimentos."}, {"cohesion", "Coesão", "Recombinação 20% mais cedo após dividir."}}
	}
	return v
}
func (r *Run) Validate(w *engine.World) error {
	if w.Config.Mode != engine.Rogue || r.Stage < 1 || r.Stage > 3 || r.Revision < 1 || w.Player(r.Player) == nil {
		return errors.New("invalid expedition checkpoint")
	}
	p := w.Player(r.Player)
	if p.Bot || int(p.Mutations[0])+int(p.Mutations[1])+int(p.Mutations[2]) != r.Stage-1 || r.Phase == "choice" && r.Stage == 3 || r.Phase == "victory" && r.Stage != 3 {
		return errors.New("invalid expedition progression")
	}
	switch r.Phase {
	case "running", "choice", "fallen", "failed", "victory":
		return nil
	}
	return errors.New("invalid expedition phase")
}

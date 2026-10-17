// Package engine owns the authoritative, fixed-step simulation. It has no I/O.
package engine

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

type Kind uint8

const (
	Food Kind = iota + 1
	Cell
	Ejected
	Virus
)

const (
	Classic                  = "classic"
	Teams                    = "teams"
	Arena                    = "arena"
	Rogue                    = "rogue"
	TickRate                 = 30
	MaxCells                 = 16
	MaxEjected               = 512
	MaxViruses               = 128
	SplitCooldownTicks       = 6
	EjectCooldownTicks       = 3
	SpawnMass                = 36.0
	EjectCost                = 14.0
	EjectMass                = 12.0
	Split              uint8 = 1
	Eject              uint8 = 2
)

type Config struct {
	Size          float64
	Food, Viruses int
	Mode          string
}

func DefaultConfig(mode string) Config {
	if mode == Rogue {
		return Config{Size: 3200, Food: 1400, Viruses: 10, Mode: Rogue}
	}
	return Config{Size: 6000, Food: 3600, Viruses: 24, Mode: mode}
}

type Entity struct {
	ID, Owner          uint32
	Kind               Kind
	X, Y, Mass, VX, VY float64
	MergeAt, Born      uint64
	Team               uint8
	Feeds              int
}

type Player struct {
	ID                       uint32
	Name, Profile            string
	Bot, Alive               bool
	Team                     uint8
	Kills, Deaths            int
	Peak                     float64
	HasRun                   bool
	RunBorn                  uint64
	RunPeak                  float64
	RunKills                 int
	Mutations                [3]uint8
	TargetX, TargetY         float64
	Actions                  uint8
	SplitAt, EjectAt, DiedAt uint64
}

// Every source/sink is recorded, including replenishment and decay. A checkpoint
// contains the PRNG state, so replay does not depend on wall time or Go's RNG.
type World struct {
	view           *viewCache
	Config         Config
	Entities       []Entity
	Players        []Player
	NextID         uint32
	Tick, RNG      uint64
	Sources, Sinks float64
}

func New(seed uint64, cfg Config) *World {
	if cfg.Size < 800 || cfg.Size > 20000 || !finite(cfg.Size) {
		cfg.Size = 6000
	}
	if cfg.Mode != Teams && cfg.Mode != Arena && cfg.Mode != Rogue {
		cfg.Mode = Classic
	}
	cfg.Food = max(0, min(cfg.Food, 12000))
	cfg.Viruses = max(0, min(cfg.Viruses, 128))
	w := &World{Config: cfg, RNG: seed ^ 0x9e3779b97f4a7c15}
	if w.RNG == 0 {
		w.RNG = 0xd1b54a32d192ed03
	}
	w.replenish()
	return w
}

func Radius(mass float64) float64 { return 4 * math.Sqrt(max(0, mass)) }
func finite(v float64) bool       { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func (w *World) random() float64 {
	x := w.RNG
	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	w.RNG = x
	return float64((x*2685821657736338717)>>11) / (1 << 53)
}
func (w *World) next() uint32 { w.NextID++; return w.NextID }
func (w *World) add(e Entity) {
	w.view = nil
	e.ID = w.next()
	e.Born = w.Tick
	w.Entities = append(w.Entities, e)
	w.Sources += e.Mass
}
func (w *World) Player(id uint32) *Player {
	for i := range w.Players {
		if w.Players[i].ID == id {
			return &w.Players[i]
		}
	}
	return nil
}
func (w *World) Cells(id uint32) []Entity {
	var result []Entity
	for _, e := range w.Entities {
		if e.Kind == Cell && e.Owner == id && e.Mass > 0 {
			result = append(result, e)
		}
	}
	return result
}
func (w *World) Mass(id uint32) float64 {
	total := 0.0
	for _, e := range w.Entities {
		if e.Kind == Cell && e.Owner == id {
			total += e.Mass
		}
	}
	return total
}
func (w *World) TotalMass() float64 {
	total := 0.0
	for _, e := range w.Entities {
		total += e.Mass
	}
	return total
}
func (w *World) AddPlayer(name string, bot bool, profile string) uint32 {
	id := w.next()
	counts := [3]int{}
	for _, p := range w.Players {
		if p.Alive {
			counts[p.Team%3]++
		}
	}
	team := uint8(0)
	for i := 1; i < 3; i++ {
		if counts[i] < counts[team] {
			team = uint8(i)
		}
	}
	w.Players = append(w.Players, Player{ID: id, Name: name, Bot: bot, Profile: profile, Team: team})
	w.spawn(id)
	return id
}
func (w *World) spawn(id uint32) {
	// Select the safest of a bounded set; no spawn immunity or hidden power.
	x, y, best := w.Config.Size/2, w.Config.Size/2, -math.MaxFloat64
	for range 24 {
		cx, cy := 40+w.random()*(w.Config.Size-80), 40+w.random()*(w.Config.Size-80)
		safety := w.Config.Size
		for _, e := range w.Entities {
			if e.Kind == Cell && e.Mass > SpawnMass*1.18 {
				safety = min(safety, math.Hypot(e.X-cx, e.Y-cy)-Radius(e.Mass))
			}
		}
		if safety > best {
			x, y, best = cx, cy, safety
		}
	}
	p := w.Player(id)
	p.Alive = true
	p.HasRun = true
	p.RunBorn = w.Tick
	p.RunPeak = SpawnMass
	p.Peak = max(p.Peak, SpawnMass)
	p.RunKills = 0
	p.Actions = 0
	p.TargetX = x
	p.TargetY = y
	p.SplitAt = w.Tick
	p.EjectAt = w.Tick
	w.add(Entity{Owner: id, Kind: Cell, X: x, Y: y, Mass: SpawnMass, Team: p.Team})
}
func (w *World) Respawn(id uint32) error {
	p := w.Player(id)
	if p == nil || p.Alive || w.Tick < p.DiedAt+60 {
		return errors.New("respawn unavailable")
	}
	w.spawn(id)
	return nil
}

func (w *World) Spectate(id uint32) {
	w.view = nil
	for i := range w.Entities {
		if w.Entities[i].Kind == Cell && w.Entities[i].Owner == id {
			w.Sinks += w.Entities[i].Mass
			w.Entities[i].Mass = 0
		}
	}
	w.compact()
	if p := w.Player(id); p != nil {
		p.Alive = false
		p.DiedAt = w.Tick
		p.HasRun = false
	}
}
func (w *World) RemovePlayer(id uint32) {
	w.view = nil
	for i := range w.Entities {
		if w.Entities[i].Kind == Cell && w.Entities[i].Owner == id {
			w.Sinks += w.Entities[i].Mass
			w.Entities[i].Mass = 0
		}
	}
	w.compact()
	for i, p := range w.Players {
		if p.ID == id {
			w.Players = append(w.Players[:i], w.Players[i+1:]...)
			break
		}
	}
}
func (w *World) Input(id uint32, x, y float64, actions uint8) error {
	p := w.Player(id)
	if p == nil || !p.Alive || !finite(x) || !finite(y) || actions & ^uint8(Split|Eject) != 0 {
		return errors.New("invalid input")
	}
	p.TargetX = max(0, min(w.Config.Size, x))
	p.TargetY = max(0, min(w.Config.Size, y))
	p.Actions |= actions
	return nil
}
func direction(x, y float64) (float64, float64) {
	d := math.Hypot(x, y)
	if d < .001 {
		return 1, 0
	}
	return x / d, y / d
}
func (w *World) clamp(e *Entity) {
	// A finite map has a finite per-cell capacity; excess is an explicit sink.
	cap := math.Pow(w.Config.Size/9, 2)
	if e.Mass > cap {
		w.Sinks += e.Mass - cap
		e.Mass = cap
	}
	r := Radius(e.Mass)
	e.X = max(r, min(w.Config.Size-r, e.X))
	e.Y = max(r, min(w.Config.Size-r, e.Y))
}
func (w *World) actions(p *Player) {
	if !p.Alive {
		p.Actions = 0
		return
	}
	pending := p.Actions
	n := len(w.Cells(p.ID))
	original := len(w.Entities)
	if p.Actions&Split != 0 && w.Tick >= p.SplitAt {
		pending &^= Split
		did := false
		for i := 0; i < original && n < MaxCells; i++ {
			e := &w.Entities[i]
			if e.Kind != Cell || e.Owner != p.ID || e.Mass < 36 {
				continue
			}
			dx, dy := direction(p.TargetX-e.X, p.TargetY-e.Y)
			e.Mass /= 2
			e.MergeAt = w.Tick + 360 + uint64(e.Mass*.3)
			if w.Config.Mode == Rogue && !p.Bot {
				e.MergeAt = w.Tick + uint64(float64(e.MergeAt-w.Tick)*(1-.2*float64(p.Mutations[2])))
			}
			child := *e
			child.ID = w.next()
			child.Born = w.Tick
			child.X += dx * (Radius(e.Mass)*2 + 8)
			child.Y += dy * (Radius(e.Mass)*2 + 8)
			child.VX = dx * 1050
			e.VX -= dx * 70
			child.VY = dy * 1050
			e.VY -= dy * 70
			w.clamp(&child)
			w.Entities = append(w.Entities, child)
			n++
			did = true
		}
		if did {
			p.SplitAt = w.Tick + SplitCooldownTicks
		}
	}
	if p.Actions&Eject != 0 && w.Tick >= p.EjectAt {
		pending &^= Eject
		did := false
		pellets := 0
		for _, e := range w.Entities {
			if e.Kind == Ejected && e.Mass > 0 {
				pellets++
			}
		}
		for i := 0; i < original; i++ {
			if pellets >= MaxEjected {
				break
			}
			e := &w.Entities[i]
			if e.Kind != Cell || e.Owner != p.ID || e.Mass < 32 {
				continue
			}
			e.Mass -= EjectCost
			w.Sinks += EjectCost - EjectMass
			dx, dy := direction(p.TargetX-e.X, p.TargetY-e.Y)
			distance := Radius(e.Mass) + Radius(EjectMass) + 3
			pellet := Entity{ID: w.next(), Owner: p.ID, Kind: Ejected, Team: p.Team, Mass: EjectMass, X: e.X + dx*distance, Y: e.Y + dy*distance, VX: dx * 650, VY: dy * 650, Born: w.Tick}
			w.clamp(&pellet)
			w.Entities = append(w.Entities, pellet)
			pellets++
			did = true
		}
		if did {
			p.EjectAt = w.Tick + EjectCooldownTicks
		}
	}
	p.Actions = pending
}

// Step advances exactly 1/30 second. Callers serialize access to a World.
func (w *World) Step() {
	w.view = nil
	w.Tick++
	for i := range w.Players {
		w.actions(&w.Players[i])
	}
	for i := range w.Entities {
		e := &w.Entities[i]
		if e.Kind == Cell {
			p := w.Player(e.Owner)
			if p != nil {
				dx, dy := p.TargetX-e.X, p.TargetY-e.Y
				d := math.Hypot(dx, dy)
				speed := 220 * math.Pow(e.Mass/SpawnMass, -.25)
				speed *= min(1, d/(Radius(e.Mass)*2.5))
				if w.Config.Mode == Rogue && !p.Bot {
					speed *= 1 + .1*float64(p.Mutations[0])
				}
				if d > .01 {
					move := min(d, speed/TickRate)
					e.X += dx / d * move
					e.Y += dy / d * move
				}
			}
			if w.Tick%30 == 0 && e.Mass > SpawnMass {
				lost := (e.Mass - SpawnMass) * .0015
				e.Mass -= lost
				w.Sinks += lost
			}
		}
		e.X += e.VX / TickRate
		e.Y += e.VY / TickRate
		drag := .88
		if e.Kind == Cell {
			drag = .92
		} else if e.Kind == Ejected {
			drag = .9
		}
		e.VX *= drag
		e.VY *= drag
		if e.Kind == Ejected && w.Tick > e.Born+900 {
			w.Sinks += e.Mass
			e.Mass = 0
		}
		w.clamp(e)
	}
	w.merge()
	w.consume()
	w.compact()
	w.replenish()
	masses := make(map[uint32]float64, len(w.Players))
	for _, e := range w.Entities {
		if e.Kind == Cell {
			masses[e.Owner] += e.Mass
		}
	}
	for i := range w.Players {
		p := &w.Players[i]
		mass := masses[p.ID]
		p.Peak = max(p.Peak, mass)
		p.RunPeak = max(p.RunPeak, mass)
		if p.Alive && mass == 0 {
			p.Alive = false
			p.Deaths++
			p.DiedAt = w.Tick
		}
	}
}
func (w *World) merge() {
	for i := range w.Entities {
		a := &w.Entities[i]
		if a.Kind != Cell || a.Mass == 0 {
			continue
		}
		for j := i + 1; j < len(w.Entities); j++ {
			b := &w.Entities[j]
			if b.Kind != Cell || b.Owner != a.Owner || b.Mass == 0 {
				continue
			}
			dx, dy := b.X-a.X, b.Y-a.Y
			d := math.Hypot(dx, dy)
			sum := Radius(a.Mass) + Radius(b.Mass)
			if w.Tick >= max(a.MergeAt, b.MergeAt) && d < sum*.65 {
				a.X = (a.X*a.Mass + b.X*b.Mass) / (a.Mass + b.Mass)
				a.Y = (a.Y*a.Mass + b.Y*b.Mass) / (a.Mass + b.Mass)
				a.Mass += b.Mass
				b.Mass = 0
				w.clamp(a)
			} else if d < sum*.75 && w.Tick > a.Born+15 && w.Tick > b.Born+15 && w.Tick < max(a.MergeAt, b.MergeAt) {
				if d < .01 {
					dx = math.Cos(float64(a.ID + b.ID))
					dy = math.Sin(float64(a.ID + b.ID))
					d = 1
				}
				push := (sum*.75 - d) * .18
				a.X -= dx / d * push
				a.Y -= dy / d * push
				b.X += dx / d * push
				b.Y += dy / d * push
				w.clamp(a)
				w.clamp(b)
			}
		}
	}
}

// Uniform buckets are rebuilt once per collision pass. Candidate iteration is
// ordered by coordinates and insertion order, never by Go map iteration.
type bucket struct{ x, y int }
type grid map[bucket][]int

func buildGrid(entities []Entity) grid {
	g := make(grid, len(entities)/4)
	for i, e := range entities {
		if e.Mass > 0 {
			key := bucket{int(e.X / 128), int(e.Y / 128)}
			g[key] = append(g[key], i)
		}
	}
	return g
}
func (g grid) near(x, y, r float64) []int {
	var out []int
	for bx := int(math.Floor((x - r) / 128)); bx <= int((x+r)/128); bx++ {
		for by := int(math.Floor((y - r) / 128)); by <= int((y+r)/128); by++ {
			out = append(out, g[bucket{bx, by}]...)
		}
	}
	return out
}
func (w *World) consume() {
	g := buildGrid(w.Entities)
	// Feed viruses before the cell pass, keeping mass transfers atomic.
	original := len(w.Entities)
	for i := 0; i < original; i++ {
		if w.Entities[i].Kind != Virus {
			continue
		}
		v := &w.Entities[i]
		for _, j := range g.near(v.X, v.Y, Radius(v.Mass)+20) {
			v = &w.Entities[i]
			e := &w.Entities[j]
			if e.Kind != Ejected || e.Mass == 0 || math.Hypot(v.X-e.X, v.Y-e.Y) > Radius(v.Mass) {
				continue
			}
			v.Mass += e.Mass
			v.Feeds++
			dx, dy := direction(e.VX, e.VY)
			e.Mass = 0
			if v.Feeds >= 7 {
				remainder := v.Mass - 100
				v.Mass = 100
				v.Feeds = 0
				viruses := 0
				for _, entity := range w.Entities {
					if entity.Kind == Virus && entity.Mass > 0 {
						viruses++
					}
				}
				if viruses >= MaxViruses {
					w.Sinks += remainder
					continue
				}
				child := Entity{ID: w.next(), Kind: Virus, X: v.X + dx*35, Y: v.Y + dy*35, Mass: remainder, VX: dx * 350, VY: dy * 350, Born: w.Tick}
				w.clamp(&child)
				w.Entities = append(w.Entities, child)
			}
		}
	}
	indices := make([]int, 0)
	for i, e := range w.Entities {
		if e.Kind == Cell && e.Mass > 0 {
			indices = append(indices, i)
		}
	}
	sort.SliceStable(indices, func(i, j int) bool {
		a, b := w.Entities[indices[i]], w.Entities[indices[j]]
		if a.Mass == b.Mass {
			return a.ID < b.ID
		}
		return a.Mass > b.Mass
	})
	for _, i := range indices {
		a := &w.Entities[i]
		if a.Mass == 0 {
			continue
		}
		for _, j := range g.near(a.X, a.Y, Radius(a.Mass)) {
			a = &w.Entities[i]
			b := &w.Entities[j]
			if i == j || b.Mass == 0 || a.Mass == 0 {
				continue
			}
			if b.Kind == Cell && (a.Owner == b.Owner || w.Config.Mode == Teams && a.Team == b.Team || a.Mass < b.Mass*1.18) {
				continue
			}
			if b.Kind == Ejected && a.Owner == b.Owner && w.Tick < b.Born+15 {
				continue
			}
			if b.Kind == Virus && a.Mass < b.Mass*1.3 {
				continue
			}
			if math.Hypot(a.X-b.X, a.Y-b.Y) > Radius(a.Mass)-Radius(b.Mass)*.35 {
				continue
			}
			kind, owner := b.Kind, b.Owner
			if kind == Food && w.Config.Mode == Rogue {
				if p := w.Player(a.Owner); p != nil && !p.Bot {
					bonus := b.Mass * .2 * float64(p.Mutations[1])
					a.Mass += bonus
					w.Sources += bonus
				}
			}
			a.Mass += b.Mass
			b.Mass = 0
			if kind == Cell && w.Mass(owner) == 0 {
				if p := w.Player(a.Owner); p != nil {
					p.Kills++
					p.RunKills++
				}
			}
			if kind == Virus {
				w.burst(i)
			}
			w.clamp(&w.Entities[i])
		}
	}
}
func (w *World) burst(index int) {
	a := w.Entities[index]
	available := MaxCells - len(w.Cells(a.Owner))
	n := min(available+1, int(a.Mass/18))
	if n < 2 {
		return
	}
	part := a.Mass / float64(n)
	w.Entities[index].Mass = part
	w.Entities[index].MergeAt = w.Tick + 420
	for i := 1; i < n; i++ {
		angle := float64(i) * 2 * math.Pi / float64(n)
		child := a
		child.ID = w.next()
		child.Mass = part
		child.MergeAt = w.Tick + 420
		child.Born = w.Tick
		child.X += math.Cos(angle) * Radius(part)
		child.Y += math.Sin(angle) * Radius(part)
		child.VX = math.Cos(angle) * 420
		child.VY = math.Sin(angle) * 420
		w.clamp(&child)
		w.Entities = append(w.Entities, child)
	}
}
func (w *World) compact() {
	n := 0
	for _, e := range w.Entities {
		if e.Mass > 0 {
			w.Entities[n] = e
			n++
		}
	}
	w.Entities = w.Entities[:n]
}
func (w *World) replenish() {
	foods, viruses := 0, 0
	for _, e := range w.Entities {
		if e.Kind == Food {
			foods++
		}
		if e.Kind == Virus {
			viruses++
		}
	}
	for i := foods; i < w.Config.Food; i++ {
		mass := 1.
		roll := w.random()
		if roll > .95 {
			mass = 3
		} else if roll > .7 {
			mass = 2
		}
		w.add(Entity{Kind: Food, X: 8 + w.random()*(w.Config.Size-16), Y: 8 + w.random()*(w.Config.Size-16), Mass: mass})
	}
	for i := viruses; i < w.Config.Viruses; i++ {
		w.add(Entity{Kind: Virus, X: 45 + w.random()*(w.Config.Size-90), Y: 45 + w.random()*(w.Config.Size-90), Mass: 100})
	}
}

func (w *World) Valid() error {
	if !finite(w.Sources) || !finite(w.Sinks) || w.Config.Size < 800 {
		return errors.New("invalid world")
	}
	ids := make(map[uint32]bool, len(w.Entities))
	for _, p := range w.Players {
		for _, rank := range p.Mutations {
			if rank > 3 || w.Config.Mode != Rogue && rank != 0 {
				return errors.New("invalid mutation rank/mode")
			}
		}
	}
	counts := map[uint32]int{}
	for _, e := range w.Entities {
		if e.ID == 0 || ids[e.ID] || e.Kind < Food || e.Kind > Virus || !finite(e.X) || !finite(e.Y) || !finite(e.Mass) || !finite(e.VX) || !finite(e.VY) || e.Mass <= 0 {
			return fmt.Errorf("invalid entity %d", e.ID)
		}
		ids[e.ID] = true
		r := Radius(e.Mass)
		if e.X < r-.001 || e.Y < r-.001 || e.X > w.Config.Size-r+.001 || e.Y > w.Config.Size-r+.001 {
			return fmt.Errorf("entity %d outside map", e.ID)
		}
		if e.Kind == Cell {
			counts[e.Owner]++
			if counts[e.Owner] > MaxCells || w.Player(e.Owner) == nil {
				return errors.New("invalid cell owner/count")
			}
		}
	}
	if math.Abs(w.TotalMass()+w.Sinks-w.Sources) > max(.0001, w.Sources*1e-9) {
		return errors.New("mass ledger mismatch")
	}
	return nil
}

func (w *World) ActionOccurred() bool {
	for _, p := range w.Players {
		if p.SplitAt == w.Tick+SplitCooldownTicks || p.EjectAt == w.Tick+EjectCooldownTicks {
			return true
		}
	}
	return false
}

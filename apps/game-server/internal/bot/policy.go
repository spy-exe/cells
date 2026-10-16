// Package bot implements visible-information utility policies, not learned models.
package bot

import (
	"github.com/spy-exe/cells/apps/game-server/internal/engine"
	"math"
)

var Profiles = []string{"aggressive", "cautious", "opportunist", "balanced", "beginner"}

type Track struct {
	X, Y, VX, VY float64
	Tick         uint64
}
type Memory struct{ Seen map[uint32]Track }
type Intent struct {
	X, Y    float64
	Actions uint8
	Reason  string
}

func ValidProfile(p string) bool {
	for _, v := range Profiles {
		if v == p {
			return true
		}
	}
	return false
}
func dist(a, b engine.Visible) float64 { return math.Hypot(a.X-b.X, a.Y-b.Y) }
func (m *Memory) remember(o engine.Observation) {
	if m.Seen == nil {
		m.Seen = map[uint32]Track{}
	}
	for _, e := range o.Entities {
		if e.Kind != engine.Cell {
			continue
		}
		t := Track{X: e.X, Y: e.Y, Tick: o.Tick}
		if old, ok := m.Seen[e.ID]; ok && o.Tick > old.Tick {
			dt := float64(o.Tick-old.Tick) / 30
			t.VX = (e.X - old.X) / dt
			t.VY = (e.Y - old.Y) / dt
		}
		m.Seen[e.ID] = t
	}
	for id, t := range m.Seen {
		if o.Tick > t.Tick+120 {
			delete(m.Seen, id)
		}
	}
}
func pointLine(x, y, ax, ay, bx, by float64) float64 {
	dx, dy := bx-ax, by-ay
	denom := dx*dx + dy*dy
	if denom < .001 {
		return math.Hypot(x-ax, y-ay)
	}
	t := math.Max(0, math.Min(1, ((x-ax)*dx+(y-ay)*dy)/denom))
	return math.Hypot(x-ax-t*dx, y-ay-t*dy)
}
func Decide(o engine.Observation, m *Memory, profile string) Intent {
	idle := Intent{X: o.X, Y: o.Y, Reason: "observe"}
	if !o.Alive {
		return idle
	}
	m.remember(o)
	aggression, caution := 1., 1.5
	switch profile {
	case "aggressive":
		aggression = 1.6
		caution = 1.05
	case "cautious":
		aggression = .55
		caution = 2.1
	case "opportunist":
		aggression = 1.3
		caution = 1.6
	case "beginner":
		aggression = .35
		caution = .8
	}
	big := engine.Visible{Mass: 0}
	smallest := math.MaxFloat64
	own := 0
	for _, e := range o.Entities {
		if e.Kind == engine.Cell && e.Owner == o.Self {
			own++
			smallest = math.Min(smallest, e.Mass)
			if e.Mass > big.Mass {
				big = e
			}
		}
	}
	if big.Mass == 0 {
		return idle
	}
	var enemies, threats, viruses []engine.Visible
	danger := false
	for _, e := range o.Entities {
		if e.Kind == engine.Virus {
			viruses = append(viruses, e)
		}
		if e.Kind != engine.Cell || e.Owner == o.Self || o.Mode == engine.Teams && e.Team == big.Team {
			continue
		}
		enemies = append(enemies, e)
		if e.Mass > smallest*1.18 {
			threats = append(threats, e)
			reach := engine.Radius(e.Mass) + engine.Radius(smallest) + 190*caution
			if e.Mass > big.Mass*2.4 {
				reach += 420
			}
			if dist(big, e) < reach {
				danger = true
			}
		}
	}
	if danger {
		bestScore := -math.MaxFloat64
		best := idle
		best.Reason = "flee"
		for i := 0; i < 24; i++ {
			angle := float64(i) * math.Pi / 12
			x, y := o.X+math.Cos(angle)*360, o.Y+math.Sin(angle)*360
			score := 0.
			margin := engine.Radius(big.Mass) + 30
			if x < margin || y < margin || x > o.Size-margin || y > o.Size-margin {
				score -= 10000
			}
			for _, e := range threats {
				track := m.Seen[e.ID]
				px, py := e.X+track.VX*.5, e.Y+track.VY*.5
				d := math.Hypot(x-px, y-py) - engine.Radius(e.Mass)
				score -= 160000 * math.Sqrt(e.Mass/big.Mass) / math.Max(20, d)
			}
			for _, v := range viruses {
				d := pointLine(v.X, v.Y, o.X, o.Y, x, y)
				if big.Mass > v.Mass*1.3 && d < engine.Radius(big.Mass)+engine.Radius(v.Mass)+10 {
					score -= 3000
				}
				// A small body can shelter behind a virus that bursts its pursuer.
				if big.Mass < v.Mass*1.3 {
					for _, e := range threats {
						if pointLine(v.X, v.Y, x, y, e.X, e.Y) < engine.Radius(v.Mass) && math.Hypot(x-v.X, y-v.Y) < math.Hypot(x-e.X, y-e.Y) {
							score += 40
						}
					}
				}
			}
			if score > bestScore {
				bestScore = score
				best.X = x
				best.Y = y
			}
		}
		best.X = math.Max(0, math.Min(o.Size, best.X))
		best.Y = math.Max(0, math.Min(o.Size, best.Y))
		return best
	}
	bestScore := 0.
	best := idle
	for _, e := range o.Entities {
		if e.Kind != engine.Food && e.Kind != engine.Ejected {
			continue
		}
		d := dist(big, e)
		score := e.Mass / (d + 18)
		for _, enemy := range threats {
			if dist(e, enemy) < engine.Radius(enemy.Mass)+170*caution {
				score *= .05
			}
		}
		if score > bestScore {
			bestScore = score
			best = Intent{X: e.X, Y: e.Y, Reason: "forage"}
		}
	}
	for _, e := range enemies {
		if big.Mass < e.Mass*1.25 {
			continue
		}
		d := dist(big, e)
		score := aggression * e.Mass / (d + 80) * .12
		mySpeed := 220 * math.Pow(big.Mass/36, -.25)
		preySpeed := 220 * math.Pow(e.Mass/36, -.25)
		if preySpeed > mySpeed && d > engine.Radius(big.Mass)+260 {
			score *= .25
		}
		unsafe := false
		for _, threat := range threats {
			if dist(e, threat) < engine.Radius(threat.Mass)+230 {
				unsafe = true
			}
		}
		if unsafe {
			score *= .02
		}
		if score <= bestScore {
			continue
		}
		bestScore = score
		t := m.Seen[e.ID]
		best = Intent{X: e.X + t.VX*.3, Y: e.Y + t.VY*.3, Reason: "hunt"}
		if profile != "beginner" && profile != "cautious" && !unsafe && o.SplitIn == 0 && own < 8 && big.Mass/2 > e.Mass*1.35 && d < engine.Radius(big.Mass)+260 && d > engine.Radius(big.Mass)*.7 {
			safe := true
			for _, threat := range threats {
				if dist(e, threat) < 600*caution {
					safe = false
				}
			}
			for _, v := range viruses {
				if pointLine(v.X, v.Y, big.X, big.Y, best.X, best.Y) < engine.Radius(v.Mass)+engine.Radius(big.Mass/2) {
					safe = false
				}
			}
			if safe {
				best.Actions = engine.Split
				best.Reason = "split-hunt"
			}
		}
	}
	if best.Reason == "observe" {
		a := float64((o.Tick/180+uint64(o.Self)*37)%360) * math.Pi / 180
		best = Intent{X: o.Size/2 + math.Cos(a)*o.Size*.3, Y: o.Size/2 + math.Sin(a)*o.Size*.3, Reason: "explore"}
	}
	if o.EjectIn == 0 && big.Mass > 150 && profile != "beginner" {
		for _, v := range viruses {
			d := dist(big, v)
			if d < 120 || d > 260 {
				continue
			}
			for _, e := range enemies {
				if e.Mass > big.Mass*1.3 && dist(big, e) > 500 && dist(v, e) < 420 && pointLine(v.X, v.Y, big.X, big.Y, e.X, e.Y) < 30 {
					best = Intent{X: v.X, Y: v.Y, Actions: engine.Eject, Reason: "feed-virus"}
				}
			}
		}
	}
	// Steer around a dangerous virus on the chosen path, rather than simply
	// subtracting its utility and then walking through it to reach food.
	for _, v := range viruses {
		if best.Reason != "feed-virus" && big.Mass >= v.Mass*1.3 && pointLine(v.X, v.Y, big.X, big.Y, best.X, best.Y) < engine.Radius(big.Mass)+engine.Radius(v.Mass)+10 {
			dx, dy := v.X-big.X, v.Y-big.Y
			d := math.Max(1, math.Hypot(dx, dy))
			best.X = big.X - dy/d*240
			best.Y = big.Y + dx/d*240
			best.Actions = 0
			best.Reason = "avoid-virus"
		}
	}
	best.X = math.Max(0, math.Min(o.Size, best.X))
	best.Y = math.Max(0, math.Min(o.Size, best.Y))
	return best
}

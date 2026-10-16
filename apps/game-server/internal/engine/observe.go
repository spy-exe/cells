package engine

import (
	"math"
	"sort"
)

type Visible struct {
	ID, Owner  uint32
	Kind       Kind
	X, Y, Mass float64
	Name       string
	Bot        bool
	Team       uint8
}
type Rank struct {
	ID   uint32
	Name string
	Mass float64
	Bot  bool
	Team uint8
}
type Observation struct {
	Tick                        uint64
	Self                        uint32
	Alive                       bool
	Mode                        string
	Size, X, Y, View, Mass      float64
	SplitIn, EjectIn, RespawnIn uint16
	Entities                    []Visible
	Leaderboard                 []Rank
	TeamMass                    [3]float64
}

func (w *World) Observe(id uint32) Observation {
	o := Observation{Tick: w.Tick, Self: id, Mode: w.Config.Mode, Size: w.Config.Size, X: w.Config.Size / 2, Y: w.Config.Size / 2, View: 800}
	if p := w.Player(id); p != nil {
		o.Alive = p.Alive
		o.Mass = w.Mass(id)
		if p.SplitAt > w.Tick {
			o.SplitIn = uint16(p.SplitAt - w.Tick)
		}
		if p.EjectAt > w.Tick {
			o.EjectIn = uint16(p.EjectAt - w.Tick)
		}
		if !p.Alive && p.DiedAt+60 > w.Tick {
			o.RespawnIn = uint16(p.DiedAt + 60 - w.Tick)
		}
		if o.Mass > 0 {
			o.X = 0
			o.Y = 0
			for _, e := range w.Cells(id) {
				o.X += e.X * e.Mass / o.Mass
				o.Y += e.Y * e.Mass / o.Mass
			}
			o.View = min(2200, 800+math.Sqrt(o.Mass)*10)
		}
	}
	for _, p := range w.Players {
		if p.Alive {
			mass := w.Mass(p.ID)
			o.Leaderboard = append(o.Leaderboard, Rank{ID: p.ID, Name: p.Name, Mass: mass, Bot: p.Bot, Team: p.Team})
			o.TeamMass[p.Team%3] += mass
		}
	}
	sort.Slice(o.Leaderboard, func(i, j int) bool {
		a, b := o.Leaderboard[i], o.Leaderboard[j]
		if a.Mass == b.Mass {
			return a.ID < b.ID
		}
		return a.Mass > b.Mass
	})
	// Spectators follow the leader with the same visibility constraints.
	if !o.Alive && len(o.Leaderboard) > 0 {
		leader := o.Leaderboard[0].ID
		mass := w.Mass(leader)
		o.X = 0
		o.Y = 0
		for _, e := range w.Cells(leader) {
			o.X += e.X * e.Mass / mass
			o.Y += e.Y * e.Mass / mass
		}
		o.View = min(2200, 800+math.Sqrt(mass)*10)
	}
	if len(o.Leaderboard) > 10 {
		o.Leaderboard = o.Leaderboard[:10]
	}
	// Important entities precede particles when the view reaches its hard cap.
	for _, kind := range []Kind{Cell, Virus, Ejected, Food} {
		for _, e := range w.Entities {
			if e.Kind != kind || len(o.Entities) >= 2048 || e.Mass <= 0 {
				continue
			}
			if !(e.Kind == Cell && e.Owner == id) && math.Hypot(e.X-o.X, e.Y-o.Y) > o.View+Radius(e.Mass) {
				continue
			}
			v := Visible{ID: e.ID, Owner: e.Owner, Kind: e.Kind, X: e.X, Y: e.Y, Mass: e.Mass, Team: e.Team}
			if e.Kind == Cell {
				if p := w.Player(e.Owner); p != nil {
					v.Name = p.Name
					v.Bot = p.Bot
				}
			}
			o.Entities = append(o.Entities, v)
		}
	}
	return o
}

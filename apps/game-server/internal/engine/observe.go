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

type center struct{ Mass, X, Y float64 }
type viewCache struct {
	tick    uint64
	grid    grid
	players map[uint32]*Player
	centers map[uint32]center
	cells   []int
	ranks   []Rank
	teams   [3]float64
}

func (w *World) visibilityCache() *viewCache {
	if w.view != nil && w.view.tick == w.Tick {
		return w.view
	}
	c := &viewCache{tick: w.Tick, grid: buildGrid(w.Entities), players: make(map[uint32]*Player, len(w.Players)), centers: make(map[uint32]center, len(w.Players))}
	for i, e := range w.Entities {
		if e.Kind == Cell && e.Mass > 0 {
			v := c.centers[e.Owner]
			v.Mass += e.Mass
			v.X += e.X * e.Mass
			v.Y += e.Y * e.Mass
			c.centers[e.Owner] = v
			c.cells = append(c.cells, i)
		}
	}
	for i := range w.Players {
		p := &w.Players[i]
		c.players[p.ID] = p
		if p.Alive {
			v := c.centers[p.ID]
			c.ranks = append(c.ranks, Rank{ID: p.ID, Name: p.Name, Mass: v.Mass, Bot: p.Bot, Team: p.Team})
			c.teams[p.Team%3] += v.Mass
		}
	}
	sort.Slice(c.ranks, func(i, j int) bool {
		a, b := c.ranks[i], c.ranks[j]
		if a.Mass == b.Mass {
			return a.ID < b.ID
		}
		return a.Mass > b.Mass
	})
	w.view = c
	return c
}
func (w *World) Observe(id uint32) Observation {
	c := w.visibilityCache()
	o := Observation{Tick: w.Tick, Self: id, Mode: w.Config.Mode, Size: w.Config.Size, X: w.Config.Size / 2, Y: w.Config.Size / 2, View: 800, TeamMass: c.teams}
	if p := c.players[id]; p != nil {
		o.Alive = p.Alive
		o.Mass = c.centers[id].Mass
		if p.SplitAt > w.Tick {
			o.SplitIn = uint16(p.SplitAt - w.Tick)
		}
		if p.EjectAt > w.Tick {
			o.EjectIn = uint16(p.EjectAt - w.Tick)
		}
		if !p.Alive && p.DiedAt+60 > w.Tick {
			o.RespawnIn = uint16(p.DiedAt + 60 - w.Tick)
		}
	}
	follow := id
	if !o.Alive && len(c.ranks) > 0 {
		follow = c.ranks[0].ID
	}
	if v := c.centers[follow]; v.Mass > 0 {
		o.X = v.X / v.Mass
		o.Y = v.Y / v.Mass
		o.View = min(2200, 800+math.Sqrt(v.Mass)*10)
	}
	o.Leaderboard = append([]Rank{}, c.ranks[:min(10, len(c.ranks))]...)
	appendVisible := func(e Entity) {
		if len(o.Entities) >= 2048 || e.Mass <= 0 {
			return
		}
		if !(e.Kind == Cell && e.Owner == id) && math.Hypot(e.X-o.X, e.Y-o.Y) > o.View+Radius(e.Mass) {
			return
		}
		v := Visible{ID: e.ID, Owner: e.Owner, Kind: e.Kind, X: e.X, Y: e.Y, Mass: e.Mass, Team: e.Team}
		if p := c.players[e.Owner]; e.Kind == Cell && p != nil {
			v.Name = p.Name
			v.Bot = p.Bot
		}
		o.Entities = append(o.Entities, v)
	}
	// Large cells can overlap the view even with a distant center. Non-cell
	// entities are small and use local buckets; both paths retain visibility rules.
	for _, index := range c.cells {
		appendVisible(w.Entities[index])
	}
	nearby := c.grid.near(o.X, o.Y, o.View+64)
	for _, kind := range []Kind{Virus, Ejected, Food} {
		for _, index := range nearby {
			e := w.Entities[index]
			if e.Kind == kind {
				appendVisible(e)
			}
		}
	}
	return o
}

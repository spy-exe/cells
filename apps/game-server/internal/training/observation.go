package training

import (
	"github.com/spy-exe/cells/apps/game-server/internal/engine"
	"math"
	"sort"
)

func clip(x float64) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return 0
	}
	return math.Max(-1, math.Min(1, x))
}

// Encode has no World pointer: all features come from the human-visible DTO.
func Encode(o engine.Observation) [80]float32 {
	var f [80]float32
	put := func(index int, v float64) { f[index] = float32(clip(v)) }
	view := math.Max(1, o.View)
	size := math.Max(1, o.Size)
	big, small, count := 0., math.MaxFloat64, 0.
	var foods, enemies, viruses []engine.Visible
	for _, e := range o.Entities {
		switch e.Kind {
		case engine.Cell:
			if e.Owner == o.Self {
				big = math.Max(big, e.Mass)
				small = math.Min(small, e.Mass)
				count++
			} else {
				enemies = append(enemies, e)
			}
		case engine.Food, engine.Ejected:
			foods = append(foods, e)
		case engine.Virus:
			viruses = append(viruses, e)
		}
	}
	norm := math.Log1p(100000)
	base := []float64{math.Log1p(o.Mass) / norm, math.Log1p(big) / norm, count / 16, 220 * math.Pow(math.Max(1, big)/36, -.25) / 440, 0, 0, 0, o.X / size, o.Y / size, view / size, engine.Radius(big) / view, math.Min(math.Min(o.X, o.Y), math.Min(size-o.X, size-o.Y)) / view}
	if big >= 36 && count < 16 && o.SplitIn == 0 {
		base[4] = 1
	}
	if big >= 32 && o.EjectIn == 0 {
		base[5] = 1
	}
	if o.Alive {
		base[6] = 1
	}
	for i, v := range base {
		put(i, v)
	}
	nearest := func(items []engine.Visible) {
		sort.Slice(items, func(i, j int) bool {
			a, b := items[i], items[j]
			da, db := math.Hypot(a.X-o.X, a.Y-o.Y), math.Hypot(b.X-o.X, b.Y-o.Y)
			if da == db {
				return a.ID < b.ID
			}
			return da < db
		})
	}
	nearest(foods)
	nearest(enemies)
	nearest(viruses)
	for i, e := range foods[:min(6, len(foods))] {
		n := 12 + i*3
		put(n, (e.X-o.X)/view)
		put(n+1, (e.Y-o.Y)/view)
		put(n+2, e.Mass/14)
	}
	team := uint8(0)
	for _, e := range o.Entities {
		if e.Kind == engine.Cell && e.Owner == o.Self {
			team = e.Team
			break
		}
	}
	for i, e := range enemies[:min(8, len(enemies))] {
		n := 30 + i*5
		put(n, (e.X-o.X)/view)
		put(n+1, (e.Y-o.Y)/view)
		put(n+2, math.Log2(e.Mass/math.Max(1, big))/4)
		if e.Mass > small*1.18 {
			put(n+3, 1)
		}
		if o.Mode == engine.Teams && e.Team == team {
			put(n+4, 1)
		}
	}
	for i, e := range viruses[:min(3, len(viruses))] {
		n := 70 + i*3
		put(n, (e.X-o.X)/view)
		put(n+1, (e.Y-o.Y)/view)
		put(n+2, engine.Radius(e.Mass)/view)
	}
	put(79, float64(len(foods))/128)
	return f
}

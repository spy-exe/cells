// Package protocol implements the bounded, little-endian Cells wire format v1.
package protocol

import (
	"encoding/binary"
	"errors"
	"github.com/spy-exe/cells/apps/game-server/internal/engine"
	"math"
)

const Magic = 206
const Version = 1

type Input struct {
	Kind, Actions uint8
	Seq           uint32
	X, Y          float64
	Echo          []byte
}

func Decode(b []byte) (Input, error) {
	bad := errors.New("invalid protocol frame")
	if len(b) < 3 || b[0] != Magic || b[1] != Version {
		return Input{}, bad
	}
	in := Input{Kind: b[2]}
	switch in.Kind {
	case 1:
		if len(b) != 16 || b[15] > 3 {
			return Input{}, bad
		}
		in.Seq = binary.LittleEndian.Uint32(b[3:])
		in.X = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[7:])))
		in.Y = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[11:])))
		in.Actions = b[15]
		if math.IsNaN(in.X) || math.IsNaN(in.Y) || math.IsInf(in.X, 0) || math.IsInf(in.Y, 0) {
			return Input{}, bad
		}
	case 3, 4:
		if len(b) != 3 {
			return Input{}, bad
		}
	case 5:
		if len(b) != 11 {
			return Input{}, bad
		}
		in.Echo = append([]byte{Magic, Version, 6}, b[3:]...)
	default:
		return Input{}, bad
	}
	return in, nil
}
func putFloat(b []byte, offset int, v float64) {
	binary.LittleEndian.PutUint32(b[offset:], math.Float32bits(float32(v)))
}
func Snapshot(o engine.Observation) []byte {
	b := make([]byte, 53, 53+len(o.Entities)*30)
	b[0] = Magic
	b[1] = Version
	b[2] = 2
	binary.LittleEndian.PutUint32(b[3:], uint32(o.Tick))
	binary.LittleEndian.PutUint32(b[7:], o.Self)
	if o.Alive {
		b[11] |= 1
	}
	if o.Mode == engine.Teams {
		b[11] |= 2
	}
	for i, v := range []float64{o.Size, o.X, o.Y, o.View, o.Mass} {
		putFloat(b, 12+i*4, v)
	}
	binary.LittleEndian.PutUint16(b[32:], o.SplitIn)
	binary.LittleEndian.PutUint16(b[34:], o.EjectIn)
	binary.LittleEndian.PutUint16(b[36:], o.RespawnIn)
	for i, v := range o.TeamMass {
		putFloat(b, 38+i*4, v)
	}
	binary.LittleEndian.PutUint16(b[50:], uint16(len(o.Entities)))
	b[52] = byte(len(o.Leaderboard))
	for _, e := range o.Entities {
		n := []byte(e.Name)
		if len(n) > 80 {
			n = n[:80]
		}
		r := make([]byte, 24)
		binary.LittleEndian.PutUint32(r, e.ID)
		binary.LittleEndian.PutUint32(r[4:], e.Owner)
		putFloat(r, 8, e.X)
		putFloat(r, 12, e.Y)
		putFloat(r, 16, e.Mass)
		r[20] = byte(e.Kind)
		r[21] = e.Team
		if e.Bot {
			r[22] = 1
		}
		r[23] = byte(len(n))
		b = append(b, r...)
		b = append(b, n...)
	}
	for _, p := range o.Leaderboard {
		n := []byte(p.Name)
		if len(n) > 80 {
			n = n[:80]
		}
		r := make([]byte, 11)
		binary.LittleEndian.PutUint32(r, p.ID)
		putFloat(r, 4, p.Mass)
		if p.Bot {
			r[8] = 1
		}
		r[9] = p.Team
		r[10] = byte(len(n))
		b = append(b, r...)
		b = append(b, n...)
	}
	return b
}

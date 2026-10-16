package protocol

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/spy-exe/cells/apps/game-server/internal/engine"
	"math"
	"os"
	"strings"
	"testing"
)

func TestInputValidation(t *testing.T) {
	b := make([]byte, 16)
	b[0] = 206
	b[1] = 1
	b[2] = 1
	binary.LittleEndian.PutUint32(b[3:], 7)
	binary.LittleEndian.PutUint32(b[7:], math.Float32bits(123))
	binary.LittleEndian.PutUint32(b[11:], math.Float32bits(456))
	b[15] = 1
	in, err := Decode(b)
	if err != nil || in.Seq != 7 || in.X != 123 || in.Y != 456 || in.Actions != 1 {
		t.Fatal(in, err)
	}
	for _, bad := range [][]byte{nil, b[:15], append(append([]byte{}, b...), 0), {206, 2, 4}, {206, 1, 99}} {
		if _, err := Decode(bad); err == nil {
			t.Fatal("invalid frame accepted")
		}
	}
	b[15] = 128
	if _, err := Decode(b); err == nil {
		t.Fatal("unknown action")
	}
	b[15] = 0
	binary.LittleEndian.PutUint32(b[7:], math.Float32bits(float32(math.NaN())))
	if _, err := Decode(b); err == nil {
		t.Fatal("NaN accepted")
	}
}
func TestSnapshotLayout(t *testing.T) {
	w := engine.New(1, engine.Config{Size: 1600, Food: 3})
	id := w.AddPlayer("João", true, "cautious")
	o := w.Observe(id)
	b := Snapshot(o)
	if b[0] != 206 || b[1] != 1 || b[2] != 2 || binary.LittleEndian.Uint32(b[7:]) != id || binary.LittleEndian.Uint16(b[50:]) != uint16(len(o.Entities)) {
		t.Fatal("snapshot header")
	}
	if len(b) < 53+len(o.Entities)*24 {
		t.Fatal("truncated snapshot")
	}
}

func TestGoldenSnapshot(t *testing.T) {
	data, err := os.ReadFile("../../../../packages/protocol/fixtures/snapshot-v1.hex")
	if err != nil {
		t.Fatal(err)
	}
	o := engine.Observation{Tick: 30, Self: 7, Alive: true, Size: 1600, X: 800, Y: 800, View: 860, Mass: 36, SplitIn: 24, EjectIn: 6, TeamMass: [3]float64{36, 0, 0}, Entities: []engine.Visible{{ID: 8, Owner: 7, X: 800, Y: 800, Mass: 36, Kind: engine.Cell, Name: "João", Bot: true}}, Leaderboard: []engine.Rank{{ID: 7, Name: "João", Mass: 36, Bot: true}}}
	if got := hex.EncodeToString(Snapshot(o)); got != strings.TrimSpace(string(data)) {
		t.Fatal("cross-language wire contract changed")
	}
}
func FuzzDecode(f *testing.F) {
	f.Add([]byte{206, 1, 3})
	f.Add([]byte{206, 1, 5, 0})
	f.Fuzz(func(t *testing.T, b []byte) { _, _ = Decode(b) })
}

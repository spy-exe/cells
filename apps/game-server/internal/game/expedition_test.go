package game

import (
	"context"
	"encoding/json"
	"github.com/spy-exe/cells/apps/game-server/internal/engine"
	"github.com/spy-exe/cells/apps/game-server/internal/protocol"
	"github.com/spy-exe/cells/apps/game-server/internal/store"
	"testing"
	"time"
)

func TestExpeditionIsPrivatePersistsChoiceAndRestartsCleanly(t *testing.T) {
	backend := &memoryCheckpoint{}
	h := New(backend, 1600)
	a, err := h.Join(store.Session{Hash: "a", Nickname: "Alice", Mode: engine.Rogue}, nil)
	if err != nil {
		t.Fatal(err)
	}
	b, err := h.Join(store.Session{Hash: "b", Nickname: "Bruno", Mode: engine.Rogue}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if a.Room == b.Room {
		t.Fatal("solo worlds were shared")
	}
	if _, err = h.ExpeditionView("stranger"); err == nil {
		t.Fatal("foreign run leaked")
	}
	r := h.rooms[a.Room]
	r.Expedition.Phase = "choice"
	r.Expedition.Revision = 2
	if err = h.Input(a, protocol.Input{Kind: 1, X: 0, Y: 0, Actions: 1}); err != nil {
		t.Fatal(err)
	}
	if err = h.Save(context.Background()); err != nil {
		t.Fatal(err)
	}
	r.Expedition.Revision = 55
	var saved Room
	for _, v := range backend.rooms {
		if v.Key == a.Room {
			if err = json.Unmarshal(v.Data, &saved); err != nil {
				t.Fatal(err)
			}
		}
	}
	if saved.Expedition.Revision != 2 {
		t.Fatal("checkpoint aliases live run")
	}
	restored := New(backend, 1600)
	if err = restored.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	view, err := restored.ExpeditionView("a")
	if err != nil || view.Phase != "choice" || len(view.Offers) != 3 {
		t.Fatal(view, err)
	}
	if err = restored.ChooseMutation("a", "digestion", 2); err != nil {
		t.Fatal(err)
	}
	if err = restored.ChooseMutation("a", "digestion", 2); err == nil {
		t.Fatal("duplicate choice accepted")
	}
	if err = restored.ChooseMutation("b", "cilia", 2); err == nil {
		t.Fatal("another run advanced")
	}
	r = restored.rooms[a.Room]
	r.World.Player(a.Player).Peak = 400
	r.Expedition.Phase = "failed"
	oldSeed := r.Seed
	restored.rooms[a.Room].Peers[a.Player] = a
	if err = restored.Input(a, protocol.Input{Kind: 3}); err != nil {
		t.Fatal(err)
	}
	if r.Seed == oldSeed || r.Expedition.Phase != "running" || r.Expedition.Stage != 1 || r.World.Player(a.Player).Mutations != [3]uint8{} || r.World.Player(a.Player).Peak != 400 {
		t.Fatal("restart retained progression or lost record")
	}
	if err = r.World.Valid(); err != nil {
		t.Fatal(err)
	}
}
func TestPausedExpeditionStopsTicksButContinuesTransport(t *testing.T) {
	h := New(&memoryCheckpoint{}, 1600)
	a, _ := h.Join(store.Session{Hash: "paused", Nickname: "Alice", Mode: engine.Rogue}, nil)
	r := h.rooms[a.Room]
	r.Expedition.Phase = "choice"
	tick := r.World.Tick
	ctx, cancel := context.WithCancel(context.Background())
	go h.Run(ctx)
	time.Sleep(120 * time.Millisecond)
	h.mu.Lock()
	got := r.World.Tick
	h.mu.Unlock()
	cancel()
	<-h.done
	if got != tick || h.Ticks.Load() < 2 {
		t.Fatal("paused simulation or transport clock incorrect", got, h.Ticks.Load())
	}
}

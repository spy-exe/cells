package game

import (
	"context"
	"fmt"
	"github.com/spy-exe/cells/apps/game-server/internal/store"
	"sync"
	"testing"
	"time"
)

func TestMatchmakingReservesDisconnectedIdentities(t *testing.T) {
	h := New(nil, 1600)
	var first *Peer
	for i := 0; i < 16; i++ {
		p, err := h.Join(store.Session{Hash: fmt.Sprintf("session-%d", i), Nickname: "Guest", Mode: "classic", Profile: "balanced"}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			first = p
		}
		h.Disconnect(p)
	}
	next, err := h.Join(store.Session{Hash: "newcomer", Nickname: "Newcomer", Mode: "classic", Profile: "balanced"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if next.Room == first.Room {
		t.Fatal("new client took a reconnecting player's reserved room slot")
	}
	returning, err := h.Join(store.Session{Hash: "session-0", Nickname: "Guest", Mode: "classic", Profile: "balanced"}, nil)
	if err != nil || returning.Room != first.Room || returning.Player != first.Player {
		t.Fatal("identity did not resume", err)
	}
}

type blockingCheckpoint struct {
	started, release chan struct{}
	once             sync.Once
}

func (b *blockingCheckpoint) LoadRooms(context.Context) ([]store.SavedRoom, error) { return nil, nil }
func (b *blockingCheckpoint) SaveRooms(ctx context.Context, _ []store.SavedRoom) error {
	b.once.Do(func() { close(b.started) })
	select {
	case <-b.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func TestCheckpointIOCannotPauseTicks(t *testing.T) {
	backend := &blockingCheckpoint{started: make(chan struct{}), release: make(chan struct{})}
	h := New(backend, 1600)
	h.saveEvery = 10 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	go h.Run(ctx)
	t.Cleanup(func() {
		close(backend.release)
		cancel()
		end, stop := context.WithTimeout(context.Background(), time.Second)
		defer stop()
		h.Close(end)
	})
	select {
	case <-backend.started:
	case <-time.After(time.Second):
		t.Fatal("checkpoint did not start")
	}
	before := h.Ticks.Load()
	time.Sleep(250 * time.Millisecond)
	if h.Ticks.Load() < before+4 {
		t.Fatal("database I/O stalled the simulation")
	}
}

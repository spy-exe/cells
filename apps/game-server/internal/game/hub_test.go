package game

import (
	"fmt"
	"github.com/spy-exe/cells/apps/game-server/internal/store"
	"testing"
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

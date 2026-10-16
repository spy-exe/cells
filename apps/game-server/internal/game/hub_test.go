package game

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spy-exe/cells/apps/game-server/internal/bot"
	"github.com/spy-exe/cells/apps/game-server/internal/store"
	"sync"
	"testing"
	"time"
)

type memoryCheckpoint struct{ rooms []store.SavedRoom }

func TestPublicPopulationAndPrivateClassicIsolation(t *testing.T) {
	for _, scene := range []struct {
		mode, room string
		want       int
	}{{"classic", "", 24}, {"arena", "", 32}, {"arena", "friends", 32}, {"classic", "friends", 0}} {
		h := New(nil, 1600)
		p, err := h.Join(store.Session{Hash: "population", Nickname: "Player", Mode: scene.mode, Room: scene.room, Profile: "balanced"}, nil)
		if err != nil {
			t.Fatal(err)
		}
		n := 0
		for _, player := range h.rooms[p.Room].World.Players {
			if player.Bot {
				n++
			}
		}
		if n != scene.want {
			t.Fatalf("%s/%s bots=%d want=%d", scene.mode, scene.room, n, scene.want)
		}
		if err = h.rooms[p.Room].World.Valid(); err != nil {
			t.Fatal(err)
		}
	}
}
func TestReloadBackfillsBotsWithoutDuplicatingExistingPopulation(t *testing.T) {
	backend := &memoryCheckpoint{}
	h := New(backend, 1600)
	p, err := h.Join(store.Session{Hash: "migration", Nickname: "Player", Mode: "arena", Profile: "aggressive"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	r := h.rooms[p.Room]
	var remove []uint32
	for _, player := range r.World.Players {
		if player.Bot && len(remove) < 24 {
			remove = append(remove, player.ID)
		}
	}
	for _, id := range remove {
		r.World.RemovePlayer(id)
	}
	if err = h.Save(context.Background()); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		loaded := New(backend, 1600)
		if err = loaded.Load(context.Background()); err != nil {
			t.Fatal(err)
		}
		bots := 0
		for _, player := range loaded.rooms[p.Room].World.Players {
			if player.Bot {
				bots++
				if player.Profile != "aggressive" {
					t.Fatal("profile changed")
				}
			}
		}
		if bots != 32 {
			t.Fatal("migration population", bots)
		}
		if err = loaded.rooms[p.Room].World.Valid(); err != nil {
			t.Fatal(err)
		}
		if err = loaded.Save(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRunSummaryUsesPrivateIdentityAndFreezesAtDeath(t *testing.T) {
	h := New(nil, 1600)
	a, _ := h.Join(store.Session{Hash: "private-a", Nickname: "A", Mode: "classic", Room: "summary"}, nil)
	b, _ := h.Join(store.Session{Hash: "private-b", Nickname: "B", Mode: "classic", Room: "summary"}, nil)
	r := h.rooms[a.Room]
	r.World.Tick = 300
	pa, pb := r.World.Player(a.Player), r.World.Player(b.Player)
	pa.RunPeak = 333
	pa.Peak = 400
	pa.RunBorn = 30
	pa.DiedAt = 240
	pa.Alive = false
	pa.RunKills = 2
	pb.RunPeak = 888
	pb.Peak = 900
	result, ok := h.RunSummary("private-a")
	if !ok || result.Peak != 333 || result.Record != 400 || result.Kills != 2 || result.Seconds != 7 || result.Alive {
		t.Fatal("private run", result)
	}
	r.World.Tick = 600
	frozen, _ := h.RunSummary("private-a")
	if frozen != result {
		t.Fatal("dead run kept accumulating survival", frozen)
	}
	if other, _ := h.RunSummary("private-b"); other.Peak != 888 {
		t.Fatal("shared records", other)
	}
	if _, ok = h.RunSummary("unknown"); ok {
		t.Fatal("unknown identity saw records")
	}
}

func (m *memoryCheckpoint) LoadRooms(context.Context) ([]store.SavedRoom, error) { return m.rooms, nil }
func (m *memoryCheckpoint) SaveRooms(_ context.Context, r []store.SavedRoom) error {
	m.rooms = r
	return nil
}
func TestGoalMemorySurvivesCheckpointAndDoesNotAlias(t *testing.T) {
	backend := &memoryCheckpoint{}
	h := New(backend, 1600)
	p, err := h.Join(store.Session{Hash: "checkpoint-player", Nickname: "Player", Mode: "arena", Profile: "balanced"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	memory := &bot.Memory{Seen: map[uint32]bot.Track{9: {X: 12}}, GoalID: 9, GoalUntil: 45}
	h.rooms[p.Room].Memories[p.Player] = memory
	if err = h.Save(context.Background()); err != nil {
		t.Fatal(err)
	}
	memory.GoalID = 0
	memory.Seen[9] = bot.Track{X: 99}
	var saved Room
	if err = json.Unmarshal(backend.rooms[0].Data, &saved); err != nil {
		t.Fatal(err)
	}
	got := saved.Memories[p.Player]
	if got.GoalID != 9 || got.GoalUntil != 45 || got.Seen[9].X != 12 {
		t.Fatal("checkpoint lost goal or shared mutable memory", got)
	}
	restored := New(backend, 1600)
	if err = restored.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	if restored.rooms[p.Room].Memories[p.Player].GoalID != 9 {
		t.Fatal("restored goal missing")
	}
}

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

func TestTacticalPopulationSpreadsWorkAndRetainsFiveHzPerAgent(t *testing.T) {
	h := New(nil, 1600)
	p, err := h.Join(store.Session{Hash: "schedule", Nickname: "Observer", Mode: "arena", Profile: "balanced"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	r := h.rooms[p.Room]
	counts := map[uint32]int{}
	peak := 0
	for tick := uint64(0); tick < 12; tick++ {
		r.World.Tick = tick
		before := map[uint32]uint64{}
		for id, m := range r.Memories {
			for _, track := range m.Seen {
				before[id] = max(before[id], track.Tick)
			}
		}
		thinkBots(r)
		changed := 0
		for _, player := range r.World.Players {
			if !player.Bot {
				continue
			}
			m := r.Memories[player.ID]
			if m == nil {
				continue
			}
			var latest uint64
			for _, track := range m.Seen {
				latest = max(latest, track.Tick)
			}
			old, existed := before[player.ID]
			if !existed || latest != old {
				changed++
				counts[player.ID]++
			}
		}
		peak = max(peak, changed)
	}
	if len(counts) != 32 || peak > 6 {
		t.Fatal("tactical work still bursts", len(counts), peak)
	}
	for id, n := range counts {
		if n != 2 {
			t.Fatalf("agent %d thought %d times in12 ticks", id, n)
		}
	}
}

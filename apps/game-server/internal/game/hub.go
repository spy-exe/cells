package game

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coder/websocket"
	"github.com/spy-exe/cells/apps/game-server/internal/bot"
	"github.com/spy-exe/cells/apps/game-server/internal/engine"
	"github.com/spy-exe/cells/apps/game-server/internal/expedition"
	"github.com/spy-exe/cells/apps/game-server/internal/protocol"
	"github.com/spy-exe/cells/apps/game-server/internal/store"
	"log/slog"
	"maps"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Peer struct {
	Conn     *websocket.Conn
	Frames   chan []byte
	Controls chan []byte
	Player   uint32
	Room     string
	closed   atomic.Bool
}
type Room struct {
	Seed       uint64
	Expedition *expedition.Run `json:",omitempty"`
	World      *engine.World
	Identities map[string]uint32
	Absent     map[uint32]time.Time
	Peers      map[uint32]*Peer `json:"-"`
	Memories   map[uint32]*bot.Memory
}
type Hub struct {
	mu                              sync.Mutex
	rooms                           map[string]*Room
	store                           CheckpointStore
	saveEvery                       time.Duration
	size                            float64
	Ticks, Bytes, Dropped, Rejected atomic.Uint64
	durations                       [1800]float64
	durationIndex                   int
	closed                          bool
	stop                            chan struct{}
	done                            chan struct{}
	saveMu                          sync.Mutex
	saveJobs                        sync.WaitGroup
	saving                          atomic.Bool
}

type CheckpointStore interface {
	LoadRooms(context.Context) ([]store.SavedRoom, error)
	SaveRooms(context.Context, []store.SavedRoom) error
}

func New(st CheckpointStore, size float64) *Hub {
	return &Hub{store: st, size: size, rooms: map[string]*Room{}, stop: make(chan struct{}), done: make(chan struct{}), saveEvery: 30 * time.Second}
}
func (h *Hub) Load(ctx context.Context) error {
	saved, err := h.store.LoadRooms(ctx)
	if err != nil {
		return err
	}
	for _, v := range saved {
		var r Room
		if err = json.Unmarshal(v.Data, &r); err != nil || r.World == nil || r.World.Valid() != nil || len(r.World.Players) > 128 || r.Identities == nil {
			return fmt.Errorf("invalid room checkpoint %s", v.Key)
		}
		if r.World.Config.Size == 6000 && r.World.Config.Food == 2000 {
			r.World.Config.Food = 3600
		}
		r.Peers = map[uint32]*Peer{}
		if r.World.Config.Mode == engine.Rogue && (r.Expedition == nil || r.Expedition.Validate(r.World) != nil) {
			return fmt.Errorf("invalid expedition checkpoint %s", v.Key)
		}
		if r.Memories == nil {
			r.Memories = map[uint32]*bot.Memory{}
		}
		r.Absent = map[uint32]time.Time{}
		for _, p := range r.World.Players {
			if !p.Bot {
				r.Absent[p.ID] = time.Now().Add(30 * time.Second)
			}
		}
		// Older checkpoints did not contain per-run statistics. Start measuring
		// active lives at migration, without inventing their earlier duration.
		for i := range r.World.Players {
			p := &r.World.Players[i]
			if !p.HasRun && p.Alive {
				p.HasRun = true
				p.RunBorn = r.World.Tick
				p.RunPeak = r.World.Mass(p.ID)
			}
		}
		h.rooms[v.Key] = &r
		profile := "balanced"
		for _, p := range r.World.Players {
			if p.Bot && bot.ValidProfile(p.Profile) {
				profile = p.Profile
				break
			}
		}
		populateBots(&r, r.World.Config.Mode, strings.Contains(v.Key, "-private-"), profile)
	}
	return nil
}

type RunSummary struct {
	HasRun  bool    `json:"has_run"`
	Alive   bool    `json:"alive"`
	Peak    float64 `json:"peak"`
	Record  float64 `json:"record"`
	Kills   int     `json:"kills"`
	Seconds float64 `json:"seconds"`
	Deaths  int     `json:"deaths"`
}

func (h *Hub) RunSummary(hash string) (RunSummary, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, r := range h.rooms {
		id, ok := r.Identities[hash]
		if !ok {
			continue
		}
		p := r.World.Player(id)
		if p == nil {
			return RunSummary{}, false
		}
		until := r.World.Tick
		if !p.Alive {
			until = p.DiedAt
		}
		seconds := 0.
		if p.HasRun && until >= p.RunBorn {
			seconds = float64(until-p.RunBorn) / engine.TickRate
		}
		return RunSummary{HasRun: p.HasRun, Alive: p.Alive, Peak: p.RunPeak, Record: p.Peak, Kills: p.RunKills, Seconds: seconds, Deaths: p.Deaths}, true
	}
	return RunSummary{}, false
}
func (h *Hub) Join(s store.Session, c *websocket.Conn) (*Peer, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return nil, errors.New("server restarting")
	}
	key := ""
	matchmakingProfile := "mixed"
	if s.Mode == engine.Arena {
		matchmakingProfile = s.Profile
	}
	for _, k := range h.keys() {
		r := h.rooms[k]
		if r.World.Config.Mode == s.Mode {
			if _, ok := r.Identities[s.Hash]; ok {
				key = k
				break
			}
		}
	}
	if key == "" && s.Room != "" {
		key = s.Mode + "-private-" + s.Room
	}
	if s.Mode == engine.Rogue {
		key = "rogue-private-" + s.Hash
	}
	if key == "" {
		for i := 0; i < 16; i++ {
			candidate := fmt.Sprintf("%s-%s-public-%d", s.Mode, matchmakingProfile, i)
			r := h.rooms[candidate]
			if r == nil || len(r.Identities) < 16 {
				key = candidate
				break
			}
		}
	}
	if key == "" {
		return nil, errors.New("arenas full")
	}
	r := h.rooms[key]
	if r == nil {
		if len(h.rooms) >= 32 {
			return nil, errors.New("arena limit")
		}
		var seed [8]byte
		if _, err := rand.Read(seed[:]); err != nil {
			return nil, err
		}
		cfg := engine.DefaultConfig(s.Mode)
		if h.size > 0 {
			cfg.Size = h.size
			cfg.Food = 900
			cfg.Viruses = 8
		}
		r = &Room{World: engine.New(binary.LittleEndian.Uint64(seed[:]), cfg), Identities: map[string]uint32{}, Absent: map[uint32]time.Time{}, Peers: map[uint32]*Peer{}, Memories: map[uint32]*bot.Memory{}}
		r.Seed = binary.LittleEndian.Uint64(seed[:])
		// Private Classic rooms stay human-only, making friend sessions and public
		// validation independent of bot interference. AI Arena always has agents.
		populateBots(r, s.Mode, s.Room != "", s.Profile)
		h.rooms[key] = r
	}
	if len(r.Peers) >= 32 {
		return nil, errors.New("arena full")
	}
	id, ok := r.Identities[s.Hash]
	if !ok {
		id = r.World.AddPlayer(s.Nickname, false, "")
		r.Identities[s.Hash] = id
		if s.Mode == engine.Rogue {
			r.Expedition = expedition.New(r.Seed, r.World, id)
		}
		if s.Spectator || s.Profile == "spectator" {
			r.World.Spectate(id)
		}
	}
	if old := r.Peers[id]; old != nil {
		old.closed.Store(true)
		go old.Conn.Close(websocket.StatusPolicyViolation, "session replaced")
	}
	p := &Peer{Conn: c, Frames: make(chan []byte, 1), Controls: make(chan []byte, 4), Player: id, Room: key}
	r.Peers[id] = p
	delete(r.Absent, id)
	p.Frames <- protocol.Snapshot(r.World.Observe(id))
	return p, nil
}

func populateBots(r *Room, mode string, private bool, profile string) {
	target := 0
	if mode == engine.Arena {
		target = 32
	} else if mode == engine.Classic && !private {
		target = 24
	}
	count := 0
	for _, p := range r.World.Players {
		if p.Bot {
			count++
		}
	}
	names := []string{"Mica", "Nori", "Vértice", "Órbita", "Talo", "Nimbo", "Coral", "Pingo", "Lume", "Brisa", "Âmbar", "Cais", "Silex", "Névoa", "Vega", "Pólux", "Musgo", "Gaia", "Iris", "Cedro", "Duna", "Fluxo", "Halo", "Lótus", "Nacre", "Onda", "Prisma", "Raiz", "Sal", "Trama", "Una", "Zênite"}
	for i := count; i < target; i++ {
		policy := bot.Profiles[i%len(bot.Profiles)]
		if mode == engine.Arena && bot.ValidProfile(profile) {
			policy = profile
		}
		r.World.AddPlayer(names[i], true, policy)
	}
}
func (h *Hub) keys() []string {
	keys := make([]string, 0, len(h.rooms))
	for k := range h.rooms {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
func (h *Hub) Disconnect(p *Peer) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if r := h.rooms[p.Room]; r != nil && r.Peers[p.Player] == p {
		delete(r.Peers, p.Player)
		r.Absent[p.Player] = time.Now().Add(30 * time.Second)
	}
}
func (h *Hub) Input(p *Peer, in protocol.Input) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	r := h.rooms[p.Room]
	if r == nil || r.Peers[p.Player] != p {
		return errors.New("stale session")
	}
	switch in.Kind {
	case 1:
		if r.Expedition != nil && r.Expedition.Phase != "running" {
			return nil
		}
		if !r.World.Player(p.Player).Alive {
			return nil
		}
		return r.World.Input(p.Player, in.X, in.Y, in.Actions)
	case 3:
		if r.Expedition != nil {
			if r.Expedition.Phase != "running" && r.Expedition.Phase != "choice" {
				restartExpedition(r, p)
			}
		} else {
			_ = r.World.Respawn(p.Player)
		}
	case 4: // A dead player observes the leader; never receives hidden map data.
	case 5:
		select {
		case p.Controls <- in.Echo:
		default:
			return errors.New("control flood")
		}
	}
	return nil
}
func (h *Hub) Serve(ctx context.Context, p *Peer) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer p.Conn.CloseNow()
	defer h.Disconnect(p)
	p.Conn.SetReadLimit(64)
	go func() {
		defer cancel()
		for {
			var b []byte
			select {
			case b = <-p.Controls:
			case b = <-p.Frames:
			case <-ctx.Done():
				return
			}
			sendCtx, done := context.WithTimeout(ctx, 3*time.Second)
			err := p.Conn.Write(sendCtx, websocket.MessageBinary, b)
			done()
			if err != nil {
				return
			}
			h.Bytes.Add(uint64(len(b)))
		}
	}()
	window, count, seq := time.Now(), 0, uint32(0)
	for {
		readCtx, done := context.WithTimeout(ctx, 20*time.Second)
		kind, b, err := p.Conn.Read(readCtx)
		done()
		if err != nil {
			return
		}
		if time.Since(window) > time.Second {
			window = time.Now()
			count = 0
		}
		count++
		in, err := protocol.Decode(b)
		if kind != websocket.MessageBinary || count > 60 || err != nil || in.Kind == 1 && in.Seq <= seq {
			h.Rejected.Add(1)
			_ = p.Conn.Close(websocket.StatusPolicyViolation, "invalid input")
			return
		}
		if in.Kind == 1 {
			seq = in.Seq
		}
		if err = h.Input(p, in); err != nil {
			h.Rejected.Add(1)
			return
		}
	}
}
func (h *Hub) Run(ctx context.Context) {
	defer close(h.done)
	ticker := time.NewTicker(time.Second / engine.TickRate)
	defer ticker.Stop()
	save := time.NewTicker(h.saveEvery)
	defer save.Stop()
	for {
		select {
		case <-h.stop:
			return
		case <-ctx.Done():
			return
		case <-save.C:
			if h.saving.CompareAndSwap(false, true) {
				h.saveJobs.Add(1)
				go func() {
					defer h.saveJobs.Done()
					defer h.saving.Store(false)
					if err := h.Save(ctx); err != nil {
						slog.Error("checkpoint failed", "error", err.Error())
					}
				}()
			}
		case <-ticker.C:
			start := time.Now()
			h.mu.Lock()
			for _, key := range h.keys() {
				r := h.rooms[key]
				active := r.Expedition == nil || r.Expedition.Phase == "running"
				if active && r.World.Tick%6 == 0 {
					for _, player := range r.World.Players {
						if !player.Bot {
							continue
						}
						if !player.Alive {
							if r.Expedition == nil && r.World.Tick > player.DiedAt+90 {
								_ = r.World.Respawn(player.ID)
							}
							continue
						}
						memory := r.Memories[player.ID]
						if memory == nil {
							memory = &bot.Memory{}
							r.Memories[player.ID] = memory
						}
						intent := bot.Decide(r.World.Observe(player.ID), memory, player.Profile)
						_ = r.World.Input(player.ID, intent.X, intent.Y, intent.Actions)
					}
				}
				for id, until := range r.Absent {
					if time.Now().After(until) {
						r.World.RemovePlayer(id)
						delete(r.Absent, id)
						for token, player := range r.Identities {
							if player == id {
								delete(r.Identities, token)
							}
						}
					}
				}
				if active {
					r.World.Step()
					if r.Expedition != nil {
						r.Expedition.Advance(r.World)
					}
				}
				if !active || r.World.Tick%2 == 0 || r.World.ActionOccurred() {
					for id, p := range r.Peers {
						b := protocol.Snapshot(r.World.Observe(id))
						select {
						case p.Frames <- b:
						default:
							select {
							case <-p.Frames:
							default:
							}
							select {
							case p.Frames <- b:
							default:
							}
							h.Dropped.Add(1)
						}
					}
				}
				if len(r.Peers) == 0 && len(r.Absent) == 0 {
					delete(h.rooms, key)
				}
			}
			h.durations[h.durationIndex%len(h.durations)] = float64(time.Since(start).Microseconds()) / 1000
			h.durationIndex++
			h.Ticks.Add(1)
			h.mu.Unlock()
		}
	}
}
func (h *Hub) Save(ctx context.Context) error {
	h.saveMu.Lock()
	defer h.saveMu.Unlock()
	h.mu.Lock()
	keys := h.keys()
	snapshots := make(map[string]Room, len(keys))
	for _, key := range keys {
		r := h.rooms[key]
		world := *r.World
		world.Entities = append([]engine.Entity{}, r.World.Entities...)
		world.Players = append([]engine.Player{}, r.World.Players...)
		memories := make(map[uint32]*bot.Memory, len(r.Memories))
		for id, m := range r.Memories {
			copy := *m
			copy.Seen = maps.Clone(m.Seen)
			memories[id] = &copy
		}
		var run *expedition.Run
		if r.Expedition != nil {
			copy := *r.Expedition
			run = &copy
		}
		snapshots[key] = Room{Seed: r.Seed, Expedition: run, World: &world, Identities: maps.Clone(r.Identities), Absent: maps.Clone(r.Absent), Memories: memories}
	}
	h.mu.Unlock()
	var saved []store.SavedRoom
	for _, key := range keys {
		data, err := json.Marshal(snapshots[key])
		if err != nil {
			return err
		}
		saved = append(saved, store.SavedRoom{Key: key, Data: data})
	}
	return h.store.SaveRooms(ctx, saved)
}
func (h *Hub) Close(ctx context.Context) {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return
	}
	h.closed = true
	close(h.stop)
	for _, r := range h.rooms {
		for _, p := range r.Peers {
			go p.Conn.Close(websocket.StatusGoingAway, "server restarting")
		}
	}
	h.mu.Unlock()
	select {
	case <-h.done:
	case <-ctx.Done():
		return
	}
	h.saveJobs.Wait()
	if err := h.Save(ctx); err != nil {
		slog.Error("final checkpoint failed")
	}
}
func (h *Hub) Stats() map[string]any {
	h.mu.Lock()
	defer h.mu.Unlock()
	players, bots, entities := 0, 0, 0
	for _, r := range h.rooms {
		players += len(r.Peers)
		entities += len(r.World.Entities)
		for _, p := range r.World.Players {
			if p.Bot {
				bots++
			}
		}
	}
	samples := append([]float64{}, h.durations[:min(h.durationIndex, len(h.durations))]...)
	sort.Float64s(samples)
	p95, p99 := 0., 0.
	if len(samples) > 0 {
		p95 = samples[int(float64(len(samples)-1)*.95)]
		p99 = samples[int(float64(len(samples)-1)*.99)]
	}
	return map[string]any{"players": players, "bots": bots, "rooms": len(h.rooms), "entities": entities, "ticks": h.Ticks.Load(), "tick_p95_ms": p95, "tick_p99_ms": p99, "bytes": h.Bytes.Load(), "dropped": h.Dropped.Load(), "rejected": h.Rejected.Load()}
}

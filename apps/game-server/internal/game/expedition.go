package game

import (
	"errors"
	"github.com/spy-exe/cells/apps/game-server/internal/bot"
	"github.com/spy-exe/cells/apps/game-server/internal/engine"
	"github.com/spy-exe/cells/apps/game-server/internal/expedition"
)

func (h *Hub) ExpeditionView(hash string) (expedition.View, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, r := range h.rooms {
		if id, ok := r.Identities[hash]; ok && r.Expedition != nil && id == r.Expedition.Player {
			return r.Expedition.View(r.World), nil
		}
	}
	return expedition.View{}, errors.New("expedition not found")
}
func (h *Hub) ChooseMutation(hash, mutation string, revision uint64) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, r := range h.rooms {
		if id, ok := r.Identities[hash]; ok && r.Expedition != nil && id == r.Expedition.Player {
			return r.Expedition.Choose(r.World, mutation, revision)
		}
	}
	return errors.New("expedition not found")
}
func restartExpedition(r *Room, peer *Peer) {
	old := *r.World.Player(peer.Player)
	seed := r.Seed + 0x9e3779b97f4a7c15
	world := engine.New(seed, r.World.Config)
	id := world.AddPlayer(old.Name, false, "")
	world.Player(id).Peak = max(old.Peak, world.Player(id).Peak)
	for hash, player := range r.Identities {
		if player == peer.Player {
			r.Identities[hash] = id
		}
	}
	delete(r.Peers, peer.Player)
	peer.Player = id
	r.Peers[id] = peer
	r.Seed = seed
	r.World = world
	r.Memories = map[uint32]*bot.Memory{}
	r.Expedition = expedition.New(seed, world, id)
}

package game

import "sync"

type PlayerID uint64 // TODO maybe string?

// Player contains player info. Pretty minimalist :)
type Player struct {
	ID   PlayerID `json:"id"`
	Name string   `json:"name"`
}

type PlayerRegistry struct {
	nextPlayerID PlayerID
	players      map[PlayerID]Player
	lock         sync.RWMutex // locks nextPlayerID and players
}

func NewPlayerRegistry() *PlayerRegistry {
	return &PlayerRegistry{
		nextPlayerID: 0,
		players:      map[PlayerID]Player{},
		lock:         sync.RWMutex{},
	}
}

// Register adds a new player to registry.
func (r *PlayerRegistry) Register(name string) Player {
	r.lock.Lock()
	defer r.lock.Unlock()
	newPlayer := Player{
		ID:   r.nextPlayerID,
		Name: name,
	}
	r.players[r.nextPlayerID] = newPlayer
	r.nextPlayerID++
	return newPlayer
}

// Get player from registry if it exists.
func (r *PlayerRegistry) Get(id PlayerID) (player Player, ok bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	player, ok = r.players[id]
	return
}

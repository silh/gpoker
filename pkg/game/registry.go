package game

import (
	"encoding/json"
	"os"
	"sync"
)

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
		nextPlayerID: 1,
		players:      map[PlayerID]Player{},
		lock:         sync.RWMutex{},
	}
}

// NewPlayerRegistryFromFile creates a registry and pre-populates players from a provided file.
func NewPlayerRegistryFromFile(filepath string) (*PlayerRegistry, error) {
	registry := &PlayerRegistry{
		nextPlayerID: 1,
		players:      map[PlayerID]Player{},
		lock:         sync.RWMutex{},
	}
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	var names []string // names
	if err = json.NewDecoder(file).Decode(&names); err != nil {
		return nil, err
	}
	for _, name := range names {
		registry.Register(name)
	}
	return registry, nil
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

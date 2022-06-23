package game

import (
	"errors"
	"sort"
	"sync"
)

var ErrGameNotFound = errors.New("game not found")
var ErrPlayerNotInGame = errors.New("player not in game")

type GameID uint64 // TODO same as the above
type Vote uint64   // TODO well that should probably be an interface? Or some enum

// Poker tracks game info. The structure is not ideal and should be reconsidered.
type Poker struct {
	ID      GameID              `json:"id"`
	Name    string              `json:"name"`
	Players map[PlayerID]Player `json:"players"`
	Votes   map[PlayerID]Vote   `json:"votes"`
}

// Dealer controls all games.
type Dealer struct {
	nextGameID GameID
	games      map[GameID]Poker
	lock       sync.RWMutex // protects nextGameID and games. This can be changed in the future to work with channels
}

func (d *Dealer) CreateGame(name string, creator Player) (*Poker, error) { // TODO not sure if this should return a pointer
	d.lock.Lock()
	defer d.lock.Unlock()
	poker := Poker{
		ID:      d.nextGameID,
		Players: map[PlayerID]Player{creator.ID: creator},
		Name:    name,
	}
	d.nextGameID++
	d.games[poker.ID] = poker
	return &poker, nil
}

func (d *Dealer) ListGameNames() []string {
	d.lock.RLock()
	defer d.lock.RUnlock()
	gamesList := make([]string, 0, len(d.games))
	for _, v := range d.games {
		gamesList = append(gamesList, v.Name)
	}
	sort.Strings(gamesList)
	return gamesList
}

func (d *Dealer) GetGame(id GameID) (Poker, bool) {
	d.lock.Lock()
	defer d.lock.Unlock()
	poker, ok := d.games[id]
	return poker, ok
}

// TODO we obviously don't handle the case where player is deleted while they are in a game
func (d *Dealer) JoinGame(gameID GameID, player Player) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	game, ok := d.games[gameID]
	if !ok {
		return ErrGameNotFound
	}
	game.Players[player.ID] = player
	return nil
}

func (d *Dealer) Vote(gameId GameID, voteReq VoteRequest) error {
	d.lock.Lock() // TODO full lock here just to vote... That's not scalable.
	defer d.lock.Unlock()
	game, ok := d.games[gameId]
	if !ok {
		return ErrGameNotFound
	}
	player, ok := game.Players[voteReq.PlayerID]
	if !ok {
		return ErrPlayerNotInGame
	}
	game.Votes[player.ID] = voteReq.Vote
	return nil
}

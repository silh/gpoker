package game

import (
	"errors"
	"sort"
	"sync"
)

var ErrGameNotFound = errors.New("game not found")
var ErrPlayerNotInGame = errors.New("player not in game")

type GameID uint64 // TODO same as the above
type Vote string   // TODO well that should probably be an interface? Or some enum

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
	games      map[GameID]*Poker
	lock       sync.RWMutex // protects nextGameID and games. This can be changed in the future to work with channels
}

// NewDealer creates a new instance of a Dealer with nextGameID set to 1.
func NewDealer() *Dealer {
	return &Dealer{
		nextGameID: 1,
		games:      make(map[GameID]*Poker), // TODO why do we store here a pointer, but in registry - a struct
		lock:       sync.RWMutex{},
	}
}

// CreateGame starts a new game with creator as participant.
func (d *Dealer) CreateGame(name string, creator Player) (GameResponse, error) { // TODO not sure if this should return a pointer
	d.lock.Lock()
	defer d.lock.Unlock()
	poker := Poker{
		ID:      d.nextGameID,
		Players: map[PlayerID]Player{creator.ID: creator},
		Votes:   map[PlayerID]Vote{},
		Name:    name,
	}
	d.nextGameID++
	d.games[poker.ID] = &poker
	return gameToResponse(&poker), nil
}

// ListGameNames returns names and ids of all present games sorted by name.
func (d *Dealer) ListGameNames() []GameListEntry {
	d.lock.RLock()
	defer d.lock.RUnlock()
	gamesList := make([]GameListEntry, 0, len(d.games))
	for _, v := range d.games {
		gamesList = append(gamesList, GameListEntry{
			ID:   v.ID,
			Name: v.Name,
		})
	}
	sort.Slice(gamesList, func(i, j int) bool { return gamesList[i].Name < gamesList[j].Name })
	return gamesList
}

// GetGame returns information about the game by its ID. Players inside a game are sorted by name.
func (d *Dealer) GetGame(id GameID) (GameResponse, bool) {
	d.lock.Lock()
	defer d.lock.Unlock()
	poker, ok := d.games[id]
	if !ok {
		return GameResponse{}, ok
	}
	return gameToResponse(poker), ok
}

// TODO we obviously don't handle the case where player is deleted while they are in a game
// Also same player can potentially be added twice (well, overwritten technically).
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

func gameToResponse(poker *Poker) GameResponse {
	resp := GameResponse{
		ID:      poker.ID,
		Name:    poker.Name,
		Players: make([]PlayerResponse, 0, len(poker.Players)),
	}
	for _, player := range poker.Players {
		resp.Players = append(resp.Players, PlayerResponse{
			ID:   player.ID,
			Name: player.Name,
			Vote: poker.Votes[player.ID],
		})
	}
	sort.Slice(resp.Players, func(i, j int) bool { return resp.Players[i].Name < resp.Players[j].Name })
	return resp
}

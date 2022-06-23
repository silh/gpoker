package game

import (
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sort"
	"sync"
)

var ErrGameNotFound = errors.New("game not found")

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

func (d *Dealer) CreateGame(name string, creator Player) (*Poker, error) {
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

func (d *Dealer) JoinGame(gameID GameID, joinReq JoinPokerRequest) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	game, ok := d.games[gameID]
	if !ok {
		return ErrGameNotFound
	}
	game.Players[joinReq.Player.ID] = joinReq.Player
	return nil
}

func (d *Dealer) AcceptVote(c *gin.Context) {
	d.lock.Lock()
	defer d.lock.Unlock()
	gameId, ok := ParamUint64(c, "gameId")
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	playerId, ok := ParamUint64(c, "playerId")
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	game, ok := d.games[GameID(gameId)]
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	var voteReq VoteRequest
	if err := c.BindJSON(&voteReq); err != nil {
		log.Printf("Failed to parse AcceptVote request = %s", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	player, ok := game.Players[PlayerID(playerId)]
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest) // player not part of the game
		return
	}
	game.Votes[player.ID] = voteReq.Vote
	c.JSON(http.StatusOK, game)
}

func (d *Dealer) sortedGamesList() []string {
	// TODO this can be quite slow if we have a lot of games.
	d.lock.RLock()
	defer d.lock.RUnlock()
	gamesList := make([]string, 0, len(d.games))
	for _, v := range d.games {
		gamesList = append(gamesList, v.Name)
	}
	sort.Strings(gamesList)
	return gamesList
}

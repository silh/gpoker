package game

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"
)

type Poker struct {
	ID      uint64            `json:"id"`
	Name    string            `json:"name"`
	Players map[uint64]Player `json:"players"`
}

type Player struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Vote uint64 `json:"vote"`
}

// Dealer controls all games.
type Dealer struct {
	counter uint64
	games   map[uint64]Poker
	lock    sync.RWMutex // protects counter and games. This can be changed in the future to work with channels

}

func (d *Dealer) CreateGame(c *gin.Context) {
	d.lock.Lock()
	defer d.lock.Unlock()
	var gameReq CreatePokerRequest
	if err := c.Bind(&gameReq); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	poker := Poker{
		ID:      d.counter,
		Players: map[uint64]Player{gameReq.Player.ID: gameReq.Player},
		Name:    gameReq.GameName,
	}
	d.counter++
	d.games[poker.ID] = poker
	c.JSON(http.StatusCreated, &poker)
}

func (d *Dealer) GetGame(c *gin.Context) {
	d.lock.Lock()
	defer d.lock.Unlock()
	id, ok := ParamUint64(c, "gameId")
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	poker, ok := d.games[id]
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, &poker)
}

func (d *Dealer) ListGameNames(c *gin.Context) {
	sortedGames := d.sortedGamesList()
	c.JSON(http.StatusOK, sortedGames)
}

func (d *Dealer) JoinGame(c *gin.Context) {
	d.lock.Lock()
	defer d.lock.Unlock()
	var joinReq JoinPokerRequest
	if err := c.BindJSON(&joinReq); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	id, ok := ParamUint64(c, "gameId")
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	game, ok := d.games[id]
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	game.Players[joinReq.Player.ID] = joinReq.Player
	c.JSON(http.StatusOK, &game)
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
	game, ok := d.games[gameId]
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
	player, ok := game.Players[playerId]
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest) // player not part of the game
		return
	}
	player.Vote = voteReq.Vote
	game.Players[playerId] = player
	c.JSON(http.StatusOK, game)
}

// ParamUint64 extracts parameter from gin.Context that is expected to be uint64.
func ParamUint64(c *gin.Context, name string) (uint64, bool) {
	idStr := c.Param(name)
	id, err := strconv.ParseUint(idStr, 10, 0)
	return id, err == nil
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

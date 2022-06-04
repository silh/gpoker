package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Poker struct {
	ID      uint64            `json:"id"`
	Players map[uint64]Player `json:"players"`
}

type Player struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Vote uint64 `json:"vote"`
}

// CreatePokerRequest to start a game.
type CreatePokerRequest struct {
	Player Player `json:"player"`
}

// JoinPokerRequest to join a game.
type JoinPokerRequest struct {
	Player Player `json:"player"`
}

// VoteRequest for player's vote.
type VoteRequest struct {
	Vote uint64 `json:"vote"`
}

func main() {
	app := gin.Default()
	dealer := Dealer{
		counter: 0,
		games:   make(map[uint64]Poker),
	}

	app.POST("/game", dealer.createGame)
	app.GET("/game/:gameId", dealer.getGame)
	app.PUT("/game/:gameId/join", dealer.joinGame)
	app.POST("/game/:gameId/player/:playerId/vote", dealer.vote)
	if err := app.Run(); err != nil {
		log.Fatalf("Error = %s", err)
	}
}

// Dealer controls all games.
type Dealer struct {
	counter uint64
	games   map[uint64]Poker
	lock    sync.RWMutex // protects counter and games

}

func (d *Dealer) createGame(c *gin.Context) {
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
	}
	d.counter++
	d.games[poker.ID] = poker
	c.JSON(http.StatusCreated, &poker)
}

func (d *Dealer) getGame(c *gin.Context) {
	d.lock.Lock()
	defer d.lock.Unlock()
	id, ok := extractUint64(c, "gameId")
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

func (d *Dealer) joinGame(c *gin.Context) {
	d.lock.Lock()
	defer d.lock.Unlock()
	var joinReq JoinPokerRequest
	if err := c.BindJSON(&joinReq); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	id, ok := extractUint64(c, "gameId")
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

func (d *Dealer) vote(c *gin.Context) {
	d.lock.Lock()
	defer d.lock.Unlock()
	gameId, ok := extractUint64(c, "gameId")
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	playerId, ok := extractUint64(c, "playerId")
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	game, ok := d.games[gameId]
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	player, ok := game.Players[playerId]
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest) // player not part of the game
		return
	}
	var voteReq VoteRequest
	if err := c.BindJSON(&voteReq); err != nil {
		log.Printf("Failed to parse vote request = %s", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	player.Vote = voteReq.Vote
	game.Players[playerId] = player
	c.JSON(http.StatusOK, game)
}

func extractUint64(c *gin.Context, name string) (uint64, bool) {
	idStr := c.Param(name)
	id, err := strconv.ParseUint(idStr, 10, 0)
	return id, err == nil
}

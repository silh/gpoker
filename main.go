package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Poker struct {
	ID      uint64   `json:"id"`
	Players []Player `json:"players"`
}

type Player struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type CreatePokerRequest struct {
	Player Player `json:"player"`
}

func main() {
	app := gin.Default()
	dealer := Dealer{
		counter: 0,
		games:   make(map[uint64]Poker),
	}

	app.POST("/game", dealer.createGame)
	app.GET("/game/:id", dealer.getGame)
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
		Players: []Player{gameReq.Player},
	}
	d.counter++
	d.games[poker.ID] = poker
	c.JSON(http.StatusCreated, &poker)
}

func (d *Dealer) getGame(c *gin.Context) {
	d.lock.Lock()
	defer d.lock.Unlock()
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
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

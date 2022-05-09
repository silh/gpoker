package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

var counter int64 // FIXME of course this is not thread safe

var games map[int64]Poker // FIXME of course this is not thread safe

type Poker struct {
	ID      int64    `json:"id"`
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
	games = make(map[int64]Poker)

	app := gin.Default()

	app.POST("/game", createGame)
	app.GET("/game/:id", getGame)
	if err := app.Run(); err != nil {
		log.Fatalf("Error = %s", err)
	}
}

func createGame(c *gin.Context) {
	var gameReq CreatePokerRequest
	if err := c.Bind(&gameReq); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	poker := Poker{
		ID:      counter,
		Players: []Player{gameReq.Player},
	}
	counter++
	games[poker.ID] = poker
	c.JSON(http.StatusCreated, &poker)
}

func getGame(c *gin.Context) {
	id := c.GetInt64("id")
	poker, ok := games[id]
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, &poker)
}

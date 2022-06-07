package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"gpoker/game"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	app := gin.Default()
	dealer := game.Dealer{
		Counter: 0,
		Games:   make(map[uint64]game.Poker),
	}
	app.POST("/api/game", dealer.CreateGame)
	app.GET("/api/game/:gameId", dealer.GetGame)
	app.PUT("/api/game/:gameId/join", dealer.JoinGame)
	app.POST("/api/game/:gameId/player/:playerId/vote", dealer.AcceptVote)
	srv := http.Server{
		Addr:    ":8080", // FIXME don't hardcode that
		Handler: app,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error = %s", err)
		}
	}()
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down the server = %s", err)
	}
}

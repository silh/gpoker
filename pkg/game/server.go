package game

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// Server is a main game server
type Server struct {
	srv *http.Server
}

func NewServer() *Server {
	app := gin.Default()
	dealer := Dealer{
		Counter: 0,
		Games:   make(map[uint64]Poker),
	}
	app.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	app.POST("/api/game", dealer.CreateGame)
	app.GET("/api/game/:gameId", dealer.GetGame)
	app.PUT("/api/game/:gameId/join", dealer.JoinGame)
	app.POST("/api/game/:gameId/player/:playerId/vote", dealer.AcceptVote)
	return &Server{
		srv: &http.Server{
			Addr:    ":8080", // FIXME don't hardcode that
			Handler: app,
		},
	}
}

func (s *Server) Start() {
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error = %s", err)
		}
	}()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

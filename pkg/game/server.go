package game

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sync"
)

// Server is a main game server
type Server struct {
	srv *http.Server

	startOnce sync.Once
}

// NewServer creates a new Server.
func NewServer() *Server {
	app := gin.Default()
	app.Use(cors.Default()) // CORS enabler for dev
	dealer := Dealer{
		counter: 0,
		games:   make(map[uint64]Poker),
	}
	app.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	app.POST("/api/games", dealer.CreateGame)
	app.GET("/api/games", dealer.ListGameNames)
	app.GET("/api/games/:gameId", dealer.GetGame)
	app.PUT("/api/games/:gameId/join", dealer.JoinGame)
	app.POST("/api/games/:gameId/players/:playerId/vote", dealer.AcceptVote)
	return &Server{
		srv: &http.Server{
			Addr:    ":8080", // FIXME don't hardcode that
			Handler: app,
		},
	}
}

// NewStartedServer creates a new Server and starts it.
func NewStartedServer() *Server {
	srv := NewServer()
	srv.Start()
	return srv
}

func (s *Server) Start() {
	s.startOnce.Do(func() {
		go func() {
			if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server error = %s", err)
			}
		}()
	})
}

func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

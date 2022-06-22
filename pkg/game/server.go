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
	srv            *http.Server
	dealer         *Dealer
	playerRegistry *PlayerRegistry

	startOnce sync.Once
}

// NewServer creates a new Server.
func NewServer() *Server {
	app := gin.Default()
	app.Use(cors.Default()) // CORS enabler for dev
	dealer := Dealer{
		nextGameID: 0,
		games:      make(map[GameID]Poker),
	}
	registry := NewPlayerRegistry()
	srv := &Server{
		srv: &http.Server{
			Addr:    ":8080", // FIXME don't hardcode that
			Handler: app,
		},
		dealer:         &dealer,
		playerRegistry: registry,
	}

	app.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	app.POST("/signup", srv.signup)

	app.POST("/api/games", srv.createGame)
	app.GET("/api/games", dealer.ListGameNames)
	app.GET("/api/games/:gameId", dealer.GetGame)
	app.PUT("/api/games/:gameId/join", dealer.JoinGame)
	app.POST("/api/games/:gameId/players/:playerId/vote", dealer.AcceptVote)
	return srv
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

func (s *Server) signup(c *gin.Context) {
	var req RegisterUserRequest
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	player := s.playerRegistry.Register(req.Name)
	c.JSON(http.StatusOK, &player)
}

func (s *Server) createGame(c *gin.Context) {
	var req CreatePokerRequest
	if err := c.BindJSON(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	player, ok := s.playerRegistry.Get(req.CreatorID)
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	game, err := s.dealer.CreateGame(req.GameName, player)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusCreated, &game)
}

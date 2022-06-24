package game

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"sync"
)

var ErrBadGameID = errors.New("game ID is not provided or is incorrect")

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
		games:      make(map[GameID]*Poker),
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
	app.POST("/api/signup", srv.signup)

	app.POST("/api/games", srv.createGame)
	app.GET("/api/games", srv.listGameNames)
	app.GET("/api/games/:gameId", srv.getGame)
	app.PUT("/api/games/:gameId/join", srv.joinGame)
	app.POST("/api/games/:gameId/vote", srv.vote) // should it rather be put? patch?
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

// TODO probably we should return names and ids so that the user can decide to join one.
func (s *Server) listGameNames(c *gin.Context) {
	names := s.dealer.ListGameNames()
	c.JSON(http.StatusOK, &names)
}

func (s *Server) getGame(c *gin.Context) {
	id, ok := ParamUint64(c, "gameId")
	if !ok {
		_ = c.AbortWithError(http.StatusBadRequest, ErrBadGameID)
		return
	}
	poker, ok := s.dealer.GetGame(GameID(id))
	if !ok {
		_ = c.AbortWithError(http.StatusNotFound, errGameNotFound(GameID(id)))
		return
	}
	c.JSON(http.StatusOK, &poker)
}

func (s *Server) joinGame(c *gin.Context) {
	var joinReq JoinPokerRequest
	if err := c.BindJSON(&joinReq); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	gameId, ok := ParamUint64(c, "gameId")
	if !ok {
		_ = c.AbortWithError(http.StatusBadRequest, ErrBadGameID) // TODO Should it be abort?
		return
	}
	player, ok := s.playerRegistry.Get(joinReq.PlayerID)
	if !ok {
		_ = c.AbortWithError(http.StatusBadRequest, fmt.Errorf("player with id %d not found", joinReq.PlayerID))
		return
	}

	err := s.dealer.JoinGame(GameID(gameId), player)

	switch err {
	case nil:
		c.Status(http.StatusOK)
	case ErrGameNotFound:
		_ = c.AbortWithError(http.StatusBadRequest, errGameNotFound(GameID(gameId)))
	default:
		_ = c.AbortWithError(http.StatusInternalServerError, err)
	}
}

func (s *Server) vote(c *gin.Context) {
	gameId, ok := ParamUint64(c, "gameId")
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	var voteReq VoteRequest
	if err := c.BindJSON(&voteReq); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err) // TODO remove this later
		return
	}
	err := s.dealer.Vote(GameID(gameId), voteReq)
	switch err {
	case nil:
		c.Status(http.StatusOK)
	case ErrGameNotFound:
		_ = c.AbortWithError(http.StatusBadRequest, errGameNotFound(GameID(gameId)))
	case ErrPlayerNotInGame:
		_ = c.AbortWithError(http.StatusBadRequest, fmt.Errorf("player %d not in game %d", voteReq.PlayerID, gameId))
	default:
		_ = c.AbortWithError(http.StatusInternalServerError, err)
	}
}

// ParamUint64 extracts parameter from gin.Context that is expected to be uint64.
func ParamUint64(c *gin.Context, name string) (uint64, bool) {
	idStr := c.Param(name)
	id, err := strconv.ParseUint(idStr, 10, 0)
	return id, err == nil
}

func errGameNotFound(gameId GameID) error {
	return fmt.Errorf("game with id %d not found", gameId)
}

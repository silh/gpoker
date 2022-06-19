package game

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"
)

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

func (d *Dealer) CreateGame(c *gin.Context) {
	d.lock.Lock()
	defer d.lock.Unlock()
	var gameReq CreatePokerRequest
	if err := c.Bind(&gameReq); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	poker := Poker{
		ID:      d.nextGameID,
		Players: map[PlayerID]Player{gameReq.Player.ID: gameReq.Player},
		Name:    gameReq.GameName,
	}
	d.nextGameID++
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
	poker, ok := d.games[GameID(id)]
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
	game, ok := d.games[GameID(id)]
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

package game_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"gpoker/gen"
	"gpoker/pkg/game"
	"net/http"
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	srv := game.NewStartedServer()
	defer srv.Stop(context.Background())
	waitForServer(t)

	createDefaultGame(t, createUser(t))
}

func TestCreateAndGetAGame(t *testing.T) {
	srv := game.NewStartedServer()
	defer srv.Stop(context.Background())
	waitForServer(t)
	createDefaultGame(t, createUser(t))

	var poker game.Poker
	resp, err := http.DefaultClient.Get(fmt.Sprintf(fullPath("/api/games/%d"), poker.ID))
	require.NoError(t, err)
	poker = game.Poker{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&poker))
}

func TestListGames(t *testing.T) {
	tests := []struct {
		name                   string
		generateCreatorsIDs    func(t *testing.T) []game.PlayerID
		numberOfGamesPerPlayer int
	}{
		{
			name:                   "zero games",
			generateCreatorsIDs:    func(t *testing.T) []game.PlayerID { return nil },
			numberOfGamesPerPlayer: 0,
		},
		{
			name: "1 game",
			generateCreatorsIDs: func(t *testing.T) []game.PlayerID {
				return []game.PlayerID{createUser(t).ID}
			},
			numberOfGamesPerPlayer: 1,
		},
		{
			name: "2 games 2 creators",
			generateCreatorsIDs: func(t *testing.T) []game.PlayerID {
				return []game.PlayerID{createUser(t).ID, createUser(t).ID}
			},
			numberOfGamesPerPlayer: 1,
		},
		{
			name: "2 games 1 creator",
			generateCreatorsIDs: func(t *testing.T) []game.PlayerID {
				return []game.PlayerID{createUser(t).ID}
			},
			numberOfGamesPerPlayer: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := game.NewStartedServer()
			defer server.Stop(context.Background())

			expectedGameNames := make([]string, 0, test.numberOfGamesPerPlayer)
			for _, creatorID := range test.generateCreatorsIDs(t) {
				for i := 0; i < test.numberOfGamesPerPlayer; i++ {
					req := game.CreatePokerRequest{
						GameName:  gen.RandLowercaseString(),
						CreatorID: creatorID,
					}
					expectedGameNames = append(expectedGameNames, req.GameName)
					createGame(t, req)
				}
			}

			resp, err := http.Get(fullPath("/api/games"))
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			var resGameNames []string
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&resGameNames))
			require.Equal(t, len(expectedGameNames), len(resGameNames))

			require.ElementsMatch(t, expectedGameNames, resGameNames)
		})
	}
}

func createUser(t *testing.T) game.Player {
	req := game.RegisterUserRequest{Name: gen.RandLowercaseString()}
	var buffer bytes.Buffer
	require.NoError(t, json.NewEncoder(&buffer).Encode(&req))
	resp, err := http.Post(fullPath("/signup"), "application/json", &buffer)
	require.NoError(t, err)
	defer resp.Body.Close()
	var player game.Player
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&player))
	require.NotZero(t, player)
	require.Equal(t, req.Name, player.Name)
	return player
}

func createGame(t *testing.T, req game.CreatePokerRequest) {
	var buffer bytes.Buffer
	require.NoError(t, json.NewEncoder(&buffer).Encode(&req))
	resp, err := http.DefaultClient.Post(fullPath("/api/games"), "application/json", &buffer)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var poker game.Poker
	err = json.NewDecoder(resp.Body).Decode(&poker) //not entirely correct...
	require.NoError(t, err)
}

func createDefaultGame(t *testing.T, player game.Player) {
	var createGameReq = game.CreatePokerRequest{
		GameName:  gen.RandLowercaseString(),
		CreatorID: player.ID,
	}
	createGame(t, createGameReq)
}

func waitForServer(t *testing.T) {
	timeout := time.After(4 * time.Second)
	var err error
	var resp *http.Response
	for done := false; !done; {
		select {
		case <-timeout:
			t.Fatalf("Server didn't start in time, last error %s", err)
		default:
			resp, err = http.Get(fullPath("/health")) // TODO fix hardcoded
			if err == nil && resp.StatusCode == http.StatusOK {
				done = true
			}
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func fullPath(apiPath string) string {
	return "http://localhost:8080" + apiPath
}

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
	pokerID := createDefaultGame(t, createUser(t))

	resp, err := http.Get(fmt.Sprintf(fullPath("/api/games/%d"), pokerID))
	require.NoError(t, err)
	var poker game.GameResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&poker))
}

func TestGetGameErrors(t *testing.T) {
	tests := []struct {
		name         string
		gameID       string
		responseCode int
	}{
		{
			name:         "Game doesn't exist",
			gameID:       "0",
			responseCode: http.StatusNotFound,
		},
		{
			name:         "Game id is not a number",
			gameID:       "notanumber",
			responseCode: http.StatusBadRequest,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := game.NewStartedServer()
			defer srv.Stop(context.Background())
			waitForServer(t)
			// use string here because we want to test that only ints are accepted
			resp, err := http.Get(fmt.Sprintf(fullPath("/api/games/%s"), test.gameID))
			require.NoError(t, err)
			require.Equal(t, test.responseCode, resp.StatusCode)
		})
	}
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

			expectedGames := make([]game.GameListEntry, 0, test.numberOfGamesPerPlayer)
			for _, creatorID := range test.generateCreatorsIDs(t) {
				for i := 0; i < test.numberOfGamesPerPlayer; i++ {
					req := game.CreatePokerRequest{
						GameName:  gen.RandLowercaseString(),
						CreatorID: creatorID,
					}
					gameID := createGame(t, req)
					expectedGames = append(expectedGames, game.GameListEntry{
						ID:   gameID,
						Name: req.GameName,
					})
				}
			}

			resp, err := http.Get(fullPath("/api/games"))
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			var resGameNames []game.GameListEntry
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&resGameNames))
			require.Equal(t, len(expectedGames), len(resGameNames))

			require.ElementsMatch(t, expectedGames, resGameNames)
		})
	}
}

func TestJoinGame(t *testing.T) {
	srv := game.NewStartedServer()
	defer srv.Stop(context.Background())
	waitForServer(t)
	creator := createUser(t)
	gameID := createDefaultGame(t, creator)
	joiners := []game.Player{
		createUser(t),
		createUser(t),
	}
	for _, player := range joiners {
		join(t, player, gameID)
	}
	players := make([]game.PlayerResponse, 0, 1+len(joiners))
	players = append(players, game.PlayerResponse{
		ID:   creator.ID,
		Name: creator.Name,
	})
	for _, joinee := range joiners {
		players = append(players, game.PlayerResponse{
			ID:   joinee.ID,
			Name: joinee.Name,
		})
	}

	resp, err := http.Get(fmt.Sprintf(fullPath("/api/games/%d"), gameID))
	require.NoError(t, err)
	var poker game.GameResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&poker))
	require.Equal(t, len(players), len(poker.Players))
	require.ElementsMatch(t, players, poker.Players)
}

func TestVote(t *testing.T) {
	srv := game.NewStartedServer()
	defer srv.Stop(context.Background())
	waitForServer(t)
	creator := createUser(t)
	gameID := createDefaultGame(t, creator)
	players := []game.Player{
		createUser(t),
		createUser(t),
	}
	for _, player := range players {
		join(t, player, gameID)
	}
	expectedPlayers := make([]game.PlayerResponse, 0, 1+len(players))
	expectedPlayers = append(expectedPlayers, game.PlayerResponse{
		ID:   creator.ID,
		Name: creator.Name,
		Vote: game.Vote(gen.RandLowercaseString()),
	})
	for _, joining := range players {
		expectedPlayers = append(expectedPlayers, game.PlayerResponse{
			ID:   joining.ID,
			Name: joining.Name,
			Vote: game.Vote(gen.RandLowercaseString()),
		})
	}

	// Vote
	for _, player := range expectedPlayers {
		vote(t, player, gameID)
	}

	resp, err := http.Get(fmt.Sprintf(fullPath("/api/games/%d"), gameID))
	require.NoError(t, err)
	var poker game.GameResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&poker))
	require.Equal(t, len(expectedPlayers), len(poker.Players))
	require.ElementsMatch(t, expectedPlayers, poker.Players)
}

func join(t *testing.T, player game.Player, gameID game.GameID) {
	var buf bytes.Buffer
	body := game.JoinPokerRequest{PlayerID: player.ID}
	require.NoError(t, json.NewEncoder(&buf).Encode(&body))
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf(fullPath("/api/games/%d/join"), gameID), &buf)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func vote(t *testing.T, player game.PlayerResponse, gameID game.GameID) {
	var buf bytes.Buffer
	request := game.VoteRequest{
		PlayerID: player.ID,
		Vote:     player.Vote,
	}
	require.NoError(t, json.NewEncoder(&buf).Encode(&request))
	resp, err := http.Post(fmt.Sprintf(fullPath("/api/games/%d/vote"), gameID), "application/json", &buf)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	buf.Reset()
}

func createUser(t *testing.T) game.Player {
	req := game.RegisterUserRequest{Name: gen.RandLowercaseString()}
	var buffer bytes.Buffer
	require.NoError(t, json.NewEncoder(&buffer).Encode(&req))
	resp, err := http.Post(fullPath("/api/signup"), "application/json", &buffer)
	require.NoError(t, err)
	defer resp.Body.Close()
	var player game.Player
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&player))
	require.NotZero(t, player)
	require.Equal(t, req.Name, player.Name)
	return player
}

func createGame(t *testing.T, req game.CreatePokerRequest) game.GameID {
	var buffer bytes.Buffer
	require.NoError(t, json.NewEncoder(&buffer).Encode(&req))
	resp, err := http.Post(fullPath("/api/games"), "application/json", &buffer)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var poker game.GameResponse
	err = json.NewDecoder(resp.Body).Decode(&poker) //not entirely correct...
	require.NoError(t, err)
	return poker.ID
}

func createDefaultGame(t *testing.T, player game.Player) game.GameID {
	var createGameReq = game.CreatePokerRequest{
		GameName:  gen.RandLowercaseString(),
		CreatorID: player.ID,
	}
	return createGame(t, createGameReq)
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

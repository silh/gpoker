package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"gpoker/pkg/game"
	"net/http"
	"testing"
	"time"
)

var createGameReq = game.CreatePokerRequest{Player: game.Player{
	ID:   1,
	Name: "Sony",
}}

func TestCreate(t *testing.T) {
	server := game.NewStartedServer()
	defer server.Stop(context.Background())
	waitForServer(t)

	createDefaultGame(t)
}

func TestCreateAndGetAGame(t *testing.T) {
	server := game.NewStartedServer()
	defer server.Stop(context.Background())
	waitForServer(t)
	createDefaultGame(t)

	var poker game.Poker
	resp, err := http.DefaultClient.Get(fmt.Sprintf("http://localhost:8080/api/games/%d", poker.ID))
	require.NoError(t, err)
	poker = game.Poker{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&poker))
}

func TestListGames(t *testing.T) {
	tests := []struct {
		name               string
		createGameRequests []game.CreatePokerRequest
	}{
		{
			name:               "zero games",
			createGameRequests: nil,
		},
		{
			name: "1 game",
			createGameRequests: []game.CreatePokerRequest{
				{
					GameName: "one",
					Player: game.Player{
						ID:   1,
						Name: "Alan",
					},
				},
			},
		},
		{
			name: "2 games",
			createGameRequests: []game.CreatePokerRequest{
				{
					GameName: "one",
					Player: game.Player{
						ID:   1,
						Name: "Alan",
					},
				},
				{
					GameName: "two",
					Player: game.Player{
						ID:   2,
						Name: "Fibo",
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := game.NewStartedServer()
			defer server.Stop(context.Background())

			for _, req := range test.createGameRequests {
				createGame(t, req)
			}

			resp, err := http.Get("http://localhost:8080/api/games")
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			var resGameNames []string
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&resGameNames))
			require.Equal(t, len(test.createGameRequests), len(resGameNames))
			expectedGameNames := make([]string, 0, len(test.createGameRequests))
			for _, g := range test.createGameRequests {
				expectedGameNames = append(expectedGameNames, g.GameName)
			}
			require.ElementsMatch(t, expectedGameNames, resGameNames)
		})
	}
}

func createGame(t *testing.T, req game.CreatePokerRequest) {
	var buffer bytes.Buffer
	require.NoError(t, json.NewEncoder(&buffer).Encode(&req))
	resp, err := http.DefaultClient.Post("http://localhost:8080/api/games", "application/json", &buffer)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var poker game.Poker
	err = json.NewDecoder(resp.Body).Decode(&poker) //not entirely correct...
	require.NoError(t, err)
}

func createDefaultGame(t *testing.T) {
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
			resp, err = http.Get("http://localhost:8080/health") // TODO fix hardcoded
			if err == nil && resp.StatusCode == http.StatusOK {
				done = true
			}
			time.Sleep(1 * time.Millisecond)
		}
	}
}

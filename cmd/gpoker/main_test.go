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

func TestCreate(t *testing.T) {
	server := game.NewServer()
	server.Start()
	defer server.Stop(context.Background())
	waitForServer(t)

	req := game.CreatePokerRequest{Player: game.Player{
		ID:   1,
		Name: "Sony",
	}}
	var buffer bytes.Buffer
	require.NoError(t, json.NewEncoder(&buffer).Encode(&req))
	resp, err := http.Post("http://localhost:8080/api/game", "application/json", &buffer)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var poker game.Poker
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&poker))
	resp, err = http.Get(fmt.Sprintf("http://localhost:8080/api/game/%d", poker.ID))
	require.NoError(t, err)
	poker = game.Poker{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&poker))
	fmt.Printf("id=%d", poker.ID)
}

//func TestCreateAndGetAGame(t *testing.T) {
//	go main()
//
//	waitForServer(t)
//
//	req := game.CreatePokerRequest{Player: game.Player{
//		ID:   1,
//		Name: "Sony",
//	}}
//	var buffer bytes.Buffer
//	json.NewEncoder(&buffer).Encode(&req)
//	resp, err := http.DefaultClient.Post("http://localhost:8080/api/game", "application/json", &buffer)
//	require.NoError(t, err)
//	require.Equal(t, http.StatusCreated, resp.StatusCode)
//	var poker game.Poker
//	err = json.NewDecoder(resp.Body).Decode(&poker) //not entirely correct...
//	require.NoError(t, err)
//	resp, err = http.DefaultClient.Get(fmt.Sprintf("http://localhost:8080/api/game/%d", poker.ID))
//	require.NoError(t, err)
//	poker = game.Poker{}
//	err = json.NewDecoder(resp.Body).Decode(&poker) //not entirely correct...
//	fmt.Printf("id=%d", poker.ID)
//}

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

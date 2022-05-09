package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func TestCreateAndGetAGame(t *testing.T) {
	go main()
	time.Sleep(400 * time.Millisecond) // TODO fix that

	req := CreatePokerRequest{Player: Player{
		ID:   1,
		Name: "Sony",
	}}
	var buffer bytes.Buffer
	json.NewEncoder(&buffer).Encode(&req)
	resp, err := http.DefaultClient.Post("http://localhost:8080/game", "application/json", &buffer)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var poker Poker
	err = json.NewDecoder(resp.Body).Decode(&poker) //not entirely correct...
	require.NoError(t, err)
	resp, err = http.DefaultClient.Get(fmt.Sprintf("http://localhost:8080/game/%d", poker.ID))
	require.NoError(t, err)
	poker = Poker{}
	err = json.NewDecoder(resp.Body).Decode(&poker) //not entirely correct...
	fmt.Printf("id=%d", poker.ID)
}

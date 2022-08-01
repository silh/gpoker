package game_test

import (
	"github.com/stretchr/testify/require"
	"gpoker/pkg/game"
	"testing"
)

func TestCreatePlayer(t *testing.T) {
	registry := game.NewPlayerRegistry()
	name := "bobby"
	player := registry.Register(name)
	require.Equal(t, name, player.Name)
}

func TestPlayersHaveUniqueIDs(t *testing.T) {
	registry := game.NewPlayerRegistry()
	player1 := registry.Register("bobby")
	player2 := registry.Register("bobby")
	require.NotEqual(t, player1, player2)
}

func TestGetPlayer(t *testing.T) {
	registry := game.NewPlayerRegistry()
	name := "bobby"
	registerPlayer := registry.Register(name)
	getPlayer, ok := registry.Get(registerPlayer.ID)
	require.True(t, ok)
	require.Equal(t, registerPlayer, getPlayer)
}

func TestGetNonexistentPlayer(t *testing.T) {
	registry := game.NewPlayerRegistry()
	getPlayer, ok := registry.Get(0)
	require.False(t, ok)
	require.Zero(t, getPlayer)
}

func TestNewPlayerRegistryFromFile(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
	}{
		{
			name:     "1 player",
			filepath: "./testdata/one_player.json",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// TODO more and better tests
			registry, err := game.NewPlayerRegistryFromFile(test.filepath)
			require.NoError(t, err)
			player, ok := registry.Get(1)
			require.True(t, ok)
			require.Equal(t, player.Name, "John")
		})
	}
}

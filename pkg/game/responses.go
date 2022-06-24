package game

type GameResponse struct {
	ID      GameID           `json:"id"`
	Name    string           `json:"name"`
	Players []PlayerResponse `json:"players"`
}

type GameListEntry struct {
	ID   GameID `json:"id"`
	Name string `json:"name"`
}

type PlayerResponse struct {
	ID   PlayerID `json:"id"`
	Name string   `json:"name"`
	Vote Vote     `json:"vote"`
}

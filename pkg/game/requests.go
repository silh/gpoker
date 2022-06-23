package game

// CreatePokerRequest to start a game.
type CreatePokerRequest struct {
	GameName  string   `json:"gameName"`
	CreatorID PlayerID `json:"creatorID"`
}

// JoinPokerRequest to join a game.
type JoinPokerRequest struct {
	PlayerID PlayerID `json:"playerID"`
}

// VoteRequest for player's vote.
type VoteRequest struct {
	PlayerID PlayerID `json:"playerID"`
	Vote     Vote     `json:"Vote"`
}

// RegisterUserRequest to add new user. We don't need passwords for now.
type RegisterUserRequest struct {
	Name string `json:"name"`
}

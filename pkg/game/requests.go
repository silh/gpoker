package game

// CreatePokerRequest to start a game.
type CreatePokerRequest struct {
	GameName  string   `json:"gameName"`
	CreatorID PlayerID `json:"creatorID"`
}

// JoinPokerRequest to join a game.
type JoinPokerRequest struct {
	Player Player `json:"player"`
}

// VoteRequest for player's vote.
type VoteRequest struct {
	Vote Vote `json:"AcceptVote"`
}

// RegisterUserRequest to add new user. We don't need passwords for now.
type RegisterUserRequest struct {
	Name string `json:"name"`
}

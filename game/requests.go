package game

// CreatePokerRequest to start a game.
type CreatePokerRequest struct {
	Player Player `json:"player"`
}

// JoinPokerRequest to join a game.
type JoinPokerRequest struct {
	Player Player `json:"player"`
}

// VoteRequest for player's vote.
type VoteRequest struct {
	Vote uint64 `json:"AcceptVote"`
}

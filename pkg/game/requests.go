package game

// CreatePokerRequest to start a game.
type CreatePokerRequest struct {
	GameName string `json:"gameName"`
	Player   Player `json:"creator"`
}

// JoinPokerRequest to join a game.
type JoinPokerRequest struct {
	Player Player `json:"player"`
}

// VoteRequest for player's vote.
type VoteRequest struct {
	Vote Vote `json:"AcceptVote"`
}

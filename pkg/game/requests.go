package game

// CreatePokerRequest to start a game.
type CreatePokerRequest struct {
	GameName  string   `json:"gameName"`
	CreatorID PlayerID `json:"creatorId"`
}

// JoinPokerRequest to join a game.
type JoinPokerRequest struct {
	PlayerID PlayerID `json:"playerId"` // FIXME need validation on all request
}

// VoteRequest for player's vote.
type VoteRequest struct {
	PlayerID PlayerID `json:"playerId"`
	Vote     Vote     `json:"Vote"`
}

// RegisterUserRequest to add new user. We don't need passwords for now.
type RegisterUserRequest struct {
	Name string `json:"name"`
}

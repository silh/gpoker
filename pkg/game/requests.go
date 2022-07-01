package game

// CreatePokerRequest to start a game.
type CreatePokerRequest struct {
	GameName  string   `json:"gameName" binding:"required"`
	CreatorID PlayerID `json:"creatorId" binding:"required"`
}

// JoinPokerRequest to join a game.
type JoinPokerRequest struct {
	PlayerID PlayerID `json:"playerId" binding:"required"`
}

// VoteRequest for player's vote.
type VoteRequest struct {
	PlayerID PlayerID `json:"playerId" binding:"required"`
	Vote     Vote     `json:"Vote" binding:"required"`
}

// RegisterUserRequest to add new user. We don't need passwords for now.
type RegisterUserRequest struct {
	Name string `json:"name" binding:"required"`
}

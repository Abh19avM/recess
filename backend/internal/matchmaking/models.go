package matchmaking

import (
	"time"
)

// MatchMode represents casual vs ranked play.
type MatchMode string

const (
	ModeCasual MatchMode = "casual"
	ModeRanked MatchMode = "ranked"
)

// MatchTicket represents a player waiting in a matchmaking queue.
type MatchTicket struct {
	TicketID     string    `json:"ticket_id"`
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	AvatarPreset string    `json:"avatar_preset,omitempty"`
	GameType     string    `json:"game_type"`
	Mode         MatchMode `json:"mode"`
	Rating       int       `json:"rating"`
	JoinedAt     time.Time `json:"joined_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// MatchedPlayer represents a participant assigned to a created match.
type MatchedPlayer struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	AvatarPreset string `json:"avatar_preset,omitempty"`
	Rating       int    `json:"rating"`
	SeatNumber   int    `json:"seat_number"`
}

// MatchResult represents a successful pairing of two players into a room.
type MatchResult struct {
	MatchID   string          `json:"match_id"`
	RoomID    string          `json:"room_id"`
	GameType  string          `json:"game_type"`
	Mode      MatchMode       `json:"mode"`
	Player1   MatchedPlayer   `json:"player1"`
	Player2   MatchedPlayer   `json:"player2"`
	MatchedAt time.Time       `json:"matched_at"`
}

// QueueJoinRequest is the HTTP body for entering matchmaking.
type QueueJoinRequest struct {
	GameType string    `json:"game_type"`
	Mode     MatchMode `json:"mode,omitempty"` // Default "casual"
}

// QueueJoinResponse is returned when a player enters the queue or is instantly paired.
type QueueJoinResponse struct {
	Status   string       `json:"status"` // "queued" or "matched"
	TicketID string       `json:"ticket_id"`
	Match    *MatchResult `json:"match,omitempty"`
}

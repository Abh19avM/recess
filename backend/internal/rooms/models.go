package rooms

import "time"

// RoomStatus defines the current state of a multiplayer room.
type RoomStatus string

const (
	StatusWaiting    RoomStatus = "waiting"
	StatusInProgress RoomStatus = "in_progress"
	StatusCompleted  RoomStatus = "completed"
)

// PlayerSlot represents a player seated in a room.
type PlayerSlot struct {
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	AvatarPreset string    `json:"avatar_preset"`
	Rating       int       `json:"rating"`
	IsHost       bool      `json:"is_host"`
	IsReady      bool      `json:"is_ready"`
	JoinedAt     time.Time `json:"joined_at"`
}

// Room represents a real-time game room.
type Room struct {
	ID         string       `json:"id"`
	Code       string       `json:"code"` // e.g. "RECESS-7X9P"
	GameType   string       `json:"game_type"`
	Title      string       `json:"title"`
	HostID     string       `json:"host_id"`
	Status     RoomStatus   `json:"status"`
	MaxPlayers int          `json:"max_players"`
	IsPrivate  bool         `json:"is_private"`
	Passcode   string       `json:"-"`
	Players    []PlayerSlot `json:"players"`
	Spectators int          `json:"spectators"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

// CreateRoomRequest represents the payload to create a new custom game room.
type CreateRoomRequest struct {
	GameType   string `json:"game_type"`
	Title      string `json:"title,omitempty"`
	IsPrivate  bool   `json:"is_private"`
	Passcode   string `json:"passcode,omitempty"`
	MaxPlayers int    `json:"max_players,omitempty"`
}

// JoinRoomRequest represents the payload to join a room.
type JoinRoomRequest struct {
	Passcode string `json:"passcode,omitempty"`
}

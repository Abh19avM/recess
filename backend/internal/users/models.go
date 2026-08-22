package users

import "time"

// User represents a player account in Recess.
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email,omitempty"`
	PasswordHash string    `json:"-"`
	IsGuest      bool      `json:"is_guest"`
	AvatarPreset string    `json:"avatar_preset"`
	Title        string    `json:"title"`
	Rating       int       `json:"rating"` // Overall Elo rating
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserProfile represents the public profile and statistics for a player.
type UserProfile struct {
	User        User       `json:"user"`
	GamesPlayed int        `json:"games_played"`
	GamesWon    int        `json:"games_won"`
	WinRate     float64    `json:"win_rate"`
	GameStats   []GameStat `json:"game_stats"`
}

// GameStat represents statistics for a specific school game.
type GameStat struct {
	GameType string `json:"game_type"`
	Rating   int    `json:"rating"`
	Played   int    `json:"played"`
	Won      int    `json:"won"`
	Lost     int    `json:"lost"`
	Drawn    int    `json:"drawn"`
}

package auth

import "github.com/Abh19avM/recess/internal/users"

// RegisterRequest represents the payload for registering a new player.
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents the credentials required to log in.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// GuestRequest represents an optional payload for creating a guest student profile.
type GuestRequest struct {
	Nickname     string `json:"nickname,omitempty"`
	AvatarPreset string `json:"avatar_preset,omitempty"`
}

// AuthResponse returns the minted JWT token alongside user details.
type AuthResponse struct {
	Token string     `json:"token"`
	User  users.User `json:"user"`
}

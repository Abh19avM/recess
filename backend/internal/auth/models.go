package auth

import "time"

// RegisterRequest represents the payload for registering a new player.
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password"`
}

// LoginRequest represents the credentials required to log in.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RefreshRequest represents the payload to exchange a refresh token for new credentials.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutRequest represents the payload to revoke an active refresh token.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// GuestRequest represents an optional payload for creating a guest student profile.
type GuestRequest struct {
	Nickname     string `json:"nickname,omitempty"`
	AvatarPreset string `json:"avatar_preset,omitempty"`
}

// SafeUser represents public user information safe to expose over the API (excluding password hash).
type SafeUser struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email,omitempty"`
	IsGuest      bool      `json:"is_guest"`
	AvatarPreset string    `json:"avatar_preset"`
	Title        string    `json:"title"`
	Rating       int       `json:"rating"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AuthResponse returns the access/refresh token pair alongside safe user details.
type AuthResponse struct {
	Tokens TokenPair `json:"tokens"`
	User   SafeUser  `json:"user"`
}

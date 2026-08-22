package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/Abh19avM/recess/internal/users"
	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	AccessTokenDuration  = 15 * time.Minute
	RefreshTokenDuration = 7 * 24 * time.Hour
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrWrongType    = errors.New("invalid token type")
)

// UserClaims defines the claims embedded in JWT tokens.
type UserClaims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	IsGuest   bool   `json:"is_guest"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// TokenPair represents the access and refresh token returned upon authentication.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Access token TTL in seconds
	TokenType    string `json:"token_type"` // "Bearer"
}

// JWTManager handles signing and verification of JWTs.
type JWTManager struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewJWTManager creates a new JWT manager instance.
func NewJWTManager(secretKey string) *JWTManager {
	if secretKey == "" {
		secretKey = "recess_jwt_default_development_secret_key_change_in_production"
	}
	return &JWTManager{
		secretKey:     []byte(secretKey),
		accessExpiry:  AccessTokenDuration,
		refreshExpiry: RefreshTokenDuration,
	}
}

// GenerateTokenPair mints a new short-lived access JWT and long-lived refresh JWT for a user.
func (m *JWTManager) GenerateTokenPair(user *users.User) (*TokenPair, string, error) {
	now := time.Now().UTC()

	// Access Token
	accessClaims := UserClaims{
		UserID:    user.ID,
		Username:  user.Username,
		IsGuest:   user.IsGuest,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        generateJTI(),
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessExpiry)),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "recess-api",
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(m.secretKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh Token
	refreshJTI := generateJTI()
	refreshClaims := UserClaims{
		UserID:    user.ID,
		Username:  user.Username,
		IsGuest:   user.IsGuest,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshJTI,
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshExpiry)),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "recess-api",
		},
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(m.secretKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	pair := &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(m.accessExpiry.Seconds()),
		TokenType:    "Bearer",
	}

	return pair, refreshJTI, nil
}

// ValidateClaims parses and validates a signed JWT string against the expected token type returning full UserClaims.
func (m *JWTManager) ValidateClaims(tokenString string, expectedType string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if expectedType != "" && claims.TokenType != expectedType {
		return nil, ErrWrongType
	}

	return claims, nil
}

// ValidateToken fulfills the middleware.TokenValidator interface.
func (m *JWTManager) ValidateToken(tokenString string, expectedType string) (userID, username string, isGuest bool, err error) {
	claims, err := m.ValidateClaims(tokenString, expectedType)
	if err != nil {
		return "", "", false, err
	}
	return claims.UserID, claims.Username, claims.IsGuest, nil
}

func generateJTI() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

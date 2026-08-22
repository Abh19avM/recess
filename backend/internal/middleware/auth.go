package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Abh19avM/recess/internal/httputil"
)

type userContextKey string

const (
	// UserClaimsContextKey stores the authenticated user's claims in request context.
	UserClaimsContextKey userContextKey = "user_claims"
	// UserIDContextKey stores the authenticated user's ID in request context.
	UserIDContextKey userContextKey = "user_id"
	// UsernameContextKey stores the authenticated user's username in request context.
	UsernameContextKey userContextKey = "username"
)

// AuthenticatedClaims represents the minimal claims extracted by the middleware.
type AuthenticatedClaims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	IsGuest   bool   `json:"is_guest"`
	TokenType string `json:"token_type"`
}

// TokenValidator defines the contract for verifying access tokens in middleware.
type TokenValidator interface {
	ValidateToken(tokenString string, expectedType string) (userID, username string, isGuest bool, err error)
}

// RequireAuth creates a middleware that strictly requires a valid Bearer access token.
func RequireAuth(validator TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				httputil.ErrorJSON(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				httputil.ErrorJSON(w, r, http.StatusUnauthorized, "INVALID_TOKEN_FORMAT", "authorization header must be in 'Bearer <token>' format")
				return
			}

			tokenString := parts[1]
			userID, username, isGuest, err := validator.ValidateToken(tokenString, "access")
			if err != nil {
				httputil.ErrorJSON(w, r, http.StatusUnauthorized, "INVALID_TOKEN", "access token is invalid or expired")
				return
			}

			claims := &AuthenticatedClaims{
				UserID:    userID,
				Username:  username,
				IsGuest:   isGuest,
				TokenType: "access",
			}

			ctx := context.WithValue(r.Context(), UserClaimsContextKey, claims)
			ctx = context.WithValue(ctx, UserIDContextKey, userID)
			ctx = context.WithValue(ctx, UsernameContextKey, username)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAuthenticatedUser extracts the claims from context if present.
func GetAuthenticatedUser(ctx context.Context) (*AuthenticatedClaims, bool) {
	claims, ok := ctx.Value(UserClaimsContextKey).(*AuthenticatedClaims)
	return claims, ok
}

// GetAuthenticatedUserID extracts the user ID from context if present.
func GetAuthenticatedUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(UserIDContextKey).(string)
	return id, ok
}

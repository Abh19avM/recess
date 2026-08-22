package auth

import (
	"testing"
	"time"

	"github.com/Abh19avM/recess/internal/users"
)

func TestJWTGenerationAndValidation(t *testing.T) {
	jwtMgr := NewJWTManager("super_secret_test_key_1234567890")

	user := &users.User{
		ID:        "usr_test_123",
		Username:  "PencilMaster",
		IsGuest:   false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	pair, jti, err := jwtMgr.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("expected non-empty access and refresh tokens")
	}

	if jti == "" {
		t.Fatal("expected non-empty refresh token JTI")
	}

	// Validate Access Token
	claims, err := jwtMgr.ValidateClaims(pair.AccessToken, TokenTypeAccess)
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}

	if claims.UserID != user.ID || claims.Username != user.Username {
		t.Errorf("claims mismatch: got user %s, %s", claims.UserID, claims.Username)
	}

	if claims.TokenType != TokenTypeAccess {
		t.Errorf("expected access token type, got %s", claims.TokenType)
	}

	// Access token should fail when checked as refresh
	_, err = jwtMgr.ValidateClaims(pair.AccessToken, TokenTypeRefresh)
	if err == nil {
		t.Error("expected error when validating access token as refresh token")
	}

	// Validate Refresh Token
	refreshClaims, err := jwtMgr.ValidateClaims(pair.RefreshToken, TokenTypeRefresh)
	if err != nil {
		t.Fatalf("failed to validate refresh token: %v", err)
	}

	if refreshClaims.UserID != user.ID {
		t.Errorf("expected user ID %s, got %s", user.ID, refreshClaims.UserID)
	}
}

func TestJWTInvalidSignature(t *testing.T) {
	jwtMgr1 := NewJWTManager("key_one")
	jwtMgr2 := NewJWTManager("key_two")

	user := &users.User{ID: "usr_1", Username: "User1"}
	pair, _, _ := jwtMgr1.GenerateTokenPair(user)

	// Validating token signed by key1 using key2 should fail
	_, err := jwtMgr2.ValidateClaims(pair.AccessToken, TokenTypeAccess)
	if err == nil {
		t.Errorf("expected validation failure with mismatched secret keys")
	}
}

package auth

import (
	"testing"
)

func TestHashAndVerifyPassword(t *testing.T) {
	password := "SecretClassroomPass123!"

	// Use lightweight params for fast unit testing
	testParams := &Argon2Params{
		Memory:      16 * 1024,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}

	hash, err := HashPasswordWithParams(password, testParams)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if hash == "" {
		t.Fatal("expected non-empty hash string")
	}

	// Verify matching password
	match, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("failed to verify password: %v", err)
	}
	if !match {
		t.Errorf("expected password to match hash")
	}

	// Verify non-matching password
	match, err = VerifyPassword("WrongPassword456", hash)
	if err != nil {
		t.Fatalf("error during wrong password verification: %v", err)
	}
	if match {
		t.Errorf("expected wrong password to NOT match hash")
	}
}

func TestVerifyInvalidHash(t *testing.T) {
	_, err := VerifyPassword("test", "invalid-hash-format")
	if err == nil {
		t.Errorf("expected error when verifying malformed hash string")
	}
}

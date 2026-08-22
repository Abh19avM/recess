package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/Abh19avM/recess/internal/users"
)

var (
	// ErrInvalidCredentials indicates that the supplied password or username is incorrect.
	ErrInvalidCredentials = errors.New("invalid username or password")
)

// Service defines authentication and token management operations.
type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	CreateGuest(ctx context.Context, req GuestRequest) (*AuthResponse, error)
}

type service struct {
	userRepo users.Repository
}

// NewService creates a new auth service with injected user repository.
func NewService(userRepo users.Repository) Service {
	return &service{userRepo: userRepo}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, fmt.Errorf("username and password are required")
	}

	newUser := &users.User{
		ID:           "usr_" + generateRandomID(8),
		Username:     req.Username,
		Email:        req.Email,
		IsGuest:      false,
		AvatarPreset: "pencil_sketch_1",
		Title:        "New Student",
		Rating:       1200,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: "jwt_token_" + newUser.ID,
		User:  *newUser,
	}, nil
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return &AuthResponse{
		Token: "jwt_token_" + user.ID,
		User:  *user,
	}, nil
}

func (s *service) CreateGuest(ctx context.Context, req GuestRequest) (*AuthResponse, error) {
	nickname := req.Nickname
	if nickname == "" {
		nickname = "RecessGuest_" + generateRandomID(4)
	}

	avatar := req.AvatarPreset
	if avatar == "" {
		avatar = "pencil_sketch_guest"
	}

	guestUser := &users.User{
		ID:           "guest_" + generateRandomID(8),
		Username:     nickname,
		IsGuest:      true,
		AvatarPreset: avatar,
		Title:        "Hall Monitor Guest",
		Rating:       1000,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_ = s.userRepo.Create(ctx, guestUser)

	return &AuthResponse{
		Token: "jwt_guest_token_" + guestUser.ID,
		User:  *guestUser,
	}, nil
}

func generateRandomID(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

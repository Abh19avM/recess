package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Abh19avM/recess/internal/users"
)

var (
	// ErrInvalidCredentials indicates that the supplied password or username is incorrect.
	ErrInvalidCredentials = errors.New("invalid username or password")
	// ErrInvalidRefreshToken indicates that the refresh token is expired, malformed, or revoked.
	ErrInvalidRefreshToken = errors.New("invalid or revoked refresh token")
	// ErrUserNotFound indicates the user associated with a token no longer exists.
	ErrUserNotFound = errors.New("user account not found")

	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// Service defines authentication and token lifecycle operations.
type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	CreateGuest(ctx context.Context, req GuestRequest) (*AuthResponse, error)
}

type service struct {
	userRepo   users.Repository
	jwtManager *JWTManager
	tokenStore TokenStore
}

// NewService creates a new auth service with injected repositories and token manager.
func NewService(userRepo users.Repository, jwtManager *JWTManager, tokenStore TokenStore) Service {
	return &service{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		tokenStore: tokenStore,
	}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	username := strings.TrimSpace(req.Username)
	if !usernameRegex.MatchString(username) {
		return nil, fmt.Errorf("username must be 3-30 characters and contain only alphanumeric letters, dashes, or underscores")
	}

	email := strings.TrimSpace(req.Email)
	if email != "" && !emailRegex.MatchString(email) {
		return nil, fmt.Errorf("invalid email address format")
	}

	if len(req.Password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}

	// Check if user already exists
	if _, err := s.userRepo.GetByUsername(ctx, username); err == nil {
		return nil, users.ErrUserAlreadyExists
	}
	if email != "" {
		if _, err := s.userRepo.GetByEmail(ctx, email); err == nil {
			return nil, users.ErrUserAlreadyExists
		}
	}

	// Hash password with Argon2id
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to secure password: %w", err)
	}

	newUser := &users.User{
		ID:           "usr_" + generateRandomID(8),
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		IsGuest:      false,
		AvatarPreset: "pencil_sketch_1",
		Title:        "Classroom Rookie",
		Rating:       1200,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	tokenPair, refreshJTI, err := s.jwtManager.GenerateTokenPair(newUser)
	if err != nil {
		return nil, err
	}

	// Store refresh token in store
	_ = s.tokenStore.StoreRefreshToken(ctx, newUser.ID, refreshJTI, time.Now().Add(RefreshTokenDuration))

	return &AuthResponse{
		Tokens: *tokenPair,
		User:   toSafeUser(newUser),
	}, nil
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	username := strings.TrimSpace(req.Username)
	if username == "" || req.Password == "" {
		return nil, ErrInvalidCredentials
	}

	var user *users.User
	var err error

	if strings.Contains(username, "@") {
		user, err = s.userRepo.GetByEmail(ctx, username)
	} else {
		user, err = s.userRepo.GetByUsername(ctx, username)
	}

	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.PasswordHash == "" {
		return nil, ErrInvalidCredentials
	}

	// Verify Argon2id password hash
	match, err := VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !match {
		return nil, ErrInvalidCredentials
	}

	tokenPair, refreshJTI, err := s.jwtManager.GenerateTokenPair(user)
	if err != nil {
		return nil, err
	}

	_ = s.tokenStore.StoreRefreshToken(ctx, user.ID, refreshJTI, time.Now().Add(RefreshTokenDuration))

	return &AuthResponse{
		Tokens: *tokenPair,
		User:   toSafeUser(user),
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	if refreshToken == "" {
		return nil, ErrInvalidRefreshToken
	}

	claims, err := s.jwtManager.ValidateClaims(refreshToken, TokenTypeRefresh)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Verify that token JTI is active and not revoked
	valid, err := s.tokenStore.IsRefreshTokenValid(ctx, claims.ID)
	if err != nil || !valid {
		return nil, ErrInvalidRefreshToken
	}

	// Invalidate previous refresh token (token rotation)
	_ = s.tokenStore.RevokeRefreshToken(ctx, claims.ID)

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	newTokenPair, newRefreshJTI, err := s.jwtManager.GenerateTokenPair(user)
	if err != nil {
		return nil, err
	}

	_ = s.tokenStore.StoreRefreshToken(ctx, user.ID, newRefreshJTI, time.Now().Add(RefreshTokenDuration))

	return &AuthResponse{
		Tokens: *newTokenPair,
		User:   toSafeUser(user),
	}, nil
}

func (s *service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}

	claims, err := s.jwtManager.ValidateClaims(refreshToken, TokenTypeRefresh)
	if err != nil {
		return nil // Graceful no-op if token is already invalid
	}

	return s.tokenStore.RevokeRefreshToken(ctx, claims.ID)
}

func (s *service) CreateGuest(ctx context.Context, req GuestRequest) (*AuthResponse, error) {
	nickname := strings.TrimSpace(req.Nickname)
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
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	_ = s.userRepo.Create(ctx, guestUser)

	tokenPair, refreshJTI, err := s.jwtManager.GenerateTokenPair(guestUser)
	if err != nil {
		return nil, err
	}

	_ = s.tokenStore.StoreRefreshToken(ctx, guestUser.ID, refreshJTI, time.Now().Add(RefreshTokenDuration))

	return &AuthResponse{
		Tokens: *tokenPair,
		User:   toSafeUser(guestUser),
	}, nil
}

func toSafeUser(u *users.User) SafeUser {
	return SafeUser{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		IsGuest:      u.IsGuest,
		AvatarPreset: u.AvatarPreset,
		Title:        u.Title,
		Rating:       u.Rating,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func generateRandomID(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

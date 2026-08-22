package users

import (
	"context"
	"fmt"
)

// Service defines user business logic operations.
type Service interface {
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserProfile(ctx context.Context, id string) (*UserProfile, error)
}

type service struct {
	repo Repository
}

// NewService creates a new user service with injected repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetUserByID(ctx context.Context, id string) (*User, error) {
	if id == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *service) GetUserProfile(ctx context.Context, id string) (*UserProfile, error) {
	if id == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	return s.repo.GetProfile(ctx, id)
}

package users

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrUserNotFound indicates that the requested user does not exist.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists indicates that a user with the same username or email already exists.
	ErrUserAlreadyExists = errors.New("user already exists")
)

// Repository defines the data access contract for users.
type Repository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
	GetProfile(ctx context.Context, id string) (*UserProfile, error)
}

// PostgresRepository implements Repository backed by PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository returns a new PostgreSQL user repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*User, error) {
	// Scaffolding: in-memory or database query in Phase 5
	return nil, ErrUserNotFound
}

func (r *PostgresRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	return nil, ErrUserNotFound
}

func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	return nil, ErrUserNotFound
}

func (r *PostgresRepository) Create(ctx context.Context, user *User) error {
	return nil
}

func (r *PostgresRepository) GetProfile(ctx context.Context, id string) (*UserProfile, error) {
	return nil, ErrUserNotFound
}

// InMemoryRepository provides an in-memory implementation for testing and development.
type InMemoryRepository struct {
	mu    sync.RWMutex
	users map[string]*User
}

// NewInMemoryRepository creates an in-memory repository with pre-seeded demo user.
func NewInMemoryRepository() *InMemoryRepository {
	repo := &InMemoryRepository{
		users: make(map[string]*User),
	}

	// Seed demo player
	demoUser := &User{
		ID:           "usr_demo_headmaster",
		Username:     "SchoolChampion",
		Email:        "champion@recess.local",
		IsGuest:      false,
		AvatarPreset: "pencil_sketch_1",
		Title:        "Prefect of Games",
		Rating:       1500,
		CreatedAt:    time.Now().Add(-24 * time.Hour),
		UpdatedAt:    time.Now(),
	}
	repo.users[demoUser.ID] = demoUser
	repo.users[demoUser.Username] = demoUser

	return repo
}

func (r *InMemoryRepository) GetByID(ctx context.Context, id string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (r *InMemoryRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[username]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (r *InMemoryRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (r *InMemoryRepository) Create(ctx context.Context, user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; exists {
		return ErrUserAlreadyExists
	}
	if _, exists := r.users[user.Username]; exists {
		return ErrUserAlreadyExists
	}

	r.users[user.ID] = user
	r.users[user.Username] = user
	return nil
}

func (r *InMemoryRepository) GetProfile(ctx context.Context, id string) (*UserProfile, error) {
	user, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &UserProfile{
		User:        *user,
		GamesPlayed: 28,
		GamesWon:    19,
		WinRate:     67.8,
		GameStats: []GameStat{
			{GameType: "hand_cricket", Rating: 1540, Played: 10, Won: 7, Lost: 3, Drawn: 0},
			{GameType: "dots_boxes", Rating: 1490, Played: 8, Won: 5, Lost: 3, Drawn: 0},
			{GameType: "xo", Rating: 1510, Played: 5, Won: 4, Lost: 1, Drawn: 0},
			{GameType: "connect4", Rating: 1475, Played: 5, Won: 3, Lost: 2, Drawn: 0},
		},
	}, nil
}

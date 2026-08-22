package rooms

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrRoomNotFound indicates that the requested room does not exist.
	ErrRoomNotFound = errors.New("room not found")
	// ErrRoomFull indicates that the room cannot accept more players.
	ErrRoomFull = errors.New("room is full")
	// ErrInvalidPasscode indicates that the supplied passcode is incorrect.
	ErrInvalidPasscode = errors.New("invalid room passcode")
)

// Repository defines data storage operations for game rooms.
type Repository interface {
	Create(ctx context.Context, room *Room) error
	GetByCode(ctx context.Context, code string) (*Room, error)
	Update(ctx context.Context, room *Room) error
	Delete(ctx context.Context, code string) error
	ListActive(ctx context.Context) ([]*Room, error)
}

// PostgresRepository implements Repository for PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, room *Room) error {
	return nil
}

func (r *PostgresRepository) GetByCode(ctx context.Context, code string) (*Room, error) {
	return nil, ErrRoomNotFound
}

func (r *PostgresRepository) Update(ctx context.Context, room *Room) error {
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, code string) error {
	return nil
}

func (r *PostgresRepository) ListActive(ctx context.Context) ([]*Room, error) {
	return []*Room{}, nil
}

// InMemoryRepository provides an in-memory implementation for testing and development.
type InMemoryRepository struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewInMemoryRepository() *InMemoryRepository {
	repo := &InMemoryRepository{
		rooms: make(map[string]*Room),
	}

	// Seed an active public demo room
	demoRoom := &Room{
		ID:         "rm_demo_cricket",
		Code:       "RECESS-CRIC",
		GameType:   "hand_cricket",
		Title:      "Recess Cricket Championship",
		HostID:     "usr_demo_headmaster",
		Status:     StatusWaiting,
		MaxPlayers: 2,
		IsPrivate:  false,
		Players: []PlayerSlot{
			{
				UserID:       "usr_demo_headmaster",
				Username:     "SchoolChampion",
				AvatarPreset: "pencil_sketch_1",
				Rating:       1500,
				IsHost:       true,
				IsReady:      true,
				JoinedAt:     time.Now().Add(-5 * time.Minute),
			},
		},
		Spectators: 0,
		CreatedAt:  time.Now().Add(-5 * time.Minute),
		UpdatedAt:  time.Now(),
	}
	repo.rooms[demoRoom.Code] = demoRoom

	return repo
}

func (r *InMemoryRepository) Create(ctx context.Context, room *Room) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.rooms[room.Code] = room
	return nil
}

func (r *InMemoryRepository) GetByCode(ctx context.Context, code string) (*Room, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	room, ok := r.rooms[code]
	if !ok {
		return nil, ErrRoomNotFound
	}
	return room, nil
}

func (r *InMemoryRepository) Update(ctx context.Context, room *Room) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.rooms[room.Code] = room
	return nil
}

func (r *InMemoryRepository) Delete(ctx context.Context, code string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.rooms, code)
	return nil
}

func (r *InMemoryRepository) ListActive(ctx context.Context) ([]*Room, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*Room, 0, len(r.rooms))
	for _, room := range r.rooms {
		if !room.IsPrivate && room.Status != StatusCompleted {
			list = append(list, room)
		}
	}
	return list, nil
}

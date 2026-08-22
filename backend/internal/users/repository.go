package users

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Abh19avM/recess/internal/database/dbgen"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

// PostgresRepository implements Repository backed by PostgreSQL and sqlc.
type PostgresRepository struct {
	pool    *pgxpool.Pool
	queries *dbgen.Queries
}

// NewPostgresRepository returns a new PostgreSQL user repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool:    pool,
		queries: dbgen.New(pool),
	}
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*User, error) {
	var uid pgtype.UUID
	if err := uid.Scan(id); err != nil {
		return nil, ErrUserNotFound
	}

	u, err := r.queries.GetUserByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &User{
		ID:        uidToString(u.ID),
		Username:  u.Username,
		Email:     textToString(u.Email),
		IsGuest:   u.IsGuest,
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}, nil
}

func (r *PostgresRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	u, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &User{
		ID:           uidToString(u.ID),
		Username:     u.Username,
		Email:        textToString(u.Email),
		PasswordHash: textToString(u.PasswordHash),
		IsGuest:      u.IsGuest,
		CreatedAt:    u.CreatedAt.Time,
		UpdatedAt:    u.UpdatedAt.Time,
	}, nil
}

func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var emailText pgtype.Text
	_ = emailText.Scan(email)

	u, err := r.queries.GetUserByEmail(ctx, emailText)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &User{
		ID:           uidToString(u.ID),
		Username:     u.Username,
		Email:        textToString(u.Email),
		PasswordHash: textToString(u.PasswordHash),
		IsGuest:      u.IsGuest,
		CreatedAt:    u.CreatedAt.Time,
		UpdatedAt:    u.UpdatedAt.Time,
	}, nil
}

func (r *PostgresRepository) Create(ctx context.Context, user *User) error {
	var emailText, passText pgtype.Text
	if user.Email != "" {
		_ = emailText.Scan(user.Email)
	}
	if user.PasswordHash != "" {
		_ = passText.Scan(user.PasswordHash)
	}

	created, err := r.queries.CreateUser(ctx, dbgen.CreateUserParams{
		Username:     user.Username,
		Email:        emailText,
		PasswordHash: passText,
		IsGuest:      user.IsGuest,
	})
	if err != nil {
		return err
	}

	user.ID = uidToString(created.ID)
	user.CreatedAt = created.CreatedAt.Time
	user.UpdatedAt = created.UpdatedAt.Time

	// Create profile
	_, err = r.queries.CreateProfile(ctx, dbgen.CreateProfileParams{
		UserID:       created.ID,
		Nickname:     user.Username,
		AvatarPreset: user.AvatarPreset,
		Title:        user.Title,
		Bio:          "",
	})
	return err
}

func (r *PostgresRepository) GetProfile(ctx context.Context, id string) (*UserProfile, error) {
	var uid pgtype.UUID
	if err := uid.Scan(id); err != nil {
		return nil, ErrUserNotFound
	}

	p, err := r.queries.GetProfileByUserID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &UserProfile{
		User: User{
			ID:           uidToString(p.UserID),
			Username:     p.Username,
			Email:        textToString(p.Email),
			IsGuest:      p.IsGuest,
			AvatarPreset: p.AvatarPreset,
			Title:        p.Title,
		},
		GamesPlayed: 0,
		GamesWon:    0,
		WinRate:     0,
		GameStats:   []GameStat{},
	}, nil
}

func uidToString(uid pgtype.UUID) string {
	if !uid.Valid {
		return ""
	}
	var src [16]byte = uid.Bytes
	return string(encodeUUID(src))
}

func textToString(t pgtype.Text) string {
	if t.Valid {
		return t.String
	}
	return ""
}

func encodeUUID(src [16]byte) []byte {
	const hexDigits = "0123456789abcdef"
	var dst [36]byte
	dst[0] = hexDigits[src[0]>>4]
	dst[1] = hexDigits[src[0]&0xf]
	dst[2] = hexDigits[src[1]>>4]
	dst[3] = hexDigits[src[1]&0xf]
	dst[4] = hexDigits[src[2]>>4]
	dst[5] = hexDigits[src[2]&0xf]
	dst[6] = hexDigits[src[3]>>4]
	dst[7] = hexDigits[src[3]&0xf]
	dst[8] = '-'
	dst[9] = hexDigits[src[4]>>4]
	dst[10] = hexDigits[src[4]&0xf]
	dst[11] = hexDigits[src[5]>>4]
	dst[12] = hexDigits[src[5]&0xf]
	dst[13] = '-'
	dst[14] = hexDigits[src[6]>>4]
	dst[15] = hexDigits[src[6]&0xf]
	dst[16] = hexDigits[src[7]>>4]
	dst[17] = hexDigits[src[7]&0xf]
	dst[18] = '-'
	dst[19] = hexDigits[src[8]>>4]
	dst[20] = hexDigits[src[8]&0xf]
	dst[21] = hexDigits[src[9]>>4]
	dst[22] = hexDigits[src[9]&0xf]
	dst[23] = '-'
	dst[24] = hexDigits[src[10]>>4]
	dst[25] = hexDigits[src[10]&0xf]
	dst[26] = hexDigits[src[11]>>4]
	dst[27] = hexDigits[src[11]&0xf]
	dst[28] = hexDigits[src[12]>>4]
	dst[29] = hexDigits[src[12]&0xf]
	dst[30] = hexDigits[src[13]>>4]
	dst[31] = hexDigits[src[13]&0xf]
	dst[32] = hexDigits[src[14]>>4]
	dst[33] = hexDigits[src[14]&0xf]
	dst[34] = hexDigits[src[15]>>4]
	dst[35] = hexDigits[src[15]&0xf]
	return dst[:]
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

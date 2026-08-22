package auth

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenStore defines operations for storing, checking, and revoking active refresh tokens.
type TokenStore interface {
	StoreRefreshToken(ctx context.Context, userID, tokenID string, expiresAt time.Time) error
	RevokeRefreshToken(ctx context.Context, tokenID string) error
	IsRefreshTokenValid(ctx context.Context, tokenID string) (bool, error)
}

// RedisTokenStore implements TokenStore backed by Redis.
type RedisTokenStore struct {
	rdb *redis.Client
}

// NewRedisTokenStore creates a new Redis token store.
func NewRedisTokenStore(rdb *redis.Client) *RedisTokenStore {
	return &RedisTokenStore{rdb: rdb}
}

func (s *RedisTokenStore) StoreRefreshToken(ctx context.Context, userID, tokenID string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}
	return s.rdb.Set(ctx, "refresh_token:"+tokenID, userID, ttl).Err()
}

func (s *RedisTokenStore) RevokeRefreshToken(ctx context.Context, tokenID string) error {
	return s.rdb.Del(ctx, "refresh_token:"+tokenID).Err()
}

func (s *RedisTokenStore) IsRefreshTokenValid(ctx context.Context, tokenID string) (bool, error) {
	exists, err := s.rdb.Exists(ctx, "refresh_token:"+tokenID).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// InMemoryTokenStore provides an in-memory token store for testing and fallback.
type InMemoryTokenStore struct {
	mu     sync.RWMutex
	tokens map[string]tokenEntry
}

type tokenEntry struct {
	userID    string
	expiresAt time.Time
}

// NewInMemoryTokenStore creates a new in-memory token store.
func NewInMemoryTokenStore() *InMemoryTokenStore {
	store := &InMemoryTokenStore{
		tokens: make(map[string]tokenEntry),
	}
	// Start periodic cleaner for expired tokens
	go store.cleaner()
	return store
}

func (s *InMemoryTokenStore) StoreRefreshToken(ctx context.Context, userID, tokenID string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tokens[tokenID] = tokenEntry{
		userID:    userID,
		expiresAt: expiresAt,
	}
	return nil
}

func (s *InMemoryTokenStore) RevokeRefreshToken(ctx context.Context, tokenID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tokens, tokenID)
	return nil
}

func (s *InMemoryTokenStore) IsRefreshTokenValid(ctx context.Context, tokenID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.tokens[tokenID]
	if !ok {
		return false, nil
	}
	if time.Now().After(entry.expiresAt) {
		return false, nil
	}
	return true, nil
}

func (s *InMemoryTokenStore) cleaner() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for id, entry := range s.tokens {
			if now.After(entry.expiresAt) {
				delete(s.tokens, id)
			}
		}
		s.mu.Unlock()
	}
}

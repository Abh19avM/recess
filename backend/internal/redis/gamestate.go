package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Abh19avM/recess/internal/metrics"
	"github.com/redis/go-redis/v9"
)

// GameCacheStore handles fast Redis caching for live matches.
type GameCacheStore struct {
	rdb *redis.Client
}

// NewGameCacheStore creates a new GameCacheStore instance.
func NewGameCacheStore(rdb *redis.Client) *GameCacheStore {
	return &GameCacheStore{rdb: rdb}
}

// SaveGameState writes an authoritative game snapshot to Redis with TTL.
func (s *GameCacheStore) SaveGameState(ctx context.Context, gameID string, state any, customTTL ...time.Duration) error {
	if s.rdb == nil {
		return nil
	}
	defer metrics.ObserveRedis("save_game_state", time.Now())

	ttl := TTLActiveGame
	if len(customTTL) > 0 && customTTL[0] > 0 {
		ttl = customTTL[0]
	}

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal game state: %w", err)
	}

	key := GameKey(gameID)
	return s.rdb.Set(ctx, key, data, ttl).Err()
}

// GetGameState retrieves a cached game snapshot from Redis.
func (s *GameCacheStore) GetGameState(ctx context.Context, gameID string, dest any) (bool, error) {
	if s.rdb == nil {
		return false, nil
	}
	defer metrics.ObserveRedis("get_game_state", time.Now())

	key := GameKey(gameID)
	val, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil // Cache miss
		}
		return false, fmt.Errorf("failed to read game state from redis: %w", err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return false, fmt.Errorf("failed to unmarshal game state: %w", err)
	}

	return true, nil
}

// DeleteGameState removes a finished or abandoned game from cache.
func (s *GameCacheStore) DeleteGameState(ctx context.Context, gameID string) error {
	if s.rdb == nil {
		return nil
	}
	key := GameKey(gameID)
	return s.rdb.Del(ctx, key).Err()
}

package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// LeaderboardEntry represents a ranked player in a game's leaderboard.
type LeaderboardEntry struct {
	UserID string  `json:"user_id"`
	Score  float64 `json:"score"`
	Rank   int64   `json:"rank"`
}

// LeaderboardStore manages Redis Sorted Sets (ZSET) for high-performance rankings.
type LeaderboardStore struct {
	rdb *redis.Client
}

// NewLeaderboardStore creates a new LeaderboardStore.
func NewLeaderboardStore(rdb *redis.Client) *LeaderboardStore {
	return &LeaderboardStore{rdb: rdb}
}

// UpdateScore updates or increments a player's score on a game's leaderboard.
func (s *LeaderboardStore) UpdateScore(ctx context.Context, gameType, userID string, score float64) error {
	if s.rdb == nil {
		return nil
	}
	key := LeaderboardKey(gameType)
	return s.rdb.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: userID,
	}).Err()
}

// IncrementScore increments a player's rating or score by delta.
func (s *LeaderboardStore) IncrementScore(ctx context.Context, gameType, userID string, delta float64) (float64, error) {
	if s.rdb == nil {
		return 0, nil
	}
	key := LeaderboardKey(gameType)
	return s.rdb.ZIncrBy(ctx, key, delta, userID).Result()
}

// GetTop returns the top N players on a game's leaderboard (highest scores first).
func (s *LeaderboardStore) GetTop(ctx context.Context, gameType string, limit int64) ([]LeaderboardEntry, error) {
	if s.rdb == nil {
		return []LeaderboardEntry{}, nil
	}
	key := LeaderboardKey(gameType)
	zEntries, err := s.rdb.ZRevRangeWithScores(ctx, key, 0, limit-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top leaderboard: %w", err)
	}

	results := make([]LeaderboardEntry, len(zEntries))
	for i, ze := range zEntries {
		results[i] = LeaderboardEntry{
			UserID: fmt.Sprintf("%v", ze.Member),
			Score:  ze.Score,
			Rank:   int64(i + 1),
		}
	}
	return results, nil
}

// GetUserRank returns a player's 1-based ranking on a game's leaderboard.
func (s *LeaderboardStore) GetUserRank(ctx context.Context, gameType, userID string) (rank int64, score float64, found bool, err error) {
	if s.rdb == nil {
		return 0, 0, false, nil
	}
	key := LeaderboardKey(gameType)

	zRank, err := s.rdb.ZRevRank(ctx, key, userID).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, 0, false, nil
		}
		return 0, 0, false, err
	}

	zScore, err := s.rdb.ZScore(ctx, key, userID).Result()
	if err != nil {
		return 0, 0, false, err
	}

	return zRank + 1, zScore, true, nil
}

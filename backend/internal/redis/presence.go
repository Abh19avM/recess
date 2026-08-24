package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Abh19avM/recess/internal/metrics"
	"github.com/redis/go-redis/v9"
)

// UserPresence represents a user's active session state in Redis.
type UserPresence struct {
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	AvatarPreset string    `json:"avatar_preset"`
	CurrentRoom  string    `json:"current_room,omitempty"`
	LastSeen     time.Time `json:"last_seen"`
}

// PresenceStore handles Redis operations for student presence.
type PresenceStore struct {
	rdb *redis.Client
}

// NewPresenceStore creates a new PresenceStore.
func NewPresenceStore(rdb *redis.Client) *PresenceStore {
	return &PresenceStore{rdb: rdb}
}

// SetOnline marks a user as online with a sliding TTL.
func (s *PresenceStore) SetOnline(ctx context.Context, presence UserPresence) error {
	if s.rdb == nil {
		return nil
	}
	defer metrics.ObserveRedis("presence_set_online", time.Now())

	presence.LastSeen = time.Now().UTC()
	data, err := json.Marshal(presence)
	if err != nil {
		return fmt.Errorf("failed to marshal presence: %w", err)
	}

	key := PresenceKey(presence.UserID)
	return s.rdb.Set(ctx, key, data, TTLPresence).Err()
}

// Heartbeat refreshes the TTL of an active user's presence.
func (s *PresenceStore) Heartbeat(ctx context.Context, userID string) error {
	if s.rdb == nil {
		return nil
	}
	defer metrics.ObserveRedis("presence_heartbeat", time.Now())
	key := PresenceKey(userID)
	return s.rdb.Expire(ctx, key, TTLPresence).Err()
}

// SetOffline removes a user's presence immediately upon logout or disconnect.
func (s *PresenceStore) SetOffline(ctx context.Context, userID string) error {
	if s.rdb == nil {
		return nil
	}
	defer metrics.ObserveRedis("presence_set_offline", time.Now())
	key := PresenceKey(userID)
	return s.rdb.Del(ctx, key).Err()
}

// GetPresence retrieves the active presence record for a user.
func (s *PresenceStore) GetPresence(ctx context.Context, userID string) (*UserPresence, error) {
	if s.rdb == nil {
		return nil, nil
	}

	key := PresenceKey(userID)
	val, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // User is offline
		}
		return nil, fmt.Errorf("failed to get presence: %w", err)
	}

	var p UserPresence
	if err := json.Unmarshal([]byte(val), &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal presence: %w", err)
	}

	return &p, nil
}

// IsOnline checks whether a user key exists in Redis.
func (s *PresenceStore) IsOnline(ctx context.Context, userID string) (bool, error) {
	if s.rdb == nil {
		return false, nil
	}
	key := PresenceKey(userID)
	n, err := s.rdb.Exists(ctx, key).Result()
	return n > 0, err
}

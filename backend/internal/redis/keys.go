package redis

import (
	"fmt"
	"time"
)

// Key Naming Conventions & Namespace Formats
const (
	// Presence & Session Keys
	KeyPrefixPresence = "online:user:" // online:user:{user_id} -> Hash/JSON presence data
	KeyPrefixSession  = "session:jwt:" // session:jwt:{token_hash} -> session metadata

	// Game & Room State Keys
	KeyPrefixGame = "game:" // game:{game_id} -> Cached GameState JSON
	KeyPrefixRoom = "room:" // room:{room_id} -> Cached Room JSON

	// Queue & Matchmaking Keys
	KeyPrefixQueue = "queue:" // queue:{game_type} -> Sorted Set or List of waiting tickets

	// Rate Limiting Keys
	KeyPrefixRateUser = "rate:user:" // rate:user:{user_id} -> sliding window counter
	KeyPrefixRateIP   = "rate:ip:"   // rate:ip:{ip_address} -> sliding window counter

	// Leaderboard & Rating Keys
	KeyPrefixLeaderboard = "leaderboard:" // leaderboard:{game_type} -> ZSET of user_id -> rating

	// Pub/Sub Channel Namespaces
	ChannelPrefixRoom = "pubsub:room:" // pubsub:room:{room_id} -> room broadcast channel
	ChannelEvents     = "pubsub:events" // global system broadcast channel
)

// Default TTL Strategy Constants
const (
	TTLPresence       = 90 * time.Second   // Refreshed by 30s ping heartbeats
	TTLActiveGame     = 2 * time.Hour      // Expiration for active matches
	TTLActiveRoom     = 24 * time.Hour     // Expiration for idle rooms
	TTLRefreshToken   = 7 * 24 * time.Hour // Session token lifetime
	TTLRateLimitWindow = 1 * time.Minute   // 1-minute rate limiting window
)

// Key Formatters

// PresenceKey returns the Redis key for user online presence.
func PresenceKey(userID string) string {
	return fmt.Sprintf("%s%s", KeyPrefixPresence, userID)
}

// GameKey returns the Redis key for cached game state.
func GameKey(gameID string) string {
	return fmt.Sprintf("%s%s", KeyPrefixGame, gameID)
}

// RoomKey returns the Redis key for cached room state.
func RoomKey(roomID string) string {
	return fmt.Sprintf("%s%s", KeyPrefixRoom, roomID)
}

// QueueKey returns the Redis key for a game's matchmaking queue.
func QueueKey(gameType string) string {
	return fmt.Sprintf("%s%s", KeyPrefixQueue, gameType)
}

// RateUserKey returns the Redis key for user-level rate limiting.
func RateUserKey(userID string) string {
	return fmt.Sprintf("%s%s", KeyPrefixRateUser, userID)
}

// RateIPKey returns the Redis key for IP-level rate limiting.
func RateIPKey(ip string) string {
	return fmt.Sprintf("%s%s", KeyPrefixRateIP, ip)
}

// LeaderboardKey returns the Redis sorted set key for a game leaderboard.
func LeaderboardKey(gameType string) string {
	return fmt.Sprintf("%s%s", KeyPrefixLeaderboard, gameType)
}

// RoomChannel returns the Pub/Sub channel name for a room.
func RoomChannel(roomID string) string {
	return fmt.Sprintf("%s%s", ChannelPrefixRoom, roomID)
}

package websocket

import (
	"encoding/json"
	"errors"
	"time"
)

// EventType represents the standard event classification.
type EventType string

const (
	// Core Lifecycle Events
	EventRoomJoin           EventType = "room.join"
	EventRoomLeave          EventType = "room.leave"
	EventRoomState          EventType = "room.state"
	EventPlayerJoined       EventType = "player.joined"
	EventPlayerLeft         EventType = "player.left"
	EventPlayerReady        EventType = "player.ready"
	EventPlayerDisconnected EventType = "player.disconnected"
	EventError              EventType = "room.error"
	EventPing               EventType = "room.ping"
	EventPong               EventType = "room.pong"

	// Broadcast / Custom Event
	EventMessage EventType = "room.message"

	// Game Engine Events
	EventGameStart   EventType = "game.start"
	EventGameMove    EventType = "game.move"
	EventGameState   EventType = "game.state"
	EventGameRematch EventType = "game.rematch"
	EventGameConfig  EventType = "game.config"

	// Session Reconnection Events
	EventSessionReconnect   EventType = "session.reconnect"
	EventSessionReconnected EventType = "session.reconnected"
	EventPlayerReconnecting EventType = "player.reconnecting"
	EventPlayerReconnected  EventType = "player.reconnected"

	// Spectator Mode Events
	EventSpectatorJoined EventType = "spectator.joined"
	EventSpectatorLeft   EventType = "spectator.left"
	EventSpectatorCount  EventType = "spectator.count"
)

// EventEnvelope is the uniform message format for all WebSocket communications.
type EventEnvelope struct {
	Type      EventType       `json:"type"`
	RoomID    string          `json:"room_id"`
	Sequence  int64           `json:"sequence,omitempty"`
	Timestamp int64           `json:"timestamp,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

// PlayerInfo holds public snapshot information about a connected player in a room.
type PlayerInfo struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	IsGuest      bool   `json:"is_guest"`
	AvatarPreset string `json:"avatar_preset,omitempty"`
	IsReady      bool   `json:"is_ready"`
}

// RoomJoinPayload is sent by a client requesting to join a room.
type RoomJoinPayload struct {
	Passcode string `json:"passcode,omitempty"`
	Role     string `json:"role,omitempty"` // "player" or "spectator"
}

// RoomStatePayload is sent to a newly joined client detailing the room's current state.
type RoomStatePayload struct {
	RoomID         string       `json:"room_id"`
	Members        []PlayerInfo `json:"members"`
	SpectatorCount int          `json:"spectator_count"`
}

// SpectatorCountPayload is broadcast when the spectator count changes.
type SpectatorCountPayload struct {
	RoomID string `json:"room_id"`
	Count  int    `json:"count"`
}

// SpectatorJoinedPayload is broadcast when a spectator joins the sideline.
type SpectatorJoinedPayload struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Count    int    `json:"count"`
}

// SpectatorLeftPayload is broadcast when a spectator departs the sideline.
type SpectatorLeftPayload struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Count    int    `json:"count"`
}

// PlayerJoinedPayload is broadcast to room members when a new player joins.
type PlayerJoinedPayload struct {
	Player PlayerInfo `json:"player"`
}

// PlayerLeftPayload is broadcast to room members when a player deliberately leaves.
type PlayerLeftPayload struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Reason   string `json:"reason,omitempty"`
}

// PlayerReadyPayload is sent and broadcast when a player changes their ready state.
type PlayerReadyPayload struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsReady  bool   `json:"is_ready"`
}

// PlayerDisconnectedPayload is broadcast when a player's socket abruptly drops.
type PlayerDisconnectedPayload struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// ErrorPayload contains structured error information.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// MessagePayload represents a generic chat or broadcast message in a room.
type MessagePayload struct {
	SenderID string `json:"sender_id"`
	Sender   string `json:"sender"`
	Text     string `json:"text"`
}

// SessionReconnectPayload is sent by a reconnecting client.
type SessionReconnectPayload struct {
	SessionID    string `json:"session_id"`
	RoomID       string `json:"room_id"`
	LastSequence int64  `json:"last_sequence"`
}

// SessionReconnectedPayload is returned to a reconnected client with state snapshot and missed events.
type SessionReconnectedPayload struct {
	SessionID       string           `json:"session_id"`
	RoomID          string           `json:"room_id"`
	CurrentSequence int64            `json:"current_sequence"`
	MissedEvents    []*EventEnvelope `json:"missed_events"`
	GameState       any              `json:"game_state,omitempty"`
	Members         []PlayerInfo     `json:"members"`
}

// PlayerReconnectingPayload is broadcast when a player abruptly disconnects during an active game.
type PlayerReconnectingPayload struct {
	UserID             string `json:"user_id"`
	Username           string `json:"username"`
	GracePeriodSeconds int    `json:"grace_period_seconds"`
}

// PlayerReconnectedPayload is broadcast when a disconnected player resumes their session.
type PlayerReconnectedPayload struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// DistributedEnvelope wraps an EventEnvelope with origin instance metadata for Redis Pub/Sub routing.
type DistributedEnvelope struct {
	OriginInstanceID string         `json:"origin_instance_id"`
	RoomID           string         `json:"room_id"`
	Envelope         *EventEnvelope `json:"envelope"`
}

// NewEnvelope creates a new EventEnvelope with the current Unix timestamp in milliseconds.
func NewEnvelope(eventType EventType, roomID string, sequence int64, payload any) (*EventEnvelope, error) {
	var rawPayload json.RawMessage
	if payload != nil {
		bytes, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		rawPayload = bytes
	}

	return &EventEnvelope{
		Type:      eventType,
		RoomID:    roomID,
		Sequence:  sequence,
		Timestamp: time.Now().UTC().UnixMilli(),
		Payload:   rawPayload,
	}, nil
}

// MustEnvelope creates an envelope without returning an error (panics if marshaling fails).
func MustEnvelope(eventType EventType, roomID string, sequence int64, payload any) *EventEnvelope {
	env, err := NewEnvelope(eventType, roomID, sequence, payload)
	if err != nil {
		panic(err)
	}
	return env
}

// Validate checks that the envelope contains required fields.
func (e *EventEnvelope) Validate() error {
	if e.Type == "" {
		return errors.New("event type is required")
	}
	if e.RoomID == "" && e.Type != EventPing && e.Type != EventPong {
		return errors.New("room_id is required")
	}
	return nil
}

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
}

// RoomStatePayload is sent to a newly joined client detailing the room's current state.
type RoomStatePayload struct {
	RoomID  string       `json:"room_id"`
	Members []PlayerInfo `json:"members"`
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

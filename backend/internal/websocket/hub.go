package websocket

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"
)

// Hub maintains all active client connections, room membership, and message routing.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client            // clientID -> *Client
	rooms   map[string]map[string]*Client // roomID -> clientID -> *Client
	roomSeq map[string]int64              // roomID -> sequence counter
}

// NewHub constructs a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
		rooms:   make(map[string]map[string]*Client),
		roomSeq: make(map[string]int64),
	}
}

// Register adds a new client to the hub registry.
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c.ID] = c
	slog.Debug("websocket client registered", "client_id", c.ID, "user_id", c.UserID, "username", c.Username)
}

// Unregister removes a client and broadcasts disconnect / leave events to any active room.
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	delete(h.clients, c.ID)
	roomID := c.GetRoomID()

	var remainingInRoom []*Client
	if roomID != "" && h.rooms[roomID] != nil {
		delete(h.rooms[roomID], c.ID)
		if len(h.rooms[roomID]) == 0 {
			delete(h.rooms, roomID)
			delete(h.roomSeq, roomID)
		} else {
			for _, member := range h.rooms[roomID] {
				remainingInRoom = append(remainingInRoom, member)
			}
		}
	}
	h.mu.Unlock()

	// Broadcast disconnect event to remaining room members
	if len(remainingInRoom) > 0 {
		disconnectEnv, _ := NewEnvelope(
			EventPlayerDisconnected,
			roomID,
			h.nextSeq(roomID),
			PlayerDisconnectedPayload{
				UserID:   c.UserID,
				Username: c.Username,
			},
		)
		for _, member := range remainingInRoom {
			member.Send(disconnectEnv)
		}
	}

	slog.Debug("websocket client unregistered", "client_id", c.ID, "user_id", c.UserID)
}

// JoinRoom adds a client to the specified room, delivers room state, and notifies existing members.
func (h *Hub) JoinRoom(c *Client, roomID string) {
	if roomID == "" {
		h.SendError(c, "", "INVALID_ROOM", "room_id cannot be empty")
		return
	}

	h.mu.Lock()
	// If currently in a different room, leave first
	oldRoom := c.GetRoomID()
	if oldRoom != "" && oldRoom != roomID && h.rooms[oldRoom] != nil {
		delete(h.rooms[oldRoom], c.ID)
		if len(h.rooms[oldRoom]) == 0 {
			delete(h.rooms, oldRoom)
			delete(h.roomSeq, oldRoom)
		}
	}

	if _, exists := h.rooms[roomID]; !exists {
		h.rooms[roomID] = make(map[string]*Client)
	}

	c.SetRoomID(roomID)
	h.rooms[roomID][c.ID] = c

	// Collect members snapshot
	var members []PlayerInfo
	var otherMembers []*Client
	for _, member := range h.rooms[roomID] {
		members = append(members, member.PlayerInfo())
		if member.ID != c.ID {
			otherMembers = append(otherMembers, member)
		}
	}

	h.roomSeq[roomID]++
	joinSeq := h.roomSeq[roomID]
	h.mu.Unlock()

	// 1. Send room.state to the joining client
	stateEnv, err := NewEnvelope(EventRoomState, roomID, joinSeq, RoomStatePayload{
		RoomID:  roomID,
		Members: members,
	})
	if err == nil {
		c.Send(stateEnv)
	}

	// 2. Broadcast player.joined to all other members in the room
	joinedEnv, err := NewEnvelope(EventPlayerJoined, roomID, joinSeq, PlayerJoinedPayload{
		Player: c.PlayerInfo(),
	})
	if err == nil {
		for _, member := range otherMembers {
			member.Send(joinedEnv)
		}
	}

	slog.Info("player joined room", "room_id", roomID, "user_id", c.UserID, "username", c.Username)
}

// LeaveRoom removes a client from their active room and notifies peers.
func (h *Hub) LeaveRoom(c *Client) {
	roomID := c.GetRoomID()
	if roomID == "" {
		return
	}

	h.mu.Lock()
	var otherMembers []*Client
	if h.rooms[roomID] != nil {
		delete(h.rooms[roomID], c.ID)
		if len(h.rooms[roomID]) == 0 {
			delete(h.rooms, roomID)
			delete(h.roomSeq, roomID)
		} else {
			for _, member := range h.rooms[roomID] {
				otherMembers = append(otherMembers, member)
			}
		}
	}
	c.SetRoomID("")
	c.SetReady(false)
	h.mu.Unlock()

	if len(otherMembers) > 0 {
		leftEnv, _ := NewEnvelope(
			EventPlayerLeft,
			roomID,
			h.nextSeq(roomID),
			PlayerLeftPayload{
				UserID:   c.UserID,
				Username: c.Username,
				Reason:   "left",
			},
		)
		for _, member := range otherMembers {
			member.Send(leftEnv)
		}
	}

	slog.Info("player left room", "room_id", roomID, "user_id", c.UserID)
}

// ToggleReady updates the client's readiness and broadcasts to all room members.
func (h *Hub) ToggleReady(c *Client, isReady bool) {
	roomID := c.GetRoomID()
	if roomID == "" {
		h.SendError(c, "", "NOT_IN_ROOM", "cannot toggle ready status when not in a room")
		return
	}

	c.SetReady(isReady)

	readyEnv, _ := NewEnvelope(
		EventPlayerReady,
		roomID,
		h.nextSeq(roomID),
		PlayerReadyPayload{
			UserID:   c.UserID,
			Username: c.Username,
			IsReady:  isReady,
		},
	)

	h.BroadcastToRoom(roomID, readyEnv, "")
}

// BroadcastToRoom broadcasts an envelope to all clients in a room.
func (h *Hub) BroadcastToRoom(roomID string, env *EventEnvelope, excludeClientID string) {
	h.mu.RLock()
	roomClients, exists := h.rooms[roomID]
	if !exists {
		h.mu.RUnlock()
		return
	}

	clients := make([]*Client, 0, len(roomClients))
	for _, c := range roomClients {
		if excludeClientID == "" || c.ID != excludeClientID {
			clients = append(clients, c)
		}
	}
	h.mu.RUnlock()

	for _, client := range clients {
		client.Send(env)
	}
}

// SendError sends a structured room.error event to a specific client.
func (h *Hub) SendError(c *Client, roomID string, code, message string) {
	errEnv, _ := NewEnvelope(
		EventError,
		roomID,
		0,
		ErrorPayload{
			Code:    code,
			Message: message,
		},
	)
	c.Send(errEnv)
}

// HandleMessage parses and routes incoming client envelopes.
func (h *Hub) HandleMessage(c *Client, env *EventEnvelope) {
	if err := env.Validate(); err != nil {
		h.SendError(c, env.RoomID, "INVALID_ENVELOPE", err.Error())
		return
	}

	switch env.Type {
	case EventRoomJoin:
		h.JoinRoom(c, env.RoomID)

	case EventRoomLeave:
		h.LeaveRoom(c)

	case EventPlayerReady:
		var payload PlayerReadyPayload
		if len(env.Payload) > 0 {
			if err := json.Unmarshal(env.Payload, &payload); err == nil {
				h.ToggleReady(c, payload.IsReady)
				return
			}
		}
		// If payload omitted, flip current ready state
		h.ToggleReady(c, !c.GetReady())

	case EventPing:
		pongEnv := &EventEnvelope{
			Type:      EventPong,
			RoomID:    env.RoomID,
			Timestamp: time.Now().UTC().UnixMilli(),
		}
		c.Send(pongEnv)

	case EventMessage:
		roomID := c.GetRoomID()
		if roomID == "" || roomID != env.RoomID {
			h.SendError(c, env.RoomID, "NOT_IN_ROOM", "cannot send messages to a room you are not in")
			return
		}

		var msgPayload MessagePayload
		if err := json.Unmarshal(env.Payload, &msgPayload); err != nil {
			h.SendError(c, roomID, "INVALID_PAYLOAD", "invalid message payload")
			return
		}
		msgPayload.SenderID = c.UserID
		msgPayload.Sender = c.Username

		broadcastEnv, err := NewEnvelope(EventMessage, roomID, h.nextSeq(roomID), msgPayload)
		if err == nil {
			h.BroadcastToRoom(roomID, broadcastEnv, "")
		}

	default:
		// For future extensible events (e.g. game moves forwarded through rooms)
		roomID := c.GetRoomID()
		if roomID != "" && roomID == env.RoomID {
			env.Sequence = h.nextSeq(roomID)
			env.Timestamp = time.Now().UTC().UnixMilli()
			h.BroadcastToRoom(roomID, env, "")
		} else {
			h.SendError(c, env.RoomID, "UNRECOGNIZED_EVENT", "unknown or unsupported event type: "+string(env.Type))
		}
	}
}

// ClientCount returns the total number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// RoomMembersCount returns the number of active clients in a room.
func (h *Hub) RoomMembersCount(roomID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if room, exists := h.rooms[roomID]; exists {
		return len(room)
	}
	return 0
}

// nextSeq increments and returns the sequence number for a room.
func (h *Hub) nextSeq(roomID string) int64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.roomSeq[roomID]++
	return h.roomSeq[roomID]
}

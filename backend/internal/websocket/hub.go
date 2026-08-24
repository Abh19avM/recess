package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	_ "github.com/Abh19avM/recess/internal/games/connect4"
	_ "github.com/Abh19avM/recess/internal/games/dotsboxes"
	"github.com/Abh19avM/recess/internal/games/engine"
	_ "github.com/Abh19avM/recess/internal/games/handcricket"
	_ "github.com/Abh19avM/recess/internal/games/npt"
	_ "github.com/Abh19avM/recess/internal/games/paperfootball"
	_ "github.com/Abh19avM/recess/internal/games/xo"
	"github.com/Abh19avM/recess/internal/metrics"
	"github.com/Abh19avM/recess/internal/redis"
	goredis "github.com/redis/go-redis/v9"
)

// DisconnectedSession holds state for a player during the disconnection grace period.
type DisconnectedSession struct {
	SessionID          string
	UserID             string
	Username           string
	IsGuest            bool
	AvatarPreset       string
	RoomID             string
	IsReady            bool
	DisconnectedAt     time.Time
	GracePeriodSeconds int
	Timer              *time.Timer
}

// MaxSpectatorsPerRoom defines the spectator capacity for a single desk match.
const MaxSpectatorsPerRoom = 50

// HubOption configures optional behaviors on the WebSocket Hub.
type HubOption func(*Hub)

// WithRedis enables Redis Pub/Sub cross-instance routing and game caching.
func WithRedis(rdb *goredis.Client) HubOption {
	return func(h *Hub) {
		h.rdb = rdb
		if rdb != nil {
			h.pubsubBroker = redis.NewPubSubBroker(rdb)
			h.gameCache = redis.NewGameCacheStore(rdb)
		}
	}
}

// WithInstanceID sets a custom instance identifier for distributed routing.
func WithInstanceID(instanceID string) HubOption {
	return func(h *Hub) {
		h.instanceID = instanceID
	}
}

// Hub maintains all active client connections, room membership, game state, and message routing.
type Hub struct {
	mu                   sync.RWMutex
	instanceID           string
	rdb                  *goredis.Client
	pubsubBroker         *redis.PubSubBroker
	gameCache            *redis.GameCacheStore
	clients              map[string]*Client               // clientID -> *Client
	rooms                map[string]map[string]*Client    // roomID -> clientID -> *Client (players)
	spectators           map[string]map[string]*Client    // roomID -> clientID -> *Client (spectators)
	roomSeq              map[string]int64                 // roomID -> sequence counter
	roomEngines          map[string]engine.Engine         // roomID -> active GameEngine
	roomTypes            map[string]engine.GameType       // roomID -> designated GameType
	roomHistory          map[string][]*EventEnvelope      // roomID -> ring buffer of recent events
	roomStartTimes       map[string]time.Time             // roomID -> game start timestamp
	disconnectedSessions map[string]*DisconnectedSession  // sessionID -> DisconnectedSession
	roomSubscriptions    map[string]*goredis.PubSub       // roomID -> redis PubSub subscription
	roomMembersCache     map[string]map[string]PlayerInfo // roomID -> userID -> PlayerInfo
}

// NewHub constructs a new Hub instance with optional Redis Pub/Sub integration.
func NewHub(opts ...HubOption) *Hub {
	h := &Hub{
		instanceID:           fmt.Sprintf("inst_%d_%d", time.Now().UnixNano(), os.Getpid()),
		clients:              make(map[string]*Client),
		rooms:                make(map[string]map[string]*Client),
		spectators:           make(map[string]map[string]*Client),
		roomSeq:              make(map[string]int64),
		roomEngines:          make(map[string]engine.Engine),
		roomTypes:            make(map[string]engine.GameType),
		roomHistory:          make(map[string][]*EventEnvelope),
		roomStartTimes:       make(map[string]time.Time),
		disconnectedSessions: make(map[string]*DisconnectedSession),
		roomSubscriptions:    make(map[string]*goredis.PubSub),
		roomMembersCache:     make(map[string]map[string]PlayerInfo),
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Register adds a new client to the hub registry.
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c.ID] = c
	metrics.WebSocketConnections.Inc()
	metrics.ActivePlayers.Set(float64(len(h.clients)))
	slog.Debug("websocket client registered", "client_id", c.ID, "user_id", c.UserID, "username", c.Username)
}

// Unregister removes a client and broadcasts disconnect / leave events to any active room.
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	delete(h.clients, c.ID)
	metrics.WebSocketConnections.Dec()
	metrics.ActivePlayers.Set(float64(len(h.clients)))
	roomID := c.GetRoomID()

	// If spectator disconnects, handle without touching match state or grace period
	if c.GetSpectator() {
		var spectCount int
		if roomID != "" && h.spectators[roomID] != nil {
			delete(h.spectators[roomID], c.ID)
			spectCount = len(h.spectators[roomID])
			if len(h.rooms[roomID]) == 0 && spectCount == 0 {
				delete(h.rooms, roomID)
				delete(h.spectators, roomID)
				delete(h.roomSeq, roomID)
				delete(h.roomEngines, roomID)
				delete(h.roomTypes, roomID)
				delete(h.roomHistory, roomID)
			}
		}
		h.mu.Unlock()

		if roomID != "" {
			spectLeftEnv, _ := NewEnvelope(
				EventSpectatorLeft,
				roomID,
				h.nextSeq(roomID),
				SpectatorLeftPayload{
					UserID:   c.UserID,
					Username: c.Username,
					Count:    spectCount,
				},
			)
			h.BroadcastToRoom(roomID, spectLeftEnv, c.ID)
		}
		slog.Debug("websocket spectator unregistered", "client_id", c.ID, "user_id", c.UserID)
		return
	}

	var remainingInRoom []*Client
	hasActiveGame := false
	if roomID != "" && h.rooms[roomID] != nil {
		delete(h.rooms[roomID], c.ID)
		if eng, hasEng := h.roomEngines[roomID]; hasEng && eng != nil && !eng.IsFinished() {
			hasActiveGame = true
		}

		if len(h.rooms[roomID]) == 0 && len(h.spectators[roomID]) == 0 && !hasActiveGame {
			delete(h.rooms, roomID)
			delete(h.spectators, roomID)
			delete(h.roomSeq, roomID)
			delete(h.roomEngines, roomID)
			delete(h.roomTypes, roomID)
			delete(h.roomHistory, roomID)
		} else {
			for _, member := range h.rooms[roomID] {
				remainingInRoom = append(remainingInRoom, member)
			}
		}
	}

	// If player disconnects during an active match, register grace session!
	if hasActiveGame && c.UserID != "" {
		sessID := c.GetSessionID()
		if sessID == "" {
			sessID = "sess_" + c.ID
		}

		graceTimer := time.AfterFunc(30*time.Second, func() {
			h.handleGracePeriodExpired(sessID, roomID, c.UserID)
		})

		h.disconnectedSessions[sessID] = &DisconnectedSession{
			SessionID:          sessID,
			UserID:             c.UserID,
			Username:           c.Username,
			IsGuest:            c.IsGuest,
			AvatarPreset:       c.AvatarPreset,
			RoomID:             roomID,
			IsReady:            c.IsReady,
			DisconnectedAt:     time.Now().UTC(),
			GracePeriodSeconds: 30,
			Timer:              graceTimer,
		}
	}
	h.mu.Unlock()

	// Broadcast disconnect or reconnecting event to remaining room members
	if len(remainingInRoom) > 0 {
		if hasActiveGame {
			reconnectingEnv, _ := NewEnvelope(
				EventPlayerReconnecting,
				roomID,
				h.nextSeq(roomID),
				PlayerReconnectingPayload{
					UserID:             c.UserID,
					Username:           c.Username,
					GracePeriodSeconds: 30,
				},
			)
			h.BroadcastToRoom(roomID, reconnectingEnv, "")
		} else {
			disconnectEnv, _ := NewEnvelope(
				EventPlayerDisconnected,
				roomID,
				h.nextSeq(roomID),
				PlayerDisconnectedPayload{
					UserID:   c.UserID,
					Username: c.Username,
				},
			)
			h.BroadcastToRoom(roomID, disconnectEnv, "")
		}
	}

	slog.Debug("websocket client unregistered", "client_id", c.ID, "user_id", c.UserID, "has_active_game", hasActiveGame)
}

// JoinRoom adds a client to the specified room.
func (h *Hub) JoinRoom(c *Client, roomID string) {
	h.JoinRoomWithRole(c, roomID, "")
}

// JoinRoomWithRole adds a client with player or spectator role, delivers room state, and notifies members.
func (h *Hub) JoinRoomWithRole(c *Client, roomID string, role string) {
	if roomID == "" {
		h.SendError(c, "", "INVALID_ROOM", "room_id cannot be empty")
		return
	}

	if role == "spectator" {
		c.SetSpectator(true)
	}

	h.mu.Lock()
	// If currently in a different room, leave first
	oldRoom := c.GetRoomID()
	if oldRoom != "" && oldRoom != roomID {
		if c.GetSpectator() && h.spectators[oldRoom] != nil {
			delete(h.spectators[oldRoom], c.ID)
		} else if h.rooms[oldRoom] != nil {
			delete(h.rooms[oldRoom], c.ID)
		}
		if len(h.rooms[oldRoom]) == 0 && len(h.spectators[oldRoom]) == 0 {
			delete(h.rooms, oldRoom)
			delete(h.spectators, oldRoom)
			delete(h.roomSeq, oldRoom)
			delete(h.roomEngines, oldRoom)
			delete(h.roomTypes, oldRoom)
		}
	}

	// 1. Handle Spectator Join
	if c.GetSpectator() {
		if _, exists := h.spectators[roomID]; !exists {
			h.spectators[roomID] = make(map[string]*Client)
		}

		if len(h.spectators[roomID]) >= MaxSpectatorsPerRoom {
			h.mu.Unlock()
			h.SendError(c, roomID, "SPECTATOR_CAPACITY_REACHED", "Spectator capacity reached for this match (max 50).")
			return
		}

		c.SetRoomID(roomID)
		h.spectators[roomID][c.ID] = c

		// Collect members
		var members []PlayerInfo
		for _, member := range h.rooms[roomID] {
			members = append(members, member.PlayerInfo())
		}
		spectatorCount := len(h.spectators[roomID])

		h.roomSeq[roomID]++
		joinSeq := h.roomSeq[roomID]

		var activeGameState *engine.GameState
		if eng, exists := h.roomEngines[roomID]; exists {
			activeGameState = eng.State()
		}
		h.mu.Unlock()

		// Send room.state to spectator
		stateEnv, err := NewEnvelope(EventRoomState, roomID, joinSeq, RoomStatePayload{
			RoomID:         roomID,
			Members:        members,
			SpectatorCount: spectatorCount,
		})
		if err == nil {
			c.Send(stateEnv)
		}

		// Ensure Redis subscription
		h.ensureRoomSubscription(roomID)

		// Send active game.state if available
		if activeGameState != nil {
			gameEnv, err := NewEnvelope(EventGameState, roomID, joinSeq, activeGameState)
			if err == nil {
				c.Send(gameEnv)
			}
		}

		// Broadcast spectator.joined to players and other spectators
		spectatorJoinedEnv, _ := NewEnvelope(EventSpectatorJoined, roomID, joinSeq, SpectatorJoinedPayload{
			UserID:   c.UserID,
			Username: c.Username,
			Count:    spectatorCount,
		})
		h.BroadcastToRoom(roomID, spectatorJoinedEnv, c.ID)

		slog.Info("spectator joined room", "room_id", roomID, "user_id", c.UserID, "username", c.Username)
		return
	}

	// 2. Handle Player Join
	if _, exists := h.rooms[roomID]; !exists {
		h.rooms[roomID] = make(map[string]*Client)
	}

	c.SetRoomID(roomID)
	h.rooms[roomID][c.ID] = c

	// Infer room game type from room code if not yet explicitly set
	if _, exists := h.roomTypes[roomID]; !exists {
		lower := strings.ToLower(roomID)
		if strings.Contains(lower, "cricket") || strings.Contains(lower, "hc") {
			h.roomTypes[roomID] = engine.GameTypeHandCricket
		} else if strings.Contains(lower, "dots") || strings.Contains(lower, "db") || strings.Contains(lower, "box") {
			h.roomTypes[roomID] = engine.GameTypeDotsBoxes
		} else if strings.Contains(lower, "connect") || strings.Contains(lower, "c4") {
			h.roomTypes[roomID] = engine.GameTypeConnect4
		} else if strings.Contains(lower, "football") || strings.Contains(lower, "pf") {
			h.roomTypes[roomID] = engine.GameTypePaperFootball
		} else if strings.Contains(lower, "npat") || strings.Contains(lower, "npt") || strings.Contains(lower, "name") {
			h.roomTypes[roomID] = engine.GameTypeNPAT
		} else {
			h.roomTypes[roomID] = engine.GameTypeXO
		}
	}

	// Collect members snapshot
	var members []PlayerInfo
	for _, member := range h.rooms[roomID] {
		members = append(members, member.PlayerInfo())
	}
	spectatorCount := len(h.spectators[roomID])

	h.roomSeq[roomID]++
	joinSeq := h.roomSeq[roomID]

	// Check if game is currently in progress
	var activeGameState *engine.GameState
	if eng, exists := h.roomEngines[roomID]; exists {
		activeGameState = eng.State()
	}
	h.mu.Unlock()

	// 1. Send room.state to the joining client
	stateEnv, err := NewEnvelope(EventRoomState, roomID, joinSeq, RoomStatePayload{
		RoomID:         roomID,
		Members:        members,
		SpectatorCount: spectatorCount,
	})
	if err == nil {
		c.Send(stateEnv)
	}

	// 2. Ensure this instance is subscribed to Redis Pub/Sub for cross-instance sync
	h.ensureRoomSubscription(roomID)

	// 3. Broadcast player.joined to all other members in the room (locally and across instances)
	joinedEnv, err := NewEnvelope(EventPlayerJoined, roomID, joinSeq, PlayerJoinedPayload{
		Player: c.PlayerInfo(),
	})
	if err == nil {
		h.BroadcastToRoom(roomID, joinedEnv, c.ID)
	}

	// 4. If there is an active game running, send game.state to newly joined client
	if activeGameState != nil {
		gameEnv, err := NewEnvelope(EventGameState, roomID, joinSeq, activeGameState)
		if err == nil {
			c.Send(gameEnv)
		}
	}

	slog.Info("player joined room", "room_id", roomID, "user_id", c.UserID, "username", c.Username)
}

// LeaveRoom removes a client (player or spectator) from their active room and notifies peers.
func (h *Hub) LeaveRoom(c *Client) {
	roomID := c.GetRoomID()
	if roomID == "" {
		return
	}

	if c.GetSpectator() {
		h.mu.Lock()
		if h.spectators[roomID] != nil {
			delete(h.spectators[roomID], c.ID)
		}
		spectCount := len(h.spectators[roomID])
		roomEmpty := len(h.rooms[roomID]) == 0 && spectCount == 0
		if roomEmpty {
			delete(h.rooms, roomID)
			delete(h.spectators, roomID)
			delete(h.roomSeq, roomID)
			delete(h.roomEngines, roomID)
			delete(h.roomTypes, roomID)
		}
		c.SetRoomID("")
		h.mu.Unlock()

		if roomEmpty {
			h.unsubscribeRoom(roomID)
		}

		spectLeftEnv, _ := NewEnvelope(
			EventSpectatorLeft,
			roomID,
			h.nextSeq(roomID),
			SpectatorLeftPayload{
				UserID:   c.UserID,
				Username: c.Username,
				Count:    spectCount,
			},
		)
		h.BroadcastToRoom(roomID, spectLeftEnv, c.ID)
		slog.Info("spectator left room", "room_id", roomID, "user_id", c.UserID)
		return
	}

	h.mu.Lock()
	roomEmpty := false
	if h.rooms[roomID] != nil {
		delete(h.rooms[roomID], c.ID)
		if len(h.rooms[roomID]) == 0 && len(h.spectators[roomID]) == 0 {
			roomEmpty = true
			delete(h.rooms, roomID)
			delete(h.spectators, roomID)
			delete(h.roomSeq, roomID)
			delete(h.roomEngines, roomID)
			delete(h.roomTypes, roomID)
		}
	}
	c.SetRoomID("")
	c.SetReady(false)
	h.mu.Unlock()

	if roomEmpty {
		h.unsubscribeRoom(roomID)
	}

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
	h.BroadcastToRoom(roomID, leftEnv, c.ID)
	slog.Info("player left room", "room_id", roomID, "user_id", c.UserID)
}

// ToggleReady updates the client's readiness and checks if all members are ready to auto-start.
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

	// Check if all players in room are ready to start a match
	h.mu.Lock()
	members, exists := h.rooms[roomID]
	allReady := false
	if exists && len(members) == 2 {
		readyCount := 0
		for _, m := range members {
			if m.GetReady() {
				readyCount++
			}
		}
		if readyCount == 2 {
			allReady = true
		}
	}
	gType := h.roomTypes[roomID]
	if gType == "" {
		gType = engine.GameTypeXO
	}
	h.mu.Unlock()

	if allReady {
		h.StartGame(roomID, gType)
	}
}

// StartGame initializes and starts a match in the given room.
func (h *Hub) StartGame(roomID string, gameType engine.GameType) {
	h.mu.Lock()
	membersMap, exists := h.rooms[roomID]
	if !exists || len(membersMap) < 2 {
		h.mu.Unlock()
		return
	}

	var players []engine.Player
	for _, m := range membersMap {
		players = append(players, engine.Player{
			ID:       m.UserID,
			Username: m.Username,
			IsBot:    false,
		})
	}

	// Sort players deterministically by UserID
	sort.Slice(players, func(i, j int) bool {
		return players[i].ID < players[j].ID
	})
	for i := range players {
		players[i].SeatNumber = i + 1
	}

	if gameType == "" {
		gameType = h.roomTypes[roomID]
		if gameType == "" {
			gameType = engine.GameTypeXO
		}
	}
	h.roomTypes[roomID] = gameType

	eng, err := engine.Create(gameType)
	if err != nil {
		h.mu.Unlock()
		slog.Error("failed to create engine for game type", "game_type", gameType, "error", err)
		return
	}

	state, err := eng.Initialize(roomID, players, nil)
	if err != nil {
		h.mu.Unlock()
		slog.Error("failed to initialize game engine", "room_id", roomID, "error", err)
		return
	}

	h.roomEngines[roomID] = eng
	h.roomStartTimes[roomID] = time.Now()
	metrics.ActiveMatches.Set(float64(len(h.roomEngines)))
	seq := h.roomSeq[roomID] + 1
	h.roomSeq[roomID] = seq
	h.mu.Unlock()

	gameEnv, _ := NewEnvelope(EventGameState, roomID, seq, state)
	h.BroadcastToRoom(roomID, gameEnv, "")
	slog.Info("game started in room", "room_id", roomID, "game_type", gameType, "players", len(players))
}

// HandleGameMove validates and applies a move to the room's active engine.
func (h *Hub) HandleGameMove(c *Client, env *EventEnvelope) {
	roomID := c.GetRoomID()
	if roomID == "" || roomID != env.RoomID {
		h.SendError(c, env.RoomID, "NOT_IN_ROOM", "cannot move in a room you are not in")
		return
	}

	h.mu.Lock()
	eng, exists := h.roomEngines[roomID]
	if !exists {
		h.mu.Unlock()
		h.SendError(c, roomID, "NO_ACTIVE_GAME", "no active game currently in progress in this room")
		return
	}

	var actionPayload struct {
		Action string `json:"action"`
	}
	_ = json.Unmarshal(env.Payload, &actionPayload)
	action := actionPayload.Action
	if action == "" {
		action = "mark"
	}

	move := engine.Move{
		PlayerID:  c.UserID,
		Action:    action,
		Data:      env.Payload,
		Timestamp: time.Now().UnixMilli(),
	}

	state, err := eng.ApplyMove(move)
	if err != nil {
		h.mu.Unlock()
		h.SendError(c, roomID, "INVALID_MOVE", err.Error())
		return
	}

	if state.Status == engine.StatusFinished {
		if startTime, ok := h.roomStartTimes[roomID]; ok {
			metrics.GameDuration.WithLabelValues(string(state.GameType)).Observe(time.Since(startTime).Seconds())
			delete(h.roomStartTimes, roomID)
		}
		metrics.ActiveMatches.Set(float64(len(h.roomEngines)))
	}

	seq := h.roomSeq[roomID] + 1
	h.roomSeq[roomID] = seq
	h.mu.Unlock()

	// Broadcast authoritative state update to all players in the room
	stateEnv, _ := NewEnvelope(EventGameState, roomID, seq, state)
	h.BroadcastToRoom(roomID, stateEnv, "")
}

// HandleRematch resets the game and starts a new round with the same players.
func (h *Hub) HandleRematch(c *Client, env *EventEnvelope) {
	roomID := c.GetRoomID()
	if roomID == "" {
		return
	}

	h.mu.RLock()
	gType := h.roomTypes[roomID]
	if gType == "" {
		if eng, exists := h.roomEngines[roomID]; exists {
			gType = eng.GameType()
		} else {
			gType = engine.GameTypeXO
		}
	}
	h.mu.RUnlock()

	h.StartGame(roomID, gType)
}

// BroadcastToRoom sends an envelope to all connected clients in a room except excludeClientID,
// including players and spectators, and broadcasts across Redis Pub/Sub to other backend instances.
func (h *Hub) BroadcastToRoom(roomID string, env *EventEnvelope, excludeClientID string) {
	h.mu.Lock()
	roomClients, exists := h.rooms[roomID]
	spectClients, hasSpects := h.spectators[roomID]
	if !exists && !hasSpects && h.pubsubBroker == nil {
		h.mu.Unlock()
		return
	}

	// Append to room history buffer (limit to 100 events for reconnection replay)
	if h.roomHistory[roomID] == nil {
		h.roomHistory[roomID] = make([]*EventEnvelope, 0, 100)
	}
	h.roomHistory[roomID] = append(h.roomHistory[roomID], env)
	if len(h.roomHistory[roomID]) > 100 {
		h.roomHistory[roomID] = h.roomHistory[roomID][len(h.roomHistory[roomID])-100:]
	}

	var clients []*Client
	if exists {
		for _, c := range roomClients {
			if excludeClientID == "" || c.ID != excludeClientID {
				clients = append(clients, c)
			}
		}
	}
	if hasSpects {
		for _, s := range spectClients {
			if excludeClientID == "" || s.ID != excludeClientID {
				clients = append(clients, s)
			}
		}
	}
	h.mu.Unlock()

	for _, client := range clients {
		client.Send(env)
	}
	metrics.WebSocketMessagesTotal.WithLabelValues(string(env.Type), "outbound").Inc()

	// Publish to Redis Pub/Sub channel for other backend instances
	if h.pubsubBroker != nil {
		distEnv := DistributedEnvelope{
			OriginInstanceID: h.instanceID,
			RoomID:           roomID,
			Envelope:         env,
		}
		_ = h.pubsubBroker.Publish(context.Background(), redis.RoomChannel(roomID), distEnv)
	}
}

// ensureRoomSubscription starts a Redis Pub/Sub listener for the room on this instance.
func (h *Hub) ensureRoomSubscription(roomID string) {
	if h.rdb == nil || h.pubsubBroker == nil {
		return
	}
	if h.roomSubscriptions[roomID] != nil {
		return
	}

	channel := redis.RoomChannel(roomID)
	pubsub := h.pubsubBroker.Subscribe(context.Background(), channel)
	if pubsub == nil {
		return
	}
	h.roomSubscriptions[roomID] = pubsub

	go func(rID string, ps *goredis.PubSub) {
		ch := ps.Channel()
		for msg := range ch {
			if msg == nil {
				continue
			}
			var dist DistributedEnvelope
			if err := json.Unmarshal([]byte(msg.Payload), &dist); err != nil {
				continue
			}
			if dist.OriginInstanceID == h.instanceID {
				continue // Avoid echoing local events
			}
			h.handleRemoteEvent(dist.RoomID, dist.Envelope)
		}
	}(roomID, pubsub)
}

// unsubscribeRoom closes the Redis subscription when no local clients remain.
func (h *Hub) unsubscribeRoom(roomID string) {
	if pubsub, exists := h.roomSubscriptions[roomID]; exists && pubsub != nil {
		_ = pubsub.Close()
		delete(h.roomSubscriptions, roomID)
	}
}

// handleRemoteEvent processes and dispatches events received from other backend instances.
func (h *Hub) handleRemoteEvent(roomID string, env *EventEnvelope) {
	if env == nil {
		return
	}

	h.mu.Lock()
	if env.Sequence > h.roomSeq[roomID] {
		h.roomSeq[roomID] = env.Sequence
	}

	if h.roomHistory[roomID] == nil {
		h.roomHistory[roomID] = make([]*EventEnvelope, 0, 100)
	}
	h.roomHistory[roomID] = append(h.roomHistory[roomID], env)
	if len(h.roomHistory[roomID]) > 100 {
		h.roomHistory[roomID] = h.roomHistory[roomID][len(h.roomHistory[roomID])-100:]
	}

	// Update local members cache if player event
	switch env.Type {
	case EventPlayerJoined:
		var p PlayerJoinedPayload
		if err := json.Unmarshal(env.Payload, &p); err == nil && p.Player.UserID != "" {
			if h.roomMembersCache[roomID] == nil {
				h.roomMembersCache[roomID] = make(map[string]PlayerInfo)
			}
			h.roomMembersCache[roomID][p.Player.UserID] = p.Player
		}
	case EventPlayerLeft, EventPlayerDisconnected:
		var p struct {
			UserID string `json:"user_id"`
		}
		if err := json.Unmarshal(env.Payload, &p); err == nil && p.UserID != "" {
			if h.roomMembersCache[roomID] != nil {
				delete(h.roomMembersCache[roomID], p.UserID)
			}
		}
	case EventPlayerReady:
		var p PlayerReadyPayload
		if err := json.Unmarshal(env.Payload, &p); err == nil && p.UserID != "" {
			if h.roomMembersCache[roomID] != nil {
				if info, exists := h.roomMembersCache[roomID][p.UserID]; exists {
					info.IsReady = p.IsReady
					h.roomMembersCache[roomID][p.UserID] = info
				}
			}
		}
	case EventGameState:
		var state engine.GameState
		if err := json.Unmarshal(env.Payload, &state); err == nil && h.gameCache != nil {
			_ = h.gameCache.SaveGameState(context.Background(), roomID, state)
		}
	}

	roomClients := h.rooms[roomID]
	spectClients := h.spectators[roomID]
	clients := make([]*Client, 0, len(roomClients)+len(spectClients))
	for _, c := range roomClients {
		clients = append(clients, c)
	}
	for _, s := range spectClients {
		clients = append(clients, s)
	}
	h.mu.Unlock()

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

	metrics.WebSocketMessagesTotal.WithLabelValues(string(env.Type), "inbound").Inc()

	// Spectators on the sideline cannot make moves or change game state
	if c.GetSpectator() {
		switch env.Type {
		case EventGameMove, EventPlayerReady, EventGameStart, EventGameRematch:
			h.SendError(c, env.RoomID, "SPECTATOR_CANNOT_MOVE", "Spectators on the sideline cannot submit moves or modify game state.")
			return
		}
	}

	switch env.Type {
	case EventRoomJoin:
		var joinPayload RoomJoinPayload
		if len(env.Payload) > 0 {
			_ = json.Unmarshal(env.Payload, &joinPayload)
		}
		h.JoinRoomWithRole(c, env.RoomID, joinPayload.Role)

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
		h.ToggleReady(c, !c.GetReady())

	case EventGameStart:
		var startPayload struct {
			GameType string `json:"game_type"`
		}
		_ = json.Unmarshal(env.Payload, &startPayload)
		gType := engine.GameType(startPayload.GameType)
		h.StartGame(env.RoomID, gType)

	case EventGameMove:
		h.HandleGameMove(c, env)

	case EventGameRematch:
		h.HandleRematch(c, env)

	case EventSessionReconnect:
		h.HandleSessionReconnect(c, env.Payload)

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

// HandleSessionReconnect processes reconnection requests, delivers state snapshots and missed events.
func (h *Hub) HandleSessionReconnect(c *Client, payloadRaw json.RawMessage) {
	var payload SessionReconnectPayload
	if err := json.Unmarshal(payloadRaw, &payload); err != nil {
		h.SendError(c, "", "INVALID_RECONNECT_PAYLOAD", "invalid reconnect payload")
		return
	}

	roomID := payload.RoomID
	if roomID == "" {
		h.SendError(c, "", "MISSING_ROOM_ID", "room_id is required for reconnection")
		return
	}

	h.mu.Lock()
	// Match disconnected session by session ID or UserID in this room
	var session *DisconnectedSession
	if sess, exists := h.disconnectedSessions[payload.SessionID]; exists {
		session = sess
		delete(h.disconnectedSessions, payload.SessionID)
		if sess.Timer != nil {
			sess.Timer.Stop()
		}
	} else {
		for sID, dSess := range h.disconnectedSessions {
			if dSess.RoomID == roomID && (dSess.UserID == c.UserID || dSess.Username == c.Username) {
				session = dSess
				delete(h.disconnectedSessions, sID)
				if dSess.Timer != nil {
					dSess.Timer.Stop()
				}
				break
			}
		}
	}

	if session != nil {
		c.UserID = session.UserID
		c.Username = session.Username
		c.IsGuest = session.IsGuest
		c.AvatarPreset = session.AvatarPreset
		c.IsReady = session.IsReady
		c.SessionID = session.SessionID
	}

	// Restore room membership
	if _, exists := h.rooms[roomID]; !exists {
		h.rooms[roomID] = make(map[string]*Client)
	}
	c.SetRoomID(roomID)
	h.rooms[roomID][c.ID] = c

	// Calculate missed events from roomHistory
	var missedEvents []*EventEnvelope
	if history, exists := h.roomHistory[roomID]; exists {
		for _, env := range history {
			if env.Sequence > payload.LastSequence {
				missedEvents = append(missedEvents, env)
			}
		}
	}

	// Get latest game state snapshot
	var gameState *engine.GameState
	if eng, exists := h.roomEngines[roomID]; exists && eng != nil {
		gameState = eng.State()
	}

	currentSeq := h.roomSeq[roomID]

	// Collect members
	var members []PlayerInfo
	for _, m := range h.rooms[roomID] {
		members = append(members, m.PlayerInfo())
	}
	h.mu.Unlock()

	// Deliver session.reconnected to the client
	reconnectedEnv, err := NewEnvelope(
		EventSessionReconnected,
		roomID,
		currentSeq,
		SessionReconnectedPayload{
			SessionID:       c.GetSessionID(),
			RoomID:          roomID,
			CurrentSequence: currentSeq,
			MissedEvents:    missedEvents,
			GameState:       gameState,
			Members:         members,
		},
	)
	if err == nil {
		c.Send(reconnectedEnv)
	}

	// Broadcast player.reconnected to other members in the room
	reconnectedNotifyEnv, _ := NewEnvelope(
		EventPlayerReconnected,
		roomID,
		h.nextSeq(roomID),
		PlayerReconnectedPayload{
			UserID:   c.UserID,
			Username: c.Username,
		},
	)
	h.BroadcastToRoom(roomID, reconnectedNotifyEnv, c.ID)
}

func (h *Hub) handleGracePeriodExpired(sessionID, roomID, userID string) {
	h.mu.Lock()
	delete(h.disconnectedSessions, sessionID)
	if len(h.rooms[roomID]) == 0 {
		delete(h.rooms, roomID)
		delete(h.roomSeq, roomID)
		delete(h.roomEngines, roomID)
		delete(h.roomTypes, roomID)
		delete(h.roomHistory, roomID)
	}
	h.mu.Unlock()

	slog.Info("grace period expired for disconnected session", "session_id", sessionID, "room_id", roomID, "user_id", userID)
}

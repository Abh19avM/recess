package websocket

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512 KB
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Configured via CORS in production
	},
}

// Client represents a single active WebSocket connection.
type Client struct {
	hub          *Hub
	conn         *websocket.Conn
	send         chan *EventEnvelope
	mu           sync.RWMutex
	ID           string
	UserID       string
	Username     string
	IsGuest      bool
	AvatarPreset string
	RoomID       string
	SessionID    string
	IsReady      bool
	IsSpectator  bool
}

// NewClient initializes a new Client wrapper.
func NewClient(hub *Hub, conn *websocket.Conn, clientID, userID, username string, isGuest bool, avatarPreset string) *Client {
	return &Client{
		hub:          hub,
		conn:         conn,
		send:         make(chan *EventEnvelope, 256),
		ID:           clientID,
		UserID:       userID,
		Username:     username,
		IsGuest:      isGuest,
		AvatarPreset: avatarPreset,
		SessionID:    "sess_" + clientID,
		IsReady:      false,
		IsSpectator:  false,
	}
}

// SetSpectator updates the client's spectator status.
func (c *Client) SetSpectator(isSpectator bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.IsSpectator = isSpectator
}

// GetSpectator returns whether the client is a spectator.
func (c *Client) GetSpectator() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.IsSpectator
}

// PlayerInfo returns a snapshot of the client's public state in a room.
func (c *Client) PlayerInfo() PlayerInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return PlayerInfo{
		UserID:       c.UserID,
		Username:     c.Username,
		IsGuest:      c.IsGuest,
		AvatarPreset: c.AvatarPreset,
		IsReady:      c.IsReady,
	}
}

// SetRoomID updates the client's current room ID.
func (c *Client) SetRoomID(roomID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.RoomID = roomID
}

// GetRoomID returns the current room ID for the client.
func (c *Client) GetRoomID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.RoomID
}

// SetSessionID updates the client's session ID.
func (c *Client) SetSessionID(sessionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SessionID = sessionID
}

// GetSessionID returns the client's session ID.
func (c *Client) GetSessionID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.SessionID
}

// SetReady updates the client's ready status.
func (c *Client) SetReady(isReady bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.IsReady = isReady
}

// GetReady returns the current ready status.
func (c *Client) GetReady() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.IsReady
}

// Send attempts to enqueue an envelope to the client's outbound channel.
func (c *Client) Send(env *EventEnvelope) bool {
	select {
	case c.send <- env:
		return true
	default:
		slog.Warn("client send buffer full, dropping message", "client_id", c.ID, "user_id", c.UserID)
		return false
	}
}

// ReadPump listens for incoming WebSocket frames and dispatches them to the Hub.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var env EventEnvelope
		err := c.conn.ReadJSON(&env)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Debug("websocket connection closed unexpectedly", "client_id", c.ID, "user_id", c.UserID, "error", err)
			}
			break
		}

		c.hub.HandleMessage(c, &env)
	}
}

// WritePump handles flushing messages to the client and sending periodic pings.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case env, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(env); err != nil {
				slog.Debug("failed to write JSON to websocket", "client_id", c.ID, "error", err)
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

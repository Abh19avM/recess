package websocket

import (
	"log/slog"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	mu         sync.RWMutex
	clients    map[string]*Client            // clientID -> Client
	roomTracks map[string]map[string]*Client // roomCode -> (clientID -> Client)
	register   chan *Client
	unregister chan *Client
	broadcast  chan *RoomMessage
}

// RoomMessage represents a payload targeted to a specific room.
type RoomMessage struct {
	RoomCode string `json:"room_code"`
	SenderID string `json:"sender_id"`
	Event    string `json:"event"`
	Data     any    `json:"data"`
}

// NewHub creates a new WebSocket Hub instance.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		roomTracks: make(map[string]map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *RoomMessage, 256),
	}
}

// Run starts the event loop for managing WebSocket client connections.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			if client.RoomCode != "" {
				if _, ok := h.roomTracks[client.RoomCode]; !ok {
					h.roomTracks[client.RoomCode] = make(map[string]*Client)
				}
				h.roomTracks[client.RoomCode][client.ID] = client
			}
			h.mu.Unlock()
			slog.Debug("websocket client registered", "client_id", client.ID, "room", client.RoomCode)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.send)
				if client.RoomCode != "" && h.roomTracks[client.RoomCode] != nil {
					delete(h.roomTracks[client.RoomCode], client.ID)
					if len(h.roomTracks[client.RoomCode]) == 0 {
						delete(h.roomTracks, client.RoomCode)
					}
				}
			}
			h.mu.Unlock()
			slog.Debug("websocket client unregistered", "client_id", client.ID, "room", client.RoomCode)

		case msg := <-h.broadcast:
			h.mu.RLock()
			if roomClients, ok := h.roomTracks[msg.RoomCode]; ok {
				for _, c := range roomClients {
					select {
					case c.send <- msg:
					default:
						close(c.send)
						delete(roomClients, c.ID)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// ClientCount returns the number of currently connected WebSocket clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

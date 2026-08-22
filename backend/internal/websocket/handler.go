package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
)

// Handler manages the WebSocket upgrade HTTP endpoint.
type Handler struct {
	hub *Hub
}

// NewHandler creates a new WebSocket upgrade handler.
func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

// ServeWS handles websocket requests from peer clients.
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}

	roomCode := r.URL.Query().Get("room")
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "guest_" + randomID(4)
	}

	clientID := "conn_" + randomID(6)

	client := NewClient(h.hub, conn, clientID, userID, roomCode)
	h.hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}

func randomID(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

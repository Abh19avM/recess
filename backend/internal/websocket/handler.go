package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"
)

// TokenValidator is an interface for validating auth JWTs on WebSocket upgrade.
type TokenValidator interface {
	ValidateToken(tokenString string, expectedType string) (userID, username string, isGuest bool, err error)
}

// Handler manages the WebSocket upgrade HTTP endpoint.
type Handler struct {
	hub       *Hub
	validator TokenValidator
}

// NewHandler creates a new WebSocket upgrade handler with token validation.
func NewHandler(hub *Hub, validator TokenValidator) *Handler {
	return &Handler{
		hub:       hub,
		validator: validator,
	}
}

// ServeWS authenticates and upgrades HTTP connections to real-time WebSockets.
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	var userID string
	var username string
	var isGuest bool

	// 1. Check for token in query params or Authorization header
	token := r.URL.Query().Get("token")
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if token != "" && h.validator != nil {
		uid, uname, guest, err := h.validator.ValidateToken(token, "access")
		if err != nil {
			slog.Warn("websocket auth failed: invalid token", "error", err)
			http.Error(w, "Unauthorized: invalid access token", http.StatusUnauthorized)
			return
		}
		userID = uid
		username = uname
		isGuest = guest
	} else {
		// 2. Allow guest fallback if explicitly provided
		guestName := r.URL.Query().Get("guest_name")
		if guestName == "" {
			guestName = r.URL.Query().Get("username")
		}
		if guestName == "" {
			guestName = "Guest_" + randomID(3)
		}

		customUID := r.URL.Query().Get("user_id")
		if customUID != "" {
			userID = customUID
		} else {
			userID = "guest_" + randomID(4)
		}
		username = guestName
		isGuest = true
	}

	avatarPreset := r.URL.Query().Get("avatar_preset")
	if avatarPreset == "" {
		avatarPreset = "pencil_sketch_1"
	}

	// 3. Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}

	clientID := "conn_" + randomID(6)
	client := NewClient(h.hub, conn, clientID, userID, username, isGuest, avatarPreset)
	if r.URL.Query().Get("role") == "spectator" || r.URL.Query().Get("spectator") == "true" {
		client.SetSpectator(true)
	}
	h.hub.Register(client)

	// 4. Auto-join initial room if requested in URL query parameter
	initialRoom := r.URL.Query().Get("room")
	if initialRoom != "" {
		h.hub.JoinRoom(client, initialRoom)
	}

	// 5. Start read and write pump routines
	go client.WritePump()
	go client.ReadPump()
}

func randomID(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

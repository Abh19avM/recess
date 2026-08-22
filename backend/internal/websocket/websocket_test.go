package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestWebSocketHubAndUpgrade(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	handler := NewHandler(hub)
	server := httptest.NewServer(http.HandlerFunc(handler.ServeWS))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?room=RECESS-TEST&user_id=usr_test_1"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	defer conn.Close()

	// Write a message
	msg := RoomMessage{
		Event: "chat",
		Data:  "hello recess",
	}
	if err := conn.WriteJSON(msg); err != nil {
		t.Fatalf("failed to write JSON to websocket: %v", err)
	}
}

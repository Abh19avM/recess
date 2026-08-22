package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Abh19avM/recess/internal/auth"
	"github.com/Abh19avM/recess/internal/users"
	"github.com/gorilla/websocket"
)

func setupTestServer(t *testing.T) (*httptest.Server, *Hub, *auth.JWTManager) {
	jwtManager := auth.NewJWTManager("test_secret_key_1234567890123456")
	hub := NewHub()

	handler := NewHandler(hub, jwtManager)
	server := httptest.NewServer(http.HandlerFunc(handler.ServeWS))

	return server, hub, jwtManager
}

func dialWS(t *testing.T, serverURL, queryParams string) (*websocket.Conn, *http.Response) {
	wsURL := "ws" + strings.TrimPrefix(serverURL, "http")
	if queryParams != "" {
		wsURL += "?" + queryParams
	}

	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, resp
	}
	return conn, resp
}

func readEnvelopeWithTimeout(t *testing.T, conn *websocket.Conn, timeout time.Duration) *EventEnvelope {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	var env EventEnvelope
	err := conn.ReadJSON(&env)
	if err != nil {
		t.Fatalf("failed to read JSON envelope: %v", err)
	}
	return &env
}

// 1. Test WebSocket upgrade with valid JWT
func TestWebSocketUpgrade_ValidJWT(t *testing.T) {
	server, hub, jwtManager := setupTestServer(t)
	defer server.Close()

	user := &users.User{
		ID:       "usr_123",
		Username: "TestPlayer",
		IsGuest:  false,
	}
	tokens, _, err := jwtManager.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	conn, resp := dialWS(t, server.URL, "token="+tokens.AccessToken)
	if conn == nil {
		t.Fatalf("expected successful connection, got resp code: %v", resp.StatusCode)
	}
	defer conn.Close()

	if hub.ClientCount() != 1 {
		t.Errorf("expected 1 connected client in hub, got %d", hub.ClientCount())
	}
}

// 2. Test WebSocket upgrade with invalid JWT (should fail with 401)
func TestWebSocketUpgrade_InvalidJWT(t *testing.T) {
	server, _, _ := setupTestServer(t)
	defer server.Close()

	conn, resp := dialWS(t, server.URL, "token=invalid_garbage_token")
	if conn != nil {
		conn.Close()
		t.Fatalf("expected connection to fail, but it succeeded")
	}

	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401 Unauthorized, got: %v", resp)
	}
}

// 3. Test WebSocket upgrade in Guest Mode
func TestWebSocketUpgrade_Guest(t *testing.T) {
	server, hub, _ := setupTestServer(t)
	defer server.Close()

	conn, _ := dialWS(t, server.URL, "guest_name=FastRunner")
	if conn == nil {
		t.Fatalf("failed to connect as guest")
	}
	defer conn.Close()

	if hub.ClientCount() != 1 {
		t.Errorf("expected 1 connected client in hub, got %d", hub.ClientCount())
	}
}

// 4. Test room join, receiving room.state, and broadcasting player.joined across 2 clients
func TestWebSocket_RoomJoinAndBroadcast(t *testing.T) {
	server, hub, _ := setupTestServer(t)
	defer server.Close()

	roomID := "RECESS-TEST-ROOM"

	// Client 1 connects and joins room
	conn1, _ := dialWS(t, server.URL, "guest_name=Alice&room="+roomID)
	if conn1 == nil {
		t.Fatalf("failed to connect conn1")
	}
	defer conn1.Close()

	// Client 1 should receive room.state
	env1 := readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	if env1.Type != EventRoomState {
		t.Fatalf("expected EventRoomState, got %s", env1.Type)
	}
	if env1.RoomID != roomID {
		t.Errorf("expected room_id %s, got %s", roomID, env1.RoomID)
	}

	var state1 RoomStatePayload
	if err := json.Unmarshal(env1.Payload, &state1); err != nil {
		t.Fatalf("failed to unmarshal room state: %v", err)
	}
	if len(state1.Members) != 1 || state1.Members[0].Username != "Alice" {
		t.Fatalf("expected 1 member (Alice), got: %+v", state1.Members)
	}

	// Client 2 connects and joins same room
	conn2, _ := dialWS(t, server.URL, "guest_name=Bob&room="+roomID)
	if conn2 == nil {
		t.Fatalf("failed to connect conn2")
	}
	defer conn2.Close()

	// Client 2 should receive room.state with 2 members
	env2 := readEnvelopeWithTimeout(t, conn2, 2*time.Second)
	if env2.Type != EventRoomState {
		t.Fatalf("expected EventRoomState for conn2, got %s", env2.Type)
	}
	var state2 RoomStatePayload
	if err := json.Unmarshal(env2.Payload, &state2); err != nil {
		t.Fatalf("failed to unmarshal room state for conn2: %v", err)
	}
	if len(state2.Members) != 2 {
		t.Fatalf("expected 2 members for conn2 state, got %d", len(state2.Members))
	}

	// Client 1 should receive player.joined (Bob)
	envJoined := readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	if envJoined.Type != EventPlayerJoined {
		t.Fatalf("expected EventPlayerJoined on conn1, got %s", envJoined.Type)
	}
	var joinedPayload PlayerJoinedPayload
	if err := json.Unmarshal(envJoined.Payload, &joinedPayload); err != nil {
		t.Fatalf("failed to unmarshal joined payload: %v", err)
	}
	if joinedPayload.Player.Username != "Bob" {
		t.Errorf("expected joined player to be Bob, got %s", joinedPayload.Player.Username)
	}

	if hub.RoomMembersCount(roomID) != 2 {
		t.Errorf("expected 2 room members, got %d", hub.RoomMembersCount(roomID))
	}
}

// 5. Test room isolation: messages in Room A do NOT reach clients in Room B
func TestWebSocket_RoomIsolation(t *testing.T) {
	server, _, _ := setupTestServer(t)
	defer server.Close()

	connA, _ := dialWS(t, server.URL, "guest_name=Alice&room=ROOM-ALPHA")
	defer connA.Close()
	_ = readEnvelopeWithTimeout(t, connA, 2*time.Second) // consume room.state

	connB, _ := dialWS(t, server.URL, "guest_name=Charlie&room=ROOM-BETA")
	defer connB.Close()
	_ = readEnvelopeWithTimeout(t, connB, 2*time.Second) // consume room.state

	// Alice sends message in ROOM-ALPHA
	msgEnv := MustEnvelope(EventMessage, "ROOM-ALPHA", 0, MessagePayload{
		Text: "Hello Alpha!",
	})
	if err := connA.WriteJSON(msgEnv); err != nil {
		t.Fatalf("failed to write msg to connA: %v", err)
	}

	// Alice receives her own room broadcast
	recvAlpha := readEnvelopeWithTimeout(t, connA, 2*time.Second)
	if recvAlpha.Type != EventMessage {
		t.Fatalf("expected EventMessage on connA, got %s", recvAlpha.Type)
	}

	// Charlie in ROOM-BETA should not receive anything
	_ = connB.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	var dummy EventEnvelope
	err := connB.ReadJSON(&dummy)
	if err == nil {
		t.Fatalf("expected timeout on connB, but received event: %+v", dummy)
	}
}

// 6. Test player ready toggle event
func TestWebSocket_PlayerReadyToggle(t *testing.T) {
	server, _, _ := setupTestServer(t)
	defer server.Close()

	roomID := "RECESS-READY-ROOM"
	conn1, _ := dialWS(t, server.URL, "guest_name=Player1&room="+roomID)
	defer conn1.Close()
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)

	conn2, _ := dialWS(t, server.URL, "guest_name=Player2&room="+roomID)
	defer conn2.Close()
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second) // player.joined on conn1

	// Player 1 sends player.ready true
	readyEnv := MustEnvelope(EventPlayerReady, roomID, 0, PlayerReadyPayload{
		IsReady: true,
	})
	if err := conn1.WriteJSON(readyEnv); err != nil {
		t.Fatalf("failed to write ready event: %v", err)
	}

	// Both conn1 and conn2 should receive the player.ready broadcast
	r1 := readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	if r1.Type != EventPlayerReady {
		t.Fatalf("expected EventPlayerReady on conn1, got %s", r1.Type)
	}

	r2 := readEnvelopeWithTimeout(t, conn2, 2*time.Second)
	if r2.Type != EventPlayerReady {
		t.Fatalf("expected EventPlayerReady on conn2, got %s", r2.Type)
	}
	var readyPayload PlayerReadyPayload
	_ = json.Unmarshal(r2.Payload, &readyPayload)
	if !readyPayload.IsReady || readyPayload.Username != "Player1" {
		t.Errorf("unexpected ready payload: %+v", readyPayload)
	}
}

// 7. Test room leave and disconnect broadcast
func TestWebSocket_RoomLeaveAndDisconnect(t *testing.T) {
	server, _, _ := setupTestServer(t)
	defer server.Close()

	roomID := "RECESS-LEAVE-ROOM"
	conn1, _ := dialWS(t, server.URL, "guest_name=User1&room="+roomID)
	defer conn1.Close()
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)

	conn2, _ := dialWS(t, server.URL, "guest_name=User2&room="+roomID)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)

	// User2 disconnects abruptly (closes connection)
	_ = conn2.Close()

	// User1 should receive player.disconnected
	discEnv := readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	if discEnv.Type != EventPlayerDisconnected {
		t.Fatalf("expected EventPlayerDisconnected, got %s", discEnv.Type)
	}
}

// 8. Test invalid message handling (error envelope returned)
func TestWebSocket_InvalidEnvelope(t *testing.T) {
	server, _, _ := setupTestServer(t)
	defer server.Close()

	conn, _ := dialWS(t, server.URL, "guest_name=Tester")
	defer conn.Close()

	// Send message without type or room_id
	badEnv := &EventEnvelope{
		Type: "",
	}
	if err := conn.WriteJSON(badEnv); err != nil {
		t.Fatalf("failed to write bad envelope: %v", err)
	}

	errEnv := readEnvelopeWithTimeout(t, conn, 2*time.Second)
	if errEnv.Type != EventError {
		t.Fatalf("expected EventError, got %s", errEnv.Type)
	}
}

package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Abh19avM/recess/internal/auth"
	"github.com/Abh19avM/recess/internal/games/dotsboxes"
	"github.com/Abh19avM/recess/internal/games/engine"
	"github.com/Abh19avM/recess/internal/users"
	"github.com/alicebob/miniredis/v2"
	"github.com/gorilla/websocket"
	goredis "github.com/redis/go-redis/v9"
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

	var count int
	for i := 0; i < 20; i++ {
		count = hub.ClientCount()
		if count == 1 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if count != 1 {
		t.Errorf("expected 1 connected client in hub, got %d", count)
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

	var count int
	for i := 0; i < 20; i++ {
		count = hub.ClientCount()
		if count == 1 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if count != 1 {
		t.Errorf("expected 1 connected client in hub, got %d", count)
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

// 9. Test Complete End-to-End XO Match over WebSockets
func TestWebSocket_FullXOMatch(t *testing.T) {
	server, _, _ := setupTestServer(t)
	defer server.Close()

	roomID := "RECESS-XO-MATCH"
	conn1, _ := dialWS(t, server.URL, "user_id=usr_alice&username=Alice&room="+roomID)
	defer conn1.Close()
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second) // consume room.state

	conn2, _ := dialWS(t, server.URL, "user_id=usr_bob&username=Bob&room="+roomID)
	defer conn2.Close()
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second) // consume room.state
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second) // consume player.joined

	// Player 1 sends ready
	_ = conn1.WriteJSON(MustEnvelope(EventPlayerReady, roomID, 0, PlayerReadyPayload{IsReady: true}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second) // player.ready on conn1
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second) // player.ready on conn2

	// Player 2 sends ready -> Auto-triggers game.start!
	_ = conn2.WriteJSON(MustEnvelope(EventPlayerReady, roomID, 0, PlayerReadyPayload{IsReady: true}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second) // player.ready on conn1
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second) // player.ready on conn2

	// Both receive game.state (initial active board)
	startEnv1 := readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	if startEnv1.Type != EventGameState {
		t.Fatalf("expected EventGameState on conn1, got %s", startEnv1.Type)
	}
	startEnv2 := readEnvelopeWithTimeout(t, conn2, 2*time.Second)
	if startEnv2.Type != EventGameState {
		t.Fatalf("expected EventGameState on conn2, got %s", startEnv2.Type)
	}

	// Move sequence: Alice marks (0,0), Bob marks (1,0), Alice (0,1), Bob (1,1), Alice (0,2) -> Alice WINS!
	moves := []struct {
		conn *websocket.Conn
		row  int
		col  int
	}{
		{conn1, 0, 0},
		{conn2, 1, 0},
		{conn1, 0, 1},
		{conn2, 1, 1},
		{conn1, 0, 2},
	}

	for idx, m := range moves {
		moveEnv := MustEnvelope(EventGameMove, roomID, 0, map[string]int{"row": m.row, "col": m.col})
		if err := m.conn.WriteJSON(moveEnv); err != nil {
			t.Fatalf("failed to write move %d: %v", idx, err)
		}

		s1 := readEnvelopeWithTimeout(t, conn1, 2*time.Second)
		s2 := readEnvelopeWithTimeout(t, conn2, 2*time.Second)
		if s1.Type != EventGameState || s2.Type != EventGameState {
			t.Fatalf("expected EventGameState after move %d, got %s / %s", idx, s1.Type, s2.Type)
		}

		if idx == len(moves)-1 {
			// Final move check: Game should be finished
			var state engine.GameState
			_ = json.Unmarshal(s1.Payload, &state)
			if state.Status != engine.StatusFinished {
				t.Fatalf("expected state status finished, got %s", state.Status)
			}
			if state.Result == nil || state.Result.WinnerID != "usr_alice" {
				t.Fatalf("expected Alice to win, got result: %+v", state.Result)
			}
		}
	}

	// Test Rematch
	rematchEnv := MustEnvelope(EventGameRematch, roomID, 0, nil)
	_ = conn1.WriteJSON(rematchEnv)

	rematch1 := readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	rematch2 := readEnvelopeWithTimeout(t, conn2, 2*time.Second)
	if rematch1.Type != EventGameState || rematch2.Type != EventGameState {
		t.Fatalf("expected EventGameState on rematch, got %s / %s", rematch1.Type, rematch2.Type)
	}

	var resetState engine.GameState
	_ = json.Unmarshal(rematch1.Payload, &resetState)
	if resetState.Status != engine.StatusActive || resetState.MoveCount != 0 {
		t.Fatalf("expected active reset game state on rematch, got %+v", resetState)
	}
}

// 10. Test Complete End-to-End Hand Cricket Match over WebSockets
func TestWebSocket_FullHandCricketMatch(t *testing.T) {
	server, _, _ := setupTestServer(t)
	defer server.Close()

	roomID := "RECESS-HC-DESK"
	conn1, _ := dialWS(t, server.URL, "user_id=usr_alice&username=Alice&room="+roomID)
	defer conn1.Close()
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second) // room.state

	conn2, _ := dialWS(t, server.URL, "user_id=usr_bob&username=Bob&room="+roomID)
	defer conn2.Close()
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second) // room.state
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second) // player.joined

	// Both ready up -> triggers hand_cricket start
	_ = conn1.WriteJSON(MustEnvelope(EventPlayerReady, roomID, 0, PlayerReadyPayload{IsReady: true}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	_ = conn2.WriteJSON(MustEnvelope(EventPlayerReady, roomID, 0, PlayerReadyPayload{IsReady: true}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// Consume initial game.state
	s1 := readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	s2 := readEnvelopeWithTimeout(t, conn2, 2*time.Second)
	if s1.Type != EventGameState || s2.Type != EventGameState {
		t.Fatalf("expected EventGameState on start")
	}

	// 1. Toss Call: Alice calls "odd"
	_ = conn1.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]string{"action": "toss_call", "call": "odd"}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// 2. Toss Throw: Alice=3, Bob=2 (Sum=5 Odd -> Alice wins toss)
	_ = conn1.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]any{"action": "toss_throw", "number": 3}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	_ = conn2.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]any{"action": "toss_throw", "number": 2}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// 3. Toss Decision: Alice chooses "bat"
	_ = conn1.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]string{"action": "toss_decision", "decision": "bat"}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// 4. Innings 1:
	// Ball 1: Alice=6, Bob=4 -> 6 runs
	_ = conn1.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"value": 6}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	_ = conn2.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"value": 4}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// Ball 2: Alice=4, Bob=4 -> OUT! (Wicket falls, Target=7)
	_ = conn1.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"value": 4}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	_ = conn2.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"value": 4}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// 5. Innings 2: Bob chases Target 7
	// Ball 1: Bob=6, Alice=1 -> 6 runs (Total: 6)
	_ = conn2.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"value": 6}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	_ = conn1.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"value": 1}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// Ball 2: Bob=4, Alice=2 -> 4 runs (Total: 10 >= Target 7 -> Bob WINS!)
	_ = conn2.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"value": 4}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	_ = conn1.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"value": 2}))
	finalEnv1 := readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	finalEnv2 := readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	if finalEnv1.Type != EventGameState || finalEnv2.Type != EventGameState {
		t.Fatalf("expected EventGameState on final ball, got %s", finalEnv1.Type)
	}

	var finalState engine.GameState
	_ = json.Unmarshal(finalEnv1.Payload, &finalState)
	if finalState.Status != engine.StatusFinished {
		t.Fatalf("expected finished status, got %s", finalState.Status)
	}
	if finalState.Result == nil || finalState.Result.WinnerID != "usr_bob" {
		t.Fatalf("expected Bob to win chase, got result: %+v", finalState.Result)
	}
}

// 11. Test Complete End-to-End Dots & Boxes Match over WebSockets
func TestWebSocket_FullDotsBoxesMatch(t *testing.T) {
	server, _, _ := setupTestServer(t)
	defer server.Close()

	roomID := "RECESS-DOTS-DESK"
	conn1, _ := dialWS(t, server.URL, "user_id=usr_alice&username=Alice&room="+roomID)
	defer conn1.Close()
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second) // room.state

	conn2, _ := dialWS(t, server.URL, "user_id=usr_bob&username=Bob&room="+roomID)
	defer conn2.Close()
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second) // room.state
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second) // player.joined

	// Both ready up -> triggers dots_boxes start
	_ = conn1.WriteJSON(MustEnvelope(EventPlayerReady, roomID, 0, PlayerReadyPayload{IsReady: true}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	_ = conn2.WriteJSON(MustEnvelope(EventPlayerReady, roomID, 0, PlayerReadyPayload{IsReady: true}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// Initial game.state
	s1 := readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	s2 := readEnvelopeWithTimeout(t, conn2, 2*time.Second)
	if s1.Type != EventGameState || s2.Type != EventGameState {
		t.Fatalf("expected EventGameState on start")
	}

	// 1. Alice draws (h, 0, 0)
	_ = conn1.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]any{"type": "h", "row": 0, "col": 0}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// 2. Bob draws (v, 0, 0)
	_ = conn2.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]any{"type": "v", "row": 0, "col": 0}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// 3. Alice draws (v, 0, 1)
	_ = conn1.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]any{"type": "v", "row": 0, "col": 1}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// 4. Bob draws (h, 1, 0) -> Completes Box (0,0)!
	_ = conn2.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]any{"type": "h", "row": 1, "col": 0}))
	res1 := readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	res2 := readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	if res1.Type != EventGameState || res2.Type != EventGameState {
		t.Fatalf("expected EventGameState after box completion")
	}

	var state engine.GameState
	_ = json.Unmarshal(res1.Payload, &state)

	var board dotsboxes.BoardState
	_ = json.Unmarshal(state.BoardState, &board)
	if board.Scores["usr_bob"] != 1 {
		t.Fatalf("expected Bob score to be 1, got %d", board.Scores["usr_bob"])
	}
	if state.CurrentTurn != "usr_bob" {
		t.Fatalf("expected Bob to get bonus turn, got %s", state.CurrentTurn)
	}
}

// 12. Test Complete End-to-End Connect 4 Match over WebSockets
func TestWebSocket_FullConnect4Match(t *testing.T) {
	server, _, _ := setupTestServer(t)
	defer server.Close()

	roomID := "RECESS-C4-TEST"
	conn1, _ := dialWS(t, server.URL, "user_id=usr_alice&username=Alice&room="+roomID)
	defer conn1.Close()
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second) // room.state

	conn2, _ := dialWS(t, server.URL, "user_id=usr_bob&username=Bob&room="+roomID)
	defer conn2.Close()
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second) // room.state
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second) // player.joined

	// Ready up -> triggers Connect 4 start
	_ = conn1.WriteJSON(MustEnvelope(EventPlayerReady, roomID, 0, PlayerReadyPayload{IsReady: true}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	_ = conn2.WriteJSON(MustEnvelope(EventPlayerReady, roomID, 0, PlayerReadyPayload{IsReady: true}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// Consume initial game.state
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// 1. Alice drops Col 0
	_ = conn1.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"col": 0}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// 2. Bob drops Col 1
	_ = conn2.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"col": 1}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// 3. Alice drops Col 0
	_ = conn1.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"col": 0}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// 4. Bob drops Col 1
	_ = conn2.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"col": 1}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// 5. Alice drops Col 0
	_ = conn1.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"col": 0}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// 6. Bob drops Col 1
	_ = conn2.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"col": 1}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// 7. Alice drops Col 0 (4th disc in Col 0 -> Vertical 4-in-a-row WIN!)
	_ = conn1.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"col": 0}))
	env1 := readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	env2 := readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	if env1.Type != EventGameState || env2.Type != EventGameState {
		t.Fatalf("expected EventGameState on winning drop")
	}

	var finalState engine.GameState
	_ = json.Unmarshal(env1.Payload, &finalState)
	if finalState.Status != engine.StatusFinished {
		t.Fatalf("expected game finished, got %s", finalState.Status)
	}
	if finalState.Result == nil || finalState.Result.WinnerID != "usr_alice" {
		t.Fatalf("expected Alice victory, got %+v", finalState.Result)
	}
}

// 13. Test Complete End-to-End Paper Football Match over WebSockets
func TestWebSocket_FullPaperFootballMatch(t *testing.T) {
	server, _, _ := setupTestServer(t)
	defer server.Close()

	roomID := "RECESS-PF-TEST"
	conn1, _ := dialWS(t, server.URL, "user_id=usr_alice&username=Alice&room="+roomID)
	defer conn1.Close()
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second) // room.state

	conn2, _ := dialWS(t, server.URL, "user_id=usr_bob&username=Bob&room="+roomID)
	defer conn2.Close()
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second) // room.state
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second) // player.joined

	// Ready up -> triggers Paper Football start
	_ = conn1.WriteJSON(MustEnvelope(EventPlayerReady, roomID, 0, PlayerReadyPayload{IsReady: true}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	_ = conn2.WriteJSON(MustEnvelope(EventPlayerReady, roomID, 0, PlayerReadyPayload{IsReady: true}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// Consume initial game.state
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// Alice flicks with 100 power -> Lands in 90-100% zone -> TOUCHDOWN! (+6 Pts)
	_ = conn1.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]any{"action": "flick", "power": 100, "angle": 0}))
	env1 := readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	env2 := readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	if env1.Type != EventGameState || env2.Type != EventGameState {
		t.Fatalf("expected EventGameState after touchdown flick")
	}

	// Alice kicks Extra Point (+1 Pt)
	_ = conn1.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]any{"action": "extra_point", "power": 50, "angle": 0}))
	env3 := readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	var midState engine.GameState
	_ = json.Unmarshal(env3.Payload, &midState)

	var board struct {
		Scores map[string]int `json:"scores"`
	}
	_ = json.Unmarshal(midState.BoardState, &board)
	if board.Scores["usr_alice"] != 7 {
		t.Fatalf("expected 7 points after TD + extra point, got %d", board.Scores["usr_alice"])
	}
}

// 14. Test Complete End-to-End Name-Place-Animal-Thing Match over WebSockets
func TestWebSocket_FullNPATMatch(t *testing.T) {
	server, _, _ := setupTestServer(t)
	defer server.Close()

	roomID := "RECESS-NPAT-TEST"
	conn1, _ := dialWS(t, server.URL, "user_id=usr_alice&username=Alice&room="+roomID)
	defer conn1.Close()
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second) // room.state

	conn2, _ := dialWS(t, server.URL, "user_id=usr_bob&username=Bob&room="+roomID)
	defer conn2.Close()
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second) // room.state
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second) // player.joined

	// Ready up -> triggers NPAT start
	_ = conn1.WriteJSON(MustEnvelope(EventPlayerReady, roomID, 0, PlayerReadyPayload{IsReady: true}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	_ = conn2.WriteJSON(MustEnvelope(EventPlayerReady, roomID, 0, PlayerReadyPayload{IsReady: true}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// Consume initial game.state
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// Alice submits words for Round 1 (S) and calls STOP!
	_ = conn1.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]any{
		"action": "call_stop",
		"name":   "Sam",
		"place":  "Spain",
		"animal": "Snake",
		"thing":  "Spoon",
	}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// Bob submits words for Round 1 (S) -> Triggers round evaluation!
	_ = conn2.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]any{
		"action": "submit_entries",
		"name":   "Sarah",
		"place":  "Seattle",
		"animal": "Spider",
		"thing":  "Scissors",
	}))
	env1 := readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	env2 := readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	if env1.Type != EventGameState || env2.Type != EventGameState {
		t.Fatalf("expected EventGameState after round evaluation")
	}

	var midState engine.GameState
	_ = json.Unmarshal(env1.Payload, &midState)

	var board struct {
		CurrentRound     int            `json:"current_round"`
		CurrentLetter    string         `json:"current_letter"`
		CumulativeScores map[string]int `json:"cumulative_scores"`
	}
	_ = json.Unmarshal(midState.BoardState, &board)

	// Both got 4 unique words = 40 points each
	if board.CumulativeScores["usr_alice"] != 40 || board.CumulativeScores["usr_bob"] != 40 {
		t.Fatalf("expected 40 points each, got %+v", board.CumulativeScores)
	}
	if board.CurrentRound != 2 || board.CurrentLetter != "A" {
		t.Fatalf("expected Round 2 with letter A, got round %d, letter %s", board.CurrentRound, board.CurrentLetter)
	}
}

// 15. Test Connection Recovery, State Snapshot, Missed Events, and Grace Period
func TestWebSocket_Reconnect_GracePeriodAndStateRecovery(t *testing.T) {
	server, _, _ := setupTestServer(t)
	defer server.Close()

	roomID := "RECESS-TTT-RECONNECT"
	conn1, _ := dialWS(t, server.URL, "user_id=usr_alice&username=Alice&room="+roomID)
	defer conn1.Close()
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second) // room.state

	conn2, _ := dialWS(t, server.URL, "user_id=usr_bob&username=Bob&room="+roomID)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second) // room.state
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second) // player.joined

	// Ready up -> match starts
	_ = conn1.WriteJSON(MustEnvelope(EventPlayerReady, roomID, 0, PlayerReadyPayload{IsReady: true}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	_ = conn2.WriteJSON(MustEnvelope(EventPlayerReady, roomID, 0, PlayerReadyPayload{IsReady: true}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// Consume initial game.state
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// Alice plays (0,0)
	_ = conn1.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"row": 0, "col": 0}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	lastBobEnv := readEnvelopeWithTimeout(t, conn2, 2*time.Second)
	lastBobSeq := lastBobEnv.Sequence

	// SIMULATE NETWORK LOSS: Bob drops connection!
	conn2.Close()

	// Alice should receive player.reconnecting notice with 30s grace period
	reconnectingEnv := readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	if reconnectingEnv.Type != EventPlayerReconnecting {
		t.Fatalf("expected EventPlayerReconnecting, got %s", reconnectingEnv.Type)
	}

	// While Bob is disconnected, game state remains intact in memory.
	// Bob re-opens connection with new socket
	connBob2, _ := dialWS(t, server.URL, "user_id=usr_bob&username=Bob&room="+roomID)
	defer connBob2.Close()
	_ = readEnvelopeWithTimeout(t, connBob2, 2*time.Second) // room.state
	_ = readEnvelopeWithTimeout(t, connBob2, 2*time.Second) // active game.state on join
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)    // player.joined on Alice

	// Bob issues session.reconnect with his last seen sequence number
	_ = connBob2.WriteJSON(MustEnvelope(EventSessionReconnect, roomID, 0, SessionReconnectPayload{
		SessionID:    "sess_usr_bob",
		RoomID:       roomID,
		LastSequence: lastBobSeq,
	}))

	// Bob receives session.reconnected with authoritative game snapshot
	reconnectedEnv := readEnvelopeWithTimeout(t, connBob2, 2*time.Second)
	if reconnectedEnv.Type != EventSessionReconnected {
		t.Fatalf("expected EventSessionReconnected, got %s", reconnectedEnv.Type)
	}

	var reconnPayload SessionReconnectedPayload
	_ = json.Unmarshal(reconnectedEnv.Payload, &reconnPayload)
	if reconnPayload.RoomID != roomID {
		t.Fatalf("expected room %s, got %s", roomID, reconnPayload.RoomID)
	}

	// Alice receives player.reconnected
	playerReconnEnv := readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	if playerReconnEnv.Type != EventPlayerReconnected {
		t.Fatalf("expected EventPlayerReconnected on Alice's client, got %s", playerReconnEnv.Type)
	}

	// Bob can now immediately continue playing his move (1,1) without issue!
	_ = connBob2.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"row": 1, "col": 1}))
	bobMoveAlice := readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	bobMoveBob := readEnvelopeWithTimeout(t, connBob2, 2*time.Second)

	if bobMoveAlice.Type != EventGameState || bobMoveBob.Type != EventGameState {
		t.Fatalf("expected EventGameState after reconnected player move")
	}
}

// 16. Test Reconnect to Finished Game Delivers Final State
func TestWebSocket_Reconnect_FinishedGame(t *testing.T) {
	server, _, _ := setupTestServer(t)
	defer server.Close()

	roomID := "RECESS-TTT-FIN-RECONN"
	conn1, _ := dialWS(t, server.URL, "user_id=usr_alice&username=Alice&room="+roomID)
	defer conn1.Close()
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)

	conn2, _ := dialWS(t, server.URL, "user_id=usr_bob&username=Bob&room="+roomID)
	defer conn2.Close()
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)

	// Start game
	_ = conn1.WriteJSON(MustEnvelope(EventPlayerReady, roomID, 0, PlayerReadyPayload{IsReady: true}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)
	_ = conn2.WriteJSON(MustEnvelope(EventPlayerReady, roomID, 0, PlayerReadyPayload{IsReady: true}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second) // initial game.state
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second) // initial game.state

	// Alice plays (0,0)
	_ = conn1.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"row": 0, "col": 0}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// Bob plays (1,0)
	_ = conn2.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"row": 1, "col": 0}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// Alice plays (0,1)
	_ = conn1.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"row": 0, "col": 1}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// Bob plays (1,1)
	_ = conn2.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"row": 1, "col": 1}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// Alice plays (0,2) -> Alice WINS (row 0 completed)!
	_ = conn1.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"row": 0, "col": 2}))
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, conn2, 2*time.Second)

	// Bob reconnects with a fresh socket to inspect final completed state
	connBobLate, _ := dialWS(t, server.URL, "user_id=usr_bob&username=Bob&room="+roomID)
	defer connBobLate.Close()
	_ = readEnvelopeWithTimeout(t, connBobLate, 2*time.Second) // room.state
	_ = readEnvelopeWithTimeout(t, connBobLate, 2*time.Second) // finished game.state on join
	_ = readEnvelopeWithTimeout(t, conn1, 2*time.Second)       // player.joined

	_ = connBobLate.WriteJSON(MustEnvelope(EventSessionReconnect, roomID, 0, SessionReconnectPayload{
		SessionID:    "sess_usr_bob",
		RoomID:       roomID,
		LastSequence: 0,
	}))

	reconnEnv := readEnvelopeWithTimeout(t, connBobLate, 2*time.Second)
	if reconnEnv.Type != EventSessionReconnected {
		t.Fatalf("expected EventSessionReconnected for finished game, got %s", reconnEnv.Type)
	}

	var reconnPayload SessionReconnectedPayload
	_ = json.Unmarshal(reconnEnv.Payload, &reconnPayload)
	if reconnPayload.GameState == nil {
		t.Fatalf("expected non-nil game state on finished match reconnection")
	}
}

// 17. Test Horizontally Scalable Distributed Match across Multiple Backend Instances
func TestWebSocket_MultiInstance_DistributedMatch(t *testing.T) {
	// Setup embedded Redis cluster/broker
	mr := miniredis.RunT(t)
	defer mr.Close()

	rdb := goredis.NewClient(&goredis.Options{
		Addr: mr.Addr(),
	})
	defer rdb.Close()

	jwtManager := auth.NewJWTManager("test_secret_key_1234567890123456")

	// Instance A
	hubA := NewHub(WithRedis(rdb), WithInstanceID("game-server-A"))
	handlerA := NewHandler(hubA, jwtManager)
	serverA := httptest.NewServer(http.HandlerFunc(handlerA.ServeWS))
	defer serverA.Close()

	// Instance B
	hubB := NewHub(WithRedis(rdb), WithInstanceID("game-server-B"))
	handlerB := NewHandler(hubB, jwtManager)
	serverB := httptest.NewServer(http.HandlerFunc(handlerB.ServeWS))
	defer serverB.Close()

	roomID := "RECESS-DISTRIB-XO"

	// Player A connects to Server A
	connA, _ := dialWS(t, serverA.URL, "user_id=usr_alice&username=Alice&room="+roomID)
	defer connA.Close()
	_ = readEnvelopeWithTimeout(t, connA, 2*time.Second) // room.state on Server A

	// Give Redis subscriptions a moment to wire
	time.Sleep(50 * time.Millisecond)

	// Player B connects to Server B
	connB, _ := dialWS(t, serverB.URL, "user_id=usr_bob&username=Bob&room="+roomID)
	defer connB.Close()
	_ = readEnvelopeWithTimeout(t, connB, 2*time.Second) // room.state on Server B

	// Player A on Server A receives player.joined broadcast from Server B via Redis Pub/Sub!
	joinedOnA := readEnvelopeWithTimeout(t, connA, 2*time.Second)
	if joinedOnA.Type != EventPlayerJoined {
		t.Fatalf("expected player.joined on Server A for Player B, got %s", joinedOnA.Type)
	}

	var joinedPayload PlayerJoinedPayload
	_ = json.Unmarshal(joinedOnA.Payload, &joinedPayload)
	if joinedPayload.Player.Username != "Bob" {
		t.Fatalf("expected Bob to join, got %s", joinedPayload.Player.Username)
	}

	// Player A on Server A readies up
	_ = connA.WriteJSON(MustEnvelope(EventPlayerReady, roomID, 0, PlayerReadyPayload{IsReady: true}))
	_ = readEnvelopeWithTimeout(t, connA, 2*time.Second) // local ready on Server A
	readyOnB := readEnvelopeWithTimeout(t, connB, 2*time.Second) // remote ready delivered to Server B via PubSub
	if readyOnB.Type != EventPlayerReady {
		t.Fatalf("expected player.ready on Server B for Alice, got %s", readyOnB.Type)
	}

	// Player B on Server B sends room message to Player A on Server A
	_ = connB.WriteJSON(MustEnvelope(EventMessage, roomID, 0, MessagePayload{Text: "Good luck from Server B!"}))
	_ = readEnvelopeWithTimeout(t, connB, 2*time.Second) // local broadcast on Server B
	msgOnA := readEnvelopeWithTimeout(t, connA, 2*time.Second) // remote message delivered to Server A
	if msgOnA.Type != EventMessage {
		t.Fatalf("expected EventMessage on Server A from Server B, got %s", msgOnA.Type)
	}

	var msgPayload MessagePayload
	_ = json.Unmarshal(msgOnA.Payload, &msgPayload)
	if msgPayload.Sender != "Bob" || msgPayload.Text != "Good luck from Server B!" {
		t.Fatalf("unexpected message payload on Server A: %+v", msgPayload)
	}
}

// 18. Test Spectator Mode: Live Event/State Reception, Move Rejection, and Count Sync
func TestWebSocket_SpectatorMode_LiveSyncAndMoveRejection(t *testing.T) {
	server, _, _ := setupTestServer(t)
	defer server.Close()

	roomID := "RECESS-SPECTATE-XO"

	// Player 1 (Alice) connects as player
	connAlice, _ := dialWS(t, server.URL, "user_id=usr_alice&username=Alice&room="+roomID)
	defer connAlice.Close()
	_ = readEnvelopeWithTimeout(t, connAlice, 2*time.Second) // room.state

	// Player 2 (Bob) connects as player
	connBob, _ := dialWS(t, server.URL, "user_id=usr_bob&username=Bob&room="+roomID)
	defer connBob.Close()
	_ = readEnvelopeWithTimeout(t, connBob, 2*time.Second) // room.state
	_ = readEnvelopeWithTimeout(t, connAlice, 2*time.Second) // player.joined

	// Both ready up to start game
	_ = connAlice.WriteJSON(MustEnvelope(EventPlayerReady, roomID, 0, PlayerReadyPayload{IsReady: true}))
	_ = readEnvelopeWithTimeout(t, connAlice, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, connBob, 2*time.Second)

	_ = connBob.WriteJSON(MustEnvelope(EventPlayerReady, roomID, 0, PlayerReadyPayload{IsReady: true}))
	_ = readEnvelopeWithTimeout(t, connAlice, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, connBob, 2*time.Second)

	// Consume initial game.state
	_ = readEnvelopeWithTimeout(t, connAlice, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, connBob, 2*time.Second)

	// Spectator (Charlie) joins the sideline
	connCharlie, _ := dialWS(t, server.URL, "role=spectator&user_id=usr_charlie&username=Charlie&room="+roomID)
	defer connCharlie.Close()

	// Charlie receives room.state (with spectator count = 1)
	charlieRoomState := readEnvelopeWithTimeout(t, connCharlie, 2*time.Second)
	if charlieRoomState.Type != EventRoomState {
		t.Fatalf("expected room.state on spectator join, got %s", charlieRoomState.Type)
	}

	var statePayload RoomStatePayload
	_ = json.Unmarshal(charlieRoomState.Payload, &statePayload)
	if statePayload.SpectatorCount != 1 {
		t.Fatalf("expected SpectatorCount=1, got %d", statePayload.SpectatorCount)
	}

	// Charlie receives active game.state snapshot
	charlieGameState := readEnvelopeWithTimeout(t, connCharlie, 2*time.Second)
	if charlieGameState.Type != EventGameState {
		t.Fatalf("expected game.state on spectator join, got %s", charlieGameState.Type)
	}

	// Alice and Bob receive spectator.joined notice
	spectJoinedAlice := readEnvelopeWithTimeout(t, connAlice, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, connBob, 2*time.Second)
	if spectJoinedAlice.Type != EventSpectatorJoined {
		t.Fatalf("expected EventSpectatorJoined on player client, got %s", spectJoinedAlice.Type)
	}

	// Charlie (spectator) attempts to submit a move -> MUST BE REJECTED!
	_ = connCharlie.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"row": 0, "col": 0}))
	errEnv := readEnvelopeWithTimeout(t, connCharlie, 2*time.Second)
	if errEnv.Type != EventError {
		t.Fatalf("expected EventError when spectator submits move, got %s", errEnv.Type)
	}

	var errPayload ErrorPayload
	_ = json.Unmarshal(errEnv.Payload, &errPayload)
	if errPayload.Code != "SPECTATOR_CANNOT_MOVE" {
		t.Fatalf("expected SPECTATOR_CANNOT_MOVE error code, got %s", errPayload.Code)
	}

	// Alice (active player) makes a valid move -> Charlie receives live synchronized game.state!
	_ = connAlice.WriteJSON(MustEnvelope(EventGameMove, roomID, 0, map[string]int{"row": 0, "col": 0}))
	_ = readEnvelopeWithTimeout(t, connAlice, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, connBob, 2*time.Second)

	charlieLiveMove := readEnvelopeWithTimeout(t, connCharlie, 2*time.Second)
	if charlieLiveMove.Type != EventGameState {
		t.Fatalf("expected spectator Charlie to receive live game.state move update, got %s", charlieLiveMove.Type)
	}

	// Charlie disconnects -> Alice and Bob receive spectator.left notice with count = 0
	connCharlie.Close()
	spectLeftAlice := readEnvelopeWithTimeout(t, connAlice, 2*time.Second)
	_ = readEnvelopeWithTimeout(t, connBob, 2*time.Second)
	if spectLeftAlice.Type != EventSpectatorLeft {
		t.Fatalf("expected EventSpectatorLeft when spectator disconnects, got %s", spectLeftAlice.Type)
	}
}










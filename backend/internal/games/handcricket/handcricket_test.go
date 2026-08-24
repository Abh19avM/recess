package handcricket_test

import (
	"encoding/json"
	"testing"

	"github.com/Abh19avM/recess/internal/games/engine"
	"github.com/Abh19avM/recess/internal/games/handcricket"
)

func makePlayers() []engine.Player {
	return []engine.Player{
		{ID: "usr_alice", Username: "Alice", SeatNumber: 1},
		{ID: "usr_bob", Username: "Bob", SeatNumber: 2},
	}
}

func makeMove(playerID, action string, payload any) engine.Move {
	data, _ := json.Marshal(payload)
	return engine.Move{
		PlayerID:  playerID,
		Action:    action,
		Data:      data,
		Timestamp: 1000,
	}
}

// 1. Test Toss Flow and Decision
func TestHandCricket_TossFlow(t *testing.T) {
	eng := handcricket.NewHandCricketEngine()
	players := makePlayers()
	state, err := eng.Initialize("hc_toss", players, nil)
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}

	var board handcricket.BoardState
	_ = json.Unmarshal(state.BoardState, &board)
	if board.Phase != handcricket.PhaseTossCall {
		t.Fatalf("expected phase toss_call, got %s", board.Phase)
	}

	// Bob attempts to call toss (Alice is toss caller)
	err = eng.ValidateMove(makeMove("usr_bob", handcricket.ActionTossCall, map[string]string{"call": "odd"}))
	if err == nil {
		t.Fatalf("expected error for non-toss caller")
	}

	// Alice calls "odd"
	_, err = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionTossCall, map[string]string{"call": "odd"}))
	if err != nil {
		t.Fatalf("failed to apply toss call: %v", err)
	}

	// Both players throw numbers: Alice=3, Bob=2 (Sum=5, Odd -> Alice wins toss!)
	_, err = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionTossThrow, map[string]int{"number": 3}))
	if err != nil {
		t.Fatalf("failed Alice toss throw: %v", err)
	}

	state, err = eng.ApplyMove(makeMove("usr_bob", handcricket.ActionTossThrow, map[string]int{"number": 2}))
	if err != nil {
		t.Fatalf("failed Bob toss throw: %v", err)
	}

	_ = json.Unmarshal(state.BoardState, &board)
	if board.Phase != handcricket.PhaseTossDecision {
		t.Fatalf("expected phase toss_decision, got %s", board.Phase)
	}
	if board.TossWinnerID != "usr_alice" {
		t.Fatalf("expected Alice to win toss, got %s", board.TossWinnerID)
	}

	// Alice chooses to BAT
	state, err = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionTossDecision, map[string]string{"decision": "bat"}))
	if err != nil {
		t.Fatalf("failed toss decision: %v", err)
	}

	_ = json.Unmarshal(state.BoardState, &board)
	if board.Phase != handcricket.PhaseInnings1 {
		t.Fatalf("expected PhaseInnings1, got %s", board.Phase)
	}
	if board.BatsmanID != "usr_alice" || board.BowlerID != "usr_bob" {
		t.Fatalf("expected Alice batting, Bob bowling. Got: %+v", board)
	}
}

// 2. Test Scoring Runs in Innings 1 and Wicket Detection
func TestHandCricket_Innings1AndWicket(t *testing.T) {
	eng := handcricket.NewHandCricketEngine()
	players := makePlayers()
	_, _ = eng.Initialize("hc_game", players, nil)

	// Complete toss (Alice bats)
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionTossCall, map[string]string{"call": "even"}))
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionTossThrow, map[string]int{"number": 2}))
	_, _ = eng.ApplyMove(makeMove("usr_bob", handcricket.ActionTossThrow, map[string]int{"number": 4})) // Sum 6 (even) -> Alice wins
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionTossDecision, map[string]string{"decision": "bat"}))

	// Ball 1: Alice picks 6, Bob picks 4 -> Alice scores 6!
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionChooseNumber, map[string]int{"value": 6}))
	state, err := eng.ApplyMove(makeMove("usr_bob", handcricket.ActionChooseNumber, map[string]int{"value": 4}))
	if err != nil {
		t.Fatalf("ball 1 failed: %v", err)
	}

	var board handcricket.BoardState
	_ = json.Unmarshal(state.BoardState, &board)
	if board.Innings1.Runs != 6 {
		t.Fatalf("expected 6 runs on ball 1, got %d", board.Innings1.Runs)
	}
	if board.Innings1.Wickets != 0 {
		t.Fatalf("expected 0 wickets on ball 1")
	}

	// Ball 2: Alice picks 4, Bob picks 2 -> Alice scores 4 (Total: 10)!
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionChooseNumber, map[string]int{"value": 4}))
	state, _ = eng.ApplyMove(makeMove("usr_bob", handcricket.ActionChooseNumber, map[string]int{"value": 2}))
	_ = json.Unmarshal(state.BoardState, &board)
	if board.Innings1.Runs != 10 {
		t.Fatalf("expected 10 runs, got %d", board.Innings1.Runs)
	}

	// Ball 3: Alice picks 3, Bob picks 3 -> OUT! Wicket falls, Innings 1 ends!
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionChooseNumber, map[string]int{"value": 3}))
	state, _ = eng.ApplyMove(makeMove("usr_bob", handcricket.ActionChooseNumber, map[string]int{"value": 3}))
	_ = json.Unmarshal(state.BoardState, &board)

	if board.Phase != handcricket.PhaseInnings2 {
		t.Fatalf("expected transition to PhaseInnings2, got %s", board.Phase)
	}
	if board.Target != 11 {
		t.Fatalf("expected target 11, got %d", board.Target)
	}
	if board.BatsmanID != "usr_bob" || board.BowlerID != "usr_alice" {
		t.Fatalf("expected roles to swap for Innings 2: %+v", board)
	}
}

// 3. Test Second Innings Chase Victory
func TestHandCricket_ChaseVictory(t *testing.T) {
	eng := handcricket.NewHandCricketEngine()
	players := makePlayers()
	_, _ = eng.Initialize("hc_chase", players, nil)

	// Setup: Alice bats 1st, scores 4 on ball 1, out on ball 2 (Total: 4, Target: 5)
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionTossCall, map[string]string{"call": "odd"}))
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionTossThrow, map[string]int{"number": 1}))
	_, _ = eng.ApplyMove(makeMove("usr_bob", handcricket.ActionTossThrow, map[string]int{"number": 2}))
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionTossDecision, map[string]string{"decision": "bat"}))

	// Innings 1: Ball 1 Alice=4, Bob=1 (4 runs). Ball 2 Alice=2, Bob=2 (OUT). Target = 5
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionChooseNumber, map[string]int{"value": 4}))
	_, _ = eng.ApplyMove(makeMove("usr_bob", handcricket.ActionChooseNumber, map[string]int{"value": 1}))
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionChooseNumber, map[string]int{"value": 2}))
	_, _ = eng.ApplyMove(makeMove("usr_bob", handcricket.ActionChooseNumber, map[string]int{"value": 2}))

	// Innings 2: Bob (Chaser) hits 6 on first ball! (Score: 6 >= Target 5 -> Bob wins!)
	_, _ = eng.ApplyMove(makeMove("usr_bob", handcricket.ActionChooseNumber, map[string]int{"value": 6}))
	state, err := eng.ApplyMove(makeMove("usr_alice", handcricket.ActionChooseNumber, map[string]int{"value": 3}))
	if err != nil {
		t.Fatalf("innings 2 move failed: %v", err)
	}

	if !eng.IsFinished() {
		t.Fatalf("expected game to finish on chase completion")
	}
	if state.Status != engine.StatusFinished {
		t.Fatalf("expected StatusFinished, got %s", state.Status)
	}
	if state.Result == nil || state.Result.WinnerID != "usr_bob" {
		t.Fatalf("expected Bob to win, got result: %+v", state.Result)
	}
}

// 4. Test Second Innings Defense Victory (Defending bowler takes wicket before target)
func TestHandCricket_DefenseVictory(t *testing.T) {
	eng := handcricket.NewHandCricketEngine()
	players := makePlayers()
	_, _ = eng.Initialize("hc_defend", players, nil)

	// Setup: Alice bats 1st, scores 10, Target is 11
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionTossCall, map[string]string{"call": "odd"}))
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionTossThrow, map[string]int{"number": 1}))
	_, _ = eng.ApplyMove(makeMove("usr_bob", handcricket.ActionTossThrow, map[string]int{"number": 2}))
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionTossDecision, map[string]string{"decision": "bat"}))

	// Alice scores 6 + 4 = 10, then gets out
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionChooseNumber, map[string]int{"value": 6}))
	_, _ = eng.ApplyMove(makeMove("usr_bob", handcricket.ActionChooseNumber, map[string]int{"value": 1}))
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionChooseNumber, map[string]int{"value": 4}))
	_, _ = eng.ApplyMove(makeMove("usr_bob", handcricket.ActionChooseNumber, map[string]int{"value": 2}))
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionChooseNumber, map[string]int{"value": 5}))
	_, _ = eng.ApplyMove(makeMove("usr_bob", handcricket.ActionChooseNumber, map[string]int{"value": 5})) // OUT! Target = 11

	// Innings 2: Bob scores 2, then gets OUT on ball 2 (Total: 2 < Target 11)
	_, _ = eng.ApplyMove(makeMove("usr_bob", handcricket.ActionChooseNumber, map[string]int{"value": 2}))
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionChooseNumber, map[string]int{"value": 1}))

	_, _ = eng.ApplyMove(makeMove("usr_bob", handcricket.ActionChooseNumber, map[string]int{"value": 4}))
	state, _ := eng.ApplyMove(makeMove("usr_alice", handcricket.ActionChooseNumber, map[string]int{"value": 4})) // OUT!

	if !eng.IsFinished() {
		t.Fatalf("expected game to finish")
	}
	if state.Result == nil || state.Result.WinnerID != "usr_alice" {
		t.Fatalf("expected Alice defense victory, got result: %+v", state.Result)
	}
}

// 5. Test Tied Match (Draw)
func TestHandCricket_Tie(t *testing.T) {
	eng := handcricket.NewHandCricketEngine()
	players := makePlayers()
	_, _ = eng.Initialize("hc_tie", players, nil)

	// Setup: Alice bats 1st, scores 4, Target 5
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionTossCall, map[string]string{"call": "odd"}))
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionTossThrow, map[string]int{"number": 1}))
	_, _ = eng.ApplyMove(makeMove("usr_bob", handcricket.ActionTossThrow, map[string]int{"number": 2}))
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionTossDecision, map[string]string{"decision": "bat"}))

	// Alice scores 4, then out
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionChooseNumber, map[string]int{"value": 4}))
	_, _ = eng.ApplyMove(makeMove("usr_bob", handcricket.ActionChooseNumber, map[string]int{"value": 1}))
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionChooseNumber, map[string]int{"value": 5}))
	_, _ = eng.ApplyMove(makeMove("usr_bob", handcricket.ActionChooseNumber, map[string]int{"value": 5})) // OUT at 4

	// Bob scores 4, then gets OUT (Scores equal at 4)
	_, _ = eng.ApplyMove(makeMove("usr_bob", handcricket.ActionChooseNumber, map[string]int{"value": 4}))
	_, _ = eng.ApplyMove(makeMove("usr_alice", handcricket.ActionChooseNumber, map[string]int{"value": 1}))
	_, _ = eng.ApplyMove(makeMove("usr_bob", handcricket.ActionChooseNumber, map[string]int{"value": 6}))
	state, _ := eng.ApplyMove(makeMove("usr_alice", handcricket.ActionChooseNumber, map[string]int{"value": 6})) // OUT at 4

	if !eng.IsFinished() {
		t.Fatalf("expected tie game finished")
	}
	if state.Result == nil || !state.Result.IsDraw || state.Result.WinnerID != "" {
		t.Fatalf("expected draw result, got: %+v", state.Result)
	}
}

// 6. Test Invalid Moves and Error Handling
func TestHandCricket_InvalidMoves(t *testing.T) {
	eng := handcricket.NewHandCricketEngine()
	players := makePlayers()
	_, _ = eng.Initialize("hc_err", players, nil)

	// Invalid number (> 6 or < 1)
	err := eng.ValidateMove(makeMove("usr_alice", handcricket.ActionTossThrow, map[string]int{"number": 7}))
	if err == nil {
		t.Fatalf("expected error for number 7")
	}
	err = eng.ValidateMove(makeMove("usr_alice", handcricket.ActionTossThrow, map[string]int{"number": 0}))
	if err == nil {
		t.Fatalf("expected error for number 0")
	}

	// Unknown player
	err = eng.ValidateMove(makeMove("usr_intruder", handcricket.ActionTossCall, map[string]string{"call": "odd"}))
	if err != engine.ErrPlayerNotInGame {
		t.Fatalf("expected ErrPlayerNotInGame, got %v", err)
	}
}

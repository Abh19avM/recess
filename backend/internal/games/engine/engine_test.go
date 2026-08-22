package engine_test

import (
	"encoding/json"
	"testing"

	"github.com/Abh19avM/recess/internal/games/engine"
	"github.com/Abh19avM/recess/internal/games/xo"
)

func makeTestPlayers() []engine.Player {
	return []engine.Player{
		{ID: "usr_alice", Username: "Alice", SeatNumber: 1},
		{ID: "usr_bob", Username: "Bob", SeatNumber: 2},
	}
}

func makeMove(playerID string, row, col int) engine.Move {
	data, _ := json.Marshal(map[string]int{"row": row, "col": col})
	return engine.Move{
		PlayerID:  playerID,
		Action:    "mark",
		Data:      data,
		Timestamp: 1000,
	}
}

// 1. Test Registry Operations
func TestRegistry(t *testing.T) {
	reg := engine.NewRegistry()

	if reg.IsRegistered(engine.GameTypeXO) {
		t.Fatalf("expected XO to not be registered initially in clean registry")
	}

	// Register XO
	err := reg.Register(engine.GameTypeXO, xo.NewXOEngine)
	if err != nil {
		t.Fatalf("failed to register XO: %v", err)
	}

	if !reg.IsRegistered(engine.GameTypeXO) {
		t.Fatalf("expected XO to be registered")
	}

	// Duplicate registration should fail
	err = reg.Register(engine.GameTypeXO, xo.NewXOEngine)
	if err != engine.ErrGameAlreadyExists {
		t.Fatalf("expected ErrGameAlreadyExists, got %v", err)
	}

	// Create XO instance
	inst, err := reg.Create(engine.GameTypeXO)
	if err != nil {
		t.Fatalf("failed to create XO engine: %v", err)
	}
	if inst.GameType() != engine.GameTypeXO {
		t.Fatalf("expected game type xo, got %s", inst.GameType())
	}

	// Unknown game creation
	_, err = reg.Create("unknown_game")
	if err != engine.ErrGameNotFound {
		t.Fatalf("expected ErrGameNotFound, got %v", err)
	}
}

// 2. Test XO Engine Initialization
func TestXOEngine_Initialization(t *testing.T) {
	eng := xo.NewXOEngine()
	players := makeTestPlayers()

	// Invalid player counts
	_, err := eng.Initialize("game_1", players[:1], nil)
	if err == nil {
		t.Fatalf("expected error for 1 player, got nil")
	}

	state, err := eng.Initialize("game_1", players, nil)
	if err != nil {
		t.Fatalf("failed to initialize XO: %v", err)
	}

	if state.Status != engine.StatusActive {
		t.Fatalf("expected status active, got %s", state.Status)
	}
	if state.CurrentTurn != "usr_alice" {
		t.Fatalf("expected Alice to have first turn, got %s", state.CurrentTurn)
	}
	if state.Version != 1 {
		t.Fatalf("expected version 1, got %d", state.Version)
	}
	if state.MoveCount != 0 {
		t.Fatalf("expected move count 0, got %d", state.MoveCount)
	}
	if eng.IsFinished() {
		t.Fatalf("expected game not finished")
	}
}

// 3. Test Move Validation & Error Cases
func TestXOEngine_MoveValidation(t *testing.T) {
	eng := xo.NewXOEngine()
	players := makeTestPlayers()
	_, _ = eng.Initialize("game_1", players, nil)

	// Bob trying to move on Alice's turn
	err := eng.ValidateMove(makeMove("usr_bob", 0, 0))
	if err != engine.ErrNotPlayerTurn {
		t.Fatalf("expected ErrNotPlayerTurn, got %v", err)
	}

	// Out of bounds coordinate
	err = eng.ValidateMove(makeMove("usr_alice", 3, 0))
	if err != engine.ErrInvalidCoordinates {
		t.Fatalf("expected ErrInvalidCoordinates, got %v", err)
	}

	// Valid move
	err = eng.ValidateMove(makeMove("usr_alice", 0, 0))
	if err != nil {
		t.Fatalf("expected valid move, got %v", err)
	}

	// Apply Alice's move at (0,0)
	_, err = eng.ApplyMove(makeMove("usr_alice", 0, 0))
	if err != nil {
		t.Fatalf("failed to apply move: %v", err)
	}

	// Bob trying to mark already occupied cell (0,0)
	err = eng.ValidateMove(makeMove("usr_bob", 0, 0))
	if err != engine.ErrCellOccupied {
		t.Fatalf("expected ErrCellOccupied, got %v", err)
	}
}

// 4. Test Row Win (Alice wins top row)
func TestXOEngine_RowWin(t *testing.T) {
	eng := xo.NewXOEngine()
	players := makeTestPlayers()
	_, _ = eng.Initialize("game_1", players, nil)

	// Alice: (0,0), Bob: (1,0)
	// Alice: (0,1), Bob: (1,1)
	// Alice: (0,2) -> WIN!
	moves := []engine.Move{
		makeMove("usr_alice", 0, 0),
		makeMove("usr_bob", 1, 0),
		makeMove("usr_alice", 0, 1),
		makeMove("usr_bob", 1, 1),
		makeMove("usr_alice", 0, 2),
	}

	var state *engine.GameState
	var err error
	for _, m := range moves {
		state, err = eng.ApplyMove(m)
		if err != nil {
			t.Fatalf("unexpected move error: %v", err)
		}
	}

	if !eng.IsFinished() {
		t.Fatalf("expected game to be finished")
	}
	if state.Status != engine.StatusFinished {
		t.Fatalf("expected StatusFinished, got %s", state.Status)
	}
	if state.Result == nil || state.Result.WinnerID != "usr_alice" {
		t.Fatalf("expected Alice to be winner, got: %+v", state.Result)
	}
	if state.Result.IsDraw {
		t.Fatalf("expected not a draw")
	}
	if state.Result.Scores["usr_alice"] != 1 || state.Result.Scores["usr_bob"] != 0 {
		t.Fatalf("unexpected scores: %+v", state.Result.Scores)
	}

	// Trying to move after game finished
	_, err = eng.ApplyMove(makeMove("usr_bob", 2, 2))
	if err != engine.ErrGameFinished {
		t.Fatalf("expected ErrGameFinished, got %v", err)
	}
}

// 5. Test Diagonal Win
func TestXOEngine_DiagonalWin(t *testing.T) {
	eng := xo.NewXOEngine()
	players := makeTestPlayers()
	_, _ = eng.Initialize("game_1", players, nil)

	// Alice: (0,0), Bob: (0,1)
	// Alice: (1,1), Bob: (0,2)
	// Alice: (2,2) -> Diagonal WIN!
	moves := []engine.Move{
		makeMove("usr_alice", 0, 0),
		makeMove("usr_bob", 0, 1),
		makeMove("usr_alice", 1, 1),
		makeMove("usr_bob", 0, 2),
		makeMove("usr_alice", 2, 2),
	}

	for _, m := range moves {
		if _, err := eng.ApplyMove(m); err != nil {
			t.Fatalf("move error: %v", err)
		}
	}

	if !eng.IsFinished() || eng.Result().WinnerID != "usr_alice" {
		t.Fatalf("expected Alice diagonal win, got result: %+v", eng.Result())
	}
}

// 6. Test Draw Scenario (Cats Game)
func TestXOEngine_Draw(t *testing.T) {
	eng := xo.NewXOEngine()
	players := makeTestPlayers()
	_, _ = eng.Initialize("game_1", players, nil)

	// Board layout for draw:
	// X O X
	// X X O
	// O X O
	moves := []engine.Move{
		makeMove("usr_alice", 0, 0), // X
		makeMove("usr_bob", 0, 1),   // O
		makeMove("usr_alice", 0, 2), // X
		makeMove("usr_bob", 1, 2),   // O
		makeMove("usr_alice", 1, 0), // X
		makeMove("usr_bob", 2, 0),   // O
		makeMove("usr_alice", 1, 1), // X
		makeMove("usr_bob", 2, 2),   // O
		makeMove("usr_alice", 2, 1), // X
	}

	var state *engine.GameState
	for _, m := range moves {
		var err error
		state, err = eng.ApplyMove(m)
		if err != nil {
			t.Fatalf("move error: %v", err)
		}
	}

	if !eng.IsFinished() {
		t.Fatalf("expected game finished on draw")
	}
	if !state.Result.IsDraw {
		t.Fatalf("expected IsDraw == true, got false")
	}
	if state.Result.WinnerID != "" {
		t.Fatalf("expected empty winner ID on draw, got %s", state.Result.WinnerID)
	}
}

// 7. Test Determinism (Two engines receiving same moves produce identical state)
func TestXOEngine_Determinism(t *testing.T) {
	eng1 := xo.NewXOEngine()
	eng2 := xo.NewXOEngine()
	players := makeTestPlayers()

	_, _ = eng1.Initialize("game_det", players, nil)
	_, _ = eng2.Initialize("game_det", players, nil)

	moves := []engine.Move{
		makeMove("usr_alice", 0, 0),
		makeMove("usr_bob", 1, 1),
		makeMove("usr_alice", 0, 2),
		makeMove("usr_bob", 0, 1),
		makeMove("usr_alice", 2, 0),
		makeMove("usr_bob", 1, 0),
		makeMove("usr_alice", 1, 2),
	}

	for _, m := range moves {
		s1, err1 := eng1.ApplyMove(m)
		s2, err2 := eng2.ApplyMove(m)
		if err1 != nil || err2 != nil {
			t.Fatalf("unexpected move error: %v / %v", err1, err2)
		}
		if s1.Version != s2.Version {
			t.Fatalf("version mismatch: %d vs %d", s1.Version, s2.Version)
		}
		if s1.CurrentTurn != s2.CurrentTurn {
			t.Fatalf("current turn mismatch: %s vs %s", s1.CurrentTurn, s2.CurrentTurn)
		}
		if string(s1.BoardState) != string(s2.BoardState) {
			t.Fatalf("board state mismatch: %s vs %s", string(s1.BoardState), string(s2.BoardState))
		}
	}
}

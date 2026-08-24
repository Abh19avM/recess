package connect4_test

import (
	"encoding/json"
	"testing"

	"github.com/Abh19avM/recess/internal/games/connect4"
	"github.com/Abh19avM/recess/internal/games/engine"
)

func makePlayers() []engine.Player {
	return []engine.Player{
		{ID: "usr_alice", Username: "Alice", SeatNumber: 1},
		{ID: "usr_bob", Username: "Bob", SeatNumber: 2},
	}
}

func makeDropMove(playerID string, col int) engine.Move {
	data, _ := json.Marshal(connect4.DropMove{Col: col})
	return engine.Move{
		PlayerID:  playerID,
		Action:    "drop_disc",
		Data:      data,
		Timestamp: 1000,
	}
}

// 1. Test Board Initialization & Gravity Drop
func TestConnect4_InitializationAndGravity(t *testing.T) {
	eng := connect4.NewConnect4Engine()
	state, err := eng.Initialize("c4_init", makePlayers(), nil)
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}

	if state.CurrentTurn != "usr_alice" {
		t.Fatalf("expected Alice turn first, got %s", state.CurrentTurn)
	}

	// Alice drops disc in Col 3 -> should land at bottom Row 5
	state, err = eng.ApplyMove(makeDropMove("usr_alice", 3))
	if err != nil {
		t.Fatalf("failed drop move: %v", err)
	}

	var board connect4.BoardState
	_ = json.Unmarshal(state.BoardState, &board)
	if board.Grid[5][3] != "usr_alice" {
		t.Fatalf("expected disc at (5,3), got %s", board.Grid[5][3])
	}
	if state.CurrentTurn != "usr_bob" {
		t.Fatalf("expected Bob turn next, got %s", state.CurrentTurn)
	}

	// Bob drops disc in Col 3 -> should stack on top at Row 4
	state, err = eng.ApplyMove(makeDropMove("usr_bob", 3))
	if err != nil {
		t.Fatalf("failed drop move: %v", err)
	}
	_ = json.Unmarshal(state.BoardState, &board)
	if board.Grid[4][3] != "usr_bob" {
		t.Fatalf("expected disc at (4,3), got %s", board.Grid[4][3])
	}
}

// 2. Test Horizontal 4-in-a-row Win
func TestConnect4_HorizontalWin(t *testing.T) {
	eng := connect4.NewConnect4Engine()
	_, _ = eng.Initialize("c4_horiz", makePlayers(), nil)

	// Alice: Cols 0, 1, 2, 3 (Row 5)
	// Bob:   Cols 0, 1, 2 (Row 4)
	_, _ = eng.ApplyMove(makeDropMove("usr_alice", 0)) // A -> (5,0)
	_, _ = eng.ApplyMove(makeDropMove("usr_bob", 0))   // B -> (4,0)
	_, _ = eng.ApplyMove(makeDropMove("usr_alice", 1)) // A -> (5,1)
	_, _ = eng.ApplyMove(makeDropMove("usr_bob", 1))   // B -> (4,1)
	_, _ = eng.ApplyMove(makeDropMove("usr_alice", 2)) // A -> (5,2)
	_, _ = eng.ApplyMove(makeDropMove("usr_bob", 2))   // B -> (4,2)

	// Alice drops winning disc in Col 3 (5,3)
	state, err := eng.ApplyMove(makeDropMove("usr_alice", 3))
	if err != nil {
		t.Fatalf("failed winning move: %v", err)
	}

	if !eng.IsFinished() {
		t.Fatalf("expected game finished on horizontal 4-in-a-row")
	}
	if state.Result == nil || state.Result.WinnerID != "usr_alice" {
		t.Fatalf("expected Alice winner, got result: %+v", state.Result)
	}
}

// 3. Test Vertical 4-in-a-row Win
func TestConnect4_VerticalWin(t *testing.T) {
	eng := connect4.NewConnect4Engine()
	_, _ = eng.Initialize("c4_vert", makePlayers(), nil)

	// Alice stacks in Col 2 (Rows 5, 4, 3, 2)
	// Bob drops in Col 3 (Rows 5, 4, 3)
	_, _ = eng.ApplyMove(makeDropMove("usr_alice", 2)) // A (5,2)
	_, _ = eng.ApplyMove(makeDropMove("usr_bob", 3))   // B (5,3)
	_, _ = eng.ApplyMove(makeDropMove("usr_alice", 2)) // A (4,2)
	_, _ = eng.ApplyMove(makeDropMove("usr_bob", 3))   // B (4,3)
	_, _ = eng.ApplyMove(makeDropMove("usr_alice", 2)) // A (3,2)
	_, _ = eng.ApplyMove(makeDropMove("usr_bob", 3))   // B (3,3)

	// Alice drops 4th disc in Col 2 (2,2) -> WINS!
	state, err := eng.ApplyMove(makeDropMove("usr_alice", 2))
	if err != nil {
		t.Fatalf("failed move: %v", err)
	}

	if !eng.IsFinished() || state.Result == nil || state.Result.WinnerID != "usr_alice" {
		t.Fatalf("expected Alice vertical win, got %+v", state.Result)
	}
}

// 4. Test Diagonal Win (\ down-right)
func TestConnect4_DiagonalWin(t *testing.T) {
	eng := connect4.NewConnect4Engine()
	_, _ = eng.Initialize("c4_diag", makePlayers(), nil)

	// Build staircase so Alice gets diagonal (5,0), (4,1), (3,2), (2,3)
	_, _ = eng.ApplyMove(makeDropMove("usr_alice", 0)) // A (5,0)
	_, _ = eng.ApplyMove(makeDropMove("usr_bob", 1))   // B (5,1)
	_, _ = eng.ApplyMove(makeDropMove("usr_alice", 1)) // A (4,1)
	_, _ = eng.ApplyMove(makeDropMove("usr_bob", 2))   // B (5,2)
	_, _ = eng.ApplyMove(makeDropMove("usr_alice", 3)) // A (5,3)
	_, _ = eng.ApplyMove(makeDropMove("usr_bob", 2))   // B (4,2)
	_, _ = eng.ApplyMove(makeDropMove("usr_alice", 2)) // A (3,2)
	_, _ = eng.ApplyMove(makeDropMove("usr_bob", 3))   // B (4,3)
	_, _ = eng.ApplyMove(makeDropMove("usr_alice", 0)) // A (4,0)
	_, _ = eng.ApplyMove(makeDropMove("usr_bob", 3))   // B (3,3)

	// Alice drops in Col 3 -> lands on (2,3) to complete diagonal (5,0)-(4,1)-(3,2)-(2,3)
	state, err := eng.ApplyMove(makeDropMove("usr_alice", 3))
	if err != nil {
		t.Fatalf("failed move: %v", err)
	}

	if !eng.IsFinished() || state.Result == nil || state.Result.WinnerID != "usr_alice" {
		t.Fatalf("expected Alice diagonal win, got %+v", state.Result)
	}
}

// 5. Test Column Full Error Validation
func TestConnect4_ColumnFullError(t *testing.T) {
	eng := connect4.NewConnect4Engine()
	_, _ = eng.Initialize("c4_full", makePlayers(), nil)

	// Fill Col 0 with 6 discs
	for i := 0; i < 6; i++ {
		if i%2 == 0 {
			_, _ = eng.ApplyMove(makeDropMove("usr_alice", 0))
		} else {
			_, _ = eng.ApplyMove(makeDropMove("usr_bob", 0))
		}
	}

	// Alice attempts 7th disc in Col 0 -> rejected ErrCellOccupied
	err := eng.ValidateMove(makeDropMove("usr_alice", 0))
	if err == nil {
		t.Fatalf("expected error when dropping in full column, got nil")
	}
}

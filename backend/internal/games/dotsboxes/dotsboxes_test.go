package dotsboxes_test

import (
	"encoding/json"
	"testing"

	"github.com/Abh19avM/recess/internal/games/dotsboxes"
	"github.com/Abh19avM/recess/internal/games/engine"
)

func makePlayers() []engine.Player {
	return []engine.Player{
		{ID: "usr_alice", Username: "Alice", SeatNumber: 1},
		{ID: "usr_bob", Username: "Bob", SeatNumber: 2},
	}
}

func makeMove(playerID, edgeType string, row, col int) engine.Move {
	data, _ := json.Marshal(dotsboxes.EdgeKey{Type: edgeType, Row: row, Col: col})
	return engine.Move{
		PlayerID:  playerID,
		Action:    "draw_edge",
		Data:      data,
		Timestamp: 1000,
	}
}

// 1. Test Board Initialization
func TestDotsBoxes_Initialization(t *testing.T) {
	eng := dotsboxes.NewDotsBoxesEngine()
	players := makePlayers()

	cfg, _ := json.Marshal(map[string]int{"rows": 2, "cols": 2})
	state, err := eng.Initialize("db_init", players, cfg)
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}

	var board dotsboxes.BoardState
	_ = json.Unmarshal(state.BoardState, &board)
	if board.Rows != 2 || board.Cols != 2 {
		t.Fatalf("expected 2x2 board, got %dx%d", board.Rows, board.Cols)
	}
	if board.TotalBoxes != 4 {
		t.Fatalf("expected 4 total boxes, got %d", board.TotalBoxes)
	}
	if state.CurrentTurn != "usr_alice" {
		t.Fatalf("expected Alice turn first")
	}
}

// 2. Test Edge Placement & Turn Alternation
func TestDotsBoxes_TurnAlternation(t *testing.T) {
	eng := dotsboxes.NewDotsBoxesEngine()
	_, _ = eng.Initialize("db_turn", makePlayers(), nil)

	// Alice draws top edge (h, 0, 0)
	state, err := eng.ApplyMove(makeMove("usr_alice", "h", 0, 0))
	if err != nil {
		t.Fatalf("failed move: %v", err)
	}
	if state.CurrentTurn != "usr_bob" {
		t.Fatalf("expected Bob's turn after no box completed, got %s", state.CurrentTurn)
	}

	// Bob draws right edge (v, 0, 1)
	state, err = eng.ApplyMove(makeMove("usr_bob", "v", 0, 1))
	if err != nil {
		t.Fatalf("failed move: %v", err)
	}
	if state.CurrentTurn != "usr_alice" {
		t.Fatalf("expected Alice's turn, got %s", state.CurrentTurn)
	}
}

// 3. Test Single Box Completion and Bonus Turn
func TestDotsBoxes_SingleBoxCompletionAndBonusTurn(t *testing.T) {
	eng := dotsboxes.NewDotsBoxesEngine()
	_, _ = eng.Initialize("db_box", makePlayers(), nil)

	// Build 3 sides of box (0,0):
	// Top: (h, 0, 0) - Alice
	// Left: (v, 0, 0) - Bob
	// Right: (v, 0, 1) - Alice
	_, _ = eng.ApplyMove(makeMove("usr_alice", "h", 0, 0))
	_, _ = eng.ApplyMove(makeMove("usr_bob", "v", 0, 0))
	_, _ = eng.ApplyMove(makeMove("usr_alice", "v", 0, 1))

	// Bob closes box (0,0) with Bottom edge: (h, 1, 0) -> Bob should score and get BONUS TURN!
	state, err := eng.ApplyMove(makeMove("usr_bob", "h", 1, 0))
	if err != nil {
		t.Fatalf("failed to complete box: %v", err)
	}

	var board dotsboxes.BoardState
	_ = json.Unmarshal(state.BoardState, &board)
	if board.Scores["usr_bob"] != 1 {
		t.Fatalf("expected Bob score to be 1, got %d", board.Scores["usr_bob"])
	}
	if board.Boxes[0][0] != "usr_bob" {
		t.Fatalf("expected box (0,0) owned by Bob, got %s", board.Boxes[0][0])
	}
	if state.CurrentTurn != "usr_bob" {
		t.Fatalf("expected Bob to retain turn (Bonus Turn), got %s", state.CurrentTurn)
	}
}

// 4. Test Double Box Completion (shared middle edge closes 2 boxes at once!)
func TestDotsBoxes_DoubleBoxCompletion(t *testing.T) {
	eng := dotsboxes.NewDotsBoxesEngine()
	cfg, _ := json.Marshal(map[string]int{"rows": 2, "cols": 1}) // 2 stacked boxes
	_, _ = eng.Initialize("db_double", makePlayers(), cfg)

	// Setup top box (0,0) and bottom box (1,0) leaving shared middle edge (h, 1, 0) open:
	// Top: (h, 0, 0)
	// Left 1: (v, 0, 0)
	// Right 1: (v, 0, 1)
	// Bottom: (h, 2, 0)
	// Left 2: (v, 1, 0)
	// Right 2: (v, 1, 1)
	_, _ = eng.ApplyMove(makeMove("usr_alice", "h", 0, 0))
	_, _ = eng.ApplyMove(makeMove("usr_bob", "v", 0, 0))
	_, _ = eng.ApplyMove(makeMove("usr_alice", "v", 0, 1))
	_, _ = eng.ApplyMove(makeMove("usr_bob", "h", 2, 0))
	_, _ = eng.ApplyMove(makeMove("usr_alice", "v", 1, 0))
	_, _ = eng.ApplyMove(makeMove("usr_bob", "v", 1, 1))

	// Alice now draws the shared middle edge (h, 1, 0) -> Closes BOTH boxes (0,0) and (1,0)!
	state, err := eng.ApplyMove(makeMove("usr_alice", "h", 1, 0))
	if err != nil {
		t.Fatalf("failed double box move: %v", err)
	}

	var board dotsboxes.BoardState
	_ = json.Unmarshal(state.BoardState, &board)
	if board.Scores["usr_alice"] != 2 {
		t.Fatalf("expected Alice score 2 for double box, got %d", board.Scores["usr_alice"])
	}
	if !eng.IsFinished() {
		t.Fatalf("expected game finished (2/2 boxes claimed)")
	}
	if state.Result == nil || state.Result.WinnerID != "usr_alice" {
		t.Fatalf("expected Alice victory, got result: %+v", state.Result)
	}
}

// 5. Test Error Handling (Occupied edge, out of bounds, wrong player)
func TestDotsBoxes_Errors(t *testing.T) {
	eng := dotsboxes.NewDotsBoxesEngine()
	_, _ = eng.Initialize("db_err", makePlayers(), nil)

	// Wrong turn
	err := eng.ValidateMove(makeMove("usr_bob", "h", 0, 0))
	if err != engine.ErrNotPlayerTurn {
		t.Fatalf("expected ErrNotPlayerTurn, got %v", err)
	}

	// Out of bounds
	err = eng.ValidateMove(makeMove("usr_alice", "h", 4, 0))
	if err != engine.ErrInvalidCoordinates {
		t.Fatalf("expected ErrInvalidCoordinates, got %v", err)
	}

	// Alice draws (h, 0, 0)
	_, _ = eng.ApplyMove(makeMove("usr_alice", "h", 0, 0))

	// Bob attempts to redraw (h, 0, 0)
	err = eng.ValidateMove(makeMove("usr_bob", "h", 0, 0))
	if err != engine.ErrCellOccupied {
		t.Fatalf("expected ErrCellOccupied, got %v", err)
	}
}

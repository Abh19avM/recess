package xo_test

import (
	"encoding/json"
	"testing"

	"github.com/Abh19avM/recess/internal/games/engine"
	"github.com/Abh19avM/recess/internal/games/xo"
)

func makePlayers() []engine.Player {
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

// 1. Test Horizontal Wins (Row 0, Row 1, Row 2)
func TestXOEngine_HorizontalWins(t *testing.T) {
	for row := 0; row < 3; row++ {
		t.Run("Row_"+string(rune('0'+row)), func(t *testing.T) {
			eng := xo.NewXOEngine()
			players := makePlayers()
			_, err := eng.Initialize("game_h", players, nil)
			if err != nil {
				t.Fatalf("failed to init: %v", err)
			}

			otherRow := (row + 1) % 3

			// Alice marks (row, 0), Bob marks (otherRow, 0)
			// Alice marks (row, 1), Bob marks (otherRow, 1)
			// Alice marks (row, 2) -> Wins!
			moves := []engine.Move{
				makeMove("usr_alice", row, 0),
				makeMove("usr_bob", otherRow, 0),
				makeMove("usr_alice", row, 1),
				makeMove("usr_bob", otherRow, 1),
				makeMove("usr_alice", row, 2),
			}

			for _, m := range moves {
				if _, err := eng.ApplyMove(m); err != nil {
					t.Fatalf("unexpected move error on row %d: %v", row, err)
				}
			}

			if !eng.IsFinished() {
				t.Fatalf("expected game to be finished on row %d win", row)
			}
			res := eng.Result()
			if res == nil || res.WinnerID != "usr_alice" || res.IsDraw {
				t.Fatalf("expected Alice to win on row %d, got: %+v", row, res)
			}
		})
	}
}

// 2. Test Vertical Wins (Col 0, Col 1, Col 2)
func TestXOEngine_VerticalWins(t *testing.T) {
	for col := 0; col < 3; col++ {
		t.Run("Col_"+string(rune('0'+col)), func(t *testing.T) {
			eng := xo.NewXOEngine()
			players := makePlayers()
			_, _ = eng.Initialize("game_v", players, nil)

			otherCol := (col + 1) % 3

			moves := []engine.Move{
				makeMove("usr_alice", 0, col),
				makeMove("usr_bob", 0, otherCol),
				makeMove("usr_alice", 1, col),
				makeMove("usr_bob", 1, otherCol),
				makeMove("usr_alice", 2, col),
			}

			for _, m := range moves {
				if _, err := eng.ApplyMove(m); err != nil {
					t.Fatalf("unexpected move error on col %d: %v", col, err)
				}
			}

			if !eng.IsFinished() || eng.Result().WinnerID != "usr_alice" {
				t.Fatalf("expected Alice vertical win on col %d, got: %+v", col, eng.Result())
			}
		})
	}
}

// 3. Test Diagonal Wins (Main diagonal & Anti-diagonal)
func TestXOEngine_DiagonalWins(t *testing.T) {
	t.Run("MainDiagonal", func(t *testing.T) {
		eng := xo.NewXOEngine()
		_, _ = eng.Initialize("game_d1", makePlayers(), nil)

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
			t.Fatalf("expected Alice main diagonal win")
		}
	})

	t.Run("AntiDiagonal", func(t *testing.T) {
		eng := xo.NewXOEngine()
		_, _ = eng.Initialize("game_d2", makePlayers(), nil)

		moves := []engine.Move{
			makeMove("usr_alice", 0, 2),
			makeMove("usr_bob", 0, 0),
			makeMove("usr_alice", 1, 1),
			makeMove("usr_bob", 0, 1),
			makeMove("usr_alice", 2, 0),
		}
		for _, m := range moves {
			if _, err := eng.ApplyMove(m); err != nil {
				t.Fatalf("move error: %v", err)
			}
		}
		if !eng.IsFinished() || eng.Result().WinnerID != "usr_alice" {
			t.Fatalf("expected Alice anti-diagonal win")
		}
	})
}

// 4. Test Draw (Cat's Game)
func TestXOEngine_DrawGame(t *testing.T) {
	eng := xo.NewXOEngine()
	_, _ = eng.Initialize("game_draw", makePlayers(), nil)

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

	for _, m := range moves {
		if _, err := eng.ApplyMove(m); err != nil {
			t.Fatalf("move error: %v", err)
		}
	}

	if !eng.IsFinished() {
		t.Fatalf("expected game finished on draw")
	}
	res := eng.Result()
	if res == nil || !res.IsDraw || res.WinnerID != "" {
		t.Fatalf("expected draw result, got: %+v", res)
	}
}

// 5. Test Invalid Moves, Wrong Player, and Move After Game Completion
func TestXOEngine_ErrorHandling(t *testing.T) {
	eng := xo.NewXOEngine()
	_, _ = eng.Initialize("game_err", makePlayers(), nil)

	// Wrong player turn
	err := eng.ValidateMove(makeMove("usr_bob", 0, 0))
	if err != engine.ErrNotPlayerTurn {
		t.Fatalf("expected ErrNotPlayerTurn, got %v", err)
	}

	// Unknown player ID
	err = eng.ValidateMove(makeMove("usr_unknown", 0, 0))
	if err != engine.ErrNotPlayerTurn {
		t.Fatalf("expected ErrNotPlayerTurn for unknown player, got %v", err)
	}

	// Out of bounds
	err = eng.ValidateMove(makeMove("usr_alice", -1, 0))
	if err != engine.ErrInvalidCoordinates {
		t.Fatalf("expected ErrInvalidCoordinates, got %v", err)
	}
	err = eng.ValidateMove(makeMove("usr_alice", 3, 2))
	if err != engine.ErrInvalidCoordinates {
		t.Fatalf("expected ErrInvalidCoordinates, got %v", err)
	}

	// Alice moves at (1,1)
	_, err = eng.ApplyMove(makeMove("usr_alice", 1, 1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Bob attempts to overwrite (1,1)
	err = eng.ValidateMove(makeMove("usr_bob", 1, 1))
	if err != engine.ErrCellOccupied {
		t.Fatalf("expected ErrCellOccupied, got %v", err)
	}

	// Complete the game (Alice wins)
	_, _ = eng.ApplyMove(makeMove("usr_bob", 0, 0))
	_, _ = eng.ApplyMove(makeMove("usr_alice", 1, 0))
	_, _ = eng.ApplyMove(makeMove("usr_bob", 0, 1))
	_, _ = eng.ApplyMove(makeMove("usr_alice", 1, 2)) // Alice wins row 1

	if !eng.IsFinished() {
		t.Fatalf("expected game finished")
	}

	// Move after completion must be rejected
	err = eng.ValidateMove(makeMove("usr_bob", 2, 2))
	if err != engine.ErrGameFinished {
		t.Fatalf("expected ErrGameFinished, got %v", err)
	}

	_, err = eng.ApplyMove(makeMove("usr_bob", 2, 2))
	if err != engine.ErrGameFinished {
		t.Fatalf("expected ErrGameFinished on ApplyMove, got %v", err)
	}
}

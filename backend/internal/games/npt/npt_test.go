package npt_test

import (
	"encoding/json"
	"testing"

	"github.com/Abh19avM/recess/internal/games/engine"
	"github.com/Abh19avM/recess/internal/games/npt"
)

func makePlayers() []engine.Player {
	return []engine.Player{
		{ID: "usr_alice", Username: "Alice", SeatNumber: 1},
		{ID: "usr_bob", Username: "Bob", SeatNumber: 2},
	}
}

func makeSubmitMove(playerID string, name, place, animal, thing string) engine.Move {
	data, _ := json.Marshal(npt.CategoryEntries{
		Name:   name,
		Place:  place,
		Animal: animal,
		Thing:  thing,
	})
	return engine.Move{
		PlayerID:  playerID,
		Action:    "submit_entries",
		Data:      data,
		Timestamp: 1000,
	}
}

// 1. Test Initialization
func TestNPAT_Initialization(t *testing.T) {
	eng := npt.NewNPATEngine()
	state, err := eng.Initialize("npt_init", makePlayers(), nil)
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}

	var board npt.BoardState
	_ = json.Unmarshal(state.BoardState, &board)
	if board.CurrentRound != 1 || board.CurrentLetter != "S" {
		t.Fatalf("expected Round 1 with Letter S, got %+v", board)
	}
}

// 2. Test Scoring (Unique vs Shared vs Invalid)
func TestNPAT_ScoringRules(t *testing.T) {
	eng := npt.NewNPATEngine()
	_, _ = eng.Initialize("npt_score", makePlayers(), nil)

	// Round 1 (Letter: S)
	// Alice: Name: "Sarah", Place: "Seattle", Animal: "Snake", Thing: "Scissors"
	// Bob:   Name: "Sarah" (shared = 5), Place: "Sydney" (unique = 10), Animal: "Tiger" (wrong letter = 0), Thing: "Sword" (unique = 10)
	_, _ = eng.ApplyMove(makeSubmitMove("usr_alice", "Sarah", "Seattle", "Snake", "Scissors"))
	state, err := eng.ApplyMove(makeSubmitMove("usr_bob", "Sarah", "Sydney", "Tiger", "Sword"))
	if err != nil {
		t.Fatalf("failed submit move: %v", err)
	}

	var board npt.BoardState
	_ = json.Unmarshal(state.BoardState, &board)

	// Alice: 5 (shared Sarah) + 10 (Seattle) + 10 (Snake) + 10 (Scissors) = 35
	if board.CumulativeScores["usr_alice"] != 35 {
		t.Fatalf("expected Alice to have 35 points, got %d", board.CumulativeScores["usr_alice"])
	}

	// Bob: 5 (shared Sarah) + 10 (Sydney) + 0 (Tiger) + 10 (Sword) = 25
	if board.CumulativeScores["usr_bob"] != 25 {
		t.Fatalf("expected Bob to have 25 points, got %d", board.CumulativeScores["usr_bob"])
	}

	// Board should have advanced to Round 2 (Letter A)
	if board.CurrentRound != 2 || board.CurrentLetter != "A" {
		t.Fatalf("expected Round 2 with Letter A, got %+v", board)
	}
}

// 3. Test Full Match to Completion
func TestNPAT_FullMatchCompletion(t *testing.T) {
	eng := npt.NewNPATEngine()
	_, _ = eng.Initialize("npt_full", makePlayers(), nil)

	// Round 1 (S): Alice = 40, Bob = 0
	_, _ = eng.ApplyMove(makeSubmitMove("usr_alice", "Sam", "Spain", "Seal", "Spoon"))
	_, _ = eng.ApplyMove(makeSubmitMove("usr_bob", "", "", "", ""))

	// Round 2 (A): Alice = 40, Bob = 0
	_, _ = eng.ApplyMove(makeSubmitMove("usr_alice", "Alex", "Austin", "Ant", "Apple"))
	_, _ = eng.ApplyMove(makeSubmitMove("usr_bob", "", "", "", ""))

	// Round 3 (M): Alice = 40, Bob = 0 -> Triggers Match Completion!
	_, _ = eng.ApplyMove(makeSubmitMove("usr_alice", "Max", "Madrid", "Monkey", "Mirror"))
	state, err := eng.ApplyMove(makeSubmitMove("usr_bob", "", "", "", ""))
	if err != nil {
		t.Fatalf("failed round 3: %v", err)
	}

	if !eng.IsFinished() {
		t.Fatalf("expected match to finish after 3 rounds")
	}
	if state.Result == nil || state.Result.WinnerID != "usr_alice" {
		t.Fatalf("expected Alice victory, got %+v", state.Result)
	}
}

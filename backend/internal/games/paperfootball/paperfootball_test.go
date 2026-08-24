package paperfootball_test

import (
	"encoding/json"
	"testing"

	"github.com/Abh19avM/recess/internal/games/engine"
	"github.com/Abh19avM/recess/internal/games/paperfootball"
)

func makePlayers() []engine.Player {
	return []engine.Player{
		{ID: "usr_alice", Username: "Alice", SeatNumber: 1},
		{ID: "usr_bob", Username: "Bob", SeatNumber: 2},
	}
}

func makeFlick(playerID string, action string, power, angle float64) engine.Move {
	data, _ := json.Marshal(paperfootball.FlickMove{
		Action: action,
		Power:  power,
		Angle:  angle,
	})
	return engine.Move{
		PlayerID:  playerID,
		Action:    action,
		Data:      data,
		Timestamp: 1000,
	}
}

// 1. Test Board Initialization
func TestPaperFootball_Initialization(t *testing.T) {
	eng := paperfootball.NewPaperFootballEngine()
	state, err := eng.Initialize("pf_init", makePlayers(), nil)
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}

	if state.CurrentTurn != "usr_alice" {
		t.Fatalf("expected Alice to have initial possession")
	}

	var board paperfootball.BoardState
	_ = json.Unmarshal(state.BoardState, &board)
	if board.BallPosition != 20.0 || board.Down != 1 {
		t.Fatalf("expected 20%% start and down 1, got %+v", board)
	}
}

// 2. Test Touchdown and Extra Point Kick
func TestPaperFootball_TouchdownAndExtraPoint(t *testing.T) {
	eng := paperfootball.NewPaperFootballEngine()
	_, _ = eng.Initialize("pf_td", makePlayers(), nil)

	// Alice flicks with 100 power -> 20 + 75 = 95% (Within 90-100% Touchdown zone!)
	state, err := eng.ApplyMove(makeFlick("usr_alice", "flick", 100, 0))
	if err != nil {
		t.Fatalf("failed flick move: %v", err)
	}

	var board paperfootball.BoardState
	_ = json.Unmarshal(state.BoardState, &board)
	if board.Scores["usr_alice"] != 6 {
		t.Fatalf("expected 6 points for Touchdown, got %d", board.Scores["usr_alice"])
	}
	if board.Phase != paperfootball.PhaseExtraPoint {
		t.Fatalf("expected ExtraPoint phase, got %s", board.Phase)
	}

	// Alice kicks extra point through uprights (power 50, angle 0)
	state, err = eng.ApplyMove(makeFlick("usr_alice", "extra_point", 50, 0))
	if err != nil {
		t.Fatalf("failed extra point: %v", err)
	}

	_ = json.Unmarshal(state.BoardState, &board)
	if board.Scores["usr_alice"] != 7 {
		t.Fatalf("expected 7 points after good extra point, got %d", board.Scores["usr_alice"])
	}
	if board.PossessionID != "usr_bob" {
		t.Fatalf("expected Bob to take possession after score, got %s", board.PossessionID)
	}
}

// 3. Test Over-flick Table Fall (Turnover)
func TestPaperFootball_TableFallTurnover(t *testing.T) {
	eng := paperfootball.NewPaperFootballEngine()
	_, _ = eng.Initialize("pf_fall", makePlayers(), nil)

	// Advance to 60% first
	_, _ = eng.ApplyMove(makeFlick("usr_alice", "flick", 50, 0)) // 20 + 37.5 = 57.5%

	// Alice over-flicks with 100 power -> 57.5 + 75 = 132.5% (> 100% table fall!)
	state, err := eng.ApplyMove(makeFlick("usr_alice", "flick", 100, 0))
	if err != nil {
		t.Fatalf("failed flick: %v", err)
	}

	var board paperfootball.BoardState
	_ = json.Unmarshal(state.BoardState, &board)
	if board.PossessionID != "usr_bob" {
		t.Fatalf("expected turnover to Bob on table fall")
	}
	if board.LastFlickResult != "table_fall" {
		t.Fatalf("expected table_fall result, got %s", board.LastFlickResult)
	}
}

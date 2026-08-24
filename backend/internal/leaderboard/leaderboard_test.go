package leaderboard_test

import (
	"context"
	"testing"

	"github.com/Abh19avM/recess/internal/leaderboard"
)

// 1. Test Elo Calculation Formula
func TestEloCalculation_EqualRatings(t *testing.T) {
	// P1: 1200, P2: 1200, P1 Wins -> Expected score 0.5 each -> delta = 32 * (1.0 - 0.5) = +16
	res := leaderboard.CalculateElo(1200, 1200, 1.0)
	if res.Player1Delta != 16 || res.Player2Delta != -16 {
		t.Fatalf("expected delta +16/-16 for equal ratings win, got %+v", res)
	}
	if res.Player1RatingAfter != 1216 || res.Player2RatingAfter != 1184 {
		t.Fatalf("expected ratings 1216/1184, got %+v", res)
	}
}

func TestEloCalculation_Draw(t *testing.T) {
	// P1: 1200, P2: 1200, Draw (0.5) -> delta = 0
	res := leaderboard.CalculateElo(1200, 1200, 0.5)
	if res.Player1Delta != 0 || res.Player2Delta != 0 {
		t.Fatalf("expected 0 delta on draw between equal ratings, got %+v", res)
	}
}

func TestEloCalculation_UnderdogVictory(t *testing.T) {
	// Underdog (1000) beats Master (1400) -> Higher delta gained by underdog
	res := leaderboard.CalculateElo(1000, 1400, 1.0)
	if res.Player1Delta <= 20 {
		t.Fatalf("expected high reward (>20) for major underdog upset, got %+v", res)
	}
	if res.Player2Delta >= -20 {
		t.Fatalf("expected high penalty (<-20) for favorite loss, got %+v", res)
	}
}

// 2. Test Match Outcome Recording & History Retrieval
func TestLeaderboardService_RecordMatchAndHistory(t *testing.T) {
	svc := leaderboard.NewService(nil, nil)
	ctx := context.Background()

	record := leaderboard.MatchOutcomeRecord{
		MatchID:       "match_test_01",
		RoomCode:      "RECESS-HC-TEST",
		GameType:      "hand_cricket",
		Player1ID:     "usr_alice",
		Player2ID:     "usr_bob",
		WinnerID:      "usr_alice",
		Player1Score:  42,
		Player2Score:  30,
		DurationSecs:  180,
	}

	eloRes, err := svc.RecordMatchOutcome(ctx, record)
	if err != nil {
		t.Fatalf("failed to record match outcome: %v", err)
	}
	if eloRes.Player1Delta <= 0 {
		t.Fatalf("expected positive delta for winner Alice, got %d", eloRes.Player1Delta)
	}

	// Verify Alice match history
	history, err := svc.GetMatchHistory(ctx, "usr_alice", 10)
	if err != nil || len(history) == 0 {
		t.Fatalf("expected Alice to have recorded history, got %v / len %d", err, len(history))
	}
	if !history[0].IsWinner || history[0].Score != 42 {
		t.Fatalf("unexpected history item: %+v", history[0])
	}
}

// 3. Test Achievements Retrieval
func TestLeaderboardService_Achievements(t *testing.T) {
	svc := leaderboard.NewService(nil, nil)
	ctx := context.Background()

	achievements, err := svc.GetUserAchievements(ctx, "usr_alice")
	if err != nil {
		t.Fatalf("failed to get achievements: %v", err)
	}
	if len(achievements) < 5 {
		t.Fatalf("expected at least 5 standard achievements, got %d", len(achievements))
	}
}

package matchmaking_test

import (
	"context"
	"testing"

	"github.com/Abh19avM/recess/internal/matchmaking"
)

// 1. Test Casual Matchmaking Pairing
func TestMatchmaking_CasualInstantPairing(t *testing.T) {
	svc := matchmaking.NewService(nil)
	ctx := context.Background()

	t1 := &matchmaking.MatchTicket{
		TicketID: "tkt_alice",
		UserID:   "usr_alice",
		Username: "Alice",
		GameType: "hand_cricket",
		Mode:     matchmaking.ModeCasual,
	}

	// Alice joins queue
	match1, err := svc.Enqueue(ctx, t1)
	if err != nil {
		t.Fatalf("failed to enqueue Alice: %v", err)
	}
	if match1 != nil {
		t.Fatalf("expected nil match when first player enqueues, got %+v", match1)
	}

	if sz := svc.GetQueueSize(ctx, "hand_cricket", matchmaking.ModeCasual); sz != 1 {
		t.Fatalf("expected queue size 1, got %d", sz)
	}

	// Bob joins queue -> should instantly pair with Alice!
	t2 := &matchmaking.MatchTicket{
		TicketID: "tkt_bob",
		UserID:   "usr_bob",
		Username: "Bob",
		GameType: "hand_cricket",
		Mode:     matchmaking.ModeCasual,
	}

	match2, err := svc.Enqueue(ctx, t2)
	if err != nil {
		t.Fatalf("failed to enqueue Bob: %v", err)
	}
	if match2 == nil {
		t.Fatalf("expected instant match for Bob, got nil")
	}

	if match2.Player1.UserID != "usr_alice" || match2.Player2.UserID != "usr_bob" {
		t.Fatalf("unexpected players in match: %+v", match2)
	}
	if match2.GameType != "hand_cricket" {
		t.Fatalf("expected hand_cricket, got %s", match2.GameType)
	}

	// Queue should now be empty
	if sz := svc.GetQueueSize(ctx, "hand_cricket", matchmaking.ModeCasual); sz != 0 {
		t.Fatalf("expected queue size 0 after pairing, got %d", sz)
	}

	// Alice polling match status should find the match
	polled, found, err := svc.PollMatch(ctx, "tkt_alice")
	if err != nil || !found || polled == nil {
		t.Fatalf("expected Alice to find match on poll: %v / %v", found, err)
	}
}

// 2. Test Ranked Matchmaking Rating Boundaries
func TestMatchmaking_RankedRatingMatch(t *testing.T) {
	svc := matchmaking.NewService(nil)
	ctx := context.Background()

	// Player 1: Rating 1000
	t1 := &matchmaking.MatchTicket{
		TicketID: "tkt_pro1",
		UserID:   "usr_pro1",
		Username: "Pro1",
		GameType: "xo",
		Mode:     matchmaking.ModeRanked,
		Rating:   1000,
	}
	_, _ = svc.Enqueue(ctx, t1)

	// Player 2: Rating 1500 (Diff 500 > 200 threshold) -> Should NOT match
	t2 := &matchmaking.MatchTicket{
		TicketID: "tkt_expert",
		UserID:   "usr_expert",
		Username: "Expert",
		GameType: "xo",
		Mode:     matchmaking.ModeRanked,
		Rating:   1500,
	}
	m2, _ := svc.Enqueue(ctx, t2)
	if m2 != nil {
		t.Fatalf("expected no match for rating gap 500")
	}

	// Player 3: Rating 1100 (Diff 100 <= 200) -> Should pair with Player 1!
	t3 := &matchmaking.MatchTicket{
		TicketID: "tkt_pro2",
		UserID:   "usr_pro2",
		Username: "Pro2",
		GameType: "xo",
		Mode:     matchmaking.ModeRanked,
		Rating:   1100,
	}
	m3, _ := svc.Enqueue(ctx, t3)
	if m3 == nil {
		t.Fatalf("expected match for compatible rating 1100")
	}
	if m3.Player1.UserID != "usr_pro1" || m3.Player2.UserID != "usr_pro2" {
		t.Fatalf("unexpected paired players: %+v", m3)
	}
}

// 3. Test Queue Cancellation / Dequeue
func TestMatchmaking_Cancellation(t *testing.T) {
	svc := matchmaking.NewService(nil)
	ctx := context.Background()

	ticket := &matchmaking.MatchTicket{
		TicketID: "tkt_cancel",
		UserID:   "usr_charlie",
		Username: "Charlie",
		GameType: "dots_boxes",
		Mode:     matchmaking.ModeCasual,
	}
	_, _ = svc.Enqueue(ctx, ticket)

	if sz := svc.GetQueueSize(ctx, "dots_boxes", matchmaking.ModeCasual); sz != 1 {
		t.Fatalf("expected queue size 1, got %d", sz)
	}

	// Cancel queue
	err := svc.Dequeue(ctx, "usr_charlie", "dots_boxes")
	if err != nil {
		t.Fatalf("failed dequeue: %v", err)
	}

	if sz := svc.GetQueueSize(ctx, "dots_boxes", matchmaking.ModeCasual); sz != 0 {
		t.Fatalf("expected queue size 0 after cancellation, got %d", sz)
	}
}

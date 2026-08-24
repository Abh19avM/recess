package matchmaking

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/Abh19avM/recess/internal/metrics"
	"github.com/redis/go-redis/v9"
)

var (
	ErrAlreadyInQueue = errors.New("player is already in matchmaking queue")
	ErrTicketNotFound = errors.New("matchmaking ticket not found or expired")
)

// Service defines the matchmaking interface.
type Service interface {
	Enqueue(ctx context.Context, ticket *MatchTicket) (*MatchResult, error)
	Dequeue(ctx context.Context, userID, gameType string) error
	PollMatch(ctx context.Context, ticketID string) (*MatchResult, bool, error)
	GetQueueSize(ctx context.Context, gameType string, mode MatchMode) int
}

type service struct {
	mu           sync.RWMutex
	rdb          *redis.Client
	memQueues    map[string][]*MatchTicket // key: gameType:mode -> []*MatchTicket
	memMatches   map[string]*MatchResult   // ticketID -> *MatchResult
	userTickets  map[string]string         // userID -> ticketID
	ticketExpiry time.Duration
}

// NewService creates a new matchmaking service instance.
func NewService(rdb *redis.Client) Service {
	return &service{
		rdb:          rdb,
		memQueues:    make(map[string][]*MatchTicket),
		memMatches:   make(map[string]*MatchResult),
		userTickets:  make(map[string]string),
		ticketExpiry: 90 * time.Second,
	}
}

// Enqueue adds a player to the matchmaking queue or instantly pairs them if an opponent is waiting.
func (s *service) Enqueue(ctx context.Context, ticket *MatchTicket) (*MatchResult, error) {
	if ticket.Mode == "" {
		ticket.Mode = ModeCasual
	}
	if ticket.TicketID == "" {
		ticket.TicketID = "tkt_" + randomHex(8)
	}
	if ticket.Rating == 0 {
		ticket.Rating = 1000 // Default starter ELO
	}
	ticket.JoinedAt = time.Now().UTC()
	ticket.ExpiresAt = ticket.JoinedAt.Add(s.ticketExpiry)

	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Check if player is already queued in memory or update ticket
	if oldTicketID, exists := s.userTickets[ticket.UserID]; exists && oldTicketID != ticket.TicketID {
		s.cleanupTicketLocked(ticket.UserID, ticket.GameType)
	}

	queueKey := fmt.Sprintf("%s:%s", ticket.GameType, ticket.Mode)
	waitingList := s.memQueues[queueKey]

	// 2. Search for a compatible waiting opponent
	now := time.Now().UTC()
	var matchedOpponent *MatchTicket
	var remaining []*MatchTicket

	for _, opp := range waitingList {
		// Filter out expired tickets or same user
		if opp.ExpiresAt.Before(now) || opp.UserID == ticket.UserID {
			continue
		}

		if matchedOpponent == nil && s.isCompatible(ticket, opp) {
			matchedOpponent = opp
		} else {
			remaining = append(remaining, opp)
		}
	}

	// 3. If an opponent was found, create match!
	if matchedOpponent != nil {
		s.memQueues[queueKey] = remaining
		metrics.MatchmakingQueueSize.WithLabelValues(ticket.GameType).Set(float64(len(remaining)))
		metrics.MatchmakingLatency.WithLabelValues(ticket.GameType).Observe(time.Since(matchedOpponent.JoinedAt).Seconds())

		delete(s.userTickets, ticket.UserID)
		delete(s.userTickets, matchedOpponent.UserID)

		matchID := "match_" + randomHex(8)
		roomID := fmt.Sprintf("MATCH-%s-%s", strings.ToUpper(ticket.GameType), randomHex(4))

		matchResult := &MatchResult{
			MatchID:  matchID,
			RoomID:   roomID,
			GameType: ticket.GameType,
			Mode:     ticket.Mode,
			Player1: MatchedPlayer{
				UserID:       matchedOpponent.UserID,
				Username:     matchedOpponent.Username,
				AvatarPreset: matchedOpponent.AvatarPreset,
				Rating:       matchedOpponent.Rating,
				SeatNumber:   1,
			},
			Player2: MatchedPlayer{
				UserID:       ticket.UserID,
				Username:     ticket.Username,
				AvatarPreset: ticket.AvatarPreset,
				Rating:       ticket.Rating,
				SeatNumber:   2,
			},
			MatchedAt: time.Now().UTC(),
		}

		// Store match result for polling lookups
		s.memMatches[ticket.TicketID] = matchResult
		s.memMatches[matchedOpponent.TicketID] = matchResult

		// If Redis is active, cache match result in Redis
		if s.rdb != nil {
			matchData, _ := json.Marshal(matchResult)
			s.rdb.Set(ctx, "match:result:"+ticket.TicketID, matchData, 5*time.Minute)
			s.rdb.Set(ctx, "match:result:"+matchedOpponent.TicketID, matchData, 5*time.Minute)
		}

		return matchResult, nil
	}

	// 4. No opponent available yet -> add to queue
	remaining = append(remaining, ticket)
	s.memQueues[queueKey] = remaining
	s.userTickets[ticket.UserID] = ticket.TicketID
	metrics.MatchmakingQueueSize.WithLabelValues(ticket.GameType).Set(float64(len(remaining)))

	return nil, nil // Nil match result means successfully queued
}

// Dequeue cancels a waiting queue ticket for a player.
func (s *service) Dequeue(ctx context.Context, userID, gameType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupTicketLocked(userID, gameType)
	return nil
}

// PollMatch checks if a ticket has been matched with an opponent.
func (s *service) PollMatch(ctx context.Context, ticketID string) (*MatchResult, bool, error) {
	s.mu.RLock()
	match, exists := s.memMatches[ticketID]
	s.mu.RUnlock()

	if exists {
		return match, true, nil
	}

	// Fallback to Redis lookup
	if s.rdb != nil {
		val, err := s.rdb.Get(ctx, "match:result:"+ticketID).Result()
		if err == nil {
			var r MatchResult
			if err := json.Unmarshal([]byte(val), &r); err == nil {
				return &r, true, nil
			}
		}
	}

	return nil, false, nil
}

// GetQueueSize returns the number of active waiting players in a queue.
func (s *service) GetQueueSize(ctx context.Context, gameType string, mode MatchMode) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	queueKey := fmt.Sprintf("%s:%s", gameType, mode)
	now := time.Now().UTC()
	count := 0
	for _, t := range s.memQueues[queueKey] {
		if t.ExpiresAt.After(now) {
			count++
		}
	}
	return count
}

func (s *service) isCompatible(t1, t2 *MatchTicket) bool {
	if t1.Mode == ModeCasual {
		return true // Instant match for casual
	}

	// For ranked: match within +/- 200 ELO rating window
	diff := math.Abs(float64(t1.Rating - t2.Rating))
	return diff <= 200
}

func (s *service) cleanupTicketLocked(userID, gameType string) {
	ticketID, exists := s.userTickets[userID]
	if !exists {
		return
	}
	delete(s.userTickets, userID)

	for key, queue := range s.memQueues {
		var filtered []*MatchTicket
		for _, t := range queue {
			if t.TicketID != ticketID && t.UserID != userID {
				filtered = append(filtered, t)
			}
		}
		s.memQueues[key] = filtered
	}
}

func randomHex(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

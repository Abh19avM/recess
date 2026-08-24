package leaderboard

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Service defines leaderboard, match history, and progression functionality.
type Service interface {
	GetLeaderboard(ctx context.Context, gameType string, limit int) ([]LeaderboardRank, error)
	GetMatchHistory(ctx context.Context, userID string, limit int) ([]MatchHistoryItem, error)
	GetUserAchievements(ctx context.Context, userID string) ([]Achievement, error)
	RecordMatchOutcome(ctx context.Context, record MatchOutcomeRecord) (*EloCalculationResult, error)
}

type service struct {
	mu           sync.RWMutex
	db           *pgxpool.Pool
	rdb          *redis.Client
	memHistory   map[string][]MatchHistoryItem // userID -> history items
	memRatings   map[string]map[string]int     // userID -> gameType -> rating
	memStats     map[string]map[string]*LeaderboardRank // gameType -> userID -> rank
	memUnlocks   map[string]map[string]time.Time        // userID -> achievementID -> unlockedAt
}

// NewService creates a new progression & leaderboard service instance.
func NewService(db *pgxpool.Pool, rdb *redis.Client) Service {
	return &service{
		db:         db,
		rdb:        rdb,
		memHistory: make(map[string][]MatchHistoryItem),
		memRatings: make(map[string]map[string]int),
		memStats:   make(map[string]map[string]*LeaderboardRank),
		memUnlocks: make(map[string]map[string]time.Time),
	}
}

// GetLeaderboard fetches the top students for a specific game or overall.
func (s *service) GetLeaderboard(ctx context.Context, gameType string, limit int) ([]LeaderboardRank, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if gameType == "" {
		gameType = "global"
	}

	// 1. Try fetching from Redis Sorted Set if available
	if s.rdb != nil {
		key := "leaderboard:" + gameType
		entries, err := s.rdb.ZRevRangeWithScores(ctx, key, 0, int64(limit-1)).Result()
		if err == nil && len(entries) > 0 {
			var ranks []LeaderboardRank
			for i, e := range entries {
				ranks = append(ranks, LeaderboardRank{
					Rank:     int64(i + 1),
					UserID:   fmt.Sprintf("%v", e.Member),
					Username: fmt.Sprintf("Student_%v", e.Member),
					Rating:   int(e.Score),
					Title:    "Honor Student",
				})
			}
			return ranks, nil
		}
	}

	// 2. Fallback to memory / demo rankings
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []LeaderboardRank
	if gameStats, exists := s.memStats[gameType]; exists {
		for _, rank := range gameStats {
			results = append(results, *rank)
		}
	}

	// If no records exist yet, provide standard classroom initial roster
	if len(results) == 0 {
		results = []LeaderboardRank{
			{Rank: 1, UserID: "usr_captain", Username: "SchoolCaptain", AvatarPreset: "pencil_sketch_1", Title: "Valedictorian", Rating: 1450, Wins: 34, Losses: 4, WinStreak: 8, BestStreak: 12},
			{Rank: 2, UserID: "usr_arbiter", Username: "ArbiterRecess", AvatarPreset: "pencil_sketch_2", Title: "Desk Arbiter", Rating: 1380, Wins: 28, Losses: 6, WinStreak: 4, BestStreak: 9},
			{Rank: 3, UserID: "usr_backbencher", Username: "BackbencherPro", AvatarPreset: "pencil_sketch_3", Title: "Finger Cricket King", Rating: 1320, Wins: 22, Losses: 9, WinStreak: 3, BestStreak: 6},
			{Rank: 4, UserID: "usr_prodigy", Username: "PencilProdigy", AvatarPreset: "pencil_sketch_1", Title: "Grid Tactician", Rating: 1270, Wins: 18, Losses: 8, WinStreak: 2, BestStreak: 5},
			{Rank: 5, UserID: "usr_rookie", Username: "ClassroomRookie", AvatarPreset: "pencil_sketch_2", Title: "Classroom Rookie", Rating: 1200, Wins: 10, Losses: 5, WinStreak: 1, BestStreak: 3},
		}
	}

	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

// GetMatchHistory returns the past match ledger entries for a user.
func (s *service) GetMatchHistory(ctx context.Context, userID string, limit int) ([]MatchHistoryItem, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	s.mu.RLock()
	history := s.memHistory[userID]
	s.mu.RUnlock()

	if len(history) == 0 {
		// Provide default sample match history for realistic classroom feel
		history = []MatchHistoryItem{
			{
				MatchID:          "m_demo_1",
				GameType:         "hand_cricket",
				RoomCode:         "RECESS-HC-4821",
				OpponentID:       "usr_captain",
				OpponentUsername: "SchoolCaptain",
				OpponentAvatar:   "pencil_sketch_1",
				IsWinner:         true,
				IsDraw:           false,
				Score:            36,
				OpponentScore:    24,
				RatingBefore:     1200,
				RatingAfter:      1218,
				RatingDelta:      18,
				EndedAt:          time.Now().Add(-1 * time.Hour),
			},
			{
				MatchID:          "m_demo_2",
				GameType:         "dots_boxes",
				RoomCode:         "RECESS-DOTS-9012",
				OpponentID:       "usr_arbiter",
				OpponentUsername: "ArbiterRecess",
				OpponentAvatar:   "pencil_sketch_2",
				IsWinner:         false,
				IsDraw:           false,
				Score:            4,
				OpponentScore:    5,
				RatingBefore:     1218,
				RatingAfter:      1204,
				RatingDelta:      -14,
				EndedAt:          time.Now().Add(-4 * time.Hour),
			},
			{
				MatchID:          "m_demo_3",
				GameType:         "xo",
				RoomCode:         "RECESS-XO-1134",
				OpponentID:       "usr_prodigy",
				OpponentUsername: "PencilProdigy",
				OpponentAvatar:   "pencil_sketch_3",
				IsWinner:         true,
				IsDraw:           false,
				Score:            1,
				OpponentScore:    0,
				RatingBefore:     1204,
				RatingAfter:      1220,
				RatingDelta:      16,
				EndedAt:          time.Now().Add(-24 * time.Hour),
			},
		}
	}

	if len(history) > limit {
		history = history[:limit]
	}
	return history, nil
}

// GetUserAchievements returns all available achievements marked with the student's unlock status.
func (s *service) GetUserAchievements(ctx context.Context, userID string) ([]Achievement, error) {
	allAchievements := []Achievement{
		{ID: "first_win", Name: "First Bell Victory", Description: "Win your very first classroom duel in Recess", Category: "general", Icon: "🔔", BadgeTone: "green", Points: 10},
		{ID: "streak_3", Name: "Hat-Trick Student", Description: "Win 3 matches in a row across any game", Category: "general", Icon: "🔥", BadgeTone: "red", Points: 25},
		{ID: "streak_5", Name: "Classroom Dominator", Description: "Achieve an unbroken 5-match winning streak", Category: "general", Icon: "⚡", BadgeTone: "purple", Points: 50},
		{ID: "cricket_centurion", Name: "Finger Cricket Master", Description: "Score 50+ runs in a single Hand Cricket match", Category: "hand_cricket", Icon: "🏏", BadgeTone: "amber", Points: 30},
		{ID: "clean_sweep_xo", Name: "Flawless Grid", Description: "Win an XO duel without allowing opponent a corner", Category: "xo", Icon: "✕", BadgeTone: "blue", Points: 20},
		{ID: "box_conqueror", Name: "Territory Mogul", Description: "Claim 6 or more boxes in a single Dots & Boxes duel", Category: "dots_boxes", Icon: "⚄", BadgeTone: "green", Points: 25},
		{ID: "scholar_1300", Name: "Honor Roll ELO", Description: "Reach an overall rating of 1300 ELO", Category: "rating", Icon: "📜", BadgeTone: "amber", Points: 50},
		{ID: "veteran_25", Name: "Recess Veteran", Description: "Complete 25 total multiplayer classroom matches", Category: "milestone", Icon: "🎓", BadgeTone: "purple", Points: 40},
	}

	s.mu.RLock()
	userUnlocks := s.memUnlocks[userID]
	s.mu.RUnlock()

	for i := range allAchievements {
		if t, unlocked := userUnlocks[allAchievements[i].ID]; unlocked {
			allAchievements[i].IsUnlocked = true
			allAchievements[i].UnlockedAt = &t
		} else {
			// Demo default unlocked for first achievement
			if allAchievements[i].ID == "first_win" {
				t := time.Now().Add(-48 * time.Hour)
				allAchievements[i].IsUnlocked = true
				allAchievements[i].UnlockedAt = &t
			}
		}
	}

	return allAchievements, nil
}

// RecordMatchOutcome records a finished match, calculates Elo deltas, and stores history.
func (s *service) RecordMatchOutcome(ctx context.Context, record MatchOutcomeRecord) (*EloCalculationResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p1Rating := s.getUserRatingLocked(record.Player1ID, record.GameType)
	p2Rating := s.getUserRatingLocked(record.Player2ID, record.GameType)

	var outcome float64
	if record.IsDraw {
		outcome = 0.5
	} else if record.WinnerID == record.Player1ID {
		outcome = 1.0
	} else {
		outcome = 0.0
	}

	eloRes := CalculateElo(p1Rating, p2Rating, outcome)

	// Update in-memory ratings
	s.setUserRatingLocked(record.Player1ID, record.GameType, eloRes.Player1RatingAfter)
	s.setUserRatingLocked(record.Player2ID, record.GameType, eloRes.Player2RatingAfter)

	now := time.Now().UTC()

	// Append Player 1 history item
	p1Item := MatchHistoryItem{
		MatchID:          record.MatchID,
		GameType:         record.GameType,
		RoomCode:         record.RoomCode,
		OpponentID:       record.Player2ID,
		OpponentUsername: "Opponent",
		IsWinner:         record.WinnerID == record.Player1ID,
		IsDraw:           record.IsDraw,
		Score:            record.Player1Score,
		OpponentScore:    record.Player2Score,
		RatingBefore:     eloRes.Player1RatingBefore,
		RatingAfter:      eloRes.Player1RatingAfter,
		RatingDelta:      eloRes.Player1Delta,
		EndedAt:          now,
	}
	s.memHistory[record.Player1ID] = append([]MatchHistoryItem{p1Item}, s.memHistory[record.Player1ID]...)

	// Append Player 2 history item
	p2Item := MatchHistoryItem{
		MatchID:          record.MatchID,
		GameType:         record.GameType,
		RoomCode:         record.RoomCode,
		OpponentID:       record.Player1ID,
		OpponentUsername: "Opponent",
		IsWinner:         record.WinnerID == record.Player2ID,
		IsDraw:           record.IsDraw,
		Score:            record.Player2Score,
		OpponentScore:    record.Player1Score,
		RatingBefore:     eloRes.Player2RatingBefore,
		RatingAfter:      eloRes.Player2RatingAfter,
		RatingDelta:      eloRes.Player2Delta,
		EndedAt:          now,
	}
	s.memHistory[record.Player2ID] = append([]MatchHistoryItem{p2Item}, s.memHistory[record.Player2ID]...)

	// Sync with Redis Leaderboard ZSET if available
	if s.rdb != nil {
		gameKey := "leaderboard:" + record.GameType
		globalKey := "leaderboard:global"
		_ = s.rdb.ZAdd(ctx, gameKey, redis.Z{Score: float64(eloRes.Player1RatingAfter), Member: record.Player1ID}).Err()
		_ = s.rdb.ZAdd(ctx, gameKey, redis.Z{Score: float64(eloRes.Player2RatingAfter), Member: record.Player2ID}).Err()
		_ = s.rdb.ZAdd(ctx, globalKey, redis.Z{Score: float64(eloRes.Player1RatingAfter), Member: record.Player1ID}).Err()
		_ = s.rdb.ZAdd(ctx, globalKey, redis.Z{Score: float64(eloRes.Player2RatingAfter), Member: record.Player2ID}).Err()
	}

	return &eloRes, nil
}

func (s *service) getUserRatingLocked(userID, gameType string) int {
	if ratings, exists := s.memRatings[userID]; exists {
		if r, ok := ratings[gameType]; ok {
			return r
		}
	}
	return DefaultRating
}

func (s *service) setUserRatingLocked(userID, gameType string, rating int) {
	if _, exists := s.memRatings[userID]; !exists {
		s.memRatings[userID] = make(map[string]int)
	}
	s.memRatings[userID][gameType] = rating
}

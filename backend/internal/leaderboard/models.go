package leaderboard

import "time"

// LeaderboardRank represents a ranked player on the classroom scoreboard.
type LeaderboardRank struct {
	Rank         int64     `json:"rank"`
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	AvatarPreset string    `json:"avatar_preset"`
	Title        string    `json:"title"`
	Rating       int       `json:"rating"`
	Wins         int       `json:"wins"`
	Losses       int       `json:"losses"`
	WinStreak    int       `json:"win_streak"`
	BestStreak   int       `json:"best_streak"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// MatchHistoryItem represents a past duel on a student's match ledger.
type MatchHistoryItem struct {
	MatchID          string    `json:"match_id"`
	GameType         string    `json:"game_type"`
	RoomCode         string    `json:"room_code"`
	OpponentID       string    `json:"opponent_id"`
	OpponentUsername string    `json:"opponent_username"`
	OpponentAvatar   string    `json:"opponent_avatar"`
	IsWinner         bool      `json:"is_winner"`
	IsDraw           bool      `json:"is_draw"`
	Score            int       `json:"score"`
	OpponentScore    int       `json:"opponent_score"`
	RatingBefore     int       `json:"rating_before"`
	RatingAfter      int       `json:"rating_after"`
	RatingDelta      int       `json:"rating_delta"`
	EndedAt          time.Time `json:"ended_at"`
}

// Achievement represents a classroom award/stamp.
type Achievement struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	Icon        string     `json:"icon"`
	BadgeTone   string     `json:"badge_tone"`
	Points      int        `json:"points"`
	IsUnlocked  bool       `json:"is_unlocked"`
	UnlockedAt  *time.Time `json:"unlocked_at,omitempty"`
}

// EloCalculationResult holds post-match rating updates.
type EloCalculationResult struct {
	Player1RatingBefore int `json:"player1_rating_before"`
	Player1RatingAfter  int `json:"player1_rating_after"`
	Player1Delta        int `json:"player1_delta"`
	Player2RatingBefore int `json:"player2_rating_before"`
	Player2RatingAfter  int `json:"player2_rating_after"`
	Player2Delta        int `json:"player2_delta"`
}

// MatchOutcomeRecord is used to record a finished match.
type MatchOutcomeRecord struct {
	MatchID       string         `json:"match_id"`
	RoomCode      string         `json:"room_code"`
	GameType      string         `json:"game_type"`
	Player1ID     string         `json:"player1_id"`
	Player2ID     string         `json:"player2_id"`
	WinnerID      string         `json:"winner_id,omitempty"`
	IsDraw        bool           `json:"is_draw"`
	Player1Score  int            `json:"player1_score"`
	Player2Score  int            `json:"player2_score"`
	SummaryData   map[string]any `json:"summary_data,omitempty"`
	DurationSecs  int            `json:"duration_seconds"`
}

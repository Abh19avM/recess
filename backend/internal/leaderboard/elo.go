package leaderboard

import (
	"math"
)

const (
	DefaultKFactor = 32.0
	DefaultRating  = 1200
)

// CalculateElo computes the updated Elo ratings for two players following a match.
// outcome: 1.0 if Player 1 won, 0.0 if Player 2 won, 0.5 for a draw.
func CalculateElo(p1Rating, p2Rating int, outcome float64, customK ...float64) EloCalculationResult {
	k := DefaultKFactor
	if len(customK) > 0 && customK[0] > 0 {
		k = customK[0]
	}

	// 1. Calculate Expected Scores
	// E_A = 1 / (1 + 10 ^ ((R_B - R_A) / 400))
	diff1 := float64(p2Rating-p1Rating) / 400.0
	expected1 := 1.0 / (1.0 + math.Pow(10.0, diff1))

	diff2 := float64(p1Rating-p2Rating) / 400.0
	expected2 := 1.0 / (1.0 + math.Pow(10.0, diff2))

	// 2. Actual Scores
	actual1 := outcome
	actual2 := 1.0 - outcome

	// 3. Compute Rating Deltas
	delta1 := int(math.Round(k * (actual1 - expected1)))
	delta2 := int(math.Round(k * (actual2 - expected2)))

	newRating1 := p1Rating + delta1
	newRating2 := p2Rating + delta2

	// Enforce floor rating of 100
	if newRating1 < 100 {
		newRating1 = 100
	}
	if newRating2 < 100 {
		newRating2 = 100
	}

	return EloCalculationResult{
		Player1RatingBefore: p1Rating,
		Player1RatingAfter:  newRating1,
		Player1Delta:        delta1,
		Player2RatingBefore: p2Rating,
		Player2RatingAfter:  newRating2,
		Player2Delta:        delta2,
	}
}

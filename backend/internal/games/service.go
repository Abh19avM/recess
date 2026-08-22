package games

import (
	"context"
	"errors"
)

var (
	// ErrGameNotFound indicates that the requested game type is not supported.
	ErrGameNotFound = errors.New("game not found")
)

// Service defines operations for retrieving game catalog and metadata.
type Service interface {
	ListGames(ctx context.Context) ([]GameMetadata, error)
	GetGameByType(ctx context.Context, gType GameType) (*GameMetadata, error)
}

type service struct {
	catalog map[GameType]GameMetadata
}

// NewService creates a new games service initialized with all 6 classic school games.
func NewService() Service {
	catalog := map[GameType]GameMetadata{
		GameHandCricket: {
			Type:              GameHandCricket,
			Name:              "Hand Cricket",
			Tagline:           "The legendary school desk cricket match",
			Description:       "Toss for odd/even, pick 1-6 fingers simultaneously. Score runs until your choices match — then you're OUT!",
			MinPlayers:        2,
			MaxPlayers:        2,
			EstimatedDuration: "3-5 mins",
			RulesSummary: []string{
				"Odd/Even toss determines who bats or bowls first",
				"Both players reveal a number from 1 to 6 simultaneously",
				"If batsman and bowler pick the same number: OUT!",
				"If numbers differ: batsman scores the runs of their choice",
				"Chase the target in the second innings to win",
			},
			BoardSurface: "notebook_ruled",
		},
		GameDotsBoxes: {
			Type:              GameDotsBoxes,
			Name:              "Dots & Boxes",
			Tagline:           "Tactical grid warfare in the math notebook",
			Description:       "Take turns connecting adjacent dots with lines. Complete a 1x1 box to capture it and earn an extra turn!",
			MinPlayers:        2,
			MaxPlayers:        2,
			EstimatedDuration: "5-8 mins",
			RulesSummary: []string{
				"Connect two adjacent dots horizontally or vertically",
				"Closing the 4th wall of a 1x1 box captures it for 1 point",
				"Capturing a box awards an immediate bonus turn",
				"Player with the most captured boxes at the end wins",
			},
			BoardSurface: "graph_paper",
		},
		GameXO: {
			Type:              GameXO,
			Name:              "XO / Tic-Tac-Toe",
			Tagline:           "Classic pencil duel in the notebook margin",
			Description:       "The timeless battle of Xs and Os across 3x3, 4x4, and 5x5 grids with rapid turn timers.",
			MinPlayers:        2,
			MaxPlayers:        2,
			EstimatedDuration: "1-2 mins",
			RulesSummary: []string{
				"Take turns placing your X or O mark on the grid",
				"First to form a continuous line (horizontal, vertical, diagonal) wins",
				"Supports classic 3x3 (3 in a row) and tactical 4x4 / 5x5 (4 in a row)",
			},
			BoardSurface: "notebook_margin",
		},
		GameConnect4: {
			Type:              GameConnect4,
			Name:              "Connect 4",
			Tagline:           "Gravity-defying strategy on the wooden classroom desk",
			Description:       "Drop colored tokens into the 7x6 vertical grid. Connect 4 tokens in a row before your opponent does!",
			MinPlayers:        2,
			MaxPlayers:        2,
			EstimatedDuration: "3-5 mins",
			RulesSummary: []string{
				"Drop your chip into any of the 7 columns",
				"Chips fall to the lowest available space in the chosen column",
				"First to align 4 chips in any direction wins",
			},
			BoardSurface: "desk_wood",
		},
		GamePaperFootball: {
			Type:              GamePaperFootball,
			Name:              "Paper Football",
			Tagline:           "Physics, friction, and desk-edge touchdowns",
			Description:       "Flick the triangular folded paper football across the desk. Stop it hanging over the table edge for a 6-point Touchdown!",
			MinPlayers:        2,
			MaxPlayers:        2,
			EstimatedDuration: "4-6 mins",
			RulesSummary: []string{
				"Drag and flick the paper football across the desk surface",
				"If the football stops hanging over the opponent's desk edge: TOUCHDOWN! (6 pts)",
				"Touchdown unlocks a Field Goal kick attempt (1 pt) through finger posts",
				"Falling off the table results in an Out of Bounds turnover",
			},
			BoardSurface: "varnished_desk",
		},
		GameNPAT: {
			Type:              GameNPAT,
			Name:              "Name–Place–Animal–Thing",
			Tagline:           "Fast-paced vocabulary rush against the classroom clock",
			Description:       "A random letter is spun. Write down a Name, Place, Animal, and Thing starting with that letter. First done hits STOP!",
			MinPlayers:        2,
			MaxPlayers:        8,
			EstimatedDuration: "5-10 mins",
			RulesSummary: []string{
				"A random letter is chosen for each round",
				"Fill valid words starting with the round letter across all 4 categories",
				"First player done slams the STOP buzzer, triggering a 10s countdown for others",
				"Unique valid answers score 10 pts, matching answers score 5 pts, invalid scores 0",
			},
			BoardSurface: "exam_sheet",
		},
	}

	return &service{catalog: catalog}
}

func (s *service) ListGames(ctx context.Context) ([]GameMetadata, error) {
	list := make([]GameMetadata, 0, len(s.catalog))
	for _, g := range s.catalog {
		list = append(list, g)
	}
	return list, nil
}

func (s *service) GetGameByType(ctx context.Context, gType GameType) (*GameMetadata, error) {
	g, ok := s.catalog[gType]
	if !ok {
		return nil, ErrGameNotFound
	}
	return &g, nil
}

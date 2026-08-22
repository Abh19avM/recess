package games

// GameType represents the identifier for one of the 6 school-time games.
type GameType string

const (
	GameHandCricket   GameType = "hand_cricket"
	GameDotsBoxes     GameType = "dots_boxes"
	GameXO            GameType = "xo"
	GameConnect4      GameType = "connect4"
	GamePaperFootball GameType = "paper_football"
	GameNPAT          GameType = "npat"
)

// GameMetadata provides information about each school game.
type GameMetadata struct {
	Type              GameType `json:"type"`
	Name              string   `json:"name"`
	Tagline           string   `json:"tagline"`
	Description       string   `json:"description"`
	MinPlayers        int      `json:"min_players"`
	MaxPlayers        int      `json:"max_players"`
	EstimatedDuration string   `json:"estimated_duration"`
	RulesSummary      []string `json:"rules_summary"`
	BoardSurface      string   `json:"board_surface"`
}

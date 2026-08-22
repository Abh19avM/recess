package engine

import (
	"encoding/json"
	"errors"
	"time"
)

// GameType represents the unique identifier for a game.
type GameType string

const (
	GameTypeXO            GameType = "xo"
	GameTypeHandCricket   GameType = "hand_cricket"
	GameTypeDotsBoxes     GameType = "dots_boxes"
	GameTypeConnect4      GameType = "connect4"
	GameTypePaperFootball GameType = "paper_football"
	GameTypeNPAT          GameType = "npat"
)

// GameStatus represents the lifecycle state of a match.
type GameStatus string

const (
	StatusWaiting   GameStatus = "waiting"
	StatusActive    GameStatus = "active"
	StatusFinished  GameStatus = "finished"
	StatusAbandoned GameStatus = "abandoned"
)

// Standard Game Engine Errors
var (
	ErrGameNotInitialized = errors.New("game has not been initialized")
	ErrGameAlreadyActive  = errors.New("game is already active")
	ErrGameFinished       = errors.New("game is already finished")
	ErrGameNotActive      = errors.New("game is not currently active")
	ErrInvalidPlayerCount = errors.New("invalid number of players for this game")
	ErrNotPlayerTurn      = errors.New("it is not this player's turn")
	ErrPlayerNotInGame    = errors.New("player is not a participant in this game")
	ErrInvalidMove        = errors.New("invalid move")
	ErrInvalidCoordinates = errors.New("coordinates are out of bounds")
	ErrCellOccupied       = errors.New("cell is already occupied")
	ErrGameNotFound       = errors.New("game type not registered")
	ErrGameAlreadyExists  = errors.New("game type is already registered")
)

// Player represents a participant in a game.
type Player struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	SeatNumber int    `json:"seat_number"`
	IsBot      bool   `json:"is_bot"`
}

// Move represents a player's action submitted to the engine.
type Move struct {
	PlayerID  string          `json:"player_id"`
	Action    string          `json:"action"`
	Data      json.RawMessage `json:"data,omitempty"`
	Timestamp int64           `json:"timestamp"`
}

// GameResult represents the final outcome of a completed game.
type GameResult struct {
	WinnerID     string         `json:"winner_id,omitempty"` // Empty string if draw
	IsDraw       bool           `json:"is_draw"`
	Scores       map[string]int `json:"scores,omitempty"`        // PlayerID -> score
	PlayerPlaces map[string]int `json:"player_places,omitempty"` // PlayerID -> rank (1st, 2nd, etc.)
	Reason       string         `json:"reason"`                  // e.g., "normal", "timeout", "resignation"
	CompletedAt  time.Time      `json:"completed_at"`
}

// GameState is the authoritative snapshot of a game at any point in time.
type GameState struct {
	GameID       string          `json:"game_id"`
	GameType     GameType        `json:"game_type"`
	Status       GameStatus      `json:"status"`
	CurrentTurn  string          `json:"current_turn,omitempty"` // PlayerID whose turn it is
	TurnDeadline *time.Time      `json:"turn_deadline,omitempty"`
	MoveCount    int             `json:"move_count"`
	BoardState   json.RawMessage `json:"board_state"`
	Result       *GameResult     `json:"result,omitempty"`
	Version      int64           `json:"version"` // Incremented sequentially on every state change
}

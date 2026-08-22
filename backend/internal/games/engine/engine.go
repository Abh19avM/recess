package engine

import "encoding/json"

// Engine defines the standard contract for all authoritative game rule engines.
// All engine implementations must be deterministic, pure, and independent of transport layers.
type Engine interface {
	// Initialize sets up the game with participants, optional config, and returns the initial state.
	Initialize(gameID string, players []Player, config json.RawMessage) (*GameState, error)

	// ValidateMove checks whether a move is legal without mutating the current state.
	ValidateMove(move Move) error

	// ApplyMove validates and executes the move, advancing the state and returning the new snapshot.
	ApplyMove(move Move) (*GameState, error)

	// State returns the current authoritative snapshot of the game.
	State() *GameState

	// IsFinished returns whether the game has reached a terminal outcome.
	IsFinished() bool

	// Result returns the final outcome of the game (or nil if still active).
	Result() *GameResult

	// NextTurn returns the PlayerID of the player whose turn it is, or empty string if simultaneous/terminal.
	NextTurn() string

	// GameType returns the unique identifier for this game.
	GameType() GameType
}

// Factory is a constructor function that instantiates a new clean Engine.
type Factory func() Engine

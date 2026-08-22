package xo

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Abh19avM/recess/internal/games/engine"
)

func init() {
	engine.MustRegister(engine.GameTypeXO, NewXOEngine)
}

// MoveData represents the coordinate parameters of an XO move.
type MoveData struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

// BoardState represents the internal snapshot of the Tic-Tac-Toe grid.
type BoardState struct {
	Grid          [3][3]string   `json:"grid"`
	PlayerSymbols map[string]string `json:"player_symbols"` // PlayerID -> "X" or "O"
	WinningLine   [][]int        `json:"winning_line,omitempty"` // Coordinates of the winning 3 cells
}

// Engine implements the engine.Engine contract for Tic-Tac-Toe.
type Engine struct {
	mu            sync.RWMutex
	gameID        string
	status        engine.GameStatus
	players       []engine.Player
	playerSymbols map[string]string // PlayerID -> "X" or "O"
	symbolPlayers map[string]string // "X" -> PlayerID
	currentTurn   string            // PlayerID
	moveCount     int
	version       int64
	grid          [3][3]string
	winningLine   [][]int
	result        *engine.GameResult
}

// NewXOEngine creates a fresh XO engine instance.
func NewXOEngine() engine.Engine {
	return &Engine{
		status:        engine.StatusWaiting,
		playerSymbols: make(map[string]string),
		symbolPlayers: make(map[string]string),
	}
}

func (e *Engine) GameType() engine.GameType {
	return engine.GameTypeXO
}

// Initialize configures the match with exactly 2 players.
func (e *Engine) Initialize(gameID string, players []engine.Player, config json.RawMessage) (*engine.GameState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(players) != 2 {
		return nil, fmt.Errorf("%w: XO requires exactly 2 players, got %d", engine.ErrInvalidPlayerCount, len(players))
	}

	e.gameID = gameID
	e.players = players
	e.status = engine.StatusActive
	e.grid = [3][3]string{}
	e.winningLine = nil
	e.result = nil
	e.moveCount = 0
	e.version = 1

	// Player 0 plays 'X' and has first move; Player 1 plays 'O'
	e.playerSymbols = map[string]string{
		players[0].ID: "X",
		players[1].ID: "O",
	}
	e.symbolPlayers = map[string]string{
		"X": players[0].ID,
		"O": players[1].ID,
	}
	e.currentTurn = players[0].ID

	return e.buildStateLocked(), nil
}

// ValidateMove checks whether a coordinate move is legal.
func (e *Engine) ValidateMove(move engine.Move) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.validateMoveLocked(move)
}

func (e *Engine) validateMoveLocked(move engine.Move) error {
	if e.status != engine.StatusActive {
		if e.status == engine.StatusFinished {
			return engine.ErrGameFinished
		}
		return engine.ErrGameNotActive
	}

	if move.PlayerID != e.currentTurn {
		return engine.ErrNotPlayerTurn
	}

	var data MoveData
	if len(move.Data) == 0 {
		return errors.New("missing move data")
	}
	if err := json.Unmarshal(move.Data, &data); err != nil {
		return fmt.Errorf("%w: invalid move data JSON: %v", engine.ErrInvalidMove, err)
	}

	if data.Row < 0 || data.Row > 2 || data.Col < 0 || data.Col > 2 {
		return engine.ErrInvalidCoordinates
	}

	if e.grid[data.Row][data.Col] != "" {
		return engine.ErrCellOccupied
	}

	return nil
}

// ApplyMove validates and executes the coordinate move.
func (e *Engine) ApplyMove(move engine.Move) (*engine.GameState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := e.validateMoveLocked(move); err != nil {
		return nil, err
	}

	var data MoveData
	_ = json.Unmarshal(move.Data, &data)

	symbol := e.playerSymbols[move.PlayerID]
	e.grid[data.Row][data.Col] = symbol
	e.moveCount++
	e.version++

	// Check if this move wins the game
	if line, won := e.checkWin(symbol); won {
		e.status = engine.StatusFinished
		e.winningLine = line
		otherPlayerID := e.getOtherPlayerID(move.PlayerID)

		e.result = &engine.GameResult{
			WinnerID: move.PlayerID,
			IsDraw:   false,
			Scores: map[string]int{
				move.PlayerID: 1,
				otherPlayerID: 0,
			},
			PlayerPlaces: map[string]int{
				move.PlayerID: 1,
				otherPlayerID: 2,
			},
			Reason:      "normal",
			CompletedAt: time.Now().UTC(),
		}
		e.currentTurn = ""
		return e.buildStateLocked(), nil
	}

	// Check for Draw (all 9 squares filled)
	if e.moveCount >= 9 {
		e.status = engine.StatusFinished
		e.result = &engine.GameResult{
			WinnerID: "",
			IsDraw:   true,
			Scores: map[string]int{
				e.players[0].ID: 0,
				e.players[1].ID: 0,
			},
			PlayerPlaces: map[string]int{
				e.players[0].ID: 1,
				e.players[1].ID: 1,
			},
			Reason:      "draw",
			CompletedAt: time.Now().UTC(),
		}
		e.currentTurn = ""
		return e.buildStateLocked(), nil
	}

	// Switch turn to the other player
	e.currentTurn = e.getOtherPlayerID(move.PlayerID)

	return e.buildStateLocked(), nil
}

func (e *Engine) State() *engine.GameState {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.buildStateLocked()
}

func (e *Engine) IsFinished() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.status == engine.StatusFinished
}

func (e *Engine) Result() *engine.GameResult {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.result
}

func (e *Engine) NextTurn() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.currentTurn
}

func (e *Engine) getOtherPlayerID(playerID string) string {
	for _, p := range e.players {
		if p.ID != playerID {
			return p.ID
		}
	}
	return ""
}

func (e *Engine) checkWin(symbol string) ([][]int, bool) {
	// 1. Check Rows
	for r := 0; r < 3; r++ {
		if e.grid[r][0] == symbol && e.grid[r][1] == symbol && e.grid[r][2] == symbol {
			return [][]int{{r, 0}, {r, 1}, {r, 2}}, true
		}
	}

	// 2. Check Columns
	for c := 0; c < 3; c++ {
		if e.grid[0][c] == symbol && e.grid[1][c] == symbol && e.grid[2][c] == symbol {
			return [][]int{{0, c}, {1, c}, {2, c}}, true
		}
	}

	// 3. Check Diagonals
	if e.grid[0][0] == symbol && e.grid[1][1] == symbol && e.grid[2][2] == symbol {
		return [][]int{{0, 0}, {1, 1}, {2, 2}}, true
	}
	if e.grid[0][2] == symbol && e.grid[1][1] == symbol && e.grid[2][0] == symbol {
		return [][]int{{0, 2}, {1, 1}, {2, 0}}, true
	}

	return nil, false
}

func (e *Engine) buildStateLocked() *engine.GameState {
	boardState := BoardState{
		Grid:          e.grid,
		PlayerSymbols: e.playerSymbols,
		WinningLine:   e.winningLine,
	}

	rawBoard, _ := json.Marshal(boardState)

	return &engine.GameState{
		GameID:      e.gameID,
		GameType:    engine.GameTypeXO,
		Status:      e.status,
		CurrentTurn: e.currentTurn,
		MoveCount:   e.moveCount,
		BoardState:  rawBoard,
		Result:      e.result,
		Version:     e.version,
	}
}

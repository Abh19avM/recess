package connect4

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Abh19avM/recess/internal/games/engine"
)

const (
	Rows = 6
	Cols = 7
)

func init() {
	engine.Register(engine.GameTypeConnect4, func() engine.Engine {
		return NewConnect4Engine()
	})
}

// CellPos represents a row-col position on the grid.
type CellPos struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

// DropMove represents a disc drop move.
type DropMove struct {
	Col int `json:"col"`
}

// BoardState represents the serializable state of the Connect 4 grid.
type BoardState struct {
	Rows         int              `json:"rows"`
	Cols         int              `json:"cols"`
	Grid         [Rows][Cols]string `json:"grid"` // UserID or empty string
	PlayerColors map[string]string `json:"player_colors"` // UserID -> "navy" or "red"
	LastMove     *CellPos         `json:"last_move,omitempty"`
	WinningCells []CellPos        `json:"winning_cells,omitempty"`
}

// Engine implements the Connect 4 server-authoritative engine.
type Engine struct {
	mu           sync.RWMutex
	gameID       string
	players      []engine.Player
	status       engine.GameStatus
	grid         [Rows][Cols]string
	playerColors map[string]string
	currentTurn  string
	lastMove     *CellPos
	winningCells []CellPos
	moveCount    int
	version      int64
	result       *engine.GameResult
}

// NewConnect4Engine creates a fresh Connect 4 engine.
func NewConnect4Engine() engine.Engine {
	return &Engine{
		status:       engine.StatusWaiting,
		playerColors: make(map[string]string),
	}
}

func (e *Engine) GameType() engine.GameType {
	return engine.GameTypeConnect4
}

// Initialize configures the match with 2 players.
func (e *Engine) Initialize(gameID string, players []engine.Player, config json.RawMessage) (*engine.GameState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(players) != 2 {
		return nil, fmt.Errorf("%w: Connect 4 requires exactly 2 players, got %d", engine.ErrInvalidPlayerCount, len(players))
	}

	e.gameID = gameID
	e.players = players
	e.status = engine.StatusActive
	e.currentTurn = players[0].ID
	e.grid = [Rows][Cols]string{}
	e.playerColors = map[string]string{
		players[0].ID: "navy",
		players[1].ID: "red",
	}
	e.lastMove = nil
	e.winningCells = nil
	e.moveCount = 0
	e.version = 1
	e.result = nil

	return e.stateLocked(), nil
}

// ValidateMove checks whether the requested move is legal.
func (e *Engine) ValidateMove(move engine.Move) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.status != engine.StatusActive {
		return engine.ErrGameFinished
	}
	if move.PlayerID != e.currentTurn {
		return engine.ErrNotPlayerTurn
	}

	var drop DropMove
	if err := json.Unmarshal(move.Data, &drop); err != nil {
		return fmt.Errorf("%w: invalid move data payload", engine.ErrInvalidMove)
	}

	if drop.Col < 0 || drop.Col >= Cols {
		return fmt.Errorf("%w: column %d out of bounds (0-%d)", engine.ErrInvalidCoordinates, drop.Col, Cols-1)
	}

	// Check if column is already full (top cell row 0 is occupied)
	if e.grid[0][drop.Col] != "" {
		return fmt.Errorf("%w: column %d is full", engine.ErrCellOccupied, drop.Col)
	}

	return nil
}

// ApplyMove validates and executes a disc drop move.
func (e *Engine) ApplyMove(move engine.Move) (*engine.GameState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status != engine.StatusActive {
		return nil, engine.ErrGameFinished
	}
	if move.PlayerID != e.currentTurn {
		return nil, engine.ErrNotPlayerTurn
	}

	var drop DropMove
	if err := json.Unmarshal(move.Data, &drop); err != nil {
		return nil, fmt.Errorf("%w: invalid move data payload", engine.ErrInvalidMove)
	}

	if drop.Col < 0 || drop.Col >= Cols {
		return nil, fmt.Errorf("%w: column %d out of bounds", engine.ErrInvalidCoordinates, drop.Col)
	}
	if e.grid[0][drop.Col] != "" {
		return nil, fmt.Errorf("%w: column %d is full", engine.ErrCellOccupied, drop.Col)
	}

	// Find the lowest available row in the selected column
	targetRow := -1
	for r := Rows - 1; r >= 0; r-- {
		if e.grid[r][drop.Col] == "" {
			targetRow = r
			break
		}
	}
	if targetRow == -1 {
		return nil, fmt.Errorf("%w: column is full", engine.ErrCellOccupied)
	}

	// Drop disc
	e.grid[targetRow][drop.Col] = move.PlayerID
	e.lastMove = &CellPos{Row: targetRow, Col: drop.Col}
	e.moveCount++
	e.version++

	// Check 4-in-a-row win condition
	if winCells := e.checkWin(targetRow, drop.Col, move.PlayerID); len(winCells) >= 4 {
		e.winningCells = winCells
		e.status = engine.StatusFinished
		e.result = &engine.GameResult{
			WinnerID: move.PlayerID,
			IsDraw:   false,
			Scores: map[string]int{
				move.PlayerID: 1,
			},
			Reason: fmt.Sprintf("Player %s connected 4 discs in a row!", move.PlayerID),
		}
		return e.stateLocked(), nil
	}

	// Check draw (all 42 cells filled)
	if e.moveCount >= Rows*Cols {
		e.status = engine.StatusFinished
		e.result = &engine.GameResult{
			WinnerID: "",
			IsDraw:   true,
			Scores:   map[string]int{},
			Reason:   "Board is full without 4 in a row. Match ended in a draw.",
		}
		return e.stateLocked(), nil
	}

	// Switch turn to opponent
	if e.currentTurn == e.players[0].ID {
		e.currentTurn = e.players[1].ID
	} else {
		e.currentTurn = e.players[0].ID
	}

	return e.stateLocked(), nil
}

// State returns the current snapshot of the game.
func (e *Engine) State() *engine.GameState {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stateLocked()
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
	if e.status != engine.StatusActive {
		return ""
	}
	return e.currentTurn
}

func (e *Engine) stateLocked() *engine.GameState {
	board := BoardState{
		Rows:         Rows,
		Cols:         Cols,
		Grid:         e.grid,
		PlayerColors: e.playerColors,
		LastMove:     e.lastMove,
		WinningCells: e.winningCells,
	}

	boardData, _ := json.Marshal(board)

	return &engine.GameState{
		GameID:      e.gameID,
		GameType:    engine.GameTypeConnect4,
		Status:      e.status,
		CurrentTurn: e.currentTurn,
		MoveCount:   e.moveCount,
		BoardState:  boardData,
		Result:      e.result,
		Version:     e.version,
	}
}

// checkWin checks horizontal, vertical, and both diagonals from (row, col)
func (e *Engine) checkWin(r, c int, playerID string) []CellPos {
	directions := [][2]int{
		{0, 1},  // Horizontal (-)
		{1, 0},  // Vertical (|)
		{1, 1},  // Diagonal down-right (\)
		{1, -1}, // Diagonal down-left (/)
	}

	for _, d := range directions {
		dr, dc := d[0], d[1]
		cells := []CellPos{{Row: r, Col: c}}

		// Scan in forward direction
		step := 1
		for {
			nr, nc := r+dr*step, c+dc*step
			if nr < 0 || nr >= Rows || nc < 0 || nc >= Cols || e.grid[nr][nc] != playerID {
				break
			}
			cells = append(cells, CellPos{Row: nr, Col: nc})
			step++
		}

		// Scan in reverse direction
		step = 1
		for {
			nr, nc := r-dr*step, c-dc*step
			if nr < 0 || nr >= Rows || nc < 0 || nc >= Cols || e.grid[nr][nc] != playerID {
				break
			}
			cells = append([]CellPos{{Row: nr, Col: nc}}, cells...)
			step++
		}

		if len(cells) >= 4 {
			return cells
		}
	}

	return nil
}

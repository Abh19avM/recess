package dotsboxes

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Abh19avM/recess/internal/games/engine"
)

func init() {
	engine.MustRegister(engine.GameTypeDotsBoxes, NewDotsBoxesEngine)
}

// EdgeKey uniquely identifies a line between two adjacent dots.
type EdgeKey struct {
	Type string `json:"type"` // "h" for horizontal, "v" for vertical
	Row  int    `json:"row"`
	Col  int    `json:"col"`
}

func (e EdgeKey) String() string {
	return fmt.Sprintf("%s_%d_%d", e.Type, e.Row, e.Col)
}

// BoxState represents a single square cell on the board.
type BoxState struct {
	Row       int    `json:"row"`
	Col       int    `json:"col"`
	OwnerID   string `json:"owner_id,omitempty"` // PlayerID of student who claimed the box
	ClaimedAt int64  `json:"claimed_at,omitempty"`
}

// BoardState represents the authoritative Dots & Boxes board.
type BoardState struct {
	Rows            int               `json:"rows"` // Box rows (default 3)
	Cols            int               `json:"cols"` // Box cols (default 3)
	HorizontalEdges [][]string        `json:"horizontal_edges"` // [Rows+1][Cols] PlayerID or ""
	VerticalEdges   [][]string        `json:"vertical_edges"`   // [Rows][Cols+1] PlayerID or ""
	Boxes           [][]string        `json:"boxes"`            // [Rows][Cols] PlayerID or ""
	PlayerInitials  map[string]string `json:"player_initials"`  // PlayerID -> Initial (e.g. "A", "B")
	PlayerColors    map[string]string `json:"player_colors"`    // PlayerID -> "navy" | "red"
	Scores          map[string]int    `json:"scores"`           // PlayerID -> number of boxes claimed
	LastEdge        *EdgeKey          `json:"last_edge,omitempty"`
	CompletedBoxes  int               `json:"completed_boxes"`
	TotalBoxes      int               `json:"total_boxes"`
}

// Engine implements the engine.Engine contract for Dots & Boxes.
type Engine struct {
	mu             sync.RWMutex
	gameID         string
	status         engine.GameStatus
	players        []engine.Player
	rows           int
	cols           int
	horizontal     [][]string // [Rows+1][Cols]
	vertical       [][]string // [Rows][Cols+1]
	boxes          [][]string // [Rows][Cols]
	scores         map[string]int
	playerInitials map[string]string
	playerColors   map[string]string
	currentTurn    string
	lastEdge       *EdgeKey
	moveCount      int
	version        int64
	result         *engine.GameResult
}

// NewDotsBoxesEngine constructs a fresh Dots & Boxes engine.
func NewDotsBoxesEngine() engine.Engine {
	return &Engine{
		status:         engine.StatusWaiting,
		rows:           3,
		cols:           3,
		scores:         make(map[string]int),
		playerInitials: make(map[string]string),
		playerColors:   make(map[string]string),
	}
}

func (e *Engine) GameType() engine.GameType {
	return engine.GameTypeDotsBoxes
}

// Initialize configures the match with 2 players and optional grid dimensions.
func (e *Engine) Initialize(gameID string, players []engine.Player, config json.RawMessage) (*engine.GameState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(players) != 2 {
		return nil, fmt.Errorf("%w: Dots & Boxes requires exactly 2 players, got %d", engine.ErrInvalidPlayerCount, len(players))
	}

	e.rows = 3
	e.cols = 3

	if len(config) > 0 {
		var cfg struct {
			Rows int `json:"rows"`
			Cols int `json:"cols"`
		}
		if err := json.Unmarshal(config, &cfg); err == nil {
			if cfg.Rows >= 1 && cfg.Rows <= 6 {
				e.rows = cfg.Rows
			}
			if cfg.Cols >= 1 && cfg.Cols <= 6 {
				e.cols = cfg.Cols
			}
		}
	}

	e.gameID = gameID
	e.players = players
	e.status = engine.StatusActive
	e.currentTurn = players[0].ID
	e.moveCount = 0
	e.version = 1
	e.result = nil
	e.lastEdge = nil

	// Initialize edge matrices
	e.horizontal = make([][]string, e.rows+1)
	for r := range e.horizontal {
		e.horizontal[r] = make([]string, e.cols)
	}

	e.vertical = make([][]string, e.rows)
	for r := range e.vertical {
		e.vertical[r] = make([]string, e.cols+1)
	}

	e.boxes = make([][]string, e.rows)
	for r := range e.boxes {
		e.boxes[r] = make([]string, e.cols)
	}

	// Player initials & colors
	e.scores = map[string]int{
		players[0].ID: 0,
		players[1].ID: 0,
	}

	p1Initial := "A"
	if len(players[0].Username) > 0 {
		p1Initial = string(players[0].Username[0])
	}
	p2Initial := "B"
	if len(players[1].Username) > 0 {
		p2Initial = string(players[1].Username[0])
	}
	if p1Initial == p2Initial {
		p2Initial = "2"
	}

	e.playerInitials = map[string]string{
		players[0].ID: p1Initial,
		players[1].ID: p2Initial,
	}
	e.playerColors = map[string]string{
		players[0].ID: "navy",
		players[1].ID: "red",
	}

	return e.buildStateLocked(), nil
}

// ValidateMove verifies edge coordinates, bounds, and legality.
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

	edge, err := parseEdge(move.Data)
	if err != nil {
		return fmt.Errorf("%w: %v", engine.ErrInvalidMove, err)
	}

	if edge.Type == "h" {
		if edge.Row < 0 || edge.Row > e.rows || edge.Col < 0 || edge.Col >= e.cols {
			return engine.ErrInvalidCoordinates
		}
		if e.horizontal[edge.Row][edge.Col] != "" {
			return engine.ErrCellOccupied
		}
	} else if edge.Type == "v" {
		if edge.Row < 0 || edge.Row >= e.rows || edge.Col < 0 || edge.Col > e.cols {
			return engine.ErrInvalidCoordinates
		}
		if e.vertical[edge.Row][edge.Col] != "" {
			return engine.ErrCellOccupied
		}
	} else {
		return errors.New("edge type must be 'h' or 'v'")
	}

	return nil
}

// ApplyMove places the edge, checks for newly closed boxes, and awards bonus turns.
func (e *Engine) ApplyMove(move engine.Move) (*engine.GameState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := e.validateMoveLocked(move); err != nil {
		return nil, err
	}

	edge, _ := parseEdge(move.Data)
	e.lastEdge = &edge
	e.moveCount++
	e.version++

	// 1. Mark edge as claimed by player
	if edge.Type == "h" {
		e.horizontal[edge.Row][edge.Col] = move.PlayerID
	} else {
		e.vertical[edge.Row][edge.Col] = move.PlayerID
	}

	// 2. Check for completed boxes
	newBoxesCompleted := 0

	if edge.Type == "h" {
		// Check box above (Row - 1, Col)
		if edge.Row > 0 && e.isBoxComplete(edge.Row-1, edge.Col) && e.boxes[edge.Row-1][edge.Col] == "" {
			e.boxes[edge.Row-1][edge.Col] = move.PlayerID
			e.scores[move.PlayerID]++
			newBoxesCompleted++
		}
		// Check box below (Row, Col)
		if edge.Row < e.rows && e.isBoxComplete(edge.Row, edge.Col) && e.boxes[edge.Row][edge.Col] == "" {
			e.boxes[edge.Row][edge.Col] = move.PlayerID
			e.scores[move.PlayerID]++
			newBoxesCompleted++
		}
	} else {
		// Check box to left (Row, Col - 1)
		if edge.Col > 0 && e.isBoxComplete(edge.Row, edge.Col-1) && e.boxes[edge.Row][edge.Col-1] == "" {
			e.boxes[edge.Row][edge.Col-1] = move.PlayerID
			e.scores[move.PlayerID]++
			newBoxesCompleted++
		}
		// Check box to right (Row, Col)
		if edge.Col < e.cols && e.isBoxComplete(edge.Row, edge.Col) && e.boxes[edge.Row][edge.Col] == "" {
			e.boxes[edge.Row][edge.Col] = move.PlayerID
			e.scores[move.PlayerID]++
			newBoxesCompleted++
		}
	}

	// 3. Check if all boxes on board are claimed -> Match Completed!
	totalBoxes := e.rows * e.cols
	claimedCount := e.scores[e.players[0].ID] + e.scores[e.players[1].ID]

	if claimedCount >= totalBoxes {
		e.status = engine.StatusFinished
		p1Score := e.scores[e.players[0].ID]
		p2Score := e.scores[e.players[1].ID]

		if p1Score > p2Score {
			e.result = &engine.GameResult{
				WinnerID: e.players[0].ID,
				IsDraw:   false,
				Scores:   e.scores,
				PlayerPlaces: map[string]int{
					e.players[0].ID: 1,
					e.players[1].ID: 2,
				},
				Reason:      fmt.Sprintf("Won %d to %d boxes", p1Score, p2Score),
				CompletedAt: time.Now().UTC(),
			}
		} else if p2Score > p1Score {
			e.result = &engine.GameResult{
				WinnerID: e.players[1].ID,
				IsDraw:   false,
				Scores:   e.scores,
				PlayerPlaces: map[string]int{
					e.players[1].ID: 1,
					e.players[0].ID: 2,
				},
				Reason:      fmt.Sprintf("Won %d to %d boxes", p2Score, p1Score),
				CompletedAt: time.Now().UTC(),
			}
		} else {
			e.result = &engine.GameResult{
				WinnerID: "",
				IsDraw:   true,
				Scores:   e.scores,
				PlayerPlaces: map[string]int{
					e.players[0].ID: 1,
					e.players[1].ID: 1,
				},
				Reason:      "Draw Game (Equal Boxes Claimed)",
				CompletedAt: time.Now().UTC(),
			}
		}
		e.currentTurn = ""
		return e.buildStateLocked(), nil
	}

	// 4. Turn management:
	// If player completed at least one box, they retain the turn (Bonus Turn!).
	// Otherwise, turn passes to the other player.
	if newBoxesCompleted == 0 {
		e.currentTurn = e.getOtherPlayerID(move.PlayerID)
	}

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

func (e *Engine) isBoxComplete(r, c int) bool {
	if r < 0 || r >= e.rows || c < 0 || c >= e.cols {
		return false
	}
	top := e.horizontal[r][c] != ""
	bottom := e.horizontal[r+1][c] != ""
	left := e.vertical[r][c] != ""
	right := e.vertical[r][c+1] != ""
	return top && bottom && left && right
}

func (e *Engine) getOtherPlayerID(playerID string) string {
	for _, p := range e.players {
		if p.ID != playerID {
			return p.ID
		}
	}
	return ""
}

func (e *Engine) buildStateLocked() *engine.GameState {
	claimedCount := e.scores[e.players[0].ID] + e.scores[e.players[1].ID]

	board := BoardState{
		Rows:            e.rows,
		Cols:            e.cols,
		HorizontalEdges: e.horizontal,
		VerticalEdges:   e.vertical,
		Boxes:           e.boxes,
		PlayerInitials:  e.playerInitials,
		PlayerColors:    e.playerColors,
		Scores:          e.scores,
		LastEdge:        e.lastEdge,
		CompletedBoxes:  claimedCount,
		TotalBoxes:      e.rows * e.cols,
	}

	rawBoard, _ := json.Marshal(board)

	return &engine.GameState{
		GameID:      e.gameID,
		GameType:    engine.GameTypeDotsBoxes,
		Status:      e.status,
		CurrentTurn: e.currentTurn,
		MoveCount:   e.moveCount,
		BoardState:  rawBoard,
		Result:      e.result,
		Version:     e.version,
	}
}

func parseEdge(data json.RawMessage) (EdgeKey, error) {
	if len(data) == 0 {
		return EdgeKey{}, errors.New("missing edge data")
	}
	var edge EdgeKey
	if err := json.Unmarshal(data, &edge); err != nil {
		return EdgeKey{}, err
	}
	if edge.Type != "h" && edge.Type != "v" {
		return EdgeKey{}, errors.New("edge type must be 'h' or 'v'")
	}
	return edge, nil
}

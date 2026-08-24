package paperfootball

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"

	"github.com/Abh19avM/recess/internal/games/engine"
)

const (
	TargetScore       = 21
	TouchdownZoneMin  = 90.0
	TouchdownZoneMax  = 100.0
	MaxDowns          = 4
	StartingBallPos   = 20.0
)

func init() {
	engine.Register(engine.GameTypePaperFootball, func() engine.Engine {
		return NewPaperFootballEngine()
	})
}

// Phase represents the current state of play in Paper Football.
type Phase string

const (
	PhaseDrive      Phase = "drive"
	PhaseExtraPoint Phase = "extra_point"
	PhaseFinished   Phase = "finished"
)

// FlickMove represents a flick action across the school desk.
type FlickMove struct {
	Action string  `json:"action"` // "flick", "extra_point", "field_goal"
	Power  float64 `json:"power"`  // 1 to 100 percentage
	Angle  float64 `json:"angle"`  // -45 to +45 degrees
}

// BoardState represents the serializable state of the paper football desk.
type BoardState struct {
	Phase           Phase             `json:"phase"`
	PossessionID    string            `json:"possession_id"`
	BallPosition    float64           `json:"ball_position"` // 0% (own end) to 100% (opponent edge)
	Down            int               `json:"down"`          // 1..4
	Scores          map[string]int    `json:"scores"`
	Quarter         int               `json:"quarter"`
	LastFlickResult string            `json:"last_flick_result,omitempty"` // "touchdown", "advance", "table_fall", "good_kick", "missed_kick"
	LastDistance    float64           `json:"last_distance"`
	Message         string            `json:"message"`
}

// Engine implements the authoritative Paper Football engine.
type Engine struct {
	mu           sync.RWMutex
	gameID       string
	players      []engine.Player
	status       engine.GameStatus
	phase        Phase
	possessionID string
	ballPosition float64
	down         int
	scores       map[string]int
	quarter      int
	lastResult   string
	lastDist     float64
	message      string
	moveCount    int
	version      int64
	result       *engine.GameResult
}

// NewPaperFootballEngine creates a clean Paper Football engine.
func NewPaperFootballEngine() engine.Engine {
	return &Engine{
		status: engine.StatusWaiting,
		scores: make(map[string]int),
	}
}

func (e *Engine) GameType() engine.GameType {
	return engine.GameTypePaperFootball
}

// Initialize configures the match with 2 players.
func (e *Engine) Initialize(gameID string, players []engine.Player, config json.RawMessage) (*engine.GameState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(players) != 2 {
		return nil, fmt.Errorf("%w: Paper Football requires 2 players, got %d", engine.ErrInvalidPlayerCount, len(players))
	}

	e.gameID = gameID
	e.players = players
	e.status = engine.StatusActive
	e.phase = PhaseDrive
	e.possessionID = players[0].ID
	e.ballPosition = StartingBallPos
	e.down = 1
	e.quarter = 1
	e.scores = map[string]int{
		players[0].ID: 0,
		players[1].ID: 0,
	}
	e.lastResult = "Kickoff! Player 1 takes possession from 20% mark."
	e.lastDist = 0
	e.message = "Flick the paper triangle towards the opponent's desk edge!"
	e.moveCount = 0
	e.version = 1
	e.result = nil

	return e.stateLocked(), nil
}

// ValidateMove validates flick parameters.
func (e *Engine) ValidateMove(move engine.Move) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.status != engine.StatusActive {
		return engine.ErrGameFinished
	}
	if move.PlayerID != e.possessionID {
		return engine.ErrNotPlayerTurn
	}

	var flick FlickMove
	if err := json.Unmarshal(move.Data, &flick); err != nil {
		return fmt.Errorf("%w: invalid move data", engine.ErrInvalidMove)
	}

	if flick.Power < 1 || flick.Power > 100 {
		return fmt.Errorf("%w: power must be between 1 and 100", engine.ErrInvalidMove)
	}

	return nil
}

// ApplyMove validates and executes a paper football flick.
func (e *Engine) ApplyMove(move engine.Move) (*engine.GameState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status != engine.StatusActive {
		return nil, engine.ErrGameFinished
	}
	if move.PlayerID != e.possessionID {
		return nil, engine.ErrNotPlayerTurn
	}

	var flick FlickMove
	if err := json.Unmarshal(move.Data, &flick); err != nil {
		return nil, fmt.Errorf("%w: invalid move data", engine.ErrInvalidMove)
	}

	e.moveCount++
	e.version++

	switch e.phase {
	case PhaseDrive:
		e.handleDriveFlick(flick, move.PlayerID)
	case PhaseExtraPoint:
		e.handleExtraPointFlick(flick, move.PlayerID)
	}

	// Check victory condition
	if e.scores[e.players[0].ID] >= TargetScore || e.scores[e.players[1].ID] >= TargetScore {
		e.finishMatch()
	}

	return e.stateLocked(), nil
}

func (e *Engine) handleDriveFlick(flick FlickMove, playerID string) {
	// Calculate distance based on power with angle friction adjustment
	// Cosine of angle penalizes off-center sideways flicks
	rad := flick.Angle * (math.Pi / 180.0)
	effectivePower := flick.Power * math.Cos(rad)
	distanceGained := effectivePower * 0.75 // 100 power = 75% max forward travel

	newPosition := e.ballPosition + distanceGained
	e.lastDist = distanceGained

	// 1. TOUCHDOWN! Paper football hanging over opponent desk edge (90% to 100%)
	if newPosition >= TouchdownZoneMin && newPosition <= TouchdownZoneMax {
		e.scores[playerID] += 6
		e.lastResult = "touchdown"
		e.message = "TOUCHDOWN! Paper football is hanging off the edge! (+6 Points)"
		e.phase = PhaseExtraPoint
		return
	}

	// 2. OVER-FLICK / TABLE FALL! (> 100%)
	if newPosition > TouchdownZoneMax {
		e.lastResult = "table_fall"
		e.message = "OVER-FLICK! Football fell off the desk! Turnover to opponent."
		e.switchPossession(StartingBallPos)
		return
	}

	// 3. Regular Advance
	e.ballPosition = newPosition
	e.lastResult = "advance"
	e.down++

	if e.down > MaxDowns {
		// Turnover on downs!
		e.lastResult = "turnover"
		e.message = "Turnover on downs! Opponent takes over desk possession."
		// Opponent starts from inverse position
		oppPos := math.Max(10.0, 100.0-e.ballPosition)
		e.switchPossession(oppPos)
	} else {
		e.message = fmt.Sprintf("Advanced to %.1f%% of the desk. Down %d of 4.", e.ballPosition, e.down)
	}
}

func (e *Engine) handleExtraPointFlick(flick FlickMove, playerID string) {
	// Extra point through finger uprights: Power >= 40 and Aim within [-20, 20] degrees
	if flick.Power >= 35 && math.Abs(flick.Angle) <= 20 {
		e.scores[playerID] += 1
		e.lastResult = "good_kick"
		e.message = "EXTRA POINT IS GOOD! Clean flick through the uprights! (+1 Point)"
	} else {
		e.lastResult = "missed_kick"
		e.message = "EXTRA POINT MISSED! Hit the upright or wide."
	}

	// Resume regular kickoff possession for opponent
	e.phase = PhaseDrive
	e.switchPossession(StartingBallPos)
}

func (e *Engine) switchPossession(startingPos float64) {
	if e.possessionID == e.players[0].ID {
		e.possessionID = e.players[1].ID
	} else {
		e.possessionID = e.players[0].ID
	}
	e.ballPosition = startingPos
	e.down = 1
}

func (e *Engine) finishMatch() {
	e.status = engine.StatusFinished
	e.phase = PhaseFinished

	p1Score := e.scores[e.players[0].ID]
	p2Score := e.scores[e.players[1].ID]

	var winnerID string
	var isDraw bool
	var reason string

	if p1Score > p2Score {
		winnerID = e.players[0].ID
		reason = fmt.Sprintf("%s won the Paper Football match %d - %d!", e.players[0].Username, p1Score, p2Score)
	} else if p2Score > p1Score {
		winnerID = e.players[1].ID
		reason = fmt.Sprintf("%s won the Paper Football match %d - %d!", e.players[1].Username, p2Score, p1Score)
	} else {
		isDraw = true
		reason = fmt.Sprintf("Match ended in a tie %d - %d!", p1Score, p2Score)
	}

	e.result = &engine.GameResult{
		WinnerID: winnerID,
		IsDraw:   isDraw,
		Scores:   e.scores,
		Reason:   reason,
	}
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
	return e.possessionID
}

func (e *Engine) stateLocked() *engine.GameState {
	board := BoardState{
		Phase:           e.phase,
		PossessionID:    e.possessionID,
		BallPosition:    e.ballPosition,
		Down:            e.down,
		Scores:          e.scores,
		Quarter:         e.quarter,
		LastFlickResult: e.lastResult,
		LastDistance:    e.lastDist,
		Message:         e.message,
	}

	boardData, _ := json.Marshal(board)

	return &engine.GameState{
		GameID:      e.gameID,
		GameType:    engine.GameTypePaperFootball,
		Status:      e.status,
		CurrentTurn: e.possessionID,
		MoveCount:   e.moveCount,
		BoardState:  boardData,
		Result:      e.result,
		Version:     e.version,
	}
}

package handcricket

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Abh19avM/recess/internal/games/engine"
)

func init() {
	engine.MustRegister(engine.GameTypeHandCricket, NewHandCricketEngine)
}

// Hand Cricket Game Phases
const (
	PhaseTossCall      = "toss_call"      // Player 1 calls "odd" or "even"
	PhaseTossThrow     = "toss_throw"     // Both players throw numbers (1-6)
	PhaseTossDecision  = "toss_decision"  // Toss winner chooses "bat" or "bowl"
	PhaseInnings1      = "innings_1"      // First innings active
	PhaseInningsBreak  = "innings_break"  // Brief pause before chase
	PhaseInnings2      = "innings_2"      // Second innings chase active
	PhaseFinished      = "finished"       // Match concluded
)

// Supported Move Actions
const (
	ActionTossCall     = "toss_call"      // data: {"call": "odd"|"even"}
	ActionTossThrow    = "toss_throw"     // data: {"number": 1..6}
	ActionTossDecision = "toss_decision"  // data: {"decision": "bat"|"bowl"}
	ActionChooseNumber = "choose_number"  // data: {"value": 1..6} or {"number": 1..6}
)

// BallRecord stores a single resolved ball in the match scorecard.
type BallRecord struct {
	Over          int    `json:"over"`
	Ball          int    `json:"ball"`
	Innings       int    `json:"innings"`
	BatsmanID     string `json:"batsman_id"`
	BowlerID      string `json:"bowler_id"`
	BatsmanChoice int    `json:"batsman_choice"`
	BowlerChoice  int    `json:"bowler_choice"`
	RunsScored    int    `json:"runs_scored"`
	IsWicket      bool   `json:"is_wicket"`
	Commentary    string `json:"commentary"`
	Timestamp     int64  `json:"timestamp"`
}

// InningsState stores the statistics for a single innings.
type InningsState struct {
	InningsNumber int          `json:"innings_number"`
	BatsmanID     string       `json:"batsman_id"`
	BowlerID      string       `json:"bowler_id"`
	Runs          int          `json:"runs"`
	Wickets       int          `json:"wickets"`
	MaxWickets    int          `json:"max_wickets"`
	BallsBowled   int          `json:"balls_bowled"`
	MaxBalls      int          `json:"max_balls"` // e.g. 6 or 12 balls (1-2 overs)
	IsCompleted   bool         `json:"is_completed"`
	BallHistory   []BallRecord `json:"ball_history"`
}

// BoardState represents the live scoreboard and match snapshot.
type BoardState struct {
	Phase           string            `json:"phase"`
	TossCallerID    string            `json:"toss_caller_id"`
	TossCall        string            `json:"toss_call"` // "odd" | "even"
	TossWinnerID    string            `json:"toss_winner_id"`
	TossDecision    string            `json:"toss_decision"` // "bat" | "bowl"
	BatsmanID       string            `json:"batsman_id"`
	BowlerID        string            `json:"bowler_id"`
	CurrentInnings  int               `json:"current_innings"`
	Innings1        InningsState      `json:"innings_1"`
	Innings2        InningsState      `json:"innings_2"`
	Target          int               `json:"target,omitempty"` // Innings 1 runs + 1
	PendingChoices  map[string]bool   `json:"pending_choices"`  // PlayerID -> has submitted secret choice
	LastResolved    *BallRecord       `json:"last_resolved,omitempty"`
	PlayerRoles     map[string]string `json:"player_roles"` // PlayerID -> "Batsman" | "Bowler"
}

// Engine implements the engine.Engine contract for Hand Cricket.
type Engine struct {
	mu             sync.RWMutex
	gameID         string
	status         engine.GameStatus
	players        []engine.Player
	phase          string
	tossCallerID   string
	tossCall       string
	tossWinnerID   string
	tossDecision   string
	tossNumbers    map[string]int
	currentInnings int
	innings1       InningsState
	innings2       InningsState
	pendingChoices map[string]int // PlayerID -> chosen secret number
	lastResolved   *BallRecord
	maxBalls       int // Default 6 balls (1 over) or 12 balls (2 overs)
	maxWickets     int // Default 1 wicket for classic lightning school play
	moveCount      int
	version        int64
	result         *engine.GameResult
}

// NewHandCricketEngine initializes a fresh Hand Cricket engine.
func NewHandCricketEngine() engine.Engine {
	return &Engine{
		status:         engine.StatusWaiting,
		phase:          PhaseTossCall,
		tossNumbers:    make(map[string]int),
		pendingChoices: make(map[string]int),
		maxBalls:       12, // 2 overs
		maxWickets:     1,  // 1 wicket
	}
}

func (e *Engine) GameType() engine.GameType {
	return engine.GameTypeHandCricket
}

// Initialize configures the Hand Cricket match with 2 players.
func (e *Engine) Initialize(gameID string, players []engine.Player, config json.RawMessage) (*engine.GameState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(players) != 2 {
		return nil, fmt.Errorf("%w: Hand Cricket requires exactly 2 players, got %d", engine.ErrInvalidPlayerCount, len(players))
	}

	e.gameID = gameID
	e.players = players
	e.status = engine.StatusActive
	e.phase = PhaseTossCall
	e.tossCallerID = players[0].ID
	e.tossCall = ""
	e.tossWinnerID = ""
	e.tossDecision = ""
	e.tossNumbers = make(map[string]int)
	e.pendingChoices = make(map[string]int)
	e.currentInnings = 1
	e.moveCount = 0
	e.version = 1
	e.result = nil
	e.lastResolved = nil

	// Parse optional custom config (e.g. max_balls)
	if len(config) > 0 {
		var cfg struct {
			MaxBalls   int `json:"max_balls"`
			MaxWickets int `json:"max_wickets"`
		}
		if err := json.Unmarshal(config, &cfg); err == nil {
			if cfg.MaxBalls > 0 {
				e.maxBalls = cfg.MaxBalls
			}
			if cfg.MaxWickets > 0 {
				e.maxWickets = cfg.MaxWickets
			}
		}
	}

	e.innings1 = InningsState{
		InningsNumber: 1,
		MaxBalls:      e.maxBalls,
		MaxWickets:    e.maxWickets,
		BallHistory:   make([]BallRecord, 0),
	}
	e.innings2 = InningsState{
		InningsNumber: 2,
		MaxBalls:      e.maxBalls,
		MaxWickets:    e.maxWickets,
		BallHistory:   make([]BallRecord, 0),
	}

	return e.buildStateLocked(), nil
}

// ValidateMove checks whether an action is legal in the current phase.
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

	if !e.isPlayerInGame(move.PlayerID) {
		return engine.ErrPlayerNotInGame
	}

	switch e.phase {
	case PhaseTossCall:
		if move.PlayerID != e.tossCallerID {
			return errors.New("only toss caller can choose odd/even")
		}
		var payload struct {
			Call string `json:"call"`
		}
		if err := json.Unmarshal(move.Data, &payload); err != nil || (payload.Call != "odd" && payload.Call != "even") {
			return errors.New("toss call must be 'odd' or 'even'")
		}

	case PhaseTossThrow:
		num, err := extractNumber(move.Data)
		if err != nil || num < 1 || num > 6 {
			return errors.New("toss number must be an integer between 1 and 6")
		}
		if _, exists := e.tossNumbers[move.PlayerID]; exists {
			return errors.New("toss number already submitted")
		}

	case PhaseTossDecision:
		if move.PlayerID != e.tossWinnerID {
			return errors.New("only toss winner can choose bat/bowl")
		}
		var payload struct {
			Decision string `json:"decision"`
		}
		if err := json.Unmarshal(move.Data, &payload); err != nil || (payload.Decision != "bat" && payload.Decision != "bowl") {
			return errors.New("toss decision must be 'bat' or 'bowl'")
		}

	case PhaseInnings1, PhaseInnings2:
		num, err := extractNumber(move.Data)
		if err != nil || num < 1 || num > 6 {
			return errors.New("choice must be an integer between 1 and 6")
		}
		if _, exists := e.pendingChoices[move.PlayerID]; exists {
			return errors.New("number already chosen for this ball")
		}

	default:
		return errors.New("invalid game phase")
	}

	return nil
}

// ApplyMove processes the player's action and advances the match.
func (e *Engine) ApplyMove(move engine.Move) (*engine.GameState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := e.validateMoveLocked(move); err != nil {
		return nil, err
	}

	e.moveCount++
	e.version++

	switch e.phase {
	case PhaseTossCall:
		var payload struct {
			Call string `json:"call"`
		}
		_ = json.Unmarshal(move.Data, &payload)
		e.tossCall = payload.Call
		e.phase = PhaseTossThrow

	case PhaseTossThrow:
		num, _ := extractNumber(move.Data)
		e.tossNumbers[move.PlayerID] = num

		// Once both players throw numbers, determine toss winner
		if len(e.tossNumbers) == 2 {
			p1Num := e.tossNumbers[e.players[0].ID]
			p2Num := e.tossNumbers[e.players[1].ID]
			sum := p1Num + p2Num
			isEven := (sum % 2 == 0)

			p1Won := (e.tossCall == "even" && isEven) || (e.tossCall == "odd" && !isEven)
			if p1Won {
				e.tossWinnerID = e.players[0].ID
			} else {
				e.tossWinnerID = e.players[1].ID
			}
			e.phase = PhaseTossDecision
		}

	case PhaseTossDecision:
		var payload struct {
			Decision string `json:"decision"`
		}
		_ = json.Unmarshal(move.Data, &payload)
		e.tossDecision = payload.Decision

		// Assign batsman & bowler for Innings 1
		otherID := e.getOtherPlayerID(e.tossWinnerID)
		if e.tossDecision == "bat" {
			e.innings1.BatsmanID = e.tossWinnerID
			e.innings1.BowlerID = otherID
		} else {
			e.innings1.BatsmanID = otherID
			e.innings1.BowlerID = e.tossWinnerID
		}

		e.innings2.BatsmanID = e.innings1.BowlerID
		e.innings2.BowlerID = e.innings1.BatsmanID

		e.phase = PhaseInnings1
		e.currentInnings = 1
		e.pendingChoices = make(map[string]int)

	case PhaseInnings1:
		num, _ := extractNumber(move.Data)
		e.pendingChoices[move.PlayerID] = num

		// Once both batsman and bowler choose, resolve ball
		if len(e.pendingChoices) == 2 {
			batsmanNum := e.pendingChoices[e.innings1.BatsmanID]
			bowlerNum := e.pendingChoices[e.innings1.BowlerID]
			e.innings1.BallsBowled++

			overNum := (e.innings1.BallsBowled - 1) / 6
			ballNum := ((e.innings1.BallsBowled - 1) % 6) + 1

			isWicket := (batsmanNum == bowlerNum)
			runsScored := 0
			commentary := ""

			if isWicket {
				e.innings1.Wickets++
				commentary = fmt.Sprintf("OUT! Both chose %d. Wicket falls!", batsmanNum)
			} else {
				runsScored = batsmanNum
				e.innings1.Runs += runsScored
				if runsScored == 6 {
					commentary = "SIX! Smashed out of the classroom window!"
				} else if runsScored == 4 {
					commentary = "FOUR! Pierces the desk boundary!"
				} else {
					commentary = fmt.Sprintf("Batsman scores %d runs.", runsScored)
				}
			}

			record := BallRecord{
				Over:          overNum,
				Ball:          ballNum,
				Innings:       1,
				BatsmanID:     e.innings1.BatsmanID,
				BowlerID:      e.innings1.BowlerID,
				BatsmanChoice: batsmanNum,
				BowlerChoice:  bowlerNum,
				RunsScored:    runsScored,
				IsWicket:      isWicket,
				Commentary:    commentary,
				Timestamp:     time.Now().UnixMilli(),
			}
			e.innings1.BallHistory = append(e.innings1.BallHistory, record)
			e.lastResolved = &record
			e.pendingChoices = make(map[string]int)

			// Check if Innings 1 is complete (Wicket or Overs completed)
			if e.innings1.Wickets >= e.innings1.MaxWickets || e.innings1.BallsBowled >= e.innings1.MaxBalls {
				e.innings1.IsCompleted = true
				e.phase = PhaseInnings2
				e.currentInnings = 2
			}
		}

	case PhaseInnings2:
		num, _ := extractNumber(move.Data)
		e.pendingChoices[move.PlayerID] = num

		// Once both players choose in 2nd innings, resolve ball
		if len(e.pendingChoices) == 2 {
			batsmanNum := e.pendingChoices[e.innings2.BatsmanID]
			bowlerNum := e.pendingChoices[e.innings2.BowlerID]
			e.innings2.BallsBowled++

			overNum := (e.innings2.BallsBowled - 1) / 6
			ballNum := ((e.innings2.BallsBowled - 1) % 6) + 1

			isWicket := (batsmanNum == bowlerNum)
			runsScored := 0
			commentary := ""
			target := e.innings1.Runs + 1

			if isWicket {
				e.innings2.Wickets++
				commentary = fmt.Sprintf("OUT! Both chose %d. Wicket falls!", batsmanNum)
			} else {
				runsScored = batsmanNum
				e.innings2.Runs += runsScored
				if runsScored == 6 {
					commentary = "SIX! Massive strike in the chase!"
				} else if runsScored == 4 {
					commentary = "FOUR! Fantastic boundary!"
				} else {
					commentary = fmt.Sprintf("Batsman scores %d runs.", runsScored)
				}
			}

			record := BallRecord{
				Over:          overNum,
				Ball:          ballNum,
				Innings:       2,
				BatsmanID:     e.innings2.BatsmanID,
				BowlerID:      e.innings2.BowlerID,
				BatsmanChoice: batsmanNum,
				BowlerChoice:  bowlerNum,
				RunsScored:    runsScored,
				IsWicket:      isWicket,
				Commentary:    commentary,
				Timestamp:     time.Now().UnixMilli(),
			}
			e.innings2.BallHistory = append(e.innings2.BallHistory, record)
			e.lastResolved = &record
			e.pendingChoices = make(map[string]int)

			// 1. Chasing team reaches or exceeds target -> Win!
			if e.innings2.Runs >= target {
				e.innings2.IsCompleted = true
				e.phase = PhaseFinished
				e.status = engine.StatusFinished
				e.result = &engine.GameResult{
					WinnerID: e.innings2.BatsmanID,
					IsDraw:   false,
					Scores: map[string]int{
						e.innings1.BatsmanID: e.innings1.Runs,
						e.innings2.BatsmanID: e.innings2.Runs,
					},
					PlayerPlaces: map[string]int{
						e.innings2.BatsmanID: 1,
						e.innings1.BatsmanID: 2,
					},
					Reason:      fmt.Sprintf("Chased target of %d runs", target),
					CompletedAt: time.Now().UTC(),
				}
			} else if e.innings2.Wickets >= e.innings2.MaxWickets || e.innings2.BallsBowled >= e.innings2.MaxBalls {
				// 2. Chaser all out or overs done
				e.innings2.IsCompleted = true
				e.phase = PhaseFinished
				e.status = engine.StatusFinished

				if e.innings2.Runs == e.innings1.Runs {
					// Match is a Tie / Draw
					e.result = &engine.GameResult{
						WinnerID: "",
						IsDraw:   true,
						Scores: map[string]int{
							e.innings1.BatsmanID: e.innings1.Runs,
							e.innings2.BatsmanID: e.innings2.Runs,
						},
						PlayerPlaces: map[string]int{
							e.innings1.BatsmanID: 1,
							e.innings2.BatsmanID: 1,
						},
						Reason:      "Match Tied",
						CompletedAt: time.Now().UTC(),
					}
				} else {
					// Defending team (Innings 1 batsman) Wins!
					margin := e.innings1.Runs - e.innings2.Runs
					e.result = &engine.GameResult{
						WinnerID: e.innings1.BatsmanID,
						IsDraw:   false,
						Scores: map[string]int{
							e.innings1.BatsmanID: e.innings1.Runs,
							e.innings2.BatsmanID: e.innings2.Runs,
						},
						PlayerPlaces: map[string]int{
							e.innings1.BatsmanID: 1,
							e.innings2.BatsmanID: 2,
						},
						Reason:      fmt.Sprintf("Defended total by %d runs", margin),
						CompletedAt: time.Now().UTC(),
					}
				}
			}
		}
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
	return e.nextTurnLocked()
}

func (e *Engine) nextTurnLocked() string {
	if e.status != engine.StatusActive {
		return ""
	}
	switch e.phase {
	case PhaseTossCall:
		return e.tossCallerID
	case PhaseTossThrow:
		for _, p := range e.players {
			if _, done := e.tossNumbers[p.ID]; !done {
				return p.ID
			}
		}
		return ""
	case PhaseTossDecision:
		return e.tossWinnerID
	case PhaseInnings1, PhaseInnings2:
		for _, p := range e.players {
			if _, done := e.pendingChoices[p.ID]; !done {
				return p.ID
			}
		}
		return ""
	default:
		return ""
	}
}

func (e *Engine) isPlayerInGame(playerID string) bool {
	for _, p := range e.players {
		if p.ID == playerID {
			return true
		}
	}
	return false
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
	pendingMap := make(map[string]bool)
	for _, p := range e.players {
		if e.phase == PhaseTossThrow {
			_, pendingMap[p.ID] = e.tossNumbers[p.ID]
		} else if e.phase == PhaseInnings1 || e.phase == PhaseInnings2 {
			_, pendingMap[p.ID] = e.pendingChoices[p.ID]
		}
	}

	activeBatsman := ""
	activeBowler := ""
	if e.phase == PhaseInnings1 {
		activeBatsman = e.innings1.BatsmanID
		activeBowler = e.innings1.BowlerID
	} else if e.phase == PhaseInnings2 {
		activeBatsman = e.innings2.BatsmanID
		activeBowler = e.innings2.BowlerID
	}

	playerRoles := make(map[string]string)
	if activeBatsman != "" {
		playerRoles[activeBatsman] = "Batsman"
	}
	if activeBowler != "" {
		playerRoles[activeBowler] = "Bowler"
	}

	target := 0
	if e.currentInnings == 2 || e.innings1.IsCompleted {
		target = e.innings1.Runs + 1
	}

	board := BoardState{
		Phase:          e.phase,
		TossCallerID:   e.tossCallerID,
		TossCall:       e.tossCall,
		TossWinnerID:   e.tossWinnerID,
		TossDecision:   e.tossDecision,
		BatsmanID:      activeBatsman,
		BowlerID:       activeBowler,
		CurrentInnings: e.currentInnings,
		Innings1:       e.innings1,
		Innings2:       e.innings2,
		Target:         target,
		PendingChoices: pendingMap,
		LastResolved:   e.lastResolved,
		PlayerRoles:    playerRoles,
	}

	rawBoard, _ := json.Marshal(board)

	return &engine.GameState{
		GameID:      e.gameID,
		GameType:    engine.GameTypeHandCricket,
		Status:      e.status,
		CurrentTurn: e.nextTurnLocked(),
		MoveCount:   e.moveCount,
		BoardState:  rawBoard,
		Result:      e.result,
		Version:     e.version,
	}
}

func extractNumber(data json.RawMessage) (int, error) {
	if len(data) == 0 {
		return 0, errors.New("empty data")
	}
	var numPayload struct {
		Number int `json:"number"`
		Value  int `json:"value"`
	}
	if err := json.Unmarshal(data, &numPayload); err == nil {
		if numPayload.Number > 0 {
			return numPayload.Number, nil
		}
		if numPayload.Value > 0 {
			return numPayload.Value, nil
		}
	}
	return 0, errors.New("number not found in data")
}

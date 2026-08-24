package npt

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/Abh19avM/recess/internal/games/engine"
)

var defaultLetters = []string{"S", "A", "M", "R", "P", "B", "C", "D", "T", "L", "K", "G", "H", "N", "F"}

const (
	DefaultRounds = 3
)

func init() {
	engine.Register(engine.GameTypeNPAT, func() engine.Engine {
		return NewNPATEngine()
	})
}

// CategoryEntries holds the 4 answers for a round.
type CategoryEntries struct {
	Name   string `json:"name"`
	Place  string `json:"place"`
	Animal string `json:"animal"`
	Thing  string `json:"thing"`
}

// RoundScore holds the scored breakdown for a player in a round.
type RoundScore struct {
	NamePoints   int             `json:"name_points"`
	PlacePoints  int             `json:"place_points"`
	AnimalPoints int             `json:"animal_points"`
	ThingPoints  int             `json:"thing_points"`
	TotalPoints  int             `json:"total_points"`
	Entries      CategoryEntries `json:"entries"`
}

// BoardState represents the serializable state of the NPAT classroom desk.
type BoardState struct {
	CurrentRound     int                      `json:"current_round"`
	TotalRounds      int                      `json:"total_rounds"`
	CurrentLetter    string                   `json:"current_letter"`
	StopCallerID     string                   `json:"stop_caller_id,omitempty"`
	SubmittedPlayers []string                 `json:"submitted_players"`
	CumulativeScores map[string]int           `json:"cumulative_scores"`
	LastRoundScores  map[string]RoundScore    `json:"last_round_scores,omitempty"`
	Phase            string                   `json:"phase"` // "writing", "round_summary", "finished"
}

// Engine implements the authoritative NPAT game engine.
type Engine struct {
	mu               sync.RWMutex
	gameID           string
	players          []engine.Player
	status           engine.GameStatus
	currentRound     int
	totalRounds      int
	currentLetter    string
	stopCallerID     string
	submittedPlayers []string
	roundEntries     map[string]CategoryEntries
	cumulativeScores map[string]int
	lastRoundScores  map[string]RoundScore
	phase            string
	moveCount        int
	version          int64
	result           *engine.GameResult
}

// NewNPATEngine creates a new NPAT engine instance.
func NewNPATEngine() engine.Engine {
	return &Engine{
		status:           engine.StatusWaiting,
		cumulativeScores: make(map[string]int),
		roundEntries:     make(map[string]CategoryEntries),
		lastRoundScores:  make(map[string]RoundScore),
	}
}

func (e *Engine) GameType() engine.GameType {
	return engine.GameTypeNPAT
}

// Initialize configures the match with players.
func (e *Engine) Initialize(gameID string, players []engine.Player, config json.RawMessage) (*engine.GameState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(players) < 2 {
		return nil, fmt.Errorf("%w: NPAT requires at least 2 players, got %d", engine.ErrInvalidPlayerCount, len(players))
	}

	e.gameID = gameID
	e.players = players
	e.status = engine.StatusActive
	e.currentRound = 1
	e.totalRounds = DefaultRounds
	e.currentLetter = defaultLetters[0]
	e.stopCallerID = ""
	e.submittedPlayers = nil
	e.roundEntries = make(map[string]CategoryEntries)
	e.cumulativeScores = make(map[string]int)
	for _, p := range players {
		e.cumulativeScores[p.ID] = 0
	}
	e.lastRoundScores = make(map[string]RoundScore)
	e.phase = "writing"
	e.moveCount = 0
	e.version = 1
	e.result = nil

	return e.stateLocked(), nil
}

// ValidateMove checks if a player is submitting valid entries.
func (e *Engine) ValidateMove(move engine.Move) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.status != engine.StatusActive {
		return engine.ErrGameFinished
	}

	// Verify player is participant
	found := false
	for _, p := range e.players {
		if p.ID == move.PlayerID {
			found = true
			break
		}
	}
	if !found {
		return engine.ErrPlayerNotInGame
	}

	return nil
}

// ApplyMove processes round submissions and evaluates scores when all players have submitted.
func (e *Engine) ApplyMove(move engine.Move) (*engine.GameState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status != engine.StatusActive {
		return nil, engine.ErrGameFinished
	}

	var entries CategoryEntries
	if err := json.Unmarshal(move.Data, &entries); err != nil {
		return nil, fmt.Errorf("%w: invalid category entries", engine.ErrInvalidMove)
	}

	e.moveCount++
	e.version++

	// Record submission
	e.roundEntries[move.PlayerID] = entries
	if !contains(e.submittedPlayers, move.PlayerID) {
		e.submittedPlayers = append(e.submittedPlayers, move.PlayerID)
	}

	if move.Action == "call_stop" && e.stopCallerID == "" {
		e.stopCallerID = move.PlayerID
	}

	// When all active players have submitted, evaluate round!
	if len(e.submittedPlayers) >= len(e.players) {
		e.evaluateRoundLocked()
	}

	return e.stateLocked(), nil
}

func (e *Engine) evaluateRoundLocked() {
	targetLetter := strings.ToUpper(e.currentLetter)
	roundResults := make(map[string]RoundScore)

	for _, p := range e.players {
		entries := e.roundEntries[p.ID]

		namePts := e.scoreCategory(entries.Name, targetLetter, "name", p.ID)
		placePts := e.scoreCategory(entries.Place, targetLetter, "place", p.ID)
		animalPts := e.scoreCategory(entries.Animal, targetLetter, "animal", p.ID)
		thingPts := e.scoreCategory(entries.Thing, targetLetter, "thing", p.ID)
		total := namePts + placePts + animalPts + thingPts

		roundResults[p.ID] = RoundScore{
			NamePoints:   namePts,
			PlacePoints:  placePts,
			AnimalPoints: animalPts,
			ThingPoints:  thingPts,
			TotalPoints:  total,
			Entries:      entries,
		}

		e.cumulativeScores[p.ID] += total
	}

	e.lastRoundScores = roundResults

	// Check if final round reached
	if e.currentRound >= e.totalRounds {
		e.finishMatchLocked()
	} else {
		// Advance to next round
		e.currentRound++
		letterIndex := (e.currentRound - 1) % len(defaultLetters)
		e.currentLetter = defaultLetters[letterIndex]
		e.stopCallerID = ""
		e.submittedPlayers = nil
		e.roundEntries = make(map[string]CategoryEntries)
		e.phase = "writing"
	}
}

func (e *Engine) scoreCategory(word, targetLetter, category, playerID string) int {
	clean := strings.TrimSpace(strings.ToUpper(word))
	if len(clean) < 2 {
		return 0 // Blank or too short
	}
	if !strings.HasPrefix(clean, targetLetter) {
		return 0 // Wrong starting letter
	}

	// Check if another player wrote the exact same word
	isShared := false
	for otherID, otherEntries := range e.roundEntries {
		if otherID == playerID {
			continue
		}
		var otherWord string
		switch category {
		case "name":
			otherWord = otherEntries.Name
		case "place":
			otherWord = otherEntries.Place
		case "animal":
			otherWord = otherEntries.Animal
		case "thing":
			otherWord = otherEntries.Thing
		}

		if strings.TrimSpace(strings.ToUpper(otherWord)) == clean {
			isShared = true
			break
		}
	}

	if isShared {
		return 5 // Shared word
	}
	return 10 // Unique valid word!
}

func (e *Engine) finishMatchLocked() {
	e.status = engine.StatusFinished
	e.phase = "finished"

	var winnerID string
	highestScore := -1
	isDraw := false

	for _, p := range e.players {
		score := e.cumulativeScores[p.ID]
		if score > highestScore {
			highestScore = score
			winnerID = p.ID
			isDraw = false
		} else if score == highestScore {
			isDraw = true
		}
	}

	var reason string
	if isDraw {
		reason = fmt.Sprintf("Match ended in a tie with %d points!", highestScore)
		winnerID = ""
	} else {
		reason = fmt.Sprintf("Player %s won with %d points!", winnerID, highestScore)
	}

	e.result = &engine.GameResult{
		WinnerID: winnerID,
		IsDraw:   isDraw,
		Scores:   e.cumulativeScores,
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
	return "" // Simultaneous round entries
}

func (e *Engine) stateLocked() *engine.GameState {
	board := BoardState{
		CurrentRound:     e.currentRound,
		TotalRounds:      e.totalRounds,
		CurrentLetter:    e.currentLetter,
		StopCallerID:     e.stopCallerID,
		SubmittedPlayers: e.submittedPlayers,
		CumulativeScores: e.cumulativeScores,
		LastRoundScores:  e.lastRoundScores,
		Phase:            e.phase,
	}

	boardData, _ := json.Marshal(board)

	return &engine.GameState{
		GameID:      e.gameID,
		GameType:    engine.GameTypeNPAT,
		Status:      e.status,
		CurrentTurn: "",
		MoveCount:   e.moveCount,
		BoardState:  boardData,
		Result:      e.result,
		Version:     e.version,
	}
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

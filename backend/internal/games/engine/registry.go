package engine

import (
	"sync"
)

// Registry manages the registration and instantiation of GameEngines by GameType.
type Registry interface {
	Register(gameType GameType, factory Factory) error
	MustRegister(gameType GameType, factory Factory)
	Create(gameType GameType) (Engine, error)
	AvailableGames() []GameType
	IsRegistered(gameType GameType) bool
}

type registry struct {
	mu        sync.RWMutex
	factories map[GameType]Factory
}

// NewRegistry initializes an empty Registry instance.
func NewRegistry() Registry {
	return &registry{
		factories: make(map[GameType]Factory),
	}
}

// DefaultRegistry is the global singleton registry for game engines.
var DefaultRegistry = NewRegistry()

// Register registers a new GameEngine factory for a specific GameType.
func (r *registry) Register(gameType GameType, factory Factory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[gameType]; exists {
		return ErrGameAlreadyExists
	}
	r.factories[gameType] = factory
	return nil
}

// MustRegister registers a factory and panics if registration fails.
func (r *registry) MustRegister(gameType GameType, factory Factory) {
	if err := r.Register(gameType, factory); err != nil {
		panic(err)
	}
}

// Create instantiates a new Engine for the specified GameType.
func (r *registry) Create(gameType GameType) (Engine, error) {
	r.mu.RLock()
	factory, exists := r.factories[gameType]
	r.mu.RUnlock()

	if !exists {
		return nil, ErrGameNotFound
	}
	return factory(), nil
}

// AvailableGames returns a list of all registered game types.
func (r *registry) AvailableGames() []GameType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	games := make([]GameType, 0, len(r.factories))
	for g := range r.factories {
		games = append(games, g)
	}
	return games
}

// IsRegistered checks whether a game type has a registered factory.
func (r *registry) IsRegistered(gameType GameType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.factories[gameType]
	return exists
}

// Package-level convenience functions targeting DefaultRegistry:

// Register registers an engine factory to the default registry.
func Register(gameType GameType, factory Factory) error {
	return DefaultRegistry.Register(gameType, factory)
}

// MustRegister registers an engine factory to the default registry, panicking on error.
func MustRegister(gameType GameType, factory Factory) {
	DefaultRegistry.MustRegister(gameType, factory)
}

// Create instantiates an engine from the default registry.
func Create(gameType GameType) (Engine, error) {
	return DefaultRegistry.Create(gameType)
}

// AvailableGames returns all registered games in the default registry.
func AvailableGames() []GameType {
	return DefaultRegistry.AvailableGames()
}

// IsRegistered checks if a game is registered in the default registry.
func IsRegistered(gameType GameType) bool {
	return DefaultRegistry.IsRegistered(gameType)
}

package action

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// Action is the contract all faction actions implement.
type Action interface {
	Name() string
	Validate(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) bool
	Inputs(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error
	Resolve(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error
	Output() ([]domain.Mutation, error)
}

// ActionFactory produces a fresh Action wired to the given Collector.
type ActionFactory func(Collector) Action

// ActionEngine orchestrates action resolution for a faction's turn.
type ActionEngine struct {
	factories []ActionFactory
}

func New() *ActionEngine {
	return &ActionEngine{}
}

// Register adds an action factory to the engine's registry.
func (ae *ActionEngine) Register(factory ActionFactory) {
	ae.factories = append(ae.factories, factory)
}

// AvailableActions instantiates each registered factory with the collector and
// returns the actions that pass Validate for the current faction and state.
func (ae *ActionEngine) AvailableActions(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook, collector Collector) []Action {
	var available []Action
	for _, factory := range ae.factories {
		a := factory(collector)
		if a.Validate(faction, factionState, rulebook) {
			available = append(available, a)
		}
	}
	return available
}

// Run calls Inputs, Resolve, and Output in order on the selected action.
func (ae *ActionEngine) Run(a Action, faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) ([]domain.Mutation, error) {
	if err := a.Inputs(faction, factionState, rulebook); err != nil {
		return nil, err
	}
	if err := a.Resolve(faction, factionState, rulebook); err != nil {
		return nil, err
	}
	return a.Output()
}

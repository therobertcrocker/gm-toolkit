package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// ActionFactory produces a fresh Action wired to the given InputCollector.
// Factories that need a roller or other engine collaborators capture them in
// the closure passed to Register.
type ActionFactory func(InputCollector) Action

// ActionEngine orchestrates action resolution for a faction's turn.
type ActionEngine struct {
	factories []ActionFactory
}

func newActionEngine() *ActionEngine {
	return &ActionEngine{}
}

// Register adds an action factory to the engine's registry.
func (ae *ActionEngine) Register(factory ActionFactory) {
	ae.factories = append(ae.factories, factory)
}

// AvailableActions instantiates each registered factory with the collector and
// returns the actions that pass Validate for the current faction and state.
// Returned actions are runnable as-is — the orchestrator does not reconstruct
// them after selection.
func (ae *ActionEngine) AvailableActions(faction *domain.Faction, factionState *state.FactionState, rulebook *loader.Rulebook, collector InputCollector) []Action {
	var available []Action
	for _, factory := range ae.factories {
		action := factory(collector)
		if action.Validate(faction, factionState, rulebook) {
			available = append(available, action)
		}
	}
	return available
}

// Run calls Inputs, Resolve, and Output in order on the selected action and
// returns the resulting mutation list.
func (ae *ActionEngine) Run(action Action, faction *domain.Faction, factionState *state.FactionState, rulebook *loader.Rulebook) ([]domain.Mutation, error) {
	if err := action.Inputs(faction, factionState, rulebook); err != nil {
		return nil, err
	}
	if err := action.Resolve(faction, factionState, rulebook); err != nil {
		return nil, err
	}
	return action.Output()
}

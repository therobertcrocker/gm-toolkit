package action

import (
	"log/slog"

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
	log       *slog.Logger
}

func New(log *slog.Logger) *ActionEngine {
	return &ActionEngine{log: log}
}

// Register adds an action factory to the engine's registry.
func (ae *ActionEngine) Register(factory ActionFactory) {
	ae.factories = append(ae.factories, factory)
}

// AvailableActions instantiates each registered factory with the collector and
// returns the actions that pass Validate for the current faction and state.
func (ae *ActionEngine) AvailableActions(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook, collector Collector, log *slog.Logger) []Action {
	log.Info("checking available actions")
	var available []Action
	for _, factory := range ae.factories {
		a := factory(collector)
		if a.Validate(faction, factionState, rulebook) {
			available = append(available, a)
		}
	}
	log.Info("available actions", "count", len(available))
	return available
}

// Run calls Inputs, Resolve, and Output in order on the selected action.
func (ae *ActionEngine) Run(a Action, faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook, log *slog.Logger) ([]domain.Mutation, error) {
	log.Info("running action", "action", a.Name())
	if err := a.Inputs(faction, factionState, rulebook); err != nil {
		return nil, err
	}
	if err := a.Resolve(faction, factionState, rulebook); err != nil {
		return nil, err
	}
	mutations, err := a.Output()
	if err != nil {
		return nil, err
	}
	log.Info("action resolved", "mutations", len(mutations))
	return mutations, nil
}

package action

import (
	"errors"
	"log/slog"
	"maps"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

var (
	ErrNoSelection = errors.New("action: no selection")
	// ErrTurnCanceled is sent on an ask Reply when the GM dismisses an action
	// prompt (Esc). Classified Recoverable, so the faction's action is skipped
	// and the cycle continues. Lives here (not the TUI adapter) so the engine's
	// error classifier can reference it without a TUI import.
	ErrTurnCanceled = errors.New("action: turn canceled by user")
	// ErrActionUnavailable is returned by the adapter's actionCollector for an
	// action-unique prompt whose per-action commit has not landed. Recoverable +
	// surfaced; the SelectAction disable-set keeps it unreachable in practice.
	ErrActionUnavailable = errors.New("action: prompt not yet available")
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

// Names returns the name of every registered action, in registration order. It
// builds each action with no collector, so a factory must not read its collector
// while it builds.
func (ae *ActionEngine) Names() []string {
	names := make([]string, 0, len(ae.factories))
	for _, factory := range ae.factories {
		names = append(names, factory(nil).Name())
	}
	return names
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

// Run calls Inputs, Resolve, and Output in order on the selected action. A failed
// action drops its mutations, so the budgets it spent on live state are restored.
func (ae *ActionEngine) Run(a Action, faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook, log *slog.Logger) ([]domain.Mutation, error) {
	log.Info("running action", "action", a.Name())
	budgets := snapshotBudgets(factionState)
	mutations, err := runSteps(a, faction, factionState, rulebook)
	if err != nil {
		restoreBudgets(factionState, budgets)
		return nil, err
	}
	log.Info("action resolved", "mutations", len(mutations))
	return mutations, nil
}

func runSteps(a Action, faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) ([]domain.Mutation, error) {
	if err := a.Inputs(faction, factionState, rulebook); err != nil {
		return nil, err
	}
	if err := a.Resolve(faction, factionState, rulebook); err != nil {
		return nil, err
	}
	return a.Output()
}

func snapshotBudgets(factionState *state.FactionState) map[string]map[string]int {
	budgets := make(map[string]map[string]int, len(factionState.Factions))
	for factionID, faction := range factionState.Factions {
		budgets[factionID] = maps.Clone(faction.HookBudgets)
	}
	return budgets
}

func restoreBudgets(factionState *state.FactionState, budgets map[string]map[string]int) {
	for factionID, faction := range factionState.Factions {
		faction.HookBudgets = budgets[factionID]
	}
}

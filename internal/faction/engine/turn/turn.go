package turn

import (
	"errors"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

var (
	ErrTurnInProgress = errors.New("a turn is already in progress")
	ErrNoTurnActive   = errors.New("no turn is currently in progress")
	ErrNoFactions     = errors.New("no factions in campaign")
)

// TurnEngine manages turn lifecycle: ordering, bookkeeping computation, and
// pause/resume cursor state. Mutation application is owned by the orchestrator.
type TurnEngine struct {
	roller   domain.Roller
	rulebook *rulebook.Rulebook
}

func New(roller domain.Roller, rulebook *rulebook.Rulebook) *TurnEngine {
	return &TurnEngine{roller: roller, rulebook: rulebook}
}

// InProgress reports whether a turn is currently active.
func (t *TurnEngine) InProgress(factionState *state.FactionState) bool {
	return factionState.CurrentTurn != nil && factionState.CurrentTurn.InProgress
}

// Start initializes a new turn by rolling faction order. Returns ErrTurnInProgress
// if a turn is already active.
func (t *TurnEngine) Start(factionState *state.FactionState) error {
	if t.InProgress(factionState) {
		return ErrTurnInProgress
	}
	if len(factionState.Factions) == 0 {
		return ErrNoFactions
	}

	factionState.CycleNumber++
	readyAllAssets(factionState)
	factionState.CurrentTurn = &domain.TurnState{
		InProgress:   true,
		CycleNumber:  factionState.CycleNumber,
		FactionOrder: buildFactionOrder(factionState.Factions, t.roller),
		CurrentIndex: 0,
		Phase:        domain.PhaseBookkeeping,
	}
	return nil
}

// CurrentFaction returns the faction whose turn it currently is.
func (t *TurnEngine) CurrentFaction(factionState *state.FactionState) (*domain.Faction, error) {
	if !t.InProgress(factionState) {
		return nil, ErrNoTurnActive
	}
	id := factionState.CurrentTurn.FactionOrder[factionState.CurrentTurn.CurrentIndex]
	if faction, ok := factionState.Factions[id]; ok {
		return faction, nil
	}
	return nil, errors.New("faction not found: " + id)
}

// Advance marks the current faction's turn complete and moves to the next.
// Returns true when all factions have acted.
func (t *TurnEngine) Advance(factionState *state.FactionState) (bool, error) {
	if !t.InProgress(factionState) {
		return false, ErrNoTurnActive
	}

	factionState.CurrentTurn.CurrentIndex++
	factionState.CurrentTurn.Phase = domain.PhaseBookkeeping

	if factionState.CurrentTurn.CurrentIndex >= len(factionState.CurrentTurn.FactionOrder) {
		factionState.CurrentTurn = nil
		return true, nil
	}
	return false, nil
}

// Abandon clears the in-progress turn. Mutations already applied remain.
func (t *TurnEngine) Abandon(factionState *state.FactionState) {
	factionState.CurrentTurn = nil
}

// readyAllAssets flips Ready=true on every asset across all factions.
func readyAllAssets(factionState *state.FactionState) {
	for _, faction := range factionState.Factions {
		for _, asset := range faction.Assets {
			asset.Ready = true
		}
	}
}

package engine

import (
	"errors"
	"math/rand/v2"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

var (
	ErrTurnInProgress = errors.New("a turn is already in progress")
	ErrNoTurnActive   = errors.New("no turn is currently in progress")
	ErrNoFactions     = errors.New("no factions in campaign")
)

// BookkeepingResult captures what happened during a faction's bookkeeping phase.
// Carries enough detail for display and history. Mutations are returned
// separately by ApplyBookkeeping so the orchestrator can apply and record them
// uniformly with mutations from other pipeline steps.
type BookkeepingResult struct {
	IncomeGained       int        // total Coin added
	WealthIncome       int        // floor(Wealth/2) component
	StatIncome         int        // floor((Force+Cunning)/4) component
	AssetsLost         []AssetRef // destroyed due to second consecutive missed payment
	AssetsUnmaintained []AssetRef // newly unmaintained due to first missed payment
}

// AssetRef identifies an asset affected during bookkeeping.
type AssetRef struct {
	ID           string // for mutation: find asset in state
	DefinitionID string // for history: resolve name via Rulebook
	Location     string // for history: where the asset was
}

// TurnEngine manages turn lifecycle: ordering, bookkeeping computation, and
// pause/resume cursor state. Mutation application is owned by the orchestrator,
// not the TurnEngine.
type TurnEngine struct{}

func newTurnEngine(_ *MutationEngine) *TurnEngine {
	return &TurnEngine{}
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
		CycleNumber:   factionState.CycleNumber,
		FactionOrder: buildFactionOrder(factionState.Factions),
		CurrentIndex: 0,
		Phase:        domain.PhaseBookkeeping,
	}
	return nil
}

// readyAllAssets flips Ready=true on every asset across all factions. Called
// at cycle start so assets bought last cycle become active (SWN: "newly bought
// asset cannot attack...until the start of next turn").
func readyAllAssets(factionState *state.FactionState) {
	for _, faction := range factionState.Factions {
		for _, asset := range faction.Assets {
			asset.Ready = true
		}
	}
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

// ApplyBookkeeping computes income and maintenance mutations for the current
// faction and advances the turn phase to PhaseAction. It does not apply the
// mutations — the orchestrator owns that step. Safe to call on resume:
// returns a zero result and nil mutations if bookkeeping has already run.
func (t *TurnEngine) ApplyBookkeeping(factionState *state.FactionState) (BookkeepingResult, []domain.Mutation, error) {
	if !t.InProgress(factionState) {
		return BookkeepingResult{}, nil, ErrNoTurnActive
	}
	if factionState.CurrentTurn.Phase != domain.PhaseBookkeeping {
		return BookkeepingResult{}, nil, nil
	}

	f, err := t.CurrentFaction(factionState)
	if err != nil {
		return BookkeepingResult{}, nil, err
	}

	wealthIncome := f.Wealth / 2
	statIncome := (f.Force + f.Cunning) / 4
	total := wealthIncome + statIncome

	var mutations []domain.Mutation
	mutations = append(mutations, domain.CoinDelta{FactionID: f.ID, Delta: total, Cause: "bookkeeping"})

	result := applyMaintenance(f, f.Coin+total, &mutations)
	result.IncomeGained = total
	result.WealthIncome = wealthIncome
	result.StatIncome = statIncome

	factionState.CurrentTurn.Phase = domain.PhaseAction
	return result, mutations, nil
}

// Advance marks the current faction's turn complete and moves to the next.
// Returns true when all factions have acted. The caller is responsible for
// triggering the Mutation and History engines before clearing turn state.
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

// Abandon clears the in-progress turn. Mutations already applied to FactionState
// during this cycle remain — only the TurnState cursor is cleared.
func (t *TurnEngine) Abandon(factionState *state.FactionState) {
	factionState.CurrentTurn = nil
}

// applyMaintenance evaluates per-asset maintenance costs against startCoin (the
// simulated post-income balance: f.Coin + income total), appends the resulting
// mutations to the provided slice, and returns display-level asset events.
// Income fields are filled by the caller.
//
// Note: structured maintenance costs per asset definition are not yet modelled —
// maintenanceCost returns 0 for all assets until that data is added to AssetDefinition.
func applyMaintenance(f *domain.Faction, startCoin int, mutations *[]domain.Mutation) BookkeepingResult {
	var result BookkeepingResult
	runningCoin := startCoin

	for _, a := range f.Assets {
		cost := maintenanceCost(a)
		if cost == 0 {
			if !a.Maintained {
				*mutations = append(*mutations, domain.AssetMaintainedFlag{FactionID: f.ID, AssetID: a.ID, Maintained: true, Cause: "bookkeeping"})
			}
			continue
		}
		ref := AssetRef{ID: a.ID, DefinitionID: a.DefinitionID, Location: a.Location}
		if runningCoin >= cost {
			runningCoin -= cost
			*mutations = append(*mutations, domain.CoinDelta{FactionID: f.ID, Delta: -cost, Cause: "bookkeeping"})
			if !a.Maintained {
				*mutations = append(*mutations, domain.AssetMaintainedFlag{FactionID: f.ID, AssetID: a.ID, Maintained: true, Cause: "bookkeeping"})
			}
		} else if !a.Maintained {
			// second consecutive missed payment — asset is lost
			result.AssetsLost = append(result.AssetsLost, ref)
			*mutations = append(*mutations, domain.AssetRemoved{FactionID: f.ID, AssetID: a.ID, Cause: "bookkeeping"})
		} else {
			result.AssetsUnmaintained = append(result.AssetsUnmaintained, ref)
			*mutations = append(*mutations, domain.AssetMaintainedFlag{FactionID: f.ID, AssetID: a.ID, Maintained: false, Cause: "bookkeeping"})
		}
	}
	return result
}

// maintenanceCost returns the per-turn Coin cost for an asset. Returns 0 until
// maintenance cost data is added to AssetDefinition and resolved via Rulebook.
func maintenanceCost(_ *domain.Asset) int {
	return 0
}

// buildFactionOrder rolls a starting index and returns faction IDs as a rotation.
// Keys are sorted before randomizing so the rotation is deterministic given the same set.
func buildFactionOrder(factions map[string]*domain.Faction) []string {
	ids := make([]string, 0, len(factions))
	for id := range factions {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	n := len(ids)
	start := rand.IntN(n)
	order := make([]string, n)
	for i := range n {
		order[i] = ids[(start+i)%n]
	}
	return order
}

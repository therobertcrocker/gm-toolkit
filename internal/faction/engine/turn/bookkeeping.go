package turn

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// BookkeepingResult captures what happened during a faction's bookkeeping phase.
type BookkeepingResult struct {
	WealthIncome       int
	StatIncome         int
	AssetsLost         []AssetRef
	AssetsUnmaintained []AssetRef
}

// AssetRef identifies an asset affected during bookkeeping.
type AssetRef struct {
	ID           string
	DefinitionID string
	Location     string
}

// ApplyBookkeeping computes income and maintenance mutations for the current
// faction and advances the turn phase to PhaseAction. Returns a zero result
// and nil mutations if bookkeeping has already run (idempotent on resume).
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
	result.WealthIncome = wealthIncome
	result.StatIncome = statIncome

	factionState.CurrentTurn.Phase = domain.PhaseAction
	return result, mutations, nil
}

// applyMaintenance evaluates per-asset maintenance costs against startCoin,
// appends the resulting mutations, and returns display-level asset events.
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
// maintenance cost data is added to AssetDefinition.
func maintenanceCost(_ *domain.Asset) int {
	return 0
}

package turn

import (
	"log/slog"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
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
	Location     domain.Location
}

// ApplyBookkeeping computes income and maintenance mutations for the current
// faction and advances the turn phase to PhaseAction. Returns a zero result
// and nil mutations if bookkeeping has already run (idempotent on resume).
func (t *TurnEngine) ApplyBookkeeping(factionState *state.FactionState, registry *hooks.Registry, log *slog.Logger) (BookkeepingResult, []domain.Mutation, error) {
	if !t.InProgress(factionState) {
		return BookkeepingResult{}, nil, ErrNoTurnActive
	}
	if factionState.CurrentTurn.BookkeepingApplied {
		return BookkeepingResult{}, nil, nil
	}

	faction, err := t.CurrentFaction(factionState, log)
	if err != nil {
		return BookkeepingResult{}, nil, err
	}

	// Clear all hook budgets so each registered hook fires at most once this turn
	// (per "once per turn" rule wording). Hooks re-accumulate from scratch each turn.
	faction.HookBudgets = nil

	wealthIncome := faction.Wealth / 2
	statIncome := (faction.Force + faction.Cunning) / 4
	total := wealthIncome + statIncome

	var mutations []domain.Mutation
	mutations = append(mutations, domain.CoinDelta{FactionID: faction.ID, Delta: total, Cause: "bookkeeping"})

	result := applyMaintenance(registry, faction, faction.Coin+total, &mutations, t.rulebook)
	result.WealthIncome = wealthIncome
	result.StatIncome = statIncome

	factionState.CurrentTurn.BookkeepingApplied = true
	return result, mutations, nil
}

// applyMaintenance evaluates per-asset maintenance costs against startCoin,
// appends the resulting mutations, and returns display-level asset events.
func applyMaintenance(registry *hooks.Registry, faction *domain.Faction, startCoin int, mutations *[]domain.Mutation, rulebook *rulebook.Rulebook) BookkeepingResult {
	var result BookkeepingResult
	runningCoin := startCoin

	counts := make(map[domain.FactionStat]int)
	for _, asset := range domain.SortedAssets(faction) {
		category := rulebook.Assets[asset.DefinitionID].Category
		counts[category]++
	}

	surcharge := map[domain.FactionStat]int{
		domain.StatForce:   max(0, counts[domain.StatForce]-faction.Force),
		domain.StatCunning: max(0, counts[domain.StatCunning]-faction.Cunning),
		domain.StatWealth:  max(0, counts[domain.StatWealth]-faction.Wealth),
	}

	for _, asset := range domain.SortedAssets(faction) {
		cost := maintenanceCost(registry, faction, asset, rulebook)
		category := rulebook.Assets[asset.DefinitionID].Category
		if surcharge[category] > 0 {
			cost++
			surcharge[category]--
		}
		if cost == 0 {
			if !asset.Maintained {
				*mutations = append(*mutations, domain.AssetMaintainedFlag{FactionID: faction.ID, AssetID: asset.ID, Maintained: true, Cause: "bookkeeping"})
			}
			continue
		}
		ref := AssetRef{ID: asset.ID, DefinitionID: asset.DefinitionID, Location: asset.Location}
		if runningCoin >= cost {
			runningCoin -= cost
			*mutations = append(*mutations, domain.CoinDelta{FactionID: faction.ID, Delta: -cost, Cause: "bookkeeping"})
			if !asset.Maintained {
				*mutations = append(*mutations, domain.AssetMaintainedFlag{FactionID: faction.ID, AssetID: asset.ID, Maintained: true, Cause: "bookkeeping"})
			}
		} else if !asset.Maintained {
			result.AssetsLost = append(result.AssetsLost, ref)
			*mutations = append(*mutations, domain.AssetRemoved{FactionID: faction.ID, AssetID: asset.ID, Cause: "bookkeeping"})
		} else {
			result.AssetsUnmaintained = append(result.AssetsUnmaintained, ref)
			*mutations = append(*mutations, domain.AssetMaintainedFlag{FactionID: faction.ID, AssetID: asset.ID, Maintained: false, Cause: "bookkeeping"})
		}
	}
	return result
}

// maintenanceCost returns the per-turn Coin cost for an asset. Returns 0 as the
// base until maintenance cost data is added to AssetDefinition; registered
// MaintenanceCostModifiers may override via the hook registry.
func maintenanceCost(registry *hooks.Registry, faction *domain.Faction, asset *domain.Asset, rulebook *rulebook.Rulebook) int {
	return dispatch.ResolveMaintenanceCost(registry, faction, asset, rulebook.Assets[asset.DefinitionID].Maintenance)
}

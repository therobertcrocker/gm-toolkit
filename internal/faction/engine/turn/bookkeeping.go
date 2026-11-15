package turn

import (
	"log/slog"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
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

	runningCoin := faction.Coin + total
	applyUpkeep(registry, faction, &runningCoin, &mutations)
	result := applyMaintenance(faction, runningCoin, &mutations, t.rulebook)
	result.WealthIncome = wealthIncome
	result.StatIncome = statIncome

	factionState.CurrentTurn.BookkeepingApplied = true
	return result, mutations, nil
}

// applyUpkeep charges each upkeep a held tag declares against runningCoin, after
// income and before asset maintenance, so tags are paid first (A.001.5 Effort 2
// Decision 10). A paid upkeep lifts the tag's lapse; an unpaid one runs the entry's
// consequence: the tag lapses, the tag is lost, or the faction takes damage.
func applyUpkeep(registry *hooks.Registry, faction *domain.Faction, runningCoin *int, mutations *[]domain.Mutation) {
	if registry == nil {
		return
	}
	for _, registered := range registry.UpkeepChargersFor(faction.ID) {
		upkeep, due := registered.Hook.UpkeepDue(faction)
		if !due {
			continue
		}
		params := upkeep.Params
		lapsed := faction.HasLapsedTag(upkeep.TagID)
		if *runningCoin >= params.Amount {
			*runningCoin -= params.Amount
			*mutations = append(*mutations, domain.CoinDelta{FactionID: faction.ID, Delta: -params.Amount, Cause: "upkeep"})
			if lapsed {
				*mutations = append(*mutations, domain.TagLapsed{FactionID: faction.ID, TagID: upkeep.TagID, Lapsed: false, Cause: "upkeep", CausedByFactionID: faction.ID})
			}
			continue
		}
		switch params.OnUnpaid {
		case domain.UnpaidLapse:
			if !lapsed {
				*mutations = append(*mutations, domain.TagLapsed{FactionID: faction.ID, TagID: upkeep.TagID, Lapsed: true, Cause: "upkeep", CausedByFactionID: faction.ID})
			}
		case domain.UnpaidLoseTag:
			*mutations = append(*mutations, domain.TagRemoved{FactionID: faction.ID, TagID: upkeep.TagID, Cause: "upkeep", CausedByFactionID: faction.ID})
		case domain.UnpaidDamage:
			*mutations = append(*mutations, domain.FactionHPDelta{FactionID: faction.ID, Delta: -params.Damage, Cause: "upkeep", CausedByFactionID: faction.ID})
		}
	}
}

// applyMaintenance evaluates per-asset maintenance costs against startCoin,
// appends the resulting mutations, and returns display-level asset events.
func applyMaintenance(faction *domain.Faction, startCoin int, mutations *[]domain.Mutation, rulebook *rulebook.Rulebook) BookkeepingResult {
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
		cost := rulebook.Assets[asset.DefinitionID].Maintenance
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

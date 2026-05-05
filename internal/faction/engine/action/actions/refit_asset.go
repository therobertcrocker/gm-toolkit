package actions

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// RefitAsset swaps one asset for another of the same category. The faction
// pays the cost difference (minimum 0); the new asset starts at full HP and
// is inactive until the start of the next turn.
// Tech-level filtering and P-flag (government permission) checks are deferred.
type RefitAsset struct {
	collector  action.Collector
	factionID  string
	refitOrder action.RefitOrder
	newAsset   domain.Asset
	costDelta  int
}

func NewRefitAsset(collector action.Collector) *RefitAsset {
	return &RefitAsset{collector: collector}
}

func (ra *RefitAsset) Name() string { return "Refit Asset" }

func (ra *RefitAsset) Validate(faction *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) bool {
	for _, asset := range faction.Assets {
		if len(validRefitReplacements(faction, asset, rulebook)) > 0 {
			return true
		}
	}
	return false
}

func (ra *RefitAsset) Inputs(faction *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) error {
	var options []action.RefitOption
	for _, asset := range faction.Assets {
		replacements := validRefitReplacements(faction, asset, rulebook)
		if len(replacements) == 0 {
			continue
		}
		options = append(options, action.RefitOption{Asset: asset, Replacements: replacements})
	}

	order, err := ra.collector.SelectRefitOrder(options, rulebook)
	if err != nil {
		return fmt.Errorf("refit asset: %w", err)
	}
	ra.factionID = faction.ID
	ra.refitOrder = order
	return nil
}

func (ra *RefitAsset) Resolve(faction *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) error {
	oldDef, ok := rulebook.Assets[ra.refitOrder.OldAsset.DefinitionID]
	if !ok {
		return fmt.Errorf("asset definition not found: %s", ra.refitOrder.OldAsset.DefinitionID)
	}
	newDef := ra.refitOrder.NewDefinition

	delta := max(0, newDef.Cost-oldDef.Cost)
	if faction.Coin < delta {
		return fmt.Errorf("insufficient Coin: need %d, have %d", delta, faction.Coin)
	}

	ra.costDelta = delta
	ra.newAsset = domain.Asset{
		ID:           fmt.Sprintf("%s-%s-%d", ra.factionID, newDef.ID, nextAssetSuffix(faction, newDef)),
		DefinitionID: newDef.ID,
		OwnerID:      ra.factionID,
		Location:     ra.refitOrder.OldAsset.Location,
		CurrentHP:    newDef.HP,
		Ready:        false,
		Maintained:   true,
	}
	return nil
}

func (ra *RefitAsset) Output() ([]domain.Mutation, error) {
	mutations := []domain.Mutation{
		domain.AssetRemoved{FactionID: ra.factionID, AssetID: ra.refitOrder.OldAsset.ID, Cause: "refit", CausedByFactionID: ra.factionID},
		domain.AssetAdded{FactionID: ra.factionID, Asset: ra.newAsset, Cause: "refit", CausedByFactionID: ra.factionID},
	}
	if ra.costDelta > 0 {
		mutations = append(mutations, domain.CoinDelta{FactionID: ra.factionID, Delta: -ra.costDelta, Cause: "refit", CausedByFactionID: ra.factionID})
	}
	return mutations, nil
}

// validRefitReplacements returns definitions the faction could refit oldAsset
// into: same category, different definition, attribute meets MinRating, and
// faction can afford the cost delta.
func validRefitReplacements(faction *domain.Faction, oldAsset *domain.Asset, rulebook *rulebook.Rulebook) []*domain.AssetDefinition {
	oldDef, ok := rulebook.Assets[oldAsset.DefinitionID]
	if !ok {
		return nil
	}
	var result []*domain.AssetDefinition
	for _, def := range rulebook.Assets {
		if def.ID == oldDef.ID {
			continue
		}
		if def.Category != oldDef.Category {
			continue
		}
		if statScore(faction, def.Category) < def.MinRating {
			continue
		}
		delta := max(0, def.Cost-oldDef.Cost)
		if faction.Coin < delta {
			continue
		}
		result = append(result, def)
	}
	return result
}

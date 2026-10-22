package actions

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// RefitAsset swaps one asset for another of the same category. The faction
// pays the cost difference (minimum 0); the new asset starts at full HP and
// is inactive until the start of the next turn.
// A replacement that needs permission asks the world's government after the
// choice (D-23). Tech-level filtering is deferred.
type RefitAsset struct {
	collector  action.Collector
	registry   *hooks.Registry
	factionID  string
	refitOrder action.RefitOrder
	refusal    *domain.PermissionRefused
	newAsset   domain.Asset
	costDelta  int
}

func NewRefitAsset(collector action.Collector, registry *hooks.Registry) *RefitAsset {
	return &RefitAsset{collector: collector, registry: registry}
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

func (ra *RefitAsset) Inputs(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
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
	refusal, err := dispatch.RequestPermission(ra.registry, ra.collector, factionState, faction, order.NewDefinition, order.OldAsset.Location.WorldID, "refit")
	if err != nil {
		return fmt.Errorf("refit asset: %w", err)
	}
	ra.refusal = refusal
	return nil
}

func (ra *RefitAsset) Resolve(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	if ra.refusal != nil {
		return nil
	}
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
		ID:           domain.NextAssetID(factionState.Factions, faction, newDef.ID),
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
	if ra.refusal != nil {
		return []domain.Mutation{*ra.refusal}, nil
	}
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
	// deliberate: no refit for a sheltered asset, because the new asset starts at full HP and would skip the repair Boltholes asks for (Effort 2 Decision Record 5).
	if oldAsset.HasCondition(domain.ConditionSheltered) {
		return nil
	}
	// deliberate: no refit for an asset in transit, because an asset between worlds cannot be acted on (Effort 4 Decision 11; the general rule is F.028).
	if oldAsset.Location.WorldID == "" {
		return nil
	}
	oldDef, ok := rulebook.Assets[oldAsset.DefinitionID]
	if !ok {
		return nil
	}
	var result []*domain.AssetDefinition
	for _, def := range rulebook.Assets {
		if def.ID == oldDef.ID || def.Quality {
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

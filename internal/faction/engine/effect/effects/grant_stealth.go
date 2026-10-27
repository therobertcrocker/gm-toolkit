package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

// grantStealthRegistrar is Stealth, a quality bought for another asset (D-19 to
// D-21). Its filters select the asset it acts on (D-20): owner = "own" reads as "the
// buyer holds it", scope = "world" as "it stands on the world of the purchase", and
// asset_type and stealthy narrow it.
var grantStealthRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOnPurchase: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterAssetType, domain.FilterStealthy},
		},
	},
	Fire:    fireGrantStealth,
	Targets: grantStealthTargets,
}

// grantStealthTargets lists the buyer's assets on the world that the filters
// select, in ID order.
func grantStealthTargets(ctx effect.TargetContext) []*domain.Asset {
	if ctx.WorldID == "" {
		return nil
	}
	filters := ctx.Entry.Filters
	var targets []*domain.Asset
	for _, asset := range domain.SortedAssets(ctx.Faction) {
		if asset.Location.WorldID != ctx.WorldID {
			continue
		}
		if filters.Stealthy != nil && asset.Stealthy != *filters.Stealthy {
			continue
		}
		def := ctx.Rulebook.Assets[asset.DefinitionID]
		if def == nil || (len(filters.AssetTypes) > 0 && !slices.Contains(filters.AssetTypes, def.Type)) {
			continue
		}
		targets = append(targets, asset)
	}
	return targets
}

func fireGrantStealth(ctx effect.FireContext) []domain.Mutation {
	if ctx.Asset.Stealthy {
		return nil
	}
	return []domain.Mutation{domain.AssetStealthApplied{
		FactionID:         ctx.Holder.Faction.ID,
		AssetID:           ctx.Asset.ID,
		Cause:             string(domain.SpecialGrantStealth),
		CausedByFactionID: ctx.Holder.Faction.ID,
	}}
}

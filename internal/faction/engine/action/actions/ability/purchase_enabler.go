package ability

import (
	"fmt"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
)

// purchaseEnabler resolves Pretech Logistics: spend the action to buy one asset of the
// configured category (Force), with required tech level at or below the cap, on the
// acting asset's world, at a surcharge (half again, rounded up). The world's own tech
// level is ignored — the ability is precisely what lets the faction buy above it. The
// asset-cost chain then runs on the surcharged price. Emits AssetAdded (Ready:false) +
// the CoinDelta.
func purchaseEnabler(ctx StepContext) ([]domain.Mutation, error) {
	params := ctx.Step.Purchase
	if params == nil {
		return nil, fmt.Errorf("purchase_enabler: step has no purchase params")
	}
	worldID := ctx.Asset.Location.WorldID
	candidates := eligiblePurchaseDefs(ctx, params)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("purchase_enabler: no affordable %s assets at tech level %d or below", params.Category, params.MaxTechLevel)
	}
	options := make([]action.PurchaseEnablerOption, len(candidates))
	for i, candidate := range candidates {
		options[i] = action.PurchaseEnablerOption{Def: candidate, Cost: enablerQuote(ctx, candidate, params).Cost}
	}
	def, err := ctx.Collector.SelectPurchaseEnablerAsset(worldID, options)
	if err != nil {
		return nil, fmt.Errorf("purchase_enabler: %w", err)
	}
	if def == nil {
		return nil, fmt.Errorf("purchase_enabler: %w", action.ErrNoSelection)
	}
	quote := enablerQuote(ctx, def, params)
	cost := quote.Cost
	discountConsumed := false
	if ctx.Faction.PendingPurchaseDiscount > 0 {
		cost = dispatch.DiscountedCost(quote, def.Cost, ctx.Faction.PendingPurchaseDiscount)
		discountConsumed = true
	}
	if ctx.Faction.Coin < cost {
		return nil, fmt.Errorf("purchase_enabler: insufficient Coin: need %d, have %d", cost, ctx.Faction.Coin)
	}
	newAsset := domain.NewPurchasedAsset(ctx.Faction, def, worldID)
	mutations := []domain.Mutation{
		domain.AssetAdded{FactionID: ctx.Faction.ID, Asset: newAsset, Cause: "ability", CausedByFactionID: ctx.Faction.ID},
		domain.CoinDelta{FactionID: ctx.Faction.ID, Delta: -cost, Cause: "ability", CausedByFactionID: ctx.Faction.ID},
	}
	if discountConsumed {
		mutations = append(mutations, domain.PurchaseDiscountConsumed{FactionID: ctx.Faction.ID, Cause: "ability", CausedByFactionID: ctx.Faction.ID})
	}
	return mutations, nil
}

// eligiblePurchaseDefs returns the rulebook definitions an enabler purchase may buy:
// matching category, required tech level at or below the cap, the faction meeting the
// min rating, holding Planetary Government when the def is P-flagged (Decision 7), and
// affordable at the surcharged price. Name-sorted for a deterministic candidate order.
func eligiblePurchaseDefs(ctx StepContext, params *domain.PurchaseParams) []*domain.AssetDefinition {
	worldID := ctx.Asset.Location.WorldID
	government := holdsGovernment(ctx, worldID)
	var eligible []*domain.AssetDefinition
	for _, def := range ctx.Rulebook.Assets {
		if def.Category != params.Category {
			continue
		}
		terms := domain.PurchaseTerms{
			Cost:            enablerQuote(ctx, def, params).Cost,
			TechCap:         params.MaxTechLevel,
			HoldsGovernment: government,
		}
		if domain.CanPurchase(ctx.Faction, def, terms) {
			eligible = append(eligible, def)
		}
	}
	sort.Slice(eligible, func(i, j int) bool { return eligible[i].Name < eligible[j].Name })
	return eligible
}

// holdsGovernment guards the (possibly nil) world engine before delegating to
// domain.HasPlanetaryGovernment. Pretech Logistics waives the tech-level requirement
// but not the P-flag permission requirement.
func holdsGovernment(ctx StepContext, worldID string) bool {
	if ctx.WorldEngine == nil || ctx.WorldEngine.Index == nil {
		return false
	}
	return domain.HasPlanetaryGovernment(ctx.Faction.ID, ctx.WorldEngine.Index.BasesByLocation[worldID])
}

// enablerQuote runs the asset-cost chain on the enabler's surcharged price, so the
// same hooks that price a Buy Asset (Local Investments, Preceptor Archive) price
// an enabler purchase on the acting asset's world.
func enablerQuote(ctx StepContext, def *domain.AssetDefinition, params *domain.PurchaseParams) hooks.AssetCostQuery {
	worldID := ctx.Asset.Location.WorldID
	var onWorld []*domain.Asset
	if ctx.WorldEngine != nil && ctx.WorldEngine.Index != nil {
		onWorld = ctx.WorldEngine.Index.AssetsByLocation[worldID]
	}
	return dispatch.ResolveAssetCost(ctx.Registry, ctx.Faction, def, worldID, surchargedCost(def.Cost, params.CostPercent), onWorld)
}

// surchargedCost applies the enabler's cost surcharge: base * percent / 100, rounded up.
// 150 ("half again") on a 5-cost asset yields 8.
func surchargedCost(base, percent int) int {
	return (base*percent + 99) / 100
}

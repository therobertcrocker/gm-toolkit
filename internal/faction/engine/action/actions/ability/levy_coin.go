package ability

import (
	"fmt"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

const levyAmount = 1

// levyCoin is Monopoly: the target faction pays the owner 1 Coin. A faction that
// cannot pay loses one asset of its choice on the world instead, from all its
// assets there, stealthed ones too (Effort 3 Decision 11), but not a sheltered one
// (Effort 2 Decision 11).
func levyCoin(ctx StepContext) ([]domain.Mutation, error) {
	target := ctx.Target.Faction
	if target.Coin >= levyAmount {
		// deliberate: the owner gains the Coin even when a Bank absorbs the target's loss, not only when the loss stands, as coin_steal does (Robert's call, Effort 3 Decision Record 8).
		return []domain.Mutation{
			domain.CoinDelta{FactionID: target.ID, Delta: -levyAmount, Cause: "ability", CausedByFactionID: ctx.Faction.ID},
			domain.CoinDelta{FactionID: ctx.Faction.ID, Delta: levyAmount, Cause: "ability", CausedByFactionID: ctx.Faction.ID},
		}, nil
	}
	var onWorld []*domain.Asset
	for _, asset := range domain.SortedAssets(target) {
		if asset.Location.WorldID == ctx.Asset.Location.WorldID && !asset.HasCondition(domain.ConditionSheltered) {
			onWorld = append(onWorld, asset)
		}
	}
	lost, err := ctx.Collector.SelectAssetToLose(target, onWorld, ctx.Rulebook)
	if err != nil {
		return nil, fmt.Errorf("levy_coin: %w", err)
	}
	if lost == nil {
		// deliberate: a plain error, not ErrNoSelection, because the book gives the target no way out of the loss (Effort 3 Decision Record 5).
		return nil, fmt.Errorf("levy_coin: %s must choose an asset to lose", target.ID)
	}
	if !slices.ContainsFunc(onWorld, func(asset *domain.Asset) bool { return asset.ID == lost.ID }) {
		return nil, fmt.Errorf("levy_coin: %s chose %s, which is not on %s", target.ID, lost.ID, ctx.Asset.Location.WorldID)
	}
	return []domain.Mutation{domain.AssetRemoved{
		FactionID:         target.ID,
		AssetID:           lost.ID,
		Cause:             string(domain.EffectLevyCoin),
		CausedByFactionID: ctx.Faction.ID,
	}}, nil
}

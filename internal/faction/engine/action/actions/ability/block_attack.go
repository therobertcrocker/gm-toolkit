package ability

import (
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// blockAttack is Seditionists: pay the rolled cost to attach to the target, which
// cannot attack while the attachment holds (domain.AttackBlocked). Attaching moves
// the attachment: this source's no_attack leaves every other asset. A faction that
// cannot pay the roll attaches nothing and pays nothing.
func blockAttack(ctx StepContext) ([]domain.Mutation, error) {
	cost := ctx.RollResult
	if ctx.Faction.Coin < cost {
		return nil, nil
	}
	condition := domain.Condition{Kind: domain.ConditionNoAttack, SourceAssetID: ctx.Asset.ID}
	mutations := []domain.Mutation{domain.CoinDelta{
		FactionID:         ctx.Faction.ID,
		Delta:             -cost,
		Cause:             "ability",
		CausedByFactionID: ctx.Faction.ID,
	}}
	for _, holderID := range slices.Sorted(maps.Keys(ctx.FactionState.Factions)) {
		for _, held := range domain.SortedAssets(ctx.FactionState.Factions[holderID]) {
			if held.ID != ctx.Target.Asset.ID && slices.ContainsFunc(held.Conditions, condition.Is) {
				mutations = append(mutations, domain.AssetConditionCleared{
					FactionID:         holderID,
					AssetID:           held.ID,
					Condition:         condition,
					Cause:             "ability",
					CausedByFactionID: ctx.Faction.ID,
				})
			}
		}
	}
	return append(mutations, domain.AssetConditionApplied{
		FactionID:         ctx.Target.Faction.ID,
		AssetID:           ctx.Target.Asset.ID,
		Condition:         condition,
		Cause:             "ability",
		CausedByFactionID: ctx.Faction.ID,
	}), nil
}

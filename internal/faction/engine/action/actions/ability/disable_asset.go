package ability

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// disableAsset is Marketers: the target faction pays the target asset's half
// price, or the asset is disabled until that price is paid. A faction that cannot
// pay takes the condition with no prompt.
func disableAsset(ctx StepContext) ([]domain.Mutation, error) {
	targetFaction, target := ctx.Target.Faction, ctx.Target.Asset
	def, ok := ctx.Rulebook.Assets[target.DefinitionID]
	if !ok {
		return nil, fmt.Errorf("disable_asset: definition not found: %s", target.DefinitionID)
	}
	price := def.HalfPrice()
	if targetFaction.Coin >= price {
		paid, err := ctx.Collector.ConfirmPayment(hooks.PaymentOffer{
			Payer:  targetFaction,
			Asset:  target,
			Amount: price,
			Reason: hooks.PaymentAvoidDisable,
		})
		if err != nil {
			return nil, fmt.Errorf("disable_asset: %w", err)
		}
		if paid {
			return []domain.Mutation{domain.CoinDelta{
				FactionID:         targetFaction.ID,
				Delta:             -price,
				Cause:             "ability",
				CausedByFactionID: ctx.Faction.ID,
			}}, nil
		}
	}
	return []domain.Mutation{domain.AssetConditionApplied{
		FactionID:         targetFaction.ID,
		AssetID:           target.ID,
		Condition:         domain.Condition{Kind: domain.ConditionDisabled, SourceAssetID: ctx.Asset.ID},
		Cause:             "ability",
		CausedByFactionID: ctx.Faction.ID,
	}}, nil
}

package ability

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// coinDelta resolves a coin ability. A target roll pays its amount on a pass. On a
// table, the row the die picked gives a signed CoinDelta: a row flagged DestroySelf
// destroys the acting asset unconditionally, and a negative payout the faction cannot
// afford also destroys it (the SWN insolvency rule). An unmatched roll pays nothing.
func coinDelta(ctx StepContext) ([]domain.Mutation, error) {
	if ctx.Step.TargetRoll != nil {
		if !ctx.Rolled.Passed {
			return nil, nil
		}
		return []domain.Mutation{coinPayout(ctx, ctx.Step.TargetRoll.Amount)}, nil
	}
	if ctx.Step.Table == nil {
		return nil, fmt.Errorf("coin_delta: step has no roll table or target roll")
	}
	outcome := ctx.Rolled.Row
	if outcome == nil {
		return nil, nil
	}
	if outcome.DestroySelf || (outcome.Coin < 0 && ctx.Faction.Coin < -outcome.Coin) {
		return []domain.Mutation{domain.AssetRemoved{
			FactionID:         ctx.Faction.ID,
			AssetID:           ctx.Asset.ID,
			Cause:             "ability",
			CausedByFactionID: ctx.Faction.ID,
		}}, nil
	}
	if outcome.Coin != 0 {
		return []domain.Mutation{coinPayout(ctx, outcome.Coin)}, nil
	}
	return nil, nil
}

func coinPayout(ctx StepContext, amount int) domain.Mutation {
	return domain.CoinDelta{
		FactionID:         ctx.Faction.ID,
		Delta:             amount,
		CausedByFactionID: ctx.Faction.ID,
		Cause:             "ability",
	}
}

package ability

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// coinDelta resolves a roll-table coin ability: the row the die picked gives a signed
// CoinDelta. A row flagged DestroySelf destroys the acting asset unconditionally; a
// negative payout the faction cannot afford also destroys it (the SWN insolvency
// rule). An unmatched roll pays nothing.
func coinDelta(ctx StepContext) ([]domain.Mutation, error) {
	if ctx.Step.Table == nil {
		return nil, fmt.Errorf("coin_delta: step has no roll table")
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
		return []domain.Mutation{domain.CoinDelta{
			FactionID:         ctx.Faction.ID,
			Delta:             outcome.Coin,
			CausedByFactionID: ctx.Faction.ID,
			Cause:             "ability",
		}}, nil
	}
	return nil, nil
}

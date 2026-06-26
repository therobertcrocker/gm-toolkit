package ability

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// purchaseDiscount resolves a roll-table discount ability (Commodities Broker): the
// rolled result is the discount on the faction's next purchase. The half-price floor
// is applied at purchase time (BuyAsset), not here. The discount equals the roll, so
// there is no Outcomes band to map through — ctx.RollResult is granted directly.
func purchaseDiscount(ctx StepContext) ([]domain.Mutation, error) {
	if ctx.Step.Roll == nil {
		return nil, fmt.Errorf("purchase_discount: step has no roll table")
	}
	if ctx.RollResult <= 0 {
		return nil, nil
	}
	return []domain.Mutation{domain.PurchaseDiscountGranted{
		FactionID:         ctx.Faction.ID,
		Amount:            ctx.RollResult,
		Cause:             "ability",
		CausedByFactionID: ctx.Faction.ID,
	}}, nil
}

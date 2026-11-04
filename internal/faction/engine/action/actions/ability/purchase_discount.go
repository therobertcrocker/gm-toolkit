package ability

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// purchaseDiscount resolves an amount-roll discount ability (Commodities Broker): the
// rolled amount is the discount on the faction's next purchase. The half-price floor
// is applied at purchase time (BuyAsset), not here.
func purchaseDiscount(ctx StepContext) ([]domain.Mutation, error) {
	if ctx.Step.Amount == nil {
		return nil, fmt.Errorf("purchase_discount: step has no amount roll")
	}
	if ctx.Rolled.Amount <= 0 {
		return nil, nil
	}
	return []domain.Mutation{domain.PurchaseDiscountGranted{
		FactionID:         ctx.Faction.ID,
		Amount:            ctx.Rolled.Amount,
		Cause:             "ability",
		CausedByFactionID: ctx.Faction.ID,
	}}, nil
}

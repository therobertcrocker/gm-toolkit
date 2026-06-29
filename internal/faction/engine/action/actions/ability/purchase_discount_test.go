package ability

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func TestPurchaseDiscount_GrantsRolledAmount(t *testing.T) {
	ctx := StepContext{
		Faction: &domain.Faction{ID: "f1"},
		Asset:   &domain.Asset{ID: "a1", OwnerID: "f1"},
		Step: domain.AbilityStep{
			Effect: domain.EffectPurchaseDiscount,
			Target: domain.TargetNone,
			Roll:   &domain.RollTable{Die: domain.DiceRoll{NumDice: 1, Sides: 8}},
		},
		RollResult: 5,
	}
	muts, err := purchaseDiscount(ctx)
	if err != nil {
		t.Fatalf("purchaseDiscount: %v", err)
	}
	if len(muts) != 1 {
		t.Fatalf("len(muts) = %d, want 1", len(muts))
	}
	granted, ok := muts[0].(domain.PurchaseDiscountGranted)
	if !ok {
		t.Fatalf("muts[0] type = %T, want PurchaseDiscountGranted", muts[0])
	}
	if granted.Amount != 5 || granted.FactionID != "f1" {
		t.Errorf("granted = %+v, want {f1, 5}", granted)
	}
}

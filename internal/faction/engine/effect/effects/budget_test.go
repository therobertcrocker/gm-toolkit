package effects

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func TestBudgetKey(t *testing.T) {
	cases := []struct {
		oncePer domain.OncePerKey
		want    string
	}{
		{domain.OncePerTargetFaction, "effect:coin_steal:DEF:beta"},
		{domain.OncePerOwnerFaction, "effect:coin_steal:DEF"},
		{domain.OncePerAsset, "effect:coin_steal:DEF:asset-1"},
	}
	for _, tc := range cases {
		if got := budgetKey(tc.oncePer, domain.SpecialCoinSteal, "DEF", "asset-1", "beta"); got != tc.want {
			t.Errorf("budgetKey(%s) = %q, want %q", tc.oncePer, got, tc.want)
		}
	}
}

func TestSpendBudget_OwnerFactionSharesOneBudgetAcrossCopies(t *testing.T) {
	owner := &domain.Faction{ID: "alpha"}
	first := budgetKey(domain.OncePerOwnerFaction, domain.SpecialCoinSteal, "DEF", "asset-1", "beta")
	second := budgetKey(domain.OncePerOwnerFaction, domain.SpecialCoinSteal, "DEF", "asset-2", "gamma")

	if !spendBudget(owner, first) {
		t.Fatal("first spend = false, want true")
	}
	if spendBudget(owner, second) {
		t.Error("second copy's spend = true, want false: owner_faction is one budget per faction")
	}
}

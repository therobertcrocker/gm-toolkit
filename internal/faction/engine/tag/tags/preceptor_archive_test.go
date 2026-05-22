package tags

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func TestPreceptorArchiveCostModifier_TL4PlusReducesCost(t *testing.T) {
	modifier := &PreceptorArchiveCostModifier{}
	buyer := &domain.Faction{}

	cases := []struct {
		tl       int
		baseCost int
		wantCost int
	}{
		{tl: 4, baseCost: 4, wantCost: 3},
		{tl: 5, baseCost: 6, wantCost: 5},
		{tl: 4, baseCost: 1, wantCost: 0},
	}
	for _, tc := range cases {
		def := &domain.AssetDefinition{TechLevel: tc.tl, Cost: tc.baseCost}
		got := modifier.ModifyAssetCost(buyer, def, "Tartarus", tc.baseCost)
		if got != tc.wantCost {
			t.Errorf("TL%d baseCost=%d: got %d, want %d", tc.tl, tc.baseCost, got, tc.wantCost)
		}
	}
}

func TestPreceptorArchiveCostModifier_BelowTL4NoChange(t *testing.T) {
	modifier := &PreceptorArchiveCostModifier{}
	buyer := &domain.Faction{}

	for _, tl := range []int{0, 1, 2, 3} {
		def := &domain.AssetDefinition{TechLevel: tl, Cost: 4}
		got := modifier.ModifyAssetCost(buyer, def, "Tartarus", 4)
		if got != 4 {
			t.Errorf("TL%d: got %d, want 4 (no change)", tl, got)
		}
	}
}

func TestPreceptorArchiveCostModifier_ZeroCostNoChange(t *testing.T) {
	modifier := &PreceptorArchiveCostModifier{}
	buyer := &domain.Faction{}

	def := &domain.AssetDefinition{TechLevel: 4, Cost: 0}
	got := modifier.ModifyAssetCost(buyer, def, "Tartarus", 0)
	if got != 0 {
		t.Errorf("zero cost TL4: got %d, want 0", got)
	}
}

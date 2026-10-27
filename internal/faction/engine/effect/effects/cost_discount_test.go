package effects_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
)

func TestCostDiscount_TagLowersTheOwnersTechLevelPurchases(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-013"}}
	beta := &domain.Faction{ID: "beta"}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialCostDiscount].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "SWN-T-013", Name: "Preceptor Archive"}},
		Entry: domain.EffectDefinition{
			Trigger: domain.TriggerAssetCostQuery, Effect: domain.SpecialCostDiscount,
			Filters:      domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, MinTechLevel: 4},
			CostDiscount: &domain.CostDiscountParams{Amount: 1},
		},
		Registry: registry,
	})
	tl4 := &domain.AssetDefinition{ID: "tl4", Cost: 4, TechLevel: 4}
	tl3 := &domain.AssetDefinition{ID: "tl3", Cost: 4, TechLevel: 3}
	free := &domain.AssetDefinition{ID: "free", Cost: 0, TechLevel: 5}

	cases := []struct {
		name  string
		buyer *domain.Faction
		def   *domain.AssetDefinition
		want  int
	}{
		{"TL4 purchase", alpha, tl4, 3},
		{"TL3 purchase", alpha, tl3, 4},
		{"never below 0", alpha, free, 0},
		{"a rival's purchase", beta, tl4, 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := dispatch.ResolveAssetCost(registry, tc.buyer, tc.def, "Tartarus", tc.def.Cost, nil).Cost; got != tc.want {
				t.Errorf("cost = %d, want %d", got, tc.want)
			}
		})
	}

	alpha.Tags = nil
	if got := dispatch.ResolveAssetCost(registry, alpha, tl4, "Tartarus", 4, nil).Cost; got != 4 {
		t.Errorf("dropped tag: cost = %d, want 4", got)
	}
}

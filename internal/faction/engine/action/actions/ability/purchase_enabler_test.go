package ability

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"go.uber.org/mock/gomock"
)

func enablerStep() domain.AbilityStep {
	return domain.AbilityStep{
		Effect:   domain.EffectPurchaseEnabler,
		Target:   domain.TargetNone,
		Purchase: &domain.PurchaseParams{Category: domain.StatForce, MaxTechLevel: 5, CostPercent: 150},
	}
}

func enablerRulebook() *rulebook.Rulebook {
	return &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"F-cheap":   {ID: "F-cheap", Name: "Aaa Cheap Force", Category: domain.StatForce, MinRating: 1, Cost: 4, HP: 6, TechLevel: 3},
		"F-pricey":  {ID: "F-pricey", Name: "Bbb Pricey Force", Category: domain.StatForce, MinRating: 1, Cost: 5, HP: 8, TechLevel: 5},
		"F-toohigh": {ID: "F-toohigh", Name: "Ccc High TL Force", Category: domain.StatForce, MinRating: 1, Cost: 4, HP: 6, TechLevel: 6},
		"F-rating":  {ID: "F-rating", Name: "Ddd High Rating Force", Category: domain.StatForce, MinRating: 9, Cost: 4, HP: 6, TechLevel: 2},
		"W-wealth":  {ID: "W-wealth", Name: "Eee Wealth", Category: domain.StatWealth, MinRating: 1, Cost: 4, HP: 6, TechLevel: 2},
		"F-perm":    {ID: "F-perm", Name: "Fff Permission Force", Category: domain.StatForce, MinRating: 1, Cost: 4, HP: 6, TechLevel: 4, Flags: []domain.AssetFlag{domain.FlagPermission}},
	}}
}

func TestPurchaseEnabler_BuysSurcharged(t *testing.T) {
	ctrl := gomock.NewController(t)
	var offered []*domain.AssetDefinition

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().
		SelectPurchaseEnablerAsset(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ string, candidates []*domain.AssetDefinition) (*domain.AssetDefinition, error) {
			offered = candidates
			for _, def := range candidates {
				if def.ID == "F-pricey" {
					return def, nil
				}
			}
			return nil, nil
		})

	ctx := StepContext{
		Faction:   &domain.Faction{ID: "f1", Force: 5, Coin: 20, Assets: map[string]*domain.Asset{}},
		Asset:     &domain.Asset{ID: "a1", OwnerID: "f1", Location: domain.Location{WorldID: "w1"}},
		Step:      enablerStep(),
		Rulebook:  enablerRulebook(),
		Collector: collector,
	}
	muts, err := purchaseEnabler(ctx)
	if err != nil {
		t.Fatalf("purchaseEnabler: %v", err)
	}

	// candidate filter: wrong category, too-high TL, too-high rating, and the P-flagged
	// def (no WorldEngine, so no Planetary Government) all excluded.
	if len(offered) != 2 {
		t.Fatalf("offered %d candidates, want 2 (F-cheap, F-pricey); got %v", len(offered), offered)
	}

	var added *domain.AssetAdded
	var coin *domain.CoinDelta
	for _, m := range muts {
		switch v := m.(type) {
		case domain.AssetAdded:
			added = &v
		case domain.CoinDelta:
			coin = &v
		default:
			t.Errorf("unexpected mutation %T", m)
		}
	}
	if added == nil || coin == nil {
		t.Fatalf("want AssetAdded + CoinDelta, got %v", muts)
	}
	if added.Asset.DefinitionID != "F-pricey" || added.Asset.Location.WorldID != "w1" || added.Asset.Ready {
		t.Errorf("new asset = %+v, want F-pricey on w1, Ready=false", added.Asset)
	}
	if coin.Delta != -8 { // ceil(5 * 1.5) = 8
		t.Errorf("CoinDelta.Delta = %d, want -8", coin.Delta)
	}
}

func TestPurchaseEnabler_InsufficientCoin(t *testing.T) {
	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	// F-cheap surcharge is ceil(4*1.5)=6 > 5; F-pricey is ceil(5*1.5)=8 > 5.
	// Both unaffordable at the filter stage — SelectPurchaseEnablerAsset never called.
	ctx := StepContext{
		Faction:   &domain.Faction{ID: "f1", Force: 5, Coin: 5, Assets: map[string]*domain.Asset{}},
		Asset:     &domain.Asset{ID: "a1", OwnerID: "f1", Location: domain.Location{WorldID: "w1"}},
		Step:      enablerStep(),
		Rulebook:  enablerRulebook(),
		Collector: collector,
	}
	if _, err := purchaseEnabler(ctx); err == nil {
		t.Fatal("want error on no affordable candidates")
	}
}

func TestPurchaseEnabler_PermissionGate(t *testing.T) {
	ctrl := gomock.NewController(t)
	var offered []*domain.AssetDefinition

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().
		SelectPurchaseEnablerAsset(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ string, candidates []*domain.AssetDefinition) (*domain.AssetDefinition, error) {
			offered = candidates
			return candidates[0], nil
		}).
		Times(2)

	// A base owned by the acting faction on w1 = Planetary Government, unlocking F-perm.
	ctx := StepContext{
		Faction:     &domain.Faction{ID: "f1", Force: 5, Coin: 20, Assets: map[string]*domain.Asset{}},
		Asset:       &domain.Asset{ID: "a1", OwnerID: "f1", Location: domain.Location{WorldID: "w1"}},
		Step:        enablerStep(),
		Rulebook:    enablerRulebook(),
		Collector:   collector,
		WorldEngine: &world.WorldEngine{Index: &world.Index{BasesByLocation: map[string][]*domain.Base{"w1": {{OwnerID: "f1"}}}}},
	}
	if _, err := purchaseEnabler(ctx); err != nil {
		t.Fatalf("purchaseEnabler with gov: %v", err)
	}
	// F-cheap, F-pricey, and F-perm (government held) — 3 candidates.
	if len(offered) != 3 {
		t.Fatalf("with Planetary Government, offered %d candidates, want 3 (incl F-perm); got %v", len(offered), offered)
	}

	// Rival base only — F-perm stays excluded.
	ctx.WorldEngine = &world.WorldEngine{Index: &world.Index{BasesByLocation: map[string][]*domain.Base{"w1": {{OwnerID: "rival"}}}}}
	if _, err := purchaseEnabler(ctx); err != nil {
		t.Fatalf("purchaseEnabler without gov: %v", err)
	}
	if len(offered) != 2 {
		t.Fatalf("without Planetary Government, offered %d candidates, want 2; got %v", len(offered), offered)
	}
}

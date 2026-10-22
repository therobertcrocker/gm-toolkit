package ability

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
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
	var offered []action.PurchaseEnablerOption

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().
		SelectPurchaseEnablerAsset(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ string, candidates []action.PurchaseEnablerOption) (*domain.AssetDefinition, error) {
			offered = candidates
			for _, option := range candidates {
				if option.Def.ID == "F-pricey" {
					return option.Def, nil
				}
			}
			return nil, nil
		})

	ctx := StepContext{
		Faction:      &domain.Faction{ID: "f1", Force: 5, Coin: 20, Assets: map[string]*domain.Asset{}},
		Asset:        &domain.Asset{ID: "a1", OwnerID: "f1", Location: domain.Location{WorldID: "w1"}},
		Step:         enablerStep(),
		FactionState: &state.FactionState{},
		Rulebook:     enablerRulebook(),
		Collector:    collector,
	}
	muts, err := purchaseEnabler(ctx)
	if err != nil {
		t.Fatalf("purchaseEnabler: %v", err)
	}

	// candidate filter: wrong category, too-high TL and too-high rating excluded; the
	// P-flagged def is offered and asks after the pick (D-35).
	if len(offered) != 3 {
		t.Fatalf("offered %d candidates, want 3 (F-cheap, F-pricey, F-perm); got %v", len(offered), offered)
	}
	// C5: the engine computes the surcharge and passes it through; the view must not
	// recompute. Each option carries surchargedCost(def.Cost, 150).
	for _, option := range offered {
		if want := surchargedCost(option.Def.Cost, 150); option.Cost != want {
			t.Errorf("option %s Cost = %d, want %d", option.Def.ID, option.Cost, want)
		}
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

func TestPurchaseEnabler_ConsumesPendingDiscount(t *testing.T) {
	selectPricey := func(ctrl *gomock.Controller) *mocks.MockCollector {
		collector := mocks.NewMockCollector(ctrl)
		collector.EXPECT().
			SelectPurchaseEnablerAsset(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ string, candidates []action.PurchaseEnablerOption) (*domain.AssetDefinition, error) {
				for _, option := range candidates {
					if option.Def.ID == "F-pricey" {
						return option.Def, nil
					}
				}
				return nil, nil
			})
		return collector
	}

	enablerCtx := func(collector action.Collector, discount int) StepContext {
		return StepContext{
			Faction:      &domain.Faction{ID: "f1", Force: 5, Coin: 20, PendingPurchaseDiscount: discount, Assets: map[string]*domain.Asset{}},
			Asset:        &domain.Asset{ID: "a1", OwnerID: "f1", Location: domain.Location{WorldID: "w1"}},
			Step:         enablerStep(),
			FactionState: &state.FactionState{},
			Rulebook:     enablerRulebook(),
			Collector:    collector,
		}
	}

	// F-pricey surcharge = ceil(5*1.5) = 8; a 3-Coin discount nets 5. The discount is
	// consumed (PurchaseDiscountConsumed emitted), mirroring Buy Asset.
	t.Run("applies and consumes", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		muts, err := purchaseEnabler(enablerCtx(selectPricey(ctrl), 3))
		if err != nil {
			t.Fatalf("purchaseEnabler: %v", err)
		}
		var coin *domain.CoinDelta
		consumed := false
		for _, m := range muts {
			switch v := m.(type) {
			case domain.CoinDelta:
				coin = &v
			case domain.PurchaseDiscountConsumed:
				consumed = true
			}
		}
		if coin == nil || coin.Delta != -5 {
			t.Errorf("CoinDelta = %v, want -5 (8 surcharge - 3 discount)", coin)
		}
		if !consumed {
			t.Error("want PurchaseDiscountConsumed emitted")
		}
	})

	// A discount larger than the surcharge clamps to half base price (5/2 = 2), never
	// raising the price back up.
	t.Run("clamps to half base floor", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		muts, err := purchaseEnabler(enablerCtx(selectPricey(ctrl), 10))
		if err != nil {
			t.Fatalf("purchaseEnabler: %v", err)
		}
		for _, m := range muts {
			if coin, ok := m.(domain.CoinDelta); ok && coin.Delta != -2 {
				t.Errorf("CoinDelta = %d, want -2 (floor at half base)", coin.Delta)
			}
		}
	})
}

// stubSurcharge raises the cost by one, as cost_surcharge does.
type stubSurcharge struct{}

func (stub *stubSurcharge) ModifyAssetCost(_ *domain.Faction, _ *domain.AssetDefinition, _ string, query *hooks.AssetCostQuery) {
	query.Cost++
	query.Surcharges = append(query.Surcharges, hooks.Surcharge{FactionID: "rival", Amount: 1})
}

// TestPurchaseEnabler_PaysRivalSurcharge: a rival's surcharge on the acting asset's
// world lands on top of the enabler's own 150% price, and survives a discount.
func TestPurchaseEnabler_PaysRivalSurcharge(t *testing.T) {
	cases := map[string]struct {
		discount  int
		wantDelta int
	}{
		"no discount":           {discount: 0, wantDelta: -9},  // ceil(5*1.5) + 1
		"discount floors below": {discount: 10, wantDelta: -3}, // half base 2 + 1
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			var offered []action.PurchaseEnablerOption
			collector := mocks.NewMockCollector(ctrl)
			collector.EXPECT().
				SelectPurchaseEnablerAsset(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ string, candidates []action.PurchaseEnablerOption) (*domain.AssetDefinition, error) {
					offered = candidates
					for _, option := range candidates {
						if option.Def.ID == "F-pricey" {
							return option.Def, nil
						}
					}
					return nil, nil
				})
			registry := hooks.NewRegistry()
			registry.RegisterAssetCostModifier(hooks.AssetScope("rival", "li"), "surcharge", &stubSurcharge{})
			rivalAsset := &domain.Asset{ID: "li", OwnerID: "rival", Location: domain.Location{WorldID: "w1"}}
			ctx := StepContext{
				Faction:      &domain.Faction{ID: "f1", Force: 5, Coin: 20, PendingPurchaseDiscount: testCase.discount, Assets: map[string]*domain.Asset{}},
				Asset:        &domain.Asset{ID: "a1", OwnerID: "f1", Location: domain.Location{WorldID: "w1"}},
				Step:         enablerStep(),
				FactionState: &state.FactionState{},
				Rulebook:     enablerRulebook(),
				Collector:    collector,
				WorldEngine:  &world.WorldEngine{Index: &world.Index{AssetsByLocation: map[string][]*domain.Asset{"w1": {rivalAsset}}}},
				Registry:     registry,
			}
			muts, err := purchaseEnabler(ctx)
			if err != nil {
				t.Fatalf("purchaseEnabler: %v", err)
			}
			for _, option := range offered {
				if want := surchargedCost(option.Def.Cost, 150) + 1; option.Cost != want {
					t.Errorf("option %s Cost = %d, want %d", option.Def.ID, option.Cost, want)
				}
			}
			total := 0
			var surcharge *domain.CoinDelta
			for _, m := range muts {
				if coin, ok := m.(domain.CoinDelta); ok {
					total += coin.Delta
					if coin.CausedByFactionID == "rival" {
						surcharge = &coin
					}
				}
			}
			if total != testCase.wantDelta {
				t.Errorf("total charged = %d, want %d", total, testCase.wantDelta)
			}
			if surcharge == nil || surcharge.Delta != -1 || surcharge.Cause != "cost_surcharge" {
				t.Errorf("surcharge charge = %+v, want its own -1 cost_surcharge caused by the rival", surcharge)
			}
		})
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

func TestPurchaseEnabler_AsksPermissionAfterThePick(t *testing.T) {
	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().
		SelectPurchaseEnablerAsset(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ string, candidates []action.PurchaseEnablerOption) (*domain.AssetDefinition, error) {
			for _, candidate := range candidates {
				if candidate.Def.ID == "F-perm" {
					return candidate.Def, nil
				}
			}
			t.Fatal("F-perm not offered")
			return nil, nil
		})
	collector.EXPECT().RequestPermission(gomock.Any()).Return(false, nil)

	ctx := StepContext{
		Faction:      &domain.Faction{ID: "f1", Force: 5, Coin: 20, Assets: map[string]*domain.Asset{}},
		Asset:        &domain.Asset{ID: "a1", OwnerID: "f1", Location: domain.Location{WorldID: "w1"}},
		Step:         enablerStep(),
		FactionState: &state.FactionState{},
		Rulebook:     enablerRulebook(),
		Collector:    collector,
	}
	mutations, err := purchaseEnabler(ctx)
	if err != nil {
		t.Fatalf("purchaseEnabler: %v", err)
	}
	want := []domain.Mutation{domain.PermissionRefused{FactionID: "f1", WorldID: "w1", DefinitionID: "F-perm", Cause: "ability", CausedByFactionID: "f1"}}
	if !reflect.DeepEqual(mutations, want) {
		t.Errorf("mutations = %+v, want only the refusal record (no asset, no charge)", mutations)
	}
}

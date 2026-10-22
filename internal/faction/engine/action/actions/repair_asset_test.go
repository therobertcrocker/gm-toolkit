package actions

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"go.uber.org/mock/gomock"
)

func makeRepairAssetRulebook() *rulebook.Rulebook {
	return &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			"tough": {ID: "tough", Name: "Infantry", Category: domain.StatForce, HP: 10, Cost: 4},
		},
	}
}

func TestRepairAsset_Validate(t *testing.T) {
	rulebook := makeRepairAssetRulebook()

	t.Run("no damaged assets", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		asset := &domain.Asset{ID: "a1", DefinitionID: "tough", CurrentHP: 10}
		faction := &domain.Faction{ID: "f1", Coin: 5, Assets: map[string]*domain.Asset{"a1": asset}}
		if NewRepairAsset(mocks.NewMockCollector(ctrl), nil).Validate(faction, nil, rulebook) {
			t.Error("expected false when no assets are damaged")
		}
	})

	t.Run("damaged but no coin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		asset := &domain.Asset{ID: "a1", DefinitionID: "tough", CurrentHP: 6}
		faction := &domain.Faction{ID: "f1", Coin: 0, Assets: map[string]*domain.Asset{"a1": asset}}
		if NewRepairAsset(mocks.NewMockCollector(ctrl), nil).Validate(faction, nil, rulebook) {
			t.Error("expected false when faction has no coin")
		}
	})

	t.Run("damaged and has coin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		asset := &domain.Asset{ID: "a1", DefinitionID: "tough", CurrentHP: 6}
		faction := &domain.Faction{ID: "f1", Coin: 5, Assets: map[string]*domain.Asset{"a1": asset}}
		if !NewRepairAsset(mocks.NewMockCollector(ctrl), nil).Validate(faction, nil, rulebook) {
			t.Error("expected true when faction has a damaged asset and coin")
		}
	})
}

// TestRepairAsset_Output: one repair order, HealCount=1.
// Force=4, missing=4 → healAmount=min(4,4)=4, cost=1.
// Emits AssetHPDelta(+4) + CoinDelta(-1).
func TestRepairAsset_Output(t *testing.T) {
	ctrl := gomock.NewController(t)
	rulebook := makeRepairAssetRulebook()
	asset := &domain.Asset{ID: "a1", DefinitionID: "tough", CurrentHP: 6}
	faction := &domain.Faction{ID: "f1", Force: 4, Coin: 5, Assets: map[string]*domain.Asset{"a1": asset}}

	orders := []action.RepairOrder{{Asset: asset, HealCount: 1}}
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRepairOrders(gomock.Any(), gomock.Any(), gomock.Any()).Return(orders, nil)

	act := NewRepairAsset(collector, nil)
	if err := act.Inputs(faction, nil, rulebook); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, nil, rulebook); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}

	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2; got %v", len(mutations), mutations)
	}
	delta, ok := mutations[0].(domain.AssetHPDelta)
	if !ok || delta.AssetID != "a1" || delta.Delta != 4 || delta.Cause != "repair" {
		t.Errorf("mutations[0] = %v, want AssetHPDelta{a1, +4, repair}", mutations[0])
	}
	coin, ok := mutations[1].(domain.CoinDelta)
	if !ok || coin.FactionID != "f1" || coin.Delta != -1 || coin.Cause != "repair" {
		t.Errorf("mutations[1] = %v, want CoinDelta{f1, -1, repair}", mutations[1])
	}
}

// TestRepairAsset_ClearsShelterAtFullHP: a sheltered asset at 0 HP, Force 4, one
// heal. The shelter clears only when the heal reaches full HP (Effort 2 Decision 12).
func TestRepairAsset_ClearsShelterAtFullHP(t *testing.T) {
	sheltered := domain.Condition{Kind: domain.ConditionSheltered, SourceAssetID: "bolt"}
	heal := domain.AssetHPDelta{FactionID: "f1", AssetID: "a1", Delta: 4, Cause: "repair", CausedByFactionID: "f1"}
	unshelter := domain.AssetConditionCleared{FactionID: "f1", AssetID: "a1", Condition: sheltered, Cause: "repair", CausedByFactionID: "f1"}
	pay := domain.CoinDelta{FactionID: "f1", Delta: -1, Cause: "repair", CausedByFactionID: "f1"}
	cases := []struct {
		name string
		hp   int
		want []domain.Mutation
	}{
		{name: "a heal to full HP clears the shelter", hp: 4, want: []domain.Mutation{heal, unshelter, pay}},
		{name: "a heal short of full HP keeps it", hp: 8, want: []domain.Mutation{heal, pay}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			rulebook := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
				"unit": {ID: "unit", Category: domain.StatForce, HP: tc.hp},
			}}
			asset := &domain.Asset{ID: "a1", DefinitionID: "unit", CurrentHP: 0, Conditions: []domain.Condition{sheltered}}
			faction := &domain.Faction{ID: "f1", Force: 4, Coin: 5, Assets: map[string]*domain.Asset{"a1": asset}}
			collector := mocks.NewMockCollector(ctrl)
			collector.EXPECT().SelectRepairOrders(gomock.Any(), gomock.Any(), gomock.Any()).Return([]action.RepairOrder{{Asset: asset, HealCount: 1}}, nil)

			act := NewRepairAsset(collector, nil)
			if err := act.Inputs(faction, nil, rulebook); err != nil {
				t.Fatalf("Inputs: %v", err)
			}
			if err := act.Resolve(faction, nil, rulebook); err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			mutations, _ := act.Output()
			if !reflect.DeepEqual(mutations, tc.want) {
				t.Errorf("mutations = %+v, want %+v", mutations, tc.want)
			}
		})
	}
}

// repairDiscountFor is a fixture RepairCostModifier that takes 1 off the price of
// one asset.
type repairDiscountFor struct {
	assetID string
}

func (modifier repairDiscountFor) ModifyRepairCost(_ *domain.Faction, asset *domain.Asset, _ *domain.AssetDefinition, query *hooks.RepairCostQuery) {
	if asset.ID == modifier.assetID {
		query.Cost--
	}
}

// TestRepairAsset_DiscountLowersEachAssetsPrice: one heal each on a1 and a2, a
// discount on a1 only → a1 costs 0, a2 costs 1, one CoinDelta of -1.
func TestRepairAsset_DiscountLowersEachAssetsPrice(t *testing.T) {
	ctrl := gomock.NewController(t)
	rulebook := makeRepairAssetRulebook()
	first := &domain.Asset{ID: "a1", DefinitionID: "tough", CurrentHP: 6}
	second := &domain.Asset{ID: "a2", DefinitionID: "tough", CurrentHP: 6}
	faction := &domain.Faction{ID: "f1", Force: 4, Coin: 5, Assets: map[string]*domain.Asset{"a1": first, "a2": second}}
	registry := hooks.NewRegistry()
	registry.RegisterRepairCostModifier(hooks.FactionScope("f1"), "fixture", repairDiscountFor{assetID: "a1"})

	orders := []action.RepairOrder{{Asset: first, HealCount: 1}, {Asset: second, HealCount: 1}}
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRepairOrders(gomock.Any(), gomock.Any(), gomock.Any()).Return(orders, nil)

	act := NewRepairAsset(collector, registry)
	if err := act.Inputs(faction, nil, rulebook); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, nil, rulebook); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, _ := act.Output()
	var coin []domain.CoinDelta
	for _, mutation := range mutations {
		if delta, ok := mutation.(domain.CoinDelta); ok {
			coin = append(coin, delta)
		}
	}
	if len(coin) != 1 || coin[0].Delta != -1 {
		t.Errorf("coin = %+v, want one CoinDelta{-1}", coin)
	}
}

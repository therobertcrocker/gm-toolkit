package actions

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"go.uber.org/mock/gomock"
)

func makeRepairAssetRulebook() *loader.Rulebook {
	return &loader.Rulebook{
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
		faction := &domain.Faction{ID: "f1", Coin: 5, Assets: []*domain.Asset{asset}}
		if NewRepairAsset(mocks.NewMockCollector(ctrl)).Validate(faction, nil, rulebook) {
			t.Error("expected false when no assets are damaged")
		}
	})

	t.Run("damaged but no coin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		asset := &domain.Asset{ID: "a1", DefinitionID: "tough", CurrentHP: 6}
		faction := &domain.Faction{ID: "f1", Coin: 0, Assets: []*domain.Asset{asset}}
		if NewRepairAsset(mocks.NewMockCollector(ctrl)).Validate(faction, nil, rulebook) {
			t.Error("expected false when faction has no coin")
		}
	})

	t.Run("damaged and has coin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		asset := &domain.Asset{ID: "a1", DefinitionID: "tough", CurrentHP: 6}
		faction := &domain.Faction{ID: "f1", Coin: 5, Assets: []*domain.Asset{asset}}
		if !NewRepairAsset(mocks.NewMockCollector(ctrl)).Validate(faction, nil, rulebook) {
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
	faction := &domain.Faction{ID: "f1", Force: 4, Coin: 5, Assets: []*domain.Asset{asset}}

	orders := []action.RepairOrder{{Asset: asset, HealCount: 1}}
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRepairOrders(gomock.Any(), gomock.Any(), gomock.Any()).Return(orders, nil)

	act := NewRepairAsset(collector)
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

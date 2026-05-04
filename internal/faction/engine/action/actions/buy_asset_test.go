package actions

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"go.uber.org/mock/gomock"
)

func makeBuyRulebook() *loader.Rulebook {
	return &loader.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			"cheap": {ID: "cheap", Name: "Militia", Category: domain.StatForce, HP: 4, Cost: 4, MinRating: 2},
		},
	}
}

func TestBuyAsset_Validate(t *testing.T) {
	rulebook := makeBuyRulebook()

	t.Run("eligible asset and sufficient coin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5, Homeworld: "Tartarus"}
		if !NewBuyAsset(mocks.NewMockCollector(ctrl)).Validate(faction, nil, rulebook) {
			t.Error("expected true when faction can afford an eligible asset")
		}
	})

	t.Run("not enough coin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		faction := &domain.Faction{ID: "f1", Force: 3, Coin: 2, Homeworld: "Tartarus"}
		if NewBuyAsset(mocks.NewMockCollector(ctrl)).Validate(faction, nil, rulebook) {
			t.Error("expected false when faction cannot afford any asset")
		}
	})

	t.Run("stat below MinRating", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		faction := &domain.Faction{ID: "f1", Force: 1, Coin: 10, Homeworld: "Tartarus"}
		if NewBuyAsset(mocks.NewMockCollector(ctrl)).Validate(faction, nil, rulebook) {
			t.Error("expected false when faction stat is below MinRating")
		}
	})
}

// TestBuyAsset_Output: purchase emits AssetAdded(Ready=false) then CoinDelta(-cost).
func TestBuyAsset_Output(t *testing.T) {
	ctrl := gomock.NewController(t)
	rulebook := makeBuyRulebook()
	def := rulebook.Assets["cheap"]
	faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5, Homeworld: "Tartarus"}

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectBuyOrder(gomock.Any(), gomock.Any()).Return(
		action.BuyOrder{World: "Tartarus", Definition: def}, nil,
	)

	act := NewBuyAsset(collector)
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
	added, ok := mutations[0].(domain.AssetAdded)
	if !ok || added.FactionID != "f1" || added.Asset.DefinitionID != "cheap" {
		t.Errorf("mutations[0] = %v, want AssetAdded{f1, cheap}", mutations[0])
	}
	if added.Asset.Ready {
		t.Error("new asset must have Ready=false until next turn start")
	}
	if added.Asset.Location != "Tartarus" {
		t.Errorf("new asset Location = %q, want Tartarus", added.Asset.Location)
	}
	coin, ok := mutations[1].(domain.CoinDelta)
	if !ok || coin.FactionID != "f1" || coin.Delta != -4 || coin.Cause != "buy" {
		t.Errorf("mutations[1] = %v, want CoinDelta{f1, -4, buy}", mutations[1])
	}
}

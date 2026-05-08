package actions

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"go.uber.org/mock/gomock"
)

func makeRefitRulebook() *rulebook.Rulebook {
	return &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			"old-def": {ID: "old-def", Name: "Militia", Category: domain.StatForce, HP: 4, Cost: 2},
			"new-def": {ID: "new-def", Name: "Infantry", Category: domain.StatForce, HP: 6, Cost: 4},
		},
	}
}

func TestRefitAsset_Validate(t *testing.T) {
	rulebook := makeRefitRulebook()

	t.Run("no assets", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5}
		if NewRefitAsset(mocks.NewMockInputCollector(ctrl)).Validate(faction, nil, rulebook) {
			t.Error("expected false when faction has no assets")
		}
	})

	t.Run("cannot afford upgrade delta", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		// delta = 4-2 = 2; Coin=1 < 2 → no valid replacement
		asset := &domain.Asset{ID: "a1", DefinitionID: "old-def", Location: "Tartarus"}
		faction := &domain.Faction{ID: "f1", Force: 3, Coin: 1, Assets: map[string]*domain.Asset{"a1": asset}}
		if NewRefitAsset(mocks.NewMockInputCollector(ctrl)).Validate(faction, nil, rulebook) {
			t.Error("expected false when faction cannot afford the cost delta")
		}
	})

	t.Run("valid refit exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		asset := &domain.Asset{ID: "a1", DefinitionID: "old-def", Location: "Tartarus"}
		faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5, Assets: map[string]*domain.Asset{"a1": asset}}
		if !NewRefitAsset(mocks.NewMockInputCollector(ctrl)).Validate(faction, nil, rulebook) {
			t.Error("expected true when faction can afford the refit")
		}
	})
}

// TestRefitAsset_Output: swap old-def(cost=2) for new-def(cost=4).
// Emits AssetRemoved + AssetAdded(Ready=false, same Location) + CoinDelta(-2).
func TestRefitAsset_Output(t *testing.T) {
	ctrl := gomock.NewController(t)
	rulebook := makeRefitRulebook()
	newDef := rulebook.Assets["new-def"]
	oldAsset := &domain.Asset{ID: "f1-old-def-1", DefinitionID: "old-def", OwnerID: "f1", Location: "Tartarus"}
	faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5, Assets: map[string]*domain.Asset{"f1-old-def-1": oldAsset}}

	order := action.RefitOrder{OldAsset: oldAsset, NewDefinition: newDef}
	collector := mocks.NewMockInputCollector(ctrl)
	collector.EXPECT().SelectRefitOrder(gomock.Any(), gomock.Any()).Return(order, nil)

	act := NewRefitAsset(collector)
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

	if len(mutations) != 3 {
		t.Fatalf("len(mutations) = %d, want 3; got %v", len(mutations), mutations)
	}
	removed, ok := mutations[0].(domain.AssetRemoved)
	if !ok || removed.AssetID != "f1-old-def-1" || removed.Cause != "refit" {
		t.Errorf("mutations[0] = %v, want AssetRemoved{f1-old-def-1, refit}", mutations[0])
	}
	added, ok := mutations[1].(domain.AssetAdded)
	if !ok || added.Asset.DefinitionID != "new-def" {
		t.Errorf("mutations[1] = %v, want AssetAdded{new-def}", mutations[1])
	}
	if added.Asset.Ready {
		t.Error("refitted asset must have Ready=false until next turn start")
	}
	if added.Asset.Location != "Tartarus" {
		t.Errorf("refitted asset Location = %q, want Tartarus", added.Asset.Location)
	}
	coin, ok := mutations[2].(domain.CoinDelta)
	if !ok || coin.FactionID != "f1" || coin.Delta != -2 || coin.Cause != "refit" {
		t.Errorf("mutations[2] = %v, want CoinDelta{f1, -2, refit}", mutations[2])
	}
}

package actions

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"go.uber.org/mock/gomock"
)

func makeSimpleRulebook() *rulebook.Rulebook {
	return &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			"sell-me": {ID: "sell-me", Name: "Infantry", Category: domain.StatForce, HP: 4, Cost: 6},
		},
	}
}

// --- SellAsset ---

func TestSellAsset_Validate(t *testing.T) {
	rulebook := makeSimpleRulebook()

	t.Run("no assets", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		faction := &domain.Faction{ID: "f1"}
		if NewSellAsset(mocks.NewMockCollector(ctrl)).Validate(faction, nil, rulebook) {
			t.Error("expected false when faction has no assets")
		}
	})

	t.Run("has assets", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		asset := &domain.Asset{ID: "a1", DefinitionID: "sell-me"}
		faction := &domain.Faction{ID: "f1", Assets: []*domain.Asset{asset}}
		if !NewSellAsset(mocks.NewMockCollector(ctrl)).Validate(faction, nil, rulebook) {
			t.Error("expected true when faction has assets")
		}
	})
}

// TestSellAsset_Output: selling an asset returns AssetRemoved + CoinDelta(cost/2 = 3).
func TestSellAsset_Output(t *testing.T) {
	ctrl := gomock.NewController(t)
	rulebook := makeSimpleRulebook()
	asset := &domain.Asset{ID: "a1", DefinitionID: "sell-me"}
	faction := &domain.Faction{ID: "f1", Assets: []*domain.Asset{asset}}

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectAsset(gomock.Any(), gomock.Any()).Return(asset, nil)

	act := NewSellAsset(collector)
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
	removed, ok := mutations[0].(domain.AssetRemoved)
	if !ok || removed.AssetID != "a1" || removed.Cause != "sell" {
		t.Errorf("mutations[0] = %v, want AssetRemoved{a1, sell}", mutations[0])
	}
	coin, ok := mutations[1].(domain.CoinDelta)
	if !ok || coin.FactionID != "f1" || coin.Delta != 3 || coin.Cause != "sell" {
		t.Errorf("mutations[1] = %v, want CoinDelta{f1, +3, sell}", mutations[1])
	}
}

// --- RepairFaction ---

func TestRepairFaction_Validate(t *testing.T) {
	t.Run("HP full", func(t *testing.T) {
		faction := &domain.Faction{ID: "f1", CurrentHP: 10, MaxHP: 10}
		if NewRepairFaction().Validate(faction, nil, nil) {
			t.Error("expected false when HP is full")
		}
	})

	t.Run("HP damaged", func(t *testing.T) {
		faction := &domain.Faction{ID: "f1", CurrentHP: 7, MaxHP: 10}
		if !NewRepairFaction().Validate(faction, nil, nil) {
			t.Error("expected true when HP is below max")
		}
	})
}

// TestRepairFaction_Output_Basic: Force=4, Cunning=2, Wealth=1.
// highest=4, lowest=1 → raw=(4+1+1)/2=3. Missing=3 → heal=3.
func TestRepairFaction_Output_Basic(t *testing.T) {
	faction := &domain.Faction{ID: "f1", Force: 4, Cunning: 2, Wealth: 1, CurrentHP: 7, MaxHP: 10}
	act := NewRepairFaction()
	if err := act.Inputs(faction, nil, nil); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, nil, nil); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}

	if len(mutations) != 1 {
		t.Fatalf("len(mutations) = %d, want 1", len(mutations))
	}
	delta, ok := mutations[0].(domain.FactionHPDelta)
	if !ok || delta.FactionID != "f1" || delta.Delta != 3 {
		t.Errorf("mutations[0] = %v, want FactionHPDelta{f1, +3}", mutations[0])
	}
}

// TestRepairFaction_Output_CappedAtMissing: raw heal exceeds missing HP.
// Force=6, Cunning=4, Wealth=2 → raw=(6+2+1)/2=4. Missing=2 → heal=2.
func TestRepairFaction_Output_CappedAtMissing(t *testing.T) {
	faction := &domain.Faction{ID: "f1", Force: 6, Cunning: 4, Wealth: 2, CurrentHP: 8, MaxHP: 10}
	act := NewRepairFaction()
	if err := act.Inputs(faction, nil, nil); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, nil, nil); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}

	if len(mutations) != 1 {
		t.Fatalf("len(mutations) = %d, want 1", len(mutations))
	}
	delta, ok := mutations[0].(domain.FactionHPDelta)
	if !ok || delta.Delta != 2 {
		t.Errorf("mutations[0] = %v, want FactionHPDelta{delta:2}", mutations[0])
	}
}

// --- AbandonGoal ---

func TestAbandonGoal_Validate(t *testing.T) {
	t.Run("no active goal", func(t *testing.T) {
		faction := &domain.Faction{ID: "f1"}
		if NewAbandonGoal().Validate(faction, nil, nil) {
			t.Error("expected false when no active goal")
		}
	})

	t.Run("has active goal", func(t *testing.T) {
		faction := &domain.Faction{ID: "f1", ActiveGoal: &domain.ActiveGoal{GoalID: "G-001"}}
		if !NewAbandonGoal().Validate(faction, nil, nil) {
			t.Error("expected true when faction has an active goal")
		}
	})
}

// TestAbandonGoal_Output: income penalty = Wealth/2 + (Force+Cunning)/4.
// Wealth=4, Force=2, Cunning=2 → income = 2+1 = 3 → CoinDelta{-3} + GoalAbandoned.
func TestAbandonGoal_Output(t *testing.T) {
	faction := &domain.Faction{
		ID: "f1", Force: 2, Cunning: 2, Wealth: 4,
		ActiveGoal: &domain.ActiveGoal{GoalID: "G-001"},
	}
	act := NewAbandonGoal()
	if err := act.Inputs(faction, nil, nil); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, nil, nil); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}

	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2", len(mutations))
	}
	coin, ok := mutations[0].(domain.CoinDelta)
	if !ok || coin.FactionID != "f1" || coin.Delta != -3 {
		t.Errorf("mutations[0] = %v, want CoinDelta{f1, -3}", mutations[0])
	}
	abandoned, ok := mutations[1].(domain.GoalAbandoned)
	if !ok || abandoned.FactionID != "f1" || abandoned.GoalID != "G-001" {
		t.Errorf("mutations[1] = %v, want GoalAbandoned{f1, G-001}", mutations[1])
	}
}

// --- Bribe ---

func TestBribe_Validate(t *testing.T) {
	t.Run("no coin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		base := &domain.Base{ID: "b1", OwnerID: "f1"}
		faction := &domain.Faction{ID: "f1", Coin: 0, Bases: []*domain.Base{base}}
		if NewBribe(mocks.NewMockCollector(ctrl)).Validate(faction, nil, nil) {
			t.Error("expected false when faction has no coin")
		}
	})

	t.Run("no own bases", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		faction := &domain.Faction{ID: "f1", Coin: 5}
		if NewBribe(mocks.NewMockCollector(ctrl)).Validate(faction, nil, nil) {
			t.Error("expected false when faction has no bases of its own")
		}
	})

	t.Run("has bases and coin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		base := &domain.Base{ID: "b1", OwnerID: "f1"}
		faction := &domain.Faction{ID: "f1", Coin: 5, Bases: []*domain.Base{base}}
		if !NewBribe(mocks.NewMockCollector(ctrl)).Validate(faction, nil, nil) {
			t.Error("expected true when faction has bases and coin")
		}
	})
}

// TestBribe_Output: emits CoinDelta(-amount) + InfluenceDelta(+amount) on target base.
func TestBribe_Output(t *testing.T) {
	ctrl := gomock.NewController(t)
	rivalBase := &domain.Base{ID: "rb1", OwnerID: "f2", Location: "Krylos"}
	ownBase := &domain.Base{ID: "ob1", OwnerID: "f1", Location: "Krylos"}
	faction := &domain.Faction{ID: "f1", Coin: 5, Bases: []*domain.Base{ownBase}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectBribeTarget(gomock.Any(), gomock.Any()).Return(rivalBase, 3, nil)

	act := NewBribe(collector)
	if err := act.Inputs(faction, factionState, nil); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, factionState, nil); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}

	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2", len(mutations))
	}
	coin, ok := mutations[0].(domain.CoinDelta)
	if !ok || coin.FactionID != "f1" || coin.Delta != -3 || coin.Cause != "bribe" {
		t.Errorf("mutations[0] = %v, want CoinDelta{f1, -3, bribe}", mutations[0])
	}
	influence, ok := mutations[1].(domain.InfluenceDelta)
	if !ok || influence.BaseID != "rb1" || influence.Delta != 3 {
		t.Errorf("mutations[1] = %v, want InfluenceDelta{rb1, +3}", mutations[1])
	}
}

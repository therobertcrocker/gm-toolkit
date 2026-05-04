package actions

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"go.uber.org/mock/gomock"
)

func makeExpandRulebook() *loader.Rulebook {
	return &loader.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			"force-unit": {
				ID:       "force-unit",
				Name:     "Infantry",
				Category: domain.StatForce,
				HP:       8,
				Attack: &domain.AttackProfile{
					AttackerStat: domain.StatForce,
					DefenderStat: domain.StatCunning,
					Damage:       domain.DiceRoll{NumDice: 1, Sides: 6},
				},
			},
		},
	}
}

func TestExpandInfluence_Validate(t *testing.T) {
	t.Run("no coin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		asset := &domain.Asset{ID: "a1", OwnerID: "f1", Location: "Krylos"}
		faction := &domain.Faction{ID: "f1", Coin: 0, MaxHP: 10, Assets: []*domain.Asset{asset}}
		factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}
		if NewExpandInfluence(mocks.NewMockCollector(ctrl), nil).Validate(faction, factionState, nil) {
			t.Error("expected false when coin < 1")
		}
	})

	t.Run("no expansion options", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		asset := &domain.Asset{ID: "a1", OwnerID: "f1", Location: "Krylos"}
		// Base on Krylos at max HP and at faction.MaxHP — no new-base, no heal, no grow option.
		base := &domain.Base{ID: "b1", OwnerID: "f1", Location: "Krylos", CurrentHP: 5, MaxHP: 5, IsHomeworld: false}
		faction := &domain.Faction{ID: "f1", Coin: 3, MaxHP: 5, Assets: []*domain.Asset{asset}, Bases: []*domain.Base{base}}
		factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}
		if NewExpandInfluence(mocks.NewMockCollector(ctrl), nil).Validate(faction, factionState, nil) {
			t.Error("expected false when no expansion options available")
		}
	})

	t.Run("world available for new base", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		asset := &domain.Asset{ID: "a1", OwnerID: "f1", Location: "Krylos"}
		faction := &domain.Faction{ID: "f1", Coin: 3, MaxHP: 10, Assets: []*domain.Asset{asset}}
		factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}
		if !NewExpandInfluence(mocks.NewMockCollector(ctrl), nil).Validate(faction, factionState, nil) {
			t.Error("expected true when faction has an asset on a world with no base")
		}
	})
}

// TestExpandInfluence_NewBase_Uncontested: no rivals on target world.
// Emits CoinDelta(-HPAmount) + BaseAdded(Ready=false).
func TestExpandInfluence_NewBase_Uncontested(t *testing.T) {
	ctrl := gomock.NewController(t)
	asset := &domain.Asset{ID: "a1", OwnerID: "f1", Location: "Krylos"}
	faction := &domain.Faction{ID: "f1", Cunning: 0, Coin: 5, MaxHP: 10, Assets: []*domain.Asset{asset}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	order := action.ExpandInfluenceOrder{Mode: action.ExpandModeNew, World: "Krylos", HPAmount: 3}
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectExpandInfluenceOrder(gomock.Any(), gomock.Any()).Return(order, nil)

	act := NewExpandInfluence(collector, &fixedRoller{values: []int{8}})
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
		t.Fatalf("len(mutations) = %d, want 2; got %v", len(mutations), mutations)
	}
	coin, ok := mutations[0].(domain.CoinDelta)
	if !ok || coin.Delta != -3 || coin.Cause != "expand" {
		t.Errorf("mutations[0] = %v, want CoinDelta{-3, expand}", mutations[0])
	}
	added, ok := mutations[1].(domain.BaseAdded)
	if !ok || added.Base.Location != "Krylos" || added.Base.CurrentHP != 3 {
		t.Errorf("mutations[1] = %v, want BaseAdded{Krylos, HP=3}", mutations[1])
	}
	if added.Base.Ready {
		t.Error("new base must have Ready=false")
	}
}

// TestExpandInfluence_Reinforce_Heal: restore HP on a damaged non-homeworld base.
// Emits BaseHealed(+amount) + CoinDelta(-amount).
func TestExpandInfluence_Reinforce_Heal(t *testing.T) {
	ctrl := gomock.NewController(t)
	base := &domain.Base{ID: "b1", OwnerID: "f1", Location: "Krylos", CurrentHP: 2, MaxHP: 5, IsHomeworld: false}
	faction := &domain.Faction{ID: "f1", Coin: 5, MaxHP: 10, Bases: []*domain.Base{base}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	order := action.ExpandInfluenceOrder{
		Mode: action.ExpandModeReinforce, BaseID: "b1",
		SubMode: action.ReinforceHeal, HPAmount: 2,
	}
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectExpandInfluenceOrder(gomock.Any(), gomock.Any()).Return(order, nil)

	act := NewExpandInfluence(collector, &fixedRoller{values: []int{1}})
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
		t.Fatalf("len(mutations) = %d, want 2; got %v", len(mutations), mutations)
	}
	healed, ok := mutations[0].(domain.BaseHealed)
	if !ok || healed.BaseID != "b1" || healed.Delta != 2 {
		t.Errorf("mutations[0] = %v, want BaseHealed{b1, +2}", mutations[0])
	}
	coin, ok := mutations[1].(domain.CoinDelta)
	if !ok || coin.Delta != -2 {
		t.Errorf("mutations[1] = %v, want CoinDelta{-2}", mutations[1])
	}
}

// TestExpandInfluence_Reinforce_Max: increase a non-homeworld base's MaxHP.
// Emits BaseExpanded(+amount) + CoinDelta(-amount).
func TestExpandInfluence_Reinforce_Max(t *testing.T) {
	ctrl := gomock.NewController(t)
	base := &domain.Base{ID: "b1", OwnerID: "f1", Location: "Krylos", CurrentHP: 5, MaxHP: 5, IsHomeworld: false}
	faction := &domain.Faction{ID: "f1", Coin: 5, MaxHP: 10, Bases: []*domain.Base{base}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	order := action.ExpandInfluenceOrder{
		Mode: action.ExpandModeReinforce, BaseID: "b1",
		SubMode: action.ReinforceMax, HPAmount: 3,
	}
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectExpandInfluenceOrder(gomock.Any(), gomock.Any()).Return(order, nil)

	act := NewExpandInfluence(collector, &fixedRoller{values: []int{1}})
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
		t.Fatalf("len(mutations) = %d, want 2; got %v", len(mutations), mutations)
	}
	expanded, ok := mutations[0].(domain.BaseExpanded)
	if !ok || expanded.BaseID != "b1" || expanded.Delta != 3 {
		t.Errorf("mutations[0] = %v, want BaseExpanded{b1, +3}", mutations[0])
	}
	coin, ok := mutations[1].(domain.CoinDelta)
	if !ok || coin.Delta != -3 {
		t.Errorf("mutations[1] = %v, want CoinDelta{-3}", mutations[1])
	}
}

// TestExpandInfluence_NewBase_Contested: rival ties the expansion roll and launches a free attack.
// Rolls: [5]=factionRoll, [5]=rivalRoll (tie → attack confirmed), [8]=attackRoll, [3]=defenseRoll, [4]=damage.
// Expects CoinDelta + BaseAdded + BaseHPDelta(-4, caused by f2) + FactionHPDelta(-4, caused by f2).
func TestExpandInfluence_NewBase_Contested(t *testing.T) {
	ctrl := gomock.NewController(t)
	rulebook := makeExpandRulebook()

	f1Asset := &domain.Asset{ID: "a1", OwnerID: "f1", Location: "Krylos"}
	f2Asset := &domain.Asset{
		ID: "a2", DefinitionID: "force-unit", OwnerID: "f2",
		Location: "Krylos", CurrentHP: 8, Ready: true, Maintained: true,
	}
	faction := &domain.Faction{ID: "f1", Cunning: 0, Coin: 5, MaxHP: 10, Assets: []*domain.Asset{f1Asset}}
	rival := &domain.Faction{ID: "f2", Force: 0, Cunning: 0, Assets: []*domain.Asset{f2Asset}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction, "f2": rival}}

	order := action.ExpandInfluenceOrder{Mode: action.ExpandModeNew, World: "Krylos", HPAmount: 5}
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectExpandInfluenceOrder(gomock.Any(), gomock.Any()).Return(order, nil)
	collector.EXPECT().ConfirmRivalFreeAttack(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	collector.EXPECT().SelectBaseAttackers(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*domain.Asset{f2Asset}, nil)

	// Roll sequence: factionRoll=5, rivalRoll=5 (tie → confirmed attack), attackRoll=8, defenseRoll=3, damage=4.
	roller := &fixedRoller{values: []int{5, 5, 8, 3, 4}}
	act := NewExpandInfluence(collector, roller)
	if err := act.Inputs(faction, factionState, rulebook); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, factionState, rulebook); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}

	if len(mutations) != 4 {
		t.Fatalf("len(mutations) = %d, want 4; got %v", len(mutations), mutations)
	}
	if _, ok := mutations[0].(domain.CoinDelta); !ok {
		t.Errorf("mutations[0] type = %T, want CoinDelta", mutations[0])
	}
	added, ok := mutations[1].(domain.BaseAdded)
	if !ok || added.Base.Location != "Krylos" {
		t.Errorf("mutations[1] = %v, want BaseAdded{Krylos}", mutations[1])
	}
	baseDelta, ok := mutations[2].(domain.BaseHPDelta)
	if !ok || baseDelta.Delta != -4 || baseDelta.CausedByFactionID != "f2" {
		t.Errorf("mutations[2] = %v, want BaseHPDelta{-4, caused by f2}", mutations[2])
	}
	factionDelta, ok := mutations[3].(domain.FactionHPDelta)
	if !ok || factionDelta.Delta != -4 || factionDelta.CausedByFactionID != "f2" {
		t.Errorf("mutations[3] = %v, want FactionHPDelta{-4, caused by f2}", mutations[3])
	}
}

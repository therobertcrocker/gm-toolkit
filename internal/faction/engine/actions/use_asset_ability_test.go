package actions

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// abilityFakeCollector provides pre-configured responses for UseAssetAbility input methods.
// Unused InputCollector methods panic to surface accidental calls in tests.
type abilityFakeCollector struct {
	selectedAssets    []*domain.Asset
	moveDestination   string
	factionTestTarget *domain.Faction
	confirmCalled     bool
}

func (c *abilityFakeCollector) SelectAbilityAssets(_ *domain.Faction, _ []*domain.Asset, _ *loader.Rulebook) ([]*domain.Asset, error) {
	return c.selectedAssets, nil
}
func (c *abilityFakeCollector) SelectMoveDestination(_ *domain.Asset, _ []string) (string, error) {
	return c.moveDestination, nil
}
func (c *abilityFakeCollector) SelectFactionTestTarget(_ *domain.Asset, _ domain.AbilityEffectType, _ []*domain.Faction) (*domain.Faction, error) {
	return c.factionTestTarget, nil
}
func (c *abilityFakeCollector) ConfirmAbilityApplied(_ *domain.Asset, _ *domain.AssetDefinition) (bool, error) {
	c.confirmCalled = true
	return false, nil
}
func (c *abilityFakeCollector) SelectAsset(_ []*domain.Asset, _ *loader.Rulebook) (*domain.Asset, error) {
	panic("SelectAsset: not used in use_asset_ability tests")
}
func (c *abilityFakeCollector) SelectRepairOrders(_ *domain.Faction, _ []*domain.Asset, _ *loader.Rulebook) ([]engine.RepairOrder, error) {
	panic("SelectRepairOrders: not used in use_asset_ability tests")
}
func (c *abilityFakeCollector) SelectBuyOrder(_ []string, _ []*domain.AssetDefinition) (engine.BuyOrder, error) {
	panic("SelectBuyOrder: not used in use_asset_ability tests")
}
func (c *abilityFakeCollector) SelectRefitOrder(_ []engine.RefitOption, _ *loader.Rulebook) (engine.RefitOrder, error) {
	panic("SelectRefitOrder: not used in use_asset_ability tests")
}
func (c *abilityFakeCollector) SelectAttackers(_ []*domain.Asset, _ *loader.Rulebook) ([]*domain.Asset, error) {
	panic("SelectAttackers: not used in use_asset_ability tests")
}
func (c *abilityFakeCollector) SelectDefender(_ *domain.Asset, _ []*domain.Asset, _ *loader.Rulebook) (*domain.Asset, error) {
	panic("SelectDefender: not used in use_asset_ability tests")
}
func (c *abilityFakeCollector) ConfirmRedirectToBase(_ *domain.Faction, _ *domain.Base, _ int) (bool, error) {
	panic("ConfirmRedirectToBase: not used in use_asset_ability tests")
}
func (c *abilityFakeCollector) SelectExpandInfluenceOrder(_ *domain.Faction, _ *state.FactionState) (engine.ExpandInfluenceOrder, error) {
	panic("SelectExpandInfluenceOrder: not used in use_asset_ability tests")
}
func (c *abilityFakeCollector) ConfirmRivalFreeAttack(_ *domain.Faction, _, _ int) (bool, error) {
	panic("ConfirmRivalFreeAttack: not used in use_asset_ability tests")
}
func (c *abilityFakeCollector) SelectBaseAttackers(_ *domain.Faction, _ []*domain.Asset, _ *loader.Rulebook) ([]*domain.Asset, error) {
	panic("SelectBaseAttackers: not used in use_asset_ability tests")
}

// makeAbilityRulebook returns a minimal Rulebook with four asset definitions:
//   - "move-asset": A-flagged, movement ability, no coin cost
//   - "move-asset-coin": A-flagged, movement ability, coin cost 2
//   - "drain-asset": A-flagged, faction_test ability with coin_drain effect (1d6)
//   - "deferred-asset": A-flagged, no ability (GM adjudication)
func makeAbilityRulebook() *loader.Rulebook {
	effectDice := &domain.DiceRoll{NumDice: 1, Sides: 6}
	return &loader.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			"move-asset": {
				ID:    "move-asset",
				Name:  "Scout Ship",
				Flags: []domain.AssetFlag{domain.FlagAction},
				HP:    4,
				Ability: &domain.AbilityDefinition{
					Steps: []domain.AbilityStep{
						{Type: domain.AbilityStepMovement, CoinCost: 0},
					},
				},
			},
			"move-asset-coin": {
				ID:    "move-asset-coin",
				Name:  "Freighter",
				Flags: []domain.AssetFlag{domain.FlagAction},
				HP:    4,
				Ability: &domain.AbilityDefinition{
					Steps: []domain.AbilityStep{
						{Type: domain.AbilityStepMovement, CoinCost: 2},
					},
				},
			},
			"drain-asset": {
				ID:    "drain-asset",
				Name:  "Informers",
				Flags: []domain.AssetFlag{domain.FlagAction},
				HP:    4,
				Ability: &domain.AbilityDefinition{
					Steps: []domain.AbilityStep{
						{
							Type:         domain.AbilityStepFactionTest,
							AttackerStat: domain.StatCunning,
							DefenderStat: domain.StatCunning,
							Effect:       domain.EffectCoinDrain,
							EffectDice:   effectDice,
						},
					},
				},
			},
			"deferred-asset": {
				ID:      "deferred-asset",
				Name:    "Tripwire Cells",
				Flags:   []domain.AssetFlag{domain.FlagAction},
				HP:      4,
				Ability: nil,
			},
		},
	}
}

// makeAbilityState builds a FactionState with an acting faction (f1) and an
// optional target faction (f2), both with one asset on world "Anchorage".
func makeAbilityState(actingAssetDefID string, includeTarget bool) (*state.FactionState, *domain.Asset) {
	actingAsset := &domain.Asset{
		ID:           "a1",
		DefinitionID: actingAssetDefID,
		OwnerID:      "f1",
		Location:     "Anchorage",
		CurrentHP:    4,
		Ready:        true,
		Maintained:   true,
	}
	factions := map[string]*domain.Faction{
		"f1": {ID: "f1", Cunning: 0, Assets: []*domain.Asset{actingAsset}},
	}
	if includeTarget {
		targetAsset := &domain.Asset{
			ID:           "t1",
			DefinitionID: "drain-asset",
			OwnerID:      "f2",
			Location:     "Anchorage",
			CurrentHP:    4,
			Ready:        true,
			Maintained:   true,
		}
		factions["f2"] = &domain.Faction{ID: "f2", Cunning: 0, Assets: []*domain.Asset{targetAsset}}
	}
	return &state.FactionState{Factions: factions}, actingAsset
}

// runAbility drives a full Inputs→Resolve→Output cycle for UseAssetAbility.
func runAbility(t *testing.T, collector *abilityFakeCollector, roller domain.Roller, faction *domain.Faction, factionState *state.FactionState, rulebook *loader.Rulebook) []domain.Mutation {
	t.Helper()
	ae := engine.NewAbilityEngine()
	action := NewUseAssetAbility(collector, roller, ae)
	if err := action.Inputs(faction, factionState, rulebook); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := action.Resolve(faction, factionState, rulebook); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := action.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}
	return mutations
}

func TestUseAssetAbility_Validate(t *testing.T) {
	rulebook := makeAbilityRulebook()

	t.Run("no A-flagged assets", func(t *testing.T) {
		asset := &domain.Asset{ID: "x1", DefinitionID: "no-flag", Ready: true, Maintained: true}
		rulebook.Assets["no-flag"] = &domain.AssetDefinition{ID: "no-flag", Flags: nil}
		faction := &domain.Faction{ID: "f1", Assets: []*domain.Asset{asset}}
		action := NewUseAssetAbility(nil, nil, nil)
		if action.Validate(faction, nil, rulebook) {
			t.Error("expected Validate false when no A-flagged assets")
		}
	})

	t.Run("A-flagged but not Ready", func(t *testing.T) {
		asset := &domain.Asset{ID: "a1", DefinitionID: "move-asset", Ready: false, Maintained: true}
		faction := &domain.Faction{ID: "f1", Assets: []*domain.Asset{asset}}
		action := NewUseAssetAbility(nil, nil, nil)
		if action.Validate(faction, nil, rulebook) {
			t.Error("expected Validate false when A-flagged asset is not Ready")
		}
	})

	t.Run("A-flagged but not Maintained", func(t *testing.T) {
		asset := &domain.Asset{ID: "a1", DefinitionID: "move-asset", Ready: true, Maintained: false}
		faction := &domain.Faction{ID: "f1", Assets: []*domain.Asset{asset}}
		action := NewUseAssetAbility(nil, nil, nil)
		if action.Validate(faction, nil, rulebook) {
			t.Error("expected Validate false when A-flagged asset is not Maintained")
		}
	})

	t.Run("usable A-flagged asset", func(t *testing.T) {
		asset := &domain.Asset{ID: "a1", DefinitionID: "move-asset", Ready: true, Maintained: true}
		faction := &domain.Faction{ID: "f1", Assets: []*domain.Asset{asset}}
		action := NewUseAssetAbility(nil, nil, nil)
		if !action.Validate(faction, nil, rulebook) {
			t.Error("expected Validate true when A-flagged, Ready, Maintained asset exists")
		}
	})
}

// TestUseAssetAbility_Resolve_Movement_NoCoinCost: movement step with CoinCost=0.
// Expects only AssetMoved, no CoinDelta.
func TestUseAssetAbility_Resolve_Movement_NoCoinCost(t *testing.T) {
	rulebook := makeAbilityRulebook()
	factionState, actingAsset := makeAbilityState("move-asset", false)
	faction := factionState.Factions["f1"]

	collector := &abilityFakeCollector{
		selectedAssets:  []*domain.Asset{actingAsset},
		moveDestination: "Tartarus",
	}

	mutations := runAbility(t, collector, &fixedRoller{values: []int{1}}, faction, factionState, rulebook)

	if len(mutations) != 1 {
		t.Fatalf("len(mutations) = %d, want 1; got %v", len(mutations), mutations)
	}
	moved, ok := mutations[0].(domain.AssetMoved)
	if !ok {
		t.Fatalf("mutations[0] type = %T, want AssetMoved", mutations[0])
	}
	if moved.AssetID != "a1" || moved.FromLocation != "Anchorage" || moved.ToLocation != "Tartarus" {
		t.Errorf("AssetMoved = %+v, want {a1, Anchorage, Tartarus}", moved)
	}
}

// TestUseAssetAbility_Resolve_Movement_WithCoinCost: movement step with CoinCost=2.
// Expects CoinDelta(f1, -2) then AssetMoved.
func TestUseAssetAbility_Resolve_Movement_WithCoinCost(t *testing.T) {
	rulebook := makeAbilityRulebook()
	factionState, actingAsset := makeAbilityState("move-asset-coin", false)
	faction := factionState.Factions["f1"]

	collector := &abilityFakeCollector{
		selectedAssets:  []*domain.Asset{actingAsset},
		moveDestination: "Tartarus",
	}

	mutations := runAbility(t, collector, &fixedRoller{values: []int{1}}, faction, factionState, rulebook)

	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2; got %v", len(mutations), mutations)
	}
	coinDelta, ok := mutations[0].(domain.CoinDelta)
	if !ok || coinDelta.FactionID != "f1" || coinDelta.Delta != -2 {
		t.Errorf("mutations[0] = %v, want CoinDelta{f1, -2}", mutations[0])
	}
	if _, ok := mutations[1].(domain.AssetMoved); !ok {
		t.Errorf("mutations[1] type = %T, want AssetMoved", mutations[1])
	}
}

// TestUseAssetAbility_Resolve_FactionTest_AttackerWins: attacker roll > defense roll.
// Expects CoinDelta(f2, -<rolled>).
// Rolls: attack=8, defense=3, drain=4.
func TestUseAssetAbility_Resolve_FactionTest_AttackerWins(t *testing.T) {
	rulebook := makeAbilityRulebook()
	factionState, actingAsset := makeAbilityState("drain-asset", true)
	faction := factionState.Factions["f1"]
	targetFaction := factionState.Factions["f2"]

	collector := &abilityFakeCollector{
		selectedAssets:    []*domain.Asset{actingAsset},
		factionTestTarget: targetFaction,
	}

	mutations := runAbility(t, collector, &fixedRoller{values: []int{8, 3, 4}}, faction, factionState, rulebook)

	if len(mutations) != 1 {
		t.Fatalf("len(mutations) = %d, want 1; got %v", len(mutations), mutations)
	}
	coinDelta, ok := mutations[0].(domain.CoinDelta)
	if !ok || coinDelta.FactionID != "f2" || coinDelta.Delta != -4 {
		t.Errorf("mutations[0] = %v, want CoinDelta{f2, -4}", mutations[0])
	}
}

// TestUseAssetAbility_Resolve_FactionTest_Tie: equal rolls — tie goes to defender.
// Expects no effect mutations.
// Rolls: attack=5, defense=5.
func TestUseAssetAbility_Resolve_FactionTest_Tie(t *testing.T) {
	rulebook := makeAbilityRulebook()
	factionState, actingAsset := makeAbilityState("drain-asset", true)
	faction := factionState.Factions["f1"]
	targetFaction := factionState.Factions["f2"]

	collector := &abilityFakeCollector{
		selectedAssets:    []*domain.Asset{actingAsset},
		factionTestTarget: targetFaction,
	}

	mutations := runAbility(t, collector, &fixedRoller{values: []int{5, 5}}, faction, factionState, rulebook)

	if len(mutations) != 0 {
		t.Errorf("expected no mutations on tie, got %v", mutations)
	}
}

// TestUseAssetAbility_Resolve_FactionTest_DefenderWins: defense roll > attack roll.
// Expects no effect mutations.
// Rolls: attack=2, defense=9.
func TestUseAssetAbility_Resolve_FactionTest_DefenderWins(t *testing.T) {
	rulebook := makeAbilityRulebook()
	factionState, actingAsset := makeAbilityState("drain-asset", true)
	faction := factionState.Factions["f1"]
	targetFaction := factionState.Factions["f2"]

	collector := &abilityFakeCollector{
		selectedAssets:    []*domain.Asset{actingAsset},
		factionTestTarget: targetFaction,
	}

	mutations := runAbility(t, collector, &fixedRoller{values: []int{2, 9}}, faction, factionState, rulebook)

	if len(mutations) != 0 {
		t.Errorf("expected no mutations when defender wins, got %v", mutations)
	}
}

// TestUseAssetAbility_Resolve_NilAbility: asset with no structured ability.
// Expects ConfirmAbilityApplied to be called and no mutations produced.
func TestUseAssetAbility_Resolve_NilAbility(t *testing.T) {
	rulebook := makeAbilityRulebook()
	factionState, actingAsset := makeAbilityState("deferred-asset", false)
	faction := factionState.Factions["f1"]

	collector := &abilityFakeCollector{
		selectedAssets: []*domain.Asset{actingAsset},
	}

	mutations := runAbility(t, collector, &fixedRoller{values: []int{1}}, faction, factionState, rulebook)

	if !collector.confirmCalled {
		t.Error("expected ConfirmAbilityApplied to be called for nil-ability asset")
	}
	if len(mutations) != 0 {
		t.Errorf("expected no mutations for nil-ability asset, got %v", mutations)
	}
}

package actions

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"go.uber.org/mock/gomock"
)

// makeAbilityRulebook returns a minimal Rulebook with two asset definitions:
//   - "SWN-C1-002": A-flagged, Informers opposed-test (reveal_stealth)
//   - "deferred-asset": A-flagged, no ability (GM adjudication)
func makeAbilityRulebook() *rulebook.Rulebook {
	return &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			"SWN-C1-002": {
				ID:    "SWN-C1-002",
				Name:  "Informers",
				Flags: []domain.AssetFlag{domain.FlagAction},
				HP:    4,
				Ability: &domain.AbilityDefinition{
					Steps: []domain.AbilityStep{{
						Effect: domain.EffectRevealStealth,
						Target: domain.TargetFaction,
						Contest: &domain.Contest{
							AttackerStat: domain.StatCunning,
							DefenderStat: domain.StatCunning,
						},
					}},
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
// optional target faction (f2) with one stealthy asset on world "Anchorage".
func makeAbilityState(actingAssetDefID string, includeTarget bool) (*state.FactionState, *domain.Asset) {
	actingAsset := &domain.Asset{
		ID:           "a1",
		DefinitionID: actingAssetDefID,
		OwnerID:      "f1",
		Location:     domain.Location{WorldID: "Anchorage"},
		CurrentHP:    4,
		Ready:        true,
		Maintained:   true,
	}
	factions := map[string]*domain.Faction{
		"f1": {ID: "f1", Cunning: 0, Assets: map[string]*domain.Asset{"a1": actingAsset}},
	}
	if includeTarget {
		targetAsset := &domain.Asset{
			ID:           "t1",
			DefinitionID: "SWN-C1-002",
			OwnerID:      "f2",
			Location:     domain.Location{WorldID: "Anchorage"},
			CurrentHP:    4,
			Ready:        true,
			Maintained:   true,
			Stealthy:     true,
		}
		factions["f2"] = &domain.Faction{ID: "f2", Cunning: 0, Assets: map[string]*domain.Asset{"t1": targetAsset}}
	}
	return &state.FactionState{Factions: factions}, actingAsset
}

// runAbility drives a full Inputs→Resolve→Output cycle for UseAssetAbility.
func runAbility(t *testing.T, collector *mocks.MockCollector, roller domain.Roller, faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) []domain.Mutation {
	t.Helper()
	act := NewUseAssetAbility(collector, roller, nil, nil)
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
	return mutations
}

func TestUseAssetAbility_Validate(t *testing.T) {
	rulebook := makeAbilityRulebook()

	t.Run("no A-flagged assets", func(t *testing.T) {
		asset := &domain.Asset{ID: "x1", DefinitionID: "no-flag", Ready: true, Maintained: true}
		rulebook.Assets["no-flag"] = &domain.AssetDefinition{ID: "no-flag", Flags: nil}
		faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"x1": asset}}
		act := NewUseAssetAbility(nil, nil, nil, nil)
		if act.Validate(faction, nil, rulebook) {
			t.Error("expected Validate false when no A-flagged assets")
		}
	})

	t.Run("A-flagged but not Ready", func(t *testing.T) {
		asset := &domain.Asset{ID: "a1", DefinitionID: "SWN-C1-002", Ready: false, Maintained: true}
		faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": asset}}
		act := NewUseAssetAbility(nil, nil, nil, nil)
		if act.Validate(faction, nil, rulebook) {
			t.Error("expected Validate false when A-flagged asset is not Ready")
		}
	})

	t.Run("A-flagged but not Maintained", func(t *testing.T) {
		asset := &domain.Asset{ID: "a1", DefinitionID: "SWN-C1-002", Ready: true, Maintained: false}
		faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": asset}}
		act := NewUseAssetAbility(nil, nil, nil, nil)
		if act.Validate(faction, nil, rulebook) {
			t.Error("expected Validate false when A-flagged asset is not Maintained")
		}
	})

	t.Run("A-flagged but destroyed", func(t *testing.T) {
		asset := &domain.Asset{ID: "a1", DefinitionID: "SWN-C1-002", Ready: true, Maintained: true, CurrentHP: 0}
		faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": asset}}
		act := NewUseAssetAbility(nil, nil, nil, nil)
		if act.Validate(faction, nil, rulebook) {
			t.Error("expected Validate false when A-flagged asset has no HP")
		}
	})

	t.Run("usable A-flagged asset", func(t *testing.T) {
		asset := &domain.Asset{ID: "a1", DefinitionID: "SWN-C1-002", Ready: true, Maintained: true, CurrentHP: 1}
		faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": asset}}
		act := NewUseAssetAbility(nil, nil, nil, nil)
		factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction, "f2": {ID: "f2"}}}
		if !act.Validate(faction, factionState, rulebook) {
			t.Error("expected Validate true when A-flagged, Ready, Maintained asset exists")
		}
	})

	t.Run("asset-target ability with no rival asset on its world", func(t *testing.T) {
		rulebook.Assets["asset-target"] = &domain.AssetDefinition{
			ID:    "asset-target",
			Flags: []domain.AssetFlag{domain.FlagAction},
			Ability: &domain.AbilityDefinition{Steps: []domain.AbilityStep{{
				Effect: domain.AbilityEffectType("fixture_asset_effect"),
				Target: domain.TargetAsset,
			}}},
		}
		asset := &domain.Asset{ID: "a1", DefinitionID: "asset-target", OwnerID: "f1", Location: domain.Location{WorldID: "Anchorage"},
			Ready: true, Maintained: true, CurrentHP: 1}
		rival := &domain.Asset{ID: "b1", OwnerID: "f2", Location: domain.Location{WorldID: "Hadrian"}}
		faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": asset}}
		factionState := &state.FactionState{Factions: map[string]*domain.Faction{
			"f1": faction,
			"f2": {ID: "f2", Assets: map[string]*domain.Asset{"b1": rival}},
		}}
		act := NewUseAssetAbility(nil, nil, nil, nil)
		if act.Validate(faction, factionState, rulebook) {
			t.Error("expected Validate false when the asset-target ability has no rival asset on its world")
		}
	})

	t.Run("A-flagged but sabotaged", func(t *testing.T) {
		asset := &domain.Asset{ID: "a1", DefinitionID: "SWN-C1-002", Ready: true, Maintained: true, CurrentHP: 1,
			Conditions: []domain.Condition{{Kind: domain.ConditionNoAbility, SourceAssetID: "sab-1", ExpiresAtTurnOf: "f2"}}}
		faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": asset}}
		act := NewUseAssetAbility(nil, nil, nil, nil)
		if act.Validate(faction, nil, rulebook) {
			t.Error("expected Validate false when the only A-flagged asset carries no_ability")
		}
	})
}

// TestUseAssetAbility_Resolve_Informers_AttackerWins: attacker roll > defense roll.
// Expects AssetStealthCleared for the stealthy target asset.
// Rolls: attack=8, defense=3.
func TestUseAssetAbility_Resolve_Informers_AttackerWins(t *testing.T) {
	rulebook := makeAbilityRulebook()
	factionState, actingAsset := makeAbilityState("SWN-C1-002", true)
	faction := factionState.Factions["f1"]
	targetFaction := factionState.Factions["f2"]

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectAbilityAssets(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*domain.Asset{actingAsset}, nil)
	collector.EXPECT().SelectAbilityTarget(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.AbilityTarget{Kind: domain.TargetFaction, Faction: targetFaction}, nil)

	mutations := runAbility(t, collector, &fixedRoller{values: []int{8, 3}}, faction, factionState, rulebook)

	if len(mutations) != 1 {
		t.Fatalf("len(mutations) = %d, want 1; got %v", len(mutations), mutations)
	}
	cleared, ok := mutations[0].(domain.AssetStealthCleared)
	if !ok {
		t.Fatalf("mutations[0] type = %T, want AssetStealthCleared", mutations[0])
	}
	if cleared.AssetID != "t1" || cleared.FactionID != "f2" {
		t.Errorf("AssetStealthCleared = %+v, want {t1, f2}", cleared)
	}
}

// TestUseAssetAbility_Resolve_Informers_Tie: equal rolls — defender wins, no mutations.
// Rolls: attack=5, defense=5.
func TestUseAssetAbility_Resolve_Informers_Tie(t *testing.T) {
	rulebook := makeAbilityRulebook()
	factionState, actingAsset := makeAbilityState("SWN-C1-002", true)
	faction := factionState.Factions["f1"]
	targetFaction := factionState.Factions["f2"]

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectAbilityAssets(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*domain.Asset{actingAsset}, nil)
	collector.EXPECT().SelectAbilityTarget(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.AbilityTarget{Kind: domain.TargetFaction, Faction: targetFaction}, nil)

	mutations := runAbility(t, collector, &fixedRoller{values: []int{5, 5}}, faction, factionState, rulebook)

	if len(mutations) != 0 {
		t.Errorf("expected no mutations on tie, got %v", mutations)
	}
}

// TestUseAssetAbility_Resolve_Informers_DefenderWins: defense roll > attack roll, no mutations.
// Rolls: attack=2, defense=9.
func TestUseAssetAbility_Resolve_Informers_DefenderWins(t *testing.T) {
	rulebook := makeAbilityRulebook()
	factionState, actingAsset := makeAbilityState("SWN-C1-002", true)
	faction := factionState.Factions["f1"]
	targetFaction := factionState.Factions["f2"]

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectAbilityAssets(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*domain.Asset{actingAsset}, nil)
	collector.EXPECT().SelectAbilityTarget(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.AbilityTarget{Kind: domain.TargetFaction, Faction: targetFaction}, nil)

	mutations := runAbility(t, collector, &fixedRoller{values: []int{2, 9}}, faction, factionState, rulebook)

	if len(mutations) != 0 {
		t.Errorf("expected no mutations when defender wins, got %v", mutations)
	}
}

// TestUseAssetAbility_LaterAssetSeesEarlierCoin: two coin_delta assets in one
// action, each costing 1 Coin, with 1 Coin to spend. The first pays; the second reads
// the action's view, sees 0 Coin and is destroyed instead of paying again.
func TestUseAssetAbility_LaterAssetSeesEarlierCoin(t *testing.T) {
	rulebook := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"postech": {
			ID:    "postech",
			Name:  "Postech",
			Flags: []domain.AssetFlag{domain.FlagAction},
			Ability: &domain.AbilityDefinition{Steps: []domain.AbilityStep{{
				Effect: domain.EffectCoinDelta,
				Target: domain.TargetNone,
				Roll: &domain.RollTable{
					Die:      domain.DiceRoll{NumDice: 1, Sides: 6},
					Outcomes: []domain.AbilityOutcome{{From: 1, To: 1, Coin: -1}},
				},
			}}},
		},
	}}
	first := &domain.Asset{ID: "p1", DefinitionID: "postech", OwnerID: "f1", CurrentHP: 1, Ready: true, Maintained: true}
	second := &domain.Asset{ID: "p2", DefinitionID: "postech", OwnerID: "f1", CurrentHP: 1, Ready: true, Maintained: true}
	faction := &domain.Faction{ID: "f1", Coin: 1, Assets: map[string]*domain.Asset{"p1": first, "p2": second}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectAbilityAssets(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*domain.Asset{first, second}, nil)

	mutations := runAbility(t, collector, &fixedRoller{values: []int{1}}, faction, factionState, rulebook)

	want := []domain.Mutation{
		domain.CoinDelta{FactionID: "f1", Delta: -1, Cause: "ability", CausedByFactionID: "f1"},
		domain.AssetRemoved{FactionID: "f1", AssetID: "p2", Cause: "ability", CausedByFactionID: "f1"},
	}
	if !reflect.DeepEqual(mutations, want) {
		t.Errorf("mutations = %+v, want %+v", mutations, want)
	}
}

// TestUseAssetAbility_SecondMarketersFizzles: two Marketers, one rival asset on the
// world. The first wins its contest and the target declines, so b1 is disabled. The
// second reads the view, finds no candidate and fizzles: no prompt, no error, and the
// first one's condition stands (Decision Record — Execution 1).
func TestUseAssetAbility_SecondMarketersFizzles(t *testing.T) {
	rulebook := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"marketers": {
			ID:    "marketers",
			Name:  "Marketers",
			Flags: []domain.AssetFlag{domain.FlagAction},
			Ability: &domain.AbilityDefinition{Steps: []domain.AbilityStep{{
				Effect:  domain.EffectDisableAsset,
				Target:  domain.TargetAsset,
				Contest: &domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatWealth},
			}}},
		},
		"cost-10": {ID: "cost-10", Cost: 10},
	}}
	at := func(id, defID, owner string) *domain.Asset {
		return &domain.Asset{ID: id, DefinitionID: defID, OwnerID: owner, Location: domain.Location{WorldID: "Anchorage"},
			CurrentHP: 1, Ready: true, Maintained: true}
	}
	first, second := at("m1", "marketers", "f1"), at("m2", "marketers", "f1")
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"m1": first, "m2": second}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{
		"f1": faction,
		"f2": {ID: "f2", Coin: 9, Assets: map[string]*domain.Asset{"b1": at("b1", "cost-10", "f2")}},
	}}

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectAbilityAssets(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*domain.Asset{first, second}, nil)
	collector.EXPECT().SelectAbilityTarget(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ *domain.Asset, _ domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
			return domain.AbilityTarget{Kind: domain.TargetAsset, Asset: candidates.Assets[0]}, nil
		}).Times(1)
	collector.EXPECT().ConfirmPayment(gomock.Any()).Return(false, nil).Times(1)

	mutations := runAbility(t, collector, &fixedRoller{values: []int{8, 3}}, faction, factionState, rulebook)

	want := []domain.Mutation{domain.AssetConditionApplied{
		FactionID:         "f2",
		AssetID:           "b1",
		Condition:         domain.Condition{Kind: domain.ConditionDisabled, SourceAssetID: "m1"},
		Cause:             "ability",
		CausedByFactionID: "f1",
	}}
	if !reflect.DeepEqual(mutations, want) {
		t.Errorf("mutations = %+v, want %+v", mutations, want)
	}
}

// TestUseAssetAbility_TwoMonopoliesSeeTheFirstLevy: f2 has 1 Coin and both
// Monopolies levy it. The first takes the Coin; the second reads the view, finds f2
// cannot pay, and asks it to lose an asset (Decision 3).
func TestUseAssetAbility_TwoMonopoliesSeeTheFirstLevy(t *testing.T) {
	assetRulebook := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"monopoly": {
			ID:    "monopoly",
			Name:  "Monopoly",
			Flags: []domain.AssetFlag{domain.FlagAction},
			Ability: &domain.AbilityDefinition{Steps: []domain.AbilityStep{{
				Effect: domain.EffectLevyCoin,
				Target: domain.TargetFaction,
			}}},
		},
	}}
	at := func(id, defID, owner string) *domain.Asset {
		return &domain.Asset{ID: id, DefinitionID: defID, OwnerID: owner, Location: domain.Location{WorldID: "Anchorage"},
			CurrentHP: 1, Ready: true, Maintained: true}
	}
	first, second := at("m1", "monopoly", "f1"), at("m2", "monopoly", "f1")
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"m1": first, "m2": second}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{
		"f1": faction,
		"f2": {ID: "f2", Coin: 1, Assets: map[string]*domain.Asset{"b1": at("b1", "", "f2")}},
	}}

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectAbilityAssets(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*domain.Asset{first, second}, nil)
	collector.EXPECT().SelectAbilityTarget(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ *domain.Asset, _ domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
			return domain.AbilityTarget{Kind: domain.TargetFaction, Faction: candidates.Factions[0]}, nil
		}).Times(2)
	collector.EXPECT().SelectAssetToLose(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ *domain.Faction, assets []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
			return assets[0], nil
		}).Times(1)

	mutations := runAbility(t, collector, &fixedRoller{values: []int{1}}, faction, factionState, assetRulebook)

	want := []domain.Mutation{
		domain.CoinDelta{FactionID: "f2", Delta: -1, Cause: "ability", CausedByFactionID: "f1"},
		domain.CoinDelta{FactionID: "f1", Delta: 1, Cause: "ability", CausedByFactionID: "f1"},
		domain.AssetRemoved{FactionID: "f2", AssetID: "b1", Cause: "levy_coin", CausedByFactionID: "f1"},
	}
	if !reflect.DeepEqual(mutations, want) {
		t.Errorf("mutations = %+v, want %+v", mutations, want)
	}
}

// TestUseAssetAbility_Resolve_NilAbility: asset with no structured ability.
// Expects ConfirmAbilityApplied to be called and no mutations produced.
func TestUseAssetAbility_Resolve_NilAbility(t *testing.T) {
	rulebook := makeAbilityRulebook()
	factionState, actingAsset := makeAbilityState("deferred-asset", false)
	faction := factionState.Factions["f1"]

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectAbilityAssets(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*domain.Asset{actingAsset}, nil)
	collector.EXPECT().ConfirmAbilityApplied(gomock.Any(), gomock.Any()).Return(false, nil)

	mutations := runAbility(t, collector, &fixedRoller{values: []int{1}}, faction, factionState, rulebook)

	if len(mutations) != 0 {
		t.Errorf("expected no mutations for nil-ability asset, got %v", mutations)
	}
}

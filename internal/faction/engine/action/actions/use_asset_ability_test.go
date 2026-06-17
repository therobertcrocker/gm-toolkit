package actions

import (
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
					AttackerStat: domain.StatCunning,
					DefenderStat: domain.StatCunning,
					Effect:       domain.EffectRevealStealth,
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
	act := NewUseAssetAbility(collector, roller)
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
		act := NewUseAssetAbility(nil, nil)
		if act.Validate(faction, nil, rulebook) {
			t.Error("expected Validate false when no A-flagged assets")
		}
	})

	t.Run("A-flagged but not Ready", func(t *testing.T) {
		asset := &domain.Asset{ID: "a1", DefinitionID: "SWN-C1-002", Ready: false, Maintained: true}
		faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": asset}}
		act := NewUseAssetAbility(nil, nil)
		if act.Validate(faction, nil, rulebook) {
			t.Error("expected Validate false when A-flagged asset is not Ready")
		}
	})

	t.Run("A-flagged but not Maintained", func(t *testing.T) {
		asset := &domain.Asset{ID: "a1", DefinitionID: "SWN-C1-002", Ready: true, Maintained: false}
		faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": asset}}
		act := NewUseAssetAbility(nil, nil)
		if act.Validate(faction, nil, rulebook) {
			t.Error("expected Validate false when A-flagged asset is not Maintained")
		}
	})

	t.Run("usable A-flagged asset", func(t *testing.T) {
		asset := &domain.Asset{ID: "a1", DefinitionID: "SWN-C1-002", Ready: true, Maintained: true}
		faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": asset}}
		act := NewUseAssetAbility(nil, nil)
		if !act.Validate(faction, nil, rulebook) {
			t.Error("expected Validate true when A-flagged, Ready, Maintained asset exists")
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
	collector.EXPECT().SelectFactionTestTarget(gomock.Any(), gomock.Any(), gomock.Any()).Return(targetFaction, nil)

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
	collector.EXPECT().SelectFactionTestTarget(gomock.Any(), gomock.Any(), gomock.Any()).Return(targetFaction, nil)

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
	collector.EXPECT().SelectFactionTestTarget(gomock.Any(), gomock.Any(), gomock.Any()).Return(targetFaction, nil)

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

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectAbilityAssets(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*domain.Asset{actingAsset}, nil)
	collector.EXPECT().ConfirmAbilityApplied(gomock.Any(), gomock.Any()).Return(false, nil)

	mutations := runAbility(t, collector, &fixedRoller{values: []int{1}}, faction, factionState, rulebook)

	if len(mutations) != 0 {
		t.Errorf("expected no mutations for nil-ability asset, got %v", mutations)
	}
}

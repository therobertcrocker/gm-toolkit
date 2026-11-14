package ability

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice/dicetest"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"go.uber.org/mock/gomock"
)

var shelterStep = domain.AbilityStep{
	Effect:  domain.EffectShelter,
	Target:  domain.TargetAsset,
	Shelter: &domain.ShelterParams{AssetTypes: []domain.AssetType{"Operative"}, Capacity: 1},
}

// makeDemiplaneFixture builds f1's Leased Demiplane d1 on Anchorage, with f1's
// Operatives op1 and op3 and Facility fac1 there, Operative op2 on Elsewhere, and
// f2's Operative b2 on Anchorage.
func makeDemiplaneFixture() (*domain.Faction, *domain.Asset, *state.FactionState, *rulebook.Rulebook) {
	demiplane := &domain.AssetDefinition{
		ID:      "demiplane",
		Name:    "Leased Demiplane",
		Flags:   []domain.AssetFlag{domain.FlagAction},
		Ability: &domain.AbilityDefinition{Steps: []domain.AbilityStep{shelterStep}},
	}
	rulebook := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"demiplane": demiplane,
		"operative": {ID: "operative", Type: "Operative"},
		"facility":  {ID: "facility", Type: "Facility"},
	}}
	at := func(id, definitionID, ownerID, worldID string) *domain.Asset {
		return &domain.Asset{ID: id, DefinitionID: definitionID, OwnerID: ownerID, Location: domain.Location{WorldID: worldID}}
	}
	acting := at("d1", "demiplane", "f1", "Anchorage")
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{
		"d1":   acting,
		"op1":  at("op1", "operative", "f1", "Anchorage"),
		"op3":  at("op3", "operative", "f1", "Anchorage"),
		"fac1": at("fac1", "facility", "f1", "Anchorage"),
		"op2":  at("op2", "operative", "f1", "Elsewhere"),
	}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{
		"f1": faction,
		"f2": {ID: "f2", Assets: map[string]*domain.Asset{"b2": at("b2", "operative", "f2", "Anchorage")}},
	}}
	return faction, acting, factionState, rulebook
}

func TestShelterCandidates(t *testing.T) {
	cases := []struct {
		name  string
		setup func(faction *domain.Faction)
		want  []string
	}{
		{name: "own assets of a listed type on its world", want: []string{"op1", "op3"}},
		{name: "an asset sheltered elsewhere is skipped", setup: func(faction *domain.Faction) {
			faction.Assets["op1"].Conditions = []domain.Condition{{Kind: domain.ConditionSheltered, SourceAssetID: "other"}}
		}, want: []string{"op3"}},
		{name: "an asset with a move order is skipped", setup: func(faction *domain.Faction) {
			faction.Assets["op1"].CurrentOrder = &domain.MovementOrder{}
		}, want: []string{"op3"}},
		{name: "a full Demiplane takes no one", setup: func(faction *domain.Faction) {
			faction.Assets["op2"].Conditions = []domain.Condition{{Kind: domain.ConditionSheltered, SourceAssetID: "d1"}}
		}, want: []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			faction, acting, factionState, rulebook := makeDemiplaneFixture()
			if tc.setup != nil {
				tc.setup(faction)
			}
			candidates, picks := stepCandidates(shelterStep, faction, acting, factionState, nil, nil, rulebook)
			if !picks {
				t.Fatal("picks = false, want a pick")
			}
			if got := assetIDs(candidates.Assets); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("candidates = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestShelterCandidates_CountsOtherOwners: a captured Demiplane that still holds
// its old owner's asset is full to its new owner.
func TestShelterCandidates_CountsOtherOwners(t *testing.T) {
	faction, acting, factionState, rulebook := makeDemiplaneFixture()
	factionState.Factions["f2"].Assets["b2"].Conditions = []domain.Condition{{Kind: domain.ConditionSheltered, SourceAssetID: "d1"}}
	candidates, _ := stepCandidates(shelterStep, faction, acting, factionState, nil, nil, rulebook)
	if len(candidates.Assets) != 0 {
		t.Errorf("candidates = %v, want none: the Demiplane is full", assetIDs(candidates.Assets))
	}
}

func TestDispatch_Shelter(t *testing.T) {
	faction, acting, factionState, rulebook := makeDemiplaneFixture()
	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectAbilityTarget(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ domain.AbilityHolder, _ domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
			return domain.AbilityTarget{Kind: domain.TargetAsset, Asset: candidates.Assets[0]}, nil
		})

	view := NewView(factionState)
	mutations, err := Dispatch(faction, domain.AbilityHolder{Asset: acting}, collector, &dicetest.Cycle{Values: []int{1}}, nil, factionState, view, rulebook, nil)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	want := []domain.Mutation{domain.AssetConditionApplied{
		FactionID:         "f1",
		AssetID:           "op1",
		Condition:         domain.Condition{Kind: domain.ConditionSheltered, SourceAssetID: "d1"},
		Cause:             "ability",
		CausedByFactionID: "f1",
	}}
	if !reflect.DeepEqual(mutations, want) {
		t.Errorf("mutations = %+v, want %+v", mutations, want)
	}
}

// TestDispatch_ShelteredActingAssetFizzles: an asset an earlier asset in the same
// action took inside a Demiplane does not use its own ability.
func TestDispatch_ShelteredActingAssetFizzles(t *testing.T) {
	faction, _, factionState, rulebook := makeDemiplaneFixture()
	view := NewView(factionState)
	view.Factions["f1"].Assets["op1"].Conditions = []domain.Condition{{Kind: domain.ConditionSheltered, SourceAssetID: "d1"}}
	rulebook.Assets["operative"].Ability = &domain.AbilityDefinition{Steps: []domain.AbilityStep{shelterStep}}

	collector := mocks.NewMockCollector(gomock.NewController(t))
	mutations, err := Dispatch(faction, domain.AbilityHolder{Asset: faction.Assets["op1"]}, collector, &dicetest.Cycle{Values: []int{1}}, nil, factionState, view, rulebook, nil)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if len(mutations) != 0 {
		t.Errorf("mutations = %+v, want none", mutations)
	}
}

func TestHasTargets_FullShelter(t *testing.T) {
	faction, acting, factionState, rulebook := makeDemiplaneFixture()
	faction.Assets["op1"].Conditions = []domain.Condition{{Kind: domain.ConditionSheltered, SourceAssetID: "d1"}}
	if HasTargets(faction, domain.AbilityHolder{Asset: acting}, &domain.AbilityDefinition{Steps: []domain.AbilityStep{shelterStep}}, factionState, nil, nil, rulebook) {
		t.Error("HasTargets = true, want false: the Demiplane is full")
	}
}

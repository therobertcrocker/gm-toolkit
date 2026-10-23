package ability

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"go.uber.org/mock/gomock"
)

// makeSeditionistsFixture builds f1's Seditionists a1 on Anchorage, holding coin,
// with f2's b2 and f3's c1 there.
func makeSeditionistsFixture(coin int) (*domain.Faction, *domain.Asset, *domain.AssetDefinition, *state.FactionState, *rulebook.Rulebook) {
	def := &domain.AssetDefinition{
		ID:    "seditionists",
		Name:  "Seditionists",
		Flags: []domain.AssetFlag{domain.FlagAction},
		Ability: &domain.AbilityDefinition{Steps: []domain.AbilityStep{{
			Effect: domain.EffectBlockAttack,
			Target: domain.TargetAsset,
			Roll:   &domain.RollTable{Die: domain.DiceRoll{NumDice: 1, Sides: 4}, Outcomes: []domain.AbilityOutcome{}},
		}}},
	}
	rulebook := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{def.ID: def}}
	acting := &domain.Asset{ID: "a1", DefinitionID: def.ID, OwnerID: "f1", Location: domain.Location{WorldID: "Anchorage"}}
	faction := &domain.Faction{ID: "f1", Coin: coin, Assets: map[string]*domain.Asset{"a1": acting}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{
		"f1": faction,
		"f2": {ID: "f2", Assets: map[string]*domain.Asset{"b2": {ID: "b2", OwnerID: "f2", Location: domain.Location{WorldID: "Anchorage"}}}},
		"f3": {ID: "f3", Assets: map[string]*domain.Asset{"c1": {ID: "c1", OwnerID: "f3", Location: domain.Location{WorldID: "Anchorage"}}}},
	}}
	return faction, acting, def, factionState, rulebook
}

func TestDispatch_BlockAttack(t *testing.T) {
	attached := domain.Condition{Kind: domain.ConditionNoAttack, SourceAssetID: "a1"}
	paid := domain.CoinDelta{FactionID: "f1", Delta: -3, Cause: "ability", CausedByFactionID: "f1"}
	appliedToB2 := domain.AssetConditionApplied{FactionID: "f2", AssetID: "b2", Condition: attached, Cause: "ability", CausedByFactionID: "f1"}
	cases := []struct {
		name        string
		coin        int
		alreadyOnC1 bool
		want        []domain.Mutation
	}{
		{
			name: "pays the roll and attaches",
			coin: 5,
			want: []domain.Mutation{paid, appliedToB2},
		},
		{
			name: "a faction that cannot pay the roll attaches nothing (Effort 3 Decision 9)",
			coin: 2,
		},
		{
			name:        "attaching moves this source's attachment off another asset",
			coin:        5,
			alreadyOnC1: true,
			want: []domain.Mutation{
				paid,
				domain.AssetConditionCleared{FactionID: "f3", AssetID: "c1", Condition: attached, Cause: "ability", CausedByFactionID: "f1"},
				appliedToB2,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			faction, acting, def, factionState, rulebook := makeSeditionistsFixture(tc.coin)
			if tc.alreadyOnC1 {
				factionState.Factions["f3"].Assets["c1"].Conditions = []domain.Condition{attached}
			}
			ctrl := gomock.NewController(t)
			collector := mocks.NewMockCollector(ctrl)
			collector.EXPECT().SelectAbilityTarget(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ *domain.Asset, _ domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
					for _, candidate := range candidates.Assets {
						if candidate.ID == "b2" {
							return domain.AbilityTarget{Kind: domain.TargetAsset, Asset: candidate}, nil
						}
					}
					t.Fatalf("b2 not among the candidates %v", assetIDs(candidates.Assets))
					return domain.AbilityTarget{}, nil
				})

			view := NewView(factionState)
			mutations, err := Dispatch(faction, acting, def, collector, &seqRoller{values: []int{3}}, nil, factionState, view, rulebook, nil)
			if err != nil {
				t.Fatalf("Dispatch: %v", err)
			}
			if !reflect.DeepEqual(mutations, tc.want) {
				t.Errorf("mutations = %+v, want %+v", mutations, tc.want)
			}
		})
	}
}

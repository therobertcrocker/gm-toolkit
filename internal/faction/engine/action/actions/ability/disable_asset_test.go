package ability

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"go.uber.org/mock/gomock"
)

// makeMarketersFixture builds f1's Marketers a1 on Anchorage and f2's b2 there,
// with b2 of the given definition and f2 holding coin.
func makeMarketersFixture(targetDefID string, coin int) (*domain.Faction, *domain.Asset, *domain.AssetDefinition, *state.FactionState, *rulebook.Rulebook) {
	def := &domain.AssetDefinition{
		ID:    "marketers",
		Name:  "Marketers",
		Flags: []domain.AssetFlag{domain.FlagAction},
		Ability: &domain.AbilityDefinition{Steps: []domain.AbilityStep{{
			Effect:  domain.EffectDisableAsset,
			Target:  domain.TargetAsset,
			Contest: &domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatWealth},
		}}},
	}
	rulebook := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		def.ID:    def,
		"cost-10": {ID: "cost-10", Cost: 10},
		"cost-1":  {ID: "cost-1", Cost: 1},
	}}
	acting := &domain.Asset{ID: "a1", DefinitionID: def.ID, OwnerID: "f1", Location: domain.Location{WorldID: "Anchorage"}}
	target := &domain.Asset{ID: "b2", DefinitionID: targetDefID, OwnerID: "f2", Location: domain.Location{WorldID: "Anchorage"}}
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": acting}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{
		"f1": faction,
		"f2": {ID: "f2", Coin: coin, Assets: map[string]*domain.Asset{"b2": target}},
	}}
	return faction, acting, def, factionState, rulebook
}

func TestDispatch_DisableAsset(t *testing.T) {
	disabled := domain.AssetConditionApplied{
		FactionID:         "f2",
		AssetID:           "b2",
		Condition:         domain.Condition{Kind: domain.ConditionDisabled, SourceAssetID: "a1"},
		Cause:             "ability",
		CausedByFactionID: "f1",
	}
	pay, decline := true, false
	cases := []struct {
		name        string
		targetDefID string
		coin        int
		rolls       []int
		answer      *bool // nil: ConfirmPayment must not be called
		want        []domain.Mutation
	}{
		{
			name:        "the target pays half the listed cost",
			targetDefID: "cost-10", coin: 9, rolls: []int{8, 3}, answer: &pay,
			want: []domain.Mutation{domain.CoinDelta{FactionID: "f2", Delta: -5, Cause: "ability", CausedByFactionID: "f1"}},
		},
		{
			name:        "the target declines and the asset is disabled",
			targetDefID: "cost-10", coin: 9, rolls: []int{8, 3}, answer: &decline,
			want: []domain.Mutation{disabled},
		},
		{
			name:        "a target that cannot pay is disabled with no prompt",
			targetDefID: "cost-10", coin: 4, rolls: []int{8, 3},
			want: []domain.Mutation{disabled},
		},
		{
			name:        "a price of 0 does nothing and asks nothing (Effort 3 Decision 7)",
			targetDefID: "cost-1", coin: 9, rolls: []int{8, 3},
		},
		{
			name:        "a lost contest asks nothing",
			targetDefID: "cost-10", coin: 9, rolls: []int{2, 9},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			faction, acting, def, factionState, rulebook := makeMarketersFixture(tc.targetDefID, tc.coin)
			ctrl := gomock.NewController(t)
			collector := mocks.NewMockCollector(ctrl)
			collector.EXPECT().SelectAbilityTarget(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ *domain.Asset, _ domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
					return domain.AbilityTarget{Kind: domain.TargetAsset, Asset: candidates.Assets[0]}, nil
				})
			if tc.answer != nil {
				collector.EXPECT().ConfirmPayment(gomock.Any()).
					DoAndReturn(func(offer hooks.PaymentOffer) (bool, error) {
						if offer.Payer.ID != "f2" || offer.Asset.ID != "b2" || offer.Amount != 5 || offer.Reason != hooks.PaymentAvoidDisable {
							t.Errorf("offer = {%s, %s, %d, %s}, want {f2, b2, 5, avoid_disable}", offer.Payer.ID, offer.Asset.ID, offer.Amount, offer.Reason)
						}
						return *tc.answer, nil
					})
			}

			view := NewView(factionState)
			mutations, err := Dispatch(faction, acting, def, collector, &seqRoller{values: tc.rolls}, nil, factionState, view, rulebook, nil)
			if err != nil {
				t.Fatalf("Dispatch: %v", err)
			}
			if !reflect.DeepEqual(mutations, tc.want) {
				t.Errorf("mutations = %+v, want %+v", mutations, tc.want)
			}
		})
	}
}

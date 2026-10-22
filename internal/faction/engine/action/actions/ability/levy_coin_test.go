package ability

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"go.uber.org/mock/gomock"
)

// makeMonopolyFixture builds f1's Monopoly a1 on Anchorage. f2 holds coin and b1
// (Anchorage, stealthed), b2 (Anchorage) and b3 (Hadrian).
func makeMonopolyFixture(coin int) (*domain.Faction, *domain.Asset, *domain.AssetDefinition, *state.FactionState, *rulebook.Rulebook) {
	def := &domain.AssetDefinition{
		ID:    "monopoly",
		Name:  "Monopoly",
		Flags: []domain.AssetFlag{domain.FlagAction},
		Ability: &domain.AbilityDefinition{Steps: []domain.AbilityStep{{
			Effect: domain.EffectLevyCoin,
			Target: domain.TargetFaction,
		}}},
	}
	rulebook := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{def.ID: def}}
	at := func(id, owner, world string, stealthy bool) *domain.Asset {
		return &domain.Asset{ID: id, OwnerID: owner, Location: domain.Location{WorldID: world}, Stealthy: stealthy}
	}
	acting := at("a1", "f1", "Anchorage", false)
	acting.DefinitionID = def.ID
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": acting}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{
		"f1": faction,
		"f2": {ID: "f2", Coin: coin, Assets: map[string]*domain.Asset{
			"b1": at("b1", "f2", "Anchorage", true),
			"b2": at("b2", "f2", "Anchorage", false),
			"b3": at("b3", "f2", "Hadrian", false),
		}},
	}}
	return faction, acting, def, factionState, rulebook
}

func TestDispatch_LevyCoin(t *testing.T) {
	selectF2 := func(collector *mocks.MockCollector) {
		collector.EXPECT().SelectAbilityTarget(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ *domain.Asset, _ domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
				return domain.AbilityTarget{Kind: domain.TargetFaction, Faction: candidates.Factions[0]}, nil
			})
	}

	t.Run("a target that can pay pays the owner 1 Coin, with no prompt", func(t *testing.T) {
		faction, acting, def, factionState, rulebook := makeMonopolyFixture(3)
		collector := mocks.NewMockCollector(gomock.NewController(t))
		selectF2(collector)

		mutations, err := Dispatch(faction, acting, def, collector, &seqRoller{values: []int{1}}, nil, factionState, NewView(factionState), rulebook, nil)
		if err != nil {
			t.Fatalf("Dispatch: %v", err)
		}
		want := []domain.Mutation{
			domain.CoinDelta{FactionID: "f2", Delta: -1, Cause: "ability", CausedByFactionID: "f1"},
			domain.CoinDelta{FactionID: "f1", Delta: 1, Cause: "ability", CausedByFactionID: "f1"},
		}
		if !reflect.DeepEqual(mutations, want) {
			t.Errorf("mutations = %+v, want %+v", mutations, want)
		}
	})

	t.Run("a target that cannot pay loses an asset of its choice on the world, stealthed ones offered", func(t *testing.T) {
		faction, acting, def, factionState, assetRulebook := makeMonopolyFixture(0)
		collector := mocks.NewMockCollector(gomock.NewController(t))
		selectF2(collector)
		collector.EXPECT().SelectAssetToLose(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(asked *domain.Faction, assets []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
				if asked.ID != "f2" {
					t.Errorf("asked faction = %s, want f2", asked.ID)
				}
				if got, want := assetIDs(assets), []string{"b1", "b2"}; !slices.Equal(got, want) {
					t.Errorf("offered assets = %v, want %v", got, want)
				}
				return assets[0], nil
			})

		mutations, err := Dispatch(faction, acting, def, collector, &seqRoller{values: []int{1}}, nil, factionState, NewView(factionState), assetRulebook, nil)
		if err != nil {
			t.Fatalf("Dispatch: %v", err)
		}
		want := []domain.Mutation{domain.AssetRemoved{FactionID: "f2", AssetID: "b1", Cause: "levy_coin", CausedByFactionID: "f1"}}
		if !reflect.DeepEqual(mutations, want) {
			t.Errorf("mutations = %+v, want %+v", mutations, want)
		}
	})

	t.Run("a sheltered asset is not offered for the loss", func(t *testing.T) {
		faction, acting, def, factionState, assetRulebook := makeMonopolyFixture(0)
		factionState.Factions["f2"].Assets["b1"].Conditions = []domain.Condition{{Kind: domain.ConditionSheltered, SourceAssetID: "bolt"}}
		collector := mocks.NewMockCollector(gomock.NewController(t))
		selectF2(collector)
		collector.EXPECT().SelectAssetToLose(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ *domain.Faction, assets []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
				if got, want := assetIDs(assets), []string{"b2"}; !slices.Equal(got, want) {
					t.Errorf("offered assets = %v, want %v", got, want)
				}
				return assets[0], nil
			})

		if _, err := Dispatch(faction, acting, def, collector, &seqRoller{values: []int{1}}, nil, factionState, NewView(factionState), assetRulebook, nil); err != nil {
			t.Fatalf("Dispatch: %v", err)
		}
	})

	t.Run("no asset chosen is an error, not a recoverable skip (Decision Record 5)", func(t *testing.T) {
		faction, acting, def, factionState, rulebook := makeMonopolyFixture(0)
		collector := mocks.NewMockCollector(gomock.NewController(t))
		selectF2(collector)
		collector.EXPECT().SelectAssetToLose(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)

		_, err := Dispatch(faction, acting, def, collector, &seqRoller{values: []int{1}}, nil, factionState, NewView(factionState), rulebook, nil)
		if err == nil || errors.Is(err, action.ErrNoSelection) {
			t.Errorf("err = %v, want a non-sentinel error", err)
		}
	})

	t.Run("an asset off the world is refused", func(t *testing.T) {
		faction, acting, def, factionState, rulebook := makeMonopolyFixture(0)
		collector := mocks.NewMockCollector(gomock.NewController(t))
		selectF2(collector)
		collector.EXPECT().SelectAssetToLose(gomock.Any(), gomock.Any(), gomock.Any()).Return(factionState.Factions["f2"].Assets["b3"], nil)

		mutations, err := Dispatch(faction, acting, def, collector, &seqRoller{values: []int{1}}, nil, factionState, NewView(factionState), rulebook, nil)
		if err == nil {
			t.Errorf("err = nil, mutations = %+v; want an error for b3 on Hadrian", mutations)
		}
	})
}

func TestFactionCandidates(t *testing.T) {
	factionIDs := func(factions []*domain.Faction) []string {
		ids := make([]string, 0, len(factions))
		for _, faction := range factions {
			ids = append(ids, faction.ID)
		}
		slices.Sort(ids)
		return ids
	}

	t.Run("a faction with only a stealthed asset on the world, or only assets elsewhere, is not a Monopoly candidate", func(t *testing.T) {
		factionState, acting := makeAssetTargetState()
		delete(factionState.Factions["f2"].Assets, "b2")
		factionState.Factions["f3"].Assets["c1"].Location.WorldID = "Hadrian"
		if got := factionCandidates(domain.EffectLevyCoin, "f1", acting, factionState); len(got) != 0 {
			t.Errorf("candidates = %v, want none", factionIDs(got))
		}
	})

	t.Run("a faction whose only unstealthed asset on the world is sheltered is not a Monopoly candidate", func(t *testing.T) {
		factionState, acting := makeAssetTargetState()
		factionState.Factions["f2"].Assets["b2"].Conditions = []domain.Condition{{Kind: domain.ConditionSheltered, SourceAssetID: "bolt"}}
		got := factionIDs(factionCandidates(domain.EffectLevyCoin, "f1", acting, factionState))
		if want := []string{"f3"}; !slices.Equal(got, want) {
			t.Errorf("candidates = %v, want %v", got, want)
		}
	})

	t.Run("a faction with an unstealthed asset on the world is a Monopoly candidate", func(t *testing.T) {
		factionState, acting := makeAssetTargetState()
		got := factionIDs(factionCandidates(domain.EffectLevyCoin, "f1", acting, factionState))
		if want := []string{"f2", "f3"}; !slices.Equal(got, want) {
			t.Errorf("candidates = %v, want %v", got, want)
		}
	})

	t.Run("Informers' effect still takes every other faction", func(t *testing.T) {
		factionState, acting := makeAssetTargetState()
		delete(factionState.Factions["f2"].Assets, "b2")
		factionState.Factions["f3"].Assets["c1"].Location.WorldID = "Hadrian"
		got := factionIDs(factionCandidates(domain.EffectRevealStealth, "f1", acting, factionState))
		if want := []string{"f2", "f3"}; !slices.Equal(got, want) {
			t.Errorf("candidates = %v, want %v", got, want)
		}
	})
}

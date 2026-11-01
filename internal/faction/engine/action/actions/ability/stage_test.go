package ability

import (
	"errors"
	"log/slog"
	"maps"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
	"go.uber.org/mock/gomock"
)

const fixtureAssetEffect = domain.AbilityEffectType("fixture_asset_effect")

// makeAssetTargetState builds f1 acting from Anchorage with a1, beside its own a2.
// f2 holds b2 (Anchorage), b1 (Anchorage, stealthed) and b3 (Hadrian); f3 holds c1
// (Anchorage).
func makeAssetTargetState() (*state.FactionState, *domain.Asset) {
	at := func(id, owner, world string, stealthy bool) *domain.Asset {
		return &domain.Asset{ID: id, OwnerID: owner, Location: domain.Location{WorldID: world}, Stealthy: stealthy}
	}
	acting := at("a1", "f1", "Anchorage", false)
	return &state.FactionState{Factions: map[string]*domain.Faction{
		"f1": {ID: "f1", Assets: map[string]*domain.Asset{"a1": acting, "a2": at("a2", "f1", "Anchorage", false)}},
		"f2": {ID: "f2", Assets: map[string]*domain.Asset{
			"b1": at("b1", "f2", "Anchorage", true),
			"b2": at("b2", "f2", "Anchorage", false),
			"b3": at("b3", "f2", "Hadrian", false),
		}},
		"f3": {ID: "f3", Assets: map[string]*domain.Asset{"c1": at("c1", "f3", "Anchorage", false)}},
	}}, acting
}

func assetIDs(assets []*domain.Asset) []string {
	ids := make([]string, 0, len(assets))
	for _, asset := range assets {
		ids = append(ids, asset.ID)
	}
	return ids
}

func TestAssetCandidates(t *testing.T) {
	t.Run("rival unstealthed assets on the acting world, in ID order", func(t *testing.T) {
		factionState, acting := makeAssetTargetState()
		got := assetIDs(assetCandidates(fixtureAssetEffect, "f1", acting, factionState))
		if want := []string{"b2", "c1"}; !slices.Equal(got, want) {
			t.Errorf("candidates = %v, want %v", got, want)
		}
	})

	t.Run("an acting asset in transit has none", func(t *testing.T) {
		factionState, acting := makeAssetTargetState()
		acting.Location.WorldID = ""
		if got := assetCandidates(fixtureAssetEffect, "f1", acting, factionState); got != nil {
			t.Errorf("candidates = %v, want nil", assetIDs(got))
		}
	})

	t.Run("the effect's rule narrows the list", func(t *testing.T) {
		previous := assetTargetRules
		assetTargetRules = maps.Clone(previous)
		assetTargetRules[fixtureAssetEffect] = func(candidate, _ *domain.Asset) bool { return candidate.ID != "c1" }
		t.Cleanup(func() { assetTargetRules = previous })

		factionState, acting := makeAssetTargetState()
		got := assetIDs(assetCandidates(fixtureAssetEffect, "f1", acting, factionState))
		if want := []string{"b2"}; !slices.Equal(got, want) {
			t.Errorf("candidates = %v, want %v", got, want)
		}
	})

	t.Run("an already-disabled rival asset is not a Marketers candidate", func(t *testing.T) {
		factionState, acting := makeAssetTargetState()
		factionState.Factions["f2"].Assets["b2"].Conditions = []domain.Condition{{Kind: domain.ConditionDisabled, SourceAssetID: "m9"}}
		got := assetIDs(assetCandidates(domain.EffectDisableAsset, "f1", acting, factionState))
		if want := []string{"c1"}; !slices.Equal(got, want) {
			t.Errorf("candidates = %v, want %v", got, want)
		}
	})

	t.Run("a rival asset that holds this source's no_attack is not a Seditionists candidate; another source's is", func(t *testing.T) {
		factionState, acting := makeAssetTargetState()
		factionState.Factions["f2"].Assets["b2"].Conditions = []domain.Condition{{Kind: domain.ConditionNoAttack, SourceAssetID: "a1"}}
		factionState.Factions["f3"].Assets["c1"].Conditions = []domain.Condition{{Kind: domain.ConditionNoAttack, SourceAssetID: "s9"}}
		got := assetIDs(assetCandidates(domain.EffectBlockAttack, "f1", acting, factionState))
		if want := []string{"c1"}; !slices.Equal(got, want) {
			t.Errorf("candidates = %v, want %v", got, want)
		}
	})

	t.Run("a sheltered rival asset is not a candidate", func(t *testing.T) {
		factionState, acting := makeAssetTargetState()
		factionState.Factions["f2"].Assets["b2"].Conditions = []domain.Condition{{Kind: domain.ConditionSheltered, SourceAssetID: "bolt"}}
		got := assetIDs(assetCandidates(fixtureAssetEffect, "f1", acting, factionState))
		if want := []string{"c1"}; !slices.Equal(got, want) {
			t.Errorf("candidates = %v, want %v", got, want)
		}
	})
}

func TestResolveTarget_Asset(t *testing.T) {
	step := domain.AbilityStep{Effect: fixtureAssetEffect, Target: domain.TargetAsset}

	t.Run("the target faction is the chosen asset's owner", func(t *testing.T) {
		factionState, acting := makeAssetTargetState()
		chosen := factionState.Factions["f2"].Assets["b2"]
		ctrl := gomock.NewController(t)
		collector := mocks.NewMockCollector(ctrl)
		collector.EXPECT().SelectAbilityTarget(domain.AbilityHolder{Asset: acting}, step, gomock.Any()).
			DoAndReturn(func(_ domain.AbilityHolder, _ domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
				if got, want := assetIDs(candidates.Assets), []string{"b2", "c1"}; !slices.Equal(got, want) {
					t.Errorf("offered candidates = %v, want %v", got, want)
				}
				return domain.AbilityTarget{Kind: domain.TargetAsset, Asset: chosen}, nil
			})

		target, err := resolveTarget(step, factionState.Factions["f1"], domain.AbilityHolder{Asset: acting}, factionState, collector, nil, nil)
		if err != nil {
			t.Fatalf("resolveTarget: %v", err)
		}
		if target.Kind != domain.TargetAsset || target.Asset != chosen || target.Faction != factionState.Factions["f2"] {
			t.Errorf("target = %+v, want {asset, b2, f2}", target)
		}
	})

	t.Run("no candidate left is errNoTargets, with no prompt", func(t *testing.T) {
		factionState, acting := makeAssetTargetState()
		acting.Location.WorldID = "Empty"
		collector := mocks.NewMockCollector(gomock.NewController(t))

		_, err := resolveTarget(step, factionState.Factions["f1"], domain.AbilityHolder{Asset: acting}, factionState, collector, nil, nil)
		if !errors.Is(err, errNoTargets) {
			t.Errorf("err = %v, want errNoTargets", err)
		}
	})

	t.Run("no asset chosen is ErrNoSelection", func(t *testing.T) {
		factionState, acting := makeAssetTargetState()
		ctrl := gomock.NewController(t)
		collector := mocks.NewMockCollector(ctrl)
		collector.EXPECT().SelectAbilityTarget(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(domain.AbilityTarget{Kind: domain.TargetAsset}, nil)

		_, err := resolveTarget(step, factionState.Factions["f1"], domain.AbilityHolder{Asset: acting}, factionState, collector, nil, nil)
		if !errors.Is(err, action.ErrNoSelection) {
			t.Errorf("err = %v, want ErrNoSelection", err)
		}
	})
}

func TestResolveTarget_World(t *testing.T) {
	engine := world.NewWithMap(&relocRouter{
		worlds:     map[string]spatial.RegionHex{"near": relocHex("r1", 1), "mid": relocHex("r1", 2)},
		techLevels: map[string]int{"near": 3, "mid": 4},
	}, slog.New(slog.DiscardHandler))
	base := func(world string) *domain.Base { return &domain.Base{Location: domain.Location{WorldID: world}} }
	faction := &domain.Faction{ID: "f1", Coin: 2, Bases: []*domain.Base{base("near"), base("mid")}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}
	holder := domain.AbilityHolder{Tag: &domain.Tag{ID: "T-013"}}

	pick := func(t *testing.T, world string) (domain.AbilityTarget, error) {
		collector := mocks.NewMockCollector(gomock.NewController(t))
		collector.EXPECT().SelectAbilityTarget(holder, teachStep, gomock.Any()).
			Return(domain.AbilityTarget{Kind: domain.TargetWorld, World: world}, nil)
		return resolveTarget(teachStep, faction, holder, factionState, collector, engine, nil)
	}

	t.Run("a candidate world is the target", func(t *testing.T) {
		target, err := pick(t, "near")
		if err != nil {
			t.Fatalf("resolveTarget: %v", err)
		}
		if target.Kind != domain.TargetWorld || target.World != "near" || target.Faction != faction {
			t.Errorf("target = %+v, want {world, near, f1}", target)
		}
	})

	t.Run("a world the filter left out is refused", func(t *testing.T) {
		if _, err := pick(t, "mid"); err == nil {
			t.Error("resolveTarget accepted mid, which is TL4 already, want an error")
		}
	})
}

func TestHasTargets(t *testing.T) {
	definition := func(steps ...domain.AbilityStep) *domain.AbilityDefinition {
		return &domain.AbilityDefinition{Steps: steps}
	}
	assetStep := domain.AbilityStep{Effect: fixtureAssetEffect, Target: domain.TargetAsset}

	t.Run("asset target with no candidate", func(t *testing.T) {
		factionState, acting := makeAssetTargetState()
		acting.Location.WorldID = "Empty"
		if HasTargets(factionState.Factions["f1"], domain.AbilityHolder{Asset: acting}, definition(assetStep), factionState, nil, nil) {
			t.Error("HasTargets = true, want false")
		}
	})

	t.Run("asset target with a candidate", func(t *testing.T) {
		factionState, acting := makeAssetTargetState()
		if !HasTargets(factionState.Factions["f1"], domain.AbilityHolder{Asset: acting}, definition(assetStep), factionState, nil, nil) {
			t.Error("HasTargets = false, want true")
		}
	})

	t.Run("faction target with one rival", func(t *testing.T) {
		acting := &domain.Asset{ID: "a1", OwnerID: "f1"}
		faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": acting}}
		factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction, "f2": {ID: "f2"}}}
		factionStep := domain.AbilityStep{Effect: domain.EffectRevealStealth, Target: domain.TargetFaction}
		if !HasTargets(faction, domain.AbilityHolder{Asset: acting}, definition(factionStep), factionState, nil, nil) {
			t.Error("HasTargets = false, want true")
		}
	})

	t.Run("only the first step counts: relocate first", func(t *testing.T) {
		factionState, acting := makeAssetTargetState()
		acting.Location.WorldID = "Empty"
		relocateStep := domain.AbilityStep{Effect: domain.EffectRelocate, Target: domain.TargetNone}
		if !HasTargets(factionState.Factions["f1"], domain.AbilityHolder{Asset: acting}, definition(relocateStep, assetStep), factionState, nil, nil) {
			t.Error("HasTargets = false, want true")
		}
	})
}

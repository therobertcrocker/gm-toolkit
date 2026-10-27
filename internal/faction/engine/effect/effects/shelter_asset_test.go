package effects_test

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

var (
	boltholesDef    = &domain.AssetDefinition{ID: "SWN-C5-003", Type: domain.TypeSpecial}
	militaryUnitDef = &domain.AssetDefinition{ID: "SWN-F1-001", Type: domain.TypeMilitaryUnit}
	facilityDef     = &domain.AssetDefinition{ID: "SWN-F2-001", Type: domain.TypeFacility}
)

func placedOn(asset *domain.Asset, worldID string) *domain.Asset {
	asset.Location = domain.Location{WorldID: worldID}
	return asset
}

// boltholesState is alpha's Boltholes bolt and Military Unit mu on Anchorage, and
// beta's asset x there too.
func boltholesState() *state.FactionState {
	beta := placedOn(activeAsset("x", "SWN-F1-001"), "anchorage")
	beta.OwnerID = "beta"
	return &state.FactionState{Factions: map[string]*domain.Faction{
		"alpha": {ID: "alpha", Assets: map[string]*domain.Asset{
			"bolt": placedOn(activeAsset("bolt", "SWN-C5-003"), "anchorage"),
			"mu":   placedOn(activeAsset("mu", "SWN-F1-001"), "anchorage"),
		}},
		"beta":  {ID: "beta", Assets: map[string]*domain.Asset{"x": beta}},
		"gamma": {ID: "gamma", Assets: map[string]*domain.Asset{}},
	}}
}

// registerBoltholes registers Boltholes' shelter_asset entry on alpha's bolt, as
// the SWN data declares it.
func registerBoltholes(factionState *state.FactionState) *hooks.Registry {
	registry := hooks.NewRegistry()
	alpha := factionState.Factions["alpha"]
	effects.Registrars()[domain.SpecialShelterAsset].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Asset: alpha.Assets["bolt"], Def: boltholesDef}, Registry: registry,
		Entry: domain.EffectDefinition{
			Trigger: domain.TriggerLethalDamage,
			Effect:  domain.SpecialShelterAsset,
			Filters: domain.EffectFilters{
				Owner:      domain.OwnerOwn,
				Scope:      domain.ScopeWorld,
				AssetTypes: []domain.AssetType{domain.TypeSpecialForces, domain.TypeMilitaryUnit},
			},
		},
	})
	return registry
}

func shelterOffer(t *testing.T, factionState *state.FactionState, ctx hooks.LethalContext) *hooks.Interception {
	t.Helper()
	interceptors := registerBoltholes(factionState).LethalInterceptorsForAsset("alpha", "bolt")
	if len(interceptors) != 1 {
		t.Fatalf("interceptors = %d, want 1", len(interceptors))
	}
	ctx.FactionState = factionState
	return interceptors[0].Hook.OfferInterception(ctx)
}

var muSheltered = domain.AssetConditionApplied{
	FactionID: "alpha", AssetID: "mu",
	Condition: domain.Condition{Kind: domain.ConditionSheltered, SourceAssetID: "bolt"},
	Cause:     "shelter_asset", CausedByFactionID: "alpha",
}

func TestShelterAsset_Offer(t *testing.T) {
	cases := []struct {
		name    string
		hpAfter int
		want    []domain.Mutation
	}{
		{
			name:    "a hit below 0 HP is set back to 0 and sheltered",
			hpAfter: -5,
			want: []domain.Mutation{
				domain.AssetHPDelta{FactionID: "alpha", AssetID: "mu", Delta: 5, Cause: "shelter_asset", CausedByFactionID: "alpha"},
				muSheltered,
			},
		},
		{name: "a hit to exactly 0 HP is only sheltered", hpAfter: 0, want: []domain.Mutation{muSheltered}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState := boltholesState()
			mu := factionState.Factions["alpha"].Assets["mu"]
			offer := shelterOffer(t, factionState, hooks.LethalContext{Target: mu, TargetDef: militaryUnitDef, HPAfter: tc.hpAfter})
			if offer == nil {
				t.Fatal("offer = nil, want one")
			}
			if offer.Ask != hooks.InterceptAutomatic || offer.Source.ID != "bolt" {
				t.Errorf("Ask = %v, Source = %s; want automatic from bolt", offer.Ask, offer.Source.ID)
			}
			if !reflect.DeepEqual(offer.Mutations, tc.want) {
				t.Errorf("mutations = %+v, want %+v", offer.Mutations, tc.want)
			}
		})
	}
}

func TestShelterAsset_NoOffer(t *testing.T) {
	cases := []struct {
		name  string
		setup func(factionState *state.FactionState) (*domain.Asset, *domain.AssetDefinition)
	}{
		{"a rival's asset", func(factionState *state.FactionState) (*domain.Asset, *domain.AssetDefinition) {
			return factionState.Factions["beta"].Assets["x"], militaryUnitDef
		}},
		{"an asset on another world", func(factionState *state.FactionState) (*domain.Asset, *domain.AssetDefinition) {
			return placedOn(factionState.Factions["alpha"].Assets["mu"], "elsewhere"), militaryUnitDef
		}},
		{"a Facility", func(factionState *state.FactionState) (*domain.Asset, *domain.AssetDefinition) {
			return factionState.Factions["alpha"].Assets["mu"], facilityDef
		}},
		{"the Boltholes itself", func(factionState *state.FactionState) (*domain.Asset, *domain.AssetDefinition) {
			return factionState.Factions["alpha"].Assets["bolt"], boltholesDef
		}},
		{"an unready Boltholes", func(factionState *state.FactionState) (*domain.Asset, *domain.AssetDefinition) {
			factionState.Factions["alpha"].Assets["bolt"].Ready = false
			return factionState.Factions["alpha"].Assets["mu"], militaryUnitDef
		}},
		{"a disabled Boltholes", func(factionState *state.FactionState) (*domain.Asset, *domain.AssetDefinition) {
			bolt := factionState.Factions["alpha"].Assets["bolt"]
			bolt.Conditions = []domain.Condition{{Kind: domain.ConditionDisabled, SourceAssetID: "m1"}}
			return factionState.Factions["alpha"].Assets["mu"], militaryUnitDef
		}},
		{"a Boltholes in transit", func(factionState *state.FactionState) (*domain.Asset, *domain.AssetDefinition) {
			placedOn(factionState.Factions["alpha"].Assets["bolt"], "")
			return placedOn(factionState.Factions["alpha"].Assets["mu"], ""), militaryUnitDef
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState := boltholesState()
			target, targetDef := tc.setup(factionState)
			if offer := shelterOffer(t, factionState, hooks.LethalContext{Target: target, TargetDef: targetDef, HPAfter: -2}); offer != nil {
				t.Errorf("offer = %+v, want nil", offer)
			}
		})
	}
}

func TestShelterAsset_LossTakesShelteredAssets(t *testing.T) {
	sheltered := []domain.Condition{{Kind: domain.ConditionSheltered, SourceAssetID: "bolt"}}
	boltKilled := func(cause string) domain.AssetRemoved {
		return domain.AssetRemoved{FactionID: "alpha", AssetID: "bolt", Cause: cause, CausedByFactionID: "beta"}
	}
	muLost := domain.AssetRemoved{FactionID: "alpha", AssetID: "mu", Cause: "attack", CausedByFactionID: "beta"}
	g1Lost := domain.AssetRemoved{FactionID: "gamma", AssetID: "g1", Cause: "attack", CausedByFactionID: "beta"}
	muShelteredByOther := muSheltered
	muShelteredByOther.Condition.SourceAssetID = "bolt2"
	cases := []struct {
		name          string
		mutations     []domain.Mutation
		disableBolt   bool
		muUnsheltered bool
		want          []domain.Mutation
	}{
		{name: "a combat kill drags every sheltered asset", mutations: []domain.Mutation{boltKilled("attack")}, want: []domain.Mutation{muLost, g1Lost}},
		{name: "a sale drags nothing", mutations: []domain.Mutation{boltKilled("sell")}},
		{
			name:      "an asset the batch already removes is skipped",
			mutations: []domain.Mutation{muLost, boltKilled("attack")},
			want:      []domain.Mutation{g1Lost},
		},
		{name: "a disabled Boltholes still drags", mutations: []domain.Mutation{boltKilled("attack")}, disableBolt: true, want: []domain.Mutation{muLost, g1Lost}},
		{
			name:          "an asset sheltered earlier in the batch is dragged",
			mutations:     []domain.Mutation{muSheltered, boltKilled("attack")},
			muUnsheltered: true,
			want:          []domain.Mutation{muLost, g1Lost},
		},
		{
			name:          "an asset another Boltholes sheltered in the batch is not dragged",
			mutations:     []domain.Mutation{muShelteredByOther, boltKilled("attack")},
			muUnsheltered: true,
			want:          []domain.Mutation{g1Lost},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState := boltholesState()
			if !tc.muUnsheltered {
				factionState.Factions["alpha"].Assets["mu"].Conditions = sheltered
			}
			g1 := placedOn(activeAsset("g1", "SWN-F1-001"), "anchorage")
			g1.OwnerID = "gamma"
			g1.Conditions = sheltered
			factionState.Factions["gamma"].Assets["g1"] = g1
			registry := registerBoltholes(factionState)
			if tc.disableBolt {
				factionState.Factions["alpha"].Assets["bolt"].Conditions = []domain.Condition{{Kind: domain.ConditionDisabled, SourceAssetID: "m1"}}
			}
			var reactors []hooks.MutationReactor
			for _, registered := range registry.MutationReactorsForAsset("alpha", "bolt") {
				reactors = append(reactors, registered.Hook)
			}

			got := runReactors(reactors, tc.mutations, factionState)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("mutations = %+v, want %+v", got, tc.want)
			}
		})
	}
}

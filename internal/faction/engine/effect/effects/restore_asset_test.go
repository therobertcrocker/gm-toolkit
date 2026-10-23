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

var medicalCenterTypes = []domain.AssetType{domain.TypeSpecialForces, domain.TypeMilitaryUnit}

// medicalCenterState is alpha (10 Coin) with Medical Center mc and Military Unit mu
// on Anchorage, and beta's Military Unit x there too.
func medicalCenterState() *state.FactionState {
	x := placedOn(activeAsset("x", "SWN-F1-001"), "anchorage")
	x.OwnerID = "beta"
	return &state.FactionState{Factions: map[string]*domain.Faction{
		"alpha": {ID: "alpha", Coin: 10, Assets: map[string]*domain.Asset{
			"mc": placedOn(activeAsset("mc", "SWN-W4-003"), "anchorage"),
			"mu": placedOn(activeAsset("mu", "SWN-F1-001"), "anchorage"),
		}},
		"beta": {ID: "beta", Assets: map[string]*domain.Asset{"x": x}},
	}}
}

// restoreOffer registers Medical Center's restore_asset entry on alpha's mc, as
// the SWN data declares it, and asks it about ctx.
func restoreOffer(t *testing.T, factionState *state.FactionState, ctx hooks.LethalContext) *hooks.Interception {
	t.Helper()
	registry := hooks.NewRegistry()
	alpha := factionState.Factions["alpha"]
	effects.Registrars()[domain.SpecialRestoreAsset].Register(effect.RegisterContext{
		Faction: alpha, Asset: alpha.Assets["mc"], Def: &domain.AssetDefinition{ID: "SWN-W4-003"}, Registry: registry,
		Entry: domain.EffectDefinition{
			Trigger: domain.TriggerLethalDamage,
			Effect:  domain.SpecialRestoreAsset,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeWorld, AssetTypes: medicalCenterTypes},
			Gates:   domain.EffectGates{OncePer: domain.OncePerAsset},
		},
	})
	interceptors := registry.LethalInterceptorsForAsset("alpha", "mc")
	if len(interceptors) != 1 {
		t.Fatalf("interceptors = %d, want 1", len(interceptors))
	}
	ctx.FactionState = factionState
	return interceptors[0].Hook.OfferInterception(ctx)
}

func TestRestoreAsset_Offer(t *testing.T) {
	factionState := medicalCenterState()
	mu := factionState.Factions["alpha"].Assets["mu"]
	unitDef := &domain.AssetDefinition{ID: "SWN-F1-001", Type: domain.TypeMilitaryUnit, Cost: 6}

	offer := restoreOffer(t, factionState, hooks.LethalContext{Target: mu, TargetDef: unitDef, HPAfter: -2, Damage: 6})

	if offer == nil {
		t.Fatal("offer = nil, want one")
	}
	if offer.Ask != hooks.InterceptRestore || offer.Source.ID != "mc" || offer.Price != 3 {
		t.Errorf("Ask = %v, Source = %s, Price = %d; want restore from mc at 3", offer.Ask, offer.Source.ID, offer.Price)
	}
	if want := "effect:restore_asset:lethal_damage:SWN-W4-003:mc"; offer.BudgetKey != want {
		t.Errorf("BudgetKey = %q, want %q", offer.BudgetKey, want)
	}
	want := []domain.Mutation{
		domain.AssetHPDelta{FactionID: "alpha", AssetID: "mu", Delta: 3, Cause: "restore_asset", CausedByFactionID: "alpha"},
		domain.CoinDelta{FactionID: "alpha", Delta: -3, Cause: "restore_asset", CausedByFactionID: "alpha"},
	}
	if !reflect.DeepEqual(offer.Mutations, want) {
		t.Errorf("mutations = %+v, want %+v", offer.Mutations, want)
	}
}

func TestRestoreAsset_PriceIsAtLeastOne(t *testing.T) {
	factionState := medicalCenterState()
	mu := factionState.Factions["alpha"].Assets["mu"]
	cheapDef := &domain.AssetDefinition{ID: "SWN-F1-001", Type: domain.TypeMilitaryUnit, Cost: 1}

	offer := restoreOffer(t, factionState, hooks.LethalContext{Target: mu, TargetDef: cheapDef, HPAfter: 0, Damage: 4})

	if offer == nil || offer.Price != 1 {
		t.Errorf("offer = %+v, want Price 1", offer)
	}
}

func TestRestoreAsset_NoOffer(t *testing.T) {
	cases := []struct {
		name  string
		setup func(factionState *state.FactionState) (*domain.Asset, *domain.AssetDefinition)
	}{
		{"a Facility", func(factionState *state.FactionState) (*domain.Asset, *domain.AssetDefinition) {
			return factionState.Factions["alpha"].Assets["mu"], facilityDef
		}},
		{"a rival's asset", func(factionState *state.FactionState) (*domain.Asset, *domain.AssetDefinition) {
			return factionState.Factions["beta"].Assets["x"], militaryUnitDef
		}},
		{"an asset on another world", func(factionState *state.FactionState) (*domain.Asset, *domain.AssetDefinition) {
			return placedOn(factionState.Factions["alpha"].Assets["mu"], "elsewhere"), militaryUnitDef
		}},
		{"an unready Medical Center", func(factionState *state.FactionState) (*domain.Asset, *domain.AssetDefinition) {
			factionState.Factions["alpha"].Assets["mc"].Ready = false
			return factionState.Factions["alpha"].Assets["mu"], militaryUnitDef
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState := medicalCenterState()
			target, def := tc.setup(factionState)
			if offer := restoreOffer(t, factionState, hooks.LethalContext{Target: target, TargetDef: def, HPAfter: -1, Damage: 5}); offer != nil {
				t.Errorf("offer = %+v, want nil", offer)
			}
		})
	}
}

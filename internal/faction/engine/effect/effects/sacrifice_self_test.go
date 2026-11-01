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

// falseFrontState is alpha's False Front ff and asset x on Anchorage, and beta's
// asset y there too.
func falseFrontState() *state.FactionState {
	y := placedOn(activeAsset("y", "SWN-F1-001"), "anchorage")
	y.OwnerID = "beta"
	return &state.FactionState{Factions: map[string]*domain.Faction{
		"alpha": {ID: "alpha", Assets: map[string]*domain.Asset{
			"ff": placedOn(activeAsset("ff", "SWN-C1-003"), "anchorage"),
			"x":  placedOn(activeAsset("x", "SWN-F1-001"), "anchorage"),
		}},
		"beta": {ID: "beta", Assets: map[string]*domain.Asset{"y": y}},
	}}
}

// falseFrontOffer registers False Front's sacrifice_self entry on alpha's ff, as
// the SWN data declares it, and asks it about ctx.
func falseFrontOffer(t *testing.T, factionState *state.FactionState, ctx hooks.LethalContext) *hooks.Interception {
	t.Helper()
	registry := hooks.NewRegistry()
	alpha := factionState.Factions["alpha"]
	effects.Registrars()[domain.SpecialSacrificeSelf].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Asset: alpha.Assets["ff"], Def: &domain.AssetDefinition{ID: "SWN-C1-003"}}, Registry: registry,
		Entry: domain.EffectDefinition{
			Trigger: domain.TriggerLethalDamage,
			Effect:  domain.SpecialSacrificeSelf,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeWorld},
		},
	})
	interceptors := registry.LethalInterceptorsForAsset("alpha", "ff")
	if len(interceptors) != 1 {
		t.Fatalf("interceptors = %d, want 1", len(interceptors))
	}
	ctx.FactionState = factionState
	return interceptors[0].Hook.OfferInterception(ctx)
}

func TestSacrificeSelf_Offer(t *testing.T) {
	factionState := falseFrontState()
	x := factionState.Factions["alpha"].Assets["x"]

	offer := falseFrontOffer(t, factionState, hooks.LethalContext{Target: x, TargetDef: militaryUnitDef, HPAfter: -3, Damage: 6})

	if offer == nil {
		t.Fatal("offer = nil, want one")
	}
	if offer.Ask != hooks.InterceptSacrifice || offer.Source.ID != "ff" {
		t.Errorf("Ask = %v, Source = %s; want sacrifice from ff", offer.Ask, offer.Source.ID)
	}
	want := []domain.Mutation{
		domain.AssetHPDelta{FactionID: "alpha", AssetID: "x", Delta: 6, Cause: "sacrifice_self", CausedByFactionID: "alpha"},
		domain.AssetRemoved{FactionID: "alpha", AssetID: "ff", Cause: "sacrifice_self", CausedByFactionID: "alpha"},
	}
	if !reflect.DeepEqual(offer.Mutations, want) {
		t.Errorf("mutations = %+v, want %+v", offer.Mutations, want)
	}
}

func TestSacrificeSelf_NoOffer(t *testing.T) {
	cases := []struct {
		name  string
		setup func(factionState *state.FactionState) *domain.Asset
	}{
		{"the False Front itself", func(factionState *state.FactionState) *domain.Asset {
			return factionState.Factions["alpha"].Assets["ff"]
		}},
		{"a rival's asset", func(factionState *state.FactionState) *domain.Asset {
			return factionState.Factions["beta"].Assets["y"]
		}},
		{"an asset on another world", func(factionState *state.FactionState) *domain.Asset {
			return placedOn(factionState.Factions["alpha"].Assets["x"], "elsewhere")
		}},
		{"an unready False Front", func(factionState *state.FactionState) *domain.Asset {
			factionState.Factions["alpha"].Assets["ff"].Ready = false
			return factionState.Factions["alpha"].Assets["x"]
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState := falseFrontState()
			target := tc.setup(factionState)
			if offer := falseFrontOffer(t, factionState, hooks.LethalContext{Target: target, TargetDef: militaryUnitDef, HPAfter: -1, Damage: 4}); offer != nil {
				t.Errorf("offer = %+v, want nil", offer)
			}
		})
	}
}

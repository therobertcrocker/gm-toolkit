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

// hostileTakeoverState is alpha's Hostile Takeover ht and asset own, and beta's
// asset x.
func hostileTakeoverState() *state.FactionState {
	x := activeAsset("x", "SWN-F1-001")
	x.OwnerID = "beta"
	return &state.FactionState{Factions: map[string]*domain.Faction{
		"alpha": {ID: "alpha", Assets: map[string]*domain.Asset{
			"ht":  activeAsset("ht", "SWN-W7-002"),
			"own": activeAsset("own", "SWN-F1-001"),
		}},
		"beta": {ID: "beta", Assets: map[string]*domain.Asset{"x": x}},
	}}
}

// seizeOffer registers Hostile Takeover's seize_target entry on alpha's ht, as the
// SWN data declares it, and asks it about ctx.
func seizeOffer(t *testing.T, factionState *state.FactionState, ctx hooks.LethalContext) *hooks.Interception {
	t.Helper()
	registry := hooks.NewRegistry()
	alpha := factionState.Factions["alpha"]
	effects.Registrars()[domain.SpecialSeizeTarget].Register(effect.RegisterContext{
		Faction: alpha, Asset: alpha.Assets["ht"], Def: &domain.AssetDefinition{ID: "SWN-W7-002"}, Registry: registry,
		Entry: domain.EffectDefinition{
			Trigger: domain.TriggerLethalDamage,
			Effect:  domain.SpecialSeizeTarget,
			Filters: domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeSelf},
		},
	})
	interceptors := registry.LethalInterceptorsForAsset("alpha", "ht")
	if len(interceptors) != 1 {
		t.Fatalf("interceptors = %d, want 1", len(interceptors))
	}
	ctx.FactionState = factionState
	return interceptors[0].Hook.OfferInterception(ctx)
}

// TestSeizeTarget_Offer: ht's blow kills x, and x joins alpha at 1 HP. The
// LethalContext carries no side, so an attack and a counter read the same; the
// counter case is TestAttack_HostileTakeoverCounterSeizesAttacker.
func TestSeizeTarget_Offer(t *testing.T) {
	factionState := hostileTakeoverState()
	ht := factionState.Factions["alpha"].Assets["ht"]
	x := factionState.Factions["beta"].Assets["x"]

	offer := seizeOffer(t, factionState, hooks.LethalContext{Target: x, TargetDef: militaryUnitDef, Striker: ht, HPAfter: -4, Damage: 8})

	if offer == nil {
		t.Fatal("offer = nil, want one")
	}
	if offer.Ask != hooks.InterceptAutomatic || offer.Source.ID != "ht" {
		t.Errorf("Ask = %v, Source = %s; want automatic from ht", offer.Ask, offer.Source.ID)
	}
	want := []domain.Mutation{
		domain.AssetHPDelta{FactionID: "beta", AssetID: "x", Delta: 5, Cause: "seize_target", CausedByFactionID: "alpha"},
		domain.AssetOwnerChanged{FactionID: "beta", AssetID: "x", ToFactionID: "alpha", Cause: "seize_target", CausedByFactionID: "alpha"},
	}
	if !reflect.DeepEqual(offer.Mutations, want) {
		t.Errorf("mutations = %+v, want %+v", offer.Mutations, want)
	}
}

func TestSeizeTarget_NoOffer(t *testing.T) {
	cases := []struct {
		name  string
		setup func(factionState *state.FactionState) (target, striker *domain.Asset)
	}{
		{"an own target", func(factionState *state.FactionState) (*domain.Asset, *domain.Asset) {
			alpha := factionState.Factions["alpha"]
			return alpha.Assets["own"], alpha.Assets["ht"]
		}},
		{"another striker", func(factionState *state.FactionState) (*domain.Asset, *domain.Asset) {
			return factionState.Factions["beta"].Assets["x"], factionState.Factions["alpha"].Assets["own"]
		}},
		{"no striker (self-damage)", func(factionState *state.FactionState) (*domain.Asset, *domain.Asset) {
			return factionState.Factions["beta"].Assets["x"], nil
		}},
		{"an unready Hostile Takeover", func(factionState *state.FactionState) (*domain.Asset, *domain.Asset) {
			ht := factionState.Factions["alpha"].Assets["ht"]
			ht.Ready = false
			return factionState.Factions["beta"].Assets["x"], ht
		}},
		{"a disabled Hostile Takeover", func(factionState *state.FactionState) (*domain.Asset, *domain.Asset) {
			ht := factionState.Factions["alpha"].Assets["ht"]
			ht.Conditions = []domain.Condition{{Kind: domain.ConditionDisabled, SourceAssetID: "marketers"}}
			return factionState.Factions["beta"].Assets["x"], ht
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState := hostileTakeoverState()
			target, striker := tc.setup(factionState)
			if offer := seizeOffer(t, factionState, hooks.LethalContext{Target: target, TargetDef: militaryUnitDef, Striker: striker, HPAfter: -1, Damage: 5}); offer != nil {
				t.Errorf("offer = %+v, want nil", offer)
			}
		})
	}
}

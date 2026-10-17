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

// zealotReactors registers both Zealots entries on alpha's "zealots" (HP 4), as
// the SWN data declares them, and returns the reactors in registration order.
func zealotReactors(t *testing.T, factionState *state.FactionState) []hooks.MutationReactor {
	t.Helper()
	registry := hooks.NewRegistry()
	alpha := factionState.Factions["alpha"]
	asset := alpha.Assets["zealots"]
	def := &domain.AssetDefinition{ID: "SWN-F3-001"}
	for _, trigger := range []domain.EffectTrigger{domain.TriggerOwnAttackHit, domain.TriggerOwnCounterHit} {
		effects.Registrars()[domain.SpecialSelfDamage].Register(effect.RegisterContext{
			Faction: alpha, Asset: asset, Def: def, Registry: registry, Roller: fixedRoller{value: 3},
			Entry: domain.EffectDefinition{
				Trigger:    trigger,
				Effect:     domain.SpecialSelfDamage,
				Filters:    domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
				SelfDamage: &domain.SelfDamageParams{Amount: domain.DiceRoll{NumDice: 1, Sides: 4}},
			},
		})
	}
	var reactors []hooks.MutationReactor
	for _, registered := range registry.MutationReactorsForAsset("alpha", asset.ID) {
		reactors = append(reactors, registered.Hook)
	}
	return reactors
}

func zealotState() *state.FactionState {
	return &state.FactionState{Factions: map[string]*domain.Faction{
		"alpha": {ID: "alpha", Assets: map[string]*domain.Asset{
			"zealots": activeAsset("zealots", "SWN-F3-001"),
		}},
		"beta": {ID: "beta"},
	}}
}

func zealotHit(counter bool) domain.AssetHPDelta {
	return domain.AssetHPDelta{FactionID: "beta", AssetID: "beta-x", Delta: -5, Cause: "attack", CausedByFactionID: "alpha", CausedByAssetID: "zealots", Counter: counter}
}

func zealotDamage(delta int) domain.AssetHPDelta {
	return domain.AssetHPDelta{FactionID: "alpha", AssetID: "zealots", Delta: delta, Cause: "self_damage", CausedByFactionID: "alpha"}
}

var zealotRemoved = domain.AssetRemoved{FactionID: "alpha", AssetID: "zealots", Cause: "self_damage", CausedByFactionID: "alpha"}

func TestSelfDamage(t *testing.T) {
	cases := []struct {
		name      string
		mutations []domain.Mutation
		want      []domain.Mutation
	}{
		{
			name:      "an attack hit fires one entry, not both",
			mutations: []domain.Mutation{zealotHit(false)},
			want:      []domain.Mutation{zealotDamage(-3)},
		},
		{
			name:      "a counterattack hit fires one entry, not both",
			mutations: []domain.Mutation{zealotHit(true)},
			want:      []domain.Mutation{zealotDamage(-3)},
		},
		{
			name: "a hit redirected to a Base still costs the Zealots",
			mutations: []domain.Mutation{
				domain.BaseHPDelta{FactionID: "beta", BaseID: "beta-base", Delta: -5, Cause: "attack", CausedByFactionID: "alpha", CausedByAssetID: "zealots"},
			},
			want: []domain.Mutation{zealotDamage(-3)},
		},
		{
			name: "a free attack on a new Base during Expand Influence costs the Zealots",
			mutations: []domain.Mutation{
				domain.BaseHPDelta{FactionID: "beta", BaseID: "beta-base", Delta: -5, Cause: "expand", CausedByFactionID: "alpha", CausedByAssetID: "zealots"},
			},
			want: []domain.Mutation{zealotDamage(-3)},
		},
		{
			name: "a hit by another asset costs nothing",
			mutations: []domain.Mutation{
				domain.AssetHPDelta{FactionID: "beta", AssetID: "beta-x", Delta: -5, Cause: "attack", CausedByFactionID: "alpha", CausedByAssetID: "someone-else"},
			},
		},
		{
			name: "damage already in the batch makes the self-damage lethal",
			mutations: []domain.Mutation{
				zealotHit(false),
				domain.AssetHPDelta{FactionID: "alpha", AssetID: "zealots", Delta: -1, Cause: "attack", CausedByFactionID: "beta", CausedByAssetID: "beta-x", Counter: true},
			},
			want: []domain.Mutation{zealotDamage(-3), zealotRemoved},
		},
		{
			name: "Zealots already removed in the batch take no further damage",
			mutations: []domain.Mutation{
				zealotHit(false),
				domain.AssetRemoved{FactionID: "alpha", AssetID: "zealots", Cause: "attack", CausedByFactionID: "beta"},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState := zealotState()
			got := runReactors(zealotReactors(t, factionState), tc.mutations, factionState)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("mutations = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestSelfDamage_InactiveZealotsTakeNothing(t *testing.T) {
	for name, deactivate := range inactiveStates {
		t.Run(name, func(t *testing.T) {
			factionState := zealotState()
			reactors := zealotReactors(t, factionState)
			deactivate(factionState.Factions["alpha"].Assets["zealots"])

			if got := runReactors(reactors, []domain.Mutation{zealotHit(false)}, factionState); len(got) != 0 {
				t.Errorf("mutations = %+v, want none", got)
			}
		})
	}
}

func TestSelfDamage_DoesNotReactToItsOwnDamage(t *testing.T) {
	factionState := zealotState()
	if got := runReactors(zealotReactors(t, factionState), []domain.Mutation{zealotDamage(-3)}, factionState); len(got) != 0 {
		t.Errorf("mutations = %+v, want none: self_damage is not an attack hit", got)
	}
}

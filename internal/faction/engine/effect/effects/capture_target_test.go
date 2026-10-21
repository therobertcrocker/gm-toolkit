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

// treacheryReactors registers Treachery's capture_target entry on alpha's
// "treachery", as the SWN data declares it.
func treacheryReactors(t *testing.T, factionState *state.FactionState) []hooks.MutationReactor {
	t.Helper()
	registry := hooks.NewRegistry()
	alpha := factionState.Factions["alpha"]
	asset := alpha.Assets["treachery"]
	effects.Registrars()[domain.SpecialCaptureTarget].Register(effect.RegisterContext{
		Faction: alpha, Asset: asset, Def: &domain.AssetDefinition{ID: "SWN-C7-003"}, Registry: registry, Roller: fixedRoller{value: 1},
		Entry: domain.EffectDefinition{
			Trigger:       domain.TriggerOwnAttackHit,
			Effect:        domain.SpecialCaptureTarget,
			Filters:       domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
			CaptureTarget: &domain.CaptureTargetParams{Coin: domain.DiceRoll{Modifier: 5}},
		},
	})
	var reactors []hooks.MutationReactor
	for _, registered := range registry.MutationReactorsForAsset("alpha", asset.ID) {
		reactors = append(reactors, registered.Hook)
	}
	return reactors
}

func treacheryState() *state.FactionState {
	return &state.FactionState{Factions: map[string]*domain.Faction{
		"alpha": {ID: "alpha", Assets: map[string]*domain.Asset{
			"treachery": activeAsset("treachery", "SWN-C7-003"),
		}},
		"beta": {ID: "beta", Assets: map[string]*domain.Asset{
			"x": activeAsset("x", "SWN-F1-001"),
		}},
	}}
}

func treacheryHit(strikerID string, counter bool) domain.AssetHPDelta {
	return domain.AssetHPDelta{FactionID: "beta", AssetID: "x", Delta: 0, Cause: "attack", CausedByFactionID: "alpha", CausedByAssetID: strikerID, Counter: counter}
}

var (
	treacherySpent   = domain.AssetRemoved{FactionID: "alpha", AssetID: "treachery", Cause: "capture_target", CausedByFactionID: "alpha"}
	treacheryCoin    = domain.CoinDelta{FactionID: "alpha", Delta: 5, Cause: "capture_target", CausedByFactionID: "alpha"}
	treacheryCapture = domain.AssetOwnerChanged{FactionID: "beta", AssetID: "x", ToFactionID: "alpha", Cause: "capture_target", CausedByFactionID: "alpha"}
)

func TestCaptureTarget(t *testing.T) {
	cases := []struct {
		name      string
		mutations []domain.Mutation
		want      []domain.Mutation
	}{
		{
			name:      "a hit spends Treachery, gains Coin and takes the target",
			mutations: []domain.Mutation{treacheryHit("treachery", false)},
			want:      []domain.Mutation{treacherySpent, treacheryCoin, treacheryCapture},
		},
		{
			name:      "a counter hit takes nothing",
			mutations: []domain.Mutation{treacheryHit("treachery", true)},
		},
		{
			name:      "another striker's hit takes nothing",
			mutations: []domain.Mutation{treacheryHit("someone-else", false)},
		},
		{
			name: "a Base hit takes nothing",
			mutations: []domain.Mutation{
				domain.BaseHPDelta{FactionID: "beta", BaseID: "beta-base", Delta: 0, Cause: "attack", CausedByFactionID: "alpha", CausedByAssetID: "treachery"},
			},
		},
		{
			name: "the capture fizzles when the batch removes the target",
			mutations: []domain.Mutation{
				treacheryHit("treachery", false),
				domain.AssetRemoved{FactionID: "beta", AssetID: "x", Cause: "attack", CausedByFactionID: "alpha"},
			},
		},
		{
			name: "a tie's counter that kills Treachery still lets the capture land (Decision 15)",
			mutations: []domain.Mutation{
				treacheryHit("treachery", false),
				domain.AssetRemoved{FactionID: "alpha", AssetID: "treachery", Cause: "attack", CausedByFactionID: "beta"},
			},
			want: []domain.Mutation{treacheryCoin, treacheryCapture},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState := treacheryState()
			got := runReactors(treacheryReactors(t, factionState), tc.mutations, factionState)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("mutations = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestCaptureTarget_InactiveTreacheryTakesNothing(t *testing.T) {
	for name, deactivate := range inactiveStates {
		t.Run(name, func(t *testing.T) {
			factionState := treacheryState()
			reactors := treacheryReactors(t, factionState)
			deactivate(factionState.Factions["alpha"].Assets["treachery"])

			if got := runReactors(reactors, []domain.Mutation{treacheryHit("treachery", false)}, factionState); len(got) != 0 {
				t.Errorf("mutations = %+v, want none", got)
			}
		})
	}
}

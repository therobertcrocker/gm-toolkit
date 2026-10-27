package effects_test

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// treacheryReactors registers Treachery's capture_target entry on alpha's
// assetID, as the SWN data declares it.
func treacheryReactors(t *testing.T, factionState *state.FactionState, assetID string) []hooks.MutationReactor {
	t.Helper()
	registry := hooks.NewRegistry()
	alpha := factionState.Factions["alpha"]
	asset := alpha.Assets[assetID]
	effects.Registrars()[domain.SpecialCaptureTarget].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Asset: asset, Def: &domain.AssetDefinition{ID: "SWN-C7-003"}}, Registry: registry, Roller: fixedRoller{value: 1},
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
			name: "the capture fizzles when the batch leaves the target at 0 HP (sheltered)",
			mutations: []domain.Mutation{
				domain.AssetHPDelta{FactionID: "beta", AssetID: "x", Delta: -6, Cause: "attack", CausedByFactionID: "alpha", CausedByAssetID: "treachery"},
				domain.AssetHPDelta{FactionID: "beta", AssetID: "x", Delta: 2, Cause: "shelter_asset", CausedByFactionID: "beta"},
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
			got := runReactors(treacheryReactors(t, factionState, "treachery"), tc.mutations, factionState)
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
			reactors := treacheryReactors(t, factionState, "treachery")
			deactivate(factionState.Factions["alpha"].Assets["treachery"])

			if got := runReactors(reactors, []domain.Mutation{treacheryHit("treachery", false)}, factionState); len(got) != 0 {
				t.Errorf("mutations = %+v, want none", got)
			}
		})
	}
}

// TestCaptureTarget_FizzlesOnAMovedTarget: the batch already moves the target (a
// Hostile Takeover seized it), so the Treachery fizzles and is not spent.
func TestCaptureTarget_FizzlesOnAMovedTarget(t *testing.T) {
	factionState := treacheryState()
	reactors := treacheryReactors(t, factionState, "treachery")
	seized := domain.AssetOwnerChanged{FactionID: "beta", AssetID: "x", ToFactionID: "alpha", Cause: "seize_target", CausedByFactionID: "alpha"}

	if got := runReactors(reactors, []domain.Mutation{treacheryHit("treachery", false), seized}, factionState); len(got) != 0 {
		t.Errorf("mutations = %+v, want none", got)
	}
}

// TestCaptureTarget_FirstCaptureWins: two strikers hit one target in a batch. The
// later Treachery fizzles, not spent, only when an earlier active capturer took
// the target (Effort 1 Decision Record 11).
func TestCaptureTarget_FirstCaptureWins(t *testing.T) {
	captureBook := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"SWN-C7-003": {ID: "SWN-C7-003", Effects: []domain.EffectDefinition{{Trigger: domain.TriggerOwnAttackHit, Effect: domain.SpecialCaptureTarget}}},
		"SWN-F1-001": {ID: "SWN-F1-001"},
	}}
	secondCaptures := []domain.Mutation{
		domain.AssetRemoved{FactionID: "alpha", AssetID: "second", Cause: "capture_target", CausedByFactionID: "alpha"},
		treacheryCoin,
		treacheryCapture,
	}
	cases := []struct {
		name            string
		firstDefinition string
		firstInactive   bool
		want            []domain.Mutation
	}{
		{name: "an earlier Treachery took the target", firstDefinition: "SWN-C7-003"},
		{name: "an earlier plain hit does not block", firstDefinition: "SWN-F1-001", want: secondCaptures},
		{name: "an earlier inactive Treachery does not block", firstDefinition: "SWN-C7-003", firstInactive: true, want: secondCaptures},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState := treacheryState()
			alpha := factionState.Factions["alpha"]
			alpha.Assets["first"] = activeAsset("first", tc.firstDefinition)
			alpha.Assets["second"] = activeAsset("second", "SWN-C7-003")
			alpha.Assets["first"].Ready = !tc.firstInactive
			mutations := []domain.Mutation{treacheryHit("first", false), treacheryHit("second", false)}

			got := runReactorsWith(treacheryReactors(t, factionState, "second"), mutations, factionState, captureBook)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("mutations = %+v, want %+v", got, tc.want)
			}
		})
	}
}

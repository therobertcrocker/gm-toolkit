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

func lockdownState() *state.FactionState {
	lockdown := activeAsset("lock", "SWN-C6-001")
	lockdown.Location.WorldID = "Krylos"
	return &state.FactionState{Factions: map[string]*domain.Faction{
		"alpha": {ID: "alpha", Assets: map[string]*domain.Asset{"lock": lockdown}},
		"beta":  {ID: "beta", Assets: map[string]*domain.Asset{"x": activeAsset("x", "SWN-F1-001")}},
	}}
}

func lockOutReactor(t *testing.T, factionState *state.FactionState) hooks.MutationReactor {
	t.Helper()
	registry := hooks.NewRegistry()
	alpha := factionState.Factions["alpha"]
	effects.Registrars()[domain.SpecialLockOut].Register(effect.RegisterContext{
		Faction: alpha, Asset: alpha.Assets["lock"], Def: &domain.AssetDefinition{ID: "SWN-C6-001"}, Registry: registry,
		Entry: domain.EffectDefinition{
			Trigger: domain.TriggerOwnAttackHit,
			Effect:  domain.SpecialLockOut,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
		},
	})
	reactors := registry.MutationReactorsForAsset("alpha", "lock")
	if len(reactors) != 1 {
		t.Fatalf("reactors = %d, want 1", len(reactors))
	}
	return reactors[0].Hook
}

var lockedOutBeta = domain.AssetConditionApplied{
	FactionID: "alpha", AssetID: "lock",
	Condition: domain.Condition{Kind: domain.ConditionLockedOut, SourceAssetID: "lock", TargetFactionID: "beta", WorldID: "Krylos"},
	Cause:     "lock_out", CausedByFactionID: "alpha",
}

func TestLockOut(t *testing.T) {
	assetHit := domain.AssetHPDelta{FactionID: "beta", AssetID: "x", Delta: 0, Cause: "attack", CausedByFactionID: "alpha", CausedByAssetID: "lock"}
	baseHit := domain.BaseHPDelta{FactionID: "beta", BaseID: "b", Delta: 0, Cause: "attack", CausedByFactionID: "alpha", CausedByAssetID: "lock"}
	counterHit := assetHit
	counterHit.Counter = true
	cases := []struct {
		name      string
		mutations []domain.Mutation
		want      []domain.Mutation
	}{
		{"a hit on an asset", []domain.Mutation{assetHit}, []domain.Mutation{lockedOutBeta}},
		{"a hit on a Base", []domain.Mutation{baseHit}, []domain.Mutation{lockedOutBeta}},
		{"two hits on one faction lock once", []domain.Mutation{assetHit, baseHit}, []domain.Mutation{lockedOutBeta}},
		{"a counter is not an attack", []domain.Mutation{counterHit}, nil},
		{"a Lockdown the batch removes locks nothing", []domain.Mutation{assetHit, domain.AssetRemoved{FactionID: "alpha", AssetID: "lock", Cause: "attack"}}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState := lockdownState()

			got := lockOutReactor(t, factionState).OnMutations(tc.mutations, factionState, nil)

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("OnMutations = %+v, want %+v", got, tc.want)
			}
		})
	}
}

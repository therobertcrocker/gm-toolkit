package effects_test

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
)

func buyStealthed(asset *domain.Asset) []domain.Mutation {
	alpha := &domain.Faction{ID: "alpha"}
	return effects.Registrars()[domain.SpecialApplyStealth].Fire(effect.FireContext{
		Holder: effect.Holder{Faction: alpha, Asset: asset, Def: &domain.AssetDefinition{ID: "SWN-F5-003"}}, Asset: asset,
		Entry: domain.EffectDefinition{Trigger: domain.TriggerOnPurchase, Effect: domain.SpecialApplyStealth, Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf}},
	})
}

func TestApplyStealth(t *testing.T) {
	bought := domain.AssetStealthApplied{FactionID: "alpha", AssetID: "assassins", Cause: "apply_stealth", CausedByFactionID: "alpha"}
	cases := []struct {
		name  string
		asset *domain.Asset
		want  []domain.Mutation
	}{
		{
			name:  "a bought asset is stealthed",
			asset: activeAsset("assassins", "SWN-F5-003"),
			want:  []domain.Mutation{bought},
		},
		{
			name:  "a bought asset that is not ready is still stealthed",
			asset: &domain.Asset{ID: "assassins", DefinitionID: "SWN-F5-003", OwnerID: "alpha", CurrentHP: 4, Ready: false},
			want:  []domain.Mutation{bought},
		},
		{
			name:  "an asset already stealthy emits nothing",
			asset: &domain.Asset{ID: "assassins", DefinitionID: "SWN-F5-003", OwnerID: "alpha", CurrentHP: 4, Stealthy: true},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := buyStealthed(tc.asset); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("mutations = %+v, want %+v", got, tc.want)
			}
		})
	}
}

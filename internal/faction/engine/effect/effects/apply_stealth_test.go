package effects_test

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
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

// TestApplyStealth_SecretiveStealthsEveryPurchase: a tag entry at faction scope
// passes the engine's check and stealths any asset its faction buys.
func TestApplyStealth_SecretiveStealthsEveryPurchase(t *testing.T) {
	secretive := &domain.Tag{ID: "SWN-T-017", Name: "Secretive", Effects: []domain.EffectDefinition{{
		Trigger: domain.TriggerOnPurchase,
		Effect:  domain.SpecialApplyStealth,
		Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
	}}}
	engine := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler))
	if err := engine.CheckTagEffects(secretive); err != nil {
		t.Fatalf("CheckTagEffects: %v", err)
	}
	rb := &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{"SWN-W1-001": {ID: "SWN-W1-001"}},
		Tags:   map[string]*domain.Tag{secretive.ID: secretive},
	}
	alpha := &domain.Faction{ID: "alpha", Tags: []string{secretive.ID}}
	bought := &domain.Asset{ID: "franchise", DefinitionID: "SWN-W1-001", OwnerID: "alpha"}

	got := engine.FirePurchase(alpha, bought, nil, rb, nil, nil)

	want := []domain.Mutation{domain.AssetStealthApplied{FactionID: "alpha", AssetID: "franchise", Cause: "apply_stealth", CausedByFactionID: "alpha"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mutations = %+v, want %+v", got, want)
	}
}

func TestApplyStealth_ValidateRefusesFactionScopeOnAnAsset(t *testing.T) {
	cases := map[string]struct {
		scope   domain.EffectScope
		wantErr bool
	}{
		"faction scope is refused": {scope: domain.ScopeFaction, wantErr: true},
		"self scope is accepted":   {scope: domain.ScopeSelf, wantErr: false},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			definition := &domain.AssetDefinition{ID: "stealthy", Effects: []domain.EffectDefinition{{
				Trigger: domain.TriggerOnPurchase,
				Effect:  domain.SpecialApplyStealth,
				Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: testCase.scope},
			}}}
			err := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler)).RegisterEffectsList(definition)
			if (err != nil) != testCase.wantErr {
				t.Errorf("RegisterEffectsList error = %v, wantErr %v", err, testCase.wantErr)
			}
		})
	}
}

package effects_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
)

func TestDriftRatingBonus_RaisesTheOwnersAssets(t *testing.T) {
	runners := &domain.Tag{ID: "HAWK-T-012", Name: "Drift-Runners"}
	alpha := &domain.Faction{ID: "alpha", Tags: []string{runners.ID}}
	beta := &domain.Faction{ID: "beta"}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialDriftRatingBonus].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Tag: runners},
		Entry: domain.EffectDefinition{
			Trigger:          domain.TriggerDriftRatingQuery,
			Effect:           domain.SpecialDriftRatingBonus,
			Filters:          domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
			DriftRatingBonus: &domain.DriftRatingBonusParams{Amount: 1},
		},
		Registry: registry,
	})
	unrated := &domain.AssetDefinition{ID: "unrated"}
	rated := &domain.AssetDefinition{ID: "rated", DriftRating: 3}

	cases := []struct {
		name       string
		faction    *domain.Faction
		definition *domain.AssetDefinition
		want       int
	}{
		{"a missing rating counts as 1, then gains the bonus", alpha, unrated, 2},
		{"a rated asset gains the bonus", alpha, rated, 4},
		{"a rival's asset keeps its rating", beta, rated, 3},
		{"a rival's unrated asset counts as 1", beta, unrated, 1},
	}
	for _, tc := range cases {
		if got := dispatch.AssetDriftRatingResolver(registry, tc.faction)(tc.definition); got != tc.want {
			t.Errorf("%s: drift rating = %d, want %d", tc.name, got, tc.want)
		}
	}

	alpha.Tags = nil
	if got := dispatch.AssetDriftRatingResolver(registry, alpha)(rated); got != 3 {
		t.Errorf("drift rating after the tag is dropped = %d, want 3", got)
	}
	if got := dispatch.AssetDriftRatingResolver(nil, alpha)(unrated); got != 1 {
		t.Errorf("nil registry, unrated asset: drift rating = %d, want 1", got)
	}
}

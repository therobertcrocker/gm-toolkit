package effects_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
)

func TestSpeedFloor_RaisesTheOwnersSlowAssets(t *testing.T) {
	mercenaries := &domain.Tag{ID: "SWN-T-008", Name: "Mercenary Group"}
	alpha := &domain.Faction{ID: "alpha", Tags: []string{mercenaries.ID}}
	beta := &domain.Faction{ID: "beta"}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialSpeedFloor].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Tag: mercenaries},
		Entry: domain.EffectDefinition{
			Trigger:    domain.TriggerSpeedQuery,
			Effect:     domain.SpecialSpeedFloor,
			Filters:    domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
			SpeedFloor: &domain.SpeedFloorParams{Speed: 1},
		},
		Registry: registry,
	})
	facility := &domain.AssetDefinition{ID: "facility"}
	starship := &domain.AssetDefinition{ID: "starship", Speed: 3}

	cases := []struct {
		name    string
		faction *domain.Faction
		def     *domain.AssetDefinition
		want    int
	}{
		{"a speed-0 asset gains the floor", alpha, facility, 1},
		{"a faster asset keeps its speed", alpha, starship, 3},
		{"a rival's asset is not raised", beta, facility, 0},
	}
	for _, tc := range cases {
		if got := dispatch.ResolveAssetSpeed(registry, tc.faction, tc.def); got != tc.want {
			t.Errorf("%s: speed = %d, want %d", tc.name, got, tc.want)
		}
	}

	alpha.Tags = nil
	if got := dispatch.ResolveAssetSpeed(registry, alpha, facility); got != 0 {
		t.Errorf("speed after the tag is dropped = %d, want 0", got)
	}
}

package effects_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
)

// TestSpeedBonus_RunningHot registers Aether Burners' two hot speed entries: +5
// for Aetherships, and a floor of 5 for everything.
func TestSpeedBonus_RunningHot(t *testing.T) {
	burners := &domain.Tag{ID: "HAWK-T-003", Name: "Aether Burners"}
	alpha := &domain.Faction{ID: "alpha", Tags: []string{burners.ID}}
	beta := &domain.Faction{ID: "beta"}
	registry := hooks.NewRegistry()
	holder := effect.Holder{Faction: alpha, Tag: burners}
	effects.Registrars()[domain.SpecialSpeedBonus].Register(effect.RegisterContext{
		Holder: holder, Registry: registry,
		Entry: domain.EffectDefinition{
			Trigger:    domain.TriggerSpeedQuery,
			Effect:     domain.SpecialSpeedBonus,
			Filters:    domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, AssetTypes: []domain.AssetType{"Aethership"}, Hot: true},
			SpeedBonus: &domain.SpeedBonusParams{Amount: 5},
		},
	})
	effects.Registrars()[domain.SpecialSpeedFloor].Register(effect.RegisterContext{
		Holder: holder, Registry: registry,
		Entry: domain.EffectDefinition{
			Trigger:    domain.TriggerSpeedQuery,
			Effect:     domain.SpecialSpeedFloor,
			Filters:    domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Hot: true},
			SpeedFloor: &domain.SpeedFloorParams{Speed: 5},
		},
	})
	frigate := &domain.AssetDefinition{ID: "frigate", Type: "Aethership", Speed: 14}
	troops := &domain.AssetDefinition{ID: "troops", Type: "Operative"}

	cases := []struct {
		name       string
		faction    *domain.Faction
		definition *domain.AssetDefinition
		hot        bool
		want       int
	}{
		{"an Aethership running hot gains 5", alpha, frigate, true, 19},
		{"an Aethership at normal speed gains nothing", alpha, frigate, false, 14},
		{"troops running hot move at the floor", alpha, troops, true, 5},
		{"troops at normal speed do not move", alpha, troops, false, 0},
		{"a rival's Aethership gains nothing", beta, frigate, true, 14},
	}
	for _, tc := range cases {
		if got := dispatch.AssetSpeedResolver(registry, tc.faction)(tc.definition, tc.hot); got != tc.want {
			t.Errorf("%s: speed = %d, want %d", tc.name, got, tc.want)
		}
	}

	alpha.LapsedTags = []string{burners.ID}
	if got := dispatch.AssetSpeedResolver(registry, alpha)(frigate, true); got != 14 {
		t.Errorf("speed while the tag is lapsed = %d, want 14", got)
	}
}

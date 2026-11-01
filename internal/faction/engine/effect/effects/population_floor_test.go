package effects_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
)

func registerPopulationFloor(registry *hooks.Registry, faction *domain.Faction, tag *domain.Tag, percent int) {
	effects.Registrars()[domain.SpecialPopulationFloor].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: faction, Tag: tag},
		Entry: domain.EffectDefinition{
			Trigger: domain.TriggerPopulationFloorQuery,
			Effect:  domain.SpecialPopulationFloor,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction,
				AssetTypes: []domain.AssetType{domain.TypeStarship}},
			PopulationFloor: &domain.PopulationFloorParams{Percent: percent},
		},
		Registry: registry,
	})
}

func TestPopulationFloor_TagHolderLowersItsTypesFloor(t *testing.T) {
	colonists := &domain.Tag{ID: "SWN-T-001", Name: "Colonists"}
	alpha := &domain.Faction{ID: "alpha", Tags: []string{colonists.ID}}
	beta := &domain.Faction{ID: "beta"}
	registry := hooks.NewRegistry()
	registerPopulationFloor(registry, alpha, colonists, 20)
	floors := map[domain.AssetType]int{domain.TypeStarship: 500000, domain.TypeFacility: 1000}
	starship := &domain.AssetDefinition{ID: "ship", Type: domain.TypeStarship}
	facility := &domain.AssetDefinition{ID: "plant", Type: domain.TypeFacility}

	cases := []struct {
		name    string
		faction *domain.Faction
		def     *domain.AssetDefinition
		want    int
	}{
		{"the holder's Starship floor is 20%", alpha, starship, 100000},
		{"another type keeps its floor", alpha, facility, 1000},
		{"a rival keeps the setting's floor", beta, starship, 500000},
	}
	for _, tc := range cases {
		if got := dispatch.ResolvePopulationFloor(registry, tc.faction, tc.def, floors); got != tc.want {
			t.Errorf("%s: floor = %d, want %d", tc.name, got, tc.want)
		}
	}

	alpha.Tags = nil
	if got := dispatch.ResolvePopulationFloor(registry, alpha, starship, floors); got != 500000 {
		t.Errorf("floor after the tag is dropped = %d, want 500000", got)
	}
}

func TestPopulationFloor_TheLowestFloorWins(t *testing.T) {
	colonists := &domain.Tag{ID: "SWN-T-001", Name: "Colonists"}
	expertise := &domain.Tag{ID: "SWN-T-018", Name: "Technical Expertise"}
	alpha := &domain.Faction{ID: "alpha", Tags: []string{expertise.ID, colonists.ID}}
	registry := hooks.NewRegistry()
	registerPopulationFloor(registry, alpha, expertise, 2)
	registerPopulationFloor(registry, alpha, colonists, 20)
	floors := map[domain.AssetType]int{domain.TypeStarship: 500000}
	starship := &domain.AssetDefinition{ID: "ship", Type: domain.TypeStarship}

	if got := dispatch.ResolvePopulationFloor(registry, alpha, starship, floors); got != 10000 {
		t.Errorf("floor = %d, want 10000 (2%% beats 20%% in either order)", got)
	}
}

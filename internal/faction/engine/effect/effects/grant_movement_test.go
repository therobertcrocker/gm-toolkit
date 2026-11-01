package effects_test

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

var transitWebDef = &domain.AssetDefinition{ID: "SWN-W7-003", Name: "Transit Web"}

func transitWebRelocate() domain.RelocateParams {
	return domain.RelocateParams{
		Scope:        domain.RelocateOther,
		Range:        domain.RangeHexes,
		MaxHex:       3,
		CoinCost:     1,
		CoinCostMode: domain.RelocateCostFlat,
		Categories:   []domain.FactionStat{domain.StatCunning, domain.StatWealth},
		ExcludeTypes: []domain.AssetType{domain.TypeStarship},
	}
}

// registerWeb gives alpha one active web on Tartarus and registers its grant.
func registerWeb(t *testing.T) (*hooks.Registry, *domain.Faction, *domain.Faction) {
	t.Helper()
	web := activeAsset("web", transitWebDef.ID)
	web.Location = tartarus
	alpha := &domain.Faction{ID: "alpha", Assets: map[string]*domain.Asset{web.ID: web}}
	beta := &domain.Faction{ID: "beta", Assets: map[string]*domain.Asset{}}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialGrantMovement].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Asset: web, Def: transitWebDef},
		Entry: domain.EffectDefinition{
			Trigger:       domain.TriggerMovementOptionsQuery,
			Effect:        domain.SpecialGrantMovement,
			Filters:       domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
			GrantMovement: &domain.GrantMovementParams{Relocate: transitWebRelocate()},
		},
		Registry: registry,
	})
	return registry, alpha, beta
}

func webGrants(registry *hooks.Registry, faction *domain.Faction) []hooks.MovementAbility {
	var abilities []hooks.MovementAbility
	for _, registered := range registry.AssetMovementGrantersForAsset("alpha", "web") {
		abilities = append(abilities, registered.Hook.GrantMovementAbilities(faction)...)
	}
	return abilities
}

func TestGrantMovement_RegistersAtAssetScope(t *testing.T) {
	registry, _, _ := registerWeb(t)
	if got := registry.AssetMovementGrantersForAsset("alpha", "web"); len(got) != 1 {
		t.Fatalf("asset-scope granters = %d, want 1", len(got))
	}
	if got := registry.AssetMovementGrantersFor("alpha", ""); len(got) != 0 {
		t.Errorf("global + faction granters = %d, want 0 (the grant is asset-scoped)", len(got))
	}
}

func TestGrantMovement_GrantsItsRelocateBlockFromItsOwnAsset(t *testing.T) {
	registry, alpha, _ := registerWeb(t)
	abilities := webGrants(registry, alpha)
	if len(abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(abilities))
	}
	if abilities[0].Source != alpha.Assets["web"] {
		t.Errorf("Source = %v, want the live web asset", abilities[0].Source)
	}
	if !reflect.DeepEqual(abilities[0].Relocate, transitWebRelocate()) {
		t.Errorf("Relocate = %+v, want the entry's relocate block", abilities[0].Relocate)
	}
}

func TestGrantMovement_RivalFactionGetsNothing(t *testing.T) {
	registry, _, beta := registerWeb(t)
	if got := webGrants(registry, beta); got != nil {
		t.Errorf("abilities for beta = %v, want nil", got)
	}
}

func TestGrantMovement_InactiveAssetGrantsNothing(t *testing.T) {
	for name, deactivate := range inactiveStates {
		t.Run(name, func(t *testing.T) {
			registry, alpha, _ := registerWeb(t)
			deactivate(alpha.Assets["web"])
			if got := webGrants(registry, alpha); got != nil {
				t.Errorf("abilities = %v, want nil while the web is inactive", got)
			}
		})
	}
}

func TestGrantMovement_RemovedAssetGrantsNothing(t *testing.T) {
	registry, alpha, _ := registerWeb(t)
	delete(alpha.Assets, "web")
	if got := webGrants(registry, alpha); got != nil {
		t.Errorf("abilities = %v, want nil after the web is removed", got)
	}
}

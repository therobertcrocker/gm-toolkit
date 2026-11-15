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
		ExcludeTypes: []domain.AssetType{domain.AssetType("Starship")},
	}
}

// registerWeb gives alpha one active web on Tartarus and registers its grant.
func registerWeb(t *testing.T) (*hooks.Registry, *domain.Faction, *domain.Faction) {
	t.Helper()
	return registerWebGrant(t, domain.GrantMovementParams{Relocate: transitWebRelocate()})
}

func registerPublicWeb(t *testing.T) (*hooks.Registry, *domain.Faction, *domain.Faction) {
	t.Helper()
	return registerWebGrant(t, domain.GrantMovementParams{Relocate: transitWebRelocate(), Public: true})
}

func registerWebGrant(t *testing.T, grant domain.GrantMovementParams) (*hooks.Registry, *domain.Faction, *domain.Faction) {
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
			GrantMovement: &grant,
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

func oneGrant(t *testing.T, abilities []hooks.MovementAbility) hooks.MovementAbility {
	t.Helper()
	if len(abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(abilities))
	}
	return abilities[0]
}

func TestGrantMovement_PublicGrantServesARivalForAFare(t *testing.T) {
	registry, alpha, beta := registerPublicWeb(t)
	grant := oneGrant(t, webGrants(registry, beta))
	if grant.Source != alpha.Assets["web"] {
		t.Errorf("Source = %v, want alpha's web", grant.Source)
	}
	if grant.Relocate.CoinCost != 1 || grant.Payee != "alpha" {
		t.Errorf("CoinCost %d, Payee %q; want 1 paid to alpha", grant.Relocate.CoinCost, grant.Payee)
	}
}

func TestGrantMovement_PublicGrantIsFreeForItsOwner(t *testing.T) {
	registry, alpha, _ := registerPublicWeb(t)
	grant := oneGrant(t, webGrants(registry, alpha))
	if grant.Relocate.CoinCost != 0 || grant.Payee != "" {
		t.Errorf("CoinCost %d, Payee %q; want a free ride with no payee", grant.Relocate.CoinCost, grant.Payee)
	}
}

func TestGrantMovement_HiddenPublicGrantServesOnlyItsOwner(t *testing.T) {
	registry, alpha, beta := registerPublicWeb(t)
	alpha.Assets["web"].Stealthy = true
	if got := webGrants(registry, beta); got != nil {
		t.Errorf("rival abilities = %v, want nil while the web is stealthed", got)
	}
	oneGrant(t, webGrants(registry, alpha))
}

func TestGrantMovement_PrivateGrantKeepsItsCost(t *testing.T) {
	registry, alpha, _ := registerWeb(t)
	grant := oneGrant(t, webGrants(registry, alpha))
	if grant.Relocate.CoinCost != 1 || grant.Payee != "" {
		t.Errorf("CoinCost %d, Payee %q; want 1 spent with no payee", grant.Relocate.CoinCost, grant.Payee)
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

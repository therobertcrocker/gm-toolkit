package effects_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

var (
	laboratoryDef = &domain.AssetDefinition{ID: "SWN-W3-002", Name: "Laboratory"}
	wealthAsset   = &domain.AssetDefinition{ID: "bought-wealth", Category: domain.StatWealth}
	forceAsset    = &domain.AssetDefinition{ID: "bought-force", Category: domain.StatForce}
)

func techLevelFloorEntry(scope domain.EffectScope, level int, categories ...domain.FactionStat) domain.EffectDefinition {
	return domain.EffectDefinition{
		Trigger:        domain.TriggerTechLevelQuery,
		Effect:         domain.SpecialTechLevelFloor,
		Filters:        domain.EffectFilters{Owner: domain.OwnerOwn, Scope: scope, Categories: categories},
		TechLevelFloor: &domain.TechLevelFloorParams{Level: level},
	}
}

// floorState gives alpha one active floor asset on Tartarus; beta is a rival.
func floorState() (*state.FactionState, *domain.Asset) {
	asset := activeAsset("lab", laboratoryDef.ID)
	asset.Location = tartarus
	return &state.FactionState{Factions: map[string]*domain.Faction{
		"alpha": {ID: "alpha", Assets: map[string]*domain.Asset{asset.ID: asset}},
		"beta":  {ID: "beta", Assets: map[string]*domain.Asset{}},
	}}, asset
}

func registerFloor(factionState *state.FactionState, entry domain.EffectDefinition) *hooks.Registry {
	registry := hooks.NewRegistry()
	alpha := factionState.Factions["alpha"]
	effects.Registrars()[domain.SpecialTechLevelFloor].Register(effect.RegisterContext{
		Faction: alpha, Asset: alpha.Assets["lab"], Def: laboratoryDef, Entry: entry, Registry: registry,
	})
	return registry
}

func alphaTechLevel(factionState *state.FactionState, registry *hooks.Registry, def *domain.AssetDefinition, world string, baseTL int) int {
	return dispatch.ResolveWorldTechLevel(registry, factionState.Factions["alpha"], def, world, baseTL)
}

func TestTechLevelFloor_WorldScopeRaisesOwnWorld(t *testing.T) {
	factionState, _ := floorState()
	registry := registerFloor(factionState, techLevelFloorEntry(domain.ScopeWorld, 4))
	if got := alphaTechLevel(factionState, registry, forceAsset, "Tartarus", 2); got != 4 {
		t.Errorf("tech level = %d, want 4", got)
	}
}

func TestTechLevelFloor_WorldScopeIgnoresOtherWorld(t *testing.T) {
	factionState, _ := floorState()
	registry := registerFloor(factionState, techLevelFloorEntry(domain.ScopeWorld, 4))
	if got := alphaTechLevel(factionState, registry, forceAsset, "Krylos", 2); got != 2 {
		t.Errorf("tech level = %d, want 2 (scope = world excludes other worlds)", got)
	}
}

func TestTechLevelFloor_FactionScopeRaisesEveryWorld(t *testing.T) {
	factionState, _ := floorState()
	registry := registerFloor(factionState, techLevelFloorEntry(domain.ScopeFaction, 4))
	if got := alphaTechLevel(factionState, registry, forceAsset, "Krylos", 2); got != 4 {
		t.Errorf("tech level = %d, want 4 (scope = faction reaches any world)", got)
	}
}

func TestTechLevelFloor_CategoryFilter(t *testing.T) {
	factionState, _ := floorState()
	registry := registerFloor(factionState, techLevelFloorEntry(domain.ScopeWorld, 5, domain.StatCunning, domain.StatWealth))
	if got := alphaTechLevel(factionState, registry, wealthAsset, "Tartarus", 2); got != 5 {
		t.Errorf("listed category: tech level = %d, want 5", got)
	}
	if got := alphaTechLevel(factionState, registry, forceAsset, "Tartarus", 2); got != 2 {
		t.Errorf("unlisted category: tech level = %d, want 2", got)
	}
}

func TestTechLevelFloor_RivalQueryUnchanged(t *testing.T) {
	factionState, _ := floorState()
	registry := registerFloor(factionState, techLevelFloorEntry(domain.ScopeWorld, 4))
	beta := factionState.Factions["beta"]
	for _, registered := range registry.WorldTechLevelModifiersForAsset("alpha", "lab") {
		if got := registered.Hook.ModifyWorldTechLevel(beta, forceAsset, "Tartarus", 2); got != 2 {
			t.Errorf("tech level = %d, want 2 (owner = own excludes a rival's query)", got)
		}
	}
}

func TestTechLevelFloor_InactiveAssetUnchanged(t *testing.T) {
	for name, deactivate := range inactiveStates {
		t.Run(name, func(t *testing.T) {
			factionState, asset := floorState()
			registry := registerFloor(factionState, techLevelFloorEntry(domain.ScopeWorld, 4))
			deactivate(asset)
			if got := alphaTechLevel(factionState, registry, forceAsset, "Tartarus", 2); got != 2 {
				t.Errorf("tech level = %d, want 2 (inactive asset)", got)
			}
		})
	}
}

func TestTechLevelFloor_RemovedAssetUnchanged(t *testing.T) {
	factionState, _ := floorState()
	registry := registerFloor(factionState, techLevelFloorEntry(domain.ScopeFaction, 4))
	alpha := factionState.Factions["alpha"]
	delete(alpha.Assets, "lab")
	for _, registered := range registry.WorldTechLevelModifiersForAsset("alpha", "lab") {
		if got := registered.Hook.ModifyWorldTechLevel(alpha, forceAsset, "Tartarus", 2); got != 2 {
			t.Errorf("tech level = %d, want 2 (asset removed after registration)", got)
		}
	}
}

func TestTechLevelFloor_NeverLowers(t *testing.T) {
	factionState, _ := floorState()
	registry := registerFloor(factionState, techLevelFloorEntry(domain.ScopeWorld, 4))
	if got := alphaTechLevel(factionState, registry, forceAsset, "Tartarus", 5); got != 5 {
		t.Errorf("tech level = %d, want 5 (a floor never lowers)", got)
	}
}

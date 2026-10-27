package effects_test

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

var localInvestmentsDef = &domain.AssetDefinition{ID: "SWN-W1-003", Name: "Local Investments"}

func costSurchargeEntry(noStack bool) domain.EffectDefinition {
	return domain.EffectDefinition{
		Trigger:       domain.TriggerAssetCostQuery,
		Effect:        domain.SpecialCostSurcharge,
		Filters:       domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeWorld},
		Gates:         domain.EffectGates{NoStack: noStack},
		CostSurcharge: &domain.CostSurchargeParams{Amount: 1},
	}
}

// localInvestments builds an active alpha Local Investments on Tartarus.
func localInvestments(id string) *domain.Asset {
	asset := activeAsset(id, localInvestmentsDef.ID)
	asset.Location = tartarus
	return asset
}

// surchargeState gives alpha the listed assets; beta is the buyer with none.
func surchargeState(assets ...*domain.Asset) *state.FactionState {
	alpha := &domain.Faction{ID: "alpha", Assets: map[string]*domain.Asset{}}
	for _, asset := range assets {
		alpha.Assets[asset.ID] = asset
	}
	return &state.FactionState{Factions: map[string]*domain.Faction{
		"alpha": alpha,
		"beta":  {ID: "beta", Assets: map[string]*domain.Asset{}},
	}}
}

// registerSurcharges registers entry for every alpha asset, in ID order.
func registerSurcharges(factionState *state.FactionState, entry domain.EffectDefinition) *hooks.Registry {
	registry := hooks.NewRegistry()
	alpha := factionState.Factions["alpha"]
	for _, asset := range domain.SortedAssets(alpha) {
		effects.Registrars()[domain.SpecialCostSurcharge].Register(effect.RegisterContext{
			Holder: effect.Holder{Faction: alpha, Asset: asset, Def: localInvestmentsDef}, Entry: entry, Registry: registry,
		})
	}
	return registry
}

// askHook runs alpha's asset hooks directly, bypassing the resolver's own-asset
// skip, so a test proves the registrar honours a filter on its own.
func askHook(registry *hooks.Registry, assetID string, buyer *domain.Faction, world string, cost int) int {
	query := &hooks.AssetCostQuery{Cost: cost, Applied: map[string]bool{}}
	for _, registered := range registry.AssetCostModifiersForAsset("alpha", assetID) {
		registered.Hook.ModifyAssetCost(buyer, &domain.AssetDefinition{ID: "bought"}, world, query)
	}
	return query.Cost
}

func onWorld(factionState *state.FactionState) []*domain.Asset {
	return domain.SortedAssets(factionState.Factions["alpha"])
}

func TestCostSurcharge_RivalBuyerOnWorldPaysMore(t *testing.T) {
	factionState := surchargeState(localInvestments("li1"))
	registry := registerSurcharges(factionState, costSurchargeEntry(true))
	got := dispatch.ResolveAssetCost(registry, factionState.Factions["beta"], &domain.AssetDefinition{ID: "bought"}, "Tartarus", 5, onWorld(factionState)).Cost
	if got != 6 {
		t.Errorf("cost = %d, want 6", got)
	}
}

func TestCostSurcharge_RecordsSurchargeOnQuery(t *testing.T) {
	factionState := surchargeState(localInvestments("li1"))
	registry := registerSurcharges(factionState, costSurchargeEntry(true))
	quote := dispatch.ResolveAssetCost(registry, factionState.Factions["beta"], &domain.AssetDefinition{ID: "bought"}, "Tartarus", 5, onWorld(factionState))
	want := []hooks.Surcharge{{FactionID: "alpha", Amount: 1}}
	if !reflect.DeepEqual(quote.Surcharges, want) {
		t.Errorf("surcharges = %+v, want %+v", quote.Surcharges, want)
	}
}

func TestCostSurcharge_OwnerBuyingPaysNothingExtra(t *testing.T) {
	factionState := surchargeState(localInvestments("li1"))
	registry := registerSurcharges(factionState, costSurchargeEntry(true))
	if got := askHook(registry, "li1", factionState.Factions["alpha"], "Tartarus", 5); got != 5 {
		t.Errorf("cost = %d, want 5 (owner = rival excludes the owner)", got)
	}
}

func TestCostSurcharge_OtherWorldPaysNothingExtra(t *testing.T) {
	factionState := surchargeState(localInvestments("li1"))
	registry := registerSurcharges(factionState, costSurchargeEntry(true))
	if got := askHook(registry, "li1", factionState.Factions["beta"], "Krylos", 5); got != 5 {
		t.Errorf("cost = %d, want 5 (scope = world excludes other worlds)", got)
	}
}

func TestCostSurcharge_InactiveAssetPaysNothingExtra(t *testing.T) {
	for name, deactivate := range inactiveStates {
		t.Run(name, func(t *testing.T) {
			asset := localInvestments("li1")
			factionState := surchargeState(asset)
			registry := registerSurcharges(factionState, costSurchargeEntry(true))
			deactivate(asset)
			if got := askHook(registry, "li1", factionState.Factions["beta"], "Tartarus", 5); got != 5 {
				t.Errorf("cost = %d, want 5 (inactive asset)", got)
			}
		})
	}
}

func TestCostSurcharge_RemovedAssetPaysNothingExtra(t *testing.T) {
	factionState := surchargeState(localInvestments("li1"))
	registry := registerSurcharges(factionState, costSurchargeEntry(true))
	delete(factionState.Factions["alpha"].Assets, "li1")
	if got := askHook(registry, "li1", factionState.Factions["beta"], "Tartarus", 5); got != 5 {
		t.Errorf("cost = %d, want 5 (asset removed after registration)", got)
	}
}

func TestCostSurcharge_NoStackCapsTwoCopiesAtOne(t *testing.T) {
	factionState := surchargeState(localInvestments("li1"), localInvestments("li2"))
	registry := registerSurcharges(factionState, costSurchargeEntry(true))
	got := dispatch.ResolveAssetCost(registry, factionState.Factions["beta"], &domain.AssetDefinition{ID: "bought"}, "Tartarus", 5, onWorld(factionState)).Cost
	if got != 6 {
		t.Errorf("cost = %d, want 6 (no_stack applies one surcharge)", got)
	}
}

func TestCostSurcharge_WithoutNoStackTwoCopiesStack(t *testing.T) {
	factionState := surchargeState(localInvestments("li1"), localInvestments("li2"))
	registry := registerSurcharges(factionState, costSurchargeEntry(false))
	got := dispatch.ResolveAssetCost(registry, factionState.Factions["beta"], &domain.AssetDefinition{ID: "bought"}, "Tartarus", 5, onWorld(factionState)).Cost
	if got != 7 {
		t.Errorf("cost = %d, want 7 (each copy applies without no_stack)", got)
	}
}

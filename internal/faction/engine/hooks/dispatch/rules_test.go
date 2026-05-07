package dispatch

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// stubAssetCostModifier reduces cost by a fixed delta.
type stubAssetCostModifier struct{ delta int }

func (stub *stubAssetCostModifier) ModifyAssetCost(_ *domain.Faction, _ *domain.AssetDefinition, _ string, cost int) int {
	return cost - stub.delta
}

// stubMaintenanceCostModifier reduces cost by a fixed delta.
type stubMaintenanceCostModifier struct{ delta int }

func (stub *stubMaintenanceCostModifier) ModifyMaintenanceCost(_ *domain.Faction, _ *domain.Asset, cost int) int {
	return cost - stub.delta
}

// stubTieResolver always returns a fixed outcome.
type stubTieResolver struct{ outcome hooks.TieOutcome }

func (stub *stubTieResolver) ResolveTie(_ hooks.RollContext, _ *state.FactionState) hooks.TieOutcome {
	return stub.outcome
}

func TestResolveAssetCost_NoModifiers(t *testing.T) {
	registry := hooks.NewRegistry()
	buyer := &domain.Faction{ID: "f1"}
	def := &domain.AssetDefinition{ID: "cheap", Cost: 5}
	got := ResolveAssetCost(registry, buyer, def, "Tartarus", 5)
	if got != 5 {
		t.Errorf("ResolveAssetCost = %d, want 5 (no modifiers)", got)
	}
}

func TestResolveAssetCost_OneModifier(t *testing.T) {
	registry := hooks.NewRegistry()
	buyer := &domain.Faction{ID: "f1"}
	def := &domain.AssetDefinition{ID: "cheap", Cost: 5}
	registry.RegisterAssetCostModifier(hooks.FactionScope("f1"), "test", &stubAssetCostModifier{delta: 1})
	got := ResolveAssetCost(registry, buyer, def, "Tartarus", 5)
	if got != 4 {
		t.Errorf("ResolveAssetCost = %d, want 4 (base 5 - 1)", got)
	}
}

func TestResolveAssetCost_TwoModifiersChain(t *testing.T) {
	registry := hooks.NewRegistry()
	buyer := &domain.Faction{ID: "f1"}
	def := &domain.AssetDefinition{ID: "cheap", Cost: 10}
	registry.RegisterAssetCostModifier(hooks.FactionScope("f1"), "first", &stubAssetCostModifier{delta: 2})
	registry.RegisterAssetCostModifier(hooks.FactionScope("f1"), "second", &stubAssetCostModifier{delta: 1})
	got := ResolveAssetCost(registry, buyer, def, "Tartarus", 10)
	// first reduces 10→8, second sees 8 and reduces 8→7
	if got != 7 {
		t.Errorf("ResolveAssetCost = %d, want 7 (10-2-1 chained)", got)
	}
}

func TestResolveMaintenanceCost_NoModifiers(t *testing.T) {
	registry := hooks.NewRegistry()
	owner := &domain.Faction{ID: "f1"}
	asset := &domain.Asset{ID: "a1", OwnerID: "f1"}
	got := ResolveMaintenanceCost(registry, owner, asset, 0)
	if got != 0 {
		t.Errorf("ResolveMaintenanceCost = %d, want 0", got)
	}
}

func TestResolveMaintenanceCost_OneModifier(t *testing.T) {
	registry := hooks.NewRegistry()
	owner := &domain.Faction{ID: "f1"}
	asset := &domain.Asset{ID: "a1", OwnerID: "f1"}
	registry.RegisterMaintenanceCostModifier(hooks.FactionScope("f1"), "test", &stubMaintenanceCostModifier{delta: -2})
	got := ResolveMaintenanceCost(registry, owner, asset, 0)
	// delta=-2 means cost increases by 2
	if got != 2 {
		t.Errorf("ResolveMaintenanceCost = %d, want 2", got)
	}
}

func TestResolveTie_NoResolver(t *testing.T) {
	registry := hooks.NewRegistry()
	ctx := hooks.RollContext{Actor: &domain.Faction{ID: "f1"}}
	got := ResolveTie(registry, ctx, nil)
	if got != hooks.TieStandard {
		t.Errorf("ResolveTie = %v, want TieStandard when no resolver registered", got)
	}
}

func TestResolveTie_DefenderWins(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterTieResolver(hooks.GlobalScope(), "fanatical", &stubTieResolver{outcome: hooks.TieDefenderWins})
	ctx := hooks.RollContext{Actor: &domain.Faction{ID: "f1"}}
	got := ResolveTie(registry, ctx, nil)
	if got != hooks.TieDefenderWins {
		t.Errorf("ResolveTie = %v, want TieDefenderWins", got)
	}
}

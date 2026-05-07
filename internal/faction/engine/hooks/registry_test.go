package hooks_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// --- stub implementations ---

type stubRollModifier struct{ name string }

func (stub *stubRollModifier) OfferModifiers(_ hooks.RollContext, _ *state.FactionState, _ *rulebook.Rulebook) []hooks.ModifierOffer {
	return nil
}

type stubRollResultHook struct{ name string }

func (stub *stubRollResultHook) OnRollResult(_ hooks.RollContext, _ hooks.RollResult, _ *state.FactionState, _ *rulebook.Rulebook) hooks.RerollDirective {
	return hooks.RerollDirective{}
}

type stubMutationReactor struct{ name string }

func (stub *stubMutationReactor) OnMutations(_ []domain.Mutation, _ *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	return nil
}

type stubAssetCostModifier struct{ name string }

func (stub *stubAssetCostModifier) ModifyAssetCost(_ *domain.Faction, _ *domain.AssetDefinition, _ string, baseCost int) int {
	return baseCost
}

type stubTieResolver struct{ name string }

func (stub *stubTieResolver) ResolveTie(_ hooks.RollContext, _ *state.FactionState) hooks.TieOutcome {
	return hooks.TieStandard
}

// --- Cat 1: RollModifier ---

func TestRollModifier_RegistrationOrderPreserved(t *testing.T) {
	registry := hooks.NewRegistry()
	hookA := &stubRollModifier{name: "A"}
	hookB := &stubRollModifier{name: "B"}
	registry.RegisterRollModifier(hooks.FactionScope("f1"), "src-A", hookA)
	registry.RegisterRollModifier(hooks.FactionScope("f1"), "src-B", hookB)

	result := registry.RollModifiersFor("f1", "")
	if len(result) != 2 {
		t.Fatalf("expected 2 hooks, got %d", len(result))
	}
	if result[0].Source != "src-A" || result[1].Source != "src-B" {
		t.Errorf("registration order not preserved: got %v, %v", result[0].Source, result[1].Source)
	}
}

func TestRollModifier_FactionScopeFilters(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.FactionScope("f1"), "src", &stubRollModifier{})

	if got := registry.RollModifiersFor("f2", ""); len(got) != 0 {
		t.Errorf("faction A hook must not appear for faction B, got %d entries", len(got))
	}
}

func TestRollModifier_AssetScopeFilters(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.AssetScope("f1", "asset-X"), "src", &stubRollModifier{})

	if got := registry.RollModifiersFor("f1", "asset-Y"); len(got) != 0 {
		t.Errorf("asset X hook must not appear for asset Y, got %d entries", len(got))
	}
}

func TestRollModifier_GlobalScopeMatchesAnyFaction(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.GlobalScope(), "global", &stubRollModifier{})

	for _, factionID := range []string{"f1", "f2", "f3"} {
		if got := registry.RollModifiersFor(factionID, ""); len(got) != 1 {
			t.Errorf("global hook missing for faction %q", factionID)
		}
	}
}

func TestRollModifier_EmptyRegistryReturnsEmptySlice(t *testing.T) {
	registry := hooks.NewRegistry()
	got := registry.RollModifiersFor("f1", "asset-X")
	if got == nil {
		// nil is acceptable, but verifying no panic
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(got))
	}
}

// --- Cat 2: RollResultHook ---

func TestRollResultHook_RegistrationOrderPreserved(t *testing.T) {
	registry := hooks.NewRegistry()
	hookA := &stubRollResultHook{name: "A"}
	hookB := &stubRollResultHook{name: "B"}
	registry.RegisterRollResultHook(hooks.FactionScope("f1"), "src-A", hookA)
	registry.RegisterRollResultHook(hooks.FactionScope("f1"), "src-B", hookB)

	result := registry.RollResultHooksFor("f1", "")
	if len(result) != 2 || result[0].Source != "src-A" || result[1].Source != "src-B" {
		t.Errorf("registration order not preserved: %v", result)
	}
}

func TestRollResultHook_FactionScopeFilters(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterRollResultHook(hooks.FactionScope("f1"), "src", &stubRollResultHook{})

	if got := registry.RollResultHooksFor("f2", ""); len(got) != 0 {
		t.Errorf("faction A hook must not appear for faction B")
	}
}

func TestRollResultHook_GlobalScopeMatchesAnyFaction(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterRollResultHook(hooks.GlobalScope(), "global", &stubRollResultHook{})

	for _, factionID := range []string{"f1", "f2"} {
		if got := registry.RollResultHooksFor(factionID, ""); len(got) != 1 {
			t.Errorf("global hook missing for faction %q", factionID)
		}
	}
}

func TestRollResultHook_EmptyRegistryReturnsEmptySlice(t *testing.T) {
	registry := hooks.NewRegistry()
	got := registry.RollResultHooksFor("f1", "")
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(got))
	}
}

// --- Cat 3: MutationReactor ---

func TestMutationReactor_RegistrationOrderPreserved(t *testing.T) {
	registry := hooks.NewRegistry()
	hookA := &stubMutationReactor{name: "A"}
	hookB := &stubMutationReactor{name: "B"}
	registry.RegisterMutationReactor(hooks.FactionScope("f1"), "src-A", hookA)
	registry.RegisterMutationReactor(hooks.FactionScope("f1"), "src-B", hookB)

	result := registry.MutationReactorsFor("f1", "")
	if len(result) != 2 || result[0].Source != "src-A" || result[1].Source != "src-B" {
		t.Errorf("registration order not preserved: %v", result)
	}
}

func TestMutationReactor_FactionScopeFilters(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterMutationReactor(hooks.FactionScope("f1"), "src", &stubMutationReactor{})

	if got := registry.MutationReactorsFor("f2", ""); len(got) != 0 {
		t.Errorf("faction A hook must not appear for faction B")
	}
}

func TestMutationReactor_GlobalScopeMatchesAnyFaction(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterMutationReactor(hooks.GlobalScope(), "global", &stubMutationReactor{})

	for _, factionID := range []string{"f1", "f2"} {
		if got := registry.MutationReactorsFor(factionID, ""); len(got) != 1 {
			t.Errorf("global hook missing for faction %q", factionID)
		}
	}
}

func TestMutationReactor_EmptyRegistryReturnsEmptySlice(t *testing.T) {
	registry := hooks.NewRegistry()
	got := registry.MutationReactorsFor("f1", "")
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(got))
	}
}

// --- Cat 4: AssetCostModifier ---

func TestAssetCostModifier_RegistrationOrderPreserved(t *testing.T) {
	registry := hooks.NewRegistry()
	hookA := &stubAssetCostModifier{name: "A"}
	hookB := &stubAssetCostModifier{name: "B"}
	registry.RegisterAssetCostModifier(hooks.FactionScope("f1"), "src-A", hookA)
	registry.RegisterAssetCostModifier(hooks.FactionScope("f1"), "src-B", hookB)

	result := registry.AssetCostModifiersFor("f1", "")
	if len(result) != 2 || result[0].Source != "src-A" || result[1].Source != "src-B" {
		t.Errorf("registration order not preserved: %v", result)
	}
}

func TestAssetCostModifier_AssetScopeFilters(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterAssetCostModifier(hooks.AssetScope("f1", "asset-X"), "src", &stubAssetCostModifier{})

	if got := registry.AssetCostModifiersFor("f1", "asset-Y"); len(got) != 0 {
		t.Errorf("asset X hook must not appear for asset Y")
	}
	if got := registry.AssetCostModifiersFor("f1", "asset-X"); len(got) != 1 {
		t.Errorf("asset X hook must appear for asset X, got %d", len(got))
	}
}

func TestAssetCostModifier_GlobalScopeMatchesAnyFaction(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterAssetCostModifier(hooks.GlobalScope(), "global", &stubAssetCostModifier{})

	for _, factionID := range []string{"f1", "f2"} {
		if got := registry.AssetCostModifiersFor(factionID, ""); len(got) != 1 {
			t.Errorf("global hook missing for faction %q", factionID)
		}
	}
}

func TestAssetCostModifier_EmptyRegistryReturnsEmptySlice(t *testing.T) {
	registry := hooks.NewRegistry()
	got := registry.AssetCostModifiersFor("f1", "")
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(got))
	}
}

// --- Cat 5: TieResolver ---

func TestTieResolver_RegistrationOrderPreserved(t *testing.T) {
	registry := hooks.NewRegistry()
	hookA := &stubTieResolver{name: "A"}
	hookB := &stubTieResolver{name: "B"}
	registry.RegisterTieResolver(hooks.FactionScope("f1"), "src-A", hookA)
	registry.RegisterTieResolver(hooks.FactionScope("f1"), "src-B", hookB)

	result := registry.TieResolversFor("f1", "")
	if len(result) != 2 || result[0].Source != "src-A" || result[1].Source != "src-B" {
		t.Errorf("registration order not preserved: %v", result)
	}
}

func TestTieResolver_FactionScopeFilters(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterTieResolver(hooks.FactionScope("f1"), "src", &stubTieResolver{})

	if got := registry.TieResolversFor("f2", ""); len(got) != 0 {
		t.Errorf("faction A hook must not appear for faction B")
	}
}

func TestTieResolver_GlobalScopeMatchesAnyFaction(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterTieResolver(hooks.GlobalScope(), "global", &stubTieResolver{})

	for _, factionID := range []string{"f1", "f2"} {
		if got := registry.TieResolversFor(factionID, ""); len(got) != 1 {
			t.Errorf("global hook missing for faction %q", factionID)
		}
	}
}

func TestTieResolver_EmptyRegistryReturnsEmptySlice(t *testing.T) {
	registry := hooks.NewRegistry()
	got := registry.TieResolversFor("f1", "")
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(got))
	}
}

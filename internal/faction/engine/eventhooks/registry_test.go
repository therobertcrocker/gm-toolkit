package eventhooks_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/eventhooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// --- stub implementations ---

type stubRollModifier struct{ name string }

func (stub *stubRollModifier) OfferModifiers(_ eventhooks.RollContext, _ *state.FactionState, _ *rulebook.Rulebook) []eventhooks.ModifierOffer {
	return nil
}

type stubRollResultHook struct{ name string }

func (stub *stubRollResultHook) OnRollResult(_ eventhooks.RollContext, _ eventhooks.RollResult, _ *state.FactionState, _ *rulebook.Rulebook) eventhooks.RerollDirective {
	return eventhooks.RerollDirective{}
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

func (stub *stubTieResolver) ResolveTie(_ eventhooks.RollContext, _ *state.FactionState) eventhooks.TieOutcome {
	return eventhooks.TieStandard
}

// --- Cat 1: RollModifier ---

func TestRollModifier_RegistrationOrderPreserved(t *testing.T) {
	registry := eventhooks.NewRegistry()
	hookA := &stubRollModifier{name: "A"}
	hookB := &stubRollModifier{name: "B"}
	registry.RegisterRollModifier(eventhooks.FactionScope("f1"), "src-A", hookA)
	registry.RegisterRollModifier(eventhooks.FactionScope("f1"), "src-B", hookB)

	result := registry.RollModifiersFor("f1", "")
	if len(result) != 2 {
		t.Fatalf("expected 2 hooks, got %d", len(result))
	}
	if result[0].Source != "src-A" || result[1].Source != "src-B" {
		t.Errorf("registration order not preserved: got %v, %v", result[0].Source, result[1].Source)
	}
}

func TestRollModifier_FactionScopeFilters(t *testing.T) {
	registry := eventhooks.NewRegistry()
	registry.RegisterRollModifier(eventhooks.FactionScope("f1"), "src", &stubRollModifier{})

	if got := registry.RollModifiersFor("f2", ""); len(got) != 0 {
		t.Errorf("faction A hook must not appear for faction B, got %d entries", len(got))
	}
}

func TestRollModifier_AssetScopeFilters(t *testing.T) {
	registry := eventhooks.NewRegistry()
	registry.RegisterRollModifier(eventhooks.AssetScope("f1", "asset-X"), "src", &stubRollModifier{})

	if got := registry.RollModifiersFor("f1", "asset-Y"); len(got) != 0 {
		t.Errorf("asset X hook must not appear for asset Y, got %d entries", len(got))
	}
}

func TestRollModifier_GlobalScopeMatchesAnyFaction(t *testing.T) {
	registry := eventhooks.NewRegistry()
	registry.RegisterRollModifier(eventhooks.GlobalScope(), "global", &stubRollModifier{})

	for _, factionID := range []string{"f1", "f2", "f3"} {
		if got := registry.RollModifiersFor(factionID, ""); len(got) != 1 {
			t.Errorf("global hook missing for faction %q", factionID)
		}
	}
}

func TestRollModifier_EmptyRegistryReturnsEmptySlice(t *testing.T) {
	registry := eventhooks.NewRegistry()
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
	registry := eventhooks.NewRegistry()
	hookA := &stubRollResultHook{name: "A"}
	hookB := &stubRollResultHook{name: "B"}
	registry.RegisterRollResultHook(eventhooks.FactionScope("f1"), "src-A", hookA)
	registry.RegisterRollResultHook(eventhooks.FactionScope("f1"), "src-B", hookB)

	result := registry.RollResultHooksFor("f1", "")
	if len(result) != 2 || result[0].Source != "src-A" || result[1].Source != "src-B" {
		t.Errorf("registration order not preserved: %v", result)
	}
}

func TestRollResultHook_FactionScopeFilters(t *testing.T) {
	registry := eventhooks.NewRegistry()
	registry.RegisterRollResultHook(eventhooks.FactionScope("f1"), "src", &stubRollResultHook{})

	if got := registry.RollResultHooksFor("f2", ""); len(got) != 0 {
		t.Errorf("faction A hook must not appear for faction B")
	}
}

func TestRollResultHook_GlobalScopeMatchesAnyFaction(t *testing.T) {
	registry := eventhooks.NewRegistry()
	registry.RegisterRollResultHook(eventhooks.GlobalScope(), "global", &stubRollResultHook{})

	for _, factionID := range []string{"f1", "f2"} {
		if got := registry.RollResultHooksFor(factionID, ""); len(got) != 1 {
			t.Errorf("global hook missing for faction %q", factionID)
		}
	}
}

func TestRollResultHook_EmptyRegistryReturnsEmptySlice(t *testing.T) {
	registry := eventhooks.NewRegistry()
	got := registry.RollResultHooksFor("f1", "")
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(got))
	}
}

// --- Cat 3: MutationReactor ---

func TestMutationReactor_RegistrationOrderPreserved(t *testing.T) {
	registry := eventhooks.NewRegistry()
	hookA := &stubMutationReactor{name: "A"}
	hookB := &stubMutationReactor{name: "B"}
	registry.RegisterMutationReactor(eventhooks.FactionScope("f1"), "src-A", hookA)
	registry.RegisterMutationReactor(eventhooks.FactionScope("f1"), "src-B", hookB)

	result := registry.MutationReactorsFor("f1", "")
	if len(result) != 2 || result[0].Source != "src-A" || result[1].Source != "src-B" {
		t.Errorf("registration order not preserved: %v", result)
	}
}

func TestMutationReactor_FactionScopeFilters(t *testing.T) {
	registry := eventhooks.NewRegistry()
	registry.RegisterMutationReactor(eventhooks.FactionScope("f1"), "src", &stubMutationReactor{})

	if got := registry.MutationReactorsFor("f2", ""); len(got) != 0 {
		t.Errorf("faction A hook must not appear for faction B")
	}
}

func TestMutationReactor_GlobalScopeMatchesAnyFaction(t *testing.T) {
	registry := eventhooks.NewRegistry()
	registry.RegisterMutationReactor(eventhooks.GlobalScope(), "global", &stubMutationReactor{})

	for _, factionID := range []string{"f1", "f2"} {
		if got := registry.MutationReactorsFor(factionID, ""); len(got) != 1 {
			t.Errorf("global hook missing for faction %q", factionID)
		}
	}
}

func TestMutationReactor_EmptyRegistryReturnsEmptySlice(t *testing.T) {
	registry := eventhooks.NewRegistry()
	got := registry.MutationReactorsFor("f1", "")
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(got))
	}
}

// --- Cat 4: AssetCostModifier ---

func TestAssetCostModifier_RegistrationOrderPreserved(t *testing.T) {
	registry := eventhooks.NewRegistry()
	hookA := &stubAssetCostModifier{name: "A"}
	hookB := &stubAssetCostModifier{name: "B"}
	registry.RegisterAssetCostModifier(eventhooks.FactionScope("f1"), "src-A", hookA)
	registry.RegisterAssetCostModifier(eventhooks.FactionScope("f1"), "src-B", hookB)

	result := registry.AssetCostModifiersFor("f1", "")
	if len(result) != 2 || result[0].Source != "src-A" || result[1].Source != "src-B" {
		t.Errorf("registration order not preserved: %v", result)
	}
}

func TestAssetCostModifier_AssetScopeFilters(t *testing.T) {
	registry := eventhooks.NewRegistry()
	registry.RegisterAssetCostModifier(eventhooks.AssetScope("f1", "asset-X"), "src", &stubAssetCostModifier{})

	if got := registry.AssetCostModifiersFor("f1", "asset-Y"); len(got) != 0 {
		t.Errorf("asset X hook must not appear for asset Y")
	}
	if got := registry.AssetCostModifiersFor("f1", "asset-X"); len(got) != 1 {
		t.Errorf("asset X hook must appear for asset X, got %d", len(got))
	}
}

func TestAssetCostModifier_GlobalScopeMatchesAnyFaction(t *testing.T) {
	registry := eventhooks.NewRegistry()
	registry.RegisterAssetCostModifier(eventhooks.GlobalScope(), "global", &stubAssetCostModifier{})

	for _, factionID := range []string{"f1", "f2"} {
		if got := registry.AssetCostModifiersFor(factionID, ""); len(got) != 1 {
			t.Errorf("global hook missing for faction %q", factionID)
		}
	}
}

func TestAssetCostModifier_EmptyRegistryReturnsEmptySlice(t *testing.T) {
	registry := eventhooks.NewRegistry()
	got := registry.AssetCostModifiersFor("f1", "")
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(got))
	}
}

// --- Cat 5: TieResolver ---

func TestTieResolver_RegistrationOrderPreserved(t *testing.T) {
	registry := eventhooks.NewRegistry()
	hookA := &stubTieResolver{name: "A"}
	hookB := &stubTieResolver{name: "B"}
	registry.RegisterTieResolver(eventhooks.FactionScope("f1"), "src-A", hookA)
	registry.RegisterTieResolver(eventhooks.FactionScope("f1"), "src-B", hookB)

	result := registry.TieResolversFor("f1", "")
	if len(result) != 2 || result[0].Source != "src-A" || result[1].Source != "src-B" {
		t.Errorf("registration order not preserved: %v", result)
	}
}

func TestTieResolver_FactionScopeFilters(t *testing.T) {
	registry := eventhooks.NewRegistry()
	registry.RegisterTieResolver(eventhooks.FactionScope("f1"), "src", &stubTieResolver{})

	if got := registry.TieResolversFor("f2", ""); len(got) != 0 {
		t.Errorf("faction A hook must not appear for faction B")
	}
}

func TestTieResolver_GlobalScopeMatchesAnyFaction(t *testing.T) {
	registry := eventhooks.NewRegistry()
	registry.RegisterTieResolver(eventhooks.GlobalScope(), "global", &stubTieResolver{})

	for _, factionID := range []string{"f1", "f2"} {
		if got := registry.TieResolversFor(factionID, ""); len(got) != 1 {
			t.Errorf("global hook missing for faction %q", factionID)
		}
	}
}

func TestTieResolver_EmptyRegistryReturnsEmptySlice(t *testing.T) {
	registry := eventhooks.NewRegistry()
	got := registry.TieResolversFor("f1", "")
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(got))
	}
}

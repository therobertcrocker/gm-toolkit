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

type stubRollOfferFilter struct{ name string }

func (stub *stubRollOfferFilter) FilterOffers(_ hooks.RollContext, offers []hooks.ModifierOffer, _ *state.FactionState) []hooks.ModifierOffer {
	return offers
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

func (stub *stubAssetCostModifier) ModifyAssetCost(_ *domain.Faction, _ *domain.AssetDefinition, _ string, _ *hooks.AssetCostQuery) {
}

type stubTieResolver struct{ name string }

func (stub *stubTieResolver) ResolveTie(_ hooks.RollContext, _ *state.FactionState) hooks.TieOutcome {
	return hooks.TieStandard
}

type stubAssetMovementGranter struct{ name string }

func (stub *stubAssetMovementGranter) GrantMovementAbilities(_ *domain.Faction) []hooks.MovementAbility {
	return nil
}

type stubRepairCostModifier struct{}

func (stub *stubRepairCostModifier) ModifyRepairCost(_ *domain.Faction, _ *domain.Asset, _ *domain.AssetDefinition, _ *hooks.RepairCostQuery) {
}

type stubLethalInterceptor struct{ name string }

func (stub *stubLethalInterceptor) OfferInterception(_ hooks.LethalContext) *hooks.Interception {
	return nil
}

// --- Cat 1: RollModifier ---

func TestRollModifier_RegistrationOrderPreserved(t *testing.T) {
	registry := hooks.NewRegistry()
	hookA := &stubRollModifier{name: "A"}
	hookB := &stubRollModifier{name: "B"}
	registry.RegisterRollModifier(hooks.FactionScope("f1"), "src-A", hookA)
	registry.RegisterRollModifier(hooks.FactionScope("f1"), "src-B", hookB)

	result := registry.RollModifiersAt(hooks.FactionScope("f1"))
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

	if got := registry.RollModifiersAt(hooks.FactionScope("f2")); len(got) != 0 {
		t.Errorf("faction A hook must not appear for faction B, got %d entries", len(got))
	}
}

func TestRollModifier_AssetScopeFilters(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.AssetScope("f1", "asset-X"), "src", &stubRollModifier{})

	if got := registry.RollModifiersAt(hooks.AssetScope("f1", "asset-Y")); len(got) != 0 {
		t.Errorf("asset X hook must not appear for asset Y, got %d entries", len(got))
	}
}

func TestRollModifier_GlobalScopeIsSeparateFromFactionScope(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.GlobalScope(), "global", &stubRollModifier{})

	if got := registry.RollModifiersAt(hooks.GlobalScope()); len(got) != 1 {
		t.Errorf("global hook missing: got %d entries", len(got))
	}
	if got := registry.RollModifiersAt(hooks.FactionScope("f1")); len(got) != 0 {
		t.Errorf("global hook must not appear in a faction lookup")
	}
}

func TestRollModifier_EmptyRegistryReturnsEmptySlice(t *testing.T) {
	registry := hooks.NewRegistry()
	got := registry.RollModifiersAt(hooks.AssetScope("f1", "asset-X"))
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(got))
	}
}

// --- Cat 1: RollOfferFilter ---

func TestRollOfferFilter_RegistrationOrderPreserved(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterRollOfferFilter(hooks.AssetScope("f1", "asset-X"), "src-A", &stubRollOfferFilter{name: "A"})
	registry.RegisterRollOfferFilter(hooks.AssetScope("f1", "asset-X"), "src-B", &stubRollOfferFilter{name: "B"})

	result := registry.RollOfferFiltersAt(hooks.AssetScope("f1", "asset-X"))
	if len(result) != 2 {
		t.Fatalf("expected 2 hooks, got %d", len(result))
	}
	if result[0].Source != "src-A" || result[1].Source != "src-B" {
		t.Errorf("registration order not preserved: got %v, %v", result[0].Source, result[1].Source)
	}
}

func TestRollOfferFilter_EmptyRegistryReturnsEmptySlice(t *testing.T) {
	registry := hooks.NewRegistry()
	if got := registry.RollOfferFiltersAt(hooks.AssetScope("f1", "asset-X")); len(got) != 0 {
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

	result := registry.RollResultHooksAt(hooks.FactionScope("f1"))
	if len(result) != 2 || result[0].Source != "src-A" || result[1].Source != "src-B" {
		t.Errorf("registration order not preserved: %v", result)
	}
}

func TestRollResultHook_FactionScopeFilters(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterRollResultHook(hooks.FactionScope("f1"), "src", &stubRollResultHook{})

	if got := registry.RollResultHooksAt(hooks.FactionScope("f2")); len(got) != 0 {
		t.Errorf("faction A hook must not appear for faction B")
	}
}

func TestRollResultHook_GlobalScopeIsSeparateFromFactionScope(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterRollResultHook(hooks.GlobalScope(), "global", &stubRollResultHook{})

	if got := registry.RollResultHooksAt(hooks.GlobalScope()); len(got) != 1 {
		t.Errorf("global hook missing: got %d entries", len(got))
	}
	if got := registry.RollResultHooksAt(hooks.FactionScope("f1")); len(got) != 0 {
		t.Errorf("global hook must not appear in a faction lookup")
	}
}

func TestRollResultHook_EmptyRegistryReturnsEmptySlice(t *testing.T) {
	registry := hooks.NewRegistry()
	got := registry.RollResultHooksAt(hooks.FactionScope("f1"))
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

	result := registry.MutationReactorsForFaction("f1")
	if len(result) != 2 || result[0].Source != "src-A" || result[1].Source != "src-B" {
		t.Errorf("registration order not preserved: %v", result)
	}
}

func TestMutationReactor_FactionScopeFilters(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterMutationReactor(hooks.FactionScope("f1"), "src", &stubMutationReactor{})

	if got := registry.MutationReactorsForFaction("f2"); len(got) != 0 {
		t.Errorf("faction A hook must not appear for faction B")
	}
}

func TestMutationReactor_GlobalScopeIsSeparateFromFactionScope(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterMutationReactor(hooks.GlobalScope(), "global", &stubMutationReactor{})

	if got := registry.MutationReactorsGlobal(); len(got) != 1 {
		t.Errorf("global hook missing: got %d entries", len(got))
	}
	if got := registry.MutationReactorsForFaction("f1"); len(got) != 0 {
		t.Errorf("global hook must not appear in a faction lookup")
	}
}

func TestMutationReactor_AssetScopeFilters(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterMutationReactor(hooks.AssetScope("f1", "a1"), "src", &stubMutationReactor{})

	if got := registry.MutationReactorsForAsset("f1", "a1"); len(got) != 1 {
		t.Errorf("asset hook missing for its own asset: got %d entries", len(got))
	}
	if got := registry.MutationReactorsForAsset("f1", "a2"); len(got) != 0 {
		t.Errorf("asset hook must not appear for another asset")
	}
	if got := registry.MutationReactorsForFaction("f1"); len(got) != 0 {
		t.Errorf("asset hook must not appear in a faction lookup")
	}
}

func TestMutationReactor_EmptyRegistryReturnsEmptySlice(t *testing.T) {
	registry := hooks.NewRegistry()
	if got := registry.MutationReactorsGlobal(); len(got) != 0 {
		t.Errorf("global: expected empty slice, got %d entries", len(got))
	}
	if got := registry.MutationReactorsForFaction("f1"); len(got) != 0 {
		t.Errorf("faction: expected empty slice, got %d entries", len(got))
	}
	if got := registry.MutationReactorsForAsset("f1", "a1"); len(got) != 0 {
		t.Errorf("asset: expected empty slice, got %d entries", len(got))
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

// --- Cat 4: AssetMovementGranter ---

func TestAssetMovementGranter_ForAssetReturnsOnlyThatBucket(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterAssetMovementGranter(hooks.GlobalScope(), "global", &stubAssetMovementGranter{})
	registry.RegisterAssetMovementGranter(hooks.FactionScope("f1"), "faction", &stubAssetMovementGranter{})
	registry.RegisterAssetMovementGranter(hooks.AssetScope("f1", "asset-X"), "asset-X", &stubAssetMovementGranter{})
	registry.RegisterAssetMovementGranter(hooks.AssetScope("f1", "asset-Y"), "asset-Y", &stubAssetMovementGranter{})

	got := registry.AssetMovementGrantersForAsset("f1", "asset-X")
	if len(got) != 1 || got[0].Source != "asset-X" {
		t.Errorf("AssetMovementGrantersForAsset = %v, want only the asset-X hook", got)
	}
}

// --- ClearAsset ---

func TestClearAsset_RemovesOnlyThatAssetsBucket(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.AssetScope("f1", "asset-X"), "asset-X", &stubRollModifier{})
	registry.RegisterMutationReactor(hooks.AssetScope("f1", "asset-X"), "asset-X", &stubMutationReactor{})
	registry.RegisterAssetMovementGranter(hooks.AssetScope("f1", "asset-X"), "asset-X", &stubAssetMovementGranter{})
	registry.RegisterLethalInterceptor(hooks.AssetScope("f1", "asset-X"), "asset-X", &stubLethalInterceptor{})
	registry.RegisterRepairCostModifier(hooks.AssetScope("f1", "asset-X"), "asset-X", &stubRepairCostModifier{})
	registry.RegisterMutationReactor(hooks.AssetScope("f1", "asset-Y"), "asset-Y", &stubMutationReactor{})
	registry.RegisterMutationReactor(hooks.AssetScope("f2", "asset-X"), "f2-asset-X", &stubMutationReactor{})
	registry.RegisterMutationReactor(hooks.FactionScope("f1"), "faction", &stubMutationReactor{})

	registry.ClearAsset("f1", "asset-X")

	if got := registry.RollModifiersAt(hooks.AssetScope("f1", "asset-X")); len(got) != 0 {
		t.Errorf("roll modifiers under f1/asset-X = %v, want none", got)
	}
	if got := registry.MutationReactorsForAsset("f1", "asset-X"); len(got) != 0 {
		t.Errorf("reactors under f1/asset-X = %v, want none", got)
	}
	if got := registry.AssetMovementGrantersForAsset("f1", "asset-X"); len(got) != 0 {
		t.Errorf("movement granters under f1/asset-X = %v, want none", got)
	}
	if got := registry.LethalInterceptorsForAsset("f1", "asset-X"); len(got) != 0 {
		t.Errorf("lethal interceptors under f1/asset-X = %v, want none", got)
	}
	if got := registry.RepairCostModifiersForAsset("f1", "asset-X"); len(got) != 0 {
		t.Errorf("repair-cost modifiers under f1/asset-X = %v, want none", got)
	}
	if len(registry.MutationReactorsForAsset("f1", "asset-Y")) != 1 ||
		len(registry.MutationReactorsForAsset("f2", "asset-X")) != 1 ||
		len(registry.MutationReactorsForFaction("f1")) != 1 {
		t.Error("ClearAsset removed a hook outside f1/asset-X")
	}
}

// --- Cat 6: LethalInterceptor ---

func TestLethalInterceptor_ForAssetReturnsOnlyThatBucket(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterLethalInterceptor(hooks.AssetScope("f1", "asset-X"), "x", &stubLethalInterceptor{name: "x"})
	registry.RegisterLethalInterceptor(hooks.AssetScope("f1", "asset-Y"), "y", &stubLethalInterceptor{name: "y"})

	for _, testCase := range []struct{ assetID, want string }{{"asset-X", "x"}, {"asset-Y", "y"}} {
		got := registry.LethalInterceptorsForAsset("f1", testCase.assetID)
		if len(got) != 1 || got[0].Source != testCase.want {
			t.Errorf("LethalInterceptorsForAsset(f1, %s) = %v, want only %q", testCase.assetID, got, testCase.want)
		}
	}
}

package effects_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// registerDiscount registers Medical Center's discount_repair entry on medicalCenter,
// as the SWN data declares it, and returns its modifier.
func registerDiscount(t *testing.T, owner *domain.Faction, medicalCenter *domain.Asset) hooks.RepairCostModifier {
	t.Helper()
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialDiscountRepair].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: owner, Asset: medicalCenter, Def: &domain.AssetDefinition{ID: "SWN-W4-003"}}, Registry: registry,
		Entry: domain.EffectDefinition{
			Trigger:        domain.TriggerRepairCostQuery,
			Effect:         domain.SpecialDiscountRepair,
			Filters:        domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeWorld, AssetTypes: medicalCenterTypes},
			Gates:          domain.EffectGates{NoStack: true},
			DiscountRepair: &domain.DiscountRepairParams{Amount: 1},
		},
	})
	modifiers := registry.RepairCostModifiersForAsset(owner.ID, medicalCenter.ID)
	if len(modifiers) != 1 {
		t.Fatalf("modifiers = %d, want 1", len(modifiers))
	}
	return modifiers[0].Hook
}

func TestDiscountRepair_TakesOneOff(t *testing.T) {
	factionState := medicalCenterState()
	alpha := factionState.Factions["alpha"]
	modifier := registerDiscount(t, alpha, alpha.Assets["mc"])

	query := &hooks.RepairCostQuery{Cost: 3, Applied: map[string]bool{}}
	modifier.ModifyRepairCost(alpha, alpha.Assets["mu"], militaryUnitDef, query)

	if query.Cost != 2 {
		t.Errorf("Cost = %d, want 2", query.Cost)
	}
}

func TestDiscountRepair_NoDiscount(t *testing.T) {
	cases := []struct {
		name  string
		setup func(alpha, beta *domain.Faction) (*domain.Faction, *domain.Asset, *domain.AssetDefinition)
	}{
		{"a rival repairer", func(alpha, beta *domain.Faction) (*domain.Faction, *domain.Asset, *domain.AssetDefinition) {
			return beta, beta.Assets["x"], militaryUnitDef
		}},
		{"an asset on another world", func(alpha, beta *domain.Faction) (*domain.Faction, *domain.Asset, *domain.AssetDefinition) {
			return alpha, placedOn(alpha.Assets["mu"], "elsewhere"), militaryUnitDef
		}},
		{"a Facility", func(alpha, beta *domain.Faction) (*domain.Faction, *domain.Asset, *domain.AssetDefinition) {
			return alpha, alpha.Assets["mu"], facilityDef
		}},
		{"an inactive Medical Center", func(alpha, beta *domain.Faction) (*domain.Faction, *domain.Asset, *domain.AssetDefinition) {
			alpha.Assets["mc"].Maintained = false
			return alpha, alpha.Assets["mu"], militaryUnitDef
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState := medicalCenterState()
			alpha, beta := factionState.Factions["alpha"], factionState.Factions["beta"]
			modifier := registerDiscount(t, alpha, alpha.Assets["mc"])
			repairer, asset, def := tc.setup(alpha, beta)

			query := &hooks.RepairCostQuery{Cost: 3, Applied: map[string]bool{}}
			modifier.ModifyRepairCost(repairer, asset, def, query)

			if query.Cost != 3 {
				t.Errorf("Cost = %d, want 3", query.Cost)
			}
		})
	}
}

func TestDiscountRepair_NoStack(t *testing.T) {
	factionState := medicalCenterState()
	alpha := factionState.Factions["alpha"]
	alpha.Assets["mc2"] = placedOn(activeAsset("mc2", "SWN-W4-003"), "anchorage")
	first := registerDiscount(t, alpha, alpha.Assets["mc"])
	second := registerDiscount(t, alpha, alpha.Assets["mc2"])

	query := &hooks.RepairCostQuery{Cost: 3, Applied: map[string]bool{}}
	first.ModifyRepairCost(alpha, alpha.Assets["mu"], militaryUnitDef, query)
	second.ModifyRepairCost(alpha, alpha.Assets["mu"], militaryUnitDef, query)

	if query.Cost != 2 {
		t.Errorf("Cost = %d, want 2 (a second Medical Center adds nothing)", query.Cost)
	}
}

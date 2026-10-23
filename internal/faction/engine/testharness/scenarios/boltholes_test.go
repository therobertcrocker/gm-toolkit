package scenarios

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

const defBoltholes = "SWN-C5-003"

// takeAction returns a SelectActionFn where factionID takes the named action and
// every other faction skips.
func takeAction(t *testing.T, factionID, name string) func(*domain.Faction, []action.Action) (action.Action, error) {
	return func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID != factionID {
			return nil, nil
		}
		for _, candidate := range available {
			if candidate.Name() == name {
				return candidate, nil
			}
		}
		t.Fatalf("%s not available for %s", name, factionID)
		return nil, nil
	}
}

// setupBoltholesShelter puts beta's Boltholes and its Security Personnel (a
// Military Unit, 3 HP) on Hadrian with alpha's attacker, and runs the cycle in
// which alpha's hit for 4 would destroy the unit.
func setupBoltholesShelter(t *testing.T) (h *testharness.Harness, alpha, beta *domain.Faction, attacker, boltholes, unit *domain.Asset) {
	t.Helper()
	h = testharness.NewHarness(t)
	alpha = h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	beta = h.AddFaction("beta", "Hadrian", 4, 3, 2)
	attacker = addActiveAsset(alpha, "attacker", testharness.DefSecurityPersonnel, "Hadrian")
	boltholes = addActiveAsset(beta, "bolt", defBoltholes, "Hadrian")
	unit = beta.Assets["beta-asset-1"]

	// Rolls: attack 10+4 vs defense 1+4, a hit; damage 1d3+1 rolls 3, so 4 damage.
	h.Engine.Rand = &testharness.FixedRoller{Values: []int{10, 1, 3}}
	h.Collector.SelectActionFn = takeAction(t, alpha.ID, "Attack")
	h.Collector.SelectAttackersFn = func([]*domain.Asset, *rulebook.Rulebook) ([]*domain.Asset, error) {
		return []*domain.Asset{attacker}, nil
	}
	h.Collector.SelectDefenderFn = func(*domain.Asset, []*domain.Asset, *rulebook.Rulebook) (*domain.Asset, error) {
		return unit, nil
	}

	runOrderedCycle(t, h, alpha.ID, beta.ID)

	testharness.CheckStep(t, "the unit stays beta's at 0 HP", beta.Assets[unit.ID] == unit && unit.CurrentHP == 0,
		"the unit was removed or its HP is not 0")
	testharness.CheckStep(t, "the unit is sheltered", unit.HasCondition(domain.ConditionSheltered), "no sheltered condition")
	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	removals := 0
	for _, record := range records {
		for _, mutation := range record.Mutations {
			if mutation.Type == "asset_removed" {
				removals++
			}
		}
	}
	testharness.CheckStep(t, "history has no asset_removed", removals == 0, "an asset was removed")
	return h, alpha, beta, attacker, boltholes, unit
}

// TestBoltholesShelterRepairAndLoss: Boltholes shelters beta's unit from a killing
// blow; beta repairs it to full HP next cycle, and the shelter ends (Effort 2
// Decision 12).
func TestBoltholesShelterRepairAndLoss(t *testing.T) {
	h, _, beta, _, _, unit := setupBoltholesShelter(t)

	beta.Coin = 1
	h.Collector.SelectActionFn = takeAction(t, beta.ID, "Repair Asset")
	h.Collector.SelectRepairOrdersFn = func(*domain.Faction, []*domain.Asset, *rulebook.Rulebook) ([]action.RepairOrder, error) {
		return []action.RepairOrder{{Asset: unit, HealCount: 1}}, nil
	}
	runOrderedCycle(t, h, "alpha", beta.ID)

	testharness.CheckStep(t, "the unit is back at full HP", unit.CurrentHP == 3, "the repair did not heal the unit to 3 HP")
	testharness.CheckStep(t, "the shelter is gone", !unit.HasCondition(domain.ConditionSheltered), "the unit is still sheltered")
}

// TestBoltholesLossTakesShelteredAsset: after the shelter, alpha destroys the
// Boltholes, and the unit it shelters is destroyed with it as a combat kill
// (Effort 2 Decision 13).
func TestBoltholesLossTakesShelteredAsset(t *testing.T) {
	h, alpha, beta, attacker, boltholes, unit := setupBoltholesShelter(t)

	boltholes.CurrentHP = 1
	h.Engine.Rand = &testharness.FixedRoller{Values: []int{10, 1, 3}}
	h.Collector.SelectActionFn = takeAction(t, alpha.ID, "Attack")
	h.Collector.SelectAttackersFn = func([]*domain.Asset, *rulebook.Rulebook) ([]*domain.Asset, error) {
		return []*domain.Asset{attacker}, nil
	}
	h.Collector.SelectDefenderFn = func(*domain.Asset, []*domain.Asset, *rulebook.Rulebook) (*domain.Asset, error) {
		return boltholes, nil
	}
	runOrderedCycle(t, h, alpha.ID, beta.ID)

	testharness.CheckStep(t, "the Boltholes is destroyed", beta.Assets[boltholes.ID] == nil, "the Boltholes is still in play")
	testharness.CheckStep(t, "the sheltered unit is destroyed with it", beta.Assets[unit.ID] == nil, "the unit is still in play")
	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	testharness.CheckStep(t, "both removals are combat kills", countByTypeAndCause(records, "asset_removed", "attack") == 2,
		"want two asset_removed with cause attack")
}

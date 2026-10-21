package scenarios

import (
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

const defTreachery = "SWN-C7-003"

// runOrderedCycle runs one cycle with the turn order pinned, so a scenario does
// not depend on the rolled order.
func runOrderedCycle(t *testing.T, h *testharness.Harness, order ...string) {
	t.Helper()
	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	h.FactionState.CurrentTurn.FactionOrder = order
	if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}
}

// TestTreacheryCaptureAdvancesSeizureSameTurn: alpha's Treachery captures the last
// rival asset on its Planetary Seizure world. Goals read the batch after the
// reactors, so phase 2 starts in the capture turn, not at alpha's next CheckLock
// (Effort 1 Decision Record 12).
func TestTreacheryCaptureAdvancesSeizureSameTurn(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	beta := h.AddFaction("beta", "Hadrian", 4, 3, 2)
	alpha.ActiveGoal = &domain.ActiveGoal{
		GoalID:          "G-004",
		ProcessPhase:    1,
		TargetFactionID: beta.ID,
		TargetWorld:     domain.Location{WorldID: "Hadrian"},
	}
	treachery := addActiveAsset(alpha, "treachery", defTreachery, "Hadrian")
	lastRival := beta.Assets["beta-asset-1"]

	// Rolls: attack 10+3 vs defense 1+3, a hit; the special attack rolls no damage.
	h.Engine.Rand = &testharness.FixedRoller{Values: []int{10, 1}}
	h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID != "alpha" {
			return nil, nil
		}
		for _, candidate := range available {
			if candidate.Name() == "Attack" {
				return candidate, nil
			}
		}
		t.Fatal("Attack not available for alpha")
		return nil, nil
	}
	h.Collector.SelectAttackersFn = func([]*domain.Asset, *rulebook.Rulebook) ([]*domain.Asset, error) {
		return []*domain.Asset{treachery}, nil
	}
	h.Collector.SelectDefenderFn = func(_ *domain.Asset, _ []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
		return lastRival, nil
	}

	runOrderedCycle(t, h, "alpha", "beta")

	testharness.CheckStep(t, "the last rival is alpha's", alpha.Assets[lastRival.ID] != nil, "the rival asset was not captured")
	testharness.CheckStep(t, "phase 2 starts in the capture cycle",
		alpha.ActiveGoal != nil && alpha.ActiveGoal.ProcessPhase == 2 && alpha.ActiveGoal.TurnsRemaining == 3,
		"Planetary Seizure did not advance on the capture")
}

// TestTreacheryCapturesAndKeepsRiders: alpha's Treachery hits beta's Heavy Drop
// Assets in flight, which carries a beta rider. The transport changes sides with
// its order; the rider stays beta's and stays aboard; next cycle the captured
// transport ticks and carries the rider with no miss (Decisions 8, 9 and 10).
func TestTreacheryCapturesAndKeepsRiders(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	beta := h.AddFaction("beta", "Hadrian", 4, 3, 2)

	hex := transportPath[1]
	inFlight := domain.Location{RegionHex: hex}
	treachery := addActiveAsset(alpha, "treachery", defTreachery, "")
	treachery.Location = inFlight
	transport := addActiveAsset(beta, "drop", testharness.DefHeavyDropAssets, "")
	transport.Location = inFlight
	rider := addActiveAsset(beta, "rider", testharness.DefSecurityPersonnel, "")
	rider.Location = inFlight
	transport.CurrentOrder = &domain.MovementOrder{
		AssetID:       transport.ID,
		Origin:        domain.Location{WorldID: "Hadrian"},
		Destination:   domain.Location{WorldID: "Krylos", RegionHex: transportPath[4]},
		Path:          transportPath,
		StepIdx:       1,
		CargoAssetIDs: []string{rider.ID},
	}

	// Rolls: attack 10+3 vs defense 1+3, a hit; the special attack rolls no damage.
	h.Engine.Rand = &testharness.FixedRoller{Values: []int{10, 1}}
	h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID != "alpha" {
			return nil, nil
		}
		for _, candidate := range available {
			if candidate.Name() == "Attack" {
				return candidate, nil
			}
		}
		t.Fatal("Attack not available for alpha")
		return nil, nil
	}
	h.Collector.SelectAttackersFn = func([]*domain.Asset, *rulebook.Rulebook) ([]*domain.Asset, error) {
		return []*domain.Asset{treachery}, nil
	}
	h.Collector.SelectDefenderFn = func(_ *domain.Asset, _ []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
		return transport, nil
	}

	runOrderedCycle(t, h, "alpha", "beta")

	captured, ok := alpha.Assets[transport.ID]
	testharness.CheckStep(t, "the transport is alpha's", ok && captured.OwnerID == "alpha" && beta.Assets[transport.ID] == nil,
		"the transport did not move to alpha")
	testharness.CheckStep(t, "the transport keeps its order and manifest",
		captured.CurrentOrder != nil && captured.CurrentOrder.StepIdx == 1 && slices.Equal(captured.CurrentOrder.CargoAssetIDs, []string{rider.ID}),
		"the order or its manifest changed")
	testharness.CheckStep(t, "the rider stays beta's", beta.Assets[rider.ID] == rider && rider.OwnerID == "beta",
		"the rider changed sides")
	testharness.CheckStep(t, "Treachery is spent", alpha.Assets[treachery.ID] == nil, "Treachery is still in play")
	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	_, found := testharness.FindMutationByTypeAndCause(records, "coin_delta", "capture_target")
	testharness.CheckStep(t, "coin_delta(cause=capture_target) in history", found, "no capture Coin recorded")
	// The next cycle re-registers every hook, so only these checks show the mid-cycle registration.
	testharness.CheckStep(t, "the transport reactor is registered under alpha within the cycle",
		len(h.Engine.Hooks.MutationReactorsForAsset("alpha", transport.ID)) == 1,
		"no reactor under alpha for the captured transport")
	testharness.CheckStep(t, "beta's old bucket for the transport is cleared (Decision Record 13)",
		len(h.Engine.Hooks.MutationReactorsForAsset("beta", transport.ID)) == 0,
		"the transport still has a reactor under beta")
	testharness.CheckStep(t, "only the captured asset registers: Treachery's spent bucket stays until the reset",
		len(h.Engine.Hooks.MutationReactorsForAsset("alpha", treachery.ID)) == 1,
		"Treachery's registration is gone, so every hook was re-primed")

	h.Collector.SelectActionFn = skipAction
	runOrderedCycle(t, h, "alpha", "beta")

	next := transportPath[2]
	testharness.CheckStep(t, "the captured transport ticks", captured.CurrentOrder != nil && captured.Location.RegionHex == next,
		"the transport did not advance under alpha")
	testharness.CheckStep(t, "the rider rides along under beta", beta.Assets[rider.ID] != nil && rider.Location.RegionHex == next,
		"the rider did not follow the captured transport")
}

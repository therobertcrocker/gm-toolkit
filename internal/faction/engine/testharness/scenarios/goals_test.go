package scenarios

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice/dicetest"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

const defHitmen = "SWN-F1-002"

// goalID finds a shipped goal's ID by its name, so the tests do not depend on
// the ID scheme.
func goalID(t *testing.T, h *testharness.Harness, name string) string {
	t.Helper()
	for id, goal := range h.Engine.Rulebook.Goals {
		if goal.Name == name {
			return id
		}
	}
	t.Fatalf("no shipped goal named %q", name)
	return ""
}

// scriptAction makes alpha take the named action and every other faction pass.
// An empty name makes alpha pass too.
func scriptAction(t *testing.T, h *testharness.Harness, actionName string) {
	t.Helper()
	scriptActionBy(t, h, "alpha", actionName)
}

// scriptActionBy makes the named faction take the named action and every other
// faction pass.
func scriptActionBy(t *testing.T, h *testharness.Harness, factionID, actionName string) {
	t.Helper()
	h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID != factionID || actionName == "" {
			return nil, nil
		}
		for _, candidate := range available {
			if candidate.Name() == actionName {
				return candidate, nil
			}
		}
		t.Fatalf("%s not available for %s", actionName, factionID)
		return nil, nil
	}
}

func scriptAttackOnFirstDefender(h *testharness.Harness) {
	h.Collector.SelectAttackersFn = func(eligible []*domain.Asset, _ *rulebook.Rulebook) ([]*domain.Asset, error) {
		return eligible, nil
	}
	h.Collector.SelectDefenderFn = func(_ *domain.Asset, eligible []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
		return eligible[0], nil
	}
}

func runOneCycle(t *testing.T, h *testharness.Harness) {
	t.Helper()
	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}
}

func checkProgress(t *testing.T, faction *domain.Faction, want int) {
	t.Helper()
	if faction.ActiveGoal == nil {
		t.Fatalf("%s has no active goal, want progress %d", faction.ID, want)
	}
	testharness.CheckStep(t, fmt.Sprintf("progress is %d", want), faction.ActiveGoal.Progress == want,
		fmt.Sprintf("progress = %d", faction.ActiveGoal.Progress))
}

// --- Peaceable Kingdom (B.013, D-16) ---

func TestPeaceableKingdom_MissedAttackResets(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 1, 3, 2)
	beta := h.AddFaction("beta", "Tartarus", 2, 2, 2)
	// Hitmen have no counter, so a miss deals no damage either way.
	beta.Assets["beta-asset-1"].DefinitionID = defHitmen
	beta.Assets["beta-asset-1"].CurrentHP = 1
	alpha.ActiveGoal = &domain.ActiveGoal{GoalID: goalID(t, h, "Peaceable Kingdom"), Progress: 2}

	scriptAction(t, h, "Attack")
	scriptAttackOnFirstDefender(h)
	// attack 1+Force(1) = 2 against defense 10+Force(2) = 12: a miss.
	h.Engine.Rand = &dicetest.Sequence{Values: []int{1, 10}}

	runOneCycle(t, h)

	checkProgress(t, alpha, 0)
	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	_, hasReset := testharness.FindMutationByTypeAndCause(records, "goal_progressed", "goal_reset")
	testharness.CheckStep(t, "goal_progressed(cause=goal_reset) in history", hasReset, "no goal reset found")
}

func TestPeaceableKingdom_PassDoesNotCount(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 1, 3, 2)
	alpha.ActiveGoal = &domain.ActiveGoal{GoalID: goalID(t, h, "Peaceable Kingdom"), Progress: 2}

	scriptAction(t, h, "")

	runOneCycle(t, h)

	checkProgress(t, alpha, 2)
}

func TestPeaceableKingdom_PeacefulActionCounts(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 1, 3, 2)
	alpha.CurrentHP = 10
	alpha.ActiveGoal = &domain.ActiveGoal{GoalID: goalID(t, h, "Peaceable Kingdom"), Progress: 2}

	scriptAction(t, h, "Repair Faction")

	runOneCycle(t, h)

	checkProgress(t, alpha, 3)
}

// --- Blood the Enemy (Decision 1) ---

func TestBloodTheEnemy_BaseHitCountsOnce(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 3, 3, 2)
	beta := h.AddFaction("beta", "Tartarus", 2, 2, 2)
	homeworldBase := testharness.AddBase(beta, "Tartarus", 10)
	homeworldBase.IsHomeworld = true
	alpha.ActiveGoal = &domain.ActiveGoal{GoalID: goalID(t, h, "Blood the Enemy")}

	scriptAction(t, h, "Attack")
	scriptAttackOnFirstDefender(h)
	h.Collector.ConfirmRedirectToBaseFn = func(_ *domain.Faction, _ *domain.Base, _ int) (bool, error) {
		return true, nil
	}
	// attack 10+Force(3) = 13 against defense 1+Force(2) = 3: a hit for 2+1 = 3.
	h.Engine.Rand = &dicetest.Sequence{Values: []int{10, 1, 2}}

	runOneCycle(t, h)

	testharness.CheckStep(t, "the Base took the hit", homeworldBase.CurrentHP == 7,
		fmt.Sprintf("Base HP = %d, want 7", homeworldBase.CurrentHP))
	checkProgress(t, alpha, 3)
}

// --- Military Conquest (D-9) ---

func TestMilitaryConquest_RoundsXPUp(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 3, 3, 2)
	h.AddFaction("beta", "Tartarus", 2, 2, 2)
	alpha.ActiveGoal = &domain.ActiveGoal{GoalID: goalID(t, h, "Military Conquest"), Progress: 2}
	xpBefore := alpha.XP

	scriptAction(t, h, "Attack")
	scriptAttackOnFirstDefender(h)
	// attack 10+Force(3) = 13 against defense 1+Force(2) = 3: a hit for 3+1 = 4,
	// which destroys beta's Security Personnel (HP 3). Progress 3 meets Force 3.
	h.Engine.Rand = &dicetest.Sequence{Values: []int{10, 1, 3}}

	runOneCycle(t, h)

	testharness.CheckStep(t, "XP rounds up", alpha.XP-xpBefore == 2,
		fmt.Sprintf("alpha gained %d XP, want 2 (3 kills / 2, rounded up)", alpha.XP-xpBefore))
	testharness.CheckStep(t, "goal cleared after completion", alpha.ActiveGoal == nil, "ActiveGoal should be nil")
}

func TestMilitaryConquest_CounterstrikeOnARivalsTurnCompletes(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 1, 3, 2)
	beta := h.AddFaction("beta", "Tartarus", 2, 2, 2)
	alpha.Assets["alpha-asset-1"].CurrentHP = 1
	beta.ActiveGoal = &domain.ActiveGoal{GoalID: goalID(t, h, "Military Conquest"), Progress: 1}
	xpBefore := beta.XP

	scriptAction(t, h, "Attack")
	scriptAttackOnFirstDefender(h)
	// attack 1+Force(1) = 2 against defense 10+Force(2) = 12: a miss, and beta's
	// Security Personnel counters for 4, which destroys alpha's asset (HP 1).
	h.Engine.Rand = &dicetest.Sequence{Values: []int{1, 10, 4}}

	runOneCycle(t, h)

	testharness.CheckStep(t, "goal cleared after completion", beta.ActiveGoal == nil, "ActiveGoal should be nil")
	testharness.CheckStep(t, "beta gained XP", beta.XP-xpBefore == 1,
		fmt.Sprintf("beta gained %d XP, want 1 (2 kills / 2)", beta.XP-xpBefore))
	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	completed, found := testharness.FindMutationType(records, "goal_completed")
	if !found {
		t.Fatal("no goal_completed in history")
	}
	var payload domain.GoalCompleted
	if err := json.Unmarshal(completed.Payload, &payload); err != nil {
		t.Fatalf("parsing goal_completed: %v", err)
	}
	testharness.CheckStep(t, "goal_completed is beta's", payload.FactionID == "beta",
		fmt.Sprintf("goal_completed faction = %q", payload.FactionID))
}

// --- Destroy the Foe (D-15, D-21) ---

func TestDestroyTheFoe_ThirdPartyBlowFailsTheGoal(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Krylos", 3, 3, 2)
	beta := h.AddFaction("beta", "Tartarus", 2, 2, 2)
	beta.CurrentHP = 1
	homeworldBase := testharness.AddBase(beta, "Tartarus", 10)
	homeworldBase.IsHomeworld = true
	h.AddFaction("gamma", "Tartarus", 3, 3, 2)
	alpha.ActiveGoal = &domain.ActiveGoal{GoalID: goalID(t, h, "Destroy the Foe"), TargetFactionID: "beta"}
	xpBefore := alpha.XP

	scriptActionBy(t, h, "gamma", "Attack")
	scriptAttackOnFirstDefender(h)
	h.Collector.SelectDefenderFn = func(_ *domain.Asset, eligible []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
		for _, defender := range eligible {
			if defender.OwnerID == "beta" {
				return defender, nil
			}
		}
		t.Fatal("no beta asset among the defenders")
		return nil, nil
	}
	h.Collector.ConfirmRedirectToBaseFn = func(_ *domain.Faction, _ *domain.Base, _ int) (bool, error) {
		return true, nil
	}
	// attack 10+Force(3) = 13 against defense 1+Force(2) = 3: a hit for 2+1 = 3 on
	// beta's homeworld Base, which takes beta from 1 HP to -2.
	h.Engine.Rand = &dicetest.Sequence{Values: []int{10, 1, 2}}

	runOneCycle(t, h)

	testharness.CheckStep(t, "goal cleared", alpha.ActiveGoal == nil, "ActiveGoal should be nil")
	testharness.CheckStep(t, "alpha's XP unchanged", alpha.XP == xpBefore,
		fmt.Sprintf("alpha gained %d XP, want 0", alpha.XP-xpBefore))
	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	failed, found := testharness.FindMutationByTypeAndCause(records, "goal_failed", domain.GoalFailedTargetLost)
	if !found {
		t.Fatal("no goal_failed(cause=target_lost) in history")
	}
	var payload domain.GoalFailed
	if err := json.Unmarshal(failed.Payload, &payload); err != nil {
		t.Fatalf("parsing goal_failed: %v", err)
	}
	testharness.CheckStep(t, "goal_failed is alpha's", payload.FactionID == "alpha",
		fmt.Sprintf("goal_failed faction = %q", payload.FactionID))
	_, completed := testharness.FindMutationType(records, "goal_completed")
	testharness.CheckStep(t, "no goal_completed", !completed, "a goal_completed is in history")
}

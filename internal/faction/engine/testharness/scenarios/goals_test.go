package scenarios

import (
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
	h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID != "alpha" || actionName == "" {
			return nil, nil
		}
		for _, candidate := range available {
			if candidate.Name() == actionName {
				return candidate, nil
			}
		}
		t.Fatalf("%s not available for alpha", actionName)
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

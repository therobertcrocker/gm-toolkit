package scenarios

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// TestChangeHomeworld_InitiateThenComplete: end-to-end lifecycle.
//
// StubSpatialMap.Distance always returns 1, so Resolve sets TurnsRemaining = 2.
//
//	Cycle 1 — action phase: ChangeHomeworld fires; ActiveGoal set with TurnsRemaining=2.
//	Cycle 2 — CheckLock: GoalTurnsTick (TurnsRemaining 2→1), LockSkip.
//	Cycle 3 — CheckLock: GoalTurnsTick + HomeworldChanged + GoalCompleted, LockSkip.
func TestChangeHomeworld_InitiateThenComplete(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	testharness.AddBase(alpha, "Krylos", 5)

	goalInitiated := false
	h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if goalInitiated {
			t.Fatalf("SelectAction called for %q after goal initiated (expected LockSkip)", faction.ID)
		}
		for _, a := range available {
			if a.Name() == "Change Homeworld" {
				return a, nil
			}
		}
		t.Fatalf("Change Homeworld not in available actions")
		return nil, nil
	}
	h.Collector.SelectChangeHomeworldTargetFn = func(_ *domain.Faction, _ *state.FactionState) (string, error) {
		return "Krylos", nil
	}

	// Cycle 1 — initiation.
	runCycle(t, h)
	goalInitiated = true

	alpha = h.FactionState.Factions["alpha"]
	if alpha.ActiveGoal == nil {
		t.Fatalf("cycle 1: ActiveGoal is nil, want G-012 initiated")
	}
	if alpha.ActiveGoal.GoalID != "G-012" {
		t.Errorf("cycle 1: ActiveGoal.GoalID = %q, want G-012", alpha.ActiveGoal.GoalID)
	}
	if alpha.ActiveGoal.TargetWorld.WorldID != "Krylos" {
		t.Errorf("cycle 1: ActiveGoal.TargetWorld.WorldID = %q, want Krylos", alpha.ActiveGoal.TargetWorld.WorldID)
	}
	const expectedTurns = 2 // 1 + stub Distance(1)
	if alpha.ActiveGoal.TurnsRemaining != expectedTurns {
		t.Errorf("cycle 1: ActiveGoal.TurnsRemaining = %d, want %d", alpha.ActiveGoal.TurnsRemaining, expectedTurns)
	}

	// Cycle 2 — first tick.
	runCycle(t, h)

	alpha = h.FactionState.Factions["alpha"]
	if alpha.ActiveGoal == nil {
		t.Fatalf("cycle 2: ActiveGoal prematurely cleared")
	}
	if alpha.ActiveGoal.TurnsRemaining != 1 {
		t.Errorf("cycle 2: TurnsRemaining = %d, want 1", alpha.ActiveGoal.TurnsRemaining)
	}

	var cycle2Lock testharness.GoalLockPayload
	for _, ev := range h.Observer.Events {
		if ev.Kind == "GoalLockApplied" && ev.Faction != nil && ev.Faction.ID == "alpha" {
			cycle2Lock = ev.Payload.(testharness.GoalLockPayload)
		}
	}
	if len(cycle2Lock.Mutations) != 1 {
		t.Errorf("cycle 2: GoalLock mutations = %d, want 1 (GoalTurnsTick only)", len(cycle2Lock.Mutations))
	}
	if _, ok := cycle2Lock.Mutations[0].(domain.GoalTurnsTick); !ok {
		t.Errorf("cycle 2: GoalLock.Mutations[0] type = %T, want GoalTurnsTick", cycle2Lock.Mutations[0])
	}

	// Cycle 3 — completion.
	runCycle(t, h)

	alpha = h.FactionState.Factions["alpha"]
	if alpha.ActiveGoal != nil {
		t.Errorf("cycle 3: ActiveGoal = %+v, want nil (goal completed)", alpha.ActiveGoal)
	}
	if alpha.Homeworld.WorldID != "Krylos" {
		t.Errorf("cycle 3: Homeworld.WorldID = %q, want Krylos", alpha.Homeworld.WorldID)
	}

	var cycle3Lock testharness.GoalLockPayload
	for _, ev := range h.Observer.Events {
		if ev.Kind == "GoalLockApplied" && ev.Faction != nil && ev.Faction.ID == "alpha" {
			cycle3Lock = ev.Payload.(testharness.GoalLockPayload)
		}
	}
	if len(cycle3Lock.Mutations) != 3 {
		t.Fatalf("cycle 3: GoalLock mutations = %d, want 3 (tick + HomeworldChanged + GoalCompleted)", len(cycle3Lock.Mutations))
	}
	if _, ok := cycle3Lock.Mutations[0].(domain.GoalTurnsTick); !ok {
		t.Errorf("cycle 3: mutations[0] type = %T, want GoalTurnsTick", cycle3Lock.Mutations[0])
	}
	if _, ok := cycle3Lock.Mutations[1].(domain.HomeworldChanged); !ok {
		t.Errorf("cycle 3: mutations[1] type = %T, want HomeworldChanged", cycle3Lock.Mutations[1])
	}
	if _, ok := cycle3Lock.Mutations[2].(domain.GoalCompleted); !ok {
		t.Errorf("cycle 3: mutations[2] type = %T, want GoalCompleted", cycle3Lock.Mutations[2])
	}
}

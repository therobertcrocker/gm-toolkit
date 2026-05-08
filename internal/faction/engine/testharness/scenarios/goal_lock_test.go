package scenarios

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

// --- scenario 2: LockSkip — Change Homeworld in transit ---

func TestRunCycle_LockSkip_ChangeHomeworld(t *testing.T) {
	h := testharness.NewHarness(t, testDataDir)
	locked := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	locked.ActiveGoal = &domain.ActiveGoal{
		GoalID:         "G-012",
		TargetWorld:    "NewHome",
		TurnsRemaining: 2,
	}
	h.AddFaction("beta", "Hadrian", 4, 3, 2)

	h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID == "alpha" {
			t.Fatalf("SelectAction called for locked faction alpha")
		}
		for _, a := range available {
			if a.Name() == "Sell Asset" {
				return a, nil
			}
		}
		t.Fatalf("Sell Asset not in available actions")
		return nil, nil
	}
	h.Collector.SelectAssetFn = func(assets []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
		return assets[0], nil
	}

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Cfg, h.Collector, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	kinds := h.Observer.Kinds()
	if got, want := testharness.CountKind(kinds, "TurnStarted"), 2; got != want {
		t.Errorf("TurnStarted count: got %d, want %d", got, want)
	}
	if got, want := testharness.CountKind(kinds, "GoalLockApplied"), 2; got != want {
		t.Errorf("GoalLockApplied count: got %d, want %d", got, want)
	}
	if got, want := testharness.CountKind(kinds, "BookkeepingApplied"), 1; got != want {
		t.Errorf("BookkeepingApplied count: got %d, want %d", got, want)
	}
	if got, want := testharness.CountKind(kinds, "ActionSelected"), 1; got != want {
		t.Errorf("ActionSelected count: got %d, want %d", got, want)
	}
	if got, want := testharness.CountKind(kinds, "TurnCompleted"), 2; got != want {
		t.Errorf("TurnCompleted count: got %d, want %d", got, want)
	}
	if got, want := testharness.CountKind(kinds, "CycleCompleted"), 1; got != want {
		t.Errorf("CycleCompleted count: got %d, want %d", got, want)
	}

	var alphaLock testharness.GoalLockPayload
	for _, ev := range h.Observer.Events {
		if ev.Kind == "GoalLockApplied" && ev.Faction != nil && ev.Faction.ID == "alpha" {
			alphaLock = ev.Payload.(testharness.GoalLockPayload)
		}
	}
	if alphaLock.Lock.Type != goal.LockSkip {
		t.Errorf("alpha lock type: got %v, want LockSkip", alphaLock.Lock.Type)
	}
	if len(alphaLock.Mutations) != 1 {
		t.Fatalf("alpha lock mutations: got %d, want 1 (GoalTurnsTick)", len(alphaLock.Mutations))
	}

	records := testharness.ReadHistory(t, h.Cfg.HistoryPath)
	rec, ok := testharness.FindMutationByCause(records, "change_homeworld_transit")
	if !ok {
		t.Fatalf("history missing change_homeworld_transit mutation; records=%+v", records)
	}
	if rec.Type != "goal_turns_tick" {
		t.Errorf("change_homeworld_transit mutation type: got %s, want goal_turns_tick", rec.Type)
	}

	if got := h.FactionState.Factions["alpha"].ActiveGoal.TurnsRemaining; got != 1 {
		t.Errorf("alpha TurnsRemaining: got %d, want 1", got)
	}
}

// --- scenario 3: LockRestrictActions — Planetary Seizure phase 1 ---

func TestRunCycle_LockRestrictActions_PlanetarySeizurePhase1(t *testing.T) {
	h := testharness.NewHarness(t, testDataDir)
	attacker := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	target := h.AddFaction("beta", "Tartarus", 2, 2, 2)
	attacker.ActiveGoal = &domain.ActiveGoal{
		GoalID:          "G-004",
		ProcessPhase:    1,
		TargetFactionID: target.ID,
		TargetWorld:     "Tartarus",
		TurnsRemaining:  3,
	}

	h.Engine.Rand = &testharness.FixedRoller{Values: []int{10, 1, 3}}

	var alphaAvailable []string
	h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID == "alpha" {
			alphaAvailable = nil
			for _, a := range available {
				alphaAvailable = append(alphaAvailable, a.Name())
			}
			for _, a := range available {
				if a.Name() == "Attack" {
					return a, nil
				}
			}
			t.Fatalf("Attack not in alpha's available actions: %v", alphaAvailable)
		}
		return nil, nil
	}
	h.Collector.SelectAttackersFn = func(eligible []*domain.Asset, _ *rulebook.Rulebook) ([]*domain.Asset, error) {
		return eligible, nil
	}
	h.Collector.SelectDefenderFn = func(_ *domain.Asset, eligible []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
		return eligible[0], nil
	}
	h.Collector.ConfirmRedirectToBaseFn = func(_ *domain.Faction, _ *domain.Base, _ int) (bool, error) {
		return false, nil
	}

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Cfg, h.Collector, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	if len(alphaAvailable) != 1 || alphaAvailable[0] != "Attack" {
		t.Fatalf("alpha available actions under LockRestrictActions: got %v, want [Attack]", alphaAvailable)
	}

	resolvedCount := 0
	for _, ev := range h.Observer.Events {
		if ev.Kind == "ActionResolved" {
			resolvedCount++
		}
	}
	if resolvedCount != 1 {
		t.Errorf("ActionResolved count: got %d, want 1", resolvedCount)
	}

	records := testharness.ReadHistory(t, h.Cfg.HistoryPath)
	if _, ok := testharness.FindMutationByCause(records, "attack"); !ok {
		t.Fatalf("history missing attack mutation; records=%+v", records)
	}

	if got := len(h.FactionState.Factions["beta"].Assets); got != 0 {
		t.Errorf("beta assets after attack: got %d, want 0", got)
	}
}

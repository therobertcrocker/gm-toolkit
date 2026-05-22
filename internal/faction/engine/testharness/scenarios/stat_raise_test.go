package scenarios

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
)

func TestRunCycle_StatRaise_XPSpent(t *testing.T) {
	h := testharness.NewHarness(t, testDataDir)
	alpha := h.AddFaction("alpha", "Tartarus", 3, 2, 1)
	alpha.XP = 6 // exactly covers Force 3→4: HPValueForRating(4) = 6

	stat := domain.StatForce
	h.Collector.SelectStatRaiseFn = func(_ *domain.Faction, _ []domain.FactionStat) (*domain.FactionStat, error) {
		return &stat, nil
	}
	h.Collector.SelectActionFn = func(_ *domain.Faction, _ []action.Action) (action.Action, error) {
		return nil, nil
	}

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Cfg, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	testharness.CheckStep(t, "Force raised to 4", alpha.Force == 4, fmt.Sprintf("Force=%d", alpha.Force))
	testharness.CheckStep(t, "XP decremented to 0", alpha.XP == 0, fmt.Sprintf("XP=%d", alpha.XP))
	testharness.CheckStep(t, "MaxHP recalculated", alpha.MaxHP == domain.CalcMaxHP(alpha),
		fmt.Sprintf("MaxHP=%d, CalcMaxHP=%d", alpha.MaxHP, domain.CalcMaxHP(alpha)))

	records := testharness.ReadHistory(t, h.Cfg.HistoryPath)

	_, hasXPSpent := testharness.FindMutationType(records, "xp_spent")
	testharness.CheckStep(t, "xp_spent in history", hasXPSpent, "no xp_spent mutation found")

	rec, hasStatRaised := testharness.FindMutationType(records, "stat_raised")
	testharness.CheckStep(t, "stat_raised in history", hasStatRaised, "no stat_raised mutation found")

	if hasStatRaised {
		var payload struct {
			OldRating int `json:"old_rating"`
			NewRating int `json:"new_rating"`
		}
		if err := json.Unmarshal(rec.Payload, &payload); err != nil {
			t.Fatalf("unmarshalling stat_raised payload: %v", err)
		}
		testharness.CheckStep(t, "OldRating=3", payload.OldRating == 3, fmt.Sprintf("OldRating=%d", payload.OldRating))
		testharness.CheckStep(t, "NewRating=4", payload.NewRating == 4, fmt.Sprintf("NewRating=%d", payload.NewRating))
	}
}

func TestRunCycle_StatRaise_Skip(t *testing.T) {
	h := testharness.NewHarness(t, testDataDir)
	alpha := h.AddFaction("alpha", "Tartarus", 3, 2, 1)
	alpha.XP = 6

	h.Collector.SelectStatRaiseFn = func(_ *domain.Faction, _ []domain.FactionStat) (*domain.FactionStat, error) {
		return nil, nil // player declines
	}
	h.Collector.SelectActionFn = func(_ *domain.Faction, _ []action.Action) (action.Action, error) {
		return nil, nil
	}

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Cfg, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	testharness.CheckStep(t, "Force unchanged at 3", alpha.Force == 3, fmt.Sprintf("Force=%d", alpha.Force))
	testharness.CheckStep(t, "XP unchanged at 6", alpha.XP == 6, fmt.Sprintf("XP=%d", alpha.XP))

	records := testharness.ReadHistory(t, h.Cfg.HistoryPath)
	_, hasXPSpent := testharness.FindMutationType(records, "xp_spent")
	testharness.CheckStep(t, "no xp_spent in history", !hasXPSpent, "unexpected xp_spent mutation found")
}

func TestRunCycle_StatRaise_IneligibleNoPrompt(t *testing.T) {
	h := testharness.NewHarness(t, testDataDir)
	alpha := h.AddFaction("alpha", "Tartarus", 3, 2, 1)
	alpha.XP = 1 // below cheapest raise: Wealth 1→2 costs HPValueForRating(2)=2

	h.Collector.SelectStatRaiseFn = func(_ *domain.Faction, _ []domain.FactionStat) (*domain.FactionStat, error) {
		t.Fatal("SelectStatRaise called when no stats are eligible")
		return nil, nil
	}
	h.Collector.SelectActionFn = func(_ *domain.Faction, _ []action.Action) (action.Action, error) {
		return nil, nil
	}

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Cfg, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	testharness.CheckStep(t, "Force unchanged at 3", alpha.Force == 3, fmt.Sprintf("Force=%d", alpha.Force))
	testharness.CheckStep(t, "XP unchanged at 1", alpha.XP == 1, fmt.Sprintf("XP=%d", alpha.XP))
}

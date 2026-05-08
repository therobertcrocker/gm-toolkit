package integration

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
)

func TestRunCycle_StatRaise_XPSpent(t *testing.T) {
	h := newHarness(t)
	alpha := h.addFaction("alpha", "Tartarus", 3, 2, 1)
	alpha.XP = 6 // exactly covers Force 3→4: HPValueForRating(4) = 6

	stat := domain.StatForce
	h.collector.SelectStatRaiseFn = func(_ *domain.Faction, _ []domain.FactionStat) (*domain.FactionStat, error) {
		return &stat, nil
	}
	h.collector.SelectActionFn = func(_ *domain.Faction, _ []action.Action) (action.Action, error) {
		return nil, nil
	}

	if err := h.engine.Turn.Start(h.factionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.engine.RunCycle(h.factionState, h.cfg, h.collector, h.observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	checkStep(t, "Force raised to 4", alpha.Force == 4, fmt.Sprintf("Force=%d", alpha.Force))
	checkStep(t, "XP decremented to 0", alpha.XP == 0, fmt.Sprintf("XP=%d", alpha.XP))
	checkStep(t, "MaxHP recalculated", alpha.MaxHP == domain.CalcMaxHP(alpha),
		fmt.Sprintf("MaxHP=%d, CalcMaxHP=%d", alpha.MaxHP, domain.CalcMaxHP(alpha)))

	records := readHistory(t, h.cfg.HistoryPath)

	_, hasXPSpent := findMutationType(records, "xp_spent")
	checkStep(t, "xp_spent in history", hasXPSpent, "no xp_spent mutation found")

	rec, hasStatRaised := findMutationType(records, "stat_raised")
	checkStep(t, "stat_raised in history", hasStatRaised, "no stat_raised mutation found")

	if hasStatRaised {
		var payload struct {
			OldRating int `json:"old_rating"`
			NewRating int `json:"new_rating"`
		}
		if err := json.Unmarshal(rec.Payload, &payload); err != nil {
			t.Fatalf("unmarshalling stat_raised payload: %v", err)
		}
		checkStep(t, "OldRating=3", payload.OldRating == 3, fmt.Sprintf("OldRating=%d", payload.OldRating))
		checkStep(t, "NewRating=4", payload.NewRating == 4, fmt.Sprintf("NewRating=%d", payload.NewRating))
	}
}

func TestRunCycle_StatRaise_Skip(t *testing.T) {
	h := newHarness(t)
	alpha := h.addFaction("alpha", "Tartarus", 3, 2, 1)
	alpha.XP = 6

	h.collector.SelectStatRaiseFn = func(_ *domain.Faction, _ []domain.FactionStat) (*domain.FactionStat, error) {
		return nil, nil // player declines
	}
	h.collector.SelectActionFn = func(_ *domain.Faction, _ []action.Action) (action.Action, error) {
		return nil, nil
	}

	if err := h.engine.Turn.Start(h.factionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.engine.RunCycle(h.factionState, h.cfg, h.collector, h.observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	checkStep(t, "Force unchanged at 3", alpha.Force == 3, fmt.Sprintf("Force=%d", alpha.Force))
	checkStep(t, "XP unchanged at 6", alpha.XP == 6, fmt.Sprintf("XP=%d", alpha.XP))

	records := readHistory(t, h.cfg.HistoryPath)
	_, hasXPSpent := findMutationType(records, "xp_spent")
	checkStep(t, "no xp_spent in history", !hasXPSpent, "unexpected xp_spent mutation found")
}

func TestRunCycle_StatRaise_IneligibleNoPrompt(t *testing.T) {
	h := newHarness(t)
	alpha := h.addFaction("alpha", "Tartarus", 3, 2, 1)
	alpha.XP = 1 // below cheapest raise: Wealth 1→2 costs HPValueForRating(2)=2

	h.collector.SelectStatRaiseFn = func(_ *domain.Faction, _ []domain.FactionStat) (*domain.FactionStat, error) {
		t.Fatal("SelectStatRaise called when no stats are eligible")
		return nil, nil
	}
	h.collector.SelectActionFn = func(_ *domain.Faction, _ []action.Action) (action.Action, error) {
		return nil, nil
	}

	if err := h.engine.Turn.Start(h.factionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.engine.RunCycle(h.factionState, h.cfg, h.collector, h.observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	checkStep(t, "Force unchanged at 3", alpha.Force == 3, fmt.Sprintf("Force=%d", alpha.Force))
	checkStep(t, "XP unchanged at 1", alpha.XP == 1, fmt.Sprintf("XP=%d", alpha.XP))
}

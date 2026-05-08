package scenarios

import (
	"encoding/json"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

// --- scenario 1: two-faction full cycle ---

func TestRunCycle_TwoFactionsBothPickSellAsset(t *testing.T) {
	h := testharness.NewHarness(t, testDataDir)
	h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	h.AddFaction("beta", "Hadrian", 4, 3, 2)

	h.Collector.SelectActionFn = func(_ *domain.Faction, available []action.Action) (action.Action, error) {
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

	wantKinds := []string{
		"TurnStarted", "GoalLockApplied", "BookkeepingApplied", "ActionSelected", "ActionResolved", "TurnCompleted",
		"TurnStarted", "GoalLockApplied", "BookkeepingApplied", "ActionSelected", "ActionResolved", "TurnCompleted",
		"CycleCompleted",
	}
	testharness.AssertKinds(t, h.Observer.Kinds(), wantKinds)

	records := testharness.ReadHistory(t, h.Cfg.HistoryPath)
	if len(records) != 4 {
		t.Fatalf("history records: got %d, want 4", len(records))
	}

	wantPerFaction := map[string]int{"alpha": 0, "beta": 0}
	sellRecords := 0
	for _, rec := range records {
		if _, known := wantPerFaction[rec.FactionID]; !known {
			t.Errorf("history: unexpected faction id %q", rec.FactionID)
			continue
		}
		wantPerFaction[rec.FactionID]++
		for _, m := range rec.Mutations {
			var fields struct {
				Cause string `json:"cause"`
			}
			if err := json.Unmarshal(m.Payload, &fields); err == nil && fields.Cause == "sell" {
				sellRecords++
			}
		}
	}
	if wantPerFaction["alpha"] != 2 || wantPerFaction["beta"] != 2 {
		t.Errorf("history record counts: got %v, want each faction = 2", wantPerFaction)
	}
	if sellRecords != 4 {
		t.Errorf("sell mutations in history: got %d, want 4", sellRecords)
	}
}

package scenarios

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
)

func TestConditionExpiresAtSourceTurnStart(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	h.AddFaction("beta", "Hadrian", 4, 3, 2)
	sabotaged := addActiveAsset(alpha, "guards", testharness.DefSecurityPersonnel, "Tartarus")
	sabotaged.Conditions = []domain.Condition{{Kind: domain.ConditionNoAbility, SourceAssetID: "beta-saboteurs", ExpiresAtTurnOf: "beta"}}
	h.Collector.SelectActionFn = skipAction

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	testharness.CheckStep(t, "condition cleared at beta's turn start", len(sabotaged.Conditions) == 0,
		"alpha's asset still carries the no_ability condition")

	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	_, found := testharness.FindMutationByTypeAndCause(records, "asset_condition_cleared", "condition_expired")
	testharness.CheckStep(t, "asset_condition_cleared(cause=condition_expired) in history", found,
		"no asset_condition_cleared found")
}

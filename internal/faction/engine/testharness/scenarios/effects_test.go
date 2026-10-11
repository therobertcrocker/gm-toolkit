package scenarios

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
)

func addPartyMachine(faction *domain.Faction) {
	assetID := faction.ID + "-party-machine"
	faction.Assets[assetID] = &domain.Asset{
		ID:           assetID,
		DefinitionID: testharness.DefPartyMachine,
		OwnerID:      faction.ID,
		Location:     faction.Homeworld,
		CurrentHP:    10,
		Ready:        true,
		Maintained:   true,
	}
}

func skipAction(*domain.Faction, []action.Action) (action.Action, error) { return nil, nil }

func TestTurnStartEffect_PartyMachinePaysAtTurnStart(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	addPartyMachine(alpha)
	h.Collector.SelectActionFn = skipAction
	initialCoin := alpha.Coin

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	// income: wealth(2)/2=1 + (force(4)+cunning(3))/4=1 = 2; Party Machine: +1
	testharness.CheckStep(t, "alpha Coin includes the Party Machine payout", alpha.Coin == initialCoin+3,
		"alpha.Coin should be bookkeeping income plus 1")

	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	_, found := testharness.FindMutationByTypeAndCause(records, "coin_delta", "coin_gain")
	testharness.CheckStep(t, "coin_gain recorded in history", found, "no coin_delta with cause coin_gain")
}

func TestTurnStartEffect_ResumedTurnDoesNotPayTwice(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	addPartyMachine(alpha)
	h.Collector.SelectActionFn = skipAction
	initialCoin := alpha.Coin

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	h.FactionState.CurrentTurn.BookkeepingApplied = true
	if _, err := h.Engine.RunFactionTurn(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunFactionTurn: %v", err)
	}

	testharness.CheckStep(t, "no payout on a turn whose bookkeeping already ran", alpha.Coin == initialCoin,
		"alpha.Coin changed on a resumed turn")
}

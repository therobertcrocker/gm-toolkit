package scenarios

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
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

func TestAttackHitEffect_FranchiseStealsFromTheTarget(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	beta := h.AddFaction("beta", "Tartarus", 2, 2, 2)
	beta.Coin = 5
	franchise := &domain.Asset{
		ID:           "alpha-franchise",
		DefinitionID: testharness.DefFranchise,
		OwnerID:      alpha.ID,
		Location:     alpha.Homeworld,
		CurrentHP:    3,
		Ready:        true,
		Maintained:   true,
	}
	alpha.Assets[franchise.ID] = franchise

	// Roll sequence: [attack, defense, damage]
	// attack:  10 → 10+wealth(2) = 12
	// defense:  1 → 1+wealth(2)  = 3 → hit; Franchise steals a flat 1, no roll
	// damage:   2
	h.Engine.Rand = &testharness.FixedRoller{Values: []int{10, 1, 2}}
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
		return []*domain.Asset{franchise}, nil
	}
	h.Collector.SelectDefenderFn = func(_ *domain.Asset, eligible []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
		return eligible[0], nil
	}

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	_, found := testharness.FindMutationByTypeAndCause(records, "coin_delta", "coin_steal")
	testharness.CheckStep(t, "coin_steal recorded in history", found, "no coin_delta with cause coin_steal")
	testharness.CheckStep(t, "the steal is spent on alpha's budget", alpha.HookBudgets["effect:coin_steal:"+testharness.DefFranchise+":beta"] == 1,
		"alpha.HookBudgets has no spent coin_steal entry for beta")
}

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

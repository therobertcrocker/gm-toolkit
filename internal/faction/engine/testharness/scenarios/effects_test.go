package scenarios

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
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
	testharness.CheckStep(t, "the steal is spent on alpha's budget", alpha.HookBudgets["effect:coin_steal:own_attack_hit:"+testharness.DefFranchise+":beta"] == 1,
		"alpha.HookBudgets has no spent coin_steal entry for beta")
}

func TestRunCycle_SecondCycleOnOneEngineRegistersEachHookOnce(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	h.AddFaction("beta", "Krylos", 2, 2, 2)
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
	h.Collector.SelectActionFn = skipAction

	for cycle := 1; cycle <= 2; cycle++ {
		if err := h.Engine.Turn.Start(h.FactionState); err != nil {
			t.Fatalf("Turn.Start (cycle %d): %v", cycle, err)
		}
		if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
			t.Fatalf("RunCycle (cycle %d): %v", cycle, err)
		}
	}

	reactors := h.Engine.Hooks.MutationReactorsForAsset(alpha.ID, franchise.ID)
	testharness.CheckStep(t, "the Franchise has one reactor after two cycles", len(reactors) == 1,
		fmt.Sprintf("got %d reactors; RunCycle registered the hook again without a reset", len(reactors)))
}

func TestEffects_SurveyorsAddDieToOwnersExpansion(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	h.AddFaction("beta", "Krylos", 2, 2, 2)
	surveyors := &domain.Asset{
		ID:           "alpha-surveyors",
		DefinitionID: testharness.DefSurveyors,
		OwnerID:      alpha.ID,
		Location:     domain.Location{WorldID: "Krylos"},
		CurrentHP:    4,
		Ready:        true,
		Maintained:   true,
	}
	alpha.Assets[surveyors.ID] = surveyors

	// Roll sequence: [expansion, Surveyors' die, rival]
	// expansion: 1, Surveyors: 9 → keep 9 → factionRoll = 9+cunning(3) = 12
	// rival:     5 → rivalRoll = 5+cunning(2) = 7 < 12 → no free attack
	// Without the die the faction roll is 4, and the rival's 7 would earn a free attack.
	h.Engine.Rand = &testharness.FixedRoller{Values: []int{1, 9, 5}}
	h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID != "alpha" {
			return nil, nil
		}
		for _, candidate := range available {
			if candidate.Name() == "Expand Influence" {
				return candidate, nil
			}
		}
		t.Fatal("Expand Influence not available for alpha")
		return nil, nil
	}
	h.Collector.SelectExpandInfluenceOrderFn = func(*domain.Faction, *state.FactionState, []string) (action.ExpandInfluenceOrder, error) {
		return action.ExpandInfluenceOrder{Mode: action.ExpandModeNew, World: "Krylos", HPAmount: 1}, nil
	}
	freeAttackPrompted := false
	h.Collector.ConfirmRivalFreeAttackFn = func(*domain.Faction, int, int) (bool, error) {
		freeAttackPrompted = true
		return false, nil
	}

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	testharness.CheckStep(t, "the Surveyors' die beats the rival, so no free attack is prompted", !freeAttackPrompted,
		"ConfirmRivalFreeAttack was called; the extra die did not reach the expansion roll")
	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	_, hasBase := testharness.FindMutationByTypeAndCause(records, "base_added", "expand")
	testharness.CheckStep(t, "base_added(cause=expand) in history", hasBase, "no base_added found")
}

func TestEffects_PanopticonUnmasksAtOwnersTurnStart(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	beta := h.AddFaction("beta", "Krylos", 2, 2, 2)
	alpha.Assets["alpha-matrix"] = &domain.Asset{
		ID:           "alpha-matrix",
		DefinitionID: testharness.DefPanopticonMatrix,
		OwnerID:      alpha.ID,
		Location:     alpha.Homeworld,
		CurrentHP:    5,
		Ready:        true,
		Maintained:   true,
	}
	spy := &domain.Asset{
		ID:           "beta-spy",
		DefinitionID: testharness.DefSecurityPersonnel,
		OwnerID:      beta.ID,
		Location:     alpha.Homeworld,
		CurrentHP:    3,
		Ready:        true,
		Maintained:   true,
		Stealthy:     true,
	}
	beta.Assets[spy.ID] = spy
	h.Collector.SelectActionFn = skipAction

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	h.FactionState.CurrentTurn.FactionOrder = []string{"beta", "alpha"}

	// Beta's turn: the Matrix's owner is not acting, so no test is rolled.
	h.Engine.Rand = &testharness.FixedRoller{}
	if _, err := h.Engine.RunFactionTurn(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunFactionTurn (beta): %v", err)
	}
	testharness.CheckStep(t, "the spy keeps its stealth on beta's turn", spy.Stealthy,
		"beta-spy lost its stealth when the Matrix's owner was not acting")

	// Alpha's turn, roll sequence: [beta's test, the Matrix's defense]
	// beta:  2 → 2+cunning(2) = 4
	// alpha: 9 → 9+cunning(3) = 12 → beta fails and loses its stealth
	h.Engine.Rand = &testharness.FixedRoller{Values: []int{2, 9}}
	if _, err := h.Engine.RunFactionTurn(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunFactionTurn (alpha): %v", err)
	}
	testharness.CheckStep(t, "the spy loses its stealth at alpha's turn start", !spy.Stealthy,
		"beta-spy is still stealthed after alpha's turn")
	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	_, found := testharness.FindMutationByTypeAndCause(records, "asset_stealth_cleared", "clear_stealth")
	testharness.CheckStep(t, "asset_stealth_cleared(cause=clear_stealth) in history", found, "no asset_stealth_cleared found")
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

func addActiveAsset(faction *domain.Faction, suffix, defID, world string) *domain.Asset {
	asset := &domain.Asset{
		ID:           faction.ID + "-" + suffix,
		DefinitionID: defID,
		OwnerID:      faction.ID,
		Location:     domain.Location{WorldID: world},
		CurrentHP:    5,
		Ready:        true,
		Maintained:   true,
	}
	faction.Assets[asset.ID] = asset
	return asset
}

// countByTypeAndCause counts history mutations of mutType whose payload carries cause.
func countByTypeAndCause(records []domain.EventRecord, mutType, cause string) int {
	count := 0
	for _, record := range records {
		for _, mutation := range record.Mutations {
			if mutation.Type != mutType {
				continue
			}
			var fields struct {
				Cause string `json:"cause"`
			}
			if err := json.Unmarshal(mutation.Payload, &fields); err == nil && fields.Cause == cause {
				count++
			}
		}
	}
	return count
}

func TestEffects_TransitWebMovesForOneCoin(t *testing.T) {
	h := testharness.NewHarness(t)
	h.SpatialMap.Worlds = []string{"Krylos", "Tartarus"}
	alpha := h.AddFaction("alpha", "Tartarus", 4, 4, 7) // AddFaction adds a Security Personnel (Force Military Unit)
	alpha.Coin = 10
	addActiveAsset(alpha, "web", testharness.DefTransitWeb, "Tartarus")
	franchise := addActiveAsset(alpha, "franchise", testharness.DefFranchise, "Tartarus")
	partyMachine := addActiveAsset(alpha, "party-machine", testharness.DefPartyMachine, "Tartarus")
	h.Collector.SelectActionFn = skipAction

	coinAtPrompt := -1
	var offered []string
	h.Collector.SelectRelocateOrdersFn = func(_ *domain.Asset, candidates []*domain.Asset, _ []string, _ int) ([]action.RelocateOrder, error) {
		coinAtPrompt = alpha.Coin
		var orders []action.RelocateOrder
		for _, candidate := range candidates {
			offered = append(offered, candidate.ID)
			orders = append(orders, action.RelocateOrder{Asset: candidate, DestinationWorld: "Krylos"})
		}
		return orders, nil
	}

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	testharness.CheckStep(t, "offered exactly the Franchise and the Party Machine",
		slices.Equal(offered, []string{franchise.ID, partyMachine.ID}),
		fmt.Sprintf("offered = %v (the Web itself, and the Force asset, must not be offered)", offered))
	testharness.CheckStep(t, "both assets are on Krylos",
		franchise.Location.WorldID == "Krylos" && partyMachine.Location.WorldID == "Krylos",
		fmt.Sprintf("franchise on %q, party machine on %q", franchise.Location.WorldID, partyMachine.Location.WorldID))
	testharness.CheckStep(t, "alpha paid exactly 1 Coin for the move", coinAtPrompt-alpha.Coin == 1,
		fmt.Sprintf("Coin at prompt %d, after the turn %d", coinAtPrompt, alpha.Coin))

	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	moves := countByTypeAndCause(records, "asset_moved", "ability")
	testharness.CheckStep(t, "one asset_moved per moved asset in history", moves == 2,
		fmt.Sprintf("asset_moved(cause=ability) = %d, want 2", moves))
}

// scriptPsychicAssassinsPurchase makes the faction's action a Buy Asset of Psychic
// Assassins on its homeworld; every other faction skips its action.
func scriptPsychicAssassinsPurchase(t *testing.T, h *testharness.Harness, buyer *domain.Faction) {
	t.Helper()
	h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID != buyer.ID {
			return nil, nil
		}
		for _, candidate := range available {
			if candidate.Name() == "Buy Asset" {
				return candidate, nil
			}
		}
		t.Fatal("Buy Asset not available")
		return nil, nil
	}
	h.Collector.SelectBuyOrderFn = func(purchasablePerWorld map[string][]*domain.AssetDefinition) (action.BuyOrder, error) {
		for _, def := range purchasablePerWorld[buyer.Homeworld.WorldID] {
			if def.ID == testharness.DefPsychicAssassins {
				return action.BuyOrder{World: buyer.Homeworld.WorldID, Definition: def}, nil
			}
		}
		t.Fatal("Psychic Assassins not purchasable on the homeworld")
		return action.BuyOrder{}, nil
	}
}

func findByDefinition(faction *domain.Faction, defID string) *domain.Asset {
	for _, asset := range domain.SortedAssets(faction) {
		if asset.DefinitionID == defID {
			return asset
		}
	}
	return nil
}

func TestEffects_PsychicAssassinsBoughtStealthed(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 5, 3, 2)
	alpha.Coin = 20
	scriptPsychicAssassinsPurchase(t, h, alpha)

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	assassins := findByDefinition(alpha, testharness.DefPsychicAssassins)
	if assassins == nil {
		t.Fatal("no Psychic Assassins in alpha's roster")
	}
	testharness.CheckStep(t, "the Assassins are stealthed", assassins.Stealthy, "bought Psychic Assassins are not stealthy")
	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	_, found := testharness.FindMutationByTypeAndCause(records, "asset_stealth_applied", "apply_stealth")
	testharness.CheckStep(t, "asset_stealth_applied(cause=apply_stealth) in history", found, "no asset_stealth_applied found")
}

func TestEffects_TripwireCellsCatchAStealthedPurchase(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 5, 3, 2)
	alpha.Coin = 20
	beta := h.AddFaction("beta", "Krylos", 2, 4, 2)
	addActiveAsset(beta, "cells", testharness.DefTripwireCells, "Tartarus")
	scriptPsychicAssassinsPurchase(t, h, alpha)

	// The Cells' contest: beta 10+cunning(4) = 14 vs alpha 1+cunning(3) = 4, so beta wins.
	h.Engine.Rand = &testharness.FixedRoller{Values: []int{10, 1}}
	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	h.FactionState.CurrentTurn.FactionOrder = []string{"alpha", "beta"}
	if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	assassins := findByDefinition(alpha, testharness.DefPsychicAssassins)
	if assassins == nil {
		t.Fatal("no Psychic Assassins in alpha's roster")
	}
	testharness.CheckStep(t, "the Cells unmask the Assassins", !assassins.Stealthy, "the Assassins kept their stealth")
	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	_, found := testharness.FindMutationByTypeAndCause(records, "asset_stealth_cleared", "clear_stealth")
	testharness.CheckStep(t, "asset_stealth_cleared(cause=clear_stealth) in history", found, "no asset_stealth_cleared found")
}

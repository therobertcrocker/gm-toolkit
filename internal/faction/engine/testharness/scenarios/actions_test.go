package scenarios

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// --- scenario 4: Buy Asset — new asset not ready until next cycle ---

func TestRunCycle_BuyAsset_ReadyNextCycle(t *testing.T) {
	h := testharness.NewHarness(t, testDataDir)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)

	// income after bookkeeping: wealth(2)/2=1 + (force(4)+cunning(3))/4=1 → 2 Coin, enough for F1-001 (cost 2).
	secondCycle := false
	h.Collector.SelectActionFn = func(_ *domain.Faction, available []action.Action) (action.Action, error) {
		if secondCycle {
			return nil, nil
		}
		for _, a := range available {
			if a.Name() == "Buy Asset" {
				return a, nil
			}
		}
		t.Fatal("Buy Asset not available in first cycle")
		return nil, nil
	}
	h.Collector.SelectBuyOrderFn = func(purchasablePerWorld map[string][]*domain.AssetDefinition) (action.BuyOrder, error) {
		for world, defs := range purchasablePerWorld {
			for _, def := range defs {
				if def.ID == testharness.DefSecurityPersonnel {
					return action.BuyOrder{World: world, Definition: def}, nil
				}
			}
		}
		t.Fatal("F1-001 not in purchasable list")
		return action.BuyOrder{}, nil
	}

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Cfg, h.Collector, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	records := testharness.ReadHistory(t, h.Cfg.HistoryPath)

	_, hasAdded := testharness.FindMutationType(records, "asset_added")
	testharness.CheckStep(t, "asset_added mutation in history", hasAdded, "no asset_added found")

	_, hasCoin := testharness.FindMutationByTypeAndCause(records, "coin_delta", "buy")
	testharness.CheckStep(t, "coin_delta(cause=buy) in history", hasCoin, "no buy coin_delta found")

	var newAsset *domain.Asset
	for _, asset := range alpha.Assets {
		if asset.DefinitionID == testharness.DefSecurityPersonnel && asset.ID != "alpha-asset-1" {
			newAsset = asset
			break
		}
	}
	testharness.CheckStep(t, "new asset present in faction", newAsset != nil, "no second F1-001 asset found")
	if newAsset != nil {
		testharness.CheckStep(t, "new asset Ready=false after purchase", !newAsset.Ready, "Ready should be false until next cycle")
	}

	// Second cycle: Turn.Start readies all assets.
	secondCycle = true
	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("second Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Cfg, h.Collector, h.Observer); err != nil {
		t.Fatalf("second RunCycle: %v", err)
	}

	if newAsset != nil {
		testharness.CheckStep(t, "new asset Ready=true after second cycle start", newAsset.Ready, "Ready should flip true at Turn.Start")
	}
}

// --- scenario 5: Expand Influence — new base placed uncontested ---

func TestRunCycle_ExpandInfluence_NewBase(t *testing.T) {
	h := testharness.NewHarness(t, testDataDir)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)

	h.Collector.SelectActionFn = func(_ *domain.Faction, available []action.Action) (action.Action, error) {
		for _, a := range available {
			if a.Name() == "Expand Influence" {
				return a, nil
			}
		}
		t.Fatal("Expand Influence not available")
		return nil, nil
	}
	h.Collector.SelectExpandInfluenceOrderFn = func(_ *domain.Faction, _ *state.FactionState, _ []string) (action.ExpandInfluenceOrder, error) {
		return action.ExpandInfluenceOrder{Mode: action.ExpandModeNew, World: "Tartarus", HPAmount: 1}, nil
	}

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Cfg, h.Collector, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	records := testharness.ReadHistory(t, h.Cfg.HistoryPath)

	_, hasBase := testharness.FindMutationByTypeAndCause(records, "base_added", "expand")
	testharness.CheckStep(t, "base_added(cause=expand) in history", hasBase, "no base_added found")

	_, hasCoin := testharness.FindMutationByTypeAndCause(records, "coin_delta", "expand")
	testharness.CheckStep(t, "coin_delta(cause=expand) in history", hasCoin, "no expand coin_delta found")

	hasBaseTartarus := false
	for _, base := range alpha.Bases {
		if base.Location == "Tartarus" {
			hasBaseTartarus = true
			break
		}
	}
	testharness.CheckStep(t, "Tartarus base in faction state", hasBaseTartarus, "no Tartarus base found in alpha.Bases")
}

// --- scenario 6: Expand Influence — contested new base ---

func TestRunCycle_ExpandInfluence_Contested(t *testing.T) {
	h := testharness.NewHarness(t, testDataDir)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	testharness.AddAssetOnWorld(alpha, "Krylos")        // gives alpha an asset on Krylos → eligible for new base there
	h.AddFaction("beta", "Krylos", 2, 2, 2)            // beta's asset is on Krylos → will contest

	// Roll sequence: [expansion, rival, attack, defense, damage]
	// expansion: 1 → factionRoll = 1+cunning(3) = 4
	// rival:    10 → rivalRoll  = 10+cunning(2) = 12 ≥ 4 → attacks
	// attack:   10 → attackRoll = 10+force(2)   = 12
	// defense:   1 → defRoll    = 1+cunning(3)  = 4  → 12≥4: hit, damage 1d3+1
	// damage:    3 → 3+1 = 4 damage → base (HP 1) destroyed
	h.Engine.Rand = &testharness.FixedRoller{Values: []int{1, 10, 10, 1, 3}}

	h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID != "alpha" {
			return nil, nil
		}
		for _, a := range available {
			if a.Name() == "Expand Influence" {
				return a, nil
			}
		}
		t.Fatal("Expand Influence not available for alpha")
		return nil, nil
	}
	h.Collector.SelectExpandInfluenceOrderFn = func(_ *domain.Faction, _ *state.FactionState, _ []string) (action.ExpandInfluenceOrder, error) {
		return action.ExpandInfluenceOrder{Mode: action.ExpandModeNew, World: "Krylos", HPAmount: 1}, nil
	}
	h.Collector.ConfirmRivalFreeAttackFn = func(_ *domain.Faction, _, _ int) (bool, error) {
		return true, nil
	}
	h.Collector.SelectBaseAttackersFn = func(_ *domain.Faction, eligible []*domain.Asset, _ *rulebook.Rulebook) ([]*domain.Asset, error) {
		return eligible, nil
	}

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Cfg, h.Collector, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	records := testharness.ReadHistory(t, h.Cfg.HistoryPath)

	_, hasBase := testharness.FindMutationByTypeAndCause(records, "base_added", "expand")
	testharness.CheckStep(t, "base_added(cause=expand) in history", hasBase, "no base_added found")

	_, hasHPDelta := testharness.FindMutationByTypeAndCause(records, "base_hp_delta", "expand")
	testharness.CheckStep(t, "base_hp_delta(cause=expand) in history — rival attacked new base", hasHPDelta, "no base_hp_delta found")

	_, hasDestroyed := testharness.FindMutationByTypeAndCause(records, "base_destroyed", "expand")
	testharness.CheckStep(t, "base_destroyed(cause=expand) in history — base HP driven to 0", hasDestroyed, "no base_destroyed found")
}

// --- scenario 7: Goal completed mid-cycle via MilitaryConquest ---

func TestRunCycle_GoalCompleted_MilitaryConquest(t *testing.T) {
	h := testharness.NewHarness(t, testDataDir)
	alpha := h.AddFaction("alpha", "Tartarus", 1, 3, 2)
	h.AddFaction("beta", "Tartarus", 2, 2, 2)

	// Progress=1, Force=1 → one Force kill pushes newProgress=2 ≥ Force=1 → complete (xp=2/2=1).
	alpha.ActiveGoal = &domain.ActiveGoal{GoalID: "G-001", Progress: 1}

	// Roll sequence: [attack, defense, damage]
	// attack:  10 → attackRoll = 10+force(1) = 11
	// defense:  1 → defRoll   = 1+force(2)  = 3  → 11≥3: hit
	// damage:   3 → 3+1 = 4 ≥ asset HP(3)        → destroyed
	h.Engine.Rand = &testharness.FixedRoller{Values: []int{10, 1, 3}}

	h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID != "alpha" {
			return nil, nil
		}
		for _, a := range available {
			if a.Name() == "Attack" {
				return a, nil
			}
		}
		t.Fatal("Attack not available for alpha")
		return nil, nil
	}
	h.Collector.SelectAttackersFn = func(eligible []*domain.Asset, _ *rulebook.Rulebook) ([]*domain.Asset, error) {
		return eligible, nil
	}
	h.Collector.SelectDefenderFn = func(_ *domain.Asset, eligible []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
		return eligible[0], nil
	}

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Cfg, h.Collector, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	records := testharness.ReadHistory(t, h.Cfg.HistoryPath)

	_, hasCompleted := testharness.FindMutationByCause(records, "goal_completed")
	testharness.CheckStep(t, "goal_completed mutation in history", hasCompleted, "no goal_completed mutation found")

	_, hasXP := testharness.FindMutationType(records, "xp_awarded")
	testharness.CheckStep(t, "xp_awarded mutation in history", hasXP, "no xp_awarded mutation found")

	testharness.CheckStep(t, "alpha.ActiveGoal cleared after completion", alpha.ActiveGoal == nil, "ActiveGoal should be nil")
}

// --- scenario 8: Bribe — rival base takes influence, attacker spends coin ---

func TestRunCycle_Bribe(t *testing.T) {
	h := testharness.NewHarness(t, testDataDir)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	beta := h.AddFaction("beta", "Tartarus", 2, 2, 2)

	// Bribe.Validate requires len(faction.Bases) > 0 on the acting faction.
	testharness.AddBase(alpha, "Tartarus", 3)
	rivalBase := testharness.AddBase(beta, "Tartarus", 3)

	h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID != "alpha" {
			return nil, nil
		}
		for _, a := range available {
			if a.Name() == "Bribe" {
				return a, nil
			}
		}
		t.Fatal("Bribe not available for alpha")
		return nil, nil
	}
	h.Collector.SelectBribeTargetFn = func(_ *domain.Faction, _ *state.FactionState) (*domain.Base, int, error) {
		return rivalBase, 1, nil
	}

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Cfg, h.Collector, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	records := testharness.ReadHistory(t, h.Cfg.HistoryPath)

	_, hasInfluence := testharness.FindMutationByTypeAndCause(records, "influence_delta", "bribe")
	testharness.CheckStep(t, "influence_delta(cause=bribe) in history", hasInfluence, "no bribe influence_delta found")

	_, hasCoin := testharness.FindMutationByTypeAndCause(records, "coin_delta", "bribe")
	testharness.CheckStep(t, "coin_delta(cause=bribe) in history", hasCoin, "no bribe coin_delta found")
}

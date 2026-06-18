package scenarios

import (
	"fmt"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

// --- scenario N: Scavengers tag — +1 Coin per asset destroyed ---

func TestScavengers_GrantsCoinOnKill(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	alpha.Tags = []*domain.Tag{{ID: "T-016"}}
	h.AddFaction("beta", "Tartarus", 2, 2, 2)

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
		t.Fatalf("Attack not available for alpha")
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
	if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	if got := len(h.FactionState.Factions["beta"].Assets); got != 0 {
		t.Fatalf("beta assets: got %d, want 0 (attack should have destroyed it)", got)
	}

	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	if _, ok := testharness.FindMutationByTypeAndCause(records, "coin_delta", "scavengers"); !ok {
		t.Error("expected coin_delta with cause=scavengers in history")
	}
}

func TestScavengers_NoBonusWithoutTag(t *testing.T) {
	h := testharness.NewHarness(t)
	h.AddFaction("alpha", "Tartarus", 4, 3, 2) // no Scavengers tag
	h.AddFaction("beta", "Tartarus", 2, 2, 2)

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
		t.Fatalf("Attack not available for alpha")
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
	if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	if _, ok := testharness.FindMutationByTypeAndCause(records, "coin_delta", "scavengers"); ok {
		t.Error("expected no coin_delta with cause=scavengers when Scavengers tag is absent")
	}
}

// --- scenario N+1: Warlike tag — +1d10 keep highest on Force attack ---

// TestWarlike_FiresOnForceAttack_ConsumesOneBudget verifies that Warlike's bonus die
// is applied, changes the outcome when the base die alone would miss, and consumes
// one budget slot for the turn.
//
// FixedRoller [2, 9, 5, 3]:
//
//	attack-base=2, warlike-bonus=9 → keep max(2,9)=9; +force(1)=10
//	defense=5+force(4)=9  →  10>9: hit; damage=3+1=4 ≥ HP(3) → destroyed
//	(without Warlike: 2+1=3 < 5+4=9 → miss)
func TestWarlike_FiresOnForceAttack_ConsumesOneBudget(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 1, 3, 2)
	alpha.Tags = []*domain.Tag{{ID: "T-020"}}
	h.AddFaction("beta", "Tartarus", 4, 3, 2)

	h.Engine.Rand = &testharness.FixedRoller{Values: []int{2, 9, 5, 3}}
	h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID != "alpha" {
			return nil, nil
		}
		for _, a := range available {
			if a.Name() == "Attack" {
				return a, nil
			}
		}
		t.Fatalf("Attack not available for alpha")
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
	if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	testharness.CheckStep(t, "beta asset destroyed by Warlike-boosted attack", len(h.FactionState.Factions["beta"].Assets) == 0,
		"beta should have no assets after attack")

	budgets := h.FactionState.Factions["alpha"].HookBudgets
	testharness.CheckStep(t, "Warlike budget consumed once", budgets["tag:Warlike"] == 1,
		fmt.Sprintf("HookBudgets[tag:Warlike] = %d, want 1", budgets["tag:Warlike"]))
}

// --- scenario N+2: Fanatical tag — auto-reroll 1s ---

// TestFanatical_RerollsOnesInAttack verifies that a rolled 1 is rerolled automatically.
//
// FixedRoller [1, 9, 6, 3]:
//
//	attack-base=1 → Fanatical rerolls → 9+force(4)=13
//	defense=6+force(2)=8  →  13>8: hit; damage=3+1=4 → destroyed
//	(without Fanatical: 1+4=5 < 6+2=8 → miss)
func TestFanatical_RerollsOnesInAttack(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	alpha.Tags = []*domain.Tag{{ID: "T-005"}}
	h.AddFaction("beta", "Tartarus", 2, 3, 2)

	h.Engine.Rand = &testharness.FixedRoller{Values: []int{1, 9, 6, 3}}
	h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID != "alpha" {
			return nil, nil
		}
		for _, a := range available {
			if a.Name() == "Attack" {
				return a, nil
			}
		}
		t.Fatalf("Attack not available for alpha")
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
	if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	testharness.CheckStep(t, "beta asset destroyed (reroll changed outcome)", len(h.FactionState.Factions["beta"].Assets) == 0,
		"beta should have no assets after Fanatical-rerolled attack")
}

// --- scenario N+3: Fanatical tag — tie loss as attacker ---

// TestFanatical_TieLoss_AttackerLoses verifies that a Fanatical attacker always
// loses ties: no damage to the defender, but counter fires and hits the attacker.
//
// FixedRoller [6, 3, 3]:
//
//	attack=6+force(1)=7, defense=3+force(4)=7 → TIE
//	TieDefenderWins → no attack damage; counter=3 destroys alpha's asset
func TestFanatical_TieLoss_AttackerLoses(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 1, 3, 2)
	alpha.Tags = []*domain.Tag{{ID: "T-005"}}
	h.AddFaction("beta", "Tartarus", 4, 3, 2)

	h.Engine.Rand = &testharness.FixedRoller{Values: []int{6, 3, 3}}
	h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID != "alpha" {
			return nil, nil
		}
		for _, a := range available {
			if a.Name() == "Attack" {
				return a, nil
			}
		}
		t.Fatalf("Attack not available for alpha")
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
	if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	testharness.CheckStep(t, "beta asset intact (Fanatical attacker lost tie)", len(h.FactionState.Factions["beta"].Assets) == 1,
		fmt.Sprintf("beta.Assets: got %d, want 1", len(h.FactionState.Factions["beta"].Assets)))
	testharness.CheckStep(t, "alpha asset destroyed by counter after tie loss", len(h.FactionState.Factions["alpha"].Assets) == 0,
		fmt.Sprintf("alpha.Assets: got %d, want 0", len(h.FactionState.Factions["alpha"].Assets)))
}

// --- scenario N+4: Preceptor Archive tag — -1 Coin on TL4+ asset purchases ---

// TestPreceptorArchive_ReducesCostOnTL4Asset verifies that buying a TL4 asset
// costs one fewer Coin when the faction has the Preceptor Archive tag.
//
// Setup: Force=2, Cunning=3, Wealth=2, initial Coin=4.
// Income from bookkeeping: wealth(2)/2=1 + (force(2)+cunning(3))/4=1 = 2.
// Pre-buy Coin = 6. SWN-F2-001 (Heavy Drop Assets) base cost=4, TL4.
// Preceptor reduces cost to 3 → final Coin = 3.
func TestPreceptorArchive_ReducesCostOnTL4Asset(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 2, 3, 2)
	alpha.Tags = []*domain.Tag{{ID: "T-013"}}
	alpha.Coin = 4

	h.Collector.SelectActionFn = func(_ *domain.Faction, available []action.Action) (action.Action, error) {
		for _, a := range available {
			if a.Name() == "Buy Asset" {
				return a, nil
			}
		}
		t.Fatal("Buy Asset not available")
		return nil, nil
	}
	h.Collector.SelectBuyOrderFn = func(purchasablePerWorld map[string][]*domain.AssetDefinition) (action.BuyOrder, error) {
		for world, defs := range purchasablePerWorld {
			for _, def := range defs {
				if def.ID == testharness.DefHeavyDropAssets {
					return action.BuyOrder{World: world, Definition: def}, nil
				}
			}
		}
		t.Fatal("SWN-F2-001 not in purchasable list")
		return action.BuyOrder{}, nil
	}

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	// income=2, reduced cost=3 → 4+2-3=3
	const wantCoin = 3
	testharness.CheckStep(t, "alpha Coin reflects Preceptor discount (cost 3, not 4)", alpha.Coin == wantCoin,
		fmt.Sprintf("alpha.Coin = %d, want %d", alpha.Coin, wantCoin))
}

// TestPreceptorArchive_NoBonusWithoutTag verifies that full cost is charged
// when the faction does not have the Preceptor Archive tag.
//
// Same setup as above: pre-buy Coin = 6, base cost = 4 → final Coin = 2.
func TestPreceptorArchive_NoBonusWithoutTag(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 2, 3, 2)
	alpha.Coin = 4

	h.Collector.SelectActionFn = func(_ *domain.Faction, available []action.Action) (action.Action, error) {
		for _, a := range available {
			if a.Name() == "Buy Asset" {
				return a, nil
			}
		}
		t.Fatal("Buy Asset not available")
		return nil, nil
	}
	h.Collector.SelectBuyOrderFn = func(purchasablePerWorld map[string][]*domain.AssetDefinition) (action.BuyOrder, error) {
		for world, defs := range purchasablePerWorld {
			for _, def := range defs {
				if def.ID == testharness.DefHeavyDropAssets {
					return action.BuyOrder{World: world, Definition: def}, nil
				}
			}
		}
		t.Fatal("SWN-F2-001 not in purchasable list")
		return action.BuyOrder{}, nil
	}

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	// income=2, full cost=4 → 4+2-4=2
	const wantCoin = 2
	testharness.CheckStep(t, "alpha Coin charged full cost without Preceptor tag", alpha.Coin == wantCoin,
		fmt.Sprintf("alpha.Coin = %d, want %d", alpha.Coin, wantCoin))
}

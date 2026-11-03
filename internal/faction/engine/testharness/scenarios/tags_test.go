package scenarios

import (
	"fmt"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice/dicetest"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

// --- scenario N: Scavengers tag — +1 Coin per asset destroyed ---

func TestScavengers_GrantsCoinOnKill(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	alpha.Tags = []string{"SWN-T-016"}
	h.AddFaction("beta", "Tartarus", 2, 2, 2)

	h.Engine.Rand = &dicetest.Sequence{Values: []int{10, 1, 3}}
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
	if _, ok := testharness.FindMutationByTypeAndCause(records, "coin_delta", "coin_gain"); !ok {
		t.Error("expected coin_delta with cause=coin_gain in history")
	}
}

func TestScavengers_NoBonusWithoutTag(t *testing.T) {
	h := testharness.NewHarness(t)
	h.AddFaction("alpha", "Tartarus", 4, 3, 2) // no Scavengers tag
	h.AddFaction("beta", "Tartarus", 2, 2, 2)

	h.Engine.Rand = &dicetest.Sequence{Values: []int{10, 1, 3}}
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
	if _, ok := testharness.FindMutationByTypeAndCause(records, "coin_delta", "coin_gain"); ok {
		t.Error("expected no coin_delta with cause=coin_gain when Scavengers tag is absent")
	}
}

// --- scenario N+1: Warlike tag — +1d10 keep highest on Force attack ---

// TestWarlike_FiresOnForceAttack_ConsumesOneBudget verifies that Warlike's bonus die
// is applied, changes the outcome when the base die alone would miss, and consumes
// one budget slot for the turn.
//
// Sequence [2, 9, 5, 3]:
//
//	attack-base=2, warlike-bonus=9 → keep max(2,9)=9; +force(1)=10
//	defense=5+force(4)=9  →  10>9: hit; damage=3+1=4 ≥ HP(3) → destroyed
//	(without Warlike: 2+1=3 < 5+4=9 → miss)
func TestWarlike_FiresOnForceAttack_ConsumesOneBudget(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 1, 3, 2)
	alpha.Tags = []string{"SWN-T-020"}
	h.AddFaction("beta", "Tartarus", 4, 3, 2)

	h.Engine.Rand = &dicetest.Sequence{Values: []int{2, 9, 5, 3}}
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

	const warlikeBudget = "effect:bonus_die:own_roll:SWN-T-020"
	budgets := h.FactionState.Factions["alpha"].HookBudgets
	testharness.CheckStep(t, "Warlike budget consumed once", budgets[warlikeBudget] == 1,
		fmt.Sprintf("HookBudgets[%s] = %d, want 1", warlikeBudget, budgets[warlikeBudget]))
}

// --- scenario N+2: Fanatical tag — auto-reroll 1s ---

// TestFanatical_RerollsOnesInAttack verifies that a rolled 1 is rerolled automatically.
//
// Sequence [1, 9, 6, 3]:
//
//	attack-base=1 → Fanatical rerolls → 9+force(4)=13
//	defense=6+force(2)=8  →  13>8: hit; damage=3+1=4 → destroyed
//	(without Fanatical: 1+4=5 < 6+2=8 → miss)
func TestFanatical_RerollsOnesInAttack(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	alpha.Tags = []string{"SWN-T-005"}
	h.AddFaction("beta", "Tartarus", 2, 3, 2)

	h.Engine.Rand = &dicetest.Sequence{Values: []int{1, 9, 6, 3}}
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
// Sequence [6, 3, 3]:
//
//	attack=6+force(1)=7, defense=3+force(4)=7 → TIE
//	TieDefenderWins → no attack damage; counter=3 destroys alpha's asset
func TestFanatical_TieLoss_AttackerLoses(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 1, 3, 2)
	alpha.Tags = []string{"SWN-T-005"}
	h.AddFaction("beta", "Tartarus", 4, 3, 2)

	h.Engine.Rand = &dicetest.Sequence{Values: []int{6, 3, 3}}
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
	alpha.Tags = []string{"SWN-T-013"}
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

// --- scenario N+5: Secretive tag — every purchase starts stealthed ---

func TestSecretive_PurchaseStartsStealthed(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 5, 3, 2)
	alpha.Tags = []string{"SWN-T-017"}
	alpha.Coin = 20
	scriptPurchase(t, h, alpha, testharness.DefFranchise)

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	franchise := findByDefinition(alpha, testharness.DefFranchise)
	if franchise == nil {
		t.Fatal("no Franchise in alpha's roster")
	}
	testharness.CheckStep(t, "the Franchise is stealthed", franchise.Stealthy, "a Secretive faction's Franchise is not stealthy")
	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	_, found := testharness.FindMutationByTypeAndCause(records, "asset_stealth_applied", "apply_stealth")
	testharness.CheckStep(t, "asset_stealth_applied(cause=apply_stealth) in history", found, "no asset_stealth_applied found")
}

// --- scenario N+6: Deep Rooted tag — lost when the homeworld changes ---

func TestDeepRooted_LostOnHomeworldChange(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	alpha.Tags = []string{"SWN-T-002"}
	alpha.ActiveGoal = &domain.ActiveGoal{GoalID: "G-012", TargetWorld: domain.Location{WorldID: "Krylos"}, TurnsRemaining: 1}
	h.Collector.SelectActionFn = skipAction

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	testharness.CheckStep(t, "alpha moved home to Krylos", alpha.Homeworld.WorldID == "Krylos",
		fmt.Sprintf("alpha.Homeworld = %q, want Krylos", alpha.Homeworld.WorldID))
	testharness.CheckStep(t, "Deep Rooted is gone", !alpha.HasTag("SWN-T-002"),
		fmt.Sprintf("alpha.Tags = %v, want no SWN-T-002", alpha.Tags))
	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	_, found := testharness.FindMutationByTypeAndCause(records, "tag_removed", "remove_tag")
	testharness.CheckStep(t, "tag_removed(cause=remove_tag) in history", found, "no tag_removed found")
}

// --- scenario N+7: Exchange Consulate tag — bonus XP on Peaceable Kingdom ---

func TestExchangeConsulate_BonusXPOnPeaceableKingdom(t *testing.T) {
	cases := []struct {
		name   string
		roll   int
		wantXP int
	}{
		{"a 4 gains the bonus", 4, 2},
		{"a 3 does not", 3, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := testharness.NewHarness(t)
			alpha := h.AddFaction("alpha", "Tartarus", 5, 3, 2)
			alpha.Tags = []string{"SWN-T-004"}
			alpha.Coin = 20
			alpha.ActiveGoal = &domain.ActiveGoal{GoalID: "G-007", Progress: 3}
			xpBefore := alpha.XP
			scriptPurchase(t, h, alpha, testharness.DefFranchise)
			h.Engine.Rand = &dicetest.Sequence{Values: []int{tc.roll}}

			if err := h.Engine.Turn.Start(h.FactionState); err != nil {
				t.Fatalf("Turn.Start: %v", err)
			}
			if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
				t.Fatalf("RunCycle: %v", err)
			}

			testharness.CheckStep(t, "XP gained", alpha.XP-xpBefore == tc.wantXP,
				fmt.Sprintf("alpha gained %d XP, want %d (1 for the goal, 1 more on a 4+)", alpha.XP-xpBefore, tc.wantXP))
		})
	}
}

// --- scenario N+8: Colonists and Technical Expertise — Starships on a small world ---

func TestPopulationFloor_TagsOpenASmallWorldToStarships(t *testing.T) {
	cases := []struct {
		name  string
		tags  []string
		offer bool
	}{
		{"no tag: the setting's 500,000 floor", nil, false},
		{"Colonists: 20% of the floor", []string{"SWN-T-001"}, true},
		{"Technical Expertise: 2% of the floor", []string{"SWN-T-018"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := testharness.NewHarness(t)
			h.SpatialMap.Populations = map[string]int{"Tartarus": 150_000}
			alpha := h.AddFaction("alpha", "Tartarus", 3, 3, 3)
			alpha.Tags = tc.tags
			alpha.Coin = 20
			offered := false
			h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
				if faction.ID != "alpha" {
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
			h.Collector.SelectBuyOrderFn = func(purchasable map[string][]*domain.AssetDefinition) (action.BuyOrder, error) {
				var fallback *domain.AssetDefinition
				for _, def := range purchasable["Tartarus"] {
					if def.ID == testharness.DefSmugglers {
						offered = true
					}
					if def.ID == testharness.DefSecurityPersonnel {
						fallback = def
					}
				}
				return action.BuyOrder{World: "Tartarus", Definition: fallback}, nil
			}

			if err := h.Engine.Turn.Start(h.FactionState); err != nil {
				t.Fatalf("Turn.Start: %v", err)
			}
			if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
				t.Fatalf("RunCycle: %v", err)
			}

			testharness.CheckStep(t, "Smugglers offered on a world of 150,000", offered == tc.offer,
				fmt.Sprintf("Smugglers offered = %v, want %v", offered, tc.offer))
		})
	}
}

// --- scenario N+9: Preceptor Archive — Teach Planetary Population ---

// TestPreceptorArchive_TeachPlanetaryPopulation: alpha holds Preceptor Archive and
// a Base on Backwater (TL 2); Tartarus, its homeworld, has no Base. Use Asset
// Ability offers the tag alone, and the world target offers Backwater alone. The
// action pays 2 Coin either way; on a 12, Backwater counts as TL4 for alpha.
func TestPreceptorArchive_TeachPlanetaryPopulation(t *testing.T) {
	cases := []struct {
		name   string
		roll   int
		wantTL int
	}{
		{"a 12 raises the world", 12, 4},
		{"an 11 pays and raises nothing", 11, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := testharness.NewHarness(t)
			h.SpatialMap.TechLevels = map[string]int{"Backwater": 2}
			alpha := h.AddFaction("alpha", "Tartarus", 2, 3, 2)
			alpha.Tags = []string{"SWN-T-013"}
			alpha.Coin = 4
			testharness.AddBase(alpha, "Backwater", 2)
			h.Engine.Rand = &dicetest.Sequence{Values: []int{tc.roll}}
			var offered []domain.AbilityHolder
			var worlds []string
			h.Collector.SelectActionFn = takeAction(t, alpha.ID, "Use Asset Ability")
			h.Collector.SelectAbilityHoldersFn = func(_ *domain.Faction, candidates []domain.AbilityHolder, _ *rulebook.Rulebook) ([]domain.AbilityHolder, error) {
				offered = candidates
				return candidates, nil
			}
			h.Collector.SelectAbilityTargetFn = func(_ domain.AbilityHolder, _ domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
				worlds = candidates.Worlds
				return domain.AbilityTarget{Kind: domain.TargetWorld, World: "Backwater"}, nil
			}

			runOrderedCycle(t, h, alpha.ID)

			testharness.CheckStep(t, "the tag is the only holder offered", len(offered) == 1 && offered[0].Tag != nil && offered[0].Tag.ID == "SWN-T-013",
				fmt.Sprintf("offered = %+v", offered))
			testharness.CheckStep(t, "Backwater is the only world offered", len(worlds) == 1 && worlds[0] == "Backwater",
				fmt.Sprintf("worlds = %v", worlds))
			testharness.CheckStep(t, "the record matches the roll", alpha.WorldTechLevels["Backwater"] == tc.wantTL,
				fmt.Sprintf("WorldTechLevels[Backwater] = %d, want %d", alpha.WorldTechLevels["Backwater"], tc.wantTL))
			records := testharness.ReadHistory(t, h.Paths.HistoryPath)
			_, paid := testharness.FindMutationByTypeAndCause(records, "coin_delta", "ability")
			testharness.CheckStep(t, "the 2 Coin is paid", paid, "no coin_delta(cause=ability) in history")
		})
	}
}

// --- scenario N+10: Eugenics Cult — Gengineered Slaves and their die ---

// TestEugenicsCult_OnlyTheTagBuysGengineeredSlaves: Buy Asset offers Gengineered
// Slaves on Tartarus (TL 5) to a faction with Eugenics Cult, and not to one
// without it.
func TestEugenicsCult_OnlyTheTagBuysGengineeredSlaves(t *testing.T) {
	cases := []struct {
		name  string
		tags  []string
		offer bool
	}{
		{"no tag", nil, false},
		{"Eugenics Cult", []string{"SWN-T-003"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := testharness.NewHarness(t)
			alpha := h.AddFaction("alpha", "Tartarus", 2, 2, 2)
			alpha.Tags = tc.tags
			alpha.Coin = 4
			offered := false
			h.Collector.SelectActionFn = takeAction(t, alpha.ID, "Buy Asset")
			h.Collector.SelectBuyOrderFn = func(purchasable map[string][]*domain.AssetDefinition) (action.BuyOrder, error) {
				var fallback *domain.AssetDefinition
				for _, def := range purchasable["Tartarus"] {
					if def.ID == "SWN-F1-004" {
						offered = true
					}
					if def.ID == testharness.DefSecurityPersonnel {
						fallback = def
					}
				}
				return action.BuyOrder{World: "Tartarus", Definition: fallback}, nil
			}

			runOrderedCycle(t, h, alpha.ID)

			testharness.CheckStep(t, "Gengineered Slaves offered", offered == tc.offer,
				fmt.Sprintf("offered = %v, want %v", offered, tc.offer))
		})
	}
}

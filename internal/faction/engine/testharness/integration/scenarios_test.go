package integration

import (
	"encoding/json"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// --- scenario 1: two-faction full cycle ---

func TestRunCycle_TwoFactionsBothPickSellAsset(t *testing.T) {
	h := newHarness(t)
	h.addFaction("alpha", "Tartarus", 4, 3, 2)
	h.addFaction("beta", "Hadrian", 4, 3, 2)

	h.collector.SelectActionFn = func(_ *domain.Faction, available []action.Action) (action.Action, error) {
		for _, a := range available {
			if a.Name() == "Sell Asset" {
				return a, nil
			}
		}
		t.Fatalf("Sell Asset not in available actions")
		return nil, nil
	}
	h.collector.SelectAssetFn = func(assets []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
		return assets[0], nil
	}

	if err := h.engine.Turn.Start(h.factionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.engine.RunCycle(h.factionState, h.cfg, h.collector, h.observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	wantKinds := []string{
		"TurnStarted", "GoalLockApplied", "BookkeepingApplied", "ActionSelected", "ActionResolved", "TurnCompleted",
		"TurnStarted", "GoalLockApplied", "BookkeepingApplied", "ActionSelected", "ActionResolved", "TurnCompleted",
		"CycleCompleted",
	}
	assertKinds(t, h.observer.Kinds(), wantKinds)

	records := readHistory(t, h.cfg.HistoryPath)
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

// --- scenario 2: LockSkip — Change Homeworld in transit ---

func TestRunCycle_LockSkip_ChangeHomeworld(t *testing.T) {
	h := newHarness(t)
	locked := h.addFaction("alpha", "Tartarus", 4, 3, 2)
	locked.ActiveGoal = &domain.ActiveGoal{
		GoalID:         "G-012",
		TargetWorld:    "NewHome",
		TurnsRemaining: 2,
	}
	h.addFaction("beta", "Hadrian", 4, 3, 2)

	h.collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID == "alpha" {
			t.Fatalf("SelectAction called for locked faction alpha")
		}
		for _, a := range available {
			if a.Name() == "Sell Asset" {
				return a, nil
			}
		}
		t.Fatalf("Sell Asset not in available actions")
		return nil, nil
	}
	h.collector.SelectAssetFn = func(assets []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
		return assets[0], nil
	}

	if err := h.engine.Turn.Start(h.factionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.engine.RunCycle(h.factionState, h.cfg, h.collector, h.observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	kinds := h.observer.Kinds()
	if got, want := countKind(kinds, "TurnStarted"), 2; got != want {
		t.Errorf("TurnStarted count: got %d, want %d", got, want)
	}
	if got, want := countKind(kinds, "GoalLockApplied"), 2; got != want {
		t.Errorf("GoalLockApplied count: got %d, want %d", got, want)
	}
	if got, want := countKind(kinds, "BookkeepingApplied"), 1; got != want {
		t.Errorf("BookkeepingApplied count: got %d, want %d", got, want)
	}
	if got, want := countKind(kinds, "ActionSelected"), 1; got != want {
		t.Errorf("ActionSelected count: got %d, want %d", got, want)
	}
	if got, want := countKind(kinds, "TurnCompleted"), 2; got != want {
		t.Errorf("TurnCompleted count: got %d, want %d", got, want)
	}
	if got, want := countKind(kinds, "CycleCompleted"), 1; got != want {
		t.Errorf("CycleCompleted count: got %d, want %d", got, want)
	}

	var alphaLock testharness.GoalLockPayload
	for _, ev := range h.observer.Events {
		if ev.Kind == "GoalLockApplied" && ev.Faction != nil && ev.Faction.ID == "alpha" {
			alphaLock = ev.Payload.(testharness.GoalLockPayload)
		}
	}
	if alphaLock.Lock.Type != goal.LockSkip {
		t.Errorf("alpha lock type: got %v, want LockSkip", alphaLock.Lock.Type)
	}
	if len(alphaLock.Mutations) != 1 {
		t.Fatalf("alpha lock mutations: got %d, want 1 (GoalTurnsTick)", len(alphaLock.Mutations))
	}

	records := readHistory(t, h.cfg.HistoryPath)
	rec, ok := findMutationByCause(records, "change_homeworld_transit")
	if !ok {
		t.Fatalf("history missing change_homeworld_transit mutation; records=%+v", records)
	}
	if rec.Type != "goal_turns_tick" {
		t.Errorf("change_homeworld_transit mutation type: got %s, want goal_turns_tick", rec.Type)
	}

	if got := h.factionState.Factions["alpha"].ActiveGoal.TurnsRemaining; got != 1 {
		t.Errorf("alpha TurnsRemaining: got %d, want 1", got)
	}
}

// --- scenario 3: LockRestrictActions — Planetary Seizure phase 1 ---

func TestRunCycle_LockRestrictActions_PlanetarySeizurePhase1(t *testing.T) {
	h := newHarness(t)
	attacker := h.addFaction("alpha", "Tartarus", 4, 3, 2)
	target := h.addFaction("beta", "Tartarus", 2, 2, 2)
	attacker.ActiveGoal = &domain.ActiveGoal{
		GoalID:          "G-004",
		ProcessPhase:    1,
		TargetFactionID: target.ID,
		TargetWorld:     "Tartarus",
		TurnsRemaining:  3,
	}

	h.engine.Rand = &testharness.FixedRoller{Values: []int{10, 1, 3}}

	var alphaAvailable []string
	h.collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID == "alpha" {
			alphaAvailable = nil
			for _, a := range available {
				alphaAvailable = append(alphaAvailable, a.Name())
			}
			for _, a := range available {
				if a.Name() == "Attack" {
					return a, nil
				}
			}
			t.Fatalf("Attack not in alpha's available actions: %v", alphaAvailable)
		}
		return nil, nil
	}
	h.collector.SelectAttackersFn = func(eligible []*domain.Asset, _ *rulebook.Rulebook) ([]*domain.Asset, error) {
		return eligible, nil
	}
	h.collector.SelectDefenderFn = func(_ *domain.Asset, eligible []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
		return eligible[0], nil
	}
	h.collector.ConfirmRedirectToBaseFn = func(_ *domain.Faction, _ *domain.Base, _ int) (bool, error) {
		return false, nil
	}

	if err := h.engine.Turn.Start(h.factionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.engine.RunCycle(h.factionState, h.cfg, h.collector, h.observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	if len(alphaAvailable) != 1 || alphaAvailable[0] != "Attack" {
		t.Fatalf("alpha available actions under LockRestrictActions: got %v, want [Attack]", alphaAvailable)
	}

	resolvedCount := 0
	for _, ev := range h.observer.Events {
		if ev.Kind == "ActionResolved" {
			resolvedCount++
		}
	}
	if resolvedCount != 1 {
		t.Errorf("ActionResolved count: got %d, want 1", resolvedCount)
	}

	records := readHistory(t, h.cfg.HistoryPath)
	if _, ok := findMutationByCause(records, "attack"); !ok {
		t.Fatalf("history missing attack mutation; records=%+v", records)
	}

	if got := len(h.factionState.Factions["beta"].Assets); got != 0 {
		t.Errorf("beta assets after attack: got %d, want 0", got)
	}
}

// --- scenario 4: Buy Asset — new asset not ready until next cycle ---

func TestRunCycle_BuyAsset_ReadyNextCycle(t *testing.T) {
	h := newHarness(t)
	alpha := h.addFaction("alpha", "Tartarus", 4, 3, 2)

	// income after bookkeeping: wealth(2)/2=1 + (force(4)+cunning(3))/4=1 → 2 Coin, enough for F1-001 (cost 2).
	secondCycle := false
	h.collector.SelectActionFn = func(_ *domain.Faction, available []action.Action) (action.Action, error) {
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
	h.collector.SelectBuyOrderFn = func(worlds []string, purchasable []*domain.AssetDefinition) (action.BuyOrder, error) {
		for _, def := range purchasable {
			if def.ID == defSecurityPersonnel {
				return action.BuyOrder{World: worlds[0], Definition: def}, nil
			}
		}
		t.Fatal("F1-001 not in purchasable list")
		return action.BuyOrder{}, nil
	}

	if err := h.engine.Turn.Start(h.factionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.engine.RunCycle(h.factionState, h.cfg, h.collector, h.observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	records := readHistory(t, h.cfg.HistoryPath)

	_, hasAdded := findMutationType(records, "asset_added")
	checkStep(t, "asset_added mutation in history", hasAdded, "no asset_added found")

	_, hasCoin := findMutationByTypeAndCause(records, "coin_delta", "buy")
	checkStep(t, "coin_delta(cause=buy) in history", hasCoin, "no buy coin_delta found")

	var newAsset *domain.Asset
	for _, asset := range alpha.Assets {
		if asset.DefinitionID == defSecurityPersonnel && asset.ID != "alpha-asset-1" {
			newAsset = asset
			break
		}
	}
	checkStep(t, "new asset present in faction", newAsset != nil, "no second F1-001 asset found")
	if newAsset != nil {
		checkStep(t, "new asset Ready=false after purchase", !newAsset.Ready, "Ready should be false until next cycle")
	}

	// Second cycle: Turn.Start readies all assets.
	secondCycle = true
	if err := h.engine.Turn.Start(h.factionState); err != nil {
		t.Fatalf("second Turn.Start: %v", err)
	}
	if err := h.engine.RunCycle(h.factionState, h.cfg, h.collector, h.observer); err != nil {
		t.Fatalf("second RunCycle: %v", err)
	}

	if newAsset != nil {
		checkStep(t, "new asset Ready=true after second cycle start", newAsset.Ready, "Ready should flip true at Turn.Start")
	}
}

// --- scenario 5: Expand Influence — new base placed uncontested ---

func TestRunCycle_ExpandInfluence_NewBase(t *testing.T) {
	h := newHarness(t)
	alpha := h.addFaction("alpha", "Tartarus", 4, 3, 2)

	h.collector.SelectActionFn = func(_ *domain.Faction, available []action.Action) (action.Action, error) {
		for _, a := range available {
			if a.Name() == "Expand Influence" {
				return a, nil
			}
		}
		t.Fatal("Expand Influence not available")
		return nil, nil
	}
	h.collector.SelectExpandInfluenceOrderFn = func(_ *domain.Faction, _ *state.FactionState) (action.ExpandInfluenceOrder, error) {
		return action.ExpandInfluenceOrder{Mode: action.ExpandModeNew, World: "Tartarus", HPAmount: 1}, nil
	}

	if err := h.engine.Turn.Start(h.factionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.engine.RunCycle(h.factionState, h.cfg, h.collector, h.observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	records := readHistory(t, h.cfg.HistoryPath)

	_, hasBase := findMutationByTypeAndCause(records, "base_added", "expand")
	checkStep(t, "base_added(cause=expand) in history", hasBase, "no base_added found")

	_, hasCoin := findMutationByTypeAndCause(records, "coin_delta", "expand")
	checkStep(t, "coin_delta(cause=expand) in history", hasCoin, "no expand coin_delta found")

	hasBaseTartarus := false
	for _, base := range alpha.Bases {
		if base.Location == "Tartarus" {
			hasBaseTartarus = true
			break
		}
	}
	checkStep(t, "Tartarus base in faction state", hasBaseTartarus, "no Tartarus base found in alpha.Bases")
}

// --- scenario 6: Expand Influence — contested new base ---

func TestRunCycle_ExpandInfluence_Contested(t *testing.T) {
	h := newHarness(t)
	alpha := h.addFaction("alpha", "Tartarus", 4, 3, 2)
	addAssetOnWorld(alpha, "Krylos")        // gives alpha an asset on Krylos → eligible for new base there
	h.addFaction("beta", "Krylos", 2, 2, 2) // beta's asset is on Krylos → will contest

	// Roll sequence: [expansion, rival, attack, defense, damage]
	// expansion: 1 → factionRoll = 1+cunning(3) = 4
	// rival:    10 → rivalRoll  = 10+cunning(2) = 12 ≥ 4 → attacks
	// attack:   10 → attackRoll = 10+force(2)   = 12
	// defense:   1 → defRoll    = 1+cunning(3)  = 4  → 12≥4: hit, damage 1d3+1
	// damage:    3 → 3+1 = 4 damage → base (HP 1) destroyed
	h.engine.Rand = &testharness.FixedRoller{Values: []int{1, 10, 10, 1, 3}}

	h.collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
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
	h.collector.SelectExpandInfluenceOrderFn = func(_ *domain.Faction, _ *state.FactionState) (action.ExpandInfluenceOrder, error) {
		return action.ExpandInfluenceOrder{Mode: action.ExpandModeNew, World: "Krylos", HPAmount: 1}, nil
	}
	h.collector.ConfirmRivalFreeAttackFn = func(_ *domain.Faction, _, _ int) (bool, error) {
		return true, nil
	}
	h.collector.SelectBaseAttackersFn = func(_ *domain.Faction, eligible []*domain.Asset, _ *rulebook.Rulebook) ([]*domain.Asset, error) {
		return eligible, nil
	}

	if err := h.engine.Turn.Start(h.factionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.engine.RunCycle(h.factionState, h.cfg, h.collector, h.observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	records := readHistory(t, h.cfg.HistoryPath)

	_, hasBase := findMutationByTypeAndCause(records, "base_added", "expand")
	checkStep(t, "base_added(cause=expand) in history", hasBase, "no base_added found")

	_, hasHPDelta := findMutationByTypeAndCause(records, "base_hp_delta", "expand")
	checkStep(t, "base_hp_delta(cause=expand) in history — rival attacked new base", hasHPDelta, "no base_hp_delta found")

	_, hasDestroyed := findMutationByTypeAndCause(records, "base_destroyed", "expand")
	checkStep(t, "base_destroyed(cause=expand) in history — base HP driven to 0", hasDestroyed, "no base_destroyed found")
}

// --- scenario 7: Goal completed mid-cycle via MilitaryConquest ---

func TestRunCycle_GoalCompleted_MilitaryConquest(t *testing.T) {
	h := newHarness(t)
	alpha := h.addFaction("alpha", "Tartarus", 1, 3, 2)
	h.addFaction("beta", "Tartarus", 2, 2, 2)

	// Progress=1, Force=1 → one Force kill pushes newProgress=2 ≥ Force=1 → complete (xp=2/2=1).
	alpha.ActiveGoal = &domain.ActiveGoal{GoalID: "G-001", Progress: 1}

	// Roll sequence: [attack, defense, damage]
	// attack:  10 → attackRoll = 10+force(1) = 11
	// defense:  1 → defRoll   = 1+force(2)  = 3  → 11≥3: hit
	// damage:   3 → 3+1 = 4 ≥ asset HP(3)        → destroyed
	h.engine.Rand = &testharness.FixedRoller{Values: []int{10, 1, 3}}

	h.collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
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
	h.collector.SelectAttackersFn = func(eligible []*domain.Asset, _ *rulebook.Rulebook) ([]*domain.Asset, error) {
		return eligible, nil
	}
	h.collector.SelectDefenderFn = func(_ *domain.Asset, eligible []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
		return eligible[0], nil
	}

	if err := h.engine.Turn.Start(h.factionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.engine.RunCycle(h.factionState, h.cfg, h.collector, h.observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	records := readHistory(t, h.cfg.HistoryPath)

	_, hasCompleted := findMutationByCause(records, "goal_completed")
	checkStep(t, "goal_completed mutation in history", hasCompleted, "no goal_completed mutation found")

	_, hasXP := findMutationType(records, "xp_awarded")
	checkStep(t, "xp_awarded mutation in history", hasXP, "no xp_awarded mutation found")

	checkStep(t, "alpha.ActiveGoal cleared after completion", alpha.ActiveGoal == nil, "ActiveGoal should be nil")
}

// --- scenario 8: Bribe — rival base takes influence, attacker spends coin ---

func TestRunCycle_Bribe(t *testing.T) {
	h := newHarness(t)
	alpha := h.addFaction("alpha", "Tartarus", 4, 3, 2)
	beta := h.addFaction("beta", "Tartarus", 2, 2, 2)

	// Bribe.Validate requires len(faction.Bases) > 0 on the acting faction.
	addBase(alpha, "Tartarus", 3)
	rivalBase := addBase(beta, "Tartarus", 3)

	h.collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
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
	h.collector.SelectBribeTargetFn = func(_ *domain.Faction, _ *state.FactionState) (*domain.Base, int, error) {
		return rivalBase, 1, nil
	}

	if err := h.engine.Turn.Start(h.factionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.engine.RunCycle(h.factionState, h.cfg, h.collector, h.observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	records := readHistory(t, h.cfg.HistoryPath)

	_, hasInfluence := findMutationByTypeAndCause(records, "influence_delta", "bribe")
	checkStep(t, "influence_delta(cause=bribe) in history", hasInfluence, "no bribe influence_delta found")

	_, hasCoin := findMutationByTypeAndCause(records, "coin_delta", "bribe")
	checkStep(t, "coin_delta(cause=bribe) in history", hasCoin, "no bribe coin_delta found")
}

// --- scenario 9: MutationReactor dispatch — CoinDelta on asset destroyed ---

// coinOnDestroyReactor emits one CoinDelta the first time it observes an
// AssetRemoved in the mutation slice. Stateful so it does not double-fire on
// recursive dispatch rounds.
type coinOnDestroyReactor struct {
	factionID string
	fired     bool
}

func (r *coinOnDestroyReactor) OnMutations(mutations []domain.Mutation, _ *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	if r.fired {
		return nil
	}
	for _, m := range mutations {
		if _, ok := m.(domain.AssetRemoved); ok {
			r.fired = true
			return []domain.Mutation{domain.CoinDelta{FactionID: r.factionID, Delta: 1}}
		}
	}
	return nil
}

func TestMutationReactorDispatch_CoinOnAssetDestroyed(t *testing.T) {
	h := newHarness(t)
	alpha := h.addFaction("alpha", "Tartarus", 4, 3, 2)
	h.addFaction("beta", "Tartarus", 2, 2, 2)

	// Roll sequence: attack hits and destroys beta's asset.
	// attack=10+force(4)=14, defense=1+force(2)=3 → hit; damage=3+1=4 ≥ HP(3) → destroyed.
	h.engine.Rand = &testharness.FixedRoller{Values: []int{10, 1, 3}}

	stub := &coinOnDestroyReactor{factionID: "alpha"}
	h.engine.Hooks.RegisterMutationReactor(hooks.FactionScope("alpha"), "stub-coin-on-destroy", stub)

	h.collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
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
	h.collector.SelectAttackersFn = func(eligible []*domain.Asset, _ *rulebook.Rulebook) ([]*domain.Asset, error) {
		return eligible, nil
	}
	h.collector.SelectDefenderFn = func(_ *domain.Asset, eligible []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
		return eligible[0], nil
	}

	initialCoin := alpha.Coin

	if err := h.engine.Turn.Start(h.factionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.engine.RunCycle(h.factionState, h.cfg, h.collector, h.observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	checkStep(t, "beta asset destroyed", len(h.factionState.Factions["beta"].Assets) == 0,
		"beta should have no assets after attack")
	checkStep(t, "reactor fired on AssetRemoved", stub.fired,
		"reactor never saw AssetRemoved in mutation slice")

	// alpha income: wealth(2)/2=1 + (force(4)+cunning(3))/4=1 = 2; reactor: +1 → total +3
	wantCoin := initialCoin + 3
	checkStep(t, "alpha Coin reflects reactor delta (+1 above bookkeeping income)", alpha.Coin == wantCoin,
		"alpha.Coin mismatch — reactor CoinDelta may not have been applied")

	records := readHistory(t, h.cfg.HistoryPath)
	_, hasRemoved := findMutationType(records, "asset_removed")
	checkStep(t, "asset_removed in history", hasRemoved, "no asset_removed mutation found")
	// reactor's CoinDelta has no Cause; bookkeeping uses Cause="bookkeeping"
	_, hasReactorCoin := findMutationByTypeAndCause(records, "coin_delta", "")
	checkStep(t, "reactor coin_delta (cause='') in history", hasReactorCoin, "no reactor coin_delta in history")
}

// --- scenario N: Scavengers tag — +1 Coin per asset destroyed ---

func TestScavengers_GrantsCoinOnKill(t *testing.T) {
	h := newHarness(t)
	alpha := h.addFaction("alpha", "Tartarus", 4, 3, 2)
	alpha.Tags = []*domain.Tag{{ID: "T-014"}}
	h.addFaction("beta", "Tartarus", 2, 2, 2)
	h.registerTags()

	h.engine.Rand = &testharness.FixedRoller{Values: []int{10, 1, 3}}
	h.collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
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
	h.collector.SelectAttackersFn = func(eligible []*domain.Asset, _ *rulebook.Rulebook) ([]*domain.Asset, error) {
		return eligible, nil
	}
	h.collector.SelectDefenderFn = func(_ *domain.Asset, eligible []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
		return eligible[0], nil
	}

	if err := h.engine.Turn.Start(h.factionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.engine.RunCycle(h.factionState, h.cfg, h.collector, h.observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	if got := len(h.factionState.Factions["beta"].Assets); got != 0 {
		t.Fatalf("beta assets: got %d, want 0 (attack should have destroyed it)", got)
	}

	records := readHistory(t, h.cfg.HistoryPath)
	if _, ok := findMutationByTypeAndCause(records, "coin_delta", "scavengers"); !ok {
		t.Error("expected coin_delta with cause=scavengers in history")
	}
}

func TestScavengers_NoBonusWithoutTag(t *testing.T) {
	h := newHarness(t)
	h.addFaction("alpha", "Tartarus", 4, 3, 2) // no Scavengers tag
	h.addFaction("beta", "Tartarus", 2, 2, 2)
	h.registerTags()

	h.engine.Rand = &testharness.FixedRoller{Values: []int{10, 1, 3}}
	h.collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
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
	h.collector.SelectAttackersFn = func(eligible []*domain.Asset, _ *rulebook.Rulebook) ([]*domain.Asset, error) {
		return eligible, nil
	}
	h.collector.SelectDefenderFn = func(_ *domain.Asset, eligible []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
		return eligible[0], nil
	}

	if err := h.engine.Turn.Start(h.factionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.engine.RunCycle(h.factionState, h.cfg, h.collector, h.observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	records := readHistory(t, h.cfg.HistoryPath)
	if _, ok := findMutationByTypeAndCause(records, "coin_delta", "scavengers"); ok {
		t.Error("expected no coin_delta with cause=scavengers when Scavengers tag is absent")
	}
}

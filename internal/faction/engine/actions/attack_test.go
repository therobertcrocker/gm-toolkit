package actions

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// fixedRoller returns values from a pre-set sequence, cycling if exhausted.
type fixedRoller struct {
	values []int
	index  int
}

func (roller *fixedRoller) Roll(_ int) int {
	val := roller.values[roller.index%len(roller.values)]
	roller.index++
	return val
}

// fakeCollector pre-configures responses for Attack input methods.
// Unused InputCollector methods panic to surface accidental calls in tests.
type fakeCollector struct {
	attackers   []*domain.Asset
	defenderSeq []*domain.Asset
	defIdx      int
	redirectSeq []bool
	redIdx      int
}

func (collector *fakeCollector) SelectAttackers(_ []*domain.Asset, _ *loader.Rulebook) ([]*domain.Asset, error) {
	return collector.attackers, nil
}
func (collector *fakeCollector) SelectDefender(_ *domain.Asset, _ []*domain.Asset, _ *loader.Rulebook) (*domain.Asset, error) {
	defender := collector.defenderSeq[collector.defIdx]
	collector.defIdx++
	return defender, nil
}
func (collector *fakeCollector) ConfirmRedirectToBase(_ *domain.Faction, _ *domain.Base, _ int) (bool, error) {
	redirect := collector.redirectSeq[collector.redIdx]
	collector.redIdx++
	return redirect, nil
}
func (collector *fakeCollector) SelectAsset(_ []*domain.Asset, _ *loader.Rulebook) (*domain.Asset, error) {
	panic("SelectAsset: not used in attack tests")
}
func (collector *fakeCollector) SelectRepairOrders(_ *domain.Faction, _ []*domain.Asset, _ *loader.Rulebook) ([]engine.RepairOrder, error) {
	panic("SelectRepairOrders: not used in attack tests")
}
func (collector *fakeCollector) SelectBuyOrder(_ []string, _ []*domain.AssetDefinition) (engine.BuyOrder, error) {
	panic("SelectBuyOrder: not used in attack tests")
}
func (collector *fakeCollector) SelectRefitOrder(_ []engine.RefitOption, _ *loader.Rulebook) (engine.RefitOrder, error) {
	panic("SelectRefitOrder: not used in attack tests")
}
func (collector *fakeCollector) SelectExpandInfluenceOrder(_ *domain.Faction, _ *state.FactionState) (engine.ExpandInfluenceOrder, error) {
	panic("SelectExpandInfluenceOrder: not used in attack tests")
}
func (collector *fakeCollector) ConfirmRivalFreeAttack(_ *domain.Faction, _, _ int) (bool, error) {
	panic("ConfirmRivalFreeAttack: not used in attack tests")
}
func (collector *fakeCollector) SelectBaseAttackers(_ *domain.Faction, _ []*domain.Asset, _ *loader.Rulebook) ([]*domain.Asset, error) {
	panic("SelectBaseAttackers: not used in attack tests")
}
func (collector *fakeCollector) SelectAbilityAssets(_ *domain.Faction, _ []*domain.Asset, _ *loader.Rulebook) ([]*domain.Asset, error) {
	panic("SelectAbilityAssets: not used in attack tests")
}
func (collector *fakeCollector) SelectMoveDestination(_ *domain.Asset, _ []string) (string, error) {
	panic("SelectMoveDestination: not used in attack tests")
}
func (collector *fakeCollector) SelectFactionTestTarget(_ *domain.Asset, _ domain.AbilityEffectType, _ []*domain.Faction) (*domain.Faction, error) {
	panic("SelectFactionTestTarget: not used in attack tests")
}
func (collector *fakeCollector) ConfirmAbilityApplied(_ *domain.Asset, _ *domain.AssetDefinition) (bool, error) {
	panic("ConfirmAbilityApplied: not used in attack tests")
}

func (collector *fakeCollector) SelectBribeTarget(_ *domain.Faction, _ *state.FactionState) (*domain.Base, int, error) {
	panic("SelectBribeTarget: not used in attack tests")
}

func (collector *fakeCollector) SelectSeizeTarget(_ *domain.Faction, _ *state.FactionState) (string, error) {
	panic("SelectSeizeTarget: not used in attack tests")
}

func (collector *fakeCollector) SelectAction(_ *domain.Faction, _ []engine.Action) (engine.Action, error) {
	panic("SelectAction: not used in attack tests")
}

func (collector *fakeCollector) AwaitCheckpoint(_ string) error { return nil }

// makeAttackRulebook returns a minimal Rulebook with three asset definitions:
//   - "force-attacker": has an Attack profile (Force vs Force, 1d6 damage)
//   - "force-defender": has a Counter (1d4) but no Attack profile
//   - "force-defender-nocounter": no Counter, no Attack profile
func makeAttackRulebook() *loader.Rulebook {
	return &loader.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			"force-attacker": {
				ID:       "force-attacker",
				Name:     "Infantry",
				Category: domain.StatForce,
				HP:       8,
				Attack: &domain.AttackProfile{
					AttackerStat: domain.StatForce,
					DefenderStat: domain.StatForce,
					Damage:       domain.DiceRoll{NumDice: 1, Sides: 6},
				},
			},
			"force-defender": {
				ID:      "force-defender",
				Name:    "Militia",
				HP:      8,
				Counter: &domain.DiceRoll{NumDice: 1, Sides: 4},
			},
			"force-defender-nocounter": {
				ID:   "force-defender-nocounter",
				Name: "Conscripts",
				HP:   8,
			},
		},
	}
}

// makeAttackState builds a minimal FactionState with one attacker asset on f1
// and one defender asset on f2, both on world "Anchorage". Stats are zeroed so
// roll outcomes are determined solely by the fixedRoller values.
func makeAttackState(attackerHP, defenderHP int, defenderDefID string, includeBase bool) (*state.FactionState, *domain.Asset, *domain.Asset) {
	attackerAsset := &domain.Asset{
		ID:           "a1",
		DefinitionID: "force-attacker",
		OwnerID:      "f1",
		Location:     "Anchorage",
		CurrentHP:    attackerHP,
		Ready:        true,
		Maintained:   true,
	}
	defenderAsset := &domain.Asset{
		ID:           "d1",
		DefinitionID: defenderDefID,
		OwnerID:      "f2",
		Location:     "Anchorage",
		CurrentHP:    defenderHP,
		Ready:        true,
		Maintained:   true,
	}
	f2 := &domain.Faction{
		ID:     "f2",
		Assets: []*domain.Asset{defenderAsset},
	}
	if includeBase {
		f2.Bases = []*domain.Base{{
			ID:        "f2-base",
			OwnerID:   "f2",
			Location:  "Anchorage",
			CurrentHP: 10,
			MaxHP:     10,
		}}
	}
	factionState := &state.FactionState{
		Factions: map[string]*domain.Faction{
			"f1": {ID: "f1", Assets: []*domain.Asset{attackerAsset}},
			"f2": f2,
		},
	}
	return factionState, attackerAsset, defenderAsset
}

// runAttack is a test helper that drives a full Validate→Inputs→Resolve→Output
// cycle and returns the resulting mutation list.
func runAttack(t *testing.T, collector *fakeCollector, roller *fixedRoller, faction *domain.Faction, factionState *state.FactionState, rulebook *loader.Rulebook) []domain.Mutation {
	t.Helper()
	attack := NewAttack(collector, roller)
	if err := attack.Inputs(faction, factionState, rulebook); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := attack.Resolve(faction, factionState, rulebook); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := attack.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}
	return mutations
}

func TestAttack_Validate(t *testing.T) {
	rulebook := makeAttackRulebook()

	t.Run("no eligible attackers", func(t *testing.T) {
		factionState, attackerAsset, _ := makeAttackState(8, 8, "force-defender", false)
		attackerAsset.Ready = false // on cooldown — ineligible
		faction := factionState.Factions["f1"]
		attack := NewAttack(&fakeCollector{}, &fixedRoller{values: []int{1}})
		if attack.Validate(faction, factionState, rulebook) {
			t.Error("expected Validate false when attacker is not Ready")
		}
	})

	t.Run("no rival assets on world", func(t *testing.T) {
		factionState, _, defenderAsset := makeAttackState(8, 8, "force-defender", false)
		defenderAsset.Location = "Tartarus" // rival is elsewhere
		faction := factionState.Factions["f1"]
		attack := NewAttack(&fakeCollector{}, &fixedRoller{values: []int{1}})
		if attack.Validate(faction, factionState, rulebook) {
			t.Error("expected Validate false when no rivals on attacker's world")
		}
	})

	t.Run("eligible attacker with target", func(t *testing.T) {
		factionState, _, _ := makeAttackState(8, 8, "force-defender", false)
		faction := factionState.Factions["f1"]
		attack := NewAttack(&fakeCollector{}, &fixedRoller{values: []int{1}})
		if !attack.Validate(faction, factionState, rulebook) {
			t.Error("expected Validate true")
		}
	})
}

// TestAttack_AttackerWins_NonLethal: attacker roll > defense roll; damage < defender HP.
// Rolls: attack=8, defense=3, damage=4.
func TestAttack_AttackerWins_NonLethal(t *testing.T) {
	rulebook := makeAttackRulebook()
	factionState, attackerAsset, defenderAsset := makeAttackState(8, 10, "force-defender-nocounter", false)
	faction := factionState.Factions["f1"]

	mutations := runAttack(t,
		&fakeCollector{attackers: []*domain.Asset{attackerAsset}, defenderSeq: []*domain.Asset{defenderAsset}},
		&fixedRoller{values: []int{8, 3, 4}},
		faction, factionState, rulebook,
	)

	if len(mutations) != 1 {
		t.Fatalf("len(mutations) = %d, want 1; got %v", len(mutations), mutations)
	}
	delta, ok := mutations[0].(domain.AssetHPDelta)
	if !ok {
		t.Fatalf("mutations[0] type = %T, want AssetHPDelta", mutations[0])
	}
	if delta.AssetID != "d1" || delta.Delta != -4 {
		t.Errorf("AssetHPDelta = {%s, %d}, want {d1, -4}", delta.AssetID, delta.Delta)
	}
}

// TestAttack_AttackerWins_Lethal: damage exceeds defender HP; AssetRemoved emitted inline.
// Rolls: attack=8, defense=3, damage=6.
func TestAttack_AttackerWins_Lethal(t *testing.T) {
	rulebook := makeAttackRulebook()
	factionState, attackerAsset, defenderAsset := makeAttackState(8, 5, "force-defender-nocounter", false)
	faction := factionState.Factions["f1"]

	mutations := runAttack(t,
		&fakeCollector{attackers: []*domain.Asset{attackerAsset}, defenderSeq: []*domain.Asset{defenderAsset}},
		&fixedRoller{values: []int{8, 3, 6}},
		faction, factionState, rulebook,
	)

	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2; got %v", len(mutations), mutations)
	}
	if _, ok := mutations[0].(domain.AssetHPDelta); !ok {
		t.Errorf("mutations[0] type = %T, want AssetHPDelta", mutations[0])
	}
	removed, ok := mutations[1].(domain.AssetRemoved)
	if !ok {
		t.Fatalf("mutations[1] type = %T, want AssetRemoved", mutations[1])
	}
	if removed.AssetID != "d1" {
		t.Errorf("AssetRemoved.AssetID = %q, want %q", removed.AssetID, "d1")
	}
}

// TestAttack_DefenderWins_NoCounter: defense roll > attack roll; defender has no counter.
// Rolls: attack=2, defense=9.
func TestAttack_DefenderWins_NoCounter(t *testing.T) {
	rulebook := makeAttackRulebook()
	factionState, attackerAsset, defenderAsset := makeAttackState(8, 8, "force-defender-nocounter", false)
	faction := factionState.Factions["f1"]

	mutations := runAttack(t,
		&fakeCollector{attackers: []*domain.Asset{attackerAsset}, defenderSeq: []*domain.Asset{defenderAsset}},
		&fixedRoller{values: []int{2, 9}},
		faction, factionState, rulebook,
	)

	if len(mutations) != 0 {
		t.Errorf("expected no mutations when defender wins with no counter, got %v", mutations)
	}
}

// TestAttack_DefenderWins_WithCounter: defense roll > attack roll; counterattack damages attacker.
// Rolls: attack=2, defense=9, counter=3.
func TestAttack_DefenderWins_WithCounter(t *testing.T) {
	rulebook := makeAttackRulebook()
	factionState, attackerAsset, defenderAsset := makeAttackState(8, 8, "force-defender", false)
	faction := factionState.Factions["f1"]

	mutations := runAttack(t,
		&fakeCollector{attackers: []*domain.Asset{attackerAsset}, defenderSeq: []*domain.Asset{defenderAsset}},
		&fixedRoller{values: []int{2, 9, 3}},
		faction, factionState, rulebook,
	)

	if len(mutations) != 1 {
		t.Fatalf("len(mutations) = %d, want 1; got %v", len(mutations), mutations)
	}
	delta, ok := mutations[0].(domain.AssetHPDelta)
	if !ok {
		t.Fatalf("mutations[0] type = %T, want AssetHPDelta", mutations[0])
	}
	if delta.AssetID != "a1" || delta.Delta != -3 {
		t.Errorf("AssetHPDelta = {%s, %d}, want {a1, -3}", delta.AssetID, delta.Delta)
	}
}

// TestAttack_Tie: equal rolls; both attack damage and counterattack apply.
// Rolls: attack=5, defense=5, damage=4, counter=2.
func TestAttack_Tie(t *testing.T) {
	rulebook := makeAttackRulebook()
	factionState, attackerAsset, defenderAsset := makeAttackState(8, 8, "force-defender", false)
	faction := factionState.Factions["f1"]

	mutations := runAttack(t,
		&fakeCollector{attackers: []*domain.Asset{attackerAsset}, defenderSeq: []*domain.Asset{defenderAsset}},
		&fixedRoller{values: []int{5, 5, 4, 2}},
		faction, factionState, rulebook,
	)

	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2; got %v", len(mutations), mutations)
	}
	defDelta, ok := mutations[0].(domain.AssetHPDelta)
	if !ok || defDelta.AssetID != "d1" || defDelta.Delta != -4 {
		t.Errorf("mutations[0] = %v, want AssetHPDelta{d1, -4}", mutations[0])
	}
	atkDelta, ok := mutations[1].(domain.AssetHPDelta)
	if !ok || atkDelta.AssetID != "a1" || atkDelta.Delta != -2 {
		t.Errorf("mutations[1] = %v, want AssetHPDelta{a1, -2}", mutations[1])
	}
}

// TestAttack_RedirectToBase_Accepted: defender faction has a Base on the world;
// GM accepts redirect. Both BaseHPDelta and FactionHPDelta are emitted.
// Rolls: attack=8, defense=3, damage=4.
func TestAttack_RedirectToBase_Accepted(t *testing.T) {
	rulebook := makeAttackRulebook()
	factionState, attackerAsset, defenderAsset := makeAttackState(8, 8, "force-defender-nocounter", true)
	faction := factionState.Factions["f1"]

	mutations := runAttack(t,
		&fakeCollector{
			attackers:   []*domain.Asset{attackerAsset},
			defenderSeq: []*domain.Asset{defenderAsset},
			redirectSeq: []bool{true},
		},
		&fixedRoller{values: []int{8, 3, 4}},
		faction, factionState, rulebook,
	)

	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2; got %v", len(mutations), mutations)
	}
	baseDelta, ok := mutations[0].(domain.BaseHPDelta)
	if !ok || baseDelta.BaseID != "f2-base" || baseDelta.Delta != -4 {
		t.Errorf("mutations[0] = %v, want BaseHPDelta{f2-base, -4}", mutations[0])
	}
	factionDelta, ok := mutations[1].(domain.FactionHPDelta)
	if !ok || factionDelta.FactionID != "f2" || factionDelta.Delta != -4 {
		t.Errorf("mutations[1] = %v, want FactionHPDelta{f2, -4}", mutations[1])
	}
}

// TestAttack_RedirectToBase_Lethal: redirect damage exceeds Base HP; BaseDestroyed emitted.
// Rolls: attack=8, defense=3, damage=5. Base HP=3.
func TestAttack_RedirectToBase_Lethal(t *testing.T) {
	rulebook := makeAttackRulebook()
	factionState, attackerAsset, defenderAsset := makeAttackState(8, 8, "force-defender-nocounter", true)
	factionState.Factions["f2"].Bases[0].CurrentHP = 3
	faction := factionState.Factions["f1"]

	mutations := runAttack(t,
		&fakeCollector{
			attackers:   []*domain.Asset{attackerAsset},
			defenderSeq: []*domain.Asset{defenderAsset},
			redirectSeq: []bool{true},
		},
		&fixedRoller{values: []int{8, 3, 5}},
		faction, factionState, rulebook,
	)

	if len(mutations) != 3 {
		t.Fatalf("len(mutations) = %d, want 3; got %v", len(mutations), mutations)
	}
	if _, ok := mutations[0].(domain.BaseHPDelta); !ok {
		t.Errorf("mutations[0] type = %T, want BaseHPDelta", mutations[0])
	}
	if _, ok := mutations[1].(domain.FactionHPDelta); !ok {
		t.Errorf("mutations[1] type = %T, want FactionHPDelta", mutations[1])
	}
	destroyed, ok := mutations[2].(domain.BaseDestroyed)
	if !ok || destroyed.BaseID != "f2-base" {
		t.Errorf("mutations[2] = %v, want BaseDestroyed{f2-base}", mutations[2])
	}
}

// TestAttack_RedirectToBase_Declined: Base exists but redirect declined; asset takes damage.
// Rolls: attack=8, defense=3, damage=4.
func TestAttack_RedirectToBase_Declined(t *testing.T) {
	rulebook := makeAttackRulebook()
	factionState, attackerAsset, defenderAsset := makeAttackState(8, 8, "force-defender-nocounter", true)
	faction := factionState.Factions["f1"]

	mutations := runAttack(t,
		&fakeCollector{
			attackers:   []*domain.Asset{attackerAsset},
			defenderSeq: []*domain.Asset{defenderAsset},
			redirectSeq: []bool{false},
		},
		&fixedRoller{values: []int{8, 3, 4}},
		faction, factionState, rulebook,
	)

	if len(mutations) != 1 {
		t.Fatalf("len(mutations) = %d, want 1; got %v", len(mutations), mutations)
	}
	if _, ok := mutations[0].(domain.AssetHPDelta); !ok {
		t.Errorf("mutations[0] type = %T, want AssetHPDelta", mutations[0])
	}
}

// TestAttack_StealthCleared: a stealthy attacker has AssetStealthCleared emitted
// before any damage mutation. Stealthy defenders cannot be targeted (filtered by
// eligibleDefenders), so only attacker stealth is exercised here.
// Rolls: attack=8, defense=3, damage=4.
func TestAttack_StealthCleared(t *testing.T) {
	rulebook := makeAttackRulebook()
	factionState, attackerAsset, defenderAsset := makeAttackState(8, 8, "force-defender-nocounter", false)
	attackerAsset.Stealthy = true
	faction := factionState.Factions["f1"]

	mutations := runAttack(t,
		&fakeCollector{attackers: []*domain.Asset{attackerAsset}, defenderSeq: []*domain.Asset{defenderAsset}},
		&fixedRoller{values: []int{8, 3, 4}},
		faction, factionState, rulebook,
	)

	// Expect: StealthCleared(a1), AssetHPDelta(d1)
	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2; got %v", len(mutations), mutations)
	}
	cleared, ok := mutations[0].(domain.AssetStealthCleared)
	if !ok || cleared.AssetID != "a1" {
		t.Errorf("mutations[0] = %v, want AssetStealthCleared{a1}", mutations[0])
	}
	if _, ok := mutations[1].(domain.AssetHPDelta); !ok {
		t.Errorf("mutations[1] type = %T, want AssetHPDelta", mutations[1])
	}
}

// TestAttack_StealthCleared_OncePerAsset: a stealthy attacker queued twice
// (via fakeCollector, exercising the dedup path) has AssetStealthCleared
// emitted only once across both matchup slots.
// Rolls: [8,3,4] for matchup 1; matchup 2 re-uses same attacker slot.
func TestAttack_StealthCleared_OncePerAsset(t *testing.T) {
	rulebook := makeAttackRulebook()
	factionState, attackerAsset, defenderAsset := makeAttackState(8, 8, "force-defender-nocounter", false)
	attackerAsset.Stealthy = true
	faction := factionState.Factions["f1"]

	// Queue a1 twice to exercise the stealthCleared dedup map.
	mutations := runAttack(t,
		&fakeCollector{
			attackers:   []*domain.Asset{attackerAsset, attackerAsset},
			defenderSeq: []*domain.Asset{defenderAsset, defenderAsset},
		},
		&fixedRoller{values: []int{8, 3, 4}},
		faction, factionState, rulebook,
	)

	stealthCount := 0
	for _, mutation := range mutations {
		if cleared, ok := mutation.(domain.AssetStealthCleared); ok && cleared.AssetID == "a1" {
			stealthCount++
		}
	}
	if stealthCount != 1 {
		t.Errorf("AssetStealthCleared for a1 emitted %d time(s), want exactly 1", stealthCount)
	}
}

// TestAttack_DestroyedAttackerSkipped: a1 appears twice in the committed sequence;
// its first matchup results in a lethal counter, so its second slot is skipped.
// Rolls: attack=2, defense=9, counter=8 (lethal); then no rolls consumed for a1's second slot.
func TestAttack_DestroyedAttackerSkipped(t *testing.T) {
	rulebook := makeAttackRulebook()
	factionState, attackerAsset, defenderAsset := makeAttackState(4, 8, "force-defender", false)
	faction := factionState.Factions["f1"]

	// fakeCollector queues a1 twice to directly exercise the re-check path.
	mutations := runAttack(t,
		&fakeCollector{
			attackers:   []*domain.Asset{attackerAsset, attackerAsset},
			defenderSeq: []*domain.Asset{defenderAsset},
		},
		&fixedRoller{values: []int{2, 9, 8}},
		faction, factionState, rulebook,
	)

	// Matchup 1: defender wins, counter=8, attacker HP=4-8=-4 → AssetHPDelta + AssetRemoved.
	// Matchup 2: a1 re-check fails (tracker shows -4 net HP) → skipped, SelectDefender not called.
	removals := 0
	for _, mutation := range mutations {
		if _, ok := mutation.(domain.AssetRemoved); ok {
			removals++
		}
	}
	if removals != 1 {
		t.Errorf("AssetRemoved count = %d, want 1 (a1 destroyed once, second slot skipped)", removals)
	}
}

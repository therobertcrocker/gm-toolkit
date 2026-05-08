package actions

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"go.uber.org/mock/gomock"
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

// makeAttackRulebook returns a minimal Rulebook with three asset definitions:
//   - "force-attacker": has an Attack profile (Force vs Force, 1d6 damage)
//   - "force-defender": has a Counter (1d4) but no Attack profile
//   - "force-defender-nocounter": no Counter, no Attack profile
func makeAttackRulebook() *rulebook.Rulebook {
	return &rulebook.Rulebook{
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
		Assets: map[string]*domain.Asset{"d1": defenderAsset},
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
			"f1": {ID: "f1", Assets: map[string]*domain.Asset{"a1": attackerAsset}},
			"f2": f2,
		},
	}
	return factionState, attackerAsset, defenderAsset
}

// runAttack is a test helper that drives a full Validate→Inputs→Resolve→Output
// cycle and returns the resulting mutation list.
func runAttack(t *testing.T, collector action.Collector, roller *fixedRoller, faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook, registry *hooks.Registry) []domain.Mutation {
	t.Helper()
	attack := NewAttack(collector, roller, registry)
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
		ctrl := gomock.NewController(t)
		attack := NewAttack(mocks.NewMockInputCollector(ctrl), &fixedRoller{values: []int{1}}, nil)
		if attack.Validate(faction, factionState, rulebook) {
			t.Error("expected Validate false when attacker is not Ready")
		}
	})

	t.Run("no rival assets on world", func(t *testing.T) {
		factionState, _, defenderAsset := makeAttackState(8, 8, "force-defender", false)
		defenderAsset.Location = "Tartarus" // rival is elsewhere
		faction := factionState.Factions["f1"]
		ctrl := gomock.NewController(t)
		attack := NewAttack(mocks.NewMockInputCollector(ctrl), &fixedRoller{values: []int{1}}, nil)
		if attack.Validate(faction, factionState, rulebook) {
			t.Error("expected Validate false when no rivals on attacker's world")
		}
	})

	t.Run("eligible attacker with target", func(t *testing.T) {
		factionState, _, _ := makeAttackState(8, 8, "force-defender", false)
		faction := factionState.Factions["f1"]
		ctrl := gomock.NewController(t)
		attack := NewAttack(mocks.NewMockInputCollector(ctrl), &fixedRoller{values: []int{1}}, nil)
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

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockInputCollector(ctrl)
	collector.EXPECT().SelectAttackers(gomock.Any(), gomock.Any()).Return([]*domain.Asset{attackerAsset}, nil)
	collector.EXPECT().SelectDefender(gomock.Any(), gomock.Any(), gomock.Any()).Return(defenderAsset, nil)

	mutations := runAttack(t, collector, &fixedRoller{values: []int{8, 3, 4}}, faction, factionState, rulebook, nil)

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

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockInputCollector(ctrl)
	collector.EXPECT().SelectAttackers(gomock.Any(), gomock.Any()).Return([]*domain.Asset{attackerAsset}, nil)
	collector.EXPECT().SelectDefender(gomock.Any(), gomock.Any(), gomock.Any()).Return(defenderAsset, nil)

	mutations := runAttack(t, collector, &fixedRoller{values: []int{8, 3, 6}}, faction, factionState, rulebook, nil)

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

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockInputCollector(ctrl)
	collector.EXPECT().SelectAttackers(gomock.Any(), gomock.Any()).Return([]*domain.Asset{attackerAsset}, nil)
	collector.EXPECT().SelectDefender(gomock.Any(), gomock.Any(), gomock.Any()).Return(defenderAsset, nil)

	mutations := runAttack(t, collector, &fixedRoller{values: []int{2, 9}}, faction, factionState, rulebook, nil)

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

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockInputCollector(ctrl)
	collector.EXPECT().SelectAttackers(gomock.Any(), gomock.Any()).Return([]*domain.Asset{attackerAsset}, nil)
	collector.EXPECT().SelectDefender(gomock.Any(), gomock.Any(), gomock.Any()).Return(defenderAsset, nil)

	mutations := runAttack(t, collector, &fixedRoller{values: []int{2, 9, 3}}, faction, factionState, rulebook, nil)

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

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockInputCollector(ctrl)
	collector.EXPECT().SelectAttackers(gomock.Any(), gomock.Any()).Return([]*domain.Asset{attackerAsset}, nil)
	collector.EXPECT().SelectDefender(gomock.Any(), gomock.Any(), gomock.Any()).Return(defenderAsset, nil)

	mutations := runAttack(t, collector, &fixedRoller{values: []int{5, 5, 4, 2}}, faction, factionState, rulebook, nil)

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

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockInputCollector(ctrl)
	collector.EXPECT().SelectAttackers(gomock.Any(), gomock.Any()).Return([]*domain.Asset{attackerAsset}, nil)
	collector.EXPECT().SelectDefender(gomock.Any(), gomock.Any(), gomock.Any()).Return(defenderAsset, nil)
	collector.EXPECT().ConfirmRedirectToBase(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)

	mutations := runAttack(t, collector, &fixedRoller{values: []int{8, 3, 4}}, faction, factionState, rulebook, nil)

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

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockInputCollector(ctrl)
	collector.EXPECT().SelectAttackers(gomock.Any(), gomock.Any()).Return([]*domain.Asset{attackerAsset}, nil)
	collector.EXPECT().SelectDefender(gomock.Any(), gomock.Any(), gomock.Any()).Return(defenderAsset, nil)
	collector.EXPECT().ConfirmRedirectToBase(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)

	mutations := runAttack(t, collector, &fixedRoller{values: []int{8, 3, 5}}, faction, factionState, rulebook, nil)

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

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockInputCollector(ctrl)
	collector.EXPECT().SelectAttackers(gomock.Any(), gomock.Any()).Return([]*domain.Asset{attackerAsset}, nil)
	collector.EXPECT().SelectDefender(gomock.Any(), gomock.Any(), gomock.Any()).Return(defenderAsset, nil)
	collector.EXPECT().ConfirmRedirectToBase(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)

	mutations := runAttack(t, collector, &fixedRoller{values: []int{8, 3, 4}}, faction, factionState, rulebook, nil)

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

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockInputCollector(ctrl)
	collector.EXPECT().SelectAttackers(gomock.Any(), gomock.Any()).Return([]*domain.Asset{attackerAsset}, nil)
	collector.EXPECT().SelectDefender(gomock.Any(), gomock.Any(), gomock.Any()).Return(defenderAsset, nil)

	mutations := runAttack(t, collector, &fixedRoller{values: []int{8, 3, 4}}, faction, factionState, rulebook, nil)

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
// (exercising the dedup path) has AssetStealthCleared emitted only once across
// both matchup slots.
// Rolls: [8,3,4] for matchup 1; matchup 2 re-uses same attacker slot.
func TestAttack_StealthCleared_OncePerAsset(t *testing.T) {
	rulebook := makeAttackRulebook()
	factionState, attackerAsset, defenderAsset := makeAttackState(8, 8, "force-defender-nocounter", false)
	attackerAsset.Stealthy = true
	faction := factionState.Factions["f1"]

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockInputCollector(ctrl)
	// Queue a1 twice to exercise the stealthCleared dedup map.
	collector.EXPECT().SelectAttackers(gomock.Any(), gomock.Any()).Return([]*domain.Asset{attackerAsset, attackerAsset}, nil)
	collector.EXPECT().SelectDefender(gomock.Any(), gomock.Any(), gomock.Any()).Return(defenderAsset, nil).Times(2)

	mutations := runAttack(t, collector, &fixedRoller{values: []int{8, 3, 4}}, faction, factionState, rulebook, nil)

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

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockInputCollector(ctrl)
	// Queue a1 twice to directly exercise the re-check path.
	collector.EXPECT().SelectAttackers(gomock.Any(), gomock.Any()).Return([]*domain.Asset{attackerAsset, attackerAsset}, nil)
	// Only one SelectDefender call — second slot is skipped (attacker already destroyed).
	collector.EXPECT().SelectDefender(gomock.Any(), gomock.Any(), gomock.Any()).Return(defenderAsset, nil)

	mutations := runAttack(t, collector, &fixedRoller{values: []int{2, 9, 8}}, faction, factionState, rulebook, nil)

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

// stubTieResolverForAttack always returns a fixed TieOutcome.
type stubTieResolverForAttack struct{ outcome hooks.TieOutcome }

func (stub *stubTieResolverForAttack) ResolveTie(_ hooks.RollContext, _ *state.FactionState) hooks.TieOutcome {
	return stub.outcome
}

// TestAttack_TieResolver_DefenderWins: a TieDefenderWins resolver is registered;
// equal rolls yield no attack damage but counter fires.
// Rolls: attack=5, defense=5 (tie — no damage roll consumed), counter=3.
func TestAttack_TieResolver_DefenderWins(t *testing.T) {
	rulebook := makeAttackRulebook()
	factionState, attackerAsset, defenderAsset := makeAttackState(8, 8, "force-defender", false)
	faction := factionState.Factions["f1"]

	registry := hooks.NewRegistry()
	registry.RegisterTieResolver(hooks.GlobalScope(), "fanatical", &stubTieResolverForAttack{outcome: hooks.TieDefenderWins})

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockInputCollector(ctrl)
	collector.EXPECT().SelectAttackers(gomock.Any(), gomock.Any()).Return([]*domain.Asset{attackerAsset}, nil)
	collector.EXPECT().SelectDefender(gomock.Any(), gomock.Any(), gomock.Any()).Return(defenderAsset, nil)
	collector.EXPECT().SelectModifiers(gomock.Any()).Return(nil).Times(2)

	mutations := runAttack(t, collector, &fixedRoller{values: []int{5, 5, 3}}, faction, factionState, rulebook, registry)

	// TieDefenderWins: attack check (5 > 5 || 5==5 && TieDefenderWins != TieDefenderWins) → false.
	// Counter check (5 > 5 || 5==5 && TieDefenderWins != TieAttackerWins) → true → counter fires.
	if len(mutations) != 1 {
		t.Fatalf("len(mutations) = %d, want 1 (counter only); got %v", len(mutations), mutations)
	}
	delta, ok := mutations[0].(domain.AssetHPDelta)
	if !ok || delta.AssetID != "a1" || delta.Delta != -3 {
		t.Errorf("mutations[0] = %v, want AssetHPDelta{a1, -3}", mutations[0])
	}
}

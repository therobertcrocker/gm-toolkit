package dispatch

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// fixedRoller returns values from a preset sequence, cycling when exhausted.
type fixedRoller struct {
	values []int
	pos    int
}

func (roller *fixedRoller) Roll(_ int) int {
	v := roller.values[roller.pos%len(roller.values)]
	roller.pos++
	return v
}

// acceptAllCollector takes all offers and confirms all rerolls.
type acceptAllCollector struct{}

func (acceptAllCollector) SelectModifiers(offers []hooks.ModifierOffer) []hooks.ModifierOffer {
	return offers
}

func (acceptAllCollector) ConfirmReroll(_ hooks.RerollDirective) bool { return true }

// rejectAllCollector rejects every offer and declines every reroll.
type rejectAllCollector struct{}

func (rejectAllCollector) SelectModifiers(_ []hooks.ModifierOffer) []hooks.ModifierOffer {
	return nil
}

func (rejectAllCollector) ConfirmReroll(_ hooks.RerollDirective) bool { return false }

// stubModifier is a Cat 1 hook that always offers a single modifier.
type stubModifier struct {
	offer hooks.ModifierOffer
}

func (stubModifier stubModifier) OfferModifiers(_ hooks.RollContext, _ *state.FactionState, _ *rulebook.Rulebook) []hooks.ModifierOffer {
	return []hooks.ModifierOffer{stubModifier.offer}
}

// stubResultHook is a Cat 2 hook that returns a fixed directive.
type stubResultHook struct {
	directive hooks.RerollDirective
}

func (hook stubResultHook) OnRollResult(_ hooks.RollContext, _ hooks.RollResult, _ *state.FactionState, _ *rulebook.Rulebook) hooks.RerollDirective {
	return hook.directive
}

func newFaction(id string) *domain.Faction {
	return &domain.Faction{ID: id, HookBudgets: make(map[string]int)}
}

func rollCtx() hooks.RollContext {
	return hooks.RollContext{Phase: hooks.PhaseAttack}
}

func baseRoll() domain.DiceRoll {
	return domain.DiceRoll{NumDice: 2, Sides: 10, Modifier: 0}
}

func TestRollWithHooks_EmptyRegistry(t *testing.T) {
	registry := hooks.NewRegistry()
	faction := newFaction("A")
	roller := &fixedRoller{values: []int{5, 7}}

	result := RollWithHooks(rollCtx(), baseRoll(), registry, acceptAllCollector{}, roller, faction, nil, nil)

	if len(result.Dice) != 2 {
		t.Fatalf("expected 2 dice, got %d", len(result.Dice))
	}
	if result.Sum != 12 {
		t.Errorf("expected sum 12, got %d", result.Sum)
	}
}

func TestRollWithHooks_ModifierAddsExtraDie(t *testing.T) {
	registry := hooks.NewRegistry()
	mod := stubModifier{offer: hooks.ModifierOffer{
		Source: "tag:Warlike",
		Apply:  func(rs *hooks.RollState) { rs.AddDie(10) },
	}}
	registry.RegisterRollModifier(hooks.GlobalScope(), "tag:Warlike", mod)

	faction := newFaction("A")
	roller := &fixedRoller{values: []int{6, 8, 9}}

	result := RollWithHooks(rollCtx(), baseRoll(), registry, acceptAllCollector{}, roller, faction, nil, nil)

	if len(result.Dice) != 3 {
		t.Fatalf("expected 3 dice (2 base + 1 extra), got %d", len(result.Dice))
	}
	if result.Sum != 6+8+9 {
		t.Errorf("expected sum %d, got %d", 6+8+9, result.Sum)
	}
}

func TestRollWithHooks_SelectModifiersReturningEmpty(t *testing.T) {
	registry := hooks.NewRegistry()
	mod := stubModifier{offer: hooks.ModifierOffer{
		Source: "tag:Warlike",
		Apply:  func(rs *hooks.RollState) { rs.AddDie(10) },
	}}
	registry.RegisterRollModifier(hooks.GlobalScope(), "tag:Warlike", mod)

	faction := newFaction("A")
	roller := &fixedRoller{values: []int{4, 7}}

	result := RollWithHooks(rollCtx(), baseRoll(), registry, rejectAllCollector{}, roller, faction, nil, nil)

	if len(result.Dice) != 2 {
		t.Fatalf("expected 2 dice (offer rejected), got %d", len(result.Dice))
	}
}

func TestRollWithHooks_BudgetGating(t *testing.T) {
	registry := hooks.NewRegistry()
	mod := stubModifier{offer: hooks.ModifierOffer{
		Source:    "tag:Warlike",
		Apply:     func(rs *hooks.RollState) { rs.AddDie(10) },
		BudgetKey: "tag:Warlike",
	}}
	registry.RegisterRollModifier(hooks.GlobalScope(), "tag:Warlike", mod)

	faction := newFaction("A")
	roller := &fixedRoller{values: []int{5, 7, 6, 8}}

	// First call: offer should be accepted, adding an extra die and incrementing the budget.
	result := RollWithHooks(rollCtx(), baseRoll(), registry, acceptAllCollector{}, roller, faction, nil, nil)
	if len(result.Dice) != 3 {
		t.Fatalf("expected 3 dice on first call (2 base + 1 extra), got %d", len(result.Dice))
	}
	if result.Sum != 5+7+6 {
		t.Errorf("expected sum %d on first call, got %d", 5+7+6, result.Sum)
	}
	if faction.HookBudgets["tag:Warlike"] != 1 {
		t.Errorf("expected budget 1 after first call, got %d", faction.HookBudgets["tag:Warlike"])
	}

	// Second call: offer should be filtered out by budget gating, resulting in only base dice.
	result = RollWithHooks(rollCtx(), baseRoll(), registry, acceptAllCollector{}, roller, faction, nil, nil)
	if len(result.Dice) != 2 {
		t.Fatalf("expected 2 dice on second call (offer budget-gated), got %d", len(result.Dice))
	}

	if faction.HookBudgets["tag:Warlike"] != 1 {
		t.Errorf("expected budget to remain at 1 after second call, got %d", faction.HookBudgets["tag:Warlike"])
	}

}

func TestRollWithHooks_ResultHookRerollsIndex(t *testing.T) {
	registry := hooks.NewRegistry()
	hook := stubResultHook{directive: hooks.RerollDirective{
		Source:  "tag:Fanatical",
		Indices: []int{0},
	}}
	registry.RegisterRollResultHook(hooks.GlobalScope(), "tag:Fanatical", hook)

	faction := newFaction("A")
	// Die 0 rolls 1 initially; reroll gives 9.
	roller := &fixedRoller{values: []int{1, 7, 9}}

	result := RollWithHooks(rollCtx(), baseRoll(), registry, acceptAllCollector{}, roller, faction, nil, nil)

	if result.Dice[0] != 9 {
		t.Errorf("expected die 0 to be rerolled to 9, got %d", result.Dice[0])
	}
	if result.Sum != 9+7 {
		t.Errorf("expected sum %d after reroll, got %d", 9+7, result.Sum)
	}
}

func TestRollWithHooks_ElectiveRerollRejected(t *testing.T) {
	registry := hooks.NewRegistry()
	hook := stubResultHook{directive: hooks.RerollDirective{
		Source:   "tag:Fanatical",
		Indices:  []int{0},
		Elective: true,
	}}
	registry.RegisterRollResultHook(hooks.GlobalScope(), "tag:Fanatical", hook)

	faction := newFaction("A")
	roller := &fixedRoller{values: []int{3, 7}}

	result := RollWithHooks(rollCtx(), baseRoll(), registry, rejectAllCollector{}, roller, faction, nil, nil)

	// Reroll was elective and collector declined — die 0 stays at 3.
	if result.Dice[0] != 3 {
		t.Errorf("expected die 0 unchanged at 3, got %d", result.Dice[0])
	}
}

func TestRollWithHooks_RerollBudgetKey(t *testing.T) {
	registry := hooks.NewRegistry()
	hook := stubResultHook{directive: hooks.RerollDirective{
		Source:    "tag:Reroll",
		Indices:   []int{0},
		BudgetKey: "tag:Reroll",
	}}
	registry.RegisterRollResultHook(hooks.GlobalScope(), "tag:Reroll", hook)

	faction := newFaction("A")
	roller := &fixedRoller{values: []int{2, 8, 6}}

	result := RollWithHooks(rollCtx(), baseRoll(), registry, acceptAllCollector{}, roller, faction, nil, nil)

	if result.Dice[0] != 6 {
		t.Errorf("expected die 0 rerolled to 6, got %d", result.Dice[0])
	}
	if faction.HookBudgets["tag:Reroll"] != 1 {
		t.Errorf("expected budget 1, got %d", faction.HookBudgets["tag:Reroll"])
	}
}

func TestRollWithHooks_MultipleModifiersFireInOrder(t *testing.T) {
	registry := hooks.NewRegistry()
	order := make([]string, 0, 2)

	modA := stubModifier{offer: hooks.ModifierOffer{
		Source: "A",
		Apply: func(rs *hooks.RollState) {
			order = append(order, "A")
			rs.AddDie(10)
		},
	}}
	modB := stubModifier{offer: hooks.ModifierOffer{
		Source: "B",
		Apply: func(rs *hooks.RollState) {
			order = append(order, "B")
			rs.AddDie(10)
		},
	}}
	registry.RegisterRollModifier(hooks.GlobalScope(), "A", modA)
	registry.RegisterRollModifier(hooks.GlobalScope(), "B", modB)

	faction := newFaction("A")
	roller := &fixedRoller{values: []int{5, 5, 5, 5}}

	result := RollWithHooks(rollCtx(), baseRoll(), registry, acceptAllCollector{}, roller, faction, nil, nil)

	if len(result.Dice) != 4 {
		t.Fatalf("expected 4 dice, got %d", len(result.Dice))
	}
	if len(order) != 2 || order[0] != "A" || order[1] != "B" {
		t.Errorf("expected order [A B], got %v", order)
	}
}

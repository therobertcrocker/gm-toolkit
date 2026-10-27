package effects_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

const (
	ownRerollKey    = "effect:reroll:own_roll:SWN-C7-002"
	forcedRerollKey = "effect:reroll:enemy_roll_at_me:SWN-C7-002"
)

var bookOfSecretsDef = &domain.AssetDefinition{ID: "SWN-C7-002", Name: "Book of Secrets"}

func rerollEntry(trigger domain.EffectTrigger, oncePer domain.OncePerKey) domain.EffectDefinition {
	owner := domain.OwnerOwn
	if trigger == domain.TriggerEnemyRollAtMe {
		owner = domain.OwnerRival
	}
	return domain.EffectDefinition{
		Trigger: trigger,
		Effect:  domain.SpecialReroll,
		Filters: domain.EffectFilters{Owner: owner, Scope: domain.ScopeWorld},
		Gates:   domain.EffectGates{OncePer: oncePer},
		Reroll:  &domain.RerollParams{Dice: 1, Elective: true},
	}
}

var (
	ownReroll    = rerollEntry(domain.TriggerOwnRoll, domain.OncePerOwnerFaction)
	forcedReroll = rerollEntry(domain.TriggerEnemyRollAtMe, domain.OncePerOwnerFaction)
)

// bookState puts alpha's Book of Secrets "bos" on Krylos; beta and gamma are rivals.
func bookState(book *domain.Asset) *state.FactionState {
	book.Location = krylos
	return &state.FactionState{Factions: map[string]*domain.Faction{
		"alpha": {ID: "alpha", Cunning: 3, Assets: map[string]*domain.Asset{"bos": book}},
		"beta":  {ID: "beta", Cunning: 2},
		"gamma": {ID: "gamma", Cunning: 1},
	}}
}

// registerBook registers each entry for alpha's Book and returns the hooks in order.
func registerBook(factionState *state.FactionState, registry *hooks.Registry, entries ...domain.EffectDefinition) []hooks.RollResultHook {
	alpha := factionState.Factions["alpha"]
	for _, entry := range entries {
		effects.Registrars()[domain.SpecialReroll].Register(effect.RegisterContext{
			Holder: effect.Holder{Faction: alpha, Asset: alpha.Assets["bos"], Def: bookOfSecretsDef}, Entry: entry, Registry: registry,
		})
	}
	var registered []hooks.RollResultHook
	for _, entry := range registry.RollResultHooksAt(hooks.AssetScope("alpha", "bos")) {
		registered = append(registered, entry.Hook)
	}
	return registered
}

// rollOn builds a roll by actor against opponent (nil for none) on world.
func rollOn(factionState *state.FactionState, actorID, opponentID, world string) hooks.RollContext {
	ctx := hooks.RollContext{Phase: hooks.PhaseAttack, Actor: factionState.Factions[actorID], World: world}
	if opponentID != "" {
		ctx.Opponent = factionState.Factions[opponentID]
	}
	return ctx
}

func rollThroughDispatcher(factionState *state.FactionState, registry *hooks.Registry, collector hooks.Collector, ctx hooks.RollContext, dice ...int) hooks.RollResult {
	return dispatch.RollWithHooks(ctx, domain.DiceRoll{NumDice: 1, Sides: 10}, registry, collector, &seqRoller{values: dice}, ctx.Actor, factionState, nil)
}

func TestReroll_OwnRerollPicksLowestDie(t *testing.T) {
	factionState := bookState(activeAsset("bos", "SWN-C7-002"))
	hook := registerBook(factionState, hooks.NewRegistry(), ownReroll)[0]

	directive := hook.OnRollResult(rollOn(factionState, "alpha", "beta", "Krylos"), hooks.RollResult{Dice: []int{2, 9}}, factionState, nil)

	if !reflect.DeepEqual(directive.Indices, []int{0}) || !directive.Elective || directive.BudgetKey != ownRerollKey {
		t.Errorf("directive = %+v, want Indices [0], Elective, BudgetKey %q", directive, ownRerollKey)
	}
	if directive.Source != "Book of Secrets" {
		t.Errorf("Source = %q, want the definition name", directive.Source)
	}
}

func TestReroll_ForcedRerollPicksHighestDie(t *testing.T) {
	factionState := bookState(activeAsset("bos", "SWN-C7-002"))
	hook := registerBook(factionState, hooks.NewRegistry(), forcedReroll)[0]

	directive := hook.OnRollResult(rollOn(factionState, "beta", "alpha", "Krylos"), hooks.RollResult{Dice: []int{2, 9}}, factionState, nil)

	if !reflect.DeepEqual(directive.Indices, []int{1}) || directive.BudgetKey != forcedRerollKey {
		t.Errorf("directive = %+v, want Indices [1], BudgetKey %q", directive, forcedRerollKey)
	}
	if !strings.HasSuffix(directive.Source, "(forced)") {
		t.Errorf("Source = %q, want a (forced) suffix", directive.Source)
	}
}

func TestReroll_ForcedRerollSpendsOwnersBudget(t *testing.T) {
	factionState := bookState(activeAsset("bos", "SWN-C7-002"))
	registry := hooks.NewRegistry()
	registerBook(factionState, registry, ownReroll, forcedReroll)
	alpha, beta := factionState.Factions["alpha"], factionState.Factions["beta"]
	collector := &rollCollector{accept: true}
	betaAttacks := rollOn(factionState, "beta", "alpha", "Krylos")

	if result := rollThroughDispatcher(factionState, registry, collector, betaAttacks, 9, 3); result.Sum != 3 {
		t.Fatalf("first roll Sum = %d, want 3: the forced reroll replaces the 9", result.Sum)
	}
	if alpha.HookBudgets[forcedRerollKey] != 1 || beta.HookBudgets != nil {
		t.Errorf("alpha budgets %v, beta budgets %v; want alpha's forced key spent and beta untouched", alpha.HookBudgets, beta.HookBudgets)
	}
	if result := rollThroughDispatcher(factionState, registry, collector, betaAttacks, 9); result.Sum != 9 {
		t.Errorf("second roll Sum = %d, want 9: the forced reroll is spent for the turn", result.Sum)
	}
}

func TestReroll_TwoBudgetsAreSeparate(t *testing.T) {
	factionState := bookState(activeAsset("bos", "SWN-C7-002"))
	registry := hooks.NewRegistry()
	registerBook(factionState, registry, ownReroll, forcedReroll)
	collector := &rollCollector{accept: true}

	rollThroughDispatcher(factionState, registry, collector, rollOn(factionState, "beta", "alpha", "Krylos"), 9, 3)
	result := rollThroughDispatcher(factionState, registry, collector, rollOn(factionState, "alpha", "beta", "Krylos"), 2, 8)

	if result.Sum != 8 || collector.rerolls != 2 {
		t.Errorf("own roll Sum = %d after %d reroll prompts; want 8 after 2", result.Sum, collector.rerolls)
	}
	if factionState.Factions["alpha"].HookBudgets[ownRerollKey] != 1 {
		t.Errorf("own budget not spent: %v", factionState.Factions["alpha"].HookBudgets)
	}
}

func TestReroll_DeclinedRerollSpendsNothing(t *testing.T) {
	factionState := bookState(activeAsset("bos", "SWN-C7-002"))
	registry := hooks.NewRegistry()
	registerBook(factionState, registry, forcedReroll)
	collector := &rollCollector{accept: false}

	result := rollThroughDispatcher(factionState, registry, collector, rollOn(factionState, "beta", "alpha", "Krylos"), 9)

	if result.Sum != 9 || collector.rerolls != 1 {
		t.Errorf("Sum = %d after %d prompts; want 9 after 1", result.Sum, collector.rerolls)
	}
	if _, spent := factionState.Factions["alpha"].HookBudgets[forcedRerollKey]; spent {
		t.Error("a declined reroll spent the budget")
	}
}

func TestReroll_Suppression(t *testing.T) {
	factionState := bookState(activeAsset("bos", "SWN-C7-002"))
	registered := registerBook(factionState, hooks.NewRegistry(), ownReroll, forcedReroll)
	own, forced := registered[0], registered[1]
	dice := hooks.RollResult{Dice: []int{2, 9}}

	cases := []struct {
		name string
		hook hooks.RollResultHook
		ctx  hooks.RollContext
	}{
		{"forced: rival rolling against a third faction", forced, rollOn(factionState, "beta", "gamma", "Krylos")},
		{"forced: rival rolling against the owner off-world", forced, rollOn(factionState, "beta", "alpha", "Tartarus")},
		{"forced: owner's own roll", forced, rollOn(factionState, "alpha", "beta", "Krylos")},
		{"own: owner rolling on another world", own, rollOn(factionState, "alpha", "beta", "Tartarus")},
		{"own: rival rolling on the Book's world", own, rollOn(factionState, "beta", "alpha", "Krylos")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if directive := tc.hook.OnRollResult(tc.ctx, dice, factionState, nil); len(directive.Indices) != 0 {
				t.Errorf("directive = %+v, want none", directive)
			}
		})
	}
}

func TestReroll_InactiveBookRerollsNothing(t *testing.T) {
	for name, deactivate := range inactiveStates {
		t.Run(name, func(t *testing.T) {
			book := activeAsset("bos", "SWN-C7-002")
			deactivate(book)
			factionState := bookState(book)
			registered := registerBook(factionState, hooks.NewRegistry(), ownReroll, forcedReroll)
			dice := hooks.RollResult{Dice: []int{2, 9}}

			if directive := registered[0].OnRollResult(rollOn(factionState, "alpha", "beta", "Krylos"), dice, factionState, nil); len(directive.Indices) != 0 {
				t.Errorf("own directive = %+v, want none", directive)
			}
			if directive := registered[1].OnRollResult(rollOn(factionState, "beta", "alpha", "Krylos"), dice, factionState, nil); len(directive.Indices) != 0 {
				t.Errorf("forced directive = %+v, want none", directive)
			}
		})
	}
}

func TestReroll_NoOncePerGateHasNoBudgetKey(t *testing.T) {
	factionState := bookState(activeAsset("bos", "SWN-C7-002"))
	hook := registerBook(factionState, hooks.NewRegistry(), rerollEntry(domain.TriggerOwnRoll, ""))[0]

	directive := hook.OnRollResult(rollOn(factionState, "alpha", "beta", "Krylos"), hooks.RollResult{Dice: []int{2}}, factionState, nil)

	if len(directive.Indices) != 1 || directive.BudgetKey != "" {
		t.Errorf("directive = %+v, want one index and an empty BudgetKey", directive)
	}
}

func TestReroll_TagHolderRerollsEveryDieShowingTheResult(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-005"}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": {ID: "beta"}}}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialReroll].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "SWN-T-005", Name: "Fanatical"}},
		Entry: domain.EffectDefinition{
			Trigger: domain.TriggerOwnRoll, Effect: domain.SpecialReroll,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
			Reroll:  &domain.RerollParams{Result: 1},
		},
		Registry: registry,
	})
	hook := registry.RollResultHooksAt(hooks.FactionScope("alpha"))[0].Hook
	roll := rollOn(factionState, "alpha", "beta", "Tartarus")

	directive := hook.OnRollResult(roll, hooks.RollResult{Dice: []int{1, 5, 1}}, factionState, nil)
	if !reflect.DeepEqual(directive.Indices, []int{0, 2}) || directive.Elective || directive.BudgetKey != "" || directive.Source != "Fanatical" {
		t.Errorf("directive = %+v, want indices [0 2], forced, unlimited, from Fanatical", directive)
	}
	if directive := hook.OnRollResult(rollOn(factionState, "beta", "alpha", "Tartarus"), hooks.RollResult{Dice: []int{1}}, factionState, nil); len(directive.Indices) != 0 {
		t.Errorf("rival's roll: directive = %+v, want none", directive)
	}
	alpha.Tags = nil
	if directive := hook.OnRollResult(roll, hooks.RollResult{Dice: []int{1}}, factionState, nil); len(directive.Indices) != 0 {
		t.Errorf("dropped tag: directive = %+v, want none", directive)
	}
}

package dispatch

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// RollWithHooks executes a dice roll with Cat 1 (RollModifier) hooks applied
// before rolling and Cat 2 (RollResultHook) hooks applied after. Budget gating,
// collector selection, and elective-reroll confirmation are handled inline.
func RollWithHooks(
	ctx hooks.RollContext,
	baseRoll domain.DiceRoll,
	registry *hooks.Registry,
	collector hooks.Collector,
	roller domain.Roller,
	faction *domain.Faction,
	factionState *state.FactionState,
	rb *rulebook.Rulebook,
) hooks.RollResult {
	assetInstanceID := ""
	if ctx.Asset != nil {
		assetInstanceID = ctx.Asset.ID
	}

	// Cat 1: gather modifier offers, filter by budget, let collector choose.
	modifiers := registry.RollModifiersFor(faction.ID, assetInstanceID)
	var allOffers []hooks.ModifierOffer
	for _, registered := range modifiers {
		allOffers = append(allOffers, registered.Hook.OfferModifiers(ctx, factionState, rb)...)
	}
	available := offersBudgetFilter(allOffers, faction)
	chosen := collector.SelectModifiers(available)

	// Apply chosen offers: expand the dice pool and consume budgets.
	var rollState hooks.RollState
	if faction.HookBudgets == nil {
		faction.HookBudgets = make(map[string]int)
	}
	for _, offer := range chosen {
		offer.Apply(&rollState)
		if offer.BudgetKey != "" {
			faction.HookBudgets[offer.BudgetKey]++
		}
	}

	// Roll base dice pool then any extra dice queued by Apply.
	result := hooks.RollResult{Modifier: baseRoll.Modifier}
	for range baseRoll.NumDice {
		result.Dice = append(result.Dice, roller.Roll(baseRoll.Sides))
	}
	for _, sides := range rollState.ExtraDice() {
		result.Dice = append(result.Dice, roller.Roll(sides))
	}
	result.Sum = diceSum(result.Dice) + result.Modifier

	// Cat 2: apply reroll directives from result hooks.
	resultHooks := registry.RollResultHooksFor(faction.ID, assetInstanceID)
	for _, registered := range resultHooks {
		directive := registered.Hook.OnRollResult(ctx, result, factionState, rb)
		if len(directive.Indices) == 0 {
			continue
		}
		if directive.BudgetKey != "" && faction.HookBudgets[directive.BudgetKey] >= 1 {
			continue
		}
		if directive.Elective && !collector.ConfirmReroll(directive) {
			continue
		}
		for _, idx := range directive.Indices {
			if idx >= 0 && idx < len(result.Dice) {
				result.Dice[idx] = roller.Roll(baseRoll.Sides)
			}
		}
		result.Sum = diceSum(result.Dice) + result.Modifier
		if directive.BudgetKey != "" {
			faction.HookBudgets[directive.BudgetKey]++
		}
	}

	return result
}

// offersBudgetFilter removes offers whose BudgetKey has already been spent
// this turn (non-zero in faction.HookBudgets). Unlimited offers (empty key) pass through.
func offersBudgetFilter(offers []hooks.ModifierOffer, faction *domain.Faction) []hooks.ModifierOffer {
	var filtered []hooks.ModifierOffer
	for _, offer := range offers {
		if offer.BudgetKey != "" && faction.HookBudgets[offer.BudgetKey] >= 1 {
			continue
		}
		filtered = append(filtered, offer)
	}
	return filtered
}

func diceSum(dice []int) int {
	total := 0
	for _, d := range dice {
		total += d
	}
	return total
}

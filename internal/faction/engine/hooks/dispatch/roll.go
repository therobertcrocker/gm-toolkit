package dispatch

import (
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// RollWithHooks executes a dice roll with Cat 1 (RollModifier) hooks applied
// before rolling and Cat 2 (RollResultHook) hooks applied after. It consults the
// buckets rollScopes lists, so the opponent's faction and asset hooks and hooks on
// the roll's world reach the roll. Each offer and directive is stamped with the
// faction that registered it, and its budget is read and spent on that faction.
func RollWithHooks(
	ctx hooks.RollContext,
	baseRoll domain.DiceRoll,
	registry *hooks.Registry,
	collector hooks.Collector,
	roller domain.Roller,
	faction *domain.Faction,
	factionState *state.FactionState,
	rulebook *rulebook.Rulebook,
) hooks.RollResult {
	var scopes []hooks.Scope
	if registry != nil {
		scopes = rollScopes(ctx, faction)
	}

	// Cat 1: gather modifier offers, run offer filters, filter by budget, let collector choose.
	var chosen []hooks.ModifierOffer
	if registry != nil {
		var allOffers []hooks.ModifierOffer
		for _, scope := range scopes {
			for _, registered := range registry.RollModifiersAt(scope) {
				for _, offer := range registered.Hook.OfferModifiers(ctx, factionState, rulebook) {
					offer.FactionID = scopeFactionID(scope, faction)
					allOffers = append(allOffers, offer)
				}
			}
		}
		for _, scope := range scopes {
			for _, registered := range registry.RollOfferFiltersAt(scope) {
				allOffers = registered.Hook.FilterOffers(ctx, allOffers, factionState)
			}
		}
		offers := offersBudgetFilter(allOffers, faction, factionState)
		if len(offers) > 0 {
			chosen = collector.SelectModifiers(offers)
		}
	}

	// Apply chosen offers: expand the dice pool and consume budgets.
	var rollState hooks.RollState
	for _, offer := range chosen {
		offer.Apply(&rollState)
		hookOwner(offer.FactionID, faction, factionState).SpendBudget(offer.BudgetKey)
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

	// Apply keep-highest trim if any offer requested it.
	if n := rollState.KeepHighest(); n > 0 && len(result.Dice) > n {
		sort.Sort(sort.Reverse(sort.IntSlice(result.Dice)))
		result.Dice = result.Dice[:n]
		result.Sum = diceSum(result.Dice) + result.Modifier
	}

	// Cat 2: apply reroll directives from result hooks.
	// deliberate: rerolls are decided right after this roll, not after both sides roll, because a split call would restructure every roll site (Effort 3 plan, Decision 15).
	if registry == nil {
		return result
	}
	for _, scope := range scopes {
		for _, registered := range registry.RollResultHooksAt(scope) {
			directive := registered.Hook.OnRollResult(ctx, result, factionState, rulebook)
			if len(directive.Indices) == 0 {
				continue
			}
			directive.FactionID = scopeFactionID(scope, faction)
			owner := hookOwner(directive.FactionID, faction, factionState)
			if owner.BudgetSpent(directive.BudgetKey) {
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
			owner.SpendBudget(directive.BudgetKey)
		}
	}

	return result
}

// rollScopes lists the registry buckets a roll consults, in order: global, the
// roller's faction, the roller's assets, the opponent's faction, then the
// opponent's assets. Every S hook checks its owner relation itself (rollMatches),
// so a hook in the opponent's bucket acts only on rolls made against its owner.
func rollScopes(ctx hooks.RollContext, roller *domain.Faction) []hooks.Scope {
	scopes := []hooks.Scope{hooks.GlobalScope(), hooks.FactionScope(roller.ID)}
	scopes = append(scopes, assetScopes(roller, ctx.World, ctx.Asset)...)
	if ctx.Opponent != nil {
		scopes = append(scopes, hooks.FactionScope(ctx.Opponent.ID))
		scopes = append(scopes, assetScopes(ctx.Opponent, ctx.World, ctx.OpponentAsset)...)
	}
	return scopes
}

// assetScopes returns the asset buckets of faction's assets on world, in ID
// order, plus named when it stands elsewhere (in flight, or a world-less roll).
func assetScopes(faction *domain.Faction, world string, named *domain.Asset) []hooks.Scope {
	var scopes []hooks.Scope
	namedSeen := false
	if world != "" {
		for _, asset := range domain.SortedAssets(faction) {
			if asset.Location.WorldID != world {
				continue
			}
			scopes = append(scopes, hooks.AssetScope(faction.ID, asset.ID))
			if named != nil && asset.ID == named.ID {
				namedSeen = true
			}
		}
	}
	if named != nil && !namedSeen {
		scopes = append(scopes, hooks.AssetScope(faction.ID, named.ID))
	}
	return scopes
}

// scopeFactionID names the faction a hook in scope belongs to; a global hook
// belongs to the roller.
func scopeFactionID(scope hooks.Scope, roller *domain.Faction) string {
	if scope.FactionID == "" {
		return roller.ID
	}
	return scope.FactionID
}

// hookOwner returns the faction whose HookBudgets a hook spends: the faction it
// belongs to, or the roller when that faction is the roller or is not in state.
func hookOwner(factionID string, roller *domain.Faction, factionState *state.FactionState) *domain.Faction {
	if factionID == roller.ID || factionState == nil {
		return roller
	}
	if owner := factionState.Factions[factionID]; owner != nil {
		return owner
	}
	return roller
}

// offersBudgetFilter removes offers whose BudgetKey has already been spent this
// turn by the faction that made them. Unlimited offers (empty key) pass through.
func offersBudgetFilter(offers []hooks.ModifierOffer, roller *domain.Faction, factionState *state.FactionState) []hooks.ModifierOffer {
	var filtered []hooks.ModifierOffer
	for _, offer := range offers {
		if hookOwner(offer.FactionID, roller, factionState).BudgetSpent(offer.BudgetKey) {
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

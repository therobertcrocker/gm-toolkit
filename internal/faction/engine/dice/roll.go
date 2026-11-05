package dice

import (
	"maps"
	"slices"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// rollWithHooks executes a dice roll with Cat 1 (RollModifier) hooks applied
// before rolling and Cat 2 (RollResultHook) hooks applied after. It consults the
// buckets rollScopes lists, so the opponent's faction and asset hooks and hooks on
// the roll's world reach the roll. Each offer and directive is stamped with the
// faction that registered it, and its budget is read and spent on that faction.
// A collector sees a card before the dice and a card after them, offers or not. A
// nil collector asks no one: Cat 1 takes every offer with no budget and skips the
// rest, and Cat 2 declines every elective reroll (overview Decision 2). A nil
// registry has no hooks, so it asks no one either.
func rollWithHooks(
	ctx hooks.RollContext,
	baseRoll domain.DiceRoll,
	registry *hooks.Registry,
	collector hooks.Collector,
	roller Roller,
	faction *domain.Faction,
	factionState *state.FactionState,
	rulebook *rulebook.Rulebook,
) hooks.RollResult {
	if registry == nil {
		result, _ := rollDice(baseRoll, &hooks.RollState{}, roller)
		return result
	}
	scopes := rollScopes(ctx, faction, factionState)

	offers, struck := gatherOffers(ctx, scopes, registry, faction, factionState, rulebook)
	var chosen []hooks.ModifierOffer
	if collector == nil {
		chosen = freeOffers(offers)
	} else {
		chosen = collector.SelectModifiers(hooks.PreRollCard{Context: ctx, Base: baseRoll, Offers: offers, Struck: struck})
	}

	var rollState hooks.RollState
	for _, offer := range chosen {
		offer.Apply(&rollState)
		hookOwner(offer.FactionID, faction, factionState).SpendBudget(offer.BudgetKey)
	}
	result, dieSides := rollDice(baseRoll, &rollState, roller)

	// deliberate: rerolls are decided right after this roll, not after both sides roll, because a split call would restructure every roll site (Effort 3 plan, Decision 15).
	step := rerollStep{
		ctx:          ctx,
		resultHooks:  gatherResultHooks(scopes, registry, faction),
		result:       result,
		dieSides:     dieSides,
		roller:       roller,
		faction:      faction,
		factionState: factionState,
		rulebook:     rulebook,
		taken:        make(map[int]bool),
	}
	// deliberate: forced rerolls run once, before the card, not again after it, because a reroll that lands on 1 stays (D-2).
	step.applyForced()
	if collector == nil {
		return step.result
	}
	for {
		offers := step.electiveOffers()
		shown := step.result
		shown.Dice = slices.Clone(step.result.Dice)
		choice := collector.SelectReroll(hooks.PostRollCard{Context: ctx, Result: shown, Rerolls: offers, Applied: step.applied})
		if choice == nil || len(offers) == 0 {
			return step.result
		}
		step.apply(*choice)
	}
}

// rollDice rolls the base dice and any extra dice, each with its own sides, and
// applies keep-highest if an offer asked for it.
func rollDice(baseRoll domain.DiceRoll, rollState *hooks.RollState, roller Roller) (hooks.RollResult, []int) {
	result := hooks.RollResult{Modifier: baseRoll.Modifier}
	var dieSides []int
	for range baseRoll.NumDice {
		result.Dice = append(result.Dice, roller.Roll(baseRoll.Sides))
		dieSides = append(dieSides, baseRoll.Sides)
	}
	for _, sides := range rollState.ExtraDice() {
		result.Dice = append(result.Dice, roller.Roll(sides))
		dieSides = append(dieSides, sides)
	}
	if keepCount := rollState.KeepHighest(); keepCount > 0 && len(result.Dice) > keepCount {
		result.Dice, dieSides = keepHighest(result.Dice, dieSides, keepCount)
	}
	result.Sum = diceSum(result.Dice) + result.Modifier
	return result, dieSides
}

// gatherOffers collects the Cat 1 offers, runs the offer filters and the budget
// filter, and returns what survives and what did not.
func gatherOffers(
	ctx hooks.RollContext,
	scopes []hooks.Scope,
	registry *hooks.Registry,
	faction *domain.Faction,
	factionState *state.FactionState,
	rulebook *rulebook.Rulebook,
) ([]hooks.ModifierOffer, []hooks.StruckOffer) {
	var offers []hooks.ModifierOffer
	for _, scope := range scopes {
		for _, registered := range registry.RollModifiersAt(scope) {
			for _, offer := range registered.Hook.OfferModifiers(ctx, factionState, rulebook) {
				offer.FactionID = scopeFactionID(scope, faction)
				offers = append(offers, offer)
			}
		}
	}
	var struck []hooks.StruckOffer
	for _, scope := range scopes {
		for _, registered := range registry.RollOfferFiltersAt(scope) {
			kept, filtered := registered.Hook.FilterOffers(ctx, offers, factionState)
			offers = kept
			struck = append(struck, filtered...)
		}
	}
	kept, spent := offersBudgetFilter(offers, faction, factionState)
	return kept, append(struck, spent...)
}

// resultHook is one Cat 2 hook in a roll, with the faction it belongs to.
type resultHook struct {
	hook      hooks.RollResultHook
	factionID string
}

func gatherResultHooks(scopes []hooks.Scope, registry *hooks.Registry, faction *domain.Faction) []resultHook {
	var resultHooks []resultHook
	for _, scope := range scopes {
		for _, registered := range registry.RollResultHooksAt(scope) {
			resultHooks = append(resultHooks, resultHook{hook: registered.Hook, factionID: scopeFactionID(scope, faction)})
		}
	}
	return resultHooks
}

// rerollStep is one roll's Cat 2 work: the dice as they stand, the hooks that may
// change them, and the rerolls applied so far.
type rerollStep struct {
	ctx          hooks.RollContext
	resultHooks  []resultHook
	result       hooks.RollResult
	dieSides     []int
	roller       Roller
	faction      *domain.Faction
	factionState *state.FactionState
	rulebook     *rulebook.Rulebook
	applied      []hooks.AppliedReroll
	taken        map[int]bool
}

// directiveAt asks the hook in slot about the dice as they stand. It reports false
// when the hook has nothing to reroll or its budget is spent.
func (step *rerollStep) directiveAt(slot int) (hooks.RerollDirective, bool) {
	entry := step.resultHooks[slot]
	directive := entry.hook.OnRollResult(step.ctx, step.result, step.factionState, step.rulebook)
	if len(directive.Indices) == 0 {
		return hooks.RerollDirective{}, false
	}
	directive.FactionID = entry.factionID
	directive.Slot = slot
	if hookOwner(directive.FactionID, step.faction, step.factionState).BudgetSpent(directive.BudgetKey) {
		return hooks.RerollDirective{}, false
	}
	return directive, true
}

func (step *rerollStep) applyForced() {
	for slot := range step.resultHooks {
		if directive, ok := step.directiveAt(slot); ok && !directive.Elective {
			step.apply(directive)
		}
	}
}

// electiveOffers lists the elective rerolls on offer now: not spent, and not taken
// already on this roll.
func (step *rerollStep) electiveOffers() []hooks.RerollDirective {
	var offers []hooks.RerollDirective
	for slot := range step.resultHooks {
		if step.taken[slot] {
			continue
		}
		if directive, ok := step.directiveAt(slot); ok && directive.Elective {
			offers = append(offers, directive)
		}
	}
	return offers
}

func (step *rerollStep) apply(directive hooks.RerollDirective) {
	note := hooks.AppliedReroll{Source: directive.Source, FactionID: directive.FactionID}
	for _, index := range directive.Indices {
		if index >= 0 && index < len(step.result.Dice) {
			note.Was = append(note.Was, step.result.Dice[index])
			step.result.Dice[index] = step.roller.Roll(step.dieSides[index])
		}
	}
	step.result.Sum = diceSum(step.result.Dice) + step.result.Modifier
	hookOwner(directive.FactionID, step.faction, step.factionState).SpendBudget(directive.BudgetKey)
	step.applied = append(step.applied, note)
	step.taken[directive.Slot] = true
}

// rollScopes lists the registry buckets a roll consults, in order: global, the
// roller's faction, the roller's assets, the opponent's faction, the opponent's
// assets, then every other faction's faction bucket in ID order. Every S hook
// checks its owner relation itself (rollMatches), so a hook in the opponent's
// bucket acts only on rolls made against its owner, and a bystander's only
// through rival_roll (Effort 3 Decision 7). A bystander's asset buckets are never
// read: only a tag holds rival_roll.
func rollScopes(ctx hooks.RollContext, roller *domain.Faction, factionState *state.FactionState) []hooks.Scope {
	scopes := []hooks.Scope{hooks.GlobalScope(), hooks.FactionScope(roller.ID)}
	scopes = append(scopes, assetScopes(roller, ctx.World, ctx.Asset)...)
	if ctx.Opponent != nil {
		scopes = append(scopes, hooks.FactionScope(ctx.Opponent.ID))
		scopes = append(scopes, assetScopes(ctx.Opponent, ctx.World, ctx.OpponentAsset)...)
	}
	if factionState == nil {
		return scopes
	}
	for _, factionID := range slices.Sorted(maps.Keys(factionState.Factions)) {
		if factionID == roller.ID || (ctx.Opponent != nil && factionID == ctx.Opponent.ID) {
			continue
		}
		scopes = append(scopes, hooks.FactionScope(factionID))
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
// turn by the faction that made them, and reports them as struck. Unlimited offers
// (empty key) pass through.
func offersBudgetFilter(offers []hooks.ModifierOffer, roller *domain.Faction, factionState *state.FactionState) ([]hooks.ModifierOffer, []hooks.StruckOffer) {
	var filtered []hooks.ModifierOffer
	var spent []hooks.StruckOffer
	for _, offer := range offers {
		if hookOwner(offer.FactionID, roller, factionState).BudgetSpent(offer.BudgetKey) {
			spent = append(spent, hooks.StruckOffer{Offer: offer})
			continue
		}
		filtered = append(filtered, offer)
	}
	return filtered, spent
}

// freeOffers returns the offers with no budget: what a roll with no one to ask
// takes. An extra die where the highest is kept never hurts the roller, while a
// once-per-turn die spent on a roll no one chose could cost a later one.
func freeOffers(offers []hooks.ModifierOffer) []hooks.ModifierOffer {
	var free []hooks.ModifierOffer
	for _, offer := range offers {
		if offer.BudgetKey == "" {
			free = append(free, offer)
		}
	}
	return free
}

// keepHighest returns the keepCount highest dice, highest first, each with its own sides.
func keepHighest(values, dieSides []int, keepCount int) ([]int, []int) {
	order := make([]int, len(values))
	for index := range order {
		order[index] = index
	}
	sort.SliceStable(order, func(i, j int) bool { return values[order[i]] > values[order[j]] })
	keptValues := make([]int, keepCount)
	keptSides := make([]int, keepCount)
	for position, index := range order[:keepCount] {
		keptValues[position] = values[index]
		keptSides[position] = dieSides[index]
	}
	return keptValues, keptSides
}

func diceSum(dice []int) int {
	total := 0
	for _, die := range dice {
		total += die
	}
	return total
}

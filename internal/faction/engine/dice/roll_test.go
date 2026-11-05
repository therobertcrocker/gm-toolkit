package dice

import (
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice/dicetest"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// acceptAllCollector takes all offers and confirms all rerolls.
type acceptAllCollector struct{}

func (acceptAllCollector) SelectModifiers(card hooks.PreRollCard) []hooks.ModifierOffer {
	return card.Offers
}

func (acceptAllCollector) SelectReroll(card hooks.PostRollCard) *hooks.RerollDirective {
	if len(card.Rerolls) == 0 {
		return nil
	}
	return &card.Rerolls[0]
}

// rejectAllCollector rejects every offer and declines every reroll.
type rejectAllCollector struct{}

func (rejectAllCollector) SelectModifiers(_ hooks.PreRollCard) []hooks.ModifierOffer {
	return nil
}

func (rejectAllCollector) SelectReroll(_ hooks.PostRollCard) *hooks.RerollDirective { return nil }

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
	return hooks.RollContext{Phase: domain.RollPhaseAttack}
}

func baseRoll() domain.DiceRoll {
	return domain.DiceRoll{NumDice: 2, Sides: 10, Modifier: 0}
}

func TestRollWithHooks_EmptyRegistry(t *testing.T) {
	registry := hooks.NewRegistry()
	faction := newFaction("A")
	roller := &dicetest.Cycle{Values: []int{5, 7}}

	result := rollWithHooks(rollCtx(), baseRoll(), registry, acceptAllCollector{}, roller, faction, nil, nil)

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
	roller := &dicetest.Cycle{Values: []int{6, 8, 9}}

	result := rollWithHooks(rollCtx(), baseRoll(), registry, acceptAllCollector{}, roller, faction, nil, nil)

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
	roller := &dicetest.Cycle{Values: []int{4, 7}}

	result := rollWithHooks(rollCtx(), baseRoll(), registry, rejectAllCollector{}, roller, faction, nil, nil)

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
	roller := &dicetest.Cycle{Values: []int{5, 7, 6, 8}}

	// First call: offer should be accepted, adding an extra die and incrementing the budget.
	result := rollWithHooks(rollCtx(), baseRoll(), registry, acceptAllCollector{}, roller, faction, nil, nil)
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
	result = rollWithHooks(rollCtx(), baseRoll(), registry, acceptAllCollector{}, roller, faction, nil, nil)
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
	roller := &dicetest.Cycle{Values: []int{1, 7, 9}}

	result := rollWithHooks(rollCtx(), baseRoll(), registry, acceptAllCollector{}, roller, faction, nil, nil)

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
	roller := &dicetest.Cycle{Values: []int{3, 7}}

	result := rollWithHooks(rollCtx(), baseRoll(), registry, rejectAllCollector{}, roller, faction, nil, nil)

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
	roller := &dicetest.Cycle{Values: []int{2, 8, 6}}

	result := rollWithHooks(rollCtx(), baseRoll(), registry, acceptAllCollector{}, roller, faction, nil, nil)

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
	roller := &dicetest.Cycle{Values: []int{5, 5, 5, 5}}

	result := rollWithHooks(rollCtx(), baseRoll(), registry, acceptAllCollector{}, roller, faction, nil, nil)

	if len(result.Dice) != 4 {
		t.Fatalf("expected 4 dice, got %d", len(result.Dice))
	}
	if len(order) != 2 || order[0] != "A" || order[1] != "B" {
		t.Errorf("expected order [A B], got %v", order)
	}
}

// recordingCollector accepts every offer and reroll, and records what it was shown.
type recordingCollector struct {
	offers     []hooks.ModifierOffer
	directives []hooks.RerollDirective
}

func (collector *recordingCollector) SelectModifiers(card hooks.PreRollCard) []hooks.ModifierOffer {
	collector.offers = append(collector.offers, card.Offers...)
	return card.Offers
}

func (collector *recordingCollector) SelectReroll(card hooks.PostRollCard) *hooks.RerollDirective {
	if len(card.Rerolls) == 0 {
		return nil
	}
	collector.directives = append(collector.directives, card.Rerolls[0])
	return &card.Rerolls[0]
}

// failOnOffersCollector fails the test if a card arrives with an offer on it.
type failOnOffersCollector struct{ t *testing.T }

func (collector failOnOffersCollector) SelectModifiers(card hooks.PreRollCard) []hooks.ModifierOffer {
	if len(card.Offers) > 0 {
		collector.t.Fatalf("offers on the card: %+v", card.Offers)
	}
	return nil
}

func (collector failOnOffersCollector) SelectReroll(_ hooks.PostRollCard) *hooks.RerollDirective {
	return nil
}

// cardCollector records each card, and answers the after card through reroll.
type cardCollector struct {
	reroll func(hooks.PostRollCard) *hooks.RerollDirective
	before []hooks.PreRollCard
	after  []hooks.PostRollCard
}

func (collector *cardCollector) SelectModifiers(card hooks.PreRollCard) []hooks.ModifierOffer {
	collector.before = append(collector.before, card)
	return nil
}

func (collector *cardCollector) SelectReroll(card hooks.PostRollCard) *hooks.RerollDirective {
	collector.after = append(collector.after, card)
	if collector.reroll == nil {
		return nil
	}
	return collector.reroll(card)
}

// factionWithAssets builds a faction with no HookBudgets map and one asset per
// entry of assetWorlds (asset ID to world ID; "" is in flight).
func factionWithAssets(id string, assetWorlds map[string]string) *domain.Faction {
	faction := &domain.Faction{ID: id, Assets: make(map[string]*domain.Asset)}
	for assetID, worldID := range assetWorlds {
		faction.Assets[assetID] = &domain.Asset{ID: assetID, Location: domain.Location{WorldID: worldID}}
	}
	return faction
}

func rollState(factions ...*domain.Faction) *state.FactionState {
	factionState := &state.FactionState{Factions: make(map[string]*domain.Faction)}
	for _, faction := range factions {
		factionState.Factions[faction.ID] = faction
	}
	return factionState
}

func contestCtx(actor, opponent *domain.Faction, world string) hooks.RollContext {
	return hooks.RollContext{Phase: domain.RollPhaseAttack, Actor: actor, Opponent: opponent, World: world}
}

func addDieModifier(source, budgetKey string) stubModifier {
	return stubModifier{offer: hooks.ModifierOffer{
		Source:    source,
		BudgetKey: budgetKey,
		Apply:     func(rs *hooks.RollState) { rs.AddDie(10) },
	}}
}

func TestRollWithHooks_OpponentAssetHookOnWorldFires(t *testing.T) {
	alpha := factionWithAssets("alpha", nil)
	beta := factionWithAssets("beta", map[string]string{"b-1": "Krylos"})
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.AssetScope("beta", "b-1"), "b-1", addDieModifier("b-1", ""))
	collector := &recordingCollector{}

	result := rollWithHooks(contestCtx(alpha, beta, "Krylos"), baseRoll(), registry, collector, &dicetest.Cycle{Values: []int{5}}, alpha, rollState(alpha, beta), nil)

	if len(result.Dice) != 3 {
		t.Fatalf("expected 3 dice (2 base + 1 from beta's asset), got %d", len(result.Dice))
	}
	if len(collector.offers) != 1 || collector.offers[0].FactionID != "beta" {
		t.Errorf("expected one offer stamped FactionID beta, got %+v", collector.offers)
	}
}

func TestRollWithHooks_OpponentFactionHookFires(t *testing.T) {
	alpha := factionWithAssets("alpha", nil)
	beta := factionWithAssets("beta", nil)
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.FactionScope("beta"), "beta-tag", addDieModifier("beta-tag", ""))
	collector := &recordingCollector{}

	result := rollWithHooks(contestCtx(alpha, beta, "Krylos"), baseRoll(), registry, collector, &dicetest.Cycle{Values: []int{5}}, alpha, rollState(alpha, beta), nil)

	if len(result.Dice) != 3 {
		t.Fatalf("expected 3 dice (2 base + 1 from beta's faction hook), got %d", len(result.Dice))
	}
	if len(collector.offers) != 1 || collector.offers[0].FactionID != "beta" {
		t.Errorf("expected one offer stamped FactionID beta, got %+v", collector.offers)
	}
}

// takeFirstReroll takes the first reroll on offer, and passes when there is none.
func takeFirstReroll(card hooks.PostRollCard) *hooks.RerollDirective {
	if len(card.Rerolls) == 0 {
		return nil
	}
	return &card.Rerolls[0]
}

// onesHook is a forced Cat 2 hook that rerolls every die showing a 1.
type onesHook struct{ source string }

func (hook onesHook) OnRollResult(_ hooks.RollContext, result hooks.RollResult, _ *state.FactionState, _ *rulebook.Rulebook) hooks.RerollDirective {
	directive := hooks.RerollDirective{Source: hook.source}
	for index, die := range result.Dice {
		if die == 1 {
			directive.Indices = append(directive.Indices, index)
		}
	}
	return directive
}

// stripSourceFilter strikes the offers from one source and keeps the rest.
type stripSourceFilter struct{ source string }

func (filter stripSourceFilter) FilterOffers(_ hooks.RollContext, offers []hooks.ModifierOffer, _ *state.FactionState) ([]hooks.ModifierOffer, []hooks.StruckOffer) {
	var kept []hooks.ModifierOffer
	var struck []hooks.StruckOffer
	for _, offer := range offers {
		if offer.Source == filter.source {
			struck = append(struck, hooks.StruckOffer{Offer: offer, StruckBy: "strip"})
			continue
		}
		kept = append(kept, offer)
	}
	return kept, struck
}

// stripAllFilter is an offer filter that removes every offer.
type stripAllFilter struct{}

func (stripAllFilter) FilterOffers(_ hooks.RollContext, offers []hooks.ModifierOffer, _ *state.FactionState) ([]hooks.ModifierOffer, []hooks.StruckOffer) {
	struck := make([]hooks.StruckOffer, 0, len(offers))
	for _, offer := range offers {
		struck = append(struck, hooks.StruckOffer{Offer: offer, StruckBy: "strip"})
	}
	return nil, struck
}

func TestRollWithHooks_OfferFilterRemovesOffers(t *testing.T) {
	alpha := factionWithAssets("alpha", nil)
	beta := factionWithAssets("beta", map[string]string{"b-1": "Krylos"})
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.GlobalScope(), "global", addDieModifier("global", ""))
	registry.RegisterRollOfferFilter(hooks.AssetScope("beta", "b-1"), "b-1", stripAllFilter{})

	result := rollWithHooks(contestCtx(alpha, beta, "Krylos"), baseRoll(), registry, failOnOffersCollector{t}, &dicetest.Cycle{Values: []int{5}}, alpha, rollState(alpha, beta), nil)

	if len(result.Dice) != 2 {
		t.Fatalf("expected 2 dice (the opponent's filter struck the offer), got %d", len(result.Dice))
	}
}

func TestRollWithHooks_OwnAssetOnWorldFires(t *testing.T) {
	alpha := factionWithAssets("alpha", map[string]string{"a-1": "Krylos"})
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.AssetScope("alpha", "a-1"), "a-1", addDieModifier("a-1", ""))

	result := rollWithHooks(contestCtx(alpha, nil, "Krylos"), baseRoll(), registry, acceptAllCollector{}, &dicetest.Cycle{Values: []int{5}}, alpha, rollState(alpha), nil)

	if len(result.Dice) != 3 {
		t.Fatalf("expected 3 dice (own asset on world), got %d", len(result.Dice))
	}
}

func TestRollWithHooks_AssetOffWorldDoesNotFire(t *testing.T) {
	alpha := factionWithAssets("alpha", map[string]string{"a-1": "Vesta"})
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.AssetScope("alpha", "a-1"), "a-1", addDieModifier("a-1", ""))

	result := rollWithHooks(contestCtx(alpha, nil, "Krylos"), baseRoll(), registry, failOnOffersCollector{t}, &dicetest.Cycle{Values: []int{5}}, alpha, rollState(alpha), nil)

	if len(result.Dice) != 2 {
		t.Fatalf("expected 2 dice (asset on another world), got %d", len(result.Dice))
	}
}

func TestRollWithHooks_NamedAssetInFlightFires(t *testing.T) {
	alpha := factionWithAssets("alpha", map[string]string{"a-1": ""})
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.AssetScope("alpha", "a-1"), "a-1", addDieModifier("a-1", ""))
	ctx := contestCtx(alpha, nil, "")
	ctx.Asset = alpha.Assets["a-1"]

	result := rollWithHooks(ctx, baseRoll(), registry, acceptAllCollector{}, &dicetest.Cycle{Values: []int{5}}, alpha, rollState(alpha), nil)

	if len(result.Dice) != 3 {
		t.Fatalf("expected 3 dice (rolling asset in flight), got %d", len(result.Dice))
	}
}

func TestRollWithHooks_OpponentBudgetSpentOnOpponent(t *testing.T) {
	alpha := factionWithAssets("alpha", nil)
	beta := factionWithAssets("beta", map[string]string{"b-1": "Krylos"})
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.AssetScope("beta", "b-1"), "b-1", addDieModifier("b-1", "k"))
	factionState := rollState(alpha, beta)
	roller := &dicetest.Cycle{Values: []int{5}}

	rollWithHooks(contestCtx(alpha, beta, "Krylos"), baseRoll(), registry, acceptAllCollector{}, roller, alpha, factionState, nil)

	if beta.HookBudgets["k"] != 1 {
		t.Errorf("expected beta budget k = 1, got %d", beta.HookBudgets["k"])
	}
	if alpha.HookBudgets != nil {
		t.Errorf("expected alpha budgets untouched (nil), got %v", alpha.HookBudgets)
	}

	result := rollWithHooks(contestCtx(alpha, beta, "Krylos"), baseRoll(), registry, failOnOffersCollector{t}, roller, alpha, factionState, nil)
	if len(result.Dice) != 2 {
		t.Fatalf("expected 2 dice on second roll (beta budget spent), got %d", len(result.Dice))
	}
}

func TestRollWithHooks_OpponentRerollBudgetSpentOnOpponent(t *testing.T) {
	alpha := factionWithAssets("alpha", nil)
	beta := factionWithAssets("beta", map[string]string{"b-1": "Krylos"})
	registry := hooks.NewRegistry()
	registry.RegisterRollResultHook(hooks.AssetScope("beta", "b-1"), "b-1", stubResultHook{directive: hooks.RerollDirective{
		Source:    "b-1",
		Indices:   []int{0},
		BudgetKey: "k",
		Elective:  true,
	}})
	factionState := rollState(alpha, beta)
	collector := &recordingCollector{}
	// First roll: [2, 8], die 0 rerolled to 6. Second roll: [3, 4], no reroll.
	roller := &dicetest.Cycle{Values: []int{2, 8, 6, 3, 4}}

	result := rollWithHooks(contestCtx(alpha, beta, "Krylos"), baseRoll(), registry, collector, roller, alpha, factionState, nil)

	if result.Dice[0] != 6 {
		t.Errorf("expected die 0 rerolled to 6, got %d", result.Dice[0])
	}
	if len(collector.directives) != 1 || collector.directives[0].FactionID != "beta" {
		t.Errorf("expected one directive stamped FactionID beta, got %+v", collector.directives)
	}
	if beta.HookBudgets["k"] != 1 {
		t.Errorf("expected beta budget k = 1, got %d", beta.HookBudgets["k"])
	}
	if alpha.HookBudgets != nil {
		t.Errorf("expected alpha budgets untouched (nil), got %v", alpha.HookBudgets)
	}

	result = rollWithHooks(contestCtx(alpha, beta, "Krylos"), baseRoll(), registry, collector, roller, alpha, factionState, nil)
	if result.Dice[0] != 3 {
		t.Errorf("expected die 0 unchanged at 3 on second roll (beta budget spent), got %d", result.Dice[0])
	}
}

func TestRollWithHooks_NoOffersStillAsks(t *testing.T) {
	registry := hooks.NewRegistry()
	faction := newFaction("A")
	collector := &cardCollector{}

	result := rollWithHooks(rollCtx(), baseRoll(), registry, collector, &dicetest.Cycle{Values: []int{5, 7}}, faction, nil, nil)

	if len(result.Dice) != 2 {
		t.Fatalf("expected 2 dice, got %d", len(result.Dice))
	}
	if len(collector.before) != 1 || len(collector.after) != 1 {
		t.Errorf("saw %d before cards and %d after cards, want one of each", len(collector.before), len(collector.after))
	}
}

func TestRollWithHooks_NilRegistryAsksNoOne(t *testing.T) {
	collector := &cardCollector{}

	result := rollWithHooks(rollCtx(), baseRoll(), nil, collector, &dicetest.Cycle{Values: []int{5, 7}}, newFaction("A"), nil, nil)

	if len(collector.before) != 0 || len(collector.after) != 0 {
		t.Errorf("saw %d before cards and %d after cards, want none", len(collector.before), len(collector.after))
	}
	if !slices.Equal(result.Dice, []int{5, 7}) || result.Sum != 12 {
		t.Errorf("dice = %v, sum %d, want [5 7], sum 12", result.Dice, result.Sum)
	}
}

func TestRollWithHooks_StruckOffersReachTheCard(t *testing.T) {
	alpha := factionWithAssets("alpha", nil)
	alpha.HookBudgets = map[string]int{"spent": 1}
	beta := factionWithAssets("beta", map[string]string{"b-1": "Krylos"})
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.GlobalScope(), "tag", addDieModifier("tag", ""))
	registry.RegisterRollOfferFilter(hooks.AssetScope("beta", "b-1"), "b-1", stripSourceFilter{source: "tag"})
	registry.RegisterRollModifier(hooks.GlobalScope(), "once", addDieModifier("once", "spent"))
	registry.RegisterRollModifier(hooks.GlobalScope(), "free", addDieModifier("free", ""))
	collector := &cardCollector{}

	rollWithHooks(contestCtx(alpha, beta, "Krylos"), baseRoll(), registry, collector, &dicetest.Cycle{Values: []int{5}}, alpha, rollState(alpha, beta), nil)

	card := collector.before[0]
	if len(card.Offers) != 1 || card.Offers[0].Source != "free" {
		t.Errorf("offers = %+v, want only free", card.Offers)
	}
	if len(card.Struck) != 2 {
		t.Fatalf("struck = %+v, want the filter's strike and the spent budget", card.Struck)
	}
	if card.Struck[0].Offer.Source != "tag" || card.Struck[0].StruckBy != "strip" {
		t.Errorf("first strike = %+v, want tag struck by strip", card.Struck[0])
	}
	if card.Struck[1].Offer.Source != "once" || card.Struck[1].StruckBy != "" {
		t.Errorf("second strike = %+v, want once with no StruckBy", card.Struck[1])
	}
}

func TestRollWithHooks_ForcedRerollRunsBeforeTheCard(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterRollResultHook(hooks.GlobalScope(), "elective", stubResultHook{directive: hooks.RerollDirective{Source: "elective", Indices: []int{1}, Elective: true}})
	registry.RegisterRollResultHook(hooks.GlobalScope(), "forced", stubResultHook{directive: hooks.RerollDirective{Source: "forced", Indices: []int{0}}})
	collector := &cardCollector{}

	result := rollWithHooks(rollCtx(), baseRoll(), registry, collector, &dicetest.Sequence{Values: []int{1, 4, 8}}, newFaction("A"), nil, nil)

	if !slices.Equal(result.Dice, []int{8, 4}) {
		t.Errorf("dice = %v, want [8 4]", result.Dice)
	}
	card := collector.after[0]
	if len(card.Applied) != 1 || card.Applied[0].Source != "forced" || !slices.Equal(card.Applied[0].Was, []int{1}) {
		t.Errorf("applied = %+v, want forced, which replaced a 1", card.Applied)
	}
	if len(card.Rerolls) != 1 || card.Rerolls[0].Source != "elective" {
		t.Errorf("rerolls = %+v, want only elective", card.Rerolls)
	}
}

func TestRollWithHooks_ForcedRerollRunsOnce(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterRollResultHook(hooks.GlobalScope(), "forced", onesHook{source: "forced"})
	registry.RegisterRollResultHook(hooks.GlobalScope(), "elective", stubResultHook{directive: hooks.RerollDirective{Source: "elective", Indices: []int{0}, Elective: true}})
	collector := &cardCollector{reroll: takeFirstReroll}
	roller := &sidesRecorder{values: []int{1, 5, 1}}

	result := rollWithHooks(rollCtx(), domain.DiceRoll{NumDice: 1, Sides: 10}, registry, collector, roller, newFaction("A"), nil, nil)

	if !slices.Equal(result.Dice, []int{1}) {
		t.Errorf("dice = %v, want [1]: a 1 that a reroll lands on stays", result.Dice)
	}
	if len(roller.sides) != 3 {
		t.Errorf("rolled %d times, want 3", len(roller.sides))
	}
}

func TestRollWithHooks_ElectiveRerollsLoopUntilAPass(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterRollResultHook(hooks.GlobalScope(), "first", stubResultHook{directive: hooks.RerollDirective{Source: "first", Indices: []int{0}, Elective: true}})
	registry.RegisterRollResultHook(hooks.GlobalScope(), "second", stubResultHook{directive: hooks.RerollDirective{Source: "second", Indices: []int{1}, Elective: true}})
	collector := &cardCollector{}
	collector.reroll = func(card hooks.PostRollCard) *hooks.RerollDirective {
		if len(collector.after) == 1 {
			return &card.Rerolls[0]
		}
		return nil
	}

	result := rollWithHooks(rollCtx(), baseRoll(), registry, collector, &dicetest.Sequence{Values: []int{2, 3, 9}}, newFaction("A"), nil, nil)

	if !slices.Equal(result.Dice, []int{9, 3}) {
		t.Errorf("dice = %v, want [9 3]", result.Dice)
	}
	if len(collector.after) != 2 {
		t.Fatalf("saw %d after cards, want 2", len(collector.after))
	}
	if len(collector.after[0].Rerolls) != 2 {
		t.Errorf("first card rerolls = %+v, want two", collector.after[0].Rerolls)
	}
	if len(collector.after[1].Rerolls) != 1 || collector.after[1].Rerolls[0].Source != "second" {
		t.Errorf("second card rerolls = %+v, want only second", collector.after[1].Rerolls)
	}
}

func TestRollWithHooks_TakenRerollIsOfferedOnce(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterRollResultHook(hooks.GlobalScope(), "unlimited", stubResultHook{directive: hooks.RerollDirective{Source: "unlimited", Indices: []int{0}, Elective: true}})
	collector := &cardCollector{reroll: takeFirstReroll}
	roller := &sidesRecorder{values: []int{2, 9}}

	result := rollWithHooks(rollCtx(), domain.DiceRoll{NumDice: 1, Sides: 10}, registry, collector, roller, newFaction("A"), nil, nil)

	if !slices.Equal(result.Dice, []int{9}) || len(roller.sides) != 2 {
		t.Errorf("dice = %v after %d rolls, want [9] after 2", result.Dice, len(roller.sides))
	}
	if len(collector.after) != 2 || len(collector.after[1].Rerolls) != 0 {
		t.Errorf("after cards = %+v, want two, the second with no rerolls", collector.after)
	}
}

func TestRollWithHooks_AnswerWithNothingOnOfferEndsTheRoll(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterRollResultHook(hooks.GlobalScope(), "unlimited", stubResultHook{directive: hooks.RerollDirective{Source: "unlimited", Indices: []int{0}, Elective: true}})
	alwaysTaken := hooks.RerollDirective{Source: "unlimited", Indices: []int{0}, Elective: true}
	collector := &cardCollector{reroll: func(hooks.PostRollCard) *hooks.RerollDirective { return &alwaysTaken }}
	roller := &sidesRecorder{values: []int{2, 9}}

	result := rollWithHooks(rollCtx(), domain.DiceRoll{NumDice: 1, Sides: 10}, registry, collector, roller, newFaction("A"), nil, nil)

	if !slices.Equal(result.Dice, []int{9}) || len(roller.sides) != 2 {
		t.Errorf("dice = %v after %d rolls, want [9] after 2: an answer to a card with no rerolls is not applied", result.Dice, len(roller.sides))
	}
	if len(collector.after) != 2 {
		t.Errorf("saw %d after cards, want 2", len(collector.after))
	}
}

func TestRollWithHooks_KeptCardKeepsItsDice(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterRollResultHook(hooks.GlobalScope(), "unlimited", stubResultHook{directive: hooks.RerollDirective{Source: "unlimited", Indices: []int{0}, Elective: true}})
	collector := &cardCollector{reroll: takeFirstReroll}

	rollWithHooks(rollCtx(), domain.DiceRoll{NumDice: 1, Sides: 10}, registry, collector, &sidesRecorder{values: []int{2, 9}}, newFaction("A"), nil, nil)

	if !slices.Equal(collector.after[0].Result.Dice, []int{2}) || !slices.Equal(collector.after[1].Result.Dice, []int{9}) {
		t.Errorf("cards show %v then %v, want [2] then [9]", collector.after[0].Result.Dice, collector.after[1].Result.Dice)
	}
}

func TestRollWithHooks_ScopeOrder(t *testing.T) {
	alpha := factionWithAssets("alpha", map[string]string{"a-1": "Krylos", "a-2": "Krylos"})
	beta := factionWithAssets("beta", map[string]string{"b-1": "Krylos"})
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.GlobalScope(), "global", addDieModifier("global", ""))
	registry.RegisterRollModifier(hooks.FactionScope("alpha"), "faction", addDieModifier("faction", ""))
	registry.RegisterRollModifier(hooks.AssetScope("alpha", "a-2"), "a-2", addDieModifier("a-2", ""))
	registry.RegisterRollModifier(hooks.AssetScope("alpha", "a-1"), "a-1", addDieModifier("a-1", ""))
	registry.RegisterRollModifier(hooks.AssetScope("beta", "b-1"), "b-1", addDieModifier("b-1", ""))
	collector := &recordingCollector{}

	rollWithHooks(contestCtx(alpha, beta, "Krylos"), baseRoll(), registry, collector, &dicetest.Cycle{Values: []int{5}}, alpha, rollState(alpha, beta), nil)

	want := []string{"global", "faction", "a-1", "a-2", "b-1"}
	if len(collector.offers) != len(want) {
		t.Fatalf("expected %d offers, got %d", len(want), len(collector.offers))
	}
	for i, source := range want {
		if collector.offers[i].Source != source {
			t.Errorf("offer %d: expected %q, got %q", i, source, collector.offers[i].Source)
		}
	}
}

// TestRollWithHooks_CapturedAssetOldBucketDoesNotFire: "b-1" is in alpha's map, but
// a hook is still filed under beta. Asset scopes come from each faction's current
// assets, so neither side's roll reaches a former owner's bucket, even before the
// orchestrator clears it (Effort 1 Decision Record 13).
func TestRollWithHooks_CapturedAssetOldBucketDoesNotFire(t *testing.T) {
	alpha := factionWithAssets("alpha", map[string]string{"b-1": "Krylos"})
	beta := factionWithAssets("beta", map[string]string{"b-2": "Krylos"})
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.AssetScope("beta", "b-1"), "b-1", addDieModifier("b-1", ""))
	factionState := rollState(alpha, beta)

	for name, rolling := range map[string]struct{ actor, opponent *domain.Faction }{
		"the old owner rolls": {beta, alpha},
		"the new owner rolls": {alpha, beta},
	} {
		t.Run(name, func(t *testing.T) {
			result := rollWithHooks(contestCtx(rolling.actor, rolling.opponent, "Krylos"), baseRoll(), registry, failOnOffersCollector{t}, &dicetest.Cycle{Values: []int{5}}, rolling.actor, factionState, nil)
			if len(result.Dice) != 2 {
				t.Fatalf("expected 2 dice (old bucket unreached), got %d", len(result.Dice))
			}
		})
	}
}

func TestRollWithHooks_NilCollectorTakesOnlyFreeOffers(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.GlobalScope(), "free", addDieModifier("free", ""))
	registry.RegisterRollModifier(hooks.GlobalScope(), "limited", addDieModifier("limited", "tag:Warlike"))
	faction := newFaction("A")

	result := rollWithHooks(rollCtx(), baseRoll(), registry, nil, &dicetest.Cycle{Values: []int{5, 7, 6}}, faction, nil, nil)

	if len(result.Dice) != 3 {
		t.Fatalf("expected 3 dice (2 base + the free die), got %d", len(result.Dice))
	}
	if faction.HookBudgets["tag:Warlike"] != 0 {
		t.Errorf("expected the limited offer skipped and its budget unspent, got %d", faction.HookBudgets["tag:Warlike"])
	}
}

func TestRollWithHooks_NilCollectorDeclinesElectiveRerolls(t *testing.T) {
	cases := []struct {
		name     string
		elective bool
		want     int
	}{
		{"a forced reroll applies", false, 9},
		{"an elective reroll is declined", true, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			registry := hooks.NewRegistry()
			registry.RegisterRollResultHook(hooks.GlobalScope(), "reroll", stubResultHook{directive: hooks.RerollDirective{
				Source: "reroll", Indices: []int{0}, Elective: tc.elective,
			}})

			result := rollWithHooks(rollCtx(), baseRoll(), registry, nil, &dicetest.Cycle{Values: []int{1, 7, 9}}, newFaction("A"), nil, nil)

			if result.Dice[0] != tc.want {
				t.Errorf("die 0 = %d, want %d", result.Dice[0], tc.want)
			}
		})
	}
}

func TestRollWithHooks_NilCollectorAppliesForcedRerollsWithNoCard(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterRollResultHook(hooks.GlobalScope(), "forced", stubResultHook{directive: hooks.RerollDirective{Source: "forced", Indices: []int{0}}})
	roller := &sidesRecorder{values: []int{1, 8}}

	result := rollWithHooks(rollCtx(), domain.DiceRoll{NumDice: 1, Sides: 10}, registry, nil, roller, newFaction("A"), nil, nil)

	if !slices.Equal(result.Dice, []int{8}) || len(roller.sides) != 2 {
		t.Errorf("dice = %v after %d rolls, want [8] after 2: one forced reroll", result.Dice, len(roller.sides))
	}
}

func TestRollWithHooks_BystanderFactionHookFires(t *testing.T) {
	alpha := factionWithAssets("alpha", nil)
	beta := factionWithAssets("beta", nil)
	gamma := factionWithAssets("gamma", map[string]string{"g-1": "Krylos"})
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.FactionScope("gamma"), "gamma-tag", addDieModifier("gamma-tag", ""))
	registry.RegisterRollModifier(hooks.AssetScope("gamma", "g-1"), "g-1", addDieModifier("g-1", ""))
	collector := &recordingCollector{}

	rollWithHooks(contestCtx(alpha, beta, "Krylos"), baseRoll(), registry, collector, &dicetest.Cycle{Values: []int{5}}, alpha, rollState(alpha, beta, gamma), nil)

	if len(collector.offers) != 1 || collector.offers[0].Source != "gamma-tag" || collector.offers[0].FactionID != "gamma" {
		t.Errorf("offers = %+v, want only gamma's faction hook, stamped gamma: a bystander's assets are not read", collector.offers)
	}
}

// sidesRecorder returns values in order and records the sides of each roll.
type sidesRecorder struct {
	values []int
	sides  []int
}

func (roller *sidesRecorder) Roll(sides int) int {
	value := roller.values[len(roller.sides)]
	roller.sides = append(roller.sides, sides)
	return value
}

func rerollRegistry(index int, apply func(*hooks.RollState)) *hooks.Registry {
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.GlobalScope(), "extra", stubModifier{offer: hooks.ModifierOffer{Source: "extra", Apply: apply}})
	registry.RegisterRollResultHook(hooks.GlobalScope(), "reroll", stubResultHook{directive: hooks.RerollDirective{Source: "reroll", Indices: []int{index}}})
	return registry
}

func TestRollWithHooks_RerollKeepsExtraDieSides(t *testing.T) {
	registry := rerollRegistry(1, func(rs *hooks.RollState) { rs.AddDie(6) })
	roller := &sidesRecorder{values: []int{4, 1, 6}}

	result := rollWithHooks(rollCtx(), domain.DiceRoll{NumDice: 1, Sides: 10}, registry, acceptAllCollector{}, roller, newFaction("A"), nil, nil)

	if !slices.Equal(roller.sides, []int{10, 6, 6}) {
		t.Errorf("rolled sides = %v, want [10 6 6]: the reroll of the extra d6 has 6 sides", roller.sides)
	}
	if !slices.Equal(result.Dice, []int{4, 6}) || result.Sum != 10 {
		t.Errorf("dice = %v, sum %d, want [4 6], sum 10", result.Dice, result.Sum)
	}
}

func TestRollWithHooks_RerollAfterKeepHighestKeepsSides(t *testing.T) {
	cases := []struct {
		name      string
		values    []int
		wantSides []int
		wantDice  []int
	}{
		{"the extra d6 is kept over the d10", []int{3, 5, 2}, []int{10, 6, 6}, []int{2}},
		{"the d10 is kept over the extra d6", []int{8, 5, 9}, []int{10, 6, 10}, []int{9}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			registry := rerollRegistry(0, func(rs *hooks.RollState) {
				rs.AddDie(6)
				rs.SetKeepHighest(1)
			})
			roller := &sidesRecorder{values: testCase.values}

			result := rollWithHooks(rollCtx(), domain.DiceRoll{NumDice: 1, Sides: 10}, registry, acceptAllCollector{}, roller, newFaction("A"), nil, nil)

			if !slices.Equal(roller.sides, testCase.wantSides) {
				t.Errorf("rolled sides = %v, want %v", roller.sides, testCase.wantSides)
			}
			if !slices.Equal(result.Dice, testCase.wantDice) {
				t.Errorf("dice = %v, want %v", result.Dice, testCase.wantDice)
			}
		})
	}
}

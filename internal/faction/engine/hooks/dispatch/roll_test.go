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

// recordingCollector accepts every offer and reroll, and records what it was shown.
type recordingCollector struct {
	offers     []hooks.ModifierOffer
	directives []hooks.RerollDirective
}

func (collector *recordingCollector) SelectModifiers(offers []hooks.ModifierOffer) []hooks.ModifierOffer {
	collector.offers = append(collector.offers, offers...)
	return offers
}

func (collector *recordingCollector) ConfirmReroll(directive hooks.RerollDirective) bool {
	collector.directives = append(collector.directives, directive)
	return true
}

// failOnSelectCollector fails the test if the dispatcher prompts for modifiers.
type failOnSelectCollector struct{ t *testing.T }

func (collector failOnSelectCollector) SelectModifiers(_ []hooks.ModifierOffer) []hooks.ModifierOffer {
	collector.t.Fatal("SelectModifiers must not be called")
	return nil
}

func (collector failOnSelectCollector) ConfirmReroll(_ hooks.RerollDirective) bool { return true }

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
	return hooks.RollContext{Phase: hooks.PhaseAttack, Actor: actor, Opponent: opponent, World: world}
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

	result := RollWithHooks(contestCtx(alpha, beta, "Krylos"), baseRoll(), registry, collector, &fixedRoller{values: []int{5}}, alpha, rollState(alpha, beta), nil)

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

	result := RollWithHooks(contestCtx(alpha, beta, "Krylos"), baseRoll(), registry, collector, &fixedRoller{values: []int{5}}, alpha, rollState(alpha, beta), nil)

	if len(result.Dice) != 3 {
		t.Fatalf("expected 3 dice (2 base + 1 from beta's faction hook), got %d", len(result.Dice))
	}
	if len(collector.offers) != 1 || collector.offers[0].FactionID != "beta" {
		t.Errorf("expected one offer stamped FactionID beta, got %+v", collector.offers)
	}
}

// stripAllFilter is an offer filter that removes every offer.
type stripAllFilter struct{}

func (stripAllFilter) FilterOffers(_ hooks.RollContext, _ []hooks.ModifierOffer, _ *state.FactionState) []hooks.ModifierOffer {
	return nil
}

func TestRollWithHooks_OfferFilterRemovesOffers(t *testing.T) {
	alpha := factionWithAssets("alpha", nil)
	beta := factionWithAssets("beta", map[string]string{"b-1": "Krylos"})
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.GlobalScope(), "global", addDieModifier("global", ""))
	registry.RegisterRollOfferFilter(hooks.AssetScope("beta", "b-1"), "b-1", stripAllFilter{})

	result := RollWithHooks(contestCtx(alpha, beta, "Krylos"), baseRoll(), registry, failOnSelectCollector{t}, &fixedRoller{values: []int{5}}, alpha, rollState(alpha, beta), nil)

	if len(result.Dice) != 2 {
		t.Fatalf("expected 2 dice (the opponent's filter struck the offer), got %d", len(result.Dice))
	}
}

func TestRollWithHooks_OwnAssetOnWorldFires(t *testing.T) {
	alpha := factionWithAssets("alpha", map[string]string{"a-1": "Krylos"})
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.AssetScope("alpha", "a-1"), "a-1", addDieModifier("a-1", ""))

	result := RollWithHooks(contestCtx(alpha, nil, "Krylos"), baseRoll(), registry, acceptAllCollector{}, &fixedRoller{values: []int{5}}, alpha, rollState(alpha), nil)

	if len(result.Dice) != 3 {
		t.Fatalf("expected 3 dice (own asset on world), got %d", len(result.Dice))
	}
}

func TestRollWithHooks_AssetOffWorldDoesNotFire(t *testing.T) {
	alpha := factionWithAssets("alpha", map[string]string{"a-1": "Vesta"})
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.AssetScope("alpha", "a-1"), "a-1", addDieModifier("a-1", ""))

	result := RollWithHooks(contestCtx(alpha, nil, "Krylos"), baseRoll(), registry, failOnSelectCollector{t}, &fixedRoller{values: []int{5}}, alpha, rollState(alpha), nil)

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

	result := RollWithHooks(ctx, baseRoll(), registry, acceptAllCollector{}, &fixedRoller{values: []int{5}}, alpha, rollState(alpha), nil)

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
	roller := &fixedRoller{values: []int{5}}

	RollWithHooks(contestCtx(alpha, beta, "Krylos"), baseRoll(), registry, acceptAllCollector{}, roller, alpha, factionState, nil)

	if beta.HookBudgets["k"] != 1 {
		t.Errorf("expected beta budget k = 1, got %d", beta.HookBudgets["k"])
	}
	if alpha.HookBudgets != nil {
		t.Errorf("expected alpha budgets untouched (nil), got %v", alpha.HookBudgets)
	}

	result := RollWithHooks(contestCtx(alpha, beta, "Krylos"), baseRoll(), registry, failOnSelectCollector{t}, roller, alpha, factionState, nil)
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
	roller := &fixedRoller{values: []int{2, 8, 6, 3, 4}}

	result := RollWithHooks(contestCtx(alpha, beta, "Krylos"), baseRoll(), registry, collector, roller, alpha, factionState, nil)

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

	result = RollWithHooks(contestCtx(alpha, beta, "Krylos"), baseRoll(), registry, collector, roller, alpha, factionState, nil)
	if result.Dice[0] != 3 {
		t.Errorf("expected die 0 unchanged at 3 on second roll (beta budget spent), got %d", result.Dice[0])
	}
}

func TestRollWithHooks_NoOffersSkipsCollector(t *testing.T) {
	registry := hooks.NewRegistry()
	faction := newFaction("A")

	result := RollWithHooks(rollCtx(), baseRoll(), registry, failOnSelectCollector{t}, &fixedRoller{values: []int{5, 7}}, faction, nil, nil)

	if len(result.Dice) != 2 {
		t.Fatalf("expected 2 dice, got %d", len(result.Dice))
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

	RollWithHooks(contestCtx(alpha, beta, "Krylos"), baseRoll(), registry, collector, &fixedRoller{values: []int{5}}, alpha, rollState(alpha, beta), nil)

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
			result := RollWithHooks(contestCtx(rolling.actor, rolling.opponent, "Krylos"), baseRoll(), registry, failOnSelectCollector{t}, &fixedRoller{values: []int{5}}, rolling.actor, factionState, nil)
			if len(result.Dice) != 2 {
				t.Fatalf("expected 2 dice (old bucket unreached), got %d", len(result.Dice))
			}
		})
	}
}

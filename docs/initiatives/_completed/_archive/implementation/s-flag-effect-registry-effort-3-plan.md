# S-flag Effect Registry (A.001.3) — Effort 3 Plan: Roll Hooks

## Session Cadence

This effort executes in four sessions, not one per commit. Commits that set a contract
get their own session for fresh re-grounding. Commits that only use a contract share a
session, with a check-in between each commit.

| Session | Commits | Why |
|---------|---------|-----|
| 1 | Commit 1 | Sets the roll-dispatch contract that all later commits use. |
| 2 | Commit 2 | Reaches outside the S side; re-grounds against the A side. |
| 3 | Commits 3, 4, 5 | Same pattern three times; the pattern carries forward. |
| 4 | Commit 6 | Changes the registrar contract and the orchestrator. |

## Context / Goal

Make roll hooks reach every faction roll and every party to it, then author the five
assets that ride them. Today `RollWithHooks` consults only the roller's own hooks, and
only the Attack action calls it. Effort 3 closes dispatch items 3 and 7 and ships
`bonus_die`, `remove_tag_dice`, `reroll`, and the turn-start half of `clear_stealth`.

- Overview plan: [`s-flag-effect-registry-plan.md`](./s-flag-effect-registry-plan.md)
  (Effort 3 row, Planning Decisions 3, 14 and 15, Shared Context).
- Discovery: [`s-flag-effect-registry-discovery.md`](../../discovery/completed/s-flag-effect-registry-discovery.md)
  (dispatch items 3 and 7; Open Questions 3 and 4).
- Builds on Effort 1's registrar contract (`effect/registrar.go`) and Effort 2's
  budget helpers (`effects/budget.go`), `active()` rule (Effort 2 Decision Record 11)
  and contest helper (`effects/contest.go`).
- Assets: Integral Protocols `SWN-F7-002` (`bonus_die` only; `restrict_attack` is
  Effort 4), Surveyors `SWN-W2-004`, Panopticon Matrix `SWN-C8-001` (both entries),
  Blackmail `SWN-C2-003`, Book of Secrets `SWN-C7-002` (both entries).

This is the only effort that reaches outside the S side: Commit 2 changes Expand
Influence and the A-side ability contest. Execution re-grounds against
`expand_influence.go` and `ability/stage.go` before it touches them.

## Decisions Ratified in Planning

**Rules readings (Robert).**

1. **Panopticon Matrix tests at the start of its owner's turn.** "At the beginning of
   every turn" reads as the owner's turn, once per round. `turn_start` as built walks
   the acting faction's assets, so the walk is unchanged. Resolves discovery Open
   Question 3.
2. **Book of Secrets forces a reroll only on a roll made against its owner.** An enemy
   rolling against the owner on that world: attack, defense, an Expand Influence rival
   roll, or an ability contest. A roll between two other factions is out of reach. This
   matches the ratified `enemy_roll_at_me` trigger. Resolves discovery Open Question 4.
3. **Book of Secrets has two separate budgets.** One own reroll and one forced reroll
   per turn. "This reroll can only be forced once per turn" would be redundant under a
   shared reading.
4. **Surveyors' die applies only to the owner's own expansion test.** The faction roll
   when the owner expands on that world. It does not apply to the owner's opposing roll
   against a rival's expansion, or to its base's defense in the free attack. The
   authored entry adds `phase = ["faction_test"]`.
5. **Panopticon's die applies to Cunning attacks and defenses only.** Not to Expand
   Influence tests or ability contests, even when they roll Cunning. The authored entry
   adds `phase = ["attack", "defense"]`.
6. **Book of Secrets' own reroll applies to any roll its owner makes on that world.**
   "Reroll one die for an action on that world" carries no phase restriction, and the
   overview authored no phase filter.

**Design.**

7. **The roll dispatcher gathers asset buckets on the roll's world, from both sides.**
   `RollWithHooks` consults: global; the roller's faction bucket; the asset bucket of
   every roller asset on `ctx.World`; the asset bucket of every opponent asset on
   `ctx.World`. The rolling asset and the opposing asset are added when they are not on
   that world (in flight, or a world-less roll). Assets iterate in ID order. **The
   opponent's faction bucket is never consulted:** tag hooks register there and assume
   the roll is their own faction's (Fanatical rerolls every 1 without checking
   `ctx.Actor`). Every S-side roll hook registers at asset scope and checks its owner
   relation against `ctx.Actor` / `ctx.Opponent` itself. That check is how the `owner`
   filter reads for roll triggers. Resolves the overview's item-3 question ("decide how
   an opponent-sourced offer is presented").

8. **The dispatcher stamps the hook's owning faction on every offer and directive, and
   spends the budget there.** `ModifierOffer` and `RerollDirective` grow `FactionID`,
   set by the dispatcher from the registration scope (the roller for a global hook),
   never by the hook. The budget is read from and spent on that faction's `HookBudgets`,
   not the roller's. This is Planning Decision 3 applied to the roll path: a forced
   reroll spends the Book owner's budget. The same field tells the collector whose
   effect it is being asked about.

9. **`RollContext` grows `OpponentAsset` and `Action`.** `OpponentAsset` is the asset
   on the other side of the roll (the attacker, for a defense roll). `Action` is the
   action's `Name()`, which gives the `action` filter a source.

10. **Phase meaning for the newly routed rolls.** `PhaseFactionTest` is the acting
    faction's own test: the Expand Influence faction roll, and the acting side of an
    ability contest. `PhaseContested` is a roll made to oppose the acting faction: each
    Expand Influence rival roll, and the target side of an ability contest. The free
    attack inside Expand Influence uses `PhaseAttack` / `PhaseDefense`.

11. **The Expand Influence faction roll has no opponent.** It is compared against N
    rival rolls, so `Opponent` is nil and no rival's opponent-side hooks reach it.
    Under Decision 2, a rival's Book of Secrets cannot force it. A fence-sign marks it.

12. **Item 7 covers Expand Influence and the A-side ability contest.** `passesGate`
    rolls through `RollWithHooks` against the real acting faction, not the
    `workingFaction` clone, so a budget spent there persists. `UseAssetAbility` and
    `ability.Dispatch` take the registry. The S-side contest (`effects/contest.go`)
    stays a direct roll with a fence-sign: a reactor has no collector to prompt.

13. **The dispatcher does not prompt `SelectModifiers` on an empty offer list.** The
    TUI opens a "Pre-roll Modifiers" overlay for every call, empty or not
    (`execution.go:475`). Routing more rolls through the dispatcher would multiply that
    empty prompt.

14. **`remove_tag_dice` is a new Cat 1 interface, `RollOfferFilter`, plus a typed
    `ModifierOffer.FromTag`.** A `RollModifier` can only add offers. The filter runs
    after every modifier has offered and before the budget filter and the collector.
    Warlike sets `FromTag`. Rejected: matching the `"tag:"` prefix of `Source`, which
    couples the effects package to the tags package's naming.

15. **Elective rerolls keep per-roll timing.** Cat 2 runs right after each roll, so the
    attacker decides on a reroll before the defense is rolled. Splitting
    `RollWithHooks` into roll and post-roll calls would restructure every call site,
    and Expand Influence compares one roll against N. A fence-sign at the dispatcher
    and a planned-work entry record it.

16. **`budgetKey` gains a trigger segment, for every entry.** Book of Secrets' two
    entries share effect and definition, so the Effort 2 key collides
    (`effect:reroll:SWN-C7-002` for both). The key becomes
    `effect:<name>:<trigger>:<defID>[:<target-or-asset>]`. The rule is the same for
    every entry. Effort 2's reactors store their trigger and pass it. Reconciles
    Planning Decision 3's key format in the overview.

17. **The `phase` filter becomes a list.** `EffectFilters.Phase` becomes
    `Phases []EffectRollPhase`, and `phase` in TOML becomes an array, like `asset_type`
    and `category`. Panopticon needs two phases (Decision 5). This changes the Effort 1
    domain contract; no shipped entry uses `phase` yet.

18. **The turn-start route sees the whole state.** `FireContext` and `FireTurnStart`
    grow `FactionState`, because Panopticon needs the rival assets on its world and the
    rival's Cunning. `HookBudgets` clearing after `FireTurnStart` (overview Effort 3
    outline) does not matter here: Panopticon's turn-start entry has no `once_per`.

19. **A bonus die keeps the highest single die.** `bonus_die` adds `count` dice of
    `sides` and sets keep-highest 1, the same as Warlike. Every faction roll is a single
    d10, and the rules read an extra die as "roll an additional d10 and keep the
    highest result". Two Panopticons on one world each offer a die; the collector picks.

## Decision Record — Execution

<!-- Empty at plan time. Execution-time decisions and reversals land here, logged
before the commit they ride. Append AND reconcile. -->

1. **The roll's world is held in `worldID`, not `world`** (Commit 2, Tasks 3 and 4).
   `expand_influence.go` imports the `world` package, so a local `world` would hide it
   inside `baseAttack.resolve`. `passesGate` uses the same name for consistency. The
   plan body is updated to match.
2. **`passesGate` names the action in a fenced constant** (Commit 2, Task 4). The
   `ability` package cannot call `UseAssetAbility.Name()`, because `actions` imports
   `ability`. The literal moves into `const actionName = "Use Asset Ability"`, with a
   fence-sign that names the import cycle. The plan body is updated to match.
3. **Roll-hook tests use a local `rollCollector`** (Commit 3, Task 10). `effects_test`
   had no collector. It does not import the harness's `ScriptedCollector`, because that
   pulls the whole engine into a registrar test. The two-method `rollCollector` (in
   `bonus_die_test.go`) accepts or declines all and counts its prompts. Commits 4 and 5
   reuse it. The Task 11 scenario adds `testharness.DefSurveyors`, as Franchise and
   Party Machine have.
4. **`hooks.md` gets two more edits than Task 10 lists** (Commit 4, Task 10). The
   per-bucket accessor list gains `RollOfferFiltersAt`. The collector-seam text no
   longer says the collector sees "the full Cat 1 modifier list": it sees what survives
   the offer filters and budgets. Both would be wrong otherwise. The dispatcher's Cat 1
   step comment also names the filter pass. The Warlike-vs-Blackmail test also asserts
   that a struck Warlike offer leaves its budget unspent.
5. **Two literal budget keys outside Task 4's search gain the trigger** (Commit 5,
   Task 4). `testharness/scenarios/effects_test.go` (Franchise's steal) is outside
   `effects/*_test.go`. `ignore_coin_loss_test.go` asserts that a key is *absent*, so an
   old-format key would pass vacuously. That test now also asserts that `bank-2`'s key
   *is* spent, as a positive control on the format. `reroll_test.go` adds three
   suppression cases to Task 8's list: a forced reroll off-world, a forced entry on the
   owner's own roll, and an own entry on a rival's roll.
6. **The `FactionState` passthrough gets its own test** (Commit 6, Task 6). The
   existing turn-start test passes a nil state, so an assertion there would pass
   vacuously. `TestFireTurnStart_PassesFactionStateToThePhaseRoute` passes a real state
   and checks that the same pointer reaches `FireContext`.
7. **Turn-start suppression cases isolate one filter each** (Commit 6, Task 6). Each
   case removes the other assets from `tripwireState`, so zero rolls proves the one
   filter under test. The open-courier case removes the spy and alpha's agent. The
   owner case removes the spy and the courier. The off-world case puts the Matrix on
   Krylos, because `tripwireState` has no rival asset off Tartarus.
8. **The scenario pins the turn order** (Commit 6, Task 6). `Turn.Start` rolls the
   order with the turn engine's own roller, not `Engine.Rand`. The test sets
   `FactionOrder = ["beta", "alpha"]` and runs `RunFactionTurn` two times. Beta's turn
   runs on an empty `FixedRoller`, so any roll there fails the test. It adds
   `testharness.DefPanopticonMatrix`.
9. **`hooks/registry.go` is gofmt'd** (Commit 6). It had a one-space alignment slip
   from Commit 4.
10. **The catalog also marks Effort 2's shipped assets** (Commit 6, Task 7). Zealots,
    Blockade Fleet, Franchise, Bank and Tripwire Cells (both of its rows) shipped in
    Effort 2 but had no ✅. Bank's hook cell becomes Cat 3 `MutationReactor`, which is
    what `ignore_coin_loss` registers. The `FireTurnStart` HP-check gap found here is
    filed as `B.013`.
11. **`clear_stealth` skips non-stealthy assets on both routes** (interim review fix).
    The `stealthy` filter was the only guard. An entry without it rolled contests
    against open assets and emitted `AssetStealthCleared` for them. Both
    `fireClearStealth` and the `stealth_arrival` reactor now skip `!Stealthy` assets.
    The reactor path had the same gap, so it is fixed too.
12. **Bonus dice stack per asset copy** (interim review). Two Surveyors on one world
    give two `bonus_die` offers, so the roll is 3d10 keep-highest. Each copy grants its
    own die. A fence-sign in `bonus_die.go` marks this.
13. **Decision 10 is confirmed for ability contests** (interim review). An ability
    contest stays `faction_test` / `contested`, also where the rule text calls it an
    attack. Panopticon's `[attack, defense]` die does not apply to it.
14. **Hook budgets are `domain.Faction` methods** (interim review fix).
    `BudgetSpent(key)` and `SpendBudget(key)` replace the two private copies in
    `dispatch` and `effects`. An empty key is unlimited. `effects.spendBudget` stays as
    the check-and-spend wrapper the reactors use.
15. **`passesGate` keeps `workingAsset`** (interim review, finding rejected). Relocate
    steps must come before gated steps, so the contest belongs on the new world.
    Other assets on that world do not move during an ability, so the real-state scan
    finds them. `assetScopes` adds the moved asset through `named`, and `scope = self`
    matches by ID. The one gap is a `scope = world` roll hook on the moved asset itself,
    because its live location is stale. No shipped asset has one, so it is not filed.
    The reviewer's other dropped findings: forced-reroll budgets reset in the owner's
    bookkeeping by design (`effects/budget.go`, Planning Decision 3). Blackmail strips
    tag bonus dice, not tag rerolls. A Base of Influence is not a `domain.Asset`, so
    `baseAttack` has no `OpponentAsset` to set.

---

## Shared Context

**Re-grounding.** Commits 3–6 bind to Commits 1 and 2's dispatcher shape. Each
execution session re-reads `hooks/dispatch/roll.go` and the effects helpers it cites
before it writes a registrar.

**Data copies.** `campaigns/test-camp/rulebook/assets/` holds copies of the SWN asset
files (Effort 1 plan, Shared Context). Every data edit in this effort applies to both
files. Note: Effort 2's entries (Zealots, Blockade Fleet, Franchise, Bank, Tripwire
Cells) were authored in `rulebooks/swn` only. That drift is not this effort's to fix,
and is raised for the pre-merge checklist.

**Test patterns.** Registrar tests live in `package effects_test` and use the existing
helpers: `seqRoller`, `fixedRoller`, `activeAsset`, `inactiveStates`. Every filter and
gate a registrar accepts gets a test that proves it suppresses (overview Shared
Context). Every roll hook gets an inactive-asset test over `inactiveStates`.

**Model.** Opus for every commit. Commits 1 and 2 change shared dispatch and two actions
outside the S side.

## Out of Scope

- **`restrict_attack` for Integral Protocols.** Effort 4, Commit 5 (Planning Decision 14).
- **Elective reroll after both rolls.** Decision 15; filed to planned-work.
- **Third-party rolls on a world** (Book of Secrets reaching a roll between two other
  factions). Decision 2.
- **S-side contests through `RollWithHooks`.** Decision 12; the reactor has no
  collector.
- **Effort 2's test-camp data drift.** See Shared Context.

---

## Work Breakdown

### Phase 1 — Dispatch reach

Items 3 and 7. The roll path learns to see the opponent and the world, and every
faction roll goes through it. No effect ships here.

#### Commit 1 — `feat: roll hooks reach the opponent and the world's assets`

##### Task 1 — `internal/faction/engine/hooks/types.go`

Give the roll context the opposing asset and the action name, and give offers and
directives the faction whose hook made them (Decisions 8, 9). `Action` lands here
because the Attack call sites change in this commit anyway. Commit 2 gives it its
other sources.

**Find:**
```go
// RollContext carries the full context of a roll being processed by hooks.
type RollContext struct {
	Phase     RollPhase
	Actor     *domain.Faction
	Opponent  *domain.Faction
	Attribute string
	Asset     *domain.Asset
	World     string
}
```
**Replace with:**
```go
// RollContext carries the full context of a roll being processed by hooks.
// Asset is the rolling asset and OpponentAsset the asset on the other side;
// either is nil when no asset stands there (a faction test, a Base defending).
// Action is the action's Name().
type RollContext struct {
	Phase         RollPhase
	Actor         *domain.Faction
	Opponent      *domain.Faction
	Attribute     string
	Asset         *domain.Asset
	OpponentAsset *domain.Asset
	World         string
	Action        string
}
```

**Find:**
```go
type ModifierOffer struct {
	Source      string
	Description string
	BudgetKey   string           // namespaced, e.g. "tag:Warlike"; empty = unlimited
	Apply       func(*RollState) // expands the dice pool before rolling
}
```
**Replace with:**
```go
// FactionID is the faction whose hook made the offer; the dispatcher sets it, and
// spends BudgetKey on that faction's HookBudgets.
type ModifierOffer struct {
	Source      string
	Description string
	BudgetKey   string           // namespaced, e.g. "tag:Warlike"; empty = unlimited
	Apply       func(*RollState) // expands the dice pool before rolling
	FactionID   string
}
```

**Find:**
```go
type RerollDirective struct {
	Source    string
	Indices   []int  // indices into RollResult.Dice to reroll
	BudgetKey string // namespaced; empty = unlimited
	Elective  bool   // if true, collector must confirm before applying
}
```
**Replace with:**
```go
// FactionID is the faction whose hook made the directive; the dispatcher sets it,
// and spends BudgetKey on that faction's HookBudgets.
type RerollDirective struct {
	Source    string
	Indices   []int  // indices into RollResult.Dice to reroll
	BudgetKey string // namespaced; empty = unlimited
	Elective  bool   // if true, collector must confirm before applying
	FactionID string
}
```

##### Task 2 — `internal/faction/engine/hooks/registry.go`

The dispatcher now walks a list of scopes it builds itself, so the Cat 1 and Cat 2
lookups become per-bucket accessors, as Cat 3's did in Effort 2. `RollModifiersFor`
and `RollResultHooksFor` have no other production caller.

**Find:**
```go
// RollModifiersFor returns all Cat 1 hooks matching the given faction and asset,
// in order: global, then faction-scoped, then asset-scoped.
func (registry *Registry) RollModifiersFor(factionID, assetInstanceID string) []RegisteredRollModifier {
	var result []RegisteredRollModifier
	result = append(result, registry.rollModifiers[GlobalScope()]...)
	result = append(result, registry.rollModifiers[FactionScope(factionID)]...)
	if assetInstanceID != "" {
		result = append(result, registry.rollModifiers[AssetScope(factionID, assetInstanceID)]...)
	}
	return result
}

// RollResultHooksFor returns all Cat 2 hooks matching the given faction and asset.
func (registry *Registry) RollResultHooksFor(factionID, assetInstanceID string) []RegisteredRollResultHook {
	var result []RegisteredRollResultHook
	result = append(result, registry.rollResultHooks[GlobalScope()]...)
	result = append(result, registry.rollResultHooks[FactionScope(factionID)]...)
	if assetInstanceID != "" {
		result = append(result, registry.rollResultHooks[AssetScope(factionID, assetInstanceID)]...)
	}
	return result
}
```
**Replace with:**
```go
// RollModifiersAt returns the Cat 1 hooks registered in one scope bucket. The roll
// dispatcher decides which buckets a roll consults.
func (registry *Registry) RollModifiersAt(scope Scope) []RegisteredRollModifier {
	return registry.rollModifiers[scope]
}

// RollResultHooksAt returns the Cat 2 hooks registered in one scope bucket.
func (registry *Registry) RollResultHooksAt(scope Scope) []RegisteredRollResultHook {
	return registry.rollResultHooks[scope]
}
```

##### Task 3 — `internal/faction/engine/hooks/dispatch/roll.go` (rewritten)

Full contents:
```go
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
// buckets rollScopes lists, so an opponent's asset hooks and hooks on the roll's
// world reach the roll. Each offer and directive is stamped with the faction that
// registered it, and its budget is read and spent on that faction.
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

	// Cat 1: gather modifier offers, filter by budget, let collector choose.
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
		offers := offersBudgetFilter(allOffers, faction, factionState)
		if len(offers) > 0 {
			chosen = collector.SelectModifiers(offers)
		}
	}

	// Apply chosen offers: expand the dice pool and consume budgets.
	var rollState hooks.RollState
	for _, offer := range chosen {
		offer.Apply(&rollState)
		spendBudget(hookOwner(offer.FactionID, faction, factionState), offer.BudgetKey)
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
			if budgetSpent(owner, directive.BudgetKey) {
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
			spendBudget(owner, directive.BudgetKey)
		}
	}

	return result
}

// rollScopes lists the registry buckets a roll consults, in order: global, the
// roller's faction, the roller's assets, then the opponent's assets.
// deliberate: the opponent's faction bucket is left out, not added, because tag hooks register there and assume the roll is their own faction's (Fanatical would reroll the opponent's 1s).
func rollScopes(ctx hooks.RollContext, roller *domain.Faction) []hooks.Scope {
	scopes := []hooks.Scope{hooks.GlobalScope(), hooks.FactionScope(roller.ID)}
	scopes = append(scopes, assetScopes(roller, ctx.World, ctx.Asset)...)
	if ctx.Opponent != nil {
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

func budgetSpent(owner *domain.Faction, key string) bool {
	return key != "" && owner.HookBudgets[key] >= 1
}

func spendBudget(owner *domain.Faction, key string) {
	if key == "" {
		return
	}
	if owner.HookBudgets == nil {
		owner.HookBudgets = make(map[string]int)
	}
	owner.HookBudgets[key]++
}

// offersBudgetFilter removes offers whose BudgetKey has already been spent this
// turn by the faction that made them. Unlimited offers (empty key) pass through.
func offersBudgetFilter(offers []hooks.ModifierOffer, roller *domain.Faction, factionState *state.FactionState) []hooks.ModifierOffer {
	var filtered []hooks.ModifierOffer
	for _, offer := range offers {
		if budgetSpent(hookOwner(offer.FactionID, roller, factionState), offer.BudgetKey) {
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
```

The old code always created the roller's `HookBudgets` map. The new code creates it only
when it spends a budget. `HookBudgets` is `omitempty`, so persisted state is unchanged.

##### Task 4 — `internal/faction/engine/action/actions/attack.go`

Give both combat rolls the opposing asset and the action name.

**Find:**
```go
			hooks.RollContext{
				Phase:     hooks.PhaseAttack,
				Actor:     faction,
				Opponent:  defenderFaction,
				Attribute: string(attackerDef.Attack.AttackerStat),
				Asset:     attacker,
				World:     attacker.Location.WorldID,
			},
```
**Replace with:**
```go
			hooks.RollContext{
				Phase:         hooks.PhaseAttack,
				Actor:         faction,
				Opponent:      defenderFaction,
				Attribute:     string(attackerDef.Attack.AttackerStat),
				Asset:         attacker,
				OpponentAsset: defender,
				World:         attacker.Location.WorldID,
				Action:        attack.Name(),
			},
```

**Find:**
```go
			hooks.RollContext{
				Phase:     hooks.PhaseDefense,
				Actor:     defenderFaction,
				Opponent:  faction,
				Attribute: string(attackerDef.Attack.DefenderStat),
				Asset:     defender,
				World:     attacker.Location.WorldID,
			},
```
**Replace with:**
```go
			hooks.RollContext{
				Phase:         hooks.PhaseDefense,
				Actor:         defenderFaction,
				Opponent:      faction,
				Attribute:     string(attackerDef.Attack.DefenderStat),
				Asset:         defender,
				OpponentAsset: attacker,
				World:         attacker.Location.WorldID,
				Action:        attack.Name(),
			},
```

##### Task 5 — `internal/faction/engine/action/actions/attack_test.go`

`TestAttack_TieResolver_DefenderWins` registers only a tie resolver, so no offer is made
and `SelectModifiers` is no longer called (Decision 13).

**Find:**
```go
	collector.EXPECT().SelectDefender(gomock.Any(), gomock.Any(), gomock.Any()).Return(defenderAsset, nil)
	collector.EXPECT().SelectModifiers(gomock.Any()).Return(nil).Times(2)
```
**Replace with:**
```go
	collector.EXPECT().SelectDefender(gomock.Any(), gomock.Any(), gomock.Any()).Return(defenderAsset, nil)
```

##### Task 6 — `internal/faction/engine/hooks/registry_test.go`

Move the Cat 1 and Cat 2 tests to the `*At` accessors, following the Cat 3 tests
Effort 2 already moved (`TestMutationReactor_*`).

- `TestRollModifier_RegistrationOrderPreserved` / `TestRollResultHook_RegistrationOrderPreserved`:
  read `RollModifiersAt(FactionScope("f1"))` / `RollResultHooksAt(FactionScope("f1"))`.
- `*_FactionScopeFilters`: assert `*At(FactionScope("f2"))` is empty.
- `TestRollModifier_AssetScopeFilters`: assert `RollModifiersAt(AssetScope("f1", "asset-Y"))`
  is empty.
- `*_GlobalScopeMatchesAnyFaction`: rename to `*_GlobalScopeIsSeparateFromFactionScope`.
  A global registration appears in `*At(GlobalScope())` and not in `*At(FactionScope(id))`,
  mirroring `TestMutationReactor_GlobalScopeIsSeparateFromFactionScope`.
- `*_EmptyRegistryReturnsEmptySlice`: read `*At(AssetScope("f1", "asset-X"))` /
  `*At(FactionScope("f1"))`.

##### Task 7 — `internal/faction/engine/hooks/dispatch/roll_test.go`

The existing cases use global registrations, a nil state and an empty world, so they
pass unchanged. Add:

- `TestRollWithHooks_OpponentAssetHookOnWorldFires`: alpha rolls against beta on
  `"Krylos"`. A `stubModifier` is registered at `AssetScope("beta", "b-1")` with `b-1`
  on Krylos. Expect 3 dice (2 base + 1), and the offer the collector sees carries
  `FactionID == "beta"`.
- `TestRollWithHooks_OpponentFactionHookDoesNotFire`: the same, but registered at
  `FactionScope("beta")`. Expect 2 dice, and the collector is never called (use a
  collector that fails the test on `SelectModifiers`).
- `TestRollWithHooks_OwnAssetOnWorldFires` and `_AssetOffWorldDoesNotFire`: an alpha
  asset hook fires when the asset is on `ctx.World` and not when it is elsewhere.
- `TestRollWithHooks_NamedAssetInFlightFires`: `ctx.World == ""`, `ctx.Asset` set. Its
  asset-scope hook fires.
- `TestRollWithHooks_OpponentBudgetSpentOnOpponent`: a beta asset hook's offer with
  `BudgetKey: "k"` is chosen. Then `beta.HookBudgets["k"] == 1`, alpha's map is still
  nil, and a second roll does not offer it.
- `TestRollWithHooks_OpponentRerollBudgetSpentOnOpponent`: the same for a Cat 2
  directive with `BudgetKey`.
- `TestRollWithHooks_NoOffersSkipsCollector`: a registry with no Cat 1 hooks never
  calls `SelectModifiers`.
- `TestRollWithHooks_ScopeOrder`: modifiers registered at global, alpha faction, alpha
  asset `a-2`, alpha asset `a-1`, and beta asset `b-1` (all on the world) reach the
  collector in the order global, faction, `a-1`, `a-2`, `b-1`.

Build the state with a helper `rollState(...)` returning a `*state.FactionState`, and
pass it as `factionState`.

##### Task 8 — `docs/architecture/engine/hooks.md`

- In **The registry and scope lookup**, after the Cat 3 paragraph, add: Cat 1 and Cat 2
  are also exposed per bucket (`RollModifiersAt(scope)`, `RollResultHooksAt(scope)`),
  because the roll dispatcher builds its own scope list.
- In **Registration in, dispatch out**, replace the `RollWithHooks` bullet's first
  sentence with: it consults global, the roller's faction, and the asset buckets of the
  roller's and the opponent's assets on the roll's world (plus the rolling and opposing
  assets when off-world), in asset-ID order. Never the opponent's faction bucket: tag
  hooks assume the roll is their own.
- In **Per-turn budgets**, replace "it increments `faction.HookBudgets[key]`" with:
  the dispatcher stamps each offer and directive with the faction that registered it
  (`FactionID`) and reads and spends the budget on that faction's `HookBudgets`. An
  opponent's effect spends its owner's budget.
- In **The collector seam**, add one sentence: `SelectModifiers` is not called when no
  offer survives the budget filter.

##### Commit message

```
feat: roll hooks reach the opponent and the world's assets

- RollContext gains OpponentAsset and Action
- RollWithHooks consults asset buckets of both sides on the roll's world
- leave the opponent's faction bucket out so tag hooks stay own-roll only
- stamp the owning faction on offers and directives; spend budgets there
- skip SelectModifiers when no offer survives
- replace RollModifiersFor/RollResultHooksFor with per-bucket accessors
```

#### Commit 2 — `feat: route every roll through RollWithHooks`

Item 7. Expand Influence's three rolls and the A-side ability contest stop calling the
roller directly. Behaviour with no hooks is unchanged: `RollWithHooks` with no offers
rolls exactly `NumDice` dice in the same order, so existing `seqRoller` sequences still
line up.

##### Task 1 — `internal/faction/engine/action/actions/expand_influence.go`

The action takes the registry.

**Find:**
```go
import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
```
**Replace with:**
```go
import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
```

**Find:**
```go
type ExpandInfluence struct {
	collector   action.Collector
	roller      domain.Roller
	index       *world.Index
	worldEngine *world.WorldEngine
	order       action.ExpandInfluenceOrder
	mutations   []domain.Mutation
}

func NewExpandInfluence(collector action.Collector, roller domain.Roller, index *world.Index, worldEngine *world.WorldEngine) *ExpandInfluence {
	return &ExpandInfluence{collector: collector, roller: roller, index: index, worldEngine: worldEngine}
}
```
**Replace with:**
```go
type ExpandInfluence struct {
	collector   action.Collector
	roller      domain.Roller
	registry    *hooks.Registry
	index       *world.Index
	worldEngine *world.WorldEngine
	order       action.ExpandInfluenceOrder
	mutations   []domain.Mutation
}

func NewExpandInfluence(collector action.Collector, roller domain.Roller, registry *hooks.Registry, index *world.Index, worldEngine *world.WorldEngine) *ExpandInfluence {
	return &ExpandInfluence{collector: collector, roller: roller, registry: registry, index: index, worldEngine: worldEngine}
}
```

##### Task 2 — `internal/faction/engine/action/actions/expand_influence.go`

The faction test and each rival roll (Decisions 10, 11).

**Find:**
```go
	// Contested roll: faction rolls 1d10 + Cunning; each rival on the world rolls the same.
	// Rivals that tie or beat the faction's roll may make a free attack against the new base.
	factionRoll := ei.roller.Roll(10) + faction.Cunning
	baseHPTracker := 0
```
**Replace with:**
```go
	// Contested roll: faction rolls 1d10 + Cunning; each rival on the world rolls the same.
	// Rivals that tie or beat the faction's roll may make a free attack against the new base.
	// deliberate: no Opponent on the faction's roll, not each rival in turn, because it is one roll compared against N rivals (Effort 3 plan, Decision 11).
	factionRoll := dispatch.RollWithHooks(
		hooks.RollContext{
			Phase:     hooks.PhaseFactionTest,
			Actor:     faction,
			Attribute: string(domain.StatCunning),
			World:     ei.order.World,
			Action:    ei.Name(),
		},
		domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: faction.Cunning},
		ei.registry, ei.collector, ei.roller, faction, factionState, rulebook,
	).Sum
	baseHPTracker := 0
```

**Find:**
```go
		rivalRoll := ei.roller.Roll(10) + rival.Cunning
```
**Replace with:**
```go
		rivalRoll := dispatch.RollWithHooks(
			hooks.RollContext{
				Phase:     hooks.PhaseContested,
				Actor:     rival,
				Opponent:  faction,
				Attribute: string(domain.StatCunning),
				World:     ei.order.World,
				Action:    ei.Name(),
			},
			domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: rival.Cunning},
			ei.registry, ei.collector, ei.roller, rival, factionState, rulebook,
		).Sum
```

**Find:**
```go
		sub := &baseAttack{collector: ei.collector, roller: ei.roller}
		if err := sub.resolve(rival, &newBase, faction, &ei.mutations, &baseHPTracker, rulebook); err != nil {
```
**Replace with:**
```go
		sub := &baseAttack{collector: ei.collector, roller: ei.roller, registry: ei.registry, action: ei.Name()}
		if err := sub.resolve(rival, &newBase, faction, &ei.mutations, &baseHPTracker, factionState, rulebook); err != nil {
```

##### Task 3 — `internal/faction/engine/action/actions/expand_influence.go`

The free attack against the new Base. The defense keeps its existing Cunning stat; the
Base has no asset, so `Asset` is nil on the defense roll.

**Find:**
```go
type baseAttack struct {
	collector action.Collector
	roller    domain.Roller
}
```
**Replace with:**
```go
type baseAttack struct {
	collector action.Collector
	roller    domain.Roller
	registry  *hooks.Registry
	action    string
}
```

**Find:**
```go
func (ba *baseAttack) resolve(rival *domain.Faction, target *domain.Base, ownerFaction *domain.Faction, mutations *[]domain.Mutation, baseHPTracker *int, rulebook *rulebook.Rulebook) error {
```
**Replace with:**
```go
func (ba *baseAttack) resolve(rival *domain.Faction, target *domain.Base, ownerFaction *domain.Faction, mutations *[]domain.Mutation, baseHPTracker *int, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
```

**Find:**
```go
		attackRoll := ba.roller.Roll(10) + statScore(rival, attackerDef.Attack.AttackerStat)
		defenseRoll := ba.roller.Roll(10) + statScore(ownerFaction, domain.StatCunning)
```
**Replace with:**
```go
		worldID := target.Location.WorldID
		attackRoll := dispatch.RollWithHooks(
			hooks.RollContext{
				Phase:     hooks.PhaseAttack,
				Actor:     rival,
				Opponent:  ownerFaction,
				Attribute: string(attackerDef.Attack.AttackerStat),
				Asset:     attacker,
				World:     worldID,
				Action:    ba.action,
			},
			domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: statScore(rival, attackerDef.Attack.AttackerStat)},
			ba.registry, ba.collector, ba.roller, rival, factionState, rulebook,
		).Sum
		defenseRoll := dispatch.RollWithHooks(
			hooks.RollContext{
				Phase:         hooks.PhaseDefense,
				Actor:         ownerFaction,
				Opponent:      rival,
				Attribute:     string(domain.StatCunning),
				OpponentAsset: attacker,
				World:         worldID,
				Action:        ba.action,
			},
			domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: statScore(ownerFaction, domain.StatCunning)},
			ba.registry, ba.collector, ba.roller, ownerFaction, factionState, rulebook,
		).Sum
```

##### Task 4 — `internal/faction/engine/action/actions/ability/stage.go`

The contest rolls both sides through the dispatcher (Decisions 10, 12).

**Find:**
```go
import (
	"fmt"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)
```
**Replace with:**
```go
import (
	"fmt"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)
```

**Find:**
```go
func passesGate(step domain.AbilityStep, faction *domain.Faction, target domain.AbilityTarget, roller domain.Roller) (bool, error) {
	if step.Contest == nil {
		// Ungated (self-move, roll-table) — always proceeds. Roll-table outcome
		// resolution lands in Effort 2.
		return true, nil
	}
	if target.Faction == nil {
		return false, fmt.Errorf("contest: target faction not resolved")
	}
	attack := roller.Roll(10) + statScore(faction, step.Contest.AttackerStat)
	defense := roller.Roll(10) + statScore(target.Faction, step.Contest.DefenderStat)
	return attack > defense, nil
}
```
**Replace with:**
```go
// passesGate rolls the step's contest through the roll hooks. faction is the real
// acting faction, so a budget a hook spends persists past this step.
func passesGate(
	step domain.AbilityStep,
	faction *domain.Faction,
	asset *domain.Asset,
	target domain.AbilityTarget,
	registry *hooks.Registry,
	collector action.Collector,
	roller domain.Roller,
	factionState *state.FactionState,
	rb *rulebook.Rulebook,
) (bool, error) {
	if step.Contest == nil {
		// Ungated (self-move, roll-table) — always proceeds. Roll-table outcome
		// resolution lands in Effort 2.
		return true, nil
	}
	if target.Faction == nil {
		return false, fmt.Errorf("contest: target faction not resolved")
	}
	worldID := asset.Location.WorldID
	// deliberate: a literal, not UseAssetAbility.Name(), because actions imports ability (import cycle).
	const actionName = "Use Asset Ability"
	attack := dispatch.RollWithHooks(
		hooks.RollContext{
			Phase:         hooks.PhaseFactionTest,
			Actor:         faction,
			Opponent:      target.Faction,
			Attribute:     string(step.Contest.AttackerStat),
			Asset:         asset,
			OpponentAsset: target.Asset,
			World:         worldID,
			Action:        actionName,
		},
		domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: statScore(faction, step.Contest.AttackerStat)},
		registry, collector, roller, faction, factionState, rb,
	)
	defense := dispatch.RollWithHooks(
		hooks.RollContext{
			Phase:         hooks.PhaseContested,
			Actor:         target.Faction,
			Opponent:      faction,
			Attribute:     string(step.Contest.DefenderStat),
			Asset:         target.Asset,
			OpponentAsset: asset,
			World:         worldID,
			Action:        actionName,
		},
		domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: statScore(target.Faction, step.Contest.DefenderStat)},
		registry, collector, roller, target.Faction, factionState, rb,
	)
	return attack.Sum > defense.Sum, nil
}
```

##### Task 5 — `internal/faction/engine/action/actions/ability/dispatch.go`

`Dispatch` takes the registry and passes the real faction and the working asset to the
gate. The working asset carries any move an earlier step made, so the contest rolls on
the world the asset is on now.

**Find:**
```go
	collector action.Collector,
	roller domain.Roller,
	factionState *state.FactionState,
	rb *rulebook.Rulebook,
	worldEngine *world.WorldEngine,
) ([]domain.Mutation, error) {
```
**Replace with:**
```go
	collector action.Collector,
	roller domain.Roller,
	registry *hooks.Registry,
	factionState *state.FactionState,
	rb *rulebook.Rulebook,
	worldEngine *world.WorldEngine,
) ([]domain.Mutation, error) {
```

**Find:**
```go
		passed, err := passesGate(step, workingFaction, target, roller)
```
**Replace with:**
```go
		// deliberate: the real faction, not workingFaction, because a roll hook's budget spent on the clone would be lost (Effort 3 plan, Decision 12).
		passed, err := passesGate(step, faction, workingAsset, target, registry, collector, roller, factionState, rb)
```

Add `"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"` to the
import block.

##### Task 6 — `internal/faction/engine/action/actions/use_asset_ability.go`

**Find:**
```go
type UseAssetAbility struct {
	collector      action.Collector
	roller         domain.Roller
	world          *world.WorldEngine
	selectedAssets []*domain.Asset
	mutations      []domain.Mutation
}

func NewUseAssetAbility(collector action.Collector, roller domain.Roller, worldEngine *world.WorldEngine) *UseAssetAbility {
	return &UseAssetAbility{
		collector: collector,
		roller:    roller,
		world:     worldEngine,
	}
}
```
**Replace with:**
```go
type UseAssetAbility struct {
	collector      action.Collector
	roller         domain.Roller
	registry       *hooks.Registry
	world          *world.WorldEngine
	selectedAssets []*domain.Asset
	mutations      []domain.Mutation
}

func NewUseAssetAbility(collector action.Collector, roller domain.Roller, registry *hooks.Registry, worldEngine *world.WorldEngine) *UseAssetAbility {
	return &UseAssetAbility{
		collector: collector,
		roller:    roller,
		registry:  registry,
		world:     worldEngine,
	}
}
```

**Find:**
```go
		mutations, err := ability.Dispatch(faction, asset, def, u.collector, u.roller, factionState, rulebook, u.world)
```
**Replace with:**
```go
		mutations, err := ability.Dispatch(faction, asset, def, u.collector, u.roller, u.registry, factionState, rulebook, u.world)
```

Add `"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"` to the
import block.

##### Task 7 — `internal/faction/engine/action/actions/register.go`

**Find:**
```go
		return NewExpandInfluence(c, resolveRoller(), worldIndex(worldEngine), worldEngine)
	})
	ae.Register(func(c action.Collector) action.Action { return NewBribe(c) })
	ae.Register(func(c action.Collector) action.Action {
		return NewUseAssetAbility(c, resolveRoller(), worldEngine)
	})
```
**Replace with:**
```go
		return NewExpandInfluence(c, resolveRoller(), hookRegistry, worldIndex(worldEngine), worldEngine)
	})
	ae.Register(func(c action.Collector) action.Action { return NewBribe(c) })
	ae.Register(func(c action.Collector) action.Action {
		return NewUseAssetAbility(c, resolveRoller(), hookRegistry, worldEngine)
	})
```

##### Task 8 — `internal/faction/engine/effect/effects/contest.go`

Mark the S-side contest as deliberately outside the roll hooks.

**Find:**
```go
func contestPasses(contest domain.EffectContest, declaring, target *domain.Faction, roller domain.Roller) bool {
	attacker, defender := declaring, target
```
**Replace with:**
```go
func contestPasses(contest domain.EffectContest, declaring, target *domain.Faction, roller domain.Roller) bool {
	// deliberate: rolls the roller directly, not through RollWithHooks, because a reactor or phase effect has no collector to prompt (Effort 3 plan, Decision 12).
	attacker, defender := declaring, target
```

##### Task 9 — call-site updates in tests

- `expand_influence_test.go`: every `NewExpandInfluence(c, r, idx, w)` becomes
  `NewExpandInfluence(c, r, nil, idx, w)` (lines 38, 50, 60, 78, 121, 161, 210).
- `use_asset_ability_test.go`: `NewUseAssetAbility(nil, nil, nil)` becomes
  `NewUseAssetAbility(nil, nil, nil, nil)` (lines 101, 110, 119, 128).
  `NewUseAssetAbility(collector, roller, nil)` becomes `(collector, roller, nil, nil)`
  (line 80).
- `ability/dispatch_test.go`: every `Dispatch(...)` gains `nil` after the roller
  argument (lines 81, 108, 175, 207, 254, 284).

##### Task 10 — `internal/faction/engine/action/actions/expand_influence_test.go`

Add `TestExpandInfluence_RollsCarryContext`. Register a recording `RollModifier` at
global scope that appends each `RollContext` it sees and offers nothing. Run a contested
new base with one rival and one confirmed free attack. Assert the recorded contexts, in
order:

| # | Phase | Actor | Opponent | Asset | OpponentAsset | Action |
|---|---|---|---|---|---|---|
| 1 | `PhaseFactionTest` | expander | nil | nil | nil | `Expand Influence` |
| 2 | `PhaseContested` | rival | expander | nil | nil | `Expand Influence` |
| 3 | `PhaseAttack` | rival | expander | rival's attacker | nil | `Expand Influence` |
| 4 | `PhaseDefense` | expander | rival | nil | rival's attacker | `Expand Influence` |

All four carry `World == order.World`. Reuse the roll sequence of
`TestExpandInfluence_NewBase_Contested`: the hook adds no dice, so the outcome is
unchanged.

##### Task 11 — `internal/faction/engine/action/actions/ability/dispatch_test.go`

Add `TestDispatch_ContestRollsThroughHooks`. A global `RollModifier` that offers one
`AddDie(10)` (`SetKeepHighest(1)`) on `PhaseFactionTest` only, with a collector that
accepts all. Contest rolls `[3, 9, 8]`: attack keeps 9, defense 8, so the gate passes.
Without the hook, `[3, 8]` fails. Assert also that a budgeted offer spends the **real**
faction's `HookBudgets`, not the clone's.

##### Task 12 — `docs/architecture/engine/hooks.md`

- In the categories table, change the Cat 1 and Cat 2 "Dispatched from" cells to
  "`RollWithHooks` (every faction roll)".
- In **Registration in, dispatch out**, change "`RollWithHooks`** wraps every combat
  roll ([attack action](actions.md))" to "wraps every faction roll: attack and defense,
  the Expand Influence test, rival rolls and free attack, and the ability contest". Add
  one sentence: S-side contests (`effects/contest.go`) roll directly, because a reactor
  has no collector.
- In **Dependencies**, change "the [actions](actions.md) engine (Cat 1/2/5 in combat…"
  to "(Cat 1/2 on every faction roll, Cat 5 in combat…".

##### Commit message

```
feat: route every roll through RollWithHooks

- Expand Influence: faction test, rival rolls and free attack use the hooks
- the ability contest rolls through the hooks against the real faction
- RollContext.Action names the action for every roll
- Expand Influence and UseAssetAbility take the hook registry
- fence the S-side contest as a direct roll
```

### Phase 2 — Roll effects

Each commit is one registrar on Phase 1's dispatcher, plus its authored assets.

#### Commit 3 — `feat: bonus_die effect`

##### Task 1 — `internal/faction/domain/effect.go`

`phase` becomes a list (Decision 17).

**Find:**
```go
type EffectFilters struct {
	Owner      EffectOwner
	Scope      EffectScope
	Attribute  FactionStat
	Phase      EffectRollPhase
	Action     string
```
**Replace with:**
```go
type EffectFilters struct {
	Owner      EffectOwner
	Scope      EffectScope
	Attribute  FactionStat
	Phases     []EffectRollPhase
	Action     string
```

**Find:**
```go
	if filters.Phase != "" {
		present = append(present, FilterPhase)
	}
```
**Replace with:**
```go
	if len(filters.Phases) > 0 {
		present = append(present, FilterPhase)
	}
```

##### Task 2 — `internal/faction/rulebook/effects.go`

**Find:**
```go
	Phase     string   `toml:"phase"`
```
**Replace with:**
```go
	Phase     []string `toml:"phase"`
```

**Find:**
```go
	if record.Phase != "" {
		phase, err := toEnum("roll phase", record.Phase, effectRollPhases)
		if err != nil {
			return domain.EffectFilters{}, err
		}
		filters.Phase = phase
	}
```
**Replace with:**
```go
	for i, value := range record.Phase {
		phase, err := toEnum("roll phase", value, effectRollPhases)
		if err != nil {
			return domain.EffectFilters{}, fmt.Errorf("filters.phase[%d]: %w", i, err)
		}
		filters.Phases = append(filters.Phases, phase)
	}
```

##### Task 3 — loader and domain test fixtures

- `rulebook/effects_test.go:59`: `phase      = "defense"` becomes
  `phase      = ["defense"]`.
- `rulebook/effects_test.go:114` and `domain/effect_test.go:28`:
  `Phase:      domain.RollPhaseDefense,` becomes
  `Phases:     []domain.EffectRollPhase{domain.RollPhaseDefense},`.
- Add a case to the loader's loud-failure table: `phase = ["defence"]` fails with an
  error containing `filters.phase[0]`.

##### Task 4 — `internal/faction/engine/effect/effects/active.go`

Roll hooks keep IDs and look their asset up at roll time, as reactors do. Append:

```go
// liveAsset returns the declaring asset from current state, or nil when its
// faction or the asset is gone. A nil state (a dispatcher test) is no asset.
func liveAsset(factionState *state.FactionState, factionID, assetID string) *domain.Asset {
	if factionState == nil {
		return nil
	}
	faction := factionState.Factions[factionID]
	if faction == nil {
		return nil
	}
	return faction.Assets[assetID]
}
```

Add `"github.com/therobertcrocker/gm-toolkit/internal/faction/state"` to the imports.

##### Task 5 — `internal/faction/engine/effect/effects/roll_match.go` (new)

The shared reading of a roll trigger and its filters, used by Commits 3, 4 and 5.

Full contents:
```go
package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

var rollPhaseNames = map[hooks.RollPhase]domain.EffectRollPhase{
	hooks.PhaseAttack:      domain.RollPhaseAttack,
	hooks.PhaseDefense:     domain.RollPhaseDefense,
	hooks.PhaseFactionTest: domain.RollPhaseFactionTest,
	hooks.PhaseContested:   domain.RollPhaseContested,
}

// rollMatches reports whether a roll fires an entry declared on asset. own_roll
// fires on the owner's own roll; enemy_roll_at_me on a rival's roll made against
// the owner. scope = "self" ties the roll to the declaring asset (rolling it for
// own_roll, rolled against for enemy_roll_at_me); scope = "world" to any such roll
// on the asset's world. The dispatcher reaches both sides' assets, so this check,
// not the registration scope, is what keeps an entry to its own side.
func rollMatches(ctx hooks.RollContext, trigger domain.EffectTrigger, ownerFactionID string, asset *domain.Asset, filters domain.EffectFilters) bool {
	if ctx.Actor == nil {
		return false
	}
	var subject *domain.Asset
	switch trigger {
	case domain.TriggerOwnRoll:
		if ctx.Actor.ID != ownerFactionID {
			return false
		}
		subject = ctx.Asset
	case domain.TriggerEnemyRollAtMe:
		if ctx.Actor.ID == ownerFactionID || ctx.Opponent == nil || ctx.Opponent.ID != ownerFactionID {
			return false
		}
		subject = ctx.OpponentAsset
	default:
		return false
	}
	switch filters.Scope {
	case domain.ScopeSelf:
		if subject == nil || subject.ID != asset.ID {
			return false
		}
	case domain.ScopeWorld:
		if ctx.World == "" || ctx.World != asset.Location.WorldID {
			return false
		}
	default:
		return false
	}
	if filters.Attribute != "" && ctx.Attribute != string(filters.Attribute) {
		return false
	}
	if len(filters.Phases) > 0 && !slices.Contains(filters.Phases, rollPhaseNames[ctx.Phase]) {
		return false
	}
	if filters.Action != "" && ctx.Action != filters.Action {
		return false
	}
	return true
}
```

##### Task 6 — `internal/faction/engine/effect/effects/bonus_die.go` (new)

Full contents:
```go
package effects

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// bonusDieRegistrar reads owner = "own" as "the owner is rolling". scope =
// "self" is a roll by the declaring asset; "world" any owner roll on its world.
var bonusDieRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnRoll: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeSelf, domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterAttribute, domain.FilterPhase, domain.FilterAction},
		},
	},
	Register: registerBonusDie,
}

func registerBonusDie(ctx effect.RegisterContext) {
	ctx.Registry.RegisterRollModifier(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialBonusDie)+":"+ctx.Def.ID,
		&bonusDieModifier{
			ownerFactionID: ctx.Faction.ID,
			assetID:        ctx.Asset.ID,
			source:         ctx.Def.Name,
			trigger:        ctx.Entry.Trigger,
			filters:        ctx.Entry.Filters,
			params:         *ctx.Entry.BonusDie,
		},
	)
}

type bonusDieModifier struct {
	ownerFactionID string
	assetID        string
	source         string
	trigger        domain.EffectTrigger
	filters        domain.EffectFilters
	params         domain.BonusDieParams
}

func (modifier *bonusDieModifier) OfferModifiers(ctx hooks.RollContext, factionState *state.FactionState, _ *rulebook.Rulebook) []hooks.ModifierOffer {
	asset := liveAsset(factionState, modifier.ownerFactionID, modifier.assetID)
	if !active(asset) || !rollMatches(ctx, modifier.trigger, modifier.ownerFactionID, asset, modifier.filters) {
		return nil
	}
	params := modifier.params
	return []hooks.ModifierOffer{{
		Source:      modifier.source,
		Description: fmt.Sprintf("roll %d extra d%d, keep highest", params.Count, params.Sides),
		Apply: func(rollState *hooks.RollState) {
			for range params.Count {
				rollState.AddDie(params.Sides)
			}
			// deliberate: keep 1, not 1 + count, because every faction roll is one d10 and an extra die is pick-the-best ("an additional d10, keep the highest").
			rollState.SetKeepHighest(1)
		},
	}}
}
```

##### Task 7 — `internal/faction/engine/effect/effects/registrars.go`

**Find:**
```go
		domain.SpecialClearStealth:   clearStealthRegistrar,
	}
```
**Replace with:**
```go
		domain.SpecialClearStealth:   clearStealthRegistrar,
		domain.SpecialBonusDie:       bonusDieRegistrar,
	}
```

##### Task 8 — asset data (both `rulebooks/swn/assets/` and `campaigns/test-camp/rulebook/assets/`)

`force_assets.toml` — after the `[assets.IntegralProtocols]` block's `description` line:
```toml

[[assets.IntegralProtocols.effects]]
trigger = "own_roll"
effect  = "bonus_die"
[assets.IntegralProtocols.effects.filters]
owner     = "own"
scope     = "self"
phase     = ["defense"]
attribute = "Cunning"
[assets.IntegralProtocols.effects.bonus_die]
sides = 10
count = 1
```

`wealth_assets.toml` — in `[assets.Surveyors]`, delete the misleading move clause
(Planning Decision 15):

**Find:**
```toml
description = "Scouts exploring potential resource and investment options on worlds. The presence of Surveyors grants one additional die on Expand Influence actions on that world. As an action, a Surveyor crew can move to any world within two hexes."
```
**Replace with:**
```toml
description = "Scouts exploring potential resource and investment options on worlds. The presence of Surveyors grants one additional die on Expand Influence actions on that world."

[[assets.Surveyors.effects]]
trigger = "own_roll"
effect  = "bonus_die"
[assets.Surveyors.effects.filters]
owner  = "own"
scope  = "world"
phase  = ["faction_test"]
action = "Expand Influence"
[assets.Surveyors.effects.bonus_die]
sides = 10
count = 1
```

`cunning_assets.toml` — after the `[assets.PanopticonMatrix]` block's `description` line:
```toml

[[assets.PanopticonMatrix.effects]]
trigger = "own_roll"
effect  = "bonus_die"
[assets.PanopticonMatrix.effects.filters]
owner     = "own"
scope     = "world"
phase     = ["attack", "defense"]
attribute = "Cunning"
[assets.PanopticonMatrix.effects.bonus_die]
sides = 10
count = 1
```

##### Task 9 — `internal/faction/engine/effect/effects/roll_match_test.go` (new)

The shared matcher is the place every filter's suppression test lives (overview Shared
Context). It is package-internal, so use `package effects`. One table, alpha owns
asset `x` on Krylos, beta is the rival:

| Case | Trigger | Filters | Context | Want |
|---|---|---|---|---|
| own roll, self | own_roll | self | Actor alpha, Asset x | true |
| own roll, other asset rolls | own_roll | self | Actor alpha, Asset y | false |
| own roll, rival rolling | own_roll | world | Actor beta, World Krylos | false |
| own roll, world | own_roll | world | Actor alpha, World Krylos | true |
| own roll, other world | own_roll | world | Actor alpha, World Tartarus | false |
| own roll, in flight | own_roll | world | Actor alpha, World "" | false |
| enemy at me, self | enemy_roll_at_me | self | Actor beta, Opponent alpha, OpponentAsset x | true |
| enemy at me, against other asset | enemy_roll_at_me | self | Actor beta, Opponent alpha, OpponentAsset y | false |
| enemy, not against owner | enemy_roll_at_me | world | Actor beta, Opponent gamma, World Krylos | false |
| enemy, no opponent | enemy_roll_at_me | world | Actor beta, Opponent nil, World Krylos | false |
| owner's own roll is not enemy | enemy_roll_at_me | world | Actor alpha, Opponent beta, World Krylos | false |
| attribute match / mismatch | own_roll | self + Cunning | Attribute Cunning / Force | true / false |
| phase in list / not in list | own_roll | self + [attack, defense] | PhaseDefense / PhaseContested | true / false |
| action match / mismatch | own_roll | world + Expand Influence | Action Expand Influence / Attack | true / false |
| unknown trigger | turn_start | self | Actor alpha, Asset x | false |

##### Task 10 — `internal/faction/engine/effect/effects/bonus_die_test.go` (new)

`package effects_test`. Register through `effects.Registrars()[domain.SpecialBonusDie]`
and read the hook back with `registry.RollModifiersAt(hooks.AssetScope(...))`.

- **Matching roll offers one die with keep-highest.** Integral Protocols' entry, its
  own Cunning defense roll. One offer whose `Source` is the definition name. Applying
  it to a `hooks.RollState` gives `ExtraDice() == [10]` and `KeepHighest() == 1`.
- **Count** 2 queues two dice.
- **Non-matching roll offers nothing.** A Force defense (attribute suppression).
- **Inactive asset offers nothing.** Over `inactiveStates`.
- **End to end through the dispatcher.** Panopticon on Krylos (alpha). Alpha attacks
  beta on Krylos with Cunning. Run `dispatch.RollWithHooks` with an accept-all collector
  and `seqRoller{[3, 9]}`. Sum is `9 + stat`. Beta's defense against alpha on the same
  world gets no offer (owner check via the opponent bucket).

##### Task 11 — `internal/faction/engine/testharness/scenarios/effects_test.go`

The harness loads `rulebooks/swn`, so construction already validates the three new
entries under rule 4. Add `TestEffects_SurveyorsAddDieToOwnersExpansion`: alpha owns
Surveyors on world W and expands there with one rival present. The scripted collector
accepts all offers, and a fixed roller is arranged so the faction roll's extra die
decides the contest. Assert no free attack is prompted. Follow the existing scenario
file's setup helpers.

##### Commit message

```
feat: bonus_die effect

- phase filter becomes a list in the domain and the loader
- shared roll matcher reads own_roll / enemy_roll_at_me and the roll filters
- bonus_die registrar as an asset-scoped RollModifier, keep-highest one
- author Integral Protocols, Surveyors and Panopticon Matrix's die
- drop the misleading move clause from Surveyors' description
```

#### Commit 4 — `feat: remove_tag_dice effect`

##### Task 1 — `internal/faction/engine/hooks/interfaces.go`

Append after `RollModifier`:

```go
// RollOfferFilter removes Cat 1 offers before the collector sees them (Cat 1).
// It runs after every RollModifier has offered, so it can strike any offer —
// including one another faction's hook made.
type RollOfferFilter interface {
	FilterOffers(ctx RollContext, offers []ModifierOffer, factionState *state.FactionState) []ModifierOffer
}
```

##### Task 2 — `internal/faction/engine/hooks/types.go`

**Find:**
```go
	Apply       func(*RollState) // expands the dice pool before rolling
	FactionID   string
}
```
**Replace with:**
```go
	Apply       func(*RollState) // expands the dice pool before rolling
	FactionID   string
	FromTag     bool // a tag's bonus die; Blackmail strips these
}
```

##### Task 3 — `internal/faction/engine/hooks/registry.go`

Add the family in the same shape as the others:

- after `RegisteredRollResultHook`:
  ```go
  type RegisteredRollOfferFilter struct {
  	Scope  Scope
  	Source string
  	Hook   RollOfferFilter
  }
  ```
- `Registry` gains `rollOfferFilters map[Scope][]RegisteredRollOfferFilter`, and
  `NewRegistry` initialises it with `make(map[Scope][]RegisteredRollOfferFilter)`.
- after `RegisterRollResultHook`:
  ```go
  // RegisterRollOfferFilter registers a Cat 1 offer filter under the given scope.
  func (registry *Registry) RegisterRollOfferFilter(scope Scope, source string, hook RollOfferFilter) {
  	registry.rollOfferFilters[scope] = append(registry.rollOfferFilters[scope], RegisteredRollOfferFilter{Scope: scope, Source: source, Hook: hook})
  }
  ```
- after `RollResultHooksAt`:
  ```go
  // RollOfferFiltersAt returns the Cat 1 offer filters registered in one scope bucket.
  func (registry *Registry) RollOfferFiltersAt(scope Scope) []RegisteredRollOfferFilter {
  	return registry.rollOfferFilters[scope]
  }
  ```

##### Task 4 — `internal/faction/engine/hooks/dispatch/roll.go`

Run the filters over the gathered offers, before the budget filter.

**Find:**
```go
		offers := offersBudgetFilter(allOffers, faction, factionState)
```
**Replace with:**
```go
		for _, scope := range scopes {
			for _, registered := range registry.RollOfferFiltersAt(scope) {
				allOffers = registered.Hook.FilterOffers(ctx, allOffers, factionState)
			}
		}
		offers := offersBudgetFilter(allOffers, faction, factionState)
```

##### Task 5 — `internal/faction/engine/tag/tags/warlike.go`

**Find:**
```go
		Source:      "tag:Warlike",
		Description: "roll an extra d10, keep highest",
		BudgetKey:   "tag:Warlike",
```
**Replace with:**
```go
		Source:      "tag:Warlike",
		Description: "roll an extra d10, keep highest",
		BudgetKey:   "tag:Warlike",
		FromTag:     true,
```

##### Task 6 — `internal/faction/engine/effect/effects/remove_tag_dice.go` (new)

Full contents:
```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// removeTagDiceRegistrar reads owner = "rival" as "a rival is rolling against the
// owner", and scope = "self" as "against the declaring asset itself" — attacking
// it, or defending against it.
var removeTagDiceRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerEnemyRollAtMe: {
			Owners: []domain.EffectOwner{domain.OwnerRival},
			Scopes: []domain.EffectScope{domain.ScopeSelf},
		},
	},
	Register: registerRemoveTagDice,
}

func registerRemoveTagDice(ctx effect.RegisterContext) {
	ctx.Registry.RegisterRollOfferFilter(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialRemoveTagDice)+":"+ctx.Def.ID,
		&tagDiceFilter{
			ownerFactionID: ctx.Faction.ID,
			assetID:        ctx.Asset.ID,
			trigger:        ctx.Entry.Trigger,
			filters:        ctx.Entry.Filters,
		},
	)
}

type tagDiceFilter struct {
	ownerFactionID string
	assetID        string
	trigger        domain.EffectTrigger
	filters        domain.EffectFilters
}

func (filter *tagDiceFilter) FilterOffers(ctx hooks.RollContext, offers []hooks.ModifierOffer, factionState *state.FactionState) []hooks.ModifierOffer {
	asset := liveAsset(factionState, filter.ownerFactionID, filter.assetID)
	if !active(asset) || !rollMatches(ctx, filter.trigger, filter.ownerFactionID, asset, filter.filters) {
		return offers
	}
	var kept []hooks.ModifierOffer
	for _, offer := range offers {
		if !offer.FromTag {
			kept = append(kept, offer)
		}
	}
	return kept
}
```

##### Task 7 — `internal/faction/engine/effect/effects/registrars.go`

**Find:**
```go
		domain.SpecialBonusDie:       bonusDieRegistrar,
	}
```
**Replace with:**
```go
		domain.SpecialBonusDie:       bonusDieRegistrar,
		domain.SpecialRemoveTagDice:  removeTagDiceRegistrar,
	}
```

##### Task 8 — asset data (both rulebooks)

`cunning_assets.toml` — after the `[assets.Blackmail.attack]` block (after
`damage = "1d4+1"`):
```toml

[[assets.Blackmail.effects]]
trigger = "enemy_roll_at_me"
effect  = "remove_tag_dice"
[assets.Blackmail.effects.filters]
owner = "rival"
scope = "self"
```

##### Task 9 — tests

- `hooks/registry_test.go`: `TestRollOfferFilter_RegistrationOrderPreserved` and
  `_EmptyRegistryReturnsEmptySlice`, matching the Cat 1 pair.
- `hooks/dispatch/roll_test.go`: `TestRollWithHooks_OfferFilterRemovesOffers`. A
  filter at the opponent's asset scope strikes the offer from a roller's global
  modifier. The roll has 2 dice and the collector is not called.
- `tag/tags/warlike_test.go`: the offer carries `FromTag == true`.
- `effect/effects/remove_tag_dice_test.go` (new, `package effects_test`):
  - Blackmail (alpha `bm` on Krylos) is attacked by beta. `FilterOffers` on beta's roll
    with `OpponentAsset: bm` drops a `FromTag` offer and keeps a non-tag offer.
  - Blackmail attacks beta. Beta's defense roll (`OpponentAsset: bm`) is stripped.
  - Beta's roll against another alpha asset is untouched (self-scope suppression).
  - Alpha's own roll with Blackmail rolling is untouched (owner suppression).
  - Inactive Blackmail strips nothing, over `inactiveStates`.
  - End to end: beta is Warlike (register `tags.WarlikeHandler{}` for beta), attacking
    Blackmail with a Force attack. `RollWithHooks` rolls one die, and the collector is
    not called.

##### Task 10 — `internal/faction/engine/hooks/doc.go` and `docs/architecture/engine/hooks.md`

- `doc.go`, Cat 1 line: `Cat 1  RollModifier / RollOfferFilter  pre-roll dice augmentation and offer removal  interfaces.go`.
  Realign the outline columns.
- `hooks.md` categories table, Cat 1 row: Interface `RollModifier`, `RollOfferFilter`;
  Fires "before a roll — offers extra dice; filters may strike offers".
- `hooks.md` **Registration in, dispatch out**, `RollWithHooks` bullet: after "gathers
  Cat 1 offers", add "runs Cat 1 offer filters over them (Blackmail strips tag dice —
  `ModifierOffer.FromTag`)".

##### Commit message

```
feat: remove_tag_dice effect

- RollOfferFilter: a Cat 1 hook that strikes offers before the collector
- ModifierOffer.FromTag, set by Warlike
- remove_tag_dice registrar on enemy_roll_at_me against the declaring asset
- author Blackmail
```

#### Commit 5 — `feat: reroll effect`

##### Task 1 — `internal/faction/engine/effect/effects/budget.go`

The key gains the trigger (Decision 16).

**Find:**
```go
// budgetKey names the once_per budget an entry spends. The effect name and the
// definition ID keep two effects, or two asset kinds, from sharing a budget.
func budgetKey(oncePer domain.OncePerKey, effectName domain.SpecialEffectType, defID, assetID, targetFactionID string) string {
	base := "effect:" + string(effectName) + ":" + defID
```
**Replace with:**
```go
// budgetKey names the once_per budget an entry spends. The effect name, trigger
// and definition ID keep two effects, two entries of one definition, or two asset
// kinds from sharing a budget.
func budgetKey(oncePer domain.OncePerKey, effectName domain.SpecialEffectType, trigger domain.EffectTrigger, defID, assetID, targetFactionID string) string {
	base := "effect:" + string(effectName) + ":" + string(trigger) + ":" + defID
```

##### Task 2 — `internal/faction/engine/effect/effects/coin_steal.go`

The reactor stores its trigger. (*Superseded by the overview plan's Decision Record
26:* the reactor now reads its entries from the rulebook and keeps no trigger.)

**Find:**
```go
			defID:          ctx.Def.ID,
			amount:         ctx.Entry.CoinSteal.Amount,
```
**Replace with:**
```go
			defID:          ctx.Def.ID,
			trigger:        ctx.Entry.Trigger,
			amount:         ctx.Entry.CoinSteal.Amount,
```

**Find:**
```go
type coinStealReactor struct {
	ownerFactionID string
	assetID        string
	defID          string
	amount         domain.DiceRoll
```
**Replace with:**
```go
type coinStealReactor struct {
	ownerFactionID string
	assetID        string
	defID          string
	trigger        domain.EffectTrigger
	amount         domain.DiceRoll
```

**Find:**
```go
budgetKey(reactor.oncePer, domain.SpecialCoinSteal, reactor.defID, reactor.assetID, targetID)
```
**Replace with:**
```go
budgetKey(reactor.oncePer, domain.SpecialCoinSteal, reactor.trigger, reactor.defID, reactor.assetID, targetID)
```

##### Task 3 — `internal/faction/engine/effect/effects/ignore_coin_loss.go`

**Find:**
```go
			defID:          ctx.Def.ID,
			oncePer:        ctx.Entry.Gates.OncePer,
		},
```
**Replace with:**
```go
			defID:          ctx.Def.ID,
			trigger:        ctx.Entry.Trigger,
			oncePer:        ctx.Entry.Gates.OncePer,
		},
```

**Find:**
```go
type ignoreCoinLossReactor struct {
	ownerFactionID string
	assetID        string
	defID          string
	oncePer        domain.OncePerKey
}
```
**Replace with:**
```go
type ignoreCoinLossReactor struct {
	ownerFactionID string
	assetID        string
	defID          string
	trigger        domain.EffectTrigger
	oncePer        domain.OncePerKey
}
```

**Find:**
```go
budgetKey(reactor.oncePer, domain.SpecialIgnoreCoinLoss, reactor.defID, copies[next].ID, loss.CausedByFactionID)
```
**Replace with:**
```go
budgetKey(reactor.oncePer, domain.SpecialIgnoreCoinLoss, reactor.trigger, reactor.defID, copies[next].ID, loss.CausedByFactionID)
```

##### Task 4 — `internal/faction/engine/effect/effects/budget_test.go`

`TestBudgetKey` passes `domain.TriggerOwnAttackHit`, and its wants become
`effect:coin_steal:own_attack_hit:DEF:beta`, `effect:coin_steal:own_attack_hit:DEF`,
`effect:coin_steal:own_attack_hit:DEF:asset-1`. The `OwnerFactionSharesOneBudget` test
passes the same trigger to both calls. Add `TestBudgetKey_TriggersDoNotShare`:
`own_roll` and `enemy_roll_at_me` on one `owner_faction` definition produce different
keys, and spending one leaves the other unspent. Any Effort 2 test that asserts a
literal budget key in `HookBudgets` (grep `effect:` in `effects/*_test.go`) gains the
trigger segment.

##### Task 5 — `internal/faction/engine/effect/effects/reroll.go` (new)

Full contents:
```go
package effects

import (
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// rerollRegistrar reads own_roll as the owner rerolling its own dice, and
// enemy_roll_at_me as the owner forcing a rival rolling against it to reroll.
// owner and scope read as in rollMatches.
var rerollRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnRoll: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf, domain.ScopeWorld},
			Gates:  []domain.EffectGateName{domain.GateOncePer},
		},
		domain.TriggerEnemyRollAtMe: {
			Owners: []domain.EffectOwner{domain.OwnerRival},
			Scopes: []domain.EffectScope{domain.ScopeSelf, domain.ScopeWorld},
			Gates:  []domain.EffectGateName{domain.GateOncePer},
		},
	},
	Register: registerReroll,
}

func registerReroll(ctx effect.RegisterContext) {
	source := ctx.Def.Name
	if ctx.Entry.Trigger == domain.TriggerEnemyRollAtMe {
		source += " (forced)"
	}
	ctx.Registry.RegisterRollResultHook(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialReroll)+":"+ctx.Def.ID,
		&rerollHook{
			ownerFactionID: ctx.Faction.ID,
			assetID:        ctx.Asset.ID,
			defID:          ctx.Def.ID,
			source:         source,
			trigger:        ctx.Entry.Trigger,
			filters:        ctx.Entry.Filters,
			oncePer:        ctx.Entry.Gates.OncePer,
			params:         *ctx.Entry.Reroll,
		},
	)
}

type rerollHook struct {
	ownerFactionID string
	assetID        string
	defID          string
	source         string
	trigger        domain.EffectTrigger
	filters        domain.EffectFilters
	oncePer        domain.OncePerKey
	params         domain.RerollParams
}

func (hook *rerollHook) OnRollResult(ctx hooks.RollContext, result hooks.RollResult, factionState *state.FactionState, _ *rulebook.Rulebook) hooks.RerollDirective {
	asset := liveAsset(factionState, hook.ownerFactionID, hook.assetID)
	if !active(asset) || !rollMatches(ctx, hook.trigger, hook.ownerFactionID, asset, hook.filters) {
		return hooks.RerollDirective{}
	}
	forced := hook.trigger == domain.TriggerEnemyRollAtMe
	directive := hooks.RerollDirective{
		Source:   hook.source,
		Indices:  pickDice(result.Dice, hook.params.Dice, forced),
		Elective: hook.params.Elective,
	}
	if hook.oncePer != "" {
		directive.BudgetKey = budgetKey(hook.oncePer, domain.SpecialReroll, hook.trigger, hook.defID, hook.assetID, otherSideID(ctx, forced))
	}
	return directive
}

// pickDice returns the indices of the n dice a reroll targets: the owner's own
// lowest dice, or a forced rival's highest.
func pickDice(dice []int, n int, highest bool) []int {
	indices := make([]int, len(dice))
	for i := range dice {
		indices[i] = i
	}
	sort.SliceStable(indices, func(a, b int) bool {
		if highest {
			return dice[indices[a]] > dice[indices[b]]
		}
		return dice[indices[a]] < dice[indices[b]]
	})
	return indices[:min(n, len(indices))]
}

// otherSideID names the faction across the roll from the owner, for a
// target_faction budget: the opponent of the owner's own roll, or the forced roller.
func otherSideID(ctx hooks.RollContext, forced bool) string {
	if forced {
		return ctx.Actor.ID
	}
	if ctx.Opponent == nil {
		return ""
	}
	return ctx.Opponent.ID
}
```

##### Task 6 — `internal/faction/engine/effect/effects/registrars.go`

**Find:**
```go
		domain.SpecialRemoveTagDice:  removeTagDiceRegistrar,
	}
```
**Replace with:**
```go
		domain.SpecialRemoveTagDice:  removeTagDiceRegistrar,
		domain.SpecialReroll:         rerollRegistrar,
	}
```

##### Task 7 — asset data (both rulebooks)

`cunning_assets.toml` — after the `[assets.BookOfSecrets]` block's `description` line:
```toml

[[assets.BookOfSecrets.effects]]
trigger = "own_roll"
effect  = "reroll"
[assets.BookOfSecrets.effects.filters]
owner = "own"
scope = "world"
[assets.BookOfSecrets.effects.gates]
once_per = "owner_faction"
[assets.BookOfSecrets.effects.reroll]
dice     = 1
elective = true

[[assets.BookOfSecrets.effects]]
trigger = "enemy_roll_at_me"
effect  = "reroll"
[assets.BookOfSecrets.effects.filters]
owner = "rival"
scope = "world"
[assets.BookOfSecrets.effects.gates]
once_per = "owner_faction"
[assets.BookOfSecrets.effects.reroll]
dice     = 1
elective = true
```

##### Task 8 — `internal/faction/engine/effect/effects/reroll_test.go` (new)

`package effects_test`. Alpha owns Book of Secrets `bos` on Krylos. Beta is the rival.
Most cases run through `dispatch.RollWithHooks` with a real `state.FactionState`, so the
budget path is proven where it lives.

- **Own reroll picks the lowest die.** `OnRollResult` on alpha's roll on Krylos with
  dice `[2, 9]` returns `Indices == [0]`, `Elective == true`, and `BudgetKey ==
  "effect:reroll:own_roll:SWN-C7-002"`.
- **Forced reroll picks the highest die.** On beta's roll against alpha on Krylos,
  `[2, 9]` gives `[1]`, `Source` ending `(forced)`, and key
  `effect:reroll:enemy_roll_at_me:SWN-C7-002`.
- **Forced reroll spends alpha's budget, not beta's.** Through the dispatcher with
  accept-all, beta's roll `[9]` is rerolled to `[3]`. Then `alpha.HookBudgets[key] ==
  1`, beta's map is nil, and a second beta roll is not rerolled.
- **The two budgets are separate.** After the forced reroll is spent, alpha's own roll
  is still offered a reroll.
- **Declined reroll spends nothing.** A collector that declines leaves the budget
  unspent.
- **Suppression.** Beta's roll against gamma on Krylos gets no forced directive. Alpha's
  roll on Tartarus gets no own directive. An inactive Book (over `inactiveStates`) gets
  none.
- **`once_per` gate absent:** an entry with no gate yields an empty `BudgetKey`.

##### Commit message

```
feat: reroll effect

- budget key gains the trigger so two entries of one definition do not share
- reroll registrar as an asset-scoped RollResultHook
- own roll rerolls the lowest die, a forced reroll the highest
- author Book of Secrets with separate own and forced budgets
```

### Phase 3 — Turn-start stealth test

#### Commit 6 — `feat: clear_stealth on turn start`

##### Task 1 — `internal/faction/engine/effect/registrar.go`

**Find:**
```go
type FireContext struct {
	Faction *domain.Faction
	Asset   *domain.Asset
	Def     *domain.AssetDefinition
	Entry   domain.EffectDefinition
	Roller  domain.Roller
}
```
**Replace with:**
```go
// FireContext carries FactionState because a phase effect may reach other
// factions' assets (Panopticon Matrix tests every rival on its world).
type FireContext struct {
	Faction      *domain.Faction
	Asset        *domain.Asset
	Def          *domain.AssetDefinition
	Entry        domain.EffectDefinition
	Roller       domain.Roller
	FactionState *state.FactionState
}
```

Add `"github.com/therobertcrocker/gm-toolkit/internal/faction/state"` to the imports.

##### Task 2 — `internal/faction/engine/effect/effect.go`

**Find:**
```go
func (e *EffectsEngine) FireTurnStart(faction *domain.Faction, rulebook *rulebook.Rulebook, roller domain.Roller) []domain.Mutation {
```
**Replace with:**
```go
func (e *EffectsEngine) FireTurnStart(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook, roller domain.Roller) []domain.Mutation {
```

**Find:**
```go
			mutations = append(mutations, e.registrars[entry.Effect].Fire(FireContext{
				Faction: faction,
				Asset:   asset,
				Def:     def,
				Entry:   entry,
				Roller:  roller,
			})...)
```
**Replace with:**
```go
			mutations = append(mutations, e.registrars[entry.Effect].Fire(FireContext{
				Faction:      faction,
				Asset:        asset,
				Def:          def,
				Entry:        entry,
				Roller:       roller,
				FactionState: factionState,
			})...)
```

##### Task 3 — `internal/faction/engine/orchestrator.go`

**Find:**
```go
		startMutations = e.Effect.FireTurnStart(faction, e.Rulebook, e.Rand)
```
**Replace with:**
```go
		startMutations = e.Effect.FireTurnStart(faction, factionState, e.Rulebook, e.Rand)
```

##### Task 4 — `internal/faction/engine/effect/effects/clear_stealth.go`

The registrar gains the turn-start trigger and its phase route.

**Find:**
```go
// clearStealthRegistrar reads owner = "rival" as "the arriving asset is a
// rival's", and scope = "world" as "it lands on the declaring asset's world".
// Arrival by purchase is Effort 5's, with the route that makes a purchase stealthy.
var clearStealthRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerStealthArrival: {
			Owners:  []domain.EffectOwner{domain.OwnerRival},
			Scopes:  []domain.EffectScope{domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterStealthy},
			Gates:   []domain.EffectGateName{domain.GateContest},
		},
	},
	Register: registerClearStealth,
}
```
**Replace with:**
```go
// clearStealthRegistrar reads owner = "rival" as "the affected asset is a
// rival's", and scope = "world" as "it stands on (stealth_arrival: lands on) the
// declaring asset's world". stealth_arrival is a reactor; turn_start tests every
// such asset when the owner's turn begins. Arrival by purchase is Effort 5's, with
// the route that makes a purchase stealthy.
var clearStealthRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerStealthArrival: {
			Owners:  []domain.EffectOwner{domain.OwnerRival},
			Scopes:  []domain.EffectScope{domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterStealthy},
			Gates:   []domain.EffectGateName{domain.GateContest},
		},
		domain.TriggerTurnStart: {
			Owners:  []domain.EffectOwner{domain.OwnerRival},
			Scopes:  []domain.EffectScope{domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterStealthy},
			Gates:   []domain.EffectGateName{domain.GateContest},
		},
	},
	Register: registerClearStealth,
	Fire:     fireClearStealth,
}

// fireClearStealth tests every rival asset on the declaring asset's world, in
// faction then asset ID order.
func fireClearStealth(ctx effect.FireContext) []domain.Mutation {
	worldID := ctx.Asset.Location.WorldID
	if worldID == "" {
		return nil
	}
	var cleared []domain.Mutation
	for _, factionID := range slices.Sorted(maps.Keys(ctx.FactionState.Factions)) {
		if factionID == ctx.Faction.ID {
			continue
		}
		target := ctx.FactionState.Factions[factionID]
		for _, asset := range domain.SortedAssets(target) {
			if asset.Location.WorldID != worldID {
				continue
			}
			if ctx.Entry.Filters.Stealthy != nil && asset.Stealthy != *ctx.Entry.Filters.Stealthy {
				continue
			}
			if ctx.Entry.Gates.Contest != nil && !contestPasses(*ctx.Entry.Gates.Contest, ctx.Faction, target, ctx.Roller) {
				continue
			}
			cleared = append(cleared, domain.AssetStealthCleared{
				FactionID:         target.ID,
				AssetID:           asset.ID,
				Cause:             string(domain.SpecialClearStealth),
				CausedByFactionID: ctx.Faction.ID,
			})
		}
	}
	return cleared
}
```

Add `"maps"` and `"slices"` to the imports.

##### Task 5 — asset data (both rulebooks)

`cunning_assets.toml` — after Panopticon Matrix's `bonus_die` entry (Commit 3 Task 8):
```toml

[[assets.PanopticonMatrix.effects]]
trigger = "turn_start"
effect  = "clear_stealth"
[assets.PanopticonMatrix.effects.filters]
owner    = "rival"
scope    = "world"
stealthy = true
[assets.PanopticonMatrix.effects.gates.contest]
roller        = "target"
attacker_stat = "Cunning"
defender_stat = "Cunning"
```

##### Task 6 — tests

- `effect/effect_test.go:361`: `FireTurnStart(faction, rb, nil)` becomes
  `FireTurnStart(faction, nil, rb, nil)`. Add an assertion through a test-local
  registrar that `ctx.FactionState` is the state passed in.
- `effect/effects/clear_stealth_test.go`: add `TestClearStealth_TurnStart`, reusing
  `tripwireState` with alpha's Matrix on Tartarus. Call
  `effects.Registrars()[domain.SpecialClearStealth].Fire(...)` with `roller = "target"`.
  Cases:
  - **Beta's spy fails its test and is unmasked.** Beta rolls low, alpha high:
    `[2, 9]` gives `spyUnmasked`.
  - **Beta's spy passes and keeps stealth.** `[9, 2]` gives nothing.
  - **A tie unmasks.** With `roller = "target"` the rival attacks and the Matrix
    defends, and ties go to the defender, so the rival's test fails
    (`contestPasses`' documented rule). Alpha Cunning 3, beta Cunning 2: `[3, 2]`
    gives `spyUnmasked`.
  - **Beta's open courier is not tested.** It is stealthy-filter suppression. Only one
    contest is rolled, proven by a `seqRoller` that panics past two values.
  - **Alpha's own stealthed agent is not tested.** Owner suppression.
  - **A rival on another world is not tested.**
  - **A Matrix in flight tests nothing.**
  - **No contest gate** unmasks unconditionally.
- `testharness/scenarios/effects_test.go`: add
  `TestEffects_PanopticonUnmasksAtOwnersTurnStart`. Alpha's active Matrix shares a world
  with beta's stealthed asset. A fixed roller makes beta fail. After alpha's turn, the
  asset's `Stealthy` is false. After beta's turn alone (Matrix owner not acting), it is
  unchanged.

##### Task 7 — `docs/architecture/engine/ability-catalog.md`

Mark the shipped rows and correct their hook labels:

- **reactive reveal-stealth** row: hook cell becomes "Cat 3 `MutationReactor` (on
  arrival); turn-start phase route (per turn)". Mark Panopticon Matrix `SWN-C8-001` ✅.
- **bonus-die** row: mark all three assets ✅.
- **reroll** row: mark Book of Secrets ✅.
- **negate-bonus** row: hook cell becomes "Cat 1 `RollOfferFilter`". Mark Blackmail ✅.

Integral Protocols' ✅ on the bonus-die row covers its die only. Its restriction row
stays unmarked until Effort 4.

##### Commit message

```
feat: clear_stealth on turn start

- turn-start route carries FactionState to its phase effects
- clear_stealth gains a turn_start phase route testing rivals on its world
- author Panopticon Matrix's stealth test
- mark the effort's shipped assets in the ability catalog
```

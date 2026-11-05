# Roll Card — Implementation Plan

## In Plain Words

- **What:** on every stat roll with someone to ask, the GM sees a card before the dice, with
  each extra die as a toggle. After the dice, a second card shows them and each reroll on
  offer, with its owner. With no reroll on offer, the dice go to the feed with no stop.
- **Why:** today's prompts do not say who rolls or for what, and the reroll prompt hides the
  dice. Psychic Academy, the last SWN tag with no data entry, can force a reroll on any
  rival's roll, so it needs a prompt that shows the roll.
- **How it fits:** this follows `R.028`, which put every hooked roll through one function,
  and it finishes the tags from `A.001.7`. `F.037` follows, to bring the card to the three
  contests that ask no one today. The same work fixes `B.016` and `B.024`, closes `F.032`,
  sweeps `R.015`, and builds the first adapter tests from `F.027`.

## Context / Goal

- Discovery doc: [F.039 roll card discovery](../discovery/F.039-roll-card-discovery.md)
- Related: `F.037` and `F.038` follow this work; `D.010` follows it for the wider page refresh.

The plan is one effort: three phases and nine commits on `feature/roll-card`.

## Decisions Ratified in Planning

Discovery's D-1 to D-12 stand. These are the plan's own, all Robert's calls in the session
unless marked as a plan-level choice.

1. **A wrong reply type on a card ask declines and reports** — resolves Open Question 1. The
   roll carries on with no die added and no reroll, and an `EvtError` shows in the status
   line and feed. Only a code bug can cause it, so it does not justify an error return
   through `Stat`, `Contest`, `Gate` and their call sites; reactors have no error path at
   all. Rejected: panic (leaves the terminal in a mess); an error return.
2. **The driver decides that the after card does not wait** — resolves Open Question 2. The
   engine calls `SelectReroll` when a roll settles, always. With no reroll on offer the
   adapter sends the dice to the feed and returns. A pass does the same, so each roll with a
   collector leaves one feed line. Rejected: the engine telling the feed, which threads an
   observer through every `dice.Engine` built with a collector.
3. **`B.024`'s shared use drops the trigger from `budgetKey`** — resolves Open Question 3.
   Book of Secrets is the only definition with two `once_per` entries of one effect, so that
   is the trigger's only job in the key. No data changes, so `A.001.5` need not wait. A
   mid-turn save gets one extra use, cleared at the owner's next bookkeeping (`B.012`'s
   pattern). Rejected: a new `once_per` value.
4. **The `-race` cycle test lives in the `execution` package** — resolves Open Question 4.
   That is where `snapshotOf` reads the live pointer that `B.038` suspects.
5. **A spent once-per offer shows on the before card as "used this turn"** — that card always
   waits anyway. The after card stays silent on spent rerolls, so it never waits with nothing
   to choose (D-5).
6. **Expand Influence's rival rolls show the expander's total** — this widens discovery
   finding 4, which passed a total only inside `Contest`. The card says "vs 9", not "to
   beat", because a rival's tie also earns the free attack.
7. **The cards are hand-drawn, with no huh form** — `↑/↓` move, `space` toggles, `enter`
   confirms. `B.029` cannot happen, a card with no offers needs no special case, and a `>`
   cursor gives `F.036`'s text cue on these cards.
8. **Every applied reroll leaves a note on the card and feed** — for example "Fanatical
   rerolled a 1". The full reroll history stays with `F.038`.
9. **Psychic Academy's entry also goes in the tracked sandbox copy** — the sandbox cannot
   test it otherwise. Precedent: `249db61`.
10. **`SelectModifiers` keeps its name; `ConfirmReroll` becomes `SelectReroll`** — plan-level
    choice. It now picks one of several, and nil means pass.
11. **The two asks rename in place** (`AskPreRollCard`, `AskPostRollCard`) — plan-level
    choice. `AskKind` is not reordered.
12. **The opponent's total travels as `Side.Total` to `RollContext.OpponentTotal`** —
    plan-level choice. `Contest` and Expand Influence both set it, and the card reads it from
    the context.
13. **A hook taken on a roll is not offered again on that roll** — plan-level choice. An
    elective reroll with no budget would otherwise return forever. The dispatcher stamps
    `RerollDirective.Slot` to tell the hooks apart.
14. **A nil registry asks no one** — plan-level choice. It narrows D-4: production always has
    a registry (`engine/core.go:52`), so every real roll with a collector still shows cards.
    It also spares the mock tests that pass a nil registry, so far fewer setups change than
    the session's first estimate of about 60.
15. **`mocks.NewRollingCollector` allows the two card calls any number of times** — plan-level
    choice. Tests that roll with a real registry swap one constructor.
16. **A card lists at most 8 rows at once and scrolls with the cursor** — plan-level choice.
    Three fact rows, a blank, 8 rows, a blank and the prompt fill the 14-row wizard band
    (`B.035`).
17. **The after card opens with the cursor on Keep** — plan-level choice. A stray `enter`
    cannot spend a once-per reroll.
18. **`B.016` restores every faction's whole `HookBudgets` on any `Run` error** — plan-level
    choice. Lobbyists' `permission_withdrawn` mark lives there too, so an aborted action
    clears that mark as well. That is right: the whole action never happened.

## Decision Record — Execution

1. **Phase 1, Commit 1: two more exact-context tests expect `OpponentTotal`.** Not in the
   plan. `dice/stat_test.go` (`TestContest_RollsTheAttackerFirstWithItsPhases`) and
   `effects/contest_test.go` (`TestEffectGate_RollPhasesAndSides`) compare whole contexts, so
   the defender's context now carries the attack's total. `expand_influence_test.go`'s
   context loop takes the pointer from the recorded context, then the new assertion checks
   the value.
2. **Phase 1, Commit 1: gap, the shim asks a declined reroll again.** Deferred on purpose.
   With two elective rerolls on offer, the GM can decline the first and take the second. The
   next after card still lists the first, so the shim asks about it again. The old prompts
   asked once per hook. It lives only between this commit and Commit 5, which deletes the
   shim.
3. **Phase 1, Commit 3: the task tables' line numbers moved.** `bonus_die_test.go:187` is now
   line 196 and `remove_tag_dice_test.go:143` is now line 156, because Commit 1 grew both
   files. The Find strings still matched. Checked Decision 3 against the data: in
   `rulebooks/swn` and `campaigns/sandbox/rulebook`, Book of Secrets is the only definition
   with two `once_per` entries of one effect.
4. **Phase 2, Commit 4: gofmt splits two observer one-liners.** Not in the plan. Task 3's
   `o.adapter.emit(` makes `OnFactionSkipped` and `OnStatRaiseSkipped` too long for gofmt to
   keep on one line, so each now takes three lines. No behaviour changes.
5. **Phase 2, Commit 5: roll cards sit flush left.** Not in the plan; the manual check found
   it. `wizardBand` places an overlay with `lipgloss.Center`, which centers each short line
   alone, so a card's rows did not line up. huh forms fill the band, so they never showed
   this. The band now uses `lipgloss.Left` for the two card asks only. The checkpoint's
   centered one-liner stays as it was.
6. **Phase 2, Commit 5: the feed wraps long lines.** Not in the plan; the manual check found
   it. The viewport cuts off a line at the column's width, so at 140 columns a feed line read
   "(Book of" and lost the rest of Decision 8's note. `streamContent` now wraps the feed to
   the stream's width. This covers every event, not only rolls; short lines look the same.
7. **Phase 2, Commit 5: the cards draw their own keys.** Not in the plan; the manual check
   found it. Nothing calls an overlay's `Help()`: the footer reads `execution.Model.Help`,
   which is always scroll and follow. huh forms draw their own hint line, but a hand-drawn
   card had none, so "space toggles" reached no screen. The before card with offers now says
   so in its prompt, and the after card gains a blank and a hint row. `preRollHelp` and
   `postRollHelp` stay, because the interface asks for them.
8. **Phase 2, Commit 5: gap, the footer is wrong under every overlay.** Deferred for Robert.
   While any overlay holds the keys, the footer still offers `tab next mode`, `↑/↓ scroll`
   and `G follow`, and `overlays.md` says the help keymaps carry the Esc contract, though
   nothing draws them. Passing the mounted overlay's `Help()` through would fix it, but it
   changes the footer of about 18 prompts, so it is not part of this commit.
9. **Phase 2, Commit 5: gap, wrapped rows can overrun the 14-row band.** Deferred. Decision
   16's sum counts one screen row per card row. A long detail line already wraps to two at
   140 columns, and offer rows can wrap too, so a card with about 7 or more rows can pass 14.
   This is reasoned from the sum, not seen: the fullest card in the check had 2 rows.
10. **Phase 2, Commit 5: Record entry 2's gap is closed.** This commit deletes the shim, so a
    declined reroll is no longer asked about again.
11. **Phase 2, Commit 5: the manual check's setup.** Task 6 did not say that the app needs
    `GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns`, or that tmux may be missing. tmux was
    not installed, so it went in through `apt-get`. The `combat` scenario has no bonus die,
    so one further run gave Veiled Hand the Machiavellian tag in `sandbox-claude`'s
    `state.toml` to see an offer row, and a reset removed it.
12. **Phase 3, Commit 6: scenario details the task left open.** The scenarios pin the turn
    order with `runOrderedCycle`, because `Turn.Start` rolls the order on the roller it took
    at construction, not `h.Engine.Rand`. Gamma's bookkeeping must come after the attack only
    in the Bookkeeping scenario. In "Holder's own roll", gamma stands on Tartarus, not on
    another world, because it makes the attack. The two tag files are identical after the
    edit, so the harness's load of `rulebooks/swn` also vouches for the sandbox copy.
13. **Phase 3, Commit 7: the sweep's check.** Read every overlay's answer against its
    method's `replyAs` type, and all match. Ran one attack and one Expand Influence, through
    a rival's roll and the free-attack prompt, in `sandbox-claude`: no error in the feed.
    That run confirmed bool, slice, pointer, string and interface answers at the table; the
    other overlays rest on the read and the table's types. The tests rebuild the world index
    in two fixtures, because `AvailableActions` and `SelectSeizeTarget` read it.
14. **Phase 3, Commit 8: `B.038` is real, and logged, not fixed.** Under `-race`, the test
    fails in 11 runs of 11 with one report: `snapshotOf` (`execution.go:214`, reading `Coin`) on
    the view's goroutine against `MutationEngine.Apply` (`mutation.go:68`, writing `Coin`) on
    the engine's. Other fields may race as well; each run reported only this pair. The plain
    gate passes. So `go test -race` over this package stays red until `B.038` is fixed. Robert
    ruled that the one note on the `B.038` row is in scope, as this commit's task says, and
    that the Out of Scope entry on tracker rows does not reach it.
15. **Phase 3, Commit 9: reversal, `overlays.md` no longer says a card fits the band.** The
    plan's "A card lists at most 8 rows, so it fits the band" is false. A detail row wrapped
    at 140 columns in Commit 7's run, and the after card adds a row per applied note. Robert
    chose this text over the plan's: "A card lists at most 8 rows at once and scrolls with
    the cursor. A wrapped row or an applied note can still push a full card past the band
    (`F.039` Decision Record 9)." Entry 9 stays open, and the sentence goes when it is fixed.
16. **Phase 3, Commit 9: three wording fixes, from checking the source.** The cards are not
    "plain data", because an offer holds an `Apply` function: they are "structs the collector
    only reads". Only the before card handles `space`. `AwaitCheckpoint` ignores its reply,
    so the `replyAs` claims say "that reads a reply". The plan body above matches.
17. **Phase 3, Commit 9: gap, `event-stream.md` still says views share no mutable state.** Its
    Key Decision "The adapter copies everything the views render" says the two goroutines
    never share mutable state, and entry 14 proves that false. This work did not make it
    false, so it is left for `B.038`'s fix or `D.010`. The same holds for `overlays.md`'s
    claim that the help keymap carries the Esc contract (entry 8).
18. **Phase 3, after Commit 9: one fix commit, for a flaky Commit 6 scenario.** Not in the
    plan. `TestPsychicAcademy_DeclinedKeepsTheBudget` failed about 1 run in 6. Every stub
    world shares one hex, so gamma's asset on Erebus was an eligible defender, and
    `eligible[0]` follows map order. The scenarios now pick beta's asset. Commit 9 landed with
    this failure showing: its command committed whatever the gate said, because the gate's
    output went through `grep`. Commits 6 to 8 passed by luck on this test.

---

## Shared Context

- **Gate for every commit:** `go build ./... && go vet ./... && go test ./internal/faction/...`.
- **Mock file:** `mockgen` is not installed, so Commit 1 edits `mock_collector.go` by hand, in
  mockgen's own shape and order. Regenerating gives the same file.
- **Running the app:** only with `--campaign sandbox-claude`, through tmux. `sandbox` is
  Robert's.
- **Test descriptions:** tasks give test stubs as code and new test cases in plain language.

## Out of Scope

- **`F.025`** — each faction's own driver; waits for `A.003`. The card asks stay `Faction: nil`.
- **`F.037`** — the three contests with no collector.
- **`F.038`** — the held result, the earlier matchups and the full reroll history.
- **`R.026`** — the rules in Go that D-1 keeps.
- **`B.038`'s fix** — Commit 8 logs what it finds and fixes nothing.
- **`D.010`** — Commit 9 changes only the parts of the pages this work makes false.
- **Tracker rows** — closing `F.032`, `B.016`, `B.024`, `R.015` and the others, and shrinking
  `F.027` (D-11), are for `/gestalt:initiative-close`.

---

## Work Breakdown

**Overview**

- [x] **Phase 1 — The engine asks with cards**
  - Commit 1 — `feat: roll cards on the hook collector` (11 tasks)
  - Commit 2 — `fix: restore hook budgets when an action aborts` (2 tasks)
  - Commit 3 — `fix: one shared Book of Secrets use per turn` (3 tasks)
- [x] **Phase 2 — The GM sees the cards**
  - Commit 4 — `feat: adapter builds roll card views` (6 tasks)
  - Commit 5 — `feat: roll card overlays` (6 tasks)
- [x] **Phase 3 — Psychic Academy and rigor**
  - Commit 6 — `feat: Psychic Academy's reroll entry` (4 tasks)
  - Commit 7 — `refactor: check every collector reply type` (4 tasks)
  - Commit 8 — `test: full cycle through the real channels under race` (1 task)
  - Commit 9 — `docs: roll card architecture pages` (4 tasks)

### Phase 1 — The engine asks with cards

**In plain words:** the roll code hands each driver a full picture of the roll in place of a
bare list, and it runs forced rerolls first, then a loop of elective ones. The GM sees nothing
new yet, because a small shim drives the old prompts from the new cards.

#### Commit 1 — `feat: roll cards on the hook collector`

##### Task 1 — `internal/faction/engine/hooks/types.go`

**Find:**
```go
// The side resisting detection never sets it (Effort 3 Decision 5).
type RollContext struct {
```
**Replace with:**
```go
// The side resisting detection never sets it (Effort 3 Decision 5). OpponentTotal is
// the other side's total when it has already rolled, and nil when it has not.
type RollContext struct {
```

**Find:**
```go
	Action            string
	Detection         bool
}
```
**Replace with:**
```go
	Action            string
	Detection         bool
	OpponentTotal     *int
}
```

**Find:**
```go
// and spends BudgetKey on that faction's HookBudgets.
type RerollDirective struct {
	Source    string
	Indices   []int  // indices into RollResult.Dice to reroll
	BudgetKey string // namespaced; empty = unlimited
	Elective  bool   // if true, collector must confirm before applying
	FactionID string
}
```
**Replace with:**
```go
// and spends BudgetKey on that faction's HookBudgets. Slot names the hook that made
// the directive; the dispatcher sets it, and does not offer a taken hook again on
// the same roll.
type RerollDirective struct {
	Source    string
	Indices   []int  // indices into RollResult.Dice to reroll
	BudgetKey string // namespaced; empty = unlimited
	Elective  bool   // if true, the collector chooses whether it applies
	FactionID string
	Slot      int
}

// StruckOffer is a Cat 1 offer that did not survive. StruckBy names the filter that
// struck it; an empty StruckBy means its budget is spent this turn.
type StruckOffer struct {
	Offer    ModifierOffer
	StruckBy string
}

// AppliedReroll is a reroll that has already changed this roll. Was holds the dice
// it replaced.
type AppliedReroll struct {
	Source    string
	FactionID string
	Was       []int
}

// PreRollCard is what the collector sees before the dice: the roll, the offers it
// may take and the offers it may not.
type PreRollCard struct {
	Context RollContext
	Base    domain.DiceRoll
	Offers  []ModifierOffer
	Struck  []StruckOffer
}

// PostRollCard is what the collector sees once forced rerolls have applied and after
// each elective reroll taken. Rerolls may be empty; Result.Dice is live, so read it
// before returning.
type PostRollCard struct {
	Context RollContext
	Result  RollResult
	Rerolls []RerollDirective
	Applied []AppliedReroll
}
```

##### Task 2 — `internal/faction/engine/hooks/collector.go`

**Find:**
```go
// Collector abstracts the two input methods that hook dispatchers require from
// the GM. SelectModifiers is called with the full Cat 1 offer list; the GM
// returns the subset to apply. ConfirmReroll is called for each elective Cat 2
// directive; returning false skips that reroll.
type Collector interface {
	SelectModifiers(offers []ModifierOffer) []ModifierOffer
	ConfirmReroll(directive RerollDirective) bool
}
```
**Replace with:**
```go
// Collector abstracts the two asks of a hooked roll. SelectModifiers is called before
// the dice on every roll, offers or not, and returns the offers to apply. SelectReroll
// is called once the roll settles, rerolls on offer or not, and again after each
// reroll it takes; it returns the reroll to take, or nil to keep the dice.
type Collector interface {
	SelectModifiers(card PreRollCard) []ModifierOffer
	SelectReroll(card PostRollCard) *RerollDirective
}
```

##### Task 3 — `internal/faction/engine/hooks/interfaces.go`

**Find:**
```go
// including one another faction's hook made.
type RollOfferFilter interface {
	FilterOffers(ctx RollContext, offers []ModifierOffer, factionState *state.FactionState) []ModifierOffer
}
```
**Replace with:**
```go
// including one another faction's hook made. It returns what it keeps and what it
// strikes, so the card can show why an offer is gone.
type RollOfferFilter interface {
	FilterOffers(ctx RollContext, offers []ModifierOffer, factionState *state.FactionState) ([]ModifierOffer, []StruckOffer)
}
```

**Find:**
```go
// RollResultHook fires after dice are rolled and may instruct the dispatcher
// to reroll specific dice (Cat 2). Elective directives are confirmed by the
// collector; non-elective directives apply automatically.
```
**Replace with:**
```go
// RollResultHook fires after dice are rolled and may instruct the dispatcher
// to reroll specific dice (Cat 2). Elective directives are chosen by the
// collector; non-elective directives apply automatically.
```

##### Task 4 — `internal/faction/engine/dice/stat.go`

**Find:**
```go
// test, a Base defending). Detects marks a side testing to detect stealth.
type Side struct {
	Faction *domain.Faction
	Asset   *domain.Asset
	Stat    domain.FactionStat
	Detects bool
}
```
**Replace with:**
```go
// test, a Base defending). Detects marks a side testing to detect stealth. Total is
// this side's total when it has already rolled, so the other side's card can show it.
type Side struct {
	Faction *domain.Faction
	Asset   *domain.Asset
	Stat    domain.FactionStat
	Detects bool
	Total   *int
}
```

**Find:**
```go
		Detection:         side.Detects,
	}
```
**Replace with:**
```go
		Detection:         side.Detects,
		OpponentTotal:     opponent.Total,
	}
```

**Find:**
```go
	outcome := Outcome{
		Attack:  engine.Stat(attackPhase, attacker, attackerOpponent, world, action),
		Defense: engine.Stat(defensePhase, defender, defenderOpponent, world, action),
	}
```
**Replace with:**
```go
	attack := engine.Stat(attackPhase, attacker, attackerOpponent, world, action)
	defenderOpponent.Total = &attack.Sum
	outcome := Outcome{
		Attack:  attack,
		Defense: engine.Stat(defensePhase, defender, defenderOpponent, world, action),
	}
```

##### Task 5 — `internal/faction/engine/dice/roll.go`

Replace the doc comment and body of `rollWithHooks` (lines 14 to 119 today) with:

```go
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
		choice := collector.SelectReroll(hooks.PostRollCard{Context: ctx, Result: step.result, Rerolls: step.electiveOffers(), Applied: step.applied})
		if choice == nil {
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
```

In `offersBudgetFilter`, report what it removes.

**Find:**
```go
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
```
**Replace with:**
```go
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
```

##### Task 6 — `internal/faction/engine/effect/effects/remove_tag_dice.go`

**Find:**
```go
		&tagDiceFilter{
			holder:  refOf(ctx.Holder),
			trigger: ctx.Entry.Trigger,
```
**Replace with:**
```go
		&tagDiceFilter{
			holder:  refOf(ctx.Holder),
			source:  ctx.Holder.Name(),
			trigger: ctx.Entry.Trigger,
```

**Find:**
```go
type tagDiceFilter struct {
	holder  holderRef
	trigger domain.EffectTrigger
```
**Replace with:**
```go
type tagDiceFilter struct {
	holder  holderRef
	source  string
	trigger domain.EffectTrigger
```

**Find:**
```go
func (filter *tagDiceFilter) FilterOffers(ctx hooks.RollContext, offers []hooks.ModifierOffer, factionState *state.FactionState) []hooks.ModifierOffer {
```
**Replace with:**
```go
func (filter *tagDiceFilter) FilterOffers(ctx hooks.RollContext, offers []hooks.ModifierOffer, factionState *state.FactionState) ([]hooks.ModifierOffer, []hooks.StruckOffer) {
```

**Find:**
```go
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
**Replace with:**
```go
		return offers, nil
	}
	var kept []hooks.ModifierOffer
	var struck []hooks.StruckOffer
	for _, offer := range offers {
		if offer.FromTag {
			struck = append(struck, hooks.StruckOffer{Offer: offer, StruckBy: filter.source})
			continue
		}
		kept = append(kept, offer)
	}
	return kept, struck
}
```

##### Task 7 — `internal/faction/engine/action/actions/expand_influence.go`

In `resolveNewBase`'s rival loop, the rival's card shows the expander's total.

**Find:**
```go
			dice.Side{Faction: rival, Stat: domain.StatCunning},
			dice.Side{Faction: faction, Stat: domain.StatCunning},
			ei.order.World, ei.Name(),
```
**Replace with:**
```go
			dice.Side{Faction: rival, Stat: domain.StatCunning},
			dice.Side{Faction: faction, Stat: domain.StatCunning, Total: &factionRoll},
			ei.order.World, ei.Name(),
```

##### Task 8 — `internal/faction/tui/adapter/action_collector.go` and `smoke.go`

The shim keeps today's two prompts until Commit 5 deletes it.

In `action_collector.go`, **find:**
```go
func (a *actionCollector) SelectModifiers(offers []hooks.ModifierOffer) []hooks.ModifierOffer {
	raw, err := a.adapter.ask(AskSelectModifiers, nil, SelectModifiersPayload{Offers: offers})
	if err != nil {
```
**Replace with:**
```go
func (a *actionCollector) SelectModifiers(card hooks.PreRollCard) []hooks.ModifierOffer {
	if len(card.Offers) == 0 {
		return nil
	}
	raw, err := a.adapter.ask(AskSelectModifiers, nil, SelectModifiersPayload{Offers: card.Offers})
	if err != nil {
```

**Find:**
```go
func (a *actionCollector) ConfirmReroll(directive hooks.RerollDirective) bool {
	raw, err := a.adapter.ask(AskConfirmReroll, nil, ConfirmRerollPayload{Directive: directive})
	if err != nil {
		return false // Esc/cancel -> skip the reroll
	}
	confirmed, _ := raw.(bool)
	return confirmed
}
```
**Replace with:**
```go
func (a *actionCollector) SelectReroll(card hooks.PostRollCard) *hooks.RerollDirective {
	for index := range card.Rerolls {
		raw, err := a.adapter.ask(AskConfirmReroll, nil, ConfirmRerollPayload{Directive: card.Rerolls[index]})
		if err != nil {
			continue // Esc/cancel -> skip the reroll
		}
		if confirmed, _ := raw.(bool); confirmed {
			return &card.Rerolls[index]
		}
	}
	return nil
}
```

In `smoke.go`, **find:**
```go
func (SmokeActionCollector) SelectModifiers(_ []hooks.ModifierOffer) []hooks.ModifierOffer {
	return nil
}
func (SmokeActionCollector) ConfirmReroll(_ hooks.RerollDirective) bool { return false }
```
**Replace with:**
```go
func (SmokeActionCollector) SelectModifiers(_ hooks.PreRollCard) []hooks.ModifierOffer {
	return nil
}
func (SmokeActionCollector) SelectReroll(_ hooks.PostRollCard) *hooks.RerollDirective { return nil }
```

##### Task 9 — `internal/faction/engine/testharness/collector.go`

**Find:**
```go
	SelectModifiersFn             func([]hooks.ModifierOffer) []hooks.ModifierOffer
	ConfirmRerollFn               func(hooks.RerollDirective) bool
```
**Replace with:**
```go
	SelectModifiersFn             func(hooks.PreRollCard) []hooks.ModifierOffer
	SelectRerollFn                func(hooks.PostRollCard) *hooks.RerollDirective
```

**Find:**
```go
// SelectModifiers takes all offered modifiers by default; override with SelectModifiersFn.
func (c *ScriptedCollector) SelectModifiers(offers []hooks.ModifierOffer) []hooks.ModifierOffer {
	if c.SelectModifiersFn != nil {
		return c.SelectModifiersFn(offers)
	}
	return offers
}

// ConfirmReroll confirms all rerolls by default; override with ConfirmRerollFn.
func (c *ScriptedCollector) ConfirmReroll(directive hooks.RerollDirective) bool {
	if c.ConfirmRerollFn != nil {
		return c.ConfirmRerollFn(directive)
	}
	return true
}
```
**Replace with:**
```go
// SelectModifiers takes all offered modifiers by default; override with SelectModifiersFn.
func (c *ScriptedCollector) SelectModifiers(card hooks.PreRollCard) []hooks.ModifierOffer {
	if c.SelectModifiersFn != nil {
		return c.SelectModifiersFn(card)
	}
	return card.Offers
}

// SelectReroll takes the first reroll on offer by default; override with SelectRerollFn.
func (c *ScriptedCollector) SelectReroll(card hooks.PostRollCard) *hooks.RerollDirective {
	if c.SelectRerollFn != nil {
		return c.SelectRerollFn(card)
	}
	if len(card.Rerolls) == 0 {
		return nil
	}
	return &card.Rerolls[0]
}
```

##### Task 10 — `internal/faction/engine/action/actions/mocks/`

In `mock_collector.go`, delete the `ConfirmReroll` mock and recorder methods (lines 122 to 134
today).

**Find:**
```go
// SelectModifiers mocks base method.
func (m *MockCollector) SelectModifiers(offers []hooks.ModifierOffer) []hooks.ModifierOffer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectModifiers", offers)
```
**Replace with:**
```go
// SelectModifiers mocks base method.
func (m *MockCollector) SelectModifiers(card hooks.PreRollCard) []hooks.ModifierOffer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectModifiers", card)
```

**Find:**
```go
func (mr *MockCollectorMockRecorder) SelectModifiers(offers any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectModifiers", reflect.TypeOf((*MockCollector)(nil).SelectModifiers), offers)
```
**Replace with:**
```go
func (mr *MockCollectorMockRecorder) SelectModifiers(card any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectModifiers", reflect.TypeOf((*MockCollector)(nil).SelectModifiers), card)
```

Directly before the `// SelectSeizeTarget mocks base method.` comment, insert:
```go
// SelectReroll mocks base method.
func (m *MockCollector) SelectReroll(card hooks.PostRollCard) *hooks.RerollDirective {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectReroll", card)
	ret0, _ := ret[0].(*hooks.RerollDirective)
	return ret0
}

// SelectReroll indicates an expected call of SelectReroll.
func (mr *MockCollectorMockRecorder) SelectReroll(card any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectReroll", reflect.TypeOf((*MockCollector)(nil).SelectReroll), card)
}

```

`mocks/rolling.go` (new) — full contents:
```go
package mocks

import "go.uber.org/mock/gomock"

// NewRollingCollector is a MockCollector for a test that rolls with a real registry
// and does not care about the roll cards: each card asks and takes nothing.
func NewRollingCollector(ctrl *gomock.Controller) *MockCollector {
	collector := NewMockCollector(ctrl)
	collector.EXPECT().SelectModifiers(gomock.Any()).AnyTimes()
	collector.EXPECT().SelectReroll(gomock.Any()).AnyTimes()
	return collector
}
```

##### Task 11 — Tests

**`internal/faction/engine/action/actions/attack_test.go` and `expand_influence_test.go`:**
replace every `mocks.NewMockCollector(ctrl)` with `mocks.NewRollingCollector(ctrl)`. The
`AnyTimes` calls do no harm where a test does not roll.

**`internal/faction/engine/action/actions/ability/dispatch_test.go`:** at the two sites that
take the offer (`TestDispatch_ContestRollsThroughHooks` and the defense-die test), **find:**
```go
	collector.EXPECT().SelectModifiers(gomock.Any()).
		DoAndReturn(func(offers []hooks.ModifierOffer) []hooks.ModifierOffer { return offers }).
		Times(1)
```
**Replace with:**
```go
	collector.EXPECT().SelectModifiers(gomock.Any()).
		DoAndReturn(func(card hooks.PreRollCard) []hooks.ModifierOffer { return card.Offers }).
		AnyTimes()
	collector.EXPECT().SelectReroll(gomock.Any()).AnyTimes()
```
In the test that registers `"test:recorder"` (line 495 today), **find:**
```go
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectAbilityTarget(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.AbilityTarget{Kind: domain.TargetFaction, Faction: targetFaction}, nil)

	recorder := &contextLog{}
```
**Replace with:**
```go
	collector := mocks.NewRollingCollector(ctrl)
	collector.EXPECT().SelectAbilityTarget(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.AbilityTarget{Kind: domain.TargetFaction, Faction: targetFaction}, nil)

	recorder := &contextLog{}
```
If any other mock test fails with an unexpected `SelectModifiers` or `SelectReroll` call, swap
its constructor the same way.

**`internal/faction/engine/effect/effects/bonus_die_test.go`:** the counters keep meaning "asked
with something to choose", so the existing assertions hold. **Find:**
```go
func (collector *rollCollector) SelectModifiers(offers []hooks.ModifierOffer) []hooks.ModifierOffer {
	collector.selections++
	if collector.accept {
		return offers
	}
	return nil
}

func (collector *rollCollector) ConfirmReroll(hooks.RerollDirective) bool {
	collector.rerolls++
	return collector.accept
}
```
**Replace with:**
```go
func (collector *rollCollector) SelectModifiers(card hooks.PreRollCard) []hooks.ModifierOffer {
	if len(card.Offers) == 0 {
		return nil
	}
	collector.selections++
	if collector.accept {
		return card.Offers
	}
	return nil
}

func (collector *rollCollector) SelectReroll(card hooks.PostRollCard) *hooks.RerollDirective {
	if len(card.Rerolls) == 0 {
		return nil
	}
	collector.rerolls++
	if collector.accept {
		return &card.Rerolls[0]
	}
	return nil
}
```

**`internal/faction/engine/effect/effects/remove_tag_dice_test.go`:** the two call sites become
`kept, _ := filter.FilterOffers(...)`. New case: in the rival-attacking case, the struck list
holds the tag offer with `StruckBy` "Blackmail".

**`internal/faction/engine/hooks/registry_test.go`:** **find:**
```go
func (stub *stubRollOfferFilter) FilterOffers(_ hooks.RollContext, offers []hooks.ModifierOffer, _ *state.FactionState) []hooks.ModifierOffer {
	return offers
}
```
**Replace with:**
```go
func (stub *stubRollOfferFilter) FilterOffers(_ hooks.RollContext, offers []hooks.ModifierOffer, _ *state.FactionState) ([]hooks.ModifierOffer, []hooks.StruckOffer) {
	return offers, nil
}
```

**`internal/faction/engine/dice/roll_test.go`:** replace the stubs.

**Find:**
```go
func (acceptAllCollector) SelectModifiers(offers []hooks.ModifierOffer) []hooks.ModifierOffer {
	return offers
}

func (acceptAllCollector) ConfirmReroll(_ hooks.RerollDirective) bool { return true }
```
**Replace with:**
```go
func (acceptAllCollector) SelectModifiers(card hooks.PreRollCard) []hooks.ModifierOffer {
	return card.Offers
}

func (acceptAllCollector) SelectReroll(card hooks.PostRollCard) *hooks.RerollDirective {
	if len(card.Rerolls) == 0 {
		return nil
	}
	return &card.Rerolls[0]
}
```

**Find:**
```go
func (rejectAllCollector) SelectModifiers(_ []hooks.ModifierOffer) []hooks.ModifierOffer {
	return nil
}

func (rejectAllCollector) ConfirmReroll(_ hooks.RerollDirective) bool { return false }
```
**Replace with:**
```go
func (rejectAllCollector) SelectModifiers(_ hooks.PreRollCard) []hooks.ModifierOffer {
	return nil
}

func (rejectAllCollector) SelectReroll(_ hooks.PostRollCard) *hooks.RerollDirective { return nil }
```

**Find:**
```go
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
```
**Replace with:**
```go
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
```

Replace every remaining `failOnSelectCollector{t}` with `failOnOffersCollector{t}`.

**Find:**
```go
func (stripAllFilter) FilterOffers(_ hooks.RollContext, _ []hooks.ModifierOffer, _ *state.FactionState) []hooks.ModifierOffer {
	return nil
}
```
**Replace with:**
```go
func (stripAllFilter) FilterOffers(_ hooks.RollContext, offers []hooks.ModifierOffer, _ *state.FactionState) ([]hooks.ModifierOffer, []hooks.StruckOffer) {
	struck := make([]hooks.StruckOffer, 0, len(offers))
	for _, offer := range offers {
		struck = append(struck, hooks.StruckOffer{Offer: offer, StruckBy: "strip"})
	}
	return nil, struck
}
```

Rename `TestRollWithHooks_NoOffersSkipsCollector` to `TestRollWithHooks_NoOffersStillAsks` and
replace its body so it asserts one before card and one after card on a roll with an empty
registry. New tests, all with `cardCollector`:

- **Nil registry** — with a nil registry, the collector sees no card and the dice still roll.
- **Struck offers** — a filter's strike and a spent budget both reach `card.Struck`, with
  `StruckBy` set for the filter and empty for the budget, and neither is in `card.Offers`.
- **Forced first** — a forced hook registered after an elective one still applies before the
  first after card, and `card.Applied` holds its `Source` and `Was`.
- **Forced once (D-2)** — roll 1, forced reroll gives 5, elective take gives 1: the die stays
  1 and the roller is called three times (`sidesRecorder`).
- **The loop (D-6)** — two elective hooks: the first after card holds two rerolls, taking one
  brings a card with one, and a pass ends the roll.
- **Taken once** — an elective hook with no budget, always taken, is rolled once, and the
  second card has no rerolls.
- **Nil collector** — a forced reroll applies and no card is built.

**`internal/faction/engine/dice/stat_test.go`:** new tests with `recordedEngine`:

- **Attack total** — in a contest, the attacker's context has a nil `OpponentTotal`, and the
  defender's holds the attack's `Sum`.
- **Same faction** — a self-contest passes the total too, with `Opponent` still nil.

**`internal/faction/engine/action/actions/expand_influence_test.go`:** new assertion in the test
that registers `"test:recorder"`: the rival's recorded context has `OpponentTotal` equal to the
expander's total, and the expander's own context has none.

##### Commit message

```
feat: roll cards on the hook collector

- hooks.Collector takes a PreRollCard and a PostRollCard; ConfirmReroll becomes SelectReroll
- every hooked roll with a collector asks before and after, choice or not (D-4)
- forced rerolls run once, first; elective rerolls loop until the collector passes (D-2, D-6)
- offer filters and the budget filter report what they strike (D-10)
- Contest and Expand Influence pass the first total to the other side's card
- applied rerolls are noted on the card
- the adapter shims the new cards onto the old prompts until the overlays land
```

#### Commit 2 — `fix: restore hook budgets when an action aborts`

##### Task 1 — `internal/faction/engine/action/action.go`

Add `"maps"` to the imports.

**Find:**
```go
// Run calls Inputs, Resolve, and Output in order on the selected action.
func (ae *ActionEngine) Run(a Action, faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook, log *slog.Logger) ([]domain.Mutation, error) {
	log.Info("running action", "action", a.Name())
	if err := a.Inputs(faction, factionState, rulebook); err != nil {
		return nil, err
	}
	if err := a.Resolve(faction, factionState, rulebook); err != nil {
		return nil, err
	}
	mutations, err := a.Output()
	if err != nil {
		return nil, err
	}
	log.Info("action resolved", "mutations", len(mutations))
	return mutations, nil
}
```
**Replace with:**
```go
// Run calls Inputs, Resolve, and Output in order on the selected action. A failed
// action drops its mutations, so the budgets it spent on live state are restored.
func (ae *ActionEngine) Run(a Action, faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook, log *slog.Logger) ([]domain.Mutation, error) {
	log.Info("running action", "action", a.Name())
	budgets := snapshotBudgets(factionState)
	mutations, err := runSteps(a, faction, factionState, rulebook)
	if err != nil {
		restoreBudgets(factionState, budgets)
		return nil, err
	}
	log.Info("action resolved", "mutations", len(mutations))
	return mutations, nil
}

func runSteps(a Action, faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) ([]domain.Mutation, error) {
	if err := a.Inputs(faction, factionState, rulebook); err != nil {
		return nil, err
	}
	if err := a.Resolve(faction, factionState, rulebook); err != nil {
		return nil, err
	}
	return a.Output()
}

func snapshotBudgets(factionState *state.FactionState) map[string]map[string]int {
	budgets := make(map[string]map[string]int, len(factionState.Factions))
	for factionID, faction := range factionState.Factions {
		budgets[factionID] = maps.Clone(faction.HookBudgets)
	}
	return budgets
}

func restoreBudgets(factionState *state.FactionState, budgets map[string]map[string]int) {
	for factionID, faction := range factionState.Factions {
		faction.HookBudgets = budgets[factionID]
	}
}
```

##### Task 2 — `internal/faction/engine/action/action_test.go` (new)

Plain-language table, with a stub action whose `Resolve` spends a budget on the acting faction
and on a rival, then returns its scripted error:

- **Recoverable error** — `ErrTurnCanceled`: both factions' budgets return to what they were,
  a nil map staying nil.
- **Fatal error** — a plain error: the same.
- **Failure in `Output`** — the same.
- **Success** — the spends stay.

##### Commit message

```
fix: restore hook budgets when an action aborts

- ActionEngine.Run snapshots every faction's HookBudgets and restores them on any error
- a prompt canceled late in an action no longer costs a once-per use (B.016)
```

#### Commit 3 — `fix: one shared Book of Secrets use per turn`

##### Task 1 — `internal/faction/engine/effect/effects/budget.go`

**Find:**
```go
// budgetKey names the once_per budget an entry spends. The effect name, trigger
// and definition ID keep two effects, two entries of one definition, or two asset
// kinds from sharing a budget.
func budgetKey(oncePer domain.OncePerKey, effectName domain.SpecialEffectType, trigger domain.EffectTrigger, defID, assetID, targetFactionID string) string {
	base := "effect:" + string(effectName) + ":" + string(trigger) + ":" + defID
```
**Replace with:**
```go
// budgetKey names the once_per budget an entry spends. The effect name and
// definition ID keep two effects or two asset kinds from sharing a budget.
// deliberate: no trigger in the key, because two entries of one effect on one definition share its use (Book of Secrets, B.024).
func budgetKey(oncePer domain.OncePerKey, effectName domain.SpecialEffectType, defID, assetID, targetFactionID string) string {
	base := "effect:" + string(effectName) + ":" + defID
```

##### Task 2 — The five call sites

| File | Find | Replace with |
|---|---|---|
| `bonus_die.go:78` | `domain.SpecialBonusDie, modifier.trigger, modifier.holder.sourceID` | `domain.SpecialBonusDie, modifier.holder.sourceID` |
| `reroll.go:97` | `domain.SpecialReroll, hook.trigger, hook.holder.sourceID` | `domain.SpecialReroll, hook.holder.sourceID` |
| `coin_steal.go:72` | `domain.SpecialCoinSteal, entry.Trigger, def.ID` | `domain.SpecialCoinSteal, def.ID` |
| `restore_asset.go:55` | `cause, domain.TriggerLethalDamage, interceptor.holder.sourceID` | `cause, interceptor.holder.sourceID` |
| `ignore_coin_loss.go:63` | `domain.SpecialIgnoreCoinLoss, reactor.trigger, reactor.holder.sourceID` | `domain.SpecialIgnoreCoinLoss, reactor.holder.sourceID` |

In `ignore_coin_loss.go`, delete the `trigger` field and its setter (lines 31 and 39 today); the
key was its only reader.

In `hooks/types.go`, **find** `"effect:bonus_die:own_roll:SWN-T-020"` and **replace with**
`"effect:bonus_die:SWN-T-020"`.

##### Task 3 — Tests

Key strings:

| File | Find | Replace with |
|---|---|---|
| `ignore_coin_loss_test.go:173`, `:176` | `"effect:ignore_coin_loss:coin_loss:"` | `"effect:ignore_coin_loss:"` |
| `coin_steal_test.go:192` | `"effect:coin_steal:own_attack_hit:SWN-F5-001:beta"` | `"effect:coin_steal:SWN-F5-001:beta"` |
| `bonus_die_test.go:187` | `"effect:bonus_die:own_roll:SWN-T-020"` | `"effect:bonus_die:SWN-T-020"` |
| `restore_asset_test.go:66` | `"effect:restore_asset:lethal_damage:SWN-W4-003:mc"` | `"effect:restore_asset:SWN-W4-003:mc"` |
| `remove_tag_dice_test.go:143` | `"effect:bonus_die:own_roll:SWN-T-020"` | `"effect:bonus_die:SWN-T-020"` |
| `reroll_test.go:272` | `"effect:reroll:rival_roll:SWN-T-014"` | `"effect:reroll:SWN-T-014"` |
| `scenarios/tags_test.go:144` | `"effect:bonus_die:own_roll:SWN-T-020"` | `"effect:bonus_die:SWN-T-020"` |
| `scenarios/effects_test.go:82` | `"effect:coin_steal:own_attack_hit:"` | `"effect:coin_steal:"` |

**`internal/faction/engine/effect/effects/budget_test.go`:** replace the first test's table and
the calls.

**Find:**
```go
		{domain.OncePerTargetFaction, "effect:coin_steal:own_attack_hit:DEF:beta"},
		{domain.OncePerOwnerFaction, "effect:coin_steal:own_attack_hit:DEF"},
		{domain.OncePerAsset, "effect:coin_steal:own_attack_hit:DEF:asset-1"},
	}
	for _, tc := range cases {
		if got := budgetKey(tc.oncePer, domain.SpecialCoinSteal, domain.TriggerOwnAttackHit, "DEF", "asset-1", "beta"); got != tc.want {
```
**Replace with:**
```go
		{domain.OncePerTargetFaction, "effect:coin_steal:DEF:beta"},
		{domain.OncePerOwnerFaction, "effect:coin_steal:DEF"},
		{domain.OncePerAsset, "effect:coin_steal:DEF:asset-1"},
	}
	for _, tc := range cases {
		if got := budgetKey(tc.oncePer, domain.SpecialCoinSteal, "DEF", "asset-1", "beta"); got != tc.want {
```

**Find:**
```go
	first := budgetKey(domain.OncePerOwnerFaction, domain.SpecialCoinSteal, domain.TriggerOwnAttackHit, "DEF", "asset-1", "beta")
	second := budgetKey(domain.OncePerOwnerFaction, domain.SpecialCoinSteal, domain.TriggerOwnAttackHit, "DEF", "asset-2", "gamma")
```
**Replace with:**
```go
	first := budgetKey(domain.OncePerOwnerFaction, domain.SpecialCoinSteal, "DEF", "asset-1", "beta")
	second := budgetKey(domain.OncePerOwnerFaction, domain.SpecialCoinSteal, "DEF", "asset-2", "gamma")
```

Replace `TestBudgetKey_TriggersDoNotShare` with:
```go
func TestBudgetKey_EntriesOfOneDefinitionShare(t *testing.T) {
	own := budgetKey(domain.OncePerOwnerFaction, domain.SpecialReroll, "DEF", "asset-1", "")
	forced := budgetKey(domain.OncePerOwnerFaction, domain.SpecialReroll, "DEF", "asset-1", "beta")

	if own != forced {
		t.Errorf("own key %q, forced key %q, want one budget", own, forced)
	}
}
```

**`internal/faction/engine/effect/effects/reroll_test.go`:**

**Find:**
```go
const (
	ownRerollKey    = "effect:reroll:own_roll:SWN-C7-002"
	forcedRerollKey = "effect:reroll:enemy_roll_at_me:SWN-C7-002"
)
```
**Replace with:**
```go
const bookRerollKey = "effect:reroll:SWN-C7-002"
```
Replace each `ownRerollKey` and `forcedRerollKey` with `bookRerollKey`.

Replace `TestReroll_TwoBudgetsAreSeparate` with:
```go
func TestReroll_BookOfSecretsSharesOneUse(t *testing.T) {
	factionState := bookState(activeAsset("bos", "SWN-C7-002"))
	registry := hooks.NewRegistry()
	registerBook(factionState, registry, ownReroll, forcedReroll)
	collector := &rollCollector{accept: true}

	rollThroughDispatcher(factionState, registry, collector, rollOn(factionState, "beta", "alpha", "Krylos"), 9, 3)
	result := rollThroughDispatcher(factionState, registry, collector, rollOn(factionState, "alpha", "beta", "Krylos"), 2)

	if result.Sum != 2 || collector.rerolls != 1 {
		t.Errorf("own roll Sum = %d after %d reroll asks; want 2 after 1: the forced reroll spent the one use", result.Sum, collector.rerolls)
	}
	if spent := factionState.Factions["alpha"].HookBudgets[bookRerollKey]; spent != 1 {
		t.Errorf("budget spent %d times, want 1", spent)
	}
}
```

##### Commit message

```
fix: one shared Book of Secrets use per turn

- budgetKey drops the trigger, so two entries of one effect on one definition share a use (B.024)
- Book of Secrets is the only definition with two such entries; no data changes
- a mid-turn save with old keys gets one extra use, cleared at the owner's next bookkeeping
```

### Phase 2 — The GM sees the cards

**In plain words:** the adapter turns each card into plain words, and two new prompts draw them
in the wizard band. When this phase lands, the GM sees the dice on every hooked roll, and each
roll leaves a line in the feed.

#### Commit 4 — `feat: adapter builds roll card views`

##### Task 1 — `internal/faction/tui/adapter/channels.go`

Append the feed event at the end of the enum.

**Find:**
```go
	EvtIndexSkipped
	EvtCycleStarted
)
```
**Replace with:**
```go
	EvtIndexSkipped
	EvtCycleStarted
	EvtRollSettled
)
```

**Find:**
```go
// Per-kind payload types for ObserverEventMsg.
```
**Replace with:**
```go
// KeepRoll is the post-roll card's answer for keeping the dice as they stand.
const KeepRoll = -1

// PreRollCardPayload is the before card in display words, with no IDs or pointers
// (D-12). Answer: []int, the indices of the Offers taken.
type PreRollCardPayload struct {
	Headline string
	Detail   string
	Dice     string
	Offers   []string
	Struck   []string
}

// PostRollCardPayload is the after card in display words. Answer: int, an index into
// Rerolls, or KeepRoll.
type PostRollCardPayload struct {
	Headline string
	Detail   string
	Dice     string
	Applied  []string
	Rerolls  []string
}

// Per-kind payload types for ObserverEventMsg.

// RollSettledPayload is a settled roll's feed line, in the card's wording.
type RollSettledPayload struct{ Line string }
```

##### Task 2 — `internal/faction/tui/adapter/roll_card.go` (new)

Full contents:
```go
package adapter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

var rollPhaseWords = map[domain.RollPhase]string{
	domain.RollPhaseAttack:      "to attack",
	domain.RollPhaseDefense:     "to defend",
	domain.RollPhaseFactionTest: "to test",
	domain.RollPhaseContested:   "to resist",
}

// preRollPayload words the before card. Safe on the engine goroutine: it is parked
// on this ask, and the payload holds only strings.
func (a *Adapter) preRollPayload(card hooks.PreRollCard) PreRollCardPayload {
	offers := make([]string, 0, len(card.Offers))
	for _, offer := range card.Offers {
		offers = append(offers, fmt.Sprintf("%s — %s, %s", offer.Description, offer.Source, a.factionName(offer.FactionID)))
	}
	struck := make([]string, 0, len(card.Struck))
	for _, row := range card.Struck {
		struck = append(struck, struckLine(row))
	}
	return PreRollCardPayload{
		Headline: rollHeadline(card.Context),
		Detail:   a.rollDetail(card.Context),
		Dice:     fmt.Sprintf("%dd%d %+d%s", card.Base.NumDice, card.Base.Sides, card.Base.Modifier, versus(card.Context)),
		Offers:   offers,
		Struck:   struck,
	}
}

// postRollPayload words the after card.
func (a *Adapter) postRollPayload(card hooks.PostRollCard) PostRollCardPayload {
	rerolls := make([]string, 0, len(card.Rerolls))
	for _, directive := range card.Rerolls {
		rerolls = append(rerolls, fmt.Sprintf("Reroll the %s — %s, %s", diceAt(card.Result.Dice, directive.Indices), directive.Source, a.factionName(directive.FactionID)))
	}
	return PostRollCardPayload{
		Headline: rollHeadline(card.Context),
		Detail:   a.rollDetail(card.Context),
		Dice:     settledDice(card),
		Applied:  appliedNotes(card.Applied),
		Rerolls:  rerolls,
	}
}

// rollLine is a settled roll's feed line.
func rollLine(card hooks.PostRollCard) string {
	line := fmt.Sprintf("%s: %s", rollHeadline(card.Context), settledDice(card))
	if notes := appliedNotes(card.Applied); len(notes) > 0 {
		line += " (" + strings.Join(notes, "; ") + ")"
	}
	return line
}

func (a *Adapter) emitRollLine(card hooks.PostRollCard) {
	a.emit(EvtRollSettled, RollSettledPayload{Line: rollLine(card)})
}

func rollHeadline(ctx hooks.RollContext) string {
	return fmt.Sprintf("%s rolls %s %s", ctx.Actor.Name, ctx.Attribute, rollPhaseWords[ctx.Phase])
}

func (a *Adapter) rollDetail(ctx hooks.RollContext) string {
	var parts []string
	if ctx.Asset != nil {
		parts = append(parts, a.assetName(ctx.Asset))
	}
	if ctx.Action != "" {
		parts = append(parts, ctx.Action)
	}
	if ctx.World != "" {
		parts = append(parts, "on "+a.worldName(ctx.World))
	}
	if ctx.Opponent != nil {
		opponent := "against " + ctx.Opponent.Name
		if ctx.OpponentAsset != nil {
			opponent += "'s " + a.assetName(ctx.OpponentAsset)
		}
		parts = append(parts, opponent)
	}
	return strings.Join(parts, " · ")
}

func settledDice(card hooks.PostRollCard) string {
	return fmt.Sprintf("%v %+d = %d%s", card.Result.Dice, card.Result.Modifier, card.Result.Sum, versus(card.Context))
}

func versus(ctx hooks.RollContext) string {
	if ctx.OpponentTotal == nil {
		return ""
	}
	return fmt.Sprintf(" vs %d", *ctx.OpponentTotal)
}

func struckLine(row hooks.StruckOffer) string {
	if row.StruckBy == "" {
		return row.Offer.Source + ": used this turn"
	}
	return fmt.Sprintf("%s: struck by %s", row.Offer.Source, row.StruckBy)
}

func appliedNotes(applied []hooks.AppliedReroll) []string {
	notes := make([]string, 0, len(applied))
	for _, reroll := range applied {
		notes = append(notes, fmt.Sprintf("%s rerolled a %s", reroll.Source, joinDice(reroll.Was)))
	}
	return notes
}

func diceAt(dice, indices []int) string {
	values := make([]int, 0, len(indices))
	for _, index := range indices {
		if index >= 0 && index < len(dice) {
			values = append(values, dice[index])
		}
	}
	return joinDice(values)
}

func joinDice(values []int) string {
	words := make([]string, 0, len(values))
	for _, value := range values {
		words = append(words, strconv.Itoa(value))
	}
	return strings.Join(words, " and ")
}

func (a *Adapter) factionName(factionID string) string {
	if faction := a.factionState.Factions[factionID]; faction != nil {
		return faction.Name
	}
	return factionID
}

func (a *Adapter) assetName(asset *domain.Asset) string {
	if def := a.engine.Rulebook.Assets[asset.DefinitionID]; def != nil {
		return def.Name
	}
	return asset.DefinitionID
}

func (a *Adapter) worldName(worldID string) string {
	if location, ok := a.engine.World.Location(worldID); ok {
		return location.Name()
	}
	return worldID
}
```

##### Task 3 — `internal/faction/tui/adapter/observer.go` and `adapter.go`

Hoist `emit`, so a collector can reach the feed.

In `observer.go`, delete:
```go
func (o *observer) emit(kind EventKind, payload any) {
	select {
	case o.adapter.eventCh <- ObserverEventMsg{Kind: kind, Payload: payload}:
	default:
		o.adapter.log.Warn("observer event dropped — eventCh full", "kind", kind)
	}
}

```
Then replace every `o.emit(` with `o.adapter.emit(`.

In `adapter.go`, **find:**
```go
func (a *Adapter) SetPerFactionCadence(on bool) { a.perFactionCadence.Store(on) }
```
**Replace with:**
```go
func (a *Adapter) SetPerFactionCadence(on bool) { a.perFactionCadence.Store(on) }

// emit sends a notification without blocking: a full buffer drops it and logs.
func (a *Adapter) emit(kind EventKind, payload any) {
	select {
	case a.eventCh <- ObserverEventMsg{Kind: kind, Payload: payload}:
	default:
		a.log.Warn("observer event dropped — eventCh full", "kind", kind)
	}
}
```

##### Task 4 — `internal/faction/tui/adapter/action_collector.go`

The shim now leaves the feed line.

**Find:**
```go
		if confirmed, _ := raw.(bool); confirmed {
			return &card.Rerolls[index]
		}
	}
	return nil
}
```
**Replace with:**
```go
		if confirmed, _ := raw.(bool); confirmed {
			return &card.Rerolls[index]
		}
	}
	a.adapter.emitRollLine(card)
	return nil
}
```

##### Task 5 — `internal/faction/tui/views/turn/execution/format.go`

**Find:**
```go
	case adapter.EvtError:
		return styles.Danger.Render("error: " + msg.Payload.(adapter.ErrorPayload).Err.Error()), nil
```
**Replace with:**
```go
	case adapter.EvtRollSettled:
		return phaseHeader(msg.Payload.(adapter.RollSettledPayload).Line), nil
	case adapter.EvtError:
		return styles.Danger.Render("error: " + msg.Payload.(adapter.ErrorPayload).Err.Error()), nil
```

##### Task 6 — Tests

`internal/faction/tui/adapter/roll_card_test.go` (new, package `adapter`), building
`New(harness.Engine, harness.FactionState, harness.Paths, slog.New(slog.DiscardHandler))` on
`testharness.NewHarness(t)`:

- **Headline and detail** — a defense roll names the roller, stat, phase words, asset, action,
  world and opponent.
- **Versus** — the dice text has "vs 9" with an `OpponentTotal`, and none without one.
- **Offer rows** — each names its description, source and owner's name.
- **Struck rows** — "Warlike: struck by Blackmail" and "Warlike: used this turn".
- **Reroll rows** — "Reroll the 7 — Psychic Academy (forced), gamma", with the owner's name.
- **Applied note** — "Fanatical rerolled a 1", on the card and in the feed line.
- **Unknown names** — a missing faction or definition falls back to its ID.
- **Plain values** — a `reflect` walk over both payload types finds only strings and string
  slices (D-12).
- **Feed** — `SelectReroll` with no rerolls sends one `EvtRollSettled` and does not ask.

`format_test.go`: `renderEvent` of an `EvtRollSettled` holds its line.

##### Commit message

```
feat: adapter builds roll card views

- each card becomes plain words with no IDs or pointers (D-12)
- every roll with a collector leaves one feed line, in the card's wording
- emit moves to the Adapter so a collector can reach the feed
```

#### Commit 5 — `feat: roll card overlays`

##### Task 1 — `internal/faction/tui/adapter/channels.go`

Rename in place, so `AskKind` does not reorder.

**Find:**
```go
	AskSelectModifiers
	AskConfirmReroll
```
**Replace with:**
```go
	AskPreRollCard
	AskPostRollCard
```

Delete:
```go
type SelectModifiersPayload struct{ Offers []hooks.ModifierOffer }
type ConfirmRerollPayload struct{ Directive hooks.RerollDirective }
```

##### Task 2 — `internal/faction/tui/adapter/action_collector.go`

Replace the `hooks.Collector` section with:
```go
// --- hooks.Collector (no error return; a bad reply declines and reports, Decision 1) ---

func (a *actionCollector) SelectModifiers(card hooks.PreRollCard) []hooks.ModifierOffer {
	raw, err := a.adapter.ask(AskPreRollCard, nil, a.adapter.preRollPayload(card))
	if err != nil {
		a.adapter.reportCardFailure(fmt.Errorf("SelectModifiers: %w", err))
		return nil
	}
	indices, ok := raw.([]int)
	if !ok {
		a.adapter.reportCardFailure(fmt.Errorf("SelectModifiers: unexpected reply type %T", raw))
		return nil
	}
	chosen := make([]hooks.ModifierOffer, 0, len(indices))
	for _, index := range indices {
		if index < 0 || index >= len(card.Offers) {
			a.adapter.reportCardFailure(fmt.Errorf("SelectModifiers: no offer %d", index))
			return nil
		}
		chosen = append(chosen, card.Offers[index])
	}
	return chosen
}

// SelectReroll waits only on a choice (D-5). With nothing on offer, or on a pass, the
// dice go to the feed.
func (a *actionCollector) SelectReroll(card hooks.PostRollCard) *hooks.RerollDirective {
	if len(card.Rerolls) == 0 {
		a.adapter.emitRollLine(card)
		return nil
	}
	raw, err := a.adapter.ask(AskPostRollCard, nil, a.adapter.postRollPayload(card))
	if err != nil {
		a.adapter.reportCardFailure(fmt.Errorf("SelectReroll: %w", err))
		a.adapter.emitRollLine(card)
		return nil
	}
	choice, ok := raw.(int)
	if !ok || choice < KeepRoll || choice >= len(card.Rerolls) {
		a.adapter.reportCardFailure(fmt.Errorf("SelectReroll: unexpected reply %#v", raw))
		a.adapter.emitRollLine(card)
		return nil
	}
	if choice == KeepRoll {
		a.adapter.emitRollLine(card)
		return nil
	}
	return &card.Rerolls[choice]
}

// reportCardFailure makes a bad card reply loud without stopping the roll: only a
// code bug can cause one, and hooks.Collector has no error to return.
func (a *Adapter) reportCardFailure(err error) {
	a.log.Error("roll card failed", "err", err)
	a.emit(EvtError, ErrorPayload{Err: err})
}
```

##### Task 3 — `internal/faction/tui/views/turn/overlay/rollcard.go` (new)

Delete `overlay/hooks.go`. Full contents of the new file:
```go
package overlay

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// rollCardRows is how many rows a card lists at once. Three fact rows, a blank, these
// rows, a blank and the prompt fill the 14-row wizard band.
const rollCardRows = 8

// visibleRows returns the rows to draw, so the cursor stays on screen.
func visibleRows(cursor, count int) (first, last int) {
	if cursor >= rollCardRows {
		first = cursor - rollCardRows + 1
	}
	return first, min(first+rollCardRows, count)
}

func moveCursor(pressed string, cursor, count int) int {
	switch pressed {
	case "up", "k":
		return max(cursor-1, 0)
	case "down", "j":
		return min(cursor+1, max(count-1, 0))
	}
	return cursor
}

func cursorMark(selected bool) string {
	if selected {
		return "> "
	}
	return "  "
}

func rollFacts(headline, detail, dice string) []string {
	return []string{styles.Strong.Render(headline), styles.Subtle.Render(detail), styles.AccentAlt.Render(dice)}
}

// PreRollCard is the before card. It only draws what the adapter worded, and its
// answer is the indices of the offers taken.
type PreRollCard struct {
	card   adapter.PreRollCardPayload
	cursor int
	chosen map[int]bool
}

func NewPreRollCard(card adapter.PreRollCardPayload) PreRollCard {
	return PreRollCard{card: card, chosen: make(map[int]bool)}
}

func (o PreRollCard) Init() tea.Cmd { return nil }

func (o PreRollCard) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	keyPress, ok := msg.(tea.KeyMsg)
	if !ok {
		return o, nil
	}
	// deliberate: no esc case, not today's "none", because a stray key must not throw away a bonus die (D-7).
	switch keyPress.String() {
	case " ":
		if o.cursor < len(o.card.Offers) {
			o.chosen[o.cursor] = !o.chosen[o.cursor]
		}
	case "enter":
		answer := make([]int, 0, len(o.chosen))
		for index := range o.card.Offers {
			if o.chosen[index] {
				answer = append(answer, index)
			}
		}
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	default:
		o.cursor = moveCursor(keyPress.String(), o.cursor, len(o.card.Offers)+len(o.card.Struck))
	}
	return o, nil
}

func (o PreRollCard) View() string {
	lines := rollFacts(o.card.Headline, o.card.Detail, o.card.Dice)
	lines = append(lines, "")
	count := len(o.card.Offers) + len(o.card.Struck)
	first, last := visibleRows(o.cursor, count)
	for row := first; row < last; row++ {
		mark := cursorMark(row == o.cursor)
		if row >= len(o.card.Offers) {
			lines = append(lines, styles.Dim.Render(mark+"--- "+o.card.Struck[row-len(o.card.Offers)]))
			continue
		}
		box := "[ ] "
		if o.chosen[row] {
			box = "[x] "
		}
		lines = append(lines, mark+box+o.card.Offers[row])
	}
	if count > 0 {
		lines = append(lines, "")
	}
	// deliberate: the keys are on the card, not only in Help, because nothing draws an overlay's Help and a hand-drawn card has no huh hint line.
	prompt := "Press enter to roll"
	if len(o.card.Offers) > 0 {
		prompt = "↑/↓ move, space toggles, enter rolls"
	}
	lines = append(lines, styles.Subtle.Render(prompt))
	return strings.Join(lines, "\n")
}

func (o PreRollCard) Help() help.KeyMap { return preRollHelp{} }

// PostRollCard is the after card. Its answer is an index into the rerolls, or
// adapter.KeepRoll. The cursor opens on Keep, so a stray enter spends nothing, and Esc
// does nothing (D-7).
type PostRollCard struct {
	card   adapter.PostRollCardPayload
	cursor int
}

func NewPostRollCard(card adapter.PostRollCardPayload) PostRollCard {
	return PostRollCard{card: card, cursor: len(card.Rerolls)}
}

func (o PostRollCard) Init() tea.Cmd { return nil }

func (o PostRollCard) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	keyPress, ok := msg.(tea.KeyMsg)
	if !ok {
		return o, nil
	}
	if keyPress.String() == "enter" {
		answer := adapter.KeepRoll
		if o.cursor < len(o.card.Rerolls) {
			answer = o.cursor
		}
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	o.cursor = moveCursor(keyPress.String(), o.cursor, len(o.card.Rerolls)+1)
	return o, nil
}

func (o PostRollCard) View() string {
	lines := rollFacts(o.card.Headline, o.card.Detail, o.card.Dice)
	for _, note := range o.card.Applied {
		lines = append(lines, styles.Subtle.Render(note))
	}
	lines = append(lines, "")
	first, last := visibleRows(o.cursor, len(o.card.Rerolls)+1)
	for row := first; row < last; row++ {
		text := "Keep"
		if row < len(o.card.Rerolls) {
			text = o.card.Rerolls[row]
		}
		lines = append(lines, fmt.Sprintf("%s%s", cursorMark(row == o.cursor), text))
	}
	lines = append(lines, "", styles.Subtle.Render("↑/↓ move, enter chooses"))
	return strings.Join(lines, "\n")
}

func (o PostRollCard) Help() help.KeyMap { return postRollHelp{} }

type preRollHelp struct{}

func (preRollHelp) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑/↓", "move")),
		key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "roll")),
	}
}
func (preRollHelp) FullHelp() [][]key.Binding { return [][]key.Binding{preRollHelp{}.ShortHelp()} }

type postRollHelp struct{}

func (postRollHelp) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑/↓", "move")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "choose")),
	}
}
func (postRollHelp) FullHelp() [][]key.Binding { return [][]key.Binding{postRollHelp{}.ShortHelp()} }

var (
	_ Overlay = PreRollCard{}
	_ Overlay = PostRollCard{}
)
```

##### Task 4 — `internal/faction/tui/views/turn/execution/execution.go`

**Find:**
```go
	adapter.AskSelectModifiers: {"Pre-roll Modifiers", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewSelectModifiers(msg.Payload.(adapter.SelectModifiersPayload).Offers)
	}},
	adapter.AskConfirmReroll: {"Reroll?", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewConfirmReroll(msg.Payload.(adapter.ConfirmRerollPayload).Directive)
	}},
```
**Replace with:**
```go
	adapter.AskPreRollCard: {"Roll", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewPreRollCard(msg.Payload.(adapter.PreRollCardPayload))
	}},
	adapter.AskPostRollCard: {"Reroll?", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewPostRollCard(msg.Payload.(adapter.PostRollCardPayload))
	}},
```

Added in execution (Decision Record 5 and 6). **Find:**
```go
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Top, m.overlay.View())
}
```
**Replace with:**
```go
	horizontal := lipgloss.Center
	// deliberate: roll cards sit flush left, not centered, because Place centers each short line alone and the card's rows would not line up.
	if m.activeAsk == adapter.AskPreRollCard || m.activeAsk == adapter.AskPostRollCard {
		horizontal = lipgloss.Left
	}
	return lipgloss.Place(width, height, horizontal, lipgloss.Top, m.overlay.View())
}

// streamContent wraps the feed to the stream's width. The viewport cuts off a line
// that is too long, and a roll's feed line can run past the center column.
func (m Model) streamContent() string {
	return lipgloss.NewStyle().Width(m.stream.Width).Render(strings.Join(m.lines, "\n"))
}
```
Then replace both `m.stream.SetContent(strings.Join(m.lines, "\n"))` calls with
`m.stream.SetContent(m.streamContent())`.

##### Task 5 — Tests

`overlay/rollcard_test.go` (new), driving `Update` with `tea.KeyMsg` and running the returned
command:

- **Toggle one row (`B.029`)** — two offers with the same label: `space` on the second, then
  `enter`, answers `[]int{1}`.
- **No offers** — `enter` answers an empty, non-nil `[]int`.
- **Struck rows** — the cursor can reach one, `space` does nothing, and the row shows.
- **Esc (D-7)** — `esc` returns no command on both cards.
- **After card opens on Keep** — `enter` straight away answers `adapter.KeepRoll`.
- **Take a reroll** — `up`, `enter` answers the last reroll's index.
- **Scrolling** — with 12 offers and the cursor on row 11, the view holds row 11 and not row 0.
- **Height** — with 12 offers, the before card's view is at most 14 lines.
- **Keys** — the before card with offers says "space toggles", with none "Press enter to roll",
  and the after card "enter chooses" (Decision Record 7).

`execution/execution_test.go` (Decision Record 5 and 6):

- **Flush left** — on both cards, the band's first line starts with the headline.
- **Feed wrap** — the end of a long `EvtRollSettled` line reaches the stream.

`adapter/roll_card_test.go`, new cases through `Action()`, answering on `AskPump`:

- **Taken offers** — `[]int{1}` returns only the second offer.
- **Reroll taken** — `0` returns the first reroll and leaves no feed line.
- **Keep** — `KeepRoll` returns nil and leaves one feed line.
- **Bad replies** — a wrong type and an out-of-range index each decline, and each sends an
  `EvtError` on the event channel.

##### Task 6 — Manual check

`go build -o gm-toolkit ./cmd/gm-toolkit`, then run `--campaign sandbox-claude` through tmux with
the `combat` scenario. Attack with an asset, and check the before card, the feed line, `vs N` on
the defense roll, and Esc doing nothing. Log anything that does not fit in the Decision Record.

##### Commit message

```
feat: roll card overlays

- two hand-drawn cards in the wizard band replace overlay/hooks.go (D-9, Decision 7)
- the before card always waits; the after card waits only on a choice (D-5)
- Esc does nothing, and the after card opens on Keep (D-7)
- a bad reply declines the roll's choice and reports it in the status line
- the asks rename in place, so AskKind does not reorder
- the manual check's fixes: cards sit flush left, draw their own keys, and the feed wraps
```

### Phase 3 — Psychic Academy and rigor

**In plain words:** Psychic Academy's reroll goes live now that a prompt can show the roll. Then
every collector reply gets its type checked, the adapter gets its first tests, and the pages
catch up with the code.

#### Commit 6 — `feat: Psychic Academy's reroll entry`

##### Task 1 — `rulebooks/swn/tags.toml`

**Find:**
```toml
effect = "Can provide psionic mentor training. Once per turn, force a rival faction to reroll any one d10, whether or not they are involved in the roll."
```
**Replace with:**
```toml
effect = "Can provide psionic mentor training. Once per turn, force a rival faction to reroll any one d10, whether or not they are involved in the roll."

[[tags.PsychicAcademy.effects]]
trigger = "rival_roll"
effect  = "reroll"
[tags.PsychicAcademy.effects.filters]
owner = "rival"
scope = "faction"
[tags.PsychicAcademy.effects.gates]
once_per = "owner_faction"
[tags.PsychicAcademy.effects.reroll]
dice     = 1
elective = true
```

##### Task 2 — `campaigns/sandbox/rulebook/tags.toml`

The same edit as Task 1 (Decision 9).

##### Task 3 — `campaigns/sandbox/TESTING.md`

**Find:**
```markdown
after the capture, the Book of Secrets reroll on Mireholt is offered to Veiled Hand, not to Ashfall.
```
**Replace with:**
```markdown
after the capture, the Book of Secrets row on a Mireholt roll card names Veiled Hand as its owner, not Ashfall.
```

##### Task 4 — `internal/faction/engine/testharness/scenarios/tags_test.go`

New scenarios, in the style of `TestScavengers_GrantsCoinOnKill`: alpha attacks beta on Tartarus,
gamma holds `SWN-T-014` on another world, and `h.Engine.Rand` is a `dicetest.Cycle`.

- **Taken** — the default collector takes the reroll: alpha's high attack die becomes a low one,
  the attack misses, gamma's `effect:reroll:SWN-T-014` is spent once, and beta's defense roll is
  not offered a second reroll.
- **Declined (A.001.7 D-7, D-3)** — `SelectRerollFn` counts the cards with rerolls and returns
  nil: both rolls are offered, the budget stays unspent, and the dice stand.
- **Holder's own roll** — gamma is not offered its own reroll on its own roll.
- **Bookkeeping clears it** — after gamma's bookkeeping, the budget is gone.

##### Commit message

```
feat: Psychic Academy's reroll entry

- the last SWN tag with no data entry goes live, in the rulebook and the sandbox copy
- the card names the Academy's owner on each rival roll (B.043)
- TESTING.md rewords the Book of Secrets step for the card
```

#### Commit 7 — `refactor: check every collector reply type`

##### Task 1 — `internal/faction/tui/adapter/phase_collector.go`

**Find:**
```go
func (p *phaseCollector) ask(kind AskKind, faction *domain.Faction, payload any) (any, error) {
	return p.adapter.ask(kind, faction, payload)
}
```
**Replace with:**
```go
// replyAs reads an ask's reply as T. An untyped nil is a legal empty answer (a skip, or
// a decline), and any other type is a code bug that must not pass as a zero value.
func replyAs[T any](method string, raw any) (T, error) {
	var empty T
	if raw == nil {
		return empty, nil
	}
	reply, ok := raw.(T)
	if !ok {
		return empty, fmt.Errorf("%s: unexpected reply type %T", method, raw)
	}
	return reply, nil
}
```

Replace every `p.ask(` with `p.adapter.ask(`.

Then replace the four methods' tails:
```go
// SelectAction
	return replyAs[action.Action]("SelectAction", raw)
// SelectStatRaise
	return replyAs[*domain.FactionStat]("SelectStatRaise", raw)
// SelectMovementDecisions (the nil guard comes with replyAs)
	return replyAs[[]world.MovementDecision]("SelectMovementDecisions", raw)
// SelectTransportCargo
	return replyAs[[]*domain.Asset]("SelectTransportCargo", raw)
```
Each line replaces everything from `if raw == nil` or `chosen, ok := raw.(…)` down to the
method's closing brace. The `if err != nil` above it stays.

##### Task 2 — `internal/faction/tui/adapter/action_collector.go`

In each method below, the listed lines replace the `raw.(…)` assertion and the `return` after
it. The `if err != nil` above each stays.

```go
// SelectAsset
	return replyAs[*domain.Asset]("SelectAsset", raw)
// SelectAssetToLose
	return replyAs[*domain.Asset]("SelectAssetToLose", raw)
// SelectAbilityTarget, TargetFaction
		chosen, err := replyAs[*domain.Faction]("SelectAbilityTarget", raw)
		return domain.AbilityTarget{Kind: domain.TargetFaction, Faction: chosen}, err
// SelectAbilityTarget, TargetAsset
		chosen, err := replyAs[*domain.Asset]("SelectAbilityTarget", raw)
		return domain.AbilityTarget{Kind: domain.TargetAsset, Asset: chosen}, err
// SelectAbilityTarget, TargetWorld
		chosen, err := replyAs[string]("SelectAbilityTarget", raw)
		return domain.AbilityTarget{Kind: domain.TargetWorld, World: chosen}, err
// selectRelocateOrders
	return replyAs[[]action.RelocateOrder]("SelectRelocateOrders", raw)
// SelectRepairOrders
	return replyAs[[]action.RepairOrder]("SelectRepairOrders", raw)
// SelectBuyOrder
	return replyAs[action.BuyOrder]("SelectBuyOrder", raw)
// SelectRefitOrder
	return replyAs[action.RefitOrder]("SelectRefitOrder", raw)
// SelectAttackers
	return replyAs[[]*domain.Asset]("SelectAttackers", raw)
// SelectDefender
	return replyAs[*domain.Asset]("SelectDefender", raw)
// ConfirmRedirectToBase
	return replyAs[bool]("ConfirmRedirectToBase", raw)
// ConfirmCounterstrike
	return replyAs[bool]("ConfirmCounterstrike", raw)
// ConfirmSacrifice
	return replyAs[bool]("ConfirmSacrifice", raw)
// requestPermission
	return replyAs[bool]("RequestPermission", raw)
// confirmPermissionContest
	return replyAs[bool]("ConfirmPermissionContest", raw)
// confirmPayment
	return replyAs[bool]("ConfirmPayment", raw)
// SelectExpandInfluenceOrder
	return replyAs[action.ExpandInfluenceOrder]("SelectExpandInfluenceOrder", raw)
// ConfirmRivalFreeAttack
	return replyAs[bool]("ConfirmRivalFreeAttack", raw)
// SelectBaseAttackers
	return replyAs[[]*domain.Asset]("SelectBaseAttackers", raw)
// SelectAbilityHolders
	return replyAs[[]domain.AbilityHolder]("SelectAbilityHolders", raw)
// ConfirmAbilityApplied
	return replyAs[bool]("ConfirmAbilityApplied", raw)
// SelectBribeTarget
	reply, err := replyAs[BribeReply]("SelectBribeTarget", raw)
	return reply.Base, reply.Amount, err
// SelectSeizeTarget
	return replyAs[string]("SelectSeizeTarget", raw)
// SelectChangeHomeworldTarget
	return replyAs[string]("SelectChangeHomeworldTarget", raw)
// SelectPurchaseEnablerAsset
	return replyAs[*domain.AssetDefinition]("SelectPurchaseEnablerAsset", raw)
```

The two card methods keep their own checks from Commit 5: an untyped nil is not a legal card
answer, because `replyAs` would read it as index 0 and spend a reroll.

##### Task 3 — `internal/faction/tui/adapter/collector_test.go` (new, package `adapter_test`)

Shared pieces, in full:
```go
package adapter_test

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
)

func newBridge(t *testing.T) (*adapter.Adapter, *testharness.Harness) {
	t.Helper()
	harness := testharness.NewHarness(t)
	return adapter.New(harness.Engine, harness.FactionState, harness.Paths, slog.New(slog.DiscardHandler)), harness
}

// answerNext answers the next ask on the channel, as the execution view does.
func answerNext(bridge *adapter.Adapter, reply any) {
	go func() {
		ask := bridge.AskPump()().(adapter.CollectorAskMsg)
		ask.Reply <- reply
	}()
}

type collectorCase struct {
	name string
	call func() (any, error)
	good any
	want any
}

func runCollectorCases(t *testing.T, bridge *adapter.Adapter, cases []collectorCase) {
	t.Helper()
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			answerNext(bridge, testCase.good)
			got, err := testCase.call()
			if err != nil || !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("good reply: got %#v, %v; want %#v", got, err, testCase.want)
			}

			answerNext(bridge, struct{}{})
			if _, err := testCase.call(); err == nil {
				t.Error("wrong reply type: got no error")
			}

			answerNext(bridge, nil)
			if _, err := testCase.call(); err != nil {
				t.Errorf("nil reply is a legal empty answer: got %v", err)
			}
		})
	}
}
```

The table holds one case for each method through `bridge.Action()`, listed in Task 2, and one for
each of `bridge.Phase()`'s four replies in Task 1, plus `ConfirmPayment`, `RequestPermission`,
`ConfirmPermissionContest` and `SelectRelocateOrders` through `Phase()` as well. Each `call`
wraps the method so it returns `(any, error)`; a method with two results returns its reply shape
as `want`. Fixtures come from the harness: `AddFaction` for factions and assets, and
`harness.Engine.Rulebook` for definitions.

Tests beside the table:

- **`SelectAbilityTarget`** — one case for each of the three target kinds.
- **`AwaitCheckpoint`** — in per-faction cadence it asks; otherwise only `cycle_summary` asks.
- **Canceled** — `action.ErrTurnCanceled` as a reply comes back as that error, through `Action()`
  and `Phase()`.

##### Task 4 — Run the sweep's check

`go test ./internal/faction/tui/...` must pass, and every overlay's answer type must match its
method's `replyAs` type. The session's read of the overlays found them matching, but this commit
is what proves it at the table, so run one attack and one Expand Influence in
`sandbox-claude` as well.

##### Commit message

```
refactor: check every collector reply type

- replyAs replaces each unchecked raw.(T) in both collectors (R.015)
- an untyped nil stays a legal empty answer; any other wrong type is an error
- SelectMovementDecisions accepts nil, and the phaseCollector.ask wrapper goes
- first adapter tests: one table over every collector method (F.027, in part)
```

#### Commit 8 — `test: full cycle through the real channels under race`

##### Task 1 — `internal/faction/tui/views/turn/execution/cycle_test.go` (new)

Full contents:
```go
package execution

import (
	"log/slog"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice/dicetest"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/turn/overlay"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

func sandboxMap(t *testing.T) *spatial.RegionMap {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "..", "..")
	regionMap, err := spatial.LoadRegionMap(filepath.Join(repoRoot, "campaigns", "sandbox", "spatial"))
	if err != nil {
		t.Fatalf("LoadRegionMap: %v", err)
	}
	return regionMap
}

// answerFor answers each ask as a GM would: alpha attacks, and everyone else passes.
func answerFor(t *testing.T, ask adapter.CollectorAskMsg) any {
	t.Helper()
	switch ask.Kind {
	case adapter.AskAwaitCheckpoint:
		return true
	case adapter.AskSelectStatRaise:
		return nil
	case adapter.AskSelectMovementDecisions:
		return []world.MovementDecision{}
	case adapter.AskSelectAction:
		if ask.Faction.ID != "alpha" {
			return nil
		}
		for _, candidate := range ask.Payload.(adapter.SelectActionPayload).Available {
			if candidate.Name() == "Attack" {
				return candidate
			}
		}
		t.Fatal("Attack not available for alpha")
	case adapter.AskSelectAttackers:
		return ask.Payload.(adapter.SelectAttackersPayload).Eligible
	case adapter.AskSelectDefender:
		return ask.Payload.(adapter.SelectDefenderPayload).Eligible[0]
	case adapter.AskConfirmRedirectToBase, adapter.AskConfirmCounterstrike:
		return false
	case adapter.AskPreRollCard:
		return []int{}
	case adapter.AskPostRollCard:
		return adapter.KeepRoll
	}
	t.Fatalf("no answer for AskKind %d", ask.Kind)
	return nil
}

// TestCycleThroughTheRealChannels drives one full cycle through the adapter and the
// real view. Its job is to run under -race, where snapshotOf reads each live faction
// pointer the way B.038 suspects. A race found here is logged on B.038, not fixed.
func TestCycleThroughTheRealChannels(t *testing.T) {
	harness := testharness.NewHarness(t)
	harness.AddFaction("alpha", "Tartarus", 4, 3, 2)
	harness.AddFaction("beta", "Tartarus", 2, 2, 2)
	harness.Engine.Rand = &dicetest.Cycle{Values: []int{7, 3}}
	bridge := adapter.New(harness.Engine, harness.FactionState, harness.Paths, slog.New(slog.DiscardHandler))
	model := New(120, 40, harness.Engine.Rulebook, sandboxMap(t))

	messages := make(chan tea.Msg)
	go func() { messages <- bridge.Run()() }()
	pump := func(next tea.Cmd) {
		for {
			msg := next()
			if msg == nil {
				return
			}
			messages <- msg
			if _, closed := msg.(adapter.StreamClosedMsg); closed {
				return
			}
		}
	}
	go pump(bridge.ObserverPump())
	go pump(bridge.AskPump())

	engineDone, streamClosed := false, false
	timeout := time.After(30 * time.Second)
	for !engineDone || !streamClosed {
		select {
		case msg := <-messages:
			switch typed := msg.(type) {
			case adapter.EngineDoneMsg:
				engineDone = true
				if typed.Err != nil {
					t.Fatalf("RunCycle: %v", typed.Err)
				}
			case adapter.StreamClosedMsg:
				streamClosed = true
			case adapter.CollectorAskMsg:
				model, _ = model.Update(typed)
				model, _ = model.Update(overlay.OverlayDoneMsg{Answer: answerFor(t, typed)})
			default:
				model, _ = model.Update(msg)
			}
		case <-timeout:
			t.Fatal("the cycle did not finish in 30 seconds")
		}
	}
	contains(t, strings.Join(model.lines, "\n"), " rolls ", "completed")
}
```

Then run `go test -race ./internal/faction/tui/views/turn/execution -run TestCycleThroughTheRealChannels`.

- **No race reported** — note it on the `B.038` row in `docs/planned-work.md`.
- **A race reported** — copy the two stacks' top frames to the `B.038` row. Do not fix it. The
  plain `go test` gate still passes, because the race detector runs only under `-race`.

##### Commit message

```
test: full cycle through the real channels under race

- one cycle through the adapter's channels and the real execution view
- the first check of B.038's suspected read; its result is logged on the row
```

#### Commit 9 — `docs: roll card architecture pages`

Load `/gestalt:style-guide` first. These edits change only what this work makes false; `D.010`
keeps the wider refresh. Check each claim against the source as committed.

##### Task 1 — `docs/architecture/engine/dice.md`

**Find:**
```markdown
3. Drop any offer whose budget is spent, then let the
   [collector](hooks.md#the-collector-seam) choose.
```
**Replace with:**
```markdown
3. Set aside any offer whose budget is spent, then show the
   [collector](hooks.md#the-collector-seam) a before card with what survives and what
   did not. It asks on every roll, offers or not.
```

**Find:**
```markdown
6. Run each Cat 2 hook. A directive whose budget is spent is skipped, an elective
   one asks the collector, and a forced one applies. A rerolled die rolls its own
   sides again.
```
**Replace with:**
```markdown
6. Run each forced Cat 2 hook, once. A directive whose budget is spent is skipped. A
   rerolled die rolls its own sides again.
7. Show the collector an after card with the dice, the rerolls applied and each
   elective reroll on offer. It takes one or passes. After a take, the card returns
   with what is left: a taken hook is not offered again on this roll, and a forced
   hook does not run again, so a reroll that lands on 1 stays.
```

Add these Key Decisions after "Elective rerolls keep per-roll timing":
```markdown
- **Every hooked roll with a collector shows the cards, choice or not.** The GM sees
  the dice about to roll, as in Foundry VTT, and the before card always waits. The
  after card waits only on a choice: `SelectReroll` is called every time, and the
  adapter sends the dice to the feed when nothing is on offer, so `dice` needs no link
  to the observer. A nil registry has no hooks and asks no one. Rejected: cards only
  on a choice, which leaves the dice unseen. (`F.039` D-4, D-5, Decisions 2 and 14.)
- **A bonus die's keep comes before any reroll, and forced rerolls before elective
  ones.** Each card shows one kept die, so the card has no die picker and `RollResult`
  carries no dropped die. Forced rerolls run once. (`F.039` D-1, D-2, D-6.)
- **A later roll's card shows the total already rolled.** `Side.Total` reaches
  `RollContext.OpponentTotal`, set by `Contest` and by Expand Influence's rival rolls.
  (`F.039` Decision 6.)
```

##### Task 2 — `docs/architecture/engine/hooks.md`

**Find:**
```go
type Collector interface {
	SelectModifiers(offers []ModifierOffer) []ModifierOffer
	ConfirmReroll(directive RerollDirective) bool
}
```
**Replace with:**
```go
type Collector interface {
	SelectModifiers(card PreRollCard) []ModifierOffer
	SelectReroll(card PostRollCard) *RerollDirective
}
```

**Find:**
```markdown
the roll offers the Cat 1 modifiers that survive the offer filters and budgets, and the collector returns the
subset to apply; for each *elective* Cat 2 directive the collector confirms or
declines. Non-elective directives apply automatically. `SelectModifiers` is not
called when no offer survives the budget filter.
```
**Replace with:**
```markdown
a before card holds the roll, the Cat 1 modifiers that survive the offer filters and
budgets, and those that did not, and the collector returns the subset to apply. An
after card holds the dice, the rerolls applied and the *elective* Cat 2 directives on
offer, and the collector returns one to take, or nil to keep the dice. Non-elective
directives apply automatically. Both are called on every roll with a collector. The
cards are structs the collector only reads, so a planner can answer them as a GM does.
```

**Find:**
```markdown
- **A budget lives on its owner's `HookBudgets`, keyed by effect, trigger,
  definition or tag ID, and (when `once_per` asks) target or asset.** A tag's key
  names the tag ID where an asset's names its definition
  (`effect:bonus_die:own_roll:SWN-T-020`; A.001.7 Effort 1 Decision Record 2). `HookBudgets` is
```
**Replace with:**
```markdown
- **A budget lives on its owner's `HookBudgets`, keyed by effect, definition or tag
  ID, and (when `once_per` asks) target or asset.** A tag's key
  names the tag ID where an asset's names its definition
  (`effect:bonus_die:SWN-T-020`; A.001.7 Effort 1 Decision Record 2). `HookBudgets` is
```

**Find:**
```markdown
not the roller's. The trigger is in the key so two entries of one definition
(Book of Secrets) do not share a budget. (s-flag-effect-registry plan, Planning
Decision 3; Effort 3 plan, Decisions 8 and 16.)
```
**Replace with:**
```markdown
not the roller's. The trigger is not in the key, so two entries of one effect on one
definition share a use: Book of Secrets has one per turn, on its own roll or a
rival's. An action that aborts restores every faction's budgets, because its
mutations never land. (s-flag-effect-registry plan, Planning Decision 3; Effort 3
plan, Decision 8; `F.039` Decisions 3 and 18, `B.016`, `B.024`.)
```

In "The collector seam keeps elective hooks out of the engine's hands", **find** `narrowed to two
methods.` and **replace with** `narrowed to two methods, one for each card.`

##### Task 3 — `docs/architecture/interface/overlays.md`

**Find:**
```markdown
`relocateorders`, `purchaseenabler`, the ability prompts — are bespoke `huh` forms
in their own files (~18 overlays in all), but
follow the identical pattern: build the form in the constructor, watch for
`huh.StateCompleted` in `Update`, emit `OverlayDoneMsg` with the collected value.
```
**Replace with:**
```markdown
`relocateorders`, `purchaseenabler`, the ability prompts — are bespoke `huh` forms
in their own files (~18 overlays in all), but
follow the identical pattern: build the form in the constructor, watch for
`huh.StateCompleted` in `Update`, emit `OverlayDoneMsg` with the collected value.

The two roll cards in `rollcard.go` use no form. Each draws its rows by hand and
handles `↑/↓` and `enter` itself, and the before card handles `space`, so rows with
one label toggle apart and a card with no offers needs no special case. They only
draw: the adapter words the whole card, and the answer is row indices.
```

**Find:**
```markdown
  the decline is an explicit option (`NoCancelConfirm`, `SelectAssetToLose`).
```
**Replace with:**
```markdown
  the decline is an explicit option (`NoCancelConfirm`, `SelectAssetToLose`). The roll
  cards ignore it too: Roll and Keep are explicit, and the after card opens on Keep.
```

Add to Key Decisions:
```markdown
- **The adapter words a roll card, and the overlay only draws it.** The card reaches
  the overlay as strings, with no IDs or pointers, and the feed's dice line uses the
  same wording. The cards sit in the wizard band, so the right panel stays free. A
  card lists at most 8 rows at once and scrolls with the cursor. A wrapped row or an
  applied note can still push a full card past the band (`F.039` Decision Record 9).
  (`F.039` D-9, D-12, Decisions 7 and 16.)
```

##### Task 4 — `docs/architecture/interface/event-stream.md`

**Find:**
```markdown
adapter-side, deliberately so the engine's `action` package stays untouched. The
engine goroutine is parked on the ask while this runs, so reading shared state is
race-free.
```
**Replace with:**
```markdown
adapter-side, deliberately so the engine's `action` package stays untouched. The
engine goroutine is parked on the ask while this runs, so reading shared state is
race-free.

The roll cards go furthest: `roll_card.go` turns each card into strings. A collector
method can also emit: `SelectReroll` sends `EvtRollSettled` with the card's dice line
when a roll settles, so each roll with a collector leaves one line in the feed. A
reply of the wrong type on a card ask declines and sends `EvtError`, because
`hooks.Collector` has no error to return. Every other collector method that reads a
reply does so through `replyAs`, which turns a wrong type into a returned error and
keeps an untyped nil as a legal empty answer.
```

Add to Key Decisions:
```markdown
- **A wrong reply type is loud.** No collector method that reads a reply takes a wrong
  one as a zero value. A card ask reports it and declines; the others return an error.
  (`R.015`; `F.039` Decision 1.)
```

##### Commit message

```
docs: roll card architecture pages

- dice and hooks: the two cards, reroll order, the opponent's total, the shared use
- overlays: the hand-drawn cards and their Esc rule
- event-stream: worded cards, the settled-roll line and reply checks
```

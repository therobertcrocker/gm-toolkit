# Dice Package — Implementation Plan

## In Plain Words

- **What:** a new package, `engine/dice`, becomes the one home for dice. Each type of game
  roll gets a named route: stat rolls (with attack-and-defense contests and gates), amount
  rolls, table rolls and target rolls. An ability's step says which type it rolls.
- **Why:** each new contest made us decide again how its roll should work. Eleven sites build
  the hooked roll by hand, two functions state one gate rule, and one ability field holds
  three types of roll. With named routes, the only question left is which type a roll is.
- **How it fits:** a refactor with no rule changes, so players see no change. It runs before
  the roll card (`R.025`), which then changes one place, and before H.A.W.K's authoring passes
  (`A.001.5`), which then use the new data shapes.

## Context / Goal

- Audit and target: [`E.004` exploration](../explorations/E.004-unifying-dice-rollhooks-exploration.md).
  There is no discovery doc, by design. `E.004` Terms and "Differences that must survive" bind
  this plan. The design calls C-5, P-4 and O-1 are settled there and are not re-opened here.
- Absorbs `B.010` and the `lock_out` case from `R.026`.
- Follow-on: [`R.025`](../../planned-work.md), `F.032`, `F.037`, `A.001.5`.

## Decisions Ratified in Planning

All are Robert's picks, 2026-11-03, unless marked as a planning call.

1. **Rerolls reach stat rolls only.** Fanatical ("any dice"), Book of Secrets ("one die for an
   action") and Psychic Academy ("any one d10") could reach amount dice on their plain words.
   The engine rerolls stat rolls only, and that reading holds. **This corrects `E.004`**, whose
   "Rules calls to carry: None" and "no rule changes an amount die" rested on a text search; a
   full read of the rulebook found these three. `E.004` stays as written; the reading goes in
   `dice.md`.
2. **A target roll is a nested `target_roll` block** with `die`, `target` and `amount`, the same
   keys as `xp_roll`. A flat `target` key is taken by `target = "world"` (`tags.toml:271`).
3. **The split step has three optional fields:** `Amount *DiceRoll`, `Table *RollTable`,
   `TargetRoll *TargetRoll`. The loader allows one at most, the pattern `EffectDefinition` uses.
4. **Amount abilities keep a bare `die`.** `die` alone is an amount; `die` with rows is a table.
   Loader checks refuse a `coin_delta` with neither rows nor a target roll, and a
   `purchase_discount` or `block_attack` with rows.
5. **Harvesters' `coin = 1` becomes the target roll's `amount`,** which is optional because Teach
   has none. The effect decides what the amount pays. "Beat" means meet or beat, as in
   `xp_roll.go:53`.
6. **`StepContext.RollResult` becomes `Rolled dice.StepResult`** (`Amount`, `Row`, `Passed`), filled
   where the die rolls today (`dispatch.go:144`): after the gate, before the handler.
7. **The lockdown's die moves to `[effects.lock_out] amount = "1d4"`,** on the shared
   `amountRecord`. `LockedOutBy` returns the asset too, walked in sorted order.
8. **The lockdown's cause becomes `string(domain.SpecialLockOut)`,** as every other effect names
   its cause and as `lock_out.go:62` already does. No code branches on the old text; the digest
   counts it as a plain Coin change (`digest/build.go:153`), so players see no change.
9. **One phase enum: `domain.RollPhase`.** `EffectRollPhase` is renamed with its text values kept,
   so TOML is unchanged. `hooks.RollPhase`, `hooks.Phase*` and `rollPhaseNames` go.
10. **Contest shape: `dice.Engine` plus `dice.Side`, with a `Gate` route.** `Engine` holds what the
    game carries; `Side` is who rolls. Same faction on both sides drops `Opponent`. `Gate` does the
    target-roller flip. `effects.contestRoll` goes away.
11. **No efforts.** Rolls are one design surface and this plan settles the design. One review, at
    the initiative close.

Planning calls, shown to Robert before writing:

12. **Literal code for what is new; a pattern and a site list for each conversion.** The refactor
    flow asks for this because per-site recipes go stale. Each conversion phase carries worked
    examples.
13. **`rollWithHooks` keeps its `faction` and base-roll parameters.** `roll_test.go` rolls 2d10 with
    a nil `Actor`, so it moves almost unchanged and guards the move. `Stat` is the only caller that
    builds the `1d10 + stat` roll.
14. **`ResolveTie` stays exported.** It is not a roll, and the `lose_ties` tests call it from
    outside the package.
15. **The sandbox re-seed rides in each commit that changes data,** not in its own commit as in
    `249db61`, so the sandbox loads at every commit.

## Decision Record — Execution

- **Plan doc and tracker land in Commit 1** — the Plan session left them uncommitted, and `git add -A` carries them with the first commit, not in a commit of their own
- **`Sequence` and `Cycle` gain `Rolled()`** (Commit 3) — seven assertions read the old fakes' private counts (`clear_stealth_test.go`, `attack_test.go`), which the planned `next` field hides; dropping them would lose the "rolled exactly this many dice" check
- **The adapter's `statRating` goes too** (Commit 5, Robert's pick) — the plan missed this third copy of the same switch in `tui/adapter/action_collector.go`, and its comment named the deleted `actions.statScore`; its one caller now uses `Faction.StatScore`. Rejected: fixing the comment only, which keeps the switch in two places
- **Two more stat conversions in tests** (Commit 5) — `expand_influence_test.go:438-441` drops `string(...)`, and the `cunningDefense` helper in `bonus_die_test.go` takes a `domain.FactionStat`; no expected value moved
- **A test parameter renamed** (Commit 6) — `rollThroughDispatcher` in `reroll_test.go` had a parameter called `dice`, which hides the package once the call reads `dice.RollWithHooks`; it is now `values`. No expected value moved
- **The import line also reaches `internal/spatial`** (Commit 8) — `go list -deps` on `engine/dice` shows it through `domain`, so the Shared Context claim of "nothing else in the project" was wrong; `spatial` imports nothing from the project, so the package still cannot close a cycle. Shared Context is corrected
- **Attack's line ranges were one low** (Commit 9) — the `dice` import from Commit 8 moved the code down a line; the matchup is `attack.go:124-167` and the counterstrike `:256-273`. Commit 9's tasks are corrected
- **An attack's tie context now carries `Action`** (Commit 9) — the old literals at the matchup and the counterstrike left it out and `Engine.Contest` fills it. No reader sees the difference: `lose_ties.go:31` is the only production resolver and reads only `Actor` and `Opponent`
- **Two gaps left as they are** (Commit 9) — the planned name `outcome` for the contest is reused by the `fireMatchup` result in the nested `own_defense_win` block; it compiles and nothing below reads the contest's, so the plan's name stays. `TestAttack_TieResolver_DefenderWins` is the only Attack-level tie-resolver test; a tie given to the attacker is held one level down, by `TestOutcome_Readings`
- **The free attack's tie context now carries `Action`** (Commit 10) — the same change as Commit 9: the old literal left it out and `Engine.Contest` fills it. No reader sees the difference, for the same reason
- **The rolls keep their old values** (Commit 10) — the expander's and each rival's modifier were `.Cunning` and are now `StatScore(domain.StatCunning)`, which returns the same field (`faction.go:156`). `RollsCarryContext` and `FreeAttackTieAsksResolvers` pass with no change to their code
- **One test comment reworded** (Commit 10) — `TestExpandInfluence_RollsCarryContext` named `RollWithHooks`, which goes private in Commit 12 and which Commit 12's list does not cover; it now says "the hooks". No expected value moved
- **Commit 12 Task 4 corrected** (Commit 10) — it still asked to rename `rollThroughDispatcher`'s `dice` parameter, which Commit 6 already did; that instruction is removed
- **The same-faction rule does not reach the ability gate** (Commit 11) — `Contest` drops `Opponent` when both sides are one faction and the old `passesGate` did not, but the two cannot meet: `otherFactions` and `assetCandidates` leave out the acting faction, and no step in `rulebooks/swn` pairs a contest with `target = "self"`

---

## Shared Context

**The routes.** A new roll calls one of these or extends one. It never builds its own.

| Type | Route | Hooks | Result |
|---|---|---|---|
| Stat roll | `Engine.Stat` | Yes | `hooks.RollResult` |
| Contest | `Engine.Contest` | Yes | `Outcome`: `AttackLands`, `CounterLands`, `GatePasses` |
| Gate | `Engine.Gate` | Yes | `bool` |
| Amount roll | `dice.Amount` | No | `int` |
| Table roll | `dice.Table` | No | The matched row, or none |
| Target roll | `dice.Target` | No | `bool` |
| Plain die | `Roller.Roll` | No | Turn order only (`turn/order.go:19`) |

**Import line.** `engine/dice` imports `domain`, `engine/hooks`, `rulebook` and `state` only.
`go list -deps` shows those four depend only on `internal/spatial` beyond each other, through
`domain`. `spatial` imports nothing from the project and none of them uses `Roller`, so the package
cannot close a cycle. `dicetest` imports nothing.

**What must survive.** Re-check these at each unit that touches them.

| Item | Where it lands |
|---|---|
| Attacker rolls first; the flip swaps who that is | `Contest`, `Gate` |
| Ability die rolls before the handler (`dispatch.go:144`) | `RollStep`, called at the same spot |
| Lockdown die rolls before the hold, even if unpaid (`arrival.go:60`) | `lockdownFee`, called at the same spot |
| Ties: `AttackLands` (Attack, counterstrike, free attack), `CounterLands` (Attack), `GatePasses` (both gates) | `Outcome` |
| Expand Influence is a plain comparison, with no resolver | Stays at its site |
| Empty `Action` on effect contests: Imperialists (`tags.toml:140`), `wealth_assets.toml:183` | `effects.contestAction` |

**Fence-signs that move or stay.**

| Fence-sign | Goes to |
|---|---|
| `roll.go:88`, per-roll reroll timing | `dice/roll.go`, unchanged |
| `attack.go:256`, no `Opponent` on a counterstrike | `Engine.Contest` |
| `expand_influence.go:95`, no `Opponent` on the expander's roll | Stays at the site |
| `ability/stage.go:104`, the real target faction | Stays in `passesGate` |
| `ability/stage.go:110`, a literal `Action` | Stays: `ability` still cannot import `actions` |
| `bonus_die.go:73`, keep-highest assumes one d10 | Stays |
| The nil collector at each of the three effect-gate callers | Stays at each caller, so `F.037` swaps one argument |

**Gate.** `go build ./...`, `go vet ./internal/faction/...`, `go test ./internal/faction/...`. The
baseline passes at `5413990`. A test that needs more than a rename while a unit moves is a signal:
stop and find out why before changing its expected values.

**Running the app.** Claude runs it only with `--campaign sandbox-claude`, in tmux. That copy is
gitignored and not on disk now, so re-copy it from `campaigns/sandbox` and keep its `campaign.toml`
ID. `sandbox` is Robert's: only its `rulebook/` folder changes, through `add-rules`.

**Search.** Find Go references with the LSP tools, not grep. Text search stays for string literals,
TOML and docs.

## Out of Scope

- **Amount rolls through hooks** — `E.004` C-4, and ruled out by Decision 1
- **Dropped dice, the roll card, collectors on effect gates** — `F.032`, `R.025`, `F.037`
- **Psychic Academy's `[[effects]]` entry** — `R.025`
- **The rest of the rule audit** — `R.026` keeps it; only the lockdown moves here
- **Other stale parts of `hooks.md`** — `D.010`; Phase 9 changes only the roll content
- **Forbidding a bare d10** — a review catches it (`E.004`)
- **`diceLabel`, `parseDice`, `parseAmount`, `DiceRoll`, `rollMatches`, `pickDice`** — stay where they are (P-4)

---

## Work Breakdown

**Overview**

- [x] **Phase 1 — Open the package**
  - Commit 1 — `refactor: add dice.Amount and retire DiceRoll.Roll` (3 tasks)
  - Commit 2 — `refactor: move Roller and RandRoller into engine/dice` (2 tasks)
  - Commit 3 — `test: share roller fakes in dicetest` (2 tasks)
- [x] **Phase 2 — Stat cleanups**
  - Commit 4 — `refactor: one roll phase enum in domain` (3 tasks)
  - Commit 5 — `refactor: type roll context stats and drop statScore wrappers` (3 tasks)
- [x] **Phase 3 — Stat routes**
  - Commit 6 — `refactor: move RollWithHooks and ResolveTie into engine/dice` (2 tasks)
  - Commit 7 — `fix: reroll keeps each die's sides` (2 tasks)
  - Commit 8 — `refactor: add Stat, Contest and Gate routes` (2 tasks)
- [x] **Phase 4 — Attack action**
  - Commit 9 — `refactor: Attack rolls through the dice routes` (3 tasks)
- [x] **Phase 5 — Expand Influence**
  - Commit 10 — `refactor: Expand Influence rolls through the dice routes` (3 tasks)
- [ ] **Phase 6 — Both gates**
  - Commit 11 — `refactor: ability gate uses dice.Gate` (1 task)
  - Commit 12 — `refactor: effect gate uses dice.Gate and rollWithHooks goes private` (4 tasks)
- [ ] **Phase 7 — Ability step split**
  - Commit 13 — `refactor: split AbilityStep.Roll into amount and table` (5 tasks)
  - Commit 14 — `refactor: Harvesters and Teach become target rolls` (6 tasks)
- [ ] **Phase 8 — Lockdown**
  - Commit 15 — `refactor: lockdown fee and cause move to data` (6 tasks)
- [ ] **Phase 9 — Docs**
  - Commit 16 — `docs: add dice architecture page and update roll references` (3 tasks)

### Phase 1 — Open the package

**In plain words:** the new package appears with its simplest route, and the die-rolling interface moves into it. Nothing about any roll changes.

The order is forced: `Roller` cannot leave `domain` while `DiceRoll.Roll` takes one.

#### Commit 1 — `refactor: add dice.Amount and retire DiceRoll.Roll`

##### Task 1 — `internal/faction/engine/dice/rolls.go`

New file, full contents:

```go
package dice

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

// Amount rolls dice whose total is the number. No hook reaches it.
func Amount(roll domain.DiceRoll, roller domain.Roller) int {
	total := roll.Modifier
	for range roll.NumDice {
		total += roller.Roll(roll.Sides)
	}
	return total
}
```

Add `rolls_test.go`, package `dice`, with a local roller fake (Commit 3 replaces it):

- A flat amount (`DiceRoll{Modifier: 1}`) returns 1 and rolls no dice.
- `2d6+1` with rolls 3 and 4 returns 8.

##### Task 2 — the 11 sites

**Pattern:** `x.Roll(roller)` becomes `dice.Amount(x, roller)`, with `*x` where `x` is a pointer.

| Site | Becomes |
|---|---|
| `attack.go:176` | `dice.Amount(*attackerDef.Attack.Damage, attack.roller)` |
| `attack.go:222` | `dice.Amount(*defenderDef.Counter, attack.roller)` |
| `attack.go:276` | `dice.Amount(*damage, attack.roller)` |
| `expand_influence.go:246` | `dice.Amount(*attackerDef.Attack.Damage, ba.roller)` |
| `ability/stage.go:161` | `dice.Amount(step.Roll.Die, roller)` (Commit 13 replaces it) |
| `capture_target.go:76` | `dice.Amount(reactor.coin, reactor.roller)` |
| `coin_gain.go:35` | `dice.Amount(ctx.Entry.CoinGain.Amount, ctx.Roller)` |
| `coin_gain.go:102` | `dice.Amount(reactor.amount, reactor.roller)` |
| `coin_steal.go:74` | `dice.Amount(entry.CoinSteal.Amount, reactor.roller)` |
| `self_damage.go:65` | `dice.Amount(reactor.amount, reactor.roller)` |
| `xp_roll.go:53` | `dice.Amount(reactor.params.Die, reactor.roller)` (Commit 14 replaces it) |

##### Task 3 — `internal/faction/domain/roll.go`

Delete the `Roll` method and its comment. `Roller` stays until Commit 2.

##### Commit message

```
refactor: add dice.Amount and retire DiceRoll.Roll

- new engine/dice package with the amount route
- all 11 DiceRoll.Roll sites call it; the method leaves domain
```

#### Commit 2 — `refactor: move Roller and RandRoller into engine/dice`

##### Task 1 — `internal/faction/engine/dice/roller.go`

New file, full contents. Delete `domain/roll.go` and `engine/roller.go`. In `rolls.go`, `Amount` takes `Roller`.

```go
package dice

import (
	"log/slog"
	"math/rand/v2"
)

// Roller produces pseudo-random die rolls. Production uses RandRoller; tests inject
// a deterministic fake. Roll returns a value in the inclusive range [1, sides].
type Roller interface {
	Roll(sides int) int
}

// RandRoller is the production Roller.
type RandRoller struct {
	log *slog.Logger
}

func NewRandRoller(log *slog.Logger) *RandRoller { return &RandRoller{log: log} }

func (roller *RandRoller) Roll(sides int) int {
	result := rand.IntN(sides) + 1
	roller.log.Debug("roll", "die", sides, "result", result)
	return result
}
```

##### Task 2 — the references

**Pattern:** `domain.Roller` becomes `dice.Roller`. Use the LSP reference list on `domain.Roller`; it holds 42 references outside `domain/roll.go`, across `action`, `effect`, `engine`, `turn` and `dispatch/roll.go`. `hooks` has none. In `core.go:52`, `NewRandRoller(log)` becomes `dice.NewRandRoller(log)`. `turn` imports `dice` for the type only: turn order stays a plain die.

##### Commit message

```
refactor: move Roller and RandRoller into engine/dice

- Roller leaves domain; only engine code uses it
- RandRoller leaves engine; core.go builds it from dice
```

#### Commit 3 — `test: share roller fakes in dicetest`

##### Task 1 — `internal/faction/engine/dice/dicetest/dicetest.go`

New file, full contents:

```go
package dicetest

// Sequence returns Values in order and panics when they run out, so a test that
// rolls more than it planned fails loudly.
type Sequence struct {
	Values []int
	next   int
}

func (roller *Sequence) Roll(int) int {
	if roller.next >= len(roller.Values) {
		panic("dicetest: Sequence exhausted")
	}
	value := roller.Values[roller.next]
	roller.next++
	return value
}

// Rolled returns how many dice have been rolled.
func (roller *Sequence) Rolled() int { return roller.next }

// Cycle returns Values in order and starts again when they run out.
type Cycle struct {
	Values []int
	next   int
}

func (roller *Cycle) Roll(int) int {
	value := roller.Values[roller.next%len(roller.Values)]
	roller.next++
	return value
}

// Rolled returns how many dice have been rolled.
func (roller *Cycle) Rolled() int { return roller.next }

// Constant returns Value on every roll.
type Constant struct {
	Value int
}

func (roller Constant) Roll(int) int { return roller.Value }
```

##### Task 2 — the nine fakes

| Fake | Becomes |
|---|---|
| `testharness.FixedRoller` (delete `testharness/roller.go`; update the comment at `register.go:18`) | `dicetest.Sequence` |
| `seqRoller` in `effects/clear_stealth_test.go` | `dicetest.Sequence` |
| `pairRoller` in `effects/contest_test.go` | `dicetest.Sequence` |
| `seqRoller` in `ability/dispatch_test.go` | `dicetest.Cycle` |
| `fixedRoller` in `actions/attack_test.go` | `dicetest.Cycle` |
| `fixedRoller` in `dispatch/roll_test.go` | `dicetest.Cycle` |
| `fixedRoller` in `arrival_test.go` | `dicetest.Constant` |
| `fixedRoller` in `effects/coin_gain_test.go` | `dicetest.Constant` |
| `constRoller` in `effect/effect_test.go` | `dicetest.Constant{Value: 1}` |

The real random fake in `turn/turn_test.go` stays. The local fake in `rolls_test.go` goes too. Each replacement keeps the old fake's behaviour, so no expected result moves.

##### Commit message

```
test: share roller fakes in dicetest

- Sequence, Cycle and Constant replace nine fakes
- the real random roller in turn_test stays
```

### Phase 2 — Stat cleanups

**In plain words:** the two roll phase types become one, and the stat a roll names becomes a typed value, not a string. The routes in Phase 3 are then written clean the first time.

#### Commit 4 — `refactor: one roll phase enum in domain`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
// EffectRollPhase names a hooks.RollPhase; domain cannot import hooks.
type EffectRollPhase string
```
**Replace with:**
```go
// RollPhase names which roll of a contest or test is being made. Rulebook filters
// name these values.
type RollPhase string
```

Then rename every `EffectRollPhase` to `RollPhase` with the LSP (19 references, including `rulebook/effects.go:148` and `EffectFilters.Phases`). The four constants keep their names and text.

##### Task 2 — `hooks/types.go`, `effects/roll_match.go`

In `hooks/types.go`, delete `RollPhase` and its constants, and type the field `Phase domain.RollPhase`.

In `roll_match.go`, delete `rollPhaseNames`, and:

**Find:**
```go
	if len(filters.Phases) > 0 && !slices.Contains(filters.Phases, rollPhaseNames[ctx.Phase]) {
```
**Replace with:**
```go
	if len(filters.Phases) > 0 && !slices.Contains(filters.Phases, ctx.Phase) {
```

##### Task 3 — the sites

| Old | New |
|---|---|
| `hooks.PhaseAttack` | `domain.RollPhaseAttack` |
| `hooks.PhaseDefense` | `domain.RollPhaseDefense` |
| `hooks.PhaseFactionTest` | `domain.RollPhaseFactionTest` |
| `hooks.PhaseContested` | `domain.RollPhaseContested` |
| `hooks.RollPhase` | `domain.RollPhase` |

**Watch for:** `hooks.PhaseAttack` was the zero value, so an unset `Phase` meant "attack". An unset phase now matches no phase filter. Every production site sets it. A test that starts failing here relied on the old zero value: set its phase to `domain.RollPhaseAttack` and keep its expected result.

##### Commit message

```
refactor: one roll phase enum in domain

- domain.RollPhase replaces EffectRollPhase and hooks.RollPhase
- rollPhaseNames goes; text values and TOML stay the same
```

#### Commit 5 — `refactor: type roll context stats and drop statScore wrappers`

##### Task 1 — `hooks/types.go`

**Find:**
```go
	Attribute         string
	OpponentAttribute string
```
**Replace with:**
```go
	Attribute         domain.FactionStat
	OpponentAttribute domain.FactionStat
```

##### Task 2 — the conversions

Drop `string(...)` at each `Attribute` and `OpponentAttribute` site: `attack.go:128-129`, `:145-146`, `:258`, `:263`; `expand_influence.go:100`, `:121-122`, `:216-217`, `:230-231`; `ability/stage.go:122-123`, `:137-138`; `effects/contest.go:70-71`. In `roll_match.go:72` and `:75`, compare `ctx.Attribute != filters.Attribute` and `ctx.OpponentAttribute != filters.OpponentAttribute`. Test literals such as `Attribute: "Cunning"` still compile; `string(domain.StatForce)` at `remove_tag_dice_test.go` and the stat conversions at `expand_influence_test.go:438-441` lose their conversions, and the `cunningDefense` helper in `bonus_die_test.go` takes a `domain.FactionStat`.

##### Task 3 — the wrappers

Delete `statScore` at `repair_asset.go:117` and `ability/stage.go:191`. Each caller calls the method: `repair_asset.go:72` becomes `faction.StatScore(def.Category)`, and the same for `attack.go:135`, `:152`, `:259`, `:264`, `expand_influence.go:222`, `:236` and `ability/stage.go:129`, `:144`. Delete the same switch, `statRating`, from `tui/adapter/action_collector.go`; its caller at `:176` calls the method too.

##### Commit message

```
refactor: type roll context stats and drop statScore wrappers

- RollContext.Attribute and OpponentAttribute are FactionStat
- both statScore wrappers and the adapter's statRating go; callers use Faction.StatScore
```

### Phase 3 — Stat routes

**In plain words:** the hooked roll and the tie rules move into the package, the reroll bug is fixed, and the new routes for a stat roll, a contest and a gate are added and tested. No site uses the new routes yet.

#### Commit 6 — `refactor: move RollWithHooks and ResolveTie into engine/dice`

##### Task 1 — the move

Move `dispatch/roll.go` to `dice/roll.go` and `dispatch/roll_test.go` to `dice/roll_test.go`, whole, with `package dice`. Their helpers (`rollScopes`, `assetScopes`, `scopeFactionID`, `hookOwner`, `offersBudgetFilter`, `freeOffers`, `diceSum`) and test helpers are used only in those files. The name stays exported until Commit 12. Fence-sign `roll.go:88` moves with the code.

Move `ResolveTie` from `dispatch/rules.go` to a new `dice/tie.go`, with `TestResolveTie_*` (`rules_test.go:212-245`) and `stubTieResolver` into `dice/tie_test.go`.

##### Task 2 — the callers

**Pattern:** `dispatch.RollWithHooks` becomes `dice.RollWithHooks`, and `dispatch.ResolveTie` becomes `dice.ResolveTie`. There are 11 production calls, seven test calls in `effects`, five production `ResolveTie` calls and two in `lose_ties_test.go`. Drop the `dispatch` import where nothing else uses it (`effects/contest.go`, `expand_influence.go`); `attack.go` and `ability/stage.go` keep it for other resolvers.

##### Commit message

```
refactor: move RollWithHooks and ResolveTie into engine/dice

- dice/roll.go and its tests move whole
- ResolveTie and its tests move to dice/tie.go
- dispatch keeps Cat 3, Cat 4 and Cat 6
```

#### Commit 7 — `fix: reroll keeps each die's sides`

Closes `B.010`. Inert today: every `sides` key in the SWN data is 10.

##### Task 1 — `dice/roll.go`

**Find:**
```go
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
```
**Replace with:**
```go
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
	result.Sum = diceSum(result.Dice) + result.Modifier

	// Apply keep-highest trim if any offer requested it.
	if n := rollState.KeepHighest(); n > 0 && len(result.Dice) > n {
		result.Dice, dieSides = keepHighest(result.Dice, dieSides, n)
		result.Sum = diceSum(result.Dice) + result.Modifier
	}
```

**Find:**
```go
					result.Dice[idx] = roller.Roll(baseRoll.Sides)
```
**Replace with:**
```go
					result.Dice[idx] = roller.Roll(dieSides[idx])
```

Add, beside `diceSum`:

```go
// keepHighest returns the n highest dice, highest first, each with its own sides.
func keepHighest(values, dieSides []int, n int) ([]int, []int) {
	order := make([]int, len(values))
	for index := range order {
		order[index] = index
	}
	sort.SliceStable(order, func(i, j int) bool { return values[order[i]] > values[order[j]] })
	keptValues := make([]int, n)
	keptSides := make([]int, n)
	for position, index := range order[:n] {
		keptValues[position] = values[index]
		keptSides[position] = dieSides[index]
	}
	return keptValues, keptSides
}
```

##### Task 2 — `dice/roll_test.go`

Add a roller fake that records the sides of each roll. New tests:

- A reroll of an extra d6 rolls with 6 sides, on a 1d10 base.
- After keep-highest keeps an extra d6 over the d10, a reroll of index 0 rolls with 6 sides.
- The existing keep-highest tests pass unchanged: the kept values and their order are the same.

##### Commit message

```
fix: reroll keeps each die's sides

- each die in a roll keeps its own sides through keep-highest
- closes B.010; inert today because every die in the data is a d10
```

#### Commit 8 — `refactor: add Stat, Contest and Gate routes`

##### Task 1 — `internal/faction/engine/dice/stat.go`

New file, full contents:

```go
package dice

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

const statDieSides = 10

// Engine is what a hooked roll needs from the game. A nil Collector asks no one
// (hooks.md, Key Decisions).
type Engine struct {
	Registry     *hooks.Registry
	Collector    hooks.Collector
	Roller       Roller
	FactionState *state.FactionState
	Rulebook     *rulebook.Rulebook
}

// Side is one party to a roll. Asset is nil when no asset stands there (a faction
// test, a Base defending). Detects marks a side testing to detect stealth.
type Side struct {
	Faction *domain.Faction
	Asset   *domain.Asset
	Stat    domain.FactionStat
	Detects bool
}

// Stat is one stat roll: 1d10 plus the side's stat, through the hooks. A zero
// opponent is a roll with no one against it.
func (engine Engine) Stat(phase domain.RollPhase, side, opponent Side, world, action string) hooks.RollResult {
	ctx := hooks.RollContext{
		Phase:             phase,
		Actor:             side.Faction,
		Opponent:          opponent.Faction,
		Attribute:         side.Stat,
		OpponentAttribute: opponent.Stat,
		Asset:             side.Asset,
		OpponentAsset:     opponent.Asset,
		World:             world,
		Action:            action,
		Detection:         side.Detects,
	}
	base := domain.DiceRoll{NumDice: 1, Sides: statDieSides, Modifier: side.Faction.StatScore(side.Stat)}
	return RollWithHooks(ctx, base, engine.Registry, engine.Collector, engine.Roller, side.Faction, engine.FactionState, engine.Rulebook)
}

// Outcome is a rolled contest. Both sides' rolls stay on it for the roll card.
type Outcome struct {
	Attack  hooks.RollResult
	Defense hooks.RollResult
	tie     hooks.TieOutcome
}

// AttackLands is the Attack action's reading: a tie lands, unless a resolver gives
// it to the defender. The counterstrike and the free attack read it too.
func (outcome Outcome) AttackLands() bool {
	if outcome.Attack.Sum == outcome.Defense.Sum {
		return outcome.tie != hooks.TieDefenderWins
	}
	return outcome.Attack.Sum > outcome.Defense.Sum
}

// CounterLands is the defender's side of the same reading: a tie lands, unless a
// resolver gives it to the attacker.
func (outcome Outcome) CounterLands() bool {
	if outcome.Attack.Sum == outcome.Defense.Sum {
		return outcome.tie != hooks.TieAttackerWins
	}
	return outcome.Defense.Sum > outcome.Attack.Sum
}

// GatePasses is a gate's reading: a tie goes to the defender, unless a resolver
// gives an attack's tie to the attacker.
func (outcome Outcome) GatePasses() bool {
	if outcome.Attack.Sum == outcome.Defense.Sum {
		return outcome.tie == hooks.TieAttackerWins
	}
	return outcome.Attack.Sum > outcome.Defense.Sum
}

// Contest rolls two stat rolls, attacker first, and asks the tie resolvers only on
// the tie of an attack. A test's tie has no resolver.
func (engine Engine) Contest(kind domain.ContestRoll, attacker, defender Side, world, action string) Outcome {
	attackPhase, defensePhase := domain.RollPhaseFactionTest, domain.RollPhaseContested
	if kind == domain.ContestRollAttack {
		attackPhase, defensePhase = domain.RollPhaseAttack, domain.RollPhaseDefense
	}
	attackerOpponent, defenderOpponent := defender, attacker
	// deliberate: Opponent nil, not the same faction, because Actor == Opponent would ask every asset hook on the world twice (overview Decision 9).
	if attacker.Faction.ID == defender.Faction.ID {
		attackerOpponent, defenderOpponent = Side{Stat: defender.Stat}, Side{Stat: attacker.Stat}
	}
	outcome := Outcome{
		Attack:  engine.Stat(attackPhase, attacker, attackerOpponent, world, action),
		Defense: engine.Stat(defensePhase, defender, defenderOpponent, world, action),
	}
	if outcome.Attack.Sum == outcome.Defense.Sum && kind == domain.ContestRollAttack {
		tieContext := hooks.RollContext{Phase: attackPhase, Actor: attacker.Faction, Opponent: attackerOpponent.Faction, Asset: attacker.Asset, World: world, Action: action}
		outcome.tie = ResolveTie(engine.Registry, tieContext, engine.FactionState)
	}
	return outcome
}

// Gate rolls a contest that decides if an ability or effect goes ahead. With roller
// "self" the declaring side attacks and must win. With "target" the target tests to
// keep what it has, and the gate passes when that test fails. Each side's stat comes
// from the contest, by the role it rolls in.
func (engine Engine) Gate(contest domain.Contest, roller domain.ContestRoller, declaring, target Side, world, action string) bool {
	attacker, defender := declaring, target
	if roller == domain.ContestRollerTarget {
		attacker, defender = target, declaring
	}
	attacker.Stat, defender.Stat = contest.AttackerStat, contest.DefenderStat
	passes := engine.Contest(contest.Roll, attacker, defender, world, action).GatePasses()
	if roller == domain.ContestRollerTarget {
		return !passes
	}
	return passes
}
```

##### Task 2 — `dice/stat_test.go`

Package `dice`, using `dicetest` and a recording Cat 1 hook. Cases:

- `Stat` totals one d10 plus the side's stat, and fills every `RollContext` field from the two sides.
- A zero opponent leaves `Opponent` and `OpponentAttribute` empty.
- `Contest` rolls the attacker first, with attack phases for `attack` and test phases for `test`.
- Same faction on both sides: no `Opponent` and no `OpponentAsset` on either roll; each roll still names the other's stat.
- A nil defender `Asset` reaches the context as nil on both rolls.
- A table of `AttackLands`, `CounterLands` and `GatePasses` over win, lose and tie, with each tie outcome.
- A resolver is asked once on an attack's tie, not on a test's tie, and not when the sums differ.
- `Gate` with roller `self`: a win passes and a tie fails.
- `Gate` with roller `target`: the target rolls first, a failed test passes, a tie passes, and `AttackerStat` goes to the target.
- `Detects` marks only the declaring side's roll, under both rollers.

##### Commit message

```
refactor: add Stat, Contest and Gate routes

- Engine and Side carry what a roll needs
- Outcome names the three tie readings
- Gate does the target-roller flip
```

### Phase 4 — Attack action

**In plain words:** the Attack action's matchup and counterstrike call the contest route. The tie rule moves from three lines at each site to one name, and tie resolvers run on ties only.

#### Commit 9 — `refactor: Attack rolls through the dice routes`

Verified in planning: `ResolveTie` has no side effects. The only production resolver (`lose_ties.go:31`) reads state only, and every test stub returns a fixed value, so asking on a tie only gives the same results.

##### Task 1 — `attack.go`, the holder

Add:

```go
func (attack *AttackAction) rolls(factionState *state.FactionState, rulebook *rulebook.Rulebook) dice.Engine {
	return dice.Engine{Registry: attack.registry, Collector: attack.collector, Roller: attack.roller, FactionState: factionState, Rulebook: rulebook}
}
```

##### Task 2 — `attack.go`, the matchup

Replace `attack.go:124-167` (from `attackResult := dice.RollWithHooks(` to the `defenderWins :=` line) with:

```go
		outcome := attack.rolls(factionState, rulebook).Contest(
			domain.ContestRollAttack,
			dice.Side{Faction: faction, Asset: attacker, Stat: attackerDef.Attack.AttackerStat},
			dice.Side{Faction: defenderFaction, Asset: defender, Stat: attackerDef.Attack.DefenderStat},
			attacker.Location.WorldID, attack.Name(),
		)
		attackerWins := outcome.AttackLands()
		defenderWins := outcome.CounterLands()
```

##### Task 3 — `attack.go`, the counterstrike

Replace `attack.go:256-273` (from `worldID :=` to the closing brace of the `if`) with the lines below. The fence-sign from `:257` now lives in `Engine.Contest`.

```go
	outcome := attack.rolls(factionState, rulebook).Contest(
		domain.ContestRollAttack,
		dice.Side{Faction: faction, Asset: attacker, Stat: attackerDef.Attack.AttackerStat},
		dice.Side{Faction: faction, Asset: attacker, Stat: attackerDef.Attack.DefenderStat},
		attacker.Location.WorldID, attack.Name(),
	)
	damage := attackerDef.Counter
	if outcome.AttackLands() {
		damage = attackerDef.Attack.Damage
	}
```

**Guards:** `TestAttack_RollsNameTheOpposingStat` and `TestAttack_CounterstrikeRollsNameTheOpposingStat` hold the contexts; `TestAttack_TieResolver_*` hold the ties. All pass unchanged.

##### Commit message

```
refactor: Attack rolls through the dice routes

- the matchup and the counterstrike call Contest
- tie resolvers run on a tie only; the results are the same
```

### Phase 5 — Expand Influence

**In plain words:** Expand Influence's own roll and each rival's roll call the stat route, and a rival's free attack calls the contest route. The plain comparison against rivals stays where it is.

#### Commit 10 — `refactor: Expand Influence rolls through the dice routes`

##### Task 1 — the expander's roll

Replace `expand_influence.go:96-106` with the lines below. The fence-sign at `:95` stays above it. `ei` gets a `rolls` helper in the shape of Commit 9 Task 1, and `baseAttack` gets its own.

```go
	factionRoll := ei.rolls(factionState, rulebook).Stat(
		domain.RollPhaseFactionTest,
		dice.Side{Faction: faction, Stat: domain.StatCunning},
		dice.Side{},
		ei.order.World, ei.Name(),
	).Sum
```

##### Task 2 — each rival's roll

Replace `expand_influence.go:116-128` with:

```go
		rivalRoll := ei.rolls(factionState, rulebook).Stat(
			domain.RollPhaseContested,
			dice.Side{Faction: rival, Stat: domain.StatCunning},
			dice.Side{Faction: faction, Stat: domain.StatCunning},
			ei.order.World, ei.Name(),
		).Sum
```

##### Task 3 — the free attack

Replace `expand_influence.go:210-245` (from `worldID :=` through `if hits {`) with the lines below. The defender has no `Asset`: a Base defends. The damage block inside is unchanged.

```go
		outcome := ba.rolls(factionState, rulebook).Contest(
			domain.ContestRollAttack,
			dice.Side{Faction: rival, Asset: attacker, Stat: attackerDef.Attack.AttackerStat},
			dice.Side{Faction: ownerFaction, Stat: domain.StatCunning},
			target.Location.WorldID, ba.action,
		)
		if outcome.AttackLands() {
```

**Guard:** `TestExpandInfluence_RollsCarryContext` passes unchanged.

##### Commit message

```
refactor: Expand Influence rolls through the dice routes

- own and rival rolls call Stat; the comparison stays at the site
- the free attack calls Contest
```

### Phase 6 — Both gates

**In plain words:** the ability gate and the effect gate become one call, so one rule has one home. After this, every stat roll goes through a named route, and the raw hooked roll goes private.

#### Commit 11 — `refactor: ability gate uses dice.Gate`

##### Task 1 — `ability/stage.go`

Keep `passesGate`, its signature and its guards, so its tests stay as they are. Replace everything from `worldID := asset.Location.WorldID` to the end of the function, keeping both fence-signs in place, and update the doc comment to point at `dice.Gate`:

```go
	// deliberate: a literal, not UseAssetAbility.Name(), because actions imports ability (import cycle).
	const actionName = "Use Asset Ability"
	engine := dice.Engine{Registry: registry, Collector: collector, Roller: roller, FactionState: factionState, Rulebook: rulebook}
	passed := engine.Gate(
		*step.Contest,
		domain.ContestRollerSelf,
		dice.Side{Faction: faction, Asset: asset, Detects: step.Effect == domain.EffectRevealStealth},
		dice.Side{Faction: defender, Asset: target.Asset},
		asset.Location.WorldID, actionName,
	)
	return passed, nil
```

**Guards:** `detection_test.go` and `TestDispatch_ContestRollsNameTheOpposingStat` pass with no edits.

##### Commit message

```
refactor: ability gate uses dice.Gate

- passesGate keeps its guards and calls the shared gate
```

#### Commit 12 — `refactor: effect gate uses dice.Gate and rollWithHooks goes private`

##### Task 1 — `effects/contest.go`

Replace the file with:

```go
package effects

// deliberate: no action name, not the effect's, because Imperialists' action = "Attack" filter must not reach an effect contest (Effort 3 Decision 4).
const contestAction = ""
```

##### Task 2 — the three callers

**Pattern:** the `contestRoll` literal becomes an `Engine` plus a `Gate` call. The nil collector and its fence-sign stay at each caller. `clear_stealth.go:58`:

```go
			rolls := dice.Engine{
				// deliberate: a nil collector, not a prompt, because a phase route has none; the roll-card follow-on row gives it one (overview Decision 2).
				Collector:    nil,
				Registry:     ctx.Registry,
				Roller:       ctx.Roller,
				FactionState: ctx.FactionState,
				Rulebook:     ctx.Rulebook,
			}
			if contest := ctx.Entry.Gates.Contest; contest != nil && !rolls.Gate(
				contest.Contest, contest.Roller,
				dice.Side{Faction: ctx.Holder.Faction, Asset: ctx.Asset, Detects: true},
				dice.Side{Faction: target, Asset: asset},
				worldID, contestAction,
			) {
				continue
			}
```

| Site | Declaring | Target | Detects |
|---|---|---|---|
| `clear_stealth.go:58` | `ctx.Holder.Faction`, `ctx.Asset` | `target`, `asset` | Yes |
| `clear_stealth.go:134` | `owner`, `self` | `target`, `target.Assets[assetID]` | Yes |
| `withdraw_permission.go:72` | `contester.owner`, `self` | `query.Requester`, no asset | No |

##### Task 3 — `dice/roll.go`, `dice/stat.go`

Rename `RollWithHooks` to `rollWithHooks`. The compiler then lists every outside caller left.

##### Task 4 — the tests

- `effects/contest_test.go`: add a local helper that builds the `Engine` and calls `Gate`. The six tests keep their cases and expected values.
- The seven outside calls (`bonus_die_test.go:153`, `:162`, `:191`, `:192`, `:216`; `remove_tag_dice_test.go:135`; `reroll_test.go:78`) call `Engine.Stat`.
- **Expected change:** `bonus_die_test.go:191-192`, `remove_tag_dice_test.go:135` and `rollThroughDispatcher` pass a base roll with no modifier, but `Stat` always adds the stat. Where the faction's stat is not 0, add it to the expected sum. The hook behaviour under test does not change.

##### Commit message

```
refactor: effect gate uses dice.Gate and rollWithHooks goes private

- the three effect-gate callers call Gate; contestRoll goes
- the empty Action is now one named constant
- RollWithHooks is unexported: a hooked roll goes through a route
```

### Phase 7 — Ability step split

**In plain words:** an ability's step now says which type of roll it makes, and its handler reads a typed result, not a bare number. Harvesters and Teach Planetary Population become target rolls, as the book words them, and their results do not change.

#### Commit 13 — `refactor: split AbilityStep.Roll into amount and table`

Commit 14 adds the target type. Teach stays a table until then.

##### Task 1 — `domain/asset.go`

**Find:**
```go
	Contest        *Contest
	Roll           *RollTable
	Relocate       *RelocateParams
```
**Replace with:**
```go
	Contest        *Contest
	Amount         *DiceRoll  // at most one of Amount and Table
	Table          *RollTable // at most one of Amount and Table
	Relocate       *RelocateParams
```

##### Task 2 — `rulebook/rulebook.go`

In `convertAbilityStep`:

**Find:**
```go
		roll := &domain.RollTable{Die: *die}
		roll.Outcomes = make([]domain.AbilityOutcome, 0, len(record.Outcomes))
		for _, o := range record.Outcomes {
			roll.Outcomes = append(roll.Outcomes, domain.AbilityOutcome{
				From:        o.From,
				To:          o.To,
				Coin:        o.Coin,
				DestroySelf: o.DestroySelf,
			})
		}
		step.Roll = roll
	}
```
**Replace with:**
```go
		if len(record.Outcomes) == 0 {
			step.Amount = die
		} else {
			table := &domain.RollTable{Die: *die}
			for _, o := range record.Outcomes {
				table.Outcomes = append(table.Outcomes, domain.AbilityOutcome{
					From:        o.From,
					To:          o.To,
					Coin:        o.Coin,
					DestroySelf: o.DestroySelf,
				})
			}
			step.Table = table
		}
	}
```

Replace the checks at `:417-430`, `:440` and `:443`:

```go
	if step.Contest != nil && (step.Amount != nil || step.Table != nil) {
		return domain.AbilityStep{}, fmt.Errorf("a step has both contest and a die; they are mutually exclusive")
	}
	if step.Effect == domain.EffectCoinDelta && step.Table == nil {
		return domain.AbilityStep{}, fmt.Errorf("coin_delta needs outcome rows")
	}
	if step.Effect == domain.EffectPurchaseDiscount && step.Amount == nil {
		return domain.AbilityStep{}, fmt.Errorf("purchase_discount needs a die and no outcome rows")
	}
	if step.Effect == domain.EffectRaiseTechLevel {
		if step.Target != domain.TargetWorld || step.Table == nil || step.RaiseTechLevel == nil {
			return domain.AbilityStep{}, fmt.Errorf("raise_tech_level requires target = %q, a die with rows and a [raise_tech_level] block", "world")
		}
		if slices.ContainsFunc(step.Table.Outcomes, func(outcome domain.AbilityOutcome) bool { return outcome.Coin != 0 || outcome.DestroySelf }) {
			return domain.AbilityStep{}, fmt.Errorf("raise_tech_level outcome rows mark a success; they carry no coin or destroy_self")
		}
	}
```

`block_attack` requires `step.Amount != nil`. `levy_coin` refuses `step.Amount != nil || step.Table != nil`. The tag check at `:733` reads `step.Table`.

##### Task 3 — `dice/rolls.go`

Add. `Table` takes over `outcomeFor` from `ability/stage.go`:

```go
// Table rolls dice that pick a row: the first whose range holds the total. No row
// means no result; the absence of a row is a zero payout.
func Table(table domain.RollTable, roller Roller) (domain.AbilityOutcome, bool) {
	total := Amount(table.Die, roller)
	for _, outcome := range table.Outcomes {
		if total >= outcome.From && total <= outcome.To {
			return outcome, true
		}
	}
	return domain.AbilityOutcome{}, false
}

// StepResult is what an ability step's die gave, by the step's type: Amount for an
// amount roll, Row for a table roll with a matching row.
type StepResult struct {
	Amount int
	Row    *domain.AbilityOutcome
}

// RollStep rolls the step's die. A step has at most one (parser-enforced), and one
// with none rolls nothing.
func RollStep(step domain.AbilityStep, roller Roller) StepResult {
	switch {
	case step.Amount != nil:
		return StepResult{Amount: Amount(*step.Amount, roller)}
	case step.Table != nil:
		if row, matched := Table(*step.Table, roller); matched {
			return StepResult{Row: &row}
		}
	}
	return StepResult{}
}
```

Delete `resolveRoll` and `outcomeFor` from `ability/stage.go`.

##### Task 4 — `ability/dispatch.go` and the handlers

In `StepContext`, replace `RollResult int` with:

```go
	Rolled       dice.StepResult // what the step's die gave; zero for a step with no die
```

**Find (`dispatch.go:144`):**
```go
			RollResult:   resolveRoll(step, roller),
```
**Replace with:**
```go
			Rolled:       dice.RollStep(step, roller),
```

This is the roll-order guarantee: `RollStep` runs in the same place, once for each step that passes its gate, before the handler. A handler's early return cannot skip a die.

| Handler | Change |
|---|---|
| `block_attack.go:15` | `cost := ctx.Rolled.Amount` |
| `purchase_discount.go` | The guard checks `ctx.Step.Amount == nil`; read `ctx.Rolled.Amount` |
| `coin_delta.go:14-20` | The guard checks `ctx.Step.Table == nil`; `outcome := ctx.Rolled.Row`, returning nil when it is nil |
| `raise_tech_level.go:14`, `:30` | The guard checks `ctx.Step.Table == nil`; the mutation depends on `ctx.Rolled.Row != nil` |

##### Task 5 — the tests

- **Step literals:** `Roll: &domain.RollTable{Die: d}` with no rows becomes `Amount: &d`; with rows, `Table: &domain.RollTable{...}`. Sites: `block_attack_test.go:24`, `coin_delta_test.go:13`, `dispatch_test.go:433`, `purchase_discount_test.go:16`, `raise_tech_level_test.go:19`, `use_asset_ability_test.go:262`, `:417`, `rulebook_test.go:231-237`, `:277-278`, `:1130`, `:1193`.
- **`RollResult` literals:** `coin_delta_test.go:45`, `:62` and `raise_tech_level_test.go:45`, `:63` set `Rolled: dice.StepResult{Row: &row}`, or leave it zero for no match. `purchase_discount_test.go:18` sets `Rolled: dice.StepResult{Amount: n}`.
- **Row selection moves:** add `dice` tests for `Table`: a total in range, on each edge and out of range. Add `RollStep` tests for each type and for a step with no die.
- **New load rules:** a test each that `coin_delta` with no rows is refused, `purchase_discount` with rows is refused, and `block_attack` with rows is refused. A fixture that breaks a new rule gets its data fixed, not the rule.

##### Commit message

```
refactor: split AbilityStep.Roll into amount and table

- AbilityStep has Amount and Table; die alone is an amount
- StepContext.Rolled replaces RollResult, filled at the same spot
- outcomeFor and resolveRoll move to dice as Table and RollStep
- the loader refuses a die shape its effect cannot read
```

#### Commit 14 — `refactor: Harvesters and Teach become target rolls`

##### Task 1 — `domain/asset.go`, `domain/effect.go`

Add beside `RollTable`, and add `TargetRoll *TargetRoll` to `AbilityStep` below `Table`. Delete `XPRollParams`; `EffectDefinition.XPRoll` becomes `*TargetRoll`.

```go
// TargetRoll is a roll that must meet or beat a number. Amount is what a pass pays;
// the effect decides in what, and it is 0 when a pass pays nothing.
type TargetRoll struct {
	Die    DiceRoll
	Target int
	Amount int
}
```

##### Task 2 — `dice/rolls.go`, `xp_roll.go`

Add below. In `RollStep`, add the `TargetRoll` case; `StepResult` gains `Passed bool`.

```go
// Target rolls dice that must meet or beat the target.
func Target(roll domain.TargetRoll, roller Roller) bool {
	return Amount(roll.Die, roller) >= roll.Target
}
```

```go
	case step.TargetRoll != nil:
		return StepResult{Passed: Target(*step.TargetRoll, roller)}
```

In `xp_roll.go`, the reactor's `params` is a `domain.TargetRoll`, and:

**Find (as Commit 1 left it):**
```go
		if dice.Amount(reactor.params.Die, reactor.roller) < reactor.params.Target {
```
**Replace with:**
```go
		if !dice.Target(reactor.params, reactor.roller) {
```

##### Task 3 — the loader

In `rulebook/effects.go`, rename `xpRollRecord` to `targetRollRecord`, and add:

```go
func convertTargetRoll(label string, record *targetRollRecord) (*domain.TargetRoll, error) {
	die, err := parseDice(record.Die)
	if err != nil {
		return nil, fmt.Errorf("%s die: %w", label, err)
	}
	if die == nil || die.NumDice == 0 {
		return nil, fmt.Errorf("%s die is required", label)
	}
	if record.Target <= 0 || record.Amount < 0 {
		return nil, fmt.Errorf("%s needs target > 0 and amount >= 0, got %d and %d", label, record.Target, record.Amount)
	}
	return &domain.TargetRoll{Die: *die, Target: record.Target, Amount: record.Amount}, nil
}
```

The `xp_roll` case calls it with `"xp_roll"` and still refuses an `amount` of 0.

In `rulebook/rulebook.go`, add `TargetRoll *targetRollRecord` with tag `toml:"target_roll"` to `abilityStepRecord`, read it in `convertAbilityStep`, and change the checks from Commit 13 Task 2:

```go
	if record.Die != "" && step.TargetRoll != nil {
		return domain.AbilityStep{}, fmt.Errorf("a step has both a die and a target_roll; use one")
	}
	if step.Contest != nil && (step.Amount != nil || step.Table != nil || step.TargetRoll != nil) {
		return domain.AbilityStep{}, fmt.Errorf("a step has both contest and a die; they are mutually exclusive")
	}
	if step.Effect == domain.EffectCoinDelta {
		if step.Table == nil && step.TargetRoll == nil {
			return domain.AbilityStep{}, fmt.Errorf("coin_delta needs outcome rows or a target_roll")
		}
		if step.TargetRoll != nil && step.TargetRoll.Amount <= 0 {
			return domain.AbilityStep{}, fmt.Errorf("coin_delta target_roll needs amount > 0, got %d", step.TargetRoll.Amount)
		}
	}
	if step.Effect == domain.EffectPurchaseDiscount && step.Amount == nil {
		return domain.AbilityStep{}, fmt.Errorf("purchase_discount needs a die and no outcome rows")
	}
	if step.Effect == domain.EffectRaiseTechLevel {
		if step.Target != domain.TargetWorld || step.TargetRoll == nil || step.RaiseTechLevel == nil {
			return domain.AbilityStep{}, fmt.Errorf("raise_tech_level requires target = %q, a target_roll and a [raise_tech_level] block", "world")
		}
		if step.TargetRoll.Amount != 0 {
			return domain.AbilityStep{}, fmt.Errorf("raise_tech_level target_roll pays nothing; it carries no amount")
		}
	}
```

`levy_coin` also refuses a target roll. The flat-block guard at `rulebook.go:322` adds `r.TargetRoll != nil`.

##### Task 4 — the handlers

`coin_delta.go`, full body:

```go
func coinDelta(ctx StepContext) ([]domain.Mutation, error) {
	if ctx.Step.TargetRoll != nil {
		if !ctx.Rolled.Passed {
			return nil, nil
		}
		return []domain.Mutation{coinPayout(ctx, ctx.Step.TargetRoll.Amount)}, nil
	}
	if ctx.Step.Table == nil {
		return nil, fmt.Errorf("coin_delta: step has no roll table or target roll")
	}
	outcome := ctx.Rolled.Row
	if outcome == nil {
		return nil, nil
	}
	if outcome.DestroySelf || (outcome.Coin < 0 && ctx.Faction.Coin < -outcome.Coin) {
		return []domain.Mutation{domain.AssetRemoved{
			FactionID:         ctx.Faction.ID,
			AssetID:           ctx.Asset.ID,
			Cause:             "ability",
			CausedByFactionID: ctx.Faction.ID,
		}}, nil
	}
	if outcome.Coin != 0 {
		return []domain.Mutation{coinPayout(ctx, outcome.Coin)}, nil
	}
	return nil, nil
}

func coinPayout(ctx StepContext, amount int) domain.Mutation {
	return domain.CoinDelta{
		FactionID:         ctx.Faction.ID,
		Delta:             amount,
		CausedByFactionID: ctx.Faction.ID,
		Cause:             "ability",
	}
}
```

`raise_tech_level.go`: the guard checks `ctx.Step.TargetRoll == nil`, and the mutation depends on `ctx.Rolled.Passed`.

##### Task 5 — the data and the re-seed

`rulebooks/swn/assets/wealth_assets.toml`:

**Find:**
```toml
[assets.Harvesters.ability]
effect = "coin_delta"
die = "1d6"
[[assets.Harvesters.ability.outcomes]]
from = 3
to = 6
coin = 1
```
**Replace with:**
```toml
[assets.Harvesters.ability]
effect = "coin_delta"
[assets.Harvesters.ability.target_roll]
die = "1d6"
target = 3
amount = 1
```

`rulebooks/swn/tags.toml`:

**Find:**
```toml
target = "world"
die    = "1d12"
[[tags.PreceptorArchive.ability.outcomes]]
from = 12
to   = 12
[tags.PreceptorArchive.ability.raise_tech_level]
```
**Replace with:**
```toml
target = "world"
[tags.PreceptorArchive.ability.target_roll]
die    = "1d12"
target = 12
[tags.PreceptorArchive.ability.raise_tech_level]
```

Results are the same: 3 or more on a d6 and 12 on a d12 are the rows' ranges, because those rows ran to the die's top face.

Re-seed, in this commit:

```
go run ./cmd/gm-toolkit campaign add-rules --rules rulebooks/swn --campaign sandbox --replace
```

Check `git status` shows only `campaigns/sandbox/rulebook/` changes.

##### Task 6 — the tests

- `rulebook_test.go:841` and `xp_roll_test.go:24`: `domain.TargetRoll` replaces `XPRollParams`.
- `TestLoad_PreceptorArchiveTeaches` wants `TargetRoll: &domain.TargetRoll{Die: 1d12, Target: 12}`.
- New `TestLoad_HarvestersRollsATarget`: `1d6`, target 3, amount 1, effect `coin_delta`.
- `raise_tech_level_test.go`: a pass raises the world, and a miss still pays the cost.
- `coin_delta_test.go`: a target roll pays its amount on a pass and nothing on a miss.
- `dice` tests: `Target` passes on the number, fails one below, and `RollStep` sets `Passed`.
- **New load rules:** both a die and a target roll; a contest and a target roll; `coin_delta` with amount 0; `raise_tech_level` with an amount or with rows; `target` of 0.
- `TestPreceptorArchive_TeachPlanetaryPopulation` passes unchanged.

##### Commit message

```
refactor: Harvesters and Teach become target rolls

- abilities gain a target_roll block, shared with xp_roll
- dice.Target states meet or beat once
- results do not change; the sandbox rulebook is re-seeded
```

### Phase 8 — Lockdown

**In plain words:** the Transport Lockdown's 1d4 fee moves from Go into its rulebook entry, and the fee's cause takes the effect's name. Players pay the same fee at the same moment.

#### Commit 15 — `refactor: lockdown fee and cause move to data`

##### Task 1 — `domain/effect.go`

Add, and add `LockOut *LockOutParams` to `EffectDefinition`:

```go
// LockOutParams is the Coin an arrival pays a Transport Lockdown.
type LockOutParams struct {
	Fee DiceRoll
}
```

##### Task 2 — `rulebook/effects.go`

- Add `LockOut *amountRecord` with tag `toml:"lock_out"` to `effectRecord`.
- In `presentParamBlocks`, add `if record.LockOut != nil { present = append(present, domain.SpecialLockOut) }`.
- In `convertEffectParams`:

```go
	case domain.SpecialLockOut:
		fee, err := convertAmount(definition.Effect, record.LockOut)
		if err != nil {
			return err
		}
		definition.LockOut = &domain.LockOutParams{Fee: fee}
```

##### Task 3 — the data and the re-seed

`rulebooks/swn/assets/cunning_assets.toml`:

**Find:**
```toml
effect  = "lock_out"
[assets.TransportLockdown.effects.filters]
owner = "own"
scope = "self"
```
**Replace with:**
```toml
effect  = "lock_out"
[assets.TransportLockdown.effects.filters]
owner = "own"
scope = "self"
[assets.TransportLockdown.effects.lock_out]
amount = "1d4"
```

Re-seed with the command from Commit 14 Task 5.

##### Task 4 — `domain/asset.go`, `relocate/relocate.go`

Replace `LockedOutBy` with the version below. `SortedAssets` makes two locks in one faction name the same asset every time. The comment above it keeps its text, with "and the asset that holds the lock".

```go
func LockedOutBy(factions map[string]*Faction, worldID, factionID string) (string, *Asset) {
	if worldID == "" {
		return "", nil
	}
	for _, holderID := range slices.Sorted(maps.Keys(factions)) {
		if holderID == factionID {
			continue
		}
		for _, asset := range SortedAssets(factions[holderID]) {
			if asset.Location.WorldID != worldID || !asset.Active() {
				continue
			}
			if slices.ContainsFunc(asset.Conditions, func(condition Condition) bool {
				return condition.Kind == ConditionLockedOut && condition.SourceAssetID == asset.ID &&
					condition.TargetFactionID == factionID && condition.WorldID == worldID
			}) {
				return holderID, asset
			}
		}
	}
	return "", nil
}
```

In `relocate.go:182`:

```go
		lockerID, _ := domain.LockedOutBy(request.FactionState.Factions, worldID, request.Faction.ID)
		return lockerID != ""
```

##### Task 5 — `engine/arrival.go`

**Find:**
```go
		if lockerID := domain.LockedOutBy(factionState.Factions, order.Destination.WorldID, faction.ID); lockerID != "" {
			hold, cost := lockdownHold(completed, order, lockerID, e.Rand.Roll(4), faction.Coin-spent[faction.ID])
			return hold, map[string]int{faction.ID: cost}, nil
		}
```
**Replace with:**
```go
		if lockerID, lockdown := domain.LockedOutBy(factionState.Factions, order.Destination.WorldID, faction.ID); lockerID != "" {
			fee, err := e.lockdownFee(lockdown)
			if err != nil {
				return nil, nil, err
			}
			hold, cost := lockdownHold(completed, order, lockerID, fee, faction.Coin-spent[faction.ID])
			return hold, map[string]int{faction.ID: cost}, nil
		}
```

Add:

```go
// lockdownFee rolls the fee on the lockdown's own lock_out entry, before the hold
// is decided: a faction that cannot pay has still rolled.
func (e *Engine) lockdownFee(lockdown *domain.Asset) (int, error) {
	if def := e.Rulebook.Assets[lockdown.DefinitionID]; def != nil {
		for _, entry := range def.Effects {
			if entry.LockOut != nil {
				return dice.Amount(entry.LockOut.Fee, e.Rand), nil
			}
		}
	}
	return 0, fmt.Errorf("arrival: lockdown %s has no lock_out fee", lockdown.ID)
}
```

In `lockdownHold`, change the comment's "1d4 fee" to "fee", and:

**Find:**
```go
		domain.CoinDelta{FactionID: completed.FactionID, Delta: -fee, Cause: "transport_lockdown", CausedByFactionID: lockerID},
```
**Replace with:**
```go
		domain.CoinDelta{FactionID: completed.FactionID, Delta: -fee, Cause: string(domain.SpecialLockOut), CausedByFactionID: lockerID},
```

##### Task 6 — the tests and the check

- `asset_test.go:249`, `:266`: take both results; the second also checks the asset.
- `arrival_test.go`: `lockKrylos` sets a `DefinitionID`, and `arrivalFixture`'s rulebook gains that definition with a `lock_out` entry and a `1d4` fee. `:229` wants `Cause: "lock_out"`. `fixedRoller` (`dicetest.Constant`) still gives 3, so the fees are unchanged.
- New: a lockdown whose definition has no fee returns an error. A faction that cannot pay still rolls one die.
- `ignore_coin_loss_test.go:91`: use `"lock_out"`.
- `rulebook_test.go:977`: `wantEffects` carries `LockOut`. New: `lock_out` with no block is refused.
- Any scenario with a hand-built lockdown gets its data fixed the same way. `lock_out_test.go` is unaffected, because registration does not read the fee.
- **Run the app:** re-copy `sandbox-claude`, run a cycle in tmux and confirm the rulebook loads. The fixed-sequence tests have already covered the rolls.

##### Commit message

```
refactor: lockdown fee and cause move to data

- lock_out carries its fee; arrival reads it from the lockdown's entry
- the cause is the effect's name; players see no change
- LockedOutBy returns the asset and walks in sorted order
- sandbox rulebook re-seeded
```

### Phase 9 — Docs

**In plain words:** a new architecture page explains the routes and the rule for adding a roll, and the old pages stop pointing at code that moved.

#### Commit 16 — `docs: add dice architecture page and update roll references`

Load `/gestalt:style-guide` and, for the `hooks.md` diagram, `/gestalt:diagram-kit`. Write against the source as it stands, from a fresh read.

##### Task 1 — `docs/architecture/engine/dice.md`

New page, from `templates/arch-page.md` in the session-modes skill. It holds:

- **Shape:** the seven terms from `E.004`, the route table, `Engine` and `Side`, `Outcome`'s three readings, and the data shape for each ability roll.
- **Taken from `hooks.md`:** the `RollWithHooks` dispatch description, the scope order, per-roll budgets on the roll path, and the Key Decisions below.
- **Key Decisions moved from `hooks.md`:** elective rerolls keep per-roll timing; the roll dispatcher reads every faction bucket; a roll with no collector asks no one.
- **New Key Decisions:** Decisions 1 (rerolls reach stat rolls only, with the three entries named), 4, 5 and 6 of this plan. The route rule itself ("a new roll calls a route or extends one") is added at the initiative close (`E.004`).

##### Task 2 — `docs/architecture/engine/hooks.md`, `hooks/doc.go`

- `hooks.md`: Cat 1, 2 and 5 rows and the `RollWithHooks` bullet point to `dice.md`. This also settles the conflict at lines 111-115 against the code. The `ResolveTie` bullet at lines 187-189 points to `dice.md`, and states the walk `rules.go` actually uses. Update the diagram's `Roll` and `Tie` nodes, and the Dependencies section.
- `hooks/doc.go:31`: `Dispatch logic lives in the hooks/dispatch sub-package; the roll families dispatch from engine/dice.`
- Add `dice.md` to `architecture-overview.md`, and to `engine/overview.md` if it lists the engine's packages.

##### Task 3 — `orchestrator.md`, `ability-catalog.md`

- `orchestrator.md:25`: `Rand dice.Roller`.
- `ability-catalog.md:93-94`: the rows name `AbilityStep.Table`. Add a row for `AbilityStep.Amount` and one for `AbilityStep.TargetRoll`. The `roll-table-coin` row moves Harvesters to a target roll.

Tick the Phase 9 box in this commit. Ticking the last box does not close the initiative: `/gestalt:initiative-close R.028` does, and it removes the `B.010` row, trims the `R.026` row and adds the route rule to `dice.md`.

##### Commit message

```
docs: add dice architecture page and update roll references

- new engine/dice.md takes the roll sections and decisions from hooks.md
- records the reading that rerolls reach stat rolls only
- hooks, orchestrator and catalog pages use the new names
```

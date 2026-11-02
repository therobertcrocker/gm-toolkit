# Unifying Dice and RollHooks — Exploration

> **Initiative:** `E.004`. It runs in the shape of the `E.002` pilot, and it adds to the
> [`D.009` findings log](../discovery/D.009-exploration-type-findings.md) at close.
> **Feeds:** `R.025` (the roll card), which changes the same roll code.
> **Status:** complete (reviewed by Robert). The audit and the options ran in one session,
> on `main`, at Robert's call. Rows are filed.

## In Plain Words

- **What:** the engine rolls dice in many places. Each new contest made us decide again
  whether its roll goes through `RollWithHooks`. This exploration maps how that works
  and asks if the logic should move into one shared utility, perhaps `contests`.
- **Short answer:** yes. One new package, `engine/dice`, gives each type of game roll a
  named route. Stat rolls (`1d10 + stat`) take over `RollWithHooks`, the two gates and the
  tie rules. Amount, table and target rolls get routes with no hooks, because no rule
  changes them. An ability's step splits the same way, so it says which type it rolls. A
  new roll calls a route or extends one, and never builds its own.
- **Why it matters:** the line between the two kinds of roll exists, but only habit keeps
  it. Eleven sites build the hooked roll by hand, and two functions state one gate rule.
  With named routes, the only question left is which kind of roll it is. Players see no
  change.
- **How it lands:** one refactor row, no rule changes, before the roll card (`R.025`). The
  card then changes one place, and can show a defense roll the total it must beat.

<br/>

## Questions

Robert's prompt: while we defined the contest types, we had to keep deciding what goes
through `RollWithHooks`. The idea to test is a unified utility, perhaps called
`contests`, where this logic lives. After reading the audit, Robert sharpened it: the
engine has three types of roll (stat, amount, plain), and each time a type comes up we
should call an existing utility or extend it, not build it again.

1. **Setup.** How does each roll in the engine work, and what decides its route?
2. **Warrant.** Does this justify a shared utility?
3. **Shape.** If it does, what moves into it, and where does the line fall between
   contests, single hooked rolls and amount rolls?
4. **Order.** How does this work sit with `R.025` and the other open rows on this code?

**Deciding criterion, from the prompt:** a new contest should not need a fresh decision
about its roll path. Robert has not set any other criterion yet.

<br/>

## Current State

From reading the code. Nothing was run for this audit.

### Three routes for a die

| Route | Production sites | Hooks reach it | Used for |
|---|---:|---|---|
| `dispatch.RollWithHooks` | 11 | Yes | Rolls of `1d10 + stat` |
| `DiceRoll.Roll` | 11 | No | Amounts, tables and targets |
| `Roller.Roll` directly | 2 | No | One fixed die and the turn order |

`RollWithHooks` and `DiceRoll.Roll` also call `Roller.Roll` inside them (`roll.go:73`,
`:76`, `:108`; `domain/roll.go:14`). Those calls are the route's own plumbing, and the
table does not count them.

**Amount rolls (`DiceRoll.Roll`):**

| Site | Rolls |
|---|---|
| `attack.go:176` | Attack damage |
| `attack.go:222` | Counterattack damage |
| `attack.go:276` | Counterstrike damage |
| `expand_influence.go:246` | Free attack damage on a new Base |
| `ability/stage.go:161` | An ability's die: an amount, a table or a target |
| `capture_target.go:76` | Coin from a capture |
| `coin_gain.go:35` | Coin at turn start |
| `coin_gain.go:102` | Coin on a combat loss |
| `coin_steal.go:74` | Coin stolen |
| `self_damage.go:65` | Damage to the asset itself |
| `xp_roll.go:53` | Exchange Consulate's XP roll |

**Direct rolls (`Roller.Roll`):**

- `arrival.go:60` — `e.Rand.Roll(4)`, the lockdown cost. The die is fixed in Go, not read
  from the rulebook. `R.026` named this case, and `R.028` now takes it.
- `turn/order.go:19` — the starting index of the faction order. This is not a game roll
  in the book's sense.

**How the types sort, from Robert's review:**

- **Stat rolls** are the first route. All the shared logic sits here.
- **Amount rolls** set a number: damage, Coin, and two ability dice (Commodities Broker,
  Seditionists). Each site is one call, so nothing is rebuilt today. A flat effect amount
  such as `1` loads as a zero-dice roll (`rulebook/dice.go:40`), so it rolls no dice.
- **Table rolls** pick a row: Postech Industry, Venture Capital and Pretech Manufactory.
  `outcomeFor` holds the logic and has two callers.
- **Target rolls** beat a number: Exchange Consulate's XP roll. Harvesters (3–6) and Teach
  Planetary Population (12) are targets written as one-row tables.
- **Plain dice** are nearly empty. The lockdown's `Roll(4)` is a 1d4 Coin amount held in
  Go, so `R.028` turns it into an amount roll. That leaves turn order.

**One call hides three types.** `ability/stage.go:161` serves amount, table and target
abilities alike. `resolveRoll` rolls before the handler runs (`ability/dispatch.go:144`),
and the handler reads the number. Two of the seven ability dice sit in a `RollTable` that
has no rows.

### Hooked rolls

The 11 sites form six kinds of roll. All six kinds roll `1d10 + stat`.

| Kind | Site | Rolls | Collector | Tie or comparison |
|---|---|---|---|---|
| Attack matchup | `attack.go:123`, `:140` | Attack, defense | Real | Both hit, unless a resolver answers |
| Counterstrike | `attack.go:257`, `:262` | Attack, defense | Real | Attack lands, unless `TieDefenderWins` |
| Free attack on a Base | `expand_influence.go:211`, `:225` | Attack, defense | Real | Attack lands, unless `TieDefenderWins` |
| Expand Influence | `expand_influence.go:96`, `:116` | Test, then one per rival | Real | A rival that ties or beats may attack; no resolver |
| Ability gate | `ability/stage.go:116`, `:132` | Attack or test, from data | Real | Defender wins, unless an attack and `TieAttackerWins` |
| Effect gate | `effects/contest.go:78`, called twice | Attack or test, from data | Nil | Defender wins, unless an attack and `TieAttackerWins` |

**Data for the two gates:**

| Asset | Gate | `roll` | `roller` |
|---|---|---|---|
| Informers | Ability | `attack` | — |
| Seductress | Ability | `attack` | — |
| Marketers | Ability | `test` | — |
| Tripwire Cells | Effect | `attack` | `self` |
| Panopticon Matrix | Effect | `test` | `target` |
| Lobbyists | Effect | `test` | `self` |

`contestPasses` has three production callers: `clear_stealth.go:58` and `:134`, and
`withdraw_permission.go:72`.

**Fewer tie rules than sites.** The table has four readings, but three kinds share more
than it shows. Attack matchup, counterstrike and free attack use one test for "the attack
lands" (`attack.go:165`, `:270`; `expand_influence.go:241-244`). The two gates share
another. Only the Attack matchup has a counter, and Expand Influence compares plain
totals. So three named rules cover every contest.

### What decides the route

No code or doc states the rule. Reading the sites, the line is:

- **A stat roll is hooked.** All 11 stat-roll sites use `RollWithHooks`. The book backs
  this: tags add a d10 to attack and defense rolls (`swn-faction-mechanics.md:146`) and to
  tests (`:196`).
- **An amount roll is not hooked.** A text search of the SWN rulebook and the rules doc
  found no rule that changes a damage, Coin or XP die. This was a pattern search, not a
  read of every entry.

Nothing enforces this line. `DiceRoll.Roll` is open to any caller, so a new stat roll can
skip the hooks. That is what `R.023` was: three contests that rolled bare dice.

**The decisions that drew the line:**

| Decision | What it settled |
|---|---|
| `A.001.7` Discovery D-16 | The three `R.023` contests move onto `RollWithHooks` |
| `A.001.7` overview Decision 2 | They pass a nil collector until a roll card reaches them |
| `A.001.7` Effort 3 Decision 3 | `contestPasses` rolls each side as a hooked roll |
| `A.001.7` Effort 3 Decision 4 | Effect contests name `attack` or `test`, as the book words them |
| `A.001.7` Effort 3 Decision Record 2 | An attack contest's tie asks the tie resolvers |
| `A.001.7` Effort 3 Decision 5 | Only the detecting side's roll is a detection roll |
| `A.001.7` Effort 5 Decision 10 | Ability contests also name `attack` or `test`; the field moves to `domain.Contest` |

These decisions came one at a time inside one initiative: in its discovery, its overview
plan and two of its efforts. The later ones made the two gates more alike.

### Repetition

- **The base roll.** `domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: …}` is written at
  all 11 sites. No named helper states "a stat roll".
- **The stat read.** It has three spellings that give the same number:
  - `statScore(…)`, a wrapper that only calls `Faction.StatScore`, defined twice
    (`actions/repair_asset.go:117`, `ability/stage.go:191`)
  - `actor.StatScore(…)` (`contest.go:77`)
  - `faction.Cunning` and `rival.Cunning` (`expand_influence.go:104`, `:126`)
- **The pair.** An attack roll followed by its defense roll is written out in four
  places: Attack matchup, counterstrike, free attack and ability gate. Only `contest.go`
  has a helper, `side`.
- **The call.** `RollWithHooks` takes 8 parameters:
  - At all 11 sites, `faction` is the same faction as `ctx.Actor`.
  - `registry`, `collector`, `roller`, `factionState` and `rulebook` are engine context
    that each caller must carry.
- **Carrying the context.** Holders store a registry and a roller only to reach this
  call (`clearStealthReactor`, `permissionContester`). `FireContext` gained `Registry` and
  `Rulebook` for the same reason (Effort 3 Decision 3).
- **The two gates.** `passesGate` (`ability/stage.go:85`) and `contestPasses`
  (`contest.go:40`) state the same rule: attack phases or test phases, a tie goes to the
  defender, and an attack tie asks the resolvers.
- **Test fakes.** Ten roller fakes exist, nine in test files and `testharness.FixedRoller`.
  They have four behaviours: three strict sequences, three cycling sequences, three
  constants and one real random (`turn_test.go`).

**How the two gates differ:**

| Point | `passesGate` | `contestPasses` |
|---|---|---|
| Who rolls as attacker | The acting faction, always | `roller`: declaring side or target |
| Collector | The action's | Nil at all three callers |
| `Action` | `"Use Asset Ability"` | Empty |
| `Detection` | The actor, on `reveal_stealth` | The declaring side, on `clear_stealth` |
| Missing target faction | Returns an error | Not a case: each caller has found the target |

### Differences that must survive

Any shared code has to keep these. They are rules or recorded decisions, not accidents.

- **Tie readings.** See the table under Hooked rolls. The book gives Attack a both-hit
  tie (`swn-faction-mechanics.md:150`) and gives tests a tie to the defender (`:196`).
- **`Action` is load-bearing.** Two rulebook entries filter on it: `tags.toml:140`
  (Imperialists, `action = "Attack"`) and `wealth_assets.toml:183` (`action = "Expand
  Influence"`). Effort 3 Decision 4 used this filter to keep Imperialists' die off
  Tripwire Cells, so the empty `Action` on effect contests matters.
- **Who is the opponent.** Four fence-signs or doc notes hold this in place:
  - No `Opponent` on the expanding faction's own roll: one roll is compared against N
    rivals (`expand_influence.go:95`).
  - No `Opponent` on a counterstrike: it would ask each asset hook on the world twice
    (`attack.go:256`).
  - No `Asset` when a Base defends (`hooks/types.go:21`).
  - Effect contests can have no target asset (`contest.go:12`).
- **The real faction.** The defender must be the state's faction, not a view, or a hook's
  budget spend is lost (`ability/stage.go:104`).
- **Per-roll reroll timing.** A reroll is decided right after each roll, not after both
  sides roll (`roll.go:88`). This is a fence-sign, and it is why each side is one call.
- **Keep-highest assumes one d10** (`bonus_die.go:73`). This is a fence-sign too.

### Findings from the audit

- **Dropped dice.** Each of the 11 sites reads only `.Sum`. `Dice` and `Modifier` reach
  the Cat 2 hooks and stop there. The GM cannot see them, which is the root of `F.032`.
- **Two similar names.** `domain.ContestRoll` is `attack` or `test`. `effects.contestRoll`
  is the struct for one contest. They differ only by a capital letter. `PermissionContest`
  and `PhaseContested` also use the word for other things. The Terms block under
  Options keeps these apart.
- **`ResolveTie` runs on every Attack matchup** (`attack.go:164`), not only on a tie. The
  other sites ask only on a tie. It has no side effects in the code read, so the result
  does not change.
- **`B.010` is still inert.** All 14 `sides` keys in the SWN data are 10. The row's line
  reference is stale: the reroll is now at `roll.go:108`, not `:86`.
- **`hooks.md` conflicts with the code.** Lines 111–115 say effect contests roll directly
  and never read the opponent's faction bucket. `contest.go:78`, `rollScopes` and the
  page's own Key Decision at line 449 say otherwise. `D.010` already covers this page, so
  this exploration flags the conflict and does not fix it.

### Related rows

| Row | Overlap |
|---|---|
| `R.025` | Roll card: one ask before each roll and one after. Replaces `SelectModifiers` and `ConfirmReroll` |
| `F.037` | Contests that ask: gives the effect gates a real collector, after `R.025` |
| `F.032` | Reroll prompt shows the roll: needs the dice that the sites drop |
| `F.025` | Per-faction roll confirmation: routes each directive to its own driver |
| `B.043` | `rival_roll` asks the wrong player, to settle with `R.025` |
| `B.024` | Book of Secrets rerolls twice a turn: a budget key rules call |
| `B.016` | Hook budgets survive an aborted action |
| `B.010` | A reroll uses the base die's sides |
| `D.010` | Refreshes `hooks.md`, among other pages |
| `R.026` | Rule values held in Go; the lockdown's 1d4 moves from it to `R.028` |
| `A.001.5` | H.A.W.K authoring: its assets would copy today's roll data shape |

<br/>

## Existing Patterns

Patterns already in the engine that a shared utility could follow.

### `engine/relocate`

The closest precedent. Relocation had two routes: an action that spends a turn and a free
move in a phase. One package now holds the logic for both.

- **One request value.** `relocate.Request` carries the faction, acting asset, params,
  world, rulebook, registry, state and collector.
- **A narrow collector.** `relocate.Collector` lists only the prompts a relocation asks.
  Both the action collector and the phase collector satisfy it.
- **Each route keeps its own policy.** A shared error (`ErrNoCandidates`) lets the action
  route surface a case that the free route skips.

### `contestRoll`

`effects/contest.go` already has the shape, for one of the two gates: one value holds the
contest, both sides, the world and the context, and `side` rolls one side. Its limit is
its package. `ability` does not import `effects`, so `passesGate` cannot call it.

### Nil collector

`dispatch.InterceptLethal` set the precedent, and `RollWithHooks` followed it (overview
Decision 2). A caller with no one to ask passes nil, and the dispatcher takes only what
needs no answer. Each call site passes its collector as one argument, so `F.037` can
swap it.

### Where a package could sit

`hooks/dispatch` imports `domain`, `engine/hooks`, `rulebook` and `state`. The packages
with hooked rolls (`action/actions`, `action/actions/ability`, `effect/effects`) all
import `dispatch`. So a new package directly above `dispatch` creates no import cycle if
it imports only those same packages.

### Data side

`domain.Contest` already serves both gates, and `EffectContest` embeds it and adds
`Roller`. The loader refuses a contest with no `roll` key. So the data has one shape, and
only the Go that rolls it is split.

<br/>

## Options

Seven words, used the same way in every option below (Robert's call, in review):

- **stat roll** — one `1d10 + stat` roll through the hooks
- **contest** — two stat rolls compared: an attack and its defense, or a test and its
  contested roll
- **gate** — a contest that decides if an ability or effect goes ahead; it deals no damage
- **amount roll** — dice from the rulebook whose result is the number; no hook reaches it
- **table roll** — dice from the rulebook whose result picks a row; no hook reaches it
- **target roll** — dice from the rulebook that must beat a number; no hook reaches it
- **plain die** — a `Roller.Roll` call that is not a game roll: turn order

**In every option, no rule changes.** Each one keeps the items under "Differences that
must survive". What differs is how much of the logic moves.

### How much moves

#### C-0 — No move; write the line down

State the rule in `hooks.md` and in a comment on `RollWithHooks`. Change no code.

- **For:** no cost, no risk.
- **Against:** a doc is the only guard, which is close to today. The repetition stays, and
  `R.025` and `F.032` still reach through each site.

#### C-1 — A stat roll helper only

Add one helper that builds the `1d10 + stat` roll and calls `RollWithHooks`, and use it at
all 11 sites.

- **For:** small. Removes the literal, the three stat spellings and the redundant `faction`
  parameter.
- **Against:** the two gates stay split and the tie rules stay at each site. A new contest
  still needs a decision about how to pair, tie and gate.

#### C-2 — The helper plus one gate

C-1, and both gates call one shared gate. The Attack action and Expand Influence keep
their own pairs and ties.

- **For:** removes the clearest duplicate rule. A new gate needs no Go beyond its effect.
- **Against:** the shared code covers some contests but not others, so a new contest still
  has to ask which kind it is. `F.032`'s "attack total to beat" still has to pass through
  the Attack sites.

#### C-3 — One home for every stat roll

Everything in C-2, and the pair and the tie rules move too. All 11 sites go through it.
Expand Influence uses its single stat roll and keeps its plain comparison. A sketch of
the shape, not of final names:

```go
// Stat is one stat roll: 1d10 + the side's stat, through the hooks.
func (engine *Engine) Stat(phase hooks.RollPhase, side, opponent Side, world, action string) hooks.RollResult

// Contest rolls a pair, attacker first, and asks the tie resolvers on an attack's tie.
func (engine *Engine) Contest(request Request) Outcome

func (outcome Outcome) AttackLands() bool  // Attack, counterstrike, free attack
func (outcome Outcome) CounterLands() bool // Attack
func (outcome Outcome) GatePasses() bool   // both gates, with the target-roller flip
```

- **For:** the only option that meets the criterion for a contest written in Go. There is
  one route, and each tie rule has one name. An `Outcome` keeps both sides' dice for the
  roll card.
- **Against:** the largest move. Five shapes of `RollContext` must fit one `Request`:
  no `Opponent` on a counterstrike, no asset on a defending Base, an empty `Action`,
  detection on one side, and the target-roller flip.
- **Limit:** holders still carry a registry and a roller, because a reactor's fire route
  gets them at registration and state at fire time. `F.037` owns that problem.

#### C-4 — One route for every die

C-3, and amount rolls go through hooks as well.

- **For:** the most literal "unified" reading. A future rule could change damage.
- **Against:** it builds a capability no rule asks for. An amount roll needs a `RollContext`
  and a collector that reactors do not have.

#### C-5 — One package, a named route for each type

From Robert's review. C-3, and amount, table and target rolls each get a named route in
the same package, with no hooks. `DiceRoll.Roll` leaves `domain`. A sketch of the routes:

```
Stat     1d10 + stat, through the hooks
Contest  two stat rolls compared
Gate     a contest that lets something go ahead
Amount   the die is the number
Table    the die picks a row
Target   the die must beat a number
```

- **For:** it answers the original problem for every game roll. The only question left is
  which type a roll is. It also closes C-3's gap: with `DiceRoll.Roll` gone, a bare d10
  needs `Roller.Roll` directly or the amount route, and either one stands out in review.
- **Against:** the amount and target routes are clearer names, not shared logic. The
  package needs a name wider than `contest`.
- **Limit:** only 11 production sites call `DiceRoll.Roll`, and no tests do, so the method
  can go.
- **The lockdown's `Roll(4)` becomes an amount roll,** with its die and its asset-named
  cause moved to data, which takes that known case from `R.026` (Robert's call).
- **`B.010` is fixed in the same work.** The move rewrites that reroll loop, so each die
  keeps its own sides (Robert's call).
- **Harvesters and Teach Planetary Population become target rolls,** as the book words them
  (Robert's call: "if it fits it sits"). Their results do not change. So target has three
  users and one data shape, and the other three assets with rows stay real tables.
- **The ability step splits by type (spot check, Robert's call).** `AbilityStep.Roll` holds
  all three types in one `RollTable`, so amount abilities carry a table with no rows, and
  the loader needs a check that Teach's rows only mark a success (`rulebook.go:424-429`).
  The step gets an amount, a table or a target, and the target shape is shared with
  `XPRollParams`. Without this, the routes change but the data type still mixes the types.

**Compared:**

| Option | New contest in Go needs a fresh decision? | Gates become one | Tie rules named once | Helps `R.025` / `F.032` | Amount rolls get a named route | Work |
|---|---|---|---|---|---|---|
| C-0 | Yes | No | No | No | No | None |
| C-1 | Yes, for pair and tie | No | No | A little | No | Small |
| C-2 | Yes, outside gates | Yes | Gates only | Partly | No | Medium |
| C-3 | No | Yes | Yes | Yes | No | Medium to large |
| C-4 | No | Yes | Yes | Yes | Yes, with hooks | Large, plus new design |
| C-5 | No | Yes | Yes | Yes | Yes, no hooks | Medium to large |

**Contest data needs no change in any option.** A contest declared in a rulebook already
gets the hooks with no Go. Only C-5 changes data, for the abilities that roll an amount or
a target.

### Where it lives

These apply to C-1, C-2, C-3 and C-5. All 25 calls in `roll_test.go` follow
`RollWithHooks` wherever it goes. The helpers it uses (`hookOwner`, `assetScopes`,
`scopeFactionID`) are used only in `roll.go`, so the file can move whole. "The package"
below is the new package, whatever its name.

#### P-1 — A new package over `dispatch`

The package calls `dispatch.RollWithHooks`, which stays exported.

- **For:** no test moves. Follows `engine/relocate`.
- **Against:** the second route stays open, and Robert's aim is one home for the logic.

#### P-2 — A new package that takes over the roll

`roll.go` and `roll_test.go` move into the package, and `RollWithHooks` becomes unexported
there. `dispatch` keeps Cat 3 to Cat 6 and `ResolveTie`, which the package imports.

- **For:** the only way to make a hooked roll is through the package. All the roll logic
  sits in one place.
- **Against:** seven calls in the `effects` tests change. The `hooks.md` dispatch section
  changes.

#### P-3 — A file inside `dispatch`

`dispatch/contest.go`. No new package.

- **For:** the least movement. Every site already imports `dispatch`.
- **Against:** `dispatch` runs hooks. Attack phases and tie rules are game rules, so it
  would hold two jobs, and the name Robert wanted is lost.

#### P-4 — A new package that takes over everything it can

From Robert's review. P-2, and everything else dice-related moves too, where the import
graph allows it. `dispatch` then keeps Cat 3, Cat 4 and Cat 6.

| Item | Today | Result |
|---|---|---|
| `RollWithHooks` and its helpers | `dispatch/roll.go` | Moves |
| `ResolveTie` | `dispatch/rules.go` | Moves; every production caller moves too |
| The gates, the pair and the tie rules | Five places | Move |
| `RandRoller` | `engine/roller.go` | Moves; only `core.go:52` uses it |
| `Roller` | `domain/roll.go` | Moves; only engine code uses it (44 references) |
| `resolveRoll`, `outcomeFor` | `ability/stage.go` | Move |
| `parseDice`, `parseAmount` | `rulebook/dice.go` | Stay: `rulebook` importing the package closes a cycle through `hooks` |
| `DiceRoll`, the roll data types | `domain` | Stay: `domain` types hold them, and `domain` cannot import the engine; `RollTable` splits by type (C-5) |
| `RollContext`, `RollResult`, hook interfaces | `hooks` | Stay: the package imports `hooks`, and the registry holds them |
| `RollContext.Attribute`, `OpponentAttribute` | `string` in `hooks` | Become `domain.FactionStat`; the conversions go (`roll_match.go:72`, `:75`) |
| `hooks.RollPhase`, `domain.EffectRollPhase` | Two enums joined by a map (`roll_match.go:11`) | Become one in `domain`, because data names the phases; the map goes |
| `statScore` wrappers | `repair_asset.go:117`, `ability/stage.go:191` | Deleted; `repair_asset.go:72` calls `Faction.StatScore` |
| `rollMatches`, `pickDice` | `effects` | Stay: they decide when a hook fires |
| `diceLabel` | TUI overlay | Stays: display code, in `E.001`'s area |
| Roller fakes | Tests and `testharness` | Three shared fakes in a small `dicetest` package; the real random in `turn_test.go` stays |

**The line this draws:** `domain` and `rulebook` state what dice are, and `engine/dice`
rolls them.

**Rejected: a low-level `faction/dice` package** for the type, the parser and `Roller`,
with no import of `hooks`. It would bring the parser and the type together, but hooked
rolls would need a second package, and all 91 `DiceRoll` references would change (Robert's
call: one package).

**No option can forbid a bare d10.** P-2 removes the second hooked route. C-5 removes
`DiceRoll.Roll`. A bare d10 then needs `Roller.Roll` directly or the amount route, and a
code review catches it. A check to forbid it is more than the problem needs.

### Order against the roll card

#### O-1 — This first, then `R.025`

- **For:** `R.025` changes one roll function and one contest function. `F.032`'s "attack
  total to beat" has one home, because only the pair knows both totals. `F.037` and `F.025`
  get one place to pick a collector.
- **Against:** `R.025` waits for a refactor that players do not see.

#### O-2 — `R.025` first, then this

- **For:** the visible feature lands sooner.
- **Against:** the defense total has to pass through four pair sites and then move again.

#### O-3 — Do this inside `R.025`

- **For:** one branch, one review.
- **Against:** a no-rule-change refactor mixed with a behaviour change is harder to check.
  Today's tests guard the refactor only if they stay unchanged while it runs.

<br/>

## Recommendation

All the calls below are Robert's, from review.

- **How much: C-5.** One package, a named route for each type, hooks on stat rolls only. It
  meets the criterion for every game roll. C-1 and C-2 leave the question open for the
  contests they skip. C-3 leaves the other rolls with no named route. C-4 builds what no
  rule asks for.
- **Where: P-4, named `engine/dice`.** Everything dice-related moves where the import graph
  allows it, with no second route. `contest` would be too narrow a name for amount rolls.
  The type and the parser stay below the engine.
- **Order: O-1.** It costs `R.025` a short wait and saves threading the same data twice.
- **H.A.W.K waits too.** `A.001.5`'s authoring passes run after `R.028`, or they copy the
  old one-row tables. Its lore-seed session writes no roll data, so it can start at any time.
- **Plain dice stay outside.** Turn order is not a game roll, and it is the only plain die
  left once the lockdown die becomes an amount roll.

**Keep it behaviour-preserving.** The existing action, effect and harness tests are the
guard, and they use fixed roll sequences, so the order of rolls must not change: attacker
first, with the flip for a target roller. `ResolveTie` moves to run on a tie only. It has
no side effects in the code read, so results stay the same. Splitting `resolveRoll` by type
changes when an ability's die is rolled, relative to its handler, so the plan must show
that the order of rolls stays the same.

### How it lands

- **`R.028` — one refactor row.** It follows the refactor flow: the pattern, then a unit
  list of Attack, Expand Influence, ability gate, effect gate, and the amount, table and
  target rolls. With the ability step split, this is more than a package move, so its plan
  should ask if it needs efforts.
- **Convert Harvesters and Teach Planetary Population in the same work.** Abilities gain a
  way to name a target roll, `coinDelta` and `raiseTechLevel` read a pass or fail, and the
  sandbox rulebook is re-seeded. It cannot be a flat `target` key, because `target = "world"`
  already uses that name (`tags.toml:271`). A nested block like `xp_roll` is the likely
  shape, and Robert decides it in the plan, along with where Harvesters' `coin = 1` goes.
- **Split the ability step.** The loader checks on `step.Roll` change with it
  (`rulebook.go:417`, `:421-429`, `:440`). The plan settles whether the two amount
  abilities' data keys change, and what replaces the single `StepContext.RollResult`.
- **Share the fakes.** Tests move to `dicetest` as each unit moves.
- **Move the fence-signs with the code.** They stay binding: `attack.go:256`,
  `expand_influence.go:95`, `ability/stage.go:104`, `roll.go:88`.
- **Update the docs.** A new `docs/architecture/engine/dice.md` takes the roll sections from
  `hooks.md`. `hooks/doc.go:31`, `orchestrator.md:25` and `ability-catalog.md:93-94` change
  to the new names. Writing the roll section against the source also settles the
  `hooks.md:111-115` conflict, if this lands before `D.010`.
- **At its close,** `dice.md` gets the Key Decision: only a stat roll takes the hooked
  route, each other game roll takes its named route, and a new roll calls a route or
  extends one, never builds its own.

### Rules calls to carry (Robert)

None. Every option keeps today's readings, and the audit found no new rules question.

<br/>

## Follow-on Rows

Filed in `planned-work.md` after Robert's review: `R.028` under Refactors (Engine), and
the row changes below.

| ID | Item | Type | Detail |
|---|---|---|---|
| `R.028` | Dice package | `refactor` | New `engine/dice` gives each game roll a named route: stat rolls (`RollWithHooks`, the pair, the tie rules, one gate), and amount, table and target rolls (no hooks; `DiceRoll.Roll` leaves `domain`); `RandRoller` and `Roller` move too; all 22 sites move; `AbilityStep.Roll` splits by type; Harvesters and Teach Planetary Population become target rolls; absorbs `B.010` and the `lock_out` case from `R.026`; one roll phase enum, a typed stat, shared fakes in `dicetest`, no `statScore` wrappers; a new `dice.md` page (`E.004` C-5, P-4) |
| `A.001.5` | H.A.W.K asset authoring (amend the row) | -- | Its authoring passes run after `R.028`; the lore-seed session stays startable |
| `R.025` | Roll card (amend the trigger) | -- | Trigger becomes `after R.028` (`E.004` O-1) |
| `B.010` | Reroll uses base die sides (amend the row) | -- | Trigger becomes `with R.028`, and the stale `roll.go:86` reference becomes the `RollWithHooks` reroll loop |
| `R.026` | Faux-data-driven rule audit (amend the row) | -- | `R.028` takes the `lock_out` known case; the rest of the audit stays here |
| `D.010` | Refresh architecture pages (amend the row) | -- | Add the `hooks.md` conflict: lines 111–115 against the code and the Key Decision at line 449 |

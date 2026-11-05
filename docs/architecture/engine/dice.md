# Dice

> **Code:** `internal/faction/engine/dice/`, `internal/faction/engine/dice/dicetest/`

## Purpose

`engine/dice` is the one home for dice. Each type of game roll has a named route:
stat rolls, with the contests and gates built from them, go through the hooks;
amount, table and target rolls do not. An ability's step says which type it rolls,
so its handler reads a typed result and not a bare number. This page is the home
for the routes, the hooked roll's dispatch and the data shape of each ability roll.
The hook families and the registry live on the [hooks](hooks.md) page.

## Shape

### Seven words

- **stat roll** — one `1d10 + stat` roll through the hooks
- **contest** — two stat rolls compared: an attack and its defense, or a test and
  its contested roll
- **gate** — a contest that decides if an ability or effect goes ahead; it deals no
  damage
- **amount roll** — dice from the rulebook whose result is the number; no hook
  reaches it
- **table roll** — dice from the rulebook whose result picks a row; no hook reaches
  it
- **target roll** — dice from the rulebook that must meet or beat a number; no hook
  reaches it
- **plain die** — a `Roller.Roll` call that is not a game roll: turn order

### The routes

| Type | Route | Hooks | Result | Used by |
|---|---|---|---|---|
| Stat roll | `Engine.Stat` | Yes | `hooks.RollResult` | Expand Influence, the expander and each rival |
| Contest | `Engine.Contest` | Yes | `Outcome` | Attack, the counterstrike, the free attack |
| Gate | `Engine.Gate` | Yes | `bool` | Ability contests, `clear_stealth`, `withdraw_permission` |
| Amount roll | `Amount` | No | `int` | Damage, Coin effects, the lockdown fee, amount abilities |
| Table roll | `Table` | No | The matched row, or none | Table abilities, through `RollStep` |
| Target roll | `Target` | No | `bool` | `xp_roll`, target abilities through `RollStep` |
| Plain die | `Roller.Roll` | No | `int` | Turn order (`turn/order.go`) |

`RollStep` is the ability step's front door: it picks `Amount`, `Table` or `Target`
by the step's field and returns a `StepResult`. `Roller` is the interface every
route rolls through. `RandRoller` is the production one, and `dicetest` holds the
shared fakes (`Sequence`, `Cycle`, `Constant`).

### Engine and Side

The hooked routes are methods on `Engine`, which holds what the game carries: the
hook registry, the collector, the roller, the faction state and the rulebook. A
`Side` is who rolls: a faction, its asset if one stands there, the stat, and
whether the side is testing to detect stealth. A zero `Side` as the opponent is a
roll with no one against it, which is how the expander rolls once against any
number of rivals.

`Contest` rolls the attacker first, then the defender, and picks the roll phases
from the contest's kind: attack and defense, or faction test and contested. When
both sides are one faction it drops `Opponent` from both rolls, so a hook on the
world is not asked twice. It asks the tie resolvers only on the tie of an attack.

`Gate` is a contest with a yes or no. With roller `self`, the declaring side
attacks and must win. With roller `target`, the target tests to keep what it has,
and the gate passes when that test fails. Each side's stat comes from the contest,
by the role it rolls in.

### Outcome's three readings

A contest returns an `Outcome`, which keeps both rolls and the tie answer. The rolls
are the same for every caller; only the reading of a tie differs.

| Reading | Read by | On a tie |
|---|---|---|
| `AttackLands` | Attack, the counterstrike, the free attack | Lands, unless a resolver gives it to the defender |
| `CounterLands` | Attack | Lands, unless a resolver gives it to the attacker |
| `GatePasses` | `Gate` | Fails, unless a resolver gives it to the attacker |

Expand Influence is the one plain comparison. The expander and each rival use
`Stat`, and the site compares the totals: a rival that ties or beats the expander
may make a free attack. No tie resolver is asked.

### The hooked roll

`rollWithHooks` is private, and `Stat` is its only caller. It covers both roll-time
hook families in one call:

1. Gather Cat 1 offers from each bucket, and stamp each with the faction that
   registered it.
2. Run the Cat 1 offer filters over them (Blackmail strips tag dice through
   `ModifierOffer.FromTag`).
3. Set aside any offer whose budget is spent, then show the
   [collector](hooks.md#the-collector-seam) a before card with what survives and what
   did not. It asks on every roll, offers or not.
4. Apply the chosen offers, spend their budgets, and roll the base dice and any
   extra dice. Each die keeps its own sides.
5. Apply keep-highest if an offer asked for it.
6. Run each forced Cat 2 hook, once. A directive whose budget is spent is skipped. A
   rerolled die rolls its own sides again.
7. Show the collector an after card with the dice, the rerolls applied and each
   elective reroll on offer. It takes one or passes. After a take, the card returns
   with what is left: a taken hook is not offered again on this roll, and a forced
   hook does not run again, so a reroll that lands on 1 stays.

**Scope order.** `rollScopes` lists the buckets, in order: global, the roller's
faction, the roller's assets, the opponent's faction, the opponent's assets, then
every other faction's faction bucket in ID order. A side's asset buckets are its
assets on the roll's world in ID order, plus the rolling or opposing asset when it
stands elsewhere. A bystander's asset buckets are never read.

**Per-roll budgets.** Each offer and directive is stamped with the faction that
registered it, and its budget is read and spent on that faction's `HookBudgets`, so
an opponent's effect spends its owner's budget. A global hook belongs to the
roller. See [per-turn budgets](hooks.md#per-turn-budgets) for the reset.

### Ties

`ResolveTie` asks the global hooks, then the attacker's side, then the defender's:
on each side its faction's hooks, then each of its assets' in ID order. The first
answer that is not `TieStandard` wins. It stays exported because it is not a roll,
and the `lose_ties` tests call it from outside the package.

### Ability rolls in data

A step has at most one of three fields, and the loader enforces it. An effect entry
uses the same shapes where it rolls.

| Step field | TOML | Handler reads | Abilities |
|---|---|---|---|
| `Amount *DiceRoll` | `die` with no rows | `Rolled.Amount` | Commodities Broker, Seditionists |
| `Table *RollTable` | `die` with `[[outcomes]]` | `Rolled.Row` | Postech Industry, Venture Capital, Pretech Manufactory |
| `TargetRoll *TargetRoll` | `[target_roll]` with `die`, `target` and an optional `amount` | `Rolled.Passed` | Harvesters, Teach Planetary Population |

Exchange Consulate's `[xp_roll]` is a `TargetRoll` too. The Transport Lockdown's fee
is `[effects.lock_out] amount = "1d4"`, on the shared `amountRecord`.

### Orders that must hold

- The attacker rolls first, and the `target` flip swaps who that is (`Contest`,
  `Gate`).
- An ability's die rolls after its gate and before its handler (`RollStep`, in
  `ability.Dispatch`).
- The lockdown's die rolls before the hold is decided, even when the faction cannot
  pay (`lockdownFee`, in `engine/arrival.go`).

## Key Decisions

- **A new roll calls a route or extends one, and never builds its own.** Only a stat
  roll takes the hooked route, each other game roll takes its named route, and turn
  order is the one plain die. So a new contest needs no fresh decision about its roll
  path; the only question left is which type it is. `rollWithHooks` is private, so the
  compiler holds the hooked half. Nothing forbids a bare d10 through `Amount` or
  `Roller.Roll`, but either one stands out in review. (`E.004` C-5; `R.028`.)
- **An amount on an effect entry always rolls above zero.** `convertAmount` refuses
  dice whose lowest roll is zero or less, for `coin_gain`, `coin_steal`, `self_damage`
  and `lock_out`, because a negative fee would pay the mover. `damage` and `counter`
  do not go through it, so a `1d4-1` there stays. (`R.028` close review.)
- **Rerolls reach stat rolls only.** Fanatical ("any dice"), Book of Secrets ("one
  die for an action") and Psychic Academy ("any one d10") could reach amount dice on
  their plain words. The engine rerolls stat rolls only, and that reading holds. A
  rule that needs to change an amount die would need a new hook path; nothing else
  in the book asks for one. (`R.028` Decision 1; this corrects `E.004`, whose text
  search missed the three entries.)
- **Amount abilities keep a bare `die`.** `die` alone is an amount, and `die` with
  rows is a table, so existing amount data does not change. The loader refuses a
  shape its effect cannot read: `coin_delta` with neither rows nor a target roll,
  `purchase_discount` or `block_attack` with rows, rows with no `die`, and a target
  roll on an effect other than `coin_delta` and `raise_tech_level`. (`R.028`
  Decision 4 and the close review.)
- **A target roll's amount is optional, and the effect decides what it pays.**
  Harvesters' `coin = 1` becomes the target roll's `amount`, and Teach Planetary
  Population has none. `coin_delta` requires one and `raise_tech_level` refuses one.
  "Beat" means meet or beat, as `xp_roll` always read it. (`R.028` Decision 5.)
- **A step's die rolls in one place, and the handler reads a typed result.**
  `StepContext.Rolled` is a `StepResult` with `Amount`, `Row` and `Passed`, filled
  after the gate and before the handler. A handler's early return cannot skip a die,
  so the dice a step rolls do not depend on what the handler does. (`R.028`
  Decision 6.)
- **Elective rerolls keep per-roll timing.** Cat 2 runs right after each roll, so an
  attacker decides on a reroll before the defense is rolled. Deciding after both
  rolls would split the roll into roll and post-roll calls at every call site, and
  Expand Influence compares one roll against N. A fence-sign at the roll marks it.
  (s-flag Effort 3 plan, Decision 15.)
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
- **The roll dispatcher reads every faction bucket, and asset buckets on the roll's
  world from both sides.** An S effect such as Book of Secrets or Panopticon Matrix
  acts on a roll made by or against its owner on its world, so `rollScopes` gathers
  both sides' asset buckets there, plus the rolling and opposing assets when they
  stand elsewhere. It also reads the opponent's faction bucket (defense-die tags such
  as Deep Rooted) and every bystander's faction bucket (Psychic Academy's
  `rival_roll`). Every roll hook checks its owner relation against `ctx.Actor` /
  `ctx.Opponent` itself (`rollMatches`), so a hook in another faction's bucket acts
  only on the rolls its entry names. An asset's roll entry cannot take scope
  `faction` or `homeworld`, because its bucket is read only on the roll's world. This
  replaces "never the opponent's faction bucket", whose reason was the Go tag hooks
  that assumed every roll was their own; `engine/tag` is gone. (s-flag Effort 3 plan,
  Decision 7; A.001.7 discovery D-7 and D-15, overview R-2, Effort 3 Decision 7,
  initiative close.)
- **A roll with no collector asks no one.** A contest that fires inside a reactor or
  a phase route (Tripwire Cells, Panopticon Matrix, Lobbyists) has no one to ask, so a
  nil collector takes every Cat 1 offer with no budget, skips every offer with one,
  declines every elective Cat 2 reroll and still applies forced ones. A free die where
  the highest is kept never hurts the roller; a once-per-turn die spent on a roll the
  player never chose could. Each call site passes its collector as one argument, nil
  for now, so a later collector is a one-line swap. Precedent: `InterceptLethal`'s nil
  collector. Rejected: a collector on every reactor (changes every reactor's
  interface; the roll card, `F.039`, would replace those prompts). (A.001.7 overview
  Decision 2.)

## Dependencies

**Depends on** `domain` (rolls, factions, assets), [hooks](hooks.md) (the registry,
the collector and the hook contracts), and the `rulebook` and `state` packages that
it passes to each hook. It does not import `hooks/dispatch`, and it holds no game
data.

**Used by** the [actions](actions.md) engine (Attack, Expand Influence and the
ability gate and step dice), the effect handlers on the
[effect & mutation](effect-mutation.md) page (effect gates and amount rolls), the
[orchestrator](orchestrator.md) (the engine's roller and the lockdown fee at arrival)
and the [turn pipeline](turn-pipeline.md) (turn order's plain die).

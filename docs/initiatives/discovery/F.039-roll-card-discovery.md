# Roll Card — Discovery

## In Plain Words

- **What:** on every stat roll that has a collector, the GM sees a card before the dice,
  with every extra die on offer as a toggle. After the dice, a second card shows them and
  every reroll on offer, a rival's included. With no reroll on offer, the dice go to the
  feed with no stop.
- **Why:** today's prompts do not show what is being rolled, and the reroll prompt does
  not show the dice. Psychic Academy, the last SWN tag with no data entry, can force a
  reroll on any rival's roll, so it needs a prompt that shows the roll.
- **How it fits:** it follows `R.028`, which put every hooked roll through one function,
  so the engine asks change in one place. It finishes the tags from `A.001.7`, and
  `F.037` follows it, to bring the card to the three contests that ask no one today.

## Problem

- **Each roll asks through two thin prompts.** `SelectModifiers` gets the offers alone
  (`dice/roll.go:59`), and `ConfirmReroll` gets one directive alone (`dice/roll.go:105`).
  Neither gets the roll's context or its dice, so the prompts cannot say who rolls, for
  what, or what the dice show (`overlay/hooks.go`; `F.032`).
- **The dice reach the GM nowhere else.** `RollResult` stops at the roll sites, which
  read only the total, and no mutation records the dice.
- **Rerolls ask one at a time.** A roll with two rerolls on offer asks two questions in
  hook order, and the GM cannot compare them.
- **Psychic Academy is half built.** Its registrar and tests exist, but its entry waits
  for this card (`A.001.7` Discovery D-7). Its `effect` text is the rule:
  "Once per turn, force a rival faction to reroll any one d10, whether or not they are
  involved in the roll."

## Architecture Context

Pages read: [dice](../../architecture/engine/dice.md),
[hooks](../../architecture/engine/hooks.md),
[overlays](../../architecture/interface/overlays.md),
[event stream](../../architecture/interface/event-stream.md). Also read: the `E.001`
[layer map](https://claude.ai/artifact/6UjHPwdY3mq8oQ24HRqvU1), built at an older commit,
so its sizes and counts are approximate.

Key Decisions that bind this work:

- **Elective rerolls keep per-roll timing** (dice). A reroll is decided right after its
  roll, before the other side rolls. A fence-sign marks it at `dice/roll.go:90`. This is
  why there is a card after each roll, not one after the contest.
- **Rerolls reach stat rolls only** (dice). So the card never appears on an amount, table
  or target roll.
- **A roll with no collector asks no one** (dice). The contests of Tripwire Cells,
  Panopticon Matrix and Lobbyists stay as they are until `F.037`, so Psychic Academy does
  not reach them (`A.001.7` overview Decision 2).
- **`rollWithHooks` is the one hooked roll** (dice). The asks change there and nowhere
  else.
- **A budget lives on its owner's `HookBudgets`** (hooks), and it clears in the owner's
  own bookkeeping. **The same decision puts the trigger in the key, so Book of Secrets'
  two entries do not share a budget. `B.024` reverses that part** (D-11).
- **A prompt inside an action that the action cannot unwind ignores Esc** (overlays).
- **`askEntries` is the one table of asks** (overlays). A new ask needs a row, and a test
  walks every kind.
- **The adapter copies everything the views render** (event stream). D-12 builds on it.
- **The message enums are append-only** (event stream). The two card asks must not reorder
  `AskKind`.

## Audit Findings

### Where the asks live

| What | Where | Count |
|---|---|---|
| The engine's ask of each kind | `dice/roll.go:59`, `dice/roll.go:105` | 1 each |
| Production collectors | `tui/adapter/action_collector.go:64`, `tui/adapter/smoke.go:80`, `engine/testharness/collector.go:240` | 3 |
| Generated mock | `action/actions/mocks/mock_collector.go` | 1 |
| Test stubs | `dice/roll_test.go`, `effects/bonus_die_test.go`, `ability/dispatch_test.go` | 3 files |
| TUI asks and overlays | `AskSelectModifiers`, `AskConfirmReroll`, `overlay/hooks.go` | 2 |

### Rolls that reach the card

| Roll | Site | Collector |
|---|---|---|
| Attack action | `actions/attack.go:128`, `:224` | Yes |
| Expand Influence, the expander and each rival | `actions/expand_influence.go:100`, `:115` | Yes |
| The free attack | `actions/expand_influence.go:206` | Yes |
| Ability gate | `ability/stage.go:111` | Yes |
| Tripwire Cells, Panopticon Matrix | `effects/clear_stealth.go:67`, `:142` | Nil (`F.037`) |
| Lobbyists | `effects/withdraw_permission.go:81` | Nil (`F.037`) |

### What offers what

| Kind | Entries | Limit |
|---|---|---|
| Extra die (`bonus_die`) | 14: 11 tag entries, 3 asset entries; each is count 1, d10 | Some once per turn, some none |
| Forced reroll, no choice | Fanatical: every die showing 1 | None |
| Elective reroll, own roll | Book of Secrets `own_roll`: lowest die | Once per turn |
| Elective reroll, rival's roll | Book of Secrets `enemy_roll_at_me`, Psychic Academy `rival_roll` (test fixture only): highest die | Once per turn |

### Six findings that shape the design

1. **Today, a roll reaches the reroll step with exactly one die.** `Stat` rolls one d10.
   The only production offer is `bonus_die`, which always keeps the highest one
   (`bonus_die.go:74`). The other die is thrown away before any reroll hook runs
   (`dice/roll.go:84-87`), and `RollResult` does not keep it. So "the holder picks the
   die" (`A.001.7` D-7) has nothing to pick from. D-1 keeps it so.
2. **The card does not reduce the stops that `A.001.7` D-7 worried about.** That decision
   calls the card the answer to "too many prompts" from Psychic Academy. A declined reroll
   spends no budget, so a holder that keeps declining is asked on every rival roll that
   has a collector, whether the question is a confirm or a card. D-4 settles this: with a
   card on every roll, Psychic Academy adds few stops of its own.
3. **"Rival asks the wrong player" (`B.043`) cannot be seen with one GM.** The TUI sends
   both asks with no faction (`action_collector.go:65`, `:78`), and one GM answers every
   prompt. The prompt does not say whose reroll it is, and that part is real: the card
   fixes it by naming each option's owner. Sending each question to its own faction's
   driver is `F.025`, which waits for `A.003`.
4. **The defense roll does not know the total to beat.** `F.032` asks for it. `Contest`
   rolls the attacker, then the defender (`dice/stat.go:97-98`), and passes the first
   total to no one. The card needs it passed in. This is one place, in `Contest`. Expand
   Influence compares at its own site.
5. **Identical rows toggle together (`B.029`).** One offer comes from each asset copy, so
   two copies give two rows with one label. The card's rows need unique labels, or a
   widget other than huh's multi-select.
6. **Nothing separates "what to say" from "how to draw it"** (layer map). Nine sandbox
   findings sit in presentation code mixed into `tui/`, and none in the driver. A card
   worded inside `overlay/` would add to that pile. D-12 keeps it out.

## Design Summary

- **Before the roll:** a card names the roller, the action, the stat, the world and the
  opponent. It lists each surviving offer with a toggle and its owner, and each struck
  offer with the reason. It always waits for a key.
- **After the roll:** forced rerolls apply first. A card then shows the dice, the modifier,
  the total and, in a contest, the total to beat, with each reroll on offer, its owner and
  the die it hits. The GM takes one reroll at a time, and the card returns with what is
  left until the GM passes. With no reroll on offer, the dice go to the feed and nothing
  waits.
- **The engine:** `hooks.Collector` keeps two methods, and each takes a card value in
  place of today's bare offers or directive. `Contest` passes the first total to the second
  roll. Offer filters report what they strike. The budgets, forced rerolls and the nil
  collector rule stay in `rollWithHooks`.
- **The driver:** the adapter turns each card value into plain names and values, with no IDs
  or pointers. The feed's dice line uses the same wording.
- **The TUI:** two bespoke overlays in the wizard band replace `overlay/hooks.go`, with their
  `askEntries` rows. They only draw. Esc does nothing on them.
- **The data:** Psychic Academy's entry goes in `rulebooks/swn/tags.toml`, in the shape of
  the existing test fixture (`reroll_test.go`): `rival_roll`, one die, elective, once per
  owner faction.

## Decisions

- **D-1 — A bonus die's keep comes before any reroll.** Robert's call. Today's order stays:
  a roll with a bonus die keeps the highest, drops the other, then rerolls. Each card
  shows one kept die, a forced reroll replaces it, and the new die stands. This narrows
  `A.001.7` D-7: with one die there is nothing to pick, so the card has no die picker and
  `RollResult` carries no dropped die for this work.
- **D-2 — Fanatical runs once per roll.** Robert's call. A reroll that lands on 1 stays,
  whether Fanatical or Book of Secrets rolled it.
- **D-3 — Psychic Academy's "once per turn" is once per cycle.** Robert's call. It uses the
  existing budget, which clears in its holder's bookkeeping, so one use covers the rival
  rolls from there to the holder's next bookkeeping. No engine change.
- **D-4 — Every hooked roll with a collector shows the cards, choice or not.** Robert's
  call, for the Foundry feel: the GM sees the dice about to roll. Rejected: cards only on a
  choice, which stops less often but leaves the dice unseen until `F.038`.
  - The rule that `SelectModifiers` is not called when no offer survives goes away.
  - The three nil-collector contests still show no card, until `F.037`.
  - Damage and the other amount, table and target rolls have no hooks, so they show no card.
- **D-5 — The before card always waits; the after card waits only on a choice.** Robert's
  call. With no reroll on offer, the after card's dice go to the feed. A contest takes two
  key presses with no reroll on offer. Each roll with a reroll on offer adds at least one,
  and each further reroll the GM takes adds one more. Rejected: both always wait; neither
  waits on no choice, which gives no pause before the dice.
- **D-6 — Forced rerolls apply first, then the after card loops.** Robert's call. The GM
  picks one reroll and sees the new die, and the card returns with what is left until the
  GM passes. D-2 holds: a forced reroll does not run again after the loop. Rejected: one
  card with toggles, where a second reroll is chosen before the first one's die shows.
- **D-7 — Esc does nothing on a roll card.** Robert's call. Roll and Keep are explicit
  choices, as on the other mid-action prompts, so a stray key cannot throw away a bonus
  die. Rejected: today's Esc as "none".
- **D-8 — This is a feature, `F.039`, rerouted from refactor `R.025`.** Robert's call. It
  adds a prompt on every roll and turns on a tag's effect. Completed docs, the arc plan and
  `E.004` keep the old ID as history.
- **D-9 — The card renders in the wizard band.** Robert's call. The right panel stays free,
  so `F.038` can show earlier matchups beside a card, and `R.016` is not triggered.
  Rejected: the roll's facts in the right panel, which covers that list.
- **D-10 — The before card shows a struck offer and why.** Robert's call. For example,
  "Warlike: struck by Blackmail", with no toggle. Offer filters must now report what they
  strike and why, in place of only removing it.
- **D-11 — Scope.** Robert's call.
  - **`F.032`** — closes here: the card shows the dice, roller, phase, stat and the total to
    beat.
  - **`B.029`** — fixed on the card; other multi-selects stay with the row.
  - **`B.043`** — the card names each option's owner; each faction's own driver stays with
    `F.025`.
  - **`B.027`** — the card shows world names; the other prompts stay with the row.
  - **`B.016`** — budgets spent in an action are restored when that action aborts.
  - **`B.024`** — Book of Secrets gets one shared use per turn, as the row reads the rule.
    This reverses the part of the hooks Key Decision that keeps the two entries' budgets
    apart, so the hooks page changes with it.
  - **`R.015`** — the whole row: all 28 unchecked reply types, the `SelectMovementDecisions`
    nil guard and the `phaseCollector.ask` wrapper. The sweep must not turn a legal empty
    reply into an error.
  - **`F.027`** — in part. One table test over every collector method through `Action()`
    and `Phase()`, with one helper that answers on the channel, and one full cycle through
    the real channels under `-race`, to check `B.038`. A race it finds is logged on
    `B.038`, not fixed here. At close, the row shrinks to the observer and `derive*` tests,
    with the trigger `with E.001`. Rejected: the whole row, because `R.013` retires the
    `derive*` code and the sandbox runs found no driver bugs; the card's asks only, which
    leaves the sweep with no tests.
  - **`F.037`** — out. Its real work is a collector in the reactor routes, and who sees a
    card on another faction's turn.
  - **`R.026`** — out. D-1 keeps "keep highest" and "a forced reroll hits the highest die"
    as they are, so that audit takes them.
- **D-12 — The adapter builds the card; the overlay only draws it.** Robert's call. The card
  reaches the overlay as plain names and values, with no IDs or pointers, as other display
  payloads do. The feed's dice line uses the same wording, so one roll is not worded twice.
  This adds no new layer and leaves the other prompts as they are. Rejected: wording inside
  the overlay, which adds to what `E.001` must untangle, and which `F.038` could not reuse.

## Relation to the Result Panel

`F.038` ([discovery doc](action-result-panel-discovery.md)) is not built, and it follows
this work. Both answer the same gap: the dice never reach the GM. The card shows a roll
while it can still change, and the panel shows the rolls after they are settled.

| Moment in an attack | Roll card | Result panel |
|---|---|---|
| Just before or after a roll | This roll and its choices | Nothing |
| Picking the next defender or a redirect | Nothing | Earlier matchups |
| Action finished | Nothing | Held result, at a pause |

What the panel asks of this work:

- **The panel's job narrows (D-4).** Cards now show each roll as it happens, so the panel
  no longer has to reveal the dice. It still holds the result once the cards are gone, and
  it still shows the earlier matchups during an attack.
- **Where the card renders (D-9).** The panel wants the right panel for earlier matchups
  during an attack, and cards fire in that same stretch. A card in the wizard band leaves
  both on screen.
- **`RollResult` is the shared contract.** It holds only the final dice, so it cannot say
  that a die was rerolled, or by whom. Add what the card needs, and leave the reroll
  history to `F.038`, with no card choice that rules it out.

What the two do to each other:

- **Stops add up.** The panel adds one pause per action. With D-5, a two-attacker attack
  takes five stops with no reroll on offer, and more with them. The wider policy for stops
  belongs to `F.018`.
- **Card values outlast the ask (D-12).** Live pointers are safe only while the engine is
  parked. A card built from plain values is safe to keep, so the panel can reuse its wording.
- **`R.016` fires with the panel.** With D-9, the held result is the fourth detail card. The
  panel's "unified detail framework" looks like the same work as `R.016`.

## Open Questions

For Plan:

1. **How does a wrong reply type fail on a card ask?** `hooks.Collector` returns no error,
   so `R.015`'s loud error has no path there. Returning one would ripple through `Stat`,
   `Contest` and `Gate`.
2. **Who decides that the after card does not wait?** The engine can ask on every roll and
   let the driver send the dice to the feed, or it can tell the feed itself.
3. **What shape does `B.024`'s shared use take?** A change to the budget key, or a new
   `once_per` value in the data.
4. **Where does the `-race` cycle test live?** The adapter cannot import the execution view,
   so a test there only reads the pointers the way the view does. A test in the view package
   is closer, and costs more.

## Related Backlog Rows

Order against this work:

- **`A.001.5`** — D-1 to D-3 keep today's reroll and bonus-die data shapes, so its tag pass
  need not wait, unless `B.024` adds new data vocabulary.
- **`D.010`** — this work rewrites the collector and roll parts of `hooks.md` and `dice.md`,
  so `D.010` follows it.
- **`A.003`, `F.025`** — an AI answers through this collector, so the card values stay plain
  data, with no display logic in them.
- **`E.001`** — D-12 gives it a worked example. Its captured idea lists `F.032` as a known
  hard part, and that note goes stale when this work closes `F.032`.

To watch during the build, not to absorb:

| Row | Watch for |
|---|---|
| `B.006` | With a card on every roll (D-4), more prompts may mean much more flicker; the cause is not known |
| `B.035` | A card with its facts and rows is taller than today's prompts |
| `F.036` | Card buttons share the problem if they use the same theme |
| `B.026` | A "no reroll" choice held as a zero value can fall into the same trap |
| `F.024` | A cousin of D-7; the card need not solve it |
| `F.034` | Two attackers with one name need the world on the card too |
| `F.026` | A cousin of D-10 |
| `R.017` | A new bespoke overlay is a natural first try for `teatest` |
| `B.038` | Live pointers are safe during an ask, not on the event channel |
| `B.012` | Matters if `B.024` changes the budget key |

Not a row: `campaigns/sandbox/TESTING.md:116` describes the Book of Secrets offer, so its
test steps need rewording for the card.

## Out of Scope

- **`F.025`** — each faction's own driver; waits for `A.003`.
- **`F.037`** — the three contests with no collector (D-11).
- **`F.038`** — the held result and the earlier matchups.
- **`R.026`** — the rules in Go that this work leaves as they are (D-11).
- **A general presenter layer** — the other prompts stay as they are; that is `E.001`'s
  question.
- **Rerolls on amount, table and target rolls** — dice Key Decision.

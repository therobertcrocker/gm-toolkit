# Dice Package — Retro

## In Plain Words

- **Dice have one home now.** A new package, `engine/dice`, holds every game roll, and
  each type of roll has a named route: stat rolls with their contests and gates, amount
  rolls, table rolls and target rolls. Turn order is the one plain die left outside it.
- **A new roll has one question left:** which type is it? Eleven sites used to build the
  hooked roll by hand; now `Stat` is its only caller, and the two gates are one `Gate`.
- **An ability's step says which type it rolls,** and its handler reads a typed result.
  Harvesters and Teach Planetary Population are target rolls, as the book words them.
- **The Transport Lockdown's fee and cause moved from Go into its rulebook entry.**
- **Players see no change.** The one bug fix (`B.010`, a reroll used the wrong die's
  sides) could not fire yet, because every die in the data is a d10.
- **Older campaigns need a re-seed.** The loader now refuses the old lockdown and Teach
  data. The sandbox is done.
- **Next:** the roll card (`R.025`) changes one place, and H.A.W.K's authoring passes
  (`A.001.5`) use the new data shapes.

## 1 — Did we do what we set out to do?

Yes. `R.028` has no discovery doc by design, so the yardstick is `E.004`'s
recommendation (C-5, P-4, O-1) and the plan's In Plain Words. All of it landed.

- **Landed as designed:** all seven route types, `Roller` and `RandRoller` in the package,
  `DiceRoll.Roll` gone from `domain`, `rollWithHooks` private with `Stat` as its one
  caller, both gates on `Gate`, the three tie readings on `Outcome`, the three-way step
  split, one `domain.RollPhase`, typed stats on `RollContext`, the shared fakes in
  `dicetest`, and the new `dice.md` page. An LSP reference list shows `Roller.Roll`
  called only inside `dice` and at `turn/order.go:20`.
- **Absorbed, as planned:** `B.010` (Commit 7) and the `lock_out` case from `R.026`
  (Commit 15).
- **Cut or deferred:** nothing. Amount rolls through hooks, dropped dice and collectors
  on effect gates stayed out of scope, as the plan said.
- **Added beyond the plan:** the adapter's `statRating` went too, on Robert's pick, as a
  third copy of the same switch; `Sequence` and `Cycle` gained `Rolled()`, to keep seven
  "rolled exactly this many dice" checks; `dice.md` gained a list of the roll orders that
  must hold; an attack's tie context now carries `Action`, which no reader sees.
- **A correction along the way:** Decision 1 found that three entries (Fanatical, Book of
  Secrets, Psychic Academy) could reach amount dice on their plain words, against
  `E.004`'s "Rules calls to carry: None". Robert ruled that rerolls reach stat rolls only.

## 2 — Is there a better way we didn't see?

The chosen shape still holds. C-3 would have left amount rolls with no named route, and
C-4 would have built hooks no rule asks for. Two smaller things would have gone better.

- **Name the anchor, not the line.** Commits 9, 12 and 13 each logged task line numbers
  that had gone stale, because an earlier commit moved the code down. Naming the symbol
  ("the matchup in `Attack`") would have saved those corrections, at no cost to the plan.
  The larger call held: literal code for what is new and a site list for each conversion
  (Decision 12) gave 1,559 plan lines for 642 inserted lines of production Go, a much
  smaller plan for its code than `A.001.7`'s.
- **Read the whole rulebook in the exploration.** `E.004` searched the text for reroll
  rules and missed three entries; the Plan session's full read caught them. A full read
  in `E.004` would have settled Decision 1 a session sooner, and the exploration would
  not need correcting from outside. Small cost, small saving, no rework either way.

One shape I looked at and would keep: the three effect-gate callers each build their own
`Engine`, which looks like repetition. One shared helper would hide the nil collector and
its fence-sign, and `F.037` needs to swap that one argument per caller.

## 3 — Does it fit the project?

Yes, and the architecture pages moved with the code, which `A.001.7` did not manage.

- **Matches the data-driven rule:** the lockdown's fee reads from its entry, and its cause
  takes the effect's name, as every other effect's does.
- **Matches hooks:** the nil-collector rule holds at each effect gate, and `dispatch` keeps
  Cat 3, Cat 4 and Cat 6. The roll sections and three Key Decisions moved to `dice.md`.
- **Matches persistence:** the strict loader takes the new keys and refuses the old shapes;
  saves and rulebooks move by a re-seed, with no migration code.
- **Matches the error rule:** `lockdownFee` returns an error and the caller forwards it.
- **Fence-signs:** every one in the plan's table moved or stayed as listed, and the
  reread of `Contest`, `passesGate` and `contestAction` shows them in place.
- **Exceptions, each named on `dice.md` or at a fence-sign:** Expand Influence stays a plain
  comparison at its site, with no resolver. `passesGate` keeps a literal action name,
  because `actions` imports `ability`. Effect contests keep an empty `Action`, so the
  Imperialists' `action = "Attack"` filter does not reach them.
- **Process drift:** the app run used a pty, not tmux, because this container has no tmux.
  Phases 7, 8 and 9 ran in one session, on Robert's call. Both are logged.

## 4 — What does it change about the project?

- **New pattern to copy:** a new roll calls a route or extends one. It cannot build its own
  hooked roll, because that function is private.
- **`R.025` changes one place,** and `Outcome` already keeps both rolls for the card.
  `F.032` and `F.037` get one home. `A.001.5` writes `target_roll` from its first pass.
- **`Stat` is always `1d10 + stat`.** `statDieSides` is a rule value held in Go. Every
  rulebook today uses it, so nothing breaks, but a rulebook with another stat die would have
  to extend the route. `R.026`'s audit is where it surfaces.
- **GM data got stricter.** A `lock_out` with no block, or a `raise_tech_level` with rows,
  no longer loads. Any campaign other than the sandbox needs
  `campaign add-rules --replace` before it runs on this code.
- **Old history keeps old text.** History files still say `transport_lockdown` as a fee's
  cause. Nothing reads it, and nothing shows a fee's cause.
- **Docs still out of date:** the other stale parts of `hooks.md` stay with `D.010`. `E.004`
  keeps its "None" for rules calls, as written; the correction lives in the plan and on
  `dice.md`. The `hooks.md` diagram changed two node labels and has not been rendered.

## 5 — Rating

| Metric | Score | Evidence |
|---|---|---|
| Delivered | 5 | Every part of C-5 landed, with both absorbed rows and no player-visible change |
| Plan held | 4 | None of the 15 ratified decisions reversed; 37 Decision Record entries, mostly small corrections to task lists and line numbers |
| Rework | | |
| Fit | 5 | Follows the data-driven rule and the loader's pattern; pages updated on the branch; each exception is named |
| Simplicity | 4 | One small package with named routes, and `contest.go` shrinks to one constant; `StepResult` carries three fields where only one applies per step |
| **Total** | | |

**Verdict:** A clean refactor that did what it set out to do, on a plan that held; the one
real catch, the rerolls reading, came from reading the whole rulebook in the Plan session.

**What surprised me:** the audit's confident "no rules calls" was the weakest claim in it,
and a text search was the reason.

## Follow-ups

| ID | Item | Trigger | Detail |
|---|---|---|---|
| `R.029` | Tag contest check is dead code | -- | Every effect a tag may carry now needs a die, and the loader refuses a contest with a die first, so `step.Contest` in `checkTagAbility` cannot fire; delete it, or keep it to hold the rule if a tag ever gets an effect with no die (`R.028` Decision Record, Commit 13) |
| `B.046` | Rows beside a target roll are dropped | -- | The loader reads `outcomes` only when a step has a `die`, so a step with `target_roll` and stray rows loads and ignores the rows; no shipped data does this; refuse it, as a die with a target roll is refused (`R.028` Decision Record, Commit 14) |

# Data-Driven Goals — Retro

## In Plain Words

- **Goals run from rulebook data now.** Each of the eleven rules goals in
  `rulebooks/_core/goals.toml` states its behavior in six parts: what the pick
  needs, how progress moves, what resets it, the threshold, what completes it, and
  the XP it pays. A goal engine checks every name when the engine starts. The twelve
  Go goal handlers are gone, and no Go code names a goal ID.
- **Goals see everything their faction causes,** on any faction's turn and in any
  phase. A kill made while defending counts, and any cause counts: an attack, a
  counterattack, an ability or an effect. A missed Attack now breaks Peaceable
  Kingdom, because each action opens its batch with an `ActionTaken` record.
- **A faction with no goal picks one at the start of its turn,** after its stat
  raise. The pick offers only the goals it may take now, asks for a target faction
  for Destroy the Foe, and may be declined. The faction wizard no longer sets a goal.
- **Destroy the Foe works:** it gets a target, completes on the faction's own blow,
  and fails with a new `GoalFailed` record when the target falls any other way.
- **Rules readings settled:** every XP formula rounds up; Inside Enemy Territory
  counts the stealthed assets held, not stealth events; Wealth of Worlds comes back
  only after Wealth rises.
- **Removed until `F.040`:** the Seize Planet and Change Homeworld actions, the
  goal-lock phase and the goal process around them. World Seizure stays in the data
  and leaves the pick until Seize Planet returns. Goal IDs are now `CORE-G-001` to
  `CORE-G-011`.

## 1 — Did we do what we set out to do?

Yes. The discovery's In Plain Words promised goals as data in six parts, goals that
read every batch, and a turn-start pick in the engine and the TUI. All three landed,
in the three efforts D-27 set.

- **Landed as designed:** the six parts and the goal-only vocabulary (D-18); strict
  loading and checks at construction; `ActionTaken` (D-19, `B.013`); every batch for
  every faction (D-13, D-14); `GoalFailed` and the lost-target rule (D-21); the pick
  after stat-raise with its `PhaseCollector` ask and overlay (D-2 to D-4, D-6); start
  inputs (D-3, `B.040`); both gates (D-11, D-12); `G-009`'s start record (D-10); the
  wizard's goal step gone (D-25, `B.041`); `CORE-G-` IDs and the re-seed (D-26);
  `difficulty` gone (`B.005`); the `G-004` and `G-012` literals gone (`B.001`). The
  rules doc took each reading in the effort that made it true (overview Decision 5).
- **Widened in Plan, on purpose:** "unregistered" became "deleted" for the two
  actions and the goal process (overview Decision 3), the goal-lock phase went with
  them (Decision 2), and Imperialists' entry left `tags.toml` (Decision 4), because
  the discovery's narrower cuts did not compile.
- **Corrected in Plan:** D-20's mechanism. The source emits `BaseHPDelta` for every
  Base hit, so the `damage` event matches no `FactionHPDelta` (Effort 1 Decision 1).
- **Added beyond the plans:** an abandoning batch moves the goal no further (Effort 1
  Decision Record 11); a destroyed faction's goal counts nothing and a dead or
  missing target fails the goal (Effort 2 Decision Records 4 to 6); a dead faction
  is not asked to pick, and a resumed turn does not ask twice (Effort 3 Decision
  Record 7); "do nothing" options lead three lists (Effort 3 Decision Record 6); a
  new `GoalPicked` mutation, since Effort 1 had deleted the only other goal writer
  (Effort 3 Decision 1).
- **Shipped with known gaps:** the goal pick cuts long descriptions at the panel
  edge (Effort 3 Decision Record 5), and a GM who picks and then quits before the
  turn's first save is asked again on resume (Effort 3 Decision Record 7).

## 2 — Is there a better way we didn't see?

The design holds. Two things would have saved time.

- **The design.** A table of events, filters, measures and formulas, compiled once
  and matched per batch, is the plain shape for "goal behavior as data". `E.002`'s
  alternatives (one `shape` field, Go kernels for the process goals) would have kept
  Go per goal. Running goals inside `applyAndRecord` (Effort 2 Decision 1) made
  "every batch" hold by construction, and Effort 3's pick batch got goal reading for
  free. Nothing on the branch argues for going back.
- **The discovery read history, not source, in two places.** D-20 took its
  mechanism from frozen log D75, and the code had moved on. D-24 removed
  `TargetWorld` without seeing that Imperialists' roll filter reads it. The overview
  and Effort 1 Plan sessions caught both before any code, so the cost was a Plan
  decision each. A source grep for each field a discovery removes, and for each
  emitter it names, would have caught them a session sooner.
- **The plans.** About 6,400 lines of plan for about 1,800 lines of production Go,
  a third of `A.001.7`'s ratio. Effort 2 and Effort 3 report no anchor drift at all,
  which says the just-in-time effort plans (overview Decision 1) paid off. Effort 1,
  at 3,200 lines, was still the largest single doc; most of it is the removal
  listings, which the compiler would have found anyway.
- **The lost-target rule took two passes.** Effort 2 planned "fails on a third
  party's blow", and its review had to settle a dead-on-arrival target and the
  holder's own early blow. Robert's principle ("lose the conditions to complete,
  the goal fails") was the simpler rule from the start; stating it in Discovery in
  place of D-15's case list would have saved the back-and-forth (Effort 2 Decision
  Record 6).

## 3 — Does it fit the project?

Yes. It follows the arc's data-driven rule and the project's mutation model. The
architecture pages are stale until this close, as planned.

- **Matches effect-mutation:** goals emit, never mutate. `GoalPicked` is the one
  writer of `ActiveGoal`, `GoalCompleted` writes the completion record, and
  `ActionTaken` is record-only, like `GoalInitiated` was.
- **Matches the data-driven rule** (Robert's standing call): no Go names a goal ID;
  thresholds, amounts and XP numbers live in the data; a bad name fails at engine
  construction, not silently in play. The goals page's "silent skip" seam is closed,
  as the page itself promised.
- **Matches persistence:** strict loader, references checked at load, prefixed IDs,
  saves moved by re-seed with no migration code (the `A.001.7` precedent).
- **Matches the arc:** ACD-2 (behavior from data), ACD-6 (goals stay in `_core`),
  AP-8 and AP-9 (prefixed IDs, tag references renamed). The arc-plan's "phases" and
  "tick" parts were narrowed out by the discovery and recorded in the arc-plan.
- **Diverges, on purpose:** the goals page's "two advancement paths, by beat" is
  now one beat on every batch; the orchestrator's "single choke point for state
  change" now reads goals before it applies; the goal-lock phase and its checkpoint
  are gone. D-18 splits goal events from effect triggers, so `kill` (any cause) and
  effects' `asset_destroyed` (combat only) mean different things by design.
- **Rides on the branch, not goals work:** `docs/initiatives/seeds/Codex_Assistant.md`
  (a GM-assistant concept seed) landed in the Effort 1 squash through `git add -A`,
  and `R.029` (TOML 1.1) ran and closed on this branch before Effort 1 (its own
  retro). Both merge with this branch.

## 4 — What does it change about the project?

- **A rulebook adds a goal with no Go,** as long as the existing events, filters,
  measures and formulas can say it. A new kind of event is still a Go entry in the
  goal package's tables (`events.go`, `measures.go`, `xp.go`). `docs/contributing/engine.md`
  has the recipe.
- **Every batch is goal-read.** A new batch path through `applyAndRecord` gets goal
  progress and its reactor pass for free, and its caller must use the batch it
  returns, which holds the goal mutations, when it tells the observer.
- **A goal can complete on a rival's turn,** so anything that reacts to a
  completion (Exchange Consulate's XP roll) can fire in any phase.
- **Each new turn phase must guard against a resumed turn.** The pick needed the
  same `BookkeepingApplied` gate the turn-start effects use; the next phase that
  asks a question will need it too.
- **Seize Planet and Change Homeworld are gone, not hidden.** `F.040` rebuilds them
  from git history, with their asks, overlays and process mutations, and puts
  Imperialists' entry back with a seizure filter. Until then World Seizure is never
  offered, Imperialists does nothing, and Deep Rooted's homeworld loss has no
  trigger in play.
- **`F.033` loses its subject:** `GoalInitiated` and `GoalPhaseAdvanced` no longer
  exist.
- **Robert's `sandbox` needs a re-compose and re-seed:** its goals copy stopped
  loading at Effort 1 (`difficulty`), and its seeds walk through Planetary Seizure.
- **Docs out of date until this close:** the goals page body (handlers, locks, the
  Planetary Seizure example, twelve goals), the orchestrator's phase list and
  checkpoints, effect-mutation (`ActionTaken`, `GoalFailed`, `GoalPicked`), the
  overlays page (`GoalPick`), the actions page (the two actions), and
  `docs/contributing/overview.md` (the goal handler registry).

## 5 — Rating

| Metric | Score | Evidence |
|---|---|---|
| Delivered | 4 | All three efforts landed and six bugs closed; the pick cuts goal descriptions at the panel edge, which blunts the overlay's main aid |
| Plan held | 4 | No overview reversal; three Plan-time widenings of discovery decisions, all made before code; Effort 2's lost-target rule reshaped at review |
| Rework | | |
| Fit | 4 | Data-driven, emit-never-mutate and strict loading all hold; two Key Decisions reversed on purpose, stale on the pages until this close; an unrelated seed doc rode in |
| Simplicity | 4 | About 900 lines of table-driven goal engine replace twelve handlers, a lock package and a goal process; the stealth measure clones the whole state per holder per batch |
| **Total** | | |

**Verdict:**

**What surprised me:** the discovery's two misses (D-20's mechanism, D-24 against
Imperialists) both came from reading history in place of source, and both were
caught by the next Plan session reading the code.

## Follow-ups

| ID | Item | Trigger | Detail |
|---|---|---|---|
| `B.048` | Goal pick cuts long descriptions | -- | huh renders the highlighted goal's description on one line and no overlay sets a form width, so most of a goal's text is lost at the panel edge (`A.001.8` Effort 3 Decision Record 5). Wrap it, or size the form to the panel |
| `B.049` | A pick before the first save is asked again | when a resume loses a pick | A GM who picks a goal and then quits at a later prompt before the turn's first save leaves `goal_picked` in history that the saved state lacks; on resume the pick is asked again. The stat raise has the same shape (`A.001.8` Effort 3 Decision Record 7) |
| `C.001` | Live phase name in the logging tests | -- | `internal/logging/handler_test.go` and `banner_test.go` use `goal_lock` as a sample two-word phase name; that phase is gone. Swap in a live one, such as `goal_pick` (`A.001.8` Effort 1 Decision Record 4) |
| `C.002` | TomlSmith pass on `goals.toml` | -- | `goals.toml` was laid out by hand; a Format Document pass in the editor should change nothing, and if it does, keep its layout (`A.001.8` Effort 1 Decision Record 6) |
| `F.040` (amend) | One name for the seizure goal | with `F.040` | The rules doc says "Planetary Seizure", `goals.toml` says "World Seizure" (`A.001.8` Effort 1 Decision Record 10); pick one when Seize Planet is rebuilt |

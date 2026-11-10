# Data-Driven Goals (A.001.8) — Implementation Plan (Overview)

## In Plain Words

- **What:** the overview plan for making the eleven rules goals work from rulebook
  data in place of twelve Go handlers, for letting every goal see every batch of
  changes, and for adding a goal pick at the start of a turn. It records the
  decisions that all three efforts share, fixes their order, and says that each
  effort gets its own Plan session. It holds no code spec.
- **Why:** a goal with no Go handler does nothing, and the goal flow is broken in
  play: no step picks a new goal, Destroy the Foe never gets a target, and Peaceable
  Kingdom misses a missed attack. The discovery designed the fix. This session
  re-read the code and settled what the discovery left open or did not see: what
  happens to the empty goal-lock phase, to the two goal-bound actions, and to the
  Imperialists tag, which reads a field the discovery removes.
- **How it fits:** the second mode of a multi-effort feature, on branch
  `feature/data-driven-goals`. Initiative `A.001.8` of arc `A.001`, the twin of
  `A.001.7` (data-driven tags). `F.040` (multi-turn actions) waits on Effort 1.

## Context / Goal

Build the three efforts the [discovery](../discovery/A.001.8-data-driven-goals-discovery.md)
designed (D-27): goals as data (Effort 1), every batch (Effort 2) and the turn-start
pick (Effort 3).

- Arc: [data-driven-rulebook](../arcs/A.001-data-driven-rulebook/A.001-data-driven-rulebook-arc-plan.md),
  initiative A.001.8 (AP-8, AP-9).
- Input: [`E.002` exploration](../explorations/E.002-data-driven-tags-goals-exploration.md),
  options G-1 (goals) and I-1 (IDs).
- Discovery decisions D-1 to D-27 stand, with one amendment made before this
  session: goal parts are TOML 1.1 inline tables in the rulebook house style, one
  header per goal, not the sub-tables of D-17 ([`R.029` exploration](../explorations/R.029-toml-1-1-migration-exploration.md)
  G-1, Robert's call 2026-11-06). `R.029` has landed on this branch, so the house
  style in `docs/contributing/engine.md` is live. The discovery's TOML block shows
  sub-tables; Effort 1 writes the same keys in the house style.
- The decisions below settle what the code showed in this session. Decisions 3 and 4
  widen D-1 and D-24 to what compiles; none of them reverses a D-decision.
- Rules source: `docs/rules/swn-faction-mechanics.md` and the book, not the
  `description` text in `goals.toml`.

## Decisions Ratified in Planning

1. **Every effort gets its own Plan session; this file carries no effort spec. DAG
   `1 → 2 → 3`, run in that order.** Robert's call. Effort 1 is too large to spec
   beside the shared decisions: the vocabulary, the loader and its checks, a matcher
   engine, eleven goals, `ActionTaken`, about 1,500 lines of handlers and tests
   removed, two actions removed, `ActiveGoal` slimmed, the IDs and the re-seed.
   Efforts 2 and 3 write literal code against the goal engine that Effort 1 creates.
   Each effort plan is written just before its effort runs, against code that exists.
   The order is D-27's. Same shape as the A.001.7 overview, Decision 1. Rejected:
   Efforts 2 and 3 specified here (their tasks would quote code that does not exist
   yet); Effort 1 specified here (too large).

2. **Effort 1 deletes the goal-lock phase.** Robert's call. `CheckLock` does work
   only for `G-004` and `G-012`. With those handlers gone (D-1, D-5) and `tick`
   replaced by `action_taken` (D-19), the phase has nothing to run. Effort 1 removes
   it with its last user:
   - `runGoalLockPhase` and the `LockSkip` short-circuit in `RunFactionTurn`
   - the `locks` package, `filterAllowedActions` and the `lock` parameter of
     `runActionPhase`
   - `CheckpointGoalLocked`, `TurnObserver.OnGoalLockApplied` and every
     implementation of it, and the adapter's `EvtGoalLockApplied` with its payload
     and feed line

   `F.040` gives actions their own in-progress state, the lock included, so it
   builds its lock on the action. Rejected: deleting it in Effort 2 (Effort 1 would
   leave a phase that calls nothing); keeping an empty phase for `F.040` (its lock
   lives on the action, so the phase may not fit it).

3. **"Unregistered" means removed. Effort 1 deletes Seize Planet, Change Homeworld
   and the goal process around them.** Robert's call. Widens D-1. Removing the two
   lines from `register.go` does not compile: `seize_planet.go` reads
   `ActiveGoal.ProcessPhase`, which D-24 removes, and both actions emit
   `GoalInitiated`. Effort 1 deletes:
   - both actions and their tests
   - `action.Collector.SelectSeizeTarget` and `SelectChangeHomeworldTarget`, their
     adapter asks and payloads, their `askEntries` rows and their smoke-collector
     stubs
   - `GoalInitiated`, `GoalTurnsTick` and `GoalPhaseAdvanced`, with their apply arms
     and their digest and feed cases (each exists only for the process, and its
     apply arm writes a field D-24 removes)
   - the scenario tests of the process (goal lock, lock skip, the Change Homeworld
     lifecycle)

   Kept: `HomeworldChanged`, which Deep Rooted's `remove_tag` reacts to, and
   `WorldGovernmentChanged`, which gains the resist field that `ruler_average`
   reads (D-8). `F.040` rebuilds the actions on action state and can take what it
   needs from git history. Rejected: keeping the files unregistered and cut back
   until they compile (dead code with tests); keeping only the target selection that
   `F.040` would reuse.

4. **Imperialists' `effects` entry leaves `tags.toml` until `F.040`, and `goal`
   leaves `bonus_die`'s filters.** Robert's call. The discovery missed this. The roll
   filter `goal` (`onGoalTarget`) matches when the holder's active goal is the named
   goal and the roll's world is that goal's `TargetWorld`. D-24 removes
   `TargetWorld`, so the discovery's "only rename the reference" does not compile.
   With the entry gone, `onGoalTarget` and the `goal` filter on `bonus_die` have no
   user, so Effort 1 deletes them. The tag text stays, and the effect waits for its
   own row, as Psychic Academy's entry waited for `R.025`. `F.040` puts the entry
   back with a seizure-in-progress filter. Exchange Consulate's `xp_roll` keeps its
   `goal` filter, which checks the completed goal's ID and no world, and its
   reference becomes `CORE-G-007`. Rejected: keeping `TargetWorld`, never written,
   until `F.040` (a narrow amendment of D-24); a goal filter that checks the ID only
   (Imperialists would add its die to every attack while World Seizure is held, and
   the wizard can set that goal until Effort 3).

5. **Each effort edits `docs/rules/swn-faction-mechanics.md` for the readings it
   lands.** Robert's call. Places the edit the discovery left to Plan, so the doc
   never runs ahead of the code:
   - Effort 1: D-8, D-9, D-10 (assets held are counted), D-14, D-15 (the faction's
     own blow), D-16, D-20 and D-22
   - Effort 2: D-13 and D-21 (D-13 is a rules call the discovery's list left out)
   - Effort 3: D-10 (assets stealthed at the start never count) and D-11

   Rejected: one docs commit in Effort 3; the edit at initiative close.

6. **The initiative close rewrites the goals architecture page body, not only its
   Key Decisions.** Robert's call. After this initiative most of the body is wrong:
   the `Handler` contract, the lock types, the Planetary Seizure worked example and
   a catalogue of twelve goals. The subsystem's model changes as a whole, so a
   separate docs row would only defer the work. Rejected: graduating Key Decisions
   only and filing the body as a `D.` row, as `A.001.7` did (`D.010`).

7. **`needs_actions` names actions by their display name (`"Seize Planet"`).**
   Robert's call. Resolves a conflict in the discovery: D-12's prose says
   `seize_planet`, while the vocabulary table and the TOML block say
   `"Seize Planet"`. Actions are known by `Name()` everywhere
   (`ActionEngine.AvailableActions`, the old lock filter), and the `[[effects]]`
   `action` filter uses the same words. Rejected: a snake-case key, which needs a
   second action identifier that no other reader uses.

## Decision Record — Execution

None yet. Each effort plan carries its own Decision Record; this one holds only
reversals of the decisions above.

---

## Effort Summary

| Done | # | Effort | Scope | File |
|------|---|--------|-------|------|
| [x] | 1 | Goals as data | The goal vocabulary and its loader, `CORE-G-` IDs, the eleven goals as data, `ActionTaken`, the handlers, the goal-lock phase and the goal process deleted, `ActiveGoal` slimmed, the re-seed | [`A.001.8.E1-data-driven-goals-plan.md`](./A.001.8.E1-data-driven-goals-plan.md) |
| [ ] | 2 | Every batch | Goals read every batch in every phase for every faction; `GoalFailed` and the lost-target rule | `A.001.8.E2-data-driven-goals-plan.md` (planned) |
| [ ] | 3 | Turn-start pick | The pick step, its ask and start inputs, both gates and their records, the TUI overlay, the wizard's goal step removed | `A.001.8.E3-data-driven-goals-plan.md` (planned) |

## Shared Context

### Branch and git cadence

- Branch `feature/data-driven-goals`, cut when Discovery started. Every effort
  session runs on it, Plan sessions included.
- Every commit for effort n carries the scope `E<n>`, its plan doc included
  (`docs(E1): ...`).
- At each effort boundary, `/gestalt:effort-close A.001.8.E<n>` runs the interim
  review, one `fix(E<n>):` commit, and the squash. Merge to `main` with `--no-ff`.
- Claude runs the app only with `--campaign sandbox-claude`, and re-seeds it in
  Effort 1. Robert re-seeds `sandbox` (D-26).

### Findings for the effort plans

Verified against the source in this session (2026-11-10). Each effort plan re-reads
the code before it relies on them.

- **Goal engine.** `goal.New(log)` (`engine/goal/goal.go:27`) registers twelve
  handlers and returns no error; `NewWithRulebook` (`engine/core.go:51`) calls it at
  `:56` and already returns an error for bad effect data, so it can host the goal
  checks. (Effort 1)
- **Goal loader.** `loadGoals` (`rulebook/rulebook.go:783`) is strict through
  `refuseUnknownKeys` and refuses a duplicate ID. `goalRecord.Difficulty` has one
  reader, the conversion at `:801`. `checkReferences` (`:811`) refuses an `effects`
  entry's `goal` filter that names no goal, so the tag references must change in
  the same commit as the IDs. (Effort 1)
- **Tag references.** `rulebooks/swn/tags.toml:68` (Exchange Consulate, `G-007`,
  `xp_roll`) and `:97` (Imperialists, `G-004`, `bonus_die`, Decision 4). (Effort 1)
- **`ActiveGoal` readers outside the handlers.** `TargetWorld`: `onGoalTarget`
  (`effect/effects/roll_match.go:95`) and the `GoalInitiated` arm
  (`engine/mutation/mutation.go:310`). `ProcessPhase`: `SeizePlanet.Validate`
  (`actions/seize_planet.go:29`) and the arms at `mutation.go:309` and `:334`.
  Scenario tests set them in `capture_test.go`, `goal_lock_test.go`,
  `movement_lockskip_test.go` and `change_homeworld_lifecycle_test.go`. (Effort 1)
- **Goal-lock reach (Decision 2).** `runGoalLockPhase` (`engine/orchestrator.go:168`),
  the `LockSkip` branch (`:125`), `filterAllowedActions` (`:705`),
  `CheckpointGoalLocked` (`:30`); `OnGoalLockApplied` in `engine/observer.go`,
  `tui/adapter/observer.go`, `testharness/observer.go` and `orchestrator_test.go`;
  `EvtGoalLockApplied` and `GoalLockAppliedPayload` in `tui/adapter/channels.go`.
  (Effort 1)
- **Process reach (Decision 3).** The two collector methods sit at
  `engine/action/collector.go:35–36`, the adapter asks at
  `tui/adapter/action_collector.go:456` and `:493`, the payloads at
  `tui/adapter/channels.go:142` and `:146`, the overlay rows at
  `tui/views/turn/execution/execution.go:301` and `:305`, and the smoke stubs at
  `tui/adapter/smoke.go:148` and `:151`. `HomeworldChanged` is read by
  `effects/remove_tag.go:49`, the digest and the feed, so it stays. (Effort 1)
- **The one goal call today.** `UpdateProgress` runs only in `runActionPhase`, for
  the acting faction (`orchestrator.go:560`), after `MutationReactors` (`:548`) and
  before a second reactor pass over the goal output alone (`:563`). A pass returns
  at `:521`, before `Action.Run`, so a pass emits no `ActionTaken` (D-16, D-19).
  (Efforts 1 and 2)
- **Every batch (D-13).** `applyAndRecord` (`orchestrator.go:660`) is the funnel
  that every batch passes through, but `MutationReactors` runs at each call site,
  not inside it, and the stat-raise, condition-lift, turn-start and bookkeeping
  batches skip the reactors. The batches: stat raise (`:349`), condition lifts (`:305`), turn-start
  effects (`:249`), bookkeeping (`:274`), movement tick (`:416`), movement
  decisions (`:452`), granted movement (`:932`) and the action (`:577`). Goal
  progress must run after the reactors and before the apply. (Effort 2)
- **Goal events and reactors.** Exchange Consulate's `xp_roll` reacts to a
  `GoalCompleted` whose `FactionID` is its holder (`effects/xp_roll.go:51`). Under
  D-13 a completion can come in any batch, on a rival's turn, so the reactor pass
  over goal output must follow the goal call in every batch. (Effort 2)
- **The pick's model.** `prepareStatRaise` (`orchestrator.go:758`) asks
  `PhaseCollector.SelectStatRaise`, skips the ask when nothing is eligible, and
  reads nil as a decline (D-4). Its overlay is `tui/views/turn/overlay/statraise.go`;
  a new ask needs its `askEntries` row (overlays page, Key Decisions). (Effort 3)
- **Wizard.** The goal group is `tui/views/manage/wizard/wizard.go:255`, with
  `formData.selectedGoal` (`:34`) and `goalsToOptions` (`:441`). (Effort 3)

### Gaps between efforts

Known and accepted. Each closes where named.

- **Until Effort 2,** goals read only the acting faction's action batch, as today.
- **Until Effort 3,** the wizard still sets goals: `CORE-G-008` has no target
  (`B.040`, as today), a wizard-set `CORE-G-004` can never complete (D-12's gate
  is not there yet), and `CORE-G-009` has no start record, so assets already
  stealthed count. `CORE-G-011` has no gate.
- **Until `F.040`,** nothing in play writes `WorldGovernmentChanged`, so
  `ruler_average` is tested with built batches only, and Imperialists has no
  effect (Decision 4).

### Pre-merge additions

- **Graduate to Key Decisions:**
  - goals page: one beat on every batch in place of two advancement paths;
    behavior from data, checked at construction, in place of the silent skip;
    goal-only events and filters (D-18); "no Go names a goal ID"
  - orchestrator page: the goal pick after stat-raise (D-2); goal progress on every
    batch (D-13); the goal-lock phase and its checkpoint removed (Decision 2), which
    retires "Goal-lock sits adjacent to the phase it gates" and changes the
    checkpoint list
  - effect-mutation page: `ActionTaken` as a record-only mutation (D-19);
    `GoalFailed` (D-21)
  - persistence page: prefixed goal IDs and the re-seed (D-26); the government
    record decision names Planetary Seizure as a writer, which `F.040` takes over
- **Goals page body:** rewrite it against the source (Decision 6).
- **Other pages:** check the actions, overlays and orchestrator page bodies for
  Seize Planet, Change Homeworld and the goal lock.
- **`planned-work.md`:**
  - close `B.001`, `B.005`, `B.013`, `B.039`, `B.040` and `B.041`
  - narrow `R.004` to its ability half
  - `F.040`: rebuild the two actions, their asks and overlays, and the process
    mutations it needs (Decision 3); restore Imperialists' entry with its new
    filter (Decision 4)
  - `F.033`: its two mutations are gone (Decision 3); re-scope it to `F.040`'s
    events, or close it
  - `R.013`: `deriveContestedWorlds` and `seizePlanetTargetWorlds` go with Seize
    Planet; re-word it
  - remove `A.001.8`

## Out of Scope

- **Multi-turn actions** — `F.040`: Seize Planet and Change Homeworld as actions,
  the seizure's completion, the resist record (D-8) and Imperialists' entry
  (Decision 4).
- **Goal state predicates** — `F.013`: a `complete` predicate kind plugs in beside
  `threshold` and `event`.
- **Scavengers' combat-only reading.** Effects' `asset_destroyed` keeps matching
  attack and counterstrike losses only (D-18).
- **H.A.W.K goals.** `_core` goals serve both rulebooks; `A.001.5` adds none here.

---

## Work Breakdown

### Effort 1 — Goals as data

**Scope:** discovery D-27, Effort 1, widened by Decisions 2 to 4.
- The goal vocabulary from the discovery's Goal Vocabulary: six parts, seven
  events, the filters, the `stealthed_assets` measure, four XP formulas and two
  gates, read by a goal engine that matches each goal's parts against a batch.
  The parts are inline tables in the house style. Every name is checked at engine
  construction.
- The goal loader reads the parts strictly and drops `difficulty` (`B.005`).
- Goal IDs `CORE-G-001` to `CORE-G-011`; `G-012` leaves `goals.toml` (D-5).
  Exchange Consulate's reference follows (D-26); Imperialists' entry leaves
  (Decision 4).
- `ActionTaken{FactionID, Action}` opens each action batch, with a no-op apply arm,
  a digest line and a feed line (D-19, `B.013`).
- The twelve handlers, the `Handler` registry and its silent skip are deleted; no Go
  names a goal ID (`B.001`).
- The goal-lock phase is deleted (Decision 2), as are Seize Planet, Change Homeworld
  and the goal process (Decision 3).
- `ActiveGoal` loses `TargetWorld`, `ProcessPhase` and `TurnsRemaining` (D-24) and
  gains the asset IDs held stealthed at start, which the measure reads and Effort 3
  fills. `WorldGovernmentChanged` gains the resist field (D-8).
- Goals still read only the acting faction's action batch.
- The rules doc for Effort 1's readings (Decision 5), and the `sandbox-claude`
  re-seed (D-26).

**Depends on:** nothing in this initiative.

**Own Plan session:** yes (Decision 1). Expected size: several phases. The
vocabulary and loader, the engine swap, and the removals are separate seams.

### Effort 2 — Every batch

**Scope:** discovery D-27, Effort 2.
- Every faction's goal reads every batch, in every phase and on every turn, after
  the reactors and before the apply, and its output gets its own reactor pass
  (D-13).
- `GoalFailed{FactionID, GoalID, Cause}` with its apply arm, digest line and feed
  line, and the built-in lost-target rule (D-21).
- The rules doc for D-13 and D-21 (Decision 5).

**Depends on:** Effort 1.

**Own Plan session:** yes (Decision 1).

### Effort 3 — Turn-start pick

**Scope:** discovery D-27, Effort 3.
- A goal-pick step after stat-raise and before bookkeeping (D-2), skipped for a
  faction that holds a goal.
- A `PhaseCollector` ask that offers the goals whose gates are open and asks only
  the start inputs each goal declares (D-3). A nil answer declines (D-4).
- Both gates: `needs_actions` by display name (D-12, Decision 7), and
  `rating_above_last_completion` with `Faction`'s record of its ratings, written
  when `GoalCompleted` applies (D-11).
- `G-009`'s start record, filled at the pick (D-10).
- The TUI overlay and its `askEntries` row (D-6).
- The wizard's goal step and `goalsToOptions` are removed (D-25).
- The rules doc for Effort 3's readings (Decision 5).
- Closes `B.039`, `B.040` and `B.041`.

**Depends on:** Effort 2.

**Own Plan session:** yes (Decision 1).

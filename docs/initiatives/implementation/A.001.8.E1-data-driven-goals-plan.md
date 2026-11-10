# Data-Driven Goals (A.001.8) — Effort 1 Plan: Goals as Data

## In Plain Words

- **What:** the eleven rules goals stop getting their behavior from Go handlers and
  get it from `rulebooks/_core/goals.toml`. Each goal states six parts in data: what
  the pick needs, how progress moves, what resets it, the threshold, what completes
  it, and the XP it pays. A new goal engine reads those parts and checks every name
  when the engine is built. On the way, this effort removes what the new model has
  no place for: the goal-lock phase, the Seize Planet and Change Homeworld actions,
  and the goal process around them.
- **Why:** a goal with no Go handler does nothing today, so a second rulebook gets
  goal behavior only by writing Go. The handlers also read the rules wrongly in
  places: XP rounds down, a missed attack does not break Peaceable Kingdom, and
  Inside Enemy Territory counts stealth events, not assets held.
- **How it fits:** Effort 1 of three in `A.001.8` (arc `A.001`), on branch
  `feature/data-driven-goals`. Effort 2 makes goals read every batch. Effort 3 adds
  the turn-start goal pick. `F.040` rebuilds the two actions as multi-turn actions.

## Context / Goal

- Overview: [`A.001.8-data-driven-goals-overview-plan.md`](./A.001.8-data-driven-goals-overview-plan.md).
  Its Decisions 2, 3 and 4 widen this effort: the goal-lock phase, the two actions
  with the goal process, and Imperialists' entry are all deleted here. Decision 5
  places the rules doc edit. Decision 7 fixes `needs_actions` to display names.
- Discovery: [`A.001.8-data-driven-goals-discovery.md`](../discovery/A.001.8-data-driven-goals-discovery.md),
  D-1, D-5, D-8 to D-10, D-14 to D-20, D-22 to D-24, D-26, and its Goal Vocabulary.
  D-17 is amended: the parts are inline tables in the house style
  (`docs/contributing/engine.md`, Rulebook TOML style).
- Architecture read: goals, orchestrator, effect-mutation (Key Decisions),
  persistence (Key Decisions).
- The discovery left no open question. Decisions 1 to 5 below settle what the code
  showed in this session.
- Until Effort 2, goals read only the acting faction's action batch, as today. Until
  Effort 3, the wizard still sets goals (overview, Gaps between efforts).

## Decisions Ratified in Planning

1. **The `damage` event matches `AssetHPDelta` and `BaseHPDelta` only.** Robert's
   call. Corrects the mechanism of D-20, not its reading. D-20 says a homeworld Base
   hit emits `FactionHPDelta` and no `BaseHPDelta` (frozen log D75). The source
   emits both for every Base hit, homeworld or not (`attack.go:156`,
   `expand_influence.go:216`), and nothing emits a negative `FactionHPDelta` alone.
   So a homeworld Base hit already counts through `BaseHPDelta`, and a
   `FactionHPDelta` match would count each Base hit twice. Rejected: the match as
   D-20 wrote it; changing the emitters to match D75 (widens the effort).

2. **A measure reads the state the batch leaves, from a copy.** Robert's call. Goals
   run before the batch applies, and `stealthed_assets` is a count read from state.
   The measure clones the state, applies the batch to the clone with the mutation
   engine, and counts there. The goal completes in the same batch as the stealth.
   Precedent: the ability dispatch's `view` (`ability/dispatch.go`). Rejected:
   reading live state before the batch (one batch late, which is a full cycle late
   until Effort 2).

3. **`ActionTaken` shows in neither the feed nor the digest.** Robert's call. Narrows
   D-19's "a digest line and a TUI line". The feed header of the same batch already
   reads `action resolved: Attack`, and the digest has no beat for an action. The
   digest gets an explicit case that drops it; the feed folds it. History records
   it, which is D-19's purpose. Rejected: a feed line; a feed line and a new digest
   beat.

4. **Stat names in goal data are `"Force"`, `"Cunning"` and `"Wealth"`.** Robert's
   call. The discovery's block writes `"force"`; every `effects` entry writes
   `"Force"`, and D-18 says goals share those value words. The goal loader converts
   them with `toFactionStat`, as the effects loader does. Rejected: lower case (a
   second spelling and a second converter).

5. **Four phases, eight commits, removals first.** Robert's call. The overview
   named three seams in the order vocabulary, swap, removals. Removals go first,
   because the two process handlers, the two actions, the lock phase and
   `ActiveGoal`'s process fields hold each other up; once they are gone the swap is
   "ten handlers become data". Commit 5 lands the new engine tested and not wired;
   Commit 6 wires it. Rejected: three phases (the middle phase sits at the line
   budget); the overview's order (the swap would carry or stub the lock phase).

6. **The loader checks shape; the goal engine checks names.** Planning call (D-18).
   `loadGoals` reads the parts strictly and converts stat names. Every goal-only
   name (event, filter value, amount, measure, formula, start input) stays a string
   in `domain.Goal`. `goal.New(rulebook, actionNames, log)` compiles each goal and
   returns an error, which `NewWithRulebook` returns (Decision Record 12). `rulebook` cannot import `engine/goal`,
   so the loader cannot hold the vocabulary.

7. **A goal has exactly one of `threshold` or `complete`.** Planning call.
   `threshold` needs `progress`; `reset` needs `progress`; `half_progress` needs
   `threshold`. `F.013` adds a third completion kind beside the two.

8. **A reset wins its batch.** Planning call. When `reset` matches, progress goes to
   0 and the batch's progress events do not count. No shipped goal can match both in
   one batch: an action batch holds one `ActionTaken`.

9. **`contested_bonus` reads live state, and the goal engine takes no world index.**
   Planning call. The index is rebuilt only at turn setup, so it goes stale inside a
   turn, and Effort 2 reads goals on every batch. The reading is D-22's: a rival
   asset or Base on the world.

10. **A faction that holds a goal ID the rulebook does not know is skipped, with a
    warning.** Planning call. An older save can hold one (`G-004`, or `G-012`). Same
    contract as a held tag (persistence, Key Decisions). This is not the old silent
    skip: every goal in the data has behavior or the engine refuses to start.

11. **`GoalProgressed.Cause` is `goal_progress` or `goal_reset`.** Planning call. The
    old causes named the goal (`military_conquest`), and no Go names a goal now. No
    reader keys on the old values.

12. **`needs_actions` names are not checked when the engine is built.** Planning
    call. Seize Planet is unregistered on purpose, so there is nothing to check
    against. A misspelt name hides the goal from the pick; Effort 3 reads the list.

13. **A threshold completes only in a batch that moved progress.** Planning call.
    Keeps today's behavior: a handler returns nothing when its batch has no gain.

14. **Tests that only the removed process can drive are deleted, not parked.**
    Planning call (overview Decision 3). That covers the scenario tests of the lock
    and the process, and `TestDeepRooted_LostOnHomeworldChange`: nothing in play
    emits `HomeworldChanged` until `F.040`. `remove_tag_test.go` still tests the
    reactor.

15. **Goal IDs change in their own commit, after the swap.** Planning call. Once
    Commit 6 lands, no Go names a goal ID, so Commit 7 touches data and tests only.
    Commits 4 to 6 keep the bare IDs.

## Decision Record — Execution

1. **Two more deleted test files held helpers that live tests use (Commit 1).**
   Drift from Task 1, which named only `lock_test.go`. `movementMutationsFor` moved
   from `movement_lockskip_test.go` to `movement_lifecycle_test.go`.
   `testHexRouter` and `testWorldLoc` moved from `change_homeworld_test.go` to
   `test_helpers_test.go`, without the `distanceFn` field: only the Change
   Homeworld tests set it. No helper from `lock_test.go` was still in use.

2. **`ImplementedActions` loses its two entries (Commit 1).** Drift: the plan did
   not list `tui/views/turn/overlay/registry.go`, which named Seize Planet and
   Change Homeworld as actions the TUI can drive. The comment in
   `tui/adapter/action_collector.go` that counted "12 actions" now gives no count.

3. **The fence-sign in `engine/world/world.go` still names ChangeHomeworld
   (Commit 1).** `WorldsWithin` carries a `// deliberate:` line that compares its
   raw hex distance with the drift-priced journey of ChangeHomeworld. A fence-sign
   is binding, and `F.040` rebuilds the action, so the line stays. It is the one
   hit of the Task 20 check. Rejected: rewording the fence-sign without Robert.

4. **The logging tests keep `goal_lock` as a sample phase name (Commit 2).**
   `internal/logging/handler_test.go` and `banner_test.go` use it to test how a
   two-word phase name becomes a banner. They test the formatter, not the phase,
   and no engine code logs that phase now. Gap left open: swap the sample for a
   live phase name if the stale word misleads.

5. **No test read an action batch by position (Commit 3).** Closes the Open
   Question. With `ActionTaken` first in every action batch, the whole suite
   passed with no expectation moved.

6. **`goals.toml` was laid out by hand, not by TomlSmith (Commit 4).** The
   formatter is an editor command and has no shell form. Every part but G-010's
   `complete` fits in 120 columns and sits on one line; G-010's opens, as Task 3
   gives it. Gap left open: a Format Document pass in the editor should change
   nothing; if it does, keep its layout. `TestGoalFilters_Used` lives in
   `domain/goal_test.go`, the option Task 5 allows.

7. **The evaluate tests share one fixture type, not loose helpers (Commit 5).**
   Task 10 named `holder(goalName, progress)` and `rival()`. A `goalFixture` built
   by `newGoalFixture(t, goalName, progress)` holds the compiled goal, the test
   rulebook, `f1`, `f2`, a third party `f3` and the state, so each test changes
   only what it needs. The ratings are the ones Task 10 gives. Three small
   additions to the cases: a `force-equal` definition (`MinRating` 3) inside the
   Invincible Valor test, for the equal-rating row the four shared definitions do
   not cover; two more damage rows that give nothing (a `FactionHPDelta` alone,
   and a Base hit caused by `f3`); and `TestCompile_AcceptsValidGoal`, which
   proves the table's base goal compiles, so each refusal row fails for its own
   change.

8. **The swap's tests share small helpers, and the missed attack hits Hitmen
   (Commit 6).** `goal_test.go` adds `newEngine` and `discardLog`;
   `TestUpdateProgress_RunsTheHeldGoal` reuses `goalFixture`, whose rulebook has a
   Force definition for the kill. `scenarios/goals_test.go` adds `scriptAction`,
   `scriptAttackOnFirstDefender`, `runOneCycle` and `checkProgress` next to
   `goalID`. Task 6 asks for a miss with "no counter damage", and the harness's
   Security Personnel counters for 1d4, so beta's defender becomes Hitmen
   (`SWN-F1-002`, no counter) and the sequence is `[1, 10]` as planned. Blood the
   Enemy's sequence is `[10, 1, 2]` (3 damage), and the test also checks the Base
   lost 3 HP. Military Conquest's is `[10, 1, 3]`. No sequence ran short or long.
   The handler tests in `engine/goal/goals/` went with the package (Task 1).

9. **The prefix reaches test messages and comments too (Commit 7).** Task 3's
   replace covers quoted literals only. Three failure messages
   (`simple_actions_test.go`, `rulebook_test.go`, `effects_test.go`) and the
   section heads of `evaluate_test.go` named bare IDs, so they take the prefix,
   and no bare goal ID is left in `internal/` or `rulebooks/`.
   `TestUpdateProgress_UnknownGoalIsSkipped` now holds `CORE-G-012`, which the
   rulebook still does not know. The re-seed ran with `GM_TOOLKIT_HOME` set to
   `campaigns/`, which Task 4's commands leave out. The menu showed neither
   removed action, cycle 1 completed, and the detail view names Inside Enemy
   Territory. The `purchase` reset made `purchase` the seed of `sandbox-claude`.

10. **The rules doc and the data name the seizure goal differently (Commit 8).**
    The rules doc's table and its new Goal Rulings say "Planetary Seizure", the
    book's name; `goals.toml` names `CORE-G-004` "World Seizure". The split is
    older than this effort, and Task 1 keeps the rules doc's name. Gap left open:
    pick one name when `F.040` rebuilds Seize Planet.

11. **A batch that abandons the held goal moves it no further (effort review).**
    The Abandon Goal batch opens with `ActionTaken{Action: "Abandon Goal"}`, which
    passes Peaceable Kingdom's `action_not = ["Attack"]`: at progress 3 the goal
    completed and paid XP in the batch that gave it up. The old handler had the
    same hole. `evaluate` now returns nothing when the batch holds the holder's
    `GoalAbandoned`.

12. **Action filter names are checked when the engine is built (effort review).**
    A misspelt `action` or `action_not` value loaded and never matched. `goal.New`
    takes the registered action names (`ActionEngine.Names()`), and
    `NewWithRulebook` builds the goal engine after the actions register. Decision
    12 stands: `needs_actions` is still not checked.

13. **The goal engine keeps the rulebook it compiled from (effort review).**
    `UpdateProgress` drops its rulebook parameter, so a goal is never evaluated
    against definitions it was not compiled with. The `log` field now carries a
    debug line from `New`, the startup channel of the two-channel logger.

14. **`state.FactionState.Clone` replaces two hand-rolled copies (effort review).**
    `project` and `ability.NewView` both use it. `NewView` used to share the live
    `Governments` map with its view; the clone copies it, and `CurrentTurn` too.

15. **Smaller review fixes; the Work Breakdown code shows the shape before them.**
    `checkXP` rejects an unknown formula before it checks the amount.
    `progressDelta` always sums `matched.amount`. The `ctx evalContext` parameters
    are now `evaluation`. The stealth measure builds the rival-homeworld set once
    per call. Rejected: projecting only the holder and the governments. A capture
    needs the losing faction in the copy, or the holder's captured stealthed asset
    goes uncounted; the full clone stays until E2 gives real per-batch costs. The
    code blocks in Phases 2 to 4 are not rewritten for entries 11 to
    15; the source wins where they differ.

---

## Open Questions — To Ratify at Implementation Time

- ~~**Tests that read an action batch by position.**~~ Closed in Commit 3: none
  broke (Decision Record 5).

## Shared Context

### Branch, cadence and checks

- Branch `feature/data-driven-goals`. One commit per plan commit, scope `E1`. At the
  effort boundary: `/gestalt:effort-close A.001.8.E1`.
- Every commit builds and passes `go build ./... && go vet ./... && go test ./...`
  from the repo root. Run `gofmt -w` on each touched Go file.
- Edit files only with Edit and Write. Delete files with `git rm`.
- `mockgen` is not installed, so Commit 1 edits
  `engine/action/actions/mocks/mock_collector.go` by hand.
- A removal task ends with "drop the imports the compiler reports unused". The
  compiler is the check; the plan does not list each import.
- Test tasks are written in plain words. A test edit given as "replace every" is an
  Edit with `replace_all: true`.
- Rulebook TOML follows the house style. Run TomlSmith's Format Document on a
  touched rulebook file; the formatter owns the line breaks.
- Paths in task titles are relative to `internal/faction/` unless they start with
  `rulebooks/`, `docs/` or `campaigns/`.

### Campaigns

- Robert's `campaigns/sandbox` is not edited (D-26). Its `rulebook/goals.toml` copy
  stops loading at Commit 4 (`difficulty` is an unknown key), and its `combat` and
  `purchase` seeds and `TESTING.md` walk through Planetary Seizure, which Commit 1
  removes. Robert re-composes and re-seeds it.
- Claude re-seeds the gitignored `campaigns/sandbox-claude` in Commit 7, and runs
  the app only with `--campaign sandbox-claude`.

### Pre-merge additions (beyond the overview's)

- **Graduate to Key Decisions (goals):** a measure reads a projected copy of the
  state (Decision 2); the loader checks shape and the engine checks names
  (Decision 6); a reset wins its batch (Decision 8).
- **Graduate to Key Decisions (effect-mutation):** the `damage` event and the
  double emission of a Base hit (Decision 1); `ActionTaken` shown nowhere
  (Decision 3).
- **`docs/contributing/overview.md`:** its "four sub-engines share one pattern"
  part still names the goal handler registry and the silent skip. Rewrite it with
  the goals page (overview Decision 6).

## Out of Scope

- Goals reading every batch, `GoalFailed` and the lost-target rule: Effort 2.
- The turn-start pick, both gates, the start record of `G-009` and the wizard's goal
  step: Effort 3. Effort 1 loads `start` and checks it, and nothing reads it.
- Seize Planet, Change Homeworld, the seizure's completion, the resist record's
  writer and Imperialists' entry: `F.040`.
- The architecture pages: the initiative close rewrites them (overview Decision 6).
- Robert's `campaigns/sandbox` files.

---

## Work Breakdown

**Overview**

- [x] **Phase 1 — Removals**
  - Commit 1 — `refactor(E1): remove the goal process and its two actions` (20 tasks)
  - Commit 2 — `refactor(E1): delete the goal-lock phase` (11 tasks)
- [x] **Phase 2 — Vocabulary**
  - Commit 3 — `feat(E1): ActionTaken opens each action batch` (6 tasks)
  - Commit 4 — `feat(E1): goal parts load from data` (6 tasks)
- [x] **Phase 3 — Matchers**
  - Commit 5 — `feat(E1): a goal engine that reads goal parts` (10 tasks)
- [x] **Phase 4 — Swap**
  - Commit 6 — `feat(E1): goals run from data` (7 tasks)
  - Commit 7 — `refactor(E1): goal IDs carry the CORE prefix` (4 tasks)
  - Commit 8 — `docs(E1): goal rules readings and the goal recipe` (2 tasks)

### Phase 1 — Removals

**In plain words:** Seize Planet, Change Homeworld, the multi-turn goal process and
the goal-lock phase leave the code. When this lands, ten goal handlers remain, each
one reads only its action batch, and no turn step runs for a goal lock.

#### Commit 1 — `refactor(E1): remove the goal process and its two actions`

Overview Decisions 3 and 4, D-5 and D-24. After this commit `G-004` is a goal with
no handler (a silent skip) until Commit 6 gives it parts.

##### Task 1 — delete files

```
git rm internal/faction/engine/action/actions/seize_planet.go \
       internal/faction/engine/action/actions/seize_planet_test.go \
       internal/faction/engine/action/actions/change_homeworld.go \
       internal/faction/engine/action/actions/change_homeworld_test.go \
       internal/faction/engine/goal/goals/planetary_seizure.go \
       internal/faction/engine/goal/goals/change_homeworld.go \
       internal/faction/engine/goal/goals/lock_test.go \
       internal/faction/engine/testharness/scenarios/goal_lock_test.go \
       internal/faction/engine/testharness/scenarios/movement_lockskip_test.go \
       internal/faction/engine/testharness/scenarios/change_homeworld_lifecycle_test.go \
       internal/faction/tui/views/turn/overlay/worldselect.go
```

`worldselect.go` holds only `NewWorldSelect`, and its two callers go in Task 10.
Two deleted test files hold helpers that live tests use; move them (Decision
Record 1): `movementMutationsFor` to `movement_lifecycle_test.go`, and
`testHexRouter` with `testWorldLoc` to `test_helpers_test.go`.

##### Task 2 — `engine/action/actions/register.go`

The `rulebook` parameter served only Change Homeworld.

**Find:**
```go
	worldEngine *world.WorldEngine,
	rulebook *rulebook.Rulebook,
	effectsEngine *effect.EffectsEngine,
) {
```
**Replace with:**
```go
	worldEngine *world.WorldEngine,
	effectsEngine *effect.EffectsEngine,
) {
```
**Find:**
```go
	ae.Register(func(_ action.Collector) action.Action { return NewAbandonGoal() })
	ae.Register(func(c action.Collector) action.Action { return NewSeizePlanet(c, worldIndex(worldEngine)) })
	ae.Register(func(c action.Collector) action.Action { return NewChangeHomeworld(c, worldEngine, rulebook) })
}
```
**Replace with:**
```go
	ae.Register(func(_ action.Collector) action.Action { return NewAbandonGoal() })
}
```
Drop the `rulebook` import.

##### Task 3 — `engine/core.go`

**Find:**
```go
	actions.RegisterDefaultActions(e.Action, func() dice.Roller { return e.Rand }, e.Hooks, e.World, e.Rulebook, e.Effect)
```
**Replace with:**
```go
	actions.RegisterDefaultActions(e.Action, func() dice.Roller { return e.Rand }, e.Hooks, e.World, e.Effect)
```

##### Task 4 — `engine/goal/goal.go`

**Find:**
```go
	engine.Register(goals.IntelligenceCoup{})
	engine.Register(goals.PlanetarySeizure{})
	engine.Register(goals.ExpandInfluence{})
```
**Replace with:**
```go
	engine.Register(goals.IntelligenceCoup{})
	engine.Register(goals.ExpandInfluence{})
```
**Find:**
```go
	engine.Register(goals.WealthOfWorlds{})
	engine.Register(goals.ChangeHomeworld{})
	return engine
```
**Replace with:**
```go
	engine.Register(goals.WealthOfWorlds{})
	return engine
```

##### Task 5 — `engine/goal/goals/helpers.go`

Delete `factionHasUnstealthedAssetOn` and `calcPlanetarySeizureXP`, which only the
seizure handler called. Drop the imports the compiler reports unused.

##### Task 6 — `engine/action/collector.go`

**Find:**
```go
	SelectBribeTarget(faction *domain.Faction, factionState *state.FactionState) (*domain.Base, int, error)
	SelectSeizeTarget(faction *domain.Faction, factionState *state.FactionState) (string, error)
	SelectChangeHomeworldTarget(faction *domain.Faction, factionState *state.FactionState) (string, error)
}
```
**Replace with:**
```go
	SelectBribeTarget(faction *domain.Faction, factionState *state.FactionState) (*domain.Base, int, error)
}
```

##### Task 7 — the collector implementations

- **`engine/action/actions/mocks/mock_collector.go`** — delete the four generated
  blocks for the two methods: `SelectChangeHomeworldTarget` on `MockCollector` and
  on `MockCollectorMockRecorder`, and the same pair for `SelectSeizeTarget`. Each
  block starts at its `// Select… mocks base method.` or
  `// Select… indicates an expected call of …` comment.
- **`engine/testharness/collector.go`** — delete the fields `SelectSeizeTargetFn`
  and `SelectChangeHomeworldTargetFn`, and the methods `SelectSeizeTarget` and
  `SelectChangeHomeworldTarget`. Run `gofmt -w`; the field column re-aligns.
- **`tui/adapter/smoke.go`** — delete the `SmokeActionCollector` methods
  `SelectSeizeTarget` and `SelectChangeHomeworldTarget`.

##### Task 8 — `tui/adapter/action_collector.go`

Delete four functions, from the first to the last in one block: `SelectSeizeTarget`,
`deriveContestedWorlds` with its comment, `SelectChangeHomeworldTarget`, and
`deriveHomeworldTargets` with its comment. The block starts after
`deriveBribeTargets` and ends before `SelectPurchaseEnablerAsset`. Leave one blank
line between those two. Drop the imports the compiler reports unused (`world`).

##### Task 9 — `tui/adapter/channels.go`

Removing two kinds renumbers the later ones. Nothing stores the numbers:
`askEntries` is a map keyed on the constants.

**Find:**
```go
	AskSelectBribeTarget
	AskSelectSeizeTarget
	AskSelectChangeHomeworldTarget
	AskSelectAttackers
```
**Replace with:**
```go
	AskSelectBribeTarget
	AskSelectAttackers
```
**Find:**
```go
type SelectSeizeTargetPayload struct {
	Worlds     []string
	WorldNames map[string]string // worldID -> display name (adapter-built)
}
type SelectChangeHomeworldTargetPayload struct {
	Worlds     []string
	WorldNames map[string]string
}
type SelectAttackersPayload struct{ Eligible []*domain.Asset }
```
**Replace with:**
```go
type SelectAttackersPayload struct{ Eligible []*domain.Asset }
```

##### Task 10 — `tui/views/turn/execution/execution.go`

**Find:**
```go
	adapter.AskSelectSeizeTarget: {"Seize Planet", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.SelectSeizeTargetPayload)
		return overlay.NewWorldSelect("Seize which world?", "", p.Worlds, p.WorldNames)
	}},
	adapter.AskSelectChangeHomeworldTarget: {"Change Homeworld", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.SelectChangeHomeworldTargetPayload)
		return overlay.NewWorldSelect("New homeworld?", "", p.Worlds, p.WorldNames)
	}},
	adapter.AskSelectAttackers: {"Select Attackers", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
```
**Replace with:**
```go
	adapter.AskSelectAttackers: {"Select Attackers", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
```

In `tui/views/turn/overlay/registry.go`, delete the `"Seize Planet"` and
`"Change Homeworld"` entries of `ImplementedActions` (Decision Record 2).

##### Task 11 — `domain/mutation.go`

Delete three types with their comments and `Type` methods: `GoalInitiated`,
`GoalTurnsTick` and `GoalPhaseAdvanced`. `GoalProgressed` stays.

**Find:**
```go
// GoalInitiated records that a faction initiated a goal. Used for bookkeeping and validation of goal-specific actions; has no direct mechanical effect.
type GoalInitiated struct {
	FactionID         string   `json:"faction_id"`
	GoalID            string   `json:"goal_id"`
	TargetWorld       Location `json:"target_world"`
	ProcessPhase      int      `json:"process_phase"`
	Cause             string   `json:"cause"`
	CausedByFactionID string   `json:"caused_by_faction_id"`
}

func (m GoalInitiated) Type() string { return "goal_initiated" }

// GoalProgressed records a progress increment (or reset) on the active goal.
```
**Replace with:**
```go
// GoalProgressed records a progress increment (or reset) on the active goal.
```
**Find:**
```go
// GoalTurnsTick decrements ActiveGoal.TurnsRemaining by 1.
type GoalTurnsTick struct {
	FactionID string `json:"faction_id"`
	GoalID    string `json:"goal_id"`
	Cause     string `json:"cause"`
}

func (m GoalTurnsTick) Type() string { return "goal_turns_tick" }

// GoalPhaseAdvanced records a phase transition in a multi-phase goal,
// setting both ProcessPhase and TurnsRemaining atomically.
type GoalPhaseAdvanced struct {
	FactionID      string `json:"faction_id"`
	GoalID         string `json:"goal_id"`
	ProcessPhase   int    `json:"process_phase"`
	TurnsRemaining int    `json:"turns_remaining"`
	Cause          string `json:"cause"`
}

func (m GoalPhaseAdvanced) Type() string { return "goal_phase_advanced" }

// XPSpent decrements a faction's XP by Amount. Emitted by the stat raise phase.
```
**Replace with:**
```go
// XPSpent decrements a faction's XP by Amount. Emitted by the stat raise phase.
```

##### Task 12 — `engine/mutation/mutation.go`

Delete the three apply arms.

**Find:**
```go
		case domain.GoalInitiated:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.ActiveGoal = &domain.ActiveGoal{
					GoalID:       v.GoalID,
					ProcessPhase: v.ProcessPhase,
					TargetWorld:  v.TargetWorld,
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.GoalProgressed:
```
**Replace with:**
```go
		case domain.GoalProgressed:
```
**Find:**
```go
		case domain.GoalTurnsTick:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if faction.ActiveGoal != nil && faction.ActiveGoal.GoalID == v.GoalID {
					faction.ActiveGoal.TurnsRemaining--
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.GoalPhaseAdvanced:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if faction.ActiveGoal != nil && faction.ActiveGoal.GoalID == v.GoalID {
					faction.ActiveGoal.ProcessPhase = v.ProcessPhase
					faction.ActiveGoal.TurnsRemaining = v.TurnsRemaining
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.XPSpent:
```
**Replace with:**
```go
		case domain.XPSpent:
```

##### Task 13 — `narrative/digest/build.go`

**Find:**
```go
			case "goal_initiated":
				// dropped — bookkeeping only
			case "goal_progressed":
				// dropped — bookkeeping only
			case "goal_turns_tick":
				// dropped — bookkeeping only
			case "goal_phase_advanced":
				// dropped — bookkeeping only
```
**Replace with:**
```go
			case "goal_progressed":
				// dropped — bookkeeping only
```

##### Task 14 — `tui/views/turn/execution/format.go`

**Find:**
```go
		// GoalTurnsTick, GoalProgressed, GoalInitiated, GoalPhaseAdvanced,
		// AssetMaintainedFlag, AssetStealthApplied, AssetStealthCleared,
```
**Replace with:**
```go
		// GoalProgressed, AssetMaintainedFlag, AssetStealthApplied, AssetStealthCleared,
```

##### Task 15 — `domain/faction.go`

D-24.

**Find:**
```go
type ActiveGoal struct {
	GoalID          string   `toml:"goal_id"`
	TargetFactionID string   `toml:"target_faction_id"`
	TargetWorld     Location `toml:"target_world"`
	Progress        int      `toml:"progress"`
	ProcessPhase    int      `toml:"process_phase"`
	TurnsRemaining  int      `toml:"turns_remaining"`
}
```
**Replace with:**
```go
type ActiveGoal struct {
	GoalID          string `toml:"goal_id"`
	TargetFactionID string `toml:"target_faction_id"`
	Progress        int    `toml:"progress"`
}
```
The state loader is not strict, so a save with the old keys still loads and drops
them.

##### Task 16 — `engine/effect/effects/roll_match.go` and `bonus_die.go`

Overview Decision 4. In `roll_match.go`:

**Find:**
```go
// matches without a rulebook. goal matches a roll against a rival on the target
// world of the owner's active goal. detection matches only a roll made by the
// side testing to detect stealth.
```
**Replace with:**
```go
// matches without a rulebook. detection matches only a roll made by the side
// testing to detect stealth.
```
**Find:**
```go
	// deliberate: a goal roll needs a rival, not only the target world, because the counterstrike rolls the attacker against itself with no Opponent (Effort 2 Decision Record 6).
	if filters.Goal != "" && (ctx.Opponent == nil || !onGoalTarget(owner, filters.Goal, ctx.World)) {
		return false
	}
	return true
}

// onGoalTarget reports whether owner's active goal is goalID and world is its
// target. A goal with no target chosen yet has none.
func onGoalTarget(owner *domain.Faction, goalID, world string) bool {
	goal := owner.ActiveGoal
	return goal != nil && goal.GoalID == goalID && world != "" && goal.TargetWorld.WorldID == world
}
```
**Replace with:**
```go
	return true
}
```
In `bonus_die.go`:

**Find:**
```go
				domain.FilterTechLevel, domain.FilterOpponentTechLevel, domain.FilterGoal, domain.FilterDetection,
```
**Replace with:**
```go
				domain.FilterTechLevel, domain.FilterOpponentTechLevel, domain.FilterDetection,
```
`domain.FilterGoal` stays: `xp_roll` accepts it.

##### Task 17 — `rulebooks/swn/tags.toml`

Imperialists keeps its text and loses its entry until `F.040`.

**Find:**
```toml
effect = "May roll an extra d10 for attacks made as part of a Seize Planet action."
effects = [
  {
    trigger = "own_roll",
    effect = "bonus_die",
    filters = { owner = "own", scope = "faction", phase = ["attack"], goal = "G-004", action = "Attack" },
    bonus_die = { sides = 10, count = 1 },
  },
]
```
**Replace with:**
```toml
effect = "May roll an extra d10 for attacks made as part of a Seize Planet action."
```

##### Task 18 — `rulebooks/_core/goals.toml`

D-5. Delete the last entry and the blank line before it.

**Find:**
```toml

[goals.ChangeHomeworld]
id = "G-012"
name = "Change Homeworld"
description = "Relocate the faction's homeworld to another world where it has a Base of Influence. The move takes a number of turns equal to the distance between the two worlds plus one."
difficulty = "none"
```
**Replace with:** nothing.

##### Task 19 — tests that named the removed parts

- **`engine/goal/goals/progress_test.go`** — delete
  `TestPlanetarySeizure_CaptureOfLastRival` and
  `TestPlanetarySeizure_ShelterOfLastRival`.
- **`engine/action/actions/spatial_index_test.go`** — delete
  `TestSeizePlanetTargetWorlds_ReturnsContestedFragments`.
- **`engine/testharness/scenarios/capture_test.go`** — delete
  `TestTreacheryCaptureAdvancesSeizureSameTurn` with its comment.
- **`engine/testharness/scenarios/tags_test.go`** — delete
  `TestDeepRooted_LostOnHomeworldChange` with its scenario comment line
  (Decision 14).
- **`engine/effect/effects/roll_match_test.go`** — delete the `planning` faction,
  the `seizure` and `otherGoal` filters, and the five cases that use them ("goal
  target world" to "goal target not yet chosen"). Remove `ActiveGoal` from `alpha`.
- **`rulebook/rulebook_test.go`** — in `TestLoad_TagBonusDice`, delete the
  `SWN-T-006` row. Add one assertion: `rb.Tags["SWN-T-006"].Effects` is empty.
- **`state/faction_state_test.go`** — in the round-trip fixture, change
  `ActiveGoal: &domain.ActiveGoal{GoalID: "G-004", TargetWorld: awkward}` to
  `ActiveGoal: &domain.ActiveGoal{GoalID: "G-004"}`.
- **`tui/adapter/collector_test.go`** — delete the `SelectSeizeTarget` and
  `SelectChangeHomeworldTarget` rows.
- **`tui/views/turn/execution/format_test.go`** — in `TestFormatMutation_Folded` and
  `TestRenderEvent_FoldedMutationsOmitted`, delete each `domain.GoalTurnsTick{}`
  entry and the word `GoalTurnsTick` from the comment and the check; the tests keep
  `GoalProgressed`.

Drop the imports the compiler reports unused in each file.

##### Task 20 — check the removal

`go build ./... && go vet ./... && go test ./...`. Then confirm nothing names the
removed parts:

```
grep -rnE "SeizePlanet|ChangeHomeworld|SelectSeizeTarget|GoalInitiated|GoalTurnsTick|GoalPhaseAdvanced|ProcessPhase|TurnsRemaining|onGoalTarget" internal/ cmd/
```

The only hit allowed is the fence-sign in `engine/world/world.go` (Decision
Record 3). `domain.HomeworldChanged` and its readers stay; the pattern does not
match them.

##### Commit message

```
refactor(E1): remove the goal process and its two actions

- Seize Planet and Change Homeworld are deleted with their collector asks,
  overlays and tests; F.040 rebuilds them as multi-turn actions
- the G-004 and G-012 handlers, GoalInitiated, GoalTurnsTick and
  GoalPhaseAdvanced are deleted; G-012 leaves goals.toml (D-5)
- ActiveGoal loses TargetWorld, ProcessPhase and TurnsRemaining (D-24)
- Imperialists' effects entry and the bonus_die goal filter are removed until
  F.040 (overview Decision 4)
```

#### Commit 2 — `refactor(E1): delete the goal-lock phase`

Overview Decision 2. No behavior change: after Commit 1 every remaining handler
returns `LockNone` and no mutations.

##### Task 1 — `engine/orchestrator.go`

Drop the `locks` import.

**Find:**
```go
	CheckpointActionResult = "action_result"
	CheckpointGoalLocked   = "goal_locked"
	CheckpointCycleSummary = "cycle_summary"
```
**Replace with:**
```go
	CheckpointActionResult = "action_result"
	CheckpointCycleSummary = "cycle_summary"
```
**Find:**
```go
// RunFactionTurn drives one faction's turn from setup through state save:
// stat-raise, bookkeeping, movement, goal-lock (gating the action), action.
```
**Replace with:**
```go
// RunFactionTurn drives one faction's turn from setup through state save:
// stat-raise, bookkeeping, movement, action.
```
**Find:**
```go
	lock, err := e.runGoalLockPhase(faction, factionState, paths, collectors, observer, turnLog)
	if err != nil {
		return false, err
	}

	if lock.Type == locks.LockSkip {
		return e.finishFactionTurn(factionState, faction, paths, collectors, observer, turnLog)
	}

	if err := e.runActionPhase(faction, factionState, paths, collectors, observer, lock, turnLog); err != nil {
```
**Replace with:**
```go
	if err := e.runActionPhase(faction, factionState, paths, collectors, observer, turnLog); err != nil {
```

Delete the whole function `runGoalLockPhase`, from `func (e *Engine) runGoalLockPhase(`
to its closing brace before `func (e *Engine) runBookkeepingPhase(`.

**Find:**
```go
	observer TurnObserver,
	lock locks.GoalLock,
	turnLog *slog.Logger,
) error {
	phaseLog := turnLog.With("phase", "action")
```
**Replace with:**
```go
	observer TurnObserver,
	turnLog *slog.Logger,
) error {
	phaseLog := turnLog.With("phase", "action")
```
**Find:**
```go
	available := e.Action.AvailableActions(faction, factionState, e.Rulebook, collectors.Action, actionLog)
	if lock.Type == locks.LockRestrictActions {
		available = filterAllowedActions(available, lock.AllowedActions)
	}

```
**Replace with:**
```go
	available := e.Action.AvailableActions(faction, factionState, e.Rulebook, collectors.Action, actionLog)

```
**Find:**
```go
// filterAllowedActions narrows a list of available actions to those whose
// Name() appears in allowed. Used when a goal lock restricts the action set
// (e.g. Planetary Seizure Phase 1 → Attack only).
// firePurchaseEffects runs the on_purchase effects of every asset the action
```
**Replace with:**
```go
// firePurchaseEffects runs the on_purchase effects of every asset the action
```
**Find:**
```go
func filterAllowedActions(available []action.Action, allowed []string) []action.Action {
	if len(allowed) == 0 {
		return nil
	}
	allowedSet := make(map[string]bool, len(allowed))
	for _, name := range allowed {
		allowedSet[name] = true
	}
	filtered := make([]action.Action, 0, len(available))
	for _, a := range available {
		if allowedSet[a.Name()] {
			filtered = append(filtered, a)
		}
	}
	return filtered
}

```
**Replace with:** nothing.

##### Task 2 — `engine/observer.go`

Drop the `locks` import.

**Find:**
```go
	OnGoalLockApplied(faction *domain.Faction, lock locks.GoalLock, mutations []domain.Mutation)
```
**Replace with:** nothing.

##### Task 3 — `engine/goal/goal.go`

Drop the `locks` import.

**Find:**
```go
	CheckLock(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) (locks.GoalLock, []domain.Mutation)
```
**Replace with:** nothing.
**Find:**
```go
// CheckLock evaluates the faction's active goal and returns the appropriate lock
// state. Must be called before bookkeeping and action selection each turn.
func (ge *GoalEngine) CheckLock(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook, log *slog.Logger) (locks.GoalLock, []domain.Mutation) {
	log.Info("checking goal lock")
	if faction.ActiveGoal == nil {
		return locks.GoalLock{Type: locks.LockNone}, nil
	}
	handler, ok := ge.handlers[faction.ActiveGoal.GoalID]
	if !ok {
		// Data-only goal: present in goals.toml, no Go handler registered. Intentional.
		return locks.GoalLock{Type: locks.LockNone}, nil
	}
	lock, mutations := handler.CheckLock(faction, factionState, rulebook)
	log.Info("goal lock", "type", lock.Type, "mutations", len(mutations))
	return lock, mutations
}

```
**Replace with:** nothing.

##### Task 4 — the ten handlers in `engine/goal/goals/`

In each of `military_conquest.go`, `commercial_expansion.go`,
`intelligence_coup.go`, `expand_influence.go`, `blood_the_enemy.go`,
`peaceable_kingdom.go`, `destroy_the_foe.go`, `inside_enemy_territory.go`,
`invincible_valor.go` and `wealth_of_worlds.go`, delete the `CheckLock` method and
the `locks` import. The method has the same body in all ten; only the receiver type
differs:

```go
func (MilitaryConquest) CheckLock(_ *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) (locks.GoalLock, []domain.Mutation) {
	return locks.GoalLock{Type: locks.LockNone}, nil
}

```

##### Task 5 — delete the `locks` package

```
git rm -r internal/faction/engine/goal/locks
```

##### Task 6 — `tui/adapter/channels.go`

Drop the `locks` import. Removing the kind renumbers the later ones; nothing stores
the numbers.

**Find:**
```go
	EvtFactionSkipped
	EvtGoalLockApplied
	EvtBookkeepingApplied
```
**Replace with:**
```go
	EvtFactionSkipped
	EvtBookkeepingApplied
```
**Find:**
```go
type GoalLockAppliedPayload struct {
	Lock      locks.GoalLock
	Mutations []domain.Mutation
}
type BookkeepingAppliedPayload struct {
```
**Replace with:**
```go
type BookkeepingAppliedPayload struct {
```

##### Task 7 — `tui/adapter/observer.go` and `tui/adapter/smoke.go`

In `observer.go`, drop the `locks` import.

**Find:**
```go
func (o *observer) OnGoalLockApplied(faction *domain.Faction, lock locks.GoalLock, muts []domain.Mutation) {
	o.adapter.emit(EvtGoalLockApplied, GoalLockAppliedPayload{Lock: lock, Mutations: muts})
}
```
**Replace with:** nothing.

In `smoke.go`, drop the `locks` import.

**Find:**
```go
func (o *SmokeObserver) OnGoalLockApplied(faction *domain.Faction, lock locks.GoalLock, mutations []domain.Mutation) {
	o.log.Info("EVT GoalLockApplied", "faction", faction.ID, "lock", lock.Type, "mutations", len(mutations))
}
```
**Replace with:** nothing.

##### Task 8 — `tui/views/turn/execution/format.go`

**Find:**
```go
	case adapter.EvtGoalLockApplied:
		return phaseHeader("goal-lock applied"), msg.Payload.(adapter.GoalLockAppliedPayload).Mutations
```
**Replace with:** nothing.

##### Task 9 — `engine/testharness/observer.go`

Drop the `locks` import.

**Find:**
```go
func (r *RecordingObserver) OnGoalLockApplied(faction *domain.Faction, lock locks.GoalLock, mutations []domain.Mutation) {
	r.Events = append(r.Events, ObservedEvent{
		Kind:    "GoalLockApplied",
		Faction: faction,
		Payload: GoalLockPayload{Lock: lock, Mutations: mutations},
	})
}

```
**Replace with:** nothing.

**Find:**
```go
// GoalLockPayload is the Payload type for GoalLockApplied events.
type GoalLockPayload struct {
	Lock      locks.GoalLock
	Mutations []domain.Mutation
}

```
**Replace with:** nothing.

##### Task 10 — tests

- **`engine/orchestrator_test.go`** — delete `TestFilterAllowedActions`, the
  `errorObserver.OnGoalLockApplied` method and the `locks` import. Delete
  `stubAction` and its methods if nothing else in the package uses them.
- **`engine/goal/goal_test.go`** — delete `TestCheckLock_NoActiveGoal` and
  `TestCheckLock_DataOnlyGoal_NoHandler` and the `locks` import.
- **`engine/testharness/scenarios/full_cycle_test.go`** — remove `"GoalLockApplied"`
  from both rows of `wantKinds`.
- **`tui/views/turn/execution/format_test.go`** — delete
  `TestRenderEvent_GoalLockApplied_IncludesMutations`.

Drop the imports the compiler reports unused in each file.

##### Task 11 — check the removal

`go build ./... && go vet ./... && go test ./...`. Then:

```
grep -rnE "GoalLock|goal_locked|CheckLock|goal/locks|filterAllowedActions" internal/ cmd/
```

No hit is allowed.

##### Commit message

```
refactor(E1): delete the goal-lock phase

- runGoalLockPhase, the LockSkip short-circuit, filterAllowedActions and the
  goal_locked checkpoint are gone; a turn runs movement, then the action
- the locks package and Handler.CheckLock are deleted
- TurnObserver.OnGoalLockApplied and the adapter's goal-lock event are deleted
  (overview Decision 2)
```

### Phase 2 — Vocabulary

**In plain words:** every action batch starts with a record of the action taken, and
`goals.toml` states each goal's six parts, which the loader reads strictly. When this
lands the handlers still run the goals; the parts are loaded and not yet read.

#### Commit 3 — `feat(E1): ActionTaken opens each action batch`

D-19 and Decision 3. Peaceable Kingdom's handler does not read it yet; Commit 6 does.

##### Task 1 — `domain/mutation.go`

**Find:**
```go
func (m GoalProgressed) Type() string { return "goal_progressed" }
```
**Replace with:**
```go
func (m GoalProgressed) Type() string { return "goal_progressed" }

// ActionTaken records the action a faction took in its action phase. It changes
// no state: goals read it from the batch, and history keeps it (D-19). Action is
// the action's display name.
type ActionTaken struct {
	FactionID string `json:"faction_id"`
	Action    string `json:"action"`
}

func (m ActionTaken) Type() string { return "action_taken" }
```

##### Task 2 — `engine/mutation/mutation.go`

**Find:**
```go
		case domain.PermissionRefused:
			// record only: the refused action bought nothing
```
**Replace with:**
```go
		case domain.PermissionRefused:
			// record only: the refused action bought nothing
		case domain.ActionTaken:
			// record only: goals read it from the batch
```

##### Task 3 — `engine/orchestrator.go`

In `runActionPhase`, after `e.Action.Run` succeeds.

**Find:**
```go
	var worldIndex *world.Index
	if e.World == nil {
		return ErrWorldEngineUnavailable
	}
```
**Replace with:**
```go
	// deliberate: the record opens the batch, and a pass or a failed action leaves none, because goals count only an action the faction took (D-16, D-19).
	actionMutations = append([]domain.Mutation{domain.ActionTaken{FactionID: faction.ID, Action: selectedAction.Name()}}, actionMutations...)

	var worldIndex *world.Index
	if e.World == nil {
		return ErrWorldEngineUnavailable
	}
```

##### Task 4 — `narrative/digest/build.go`

**Find:**
```go
			case "goal_progressed":
				// dropped — bookkeeping only
```
**Replace with:**
```go
			case "goal_progressed":
				// dropped — bookkeeping only
			case "action_taken":
				// dropped — the action shows through its own mutations (Decision 3)
```

##### Task 5 — `tui/views/turn/execution/format.go`

**Find:**
```go
		// GoalProgressed, AssetMaintainedFlag, AssetStealthApplied, AssetStealthCleared,
```
**Replace with:**
```go
		// ActionTaken (the header names the action), GoalProgressed,
		// AssetMaintainedFlag, AssetStealthApplied, AssetStealthCleared,
```

##### Task 6 — tests

- **`engine/mutation/mutation_test.go`** — `TestApply_ActionTakenChangesNoState`:
  applying `ActionTaken{FactionID: "f1", Action: "Attack"}` returns no error and
  leaves the faction equal to a clone taken before the apply. A second case names a
  faction that does not exist and still returns no miss.
- **`engine/testharness/scenarios/actions_test.go`** —
  `TestRunCycle_ActionBatchOpensWithActionTaken`: alpha takes Repair Faction, beta
  passes. In history, the first mutation of alpha's action record is `action_taken`
  with `action = "Repair Faction"`. No record of beta's turn holds an
  `action_taken`.
- **`tui/views/turn/execution/format_test.go`** — add `domain.ActionTaken{}` to the
  list in `TestFormatMutation_Folded`.
- **`narrative/digest/digest_test.go`** — `TestBuild_ActionTakenLeavesNoNote`: a
  record that holds only an `action_taken` builds a digest in which no beat has a
  note.

Then run the whole suite. A test that reads an action batch by position, or counts
history records, may need its expectation moved by one (Open Questions).

##### Commit message

```
feat(E1): ActionTaken opens each action batch

- a record-only ActionTaken{FactionID, Action} is the first mutation of every
  action batch; a pass or a failed action emits none (D-16, D-19)
- no-op apply arm; the digest drops it and the feed folds it, because the feed
  header already names the action
```

#### Commit 4 — `feat(E1): goal parts load from data`

D-18, Decisions 4 and 6, `B.005`. The IDs stay bare until Commit 7.

##### Task 1 — `domain/goal.go` (new) and `domain/faction.go`

In `faction.go`, delete the `Goal` type:

**Find:**
```go
type Goal struct {
	ID          string `toml:"id"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Difficulty  string `toml:"difficulty"`
}

```
**Replace with:** nothing.

`domain/goal.go` (new) — full contents:
```go
package domain

// Goal is one rulebook goal: its text and the parts that state its behavior.
// Stat names are converted at load. Every other name in a part is a plain string
// that the goal engine checks when it is built (A.001.8 Effort 1 Decision 6).
type Goal struct {
	ID          string
	Name        string
	Description string
	Start       *GoalStart
	Progress    *GoalProgress
	Reset       *GoalEvent
	Threshold   *GoalThreshold
	Complete    *GoalEvent
	XP          *GoalXP
}

// GoalStart is what the pick needs before it may offer the goal.
type GoalStart struct {
	Inputs       []string
	NeedsActions []string
	Gate         *GoalGate
}

type GoalGate struct {
	RatingAboveLastCompletion FactionStat
}

// GoalEvent names an event and the filters a mutation must pass to match it.
type GoalEvent struct {
	Event   string
	Filters GoalFilters
}

// GoalProgress moves progress one of two ways: Event with Amount adds for each
// match, and Measure sets progress to a count read from state.
type GoalProgress struct {
	Event   string
	Amount  string
	Measure string
	Filters GoalFilters
}

// GoalThreshold is the progress that completes the goal: Constant, or the sum of
// the holder's Ratings times Multiplier (1 when it is 0).
type GoalThreshold struct {
	Ratings    []FactionStat
	Multiplier int
	Constant   int
}

type GoalXP struct {
	Formula        string
	Amount         int
	ContestedBonus int
}

const (
	GoalFilterOwner              = "owner"
	GoalFilterCausedBy           = "caused_by"
	GoalFilterFaction            = "faction"
	GoalFilterCategory           = "category"
	GoalFilterRatedAbove         = "rated_above"
	GoalFilterNewWorld           = "new_world"
	GoalFilterTo                 = "to"
	GoalFilterTarget             = "target"
	GoalFilterAction             = "action"
	GoalFilterActionNot          = "action_not"
	GoalFilterGround             = "ground"
	GoalFilterExcludeHeldAtStart = "exclude_held_at_start"
)

// GoalFilters narrows which mutations match a goal event or what a measure
// counts. A zero field is an absent filter; only "= true" filters for the bools.
type GoalFilters struct {
	Owner              string
	CausedBy           string
	Faction            string
	Category           []FactionStat
	RatedAbove         FactionStat
	NewWorld           bool
	To                 string
	Target             string
	Action             []string
	ActionNot          []string
	Ground             string
	ExcludeHeldAtStart bool
}

func (filters GoalFilters) Used() []string {
	var present []string
	if filters.Owner != "" {
		present = append(present, GoalFilterOwner)
	}
	if filters.CausedBy != "" {
		present = append(present, GoalFilterCausedBy)
	}
	if filters.Faction != "" {
		present = append(present, GoalFilterFaction)
	}
	if len(filters.Category) > 0 {
		present = append(present, GoalFilterCategory)
	}
	if filters.RatedAbove != "" {
		present = append(present, GoalFilterRatedAbove)
	}
	if filters.NewWorld {
		present = append(present, GoalFilterNewWorld)
	}
	if filters.To != "" {
		present = append(present, GoalFilterTo)
	}
	if filters.Target != "" {
		present = append(present, GoalFilterTarget)
	}
	if len(filters.Action) > 0 {
		present = append(present, GoalFilterAction)
	}
	if len(filters.ActionNot) > 0 {
		present = append(present, GoalFilterActionNot)
	}
	if filters.Ground != "" {
		present = append(present, GoalFilterGround)
	}
	if filters.ExcludeHeldAtStart {
		present = append(present, GoalFilterExcludeHeldAtStart)
	}
	return present
}
```

##### Task 2 — `rulebook/goals.go` (new) and `rulebook/rulebook.go`

In `rulebook.go`, delete the goal loader. `checkReferences` and its comment stay.

**Find:**
```go
type goalFile struct {
	Goals map[string]goalRecord `toml:"goals"`
}

type goalRecord struct {
	ID          string `toml:"id"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Difficulty  string `toml:"difficulty"`
}

func loadGoals(path string) (map[string]*domain.Goal, error) {
	var f goalFile
	metadata, err := toml.DecodeFile(path, &f)
	if err != nil {
		return nil, err
	}
	if err := refuseUnknownKeys(path, metadata); err != nil {
		return nil, err
	}
	result := make(map[string]*domain.Goal, len(f.Goals))
	for _, r := range f.Goals {
		if _, exists := result[r.ID]; exists {
			return nil, fmt.Errorf("duplicate goal ID %q in %s", r.ID, path)
		}
		result[r.ID] = &domain.Goal{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Difficulty:  r.Difficulty,
		}
	}
	return result, nil
}

```
**Replace with:** nothing.

`rulebook/goals.go` (new) — full contents:
```go
package rulebook

import (
	"fmt"
	"maps"
	"slices"

	"github.com/BurntSushi/toml"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

type goalFile struct {
	Goals map[string]goalRecord `toml:"goals"`
}

type goalRecord struct {
	ID          string               `toml:"id"`
	Name        string               `toml:"name"`
	Description string               `toml:"description"`
	Start       *goalStartRecord     `toml:"start"`
	Progress    *goalProgressRecord  `toml:"progress"`
	Reset       *goalEventRecord     `toml:"reset"`
	Threshold   *goalThresholdRecord `toml:"threshold"`
	Complete    *goalEventRecord     `toml:"complete"`
	XP          *goalXPRecord        `toml:"xp"`
}

type goalStartRecord struct {
	Inputs       []string        `toml:"inputs"`
	NeedsActions []string        `toml:"needs_actions"`
	Gate         *goalGateRecord `toml:"gate"`
}

type goalGateRecord struct {
	RatingAboveLastCompletion string `toml:"rating_above_last_completion"`
}

type goalEventRecord struct {
	Event   string             `toml:"event"`
	Filters *goalFiltersRecord `toml:"filters"`
}

type goalProgressRecord struct {
	Event   string             `toml:"event"`
	Amount  string             `toml:"amount"`
	Measure string             `toml:"measure"`
	Filters *goalFiltersRecord `toml:"filters"`
}

type goalThresholdRecord struct {
	Ratings    []string `toml:"ratings"`
	Multiplier int      `toml:"multiplier"`
	Constant   int      `toml:"constant"`
}

type goalXPRecord struct {
	Formula        string `toml:"formula"`
	Amount         int    `toml:"amount"`
	ContestedBonus int    `toml:"contested_bonus"`
}

type goalFiltersRecord struct {
	Owner              string   `toml:"owner"`
	CausedBy           string   `toml:"caused_by"`
	Faction            string   `toml:"faction"`
	Category           []string `toml:"category"`
	RatedAbove         string   `toml:"rated_above"`
	NewWorld           bool     `toml:"new_world"`
	To                 string   `toml:"to"`
	Target             string   `toml:"target"`
	Action             []string `toml:"action"`
	ActionNot          []string `toml:"action_not"`
	Ground             string   `toml:"ground"`
	ExcludeHeldAtStart bool     `toml:"exclude_held_at_start"`
}

// loadGoals reads the goals strictly and converts their stat names. It does not
// check a goal-only name (an event, a formula): rulebook must not import the goal
// engine, so engine construction checks those (A.001.8 Effort 1 Decision 6).
func loadGoals(path string) (map[string]*domain.Goal, error) {
	var file goalFile
	metadata, err := toml.DecodeFile(path, &file)
	if err != nil {
		return nil, err
	}
	if err := refuseUnknownKeys(path, metadata); err != nil {
		return nil, err
	}
	result := make(map[string]*domain.Goal, len(file.Goals))
	for _, key := range slices.Sorted(maps.Keys(file.Goals)) {
		record := file.Goals[key]
		if _, exists := result[record.ID]; exists {
			return nil, fmt.Errorf("duplicate goal ID %q in %s", record.ID, path)
		}
		goal, err := convertGoal(record)
		if err != nil {
			return nil, fmt.Errorf("goal %q in %s: %w", record.ID, path, err)
		}
		result[record.ID] = goal
	}
	return result, nil
}

func convertGoal(record goalRecord) (*domain.Goal, error) {
	goal := &domain.Goal{ID: record.ID, Name: record.Name, Description: record.Description}
	if start := record.Start; start != nil {
		goal.Start = &domain.GoalStart{Inputs: start.Inputs, NeedsActions: start.NeedsActions}
		if start.Gate != nil {
			rating, err := toFactionStat(start.Gate.RatingAboveLastCompletion)
			if err != nil {
				return nil, fmt.Errorf("start.gate rating_above_last_completion: %w", err)
			}
			goal.Start.Gate = &domain.GoalGate{RatingAboveLastCompletion: rating}
		}
	}
	if progress := record.Progress; progress != nil {
		filters, err := convertGoalFilters(progress.Filters)
		if err != nil {
			return nil, fmt.Errorf("progress: %w", err)
		}
		goal.Progress = &domain.GoalProgress{Event: progress.Event, Amount: progress.Amount, Measure: progress.Measure, Filters: filters}
	}
	reset, err := convertGoalEvent("reset", record.Reset)
	if err != nil {
		return nil, err
	}
	goal.Reset = reset
	if threshold := record.Threshold; threshold != nil {
		ratings, err := toFactionStats(threshold.Ratings)
		if err != nil {
			return nil, fmt.Errorf("threshold ratings%w", err)
		}
		goal.Threshold = &domain.GoalThreshold{Ratings: ratings, Multiplier: threshold.Multiplier, Constant: threshold.Constant}
	}
	complete, err := convertGoalEvent("complete", record.Complete)
	if err != nil {
		return nil, err
	}
	goal.Complete = complete
	if xp := record.XP; xp != nil {
		goal.XP = &domain.GoalXP{Formula: xp.Formula, Amount: xp.Amount, ContestedBonus: xp.ContestedBonus}
	}
	return goal, nil
}

func convertGoalEvent(part string, record *goalEventRecord) (*domain.GoalEvent, error) {
	if record == nil {
		return nil, nil
	}
	filters, err := convertGoalFilters(record.Filters)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", part, err)
	}
	return &domain.GoalEvent{Event: record.Event, Filters: filters}, nil
}

func convertGoalFilters(record *goalFiltersRecord) (domain.GoalFilters, error) {
	if record == nil {
		return domain.GoalFilters{}, nil
	}
	filters := domain.GoalFilters{
		Owner:              record.Owner,
		CausedBy:           record.CausedBy,
		Faction:            record.Faction,
		NewWorld:           record.NewWorld,
		To:                 record.To,
		Target:             record.Target,
		Action:             record.Action,
		ActionNot:          record.ActionNot,
		Ground:             record.Ground,
		ExcludeHeldAtStart: record.ExcludeHeldAtStart,
	}
	category, err := toFactionStats(record.Category)
	if err != nil {
		return domain.GoalFilters{}, fmt.Errorf("filters.category%w", err)
	}
	filters.Category = category
	if record.RatedAbove != "" {
		ratedAbove, err := toFactionStat(record.RatedAbove)
		if err != nil {
			return domain.GoalFilters{}, fmt.Errorf("filters.rated_above: %w", err)
		}
		filters.RatedAbove = ratedAbove
	}
	return filters, nil
}

// toFactionStats converts a list of stat names. Its error starts with the index,
// so the caller's message reads "filters.category[1]: unknown faction stat".
func toFactionStats(names []string) ([]domain.FactionStat, error) {
	if len(names) == 0 {
		return nil, nil
	}
	stats := make([]domain.FactionStat, 0, len(names))
	for i, name := range names {
		stat, err := toFactionStat(name)
		if err != nil {
			return nil, fmt.Errorf("[%d]: %w", i, err)
		}
		stats = append(stats, stat)
	}
	return stats, nil
}
```

Drop the imports `rulebook.go` no longer uses, if the compiler reports any.

##### Task 3 — `rulebooks/_core/goals.toml`

Full contents. Key order in an entry: `id`, `name`, `description`, then the parts in
the order `start`, `progress`, `reset`, `threshold`, `complete`, `xp`.

```toml
[goals.MilitaryConquest]
id = "G-001"
name = "Military Conquest"
description = "Destroy a number of Force assets of rival factions equal to your faction's Force rating."
progress = { event = "kill", amount = "one", filters = { owner = "rival", caused_by = "own", category = ["Force"] } }
threshold = { ratings = ["Force"] }
xp = { formula = "half_progress" }

[goals.CommercialExpansion]
id = "G-002"
name = "Commercial Expansion"
description = "Destroy a number of Wealth assets of rival factions equal to your faction's Wealth rating."
progress = { event = "kill", amount = "one", filters = { owner = "rival", caused_by = "own", category = ["Wealth"] } }
threshold = { ratings = ["Wealth"] }
xp = { formula = "half_progress" }

[goals.IntelligenceCoup]
id = "G-003"
name = "Intelligence Coup"
description = "Destroy a number of Cunning assets of rival factions equal to your faction's Cunning rating."
progress = { event = "kill", amount = "one", filters = { owner = "rival", caused_by = "own", category = ["Cunning"] } }
threshold = { ratings = ["Cunning"] }
xp = { formula = "half_progress" }

[goals.WorldSeizure]
id = "G-004"
name = "World Seizure"
description = "Take control of a world, becoming the world's legitimate government. If no opposing faction resists, counts as difficulty low."
start = { needs_actions = ["Seize Planet"] }
complete = { event = "government_taken", filters = { to = "own" } }
xp = { formula = "ruler_average" }

[goals.ExpandInfluence]
id = "G-005"
name = "Expand Influence"
description = "Plant a Base of Influence on a new world. Difficulty increases by one if the attempt is contested by a rival faction."
complete = { event = "base_planted", filters = { caused_by = "own", new_world = true } }
xp = { formula = "constant", amount = 1, contested_bonus = 1 }

[goals.BloodTheEnemy]
id = "G-006"
name = "Blood the Enemy"
description = "Inflict a number of hit points of damage on enemy faction assets or bases equal to your faction's total Force, Cunning, and Wealth ratings."
progress = { event = "damage", amount = "damage", filters = { owner = "rival", caused_by = "own" } }
threshold = { ratings = ["Force", "Cunning", "Wealth"] }
xp = { formula = "constant", amount = 2 }

[goals.PeaceableKingdom]
id = "G-007"
name = "Peaceable Kingdom"
description = "Do not take an Attack action for four consecutive turns."
progress = { event = "action_taken", amount = "one", filters = { faction = "own", action_not = ["Attack"] } }
reset = { event = "action_taken", filters = { faction = "own", action = ["Attack"] } }
threshold = { constant = 4 }
xp = { formula = "constant", amount = 1 }

[goals.DestroyTheFoe]
id = "G-008"
name = "Destroy the Foe"
description = "Destroy a rival faction entirely."
start = { inputs = ["target_faction"] }
complete = { event = "faction_destroyed", filters = { target = "start_input", caused_by = "own" } }
xp = { formula = "target_average" }

[goals.InsideEnemyTerritory]
id = "G-009"
name = "Inside Enemy Territory"
description = "Have a number of stealthed assets on worlds with other governments equal to your Cunning score. Units already stealthed when this goal is adopted do not count."
progress = { measure = "stealthed_assets", filters = { ground = "rival", exclude_held_at_start = true } }
threshold = { ratings = ["Cunning"] }
xp = { formula = "constant", amount = 2 }

[goals.InvincibleValor]
id = "G-010"
name = "Invincible Valor"
description = "Destroy a Force asset with a minimum purchase rating higher than your faction's Force rating."
complete = {
  event = "kill",
  filters = { owner = "rival", caused_by = "own", category = ["Force"], rated_above = "Force" },
}
xp = { formula = "constant", amount = 2 }

[goals.WealthOfWorlds]
id = "G-011"
name = "Wealth of Worlds"
description = "Spend Coins equal to four times your faction's Wealth rating on bribes and influence. The faction's Wealth rating must increase before this goal can be selected again."
start = { gate = { rating_above_last_completion = "Wealth" } }
progress = { event = "influence_bought", amount = "delta", filters = { caused_by = "own" } }
threshold = { ratings = ["Wealth"], multiplier = 4 }
xp = { formula = "constant", amount = 2 }
```

Run Format Document on the file and keep what it lays out.

##### Task 4 — `rulebook/rulebook_test.go`

- In the "goals populated" subtest, replace the `Difficulty` assertion: `G-001`
  loads with `Progress` equal to
  `&domain.GoalProgress{Event: "kill", Amount: "one", Filters: domain.GoalFilters{Owner: "rival", CausedBy: "own", Category: []domain.FactionStat{domain.StatForce}}}`,
  `Threshold` equal to `&domain.GoalThreshold{Ratings: []domain.FactionStat{domain.StatForce}}`,
  `XP` equal to `&domain.GoalXP{Formula: "half_progress"}`, and `Start`, `Reset` and
  `Complete` nil.
- In the strict-loader table, the "goal unknown key" row keeps its body
  (`dificulty` is still unknown). Add these rows, each with `load: loadGoalFile`:
  - "goal old difficulty key": a goal with `difficulty = "low"` fails with
    `unknown keys in`.
  - "goal unknown key inside a part": `xp = { formula = "constant", amout = 1 }`
    fails with `unknown keys in`.
  - "goal unknown filter key": `complete = { event = "kill", filters = { ownr = "rival" } }`
    fails with `unknown keys in`.
  - "goal bad category stat": `filters = { category = ["force"] }` on `progress`
    fails with `progress: filters.category[0]: unknown faction stat`.
  - "goal bad rated_above stat": fails with `complete: filters.rated_above: unknown faction stat`.
  - "goal bad threshold rating": `threshold = { ratings = ["Might"] }` fails with
    `threshold ratings[0]: unknown faction stat`.
  - "goal bad gate rating": `start = { gate = { rating_above_last_completion = "" } }`
    fails with `start.gate rating_above_last_completion: unknown faction stat`.

##### Task 5 — `rulebook/goals_test.go` (new)

- `TestLoad_ShippedGoals`: the shipped rulebook (`loadShippedRulebook`) holds
  exactly eleven goals. Table by goal ID with the parts that are set:
  `G-004` has `Start.NeedsActions == ["Seize Planet"]`, a `Complete` on
  `government_taken` with `To == "own"`, and no `Progress` or `Threshold`; `G-007`
  has a `Reset` on `action_taken` with `Action == ["Attack"]` and
  `Threshold.Constant == 4`; `G-008` has `Start.Inputs == ["target_faction"]`;
  `G-009` has `Progress.Measure == "stealthed_assets"` with `Ground == "rival"` and
  `ExcludeHeldAtStart`; `G-010` has `Complete.Filters.RatedAbove == domain.StatForce`;
  `G-011` has `Start.Gate.RatingAboveLastCompletion == domain.StatWealth` and
  `Threshold.Multiplier == 4`.
- `TestGoalFilters_Used`: a `GoalFilters` with every field set reports all twelve
  names; a zero value reports none; `NewWorld: false` and
  `ExcludeHeldAtStart: false` report nothing. This test may live in
  `domain/goal_test.go` instead.

##### Task 6 — other users of `Goal.Difficulty`

`grep -rn "Difficulty" internal/ cmd/` must return nothing. The wizard and the
detail view read only `ID`, `Name` and `Description`.

##### Commit message

```
feat(E1): goal parts load from data

- each goal in _core/goals.toml states start, progress, reset, threshold,
  complete and xp as inline tables; difficulty is gone (B.005)
- the goal loader reads the parts strictly and converts stat names; goal-only
  names stay strings for the goal engine to check (D-18)
- domain.Goal carries the parts; the handlers still run the goals
```

### Phase 3 — Matchers

**In plain words:** a new goal engine turns a goal's parts into checks and matchers,
and works out progress, completion and XP from a batch of mutations. When this lands
it is fully tested against the shipped goals, and nothing calls it yet.

#### Commit 5 — `feat(E1): a goal engine that reads goal parts`

The Goal Vocabulary of the discovery, with Decisions 1, 2, 7 to 9, 11 and 13. The new
files sit beside `goal.go` and the handlers; Commit 6 switches over.

##### Task 1 — `domain/faction.go`

The start record that `exclude_held_at_start` reads (D-10). Effort 3 fills it.

**Find:**
```go
type ActiveGoal struct {
	GoalID          string `toml:"goal_id"`
	TargetFactionID string `toml:"target_faction_id"`
	Progress        int    `toml:"progress"`
}
```
**Replace with:**
```go
type ActiveGoal struct {
	GoalID          string `toml:"goal_id"`
	TargetFactionID string `toml:"target_faction_id"`
	Progress        int    `toml:"progress"`
	// StealthedAtStart holds the IDs of the assets already stealthed when the goal
	// was picked; a measure with exclude_held_at_start never counts them (D-10).
	StealthedAtStart []string `toml:"stealthed_at_start,omitempty"`
}
```
**Find:**
```go
		goalCopy := *faction.ActiveGoal
		clone.ActiveGoal = &goalCopy
```
**Replace with:**
```go
		goalCopy := *faction.ActiveGoal
		goalCopy.StealthedAtStart = slices.Clone(faction.ActiveGoal.StealthedAtStart)
		clone.ActiveGoal = &goalCopy
```

##### Task 2 — `domain/mutation.go`

The resist record that `ruler_average` reads (D-8). `F.040` fills it.

**Find:**
```go
// WorldGovernmentChanged sets the faction that governs a world. An empty
// ToFactionID makes the world independent; FromFactionID is the ruler it replaces,
// empty for an independent world.
type WorldGovernmentChanged struct {
	WorldID           string `json:"world_id"`
	FromFactionID     string `json:"from_faction_id"`
	ToFactionID       string `json:"to_faction_id"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}
```
**Replace with:**
```go
// WorldGovernmentChanged sets the faction that governs a world. An empty
// ToFactionID makes the world independent; FromFactionID is the ruler it replaces,
// empty for an independent world. Resisted says the replaced ruler had an
// unstealthed asset on the world when the seizure began (D-8).
type WorldGovernmentChanged struct {
	WorldID           string `json:"world_id"`
	FromFactionID     string `json:"from_faction_id"`
	ToFactionID       string `json:"to_faction_id"`
	Resisted          bool   `json:"resisted,omitempty"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}
```

##### Task 3 — `engine/goal/vocabulary.go` (new)

Full contents:
```go
package goal

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

const (
	sideOwn   = "own"
	sideRival = "rival"

	targetStartInput   = "start_input"
	inputTargetFaction = "target_faction"
	groundRival        = "rival"

	amountOne    = "one"
	amountDamage = "damage"
	amountDelta  = "delta"

	eventGovernmentTaken = "government_taken"

	formulaConstant      = "constant"
	formulaHalfProgress  = "half_progress"
	formulaTargetAverage = "target_average"
	formulaRulerAverage  = "ruler_average"
)

// evalContext is one faction's goal read against one batch. factionState is the
// state before the batch applies.
type evalContext struct {
	faction      *domain.Faction
	batch        []domain.Mutation
	factionState *state.FactionState
	rulebook     *rulebook.Rulebook
}

// match is one mutation that passed an event's filters. world is empty for an
// event with no world; mutation is set only where an XP formula reads it.
type match struct {
	amount   int
	world    string
	mutation domain.Mutation
}

// eventSpec is one goal event: the filters and amounts the data may give it, and
// the matcher over a batch. hasWorld says its matches name a world.
type eventSpec struct {
	filters  []string
	amounts  []string
	hasWorld bool
	match    func(ctx evalContext, filters domain.GoalFilters) []match
}

type measureSpec struct {
	filters []string
	count   func(ctx evalContext, filters domain.GoalFilters) int
}

var events = map[string]eventSpec{
	"kill": {
		filters:  []string{domain.GoalFilterOwner, domain.GoalFilterCausedBy, domain.GoalFilterCategory, domain.GoalFilterRatedAbove},
		amounts:  []string{amountOne},
		hasWorld: true,
		match:    matchKill,
	},
	"damage": {
		filters: []string{domain.GoalFilterOwner, domain.GoalFilterCausedBy},
		amounts: []string{amountDamage},
		match:   matchDamage,
	},
	"base_planted": {
		filters:  []string{domain.GoalFilterCausedBy, domain.GoalFilterNewWorld},
		amounts:  []string{amountOne},
		hasWorld: true,
		match:    matchBasePlanted,
	},
	"influence_bought": {
		filters: []string{domain.GoalFilterCausedBy},
		amounts: []string{amountDelta},
		match:   matchInfluenceBought,
	},
	eventGovernmentTaken: {
		filters:  []string{domain.GoalFilterTo},
		amounts:  []string{amountOne},
		hasWorld: true,
		match:    matchGovernmentTaken,
	},
	"faction_destroyed": {
		filters: []string{domain.GoalFilterTarget, domain.GoalFilterCausedBy},
		amounts: []string{amountOne},
		match:   matchFactionDestroyed,
	},
	"action_taken": {
		filters: []string{domain.GoalFilterFaction, domain.GoalFilterAction, domain.GoalFilterActionNot},
		amounts: []string{amountOne},
		match:   matchActionTaken,
	},
}

var measures = map[string]measureSpec{
	"stealthed_assets": {
		filters: []string{domain.GoalFilterGround, domain.GoalFilterExcludeHeldAtStart},
		count:   countStealthedAssets,
	},
}

// sideMatches reads an own/rival filter against the goal's holder. An absent
// filter matches every faction.
func sideMatches(want, factionID, holderID string) bool {
	switch want {
	case sideOwn:
		return factionID == holderID
	case sideRival:
		return factionID != "" && factionID != holderID
	default:
		return true
	}
}
```

##### Task 4 — `engine/goal/events.go` (new)

Full contents:
```go
package goal

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// matchKill reads the removed asset from the state before the batch, where it
// still exists. Any cause counts when the holder caused it (D-14).
func matchKill(ctx evalContext, filters domain.GoalFilters) []match {
	var matches []match
	for _, mutation := range ctx.batch {
		removed, ok := mutation.(domain.AssetRemoved)
		if !ok || !sideMatches(filters.Owner, removed.FactionID, ctx.faction.ID) || !sideMatches(filters.CausedBy, removed.CausedByFactionID, ctx.faction.ID) {
			continue
		}
		owner := ctx.factionState.Factions[removed.FactionID]
		if owner == nil || owner.Assets[removed.AssetID] == nil {
			continue
		}
		asset := owner.Assets[removed.AssetID]
		definition := ctx.rulebook.Assets[asset.DefinitionID]
		if definition == nil {
			continue
		}
		if len(filters.Category) > 0 && !slices.Contains(filters.Category, definition.Category) {
			continue
		}
		if filters.RatedAbove != "" && definition.MinRating <= ctx.faction.StatScore(filters.RatedAbove) {
			continue
		}
		matches = append(matches, match{amount: 1, world: asset.Location.WorldID})
	}
	return matches
}

func matchDamage(ctx evalContext, filters domain.GoalFilters) []match {
	var matches []match
	for _, mutation := range ctx.batch {
		var ownerID, causedByID string
		var delta int
		// deliberate: FactionHPDelta is not damage here, because every Base hit emits one beside its BaseHPDelta and would count twice (Decision 1).
		switch hit := mutation.(type) {
		case domain.AssetHPDelta:
			ownerID, causedByID, delta = hit.FactionID, hit.CausedByFactionID, hit.Delta
		case domain.BaseHPDelta:
			ownerID, causedByID, delta = hit.FactionID, hit.CausedByFactionID, hit.Delta
		default:
			continue
		}
		if delta >= 0 || !sideMatches(filters.Owner, ownerID, ctx.faction.ID) || !sideMatches(filters.CausedBy, causedByID, ctx.faction.ID) {
			continue
		}
		matches = append(matches, match{amount: -delta})
	}
	return matches
}

func matchBasePlanted(ctx evalContext, filters domain.GoalFilters) []match {
	var matches []match
	for _, mutation := range ctx.batch {
		added, ok := mutation.(domain.BaseAdded)
		if !ok || !sideMatches(filters.CausedBy, added.CausedByFactionID, ctx.faction.ID) {
			continue
		}
		world := added.Base.Location.WorldID
		if filters.NewWorld && ctx.faction.HasBaseOn(world) {
			continue
		}
		matches = append(matches, match{amount: 1, world: world})
	}
	return matches
}

func matchInfluenceBought(ctx evalContext, filters domain.GoalFilters) []match {
	var matches []match
	for _, mutation := range ctx.batch {
		bought, ok := mutation.(domain.InfluenceDelta)
		if !ok || bought.Delta <= 0 || !sideMatches(filters.CausedBy, bought.CausedByFactionID, ctx.faction.ID) {
			continue
		}
		matches = append(matches, match{amount: bought.Delta})
	}
	return matches
}

func matchGovernmentTaken(ctx evalContext, filters domain.GoalFilters) []match {
	var matches []match
	for _, mutation := range ctx.batch {
		changed, ok := mutation.(domain.WorldGovernmentChanged)
		if !ok || !sideMatches(filters.To, changed.ToFactionID, ctx.faction.ID) {
			continue
		}
		matches = append(matches, match{amount: 1, world: changed.WorldID, mutation: changed})
	}
	return matches
}

// matchFactionDestroyed matches the one HP loss that takes a faction to 0 or
// less, measured against its HP before the batch plus the batch's earlier losses.
// caused_by reads that loss, so a third party's last blow does not match (D-15).
func matchFactionDestroyed(ctx evalContext, filters domain.GoalFilters) []match {
	var matches []match
	lost := make(map[string]int)
	for _, mutation := range ctx.batch {
		delta, ok := mutation.(domain.FactionHPDelta)
		if !ok {
			continue
		}
		target := ctx.factionState.Factions[delta.FactionID]
		if target == nil {
			continue
		}
		before := target.CurrentHP + lost[delta.FactionID]
		lost[delta.FactionID] += delta.Delta
		if before <= 0 || before+delta.Delta > 0 {
			continue
		}
		if filters.Target == targetStartInput && delta.FactionID != ctx.faction.ActiveGoal.TargetFactionID {
			continue
		}
		if !sideMatches(filters.CausedBy, delta.CausedByFactionID, ctx.faction.ID) {
			continue
		}
		matches = append(matches, match{amount: 1, mutation: delta})
	}
	return matches
}

func matchActionTaken(ctx evalContext, filters domain.GoalFilters) []match {
	var matches []match
	for _, mutation := range ctx.batch {
		taken, ok := mutation.(domain.ActionTaken)
		if !ok || !sideMatches(filters.Faction, taken.FactionID, ctx.faction.ID) {
			continue
		}
		if len(filters.Action) > 0 && !slices.Contains(filters.Action, taken.Action) {
			continue
		}
		if slices.Contains(filters.ActionNot, taken.Action) {
			continue
		}
		matches = append(matches, match{amount: 1})
	}
	return matches
}
```

##### Task 5 — `engine/goal/measures.go` (new)

Decision 2. `isRivalGround` is the old `isEnemyTerritory`, moved here with an
empty-world guard; the handler's copy goes with the handlers in Commit 6.

Full contents:
```go
package goal

import (
	"log/slog"
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/mutation"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// countStealthedAssets counts the holder's stealthed assets in the state the
// batch leaves (Decision 2, D-10). An asset in transit has no world, so it is on
// no one's ground.
func countStealthedAssets(ctx evalContext, filters domain.GoalFilters) int {
	projected := project(ctx.factionState, ctx.batch)
	holder := projected.Factions[ctx.faction.ID]
	if holder == nil {
		return 0
	}
	count := 0
	for _, asset := range holder.Assets {
		if !asset.Stealthy {
			continue
		}
		if filters.Ground == groundRival && !isRivalGround(holder.ID, asset.Location.WorldID, projected) {
			continue
		}
		if filters.ExcludeHeldAtStart && slices.Contains(ctx.faction.ActiveGoal.StealthedAtStart, asset.ID) {
			continue
		}
		count++
	}
	return count
}

// project returns a copy of factionState with batch applied. The goal engine
// runs before the batch applies, and a measure counts what the batch leaves.
func project(factionState *state.FactionState, batch []domain.Mutation) *state.FactionState {
	projected := &state.FactionState{
		CampaignID:  factionState.CampaignID,
		CycleNumber: factionState.CycleNumber,
		Factions:    make(map[string]*domain.Faction, len(factionState.Factions)),
		Governments: maps.Clone(factionState.Governments),
		CurrentTurn: factionState.CurrentTurn,
	}
	for factionID, faction := range factionState.Factions {
		projected.Factions[factionID] = faction.Clone()
	}
	discard := slog.New(slog.DiscardHandler)
	// deliberate: a miss on the copy is dropped, not reported, because the orchestrator's own apply of this batch reports it.
	_ = mutation.New(discard).Apply(projected, batch, discard)
	return projected
}

// isRivalGround reports whether worldID is a rival's ground for factionID: a
// rival governs it, or it is independent and a rival's homeworld. A world the
// faction governs never is, a seized rival homeworld included.
func isRivalGround(factionID, worldID string, factionState *state.FactionState) bool {
	if worldID == "" {
		return false
	}
	if !factionState.IsIndependent(worldID) {
		return factionState.Ruler(worldID) != factionID
	}
	for otherID, other := range factionState.Factions {
		if otherID != factionID && other.Homeworld.WorldID == worldID {
			return true
		}
	}
	return false
}
```

`CurrentTurn` is shared with the original, not copied: no apply arm writes it.

##### Task 6 — `engine/goal/compile.go` (new)

Full contents:
```go
package goal

import (
	"fmt"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// compiledGoal is a goal whose every name is known: each part is bound to the
// spec that runs it.
type compiledGoal struct {
	goal     *domain.Goal
	progress *eventSpec
	measure  *measureSpec
	reset    *eventSpec
	complete *eventSpec
}

// compile checks every name and combination in goal's parts and binds them. Its
// errors do not name the goal; the caller does.
func compile(goal *domain.Goal) (*compiledGoal, error) {
	compiled := &compiledGoal{goal: goal}
	if err := checkStart(goal.Start); err != nil {
		return nil, err
	}
	if (goal.Threshold == nil) == (goal.Complete == nil) {
		return nil, fmt.Errorf("a goal needs exactly one of threshold or complete")
	}
	if goal.Threshold != nil {
		if goal.Progress == nil {
			return nil, fmt.Errorf("threshold needs a progress part")
		}
		if err := checkThreshold(goal.Threshold); err != nil {
			return nil, err
		}
	}
	if goal.Reset != nil && goal.Progress == nil {
		return nil, fmt.Errorf("reset needs a progress part")
	}
	if progress := goal.Progress; progress != nil {
		if (progress.Event == "") == (progress.Measure == "") {
			return nil, fmt.Errorf("progress needs exactly one of event or measure")
		}
		if progress.Event != "" {
			spec, err := eventNamed("progress", progress.Event, progress.Filters, goal)
			if err != nil {
				return nil, err
			}
			if !slices.Contains(spec.amounts, progress.Amount) {
				return nil, fmt.Errorf("progress: event %q does not accept amount %q (want one of %v)", progress.Event, progress.Amount, spec.amounts)
			}
			compiled.progress = spec
		} else {
			spec, ok := measures[progress.Measure]
			if !ok {
				return nil, fmt.Errorf("progress: unknown measure %q", progress.Measure)
			}
			if progress.Amount != "" {
				return nil, fmt.Errorf("progress: measure %q takes no amount", progress.Measure)
			}
			if err := checkFilters("progress", spec.filters, progress.Filters, goal); err != nil {
				return nil, err
			}
			compiled.measure = &spec
		}
	}
	if goal.Reset != nil {
		spec, err := eventNamed("reset", goal.Reset.Event, goal.Reset.Filters, goal)
		if err != nil {
			return nil, err
		}
		compiled.reset = spec
	}
	if goal.Complete != nil {
		spec, err := eventNamed("complete", goal.Complete.Event, goal.Complete.Filters, goal)
		if err != nil {
			return nil, err
		}
		compiled.complete = spec
	}
	if err := checkXP(goal); err != nil {
		return nil, err
	}
	return compiled, nil
}

func eventNamed(part, name string, filters domain.GoalFilters, goal *domain.Goal) (*eventSpec, error) {
	spec, ok := events[name]
	if !ok {
		return nil, fmt.Errorf("%s: unknown event %q", part, name)
	}
	if err := checkFilters(part, spec.filters, filters, goal); err != nil {
		return nil, err
	}
	return &spec, nil
}

func checkFilters(part string, accepted []string, filters domain.GoalFilters, goal *domain.Goal) error {
	for _, name := range filters.Used() {
		if !slices.Contains(accepted, name) {
			return fmt.Errorf("%s: filter %q is not accepted here (want one of %v)", part, name, accepted)
		}
	}
	values := []struct {
		name    string
		value   string
		allowed []string
	}{
		{domain.GoalFilterOwner, filters.Owner, []string{sideOwn, sideRival}},
		{domain.GoalFilterCausedBy, filters.CausedBy, []string{sideOwn, sideRival}},
		{domain.GoalFilterFaction, filters.Faction, []string{sideOwn}},
		{domain.GoalFilterTo, filters.To, []string{sideOwn}},
		{domain.GoalFilterTarget, filters.Target, []string{targetStartInput}},
		{domain.GoalFilterGround, filters.Ground, []string{groundRival}},
	}
	for _, check := range values {
		if check.value != "" && !slices.Contains(check.allowed, check.value) {
			return fmt.Errorf("%s: filter %s = %q (want one of %v)", part, check.name, check.value, check.allowed)
		}
	}
	if filters.Target == targetStartInput && !hasInput(goal, inputTargetFaction) {
		return fmt.Errorf("%s: filter target = %q needs start.inputs to name %q", part, targetStartInput, inputTargetFaction)
	}
	return nil
}

func checkStart(start *domain.GoalStart) error {
	if start == nil {
		return nil
	}
	for _, input := range start.Inputs {
		if input != inputTargetFaction {
			return fmt.Errorf("start: unknown input %q (want %q)", input, inputTargetFaction)
		}
	}
	return nil
}

func hasInput(goal *domain.Goal, input string) bool {
	return goal.Start != nil && slices.Contains(goal.Start.Inputs, input)
}

func checkThreshold(threshold *domain.GoalThreshold) error {
	if threshold.Constant < 0 || threshold.Multiplier < 0 {
		return fmt.Errorf("threshold: constant and multiplier must be >= 0, got %d and %d", threshold.Constant, threshold.Multiplier)
	}
	if (len(threshold.Ratings) > 0) == (threshold.Constant > 0) {
		return fmt.Errorf("threshold needs exactly one of ratings or constant")
	}
	if threshold.Multiplier > 0 && len(threshold.Ratings) == 0 {
		return fmt.Errorf("threshold: multiplier needs ratings")
	}
	return nil
}

func checkXP(goal *domain.Goal) error {
	xp := goal.XP
	if xp == nil {
		return fmt.Errorf("a goal needs an xp part")
	}
	if xp.Formula != formulaConstant && (xp.Amount != 0 || xp.ContestedBonus != 0) {
		return fmt.Errorf("xp: formula %q takes no amount or contested_bonus", xp.Formula)
	}
	switch xp.Formula {
	case formulaConstant:
		if xp.Amount <= 0 || xp.ContestedBonus < 0 {
			return fmt.Errorf("xp: constant needs amount > 0 and contested_bonus >= 0, got %d and %d", xp.Amount, xp.ContestedBonus)
		}
		if xp.ContestedBonus > 0 && (goal.Complete == nil || !events[goal.Complete.Event].hasWorld) {
			return fmt.Errorf("xp: contested_bonus needs a complete event that names a world")
		}
	case formulaHalfProgress:
		if goal.Threshold == nil {
			return fmt.Errorf("xp: %s needs a threshold", formulaHalfProgress)
		}
	case formulaTargetAverage:
		if !hasInput(goal, inputTargetFaction) {
			return fmt.Errorf("xp: %s needs start.inputs to name %q", formulaTargetAverage, inputTargetFaction)
		}
	case formulaRulerAverage:
		if goal.Complete == nil || goal.Complete.Event != eventGovernmentTaken {
			return fmt.Errorf("xp: %s needs complete.event = %q", formulaRulerAverage, eventGovernmentTaken)
		}
	default:
		return fmt.Errorf("xp: unknown formula %q", xp.Formula)
	}
	return nil
}
```

##### Task 7 — `engine/goal/evaluate.go` (new)

Full contents:
```go
package goal

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

const (
	causeProgress  = "goal_progress"
	causeReset     = "goal_reset"
	causeCompleted = "goal_completed"
)

// evaluate reads one batch for the holder's active goal and returns the goal
// mutations it causes. It must run before the batch applies: the matchers read
// about-to-be-removed assets from live state, and XP is worked out from the state
// before the loss (frozen log D119).
func (compiled *compiledGoal) evaluate(ctx evalContext) []domain.Mutation {
	goal := compiled.goal
	progress := ctx.faction.ActiveGoal.Progress
	if compiled.reset != nil && len(compiled.reset.match(ctx, goal.Reset.Filters)) > 0 {
		// deliberate: a reset ends the batch for this goal, not only zeroes it, because progress counted in the same batch would outlive the reset (Decision 8).
		if progress == 0 {
			return nil
		}
		return []domain.Mutation{compiled.progressed(ctx, -progress, causeReset)}
	}
	var mutations []domain.Mutation
	delta := compiled.progressDelta(ctx, progress)
	if delta != 0 {
		mutations = append(mutations, compiled.progressed(ctx, delta, causeProgress))
		progress += delta
	}
	completing, completed := compiled.completion(ctx, progress, delta != 0)
	if !completed {
		return mutations
	}
	xp := compiled.xp(ctx, progress, completing)
	return append(mutations,
		domain.GoalCompleted{FactionID: ctx.faction.ID, GoalID: goal.ID, XPAwarded: xp, Cause: causeCompleted},
		domain.XPAwarded{FactionID: ctx.faction.ID, Amount: xp, Cause: causeCompleted},
	)
}

func (compiled *compiledGoal) progressed(ctx evalContext, delta int, cause string) domain.GoalProgressed {
	return domain.GoalProgressed{FactionID: ctx.faction.ID, GoalID: compiled.goal.ID, Delta: delta, Cause: cause}
}

// progressDelta is how far this batch moves progress: the sum of an event's
// matches, or the gap between a measure's count and the progress held.
func (compiled *compiledGoal) progressDelta(ctx evalContext, progress int) int {
	part := compiled.goal.Progress
	switch {
	case compiled.measure != nil:
		return compiled.measure.count(ctx, part.Filters) - progress
	case compiled.progress != nil:
		matches := compiled.progress.match(ctx, part.Filters)
		if part.Amount == amountOne {
			return len(matches)
		}
		total := 0
		for _, matched := range matches {
			total += matched.amount
		}
		return total
	default:
		return 0
	}
}

// completion reports whether the goal completes on this batch, and the match
// that completed it when a complete event did. A threshold completes only in a
// batch that moved progress (Decision 13).
func (compiled *compiledGoal) completion(ctx evalContext, progress int, moved bool) (match, bool) {
	if compiled.complete != nil {
		matches := compiled.complete.match(ctx, compiled.goal.Complete.Filters)
		if len(matches) == 0 {
			return match{}, false
		}
		return matches[0], true
	}
	return match{}, moved && progress >= thresholdOf(compiled.goal.Threshold, ctx.faction)
}

func thresholdOf(threshold *domain.GoalThreshold, faction *domain.Faction) int {
	if threshold.Constant > 0 {
		return threshold.Constant
	}
	total := 0
	for _, stat := range threshold.Ratings {
		total += faction.StatScore(stat)
	}
	if threshold.Multiplier > 0 {
		total *= threshold.Multiplier
	}
	return total
}
```

##### Task 8 — `engine/goal/xp.go` (new)

D-8, D-9 and D-22. Every formula rounds up.

Full contents:
```go
package goal

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// xp is the XP a completion pays. progress is the progress after this batch;
// completing is the match of the complete event, zero for a threshold.
func (compiled *compiledGoal) xp(ctx evalContext, progress int, completing match) int {
	spec := compiled.goal.XP
	switch spec.Formula {
	case formulaHalfProgress:
		return ceilDiv(progress, 2)
	case formulaTargetAverage:
		target := ctx.factionState.Factions[ctx.faction.ActiveGoal.TargetFactionID]
		if target == nil {
			return 1
		}
		return 1 + ceilDiv(ratingSum(target), 3)
	case formulaRulerAverage:
		changed, _ := completing.mutation.(domain.WorldGovernmentChanged)
		ruler := ctx.factionState.Factions[changed.FromFactionID]
		if ruler == nil || !changed.Resisted {
			return 1
		}
		return max(1, ceilDiv(ratingSum(ruler), 6))
	default:
		amount := spec.Amount
		if spec.ContestedBonus > 0 && rivalPresentOn(completing.world, ctx.faction.ID, ctx.factionState) {
			amount += spec.ContestedBonus
		}
		return amount
	}
}

func ceilDiv(value, divisor int) int {
	return (value + divisor - 1) / divisor
}

func ratingSum(faction *domain.Faction) int {
	return faction.Force + faction.Cunning + faction.Wealth
}

// rivalPresentOn reports whether a faction other than factionID has an asset or
// a Base on worldID (D-22). It reads live state, not the world index, which is
// rebuilt only at turn setup (Decision 9).
func rivalPresentOn(worldID, factionID string, factionState *state.FactionState) bool {
	if worldID == "" {
		return false
	}
	for otherID, other := range factionState.Factions {
		if otherID == factionID {
			continue
		}
		if other.HasBaseOn(worldID) {
			return true
		}
		for _, asset := range other.Assets {
			if asset.Location.WorldID == worldID {
				return true
			}
		}
	}
	return false
}
```

##### Task 9 — `engine/goal/compile_test.go` (new)

One table test, `TestCompile_Refuses`. Each row builds one `domain.Goal` in Go and
names the error text it wants. A helper `validThresholdGoal()` returns a goal with
`progress` on `kill` / `one`, `threshold = {constant = 2}` and
`xp = {formula = "constant", amount = 1}`; each row changes one thing.

| Row | Change | Error contains |
|---|---|---|
| neither end | no `threshold`, no `complete` | `exactly one of threshold or complete` |
| both ends | `threshold` and `complete` | `exactly one of threshold or complete` |
| threshold with no progress | drop `progress` | `threshold needs a progress part` |
| reset with no progress | `complete` goal with `reset` | `reset needs a progress part` |
| event and measure | both set on `progress` | `exactly one of event or measure` |
| neither event nor measure | empty `progress` | `exactly one of event or measure` |
| unknown progress event | `event = "kil"` | `progress: unknown event "kil"` |
| unknown reset event | | `reset: unknown event` |
| unknown complete event | | `complete: unknown event` |
| wrong amount | `kill` with `amount = "damage"` | `does not accept amount "damage"` |
| missing amount | `kill` with no amount | `does not accept amount ""` |
| unknown measure | `measure = "assets"` | `unknown measure "assets"` |
| measure with amount | | `takes no amount` |
| filter not accepted by event | `damage` with `category` | `filter "category" is not accepted here` |
| filter not accepted by measure | `stealthed_assets` with `owner` | `filter "owner" is not accepted here` |
| bad owner value | `owner = "enemy"` | `filter owner = "enemy"` |
| bad caused_by value | | `filter caused_by =` |
| bad faction value | `faction = "rival"` | `filter faction = "rival"` |
| bad to value | | `filter to =` |
| bad target value | | `filter target =` |
| bad ground value | `ground = "own"` | `filter ground = "own"` |
| target with no start input | `target = "start_input"`, no `start` | `needs start.inputs to name "target_faction"` |
| unknown start input | `inputs = ["target_world"]` | `start: unknown input "target_world"` |
| threshold with both | `ratings` and `constant` | `exactly one of ratings or constant` |
| threshold with neither | | `exactly one of ratings or constant` |
| negative constant | | `must be >= 0` |
| multiplier with constant | `constant = 2, multiplier = 3` | `multiplier needs ratings` |
| no xp | | `needs an xp part` |
| unknown formula | `formula = "double"` | `xp: unknown formula "double"` |
| constant with no amount | | `constant needs amount > 0` |
| amount on another formula | `half_progress` with `amount = 1` | `takes no amount or contested_bonus` |
| contested_bonus on a threshold goal | | `contested_bonus needs a complete event that names a world` |
| contested_bonus on an event with no world | `complete` on `faction_destroyed` | same |
| half_progress with no threshold | on a `complete` goal | `half_progress needs a threshold` |
| target_average with no input | | `target_average needs start.inputs` |
| ruler_average on another event | `complete` on `kill` | `ruler_average needs complete.event` |

`TestCompile_AcceptsShippedGoals`: every goal of the shipped rulebook compiles. The
helper `shippedRulebook(t)` composes `rulebooks/swn` into a temp campaign and loads
it; its body:

```go
func shippedRulebook(t *testing.T) *rulebook.Rulebook {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(file), "..", "..", "..", "..")
	camp := &campaigns.Campaign{Root: t.TempDir()}
	if _, err := campaigns.ComposeRulebook(camp, filepath.Join(repoRoot, "rulebooks", "swn"), false); err != nil {
		t.Fatalf("compose rulebook: %v", err)
	}
	loaded, err := rulebook.Load(camp.Paths().FactionDataDir)
	if err != nil {
		t.Fatalf("rulebook.Load: %v", err)
	}
	return loaded
}

func shippedGoal(t *testing.T, loaded *rulebook.Rulebook, name string) *compiledGoal {
	t.Helper()
	for _, goal := range loaded.Goals {
		if goal.Name == name {
			compiled, err := compile(goal)
			if err != nil {
				t.Fatalf("compile %s: %v", name, err)
			}
			return compiled
		}
	}
	t.Fatalf("no shipped goal named %q", name)
	return nil
}
```

Tests name a shipped goal by its display name, so Commit 7's ID change does not
touch them.

##### Task 10 — `engine/goal/evaluate_test.go` (new)

Each test compiles a shipped goal with `shippedGoal`, builds a small state and a
batch in Go, and calls `evaluate`. Helpers: `holder(goalName string, progress int)`
gives faction `f1` with ratings Force 3, Cunning 3, Wealth 2 and the goal active
under the ID the rulebook gives it; `rival()` gives `f2`; `testRulebook` adds asset
definitions `force-def` (Force, `MinRating` 1), `force-high` (Force, `MinRating` 5),
`wealth-def` and `cunning-def` to the shipped rulebook.

**Kills (`G-001` to `G-003`, `G-010`), D-9 and D-14**

- Military Conquest: one Force kill caused by `f1` gives `GoalProgressed{Delta: 1, Cause: "goal_progress"}` and no completion at Force 3.
- A kill with cause `ability` or `counterstrike` counts as an `attack` kill does.
- A Wealth kill, a kill caused by `f3`, a kill of `f1`'s own asset, and a removal
  with no `CausedByFactionID` each give nothing.
- Progress 2 plus one kill at Force 3 completes: `GoalProgressed`, `GoalCompleted`,
  `XPAwarded`, with XP 2 (⌈3 / 2⌉).
- Two kills in one batch at progress 2 give `Delta: 2` and XP 2 (⌈4 / 2⌉).
- Commercial Expansion and Intelligence Coup: one row each, counting only their
  category against their own rating.
- Invincible Valor: a kill of `force-high` at Force 3 completes with XP 2 and no
  `GoalProgressed`; a kill of `force-def` gives nothing; a `MinRating` equal to the
  Force rating gives nothing.

**Damage (`G-006`), Decision 1**

- `AssetHPDelta{Delta: -3}` on a rival caused by `f1` gives `Delta: 3`.
- A Base hit written as the Attack action writes it (`BaseHPDelta{Delta: -4}` and
  `FactionHPDelta{Delta: -4}`, both caused by `f1`) gives `Delta: 4`, not 8.
- A positive delta, a zero delta, damage to `f1`'s own asset and damage caused by
  `f3` give nothing.
- Progress 6 plus 2 damage at ratings 3/3/2 completes with XP 2.

**Base planted (`G-005`), D-22**

- A `BaseAdded` caused by `f1` on a world where it has no Base completes with XP 1.
- The same with a rival asset on the world pays 2; with only a rival Base on the
  world pays 2.
- A `BaseAdded` on a world where `f1` already has a Base gives nothing.
- A `BaseAdded` caused by `f2` gives nothing.

**Influence (`G-011`)**

- `InfluenceDelta{Delta: 3}` caused by `f1` gives `Delta: 3`; a negative delta and
  one caused by `f2` give nothing.
- Progress 6 plus 2 at Wealth 2 (threshold 8) completes with XP 2.

**Action taken (`G-007`), D-16 and `B.013`**

- `ActionTaken{f1, "Buy Asset"}` gives `Delta: 1`.
- `ActionTaken{f1, "Attack"}` at progress 2 gives one `GoalProgressed{Delta: -2, Cause: "goal_reset"}`; at progress 0 it gives nothing.
- An Attack batch with no damage and no stealth loss (a clean miss) still resets.
- `ActionTaken{f2, "Buy Asset"}` gives nothing. An empty batch gives nothing.
- Progress 3 plus a peaceful action completes with XP 1.

**Government taken (`G-004`), D-8 and D-9**

- `WorldGovernmentChanged{FromFactionID: "f2", ToFactionID: "f1", Resisted: true}`
  with `f2` at ratings 3/4/4 completes with XP 2 (⌈11 / 6⌉).
- The same with `Resisted: false` pays 1. An independent world (`FromFactionID`
  empty) pays 1. A resisting ruler at 1/1/1 pays 1.
- A change to `f2` gives nothing.

**Faction destroyed (`G-008`), D-9 and D-15**

- `f1` targets `f2` (HP 5, ratings 3/4/4). `FactionHPDelta{f2, -5}` caused by `f1`
  completes with XP 5 (1 + ⌈11 / 3⌉).
- Two losses in one batch, -3 by `f3` then -2 by `f1`, complete.
- -3 by `f1` then -2 by `f3` gives nothing: the last blow is not `f1`'s.
- A loss that leaves `f2` at 1 HP gives nothing. The destruction of `f3`, which is
  not the target, gives nothing.
- No target set (`TargetFactionID` empty) gives nothing.

**Stealthed assets (`G-009`), D-10 and Decision 2**

- `f1` (Cunning 3) holds two stealthed assets on a world `f2` governs. A batch that
  stealths a third there (`AssetStealthApplied`) gives `Delta: 1` and completes
  with XP 2 in that batch.
- The count follows the state: an `AssetStealthCleared` in the batch gives
  `Delta: -1`; an asset moved off rival ground in the batch stops counting.
- Ground: the seven worlds of the old `TestInsideEnemyTerritory_CountsRivalGround`
  (governed by `f1`, governed by a rival, independent, a rival's independent
  homeworld, a seized rival homeworld, `f1`'s lost homeworld, in transit). Only the
  rival-governed world, the rival's independent homeworld and the lost homeworld
  count.
- An asset ID in `StealthedAtStart` never counts.
- `project` leaves the original state untouched: after `evaluate`, the live asset
  is still not stealthed.

**Order of output**

- A completing batch returns `GoalProgressed` first, then `GoalCompleted`, then
  `XPAwarded`, and `GoalCompleted.XPAwarded` equals `XPAwarded.Amount`.

##### Commit message

```
feat(E1): a goal engine that reads goal parts

- compile checks every event, filter, amount, measure and formula name in a
  goal's parts, and every combination of parts (D-18)
- seven event matchers, the stealthed_assets measure and four XP formulas; each
  formula rounds up (D-9)
- damage counts asset and Base hits only, because a Base hit also emits a
  FactionHPDelta (corrects D-20's mechanism)
- a measure counts on a projected copy of the state (D-10)
- ActiveGoal gains StealthedAtStart; WorldGovernmentChanged gains Resisted (D-8)
- not wired yet: the handlers still run the goals
```

### Phase 4 — Swap

**In plain words:** the engine starts to run goals from their data, and the ten Go
handlers are deleted. Then goal IDs take the `CORE-` prefix, and the rules doc and
the contributor guide say what the code now does. When this lands, no Go names a
goal.

#### Commit 6 — `feat(E1): goals run from data`

The behavior change of the effort: D-9, D-10, D-14, D-15, D-16 and D-20 take effect
here, and `B.001` and `B.013` close.

##### Task 1 — delete the handlers

```
git rm -r internal/faction/engine/goal/goals
```

##### Task 2 — `engine/goal/goal.go`

Full contents:
```go
package goal

import (
	"fmt"
	"log/slog"
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// GoalEngine advances faction goals from the parts each goal states in the
// rulebook. It holds no goal ID of its own.
type GoalEngine struct {
	goals map[string]*compiledGoal
	log   *slog.Logger
}

// New compiles every goal in the rulebook. A goal whose parts name something the
// engine cannot run is an error, not a goal that does nothing.
func New(rulebook *rulebook.Rulebook, log *slog.Logger) (*GoalEngine, error) {
	goalEngine := &GoalEngine{goals: make(map[string]*compiledGoal, len(rulebook.Goals)), log: log}
	for _, goalID := range slices.Sorted(maps.Keys(rulebook.Goals)) {
		compiled, err := compile(rulebook.Goals[goalID])
		if err != nil {
			return nil, fmt.Errorf("goal %q: %w", goalID, err)
		}
		goalEngine.goals[goalID] = compiled
	}
	return goalEngine, nil
}

// UpdateProgress reads a batch for the acting faction's active goal and returns
// the goal mutations it causes. It must be called before MutationEngine.Apply.
func (goalEngine *GoalEngine) UpdateProgress(
	actingFactionID string,
	mutations []domain.Mutation,
	factionState *state.FactionState,
	rulebook *rulebook.Rulebook,
	log *slog.Logger,
) []domain.Mutation {
	actingFaction, ok := factionState.Factions[actingFactionID]
	if !ok || actingFaction.ActiveGoal == nil {
		return nil
	}
	compiled, ok := goalEngine.goals[actingFaction.ActiveGoal.GoalID]
	if !ok {
		// deliberate: a held goal the rulebook does not know is skipped, not an error, because an older save can hold one (Decision 10).
		log.Warn("faction holds an unknown goal", "faction", actingFactionID, "goal", actingFaction.ActiveGoal.GoalID)
		return nil
	}
	return compiled.evaluate(evalContext{
		faction:      actingFaction,
		batch:        mutations,
		factionState: factionState,
		rulebook:     rulebook,
	})
}
```

##### Task 3 — `engine/core.go`

**Find:**
```go
	e.Goal = goal.New(log)
```
**Replace with:**
```go
	goalEngine, err := goal.New(rulebook, log)
	if err != nil {
		return nil, fmt.Errorf("engine: %w", err)
	}
	e.Goal = goalEngine
```

##### Task 4 — `engine/orchestrator.go`

The goal engine reads live state for a contested world (Decision 9), so the action
phase needs no world index. The movement phase already refuses a nil world engine
earlier in the turn.

**Find:**
```go
	var worldIndex *world.Index
	if e.World == nil {
		return ErrWorldEngineUnavailable
	}
	worldIndex = e.World.Index
	actionMutations = append(actionMutations, e.firePurchaseEffects(faction, actionMutations, factionState)...)
```
**Replace with:**
```go
	actionMutations = append(actionMutations, e.firePurchaseEffects(faction, actionMutations, factionState)...)
```
**Find:**
```go
	goalMutations := e.Goal.UpdateProgress(faction.ID, combined, factionState, e.Rulebook, worldIndex, phaseLog.With("engine", "goal"))
```
**Replace with:**
```go
	goalMutations := e.Goal.UpdateProgress(faction.ID, combined, factionState, e.Rulebook, phaseLog.With("engine", "goal"))
```

##### Task 5 — `engine/goal/goal_test.go`

Replace the file's tests. It keeps `shippedRulebook` from `compile_test.go`.

- `TestNew_CompilesTheShippedGoals`: `New` on the shipped rulebook returns no error
  and holds eleven goals.
- `TestNew_RefusesABadGoal`: a rulebook with one goal `X-1` that has no `xp` part
  fails with an error that contains `goal "X-1"` and `needs an xp part`.
- `TestUpdateProgress_NoActiveGoal`: returns nil.
- `TestUpdateProgress_UnknownFaction`: returns nil.
- `TestUpdateProgress_UnknownGoalIsSkipped`: a faction that holds `G-012` returns
  nil and does not panic (Decision 10).
- `TestUpdateProgress_RunsTheHeldGoal`: a faction that holds Military Conquest (by
  the ID the shipped rulebook gives that name) and a batch with one Force kill
  returns one `GoalProgressed`.

##### Task 6 — scenario and construction tests

- **`engine/core_test.go`** — `TestNewWithRulebook_RefusesBadGoalData`: a rulebook
  whose one goal names `xp = {formula: "double"}` makes `NewWithRulebook` return an
  error that contains `engine: goal` and `unknown formula`.
- **`engine/testharness/scenarios/goals_test.go`** (new):
  - `TestPeaceableKingdom_MissedAttackResets` (`B.013`): alpha holds Peaceable
    Kingdom at progress 2 and attacks beta with rolls that miss (attack die 1,
    defense die 10) and no counter damage. After the cycle alpha's progress is 0,
    and history holds a `goal_progressed` with cause `goal_reset`.
  - `TestPeaceableKingdom_PassDoesNotCount` (D-16): alpha holds it at progress 2
    and passes. Progress stays 2.
  - `TestPeaceableKingdom_PeacefulActionCounts`: alpha takes Repair Faction.
    Progress is 3.
  - `TestBloodTheEnemy_BaseHitCountsOnce` (Decision 1): alpha holds Blood the Enemy
    and attacks beta on beta's homeworld; `ConfirmRedirectToBaseFn` answers yes.
    Alpha's progress equals the damage rolled, once.
  - `TestMilitaryConquest_RoundsXPUp` (D-9): alpha at Force 3 and progress 2 kills
    a Force asset. Alpha gains 2 XP.
  - Each test sets `ActiveGoal.GoalID` through a helper
    `goalID(t, h, "Peaceable Kingdom")` that finds the ID by name in
    `h.Engine.Rulebook.Goals`.
- **`engine/testharness/scenarios/actions_test.go`** —
  `TestRunCycle_GoalCompleted_MilitaryConquest` must pass unchanged (progress 1,
  Force 1, one kill: XP ⌈2 / 2⌉ = 1).
- **`engine/testharness/scenarios/tags_test.go`** —
  `TestExchangeConsulate_BonusXPOnPeaceableKingdom` must pass unchanged: the
  purchase is a peaceful action, read from `ActionTaken` now.

If a scenario's `dicetest.Sequence` runs short or long, fix its values and log the
change in the Decision Record.

##### Task 7 — check the swap

`go build ./... && go vet ./... && go test ./...`. Then:

```
grep -rnE '"G-0[0-9]{2}"' internal/ cmd/ --include='*.go' | grep -v _test.go
```

No hit is allowed: no production Go names a goal ID (`B.001`).

##### Commit message

```
feat(E1): goals run from data

- goal.New compiles every goal in the rulebook and NewWithRulebook returns its
  error; the ten Go handlers and the Handler registry are deleted (B.001)
- any kill or damage the faction causes counts, whatever the cause (D-14)
- every XP formula rounds up (D-9)
- Peaceable Kingdom reads ActionTaken, so a missed attack resets it (B.013) and
  a pass neither counts nor resets (D-16)
- Inside Enemy Territory counts stealthed assets held on rival ground, not
  stealth events (D-10)
- a faction that holds a goal the rulebook does not know is skipped with a
  warning
```

#### Commit 7 — `refactor(E1): goal IDs carry the CORE prefix`

D-26 and `E.002` I-1. Data and tests only.

##### Task 1 — `rulebooks/_core/goals.toml`

Change the eleven `id` lines: `"G-001"` becomes `"CORE-G-001"`, and so on to
`"CORE-G-011"`. Nothing else in the file changes.

##### Task 2 — `rulebooks/swn/tags.toml`

Exchange Consulate's reference. `checkReferences` refuses the load if it is left
behind.

**Find:**
```toml
    filters = { owner = "own", scope = "faction", goal = "G-007" },
```
**Replace with:**
```toml
    filters = { owner = "own", scope = "faction", goal = "CORE-G-007" },
```

##### Task 3 — tests

In each file below, replace every `"G-0` with `"CORE-G-0`:

- `internal/faction/rulebook/rulebook_test.go`
- `internal/faction/rulebook/goals_test.go`
- `internal/faction/rulebook/effects_test.go`
- `internal/faction/engine/effect/effects/xp_roll_test.go`
- `internal/faction/engine/goal/goal_test.go`
- `internal/faction/engine/testharness/scenarios/actions_test.go`
- `internal/faction/engine/testharness/scenarios/tags_test.go`
- `internal/faction/engine/action/actions/simple_actions_test.go`
- `internal/faction/state/faction_state_test.go`

In `rulebook_test.go` the duplicate-ID row's wanted text follows:
`duplicate goal ID "CORE-G-001"`. Other literals such as `"G-404"` stay as they are;
they name a goal that does not exist.

Then `grep -rn '"G-0' internal/ rulebooks/` must return nothing.

##### Task 4 — re-seed sandbox-claude (not committed)

The campaign is gitignored. Robert's `campaigns/sandbox` is his to re-seed.

1. `go run ./cmd/gm-toolkit campaign add-rules --campaign sandbox-claude --rules rulebooks/swn --replace`
2. In `campaigns/sandbox-claude/state/state.seed.toml` and each
   `campaigns/sandbox-claude/state/scenarios/*.toml`, use Edit on each
   `active_goal` line:
   - a line with `goal_id = "G-004"` is deleted: the seizure it staged is gone
     (D-1). The faction starts with no goal.
   - any other line keeps `goal_id`, `target_faction_id` and `progress`, takes the
     prefix, and loses `process_phase`, `turns_remaining` and `target_world`. So
     `purchase.toml`'s Veiled Hand line becomes
     `active_goal = { goal_id = "CORE-G-009", target_faction_id = "", progress = 0 }`.
3. `./cmd/scripts/reset-campaign.sh sandbox-claude`
4. Run one cycle against `--campaign sandbox-claude` in tmux. Confirm the campaign
   loads, the action menu shows neither Seize Planet nor Change Homeworld, and the
   cycle completes. Then reset to the `purchase` scenario and confirm the detail
   view names Veiled Hand's goal as Inside Enemy Territory.

##### Commit message

```
refactor(E1): goal IDs carry the CORE prefix

- goal IDs are CORE-G-001 to CORE-G-011 in _core/goals.toml (E.002 I-1)
- Exchange Consulate's xp_roll names CORE-G-007, checked at load
- saves move by a sandbox re-seed, with no migration code (D-26)
```

#### Commit 8 — `docs(E1): goal rules readings and the goal recipe`

Overview Decision 5 for the rules doc. The contributor guide's goal recipe and its
skip rule describe code that Commits 2 and 6 deleted.

##### Task 1 — `docs/rules/swn-faction-mechanics.md`

Load `/gestalt:style-guide` first.

**Find:**
```markdown
| Military Conquest | Destroy Force assets of rivals equal to your Force rating | half assets destroyed |
| Commercial Expansion | Destroy Wealth assets of rivals equal to your Wealth rating | half assets destroyed |
| Intelligence Coup | Destroy Cunning assets of rivals equal to your Cunning rating | half assets destroyed |
| Planetary Seizure | Take control of a planet, becoming its government | half avg of ruling faction's F/C/W (min 1) |
| Expand Influence | Plant a Base of Influence on a new planet | 1, +1 if contested |
| Blood the Enemy | Inflict HP damage on enemy assets/bases equal to total F+C+W | 2 |
| Peaceable Kingdom | Don't take an Attack action for four turns | 1 |
| Destroy the Foe | Destroy a rival faction | 1 + avg of target faction's F/C/W |
```
**Replace with:**
```markdown
| Military Conquest | Destroy Force assets of rivals equal to your Force rating | half assets destroyed, rounded up |
| Commercial Expansion | Destroy Wealth assets of rivals equal to your Wealth rating | half assets destroyed, rounded up |
| Intelligence Coup | Destroy Cunning assets of rivals equal to your Cunning rating | half assets destroyed, rounded up |
| Planetary Seizure | Take control of a planet, becoming its government | half avg of the resisting ruler's F/C/W, rounded up (min 1); 1 if no faction resisted |
| Expand Influence | Plant a Base of Influence on a new planet | 1, +1 if contested |
| Blood the Enemy | Inflict HP damage on enemy assets/bases equal to total F+C+W | 2 |
| Peaceable Kingdom | Don't take an Attack action for four turns | 1 |
| Destroy the Foe | Destroy a rival faction | 1 + avg of target faction's F/C/W, rounded up |
```

**Find:**
```markdown
| Wealth of Worlds | Spend Coin equal to 4× Wealth rating on bribes (lost); Wealth must increase before selecting again | 2 |
```
**Replace with:**
```markdown
| Wealth of Worlds | Spend Coin equal to 4× Wealth rating on bribes (lost); Wealth must increase before selecting again | 2 |

### Goal Rulings

The book leaves these points open. The engine reads them as follows.

- **Rounding** — every difficulty that divides rounds up. Three kills pay 2 XP. A
  resisting ruler at Force 3, Cunning 4, Wealth 4 pays 2 for a Planetary Seizure and
  1 + 4 for Destroy the Foe.
- **What counts as destroying or damaging** — any kill or damage the faction
  causes: an attack, a counterattack, an ability or an asset's special effect.
- **Bases** — damage to any rival Base counts for Blood the Enemy, the homeworld
  Base included. It counts once, although the Base and the faction both lose the
  hit points.
- **Planetary Seizure** — the previous ruler resisted when it had an unstealthed
  asset on the world as the seizure began. Stealthed assets are not resistance. With
  no resisting ruler, and on a world with no faction ruler, the difficulty is 1.
- **Expand Influence** — the attempt is contested when a rival asset or Base is on
  the world.
- **Peaceable Kingdom** — a turn counts when the faction takes an action that is
  not Attack. An Attack resets the count even if every attack misses. A turn with
  no action neither counts nor resets.
- **Destroy the Foe** — the goal completes only when the faction itself lands the
  blow that takes the target to 0 hit points.
- **Inside Enemy Territory** — the goal counts the stealthed assets the faction
  holds at one time on worlds a rival governs, or on a rival's homeworld that has no
  faction government. An asset that loses stealth or leaves stops counting.
```

Effort 2 adds the rulings for D-13 and D-21, and Effort 3 those for the pick, the
start record of Inside Enemy Territory and the Wealth of Worlds gate.

##### Task 2 — `docs/contributing/engine.md`

**Find:**
```markdown
- **Data-only TOML silently skips.** A tag/asset/goal ID with no registered handler
  is an intentional no-op, not an error — so **add the rulebook entry first**, or
  your handler is keyed against an ID that never matches and never runs.
```
**Replace with:**
```markdown
- **Data-only TOML silently skips, except where the data states behavior.** An
  asset ID with no registered handler is an intentional no-op, not an error — so
  **add the rulebook entry first**, or your handler is keyed against an ID that
  never matches and never runs. An `effects` entry or a goal part is stricter: a
  name the engine cannot run fails engine construction.
```

**Find:**
```markdown
`full_cycle_test.go`, `hooks_test.go`, `tags_test.go`, `goal_lock_test.go`,
```
**Replace with:**
```markdown
`full_cycle_test.go`, `hooks_test.go`, `tags_test.go`, `goals_test.go`,
```

**Find:**
```markdown
- **Key order.** The entry's own keys first, then its parts: `attack`, `transport`,
  `ability`, `effects`. Inside an effect: `trigger`, `effect`, `filters`, `gates`,
```
**Replace with:**
```markdown
- **Key order.** The entry's own keys first, then its parts: `attack`, `transport`,
  `ability`, `effects`. A goal's parts: `start`, `progress`, `reset`, `threshold`,
  `complete`, `xp`. Inside an effect: `trigger`, `effect`, `filters`, `gates`,
```

Replace the whole recipe, from the heading `### Implement a goal handler` to the
end of its `**Test:**` paragraph, with:

```markdown
### Add a goal

> Shape & why: [goals](../architecture/engine/goals.md)

A goal is data. Add it with no Go when the existing vocabulary can state its rule.

1. **rulebook `goals.toml`** — one entry with `id`, `name`, `description` and its
   parts. A goal needs `xp`, and exactly one of `threshold` (with `progress`) or
   `complete`. `start` and `reset` are optional.
2. **`engine/goal/vocabulary.go`** — only when the rule needs a new word. An event
   is a row in `events` with the filters and amounts it accepts and a matcher in
   `events.go`; a count read from state is a row in `measures`; an XP rule is a
   formula in `xp.go` with its check in `checkXP`.
3. **`domain/goal.go` and `rulebook/goals.go`** — only for a new filter or key: the
   field, its name constant, its line in `Used`, and the record field.

> A matcher reads the state **before** the batch applies, where an
> about-to-be-removed asset still exists. A measure reads the state the batch
> leaves, from a copy.

**Test:** a row in `TestCompile_Refuses` for each new check, and an
`evaluate_test.go` case built from the shipped goal for each new word.
```

##### Commit message

```
docs(E1): goal rules readings and the goal recipe

- the rules doc states the goal readings this effort lands: rounding, any
  caused kill or damage, Base hits, resistance, contested, Peaceable Kingdom's
  turns, the faction's own blow, assets held (D-8 to D-10, D-14 to D-16, D-20,
  D-22)
- the contributor guide's goal recipe is "add a goal" in data; the skip rule
  and the TOML key order follow
- ticks Phase 4 and Effort 1 in the overview's Effort Summary
```

The last commit of the effort ticks the Phase 4 box above and the Effort 1 box in
the overview plan's Effort Summary, and changes that row's File cell to a link to
this plan.

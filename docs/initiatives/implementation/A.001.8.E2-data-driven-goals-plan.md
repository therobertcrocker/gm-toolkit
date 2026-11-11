# Data-Driven Goals (A.001.8) — Effort 2 Plan: Every Batch

## In Plain Words

- **What:** every faction's goal reads every batch of changes, on any faction's
  turn and in any phase, not only the acting faction's action. A goal that names a
  target faction fails, with a new `GoalFailed` record, when another faction
  destroys that target. The digest and the TUI feed credit each goal line to the
  faction that holds the goal.
- **Why:** today a kill made while defending, or anything in a movement or
  bookkeeping batch, never counts toward a goal, and the rules name no turn. And
  Destroy the Foe has no way to end when a third party lands the last blow.
- **How it fits:** Effort 2 of three in `A.001.8` (arc `A.001`), on branch
  `feature/data-driven-goals`. Effort 1 made goals data. Effort 3 adds the
  turn-start goal pick, which also writes the target that the lost-target rule
  reads.

## Context / Goal

- Overview: [`A.001.8-data-driven-goals-overview-plan.md`](./A.001.8-data-driven-goals-overview-plan.md),
  Effort 2. Its Decision 5 places the rules doc edit for D-13 and D-21 here.
- Discovery: [`A.001.8-data-driven-goals-discovery.md`](../discovery/A.001.8-data-driven-goals-discovery.md),
  D-13, D-15 and D-21, and its Built-in rules.
- Effort 1: [`A.001.8.E1-data-driven-goals-plan.md`](./A.001.8.E1-data-driven-goals-plan.md).
  Its Decision Record 15 left the measure's full state copy for this effort to
  judge (Decision 10).
- Architecture read: goals, orchestrator, effect-mutation (Key Decisions),
  narrative-digest.
- The discovery left no open question. Decisions 1 to 4 below are Robert's calls
  from this session; the rest settle what the code showed.
- Until Effort 3, nothing in play writes `ActiveGoal.TargetFactionID`, so the
  lost-target rule runs only in tests (overview, Gaps between efforts).

## Decisions Ratified in Planning

1. **Goals run inside `applyAndRecord`.** Robert's call. Every batch passes
   through it, so "every batch" (D-13) holds by construction and a batch added
   later cannot skip goals. It reads every faction's goal, runs the reactors over
   the goal output alone, applies, records, and returns the batch as applied; each
   of the eight call sites passes that batch on to its observer. The action phase
   loses its own goal call. Rejected: a helper called at each site (a new batch
   must remember it).

2. **The feed's goal-completed and goal-failed lines name the holder.** Robert's
   call. `goal completed: beta (+2 XP)`, `goal failed: beta (target_lost)`, by
   faction ID, as `captured by beta` does; `formatMutation` stays a pure function
   of the mutation. The `goal abandoned` line stays as it is: Abandon Goal is the
   holder's own action, under its own header. Rejected: a name only for a rival's
   goal (the formatter would need the acting faction); no name.

3. **`GoalFailed`'s headline is decided per cause; `target_lost` gets none.**
   Robert's call. A goal can fail for other causes later, and each weighs its own
   headline. `target_lost` follows a faction's fall, which is the bigger story. The
   digest's `GoalEvent` carries the cause, the wire line's templates are chosen by
   cause, and a fence-sign at the Goal Abandoned rule in `selectHeadline` names the
   choice. Rejected: one headline for every `GoalFailed`; none, with no hook for a
   later cause.

4. **One phase, three commits.** Robert's call. About 1,100 lines, under the phase
   budget. Rejected: two phases split at `GoalFailed` (two short sessions on one
   reviewable surface).

5. **The digest credits a goal completion and its XP to the holder's beat.**
   Planning call; the discovery and the overview missed it. `goal_completed` and
   `xp_awarded` were added to the beat of the record's acting faction. Under D-13 a
   rival's goal completes inside another faction's batch, so both cases now read
   the mutation's `FactionID`, as `tag_removed` does. `goal_failed` does the same.
   `goal_abandoned` keeps the acting beat; only the holder emits it.

6. **`UpdateProgress` reads every holder, in faction ID order, against the state
   before the batch.** Planning call. It drops `actingFactionID`. No goal reads
   another goal's output in the same batch. A held goal the rulebook does not know
   now warns once per batch, not once per action batch; accepted.

7. **The goal output gets a reactor pass in every batch.** Planning call, from the
   overview's findings. The turn-start, condition-lift, bookkeeping and stat-raise
   batches run no reactors over their own mutations, but Exchange Consulate's
   `xp_roll` reacts to a `GoalCompleted` wherever it lands. The two fence-signs on
   the old goal call move with it into `withGoalProgress`, unchanged.

8. **A batch that loses the target returns only `GoalFailed`.** Planning call. The
   goal ends there; progress in the same batch would end with it.

9. **The lost-target rule runs after completion, in every batch.** Planning call.
   `evaluate` splits into `advance` (reset, progress, completion) and the rule, so
   a reset batch cannot hide a lost target. An abandoning batch still returns
   nothing (Effort 1 Decision Record 11).

10. **The stealth measure keeps its full state copy.** Planning call; closes Effort
    1 Decision Record 15. Now it runs once per Inside Enemy Territory holder per
    batch. States hold a handful of factions and tens of assets, so the copy is
    cheap. Rejected: a whitelist of mutations that can move the count (it goes
    stale when a mutation is added).

11. **The `GoalFailed` cause is a `domain` constant.** Planning call. The goal
    engine writes it and the wire renderer keys its templates on it.

12. **The rules doc gets D-13 and D-21 only.** Planning call (overview Decision 5).
    "The faction picks its next goal at the start of its next turn" waits for
    Effort 3's pick, so the doc never runs ahead of the code.

## Decision Record — Execution

1. **No anchor drift in Commits 1 and 2.** Every Find matched the source. The
   `goalFailedTemplates` map needed no new import: `wire_templates.go` already
   imports `domain`.

2. **The third-party scenario places alpha on Krylos and picks beta's asset by
   owner (Commit 2).** The plan named no world for alpha, and
   `scriptAttackOnFirstDefender` gave gamma alpha's asset first, so the attack
   missed beta. The test keeps that helper for the attackers and sets its own
   defender script, which picks the asset beta owns, as Task 8 states. The dice
   `[10, 1, 2]` held: the Base takes 3 and beta goes from 1 HP to -2.

3. **`TestWireRendererGoalFailed` renders a whole digest with a quiet headline
   (Commit 2).** The goal line is the only content, so the check that the output
   holds the faction and the goal name reads that line alone.

---

## Shared Context

### Branch, cadence and checks

- Branch `feature/data-driven-goals`. One commit per plan commit, scope `E2`. At the
  effort boundary: `/gestalt:effort-close A.001.8.E2`.
- Every commit builds and passes `go build ./... && go vet ./... && go test ./...`
  from the repo root. Run `gofmt -w` on each touched Go file.
- Edit files only with Edit and Write.
- Test tasks are written in plain words. A test edit given as "replace every" is an
  Edit with `replace_all: true`. A dice sequence in a scenario test is the
  intended roll order; fix it against the harness if a roll runs short or long, and
  log the change.
- A task that adds a cross-package name ends with "add the import the compiler
  asks for".
- Paths in task titles are relative to `internal/faction/` unless they start with
  `docs/`.

### Pre-merge additions (beyond the overview's)

- **Graduate to Key Decisions (orchestrator):** goals run inside `applyAndRecord`,
  which now reads goals before it applies (Decision 1). This changes "the single
  choke point for state change" on that page.
- **Graduate to Key Decisions (narrative-digest):** goal events go to the holder's
  beat (Decision 5); a `GoalFailed` headline is weighed per cause (Decision 3).
- **Graduate to Key Decisions (goals):** a lost target fails the goal after
  completion is checked (Decisions 8 and 9).

## Out of Scope

- The turn-start pick, the writer of `TargetFactionID`, both gates and the wizard:
  Effort 3.
- A `GoalFailed` headline for `target_lost` (Decision 3); a feed name on the
  abandon line (Decision 2).
- A scenario test of Exchange Consulate's roll on a rival's turn. Its only goal is
  Peaceable Kingdom, which completes only on its holder's own action.
- The architecture pages: the initiative close rewrites them (overview Decision 6).

---

## Work Breakdown

**Overview**

- [x] **Phase 1 — Every batch**
  - Commit 1 — `feat(E2): goals read every batch for every faction` (5 tasks)
  - Commit 2 — `feat(E2): a lost target fails the goal` (8 tasks)
  - Commit 3 — `docs(E2): goal rules for every batch and a lost target` (2 tasks)

### Phase 1 — Every batch

**In plain words:** goals stop reading only the acting faction's action and read
every batch for every faction, with goal lines credited to their holder; then a
goal whose target falls to another faction fails, and the rules doc says both.

#### Commit 1 — `feat(E2): goals read every batch for every faction`

##### Task 1 — `engine/goal/goal.go`

`UpdateProgress` reads every faction that holds a goal (Decision 6).

**Find:**
```go
// UpdateProgress reads a batch for the acting faction's active goal and returns
// the goal mutations it causes. It must be called before MutationEngine.Apply.
func (goalEngine *GoalEngine) UpdateProgress(
	actingFactionID string,
	mutations []domain.Mutation,
	factionState *state.FactionState,
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
		rulebook:     goalEngine.rulebook,
	})
}
```
**Replace with:**
```go
// UpdateProgress reads one batch for every faction's active goal, in faction ID
// order, and returns the goal mutations it causes. Every goal reads the state
// before the batch, so it must be called before MutationEngine.Apply.
func (goalEngine *GoalEngine) UpdateProgress(
	mutations []domain.Mutation,
	factionState *state.FactionState,
	log *slog.Logger,
) []domain.Mutation {
	var goalMutations []domain.Mutation
	for _, factionID := range slices.Sorted(maps.Keys(factionState.Factions)) {
		faction := factionState.Factions[factionID]
		if faction.ActiveGoal == nil {
			continue
		}
		compiled, ok := goalEngine.goals[faction.ActiveGoal.GoalID]
		if !ok {
			// deliberate: a held goal the rulebook does not know is skipped, not an error, because an older save can hold one (Decision 10).
			log.Warn("faction holds an unknown goal", "faction", factionID, "goal", faction.ActiveGoal.GoalID)
			continue
		}
		goalMutations = append(goalMutations, compiled.evaluate(evalContext{
			faction:      faction,
			batch:        mutations,
			factionState: factionState,
			rulebook:     goalEngine.rulebook,
		})...)
	}
	return goalMutations
}
```

##### Task 2 — `engine/orchestrator.go`

`applyAndRecord` runs the goals (Decision 1); the action phase loses its own call;
every call site takes the batch back.

**2a.** The funnel. **Find:**
```go
// applyAndRecord applies mutations to in-memory state and appends a single
// EventRecord to the history file. It does not persist state to disk — callers
// are responsible for calling state.Save at phase-gate checkpoints.
func (e *Engine) applyAndRecord(
	factionState *state.FactionState,
	faction *domain.Faction,
	mutations []domain.Mutation,
	paths *campaigns.Paths,
	phaseLog *slog.Logger,
) error {
	if len(mutations) == 0 {
		return nil
	}
	if err := e.Mutation.Apply(factionState, mutations, phaseLog.With("engine", "mutation")); err != nil {
		for _, miss := range err.Misses {
			phaseLog.Error("mutation miss",
				"mutation_type", miss.MutationType,
				"faction_id", miss.FactionID,
				"entity_id", miss.EntityID,
			)
		}
		return fmt.Errorf("applying mutations: %w", err)
	}
	if err := turn.RecordHistory(paths.HistoryPath, factionState, faction, mutations, phaseLog.With("engine", "turn")); err != nil {
		return fmt.Errorf("recording history: %w", err)
	}
	return nil
}
```
**Replace with:**
```go
// applyAndRecord reads the batch for every faction's goal, applies the batch and
// its goal mutations to in-memory state, and appends a single EventRecord to the
// history file. It returns the batch as applied, goal mutations included. It
// does not persist state to disk — callers are responsible for calling
// state.Save at phase-gate checkpoints.
func (e *Engine) applyAndRecord(
	factionState *state.FactionState,
	faction *domain.Faction,
	mutations []domain.Mutation,
	paths *campaigns.Paths,
	phaseLog *slog.Logger,
) ([]domain.Mutation, error) {
	if len(mutations) == 0 {
		return nil, nil
	}
	mutations, err := e.withGoalProgress(faction, mutations, factionState, phaseLog)
	if err != nil {
		return nil, err
	}
	if err := e.Mutation.Apply(factionState, mutations, phaseLog.With("engine", "mutation")); err != nil {
		for _, miss := range err.Misses {
			phaseLog.Error("mutation miss",
				"mutation_type", miss.MutationType,
				"faction_id", miss.FactionID,
				"entity_id", miss.EntityID,
			)
		}
		return nil, fmt.Errorf("applying mutations: %w", err)
	}
	if err := turn.RecordHistory(paths.HistoryPath, factionState, faction, mutations, phaseLog.With("engine", "turn")); err != nil {
		return nil, fmt.Errorf("recording history: %w", err)
	}
	return mutations, nil
}

// withGoalProgress appends every faction's goal mutations for the batch, after a
// reactor pass over them. Where the phase has reactors, they have already run
// over the batch.
func (e *Engine) withGoalProgress(
	faction *domain.Faction,
	mutations []domain.Mutation,
	factionState *state.FactionState,
	phaseLog *slog.Logger,
) ([]domain.Mutation, error) {
	// deliberate: goals read the batch after the reactors, not before, because a reactor's capture or removal is progress too (Effort 1 Decision Record 12).
	goalMutations := e.Goal.UpdateProgress(mutations, factionState, phaseLog.With("engine", "goal"))
	if len(goalMutations) == 0 {
		return mutations, nil
	}
	// deliberate: a second pass over the goal mutations alone, not the whole batch again, because the first pass already reacted to the rest (Effort 3 Decision 6).
	goalMutations, err := dispatch.MutationReactors(e.Hooks, faction, goalMutations, factionState, e.Rulebook)
	if err != nil {
		return nil, err
	}
	return append(mutations, goalMutations...), nil
}
```

**2b.** The action phase. **Find:**
```go
	// deliberate: goals read the batch after the reactors, not before, because a reactor's capture or removal is progress too (Effort 1 Decision Record 12).
	goalMutations := e.Goal.UpdateProgress(faction.ID, combined, factionState, phaseLog.With("engine", "goal"))
	if len(goalMutations) > 0 {
		// deliberate: a second pass over the goal mutations alone, not the whole batch again, because the first pass already reacted to the rest (Effort 3 Decision 6).
		goalMutations, err = dispatch.MutationReactors(e.Hooks, faction, goalMutations, factionState, e.Rulebook)
		if err != nil {
			if factionErrors.IsRecoverable(err) {
				phaseLog.Warn("recoverable failure", "err", err)
				observer.OnError(faction, err)
				return nil // continue past this sub-engine call
			}
			phaseLog.Error("failure", "err", err)
			observer.OnError(faction, err)
			return err
		}
	}
	combined = append(combined, goalMutations...)

	if err := e.applyAndRecord(factionState, faction, combined, paths, phaseLog); err != nil {
```
**Replace with:**
```go
	if combined, err = e.applyAndRecord(factionState, faction, combined, paths, phaseLog); err != nil {
```

**2c.** Turn-start effects. **Find:**
```go
		if err := e.applyAndRecord(factionState, faction, startMutations, paths, phaseLog); err != nil {
```
**Replace with:**
```go
		if startMutations, err = e.applyAndRecord(factionState, faction, startMutations, paths, phaseLog); err != nil {
```

**2d.** Bookkeeping. **Find:**
```go
		if err := e.applyAndRecord(factionState, faction, bookMutations, paths, phaseLog); err != nil {
```
**Replace with:**
```go
		if bookMutations, err = e.applyAndRecord(factionState, faction, bookMutations, paths, phaseLog); err != nil {
```

**2e.** Condition lifts, in `runConditionLifts`. **Find:**
```go
		err = e.applyAndRecord(factionState, faction, mutations, paths, phaseLog)
```
**Replace with:**
```go
		mutations, err = e.applyAndRecord(factionState, faction, mutations, paths, phaseLog)
```

**2f.** Stat raise. **Find:**
```go
		if err := e.applyAndRecord(factionState, faction, raiseMutations, paths, phaseLog); err != nil {
```
**Replace with:**
```go
		if raiseMutations, err = e.applyAndRecord(factionState, faction, raiseMutations, paths, phaseLog); err != nil {
```

**2g.** Movement tick. **Find:**
```go
		if err := e.applyAndRecord(factionState, faction, tickMutations, paths, phaseLog); err != nil {
```
**Replace with:**
```go
		if tickMutations, err = e.applyAndRecord(factionState, faction, tickMutations, paths, phaseLog); err != nil {
```

**2h.** Movement decisions. **Find:**
```go
		if err := e.applyAndRecord(factionState, faction, decisionMutations, paths, phaseLog); err != nil {
```
**Replace with:**
```go
		if decisionMutations, err = e.applyAndRecord(factionState, faction, decisionMutations, paths, phaseLog); err != nil {
```

**2i.** Granted movement, in `runGrantedMovement`. **Find:**
```go
		if err := e.applyAndRecord(factionState, faction, mutations, paths, phaseLog); err != nil {
			return resolved, err
		}
```
**Replace with:**
```go
		if mutations, err = e.applyAndRecord(factionState, faction, mutations, paths, phaseLog); err != nil {
			return resolved, err
		}
```

In every site above, `err` is already declared in the enclosing scope (the
sub-engine call before it). If the compiler reports one that is not, declare it
with `var err error` just above that call, not with `:=`, which would shadow the
batch.

##### Task 3 — `narrative/digest/build.go`

A goal completion and its XP go to the holder's beat (Decision 5).

**3a. Find:**
```go
			case "goal_completed":
				var m domain.GoalCompleted
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				beat.GoalEvents = append(beat.GoalEvents, GoalEvent{
```
**Replace with:**
```go
			case "goal_completed":
				var m domain.GoalCompleted
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				// deliberate: the holder's beat, not the acting faction's, because a goal can complete on a rival's turn (D-13).
				holder := getOrCreateBeat(beats, m.FactionID, factionState)
				seenFactionIDs[m.FactionID] = struct{}{}
				holder.GoalEvents = append(holder.GoalEvents, GoalEvent{
```

**3b. Find:**
```go
				beat.XPGained += m.Amount
				// A goal's xp_roll bonus follows its completion in the batch, so it
				// adds to the latest GoalCompleted.
				if m.Cause == string(domain.SpecialXPRoll) {
					for i := len(beat.GoalEvents) - 1; i >= 0; i-- {
						if beat.GoalEvents[i].Kind == GoalCompleted {
							beat.GoalEvents[i].XPAwarded += m.Amount
							break
						}
					}
				} else {
					// Fold into matching GoalCompleted if XPAwarded wasn't set by the mutation.
					for i := range beat.GoalEvents {
						if beat.GoalEvents[i].Kind == GoalCompleted && beat.GoalEvents[i].XPAwarded == 0 {
							beat.GoalEvents[i].XPAwarded = m.Amount
							break
						}
					}
				}
```
**Replace with:**
```go
				holder := getOrCreateBeat(beats, m.FactionID, factionState)
				seenFactionIDs[m.FactionID] = struct{}{}
				holder.XPGained += m.Amount
				// A goal's xp_roll bonus follows its completion in the batch, so it
				// adds to the latest GoalCompleted.
				if m.Cause == string(domain.SpecialXPRoll) {
					for i := len(holder.GoalEvents) - 1; i >= 0; i-- {
						if holder.GoalEvents[i].Kind == GoalCompleted {
							holder.GoalEvents[i].XPAwarded += m.Amount
							break
						}
					}
				} else {
					// Fold into matching GoalCompleted if XPAwarded wasn't set by the mutation.
					for i := range holder.GoalEvents {
						if holder.GoalEvents[i].Kind == GoalCompleted && holder.GoalEvents[i].XPAwarded == 0 {
							holder.GoalEvents[i].XPAwarded = m.Amount
							break
						}
					}
				}
```

##### Task 4 — `tui/views/turn/execution/format.go`

The goal-completed line names the holder (Decision 2).

**Find:**
```go
	case domain.GoalCompleted:
		return fmt.Sprintf("goal completed (+%d XP)", mut.XPAwarded), true
```
**Replace with:**
```go
	case domain.GoalCompleted:
		return fmt.Sprintf("goal completed: %s (+%d XP)", mut.FactionID, mut.XPAwarded), true
```

##### Task 5 — tests

- **`engine/goal/goal_test.go`:**
  - Replace every `UpdateProgress("f1", ` with `UpdateProgress(`.
  - Delete `TestUpdateProgress_UnknownFaction`; with no acting faction it tests
    nothing.
  - Add, after `TestUpdateProgress_RunsTheHeldGoal`:

    ```go
    func TestUpdateProgress_ReadsEveryHolder(t *testing.T) {
    	fixture := newGoalFixture(t, "Military Conquest", 0)
    	bloodTheEnemy := shippedGoal(t, fixture.rulebook, "Blood the Enemy").goal.ID
    	fixture.rival.ActiveGoal = &domain.ActiveGoal{GoalID: bloodTheEnemy}
    	fixture.give(fixture.rival, "a1", "force-def", "Krylos")
    	fixture.give(fixture.holder, "b1", "force-def", "Krylos")
    	goalEngine := newEngine(t, fixture.rulebook)
    	batch := []domain.Mutation{
    		domain.AssetHPDelta{FactionID: "f1", AssetID: "b1", Delta: -2, Cause: "counterstrike", CausedByFactionID: "f2"},
    		kill("f2", "a1", "attack", "f1"),
    	}

    	got := goalEngine.UpdateProgress(batch, fixture.factionState, discardLog())

    	fixture.want(got,
    		fixture.progressed(1, causeProgress),
    		domain.GoalProgressed{FactionID: "f2", GoalID: bloodTheEnemy, Delta: 2, Cause: causeProgress},
    	)
    }
    ```

- **`engine/orchestrator_apply_test.go`:** replace every
  `err := eng.applyAndRecord(` with `_, err := eng.applyAndRecord(`.

- **`engine/testharness/scenarios/goals_test.go`:** add, after
  `TestMilitaryConquest_RoundsXPUp`, a test named
  `TestMilitaryConquest_CounterstrikeOnARivalsTurnCompletes`:
  - alpha (Force 1, Cunning 3, Wealth 2) and beta (Force 2, Cunning 2, Wealth 2),
    both on Tartarus. alpha's Security Personnel has 1 HP. beta holds Military
    Conquest at progress 1.
  - `scriptAction(t, h, "Attack")` and `scriptAttackOnFirstDefender(h)`; dice
    `[1, 10, 4]`: alpha's attack 1+1 = 2 against beta's defense 10+2 = 12 misses,
    and beta's Security Personnel counters for 4, which destroys alpha's asset.
  - After one cycle: beta's goal is cleared, beta gained 1 XP (2 kills / 2), and
    history holds a `goal_completed` whose faction is beta.

- **`narrative/digest/digest_test.go`:** add, after
  `TestGoalCompleted_XPRollBonusJoinsTheGoalLine`, a test named
  `TestGoalCompleted_OnARivalsTurnGoesToTheHolder`: one record for `faction-a`
  holding a `goal_completed` and an `xp_awarded` (2 XP, cause `goal_completed`)
  for `faction-b`. With both factions in the state, `faction-b`'s beat has the
  one `GoalCompleted` event and `XPGained` 2, and `faction-a`'s beat has no goal
  event and `XPGained` 0.

- **`tui/views/turn/execution/format_test.go`:** the `GoalCompleted` row becomes
  `domain.GoalCompleted{FactionID: "beta", XPAwarded: 5}`, want
  `"goal completed: beta (+5 XP)"`.

##### Commit message

```
feat(E2): goals read every batch for every faction

- applyAndRecord reads every faction's goal on every batch, runs the reactors
  over the goal output, and returns the batch as applied (D-13)
- the action phase loses its own goal call; all eight call sites take the
  batch back for their observer
- UpdateProgress reads every holder in faction ID order
- the digest credits a goal completion and its XP to the holder's beat
- the feed's goal-completed line names the holder
```

#### Commit 2 — `feat(E2): a lost target fails the goal`

##### Task 1 — `domain/mutation.go`

`GoalFailed` and its one cause (D-21, Decision 11).

**Find:**
```go
func (m GoalAbandoned) Type() string { return "goal_abandoned" }
```
**Replace with:**
```go
func (m GoalAbandoned) Type() string { return "goal_abandoned" }

// GoalFailed records that a faction's active goal ended without completing. It
// pays no XP and costs no income. Cause says why.
type GoalFailed struct {
	FactionID string `json:"faction_id"`
	GoalID    string `json:"goal_id"`
	Cause     string `json:"cause"`
}

func (m GoalFailed) Type() string { return "goal_failed" }

// GoalFailedTargetLost is the GoalFailed cause when another faction destroys the
// goal's target faction (D-21).
const GoalFailedTargetLost = "target_lost"
```

##### Task 2 — `engine/mutation/mutation.go`

The apply arm clears the goal, as `GoalAbandoned`'s does.

**Find:**
```go
		case domain.GoalCompleted:
```
**Replace with:**
```go
		case domain.GoalFailed:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.ActiveGoal = nil
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.GoalCompleted:
```

##### Task 3 — `engine/goal/evaluate.go`

`evaluate` splits into `advance` and the lost-target rule (Decisions 8 and 9).

**Find:**
```go
func (compiled *compiledGoal) evaluate(evaluation evalContext) []domain.Mutation {
	// deliberate: a batch that abandons the held goal moves it no further, because a completion there would pay for a goal given up.
	if abandonsHeldGoal(evaluation) {
		return nil
	}
	goal := compiled.goal
	progress := evaluation.faction.ActiveGoal.Progress
	if compiled.reset != nil && len(compiled.reset.match(evaluation, goal.Reset.Filters)) > 0 {
		// deliberate: a reset ends the batch for this goal, not only zeroes it, because progress counted in the same batch would outlive the reset (Decision 8).
		if progress == 0 {
			return nil
		}
		return []domain.Mutation{compiled.progressed(evaluation, -progress, causeReset)}
	}
	var mutations []domain.Mutation
	delta := compiled.progressDelta(evaluation, progress)
	if delta != 0 {
		mutations = append(mutations, compiled.progressed(evaluation, delta, causeProgress))
		progress += delta
	}
	completing, completed := compiled.completion(evaluation, progress, delta != 0)
	if !completed {
		return mutations
	}
	xp := compiled.xp(evaluation, progress, completing)
	return append(mutations,
		domain.GoalCompleted{FactionID: evaluation.faction.ID, GoalID: goal.ID, XPAwarded: xp, Cause: causeCompleted},
		domain.XPAwarded{FactionID: evaluation.faction.ID, Amount: xp, Cause: causeCompleted},
	)
}
```
**Replace with:**
```go
func (compiled *compiledGoal) evaluate(evaluation evalContext) []domain.Mutation {
	// deliberate: a batch that abandons the held goal moves it no further, because a completion there would pay for a goal given up.
	if abandonsHeldGoal(evaluation) {
		return nil
	}
	mutations, completed := compiled.advance(evaluation)
	if !completed && compiled.targetLost(evaluation) {
		// deliberate: the failure replaces the batch's progress, not joins it, because the goal ends here and its progress ends with it (E2 Decision 8).
		return []domain.Mutation{domain.GoalFailed{FactionID: evaluation.faction.ID, GoalID: compiled.goal.ID, Cause: domain.GoalFailedTargetLost}}
	}
	return mutations
}

// advance moves the goal's progress for one batch and completes the goal when
// it is done. completed reports a completion.
func (compiled *compiledGoal) advance(evaluation evalContext) ([]domain.Mutation, bool) {
	goal := compiled.goal
	progress := evaluation.faction.ActiveGoal.Progress
	if compiled.reset != nil && len(compiled.reset.match(evaluation, goal.Reset.Filters)) > 0 {
		// deliberate: a reset ends the batch for this goal, not only zeroes it, because progress counted in the same batch would outlive the reset (Decision 8).
		if progress == 0 {
			return nil, false
		}
		return []domain.Mutation{compiled.progressed(evaluation, -progress, causeReset)}, false
	}
	var mutations []domain.Mutation
	delta := compiled.progressDelta(evaluation, progress)
	if delta != 0 {
		mutations = append(mutations, compiled.progressed(evaluation, delta, causeProgress))
		progress += delta
	}
	completing, completed := compiled.completion(evaluation, progress, delta != 0)
	if !completed {
		return mutations, false
	}
	xp := compiled.xp(evaluation, progress, completing)
	return append(mutations,
		domain.GoalCompleted{FactionID: evaluation.faction.ID, GoalID: goal.ID, XPAwarded: xp, Cause: causeCompleted},
		domain.XPAwarded{FactionID: evaluation.faction.ID, Amount: xp, Cause: causeCompleted},
	), true
}

// targetLost reports whether the batch destroys the goal's target faction, by
// any hand: a built-in rule, not a part (D-21). Only a goal whose start names a
// target_faction input has a target.
func (compiled *compiledGoal) targetLost(evaluation evalContext) bool {
	if !hasInput(compiled.goal, inputTargetFaction) {
		return false
	}
	return len(matchFactionDestroyed(evaluation, domain.GoalFilters{Target: targetStartInput})) > 0
}
```

##### Task 4 — `narrative/digest/types.go` and `build.go`

A failed goal is a goal event with its cause (Decision 3).

**4a.** `types.go`. **Find:**
```go
	GoalTagLost
	GoalTechLevelRaised
)

type GoalEvent struct {
	Kind      GoalEventKind
	GoalID    string
	GoalName  string
	XPAwarded int
	FromWorld string
	ToWorld   string
	TagName   string
	TechLevel int
}
```
**Replace with:**
```go
	GoalTagLost
	GoalTechLevelRaised
	GoalFailed
)

type GoalEvent struct {
	Kind      GoalEventKind
	GoalID    string
	GoalName  string
	XPAwarded int
	FromWorld string
	ToWorld   string
	TagName   string
	TechLevel int
	Cause     string
}
```

**4b.** `build.go`, after the `goal_abandoned` case. **Find:**
```go
					Kind:     GoalAbandoned,
					GoalID:   m.GoalID,
					GoalName: resolveGoalName(m.GoalID, rulebook),
				})
```
**Replace with:**
```go
					Kind:     GoalAbandoned,
					GoalID:   m.GoalID,
					GoalName: resolveGoalName(m.GoalID, rulebook),
				})

			case "goal_failed":
				var m domain.GoalFailed
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				holder := getOrCreateBeat(beats, m.FactionID, factionState)
				seenFactionIDs[m.FactionID] = struct{}{}
				holder.GoalEvents = append(holder.GoalEvents, GoalEvent{
					Kind:     GoalFailed,
					GoalID:   m.GoalID,
					GoalName: resolveGoalName(m.GoalID, rulebook),
					Cause:    m.Cause,
				})
```

##### Task 5 — `narrative/wire_templates.go` and `wire_renderer.go`

The wire line's text is chosen by cause (Decision 3).

**5a.** `wire_templates.go`. **Find:**
```go
var goalHomeworldTemplates = []string{
```
**Replace with:**
```go
// goalFailedTemplates holds one template set per GoalFailed cause, so a later
// cause never inherits wording about a target.
var goalFailedTemplates = map[string][]string{
	domain.GoalFailedTargetLost: {
		"{faction}'s {goal} ends unfinished — another hand struck the target down first.",
		"{faction} lost the chance at {goal}: a rival finished the foe before they could.",
	},
}

var goalHomeworldTemplates = []string{
```
Add the import the compiler asks for.

**5b.** `wire_renderer.go`, in `renderGoalEvent`. **Find:**
```go
	case digest.GoalAbandoned:
		tmpl := pick(rng, goalAbandonedTemplates)
		result := strings.ReplaceAll(tmpl, "{faction}", factionName)
		return strings.ReplaceAll(result, "{goal}", ge.GoalName)
```
**Replace with:**
```go
	case digest.GoalAbandoned:
		tmpl := pick(rng, goalAbandonedTemplates)
		result := strings.ReplaceAll(tmpl, "{faction}", factionName)
		return strings.ReplaceAll(result, "{goal}", ge.GoalName)
	case digest.GoalFailed:
		templates, ok := goalFailedTemplates[ge.Cause]
		if !ok {
			return fmt.Sprintf("%s's %s failed.", factionName, ge.GoalName)
		}
		result := strings.ReplaceAll(pick(rng, templates), "{faction}", factionName)
		return strings.ReplaceAll(result, "{goal}", ge.GoalName)
```

##### Task 6 — `narrative/digest/headline.go`

The fence-sign that names Decision 3 where a later cause would add its rule.

**Find:**
```go
	// Priority 5: Goal Abandoned
```
**Replace with:**
```go
	// deliberate: a failed goal gets no headline here, not one beside Goal Abandoned, because its one cause, target_lost, follows a faction's fall, the bigger story; each new cause weighs its own (A.001.8 E2 Decision 3).
	// Priority 5: Goal Abandoned
```

##### Task 7 — `tui/views/turn/execution/format.go`

The goal-failed line names the holder and the cause (Decision 2).

**Find:**
```go
	case domain.GoalCompleted:
		return fmt.Sprintf("goal completed: %s (+%d XP)", mut.FactionID, mut.XPAwarded), true
```
**Replace with:**
```go
	case domain.GoalCompleted:
		return fmt.Sprintf("goal completed: %s (+%d XP)", mut.FactionID, mut.XPAwarded), true
	case domain.GoalFailed:
		return fmt.Sprintf("goal failed: %s (%s)", mut.FactionID, mut.Cause), true
```

##### Task 8 — tests

- **`engine/goal/evaluate_test.go`:** replace `TestEvaluate_DestroyTheFoe` with:

  ```go
  func TestEvaluate_DestroyTheFoe(t *testing.T) {
  	cases := []struct {
  		name       string
  		targetID   string
  		batch      []domain.Mutation
  		wantXP     int
  		wantFailed bool
  	}{
  		{"the holder's blow destroys the target", "f2", []domain.Mutation{hpLoss("f2", -5, "f1")}, 5, false},
  		{"the holder's last blow after a third party's", "f2", []domain.Mutation{hpLoss("f2", -3, "f3"), hpLoss("f2", -2, "f1")}, 5, false},
  		{"a third party's last blow fails the goal", "f2", []domain.Mutation{hpLoss("f2", -3, "f1"), hpLoss("f2", -2, "f3")}, 0, true},
  		{"a third party's only blow fails the goal", "f2", []domain.Mutation{hpLoss("f2", -5, "f3")}, 0, true},
  		{"a third party's loss that leaves the target at 1 HP", "f2", []domain.Mutation{hpLoss("f2", -4, "f3")}, 0, false},
  		{"the destruction of a faction not targeted", "f2", []domain.Mutation{hpLoss("f3", -2, "f1")}, 0, false},
  		{"no target set", "", []domain.Mutation{hpLoss("f2", -5, "f1")}, 0, false},
  	}
  	for _, testCase := range cases {
  		t.Run(testCase.name, func(t *testing.T) {
  			fixture := newGoalFixture(t, "Destroy the Foe", 0)
  			fixture.holder.ActiveGoal.TargetFactionID = testCase.targetID

  			got := fixture.evaluate(testCase.batch...)

  			switch {
  			case testCase.wantFailed:
  				fixture.want(got, domain.GoalFailed{FactionID: "f1", GoalID: fixture.compiled.goal.ID, Cause: domain.GoalFailedTargetLost})
  			case testCase.wantXP > 0:
  				fixture.want(got, fixture.completed(testCase.wantXP)...)
  			default:
  				fixture.want(got)
  			}
  		})
  	}
  }
  ```

  Add `TestEvaluate_LostTargetNeedsATargetInput`: Military Conquest held by `f1`,
  with `TargetFactionID` set to `f2` by hand, and a batch where `f3` takes `f2`
  from 5 HP to 0. The result is empty: a goal with no `target_faction` input has
  no target to lose.

- **`engine/mutation/mutation_test.go`:** add `TestApply_GoalFailedClearsTheGoal`.
  `f1` holds a goal with a target; applying `GoalFailed` for `f1` clears
  `ActiveGoal` and leaves XP and Coin as they were. A second case: a `GoalFailed`
  for an unknown faction is a miss.

- **`narrative/digest/digest_test.go`:** add `TestGoalFailed`, after
  `TestGoalAbandoned`: one record for `faction-a` holding a `goal_failed` for
  `faction-b` (cause `target_lost`). `faction-b`'s beat has one `GoalFailed`
  event with the goal's name and cause `target_lost`; `faction-a`'s beat has
  none. The digest's headline is not `HeadlineGoalAbandoned`.

- **`narrative/wire_renderer_test.go`:** add `TestWireRendererGoalFailed`: a
  digest with one `GoalFailed` event (cause `target_lost`, goal name "Destroy the
  Foe") renders a line that holds the faction name and "Destroy the Foe" for
  seeds 1 to 10, and an event with an unknown cause renders
  "<faction>'s Destroy the Foe failed.".

- **`tui/views/turn/execution/format_test.go`:** add the row
  `{"GoalFailed", domain.GoalFailed{FactionID: "beta", Cause: domain.GoalFailedTargetLost}, "goal failed: beta (target_lost)"}`.

- **`engine/testharness/scenarios/goals_test.go`:**
  - Split `scriptAction` so a named faction can act:

    **Find:**
    ```go
    // scriptAction makes alpha take the named action and every other faction pass.
    // An empty name makes alpha pass too.
    func scriptAction(t *testing.T, h *testharness.Harness, actionName string) {
    	t.Helper()
    	h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
    		if faction.ID != "alpha" || actionName == "" {
    			return nil, nil
    		}
    		for _, candidate := range available {
    			if candidate.Name() == actionName {
    				return candidate, nil
    			}
    		}
    		t.Fatalf("%s not available for alpha", actionName)
    		return nil, nil
    	}
    }
    ```
    **Replace with:**
    ```go
    // scriptAction makes alpha take the named action and every other faction pass.
    // An empty name makes alpha pass too.
    func scriptAction(t *testing.T, h *testharness.Harness, actionName string) {
    	t.Helper()
    	scriptActionBy(t, h, "alpha", actionName)
    }

    // scriptActionBy makes the named faction take the named action and every other
    // faction pass.
    func scriptActionBy(t *testing.T, h *testharness.Harness, factionID, actionName string) {
    	t.Helper()
    	h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
    		if faction.ID != factionID || actionName == "" {
    			return nil, nil
    		}
    		for _, candidate := range available {
    			if candidate.Name() == actionName {
    				return candidate, nil
    			}
    		}
    		t.Fatalf("%s not available for %s", actionName, factionID)
    		return nil, nil
    	}
    }
    ```
  - Add `TestDestroyTheFoe_ThirdPartyBlowFailsTheGoal`, under a
    `// --- Destroy the Foe (D-15, D-21) ---` head:
    - alpha holds Destroy the Foe with `TargetFactionID` "beta". beta is on
      Tartarus with 1 faction HP and a homeworld Base (as in
      `TestBloodTheEnemy_BaseHitCountsOnce`). gamma (Force 3) attacks.
    - `scriptActionBy(t, h, "gamma", "Attack")`; the defender script picks beta's
      asset, and the redirect to the Base is confirmed. Dice `[10, 1, 2]`: a hit for
      3 on the Base takes beta from 1 HP to -2.
    - After one cycle: alpha's goal is cleared, alpha's XP is unchanged, history
      holds a `goal_failed` for alpha with cause `target_lost`, and no
      `goal_completed`.

##### Commit message

```
feat(E2): a lost target fails the goal

- GoalFailed{FactionID, GoalID, Cause} clears the goal, pays no XP and costs
  no income; target_lost is its one cause (D-21)
- a goal with a target_faction input fails when another faction destroys the
  target and its own complete did not match; evaluate splits into advance and
  the lost-target rule
- the digest records it on the holder's beat with its cause; the wire line is
  chosen by cause; no headline for target_lost
- the feed line names the holder and the cause
```

#### Commit 3 — `docs(E2): goal rules for every batch and a lost target`

##### Task 1 — `docs/rules/swn-faction-mechanics.md`

The readings of D-13 and D-21 (Decision 12).

**1a. Find:**
```markdown
- **What counts as destroying or damaging** — any kill or damage the faction
  causes: an attack, a counterattack, an ability or an asset's special effect.
```
**Replace with:**
```markdown
- **When goals count** — a goal counts what happens on any faction's turn and in
  any phase of it, not only during its holder's action. A kill made while
  defending counts, and a goal can complete on another faction's turn.
- **What counts as destroying or damaging** — any kill or damage the faction
  causes: an attack, a counterattack, an ability or an asset's special effect.
```

**1b. Find:**
```markdown
- **Destroy the Foe** — the goal completes only when the faction itself lands the
  blow that takes the target to 0 hit points.
```
**Replace with:**
```markdown
- **Destroy the Foe** — the goal completes only when the faction itself lands the
  blow that takes the target to 0 hit points. When another faction lands that
  blow, the goal fails: it ends with no XP, and the faction loses no income.
```

##### Task 2 — tick the boxes

- In this plan's Overview, tick Phase 1: `- [ ] **Phase 1 — Every batch**` becomes
  `- [x] **Phase 1 — Every batch**`.
- In the overview plan's Effort Summary, tick Effort 2's Done cell: `| [ ] | 2 |`
  becomes `| [x] | 2 |`.

##### Commit message

```
docs(E2): goal rules for every batch and a lost target

- rules doc: goals count on any turn and in any phase (D-13)
- rules doc: Destroy the Foe fails when another faction lands the blow (D-21)
- tick Phase 1 and Effort 2
```

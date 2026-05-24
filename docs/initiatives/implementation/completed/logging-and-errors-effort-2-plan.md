# Logging + Errors — Effort 2: Errors

`F-003`, effort 2 of 2. Implements the error-handling conventions described in the discovery doc's [Errors](../discovery/logging-and-errors-discovery.md#errors) section.

## Context / Goal

- Top-level plan: [`logging-and-errors-plan.md`](./logging-and-errors-plan.md)
- Discovery doc: [`logging-and-errors-discovery.md`](../discovery/logging-and-errors-discovery.md)
- Effort 1 (logging) is a prerequisite: this effort calls `phaseLog.Warn` / `phaseLog.Error` at the orchestrator's Recoverable/Fatal branch and emits per-miss `Error` lines at the `MutationApplyError` consumer. Effort 2 assumes the logger field, cascade, and `applyAndRecord` logger parameter are already in place.

Declare promoted sentinels, demote the audit's six bare-string `fmt.Errorf`s, wire the orchestrator's Recoverable/Fatal branch around every sub-engine call, and consume `MutationApplyError`'s structured payload at `applyAndRecord` with per-miss diagnostic log lines. After this effort lands, recoverable errors (collector-aborted required selections) log `Warn` and the turn continues; everything else logs `Error` and propagates.

## Decisions Ratified in Planning

### Q1 — Sentinels to declare (two, not three)

Two sentinels are declared in this effort:

- `engine.ErrWorldEngineUnavailable` — collapses the two `fmt.Errorf("world engine not found")` duplicates at `orchestrator.go:211, :290`. **Fatal** (engine-integrity failure).
- `action.ErrNoSelection` — collapses the four `"no <X> selected"` leaves at `change_homeworld.go:52`, `seize_planet.go:47`, `sell_asset.go:38`, `ability/informers.go:28`. **Recoverable**.

**`action.ErrPreconditionFailed` is deferred.** Discovery floated this as a third sentinel for the leaf-origin action precondition errors (insufficient Coin, unknown sub-mode, missing target). Adding it would require retroactively wrapping 5–10 audit-cleared leaf sites with `%w: ErrPreconditionFailed` *and* would change behavior — action precondition failures would become skip-and-continue rather than turn-abort. Today every action.Run error aborts. F-003's mandate is a convention layer; making a behavior change to action-precondition handling belongs to a future initiative that has it as the actual goal. The recoverable list ships with one entry; growing it is a deliberate later step.

### Q2 — One shared `action.ErrNoSelection`

Single sentinel for all four sites. The four leaves are semantically identical (collector returned no selection for a required field); declaring four distinct sentinels would inflate the recoverable list and pre-bake per-action granularity F-005 hasn't asked for. If the TUI later needs per-action rendering distinction, per-action sentinels can be layered on top via `%w: ErrNoSelection` wrapping — additive, not retrofitted.

Location: `internal/faction/engine/action/errors.go` (new file, package `action`). Sub-packages `actions/` and `actions/ability/` import `action` and reference `action.ErrNoSelection`.

### Q3 — Direct field access at the `MutationApplyError` consumer

`Mutation.Apply` returns `*MutationApplyError` (concrete pointer) at `mutation.go:44`, not `error`. At `applyAndRecord`'s call site (`orchestrator.go:359`), `err` is already the typed value before any wrap. The consumer uses direct field access — no `errors.As` ceremony at this site:

```go
if err := e.Mutation.Apply(factionState, mutations); err != nil {
    for _, miss := range err.Misses {
        log.Error("mutation miss",
            "mutation_type", miss.MutationType,
            "faction_id", miss.FactionID,
            "entity_id", miss.EntityID,
        )
    }
    return fmt.Errorf("applying mutations: %w", err)
}
```

The `%w` wrap continues to preserve the typed payload in the chain for any future downstream consumer. `errors.As` remains the canonical idiom for *that* hypothetical consumer (if it ever arrives), but Effort 2 doesn't add one.

### Q4 — `engine/recoverable.go`

Dedicated file holds the recoverable list and the helper:

```go
// internal/faction/engine/recoverable.go
package engine

import (
    "errors"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
)

// recoverableErrors enumerates sentinels whose return value is treated as a
// gameplay outcome rather than an engine failure. The orchestrator's
// Recoverable/Fatal branch logs at Warn level and continues past the failing
// sub-engine call for any error matching one of these via errors.Is.
//
// Adding to this list is a deliberate decision — it changes the orchestrator's
// abort semantics for the new error category.
var recoverableErrors = []error{
    action.ErrNoSelection,
}

func isRecoverable(err error) bool {
    for _, sentinel := range recoverableErrors {
        if errors.Is(err, sentinel) {
            return true
        }
    }
    return false
}
```

Unexported — both `recoverableErrors` and `isRecoverable` are package-internal. The orchestrator is the only caller. Keeping the helper out of `orchestrator.go` (~520 lines) preserves that file's focus on phase pipelines.

### Q5 — Phase order: A → B → C (sentinels → branch → consumer)

Three commits, in this order:

| # | Commit | Depends on | Recommended model |
|---|--------|------------|-------------------|
| A | `feat(errors): declare promoted sentinels and demote bare-string fmt.Errorfs` | — | Sonnet |
| B | `feat(errors): wire Recoverable/Fatal branch into orchestrator` | A (references `action.ErrNoSelection`) | Opus |
| C | `feat(errors): consume MutationApplyError payload at applyAndRecord` | Effort 1 (logger on `applyAndRecord`) | Sonnet |

Sentinels first — small, mechanical, unblocks B. The orchestrator branch is the substantive change and lands second. The typed-error consumer (independent of A and B) lands last as a small leaf addition.

## Open Questions — To Ratify at Implementation Time

- **Helper for the per-call-site surfacing pattern.** Commit B touches 26 error sites in `orchestrator.go`, each doing the same `if isRecoverable(err) { log.Warn + OnError + return nil } else { log.Error + OnError + return err }` scaffolding. A small helper (`e.surface(log, faction, observer, err) bool` returning the recoverability bool) collapses each site to two lines. Inline keeps control flow visible per site; helper saves ~80 lines. Open: ratify at execution time after seeing the first three call sites converted both ways. Working assumption: introduce the helper if the inline pattern feels mechanical and repetitive in practice.
- **Prefix-discipline cleanup at the two `world engine not found` sites.** Today neither site at `orchestrator.go:211, :290` calls `observer.OnError` before returning — the inconsistency discovery flagged. Promotion to `engine.ErrWorldEngineUnavailable` plus Commit B's unified branch fixes the inconsistency automatically (the branch always fires `OnError`). No standalone fix needed; flagged so it doesn't get touched twice.
- **Whether `runActionPhase` should fire `OnActionResolved` on a recoverable skip.** Today `selectedAction == nil` triggers `OnFactionSkipped` and returns `nil` (`orchestrator.go:275-278`). Under Commit B, an `ErrNoSelection` return from `e.Action.Run(...)` becomes recoverable — should that path emit `OnFactionSkipped` too, or stay silent on the observer (since `Action.Run` already partially executed)? Working assumption: stay silent — `OnError` already fires via the branch and that's the only observer notification a recoverable action-skip warrants. Confirm at execution time when wiring `runActionPhase`.
- **Test coverage shape for the recoverable path.** No existing test exercises a recoverable error scenario (there were no recoverable errors before this effort). At least one orchestrator-level test should drive a collector to return `action.ErrNoSelection` and assert the turn continues. Test scaffolding details settled at execution time; commit B's task list reserves room for it.

## Work Breakdown

3 commits, each = one execution session. Branch: `feature/logging-and-errors` (shared with Effort 1).

### Commit 1 — `feat(errors): declare promoted sentinels and demote bare-string fmt.Errorfs`

**Recommended model:** Sonnet. Mechanical: declare two sentinels, demote six call sites.

#### Task 1 — `internal/faction/engine/action/errors.go` (new file)

```go
package action

import "errors"

// ErrNoSelection is returned by an action's Run method when the collector
// produced no value for a field the action requires (target world, target
// faction, asset to sell, etc.). Recoverable at the orchestrator: a gameplay
// outcome ("the faction tried X, no selection was made"), not an engine
// failure.
var ErrNoSelection = errors.New("action: no selection")
```

The `"action: "` prefix follows discovery's optional package-ish scope (3rd guideline of [Prefix discipline](../discovery/logging-and-errors-discovery.md#prefix-discipline)) — useful for standalone readability when the error is logged before any wrap context is added.

#### Task 2 — `internal/faction/engine/core.go`

Add `ErrWorldEngineUnavailable` to the existing `var (...)` block at `core.go:20-22`:

```go
var (
    ErrSpatialDataDirRequired = errors.New("spatial data dir is required for world engine")
    ErrWorldEngineUnavailable = errors.New("engine: world engine unavailable")
)
```

#### Task 3 — Demote bare-string sites (six total)

Four `action.ErrNoSelection` demotions. Each site replaces a bare-string `fmt.Errorf` with the sentinel-of-`%w` pattern (the audit's expressive `spatial`-style shape):

| File:line | Before | After |
|-----------|--------|-------|
| `engine/action/actions/change_homeworld.go:52` | `fmt.Errorf("change homeworld: no target selected")` | `fmt.Errorf("change homeworld: %w", action.ErrNoSelection)` |
| `engine/action/actions/seize_planet.go:47` | `fmt.Errorf("seize planet: no target world selected")` | `fmt.Errorf("seize planet: %w", action.ErrNoSelection)` |
| `engine/action/actions/sell_asset.go:38` | `fmt.Errorf("sell asset: no asset selected")` | `fmt.Errorf("sell asset: %w", action.ErrNoSelection)` |
| `engine/action/actions/ability/informers.go:28` | `fmt.Errorf("informers: no target faction selected")` | `fmt.Errorf("informers: %w", action.ErrNoSelection)` |

The operation-name prefix is retained per discovery's [Prefix discipline](../discovery/logging-and-errors-discovery.md#prefix-discipline) rule 1 ("prefer operation/function name").

Add the import `"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"` to each site. (Verify import-cycle safety at execution time: `action/actions/` and `action/actions/ability/` should import `action` cleanly — they're sub-packages of `action`, no cycle.)

Two `ErrWorldEngineUnavailable` demotions:

| File:line | Before | After |
|-----------|--------|-------|
| `engine/orchestrator.go:211` | `return fmt.Errorf("world engine not found")` | `return ErrWorldEngineUnavailable` |
| `engine/orchestrator.go:290` | `return fmt.Errorf("world engine not found")` | `return ErrWorldEngineUnavailable` |

Both sites currently lack an `observer.OnError` call (the inconsistency discovery flagged). Don't fix that here — Commit B's unified branch will handle it via the new pattern. Resist the urge to add `observer.OnError(faction, err)` inline; the branch will do it.

#### Task 4 — Tests

- `internal/faction/engine/action/errors_test.go` — minimal `errors.Is` smoke test: a wrap with `%w` passes `errors.Is(err, action.ErrNoSelection)`.
- Existing action tests: scan `internal/faction/engine/action/actions/**/*_test.go` for any assertion that string-matches the old messages (e.g., `assert.EqualError(err, "change homeworld: no target selected")`). Convert to `assert.ErrorIs(err, action.ErrNoSelection)` where they exist. If none exist (likely — the audit found no consumers of these errors), no test changes are needed beyond the new smoke test.
- No new tests for `ErrWorldEngineUnavailable` in this commit — Commit B's orchestrator tests cover the branch behavior. A bare `errors.Is` smoke test isn't load-bearing for an unexported-target sentinel.

---

### Commit 2 — `feat(errors): wire Recoverable/Fatal branch into orchestrator`

**Recommended model:** Opus. The substantive change: replacing 26 error sites with the unified branch, deciding helper-vs-inline at the first-three-site mark, and resolving the recoverable-path observer semantics in `runActionPhase`.

#### Task 1 — `internal/faction/engine/recoverable.go` (new file)

Exactly the file from Q4 above. Initial list contains one entry (`action.ErrNoSelection`).

#### Task 2 — Convert the 26 orchestrator error sites to the unified branch

Every site in `orchestrator.go` that today reads:

```go
if err := someSubEngine(...); err != nil {
    observer.OnError(faction, err)
    return err  // or: return fmt.Errorf("prefix: %w", err)
}
```

…becomes:

```go
if err := someSubEngine(...); err != nil {
    if isRecoverable(err) {
        phaseLog.Warn("recoverable failure", "err", err)
        observer.OnError(faction, err)
        return nil  // continue past this sub-engine call
    }
    phaseLog.Error("failure", "err", err)
    observer.OnError(faction, err)
    return err  // or: return fmt.Errorf("prefix: %w", err)
}
```

Execution-time decision: after converting the first three sites, evaluate whether the inline scaffolding is mechanical enough to extract a helper:

```go
// candidate helper, if extracted
func (e *Engine) surface(log *slog.Logger, faction *domain.Faction, observer TurnObserver, err error) bool {
    if isRecoverable(err) {
        log.Warn("recoverable failure", "err", err)
        observer.OnError(faction, err)
        return true
    }
    log.Error("failure", "err", err)
    observer.OnError(faction, err)
    return false
}
```

If extracted, each site reduces to:

```go
if err := someSubEngine(...); err != nil {
    if e.surface(phaseLog, faction, observer, err) {
        return nil
    }
    return err  // or wrapped
}
```

Decide once, apply to all 26 sites. Don't mix patterns within `orchestrator.go`.

**Logger naming at each site.** Use the phase-bound logger that Effort 1 introduced (`phaseLog`, `turnLog`, or `engineLog` depending on scope at the site). The cascade has already bound `turn`, `faction`, `phase`, so the `Warn` / `Error` line picks up that context automatically; the only fresh attr per site is `err`.

**Site inventory** (26 from audit). Verify against current `orchestrator.go` at execution time; line numbers will drift after Effort 1's edits. Use LSP `goToReferences` on `OnError` to enumerate:

- `setupFactionTurn` — 2 sites (`World.RebuildIndex`, `Turn.CurrentFaction`). `nil` faction.
- `runGoalLockPhase` — 2 sites (`applyAndRecord`, `AwaitCheckpoint`).
- `runBookkeepingPhase` — 3 sites (`Turn.ApplyBookkeeping`, `applyAndRecord`, `AwaitCheckpoint`).
- `runStatRaisePhase` — 2 sites (`prepareStatRaise`, `applyAndRecord`).
- `runMovementPhase` — 7 sites (incl. the `ErrWorldEngineUnavailable` early-return, two `dispatch.MutationReactors`, two `applyAndRecord`, `prepareMovementDecisions`, `AwaitCheckpoint`).
- `runActionPhase` — 6 sites (incl. the second `ErrWorldEngineUnavailable` early-return, `SelectAction`, `Action.Run`, `dispatch.MutationReactors`, `applyAndRecord`, `AwaitCheckpoint`).
- `finishFactionTurn` — 3 sites (`Turn.Advance`, `state.Save`, `AwaitCheckpoint`).
- Direct `state.Save` calls in `RunFactionTurn` — 1 site (the `"saving state after bookkeeping"` site at line 79).

Cross-check the total against the audit's 26 at execution time; the two `world engine not found` sites add to that count because Commit 1 turned them into bare `return ErrWorldEngineUnavailable` (no `OnError` today — the branch now wraps them too).

#### Task 3 — Recoverable-path semantics in `runActionPhase`

Special case: when `e.Action.Run(...)` returns `ErrNoSelection`, the recoverable branch fires `Warn` + `OnError` and returns `nil`. Open question listed above: does `OnFactionSkipped` also fire? Working assumption: no — the new `OnError` call is the single observer signal for a recoverable action skip. The existing `selectedAction == nil` path (lines 275-278) keeps its `OnFactionSkipped` call because that's a different flow (collector returned no action, action.Run was never called).

Ratify at execution time after writing the test from Task 5 — the test will surface whether the observer trail makes sense.

#### Task 4 — Recoverable-path semantics in collector sites

`prepareStatRaise` (line 184), `prepareMovementDecisions` (line 232), `SelectAction` (line 270), `SelectTransportCargo` (line 470), and `Phase.AwaitCheckpoint` calls currently return their own (non-`ErrNoSelection`) errors. None of these will match `isRecoverable` today — they all flow through the Fatal branch and propagate. No behavior change at those sites; the branch just runs them through the standard `Error` log + propagate path.

If the TUI later wants to make collector aborts recoverable (e.g., GM cancels mid-selection), declaring a new sentinel (`collector.ErrAborted`?) and adding it to the recoverable list is the path — out of scope here.

#### Task 5 — Tests

Two new tests in `internal/faction/engine/`:

- `recoverable_test.go` — `isRecoverable` unit test: passes for `action.ErrNoSelection` and a `fmt.Errorf("%w", action.ErrNoSelection)` wrap; fails for arbitrary errors.
- An orchestrator-level integration test (location: existing `engine/orchestrator_test.go` if present, or new file) that drives a `RunFactionTurn` through a recoverable scenario:
  - Set up a collector that produces a `selectedAction` whose `Run` returns `ErrNoSelection` (e.g., `sell_asset` with no asset selection).
  - Assert: `RunFactionTurn` returns `nil` (turn continues).
  - Assert: `observer.Events` contains the `Error` event for `ErrNoSelection`.
  - Assert: the `OnActionResolved` event is *not* recorded.
  - Assert: `OnFactionTurnCompleted` *is* recorded (finishFactionTurn ran).

Existing tests asserting on full propagation of these errors (if any) need their assertions updated — the action path no longer aborts on `ErrNoSelection`. Scan via LSP `goToReferences` on each demoted action's name at test-write time.

---

### Commit 3 — `feat(errors): consume MutationApplyError payload at applyAndRecord`

**Recommended model:** Sonnet. Small, contained change at one site plus a focused test.

#### Task 1 — `internal/faction/engine/orchestrator.go`: `applyAndRecord`

After Effort 1, `applyAndRecord`'s signature includes `log *slog.Logger`. Convert the `Mutation.Apply` call:

```go
// Before (current orchestrator.go:359-361):
if err := e.Mutation.Apply(factionState, mutations); err != nil {
    return fmt.Errorf("applying mutations: %w", err)
}

// After:
if err := e.Mutation.Apply(factionState, mutations); err != nil {
    for _, miss := range err.Misses {
        log.Error("mutation miss",
            "mutation_type", miss.MutationType,
            "faction_id", miss.FactionID,
            "entity_id", miss.EntityID,
        )
    }
    return fmt.Errorf("applying mutations: %w", err)
}
```

Direct field access (per Q3) — `err` is already `*MutationApplyError`. No `errors.As` here.

`miss.EntityID` is documented as "empty for faction-level misses" (`mutation.go:15`). The `entity_id` attr will be emitted as an empty string for those — acceptable per the canonical attr vocabulary (consumer can distinguish empty vs. populated). Don't conditionally suppress the attr; uniform shape is more grep-friendly.

#### Task 2 — Resolve the aspirational comment

`mutation.go:18-20` reads:

```go
// MutationApplyError is returned by Apply when one or more referenced entities
// were not found. The orchestrator decides whether to treat this as a hard turn
// failure or a warning.
```

The decision is now made — and it's "log per miss at Error level, abort the turn." Update the comment to reflect the ratified semantics:

```go
// MutationApplyError is returned by Apply when one or more referenced entities
// were not found. Fatal at the orchestrator: applyAndRecord emits one
// structured Error-level log line per Miss and propagates the error to abort
// the turn.
```

#### Task 3 — Tests

- `internal/faction/engine/orchestrator_apply_test.go` (or extend an existing orchestrator test file) — feed `applyAndRecord` a mutation that references a missing entity, capture log output via a buffer-backed `slog.NewTextHandler`, assert:
  - The function returns a non-nil error wrapping `*MutationApplyError`.
  - The log buffer contains an `Error`-level line with `mutation_type`, `faction_id`, `entity_id` attrs matching the miss.
  - For multi-miss scenarios: N log lines for N misses.

Use the buffer-handler pattern documented in the top-level plan's [Testharness logger](./logging-and-errors-plan.md#testharness-logger) section.

---

## End of Effort 2

After Commit 3 lands, both efforts on `feature/logging-and-errors` are complete. The pre-merge checklist (code review, dev journal update, planned-work update, decisions log update) runs once at branch level per `feedback_branch_initiative_layering`. Squash to two conventional commits (`feat(logging): ...` and `feat(errors): ...`) — or a single `feat: logging + structured errors` — at merge time.

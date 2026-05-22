# Movement Phase — Tick Before Decisions

> Branch: `feature/movement-redesign` · Single-commit fix-up between Effort 2 Phase 3 and Phase 4. Conventional commit: `wip: fix: apply movement ticks before building decisions`.

## Why

`runMovementPhase` (`internal/faction/engine/orchestrator.go:194-234`) currently:

1. Builds `tickMutations` (Progressed / Completed)
2. Builds `decisionMutations` (Issue / Revise / Cancel) reading **pre-tick** asset state
3. Applies both together in one `applyAndRecord`

Because step 2 reads `asset.Location` before step 1 has been applied, a Revise built on a tick turn originates its new `Path` from the pre-tick hex. After Apply, `asset.Location` (set by Progressed) and `CurrentOrder.Path[0]` (built from pre-tick hex) disagree — the asset "jumps" on the next tick.

## Change

Split the phase into two apply cycles:

```go
// 1. Tick — physics, no collector input
tickMutations, _ := e.World.TickMovementOrders(faction, e.Rulebook)
if len(tickMutations) > 0 {
    applyAndRecord(factionState, faction, tickMutations, cfg)
    observer.OnMovementTicked(faction, tickMutations)
}

// 2. Decisions — collector sees post-tick state
decisionMutations, _ := prepareMovementDecisions(faction, ...)
if len(decisionMutations) > 0 {
    applyAndRecord(factionState, faction, decisionMutations, cfg)
}
observer.OnMovementResolved(faction, decisionMutations)   // phase-boundary marker; fires unconditionally
```

`OnMovementTicked` fires only when ticks happened. `OnMovementResolved` keeps its phase-boundary semantics and fires unconditionally, but now its payload is only the decision mutations.

## Files to touch

| File | Change |
|---|---|
| `internal/faction/engine/observer.go` | Add `OnMovementTicked(faction, mutations)` to `TurnObserver` interface |
| `internal/faction/engine/orchestrator.go` | Restructure `runMovementPhase` per above |
| `internal/faction/engine/world/movement.go` | In `BuildMovementMutations`, add nil-check on `asset.CurrentOrder` in Revise/Cancel branches — a same-turn Completed tick can clear it before the decision is built. Return an error if collector asks to revise/cancel an orderless asset |
| `internal/faction/engine/testharness/observer.go` | Implement `OnMovementTicked` on `RecordingObserver` — append `ObservedEvent{Kind: "MovementTicked", ...}` |
| `internal/faction/engine/testharness/scenarios/full_cycle_test.go` | Update `wantKinds` to include `MovementTicked` where applicable (only fires when a faction has in-flight orders — which is never in that test, so no change likely needed; verify) |
| `internal/faction/engine/testharness/scenarios/movement_lockskip_test.go` | `TestLockSkip_InFlightOrderTicks` — verify history assertion still finds Progressed; should be fine since both events still record |
| `internal/faction/engine/testharness/scenarios/movement_lifecycle_test.go` | The three tests likely still pass — verify. The cancel test's `lifecyclePath[2]` assertion now reflects post-tick Apply (was already true under old ordering; reasoning is now cleaner) |

## New test to add

In `movement_lifecycle_test.go`, add a test that exercises the fix: post-tick decision build.

- Issue (path length 7, Speed 2)
- Tick on T2 — asset at `Path[2]`
- Revise on T3 — assert `Revised.NewOrder.Path[0] == Path[2]` (post-tick hex), not `Path[0]` (original origin)

This is the test the old ordering would have failed.

## Open question to resolve in-session

`OnMovementTicked` event naming. Alternatives: `OnMovementProgressed` (matches mutation type) / `OnMovementOrdersAdvanced`. Pick what reads clearly alongside `OnMovementResolved` — I lean `OnMovementTicked` because "Progressed" already names a specific mutation.

## Pre-commit checklist

1. `go test ./...` green
2. No decisions-log entry needed (per session that landed Commit 5 — Robert confirmed)
3. Stage `observer.go`, `orchestrator.go`, `world/movement.go`, `testharness/observer.go`, scenario tests touched, this brief
4. Move this file to `docs/initiatives/implementation/completed/` after commit
5. Commit message: `wip: fix: apply movement ticks before building decisions`

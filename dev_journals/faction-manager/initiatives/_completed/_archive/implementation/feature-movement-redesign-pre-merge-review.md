# feature/movement-redesign — Pre-Merge Code Review

**Reviewer:** Claude (Opus 4.7)
**Branch:** `feature/movement-redesign` (38 commits ahead of `main`)
**Scope:** 104 files changed, +7,890 / −2,406 LOC
**Date:** 2026-05-22
**Build/test status:** `go build ./...` clean; `go test ./...` all 20 testable packages pass

## Scope of this review

The branch bundles six distinct initiatives that landed together:

1. **Spatial library redesign** — `hybrid_map` → `region_map`, `RegionHex` unification, narrowed `SpatialMap` interface, `HexRouter` consumer-side composition
2. **Input/PhaseCollector refactor** — split orchestrator-owned prompts from sub-engine prompts
3. **Ability engine redesign (R-001)** — retired `engine/ability/` sub-engine; moved to per-asset dispatch under `action/actions/ability/`
4. **Asset Movement Redesign (F-002)** — `Location` struct, `MovementOrder`, `Speed`, new Movement Phase in orchestrator
5. **Transport system** — `TransportProfile`, `TransportReactor`, `SelectTransportCargo` collector, `EffectsEngine`
6. **Hex-scoped attack targeting + ChangeHomeworld action**

This review intentionally **does not** re-cover ground from `CODEBASE_REVIEW.md` (commissioned earlier on this same branch). It focuses on what was added after that review and on the architectural coherence of the bundled work.

---

## Verdict

**Mergeable, with two cheap fixes recommended before merge.**

- The engine refactors are disciplined; the test coverage is real (≈2,057 LOC of scenario tests, including 810 LOC newly added for movement / transport / homeworld).
- The orchestrator's phase-extracted shape is clean and the `PhaseCollector` / `action.Collector` split is principled.
- Two correctness items (C1, C2) should be fixed before merge — both are mechanical, under 10 minutes combined.
- One architectural gap (C3, transport reactor wiring) is already deferred per Decision #234, but is now load-bearing for any production driver. Worth surfacing in planned-work.md.

---

## Critical / Architectural Concerns

### C1 — Movement mutations missing JSON struct tags

**Files:** `internal/faction/domain/mutation.go:309–359`

The five new mutation types (`MovementOrderIssued`, `MovementOrderProgressed`, `MovementOrderRevised`, `MovementOrderCancelled`, `MovementOrderCompleted`) have no `json:"..."` tags, while every other mutation in the file uses snake_case JSON tags.

**Why this matters:** History records for movement events will serialize with default Go field names (`"FactionID"` instead of `"faction_id"`). Inconsistent with the rest of the audit trail. The digest builder (per `CODEBASE_REVIEW.md` section 4.8) is built around cause-keyed disambiguation of mutation history — the moment anything reads movement history records, the casing mismatch will surface as a bug.

**Suggested fix:** Add JSON tags matching the convention used for the other 28 mutations. Pattern:
```go
type MovementOrderIssued struct {
    FactionID         string        `json:"faction_id"`
    AssetID           string        `json:"asset_id"`
    Order             MovementOrder `json:"order"`
    Cause             string        `json:"cause"`
    CausedByFactionID string        `json:"caused_by_faction_id"`
}
```

(And consider whether `MovementOrder` itself, which currently has only `toml` tags, should also gain `json` tags now that it appears inside history records.)

---

### C2 — Movement mutations missing `Cause` on the happy path

**Files:** `internal/faction/engine/world/movement.go:97–102, 127–132`; `internal/faction/engine/action/actions/change_homeworld.go:73–79`

In `BuildMovementMutations`:
- `MovementOrderIssued` is constructed with **no `Cause` field set** (line 97)
- `MovementOrderRevised` is constructed with no `Cause` field set (line 127)

Both structs have a `Cause` field (mutation.go:313, 335). Only the Tick and Cancel paths set it.

In `ChangeHomeworld.Output`:
- `GoalPhaseAdvanced` is constructed with no `Cause` field set (line 73)

**Why this matters:** The digest builder uses `Cause` to disambiguate composite events. The "movement_order_issued" and "movement_order_revised" history rows will land with `"cause": ""`. The system silently degrades: no test fails, but the narrative renderer (or any future cause-aware consumer) loses information.

**Suggested fix:**
- Define `CauseMovementIssued` in `domain/mutation.go` to sit alongside the existing `CauseMovementTick`/`CauseMovementRevision`/`CauseMovementCancellation` constants
- Set `Cause: domain.CauseMovementIssued` and `Cause: domain.CauseMovementRevision` in the respective `BuildMovementMutations` branches
- Set `Cause: "change_homeworld"` (or a domain constant) on the `GoalPhaseAdvanced` in `ChangeHomeworld.Output`

---

### C3 — Transport reactors not registered in production wiring (deferred but now load-bearing)

**Files:** `internal/faction/engine/core.go:68–72`, `internal/faction/engine/orchestrator.go`, `internal/faction/engine/testharness/harness.go:103`

`engine.NewWithRulebook` calls `e.Effect.Register(...)` for each transport-capable asset definition. But `EffectsEngine.Register` only adds a `Handler` to the engine's map. The actual hook-registry registration happens inside `Handler.Apply`, which is only invoked from `EffectsEngine.ApplyAll`.

`ApplyAll` is called from exactly one place in the codebase:
```
internal/faction/engine/testharness/harness.go:103
```

The orchestrator never calls it. Same issue applies to `TagEngine.ApplyAll` (Scavengers reactor) — but that predates this branch.

**Why this matters:** In tests, transport behaves correctly because the harness wires it up. In production (when a CLI/TUI materializes), `engine.New(cfg)` returns an engine whose transport reactors will *never* register — cargo will silently fail to follow the transport, with no error path. The orchestrator looks complete but isn't.

This is **acknowledged** in Decision #234: production wiring is deferred to F-004 (CLI Rebuild). But this branch *adds the second consumer* of the pattern (transport, now joined to tags) without addressing the gap. Whatever wires up the CLI eventually now has two `ApplyAll`s to remember — and the cost of forgetting either is "silent feature failure," not a crash.

**Suggested fix (pick one):**
- **Cheapest:** Add an `Engine.PrepareForRun(factionState)` method that calls all `ApplyAll`s, and require callers to invoke it before `RunCycle`. Make the testharness call it too — then there's one production seam.
- **Most self-contained:** Call `Effect.ApplyAll` and `Tag.ApplyAll` from `orchestrator.setupFactionTurn` (idempotent over the cycle anyway, since registrations are keyed by source).
- **At minimum:** Add a comment in `core.NewWithRulebook` and `setupFactionTurn` documenting the wiring requirement.

This is **not** a blocker for merging the branch (the test scenarios cover transport correctness), but it should be a tracked item in planned-work.md, ideally as a blocker on F-004.

---

### C4 — `runMovementPhase` errors when `e.World == nil`, but other phases tolerate it

**File:** `internal/faction/engine/orchestrator.go:202–204`

```go
if e.World == nil {
    return fmt.Errorf("world engine not found")
}
```

`runActionPhase` (lines 280–283) gracefully accepts `nil` World and passes `worldIndex = nil`. `core.New` permits `cfg.SpatialDataDir == ""` to skip World construction (lines 50–56). The composition root supports nil-World; the movement phase does not.

**Why this matters:** Today, the test harness always provides a World (via `StubSpatialMap`), so this never triggers. But any production caller that lacks spatial config will fail at movement with the message "world engine not found" — which is misleading: the engine *is* found, it's just nil because no spatial dir was configured. More structurally: the codebase is split on whether World is optional.

**Suggested fix:** Decide one way:
- **Optional:** `if e.World == nil { return nil }` — skip the phase silently. Document in the doc comment that movement is a no-op without a World engine.
- **Required:** Remove the `e.World == nil` checks from `runActionPhase` and `core.New`. Make `SpatialDataDir` a required config field.

---

## Architectural Observations (not blocking)

### A1 — Stat-raise phase lacks observer + checkpoint on decline

**File:** `internal/faction/engine/orchestrator.go:173–193`

`runStatRaisePhase` only fires `OnStatRaiseApplied` and `AwaitCheckpoint` if mutations were emitted. If the player has eligible raises but declines (`prepareStatRaise` returns `stat == nil`), no observation occurs and no checkpoint hits.

Compare `runActionPhase` (line 268), which calls `OnFactionSkipped` on a nil selection. The pattern is asymmetric — either both phases observe a skip, or neither.

This is pre-existing (the stat raise refactor happened earlier), but the new movement phase reproduces the same shape ("only observe/checkpoint when activity happened"), which now creates a wider inconsistency with the action phase.

**Suggestion:** Decide on the canonical pattern, apply uniformly across all phases that prompt the collector.

### A2 — Hardcoded goal ID `"G-012"` in `ChangeHomeworld` action

**File:** `internal/faction/engine/action/actions/change_homeworld.go:67, 75`

The literal `"G-012"` appears twice. If goal IDs ever get renumbered, `ChangeHomeworld` silently issues a `GoalInitiated` for the wrong goal.

**Suggestion:** Define `domain.GoalIDChangeHomeworld = "G-012"` and reference it. Or load the ID from the rulebook (the action already has a rulebook ref).

### A3 — Cargo can be issued an independent movement order while on transport

**File:** `internal/faction/engine/orchestrator.go:483–492` (`eligibleMovableAssets`)

A cargo asset whose Location has been teleported by `AssetMoved` still has `CurrentOrder == nil`. If its definition has `Speed > 0`, it appears in `eligibleMovableAssets` and the player can issue a separate movement order for it — effectively "unloading" mid-flight at the current hex.

Is this intended (a tactical option: "drop off at the current hex")? If yes, document. If no, exclude assets currently appearing in any transport's `CurrentOrder.CargoAssetIDs`.

I lean **not intentional** — the transport pattern reads as "cargo follows the transport, full stop." Allowing the cargo to break out of that without a corresponding `transport_unload` decision lets state drift.

### A4 — `liveDefendersOnWorld` appears unused after hex-scoping

**File:** `internal/faction/engine/action/actions/attack.go:241–243`

Defined alongside `liveDefendersAtHex`, but I didn't find a caller after the hex-scoping refactor. If unused, delete (per the codebase's no-dead-code discipline noted in CLAUDE.md and `feedback_unused_code`). If kept for symmetry with the world-scoped path, document the intended caller.

### A5 — Mutation file stylistic inconsistency widened

**File:** `internal/faction/domain/mutation.go`

The new Movement* mutations (and earlier additions `GoalTurnsTick`, `GoalPhaseAdvanced`, `XPSpent`, `StatRaised`) use receiver name `m`; the older mutations use `mutation`. Per `feedback_code_style.md`, full-word parameter names are the project norm.

Additionally, `GoalTurnsTick`, `GoalPhaseAdvanced`, `XPSpent`, `StatRaised` lack a `CausedByFactionID` field. The new Movement* mutations restore it. Worth a design call: is "self-caused mutations skip CausedByFactionID" a deliberate pattern, or accidental drift? If deliberate, document. If not, normalize.

---

## What Looked Good

- **Phase extraction in `orchestrator.go`** — `runGoalLockPhase`, `runStatRaisePhase`, `runMovementPhase`, `runActionPhase`, `finishFactionTurn` all share the same shape. Reads like a pipeline diagram.
- **`HexRouter` consumer-side interface in `world/world.go`** — narrowing `spatial.SpatialMap` to `Location()` only and pushing `Distance`/`Path` into the consumer is exactly the right shape. Decisions #219, #222 articulated this well.
- **`TransportReactor.OnMutations` shape** — five cases for the five movement mutations, plus a documented no-op on `Revised` because cargo manifest carries forward in `BuildMovementMutations`. Clean.
- **`cargoFollows` helper docstring** — names every "not my problem" case the reactor silently skips. Future maintainers don't have to guess.
- **`isCoLocated` is one line, used once, but named** — makes `eligibleCargoForTransport` readable. Tiny but right.
- **Test coverage is honest** — `movement_lifecycle_test.go` (340 LOC), `transport_lifecycle_test.go` (356 LOC), `change_homeworld_lifecycle_test.go` (113 LOC), `movement_lockskip_test.go` (183 LOC). Plus unit-level coverage in `world/movement_test.go`, `world/world_test.go`, `actions/change_homeworld_test.go`, `actions/attack_test.go`. Not theatrical.
- **Data-driven transport handler registration** — `for _, def := range rulebook.Assets { if def.Transport != nil { ... } }` is the right pattern. The CODEBASE_REVIEW already praised it.

---

## Pre-Merge Recommendation

1. **Apply C1 + C2 fixes** before merge. ~10 minutes of mechanical work.
2. **Add C3 to planned-work.md** as a deferred item with explicit "blocks F-004" semantics.
3. **Leave A1–A5 for a follow-up pass** — none are correctness-blocking, but worth a triage session before they ossify further.
4. **Run `go test ./...` one more time after C1 + C2 are applied** to confirm no test relied on the broken serialization.

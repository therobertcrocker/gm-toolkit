# Asset Movement Redesign — Effort 2: Movement Engine

> Overview: [`asset-movement-redesign-plan.md`](asset-movement-redesign-plan.md)
> Depends on: [`asset-movement-redesign-effort-1-plan.md`](asset-movement-redesign-effort-1-plan.md) — all Effort 1 phases must be merged before starting Effort 2.

Two phases. Each is its own execution session.

After Effort 2 lands, the Movement Phase is wired into `RunFactionTurn` and the transport ability step type exists. In practice, the phase is dormant — no asset has `Speed > 0` yet (Effort 3 Phase 5) and no TOML uses `type = "transport"` yet (Effort 3 Phase 5). The collector is offered the issue/revise/cancel prompt every Movement Phase but should return "no orders" by default until orders are programmatically driven by tests or (eventually) a UI.

<br/>
<br/>

# Phase 3 — Movement Phase Orchestrator

Insert a new phase into `RunFactionTurn` between Phase 2B (Stat Raise) and Phase 3 (Action). Resolve in-flight orders by ticking them, emit progress/completion mutations, then offer the collector a window to issue, revise, or cancel orders.

After Phase 3: `go test ./...` passes; new harness tests cover order tick → completion, issue, revise (1 Coin cost), cancel, and `LockSkip` movement-only behavior.

<br/>

## Commit 1 — `feat: add CheckpointMovement and PhaseMovement orchestrator hook`

### Task 1 — `internal/faction/engine/orchestrator.go`

Insert a `runMovementPhase` call between Phase 2B and Phase 3 in `RunFactionTurn`. Skeleton:

```go
// Phase 2.5: Movement Phase. The engine ticks in-flight orders, then offers the
// collector a window to issue, revise, or cancel orders for any movable assets
// owned by this faction.
movementMutations, err := e.runMovementPhase(faction, factionState, collector, lock)
if err != nil {
    observer.OnError(faction, err)
    return false, err
}
if len(movementMutations) > 0 {
    if err := e.applyAndRecord(factionState, faction, movementMutations, cfg); err != nil {
        observer.OnError(faction, err)
        return false, err
    }
}
observer.OnMovementResolved(faction, movementMutations)
if err := collector.AwaitCheckpoint(CheckpointMovement); err != nil {
    observer.OnError(faction, err)
    return false, err
}
```

`runMovementPhase` is implemented in Commit 2. `OnMovementResolved` is a new observer method (add to `TurnObserver` interface; existing implementations no-op).

### Task 2 — `internal/faction/engine/observer.go` (or wherever `TurnObserver` is defined)

Add `OnMovementResolved(faction *domain.Faction, mutations []domain.Mutation)` to the `TurnObserver` interface. Update any existing implementations (test observers, real observer if one is wired) to add a no-op or pass-through method.

### Task 3 — `internal/faction/engine/orchestrator.go` (LockSkip path)

The existing `lock.Type == goal.LockSkip` branch (line ~56) currently routes directly to `finishFactionTurn` without going through bookkeeping. Per Q5 (LockSkip allows movement), this branch needs to run the Movement Phase before finishing:

```go
if lock.Type == goal.LockSkip {
    // Apply any lock mutations
    if len(lockMutations) > 0 { ... }
    if err := collector.AwaitCheckpoint(CheckpointGoalLocked); err != nil { ... }

    // Run Movement Phase even when locked — assets can still tick / issue
    movementMutations, err := e.runMovementPhase(faction, factionState, collector, lock)
    // ... apply, observe, checkpoint as above ...

    return e.finishFactionTurn(...)
}
```

Note: bookkeeping does **not** run during `LockSkip` (existing behavior). Per the rules, a locked faction's stat raise and bookkeeping are skipped — but movement is allowed. Keep that asymmetry; flag for Robert at execution time if the implementation feels wrong.

<br/>

## Commit 2 — `feat: tick in-flight orders and emit progress/completion mutations`

### Task 4 — `internal/faction/engine/orchestrator.go` (new helper)

```go
func (e *Engine) runMovementPhase(
    faction *domain.Faction,
    factionState *state.FactionState,
    collector InputCollector,
    lock goal.GoalLock,
) ([]domain.Mutation, error) {
    var mutations []domain.Mutation

    // (1) Tick all in-flight orders on this faction's assets.
    tickMutations := tickMovementOrders(faction, e.Rulebook, e.spatialMap)
    mutations = append(mutations, tickMutations...)

    // (2) Offer the collector a window to issue/revise/cancel orders.
    decisionMutations, err := collectMovementDecisions(faction, factionState, collector, e.Rulebook, e.spatialMap)
    if err != nil {
        return nil, err
    }
    mutations = append(mutations, decisionMutations...)

    return mutations, nil
}
```

### Task 5 — `internal/faction/engine/movement/tick.go` *(new file)*

```go
package movement

func TickMovementOrders(
    faction *domain.Faction,
    rulebook *rulebook.Rulebook,
    spatialMap spatial.SpatialMap,
) []domain.Mutation
```

For each asset in `faction.Assets` with `asset.CurrentOrder != nil`:

- `def := rulebook.Assets[asset.DefinitionID]`
- `newStepIdx := asset.CurrentOrder.StepIdx + def.Speed`
- If `newStepIdx >= len(asset.CurrentOrder.Path) - 1`: emit `MovementOrderCompleted` (with `FinalLocation = asset.CurrentOrder.Destination`) and **skip** emitting `MovementOrderProgressed`. Done.
- Else: emit `MovementOrderProgressed` with `NewStepIdx = newStepIdx`, `NewHexCoords = asset.CurrentOrder.Path[newStepIdx]`. Also include the region by looking up the hex's region via the spatial map (helper needed — see Task 6).

`Cause: "movement_tick"`, `CausedByFactionID: faction.ID`.

### Task 6 — `internal/faction/engine/movement/region_lookup.go` *(new file or helper in tick.go)*

A small helper that, given a `spatial.HexCoord` and the `spatial.SpatialMap`, returns the `regionID` for that hex. Implementation: walk the `HybridMap.regions` map looking for one that contains the hex. (This is not currently exposed on the `SpatialMap` interface — either add `RegionOfHex(coord HexCoord) (string, bool)` to the interface and implement it on `HybridMap`, or capture region from the path-building step where it's known for free.)

**Recommended:** add `RegionOfHex` to the `SpatialMap` interface. The spatial-map types already iterate regions internally — this is a 5-line addition.

### Task 7 — Tick unit tests

`internal/faction/engine/movement/tick_test.go`:

| Test | Setup | Assertion |
|---|---|---|
| No orders | All assets `CurrentOrder == nil` | Empty mutation list |
| Mid-flight tick | Order with `StepIdx=0, Speed=1, Path len=5` | One `Progressed` mutation, `NewStepIdx=1` |
| Completion tick | Order with `StepIdx=3, Speed=2, Path len=5` | One `Completed` mutation; no `Progressed` |
| Multi-hex per tick | Order with `Speed=3, Path len=10, StepIdx=0` | `Progressed` with `NewStepIdx=3` |

<br/>

## Commit 3 — `feat: collector interface for order issue/revise/cancel (1 Coin revision cost)`

### Task 8 — `internal/faction/engine/input_collector.go` (or wherever `InputCollector` is defined)

Add a single method (one round-trip to keep the interface simple — the collector returns all movement decisions for this faction's turn at once):

```go
type MovementDecision struct {
    AssetID     string
    Kind        MovementDecisionKind // Issue | Revise | Cancel
    Destination *Location            // for Issue/Revise; nil for Cancel
}

type MovementDecisionKind int

const (
    MovementDecisionIssue MovementDecisionKind = iota
    MovementDecisionRevise
    MovementDecisionCancel
)

// On the InputCollector interface:
CollectMovementDecisions(faction *domain.Faction, eligibleAssets []*domain.Asset) ([]MovementDecision, error)
```

`eligibleAssets` includes every asset owned by `faction` with `Speed > 0` (and may include those with existing `CurrentOrder` — they're candidates for revise/cancel). The collector returns an empty slice when no decisions are made.

Update every existing `InputCollector` implementation (test harness collectors at minimum) to add a no-op `CollectMovementDecisions` returning `nil, nil`. Real TUI/CLI implementations get stubs flagging "not yet implemented" — UI is out of scope.

### Task 9 — `internal/faction/engine/movement/decisions.go` *(new file)*

```go
package movement

func CollectMovementDecisions(
    faction *domain.Faction,
    factionState *state.FactionState,
    collector InputCollector,
    rulebook *rulebook.Rulebook,
    spatialMap spatial.SpatialMap,
) ([]domain.Mutation, error)
```

- Build `eligibleAssets` (Speed > 0 owned by faction)
- Call `collector.CollectMovementDecisions(faction, eligibleAssets)`
- For each decision, produce mutations:

**Issue:**
```go
def := rulebook.Assets[asset.DefinitionID]
crossingCost := rulebook.DriftCost(def.DriftRating)
path, _, err := spatialMap.Path(asset.Location.WorldID, decision.Destination.WorldID, crossingCost)
if err != nil { return nil, fmt.Errorf("movement: %w", err) }
order := domain.MovementOrder{
    AssetID:     asset.ID,
    Origin:      asset.Location,
    Destination: *decision.Destination,
    Path:        path,
    StepIdx:     0,
    DriftRating: def.DriftRating,
}
mutations = append(mutations, domain.MovementOrderIssued{... Order: order ...})
```

**Revise** (1 Coin cost — Q1):
```go
// Build new path from asset's *current* hex, not origin.
path, _, err := spatialMap.Path(currentHexWorldID, decision.Destination.WorldID, crossingCost)
// (Note: asset is mid-flight, so currentHexWorldID may be empty — Path needs to accept HexCoord origin, not just world ID. See Task 10.)
mutations = append(mutations,
    domain.CoinDelta{FactionID: faction.ID, Delta: -1, Cause: "movement_revision", CausedByFactionID: faction.ID},
    domain.MovementOrderRevised{... NewOrder: newOrder ...},
)
```

**Cancel:**
```go
mutations = append(mutations, domain.MovementOrderCancelled{
    AssetID:    asset.ID,
    StrandedAt: asset.Location, // already mid-flight if applicable
    Cause:      "movement_cancellation",
})
```

### Task 10 — Open question for execution

`spatial.Path` (from Effort 1 Phase 1) takes two world IDs. A mid-flight asset's `Location.WorldID` is empty — there's no world to path *from*. Options:

1. Add a sibling `PathFromHex(originHex HexCoord, originRegion string, toID string, crossingCost int) ([]HexCoord, int, error)` and call that for revision pathing.
2. Construct a synthetic "hex node" inline in `Path` when the origin world ID is empty — pass `(originHexCoord, originRegion)` instead.

**Recommended: Option 1.** Cleaner separation. Add this in Phase 3 Commit 3 alongside the revision path build. Update `spatial.SpatialMap` interface signature accordingly.

Surface this to Robert during execution if the call shape feels off.

### Task 11 — Revision cost test

Harness test in `decisions_test.go`:

- Build a faction with one mid-flight asset (programmatically set `CurrentOrder`)
- Collector returns one `MovementDecisionRevise` with a new destination
- Assert the returned mutations contain exactly one `CoinDelta{Delta: -1}` and one `MovementOrderRevised`

<br/>

## Commit 4 — `feat: LockSkip allows movement, denies action`

The orchestrator wiring from Commit 1 already added `runMovementPhase` to the `LockSkip` branch. This commit verifies the semantics with tests.

### Task 12 — Harness test

`internal/faction/engine/orchestrator_test.go` (or sibling): test a faction with an active `ChangeHomeworld` goal (in `LockSkip` state):

| Scenario | Assertion |
|---|---|
| LockSkip + no orders | Movement Phase runs (no mutations), `CheckpointMovement` fires |
| LockSkip + in-flight order | Order ticks, progress mutation emitted, checkpoint fires |
| LockSkip + collector issues a new order on a different asset | Order is issued; `CoinDelta` not emitted (issue is free; only revisions cost) |
| LockSkip + action attempted | Action selection is skipped (existing behavior) |

The LockSkip branch should not run bookkeeping or stat-raise — confirm test still asserts this.

<br/>

## Commit 5 — `test: movement-phase integration tests`

### Task 13 — Full lifecycle test

`internal/faction/engine/movement/integration_test.go`:

Build a two-faction state with a known spatial map (use the existing test fixture if available; else copy from `internal/spatial`'s test data). Walk through a multi-turn lifecycle:

1. Turn 1: faction A issues an order on asset X from world A to world C (distance 4 hexes).
2. Turn 1: asset X's `Location.WorldID = ""`, `CurrentOrder` is set with `Path` of length 5.
3. Turn 2: tick advances `StepIdx` by `Speed` (e.g. 2). Progress mutation emitted.
4. Turn 3: tick reaches end of path. Completion mutation emitted. Asset's `Location.WorldID = "c"`.

| Assertion | At |
|---|---|
| `MovementOrderIssued` emitted, asset Location cleared | Turn 1 |
| `MovementOrderProgressed` emitted, `StepIdx == 2` | Turn 2 |
| `MovementOrderCompleted` emitted, asset on world C | Turn 3 |
| `Asset.CurrentOrder == nil` | Turn 3 (after Apply) |

### Task 14 — Cancellation test

Similar setup; mid-flight, faction A cancels the order. Assertions:
- `MovementOrderCancelled` emitted
- `asset.CurrentOrder == nil` after Apply
- `asset.Location.WorldID == ""` (stranded at current hex)
- `asset.Location.HexCoords` matches the cancellation-tick's hex

### Task 15 — Revision test

Mid-flight, faction A revises destination from C to D. Assertions:
- Both `CoinDelta{Delta: -1}` and `MovementOrderRevised` emitted in the same Apply
- New `Path` originates from current hex, not original origin
- Future ticks advance along the new path

<br/>
<br/>

# Phase 4 — Transport Ability Step + Handler

Introduce `AbilityStepTransport` as a distinct step type. The handler is invoked during the Movement Phase when a transport asset's controller chooses to activate transport (rather than self-move). Cargo's `Location` mirrors the transport's hex through transit; cargo is individually targetable.

After Phase 4: a new step type is registered in the rulebook decoder and ability engine, but no TOML uses it yet (Effort 3 Phase 5 reclassifies the transport-pattern entries). Tests use synthetic asset definitions to verify behavior.

### Sub-engine shape note

`ability` is mid-alignment per the [sub-engine shapes catalog](../architecture-overview.md#sub-engine-shapes) — eventual target is Shape 2 open-ended with step handlers in an `ability/steps/` sibling package and external registration via a bootstrap function. That refactor is deferred until after the asset movement redesign lands (see [`docs/initiatives/discovery/sub-engine-alignment-discovery.md`](../discovery/sub-engine-alignment-discovery.md)).

**For this phase:** add `transportStepHandler` inline in the `ability/` package alongside the existing `movementStepHandler` and `factionTestStepHandler`, and register it inline in `ability.New()` (matches Task 6's instruction). **Do not** pre-position it in `ability/steps/`. A mixed inline/sibling state is a worse asymmetry than the current uniform-inline state.

The post-asset-movement alignment refactor will then move both remaining step handlers (`factionTestStepHandler` + `transportStepHandler` — `movementStepHandler` is deleted in Effort 3 Phase 5) into `ability/steps/` as a single unit, with bootstrap wiring through `steps.RegisterDefaultSteps(eng)`.

<br/>

## Commit 1 — `feat: add AbilityStepTransport step type and rulebook decoder`

### Task 1 — `internal/faction/domain/asset.go`

Add the new step type constant:

```go
const (
    AbilityStepMovement    AbilityStepType = "movement"
    AbilityStepFactionTest AbilityStepType = "faction_test"
    AbilityStepTransport   AbilityStepType = "transport"
)
```

Reuse the existing `AbilityStep` fields (`MaxHex`, `CoinCost`) — transport has the same shape as movement at the step level. Cargo restrictions live on the handler, not the step record.

### Task 2 — `internal/faction/rulebook/rulebook.go`

Add a case in `convertAbilityStep`:

```go
case "transport":
    return domain.AbilityStep{
        Type:     domain.AbilityStepTransport,
        MaxHex:   r.MaxHex,
        CoinCost: r.CoinCost,
    }, nil
```

Keep the existing `case "movement":` for now — Effort 3 Phase 5 removes it.

### Task 3 — Rulebook decoder test

`internal/faction/rulebook/rulebook_test.go`: add a test loading a synthetic asset TOML with `type = "transport"` and assert the decoded `AbilityStep.Type == domain.AbilityStepTransport`.

<br/>

## Commit 2 — `feat: transport step handler with cargo co-location`

### Task 4 — `internal/faction/engine/ability/transport.go` *(new file)*

```go
package ability

func transportStepHandler(
    faction *domain.Faction,
    asset *domain.Asset,
    step domain.AbilityStep,
    collector Collector,
    _ domain.Roller,
    factionState *state.FactionState,
    rulebook *rulebook.Rulebook,
) ([]domain.Mutation, error)
```

Flow:

1. Build the cargo-eligibility set: same-faction assets with `Speed == 0`, currently co-located with the transport (`asset.Location.WorldID == cargo.Location.WorldID`), within `step.MaxHex` of the transport's hex.
2. Call `collector.SelectTransportCargo(asset, eligibleCargo, step.MaxHex)` — new collector method.
3. For each selected cargo asset, emit a `MovementOrderIssued` mutation that points to the same destination as the transport's order, with the transport's `Path`. The cargo asset's `Speed` is 0, but its `CurrentOrder` ticks alongside the transport's because the engine tick logic should walk **all** assets including those with `Speed == 0` when they have an order.
4. If `step.CoinCost > 0`, emit a `CoinDelta` for the transport cost.

### Task 5 — Cargo tick semantics

Update `TickMovementOrders` (Phase 3 Commit 2 Task 5) to handle the cargo case. Specifically: if an asset has `CurrentOrder != nil` but `Speed == 0` (cargo), tick it at the same rate as its "transport carrier." Cargo's `CurrentOrder.Path` is identical to the transport's, so the cargo's `StepIdx` should equal the transport's `StepIdx` at all times.

**Implementation:** when emitting mutations, the engine can compute the cargo's tick by reusing the transport's `Speed`. To avoid surfacing the transport→cargo relationship at the order level, the simpler approach is to tick every order by `max(asset.Speed, transportSpeedForCargoIfApplicable)`. But this means cargo needs to know its transport — which adds a field to `MovementOrder`.

**Recommended:** add `TransportAssetID string` (optional) to `MovementOrder`. For cargo orders, this points to the carrying transport. The tick function looks up the transport, uses its Speed for the tick. Surface this to Robert at execution time if a cleaner approach emerges — it's a small surface.

### Task 6 — `internal/faction/engine/ability/ability.go`

Register the handler in `New()`:

```go
ae.stepHandlers[domain.AbilityStepTransport] = transportStepHandler
```

### Task 7 — Collector interface

Add `SelectTransportCargo` to the `Collector` interface (in `internal/faction/engine/ability/collector.go` or wherever `Collector` is defined):

```go
SelectTransportCargo(transport *domain.Asset, eligibleCargo []*domain.Asset, maxHex int) ([]*domain.Asset, error)
```

Stub all existing implementations.

<br/>

## Commit 3 — `test: transport handler integration tests`

### Task 8 — Transport happy-path test

Build a faction with:
- One transport asset (synthetic def with `Speed=2`, ability with `AbilityStepTransport, MaxHex=2`)
- One cargo asset (synthetic def with `Speed=0`)
- Both co-located at world A

Test: transport issues movement to world C. During order issuance, the test collector includes a `SelectTransportCargo` call returning the cargo asset.

| Assertion |
|---|
| Transport and cargo both have `CurrentOrder` set with identical paths |
| Cargo's `MovementOrder.TransportAssetID == transport.ID` |
| Cargo's `Location.WorldID == ""` after issue (matches transport) |
| Subsequent ticks advance both assets' `StepIdx` in lockstep |
| Completion: both `Location.WorldID == "c"` |

### Task 9 — Cargo restriction tests

| Test | Assertion |
|---|---|
| Cargo with `Speed > 0` | Excluded from `eligibleCargo` |
| Cargo from a different faction | Excluded from `eligibleCargo` |
| Cargo more than `MaxHex` away | Excluded from `eligibleCargo` |
| Transport can carry self via `Speed`, not transport ability | Verify self isn't listed as cargo candidate |

### Task 10 — Transport cost test

Synthetic transport def with `CoinCost: 2`. Issuing transport order emits `CoinDelta{Delta: -2}` alongside the `MovementOrderIssued` mutations.

# Asset Movement Redesign — Effort 2: Movement Engine

> Overview: [`asset-movement-redesign-plan.md`](asset-movement-redesign-plan.md)
> Depends on: [`asset-movement-redesign-effort-1-plan.md`](asset-movement-redesign-effort-1-plan.md) — all Effort 1 phases must be merged before starting Effort 2.

Two phases. Each is its own execution session.

After Effort 2 lands, the Movement Phase is wired into `RunFactionTurn` and the transport S-flag hook handler exists, registered through the new Effects Engine package. In practice, the phase is dormant — no asset has `Speed > 0` yet (Effort 3 Phase 5) and no TOML carries a `[transport]` block yet (Effort 3 Phase 5). The collector is offered the issue/revise/cancel prompt every Movement Phase but should return "no orders" by default until orders are programmatically driven by tests or (eventually) a UI.

<br/>
<br/>

# Phase 3 — Movement Phase Orchestrator

Insert a new phase into `RunFactionTurn` between Phase 2B (Stat Raise) and Phase 3 (Action). Resolve in-flight orders by ticking them, emit progress/completion mutations, then offer the collector a window to issue, revise, or cancel orders.

After Phase 3: `go test ./...` passes; new harness tests cover order tick → completion, issue, revise (1 Coin cost), cancel, and `LockSkip` movement-only behavior.

<br/>

## Commit 1 — `feat: add CheckpointMovement and PhaseMovement orchestrator hook`

Done. `runMovementPhase` wired into `RunFactionTurn` in `orchestrator.go`; `OnMovementResolved` added to `TurnObserver`; `CheckpointMovement` constant defined. `LockSkip` runs movement before branching to `finishFactionTurn` — movement is inserted before the `LockSkip` check rather than duplicated inside the branch, same semantics, simpler structure.

<br/>

## Commit 2 — `feat: tick in-flight orders and emit progress/completion mutations`

Done. `TickMovementOrders` lives as a `WorldEngine` method in `world/movement.go`. `RegionOfHex` is accessed via type-switch on `*spatial.HybridMap` — no interface change needed since `HexMap` and `GraphMap` have no region concept. All four tick cases covered in `world/world_test.go`.

<br/>

## Commit 3 — `feat: collector interface for order issue/revise/cancel (1 Coin revision cost)`

`MovementDecision`, `MovementDecisionKind`, and `SelectMovementDecisions` on `InputCollector` are done (landed in the input-collector-refactor commit). `BuildMovementMutations` on `WorldEngine` and `eligibleMovableAssets` in `orchestrator.go` are stubbed — this commit fills them in.

### Task 8 — `internal/faction/engine/world/movement.go`

Implement `BuildMovementMutations`. The stub already has the correct signature:

```go
func (engine *WorldEngine) BuildMovementMutations(decisions []MovementDecision, faction *domain.Faction, rulebook *rulebook.Rulebook) ([]domain.Mutation, error)
```

For each decision, look up the asset by `decision.AssetID` in `faction.Assets` and produce mutations:

**Issue:**
```go
def := rulebook.Assets[asset.DefinitionID]
crossingCost := rulebook.DriftCost(def.DriftRating)
path, _, err := engine.spatialMap.Path(asset.Location.WorldID, decision.Destination.WorldID, crossingCost)
if err != nil { return nil, fmt.Errorf("movement: %w", err) }
order := domain.MovementOrder{
    AssetID:     asset.ID,
    Origin:      asset.Location,
    Destination: *decision.Destination,
    Path:        path,
    StepIdx:     0,
    DriftRating: def.DriftRating,
}
mutations = append(mutations, domain.MovementOrderIssued{Order: order, CausedByFactionID: faction.ID})
```

**Revise** (1 Coin cost — path origin: see Task 9):
```go
mutations = append(mutations,
    domain.CoinDelta{FactionID: faction.ID, Delta: -1, Cause: "movement_revision", CausedByFactionID: faction.ID},
    domain.MovementOrderRevised{NewOrder: newOrder, CausedByFactionID: faction.ID},
)
```

**Cancel:**
```go
mutations = append(mutations, domain.MovementOrderCancelled{
    AssetID:           asset.ID,
    StrandedAt:        asset.Location,
    Cause:             "movement_cancellation",
    CausedByFactionID: faction.ID,
})
```

### Task 9 — `internal/faction/engine/orchestrator.go`

Implement `eligibleMovableAssets`: return all assets owned by `faction` where the def's `Speed > 0`.

```go
func eligibleMovableAssets(faction *domain.Faction, rulebook *rulebook.Rulebook) []*domain.Asset {
    var eligible []*domain.Asset
    for _, asset := range faction.Assets {
        def := rulebook.Assets[asset.DefinitionID]
        if def != nil && def.Speed > 0 {
            eligible = append(eligible, asset)
        }
    }
    return eligible
}
```

### Task 10 — Open question for execution: mid-flight revision pathing

`spatial.Path` takes two world IDs. A mid-flight asset's `Location.WorldID` is empty — there's no world to path *from*. Options:

1. Add `PathFromHex(originHex HexCoord, originRegion string, toID string, crossingCost int) ([]HexCoord, int, error)` to `SpatialMap` and implement on `HybridMap`. `HexMap` and `GraphMap` stubs return `ErrNotImplemented`. Call this inside `BuildMovementMutations` for the Revise path.
2. Construct a synthetic hex-origin path inline in `BuildMovementMutations`.

**Recommended: Option 1.** Cleaner separation; `HybridMap` already has the internal graph needed. Update `spatial.SpatialMap` interface alongside this commit.

Surface this during execution if a cleaner approach emerges.

### Task 11 — Revision cost test

`world/world_test.go` (or `world/movement_test.go`):

- Build a faction with one mid-flight asset (programmatically set `CurrentOrder`)
- Call `BuildMovementMutations` directly with one `MovementDecisionRevise` decision
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

`LockSkip` gates only the action phase — bookkeeping, stat-raise, and movement all run normally. The test should confirm action selection is the only skipped phase.

<br/>

## Commit 5 — `test: movement-phase integration tests`

### Task 13 — Full lifecycle test

`internal/faction/engine/movement_test.go`:

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

# Phase 4 — Transport as S-flag Hook Handler

> Replaces the original Phase 4 (`AbilityStepTransport` step + handler), superseded under R-001. See [`docs/initiatives/implementation/ability-engine-redesign-plan.md`](ability-engine-redesign-plan.md) for the reversal context. Decision #6 / #7 in [`asset-movement-redesign-plan.md`](asset-movement-redesign-plan.md) reflect the new shape.

Transport is **not** an ability step. It is an S-flag asset feature: a per-asset `MutationReactor` (Cat 3 hook) that fires reactively on the transport's `MovementOrder{Issued, Progressed, Completed, Cancelled, Revised}` mutations and emits cargo-following `AssetMoved` updates plus the transport's Coin cost. Cargo selection is the one piece that lives outside the hook — `hooks.MutationReactor` has no `Collector`, so the player chooses cargo at order issuance via a new phase-collector method, and the chosen IDs ride along on the transport's `MovementOrder`.

After Phase 4: the Effects Engine (new package mirroring `TagEngine`) is wired into engine boot with one occupant — the transport handler covering Smugglers and Blockade Runners. No TOML uses the new `[assets.X.transport]` block yet (Effort 3 Phase 5 migrates Smugglers + Blockade Runners from `[[ability.steps]]` to `[transport]` and flips their flags from `["A"]` to `["S"]`). Phase 4 tests use synthetic asset defs to verify behavior.

### Shape notes

- **Effects Engine is new.** No per-asset S-flag registration mechanism existed before; tags register via `TagEngine.ApplyAll` but assets had no analog. Phase 4 introduces `internal/faction/engine/effect/` with the same shape: a `Handler` interface keyed by `AssetDefinitionID`, an `EffectsEngine.ApplyAll(state, registry)` that walks each faction's assets, and per-asset handlers under `effect/effects/`. Transport is the first occupant; Boltholes, Tripwire Cells, etc. follow in their own future initiatives.
- **AssetScope MutationReactors are inert under current dispatch.** `dispatch.MutationReactors` calls `MutationReactorsFor(faction.ID, "")` — empty asset instance ID — which short-circuits the AssetScope branch in `Registry.MutationReactorsFor`. Transport handlers therefore register at **`FactionScope`** and filter mutations by asset ID inside `OnMutations`. This matches the `ScavengersReactor` pattern (`tag/tags/scavengers.go`).
- **Movement Phase does not yet dispatch reactors.** `runMovementPhase` in `orchestrator.go` skips the `dispatch.MutationReactors` call that `runActionPhase` makes. Phase 4 wires the dispatch into the movement phase — Commit 2 includes this.
- **Asymmetric transport profiles.** Smugglers: 1 Coin / 2 hex / `[SpecialForces]` cargo. Blockade Runners: 2 Coin / 3 hex / `[MilitaryUnit, SpecialForces]` cargo. The Coin cost in Decision #7 ("1 Coin transport cost") is Smugglers-specific; the design carries asset-specific cost via `TransportProfile.CoinCost`.
- **No re-issuance cost on revise.** Cargo is loaded at issuance, not at each route segment. Revise pays only the existing 1 Coin revision cost; the transport reactor does not emit a new `TransportProfile.CoinCost` `CoinDelta` on `MovementOrderRevised`. Carry-forward of the cargo manifest is mandatory; players who want to change cargo must cancel and re-issue.

<br/>

## Commit 1 — `feat: TransportProfile, cargo manifest on order, SelectTransportCargo collector`

Adds the data shape (TOML + domain) and the issuance-time cargo selection plumbing. After this commit, `BuildMovementMutations` is capable of issuing a transport order with a populated cargo manifest, but no reactor exists yet — no cargo location updates and no Coin cost are emitted. The commit is self-contained: build passes, tests for the decoder + manifest field pass, no behavioral integration test yet.

### Task 1 — `internal/faction/domain/asset.go`

Add the `TransportProfile` struct and field on `AssetDefinition`:

```go
type TransportProfile struct {
    MaxHex     int
    CoinCost   int
    CargoTypes []AssetType
    MaxCargo   int
}

type AssetDefinition struct {
    // ... existing fields ...
    Transport *TransportProfile
}
```

`Transport` is `nil` for non-transport assets. Presence is the marker.

### Task 2 — `internal/faction/domain/location.go`

Add `CargoAssetIDs []string` to `MovementOrder`:

```go
type MovementOrder struct {
    AssetID       string             `toml:"asset_id"`
    Origin        Location           `toml:"origin"`
    Destination   Location           `toml:"destination"`
    Path          []spatial.HexCoord `toml:"path"`
    StepIdx       int                `toml:"step_idx"`
    DriftRating   int                `toml:"drift_rating"`
    CargoAssetIDs []string           `toml:"cargo_asset_ids,omitempty"`
}
```

`CargoAssetIDs` is empty for non-transport orders. Cargo manifest is captured at issuance and preserved through revise (carry-forward).

### Task 3 — `internal/faction/rulebook/rulebook.go`

Add a TOML record for the `[assets.X.transport]` block:

```go
type transportRecord struct {
    MaxHex     int      `toml:"max_hex"`
    CoinCost   int      `toml:"coin_cost"`
    CargoTypes []string `toml:"cargo_types"`
    MaxCargo   int      `toml:"max_cargo"`
}

// attached to the existing assetRecord:
Transport *transportRecord `toml:"transport"`
```

Update `convertAsset` (or equivalent) to populate `def.Transport` from the record when non-nil. Validate that each `CargoTypes` entry parses to a known `domain.AssetType`; return a decode error on mismatch. Default `MaxCargo = 1` when zero (SWN rule shape — one cargo per transport).

### Task 4 — Rulebook decoder test

`internal/faction/rulebook/rulebook_test.go`: add a test loading a synthetic asset TOML with a `[transport]` block. Assert the decoded `def.Transport` matches expected `MaxHex`, `CoinCost`, `CargoTypes`, `MaxCargo`. Add a negative test: an unknown cargo type string returns a decode error.

### Task 5 — `internal/faction/engine/collector.go`

Add `SelectTransportCargo` to the phase-side `InputCollector` interface:

```go
SelectTransportCargo(transport *domain.Asset, eligibleCargo []*domain.Asset, profile *domain.TransportProfile) ([]*domain.Asset, error)
```

Stub the testharness collector (`internal/faction/engine/testharness/collector.go`) with a configurable response (a slice-returning func field, matching the existing harness style). Update the mock collector under `internal/faction/engine/action/actions/mocks/` if it implements `InputCollector` for shared paths — verify by build.

### Task 6 — `internal/faction/engine/world/movement.go`

Extend `BuildMovementMutations` so that when issuing for an asset whose def has `Transport != nil`, it:

1. Builds the cargo-eligibility set: same-faction assets where the cargo's def `Type` ∈ `Transport.CargoTypes`, currently co-located with the transport (`cargo.Location.WorldID == transport.Location.WorldID && cargo.Location.HexCoords == transport.Location.HexCoords`), `cargo.CurrentOrder == nil` (no in-flight cargo), and not the transport itself.
2. Calls `collector.SelectTransportCargo(transport, eligible, def.Transport)`.
3. Validates the returned slice: ≤ `Transport.MaxCargo`, all in the eligibility set, destination within `Transport.MaxHex` of the transport's origin (path length check against `Transport.MaxHex`). On violation, return an error — the collector returned an invalid manifest.
4. Stores the cargo asset IDs on the issued `MovementOrder.CargoAssetIDs`.

For non-transport issuances (`def.Transport == nil`), behavior is unchanged. Revise and Cancel decisions do not invoke cargo selection — Revise carries forward `CargoAssetIDs` from the existing order; Cancel is reactor-handled in Commit 2.

### Task 7 — Manifest field test

`internal/faction/engine/world/world_test.go` (or sibling): given a synthetic transport def with `MaxCargo=1`, `CargoTypes=[SpecialForces]`, a `SelectTransportCargo`-stub returning one eligible cargo, assert the produced `MovementOrderIssued.Order.CargoAssetIDs` contains that one asset's ID and nothing else. Negative test: collector returns two cargo when `MaxCargo=1` → `BuildMovementMutations` returns an error.

<br/>

## Commit 2 — `feat: EffectsEngine + transport reactor; dispatch reactors in Movement Phase`

Introduces the Effects Engine, the transport reactor (handling all five movement mutations), and wires `dispatch.MutationReactors` into `runMovementPhase`. After this commit, transport behavior is fully reactive: Coin cost, cargo co-location, stranding on cancel, and following on revise all happen via the hook subsystem.

### Task 1 — New package `internal/faction/engine/effect/`

Mirror `internal/faction/engine/tag/`:

```go
// effect/effect.go
package effect

type Handler interface {
    AssetDefinitionID() string
    Apply(faction *domain.Faction, asset *domain.Asset, hookRegistry *hooks.Registry)
}

type EffectsEngine struct {
    handlers map[string]Handler
}

func New() *EffectsEngine { /* registers defaults — see Task 3 */ }

func (e *EffectsEngine) Register(handler Handler) {
    e.handlers[handler.AssetDefinitionID()] = handler
}

func (e *EffectsEngine) ApplyAll(factionState *state.FactionState, rulebook *rulebook.Rulebook, hookRegistry *hooks.Registry) {
    for _, faction := range factionState.Factions {
        for _, asset := range faction.Assets {
            def := rulebook.Assets[asset.DefinitionID]
            if def == nil { continue }
            handler, ok := e.handlers[def.ID]
            if !ok { continue }
            handler.Apply(faction, asset, hookRegistry)
        }
    }
}
```

`ApplyAll` takes the `rulebook` (which `TagEngine.ApplyAll` does not) because handlers may need the def to register profile-aware reactors. Asset-without-def cases are silent skips (consistent with the tag-engine "data-only" treatment).

### Task 2 — Generic transport handler `effect/effects/transport.go` *(new file)*

Single handler keyed by `AssetDefinitionID == ""` won't work — handlers map by definition ID, but transport applies to *any* asset with `def.Transport != nil`. Choices:

- **Option A (Recommended):** drop the per-definition-ID keying for the transport handler. Register the transport handler under a sentinel like `""` and have `ApplyAll` also walk a second pass calling a generalized "for-each-asset matcher". *Rejected — diverges from the TagEngine pattern.*
- **Option B (Recommended):** register one `TransportHandler` per transport-capable definition ID. When the rulebook is loaded, the rulebook decoder (Commit 1, Task 3) populates `def.Transport`. At engine `New()`, walk `rulebook.Assets` for any def with `Transport != nil` and call `EffectsEngine.Register(NewTransportHandler(def.ID))`. The handler's `AssetDefinitionID()` returns its bound def ID; `Apply` registers a `TransportReactor` for that (faction, asset, profile) tuple. **This is the chosen approach — symmetric with TagEngine, data-driven by `def.Transport`.**

Engine struct change: `engine.New(...)` now needs the rulebook before constructing `Effect`. Registration call:

```go
effects := effect.New()
for _, def := range rulebook.Assets {
    if def.Transport != nil {
        effects.Register(effects.NewTransportHandler(def))
    }
}
```

Reactor:

```go
type TransportReactor struct {
    FactionID  string
    AssetID    string
    Profile    *domain.TransportProfile
}

func (reactor *TransportReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
    var extra []domain.Mutation
    for _, mutation := range mutations {
        switch m := mutation.(type) {
        case domain.MovementOrderIssued:
            if m.AssetID != reactor.AssetID || len(m.Order.CargoAssetIDs) == 0 { continue }
            // Charge transport Coin cost.
            extra = append(extra, domain.CoinDelta{
                FactionID: reactor.FactionID,
                Delta:     -reactor.Profile.CoinCost,
                Cause:     "transport_cost",
                CausedByFactionID: reactor.FactionID,
            })
            // Place each cargo at Path[0] with WorldID cleared, matching the transport's post-issue location.
            for _, cargoID := range m.Order.CargoAssetIDs {
                extra = append(extra, buildCargoMoved(factionState, reactor.FactionID, cargoID, domain.Location{
                    HexCoords: m.Order.Path[0],
                    Region:    m.Order.Origin.Region,
                }, "transport_issued"))
            }
        case domain.MovementOrderProgressed:
            // Look up the transport's order (from factionState) to find current cargo manifest.
            // Emit AssetMoved for each cargo to the new hex.
        case domain.MovementOrderCompleted:
            // Emit AssetMoved for each cargo to FinalLocation.
        case domain.MovementOrderCancelled:
            // Emit AssetMoved for each cargo to the transport's current Location (stranded; WorldID="").
        case domain.MovementOrderRevised:
            // No-op for the reactor — the new order's CargoAssetIDs were copied forward by BuildMovementMutations (Task 5 below);
            // subsequent Progressed events drive cargo updates along the new path.
        }
    }
    return extra
}
```

`buildCargoMoved` is a package-local helper that resolves the cargo asset, computes the `FromLocation` from current `factionState`, and constructs an `AssetMoved` mutation. Cargo's `OwnerID` is the transport's faction (cargo is always same-faction per the eligibility filter).

### Task 3 — Wire EffectsEngine into the Engine struct

`internal/faction/engine/core.go` (or wherever the `Engine` struct is defined): add `Effect *effect.EffectsEngine` field, analogous to `Tag *tag.TagEngine`. Wire `effects.ApplyAll(state, rulebook, hooks)` into engine boot — same call site as `Tag.ApplyAll`. Order is irrelevant for transport (no tag interacts with transport mutations).

If `engine.New` doesn't currently take the rulebook before engine struct construction, restructure so that `effect.New(rulebook)` (or the populated `Register` calls) runs once the rulebook is loaded.

### Task 4 — Dispatch reactors in Movement Phase

`internal/faction/engine/orchestrator.go`, in `runMovementPhase`: insert a `dispatch.MutationReactors` call between the `decisionMutations` append and `applyAndRecord`:

```go
mutations = append(mutations, decisionMutations...)

mutations, err = dispatch.MutationReactors(e.Hooks, faction, mutations, factionState, e.Rulebook)
if err != nil {
    observer.OnError(faction, err)
    return err
}

if len(mutations) > 0 {
    if err := e.applyAndRecord(...); err != nil { ... }
}
```

The dispatcher's existing recursion-bound (depth 5) handles any cascade from transport-emitted mutations. Transport-emitted mutations are `CoinDelta` and `AssetMoved`, neither of which triggers further transport reactor activity (the reactor only listens to `MovementOrder*` mutations) — depth 2 in practice.

### Task 5 — Revise carries cargo manifest forward

`internal/faction/engine/world/movement.go`, in `BuildMovementMutations`'s `MovementDecisionRevise` branch: copy `CargoAssetIDs` from `asset.CurrentOrder.CargoAssetIDs` into the new `MovementOrder.CargoAssetIDs`. Revise does not re-invoke `SelectTransportCargo`. Cargo cannot be added or dropped via revise; player must cancel and re-issue to change manifest.

If `asset.CurrentOrder` is `nil` at revise time, that's an invariant violation (revise requires an in-flight order) — error out, same as existing revise handling.

### Task 6 — Mutation Apply for `CargoAssetIDs`

`internal/faction/engine/mutation/mutation.go`'s `MovementOrderIssued`, `MovementOrderRevised`, `MovementOrderCancelled`, and `MovementOrderCompleted` handlers already deep-copy the order — no Apply change needed. Sanity-check the `MovementOrderCompleted` apply zeroes `CurrentOrder` (it does, line 226) so the cargo manifest is cleaned up implicitly on transport completion.

<br/>

## Commit 3 — `test: transport lifecycle, restrictions, cancel/revise`

Integration tests covering the transport behavior end-to-end. Uses synthetic asset defs (`def.Transport` populated programmatically) since no rulebook TOML carries `[transport]` blocks yet — that's Effort 3 Phase 5.

### Task 1 — Reactor unit tests

`internal/faction/engine/effect/effects/transport_test.go`:

| Test | Assertion |
|---|---|
| Issued with non-empty cargo manifest | `CoinDelta{Delta: -Profile.CoinCost}` emitted; one `AssetMoved` per cargo, `ToLocation.WorldID == ""`, `HexCoords == Path[0]` |
| Issued with empty cargo manifest | No reactor output |
| Issued for unrelated asset | No reactor output (asset-ID filter works) |
| Progressed for the transport | One `AssetMoved` per cargo, `ToLocation.HexCoords == NewHexCoords`, `WorldID == ""` |
| Completed for the transport | One `AssetMoved` per cargo, `ToLocation == FinalLocation` |
| Cancelled for the transport | One `AssetMoved` per cargo, `ToLocation == transport's current Location` (mid-flight, `WorldID == ""`) |
| Revised for the transport | No reactor output on the Revised mutation itself |

### Task 2 — Cargo eligibility tests

`internal/faction/engine/world/movement_test.go` (or sibling): tests on `BuildMovementMutations`'s cargo-eligibility filter, using a synthetic transport def.

| Test | Assertion |
|---|---|
| Cargo of allowed type, co-located | Included in eligible set |
| Cargo of disallowed type (not in `CargoTypes`) | Excluded |
| Cargo from a different faction | Excluded |
| Cargo on a different world | Excluded |
| Cargo on the same world but different hex | Excluded |
| Cargo with `CurrentOrder != nil` (already in flight) | Excluded |
| Transport asset itself | Excluded from its own eligibility set |
| Collector returns cargo beyond `MaxCargo` | `BuildMovementMutations` returns an error |
| Destination beyond `Transport.MaxHex` from origin | `BuildMovementMutations` returns an error |

### Task 3 — Full lifecycle integration test

`internal/faction/engine/movement_test.go` (extend the existing file or add `transport_test.go` sibling):

Synthetic faction with one transport (`Speed=2`, `Transport.MaxHex=3`, `Transport.CoinCost=1`, `Transport.CargoTypes=[SpecialForces]`, `Transport.MaxCargo=1`) and one cargo (`Speed=0`, type `SpecialForces`). Both at world A.

Walk multi-turn lifecycle:

| At | Assertion |
|---|---|
| Turn 1 issuance | `MovementOrderIssued.Order.CargoAssetIDs == [cargo.ID]`; `CoinDelta{Delta: -1, Cause: "transport_cost"}` emitted reactively; cargo's `Location.WorldID == ""`, `HexCoords == Path[0]` |
| Turn 2 tick | `MovementOrderProgressed` emitted for transport; cargo's `Location.HexCoords` matches transport's new hex |
| Turn 3 tick | `MovementOrderCompleted` emitted; both transport and cargo's `Location == FinalLocation`; `transport.CurrentOrder == nil` |

### Task 4 — Cancel mid-flight test

Issue transport+cargo, advance one tick, then cancel.

| Assertion |
|---|
| `MovementOrderCancelled` emitted for transport |
| `AssetMoved` emitted for cargo with `ToLocation` matching transport's current (post-progressed) `Location` |
| After apply: `transport.CurrentOrder == nil`, `cargo.Location` matches the stranded hex, `cargo.Location.WorldID == ""` |

### Task 5 — Revise mid-flight test

Issue transport+cargo (dest C), advance one tick, then revise to a new destination (dest D, still within `Transport.MaxHex` from current hex).

| Assertion |
|---|
| `MovementOrderRevised.NewOrder.CargoAssetIDs == [cargo.ID]` (carried forward) |
| `CoinDelta{Delta: -1, Cause: "movement_revision"}` emitted (existing revision cost; no new `transport_cost` `CoinDelta`) |
| Subsequent Progressed mutations move cargo along the new path |
| At completion: both transport and cargo land at dest D |

### Task 6 — Boot-time registration test

`internal/faction/engine/effect/effect_test.go`: build a `*rulebook.Rulebook` containing one def with `Transport != nil` and one without. Call `effect.New(rulebook)` + `ApplyAll`. Assert the `hooks.Registry` has one `MutationReactor` registered at `FactionScope(faction.ID)` from source `"transport:<defID>"` for the transport-owning faction; zero for a faction with no transport-capable assets.

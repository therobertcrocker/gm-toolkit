# Spatial Interface Redesign

## Context

The `spatial.SpatialMap` interface promises methods that aren't honest for other map kinds. A pure graph map (Codex's anticipated use case) cannot honestly implement `Path(...) []HexCoord` or `Distance(..., crossingCost int)` — `HexCoord` and `crossingCost` are `RegionMap`-specific concepts. The faction engine, the only current consumer, reaches past the interface and type-asserts back to `*spatial.RegionMap` to call methods (`PathFromHex`, `RegionOfHex`) that aren't on the contract. The interface is both too broad (claims things `RegionMap`-specific) and too narrow (missing what `world` actually needs).

Audit findings during planning:

- `WorldEngine.Distance` has zero production callers (LSP `findReferences`).
- `WorldEngine.PathFromHex` is a thin wrapper used only by `world/movement.go`.
- `Asset.Location` already carries `HexCoords` (and `Region`). The WorldID-based `Path` variant is redundant — every caller already has hex coords at the call site, so one hex-based `Path` covers both `MovementDecisionIssue` and `MovementDecisionRevise`.
- The stub map types (`HexMap`, `GraphMap`) satisfy `SpatialMap` only by returning `ErrNotImplemented` from every method — architectural placeholders with no backing implementation.

## Approach

Three principles drive the redesign: extensibility, readability, YAGNI.

1. **Narrow `spatial.SpatialMap`** to the only honest universal operation:
   ```go
   type SpatialMap interface {
       Location(id string) (Location, bool)
   }
   ```
   All routing methods come off the interface. They stay as concrete methods on `*RegionMap`, which is the library's first complete spatial implementation. Future map kinds (graph-only, voxel) implement the narrow `SpatialMap` and add their own routing API to match their shape.

2. **Collapse routing to one axis using `RegionHex`.** `HexCoord` alone is not globally unique — the same `(Q, R)` can appear in two different regions, because regions have local coordinate systems. Routing must identify a hex by both its coordinate and its region. `RegionHex` (defined in `region_map.go`, alongside the type that uses it) bundles these:
   ```go
   type RegionHex struct {
       RegionID string
       Coord    HexCoord
   }
   ```
   `RegionMap.Path` and `RegionMap.Distance` take `(from, to RegionHex, crossingCost int)`. Callers always have the region at the call site via `Asset.Location.Region`. `Path` returns `[]RegionHex` so consumers can read the region of each step directly — eliminating any need for a reverse-lookup. The world-ID-based variants are deleted, not replaced.

3. **Define `world.HexRouter`** in `internal/faction/engine/world/`. It embeds `spatial.SpatialMap` and adds the routing methods the world engine consumes:
   ```go
   type HexRouter interface {
       spatial.SpatialMap
       Distance(from, to spatial.RegionHex, crossingCost int) (int, error)
       Path(from, to spatial.RegionHex, crossingCost int) ([]spatial.RegionHex, int, error)
   }
   ```
   `RegionOfHex` is removed from the interface entirely — `Path` returning `[]RegionHex` means consumers read the region directly from each step. `WorldEngine` holds a `HexRouter`. The compile-time satisfaction assert lives at the consumer (`world` knows what shape of map it accepts); the spatial-side assert is reduced to the narrower `SpatialMap` contract.

Drive-by cleanups that fall out of touching these files: silent skips in `RebuildIndex` become a returned error, dead wrappers on `WorldEngine` are deleted, untyped `Cause` strings in `movement.go` become typed constants, scaffolding comments in `RegionMap.neighbors()` are stripped, `WorldEngine.Location` is fixed to return `(spatial.HexLocation, bool)`.

## Out of Scope

- `Location` interface fields (`TechLevel`, `Population`) stay as-is. SWN-specificity is a future concern; revisit when Codex actually wants a different shape.
- `Location` storing three derived axes (`WorldID`, `HexCoords`, `Region`) — captured as backlog R-005.
- Reverse-boundary index in `RegionMap.neighbors` — already in spatial-engine backlog (Deferred Minor #1).
- Broader `Cause` typing across all mutations — this initiative types only the movement causes.

## Phasing

Two `wip:` commits on `feature/movement-redesign`. Squashed to one `refactor:` commit on merge.

<br/>
<br/>

# Commit 1 — `wip: refactor: reshape spatial library; adopt HexRouter in world engine`

Type rename, interface shape, library API consolidation, and consumer adoption land together so the build stays green at the commit boundary.

## Spatial package

### Rename `HybridMap` → `RegionMap`

Pure rename — no behavior change. Driven first so the rest of the commit's tasks reference the new name consistently.

- Rename the type `HybridMap` → `RegionMap` throughout `internal/spatial/`.
- Rename files: `hybrid_map.go` → `region_map.go`; `hybrid_map_test.go` → `region_map_test.go`.
- Rename the constructor `LoadHybrid` → `LoadRegionMap`.
- Rename receivers and parameter names: `hybridMap` → `regionMap`.
- Update the compile-time assert: `var _ SpatialMap = (*RegionMap)(nil)`.

"`HybridMap`" described the implementation history (hex grid + graph edges combined); "`RegionMap`" describes what the type *is* — a spatial map organized into named regions, each internally a hex grid with boundary connections to neighbors.

### `internal/spatial/spatial.go`

- Narrow `SpatialMap` to a single method:
  ```go
  type SpatialMap interface {
      Location(id string) (Location, bool)
  }
  ```
- Remove the compile-time asserts for `(*HexMap)` and `(*GraphMap)` (those types are being deleted).
- Keep `var _ SpatialMap = (*RegionMap)(nil)` and `var _ HexLocation = (*World)(nil)`.
- Leave the sentinel error block unchanged.

### `internal/spatial/region_map.go`

- Define `RegionHex` here (alongside the type that uses it, not in `spatial.go`):
  ```go
  type RegionHex struct {
      RegionID string
      Coord    HexCoord
  }
  ```
- Collapse `Path` and `PathFromHex` into one `RegionHex`-based method:
  ```go
  func (regionMap *RegionMap) Path(from, to RegionHex, crossingCost int) ([]RegionHex, int, error)
  ```
  Extract an unexported `pathBetween(source, target hexNode, crossingCost int) ([]RegionHex, int, error)` helper that owns the dijkstra call, path reconstruction, and reversal — eliminates the duplication that existed between the two old methods. `hexNode` already carries `regionID`, so the return conversion is direct.
- Switch `Distance` to `RegionHex` inputs:
  ```go
  func (regionMap *RegionMap) Distance(from, to RegionHex, crossingCost int) (int, error)
  ```
- `RegionOfHex` stays as a method on `*RegionMap` but is **not** on any interface — it is no longer needed by any caller once `Path` returns `[]RegionHex`.
- Strip the `(1)…(2)…(3)…(4)` scaffolding comment blocks in `neighbors()`. Keep the implementation; remove the duplicated pseudocode lines.

### `internal/spatial/hex_map.go` and `internal/spatial/graph_map.go`

Delete both files. They were architectural placeholders satisfying the interface only via `ErrNotImplemented`. When a real graph or hex-only map shows up, it will arrive with a real implementation.

### `internal/spatial/region_map_test.go`

Update `Distance` and `Path` test cases (and consolidate the `PathFromHex` tests into the unified `Path` tests) to the new `RegionHex` signatures. Test cases construct `RegionHex{RegionID: "...", Coord: HexCoord{...}}` directly — no world lookup needed. The test scenarios stay equivalent; only the call sites change.

## World package

### `internal/faction/engine/world/world.go`

- Define the consumer contract at the top of the file:
  ```go
  type HexRouter interface {
      spatial.SpatialMap
      Distance(from, to spatial.RegionHex, crossingCost int) (int, error)
      Path(from, to spatial.RegionHex, crossingCost int) ([]spatial.RegionHex, int, error)
  }

  var _ HexRouter = (*spatial.RegionMap)(nil)
  ```
  `RegionOfHex` is gone — callers read the region from path steps directly.
- Change `WorldEngine.spatialMap` field type to `HexRouter`.
- Change `NewWithMap(spatialMap HexRouter) *WorldEngine`.
- Fix `WorldEngine.Location` return type:
  ```go
  func (engine *WorldEngine) Location(id string) (spatial.HexLocation, bool) {
      loc, ok := engine.spatialMap.Location(id)
      if !ok {
          return nil, false
      }
      hexLoc, ok := loc.(spatial.HexLocation)
      return hexLoc, ok
  }
  ```
- Delete `WorldEngine.Distance` (zero production callers, confirmed).
- Delete `WorldEngine.PathFromHex` (its only caller is `movement.go`, which now calls `engine.spatialMap.Path(...)` directly).

### `internal/faction/domain/` (new scope, falls here)

`domain.MovementOrder.Path` changes from `[]spatial.HexCoord` to `[]spatial.RegionHex`. Path steps now carry their region so `TickMovementOrders` can label movements without a reverse lookup.

### `internal/faction/engine/world/movement.go`

- In `TickMovementOrders`, drop the `*spatial.RegionMap` type assertion entirely. Read region from the path step:
  ```go
  step := asset.CurrentOrder.Path[newStepIdx]
  region := step.RegionID
  newHexCoords := step.Coord
  ```
- In `BuildMovementMutations`:
  - `MovementDecisionIssue`: look up the destination world to build `RegionHex` inputs:
    ```go
    destLoc, _ := engine.spatialMap.Location(decision.Destination.WorldID)
    destHex := destLoc.(spatial.HexLocation)
    q, r := destHex.Coords()
    to := spatial.RegionHex{RegionID: destHex.RegionID(), Coord: spatial.HexCoord{Q: q, R: r}}
    from := spatial.RegionHex{RegionID: asset.Location.Region, Coord: asset.Location.HexCoords}
    path, _, err := engine.spatialMap.Path(from, to, crossingCost)
    ```
  - `MovementDecisionRevise`: build `RegionHex` directly from `asset.Location` and `decision.Destination`:
    ```go
    from := spatial.RegionHex{RegionID: asset.Location.Region, Coord: asset.Location.HexCoords}
    to := spatial.RegionHex{RegionID: decision.Destination.Region, Coord: decision.Destination.HexCoords}
    path, _, err := engine.spatialMap.Path(from, to, crossingCost)
    ```

### `internal/faction/engine/world/world_test.go`

Update `stubMap` to satisfy `HexRouter` — switches `Path` and `Distance` to `RegionHex` signatures; no `RegionOfHex` stub needed.

### `internal/faction/engine/testharness/harness.go`

Per decision #207, `testharness.NewHarness` wires `stubSpatialMap` into `eng.World` via `world.NewWithMap`. Update its `Path` and `Distance` signatures to `RegionHex`; remove `RegionOfHex` stub. Update compile-time assert to `world.HexRouter`.

## Validation

- `go build ./...` green.
- `go test ./...` green.
- No type assertions to `*spatial.RegionMap` remain in the faction engine.
- No bare `HexCoord` appears as a routing input — all routing goes through `RegionHex`.

<br/>
<br/>

# Commit 2 — `wip: refactor: surface index-build errors and type movement causes`

Two independent cleanups that touch the same files this initiative already modified. Bundled per "group as much as possible" preference.

## RebuildIndex error return

### `internal/faction/engine/world/world.go`

Change `RebuildIndex` to:

```go
func (engine *WorldEngine) RebuildIndex(factionState *state.FactionState) ([]string, error)
```

- Returned slice carries skipped location IDs (assets or bases whose `Location.WorldID` doesn't resolve via `engine.spatialMap.Location(...)`).
- Error reserved for hard failures (none today, but the signature shape is forward-compatible).
- Remove both `// TODO: surface skipped locations once a logging layer exists` comments.

### `internal/faction/engine/orchestrator.go`

The caller of `RebuildIndex` forwards the returned skipped IDs to the observer. Per `feedback_error_return_over_observer`, the orchestrator owns observer wiring — `world` only returns the data.

Locate the `RebuildIndex` call site in the orchestrator (modified in current branch state) and:

1. Capture the returned `([]string, error)`.
2. If error: propagate.
3. If skipped non-empty: forward via the appropriate observer hook. Identify the right hook by inspecting the existing observer interface — if no fitting hook exists, add one (e.g., `OnIndexSkipped(skipped []string)`).

## Typed movement Cause constants

### `internal/faction/domain/mutation.go`

Add a new typed constants block near the top of the file (the `Cause` field is already defined here on multiple mutation types):

```go
const (
    CauseMovementTick         = "movement_tick"
    CauseMovementRevision     = "movement_revision"
    CauseMovementCancellation = "movement_cancellation"
)
```

Scope is limited to movement causes — broader `Cause` typing across all mutations is out of scope.

### `internal/faction/engine/world/movement.go`

Replace all four string-literal `Cause` values:
- `"movement_tick"` (two sites) → `domain.CauseMovementTick`
- `"movement_revision"` (two sites, including the `CoinDelta` mutation) → `domain.CauseMovementRevision`
- `"movement_cancellation"` (one site) → `domain.CauseMovementCancellation`

## Validation

- `go build ./...` green.
- `go test ./...` green.

<br/>
<br/>

# Pre-merge checklist (separate session)

Per `feedback_branch_initiative_layering`: pre-merge work is branch-level. Defer until movement redesign is otherwise ready to merge (`project_movement_branch_merge`).

- **Code review** as a senior reviewing a junior's PR — focus on the interface boundary, the test updates, and whether any consumer outside `faction/` is reaching into `spatial` in ways the narrow interface no longer supports.
- **Decisions log entries**:
  - *Faction-manager log*: new section for this branch's work — entries for `world.HexRouter` (consumer-side interface), routing API collapse to hex coords, deletion of dead `WorldEngine.Distance`/`PathFromHex` wrappers, R-005 backlog addition.
  - *Spatial-engine log*: new section — entries for `SpatialMap` narrowing to `Location(id)`, stub map deletion, `Path`/`Distance` consolidation to hex-coord signatures, scaffolding-comment cleanup.
- **`planned-work.md`**: R-005 already added during planning; nothing else to update.
- **Squash** both `wip:` commits into one `refactor: ...` conventional commit on merge.

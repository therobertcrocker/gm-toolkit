# Asset Movement Redesign — Effort 1: Foundation

> Overview: [`asset-movement-redesign-plan.md`](asset-movement-redesign-plan.md)
> Discovery: [`docs/initiatives/discovery/asset-movement-redesign-discovery.md`](../discovery/asset-movement-redesign-discovery.md)

Two phases. Each is its own execution session.

After Effort 1 lands, the system behaves identically to today — only the shapes have shifted. No assets can issue movement orders yet (no `Speed > 0` values; no Movement Phase in the orchestrator). The point of this effort is to land all type-shape changes without introducing new behavior, so the engine work in Effort 2 has a clean foundation.

<br/>
<br/>

# Phase 1 — Additive Domain Types & `spatial.Path`

Zero blast radius. All new types are additive — `Asset.Location` remains a string; `Faction.Homeworld` remains a string; `Base.Location` remains a string. The new `Location` struct exists alongside them, used only by the new (also-additive) `MovementOrder` struct, the new mutations, and `spatial.Path`'s return signature.

After Phase 1: `go build ./...` passes; no existing tests fail; no existing behavior changes; `Asset.CurrentOrder` is always nil; no asset has `Speed > 0`.

<br/>

## Commit 1 — `feat: add spatial.Path pathfinding primitive`

### Task 1 — `internal/spatial/hybrid_map.go`

Add `Path` alongside the existing `Distance`:

```go
func (hybridMap *HybridMap) Path(fromID, toID string, crossingCost int) ([]HexCoord, int, error)
```

Returns the sequence of hex coords from origin to destination inclusive (so `Path[0]` is the origin's hex and `Path[len-1]` is the destination's hex), plus the total cost (same value `Distance` would return), plus an error matching `Distance`'s error contract (`ErrInvalidCost`, `ErrUnknownWorld`, `ErrNoPath`).

Implementation: reuse the Dijkstra body in `Distance`. Track a `predecessor map[hexNode]hexNode` alongside the `dist` map; on each relaxation, record `predecessor[next.to] = current.node`. On target hit, walk predecessors back to the source, reverse, and return the `[]HexCoord` (drop the region IDs — callers only need coords). Total cost is the existing `current.cost` at target.

The `Distance` function should be refactored to call `Path` internally and discard the path, **or** the two functions share a common Dijkstra helper. Pick whichever leaves the diff cleaner at execution time.

### Task 2 — `internal/spatial/spatial.go`

Extend the `SpatialMap` interface with `Path`:

```go
type SpatialMap interface {
    Location(id string) (Location, bool)
    Distance(fromID, toID string, crossingCost int) (int, error)
    Path(fromID, toID string, crossingCost int) ([]HexCoord, int, error)
}
```

The other map implementations (`HexMap`, `GraphMap`) are mock/test-only — they get `Path` stubs that return `nil, 0, ErrNotImplemented` (matching the existing pattern for `Distance` on `GraphMap` if applicable; verify at execution time).

### Task 3 — `internal/spatial/hybrid_map_test.go`

Add table-driven tests for `Path`:

| Case | From | To | Expected |
|---|---|---|---|
| Same world | `"a"` | `"a"` | `[]HexCoord{a.Hex}`, dist 0 |
| One step intra-region | adjacent hexes | adjacent hexes | 2-element path, dist 1 |
| Cross-region | A → B (via boundary) | A → B | path includes A's boundary hex then B's hex, dist matches `Distance` |
| Unknown world | `"unknown"` | `"a"` | `ErrUnknownWorld` |
| Negative cost | any | any | `ErrInvalidCost` |

Reuse the existing `HybridMap` test fixtures.

<br/>

## Commit 2 — `feat: add Location, MovementOrder, Speed, PhaseMovement domain types`

### Task 4 — `internal/faction/domain/location.go` *(new file)*

Define the `Location` struct:

```go
package domain

import "github.com/therobertcrocker/gm-toolkit/internal/spatial"

type Location struct {
    WorldID   string          `toml:"world_id"`
    HexCoords spatial.HexCoord `toml:"hex_coords"`
    Region    string          `toml:"region"`
}
```

If TOML doesn't decode `spatial.HexCoord` directly (it's a struct with `Q, R int` fields), inline the hex fields:

```go
type Location struct {
    WorldID string `toml:"world_id"`
    HexQ    int    `toml:"hex_q"`
    HexR    int    `toml:"hex_r"`
    Region  string `toml:"region"`
}

func (location Location) Coords() spatial.HexCoord {
    return spatial.HexCoord{Q: location.HexQ, R: location.HexR}
}
```

Pick whichever shape TOML cleanly serializes. Verify at execution time by writing a roundtrip test (`TestLocation_TomlRoundtrip`).

### Task 5 — `internal/faction/domain/movement_order.go` *(new file)*

```go
package domain

import "github.com/therobertcrocker/gm-toolkit/internal/spatial"

type MovementOrder struct {
    AssetID     string             `toml:"asset_id"`
    Origin      Location           `toml:"origin"`
    Destination Location           `toml:"destination"`
    Path        []spatial.HexCoord `toml:"path"`
    StepIdx     int                `toml:"step_idx"`
    DriftRating int                `toml:"drift_rating"`
}
```

`StepIdx` advances by `Speed` per tick. Asset's current hex during transit = `Path[StepIdx]`. Order completes when `StepIdx >= len(Path) - 1`.

### Task 6 — `internal/faction/domain/asset.go`

- Add `Speed int` to `AssetDefinition` with TOML tag `speed`. No default; zero means non-movable.
- Add `CurrentOrder *MovementOrder` to `Asset` with TOML tag `current_order` (and `omitempty` if TOML supports it; otherwise leave as a nil pointer — verify TOML encoding behavior).

Do **not** touch `Asset.Location` in this phase — it stays a `string`. Phase 2 migrates it.

### Task 7 — `internal/faction/engine/turn/turn.go` (or wherever `TurnPhase` is defined)

Locate the `TurnPhase` enum (grep for `PhaseBookkeeping` if the file path isn't obvious — it may live in `engine/orchestrator.go` or a sibling). Add `PhaseMovement` between `PhaseBookkeeping` and `PhaseAction`:

```go
const (
    PhaseBookkeeping TurnPhase = iota
    PhaseMovement
    PhaseAction
    PhaseComplete
)
```

If `PhaseStatRaise` or similar exists between Bookkeeping and Action, place `PhaseMovement` after it (matching the discovery's flow: 2B → Movement → 3).

### Task 8 — `internal/faction/engine/orchestrator.go`

Add the checkpoint constant in the same block as the existing ones:

```go
const (
    CheckpointBookkeeping  = "bookkeeping"
    CheckpointMovement     = "movement"
    CheckpointActionResult = "action_result"
    CheckpointGoalLocked   = "goal_locked"
    CheckpointCycleSummary = "cycle_summary"
)
```

Do not wire it into `RunFactionTurn` yet — that's Phase 3.

<br/>

## Commit 3 — `feat: add movement-order mutations`

### Task 9 — `internal/faction/domain/mutation.go`

Append five new mutation types. Match the existing pattern (struct with `FactionID`, `Cause`, `CausedByFactionID`, plus type-specific fields; implements `Type() string`):

```go
type MovementOrderIssued struct {
    FactionID         string
    AssetID           string
    Order             MovementOrder
    Cause             string
    CausedByFactionID string
}
func (m MovementOrderIssued) Type() string { return "movement_order_issued" }

type MovementOrderProgressed struct {
    FactionID         string
    AssetID           string
    NewStepIdx        int
    NewHexCoords      spatial.HexCoord
    Cause             string
    CausedByFactionID string
}
func (m MovementOrderProgressed) Type() string { return "movement_order_progressed" }

type MovementOrderRevised struct {
    FactionID         string
    AssetID           string
    NewOrder          MovementOrder
    Cause             string
    CausedByFactionID string
}
func (m MovementOrderRevised) Type() string { return "movement_order_revised" }

type MovementOrderCancelled struct {
    FactionID         string
    AssetID           string
    StrandedAt        Location // current hex at cancellation; WorldID empty if mid-flight
    Cause             string
    CausedByFactionID string
}
func (m MovementOrderCancelled) Type() string { return "movement_order_cancelled" }

type MovementOrderCompleted struct {
    FactionID         string
    AssetID           string
    FinalLocation     Location
    Cause             string
    CausedByFactionID string
}
func (m MovementOrderCompleted) Type() string { return "movement_order_completed" }
```

### Task 10 — `internal/faction/engine/mutation/mutation.go`

Add Apply cases for each new mutation. All five mutate `Asset.CurrentOrder` and (for `MovementOrderProgressed` / `MovementOrderCompleted` / `MovementOrderCancelled`) `Asset.Location`. Note `Asset.Location` is still a `string` in this phase — so `MovementOrderProgressed` and `MovementOrderCancelled` should set it to `""` (mid-flight sentinel — same as the old "Astral Sea" semantics, just empty for now). `MovementOrderCompleted` sets it to `FinalLocation.WorldID`. Phase 2 rewrites all of these to set the full `Location` struct.

```go
case domain.MovementOrderIssued:
    if faction, ok := factionState.Factions[v.FactionID]; ok {
        if asset, ok := faction.Assets[v.AssetID]; ok {
            order := v.Order
            asset.CurrentOrder = &order
            asset.Location = "" // WorldID clears on issue (Q3); Phase 2 swaps to Location{}.
        }
    }
case domain.MovementOrderProgressed:
    if faction, ok := factionState.Factions[v.FactionID]; ok {
        if asset, ok := faction.Assets[v.AssetID]; ok && asset.CurrentOrder != nil {
            asset.CurrentOrder.StepIdx = v.NewStepIdx
        }
    }
case domain.MovementOrderRevised:
    if faction, ok := factionState.Factions[v.FactionID]; ok {
        if asset, ok := faction.Assets[v.AssetID]; ok {
            order := v.NewOrder
            asset.CurrentOrder = &order
        }
    }
case domain.MovementOrderCancelled:
    if faction, ok := factionState.Factions[v.FactionID]; ok {
        if asset, ok := faction.Assets[v.AssetID]; ok {
            asset.CurrentOrder = nil
        }
    }
case domain.MovementOrderCompleted:
    if faction, ok := factionState.Factions[v.FactionID]; ok {
        if asset, ok := faction.Assets[v.AssetID]; ok {
            asset.CurrentOrder = nil
            asset.Location = v.FinalLocation.WorldID // Phase 2 swaps to v.FinalLocation.
        }
    }
```

### Task 11 — `internal/faction/engine/mutation/mutation_test.go`

Add minimal Apply unit tests for each of the five new mutations:

| Test | Assertion |
|---|---|
| `Issued` | `asset.CurrentOrder` is set; `asset.Location == ""` |
| `Progressed` | `asset.CurrentOrder.StepIdx` updates |
| `Revised` | `asset.CurrentOrder` swapped to new order |
| `Cancelled` | `asset.CurrentOrder == nil` |
| `Completed` | `asset.CurrentOrder == nil`; `asset.Location` matches final |

<br/>
<br/>

# Phase 2 — `Location` Struct Migration

Mechanical refactor. Every site that reads or writes a location string updates to read or write a `Location` struct (or its `WorldID` field, depending on the call site's intent). No new behavior — every existing test should pass after migration.

**Read this phase's commits in order — they have soft dependencies.** Commit 1 (Asset) is foundational; Commit 2 (Base) is parallel-shaped; Commit 3 (Homeworld) leans on Commit 1's helpers; Commit 4 (Mutations) requires all three to be reaching the new struct; Commit 5 (Astral Sea) is the cleanup pass.

After Phase 2: same external behavior, but `Asset.Location`, `Base.Location`, `Faction.Homeworld`, `AssetMoved.From/ToLocation`, and `HomeworldChanged.From/ToWorld` are all `Location` structs (or the equivalent — confirm at execution).

<br/>

## Commit 1 — `refactor: migrate Asset.Location to Location struct`

### Task 1 — `internal/faction/domain/asset.go`

Change `Asset.Location` from `string` to `Location`. Update the TOML tag scope as the new struct demands (nested `[location]` block, or inlined fields — see `Location` struct task in Phase 1 / Task 4).

### Task 2 — Update every reader

Sites identified during planning. Each gets `asset.Location` → `asset.Location.WorldID` (when the existing code is comparing world IDs) or `asset.Location` → `asset.Location` (when passing the whole thing through):

| File | Line(s) | Change |
|---|---|---|
| `internal/faction/engine/ability/ability.go` | 104 | `FromLocation: asset.Location` → keep as `Location` (Commit 4 migrates the mutation field) |
| `internal/faction/engine/ability/ability.go` | 121 | `asset.Location` (passed to `factionTestCandidates`) — update signature to take `worldID string`, pass `asset.Location.WorldID` |
| `internal/faction/engine/ability/ability.go` | 151, 174 | `== world` and `== asset.Location` comparisons → compare `WorldID` |
| `internal/faction/engine/ability/ability.go` | 210, 211 | `asset.Location != "" / seen[asset.Location]` → `asset.Location.WorldID != "" / seen[asset.Location.WorldID]` |
| `internal/faction/engine/ability/ability.go` | 215, 216 | same for `base.Location` (handled in Commit 2; flag it for Commit 2's audit) |
| `internal/faction/engine/action/actions/eligibility.go` | 37 | `base.Location == world` (Commit 2) |
| `internal/faction/engine/action/actions/refit_asset.go` | 75 | `Location: ra.refitOrder.OldAsset.Location` → preserve as `Location` struct |
| `internal/faction/engine/action/actions/buy_asset.go` | 123, 139 | `asset.Location != world / seen[asset.Location]` → `.WorldID` |
| `internal/faction/engine/action/actions/attack.go` | 79, 126, 140, 152, 159, 234 | `attacker.Location` → `attacker.Location.WorldID` (for index lookups and `World` field on `RollContext`); `liveDefenders` / `eligibleDefenders` / `factionBaseOnWorld` signatures may accept `string` or be widened — keep them on `string` (world ID) for this commit; Phase 7 widens them |
| `internal/faction/engine/action/actions/expand_influence.go` | 164, 219, 239, 297 | `asset.Location` → `.WorldID` |
| `internal/faction/engine/action/actions/expand_influence.go` | 243 | `base.Location` (Commit 2) |
| `internal/faction/engine/action/actions/seize_planet.go` | 72 | `asset.Location` → `.WorldID` |
| `internal/faction/engine/goal/lock.go` | 26 | `asset.Location == world` → `.WorldID` |
| `internal/faction/engine/goal/progress.go` | 80, 216 | `asset.Location ==` → `.WorldID` |
| `internal/faction/engine/goal/progress.go` | 102, 106, 334 | `base.Location` (Commit 2) |
| `internal/faction/engine/turn/bookkeeping.go` | 92 | `AssetRef{... Location: asset.Location}` — update `AssetRef.Location` field type to `Location` |
| `internal/faction/engine/world/world.go` | 38, 42 | `spatialMap.Location(asset.Location)` and `index.AssetsByLocation[asset.Location]` → use `asset.Location.WorldID` for both |
| `internal/faction/narrative/digest/build.go`, `resolve.go` | various | `base.Location` (Commit 2) |

Use grep at execution time to catch any I missed; the above is the audited list as of the plan session.

### Task 3 — Update writers

| File | Site | Change |
|---|---|---|
| `internal/faction/engine/mutation/mutation.go` | `case domain.AssetMoved` | `asset.Location = v.ToLocation` — `v.ToLocation` is still a string in this commit; in Commit 4 it becomes `Location`. For this commit, assign via `asset.Location = Location{WorldID: v.ToLocation}` as a placeholder. **Caveat:** the `HexCoords` and `Region` fields will be zero-valued; this is acceptable because Commit 4 immediately follows and rewrites this. |
| `internal/faction/engine/testharness/harness.go` | asset construction sites | populate the `Location` struct (probably needs a small helper `harness.LocationOnWorld(spatialMap, worldID) Location`) |

### Task 4 — Test the migration

Run `go build ./... && go test ./...`. All existing tests must pass. The `harness` change in Task 3 will ripple through every test that builds factions — expect compile errors first, then fix tests by replacing string-literal locations with `harness.LocationOnWorld(...)` calls.

<br/>

## Commit 2 — `refactor: migrate Base.Location to Location struct`

### Task 5 — `internal/faction/domain/base.go`

Change `Base.Location` from `string` to `Location`.

### Task 6 — Update readers/writers

| File | Line(s) | Change |
|---|---|---|
| `internal/faction/engine/ability/ability.go` | 215, 216 | `base.Location` → `.WorldID` |
| `internal/faction/engine/action/actions/eligibility.go` | 37 | `base.Location == world` → `.WorldID` |
| `internal/faction/engine/action/actions/expand_influence.go` | 243 | `base.Location` → `.WorldID` |
| `internal/faction/engine/goal/progress.go` | 102, 106, 334 | `base.Location` → `.WorldID` |
| `internal/faction/engine/world/world.go` | 45, 49 | same pattern as asset Commit 1 |
| `internal/faction/narrative/digest/resolve.go` | 61 | `return base.Location` — caller likely wants a display string; change return to `base.Location.WorldID` (or update the narrative consumer to handle `Location`) |
| `internal/faction/narrative/digest/build.go` | 338 | `expand.location = m.Base.Location` — same as resolve.go; preserve string-display intent by using `.WorldID` |
| `internal/faction/engine/testharness/harness.go` | base construction sites (lines around 209) | populate the `Location` struct |

### Task 7 — Test the migration

Same as Task 4 — `go build && go test`. Expect failures around digest/narrative; fix by using `.WorldID` where the consumer wants a string.

<br/>

## Commit 3 — `refactor: migrate Faction.Homeworld to Location struct`

### Task 8 — `internal/faction/domain/faction.go`

Change `Faction.Homeworld` from `string` to `Location`.

### Task 9 — Update readers/writers

| File | Line(s) | Change |
|---|---|---|
| `internal/faction/engine/action/actions/buy_asset.go` | 137 | `seen := map[string]struct{}{faction.Homeworld: {}}` → use `faction.Homeworld.WorldID` |
| `internal/faction/engine/goal/lock.go` | 61 | `FromWorld: faction.Homeworld` — `FromWorld` field on `HomeworldChanged` is migrated in Commit 4. For this commit, pass `faction.Homeworld.WorldID` (preserves string compat) |
| `internal/faction/engine/mutation/mutation.go` | 128 | `faction.Homeworld = v.ToWorld` → `faction.Homeworld = Location{WorldID: v.ToWorld}` placeholder; Commit 4 fixes this properly |
| `internal/faction/engine/testharness/harness.go` | 92 | `Homeworld: homeworld` → use `Location{WorldID: homeworld, ...}` via the helper from Commit 1 |

### Task 10 — Test the migration

`go build && go test`.

<br/>

## Commit 4 — `refactor: migrate AssetMoved and HomeworldChanged mutations to Location`

### Task 11 — `internal/faction/domain/mutation.go`

Change field types:

```go
type AssetMoved struct {
    FactionID         string
    AssetID           string
    FromLocation      Location // was string
    ToLocation        Location // was string
    Cause             string
    CausedByFactionID string
}

type HomeworldChanged struct {
    FactionID         string
    FromWorld         Location // was string
    ToWorld           Location // was string
    Cause             string
    CausedByFactionID string
}
```

### Task 12 — Fix emitters

| File | Site | Change |
|---|---|---|
| `internal/faction/engine/ability/ability.go` | 100–108 (`movementStepHandler`) | Build `AssetMoved` with full `Location` structs for `FromLocation` / `ToLocation`. **Note:** this handler is removed in Effort 3 Phase 5, but keep it correct in the interim. The `destination` returned by `collector.SelectMoveDestination` is currently a string world ID — wrap it via a helper that looks up the spatial map and builds the `Location`. |
| `internal/faction/engine/goal/lock.go` | 59–64 | `HomeworldChanged.FromWorld = faction.Homeworld` (struct copy); `ToWorld` built from `faction.ActiveGoal.TargetWorld` via the same spatial-map lookup helper |

### Task 13 — Fix Apply

`internal/faction/engine/mutation/mutation.go`:

```go
case domain.AssetMoved:
    if faction, ok := factionState.Factions[v.FactionID]; ok {
        if asset, ok := faction.Assets[v.AssetID]; ok {
            asset.Location = v.ToLocation // now a Location, not a string
        }
    }
case domain.HomeworldChanged:
    if faction, ok := factionState.Factions[v.FactionID]; ok {
        faction.Homeworld = v.ToWorld // now a Location
    }
```

Also revisit the Phase 1 placeholder cases (`MovementOrderProgressed`, `MovementOrderCompleted`, `MovementOrderCancelled`) and switch them to write proper `Location` structs:

- `Progressed`: `asset.Location = Location{WorldID: "", HexCoords: v.NewHexCoords, Region: <regionOfHex>}` — need the region; the emitter (Phase 3 Commit 2) should populate `Region` on the mutation if it isn't already, or pass a spatial-map lookup into Apply. Pick at execution.
- `Completed`: `asset.Location = v.FinalLocation`
- `Cancelled`: clears `CurrentOrder`; existing logic stands (don't touch `asset.Location` — it stays at its current mid-flight hex)
- `Issued`: `asset.Location = Location{WorldID: "", HexCoords: v.Order.Path[0], Region: v.Order.Origin.Region}` (Q3: clear WorldID on issue)

### Task 14 — Narrative + digest consumers

`internal/faction/narrative/wire_renderer.go` and `narrative/digest/build.go`:

- `headlineHomeworldTemplates` and `goalHomeworldTemplates` consume `ToWorld` as a display string; update template-fill sites to use `v.ToWorld.WorldID` (or extend to display name via spatial lookup if that pattern already exists in the renderer).
- `HomeworldChanged` lookup in `build.go:391` decodes the mutation — update field reads to use `Location` shape.

### Task 15 — `go build && go test ./...`

All tests pass. This is the largest migration commit; expect to spend time on test harness fixtures.

<br/>

## Commit 5 — `chore: retire "Astral Sea" sentinel`

### Task 16 — `internal/faction/engine/ability/ability.go`

Remove line 225: `worlds = append(worlds, "Astral Sea")`. The `worldsFromState` function is consumed by `collector.SelectMoveDestination` in the old `movementStepHandler`. With `Astral Sea` gone, the collector's "move to empty space" option is also gone — which is correct under the redesign (mid-flight is a state of having an order, not a destination).

### Task 17 — Sentinel sweep

```bash
grep -rn "Astral Sea\|astral_sea\|AstralSea" --include="*.go" --include="*.toml" .
```

Should return zero hits after Task 16. If anything else lingers, remove it.

### Task 18 — Discovery TOML / state fixture sweep

Grep test fixtures (`testdata/`, `internal/faction/state/testdata/`) for `"Astral Sea"` string literals. Update any fixture that uses it as a location to either:
- A real world ID (if the test cares about being on-world)
- An empty `Location{}` (if the test cares about being mid-flight; though no existing test should — Phase 1 didn't add mid-flight semantics yet)

### Task 19 — `go test ./...`

Final verification. All tests pass.

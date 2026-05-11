# Spatial Model — Effort 2: Faction Engine Integration

> Overview: [spatial-model-plan.md](spatial-model-plan.md)
> Depends on: [spatial-model-effort-1-plan.md](spatial-model-effort-1-plan.md) — `internal/spatial` must be complete before starting this effort.

Three phases. Each is its own execution session.

<br/>
<br/>

# Phase 2 — Engine Wiring

Wire `HybridMap` into the faction engine. Replace all O(factions × assets) world scans with spatial index lookups. No new rule enforcement — existing behavior is preserved; the scans just become index lookups.

**Prerequisite:** Before running the engine, the campaign TOML state file must be updated so all `Asset.Location` values are Fragment IDs (e.g. `"tartarus"`) rather than bare display names (e.g. `"Tartarus"`). Do this manually before starting Phase 2.

<br/>

## Commit 1 — `feat: add SPATIAL_DATA_DIR to config and DriftRating to AssetDefinition`

### Task 1 — `cmd/faction-manager/config.go`

Add `SpatialDataDir string` to the `Config` struct. Load from `SPATIAL_DATA_DIR` env var. Return an error from `LoadConfig()` if the var is unset, matching the existing pattern for `FACTION_DATA_DIR`.

### Task 2 — `internal/faction/domain/asset.go`

Add `DriftRating int` to `AssetDefinition` with TOML tag `drift_rating`. No default or fallback — assets without a drift rating leave the field at zero, which is valid for non-starship assets that never cross regions.

<br/>

## Commit 2 — `feat: load drift_costs.toml into Rulebook`

### Task 3 — `internal/faction/data/drift_costs.toml`

Create the default drift cost file in the faction data directory:

```toml
# drift_costs[i] = hex-equivalent crossing cost for drift rating (i+1)
drift_costs = [5, 4, 3, 2, 1]
```

### Task 4 — `internal/faction/rulebook/rulebook.go`

- Add `DriftCosts []int` field to `Rulebook`
- In `Load()`, decode `drift_costs.toml` from `dataDir` and populate the field
- Add helper:

```go
func (r *Rulebook) DriftCost(driftRating int) int {
    if driftRating < 1 || driftRating > len(r.DriftCosts) {
        return r.DriftCosts[0] // fallback to worst cost
    }
    return r.DriftCosts[driftRating-1]
}
```

<br/>

## Commit 3 — `feat: add SpatialIndex and wire HybridMap into engine`

### Task 5 — `internal/faction/engine/spatial_index.go`

Define the index type and build function:

```go
type SpatialIndex struct {
    AssetsByFragment map[string][]*domain.Asset
    BasesByFragment  map[string][]*domain.Base
}

func BuildSpatialIndex(factionState *state.FactionState, spatialMap *spatial.HybridMap) *SpatialIndex
```

`BuildSpatialIndex` scans all factions and all assets/bases in `factionState` once, grouping them by `Asset.Location` (which is now a Fragment ID). Only include assets with a `Location` that exists in `spatialMap` — log and skip any that don't resolve (handles stale state from before the migration).

### Task 6 — `internal/faction/engine/core.go`

- Add `spatialMap *spatial.HybridMap` field to `Engine`
- In the engine constructor (or `NewEngine`), call `spatial.LoadHybrid(cfg.SpatialDataDir)` and store the result
- At turn start (before action resolution begins), call `BuildSpatialIndex(factionState, engine.spatialMap)` and store the result on the engine for the duration of the turn
- Pass the `*SpatialIndex` into sub-engines that currently perform O(n) scans (see Commit 4)

<br/>

## Commit 4 — `refactor: replace O(n) world scans with SpatialIndex lookups`

Update the four scan sites. Each function gains a `*SpatialIndex` parameter. Callers (orchestrator, action factory) are updated to pass it through.

### Task 7 — `internal/faction/engine/action/actions/eligibility.go`

`eligibleDefenders(factionState, attackerFactionID, world, index)` — replace faction-loop with `index.AssetsByFragment[world]`, then filter out assets owned by `attackerFactionID`.

`liveDefenders(factionState, attackerFactionID, world, assetHPTracker, index)` — same replacement; apply the HP tracker filter on the index result.

### Task 8 — `internal/faction/engine/action/actions/expand_influence.go`

`rivalsOnWorld(factionState, factionID, world, index)` — replace faction-loop with `index.AssetsByFragment[world]`, collect unique owning faction IDs that are not `factionID`, return the corresponding `*Faction` values.

### Task 9 — `internal/faction/engine/goal/progress.go`

`worldHasRivalPresence(world, actingFactionID, factionState, index)` — check `len(index.AssetsByFragment[world]) > 0 || len(index.BasesByFragment[world]) > 0` after filtering out the acting faction's own entries.

`rivalHasPlanetaryGovernmentOnWorld(actingFactionID, world, factionState, index)` — iterate `index.BasesByFragment[world]`; return true if any base belongs to a rival and has type `Planetary Government`.

### Task 10 — `internal/faction/engine/action/actions/seize_planet.go`

Replace the defender scan in `seizePlanetTargetWorlds` with a lookup against `index.AssetsByFragment` and `index.BasesByFragment` to determine which worlds are contested (acting faction present AND rival present).

<br/>

## Commit 5 — `test: integration tests for SpatialIndex scan replacements`

### Task 11 — one harness test per replaced scan

Each test builds a `FactionState` with two factions and assets at known Fragment IDs, constructs a `SpatialIndex` from it, and asserts the scan-replacement function returns the same result the old loop would have. Cover:

| Test | Function under test | Assertion |
|---|---|---|
| Defender eligibility | `eligibleDefenders` | Returns only rival assets on target Fragment |
| Live defenders | `liveDefenders` | Excludes assets at zero HP per tracker |
| Rivals on world | `rivalsOnWorld` | Returns correct rival factions |
| Rival presence (assets) | `worldHasRivalPresence` | True when rival asset present |
| Rival presence (bases) | `worldHasRivalPresence` | True when rival base present, no assets |
| Planetary Government | `rivalHasPlanetaryGovernmentOnWorld` | True only when rival holds PG base |
| Seize target worlds | `seizePlanetTargetWorlds` | Returns only contested Fragments |

<br/>
<br/>

# Phase 3 — Validation Enforcement

Add rule enforcement that was previously skipped for lack of spatial data. Depends on Phase 2 (SpatialIndex and HybridMap must be wired in).

<br/>

## Commit 1 — `feat: enforce tech level in BuyAsset and ExpandInfluence`

### Task 1 — `internal/faction/engine/action/actions/buy_asset.go`

In `purchasableDefinitions` (or equivalent filter), add:

```go
fragment, ok := engine.spatialMap.Location(targetFragmentID)
if !ok {
    // skip — unknown fragment, treat as no-purchase
}
// filter: def.TechLevel <= fragment.TechLevel()
```

Only asset definitions whose `TechLevel` is ≤ the target Fragment's tech level are returned.

### Task 2 — `internal/faction/engine/action/actions/expand_influence.go`

In the destination-filter step of `Inputs` or `Validate`, look up each candidate Fragment via `engine.spatialMap.Location(id)` and exclude any whose tech level is below the minimum required by the assets the faction is trying to place. Identify the specific filter point by reading the current `expand_influence.go` at execution time.

<br/>

## Commit 2 — `feat: enforce P-flag in BuyAsset`

### Task 3 — `internal/faction/engine/action/actions/buy_asset.go`

After tech level filtering, add P-flag check: if a definition has the `P` flag, include it in the purchasable list only if the buying faction has a base of type `Planetary Government` on the target Fragment.

```go
if def.HasFlag(domain.FlagP) {
    bases := index.BasesByFragment[targetFragmentID]
    if !factionHasPlanetaryGovernment(faction.ID, bases) {
        continue
    }
}
```

Add `factionHasPlanetaryGovernment(factionID string, bases []*domain.Base) bool` as a package-level helper in `buy_asset.go`.

<br/>

## Commit 3 — `feat: enforce MaxHex in UseAssetAbility movement`

### Task 4 — `internal/faction/engine/ability/` — movement step handler

Locate the movement ability step handler (the code that processes `AbilityStep.Type == "movement"`). Add a distance check before accepting the GM's selected destination:

```go
driftCost := rulebook.DriftCost(assetDef.DriftRating)
dist, err := engine.spatialMap.Distance(asset.Location, destination, driftCost)
if err != nil || dist > step.MaxHex {
    // reject destination — return error to collector
}
```

<br/>

## Commit 4 — `test: validation enforcement integration tests`

### Task 5 — harness tests for all enforcement paths

| Test | Assertion |
|---|---|
| Tech level block | `BuyAsset` menu excludes TL-2 asset when target Fragment is TL-1 |
| Tech level pass | `BuyAsset` menu includes TL-1 asset when target Fragment is TL-2 |
| P-flag block | P-flagged asset excluded when faction has no Planetary Government on Fragment |
| P-flag pass | P-flagged asset included when faction holds Planetary Government |
| MaxHex block | Movement to Fragment beyond MaxHex is rejected |
| MaxHex pass | Movement to Fragment within MaxHex is allowed |

<br/>
<br/>

# Phase 4 — Distance Mechanics

Depends on Phase 2 (HybridMap wired in, DriftCost available).

<br/>

## Open Question — resolve before starting this phase

| # | Question | Where to decide |
|---|---|---|
| 1 | Which drift rating is used for `Change Homeworld` distance computation — the faction's best available starship, a fixed GM-provided value, or something else? | Discuss before Phase 4 execution session |

<br/>

## Commit 1 — `feat: compute Change Homeworld duration from spatial distance`

### Task 1 — locate Change Homeworld goal inputs

Read `internal/faction/engine/goal/` to find where `TurnsRemaining` is set for the `Change Homeworld` goal. This is the collector call where the GM currently enters a manual turn count.

### Task 2 — replace manual entry with computed distance

Replace the collector prompt for turn count with:

```go
crossingCost := rulebook.DriftCost(/* resolved drift rating — per open question */)
dist, err := engine.spatialMap.Distance(faction.HomeworldID, destinationID, crossingCost)
turnsRemaining := 1 + dist
```

Apply `GoalTurnsTick` with the computed value. If `Distance` returns an error (no path), surface it to the GM via the observer before aborting the goal selection.

<br/>

## Commit 2 — `test: Change Homeworld distance integration test`

### Task 3 — harness test

Build a two-region `HybridMap`, place a faction's homeworld in Region A, select a destination in Region B, and assert that `TurnsRemaining` equals `1 + expected distance` after the goal is set.

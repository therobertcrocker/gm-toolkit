# Asset Movement Redesign — Effort 3: Cutover & Integration

> Overview: [`asset-movement-redesign-plan.md`](asset-movement-redesign-plan.md)
> Depends on: [`asset-movement-redesign-effort-2-plan.md`](asset-movement-redesign-effort-2-plan.md) — all Effort 2 phases must be merged before starting Effort 3.

Three phases. Each is its own execution session.

After Effort 3 lands, the redesign is **live**: assets can actually move via `Speed`, transport abilities work, `ChangeHomeworld` is distance-aware, and attacks can target in-transit assets at the hex level. The old `AbilityStepMovement` is purged from the codebase entirely.

<br/>
<br/>

# Phase 5 — Asset TOML Migration & Self-Movement Step Removal

The cutover moment. Assets get `Speed` values; mixed-mode transport entries are reclassified as `type = "transport"`; self-mover entries are removed; the legacy `AbilityStepMovement` constant + handler + decoder case are deleted.

After Phase 5: assets in the rulebook actually move. The old `movementStepHandler` is gone; the only way an asset moves is via the Movement Phase.

<br/>

## Pre-Phase audit

Before starting Phase 5, run the full audit:

```bash
grep -n "type = \"movement\"" internal/faction/data/*_assets.toml
```

As of the plan session, the audit shows 17 entries across `cunning_assets.toml`, `force_assets.toml`, and `wealth_assets.toml`. For each entry, read its parent asset's `description` field to classify it as **self-mover** (asset moves itself only) or **transport-pattern** (asset moves other assets, possibly including itself). The discovery doc lists known examples; the audit below applies the same classification rules.

| Asset | File | Pattern | Action |
|---|---|---|---|
| FreighterContract (W?-???) | wealth | Transport (mixed; "move any one non-Force asset including this one") | Reclassify step → `transport`; assign `speed = 2` (self-mover part absorbed) |
| Surveyors (W2-004) | wealth | Self-mover | Remove step; assign `speed = 2` |
| Mercenaries (W3-003) | wealth | Self-mover | Remove step; assign `speed = 1` |
| ShippingCombine (W4-001) | wealth | Transport (multiple cargo, "non-Force assets including itself") | Reclassify step → `transport`; assign `speed = 2`. Note: multi-cargo support is a planning concern — confirm if the collector returns one cargo asset at a time or a slice. See Effort 2 Phase 4 Task 7. |
| (others) | wealth/force/cunning | TBD at execution | Classify by description |

**At Phase 5 execution time, complete the audit for the remaining ~13 entries.** Each entry's description text dictates the classification.

<br/>

## Commit 1 — `chore: assign Speed to formerly self-moving assets in TOMLs`

### Task 1 — Update each `*_assets.toml`

For every asset classified as self-mover, add the `speed` field next to `tech_level`:

```toml
[assets.Surveyors]
id = "W2-004"
...
speed = 2
```

Match the `speed` value to the original `max_hex` of the removed step (or use Robert's judgment if the description suggests otherwise — e.g., a "fast frigate" might warrant `speed = 3` even if its old `max_hex = 2`).

### Task 2 — Update each `*_assets.toml` for transport carriers

Transport-carrying assets also get a `speed` so they can self-move when not transporting. Default to the same value as the transport step's `max_hex`, unless the description implies a different self-speed.

### Task 3 — Rulebook test

Update `rulebook_test.go` fixtures to expect `Speed` populated correctly on the test assets.

<br/>

## Commit 2 — `refactor: reclassify retained transport ability steps in TOMLs`

### Task 4 — Update transport entries

For every transport-pattern entry, rename `type = "movement"` → `type = "transport"`:

```toml
[[assets.FreighterContract.ability.steps]]
type = "transport"
max_hex = 2
coin_cost = 1
```

### Task 5 — Remove self-mover entries

For every self-mover entry, delete the entire `[[assets.X.ability.steps]]` block. If the parent asset's `ability` no longer has any steps, the empty `ability = {}` or omitted `ability` should leave the asset with `Ability == nil` after decoding — verify the rulebook decoder handles this (`convertAbility` returns `nil` for nil records — check). If the asset has other ability steps (e.g. a faction_test step alongside the movement step), only remove the movement step.

### Task 6 — Mixed-mode reconciliation (Smugglers, Blockade Runner)

These two assets (per discovery) have descriptions implying both self-movement and transport. Under the redesign:

- Their self-movement is now their `Speed`
- Their transport ability stays at the original cost and `max_hex`
- Their `Speed` should match what their description implies the asset's pace is

Confirm at execution by reading their descriptions in the TOMLs and applying judgment.

### Task 7 — Rulebook integration test

A test that loads the full rulebook from `internal/faction/data/` and asserts:

- No asset definition has an `AbilityStepMovement`-typed step
- Transport-carriers have an `AbilityStepTransport` step with expected `MaxHex` and `CoinCost`
- Self-mover assets have `Speed > 0` and no movement-type ability step

<br/>

## Commit 3 — `chore: remove AbilityStepMovement constant, handler, and rulebook case`

### Task 8 — `internal/faction/domain/asset.go`

Delete the constant:

```diff
- AbilityStepMovement    AbilityStepType = "movement"
```

### Task 9 — `internal/faction/engine/ability/ability.go`

- Delete the `movementStepHandler` function
- Delete the `ae.stepHandlers[domain.AbilityStepMovement] = movementStepHandler` registration in `New()`
- Delete the `worldsFromState` helper (it was already pruned of `"Astral Sea"` in Effort 1 — at this point it's only called from `movementStepHandler` which is being deleted; verify and clean up)
- Delete `collector.SelectMoveDestination` from the `Collector` interface and every implementation — no remaining consumer

### Task 10 — `internal/faction/rulebook/rulebook.go`

Delete the `case "movement":` branch from `convertAbilityStep`:

```diff
- case "movement":
-     return domain.AbilityStep{
-         Type:     domain.AbilityStepMovement,
-         MaxHex:   r.MaxHex,
-         CoinCost: r.CoinCost,
-     }, nil
```

### Task 11 — `internal/faction/engine/ability/ability_test.go`

Delete tests that reference `AbilityStepMovement`. Any test that needs an ability with a movement-like step should reclassify to `AbilityStepTransport` (with synthetic cargo).

### Task 12 — `internal/faction/engine/action/actions/use_asset_ability_test.go`

Same as Task 11 — remove or reclassify any `AbilityStepMovement` references.

### Task 13 — Final grep

```bash
grep -rn "AbilityStepMovement\|movementStepHandler\|SelectMoveDestination" --include="*.go" .
```

Zero hits. If anything remains, remove it.

### Task 14 — `go test ./...`

All tests pass.

<br/>
<br/>

# Phase 6 — ChangeHomeworld Distance Reconciliation

Replace the hardcoded `TurnsRemaining: 3` at goal-initiation with `1 + spatial.Distance(homeworld, target, drift_costs[2])`. The lock-time tick logic at `internal/faction/engine/goal/lock.go:49` is unchanged — it already decrements `TurnsRemaining` each turn.

Retires the paused **Spatial Phase 4** work from `docs/initiatives/implementation/completed/spatial-model-effort-2-plan.md`.

After Phase 6: `ChangeHomeworld` goals' duration scales with hex distance, using the faction-level default drift rating (3).

<br/>

## Commit 1 — `feat: compute ChangeHomeworld TurnsRemaining from spatial distance`

### Task 1 — Locate the goal-initiation site

`internal/faction/engine/goal/progress.go:90` sets `TurnsRemaining: 3` when a `ChangeHomeworld` goal is initiated. Read the surrounding function (line ~75–100 in the current file) to understand the context — it's likely a `case` branch in a goal-selection collector flow.

### Task 2 — Replace hardcoded turns with computed distance

```go
crossingCost := rulebook.DriftCost(3) // faction-level default drift rating
dist, err := spatialMap.Distance(faction.Homeworld.WorldID, targetWorld.WorldID, crossingCost)
if err != nil {
    return nil, fmt.Errorf("change_homeworld: computing distance: %w", err)
}
turnsRemaining := 1 + dist
```

If the goal-initiation function doesn't currently have access to `spatialMap` or `rulebook`, plumb them through. The engine's existing wiring (Effort 1 Phase 1's `e.spatialMap`) makes this straightforward — the calling site is already inside engine code.

### Task 3 — Error surfacing

If `spatialMap.Distance` returns `ErrNoPath` (no path between the two worlds), the goal cannot be initiated. Surface this to the collector via the existing observer/error pattern and abort goal selection.

`ErrUnknownWorld` should be similarly surfaced — though it'd indicate a bug since both worlds should exist in the rulebook by the time selection runs.

<br/>

## Commit 2 — `test: ChangeHomeworld distance integration test`

### Task 4 — Harness test

`internal/faction/engine/goal/progress_test.go`:

Build a two-region `HybridMap` (or use the existing test fixture):
- Homeworld on world A in region 1
- Target world C in region 2 (forces a region boundary crossing)

Initiate `ChangeHomeworld` selecting C. Assert:
- `goal.TurnsRemaining == 1 + expectedDistance`
- Where `expectedDistance` matches what `spatial.Distance(A, C, DriftCost(3))` returns

### Task 5 — Same-region test

Single-region setup. Distance from A to B is purely hex-distance, no region cost. Assert `TurnsRemaining == 1 + hexDistance`.

### Task 6 — Verify existing tick still works

The existing `checkLockChangeHomeworld` logic (decrementing `TurnsRemaining` each turn) should continue to work unchanged. The harness test in Task 4 should run for `TurnsRemaining` turns and confirm:
- Each turn emits `GoalTurnsTick`
- The final turn emits `HomeworldChanged` and `GoalCompleted`

<br/>
<br/>

# Phase 7 — Hex-Level Attack Targeting

Add a hex-keyed lookup to `world.Index` and let `liveDefenders` / `eligibleDefenders` accept either a `WorldID` (for on-world attacks) or a `HexCoord` (for in-transit defenders). Combat resolution itself is unchanged.

After Phase 7: an attacker whose current hex matches an in-transit asset's hex can target that asset. The collector-level defender-selection UI is out of scope (per discovery).

Retires the paused **Spatial Phase 3 Commit 3** (`MaxHex` enforcement) implicitly — the original Commit 3 enforcement is no longer needed because movement no longer flows through `UseAssetAbility`. Self-movement is governed by `Speed`; transport range is enforced in the transport handler.

<br/>

## Commit 1 — `feat: add AssetsAtHex lookup to world.Index`

### Task 1 — `internal/faction/engine/world/world.go`

Extend `Index`:

```go
type Index struct {
    AssetsByLocation map[string][]*domain.Asset
    BasesByLocation  map[string][]*domain.Base
    AssetsByHex      map[spatial.HexCoord][]*domain.Asset
}
```

In `RebuildIndex`, populate `AssetsByHex` alongside `AssetsByLocation`:

```go
for _, faction := range factionState.Factions {
    for _, asset := range faction.Assets {
        hex := asset.Location.HexCoords
        index.AssetsByHex[hex] = append(index.AssetsByHex[hex], asset)

        if asset.Location.WorldID != "" {
            if _, ok := engine.spatialMap.Location(asset.Location.WorldID); !ok {
                continue
            }
            index.AssetsByLocation[asset.Location.WorldID] = append(index.AssetsByLocation[asset.Location.WorldID], asset)
        }
        // Mid-flight assets (empty WorldID) drop out of AssetsByLocation but
        // are present in AssetsByHex.
    }
    // (bases — bases are always on-world; same loop as before)
}
```

### Task 2 — Hex-index unit test

Build a fixture with three assets:
- Asset X on world A (hex coord matches A's hex)
- Asset Y mid-flight at hex (5, 3) with empty WorldID
- Asset Z on world A (same hex as X)

Assertions:
- `AssetsByLocation["a"]` contains X and Z but not Y
- `AssetsByHex[A.Hex]` contains X and Z
- `AssetsByHex[{5,3}]` contains Y

<br/>

## Commit 2 — `refactor: liveDefenders accepts WorldID or HexCoord`

### Task 3 — `internal/faction/engine/action/actions/eligibility.go`

Widen `eligibleDefenders` and `liveDefenders` to accept a location-handle that's either a world ID or a hex coord. Two clean shapes:

**Option A:** overload via a typed enum:
```go
type DefenderTarget struct {
    WorldID string
    Hex     *spatial.HexCoord // non-nil means hex-targeted (overrides WorldID)
}
```

**Option B:** two functions: `eligibleDefendersOnWorld(...)` and `eligibleDefendersAtHex(...)`. Caller picks which based on whether the attacker is on-world or in-transit.

**Recommended: Option B** — simpler, each function has a single responsibility. Attacker-side code branches on `attacker.Location.WorldID != ""` and calls the appropriate function.

### Task 4 — Update `AttackAction.Resolve`

`internal/faction/engine/action/actions/attack.go:79`:

```go
var defenders []*domain.Asset
if attacker.Location.WorldID != "" {
    defenders = liveDefendersOnWorld(faction.ID, attacker.Location.WorldID, assetHPTracker, attack.index)
} else {
    defenders = liveDefendersAtHex(faction.ID, attacker.Location.HexCoords, assetHPTracker, attack.index)
}
```

Same pattern for `attackerHasTarget` at line 229.

### Task 5 — `internal/faction/engine/action/actions/eligibility.go` implementations

`eligibleDefendersOnWorld` is the existing function renamed.

`eligibleDefendersAtHex`:
```go
func eligibleDefendersAtHex(attackerFactionID string, hex spatial.HexCoord, index *world.Index) []*domain.Asset {
    candidates := index.AssetsByHex[hex]
    return filterOpposingDefenders(attackerFactionID, candidates)
}
```

Extract `filterOpposingDefenders` as a private helper shared by both functions (exclude attacker's own assets; whatever stealth/eligibility filters the original function applied).

<br/>

## Commit 3 — `test: hex-level attack targeting integration tests`

### Task 6 — In-transit defender test

Build a two-faction state:
- Faction A: attacker asset X currently mid-flight at hex (5, 3) with empty WorldID
- Faction B: defender asset Y currently mid-flight at the same hex (5, 3) with empty WorldID

Run an `Attack` action. Assertions:
- `eligibleDefenders` includes Y
- Combat resolves normally; damage applies to Y as usual

### Task 7 — Mixed hex test

Same setup but:
- Faction A's attacker is on world C at hex (5, 3)
- Faction B's defender is mid-flight passing through hex (5, 3)

`liveDefendersAtHex` is the right function here? No — the attacker is on-world, so `liveDefendersOnWorld(C)` is used. Y is mid-flight (empty WorldID) → it's not in `AssetsByLocation["c"]` → it's **not** a target.

**This is the intended behavior:** an on-world attacker can only target on-world defenders. An in-transit attacker can target in-transit defenders (same hex). On-world ↔ in-transit cross-targeting is not supported by this commit — flag for Robert at execution if the rule should be different.

### Task 8 — No defenders test

Faction A attacker at an empty hex (no other factions there). `eligibleDefenders` returns empty; `Attack` action validates as no-target.

# Asset Movement Redesign — Effort 3: Cutover & Integration

> Overview: [`asset-movement-redesign-plan.md`](asset-movement-redesign-plan.md)
> Depends on: [`asset-movement-redesign-effort-2-plan.md`](asset-movement-redesign-effort-2-plan.md) — all Effort 2 commits must be on `feature/movement-redesign` before starting Effort 3. The branch is not merged to `main` until the whole movement initiative (Efforts 1–3) is complete.

Three phases. Each is its own execution session.

After Effort 3 lands, the redesign is **live**: assets can actually move via `Speed`, transport behavior works through the S-flag hook handler, `ChangeHomeworld` is distance-aware, and attacks can target in-transit assets at the hex level. The old `[[ability.steps]]` movement TOML rows are purged from the rulebook data; `AbilityStepMovement` itself was deleted earlier under R-001's ability-engine retirement.

<br/>
<br/>

# Phase 5 — Asset TOML Migration

The cutover moment. Assets get `Speed` values; transport-pattern entries are migrated from `[[ability.steps]] type = "movement"` rows to the new `[assets.X.transport]` block (with their `flags` flipped from `["A"]` to `["S"]`); self-mover entries lose their movement step rows entirely.

The `AbilityStepMovement` constant, `movementStepHandler`, and `case "movement":` decoder branch were already deleted under R-001's ability-engine retirement. Phase 5 only touches TOML data plus a rulebook integration test — no Go-code purge work remains.

After Phase 5: assets in the rulebook actually move. Effort 2's Movement Phase + transport reactor have data to operate on.

<br/>

## Pre-Phase audit

Before starting Phase 5, run the full audit:

```bash
grep -n "type = \"movement\"" internal/faction/data/*_assets.toml
```

These rows are currently silently ignored by the rulebook decoder (R-001 removed the `case "movement":` branch in `convertAbilityStep`). The audit lists the TOML rows that still need to be removed *and* classified — each parent asset is either a **self-mover** (the row just gets removed, plus `speed = N` added at top level) or a **transport-pattern** (the row is removed *and* replaced with a top-level `[assets.X.transport]` block, *and* the asset's `flags` flips from `["A"]` to `["S"]`, *and* `speed = N` is added).

| Asset | File | Pattern | Action |
|---|---|---|---|
| FreighterContract (W?-???) | wealth | Transport (mixed; "move any one non-Force asset including this one") | Remove `[[ability.steps]]` row; add `[transport]` block (`max_hex = N`, `coin_cost = N`, `cargo_types`, `max_cargo`); flip `flags` to `["S"]`; assign top-level `speed = 2` |
| Surveyors (W2-004) | wealth | Self-mover | Remove `[[ability.steps]]` row; assign `speed = 2`; flags unchanged |
| Mercenaries (W3-003) | wealth | Self-mover | Remove `[[ability.steps]]` row; assign `speed = 1`; flags unchanged |
| ShippingCombine (W4-001) | wealth | Transport (multiple cargo, "non-Force assets including itself") | Remove `[[ability.steps]]` row; add `[transport]` block with `max_cargo > 1`; flip `flags` to `["S"]`; assign `speed = 2`. Note: multi-cargo collector behavior is covered by Effort 2 Phase 4 Commit 1 Task 6 — the `MaxCargo` cap is enforced there. |
| Smugglers (C1-001) | cunning | Transport (mixed; "transport itself and/or any one Special Forces unit") | Remove `[[ability.steps]]` row; add `[transport]` block (`max_hex = 2`, `coin_cost = 1`, `cargo_types = ["Special Forces"]`, `max_cargo = 1`); flip `flags` from `["A"]` to `["S"]`; assign `speed = 2` |
| BlockadeRunners (W5-003) | wealth | Transport (mixed; "move itself or any one Military Unit or Special Forces") | Remove `[[ability.steps]]` row; add `[transport]` block (`max_hex = 3`, `coin_cost = 2`, `cargo_types = ["Military Unit", "Special Forces"]`, `max_cargo = 1`); flip `flags` from `["A"]` to `["S"]`; assign `speed = 3` |
| (others) | wealth/force/cunning | TBD at execution | Classify by description |

**At Phase 5 execution time, complete the audit for any remaining entries.** Each entry's description text dictates the classification.

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

## Commit 2 — `feat: migrate transport entries to [transport] block and S-flag`

> **Session handoff (Commit 1 complete).** Speed values are written and the decoder wires them through. Commit 2 is new scope vs. the original plan: `ExcludeCategories []FactionStat` was added to `TransportProfile` to express "non-Force" and "non-Starship" restrictions that can't be represented by `cargo_types` alone (those filter by asset `Type`; category is a separate field). All three Go touches are small — domain field, decoder field + loop, one eligibility guard — but must land before the TOML blocks are written.

### Task 3a — Add `ExcludeCategories` to `TransportProfile`

Three files, all small:

1. **`internal/faction/domain/asset.go`** — add `ExcludeCategories []FactionStat` to `TransportProfile`.
2. **`internal/faction/rulebook/rulebook.go`** — add `ExcludeCategories []string \`toml:"exclude_categories"\`` to `transportRecord`; loop through it in `convertTransport` using `toFactionStat`, appending to a slice, and populate `TransportProfile.ExcludeCategories`.
3. **`internal/faction/engine/orchestrator.go`** — in `eligibleCargoForTransport` (line 504), add `!slices.Contains(profile.ExcludeCategories, def.Category)` to the existing `slices.Contains(profile.CargoTypes, def.Type)` guard.

### Task 4 — Convert transport-pattern assets

Eight assets get a `[transport]` block. Reference table (all values resolved during Commit 1 planning):

| Asset | max_hex | coin_cost | cargo_types | max_cargo | exclude_categories |
|---|---|---|---|---|---|
| FreighterContract (W2-001) | 2 | 1 | all types | 1 | `["Force"]` |
| ShippingCombine (W4-001) | 2 | 1 | all types | 10 | `["Force"]` |
| BlockadeRunners (W5-003) | 3 | 2 | `["Military Unit","Special Forces"]` | 1 | — |
| Smugglers (C1-001) | 2 | 1 | `["Special Forces"]` | 1 | — |
| HeavyDropAssets (F2-001) | 1 | 1 | non-Starship | 1 | — |
| BeachheadLanders (F4-001) | 1 | 1 | all types | 10 | — |
| ExtendedTheater (F4-002) | 2 | 1 | non-Starship | 1 | — |
| DeepStrikeLanders (F7-001) | 3 | 2 | non-Starship | 1 | — |

"all types" = `["Facility", "Starship", "Military Unit", "Special Forces", "Tactic", "Logistics Facility"]`

"non-Starship" = `["Facility", "Military Unit", "Special Forces", "Tactic", "Logistics Facility"]`

Steps per asset:
1. Delete the `[[assets.X.ability.steps]]` block carrying `type = "movement"`.
2. Add `[assets.X.transport]` block using the values above.
3. Flip `flags = ["A"]` → `flags = ["S"]` on the parent asset. (Transport is reactive, not player-action-triggered.) **Note:** the `S` flag (`FlagSpecial`) is currently inert — no engine code branches on it; only the `[transport]` block presence drives reactor registration (see Effort 2 Phase 4 Commit 2 Task 2).
4. The `speed = N` line (added in Commit 1) handles self-movement.

If the parent asset's `ability` no longer has any remaining content, remove the empty `[assets.X.ability]` header along with the step row.

### Task 5 — Remove self-mover and deferred-transport entries

**Self-movers** (7 assets — Surveyors, Mercenaries, ScavengerFleet, Seductress, StrikeFleet, SpaceMarines, CapitalFleet): delete the `[[assets.X.ability.steps]]` block. Keep `flags` unchanged (still `["A"]` if the asset has other action-triggered abilities; cleared entirely if movement was its only ability).

**Deferred-transport assets** (CovertShipping C3-003, CovertTransitNet C6-002): these are stationary logistics hubs whose "move other assets" mechanic is deferred to a future special-ability system. Delete the `[[ability.steps]] type = "movement"` block; do NOT add a `[transport]` block. Keep `flags` unchanged on both — the `A` flag is a placeholder for the future action ability.

### Task 6 — Rulebook integration test

A test that loads the full rulebook from `internal/faction/data/` and asserts:

- No asset definition's TOML row carries `type = "movement"` in any `[[ability.steps]]` block (rulebook decoder no longer accepts this; this assertion is a data-shape sanity check)
- Each transport-pattern asset has `def.Transport != nil` with `MaxHex`, `CoinCost`, `CargoTypes`, `MaxCargo`, and `ExcludeCategories` matching the expected per-asset values from the table above
- Each transport-pattern asset has `domain.FlagSpecial` (`"S"`) in `def.Flags`
- Each self-mover asset has `def.Speed > 0` and `def.Transport == nil`
- CovertShipping and CovertTransitNet have `def.Transport == nil` (deferred) and `def.Ability == nil` (step removed)

<br/>
<br/>

# Phase 6 — ChangeHomeworld: initiation pathway + distance-based duration

Build the missing piece. Today the `ChangeHomeworld` goal handler is registered and its `CheckLock` ticks `TurnsRemaining` down + emits completion mutations, but **nothing in the codebase initiates the goal** — there is no action, no collector entry, no `GoalInitiated` emission for `G-012`. The previous plan iteration assumed an existing initiation site with a hardcoded `TurnsRemaining: 3`; that site does not exist.

Phase 6 covers two commits:

1. Add a `ChangeHomeworld` action that mirrors `SeizePlanet`: the player picks it via `SelectAction`, it computes `1 + Distance(homeworld, target, DriftCost(3))`, and emits `GoalInitiated` + `GoalPhaseAdvanced` to set `ActiveGoal.TargetWorld` and `TurnsRemaining`.
2. Integration tests covering end-to-end: take action → tick → complete.

`change_homeworld.go`'s existing handler structure (tick + completion in `CheckLock`, `UpdateProgress` as a no-op) is **already correct** — see Shape notes below. No refactor commit needed.

Retires the paused **Spatial Phase 4** work from `docs/initiatives/implementation/completed/spatial-model-effort-2-plan.md`.

After Phase 6: a faction can take the `ChangeHomeworld` action targeting any reachable world where it has a Base of Influence; `TurnsRemaining` scales with hex distance (using drift rating 3 by default); the goal ticks down each turn via the existing `CheckLock` logic; on completion the homeworld changes.

### Shape notes

- **Decision 112 is being reversed.** D112 stated "Change Homeworld is a goal, not a registered action," with the rationale that multi-turn state belongs to the Goal Engine. The reversal: D113 already established the precedent that an *action* can initiate a multi-turn goal (SeizePlanet → PlanetarySeizure). ChangeHomeworld follows the same shape — the action initiates, the goal handler manages subsequent turns. Goal-Engine ownership of multi-turn state is preserved. The reversal must be captured as a new decisions-log entry, written **before** the Commit 1 commit per CLAUDE.md.
- **Decision 122 still holds — do not touch the goal handler's structure.** D122: "Change Homeworld completion fires inside `CheckLock` when `TurnsRemaining` reaches 0, not inside `UpdateProgress`." This remains correct in current code: `orchestrator.go:85` early-returns for `LockSkip` factions before reaching `runActionPhase`, and `UpdateProgress` is only called inside `runActionPhase` (line 284). For a `LockSkip` faction, `UpdateProgress` never fires. Tick + completion must live in `CheckLock`. The current `change_homeworld.go` already does this correctly; `UpdateProgress` is a no-op and stays that way (there's no event-driven phase advancement for this goal — it's a pure timer).
- **PlanetarySeizure is the precedent for goal-handler structure.** Its phase-2 `CheckLock` ticks `TurnsRemaining` and emits `GoalCompleted` + `TagAdded` + `XPAwarded` directly when `newTurns == 0`. ChangeHomeworld's `CheckLock` already mirrors this shape. The only differences: ChangeHomeworld uses `LockSkip` (transit ⇒ no action) where PlanetarySeizure phase 2 uses `LockNone` (occupation ⇒ faction can still act).
- **Action contract.** `SeizePlanet` implements the full `Action` interface — `Name()`, `Validate()`, `Inputs()` (collector call), `Resolve()` (internal computation), `Output()` (mutation emission) — and emits a single `GoalInitiated` mutation from `Output()`. The constructor takes `(collector action.Collector, index *world.Index)` and stashes them on the struct. `ChangeHomeworld` follows the contract; it adds a `*world.WorldEngine` dep for the `Distance` lookup.
- **GoalInitiated + GoalPhaseAdvanced pair.** `GoalInitiated` (`domain/mutation.go:245`) carries `FactionID`, `GoalID`, `ProcessPhase`, `TargetWorld`, `Cause`, `CausedByFactionID` — no `TurnsRemaining`. Rather than extend the shared mutation type (which would touch `SeizePlanet` too), `Output()` emits the pair: `GoalInitiated` sets `ActiveGoal.TargetWorld`, then `GoalPhaseAdvanced{ProcessPhase: 0, TurnsRemaining: 1+dist}` writes the timer. The mutation engine's apply for `GoalPhaseAdvanced` (`mutation/mutation.go:165`) already writes `TurnsRemaining` directly.
- **WorldEngine needs a public Distance method.** `WorldEngine.spatialMap` is private; only `Location(id)` is exposed today. Phase 6 adds `func (engine *WorldEngine) Distance(from, to spatial.RegionHex, crossingCost int) (int, error)` as a thin passthrough — same shape as the `HexRouter` interface method (Decision 222) and `RegionMap.Distance` underneath. The action receives `*world.WorldEngine` from the engine wiring.
- **Distance is computed once, at initiation.** The action's `Resolve()` computes hex distance from the faction's current homeworld to the chosen target, stores it on the action struct, and `Output()` writes it into `TurnsRemaining`. The goal handler doesn't re-compute distance mid-transit; the timer ticks down deterministically.
- **Base-of-Influence precondition.** Per `goals.toml` (line 70), ChangeHomeworld targets "a world where it has a Base of Influence." Every `domain.Base` is a Base of Influence by definition (`domain/base.go:3`), so the precondition is "faction.Bases has a Base whose `Location.WorldID == target`." The `changeHomeworldTargets()` eligibility filter enforces this alongside the reachability check.
- **Location field naming (Decision 226).** `domain.Location` uses `WorldID string` + `RegionHex spatial.RegionHex` — no `HexCoords` field. `faction.Homeworld` is type `Location`, so the routing input is `faction.Homeworld.RegionHex`.

<br/>

## Commit 1 — `feat: ChangeHomeworld action with distance-based TurnsRemaining`

Adds the missing initiation pathway. After this commit, a faction can take the `ChangeHomeworld` action; the goal starts with a distance-derived timer; the existing tick + completion logic in `CheckLock` runs it down.

### Pre-commit — decisions-log entry

Before staging this commit, append a new entry to `docs/dev_journals/faction-manager/decisions-log.md` under the `feature/movement-redesign — engine (Effort 2)` section (or a new `feature/movement-redesign — Effort 3` section if cleaner). The entry text:

> **D###** — D112 reversed: ChangeHomeworld is initiated by a registered action (`ChangeHomeworld`), mirroring SeizePlanet→PlanetarySeizure (D113). | D112's rationale (Goal-Engine ownership of multi-turn state) is preserved — the action only emits `GoalInitiated` + `GoalPhaseAdvanced`; subsequent turns are managed by the goal handler's `CheckLock` per D122. The goal-engine-driven initiation pathway D112 implied was never built; the action shape is symmetric with the only other multi-turn goal in the codebase.

Update the section's index entry to include the new decision number.

### Task 1 — `internal/faction/engine/world/world.go`

Add a public `Distance` method on `WorldEngine`:

```go
func (engine *WorldEngine) Distance(from, to spatial.RegionHex, crossingCost int) (int, error) {
    return engine.spatialMap.Distance(from, to, crossingCost)
}
```

Thin passthrough — matches the existing `Location(id)` shape.

### Task 2 — New action `internal/faction/engine/action/actions/change_homeworld.go`

Mirror `seize_planet.go`'s full structure (`Name`/`Validate`/`Inputs`/`Resolve`/`Output`). The struct stashes deps via constructor; `Inputs` collects the target via the collector; `Resolve` computes distance and stores it; `Output` emits the mutation pair.

```go
type ChangeHomeworld struct {
    factionID      string
    collector      action.Collector
    world          *world.WorldEngine
    rulebook       *rulebook.Rulebook
    targetWorld    *domain.Location
    turnsRemaining int
}

func NewChangeHomeworld(collector action.Collector, worldEngine *world.WorldEngine, rulebook *rulebook.Rulebook) *ChangeHomeworld {
    return &ChangeHomeworld{collector: collector, world: worldEngine, rulebook: rulebook}
}

func (c *ChangeHomeworld) Name() string { return "Change Homeworld" }

func (c *ChangeHomeworld) Validate(faction *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) bool {
    if faction.ActiveGoal != nil {
        return false // can't initiate while another goal is active
    }
    return len(changeHomeworldTargets(faction, c.world)) > 0
}

func (c *ChangeHomeworld) Inputs(faction *domain.Faction, factionState *state.FactionState, _ *rulebook.Rulebook) error {
    targetWorldID, err := c.collector.SelectChangeHomeworldTarget(faction, factionState)
    if err != nil {
        return fmt.Errorf("change homeworld: %w", err)
    }
    target, ok := c.world.Location(targetWorldID)
    if !ok {
        return fmt.Errorf("change homeworld: unknown world %q", targetWorldID)
    }
    c.targetWorld = &domain.Location{WorldID: targetWorldID, RegionHex: target.RegionHex()}
    c.factionID = faction.ID
    return nil
}

func (c *ChangeHomeworld) Resolve(faction *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) error {
    if c.targetWorld == nil {
        return fmt.Errorf("change homeworld: no target selected")
    }
    crossingCost := c.rulebook.DriftCost(3) // faction-level default drift rating
    dist, err := c.world.Distance(faction.Homeworld.RegionHex, c.targetWorld.RegionHex, crossingCost)
    if err != nil {
        return fmt.Errorf("change homeworld: computing distance: %w", err)
    }
    c.turnsRemaining = 1 + dist
    return nil
}

func (c *ChangeHomeworld) Output() ([]domain.Mutation, error) {
    return []domain.Mutation{
        domain.GoalInitiated{
            FactionID:         c.factionID,
            GoalID:            "G-012",
            ProcessPhase:      0,
            TargetWorld:       *c.targetWorld,
            Cause:             "change homeworld",
            CausedByFactionID: c.factionID,
        },
        domain.GoalPhaseAdvanced{
            FactionID:      c.factionID,
            GoalID:         "G-012",
            ProcessPhase:   0,
            TurnsRemaining: c.turnsRemaining,
        },
    }, nil
}
```

`changeHomeworldTargets(faction, worldEngine)` is the analog of `seizePlanetTargetWorlds`. It returns the list of world IDs that satisfy **both** preconditions:

1. **Reachability** — `Distance(faction.Homeworld.RegionHex, world.RegionHex, DriftCost(3))` returns no error.
2. **Base of Influence at target** — `faction.Bases` contains a `Base` whose `Location.WorldID == world.ID`.

The current homeworld is excluded (no point relocating to where you already are). Because the Base-of-Influence filter typically narrows the candidate set drastically (a faction has only a handful of bases), iterate `faction.Bases` first and check reachability per-base:

```go
func changeHomeworldTargets(faction *domain.Faction, worldEngine *world.WorldEngine, rulebook *rulebook.Rulebook) []string {
    var targets []string
    crossingCost := rulebook.DriftCost(3)
    for _, base := range faction.Bases {
        if base.Location.WorldID == faction.Homeworld.WorldID {
            continue // current homeworld is excluded
        }
        target, ok := worldEngine.Location(base.Location.WorldID)
        if !ok {
            continue
        }
        if _, err := worldEngine.Distance(faction.Homeworld.RegionHex, target.RegionHex(), crossingCost); err != nil {
            continue
        }
        targets = append(targets, base.Location.WorldID)
    }
    sort.Strings(targets)
    return targets
}
```

Pass `rulebook` to the helper from the action's `Validate` (the struct already holds it).

### Task 3 — Add `SelectChangeHomeworldTarget` to `action.Collector`

Add the method to `internal/faction/engine/action/collector.go` (matching the existing `SelectSeizeTarget(faction, factionState) (string, error)` signature). Stub the testharness implementation in `internal/faction/engine/testharness/collector.go` with a configurable response field. Update any mock collector (`internal/faction/engine/action/actions/mocks/`) to add the method — verify by build.

### Task 4 — Register the action

Wire `NewChangeHomeworld` into the action engine's action list (same place `SeizePlanet` is constructed and registered). The action needs the new `*world.WorldEngine` dep, which the engine has on `core.go:41`.

### Task 5 — Error surfacing

`Distance` returns `ErrNoPath` when no path exists between the two worlds. `Resolve` wraps it as `change homeworld: computing distance: <err>`; the orchestrator's action-error pathway surfaces it to the observer.

`ErrUnknownWorld` would indicate a bug (target world picked from the eligibility set must exist in the rulebook); wrap identically.

If `Resolve` errors out, `Output` is never called — no mutations are emitted and `faction.ActiveGoal` remains `nil`. This is the desired no-partial-state guarantee tested in Commit 2 Task 9.

<br/>

## Commit 2 — `test: ChangeHomeworld initiation + lifecycle integration tests`

End-to-end coverage of the new action and the existing goal handler.

### Task 6 — Cross-region distance test

`internal/faction/engine/action/actions/change_homeworld_test.go` (or extend `progress_test.go` — pick by what minimizes test-fixture duplication):

Build a two-region `RegionMap` (use the existing test fixture if available; else copy from `internal/spatial`'s test data):
- Homeworld on world A in region 1
- Faction has a Base on world C in region 2 (forces a region boundary crossing)

Take the `ChangeHomeworld` action with target C. Assert:
- `GoalInitiated` and `GoalPhaseAdvanced` both emitted
- After Apply: `faction.ActiveGoal.GoalID == "G-012"`, `faction.ActiveGoal.TargetWorld.WorldID == "c"`, `faction.ActiveGoal.TurnsRemaining == 1 + expectedDistance`
- Where `expectedDistance == world.Distance(A.RegionHex, C.RegionHex, DriftCost(3))`

### Task 7 — Same-region distance test

Single-region setup. Faction has a Base on world B in the same region. Distance from A to B is purely hex-distance, no region cost. Assert `ActiveGoal.TurnsRemaining == 1 + hexDistance`.

### Task 8 — Full lifecycle test

Use the cross-region setup from Task 6. Run the engine for `TurnsRemaining` turns and confirm:
- Each turn: `CheckLock` emits `GoalTurnsTick`, returns `LockSkip` (action phase is skipped, `UpdateProgress` never fires)
- Final turn: `CheckLock` emits the tick + `HomeworldChanged` + `GoalCompleted`
- After Apply on the final turn: `faction.Homeworld.WorldID == "c"`, `faction.ActiveGoal == nil`

### Task 9 — Error path test

Build a setup where the faction has no Base on any non-homeworld (i.e. eligible target set is empty). Assert `Validate` returns false — the action is not offered. Then build a second setup with a Base on an unreachable world: assert `Validate` returns true (Base exists) but `Resolve` errors out and `faction.ActiveGoal` remains `nil` (no partial state). The first case proves the eligibility filter; the second proves the no-partial-state guarantee.

### Task 10 — Base-of-Influence precondition test

Two-region setup:
- Faction has bases on worlds B (same region as homeworld A) and C (different region)
- A third world D is in the cross-region region but the faction has no base there

Assert `changeHomeworldTargets()` returns exactly `["b", "c"]` — D is excluded despite being reachable. Verify the `Validate` → `Inputs` flow rejects a collector that tries to pick D (the collector contract assumes targets are filtered, but if a test injects an out-of-set value, `Inputs` should not silently accept it — flag for execution whether to add an explicit re-check).

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

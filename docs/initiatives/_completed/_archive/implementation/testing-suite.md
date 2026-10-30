# Implementation Plan: Comprehensive Test Suite

## Context

The engine is feature-complete (all 9 actions, Goal Engine, narrative renderer). Most internal packages have targeted unit tests, but 9 of 11 action types have no unit tests, the Goal Engine and Ability Engine have none, and the orchestrator integration tests cover only 3 of ~15 meaningful scenarios. This plan fills those gaps in three execution sessions (phases). `cmd/` (TUI/CLI) is explicitly out of scope.

---

## Phase 1: Action Unit Tests

**Deliverable:** `_test.go` for every untested action in `internal/faction/engine/actions/`.

### Pattern (from `attack_test.go`)

- Package declaration: `package actions` (same-package access)
- Real rulebook: `const testDataDir = "../data"` → `loader.LoadRulebook(testDataDir)`
- Minimal `fakeCollector` struct per file — only implement the methods the action calls; all others `panic("not used in <action> tests")`
- Fixture helpers: build a `*state.FactionState` with just enough factions/assets/bases to exercise the case
- Table-driven where multiple Validate/Resolve variants share setup; separate named funcs where scenarios diverge

### Files to create

#### `simple_actions_test.go`

Groups four low-complexity actions. One shared `fakeCollector` stub (all methods that aren't needed panic).

| Action | Validate cases | Resolve assertions |
|---|---|---|
| **SellAsset** | no assets → false; has assets → true | `AssetRemoved` + `CoinDelta` (cause: "sell") |
| **RepairFaction** | HP full → false; no coin → false; HP damaged + coin → true | `FactionHPDelta` + `CoinDelta` |
| **AbandonGoal** | no active goal → false; has goal → true | `GoalCleared` mutation present |
| **Bribe** | no coin → false; no rival bases on any world → false; valid target → true | `BaseHPDelta` on target base + `CoinDelta` |

For Bribe, the `fakeCollector.SelectBribeTarget` returns a scripted base and coin amount.

#### `buy_asset_test.go`

- Validate: faction with no purchasable assets (stat rating too low) → false; faction with eligible assets and enough coin → true; not enough coin → false
- Resolve: `AssetAdded` mutation with `Ready = false`; `CoinDelta` with correct amount; asset appears in Output()
- `SelectBuyOrder` returns a scripted `BuyOrder{DefinitionID: "F1-001", World: "Tartarus"}`

#### `repair_asset_test.go`

- Validate: no damaged assets → false; damaged assets but no coin → false; damaged + coin → true
- Resolve: one `AssetHPDelta` per selected asset; `CoinDelta` totaling repair costs
- `SelectRepairOrders` returns a scripted `[]RepairOrder`

#### `refit_asset_test.go`

- Validate: no assets → false; no affordable upgrade path → false; valid refit option exists → true
- Resolve: `AssetRemoved` for old asset; `AssetAdded` for new definition; coin delta matches cost difference
- `SelectRefitOrder` returns a scripted `RefitOrder`

#### `seize_planet_test.go`

- Validate: faction has no Force assets on target world → false; target world already has a faction BoI → false; valid target → true
- Resolve: appropriate seizure mutation(s) emitted with correct world and faction IDs
- `SelectSeizeTarget` returns a scripted world string

#### `expand_influence_test.go`

Three sub-scenarios matching the three `ExpandInfluence.Mode` branches:

1. **NewBase** — `SelectExpandInfluenceOrder` returns `Mode: ModeNewBase, World: "Krylos"`. Assert `BaseAdded` mutation + `CoinDelta`. No rivals → no contested roll needed.
2. **Reinforce** — existing base below MaxHP. Assert `BaseHPDelta` + `CoinDelta`.
3. **Repair** — damaged base. Assert `BaseHPDelta` + `CoinDelta`.
4. **Contested new base** — add a rival faction with an asset on "Krylos". Use `fixedRoller` so the rival ties/beats the expanding faction roll. Assert `ConfirmRivalFreeAttack` is called and (if confirmed) the rival's base-attack mutations appear.

---

## Phase 2: Sub-Engine Unit Tests

**Deliverable:** `_test.go` in `engine/goal/` and `engine/ability/`.

### `engine/goal/lock_test.go`

Tests `ComputeLock(faction, factionState, rulebook)` for each lock outcome.

| Scenario | Setup | Expected `LockType` |
|---|---|---|
| No active goal | `faction.ActiveGoal = nil` | `LockNone` |
| Change Homeworld in transit | `GoalID: "G-012"`, `TurnsRemaining: 2` | `LockSkip` |
| Change Homeworld completing | `TurnsRemaining: 1` | `LockComplete` (ticks to 0) |
| Planetary Seizure phase 1 | `GoalID: "G-004"`, `ProcessPhase: 1` | `LockRestrictActions` (only Attack allowed) |

Assert that `LockSkip` results include a `GoalTurnsTick` mutation. Assert `LockRestrictActions` returns the correct permitted-action list.

### `engine/goal/progress_test.go`

Tests the progress functions directly using a constructed mutation list.

| Goal | Trigger | Expected mutations |
|---|---|---|
| MilitaryConquest | kill Force asset belonging to rival | `GoalProgressed` delta = kill count |
| MilitaryConquest (complete) | kills push progress ≥ Force rating | `GoalProgressed` + `GoalCompleted` + `CoinDelta` (XP reward) |
| CommercialExpansion | kill Wealth asset | same pattern |
| IntelligenceCoup | kill Cunning asset | same pattern |
| No relevant kills | mutation list has no kills matching goal | no mutations returned |

Use synthetic `domain.Mutation` lists (AssetRemoved with matching asset definitions) rather than running a full action.

### `engine/ability/ability_test.go`

Tests the `AbilityEngine.Resolve()` for the two main ability effect types.

- **Move asset** — pre-load an asset on "Tartarus"; ability moves it to "Krylos". Assert `AssetLocationChanged` mutation.
- **Faction test** — ability targets a rival faction. Use `fixedRoller` to control the test roll outcome. Assert the correct stat-damage mutation fires on success, none on failure.
- **ConfirmAbilityApplied = false** — assert no mutations emitted.

---

## Phase 3: Integration Scenarios

**Deliverable:** Additional test functions in `internal/faction/engine/orchestrator_test.go`.

All tests use the existing `newHarness(t)` / `addFaction(...)` / `ScriptedCollector` pattern.

### New test functions

#### `TestRunCycle_BuyAsset_ReadyNextCycle`

- Faction picks Buy Asset (scripted `SelectBuyOrderFn` → "F1-001" on homeworld).
- After `RunCycle`, assert new asset is in faction's asset list with `Ready = false`.
- Run a second `RunCycle` (faction takes any action). Assert asset's `Ready` flips to `true` at turn start (the `Turn.Start` ready-marking step).

#### `TestRunCycle_ExpandInfluence_NewBase`

- Faction has coin; no existing base on "Krylos". `SelectExpandInfluenceOrderFn` returns `Mode: ModeNewBase, World: "Krylos"`.
- Assert history contains `BaseAdded` mutation for "Krylos".
- Assert faction's Bases slice includes the new entry.

#### `TestRunCycle_ExpandInfluence_Contested`

- Add rival faction with a Force asset on "Krylos". Use `fixedRoller` so rival ties the expansion roll.
- Set `ConfirmRivalFreeAttackFn` → `true`.
- `SelectBaseAttackersFn` → return rival's asset.
- Assert history contains both `BaseAdded` and an attack mutation against the new base.

#### `TestRunCycle_GoalCompleted_LockComplete`

- Faction has `ActiveGoal` for MilitaryConquest near completion threshold.
- Scripted to Attack the rival and kill its asset (use `fixedRoller` for guaranteed kill).
- After `RunCycle`, assert `GoalCompleted` mutation in history.
- Assert `faction.ActiveGoal == nil` in final state.
- Assert `CoinDelta` for XP reward in history.

#### `TestRunCycle_Bookkeeping_AssetDestroyedSecondMiss`

- Faction with one asset; `Coin = 0` (can't pay maintenance).
- Run `RunCycle` → asset marked `Maintained = false`.
- Run second `RunCycle` → second consecutive miss → `AssetRemoved` mutation fires.
- Assert asset is gone from faction's asset list after cycle 2.

#### `TestRunCycle_Bribe`

- Rival faction has a base on a shared world. Attacker has enough coin.
- `SelectBribeTargetFn` returns the rival's base and coin amount.
- Assert history has `BaseHPDelta` on rival's base + `CoinDelta` on attacker.

---

## Verification

After each phase, run from repo root:

```bash
go test ./internal/faction/...
```

All packages must pass with zero failures. Use `-v` to confirm individual test names are running. No new `[no test files]` packages should remain in `engine/actions/`, `engine/goal/`, or `engine/ability/` after phase 2.

After phase 3:

```bash
go test ./internal/faction/engine/ -v -run TestRunCycle
```

Should list all 6 orchestrator scenarios passing.

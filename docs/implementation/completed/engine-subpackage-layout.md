# Engine Sub-Package Layout

Design document for splitting `internal/faction/engine` into sub-packages. Produced during the Phase 3 session of `feature/core-engine-orchestrator`.

**Branch:** `feature/core-engine-orchestrator` (Phase 3)

<br/>

## Goal

Make the engine package human-readable and human-maintainable. Each sub-engine becomes its own package. The top-level `engine` package becomes a clean composition root: it owns the orchestrator, the public interfaces, and the `Engine` struct. Nothing more.

<br/>

## Resolved Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | Option A — `Action` interface and `Collector` interfaces live in their sub-packages; top-level `engine` composes them | Keeps sub-packages self-contained; no shared contracts dumping ground; idiomatic Go |
| 2 | `InputCollector` embeds `action.Collector` + `ability.Collector` and adds orchestrator-only methods | Ownership is clear — each sub-package defines the collector contract it requires; `InputCollector` is the full composition; no method duplication |
| 3 | `TurnEngine` keeps its struct shape and gains a `domain.Roller` field | Roller injection gives it real state; consistent struct-based shape across all six sub-engines |
| 4 | All sub-engine constructors named `New(...)` within their package | Uniform: `turn.New(roller)`, `goal.New()`, `mutation.New()`, etc. Top-level `engine.New()` calls each |
| 5 | `NewWithRulebook(*loader.Rulebook)` added to top-level `engine` | Unlocks hermetic tests that don't hit disk |
| 6 | `HistoryEngine.Record` signature changes to take `(historyPath, factionState, faction, mutations)` | `buildEventRecord` becomes an unexported helper inside `engine/history`; orchestrator's `applyAndRecord` shrinks; history package owns its full write responsibility |
| 7 | `Phase*` checkpoint constants renamed to `Checkpoint*` and moved to `observer.go` | Resolves collision with `domain.Phase*` turn-cursor constants; `observer.go` is the right semantic home for pipeline-phase names |
| 8 | `Engine.AbilityEngine` field renamed to `Engine.Ability` | Matches the `Role *RoleEngine` naming pattern of all other fields |
| 9 | `event_record.go` deleted; `buildEventRecord` moves into `engine/history` | Was a 25-line file with one caller; belongs with the package that owns history writes |
| 10 | Order types (`BuyOrder`, `RepairOrder`, `RefitOrder`, `RefitOption`, `ExpandInfluenceOrder`, `ExpandMode`, `ReinforceMode`) move to `engine/action` | Live closest to what uses them; resolves the `engine/action` → `engine` import cycle |

<br/>

## Package Tree

```
internal/faction/
├── config/                     — unchanged
│   └── config.go
│
└── engine/                     — composition root: Engine struct, orchestrator, public interfaces
    ├── engine.go               — Engine struct, New(), NewWithRulebook()     [rename from core.go]
    ├── orchestrator.go         — RunFactionTurn, RunCycle, applyAndRecord,
    │                             finishFactionTurn, filterAllowedActions
    ├── input_collector.go      — InputCollector (embeds action.Collector +
    │                             ability.Collector; adds SelectAction, AwaitCheckpoint)
    ├── observer.go             — TurnObserver interface, Checkpoint* constants
    ├── event_hook.go           — EventHook interface (unchanged)
    ├── roller.go               — RandRoller (unchanged)
    ├── orchestrator_test.go    — integration tests (unchanged)
    │
    ├── turn/
    │   ├── turn.go             — TurnEngine struct, New(roller), Start, InProgress,
    │   │                         CurrentFaction, Advance, Abandon
    │   ├── bookkeeping.go      — ApplyBookkeeping, applyMaintenance, maintenanceCost,
    │   │                         BookkeepingResult, AssetRef, readyAllAssets
    │   ├── order.go            — buildFactionOrder
    │   └── turn_test.go        — moved from engine/turn_test.go
    │
    ├── goal/
    │   ├── goal.go             — GoalEngine struct, New(), CheckLock, UpdateProgress
    │   ├── lock.go             — GoalLock, LockType (LockNone/LockSkip/LockRestrictActions),
    │   │                         checkLockChangeHomeworld, checkLockPlanetarySeizure
    │   ├── progress.go         — all progressXxx handlers (11 goal types)
    │   └── helpers.go          — completeGoal, countAssetKillsByCategory, findAsset,
    │                             factionHasUnstealthedAssetOn, factionHasBaseOn,
    │                             worldHasRivalPresence, rivalHasPlanetaryGovernmentOnWorld,
    │                             factionHasPlanetaryGovernmentTag, calcPlanetarySeizureXP
    │
    ├── mutation/
    │   ├── mutation.go         — MutationEngine struct, New(), Apply
    │   └── mutation_test.go    — moved from engine/mutation_test.go
    │
    ├── history/
    │   ├── history.go          — HistoryEngine struct, New(), Record(path, factionState,
    │   │                         faction, mutations) — signature change from decision #6
    │   ├── record.go           — buildEventRecord (unexported; moved from event_record.go)
    │   └── history_test.go     — moved from engine/history_engine_test.go
    │
    ├── action/
    │   ├── action.go           — Action interface, ActionEngine struct, New(),
    │   │                         Register, AvailableActions, Run
    │   │                         ActionFactory type: func(Collector) Action
    │   └── collector.go        — action.Collector interface (all 14 action-input methods),
    │                             order types: BuyOrder, RepairOrder, RefitOrder, RefitOption,
    │                             ExpandInfluenceOrder, ExpandMode, ReinforceMode
    │
    ├── ability/
    │   ├── ability.go          — AbilityEngine struct, New(), Run, RegisterCustomHandler,
    │   │                         movementStepHandler, factionTestStepHandler,
    │   │                         factionTestCandidates, applyAbilityEffect,
    │   │                         worldsFromState, abilityStatScore
    │   │                         StepHandler and CustomAbilityHandler types
    │   └── collector.go        — ability.Collector interface
    │                             (SelectMoveDestination, SelectFactionTestTarget)
    │
    ├── actions/                — unchanged location; imports updated
    │   ├── register.go         — RegisterDefaultActions(*engine.Engine)
    │   ├── attack.go
    │   ├── attack_test.go
    │   ├── buy_asset.go
    │   ├── sell_asset.go
    │   ├── repair_asset.go
    │   ├── repair_faction.go
    │   ├── refit_asset.go
    │   ├── expand_influence.go
    │   ├── bribe.go
    │   ├── seize_planet.go
    │   ├── abandon_goal.go
    │   ├── use_asset_ability.go
    │   ├── use_asset_ability_test.go
    │   └── eligibility.go
    │
    └── testharness/            — unchanged location; imports updated
        ├── collector.go
        └── observer.go
```

<br/>

## Import Graph

Arrows show `imports`. No cycles.

```
domain, loader, state, config
    ▲  ▲  ▲  ▲  ▲  ▲  ▲
    │  │  │  │  │  │  │
  turn goal mut hist act abil
    │              │    │
    └──────────────┴────┘
              ▲
              │  (also imports turn, goal, mut, hist, act, abil)
           engine   ◄── engine/actions (register only; engine never imports actions)
              ▲
           engine/testharness
           engine/orchestrator_test.go
```

Detailed:

| Package | Imports |
|---------|---------|
| `engine/turn` | `domain`, `state` |
| `engine/goal` | `domain`, `loader`, `state` |
| `engine/mutation` | `domain`, `state` |
| `engine/history` | `domain`, `state` |
| `engine/action` | `domain`, `loader`, `state` — order types defined here, no upward import |
| `engine/ability` | `domain`, `loader`, `state` |
| `engine/actions` | `engine` (InputCollector), `engine/action` (Action, Collector, order types), `engine/ability` (AbilityEngine), `domain`, `loader`, `state` |
| `engine` (top-level) | `engine/turn`, `engine/goal`, `engine/mutation`, `engine/history`, `engine/action`, `engine/ability`, `domain`, `loader`, `state`, `config` |
| `engine/testharness` | `engine`, `engine/action` (Action), `engine/goal` (GoalLock, LockType), `engine/turn` (BookkeepingResult), `domain`, `loader`, `state` |

<br/>

## Key Interface Definitions After Restructure

### `action.Collector` (`engine/action/collector.go`)

All methods an action implementation may call on the collector. `InputCollector` embeds this. Order types live here too (decision #10) — closest to what uses them.

```go
type Collector interface {
    SelectAsset(assets []*domain.Asset, rulebook *loader.Rulebook) (*domain.Asset, error)
    SelectRepairOrders(faction *domain.Faction, damaged []*domain.Asset, rulebook *loader.Rulebook) ([]RepairOrder, error)
    SelectBuyOrder(worlds []string, purchasable []*domain.AssetDefinition) (BuyOrder, error)
    SelectRefitOrder(options []RefitOption, rulebook *loader.Rulebook) (RefitOrder, error)
    SelectAttackers(eligible []*domain.Asset, rulebook *loader.Rulebook) ([]*domain.Asset, error)
    SelectDefender(attacker *domain.Asset, eligible []*domain.Asset, rulebook *loader.Rulebook) (*domain.Asset, error)
    ConfirmRedirectToBase(defenderFaction *domain.Faction, base *domain.Base, damage int) (bool, error)
    SelectExpandInfluenceOrder(faction *domain.Faction, factionState *state.FactionState) (ExpandInfluenceOrder, error)
    ConfirmRivalFreeAttack(rival *domain.Faction, rivalRoll, factionRoll int) (bool, error)
    SelectBaseAttackers(rival *domain.Faction, eligible []*domain.Asset, rulebook *loader.Rulebook) ([]*domain.Asset, error)
    SelectAbilityAssets(faction *domain.Faction, candidates []*domain.Asset, rulebook *loader.Rulebook) ([]*domain.Asset, error)
    ConfirmAbilityApplied(asset *domain.Asset, def *domain.AssetDefinition) (bool, error)
    SelectBribeTarget(faction *domain.Faction, factionState *state.FactionState) (*domain.Base, int, error)
    SelectSeizeTarget(faction *domain.Faction, factionState *state.FactionState) (string, error)
}

// Order types — live here because they are inputs to actions; defined closest to their use site.
type RepairOrder struct { ... }
type BuyOrder struct { ... }
type RefitOption struct { ... }
type RefitOrder struct { ... }
type ExpandInfluenceOrder struct { ... }
type ExpandMode string
type ReinforceMode string
```

### `ability.Collector` (`engine/ability/collector.go`)

```go
type Collector interface {
    SelectMoveDestination(asset *domain.Asset, worlds []string) (string, error)
    SelectFactionTestTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Faction) (*domain.Faction, error)
}
```

### `engine.InputCollector` (`engine/input_collector.go`)

```go
type InputCollector interface {
    action.Collector
    ability.Collector
    SelectAction(faction *domain.Faction, available []action.Action) (action.Action, error)
    AwaitCheckpoint(phase string) error
}
```

### `Engine` struct after restructure (`engine/engine.go`)

```go
type Engine struct {
    Rulebook *loader.Rulebook
    Rand     domain.Roller
    Turn     *turn.TurnEngine
    Mutation *mutation.MutationEngine
    Action   *action.ActionEngine
    Ability  *ability.AbilityEngine
    History  *history.HistoryEngine
    Goal     *goal.GoalEngine
}

func New(dataDir string) (*Engine, error) {
    rb, err := loader.Load(dataDir)
    if err != nil {
        return nil, err
    }
    return NewWithRulebook(rb)
}

func NewWithRulebook(rb *loader.Rulebook) *Engine {
    e := &Engine{Rulebook: rb, Rand: NewRandRoller()}
    e.Turn    = turn.New(e.Rand)
    e.Mutation = mutation.New()
    e.Action  = action.New()
    e.Ability = ability.New()
    e.History = history.New()
    e.Goal    = goal.New()
    return e
}
```

<br/>

## Open Questions

None — all design questions resolved.

<br/>

## Verification

After each sub-package extraction:

```bash
go build ./...
go vet ./...
go test ./...
```

The three orchestrator integration tests (`TestRunCycle_*`) are the load-bearing safety net — they must stay green after every commit.

<br/>

## Commit Pattern

One commit per sub-package extraction, in dependency order (lowest first):

```
refactor: extract engine/mutation sub-package
refactor: extract engine/history sub-package
refactor: extract engine/turn sub-package
refactor: extract engine/goal sub-package
refactor: extract engine/ability sub-package
refactor: extract engine/action sub-package
refactor: update engine/actions and engine/testharness imports
refactor: collapse engine top-level to composition root
```

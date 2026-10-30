# Engine Collector Reshape — Implementation Plan

## Context / Goal

Drop `engine.InputCollector` entirely and replace it with two purpose-scoped types:

- `PhaseCollector` — 4 methods owned by the orchestrator
- `Collectors` struct — bundles `PhaseCollector` + `action.Collector`

The orchestrator receives `Collectors` instead of the flat `InputCollector`. Sub-engines keep their existing narrow contracts unchanged.

- Discovery doc: [`engine-collector-reshape-discovery.md`](../discovery/engine-collector-reshape-discovery.md)

## Decisions Ratified in Planning

1. **`Collectors.Phase` field name** — `Phase`. Reads naturally alongside `Action`; "phase-level orchestrator prompts" is an accurate description.

2. **Mock fate** — Regenerate against `action.Collector`. Mock surface shrinks (drops 4 orchestrator-only methods). All action test files using `MockInputCollector` → `MockCollector` (one mechanical rename).

3. **`hooks.Collector` embed on `action.Collector` is load-bearing — do not remove.** The discovery doc stated the embed was "verified vestigial." Re-grounding found this is incorrect: `attack.go` passes its `action.Collector` to `dispatch.RollWithHooks`, which calls `SelectModifiers` and `ConfirmReroll`. No action calls those methods *directly*, but they are called indirectly through the pass-through. Removing the embed would break the attack action. Removing it would also cross into the ActionFactory pattern change, which is explicitly out of scope.

4. **`Collectors` struct has 2 fields, not 4.** The discovery doc's 4-field design (`Phase`, `Action`, `Hooks`, `Ability`) was premised on removing the `hooks.Collector` embed. Since the embed stays, `collectors.Action` already transitively satisfies `hooks.Collector` and `ability.Collector`. Separate `Hooks` and `Ability` fields would be false explicitness. Final shape:

   ```go
   type Collectors struct {
       Phase  PhaseCollector
       Action action.Collector  // satisfies hooks.Collector + ability.Collector transitively
   }
   ```

5. **`Harness.Collectors` field added to `testharness.Harness`** — initialized in `NewHarness` as `engine.Collectors{Phase: scriptedCollector, Action: scriptedCollector}`. Scenario tests pass `h.Collectors` to `RunCycle`/`RunFactionTurn` instead of building the struct inline at each call site.

---

## Out of Scope

- `steps.Collector` consolidation — owned by R-001
- `ActionFactory(Collector)`-stored-as-field pattern change
- `ability.Collector` embed on `action.Collector` — load-bearing, unchanged
- `hooks.Collector` embed removal from `action.Collector` — load-bearing (see Decision 3)
- Ability-handler shape changes

---

## Work Breakdown

Single commit: `refactor: replace InputCollector with PhaseCollector + Collectors`

### Step 1 — `internal/faction/engine/collector.go`

- Delete the `InputCollector` interface and all its imports.
- Delete the `//go:generate` directive (moves to `action/collector.go`).
- Add `PhaseCollector` interface:

  ```go
  type PhaseCollector interface {
      AwaitCheckpoint(phase string) error
      SelectAction(faction *domain.Faction, available []action.Action) (action.Action, error)
      SelectStatRaise(faction *domain.Faction, eligible []domain.FactionStat) (*domain.FactionStat, error)
      SelectMovementDecisions(faction *domain.Faction, eligible []*domain.Asset) ([]world.MovementDecision, error)
  }
  ```

- Add `Collectors` struct:

  ```go
  type Collectors struct {
      Phase  PhaseCollector
      Action action.Collector
  }
  ```

- Update the package-overview comment in `core.go` (the `collector.go` bullet) so it names `PhaseCollector` + `Collectors` instead of `InputCollector`.
- Update the orchestrator file-level doc comment near `orchestrator.go:18` — it currently reads "for caller acknowledgement via InputCollector.AwaitCheckpoint." Change `InputCollector` to `PhaseCollector` (the checkpoint method lives on `PhaseCollector` after the refactor).
- Remove the now-unused `hooks` import from `engine/collector.go` once `SelectModifiers` / `ConfirmReroll` are gone (the `world` import stays — `PhaseCollector.SelectMovementDecisions` still needs it).

### Step 2 — `internal/faction/engine/action/collector.go`

- Add `//go:generate` directive targeting `action.Collector`:

  ```go
  //go:generate mockgen -destination=actions/mocks/mock_collector.go -package=mocks github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action Collector
  ```

- No interface changes. Both embeds (`hooks.Collector`, `ability.Collector`) remain.

### Step 3 — `internal/faction/engine/orchestrator.go`

Replace `collector InputCollector` with `collectors Collectors` across all method signatures:

- `RunCycle`, `RunFactionTurn`, `runGoalLockPhase`, `runBookkeepingPhase`, `runStatRaisePhase`, `runMovementPhase`, `runActionPhase`, `finishFactionTurn`

Thread appropriately:

- `collectors.Phase` for all `AwaitCheckpoint`, `SelectAction`, `SelectStatRaise`, `SelectMovementDecisions` calls.
- `collectors.Action` for the `e.Action.AvailableActions(..., collectors.Action)` call.

Update package-level helpers:

- `prepareStatRaise(faction *domain.Faction, collector PhaseCollector)` — change parameter type from `InputCollector` to `PhaseCollector`.
- `prepareMovementDecisions(faction *domain.Faction, collector PhaseCollector, ...)` — same.

### Step 4 — `internal/faction/engine/testharness/harness.go`

- Add `Collectors engine.Collectors` field to `Harness`. **Keep the existing `Collector *ScriptedCollector` field** — it stays as the scripting handle tests use to set `SelectXxxFn` overrides before calling `RunCycle`. `Collectors` is the dispatch wrapper consumed by the orchestrator; both must point at the *same* `*ScriptedCollector` instance or test scripts will be silently dropped.
- In `NewHarness`, currently the `ScriptedCollector` is constructed inline inside the struct literal (`harness.go:83`). Hoist it to a named local *before* the `return` so the same instance can be shared by both fields:

  ```go
  scriptedCollector := &ScriptedCollector{}
  return &Harness{
      Engine:       eng,
      FactionState: &state.FactionState{...},
      Cfg:          cfg,
      Collector:    scriptedCollector,
      Collectors:   engine.Collectors{Phase: scriptedCollector, Action: scriptedCollector},
      Observer:     &RecordingObserver{},
  }
  ```

  Do not write `&ScriptedCollector{}` twice — that would create two distinct instances and any `h.Collector.SelectXxxFn = ...` mutation a test makes wouldn't be visible through `h.Collectors`.

### Step 5 — testharness doc comments

- `internal/faction/engine/testharness/collector.go`: update the `ScriptedCollector` doc comment — remove the reference to `InputCollector`; describe the four interfaces it structurally satisfies (`PhaseCollector`, `action.Collector`, `hooks.Collector`, `ability.Collector`).
- `internal/faction/engine/testharness/observer.go:1`: the package doc currently says "Package testharness provides headless implementations of InputCollector and TurnObserver…". Update to reference `PhaseCollector` + `action.Collector` (or just "the engine's collector interfaces") instead of the dead type name.

### Step 6 — Regenerate mock

Run from the repo root:

```
go generate ./internal/faction/engine/action/...
```

This regenerates `internal/faction/engine/action/actions/mocks/mock_collector.go` against `action.Collector`. The generated type becomes `MockCollector` (was `MockInputCollector`). The mock surface shrinks by 4 methods (the orchestrator-owned ones: `SelectAction`, `SelectStatRaise`, `SelectMovementDecisions`, `AwaitCheckpoint`).

### Step 7 — Update action test files

In all files under `internal/faction/engine/action/actions/`:

- `mocks.NewMockInputCollector(ctrl)` → `mocks.NewMockCollector(ctrl)`
- `*mocks.MockInputCollector` → `*mocks.MockCollector`

Files affected: `attack_test.go`, `use_asset_ability_test.go`, `repair_asset_test.go`, `refit_asset_test.go`, `buy_asset_test.go`, `seize_planet_test.go`, `expand_influence_test.go`, `simple_actions_test.go`.

### Step 8 — Update scenario test files

In all files under `internal/faction/engine/testharness/scenarios/`:

- `h.Engine.RunCycle(h.FactionState, h.Cfg, h.Collector, h.Observer)` → `h.Engine.RunCycle(h.FactionState, h.Cfg, h.Collectors, h.Observer)`

Files affected: `full_cycle_test.go`, `hooks_test.go`, `goal_lock_test.go`, `actions_test.go`, `tags_test.go`, `stat_raise_test.go`.

### Verification

After all changes:

- `go build ./...` — must pass clean.
- `go vet ./...` — must pass clean. Catches unused-imports (e.g. the `hooks` import removed from `engine/collector.go` in Step 1) and any shadowed-var artifacts from the orchestrator parameter rename.
- `go test ./...` — must pass with no skips.
- Spot-check the regenerated `internal/faction/engine/action/actions/mocks/mock_collector.go`:
  - Source comment on line 2 should now read `Source: …/action (interfaces: Collector)` (was `…/engine (interfaces: InputCollector)`).
  - Type renamed: `MockCollector` / `MockCollectorMockRecorder` (was `MockInputCollector` / `MockInputCollectorMockRecorder`).
  - The four orchestrator-only methods (`AwaitCheckpoint`, `SelectAction`, `SelectStatRaise`, `SelectMovementDecisions`) should be absent — confirms Decision 2's mock-surface shrink actually happened.

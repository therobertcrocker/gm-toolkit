# Engine Collector Reshape — Discovery

## Problem

The `input-collector-refactor` (commit 5cfa31f) flattened `engine.InputCollector` by replacing embedded sub-package interfaces with explicit method declarations. The flatten produced a clearer engine-boundary contract but doubled the maintenance points for any new prompt that flows through a sub-engine.

Adding a single input method now touches 2–3 places: the narrow sub-package interface (e.g. `ability.Collector`) that consumes it, the flat `engine.InputCollector` that re-declares it, and — for ability inputs — the `ability/steps/Collector` duplicate. The pre-refactor megainterface was assembled by embedding; the post-refactor megainterface is assembled by hand. Both shapes share the same root cause: the engine boundary attempts to act as a contract document enumerating every input method, when Go's structural typing means each sub-package's narrow interface is already a sufficient contract for its own consumer.

Secondary smells surface alongside the duplication:

- `action.Collector` still embeds `hooks.Collector` — verified vestigial. No production action implementation calls `SelectModifiers` or `ConfirmReroll`; the hooks dispatcher invokes those itself via `e.Hooks`. Only test mocks reference them through this embed.
- `action.Collector` embeds `ability.Collector` for a real reason: `use_asset_ability.go:77` forwards its collector to `abilityEngine.Run`, which expects an `ability.Collector`. The embed is load-bearing.
- The flatten was applied at the engine layer but not propagated downward. The engine boundary is flat; the action boundary is layered. There is no principle distinguishing the two — it is simply where the last refactor stopped.

## Design Summary

Drop `engine.InputCollector` entirely. The orchestrator declares only what it itself calls — a small `PhaseCollector` interface with four methods (`SelectAction`, `SelectStatRaise`, `SelectMovementDecisions`, `AwaitCheckpoint`). Sub-engines keep their existing narrow contracts unchanged in shape, except for removing the vestigial `hooks.Collector` embed on `action.Collector`.

The orchestrator receives a `Collectors` struct bundling four narrow types:

```go
type Collectors struct {
    Phase   PhaseCollector
    Action  action.Collector   // 14 methods + ability.Collector embed (unchanged)
    Hooks   hooks.Collector
    Ability ability.Collector
}

func (e *Engine) RunFactionTurn(
    factionState *state.FactionState,
    cfg *config.Config,
    collectors Collectors,
    observer TurnObserver,
) (bool, error)
```

The concrete caller (test harness today, GM TUI later) provides one type that structurally satisfies all four narrow interfaces. Composition happens at call site:

```go
gm := &TUIPrompter{...}
collectors := Collectors{Phase: gm, Action: gm, Hooks: gm, Ability: gm}
```

No method is declared in more than one place. A future prompt belongs in exactly one of the four interfaces — the one whose consumer calls it.

## Audit Findings

### Interface inventory (post-flatten, pre-reshape)

| Layer | Interface | Methods | Embeds | Real consumer |
|---|---|---|---|---|
| Engine | `engine.InputCollector` | ~22, flat | none | orchestrator (4 methods) + propagation only |
| Action | `action.Collector` | 14 explicit | `hooks.Collector`, `ability.Collector` | action implementations |
| Hooks | `hooks.Collector` | 2 | none | `dispatch/` package |
| Ability | `ability.Collector` | 2 | none | `ability.Run` / `runStep` |
| Steps | `ability/steps/Collector` | 2 | none | step handlers (`steps.Movement`, `steps.FactionCheck`) |

### Method-by-method ownership

Of the 22 methods on `engine.InputCollector`:

- 4 are orchestrator-owned (`SelectAction`, `SelectStatRaise`, `SelectMovementDecisions`, `AwaitCheckpoint`) — used only in `orchestrator.go`.
- 14 are action-owned — re-declared on `action.Collector`.
- 2 are hooks-owned — re-declared on `hooks.Collector`.
- 2 are ability-owned — re-declared on `ability.Collector` *and* on `steps.Collector`.

Every non-orchestrator method has at least two declarations. Ability methods have three.

### Embedded interface usage

| Embed | In | Real use? |
|---|---|---|
| `hooks.Collector` in `action.Collector` | ✗ Vestigial. No production caller; only test mocks reference. |
| `ability.Collector` in `action.Collector` | ✓ Load-bearing. `use_asset_ability.go:77` forwards collector to `abilityEngine.Run`. |

## Per-Area Design Details

### `engine/collector.go`

- Delete the `InputCollector` interface.
- Add `PhaseCollector` (4 methods listed above).
- Add `Collectors` struct (four typed fields).
- The `//go:generate mockgen` directive for `InputCollector` is dropped or replaced with directives for each narrow interface that still needs a mock. Mocks for `action.Collector` already live under `action/actions/mocks/`.

### `engine/orchestrator.go`

- Replace every `collector InputCollector` parameter with `collectors Collectors`.
- Each phase method uses the appropriate field:
  - `runMovementPhase`, `runStatRaisePhase`, finish-turn checkpoint calls → `collectors.Phase`
  - `runActionPhase` → forwards `collectors.Action` into `e.Action.AvailableActions` and uses `collectors.Phase.SelectAction`
- `prepareStatRaise` and `prepareMovementDecisions` take `PhaseCollector` rather than `InputCollector`.

### `engine/action/collector.go`

- Remove `hooks.Collector` embed.
- Keep `ability.Collector` embed (load-bearing).
- Surface method count unchanged from caller's perspective.

### `engine/testharness/collector.go`

- `ScriptedCollector` retains all current methods — it structurally satisfies `PhaseCollector + action.Collector + hooks.Collector` (and `ability.Collector` transitively via the action embed).
- No struct-shape changes; only the interfaces it advertises change.

### `engine/action/actions/mocks/mock_collector.go`

- Regenerate against `action.Collector` (the narrow interface) instead of `engine.InputCollector`. The mock surface shrinks — orchestrator-only methods (`SelectAction`, `SelectStatRaise`, `SelectMovementDecisions`, `AwaitCheckpoint`) drop from the mock.

## Out of Scope

- **`steps.Collector` consolidation.** The duplicate of `ability.Collector` in `ability/steps/` is owned by R-001 (Ability Engine Redesign). That initiative's discovery decides the shape of the `ability/steps/` package; the duplication is one of the things its findings will address. Per the deferred-refactor discipline (`session-modes.md` section 5 and 8), this refactor does not pre-empt R-001's design.
- **`ActionFactory(Collector)`-stored-as-field pattern.** Each action holding a `Collector` field at construction is unchanged. The aggregate shape of `action.Collector` (single field passing for ability propagation) remains a readability win at the action layer.
- **Ability-handler shape change.** Whether ability step handlers should be Shape 2 self-registering with stored collectors is R-001's domain.
- **`ability.Collector` embed on `action.Collector`.** Removing it would require restructuring how `use_asset_ability` invokes the ability engine. Kept as-is.

## Replaces / Retires

- `engine.InputCollector` interface — deleted.
- The bulk of `engine/collector.go` — replaced with `PhaseCollector` + `Collectors`.
- The `hooks.Collector` embed in `action.Collector` — removed.
- The `mockgen` directive targeting `engine.InputCollector` — replaced or removed.

## Open Questions

- **Naming of the `Phase` field on `Collectors`.** `Phase` reads well alongside `Action` / `Hooks` / `Ability`, but slightly conflates "phase-level orchestrator prompts" with the phase pipeline itself. Alternatives: `Turn`, `Orchestrator`, `Top`. Plan session decides.
- **Engine-level mock necessity.** Today's `mock_collector.go` is consumed by action tests, not engine tests. After the reshape, action tests can mock `action.Collector` directly. Plan session decides whether to delete the engine-level mock entirely or regenerate against `action.Collector`.

## Reference Exemplars

- `docs/initiatives/implementation/input-collector-refactor-plan.md` — prior in-line refactor that established the flat `InputCollector` this work supersedes.
- R-001 (Ability Engine Redesign) entry in `planned-work.md` Up Next — describes the deferred work that owns `steps.Collector` consolidation and ability-handler shape change.

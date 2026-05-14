# Sub-Engine Self-Bootstrap — Implementation Plan

> **Supersedes:** [`sub-engine-alignment-plan.md`](sub-engine-alignment-plan.md) (do not execute that one)
> **Discovery:** [`../discovery/sub-engine-alignment-discovery.md`](../discovery/sub-engine-alignment-discovery.md)

Brings the sub-engines under `internal/faction/engine/` into alignment with a **revised shape catalog**, where each sub-engine (except `action`) self-bootstraps its default handlers from its sibling package — eliminating engine-passing entirely from the tag, goal, and ability trees.

<br/>

## Why a new plan

The earlier plan ([`sub-engine-alignment-plan.md`](sub-engine-alignment-plan.md)) flipped the dependency direction inside `tag/tag.go` but kept a `tag/tags/register.go` bootstrap that still imported `engine` and took `*engine.Engine`. That **concentrated** the dep inversion into one designated file per sub-engine; it didn't **eliminate** it. The convention "only `foo/foos/register.go` may import engine" is a confession that the import is still happening.

It also collapsed two genuinely different patterns into one bucket ("Shape 2 open-ended"): `action` (factory closures over heterogeneous engine fields) and `hooks` (runtime-populated registry written to by other systems during play). Those are different shapes that should not share a label.

This plan replaces that approach:

1. Each non-`action` sub-engine's `New()` imports its sibling handler package and registers defaults inline. Nothing in `tag/`, `tag/tags/`, `goal/`, `goal/goals/`, `ability/`, or `ability/steps/` ever imports the parent `engine` package.
2. The composition root (`core.go`) becomes pure assembly — one line per sub-engine, no paired `RegisterDefault*` call.
3. The shape catalog is revised from three shapes to five, reflecting honest distinctions the original catalog flattened.
4. `action` is documented as the genuine exception: its factories close over heterogeneous engine fields, so its bootstrap remains in a sibling `actions/register.go` that takes `*engine.Engine`.

<br/>

## Revised Shape Catalog

The discovery's three-shape catalog over-fit two patterns into "Shape 2 open-ended" and missed that `ability`'s two dispatch tables have different shapes. Replacing it with five named shapes that label **dispatch tables**, not whole sub-engines (a struct may carry more than one):

| # | Shape | When to use | Example |
|---|---|---|---|
| 1 | **Closed dispatcher** | Set of cases is closed by a domain enum or type switch. Implement as a `switch`, no registry, no registration. | `mutation`, ability *step* dispatch |
| 2 | **Self-bootstrapping registry** | Set of entries is open (rulebook-defined) but defaults are known at compile time. `New()` registers defaults; `Register(...)` available for extensions; lookup site documents the data-only fall-through. | `tag`, `goal`, ability *custom* dispatch |
| 3 | **Factory registry with engine deps** | Handlers have heterogeneous engine dependencies and need a factory closure. Bootstrap function lives in sibling package and takes `*engine.Engine` because no single sub-engine reference suffices. | `action` |
| 4 | **Runtime-populated scoped registry** | Registry exposed for other systems to push into during play. No startup wiring. Dispatch logic lives in a sibling `dispatch/` package; concrete hooks are registered from elsewhere (tags, goals, abilities) when game state demands. | `hooks` |
| 5 | **State-owning subsystem** | Constructor takes deps, may return errors, owns lifecycle state. | `turn`, `world` |

**Mixed-shape sub-engines are legitimate.** `ability` carries one Shape 1 dispatch table (step types, keyed by `domain.AbilityStepType`) and one Shape 2 dispatch table (custom asset handlers, keyed by rulebook `defID`). Each gets the shape that fits its key universe.

**The `foo/foos/` sibling pattern** is retained for Shape 2 sub-engines because it gives each rulebook concept its own file and keeps the parent package focused on engine mechanics. The key change from the original plan: the sibling package contains **only handler types**. No `register.go`, no bootstrap function, no `engine` import. `New()` in the parent does the wiring.

<br/>

## Decisions Ratified in Planning

| # | Decision |
|---|---|
| 1 | Each non-`action` sub-engine's `New()` registers its default handlers from its sibling package inline. No `RegisterDefaultX(eng)` functions exist in the tag, goal, or ability trees. |
| 2 | Sibling packages (`tag/tags/`, `goal/goals/`, `ability/steps/`) contain handler types and helpers only — they never import the parent `engine` package. |
| 3 | `core.go` is pure assembly: one line per sub-engine (`e.Tag = tag.New()`, etc.). No paired bootstrap calls. |
| 4 | The test harness drops its `tags.RegisterDefaultTags(eng)` call (and never adds `goals` or `steps` equivalents). It relies on `New()` having done the wiring. |
| 5 | `action` retains its current pattern: `action/actions/register.go` takes `*engine.Engine` because its factory closures legitimately need heterogeneous engine fields (`e.Rand`, `e.Hooks`, `e.World`, `e.Ability`, `worldIndex(e)`). This is documented as the exception, not the template. |
| 6 | `ability`'s `stepHandlers` map collapses to a `switch` on `domain.AbilityStepType`. The map adds no value when the key set is a closed domain enum. The step handler implementations move to `ability/steps/` as functions called from the switch — no registration. |
| 7 | `ability`'s `customHandlers` map stays. Its keys are rulebook `defID` values (open). `RegisterCustomHandler` remains the public API. No defaults are wired today (nothing currently registers a custom handler). |
| 8 | `goal` is promoted to Shape 2 (self-bootstrapping registry). 12 handlers move to `goal/goals/`; the switches in `CheckLock` and `UpdateProgress` are replaced with registry lookup + data-only fall-through. |
| 9 | `history` is demoted into `turn/` as a package-level `turn.RecordHistory(...)` function. The `History` field on `*Engine` is removed; `internal/faction/engine/history/` is deleted. (Same as the prior plan's Effort 4.) |
| 10 | The discovery's three-shape catalog is superseded by the five-shape catalog above. Future sub-engines should fit one of the five shapes; mixed shapes (multiple dispatch tables in one struct) are allowed and expected. |

<br/>

## Current State Assessment

What's already landed on `refactor/sub-engine-alignment` vs. what still needs to change.

### tag (partially landed, needs realignment)

- **Done:** `tag/tag.go` has `Handler` interface, `Register()`, `ApplyAll()`. No `engine` import. ✓
- **Done:** Handler structs (`ScavengersHandler`, `WarlikeHandler`, `FanaticalHandler`, `PreceptorArchiveHandler`) exist in `tag/tags/*.go` and implement `tag.Handler`. ✓
- **Done:** `core.go` has `e.Tag = tag.New()` at line 69. ✓
- **Needs change:** `tag/tags/register.go` exists and imports `engine`. **Delete this file.**
- **Needs change:** `testharness/harness.go:73` calls `tags.RegisterDefaultTags(eng)`. **Remove this call** and the `tags` import on line 16.
- **Needs change:** `tag/tag.go`'s `New()` must self-register the four handlers (importing `tag/tags`).

### ability (not yet refactored)

- **Current:** `ability/ability.go` has `stepHandlers` map *and* `customHandlers` map. `New()` does inline registration of two step handlers at lines 43–44. Step handler bodies and helpers (`movementStepHandler`, `factionTestStepHandler`, `factionTestCandidates`, `applyAbilityEffect`, `worldsFromState`, `abilityStatScore`) live inline in the same file at lines 84–240.
- **Sibling exists:** `ability/steps/` is an empty directory (per `git status`).
- **Target:** Collapse `stepHandlers` to a switch. Move step handler functions to `ability/steps/`. Drop the action-side `c.(engine.InputCollector)` cast (depends on `action.Collector` composing `ability.Collector`).

### goal (not yet refactored)

- **Current:** `goal/goal.go` is an empty `GoalEngine` struct with two hardcoded switches: `CheckLock` (2 cases for G-004, G-012) and `UpdateProgress` (11 cases for G-001 through G-011). Handler functions live in `goal/progress.go` and `goal/lock.go`.
- **Target:** Shape 2 registry. 12 handlers in `goal/goals/` (G-001 through G-012). Lookup-site fall-through replaces the switches. `goal.New()` self-registers all 12.

### history (not yet refactored)

- **Current:** `history.HistoryEngine` is an empty struct (`history/history.go`) holding a single `Record` method. `core.go:40` declares `History *history.HistoryEngine`; `core.go:67` initializes it; `orchestrator.go:253` calls `e.History.Record(...)`.
- **Target:** `turn.RecordHistory(...)` package-level function. Delete `engine/history/` entirely. Remove the `History` field.

### action (no changes)

`action/actions/register.go` stays. It is the **only** file in the engine tree (after this plan) permitted to take `*engine.Engine`. The factory closures at lines 14–23 close over `e.Rand`, `e.Hooks`, `e.World`, `e.Ability`, and `worldIndex(e)` — no single sub-engine reference suffices.

The `c.(engine.InputCollector)` cast at line 20 still needs to go (Effort 2.2 below) but that's a separate fix from the broader bootstrap question.

<br/>

## Effort & Phase Map

| Effort | Phase | Goal | Touches | Behavior change |
|---|---|---|---|---|
| 1 — Tag realignment | 1.1 | `tag.New()` self-registers defaults; delete `tag/tags/register.go`; drop harness call. | `tag/tag.go`, `tag/tags/register.go` (delete), `testharness/harness.go` | None |
| 2 — Ability | 2.1 | Collapse `stepHandlers` map to a `switch`; move step handlers to `ability/steps/` as functions (no registration). | `ability/ability.go`, new `ability/steps/movement.go`, new `ability/steps/faction_check.go` | None |
| 2 — Ability | 2.2 | Compose `ability.Collector` into `action.Collector`; drop `engine.InputCollector` cast. | `action/collector.go`, `action/actions/register.go`, `action/actions/use_asset_ability.go`, `engine/collector.go` | None |
| 3 — Goal | 3.1 | Introduce `goal.Handler` interface and registry; parallel-path dispatch. | `goal/goal.go` | None |
| 3 — Goal | 3.2 | Migrate 12 handlers to `goal/goals/`; `goal.New()` self-registers; delete switches and old helper files. | `goal/`, new `goal/goals/`, no harness change | None |
| 4 — History | 4.1 | Move `Record` to `turn.RecordHistory`; delete `engine/history/`. | `turn/` (new file), `core.go`, `orchestrator.go`, deletes `history/` | None |

Every phase is behavior-preserving. Tests stay green throughout.

<br/>

## Cross-Cutting Notes

### Self-bootstrap convention

After this initiative, Shape 2 sub-engines all follow the same pattern. Example (`tag`):

```go
// internal/faction/engine/tag/tag.go
package tag

import (
    "github.com/.../internal/faction/engine/tag/tags"
    // ... other imports (domain, hooks, state) — never engine
)

func New() *TagEngine {
    e := &TagEngine{handlers: make(map[string]Handler)}
    e.Register(tags.ScavengersHandler{})
    e.Register(tags.WarlikeHandler{})
    e.Register(tags.FanaticalHandler{})
    e.Register(tags.PreceptorArchiveHandler{})
    return e
}
```

Concrete handler files in `tag/tags/*.go` import only `domain`, `state`, `rulebook`, `hooks` — never `engine`, never the parent `tag` package (interface satisfaction is implicit). The single rule is straightforward: **nothing in the tag/goal/ability trees imports `engine`.**

### Data-mirroring lookup-site contract

Shape 2 sub-engines (`tag` already, `goal` after Effort 3, ability `customHandlers` already) must treat "ID exists but no handler registered" as a designed state, not a fall-through bug. Standard form at the lookup site:

```go
handler, ok := e.handlers[id]
if !ok {
    // Data-only entry: known to the rulebook, not implemented in code. Intentional.
    return defaultZeroValue
}
```

The comment is required at every Shape 2 lookup site — this remains the one "no comment unless WHY is non-obvious" exception in this plan.

### No production bootstrap caller yet

`action/actions/RegisterDefaultActions(eng)` is invoked only from `testharness/harness.go`. There is no `cmd/` caller yet. After this initiative, that situation is unchanged: action remains the only bootstrap function, and only the harness calls it. When a future CLI binary needs the engine, it will call `actions.RegisterDefaultActions(eng)` after `engine.New(cfg)` — and nothing else, because the other sub-engines have self-wired.

### Why action stays the exception

Action's factory closures (`action/actions/register.go:11–23`) close over heterogeneous engine fields:

```go
e.Action.Register(func(c action.Collector) action.Action { return NewBuyAsset(c, e.Hooks, e.World) })
e.Action.Register(func(c action.Collector) action.Action { return NewAttack(c, e.Rand, e.Hooks, worldIndex(e)) })
// ...eleven actions, each closing over a different subset of {Rand, Hooks, World, Ability, World.Index}
```

There is no single sub-engine reference you could pass instead of `*engine.Engine` here — different factories need different combinations of `Rand`, `Hooks`, `World`, `Ability`, and `World.Index`. Putting this inside `action.New()` would require `action` to know about hooks, world, ability, and rand — which would push the dep tangle into the action sub-engine itself, the opposite of what we want.

Action is therefore the documented exception. Its bootstrap stays in `actions/register.go` and takes `*engine.Engine`.

### Test harness simplification

After all efforts merge, `testharness/harness.go`'s engine setup collapses from this:

```go
eng, err := engine.New(cfg)
// ...
actions.RegisterDefaultActions(eng)
tags.RegisterDefaultTags(eng)
```

to this:

```go
eng, err := engine.New(cfg)
// ...
actions.RegisterDefaultActions(eng)
```

One call. The `tags` import is dropped. No `goals` or `steps` imports are added. Tag, goal, and ability defaults are wired automatically by their respective `New()` constructors.

### Pre-merge checklist (per CLAUDE.md)

Before merging the branch carrying any effort:

1. Self-review as a senior reviewing a junior. Be critical, constructive.
2. Update `docs/dev_journals/faction-manager/dev-journal-factions.md` with effort summary + notes.
3. Update `docs/dev_journals/faction-manager/planned-work.md` (remove completed items, add anything that surfaced).
4. Update `docs/dev_journals/faction-manager/decisions-log.md` only for decisions ratified *during execution*. The ten decisions above belong in this plan, not the log.
5. Assess version bump (patch/minor/major). Tag after merge.

### Model selection

| Phase | Recommended model | Notes |
|---|---|---|
| 1.1 | Sonnet | Small mechanical refactor: move four `Register` calls, delete one file, delete one harness call. |
| 2.1 | Sonnet | Mechanical extraction; switch is straightforward. |
| 2.2 | Sonnet | Three-line interface change. |
| 3.1 | Sonnet | Interface + parallel path. |
| 3.2 | **Opus** | 12 handler extractions + test migration. The biggest design surface in the initiative. |
| 4.1 | Sonnet | Function move + delete. |

Switch via `/model` at each phase boundary.

<br/>

---

## Effort 1 — Tag realignment

**Why first:** the existing tag refactor (`b62e87d`) ratified the wrong direction. Realigning tag first establishes the self-bootstrap pattern that Efforts 2 and 3 follow directly.

### Phase 1.1 — `tag.New()` self-registers; delete bootstrap

**Files to modify:**

1. `internal/faction/engine/tag/tag.go` — `New()` self-registers the four handlers:
   ```go
   package tag

   import (
       "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
       "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
       "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/tag/tags"
       "github.com/therobertcrocker/gm-toolkit/internal/faction/state"
   )

   type Handler interface {
       TagID() string
       Apply(faction *domain.Faction, hookRegistry *hooks.Registry)
   }

   type TagEngine struct {
       handlers map[string]Handler
   }

   func New() *TagEngine {
       e := &TagEngine{handlers: make(map[string]Handler)}
       e.Register(tags.ScavengersHandler{})
       e.Register(tags.WarlikeHandler{})
       e.Register(tags.FanaticalHandler{})
       e.Register(tags.PreceptorArchiveHandler{})
       return e
   }

   func (e *TagEngine) Register(handler Handler) {
       e.handlers[handler.TagID()] = handler
   }

   // ApplyAll walks every faction-tag in factionState and invokes the registered
   // handler for that tag ID. Tags with no registered handler are silently
   // skipped — they are data-only entries from the rulebook.
   func (e *TagEngine) ApplyAll(factionState *state.FactionState, hookRegistry *hooks.Registry) {
       for _, faction := range factionState.Factions {
           for _, tag := range faction.Tags {
               handler, ok := e.handlers[tag.ID]
               if !ok {
                   continue
               }
               handler.Apply(faction, hookRegistry)
           }
       }
   }
   ```
   Note: `tag` now imports `tag/tags`. Parent-imports-child is fine in Go and doesn't create a cycle because the child package never imports the parent (handler structs satisfy `tag.Handler` implicitly).

2. `internal/faction/engine/testharness/harness.go`:
   - Remove the `tags.RegisterDefaultTags(eng)` call at line 73.
   - Remove the `tag/tags` import at line 16.

**Files to delete:**

3. `internal/faction/engine/tag/tags/register.go` — entire file. The bootstrap is gone.

**Test impact:**

- Tag tests in `tag/tags/*_test.go` test the reactor structs directly; no test changes.
- Any scenario using the harness gets the same four handlers wired, just via `tag.New()` instead of `tags.RegisterDefaultTags`. Behavior identical.

**Acceptance criteria:**

- [ ] `grep -rn "engine" internal/faction/engine/tag/` returns no results other than `engine/hooks` and `engine/state` (re-verify: nothing imports the parent `engine` package).
- [ ] `internal/faction/engine/tag/tags/register.go` does not exist.
- [ ] `testharness/harness.go` has no reference to `tags.RegisterDefaultTags`.
- [ ] `go build ./...` succeeds.
- [ ] `go test ./internal/faction/...` is green.

**Commit message:** `refactor(tag): self-bootstrap defaults in tag.New(); drop register.go`

<br/>

---

## Effort 2 — Ability

**Why:** `ability.New()` registers two step handlers inline at lines 43–44 because the discovery treated this as Shape 2. But `domain.AbilityStepType` is a closed domain enum (two values: `AbilityStepMovement`, `AbilityStepFactionTest`) — a registry over a closed enum is ceremony. The honest shape is a `switch`. Separately, `action.Collector` doesn't compose `ability.Collector`, forcing a `c.(engine.InputCollector)` cast in `action/actions/register.go:20`.

### Phase 2.1 — Step dispatch collapses to switch; handlers move to ability/steps/

**Files to create:**

1. `internal/faction/engine/ability/steps/movement.go`:
   - Move `movementStepHandler` (currently `ability/ability.go:84–110`) into this file as exported `Movement`.
   - Move `worldsFromState` (lines 206–227) as unexported `worldsFromState`.
   - Package: `steps`. Imports: `domain`, `rulebook`, `state`, `ability` (for `Collector` type only — no `engine` import).
   - Public surface:
     ```go
     func Movement(
         faction *domain.Faction,
         asset *domain.Asset,
         step domain.AbilityStep,
         collector ability.Collector,
         roller domain.Roller,
         factionState *state.FactionState,
         rulebook *rulebook.Rulebook,
     ) ([]domain.Mutation, error)
     ```

2. `internal/faction/engine/ability/steps/faction_check.go` (note: not `faction_test.go` — Go would treat that as a test file):
   - Move `factionTestStepHandler` (lines 112–138) as exported `FactionCheck`.
   - Move `factionTestCandidates` (lines 140–161), `applyAbilityEffect` (lines 163–204), `abilityStatScore` (lines 229–240) as unexported helpers.
   - Same imports as `movement.go`.

**Files to modify:**

3. `internal/faction/engine/ability/ability.go`:
   - Remove the `stepHandlers` map field entirely.
   - Strip `StepHandler` type definition (lines 12–21) — it's no longer needed.
   - Remove the `RegisterStepHandler` plan from the previous doc (it never existed in the current code).
   - `New()` becomes:
     ```go
     func New() *AbilityEngine {
         return &AbilityEngine{
             customHandlers: make(map[string]CustomAbilityHandler),
         }
     }
     ```
   - Modify `Run` (lines 54–82) to dispatch step types via a switch instead of map lookup:
     ```go
     for _, step := range def.Ability.Steps {
         stepMutations, err := runStep(faction, asset, step, collector, roller, factionState, rulebook)
         if err != nil {
             return nil, err
         }
         mutations = append(mutations, stepMutations...)
     }
     ```
     With a private package-level `runStep`:
     ```go
     func runStep(
         faction *domain.Faction,
         asset *domain.Asset,
         step domain.AbilityStep,
         collector Collector,
         roller domain.Roller,
         factionState *state.FactionState,
         rulebook *rulebook.Rulebook,
     ) ([]domain.Mutation, error) {
         switch step.Type {
         case domain.AbilityStepMovement:
             return steps.Movement(faction, asset, step, collector, roller, factionState, rulebook)
         case domain.AbilityStepFactionTest:
             return steps.FactionCheck(faction, asset, step, collector, roller, factionState, rulebook)
         default:
             return nil, fmt.Errorf("no handler for ability step type %q", step.Type)
         }
     }
     ```
   - Delete the now-orphaned inline functions (`movementStepHandler`, `factionTestStepHandler`, and the helpers — they moved to `steps/`).
   - Add import: `"github.com/.../internal/faction/engine/ability/steps"`.

**Test impact:**

- `ability/ability_test.go` exercises `Run()` end-to-end; unaffected by the internal switch reorganization.
- `ability/steps/` may want a small focused test file later, but no migration is required for this phase.

**Acceptance criteria:**

- [ ] `AbilityEngine` has no `stepHandlers` field.
- [ ] `ability/ability.go` has no `movementStepHandler` or `factionTestStepHandler`.
- [ ] `ability/steps/movement.go` and `ability/steps/faction_check.go` exist with exported `Movement` and `FactionCheck`.
- [ ] `grep -rn "engine\"" internal/faction/engine/ability/` returns nothing (no parent-`engine` import in the ability tree).
- [ ] `go build ./...` succeeds.
- [ ] `go test ./internal/faction/...` is green.

**Commit message:** `refactor(ability): collapse step dispatch to switch; extract handlers to ability/steps/`

### Phase 2.2 — Compose ability.Collector into action.Collector

(Identical to the original plan's 2.2 — the cast removal is independent of the bootstrap question.)

**Files to modify:**

1. `internal/faction/engine/action/collector.go`:
   - Add `ability.Collector` to the `Collector` interface composition:
     ```go
     type Collector interface {
         hooks.Collector
         ability.Collector
         SelectAsset(...) (*domain.Asset, error)
         // ...rest unchanged
     }
     ```
   - Add import: `"github.com/.../internal/faction/engine/ability"`.

2. `internal/faction/engine/action/actions/register.go`:
   - Line 20: change `NewUseAssetAbility(c.(engine.InputCollector), e.Rand, e.Ability)` to `NewUseAssetAbility(c, e.Rand, e.Ability)`.

3. `internal/faction/engine/action/actions/use_asset_ability.go`:
   - Drop `engine` import.
   - Add `action` import.
   - Retype `collector` field and `NewUseAssetAbility`'s first parameter from `engine.InputCollector` to `action.Collector`.

4. `internal/faction/engine/collector.go`:
   - `InputCollector` no longer needs to explicitly embed `ability.Collector` (now transitively via `action.Collector`). Remove that line and the `ability` import.

**Test impact:** mock at `action/actions/mocks/mock_collector.go` is generated from `InputCollector`. Regeneration is safer; otherwise the existing mock still satisfies the equivalent interface.

**Acceptance criteria:**

- [ ] `grep -n "engine.InputCollector" internal/faction/engine/action/` returns no results.
- [ ] `grep -rn "c\.(engine\.InputCollector)" internal/faction/` returns no results.
- [ ] `go build ./...` succeeds.
- [ ] `go test ./internal/faction/...` is green.

**Commit message:** `refactor(action): compose ability.Collector; drop engine.InputCollector cast`

<br/>

---

## Effort 3 — Goal data-mirroring promotion

**Why:** `goal.GoalEngine` is an empty struct dispatching through hardcoded switches over 11 progress IDs and 2 lock IDs. The rulebook's `goals.toml` is an open data registry — a GM-authored goal silently no-ops because the switch has no case. Promoting `goal` to Shape 2 with the self-bootstrap pattern makes "data exists, code doesn't" a first-class state and aligns goal with tag.

### Phase 3.1 — Introduce parallel registry path

Land the interface and registry; keep the switches as a parallel fall-through path so behavior is unchanged. Zero handler migrations in this phase.

**Files to modify:**

1. `internal/faction/engine/goal/goal.go`:
   - Add `Handler` interface (note: still uses `actingFaction *domain.Faction`, not `actingFactionID string`, to align with how handlers will be shaped):
     ```go
     type Handler interface {
         GoalID() string
         CheckLock(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) (GoalLock, []domain.Mutation)
         UpdateProgress(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, rulebook *rulebook.Rulebook, index *world.Index) []domain.Mutation
     }
     ```
   - Add map + `Register`:
     ```go
     type GoalEngine struct {
         handlers map[string]Handler
     }

     func New() *GoalEngine {
         return &GoalEngine{handlers: make(map[string]Handler)}
         // Phase 3.2 will add e.Register(goals.Foo{}) calls here.
     }

     func (e *GoalEngine) Register(handler Handler) {
         e.handlers[handler.GoalID()] = handler
     }
     ```
   - Modify `CheckLock` and `UpdateProgress` to consult the registry first and fall through to the switch:
     ```go
     func (ge *GoalEngine) CheckLock(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) (GoalLock, []domain.Mutation) {
         if faction.ActiveGoal == nil {
             return GoalLock{Type: LockNone}, nil
         }
         if handler, ok := ge.handlers[faction.ActiveGoal.GoalID]; ok {
             return handler.CheckLock(faction, factionState, rulebook)
         }
         // Parallel path during Effort 3 migration; removed in Phase 3.2.
         switch faction.ActiveGoal.GoalID {
         case "G-012":
             return checkLockChangeHomeworld(faction)
         case "G-004":
             return checkLockPlanetarySeizure(faction, factionState, rulebook)
         }
         return GoalLock{Type: LockNone}, nil
     }
     ```
   - Same symmetric treatment for `UpdateProgress`.

**Test impact:** none. All goal tests still call unexported `progress*` / `checkLock*` functions directly; the switch still routes to them.

**Acceptance criteria:**

- [ ] `goal.Handler` interface defined.
- [ ] `GoalEngine.handlers` map and `Register` method exist.
- [ ] `CheckLock` and `UpdateProgress` check registry first; switches retained as fall-through.
- [ ] `go test ./internal/faction/engine/goal/...` is green.

**Commit message:** `refactor(goal): introduce Handler interface and parallel registry path`

### Phase 3.2 — Migrate 12 handlers; self-bootstrap; delete switches

**The big one.** Pull every progress/lock function into a per-goal struct in `goal/goals/`, migrate the tests, register everything in `goal.New()`, then delete the parallel switches and the helper files.

**Files to create:**

One handler file per goal ID (12 total) in `internal/faction/engine/goal/goals/`. Most handlers have a no-op `CheckLock`; only G-004 and G-012 implement non-trivial locks.

| File | Goal ID | Source progress fn | Source lock fn |
|---|---|---|---|
| `military_conquest.go` | G-001 | `progressMilitaryConquest` | — |
| `commercial_expansion.go` | G-002 | `progressCommercialExpansion` | — |
| `intelligence_coup.go` | G-003 | `progressIntelligenceCoup` | — |
| `planetary_seizure.go` | G-004 | `progressPlanetarySeizure` | `checkLockPlanetarySeizure` |
| `expand_influence.go` | G-005 | `progressExpandInfluence` | — |
| `blood_the_enemy.go` | G-006 | `progressBloodTheEnemy` | — |
| `peaceable_kingdom.go` | G-007 | `progressPeaceableKingdom` | — |
| `destroy_the_foe.go` | G-008 | `progressDestroyTheFoe` | — |
| `inside_enemy_territory.go` | G-009 | `progressInsideEnemyTerritory` | — |
| `invincible_valor.go` | G-010 | `progressInvincibleValor` | — |
| `wealth_of_worlds.go` | G-011 | `progressWealthOfWorlds` | — |
| `change_homeworld.go` | G-012 | — | `checkLockChangeHomeworld` |

Standard handler shape (G-001 illustration; all progress-only handlers are identical in structure):

```go
package goals

import (
    "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type MilitaryConquest struct{}

func (MilitaryConquest) GoalID() string { return "G-001" }

func (MilitaryConquest) CheckLock(_ *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) (goal.GoalLock, []domain.Mutation) {
    return goal.GoalLock{Type: goal.LockNone}, nil
}

func (MilitaryConquest) UpdateProgress(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, rulebook *rulebook.Rulebook, _ *world.Index) []domain.Mutation {
    // body of the old progressMilitaryConquest, verbatim
}
```

**Helpers (file: `goal/goals/helpers.go`):**

Move these unexported helpers from `goal/progress.go` and `goal/lock.go` into `goals` package as unexported package-level functions:

- From `progress.go`: `completeGoal`, `countAssetKillsByCategory`, `findAsset`, `factionHasBaseOn`, `worldHasRivalPresence`, `rivalHasPlanetaryGovernmentOnWorld`, `factionHasPlanetaryGovernmentTag`
- From `lock.go`: `factionHasUnstealthedAssetOn`, `calcPlanetarySeizureXP`

All move verbatim. Imports rebuilt for the new package context.

**Self-bootstrap in `goal/goal.go`'s `New()`:**

```go
import (
    "github.com/.../internal/faction/engine/goal/goals"
    // ... other imports
)

func New() *GoalEngine {
    e := &GoalEngine{handlers: make(map[string]Handler)}
    e.Register(goals.MilitaryConquest{})
    e.Register(goals.CommercialExpansion{})
    e.Register(goals.IntelligenceCoup{})
    e.Register(goals.PlanetarySeizure{})
    e.Register(goals.ExpandInfluence{})
    e.Register(goals.BloodTheEnemy{})
    e.Register(goals.PeaceableKingdom{})
    e.Register(goals.DestroyTheFoe{})
    e.Register(goals.InsideEnemyTerritory{})
    e.Register(goals.InvincibleValor{})
    e.Register(goals.WealthOfWorlds{})
    e.Register(goals.ChangeHomeworld{})
    return e
}
```

**Files to modify:**

`goal/goal.go`:
- Delete both parallel switches.
- Replace each with the data-only fall-through:
  ```go
  func (ge *GoalEngine) CheckLock(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) (GoalLock, []domain.Mutation) {
      if faction.ActiveGoal == nil {
          return GoalLock{Type: LockNone}, nil
      }
      handler, ok := ge.handlers[faction.ActiveGoal.GoalID]
      if !ok {
          // Data-only goal: present in goals.toml, no Go handler registered. Intentional.
          return GoalLock{Type: LockNone}, nil
      }
      return handler.CheckLock(faction, factionState, rulebook)
  }
  ```
  Same shape for `UpdateProgress` (returning `nil` mutations on miss).
- After migration, `goal.go` only needs `domain`, `state`, `rulebook`, `world`, and `goal/goals` imports.

**No harness change.** `goal.New()` self-wires; no `goals.RegisterDefaultGoals(eng)` exists.

**Files to delete:**

- `internal/faction/engine/goal/progress.go` — empty after migration.
- `internal/faction/engine/goal/lock.go` — empty after migration.

  **Important:** before deleting `lock.go`, move the `LockType`, `LockNone`, `LockSkip`, `LockRestrictActions`, and `GoalLock` declarations to `goal/goal.go` (or a new `goal/types.go`). The `goals/` handlers depend on these types.

**Test migration:**

The 19 existing goal tests in `package goal` call unexported functions directly. Move into `package goals`:

| Old file | New file | What changes |
|---|---|---|
| `goal/progress_test.go` (9 tests) | `goal/goals/progress_test.go` | `package goals`; rewrite each call from `progressMilitaryConquest(acting, ...)` to `MilitaryConquest{}.UpdateProgress(acting, ...)`; types like `GoalLock` reference as `goal.GoalLock`. |
| `goal/lock_test.go` (4 tests) | `goal/goals/lock_test.go` | `package goals`; same pattern. |
| `goal/world_index_test.go` (6 tests) | `goal/goals/world_index_test.go` | `package goals`; same pattern. |

Test helpers (e.g., `makeProgressRulebook`) move into `goals` package or `goals_test` as appropriate.

**Acceptance criteria:**

- [ ] `internal/faction/engine/goal/progress.go` and `lock.go` are deleted.
- [ ] Twelve handler files exist in `goal/goals/`, plus `helpers.go`.
- [ ] **No `goal/goals/register.go` exists** — defaults are wired by `goal.New()`.
- [ ] `goal/goal.go` contains no `switch` over goal IDs; the data-only comment exists at both fall-throughs.
- [ ] Tests have moved into `goal/goals/` and reference handler structs directly.
- [ ] `grep -rn "engine\"" internal/faction/engine/goal/` returns nothing (no `engine` import in the goal tree).
- [ ] `grep -rn "case \"G-" internal/faction/engine/goal/` returns no results.
- [ ] `go test ./internal/faction/...` is green.
- [ ] Smoke check: adding a dummy `G-099` to a test rulebook fixture, confirm `CheckLock`/`UpdateProgress` return `nil` cleanly.

**Commit message:** `refactor(goal): promote to Shape 2 self-bootstrapping; migrate 12 handlers to goal/goals/`

<br/>

---

## Effort 4 — History demotion into turn/

**Why:** `HistoryEngine` is an empty struct holding a single method that writes one JSON line. Demoting it to a package-level function in `turn/` reflects that the record is conceptually a turn artifact and removes one over-structured sub-engine from the composition root.

### Phase 4.1 — Move Record into turn/, delete engine/history/

**Files to create:**

1. `internal/faction/engine/turn/history.go`:
   - Package-level `RecordHistory(historyPath, factionState, faction, mutations) error` mirroring the body of `history.HistoryEngine.Record` (currently `engine/history/history.go:18–39`).

2. `internal/faction/engine/turn/event_record.go`:
   - Move `buildEventRecord` from `engine/history/record.go:11–26` here as unexported.

3. `internal/faction/engine/turn/history_test.go`:
   - Move `TestHistoryEngine_Record` from `engine/history/history_test.go`. Rename to `TestRecordHistory`. Replace `he := New(); he.Record(...)` with `turn.RecordHistory(...)`. Match the existing `turn/*_test.go` package convention (`turn` or `turn_test`).

**Files to modify:**

4. `internal/faction/engine/core.go`:
   - Remove the `history` import (line 9).
   - Remove the `History *history.HistoryEngine` field (line 40).
   - Remove `e.History = history.New()` from `NewWithRulebook` (line 67).
   - Update the doc comment around lines 26–28 to drop `history/` from the supporting-packages list.

5. `internal/faction/engine/orchestrator.go`:
   - In `applyAndRecord` (around line 253), replace `e.History.Record(...)` with `turn.RecordHistory(...)`. Verify the `turn` import already exists; add if not.

**Files to delete:**

6. `internal/faction/engine/history/` — entire directory.

**Acceptance criteria:**

- [ ] `internal/faction/engine/history/` no longer exists.
- [ ] `grep -rn "engine/history\|history\.HistoryEngine\|\.History\." internal/faction/ cmd/` returns no results.
- [ ] `*Engine` no longer has a `History` field.
- [ ] `turn.RecordHistory` is the only history-writing path.
- [ ] `go test ./internal/faction/...` is green.

**Commit message:** `refactor(history): demote to turn.RecordHistory; delete engine/history package`

<br/>

---

## Post-Initiative State

After all four efforts merge:

- **Nine sub-engines collapse to eight** (`history` is gone; its function lives in `turn`).
- **Every remaining sub-engine fits the revised five-shape catalog:**
  - Shape 1 (closed dispatcher): `mutation`, ability *step* dispatch
  - Shape 2 (self-bootstrapping registry): `tag`, `goal`, ability *custom* dispatch
  - Shape 3 (factory registry with engine deps): `action`
  - Shape 4 (runtime-populated scoped registry): `hooks`
  - Shape 5 (state-owning subsystem): `turn`, `world`
- **Dependency direction is uniform:** every sub-engine package (except `action/actions/`) imports only `domain`, `state`, `rulebook`, `hooks`, and sibling sub-engines / sibling handler packages as needed. **Nothing in the tag, goal, or ability trees imports the parent `engine` package.**
- **The "designated cheat file" convention is gone.** No rule like "only `foo/foos/register.go` may import engine" exists. Action's `actions/register.go` is the only file that imports `engine` in the entire sub-engine tree, and it's the documented exception, not the template.
- **The composition root is pure assembly.** `core.go`'s `NewWithRulebook` is a one-line-per-sub-engine sequence: each `New()` returns a fully-wired engine.
- **The test harness's engine setup is one bootstrap call** — `actions.RegisterDefaultActions(eng)` — instead of two.
- **Discovery Open Questions 1–5 are closed** in the Decisions Ratified table.

The five-shape catalog supersedes the discovery's three-shape catalog as the reference for future sub-engines. A new sub-engine that doesn't fit one of the five shapes should prompt re-examination — and if it's genuinely a new shape, the catalog grows. Mixed-shape sub-engines (multiple dispatch tables in one struct) are legitimate.

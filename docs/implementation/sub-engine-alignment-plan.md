# Sub-Engine Alignment — Implementation Plan

> Discovery: [`../discovery/sub-engine-alignment-discovery.md`](../discovery/sub-engine-alignment-discovery.md)

Brings the nine packages under `internal/faction/engine/` into structural alignment with the three sub-engine shapes catalogued in the discovery. Four efforts, six phases, six commits. Single-file plan: each phase is self-contained and can be executed in its own session without re-reading the discovery.

<br/>

## Decisions Ratified in Planning

| # | Decision |
|---|---|
| 1 | `goal` is promoted to **Shape 2 data-mirroring** (full parity with the proposed `tag` shape). The TOML rulebook owns the goal ID universe; the handler set is an explicit subset, and known-data/no-handler is a documented "data-only" state at the lookup site. |
| 2 | `action.Collector` **composes** `ability.Collector`. The `c.(engine.InputCollector)` cast at `action/actions/register.go:20` is removed; `UseAssetAbility.collector` is retyped to `action.Collector`. |
| 3 | `history` is **demoted into `turn/`** as a package-level `turn.RecordHistory(...)` function. The `History` field is removed from `*Engine`; the `internal/faction/engine/history/` package is deleted. |
| 4 | `world` stays optional. The `if e.World != nil` guards at `orchestrator.go:36` and `:146` (and `action/actions/register.go:27`) remain, documented but not removed. |
| 5 | `world.Engine` → `world.WorldEngine` rename is **already complete** (verified pre-plan; visible in `core.go:40`). Discovery Open Question 5 is closed. |
| 6 | Plan structure: single file in `docs/implementation/sub-engine-alignment-plan.md`. No per-effort sub-plans — total content is comfortably one document and cross-cutting context lives once. |

<br/>

## Sub-Engine Shape Recap

The three canonical shapes (from the discovery) — kept here so the plan stands alone:

- **Shape 1 — Stateless Dispatcher.** Empty struct, bare `New()`, closed dispatch, no registration. `mutation` is the exemplar.
- **Shape 2 — Open Registry.** Engine holds a registry, `New()` is bare, sibling `foo/foos/` package holds concrete handlers, sibling-package `RegisterDefaultFoos(eng)` bootstrap. Two sub-flavors:
  - *Open-ended* — handler registry defines what exists (`action`, `hooks`).
  - *Data-mirroring* — rulebook TOML defines the ID universe; handler set is a subset; known-but-unhandled is a documented data-only state (`tag` target, `goal` target).
- **Shape 3 — State-Owning Subsystem.** Constructor takes deps and may return an error, capability-injection alt-constructor for tests, owns lifecycle state. `turn`, `world` are the exemplars.

<br/>

## Effort & Phase Map

| Effort | Phase | Goal | Touches | Behavior change |
|---|---|---|---|---|
| 1 — Tag refactor | 1.1 | Convert `tag` to Shape 2 data-mirroring; flip the dep direction. | `tag/`, `tag/tags/`, `core.go`, `testharness/harness.go` | None |
| 2 — Ability + Action | 2.1 | Extract ability step handlers to `ability/steps/`; bare `ability.New()`. | `ability/`, new `ability/steps/`, `testharness/harness.go` | None |
| 2 — Ability + Action | 2.2 | Compose `ability.Collector` into `action.Collector`; drop the cast. | `action/collector.go`, `action/actions/register.go`, `action/actions/use_asset_ability.go`, `engine/collector.go` | None |
| 3 — Goal promotion | 3.1 | Introduce `goal.Handler` interface + registry; keep existing switches as parallel path. | `goal/goal.go` | None |
| 3 — Goal promotion | 3.2 | Migrate all 13 handlers to `goal/goals/`; migrate tests; delete switches. | `goal/`, new `goal/goals/`, `testharness/harness.go` | None |
| 4 — History demotion | 4.1 | Move `Record` into `turn/`; delete `engine/history/`. | `turn/` (new file), `core.go`, `orchestrator.go`, deletes `history/` | None |

Every phase is **behavior-preserving**. Tests stay green throughout.

<br/>

## Cross-Cutting Notes

### Bootstrap convention

After this initiative, every Shape 2 sub-engine follows the same bootstrap pattern, mirroring `action/actions/register.go`:

```go
// In package foo/foos:
func RegisterDefaultFoos(eng *engine.Engine) {
    eng.Foo.Register(HandlerA{})
    eng.Foo.Register(HandlerB{})
    // ...
}
```

The bootstrap file is the *only* file in `foo/foos/` permitted to import `engine`. Concrete handler files (`foo/foos/handler_a.go`) import only `domain`, `state`, `rulebook`, and sibling sub-engines as needed — never `engine`.

### No production bootstrap caller yet

Neither `RegisterDefaultActions`, `RegisterDefaultTags`, nor any future `RegisterDefaultSteps` / `RegisterDefaultGoals` is currently invoked from `cmd/`. The only caller of all of them is `internal/faction/engine/testharness/harness.go`. This means every phase below updates *one* call site (the test harness) — not a CLI binary too. When the CLI gains a faction-engine entry point, it will need to invoke all four bootstraps.

### Data-mirroring lookup-site contract

For data-mirroring sub-engines (`tag` after Effort 1, `goal` after Effort 3), the lookup site must treat "ID exists in the rulebook data but no handler is registered" as a *designed state*, not a fall-through. Standard form:

```go
handler, ok := e.handlers[id]
if !ok {
    // Data-only entry: known to the rulebook, not implemented in code. Intentional.
    return nil
}
```

A comment at the lookup site is required — this is the only "no comment unless WHY is non-obvious" exception in this plan.

### Pre-merge checklist (per CLAUDE.md)

Before merging the branch carrying any effort:

1. Self-review as a senior reviewing a junior. Be critical, constructive.
2. Update `docs/dev_journals/faction-manager/dev-journal-factions.md` with effort summary + notes.
3. Update `docs/dev_journals/faction-manager/planned-work.md` (remove completed items, add anything that surfaced).
4. Update `docs/dev_journals/faction-manager/decisions-log.md` only for decisions ratified *during execution*. The six decisions above belong in this plan, not the log.
5. Assess version bump (patch/minor/major). Tag after merge.

### Model selection

| Phase | Recommended model | Notes |
|---|---|---|
| 1.1 | Sonnet | Mechanical refactor + interface introduction. |
| 2.1 | Sonnet | Mechanical extraction. |
| 2.2 | Sonnet | Three-line change at the interface + cast site. |
| 3.1 | Sonnet | New interface + parallel-path wiring. |
| 3.2 | **Opus** | 13 handler extractions + test migration; biggest design surface in the initiative. |
| 4.1 | Sonnet | Function move + delete. |

Switch via `/model` at each phase boundary.

<br/>

---

## Effort 1 — Tag refactor

**Why first:** the architectural problem. `tag/` currently imports `engine/`, inverting the dep direction every other sub-engine respects. Fixing this unblocks treating tag as a normal sub-engine going forward, and produces the template that Effort 3 (`goal`) follows directly.

### Phase 1.1 — Convert `tag` to Shape 2 data-mirroring

**Files to create:**

1. `internal/faction/engine/tag/tags/register.go` — bootstrap function. Imports `engine`.
   ```go
   package tags

   import "github.com/therobertcrocker/gm-toolkit/internal/faction/engine"

   func RegisterDefaultTags(eng *engine.Engine) {
       eng.Tag.Register(ScavengersHandler{})
       eng.Tag.Register(WarlikeHandler{})
       eng.Tag.Register(FanaticalHandler{})
       eng.Tag.Register(PreceptorArchiveHandler{})
   }
   ```

**Files to modify:**

2. `internal/faction/engine/tag/tag.go` — replace entire file contents:
   ```go
   package tag

   import (
       "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
       "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
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
       return &TagEngine{handlers: make(map[string]Handler)}
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
   Note the import list: **no `engine` import**. Dependency direction flipped.

3. `internal/faction/engine/tag/tags/scavengers.go` — replace `RegisterScavengers` free function (lines 40–46) with a handler struct. The existing `ScavengersReactor` stays untouched. Drop the `engine` import.
   ```go
   type ScavengersHandler struct{}

   func (ScavengersHandler) TagID() string { return ScavengersTagID }

   func (ScavengersHandler) Apply(faction *domain.Faction, hookRegistry *hooks.Registry) {
       hookRegistry.RegisterMutationReactor(
           hooks.FactionScope(faction.ID),
           "scavengers",
           &ScavengersReactor{FactionID: faction.ID},
       )
   }
   ```
   After this change, the file imports only `domain`, `hooks`, `rulebook`, `state`.

4. `internal/faction/engine/tag/tags/warlike.go` — same treatment. Replace `RegisterWarlike` with `WarlikeHandler{}`; drop `engine` import. The hook category Warlike registers is `hooks.RollModifier` — preserve the current registration call inside `Apply`. Use `WarlikeTagID` for `TagID()`.

5. `internal/faction/engine/tag/tags/fanatical.go` — same. `FanaticalHandler{}`, drop `engine` import, preserve the `hooks.RollResultHook` registration. Use `FanaticalTagID`.

6. `internal/faction/engine/tag/tags/preceptor_archive.go` — same. `PreceptorArchiveHandler{}`, drop `engine` import, preserve the `hooks.AssetCostModifier` registration. Use `PreceptorArchiveTagID`.

7. `internal/faction/engine/core.go`:
   - Add import: `"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/tag"`
   - Add field to `Engine` struct (alphabetical position): `Tag *tag.TagEngine`
   - In `NewWithRulebook`, add: `e.Tag = tag.New()` (after the other sub-engines).

8. `internal/faction/engine/testharness/harness.go`:
   - Add import: `"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/tag/tags"`
   - In `NewHarness` (after `actions.RegisterDefaultActions(eng)` at line 69): add `tags.RegisterDefaultTags(eng)`.
   - Replace the body of `RegisterTags` (lines 80–82) with:
     ```go
     func (h *Harness) RegisterTags() {
         h.Engine.Tag.ApplyAll(h.FactionState, h.Engine.Hooks)
     }
     ```
   - The old `tag` import at line 16 stays (now used via `eng.Tag` indirectly — actually no, the harness now imports `tag/tags` for the bootstrap, and reaches the engine via `h.Engine.Tag` for `ApplyAll`. The bare `tag` import can be removed).

**Files to delete:** none.

**Test impact:**

- Tag tests (`tag/tags/*_test.go`) test the reactor/modifier structs directly, not the registration entry points. No test changes required. Run `go test ./internal/faction/engine/tag/...` to confirm green.
- All faction-engine scenarios that depend on tags exercise `h.RegisterTags()` — they should pass unchanged because `ApplyAll` does the same walk as the old `RegisterDefaultTags`.

**Acceptance criteria:**

- [ ] `grep -r "engine" internal/faction/engine/tag/*.go internal/faction/engine/tag/tags/*.go | grep -v register.go` returns no results. (Only `register.go` may import the parent `engine` package.)
- [ ] `go build ./...` succeeds.
- [ ] `go test ./internal/faction/...` is green.
- [ ] `*Engine.Tag` is non-nil after `NewWithRulebook`.
- [ ] Test harness's `RegisterTags()` produces the same hook registrations as before.

**Commit message:** `refactor(tag): convert to Shape 2 data-mirroring sub-engine`

<br/>

---

## Effort 2 — Ability + Action template alignment

**Why:** `ability.New()` has inline registration of two step handlers, which violates Shape 2 (built-ins must be registered externally via the sibling-package bootstrap). And `action/actions/register.go` carries a `c.(engine.InputCollector)` cast because `action.Collector` doesn't include the ability methods. Both leaks are small and contained, and closing them brings `action` and `ability` fully into line with the template that Effort 3 follows.

### Phase 2.1 — Extract ability step handlers to `ability/steps/`

**Files to create:**

1. `internal/faction/engine/ability/steps/movement.go`:
   - Move `movementStepHandler` (currently `ability/ability.go:84–110`) into this file as `MovementStepHandler`.
   - Move `worldsFromState` helper (lines 206–227) into this file as unexported `worldsFromState`.
   - Package: `steps`. Imports: `domain`, `rulebook`, `state`, `ability` (for the `Collector` type referenced in the handler signature).
   - Public surface: `func MovementStepHandler(faction *domain.Faction, asset *domain.Asset, step domain.AbilityStep, collector ability.Collector, roller domain.Roller, factionState *state.FactionState, rulebook *rulebook.Rulebook) ([]domain.Mutation, error)`

2. `internal/faction/engine/ability/steps/faction_test.go`:
   - Move `factionTestStepHandler` (lines 112–138) into this file as `FactionTestStepHandler`.
   - Move `factionTestCandidates` (lines 140–161), `applyAbilityEffect` (lines 163–204), `abilityStatScore` (lines 229–240) into this file as unexported helpers.
   - **Warning:** name this file something other than `faction_test.go` so Go doesn't treat it as a test file. Use `faction_check.go` (since "faction test" is the rulebook term and the file is not a Go test). Update the function names accordingly if useful — `FactionCheckStepHandler` is clearer but `FactionTestStepHandler` preserves the rulebook vocabulary. Recommend: keep `FactionTestStepHandler`, name the file `faction_check.go`.

3. `internal/faction/engine/ability/steps/register.go`:
   ```go
   package steps

   import (
       "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
       "github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
   )

   func RegisterDefaultSteps(eng *engine.Engine) {
       eng.Ability.RegisterStepHandler(domain.AbilityStepMovement, MovementStepHandler)
       eng.Ability.RegisterStepHandler(domain.AbilityStepFactionTest, FactionTestStepHandler)
   }
   ```

**Files to modify:**

4. `internal/faction/engine/ability/ability.go`:
   - Strip the inline registrations from `New()` (lines 43–44). New body:
     ```go
     func New() *AbilityEngine {
         return &AbilityEngine{
             stepHandlers:   make(map[domain.AbilityStepType]StepHandler),
             customHandlers: make(map[string]CustomAbilityHandler),
         }
     }
     ```
   - Add the exported method:
     ```go
     func (ae *AbilityEngine) RegisterStepHandler(stepType domain.AbilityStepType, handler StepHandler) {
         ae.stepHandlers[stepType] = handler
     }
     ```
   - Delete `movementStepHandler`, `factionTestStepHandler`, `factionTestCandidates`, `applyAbilityEffect`, `worldsFromState`, `abilityStatScore`. They're now in `steps/`.

5. `internal/faction/engine/testharness/harness.go`:
   - Add import: `"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/ability/steps"`
   - In `NewHarness` (after `actions.RegisterDefaultActions(eng)`): `steps.RegisterDefaultSteps(eng)`. Place it before or after the `tags.RegisterDefaultTags(eng)` call from Phase 1.1 — order doesn't matter.

**Test impact:**

- `ability/ability_test.go` exists but does not reference the helper functions directly (verified via grep). Run `go test ./internal/faction/engine/ability/...` to confirm.
- Any scenario test that relied on movement or faction-test resolution implicitly exercises the new step-handler registrations.

**Acceptance criteria:**

- [ ] `ability.New()` body is the two-line struct construction; no `ae.stepHandlers[...] = ...` calls.
- [ ] `MovementStepHandler` and `FactionTestStepHandler` are exported from `ability/steps/`.
- [ ] `go build ./...` succeeds.
- [ ] `go test ./internal/faction/...` is green.

**Commit message:** `refactor(ability): extract step handlers to ability/steps/ sibling package`

### Phase 2.2 — Compose `ability.Collector` into `action.Collector`

**Files to modify:**

1. `internal/faction/engine/action/collector.go`:
   - Add import: `"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/ability"`
   - Modify the `Collector` interface (line 11) to embed `ability.Collector`:
     ```go
     type Collector interface {
         hooks.Collector
         ability.Collector
         SelectAsset(...) (*domain.Asset, error)
         // ... rest unchanged
     }
     ```

2. `internal/faction/engine/action/actions/register.go`:
   - Line 20 currently reads:
     ```go
     return NewUseAssetAbility(c.(engine.InputCollector), e.Rand, e.Ability)
     ```
   - Replace with:
     ```go
     return NewUseAssetAbility(c, e.Rand, e.Ability)
     ```

3. `internal/faction/engine/action/actions/use_asset_ability.go`:
   - Drop the `engine` import (line 7).
   - Add the `action` import: `"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"`
   - Change the `collector` field type (line 16) from `engine.InputCollector` to `action.Collector`.
   - Change `NewUseAssetAbility`'s first parameter type (line 23) from `engine.InputCollector` to `action.Collector`.

4. `internal/faction/engine/collector.go`:
   - The `InputCollector` interface (line 15) currently composes `action.Collector`, `ability.Collector`, `hooks.Collector`. After this phase, `action.Collector` already embeds `ability.Collector`, making the explicit `ability.Collector` line in `InputCollector` redundant. Remove it:
     ```go
     type InputCollector interface {
         action.Collector  // now transitively includes ability.Collector
         hooks.Collector
         SelectAction(faction *domain.Faction, available []action.Action) (action.Action, error)
         AwaitCheckpoint(phase string) error
         SelectStatRaise(faction *domain.Faction, eligible []domain.FactionStat) (*domain.FactionStat, error)
     }
     ```
   - Drop the `ability` import from this file (no longer referenced).

**Test impact:**

- The mockgen directive at `collector.go:3` generates `action/actions/mocks/mock_collector.go` from `InputCollector`. Re-run mockgen if the team policy is to regenerate, otherwise the existing mock continues to satisfy the (now-equivalent) interface. Regeneration is safer; the command is in the `//go:generate` directive.
- `action/actions/use_asset_ability_test.go` uses the generated mock collector — confirm it still compiles after the interface composition change.

**Acceptance criteria:**

- [ ] `grep -n "engine.InputCollector" internal/faction/engine/action/` returns no results.
- [ ] `grep -n "\.(engine\.InputCollector)" internal/faction/engine/` returns no results (the cast is gone).
- [ ] `go build ./...` succeeds.
- [ ] `go test ./internal/faction/...` is green.

**Commit message:** `refactor(action): compose ability.Collector to close UseAssetAbility wiring leak`

<br/>

---

## Effort 3 — Goal data-mirroring promotion

**Why:** `goal.GoalEngine` is an empty struct dispatching through hardcoded switches (11 progress IDs in `UpdateProgress`, 2 lock IDs in `CheckLock`). The rulebook already loads `goals.toml` as an open data registry — a GM-authored goal added to TOML today silently no-ops because the switch has no case. Promoting `goal` to Shape 2 data-mirroring brings it into parity with `tag` (after Effort 1) and makes "data exists, code doesn't" a first-class state with a comment.

### Phase 3.1 — Introduce parallel registry path

This phase lands the new infrastructure (interface, registry, lookup-first dispatch) without migrating any handlers. The registry is empty after `New()`, so both `CheckLock` and `UpdateProgress` fall through to the existing switches. Zero behavior change. Test surface unchanged.

**Files to modify:**

1. `internal/faction/engine/goal/goal.go`:
   - Add a `Handler` interface:
     ```go
     type Handler interface {
         GoalID() string
         CheckLock(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) (GoalLock, []domain.Mutation)
         UpdateProgress(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, rulebook *rulebook.Rulebook, index *world.Index) []domain.Mutation
     }
     ```
   - Change `GoalEngine`:
     ```go
     type GoalEngine struct {
         handlers map[string]Handler
     }

     func New() *GoalEngine {
         return &GoalEngine{handlers: make(map[string]Handler)}
     }

     func (e *GoalEngine) Register(handler Handler) {
         e.handlers[handler.GoalID()] = handler
     }
     ```
   - Modify `CheckLock` to consult the registry first:
     ```go
     func (ge *GoalEngine) CheckLock(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) (GoalLock, []domain.Mutation) {
         if faction.ActiveGoal == nil {
             return GoalLock{Type: LockNone}, nil
         }
         if handler, ok := ge.handlers[faction.ActiveGoal.GoalID]; ok {
             return handler.CheckLock(faction, factionState, rulebook)
         }
         // Parallel path during Effort 3 migration; remove after Phase 3.2.
         switch faction.ActiveGoal.GoalID {
         case "G-012":
             return checkLockChangeHomeworld(faction)
         case "G-004":
             return checkLockPlanetarySeizure(faction, factionState, rulebook)
         }
         return GoalLock{Type: LockNone}, nil
     }
     ```
   - Modify `UpdateProgress` symmetrically: try `ge.handlers[goalID]` first, fall through to the existing switch.

**Test impact:** none. All goal tests still call the unexported `progress*` and `checkLock*` functions directly; the switch still routes to them.

**Acceptance criteria:**

- [ ] `goal.Handler` interface defined.
- [ ] `GoalEngine.handlers` map exists; `Register` exists.
- [ ] `CheckLock` and `UpdateProgress` check registry first, fall through to switch.
- [ ] `go test ./internal/faction/engine/goal/...` is green.
- [ ] The two switch statements still carry every case they had before. (Verified by `grep -c "case \"G-" goal/goal.go` returning 13.)

**Commit message:** `refactor(goal): introduce Handler interface and parallel registry path`

### Phase 3.2 — Migrate all 13 handlers + tests; delete switches

**The big one.** Pull every progress/lock function into a per-goal struct in `goal/goals/`, migrate the tests, register everything, then delete the parallel switches and the now-empty helper files in `goal/`.

**Files to create (in `internal/faction/engine/goal/goals/`):**

One file per goal ID. Each file holds the handler struct + its `GoalID()`, `CheckLock`, `UpdateProgress` methods. Most goals have a no-op `CheckLock`; only G-004 and G-012 implement non-trivial locks.

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

**Standard handler shape** (illustrated for G-001; the other ten progress-only handlers follow identically):

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

**Helper migration:**

The following functions in `goal/progress.go` and `goal/lock.go` are used by multiple handlers and must move into `goal/goals/helpers.go` as unexported package-level functions in `package goals`:

- From `progress.go`: `completeGoal`, `countAssetKillsByCategory`, `findAsset`, `factionHasBaseOn`, `worldHasRivalPresence`, `rivalHasPlanetaryGovernmentOnWorld`, `factionHasPlanetaryGovernmentTag`
- From `lock.go`: `factionHasUnstealthedAssetOn`, `calcPlanetarySeizureXP`

All move verbatim. Imports get rebuilt for the new package context (`domain`, `state`, `rulebook`, `world` for the index-touching ones).

**Files to create:**

1. Twelve handler files (per the table above).
2. `goal/goals/helpers.go` — the 9 helpers above.
3. `goal/goals/register.go`:
   ```go
   package goals

   import "github.com/therobertcrocker/gm-toolkit/internal/faction/engine"

   func RegisterDefaultGoals(eng *engine.Engine) {
       eng.Goal.Register(MilitaryConquest{})
       eng.Goal.Register(CommercialExpansion{})
       eng.Goal.Register(IntelligenceCoup{})
       eng.Goal.Register(PlanetarySeizure{})
       eng.Goal.Register(ExpandInfluence{})
       eng.Goal.Register(BloodTheEnemy{})
       eng.Goal.Register(PeaceableKingdom{})
       eng.Goal.Register(DestroyTheFoe{})
       eng.Goal.Register(InsideEnemyTerritory{})
       eng.Goal.Register(InvincibleValor{})
       eng.Goal.Register(WealthOfWorlds{})
       eng.Goal.Register(ChangeHomeworld{})
   }
   ```

**Files to modify:**

4. `internal/faction/engine/goal/goal.go`:
   - Delete both parallel-path switches (the `case "G-012"` / `case "G-004"` block in `CheckLock` and the 11-case block in `UpdateProgress`).
   - Replace each with the data-only fallthrough:
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
     Same shape for `UpdateProgress`.
   - The file should no longer import anything it doesn't need. After the migration, the imports are `domain`, `state`, `rulebook`, `world` (still referenced in the `UpdateProgress` signature).

5. `internal/faction/engine/testharness/harness.go`:
   - Add import: `"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/goals"`
   - In `NewHarness` (after the other registrations): `goals.RegisterDefaultGoals(eng)`.

**Files to delete:**

6. `internal/faction/engine/goal/progress.go` — empty after migration.
7. `internal/faction/engine/goal/lock.go` — empty after migration. **Keep** the `LockType`, `LockNone`, `LockSkip`, `LockRestrictActions`, and `GoalLock` declarations — move them to `goal/goal.go` (or a new `goal/types.go`) so the public surface survives. The `goals/` handlers depend on these types.

**Test migration:**

The 19 existing goal tests live in `package goal` and call unexported functions directly. They need to move into `package goals` (the new sibling) and either:

- **Option A (recommended):** call the handler struct's method directly (`MilitaryConquest{}.UpdateProgress(...)`). Closest to the current shape; preserves the unit-test character.
- **Option B:** route through `*GoalEngine`. More integration-flavored.

Use Option A. File mapping:

| Old file | New file | What changes |
|---|---|---|
| `goal/progress_test.go` (9 tests) | `goal/goals/progress_test.go` | `package goals`; rewrite each call from `progressMilitaryConquest(acting, ...)` to `MilitaryConquest{}.UpdateProgress(acting, ...)`; types like `GoalLock` now reach in as `goal.GoalLock`. |
| `goal/lock_test.go` (4 tests) | `goal/goals/lock_test.go` | `package goals`; rewrite calls; reference `goal.LockType` etc. |
| `goal/world_index_test.go` (6 tests) | `goal/goals/world_index_test.go` | `package goals`; same pattern. |

Test helpers like `makeProgressRulebook` (referenced in `progress_test.go:124`) also need to move into the `goals` package (or `goals_test` if they're only used in tests).

**Acceptance criteria:**

- [ ] `internal/faction/engine/goal/progress.go` and `lock.go` are deleted.
- [ ] Twelve handler files exist in `goal/goals/`, plus `helpers.go` and `register.go`.
- [ ] `goal/goal.go` contains no `switch` over goal IDs; the data-only comment exists at both fallthroughs.
- [ ] Tests have moved into `goal/goals/` and reference handler structs directly.
- [ ] `go test ./internal/faction/...` is green.
- [ ] `grep -rn "case \"G-" internal/faction/engine/goal/` returns no results.
- [ ] Adding a new entry to `goals.toml` without a corresponding Go handler does not break any test (manual smoke-check: add a dummy `G-099` to a test rulebook fixture, confirm `CheckLock`/`UpdateProgress` return `nil` cleanly).

**Commit message:** `refactor(goal): promote to Shape 2 data-mirroring; migrate 12 handlers to goal/goals/`

<br/>

---

## Effort 4 — History demotion into `turn/`

**Why:** `HistoryEngine` is an empty struct holding a single method that writes one JSON line. It's ceremony around `os.OpenFile` + `json.Marshal`. Demoting it to a package-level function in `turn/` reflects that the record is conceptually a turn artifact and removes one over-structured sub-engine from the composition root.

### Phase 4.1 — Move `Record` into `turn/`, delete `engine/history/`

**Files to create:**

1. `internal/faction/engine/turn/history.go`:
   - Move the body of `history.HistoryEngine.Record` (currently `engine/history/history.go:18–39`) into a package-level function:
     ```go
     package turn

     import (
         "encoding/json"
         "fmt"
         "os"

         "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
         "github.com/therobertcrocker/gm-toolkit/internal/faction/state"
     )

     // RecordHistory builds an EventRecord from factionState, faction, and
     // mutations, then appends it as a single JSON line to historyPath.
     func RecordHistory(historyPath string, factionState *state.FactionState, faction *domain.Faction, mutations []domain.Mutation) error {
         event, err := buildEventRecord(factionState, faction, mutations)
         if err != nil {
             return fmt.Errorf("building event record: %w", err)
         }

         file, err := os.OpenFile(historyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
         if err != nil {
             return fmt.Errorf("opening history file: %w", err)
         }
         defer file.Close()

         data, err := json.Marshal(event)
         if err != nil {
             return fmt.Errorf("marshaling event: %w", err)
         }

         if _, err := fmt.Fprintf(file, "%s\n", data); err != nil {
             return fmt.Errorf("writing event: %w", err)
         }
         return nil
     }
     ```

2. `internal/faction/engine/turn/event_record.go`:
   - Move `buildEventRecord` from `engine/history/record.go:11–26` into this file (as unexported `buildEventRecord` in `package turn`).

3. `internal/faction/engine/turn/history_test.go`:
   - Move `TestHistoryEngine_Record` from `engine/history/history_test.go`. Rename to `TestRecordHistory`. Replace `he := New(); he.Record(...)` with `turn.RecordHistory(...)`. Package: `turn` (or `turn_test` if the existing test files use the black-box pattern — verify by reading any existing `turn/*_test.go`).

**Files to modify:**

4. `internal/faction/engine/core.go`:
   - Remove the `history` import (line 9).
   - Remove the `History *history.HistoryEngine` field from `Engine` (line 38).
   - Remove `e.History = history.New()` from `NewWithRulebook` (line 65).

5. `internal/faction/engine/orchestrator.go`:
   - In `applyAndRecord` (line 253), replace `e.History.Record(...)` with `turn.RecordHistory(...)`. The `turn` package is already imported (used elsewhere in the orchestrator); no new import needed. **Verify** the `turn` import exists at the top of the file before relying on it — if not, add it.

**Files to delete:**

6. `internal/faction/engine/history/` — entire directory: `history.go`, `record.go`, `history_test.go`.

**Acceptance criteria:**

- [ ] `internal/faction/engine/history/` no longer exists.
- [ ] `grep -rn "engine/history\|history\.HistoryEngine\|\.History\." internal/faction/ cmd/` returns no results.
- [ ] `*Engine` no longer has a `History` field.
- [ ] `turn.RecordHistory` is the only history-writing path.
- [ ] `go test ./internal/faction/...` is green; the migrated test passes.

**Commit message:** `refactor(history): demote to turn.RecordHistory; delete engine/history package`

<br/>

---

## Post-Initiative State

After all four efforts merge:

- **Nine sub-engines collapse to eight** (`history` is gone; its function lives in `turn`).
- **Every remaining sub-engine fits exactly one shape:**
  - Shape 1: `mutation`
  - Shape 2 open-ended: `action`, `hooks`, `ability`
  - Shape 2 data-mirroring: `tag`, `goal`
  - Shape 3: `turn`, `world`
- **Dependency direction is uniform:** every sub-engine package imports only `domain`, `state`, `rulebook`, `hooks`, and (in some cases) sibling sub-engines. Bootstrap files in `foo/foos/register.go` are the only files that import the parent `engine` package.
- **Bootstrap surface is uniform:** four `RegisterDefaultX(eng)` calls, all state-free, callable in any order at engine startup. The test harness invokes all four; a future CLI binary will do the same.
- **Discovery Open Questions 1–5 are all closed** in the Decisions Ratified table above.

No residual misalignments. The shape catalog in the discovery document remains the reference for future sub-engines — anything that doesn't fit one of the three should prompt re-examination, not be wedged in.

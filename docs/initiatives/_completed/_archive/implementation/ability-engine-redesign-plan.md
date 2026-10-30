# Ability Engine Redesign — Implementation Plan

> Discovery: [`docs/initiatives/discovery/ability-engine-redesign-discovery.md`](../discovery/ability-engine-redesign-discovery.md)
> Initiative: R-001 (`planned-work.md`)
> Branch: `feature/movement-redesign` (lands as wip commits alongside F-002 work; see *Branch context* below)

## Context / Goal

Retire the `internal/faction/engine/ability/` sub-engine and fold the residual A-flag domain into `UseAssetAbility` inside the action sub-engine, using an asset-ID-keyed handler map. Reshape `domain.AbilityDefinition` to a slim flat-parameter struct, port Informers' implementation into the new structure, stub the other 9 residual abilities to fall through to `ConfirmAbilityApplied`, correct three TOML flag mismatches, and update the Movement Redesign plan to record the Decision #6 reversal.

All six discovery questions are resolved (Q1 asset-ID handlers, Q2 slimmed `[assets.X.ability]` block, Q3 shared helpers under `action/actions/` *when needed*, Q4 Tripwire flag fix, Q5 structural-only, Q6 retire/reshape list ratified). This plan turns those resolutions into commits.

## Decisions Ratified in Planning

1. **Slimmed `domain.AbilityDefinition` is a flat optional-fields bag** (not a sum type, not a `map[string]any`). Concrete shape in *Shared Context → Struct shape*. Rationale: matches Q1 Option A (asset-ID dispatch, each handler reads what it needs); avoids re-introducing shape-typing that Q1 explicitly rejected; keeps `def.Ability.AttackerStat` as the natural read site for handlers.
2. **New dispatch package is `internal/faction/engine/action/actions/ability/`**, package name `ability`. The existing flat file `actions/use_asset_ability.go` remains as the `Action` interface impl and delegates into `ability.Dispatch(...)`. Rationale: clean Go naming (no underscores, single word); different import path from the old `internal/faction/engine/ability/` package so no collision; the old package is gone by the end of commit 4 anyway. The single file (`use_asset_ability.go`) that swaps from old-import to new-import does so atomically — no commit window where both are imported.
3. **The shared dice helper sub-subpackage is deferred to F-014**, not created in R-001. R-001 has zero callers (Informers is opposed-test; the 9 stubs call `ConfirmAbilityApplied`). Per YAGNI and [[feedback_yagni]], an empty `ability/dice/` package would be scaffolding without justification. Q3's resolution stands as the agreed *location for when needed* — F-014's first economy handler lands the helper.
4. **Commit cadence is additive-then-cutover-then-deletion.** Five refactor commits + one docs commit. Each wip commit must build clean. Specifically: domain reshape is additive (commit 1), TOML migrates next (commit 2), new dispatch lands while old engine still exists (commit 3), call site cutover and old-package deletion together (commit 4), legacy domain types removed once unused (commit 5), movement plan doc edit (commit 6). See *Work Breakdown*.
5. **`Engine.Ability` deletion includes its threading.** Commit 4 also deletes the `e.Ability` argument at `register.go:20` and the `e.Ability = ability.New()` line at `core.go:64`. `core.go`'s package-doc subpackage list also loses the `ability/` mention.
6. **`AbilityEffectType` constants survive the reshape.** `EffectRevealStealth` is read by the Informers handler; `EffectCoinDrain` / `EffectCoinSteal` are retained for use by future handlers (named effects in the TOML schema). Renaming or removing them is out of scope here.

## Open Questions — To Ratify at Implementation Time

| # | Question | Phase / Commit |
|---|---|---|
| 1 | **Outcome-table TOML shape.** The `Outcomes []AbilityOutcome` field on the slimmed struct anticipates Harvesters/Postech Industry/Venture Capital style mechanics, but R-001's only ported handler (Informers) doesn't use it. Concrete `[[ability.outcomes]]` row shape (`from`/`to`/`coin` vs `range`/`coin` vs other) ratifies when F-014 lands the first economy handler. R-001 declares the Go struct (see Shared Context) so the field exists; the TOML key shape isn't pinned. | Defer to F-014. |
| 2 | **`Engine.Ability` field deletion confirmation against `engine` package callers.** Discovery confirms 2 importers of `internal/faction/engine/ability/`. Execution Commit 4 should re-check via LSP `findReferences` on `engine.Engine.Ability` immediately before deletion in case another caller has landed since this plan. | Ratify at top of Commit 4. |

## Shared Context

### Struct shape — slimmed `domain.AbilityDefinition`

Commit 1 adds new fields additively. Commit 5 deletes `Steps`. Final shape:

```go
type AbilityDefinition struct {
    // Roll parameter (used by dice-based abilities — F-014's economy handlers)
    Die *DiceRoll

    // Outcome table (optional; populated by handlers that map roll → Coin delta)
    Outcomes []AbilityOutcome

    // Opposed-test parameters (used by Informers; extensible to Marketers)
    AttackerStat FactionStat
    DefenderStat FactionStat
    Effect       AbilityEffectType
}

type AbilityOutcome struct {
    From int  // inclusive lower bound on the roll
    To   int  // inclusive upper bound
    Coin int  // delta applied to faction Coin (may be negative)
}
```

All fields are zero-valued when unused. Each handler reads only what it cares about; no validation of field combinations is added (handlers are bespoke per Q1 Option A — invalid combinations are caller errors discovered at handler authoring time, not runtime guards).

`AbilityOutcome` lands additively in Commit 1 even though no R-001 handler reads it — keeps the F-014 follow-up free of domain-layer changes. (Mild YAGNI exception, justified by reducing F-014's commit footprint and avoiding a second sweep through `domain/asset.go`.)

### Package layout — `internal/faction/engine/action/actions/ability/`

```
action/actions/
├── use_asset_ability.go           # Action interface impl; calls into ability.Dispatch
└── ability/
    └── dispatch.go                # AbilityHandler type, handlers map, Dispatch func
    └── informers.go               # Informers' opposed-test handler
```

`dispatch.go` exports:

```go
package ability

type AbilityHandler func(
    faction *domain.Faction,
    asset *domain.Asset,
    def *domain.AssetDefinition,
    collector action.Collector,
    roller domain.Roller,
    factionState *state.FactionState,
    rulebook *rulebook.Rulebook,
) ([]domain.Mutation, error)

var handlers = map[string]AbilityHandler{
    "C1-002": informers,         // Informers (real handler)

    // Structural stubs — Q5 Option A. Each falls through to ConfirmAbilityApplied.
    // Real implementations land in F-014. Listed by TOML ID for grep clarity.
    "W1-002": confirmApplied,    // Harvesters
    "W3-001": confirmApplied,    // Postech Industry
    "W7-001": confirmApplied,    // Pretech Manufactory
    "F5-002": confirmApplied,    // Pretech Logistics
    "W6-001": confirmApplied,    // Venture Capital
    "W6-003": confirmApplied,    // Commodities Broker
    "C4-004": confirmApplied,    // Seditionists
    "W5-001": confirmApplied,    // Marketers
    "W4-002": confirmApplied,    // Monopoly
}

func Dispatch(/* same args as handler */) ([]domain.Mutation, error) {
    handler, ok := handlers[def.ID]
    if !ok {
        // Asset has A-flag but no registered handler — defensive; should not happen
        // post-flag-corrections. Return confirmApplied for parity with old fall-through.
        return confirmApplied(...)
    }
    return handler(...)
}
```

`confirmApplied` is a shared private helper inside the `ability` package that calls `collector.ConfirmAbilityApplied(asset, def)` and returns no mutations. Identical behavior to the old fall-through at `use_asset_ability.go:71-76`.

`informers.go` implements the existing logic that lives in `ability/steps/faction_check.go` today (opposed Cunning vs Cunning roll, on success emit reveal-stealth mutation). Commit 3 reads the current step handler and ports its body verbatim into the Informers handler — same collector calls, same mutation shape.

### Branch context

R-001 lands as wip commits on `feature/movement-redesign` per [[feedback_branch_initiative_layering]] and [[project_movement_branch_merge]]. No sub-branch is cut. The merge gate on `feature/movement-redesign` (no merge until all movement work is complete) means R-001 ships, F-002 Phase 4 is re-planned as a hooks handler in a separate plan session, F-002 Phase 4 lands, then the branch merges.

The branch must build at every wip commit so that movement-side execution sessions can resume mid-stream if needed — the commit cadence below enforces this.

### Per-commit model discipline

| Commit | Model | Notes |
|---|---|---|
| 1 — domain reshape (additive) | Sonnet | Mechanical add of optional fields |
| 2 — TOML migration + flag corrections | Sonnet | Data edits; no design |
| 3 — new dispatch + Informers port + 9 stubs | Sonnet, suggest Opus if Informers' state.FactionState reads expose any ambiguity | Mostly mechanical; the only design surface is the dispatch func signature, which is pinned in Shared Context above |
| 4 — wire UseAssetAbility through new dispatch; delete `ability/` package | Sonnet | Includes Engine.Ability field removal and register.go threading cleanup |
| 5 — drop AbilityStep types from domain | Sonnet | Strictly deletion after all callers gone |
| 6 — docs: record Decision #6 reversal | Sonnet (mid-session shift from execution per [[feedback_model_selection]]) | Doc-only edit to movement plan |
| Pre-merge checklist | Sonnet | Per `session-modes.md` section 8 |

## Out of Scope

- **Implementing the 9 stubbed handlers.** Tracked as F-014; each handler stays a `confirmApplied` entry until then. Per Q5 Option A.
- **F-002 Phase 4 transport implementation.** This plan only edits F-002's plan doc (commit 6) to note the Decision #6 reversal. F-002 re-planning as a hooks handler is a separate plan session after R-001 merges.
- **Creating `ability/dice/` helper subpackage.** Q3 establishes the location; first call site lands the package. Per Decision #3.
- **Renaming or pruning `AbilityEffectType` constants.** Per Decision #6.
- **Touching the hooks subsystem.** R-001's only interaction with hooks is documentary (commit 6 notes the transport-as-hook approach in F-002's plan).
- **Updating `cunning_assets.toml` Tripwire Cells description or counter.** Only the `flags` field changes in commit 2.

---

## Work Breakdown

Six commits. Each is one execution session.

### Commit 1 — `refactor: add slim parameter fields to AbilityDefinition`

##### Task 1 — `internal/faction/domain/asset.go`

Add new fields to `AbilityDefinition` *alongside* existing `Steps []AbilityStep`. Add new `AbilityOutcome` struct. Final state of struct after this commit (with `Steps` still present):

```go
type AbilityDefinition struct {
    Steps []AbilityStep    // legacy — removed in Commit 5

    Die          *DiceRoll
    Outcomes     []AbilityOutcome
    AttackerStat FactionStat
    DefenderStat FactionStat
    Effect       AbilityEffectType
}

type AbilityOutcome struct {
    From int
    To   int
    Coin int
}
```

##### Task 2 — Verify build

No other file changes. `go build ./...` must pass. No tests added (additive fields; no behavior change).

---

### Commit 2 — `chore: migrate Informers ability TOML and fix A/S flag mismatches`

##### Task 1 — `internal/faction/data/cunning_assets.toml`

Replace lines 41-45 (`[[assets.Informers.ability.steps]]` block) with:

```toml
[assets.Informers.ability]
attacker_stat = "Cunning"
defender_stat = "Cunning"
effect = "reveal_stealth"
```

Also at line 238: change Tripwire Cells `flags = ["A", "S"]` → `flags = ["S"]`. (Per Q4 / *TOML flag corrections required* in discovery.)

##### Task 2 — `internal/faction/data/wealth_assets.toml`

- Line ~368 (Pretech Manufactory): `flags = ["S"]` → `flags = ["A"]`.
- Find Monopoly's `flags` line (near line 207-219); confirm current value and change to `["A"]` if currently `["S"]`. If already `["A"]`, no edit needed — discovery flagged the discrepancy from the rulebook but didn't verify TOML state at that point. Execution session verifies before editing.

##### Task 3 — Verify build and Informers test path

`go build ./...` plus run `internal/faction/engine/action/actions/use_asset_ability_test.go` if it exercises Informers (verify). The Informers fall-through path through the *old* `ability.AbilityEngine.Run` still resolves: it reads the new `def.Ability.AttackerStat` etc. via the legacy `Steps` decoding only if the TOML still has `[[steps]]`. After this commit, Informers' `[[steps]]` is gone — old engine will see `def.Ability != nil` but `len(def.Ability.Steps) == 0` and emit no mutations.

**Note for execution**: This is the only commit in the cadence where Informers' behavior temporarily regresses (between commit 2 and commit 4). The old `ability.AbilityEngine.Run` returns zero mutations for Informers because `Steps` is empty after the TOML migration. Tests exercising Informers' reveal-stealth path will fail on commit 2 and pass again on commit 4. **Acceptable** because the branch isn't shipping mid-cadence — this is an intentional, short-lived broken-test window.

If broken Informers tests cause too much noise, an alternative is to reorder: commit 3 (new dispatch) before commit 2 (TOML migration). The execution session may choose to swap if Informers tests are loud — flag in dev journal.

---

### Commit 3 — `refactor: introduce per-asset ability dispatch under action/actions/ability`

##### Task 1 — `internal/faction/engine/action/actions/ability/dispatch.go` (new file)

Create the package with `AbilityHandler` type, the `handlers` map (asset-ID → handler), the `Dispatch` function, and the private `confirmApplied` helper. Exact shapes per *Shared Context → Package layout*.

##### Task 2 — `internal/faction/engine/action/actions/ability/informers.go` (new file)

Port the Informers-specific logic from `internal/faction/engine/ability/steps/faction_check.go`. The function signature matches `AbilityHandler`. Reads `def.Ability.AttackerStat`, `def.Ability.DefenderStat`, `def.Ability.Effect`. Calls the same collector method the current `faction_check.go` uses (re-grounding step: verify the collector method name against current source at execution time — `engine-collector-reshape` recently restructured collectors per [[feedback_branch_initiative_layering]]). Emits the same mutation shape on success.

##### Task 3 — `internal/faction/engine/action/actions/ability/dispatch_test.go` (new file, recommended)

Add a test that:
- Calls `Dispatch` for `C1-002` with a stub collector that records `RevealStealth`-style calls and a deterministic roller; asserts the handler resolves correctly.
- Calls `Dispatch` for one stub ID (e.g. `W1-002`) and asserts `ConfirmAbilityApplied` was invoked on the collector.

Test scope kept minimal — broader handler tests are F-014's territory.

##### Task 4 — Verify build

`go build ./...` must pass. The old `internal/faction/engine/ability/` package and `Engine.Ability` field are untouched. `UseAssetAbility` still routes through the old engine — the new dispatch is dormant until commit 4.

---

### Commit 4 — `refactor: route UseAssetAbility through new dispatch and retire ability sub-engine`

##### Task 0 — Re-grounding

Per session-modes section 8: verify `engine.Engine.Ability`'s callers haven't grown beyond `register.go:20` and `core.go:64`. LSP `findReferences` on the field. If new callers exist, pause and discuss before proceeding.

##### Task 1 — `internal/faction/engine/action/actions/use_asset_ability.go`

- Drop the `abilityEngine *ability.AbilityEngine` field and the `ability` import.
- Update `NewUseAssetAbility` signature: drop the third arg.
- In `Resolve`: replace the entire if/else (lines 71-82) with a single call to `ability.Dispatch(...)`. The handler routes Informers to its real impl and the 9 stubbed IDs to `confirmApplied` (which calls `ConfirmAbilityApplied` internally). The post-loop behavior (`u.mutations = append(...)`) is unchanged.
- Swap the import: drop `internal/faction/engine/ability`, add `internal/faction/engine/action/actions/ability`. Both bind to the local name `ability`, but only one is imported at any time within this file (atomic edit). No alias needed.

##### Task 2 — `internal/faction/engine/action/actions/register.go`

- Line 19-21: drop `e.Ability` from `NewUseAssetAbility` call site. New shape: `NewUseAssetAbility(c, e.Rand)`.

##### Task 3 — `internal/faction/engine/core.go`

- Line 38: remove `Ability *ability.AbilityEngine` field from `Engine` struct.
- Line 64: remove `e.Ability = ability.New()` line.
- Lines 6 (import) and 25 (package doc comment listing `ability/` as a sub-package): remove both.

##### Task 4 — Delete `internal/faction/engine/ability/`

Delete the entire directory:
- `ability.go`
- `steps/collector.go`
- `steps/faction_check.go`
- `steps/movement.go` *(was already movement-redesign-orphan; F-002 carried it forward as F-002 Phase 4; verify with `git log -- internal/faction/engine/ability/steps/movement.go` that no recent F-002 work is in flight here)*

##### Task 5 — `internal/faction/engine/action/actions/use_asset_ability_test.go`

Update test setup to drop the `abilityEngine` arg from `NewUseAssetAbility` calls. If the test exercised Informers' reveal-stealth through the old engine, route it through the new `useability.Dispatch` instead — likely just a no-op since the call site is the same; verify at execution time.

##### Task 6 — Verify

`go build ./...` and full `go test ./...` must pass. Informers' reveal-stealth path is now back online via the new dispatch. The 9 stubbed abilities continue to route to `ConfirmAbilityApplied` exactly as they did pre-redesign.

---

### Commit 5 — `refactor: drop AbilityStep types from domain`

##### Task 1 — `internal/faction/domain/asset.go`

After commit 4, no caller references `AbilityStep`, `AbilityStepType`, `AbilityStepMovement`, or `AbilityStepFactionTest`. Delete:

- Lines 41-46: `AbilityStepType` and its constants.
- Lines 56-68: `AbilityStep` struct.
- Line 71 (inside `AbilityDefinition`): the `Steps []AbilityStep` field.

`AbilityEffectType` and its constants (`EffectRevealStealth` / `EffectCoinDrain` / `EffectCoinSteal`) survive — `EffectRevealStealth` is read by the Informers handler.

##### Task 2 — Verify

`go build ./...` and `go test ./...` must pass. If any test fixture still references `AbilityStep` (e.g. rulebook test data), update it — these were the only callers and should already be gone after commit 4, but re-grounding via build is mandatory.

---

### Commit 6 — `docs: record Decision #6 reversal in movement redesign plan`

##### Task 1 — `docs/initiatives/implementation/asset-movement-redesign-plan.md`

In the *Decisions Ratified in Planning* table (lines 12-27), update Decisions #6 and #7:

- **Decision #6** — strikethrough or replace with: *Decision #6 reversed under R-001 — transport is not an ability step. It is an S-flag hook handler firing on `MovementOrderProgressed` / `MovementOrderCompleted` mutations. `AbilityStepTransport` is never introduced; `AbilityStepMovement` is deleted as part of R-001 (the `ability/` package retire). See R-001 plan for the new shape.*
- **Decision #7** — append note: *Reframed under R-001 — `Smugglers` and `Blockade Runner`'s cargo behavior moves to the S-flag hooks subsystem; the 1 Coin cost is the hook handler's payload, not an ability cost.*

Also update *Per-Effort Model Discipline* (line 105-106): Effort 2 Phase 4 row now references "hooks handler addition" instead of "transport handler design."

Add a paragraph under *Cross-Cutting Context* (around line 50) titled **Transport as hook (post-R-001)** with the F-002 re-planning prompt: *Effort 2 Phase 4 must be re-planned in a fresh plan session — the transport mechanic is now a `MutationReactor`-style hook handler against movement mutations, not an ability step. The phase's commit cadence and unit-of-work change; the original Phase 4 plan content is superseded.*

##### Task 2 — `docs/dev_journals/faction-manager/planned-work.md`

Update F-002's status note to remove the "blocked by R-001" line and add: *Phase 4 needs re-planning as a hooks handler; new plan session before execution resumes.*

R-001's row gets removed from *Planned Initiatives* on branch merge per CLAUDE.md rule 7-8 (pre-merge checklist), not in this commit.

##### Task 3 — Verify

No code changes. `git diff` reviews cleanly; no formatting drift.

---

## Pre-Merge Checklist (deferred to branch merge — not part of R-001 commits)

Per [[feedback_branch_initiative_layering]], the pre-merge checklist is branch-level, not initiative-level. R-001 doesn't run its own pre-merge — the `feature/movement-redesign` branch's pre-merge runs once all movement work plus R-001 plus F-002 Phase 4 are complete. R-001 contributes to that checklist by:

- Updating `decisions-log.md` with any execution-time decisions made during R-001's commits (per CLAUDE.md rule 7.1 and [[feedback_decisions_log_scope]]).
- Removing R-001's row from `planned-work.md`'s *Planned Initiatives* table at the branch's pre-merge moment.
- Senior-engineer review of R-001's diff as part of the branch-level review pass.

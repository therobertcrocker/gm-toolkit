# Event Hooks Subsystem — Implementation Plan

> **Final destination:** `docs/implementation/event-hooks.md`. Plan mode constrains edits to the harness path; this file moves to the repo location once the plan is approved.

## Context

Read: `docs/discovery/event-hooks.md` for the discovery process that led to this plan, including the open questions and the family of five interfaces.

The Core Engine ships with a single-interface stub (`EventHook`) and a documented dispatch site, but no registry, dispatcher, or consumers. Decisions #137 and #138 deferred this work because the Tag Engine wasn't ready and there was nothing to dispatch to. That deferral is now expiring: the next major effort is the TUI rebuild, which assumes a stable Core Engine collaboration surface, and reactive mechanics (tags + `S`-flagged asset effects) are the largest unbuilt piece of that surface.

A close reading of the SWN faction rules surfaced **five distinct shapes** of reactive mechanic, only one of which fits the existing stub. The discovery doc (`docs/discovery/event-hooks.md`) ratified the family of five interfaces, the per-turn budget machinery, and the litmus test: *expanding tags or `S`-flag effects must not require changes to `core_*.go` or `cmd/faction-manager/tui/`*.

This plan ships the **complete framework** — all five interfaces, registries, dispatchers, budget bookkeeping — plus **one proof-of-life hook per applicable category**. The Tag Engine ships in skeleton form (one package, three registered tags). The Effects Engine and the remaining tag handlers are deferred to follow-up branches.

## Settled Decisions (from discovery's open questions)

| # | Question | Resolution |
|---|---|---|
| 1 | Cat 1+2: one interface or two? | **Two.** Different shapes (offers vs. directives), different timings, different GM-input behavior. |
| 2 | `RuleModifier` package home — `hooks` or sibling? | **Same package** (`hooks`). Mechanics (typed registry, source-attributed dispatch) are identical to the other categories; semantic difference doesn't justify duplicating registry machinery. |
| 3 | Registration lifecycle — register-once at engine start? | **Yes.** Dynamic mid-turn registration deferred. |
| 4 | Persisted budget format — flat map or structured? | **Flat `map[string]int`** keyed by namespaced `BudgetKey` (e.g., `"tag:Warlike"`, `"asset:BookOfSecrets:C7-001"`). |
| 5 | Cat 1 proof-of-life — Warlike or Eugenics Cult? | **Warlike.** Asset-scoped registration is a registry concern, exercised end-to-end when the Effects Engine ships. Asset-scope path covered by a small unit test in this branch. |
| 6 | `MutationReactor` signature — keep `OnMutations` or enrich? | **Keep current signature.** Action context can be added via a segregable `ContextualMutationReactor` interface when a consumer needs it. |

## Settled Decisions (from this plan session)

| # | Question | Resolution |
|---|---|---|
| A | Where do `SelectModifiers` / `ConfirmReroll` live on `InputCollector`? | New `hooks.Collector` interface, embedded into `engine.InputCollector` alongside `action.Collector` and `ability.Collector`. |
| B | How do tags bind to hooks? | Mirror the asset → ability pattern (`engine/ability/ability.go:48-51`). `tags.toml` keeps current schema; Tag Engine looks up handlers by tag `id`. Future tags may add structured fields to `tags.toml` when patterns emerge — YAGNI for proof-of-life. |
| C | How do hook-eligible roll sites get richer dice state? | Introduce a new `RollWithHooks` function returning per-die slice + sum. Migrate hook-eligible call sites; leave non-hook sites (e.g., faction order) on the existing `DiceRoll.Roll(Roller) int`. |
| D | Test-fake collateral when `InputCollector` extends. | Refactor `attack_test.go` and `use_asset_ability_test.go` local fakes to use the gomock-generated `MockCollector` (decision #149). Lands in Phase 2. |

## Phase Overview

| # | Phase | Deliverable |
|---|---|---|
| 1 | Interfaces + registry foundation | Five interface families, scoped registries, no dispatch wiring |
| 2 | Budgets + `InputCollector` extension + test-fake refactor | `Faction.HookBudgets`, refill, two new collector methods, mocks regenerated, local fakes retired |
| 3a | Cat 3 dispatch + `EventHook` rename | `MutationReactor` dispatch wired into orchestrator; error returned on depth cap |
| 3b | Cat 1+2 dispatch infrastructure | `RollWithHooks` function, registry consultation logic; not yet wired into call sites |
| 3c | Cat 4+5 dispatch sites | `RuleModifier` family dispatched at each rule path; `TieResolver` dispatched at attack tie site |
| 4a | Tag Engine skeleton + Scavengers (Cat 3) | Package, registration entry point, one Cat 3 tag end-to-end |
| 4b | Warlike (Cat 1) + Fanatical (Cat 2 + Cat 5) + roll-site migration | Two tags wired through the new `RollWithHooks` path |
| 4c | `RuleModifier` proof-of-life (Preceptor or Pirates) | One Cat 4 tag end-to-end; litmus test demonstrated |

---

## Phase 1 — Interfaces + Registry Foundation

**Session deliverable:** Greenfield `hooks` package with all five interface families defined and per-category registries built. No dispatch wiring, no consumers. Compiles and passes registry unit tests.

### Tasks

1. Create the package directory `internal/faction/engine/hooks/`.
2. Add `hooks/types.go` defining shared types: `RollPhase` enum (`PhaseAttack`, `PhaseDefense`, `PhaseFactionTest`, `PhaseContested`), `RollContext` struct (Phase, Actor, Opponent, Attribute, Asset, World), `RollResult` struct (Dice []int, Modifier, Sum), `ModifierOffer` struct (Source, Description, BudgetKey, Apply func), `RerollDirective` struct (Source, Indices, BudgetKey, Elective), `TieOutcome` enum (`TieStandard`, `TieAttackerWins`, `TieDefenderWins`).
3. Add `hooks/roll_modifier.go` defining `RollModifier` interface with `OfferModifiers(ctx RollContext, factionState, rulebook) []ModifierOffer`.
4. Add `hooks/roll_result_hook.go` defining `RollResultHook` interface with `OnRollResult(ctx RollContext, result RollResult, factionState, rulebook) RerollDirective`.
5. Add `hooks/mutation_reactor.go` defining `MutationReactor` interface with `OnMutations(mutations, factionState, rulebook) []domain.Mutation`. (Replaces the soon-renamed `EventHook` — full rename happens in Phase 3a.)
6. Add `hooks/rule_modifier.go` defining the typed family: `AssetCostModifier`, `MaintenanceCostModifier`, `WorldTechLevelModifier`, `AssetMovementGranter`. Each interface method takes its specific question (buyer/asset/world/baseCost) and returns the modified answer.
7. Add `hooks/tie_resolver.go` defining `TieResolver` interface with `ResolveTie(ctx RollContext, factionState) TieOutcome`.
8. Add `hooks/scope.go` defining the `Scope` type: `Scope{Kind: ScopeFaction|ScopeAsset|ScopeGlobal, FactionID string, AssetInstanceID string}`.
9. Add `hooks/registry.go` defining `Registry` struct with one map per category, each keyed by `Scope`. Public methods: `RegisterRollModifier(scope Scope, source string, hook RollModifier)`, similar for each category. Internal lookup methods (used by dispatchers in later phases): `RollModifiersFor(faction, asset) []RegisteredRollModifier`, etc.
10. Add `RegisteredRollModifier` (and equivalents per category) wrapper struct carrying `Source string` + the hook itself, so dispatchers can attribute offers and decisions.
11. Document the package in a top-of-file comment in `hooks/registry.go`: who registers, who consults, when registration happens (engine startup), ordering guarantee (registration order).
12. Confirm the package compiles standalone (`go build ./internal/faction/engine/hooks/...`).

### Tests

13. Add `hooks/registry_test.go` covering:
    - Registration order is preserved on lookup (Cat 1, 2, 3, 4, 5 each).
    - Faction scope filters correctly (a hook scoped to faction A is not returned for faction B).
    - Asset scope filters correctly (a hook scoped to asset instance ID X is not returned for asset Y).
    - Global scope hooks are returned for any faction/asset query.
    - Empty registry returns empty slice (no nil-deref hazards).

### Files created

- `internal/faction/engine/hooks/types.go`
- `internal/faction/engine/hooks/roll_modifier.go`
- `internal/faction/engine/hooks/roll_result_hook.go`
- `internal/faction/engine/hooks/mutation_reactor.go`
- `internal/faction/engine/hooks/rule_modifier.go`
- `internal/faction/engine/hooks/tie_resolver.go`
- `internal/faction/engine/hooks/scope.go`
- `internal/faction/engine/hooks/registry.go`
- `internal/faction/engine/hooks/registry_test.go`

### Definition of done

- All interfaces compile.
- `go test ./internal/faction/engine/hooks/...` passes.
- No wiring yet to `core_engine.go` — `Engine` does not yet hold a `Hooks *hooks.Registry` field. That arrives in Phase 3a (when the first dispatcher needs it).

### Out of scope for this phase

- Any dispatch site changes in `core_*.go` or sub-engines.
- Any hook implementations.
- Wiring of `Registry` to `Engine`.

---

## Phase 2 — Budgets + `InputCollector` Extension + Test-Fake Refactor

**Session deliverable:** `Faction.HookBudgets` field added with TOML round-trip; refill wired into bookkeeping; `InputCollector` gains the two new methods via embedded `hooks.Collector`; existing local test fakes retired in favor of gomock.

### Tasks

1. Add `HookBudgets map[string]int` to `domain.Faction` at `internal/faction/domain/faction.go:43-59`. Add `toml:"hook_budgets,omitempty"` tag.
2. Initialize `HookBudgets` to `nil` (omitempty) at faction creation in the wizard / state load paths. Confirm round-trip with a unit test.
3. Define `hooks.Collector` interface in `internal/faction/engine/hooks/collector.go` with two methods:
    - `SelectModifiers(offers []ModifierOffer) []ModifierOffer`
    - `ConfirmReroll(directive RerollDirective) bool`
4. Embed `hooks.Collector` into `engine.InputCollector` at `internal/faction/engine/core_input_collector.go:12-17`, alongside `action.Collector` and `ability.Collector`.
5. Update `ScriptedCollector` at `internal/faction/engine/testharness/collector.go` to implement the two new methods. Default behaviour: take all offers, confirm all rerolls. Add scripting hooks (`SetModifierSelections`, `SetRerollConfirmations`) so tests can override.
6. Locate and update any other production `InputCollector` implementations (TUI collector if present in `cmd/faction-manager/tui/`). Add the two new methods.
7. **Refactor test fakes to gomock**:
    - Confirm gomock setup matches decision #149: identify the `MockCollector` source and how it's generated. If no `go:generate` directive exists, add one (likely on `core_input_collector.go`) or document the manual `mockgen` invocation.
    - Replace local `fakeCollector` struct in `internal/faction/engine/action/actions/attack_test.go:89-93` with `MockCollector` + `gomock.Controller`.
    - Replace `abilityFakeCollector` in `internal/faction/engine/action/actions/use_asset_ability_test.go:74-78` with `MockCollector`.
    - Run the affected tests; confirm all pass.
8. Add the budget-refill hook at the top of `Turn.ApplyBookkeeping` in `internal/faction/engine/turn/bookkeeping.go:26-52`. Refill semantics: clear the map (or zero out only declared budget keys — clear is simpler and acceptable since registered hook descriptors will repopulate on demand). Document the choice in a one-line comment with the **why** (matches "once per turn" rule wording).
9. Confirm `state.Save` and `state.Load` (in `internal/faction/state/faction_state.go`) round-trip the new field. Add an integration test that saves a faction with non-empty `HookBudgets`, loads it, and confirms equality.

### Tests

10. Unit test in `internal/faction/domain/faction_test.go` (create if missing): TOML round-trip of `HookBudgets` (nil → nil, populated → populated).
11. Unit test in `internal/faction/engine/turn/bookkeeping_test.go`: after `ApplyBookkeeping`, `HookBudgets` is empty regardless of pre-state.
12. Updated `attack_test.go` and `use_asset_ability_test.go` use `MockCollector` and pass.

### Files modified

- `internal/faction/domain/faction.go` (add field)
- `internal/faction/engine/core_input_collector.go` (embed new collector)
- `internal/faction/engine/testharness/collector.go` (implement new methods + scripting hooks)
- `internal/faction/engine/turn/bookkeeping.go` (refill hook at function top)
- `internal/faction/engine/action/actions/attack_test.go` (replace fake with mock)
- `internal/faction/engine/action/actions/use_asset_ability_test.go` (replace fake with mock)
- TUI collector implementation (if present) — add new methods
- Mock regeneration target (if a `go:generate` directive needs to be added)

### Files created

- `internal/faction/engine/hooks/collector.go`
- `internal/faction/domain/faction_test.go` (if not already present, for round-trip test)

### Definition of done

- `go test ./...` passes.
- `MockCollector` is the only collector mock used in unit tests; no local `fakeCollector` structs remain.
- A faction round-tripped through TOML preserves a non-empty `HookBudgets`.
- `ApplyBookkeeping` clears `HookBudgets` at turn start.
- Still no dispatch wiring — `Registry` still not held on `Engine`.

### Out of scope for this phase

- Dispatching budget consultation (lands in Phase 3b).
- Wiring Registry to Engine (lands in Phase 3a).

---

## Phase 3a — Cat 3 Dispatch + `EventHook` Rename

**Session deliverable:** `MutationReactor` dispatch site wired into the orchestrator via the new `dispatch/` package. Depth cap returns an error to the caller rather than dispatching to the observer. `EventHook` renamed everywhere. `Engine` now holds `Hooks *hooks.Registry`.

### Tasks

1. Add `Hooks *hooks.Registry` field to `Engine` at `internal/faction/engine/core_engine.go:31` (next to `Rand`).
2. Initialize `Hooks: hooks.NewRegistry()` in both `engine.New(dataDir)` and `engine.NewWithRulebook(rb)` constructors.
3. Delete `internal/faction/engine/core_event_hook.go`. The interface now lives in `hooks/mutation_reactor.go` (added in Phase 1).
4. Update the comment at `core_engine.go:22` (currently references `EventHook`) to reference `MutationReactor` and the `hooks` package.
5. Replace the dispatch-site comment block at `orchestrator.go:120-123` with a call to `dispatch.MutationReactors`. Algorithm (in `dispatch/mutations.go`):
    1. Take the `combined` mutation slice.
    2. Look up `MutationReactors` registered for the acting faction (faction-scoped) and global scope.
    3. Iterate in registration order. For each reactor, call `OnMutations(combined, factionState, rulebook)`.
    4. Append returned mutations to `combined`. If any new mutations were added, recurse.
    5. Track depth; bound at 5. On cap trip, return the accumulated mutations alongside an error — the orchestrator passes the error to `observer.OnError`.
6. `dispatch.MutationReactors` signature: `(registry, faction, combined, factionState, rulebook) ([]domain.Mutation, error)`. No observer dependency in the `dispatch` package.
7. Confirm no other production code references `EventHook` (search `grep -r "EventHook" --include="*.go"`).

### Tests

8. Unit test for `dispatchMutationReactors` in `core_dispatch_test.go`:
    - Empty registry → mutations passthrough unchanged.
    - One reactor adding one mutation → final slice has both.
    - Reactor adding mutation that triggers same reactor again → recurses (verify count of invocations).
    - Recursion depth cap: a reactor that always adds a mutation → stops at depth 5, returns a non-nil error.
    - Multiple reactors fire in registration order (verify with a recording reactor that appends source markers to mutations).
9. Integration test added to `internal/faction/engine/test_harness/integration_test/scenarios_test.go`: register a stub `MutationReactor` that emits a `CoinDelta` whenever it sees an `AssetDestroyed`. Run an attack that destroys an asset. Confirm the additional `CoinDelta` was applied. (This is a structural test — the actual Scavengers tag binding lands in Phase 4a, but the dispatch path is exercised here.)

### Files created

- `internal/faction/engine/hooks/dispatch/mutations.go`
- `internal/faction/engine/hooks/dispatch/mutations_test.go`

### Files modified

- `internal/faction/engine/engine.go` (add `Hooks` field, initialize in both constructors, update comment)
- `internal/faction/engine/orchestrator.go` (replace dispatch-site stub at line 120 with call to `dispatch.MutationReactors`)
- `internal/faction/engine/testharness/integration_test/scenarios_test.go` (add structural test)

### Files deleted

- `internal/faction/engine/core_event_hook.go`

### Definition of done

- `go test ./...` passes.
- No `EventHook` symbol exists anywhere in the repo.
- Stub-reactor integration test demonstrates Cat 3 end-to-end through the orchestrator.

### Out of scope

- Cat 1, 2, 4, 5 dispatch.
- Tag Engine.

---

## Phase 3b — Cat 1+2 Dispatch Infrastructure

**Session deliverable:** New `RollWithHooks` function that consults the registry for `RollModifier` and `RollResultHook` hooks, integrates `SelectModifiers` and `ConfirmReroll` collector calls, and decrements faction budgets. No call sites migrated yet.

### Tasks

1. In `dispatch/roll.go` (new file), implement `RollWithHooks`:
    ```
    func RollWithHooks(
        ctx RollContext,
        baseRoll domain.DiceRoll,
        registry *Registry,
        collector Collector,
        roller domain.Roller,
        faction *domain.Faction,
        factionState *state.FactionState,
        rulebook *rulebook.Rulebook,
    ) RollResult
    ```
2. Algorithm for `RollWithHooks`:
    1. Build the working `RollResult{Dice: nil, Modifier: baseRoll.Modifier, Sum: 0}`.
    2. Look up `RollModifier` hooks scoped to the faction (and global scope).
    3. For each, call `OfferModifiers(ctx, factionState, rulebook)` and collect all offers.
    4. Filter offers: an offer with non-empty `BudgetKey` is excluded if `faction.HookBudgets[BudgetKey] >= 1` (already used this turn). (Note: budgets are *consumed*, not capped — Phase 2's refill clears them, so any non-zero count means "used".)
    5. Pass the remaining offers to `collector.SelectModifiers(offers)`. The collector returns the chosen subset.
    6. For each chosen offer: invoke `offer.Apply(rollState)` (rollState is a thin wrapper over the working dice slice that lets `Apply` push additional dice), and increment `faction.HookBudgets[BudgetKey]` if non-empty.
    7. Roll the (possibly augmented) dice pool: `for i := 0; i < len(diceToRoll); i++ { result.Dice = append(result.Dice, roller.Roll(baseRoll.Sides)) }`.
    8. Compute `result.Sum = sum(result.Dice) + result.Modifier`.
    9. Look up `RollResultHook` hooks (faction + global scope). Iterate in registration order.
    10. For each, call `OnRollResult(ctx, result, factionState, rulebook)` to get a `RerollDirective`.
    11. Filter directives by budget the same way.
    12. For elective directives: call `collector.ConfirmReroll(directive)`. Skip if rejected.
    13. For accepted/non-elective directives: reroll the indicated indices using `roller.Roll(sides)`, update `result.Dice` and `result.Sum`, increment budget if `BudgetKey` non-empty.
    14. Return final `result`.
3. Define `RollState` helper in `hooks/types.go` with methods `AddDie(sides int)` (queues a new die for the upcoming roll) — to give `ModifierOffer.Apply` a clean mutation interface. Decide whether `Apply` mutates pre-roll dice count or runs after roll (pre-roll is simpler — Warlike's "+1d10 keep highest" expands the dice pool before any are rolled).

### Tests

4. Unit tests in `dispatch/roll_test.go`:
    - Empty registry → result matches plain `DiceRoll.Roll` semantics.
    - One `RollModifier` adding +1d10 → final `result.Dice` has one extra die.
    - Modifier with `BudgetKey` set: first call applies, increments budget; second call (same turn, same faction) skips because budget is non-zero.
    - `SelectModifiers` returning empty subset → no offers applied even if available.
    - `RollResultHook` returning `Indices: []int{0}` → die 0 is rerolled (verify with deterministic `FixedRoller`).
    - Elective reroll with `ConfirmReroll` returning false → reroll skipped.
    - Reroll with `BudgetKey` decrements budget.
    - Multiple modifiers from multiple sources fire in registration order; budgets attributed correctly per source.

### Files created

- `internal/faction/engine/hooks/dispatch/roll.go`
- `internal/faction/engine/hooks/dispatch/roll_test.go`

### Files modified

- `internal/faction/engine/hooks/types.go` (add `RollState`)

### Definition of done

- `RollWithHooks` exists and is unit-tested with stub hooks.
- No call sites migrated yet — production behaviour unchanged.
- `go test ./...` passes.

### Out of scope

- Migrating any of the 14 existing roll call sites (lands in Phase 4b).
- Cat 4/5 dispatch.

---

## Phase 3c — Cat 4 + Cat 5 Dispatch Sites

**Session deliverable:** Each `RuleModifier` family interface dispatched at its rule path. `TieResolver` dispatched at the attack tie site. All sites no-op when no hooks registered.

### Tasks

1. **Asset cost dispatch:** in `internal/faction/engine/action/actions/buy_asset.go` near line 76 (`ba.cost = def.Cost`):
    1. Replace the assignment with a call to a new helper `hooks.ResolveAssetCost(registry, buyer, def, world, baseCost)`.
    2. Helper iterates registered `AssetCostModifier`s for the faction, applies in registration order, returns final cost.
2. **Maintenance cost dispatch:** in `internal/faction/engine/turn/bookkeeping.go:86-90`:
    1. Update `maintenanceCost(asset)` to take a registry and faction: `maintenanceCost(registry, owner, asset) int`. Default body still returns 0 (decision #36 stub stands until structured cost data exists), but the body now calls `hooks.ResolveMaintenanceCost(registry, owner, asset, 0)` so registered modifiers can override.
    2. Update the call site at line 61 to pass the registry and owner.
3. **Tech-level dispatch:** add `hooks.ResolveWorldTechLevel(registry, faction, world, baseTL)` helper. Document at the top of `buy_asset.go` (where the line 17 deferred-TL comment lives) that this helper exists; no migration required this branch since TL filtering itself is deferred (decision in `buy_asset.go:17`).
4. **Movement granters:** add `hooks.GrantedMovementAbilities(registry, asset)` helper returning `[]MovementAbility`. Document for use when an action wants to know "what extra movement abilities does this asset have via tags/effects?". No migration required this branch.
5. **Tie resolver dispatch:** in `internal/faction/engine/action/actions/attack.go:120` (and `:154`):
    1. Before the `>=` comparison, call `hooks.ResolveTie(registry, ctx, factionState)`.
    2. If returned outcome is `TieStandard`, fall back to current `>=` behaviour.
    3. If `TieAttackerWins`, attacker wins ties unconditionally (current default — but explicit).
    4. If `TieDefenderWins`, defender wins ties (Fanatical's case).
    5. Apply equivalent logic at the counter-damage tie site at `:154`.
6. Ensure all dispatch helpers live in `dispatch/` (e.g., `dispatch/rules.go`), keeping consumer-facing call sites tidy.

### Tests

7. Unit tests in `dispatch/rules_test.go`:
    - `ResolveAssetCost` with no modifiers → returns base.
    - One modifier reducing cost by 1 → returns base-1.
    - Two modifiers chain in registration order (each sees the previous one's output).
    - `ResolveMaintenanceCost` similar.
    - `ResolveTie` with no resolver → returns `TieStandard`.
    - `ResolveTie` with resolver returning `TieDefenderWins` → returns that.
8. Integration test in `attack_test.go` (or new `attack_tie_test.go`): register a stub `TieResolver` returning `TieDefenderWins`. Roll equal attack and defense values via `FixedRoller`. Confirm defender's behaviour wins (no damage applied to defender; counter triggers if applicable).
9. Integration test in `buy_asset_test.go` for `ResolveAssetCost` with a stub modifier reducing cost by 1.

### Files created

- `internal/faction/engine/hooks/dispatch/rules.go`
- `internal/faction/engine/hooks/dispatch/rules_test.go`

### Files modified

- `internal/faction/engine/action/actions/buy_asset.go` (cost dispatch at line 76)
- `internal/faction/engine/turn/bookkeeping.go` (maintenance signature change at line 61, 86-90)
- `internal/faction/engine/action/actions/attack.go` (tie resolution at line 120, 154)
- Any tests touched by the maintenance signature change

### Definition of done

- All four `RuleModifier` family interfaces have helper-based dispatch.
- `TieResolver` dispatch fires at both tie sites in `attack.go`.
- `go test ./...` passes; existing behaviour preserved when no hooks registered.

### Out of scope

- Roll-site migration to `RollWithHooks` (Phase 4b).
- Tag Engine.

---

## Phase 4a — Tag Engine Skeleton + Scavengers (Cat 3)

**Session deliverable:** `tag_engine` package created. `RegisterDefaultTags(*Engine)` walks `rulebook.Tags` and binds known tag IDs to coded handlers. Scavengers (Cat 3, faction-scoped, no roll-site migration needed) lands as the first registered tag with end-to-end test coverage.

### Tasks

1. Create `internal/faction/engine/tag/` package directory.
2. Add `tag/tag.go` with:
    - `RegisterDefaultTags(eng *engine.Engine)` mirroring `RegisterDefaultActions` from `action/actions/register.go:8-23`.
    - Private dispatch table: `var handlers = map[string]func(*engine.Engine, *domain.Tag){...}` keyed by tag ID.
    - For each tag a faction owns (`faction.Tags`), look up by ID; if a handler exists, register the corresponding hooks scoped to that faction. Tags without handlers are silently no-op (data-only).
3. Wire registration into engine setup. **Decision point:** `RegisterDefaultTags` walks the *current* faction state to register faction-scoped hooks at engine start. This means `RegisterDefaultTags` takes the `FactionState`, not just the Engine. Adjust signature: `RegisterDefaultTags(eng *engine.Engine, factionState *state.FactionState)`. Call site lives wherever the existing `RegisterDefaultActions(e)` is called (currently `integration_test/harness_test.go:38`; confirm production wiring location and add there too — likely `cmd/faction-manager/` startup).
4. Add `tag/tags/scavengers.go`: implement `ScavengersReactor` as a `MutationReactor`. Logic:
    - Inspect each mutation in the slice; if any is `AssetDestroyed`, emit a `CoinDelta{FactionID: <owner of this tag>, Delta: +1}` per destroyed asset (own *or* rival, per rules p. 219).
    - Returns the new mutations to be appended (the dispatcher recurses, but since the new mutations are `CoinDelta` not `AssetDestroyed`, no infinite loop).
5. Register Scavengers in the `handlers` map keyed by the tag's ID (`"T-014"` per `tags.toml:93-97` — confirm exact ID).

### Tests

6. Unit test in `tag/tags/scavengers_test.go`: feed `ScavengersReactor.OnMutations` a mutation slice containing one `AssetDestroyed`; assert one `CoinDelta{Delta: +1}` returned.
7. Integration test in `integration_test/scenarios_test.go`: faction A owns Scavengers tag, attacks and destroys faction B's asset. Confirm A's coin balance increases by 1 (above any existing destruction-driven coin changes).
8. Negative test: faction without Scavengers in same scenario gets no bonus coin.

### Files created

- `internal/faction/engine/tag/tag.go`
- `internal/faction/engine/tag/tags/scavengers.go`
- `internal/faction/engine/tag/tags/scavengers_test.go`

### Files modified

- `internal/faction/engine/test_harness/integration_test/harness_test.go` (call `RegisterDefaultTags`)
- Production startup site (likely `cmd/faction-manager/cmd/<root>.go` or wherever `RegisterDefaultActions` is invoked) — confirm in-session and update.
- `internal/faction/engine/test_harness/integration_test/scenarios_test.go` (add Scavengers integration test)

### Definition of done

- `go test ./...` passes.
- A faction with the Scavengers tag gains +1 Coin per asset destroyed (own or rival) in any action.
- Tags without a registered handler in `tag_engine.handlers` continue to load as data with no behavioural effect.

### Out of scope

- Other tag categories (Phase 4b).
- `RuleModifier` tags (Phase 4c).

---

## Phase 4b — Warlike (Cat 1) + Fanatical (Cat 2 + Cat 5) + Roll-Site Migration

**Session deliverable:** Two more tags wired through `RollWithHooks`. The hook-eligible roll sites in `attack.go` (and any other site exercised by Warlike/Fanatical proof-of-life) are migrated to use `RollWithHooks`.

### Tasks

1. **Migrate hook-eligible roll sites in `attack.go`** at lines 116, 117, 121, 156:
    1. For each, build a `RollContext` (Phase: `PhaseAttack` / `PhaseDefense`, Actor, Opponent, Attribute from the action, Asset from the action target).
    2. Replace direct `roller.Roll(10)` and `DiceRoll.Roll(roller)` calls with `hooks.RollWithHooks(ctx, diceRoll, registry, collector, roller, faction, factionState, rulebook)`.
    3. The function returns `RollResult`; use `result.Sum` where the old code used the int return. Use `result.Dice` for any per-die logic (none currently in `attack.go`).
2. Migrate any *other* hook-eligible sites that proof-of-life consumers depend on. For Fanatical (auto-reroll 1s), per the rules this fires on *all* faction rolls — but for proof-of-life scope, just attack/defense rolls is acceptable. Decide in-session whether to expand to `expand_influence.go` and `ability.go` faction-test sites; if yes, migrate those four sites as well. (Other 7 sites that aren't faction tests stay on the simple `Roll` form.)
3. **Implement Warlike** in `tag/tags/warlike.go`:
    - `WarlikeRollModifier` implements `RollModifier`.
    - `OfferModifiers` returns one `ModifierOffer` when `ctx.Phase == PhaseAttack` AND `ctx.Attribute == FactionStatForce`. The offer's `Apply` adds 1d10 to the dice pool and instructs the result to keep the highest n dice.
    - `BudgetKey: "tag:Warlike"`, `Source: "tag:Warlike"`.
    - "Keep highest" semantics: since the existing `RollResult` returns `Sum = sum(Dice) + Modifier`, "keep highest" needs explicit handling. Decide: extend `ModifierOffer` with a post-roll trim hook, OR have the offer's `Apply` register a paired `RollResultHook` that drops the lowest die. Recommended approach in this plan: post-roll trim hook on `RollState` — simpler.
    - Document the design choice with a one-line comment.
4. Register Warlike handler in `tag/tag.go` `handlers` map keyed by Warlike's tag ID (confirm from `tags.toml:117-122` — likely `"T-016"` or similar).
5. **Implement Fanatical** in `tag/tags/fanatical.go`:
    - `FanaticalRollResultHook` implements `RollResultHook`. `OnRollResult` returns a `RerollDirective` listing indices in `result.Dice` where the value is 1, with `Elective: false` (auto-reroll, non-elective per rules) and no `BudgetKey` (unlimited).
    - `FanaticalTieResolver` implements `TieResolver`. `ResolveTie` returns `TieDefenderWins` when `ctx.Actor.HasTag("Fanatical")` and the actor is the attacker, OR `TieAttackerWins` when the Fanatical faction is the defender. (Per rules: "they always lose ties during attacks", which means: if Fanatical is attacker, attacker loses, so defender wins; if Fanatical is defender, defender loses, so attacker wins.)
6. Register both Fanatical hooks (one source, two interface registrations) in `tag_engine.handlers` under Fanatical's tag ID.

### Tests

7. Unit tests:
    - `warlike_test.go`: `OfferModifiers` returns offer when Force attack; empty otherwise (Cunning attack, defense, faction test).
    - `fanatical_test.go`: `OnRollResult` flags exactly the indices where `Dice[i] == 1`. Returns empty `RerollDirective` when no 1s.
    - `fanatical_test.go`: `ResolveTie` returns `TieDefenderWins` when Fanatical is attacker.
8. Integration tests in `integration_test/scenarios_test.go`:
    - Faction A has Warlike tag, attacks with Force-attribute asset. With `FixedRoller` set so the +1d10 keep-highest changes the outcome, confirm the bonus die was rolled and the higher result was used. Confirm `HookBudgets["tag:Warlike"]` increments to 1.
    - Faction A attacks twice in a single turn with Warlike: budget is consumed first call, second call gets no offer.
    - Faction A has Fanatical tag, attack roll dice include a 1. Confirm reroll occurred (verify final `result.Dice` no longer contains the original 1, with seeded `FixedRoller`).
    - Faction A has Fanatical, attacks with equal attack/defense rolls. Confirm defender wins (no damage to defender, counter logic fires if applicable).

### Files created

- `internal/faction/engine/tag/tags/warlike.go`
- `internal/faction/engine/tag/tags/warlike_test.go`
- `internal/faction/engine/tag/tags/fanatical.go`
- `internal/faction/engine/tag/tags/fanatical_test.go`

### Files modified

- `internal/faction/engine/action/actions/attack.go` (migrate roll sites at lines 116, 117, 121, 156)
- `internal/faction/engine/action/actions/expand_influence.go` (optional — decide in-session)
- `internal/faction/engine/ability/ability.go` (optional — decide in-session)
- `internal/faction/engine/tag/tag.go` (register Warlike + Fanatical)
- `internal/faction/engine/test_harness/integration_test/scenarios_test.go` (add Warlike + Fanatical scenarios)

### Definition of done

- `go test ./...` passes.
- Warlike fires on Force attacks, decrements budget, lapses on second attack within turn, refills next turn.
- Fanatical auto-rerolls 1s and forces tie loss when its owner is attacker.
- Hook-eligible attack/defense roll sites use `RollWithHooks`. Faction-order roll site (`turn/order.go:19`) deliberately stays on simple `Roll`.

### Out of scope

- `RuleModifier` tags (Phase 4c).
- Other tags from `tags.toml` (deferred follow-up).

---

## Phase 4c — `RuleModifier` Proof-of-Life

**Session deliverable:** Either Preceptor Archive or Pirates implemented as a `RuleModifier` consumer, demonstrating the Cat 4 path end-to-end.

### Tasks

1. **Decide in-session:** Preceptor Archive (`AssetCostModifier`: -1 Coin on TL4+ asset purchases) vs. Pirates (`AssetCostModifier` or movement-cost modifier: rivals pay +1 Coin to move onto Pirate-BoI worlds).
    - Tradeoff: Preceptor exercises the buy-asset cost path that's already plumbed in Phase 3c. Pirates would either also use the cost path (simpler) or exercise a different rule path (more code). Recommendation: pick whichever has a cleaner rule statement; default to **Preceptor** for path coverage simplicity.
2. **If Preceptor:** create `tag/tags/preceptor_archive.go`:
    - `PreceptorArchiveCostModifier` implements `AssetCostModifier`.
    - `ModifyAssetCost(buyer, def, world, baseCost) int` returns `baseCost - 1` when `def.TechLevel >= 4` and `baseCost > 0`; else returns `baseCost`.
    - Register in `tag_engine.handlers` for Preceptor Archive's tag ID (confirm from `tags.toml:75-79`).
3. **If Pirates:** create `tag/tags/pirates.go` with the equivalent logic for the movement/move-onto-world path. Note the Phase 3c helper for movement may need a small extension since this branch didn't migrate the movement cost site.

### Tests

4. Unit test in the chosen tag's `_test.go`: cost reduction fires on TL4+ assets, no-op on TL3 and below.
5. Integration test in `scenarios_test.go`: faction A has the chosen tag, buys a TL4+ asset (or moves onto a Pirate-BoI world, depending on choice). Confirm the modified cost was charged.
6. Negative test: same scenario, no tag, base cost charged.

### Files created

- `internal/faction/engine/tag/tags/preceptor_archive.go` *or* `pirates.go`
- Corresponding `_test.go`

### Files modified

- `internal/faction/engine/tag/tag.go` (register the chosen handler)
- `internal/faction/engine/test_harness/integration_test/scenarios_test.go` (add scenario)

### Definition of done

- `go test ./...` passes.
- Chosen tag fires at the chosen rule path; cost is modified end-to-end.
- Litmus-test demonstration complete: a new tag was added without touching `core_*.go` or `cmd/faction-manager/tui/`.

### Out of scope

- The non-chosen tag (Pirates if Preceptor was picked, etc.).
- All other unimplemented tags.
- Effects Engine.

---

## Deferred (explicitly out of scope this branch)

| Item | Why deferred | Re-entry trigger |
|---|---|---|
| Effects Engine — `S`-flag asset hooks (Boltholes, Tripwire, Cracked Comms, Blockade Fleets, etc.) | Discovery non-goal; large scope; foundation alone is the win | Effects Engine branch, post-TUI rebuild |
| Asset-scoped registration consumer | No proof-of-life consumer this branch (registry path exercised by unit test only) | Lands with Effects Engine |
| Dynamic registration lifecycle (mid-turn tag acquisition, asset destroyed mid-turn) | Foundation registers at engine start only; rules-driven tag changes happen between turns or at well-defined points anyway | First branch where a rule needs mid-turn tag acquisition (e.g., Change Homeworld → Deep Rooted re-targeting) |
| Remaining tag handlers (most of `tags.toml` — Plutocratic, Machiavellian, Theocratic, Imperialists, Savage, Deep Rooted, Eugenics Cult, Exchange Consulate, Perimeter Agency, Psychic Academy, Tripwire Cells, Lobbyists, Colonists, Technical Expertise, Capital Fleets, Mercenary Group, Secretive, Book of Secrets) | Beyond proof-of-life scope; many depend on `RuleModifier` paths or `S`-flag effects that don't yet exist | Subsequent branches, one or two tags per branch |
| Richer `MutationReactor` context (action name, actor, etc.) | Current `OnMutations` shape suffices for Scavengers; richer context can be added via segregable interface (`ContextualMutationReactor`) when a consumer needs it | First reactor that needs action context (likely Cracked Comms) |
| TL filtering enforcement | Already deferred (`buy_asset.go:17`); `RuleModifier.ResolveWorldTechLevel` helper exists in Phase 3c but no consumer migrated | Branch that enforces TL gates |
| Movement granters real consumer | Helper exists, no migration | Mercenary Group tag branch |
| Maintenance cost real values | Decision #36 stub; `RuleModifier` path plumbed but base cost still 0 | When `AssetDefinition` carries structured maintenance cost data |
| TUI affordances for new collector methods | `SelectModifiers` / `ConfirmReroll` get default scripted-collector implementations; the production TUI gains real prompts later | TUI rebuild branch |
| Budget rendering in TUI | Out-of-scope here; foundation persists budget data, TUI shows it later | TUI rebuild branch |

## Verification (cross-phase)

End-to-end checks to run after each phase:

- `go build ./...` from repo root — all phases.
- `go test ./...` from repo root — all phases.

## Critical Files Reference

| File | Phase touching it |
|---|---|
| `internal/faction/engine/core_event_hook.go` | 3a (deleted) |
| `internal/faction/engine/core_engine.go:22, 31` | 3a (Hooks field) |
| `internal/faction/engine/core_orchestrator.go:120-123` | 3a (dispatch site) |
| `internal/faction/engine/core_input_collector.go:12-17` | 2 (embed hooks.Collector) |
| `internal/faction/domain/faction.go:43-59` | 2 (HookBudgets field) |
| `internal/faction/state/faction_state.go:20-45` | 2 (round-trip verified) |
| `internal/faction/engine/turn/bookkeeping.go:26-90` | 2 (refill), 3c (maintenance signature) |
| `internal/faction/domain/roll.go:11-17` | (preserved as-is — non-hook sites still use this) |
| `internal/faction/engine/action/actions/attack.go:116-156` | 3c (tie), 4b (roll-site migration) |
| `internal/faction/engine/action/actions/buy_asset.go:76` | 3c (cost dispatch) |
| `internal/faction/engine/action/actions/use_asset_ability_test.go:74-78` | 2 (mock refactor) |
| `internal/faction/engine/action/actions/attack_test.go:89-93` | 2 (mock refactor) |
| `internal/faction/engine/action/actions/register.go:8-23` | 4a (mirror pattern for `RegisterDefaultTags`) |
| `internal/faction/engine/ability/ability.go:48-51` | (reference pattern only — `RegisterCustomHandler` is the model for tag binding by ID) |
| `internal/faction/data/tags.toml` | 4a, 4b, 4c (read tag IDs) |

# Sub-Engine Alignment — Discovery

Audits the nine packages under `internal/faction/engine/` against a shared structural pattern. Identifies three canonical sub-engine *shapes*, classifies each existing sub-engine against those shapes, and proposes per-engine changes to bring the system into alignment. Driven by the observation that `action` is the strongest-shaped sub-engine and a useful template for evaluating the rest.

---

## Problem

The faction engine composes nine sub-packages (`ability`, `action`, `goal`, `history`, `hooks`, `mutation`, `tag`, `turn`, `world`), each treated as a "sub-engine" by the core composition root. In practice these packages have drifted into inconsistent shapes: constructor patterns differ, some are open registries while others are closed switches, one (`tag`) inverts the dependency direction by importing the core engine, and the reference template (`action`) leaks its own `Collector` abstraction at one call site. There is no shared specification for what a "sub-engine" *is*, which makes new sub-engines easy to mis-shape and existing ones hard to compare. This discovery defines the shapes, audits current state, and proposes targeted alignment changes.

---

## Audit Findings

### Inventory

| Sub-engine | Primary purpose | Held on `*Engine`? | Invoked from orchestrator? |
|---|---|---|---|
| `action` | Decide available actions; drive selected action's lifecycle | yes | yes |
| `ability` | Resolve an asset's ability via per-step handlers | yes | indirectly (called by an action) |
| `goal` | Evaluate goal locks pre-turn; compute progress mutations post-action | yes | yes |
| `history` | Append event records to disk | yes | yes |
| `hooks` | Store + query scoped hooks across 8 categories | yes | yes (via `dispatch/` subpackage) |
| `mutation` | Apply mutations to in-memory state | yes | yes |
| `tag` | Walk faction tags and register hooks per tag | **no** | no (bootstrap-time only) |
| `turn` | Manage turn lifecycle (start, cursor, advance, bookkeeping) | yes | yes |
| `world` | Load spatial map; expose location index, distance | yes (conditional) | yes |

### Cross-axis comparison

The audit looked at four axes: *what does the sub-engine do, what does it need, how does it construct itself, how does the core engine use it.* Key divergences:

- **Constructor signature.** `turn.New(roller, rulebook)` takes deps; `world.New(dataDir)` returns an error; every other sub-engine is `New() *FooEngine` with no args. `hooks.NewRegistry()` uses different naming.
- **Construction side-effects.** `ability.New()` registers two built-in step handlers inline; every other constructor is bare.
- **Dispatch style.** `action` and `hooks` are open registries with external registration. `goal` and `mutation` are closed switches over hardcoded IDs/types. `ability` is mixed (custom handlers open, step handlers closed). `history` is a single I/O method.
- **Identity naming.** Most sub-engines use `<Pkg>Engine` (e.g. `ActionEngine`). `hooks.Registry` and `world.Engine` deviate.
- **Composition root presence.** Every sub-engine is held as a field on `*Engine` and constructed in `NewWithRulebook` — except `tag`, which is held nowhere and called separately via `RegisterDefaultTags(*engine.Engine, *state.FactionState)`.
- **Dependency direction.** `tag/` imports `engine` (the parent package) and reaches into `*engine.Engine` to call `eng.Hooks.Register*(...)`. Every other sub-engine imports only domain/state/rulebook and (in some cases) sibling sub-engines.

### Reference exemplar

`action` is the closest match to a clean shared pattern: bare `New()`, open `Register(factory)` extension point, factories that wire each action's own dependencies via closure, concrete implementations in a sibling `actions/` package, and a `RegisterDefaultActions(eng)` bootstrap function. One leak: `actions/register.go` performs `c.(engine.InputCollector)` when wiring `UseAssetAbility`, because the action's needs exceed `action.Collector`'s contract.

### Data Registry vs Handler Registry

A pattern the initial audit missed: the rulebook (`internal/faction/rulebook/`) loads `tags.toml` and `goals.toml` at startup and exposes `Rulebook.Tags map[string]*domain.Tag` and `Rulebook.Goals map[string]*domain.Goal`. This is a third kind of registry the sub-engines interact with — a **data registry** owning identity (which IDs exist) and presentation (name, description, text effect). The handler-side sub-engines (`tag`, `goal`) are **behavior registries** mapping a subset of those IDs to Go code.

Current coverage:

| | Data entries (TOML) | Go handlers | Posture |
|---|---|---|---|
| `tag` | 20 (T-001…T-020) | 4 (Scavengers, Warlike, Fanatical, PreceptorArchive) | Handler set is an explicit subset; missing handler = "data-only", documented in code comment |
| `goal` | 12 (G-001…G-012) | 11 in `UpdateProgress`, 2 in `CheckLock` | Handler set tracks data 1:1 by convention only; missing handler falls through to `return nil` without comment |

`tag` already operates in the data-mirroring posture; `goal` is structurally identical to a data-mirroring engine but treats the relationship as implicit. This reframes the open question about `goal`'s target shape (see *Open Questions*) — the data registry is already open, so the live question is whether the engine should reflect that openness.

---

## Sub-Engine Shape Specifications

Three canonical shapes emerge from the audit. Every sub-engine should fit exactly one.

<br/>

### Shape 1 — Stateless Dispatcher

**Use when** the variant set is sealed (compile-time exhaustiveness is desired), or the operation is a one-shot transformation with no extension points.

**Contract:**

```go
type FooEngine struct{}                              // empty
func New() *FooEngine                                // bare, no args, no errors, no side effects
func (e *FooEngine) Op(perCallState ...) returnType  // operates entirely on passed-in state
```

**Rules:**

- Empty (or near-empty) struct
- `New()` takes no arguments, returns no error, performs no registration or I/O
- All operations receive per-call state; the engine holds nothing between calls
- Dispatch is closed: a `switch` over a domain enum, a type-switch over a sealed interface, or a flat call
- **No registration mechanism, no sibling handler package**
- Adding a new variant requires editing the engine file — this is a *feature*, not a bug

**Reference exemplar:** `mutation` — closed type-switch over `domain.Mutation`, `default: panic(...)` to force exhaustiveness.

<br/>

### Shape 2 — Open Registry

**Use when** new variants are added by code, rulebook authors, or downstream campaigns; or the operation has many independent handlers that benefit from isolation.

**Contract:**

```go
// In package foo/:
type Foo interface {                               // the unit-of-work contract
    ID() string                                    // or some identity method
    // ...domain methods
}

type FooFactory func(deps ...) Foo                 // optional — if instances need request-scoped wiring

type FooEngine struct {
    handlers []FooFactory                          // or map[ID]Handler for keyed lookup
}

func New() *FooEngine                              // bare; registry starts empty
func (e *FooEngine) Register(handler FooFactory)   // open extension point
func (e *FooEngine) Lookup/Run/Dispatch(...)       // query + drive

// In sibling package foo/foos/:
type concreteFooA struct { /* deps */ }
func NewFooA(deps ...) *concreteFooA               // each concrete declares its own deps

func RegisterDefaultFoos(eng *engine.Engine)       // bootstrap wiring; called from core
```

**Rules:**

- Engine holds a registry (slice or map) of handlers
- `New()` is bare; built-ins are registered externally by a sibling-package bootstrap function — **never inside `New()`**
- Concrete handlers live in a sibling sub-package (`foo/foos/`)
- Each handler declares its own dependency surface; factories close over what each needs
- Operations are query (`Available...`) and/or drive (`Run`, `Dispatch`)

**Reference exemplar:** `action` — modulo the `c.(engine.InputCollector)` leak in `UseAssetAbility` wiring.

**Variants:**

Shape 2 splits into two sub-flavors based on who owns the ID universe:

- **Open-ended.** The handler registry itself defines what exists. There is no external authority on "which IDs are valid" — an ID is valid iff a handler is registered. Examples: `action`, `hooks`.
- **Data-mirroring.** A rulebook data file (TOML) is the authority on the ID universe. The handler set is a subset of that universe. A known ID with no registered handler is a legitimate "data-only" state, not a bug. Examples: `tag` (current), `goal` (target).

The structural contract is identical; the runtime semantics differ at the lookup site. Open-ended dispatchers fail or panic on unknown IDs; data-mirroring dispatchers must explicitly treat "unknown to code, known to data" as a designed state. Bootstrap functions in the data-mirroring variant register a subset of a known universe — they cannot extend the universe itself, only the behavior coverage of it.

<br/>

### Shape 3 — State-Owning Subsystem

**Use when** the engine owns heavyweight resources (loaded data, indexes, connections), or manages lifecycle state that spans multiple operations.

**Contract:**

```go
type FooEngine struct {
    dep1  SomeDependency
    dep2  OtherDependency
    state SomeOwnedState                           // mutable, lifecycle-managed
}

func New(dep1, dep2) (*FooEngine, error)           // takes deps; may return error (I/O, validation)
func NewWith{Resource}(resource) *FooEngine        // capability-injection alt constructor for tests
func (e *FooEngine) Op(...)                        // operates on owned state, possibly per-call state too
```

**Rules:**

- Constructor takes dependencies as args
- Constructor may return an error when construction involves I/O or validation
- Provide a capability-injection alt constructor so tests can inject stubs without going through the I/O path
- Struct fields persist across calls
- Operations may mutate owned state, but should still take per-call domain state as args — never store the domain state

**Reference exemplar:** `world` — `New(dataDir)` for production, `NewWithMap(map)` for tests, owns `spatialMap` + `Index`.

---

## Per-Engine Shape Assessment

| Sub-engine | Current shape | Target shape | Aligned? |
|---|---|---|---|
| `action` | Shape 2 open-ended (with collector leak) | Shape 2 open-ended | mostly |
| `ability` | Mixed Shape 1 / Shape 2 | Shape 2 open-ended *(verify whether step types are data-loaded during plan session)* | no |
| `goal` | Shape 1 (data-mirroring by accident) | Shape 2 data-mirroring | no |
| `history` | Shape 1 (over-structured) | Shape 1 *(or demote to free function)* | yes |
| `hooks` | Shape 2 open-ended (specialized) | Shape 2 open-ended | yes |
| `mutation` | Shape 1 | Shape 1 | yes |
| `tag` | Data-mirroring posture, mis-packaged | Shape 2 data-mirroring | no |
| `turn` | Shape 3 | Shape 3 | yes |
| `world` | Shape 3 | Shape 3 | yes |

The misaligned cases — `ability`, `goal`, `tag`, and the `action` leak — define the alignment work.

---

## Proposed: `tag` as a Proper Sub-Engine

Tag's job is to map a faction's data-driven tags to hook registrations. The ID universe is owned by `tags.toml` (loaded by the rulebook); the handler set is a subset, and tags without a handler are legitimately data-only. This fits **Shape 2, data-mirroring variant** cleanly. The current implementation already operates in this posture at the runtime-contract level — the refactor packages that posture into proper structural form and fixes the dependency inversion.

### Contract

```go
// package tag
type TagHandler interface {
    TagID() string
    Apply(faction *domain.Faction, hookRegistry *hooks.Registry)
}

type TagEngine struct {
    handlers map[string]TagHandler                   // keyed by TagID for O(1) lookup
}

func New() *TagEngine
func (e *TagEngine) Register(handler TagHandler)

// Walks factionState and invokes the matching handler per faction-tag.
// Faction-tags with no registered handler are silently skipped (data-only tags).
func (e *TagEngine) ApplyAll(factionState *state.FactionState, hookRegistry *hooks.Registry)
```

### Handler side

```go
// package tag/tags
type ScavengersHandler struct{}

func (ScavengersHandler) TagID() string { return "T-014" }

func (ScavengersHandler) Apply(faction *domain.Faction, hookRegistry *hooks.Registry) {
    hookRegistry.RegisterMutationReactor(
        hooks.FactionScope(faction.ID),
        "scavengers",
        &ScavengersReactor{FactionID: faction.ID},
    )
}

func RegisterDefaultTags(eng *engine.Engine) {
    eng.Tag.Register(ScavengersHandler{})
    eng.Tag.Register(WarlikeHandler{})
    eng.Tag.Register(FanaticalHandler{})
    eng.Tag.Register(PreceptorArchiveHandler{})
}
```

### Bootstrap flow

1. `engine.New(cfg)` builds core + sub-engines including `e.Tag` (held as `Tag *tag.TagEngine`)
2. `tags.RegisterDefaultTags(eng)` registers handler types — state-free, called once at startup
3. At faction-load time: `eng.Tag.ApplyAll(factionState, eng.Hooks)` applies the right handlers to the right factions

### What this fixes

| Current | Proposed |
|---|---|
| `tag/tag.go` imports `engine` (dependency inversion) | `tag/tag.go` imports only `domain`, `state`, `hooks` |
| `tag/tags/scavengers.go` imports `engine` | `tag/tags/scavengers.go` imports only `domain`, `hooks` |
| No `Tag` field on `*Engine`; invoked separately from bootstrap | `Tag *tag.TagEngine` on core, constructed in `NewWithRulebook` |
| `RegisterDefaultTags(eng, factionState)` couples bootstrap to state | `RegisterDefaultTags(eng)` is state-free; state coupling moves to `ApplyAll` |
| Tag handlers are free functions in a package-level map | Tag handlers are interface implementations — discoverable, testable, mockable |

---

## Alignment Plan — Per Engine

### `action` — close the collector leak

**Problem:** `actions/register.go` performs `c.(engine.InputCollector)` when wiring `UseAssetAbility`, because the action calls into `ability.Engine.Run(...)` which requires `ability.Collector` methods that `action.Collector` does not include.

**Proposed change:** widen `action.Collector` to compose `ability.Collector`:

```go
type Collector interface {
    hooks.Collector
    ability.Collector       // new
    // ...existing methods
}
```

The cost is that every action implementation has potential access to ability-collector methods even when it doesn't use them; the benefit is that the type assertion goes away and the contract reflects reality.

**Effort:** small. One interface change, drop the type assertion.

<br/>

### `ability` — finish the Shape 2 conversion

**Problem:** `ability.New()` registers `movementStepHandler` and `factionTestStepHandler` inline. Custom handlers are registered externally via `RegisterCustomHandler`. The internal asymmetry — one half open, one half closed — violates the Shape 2 contract.

**Proposed change:**

- Create sibling package `ability/steps/` containing `movementStepHandler` and `factionTestStepHandler` (one file per step type)
- Add `ability.RegisterStepHandler(stepType, handler)` as a public method
- Add `steps.RegisterDefaultSteps(eng *engine.Engine)` bootstrap function
- Strip the inline registrations out of `New()`
- Wire the call to `steps.RegisterDefaultSteps(eng)` from the core bootstrap path, alongside `actions.RegisterDefaultActions`

After: ability matches action's template exactly.

**Effort:** medium. Mechanical refactor, no behavior change.

<br/>

### `goal` — promote from Shape 1 to Shape 2 (data-mirroring)

**Problem:** `goal.go`, `lock.go`, and `progress.go` together hardcode 11 goal IDs across two switch statements. Adding a goal requires editing three files. No isolation between goals; no extension point.

**Proposed change:**

```go
// package goal
type Handler interface {
    GoalID() string
    CheckLock(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) (GoalLock, []domain.Mutation)
    UpdateProgress(faction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, rulebook *rulebook.Rulebook, index *world.Index) []domain.Mutation
}

type GoalEngine struct {
    handlers map[string]Handler
}

func New() *GoalEngine
func (e *GoalEngine) Register(handler Handler)
```

Each existing `progress*` / `checkLock*` pair moves into its own struct in a sibling `goal/goals/` package (one file per goal). `goals.RegisterDefaultGoals(eng *engine.Engine)` wires them.

**Effort:** large. Twelve goal handlers to extract.

**Motivation — parity with `tag`:** the rulebook already loads `goals.toml` as a data registry, just like `tags.toml`. A GM-authored goal added to TOML today has no Go behavior and the current switch silently `return nil`s — which is the data-only state, just undocumented. Promoting `goal` to Shape 2's data-mirroring variant brings it into structural parity with `tag`: the handler set is an explicit subset of the data set, missing-handler is a first-class state with documented semantics, and the lookup site treats "no handler for this ID" as designed rather than as a fall-through. The mechanical refactor is the same either way — what changes is that the design question is no longer "is the goal set sealed?" (the TOML answers that: no) but "should the engine reflect the openness of the data?".

<br/>

### `history` — leave alone, or demote to a free function

**Problem:** empty struct, single method, no state. Pure ceremony around `os.OpenFile` + JSON encoding.

**Options:**

- **Option A — leave as-is.** Cost is zero; aligns with `mutation`'s level of over-structure.
- **Option B — demote.** `package history; func Record(path, state, faction, mutations) error`. Drop the `History` field on core; orchestrator calls `history.Record(...)` directly.

Option B is more honest about what the package does. Option A is more uniform with the other sub-engines. Either is acceptable; flagging as a low-stakes call.

**Effort:** tiny either way.

<br/>

### `hooks` — leave alone

Already Shape 2, correctly named (`Registry` reflects the registry-not-engine identity), dispatch correctly factored into `hooks/dispatch/` subpackage. The 8-category interface family is justified by the dispatch differences between categories. No changes needed.

<br/>

### `mutation` — leave alone

Closed type-switch is correct. Mutations are a sealed event log; the `default: panic(...)` enforces exhaustiveness at runtime, and adding a new mutation type is *exactly the kind of change that should require an explicit edit here*. Promoting to Shape 2 would lose this property without a corresponding benefit.

<br/>

### `tag` — see *Proposed: `tag` as a Proper Sub-Engine*

**Effort:** medium. Refactor handlers into the new interface, add `TagEngine`, wire into core, update `testharness.RegisterTags` to call `eng.Tag.ApplyAll(...)`.

<br/>

### `turn` — leave alone

Shape 3, deliberately. Roller and rulebook are used in nearly every method; threading them per-call would be noise. Lifecycle state (cursor, phase) lives across calls by design.

<br/>

### `world` — minor cleanup

- Consider renaming `world.Engine` → `world.WorldEngine` for visual uniformity with other sub-engines. Low value; breaks import sites.
- The orchestrator's `if e.World != nil` guard reflects that `NewWithRulebook` does not construct a world (only the cfg-driven `New(cfg)` does). Either accept this asymmetry as documented, or make world required (with a stub fallback) so the guard goes away.

**Effort:** tiny. Cosmetic.

---

## Open Questions

| # | Question | Relevant area |
|---|---|---|
| 1 | Confirm: should `goal` adopt the data-mirroring posture that `tag` already has — handler set as an explicit subset of the rulebook data set, with missing-handler treated as a documented "data-only" state? The TOML registry is already open, so this is really asking whether the engine should reflect that openness or keep enforcing handler/data lockstep by convention. | `goal` |
| 2 | Should `action.Collector` compose `ability.Collector` directly, or should `UseAssetAbility` use a different mechanism to access the broader collector? | `action` |
| 3 | Should `history` keep its struct wrapper for uniformity, or demote to a free function for honesty about what the package does? | `history` |
| 4 | Should `world` be required (with a stub fallback when no data) so the `if e.World != nil` guard goes away, or stay optional and documented? | `world` |
| 5 | Should `world.Engine` rename to `WorldEngine` for naming uniformity, accepting the import-site churn? | `world` |

---

## Recommended Phasing

If the full alignment is undertaken, the natural ordering — chosen so each step is contained and behavior-preserving:

1. **`tag` refactor.** Highest architectural win (fixes the dependency inversion). Contained. Behavior unchanged.
2. **`ability` alignment.** Small mechanical follow-up that mirrors action's template.
3. **`action` collector fix.** Close the leak before any larger work depends on the action contract.
4. **`goal` decision and refactor.** Resolve Open Question 1. If promoting to Shape 2 (data-mirroring), the `tag` refactor (step 1) provides the structural template — handler interface, sibling package, bootstrap, and explicit data-only handling at the lookup site. If keeping the current shape, document the rationale in the decisions log.
5. **`history` demotion** *(optional)*. Resolve Open Question 3.
6. **`world` cleanup** *(optional)*. Resolve Open Questions 4 and 5.

Each step is a candidate for its own implementation plan and session.

---

## Notes

- This audit treats `tag`'s current shape as the most significant problem because it inverts the dependency direction between a sub-engine and the core engine. Other misalignments (`ability`'s internal asymmetry, `action`'s collector leak) are local; `tag`'s placement is architectural.
- The three shapes are descriptive of the current codebase, not normative beyond it. Future sub-engines that don't fit any of the three should prompt a re-examination of the shape catalog, not be wedged in.
- After the data-registry framing, `goal`'s target shape is no longer undecided — it's Shape 2 data-mirroring, matching `tag`. The remaining question is confirmation of the posture (Open Question 1), not selection between competing shapes.

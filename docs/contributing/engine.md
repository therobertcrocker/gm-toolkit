# Contributing — Engine

> How to extend the headless faction-turn engine. Read the
> [contributing overview](overview.md) first for setup and the shared extension
> pattern; this guide carries the engine's test tooling, the conventions you'll
> trip over, and six extension recipes.

## Orientation

The engine is a composition root (`engine.Engine`) owning a rulebook, a roller, a
hook registry, and seven sub-engines, driven by a phase-oriented orchestrator. It
is **headless**: it knows nothing about who drives it, talking to the outside
through a collector (decisions in) and an observer (events out). You extend it by
**registering handlers** that emit mutations and, where new behavior is reactive,
**registering hooks** the orchestrator dispatches at fixed pipeline points. Shape
& why: [engine overview](../architecture/engine/overview.md).

<br/>
<br/>

# Conventions that bite

Four rules that will cost you a debugging session if you miss them. Each links the
page that explains it — this list is the trap, not the rationale.

- **Emit, never mutate.** A sub-engine returns `[]domain.Mutation`; it never writes
  `FactionState` directly. `MutationEngine.Apply` is the single choke point that
  writes state. If you mutate in a handler, the change escapes history and replay.
  ([effect & mutation → Key Decisions](../architecture/engine/effect-mutation.md#key-decisions).)
- **Data-only TOML silently skips.** A tag/asset/goal ID with no registered handler
  is an intentional no-op, not an error — so **add the rulebook entry first**, or
  your handler is keyed against an ID that never matches and never runs.
  ([the extension pattern](overview.md#the-extension-pattern).)
- **Explicit deps dodge the engine import cycle.** `RegisterDefaultActions` and the
  registrars take their dependencies (roller, hook registry, world, rulebook) as
  *parameters*, so the `actions` package never imports `engine`. Reach for the
  parent engine and you reintroduce the cycle.
  ([actions → Key Decisions](../architecture/engine/actions.md#key-decisions).)
- **Recoverable sentinels live in the action package.** `ErrNoSelection`,
  `ErrTurnCanceled`, and `ErrActionUnavailable` are defined in `action`, not the
  TUI, so the error classifier can call a GM-cancelled action *recoverable* (skip,
  continue) without importing interface code. A new cancel path adds its sentinel
  here. ([logging & errors](../architecture/engine/logging-errors.md).)

<br/>
<br/>

# Test tooling

The engine is exercised end-to-end by the `testharness`
(`internal/faction/engine/testharness/`). `NewHarness(t)` scaffolds a temp
campaign, copies the bundled `rulebooks/swn/` rulebook, and builds a *live* engine
wired to a `StubSpatialMap` (every location resolves), a `ScriptedCollector` (you
queue the GM's decisions ahead of the run), and a `RecordingObserver`.

A scenario test follows one arc: build state with `AddFaction` / `AddBase` /
`AddAssetOnWorld`, script the collector, run a cycle through the engine, then
assert against the **recorded history** — `ReadHistory` reads `history.jsonl` back
into `EventRecord`s, and `AssertKinds` / `CountKind` / `FindMutationByCause` /
`FindMutationType` check that the right mutations landed for the right reasons.

Scenario tests live in `testharness/scenarios/` —
`full_cycle_test.go`, `hooks_test.go`, `tags_test.go`, `goal_lock_test.go`,
`transport_lifecycle_test.go`, and kin. The nearest-shaped one is your fixture
template: copy it, swap in your setup and assertions. The fixture rulebook is the
real `rulebooks/swn/` at the repo root, so your data ID must exist there (or in a
test rulebook you point the harness at) for the handler to fire.

<br/>
<br/>

# Recipes

### Add an action

> Shape & why: [actions](../architecture/engine/actions.md)

Add one more discretionary thing a faction can choose to do on its turn.

1. **`engine/action/actions/<name>.go`** — implement the four-method `Action`
   contract: `Validate` (cheap, pure eligibility — must not prompt), `Inputs`
   (gather GM decisions through the `Collector`), `Resolve` (compute the outcome),
   `Output` (return `[]domain.Mutation`). The split is loose about where the work
   sits; put it where it reads cleanest.
2. **`engine/action/actions/register.go`** — add a factory to
   `RegisterDefaultActions`, a closure building a fresh action bound to the turn's
   collector. Pull dependencies from the function's explicit parameters (roller
   resolver, hook registry, world, rulebook) — never the parent engine. Call
   `resolveRoller()` *inside* the closure, not at registration, or you defeat the
   deterministic-test seam (it's fence-signed in this file).
3. **`engine/action/collector.go`** — if the action needs GM input no existing
   method supplies, add the `Collector` method *signature* (and any order/option
   struct it exchanges) here. The interface side implements it.
4. **`engine/action/` (errors)** — if the action can be cancelled mid-prompt, add
   a recoverable sentinel beside `ErrNoSelection` / `ErrTurnCanceled` so the
   classifier treats the cancel as recoverable.
5. **`engine/action/actions/eligibility.go`** — if it's combat-shaped, reuse or add
   a shared targeting predicate (`eligibleAttackers`, `eligibleDefendersOnWorld`)
   rather than re-deriving eligibility inline.

**Test:** a `testharness` cycle that scripts the action and asserts its emitted
mutations; a unit test on `Validate` (the availability gate) and `Resolve` (the
outcome math).

**See also:** the GM-facing prompt for your new `Collector` method, the overlay,
and the `ImplementedActions` gate are the interface guide's
[Wire a prompt](interface.md#wire-a-prompt) recipe — the action isn't drivable
until that half lands.

<br/>

### Implement a tag handler

> Shape & why: [tags & effects](../architecture/engine/effect-mutation.md) ·
> the [hook category](../architecture/engine/hooks.md) you register into

Turn a faction tag from inert rulebook data into a live rule-bender.

1. **`engine/tag/tags/<name>.go`** — implement `tag.Handler`: `TagID()` returns the
   rulebook tag ID, `Apply(faction, hookRegistry)` registers the hooks the effect
   needs. Pick the hook categories deliberately — Warlike (Cat 1), Fanatical (Cat 2
   + Cat 5), Scavengers (Cat 3), and Preceptor Archive (Cat 4) are the four worked
   precedents, and a tag may register in any combination.
2. **`engine/tag/tag.go`** — add `e.Register(tags.<Name>Handler{})` in `New`, beside
   the existing four.
3. **rulebook `tags.toml`** — the tag ID must already exist there. A faction holds
   the tag as faction-state data; the behavior is your code.

**Test:** a `testharness` faction holding the tag; assert the registered hook fires
at its category's dispatch beat (e.g. a Cat 3 reactor's mutation appears in history
after the triggering phase).

<br/>

### Implement an asset-effect (`S`-flag) handler

> Shape & why: [tags & effects](../architecture/engine/effect-mutation.md) ·
> [hooks](../architecture/engine/hooks.md)

Give an asset passive/reactive specialness it carries just by existing. (The `S`
flag is the passive side; `A`-flag *active* abilities are the
[actions](../architecture/engine/actions.md) ability dispatch, not here.)

1. **`engine/effect/effects/<feature>.go`** — implement `effect.Handler`:
   `AssetDefinitionID()` and `Apply(faction, asset, hookRegistry)`, which registers
   the reactive hook for each owning asset. `transport.go` is the worked model — a
   handler that registers a `MutationReactor` per instance.
2. **`engine/core.go`** — in `NewWithRulebook`, add a registration arm to the
   rulebook loop, keyed on the asset definition's profile (the `def.Transport != nil`
   arm is the precedent): `e.Effect.Register(effects.New<Feature>Handler(def))`.
   This is where effect handlers get wired — `effect.New` itself starts empty,
   unlike the tag engine.
3. **rulebook asset TOML** — the asset definition carries the `S`-flag / profile the
   loop tests for.

**Test:** a `testharness` faction owning the asset; assert the reactor emits its
mutation on the trigger.

<br/>

### Implement a goal handler

> Shape & why: [goals](../architecture/engine/goals.md)

Make a goal do something to a faction's turn — constrain it, or advance on events.

1. **`engine/goal/goals/<goal>.go`** — implement the `Handler` contract: `GoalID()`,
   `CheckLock` (start-of-turn lock + time/condition-driven mutations),
   `UpdateProgress` (post-action, event-driven mutations). Implement whichever
   beat(s) the goal needs — Change Homeworld is pure `CheckLock`, Military Conquest
   pure `UpdateProgress`, Planetary Seizure both.
2. **`engine/goal/goal.go`** — register the handler in `GoalEngine.New`, beside the
   twelve standard ones.
3. **rulebook `goals.toml`** — the goal ID must exist there.

> `UpdateProgress` runs **before** `MutationEngine.Apply` — it inspects
> about-to-be-destroyed entities in live state. Count first, apply after, or the
> kills are uncountable.

**Test:** a multi-turn `testharness` run; assert progress increments, lock behavior
(`LockSkip` / `LockRestrictActions`), and the completion mutation bundle.

<br/>

### Add a mutation type

> Shape & why:
> [the mutation apply layer](../architecture/engine/effect-mutation.md)

Introduce one more kind of state change to the engine's vocabulary.

1. **`domain/mutation.go`** — define the struct (carry `FactionID`, the changed
   fields, a `Cause`, and `CausedByFactionID` where provenance matters) and its
   `Type()` discriminator — the stable string written to history.
2. **`engine/mutation/mutation.go`** — add an arm to `MutationEngine.Apply`'s type
   switch. The switch is **exhaustive by panic**: skip this and the first emission
   panics with `unhandled mutation type` rather than silently dropping.
3. **the emitting sub-engine** — return the new mutation from whichever engine owns
   that change.
4. **arch page** — add a row to the
   [Mutation catalogue](../architecture/engine/effect-mutation.md#mutation-catalogue)
   so the vocabulary list stays complete.

**Test:** an apply-layer unit test asserting the state change; a history-serialization
round-trip (the `Type()` string is the history key — a typo there is a silent
read-back failure).

<br/>

### Add a hook at a new dispatch site

> Shape & why: [hooks](../architecture/engine/hooks.md)

Consult a reactive mechanic at a *new* point in the pipeline. (Rarer and
framework-level — registering a reactor into an existing category is the tag/effect
recipes above; this adds the dispatch beat itself.)

1. **`engine/hooks/`** — if you need a new *category*, define its typed hook
   interface (the Cat 4 family is the model — one small interface per question, so
   the compiler keeps argument and return types honest). Otherwise reuse an existing
   category.
2. **the orchestrator or action site** — add the dispatch call at the new pipeline
   point, treating a `nil` registry and an empty match list as "no modification"
   (pass the base value through).
3. **honor the framework invariants** — only Cat 3 reactors recurse, bounded at
   depth 5; lookup precedence is global → faction → asset; reactors recurse on
   **newly emitted** mutations only, never the accumulated set.

**Test:** a `testharness` scenario asserting the hook is consulted at the new beat
(a registered hook's effect shows up; with no hook registered, behavior is
unchanged).

**See also:** the tag and effect recipes above *register reactors into* the
categories this recipe *dispatches* — the two halves meet at the registry.

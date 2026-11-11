# Orchestrator

> **Code:** `internal/faction/engine/{core,orchestrator,collector,observer,roller}.go`

## Purpose

The orchestrator is the composition root and turn driver for the faction
engine. `core.go` wires the `Engine` — the rulebook, a roller, a hook registry,
and seven sub-engines — and `orchestrator.go` drives one faction's turn through
a fixed sequence of phases, delegating the mechanics of each phase to a
sub-engine and owning everything between them: hook dispatch, mutation
application, history, persistence, and the pause points where a caller steps in.
If you are tracing how a turn actually runs end to end, this is the page.

## Shape

### The composition root

`Engine` (`core.go`) holds the rulebook plus the nine collaborators the
pipeline calls into:

```go
type Engine struct {
	Rulebook *rulebook.Rulebook
	Rand     dice.Roller
	Hooks    *hooks.Registry
	Turn     *turn.TurnEngine
	Tag      *tag.TagEngine
	Effect   *effect.EffectsEngine
	Mutation *mutation.MutationEngine
	Action   *action.ActionEngine
	Goal     *goal.GoalEngine
	World    *world.WorldEngine
	log      *slog.Logger
}
```

`NewWithRulebook` constructs every sub-engine in one pass, then performs the two
data-driven registration steps that turn rulebook content into runtime hooks:
it registers a `TransportHandler` for each asset definition carrying a
`Transport` profile, and calls `RegisterDefaultActions` with the explicit deps
the action factories need. The `World` engine is passed in rather than built
here — it is constructed against the campaign's spatial map before the faction
engine exists.

### Two cycle loops

`RunCycle` is the entry point. It logs a run header, primes the cycle's
data-driven hook layer once, up front — `Tag.ApplyAll` registers hooks keyed off
faction state, `Effect.ApplyAll` off rulebook asset data — then loops
`RunFactionTurn` until a turn reports the cycle closed out. The caller must have already called `Turn.Start` (or be
resuming an in-progress turn); the orchestrator does not open the cycle, it runs
it.

`RunFactionTurn` drives a single faction through the pipeline and returns
`(cycleDone bool, err error)`. The boolean reports whether advancing the turn
cursor past this faction closed the cycle — that is the signal `RunCycle` loops
on. The fixed phase order:

```mermaid
---
config:
  layout: elk
  theme: base
  themeVariables:
    background: '#1e1e2e'
    primaryTextColor: '#cdd6f4'
    lineColor: '#ff2e97'
---
flowchart LR
  Setup["setup"]:::process
  Stat["stat-raise"]:::process
  Pick["goal pick"]:::process
  Book["bookkeeping"]:::process
  Save1[("save")]:::muted
  Move["movement"]:::process
  Action["action"]:::process
  Save2[("save")]:::muted
  Finish["finish<br>save"]:::process

  Setup --> Stat --> Pick --> Book --> Save1 --> Move --> Action --> Save2 --> Finish

  classDef process  fill:#89b4fa,stroke:#5a7fd6,color:#1e1e2e
  classDef muted    fill:#45475a,stroke:#585b70,color:#cdd6f4
```

Each phase delegates its mechanics to a sub-engine, then runs the same
five-step beat around that call — zoomed in below:

```mermaid
---
config:
  theme: base
  themeVariables:
    background: '#1e1e2e'
    primaryTextColor: '#cdd6f4'
    lineColor: '#ff2e97'
---
flowchart LR
  Sub["① <br> call sub-engine<br>compute · emit mutations"]:::process
  Hooks["② <br> dispatch hooks<br>over emitted mutations"]:::cond
  Apply["③ <br> applyAndRecord<br>goal progress · Mutation.Apply · append history"]:::core
  Notify["④ <br> notify observer<br>On…(event)"]:::process
  Check["⑤ <br> await checkpoint<br>AwaitCheckpoint"]:::cond

  Sub --> Hooks --> Apply --> Notify --> Check

  classDef process fill:#89b4fa,stroke:#5a7fd6,color:#1e1e2e
  classDef core    fill:#cba6f7,stroke:#a47fd6,color:#1e1e2e
  classDef cond    fill:#89b4fa,stroke:#5a7fd6,color:#1e1e2e,stroke-dasharray:4 3
```

Setup rebuilds the world index and resolves the current faction. The goal pick
runs after stat-raise: a faction with no goal it can run, not destroyed, and not
resuming a turn whose bookkeeping already ran, gets `PhaseCollector.SelectGoal`
with the goals the goal engine offers; a nil answer declines, and the ask comes
back next turn. Every phase runs on to the action; nothing skips a turn. State is
saved at two interior gates (after bookkeeping, after action) and again in
finish — never mid-phase.


### The per-phase beat

Each phase function (`runGoalPickPhase`, `runBookkeepingPhase`, … ) is a
variation on one beat:

1. **Call the sub-engine** for this phase's mechanics (e.g.
   `Goal.Offers` and `Goal.Pick`, `Turn.ApplyBookkeeping`, `World.TickMovementOrders`,
   `Action.Run`). Sub-engines compute and *emit mutations*; they never touch
   state directly.
2. **Dispatch hooks** over the emitted mutations where the phase has reactive
   rules — `dispatch.MutationReactors` in the movement and action phases.
3. **Apply and record** via `applyAndRecord`: read the batch for every faction's
   goal and add the goal mutations after a reactor pass over them alone, apply
   the whole batch to in-memory state through `Mutation.Apply`, then append one
   `EventRecord` to the history file. It returns the batch as applied, which is
   what the phase passes to its observer.
4. **Notify the observer** of what happened (`OnBookkeepingApplied`,
   `OnMovementResolved`, … ).
5. **Await a checkpoint** where the phase is a GM pause point
   (`collectors.Phase.AwaitCheckpoint`).

This beat is a synthesis, not a literal invariant. Phases vary in both which
steps they have and their order: the goal pick has no hook dispatch and no
checkpoint; the action phase opens its batch with a record-only `ActionTaken` and
folds the `on_purchase` effects of every bought asset into the action's own
before dispatching. What is uniform is the recoverable-vs-fatal
handling wrapped around every fallible step (see Key Decisions).

### The two input contracts

`collector.go` defines the seam to whoever is driving the turn — GM, AI, or a
test:

```go
type Collectors struct {
	Phase  PhaseCollector  // orchestrator-owned prompts
	Action action.Collector // threaded to sub-engines
}
```

`PhaseCollector` is the orchestrator's own prompt surface: `ConfirmPayment` (embedded `hooks.PaymentCollector`; the turn-start lift), `RequestPermission` and `ConfirmPermissionContest` (embedded `hooks.PermissionCollector`; an arriving asset's permission request, and a rival Lobbyists' contest of the grant), `AwaitCheckpoint`,
`SelectAction`, `SelectStatRaise`, `SelectGoal`, `SelectMovementDecisions`,
`SelectTransportCargo`, `SelectRelocateOrders` (a free movement grant's prompt; it
shares the signature of the action collector's, so both satisfy
`relocate.Collector`). `action.Collector` is a separate, richer contract passed
down into the action sub-engine (it transitively satisfies `hooks.Collector` and
`ability.Collector`). The orchestrator threads both but calls only `Phase`
itself.

### The output channel

`observer.go` defines `TurnObserver`, the engine's fire-and-forget output
channel: one `On…` method per pipeline event. The engine never waits on the
observer and the observer cannot affect game state — it exists for display,
logging, and narration. Errors do **not** flow through the observer as the
control path: every phase function *returns* its error to `RunFactionTurn`,
which returns it to `RunCycle` and out to the caller. `OnError` is fired
alongside the return purely so observers can react; the caller's `err` is the
source of truth.

### applyAndRecord and the save cadence

`applyAndRecord` is the single choke point for state change: it reads goals over
a mutation batch, applies the batch with its goal mutations, and appends history,
and it deliberately does **not** persist to disk. `Mutation.Apply` reports per-mutation misses, which the orchestrator logs
before failing. `state.Save` is called only at phase gates — after bookkeeping,
after action resolution, and in `finish` — so a turn's on-disk state always
lands on a phase boundary, never inside one.

## Key Decisions

- **Single-pass, self-registering construction.** `NewWithRulebook` builds every
  sub-engine and registers its data-driven hooks in one call rather than a
  staged bootstrap. The engine owns its own wiring; callers hand it a rulebook
  and a world engine and get a ready engine back. (Resolves R-012; frozen log
  285–294.) Construction can fail: it is the first point that knows both the
  rulebook's `effects` data and the effect registrars, so it checks every
  entry there and `NewWithRulebook` returns an error on bad effect data.
  (s-flag-effect-registry plan, Planning Decision 7.)
- **Two input contracts, split by audience.** `PhaseCollector` is the
  orchestrator's own prompt surface (checkpoints, action/stat/movement/cargo/
  granted-relocate selection); `action.Collector` is a separate, richer contract threaded down
  into the action sub-engine, where it transitively satisfies `hooks.Collector`
  and `ability.Collector`. They are kept distinct because they serve different
  callers — the orchestrator drives `Phase`, the sub-engines consume `Action`.
  (Frozen log 46–53.)
- **Fire-and-forget observer; errors return.** `TurnObserver` is output-only and
  cannot change state; the control path for failure is the returned `error`.
  Internal packages return errors and the caller (here, the orchestrator)
  forwards them to the observer — the engine never depends on an observer for
  correctness. This is the error-return-over-observer rule.
- **Recoverable-vs-fatal branching at every fallible step.** Every sub-engine
  call, mutation apply, and checkpoint is wrapped in the same branch:
  `factionErrors.IsRecoverable(err)` → warn, fire `OnError`, and continue (or
  return `nil` to pause without aborting); otherwise log, fire `OnError`, and
  return the wrapped error. Recoverable failures degrade the turn; fatal ones
  abort it. (See [logging & errors](logging-errors.md) for the classifier.)
- **`applyAndRecord` does not save.** Mutation application and history append are
  decoupled from disk persistence. `state.Save` runs only at phase gates, so
  durable state always reflects a completed phase. (Frozen log 195–196.)
- **Checkpoints are the named pause points.** The `Checkpoint*` constants
  (`bookkeeping`, `movement`, `action_result`, `cycle_summary`)
  enumerate where the orchestrator yields to `AwaitCheckpoint`. A GM-driven
  caller blocks there for acknowledgement; an AI or test caller returns
  immediately. The pause points are part of the contract, not an interface
  artifact.
- **No goal-lock phase.** The phase ran only for the two process goals (Planetary
  Seizure, Change Homeworld). With those gone to the multi-turn actions of
  `F.040`, it had nothing to run, so it went with its `goal_locked` checkpoint and
  `OnGoalLockApplied`. A future lock lives on the action that holds it, not in a
  turn phase. Retires "goal-lock sits adjacent to the phase it gates". (A.001.8
  overview Decision 2.)
- **The goal pick runs after stat-raise and before bookkeeping.** The rules put
  the pick at the start of the turn, as the XP spend is; with the raise first, a
  Wealth raise this turn can open Wealth of Worlds for the same turn's pick. A
  resumed turn whose bookkeeping ran skips the pick, so a goal completed in
  bookkeeping is not replaced mid-turn. Rejected: the pick before stat-raise.
  (A.001.8 discovery D-2; Effort 3 Decision Record 7.)
- **Goals run inside `applyAndRecord`.** Every batch in every phase passes
  through it, so every faction's goal reads every batch by construction, and a
  batch added later cannot skip goals. The goal output gets its own reactor pass
  there, because a completion can land in a batch whose phase runs no reactors.
  Rejected: a goal call at each site. (A.001.8 discovery D-13; Effort 2
  Decisions 1 and 7.)
- **The orchestrator owns the inter-phase work, sub-engines own the mechanics.**
  Sub-engines emit mutations and never mutate state; hook dispatch, mutation
  application, history, persistence, and checkpoint handling all live in the
  orchestrator. This keeps mutations the only state-change currency and keeps
  sub-engines independently testable.

## Dependencies

**Depends on** every engine subsystem: [turn pipeline](turn-pipeline.md)
(cursor, bookkeeping, history), [actions](actions.md),
[goals](goals.md), [effect & mutation](effect-mutation.md) (apply layer + tag/
effect registration), [hooks](hooks.md) (dispatch), and
[world & movement](world-movement.md) (index rebuild, movement orders). It also
calls [persistence & static data](persistence.md) for `state.Save` and the
rulebook, and [logging & errors](logging-errors.md) for the run header and the
recoverable-error classifier.

**Depended on by** the [interface](../interface/overview.md) side: the
[event stream](../interface/event-stream.md) adapter constructs the `Collectors`
and a `TurnObserver`, runs the engine on a goroutine, and calls `RunCycle`. Any
future frontend (CLI, web) drives the engine through these same two seams.

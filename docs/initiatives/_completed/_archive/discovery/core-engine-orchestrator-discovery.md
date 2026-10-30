# Core Engine — Orchestrator Design

A design exploration of moving turn orchestration into the Core Engine and introducing a three-interface collaboration model.

<br/>

## Background

The Core Engine (`engine/core.go`) currently acts as a toolbox: it assembles the sub-engines at startup and exposes them for callers to use. The actual turn pipeline — the sequence of bookkeeping, action selection, resolution, mutation application, history recording, and state persistence — is driven by `TurnModel` in the TUI layer.

This creates two problems:

- The turn pipeline cannot run without the Bubbletea event loop. There is no way to execute a complete turn headlessly for testing, AI batch runs, or future non-TUI frontends.
- Adding a new action that needs mid-resolution prompts requires touching the TUI state machine, the event channels, and the pending message fields in `TurnModel` — with no shared template to follow. The pattern is re-derived each time.

The intended design is for the Core Engine to be the orchestrator — the "hand" that uses the sub-engines as tools — and for the TUI to be a collaborator that the Engine talks to, rather than a driver that tells the Engine what to do next.

<br/>

## Responsibilities

### Core Engine (orchestrator)

- Owns the full per-faction turn pipeline: goal lock check → bookkeeping → action selection → input collection → resolution → mutation application → history recording → state save → goal progress update
- Calls sub-engines in sequence; sub-engines do not call each other
- Communicates with callers through three interfaces (see below)
- Holds no knowledge of UI framework, terminal, or display concerns

### Sub-engines (tools)

Each sub-engine retains its current single responsibility. Nothing changes about how they work internally.

- **TurnEngine** — turn lifecycle, faction ordering, bookkeeping calculation
- **MutationEngine** — applies mutations to campaign state
- **ActionEngine** — validates and runs actions through their lifecycle
- **AbilityEngine** — resolves asset ability special effects
- **HistoryEngine** — appends event records to the JSONL history log
- **GoalEngine** — evaluates goal locks and advances goal progress

<br/>

## The Three Interfaces

### `InputCollector` (existing)

The Engine's input channel. Called when the Engine needs a decision it cannot make itself — which action to take, whether to redirect damage to a Base, which assets to send into a contested roll. Blocks until answered.

The caller (TUI, AI agent, test mock) implements this interface and answers in whatever way is appropriate. The Engine cannot distinguish one implementation from another.

**Direction:** Engine ← caller. No effect on game state.

<br/>

### `TurnObserver` (new)

The Engine's output channel. Called when the Engine has something to communicate — bookkeeping results, action outcomes, faction skips, cycle summaries. Fire-and-forget; the Engine does not wait for the observer to finish rendering.

The TUI implements this by sending messages to the Bubbletea event loop. A logger implements it by writing to a file. A headless test implements it by appending to a slice for later assertion. An observer can never affect game state.

**Direction:** Engine → caller. No effect on game state.

<br/>

### `EventHook` (new)

The Engine's reactive game mechanic channel. Called when a game event occurs — an asset is destroyed, an attack lands, bookkeeping completes. The hook returns a list of mutations the Engine applies immediately, as if they were part of the original action.

This is the natural hook point for the Tag Engine (future) and for asset special effects that trigger passively rather than being activated. A tag like "when this faction loses an asset, gain 1 Coin" is an `EventHook` that returns a `CoinDelta` when it sees the right event.

Multiple hooks can be registered. The Engine calls each in registration order and applies all returned mutations before continuing. Returned mutations are recorded in history alongside the mutations that triggered them.

**Direction:** Engine ↔ game logic. Can affect game state.

<br/>

## Interface Summary

| Interface | Direction | Affects game state | Implemented by |
|---|---|---|---|
| `InputCollector` | Engine ← caller | No | TUI, AI agent, test mock |
| `TurnObserver` | Engine → caller | No | TUI, logger, test spy |
| `EventHook` | Engine ↔ game logic | Yes | Tag Engine, mechanics hooks |

<br/>

## TUI Impact

With orchestration in the Engine, the TUI simplifies from a state machine that drives the pipeline to a display layer that reacts to it. The TUI implements both `InputCollector` (via the existing `TUICollector`) and `TurnObserver`.

The per-action goroutine/channel bridges remain — they are the correct pattern for mid-resolution input in an event-driven UI — but they generalise. The three structurally identical `*CompletedMsg` types consolidate into one. Most of the current `turnState` values collapse because the Engine drives sequencing, not the TUI.

<br/>

## Headless Execution

With the pipeline in the Engine, a complete faction turn can be exercised without a UI:

- A `testObserver` implements `TurnObserver` by recording events for assertion
- A `testCollector` implements `InputCollector` by returning pre-set values
- No `EventHook` is required for simple tests

This enables integration tests that cover the full turn pipeline — goal lock evaluation, bookkeeping, action resolution, mutation application, history recording, and state persistence — in a single test.

<br/>

## Open Questions

| # | Question | Relevant area |
|---|---|---|
| 1 | Does `TurnObserver` need acknowledgement signals for events where the GM must read output before proceeding (e.g. bookkeeping results)? Or does this belong in `InputCollector` as a dedicated method? | `TurnObserver` / `InputCollector` boundary |
| 2 | What is the ordering guarantee for multiple registered `EventHook` implementations? Registration order is simplest; alphabetical by hook name gives stability across restarts | `EventHook` |
| 3 | If a hook returns a mutation that itself triggers another hook (e.g. a tag gives Coin, and another tag reacts to Coin gain), does the Engine recurse or apply only one level deep? | `EventHook` |
| 4 | Does action selection move into `InputCollector` as a new method (`SelectAction`), or does the Engine call `ActionEngine.AvailableActions` and pass the result to the observer before blocking on the collector? | `InputCollector` / `TurnObserver` boundary |

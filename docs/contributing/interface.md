# Contributing — Interface

> How to extend the TUI projection over the engine. Read the
> [contributing overview](overview.md) first for setup and the shared extension
> pattern; this guide carries the interface's test tooling, the conventions you'll
> trip over, and three extension recipes.

## Orientation

The interface is a Bubble Tea program projected over the headless engine: a root
`Model` holding a mode bar and one sub-model per mode (`manage`, `turn`), with the
engine running on its own goroutine behind a two-channel [adapter](../architecture/interface/event-stream.md).
It is a **projection, never a player** — you never call the engine from a view.
You extend it by adding prompts, modes, and screens, and let messages carry the
GM's intent *down* and engine state *up*. Shape & why:
[interface overview](../architecture/interface/overview.md).

<br/>
<br/>

# Conventions that bite

Four rules that will cost you a debugging session if you miss them. Each links the
page that explains it — this list is the trap, not the rationale.

- **The Update Discipline.** The root *MAY* forward messages to the active sub,
  handle global keys, switch modes, and return `tea.Cmd`s for adapter I/O. It *MAY
  NOT* call the engine, mutate `*state.FactionState` or any `*domain.*` value, or
  embed game logic. Reach past that line and the engine stops being headless.
  ([state machine → Update Discipline](../architecture/interface/state-machine.md#the-update-discipline).)
- **Views render from snapshots, never engine pointers.** The UI runs on one
  goroutine, the engine on another; every value that crosses the bridge is a copy
  the UI owns outright. Hold a live `*domain.*` pointer in a view and you race the
  engine goroutine.
  ([event stream → Key Decisions](../architecture/interface/event-stream.md#key-decisions).)
- **Shared messages live in leaf `msgs` packages.** A sub-view signals its parent
  by emitting a type from a `msgs` leaf both import — never by importing the parent,
  which would be an import cycle.
  ([views → Key Decisions](../architecture/interface/views.md#key-decisions).)
- **lipgloss only inside `tui/`.** All styling flows through the `styles`
  sub-package (palette → tokens → semantic), with raw hex confined to `palette.go`;
  no engine package imports lipgloss. A view names a token, never a color.
  ([regions → Key Decisions](../architecture/interface/regions.md#key-decisions).)

<br/>
<br/>

# Test tooling

The interface has two test surfaces, and neither needs a terminal.

The **headless smoke / dry-run harness** (`tui/dryrun.go`'s `RunDryRun`, backed by
the `Smoke*` collectors and observer in `tui/adapter/smoke.go`) runs one full
`RunCycle` with no TUI: `SmokePhaseCollector` picks the first available action and
no-ops the rest, `SmokeActionCollector` returns the recoverable
`ErrActionUnavailable` for any prompt that shouldn't be reached, and `SmokeObserver`
logs every event to slog. It exercises the engine⇄collector contract end to end —
which is what makes the bridge's two halves independently verifiable without a
screen.

**Component unit tests** construct an overlay or a model directly and drive it
through its `Update` with `tea.Msg`s — `tea.KeyMsg` for keystrokes, the adapter's
ask/event messages for bridge behavior — then execute the returned `tea.Cmd` and
assert the message it emits: `OverlayDoneMsg.Answer` for an overlay, a transition
message for a router. `views/turn/overlay/movement_test.go` is the worked pattern.
(There is no `teatest` harness — the suite drives models by hand.)

<br/>
<br/>

# Recipes

### Wire a prompt

> Shape & why: [overlays](../architecture/interface/overlays.md) ·
> the [engine⇄TUI bridge](../architecture/interface/event-stream.md) the ask crosses

Give the GM a new modal prompt during a turn — and complete the interface half of
a GM-driven action.

1. **`tui/adapter/channels.go`** — add an `AskKind` constant at the *end* of the
   enum. It is **append-only**: the kinds are a positional wire contract, and a
   reorder silently re-maps them.
2. **`tui/adapter/action_collector.go`** (or **`phase_collector.go`**) — implement
   the `Collector` method the engine calls. Build the display payload *here*, copying
   every value out of engine ownership, then call the shared `ask` helper — it sends
   on the unbuffered `askCh`, parks the engine goroutine, and blocks on the reply.
   The view never sees an engine pointer.
3. **`tui/views/turn/overlay/<name>.go`** — implement the overlay. Reach for a
   generic archetype (`NewSelect` / `NewMultiSelect` / `NewConfirm`, `archetypes.go`)
   for a list or yes/no; hand-build a `huh` form when it needs more. Either way:
   watch for `huh.StateCompleted` in `Update` and emit `OverlayDoneMsg{Answer}`. The
   overlay never touches the reply channel — the execution view owns that send.
4. **`tui/views/turn/execution/execution.go`** — add an arm to the `newOverlay`
   switch (append-only; its `default` builds a labeled placeholder, so a missing arm
   degrades visibly instead of panicking) constructing your overlay from the ask
   payload plus the cycle number, rulebook, and spatial map the view supplies but the
   ask doesn't carry.
5. **`tui/views/turn/overlay/formhelp.go`** — pick the Esc keymap that matches what
   an *empty* answer means: `formHelp` (no empty answer exists), `declineFormHelp`
   (empty is a legal skip), or `cancelFormHelp` (empty aborts the action, emitting
   `action.ErrTurnCanceled` down the ordinary reply path).
6. **`tui/views/turn/overlay/registry.go`** — if this completes an action's
   drivability, add its name to `ImplementedActions`, so `SelectAction` stops
   rendering it with the `"(not yet available)"` suffix.

**Test:** the dry-run/smoke harness exercises the collector→ask contract end to end;
an overlay unit test feeds the widget `tea.KeyMsg`s through `Update`, runs the
returned `tea.Cmd`, and asserts the `OverlayDoneMsg.Answer` (the `movement_test.go`
pattern).

**See also:** the engine-side `Action` contract, factory registration, and the
`Collector` method *signature* are the engine guide's
[Add an action](engine.md#add-an-action) recipe — this recipe owns the interface
half (payload, overlay, Esc semantics, gate), and the action isn't drivable until
both land.

<br/>

### Enable a mode slot

> Shape & why: [root model & state machine](../architecture/interface/state-machine.md) ·
> [regions](../architecture/interface/regions.md) for the layout budget

Add a top-level mode beside Manage and Turn. (The disabled Spatial slot is the
worked precedent — a *registered absence*: already rendered, skipped on cycle, and
routed to a placeholder. Enabling it is the smallest version of this recipe.)

1. **`tui/views/modebar/modebar.go`** — add a `Mode` constant and a `slot{mode,
   disabled}` to the seeded slice `New` builds. Enabling an existing disabled slot
   (Spatial) is just flipping its `disabled` to `false`; `Next`/`Prev` skip disabled
   slots and wrap automatically.
2. **`tui/model.go`** — register the sub-model in the `subs map[modebar.Mode]tea.Model`
   that `NewModel` builds. The root forwards messages to it and calls `View()`; it
   never reaches inside.
3. **capability predicates** — implement `CapturesInput() bool` if the mode owns the
   keyboard while a form or text field is live, and/or `ModeLocked() bool` if it must
   forbid Tab-away (the turn view locks while its pumps run). The root discovers both
   by interface assertion — no root change, no `case` per mode.
4. **sizing & layout** — nothing to assign. `resizeSubs` already forwards the content
   budget to *every* registered sub as a synthetic `tea.WindowSizeMsg`, so a new mode
   is sized for free. If it wants the three-column working area, compose its content
   into `layout.Region`s via `layout.Compose` (`tui/layout/layout.go`) — but a
   `Region` is an intra-view column key (`Left`/`Center`/`Right`), **not** a per-mode
   handle.

**Test:** a root-model unit test that switches into the mode (drive `Update` with the
Tab `tea.KeyMsg`) and asserts key routing and the mode's lifecycle; if it reports
`ModeLocked`, assert Tab is refused while locked.

<br/>

### Add a manage sub-view

> Shape & why:
> [views — the per-mode composition pattern](../architecture/interface/views.md)

Add one more screen to the Manage mode's navigation graph
(list → detail → create → delete-confirm).

1. **`tui/views/manage/<name>/`** — build the sub-view as its own `tea.Model`; the
   `list` / `detail` / `wizard` siblings are the template. Copy the nearest-shaped one.
2. **`tui/views/manage/manage.go`** — add a `view` enum constant and a field for the
   sub-model, and handle its transition/completion message centrally in `Update`.
   Everything else falls through `routeForward` to the active sub-view — the mode
   model is a switchboard, the sub-views are the screens.
3. **`backTarget`** (`manage.go`) — add the new view's parent, so a `CancelMsg` routes
   back by map lookup rather than scattered branching.
4. **`tui/views/manage/msgs/`** — if the sub-view signals up (a request to open it, a
   completion result), add the message type to the leaf `msgs` package both the router
   and the sub-views import. That leaf is the import-cycle break.
5. **atomic CRUD** — if the sub-view mutates state, do *not* write from it: emit a
   completion message (the `CreatedMsg` / `DeletedMsg` shape) and let `manage.Update`
   perform the atomic `state.*` write (with in-memory rollback), rebuild the faction
   list, and surface a save error inline with state intact. Report `CapturesInput()`
   only if the sub-view has free-text fields.

**Test:** a manage-model unit test that drives navigation into the sub-view and
asserts the `backTarget` return path; if it writes, assert the list rebuilds and a
failed save renders inline with state untouched.

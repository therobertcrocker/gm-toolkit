# Interface — Overview

> The interface side's spine. Cross-subsystem narrative plus the page index.
> Anything about a single subsystem lives on that subsystem's page; only
> narrative that spans subsystems lives here.

## What the interface side is

The interface side is everything that *drives* the headless
[engine](../engine/overview.md). Today that is a single Bubble Tea TUI
(`internal/faction/tui/`). The side is named "interface" rather than "tui"
because the same seam — a collector in, an observer out — is meant to back a
rebuilt CLI and any future frontend, none of which the engine can tell apart.

This overview describes the TUI as it exists. As other frontends land, their
shared conventions live here and their specifics get their own pages.

<br/>

## Modes

The mode bar (`views/modebar`) defines the top-level interaction modes. Two are
live and one is a reserved slot:

- **Manage** — the review-and-browse surface for faction and asset state:
  faction summaries, asset lists, history. (This is the "Review" of the original
  design.)
- **Turn** — steps a faction through its turn cycle; the only mode that drives
  the engine.
- **Spatial** *(disabled)* — a reserved third slot, currently disabled in the
  bar (`{mode: ModeSpatial, disabled: true}`). Its eventual role isn't settled;
  the current direction is a faction-utilities / edit surface — actions you can
  perform on factions outside a turn (the original design's "Edit Mode").

The disabled slot is the growth point the root model is built around; the
[state machine](state-machine.md) page covers how a disabled mode is skipped in
the bar yet kept in the layout.

<br/>

## Composition story

A root `Model` (`tui/model.go`) holds the mode bar, a `map[Mode]tea.Model` of
sub-models (one per live mode), and the help + confirm-exit overlays. Each mode
*is* a sub-model satisfying `tea.Model`; the root frames the active sub with
chrome (mode bar, status line) and forwards messages to it.

When the GM enters **Turn** mode and starts a cycle, the engine runs on a
goroutine behind an **adapter** (`tui/adapter`). The adapter's `Run()` returns a
`tea.Cmd` that calls `engine.RunCycle`; the engine then talks to the UI through
two channels: an **unbuffered ask channel** that parks the engine whenever it
needs a GM decision (the collector calls block until the UI replies), and a
**buffered event channel** that carries observer events back without blocking
the engine. The [event stream](event-stream.md) page covers the bridge — pumps,
re-arm, teardown — in full.

<br/>

## Cross-subsystem conventions

Patterns that recur across interface subsystems, each verified in source:

- **The Update Discipline.** `tui/update.go` opens with a MAY / MAY-NOT header
  (cited from `tui-rebuild-arc-discovery.md`). The root `Update` MAY forward to
  the active sub, handle global keys, switch modes, and return adapter `tea.Cmd`s;
  it MAY NOT call engine methods, mutate state, do I/O outside a `tea.Cmd`, or
  embed game logic. Update is *a projection of engine state and a forwarder of
  user intent* — nothing more.
- **Capability predicates over type switches.** The root discovers what a
  sub-model can do by asserting small interfaces, not by switching on concrete
  type: `inputCapturer` (`CapturesInput()` — owns the keyboard, suppresses
  globals) and `modeLocker` (`ModeLocked()` — forbids mode-switch while a cycle
  runs, so its event/ask pumps are never orphaned).
- **Views hold display snapshots, never engine pointers.** Sub-models render
  from copied display payloads built by the adapter, not from live engine or
  domain values (`adapter/run.go`: *"the order is copied with display names so
  the execution view never holds engine pointers"*). This keeps rendering
  race-safe against the engine goroutine.
- **Message sub-packages break import cycles.** Each mode that needs shared
  message types puts them in a leaf `msgs` package (`views/manage/msgs`,
  `views/turn/msgs`) so sibling sub-views can exchange messages without a cycle.
- **lipgloss lives only inside `tui/`.** All styling goes through the `tui/styles`
  package; the engine and CLI layers carry no lipgloss dependency.

<br/>

## Page index

| Page | Scope | Status |
|------|-------|--------|
| [regions](regions.md) | The three-column Region/Panel layout system and root-owned budget | Written |
| [state machine](state-machine.md) | Root model, the Update Discipline, capability predicates, global keys | Written |
| [event stream](event-stream.md) | The engine⇄TUI bridge: goroutine, ask/event channels, pumps, teardown | Written |
| [overlays](overlays.md) | The `Overlay` interface and the prompt-overlay flow | Written |
| [views](views.md) | The per-mode composition pattern (manage + turn sub-views) | Written |

# Overlays — The Prompt Surface

> **Code:** `internal/faction/tui/views/turn/overlay/`

## Purpose

Overlays are the modal prompts the GM answers during a turn — pick an action,
choose attackers, confirm a reroll, name a Coin amount. Each is a small,
self-contained widget that renders a form and reports a single answer. The
defining constraint is that overlays know *nothing* about the engine bridge:
they never see the [ask channel](event-stream.md) or its reply, they just emit an
`OverlayDoneMsg{Answer}` and let the [execution view](views.md) put it on the
wire. That keeps every prompt a pure render-and-answer component, swappable and
testable in isolation.

## Shape

### The Overlay interface

```go
type Overlay interface {
	Init() tea.Cmd
	Update(tea.Msg) (Overlay, tea.Cmd)
	View() string
	Help() help.KeyMap
}

type OverlayDoneMsg struct{ Answer any }
```

It is a Bubble Tea model in miniature, with one twist: `Update` returns an
`Overlay`, not a `tea.Model`, so the host holds it in a typed field. An overlay's
*only* output is `OverlayDoneMsg{Answer}`, emitted via a `tea.Cmd` when its form
completes. The execution sub-model receives that message and forwards `Answer` on
the active ask's reply channel — *"overlays never touch the adapter's ask/Reply
plumbing… the execution sub-model owns the channel send"* (`overlay.go`). An
overlay couldn't reply if it wanted to; it has no reference to the channel.

### The acyclic import edge

Some richer overlays import plain *value* types from the adapter
(`RepairTarget`, `BribeReply`) to unpack their payloads. That edge is one-way and
deliberately so: **the adapter imports neither `overlay` nor `execution`**, so
`overlay → adapter` cannot form a cycle. Payload data flows up to the overlay;
control flows back down through the host, never overlay-to-adapter directly.

### Three archetypes, plus bespoke prompts

`archetypes.go` provides three generic widgets that cover the common shapes, each
a thin wrapper over a `huh` form themed with `styles.FormTheme()`:

- **`SelectOverlay[T]`** — single-select; a skip/decline sentinel is just an
  option whose value is the zero `T`.
- **`MultiSelectOverlay[T]`** — multi-select with an optional cap (validated).
- **`ConfirmOverlay`** — yes/no.

Prompts that need more than a list — `buyorder`, `refitorder`, `repairorders`,
`bribe`, `expand`, `movement`, `cargo`, `attack`, `selectaction`, the ability
prompts — are bespoke `huh` forms in their own files (~18 overlays in all), but
follow the identical pattern: build the form in the constructor, watch for
`huh.StateCompleted` in `Update`, emit `OverlayDoneMsg` with the collected value.

### Esc semantics are a per-overlay contract

What the *empty* answer means differs by prompt, and that difference is encoded in
which help keymap the overlay returns (`formhelp.go`):

| Keymap | Esc behavior | Used by | Why |
|--------|-------------|---------|-----|
| `formHelp` | Esc absent | checkpoint, stat-raise | no empty answer exists — decline is an explicit form option |
| `declineFormHelp` | Esc = skip | movement, cargo | empty is a *legal* answer (issue no orders) |
| `cancelFormHelp` | Esc = cancel action | the action overlays | empty means abort this action |

The action overlays share `cancelOnEsc`: on Esc it emits
`OverlayDoneMsg{Answer: action.ErrTurnCanceled}`. That sentinel rides the *same*
reply path as a normal answer; the adapter's `ask` helper turns it into a
returned error the orchestrator classifies **recoverable** — the faction's action
is skipped and the turn continues. Canceling an action is just answering with a
particular value, not a separate control path.

### The ImplementedActions gate

`registry.go` holds `ImplementedActions`, the set of action names the TUI can
currently drive. The `SelectAction` overlay lists every action the engine's
`Validate` admitted, but any name not in the set renders with a `"(not yet
available)"` suffix and is rejected by the form's `Validate` (huh has no native
disabled option, so the gate is label-suffix + reject). This decouples *engine*
action availability from *TUI* drivability, which let the per-action overlays land
one commit at a time; all twelve actions are now wired.

### Mounting: the ask→overlay dispatch

Overlays don't mount themselves. When the execution view receives a
`CollectorAskMsg`, its `newOverlay` method switches on the `AskKind` and
constructs the matching overlay from the ask payload, supplying the cycle number,
rulebook, and spatial map the overlay needs but the ask doesn't carry. The switch
is append-only alongside the `AskKind` enum, and its `default` arm builds a
labeled `Placeholder` — so an ask with no overlay yet degrades to a visible
placeholder rather than a panic. The overlay renders into the execution view's
fixed **wizard band** and, while mounted, makes the execution view report
`CapturesInput()` so the root routes every key to the form (see
[state machine](state-machine.md)).

## Key Decisions

- **Overlays never touch the reply channel.** A prompt's only output is
  `OverlayDoneMsg{Answer}`; the execution view owns the send onto the ask's reply
  channel. Overlays hold no adapter channel reference, which keeps them pure
  widgets and keeps all bridge plumbing in one place. (`overlay.go`; Decision 9.)
- **The overlay→adapter import edge is one-way.** Overlays may import adapter
  *value* types to read payloads; the adapter imports neither `overlay` nor
  `execution`, so no cycle can form. Data up, control down.
- **Esc-meaning is encoded in the help keymap.** The three keymaps
  (`formHelp` / `declineFormHelp` / `cancelFormHelp`) make each overlay declare
  whether empty is impossible, legal, or a cancel — and the cancel case routes a
  recoverable `ErrTurnCanceled` through the ordinary answer path rather than a
  special one. (Decision 2.)
- **Three archetypes cover the common shapes; the rest are bespoke.** Generic
  `Select`/`MultiSelect`/`Confirm` wrappers handle list and yes/no prompts;
  prompts needing more structure get hand-built `huh` forms following the same
  complete-and-emit contract. All themed through `styles.FormTheme` so overlays
  match the rest of the UI.
- **`ImplementedActions` decouples engine availability from TUI drivability.**
  The `SelectAction` overlay disables (label-suffix + `Validate`-reject) actions
  the TUI can't yet drive, which let per-action overlays ship incrementally
  without the engine knowing or caring which were wired.
- **`newOverlay` is the single mount point, with a placeholder default.** One
  switch arm per `AskKind`, append-only, defaulting to a labeled placeholder — a
  new ask kind shows a placeholder instead of crashing, so the bridge and the
  prompt surface can grow out of step safely.

## Dependencies

**Depends on** `huh` and `bubbles/help` (the form and keymap machinery), the
[regions](regions.md) `styles` vocabulary (`FormTheme`), and adapter *value*
types (`RepairTarget`, `BribeReply`, the payload structs) plus engine `action`/
`hooks`/`domain` types for the answers it builds. The `action.ErrTurnCanceled`
sentinel it emits on cancel is the [actions](../engine/actions.md) recoverable
contract.

**Depended on by** the execution [view](views.md), which is the overlay host: it
constructs overlays in `newOverlay`, mounts them in the wizard band, drives their
`Update`, and forwards `OverlayDoneMsg.Answer` onto the [event stream](event-stream.md)'s
reply channel. No other subsystem references the overlay package.

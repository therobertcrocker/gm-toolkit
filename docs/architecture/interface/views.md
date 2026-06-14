# Views — The Per-Mode Composition Pattern

> **Code:** `internal/faction/tui/views/manage/` (`list/`, `detail/`, `wizard/`, `deleteconfirm/`, `msgs/`), `internal/faction/tui/views/turn/` (`setup/`, `execution/`, `msgs/`)

## Purpose

A mode is not one screen but a small family of them — Manage is a list, a detail
card, a creation wizard, and a delete confirm; Turn is a setup roster and a
running-cycle view. This page covers the **composition pattern** both modes share:
each is a router model holding a `view` enum and one sub-model per screen, with a
message-driven `Update` that swaps the active screen and forwards everything else.
The pattern is uniform enough that learning one mode teaches the other; the
interesting differences are in what each routes *to*.

## Shape

### The router pattern

Every mode model has the same skeleton: a `view int` enum, a field per sub-view,
and an `Update` that splits into two kinds of handling.

```go
type view int
const ( viewList view = iota; viewDetail; viewCreate; viewDeleteConfirm )

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case msgs.RequestDetailMsg: /* build the sub-view, switch m.view */
	// … other transition/completion messages …
	}
	return m.routeForward(msg) // anything else → the active sub-view
}
```

Typed **transition/completion messages** are handled centrally (they change which
screen is active or commit a result); everything else falls through
`routeForward` to the currently active sub-view's `Update`. The mode model is a
switchboard, the sub-views are the screens, and they talk only in messages.

### Manage: a navigation graph as data

Manage routes over `list / detail / create / deleteConfirm`. Where it goes *back*
is a declared map rather than scattered logic:

```go
var backTarget = map[view]view{
	viewDetail:        viewList,
	viewCreate:        viewList,
	viewDeleteConfirm: viewDetail,
}
```

A `msgs.CancelMsg` sets `m.view = backTarget[m.view]`; a completed create returns
to `backTarget[viewCreate]`. The navigation graph is data you can read at a
glance, not control flow to trace.

CRUD is handled centrally and atomically. Sub-views never write state — the
wizard emits `msgs.CreatedMsg{Faction}`, the confirm emits `msgs.DeletedMsg{ID}`,
and `manage.Update` performs the `state.CreateFaction` / `state.DeleteFaction`
write (atomic with in-memory rollback — see [persistence](../engine/persistence.md)),
rebuilds the faction list, and returns to the parent view. A failed save stashes
`saveErr` and renders it inline with state untouched. Only the **create** view
reports `CapturesInput()`, because the wizard has free-text fields (name, Coin)
and its own field navigation that the root must not interrupt with global keys.

### Turn: two screens and the adapter lifecycle

Turn routes over just `setup` and `execution`. The `setup` view is a
`bubbles/list` roster of factions; pressing Start with at least one faction emits
`msgs.StartCycleMsg`. The turn router handles that by **building a fresh adapter
and execution model**, batching the adapter's `Run`/`ObserverPump`/`AskPump`
commands, and switching to the execution view. On clean completion
(`StreamClosedMsg`) or fatal recovery (`ReturnToSetupMsg`) it discards the adapter
(`m.adapter = nil`) and rebuilds setup with post-cycle HP/Coin. The adapter is
*built per cycle and discarded on return*, so no stale channels or goroutine
outlive a finished cycle (the lifecycle detail is on the
[event stream](event-stream.md) page; here it is one half of the turn router's
composition). While the execution view is active the router reports
`ModeLocked()`, protecting the in-flight pumps.

### The execution view: a three-band layout

`execution` is the most structured view — it renders a running cycle as three
regions composed through [layout](regions.md):

```
┌──────────┬───────────────────────────┬──────────────┐
│  rail    │  ▶ Faction · Phase        │  detail card │
│ (turn    │ ┌─ wizard band (overlay) ─┐│  (acting     │
│  order)  │ ├──────── divider ────────┤│   faction)   │
│          │ │   event stream          ││              │
└──────────┴───────────────────────────┴──────────────┘
```

The center column is a fixed-height **wizard band** over a scrolling **event
stream**. The band reserves a constant `wizardBandHeight = 14` rows — sized to the
tallest form, since huh caps a Select at 10 visible options — so the stream below
doesn't reflow each time an [overlay](overlays.md) mounts or unmounts. The band
hosts the active overlay; the stream appends a rendered line per
`ObserverEventMsg`; the left rail highlights the acting faction in the turn order;
the right card shows the acting faction.

### Race-safe display snapshots

The execution view runs on the UI goroutine while the engine runs on another, so
it never holds a live engine or domain pointer. It renders from copies:
`factionSnapshot` (the acting faction's display scalars, refreshed from each
`*domain.Faction` event) and `movableLine` (a resolved movement summary). Reads of
live pointers happen *only* inside `resolveMovables`, which runs while the engine
is parked on a movement ask — quiescent, so the read is safe. One subtlety: the
right card chooses its specialized layout off the **mounted modal**
(`activeAsk`), not off turn events, because the engine ticks in-flight orders and
fires `MovementTicked` *before* the movement ask — an event-derived phase would
already have run past movement by the time its modal opened.

### Message sub-packages

Each mode that needs shared message types puts them in a leaf `msgs` package
(`manage/msgs`, `turn/msgs`). `manage/msgs` carries completion messages
(`CreatedMsg`, `DeletedMsg`, `CancelMsg`, `SaveErrorMsg`) and transition requests
(`RequestDetailMsg`, `RequestCreateMsg`, `RequestDeleteMsg`); `turn/msgs` carries
`StartCycleMsg` and `ReturnToSetupMsg`. Because the messages live in a leaf both
the parent router and the sibling sub-views import, a sub-view can emit a message
the parent handles without the sub-view importing the parent — which would be a
cycle.

## Key Decisions

- **Every mode is a router over sub-views.** A `view` enum, one sub-model per
  screen, and an `Update` that handles transition/completion messages centrally
  while forwarding the rest to the active sub-view. The uniform shape is why
  Manage and Turn read the same despite doing very different work.
- **Manage's navigation graph is data.** `backTarget` declares each view's parent
  as a map, so cancel and completion route by lookup rather than scattered
  branches. The whole back-navigation is legible in three lines.
- **CRUD is central and atomic; sub-views never write.** Sub-views emit
  `CreatedMsg`/`DeletedMsg`; `manage.Update` performs the atomic
  `state.CreateFaction`/`DeleteFaction`, rebuilds the list, and surfaces a save
  error inline with state intact. Keeping the write in one place keeps the wizard
  a pure form. (Frozen log 274–284.)
- **The turn router owns the adapter lifecycle.** A fresh adapter is built per
  cycle and nilled on return to setup, so a completed cycle leaves no live
  goroutine or open channels behind. (`turn.go`.)
- **The execution view renders from snapshots, never engine pointers.**
  `factionSnapshot` and `movableLine` are copies built from events and ask
  payloads; live pointers are read only while the engine is parked. This makes
  rendering race-safe against the engine goroutine. (`execution.go`.)
- **The right card's phase keys off the mounted modal, not turn events.** The
  modal is the only reliable phase signal mid-cycle, because the engine emits
  `MovementTicked` before the movement ask — so an event-derived phase would lead
  the modal. (`execution.go`.)
- **The wizard band is a fixed height sized to huh's option cap.** Reserving a
  constant 14 rows for the overlay keeps the event stream from reflowing as
  overlays mount and unmount. (`execution.go`.)
- **Shared messages live in leaf `msgs` packages.** Sub-views and their parent
  both import the leaf, so a sub-view can signal the parent without importing it —
  the import-cycle break. (Frozen log 274–284.)

## Dependencies

**Depends on** [regions](regions.md) (`layout.Compose`/`RegionWidths`, the
`styles` vocabulary, `chrome.StatusLiner`), the [overlays](overlays.md) the
execution view hosts in its wizard band, the [event stream](event-stream.md)
adapter the turn router drives, and [persistence](../engine/persistence.md)
(`state.CreateFaction`/`DeleteFaction`) plus the `rulebook` and `spatial` map for
the detail and wizard views. It reads `domain` values to render.

**Depended on by** the [state machine](state-machine.md): the root `Model`
constructs `manage.New` and `turn.New` as its two mode sub-models, forwards
messages to the active one, and discovers their `CapturesInput`/`ModeLocked`/
`StatusLine` through the capability predicates. The views are the leaves of the
interface tree below the root.

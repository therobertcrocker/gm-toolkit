# TUI Rebuild — Arc Discovery

`F-005`.

## Problem

The toolkit currently has no interactive surface. Operating the engine means hand-running `main.go`; faction state is authored by hand-editing TOML. The TUI rebuild restores an interaction surface against the reshaped engine and serves three GM workflows:

1. **Faction management** — creating new factions, configuring them, and inspecting/editing existing ones. The TUI replaces hand-edited TOML across the entire authoring loop.
2. **Interactive turn execution** — driving the engine forward turn by turn during play; rendering structured mutations, narrative, and errors as they emit; advancing through phases under GM control.
3. **Direct state editing** _(stretch)_ — GM-side surgical edits to faction state outside the engine's mutation pipeline. Structurally distinct from (1) and (2): it bypasses the engine rather than consuming it. Out of MVP unless time permits.

The user is a single GM doing solo prep before a session and looking things up at the table. No multi-user, no remote, no concurrent-edit concerns. The TUI is a desk-side tool.

Two constraints shape the arc. **First**, the hard inheritance from V1's failure: the TUI is a pure engine consumer. It reads engine state and submits player intent through the orchestrator; it never owns game logic, never holds canonical state, never decides turn progression rules. This constraint binds every initiative in the arc. **Second**, the arc ships as a series of layered initiatives, each expanding the feature set without rebuilding the prior foundation. The first initiative's structural decisions — engine boundary, sub-model routing, adapter sub-package — must be complete enough to absorb later initiatives without overhaul, even though its feature surface is intentionally small.

## Scope of This Doc

This is an **arc-discovery** doc per `docs/process/session-modes.md` section 8: it ratifies cross-cutting architectural decisions that constrain every initiative in the arc — framework, boundaries, conventions, extensibility patterns. Slicing the work into specific initiatives is deferred to **Arc-Plan**. Where this doc references "the first initiative," it means the first thing that ships in the arc; per-initiative scope is Arc-Plan's call.

## Design Summary

The TUI is a **greenfield Bubbletea / lipgloss / huh application** built strictly as an engine consumer. Its call surface against the engine is exactly two methods (`RunFactionTurn`, `RunCycle`) and two interfaces (`PhaseCollector`, `TurnObserver`) — every prompt is the engine asking, every event is the engine telling. The TUI never owns canonical game state and never decides turn progression.

The Model/Update/View structure is held to **binding discipline**: Update may transform Model and emit `tea.Cmd`s but may not call into the engine synchronously, mutate engine state, or perform game logic. Engine work runs in a goroutine; collectors block on reply channels; observers pump callbacks onto Bubbletea's message queue. These channels and the collector / observer implementations are the **TUI-Engine adapter**, a self-contained sub-package isolating the boundary. Per-faction vs whole-cycle execution is a **cadence flag** the adapter consults at `AwaitCheckpoint` — not a separate mode — and is GM-flippable mid-cycle.

The UI is a **persistent mode bar** (`Manage | Turn | Spatial | Quit`) over mode-owned content. Default landing is Manage — a faction list with detail-on-Enter and edit as its own view. Turn opens a setup view (cadence + roster), then an execution layout with the processing faction's state on the left and a scrolling event log on the right; collector prompts surface as a single modal overlay component (a `huh` form for decisions, a key-listener for acks). Spatial is a greyed placeholder for `F-012`.

The **input grammar is letter actions plus arrow navigation**. `Tab` / `Shift-Tab` cycles modes, `?` opens a help overlay, `q` quits, and `Esc` resolves prompts with a cancel sentinel the adapter routes as a turn-abort error. Letter actions are view-local; only `?`, `Tab` / `Shift-Tab`, and `q` are reserved globally.

**Extensibility is structural, not configurational.** New modes are new sub-models plus bar slots; new sub-views extend a mode's internal view stack; engine-side growth (new collector methods, observer events, mutation types) flows through the adapter without changing the overlay or event-stream infrastructure. Theming and cross-launch persistence are deferred to later initiatives within the arc.

Package layout is `internal/faction/tui/` with `views/` and `adapter/` subdirectories. `gm-toolkit faction` opens the TUI directly.

## Open Questions

The items below are decisions Discovery deliberately deferred. Each carries a real branching cost — choosing one path changes the implementation shape enough that surfacing the choice early matters. **Arc-Plan routes** each item to a specific initiative; the **per-initiative Plan** for that initiative then picks.

- **Faction CRUD form scope.** What is editable in `faction-create` vs `faction-edit`? Identity and stats only, with goals and assets handled in separate sub-views — or inline at create time? The narrower scope ships sooner; the broader scope mirrors how factions actually exist in the TOML on disk. The per-initiative Plan picks which.

- **What `Start` invokes from the Turn setup view.** `RunCycle` (one engine call, abort unwinds the whole cycle) or `RunFactionTurn` in a loop (one call per faction, abort unwinds only the current faction)? The choice ripples into abort semantics, mid-cycle cadence flips, and how the GM perceives "stop now". Both shapes are supported by the engine surface; the per-initiative Plan picks based on which abort granularity matches the GM's mental model.

- **Error rendering shape.** Where does `OnError` surface visually? The error conventions branch distinguishes `Recoverable` from `Fatal`. Candidate behaviors: event stream only; event stream plus a persistent error chip in the state pane; modal interrupt for fatals, inline for recoverables. The per-initiative Plan picks the rendering grammar.

- **Empty-state UX.** First launch with zero factions: how do Manage and Turn modes behave? Likely Manage's list shows a "press `n` to create" hint and Turn's setup view blocks with a "no factions" message — but the exact onboarding moment shapes first-session impressions and is worth nailing in the per-initiative Plan rather than discovering through implementation.

---

## Framework

**Decision: Charm stack (`bubbletea` + `lipgloss` + `huh`).**

Rationale:
- **CLAUDE.md alignment.** Project conventions already mandate `lipgloss` for TUI styling and `huh` for forms. Bubbletea is the missing third piece — adopting it completes the canonical Charm pairing rather than introducing a fourth ecosystem.
- **Pure-consumer fit.** Bubbletea's Model/Update/View architecture is structurally a *projection*: the Model is rebuilt from messages, not authoritative on its own. That aligns the "TUI is a pure engine consumer" constraint with the framework's own grain — the Model is by construction a snapshot of engine state, not state itself.
- **Workflow fit.** Faction CRUD is form-heavy (`huh`'s native territory); interactive turn execution is event-driven rendering (`bubbletea`'s native territory). The two MVP workflows each match the strongest part of the stack.

V1 was also Charm-based. Its failure was structural, not framework-driven: the Model held canonical state and Update functions performed game logic, both of which the framework permits but does not require. The discipline against recurrence lives in *how Model is structured* and *what Update is allowed to do* — addressed in **Architecture** below — not in framework choice.

**Dependencies to add.** `go.mod` currently lists only `BurntSushi/toml` and `go.uber.org/mock`; the Charm deps were removed with V1. Plan phase reintroduces: `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`, `github.com/charmbracelet/huh`.

## Architecture

### Engine-consumption boundary

The engine already exposes the IoC shape the TUI needs. The TUI's call surface is exactly two functions:

- `(*Engine).RunFactionTurn(state, cfg, collectors, observer) (bool, error)` — drives one faction's turn within a cycle; the `bool` (`cycleDone`) signals whether the cycle has completed.
- `(*Engine).RunCycle(state, cfg, collectors, observer) error` — drives a full cycle to completion.

Both call back into the TUI through two interfaces:

- **`PhaseCollector`** — the engine asks for GM input. Methods: `AwaitCheckpoint`, `SelectAction`, `SelectStatRaise`, `SelectMovementDecisions`, `SelectTransportCargo`.
- **`TurnObserver`** — the engine reports activity. 14 callbacks covering turn lifecycle, mutations, movement events, errors, and cycle completion.

The TUI implements both interfaces. **The engine drives the turn; the TUI responds to questions and observes progress.** V1's failure mode (TUI as orchestrator) is structurally impossible at this boundary: the engine does not expose step-by-step entry points the TUI could chain. There is one entry point per faction turn (or per cycle), and it calls back into the TUI for every decision and event.

The boundary is enforced by interface. The TUI imports engine sub-packages (`action`, `turn`, `domain`, etc.) only for the types passed through interface signatures. No engine method outside `RunFactionTurn` / `RunCycle` is on the TUI's call surface.

### Execution cadence

Per-faction and whole-cycle execution are *cadences within a cycle*, not separate modes. The engine owns cycle ordering and boundaries; the TUI controls whether to pause at engine-emitted checkpoints.

The engine emits checkpoints via `PhaseCollector.AwaitCheckpoint(phase)` at five known points: `CheckpointGoalLocked`, `CheckpointBookkeeping`, `CheckpointActionResult`, `CheckpointMovement`, `CheckpointCycleSummary`. The TUI's collector reads a cadence flag from a shared reference:

- **Per-faction cadence** — `AwaitCheckpoint` pauses for GM acknowledgement at each checkpoint, showing the relevant state and blocking until the GM advances.
- **Whole-cycle cadence** — `AwaitCheckpoint` returns immediately at non-cycle-end checkpoints; the GM only sees the `CycleSummary` pause.

The cadence flag is flippable mid-cycle. Mid-cycle the GM can shift from per-faction to whole-cycle ("I've seen enough, run the rest"); the engine notices nothing because the cadence is purely the collector's behavior at checkpoints. The TUI never invents cycle boundaries — those are engine-owned and surfaced via `OnCycleCompleted`.

### Concurrency model

Bubbletea owns the main thread; `RunFactionTurn` / `RunCycle` are synchronous and call back into collectors that block. The two event loops reconcile via:

- **Engine runs in a goroutine** launched by a `tea.Cmd`.
- **Collectors block on channels.** `SelectAction` (etc.) sends an "ask the GM" Message to Bubbletea, then blocks reading a reply channel. Bubbletea handles the prompt via `huh` form or selectable list, then sends the answer back through the channel. The collector method returns and the engine continues.
- **Observers push channels.** Each observer callback sends an event Message to Bubbletea via a channel-pump `tea.Cmd`. The engine never blocks on rendering.
- **Cadence flag** is read by the collector from a shared reference (guarded by a mutex or atomic).

These channels, the cadence reference, and the Collectors/Observer implementations are the **TUI-Engine adapter** — a self-contained sub-package whose only job is bridging the two event loops.

### Package layout

```
internal/faction/tui/
├── tui.go            -- entry point (called from main.go's faction dispatch)
├── model.go          -- Bubbletea Model
├── update.go         -- Update function and Message types
├── view.go           -- top-level View routing
├── views/            -- per-view rendering (faction-list, faction-detail, turn-execution, ...)
└── adapter/          -- TUI-Engine adapter (Collectors + Observer + channels)
```

`internal/faction/tui/` co-locates the TUI with the rest of the faction tool's internals (`internal/faction/{config,data,domain,engine,...}`). No `cmd/faction-manager/` directory — `main.go` already dispatches to "Tools," and `gm-toolkit faction` opens the TUI directly. Future F-004 CLI commands become siblings (`gm-toolkit faction list`, etc.).

### Update discipline

Update is the seam V1 violated. Explicit, binding rules:

**Update MAY:**
- Read the Model.
- Transform Messages into a new Model.
- Emit `tea.Cmd` values for side-effecting work (engine calls, file I/O, channel reads).

**Update MAY NOT:**
- Call into the engine synchronously (engine calls live in `tea.Cmd`s that run off the main thread).
- Mutate engine state directly.
- Perform game logic — checking rule conditions, applying mutations, deciding turn progression. All such logic lives in the engine.

The same rules apply transitively:
- **Model** holds data and projections only. Model methods are pure functions of Model state.
- **View** is read-only on Model. View functions never mutate, never call out.

These rules are binding. Plan and Execution treat violations as bugs. Revision is allowed but must go through the decisions log; ad-hoc exceptions are not.

## View Structure

The TUI presents a persistent **mode bar** along the top and a **mode-owned content area** beneath it. Mode bar slots: `Manage`, `Turn`, `Spatial`, `Quit`. Default landing is `Manage`. `Spatial` is a greyed-out placeholder for `F-012` — reserving the slot in MVP advertises the growth axis and lets the future initiative add a sub-model without reshuffling the bar's routing. `Quit` is a slot for discoverability *and* a `q` keybinding for convention; both dispatch to the same confirm-and-exit action.

Each mode owns a sub-model with its own internal view stack and per-mode split shape. The root Model holds a `mode` field and dispatches Update/View to the active sub-model; the mode bar itself never knows which sub-view is active within a mode. This is the standard Bubbletea router-model pattern, and it bears directly on the "MVP grows without overhaul" constraint: adding a mode or a sub-view never touches the root's Update logic.

### Manage

Detail-on-Enter split. Mode entry is a faction list (left) plus a thin context strip (right) — the strip holds orienting state that doesn't compete for attention: last cycle number, next cycle, hotkey hints, last error preview. `Enter` on a list row opens **faction-detail**, which takes over the content area below the bar. Edit is its own view, reached from detail. The internal view stack:

```
list ──Enter──► detail ──e──► edit
```

CRUD entry points (create from list, delete-confirm from detail) are sub-views of the same stack. Faction CRUD is the only TUI workflow that does not go through the engine — it reads and writes faction state directly. The "TUI is a pure consumer" constraint applies precisely to engine-mediated work; CRUD owns a draft buffer that commits to disk on save, which is structurally distinct from holding canonical *game* state and does not violate the constraint.

### Turn

Mode entry opens a **setup view** that collects cadence (per-faction vs whole-cycle) and confirms the faction roster for this cycle. Pressing Start launches execution; the engine call (`RunFactionTurn` or `RunCycle`) is dispatched from there per the Concurrency model. Cadence is a TUI behavior at `AwaitCheckpoint`, not an engine concept — having the GM select it in setup makes that abstraction visible rather than hidden behind defaults.

The **execution view** is a left/right split: left pane shows the currently-processing faction's state (stats, current goal, cadence flag, cycle number); right pane is a scrolling event log fed by `TurnObserver` callbacks. The same callback that updates the left pane (e.g., `OnFactionTurnStarted`) appends a line to the right pane's log — one observer dimension, two rendering channels.

Collector prompts — both `AwaitCheckpoint` acks and the four `Select*` decisions — surface as **modal overlays** floating over the content area. There is one overlay component; for decisions it wraps a `huh` form, for acks it wraps a single-key listener. Resolution sends the answer back through the adapter's reply channel and the overlay dismisses. A new collector method added by the engine in the future requires a new overlay content but no change to the overlay infrastructure.

Cadence remains toggleable mid-cycle via keybinding from the execution view. The flag lives in shared collector state per Execution cadence; the engine notices nothing.

### Spatial

Greyed-out slot in the mode bar; selecting it is a no-op until `F-012` ships a real sub-model.

### Quit

Bar slot for discoverability, `q` keybinding for convention. Both routes call the same confirm-and-exit action.

## Input Model

The grammar is **letter actions plus arrow navigation**: arrows (with `h`/`j`/`k`/`l` as aliases) move within a view, single letters trigger actions (`e` for edit, `n` for new, etc.), `Enter` moves forward through a view stack, and `Esc` moves back one step. The keybinding budget is sustained by making letter actions **view-local** — only `?`, `Tab` / `Shift-Tab`, and `q` are reserved globally, so each view reuses its own letter namespace freely. `huh` owns input grammar inside forms (Tab field-advance, Enter submit, etc.) and the TUI does not override it.

### Mode switching

`Tab` advances to the next mode bar slot; `Shift-Tab` reverses; greyed slots (`Spatial` in MVP) are skipped. Mode-switch is suppressed while a modal overlay is open — and Tab inside a `huh` form means field-advance, per huh defaults. The grammar is context-dependent in exactly the way every TUI with both navigation and forms has to be; the help overlay teaches it.

### Modal overlay capture

Collector prompts (both `AwaitCheckpoint` acks and `Select*` decisions) appear as modal overlays per View Structure. While an overlay is open:

- All input outside the overlay's own grammar is captured. `Tab`, `q`, and view-local letters are inert.
- `?` remains active, opening a help overlay layered above the prompt.
- `Esc` resolves the overlay with a **cancel sentinel** the adapter forwards as a turn-abort error. The orchestrator unwinds the current faction turn, fires `OnError`, and returns control to the TUI's execution view. This is the GM's escape hatch from a misclick or a wrong-faction setup.

The cancel path leans on existing engine error conventions (`4ebebc4`) — the adapter classifies the abort as `Recoverable` and the orchestrator already handles that branch.

### Event stream

The Turn execution view's event stream **auto-follows** new events from the bottom by default. **Scrolling up locks** the viewport: new events still append to the buffer, but the display stays where the GM put it. `End` (or `G`) jumps back to bottom and re-engages auto-follow. Standard tail / `less` behavior; matches expectations from terminal-native tooling.

### Help

`?` from any view opens a help overlay listing the current view's keybindings plus the global keys. The overlay is dismissed by any key. There is no persistent footer hint strip — the row is reserved for content density, and the GM is expected to learn the small global set quickly and rely on `?` for view-specific keys.

## Extensibility Shape

The Problem statement's "structurally complete, capable of growing to feature-complete without a major overhaul" constraint is satisfied when each anticipated growth axis maps cleanly onto an existing pattern in the architecture above. This section names those patterns as the certified growth path — not new structure, but explicit certification that the structure that's there is enough.

### New top-level modes

Pattern: **new sub-model plus new mode bar slot.** The root Model dispatches Update/View to whichever sub-model is active; adding a slot is a one-line bar change and a sub-model registration. The root's Update logic and the existing modes' sub-models are untouched.

Concrete near-term examples: `F-012` (Spatial) replaces the greyed slot with a real Spatial sub-model; direct state editing — the stretch workflow from the Problem statement — slots in as a new Edit mode when it ships. Neither requires modifying any existing mode.

### New sub-views within a mode

Pattern: **extend the mode's internal view stack.** Each mode owns a stack (e.g., Manage's `list → detail → edit`); a new sub-view is a new entry in the mode's internal `view` field plus the rendering and Update logic that activates it. The mode bar never knows; sibling modes never know.

Concrete near-term examples: faction-detail growing tabs (Stats / Assets / Goals / History) is sub-view growth within Manage; a turn-execution post-mortem view ("review last cycle") is sub-view growth within Turn.

### Engine-surface growth

The engine boundary itself — `RunFactionTurn` / `RunCycle` + `PhaseCollector` + `TurnObserver` — is fixed from the TUI's perspective. The TUI does not grow alternative bypass routes around it. When a TUI need cannot be expressed through this surface, the *engine* grows; not the TUI.

When the engine does grow, the TUI absorbs the change through three patterns:

- **New collector methods.** The engine adds a `Select…` to `PhaseCollector`; the adapter wraps it as another channel-blocked call. New overlay content (typically a `huh` form definition) renders the prompt. The overlay infrastructure is unchanged.
- **New observer events.** The engine adds an `On…` method; the adapter pumps it as a new Message type onto Bubbletea's queue. A new Update branch routes it to the appropriate sub-model's state and to the event stream renderer. Events structurally similar to existing ones (e.g., another `On…Applied` carrying mutations) reuse the existing rendering pipeline.
- **New mutation types.** Engine adds a kind to the mutation taxonomy; the TUI's mutation renderer adds a case. Adapter, sub-models, and observer routing are unaffected.

### Deferrals

**Deferred to later initiatives within the arc** — out of scope for the first initiative, expected to land in subsequent ones:

- **Theming** — swappable palettes. Style is a single fixed sheet in `internal/faction/tui/styles.go` until a theming initiative ships.
- **Cross-launch persistence** — remembered preferences. Launch state is recomputed each run.
- **Direct state editing** — the third workflow named in the Problem statement (GM-side surgical edits outside the engine's mutation pipeline). Slots in as a new Edit mode when its initiative ships.
- **Deeper cycle-execution control** — pause/resume mid-cycle, skip-faction-without-aborting, reorder-on-the-fly. The first initiative stops at `Esc`-aborts-current-turn.

**Out of arc:** no items currently flagged in Discovery. Arc-Plan may surface explicit out-of-arc deferrals.

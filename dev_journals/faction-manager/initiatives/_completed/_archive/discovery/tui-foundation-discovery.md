# TUI Foundation — Discovery

`F-005`, Initiative 1 of the [TUI Rebuild arc](./tui-rebuild-arc-plan.md). Companion to [`tui-rebuild-arc-discovery.md`](./tui-rebuild-arc-discovery.md) — that doc ratified cross-cutting decisions for the whole arc (framework, engine boundary, Update discipline, mode-bar shape). This doc ratifies the **initiative-specific** decisions Foundation needs before Plan: the adapter's internal shape, the Bubbletea ↔ adapter message boundary, the dry-run smoke exercise design, and the root Model's dispatch.

## Problem

Foundation is the load-bearing initiative of the arc. The TUI rebuild's V1 inheritance is the constraint: V1 collapsed because Model held canonical state and Update performed game logic. Arc-discovery ratified the structural defense — engine-consumption boundary, adapter sub-package, Update discipline — but those decisions are claims until proven. Foundation's job is to **prove the adapter against the real engine before either feature workflow is built on top of it**.

Three concrete pieces of risk concentrate here:

1. **The adapter is the architecture.** The arc-discovery's framing — "Bubbletea is by construction a projection" — only holds if the engine is genuinely off the main thread and the collectors genuinely block on reply channels. If Foundation lands without exercising that pattern E2E, Initiative 2 (Manage) ships before Initiative 3 (Turn) discovers the wiring is wrong.

2. **The engine surface is wider than arc-discovery enumerated.** The engine's `Collectors` struct (`internal/faction/engine/collector.go:23`) bundles two interfaces: `PhaseCollector` (5 methods) **and** `action.Collector` (~17 methods, action sub-engine prompts like `SelectAsset`, `SelectAttackers`, `ConfirmAbilityApplied`). Arc-discovery's "TUI implements two interfaces" referred to `PhaseCollector` and `TurnObserver`; `action.Collector` is the unmentioned third. Foundation must implement it to call `RunFactionTurn` / `RunCycle` at all.

3. **The mode-bar router pattern needs to be built such that adding Manage's faction-list and Turn's setup/execution views is purely additive.** The "MVP grows without overhaul" constraint from arc-discovery binds here. Root Update logic must not have to change when new modes' sub-views land.

Foundation ships package skeleton + mode-bar router + adapter sub-package + a dry-run smoke exercise that drives `RunCycle` against a trivial in-memory faction state. After this initiative, Initiative 2 and Initiative 3 add features without revisiting the boundary.

## Design Summary

Foundation builds `internal/faction/tui/` from scratch with the layout arc-discovery prescribed (`tui.go`, `model.go`, `update.go`, `view.go`, `views/`, `adapter/`, `styles.go`). Charm dependencies (`bubbletea`, `lipgloss`, `huh`) land in `go.mod` in the same commit as the skeleton.

The **adapter sub-package** is a single flat package under `internal/faction/tui/adapter/`, with files split by concern (`adapter.go`, `phase_collector.go`, `action_collector.go`, `observer.go`, `channels.go`, `run.go`). The `Adapter` struct owns the engine-side wiring: a request/reply channel for collector asks, an event channel for observer pushes, the cadence flag (an `atomic.Bool`), and the engine goroutine's lifecycle. Foundation implements `PhaseCollector` and `TurnObserver` for real, and `action.Collector` as **error-returning stubs** — methods exist to satisfy the interface but return `errors.New("action.Collector: not implemented in foundation")` if called. The dry-run is shaped so they're never called.

The **Bubbletea ↔ adapter boundary** crosses through three typed `tea.Msg` types defined in `adapter/channels.go`: `CollectorAskMsg` (carries a `Kind` discriminator, faction context, payload, and a reply channel), `ObserverEventMsg` (carries a `Kind` discriminator and payload), and `EngineDoneMsg` (carries the cycle's terminal error or nil). Update has a case per Msg and dispatches to the active sub-model; sub-models reply by writing the chosen value to the `CollectorAskMsg.Reply` channel and emitting a `tea.Cmd` that closes the local turn. The observer pump is a single `tea.Cmd` launched at adapter start that loops reading the event channel and emitting `ObserverEventMsg`s.

The **root Model** holds `mode Mode` (typed enum), `subs map[Mode]tea.Model` (the registered sub-models), a `modebar.Model`, a help-overlay stub, and the `*adapter.Adapter`. Update handles the global keys (`Tab`, `Shift-Tab`, `?`, `q`) and dispatches all other messages to `subs[mode]`. View calls the active sub-model's `View()` and composes it with the mode bar and help overlay. Adding a new mode is one map entry plus one bar slot — exactly the extensibility shape arc-discovery's "Extensibility" section certified.

The **mode bar** is its own component at `internal/faction/tui/views/modebar/`. It owns slot definitions, the active-slot highlight, the greyed-out state for Spatial, and `Tab` / `Shift-Tab` cycling. Foundation's empty Manage and Turn shells live at `views/manage/` and `views/turn/`; each renders a single styled banner identifying the mode and the initiative that fills it in (`"Manage — faction CRUD ships in Initiative 2"`).

The **dry-run smoke exercise** is gated by a Cobra sub-command flag: `gm-toolkit faction --dryrun`. With the flag, the command bypasses Bubbletea entirely — it constructs a trivial in-memory `FactionState`, instantiates the `Adapter`, calls `engine.RunCycle` through the adapter on the main thread (no Bubbletea event loop needed — the smoke uses an auto-acking collector implementation and a stdout-printing observer that don't depend on Tea messages), prints observer events as they arrive, and exits with the cycle's error code. The smoke proves the **structural wiring** of the adapter (interface satisfaction, channel types, goroutine pattern) without requiring the Bubbletea half. The Bubbletea half is proven by the empty shells rendering and the mode bar cycling — that's a separate axis of confidence, exercised by `gm-toolkit faction` (without the flag).

The **cadence flag** is a `sync/atomic.Bool` field on the `Adapter` struct. Read by `PhaseCollector.AwaitCheckpoint`; written by Initiative 3's execution-view keybinding (Foundation leaves it at default-zero, which the smoke interprets as whole-cycle and which is fine for an automated smoke). `false = whole-cycle`, `true = per-faction` — the convention is committed in code and documented in adapter.go.

## Per-Area Design Details

### Adapter sub-package

File layout (all in `internal/faction/tui/adapter/`):

| File | Contains |
|---|---|
| `adapter.go` | `Adapter` struct, `New(engine *engine.Engine, factionState *state.FactionState, cfg *config.Config) *Adapter`, `Stop()`, internal lifecycle helpers |
| `phase_collector.go` | `phaseCollector` type implementing `engine.PhaseCollector`; all five methods send `CollectorAskMsg`s and block on a reply channel; `AwaitCheckpoint` consults `Adapter.perFactionCadence` before deciding to ask |
| `action_collector.go` | `stubActionCollector` type implementing `action.Collector`; all ~17 methods return a single shared `ErrActionNotImplementedInFoundation` error |
| `observer.go` | `observer` type implementing `engine.TurnObserver`; all 14 callbacks push `ObserverEventMsg`s onto the adapter's event channel; `OnError` and `OnCycleCompleted` get their own typed payloads since downstream rendering will branch on them |
| `channels.go` | `CollectorAskMsg`, `ObserverEventMsg`, `EngineDoneMsg` definitions; `AskKind` and `EventKind` enums; reply channel typedefs |
| `run.go` | The `tea.Cmd` that launches the engine goroutine (calls `engine.RunCycle` or `engine.RunFactionTurn`, returns `EngineDoneMsg` when done); the observer pump `tea.Cmd` that loops on the event channel |

The `Adapter` struct exposes only what the TUI layer needs: a constructor, a `Run() tea.Cmd` that returns the engine-launch command, an `ObserverPump() tea.Cmd` for the observer messages, the cadence-flag setter for Initiative 3, and `Stop()` for clean shutdown. Internal channels are unexported.

### Bubbletea ↔ adapter message boundary

```go
// adapter/channels.go

type AskKind int

const (
    AskAwaitCheckpoint AskKind = iota
    AskSelectAction
    AskSelectStatRaise
    AskSelectMovementDecisions
    AskSelectTransportCargo
)

type CollectorAskMsg struct {
    Kind    AskKind
    Faction *domain.Faction  // nil for AskAwaitCheckpoint
    Payload any              // typed by Kind; e.g. []action.Action for AskSelectAction
    Reply   chan<- any       // sub-model writes the chosen value here
}

type EventKind int

const (
    EvtFactionTurnStarted EventKind = iota
    EvtFactionSkipped
    EvtGoalLockApplied
    // ... one per TurnObserver method
    EvtError
)

type ObserverEventMsg struct {
    Kind    EventKind
    Payload any  // typed by Kind
}

type EngineDoneMsg struct {
    Err error  // nil on clean completion
}
```

`Payload` and `Reply` are typed `any` at the boundary; sub-models cast based on `Kind`. The choice trades type safety at the message boundary for a stable Msg surface — adding a new collector method in the engine becomes "add an `AskKind` constant and a new payload case" rather than "add a new `tea.Msg` type and a new Update case at the root."

### Mode-bar router

```go
// model.go

type Mode int

const (
    ModeManage Mode = iota
    ModeTurn
    ModeSpatial
    ModeQuit  // not actually a sub-model; selecting it dispatches confirm-and-exit
)

type Model struct {
    mode    Mode
    subs    map[Mode]tea.Model
    bar     modebar.Model
    help    help.Model
    adapter *adapter.Adapter
}
```

Root `Update` handles global keys (`Tab`, `Shift-Tab`, `?`, `q`) directly. All other messages — `CollectorAskMsg`, `ObserverEventMsg`, `EngineDoneMsg`, mouse, window-resize, sub-model-local messages — are forwarded to `subs[m.mode].Update(msg)`. The forwarded sub-model returns its updated form and `tea.Cmd`; the root writes it back and returns.

This shape means: in Foundation, `subs[ModeManage]` and `subs[ModeTurn]` are the empty shells. In Initiative 2, the Manage entry is replaced with the real Manage sub-model — Update's root logic is unchanged. Same for Initiative 3 and Turn.

`ModeSpatial` registers a sub-model that renders a "reserved for F-012" banner; the mode bar greys the slot. `ModeQuit` is a bar slot but not a registered sub-model — selecting it bypasses the dispatch and triggers the confirm-and-exit flow directly.

### Empty mode shells

Each empty shell is a `tea.Model` with no internal state beyond what it inherits, an Update that handles no keys (everything propagates through root), and a View that returns a single lipgloss-styled banner:

```go
// views/manage/manage.go

type Model struct{}

func New() Model { return Model{} }

func (m Model) Init() tea.Cmd { return nil }
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m Model) View() string {
    return styles.Placeholder.Render(
        "Manage — faction CRUD ships in Initiative 2 (tui-manage)")
}
```

The placeholder style lives in `internal/faction/tui/styles.go`. Initiative 2 replaces `views/manage/manage.go`'s body wholesale; the file path and the `Model`/`Init`/`Update`/`View` quartet stay.

### Dry-run smoke exercise

The smoke is a separate code path from the TUI launch. Both are dispatched from `cmd/gm-toolkit/faction.go` (or wherever the faction Cobra command lives — Foundation reads the current dispatch layout and slots in):

```go
// pseudocode

var dryrun bool
factionCmd.Flags().BoolVar(&dryrun, "dryrun", false, "run a one-faction smoke cycle and exit")

factionCmd.Run = func(cmd *cobra.Command, args []string) {
    if dryrun {
        runDryRun(...)  // bypasses Bubbletea entirely
        return
    }
    runTUI(...)         // launches Bubbletea program
}
```

`runDryRun` constructs:
- A trivial `FactionState` with one faction (no assets, a goal already locked, no actionable conditions)
- The real `Adapter`, but instantiated with **smoke-mode collectors** — these are auto-acking variants that satisfy the same interfaces but don't go through the Bubbletea message channel. They live in `adapter/smoke.go` (or `tui.go` in the cmd layer — placement is a Plan-level call).
- A real `engine.Engine` built from the existing `engine.New` constructor

Then calls `engine.RunCycle` directly on the main thread. Observer events print to stdout via a stdout-printing observer. The cycle completes, the program exits.

What this proves:
- `PhaseCollector` interface satisfaction by Foundation's implementation
- `action.Collector` interface satisfaction by the stub (interface alignment at compile time; the dry-run state never invokes its methods)
- `TurnObserver` interface satisfaction by Foundation's implementation
- The engine's `RunCycle` call shape works end-to-end with Foundation's adapter types in the seats

What this does **not** prove:
- The channel/goroutine plumbing — the smoke is single-threaded and uses auto-acking collectors. The real channel-blocked pattern is proven by Initiative 3's first turn execution.

The asymmetry is intentional. Foundation cannot prove the goroutine pattern without a real Bubbletea loop consuming the channels (which means modal overlays, which means Initiative 3 work). Foundation can prove **interface and type alignment with the real engine**, which is the part that, if wrong, would force a structural rewrite. The goroutine pattern is comparatively boring — it's the canonical Bubbletea concurrency idiom and the risk of getting it structurally wrong is low.

The arc-plan's text ("validating the channel wiring, the goroutine-and-blocking-collector model, and the observer pump end-to-end") is more aspirational than the realistic Foundation scope. Plan should ratify the narrower smoke scope explicitly.

### Cadence-flag storage

```go
// adapter/adapter.go

type Adapter struct {
    perFactionCadence atomic.Bool  // false = whole-cycle, true = per-faction
    // ...
}

func (a *Adapter) SetPerFactionCadence(on bool) {
    a.perFactionCadence.Store(on)
}

// phase_collector.go

func (p *phaseCollector) AwaitCheckpoint(phase string) error {
    if !p.adapter.perFactionCadence.Load() && phase != engine.CheckpointCycleSummary {
        return nil  // whole-cycle: only the cycle-summary checkpoint pauses
    }
    // send CollectorAskMsg, block on reply, return resulting error (cancel-sentinel handling)
    // — Foundation's smoke variant auto-acks instead
}
```

The setter is the only writer. Initiative 3 binds a keybinding to call it from the execution view. The flag's zero value (`false` / whole-cycle) is the sensible default — a GM driving a turn the first time wants to see each step (per-faction), but the smoke wants no pauses at all (whole-cycle works because the smoke's auto-acking collectors don't ask anyway).

## Decisions Ratified in Discovery

1. **`action.Collector` is implemented as error-returning stubs.** Foundation must provide the interface to call the engine; the dry-run is shaped to never reach action-sub-engine collector calls. Stubs are deleted and replaced with real channel-blocked implementations in Initiative 3 (Turn).
2. **Adapter is one flat package with files split by concern.** Not split by direction (in/out), not single-file. Files: `adapter.go`, `phase_collector.go`, `action_collector.go`, `observer.go`, `channels.go`, `run.go`.
3. **Bubbletea ↔ adapter boundary uses three typed `tea.Msg` types with `Kind` discriminators.** `CollectorAskMsg`, `ObserverEventMsg`, `EngineDoneMsg`. Adding a new collector method in the engine becomes a new `Kind` constant plus a new payload case, not a new Msg type at the boundary.
4. **Dry-run smoke is a Cobra sub-command flag**, `gm-toolkit faction --dryrun`. It bypasses Bubbletea entirely, exercises the adapter's type alignment with the engine using auto-acking collectors and stdout observer, prints events, exits.
5. **Dry-run scope is the realistic minimum**: interface and type alignment with the real engine, proven by calling `RunCycle` on a trivial faction state. The full channel/goroutine pattern is **not** proven by Foundation — that proof lands in Initiative 3 with the first real turn execution. The arc-plan's broader phrasing is acknowledged but the narrower scope is what Foundation actually ships.
6. **Cadence flag is an `atomic.Bool` on the `Adapter` struct.** Single bit, lock-free read in `AwaitCheckpoint`, single setter for Initiative 3. `false = whole-cycle`, `true = per-faction`.
7. **Root Model uses `map[Mode]tea.Model` for sub-model dispatch.** Update handles global keys, forwards everything else to `subs[mode]`. Adding a new mode is one map entry and one bar slot.
8. **Mode bar is its own component** at `internal/faction/tui/views/modebar/`. Owns slots, active-slot highlight, greyed-out state, `Tab`/`Shift-Tab` cycling. Root Model holds an instance.
9. **Empty Manage and Turn shells render a single styled banner** identifying the mode and the initiative that fills it in. Not blank, not a shared placeholder component — just a per-mode `View()` returning one styled string.

## Open Questions

None. All initiative-internal design choices were resolved in this Discovery session. Decisions that were deferred at the arc level (CRUD form scope, `Start` action, error rendering, empty-state UX) belong to Initiative 2 and Initiative 3 per the arc-plan's open-question routing table.

## Out of Scope

Foundation explicitly does **not** ship:

- **Faction list, faction-detail, faction-create, faction-edit views.** Initiative 2.
- **Turn setup view, Turn execution view, modal overlay infrastructure.** Initiative 3.
- **Real `action.Collector` implementation.** Initiative 3 replaces the stubs with channel-blocked methods backed by modal overlays.
- **Real channel/goroutine exercise of the adapter.** Initiative 3 is where the goroutine + reply-channel pattern actually runs against the Bubbletea loop. Foundation proves type alignment, not concurrency behavior.
- **Help overlay content.** Foundation ships a stub `?` handler; Initiative 2 and Initiative 3 populate per-view bindings.
- **Empty-state UX grammar** (hint copy, key prompt for "no factions"). Decided in Initiative 2; Turn inherits.
- **Cancel sentinel routing for `Esc`.** The arc-discovery's `Esc`-aborts-current-turn is an Initiative 3 concern — there's nothing to cancel in Foundation. The cancel sentinel type may be defined in `adapter/channels.go` for completeness, but no routing logic ships.
- **Mid-cycle cadence-flag flipping from a keybinding.** Foundation ships the storage; Initiative 3 wires the keybinding.
- **Charm dependency versions pinned beyond `go.mod` defaults.** Plan picks the specific version range; Foundation absorbs whatever Plan chooses.

## Reference Exemplars

- **Arc-Discovery — Architecture section** (`tui-rebuild-arc-discovery.md`, "Architecture") — ratifies the engine-consumption boundary, concurrency model, package layout, and Update discipline that Foundation implements.
- **Arc-Discovery — Extensibility Shape section** (same doc) — the "new modes = new sub-model + new bar slot" pattern is the certification Foundation's router shape satisfies.
- **Arc-Plan — Initiative 1 entry** ([`tui-rebuild-arc-plan.md`](./tui-rebuild-arc-plan.md), Initiative 1 — `tui-foundation`) — the In Scope / Out of Scope boundary this Discovery refines.
- **Engine surface** (`internal/faction/engine/collector.go`, `internal/faction/engine/observer.go`, `internal/faction/engine/orchestrator.go`) — the `Collectors` struct (Phase + Action), the 14-method `TurnObserver`, the `RunCycle` / `RunFactionTurn` signatures, and the five `Checkpoint*` constants Foundation's adapter binds against.
- **Error conventions branch** (commit `4ebebc4`) — the `Recoverable` / `Fatal` classification the cancel sentinel will lean on in Initiative 3. Foundation does not implement cancel routing but the error conventions are the inherited contract.
- **Structured logging branch** (commit `cf379fc`) — Foundation's adapter and dry-run command use the structured logger inherited without modification.

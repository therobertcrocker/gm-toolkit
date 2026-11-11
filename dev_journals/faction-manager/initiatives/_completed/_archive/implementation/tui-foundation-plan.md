# tui-foundation — Implementation Plan

## Context / Goal

This plan turns [`tui-foundation-discovery.md`](../arcs/tui-rebuild/tui-foundation-discovery.md) into an executable sequence of commits. Foundation is Initiative 1 of the [TUI Rebuild arc](../arcs/tui-rebuild/tui-rebuild-arc-plan.md). After this initiative ships, the TUI can be launched, the mode bar cycles between empty Manage / Turn / Spatial / Quit slots, and a `--dryrun` Cobra flag exercises the adapter's type alignment with the real engine end-to-end. Neither feature workflow is yet implemented; that work lands in Initiatives 2 (`tui-manage`) and 3 (`tui-turn`).

Discovery ratified the structural decisions — adapter sub-package shape, three-Msg boundary, mode-bar router, dry-run smoke scope, cadence-flag storage. This plan ratifies the implementation-time choices Discovery left open and breaks the work into commits sized one-per-execution-session.

**Related artifacts:**
- Discovery: [`tui-foundation-discovery.md`](../arcs/tui-rebuild/tui-foundation-discovery.md)
- Arc-Plan: [`tui-rebuild-arc-plan.md`](../arcs/tui-rebuild/tui-rebuild-arc-plan.md)
- Arc-Discovery: [`tui-rebuild-arc-discovery.md`](../arcs/tui-rebuild/tui-rebuild-arc-discovery.md)

## Decisions Ratified in Planning

1. **Single-file plan, Phases → Commits → Tasks granularity.** The four phases share heavy cross-cutting context (channel types, `Adapter` struct shape); per-effort splits would fragment that context. File written at self-implementable detail — full signatures, struct fields, channel ops, file ordering within each commit.
2. **Adopt Cobra in Foundation.** Re-grounding revealed the repo's current `main.go` is a 14-line `os.Args`-printing stub; there is no Cobra dispatcher to "slot into." Discovery's `factionCmd.Flags().BoolVar(...)` pseudocode already assumed Cobra; this plan honors that assumption. F-004 (CLI Rebuild) will add sibling `faction list` etc. commands rather than introducing the dispatcher itself.
3. **`hooks.Collector` stub methods panic on call; `action.Collector`'s direct methods return `ErrActionNotImplementedInFoundation`.** Re-grounding surfaced that `action.Collector` embeds `hooks.Collector`, and `hooks.Collector`'s two methods are non-error-returning (`SelectModifiers` returns `[]ModifierOffer`, `ConfirmReroll` returns `bool`). Discovery's "all methods return an error" stub strategy doesn't fit those two — panic is the equivalent loud-failure for the dry-run's invariant. The dry-run state is shaped so neither method is reachable; if they ever are, we want the crash.
4. **Smoke-mode collectors live in `internal/faction/tui/adapter/smoke.go`.** Discovery flagged this as a Plan-level call. Co-locating with production collectors keeps interface implementations in one package — if `PhaseCollector` grows a method, real and smoke variants update in lockstep.
5. **CLI layout: `cmd/gm-toolkit/{main.go, faction.go}`.** Standard Go binary layout. The repo-root `main.go` is deleted and replaced by `cmd/gm-toolkit/main.go`. Cobra root command and `main()` live in `main.go`; the `faction` subcommand lives in `faction.go`.
6. **Charm and Cobra versions: latest stable at execution time.** No explicit floor or ceiling beyond Go module semantics. Commit 1's `go.mod` and `go.sum` capture the resolved versions. No `replace` directives.
7. **Five commits across four phases.** Phase 1 (Bootstrap): Commit 1. Phase 2 (Adapter sub-package): Commits 2 and 3. Phase 3 (Root Model + mode bar + empty shells): Commit 4. Phase 4 (Dry-run smoke): Commit 5. Arc-Plan estimated 4–6; this lands in the middle.
8. **Cancel-sentinel type is defined in `channels.go` for completeness; no routing logic.** Per Discovery's Out of Scope. The type is `ErrTurnCanceled` (a package-level `error` variable), referenced by future code in Initiative 3. Foundation imports it nowhere.
9. **`ModeQuit` rendering: confirm-exit overlay in root View, not a registered sub-model.** Per Discovery: "selecting it bypasses the dispatch and triggers the confirm-and-exit flow directly." Foundation implements the simplest form: when `m.mode == ModeQuit`, root `View()` renders a one-line `"Press Enter to quit, Esc to cancel"` overlay; Enter dispatches `tea.Quit`; Esc restores the prior mode. `q` from any mode jumps to `ModeQuit`.
10. **Update discipline lives as a comment block at the top of `update.go`.** Two lists: MAY (forward messages, handle global keys, update mode-bar state, return `tea.Cmd`s) and MAY NOT (call engine methods, mutate `*state.FactionState`, perform I/O outside `tea.Cmd`s, embed game logic). Enforced by the senior-engineer pre-merge review per arc-plan.
11. **Test scope is the smoke plus one mode-bar unit test.** The smoke validates engine ↔ adapter type alignment end-to-end. `modebar_test.go` covers Tab / Shift-Tab cycling logic (pure function, easy to test, hard to inspect from the smoke). All other unit tests deferred — channel-blocking behavior is exercised in Initiative 3 when real overlays consume the channels.

## Shared Context

This section holds the structural definitions that multiple commits reference. Tasks under Work Breakdown point back here rather than re-stating struct shapes.

### Final Package Layout

```
gm-toolkit/
├── cmd/
│   └── gm-toolkit/
│       ├── main.go               # Cobra root + main()
│       └── faction.go            # `gm-toolkit faction` + --dryrun
├── internal/
│   └── faction/
│       ├── engine/...            # untouched
│       └── tui/
│           ├── tui.go            # entry — builds *tea.Program and Run()
│           ├── model.go          # root Model struct
│           ├── update.go         # root Update (with discipline comment block)
│           ├── view.go           # root View (mode bar + sub-model + overlays)
│           ├── styles.go         # lipgloss styles (palette + helpers)
│           ├── dryrun.go         # runDryRun (called from faction.go)
│           ├── adapter/
│           │   ├── adapter.go
│           │   ├── channels.go
│           │   ├── phase_collector.go
│           │   ├── action_collector.go
│           │   ├── observer.go
│           │   ├── run.go
│           │   └── smoke.go
│           └── views/
│               ├── modebar/
│               │   ├── modebar.go
│               │   └── modebar_test.go
│               ├── manage/
│               │   └── manage.go   # empty shell
│               └── turn/
│                   └── turn.go     # empty shell
```

### Dependency Additions

`go.mod` additions in Commit 1:

```
github.com/charmbracelet/bubbletea  v1.x (latest)
github.com/charmbracelet/lipgloss   v1.x (latest)
github.com/charmbracelet/huh        v0.x (latest)
github.com/spf13/cobra              v1.x (latest)
```

Resolve with `go get` at commit time; record exact versions in `go.mod`/`go.sum`.

### Engine Surface (Verified at Plan Time)

The adapter must satisfy three interfaces. Method counts verified against current source at Plan time:

**`engine.PhaseCollector`** (5 methods, `internal/faction/engine/collector.go`):

```go
AwaitCheckpoint(phase string) error
SelectAction(faction *domain.Faction, available []action.Action) (action.Action, error)
SelectStatRaise(faction *domain.Faction, eligible []domain.FactionStat) (*domain.FactionStat, error)
SelectMovementDecisions(faction *domain.Faction, eligible []*domain.Asset) ([]world.MovementDecision, error)
SelectTransportCargo(transport *domain.Asset, eligibleCargo []*domain.Asset, profile *domain.TransportProfile) ([]*domain.Asset, error)
```

**`engine.TurnObserver`** (14 methods, `internal/faction/engine/observer.go`):

```go
OnFactionTurnStarted(faction *domain.Faction)
OnFactionSkipped(faction *domain.Faction)
OnGoalLockApplied(faction *domain.Faction, lock locks.GoalLock, mutations []domain.Mutation)
OnBookkeepingApplied(faction *domain.Faction, result turn.BookkeepingResult, mutations []domain.Mutation)
OnActionSelected(faction *domain.Faction, selectedAction action.Action)
OnActionResolved(faction *domain.Faction, selectedAction action.Action, mutations []domain.Mutation)
OnFactionTurnCompleted(faction *domain.Faction)
OnCycleCompleted(cycleNumber int, factionState *state.FactionState)
OnError(faction *domain.Faction, err error)
OnStatRaiseApplied(faction *domain.Faction, raised *domain.FactionStat, mutations []domain.Mutation)
OnStatRaiseSkipped(faction *domain.Faction)
OnMovementTicked(faction *domain.Faction, mutations []domain.Mutation)
OnMovementResolved(faction *domain.Faction, mutations []domain.Mutation)
OnIndexSkipped(skipped []string)
```

**`action.Collector`** (18 methods total — 2 from embedded `hooks.Collector` + 16 explicit; `internal/faction/engine/action/collector.go`):

```go
// embedded hooks.Collector
SelectModifiers(offers []hooks.ModifierOffer) []hooks.ModifierOffer   // non-error-returning
ConfirmReroll(directive hooks.RerollDirective) bool                    // non-error-returning

// action.Collector direct
SelectFactionTestTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Faction) (*domain.Faction, error)
SelectAsset(assets []*domain.Asset, rulebook *rulebook.Rulebook) (*domain.Asset, error)
SelectRepairOrders(faction *domain.Faction, damaged []*domain.Asset, rulebook *rulebook.Rulebook) ([]action.RepairOrder, error)
SelectBuyOrder(purchasablePerWorld map[string][]*domain.AssetDefinition) (action.BuyOrder, error)
SelectRefitOrder(options []action.RefitOption, rulebook *rulebook.Rulebook) (action.RefitOrder, error)
SelectAttackers(eligible []*domain.Asset, rulebook *rulebook.Rulebook) ([]*domain.Asset, error)
SelectDefender(attacker *domain.Asset, eligible []*domain.Asset, rulebook *rulebook.Rulebook) (*domain.Asset, error)
ConfirmRedirectToBase(defenderFaction *domain.Faction, base *domain.Base, damage int) (bool, error)
SelectExpandInfluenceOrder(faction *domain.Faction, factionState *state.FactionState, eligibleNewBaseWorlds []string) (action.ExpandInfluenceOrder, error)
ConfirmRivalFreeAttack(rival *domain.Faction, rivalRoll, factionRoll int) (bool, error)
SelectBaseAttackers(rival *domain.Faction, eligible []*domain.Asset, rulebook *rulebook.Rulebook) ([]*domain.Asset, error)
SelectAbilityAssets(faction *domain.Faction, candidates []*domain.Asset, rulebook *rulebook.Rulebook) ([]*domain.Asset, error)
ConfirmAbilityApplied(asset *domain.Asset, def *domain.AssetDefinition) (bool, error)
SelectBribeTarget(faction *domain.Faction, factionState *state.FactionState) (*domain.Base, int, error)
SelectSeizeTarget(faction *domain.Faction, factionState *state.FactionState) (string, error)
SelectChangeHomeworldTarget(faction *domain.Faction, factionState *state.FactionState) (string, error)
```

### The `Adapter` Struct (Final Shape)

```go
// internal/faction/tui/adapter/adapter.go

type Adapter struct {
    engine       *engine.Engine
    factionState *state.FactionState

    askCh   chan CollectorAskMsg
    eventCh chan ObserverEventMsg
    doneCh  chan EngineDoneMsg

    perFactionCadence atomic.Bool   // false = whole-cycle, true = per-faction

    log *slog.Logger
}

func New(engine *engine.Engine, factionState *state.FactionState, log *slog.Logger) *Adapter
func (a *Adapter) Phase() engine.PhaseCollector
func (a *Adapter) Action() action.Collector
func (a *Adapter) Observer() engine.TurnObserver
func (a *Adapter) Collectors() engine.Collectors            // bundles Phase + Action
func (a *Adapter) Run() tea.Cmd                              // launches engine goroutine; returns EngineDoneMsg when done
func (a *Adapter) ObserverPump() tea.Cmd                     // recursive Cmd: read one event, emit ObserverEventMsg, re-pump
func (a *Adapter) SetPerFactionCadence(on bool)
func (a *Adapter) Stop()                                     // closes channels, signals shutdown
```

Channel buffer sizes: `askCh` unbuffered (synchronous request/reply pattern); `eventCh` buffered (default 64) so the observer pump doesn't block the engine when the UI thread is busy; `doneCh` buffered (1) for the terminal message.

### Channel and Msg Types (Final Shape)

```go
// internal/faction/tui/adapter/channels.go

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
    Faction *domain.Faction   // nil for AskAwaitCheckpoint
    Payload any               // typed by Kind; see Per-Kind Payload Types below
    Reply   chan<- any        // sub-model writes the chosen value (or error sentinel) here
}

type EventKind int

const (
    EvtFactionTurnStarted EventKind = iota
    EvtFactionSkipped
    EvtGoalLockApplied
    EvtBookkeepingApplied
    EvtActionSelected
    EvtActionResolved
    EvtFactionTurnCompleted
    EvtCycleCompleted
    EvtError
    EvtStatRaiseApplied
    EvtStatRaiseSkipped
    EvtMovementTicked
    EvtMovementResolved
    EvtIndexSkipped
)

type ObserverEventMsg struct {
    Kind    EventKind
    Payload any   // typed by Kind; see Per-Kind Payload Types below
}

type EngineDoneMsg struct {
    Err error   // nil on clean completion
}

// Per-Kind Payload Types — concrete struct per kind, the boundary stays any/cast-on-receipt.

type AwaitCheckpointPayload struct{ Phase string }
type SelectActionPayload struct{ Available []action.Action }
type SelectStatRaisePayload struct{ Eligible []domain.FactionStat }
type SelectMovementDecisionsPayload struct{ Eligible []*domain.Asset }
type SelectTransportCargoPayload struct {
    Transport     *domain.Asset
    EligibleCargo []*domain.Asset
    Profile       *domain.TransportProfile
}

type GoalLockAppliedPayload struct {
    Lock      locks.GoalLock
    Mutations []domain.Mutation
}
type BookkeepingAppliedPayload struct {
    Result    turn.BookkeepingResult
    Mutations []domain.Mutation
}
type ActionSelectedPayload struct{ Selected action.Action }
type ActionResolvedPayload struct {
    Selected  action.Action
    Mutations []domain.Mutation
}
type CycleCompletedPayload struct {
    CycleNumber  int
    FactionState *state.FactionState
}
type ErrorPayload struct{ Err error }
type StatRaiseAppliedPayload struct {
    Raised    *domain.FactionStat
    Mutations []domain.Mutation
}
type MovementTickedPayload struct{ Mutations []domain.Mutation }
type MovementResolvedPayload struct{ Mutations []domain.Mutation }
type IndexSkippedPayload struct{ Skipped []string }

// FactionStarted/Skipped/Completed/StatRaiseSkipped events carry only Faction —
// the *domain.Faction is in ObserverEventMsg.Payload directly (no wrapper struct).

// Cancel sentinel (defined here for completeness; Foundation routes nothing through it).
var ErrTurnCanceled = errors.New("turn canceled by user")
```

### Update Discipline Comment Block

Inlined at the top of `internal/faction/tui/update.go` (also referenced from `update.go`'s package doc):

```go
// Update Discipline (per tui-rebuild-arc-discovery.md "Update Discipline").
//
// MAY:
//   - Forward messages to the active sub-model (subs[mode]).
//   - Handle global key bindings (Tab, Shift-Tab, q, ?, Esc) at the root.
//   - Update mode-bar state and switch the active mode.
//   - Update root-overlay state (help overlay, confirm-exit overlay).
//   - Return tea.Cmds for I/O performed by the adapter (engine launch, observer pump).
//
// MAY NOT:
//   - Call engine methods directly (engine.RunCycle, engine.RunFactionTurn, etc.).
//   - Mutate *state.FactionState or any *domain.* value.
//   - Perform I/O outside a tea.Cmd (filesystem reads, network calls, logging beyond
//     the structured logger's non-blocking calls).
//   - Embed game logic — turn ordering, action selection, mutation application all
//     belong to the engine. Update is a projection of engine state and a forwarder
//     of user intent.
//
// Violations are senior-engineer code review concerns at pre-merge time.
```

### Smoke Invariant

The dry-run's trivial faction state must be shaped so that no `action.Collector` method (direct or via `hooks.Collector` embed) is ever called. Concretely: one faction, zero assets, all goals pre-locked, no `action.Action` whose execution requires action-sub-engine collector calls. The smoke chooses `AbandonGoal` or `RepairFaction` as the available action — both are simple enough to resolve without `Select*`/`Confirm*` callbacks on `action.Collector`. If a panic from `action_collector.go` or `hooks.Collector` stub methods is observed during smoke, the smoke's faction state has drifted from the invariant; the fix is the state, not the stub.

## Out of Scope

Inherited from Discovery (`tui-foundation-discovery.md` § Out of Scope), restated here for the executor's convenience:

- Faction list, faction-detail, faction-create, faction-edit views (Initiative 2).
- Turn setup view, Turn execution view, modal overlay infrastructure (Initiative 3).
- Real `action.Collector` implementation (Initiative 3; replaces stubs with channel-blocked methods).
- Real channel/goroutine exercise of the adapter under a live Bubbletea loop (Initiative 3; Foundation only exercises interface and type alignment via the smoke's auto-acking collectors).
- Help overlay content (stub `?` handler only; per-view bindings populate in Initiative 2 and 3).
- Empty-state UX grammar ("no factions" hint copy, key prompt for create — Initiative 2 decides, Turn inherits).
- Cancel sentinel routing for `Esc` (Initiative 3 wires the routing; Foundation defines the sentinel type).
- Mid-cycle cadence-flag flipping from a keybinding (Initiative 3; Foundation ships the `SetPerFactionCadence` setter and the storage).
- Confirm-exit polish (multi-key confirmation flows, dirty-state prompts) — Foundation ships the simplest single-line overlay.

Out of Scope from Plan (not in Discovery's list, captured here):

- F-004 (CLI Rebuild). Foundation introduces Cobra and `gm-toolkit faction`. Sibling commands (`faction list`, etc.) belong to F-004's initiative.
- Theming. Single fixed palette in `styles.go`. Theme-swap belongs to a later initiative.

---

## Work Breakdown

### Phase 1 — Bootstrap

Adds the dependency set, the Cobra dispatcher, and the empty `tui` package skeleton. After this phase the binary compiles and `gm-toolkit faction` runs but prints a placeholder. No adapter, no real root Model yet.

#### Commit 1 — `feat(tui): bootstrap tui package with charm and cobra deps`

File-creation order (each step compiles before moving to the next):

##### Task 1 — `go.mod` / `go.sum`

Run from the repo root:

```sh
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/huh@latest
go get github.com/spf13/cobra@latest
go mod tidy
```

Verify all four resolve. Commit the resulting `go.mod` and `go.sum` with the rest of this commit's files.

##### Task 2 — `internal/faction/tui/styles.go`

Initial style palette. Foundation uses these styles in Phase 3 (modebar, empty shells, confirm-exit overlay). All styles exported, all named for what they render — not for the color (theme-swap-friendly later).

```go
package tui

import "github.com/charmbracelet/lipgloss"

var (
    // Mode bar slot styles.
    ModeBarActive   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Background(lipgloss.Color("0")).Padding(0, 2)
    ModeBarInactive = lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Padding(0, 2)
    ModeBarDisabled = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Padding(0, 2)
    ModeBarSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).SetString(" │ ")

    // Empty-shell placeholder banner.
    Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Padding(2, 4).Italic(true)

    // Confirm-exit overlay.
    ConfirmExit = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9")).Padding(1, 2)

    // Help-overlay stub.
    HelpStub = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Padding(1, 2)
)
```

##### Task 3 — `internal/faction/tui/tui.go`

Entry point. Foundation ships a stub `Run` that prints a placeholder and returns nil; Commit 4 replaces the body with `tea.NewProgram(...).Run()`.

```go
package tui

import (
    "fmt"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// Run launches the TUI. Foundation's stub prints a placeholder; Commit 4 wires
// up tea.NewProgram with the root Model.
func Run(factionState *state.FactionState) error {
    fmt.Println("TUI not yet wired (Commit 4 replaces this stub).")
    return nil
}
```

The `*state.FactionState` parameter is wired through from the Cobra command. Even though Foundation's stub doesn't use it, the signature is final — Commit 4 consumes it.

##### Task 4 — `cmd/gm-toolkit/main.go`

Create the directory and the Cobra root. `main()` calls `rootCmd.Execute()`.

```go
package main

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "gm-toolkit",
    Short: "GM Toolkit — tools for running Stars Without Number factions and worlds",
    Long:  "GM Toolkit bundles utilities for the GM running Stars Without Number campaigns. Run `gm-toolkit <subcommand> --help` for per-subcommand details.",
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

##### Task 5 — `cmd/gm-toolkit/faction.go`

`faction` subcommand. Foundation's body calls `tui.Run`; the `--dryrun` flag is declared but unused (Commit 5 wires it).

```go
package main

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui"
    // factionState loader import path — verify at execution time; the existing
    // config layer's loader func is reused. If the loader signature is
    // load(path string) (*state.FactionState, error), call it here.
)

var (
    factionDryRun bool
)

var factionCmd = &cobra.Command{
    Use:   "faction",
    Short: "Open the faction TUI",
    Long:  "Launches the faction-manager TUI. Use --dryrun to run a one-faction smoke cycle through the engine and exit (developer-only).",
    RunE:  factionRun,
}

func init() {
    factionCmd.Flags().BoolVar(&factionDryRun, "dryrun", false, "run a one-faction smoke cycle and exit")
    rootCmd.AddCommand(factionCmd)
}

func factionRun(cmd *cobra.Command, args []string) error {
    if factionDryRun {
        fmt.Fprintln(os.Stderr, "--dryrun not yet wired (Commit 5 replaces this branch).")
        return nil
    }

    // TODO Commit 5: load FactionState from config and pass to tui.Run.
    return tui.Run(nil)
}
```

Note: the factionState loader call is left as a `TODO Commit 5` because the smoke also needs to construct a trivial state — Commit 5 is the right place to centralize the loader-or-construct branching.

##### Task 6 — Delete repo-root `main.go`

The repo-root `main.go` is replaced by `cmd/gm-toolkit/main.go`. Delete it.

```sh
rm main.go
```

Verify the binary now builds from `cmd/gm-toolkit/`:

```sh
go build ./cmd/gm-toolkit
./gm-toolkit faction        # should print the Commit 4 stub message
./gm-toolkit faction --dryrun   # should print the Commit 5 stub message
./gm-toolkit                 # should print Cobra usage
rm gm-toolkit
```

##### Commit message

```
feat(tui): bootstrap tui package with charm and cobra deps

- add bubbletea, lipgloss, huh, cobra to go.mod
- move binary entry to cmd/gm-toolkit/
- introduce `gm-toolkit faction` (stub) and `--dryrun` flag (stub)
- scaffold internal/faction/tui/ with stub Run and initial style palette
```

---

### Phase 2 — Adapter Sub-Package

Lays down the adapter package in two commits. Commit 2 ships the static surface — channel/msg types, the `Adapter` struct shell, and the `action.Collector` stub. Commit 3 fills in the live wiring — real `PhaseCollector`, `TurnObserver`, the `tea.Cmd`s, and the `Adapter`'s lifecycle methods.

#### Commit 2 — `feat(tui/adapter): channel types, adapter struct, action.Collector stub`

After this commit the `adapter` package compiles, the `action.Collector` interface is satisfied (the panicking stub is a valid implementation), and the channel/msg types are ready for Commit 3 to wire against. Nothing executes yet.

File-creation order:

##### Task 1 — `internal/faction/tui/adapter/channels.go`

All type defs from Shared Context § "Channel and Msg Types (Final Shape)". Copy that block verbatim into this file. The file should consist of:

1. Package clause: `package adapter`
2. Imports: `errors`, `github.com/therobertcrocker/gm-toolkit/internal/faction/domain`, `internal/faction/engine/action`, `internal/faction/engine/goal/locks`, `internal/faction/engine/turn`, `internal/faction/state`
3. `AskKind` enum and constants
4. `CollectorAskMsg` struct
5. `EventKind` enum and constants
6. `ObserverEventMsg` struct
7. `EngineDoneMsg` struct
8. Per-kind payload struct types (one block, all listed under "Per-Kind Payload Types")
9. `ErrTurnCanceled` sentinel

No methods on these types — they're pure data. The file is purely declarative.

##### Task 2 — `internal/faction/tui/adapter/adapter.go` (struct shell)

Adapter struct + constructor. Methods are declared but bodies are stubbed (panic with "wired in Commit 3") for the lifecycle methods; `New` is implemented because the next file needs it to compile.

```go
package adapter

import (
    "log/slog"
    "sync/atomic"

    tea "github.com/charmbracelet/bubbletea"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type Adapter struct {
    engine       *engine.Engine
    factionState *state.FactionState

    askCh   chan CollectorAskMsg
    eventCh chan ObserverEventMsg
    doneCh  chan EngineDoneMsg

    perFactionCadence atomic.Bool

    log *slog.Logger
}

const eventChanBuffer = 64

func New(eng *engine.Engine, factionState *state.FactionState, log *slog.Logger) *Adapter {
    return &Adapter{
        engine:       eng,
        factionState: factionState,
        askCh:        make(chan CollectorAskMsg),
        eventCh:      make(chan ObserverEventMsg, eventChanBuffer),
        doneCh:       make(chan EngineDoneMsg, 1),
        log:          log,
    }
}

func (a *Adapter) SetPerFactionCadence(on bool) { a.perFactionCadence.Store(on) }

// Phase / Action / Observer / Collectors return interface impls. Stubbed until Commit 3.
func (a *Adapter) Phase() engine.PhaseCollector    { panic("adapter.Phase: wired in Commit 3") }
func (a *Adapter) Action() action.Collector         { return &stubActionCollector{} }
func (a *Adapter) Observer() engine.TurnObserver    { panic("adapter.Observer: wired in Commit 3") }
func (a *Adapter) Collectors() engine.Collectors {
    return engine.Collectors{Phase: a.Phase(), Action: a.Action()}
}

func (a *Adapter) Run() tea.Cmd          { panic("adapter.Run: wired in Commit 3") }
func (a *Adapter) ObserverPump() tea.Cmd { panic("adapter.ObserverPump: wired in Commit 3") }
func (a *Adapter) Stop()                 { /* no-op until Commit 3 */ }
```

The panic stubs are intentional — they make it loud if anything in Commit 2's scope accidentally calls them.

##### Task 3 — `internal/faction/tui/adapter/action_collector.go`

The 18-method stub. Direct `action.Collector` methods return `ErrActionNotImplementedInFoundation`; the two `hooks.Collector` embedded methods panic.

```go
package adapter

import (
    "errors"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// ErrActionNotImplementedInFoundation is returned by every error-returning
// method on action.Collector in Foundation. Initiative 3 replaces this stub
// with channel-blocked implementations backed by modal overlays.
var ErrActionNotImplementedInFoundation = errors.New("action.Collector: not implemented in foundation (Initiative 3 replaces stubs)")

// stubActionCollector satisfies action.Collector. The dry-run is shaped so none
// of these methods is reachable; if one is called, the failure is loud.
type stubActionCollector struct{}

// --- hooks.Collector (embedded; non-error-returning, so panic on call) ---

func (s *stubActionCollector) SelectModifiers(offers []hooks.ModifierOffer) []hooks.ModifierOffer {
    panic("hooks.Collector.SelectModifiers: not implemented in foundation; only reachable from the action sub-engine, which the dry-run never invokes")
}

func (s *stubActionCollector) ConfirmReroll(directive hooks.RerollDirective) bool {
    panic("hooks.Collector.ConfirmReroll: not implemented in foundation; only reachable from the action sub-engine, which the dry-run never invokes")
}

// --- action.Collector direct methods (error-returning) ---

func (s *stubActionCollector) SelectFactionTestTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Faction) (*domain.Faction, error) {
    return nil, ErrActionNotImplementedInFoundation
}
func (s *stubActionCollector) SelectAsset(assets []*domain.Asset, rb *rulebook.Rulebook) (*domain.Asset, error) {
    return nil, ErrActionNotImplementedInFoundation
}
func (s *stubActionCollector) SelectRepairOrders(faction *domain.Faction, damaged []*domain.Asset, rb *rulebook.Rulebook) ([]action.RepairOrder, error) {
    return nil, ErrActionNotImplementedInFoundation
}
func (s *stubActionCollector) SelectBuyOrder(purchasablePerWorld map[string][]*domain.AssetDefinition) (action.BuyOrder, error) {
    return action.BuyOrder{}, ErrActionNotImplementedInFoundation
}
func (s *stubActionCollector) SelectRefitOrder(options []action.RefitOption, rb *rulebook.Rulebook) (action.RefitOrder, error) {
    return action.RefitOrder{}, ErrActionNotImplementedInFoundation
}
func (s *stubActionCollector) SelectAttackers(eligible []*domain.Asset, rb *rulebook.Rulebook) ([]*domain.Asset, error) {
    return nil, ErrActionNotImplementedInFoundation
}
func (s *stubActionCollector) SelectDefender(attacker *domain.Asset, eligible []*domain.Asset, rb *rulebook.Rulebook) (*domain.Asset, error) {
    return nil, ErrActionNotImplementedInFoundation
}
func (s *stubActionCollector) ConfirmRedirectToBase(defenderFaction *domain.Faction, base *domain.Base, damage int) (bool, error) {
    return false, ErrActionNotImplementedInFoundation
}
func (s *stubActionCollector) SelectExpandInfluenceOrder(faction *domain.Faction, factionState *state.FactionState, eligibleNewBaseWorlds []string) (action.ExpandInfluenceOrder, error) {
    return action.ExpandInfluenceOrder{}, ErrActionNotImplementedInFoundation
}
func (s *stubActionCollector) ConfirmRivalFreeAttack(rival *domain.Faction, rivalRoll, factionRoll int) (bool, error) {
    return false, ErrActionNotImplementedInFoundation
}
func (s *stubActionCollector) SelectBaseAttackers(rival *domain.Faction, eligible []*domain.Asset, rb *rulebook.Rulebook) ([]*domain.Asset, error) {
    return nil, ErrActionNotImplementedInFoundation
}
func (s *stubActionCollector) SelectAbilityAssets(faction *domain.Faction, candidates []*domain.Asset, rb *rulebook.Rulebook) ([]*domain.Asset, error) {
    return nil, ErrActionNotImplementedInFoundation
}
func (s *stubActionCollector) ConfirmAbilityApplied(asset *domain.Asset, def *domain.AssetDefinition) (bool, error) {
    return false, ErrActionNotImplementedInFoundation
}
func (s *stubActionCollector) SelectBribeTarget(faction *domain.Faction, factionState *state.FactionState) (*domain.Base, int, error) {
    return nil, 0, ErrActionNotImplementedInFoundation
}
func (s *stubActionCollector) SelectSeizeTarget(faction *domain.Faction, factionState *state.FactionState) (string, error) {
    return "", ErrActionNotImplementedInFoundation
}
func (s *stubActionCollector) SelectChangeHomeworldTarget(faction *domain.Faction, factionState *state.FactionState) (string, error) {
    return "", ErrActionNotImplementedInFoundation
}
```

Verify with `go build ./internal/faction/tui/adapter` — should compile. Verify the interface is satisfied by adding a compile-time check at the bottom of the file:

```go
var _ action.Collector = (*stubActionCollector)(nil)
```

##### Commit message

```
feat(tui/adapter): channel types, adapter struct, action.Collector stub

- declare CollectorAskMsg / ObserverEventMsg / EngineDoneMsg
- declare per-kind payload structs and ErrTurnCanceled sentinel
- introduce Adapter struct with channels, cadence flag, and constructor
- ship stubActionCollector — error-returning for action.Collector methods,
  panicking for the two embedded hooks.Collector methods (non-error returns)
- live methods (Run, ObserverPump, Phase, Observer) panic until Commit 3
```

---

#### Commit 3 — `feat(tui/adapter): phase collector, observer, lifecycle`

Replaces the panicking stubs in `adapter.go` with real implementations and adds three new files: `phase_collector.go`, `observer.go`, `run.go`. After this commit the adapter is a real participant in an engine call — the smoke (Commit 5) can build directly on it.

File-creation order (each step compiles before next):

##### Task 1 — `internal/faction/tui/adapter/observer.go`

The `TurnObserver` implementation. All 14 methods construct an `ObserverEventMsg` (with the appropriate `EventKind` and payload) and push it onto `Adapter.eventCh`. Sends are non-blocking via `select`/`default` — if the channel is full, drop the event and log a warning. The observer must never block the engine.

```go
package adapter

import (
    "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/turn"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type observer struct{ adapter *Adapter }

func (a *Adapter) Observer() engine.TurnObserver { return &observer{adapter: a} }

func (o *observer) emit(kind EventKind, payload any) {
    select {
    case o.adapter.eventCh <- ObserverEventMsg{Kind: kind, Payload: payload}:
    default:
        o.adapter.log.Warn("observer event dropped — eventCh full", "kind", kind)
    }
}

func (o *observer) OnFactionTurnStarted(faction *domain.Faction)   { o.emit(EvtFactionTurnStarted, faction) }
func (o *observer) OnFactionSkipped(faction *domain.Faction)       { o.emit(EvtFactionSkipped, faction) }
func (o *observer) OnFactionTurnCompleted(faction *domain.Faction) { o.emit(EvtFactionTurnCompleted, faction) }
func (o *observer) OnStatRaiseSkipped(faction *domain.Faction)     { o.emit(EvtStatRaiseSkipped, faction) }

func (o *observer) OnGoalLockApplied(faction *domain.Faction, lock locks.GoalLock, muts []domain.Mutation) {
    o.emit(EvtGoalLockApplied, GoalLockAppliedPayload{Lock: lock, Mutations: muts})
}
func (o *observer) OnBookkeepingApplied(faction *domain.Faction, result turn.BookkeepingResult, muts []domain.Mutation) {
    o.emit(EvtBookkeepingApplied, BookkeepingAppliedPayload{Result: result, Mutations: muts})
}
func (o *observer) OnActionSelected(faction *domain.Faction, selected action.Action) {
    o.emit(EvtActionSelected, ActionSelectedPayload{Selected: selected})
}
func (o *observer) OnActionResolved(faction *domain.Faction, selected action.Action, muts []domain.Mutation) {
    o.emit(EvtActionResolved, ActionResolvedPayload{Selected: selected, Mutations: muts})
}
func (o *observer) OnCycleCompleted(cycleNumber int, factionState *state.FactionState) {
    o.emit(EvtCycleCompleted, CycleCompletedPayload{CycleNumber: cycleNumber, FactionState: factionState})
}
func (o *observer) OnError(faction *domain.Faction, err error) {
    o.emit(EvtError, ErrorPayload{Err: err})
}
func (o *observer) OnStatRaiseApplied(faction *domain.Faction, raised *domain.FactionStat, muts []domain.Mutation) {
    o.emit(EvtStatRaiseApplied, StatRaiseAppliedPayload{Raised: raised, Mutations: muts})
}
func (o *observer) OnMovementTicked(faction *domain.Faction, muts []domain.Mutation) {
    o.emit(EvtMovementTicked, MovementTickedPayload{Mutations: muts})
}
func (o *observer) OnMovementResolved(faction *domain.Faction, muts []domain.Mutation) {
    o.emit(EvtMovementResolved, MovementResolvedPayload{Mutations: muts})
}
func (o *observer) OnIndexSkipped(skipped []string) {
    o.emit(EvtIndexSkipped, IndexSkippedPayload{Skipped: skipped})
}

var _ engine.TurnObserver = (*observer)(nil)
```

Note the `faction *domain.Faction` parameter is sometimes ignored (the payload struct doesn't carry it). For the events whose payload is just the faction (Started/Skipped/Completed/StatRaiseSkipped), the `Payload` field carries the `*domain.Faction` directly — Update casts to `*domain.Faction` for those kinds.

##### Task 2 — `internal/faction/tui/adapter/phase_collector.go`

The `PhaseCollector` implementation. All 5 methods construct a `CollectorAskMsg`, push to `askCh`, and block on the reply channel. `AwaitCheckpoint` short-circuits when the cadence flag is `false` (whole-cycle mode) for non-cycle-summary phases.

```go
package adapter

import (
    "fmt"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
)

type phaseCollector struct{ adapter *Adapter }

func (a *Adapter) Phase() engine.PhaseCollector { return &phaseCollector{adapter: a} }

// ask sends a CollectorAskMsg to the UI thread, blocks on the reply, and returns the result.
func (p *phaseCollector) ask(kind AskKind, faction *domain.Faction, payload any) (any, error) {
    reply := make(chan any, 1)
    p.adapter.askCh <- CollectorAskMsg{Kind: kind, Faction: faction, Payload: payload, Reply: reply}
    received := <-reply
    if err, ok := received.(error); ok {
        return nil, err
    }
    return received, nil
}

func (p *phaseCollector) AwaitCheckpoint(phase string) error {
    if !p.adapter.perFactionCadence.Load() && phase != engine.CheckpointCycleSummary {
        return nil
    }
    _, err := p.ask(AskAwaitCheckpoint, nil, AwaitCheckpointPayload{Phase: phase})
    return err
}

func (p *phaseCollector) SelectAction(faction *domain.Faction, available []action.Action) (action.Action, error) {
    raw, err := p.ask(AskSelectAction, faction, SelectActionPayload{Available: available})
    if err != nil {
        return nil, err
    }
    chosen, ok := raw.(action.Action)
    if !ok {
        return nil, fmt.Errorf("phase_collector.SelectAction: unexpected reply type %T", raw)
    }
    return chosen, nil
}

func (p *phaseCollector) SelectStatRaise(faction *domain.Faction, eligible []domain.FactionStat) (*domain.FactionStat, error) {
    raw, err := p.ask(AskSelectStatRaise, faction, SelectStatRaisePayload{Eligible: eligible})
    if err != nil {
        return nil, err
    }
    chosen, ok := raw.(*domain.FactionStat)
    if !ok {
        return nil, fmt.Errorf("phase_collector.SelectStatRaise: unexpected reply type %T", raw)
    }
    return chosen, nil
}

func (p *phaseCollector) SelectMovementDecisions(faction *domain.Faction, eligible []*domain.Asset) ([]world.MovementDecision, error) {
    raw, err := p.ask(AskSelectMovementDecisions, faction, SelectMovementDecisionsPayload{Eligible: eligible})
    if err != nil {
        return nil, err
    }
    chosen, ok := raw.([]world.MovementDecision)
    if !ok {
        return nil, fmt.Errorf("phase_collector.SelectMovementDecisions: unexpected reply type %T", raw)
    }
    return chosen, nil
}

func (p *phaseCollector) SelectTransportCargo(transport *domain.Asset, eligibleCargo []*domain.Asset, profile *domain.TransportProfile) ([]*domain.Asset, error) {
    raw, err := p.ask(AskSelectTransportCargo, nil, SelectTransportCargoPayload{Transport: transport, EligibleCargo: eligibleCargo, Profile: profile})
    if err != nil {
        return nil, err
    }
    chosen, ok := raw.([]*domain.Asset)
    if !ok {
        return nil, fmt.Errorf("phase_collector.SelectTransportCargo: unexpected reply type %T", raw)
    }
    return chosen, nil
}

var _ engine.PhaseCollector = (*phaseCollector)(nil)
```

Verify `engine.CheckpointCycleSummary` constant exists — if not, replace with whatever the engine exports for "cycle summary checkpoint." Re-grounding step at execution time.

##### Task 3 — `internal/faction/tui/adapter/run.go`

Two `tea.Cmd`s: `Run` (launches the engine goroutine, returns `EngineDoneMsg`) and `ObserverPump` (recursive — reads one event from `eventCh`, returns `ObserverEventMsg`, and the consumer re-issues the pump command).

```go
package adapter

import (
    tea "github.com/charmbracelet/bubbletea"
)

func (a *Adapter) Run() tea.Cmd {
    return func() tea.Msg {
        err := a.engine.RunCycle(a.factionState, a.Collectors(), a.Observer())
        return EngineDoneMsg{Err: err}
    }
}

func (a *Adapter) ObserverPump() tea.Cmd {
    return func() tea.Msg {
        msg, ok := <-a.eventCh
        if !ok {
            return nil
        }
        return msg
    }
}

func (a *Adapter) Stop() {
    // close eventCh so any pending ObserverPump returns nil and Update can stop re-issuing.
    close(a.eventCh)
}
```

Note: Foundation's smoke is single-threaded and uses smoke-mode collectors; the smoke does **not** exercise the `Run`/`ObserverPump` `tea.Cmd`s — they exist for Initiative 3. The smoke calls `engine.RunCycle` directly on the main thread. Foundation ships the `tea.Cmd`s with no live exerciser; that's intentional.

The `engine.RunCycle` signature is asserted from current source; verify at execution time and adjust the call shape if it has drifted.

##### Task 4 — `internal/faction/tui/adapter/adapter.go` (remove panic stubs)

Replace the four panicking stubs (`Phase`, `Observer`, `Run`, `ObserverPump`) with the real implementations defined in the new files. They already exist on the `*Adapter` receiver from Tasks 1–3, so this is a deletion: remove the panicking method declarations from `adapter.go`. The real methods on `*Adapter` (Run, ObserverPump, Stop) and the wrapper methods on phaseCollector/observer take over.

After deletion, `adapter.go` should contain only:

```go
package adapter

import (
    "log/slog"
    "sync/atomic"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

const eventChanBuffer = 64

type Adapter struct {
    engine       *engine.Engine
    factionState *state.FactionState

    askCh   chan CollectorAskMsg
    eventCh chan ObserverEventMsg
    doneCh  chan EngineDoneMsg

    perFactionCadence atomic.Bool

    log *slog.Logger
}

func New(eng *engine.Engine, factionState *state.FactionState, log *slog.Logger) *Adapter {
    return &Adapter{
        engine:       eng,
        factionState: factionState,
        askCh:        make(chan CollectorAskMsg),
        eventCh:      make(chan ObserverEventMsg, eventChanBuffer),
        doneCh:       make(chan EngineDoneMsg, 1),
        log:          log,
    }
}

func (a *Adapter) SetPerFactionCadence(on bool) { a.perFactionCadence.Store(on) }
func (a *Adapter) Action() action.Collector     { return &stubActionCollector{} }
func (a *Adapter) Collectors() engine.Collectors {
    return engine.Collectors{Phase: a.Phase(), Action: a.Action()}
}
```

`Phase()` and `Observer()` live in `phase_collector.go` and `observer.go`; `Run()`, `ObserverPump()`, and `Stop()` live in `run.go`.

##### Commit message

```
feat(tui/adapter): phase collector, observer, lifecycle

- implement engine.PhaseCollector — 5 methods, channel-blocked, cadence-aware
- implement engine.TurnObserver — 14 methods, non-blocking emit to eventCh
- add Run() and ObserverPump() tea.Cmds and Stop() lifecycle
- remove the Commit 2 panic stubs from adapter.go
```

---

### Phase 3 — Root Model, Mode Bar, Empty Shells

Stands up the Bubbletea half. After this phase `gm-toolkit faction` launches a real `tea.Program`, the mode bar cycles, and the empty Manage / Turn shells render. The adapter is constructed and `SetPerFactionCadence` is callable, but no `Run`/`ObserverPump` is yet issued — that's Initiative 3's concern.

#### Commit 4 — `feat(tui): root model with mode-bar dispatch and empty shells`

File-creation order:

##### Task 1 — `internal/faction/tui/views/manage/manage.go`

Empty shell. The four-method `tea.Model` contract; `View` renders a single styled banner. Independent of root Model.

```go
package manage

import (
    tea "github.com/charmbracelet/bubbletea"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui"
)

type Model struct{}

func New() Model                                                  { return Model{} }
func (m Model) Init() tea.Cmd                                     { return nil }
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)           { return m, nil }
func (m Model) View() string {
    return tui.Placeholder.Render("Manage — faction CRUD ships in Initiative 2 (tui-manage)")
}
```

Note: importing the `tui` package for `Placeholder` works because `manage` is a sub-package (`tui/views/manage`). No circular import — `tui` does not import `tui/views/manage` from `styles.go`.

If a circular-import problem surfaces at execution time (because `model.go` in `tui` imports `tui/views/manage`), move `Placeholder` to a separate `internal/faction/tui/styles/` package — split the import. Decide at execution time; the simplest layout is co-located styles, and Foundation only flips to split-package if Go forces it.

##### Task 2 — `internal/faction/tui/views/turn/turn.go`

Same shape as Manage's shell.

```go
package turn

import (
    tea "github.com/charmbracelet/bubbletea"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui"
)

type Model struct{}

func New() Model                                                  { return Model{} }
func (m Model) Init() tea.Cmd                                     { return nil }
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)           { return m, nil }
func (m Model) View() string {
    return tui.Placeholder.Render("Turn — interactive cycle execution ships in Initiative 3 (tui-turn)")
}
```

##### Task 3 — `internal/faction/tui/views/modebar/modebar.go`

The mode-bar component. Owns slot definitions, the active-index, cycling, and `View`. The `Mode` type lives in the top-level `tui` package (declared in `model.go` in Task 5); modebar accepts it as an opaque comparable.

To avoid circular imports, modebar declares its own `Mode` type as an `int`, and `model.go` declares `tui.Mode` as `modebar.Mode`. Cleaner: declare `Mode` and the constants in modebar; root `tui` package imports them.

```go
package modebar

import (
    "strings"

    "github.com/charmbracelet/lipgloss"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui"
)

type Mode int

const (
    ModeManage Mode = iota
    ModeTurn
    ModeSpatial
    ModeQuit
)

func (m Mode) String() string {
    switch m {
    case ModeManage:
        return "Manage"
    case ModeTurn:
        return "Turn"
    case ModeSpatial:
        return "Spatial"
    case ModeQuit:
        return "Quit"
    default:
        return "?"
    }
}

type slot struct {
    mode     Mode
    disabled bool
}

type Model struct {
    slots  []slot
    active int
}

func New() Model {
    return Model{
        slots: []slot{
            {mode: ModeManage, disabled: false},
            {mode: ModeTurn, disabled: false},
            {mode: ModeSpatial, disabled: true},
            {mode: ModeQuit, disabled: false},
        },
        active: 0,
    }
}

func (m Model) Active() Mode { return m.slots[m.active].mode }

// Next moves the active highlight forward, skipping disabled slots. Wraps.
func (m Model) Next() Model {
    n := len(m.slots)
    for i := 1; i <= n; i++ {
        candidate := (m.active + i) % n
        if !m.slots[candidate].disabled {
            m.active = candidate
            return m
        }
    }
    return m
}

// Prev mirrors Next in reverse.
func (m Model) Prev() Model {
    n := len(m.slots)
    for i := 1; i <= n; i++ {
        candidate := (m.active - i + n) % n
        if !m.slots[candidate].disabled {
            m.active = candidate
            return m
        }
    }
    return m
}

// SetActive jumps to a specific Mode if it exists and is enabled; otherwise no-op.
func (m Model) SetActive(target Mode) Model {
    for i, s := range m.slots {
        if s.mode == target && !s.disabled {
            m.active = i
            return m
        }
    }
    return m
}

func (m Model) View() string {
    var parts []string
    for i, s := range m.slots {
        label := s.mode.String()
        switch {
        case s.disabled:
            parts = append(parts, tui.ModeBarDisabled.Render(label))
        case i == m.active:
            parts = append(parts, tui.ModeBarActive.Render(label))
        default:
            parts = append(parts, tui.ModeBarInactive.Render(label))
        }
    }
    return strings.Join(parts, lipgloss.NewStyle().SetString(" │ ").String())
}
```

##### Task 4 — `internal/faction/tui/views/modebar/modebar_test.go`

Unit test for cycling. Covers: initial Active is Manage; Next wraps past Quit back to Manage; Prev wraps past Manage to Quit; Spatial is skipped; SetActive(ModeSpatial) is a no-op; SetActive(ModeTurn) jumps.

```go
package modebar

import "testing"

func TestNewActiveIsManage(t *testing.T) {
    if got := New().Active(); got != ModeManage {
        t.Errorf("New().Active() = %v, want %v", got, ModeManage)
    }
}

func TestNextSkipsSpatialAndWraps(t *testing.T) {
    m := New()
    // Manage -> Turn
    m = m.Next()
    if m.Active() != ModeTurn {
        t.Fatalf("after Next: got %v, want %v", m.Active(), ModeTurn)
    }
    // Turn -> Quit (Spatial skipped)
    m = m.Next()
    if m.Active() != ModeQuit {
        t.Fatalf("after Next x2: got %v, want %v", m.Active(), ModeQuit)
    }
    // Quit -> Manage (wrap)
    m = m.Next()
    if m.Active() != ModeManage {
        t.Fatalf("after Next x3: got %v, want %v", m.Active(), ModeManage)
    }
}

func TestPrevWrapsAndSkipsSpatial(t *testing.T) {
    m := New()
    // Manage -> Quit (wrap)
    m = m.Prev()
    if m.Active() != ModeQuit {
        t.Fatalf("after Prev: got %v, want %v", m.Active(), ModeQuit)
    }
    // Quit -> Turn (Spatial skipped)
    m = m.Prev()
    if m.Active() != ModeTurn {
        t.Fatalf("after Prev x2: got %v, want %v", m.Active(), ModeTurn)
    }
}

func TestSetActiveDisabledIsNoOp(t *testing.T) {
    m := New().SetActive(ModeSpatial)
    if m.Active() != ModeManage {
        t.Errorf("SetActive(ModeSpatial): got %v, want %v (unchanged)", m.Active(), ModeManage)
    }
}

func TestSetActiveEnabledJumps(t *testing.T) {
    m := New().SetActive(ModeTurn)
    if m.Active() != ModeTurn {
        t.Errorf("SetActive(ModeTurn): got %v, want %v", m.Active(), ModeTurn)
    }
}
```

##### Task 5 — `internal/faction/tui/model.go`

Root `Model`. Holds active mode, sub-model map, mode bar, adapter, and overlay state. `Init` returns the adapter's observer pump `tea.Cmd` (Initiative 3 may extend; Foundation's pump receives nothing because the engine isn't running yet, but the pump command is correct to issue here for the architecture's sake).

Actually — Foundation does *not* start the observer pump at init time. Foundation never calls `adapter.Run()` from the TUI side (only the smoke does, and the smoke bypasses Bubbletea). So `Init` returns `nil`. Initiative 3 changes `Init` to issue both `Run` and `ObserverPump`.

```go
package tui

import (
    tea "github.com/charmbracelet/bubbletea"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/modebar"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/turn"
)

// Mode is an alias of modebar.Mode for convenience at the root layer.
type Mode = modebar.Mode

type Model struct {
    bar     modebar.Model
    subs    map[Mode]tea.Model
    adapter *adapter.Adapter

    // overlay state — Foundation only uses confirmExit.
    confirmExit  bool
    priorMode    Mode  // restored on Esc from confirm-exit overlay
    showHelpStub bool
}

func NewModel(adp *adapter.Adapter) Model {
    bar := modebar.New()
    return Model{
        bar:     bar,
        adapter: adp,
        subs: map[Mode]tea.Model{
            modebar.ModeManage: manage.New(),
            modebar.ModeTurn:   turn.New(),
            // ModeSpatial and ModeQuit are intentionally absent — they're
            // handled by root View directly, not dispatched.
        },
    }
}

func (m Model) Init() tea.Cmd { return nil }
```

##### Task 6 — `internal/faction/tui/update.go`

Root `Update`. Starts with the update discipline comment block (see Shared Context). Handles global keys (Tab, Shift-Tab, q, ?, Esc, Enter) and forwards everything else to `subs[bar.Active()]`. Special-cases the confirm-exit overlay state.

```go
package tui

// Update Discipline (per tui-rebuild-arc-discovery.md "Update Discipline").
//
// MAY:
//   - Forward messages to the active sub-model (subs[mode]).
//   - Handle global key bindings (Tab, Shift-Tab, q, ?, Esc) at the root.
//   - Update mode-bar state and switch the active mode.
//   - Update root-overlay state (help overlay, confirm-exit overlay).
//   - Return tea.Cmds for I/O performed by the adapter (engine launch, observer pump).
//
// MAY NOT:
//   - Call engine methods directly (engine.RunCycle, engine.RunFactionTurn, etc.).
//   - Mutate *state.FactionState or any *domain.* value.
//   - Perform I/O outside a tea.Cmd (filesystem reads, network calls, logging beyond
//     the structured logger's non-blocking calls).
//   - Embed game logic — turn ordering, action selection, mutation application all
//     belong to the engine. Update is a projection of engine state and a forwarder
//     of user intent.
//
// Violations are senior-engineer code review concerns at pre-merge time.

import (
    tea "github.com/charmbracelet/bubbletea"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/modebar"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Confirm-exit overlay intercepts all keys.
        if m.confirmExit {
            switch msg.String() {
            case "enter":
                return m, tea.Quit
            case "esc":
                m.confirmExit = false
                m.bar = m.bar.SetActive(m.priorMode)
                return m, nil
            }
            return m, nil // swallow other keys while overlay is up
        }

        switch msg.String() {
        case "tab":
            m.bar = m.bar.Next()
            return m.afterModeChange()
        case "shift+tab":
            m.bar = m.bar.Prev()
            return m.afterModeChange()
        case "q":
            m.priorMode = m.bar.Active()
            m.bar = m.bar.SetActive(modebar.ModeQuit)
            m.confirmExit = true
            return m, nil
        case "?":
            m.showHelpStub = !m.showHelpStub
            return m, nil
        case "enter":
            if m.bar.Active() == modebar.ModeQuit {
                m.priorMode = modebar.ModeManage // arbitrary; user explicitly chose Quit
                m.confirmExit = true
                return m, nil
            }
            // fall through to sub-model
        }
    }

    // Forward to active sub-model — unless mode is Quit or Spatial (no registered sub-model).
    active := m.bar.Active()
    if sub, ok := m.subs[active]; ok {
        updated, cmd := sub.Update(msg)
        m.subs[active] = updated
        return m, cmd
    }
    return m, nil
}

// afterModeChange handles the side-effects of a Tab/Shift-Tab cycle landing on
// a non-dispatched mode (Quit). For Foundation this is a no-op — the View
// branches on m.bar.Active() to render Spatial/Quit overlays.
func (m Model) afterModeChange() (Model, tea.Cmd) {
    return m, nil
}
```

##### Task 7 — `internal/faction/tui/view.go`

Root `View`. Composes: mode bar on top, then either the active sub-model's view, the Spatial placeholder, or the confirm-exit overlay. Help stub appears at the bottom if toggled.

```go
package tui

import (
    "fmt"
    "strings"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/modebar"
)

func (m Model) View() string {
    var sb strings.Builder

    sb.WriteString(m.bar.View())
    sb.WriteString("\n\n")

    if m.confirmExit {
        sb.WriteString(ConfirmExit.Render("Quit? Press Enter to confirm, Esc to cancel."))
    } else {
        switch m.bar.Active() {
        case modebar.ModeSpatial:
            sb.WriteString(Placeholder.Render("Spatial — reserved for F-012 (Spatial Map CLI)"))
        case modebar.ModeQuit:
            sb.WriteString(ConfirmExit.Render("Quit slot active. Press Enter to confirm."))
        default:
            if sub, ok := m.subs[m.bar.Active()]; ok {
                sb.WriteString(sub.View())
            } else {
                sb.WriteString(Placeholder.Render(fmt.Sprintf("Mode %v has no registered sub-model.", m.bar.Active())))
            }
        }
    }

    if m.showHelpStub {
        sb.WriteString("\n\n")
        sb.WriteString(HelpStub.Render("Help: Tab / Shift-Tab cycle modes  ·  q quit  ·  ? toggle help"))
    }

    return sb.String()
}
```

##### Task 8 — `internal/faction/tui/tui.go` (replace stub)

Replace the stub `Run` body with a real `tea.NewProgram` invocation. Constructs the adapter and the root Model.

```go
package tui

import (
    "log/slog"
    "os"

    tea "github.com/charmbracelet/bubbletea"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/state"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
)

// Run launches the TUI against a constructed engine and faction state.
func Run(eng *engine.Engine, factionState *state.FactionState, log *slog.Logger) error {
    adp := adapter.New(eng, factionState, log)
    defer adp.Stop()

    program := tea.NewProgram(NewModel(adp), tea.WithAltScreen())
    _, err := program.Run()
    return err
}

// init nudge — silence the unused-import linter if os isn't referenced yet.
var _ = os.Stdout
```

The `os` import / `var _ = os.Stdout` line is a tactical placeholder — drop the import once a real use shows up (likely Commit 5 if dryrun.go shares the package). If `os` isn't actually needed, remove the import and the placeholder line.

Update the Phase 1 stub's signature: now takes `*engine.Engine` and `*slog.Logger` as well. `cmd/gm-toolkit/faction.go` will be updated in Commit 5 to construct the engine and pass it through.

For Phase 3 to compile, the `cmd/gm-toolkit/faction.go` stub from Commit 1 must be updated minimally to match the new `Run` signature. Add a Task 9:

##### Task 9 — `cmd/gm-toolkit/faction.go` (sync signature)

Update the `factionRun` non-dryrun branch to construct stub args so the call compiles:

```go
func factionRun(cmd *cobra.Command, args []string) error {
    if factionDryRun {
        fmt.Fprintln(os.Stderr, "--dryrun not yet wired (Commit 5 replaces this branch).")
        return nil
    }

    // TODO Commit 5: construct real engine + factionState from config.
    return tui.Run(nil, nil, slog.Default())
}
```

Add the `log/slog` import. `nil` for `*engine.Engine` and `*state.FactionState` is fine — Foundation's root Model doesn't dereference them. (Commit 5 wires real construction.)

##### Commit message

```
feat(tui): root model with mode-bar dispatch and empty shells

- introduce mode-bar component with Tab/Shift-Tab cycling, disabled-slot skip,
  and a unit test covering the cycling logic
- ship empty Manage and Turn shells with placeholder banners
- root Model dispatches to subs[mode]; root View renders bar + sub-view + overlays
- root Update handles Tab, Shift-Tab, q (→ confirm-exit), ?, Esc, Enter
- ModeQuit and ModeSpatial render overlays in root View (not via subs)
- update discipline encoded as comment block at top of update.go
- wire tea.NewProgram in tui.Run
```

---

### Phase 4 — Dry-Run Smoke

Adds the auto-acking smoke collectors and the `--dryrun` execution path. After this commit `gm-toolkit faction --dryrun` constructs a trivial in-memory faction state, instantiates the real `Adapter`, and calls `engine.RunCycle` directly on the main thread with smoke collectors. Observer events print to stdout.

#### Commit 5 — `feat(tui): dry-run smoke exercises adapter against engine`

File-creation order:

##### Task 1 — `internal/faction/tui/adapter/smoke.go`

Auto-acking variants of `PhaseCollector` and `TurnObserver`. `action.Collector` reuses `stubActionCollector` from Commit 2 (the smoke invariant guarantees nothing reaches it).

```go
package adapter

import (
    "fmt"
    "log/slog"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/turn"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// SmokePhaseCollector auto-acks every PhaseCollector method. AwaitCheckpoint
// returns nil immediately; Select* methods return the first eligible option or
// a sensible default. The smoke's invariant is that all returns are safely
// consumable by the engine for the trivial single-faction cycle.
type SmokePhaseCollector struct{ log *slog.Logger }

func NewSmokePhaseCollector(log *slog.Logger) *SmokePhaseCollector { return &SmokePhaseCollector{log: log} }

func (s *SmokePhaseCollector) AwaitCheckpoint(phase string) error {
    s.log.Info("smoke: AwaitCheckpoint auto-ack", "phase", phase)
    return nil
}

func (s *SmokePhaseCollector) SelectAction(faction *domain.Faction, available []action.Action) (action.Action, error) {
    if len(available) == 0 {
        return nil, fmt.Errorf("smoke: no actions available — faction state shaped wrong")
    }
    s.log.Info("smoke: SelectAction", "faction", faction.ID, "choice", available[0].Name())
    return available[0], nil
}

func (s *SmokePhaseCollector) SelectStatRaise(faction *domain.Faction, eligible []domain.FactionStat) (*domain.FactionStat, error) {
    if len(eligible) == 0 {
        return nil, nil
    }
    s.log.Info("smoke: SelectStatRaise", "faction", faction.ID)
    return &eligible[0], nil
}

func (s *SmokePhaseCollector) SelectMovementDecisions(faction *domain.Faction, eligible []*domain.Asset) ([]world.MovementDecision, error) {
    s.log.Info("smoke: SelectMovementDecisions (no assets, returns empty)", "faction", faction.ID)
    return nil, nil
}

func (s *SmokePhaseCollector) SelectTransportCargo(transport *domain.Asset, eligibleCargo []*domain.Asset, profile *domain.TransportProfile) ([]*domain.Asset, error) {
    s.log.Info("smoke: SelectTransportCargo (no cargo, returns empty)")
    return nil, nil
}

// SmokeObserver prints each event to the structured logger.
type SmokeObserver struct{ log *slog.Logger }

func NewSmokeObserver(log *slog.Logger) *SmokeObserver { return &SmokeObserver{log: log} }

func (o *SmokeObserver) OnFactionTurnStarted(faction *domain.Faction)   { o.log.Info("EVT FactionTurnStarted", "faction", faction.ID) }
func (o *SmokeObserver) OnFactionSkipped(faction *domain.Faction)       { o.log.Info("EVT FactionSkipped", "faction", faction.ID) }
func (o *SmokeObserver) OnFactionTurnCompleted(faction *domain.Faction) { o.log.Info("EVT FactionTurnCompleted", "faction", faction.ID) }
func (o *SmokeObserver) OnStatRaiseSkipped(faction *domain.Faction)     { o.log.Info("EVT StatRaiseSkipped", "faction", faction.ID) }
func (o *SmokeObserver) OnGoalLockApplied(faction *domain.Faction, lock locks.GoalLock, muts []domain.Mutation) {
    o.log.Info("EVT GoalLockApplied", "faction", faction.ID, "lock", lock, "mutations", len(muts))
}
func (o *SmokeObserver) OnBookkeepingApplied(faction *domain.Faction, result turn.BookkeepingResult, muts []domain.Mutation) {
    o.log.Info("EVT BookkeepingApplied", "faction", faction.ID, "mutations", len(muts))
}
func (o *SmokeObserver) OnActionSelected(faction *domain.Faction, selected action.Action) {
    o.log.Info("EVT ActionSelected", "faction", faction.ID, "action", selected.Name())
}
func (o *SmokeObserver) OnActionResolved(faction *domain.Faction, selected action.Action, muts []domain.Mutation) {
    o.log.Info("EVT ActionResolved", "faction", faction.ID, "action", selected.Name(), "mutations", len(muts))
}
func (o *SmokeObserver) OnCycleCompleted(cycleNumber int, factionState *state.FactionState) {
    o.log.Info("EVT CycleCompleted", "cycle", cycleNumber)
}
func (o *SmokeObserver) OnError(faction *domain.Faction, err error) {
    o.log.Error("EVT Error", "faction", faction.ID, "err", err)
}
func (o *SmokeObserver) OnStatRaiseApplied(faction *domain.Faction, raised *domain.FactionStat, muts []domain.Mutation) {
    o.log.Info("EVT StatRaiseApplied", "faction", faction.ID, "mutations", len(muts))
}
func (o *SmokeObserver) OnMovementTicked(faction *domain.Faction, muts []domain.Mutation) {
    o.log.Info("EVT MovementTicked", "faction", faction.ID, "mutations", len(muts))
}
func (o *SmokeObserver) OnMovementResolved(faction *domain.Faction, muts []domain.Mutation) {
    o.log.Info("EVT MovementResolved", "faction", faction.ID, "mutations", len(muts))
}
func (o *SmokeObserver) OnIndexSkipped(skipped []string) {
    o.log.Info("EVT IndexSkipped", "count", len(skipped))
}

// Compile-time interface satisfaction checks.
var (
    _ engine.PhaseCollector = (*SmokePhaseCollector)(nil)
    _ engine.TurnObserver   = (*SmokeObserver)(nil)
)
```

##### Task 2 — `internal/faction/tui/dryrun.go`

`RunDryRun` constructs a trivial state, builds the engine, and calls `engine.RunCycle` with smoke collectors. Stays in the `tui` package so it shares the styles file and any common helpers.

```go
package tui

import (
    "log/slog"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/state"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
)

// RunDryRun executes a one-faction smoke cycle through the real engine using
// auto-acking smoke collectors and a logging observer. It bypasses Bubbletea
// entirely — the engine runs on the main thread.
//
// What this proves:
//   - PhaseCollector / action.Collector / TurnObserver interface satisfaction.
//   - engine.RunCycle's call shape works end-to-end with Foundation's adapter types.
//
// What this does NOT prove:
//   - Channel/goroutine plumbing (the smoke is single-threaded).
//   - Real action sub-engine collector behavior (the smoke invariant ensures
//     action.Collector methods are never called; stubActionCollector would
//     panic/error if they were).
func RunDryRun(log *slog.Logger) error {
    factionState, err := newSmokeFactionState()
    if err != nil {
        return err
    }

    eng, err := engine.New( /* args: verify engine.New signature at execution time */ )
    if err != nil {
        return err
    }

    collectors := engine.Collectors{
        Phase:  adapter.NewSmokePhaseCollector(log),
        Action: &smokeActionCollector{}, // reuses the stub; smoke invariant says it's unreachable
    }
    observer := adapter.NewSmokeObserver(log)

    log.Info("dryrun: starting RunCycle")
    if err := eng.RunCycle(factionState, collectors, observer); err != nil {
        log.Error("dryrun: RunCycle returned error", "err", err)
        return err
    }
    log.Info("dryrun: RunCycle completed cleanly")
    return nil
}

// newSmokeFactionState builds a trivial in-memory state: one faction, zero
// assets, all goals pre-locked, no action requiring action-sub-engine
// collector calls. See plan § "Smoke Invariant".
func newSmokeFactionState() (*state.FactionState, error) {
    // TODO at execution time: re-ground on state.FactionState's current shape
    // and constructor. Fields below are the minimum the engine needs; adjust
    // to match current struct.
    smokeFaction := &domain.Faction{
        ID:   "smoke-faction",
        Name: "Smoke Faction",
        // ... minimum required fields. Re-ground on current domain.Faction.
    }
    fs := &state.FactionState{
        Factions: []*domain.Faction{smokeFaction},
    }
    return fs, nil
}

// smokeActionCollector is a separate type from adapter.stubActionCollector to
// avoid leaking the unexported type. It satisfies action.Collector by
// embedding a no-op shim — actual smoke invariant says it's unreachable.
type smokeActionCollector struct{ adapter.stubActionCollectorExported } // see Task 1 follow-up
```

Note: The last paragraph above suggests exposing `stubActionCollector` via a wrapper. Simpler: in Task 1, also add an exported constructor `NewStubActionCollector() action.Collector` to the adapter package so `dryrun.go` can call it without re-implementing. Add that constructor to `action_collector.go`:

```go
// NewStubActionCollector returns the panicking/error-returning stub used by
// the smoke and (in Foundation) by the real adapter.
func NewStubActionCollector() action.Collector { return &stubActionCollector{} }
```

Then in `dryrun.go`:

```go
collectors := engine.Collectors{
    Phase:  adapter.NewSmokePhaseCollector(log),
    Action: adapter.NewStubActionCollector(),
}
```

Adjust the file body accordingly when implementing. (The plan reflects the cleaner shape after iteration.)

##### Task 3 — `cmd/gm-toolkit/faction.go` (wire `--dryrun`)

Replace the dryrun-branch stub with the real call:

```go
package main

import (
    "log/slog"
    "os"

    "github.com/spf13/cobra"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui"
)

var (
    factionDryRun bool
)

var factionCmd = &cobra.Command{
    Use:   "faction",
    Short: "Open the faction TUI",
    Long:  "Launches the faction-manager TUI. Use --dryrun to run a one-faction smoke cycle through the engine and exit (developer-only).",
    RunE:  factionRun,
}

func init() {
    factionCmd.Flags().BoolVar(&factionDryRun, "dryrun", false, "run a one-faction smoke cycle and exit")
    rootCmd.AddCommand(factionCmd)
}

func factionRun(cmd *cobra.Command, args []string) error {
    log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

    if factionDryRun {
        return tui.RunDryRun(log)
    }

    // TODO (Initiative 2): construct real engine + factionState from config and pass through.
    return tui.Run(nil, nil, log)
}
```

Verify:

```sh
go build ./cmd/gm-toolkit
./gm-toolkit faction --dryrun
# Should print: "dryrun: starting RunCycle" then a stream of EVT events then "dryrun: RunCycle completed cleanly".
./gm-toolkit faction
# Should launch the TUI; Tab/Shift-Tab cycle, q opens confirm-exit, Enter quits.
```

If the `RunCycle` call panics in `stubActionCollector` or `hooks.Collector` methods, the smoke faction state has violated the invariant — fix `newSmokeFactionState`, not the stubs.

##### Commit message

```
feat(tui): dry-run smoke exercises adapter against engine

- add smoke collectors and stdout-style observer in adapter/smoke.go
- introduce tui.RunDryRun: constructs trivial faction state, builds engine,
  calls engine.RunCycle directly on the main thread
- wire `gm-toolkit faction --dryrun` into RunDryRun
- export NewStubActionCollector so the smoke can reuse the foundation stub
- smoke proves interface and type alignment with the real engine; the
  channel/goroutine pattern remains proven by Initiative 3's first real turn
```

---

## Test Plan

- **Smoke (run manually after Commit 5):** `gm-toolkit faction --dryrun` — completes cleanly, prints expected EVT lines, exits zero. Any panic from `stubActionCollector` or `hooks.Collector` stub methods indicates the smoke faction state violates the invariant — fix the state, not the stub.
- **Unit test (Commit 4):** `modebar_test.go` covers Tab/Shift-Tab cycling, Spatial skip, SetActive enabled/disabled.
- **Compile-time:** every package compiles after every commit; `var _ <iface> = (*<impl>)(nil)` lines at the end of each implementer file enforce interface satisfaction at build time.
- **Visual (manual after Commit 4):** `gm-toolkit faction` launches; Tab cycles Manage → Turn → Quit (Spatial skipped); Shift-Tab reverses; `q` opens confirm-exit; `Enter` on Quit exits; `Esc` from confirm-exit returns to the prior mode; `?` toggles the help stub.
- **Out of test scope:** channel ops under contention, goroutine cleanup, `Esc`-cancel routing. Initiative 3 covers these.

## Pre-Merge Checklist Additions

Beyond the standard CLAUDE.md checklist (senior-engineer code review, dev journal, planned-work doc):

1. **Verify `gm-toolkit faction --dryrun` completes cleanly** as the final manual check.
2. **Confirm no production code references `stubActionCollector` or `hooks.Collector` stub methods reachably** — Initiative 3 deletes them; Foundation must not bake in any dependency.
3. **Confirm Update discipline comment block is present at the top of `update.go`** and accurately reflects what Update does and does not do.
4. **Confirm the `cancel sentinel`** (`ErrTurnCanceled`) is defined in `channels.go` with no use sites (Initiative 3 wires the routing).
5. **Confirm `SetPerFactionCadence`** is exposed on `*Adapter` and storage is `atomic.Bool` (Initiative 3 calls it from the execution view).

## Open Questions — To Ratify at Implementation Time

None deferred to Execution. Two re-grounding steps remain:

1. **`engine.New` constructor signature** — verified at Commit 5 task 2; the `engine.New( /* args: verify... */ )` placeholder is replaced with the current signature.
2. **`engine.CheckpointCycleSummary` constant name** — verified at Commit 3 task 2 (`AwaitCheckpoint`'s phase short-circuit); if the engine exports a different name for the cycle-summary checkpoint, swap it in.
3. **`state.FactionState` constructor / required fields** — verified at Commit 5 task 2 (`newSmokeFactionState`); if the struct has shifted since Plan, adjust the minimum field list.

These are not design choices; they are signature lookups at execution time. The design they implement is fixed.

# TUI Turn — Discovery

`F-005.3`, Initiative 3 of the [TUI Rebuild arc](../arcs/tui-rebuild/tui-rebuild-arc-plan.md). Companion to [`tui-rebuild-arc-discovery.md`](../arcs/tui-rebuild/tui-rebuild-arc-discovery.md), [`tui-foundation-discovery.md`](./completed/tui-foundation-discovery.md), and [`tui-manage-discovery.md`](./completed/tui-manage-discovery.md) — Turn inherits the engine boundary, root Model dispatch, mode-bar router, adapter sub-package, router sub-model + sub-model-per-view pattern, completion-message transition seam, help-overlay wiring, and empty-state grammar from those docs and does not re-litigate them. This doc ratifies the **initiative-specific** decisions Turn needs before Plan: how a faction turn is walked through the UI, the setup → execution view shape, the modal-overlay infrastructure that backs every collector prompt, the event-stream rendering, and the `Esc`-cancel path.

## Problem

Turn is the **engine-crossing** workflow — the mirror of Manage, which bypasses the engine to do plain CRUD on faction state. Where Manage reads and writes `FactionState` directly, Turn hands control *to* the engine and drives a whole faction cycle to completion.

The central tension is concurrency. `RunCycle` (orchestrator.go) executes synchronously: it loops every faction through a fixed phase sequence, calling `PhaseCollector` / `action.Collector` methods inline and **blocking** on each until the GM answers, while emitting `TurnObserver` callbacks as it mutates state. A TUI cannot block its `Update` loop, so Foundation's adapter runs the cycle on a **goroutine** and marshals everything across channels: observer callbacks drain onto a buffered event channel (the pump), and each blocking collector call sends an ask on an unbuffered channel and waits for a typed reply. The engine goroutine and the TUI render loop never touch shared state directly — only messages cross the boundary.

So a turn walk-through must surface three things the engine produces as it runs: **collector prompts as modal decisions** (the blocking asks — what the GM must answer for the cycle to proceed), **observer callbacks as a streaming event log** (the fire-and-forget play-by-play), and the **`Esc`-abort path** (a reply that is an error, which the orchestrator classifies and either recovers from or unwinds). `internal/faction/engine/orchestrator.go` is the authority on how a turn actually executes and grounds the prompt/event inventory below.

## Design Summary

Turn is built as a **router sub-model** at `internal/faction/tui/views/turn/`, mirroring Manage one level down: a `view` enum (`viewSetup`, `viewExecution`) with a sub-model per view. **Setup** shows a read-only roster and dispatches `Start`. **Execution** is a three-pane layout reusing Manage's Region split — left turn-order progress rail, center scrolling event stream, right processing-faction detail — with a **modal overlay** that floats over it for every collector prompt.

The engine is driven through Foundation's adapter, which already exists and already made the load-bearing channel decisions. `Adapter.Run()` launches `RunCycle` as a `tea.Cmd` on a goroutine; `ObserverPump()` re-issues a command that drains observer events into the stream. Collector prompts arrive on a **single discriminated channel** (`askCh chan CollectorAskMsg{Kind, Faction, Payload, Reply}`): the engine goroutine blocks sending an ask, the TUI receives it, opens the matching overlay, and sends the typed answer (or `ErrTurnCanceled`) back on the per-ask `Reply` channel.

The initiative ships in **three Efforts** (see Effort Breakdown). Effort 1 proves the whole loop against the real engine with placeholder/auto-answering overlays — no real decisions, but every seam in place. Efforts 2 and 3 fill in real modals additively, without restructuring Effort 1's plumbing.

## Decisions Ratified in Discovery

1. **Turn ships in three Efforts** — (1) Turn-flow scaffold, (2) phases except action, (3) the action phase — each with its own Plan, under this one Discovery doc. Mirrors the `asset-movement-redesign` / `logging-and-errors` multi-Effort precedent.
2. **Effort 1 drives the REAL engine (Option A), not a static shell.** `Start` dispatches `Adapter.Run()` → `RunCycle` against the real adapter (real `PhaseCollector` + `observer` from Foundation). Every prompt is auto-answered with a default (ack / nil / empty), so a real cycle runs to completion while the GM makes no meaningful decisions. This is the only split that lets Efforts 2 & 3 add modals without refactoring the consumption loop or event stream.
3. **`SelectAction` belongs to Effort 3, not Effort 2.** It is a `PhaseCollector` method, but semantically it *is* the action phase; making it real before `action.Collector` exists would panic mid-turn. In Efforts 1–2 it auto-returns `nil` — a legal "skip faction" path (orchestrator.go:418), which keeps the stubbed `action.Collector` unreachable until Effort 3 deliberately reaches it.
4. **The unified discriminated ask-channel is inherited from Foundation, not re-litigated.** One `askCh` carrying `CollectorAskMsg{Kind, Payload, Reply}`. Effort 3's action prompts are new `AskKind` values + payload types + a real `action.Collector` calling the same `ask()` helper — purely additive at the adapter layer. No per-collector channels.
5. **The load-bearing seam for "no refactoring" is the TUI-side dispatch on `CollectorAskMsg.Kind`.** Effort 1 enumerates all current ask kinds in a switch, each routing to a placeholder overlay; Efforts 2 & 3 swap the placeholder for a real overlay per kind (and Effort 3 adds new kinds). Adding/replacing a branch is filling-in, not restructuring.
6. **`Start` invokes `RunCycle`** (closes OQ1). The TUI does not loop `RunFactionTurn` itself; `Adapter.Run()` drives the whole `RunCycle` (run.go:9) in one call.
7. **Effort 3 ships all actions, but each action is its own self-contained, non-blocking commit** (closes OQ3). Effort 3 opens with an **action-flow foundation commit** (real channel-backed `action.Collector` + `SelectAction` modal, all interface methods present but safe-stubbed — **refined by Decision 14**: the *shared* prompts (`SelectAsset`, the two hook modals) are made real in this foundation commit; only action-*unique* methods stay stubbed until their action's commit), then one commit per action replaces the stubs for that action's collector-method set. An unimplemented action remains a legitimate landing state. **Requirement this imposes:** the not-yet-built path must be *recoverable and surfaced*, not the current behavior (`ErrActionNotImplemented` is unclassified; the two `hooks.Collector` methods `panic`). Preferred: the `SelectAction` modal **disables unimplemented actions** so they can't be selected. Action→method mapping and the implemented-action registry are Effort-3 Plan detail.
8. **Errors surface in the persistent contextual bottom bar, not a modal** (closes OQ on error rendering). The modal overlay is reserved for collector prompts (things the engine blocks on); errors are notifications, not decisions. Every error is recorded in the event stream (`EvtError`, rendered from Effort 1 onward — the permanent log). The bottom bar — today's toggle-gated help footer (`view.go:65`), reworked into an **always-present, priority-resolved contextual bar** (app-wide root chrome) — mirrors the **latest** error, color-coded recoverable (cycle continued) vs. fatal (cycle halted via `EngineDoneMsg{Err}`). Content priority: a live error outranks help; otherwise the bar shows help. The `?` toggle now switches help **compact ↔ full**, not the bar's existence. Turn contributes error content via a `StatusLiner` interface on the active sub-model, paralleling the existing `Helper` check (`view.go:67`). See *Contextual bottom bar*.
9. **Logging is app-wide, constructed from `logging.New(paths.LogsDir, debug)` in the `faction` command** (closes OQ on logging). A `--debug` bool flag is added to `cmd/gm-toolkit/faction.go`; the logger replaces the current stderr `slog` handler (`faction.go:28`) and is built **inside `factionRun` after `paths` resolves** (`faction.go:51`), then threaded into `engine.NewWithRulebook` and `tui.Run` (both already accept `*slog.Logger`, so engine-cycle and TUI logs share one campaign file). Pre-campaign errors (registry/active-campaign resolution) stay on stderr. Lands early in Effort 1 since it benefits the whole app, not just Turn.
10. **`Turn.Start` is folded into `Adapter.Run()`** (Effort 1; closes the setup-dispatch gap). `RunCycle` requires the turn order to already be rolled (orchestrator.go:31-32), but Foundation's `Run()` never calls `Turn.Start` — so a real cycle would fail at `CurrentFaction`. `Run()` is changed to roll the order itself: **not `InProgress` → `Turn.Start` then `RunCycle`; already `InProgress` → skip Start and resume.** The setup view's Start button just dispatches `Run()` and stays engine-ignorant. This is the minimal "don't re-roll a live turn" guard; full pause/resume UX remains Out of Scope.
11. **The overlay seam — `Overlay` interface + execution-owned dispatch — is ratified here** (makes Decision 5's "no churn" concrete). An overlay is a pure `tea.Model` emitting `OverlayDoneMsg{Answer any}`; it never touches the adapter channels. The **execution sub-model** owns the optional `overlay` + the pending `Reply chan<- any`, dispatches `CollectorAskMsg.Kind` through one factory switch, and does the channel send on completion. The overlay **floats** over the three panes (internal state, not a router peer view). Effort 1 backs every branch with one keypress-gated generic placeholder; Efforts 2/3 swap/add branches additively. See Modal overlay infrastructure for the full contract.
12. **Turn runs whole-cycle only; per-faction cadence is dropped from the initiative** (deferred). No setup cadence choice, no mid-cycle toggle, no cadence indicator — the control surface wasn't worth its cost. The adapter's `perFactionCadence` flag stays at its zero value, unused. The only checkpoint that pauses is `cycle_summary` (the adapter forces it, phase_collector.go:28); the `Select*` asks fire regardless of cadence, so the overlay seam is unaffected. See Whole-cycle only.
13. **Effort 2 modals are `huh.Form`-backed overlays** (closes Effort 2 modal design). `SelectStatRaise` → single-select + in-modal "Decline" (never `Esc`); `SelectMovementDecisions` → a **single multi-group form** (one row per movable asset: kind select + conditional destination-**world** select; destination is world-granular — the engine derives hex/path); `SelectTransportCargo` → capped multi-select per transport; `AwaitCheckpoint(cycle_summary)` → bare ack (the only checkpoint that pauses). The right detail pane becomes **phase-reactive**. See *Effort 2 modals*.
14. **Effort 3 builds a reusable archetype toolkit, and the action surface is a star dependency** (closes Effort 3 modal design). The ~19 action prompts collapse into six generic `Overlay` archetypes (select / multiselect / confirm / two-step / numeric / wizard) instantiated per prompt with an injected context-renderer. The per-action inventory (verified against `register.go` + each action) proves the **only** cross-action prompt is `SelectAsset` (Buy + Sell); it plus the cross-cutting hooks (`SelectModifiers` / `ConfirmReroll`, made **real** in the foundation commit since they fire during any roll) are hoisted into the foundation commit, making every per-action commit self-contained and non-blocking. The two `hooks.Collector` methods return no `error`, so `Esc` inside a hook modal degrades to the legal no-op, not a turn-cancel. The **per-action contract** (prompt set, archetype, answer, decision-relevant context) is locked in Discovery; exact per-prompt fields/layout are Effort-3 Plan detail. Effort 3 ships **all 12** registered actions (Repair Faction + Abandon Goal are zero-prompt). See *Effort 3 modals → Per-action contract*. (Refines Decision 7's "all methods safe-stubbed" — shared prompts are real in foundation; only action-unique methods are stubbed until their commit.)

## Effort Breakdown

### Effort 1 — Turn-flow scaffold (drives the real engine; no real decisions)

The structural Effort. Lands a fully-wired turn that runs end-to-end with every decision defaulted.

- `turn` router sub-model (`viewSetup` / `viewExecution`) + the two view packages.
- Execution layout: three panes (left turn-order rail, center event stream, right faction detail) reusing Manage's promoted Region helper; overlay mount over the lot.
- **Ask-pump** `tea.Cmd` reading `askCh` → surfaces `CollectorAskMsg` as a `tea.Msg`. *New work: Foundation's smoke exercise used the inline `Smoke*` collectors, so the real `askCh` round-trip (block-send → receive → reply) is proven for the first time here. There is an `ObserverPump` but no symmetric ask-pump yet.*
- **Dispatch skeleton**: `switch CollectorAskMsg.Kind` with all 5 current kinds enumerated; each routes to a placeholder overlay that auto-answers the default and sends it on `msg.Reply`.
- Event-stream renderer covering all 14 observer events (they all fire on a real cycle), via the hybrid header + key-mutation `Mutation → string` formatter (Decision 8 / *Event stream rendering*).
- **Contextual bottom bar** (app-wide root chrome, Decision 8): rework `view.go`'s toggle-gated help footer into the always-present, priority-resolved bar (error > help; `?` = compact↔full help); add the `StatusLiner` interface. Lands early alongside logging since it touches the root.
- Whole-cycle only (Decision 12): no cadence control. The adapter's `perFactionCadence` stays at its zero value; the cycle streams straight through, pausing only for the `Select*` decisions and once at `cycle_summary` (phase_collector.go:28).
- Empty-state (no factions) in setup, inheriting Manage's grammar.
- `Adapter.Stop()` safety: do not close `eventCh` while a cycle is live (the close-during-emit panic risk Manage deferred here).
- **Outcome:** launch Turn → Start → keypress through each prompt → watch a real cycle stream events to completion, every decision auto-defaulted.

### Effort 2 — Phases except action

Swap Effort 1's placeholder overlays for real modals on the four non-action phase kinds; enrich the faction-detail pane.

- Real overlays: `AwaitCheckpoint` (ack), `SelectStatRaise`, `SelectMovementDecisions`, `SelectTransportCargo`.
- Per-phase enrichment of the right faction-detail pane (Effort 1 ships the basic identity + stats card; Effort 2 makes it reflect the active phase).
- Additive: swaps overlay constructors per `Kind`; no plumbing change.

### Effort 3 — The action phase

The largest Effort. Ships **all 12** registered actions, structured so each action is an independent, non-blocking commit — any action may be left unimplemented without blocking the others. The ~19 prompts are built as a reusable **archetype toolkit**, and the per-action prompt inventory (see *Effort 3 modals → Per-action contract*) proves a **star** dependency graph: every action depends only on the foundation commit, never on a sibling.

**Commit shape:**
1. **Action-flow foundation commit.** Ships everything *shared*: the archetype toolkit; the real `PhaseCollector.SelectAction` overlay (flips the Effort-1 auto-skip into real selection) + the disable-unimplemented set; the two hook modals (`SelectModifiers` / `ConfirmReroll`) made **real** here since they fire cross-cutting during any rolling action; **`SelectAsset`** (the only prompt shared between actions — Buy + Sell); `Esc`-cancel routing (`ErrTurnCanceled`, which Foundation's channels.go:101 comment defers to here); and the two zero-prompt actions (Repair Faction, Abandon Goal) registered by `Name()` only. The real channel-backed `action.Collector` replaces `stubActionCollector`; action-*unique* methods remain **safe-stubbed** (recoverable, surfaced — never `panic`, never unclassified-fatal) until their action's commit.
2. **One commit per remaining action**, instantiating archetypes for *only that action's unique prompts* and adding its `Name()` to the implemented set. Per the verified inventory: Attack (`SelectAttackers` / `SelectDefender` / `ConfirmRedirectToBase`), Expand Influence (`SelectExpandInfluenceOrder` / `ConfirmRivalFreeAttack` / `SelectBaseAttackers`), Buy (`SelectBuyOrder` + shared `SelectAsset`), Refit (`SelectRefitOrder`), Repair Asset (`SelectRepairOrders`), Use Asset Ability (`SelectAbilityAssets` / `ConfirmAbilityApplied` / `SelectFactionTestTarget`), Bribe (`SelectBribeTarget`), Seize Planet (`SelectSeizeTarget`), Change Homeworld (`SelectChangeHomeworldTarget`), Sell Asset (shared `SelectAsset` only).

**Plan-level detail (deferred to Effort 3 Plan):** the exact per-prompt fields/layout (the contract — prompt set, archetype, answer, decision-relevant context — is locked in *Per-action contract*); the implemented-action registry shape; and the safe-stub classification for not-yet-built action-unique methods.

Additive at both the adapter (new `AskKind` values + payloads) and the TUI dispatch (new switch cases) — Efforts 1 & 2 are untouched.

## Open Questions

All four open questions raised in this Discovery are resolved — see Decisions Ratified:

- OQ1 (`Start` → `RunCycle`) → Decision 6.
- OQ "`action.Collector` scope" → Decision 7.
- OQ "error rendering shape" → Decision 8.
- OQ "logging scope / construction point" → Decision 9.

None remaining at the Discovery level.

## Remaining Discovery Work

Design is **complete** across all three Efforts:
- **Effort 1 UI** — Q1–Q7 / Decisions 10–12, in *Setup view*, *Execution view*, *Event stream rendering*, *Modal overlay infrastructure*, *Contextual bottom bar*.
- **Effort 2 modals** — Decision 13, in *Effort 2 modals*.
- **Effort 3 modals** — Decision 14, in *Effort 3 modals* (archetype toolkit + per-action contract + commit shape).
- **Stub subsections** — `Problem`, `Turn view stack and sub-model layout`, `Help overlay wiring`, `State Storage`, `User-Facing Impact` all filled.

**Discovery is complete and the restructuring pass is done — this doc is ready to drive Plan.** Grounding subsections (turn pipeline + adapter contract) lead the *Per-Area Design Details*, with the design subsections following.

**Overall initiative flow (for orientation):** **one Discovery doc** (this file) → then **an Effort Overview doc + 3 Effort Plan docs** (Plan mode, one session per doc) → then execution. All Effort design lands *here*; the per-Effort splitting happens only at the Plan stage.

---

## Per-Area Design Details

The first two subsections are **grounding** — the turn pipeline and the adapter contract, verified against engine and adapter source. Everything after is **design**.

### Turn pipeline inventory (from `orchestrator.go`, verified 2026-05-29)

This is the grounding for "how we walk through a turn." `RunCycle` loops `RunFactionTurn` until the turn cursor wraps. `RunFactionTurn` drives one faction through a fixed phase sequence. The pipeline is **synchronous**: collector methods are called inline and block; the adapter runs the whole thing on a goroutine, so a blocking collector call is exactly where the UI must open a prompt.

**Per-faction phase sequence:**

| # | Phase / step | GM input (blocking) | Observer events emitted |
|---|---|---|---|
| 0 | `setupFactionTurn` | — | `OnIndexSkipped` (if any), `OnFactionTurnStarted` |
| 1 | `runGoalLockPhase` | `AwaitCheckpoint(goal_locked)` — **only if `LockSkip`** | `OnGoalLockApplied` |
| 2 | `runStatRaisePhase` | `SelectStatRaise` (only if eligible; may decline → nil) | `OnStatRaiseApplied` / `OnStatRaiseSkipped` |
| 3 | `runBookkeepingPhase` | `AwaitCheckpoint(bookkeeping)` | `OnBookkeepingApplied` |
|   | *state save (gate)* | — | `OnError` on failure |
| 4 | `runMovementPhase` | `SelectMovementDecisions`, then `SelectTransportCargo` (per transport issuing a move); `AwaitCheckpoint(movement)` | `OnMovementTicked`, `OnMovementResolved` |
| 5 | `runActionPhase` — **skipped entirely if `LockSkip`** | `SelectAction` (may skip → nil); **+ the entire `action.Collector` surface** fires inside `Action.Run`; `AwaitCheckpoint(action_result)` | `OnActionSelected`, `OnActionResolved` / `OnFactionSkipped` |
|   | *state save (gate)* | — | `OnError` on failure |
| 6 | `finishFactionTurn` | `AwaitCheckpoint(cycle_summary)` — **only on the faction that closes the cycle** | `OnFactionTurnCompleted`; `OnCycleCompleted` (cycle close only) |

**Prompt surface = TWO collectors.** The modal overlay must back both:

1. **`PhaseCollector`** (5 methods, orchestrator-owned): `AwaitCheckpoint(phase)` — single-key ack; `SelectAction` — pick-one-or-skip; `SelectStatRaise` — pick-one-or-decline; `SelectMovementDecisions` — multi-decision; `SelectTransportCargo` — multi-select cargo.
2. **`action.Collector`** (~18 methods + embedded `hooks.Collector`'s `SelectModifiers` / `ConfirmReroll`), threaded into the sub-engines during `runActionPhase`. Fires *conditionally* on which action the GM selects: e.g. `SelectAttackers` / `SelectDefender` / `ConfirmRedirectToBase` (Attack), `SelectBuyOrder` (Buy Asset), `SelectRefitOrder` (Refit), `SelectExpandInfluenceOrder` (Expand Influence), `SelectBribeTarget` / `SelectSeizeTarget` / `SelectChangeHomeworldTarget`, `SelectRepairOrders`, ability prompts (`SelectAbilityAssets`, `ConfirmAbilityApplied`, `SelectFactionTestTarget`), and the two hook prompts.

**Observer surface** (14 methods) is fire-and-forget output → the scrolling event stream. `OnError` is the one needing a rendering decision (resolved — Decision 8, *Contextual bottom bar*).

### Adapter — implementation status and consumption contract

Verified against `internal/faction/tui/adapter/` (2026-05-29). Foundation built the adapter; Turn consumes it. Two facets ground the work: which collector/observer methods are already real vs. stubbed, and the channel contract Turn drives the cycle through.

**Implementation status (Foundation vs. Turn).**

- **`phaseCollector` — REAL.** Channel-backed; all 5 methods implemented (`phase_collector.go`).
- **`observer` — REAL.** All 14 methods implemented; feeds the event channel (`observer.go`).
- **`stubActionCollector` — STUB.** Every `action.Collector` method returns `ErrActionNotImplemented`; the two embedded `hooks.Collector` methods **panic** (`action_collector.go`). The dry-run state is shaped so none are reachable.

**Implication for scope.** Making the action phase actually interactive means Turn must build a **real, channel-backed `action.Collector`** — ~20 methods, each with its own modal shape. This is the single largest piece of Turn work and is *not* called out in the arc-plan's Initiative 3 entry (which only mentions `Select*` prompts generically). Scoped by Decision 7 (the action-phase Effort).

**Consumption contract (driving the cycle).** The contract Turn consumes:

- **Start** → `Adapter.Run() tea.Cmd` (run.go:7) runs `RunCycle` on the bubbletea command goroutine and returns `EngineDoneMsg{Err}` on completion.
- **Events** → `Adapter.ObserverPump() tea.Cmd` (run.go:16) reads one `ObserverEventMsg` from the buffered `eventCh` (cap 64) and returns it; the root Update re-issues it after each receipt to keep the pump alive. Returns `nil` when `eventCh` closes.
- **Asks** → `askCh chan CollectorAskMsg` is **unbuffered**; the engine goroutine blocks on send until the TUI receives. **Turn must build the symmetric ask-pump** — there is no equivalent of `ObserverPump` for asks yet, and Foundation's smoke test bypassed `askCh` entirely by wiring the inline `Smoke*` collectors. So Effort 1 is the first code to exercise the real ask round-trip.
- **Reply** → each `CollectorAskMsg` carries `Reply chan<- any` (buffered cap 1, created per-ask in `phaseCollector.ask`). The TUI sends the typed answer; the collector type-asserts it (e.g. `raw.(action.Action)`). Sending an `error` value (e.g. `ErrTurnCanceled`) makes the collector return that error — this is the cancel hook.

### Turn view stack and sub-model layout

Mirrors Manage one level down (`views/manage/manage.go`): a router `Model` with a `view` enum and a sub-model per view. Package sketch under `internal/faction/tui/views/turn/`:

```
turn/
  turn.go        router Model: view enum {viewSetup, viewExecution}, dispatch, Help, CapturesInput
  messages.go    Turn-local tea.Msg types (CollectorAskMsg surfaced from the ask-pump,
                 EngineDoneMsg, cycle-started msg, OverlayDoneMsg)
  setup/         setup.Model — read-only roster + Start (dispatches Adapter.Run)
  execution/     execution.Model — three-pane layout + overlay ownership + StatusLiner
  overlay/       the Overlay archetype toolkit (Effort 3: select/multiselect/confirm/…);
                 Effort 1 ships only the generic placeholder here
```

The Region split helper is **not** local to `turn` — it's the promoted shared `tui` layout package both Manage and Turn import (*Execution view*).

`turn.Model` fields (sketch; exact shape is Plan detail):

```go
type Model struct {
    factionState *state.FactionState
    paths        *campaigns.Paths
    rulebook     *rulebook.Rulebook
    log          *slog.Logger

    adapter   *adapter.Adapter   // constructed at Start; drives RunCycle on a goroutine
    view      view               // viewSetup | viewExecution
    setup     setup.Model
    execution execution.Model

    termWidth, termHeight int
}
```

The router owns the adapter (built when Setup dispatches Start) and threads its channels — `ObserverPump`, the ask-pump, `EngineDoneMsg` — into `execution`. `execution.Model` holds the live-cycle state: the captured `FactionOrder` + cycle number (from the cycle-started message), the turn-order rail, the event-stream `viewport`, the snapshot-on-event detail copy, the optional `overlay Overlay` + `pendingReply chan<- any`, and the latest error for `StatusLiner`. Setup stays engine-ignorant.

### Setup view — roster + Start

Setup is a launch gate, not an editor — two elements, both light:

- **Read-only roster** — reuses Manage's list-item format verbatim: faction name + `scale · HP x/y · Coin z` (`list.go:26-27`), sorted **alphabetically**. It is display-only: there is **no per-faction opt-out or reorder**, because the engine has no such capability (`RunCycle` runs whatever is in state). Crucially, the roster does **not** show turn order — order is a random rotation rolled inside `Turn.Start` (`order.go`, `turn.go:50`) and does not exist until the cycle begins; the real order is revealed by the `OnFactionTurnStarted` events streaming into the execution event log.
- **Start** — dispatches the cycle (see below).

**Empty-state ("no factions")** inherits Manage's grammar (`list.go:115` — centered "No factions yet." over a dim hint), with the hint pointing the GM back to Manage to create factions rather than Manage's local `n`.

**Start dispatch (Decision 10).** `Turn.Start` (the order-roll + `CycleNumber++`) is **folded into `Adapter.Run()`**, not called by the setup view. Today `Run()` calls `RunCycle` directly and never rolls the order (Foundation's smoke path did it inline in `dryrun.go:47`), so as wired a real cycle would fail at `CurrentFaction` with `ErrNoTurnActive`. Effort 1 changes `Run()` to: **if the turn is not already `InProgress` → call `Turn.Start` (roll order, increment cycle), then `RunCycle`; if it is `InProgress` → skip Start and resume.** This keeps the setup view ignorant of engine internals (it just dispatches `Run()`), centralizes start-vs-resume in the adapter, and makes a persisted in-progress turn resume automatically. Full pause/resume *UX* stays out of scope — this is only the minimal "don't re-roll a live turn" guard.

### Execution view — three-pane layout

**Three columns mirroring Manage's Region split** (`layout.go` — `Left` ¼ · `Center` ½ · `Right` ¼, vertical-rule separators). The `Region`/`compose`/`regionWidths` helper is **promoted out of `package manage` into a shared `tui` layout package** that both Manage and Turn import (small extraction, touches Manage).

- **Left (¼) — turn-order progress rail.** The cycle's factions in *turn order*, each marked **done / current (highlighted) / upcoming**. This replaces setup's alphabetical roster once the cycle is live — alphabetical there because no order exists yet; turn-ordered here because it now does. The current marker advances on each `OnFactionTurnStarted`.
- **Center (½) — event stream.** Scrolling log fed by the observer callbacks; auto-follow / scroll-locks-tail (Q3).
- **Right (¼) — processing-faction detail.** Identity + scale + Force/Cunning/Wealth + HP + Coin. A compact "who's acting and how strong" card. (Cycle number already lives in the app title bar, view.go:26 — not repeated here.)

**Data discipline — snapshot-on-event.** Every observer event carries a `*domain.Faction` that is a **live pointer** into the engine-mutated `factionState`. The execution model never renders through that pointer (the engine mutates it on its own goroutine — a data race). Instead, when an event arrives the model **copies the display fields** (name, scale, stats, HP, Coin) into its own value and renders from that. Events are emitted right after their mutations land, so the copy is current; nothing shared is held across frames.

**Turn-order rail data (Effort-1 plumbing).** The rail needs the full `FactionOrder`, which `Turn.Start` now rolls inside `Adapter.Run()` on the command goroutine (Decision 10). The model cannot read `CurrentTurn.FactionOrder` directly (race). The adapter instead surfaces a **copy** of the order (+ cycle number) as a **cycle-started message** emitted at the top of `Run()` before the engine loop proceeds — the channel send is the happens-before. The model captures it once; subsequent `OnFactionTurnStarted` events only move the current marker. Exact message type and channel are Plan detail.

**Overlay mount.** Collector prompts float over this layout as a modal overlay (Q4/Q5).

### Event stream rendering

The center pane is the live play-by-play — terse, deterministic, one entry per observer event as it fires.

- **Widget: `bubbles/viewport`.** Scroll, follow, and key handling come for free; no custom buffer.
- **Auto-follow / scroll-locks-tail (arc-plan behavior, confirmed).** The viewport auto-scrolls to the tail as events arrive. Scrolling up manually **locks** the tail (suspends auto-follow so the GM can read history mid-cycle); `End` / `G` re-engages auto-follow and jumps back to live.
- **Hybrid rendering — header + key mutations.** Each observer event renders as a **header line** (faction + phase/result phrase) followed by a few **indented key-mutation lines**. This requires a **`Mutation → string` formatter built in Effort 1** (the ~30 types in `mutation.go`). The formatter surfaces *player-meaningful* deltas — faction/asset/base HP, Coin, asset added/removed, stat raised, goal completed/abandoned, homeworld changed, movement issued/completed — and **folds or omits bookkeeping noise** (`GoalTurnsTick`, `GoalProgressed`, `AssetMaintainedFlag`, the stealth flags). The exact "key" set and per-type phrasing are Plan detail.
- **Header phrasing** covers all 14 events: turn started / skipped, goal-lock applied, stat raise applied / skipped, bookkeeping applied, movement ticked / resolved, action selected / resolved, faction skipped, turn completed, cycle completed, index skipped. `OnError` → a distinct **error-styled** line (Decision 8 — the permanent record; the contextual bottom bar mirrors the latest).
- **Not** routed through the narrative renderer (that's digest-driven post-cycle prose; reserved for the deferred "review last cycle" view). The formatter is the execution view's own concern — likely a small helper in the turn execution package (Plan detail).

### Contextual bottom bar (Decision 8)

The old toggle-gated help footer (`view.go:65`) is reworked into a **persistent, always-present bar** whose content is **resolved by priority** for the current situation. This is **app-wide root chrome** (every mode renders through `view.go`), so it lands in Effort 1 alongside the other app-wide work (logging). Errors are notifications, not decisions, so they never use the modal overlay.

**Content priority** (highest wins):
1. **Live error** — the latest, color-coded recoverable (cycle continued) vs. fatal (cycle halted; arrives as `EngineDoneMsg{Err}`). Always visible while live, regardless of the help toggle.
2. **Help** — the default content. The `?` toggle switches help **compact ↔ full** (short single-row vs. expanded multi-row), *not* the bar's existence. The bar is always at least showing compact help.

The bar is always rendered (one persistent row), so the three-pane layout never shifts when an error appears or clears.

**Wiring.** The root resolves the bar each frame. It already pulls the active sub-model's keymap via the `Helper` interface (`view.go:67`); it gains a parallel **`StatusLiner`** check — an optional interface (e.g. `StatusLine() (text string, severity Severity)`) that the active sub-model implements. Turn's execution sub-model returns the latest error (from `EvtError` / `EngineDoneMsg`); when the returned text is non-empty it outranks help. Modes that don't implement `StatusLiner` simply contribute nothing, and the bar falls through to help.

**Permanent record (unchanged).** Every `OnError` → `EvtError` → a line in the event stream (the durable log); the bar only mirrors the *latest*.

### Modal overlay infrastructure (Decision 11)

The seam Efforts 2 & 3 build on — the contract that makes them purely additive (Decision 5's "no churn" promise, now concrete).

**Contract — `Overlay` interface.** An overlay is a pure `tea.Model` (Init/Update/View/Help) that, on completion, emits an `OverlayDoneMsg{Answer any}` via `tea.Cmd` — exactly as Manage's `deleteconfirm` emits `DeletedMsg`/`CancelMsg` (deleteconfirm.go:41). Overlays **never import the adapter's channels** and never see `Reply`; they only know how to render a prompt and emit a typed answer. (Cancel is a variant — an `OverlayDoneMsg` carrying `ErrTurnCanceled`, or a sibling message; Effort 3 detail.)

**Ownership — internal state on the execution sub-model.** Execution holds `overlay Overlay` (nil when no prompt is up) and `pendingReply chan<- any` (the `Reply` from the active `CollectorAskMsg`). The overlay **floats**: execution's `View` composites it over the three panes (panes dimmed behind). It is *not* a peer view in the turn router — a router view would be a full swap (deleteconfirm style) and couldn't float over live panes.

**Dispatch — the one switch that must not churn.** On a `CollectorAskMsg`, the execution `Update`: (1) stores `msg.Reply` as `pendingReply`; (2) constructs the overlay via a `switch msg.Kind` factory enumerating **all 5 current kinds**; (3) mounts it. While an overlay is present, keys route to it. On `OverlayDoneMsg`, `Update` sends `msg.Answer` on `pendingReply`, then clears both. Swapping a branch from placeholder to real overlay (Effort 2), or adding a branch for a new kind (Effort 3), is filling-in — never restructuring.

**Effort 1 placeholder.** All 5 branches point at **one generic placeholder** component (no per-kind files yet). It **renders and waits for a keypress** — surfaces which ask fired for which faction, and on a key sends that kind's default answer on the `Reply`. This makes the ask round-trip *visible* (Effort 1 is the first code to exercise the real `askCh`) and paces the cycle. Defaults per kind (the legal "no decision" answer): `AwaitCheckpoint` → ack; `SelectAction` → nil (skip faction, orchestrator.go:418); `SelectStatRaise` → nil (decline); `SelectMovementDecisions` → no decisions; `SelectTransportCargo` → empty selection. Exact zero-values are Plan detail.

### Effort 2 modals — phase-prompt UX

Effort 2 swaps the Effort-1 generic placeholder for a real overlay on each of the four non-action phase kinds. All four are built as **`huh.Form`-backed `Overlay` implementations** following the wizard idiom (`wizard.go`): `huh.NewSelect` / `huh.NewMultiSelect` with `Validate`, themed via `styles.FormTheme()`, emitting `OverlayDoneMsg{Answer}` on `huh.StateCompleted`. The orchestrator re-validates caps and eligibility (cap → orchestrator.go:672, ineligible cargo → :682), so in-modal validation is for UX, not integrity.

**`SelectStatRaise` — single-select + decline.** `huh.NewSelect[*domain.FactionStat]` over `eligible`, each option labeled with the stat and its current rating (so the GM knows what they're bumping), plus an explicit **"Decline raise"** sentinel mapping to `nil`. Decline is an in-modal option, never `Esc` (`Esc` = turn-cancel, Effort 3 — overloading it here would collide). Answer type: `*domain.FactionStat`.

**`SelectMovementDecisions` — single multi-group form** (chosen over per-asset-sequential and roster-drill-down). One `huh.Group` per eligible (`Speed > 0`) asset: a **kind** `Select` {None, Issue, Revise, Cancel} plus a destination-**world** `Select` shown only when the kind is Issue/Revise (conditional field visibility — `huh`'s `WithHideFunc`-style gate keyed on the row's kind; the first conditional-field form in the codebase, so it's a Plan detail, not assumed free). Revise/Cancel are offered only for assets that already hold a `CurrentOrder`; assets left at None are omitted from the returned slice. Submit assembles `[]world.MovementDecision`. **Destination granularity is world** — the engine derives the hex and runs `spatialMap.Path` itself (movement.go:82-93), so the control is the wizard's `worldsToOptions` picker with no hex/path entry. Cargo is **not** part of this form (see next). Answer type: `[]world.MovementDecision`.
- *Engine wrinkle to confirm at Plan (does not change the modal):* Issue ignores `Destination.RegionHex` and derives the hex from the world (movement.go:91); Revise trusts the passed `RegionHex` (movement.go:124). The modal populates both `WorldID` and the chosen world's hex, satisfying both paths — flagged so Plan doesn't trip on the inconsistency.

**`SelectTransportCargo` — capped multi-select.** `huh.NewMultiSelect[*domain.Asset]` over `eligibleCargo`, `Validate` enforcing `profile.MaxCargo` — structurally identical to the wizard's "choose up to N" groups. Fires once **per transport** that issued an Issue move (the orchestrator.go:668 loop), i.e. as sequential follow-on overlays *after* the movement form closes, not as part of it. Answer type: `[]*domain.Asset`.

**`AwaitCheckpoint(cycle_summary)` — bare ack.** In whole-cycle mode this is the **only** checkpoint that pauses (Decision 12 / phase_collector.go:28); `goal_locked` / `bookkeeping` / `movement` / `action_result` auto-nil at the adapter and never reach the TUI. So the real `AwaitCheckpoint` overlay is exactly the end-of-cycle wrap: a centered "Cycle N complete — Enter to return to setup." **No summary content** — the event stream is the durable record and a post-cycle review view is deferred (Out of Scope). Answer: the checkpoint ack default.

**Faction-detail pane — phase-reactive.** Effort 2 makes the right pane reflect the active phase atop Effort 1's identity + stats card:
- *stat_raise* — emphasize the three stat ratings (the raise targets).
- *movement* — list the faction's movable assets with current location / order.
- *bookkeeping, goal-lock* — the base identity card (Coin + HP already present); no special enrichment.
- *(action — Effort 3.)*

The pane renders from the snapshot-on-event copy, never the live pointer (*Execution view → Data discipline*). The "active phase" signal is derived from the observer event stream (phase-boundary events such as `OnFactionTurnStarted`); the exact signal is Plan detail.

### Effort 3 modals — the action phase

The largest Effort. The action surface is **two collectors** — `action.Collector` (16 methods) + embedded `hooks.Collector` (2 methods) — plus `PhaseCollector.SelectAction`. Verified against `action/collector.go` and `hooks/collector.go` (2026-05-29).

**Structural decision — a reusable archetype toolkit, not a bespoke overlay per prompt.** The ~19 prompts collapse into **six interaction archetypes**, twelve of them pure `huh` primitives. Effort 3 builds the archetypes as generic `Overlay` types (Go generics carry the answer type), and each prompt instantiates one with its label, options, validation, and an injected **context-renderer** for the engine detail it surfaces. This is what keeps the per-action commits thin and additive.

| Archetype | `Overlay` type | Prompts | Answer | Key engine context shown |
|---|---|---|---|---|
| Single-select (+optional skip) | `SelectOverlay[T]` | `SelectAction` (skip→nil), `SelectDefender`, `SelectAsset`, `SelectFactionTestTarget`, `SelectSeizeTarget`, `SelectChangeHomeworldTarget` | `T` / `nil` / `string` | candidate identity + the stat/HP/owner relevant to the choice |
| Multi-select (sometimes capped) | `MultiSelectOverlay[T]` | `SelectAttackers`, `SelectBaseAttackers`, `SelectAbilityAssets`, `SelectModifiers` | `[]T` | per-candidate attack/ability profile; offer source + budget |
| Yes/No confirm | `ConfirmOverlay` | `ConfirmRedirectToBase`, `ConfirmRivalFreeAttack`, `ConfirmAbilityApplied`, `ConfirmReroll` | `bool` | the stakes: damage, the two rolls, the ability, the reroll target |
| Two-step parent→child | `twostep` composite | `SelectBuyOrder` (world→def), `SelectRefitOrder` (asset→replacement) | `BuyOrder` / `RefitOrder` | child options annotated cost vs. the faction's Coin |
| Per-item numeric | `numeric` composite | `SelectRepairOrders` (per damaged asset → heal count), `SelectBribeTarget` (base + amount) | `[]RepairOrder` / `(*Base, int)` | HP gap and repair/bribe cost against Coin budget |
| Branching wizard | `expand` composite | `SelectExpandInfluenceOrder` (mode→world/base→submode→HP) | `ExpandInfluenceOrder` | new-vs-reinforce, target, heal-vs-max, HP amount + cost |

**Per-prompt context principle.** Each modal surfaces exactly what the GM needs to make *that* call — combat prompts show attacker/defender stats and damage profile; buy/refit show option cost against current Coin; target selectors show the target's owner/HP. The context-renderer is the per-prompt knob on an otherwise generic archetype; the exact fields/layout per prompt are Effort-3 Plan detail, but the **contract** (prompt set, archetype, answer type, decision-relevant context) is locked per action below.

**Per-action contract (independence inventory, verified 2026-05-29).** All 12 registered actions (register.go), mapped to the collector prompts they invoke — direct, plus transitive through the `ability` and `hooks/dispatch` packages. This is the artifact that proves each per-action commit is **self-contained and non-blocking**: every prompt below is invoked by *exactly one* action **except** `SelectAsset` (Buy + Sell) and the cross-cutting hooks — which is why those are hoisted into the foundation commit (see *Commit shape*).

| Action | Prompts → archetype | Answer | Decision-relevant context |
|---|---|---|---|
| **Sell Asset** | `SelectAsset` → Select *(shared → foundation)* | `*Asset` | owned assets + identity/stats |
| **Buy Asset** | `SelectBuyOrder` → two-step; `SelectAsset` → Select *(shared)* | `BuyOrder`; `*Asset` | world → purchasable defs, cost vs. Coin |
| **Repair Asset** | `SelectRepairOrders` → numeric | `[]RepairOrder` | damaged assets, HP gap, heal cost vs. Coin |
| **Refit Asset** | `SelectRefitOrder` → two-step | `RefitOrder` | refittable asset → valid replacements + cost |
| **Attack** | `SelectAttackers` → MultiSelect; `SelectDefender` → Select; `ConfirmRedirectToBase` → Confirm | `[]*Asset`; `*Asset`; `bool` | attacker/defender stats, HP, damage profile; base + damage on redirect |
| **Expand Influence** | `SelectExpandInfluenceOrder` → wizard; `ConfirmRivalFreeAttack` → Confirm; `SelectBaseAttackers` → MultiSelect | `ExpandInfluenceOrder`; `bool`; `[]*Asset` | new-vs-reinforce, world/base, heal-vs-max, HP+cost; the two rolls; eligible attackers vs. rival base |
| **Bribe** | `SelectBribeTarget` → numeric | `(*Base, int)` | bribable bases + Coin budget |
| **Use Asset Ability** | `SelectAbilityAssets` → MultiSelect; `ConfirmAbilityApplied` → Confirm; `SelectFactionTestTarget` → Select | `[]*Asset`; `bool`; `*Faction` | ability-capable assets + effect; the ability def; candidate faction + effect-relevant stat |
| **Seize Planet** | `SelectSeizeTarget` → Select | `string` | seizable world/base ids + owner |
| **Change Homeworld** | `SelectChangeHomeworldTarget` → Select | `string` | candidate worlds |
| **Repair Faction** | *(none — zero-prompt)* | — | register `Name()` only |
| **Abandon Goal** | *(none — zero-prompt)* | — | register `Name()` only |

Hooks (`SelectModifiers` → MultiSelect, `ConfirmReroll` → Confirm) fire transitively during any rolling action's resolution (hooks/dispatch/roll.go) — cross-cutting, hence foundation-owned.

**`SelectAction` + disable-unimplemented (Decision 7).** `SelectAction` receives `[]action.Action`, already `Validate`-filtered by the engine (action.go:45). Because the collector is wired *into* each action at construction (`ActionFactory(collector)`, action.go:26/49) and prompts fire inside `Action.Inputs()`, an action with stubbed collector methods still validates and would appear selectable — so the modal **disables** any action whose `Name()` is not in the TUI's **implemented-action set**. `Action` exposes only `Name()`, so that set is keyed by name; it grows by one entry per per-action commit. Disabled options render but can't be selected (with a "not yet available" hint). The exact registry shape stays Effort-3 Plan detail per Decision 7.

**Hooks prompts — real in the foundation commit; Esc degrades to no-op.** `SelectModifiers` / `ConfirmReroll` are cross-cutting (they fire during *any* action's roll resolution), so leaving them stubbed would mean the first per-action commit hits a no-op mid-resolution. Both modals are therefore built in the **action-flow foundation commit**, ahead of any per-action work: `SelectModifiers` → `MultiSelectOverlay[hooks.ModifierOffer]` over the offer list; `ConfirmReroll` → `ConfirmOverlay`. **Critically, these two methods return no `error`** (collector.go:8-9), so they structurally cannot carry `ErrTurnCanceled` — `Esc` inside a hook modal resolves to the **legal no-op** (`SelectModifiers` → none applied, `ConfirmReroll` → false/skip), *not* a turn-cancel. This is the one place the `Esc`-cancel path can't reach.

**Commit shape (refines Decision 7).** The **foundation commit** ships everything shared, so the dependency graph is a **star** (each action → foundation only; no action → action edge):
- the archetype toolkit (all six archetypes, including the two-step / numeric / wizard composites);
- the real `SelectAction` modal + the disable-unimplemented set;
- both hook modals (`SelectModifiers`, `ConfirmReroll`) — cross-cutting per the contract above;
- **`SelectAsset`** — the *only* prompt shared between actions (Buy + Sell), hoisted here rather than landing "with whichever action first" (the arc's phrasing), which would have been the lone mesh edge;
- `Esc`-cancel routing for the error-returning `action.Collector` methods (`ErrTurnCanceled` on `Reply` → orchestrator classifies; see *Esc-cancel path*);
- the two **zero-prompt actions** (Repair Faction, Abandon Goal): no modal — just their `Name()` added to the implemented set so `SelectAction` enables them.

Then **one commit per remaining action** instantiates archetypes for *only that action's unique prompts*, adds its `Name()` to the implemented set, and adds any new `AskKind` values + payload types. Because no per-action commit shares a prompt with another, each is self-contained; an unimplemented action stays disabled in `SelectAction` and never blocks the others. (Effort 3 thus ships **all 12** registered actions — the inventory surfaced Sell Asset, Repair Faction, and Abandon Goal, which the arc-plan's generic "`Select*` prompts" blurb had not called out.)

### Whole-cycle only — no cadence control (Decision 12)

Turn runs **whole-cycle only**. The setup cadence choice, the mid-cycle toggle keybinding, and the cadence indicator are all dropped — they added a control surface (a flippable mode competing for screen real estate) for marginal value, and per-faction stepping is deferred.

- The adapter's `perFactionCadence` (`atomic.Bool`) and `SetPerFactionCadence` (adapter.go:41) **remain** (Foundation built them) but Turn never wires a control; the flag sits at its zero value (`false` = whole-cycle).
- **What still pauses.** Cadence only ever gated `AwaitCheckpoint`. In whole-cycle mode `phaseCollector.AwaitCheckpoint` (phase_collector.go:28) returns `nil` immediately for `goal_locked` / `bookkeeping` / `movement` / `action_result`, **except** `CheckpointCycleSummary`, which always pauses (kept — the GM acks the cycle wrap before returning to setup). The four `Select*` asks fire **regardless of cadence**, so the overlay dispatch and seam are unchanged.
- Net: the cycle streams straight through, stopping only for genuine decisions (`Select*`) and the final cycle summary.

Per-faction cadence is **Out of Scope** (revisitable later without disturbing this design — the flag and gating are already in place).

### Esc-cancel path

Foundation laid the sentinel but **deferred the routing to this initiative** (channels.go:101 comment → "the action-selection initiative" = Effort 3).

- `ErrTurnCanceled` exists. Cancelling an ask = sending it on the ask's `Reply` channel; the collector returns it as an error.
- The orchestrator classifies the returned error: `IsRecoverable` → logs a warning, calls `observer.OnError`, and **continues** (returns `nil`/`false`) rather than aborting the whole run (e.g. orchestrator.go:174–183). Whether `ErrTurnCanceled` should be `Recoverable` (skip current faction, keep cycling) or unwind differently is an Effort 3 design point — confirm classification against the error conventions then.
- **Effort 1** establishes only that a reply *can* be an error and the loop survives it; full Esc-cancel UX is Effort 3.

### Help overlay wiring

Same idiom as Manage (`manage.go:186`): the root pulls the active sub-model's keymap via the `Helper` interface, and `turn.Model.Help()` switches on `view` to return `setup.Help()` or `execution.Help()`, which the root composes with the global bindings. (This is the existing footer plumbing the contextual bottom bar reworks — *Contextual bottom bar*; help is the fall-through content when no error outranks it.)

The one Turn-specific wrinkle is the **overlay**: while an overlay floats over the execution panes, keys route to it, so `execution.Help()` must return the overlay's keymap (e.g. huh form nav — tab/space/enter, plus the in-modal decline/cancel) rather than the panes' navigation. When no overlay is up, it returns the execution view's own bindings (event-stream scroll/follow). Like Manage's wizard (`CapturesInput`, `manage.go:182`), an overlay with free-text or its own field navigation must signal that it captures input so the root doesn't steal `q`/`tab`/`?`.

### Logging wiring

The Turn workflow must run with a logger from the `internal/logging` package that writes to the campaign's logs directory.

Confirmed facts (verified 2026-05-29):
- `logging.New(logsDir string, debug bool) (*slog.Logger, error)` is the constructor — returns a standard `*slog.Logger`.
- `campaigns.Paths.LogsDir` already resolves to `state/logs` under the campaign root (e.g. `campaigns/test-camp/state/logs`, currently empty).
- The TUI root Model already receives `*campaigns.Paths` (Manage threads it for CRUD), so `logging.New(paths.LogsDir, debug)` needs no new plumbing to reach the directory.

**Resolved (Decision 9): app-wide, constructed in the `faction` command.**
- Today `cmd/gm-toolkit/faction.go:28` builds a *stderr* `slog` handler — not `internal/logging`, not the campaign dir. This is replaced.
- Add a `--debug` bool flag to the `faction` command. Construct `logging.New(paths.LogsDir, debug)` **inside `factionRun` after `paths` resolves** (`faction.go:51`) — it cannot be built at line 28 because `paths` aren't known until the active campaign is resolved.
- Thread the resulting `*slog.Logger` into `engine.NewWithRulebook(rb, log)` and `tui.Run(..., log)` exactly as today (both already accept it). The engine cycle and the TUI then write to one campaign log file.
- Pre-campaign errors (registry load, active-campaign resolution — before `paths` exist) stay on stderr / returned errors.
- App-wide, so it lands early in Effort 1 and Foundation/Manage inherit it.

## State Storage

**No new on-disk schema.** Turn drives the engine, and the engine already owns all cycle persistence: `RunCycle` calls `state.Save(paths.StatePath, factionState)` at the phase-gate checkpoints and at cycle close (orchestrator.go:87, 109, 508), and `applyAndRecord` appends `EventRecord`s to the history file as mutations land (orchestrator.go:537-538). The TUI reads nothing new from disk — every value it renders arrives over the observer/ask channels as the cycle runs (the snapshot-on-event copies, *Execution view → Data discipline*). Manage's own CRUD persistence (`state.CreateFaction` / `DeleteFaction`) is separate and unchanged. The only new persistent artifact is the **campaign log file** (`logging.New(paths.LogsDir, …)`, Decision 9 / *Logging wiring*), which is operator-facing output, not game state.

## User-Facing Impact

After F-005.3 lands, the GM can drive a full engine cycle interactively from the TUI — the first time the engine runs anywhere but tests and the dry-run smoke path. The cumulative experience across the three Efforts:

- **Launch Turn** from the mode bar → a read-only roster of the campaign's factions (alphabetical), with an empty-state pointing back to Manage if there are none.
- **Press Start** → the adapter rolls turn order, increments the cycle, and begins `RunCycle` on a goroutine. The view switches to the three-pane execution layout: turn-order rail (left), live event stream (center), processing-faction detail (right).
- **Watch the cycle stream** event-by-event as each faction is walked through its phases; the rail's "current" marker advances, the detail pane reflects the active phase, and the event log auto-follows (scroll up to lock the tail, `End`/`G` to re-engage).
- **Answer decisions as they arise** through floating modals — stat raise, movement + transport cargo (Effort 2); action selection and the full action surface — attack, buy, refit, expand influence, repair, abilities, bribe, seize, change homeworld (Effort 3). Errors surface in the contextual bottom bar (recoverable vs. fatal) and are logged permanently in the event stream.
- **Ack the cycle wrap** at the forced `cycle_summary` checkpoint → return to setup, state and history already persisted by the engine.
- **Esc** cancels the current ask (Effort 3); the orchestrator classifies the cancel and the cycle either skips the faction and continues or unwinds.

**What still does not work** after F-005.3:
- **Per-faction cadence / mid-cycle stepping** — Turn runs whole-cycle only (Decision 12).
- **Pause / resume mid-cycle, skip-a-faction-without-aborting, reorder-on-the-fly** — beyond `Esc`-aborts-current-turn (arc deferral).
- **Reviewing a past cycle** — no post-mortem / "review last cycle" view; the event stream is live-only and not replayable.
- **Unimplemented actions** (if any are left for a later pass) appear **disabled** in the action picker rather than selectable.

## Out of Scope

<!-- Seeded from arc-plan Initiative 3 "Out of Scope" + deferrals. Refine. -->

- **Per-faction cadence (mid-cycle stepping).** Turn runs whole-cycle only (Decision 12). The adapter's cadence flag and gating stay in place, so this can return later without disturbing the design.
- **Pause / resume mid-cycle, skip-faction-without-aborting, reorder-on-the-fly.** The arc stops at `Esc`-aborts-current-turn. (Arc-plan deferral.)
- **Turn post-mortem / "review last cycle" view.** Sub-view growth deferred.
- **Theming, cross-launch persistence.** Arc-wide deferrals.
<!-- TODO(fill): add Turn-specific non-goals discovered during the session. -->

## Reference Exemplars

- **Arc-Discovery — Turn section** (`tui-rebuild-arc-discovery.md`, "View Structure → Turn") — setup/execution split, event-stream + state-pane framing.
- **Arc-Discovery — Update discipline** (same doc) — Update may emit `tea.Cmd`s for side-effects but may not block; collector replies flow through the adapter, not inline.
- **Arc-Plan — Initiative 3 entry** (`tui-rebuild-arc-plan.md`) — the In Scope / Out of Scope boundary this Discovery refines, plus the two routed open questions.
- **Manage-Discovery** (`tui-manage-discovery.md`) — the router sub-model, sub-model-per-view, completion-message transition seam, help-overlay wiring, and empty-state grammar Turn inherits.
- **`internal/faction/engine/orchestrator.go`** — the authority on how a faction turn / cycle actually executes; grounds the collector-prompt and observer-event inventory.
- **Foundation adapter sub-package** (`internal/faction/tui/adapter/`) — `PhaseCollector` / `TurnObserver` implementations, reply channels, observer pump, and cancel-sentinel plumbing that Turn consumes. (The cadence flag exists but Turn leaves it unused — Decision 12.)
- **Error conventions** — `Recoverable` / `Fatal` branch the `Esc`-cancel path leans on.

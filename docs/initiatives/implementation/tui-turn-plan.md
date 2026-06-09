# TUI Turn — Implementation Plan (Overview)

`F-005.3`, Initiative 3 of the TUI Rebuild arc. Top-level plan for the Turn workflow — the engine-crossing TUI view that drives a whole faction cycle to completion. Ships in **three Efforts on one feature branch**; this file holds the cross-effort context and routes the per-Effort detail to three separate Plan sessions.

- Discovery: [`tui-turn-discovery.md`](../discovery/tui-turn-discovery.md)
- Arc: [`tui-rebuild-arc-plan.md`](../arcs/tui-rebuild/tui-rebuild-arc-plan.md) · [`tui-rebuild-arc-discovery.md`](../arcs/tui-rebuild/tui-rebuild-arc-discovery.md)
- Effort 1 — Turn-flow scaffold: [`tui-turn-effort-1-plan.md`](./tui-turn-effort-1-plan.md)
- Effort 2 — Phases except action: [`tui-turn-effort-2-plan.md`](./tui-turn-effort-2-plan.md)
- Effort 3 — The action phase: [`tui-turn-effort-3-plan.md`](./tui-turn-effort-3-plan.md)

## Context / Goal

Turn is the mirror of Manage: where Manage bypasses the engine to do plain CRUD on `FactionState`, Turn hands control **to** the engine and walks a full `RunCycle` through the UI. The engine runs synchronously and blocks on collector prompts; Foundation's adapter already marshals that across channels (observer events onto a buffered pump, blocking asks onto an unbuffered `askCh` with a per-ask `Reply` channel). Turn consumes that contract.

Discovery ratified all fourteen design decisions and left **no open questions at the Discovery level** — the design is complete across all three Efforts. This plan's job is therefore not to re-derive design but to (1) ratify the few *organizational* decisions Plan adds, (2) lock the one cross-effort contract every Effort builds on (the `Overlay` seam), and (3) route discovery's scattered "Plan detail" deferrals to the Effort whose Plan session resolves each.

The Efforts are layered so that **Effort 1 is structural and Efforts 2–3 are purely additive**: Effort 1 proves the whole loop against the real engine with auto-answering placeholder overlays; Efforts 2 and 3 swap placeholders for real modals without restructuring the consumption loop, the event stream, or the dispatch switch.

## Decisions Ratified in Planning

1. **Single feature branch `feature/tui-turn`; all three Efforts, one merge.** wip: commits per layer during execution; the pre-merge checklist runs **once** at branch level after Effort 3 lands; squash to one or more conventional commits at merge (Robert's call at merge time). Matches the multi-Effort precedent (`logging-and-errors`, `asset-movement-redesign` / `project_movement_branch_merge`) and the arc model where each *initiative* is one branch. Rationale: Turn is not independently shippable until Efforts 2–3 fill the modals — Effort 1 alone runs only an auto-defaulting cycle, so per-Effort merges would ship half-built states.
2. **Effort order is dependency-fixed: 1 → 2 → 3.** Effort 1 lays every seam (router, three-pane layout, ask-pump, dispatch switch, event-stream formatter, app-wide chrome). Effort 2 swaps four non-action overlays. Effort 3 swaps `SelectAction` and adds the action surface. No Effort depends on a *sibling*; 2 and 3 each depend only on Effort 1's seam. (Effort 3's internal commits form a star on its own foundation commit — Effort-3 Plan detail.)
3. **The `Overlay` seam is ratified at overview level** (see *Shared Context → The Overlay seam*). It is the single cross-effort contract whose stability is the precondition for Decision 2's "purely additive" promise, so it is pinned here rather than in any one Effort's plan — mirroring how `asset-movement-redesign-plan.md` pinned the `Location` struct in its overview.
4. **App-wide work lands in Effort 1.** Logging (Discovery Decision 9) and the contextual bottom bar (Discovery Decision 8) are root chrome that benefits the whole app, not just Turn. Both land early in Effort 1; Foundation and Manage inherit them. They are not deferred to whichever Effort "needs" them.
5. **Decisions-log scope.** Discovery and Plan decisions live in their own artifacts. The decisions log (`docs/dev_journals/faction-manager/decisions-log.md`) records only revisions or new decisions made **during execution** (per `feedback_decisions_log_scope`). This overview's decisions are not back-filled there.

## Open Questions — To Ratify at Implementation Time

Discovery resolved every Discovery-level question but tagged numerous items "Plan detail." Collected and routed here; each Effort's Plan session resolves its rows.

| # | Question | Effort / when |
|---|----------|---------------|
| 1 | Exact `turn.Model` / `execution.Model` / `setup.Model` struct shapes and field sets. | Effort 1 Plan |
| 2 | The **cycle-started message** type and which channel carries it (surfaces a copy of `FactionOrder` + cycle number before the engine loop proceeds). | Effort 1 Plan |
| 3 | `AskPump() tea.Cmd` shape — the symmetric ask-pump reading `askCh` (no equivalent exists yet; `run.go` has only `ObserverPump`). | Effort 1 Plan |
| 4 | `Adapter.Stop()` safety reshape — current `Stop()` closes `eventCh` unconditionally; must not close while a cycle is live (close-during-emit panic). | Effort 1 Plan |
| 5 | `Mutation → string` formatter: the exact "key" mutation set, per-type phrasing, and which bookkeeping mutations fold/omit (~30 types in `mutation.go`). | Effort 1 Plan |
| 6 | `StatusLiner` exact signature (sketch: `StatusLine() (text string, severity Severity)`) and the `Severity` type. | Effort 1 Plan |
| 7 | Placeholder overlay's per-kind default answer zero-values (legal "no decision" reply per `AskKind`). | Effort 1 Plan |
| 8 | `SelectMovementDecisions` conditional-field `huh` form (per-row destination shown only on Issue/Revise) — first conditional-field form in the codebase; confirm the `WithHideFunc`-style gate is viable. | Effort 2 Plan |
| 9 | `movement.go` Issue-vs-Revise `RegionHex` inconsistency (Issue derives hex from world; Revise trusts the passed hex) — confirm the modal populates both paths correctly. Does not change the modal. | Effort 2 Plan |
| 10 | The "active phase" signal the phase-reactive detail pane derives from the observer event stream. | Effort 2 Plan |
| 11 | Exact per-prompt fields/layout per archetype (the *contract* — prompt set, archetype, answer, context — is locked in Discovery; only layout is open). | Effort 3 Plan |
| 12 | Implemented-action registry shape (keyed by `Action.Name()`; grows one entry per per-action commit). | Effort 3 Plan |
| 13 | Safe-stub classification for not-yet-built action-unique methods (must be recoverable + surfaced — never `panic`, never unclassified-fatal). | Effort 3 Plan |
| 14 | `ErrTurnCanceled` classification (`Recoverable` → skip faction & continue, vs. unwind) confirmed against the error conventions; and the cancel message shape (an `OverlayDoneMsg` carrying the error vs. a sibling msg). | Effort 3 Plan |

## Effort Summary

Three Efforts, one branch. Each Effort gets its own Plan session (Opus) producing a `-effort-N-plan.md` with the Phase/Commit/Task breakdown; each commit in those files is one execution session (Sonnet default).

| # | Effort | Scope | File |
|---|--------|-------|------|
| 1 | Turn-flow scaffold | Router sub-model + setup/execution views; three-pane layout (Region promotion); **ask-pump**; dispatch switch over all 5 `AskKind` → generic placeholder overlay (auto-answers defaults); event-stream renderer over all 14 events + `Mutation → string` formatter; app-wide logging + contextual bottom bar + `StatusLiner`; whole-cycle-only; empty-state; `Stop()` safety. Drives the **real** engine end-to-end with every decision defaulted. | [`tui-turn-effort-1-plan.md`](./tui-turn-effort-1-plan.md) |
| 2 | Phases except action | Swap placeholders for real `huh.Form` overlays on the four non-action kinds (`AwaitCheckpoint` ack, `SelectStatRaise`, `SelectMovementDecisions`, `SelectTransportCargo`); phase-reactive faction-detail pane. Additive — swaps overlay constructors per `Kind`, no plumbing change. | [`tui-turn-effort-2-plan.md`](./tui-turn-effort-2-plan.md) |
| 3 | The action phase | Reusable six-archetype overlay toolkit; foundation commit ships everything shared (real `SelectAction` + disable-unimplemented set, both hook modals, `SelectAsset`, `Esc`-cancel routing, the two zero-prompt actions); then one self-contained commit per remaining action. Ships **all 12** registered actions. Additive at adapter (new `AskKind` + payloads) and TUI (new switch cases). | [`tui-turn-effort-3-plan.md`](./tui-turn-effort-3-plan.md) *(later session)* |

## Shared Context

Concerns that span the Efforts, pinned here so each Effort file references rather than re-derives them.

### Branch and merge strategy

- One feature branch: `feature/tui-turn`.
- wip: commits during execution (one per layer, per CLAUDE.md item 10 / `feedback_git_commits`).
- Pre-merge checklist runs **once** at branch level after Effort 3 (code review, dev journal update, planned-work update) — per `feedback_branch_initiative_layering`.
- Squash to conventional commits at merge — likely one `feat(tui/turn): …` per Effort, or a single combined commit; Robert's call at merge time.
- Post-merge: assess patch/minor/major bump (CLAUDE.md item 9). Turn is a substantial user-facing capability — likely a minor bump; confirm at merge.

### Model discipline

| Session | Model | Why |
|---------|-------|-----|
| This overview Plan session | Opus | Plan mode |
| Effort 1 / 2 / 3 Plan sessions | Opus | Plan mode (always) |
| Effort 1 execution | Sonnet default; **suggest Opus** for the dispatch-switch + ask-pump + formatter commits | The channel round-trip seam and the contextual-bar rework carry real design weight |
| Effort 2 execution | Sonnet | `huh.Form` overlays follow the existing wizard idiom; the conditional-field form (OQ 8) may warrant Opus for its one commit |
| Effort 3 execution | Sonnet default; **suggest Opus** for the archetype-toolkit foundation commit | The generic archetype types + two-step/numeric/wizard composites are the design-heavy piece; per-action commits are mechanical instantiation |
| End-of-phase docs & pre-merge checklist | Sonnet | Always — doc/checklist work |

Prompt to switch model at each transition (`/model`), including the mid-session shift into the pre-merge checklist.

### The Overlay seam (the contract Efforts 2 & 3 build on)

This is the load-bearing artifact — Decision 2's "purely additive" promise holds only because this contract is fixed before Effort 1. Ratified here; Efforts swap/add overlay implementations behind it without touching the dispatch.

**The `Overlay` interface.** An overlay is a self-contained `tea.Model`-shaped component that renders one prompt and, on completion, emits its answer as a message via `tea.Cmd` — exactly as `deleteconfirm` emits `DeletedMsg` / `CancelMsg`. It is held by the execution sub-model as an **interface variable** (one field, many concrete overlay types via the factory switch), so unlike Manage's concrete-typed sub-models its `Update` returns the interface — avoiding a per-frame type-assertion:

```go
// internal/faction/tui/views/turn/overlay (exact package path is Effort-1 Plan detail)
type Overlay interface {
    Init() tea.Cmd
    Update(tea.Msg) (Overlay, tea.Cmd)
    View() string
    Help() help.KeyMap
}

// Emitted by an overlay on completion; carries the typed answer the collector expects.
type OverlayDoneMsg struct {
    Answer any
}
```

- Overlays **never import the adapter channels** and never see `Reply` — they only render a prompt and emit `OverlayDoneMsg{Answer}`. The execution sub-model alone bridges overlay ↔ channel.
- **Cancel** is a variant of completion. Its exact shape (an `OverlayDoneMsg` carrying `ErrTurnCanceled`, or a sibling message) is **deferred to Effort 3** (OQ 14) — Effort 1 needs only the happy-path `Answer`.

**Ownership & dispatch (execution sub-model).** Execution holds `overlay Overlay` (nil when no prompt is up) and `pendingReply chan<- any` (the `Reply` from the active `CollectorAskMsg`). The flow that **must not churn** across Efforts:

1. On `CollectorAskMsg`: store `msg.Reply` as `pendingReply`; build the overlay via a factory `switch msg.Kind` enumerating **all current `AskKind` values**; mount it (`overlay.Init()`).
2. While `overlay != nil`: route key messages to `overlay.Update`; the overlay reports it `CapturesInput()` so the root doesn't steal `q`/`tab`/`?`.
3. On `OverlayDoneMsg`: send `msg.Answer` on `pendingReply` (the buffered cap-1 reply channel), then clear both `overlay` and `pendingReply`.

Swapping a branch from placeholder → real overlay (Effort 2), or adding a branch for a new `AskKind` (Effort 3), is **filling-in**, never restructuring. The five current kinds are already declared in `adapter/channels.go` (`AskAwaitCheckpoint`, `AskSelectAction`, `AskSelectStatRaise`, `AskSelectMovementDecisions`, `AskSelectTransportCargo`); Effort 3 appends action kinds.

**Effort 1 placeholder.** All branches point at one generic placeholder overlay that renders which ask fired for which faction and, on a keypress, emits that kind's default answer (the legal "no decision" reply). This is the first code to exercise the real `askCh` round-trip.

### Turn package layout

Mirrors Manage one level down, under `internal/faction/tui/views/turn/` (sketch; exact file/field shapes are Effort-1 Plan detail, OQ 1):

```
turn/
  turn.go        router Model: view enum {viewSetup, viewExecution}, dispatch, Help, CapturesInput
  messages.go    Turn-local tea.Msg types (cycle-started, OverlayDoneMsg, surfaced CollectorAskMsg)
  setup/         setup.Model — read-only roster + Start (dispatches Adapter.Run)
  execution/     execution.Model — three-pane layout + overlay ownership + StatusLiner
  overlay/       the Overlay interface + Effort-1 generic placeholder; Effort-3 archetype toolkit
```

The router owns the `adapter.Adapter` (built when Setup dispatches Start) and threads its commands — `ObserverPump`, the new `AskPump`, `EngineDoneMsg` — into `execution`. Setup stays engine-ignorant.

### App-wide chrome — lands in Effort 1

Two pieces of root chrome land early in Effort 1 (Decision 4) because every mode renders through them:

- **Logging** (Discovery Decision 9): add a `--debug` flag to `cmd/gm-toolkit/faction.go`; replace the stderr `slog` handler with `logging.New(paths.LogsDir, debug)` constructed **after `paths` resolves** inside `factionRun`; thread the `*slog.Logger` into `engine.NewWithRulebook` and `tui.Run` (both already accept it). Pre-campaign errors stay on stderr.
- **Contextual bottom bar** (Discovery Decision 8): rework `view.go`'s `showHelp`-gated footer into an always-present, priority-resolved bar — a live error (recoverable vs. fatal, color-coded) outranks help; otherwise help shows, and `?` toggles help **compact ↔ full** rather than toggling the bar's existence. Add a `StatusLiner` interface (OQ 6) parallel to the existing `Helper` check; the active sub-model contributes error content. Modes that don't implement it fall through to help.

> Re-grounding note for the Effort-1 Plan session: the recent theming refactor (commits `e3a4538`, `eb772d4`) shifted line numbers, so Discovery's `view.go:65/67`, `faction.go:28/51` citations must be re-verified against current source before coding. The *structures* (the `showHelp` footer block, the `Helper` check, the stderr handler, the `factionRun` paths-resolution point) are all confirmed present as of this writing.

### Region layout promotion

The three-pane split (`Left` ¼ · `Center` ½ · `Right` ¼ with vertical rules) is Manage's `Region`/`compose`/`regionWidths` helper, **promoted out of `package manage` into a shared `tui` layout package** that both Manage and Turn import. This is a small extraction that touches Manage; it lands in Effort 1. Whether it becomes `internal/faction/tui/layout` or similar is Effort-1 Plan detail.

## Out of Scope

Inherited from Discovery; restated for emphasis because Efforts may be tempted to creep into them.

- **Per-faction cadence / mid-cycle stepping.** Turn runs whole-cycle only (Discovery Decision 12). The adapter's `perFactionCadence` flag and its gating stay in place (unused at zero value), so this can return later without disturbing the design.
- **Pause / resume mid-cycle, skip-faction-without-aborting, reorder-on-the-fly.** The initiative stops at `Esc`-aborts-current-turn (Effort 3). (Arc deferral.)
- **Turn post-mortem / "review last cycle" view.** The event stream is live-only, not replayable; no post-cycle summary content beyond the bare `cycle_summary` ack. (Sub-view growth deferred.)
- **Full pause/resume UX.** Effort 1 adds only the minimal "don't re-roll a live turn" guard in `Adapter.Run()` (resume an `InProgress` turn instead of re-rolling); the surrounding UX is not built.
- **Narrative-renderer integration.** The event stream is the turn's own terse formatter, not the digest-driven prose renderer (reserved for the deferred review view).
- **New on-disk schema.** The engine already owns all cycle persistence; the only new persistent artifact is the campaign log file. (Discovery *State Storage*.)
- **Theming, cross-launch persistence.** Arc-wide deferrals.

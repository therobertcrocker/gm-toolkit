# TUI Rebuild — Arc Plan

`F-005`. Companion to [`tui-rebuild-arc-discovery.md`](./tui-rebuild-arc-discovery.md).

## Context / Goal

This Arc-Plan slices the TUI rebuild into shippable initiatives, fixes their order and boundaries, and routes the arc-discovery's open questions to the initiative that owns each one. The cross-cutting architecture (engine-consumption boundary, Bubbletea / lipgloss / huh stack, mode-bar router, adapter sub-package, Update discipline) is already ratified in arc-discovery and is not re-litigated here.

The arc ships in **three layered initiatives**: a foundation that proves the adapter wiring against the real engine, a Manage initiative that delivers faction CRUD, and a Turn initiative that delivers interactive turn execution. Each initiative runs the standard per-Type Feature Flow (`docs/process/session-modes.md` section 4): Discovery → Plan → Execution → Pre-Merge Checklist. Per-initiative Discovery is *light* — it does not re-open arc-level decisions, only the initiative-specific scope and the open questions routed here.

## Decisions Ratified in Planning

1. **Three-initiative slicing — Foundation, then Manage, then Turn.** Strict sequence. Resolves the slicing question deferred to Arc-Plan. Foundation comes first because the adapter is the load-bearing piece of the architecture (it is where V1 failed) and it must be proven before either feature workflow is built on top. Manage precedes Turn because it ships a usable authoring tool sooner and because Turn's setup view consumes the faction roster Manage creates.
2. **Charm dependency reintroduction is absorbed into Foundation's first commit.** No separate chore initiative. `go.mod` adds `bubbletea`, `lipgloss`, `huh` alongside the skeleton commit so the dep add and the first import land together.
3. **Foundation includes a minimum end-to-end adapter exercise.** Not scaffold-only. Foundation ships a dry-run cycle driven through real `PhaseCollector` and `TurnObserver` implementations against the real engine — validating the channel wiring, the goroutine-and-blocking-collector model, and the observer pump before Manage or Turn are built on top.
4. **Direct state editing is out-of-arc.** The third workflow named in arc-discovery's Problem statement — GM-side surgical edits outside the engine's mutation pipeline — does not ship in this arc. Captured as a separate backlog item with the expectation it spawns its own arc later.
5. **Empty-state UX is decided in Manage; Turn inherits the pattern.** Resolves how the empty-state open question is split. Manage's per-initiative Plan picks the grammar (hint copy, key prompt, visual treatment); Turn's setup view applies it to its "no factions" blocking state.

## Initiative List

The arc is a strict sequence — no parallelization. Each initiative depends on the prior one's branch having landed.

| # | Initiative | Type | Depends On | Status |
|---|------------|------|------------|--------|
| 1 | `tui-foundation` | feature | — | Backlog (promote to Up Next next) |
| 2 | `tui-manage` | feature | Initiative 1 merged | Backlog |
| 3 | `tui-turn` | feature | Initiative 2 merged | Backlog |

Architecturally, Manage and Turn touch disjoint regions of the codebase (Manage is engine-bypassing CRUD; Turn is engine-mediated execution). They could parallelize if multiple contributors ran them — single-developer cadence keeps them serial. The DAG is captured above to make the parallelism *possible* if circumstances change, without committing to it.

---

## Initiative 1 — `tui-foundation`

**Goal.** Stand up the TUI package skeleton, the mode-bar router, the adapter sub-package (collectors + observer + channels), and a dry-run end-to-end exercise of the adapter against the real engine. After this initiative, the TUI can be launched, mode-switched between empty mode shells, and observed driving a no-op cycle through the adapter — but neither feature workflow is yet implemented.

**In Scope.**
- Add `bubbletea`, `lipgloss`, `huh` to `go.mod`.
- Create `internal/faction/tui/` with `tui.go`, `model.go`, `update.go`, `view.go`, `views/`, `adapter/`, `styles.go`.
- Root `Model` with `mode` field; dispatch Update / View to sub-models.
- Mode-bar component with slots `Manage | Turn | Spatial | Quit`. Spatial greyed; Quit dispatches to confirm-and-exit. `Tab` / `Shift-Tab` mode cycling, `?` help-overlay stub, `q` quit binding.
- Empty mode shells for Manage and Turn — placeholder content only, just enough to render.
- `adapter/` sub-package: full `PhaseCollector` and `TurnObserver` implementations, reply channels, observer pump `tea.Cmd`, cadence flag (shared reference; mutex- or atomic-guarded), cancel-sentinel plumbing for `Esc`.
- Dry-run smoke exercise: a developer-only entry point (gated by a flag or a hidden keybinding) that runs `RunCycle` against a trivial in-memory faction state via the real adapter. Proves channels, goroutine launch, blocking collectors, observer pump, and cadence flag all work end-to-end.
- `gm-toolkit faction` opens the TUI per arc-discovery's package-layout decision.
- Update discipline rules from arc-discovery (Update MAY / MAY NOT) are encoded as comments at the top of `update.go` and enforced by the senior-engineer pre-merge review.

**Out of Scope (punted to later initiatives).**
- Faction list rendering, faction-detail, faction-create, faction-edit (Initiative 2).
- Turn setup view, Turn execution view, modal overlay infrastructure for real collector prompts, event stream rendering (Initiative 3).
- Help overlay content (stub only in Foundation; real per-view bindings populated as Manage and Turn add views).
- Empty-state UX grammar (decided in Manage).

**Per-initiative open questions** — none routed here. The arc-discovery's open questions all land in Manage or Turn.

**Estimated execution shape.** ~4–6 commits. Plan session breaks them down.

## Initiative 2 — `tui-manage`

**Goal.** Ship the Manage mode end-to-end: faction list, faction-detail, faction-create, faction-edit. After this initiative, the TUI replaces hand-edited TOML for the entire faction authoring loop.

**In Scope.**
- Manage sub-model with the internal view stack: `list → detail → edit`, plus `list → create` and `detail → delete-confirm` sub-views.
- Faction list view with detail-on-Enter, context strip on the right per arc-discovery's View Structure.
- Faction-detail view.
- Faction-create form (`huh`).
- Faction-edit form (`huh`).
- Delete-confirm sub-view.
- Draft-buffer-commits-on-save semantics for create and edit. Reads and writes faction state directly via the existing config layer — no engine boundary involved.
- Empty-state UX: Manage's "no factions" affordance (hint copy + create keybinding), establishing the pattern Turn will inherit.
- Help overlay populated with Manage's per-view bindings.

**Out of Scope.**
- Anything Turn-related.
- Tabbed faction-detail (Stats / Assets / Goals / History) — sub-view growth deferred per arc-discovery's Extensibility section.

**Per-initiative open questions to resolve in this initiative's Plan.**
- **Faction CRUD form scope** (from arc-discovery). Identity and stats only with goals and assets in separate sub-views, or inline at create time? Plan session picks.
- **Empty-state UX** (from arc-discovery). Plan session picks the grammar; Turn inherits it.

**Estimated execution shape.** ~5–8 commits depending on CRUD form scope. Plan session breaks them down.

## Initiative 3 — `tui-turn`

**Goal.** Ship the Turn mode end-to-end: setup view (cadence + roster), execution view (state pane + event stream), modal overlay infrastructure, cadence-flag mid-cycle flipping, `Esc`-aborts-current-turn cancel path. After this initiative, the GM can drive engine cycles interactively from the TUI.

**In Scope.**
- Turn sub-model with internal views: `setup → execution`.
- Setup view: cadence selection (per-faction vs whole-cycle), roster confirmation, `Start` dispatch.
- Execution view: left pane (processing faction's state, cadence flag, cycle number), right pane (scrolling event log fed by `TurnObserver` callbacks).
- Modal overlay component: one component, wraps `huh` for decisions, wraps single-key listener for `AwaitCheckpoint` acks. All `Select*` collector prompts route through it.
- Auto-follow event stream with scroll-locks-tail and `End` / `G` to re-engage.
- Cadence flag mid-cycle toggle via a keybinding from the execution view.
- `Esc`-cancel path: routes through adapter as turn-abort error, classified `Recoverable`, orchestrator unwinds current faction, control returns to execution view.
- Empty-state in Turn setup view applies the pattern Manage established.
- Help overlay populated with Turn's per-view bindings.

**Out of Scope.**
- Pause / resume mid-cycle, skip-faction-without-aborting, reorder-on-the-fly — deferred per arc-discovery's Deferrals section.
- Turn post-mortem / "review last cycle" view — sub-view growth deferred.
- Theming, cross-launch persistence — deferred.

**Per-initiative open questions to resolve in this initiative's Plan.**
- **What `Start` invokes from the Turn setup view** (from arc-discovery). `RunCycle` vs `RunFactionTurn` in a loop? Plan session picks based on which abort granularity matches the GM's mental model.
- **Error rendering shape** (from arc-discovery). Where does `OnError` surface? Plan session picks; Foundation has already laid the overlay and event-stream infrastructure such that none of the candidate options (event-stream-only / event-stream + chip / modal-for-fatal-inline-for-recoverable) is foreclosed.

**Estimated execution shape.** ~5–7 commits. Plan session breaks them down.

---

## Open Question Routing — Summary

| Open Question (from arc-discovery) | Routed To | Resolved In |
|---|---|---|
| Faction CRUD form scope | Initiative 2 — `tui-manage` | Per-initiative Plan |
| What `Start` invokes from Turn setup view | Initiative 3 — `tui-turn` | Per-initiative Plan |
| Error rendering shape | Initiative 3 — `tui-turn` | Per-initiative Plan |
| Empty-state UX | Initiative 2 — `tui-manage` (Turn inherits) | Per-initiative Plan |

## Cross-Arc Dependencies

All prerequisites are already merged on `main`. No engine-side or data-layer work blocks the arc.

- **Engine call surface** — `(*Engine).RunFactionTurn`, `(*Engine).RunCycle`, `PhaseCollector`, `TurnObserver`. In place (per arc-discovery's Architecture section).
- **Error conventions** — `Recoverable` / `Fatal` branch the `Esc`-cancel path leans on. Landed in `4ebebc4 feat(errors): error conventions, Recoverable/Fatal branch, MutationApplyError consumer`.
- **Structured logging** — landed in `cf379fc feat(logging): :art: added structured logging`. TUI inherits without action.
- **Asset movement redesign** — landed in `f09e4cc`. Movement-phase collector methods and observer events are in their final shape, no churn risk for Initiative 3.
- **Config layer** — faction TOML read / write path used by Manage already exists. Initiative 2 consumes it.

## Out-of-Arc Deferrals

Captured here so they are not silently dropped. Each should land in `planned-work.md` as a separate Backlog entry when this Arc-Plan is committed.

- **Direct state editing** (the stretch workflow from arc-discovery's Problem statement). GM-side surgical edits outside the engine's mutation pipeline. Expected to spawn its own arc. Not in this arc.
- **`F-012` Spatial mode.** The mode-bar slot is reserved in Initiative 1; the real Spatial sub-model is a separate initiative outside this arc.
- **`F-004` CLI commands.** `gm-toolkit faction list`, etc. — siblings to the TUI under the `faction` dispatch, but their own initiative.
- **Theming.** Swappable palettes. Single fixed sheet in `internal/faction/tui/styles.go` until a theming initiative ships.
- **Cross-launch persistence.** Remembered preferences. Launch state recomputed each run.
- **Deeper cycle-execution control.** Pause / resume mid-cycle, skip-faction-without-aborting, reorder-on-the-fly. The arc stops at `Esc`-aborts-current-turn.
- **Tabbed faction-detail.** Sub-view growth within Manage (Stats / Assets / Goals / History tabs).
- **Turn post-mortem view.** Sub-view growth within Turn.

## Post-Session Actions

Per session-modes section 8.2 last paragraph, after this Arc-Plan is committed each initiative is added to `planned-work.md` as a Backlog or Up Next entry that references this plan. That step is its own session — not part of the Plan session whose output is this file.

The three entries are:
- `tui-foundation` (feature) — promote to Up Next immediately if Robert is ready to start.
- `tui-manage` (feature) — Backlog, trigger "when `tui-foundation` merges".
- `tui-turn` (feature) — Backlog, trigger "when `tui-manage` merges".

The Out-of-Arc Deferrals list above should also land in `planned-work.md` Backlog as separate entries when this plan commits.

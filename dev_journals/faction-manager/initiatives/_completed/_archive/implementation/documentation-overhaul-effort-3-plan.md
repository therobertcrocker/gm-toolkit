# Documentation Overhaul — Effort 3 Plan (Architecture Overviews)

> Per-effort plan under `documentation-overhaul-plan.md` (the overview plan), per
> `session-modes.md` section 8. This effort builds the `docs/architecture/` tree:
> thin spines + seeded per-subsystem wiki pages (overview plan, Decision 5).
>
> This plan was written after a full read-through of the engine and interface
> source in the Plan session itself (Fable). Page boundaries, scope notes, Key
> Decisions candidates, and code anchors below are verified against source as of
> commit `489cf0b` — execution sessions re-verify against current source but do
> not re-derive them.

## Context / Goal

Effort 3 is the largest effort of D-002 and the one that makes the push-not-pull
decision system real: the architecture pages it produces are the primary read
surface (`session-modes.md` section 10), required reading during Discovery/Plan
re-grounding. Until they exist, two CLAUDE.md routing-table rows and the
re-grounding rule carry "lands in D-002 Effort 3" interim notes pointing at
`docs/architecture-overview.md`.

Deliverables: `docs/architecture/engine/` and `docs/architecture/interface/`
(one `overview.md` spine + seeded subsystem pages each), the arch-page template,
the Effort 1 deferred narrative migration, and the top-level
`architecture/architecture-overview.md` orientation index. (The legacy
`architecture-overview.md` was already deleted in `f09e4cc`, so the planned
supersession became a no-op — see Session 5 in the Decision Record.)

Execution sessions inherit decisions and source pointers from this plan, not
design work. What remains discover-at-write is the page *prose* — the scope,
boundaries, source maps, and Key Decisions candidates are fixed here.

## Decisions Ratified in Planning

1. **Seeded page list — engine (11 pages):** orchestrator, turn pipeline,
   actions, goals, hooks, world & movement, spatial, effect & mutation,
   persistence & static data, narrative digest, logging & errors. Notable
   calls, all verified against source:
   - **World & movement is one page** (was "movement" in the candidate list).
     `engine/world/` is a single package telling one story: the per-turn
     spatial index (`world.go`) plus movement orders (`movement.go`). The
     `HexRouter` interface in `world.go` is the seam to the spatial page.
   - **Tag rides the effect & mutation page; world does not.** The tag and
     effect engines are structural twins — both `ApplyAll` at cycle start to
     register hooks from rulebook data. The page covers tag + effect +
     mutation: how data becomes hooks, how mutations become state.
   - **Logging & errors is one page.** `faction/errors/recoverable.go` defines
     the orchestrator's abort semantics and shipped with the logging
     initiative; splitting them would orphan the error half.
   - **Persistence & static data is one combined page** spanning
     `faction/state/`, `internal/campaigns/`, and `faction/rulebook/` — they
     share the durable decisions (GM-owned data, file layout, Paths bundle).
   - **Three pages beyond the original candidate set** (effect & mutation,
     narrative digest, logging & errors), each clearing the
     accumulated-rationale bar (Decision 5).
2. **Seeded page list — interface (5 pages):** regions, state machine, event
   stream, overlays, views. The candidate four plus views (the views-layer
   composition pattern accumulated real rationale across the TUI rebuild arc).
3. **Spine depth: cross-subsystem narrative + index.** Each `overview.md`
   holds narrative that spans subsystems plus the page index. **Routing rule
   for all future content: anything about a single subsystem lives on that
   subsystem's page; only cross-subsystem narrative may live in the spine.**
   Spine outlines are pre-drafted in this plan (Shared Context below).
4. **Arch-page template: code-map header + Purpose / Shape / Key Decisions /
   Dependencies.** Skeleton pre-drafted in this plan (Shared Context below);
   Session 1 validates it against the orchestrator pilot before the page
   clusters run, then lands it at `docs/process/templates/arch-page.md`. No
   status/last-verified line.
5. **Dependencies section means both axes, briefly:** what the subsystem
   depends on / what depends on it (code axis), with named subsystems linked
   to their pages where they exist (doc axis).
6. **Session slicing: five execution sessions** (Robert's call; four merged from
   an initially proposed six, plus Session 5 added during the Session 3–4
   structural review — see Decision Record). Session 1 first; 2–3 (engine) and 4
   (interface) independent after it; Session 5 (top-level orientation index +
   closeout) runs last because it orients over all 16 pages and its closeout
   re-points routing at the new front door. Clustering principle: pages that mine
   the same source material are written together.
7. **Source maps are fixed in this plan; sessions mine, never re-derive.**
   Each page's decisions-log sections (by decision number from the frozen
   index), completed initiative docs, and code anchors are assigned below.
   Read log sections via the index + `offset`/`limit` — never the full file.
   Every page also checks `architecture-overview.md` for salvage.
8. **Plan-session read-through (process).** The "discover-at-write" license
   covers page prose, not page boundaries. Boundary/seam decisions require
   reading the source, so the Fable Plan session read the codebase before
   writing this plan — three candidate-list corrections (world & movement,
   tag placement, logging & errors) came directly from that read.

## Open Questions — To Ratify at Execution Time

None routed forward. The overview plan routed three questions here (page list,
spine depth, template); all are ratified above. Effort 5 (diagrams) questions
remain with Effort 5.

## Effort 3 Session Summary

| # | Session | Deliverables | Depends on |
|---|---------|--------------|------------|
| 1 | Skeleton + pilot | `templates/arch-page.md`; engine + interface `overview.md` (incl. journal migration); pilot page: orchestrator | — |
| 2 | Engine: turn resolution | Pages: turn pipeline, actions, goals, effect & mutation | 1 |
| 3 | Engine: world + periphery | Pages: world & movement, spatial, hooks, persistence & static data, narrative digest, logging & errors | 1 |
| 4 | Interface cluster | Pages: regions, state machine, event stream, overlays, views | 1 |
| 5 | Orientation index + closeout | `architecture/architecture-overview.md` (top-level front door); supersede legacy `architecture-overview.md`; flip interim routing notes to the front door | 2–4 |

One `wip:` commit per session. All sessions run on **Opus**.

```
            ┌──► Session 2 ──┐
Session 1 ──┼──► Session 3 ──┼──► Session 5 (orientation + closeout)
            └──► Session 4 ──┘
```

## Per-Session Breakdown

### Session 1 — Skeleton + pilot

**Scope:** Everything the page clusters conform to: the template, both spines,
the journal migration, and one real page to validate the template against
(Decision Record, Effort 2 entry 6).

**Deliverables:**

- `docs/process/templates/arch-page.md` — the skeleton from Shared Context,
  with inline HTML-comment guidance matching the other templates. Register it
  in `session-modes.md` section 11.
- `docs/architecture/engine/overview.md` — per the engine spine outline
  (Shared Context). Carries the migrated **AI Decision-Making (Planned)** note
  (a brief, seam-anchored note, not the journal's fuller framing) and a new
  headless-by-design framing in section 1. (Execution reversal: the `## Modes`
  narrative migrates to `interface/overview.md` instead — see Decision Record.)
- The migration is **Effort 1's deferred tail and a sanctioned one-time edit of
  the append-only journal** — ratified in Effort 1 before the freeze semantics
  applied; logged in the Decision Record. Both `## Modes` and `## AI
  Decision-Making (Planned)` move out of `dev-journal-factions.md`; `## Quick
  Reference` and `## Design Principles` stay. The Modes content is rewritten to
  match the implemented modebar (Manage / Turn / Spatial-disabled), not migrated
  verbatim — the journal's Review/Turn/Edit taxonomy is stale against source.
- `docs/architecture/interface/overview.md` — per the interface spine outline.
- `docs/architecture/engine/orchestrator.md` — pilot page (source map below).
  Chosen because it is the richest subsystem and stresses every template
  section.
- Both spine indexes list all 16 seeded pages, with not-yet-written ones
  marked forthcoming — the index is complete from day one.

**Dependencies:** none. Gates everything.

### Session 2 — Engine: turn resolution cluster

**Scope:** The single data-flow story of how a turn resolves into applied
state: turn pipeline, actions, goals, effect & mutation (source maps below).

**Dependencies:** Session 1 (template + spine).

### Session 3 — Engine: world + periphery cluster

**Scope:** The world-shape subsystems (world & movement, spatial, hooks —
sharing the asset-movement-redesign and event-hooks rationale) plus the
standalone periphery (persistence & static data, narrative digest, logging &
errors). Source maps below.

**Dependencies:** Session 1. Independent of Session 2.

### Session 4 — Interface cluster

**Scope:** All five interface pages (source maps below).

**Dependencies:** Session 1 (template + spine).

### Session 5 — Orientation index + closeout

**Scope:** The top-level front door for `docs/architecture/`, then the closeout
that removes every interim pointer in the doc set. The page's job is to orient a
reader and route them — not to re-narrate the subsystems.

**Deliverables:**

- `docs/architecture/architecture-overview.md` — the top-level orientation
  index. Names the engine/interface split and what each side covers, explains
  how the pages work (the spine-plus-subsystem-pages shape, the Key Decisions
  read surface), and routes the reader to the two `overview.md` spines.
  **Orientation only** — no per-subsystem prose (that lives on the pages). Built
  from scratch (Robert's call): the legacy `architecture-overview.md` was already
  deleted in `f09e4cc`, so there was nothing to relocate or salvage — see
  Decision Record.
- Flip the interim routing notes to the new front door
  (`docs/architecture/architecture-overview.md`): the two CLAUDE.md
  routing-table rows and the `session-modes.md` section 10 re-grounding note
  named here, plus any others the closeout sweep surfaces (the Decision Record
  logs three more).

**Dependencies:** Sessions 2–4 — the index orients over all 16 pages, so they
must be written first. Runs last.

## Per-Page Source Maps

Format per page: **Code** (anchors verified in this Plan session) · **Log**
(decision numbers in `archive/decisions-log.md` — read via index +
offset/limit) · **Docs** (under `docs/initiatives/`, `completed/` unless
noted) · **Scope / Key Decisions candidates** (what the page covers; durable
"why"s spotted during the read — candidates, not exhaustive).

### Engine

**orchestrator** (Session 1, pilot)
- Code: `engine/{core,orchestrator,collector,observer,roller}.go`
- Log: 31–39, 46–53, 135–148, 215
- Docs: `core-engine-orchestrator-{discovery,plan}`, `sub-engine-alignment-{discovery,plan}`, `sub-engine-self-bootstrap-plan`, `engine-collector-reshape-{discovery,plan}`, `input-collector-refactor-plan`, `engine-subpackage-layout`, `CODEBASE_REVIEW`
- Scope: composition root (`Engine` owns Rulebook, Roller, Hooks, seven
  sub-engines); the phase pipeline (setup/index-rebuild → goal-lock →
  stat-raise → bookkeeping → movement → action → finish) and its per-phase
  beat (sub-engine call → hook dispatch → `applyAndRecord` → observer →
  checkpoint). Key Decisions candidates: two input contracts
  (`PhaseCollector` vs `action.Collector`) and why they're split; fire-and-
  forget `TurnObserver` (engine returns errors, callers forward — the
  error-return-over-observer rule); recoverable-vs-fatal branching at every
  phase; `applyAndRecord` does *not* save — `state.Save` only at phase gates;
  checkpoint constants as the pause points; single-phase self-registering
  engine construction (log 285–294, resolves R-012).

**turn pipeline** (Session 2)
- Code: `engine/turn/{turn,bookkeeping,order,history,event_record}.go`
- Log: 31–39, 40–45, 54–58, 195–196
- Docs: `turn-engine-discovery`, `testing-suite`
- Scope: turn lifecycle (Start/CurrentFaction/Advance/Abandon, the
  pause/resume cursor in `TurnState`), bookkeeping math (income, maintenance
  with category surcharge, hook-budget reset), deterministic rotation order,
  history as JSONL `EventRecord` appends. Key Decisions candidates:
  bookkeeping idempotent on resume (`BookkeepingApplied` flag); history
  vs. state write cadence (log 195–196); sorted-then-rotated faction order
  for determinism; mutation application deliberately owned by the
  orchestrator, not this engine.

**actions** (Session 2)
- Code: `engine/action/{action,collector}.go`, `engine/action/actions/`
  (`register.go`, `eligibility.go`, ~12 actions)
- Log: 46–53, 64–70, 71–76, 90–95, 96–109, 216–217, 236–238
- Docs: `action-resolution-discovery`, `expand-influence-discovery`, `use-asset-ability-implementation`, `ability-engine-redesign-{discovery,plan}`, `xp-stat-raise`
- Scope: the `Action` contract (Validate → Inputs → Resolve → Output),
  factory registration, the 17-method `action.Collector`, eligibility
  helpers, the ability fold. Key Decisions candidates: factories take the
  collector so each turn gets fresh wired actions; `RegisterDefaultActions`
  takes explicit deps to avoid an engine import cycle; roller resolved at
  factory-invocation time (swappable test seam — fence-signed in
  `register.go`); recoverable sentinels (`ErrNoSelection`, `ErrTurnCanceled`)
  live here so the error classifier needn't import the TUI.

**goals** (Session 2)
- Code: `engine/goal/goal.go`, `engine/goal/goals/`, `engine/goal/locks/locks.go`
- Log: 110–123, 215
- Docs: `goal-engine-{discovery,implementation}`, `goal-engine-mutation-refactor`
- Scope: per-goal `Handler` (CheckLock + UpdateProgress), the lock taxonomy
  (None / Skip / RestrictActions), progress driven by inspecting the turn's
  mutation list. Key Decisions candidates: `UpdateProgress` must run before
  `MutationEngine.Apply` (destructive mutations haven't fired yet); data-only
  goals (TOML entry, no handler) intentionally no-op; sibling `locks` package
  as the import-cycle resolution (log 215).

**effect & mutation** (Session 2)
- Code: `engine/tag/`, `engine/effect/` + `effect/effects/transport.go`,
  `engine/mutation/mutation.go`, `domain/mutation.go`
- Log: 40–45, 54–63, 191–194, 227–235, 239–240
- Docs: `goal-engine-mutation-refactor`, `asset-movement-redesign-effort-2-plan`, `faction-assets-map-plan`
- Scope: the data-driven rule layer and the apply layer. Tag and effect
  engines as structural twins (`ApplyAll` at cycle start, registering hooks
  from rulebook data — e.g. `TransportHandler` → per-asset
  `TransportReactor`); the mutation vocabulary (`domain.Mutation`, Type
  discriminators, Cause fields) and `MutationEngine.Apply`'s switch. Key
  Decisions candidates: EffectsEngine registration is data-driven
  (`def.Transport != nil`, log 227–235); fail-loud `MutationApplyError` with
  per-miss reporting (log 239–240); mutations are the *only* state-change
  currency (engines emit, never mutate); assets-map conversion for entity
  addressing (log 191–194); transport cargo location kept in sync so cargo
  is recoverable if the transport dies mid-flight (`transport.go` comment).

**hooks** (Session 3)
- Code: `engine/hooks/` (`doc.go`, `registry.go`, `interfaces.go`,
  `rule_modifier.go`, `scope.go`), `engine/hooks/dispatch/`
- Log: 158–188
- Docs: `event-hooks` (discovery + implementation), `xp-stat-raise`
- Scope: the five hook categories (doc.go is a ready-made outline), the
  scope system (global → faction → asset lookup order), dispatch. Key
  Decisions candidates: Cat 4 as a typed family rather than one generic
  interface (compile-time safety — stated in doc.go); only Cat 3 recurses,
  bounded at depth 5; reactors recurse on *newly emitted* mutations only
  (no re-firing on originals — `dispatch/mutations.go` comment).

**world & movement** (Session 3)
- Code: `engine/world/{world,movement}.go`; movement phase consumers in
  `orchestrator.go` (`runMovementPhase`, `prepareMovementDecisions`,
  cargo-eligibility helpers)
- Log: 197–210, 218–226, 227–235, 236–240
- Docs: `asset-movement-redesign-{discovery,plan}`, `asset-movement-redesign-effort-{1,2,3}-plan`, `asset-movement-redesign-tick-decision-ordering`, `asset-movement-phase2-commit1-state`, `feature-movement-redesign-pre-merge-{plan,review}`
- Scope: the per-turn spatial `Index` (AssetsByLocation / BasesByLocation /
  AssetsByHex, rebuilt each faction turn, unknown worlds skipped + surfaced
  via `OnIndexSkipped`), movement orders (tick → progress/complete;
  issue/revise/cancel decisions), the `HexRouter` seam to spatial. Key
  Decisions candidates: tick-before-decisions ordering (own completed doc);
  two-dispatch movement phase (log 227–235); in-flight assets indexed by hex
  only; revision costs 1 Coin and re-paths from current hex; cargo manifest
  carried by `BuildMovementMutations`, cargo-follow owned by the transport
  reactor (page cross-link to effect & mutation).

**spatial** (Session 3)
- Code: `internal/spatial/{spatial,region_map}.go`, `internal/spatial/builder/`
- Log: 197–210, 218–226, 269–271, 272–273
- Docs: `spatial-model-{discovery,plan}`, `spatial-model-effort-{1,2}-plan`, `spatial-map-cli-{discovery,plan}`, `spatial-interface-redesign-plan`, `spatial-warp-redesign-plan`
- Scope: the interface ladder (`SpatialMap` → `Location` →
  `RegionLocation`), `RegionMap` (regions, worlds, hex index, bidirectional
  warps, Dijkstra pathing with crossing costs), the multi-phase builder. Key
  Decisions candidates: interface narrowing + `RegionHex` as the routing
  type (log 218–226); warp canonical orientation + dedup (log 269–271);
  derivation state struct, no user-defined generics (log 272–273); package
  deliberately engine-agnostic (no faction imports).

**persistence & static data** (Session 3)
- Code: `faction/state/{faction_state,crud}.go`, `internal/campaigns/`
  (`campaigns.go`, `registry.go`, `scaffold.go`), `faction/rulebook/rulebook.go`
- Log: 1–18, 24–30, 195–196, 262–268
- Docs: `campaign-manager-{discovery,plan}`
- Scope: where data lives and who owns it. State TOML + seed-file mechanism
  (`ensureSeed`); atomic CRUD with in-memory rollback on failed save;
  history JSONL (cross-link turn pipeline); campaign manifest + registry at
  `$GM_TOOLKIT_HOME`; the tag-driven `Paths` bundle; rulebook TOML loading
  (globbed `*_assets.toml`, tags, goals, drift costs). Key Decisions
  candidates: GMs own all rulebook TOML — no embedded defaults, no
  hard-coded data path; `Paths()` derives from struct tags so adding an
  entry is one field (`campaigns.go` comment); TOML for static/state, JSONL
  for history (log 1–18); save-at-phase-gates cadence (log 195–196).

**narrative digest** (Session 3)
- Code: `faction/narrative/{renderer,wire_renderer,wire_templates,history_reader}.go`, `narrative/digest/{types,build}.go`
- Log: 84–89, 124–134
- Docs: `narrative-renderer-{discovery,implementation-plan}`
- Scope: the two-layer split — digest builder (history → typed `CycleDigest`:
  beats, cross events, headline) and `Renderer` interface (wire-service v1).
  Key Decisions candidates: post-hoc narrative from history rather than live
  events (log 84–89); digest as the stable intermediate so renderers can
  multiply; seeded randomness for reproducible copy.

**logging & errors** (Session 3)
- Code: `internal/logging/{logging,handler,banner}.go`, `faction/errors/recoverable.go`
- Log: 241–254
- Docs: `logging-and-errors-{discovery,plan}`, `logging-and-errors-effort-{1,2}-plan`
- Scope: the custom slog handler (flat key=value, suppressed structural
  attrs, banners), per-campaign `LogsDir`, and the recoverable-error
  classifier. Key Decisions candidates: two-channel logger for sub-engines
  (field for internals, parameter for cascaded orchestrator calls — log
  241–250); sentinel list as the *deliberate* definition of the
  orchestrator's abort semantics (`recoverable.go` comment); turn=cycle
  vocabulary fix (log 241–250).

### Interface

All five pages additionally mine `arcs/tui-rebuild/tui-rebuild-arc-{discovery,plan}`
— the arc artifacts are primary source for cross-page conventions (cited by
name in `update.go`).

**regions** (Session 4)
- Code: `tui/layout/layout.go`, `tui/chrome/chrome.go`, `tui/styles/`, root
  layout budget in `tui/update.go` (`headerHeight`, `contentHeight`,
  `resizeSubs`)
- Log: 255–261, 274–284
- Docs: `tui-foundation-{discovery,plan}`, `tui-style-system-plan`
- Scope: the three-column `Region`/`Panel`/`Compose` system, root-owned
  layout budget (subs never see chrome), `StatusLiner` severity contract.
  Key Decisions candidates: root-owned layout budget (log 274–284);
  moving a component = reassigning its Region key (`layout.go` comment);
  styles as a sub-package (log 255–261).

**state machine** (Session 4)
- Code: `tui/{model,update,view}.go`, `tui/views/modebar/modebar.go`
- Log: 77–83, 274–284, 285–294
- Docs: `tui-discovery`, `tui-implementation`, `tui-foundation-discovery`
- Scope: root `Model` (mode bar + subs map + help + confirm-exit), the
  Update Discipline (the MAY/MAY-NOT header in `update.go` — reproduce it),
  capability predicates (`inputCapturer`, `modeLocker`), global keys. Key
  Decisions candidates: key-routing capability predicate over type switches
  (log 274–284); `modeLocker` exists because mode-switch would orphan the
  pumps (cross-link event stream); Update as projection + intent forwarder,
  never engine caller; disabled Spatial slot as growth point.

**event stream** (Session 4)
- Code: `tui/adapter/` (`adapter.go`, `run.go`, `channels.go`,
  `observer.go`, `phase_collector.go`, `action_collector.go`, `smoke.go`),
  `tui/dryrun.go`
- Log: 77–83, 90–95, 255–261, 285–294
- Docs: `tui-foundation-{discovery,plan}`, `tui-turn-plan`, `tui-turn-effort-{1,2,3}-plan`; `discovery/tui-turn-discovery.md` (still in active `discovery/` — flag for housekeeping)
- Scope: the engine⇄TUI bridge. Engine runs in a `tea.Cmd` goroutine;
  unbuffered `askCh` parks the engine on GM input, buffered `eventCh` (64)
  carries observer events; pump-re-arm pattern; `StreamClosedMsg` as the
  teardown signal so buffered events render before reset; append-only
  `AskKind`/`EventKind` enums; adapter-built display payloads (views never
  hold engine pointers). Key Decisions candidates: goroutine bridge (log
  77–83, extended 90–95); StreamClosed-not-EngineDone teardown
  (`channels.go` comment); synthetic `EvtCycleStarted` emitted
  single-threaded before `RunCycle`; smoke/dryrun harness.

**overlays** (Session 4)
- Code: `tui/views/turn/overlay/` (`overlay.go`, `registry.go`, ~18 prompt
  overlays, `archetypes.go`, `formhelp.go`)
- Log: 274–284, 285–294
- Docs: `tui-turn-effort-{1,2,3}-plan`, `tui-turn-code-review-fixes-plan`; `discovery/action-result-panel-discovery.md` (active — pending work, check `planned-work.md` before writing)
- Scope: the `Overlay` interface (Init/Update/View/Help), `OverlayDoneMsg`
  flow (overlays never touch ask/Reply plumbing — the execution sub-model
  owns the channel send; `overlay.go` comment), the `ImplementedActions`
  gate, form archetypes. Key Decisions candidates: overlay/adapter import
  edge kept acyclic; execution overlay scoped to the center region (log
  285–294, revised plan Decision 11); Esc-declines for movement/cargo (log
  285–294).

**views** (Session 4)
- Code: `tui/views/manage/` (`manage.go`, `list/`, `detail/`, `wizard/`,
  `deleteconfirm/`, `msgs/`), `tui/views/turn/` (`turn.go`, `setup/`,
  `execution/`, `msgs/`)
- Log: 274–284, 285–294
- Docs: `tui-manage-{discovery,plan}`, `tui-turn-plan` + effort plans, `tui-implementation`
- Scope: the composition pattern — each mode is a router over sub-views
  (manage: list/detail/create/deleteConfirm with a `backTarget` map; turn:
  setup/execution with adapter lifecycle per cycle), message sub-packages,
  the execution view's band layout (rail / banner + wizard band + stream /
  detail card). Key Decisions candidates: message sub-packages for
  import-cycle avoidance (log 274–284); adapter built per cycle, discarded
  on return to setup (`turn.go`); race-safe display snapshots
  (`factionSnapshot`, `movableLine` — `execution.go` comments); fixed
  wizard band height sized to huh's 10-option cap (`execution.go` comment);
  atomic CRUD mutations from the wizard (log 274–284).

## Shared Context

### Arch-page template skeleton (Session 1 validates, then lands in `templates/arch-page.md`)

```markdown
# <Subsystem Name>

> **Code:** `internal/<path>` (, `internal/<path>`…)

## Purpose

<!-- 2–5 sentences: what this subsystem is for and the one-line version of
how it does it. A Discovery reader should know after this section whether
this page is the one they need. -->

## Shape

<!-- The structural story: the package's key types/interfaces, the data flow
through them, and the boundaries with neighbors. Prose-first; small code
signatures where they carry more than prose. This is the longest section. -->

## Key Decisions

<!-- The durable "why this subsystem is shaped this way," topic-organized.
One bold lead-in per decision + rationale. This is the primary read surface
of the push-not-pull system: pre-merge promotion appends here. Provenance
links (plan docs, frozen-log numbers) welcome but optional. -->

- **<Decision>** — <rationale>.

## Dependencies

<!-- Both axes, briefly: what this subsystem depends on and what depends on
it (code axis), each named subsystem linked to its page where one exists
(doc axis). -->
```

### Engine spine outline (`docs/architecture/engine/overview.md`)

1. **What the engine is** — the SWN faction-turn engine: a composition root
   (`engine.Engine`) owning a rulebook, a roller, a hook registry, and seven
   sub-engines, driven by a phase-oriented orchestrator. Leads with
   **headless-by-design**: the collector/observer seam backs three drivers
   (test harness, UI, AI planner), none privileged. One short paragraph.
2. **AI Decision-Making (Planned)** — brief seam-anchored note: the
   source-agnostic collector means an AI planner drives the engine through the
   same interface the GM uses. (Modes moved to the interface spine — Decision
   Record.)
3. **How a cycle runs** — one tight walkthrough of the phase pipeline,
   each phase naming the page that owns it. (Effort 5 embeds the flow
   diagram here.)
4. **Cross-subsystem conventions** — the spine-worthy narrative found in the
   read-through: handler registry keyed by data ID (three instances —
   tag/effect/goal; action factory slice and scope-keyed hook registry are
   distinct kin); data-only TOML entries silently and intentionally skipped;
   mutations as the only state-change currency; collectors in / observer out;
   recoverable vs. fatal error discipline; two-channel logging;
   save-at-phase-gates.
5. **Page index** — table: page, one-line scope. All 11 pages from day one.

### Interface spine outline (`docs/architecture/interface/overview.md`)

1. **What the interface side is** — the Bubble Tea TUI today; the side is
   named "interface" because it also covers the CLI rebuild and any future
   frontend (Decision Record, Effort 2 entry 11).
2. **Modes** — migrated from the journal but rewritten to match the modebar
   (`modebar.go`): Manage (the Review/browse surface), Turn, Spatial-disabled
   (reserved slot, unsettled direction toward a faction-utilities/edit surface).
3. **Composition story** — root model (chrome + mode bar + subs map), modes
   as sub-models, engine on a goroutine behind the adapter, asks park the
   engine / events stream past it.
4. **Cross-subsystem conventions** — the Update Discipline (MAY/MAY-NOT);
   message sub-packages to break import cycles; views hold display
   snapshots, never engine pointers; lipgloss only inside `tui/`;
   capability predicates over type assertions at the root.
5. **Page index** — table, all 5 pages.

### Target tree

```
docs/architecture/
  architecture-overview.md   ← top-level front door (Session 5)
  engine/
    overview.md          orchestrator.md      turn-pipeline.md
    actions.md           goals.md             hooks.md
    world-movement.md    spatial.md           effect-mutation.md
    persistence.md       narrative-digest.md  logging-errors.md
  interface/
    overview.md          regions.md           state-machine.md
    event-stream.md      overlays.md          views.md
```

## Decision Record — Execution

*(Append and reconcile: reversals of planned decisions are logged here and the
plan body fixed. Architecturally-durable entries graduate to arch-page `Key
Decisions` at pre-merge — for this effort that promotion is usually a no-op,
since the deliverables ARE the Key Decisions surfaces. Session 1 must log the
sanctioned journal edit here.)*

### Session 1 — Skeleton + pilot

- **Sanctioned journal edit (planned, executed).** Removed `## Modes` and `## AI
  Decision-Making (Planned)` from `dev-journal-factions.md` — the append-only
  journal's temporary Design narrative. Ratified in Effort 1 before the freeze
  semantics applied. The journal retains Overview, Quick Reference, and Design
  Principles.
- **Modes narrative → interface overview, not engine (reversal).** The plan
  (Session 1 deliverables + engine spine outline section 2) routed the Modes
  narrative into `engine/overview.md`. Robert's call at execution: modes
  describe how the *user drives* the engine — an interface concern, not an
  engine concern. Modes now lives in `interface/overview.md`. The engine
  overview instead leads with a **headless-by-design** framing naming the three
  drivers the collector/observer seam supports: test harness, UI, AI planner.
  Plan body reconciled below.
- **AI Decision-Making stays engine-side, trimmed (partial reversal).** Robert
  kept the AI Decision-Making (Planned) narrative in `engine/overview.md` —
  anchored on the source-agnostic collector seam being an *engine* property —
  but as a brief forward-looking note, not the journal's fuller framing.
- **Modes rewritten to match the modebar, not migrated verbatim (grounding
  correction).** The journal's Review/Turn/Edit taxonomy is stale against
  `modebar.go` (Manage / Turn / Spatial-disabled). Per the docs-reflect-source
  rule, the interface Modes section was rewritten to the implemented set with
  Robert's mapping: Manage = the Review/browse surface, Turn = Turn, Spatial = a
  reserved disabled slot whose unsettled direction is a faction-utilities/edit
  surface. This overrides the plan's "editorial fitting only" for the *Modes*
  content; editorial-fitting still governs the AI note.
- **Orchestrator pilot re-grounded against source.** Three claims carried from
  the plan's prose were corrected against current source while authoring the
  pilot: `Tag.ApplyAll` registers from faction state, not the rulebook (only
  `Effect.ApplyAll` takes the rulebook); the "registry+handler keyed by data ID"
  convention is **three** instances (tag/effect/goal), with the action factory
  slice and the scope-keyed hook registry as distinct kin — not "five"; and the
  per-phase "beat" is a synthesis, not a literal invariant (goal-lock notifies
  before applying). Establishes the session standard: every load-bearing
  arch-page claim is verified against current source, never carried from the
  plan, the frozen log, or other docs.

### Session 2 — Engine: turn resolution cluster

Pages written: `turn-pipeline`, `actions`, `goals`, `effect-mutation`. Index
rows flipped to Written.

- **Per-page catalogues added beyond the seeded scope (Robert's call).** `actions`
  gained an **Actions catalogue** (12-row table: prompts → emitted mutations) and
  `goals` a parallel **Goals catalogue** (advances-by / completes-when), each with
  a "rules behind the numbers" pointer to `swn-faction-mechanics.md` to hold the
  arch pages at structural altitude. `effect-mutation` carries the canonical
  **Mutation catalogue** (all 31 `domain.Mutation` types, grouped). These are
  additive to the plan's page scope, not reversals.
- **Effect & mutation rule layer expanded (Robert's call).** The seeded scope
  framed tag/effect as "data → hooks" with the transport example. On review that
  undersold both engines, so the rule-layer section was expanded to: the five hook
  categories the four implemented tag handlers span (Warlike/Fanatical/Scavengers/
  Preceptor), the `S`-flag effect engine as the home for *all* passive/reactive
  asset features (transport wired; Zealots/Blockade Fleet/Integral Protocols/etc.
  data-awaiting), and the `S` vs `A` vs `P` flag split. Hook-category *mechanics
  and dispatch* deferred to `hooks.md` (Session 3) per the spine routing rule.
- **Re-grounding corrections against source.** (1) `turn-pipeline`: `bookkeeping.go`'s
  doc comment claims `ApplyBookkeeping` "advances the turn phase to `PhaseAction`,"
  but `TurnState` has no phase field — claim dropped (stale source comment, flagged
  for a later one-line code fix). (2) `goals`: the plan framed `CheckLock` as a gate
  only; source shows it is *also* the time/condition-driven advancement path
  (ticks, transit completion), distinct from `UpdateProgress`'s event-driven path —
  page documents both beats. (3) `effect-mutation`: confirmed the tag/effect
  source-of-truth split (tag handlers a fixed Go set over faction-state assignments;
  effect handlers registered data-drivenly from the rulebook) the pilot first noted.
- **Tag/data discrepancy found, fixed by Robert mid-session.** `scavengers.go`
  keyed its handler to `T-014`, but `tags.toml` lists `T-016` as Scavengers
  (`T-014` is Psychic Academy). Surfaced while authoring the tag table; Robert
  corrected the const. The page cites tags by name, not number, so it was
  unaffected.

### Session 3 — Engine: world + periphery cluster

Pages written: `hooks`, `world-movement`, `spatial`, `persistence`,
`narrative-digest`, `logging-errors`. All six index rows flipped to Written.

- **Narrative recap is unwired — page corrected, `F.021` added (re-grounding +
  planned-work).** The plan's source map and a first draft both implied a CLI/TUI
  recap surface consuming the digest. Source check (`findReferences` on
  `NewWireRenderer` and `LoadCycle`) shows **only tests** consume the `digest`
  builder and wire `Renderer` — no `cmd/` command, no TUI view. The
  `narrative-digest` page now states the system is built-and-tested but unwired,
  and a new Backlog feature `F.021` (Narrative Recap Surface) was added to
  `planned-work.md` per the "if it doesn't exist, add it" instruction.
- **Registry filename corrected against a stale comment (re-grounding).** The
  `campaigns` doc comments name the registry `$GM_TOOLKIT_HOME/campaigns.toml`,
  but `RegistryPath()` returns `campaigns-registry.toml`. The `persistence` page
  documents the real filename. (Stale source comment; candidate for a one-line
  code-comment fix, not made here.)
- **Recoverable sentinel list is three, not two.** The Session 2 `actions` page
  cited `ErrNoSelection`/`ErrTurnCanceled`; `recoverable.go` also lists
  `ErrActionUnavailable`. The `logging-errors` page documents all three.
- **Dispatch sites verified, not carried.** Hook dispatch was confirmed against
  source rather than the plan prose: Cat 1/2/5 from `action/actions/attack.go`,
  Cat 3 from three `orchestrator.go` sites, Cat 4 from `buy_asset.go` and
  `turn/bookkeeping.go`. This grounded the `hooks` page's "dispatch is
  distributed, not centralized" framing.
- **Two scope notes surfaced beyond the seeded maps (additive, not reversals).**
  `world-movement` documents that **only immobile (`Speed == 0`) assets are
  cargo-eligible** (`eligibleCargoForTransport`); `spatial` foregrounds the
  package's **two halves** — a runtime `RegionMap` and an offline authoring
  `builder` that compiles an ASCII hex grid to canonical TOML — which the seeded
  scope under-weighted.

### Structural review (between Sessions 3 and 4)

A review of the landed engine pages (Opus, no code written) produced three plan
amendments, logged here as provenance:

- **Catalogue convention codified.** The data-backed catalogue element that
  emerged across Sessions 2–3 (Actions/Goals/Mutation catalogues) was made an
  explicit optional section in `templates/arch-page.md`, scoped to pages with
  rulebook data behind them — so interface pages (Session 4) know it isn't for
  them. Additive to the template, not a reversal.
- **Tag discoverability fixed.** `effect-mutation.md` is the canonical home for
  faction tags, but neither its title nor the spine index row surfaced the word.
  Title → "Tags, Effects & Mutation"; the engine `overview.md` index row → "Faction
  tags, asset effects, and the mutation apply layer." Page kept combined (tag and
  effect are structural twins sharing the registrar/`ApplyAll` shape — splitting
  would orphan that framing); filename and inbound link text left untouched (the
  gap was the visible words, not the link target).
- **Session 5 added; closeout moved out of Session 4 (amends Decision 6).** The
  flat-per-side tree has no top-level entry point at `docs/architecture/` — a
  reader landing there sees two directories and must guess. Session 5 adds
  `architecture/architecture-overview.md` as the orientation front door. The
  closeout (legacy supersession + interim-pointer flips) moved from Session 4 to
  Session 5 so the routing notes target the front door rather than the bare
  directory, and so the front door exists before anything points at it. Session 4
  is now the interface cluster only. Decision 6, the session table, and the
  dependency diagram are reconciled above.

### Session 4 — Interface cluster

Pages written: `regions`, `state-machine`, `event-stream`, `overlays`, `views`.
All five interface index rows flipped to Written. No reversals of planned
decisions — every page matched its seeded scope and source map. Notes:

- **Stale `OverlayDoneMsg` comment found (re-grounding).** `overlay.go`'s
  `OverlayDoneMsg` doc still reads *"Cancel is deferred to Effort 3 … Effort 1
  needs only the happy-path Answer,"* but Effort 3 has since landed:
  `archetypes.go`'s `cancelOnEsc` emits `OverlayDoneMsg{Answer:
  action.ErrTurnCanceled}`, which the orchestrator classifies recoverable. The
  `overlays` page documents the live cancel flow, not the stale "deferred"
  framing. (Stale source comment; candidate for a one-line fix, not made here.)
- **Disabled-Spatial framing reconciled across two surfaces (no change).** The
  `view.go` placeholder reads *"Spatial — reserved for F-012 (Spatial Map CLI)"*
  while the Session-1 interface `overview.md` Modes section frames the slot's
  direction as a faction-utilities/edit surface. Both are accurate to their own
  grounding. The `state-machine` page documents the slot *mechanics* (registered
  absence: rendered, skipped on cycle, placeholder view) and quotes the
  placeholder string verbatim, without re-litigating the Modes narrative — the
  spine owns that per the routing rule.
- **`action-result-panel` housekeeping confirmed (no false claim).** The plan
  flagged `discovery/action-result-panel-discovery.md` as active pending work.
  Confirmed against `planned-work.md` (`F.018`, UX Pass: Confirm-on-Choice — "First
  instance ships in the action-result-panel work"); the panel is not built, so the
  `overlays` page documents only the shipped prompt surface and makes no
  action-result-panel claim.
- **Three additive structural stories surfaced beyond the seeded maps (not
  reversals).** `regions` foregrounds the **three-layer styles cascade**
  (palette → tokens → semantic, hex confined to `palette.go`) and the
  `chrome.StatusLiner` severity-in/style-out contract, which the seeded "styles as
  a sub-package" line under-weighted. `overlays` documents the **Esc-meaning-as-
  help-keymap** contract (`formHelp`/`declineFormHelp`/`cancelFormHelp` encoding
  no-empty / legal-skip / cancel). `views` foregrounds the **detail card keying
  off the mounted modal, not turn events** (the engine fires `MovementTicked`
  before the movement ask, so an event-derived phase would lead the modal).

### Session 5 — Orientation index + closeout

Front door written: `architecture/architecture-overview.md`. Interim routing
notes flipped. Index reconciled.

- **Salvage + supersede dropped; built from scratch (reversal + reconciliation).**
  The plan (Context/Goal, Session 5 deliverables) had Session 5 relocate the
  legacy `architecture-overview.md` into the tree, salvage its unique content,
  and banner the original as superseded. Two facts collapsed that: (1) the legacy
  file was **already deleted** in commit `f09e4cc` (asset-movement redesign), so
  there is no file to relocate or banner; (2) Robert's call at execution was to
  **build the front door from scratch** rather than mine the deleted legacy
  content. The front door is therefore a fresh orientation-only page; no salvage,
  no supersession banner. Plan body reconciled below.
- **Three named interim notes flipped, plus three more found in the closeout
  sweep.** The plan named the two CLAUDE.md routing-table rows and the
  `session-modes.md` section 10 re-grounding note. The closeout's mandate is to
  remove *every* interim pointer in the doc set (Session 5 scope), so a sweep
  caught three more live ones, all flipped to the new front door: CLAUDE.md
  pre-merge checklist step 2 (Robert's call — confirmed in-session), the
  `templates/discovery.md` Architecture Context note, and the broken
  `dev-journal-factions.md` Quick Reference link (pointed at the deleted legacy
  path). Frozen `completed/` artifacts and the plan doc's own historical prose
  were left untouched.
- **Second sanctioned journal touch.** Repointing the `dev-journal-factions.md`
  Quick Reference link is a closeout pointer-fix on the append-only journal —
  the link target was deleted, so the row was already dangling. Logged here for
  the same reason Session 1's journal edit was: the journal's freeze semantics
  make any edit a sanctioned exception.

## Out of Scope

- **Diagrams** — Effort 5; pages may leave a `<!-- diagram: ... -->`
  breadcrumb where one is wanted, nothing more.
- **Contributor guides** — Effort 4 (`docs/contributing/`).
- **Exhaustive page stubbing** — subsystems not listed here (`domain` types,
  `testharness`, the Cobra `cmd/` layer) earn pages later, when rationale
  accumulates (Decision 5).
- **Rewriting the migrated Modes/AI narrative** — editorial fitting only;
  substantive redesign of the modes story is not this effort.
- **Housekeeping flagged, not fixed:** `tui-turn-discovery.md` still sits in
  active `discovery/` despite the initiative having shipped; moving it is a
  one-line chore outside this effort.

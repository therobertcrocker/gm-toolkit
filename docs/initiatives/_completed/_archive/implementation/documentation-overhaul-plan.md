# Documentation Overhaul — Implementation Plan

## Context / Goal

The TUI Rebuild arc just closed. The documentation set has accumulated organically and no longer matches how the project actually works: some artifacts are write-only, others are stale, and the process docs predate formats we now use (Arcs, effort-based initiatives, multi-effort work). This initiative brings the docs in line with the real shape of development and builds the durable artifacts the project has been missing.

This is a **multi-effort `docs` initiative**. It exceeds the single-session `docs` flow (which `session-modes.md` §7 anticipates: "if a docs initiative grows large enough… promote it"), so it borrows the effort structure from the feature flow — one planning doc here, then one (or a few) execution sessions per effort. Unlike a feature, the *generative* efforts discover-at-write: the act of writing an architecture page **is** the act of exploring and documenting that subsystem, so the per-doc detail is not pre-specified here the way code tasks would be.

The work splits into two tracks:

- **Meta track** — decisions about existing process/journal artifacts (Efforts 1–2). Upstream: these decisions shape what the generative docs describe.
- **Generative track** — new docs written by exploring the codebase (Efforts 3–5).

Design rationale for every decision below was worked out in the planning session; this doc is the authoritative record (there is no separate discovery doc for this initiative).

## Decisions Ratified in Planning

1. **Process model = multi-effort `docs` initiative.** One planning doc (this file) covers all efforts; each effort is its own execution session (large efforts get their own lightweight Plan session first). Lightest structured shape that still gives ordering and shared decisions. Heavy per-doc Discovery is skipped because generative docs discover-at-write.

2. **`decisions-log.md` is frozen and archived, not killed.** The existing 294 decisions are kept as a frozen historical record. The per-commit *write* discipline is retained (Robert does not object to the write tax) but **demoted from primary read surface to provenance**. The log's failure mode was never storage — it was that it was a *pull* resource organized by branch/time, an axis nobody retrieves on. Decisions now live where they get *read*.

3. **Push-not-pull retrieval system (cross-cutting).** Durable decisions are attached to moments that cannot be skipped, organized by the axis people query on (subsystem), via three surfaces:
   - **Surface 1 — Architecture wiki pages (primary).** Each per-subsystem page carries a `Key Decisions` section holding the durable "why this subsystem is shaped this way," topic-organized. Reading the relevant arch page becomes a **required step** in the Discovery template and the re-grounding rule — so the rationale is in the doc you are *forced* to open. Pull becomes push.
   - **Surface 2 — Code fence-signs.** A one-line `// deliberate: X, not Y, because <reason>` at the fence itself, surfaced during execution/refactor re-grounding. CLAUDE.md's "no comments unless the WHY is non-obvious" rule already licenses exactly this case.
   - **Surface 3 — Plan-doc decision records (provenance).** Each initiative's execution-time decisions are logged in its own plan doc (kept write discipline) and ship to `completed/` as source material — not the routine read surface.
   - **Promotion step.** The pre-merge checklist graduates architecturally-durable decisions from the plan-doc record into the relevant arch wiki page's `Key Decisions`. Keeping a descriptive doc in sync with the code it describes is the baseline cost of having an architecture doc at all, not new tax. Micro-decisions (naming, a test strategy) stay as plan-doc provenance + code comments; only architecturally-significant "why" graduates.

4. **`dev-journal-factions.md` → append-only north-star / index.** It becomes a stable navigational spine: links to other docs (Quick Reference) plus durable Design Principles. It holds nothing that needs *editing* — only *adding*. Its mutable content evicts: Feature List + Open Questions → `planned-work.md`; the Modes / behavior narrative → the Architecture Overview (Effort 3).

5. **Architecture docs = thin spine + per-subsystem wiki pages.** Not a sectioned monolith. One stable `overview.md` spine per side (engine, interface) that indexes into one page per subsystem. Each page uses a uniform template (Purpose / Shape / Key Decisions / Dependencies). This partitions growth along the *subsystem* axis (which matches the code's change axis), keeps each Discovery read scoped to one file, and reinforces the push mechanism in Decision 3. Pages are **seeded for subsystems that already have accumulated rationale and grown organically** — no front-loaded stubs. The exact page list is an Effort-3-Plan concern.

6. **Contributor guides = a shared front door plus two per-side guides, organized around extension recipes.** *(Reversed at Effort 4 planning — see Decision Record, Effort 4. Original framing: a single stable, non-wiki-fied doc per side.)* `docs/contributing/overview.md` holds setup/build and the shared extension pattern; `engine.md` and `interface.md` carry side-specific test tooling, conventions, and a set of per-subsystem **extension recipes** ("Add an action," "Wire a prompt"). Each recipe links its architecture page for shape and rationale and spends its own words only on steps — task-oriented still, but structured on the same subsystem axis as the arch wiki. Effort 4 gets its own Plan session (`documentation-overhaul-effort-4-plan.md`) to settle the recipe template, coverage, and slicing.

7. **Docs stay portable markdown in git; tooling is viewer-agnostic.** No tool-specific syntax committed (no `[[wikilinks]]`, Dataview, callouts, transclusion). The primary consumer during the workflow is the CLI agent (Read/Grep) plus PR review on GitHub; the lowest-common-denominator format is a feature, not a limitation. Obsidian or any editor may be pointed at the repo as a personal viewer for free. Diagrams use **Mermaid** (renders on GitHub, in-IDE, and in any future doc-site).

### Mutability contracts (state explicitly to avoid blur)

Two artifacts live in the same dev-journal folder with **opposite** contracts; the process docs (Effort 2) must say so:

- **Plan-doc decision record** — *mutable*: append **and reconcile** (when execution overrides a planned decision, log the reversal with rationale and fix the plan body so it doesn't lie).
- **North-star journal** — *append-only*: never edited, only added to.

And two indexes on **different axes**, kept distinct and cross-linked:

- **Dev-journal north-star** indexes *process/tracking* artifacts (process axis).
- **Architecture spine** indexes *system* pages (system axis).

## Open Questions — To Ratify at Implementation Time

- **Effort 1:** Resolved — see Decision Record, Effort 1 entries 1–3.
- **Effort 2:** Resolved — see Decision Record, Effort 2 entries 1–11.
- **Effort 3:** The seeded subsystem page list for the engine and interface sides. Spine depth. The page template (Purpose / Shape / Key Decisions / Dependencies) finalization — **includes defining the template itself (Decision Record, Effort 2 entry 6)**.
- **Effort 5:** Resolved — see Decision Record, Effort 5 entries 1–3.

## Decision Record — Execution

*(Append and reconcile: reversals of planned decisions are logged here and the plan body fixed. Architecturally-durable entries graduate to arch-page `Key Decisions` at pre-merge.)*

**Effort 1 (2026-06-08):**

1. **Frozen log archived by moving to `archive/`** — chosen over an in-place `ARCHIVED` banner; the path change itself signals the freeze, and references were updated at the same time.
2. **Feature List dropped** — not kept as a "shipped features" reference; planned-work tracks forward work, git history tracks shipped work. Added `completed-work.md` to the dev_journal to track shipped features if a reference is desired.
3. **Open Questions moved into `planned-work.md` as a table** with a `Backlog Item` column tying each question to the item that answers it.

**Effort 2 (2026-06-10):**

1. **Arc is a tracker Type** (`A.` ID prefix) — supersedes the "arc is not a Type, only a granularity" framing. The arc's row stays in Current Initiatives for the arc's whole duration; constituent initiatives get their own rows referencing the arc plan.
2. **Effort planning is overview + as-needed** — the first Plan session always produces the top-level plan (shared decisions, effort slicing, DAG); only large or design-heavy efforts get their own lightweight Plan session. Codified as the new `session-modes.md` section 8.
3. **Concept artifact splits by level** — an initiative's Concept is its planned-work write-up (now named as such); an arc's Concept is a standalone templated doc in `docs/initiatives/arcs/<arc-name>/`.
4. **Docs flow is tiered in place** — single-session (default) vs. multi-effort docs initiative; growth never reroutes a docs initiative to `feature`/`refactor`. Reverses the old section 7 guidance.
5. **Template set:** new `docs-plan.md` + `arc-concept.md`; `discovery.md` gains a required Architecture Context section (the push mechanism); `implementation-plan.md` gains a Decision Record section. No separate docs-discovery template — generative docs discover-at-write.
6. **Arch-page template deferred to Effort 3** — defined alongside the first real pages it is validated against, per the no-front-loaded-stubs principle.
7.  **Fence-sign codified as planned** — `// deliberate: X, not Y, because <reason>`, one line, under CLAUDE.md Code Style; binding during re-grounding unless explicitly revisited.
8.  **Arc artifacts are permanent in `arcs/<arc-name>/`** — no move to `completed/` on arc completion (matches tui-rebuild on disk; may revisit an `arcs/completed/` split if arc count grows).
9.  **`session-modes.md` renumbered** — Efforts is the new section 8; Arc-Level Flow → 9, Session Discipline → 10, Templates → 11. This plan's "§9 (decisions-log scope)" deliverable landed in section 10 as "Decision records."
10. **Fable model tier added** — Fable for structure-setting sessions whose output constrains many future sessions (Arc-Discovery, Arc-Plan, process redesign); Opus stays the default for initiative-level Discovery/Plan and reviews. Effort 3's Plan session runs on Fable: the page decomposition is a one-shot taxonomy decision.
11. **Architecture wiki sides renamed `engine` / `interface`** (was `backend` / `tui`) — "engine" is the project's native vocabulary; "interface" covers the TUI, the CLI rebuild, and any future frontend without needing a third spine. Applies to `docs/architecture/` and `docs/contributing/`.

**Effort 4 (2026-06-11):**

1. **Contributor-guide shape reversed to hybrid recipe-based (reverses Decision 6).** Decision 6 scoped the guides as one task-oriented, non-wiki-fied doc per side, with Effort 4 small and straight-to-execution. Robert's call: organize each guide around per-subsystem **extension recipes** ("Add an action," "Implement a tag handler," "Wire a prompt") that cross-link the Effort 3 architecture pages for shape and rationale. This restructures the guides on the subsystem axis without abandoning task-orientation — the recipe *is* the unit of task. Decision 6 body reconciled above.
2. **Three deliverable docs, not two.** A shared `contributing/overview.md` front door (setup/build is identical for both sides — one Go module, one binary) plus the two per-side guides. Reconciled in the Effort 4 breakdown and the Effort Summary above.
3. **Effort 4 promoted to a Plan session.** No longer straight-to-execution; `documentation-overhaul-effort-4-plan.md` settles the recipe template, the six-engine/three-interface coverage set, the per-side handling of the cross-boundary "Add an action" seam, and a two-session slice. That Plan session ran on Opus; execution also runs on **Opus** — Robert's call: he does not trust Sonnet for documentation work, overriding the table's Sonnet default for docs execution.

**Effort 5 (2026-06-12):**

1. **Inventory = 12 diagrams.** The 10 placeholders Effort 3 seeded in-page (discover-at-write), plus two additions: a routing/flow diagram in `interface/state-machine.md` (the named "TUI state machine" plan candidate, the most flowchart-shaped page — the Update priority tree) and a top-level two-sides diagram in `architecture-overview.md` (the engine↔interface seam: collector in / observer out). `engine/effect-mutation.md` was considered and **left without a diagram** — the mutation-as-currency flow is already carried by the phase-pipeline diagrams (#1/#2). **Reconciled (2026-06-13): final inventory is 11 Mermaid + 2 ASCII-carried.** The two interface wireframes (regions #9, views #11) were dropped — built as `block` they fought sizing and nesting-inversion, and each page already had an Effort-3 ASCII chrome diagram that is WYSIWYG-superior for terminal frames (a diagram only lives where it earns its surface; cf. effect-mutation). Net vs. 12: −2 wireframes, +1 (orchestrator carries two flowcharts — fixed phase order + per-phase five-step beat — where this entry counted one) = 11. No placeholders remain. effect-mutation diagram-less stands.
2. **All embedded inline; no `diagrams/` directory.** Every diagram lives in its page's Mermaid fence. `docs/architecture/diagrams/` is created only if a diagram outgrows its page; none currently do. (Narrows the plan's "standalone where a diagram earns its own surface" — none earned it.)
3. **Level of detail matches each page's prose altitude.** Conceptual flow — named beats/phases as nodes, no code-level exhaustiveness — stands as authoring *guidance*. Final visual composition and styling are the author's discretion (see entry 4).
4. **Diagrams are hand-authored by Robert, not Claude-generated.** Reverses the plan's assumption that Effort 5 is a Claude execution effort. Rationale: a Claude pass produced valid, on-palette (Catppuccin Frappé) Mermaid, but Mermaid's dagre auto-layout couldn't be driven to the desired composition through theme/config tuning — nested subgraphs with cross-boundary and back edges sprawl, and `direction` inside a connected subgraph is ignored. Robert authors the Mermaid by hand. The agent's Effort 5 role is reduced to **marking the 12 slots** (`<!-- diagram: … Effort 5 -->` placeholders, all in place) plus eventual mechanical wiring/closeout. Decision 2 is **preserved** — hand-authored Mermaid is still inline and GitHub-renderable per Decision 7; no `diagrams/` directory and no image-file pipeline. **Reconciled (entries 5–8): the hand-authored stance was reversed in execution — a type-by-topology diagram kit lets the agent draft, with Robert reviewing layout. The dagre blocker was wrong-tool, not authoring. Slot-marking stands as the starting point.**

5. **Diagram tooling = a type-by-topology kit (reverses entry 4's authoring stance).** Diagrams are drafted by the agent off a committed kit (`docs/process/diagram-kit.md`) and reviewed by Robert, not hand-authored from scratch. Entry 4's dagre blocker was a wrong-tool problem: it dissolves once type follows content shape — **`block-beta`** for spatial layouts/seams, **`flowchart`/dagre** for flows, **ELK flowchart or `sequenceDiagram`** for cyclic/cross-actor topologies.
6. **`layout: fixed` rejected; diagrams declared `block-beta`.** `fixed` resolves only in the mermaid.ai visual editor, not the open-source renderer GitHub and the IDE ship — fails Decision 7. Block diagrams declare `block-beta` (documented, GitHub-confirmed); bare `block` is reserved for nested composite blocks. **Reconciled (2026-06-13): spatial diagrams declare bare `block`, not `block-beta`** (Robert's call — it renders where the diagrams are read). The `layout: fixed` rejection stands. Existing `block-beta` references were left untouched per Robert; the diagram-kit gotcha that prescribes `block-beta` was reconciled to bare `block` at pre-merge (all five kit references flipped; the gotcha now carries the column-span lesson rather than the falsified renderer-support claim). Lesson: column spans (`:n`) are reliable only at the top level of a *flat* grid — a span inside a nested block has nothing to stretch into.
7. **Palette confirmed Frappé (entry-4 palette stands), now systematized.** Catppuccin Frappé retained, with a role→hue map (external/seam/core/process/decision), subdued semantic *fills* on a `#303446` canvas, and **neon-pink `#ff2e97` edges on flowcharts** (muted block-arrows on spatial diagrams). A neon-bright variant was tried and rejected as too loud. **Reconciled (2026-06-13): palette is Catppuccin *Mocha* on a `#1e1e2e` canvas, not Frappé/`#303446`** — the committed kit (`diagram-kit.md`) standardized on Mocha and all 11 diagrams use it. The role→hue map and neon-pink `#ff2e97` flowchart edges stand; sequence-diagram signal lines are kept quiet (`#cdd6f4`) because dense arrows would scream in neon.
8. **`type:` resolved across the 12 placeholders.** Each brief reconciled to the kit: #8 seam flowchart→`block-beta` (built as the kit's reference example); wireframes #9/#11 → `block-beta`; #12 event-stream → `sequenceDiagram`; #10 priority tree → `flowchart`; #6 hooks → ELK flowchart (recursion back-edge); the six dagre-safe flows confirmed. **Reconciled (2026-06-13): #9/#11 wireframes dropped (entry 1); #8 seam built as bare `block`; #10/#12 shipped as planned.**
9. **`stateDiagram-v2` explored and rejected for `state-machine.md`; the flowchart cascade shipped (2026-06-13).** A true mode/overlay state machine was drafted (modes as states, the `modeLocked` condition as a `Turn` substate — genuinely the better fit for the page's "state machine" title). It rendered as chaos: off-palette green edge-label backgrounds, clashing default notes, and the confirm-exit overlay tangled when modeled as transitions (an overlay is a control-flow interrupt, not a state). The Update key-routing cascade (flowchart, blue-`root`/purple-`delegate` leaf split) is clearer and is what shipped. A clean correct diagram beats a chaotic more-correct one.
10. **ELK `nodePlacementStrategy: NETWORK_SIMPLEX` is the default for cyclic/tall flows (2026-06-13).** logging-errors, spatial, world-movement, and hooks use it — it aligns ranks into clean columns and shortens edges. `cycleBreakingStrategy` was *not* needed (hooks, with two cycles, rendered clean without it). Two Mermaid gotchas surfaced: `;` is a statement separator and breaks any bare diagram text (sequence Note text especially); spatial diagrams clip rather than wrap overlong labels, so labels must fit the *narrowest* cell.

**Quality Review (2026-06-13):** Full-corpus pre-merge review (`documentation-overhaul-review.md`, CLAUDE.md §8) surfaced 13 findings across the ~30-doc corpus; all applied in one pass.

1. **Spatial slot purpose resolved → faction-utilities / edit surface (finding M2).** Three sources disagreed on the disabled third TUI mode: `interface/overview.md` framed it as an "Edit Mode" direction, while the source placeholder and `interface/state-machine.md` both said "reserved for F-012 (Spatial Map CLI)" — yet F.012 had already shipped as a standalone CLI and map *rendering* is F.005.2 inside Manage, so the slot pointed at a shipped/relocated feature. Robert's call: the slot is a faction-utilities / edit surface. Dropped the F-012 / Spatial-Map framing from the source placeholder (`internal/faction/tui/view.go:53`) and `state-machine.md`; `overview.md` was already canonical. Subsumes finding L5 (the hyphenated `F-012` string is gone). **Durable — candidate to graduate to an interface-page Key Decision at pre-merge.**
2. **Rules-doc path fixed corpus-wide and at the seed (finding H1).** Dead `../../swn-faction-mechanics.md` links in `actions.md`/`effect-mutation.md`/`goals.md` repointed to `../../rules/swn-faction-mechanics.md`; the wrong path was also fixed in its source — `process/templates/arch-page.md` (stops recurrence on the next data-backed page) and `CLAUDE.md`'s routing table. Robert's call: wide scope, one pass.
3. **Goal-lock prose reconciled to the late reorder (finding H2).** `goals.md` `CheckLock` framing ("start of turn… before bookkeeping") corrected to "the goal-lock phase, after bookkeeping and movement, before action selection," matching the `orchestrator.go` phase reorder already reflected in `engine/overview.md` and `orchestrator.md`. **Same durable reorder as the orchestrator.md "Goal-lock sits adjacent to the phase it gates" Key Decision** — graduate together if at all.
4. **StreamClosedMsg reattributed UI-side in the event-stream diagram (finding L7).** Robert's call: `Eng--)UI: StreamClosedMsg` → a `UI->>UI` self-message, since `ObserverPump` generates it from the drained-and-closed channel; matches the page's prose and the existing pump self-message convention (line 71).
5. **Mechanical corrections bundled (findings M1, M3, L1–L4, L6).** Hook-category count four→all five; `state-machine.md` engine-field claim corrected (the root forwards `*engine.Engine` to the turn view and never stores it; it holds `*state.FactionState`); action-count / collaborator-count / diagram-count / template-count nits; `logging-errors.md` mermaid frontmatter de-indented to the kit's flush-left style. Finding L8 left as-is (bare provenance filename, not a link). **Out-of-scope aside, not actioned:** a second hyphenated `F-012` at `cmd/gm-toolkit/faction.go:69` and the stale `RunFactionTurn` doc-comment at `orchestrator.go:55` are code-side, deferred to the `orchestrator.go` commit at Robert's discretion.

## Effort Summary

| # | Effort | Track | Scope | Depends on |
|---|--------|-------|-------|------------|
| 1 | Dev Journal & Decisions-Log Consolidation | meta | Freeze the log; convert the journal to a north-star index; evict mutable content | — (final tail rides Effort 3) |
| 2 | Process-Docs Alignment | generative | Update `session-modes.md` + CLAUDE.md to match real process; codify push-not-pull, the freeze rule, and a docs-initiative plan format | Effort 1 |
| 3 | Architecture Overviews | generative | Build `docs/architecture/{engine,interface}/` — thin spines + seeded per-subsystem wiki pages | Effort 1 (receives migrated narrative) |
| 4 | Contributor Guides | generative | `docs/contributing/` — front door + two per-side guides of extension recipes (Plan: `…-effort-4-plan.md`) | Effort 2 |
| 5 | Architecture Diagrams | generative | Mermaid diagrams (data/logic flow), embedded in arch pages + standalone where useful | Effort 3 |

Large or design-heavy efforts (Effort 3, and Effort 4 after its Decision-6 reversal) get their own lightweight Plan session to slice into execution sessions; Effort 1 goes straight to execution from this plan.

## Per-Effort Breakdown

### Effort 1 — Dev Journal & Decisions-Log Consolidation

**Scope:** Rationalize `docs/dev_journals/faction-manager/`. Touches dev-journal *artifacts* only; the process *rules* that describe these changes are Effort 2.

**Deliverables (non-blocked, done now):**
- Freeze `decisions-log.md` (archival mechanism per open question). Stop future appends.
- Convert `dev-journal-factions.md` into the north-star index: keep Quick Reference (links) + Design Principles; remove Feature List, Open Questions, and the Modes narrative.
- Migrate evicted content: Feature List + Open Questions → `planned-work.md`.

**Deferred tail (rides Effort 3):** migrate the Modes / design narrative into the engine Architecture Overview, then confirm it is stripped from the journal.

**Dependencies:** none to start; the tail needs Effort 3's overview to exist as a destination.

### Effort 2 — Process-Docs Alignment

**Scope:** Make the process docs describe the process we actually run, and codify the decisions ratified here.

**Deliverables:**
- `session-modes.md`: ensure Arc flow and effort-based flow are accurate; add/clarify the multi-effort **docs** flow; amend §9 (decisions-log scope) to the freeze rule + push-not-pull system (arch-page `Key Decisions` as read surface, mandatory Discovery read of the relevant arch page, pre-merge promotion step).
- `CLAUDE.md`: amend rule 7.1 (decisions log → plan-doc provenance + promotion); update the routing table (add `architecture/` wiki and `contributing/`); make the `// deliberate:` fence-sign carve-out explicit under Code Style.
- New template(s): a docs-initiative plan format (the lighter format this very doc improvised — it is the worked example), under `docs/process/templates/`.
- `style-guide.md`: update for the new doc-structure conventions if needed.

**Dependencies:** Effort 1 (the artifact decisions exist before the rules describing them are written). Gates Effort 4.

### Effort 3 — Architecture Overviews

**Scope:** Build the `docs/architecture/` tree: thin spines + seeded per-subsystem wiki pages, per Decision 5.

**Deliverables:**
- `docs/architecture/engine/overview.md` (spine + index) and seeded engine subsystem pages (candidate set: turn pipeline, orchestrator, actions, goals, hooks, spatial, movement, persistence — finalized in this effort's Plan).
- `docs/architecture/interface/overview.md` (spine + index) and seeded interface subsystem pages (candidate set: regions, overlays, state machine, event stream — finalized in this effort's Plan).
- The per-page template (Purpose / Shape / Key Decisions / Dependencies).
- Absorb Effort 1's deferred narrative (Modes, design principles) into the engine overview.

**Dependencies:** receives Effort 1's deferred content; feeds Effort 5. Largest effort — its Plan slices it into execution sessions (spine first, then page clusters).

### Effort 4 — Contributor Guides

**Scope:** How to extend each side of the system, organized around per-subsystem extension recipes. Split engine vs interface, over a shared setup/build front door. The recipe shape, coverage, and slicing are settled in `documentation-overhaul-effort-4-plan.md`.

**Deliverables:**
- `docs/contributing/overview.md` — front door: setup, build, run, and the shared extension pattern.
- `docs/contributing/engine.md` — engine test tooling + conventions + six extension recipes.
- `docs/contributing/interface.md` — interface test tooling + conventions + three extension recipes.

**Dependencies:** Effort 2 (the workflow the guides describe). Cross-links Effort 3 arch pages (each recipe links its subsystem's page). Gets its own Plan session (Decision Record, Effort 4).

### Effort 5 — Architecture Diagrams

**Scope:** Mermaid diagrams walking through data and logic flow, embedded into arch pages and standalone where a diagram earns its own surface.

**Deliverables:**
- Diagrams embedded in the spines and key subsystem pages (candidates: turn pipeline flow, orchestrator/sub-engine dispatch, data flow, TUI state machine).
- `docs/architecture/diagrams/` for any standalone large diagrams.

**Dependencies:** Effort 3 (the pages the diagrams embed into must exist).

## Sequencing / DAG

```
Effort 1 ──┬──► Effort 2 ──► Effort 4
           └──► Effort 3 ──► Effort 5
                    ▲
            (Effort 1 tail migration
             lands during Effort 3)
```

- **Effort 1 is first** and unblocks everything.
- **Efforts 2 and 3 can run concurrently** after Effort 1 (they touch disjoint files); only Effort 1's tail migration links them.
- **Effort 4** waits on Effort 2. **Effort 5** waits on Effort 3.
- Recommended linear order if not parallelizing: 1 → 2 → 3 → 4 → 5.

## Shared Context

**Target `docs/` layout (additions):**

```
docs/
  architecture/
    engine/    overview.md + per-subsystem pages
    interface/ overview.md + per-subsystem pages
    diagrams/  standalone Mermaid (as needed)
  contributing/
    engine.md
    interface.md
  process/
    templates/ (+ docs-initiative plan format)
  dev_journals/faction-manager/
    dev-journal-factions.md  (north-star index, append-only)
    planned-work.md          (workhorse; absorbs evicted status/questions)
    decisions-log.md         (frozen archive)
```

**Adapted work breakdown:** this plan uses a per-effort breakdown rather than the code-oriented Phase→Commit→Task structure in `implementation-plan.md`. Generative docs discover-at-write, so per-file edits cannot be pre-specified. This mismatch is itself the motivation for the docs-initiative plan format produced in Effort 2.

**Model discipline:** this Plan session is Opus. Per-effort execution: suggest **Opus** for design-heavy efforts (2, 3, 5 — structure and rationale decisions); **Sonnet** was used for the mechanical Effort 1. **Effort 4 runs entirely on Opus** (both its Plan session and its execution sessions) — Robert does not trust Sonnet for documentation work; this overrides the table's Sonnet-for-docs-execution default for the generative efforts. Effort 3's own Plan session runs on **Fable** — one-shot taxonomy decision (Decision Record, Effort 2 entry 10). Surface and confirm the model at each transition.

## Out of Scope

- **Codex-tool documentation** — the Codex tool is not yet built (see `project-plan.md`); its docs come with that tool.
- **Exhaustive subsystem-page stubbing** — pages are seeded and grown, not papered up front (Decision 5).
- **Static-site generator (MkDocs/mdBook/Docusaurus)** — explicit non-goal for this initiative. Source stays portable markdown; a rendered doc-site is a possible future once the corpus matures, registered separately if pursued.
- **Tool-specific doc platforms (Obsidian vault, Notion, Confluence)** — rejected (Decision 7).
- **Rewriting `swn-faction-mechanics.md`** — rules reference doc, untouched unless a cross-link needs fixing.

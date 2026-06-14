# Session Modes & Initiative Flows

## 1 — Purpose

This document describes how work moves through the GM Toolkit project, from "this might be worth doing" to "shipped." It defines the **modes** of work (Discovery, Plan, Execution), the **artifacts** they produce (Concept, Discovery doc, Plan), the **flows** that combine them per work type (arc, feature, refactor, bugfix, docs, chore), the **effort** mechanism for splitting large initiatives, and the **discipline** that keeps each mode honest.

**Read this before:**
- Starting a new initiative — to confirm which flow applies
- Promoting an item from `planned-work.md` to Up-Next or Current
- Proposing a change to the process itself

**Audience:** Robert and Claude, working together. This is the substantive guide; `CLAUDE.md` carries only a pointer and the routing table.

**Why this doc exists:** Our process was originally designed for feature work. A retrospective surfaced that refactors fail silently under the same flow — refactor artifacts describe existing code (which drifts), where feature artifacts invent new structure (which is authoritative). This doc makes the distinction explicit and gives each work type its own flow.

## 2 — The Initiative Lifecycle

Every initiative passes through the same stages, regardless of type. All stages are tracked in `planned-work.md` until the initiative ships.

### Stage 1 — Backlog

Lands in the **Backlog** section of `planned-work.md`, grouped into per-type subsections (Features, Refactors, Bugfixes — others added as needed). Each table has columns:

- **ID** — stable identifier; the prefix encodes the type: `A.` arc, `F.` feature, `R.` refactor, `B.` bugfix, `D.` docs, `C.` chore
- **Item** — short title
- **Trigger** — the condition that warrants promotion ("when X is done", "when Y becomes blocking")
- **Detail** — one-line context, problem, or rationale

The ID prefix determines which flow applies when the item is eventually worked (see section 3). Backlog items are unordered. They sit until their trigger fires.

### Stage 2 — Up Next

Promoted into the **Up Next** table when the trigger condition is close or already met. Items move (not copied) from Backlog — Up Next is the prioritized queue, not a projection. Single table; columns ID / Item / Type / Trigger / Detail.

Promotion from Backlog to Up Next is Robert's explicit call.

### Stage 3 — Current Initiative

Promoted into the **Current Initiatives** table at the top of `planned-work.md` and given a full write-up further down in the same file. The table is the index; each row links to the write-up.

The write-up is the initiative's **Concept** artifact — the basic idea, recorded before any Discovery happens. It adds:

- **Problem** — what's broken or missing
- **Approach** — how we plan to address it
- **Unlocks** — what becomes possible once it lands
- **Trigger** — copied from the Backlog/Up Next entry
- **Status** — see below

**Arcs are the exception:** an arc's Concept is a standalone doc at `docs/initiatives/arcs/<arc-name>/<arc-name>-arc-concept.md` (template in `templates/`), because it outlives the planned-work entry and anchors the arc's folder. The arc's planned-work write-up reduces to a pointer at that doc.

**Status values:**

- `In-Progress` — actively being worked. Colloquially called "Current." Capped at 2–3 at any time; more than that is wishful thinking.
- `Blocked` — cannot proceed until something else lands. **Required `by:` note** in the write-up naming the blocker (another initiative, an external dep, a design decision).

Promotion from Up Next to Current Initiative is Robert's explicit call. The write-up (or arc concept doc) is drafted at promotion time, not in advance.

**Arc rows:** when an arc is promoted, its row stays in Current Initiatives for the arc's whole duration — the queue should show the arc is live. Its constituent initiatives get their own rows (and write-ups) as they are promoted, referencing the arc plan. The arc row is removed when the final initiative ships.

### Stage 4 — Done

When all execution phases land and the branch is merged:

1. Remove the row from the Current Initiatives table and the write-up from `planned-work.md`
2. Tag the version bump if warranted (per CLAUDE.md)

### Minor items

Bugfix (`B.`) and chore (`C.`) items may skip Stages 2 and 3 entirely — they go straight from Backlog to a single execution session and are removed when done. See section 6 for the bugfix flow and section 7 for docs/chore.

### Promotion is manual

Movement between stages is always Robert's explicit call, not automatic. Claude may *suggest* a promotion ("this feels ready for Up Next") but does not execute it without confirmation.

## 3 — Routing by Type

The type of a Backlog entry (encoded in its ID prefix) determines which flow the work will run. Set it when the item lands in Backlog; revisit only if the nature of the work changes (e.g. a "bugfix" turns out to need a redesign and becomes a "refactor").

| Type | Flow | When to use |
|------|------|-------------|
| `arc` | Arc-Level Flow (section 9) | A collection of conceptually connected initiatives contributing to a greater whole. Cross-cutting decisions ratified once bind every initiative in it. |
| `feature` | Feature Flow (section 4) | New capability, new domain concept, new user-visible behavior. Default when in doubt. |
| `refactor` | Refactor Flow (section 5) | Restructuring existing code without changing behavior. Pattern conversions, package reorganization, API reshaping. |
| `bugfix` | Bugfix Flow (section 6) | Existing behavior is wrong. Tiered by severity (see section 6). |
| `docs` | Docs & Chore Flow (section 7) | Documentation work — new docs, restructures, doc-only refactors of existing content. Tiered: single-session or multi-effort (see section 7). |
| `chore` | Docs & Chore Flow (section 7) | Mechanical maintenance — dependency bumps, tooling config, cleanup. No design decisions. |

**Mis-routing is recoverable.** If an initiative is partway through one flow and turns out to be a different type, change the ID prefix, move to the appropriate flow, and add a note in `planned-work.md` explaining the reroute. The flows aren't sealed — they're defaults that match common shapes of work.

### Types × artifacts at a glance

Every flow is a different weighting of the same three artifacts (Concept → Discovery → Plan) followed by Execution. This matrix is the skeleton; the per-type sections below carry the detail.

| Type | Concept | Discovery | Plan | Execution |
|------|---------|-----------|------|-----------|
| **Arc** | Standalone doc in `arcs/<arc>/` | Arc-Discovery, always | Arc-Plan (initiative slicing + DAG) | N initiatives, each running its own flow |
| **Feature** | planned-work write-up | Always (1–2 sessions) | Overview plan; per-effort plans as needed (section 8) | One session per commit |
| **Refactor** | planned-work write-up | Often skipped; audit-first when it fires | Pattern + unit list | One session per unit, mandatory re-grounding |
| **Bugfix** | Backlog row only | Never (diagnosis = first half of Plan) | Skipped (trivial) or lightweight (non-trivial) | Single session |
| **Docs** | Backlog row, or write-up if initiative-sized | Skipped — generative docs discover-at-write | Skipped (single-session) or docs-initiative plan (multi-effort) | One session per effort/doc |
| **Chore** | Backlog row only | Never | Never | Single session |

## 4 — Feature Flow

Feature work follows three modes in sequence: **Discovery → Plan → Execution.** Each mode is its own session (or several). Each produces a specific artifact. Boundaries between modes are strict — see section 10 for the discipline.

### Discovery (1–2 sessions)

**Goal:** understand the problem space, survey relevant existing code, identify approach options, surface tradeoffs and open questions.

**Artifact:** `docs/initiatives/discovery/<initiative-name>-discovery.md` — see template at `docs/process/templates/discovery.md` for required and optional sections.

**Session ends when:** the discovery doc is written and Robert has signed off. Do not begin planning in the same session.

### Plan (1–2 sessions)

**Goal:** turn the discovery doc into an executable plan — work broken into commits, each commit sized for one execution session.

**Artifacts** — see template at `docs/process/templates/implementation-plan.md`:

- **Top-level plan** — `docs/initiatives/implementation/<initiative-name>-plan.md`. Always required.
- **Per-effort plans** — `docs/initiatives/implementation/<initiative-name>-effort-N-plan.md`. Used when the initiative splits into efforts — see section 8 for when to split and how effort planning is paced.

**Session ends when:** the plan is written and Robert has signed off. Do not begin execution in the same session.

### Execution (N sessions, one per commit)

**Goal:** implement one commit's worth of work per session.

**Activities:**
- Read the implementation plan, focused on the current commit's tasks
- Read the relevant current code (it may have shifted since the plan was written)
- Implement the tasks
- Commit using conventional commit format (per CLAUDE.md)
- If the work surfaces a scope change, pause and discuss before proceeding

**Artifact:** code, plus a `wip:` commit per the git-commit cadence (squashed to one conventional commit on merge).

**Session ends when:** the commit lands. The next commit in the plan is the next session.

### Pre-Merge Checklist

Before merging the feature branch, run the checklist in CLAUDE.md (Collaboration, item 8):

1. Senior-engineer code review of the branch
2. Promote durable decisions: graduate architecturally-significant entries from the plan's Decision Record into the relevant architecture page's `Key Decisions` section (see section 10, Decision records)
3. Update planned-work doc (remove the initiative; capture deferred items that surfaced)
4. Move the initiative's discovery and plan docs to their `completed/` subdirectories

The pre-merge checklist is itself a session (or part of the final execution session) — see section 10 for model selection.

## 5 — Refactor Flow

Refactor work uses the same three modes as feature work, but with significant deltas. The deltas exist because of one structural asymmetry:

> **Feature artifacts invent future structure.** They are *authoritative* — code is changed to match the doc.
>
> **Refactor artifacts describe existing structure.** They are *descriptive* — the doc is a snapshot of code that may have shifted since it was written.

This asymmetry drives every difference below.

### Discovery (0–1 sessions; often skippable)

Refactors usually start with a known target shape — "convert all sub-engines to Shape 2," "migrate Asset.Location to a Location struct," "rename foo to bar." If the target shape is describable in a paragraph, **skip Discovery entirely** and go straight to Plan.

Discovery fires for a refactor only when:
- The target pattern itself needs investigation (e.g. "what should sub-engines look like?")
- An audit is needed before any target can be chosen
- Multiple incompatible target shapes are on the table

When Discovery does fire for a refactor, it's audit-first rather than design-first: inventory the current state, classify divergences from the candidate target(s), then decide on the target. The discovery doc emphasizes *findings* (what's there now) over *invention* (what should be).

### Plan (1 session)

The refactor plan describes the **target pattern** and lists the **units** to convert (one sub-engine, one package, one struct, etc.). It does not write a step-by-step recipe per unit.

Why: recipes drift. Each unit's current state is slightly different and changes as the work progresses. A pattern is stable; a recipe rots.

Phases are scoped smaller than feature phases — typically one unit per phase. This keeps each execution session focused on a single conversion, with the re-grounding step (below) bounded to that unit's code.

### Execution (N sessions, one per unit, with mandatory re-grounding)

Every refactor execution session opens with a **re-grounding step:**

1. Read the current state of the unit being converted
2. Verify it still matches what the plan assumed
3. If it has drifted, update the plan before coding

Skipping re-grounding is the most common refactor failure mode. The plan was written against a snapshot; the snapshot decays; coding against the stale snapshot produces broken or out-of-date conversions.

After re-grounding:
- Apply the target pattern to the unit
- Commit with conventional commit format (typically `refactor:`)
- If the conversion surfaces a structural problem the plan didn't anticipate, pause and discuss before proceeding

### Deferred Refactors

Sometimes a refactor is scoped but not executed immediately — for example, when new code is added during a feature initiative that "should" follow the target pattern but the refactor hasn't shipped yet. In that case, **new code matches the existing (imperfect) pattern, not the target.** Mixed state is worse than uniform pre-refactor state. The refactor will later catch the new code along with everything else.

### Pre-Merge Checklist

Same as Feature Flow — see section 4.

## 6 — Bugfix Flow

Bugfix work is **tiered** — most bugfixes skip Discovery and Plan entirely. The tier is determined by what's known going in.

### Trivial Bugfix (single Execution session)

Use when:
- The root cause is known or trivially diagnosable
- The fix is describable in one sentence
- One file, or a few obviously related files
- One reasonable fix approach

Flow: **Backlog → Execution → Done.** No discovery, no plan. Single session: diagnose (if needed), fix, commit (`fix: <description>`), update tests.

Example: the `BuyAsset` stealth ID hardcode (item 1 in current `Deferred — Minor`). Known cause, known fix, single file.

### Non-Trivial Bugfix (lightweight plan + Execution)

Use when:
- Root cause needs investigation before a fix can be chosen
- Multiple plausible fix approaches with meaningful tradeoffs
- The fix touches several files or crosses package boundaries
- Risk of regression is non-trivial

Flow: **Backlog → Plan → Execution → Done.**

The plan is **lightweight** — typically a single section in `docs/initiatives/implementation/<bugfix-name>-plan.md` covering:

- **Root cause** — what's actually wrong
- **Fix approach** — the chosen approach, with brief rationale if alternatives existed
- **Test plan** — how the fix is verified, and what regression coverage is added

No separate Discovery session. If diagnosis is the hard part, that's the first half of the Plan session, not its own mode.

Execution is typically a single session, occasionally split if the fix is genuinely multi-commit.

### Promoting a bugfix to a refactor or feature

If investigation reveals the bug is symptomatic of a structural problem, the right fix may not be a bugfix at all. In that case: stop, change the `Type` field on the planned-work entry from `bugfix` to `refactor` or `feature`, and re-route through the appropriate flow. Mis-routing is recoverable (see section 3).

### Pre-Merge Checklist

For trivial bugfixes the pre-merge checklist may be lightweight or skipped (single-commit fixes don't need a senior-engineer code review pass). For non-trivial bugfixes, follow the full checklist as in Feature Flow.

## 7 — Docs & Chore Flow

### Docs

Documentation work — new docs, restructures, or doc-only refactors of existing content. Conventional commit prefix: `docs:`. The flow is **tiered**; a docs initiative never reroutes to `feature` or `refactor` just because it grows — it stays `docs` and scales up within this flow.

**Single-session docs (default).** Flow: **Backlog → Execution → Done.** No Discovery, no Plan. Most docs work lands here.

**Multi-effort docs initiative.** Use when the work has structural decisions about the doc set and clearly exceeds one session (worked example: the Documentation Overhaul, `documentation-overhaul-plan.md`). Flow: **Backlog → Concept write-up → Plan → Execution per effort → Done.**

- **No Discovery doc.** Generative docs *discover-at-write* — the act of writing an architecture page is the act of exploring that subsystem, so per-doc design cannot be usefully pre-specified. Design rationale is worked out in the Plan session and recorded in the plan's "Decisions Ratified in Planning."
- **The plan uses the docs-initiative format** (`templates/docs-plan.md`): a per-effort breakdown (Scope / Deliverables / Dependencies) plus a sequencing DAG, instead of the code-oriented Phase → Commit → Task structure — per-file edits can't be pre-specified for discover-at-write work.
- **Effort pacing follows section 8:** small efforts execute straight from the overview plan; large or design-heavy efforts get their own lightweight Plan session.

### Chore

Mechanical maintenance — dependency bumps, tooling config, formatting passes, cleanup. No design decisions. Single execution session. Flow: **Backlog → Execution → Done.** Conventional commit prefix: `chore:`.

If a chore turns out to require design (e.g. "bump this dep" reveals an API change that needs a refactor), stop and reroute, same as bugfix promotion (see section 6).

### Pre-Merge Checklist

Minimal for single-session docs and chores: no senior-engineer code review pass for docs-only changes; remove the entry from `planned-work.md`. Multi-effort docs initiatives run the standard checklist (section 4), minus the code review when no code changed.

## 8 — Efforts

An **effort** is the scaling mechanism for initiatives too large to be contained in one coherent Plan. It is not a work type — any type can split into efforts when its plan exceeds what one plan can coherently hold (features, refactors, and docs initiatives have all done it: `asset-movement-redesign`, `logging-and-errors`, the Documentation Overhaul).

### The multi-effort flow

**Discovery → Overview Plan → per effort: (Plan as needed) → Execution.**

- **One Discovery, shared.** The initiative's discovery doc designs across all efforts; efforts do not get their own Discovery.
- **One overview plan, always.** The first Plan session produces the top-level plan (`<initiative-name>-plan.md`): shared decisions, the effort slicing, and the dependency DAG between efforts. Small efforts are specified fully here and execute straight from it.
- **Per-effort Plan sessions, as needed.** A large or design-heavy effort gets its own lightweight Plan session producing `<initiative-name>-effort-N-plan.md`. The overview plan names which efforts need one. Each per-effort Plan session is its own session, subject to the usual mode boundaries.
- **Execution per effort.** Each effort's commits run as normal execution sessions for the initiative's type (per-commit for features, per-unit for refactors, per-doc for docs).

### Interim Effort Review (optional)

An optional code review can run at an effort boundary — after an effort's commits land, before the next effort's Plan session. It is a senior-engineer review of just that effort's work (the diff since the effort began), catching structural issues while the next effort is still cheap to redirect.

It is **distinct from the Pre-Merge Checklist**: the checklist still runs once at branch level after the final effort. An interim review does not trigger a merge or a planned-work update — it is review only. Like all reviews, it runs on Opus (section 10). It is Robert's call whether a given effort boundary warrants one.

### Effort vs. arc

If the slices are independently shippable initiatives (each with its own branch, merge, and lifecycle entry), it's an arc (section 9), not a multi-effort initiative. Efforts share one branch, one Discovery, and one pre-merge checklist; arc initiatives don't.

## 9 — Arc-Level Flow

Some work spans multiple initiatives that share a common architecture — for example, a TUI rebuild that ships as a series of layered initiatives, each adding features against the same engine boundary. These are **arcs.** An arc is a Type in the tracker (`A.` prefix, see sections 2–3), but unlike the other types it is a *container*: it ships nothing directly. The standard Discovery and Plan modes apply at two levels — once for the arc, once for each initiative within it.

**When to use:**
- The work clearly requires multiple shippable initiatives (not a single multi-phase initiative).
- Cross-cutting architectural decisions need to be ratified once and bind every initiative in the arc.
- The order in which the initiatives ship is itself a design question worth deliberation.

If the work fits in one initiative — even a multi-effort one — use the Feature or Refactor Flow directly. Arcs add overhead that only pays off when cross-initiative coordination is real.

**Arc artifacts live together:** every arc-level doc sits in `docs/initiatives/arcs/<arc-name>/` — the concept, the arc-discovery, and the arc-plan. Per-initiative artifacts stay in the standard `discovery/` and `implementation/` locations.

### Arc-Concept

The arc's Concept is a standalone doc, `arcs/<arc-name>/<arc-name>-arc-concept.md` (template: `templates/arc-concept.md`), drafted when the arc is promoted to Current Initiatives (section 2). It captures the vision, why the work is an arc rather than one initiative, and the rough candidate initiatives — before any Discovery.

### Arc-Discovery (1–2 sessions)

**Goal:** ratify the cross-cutting architectural decisions that will constrain *every* initiative in the arc — framework, boundaries, conventions, extensibility patterns. Do not slice into initiatives yet; that is Arc-Plan's job.

**Artifact:** `docs/initiatives/arcs/<arc-name>/<arc-name>-arc-discovery.md`. Structurally similar to a per-initiative discovery doc, but its decisions apply to the whole arc, and its Open Questions may be routed to specific initiatives later (in Arc-Plan).

**Session ends when:** the arc-discovery doc is written and Robert has signed off. Do not begin Arc-Plan in the same session.

### Arc-Plan (1 session)

**Goal:** define the initiatives that make up the arc, in what order, with what boundaries.

**Artifact:** `docs/initiatives/arcs/<arc-name>/<arc-name>-arc-plan.md`. Contains:

- **Initiative list** — ordered, with dependencies surfaced (a DAG, not necessarily a flat sequence — some entries can parallelize).
- **Per-initiative boundary** — what is in scope for each initiative, what is punted to a later entry.
- **Open Question routing** — each open question from Arc-Discovery assigned to a specific initiative.
- **Cross-arc dependencies** — engine-side prerequisites, data-layer prereqs, anything outside the arc that gates an initiative.
- **Out-of-arc deferrals** — explicit list of work the arc does not address.

**Session ends when:** the arc-plan is written and Robert has signed off. Each initiative in the arc is added to `planned-work.md` as a standard Backlog or Up-Next entry that references the arc-plan.

### Initiatives within the arc

Each initiative listed in the Arc-Plan goes through the standard lifecycle from section 2 and the per-Type flow from section 3.

The arc-level artifacts make per-initiative Discovery and Plan *lighter*, not absent. Per-initiative Discovery does not re-litigate decisions already ratified at the arc level — it focuses on initiative-specific decisions (scope, internal slicing, the open questions Arc-Plan routed to this initiative). Per-initiative Plan inherits the arc's boundary for the initiative.

### Arc completion

The arc is complete when its final initiative ships. There is no separate arc pre-merge checklist — each initiative runs its own per section 4 or 5. The arc's row leaves Current Initiatives at that point (section 2). Arc artifacts stay in `arcs/<arc-name>/` — the folder is the arc's permanent home; nothing moves to `completed/`.

## 10 — Session Discipline

The flows in sections 4–9 work because of a small set of cross-cutting rules. These rules apply to every session regardless of which flow it serves.

### One mode per session

Discovery, Plan, and Execution are separate sessions. A session does not change modes mid-flight. Even when the inputs to the next mode are ready, the next mode starts in a new session.

The most-violated boundary is **Plan → Execution.** When the current session's deliverable is an implementation plan, stop after the plan is written, approved, or handed off. Do not write code, edit source files, or stage edits. If pushed to keep going, suggest opening a new session.

This applies to refactor and bugfix Plan sessions as well, even when the plan is short.

### Plan mode is a strong signal

When plan mode is active and the stated deliverable is a file (a discovery doc or implementation plan), **that file is the entire session output.** Exiting plan mode is not a handoff into implementation — it is the end of the session. Do not queue file edits, Bash commands, or follow-up work to run after exit. End the turn after the file is written.

### Re-grounding before action

Every Plan session and every Execution session must verify its inputs against the current state of the code before acting:

- **Discovery and Plan sessions** — read the relevant architecture page(s) under `docs/architecture/` before designing (start at `docs/architecture/architecture-overview.md`); their `Key Decisions` sections are binding context. This is the push half of the push-not-pull decision system (see Decision records below).
- **Plan sessions** — re-read the relevant code. The discovery doc is a snapshot; the snapshot may have decayed. Verify the plan's assumptions against the current source, not against what the discovery doc said.
- **Execution sessions** — re-read the current code for the unit being touched. The plan was written against a snapshot that may have shifted, especially for refactors where prior phases have already changed the surrounding code. `// deliberate:` fence-signs encountered in that code are binding unless explicitly revisited with Robert.

Re-grounding is most load-bearing for refactors (see section 5) but applies to all flows. Reasoning from a stale doc instead of the current code is the most common drift failure.

### Model selection

| Session type | Model | Notes |
|--------------|-------|-------|
| Arc-Discovery / Arc-Plan / process redesign | Fable | Structure-setting sessions whose output constrains many future sessions (taxonomy, cross-cutting architecture, the process itself); most initiatives never need it |
| Discovery | Opus | Always |
| Plan | Opus | Always |
| Execution | Sonnet | Default; suggest Opus when the phase involves heavy design or unusual complexity |
| Code review (interim or pre-merge) | Opus | Always — critical design review benefits from Opus reasoning |
| End-of-phase docs and pre-merge checklist | Sonnet | Doc/journal/planned-work updates don't need Opus; the code-review step within the checklist uses Opus |

Surface the recommended model and prompt Robert to switch (`/model`) before each mode transition, including mid-session shifts (e.g. moving from execution into the pre-merge checklist).

### Promotion is manual

Movement between lifecycle stages (Backlog → Up Next → Current Initiative → Done) is always Robert's explicit call. Claude may suggest a promotion but does not execute it without confirmation. See section 2.

### File locations

- **Discovery docs** — `docs/initiatives/discovery/<initiative-name>-discovery.md`
- **Implementation plans** — `docs/initiatives/implementation/<initiative-name>-plan.md`, plus per-effort plans at `docs/initiatives/implementation/<initiative-name>-effort-N-plan.md` when needed
- **Arc artifacts** — `docs/initiatives/arcs/<arc-name>/` holds the arc concept, arc-discovery, and arc-plan; this is the arc's permanent home
- **Completed artifacts** — moved to `docs/initiatives/discovery/completed/` and `docs/initiatives/implementation/completed/` when the initiative ships (arc artifacts stay put)

Plans never live in `~/.claude/plans/` or other harness defaults — they are repo-tracked artifacts.

### Decision records

Decisions are recorded where they are made and *read* where they cannot be skipped — push, not pull. The old branch/time-organized decisions log failed because nobody retrieves on that axis; it is now a frozen historical archive (`docs/dev_journals/faction-manager/archive/decisions-log.md`, 294 decisions, never appended).

**Where decisions are written:**

- Discovery decisions live in the discovery doc
- Plan decisions live in the plan's "Decisions Ratified in Planning" section
- Execution-time decisions (and reversals of planned decisions) land in the plan's **Decision Record** section

**Where decisions are read (the push surfaces):**

1. **Architecture pages (primary).** Each subsystem page under `docs/architecture/` carries a `Key Decisions` section — the durable "why this subsystem is shaped this way," organized by the axis people actually query on. Reading the relevant page is a required re-grounding step (above), so the rationale sits in the doc every Discovery is forced to open.
2. **Code fence-signs.** A one-line `// deliberate: X, not Y, because <reason>` at the fence itself, surfaced during execution and refactor re-grounding (convention defined in CLAUDE.md, Code Style).
3. **Completed plan docs (provenance).** Each initiative's plan ships to `completed/` carrying its Decision Record — source material, not a routine read surface.

**Promotion.** The pre-merge checklist (section 4, step 2) graduates architecturally-durable decisions from the plan's Decision Record into the relevant architecture page's `Key Decisions`. Micro-decisions (naming, a test strategy) stay behind as plan-doc provenance or fence-signs — only architecturally-significant "why" graduates.

**Mutability contracts.** Two artifacts in the dev-journal folder have opposite contracts, deliberately:

- **Plan Decision Record** — *mutable*: append **and reconcile**. When execution overrides a planned decision, log the reversal with rationale and fix the plan body so it doesn't lie.
- **Dev-journal north-star** (`dev-journal-factions.md`) — *append-only*: never edited, only added to.

And two indexes sit on different axes, kept distinct and cross-linked: the dev-journal north-star indexes *process/tracking* artifacts; the architecture spine indexes *system* pages.

### Deferred refactor — match existing pattern

When a refactor is scoped but not yet executed, new code added in the meantime matches the existing (imperfect) pattern, not the target. Mixed state is worse than uniform pre-refactor state. See section 5.

## 11 — Templates

Templates for the file artifacts produced by Concept, Discovery, and Plan sessions:

- **Discovery doc** — [`templates/discovery.md`](./templates/discovery.md). Copy to `docs/initiatives/discovery/<initiative-name>-discovery.md` and fill in.
- **Implementation plan** — [`templates/implementation-plan.md`](./templates/implementation-plan.md). Copy to `docs/initiatives/implementation/<initiative-name>-plan.md` (or split into per-effort files for large initiatives) and fill in.
- **Docs-initiative plan** — [`templates/docs-plan.md`](./templates/docs-plan.md). For multi-effort `docs` initiatives (section 7). Copy to `docs/initiatives/implementation/<initiative-name>-plan.md` and fill in.
- **Arc concept** — [`templates/arc-concept.md`](./templates/arc-concept.md). Copy to `docs/initiatives/arcs/<arc-name>/<arc-name>-arc-concept.md` when an arc is promoted.
- **Architecture page** — [`templates/arch-page.md`](./templates/arch-page.md). Code-map header + Purpose / Shape / Key Decisions / Dependencies. Copy to `docs/architecture/<side>/<subsystem>.md` (`<side>` = `engine` or `interface`) and fill in. Landed and validated against the orchestrator pilot in D-002 Effort 3.

All templates carry inline guidance (HTML comments) explaining required sections, optional sections, and how to scale the structure to the size of the work. Required sections are marked `REQUIRED`; everything else is included only when relevant.

Templates evolve as we learn what artifacts actually need. Treat them as living documents — update when a real plan or discovery surfaces a missing section or an unused one.

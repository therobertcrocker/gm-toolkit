# Session Modes & Initiative Flows

## 1 — Purpose

This document describes how work moves through the GM Toolkit project, from "this might be worth doing" to "shipped." It defines the **modes** of work (Discovery, Plan, Execution), the **flows** that combine them (feature, refactor, bugfix, docs/chore), and the **discipline** that keeps each mode honest.

**Read this before:**
- Starting a new initiative — to confirm which flow applies
- Promoting an item from `planned-work.md` to Up-Next or Current
- Proposing a change to the process itself

**Audience:** Robert and Claude, working together. This is the substantive guide; `CLAUDE.md` carries only a pointer and the routing table.

**Why this doc exists:** Our process was originally designed for feature work. A retrospective surfaced that refactors fail silently under the same flow — refactor artifacts describe existing code (which drifts), where feature artifacts invent new structure (which is authoritative). This doc makes the distinction explicit and gives each work type its own flow.

## 2 — The Initiative Lifecycle

Every initiative passes through the same stages, regardless of type. All stages are tracked in `planned-work.md` until the initiative ships.

### Stage 1 — Backlog

Lands in the **Backlog** table in `planned-work.md`. A single flat table with columns:

- **#** — stable identifier
- **Item** — short title
- **Type** — `feature`, `refactor`, `bugfix`, `docs`, or `chore` (determines which flow applies; see section 3)
- **Trigger** — the condition that warrants promotion ("when X is done", "when Y becomes blocking")
- **Detail** — one-line context, problem, or rationale

Backlog items are unordered. They sit until their trigger fires.

### Stage 2 — Up Next

Promoted into the **Up Next** table when the trigger condition is close or already met. Items move (not copied) from Backlog — Up Next is the prioritized queue, not a projection. Same columns as Backlog.

Promotion from Backlog to Up Next is Robert's explicit call.

### Stage 3 — Planned Initiative

Promoted into the **Planned Initiatives** table at the top of `planned-work.md` and given a full write-up further down in the same file. The table is the index; each row links to the write-up.

The write-up adds:

- **Problem** — what's broken or missing
- **Approach** — how we plan to address it
- **Unlocks** — what becomes possible once it lands
- **Trigger** — copied from the Backlog/Up Next entry
- **Status** — see below

**Status values:**

- `In-Progress` — actively being worked. Colloquially called "Current." Capped at 2–3 at any time; more than that is wishful thinking.
- `Blocked` — cannot proceed until something else lands. **Required `by:` note** in the write-up naming the blocker (another initiative, an external dep, a design decision).

Promotion from Up Next to Planned Initiative is Robert's explicit call. The write-up is drafted at promotion time, not in advance.

### Stage 4 — Done

When all execution phases land and the branch is merged:

1. Remove the row from the Planned Initiatives table and the write-up from `planned-work.md`
3. Tag the version bump if warranted (per CLAUDE.md)

### Minor items

Items with `Type: bugfix` or `Type: chore` may skip Stages 2 and 3 entirely — they go straight from Backlog to a single execution session and are removed when done. See section 6 for the bugfix flow and section 7 for docs/chore.

### Promotion is manual

Movement between stages is always Robert's explicit call, not automatic. Claude may *suggest* a promotion ("this feels ready for Up Next") but does not execute it without confirmation.

## 3 — Routing by Type

The `Type` field on a Backlog entry determines which flow the initiative will run when it reaches Stage 3. Set it when the item lands in Backlog; revisit only if the nature of the work changes (e.g. a "bugfix" turns out to need a redesign and becomes a "refactor").

| Type | Flow | When to use |
|------|------|-------------|
| `feature` | Feature Flow (section 4) | New capability, new domain concept, new user-visible behavior. Default when in doubt. |
| `refactor` | Refactor Flow (section 5) | Restructuring existing code without changing behavior. Pattern conversions, package reorganization, API reshaping. |
| `bugfix` | Bugfix Flow (section 6) | Existing behavior is wrong. Tiered by severity (see section 6). |
| `docs` | Docs & Chore Flow (section 7) | Documentation work — new docs, restructures, doc-only refactors of existing content. |
| `chore` | Docs & Chore Flow (section 7) | Mechanical maintenance — dependency bumps, tooling config, cleanup. No design decisions. |

**Mis-routing is recoverable.** If an initiative is partway through one flow and turns out to be a different type, change the `Type` field, move to the appropriate flow, and add a note in the dev journal explaining the reroute. The flows aren't sealed — they're defaults that match common shapes of work.

## 4 — Feature Flow

Feature work follows three modes in sequence: **Discovery → Plan → Execution.** Each mode is its own session (or several). Each produces a specific artifact. Boundaries between modes are strict — see section 8 for the discipline.

### Discovery (1–2 sessions)

**Goal:** understand the problem space, survey relevant existing code, identify approach options, surface tradeoffs and open questions.

**Artifact:** `docs/initiatives/discovery/<initiative-name>-discovery.md` — see template at `docs/process/templates/discovery.md` for required and optional sections.

**Session ends when:** the discovery doc is written and Robert has signed off. Do not begin planning in the same session.

### Plan (1–2 sessions)

**Goal:** turn the discovery doc into an executable plan — work broken into commits, each commit sized for one execution session.

**Artifacts** — see template at `docs/process/templates/implementation-plan.md`:

- **Top-level plan** — `docs/initiatives/implementation/<initiative-name>-plan.md`. Always required.
- **Per-effort plans** — `docs/initiatives/implementation/<initiative-name>-effort-N-plan.md`. Used when the initiative is large enough to split into multiple efforts; small initiatives keep everything in the top-level plan.

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

Before merging the feature branch, run the checklist in CLAUDE.md (Collaboration, item 7):

1. Senior-engineer code review of the branch
2. Update dev journal
3. Update planned-work doc (remove the initiative; capture deferred items that surfaced)

The pre-merge checklist is itself a session (or part of the final execution session) — see section 8 for model selection.

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

Both `docs` and `chore` initiatives skip Discovery and Plan. Flow: **Backlog → Execution → Done.**

### Docs

Documentation work — new docs, restructures, or doc-only refactors of existing content. Single execution session. Conventional commit prefix: `docs:`.

If a docs initiative grows large enough to warrant breaking up (multi-session work, structural decisions about the doc set), promote it to a `feature` or `refactor` and re-route. The `docs` type is for work that fits in a single session.

### Chore

Mechanical maintenance — dependency bumps, tooling config, formatting passes, cleanup. No design decisions. Single execution session. Conventional commit prefix: `chore:`.

If a chore turns out to require design (e.g. "bump this dep" reveals an API change that needs a refactor), stop and reroute, same as bugfix promotion (see section 6).

### Pre-Merge Checklist

Minimal for both. Update the dev journal only if the work warrants a note (most chores don't; most non-trivial docs work does). No senior-engineer code review pass for docs-only changes. Remove the entry from `planned-work.md`.

## 8 — Session Discipline

The flows in sections 4–7 work because of a small set of cross-cutting rules. These rules apply to every session regardless of which flow it serves.

### One mode per session

Discovery, Plan, and Execution are separate sessions. A session does not change modes mid-flight. Even when the inputs to the next mode are ready, the next mode starts in a new session.

The most-violated boundary is **Plan → Execution.** When the current session's deliverable is an implementation plan, stop after the plan is written, approved, or handed off. Do not write code, edit source files, or stage edits. If pushed to keep going, suggest opening a new session.

This applies to refactor and bugfix Plan sessions as well, even when the plan is short.

### Plan mode is a strong signal

When plan mode is active and the stated deliverable is a file (a discovery doc or implementation plan), **that file is the entire session output.** Exiting plan mode is not a handoff into implementation — it is the end of the session. Do not queue file edits, Bash commands, or follow-up work to run after exit. End the turn after the file is written.

### Re-grounding before action

Every Plan session and every Execution session must verify its inputs against the current state of the code before acting:

- **Plan sessions** — re-read the relevant code. The discovery doc is a snapshot; the snapshot may have decayed. Verify the plan's assumptions against the current source, not against what the discovery doc said.
- **Execution sessions** — re-read the current code for the unit being touched. The plan was written against a snapshot that may have shifted, especially for refactors where prior phases have already changed the surrounding code.

Re-grounding is most load-bearing for refactors (see section 5) but applies to all flows. Reasoning from a stale doc instead of the current code is the most common drift failure.

### Model selection

| Session type | Model | Notes |
|--------------|-------|-------|
| Discovery | Opus | Always |
| Plan | Opus | Always |
| Execution | Sonnet | Default; suggest Opus when the phase involves heavy design or unusual complexity |
| End-of-phase docs and pre-merge checklist | Sonnet | Always — doc and checklist work doesn't need Opus |

Surface the recommended model and prompt Robert to switch (`/model`) before each mode transition, including mid-session shifts (e.g. moving from execution into the pre-merge checklist).

### Promotion is manual

Movement between lifecycle stages (Backlog → Up Next → Planned Initiative → Done) is always Robert's explicit call. Claude may suggest a promotion but does not execute it without confirmation. See section 2.

### File locations

- **Discovery docs** — `docs/initiatives/discovery/<initiative-name>-discovery.md`
- **Implementation plans** — `docs/initiatives/implementation/<initiative-name>-plan.md`, plus per-effort plans at `docs/initiatives/implementation/<initiative-name>-effort-N-plan.md` when needed
- **Completed artifacts** — moved to `docs/initiatives/discovery/completed/` and `docs/initiatives/implementation/completed/` when the initiative ships

Plans never live in `~/.claude/plans/` or other harness defaults — they are repo-tracked artifacts.

### Decisions log scope

The decisions log (`docs/dev_journals/faction-manager/decisions-log.md`) tracks **decisions ratified during implementation**, not decisions made during Discovery or Plan.

- Discovery decisions live in the discovery doc
- Plan decisions live in the plan's "Decisions Ratified in Planning" section
- Execution-time decisions (and revisions to earlier decisions) land in the decisions log

Do not back-fill the log with decisions from Discovery or Plan unless they were materially revised during execution.

### Deferred refactor — match existing pattern

When a refactor is scoped but not yet executed, new code added in the meantime matches the existing (imperfect) pattern, not the target. Mixed state is worse than uniform pre-refactor state. See section 5.

## 9 — Templates

Templates for the two file artifacts produced by Discovery and Plan sessions:

- **Discovery doc** — [`templates/discovery.md`](./templates/discovery.md). Copy to `docs/initiatives/discovery/<initiative-name>-discovery.md` and fill in.
- **Implementation plan** — [`templates/implementation-plan.md`](./templates/implementation-plan.md). Copy to `docs/initiatives/implementation/<initiative-name>-plan.md` (or split into per-effort files for large initiatives) and fill in.

Both templates carry inline guidance (HTML comments) explaining required sections, optional sections, and how to scale the structure to the size of the work. Required sections are marked `REQUIRED`; everything else is included only when relevant.

Templates evolve as we learn what artifacts actually need. Treat them as living documents — update when a real plan or discovery surfaces a missing section or an unused one.

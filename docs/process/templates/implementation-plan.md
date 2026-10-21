# [Initiative Name] — Implementation Plan

<!--
COPY THIS FILE TO: docs/initiatives/implementation/<initiative-name>-plan.md

LEVELS (session-modes.md section 4, Work breakdown levels):
  Initiative → Effort → Phase → Commit → Task. One phase = one execution
  session.

This template covers three file shapes:

  1. Single-file plan (one effort): every section below, with phases in
     Work Breakdown. The overview and the effort plan are the same file.

  2. Overview plan (multi-effort): every section above Work Breakdown, plus
     the Effort Summary. Work Breakdown holds one "### Effort N" section per
     effort: its scope, and whether it gets its own Plan session. Effort
     pacing is defined in session-modes.md section 8.

  3. Effort plan (multi-effort):
         docs/initiatives/implementation/<initiative-name>-effort-N-plan.md
     Same sections as a single-file plan, scoped to the effort: its own
     In Plain Words, Decisions Ratified, Decision Record, Shared Context,
     Out of Scope, then phases in Work Breakdown. It cites the overview's
     decisions; it does not repeat them.

SIZE: two budgets, for single-file and effort plans. The overview has none.
  - Front matter (every line above Work Breakdown): about 400. Every phase
    session reads it again. Past it, the effort carries too many decisions:
    split the effort.
  - Each phase: about 800 lines, from the phase heading to the next one.
    Past it, split the phase.
Spend lines on literal code in tasks, not on prose. Commits have no line
budget; split them only at real seams (KEY RULE 1). Past a budget, split, or
record why not in one line under Decisions Ratified in Planning.

  awk '/^## Work Breakdown/{print NR-1, "front matter"; exit}' <plan-file>
  awk '/^### Phase/{if(h)print n,h;h=$0;n=0;next} h{n++} END{print n,h}' <plan-file>

Multi-effort DOCS initiatives use templates/docs-plan.md instead — generative
docs discover-at-write and don't fit the Phase/Commit/Task structure.

Delete these comments before writing.
-->

## In Plain Words

<!--
REQUIRED. Under 15 lines. Plain words for a reader who has not read the rest of
this plan: no project shorthand unless it is defined in the same sentence.

  - What: what this plan builds.
  - Why: the problem it solves, for the GM or for the codebase.
  - How it fits: where it sits in the initiative, the arc and the project.

Every execution session reads this section first (session-modes.md section 10,
Re-grounding). If execution changes the what, why or fit, fix it here too.
-->

## Context / Goal

<!--
REQUIRED. What this plan accomplishes. Brief — link back to the discovery doc for
the full design rationale.

  - Discovery doc: [link]
  - Related plans: [links if applicable]
-->

## Decisions Ratified in Planning

<!--
REQUIRED. Design decisions made *during* the plan session, not in Discovery.

Each decision: what was decided, brief rationale, and which open question (from
Discovery) it resolves if applicable. Architecturally-durable decisions graduate
to the relevant architecture page's Key Decisions at pre-merge.
-->

1. **[Decision]** — [rationale]
2. **[Decision]** — [rationale]

## Decision Record — Execution

<!--
REQUIRED once execution begins (empty at plan time). Execution-time decisions
and reversals of planned decisions land here, logged before the commit they
ride (CLAUDE.md, Collaboration 7.1).

Contract: append AND reconcile — when execution overrides a planned decision,
log the reversal with rationale here and fix the plan body so it doesn't lie.
At pre-merge, architecturally-durable entries graduate to the relevant
architecture page's Key Decisions section; the rest stay here as provenance
when the plan ships to completed/.
-->

---

<!--
OPTIONAL SECTIONS BELOW. Include only the ones relevant to this initiative.
Shared Context is the catch-all for context that doesn't justify its own
top-level section (model discipline, pre-merge additions, naming conventions,
test harness setup, etc.).
-->

## Open Questions — To Ratify at Implementation Time

<!--
Questions deliberately deferred to Execution. Each should name when it will be
resolved (which phase, which commit, or "first use").
-->

## Effort Summary

<!--
Use only when the initiative is split into multiple efforts. One-line scope per
effort in a table, with a link to each per-effort plan file.
-->

| # | Effort | Scope | File |
|---|--------|-------|------|
| 1 | [name] | [one-line scope] | [`-effort-1-plan.md`](./<initiative-name>-effort-1-plan.md) |

## Shared Context

<!--
Cross-cutting context with no single natural home: naming conventions, test
harness setup, model discipline per phase, initiative-specific pre-merge
checklist additions, package conventions.

NOT for per-file implementation detail. Signatures, struct definitions, field
lists, formatter tables, dispatch flows — all belong in the task that builds
them (see Work Breakdown), never here. Hoisting detail up and referencing it
back down splits one spec across two locations; the execution session reading a
task loses locality and has to reassemble the spec. If a section here is named
after a single commit ("Adapter contract changes (Commit 4)"), it is in the
wrong place — move it into that commit's task.
-->

## Out of Scope

<!--
Explicit non-goals. Things that look related but won't be addressed by this plan.
Deferred items belong in planned-work.md as separate Backlog entries.
-->

---

## Work Breakdown

<!--
REQUIRED. The actual sequence of work, structured at whichever granularity fits.

GRANULARITY GUIDE:
  - Small refactor or bugfix:    Steps only.
  - Single-file or effort plan:  Phases → Commits → Tasks. Work that fits one
                                 session is one phase; keep the phase heading.
  - Overview plan:               "### Effort N" sections (file shape 2 above).

KEY RULE 1: Each phase = one execution session. Size each phase to the phase
budget above, typically 2–3 commits. Split phases and commits only at real seams:
a green-build point, a type-system or cross-package boundary, or a behavior
change. Title each commit with its conventional commit message so the execution
session knows the deliverable shape going in.

KEY RULE 2: Every task is self-contained and executable on its own — it IS the
edit, not a description of one (see the per-task format below). The full
implementation detail lives in the task as literal code: real signatures, struct
definitions with field lists, per-kind tables — never a pointer back to a section
above, never pseudocode the executor must flesh out. When a structure is
genuinely shared across commits, the FIRST task that needs it defines it in full;
later tasks cite that task precisely ("uses turn.Model, defined in Commit 5
Task 1") rather than redefining it or referencing a floating section. The only
back-reference allowed is a task-to-task citation.
-->

### Phase 1 — [Title]

<!-- Phase-level goal in 1–2 sentences. One phase = one execution session. -->

#### Commit 1 — `feat: [conventional commit message]`

##### Task 1 — `path/to/file.go`

<!--
A task is the EDIT, not a description of it. The executor (human or Claude)
applies it without interpreting prose back into code and without reading
anything above the task. Pick the format by edit size:

SURGICAL EDIT to an existing file — Find → Replace. Pull the anchor verbatim
from the current source so it locates the spot exactly:

  In `FuncName`, <one line of why this changes>.

  **Find:**
  ```go
  <verbatim current lines, enough to be unique>
  ```
  **Replace with:**
  ```go
  <exact new lines>
  ```

NEW OR REWRITTEN file — full content. No anchor needed; there's nothing to
locate. Give the complete file body in one block:

  `path/to/file.go` (new) — full contents:
  ```go
  <the entire file>
  ```

Either way the code is literal: real signatures, real struct field lists, real
per-kind tables — never "see Shared Context," never pseudocode the executor must
flesh out. If this task defines a structure later tasks reuse, define it in full
here; later tasks cite it by task ("uses turn.Model, defined in Commit N Task M").
-->

##### Task 2 — `path/to/other_file.go`

<!-- ... -->

##### Commit message

<!--
REQUIRED per commit. The fenced conventional-commit message the execution
session will use, so the deliverable shape is fixed going in:

```
type(scope): summary line

- bullet per material change
```
-->

#### Commit 2 — `refactor: [conventional commit message]`

<!-- ... -->

### Phase 2 — [Title]

<!-- ... -->

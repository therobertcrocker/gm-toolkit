# [Initiative Name] — Implementation Plan

<!--
COPY THIS FILE TO: docs/implementation/<initiative-name>-plan.md

This template covers two cases:

  1. Single-file plans for small to medium initiatives — keep everything below in
     one file.

  2. Top-level + per-effort plans for large initiatives — keep the sections above
     "Work Breakdown" in the top-level file; move each effort's Phase/Commit/Task
     breakdown to a separate per-effort file at:
         docs/implementation/<initiative-name>-effort-N-plan.md
     Each per-effort file follows the "Work Breakdown" section structure below.

Delete these comments before writing.
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
Discovery) it resolves if applicable. These become the source for any new
decisions-log entries.
-->

1. **[Decision]** — [rationale]
2. **[Decision]** — [rationale]

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
Concepts, conventions, or constraints that span multiple phases or efforts.
Use subsections per topic.

Examples: struct shape decisions, naming conventions, test harness setup,
mutation versioning, model discipline per phase, initiative-specific pre-merge
checklist additions, package conventions.
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
  - Medium initiative:           Commits only (each commit = one execution session).
  - Large initiative (one file): Phases → Commits → Tasks.
  - Large initiative (split):    Effort Summary above; this section lives in the
                                 per-effort plan files instead.

The structure below shows the full Phase → Commit → Task pattern. Trim levels as
appropriate for the work size.

KEY RULE: Each commit = one execution session. Title each commit with its
conventional commit message so the execution session knows the deliverable shape
going in.
-->

### Phase 1 — [Title]

<!-- Phase-level goal in 1–2 sentences. -->

#### Commit 1 — `feat: [conventional commit message]`

##### Task 1 — `path/to/file.go`

<!--
What changes in this file. Include signatures, struct definitions, or pseudocode
where the design is non-obvious. The execution session reads this and writes the
code.
-->

##### Task 2 — `path/to/other_file.go`

<!-- ... -->

#### Commit 2 — `refactor: [conventional commit message]`

<!-- ... -->

### Phase 2 — [Title]

<!-- ... -->

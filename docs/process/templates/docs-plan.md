# [Initiative Name] — Implementation Plan

<!--
COPY THIS FILE TO: docs/initiatives/implementation/<initiative-name>-plan.md

This is the plan format for MULTI-EFFORT DOCS INITIATIVES (session-modes.md
section 7). Single-session docs work needs no plan at all — if you're reaching
for this template for one doc, stop and just write the doc.

How this format differs from implementation-plan.md, and why:

  - NO Phase → Commit → Task breakdown. Generative docs discover-at-write —
    the act of writing a doc IS the act of exploring its subject, so per-file
    edits cannot be pre-specified the way code tasks can.
  - Deliverables are NAMED DOCS with scope statements, not literal edits.
  - There is usually NO discovery doc; design rationale is worked out in the
    Plan session and recorded under "Decisions Ratified in Planning."

Worked example: documentation-overhaul-plan.md (the initiative that created
this template).

Delete these comments before writing.
-->

## Context / Goal

<!--
REQUIRED. What's wrong with the current doc set, what this initiative builds,
and why the work is multi-effort. If the work splits into tracks (e.g. meta
decisions vs. generative writing), name them here.
-->

## Decisions Ratified in Planning

<!--
REQUIRED. The doc-set design decisions made in the Plan session: structure,
formats, retrieval/maintenance rules, tooling constraints. For a docs
initiative this section carries the weight Discovery would normally carry —
be generous with rationale.
-->

1. **[Decision]** — [rationale]
2. **[Decision]** — [rationale]

## Open Questions — To Ratify at Execution Time

<!--
OPTIONAL. Questions deliberately deferred, each routed to the effort that
resolves it ("Effort 2: exact template set", "Effort 3: the seeded page list").
-->

## Effort Summary

<!--
REQUIRED. One row per effort. "Depends on" makes the DAG scannable. Note here
which efforts get their own lightweight Plan session and which execute straight
from this file (session-modes.md section 8).
-->

| # | Effort | Scope | Depends on |
|---|--------|-------|------------|
| 1 | [name] | [one-line scope] | — |

## Per-Effort Breakdown

<!--
REQUIRED. For each effort: Scope (what it touches), Deliverables (the named
docs/edits it produces — specific enough that "done" is checkable), and
Dependencies (what gates it, what it feeds). Deferred tails that ride a later
effort are called out explicitly.
-->

### Effort 1 — [Name]

**Scope:**

**Deliverables:**

**Dependencies:**

## Sequencing / DAG

<!--
OPTIONAL but recommended for 3+ efforts. An ASCII DAG plus the recommended
linear order if not parallelizing.
-->

## Decision Record — Execution

<!--
REQUIRED once execution begins (empty at plan time). Execution-time decisions
and reversals land here per commit (CLAUDE.md, Collaboration 7.1). Contract:
append AND reconcile — if execution overrides a planned decision, log the
reversal with rationale and fix the plan body above so it doesn't lie.
-->

## Out of Scope

<!--
OPTIONAL. Explicit non-goals — especially future doc tooling (site generators,
viewers) and docs owned by unbuilt features.
-->

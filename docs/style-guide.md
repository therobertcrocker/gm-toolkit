# Documentation Style Guide

A reference for writing and maintaining dev journals and tracking journals consistently.

<br/>

## Document Types

**Dev Journal** (`docs/dev_journals/<tool>/dev-journal-<tool>.md`) — the living design document for a feature area. Covers intent, architecture, decisions, and progress. Updated on every branch merge.

**Tracking Journal** (`docs/dev_journals/<tool>/*-journal.md`) — a focused tracker for a specific engine or subsystem. Covers feature status, decisions made during development, and open design questions.

**Discovery Doc** (`docs/initiatives/discovery/*-discovery.md`) — written before implementation begins. Covers inputs, resolution steps, outputs, and notes for each feature or action. Not updated after the fact — it is a planning artifact.

**Implementation Plan** (`docs/initiatives/implementation/*-plan.md`) — a step-by-step plan for implementing a new feature. Includes a phased breakdown of the tasks, split by commit - with notes on the design rationale and any open questions to resolve during implementation. Updated as needed during implementation, but not after completion.

<br/>

## Section Order

### Dev Journal
1. Overview
2. Core Behaviors / Feature Summary
3. Domain-specific content (Modes, Actions, Architecture, etc.)
4. Open Questions
5. Progress (subsections in order: Up Next, Deferred, Completed)
6. Decisions Log

### Tracking Journal
1. Feature tracker table
2. Decisions Log
3. Open Questions

<br/>
<br/>

# Heading Levels

- `#` (H1) — major thematic sections (Decisions Log, Progress, Open Questions, Data Architecture)
- `##` (H2) — subsections within a major section (Core Behaviors, Design Principles, feature-level groupings)
- `###` (H3) — sub-subsections (branch names, phase names, individual mode descriptions)

**Spacing:** Use `<br/><br/>` before a new `#` section. Use a single `<br/>` between `##` sections.

<br/>
<br/>

# Tables

## Decisions Log

- Grouped by feature branch (`### feature/branch-name`) or phase (`### Discovery & Planning`)
- Numbered sequentially across all groups — do not restart per group
- Columns: `#` | `Decision` | `Rationale`
- Immediately followed by a **Notable alternatives rejected** table when meaningful alternatives were considered

**Notable alternatives rejected** table columns: `Decision #` | `Alternative` | `Why rejected`
- Only include when the alternative is non-obvious — don't document every possible path not taken
- Reference the decision number so the context is clear

## Features Tracker

Columns: `#` | `Feature` | `Status` | `Notes`

Status values:
- `Not started`
- `In progress`
- `Complete`

## Open Questions

Columns: `#` | `Question` | `Relevant Feature`

<br/>
<br/>

# Writing Style

- Rationale is concise — sentence fragments are fine
- Em dash (`—`) separates a term from its description in bullet lists: `**Term** — description`
- Code identifiers (types, functions, files, flags) use backticks inline
- No trailing periods on table cell content
- `(planned)` suffix on items not yet implemented
- `(Planned)` suffix on section headings for future features

<br/>

## Bullet Lists vs Tables

Use **bullet lists** for: behaviors, principles, mode descriptions, progress items — anything prose-like.

Use **tables** for: decisions, features, questions, data architecture — anything that benefits from structured columns.

<br/>
<br/>

# Section-Specific Guidance

## Decisions Log

Only record decisions where the rationale is non-obvious or where a meaningful alternative was rejected. Implementation details that follow naturally from prior decisions don't need their own entry.

## Progress

Three subsections: **Completed**, **Deferred**, **Up Next**
- **Completed** — done and merged; one line per meaningful unit of work
- **Deferred** — known gaps with a reason why they were deferred and what would unblock them
- **Up Next** — next planned work, in rough priority order

## Open Questions

For unresolved *design* questions only — not deferred implementation work. A question belongs here when the answer will meaningfully affect how a feature is built and we don't yet have an answer. Remove or resolve entries when a decision is made; log the decision in the Decisions Log.

<br/>

# Discovery Docs Style Guide

Referenced via relative paths in a blockquote at the top of the relevant tracking journal:

```
> Discovery docs: [name](../discovery/file.md) | [name](../discovery/file.md)
```

Discovery docs are planning artifacts — they capture the design before implementation and are not updated retroactively. Post-implementation decisions belong in the Decisions Log, not the discovery doc.

<br/>
<br/>

# Discovery Doc Style

Discovery docs follow the same heading and spacing conventions as journals, with a structure specific to their purpose.

## Heading Levels

- `#` — document title only
- `##` — major sections: numbered features (e.g. `## 1. Turn Scaffolding`) or named concepts (e.g. `## Sell Asset`, `## Responsibilities`)
- `###` — sub-sections within a major section: Inputs, Resolution Steps, Outputs, Notes, Sub-tasks, Design Notes, etc.
- `####` — sub-sub-sections where needed (e.g. Phases within a multi-phase process)

## Tone and Content

- Discovery docs describe intent and design, not implementation — write in terms of what the system does, not how the code does it
- Flag unresolved dependencies explicitly in Notes using `**Depends on X (planned)**`
- Cross-reference related discovery docs with blockquote links: `> Full breakdown in [file.md](file.md)`
- Do not update retroactively once implementation begins — post-implementation decisions belong in the Decisions Log

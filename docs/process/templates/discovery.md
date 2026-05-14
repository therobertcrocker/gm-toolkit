# [Initiative Name] — Discovery

<!--
COPY THIS FILE TO: docs/discovery/<initiative-name>-discovery.md

A discovery doc captures the design for an initiative before implementation begins.
The shape varies by what's being designed — feature work and refactor work call for
different sections. Required sections are always present; optional sections are
included when relevant.

Delete these comments before writing.
-->

## Problem

<!--
REQUIRED. What is broken, missing, limiting, or worth investigating?

For feature work: name the gap or pain point. Be richer than the planned-work blurb.
For refactor work: name the structural issue and what it costs (e.g. drift between
sub-engines, friction in the abstraction).

Reference specific code paths, files, and line numbers where helpful.
-->

## Design Summary

<!--
REQUIRED.

For feature work: name the chosen approach in a few paragraphs, with rationale.
For refactor work: name the target pattern (or, if Discovery is figuring out the
target, this section becomes "Audit Findings" — see optional sections below).

This is the section a reader skims to understand "what are we doing about it."
-->

## Open Questions

<!--
REQUIRED. Anything unresolved that the Plan session needs to decide.

If there are no open questions, write "None." Don't omit the section.
-->

---

<!--
OPTIONAL SECTIONS BELOW. Include the ones relevant to this work; delete the rest
along with this comment. Add new sections freely — discovery docs are not
template-bound, and the work itself dictates what's needed.
-->

## Audit Findings

<!--
Common in REFACTOR discovery. Inventory the current state of what's being refactored,
classify divergences from the target pattern, identify reference exemplars.

Subsections often include: Inventory, Cross-axis comparison, Reference exemplar.
-->

## Domain Model Changes

<!--
New types, field changes, struct shape decisions. One subsection per type or area.
Show signatures or struct definitions where the shape is non-obvious.
-->

## Per-Area Design Details

<!--
One section per major component, flow, or sub-feature being touched. Use
descriptive names (e.g. "Movement Phase Flow", "Bootstrap Convention",
"Cargo co-location"), not "Area 1 / Area 2".
-->

## Mutation Surface

<!--
For engine/turn/state work. New mutations, changed mutations, removed mutations.
List with brief rationale per mutation.
-->

## State Storage

<!--
What gets persisted where. TOML schema additions, FactionState field changes,
on-disk file additions.
-->

## User-Facing Impact

<!--
TUI changes, CLI changes, narrative output changes. What does the GM see differently?
-->

## Replaces / Retires

<!--
What existing code, types, patterns, or behaviors get removed or superseded.
-->

## Out of Scope

<!--
Explicit non-goals. Things that look like they're in scope but aren't. Deferred
work that should be filed in planned-work.md as a separate Backlog entry.
-->

## Reference Exemplars

<!--
Prior art, related decisions log entries, similar patterns elsewhere in the codebase.
-->

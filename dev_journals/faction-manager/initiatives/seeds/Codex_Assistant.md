# Initiative Seed — GM Assistant / Codex

> **Type:** Initiative seed, not a plan or a directive
> **Status of the work it describes:** Concept (pre-discovery). Nothing here is approved, scheduled, or promoted.
> **Source:** Brainstorm sessions in a separate Claude Project, June 2026. Canonical concept: `GM_ASSISTANT_CONCEPT.md`.
> **Written:** 2026-11-10

<br/>

## What an initiative seed is

A seed sits *before* the initiative lifecycle (Backlog → Up Next → Planned Initiative → Done). It records a piece of work that has been thought through at the concept level but has not entered the backlog. A seed carries no commitment: it may become a backlog item, change shape, or never be planted. Whether it enters the lifecycle is Robert's call.

<br/>

## Read this first

This file exists so that sessions in this repo know about a large piece of future work and don't accidentally build against it or foreclose it. It is **awareness, not authorization.**

- **Do not start work on anything described here** — no scaffolding, no `internal/codex` package, no interfaces "for later."
- **Do not promote it** in any `planned-work.md`. Promotion is Robert's call. You may suggest; you may not move it.
- **Re-ground before relying on any repo fact below.** The repo descriptions here are a snapshot from mid-2026 (`main` @ `e3a4538`). The TUI arc, the engine freeze, and package layout may have moved since. Read the current code and `planned-work.md` first.
- **YAGNI still applies.** Knowing this work is coming is not a reason to add extension points today. The only legitimate effect of this seed on present work is *avoiding a choice that would make the future work painful*, and flagging it when you see one.

<br/>
<br/>

# What's coming

**The GM Assistant** is a long-horizon concept: a structured campaign-and-rules substrate with a deterministic core, with "LLM-shaped sockets" at the read and write edges. The committed part is the substrate and core; any LLM integration is deferred and pluggable.

Three real pieces:

- **The Codex** — the campaign as a structured, mutable entity graph: characters, factions, events, locations, lore terms, artifacts, connected by typed edges. This is the planned **tool #2** in `docs/project-plan.md`, and the near-term bridge to the larger vision.
- **The Rulebook (PF2e)** — the Pathfinder 2e ruleset as structured, read-only data, imported from existing datasets (Foundry VTT `pf2e`, pf2ools) rather than hand-keyed. Used for lookup, aggregation, and rules-correct generation (creature/NPC stat blocks, encounter math).
- **The deterministic core** — ingestion/commit, retrieval/context-assembly, and generation, exposed as tools.

And two edges:

- **Write socket** — producers propose structured operations (events carrying mutations and entity mints); the core validates, resolves identity, and commits. **The faction manager is the first intended producer.**
- **Read socket** — a consumer (eventually an LLM assistant; possibly a CLI) composes queries; the core answers precisely.

Guiding rule: **code computes, the model interprets.** No LLM is ever in a precision-critical path.

<br/>
<br/>

# Where it touches this repo

These are the points of contact with existing code. They are the reason this seed exists.

<br/>

## 1. The faction manager is the first write-socket producer

The concept positions the faction manager as proof that a non-LLM producer can feed the Codex through a defined contract. That contract — the **event contract** (event-entity schema + mutation/mint format) — is not designed yet.

Its likely raw material already exists here: the **mutation engine** (single write path that writes state and records the change) and the **JSONL turn history** that `narrate` renders from.

*Practical effect today:* when touching the mutation engine or the history format, prefer choices that keep a mutation's target, old → new value, and cause recoverable from the record. Do **not** build an export, adapter, or event contract now.

<br/>

## 2. The Codex shares `internal/spatial`

`internal/spatial` is deliberately standalone so the Codex can depend on it. The plan is for the Codex to own the canonical Fragment/Region TOML that the faction tool reads.

*Practical effect today:* keep `internal/spatial` free of faction-specific imports and assumptions. That was already the rule; this is the reason behind it.

<br/>

## 3. The same write-path philosophy, on both sides

The faction engine's "mutations are the single write path; the entity owns the truth; not event-sourced" is the **same commitment** the Codex is built on. If a change here starts deriving current state by replaying history, it is drifting away from both designs — flag it.

<br/>

## 4. Naming collision: "Rulebook"

`engine.Engine` already owns a `Rulebook` (the SWN-inspired faction rules, data under `rulebooks/swn/`). The GM Assistant's **Rulebook** is the *PF2e* ruleset — a different thing in a different domain.

*Practical effect:* when discussing or naming things, disambiguate (e.g. "faction rulebook" vs. "PF2e Rulebook"). Whether to rename either is an open decision for discovery, not for now.

<br/>
<br/>

# Design commitments to uphold

If present-day work touches any of these, keep them intact — and say so openly if one stops fitting rather than quietly working around it.

- **Entities own the truth; not event-sourced.** Current state is read off the entity. A change-log is written on every mutation and read only for provenance, ordering, or deliberate repair — never to answer a query.
- **Events are entities**, carrying mutations/mints that apply once at commit and are then frozen as provenance. Narrative events ("what happened") are a subset of mutation causes; a typo fix is not a narrative event.
- **One source of truth per fact.**
- **Composition for per-entity attributes; edges for relationships.** ECS's composition lens, not its architecture.
- **Determinism at the commit boundary.** Fuzzy producer output is validated (ideally human-confirmed) before it becomes canonical.
- **Explicitly not this:** Datomic / XTDB-style immutable fact logs with projected state, and bitemporality. These were considered and rejected.

<br/>
<br/>

# Open questions (for eventual discovery — not to be answered now)

- The event contract: event-entity schema and mutation/mint format.
- Entity resolution at both edges — the single riskiest area.
- The graph-primitive / query-surface contract between core and sockets.
- Persistence per store (files / SQLite / graph store) — still deferred, consistent with the standing deferral.
- Knowledge-state grain (player-known vs. secret: entity, edge, or fact).
- Which PF2e dataset to import, what the transform must preserve, and licensing (ORC / OGL / Apache-2.0 / Paizo terms).
- Whether the faction rulebook and PF2e Rulebook naming should be reconciled.

<br/>
<br/>

# If this seed is planted (Robert to decide)

If and when Robert moves this into the backlog, a reasonable shape for the entry:

- **Type:** `feature` (likely an **arc**, given it spans multiple shippable initiatives over shared architecture)
- **Title:** Codex — campaign knowledge graph (tool #2)
- **Stage:** Backlog
- **Note:** Pre-discovery. See `GM_ASSISTANT_CONCEPT.md` and this seed. Blocked on nothing technical; sequencing relative to the TUI arc is Robert's call.

<br/>
<br/>

# Mini-glossary additions

- **Codex** — the campaign knowledge graph (planned tool #2).
- **PF2e Rulebook** — imported, read-only PF2e rules data. Not the faction engine's `Rulebook`.
- **Deterministic core** — ingestion/commit, retrieval, generation; the trustworthy middle.
- **Socket** — a defined, LLM-shaped interface at the read or write edge; deferred and pluggable.
- **Event contract** — the format producers (faction manager, future session parser) use to propose changes to the Codex.
- **Parsable component** — a mutation or mint carried on an event entity; proposed before commit, applied at commit, frozen as provenance after.
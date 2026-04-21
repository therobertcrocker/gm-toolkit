# Turn Engine — Tracking Journal

A high-level tracker for features, tasks, and decisions made during turn engine development.

> Discovery docs: [turn-engine-discovery.md](../discovery/turn-engine-discovery.md) | [action-resolution-discovery.md](../discovery/action-resolution-discovery.md) | [goal-engine-discovery.md](../discovery/goal-engine-discovery.md)


<br/>

## Features

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 1 | Turn Scaffolding | Complete | |
| 2 | Per-Faction Bookkeeping | In progress | Income and maintenance skeleton done; maintenance costs deferred until AssetDefinition carries structured cost data |
| 3 | Action Selection | Not started | |
| 4 | Action Resolution | Not started | Depends on Tag Engine, Ability Engine |
| 5 | State Mutation | Not started | |
| 6 | Event Recording | Not started | |
| 7 | Goal Engine | Not started | Depends on Action Resolution |
| 8 | Tag Engine | Not started | Scope TBD; discovery doc pending |

<br/>
<br/>

# Decisions Log

### Discovery & Planning

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | Actions implement a common interface (Inputs, Validate, Resolve, Output) | Keeps the Action Engine source-agnostic; same interface works for GM input and future AI agent |
| 2 | Actions produce a mutation list and event record; engine applies both at turn end | Decouples action logic from state writes; keeps actions testable; prevents partial writes on paused turns |
| 3 | Change Homeworld and Seize Planet moved to Goal Engine | Both are multi-turn stateful processes with locking behavior — closer to goal tracking than single-turn action resolution |
| 4 | Tag Engine is a shared utility, not owned by any action | Tags surface in Attack, faction tests, and potentially elsewhere; logic should not be duplicated |
| 5 | Ability Engine handles asset ability resolution | Each `A`-flagged asset has bespoke logic; a dedicated sub-engine keeps the action engine clean |
| 6 | Faction test is a shared engine utility, not owned by Ability Engine | Faction tests may surface outside of asset abilities |
| 7 | Defender asset selection is manual for now | AI agent defender logic is a future concern |
| 8 | Hex distance for Change Homeworld is entered manually for now | Map engine is planned but not yet scoped; door left open for automatic calculation |
| 9 | Pub/sub deferred for state mutation; may revisit for event recording | State mutation has one consumer (state writer) — pub/sub is overkill; event recording may eventually fan out to narrative renderer and AI observer |
| 10 | Edit Mode added as a third top-level mode | GMs need to manipulate faction state directly for campaign setup without mechanical validation |

**Notable alternatives rejected:**

| Decision # | Alternative | Why rejected |
|------------|-------------|--------------|
| #2 | Full event sourcing instead of mutation list | Overkill for this data volume and access pattern; mutation list is sufficient |
| #5 | Actions as TOML-driven data | Resolution logic is too conditional to live in data; code is the right home |

### feature/turn-engine-scaffolding

| # | Decision | Rationale |
|---|----------|-----------|
| 11 | `Engine` restructured as core orchestrator composing sub-engines (`TurnEngine`, with `ActionEngine`, `GoalEngine`, `TagEngine` as future fields) | Maps to discovery doc design; keeps sub-engines focused and testable; dependencies explicit |
| 12 | `TurnEngine` holds no back-reference to `Engine` | No dependency needed now; if Rulebook access is required later, pass `*loader.Rulebook` directly |
| 13 | Faction order is a rotation, not a shuffle | Rules specify "roll a die, proceed in order" — a starting index rotation satisfies this |
| 14 | `ApplyBookkeeping` is idempotent via `TurnPhase` guard | Skips silently if phase is not `PhaseBookkeeping`; prevents double-apply on resume |
| 15 | `Advance` returning `true` is the seam for Mutation and History engines | Single clean signal for turn completion; future engines plug in here |
| 16 | `maintenanceCost` returns 0 until `AssetDefinition` carries structured cost data | Maintenance costs are in description text only; deferred until modelled and Rulebook-resolved |
| 17 | Mutation and History writes commit together at the end of each faction's turn | Keeps state and history always in sync; pause/resume is correct because each faction's commit lands before the next faction's bookkeeping runs |
| 18 | `Asset.Maintained` doubles as the consecutive-miss tracker for the two-turn loss rule | First missed payment sets `Maintained = false`; second consecutive miss (already false when we try to pay) destroys the asset — no extra field needed |
| 19 | `TurnPhase` is an int/iota enum replacing `BookkeepingApplied bool` | Three phases needed (Bookkeeping, Action, Complete) for correct pause/resume across action resolution; a bool cannot represent the Action phase |

**Notable alternatives rejected:**

| Decision # | Alternative | Why rejected |
|------------|-------------|--------------|
| #11 | All methods on a single `Engine` struct | Would work today but makes sub-engine dependencies implicit as the codebase grows; composition keeps them explicit and testable in isolation |
| #13 | Full shuffle instead of rotation | Rules specify a fixed list order with a random start, not random ordering each turn |
| #17 | Defer all writes to end of full round | If interrupted mid-round, all prior factions would need to re-act on resume; per-faction commits are the correct granularity given pause/resume is first-class |

<br/>
<br/>

# Open Questions

Design questions that are unresolved and will need answers before the relevant feature can be built.

| # | Question | Relevant Feature |
|---|----------|-----------------|
| 1 | How does the turn command surface resume detection — prompt to resume or abandon at startup, or a dedicated sub-command? | `turn` command |
| 2 | How does the turn command handle an action-locked faction (mid-move, mid-seize) — skip automatically with a message, or present it as a distinct "no action" step? | Goal Engine / `turn` command |
| 3 | What is the interaction model for the turn command — a single interactive wizard stepping through all factions, or discrete sub-commands per step? | `turn` command |
| 4 | How does the Goal Engine communicate lock state back to the turn flow — method call, return value, or flag on `TurnState`? | Goal Engine |

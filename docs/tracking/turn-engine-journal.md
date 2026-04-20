# Turn Engine — Tracking Journal

A high-level tracker for features, tasks, and decisions made during turn engine development.

> Discovery docs: [turn-engine-discovery.md](../discovery/turn-engine-discovery.md) | [action-resolution-discovery.md](../discovery/action-resolution-discovery.md) | [goal-engine-discovery.md](../discovery/goal-engine-discovery.md)

---

## Features

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 1 | Turn Scaffolding | Not started | |
| 2 | Per-Faction Bookkeeping | Not started | |
| 3 | Action Selection | Not started | |
| 4 | Action Resolution | Not started | Depends on Tag Engine, Ability Engine |
| 5 | State Mutation | Not started | |
| 6 | Event Recording | Not started | |
| 7 | Goal Engine | Not started | Depends on Action Resolution |
| 8 | Tag Engine | Not started | Scope TBD; discovery doc pending |

---

## Decisions Log

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

---

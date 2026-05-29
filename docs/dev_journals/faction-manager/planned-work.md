# Planned Work

Pre-discovery initiative tracker. Deferred items move to Planned Initiatives when scoped -- at that point a write-up is added below and a discovery doc is written before implementation begins.

> **Scheduling note (2026-05-23):** The TUI rebuild (`F-005`) is the first of a planned arc to bring the TUI to feature completeness. No new engine features will be promoted from Backlog to Up Next until that arc is complete. Bugfixes and refactors are unaffected.
>
> **Addendum (2026-05-25):** `F-012` (spatial-map-cli) promoted as content-authoring tooling, exempt from the engine-feature freeze. It is a prerequisite for `F-005.2` (Manage needs an authored spatial map to render).
>
> **Addendum (2026-05-26):** `F-012` shipped (`feature/spatial-map-cli` merged). Prerequisite for `F-005.2` is met.

<br/>

## Current Initiatives

Full write-ups below. Each item has been scoped enough to warrant a dedicated discovery phase.

| ID      | Item | Type | Status | Detail |
|---------|------|------|--------|--------|
| `A-005` | [TUI Rebuild](#f-005--tui-rebuild) | Arc | In-Progress | First user-facing surface against the reshaped engine; greenfield rebuild. |


---
<br/>

## Up Next (Ready or Near Ready)
The queue of deferred items that are ready to become initiatives. These are scoped and waiting for their trigger conditions to be met.

| ID      | Item | Type | Trigger | Detail |
|---------|------|------|---------|--------|
| `F-005.3` | [tui-turn](#f-0053--tui-turn) | feature | `F-005.2` merged ✓ | Turn mode: setup, execution, modal overlay, Esc-cancel path. |

---
<br />

## Backlog

Unscoped items waiting for their trigger. Move to Up Next when the trigger is close; move to Planned Initiatives when fully scoped for discovery.

| ID      | Item | Type | Trigger | Detail |
|---------|------|------|---------|--------|
| `F-004` | CLI Rebuild | feature | -- | -- |
| `F-010` | Tag-Granted Assets | feature | -- | -- |
| `F-011` | Tag Reminders | feature | -- | -- |
| `R-003` | Action Preconditions | refactor | AI/planner | Lift `Action.Validate` into `Action.Preconditions []Precondition` |
| `F-013` | Goal State Predicates | feature | AI/planner | Add `Satisfied(state) bool` alongside per-goal `progressX`; both shapes coexist |
| `B-001` | Hardcoded IDs | bugfix | -- | `stealth_applicator` (asset) and `"G-012"` (ChangeHomeworld goal) are hardcoded; silently break if TOML IDs change |
| `B-002` | Goal/Tag Display IDs | bugfix | -- | Key `Rulebook.Goals`/`Rulebook.Tags` on semantic table key; `id` becomes display-only. |
| `B-003` | `SeizePlanet` History | bugfix | -- | Currently invisible in history. `Output()` returns no mutations; add `GoalPhaseAdvanced` mutation |
| `B-004` | `Ability` Narration | bugfix | -- | Add fallback text for asset ability narration |
| `B-005` | Goal XP | bugfix | -- | Retype `Difficulty` from `string` to int or tagged sum so engine can dispatch XP |
| `F-014` | A-flag Abilities | feature | `R-001` | Nine A-flag abilities stubbed |
| `R-004` | Data-Driven Handlers | refactor | High Pain | Replace hardcoded goal/ability handler dispatch with TOML-defined handlers + typed primitive registry |
| `R-005` | Adjacency vs. Warp Cost | refactor | Deferred from `F-012` | Split `Region.Boundary` into adjacency-edges (cost 1) and warp-edges (cost `crossingCost`) |
| `D-001` | Architecture Overview Doc | docs | -- | `docs/architecture-overview.md` does not exist. |
| `F-015` | TUI Manage: Map View | feature | -- | add a map view to the manage screen, showing the faction's homeworld and the planets it controls. |
| `F-016` | TUI Manage: Right Panel | feature | -- | add a right-hand panel to the manage screen, showing additional faction details at a glance. |
| `R-006` | Shared `styles.Dim` constant | refactor | -- | `#6C7086` duplicated across detail/list/manage/layout; extract one shared constant. |
| `R-007` | Extract `rebuildList()` helper | refactor | -- | Created/Deleted branches in manage.go copy-paste the rebuild sequence; extract a shared helper. |
| `R-008` | Reuse `world.Location` + `campaigns.ValidateID` in wizard | refactor | -- | `synthesize`/`slugify` reinvent location handling and ID validation; delegate to existing helpers. |
| `R-009` | Remove write-only `detail.Model.height` | refactor | -- | Field is assigned but never read; dead weight in the detail model. |
| `R-010` | Unify `WindowSizeMsg` routing in manage.go | refactor | -- | Per-view asymmetry in how size messages are forwarded; normalise to one pattern. |
| `R-011` | Cache per-frame allocations in `compose()` / root `View()` | refactor | -- | Low priority: repeated allocations on hot path; pre-allocate or cache where safe. |
---
<br />
<br />

# Initiative Write-Ups
This section contains detailed write-ups for each planned initiative, including problem statements, proposed approaches, tradeoffs, and triggers. This will be the source of truth when it comes time to start discovery work on any of these items.

<br />

### A-005 — TUI Rebuild

**Status:** In-Progress
**Type:** Arc
**Trigger:** Movement-redesign landing (now complete — `f09e4cc`).

**Problem.** The engine reshape arc — movement-redesign, engine-collector-reshape, mutation taxonomy, structured logging, error conventions — was undertaken in service of a clean TUI rebuild. V1 was deleted because it had grown into an orchestrator masquerading as a UI: logic that belonged in the engine lived in TUI code, making both harder to reason about. With that work landed, the engine now exposes the surface a UI can consume cleanly. There is currently no user-facing surface at all; F-004 (CLI rebuild) is still in Backlog, so the TUI rebuild is the toolkit's first interaction surface against the reshaped engine.

**Approach.** Greenfield rebuild — new design, new implementation, no carry-over from V1. The TUI is a pure consumer: it reads engine state and forwards user intent through the orchestrator, never owning game logic. Concrete shape (file layout, framework choice, view structure, input model) is the work of Discovery.

**Unlocks.**
- Restores any interactive surface to the toolkit.
- First real consumer of the reshaped engine's external API — will surface gaps and friction before F-004 inherits them.
- Gives mutations, narrative, and errors a structured rendering venue rather than living only in logs.

<br />

### F-005.3 — tui-turn

**Status:** Backlog
**Type:** Feature
**Trigger:** `F-005.2` merged.
**Part Of:** [A-005 — TUI Rebuild](#a-005--tui-rebuild) — Initiative 3 of 3.

**Goal.** Ship the Turn mode end-to-end: setup view, execution view, modal overlay for collector prompts, cadence-flag toggle, and Esc-aborts-current-turn cancel path. After this initiative the GM can drive engine cycles interactively from the TUI.

**Open Questions (resolve in Plan).**
- What `Start` invokes — `RunCycle` vs `RunFactionTurn` in a loop?
- Error rendering shape — where does `OnError` surface in the UI?

See [`tui-rebuild-arc-plan.md`](../../initiatives/arcs/tui-rebuild/tui-rebuild-arc-plan.md) Initiative 3 for full scope and estimated commit shape.

<br />



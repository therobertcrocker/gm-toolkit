# Planned Work

Pre-discovery initiative tracker. Deferred items move to Planned Initiatives when scoped -- at that point a write-up is added below and a discovery doc is written before implementation begins.

> **Scheduling note (2026-05-23):** The TUI rebuild (`F-005`) is the first of a planned arc to bring the TUI to feature completeness. No new engine features will be promoted from Backlog to Up Next until that arc is complete. Bugfixes and refactors are unaffected.

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
| `F-005.2` | [tui-manage](#f-0052--tui-manage) | feature | `F-005.1` merged | Manage mode: faction list, detail, create, edit, delete. |
| `F-005.3` | [tui-turn](#f-0053--tui-turn) | feature | `F-005.2` merged | Turn mode: setup, execution, modal overlay, Esc-cancel path. |

---
<br />

## Backlog

Unscoped items waiting for their trigger. Move to Up Next when the trigger is close; move to Planned Initiatives when fully scoped for discovery.

| ID      | Item | Type | Trigger | Detail |
|---------|------|------|---------|--------|
| `F-004` | CLI Rebuild | feature | -- | -- |
| `F-007` | Campaign-Scoped Static Data | feature | -- | -- |
| `F-010` | Tag-Granted Assets | feature | -- | -- |
| `F-011` | Tag Reminders | feature | -- | -- |
| `F-012` | Spatial Map CLI | feature | -- | Generate `worlds.toml` from a user-provided template or data source |
| `R-003` | Action Preconditions | refactor | AI/planner | Lift `Action.Validate` into `Action.Preconditions []Precondition` |
| `F-013` | Goal State Predicates | feature | AI/planner | Add `Satisfied(state) bool` alongside per-goal `progressX`; both shapes coexist |
| `B-001` | Hardcoded IDs | bugfix | -- | `stealth_applicator` (asset) and `"G-012"` (ChangeHomeworld goal) are hardcoded; silently break if TOML IDs change |
| `B-002` | Goal/Tag Display IDs | bugfix | -- | Key `Rulebook.Goals`/`Rulebook.Tags` on semantic table key; `id` becomes display-only. |
| `B-003` | `SeizePlanet` History | bugfix | -- | Currently invisible in history. `Output()` returns no mutations; add `GoalPhaseAdvanced` mutation |
| `B-004` | `UseAssetAbility` Narration | bugfix | -- | Add fallback text for asset ability narration |
| `B-005` | Goal XP | bugfix | -- | Retype `Difficulty` from `string` to int or tagged sum so engine can dispatch XP |
| `F-014` | A-flag Abilities | feature | `R-001` | Nine A-flag abilities stubbed |
| `R-004` | Data-Driven Handlers | refactor | High Pain | Replace hardcoded goal/ability handler dispatch with TOML-defined handlers + typed primitive registry |
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

### F-005.2 — tui-manage

**Status:** Backlog
**Type:** Feature
**Trigger:** `F-005.1` merged.
**Part Of:** [A-005 — TUI Rebuild](#a-005--tui-rebuild) — Initiative 2 of 3.

**Goal.** Ship the Manage mode end-to-end: faction list, detail, create, edit, delete. After this initiative the TUI replaces hand-edited TOML for the full faction authoring loop.

**Open Questions (resolve in Plan).**
- Faction CRUD form scope — identity and stats only, or goals/assets inline at create?
- Empty-state UX grammar (hint copy, key prompt, visual treatment) — Turn inherits the choice.
- `Adapter.Stop()` cancellation design: `close(eventCh)` panics if the engine goroutine is still emitting. Choose a strategy (context, done channel, or guarantee engine completes before TUI exits) before wiring `Adapter.Run()`.
- `tui.Run(nil, nil, nil, log)`: update to pass real engine/state/cfg at wiring time; current nil parameters are safe only while Foundation's TUI paths don't reach engine methods.

See [`tui-rebuild-arc-plan.md`](../../initiatives/arcs/tui-rebuild/tui-rebuild-arc-plan.md) Initiative 2 for full scope and estimated commit shape.

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



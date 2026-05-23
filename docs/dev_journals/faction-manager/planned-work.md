# Planned Work

Pre-discovery initiative tracker. Deferred items move to Planned Initiatives when scoped -- at that point a write-up is added below and a discovery doc is written before implementation begins.

<br/>

## Current Initiatives

Full write-ups below. Each item has been scoped enough to warrant a dedicated discovery phase.

| ID      | Item | Type | Status | Detail |
|---------|------|------|--------|--------|


---
<br/>

## Up Next (Ready or Near Ready)
The queue of deferred items that are ready to become initiatives. These are scoped and waiting for their trigger conditions to be met.

| ID      | Item | Type | Trigger | Detail |
|---------|------|------|--------|--------|
| `F-005` | TUI Rebuild | feature | -- | Build a full TUI for the new engine's richer state and more complex interactions. |
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
| `B-001` | Hardcoded IDs | bugfix | -- | `stealth_applicator` (asset) and `"G-012"` (ChangeHomeworld goal) are hardcoded; silently break if TOML IDs change; add typed constants or load from rulebook |
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



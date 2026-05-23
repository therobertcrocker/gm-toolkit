# Planned Work

Pre-discovery initiative tracker. Deferred items move to Planned Initiatives when scoped -- at that point a write-up is added below and a discovery doc is written before implementation begins.

<br/>

## Current Initiatives

Full write-ups below. Each item has been scoped enough to warrant a dedicated discovery phase.

| ID      | Item | Type | Status | Detail |
|---------|------|------|--------|--------|
| `F-003` | Logging + Errors | feature | In-Progress (Effort 2) | Effort 1 (logging) merged; Effort 2 (error conventions, Recoverable/Fatal branch) on `feature/errors` |


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

## F-003: Logging + Errors

### Guiding Principles

- **Consistency** — one pattern for errors, one pattern for logging, everywhere
- **Ease of use** — should not require ceremony; the right behavior should be the default behavior
- **Structured over vague** — errors carry typed context; log output is human readable but also machine-parseable; avoid unstructured strings where possible
- **Surface over silence** — when something goes wrong, it should be visible in a log, not just an error return that may or may not be checked

### Problem

No structured logging exists in the project. `main.go` has a single `fmt.Println` for usage output; nothing else reaches a log surface. Error handling across 25+ files is ad hoc: some packages define sentinel errors (`spatial`, `engine/core`), others inline `fmt.Errorf` strings with no consistency. `TurnObserver.OnError` accepts a raw `error` — there is no type information, no severity, and no structured context attached. When something goes wrong, there is no trail and no way to distinguish a recoverable data-load failure from a logic bug.

### Approach

Discovery decides. Expected scope: audit current error-return patterns across all packages; evaluate `log/slog` (stdlib) as the logging backend; decide on a log output strategy (file-first, with level-gated stderr fallback likely); propose error conventions — sentinel vs. typed vs. wrapped — and where each belongs. No pre-determined conclusions beyond the guiding principles above.

### Unlocks

- Reliable debug trail during active development and GM sessions
- Cleaner foundation for `F-005` (TUI Rebuild) — TUI needs a log surface separate from its display output

### Trigger

Active — no blocking dependencies.

### Status

In-Progress. Effort 1 (logging) complete and merged on `feature/logging`. Effort 2 (error conventions, `isRecoverable` orchestrator branch, `MutationApplyError` consumer) in execution on `feature/errors`.

<br/>

## Ability Engine Redesign

### Problem

The ability sub-system has accumulated debt across several refactor cycles and was never given a deliberate design of its own. How abilities are dispatched, how step handlers are structured, how input is collected, and where package boundaries should sit are all open questions. Discovery should start from scratch -- audit what exists, identify the pain points, and decide what the right model looks like.

### Approach

Discovery decides. No pre-determined conclusions.

### Unlocks

- Clean foundation for `F-014` (nine A-flag abilities stubbed, awaiting dispatch)

### Trigger

Active -- `R-004` complete; `F-014` is the next blocked work.

### Status

In-Progress -- discovery complete; structural fold into action sub-engine underway on `feature/ability-engine-redesign`. The `ability/` package will be retired and its contents redistributed as appropriate.

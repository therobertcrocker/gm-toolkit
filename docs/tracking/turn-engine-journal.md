# Turn Engine — Tracking Journal

A high-level tracker for features, tasks, and decisions made during turn engine development.

> Discovery docs: [turn-engine-discovery.md](../discovery/turn-engine-discovery.md) | [action-resolution-discovery.md](../discovery/action-resolution-discovery.md) | [goal-engine-discovery.md](../discovery/goal-engine-discovery.md)


<br/>

## Features

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 1 | Turn Scaffolding | Complete | |
| 2 | Per-Faction Bookkeeping | In progress | Income and maintenance skeleton done; maintenance costs deferred until AssetDefinition carries structured cost data |
| 3 | Action Selection | In progress | ActionEngine and Action interface built; Sell Asset wired; remaining actions not started |
| 4 | Action Resolution | In progress | Sell Asset complete; remaining actions not started; depends on Tag Engine, Ability Engine for Attack and Use Asset Ability |
| 5 | State Mutation | In progress | MutationEngine scaffolded; Apply wired into bookkeeping and action phase; history recording deferred |
| 6 | Event Recording | Complete | Per-faction JSONL records; one EventRecord per faction per cycle; HistoryEngine wired into turn wizard |
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
| #17 | Defer all writes to end of full Cycle | If interrupted mid-Cycle, all prior factions would need to re-act on resume; per-faction commits are the correct granularity given pause/resume is first-class |

### feature/turn-command

| # | Decision | Rationale |
|---|----------|-----------|
| 20 | Resume detection prompts at startup if an in-progress turn exists | Natural entry point; GM doesn't need a separate sub-command to resume normal flow |
| 21 | Action-locked factions are presented to the GM for confirmation before skipping | Keeps the GM informed; avoids silent skips that could be confusing mid-campaign |
| 22 | `turn` command is a fully interactive wizard stepping through all factions in sequence | One continuous flow per turn; consistent with the `faction create` wizard pattern |
| 23 | `Mutation` interface defined in `domain` with `Type()` and `Describe()` | `Type()` gives a stable discriminator for history serialization; `Describe()` is the narrative renderer contract; domain owns the contract, engine owns the logic |
| 24 | `MutationEngine` is the sole writer to campaign state; `BookkeepingResult.Mutations` bridges calculation to application | Centralizes all state writes; makes bookkeeping logic testable without side effects; slots naturally into history recording when that engine is built |
| 25 | `applyMaintenance` receives a pointer-to-slice and a `startCoin` (post-income balance) | Income mutation is prepended by the caller before maintenance runs, so the simulated running balance is correct; pointer-to-slice avoids a messy return tuple |
| 26 | `godotenv` + `config.go` for environment variable management | Separates runtime config from the binary; dev path lives in `.env`, never hardcoded; `LoadConfig()` is the single point of failure with a clear error message |
| 27 | ANSI escape codes for terminal styling in the turn wizard; `lipgloss` deferred to TUI layer | `lipgloss` is a layout library designed for Bubbletea TUI components, not sequential CLI text output; ANSI codes are sufficient and add no dependency |

### feature/action-engine

| # | Decision | Rationale |
|---|----------|-----------|
| 28 | `Action` interface lives in `engine`, not `domain` | Depends on `loader.Rulebook` and `state.FactionState`; domain cannot import either |
| 29 | Actions are stateful structs; Inputs populates fields, Resolve reads them | Keeps interface signatures uniform; no generic input bag or type assertions needed |
| 30 | `ActionEngine` uses factories (`func() Action`) rather than registered instances | Ensures a fresh zero-value struct per faction turn; prevents stale state from a previous faction carrying over |
| 31 | `InputCollector` interface injected into actions via factory; `GMCollector` in `cmd/forms` | Keeps `huh` out of the engine; AI agent implements the same interface with goal-driven logic; `Inputs` method unchanged between GM and AI modes |
| 32 | Action selection prompt lives in the wizard, not `ActionEngine` | `ActionEngine` is pure orchestration; UI belongs in the command layer |
| 33 | `No Action` hardcoded in the wizard, not registered as an action | GM UI affordance, not a game mechanic; keeps it out of AI action selection |

**Notable alternatives rejected:**

| Decision # | Alternative | Why rejected |
|------------|-------------|--------------|
| #30 | Register action instances directly | Shared instance retains state from prior faction's action phase if resolution fails mid-way |
| #31 | `huh` calls directly inside action `Inputs` | Couples engine to a UI library; AI agent would require a different concrete type rather than a different collector |

### feature/history-engine

| # | Decision | Rationale |
|---|----------|-----------|
| 34 | `Describe()` removed from `Mutation` interface | Description is a renderer concern; coupling prose to the mutation type locks the renderer to pre-baked strings and mixes display logic into the domain |
| 35 | `MutationRecord` stores a `json.RawMessage` payload alongside the type discriminator | Preserves the full structured mutation data for querying; avoids custom marshalers; sidesteps interface serialisation issues |
| 36 | History records at per-faction granularity, not per-Cycle | Consistent with per-faction state commits; a per-Cycle record would require buffering history until Cycle end while state already commits per-faction, creating a sync gap on interrupted Cycles |
| 37 | `HistoryEngine` is a separate engine, not folded into `MutationEngine.Apply` | Single responsibility; history engine can grow (readers, narrative renderer) without touching the mutation path |
| 38 | Index-based `huh.Select` for action selection | Interface equality is unreliable in huh's option matching; integer indices are unambiguous and avoid the lookup entirely |

**Notable alternatives rejected:**

| Decision # | Alternative | Why rejected |
|------------|-------------|--------------|
| #36 | Per-Cycle event record (original discovery doc design) | State commits per-faction for pause/resume correctness; deferring history to Cycle end would create a sync gap on interrupted Cycles |
| #38 | `engine.Action` directly as huh option value | huh's option matching behaved unexpectedly with interface values; integer indices are unambiguous |

### chore/code-review-2

| # | Decision | Rationale |
|---|----------|-----------|
| 39 | `BookkeepingResult.Mutations` renamed to `RecordedMutations` | Mutations are already applied inside `ApplyBookkeeping`; the field is carried for history recording only — the old name implied the caller should apply them |
| 40 | Concrete actions moved to `engine/actions` sub-package | `ActionEngine` holds only the interface and factory registry and never references concrete types; `engine/actions` imports `engine` for the contract with no circular import; establishes the correct home before the action list grows |

<br/>
<br/>

# Open Questions

Design questions that are unresolved and will need answers before the relevant feature can be built.

| # | Question | Relevant Feature |
|---|----------|-----------------|
| 1 | How does the Goal Engine communicate lock state back to the turn flow — method call, return value, or flag on `TurnState`? | Goal Engine |

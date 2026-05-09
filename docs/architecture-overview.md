# GM Toolkit — Architecture Overview

This document is written for a mid-level Go engineer taking over active development of the gm-toolkit project. It assumes you're comfortable with Go, have built CLI tools before, and just need to understand how this project is structured and why.

---

## What This Project Is

GM Toolkit is a command-line tool for tabletop RPG game masters. Its current focus is the **faction tracker** for *Stars Without Number* (SWN), a sci-fi tabletop RPG. Factions are organizations — corporations, governments, pirate fleets, cults — that compete for influence between player sessions. The GM runs faction turns, during which each faction earns income, pays maintenance on its assets, optionally raises a stat, and takes one action (attacking a rival, buying new units, expanding to a new world, etc.).

The tool is a single Go binary with no server component and no external dependencies beyond the local filesystem. Campaign state lives in TOML files on disk.

Understanding the game domain matters here. The codebase models real game mechanics — stat ratings, asset costs, attack formulas — so decisions in the code are often driven by rules in the rulebook (`docs/swn-faction-mechanics.md`). When something seems arbitrary, check there first.

---

## Repo Layout

```
gm-toolkit/
├── cmd/
│   └── faction-manager/          # CLI entry point
│       ├── main.go               # Root command bootstrap
│       ├── config.go             # Runtime config loading (FACTION_DATA_DIR)
│       ├── commands/             # Cobra subcommands
│       │   ├── app.go            # Root command, engine wiring
│       │   ├── faction/          # Faction CRUD (create, delete)
│       │   │   └── wizard/       # Multi-step huh forms for faction setup
│       │   └── narrate/          # Narrative generation from cycle history
│       └── paths/                # Campaign path resolution (state/history/narratives)
├── internal/
│   └── faction/                  # All faction domain and business logic
│       ├── domain/               # Pure data types — no I/O, no framework imports
│       ├── state/                # TOML persistence layer
│       ├── rulebook/             # Loads static TOML files into a Rulebook struct
│       ├── data/                 # Static game data (assets, tags, goals) as TOML files
│       ├── config/               # Engine config struct (StatePath, HistoryPath)
│       ├── narrative/            # Cycle history → markdown narrative renderer
│       └── engine/               # Turn lifecycle, actions, mutations, hooks, history
│           ├── core.go           # Composition root; builds all sub-engines
│           ├── orchestrator.go   # Drives one faction's turn through the pipeline
│           ├── observer.go       # TurnObserver interface (fire-and-forget output)
│           ├── collector.go      # InputCollector interface (GM / AI / test input)
│           ├── roller.go         # Roller interface + default math/rand implementation
│           ├── action/           # Action interface, factory, and all action types
│           ├── ability/          # Asset ability step handlers
│           ├── goal/             # Goal locking, progress tracking, XP computation
│           ├── history/          # Appends EventRecords to JSONL history file
│           ├── hooks/            # Five-category hook system + scoped registry
│           ├── mutation/         # Applies mutation lists to FactionState
│           ├── tag/              # Registers faction hooks for known tag types
│           ├── turn/             # Turn order, bookkeeping, phase advancement
│           └── testharness/      # Scripted collector + recording observer for tests
├── docs/                         # Developer documentation and design journals
└── reference/                    # SWN rulebook excerpts and reference material
```

The split between `cmd/` and `internal/` is intentional and meaningful. Everything under `internal/faction/` is pure game logic — it knows nothing about terminals, prompts, or Cobra. Everything under `cmd/faction-manager/` is delivery mechanism — it knows about the terminal but delegates all game rules to the internal packages. This boundary makes the engine testable in isolation and means you can reason about game logic without thinking about UI.

---

## Core Domain Model

### Faction

A `Faction` is the central entity. It has:

- **Three stat pillars** — Force, Cunning, and Wealth, each rated 1–8. These determine what the faction can buy, how it fights, and how it earns income.
- **HP** — a health pool. When a faction takes damage (typically through Base destruction), its HP drops. Reaching zero means the faction is in serious trouble.
- **Coin** — the in-game currency used to purchase assets and take certain actions. Income is earned each cycle based on Wealth rating.
- **XP** — accumulated experience points, spent to raise stats. Awarded on goal completion.
- **Scale** — minor, major, or hegemon. Scale sets hard caps on stat ratings and asset counts.
- **Assets** — units the faction owns (see below).
- **Bases of Influence** — footholds on worlds (see below).
- **ActiveGoal** — a multi-turn objective the faction is working toward. Tracked as an `ActiveGoal` struct holding goal ID, target, progress, and phase state.
- **Tags** — special abilities or traits that modify faction behavior (e.g., `Scavengers`, `Fanatical`).

### Asset

An `Asset` is a unit the faction owns — a starship, a military battalion, a spy ring, a financial front. Assets are the main interactive element of the game. They can attack, be attacked, be repaired, be moved, and be sold.

Each owned asset references an `AssetDefinition` from the Rulebook — a static template holding the asset's category, cost, HP, attack/counter stats, and any special ability. The owned `Asset` struct tracks runtime state: current HP, location, and three flags:

- **Ready** — newly purchased assets are not ready until the start of the next cycle.
- **Maintained** — assets require upkeep each cycle. A second consecutive missed payment destroys the asset.
- **Stealthy** — asset is hidden; cleared on certain interactions.

### Base of Influence

A `Base` represents a faction's foothold on a world. Bases are both territory and a proxy for faction health — damage dealt to a Base is also dealt directly to Faction HP. A faction's homeworld Base is special: its max HP is recalculated dynamically from the faction's current stat ratings.

### TurnState

`TurnState` tracks the in-progress turn. It records which cycle is active, the order factions will act in (randomized at turn start), which faction is currently up, and what phase that faction is in. `TurnState` is persisted to disk as part of `FactionState` so a turn is always resumable from wherever it left off.

### Mutation

Rather than mutating `FactionState` directly, every state change is expressed as a typed `Mutation` value. Actions and pipeline phases produce mutation lists; the `MutationEngine` is the single writer. Current mutation types: `CoinDelta`, `XPAwarded`, `XPSpent`, `AssetAdded`, `AssetRemoved`, `AssetHPDelta`, `AssetMaintainedFlag`, `AssetStealthCleared`, `FactionHPDelta`, `BaseHPDelta`, `BaseDestroyed`, `GoalProgressed`, `GoalCompleted`, `GoalAbandoned`, `GoalTurnsTick`, `HomeworldChanged`, `StatRaised`.

This pattern has two benefits. First, it makes history recording trivial — mutations are serializable and written to the history log as-is. Second, it keeps action logic clean and testable — an action produces a mutation list; it doesn't write state or know how persistence works.

### EventRecord

At the end of each pipeline write, all mutations are wrapped into an `EventRecord` and appended to a JSONL history file. The history file is append-only and machine-parseable. The `narrative` package reads it to produce markdown summaries.

---

## Static Data and the Rulebook

The game's asset catalog, faction tags, and goal templates are defined in TOML files under `internal/faction/data/`:

- `force_assets.toml`, `cunning_assets.toml`, `wealth_assets.toml` — all asset definitions
- `tags.toml` — tag definitions (ID, name, description, effect)
- `goals.toml` — goal definitions (ID, name, description, difficulty)

At startup, `rulebook.Load(dataDir)` reads all of these files and builds a `Rulebook` struct — maps of asset definitions, tags, and goals keyed by ID. The Rulebook is passed into the engine at construction time and is read-only for the entire lifetime of the process.

The data directory is not embedded in the binary. It lives with the campaign and is located via the `FACTION_DATA_DIR` environment variable. This means different campaigns can carry different rulesets.

---

## Persistence

Campaign state is one TOML file per campaign: `campaigns/<campaign-id>/faction_state.toml`. The `state` package is the only place that reads or writes this file.

TOML was chosen deliberately. The file is human-readable and hand-editable — a GM should be able to open the file and fix something if the tool gets into a bad state. There is no migration system; schema changes require manual edits to existing campaign files.

History is a JSONL file (`campaigns/<campaign-id>/history.jsonl`) that grows over time. It is never read back by the engine — it is a log for the `narrative` package and the GM's reference.

**Write cadences differ by purpose.** `applyAndRecord` writes history on every mutation batch — the JSONL log must be complete. `state.Save` is called only at phase-gate checkpoints (after bookkeeping, after action resolution, and after the turn cursor advances in `finishFactionTurn`). The state file is a resume checkpoint: its purpose is "if the process dies, where do we restart from?" Writing it after every intermediate mutation batch adds no resume value because `TurnState.Phase` already encodes which phase the turn is in. Do not merge these two write paths back together.

---

## Engine Architecture

The `Engine` struct in `internal/faction/engine/core.go` is the composition root. It owns:

- **Rulebook** — static game data, loaded once at startup
- **Rand** — `domain.Roller` interface, injected for deterministic dice in tests
- **Hooks** — the hook registry (see below)
- **Turn** — manages turn lifecycle (order, bookkeeping, phase advancement)
- **Mutation** — applies mutation lists to `FactionState`
- **Action** — action interface, factory, and resolution
- **Ability** — asset ability step handlers
- **History** — appends `EventRecord` to the JSONL file
- **Goal** — goal locking, progress tracking, XP logic

The sub-engines do not call each other. The `orchestrator.go` coordinates them in sequence across a well-defined pipeline.

### The InputCollector and TurnObserver Interfaces

The engine communicates with its caller through two interfaces.

**`TurnObserver`** is a fire-and-forget output channel. The observer is notified at each meaningful pipeline event — turn started, bookkeeping applied, action selected, action resolved, cycle completed — but cannot affect state. A CLI can use this to print output; a test can use it to record assertions.

**`InputCollector`** is the input abstraction. When the engine needs a decision from the GM — which action to take, whether to raise a stat, which ability to use — it calls a method on the collector and waits for the return value. This interface is the only place where the engine blocks on input. A TUI, a CLI prompt, an AI agent, or a scripted test harness all implement the same interface.

This design means the engine has no knowledge of how input is presented or collected. The goroutine/channel pattern from earlier versions of the codebase (needed when Bubbletea's event loop couldn't be blocked) is no longer needed.

### The Action Interface

Every action implements a three-method interface:

- `Validate(faction, factionState, rulebook)` — can this action run right now? Used to filter the menu before presenting it to the GM.
- `Inputs(faction, factionState, rulebook, collector)` — collect whatever the GM needs to provide.
- `Resolve(faction, factionState, rulebook)` — execute the action's logic, produce a mutation list.

Actions are stateful structs. Inputs collected in `Inputs()` are stored on the struct and read during `Resolve()`. Each action is registered as a factory in `engine/action/actions/register.go`. Adding a new action means: create a file in `actions/`, implement the interface, register the factory.

---

## The Turn Pipeline, End to End

`RunFactionTurn()` in `orchestrator.go` drives one faction's turn through five phases. `RunCycle()` loops over it until the cycle is complete.

**Phase 1 — Goal Lock Check.** `Goal.CheckLock()` evaluates the faction's active goal and returns a `GoalLock` with one of three types:
- `LockNone` — no restriction, proceed normally.
- `LockSkip` — the goal forces the faction to skip its action this turn (e.g., turn counter ticking).
- `LockRestrictActions` — only certain actions are allowed (e.g., Planetary Seizure restricts to Attack during its combat phase).

Any lock mutations are applied and the observer is notified. If `LockSkip`, the turn advances immediately.

**Phase 2 — Bookkeeping.** `Turn.ApplyBookkeeping()` calculates income (`floor(Wealth/2) + floor((Force+Cunning)/4)`), marks last cycle's purchases ready, and deducts asset maintenance costs. If Coin runs out mid-deduction, remaining assets are marked unmaintained; a second consecutive unmaintained cycle destroys the asset. All changes are emitted as mutations and applied before the action phase. The collector's `AwaitCheckpoint(bookkeeping)` pauses for the caller to display results.

**Phase 2B — Stat Raise (optional).** If the faction has enough XP to afford a stat increase, the orchestrator calls `collector.SelectStatRaise()`. If accepted, `XPSpent` and `StatRaised` mutations are applied.

**Phase 3 — Action Selection and Resolution.** `Action.AvailableActions()` calls `Validate()` on every registered action and returns those that pass. If a `LockRestrictActions` is in effect, the list is filtered further. The collector picks one via `SelectAction()`. The selected action's `Inputs()` and `Resolve()` run in sequence; `Resolve()` returns a mutation list.

**Phase 4 — Goal Progress and Hook Dispatch.** `Goal.UpdateProgress()` inspects the action's mutation list for goal-relevant events (kills, seizures, captures) and appends any goal progress or completion mutations. Then `dispatch.MutationReactors()` fires Cat 3 hooks (see below) in registration order, recursively, until no new mutations are produced (capped at depth 5).

**Phase 5 — Apply, Record, and Save.** The full ordered mutation list is applied to `FactionState`, written to the JSONL history file as a single `EventRecord`, and persisted to disk via `state.Save()`. The observer fires `OnActionResolved`; the collector awaits `CheckpointActionResult`. Then the turn cursor advances.

**Cycle Completion.** When the last faction's turn ends, `OnCycleCompleted` fires and the collector awaits `CheckpointCycleSummary`. `TurnState` is cleared.

---

## The Hooks System

The hooks system allows tags and other registered handlers to modify engine behavior without altering core logic. Hooks are registered at faction setup time (after the Rulebook is loaded) and are looked up from a scoped registry — global, per-faction, or per-asset.

Five hook categories, each with a distinct dispatch protocol:

- **Cat 1 — RollModifier.** Fires before a dice roll. Handlers can add bonus dice or flat modifiers. The collector selects which offers to apply (some may be elective).
- **Cat 2 — RollResultHook.** Fires after a roll result is known. Handlers can inspect and issue reroll directives; elective rerolls are confirmed by the collector.
- **Cat 3 — MutationReactor.** Fires after action resolution, given the full mutation list. Handlers return new mutations (e.g., `Scavengers` grants +1 Coin per enemy asset destroyed). Dispatch recurses until no new mutations are produced.
- **Cat 4 — Rule Modifiers.** A family of query-time interfaces (`AssetCostModifier`, `MaintenanceCostModifier`, `WorldTechLevelModifier`, `AssetMovementGranter`) that override specific rule calculations. Called synchronously during validation or bookkeeping.
- **Cat 5 — TieResolver.** Resolves attack/defense ties. The first registered handler wins.

Tag handlers register into this system. For example, `Warlike` registers a Cat 1 hook that adds a bonus die to attack rolls; `Scavengers` registers a Cat 3 reactor that grants Coin on kills.

---

## The Goal Engine

The goal engine in `engine/goal/` handles three responsibilities.

**Lock evaluation.** `CheckLock()` inspects the faction's `ActiveGoal` and returns a `GoalLock`. Goals with multi-turn timers tick their counter here and return `LockSkip` when the faction must wait. Goals with combat phases (e.g., Planetary Seizure) return `LockRestrictActions` with an explicit allowed-action list.

**Progress tracking.** `UpdateProgress()` receives the action's mutation list and pattern-matches against goal completion conditions. If the goal advances or completes, it appends the appropriate `GoalProgressed`, `GoalCompleted`, or `GoalTurnsTick` mutations.

**XP.** Goal completion appends an `XPAwarded` mutation. XP accumulates on the faction and is spent via `SelectStatRaise` to raise stats. Eleven specific goal types are handled; unlisted goals produce no progress mutations.

---

## Narrative Package

The `narrative` package in `internal/faction/narrative/` reads the JSONL history file and renders markdown summaries.

`LoadCycle(historyPath, cycleNumber)` parses the history file and returns all `EventRecord` entries for the given cycle.

The `digest` subpackage builds a `CycleDigest` from those records — per-faction beats (acquisitions, losses, goal events), cross-faction attack events, and a single headline (the most significant event: a homeworld attack, goal completion, or faction destruction).

`WireRenderer` implements the `Renderer` interface. Given a `CycleDigest` and a random seed, it renders a full markdown narrative with seeded template variants for natural-sounding prose. The `narrate` CLI command is the current consumer.

---

## CLI Layer

The current delivery mechanism is a pure CLI. There is no interactive TUI in the current codebase — the interactive turn-running TUI is the next major frontend milestone, deferred until the engine mechanics were fully proven in headless operation.

**What exists now:**

- **`faction create` / `faction delete`** — wizard-style faction management using `huh` forms. Multi-step selection of stats, starting assets, tags, and goal.
- **`narrate`** — reads the history file for a given cycle and renders a markdown narrative.
- **`review`** — placeholder for cycle state inspection.

The `huh` library is used for all interactive CLI prompts. `huh` forms run synchronously — build a form, call `.Run()`, read the result — and live in the Cobra command handlers.

**Planned TUI.** The interactive turn-running loop will be implemented as a Bubbletea TUI. It will implement `InputCollector` and `TurnObserver` and drive `RunCycle()` — the engine does not need to change. Sub-models for each turn phase (bookkeeping display, action selection, ability input, cycle summary) are the expected structure.

---

## Test Harness

`internal/faction/engine/testharness/` provides test infrastructure for action and pipeline tests.

- **`Harness`** — composes Engine, FactionState, Config, ScriptedCollector, and RecordingObserver into a single test fixture. `NewHarness(t, dataDir)` sets everything up.
- **`ScriptedCollector`** — implements `InputCollector` with pre-scripted selections. Tests feed in expected choices (action name, stat to raise, modifier decisions) and the collector returns them in order.
- **`RecordingObserver`** — implements `TurnObserver`. Records all observer calls in memory for assertion.
- **`scenarios/`** — reusable helper functions for building common faction/state configurations.

Tests use `go.uber.org/mock/gomock` for the `MockCollector` pattern where scripted selection alone isn't sufficient.

---

## Dev Workflow

**Build.** From `cmd/faction-manager/`:

```
go build -o bin/faction-manager .
```

**Run.** The binary needs to know where to find the static game data:

```
FACTION_DATA_DIR=/workspaces/gm-toolkit/internal/faction/data ./bin/faction-manager --campaign test <subcommand>
```

**Test.** From the repo root:

```
go test ./...
```

**Commits.** Conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`. Imperative mood, concise.

**Branches.** Feature branches off `main`. Before merging: code review, update `docs/dev_journals/faction-manager/dev-journal-factions.md` and `docs/dev_journals/faction-manager/decisions-log.md`, assess whether the work warrants a version bump.

**Docs.** Read `docs/` before implementing anything non-trivial. The CLAUDE.md routing table maps work areas to the relevant doc to read first.

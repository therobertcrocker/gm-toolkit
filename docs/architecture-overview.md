# GM Toolkit — Architecture Overview

This document is written for a mid-level Go engineer taking over active development of the gm-toolkit project. It assumes you're comfortable with Go, have built CLI tools before, and just need to understand how this project is structured and why.

---

## What This Project Is

GM Toolkit is a command-line tool for tabletop RPG game masters. Its current focus is the **faction tracker** for *Stars Without Number* (SWN), a sci-fi tabletop RPG. Factions are organizations — corporations, governments, pirate fleets, cults — that compete for influence between player sessions. The GM runs faction turns, during which each faction earns income, pays maintenance on its assets, and takes one action (attacking a rival, buying new units, expanding to a new world, etc.).

The tool is a single Go binary with no server component and no external dependencies beyond the local filesystem. Campaign state lives in TOML files on disk. The GM runs commands, fills out terminal prompts, and the tool tracks everything that would otherwise require spreadsheets or scribbled notes.

Understanding the game domain matters here. The codebase models real game mechanics — stat ratings, asset costs, attack formulas — so decisions in the code are often driven by rules in the rulebook (`docs/swn-faction-mechanics.md`). When something seems arbitrary, check there first.

---

## Repo Layout

```
gm-toolkit/
├── cmd/
│   └── faction-manager/        # The CLI application
│       ├── main.go             # Bootstrap: loads config, builds engine, registers commands
│       ├── config.go           # Reads FACTION_DATA_DIR from environment
│       ├── commands/           # One file (or subdirectory) per Cobra command
│       ├── tui/                # All Bubbletea UI code
│       │   ├── model.go        # Root TUI state machine
│       │   ├── collector.go    # TUICollector: bridges action inputs to the TUI
│       │   ├── narrate.go      # Human-readable mutation descriptions
│       │   ├── layout.go       # Split-panel layout helper
│       │   ├── eligibility.go  # TUI-side asset filtering helpers
│       │   ├── style/          # Lipgloss style definitions
│       │   ├── phases/         # Sub-models for turn phases (resume, bookkeeping, etc.)
│       │   └── inputs/         # Sub-models for action input collection
│       └── paths/              # File path resolution for campaign directories
├── internal/
│   └── faction/                # All faction domain and business logic
│       ├── domain/             # Pure data types — no I/O, no framework imports
│       ├── state/              # TOML persistence layer
│       ├── loader/             # Builds the Rulebook from static TOML files
│       ├── engine/             # Orchestration: turn lifecycle, actions, mutations, history
│       │   └── actions/        # One file per action type
│       └── data/               # Static game data (assets, tags, goals) as TOML files
├── docs/                       # Developer documentation and design journals
└── reference/                  # SWN rulebook excerpts and reference material
```

The split between `cmd/` and `internal/` is intentional and meaningful. Everything under `internal/faction/` is pure game logic — it knows nothing about terminals, prompts, or Cobra. Everything under `cmd/faction-manager/` is delivery mechanism — it knows about the terminal but delegates all game rules to the internal packages. This boundary makes the engine testable in isolation and means you can reason about game logic without thinking about UI.

---

## Core Domain Model

### Faction

A `Faction` is the central entity. It has:

- **Three stat pillars** — Force, Cunning, and Wealth, each rated 1–8. These determine what the faction can buy, how it fights, and how it earns income.
- **HP** — a health pool. When a faction takes damage (typically through Base destruction), its HP drops. Reaching zero means the faction is in serious trouble.
- **Coin** — the in-game currency used to purchase assets and take certain actions. Income is earned each cycle based on Wealth rating.
- **Scale** — minor, major, or hegemon. Scale sets hard caps on stat ratings and asset counts. A minor faction simply cannot field the same forces as a hegemon.
- **Assets** — units the faction owns (see below).
- **Bases of Influence** — footholds on worlds (see below).
- **Goal** — a multi-turn objective the faction is working toward. Goals earn XP when completed.
- **Tags** — special abilities or traits that modify faction behavior (e.g., "Colonists" or "Fanatical").

### Asset

An `Asset` is a unit the faction owns — a starship, a military battalion, a spy ring, a financial front. Assets are the main interactive element of the game. They can attack, be attacked, be repaired, be moved, and be sold.

Each owned asset references an `AssetDefinition` — a static template loaded from the Rulebook at startup. The definition holds the asset's category (Force/Cunning/Wealth), cost, HP, attack/counter stats, and any special ability. The owned `Asset` struct tracks the runtime state: current HP, location, readiness, whether maintenance was paid.

Two flags on Asset are worth knowing:

- **Ready** — newly purchased assets are not ready until the start of the next cycle. This prevents same-turn purchasing and attacking.
- **Maintained** — assets require upkeep each cycle. A faction that can't afford maintenance marks the asset unmaintained; a second consecutive missed payment destroys it.

### Base of Influence

A `Base` represents a faction's foothold on a world. Bases are both territory and a proxy for faction health — damage dealt to a Base is also dealt directly to Faction HP. This mechanic means that attacking a faction's Bases is a meaningful way to weaken the faction itself, not just its territory.

A faction's homeworld Base is special: its max HP is recalculated dynamically from the faction's current stat ratings, so improving the faction improves its homeworld resilience.

### TurnState

`TurnState` tracks the in-progress turn. It records which cycle is active, the order factions will act in (randomized at turn start), which faction is currently up, and what phase that faction is in (Bookkeeping → Action → Complete).

Critically, `TurnState` is persisted to disk as part of `FactionState`. If the program crashes mid-turn or the GM closes the terminal, the turn resumes exactly where it left off. The resumption prompt at turn start handles this case.

### Mutation

Rather than mutating `FactionState` directly in action code, every state change is expressed as a typed `Mutation` value — a `CoinDelta`, an `AssetAdded`, a `BaseHPDelta`, and so on. Actions collect these mutations and hand them off to the `MutationEngine`, which is the single writer to campaign state.

This pattern has two benefits. First, it makes history recording trivial — mutations are serializable, so they can be written to the history log as-is. Second, it keeps action logic clean and testable — an action just produces a list of mutations; it doesn't have to know how state is structured or saved.

### EventRecord

At the end of each faction's turn, all mutations are wrapped into an `EventRecord` and appended to a JSONL history file. Each line in the file is one complete record for one faction's turn in one cycle. The history file is append-only and meant for machine parsing (for a future narrative egine or post-turn report generator). The `EventRecord` struct captures all relevant data: faction ID, cycle number, and the full mutation list.

---

## Static Data and the Rulebook

The game's asset catalog, faction tags, and goal templates are defined in TOML files under `internal/faction/data/`. These files are the source of truth for what assets exist, what they cost, and what abilities they have. They're edited manually when adding new content to the game system.

At startup, `loader.LoadRulebook()` reads all of these files and builds a `Rulebook` struct — maps of asset definitions, tags, and goals keyed by ID. The Rulebook is passed into the engine at construction time and is read-only for the entire lifetime of the process. Nothing re-reads the data directory at runtime.

When you need to understand what an asset does or why an action has a particular validation rule, the Rulebook and the TOML files are the first place to look.

---

## Persistence

Campaign state is one TOML file per campaign: `campaigns/<campaign-id>/faction_state.toml`. The `state` package is the only place that reads or writes this file. All mutations flow through `MutationEngine.Apply()`, and after mutations are applied, the caller saves state via `state.Save()`.

TOML was chosen deliberately. The file is human-readable and hand-editable, which matters for a GM tool — a GM should be able to open the file and fix something if the tool gets into a bad state. There's no migration system; schema changes require manual edits to existing campaign files when they occur.

History is a JSONL file (`campaigns/<campaign-id>/history.jsonl`) that grows over time. It's never read back by the engine — it's a log for the GM's reference.

---

## Engine Architecture

The `Engine` struct in `internal/faction/engine/core.go` is the top-level orchestration object. It owns:

- **Rulebook** — the static game data, loaded once at startup
- **TurnEngine** — manages the turn lifecycle (start, advance, bookkeeping)
- **ActionEngine** — runs actions through their lifecycle and collects mutations
- **MutationEngine** — applies mutation lists to FactionState
- **AbilityEngine** — handles the special-ability logic for Use Asset Ability actions
- **HistoryEngine** — appends EventRecords to the history log

The sub-engines don't call each other. The TUI (specifically `TurnModel`) orchestrates calls across them in sequence: bookkeeping runs → mutations applied → action selected → inputs collected → action resolves → mutations applied → history recorded → state saved. The engine is the tool; the TUI is the hand holding it.

### The Action Interface

Every action in the game implements a four-method interface:

- `Validate(faction, state, rulebook)` — can this action run right now? Used to filter the action menu before presenting it to the GM.
- `Inputs(faction, state, rulebook)` — collect whatever input the GM needs to provide (which asset to buy, which faction to attack, etc.).
- `Resolve(faction, state, rulebook)` — execute the action's logic against the current state.
- `Output()` — return the list of mutations the action produced.

Actions are stateful structs. The inputs collected in `Inputs()` are stored on the struct and read during `Resolve()`. Each action is instantiated fresh for each use via a factory function registered with the ActionEngine.

This interface makes adding new actions straightforward: create a file in `engine/actions/`, implement the four methods, and register the factory. Nothing else needs to change.

---

## Action Resolution Pipeline, End to End

Here's what happens from the moment a GM starts a turn to the moment the cycle ends.

**Turn Start.** The TurnEngine creates a `TurnState`, assigns a randomized faction order (sorted IDs shuffled), and marks all assets `Ready = true`, making last cycle's purchases active. If a turn was already in progress (resumed from disk), this step is skipped.

**Bookkeeping.** For the active faction, the TurnEngine calculates income (`floor(Wealth/2) + floor((Force+Cunning)/4)`), deducts asset maintenance costs in order, and marks any asset unmaintained if Coin runs out. A second consecutive unmaintained cycle destroys the asset. All of these changes are emitted immediately as mutations and applied before the action phase begins.

**Action Selection.** The TUI calls `Validate()` on every registered action against the current faction and presents only the eligible ones. The GM picks one.

**Input Collection.** The selected action's `Inputs()` method runs. For simple actions, this might be a single prompt (which asset to sell?). For complex actions, it may involve multiple steps with conditional branching (which attackers? which defender? redirect to base?). Input is collected through the `InputCollector` interface — more on this below.

**Resolution.** `Resolve()` runs the action's logic. It reads the inputs collected in the previous step, consults faction state and rulebook rules, and computes outcomes (dice rolls, damage calculations, eligibility checks). It does not directly modify state — it builds up a mutation list.

**Mutation Application.** The MutationEngine applies the mutation list to `FactionState` in order. Each mutation type has a corresponding case in the engine that knows how to apply it.

**History Recording.** The TUI builds an `EventRecord` from the faction ID, cycle number, and full mutation list, then hands it to the HistoryEngine to append to the JSONL file.

**Advance.** The TurnEngine increments `CurrentIndex`. If this was the last faction, the cycle is marked complete, `TurnState` is cleared, and a cycle summary is displayed. Otherwise, the next faction's bookkeeping begins.

**Save.** After every mutation application, `state.Save()` writes the updated `FactionState` to disk. The turn is always resumable.

---

## The Goroutine/Channel Pattern for Mid-Resolution Input

This is the most non-obvious pattern in the codebase. Understanding it is essential for working on Attack, Expand Influence, or Use Asset Ability.

Simple actions (Sell Asset, Repair Asset) collect all their inputs upfront in `Inputs()`, then resolve without needing any more GM decisions. Complex actions need GM decisions *during* resolution — for example, the Attack action needs to know whether the defending faction wants to redirect damage to a Base instead of taking it directly to faction HP.

The problem is that `Resolve()` needs to pause mid-execution and wait for a GM response, but Bubbletea's event loop can't be blocked. The solution is a goroutine with channels.

For complex actions, `Resolve()` is launched in a goroutine. When it needs a mid-resolution GM decision, it sends a typed message on an event channel that the TUI is listening to, then blocks on a separate response channel. The TUI receives the message, transitions to an intermediate state (e.g., `stateAttackRedirect`), renders a confirmation prompt, and waits for a keypress. When the GM responds, the TUI sends the answer back on the response channel. The goroutine unblocks and continues.

When resolution is fully complete, the goroutine sends a completion message (`AttackCompletedMsg`, `ExpandInfluenceCompletedMsg`, etc.) containing the final mutation list. The TUI receives it, exits the intermediate state, and proceeds to history recording.

This is why `TurnModel` has so many states with "mid-resolution" in the conceptual name — `stateAttackRedirect`, `stateExpandInfluenceRivalConfirm`, `stateAbilityMoveDestination`, and so on. Each one corresponds to a point where a complex action paused and asked a question.

The `TUICollector` struct in `tui/collector.go` is the bridge. It implements the `InputCollector` interface and manages the event and response channels. Action code calls methods like `ConfirmRedirectToBase()` without knowing anything about Bubbletea — the TUICollector handles the channel handshake.

---

## TUI Layer

The TUI is built with [Bubbletea](https://github.com/charmbracelet/bubbletea), a framework that models the terminal UI as a pure function: given a model and a message, produce a new model and some commands. All UI state lives in the model; all updates flow through a single `Update()` function.

### TurnModel

`TurnModel` in `tui/model.go` is the root model for the turn flow. It's a state machine — a `turnState` integer constant determines what's currently displayed and what inputs are accepted. Each state corresponds to a screen or prompt:

- `stateResumePrompt` — should we resume, start fresh, or abandon?
- `stateBookkeeping` — show bookkeeping results
- `stateActionSelect` — GM chooses an action
- `stateActionInput` — a sub-model collects action-specific inputs
- `stateActionResult` — display what happened
- Several mid-resolution states for complex actions
- `stateCycleSummary` — end-of-cycle report
- `stateDone` — exit

Transitions happen via messages. When a sub-model finishes its job, it emits a typed message (e.g., `ActionSelectedMsg`, `BuyOrderSelectedMsg`). `TurnModel.Update()` receives the message, stores any relevant data, and transitions to the next state.

### Sub-Models: Phases and Inputs

Sub-models live in two directories:

- `tui/phases/` — models for turn phase screens (resume prompt, skip prompt, bookkeeping display, action selection, cycle summary). These are mostly display + simple selection.
- `tui/inputs/` — models for action input collection (buy order, attack inputs, repair selection, etc.). These can be multi-step: select asset, then select target, then confirm.

Each sub-model is a self-contained Bubbletea model. `TurnModel` holds the active sub-model as a field and delegates `Update()` and `View()` calls to it while that phase is active.

### Huh Forms

For simpler CLI flows that don't need the full Bubbletea loop — faction creation, faction deletion — the project uses [huh](https://github.com/charmbracelet/huh), a form library from the same family as Bubbletea. Huh forms run synchronously: build a form, call `.Run()`, read the result. These live in the Cobra command handlers rather than the TUI layer.

### Layout and Styles

The turn TUI uses a two-panel layout: a narrow left panel (fixed at 30 characters) showing the active faction's stats, and a wider right panel showing the current phase content. The `renderSplitPanel()` helper in `tui/layout.go` handles height-balancing the two panels so borders align.

All lipgloss styles are defined once in `tui/style/styles.go` and referenced by name everywhere else. If you're adding new output, use existing styles before defining new ones — consistency matters for readability in a terminal tool.

---

## Dev Workflow

**Build.** Work from `cmd/faction-manager/` as the working directory:

```
go build -o bin/faction-manager .
```

**Run.** The binary needs to know where to find the static game data:

```
FACTION_DATA_DIR=/workspaces/gm-toolkit/internal/faction/data ./bin/faction-manager --campaign test turn
```

The `--campaign` flag selects which campaign directory to use. The test campaign lives at `cmd/faction-manager/campaigns/test/faction_state.toml`.

**Test.** Run from the repo root:

```
go test ./...
```

Tests cover engine logic, mutation application, and action validation. The `domain.Roller` interface is injected for deterministic dice in tests.

**Commits.** Use conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`. Imperative mood, concise. Example: `feat: add Refit Asset action`.

**Branches.** Feature branches off `main`. Before merging: code review, update `docs/dev-journal-factions.md` and `docs/decisions-log.md`, assess whether the work warrants a version bump. The current version is tracked in `CLAUDE.md`.

**Docs.** Read `docs/` before implementing anything non-trivial. The dev journal records why decisions were made; the decisions log records what was ratified. If you make a meaningful architectural choice, it belongs in one or both.

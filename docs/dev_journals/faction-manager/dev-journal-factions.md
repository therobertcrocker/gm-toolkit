# Dev Journal: Faction Tracker

Entry point for the faction tracker's design documentation. The two living sections are **Feature List** and **Open Questions** — updated as features land and design questions arise.

<br/>

## Quick Reference

| Document | Purpose |
|----------|---------|
| [Architecture Overview](../architecture-overview.md) | System shape, domain model, turn pipeline, and package layout |
| [Planned Work](planned-work.md) | Pre-discovery initiative tracker; all known deferred work |
| [Decisions Log](decisions-log.md) | All ratified implementation decisions, numbered and grouped by branch |
| [Contributor Guide](../contributor-guide.md) | Setup, workflow, and conventions for development |
| [Implementation Plans](../implementation/) | Per-feature implementation plans, written before execution |
| [Discovery Docs](../discovery/) | Pre-implementation design explorations |

<br/>

## Modes

### Review Mode (read-only)

The GM can browse the current state of the campaign without making any changes:

- View faction summaries (stats, HP, Coin, current goal)
- View asset lists by faction or by world
- Browse past turn history and event logs
- Generate a human-readable narrative of past turns

### Turn Mode (stateful, step-by-step)

The GM executes a new faction turn:

- Factions are ordered randomly at the start of the turn
- The GM steps through each faction one at a time
- For each faction: income is calculated, maintenance is paid, and an action is chosen and processed
- State is updated after each action; history is recorded
- A turn can be paused mid-execution and resumed later

### Edit Mode (Planned)

The GM makes freeform changes to faction state outside of mechanical rules — for campaign setup, world-building, or corrections:

- Edit faction details (name, homeworld, scale, stats, HP, Coin, goal, tags)
- Add, remove, or modify assets directly
- No mechanical validation; changes are applied as entered

<br/>

## AI Decision-Making (Planned)

The tool is designed from the start to support automated faction behavior. The intent is traditional game AI (not LLM-based) that can drive faction decisions during Turn Mode. Two modes are planned:

- **Goal-oriented** — the AI plays optimally toward the faction's current goal
- **In-character** — the AI factors in faction personality and tags when making decisions

The action system's source-agnostic design means the AI and the GM use the same interface.

<br/>

## Design Principles

- **CLI-first** — core logic is decoupled from any UI layer
- **Source-agnostic actions** — actions can be driven by human input or AI equally
- **Separation of concerns** — review and turn execution are distinct flows
- **Extensible static data** — GMs customize assets and flavor without touching code
- **Robust history** — every state change is recorded; nothing is lost
- **Built to grow** — a TUI or web frontend can be layered on later without touching core logic

<br/>

## Feature List

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 1 | Faction CRUD | Complete | `faction create`, `list`, `delete` via Cobra commands |
| 2 | Turn Engine | Complete | Bookkeeping, stat raise, faction ordering, mid-turn resume |
| 3 | All 9 SWN Actions | Complete | Attack, BuyAsset, SellAsset, RepairAsset, RepairFaction, RefitAsset, ExpandInfluence, Bribe, UseAssetAbility |
| 4 | Goal Engine | Complete | 12 goals, multi-turn objectives, XP rewards, `CheckLock` and `UpdateProgress` pipeline |
| 5 | Event Hooks System | Complete | Five-category hook system, scoped registry, `RollWithHooks` |
| 6 | Tag Automation | In progress | 4 tags live (Scavengers, Warlike, Fanatical, Preceptor Archive); 16 remaining + Effects Engine (`S`-flag assets) |
| 7 | Narrative Renderer | Complete | `narrate <cycle>` command; markdown output from JSONL history |
| 8 | Test Suite | Complete | Unit + integration tests; `testharness` with `ScriptedCollector`, `RecordingObserver`, `FixedRoller` |
| 9 | TUI | Not started | Fresh Bubbletea TUI against `InputCollector` + `TurnObserver` interfaces |
| 10 | Edit Mode | Not started | Freeform state manipulation outside turn rules; see Modes above |
| 11 | AI Decision-Making | Not started | Goal-oriented and in-character modes; see above |
| 12 | Spatial Model | In progress | Effort 1 Phase 1 complete — standalone `internal/spatial` package: SpatialMap/Location interfaces, HybridMap with TOML loader and Dijkstra distance, HexMap/GraphMap stubs. Effort 2 (engine integration: index/scan replacement, validation, distance-based mechanics) ahead |

<br/>
<br/>

# Open Questions

Design questions that are unresolved and will need answers before the relevant feature can be built.

| # | Question | Relevant Feature |
|---|----------|-----------------|
| 1 | Starting Coin for new factions — no explicit SWN rule; what is the right default (Wealth rating, fixed amount, GM prompt)? | `faction create`, Edit Mode |
| 2 | How does Edit Mode integrate into the Cobra command tree — sub-commands under a top-level `edit` command, or per-entity sub-commands (e.g. `faction edit`)? | Edit Mode |
| 3 | When AI decision-making is added, how does the GM designate which factions are AI-driven vs. manually controlled? | AI decision-making |

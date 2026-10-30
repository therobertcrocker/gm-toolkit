# Turn Mode TUI — Discovery & Design

A planning artifact for the Bubbletea TUI that replaces the `turn` command's huh wizard. Covers architecture, state machine, component design, and the strategy for bridging Bubbletea's event-driven model to the existing action engine.

> Related: [turn-engine-discovery.md](turn-engine-discovery.md)

<br/>

## 1. Overview and Scope

The TUI replaces the sequential huh-based wizard in `cmd/faction-manager/commands/turn/wizard.go` with a full-screen Bubbletea interface. It covers all phases of a faction turn: resume detection, per-faction bookkeeping, action selection, action input collection, and the end-of-cycle summary.

### What changes

- `cmd/faction-manager/tui/` — new package; the Bubbletea program lives here
- `cmd/faction-manager/commands/turn/wizard.go` — deleted; replaced by `tui.RunTurnTUI`
- `cmd/faction-manager/forms/collector.go` (`GMCollector`) — deleted; replaced by `TUICollector`
- `cmd/faction-manager/commands/turn/cmd.go` — updated to call `tui.RunTurnTUI(e, factionState, p)`

### What does not change

- All engine sub-systems: `TurnEngine`, `ActionEngine`, `MutationEngine`, `HistoryEngine`
- The `InputCollector` interface in `internal/faction/engine/input_collector.go` — same 7 method signatures
- The `Action` interface and all concrete actions in `engine/actions/`
- All domain types, state serialization, and history recording
- All non-turn commands: `faction create`, `faction list`, `faction delete`, and the `review` group stay as Cobra huh-wizard commands

### Dependencies

All three packages are already present in `go.mod` as indirect dependencies; they become direct:

- `charmbracelet/bubbletea` — the TUI runtime (event loop, model/update/view)
- `charmbracelet/bubbles` — reusable components (`list.Model` for all select inputs)
- `charmbracelet/lipgloss` — styling and layout (previously deferred to this layer in decision #44)
- `charmbracelet/huh` — remains for non-turn commands

<br/>

## 2. Layout

The terminal is divided into two persistent panels for the duration of a faction's turn.

```
┌──────────────────────────┬──────────────────────────────────────────────┐
│  FACTION NAME             │  Phase Title                                  │
│  Scale · HP 14/18         │                                               │
│  Coin: 32                 │  (Right panel content changes per phase)      │
│  Goal: Seize Planet       │                                               │
│                           │                                               │
│  Stats                    │                                               │
│  Force   4                │                                               │
│  Cunning 3                │                                               │
│  Wealth  5                │                                               │
│                           │                                               │
│  Assets                   │                                               │
│  War Fleet       F4  12/12│                                               │
│  Informers       C2   6/8 │                                               │
│  Merchant Caravan W3   4/4│                                               │
│                           │                                               │
│  Bases                    │                                               │
│  Avernus (home)  HP 18/18 │                                               │
└──────────────────────────┴──────────────────────────────────────────────┘
```

### Left panel

Always visible during a faction's turn. Content:
- Faction name (styled header), scale, HP (current/max), Coin
- Force, Cunning, Wealth ratings
- Active goal name
- Asset list: name, category+rating abbreviation (e.g. F4), HP, status flags (not ready, unmaintained, stealthy)
- Bases list: location, HP, homeworld flag

The left panel does not update mid-turn as mutations are applied. It reflects the state at the start of the faction's turn and is refreshed at the start of the next faction's turn.

### Right panel

Fills the remainder of the terminal width. Content varies by state (see state machine). All interactive elements (lists, prompts, confirmations) live here.

### Full-screen states (no split)

The resume prompt and cycle summary fill the entire terminal width — there is no active faction to show in the left panel during these states.

<br/>

## 3. State Machine

`TurnModel` is the top-level `tea.Model`. It owns the current state and delegates rendering and input to embedded sub-models.

### States

| State | Right panel content | Trigger to advance |
|-------|--------------------|--------------------|
| `stateResumePrompt` | Resume / Start New / Abandon choice | User selects |
| `stateSkipPrompt` | "Skip [Faction]'s turn?" yes/no | User selects |
| `stateBookkeeping` | Income breakdown, maintenance results, assets lost/unmaintained | Any key |
| `stateActionSelect` | Filtered action list + No Action option | User selects |
| `stateActionInput` | Action-specific input sub-model | Sub-model signals done |
| `stateActionResult` | One-line summary of what the action did | Any key |
| `stateCycleSummary` | Table: all factions, HP, Coin, action taken | Any key → quit |
| `stateDone` | — | `tea.Quit` returned |

### Transitions

```
start
  └→ stateResumePrompt
       ├→ (resume or start new) → stateSkipPrompt [faction 0]
       └→ (abandon) → stateDone

stateSkipPrompt
  ├→ (skip) → advance faction → stateSkipPrompt [next] or stateCycleSummary [done]
  └→ (no skip) → ApplyBookkeeping → stateBookkeeping

stateBookkeeping
  └→ (any key) → stateActionSelect

stateActionSelect
  ├→ (No Action) → commit + history → advance → stateSkipPrompt [next] or stateCycleSummary
  └→ (action selected) → stateActionInput

stateActionInput
  └→ (sub-model done) → Run action → stateActionResult

stateActionResult
  └→ (any key) → commit + history → advance → stateSkipPrompt [next] or stateCycleSummary

stateCycleSummary
  └→ (any key) → stateDone
```

### Commit and history

State save and history recording happen when advancing past `stateActionResult` (or past the skip path), matching the existing per-faction commit granularity. This preserves pause/resume correctness: each faction's data is persisted before the next faction's bookkeeping runs.

<br/>

## 4. Component Map

New package: `cmd/faction-manager/tui/`

| File | Responsibility |
|------|---------------|
| `model.go` | `TurnModel` struct; `Init`, `Update`, `View`; `RunTurnTUI` entry point |
| `layout.go` | Lipgloss helpers: left/right panel split, borders, sizing from terminal width |
| `styles.go` | Lipgloss style constants: colors for HP, Coin, headers, muted text, borders |
| `messages.go` | All `tea.Msg` types used across sub-models |
| `phases/resume_prompt.go` | `ResumeTurnModel` — resume / start new / abandon |
| `phases/skip_prompt.go` | `SkipTurnModel` — skip this faction? |
| `phases/bookkeeping.go` | `BookkeepingModel` — renders `engine.BookkeepingResult` |
| `phases/action_select.go` | `ActionSelectModel` — wraps `bubbles/list.Model`; emits `ActionSelectedMsg` |
| `phases/summary.go` | `CycleSummaryModel` — end-of-cycle recap table |
| `inputs/sell_asset.go` | `SelectAssetModel` — single-select from faction assets |
| `inputs/buy_asset.go` | `BuyOrderModel` — world select → asset definition select |
| `inputs/refit_asset.go` | `RefitOrderModel` — owned asset select → replacement select |
| `inputs/repair_asset.go` | `RepairOrdersModel` — multi-select assets + per-asset heal count |
| `inputs/attack_inputs.go` | `AttackInputsModel` — sequential: attackers → defenders → redirect confirms |
| `collector.go` | `TUICollector` — `InputCollector` implementation backed by pre-collected data |

<br/>

## 5. InputCollector Bridge

### The problem

`action.Inputs()` calls `collector.Select*()` synchronously and blocks until the user chooses. Inside a Bubbletea `Update` function, blocking is not possible — the event loop must return immediately.

### Solution: pre-collection pattern

1. When the user selects an action in `stateActionSelect`, `TurnModel` knows the action type
2. `TurnModel` transitions to `stateActionInput` and embeds the appropriate input sub-model for that action
3. The sub-model collects all required inputs through Bubbletea views (arrow keys, Enter, etc.)
4. When collection is complete, the sub-model sends a typed `tea.Msg` back to `TurnModel` with the collected data
5. `TurnModel` builds a `TUICollector` instance with the pre-collected data stored in its fields
6. `TurnModel` calls `e.Action.Run(action, faction, factionState, rulebook)` — this calls `action.Inputs()`, which calls `collector.Select*()`, which returns the pre-stored data immediately (no prompt, no blocking)
7. `action.Resolve()` and `action.Output()` run synchronously as before

The engine call sequence — `Inputs` → `Resolve` → `Output` — is unchanged. The collector simply returns pre-collected values instead of prompting.

### `TUICollector` fields

```
selectedAsset     *domain.Asset                 — SellAsset
buyOrder          engine.BuyOrder               — BuyAsset
refitOrder        engine.RefitOrder             — RefitAsset
repairOrders      []engine.RepairOrder          — RepairAsset
attackers         []*domain.Asset               — Attack
defenders         map[string]*domain.Asset      — Attack (attacker ID → defender)
redirectConfirms  map[string]bool               — Attack (matchup key → bool)
```

Each `InputCollector` method reads exactly one field and returns it. If the wrong method is called for a given action, the field will be its zero value — this would indicate a programming error, not a runtime user error.

<br/>

## 6. Action Input Sub-Models

Each sub-model is a self-contained `tea.Model`. `TurnModel` embeds the active sub-model during `stateActionInput`, delegates `Update` and `View` to it, and listens for a completion message.

### `SelectAssetModel` — SellAsset

- Wraps `bubbles/list.Model` populated with the faction's assets
- Each list item shows asset name, category, HP, and location
- Enter selects; sends `AssetSelectedMsg{asset *domain.Asset}` to parent

### `BuyOrderModel` — BuyAsset

Two-step sequential model:
1. `bubbles/list` of worlds where the faction has a Base of Influence (or homeworld)
2. `bubbles/list` of purchasable asset definitions at the selected world (pre-filtered: attribute meets MinRating, Coin ≥ Cost, TechLevel met)
- Sends `BuyOrderSelectedMsg{order engine.BuyOrder}` when both steps complete

### `RefitOrderModel` — RefitAsset

Two-step sequential model:
1. `bubbles/list` of owned assets that have at least one valid replacement (pre-filtered)
2. `bubbles/list` of valid replacement definitions for the selected asset (same category, different def, attribute ≥ MinRating, Coin ≥ max(0, newCost-oldCost))
- Sends `RefitOrderSelectedMsg{order engine.RefitOrder}`

### `RepairOrdersModel` — RepairAsset

Multi-step model:
1. `bubbles/list` multi-select: which damaged assets to repair (Space toggles selection, Enter confirms)
2. For each selected asset in order: display asset name + current/max HP; +/- keys (or number input) to set heal count; Enter confirms
- Sends `RepairOrdersSelectedMsg{orders []engine.RepairOrder}`

### `AttackInputsModel` — Attack

Sequential state machine (most complex input):
1. `bubbles/list` multi-select: which eligible assets to attack with (Space toggles, Enter confirms)
2. For each selected attacker in order: `bubbles/list` single-select defender from eligible enemy assets on the same world
3. For each attacker–defender matchup where the defender's faction has a Base on that world: yes/no confirm redirect
- Sends `AttackInputsSelectedMsg{attackers []*domain.Asset, defenders map[string]*domain.Asset, redirectConfirms map[string]bool}`

### RepairFaction — no sub-model

`RepairFaction` has no `InputCollector` dependency; `action.Inputs()` is a no-op. `TurnModel` can call `e.Action.Run` immediately upon selection.

<br/>

## 7. Bubbletea Patterns

### Model delegation

Sub-models receive `tea.Msg` by delegation from the parent's `Update`. Each sub-model's `Update` returns a `(tea.Model, tea.Cmd)`. The parent stores the updated sub-model and propagates the `tea.Cmd`. When a sub-model is complete, it returns a `tea.Cmd` that produces a completion `tea.Msg`; the parent's `Update` catches this message and transitions state.

### List component

`bubbles/list.Model` is used for all single and multi-select inputs. It handles arrow key navigation, wrapping, and optionally item filtering. Each list item implements `list.Item` (two methods: `Title() string`, `Description() string`). Multi-select behavior requires tracking selection state separately since `bubbles/list` is single-select by default — a selected-items map is maintained alongside the list model.

### Layout

`lipgloss.JoinHorizontal(lipgloss.Top, leftContent, rightContent)` composes the two panels. Panel widths are computed from the terminal width reported by `tea.WindowSizeMsg`. The left panel uses a fixed width (approximately 28 columns); the right panel takes the remainder.

### Key bindings

- Arrow keys / `j` / `k` — navigate lists
- `Space` — toggle selection in multi-select views
- `Enter` — confirm selection or advance phase
- `q` / `Ctrl+C` — quit (handled at top-level `TurnModel`)

<br/>

## 8. Design Decisions

### Feature/tui

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | Pre-collection pattern for `InputCollector` bridge | Avoids goroutines and channels; engine calls remain synchronous; `TUICollector` is a plain data struct |
| 2 | `TUICollector` replaces `GMCollector`; `InputCollector` interface unchanged | Interface is the designed seam; no engine changes needed; AI collector plugs in the same way |
| 3 | `GMCollector` and `cmd/forms` package deleted | Sole consumer is the turn wizard, which is replaced entirely |
| 4 | `bubbles/list` for all select inputs | Already in `go.mod`; handles navigation and wrapping; no new dependency |
| 5 | `huh` retained for non-turn commands | `faction create` wizard is out of scope; removing `huh` would require a parallel rewrite |
| 6 | Input sub-models are embedded `tea.Model` values | Consistent with Bubbletea patterns; each input flow is isolated and independently testable |
| 7 | Left panel reflects state at start of faction's turn (not live-updated) | Avoids mid-turn partial state display; faction context is stable for the entire action sequence |
| 8 | Full-screen layout for resume prompt and cycle summary | No active faction during these states; the left panel would be empty and misleading |

**Notable alternatives rejected:**

| Decision # | Alternative | Why rejected |
|---|---|---|
| #1 | Channel-based async bridge: goroutine runs `action.Inputs`, channels carry request/response | Concurrency adds complexity with no user-visible benefit; pre-collection is simpler and provably correct |
| #6 | Single monolithic `TurnModel.Update` with all input logic inline | Unmanageable as action count grows; one switch with nested switch per action state is unreadable |

<br/>

## Open Questions

| # | Question | Relevant Feature |
|---|----------|-----------------|
| 1 | Should `stateActionResult` display a structured diff of mutations (e.g. asset removed, Coin delta) or just the action name and a one-line summary? | Action result display |
| 2 | Should the cycle summary include factions that were skipped, and if so, how are they marked? | Cycle summary |

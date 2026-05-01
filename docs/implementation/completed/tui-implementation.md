# TUI Implementation Plan

A phased build guide for the `feature/tui` branch. Written for a fresh session with no prior context — read this after `docs/discovery/tui-discovery.md`.

<br/>

## Context

This plan implements the Bubbletea TUI for Turn Mode, replacing the existing huh wizard with a full-screen split-panel interface. The engine layer is entirely unchanged — only the presentation and input-collection layers change.

### Key files to read before starting

| File | Why |
|------|-----|
| `docs/discovery/tui-discovery.md` | Full architecture: state machine, component map, layout, InputCollector bridge |
| `cmd/faction-manager/commands/turn/wizard.go` | The code being replaced — understand the full turn flow before writing the TUI |
| `internal/faction/engine/input_collector.go` | The interface `TUICollector` must implement (7 methods) |
| `internal/faction/engine/action.go` | `Action` interface — `Inputs`, `Resolve`, `Output` |
| `internal/faction/engine/core.go` | `Engine` struct — what sub-engines are available |
| `internal/faction/engine/turn_engine.go` | `TurnEngine` methods and `BookkeepingResult` struct |
| `internal/faction/engine/action_engine.go` | `AvailableActions` and `Run` signatures |
| `cmd/faction-manager/forms/collector.go` | `GMCollector` — reference for what each `InputCollector` method does today |

### Dependency versions (already in go.mod as indirect)

Run `go get` to promote to direct before starting:

```
go get github.com/charmbracelet/bubbletea@v1.3.6
go get github.com/charmbracelet/bubbles@v0.21.1-0.20250623103423-23b8fd6302d7
go get github.com/charmbracelet/lipgloss@v1.1.0
```

<br/>

## Phase 1 — Scaffold the TUI package

**Goal:** `tui.RunTurnTUI` is callable from the turn command and compiles cleanly.

### Files to create

**`cmd/faction-manager/tui/styles.go`**

Define a `styles` package-level var or a `Styles` struct with lipgloss style constants:
- `HeaderStyle` — bold, faction name header
- `MutedStyle` — dim text for secondary info
- `HPStyle` / `LowHPStyle` — green/red HP values
- `CoinStyle` — yellow Coin values
- `PanelBorderStyle` — border for left and right panels
- `SectionTitleStyle` — bold for section labels (Assets, Bases, etc.)

**`cmd/faction-manager/tui/messages.go`**

Define all `tea.Msg` types used across sub-models:

```go
type ResumeChoiceMsg    struct{ choice resumeChoice }  // resume | startNew | abandon
type SkipChoiceMsg      struct{ skip bool }
type BookkeepingDoneMsg struct{}
type ActionSelectedMsg  struct{ index int }            // -1 = No Action
type AssetSelectedMsg   struct{ asset *domain.Asset }
type BuyOrderSelectedMsg   struct{ order engine.BuyOrder }
type RefitOrderSelectedMsg struct{ order engine.RefitOrder }
type RepairOrdersSelectedMsg struct{ orders []engine.RepairOrder }
type AttackInputsSelectedMsg struct {
    attackers       []*domain.Asset
    defenders       map[string]*domain.Asset
    redirectConfirms map[string]bool
}
type ActionResultDoneMsg struct{}
type SummaryDoneMsg      struct{}
```

**`cmd/faction-manager/tui/layout.go`**

```go
func renderSplitPanel(left, right string, totalWidth int) string
```

Uses `lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)`. Left panel is fixed width (~28 cols with border); right panel takes the remainder. Both panels are wrapped in `PanelBorderStyle` with the same height (pad shorter side to match).

**`cmd/faction-manager/tui/model.go`**

```go
type turnState int
const (
    stateResumePrompt turnState = iota
    stateSkipPrompt
    stateBookkeeping
    stateActionSelect
    stateActionInput
    stateActionResult
    stateCycleSummary
    stateDone
)

type TurnModel struct {
    engine       *engine.Engine
    factionState *state.FactionState
    paths        paths.Paths
    width        int
    height       int
    state        turnState
    currentFaction *domain.Faction
    bookkeepingResult engine.BookkeepingResult
    pendingMutations  []domain.Mutation
    actionResultText  string
    subModel     tea.Model  // active sub-model during input/prompt states
}

func (m TurnModel) Init() tea.Cmd
func (m TurnModel) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m TurnModel) View() string

func RunTurnTUI(e *engine.Engine, factionState *state.FactionState, p paths.Paths) error
```

`RunTurnTUI` creates a `TurnModel`, runs `tea.NewProgram(model, tea.WithAltScreen())`, and returns any error.

`Init` initializes the resume prompt sub-model.

`Update` handles `tea.WindowSizeMsg` (store width/height), delegates to `subModel.Update`, and catches completion messages to drive state transitions.

`View` calls `subModel.View()` for full-screen states (resume, summary) or calls `renderSplitPanel(renderLeft(m), subModel.View(), m.width)` for faction-turn states.

`renderLeft(m TurnModel) string` — renders the left panel content from `m.currentFaction`.

**Verification:** `go build ./...` passes from repo root. Wire the stub `RunTurnTUI` into `cmd/faction-manager/commands/turn/cmd.go` (keep old wizard call intact for now, just ensure it compiles alongside).

<br/>

## Phase 2 — Resume and skip prompts

**Goal:** Launching `turn --campaign test` shows the resume/start prompt; skip prompt appears per faction.

### Files to create

**`cmd/faction-manager/tui/phases/resume_prompt.go`**

```go
type resumeChoice int
const (
    choiceResume resumeChoice = iota
    choiceStartNew
    choiceAbandon
)

type ResumeTurnModel struct {
    list     list.Model
    inProgress bool
}

func NewResumeTurnModel(inProgress bool) ResumeTurnModel
func (m ResumeTurnModel) Init() tea.Cmd
func (m ResumeTurnModel) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m ResumeTurnModel) View() string
```

If `inProgress` is false, the list shows only "Start New Turn" (no resume or abandon options). On Enter, sends `ResumeChoiceMsg`.

**`cmd/faction-manager/tui/phases/skip_prompt.go`**

```go
type SkipTurnModel struct {
    factionName string
    selected    bool
}

func NewSkipTurnModel(factionName string) SkipTurnModel
func (m SkipTurnModel) Init() tea.Cmd
func (m SkipTurnModel) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m SkipTurnModel) View() string
```

Renders "Skip [name]'s turn? (y/n)" or a two-item list. On choice, sends `SkipChoiceMsg{skip bool}`.

### Wire into `TurnModel.Update`

On `ResumeChoiceMsg`:
- `choiceResume`: call `e.Turn` — cycle already in progress, load `currentFaction` from `e.Turn.CurrentFaction`
- `choiceStartNew`: call `e.Turn.Start(factionState)`, load `currentFaction`
- `choiceAbandon`: call `e.Turn.Abandon(factionState)`, save state, transition to `stateDone`

After resume/start: transition to `stateSkipPrompt`, embed `NewSkipTurnModel(currentFaction.Name)`.

On `SkipChoiceMsg`:
- skip=true: call `e.Turn.Advance(factionState)`, save state, record empty history event; if done → `stateCycleSummary`, else load next faction, stay in `stateSkipPrompt`
- skip=false: transition to `stateBookkeeping`

**Verification:** Run `turn --campaign test`. Resume/start prompt appears. Selecting start new shows skip prompt for faction 0. Selecting skip advances to faction 1.

<br/>

## Phase 3 — Bookkeeping display

**Goal:** Bookkeeping results render in the right panel; any key advances to action selection.

### Files to create

**`cmd/faction-manager/tui/phases/bookkeeping.go`**

```go
type BookkeepingModel struct {
    result engine.BookkeepingResult
    rulebook *loader.Rulebook
}

func NewBookkeepingModel(result engine.BookkeepingResult, rulebook *loader.Rulebook) BookkeepingModel
func (m BookkeepingModel) Init() tea.Cmd
func (m BookkeepingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m BookkeepingModel) View() string
```

`View` renders:
- Income line: `Income: +N Coin (Wealth N/2 = N, Stats (F+C)/4 = N)`
- Maintenance section: one line per asset — asset name, cost (0 if current system has no cost data), maintained/not maintained
- Assets lost section (if any): asset name + "lost (unpaid 2 turns)"
- Assets unmaintained (if any): asset name + "first missed payment"
- "Press any key to continue" footer

`Update` returns `BookkeepingDoneMsg` on any `tea.KeyMsg`.

### Wire into `TurnModel.Update`

On transition from `stateSkipPrompt` (skip=false):
- Call `e.Turn.ApplyBookkeeping(factionState)` — store `BookkeepingResult` in `m.bookkeepingResult`
- Store `result.RecordedMutations` in `m.pendingMutations` (bookkeeping mutations carried for history)
- Transition to `stateBookkeeping`, embed `NewBookkeepingModel(result, e.Rulebook)`

On `BookkeepingDoneMsg`: transition to `stateActionSelect`.

**Verification:** Skip prompt → no skip → bookkeeping screen shows income and maintenance. Left panel shows current faction info. Any key advances to action selection (stub for now).

<br/>

## Phase 4 — Action selection

**Goal:** Filtered action list renders in the right panel; selecting an action or No Action transitions correctly.

### Files to create

**`cmd/faction-manager/tui/phases/action_select.go`**

```go
type ActionSelectModel struct {
    list    list.Model
    actions []engine.Action  // parallel to list items; index -1 appended for No Action
}

func NewActionSelectModel(actions []engine.Action) ActionSelectModel
func (m ActionSelectModel) Init() tea.Cmd
func (m ActionSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m ActionSelectModel) View() string
```

List items: one per available action (from `e.Action.AvailableActions`) + "No Action" appended. Arrow keys navigate; Enter sends `ActionSelectedMsg{index}` where index is the position in `actions` slice, or -1 for No Action.

### Wire into `TurnModel.Update`

On `ActionSelectedMsg`:
- index == -1 (No Action): commit + record history with only bookkeeping mutations; call `e.Turn.Advance`; if done → `stateCycleSummary`, else load next faction → `stateSkipPrompt`
- index >= 0: store the selected action; transition to `stateActionInput` (handled in Phase 5+)

**Verification:** Action list renders with all valid actions for the current faction. No Action works end-to-end (completes faction, advances to next).

<br/>

## Phase 5 — TUICollector and simple action inputs

**Goal:** SellAsset, BuyAsset, and RefitAsset complete a full cycle end-to-end.

### Files to create

**`cmd/faction-manager/tui/collector.go`**

```go
type TUICollector struct {
    selectedAsset    *domain.Asset
    buyOrder         engine.BuyOrder
    refitOrder       engine.RefitOrder
    repairOrders     []engine.RepairOrder
    attackers        []*domain.Asset
    defenders        map[string]*domain.Asset
    redirectConfirms map[string]bool
}

func (c *TUICollector) SelectAsset(_ []*domain.Asset, _ *loader.Rulebook) (*domain.Asset, error)
func (c *TUICollector) SelectRepairOrders(_ *domain.Faction, _ []*domain.Asset, _ *loader.Rulebook) ([]engine.RepairOrder, error)
func (c *TUICollector) SelectBuyOrder(_ []string, _ []*domain.AssetDefinition) (engine.BuyOrder, error)
func (c *TUICollector) SelectRefitOrder(_ []engine.RefitOption, _ *loader.Rulebook) (engine.RefitOrder, error)
func (c *TUICollector) SelectAttackers(_ []*domain.Asset, _ *loader.Rulebook) ([]*domain.Asset, error)
func (c *TUICollector) SelectDefender(_ *domain.Asset, _ []*domain.Asset, _ *loader.Rulebook) (*domain.Asset, error)
func (c *TUICollector) ConfirmRedirectToBase(_ *domain.Faction, _ *domain.Base, _ int) (bool, error)
```

Each method ignores its arguments and returns the pre-filled field. No prompts, no blocking.

**`cmd/faction-manager/tui/inputs/sell_asset.go`**

```go
type SelectAssetModel struct {
    list   list.Model
    assets []*domain.Asset
}

func NewSelectAssetModel(assets []*domain.Asset, rulebook *loader.Rulebook) SelectAssetModel
func (m SelectAssetModel) Init() tea.Cmd
func (m SelectAssetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m SelectAssetModel) View() string
```

List items show asset name, category, HP, location. Enter sends `AssetSelectedMsg{asset}`.

**`cmd/faction-manager/tui/inputs/buy_asset.go`**

```go
type BuyOrderModel struct {
    step        int  // 0 = world select, 1 = definition select
    worldList   list.Model
    defList     list.Model
    worlds      []string
    definitions []*domain.AssetDefinition
    selectedWorld string
}

func NewBuyOrderModel(worlds []string, purchasable []*domain.AssetDefinition) BuyOrderModel
func (m BuyOrderModel) Init() tea.Cmd
func (m BuyOrderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m BuyOrderModel) View() string
```

Step 0: world list. On Enter: store selected world, build filtered definition list for that world, advance to step 1. Step 1: definition list. On Enter: send `BuyOrderSelectedMsg{BuyOrder{World, Definition}}`.

**`cmd/faction-manager/tui/inputs/refit_asset.go`**

```go
type RefitOrderModel struct {
    step        int  // 0 = asset select, 1 = replacement select
    assetList   list.Model
    replList    list.Model
    options     []engine.RefitOption
    selectedOption *engine.RefitOption
}

func NewRefitOrderModel(options []engine.RefitOption, rulebook *loader.Rulebook) RefitOrderModel
func (m RefitOrderModel) Init() tea.Cmd
func (m RefitOrderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m RefitOrderModel) View() string
```

Step 0: list of assets with at least one valid replacement. Step 1: list of replacement definitions for selected asset. On completion: send `RefitOrderSelectedMsg`.

### Wire into `TurnModel.Update` — `stateActionInput`

```go
case ActionSelectedMsg:
    action := m.availableActions[msg.index]
    switch action.Name() {
    case "Sell Asset":
        assets := /* faction assets eligible for sell */
        m.subModel = inputs.NewSelectAssetModel(assets, m.engine.Rulebook)
    case "Buy Asset":
        worlds, purchasable := /* compute from factionState + rulebook */
        m.subModel = inputs.NewBuyOrderModel(worlds, purchasable)
    case "Refit Asset":
        options := /* compute RefitOptions */
        m.subModel = inputs.NewRefitOrderModel(options, m.engine.Rulebook)
    case "Repair Faction":
        // no inputs — go straight to running the action
        m.runAction(action, &TUICollector{})
    }
    m.state = stateActionInput
```

On `AssetSelectedMsg`, `BuyOrderSelectedMsg`, `RefitOrderSelectedMsg`:
- Build `TUICollector` with collected data
- Call `e.Action.Run(action, currentFaction, factionState, e.Rulebook)` → get mutations
- Apply mutations: `e.Mutation.Apply(factionState, mutations)`
- Append to `m.pendingMutations`
- Set `m.actionResultText` from action name + one-line summary
- Transition to `stateActionResult`

On `ActionResultDoneMsg` (any key):
- Commit: save state, record history event from `m.pendingMutations`
- Call `e.Turn.Advance`; if done → `stateCycleSummary`, else load next faction → `stateSkipPrompt`

**Verification:** SellAsset: asset list renders, selection sells the asset, Coin increases. BuyAsset: world list → definition list → asset purchased. RefitAsset: old asset → replacement → asset refitted.

<br/>

## Phase 6 — RepairAsset inputs

**Goal:** RepairAsset selects assets to repair and sets heal counts.

### Files to create

**`cmd/faction-manager/tui/inputs/repair_asset.go`**

```go
type repairStep int
const (
    repairStepSelect repairStep = iota  // multi-select damaged assets
    repairStepCounts                    // per-asset heal count entry
)

type RepairOrdersModel struct {
    step         repairStep
    assetList    list.Model
    selected     map[string]bool         // asset ID → selected
    assets       []*domain.Asset
    healCounts   map[string]int          // asset ID → count
    countOrder   []*domain.Asset         // assets in count-entry order
    countIndex   int                     // current asset being configured
    faction      *domain.Faction
    rulebook     *loader.Rulebook
}

func NewRepairOrdersModel(faction *domain.Faction, damagedAssets []*domain.Asset, rulebook *loader.Rulebook) RepairOrdersModel
func (m RepairOrdersModel) Init() tea.Cmd
func (m RepairOrdersModel) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m RepairOrdersModel) View() string
```

Step `repairStepSelect`: `bubbles/list` with Space to toggle selection, Enter to confirm. Show current Coin balance in footer. At least one asset must be selected to proceed.

Step `repairStepCounts`: for each selected asset in order — display asset name, current HP, max HP; +/- keys or j/k to adjust heal count; show running Coin cost; Enter confirms and advances. Prevent setting heal count that would exceed Coin or max HP. After last asset: send `RepairOrdersSelectedMsg{orders}`.

### Wire into Phase 5's action switch

```go
case "Repair Asset":
    damagedAssets := /* filter faction.Assets for CurrentHP < MaxHP */
    m.subModel = inputs.NewRepairOrdersModel(currentFaction, damagedAssets, e.Rulebook)
```

**Verification:** RepairAsset multi-selects assets, +/- sets heal counts, Coin cost tracked live, resolves correctly.

<br/>

## Phase 7 — Attack inputs

**Goal:** Attack completes a full cycle including attacker/defender/redirect steps.

### Files to create

**`cmd/faction-manager/tui/inputs/attack_inputs.go`**

```go
type attackStep int
const (
    attackStepSelectAttackers attackStep = iota
    attackStepSelectDefender              // repeated per attacker
    attackStepConfirmRedirect             // repeated per matchup with a base
)

type AttackInputsModel struct {
    step            attackStep
    attackerList    list.Model
    defenderList    list.Model
    selectedAttackers map[string]bool      // asset ID → selected
    attackers       []*domain.Asset
    attackerQueue   []*domain.Asset        // attackers awaiting defender selection
    currentAttacker *domain.Asset
    defenders       map[string]*domain.Asset    // attacker ID → defender
    redirectConfirms map[string]bool            // attacker ID → confirm
    redirectQueue   []redirectPrompt            // matchups awaiting redirect confirm
    currentRedirect *redirectPrompt
    factionState    *state.FactionState
    rulebook        *loader.Rulebook
}

type redirectPrompt struct {
    attackerID      string
    defenderFaction *domain.Faction
    base            *domain.Base
    damage          int  // estimated; actual computed in Resolve
}

func NewAttackInputsModel(eligible []*domain.Asset, factionState *state.FactionState, rulebook *loader.Rulebook) AttackInputsModel
func (m AttackInputsModel) Init() tea.Cmd
func (m AttackInputsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m AttackInputsModel) View() string
```

Step `attackStepSelectAttackers`: multi-select from eligible attackers (Ready, HP > 0, Maintained). Space to toggle, Enter to confirm. Must select at least one.

Step `attackStepSelectDefender`: for current attacker, show eligible defenders on same world (not stealthy, Ready, HP > 0, Maintained, rival faction). Enter selects. Advance to next attacker or to redirect step.

Step `attackStepConfirmRedirect`: for current redirect prompt — show defender faction name, base location, estimated damage. y/n or list select. After all prompts: send `AttackInputsSelectedMsg`.

To populate redirect queue: after all defenders are selected, for each attacker–defender pair, check if the defender's faction has a Base on the same world (`factionBaseOnWorld` logic from `eligibility.go`). If yes, add a redirect prompt for that pair.

**Note on estimated damage:** The redirect prompt happens before dice are rolled. Display the attacker's damage dice expression (from `AssetDefinition.Attack.Damage`) as the estimate rather than a rolled value.

### Wire into Phase 5's action switch

```go
case "Attack":
    eligible := /* eligibleAttackers from factionState */
    m.subModel = inputs.NewAttackInputsModel(eligible, factionState, e.Rulebook)
```

On `AttackInputsSelectedMsg`:
- Build `TUICollector` with `attackers`, `defenders`, `redirectConfirms`
- Call `e.Action.Run` with a fresh `Attack` action wired to the collector
- Continue as per Phase 5 completion flow

**Verification:** Attack selects multiple attackers. Defender list appears per attacker. Redirect confirm appears when applicable. Full attack resolves, mutations applied, history recorded.

<br/>

## Phase 8 — Cycle summary and history recording

**Goal:** Full cycle completes; history.jsonl and faction_state.toml are correct; cycle summary renders.

### History recording

Move the history commit logic from the old wizard into `TurnModel`. It should fire at the same point: after each faction's action phase resolves and `Advance` is called. Use the same `buildEventRecord` logic as the old wizard (or inline equivalent):

```go
func buildEventRecord(factionState *state.FactionState, faction *domain.Faction, mutations []domain.Mutation) (domain.EventRecord, error)
```

This already exists in `cmd/faction-manager/commands/turn/wizard.go` — move it to `cmd/faction-manager/tui/model.go` or a `tui/history.go` helper.

### Files to create

**`cmd/faction-manager/tui/phases/summary.go`**

```go
type factionSummaryRow struct {
    name       string
    startHP    int
    endHP      int
    startCoin  int
    endCoin    int
    action     string  // action taken or "Skipped" / "No Action"
}

type CycleSummaryModel struct {
    rows  []factionSummaryRow
    cycle int
}

func NewCycleSummaryModel(cycleNumber int, rows []factionSummaryRow) CycleSummaryModel
func (m CycleSummaryModel) Init() tea.Cmd
func (m CycleSummaryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m CycleSummaryModel) View() string
```

`View` renders a table: Faction | HP | Coin | Action Taken. HP and Coin show start → end (color-coded for increases/decreases). Skipped factions show "—". Press any key to quit.

`Update` returns `SummaryDoneMsg` on any `tea.KeyMsg`.

### Snapshot tracking in `TurnModel`

To compute deltas for the summary, snapshot each faction's HP and Coin at the start of their turn (before bookkeeping). Store in a `map[string]factionSnapshot` on `TurnModel`. Compare against end-of-turn values when building summary rows.

### Wire into `TurnModel.Update`

When `e.Turn.Advance` returns `done=true`:
- Build summary rows from snapshots vs current state
- Transition to `stateCycleSummary`, embed `NewCycleSummaryModel(cycleNumber, rows)`

On `SummaryDoneMsg`: `stateDone` → return `tea.Quit`.

**Verification:** Run a full cycle. Check `campaigns/test/history.jsonl` — one record per faction. Check `campaigns/test/faction_state.toml` — Coin and HP updated. Cycle summary renders with correct deltas.

<br/>

## Phase 9 — Replace old wizard, delete GMCollector

**Goal:** Old code is removed; `turn` command runs the TUI exclusively.

### Steps

1. Update `cmd/faction-manager/commands/turn/cmd.go`:
   - Replace `runTurnWizard(e, factionState, p)` call with `tui.RunTurnTUI(e, factionState, p)`
   - Remove any remaining wizard imports

2. Delete `cmd/faction-manager/commands/turn/wizard.go`

3. Delete `cmd/faction-manager/forms/collector.go`

4. If `cmd/faction-manager/forms/` is now empty, delete the directory

5. Check whether `cmd/faction-manager/commands/turn/cmd.go` still registers action factories with the engine — if that registration was in `wizard.go`, move it to `cmd.go` or `app.go`

6. Run `go build ./...` and fix any broken imports

7. Run `go test ./...` — all existing tests must pass (engine tests are unaffected by UI changes)

**Verification:**

```bash
cd cmd/faction-manager
go build -o bin/faction-manager .
FACTION_DATA_DIR=/workspaces/gm-toolkit/internal/faction/data ./bin/faction-manager turn --campaign test-attack
```

Complete a full cycle with varied actions (attack, buy, repair, no action). Confirm:
- `campaigns/test-attack/faction_state.toml` reflects all changes
- `campaigns/test-attack/history.jsonl` has one record per faction
- `go test ./...` passes from repo root

<br/>

## End-to-End Verification Checklist

- [ ] `go build ./...` clean from repo root
- [ ] `go test ./...` all pass
- [ ] `turn --campaign test-attack` launches full-screen TUI
- [ ] Resume detection works (start new; re-run with turn in progress to test resume)
- [ ] Bookkeeping screen shows correct income and maintenance per faction
- [ ] All 6 actions (SellAsset, BuyAsset, RefitAsset, RepairAsset, RepairFaction, Attack) complete successfully
- [ ] No Action skips the action phase correctly
- [ ] Cycle summary shows all factions with correct deltas
- [ ] `history.jsonl` has one record per non-skipped faction per cycle
- [ ] `faction_state.toml` updated correctly after each faction
- [ ] Ctrl+C / q exits cleanly at any point without corrupting state

# Contributor Guide — Making Code Changes

This guide walks through the most common code-addition flows in this repo. It assumes you've read `architecture-overview.md` and understand the basic layer split: `internal/faction/` is pure game logic, `cmd/faction-manager/` is the UI and delivery layer.

---

## Quick Reference: Which Files Change for What

| Task | Files touched |
|------|---------------|
| Add a new action | `engine/actions/<name>.go`, `tui/inputs/<name>.go`, `tui/model.go`, `commands/turn.go` |
| Add a new Mutation type | `domain/mutation.go`, `engine/mutation_engine.go`, `tui/narrate.go` |
| Add a new static data field | `domain/` types, `loader/`, `internal/faction/data/*.toml` |
| Add a new Cobra command | `cmd/faction-manager/commands/<name>.go`, `cmd/faction-manager/commands/app.go` |
| Add a new TUI phase screen | `tui/phases/<name>.go`, `tui/model.go` |

---

## Flow 1: Adding a New Action

This is the most common flow. Each action lives in two places: engine logic and TUI input collection.

### Step 1 — Write the engine action

Create `internal/faction/engine/actions/<your_action>.go`. Every action is a struct that implements four methods:

```go
type YourAction struct {
    collector     engine.InputCollector
    factionID     string
    // ... fields populated during Inputs() and used in Resolve()
}

func NewYourAction(collector engine.InputCollector) *YourAction {
    return &YourAction{collector: collector}
}

func (a *YourAction) Name() string { return "Your Action" }

func (a *YourAction) Validate(faction *domain.Faction, factionState *state.FactionState, rulebook *loader.Rulebook) bool {
    // return true when this action is eligible for the current faction
}

func (a *YourAction) Inputs(faction *domain.Faction, factionState *state.FactionState, rulebook *loader.Rulebook) error {
    // call a.collector.Select*() to gather GM choices; store on the struct
}

func (a *YourAction) Resolve(faction *domain.Faction, factionState *state.FactionState, rulebook *loader.Rulebook) error {
    // compute outcomes from stored inputs; populate mutation fields
    // do NOT call state.Save() — mutations are applied by the engine
}

func (a *YourAction) Output() ([]domain.Mutation, error) {
    return []domain.Mutation{
        domain.CoinDelta{...},
        // ...
    }, nil
}
```

See `sell_asset.go` for a minimal example. See `buy_asset.go` for a multi-step inputs example.

**Key rules:**
- `Resolve()` never writes to state directly — it only builds the mutation list.
- `Validate()` is called before the action menu is shown; keep it cheap (no I/O).
- All inputs collected in `Inputs()` must be stored on the struct, not passed as arguments to `Resolve()`.

### Step 2 — Wire the input collector method

If your action needs a new kind of GM selection (not just `SelectAsset` or `SelectBuyOrder`), add a method to the `InputCollector` interface in `internal/faction/engine/input_collector.go`, then implement it on `TUICollector` in `cmd/faction-manager/tui/collector.go`.

If your action only reuses an existing collector method, skip this step.

### Step 3 — Write the TUI input sub-model

Create `cmd/faction-manager/tui/inputs/<your_action>.go`. This is a Bubbletea sub-model that collects the inputs and emits a completion message when done:

```go
type YourActionSelectedMsg struct {
    // the fields TurnModel needs to start resolution
}

type YourActionModel struct {
    list list.Model
    // ...
}

func NewYourActionModel(...) YourActionModel { ... }

func (m YourActionModel) Init() tea.Cmd { return nil }

func (m YourActionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.Type == tea.KeyEnter {
            // emit the completion message
            return m, func() tea.Msg { return YourActionSelectedMsg{...} }
        }
    }
    var cmd tea.Cmd
    m.list, cmd = m.list.Update(msg)
    return m, cmd
}

func (m YourActionModel) View() string { return m.list.View() }
```

For multi-step input (select world, then select asset), use a `step int` field and advance it on each Enter press, like `BuyOrderModel` does.

### Step 4 — Wire into TurnModel

In `cmd/faction-manager/tui/model.go`:

1. **Handle the completion message** — in `Update()`, add a case for `inputs.YourActionSelectedMsg`:

```go
case inputs.YourActionSelectedMsg:
    return m.handleYourActionSelected(msg)
```

2. **Launch the sub-model** — in the `startActionInput()` function (or wherever action inputs are dispatched by name), add a branch for your action:

```go
case "Your Action":
    m.state = stateActionInput
    m.subModel = m.resizeSub(inputs.NewYourActionModel(...))
    return m, m.subModel.Init()
```

3. **Handle the result** — add `handleYourActionSelected` to populate `pendingAction` inputs and transition to resolution.

### Step 5 — Register the action

In `cmd/faction-manager/commands/turn.go`, add one line:

```go
a.Engine.Action.Register(func() engine.Action { return actions.NewYourAction(nil) })
```

The `nil` collector is replaced at runtime by the engine with the live `TUICollector`.

### Step 6 — Build and test

```bash
cd cmd/faction-manager
go build -o bin/faction-manager .
FACTION_DATA_DIR=/workspaces/gm-toolkit/internal/faction/data ./bin/faction-manager --campaign test turn
```

Run the turn and select your action. Verify the mutation list in `campaigns/test/history.jsonl` after completing the turn.

---

## Flow 2: Adding a New Mutation Type

Mutations are how all state changes are expressed. Adding a new kind of state change means touching three places.

### Step 1 — Define the type

In `internal/faction/domain/mutation.go`, add a struct that implements the `Mutation` interface (a single `mutationTag()` marker method):

```go
type YourMutation struct {
    FactionID         string
    // ... fields
    Cause             string
    CausedByFactionID string
}

func (m YourMutation) mutationTag() {}
```

Follow the naming and field conventions of the surrounding types. Always include `Cause` and `CausedByFactionID` for the history log.

### Step 2 — Handle it in the MutationEngine

In `internal/faction/engine/mutation_engine.go`, add a case in the `Apply()` method's type switch:

```go
case domain.YourMutation:
    // update factionState.Factions[m.FactionID] accordingly
```

### Step 3 — Add narration

In `cmd/faction-manager/tui/narrate.go`, add a case in the narrate function so the TUI can describe it to the GM in human-readable form. If you skip this, the TUI will silently ignore the mutation in result display.

---

## Flow 3: Adding a Static Data Field

If you need to add a field to the asset catalog, tags, or goals (e.g., a new property on `AssetDefinition`):

1. **Add the Go field** to the relevant type in `internal/faction/domain/`.
2. **Update the loader** in `internal/faction/loader/` to read the new field from TOML.
3. **Update the TOML data files** in `internal/faction/data/` to include the new field on the relevant entries.

Because the Rulebook is read-only after startup, you don't need to worry about concurrent writes. If the field is optional, give it a zero value and let callers check it before use.

---

## Flow 4: Adding a New Cobra Command

For commands that don't need the full turn TUI (faction creation, deletion, one-off queries):

1. Create `cmd/faction-manager/commands/<your_command>.go`:

```go
func (a *App) yourCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "your-command",
        Short: "One-line description",
        RunE: func(cmd *cobra.Command, args []string) error {
            // use huh for interactive prompts
            // call a.Engine.* for game logic
            return nil
        },
    }
}
```

2. Register it in `cmd/faction-manager/commands/app.go` by adding `rootCmd.AddCommand(a.yourCmd())`.

Use `huh` for forms and confirmations. Don't start a Bubbletea program unless you need a full interactive TUI. Most management commands (create, delete, list) are fine as synchronous huh flows.

---

## Flow 5: Adding a New TUI Phase Screen

Phase screens live in `cmd/faction-manager/tui/phases/`. These are Bubbletea sub-models used for turn flow screens (bookkeeping display, skip prompt, cycle summary, etc.) rather than input collection.

1. Create `tui/phases/<your_phase>.go` with the standard Bubbletea model shape (`Init`, `Update`, `View`).
2. Define a completion message type in the same file.
3. In `tui/model.go`:
   - Add a `turnState` constant for the new phase.
   - Handle the entry transition (set `m.state`, create `m.subModel`, return `m.subModel.Init()`).
   - Handle the completion message in `Update()`.
   - Add a `View()` case that delegates to `m.subModel.View()`.

---

## Build and Test Checklist

Before considering any change done:

```bash
# from repo root
go test ./...

# from cmd/faction-manager
go build -o bin/faction-manager .
FACTION_DATA_DIR=/workspaces/gm-toolkit/internal/faction/data ./bin/faction-manager --campaign test turn
```

- `go test ./...` covers engine logic, mutation application, and action validation.
- Manual turn run is required for any action or TUI change — tests don't cover UI paths.
- After a turn, inspect `cmd/faction-manager/campaigns/test/history.jsonl` to verify mutations were recorded correctly.
- Check `cmd/faction-manager/campaigns/test/faction_state.toml` to verify state was persisted correctly.

---

## Common Pitfalls

**Forgetting `resizeSub`** — sub-models created mid-loop never receive Bubbletea's initial `WindowSizeMsg`. Always wrap new sub-models with `m.resizeSub(...)` before assigning to `m.subModel`.

**Writing state in `Resolve()`** — `Resolve()` must only build the mutation list. Direct writes to `FactionState` will bypass history recording and may be overwritten when mutations are applied.

**Reusing the `nil` collector in tests** — the action's `collector` field is `nil` when registered (see `turn.go`); the engine substitutes the live collector at runtime. If you're writing a test that calls `Inputs()`, you'll need to provide a mock or stub collector.

**Forgetting `mutationTag()`** — new `Mutation` types must implement the marker interface or the compiler will reject them where a `Mutation` is expected.

**Missing narration case** — if you add a mutation but don't add a case in `narrate.go`, the action result screen will be blank for that mutation. It won't crash, but it will confuse the GM.

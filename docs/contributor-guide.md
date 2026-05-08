# Contributor Guide — Common Patterns

This guide documents the most common code-addition flows in this codebase. It assumes you've read `architecture-overview.md` and understand the basic layer split: `internal/faction/` is pure game logic; `cmd/faction-manager/` is the delivery layer.

---

## Common Commands

`go test ./...` — run all tests from the repo root.
`go test ./internal/faction/engine/... -v -run TestSomething` — run a specific test package with verbose output.

---

## Quick Reference: Which Files Change for What

| Task | Files touched |
|------|---------------|
| Add a new action | `engine/action/actions/<name>.go`, `engine/action/actions/register.go` |
| Add a new mutation type | `domain/mutation.go`, `engine/mutation/mutation.go`, `narrative/digest/build.go` |
| Add a tag handler | `engine/tag/tags/<name>.go`, `engine/tag/tag.go` |
| Add a hook handler | `engine/hooks/` (if new interface needed), registering package |
| Add a goal type | `data/goals.toml`, `engine/goal/goal.go`, `engine/goal/progress.go` |
| Add a Cobra command | `commands/<name>.go`, `commands/app.go` |
| Add a narrative beat | `narrative/digest/types.go` (if new event type), `narrative/digest/build.go` |
| Add a static data field | `domain/` type, `rulebook/` loader, `data/*.toml` files |

---

## Flow 1: Adding a New Action

### Step 1 — Write the engine action

Create `internal/faction/engine/action/actions/<your_action>.go`. Every action is a struct implementing a five-method interface:

```go
type YourAction struct {
    collector     action.Collector
    factionID     string
    selectedAsset *domain.Asset
    // ... other fields populated by Inputs() and read by Resolve()
}

func NewYourAction(collector action.Collector) *YourAction {
    return &YourAction{collector: collector}
}

func (a *YourAction) Name() string { return "Your Action" }

func (a *YourAction) Validate(faction *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) bool {
    // return true when this action is eligible for the current faction
    return len(faction.Assets) > 0
}

func (a *YourAction) Inputs(faction *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) error {
    // call a.collector.Select*() to gather GM choices; store on the struct
    selected, err := a.collector.SelectAsset(faction.Assets, rulebook)
    if err != nil {
        return fmt.Errorf("your action: %w", err)
    }
    a.selectedAsset = selected
    a.factionID = faction.ID
    return nil
}

func (a *YourAction) Resolve(_ *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) error {
    // compute outcomes from stored inputs
    // do NOT call state.Save() or write to factionState — mutations handle all state changes
    return nil
}

func (a *YourAction) Output() ([]domain.Mutation, error) {
    return []domain.Mutation{
        domain.CoinDelta{
            FactionID:         a.factionID,
            Delta:             5,
            Cause:             "your_action",
            CausedByFactionID: a.factionID,
        },
    }, nil
}
```

See `sell_asset.go` for a minimal example. See `buy_asset.go` for a multi-step inputs example.

**Key rules:**
- `Resolve()` never writes to state directly — it only computes data for `Output()`.
- `Validate()` is called before the menu is shown; keep it cheap (no I/O).
- All inputs from `Inputs()` must be stored on the struct, not passed to `Resolve()`.
- `Cause` and `CausedByFactionID` on every mutation are required for history and narrative correctness.

### Step 2 — Add a collector method if needed

If your action requires a GM decision not covered by the existing methods on `action.Collector`, add a new method to the interface in `internal/faction/engine/action/collector.go`. You must then add the corresponding `Fn` field and default implementation to `ScriptedCollector` in `testharness/collector.go` — a nil default is fine if the test doesn't exercise it, but the method must compile.

If your action reuses existing collector methods, skip this step.

### Step 3 — Register the action

In `internal/faction/engine/action/actions/register.go`, add one line to `RegisterDefaultActions`:

```go
e.Action.Register(func(c action.Collector) action.Action { return NewYourAction(c) })
```

The factory receives the live collector at runtime. If your action also needs the hooks registry or the roller, follow the pattern of `NewBuyAsset(c, e.Hooks)` or `NewAttack(c, e.Rand, e.Hooks)`.

### Step 4 — Build and verify

```bash
go build -o cmd/faction-manager/bin/faction-manager ./cmd/faction-manager
FACTION_DATA_DIR=/workspaces/gm-toolkit/internal/faction/data \
  ./cmd/faction-manager/bin/faction-manager --campaign test turn
```

Inspect `campaigns/test/history.jsonl` after the turn to verify mutations were recorded correctly.

---

## Flow 2: Adding a New Mutation Type

Mutations are the only path to state change. Adding a new kind means touching three places.

### Step 1 — Define the type

In `internal/faction/domain/mutation.go`, add a struct implementing the `Mutation` interface:

```go
type YourMutation struct {
    FactionID         string `json:"faction_id"`
    // ... fields
    Cause             string `json:"cause"`
    CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m YourMutation) Type() string { return "your_mutation" }
```

Follow the field conventions of surrounding types. Always include `Cause` and `CausedByFactionID`. The string returned by `Type()` is the stable discriminator written to the history JSONL — once chosen, treat it as immutable.

### Step 2 — Handle it in the MutationEngine

In `internal/faction/engine/mutation/mutation.go`, add a case in `Apply()`'s type switch:

```go
case domain.YourMutation:
    if faction, ok := factionState.Factions[v.FactionID]; ok {
        // update faction state
    }
```

### Step 3 — Handle it in the narrative digest

In `internal/faction/narrative/digest/build.go`, add a case in the `mr.Type` switch inside `Build()`. If the mutation has no narrative significance, add it with a `// dropped` comment. If it does, decode it and update the appropriate `FactionBeat` or `CrossEvent` fields:

```go
case "your_mutation":
    var m domain.YourMutation
    if err := json.Unmarshal(mr.Payload, &m); err != nil {
        return CycleDigest{}, err
    }
    // update beat fields
```

Every type must appear in this switch. The `default:` branch appends an `"unknown mutation"` note to the beat, which is the intended fallback for forward-compatibility — not a substitute for explicit handling.

---

## Flow 3: Adding a Tag Handler

Tags are data-only until a handler registers hooks for them. A handler wires one or more hooks into the engine for any faction that carries the tag.

### Step 1 — Write the handler

Create `internal/faction/engine/tag/tags/<your_tag>.go`:

```go
package tags

const YourTagID = "T-XXX"

type YourTagReactor struct {
    FactionID string
}

func (reactor *YourTagReactor) OnMutations(
    mutations []domain.Mutation,
    _ *state.FactionState,
    _ *rulebook.Rulebook,
) []domain.Mutation {
    var extra []domain.Mutation
    for _, m := range mutations {
        if v, ok := m.(domain.AssetRemoved); ok && v.Cause == "attack" {
            extra = append(extra, domain.CoinDelta{
                FactionID:         reactor.FactionID,
                Delta:             1,
                Cause:             "your_tag",
                CausedByFactionID: reactor.FactionID,
            })
        }
    }
    return extra
}

func RegisterYourTag(eng *engine.Engine, faction *domain.Faction) {
    eng.Hooks.RegisterMutationReactor(
        hooks.FactionScope(faction.ID),
        "your_tag",
        &YourTagReactor{FactionID: faction.ID},
    )
}
```

The reactor struct implements one or more hook interfaces (`hooks.MutationReactor`, `hooks.RollModifier`, `hooks.RollResultHook`, `hooks.TieResolver`). Each is registered separately via the corresponding registry method.

See `scavengers.go` for a complete minimal example; `warlike.go` for a `RollModifier` example.

### Step 2 — Wire into the tag router

In `internal/faction/engine/tag/tag.go`, add an entry to the `handlers` map:

```go
var handlers = map[string]func(*engine.Engine, *domain.Faction){
    // ... existing entries
    tags.YourTagID: tags.RegisterYourTag,
}
```

`RegisterDefaultTags(eng, factionState)` is called once at startup and walks all faction tags. Factions without the tag are unaffected.

---

## Flow 4: Adding a Hook Handler

Use this when you need engine-modifiable behavior that isn't tied to a specific tag — for example, a campaign-wide rule override or a one-off ability effect.

### Cat 1 — RollModifier (pre-roll dice pool)

Implement `hooks.RollModifier` (`OfferModifiers(ctx, factionState, rulebook) []ModifierOffer`) and register with `eng.Hooks.RegisterRollModifier(scope, key, handler)`. Each `ModifierOffer` has a `Description` and an `Apply(*RollState)` function that calls `rollState.AddDie(sides)` to expand the pool. The collector selects which offers the GM accepts.

### Cat 2 — RollResultHook (post-roll reroll)

Implement `hooks.RollResultHook` (`OnRollResult(ctx, result, factionState, rulebook) RerollDirective`) and register with `eng.Hooks.RegisterRollResultHook(scope, key, handler)`. Return a `RerollDirective` naming the indices into `result.Dice` to reroll. Set `Elective: true` for optional rerolls the collector must confirm.

### Cat 3 — MutationReactor (post-action mutations)

Implement `hooks.MutationReactor` (`OnMutations(mutations, factionState, rulebook) []domain.Mutation`) and register with `eng.Hooks.RegisterMutationReactor(scope, key, handler)`. Return any additional mutations to append. The dispatcher recurses until no new mutations are produced (depth cap: 5). See `ScavengersReactor` for a reference implementation.

### Cat 4 — Rule Modifiers (synchronous query overrides)

Implement one of the typed modifier interfaces (`hooks.AssetCostModifier`, `hooks.MaintenanceCostModifier`, `hooks.WorldTechLevelModifier`, `hooks.AssetMovementGranter`) and register with the corresponding registry method. These fire synchronously during validation and bookkeeping, not in the mutation pipeline.

### Cat 5 — TieResolver (attack/defense tie)

Implement `hooks.TieResolver` (`ResolveTie(ctx, factionState) TieOutcome`) and register with `eng.Hooks.RegisterTieResolver(scope, key, handler)`. The first registered handler wins. Return `TieAttackerWins`, `TieDefenderWins`, or `TieStandard` (default `>=` comparison).

**Scoping.** Use `hooks.GlobalScope()` for campaign-wide behavior, `hooks.FactionScope(factionID)` for a single faction, or `hooks.AssetScope(assetID)` for a specific asset.

---

## Flow 5: Adding a Goal Type

Goals live in two places: the TOML data file (definition) and the goal engine (progress and locking logic).

### Step 1 — Add to goals.toml

In `internal/faction/data/goals.toml`, add an entry:

```toml
[[goals]]
id          = "G-013"
name        = "Your Goal"
description = "What the faction is trying to achieve."
difficulty  = 2
```

The `id` is the stable discriminator used in save files and code — treat it as immutable.

### Step 2 — Add progress logic

In `internal/faction/engine/goal/progress.go`, add a progress function:

```go
func progressYourGoal(
    actingFaction *domain.Faction,
    mutations []domain.Mutation,
    factionState *state.FactionState,
    rulebook *rulebook.Rulebook,
) []domain.Mutation {
    // inspect mutations for goal-relevant events
    // use completeGoal(actingFaction, xp) to produce completion mutations
    // use domain.GoalProgressed{...} for incremental progress
    return nil
}
```

Then wire it into `UpdateProgress()` in `goal.go`:

```go
case "G-013":
    return progressYourGoal(actingFaction, mutations, factionState, rulebook)
```

`completeGoal(faction, xp)` returns a `GoalCompleted` + `XPAwarded` pair — always use it rather than emitting those mutations directly.

### Step 3 — Add lock logic (if needed)

If your goal restricts the faction's action choices during some turns (a timer-based skip, or a combat-phase restriction), add a case in `CheckLock()` in `goal.go`:

```go
case "G-013":
    return checkLockYourGoal(faction)
```

See `checkLockChangeHomeworld` and `checkLockPlanetarySeizure` for reference implementations of `LockSkip` and `LockRestrictActions` respectively.

---

## Flow 6: Adding a New Cobra Command

For commands that don't require the full turn pipeline — faction management, report generation, one-off queries:

### Step 1 — Create the command file

Create `cmd/faction-manager/commands/<your_command>.go`:

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

### Step 2 — Register it

In `cmd/faction-manager/commands/app.go`, add `rootCmd.AddCommand(a.yourCmd())`.

Use `huh` for forms and confirmations. Most management commands are fine as synchronous `huh` flows. Only start a Bubbletea program if you need a full interactive TUI — which is deferred; see below.

---

## Flow 7: Adding a Narrative Beat

The narrative digest converts raw mutation history into semantic events for the renderer. Adding a new beat means teaching `digest/build.go` about a new mutation type or event shape.

### Step 1 — Add an event type (if needed)

If your beat doesn't fit any existing struct in `narrative/digest/types.go`, add one. Keep it data-only — no methods, no logic.

### Step 2 — Handle the mutation in Build()

In `internal/faction/narrative/digest/build.go`, add a case in the `mr.Type` switch inside `Build()`:

```go
case "your_mutation":
    var m domain.YourMutation
    if err := json.Unmarshal(mr.Payload, &m); err != nil {
        return CycleDigest{}, err
    }
    beat.YourEvents = append(beat.YourEvents, YourEvent{...})
```

If the mutation has no narrative significance, still add a `case "your_mutation": // dropped` line — the `default:` branch appends an `"unknown mutation"` note which will show up in rendered output.

### Step 3 — Render it

In `internal/faction/narrative/renderer.go` or `wire_renderer.go`, add a template or conditional to include your new event type in rendered markdown output.

---

## Flow 8: Writing a Test with the Harness

The test harness provides a complete fixture for action and pipeline tests.

### Basic setup

```go
func TestYourScenario(t *testing.T) {
    h := testharness.NewHarness(t, "/workspaces/gm-toolkit/internal/faction/data")

    // AddFaction sets up a faction with one default asset and adds it to FactionState.
    attacker := h.AddFaction("alpha", "Andoni", 3, 2, 4)
    _ = h.AddFaction("beta", "Vance", 2, 2, 2)

    // If factions have tags that register hooks, call this after all factions are added.
    h.RegisterTags()

    // Start the turn (randomises faction order; skipped if TurnState is already in progress).
    if err := h.Engine.Turn.Start(h.FactionState); err != nil {
        t.Fatalf("start turn: %v", err)
    }

    // Script the collector to pick a specific action by name.
    h.Collector.SelectActionFn = func(_ *domain.Faction, available []action.Action) (action.Action, error) {
        for _, a := range available {
            if a.Name() == "Sell Asset" {
                return a, nil
            }
        }
        return nil, fmt.Errorf("Sell Asset not available")
    }
    h.Collector.SelectAssetFn = func(assets []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
        return assets[0], nil
    }

    _, err := h.Engine.RunFactionTurn(h.FactionState, h.Cfg, h.Collector, h.Observer)
    if err != nil {
        t.Fatalf("RunFactionTurn: %v", err)
    }

    // Assert via history JSONL.
    records := testharness.ReadHistory(t, h.Cfg.HistoryPath)
    _, found := testharness.FindMutationType(records, "asset_removed")
    testharness.CheckStep(t, "asset removed from roster", found, "expected asset_removed in history")
}
```

### Key harness helpers

- `h.AddFaction(id, homeworld, force, cunning, wealth)` — faction with one default asset.
- `testharness.AddBase(faction, world, hp)` — adds a Base of Influence.
- `testharness.AddAssetOnWorld(faction, world)` — adds a second asset at a specific location.
- `h.RegisterTags()` — registers tag hooks; call after all factions are added.
- `testharness.ReadHistory(t, path)` — parses history JSONL into `[]domain.EventRecord`.
- `testharness.FindMutationType(records, type)` — first mutation of a given type.
- `testharness.FindMutationByCause(records, cause)` — first mutation with a given cause.
- `testharness.FindMutationByTypeAndCause(records, type, cause)` — first matching both.
- `testharness.CheckStep(t, desc, ok, detail)` — logs ✓/✗ for readable step-by-step assertions.

If `ScriptedCollector`'s `Fn` overrides aren't sufficient (e.g., you need to assert exact call arguments), use `go.uber.org/mock/gomock` with the generated `MockCollector` in `engine/action/actions/mocks/`.

For complete multi-faction scenarios, see `engine/testharness/scenarios/helpers.go`.

---

## Flow 9: Adding a Static Data Field

If you need a new field on the asset catalog, tags, or goals:

1. **Add the Go field** to the relevant type in `internal/faction/domain/`.
2. **Update the loader** in `internal/faction/rulebook/` to read the new field from TOML.
3. **Update the TOML files** in `internal/faction/data/` to include the new field on relevant entries.

Because the Rulebook is read-only after startup, there are no concurrency concerns. Optional fields should carry a safe zero value; callers check before use.

---

## TUI Flows (Deferred)

The interactive Bubbletea TUI is the next major frontend milestone but is not yet implemented. When built, two new flow categories will be added here:

- **Adding a TUI phase screen** — Bubbletea sub-models for turn flow screens (bookkeeping display, action selection, cycle summary).
- **Adding a TUI input sub-model** — Bubbletea models for action input collection, wired into `TurnModel`.

The engine doesn't change when the TUI lands — it implements `InputCollector` and `TurnObserver` and drives `RunCycle()` from the outside.

---

## Build and Test Checklist

Before considering any change done:

```bash
# from repo root
go test ./...

# build
go build -o cmd/faction-manager/bin/faction-manager ./cmd/faction-manager

# smoke-test (pick a relevant subcommand)
FACTION_DATA_DIR=/workspaces/gm-toolkit/internal/faction/data \
  ./cmd/faction-manager/bin/faction-manager --campaign test <subcommand>
```

After any turn-pipeline change, inspect `campaigns/test/history.jsonl` to verify mutations were recorded. After any state-change, inspect `campaigns/test/faction_state.toml` to verify persistence.

---

## Common Pitfalls

**Writing state in `Resolve()`** — `Resolve()` must only compute data for `Output()`. Direct writes to `FactionState` bypass history recording and may be overwritten when mutations are applied.

**Missing `Type()` on a new mutation** — new `Mutation` types must implement `Type() string` or the compiler rejects them. The returned string is written to history JSONL — once chosen, treat it as immutable.

**Missing a case in `Build()`** — every mutation type needs a case in `narrative/digest/build.go`, even a `// dropped` comment. The `default:` branch appends `"unknown mutation: <type>"` to the narrative output.

**Missing `CausedByFactionID`** — mutations without it break cross-faction attribution in the narrative digest. All combat and ability mutations must carry it.

**Not calling `h.RegisterTags()` in tag tests** — tags register their hooks at startup via `RegisterDefaultTags`. Skipping this call means the hook never fires, even if the faction has the tag.

**ScriptedCollector defaults to accept-all** — `SelectAction` picks the first available action; `ConfirmAbilityApplied` returns `true`; `SelectModifiers` accepts all offers; `ConfirmReroll` returns `true`. Override explicitly when the test depends on specific behavior, or you may be testing more than you intended.

# Use Asset Ability — Implementation Plan

A phased build guide for the `feature/use-asset-ability` branch. Written for a fresh session — read this doc first, then the key files listed per phase.

**Branch:** `feature/use-asset-ability`  
**Phase 1 committed:** `259e3ad` — domain types, loader parsing, TOML ability data

<br/>

## Design Decisions (read before any phase)

These were ratified in the design session and must be followed precisely:

| # | Decision |
|---|----------|
| 1 | Movement does not enforce hex distance — GM confirms, engine updates `Asset.Location` only |
| 2 | Assets with `Ability == nil` use GM fallback: show description, `ConfirmAbilityApplied` |
| 3 | `AbilityEngine` has a step handler registry keyed by `AbilityStepType` plus a custom handler override map keyed by asset definition ID |
| 4 | World selection for movement includes all worlds derived from current `FactionState` plus a hardcoded **"Astral Sea"** option appended at the end |
| 5 | Movement Coin cost is emitted as a `CoinDelta` mutation (negative) when `CoinCost > 0` |
| 6 | Faction test uses the existing `Roller` interface — same as Attack |
| 7 | `reveal_stealth` emits one `AssetStealthCleared` per stealthed asset on the acting asset's planet belonging to the target faction |
| 8 | `coin_drain` emits one `CoinDelta` (negative) on the target faction |
| 9 | `coin_steal` emits one `CoinDelta` (negative) on target + one `CoinDelta` (positive) on acting faction |
| 10 | Faction test tie goes to the defender (same rule as Attack) — ability effect does **not** apply on a tie |
| 11 | `UseAssetAbility` selects assets up-front (ordered list), then resolves each in sequence — same committed-upfront pattern as Attack |

<br/>

## Ability Step Types Reference

```
AbilityStepMovement    = "movement"
AbilityStepFactionTest = "faction_test"

EffectRevealStealth = "reveal_stealth"
EffectCoinDrain     = "coin_drain"
EffectCoinSteal     = "coin_steal"
```

Assets without an `[assets.X.ability]` TOML section have `AssetDefinition.Ability == nil` and fall through to the GM adjudication path. Current deferred assets: Pretech Logistics (`F5-002`), Tripwire Cells (`C4-003`), Seditionists (`C4-004`).

<br/>

## Deriving the World List

For `SelectMoveDestination`, build the world list from `FactionState` at resolve time:

```go
func worldsFromState(factionState *state.FactionState) []string {
    seen := map[string]bool{}
    for _, faction := range factionState.Factions {
        for _, asset := range faction.Assets {
            if asset.Location != "" {
                seen[asset.Location] = true
            }
        }
        for _, base := range faction.Bases {
            if base.Location != "" {
                seen[base.Location] = true
            }
        }
    }
    worlds := make([]string, 0, len(seen))
    for w := range seen {
        worlds = append(worlds, w)
    }
    sort.Strings(worlds)
    worlds = append(worlds, "Astral Sea") // always last
    return worlds
}
```

<br/>
<br/>

## Phase 2 — AssetMoved Mutation + AbilityEngine

**Goal:** `AbilityEngine` compiles and step handlers for `movement` and `faction_test` are registered and unit-testable.

### Key files to read

| File | Why |
|------|-----|
| `internal/faction/domain/asset.go` | `AbilityDefinition`, `AbilityStep`, step/effect type constants — added in Phase 1 |
| `internal/faction/domain/mutation.go` | Existing mutation types; add `AssetMoved` here |
| `internal/faction/engine/mutation_engine.go` | `Apply` switch — add `AssetMoved` case |
| `internal/faction/engine/action.go` | `Action` interface; `InputCollector` interface to extend |
| `internal/faction/engine/actions/attack.go` | Reference for `Roller` usage and faction test roll pattern |
| `internal/faction/engine/core.go` | `Engine` struct — add `AbilityEngine` field here |

### Tasks

#### 1. Add `AssetMoved` mutation (`internal/faction/domain/mutation.go`)

```go
type AssetMoved struct {
    FactionID    string `json:"faction_id"`
    AssetID      string `json:"asset_id"`
    FromLocation string `json:"from_location"`
    ToLocation   string `json:"to_location"`
}

func (m AssetMoved) Type() string { return "asset_moved" }
```

#### 2. Wire `AssetMoved` into `MutationEngine.Apply`

In `internal/faction/engine/mutation_engine.go`, add a case to the `Apply` switch:

```go
case domain.AssetMoved:
    faction := factionState.Factions[m.FactionID]
    faction.Assets[m.AssetID].Location = m.ToLocation
```

#### 3. Create `AbilityEngine` (`internal/faction/engine/ability_engine.go`)

```go
package engine

import (
    "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// StepHandler resolves one ability step for an asset.
// Returns mutations to apply; does not apply them itself.
type StepHandler func(
    faction *domain.Faction,
    asset *domain.Asset,
    step domain.AbilityStep,
    collector InputCollector,
    roller Roller,
    factionState *state.FactionState,
    rulebook *loader.Rulebook,
) ([]domain.Mutation, error)

// CustomAbilityHandler is a full override for a specific asset definition ID.
// Used for bespoke abilities that don't fit the step model.
type CustomAbilityHandler func(
    faction *domain.Faction,
    asset *domain.Asset,
    collector InputCollector,
    roller Roller,
    factionState *state.FactionState,
    rulebook *loader.Rulebook,
) ([]domain.Mutation, error)

type AbilityEngine struct {
    stepHandlers   map[domain.AbilityStepType]StepHandler
    customHandlers map[string]CustomAbilityHandler // keyed by AssetDefinition ID
}

func NewAbilityEngine() *AbilityEngine {
    ae := &AbilityEngine{
        stepHandlers:   make(map[domain.AbilityStepType]StepHandler),
        customHandlers: make(map[string]CustomAbilityHandler),
    }
    ae.stepHandlers[domain.AbilityStepMovement] = movementStepHandler
    ae.stepHandlers[domain.AbilityStepFactionTest] = factionTestStepHandler
    return ae
}

// RegisterCustomHandler registers a bespoke handler for a specific asset definition ID,
// overriding the step-based resolution path.
func (ae *AbilityEngine) RegisterCustomHandler(defID string, handler CustomAbilityHandler) {
    ae.customHandlers[defID] = handler
}

// Run resolves an asset's ability. Returns mutations; does not apply them.
// If the asset's definition has no Ability, returns nil (caller uses GM fallback).
func (ae *AbilityEngine) Run(
    faction *domain.Faction,
    asset *domain.Asset,
    def *domain.AssetDefinition,
    collector InputCollector,
    roller Roller,
    factionState *state.FactionState,
    rulebook *loader.Rulebook,
) ([]domain.Mutation, error) {
    if handler, ok := ae.customHandlers[def.ID]; ok {
        return handler(faction, asset, collector, roller, factionState, rulebook)
    }
    if def.Ability == nil {
        return nil, nil
    }
    var mutations []domain.Mutation
    for _, step := range def.Ability.Steps {
        handler, ok := ae.stepHandlers[step.Type]
        if !ok {
            return nil, fmt.Errorf("no handler for ability step type %q", step.Type)
        }
        stepMutations, err := handler(faction, asset, step, collector, roller, factionState, rulebook)
        if err != nil {
            return nil, err
        }
        mutations = append(mutations, stepMutations...)
    }
    return mutations, nil
}
```

#### 4. Implement `movementStepHandler`

Logic:
1. Call `collector.SelectMoveDestination(asset, worldsFromState(factionState))` — see world list derivation above
2. If `step.CoinCost > 0`, append `CoinDelta{FactionID: faction.ID, Delta: -step.CoinCost}`
3. Append `AssetMoved{FactionID: faction.ID, AssetID: asset.ID, FromLocation: asset.Location, ToLocation: destination}`

#### 5. Implement `factionTestStepHandler`

Logic:
1. Build candidate target factions: all factions except the acting faction that have at least one asset on `asset.Location` (for non-`reveal_stealth` effects) — or any faction for `reveal_stealth` (Informers can target even with no visible assets)
2. Call `collector.SelectFactionTestTarget(asset, step.Effect, candidates)`
3. Roll `1d10 + acting faction's relevant stat` vs `1d10 + target faction's relevant stat`
   - Attacker stat: `step.AttackerStat`; defender stat: `step.DefenderStat`
   - Ties go to defender — no effect applied
4. On attacker strictly greater, apply effect:
   - `reveal_stealth`: for each of target faction's assets on `asset.Location` where `Stealthy == true`, append `AssetStealthCleared`
   - `coin_drain`: append `CoinDelta{FactionID: targetFaction.ID, Delta: -roll(step.EffectDice)}`
   - `coin_steal`: append `CoinDelta` (negative) on target + `CoinDelta` (positive) on acting faction

#### 6. Add `AbilityEngine` to `Engine` (`internal/faction/engine/core.go`)

Add `AbilityEngine *AbilityEngine` field to the `Engine` struct and initialize it in the constructor.

#### 7. Extend `InputCollector` interface (`internal/faction/engine/action.go`)

Add three new methods:

```go
SelectMoveDestination(asset *domain.Asset, worlds []string) (string, error)
SelectFactionTestTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Faction) (*domain.Faction, error)
ConfirmAbilityApplied(asset *domain.Asset, def *domain.AssetDefinition) (bool, error)
```

Update the existing test fake collector in `engine/actions/` tests and the `TUICollector` stub to satisfy the extended interface (return zero values / `false, nil` for new methods).

### Commit message

```
feat: phase 2 — AssetMoved mutation and AbilityEngine with step handlers
```

<br/>
<br/>

## Phase 3 — UseAssetAbility Action

**Goal:** `UseAssetAbility` is registered in the `ActionEngine`, passes validation, and produces correct mutations in unit tests.

### Key files to read

| File | Why |
|------|-----|
| `internal/faction/engine/action.go` | `Action` interface + `InputCollector` (extended in Phase 2) |
| `internal/faction/engine/ability_engine.go` | `AbilityEngine.Run` — called from `Resolve` |
| `internal/faction/engine/actions/sell_asset.go` | Simplest reference for the Validate→Inputs→Resolve→Output pattern |
| `internal/faction/engine/actions/attack.go` | Reference for committed-upfront asset selection + per-asset loop |
| `internal/faction/engine/action_engine.go` | `Register` call — add `UseAssetAbility` factory here |

### Tasks

#### 1. Create `internal/faction/engine/actions/use_asset_ability.go`

```go
type UseAssetAbility struct {
    collector     InputCollector
    selectedAssets []*domain.Asset // set by Inputs, consumed by Resolve
}
```

**`Validate`** — returns nil if the faction has at least one A-flagged, Ready (not inactive), and Maintained asset:
```go
for _, asset := range faction.Assets {
    def := rulebook.Assets[asset.DefinitionID]
    if def.HasFlag(domain.FlagAction) && asset.Ready && asset.Maintained {
        return nil
    }
}
return fmt.Errorf("no usable A-flagged assets")
```

**`Inputs`** — build the candidate list (A-flagged, Ready, Maintained), call `collector.SelectAbilityAssets`. The returned slice is the ordered resolution sequence.

**`Resolve`** — for each asset in order:
1. Look up `def := rulebook.Assets[asset.DefinitionID]`
2. Call `engine.AbilityEngine.Run(faction, asset, def, collector, roller, factionState, rulebook)`
3. If `Run` returns `nil, nil` (no ability / deferred) → call `collector.ConfirmAbilityApplied(asset, def)`
4. Accumulate returned mutations

**`Output`** — return accumulated mutations as-is.

#### 2. Add `SelectAbilityAssets` to `InputCollector`

```go
SelectAbilityAssets(faction *domain.Faction, candidates []*domain.Asset, rulebook *loader.Rulebook) ([]*domain.Asset, error)
```

Returns an ordered list — order matters for chaining (per SWN rules: "each form of asset must be fully used before the next type is triggered").

#### 3. Register in `ActionEngine`

In `internal/faction/engine/action_engine.go`, add a factory registration:

```go
engine.Register("UseAssetAbility", func() Action {
    return &actions.UseAssetAbility{Collector: collector}
})
```

#### 4. Unit tests (`engine/actions/use_asset_ability_test.go`)

Test cases to cover:
- Validate returns error when no A-flagged assets are available
- Validate returns nil when a usable A-flagged asset exists
- Resolve with a movement-only asset produces `AssetMoved` (and `CoinDelta` when `CoinCost > 0`)
- Resolve with a faction_test asset produces correct mutations on attacker win
- Resolve with a faction_test asset produces no effect mutations on tie/defender win
- Resolve with a nil-ability asset calls `ConfirmAbilityApplied` and produces no mutations

Use a fake collector (same pattern as attack tests) and deterministic roller.

### Commit message

```
feat: phase 3 — UseAssetAbility action with validate, inputs, resolve, output
```

<br/>
<br/>

## Phase 4 — TUI, Input Models, and Narration

**Goal:** Use Asset Ability is fully playable in the TUI with play-by-play log output.

### Key files to read

| File | Why |
|------|-----|
| `cmd/faction-manager/tui/turn_model.go` | Nine-state state machine; add new states and message types here |
| `cmd/faction-manager/tui/inputs/base_attackers.go` | Multi-select input model — reference for `SelectAbilityAssets` |
| `cmd/faction-manager/tui/inputs/expand_influence.go` | Wizard with step-skipping — reference for multi-step ability |
| `cmd/faction-manager/tui/narrate.go` | Per-action narrators — add `narrateUseAssetAbility` here |
| `internal/faction/engine/input_collector.go` | Full `InputCollector` interface — `TUICollector` must implement all methods |

### New TUI states

Add to the state machine enum in `turn_model.go`:

| State | Triggers | Exits to |
|---|---|---|
| `stateSelectAbilityAssets` | action = UseAssetAbility | goroutine starts; → `stateActionResult` |
| `stateAbilityMoveDestination` | movement step mid-Resolve | goroutine unblocks; → back to resolution |
| `stateAbilityFactionTestTarget` | faction_test step mid-Resolve | goroutine unblocks; → back to resolution |
| `stateAbilityConfirmApplied` | nil-ability asset mid-Resolve | goroutine unblocks; → back to resolution |

### Goroutine/channel bridge

Follows the same pattern as `ConfirmRedirectToBase` (Attack) and `ConfirmRivalFreeAttack` (Expand Influence). Add to `TurnModel`:

```go
abilityMoveCh         chan string           // response for SelectMoveDestination
abilityFactionTestCh  chan *domain.Faction  // response for SelectFactionTestTarget
abilityConfirmCh      chan bool             // response for ConfirmAbilityApplied
```

Each `TUICollector` method sends a typed message to the BubbleTea event loop and blocks on the corresponding channel.

### New message types

```go
type AbilityMoveMsg struct {
    Asset  *domain.Asset
    Worlds []string
}
type AbilityFactionTestMsg struct {
    Asset      *domain.Asset
    Effect     domain.AbilityEffectType
    Candidates []*domain.Faction
}
type AbilityConfirmMsg struct {
    Asset *domain.Asset
    Def   *domain.AssetDefinition
}
```

### New input models (`cmd/faction-manager/tui/inputs/`)

| File | Model | Notes |
|---|---|---|
| `ability_assets.go` | Ordered multi-select list of A-flagged assets | Show asset name + category abbrev; selected order determines resolution order |
| `move_destination.go` | Single-select world list | Worlds sorted alphabetically; "Astral Sea" always last |
| `faction_test_target.go` | Single-select faction list | Show faction name only |
| `ability_confirm.go` | Yes/No confirm | Show asset name + full description text |

### `narrateUseAssetAbility`

Add to `cmd/faction-manager/tui/narrate.go`. Called before `Mutation.Apply` (same pattern as other narrators). Derive output from the mutations list:

- `AssetMoved` → `"<AssetName> relocated to <ToLocation>"`
- `CoinDelta` (negative, acting faction) → `"Cost: <N> Coin"`
- `CoinDelta` (negative, other faction) → `"<FactionName> lost <N> Coin"`
- `CoinDelta` (positive, acting faction) → `"<FactionName> gained <N> Coin"`
- `AssetStealthCleared` → `"<AssetName> revealed"`
- No mutations (GM fallback) → `"Ability applied (GM adjudicated)"`

### Commit message

```
feat: phase 4 — TUI states, input models, and narration for Use Asset Ability
```

<br/>
<br/>

## Post-Phase 4 Checklist

Before merging to `main`:

- [ ] All tests pass (`go test ./...` from repo root)
- [ ] Manual smoke test with `--campaign test` campaign: trigger Use Asset Ability with a movement asset, a faction_test asset, and a deferred asset
- [ ] Senior engineer code review (per CLAUDE.md)
- [ ] Update `docs/dev-journal-factions.md` Progress section
- [ ] Update `docs/decisions-log.md` with Phase 2–4 decisions
- [ ] Version bump assessment: Use Asset Ability completes all nine actions → minor bump (`v0.12.0`)

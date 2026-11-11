# Input Collector Refactor — Flatten Embedded Hierarchy

> Context: identified during `feature/movement-redesign` Effort 2 Phase 3 implementation.

**Problem:** `engine.InputCollector` is assembled by embedding sub-package collector interfaces (`action.Collector`, `hooks.Collector`), which propagates sub-package contracts upward into the engine boundary. `hooks.Collector` is embedded twice. `ability/steps/collector.go` is a verbatim duplicate of `ability/collector.go`.

**Resolution:** `engine.InputCollector` becomes a flat, explicitly-declared interface that owns the full GM input contract. Sub-package interfaces (`action.Collector`, `hooks.Collector`, `ability.Collector`) survive as internal narrow views used by their sub-engines — they no longer propagate upward via embedding. Adding `SelectMovementDecisions` is bundled into this commit as the first consumer of the new shape.

One commit: `refactor: flatten InputCollector and add SelectMovementDecisions`

After this commit: `go build ./...` passes. `engine.InputCollector` has no embedded types. `ability/steps/collector.go` is deleted. `SelectMovementDecisions` is stubbed in all implementations.

<br/>

## Task 1 — `internal/faction/engine/world/movement.go` *(new file)*

Create `movement.go` in the `world` package. Move `TickMovementOrders` here from `world.go`. Add `MovementDecision` and `MovementDecisionKind`. All movement-related logic in the `world` package lives here — types, tick, and (in Commit 3) `buildMovementMutations`.

```go
type MovementDecisionKind int

const (
    MovementDecisionIssue  MovementDecisionKind = iota
    MovementDecisionRevise
    MovementDecisionCancel
)

type MovementDecision struct {
    AssetID     string
    Kind        MovementDecisionKind
    Destination *domain.Location // nil for Cancel
}

func (engine *WorldEngine) TickMovementOrders(...) { /* moved from world.go */ }
```

<br/>

## Task 2 — `internal/faction/engine/collector.go`

Replace all embedded types with explicit method declarations. Reference `world.MovementDecision` for `SelectMovementDecisions`.

```go
type InputCollector interface {
    AwaitCheckpoint(phase string) error

    // Top-level phase inputs
    SelectAction(faction *domain.Faction, available []action.Action) (action.Action, error)
    SelectStatRaise(faction *domain.Faction, eligible []domain.FactionStat) (*domain.FactionStat, error)
    SelectMovementDecisions(faction *domain.Faction, eligible []*domain.Asset) ([]world.MovementDecision, error)

    // Action sub-phase (previously via action.Collector embed)
    SelectAsset(assets []*domain.Asset, rulebook *rulebook.Rulebook) (*domain.Asset, error)
    SelectRepairOrders(faction *domain.Faction, damaged []*domain.Asset, rulebook *rulebook.Rulebook) ([]action.RepairOrder, error)
    SelectBuyOrder(purchasablePerWorld map[string][]*domain.AssetDefinition) (action.BuyOrder, error)
    SelectRefitOrder(options []action.RefitOption, rulebook *rulebook.Rulebook) (action.RefitOrder, error)
    SelectAttackers(eligible []*domain.Asset, rulebook *rulebook.Rulebook) ([]*domain.Asset, error)
    SelectDefender(attacker *domain.Asset, eligible []*domain.Asset, rulebook *rulebook.Rulebook) (*domain.Asset, error)
    ConfirmRedirectToBase(defenderFaction *domain.Faction, base *domain.Base, damage int) (bool, error)
    SelectExpandInfluenceOrder(faction *domain.Faction, factionState *state.FactionState, eligibleNewBaseWorlds []string) (action.ExpandInfluenceOrder, error)
    ConfirmRivalFreeAttack(rival *domain.Faction, rivalRoll, factionRoll int) (bool, error)
    SelectBaseAttackers(rival *domain.Faction, eligible []*domain.Asset, rulebook *rulebook.Rulebook) ([]*domain.Asset, error)
    SelectAbilityAssets(faction *domain.Faction, candidates []*domain.Asset, rulebook *rulebook.Rulebook) ([]*domain.Asset, error)
    ConfirmAbilityApplied(asset *domain.Asset, def *domain.AssetDefinition) (bool, error)
    SelectBribeTarget(faction *domain.Faction, factionState *state.FactionState) (*domain.Base, int, error)
    SelectSeizeTarget(faction *domain.Faction, factionState *state.FactionState) (string, error)

    // Hooks (previously via hooks.Collector embed)
    SelectModifiers(offers []hooks.ModifierOffer) []hooks.ModifierOffer
    ConfirmReroll(directive hooks.RerollDirective) bool

    // Ability steps (previously via ability.Collector embed)
    SelectMoveDestination(asset *domain.Asset, worlds []string) (string, error)
    SelectFactionTestTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Faction) (*domain.Faction, error)
}
```

<br/>

## Task 3 — `internal/faction/engine/testharness/collector.go`

Add `SelectMovementDecisionsFn` field and `SelectMovementDecisions` method (no-op returning `nil, nil`). No other changes — `ScriptedCollector` already implements every method explicitly.

```go
SelectMovementDecisionsFn func(*domain.Faction, []*domain.Asset) ([]world.MovementDecision, error)

func (c *ScriptedCollector) SelectMovementDecisions(faction *domain.Faction, eligible []*domain.Asset) ([]world.MovementDecision, error) {
    if c.SelectMovementDecisionsFn != nil {
        return c.SelectMovementDecisionsFn(faction, eligible)
    }
    return nil, nil
}
```

<br/>

## Task 4 — `internal/faction/engine/orchestrator.go`

Replace the current `collector.CollectMovementDecisions(...)` call (wrong signature, does not exist) with `prepareMovementDecisions`. Add stubs for the helpers filled in by Effort 2 Phase 3 Commit 3:

```go
func prepareMovementDecisions(
    faction *domain.Faction,
    collector InputCollector,
    worldEngine *world.WorldEngine,
    rulebook *rulebook.Rulebook,
) ([]domain.Mutation, error) {
    eligible := eligibleMovableAssets(faction, rulebook)
    if len(eligible) == 0 {
        return nil, nil
    }
    decisions, err := collector.SelectMovementDecisions(faction, eligible)
    if err != nil {
        return nil, fmt.Errorf("selecting movement decisions: %w", err)
    }
    if len(decisions) == 0 {
        return nil, nil
    }
    return worldEngine.BuildMovementMutations(decisions, faction, rulebook)
}

func eligibleMovableAssets(faction *domain.Faction, rulebook *rulebook.Rulebook) []*domain.Asset {
    return nil // filled in Effort 2 Phase 3 Commit 3
}
```

`BuildMovementMutations` is a method stub on `WorldEngine` in `movement.go` (also filled in Commit 3).

<br/>

## Task 5 — Build check

`go build ./...` must pass. The compiler will surface any `InputCollector` implementations missing `SelectMovementDecisions` — fix each with a no-op stub.

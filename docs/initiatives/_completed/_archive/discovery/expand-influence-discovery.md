# Expand Influence — Discovery & Implementation Plan

> Branch: `feature/expand-influence`

<br/>

## Overview

Expand Influence lets a faction establish a new Base of Influence on a world or reinforce an existing one. New bases trigger a contested roll against rivals present on that world; ties or losses give those rivals the option of a free attack directly against the new base.

**Complexity:** Medium-High — Coin transaction, contested roll, optional rival free-attack sub-flow.

<br/>

## Inputs

- Faction performing the action
- Mode: new Base of Influence or reinforce existing
- World selected
- If reinforce: target base and sub-mode (heal or expand)
- HP amount to purchase or restore

<br/>

## Resolution Steps

### New Base of Influence

1. GM selects world (must have at least one of the faction's assets there)
2. GM selects HP amount to purchase (cost: 1 Coin per HP; max: `faction.MaxHP`)
3. Deduct Coin from faction balance
4. Place new Base on the world, flagged `Ready: false` (inactive until next turn)
5. Roll contested check: faction rolls `1d10 + Cunning`; each rival faction with assets on that world rolls the same
6. For each rival that ties or beats the faction's roll: prompt GM whether that rival makes a free attack against the new Base
7. For each confirmed free attack: resolve via `baseAttack` sub-flow (see below)

### Reinforce — Heal

1. GM selects an existing non-homeworld base with `CurrentHP < MaxHP`
2. GM selects HP amount to restore (cost: 1 Coin per HP; capped at `MaxHP - CurrentHP`)
3. Deduct Coin; increase `CurrentHP` only — `MaxHP` unchanged

### Reinforce — Expand

1. GM selects an existing non-homeworld base with `MaxHP < faction.MaxHP`
2. GM selects HP amount to add (cost: 1 Coin per HP; capped at `faction.MaxHP - base.MaxHP`)
3. Deduct Coin; increase both `MaxHP` and `CurrentHP` by the selected amount

<br/>

## Base Attack Sub-Flow

An unexported `baseAttack` struct in `expand_influence.go`. Not registered with the action engine — only invoked from within `ExpandInfluence.Resolve`.

### Inputs
- Rival faction
- Eligible rival assets on that world (non-stealthy, alive, not inactive)
- Target base (the newly placed one)
- `Roller`

### Resolution
1. GM selects which of the rival's eligible assets attack the base (via `SelectBaseAttackers`)
2. For each attacker:
   - Roll attack: `1d10 + attacker's relevant stat`
   - Roll defense: `1d10 + owning faction's Cunning`
   - Attacker wins (strictly greater) or ties: base takes attacker's listed damage; emit `BaseHPDelta + FactionHPDelta`; if HP ≤ 0 emit `BaseDestroyed`
   - Defender wins: no damage
   - No counterattack — bases do not counter

<br/>

## Outputs

- New Base placed on world (`BaseAdded`) or existing Base HP changed (`BaseHealed` / `BaseExpanded`)
- Faction Coin reduced (`CoinDelta`)
- Rival free-attack outcomes: `BaseHPDelta`, `FactionHPDelta`, `BaseDestroyed` as applicable

<br/>

## New Mutations Required

| Mutation | Fields | Behavior |
|----------|--------|----------|
| `BaseAdded` | `FactionID`, `Base` | Appends base to `faction.Bases` |
| `BaseHealed` | `FactionID`, `BaseID`, `Delta` | Increases `CurrentHP` only; capped at `EffectiveMaxHP` |
| `BaseExpanded` | `FactionID`, `BaseID`, `Delta` | Increases both `MaxHP` and `CurrentHP` |

`BaseHPDelta` remains attack-only (its paired `FactionHPDelta` semantics don't apply to healing).

<br/>

## New InputCollector Methods

```
SelectExpandInfluenceOrder(faction, factionState) (ExpandInfluenceOrder, error)
ConfirmRivalFreeAttack(rival *domain.Faction, rivalRoll, factionRoll int) (bool, error)
SelectBaseAttackers(rival *domain.Faction, eligible []*domain.Asset, rulebook) ([]*domain.Asset, error)
```

### New Order Types

```
ExpandInfluenceOrder {
    Mode     ExpandMode     // "new" | "reinforce"
    World    string
    BaseID   string         // reinforce only
    SubMode  ReinforceMode  // "heal" | "expand" — reinforce only
    HPAmount int
}

ExpandMode    = "new" | "reinforce"
ReinforceMode = "heal" | "expand"
```

<br/>

## Implementation Layers

### 1. `internal/faction/domain/mutation.go`
Add `BaseAdded`, `BaseHealed`, `BaseExpanded`.

### 2. `internal/faction/engine/mutation_engine.go`
Add `Apply` cases for the three new mutations.

### 3. `internal/faction/engine/input_collector.go`
Add three new methods and two new order types.

### 4. `internal/faction/engine/actions/expand_influence.go` *(new file)*
- `ExpandInfluence` struct — registered action, standard `Validate`/`Inputs`/`Resolve`/`Output` interface
- `baseAttack` struct — unexported, not registered, called from `Resolve`
- Base ID convention: `{factionID}-base-{world}-{suffix}` (monotonic suffix, same pattern as assets)

### 5. `cmd/faction-manager/tui/inputs/expand_influence_model.go` *(new file)*
Multi-step form: mode → world → (base + sub-mode if reinforce) → HP amount.

### 6. `cmd/faction-manager/tui/model.go`
- Add `ExpandInfluenceMsg` type for the mid-`Resolve` goroutine bridge (same pattern as `AttackRedirectMsg`)
- New TUI state `stateExpandInfluenceRivalAttack`
- Wire `SelectExpandInfluenceOrder` TUI sub-model into the action phase

### 7. `cmd/faction-manager/tui/narrate.go`
Add `narrateExpandInfluence` covering: new base, reinforce (heal vs expand), contested roll outcome, rival attacks.

### 8. `internal/faction/engine/action_engine.go`
Register `NewExpandInfluence(collector, roller)`.

<br/>

## Design Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | `baseAttack` is an unexported struct in `expand_influence.go`, not a registered action | Free attack is triggered by the engine mid-resolution, never by GM input; registering it would expose it incorrectly |
| 2 | Reinforce has two distinct sub-modes: Heal and Expand | Heal restores a damaged base without growing it; Expand grows capacity — different intent, different mutation shape |
| 3 | `BaseHPDelta` retained as attack-only; new `BaseHealed`/`BaseExpanded` mutations for reinforce | `BaseHPDelta` carries an implicit `FactionHPDelta` coupling (per SWN damage rules); healing should not trigger faction HP side effects |
| 4 | `ConfirmRivalFreeAttack` and `SelectBaseAttackers` use the same goroutine/channel bridge pattern as `ConfirmRedirectToBase` | Rival attacks fire mid-`Resolve` inside a goroutine; the TUI must receive async messages to render the prompts |

<br/>

## Notes

- `InputCollector` will reach 10 methods after this branch. The wide-interface shape was ratified during `simple-actions`; revisit after remaining actions land
- Homeworld bases are excluded from reinforce targets (they track faction MaxHP automatically)
- Validation for new base: faction has ≥1 Coin and ≥1 asset on at least one world
- Validation for reinforce heal: at least one non-homeworld base exists with `CurrentHP < MaxHP` and faction has ≥1 Coin
- Validation for reinforce expand: at least one non-homeworld base exists with `MaxHP < faction.MaxHP` and faction has ≥1 Coin

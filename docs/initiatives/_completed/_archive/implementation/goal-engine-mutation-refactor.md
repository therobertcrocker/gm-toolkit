# Goal Engine — Mutation Refactor

A targeted refactor guide for a single branch. Written for a fresh session — read this doc first, then the key files listed per step.

**Branch:** `refactor/goal-engine-mutations`

<br/>

## Context

The project invariant is that all state changes flow through `MutationEngine.Apply`. Actions honor this: every `Resolve` returns `[]domain.Mutation` and never writes to `FactionState` directly.

`goal_engine.go` currently violates this in two ways:

1. **Direct field writes** — `Progress`, `TurnsRemaining`, `ProcessPhase`, and `ActiveGoal = nil` are all set by hand inside goal handlers and `completeGoal`.
2. **Double-writes** — `completeGoal` sets `faction.ActiveGoal = nil` *and* returns a `GoalCompleted` mutation, which `MutationEngine.Apply` also clears. `checkLockPlanetarySeizure`'s occupation-failure path has the same pattern with `GoalAbandoned`.

Mutation types `GoalProgressed` and `GoalInitiated` already exist in `domain/mutation.go` and are already handled in `mutation_engine.go`'s `Apply` switch — but the goal engine never emits them; it writes `Progress` directly instead. Two state changes (`TurnsRemaining--` and phase transitions) have no mutation type at all.

**Goal:** make `goal_engine.go` read-only with respect to `FactionState`. All state changes must be expressed as returned mutations and applied by `MutationEngine.Apply`.

<br/>

## Key Files

| File | Role |
|------|------|
| `internal/faction/domain/mutation.go` | Add two new mutation types; fix inaccurate `GoalProgressed` comment |
| `internal/faction/engine/mutation_engine.go` | Add `Apply` cases for the two new types |
| `internal/faction/engine/goal_engine.go` | Remove all direct writes; emit mutations instead |
| `internal/faction/narrative/digest/build.go` | Silent-drop cases for goal bookkeeping mutation types |

<br/>

## Step 1 — New Mutation Types (`domain/mutation.go`)

The existing `GoalProgressed` comment claims it "has no direct mechanical effect" — this is wrong; `mutation_engine.go` already applies it to `Progress`. Fix the comment.

Add two new types after `GoalProgressed`:

```go
// GoalTurnsTick decrements ActiveGoal.TurnsRemaining by 1.
type GoalTurnsTick struct {
    FactionID string `json:"faction_id"`
    GoalID    string `json:"goal_id"`
    Cause     string `json:"cause"`
}

func (mutation GoalTurnsTick) Type() string { return "goal_turns_tick" }

// GoalPhaseAdvanced records a phase transition in a multi-phase goal,
// setting both ProcessPhase and TurnsRemaining atomically.
type GoalPhaseAdvanced struct {
    FactionID      string `json:"faction_id"`
    GoalID         string `json:"goal_id"`
    ProcessPhase   int    `json:"process_phase"`
    TurnsRemaining int    `json:"turns_remaining"`
    Cause          string `json:"cause"`
}

func (mutation GoalPhaseAdvanced) Type() string { return "goal_phase_advanced" }
```

<br/>

## Step 2 — Apply Cases (`mutation_engine.go`)

Add before the `default` panic in `Apply`:

```go
case domain.GoalTurnsTick:
    if faction, ok := factionState.Factions[v.FactionID]; ok {
        if faction.ActiveGoal != nil && faction.ActiveGoal.GoalID == v.GoalID {
            faction.ActiveGoal.TurnsRemaining--
        }
    }

case domain.GoalPhaseAdvanced:
    if faction, ok := factionState.Factions[v.FactionID]; ok {
        if faction.ActiveGoal != nil && faction.ActiveGoal.GoalID == v.GoalID {
            faction.ActiveGoal.ProcessPhase = v.ProcessPhase
            faction.ActiveGoal.TurnsRemaining = v.TurnsRemaining
        }
    }
```

<br/>

## Step 3 — `goal_engine.go` Rewrites

Work through each function in order. The function signatures do not change — only the bodies.

### `completeGoal`

Remove `faction.ActiveGoal = nil`. `GoalCompleted` in `Apply` already handles it. The `goalID` local is still needed.

```go
func completeGoal(faction *domain.Faction, xp int) []domain.Mutation {
    goalID := faction.ActiveGoal.GoalID
    return []domain.Mutation{
        domain.GoalCompleted{
            FactionID: faction.ID,
            GoalID:    goalID,
            XPAwarded: xp,
            Cause:     "goal_completed",
        },
        domain.XPAwarded{
            FactionID: faction.ID,
            Amount:    xp,
            Cause:     "goal_completed",
        },
    }
}
```

### `checkLockChangeHomeworld`

Compute `newTurns` locally instead of decrementing directly. Emit `GoalTurnsTick` in both branches; on completion, append the homeworld and goal mutations after it. Remove `faction.ActiveGoal = nil`.

```go
func checkLockChangeHomeworld(faction *domain.Faction) (GoalLock, []domain.Mutation) {
    newTurns := faction.ActiveGoal.TurnsRemaining - 1
    tick := domain.GoalTurnsTick{
        FactionID: faction.ID,
        GoalID:    faction.ActiveGoal.GoalID,
        Cause:     "change_homeworld_transit",
    }
    if newTurns == 0 {
        return GoalLock{Type: LockSkip}, []domain.Mutation{
            tick,
            domain.HomeworldChanged{
                FactionID: faction.ID,
                FromWorld: faction.Homeworld,
                ToWorld:   faction.ActiveGoal.TargetWorld,
                Cause:     "goal_completed",
            },
            domain.GoalCompleted{
                FactionID: faction.ID,
                GoalID:    faction.ActiveGoal.GoalID,
                XPAwarded: 0,
            },
        }
    }
    return GoalLock{Type: LockSkip}, []domain.Mutation{tick}
}
```

### `checkLockPlanetarySeizure`

Three direct writes to replace:

1. **Occupation failure** — remove `faction.ActiveGoal = nil`; `GoalAbandoned` in `Apply` handles it.
2. **Occupation countdown** — replace `goal.TurnsRemaining--` with a `GoalTurnsTick`; compute `newTurns` locally to decide the branch.
3. **Occupation completion** — remove `faction.ActiveGoal = nil`; prepend `GoalTurnsTick` before the existing `TagAdded`/`GoalCompleted`/`XPAwarded` mutations.

```go
func checkLockPlanetarySeizure(faction *domain.Faction, factionState *state.FactionState, rulebook *loader.Rulebook) (GoalLock, []domain.Mutation) {
    goal := faction.ActiveGoal
    if goal.ProcessPhase == 0 {
        return GoalLock{Type: LockNone}, nil
    }
    if goal.ProcessPhase == 1 {
        return GoalLock{Type: LockRestrictActions, AllowedActions: []string{"Attack"}}, nil
    }
    // Phase 2: occupation.
    if !factionHasUnstealthedAssetOn(faction, goal.TargetWorld) {
        return GoalLock{Type: LockNone}, []domain.Mutation{
            domain.GoalAbandoned{
                FactionID: faction.ID,
                GoalID:    goal.GoalID,
                Cause:     "occupation_failed",
            },
        }
    }
    tick := domain.GoalTurnsTick{
        FactionID: faction.ID,
        GoalID:    goal.GoalID,
        Cause:     "planetary_seizure_occupation",
    }
    newTurns := goal.TurnsRemaining - 1
    if newTurns == 0 {
        xp := calcPlanetarySeizureXP(goal.TargetFactionID, factionState)
        var pgTag domain.Tag
        if t, ok := rulebook.Tags["T-011"]; ok {
            pgTag = *t
        }
        return GoalLock{Type: LockNone}, []domain.Mutation{
            tick,
            domain.TagAdded{FactionID: faction.ID, Tag: pgTag, Cause: "goal_completed"},
            domain.GoalCompleted{FactionID: faction.ID, GoalID: goal.GoalID, XPAwarded: xp},
            domain.XPAwarded{FactionID: faction.ID, Amount: xp, Cause: "goal_completed"},
        }
    }
    return GoalLock{Type: LockNone}, []domain.Mutation{tick}
}
```

### `progressPlanetarySeizure`

Replace the two direct field writes with a `GoalPhaseAdvanced` mutation. The function currently returns `nil` on success — after this change it returns a one-element slice.

```go
// No rivals remain — transition to occupation (3 turns).
return []domain.Mutation{
    domain.GoalPhaseAdvanced{
        FactionID:      actingFaction.ID,
        GoalID:         goal.GoalID,
        ProcessPhase:   2,
        TurnsRemaining: 3,
        Cause:          "planetary_seizure_phase_advance",
    },
}
```

### Progress handlers — general pattern

Each handler currently: (a) increments `Progress` directly, (b) checks the threshold against the updated value, (c) calls `completeGoal` (which also writes `ActiveGoal = nil`). 

New pattern: compute `newProgress` locally without writing, emit `GoalProgressed{Delta: delta}` when below threshold, prepend it before `completeGoal(...)` when at or above. Never write to `actingFaction.ActiveGoal.Progress`.

**`progressMilitaryConquest` / `progressCommercialExpansion` / `progressIntelligenceCoup`**

All three are identical in shape. Example for Military Conquest; apply the same transform to the other two (swap the `StatForce`/`Force` references for `StatWealth`/`Wealth` and `StatCunning`/`Cunning`):

```go
kills := countAssetKillsByCategory(...)
if kills == 0 {
    return nil
}
newProgress := actingFaction.ActiveGoal.Progress + kills
progressed := domain.GoalProgressed{
    FactionID: actingFaction.ID,
    GoalID:    actingFaction.ActiveGoal.GoalID,
    Delta:     kills,
    Cause:     "military_conquest",
}
if newProgress < actingFaction.Force {
    return []domain.Mutation{progressed}
}
return append([]domain.Mutation{progressed}, completeGoal(actingFaction, newProgress/2)...)
```

Cause strings for the other two: `"commercial_expansion"`, `"intelligence_coup"`.

**`progressBloodTheEnemy`**

```go
if damage == 0 {
    return nil
}
newProgress := actingFaction.ActiveGoal.Progress + damage
threshold := actingFaction.Force + actingFaction.Cunning + actingFaction.Wealth
progressed := domain.GoalProgressed{
    FactionID: actingFaction.ID,
    GoalID:    actingFaction.ActiveGoal.GoalID,
    Delta:     damage,
    Cause:     "blood_the_enemy",
}
if newProgress < threshold {
    return []domain.Mutation{progressed}
}
return append([]domain.Mutation{progressed}, completeGoal(actingFaction, 2)...)
```

**`progressPeaceableKingdom`**

The reset-to-zero case uses a negative delta. The increment-and-check case follows the general pattern.

```go
if attacked {
    if actingFaction.ActiveGoal.Progress == 0 {
        return nil
    }
    return []domain.Mutation{domain.GoalProgressed{
        FactionID: actingFaction.ID,
        GoalID:    actingFaction.ActiveGoal.GoalID,
        Delta:     -actingFaction.ActiveGoal.Progress,
        Cause:     "peaceable_kingdom_reset",
    }}
}
newProgress := actingFaction.ActiveGoal.Progress + 1
progressed := domain.GoalProgressed{
    FactionID: actingFaction.ID,
    GoalID:    actingFaction.ActiveGoal.GoalID,
    Delta:     1,
    Cause:     "peaceable_kingdom",
}
if newProgress < 4 {
    return []domain.Mutation{progressed}
}
return append([]domain.Mutation{progressed}, completeGoal(actingFaction, 1)...)
```

**`progressInsideEnemyTerritory`**

The loop currently increments `Progress` on each qualifying iteration. Replace with a local `gained` counter; emit one `GoalProgressed` after the loop.

```go
gained := 0
for _, mutation := range mutations {
    // ... existing qualification checks, increment gained instead of Progress ...
    gained++
}
if gained == 0 {
    return nil
}
newProgress := actingFaction.ActiveGoal.Progress + gained
progressed := domain.GoalProgressed{
    FactionID: actingFaction.ID,
    GoalID:    actingFaction.ActiveGoal.GoalID,
    Delta:     gained,
    Cause:     "inside_enemy_territory",
}
if newProgress < actingFaction.Cunning {
    return []domain.Mutation{progressed}
}
return append([]domain.Mutation{progressed}, completeGoal(actingFaction, 2)...)
```

**`progressWealthOfWorlds`**

```go
if spent == 0 {
    return nil
}
newProgress := actingFaction.ActiveGoal.Progress + spent
progressed := domain.GoalProgressed{
    FactionID: actingFaction.ID,
    GoalID:    actingFaction.ActiveGoal.GoalID,
    Delta:     spent,
    Cause:     "wealth_of_worlds",
}
if newProgress < 4*actingFaction.Wealth {
    return []domain.Mutation{progressed}
}
return append([]domain.Mutation{progressed}, completeGoal(actingFaction, 2)...)
```

<br/>

## Step 4 — Digest Builder (`narrative/digest/build.go`)

The `default` case appends `"unknown mutation: <type>"` to beat notes. Add silent-drop cases before `default` for all goal bookkeeping types that carry no narrative meaning. `goal_initiated` is already emitted by `seize_planet.go` but unhandled — this drop fixes that pre-existing gap too.

```go
case "goal_initiated":
    // dropped — bookkeeping only
case "goal_progressed":
    // dropped — bookkeeping only
case "goal_turns_tick":
    // dropped — bookkeeping only
case "goal_phase_advanced":
    // dropped — bookkeeping only
```

<br/>

## Verification

1. `go build ./...` from repo root — new types compile; `Apply` switch remains exhaustive.
2. `go test ./...` — existing tests must pass unchanged. Goal engine tests check for the presence of `GoalCompleted`/`XPAwarded` in returned slices; the handlers now prepend `GoalProgressed` before those, which is additive and does not break existing assertions. Confirm no new `panic: unhandled mutation type` failures.
3. Manual smoke — run a faction with Change Homeworld active through multiple cycles; confirm `TurnsRemaining` decrements turn-over-turn and the homeworld shifts on the final turn.
4. `narrate <cycle>` on a cycle with a goal completion — confirm no `"unknown mutation"` lines appear in the output.

# Goal Engine — Implementation Plan

A phased build guide for the `feature/goal-engine` branch. Written for a fresh session — read this doc first, then the key files listed per phase.

**Branch:** `feature/goal-engine`

<br/>

## Design Decisions (read before any phase)

These were ratified in the discovery session and must be followed precisely:

| # | Decision |
|---|----------|
| 1 | `Faction.Goal *Goal` is replaced by `Faction.ActiveGoal *ActiveGoal` — a richer struct that holds live progress state |
| 2 | TOML holds metadata only (name, description, difficulty label); all goal behavior lives in code as per-goal-type handlers |
| 3 | Change Homeworld is **not a registered action** — it is a goal with Difficulty 0 added to `goals.toml`; when selected, it prompts for destination world and hex distance |
| 4 | Seize Planet (action) → Planetary Seizure (goal) follows the Expand Influence pattern: the action initiates the process; the Goal Engine manages subsequent turns via lock state |
| 5 | Abandon Goal is a **registered action** — appears in the action menu, recalculates and deducts that turn's income, clears `ActiveGoal` |
| 6 | Goal Engine inspects the acting faction's mutation list **after** `ActionEngine.Run` returns and **before** `MutationEngine.Apply` fires (Option A) |
| 7 | All mutations gain attribution fields (e.g.): `CausedByFactionID string`, `Cause string` |
| 8 | `AssetStealthApplied` is emitted during **Buy Asset** when `Stealth` Cunning asset is purchased — not during Use Asset Ability |
| 9 | Lock state is communicated via `GoalEngine.CheckLock(faction, factionState) GoalLock` — called by the turn wizard before action selection (Option B) |
| 10 | Each goal type has its own `CalculateXP` formula function; XP for variable-difficulty goals is calculated at completion time, before any destructive mutations from the triggering event are applied |
| 11 | `Base.Influence int` is a new field; the Bribe (new) action is the only thing that modifies it; it has no mechanical effect beyond satisfying Wealth of Worlds |
| 12 | Inside Enemy Territory tracks stealth events (`AssetStealthApplied` on a world with a rival's Planetary Government tag) — only assets stealthed after the goal was adopted count |
| 13 | Change Homeworld goal completion (TurnsRemaining reaching 0) is handled inside `CheckLock`, not `UpdateProgress` — it fires at the start of the faction's locked turn, not after an action |
| 14 | Destroy the Foe XP must be calculated before the target faction is removed from state |

<br/>

## Domain Types Reference

### ActiveGoal

```go
type ActiveGoal struct {
    GoalID          string // references goal ID from goals.toml
    TargetFactionID string // Destroy the Foe
    TargetWorld     string // Planetary Seizure (target world), Change Homeworld (destination)
    Progress        int    // cumulative counter: kills, damage dealt, Coin spent, turns without attacking, stealthed assets
    ProcessPhase    int    // Seize Planet: 0 = combat, 1 = occupation
    TurnsRemaining  int    // Change Homeworld: turns until move completes; also used as occupation turn counter
}
```

### GoalLock

```go
type LockType int

const (
    LockNone           LockType = iota // unrestricted — normal flow
    LockSkip                           // skip faction entirely (Change Homeworld in-progress)
    LockRestrictActions                // restrict action menu (Seize Planet combat phase: Attack only)
)

type GoalLock struct {
    Type           LockType
    AllowedActions []string // action names; non-nil only when Type == LockRestrictActions
}
```

<br/>

## Goal Handler Summary

| Goal ID | Goal Name | Completion Trigger | Progress Counter | XP Formula |
|---------|-----------|-------------------|-----------------|------------|
| G-001 | Military Conquest | `AssetRemoved` (cause=attack, category=Force, causedBy=acting faction) | rival Force kills | `floor(kills / 2)` |
| G-002 | Commercial Expansion | `AssetRemoved` (cause=attack, category=Wealth, causedBy=acting faction) | rival Wealth kills | `floor(kills / 2)` |
| G-003 | Intelligence Coup | `AssetRemoved` (cause=attack, category=Cunning, causedBy=acting faction) | rival Cunning kills | `floor(kills / 2)` |
| G-004 | Planetary Seizure | Seize Planet action initiates; phases managed via `CheckLock` | phase + occupation turns | `max(1, floor(avg(target F/C/W) / 2))` |
| G-005 | Expand Influence | `BaseAdded` (causedBy=acting faction, world not previously held) | — (single event) | 1, +1 if contested |
| G-006 | Blood the Enemy | `AssetHPDelta` + `BaseHPDelta` (negative delta, causedBy=acting faction, rival-owned) | cumulative HP damage | 2 |
| G-007 | Peaceable Kingdom | No Attack action taken — 4 consecutive turns | turns without attacking | 1 |
| G-008 | Destroy the Foe | Target faction HP reaches 0 | — (single event) | `1 + floor(avg(target F/C/W))` |
| G-009 | Inside Enemy Territory | `AssetStealthApplied` on a world with a rival's Planetary Government tag | stealthed asset count (post-adoption) | 2 |
| G-010 | Invincible Valor | `AssetRemoved` (cause=attack, Force category, minRating > acting faction's Force) | — (single event) | 2 |
| G-011 | Wealth of Worlds | `InfluenceDelta` (Bribe action, causedBy=acting faction) | cumulative Coin spent on bribes | 2 |
| G-012 | Change Homeworld | `TurnsRemaining` countdown via `CheckLock` | — (countdown) | 0 |

<br/>

## New Mutations

### New Types

| Mutation Type | Fields | Purpose |
|---------------|--------|---------|
| `AssetStealthApplied` | `FactionID`, `AssetID` | Asset became stealthed via Buy Asset (Stealth-type purchase) |
| `GoalAbandoned` | `FactionID`, `GoalID` | Faction abandoned their goal; income deducted via separate `CoinDelta` |
| `GoalCompleted` | `FactionID`, `GoalID`, `XPAwarded int` | Goal resolved successfully |
| `XPAwarded` | `FactionID`, `Amount int` | XP added to faction |
| `HomeworldChanged` | `FactionID`, `FromWorld`, `ToWorld` | Homeworld updated on Change Homeworld completion |
| `TagAdded` | `FactionID`, `TagID` | Tag granted to faction (Planetary Government on Seize Planet) |
| `InfluenceDelta` | `FactionID`, `BaseID`, `Delta int` | Influence added to a Base (Bribe action) |

### Updated Existing Types

- `Cause string` — one of `"attack"`, `"sell"`, `"refit"`, `"bookkeeping"`
- `CausedByFactionID string` — faction whose action caused the removal. 

All existing call sites that construct mutations must be updated to populate the new fields.

<br/>

## Turn Flow Integration

```
Per-faction turn:

  1. GoalEngine.CheckLock(faction, factionState)
     → LockSkip:
         decrement faction.ActiveGoal.TurnsRemaining
         if TurnsRemaining == 0:
             emit HomeworldChanged + GoalCompleted (XP=0) + clear ActiveGoal
         apply those mutations; advance to next faction (no bookkeeping, no action)
     → LockRestrictActions:
         filter AvailableActions to GoalLock.AllowedActions only; continue normally
     → LockNone:
         normal flow

  2. ApplyBookkeeping — unchanged

  3. Action selection (filtered if LockRestrictActions)
     Action runs → mutation list returned

  4. GoalEngine.UpdateProgress(actingFactionID, mutations, factionState, rulebook)
     → inspect mutations for goal-relevant events
     → update faction.ActiveGoal.Progress
     → check completion threshold
     → if complete: calculate XP (before destructive mutations apply), emit GoalCompleted + XPAwarded
       (+ TagAdded for Planetary Seizure)
     → return supplemental goal mutations appended to action mutations

  5. MutationEngine.Apply(all mutations combined)

  6. History record + state save — unchanged
```

<br/>

## Phase 1 — Domain Foundation

**Goal:** data model only — no behavior. Existing tests must still pass after this phase.

### Files to modify

**`internal/faction/domain/faction.go`**
- Add `ActiveGoal` struct (fields above)
- Replace `Goal *Goal` field on `Faction` with `ActiveGoal *ActiveGoal` (TOML tag: `active_goal`)
- Keep `Goal` and `Tag` domain types as-is — still used by loader and display

**`internal/faction/domain/mutation.go`**
- Add all seven new mutation types (fields and `Type()` string methods)
- Add `Cause string` and `CausedByFactionID string` to each existing mutation type.


**`internal/faction/domain/base.go`**
- Add `Influence int` field with `toml:"influence"` tag

**`internal/faction/data/goals.toml`**
- Add G-012 Change Homeworld entry

**Fix compilation breakage** — anywhere `Faction.Goal` is read must be updated to `Faction.ActiveGoal`:
- `cmd/faction-manager/commands/faction/cmd.go` — faction list display
- `cmd/faction-manager/commands/faction/create.go` — faction create (set `ActiveGoal` from selected goal)
- `cmd/faction-manager/tui/model.go` — faction info panel

**`cmd/faction-manager/commands/faction/create.go`** — faction create sets `ActiveGoal.GoalID` from the selected `Goal.ID`; `ActiveGoal.Progress`, `TargetFactionID`, etc. are zero values.

**Update mutation call sites** — all existing `AssetRemoved`, `AssetHPDelta`, `BaseHPDelta` construction sites need the new fields. Cause values:
- Attack action: `Cause = "attack"`, `CausedByFactionID = faction.ID`
- Sell Asset: `Cause = "sell"`, `CausedByFactionID = faction.ID`
- Refit Asset: `Cause = "refit"`, `CausedByFactionID = faction.ID`
- Bookkeeping (asset lost to maintenance): `Cause = "bookkeeping"`, `CausedByFactionID = ""`

**`internal/faction/engine/mutation_engine.go`**
- Add `Apply` cases for all seven new mutation types
- `GoalCompleted` / `GoalAbandoned`: clear `faction.ActiveGoal`
- `XPAwarded`: increment `faction.XP`
- `HomeworldChanged`: update `faction.Homeworld`
- `TagAdded`: append tag to `faction.Tags` (look up from `Rulebook.Tags` by ID)
- `InfluenceDelta`: find base by ID across faction bases, add delta to `Influence`
- `AssetStealthApplied`: set `asset.Stealthy = true`

**Commit:** `feat: phase 1 — ActiveGoal domain type, new mutations, attribution fields`

<br/>

## Phase 2 — GoalEngine Core

**Goal:** `GoalEngine` struct with `CheckLock` and `UpdateProgress`; all 12 goal handlers; wired into `Engine`. Unit-testable in isolation.

### Files to create

**`internal/faction/engine/goal_engine.go`**

```go
type GoalEngine struct{}

func NewGoalEngine() *GoalEngine { ... }

// CheckLock evaluates the faction's active goal and returns the appropriate
// lock state. For LockSkip (Change Homeworld), it also decrements TurnsRemaining
// and emits completion mutations when the countdown reaches zero.
func (ge *GoalEngine) CheckLock(faction *domain.Faction, factionState *state.FactionState) (GoalLock, []domain.Mutation)

// UpdateProgress inspects the acting faction's mutation list for goal-relevant
// events, updates ActiveGoal.Progress, and returns supplemental goal mutations
// (GoalCompleted, XPAwarded, TagAdded) if the goal is now complete.
// Must be called before MutationEngine.Apply so destructive mutations have not
// yet fired (Destroy the Foe XP calculation reads live target faction state).
func (ge *GoalEngine) UpdateProgress(
    actingFactionID string,
    mutations []domain.Mutation,
    factionState *state.FactionState,
    rulebook *loader.Rulebook,
) []domain.Mutation
```

Each of the 12 goals gets an unexported handler following this shape:

```go
// progressMilitaryConquest inspects mutations for rival Force asset destructions
// caused by the acting faction and returns any supplemental mutations on completion.
func progressMilitaryConquest(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, rulebook *loader.Rulebook) []domain.Mutation
```

Goal target thresholds:
- G-001/002/003: `Progress >= faction stat rating` (Force/Wealth/Cunning)
- G-006: `Progress >= faction.Force + faction.Cunning + faction.Wealth`
- G-007: `Progress >= 4` (turns without attacking)
- G-009: `Progress >= faction.Cunning`
- G-011: `Progress >= 4 * faction.Wealth`

**CheckLock return values:**
- Change Homeworld (G-012) active: `LockSkip`
- Planetary Seizure (G-004) active + `ProcessPhase == 0` (combat): `LockRestrictActions{AllowedActions: ["Attack"]}`
- Planetary Seizure (G-004) active + `ProcessPhase == 1` (occupation): `LockNone`
- All other goals: `LockNone`

**`internal/faction/engine/core.go`**
- Add `Goal *GoalEngine` field
- Wire: `e.Goal = NewGoalEngine()` in `New()`

**Commit:** `feat: phase 2 — GoalEngine with CheckLock and UpdateProgress, all 12 goal handlers`

<br/>

## Phase 3 — Turn Flow and TUI Integration

**Goal:** `CheckLock` wired into the turn wizard; goal selection extended for goals that need inputs; `AbandonGoal` action registered.

### Files to modify

**`cmd/faction-manager/tui/model.go`**

Add a `stateGoalLocked` TUI state for when `CheckLock` returns `LockSkip` — display a "Moving homeworld: N turns remaining" message and auto-advance after the GM acknowledges (Enter). This follows the same press-Enter pattern as the skip prompt.

When `CheckLock` returns `LockRestrictActions`, pass `AllowedActions` through to the action selection phase so `AvailableActions` is filtered against it.

**`cmd/faction-manager/tui/phases/bookkeeping.go`** (or a new `goal_select.go` phase)

Goal selection fires when `faction.ActiveGoal == nil`. The goal selection model must handle two goal types that require additional inputs after selection:
- **Change Homeworld (G-012):** prompt for destination world (derived from faction's non-homeworld Bases), then hex distance (integer input). Calculate `TurnsRemaining = distance + 1`.
- **Planetary Seizure (G-004):** prompt for target world (any world where the faction has an unstealthed asset and a rival has an unstealthed asset).

All other goals set `ActiveGoal.GoalID` with zero-value Progress/Target fields.

**`internal/faction/engine/actions/abandon_goal.go`**

```
Validate: faction.ActiveGoal != nil
Inputs:   none
Resolve:  recalculate income (floor(Wealth/2) + floor((Force+Cunning)/4))
          emit CoinDelta{Delta: -income} + GoalAbandoned{GoalID: faction.ActiveGoal.GoalID}
Output:   [CoinDelta, GoalAbandoned]
```

Register in `cmd/faction-manager/commands/turn/cmd.go` (wherever actions are registered).

**Commit:** `feat: phase 3 — TUI CheckLock integration, goal selection inputs, AbandonGoal action`

<br/>

## Phase 4 — New Actions

**Goal:** `Bribe` and `SiezePlanet` registered; `BuyAsset` extended for Stealth asset detection.

### Files to create

**`internal/faction/engine/actions/bribe.go`**

```
Validate: faction has at least one Base; faction.Coin >= 1
Inputs:   SelectBribeTarget — select a Base owned by this faction (any world); select Coin amount (1 to faction.Coin)
Resolve:  emit CoinDelta{Delta: -amount} + InfluenceDelta{BaseID, Delta: amount}
Output:   [CoinDelta, InfluenceDelta]
```

Add `SelectBribeTarget` to `InputCollector` interface (`internal/faction/engine/input_collector.go`).

**`internal/faction/engine/actions/seize_planet.go`**

```
Validate: faction has at least one unstealthed asset on a world where a rival also has an unstealthed asset;
          faction.ActiveGoal == nil OR (ActiveGoal.GoalID == G-004 AND ActiveGoal.TargetWorld matches)
Inputs:   SelectSiezeTarget — select target world (pre-filtered to valid worlds)
Resolve:  set faction.ActiveGoal = &ActiveGoal{GoalID: "G-004", TargetWorld: selected, ProcessPhase: 0}
          emit no mutations — the Goal Engine manages subsequent turns
Output:   [] (empty — goal state is set inline; Goal Engine takes over next turn)
```

Note: `SiezePlanet.Resolve` mutates `faction.ActiveGoal` directly rather than via a mutation, because goal initiation is not something that needs to be recorded in history as a standalone mutation — the `GoalCompleted` mutation at the end of the process is the meaningful history event.

### Files to modify

**`internal/faction/engine/actions/buy_asset.go`**

After emitting `AssetAdded`, check if the purchased asset is a Stealth-type Cunning asset (identify by checking the asset definition's description or a new flag — TBD during implementation based on what's in the TOML data). If so, find eligible Special Forces units owned by the same faction on the same world and emit `AssetStealthApplied` for each eligible unit.

**`internal/faction/engine/input_collector.go`**

Add:
- `SelectBribeTarget(faction *domain.Faction, factionState *state.FactionState) (*domain.Base, int, error)`
- `SelectSiezeTarget(faction *domain.Faction, factionState *state.FactionState) (string, error)`

**TUI collector and input models** — follow the existing pattern in `cmd/faction-manager/tui/inputs/` for each new prompt.

**Commit:** `feat: phase 4 — Bribe action, SiezePlanet action, BuyAsset Stealth extension`

<br/>

## Phase 5 — Narration and Polish

**Goal:** goal events surfaced in the play-by-play log; cycle summary updated; dev journal and decisions log updated.

**`cmd/faction-manager/tui/narrate.go`**

Add narrators for:
- `GoalCompleted` — "Faction completed [Goal Name] and earned N XP"
- `GoalAbandoned` — "Faction abandoned [Goal Name] (income forfeited)"
- `HomeworldChanged` — "Faction relocated homeworld from X to Y"
- `TagAdded` — "Faction gained the Planetary Government tag on [World]"
- `XPAwarded` — surfaced as part of `GoalCompleted` narration, not standalone

**`cmd/faction-manager/tui/model.go` / cycle summary**

The `stateGoalLocked` screen (Change Homeworld) should show the remaining turn count and destination world in the left faction panel, not just a blank action slot.

**docs/**
- Update `docs/dev-journal-factions.md` — progress and decisions
- Update `docs/decisions-log.md` — all ratified decisions from discovery
- Update `docs/tracking/turn-engine-journal.md` — mark Feature #7 as Complete; close Open Question #1
- Rewrite `docs/discovery/goal-engine-discovery.md` — the old doc is superseded; replace with the design as built

**Commit:** `feat: phase 5 — goal narration, polish, and doc updates`

<br/>

## Verification

- `go test ./...` from repo root — must pass after every phase
- Manual (`--campaign test`): full turn cycle fires goal selection prompt when `ActiveGoal` is nil; progress updates visible in narration log; Peaceable Kingdom completes after 4 turns without Attack
- Manual: Change Homeworld goal — select destination + distance, faction skips N turns, homeworld updates on completion
- Manual: Seize Planet — take action, goal locks to Attack-only, transitions to occupation phase, awards Planetary Government tag on completion
- Manual: Abandon Goal — deducts correct income amount, clears goal
- Manual: Bribe action — increments `Base.Influence`; Wealth of Worlds progress counter increments

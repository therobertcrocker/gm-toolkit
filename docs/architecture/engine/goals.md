# Goals

> **Code:** `internal/faction/engine/goal/` (`goal.go`, `goals/`, `locks/locks.go`)

## Purpose

A faction pursues one active goal at a time; the goal engine decides what that
goal *does* to a turn — whether it constrains the faction's choices, and whether
this turn's events advanced or completed it. Each goal is a small `Handler`
registered by goal ID; the engine looks one up for the faction's active goal and
asks it two questions per turn. Like every sub-engine it emits mutations and
applies nothing.

## Shape

### The `Handler` contract and registry

```go
type Handler interface {
	GoalID() string
	CheckLock(faction, factionState, rulebook) (locks.GoalLock, []domain.Mutation)
	UpdateProgress(actingFaction, mutations, factionState, rulebook, index) []domain.Mutation
}
```

`GoalEngine` holds a `map[string]Handler` keyed by `GoalID()`, populated in `New`
with the twelve standard goal handlers (Military Conquest, Planetary Seizure,
Change Homeworld, …). This is the same handler-registry-keyed-by-data-ID pattern
the tag and effect engines use. A faction's `ActiveGoal.GoalID` with no
registered handler is a **data-only goal** — present in `goals.toml` for flavor,
silently skipped, not an error (the fence-signed twin of the tag/effect skip).

### The lock taxonomy

`CheckLock` returns a `locks.GoalLock` from the leaf `locks` package:

```go
const (
	LockNone            // normal flow
	LockSkip            // skip the faction's turn entirely
	LockRestrictActions // limit the action menu to AllowedActions
)
```

`LockSkip` short-circuits the whole turn (a faction mid-transit during Change
Homeworld does nothing); the orchestrator's goal-lock phase carries a `LockSkip`
straight to finish. `LockRestrictActions` narrows the action menu to a named set
(Planetary Seizure's combat phase allows only `Attack`). `LockNone` is the
common case.

### Two advancement paths

A goal can advance on either of two distinct beats, and the split is the heart of
this engine:

- **`CheckLock` — the goal-lock phase, time/condition-driven.** Called after
  bookkeeping and movement, immediately *before* action selection. Besides
  returning the lock, it emits the mutations for goals that advance by the
  *passage of a turn* rather
  than by an action: `GoalTurnsTick` countdowns, `HomeworldChanged` on arrival,
  occupation completions. Change Homeworld lives entirely here — it `LockSkip`s
  each transit turn and decrements `TurnsRemaining` until the move completes.
- **`UpdateProgress` — after the action, event-driven.** Called once the faction's
  action has produced its mutations, it inspects that list for goal-relevant
  events and emits supplemental mutations. Military Conquest counts
  `AssetRemoved` mutations caused by this faction's attacks against Force-category
  rivals and emits `GoalProgressed` / `GoalCompleted`.

`UpdateProgress` must run **before** `MutationEngine.Apply`. Progress counting
looks up the about-to-be-destroyed rival assets in state to read their category;
if the destructive mutations had already applied, the assets would be gone and
the kills uncountable. The orchestrator therefore folds `UpdateProgress`'s output
into the action's mutations and applies the combined batch in one pass.

### Worked example — Planetary Seizure

`PlanetarySeizure` (G-004) exercises every mechanism across a multi-turn
`ProcessPhase`:

- **Phase 0 → 1:** the [`Seize Planet`](actions.md) action emits `GoalInitiated`
  to start the goal on a contested world.
- **Phase 1 (combat):** `CheckLock` returns `LockRestrictActions` allowing only
  `Attack`. `UpdateProgress` watches for the moment no unstealthed rival asset
  remains on the target world and emits `GoalPhaseAdvanced` to phase 2.
- **Phase 2 (occupation):** `CheckLock` ticks `TurnsRemaining` down each turn,
  abandons the goal if the faction loses its foothold, and on the final tick
  emits the completion bundle — `GoalCompleted`, `XPAwarded`, and a
  `TagAdded` granting the `T-011` Planetary Government [tag](effect-mutation.md).

## Goals catalogue

The twelve registered goal handlers. "Advances by" names which beat does the
work — most read action mutations in `UpdateProgress`; Change Homeworld counts
down in `CheckLock`; Planetary Seizure does both. Only G-004 and G-012 impose a
turn lock; the rest run `LockNone`. Completion awards XP per the rules in
[`swn-faction-mechanics.md`](../../rules/swn-faction-mechanics.md).

| Goal | Advances by | Completes when |
|------|-------------|----------------|
| Military Conquest (G-001) | Destroying rival **Force** assets in combat | Kills reach the faction's Force rating |
| Commercial Expansion (G-002) | Destroying rival **Wealth** assets | Kills reach the faction's Wealth rating |
| Intelligence Coup (G-003) | Destroying rival **Cunning** assets | Kills reach the faction's Cunning rating |
| Planetary Seizure (G-004) | Multi-phase: clear, then occupy the target world | Occupation countdown elapses (grants the T-011 tag) |
| Expand Influence (G-005) | Placing a Base on a world it held none on | The base is placed — single turn |
| Blood the Enemy (G-006) | Dealing damage to rival assets and bases | Cumulative damage reaches Force + Cunning + Wealth |
| Peaceable Kingdom (G-007) | Passing turns without attacking (resets to 0 if it attacks) | Four uninterrupted peaceful turns |
| Destroy the Foe (G-008) | Reducing the target faction's HP | The target faction reaches 0 HP |
| Inside Enemy Territory (G-009) | Stealthing assets on rival-government worlds | Stealthed assets reach the faction's Cunning rating |
| Invincible Valor (G-010) | Destroying a Force asset rated above the faction's own Force | Any such kill — single turn |
| Wealth of Worlds (G-011) | Spending Coin to raise rival Base influence | Cumulative spend reaches 4 × Wealth |
| Change Homeworld (G-012) | Transit countdown that skips the faction's turns | Arrival at the destination world |

The three stat-kill goals (G-001/002/003) share `countAssetKillsByCategory` and
differ only in the category counted and the threshold stat — the keyed-handler
pattern paying off as near-identical small types.

## Key Decisions

- **Two advancement paths, by beat.** Time- and condition-based progress runs in
  `CheckLock` in the goal-lock phase (after bookkeeping, before the action);
  event-based progress runs in `UpdateProgress` after the action. A goal
  implements whichever path(s) it needs
  — Change Homeworld is pure `CheckLock`, Military Conquest pure `UpdateProgress`,
  Planetary Seizure both. (Frozen log 110–123.)
- **`UpdateProgress` runs before mutations apply.** The ordering is a real data
  dependency, not convention: progress is counted by inspecting
  about-to-be-destroyed entities in live state. Apply-after-inspect keeps the
  kill counts correct. (Frozen log 110–123.)
- **`locks` is a separate leaf package.** `LockType`/`GoalLock` live in their own
  package because both `goal` (the engine) and `goals` (the handlers) reference
  them, and `goal` already imports `goals`. A shared leaf breaks the import cycle
  that putting the lock types in either package would create. (Frozen log 215.)
- **Data-only goals are intentionally no-ops.** A goal ID in `goals.toml` with no
  registered handler returns `LockNone` and no progress — flavor data a GM can add
  without writing Go, the same silent-skip contract as tags and effects.
- **Goals emit, never mutate.** Every handler returns `[]domain.Mutation`;
  completion, XP, ticks, phase advances, even the Planetary Government tag are all
  expressed as mutations the orchestrator applies. The goal engine never touches
  faction state directly.

## Dependencies

**Depends on** `domain` (the goal mutation vocabulary — `GoalProgressed`,
`GoalCompleted`, `GoalPhaseAdvanced`, `GoalTurnsTick`, `GoalAbandoned`,
`XPAwarded`, `HomeworldChanged` — and `Faction.ActiveGoal`), the leaf `locks`
package, [world & movement](world-movement.md) for the spatial `Index` that
location-aware goals query, and [persistence & static data](persistence.md) for
`FactionState` and the rulebook (goal definitions, the `T-011` tag).

**Depended on by** the [orchestrator](orchestrator.md), which calls `CheckLock`
in the goal-lock phase and `UpdateProgress` in the action phase. Goals are
*started* by the goal-initiating [actions](actions.md) (`Seize Planet`, `Change
Homeworld`) and dropped by `Abandon Goal`; the mutations goals emit are applied
through the [effect & mutation](effect-mutation.md) layer.

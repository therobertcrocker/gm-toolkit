# Asset Movement Redesign — Discovery

Replaces the current `UseAssetAbility`-coupled movement model with a dedicated movement phase, per-asset orders that tick over multiple turns, and an asset-level `Speed` field. Subsumes paused work in the Spatial initiative (Phase 3 Commit 3 — `MaxHex` enforcement; Phase 4 — `ChangeHomeworld` distance awareness).

---

## Problem

Movement today is coupled to the action slot — a faction's one action per turn is consumed by `UseAssetAbility` when any movement is taken. There is no concept of multi-turn travel for arbitrary assets; the only multi-turn pattern is the `ChangeHomeworld` goal, which is faction-level and homeworld-only. Per-asset speed lives on an `AbilityStep.MaxHex` field — conflating "this asset's pace" with "this ability's range" — and ability-step movement bypasses the spatial layer's `Distance` entirely. In play, this means: assets cannot move and act in the same turn, journeys longer than one ability step require sequential full-turn hops, and there is no place to model FTL-style multi-region travel for individual assets.

---

## Design Summary

- A new **Movement Phase** runs for every faction every turn, between Bookkeeping (engine Phase 2 / 2B) and Action Selection (engine Phase 3).
- During its movement phase a faction may issue, revise, or cancel orders for **any number of movable assets**.
- `AssetDefinition` gains a `Speed` field (hex/turn). `DriftRating` already exists on `AssetDefinition` and feeds `spatial.Distance` via the existing `drift_costs` table.
- Each movable asset can hold at most **one in-flight `MovementOrder`** at a time. The order stores a pre-computed hex path; each turn the asset advances along the path by `Speed` hexes.
- `Asset.Location` is promoted from `string` to a struct (`WorldID`, `HexCoords`). `WorldID` is populated only when the asset is on a world; mid-flight, `WorldID` is empty and `HexCoords` identifies the current empty-space hex.
- Self-movement via `UseAssetAbility` is **removed**. Transport-pattern abilities (moving non-moving cargo) are **retained** and now fire during the movement phase, consuming the transport's own movement that turn.
- `ChangeHomeworld` remains a goal; its timer is now distance-aware, set at goal start to `spatial.Distance(homeworld, target, drift=3)`.

---

## Domain Model Changes

### `Location` struct

Replaces `Asset.Location string` and the analogous fields on `Base`, mutations, and digest types.

```go
type Location struct {
    WorldID   string    // empty when the entity is in empty space mid-transit
    HexCoords HexCoord  // always populated
}
```

Extensible by design — additional fields (e.g. orbital tags, faction-of-occupation pointers) can be added later without further migrations. For now `WorldID + HexCoords` is sufficient.

### `AssetDefinition.Speed`

New field on `AssetDefinition`. Type `int`, units of hex/turn.

- `Speed > 0` ⇒ the asset is **movable**; the engine permits a `MovementOrder` for it.
- `Speed == 0` (the default) ⇒ the asset is **non-moving**; it cannot issue a movement order and is eligible as transport cargo.

No separate `Movable` flag — `Speed > 0` is the predicate.

### `MovementOrder`

```go
type MovementOrder struct {
    AssetID      string
    Origin       Location   // captured at issue time
    Destination  Location   // target world
    Path         []HexCoord // pre-computed at issue (and on revision)
    StepIdx      int        // index into Path; asset's current HexCoords = Path[StepIdx]
    DriftRating  int        // captured from the asset definition at issue time
}
```

Stored on the `Asset` itself: `Asset.CurrentOrder *MovementOrder` (nil when stationary). One order per asset; revision rebuilds the path from the asset's current hex.

### `TurnPhase` enum

Insert `PhaseMovement` between `PhaseBookkeeping` and `PhaseAction`:

```go
const (
    PhaseBookkeeping TurnPhase = iota
    PhaseMovement
    PhaseAction
    PhaseComplete
)
```

### `Faction.Homeworld` (open)

Currently `string` (world ID). Two options: leave it as a world ID (homeworld is always on-world by definition) or migrate it to the same `Location` struct for type consistency. Recommended: **leave as `string`** — homeworld is conceptually a world reference, not a position; converting would introduce a meaningless `HexCoords` value. Open for review.

---

## Movement Phase Flow

### Slotting in the orchestrator

`RunFactionTurn` in `internal/faction/engine/orchestrator.go` currently runs: Phase 1 (Goal Lock) → Phase 2 (Bookkeeping) → Phase 2B (Stat Raise) → Phase 3 (Action) → Phase 4 (MutationReactor) → Phase 5 (Persist).

The new Movement Phase slots after Phase 2B and before Phase 3. A new `CheckpointMovement` checkpoint constant fires after the phase resolves, mirroring the existing bookkeeping/action checkpoints.

### Per-faction iteration

For each faction, in turn order:

1. Engine resolves any in-flight `MovementOrder` on the faction's assets (one tick per order, `StepIdx += Speed`, clamped to the path length).
2. Completed orders (`StepIdx >= len(Path) - 1`) emit `AssetMoved` and clear `CurrentOrder`.
3. The collector is offered the opportunity to **issue new orders, revise existing in-flight orders, or cancel** them, for any movable assets owned by the faction.
4. New orders compute their `Path` via `spatial.Distance` (or its sibling pathfinding primitive) using the asset's `DriftRating`.

Goal-lock `LockSkip` factions (e.g. mid-`ChangeHomeworld`) still tick their existing orders, but cannot issue/revise. Open: confirm with planning.

### Order lifecycle

- **Issued** — `MovementOrderIssued` mutation; `CurrentOrder` set; asset's `Location.WorldID` cleared on the first tick if not already empty (or kept until first tick — see Open Questions).
- **Progressed** — `MovementOrderProgressed` mutation each tick that advances `StepIdx` short of completion.
- **Revised** — `MovementOrderRevised` mutation; path recomputed from current hex; cost TBD (see Open Questions).
- **Cancelled** — `MovementOrderCancelled` mutation; asset's `Location` remains at its current hex (`WorldID` stays empty if mid-flight). The asset is stranded in empty space until a new order is issued.
- **Completed** — `MovementOrderCompleted` and `AssetMoved` mutations; `Location.WorldID` set to destination world ID; `CurrentOrder` cleared.

---

## Drift Integration

No new mechanic. `spatial.Distance(fromID, toID, crossingCost int)` already exists and the `drift_costs` table at `internal/faction/data/drift_costs.toml` already maps drift rating → per-region-crossing cost. The redesign reuses both:

- At order issuance, the engine looks up `crossingCost = rulebook.DriftCosts[asset.DriftRating - 1]` and calls `spatial.Distance(origin.WorldID, destination.WorldID, crossingCost)` to compute total hex-equivalent distance and the path.
- Speed is per-turn pace; drift is trip-length modifier. They are non-overlapping.

---

## Transport Ability

### Mechanics

- Activating a transport ability **consumes the transport asset's movement for that turn** — it cannot also issue a self-`MovementOrder`.
- A transport may carry **only same-faction, non-moving assets** (`Speed == 0`).
- Cargo is selected during order issuance for the transport.

### Cargo co-location

While in transit, each cargo asset's `Location` mirrors the transport's hex: `cargo.Location = {WorldID: "", HexCoords: transport.Path[StepIdx]}`. Cargo is **individually targetable** — an attacker who reaches the transport's hex can target the transport or any of its cargo.

On completion, the transport and all cargo gain `Location.WorldID = destination`.

### Schema question (open for planning)

The existing `AbilityStepMovement` step type is used for **both** self-movement (to be removed) and transport. Two options for planning:

1. Introduce `AbilityStepTransport` as a distinct step type; the movement-step type is deleted entirely.
2. Keep `AbilityStepMovement` and add a `transport bool` (or `cargo_only`) flag on the step; rulebook entries without the flag are removed.

Recommended: **option (1)**. It eliminates an ambiguous step type and makes the new model self-documenting at the rulebook level.

---

## Ability-Step Movement Removal

The current rulebook has `type = "movement"` ability steps across multiple asset definitions. They split into two groups:

### Removed (self-movement; redundant with `Speed`)

These assets simply gain a `Speed` value matching (or close to) their former `MaxHex`. Their ability is either removed or repurposed if the ability had other steps. Identified in `internal/faction/data/`:

- `Strike Fleet` (F4-003) — `max_hex=1`
- `Integral Protocols` (F7-002) — `max_hex=1`
- `Capital Fleet` (F8-xxx)
- Others identified at planning time via a scan for non-transport `type = "movement"` steps

### Retained (transport pattern)

These survive the redesign as `AbilityStepTransport` (per the recommended schema option above):

- `Cargo Lighters` (F2-xxx) — `max_hex=1`, transports any one non-Starship asset within one hex
- `Extended Theater` (F4-002) — `max_hex=2`, transports non-Starship assets within two hexes
- `Logistics Facility` (F6-xxx) — `max_hex=3`, transports any non-Starship within three hexes, ignores government opposition
- `Smugglers` (cunning) — `max_hex=2`, transports self or one Special Forces unit
- `Blockade Runner` (wealth) — `max_hex=3`, transports self or one Military Unit / Special Forces

**Mixed-mode reconciliation:** `Smugglers` and `Blockade Runner` currently describe both self-movement and transport. Under the redesign, "self" is handled by their `Speed`; the ability retains only the transport mode. Planning to confirm cost and hex-range carry-over.

### Migration

Every asset TOML needs a one-pass update:

1. Add `speed` (int) — non-zero for assets that are inherently movable per their description (Starships, Special Forces, etc.).
2. Remove `[ability.steps]` entries with `type = "movement"` for self-movers.
3. Reclassify remaining `type = "movement"` entries as the new transport step type.

---

## Change Homeworld Reconciliation

`ChangeHomeworld` remains a goal — not a movement order. The goal's existing `ActiveGoal.TurnsRemaining` field is repurposed:

- At goal start: `TurnsRemaining = spatial.Distance(faction.Homeworld, ActiveGoal.TargetWorld, drift_costs[3-1])` (drift rating 3 — the faction-level default).
- Each turn: existing `GoalTurnsTick` + `LockSkip` machinery in `engine/goal/lock.go:49` unchanged.
- Completion: existing `HomeworldChanged` mutation unchanged.

Retires Spatial Phase 4 ("Change Homeworld distance from spatial path") — the substance is captured here.

---

## Mid-Move Attack and Hex-Level Targeting

The "vulnerable mid-move" rule (factions may elect to attack a moving asset) requires the engine to resolve attack targets at the hex level when a defender is in transit.

### Required change

`world.Index` today is world-keyed: `assetsAt(worldID) → []*Asset`. The redesign adds a hex-keyed lookup: `assetsAtHex(HexCoord) → []*Asset` returning every asset (including cargo) whose `Location.HexCoords == coord`, regardless of `WorldID`.

`attack.go:79`'s `liveDefenders(faction.ID, attacker.Location, ...)` already takes a location handle; the change is to accept a hex coord when targeting in-transit assets and a `WorldID` when targeting on-world assets. The choice surfaces at the collector level (defender selection UI is out of scope for discovery).

### No auto-engagement

Moving through a hostile hex does **not** auto-trigger combat. Combat is only initiated by an explicit `Attack` action during the attacker's own action phase, targeting an asset whose current hex (on-world or in-transit) is reachable per the attacker's range rules. Combat rules themselves are unchanged.

---

## Mutation Surface

New mutations introduced by this redesign:

| Mutation | Emitted when |
|---|---|
| `MovementOrderIssued` | A new order is created during a movement phase |
| `MovementOrderProgressed` | A tick advances `StepIdx` short of completion |
| `MovementOrderRevised` | Faction revises destination mid-flight |
| `MovementOrderCancelled` | Faction cancels an in-flight order |
| `MovementOrderCompleted` | `StepIdx` reaches path end |
| `AssetMoved` | Asset's `Location.WorldID` changes (fires alongside `MovementOrderCompleted`) |

Existing mutations affected:

- Mutation `from_location` / `to_location` string fields migrate to `Location` (or remain string `WorldID`s if the planning step confirms no hex info is needed in narration).

---

## State Storage

- `MovementOrder` stored on the asset (`Asset.CurrentOrder *MovementOrder`).
- Persisted as part of normal `FactionState` save/load via TOML; no new top-level state file.
- Path is pre-computed and stored; recomputed only on order revision or (planning concern) world-graph mutation mid-flight.

---

## Replaces / Retires

This redesign supersedes the following items from `docs/implementation/spatial-model-effort-2-plan.md`:

- **Phase 3 Commit 3** — `MaxHex` enforcement on `UseAssetAbility` movement. Self-movement no longer flows through `UseAssetAbility`; transport-range enforcement is handled by the new transport step's `max_hex`.
- **Phase 4** — `ChangeHomeworld` distance from spatial path. Captured under "Change Homeworld Reconciliation" above.

A new implementation plan covering the full redesign will be written in the subsequent plan session.

---

## Open Questions

Carried forward for the plan session (or to ratify post-discovery):

1. **Cost of mid-flight revision.** Some cost is expected (Coin? partial path penalty?); value TBD.
2. **In-flight on world-graph change.** Robert's call: order remains intact unless the user revises. Planning needs to define behavior when a destination world is destroyed or merged out of existence (likely auto-cancel with strand-at-current-hex).
3. **First-tick `WorldID` clear.** Does an asset's `WorldID` clear (a) the moment an order is issued, or (b) on its first tick? Affects mid-movement-phase targeting consistency.
4. **`Faction.Homeworld` shape.** Stay as `string` (recommended) or migrate to `Location`.
5. **Goal-locked factions issuing orders.** Can a faction under `LockSkip` (e.g. mid-`ChangeHomeworld`) still issue/revise orders for its other assets? The rules doc says "faction may take no actions during the move" — silence on movement.
6. **Transport step schema.** Confirm option (1) — distinct `AbilityStepTransport` step type — vs. option (2) — flagged `AbilityStepMovement`.
7. **Mixed-mode ability reconciliation cost.** `Smugglers` and `Blockade Runner` lose their self-movement option; should the cost listed in their transport ability change?
8. **Multi-faction targeting in same movement phase.** Order of operations when faction A's movement-phase tick lands an asset at the same hex as faction B's already-in-transit asset — does B get to react before A acts? Planning concern.

---

## Out of Scope

- UI/UX of the movement phase (TUI wizard, observer events) — follows current patterns per Robert's note.
- AI faction movement decision-making — separate initiative.
- Faction-merge interactions with in-flight orders — separate initiative.
- Spatial map creation tooling (Deferred Major #9 in `planned-work.md`).
- Pirates tag mechanic — unlocked by this redesign but implemented separately under the programmatic tag initiative.

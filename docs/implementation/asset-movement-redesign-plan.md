# Asset Movement Redesign — Plan Overview

> Discovery: [`docs/discovery/asset-movement-redesign-discovery.md`](../discovery/asset-movement-redesign-discovery.md)
> Effort 1 (Foundation): [`asset-movement-redesign-effort-1-plan.md`](asset-movement-redesign-effort-1-plan.md)
> Effort 2 (Movement Engine): [`asset-movement-redesign-effort-2-plan.md`](asset-movement-redesign-effort-2-plan.md)
> Effort 3 (Cutover & Integration): [`asset-movement-redesign-effort-3-plan.md`](asset-movement-redesign-effort-3-plan.md)

Replaces the `UseAssetAbility`-coupled movement model with a dedicated Movement Phase that runs every turn between Bookkeeping (2B) and Action Selection (3). Per-asset `Speed` plus pre-computed hex paths let assets travel multiple turns without consuming the faction's action slot. Subsumes the paused Spatial Phase 3 Commit 3 (`MaxHex` enforcement) and Phase 4 (`ChangeHomeworld` distance) work.

<br/>

## Decisions Ratified in Planning

The discovery flagged eight open questions plus one plan-level question. All are settled:

| # | Decision |
|---|---|
| 1 | Mid-flight revision costs **1 Coin** (flat). Charged at the moment of revision via `CoinDelta`. |
| 2 | If a destination world disappears mid-flight, the order is **not auto-cancelled** — the asset arrives at the (now-empty) hex with `Location.WorldID = ""`. The hex is the durable address; the world is metadata. |
| 3 | `Asset.Location.WorldID` clears **on issue**, not on first tick. Issuance and progression remain separate code paths with no special-case "first tick" logic. |
| 4 | `Faction.Homeworld` migrates from `string` to `Location` for type consistency with the rest of the domain. |
| 5 | `LockSkip` factions (mid-`ChangeHomeworld`) **may not** select an action but **may** issue, revise, cancel, or tick movement orders. |
| 6 | A distinct `AbilityStepTransport` step type is introduced. `AbilityStepMovement` is deleted entirely once the TOML cutover is complete. |
| 7 | `Smugglers` (cunning) and `Blockade Runner` (wealth) become transport-only abilities; their existing Coin cost is retained. Self-movement is handled by their `Speed`. |
| 8 | Co-location is **state, not an interrupt** — no reaction window inside the Movement Phase. Combat is initiated only via an explicit `Attack` action in some attacker's later turn. |
| 9 | The `spatial.Path` primitive (path + total distance) lands in Phase 1 as part of the additive domain types phase. |

<br/>

## Open Questions — To Ratify at Implementation Time

| # | Question | Effort / Phase |
|---|---|---|
| 1 | **Where does the Movement Phase code live?** Effort 2 Phase 3 currently sketches a new `internal/faction/engine/movement/` package of free functions — which does not match any of the three sub-engine shapes (see `architecture-overview.md` → *Sub-Engine Shapes*) and would establish a fourth pattern under `engine/`. Alternative: host the Movement Phase under the existing `turn/` package as `(*TurnEngine).RunMovementPhase(...)` plus `movement_*.go` files. Turn is already Shape 3; `turn.New`'s arg list extends to take `spatialMap`. Movement is conceptually a turn phase (same family as Bookkeeping), and `turn`-hosting keeps the package boundary aligned with the conceptual owner. | Ratify at the start of Effort 2 Phase 3. See [`docs/discovery/sub-engine-alignment-discovery.md`](../discovery/sub-engine-alignment-discovery.md) for the shape catalog context. |

<br/>

## Effort Summary

Three efforts. Seven phases. Each phase is its own Sonnet execution session.

| Effort | Phases | Theme | What lands |
|---|---|---|---|
| [1 — Foundation](asset-movement-redesign-effort-1-plan.md) | 1, 2 | Domain prep | Additive types + `spatial.Path` + `Location`-struct migration of `Asset.Location`, `Base.Location`, `Faction.Homeworld`, `AssetMoved`, `HomeworldChanged`. Zero behavioral change. |
| [2 — Movement Engine](asset-movement-redesign-effort-2-plan.md) | 3, 4 | The mechanic | Movement Phase wired into the orchestrator with tick/issue/revise/cancel lifecycle, `LockSkip` semantics, and the transport ability step type + handler. Dormant in practice (no asset has `Speed > 0` and no TOML uses transport yet). |
| [3 — Cutover & Integration](asset-movement-redesign-effort-3-plan.md) | 5, 6, 7 | Player-facing change | TOML cutover (Speed values, transport reclassification, old step purged); `ChangeHomeworld` distance reconciliation; hex-level attack targeting. |

<br/>

## Cross-Cutting Context

A few concerns apply across efforts and are called out here so each effort file doesn't have to repeat the rationale.

### Location struct shape

```go
type Location struct {
    WorldID   string    // empty when the entity is in empty space mid-transit
    HexCoords HexCoord  // always populated post-migration
    Region    string    // captured from spatial.Location lookup at write time
}
```

`Region` is denormalized for now — it's looked up from the spatial map at the point a `Location` is assigned and stored alongside the hex coord. Avoids forcing readers to consult the spatial map for region context. (YAGNI: not adding any other fields until a concrete reader needs them.)

### `world.Index` keying

The existing `Index` is `map[string][]*domain.Asset` keyed by world ID string. Effort 1 keeps this shape — readers just feed `asset.Location.WorldID` instead of `asset.Location`. Mid-flight assets (empty `WorldID`) naturally drop out of the world-keyed index, which is semantically correct (they're not "on" any world).

Effort 3 Phase 7 adds a parallel `AssetsByHex map[HexCoord][]*domain.Asset` for hex-level targeting. Both indices stay in sync; world-keyed lookups are used by every existing scan site, hex-keyed lookups are used only by `Attack` targeting for in-transit defenders.

### `Astral Sea` sentinel retirement

`ability.go:225`'s `worldsFromState` appends the literal string `"Astral Sea"` to its returned world list — the legacy way to represent "in empty space." Under the redesign this is replaced by `Location.WorldID = ""`. The sentinel is removed in Effort 1 Phase 2; any remaining string match elsewhere is flagged for the same migration.

### Mutation versioning

No history-format migration is needed for past `AssetMoved` events. The mutation's struct shape changes (string → `Location`) but old `EventRecord` history files use TOML serialization with field tags — the next write of `AssetMoved` will simply emit the new shape. Historical reads of old `AssetMoved` entries are not required by any current consumer.

### Test harness

The existing `internal/faction/engine/testharness` helpers build factions with hardcoded string locations and homeworlds. Effort 1 Phase 2 migrates these to use `Location{WorldID: ..., HexCoords: ...}`. Per-effort test additions extend the harness with order-issuance helpers (Effort 2), transport setup helpers (Effort 2 Phase 4), and hex-coord lookup helpers (Effort 3 Phase 7).

<br/>

## Out of Scope

Per the discovery doc and Robert's notes:

- UI/UX of the movement phase (TUI wizard, observer events) — separate initiative
- AI faction movement decision-making — separate initiative
- Faction-merge interactions with in-flight orders — separate initiative
- Spatial map creation tooling (Deferred Major #9)
- Pirates tag mechanic (unlocked by this redesign, but implemented separately under the programmatic tag initiative)
- Existing campaign migration — no campaigns to migrate

<br/>

## Per-Effort Model Discipline

| Session | Model |
|---|---|
| Effort 1 Phase 1 (additive types, no design choices) | Sonnet |
| Effort 1 Phase 2 (mechanical refactor) | Sonnet |
| Effort 2 Phase 3 (orchestrator design — collector interface shape matters) | Sonnet, suggest Opus if collector interface design proves nontrivial during the session |
| Effort 2 Phase 4 (transport handler design — cargo lifecycle) | Sonnet, suggest Opus if cargo-lifecycle edge cases proliferate |
| Effort 3 Phase 5 (TOML edits + constant removal) | Sonnet |
| Effort 3 Phase 6 (small reconciliation) | Sonnet |
| Effort 3 Phase 7 (hex-index addition + targeting widening) | Sonnet |
| End-of-phase docs & pre-merge checklist (every phase) | Sonnet |

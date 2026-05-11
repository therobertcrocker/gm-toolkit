# Planned Work

Pre-discovery initiative tracker. Deferred items move to Planned Initiatives when scoped — at that point a write-up is added below and a discovery doc is written before implementation begins.

<br/>

## Initiatives

Full write-ups below. Each item has been scoped enough to warrant a dedicated discovery phase.

| # | Item | Size | Status | Trigger |
|---|------|------|--------|---------|
| 1 | [Spatial Model and World Graph](#spatial-model-and-world-graph) | Major | In-Progress | Hex distance, tech level, or Pirates tag becomes blocking |
| 2 | [Asset Movement Redesign](#asset-movement-redesign) | Major | Pending Discovery | Current movement model felt limiting in play; overlaps with Spatial Phase 3 Commit 3 and Phase 4 |

---
<br/>

## Deferred — Major

Items that will eventually warrant a full initiative entry. Each moves to Planned Initiatives when scoped.

| # | Item | Trigger |
|---|------|---------|
| 0 | CLI Rebuild | When tool is mechanically complete |
| 1 | TUI Rebuild | When tool is mechanically complete (and CLI Rebuild is done) |
| 2 | Programmatic tag handling | Scheduled tag milestone; 18 remaining tags + Effects Engine (`S`-flag assets); design in `tag-engine-discovery.md` |
| 3 | Campaign-scoped static data | No concrete trigger; revisit when distribution story is resolved |
| 4 | Command layer state mutation | Engine has sufficient substance to absorb `engine.CreateFaction` |
| 5 | Maintenance costs per asset | Cost data added to `AssetDefinition` in TOML |
| 6 | Starting Coin | Coin tracking designed and landed |
| 7 | Tag-granted assets | Engine action resolution complete |
| 8 | Tag reminder affordance | Alongside programmatic tag handling |
| 9 | Spatial map creation helper | When spatial model work begins; would be a CLI tool to generate the `worlds.toml` file from a user-provided template or data source |
| 10 | Goal Registration Refactor | When the tool is mechanically stable; would refactor to match action and tag registration patterns |

---
<br/>

## Deferred — Minor

Small, targeted fixes. No write-up needed — tracked here until scheduled.

| # | Item | Detail |
|---|------|--------|
| 1 | `BuyAsset` stealth ID hardcode (`"C3-002"`) | Add `stealth_applicator` flag or typed `TypeStealth` constant; silently breaks if the TOML ID changes |
| 2 | Goal/tag dispatch hardcodes display IDs | Key `Rulebook.Goals`/`Rulebook.Tags` on semantic table key; `id` becomes display-only; one-time migration |
| 3 | `SeizePlanet` invisible in history | `Output()` returns no mutations; add `GoalPhaseAdvanced` mutation |
| 4 | `narrateUseAssetAbility` fallback text | Moot until TUI rebuild; resurfaces when narration is re-implemented against the new observer interface |
| 5 | Structured logging layer | No `log` package calls anywhere; `BuildSpatialIndex` silently skips stale fragments with a TODO; add a lightweight, consistent logging approach across the tool |

---
<br />
<br />

# Planned Initiatives
This section contains detailed write-ups for each planned initiative, including problem statements, proposed approaches, tradeoffs, and triggers. This will be the source of truth when it comes time to start discovery work on any of these items.

<br />

## Spatial Model and World Graph

### Problem

Worlds are bare strings (`asset.Location = "Tartarus"`). There is no registry of what worlds exist, where they are, or what properties they have. All spatial queries are derived at runtime by scanning faction state — finding assets on a world requires iterating every faction then every asset within each faction, filtering by `.Location`. This O(factions × assets) pattern appears in `attack.go` (`eligibleDefenders`, `liveDefenders`), `expand_influence.go` (`rivalsOnWorld`), `goal/progress.go` (`worldHasRivalPresence`, `rivalHasPlanetaryGovernmentOnWorld`), and `seize_planet.go`. At current scale it is immeasurable; as faction and asset counts grow it compounds.

The deeper cost is what the missing model prevents from being enforced:

- **Hex distance** — `Change Homeworld` sets `TurnsRemaining` to a fixed value that ignores distance; `UseAssetAbility` movement defers hex-distance enforcement to GM adjudication; the `MaxHex` field is parsed and stored but never checked
- **Tech level** — `BuyAsset.purchasableDefinitions` does not filter by `AssetDefinition.TechLevel`; there is nowhere to look up a world's tech level
- **Government permission (flag `P`)** — `BuyAsset.Validate` does not check the `P` flag; enforcing it requires knowing which faction holds the `Planetary Government` tag on the target world
- **Pirates tag** — the movement cost mechanic requires knowing which world an asset moved from and to; without a world graph this is unimplementable

### Approach

Introduce a `World` domain type with at minimum `{ ID string, TechLevel int, Population int }` and a hex-coordinate or adjacency structure for distance queries. Layer a derived spatial index on top — a `map[worldID][]*Asset` and `map[worldID][]*Base` — rebuilt at the start of each turn and treated as read-only during resolution. This index replaces the O(n) world scans throughout the engine.

Campaign state gains a `Worlds map[string]*World` field in `FactionState`. The `faction create` wizard gains a world-setup step. The `FACTION_DATA_DIR` data files gain a per-campaign `worlds.toml`.

### Unlocks

- Hex-distance enforcement for `Change Homeworld` goal timer and asset movement
- Tech level validation in `BuyAsset` and `ExpandInfluence`
- Government permission (`P` flag) enforcement in `BuyAsset`
- Pirates tag mechanic (movement cost)
- Full world-list for `SelectMoveDestination` (currently derived from live asset locations)

### Trigger

Pick up when the first of these becomes blocking: tech level enforcement is needed for a campaign, movement distance needs to be enforced (AI decision-making), or the Pirates tag is scheduled for implementation.

<br/>

### Status

- **Phase 1 — `internal/spatial` package:** complete
- **Phase 2 — engine wiring with `SpatialIndex`:** complete (`feat/spatial-effort-2`)
- **Phase 3 Commits 1 & 2 — TL and P-flag enforcement:** complete (`feat/spatial-effort-2`)
- **Paused pending Asset Movement Redesign:** Phase 3 Commit 3 (MaxHex enforcement on `UseAssetAbility` movement), the MaxHex portion of Phase 3 Commit 4 tests, and Phase 4 (Change Homeworld distance from spatial path). These will be re-planned as part of the movement redesign rather than executed against the current `spatial-model-effort-2-plan.md` — a new implementation plan covering them alongside the redesign will be written when the movement initiative reaches its plan session

<br/>
<br/>

## Asset Movement Redesign

### Problem

Movement is coupled to the faction's action slot — moving an asset consumes the entire turn's action via `UseAssetAbility`. Asset movement is also single-turn-only: an asset moves up to `AbilityStep.MaxHex` hexes per use, with no notion of multi-turn travel. The one multi-turn pattern, `Change Homeworld`, is modeled as a goal and applies only to a faction's homeworld — there is no general way for individual assets to set up longer journeys. `MaxHex` lives on the ability step rather than the asset definition, conflating "this asset's speed" with "this ability's range."

In play this means moving and acting in the same turn is impossible for any asset, and traveling further than `MaxHex` requires a full-turn action per hop in sequence.

### Approach

Decouple movement from the action phase. Promote per-turn speed to an asset-level field on `AssetDefinition`. Generalize the multi-turn pattern: any asset whose chosen destination exceeds per-turn speed sets up a `MovementOrder` (destination + remaining hexes) that ticks down each turn. The controller (GM, AI, test) can issue, revise, or cancel orders each turn. New mutations cover the order lifecycle; `AssetMoved` still fires on completion. The spatial layer's `Distance` primitive supplies hex counts.

Existing ability-step movement (e.g. `Strike Fleet`, `Capital Fleet`, `Integral Protocols`) needs to be reconciled with the new model — either as bonus hops on top of base speed or folded entirely into the new system. To be decided in discovery.

### Unlocks

- Decouples movement from the action economy — assets can move and act in the same turn
- Multi-turn moves for any asset, not just the homeworld
- Single coherent movement model (collapses ability-step movement and goal-based movement)
- Subsumes Spatial Phase 3 Commit 3 (MaxHex enforcement) and Phase 4 (Change Homeworld distance) into one coherent design

### Trigger

Active — current movement model felt limiting in play; design overlap with Spatial Phase 3 Commit 3 and Phase 4 makes this the right time to pause spatial work and redesign. Discovery is the next session.


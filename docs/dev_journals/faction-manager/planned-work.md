# Planned Work

Pre-discovery initiative tracker. Deferred items move to Planned Initiatives when scoped — at that point a write-up is added below and a discovery doc is written before implementation begins.

<br/>

## Planned Initiatives

Full write-ups below. Each item has been scoped enough to warrant a dedicated discovery phase.

| # | Item | Type | Status | Detail |
|---|------|------|--------|--------|
| 1 | [Spatial Model and World Graph](#spatial-model-and-world-graph) | feature | Blocked | — |
| 2 | [Asset Movement Redesign](#asset-movement-redesign) | feature | In-Progress | — |

---
<br/>

## Up Next (Ready or Near Ready)
The queue of deferred items that are ready to become initiatives. These are scoped and waiting for their trigger conditions to be met.

| # | Item | Type | Trigger | Detail |
|---|------|------|---------|--------|
| 0 | CLI Rebuild | feature | When tool is mechanically complete | — |
| 9 | Spatial map CLI | feature | When spatial model work begins | CLI tool to generate the `worlds.toml` file from a user-provided template or data source |

---
<br />

## Backlog

Unscoped items waiting for their trigger. Move to Up Next when the trigger is close; move to Planned Initiatives when fully scoped for discovery.

| # | Item | Type | Trigger | Detail |
|---|------|------|---------|--------|
| 1 | TUI Rebuild | feature | When tool is mechanically complete (and CLI Rebuild is done) | — |
| 2 | Programmatic tag handling | feature | Scheduled tag milestone; 18 remaining tags + Effects Engine (`S`-flag assets); design in `tag-engine-discovery.md` | — |
| 3 | Campaign-scoped static data | feature | No concrete trigger; revisit when distribution story is resolved | — |
| 4 | Command layer state mutation | refactor | Engine has sufficient substance to absorb `engine.CreateFaction` | — |
| 5 | Maintenance costs per asset | feature | Cost data added to `AssetDefinition` in TOML | — |
| 6 | Starting Coin | feature | Coin tracking designed and landed | — |
| 7 | Tag-granted assets | feature | Engine action resolution complete | — |
| 8 | Tag reminder affordance | feature | Alongside programmatic tag handling | — |
| 11 | Declarative Action Preconditions | refactor | When AI/planner integration becomes scoped | lift `Action.Validate(...) bool` into `Action.Preconditions(...) []Precondition` — unlocks backward search, self-explanation, and structured error messages for free |
| 12 | Unified Goal State Predicates | feature | When AI/planner integration becomes scoped | add a uniform `Satisfied(state) bool` alongside the per-goal `progressX` functions. Process-style progression stays for narrative; predicate form lets agents set arbitrary world states as goals. Both shapes coexist |
| 13 | `BuyAsset` stealth ID hardcode (`"C3-002"`) | bugfix | — | Add `stealth_applicator` flag or typed `TypeStealth` constant; silently breaks if the TOML ID changes |
| 14 | Goal/tag dispatch hardcodes display IDs | bugfix | — | Key `Rulebook.Goals`/`Rulebook.Tags` on semantic table key; `id` becomes display-only; one-time migration |
| 15 | `SeizePlanet` invisible in history | bugfix | — | `Output()` returns no mutations; add `GoalPhaseAdvanced` mutation |
| 16 | `narrateUseAssetAbility` fallback text | bugfix | — | Moot until TUI rebuild; resurfaces when narration is re-implemented against the new observer interface |
| 17 | Structured logging layer | feature | — | No `log` package calls anywhere; `BuildSpatialIndex` silently skips stale fragments with a TODO; add a lightweight, consistent logging approach across the tool |
| 18 | Goal XP not derived from `difficulty` | bugfix | — | Each handler in `goal/goals/` hardcodes its XP value (constants like `2` for "moderate", `1` for "low", or inline formulas) and `completeGoal(faction, xp)` accepts the result. The `difficulty` field in `goals.toml` is descriptive-only — a GM edit silently diverges from awarded XP. Two flavors to wire: constants (`"low"`/`"moderate"`/`"none"`) and formulas (`"half_assets_destroyed"`, `"half_avg_ruling_faction"`, `"one_plus_avg_target"`, `"low_plus_contested"`). Likely needs `Difficulty` retyped from `string` to either an int or a tagged sum so the engine can dispatch. |
| 19 | Ability Engine Redesign | refactor | After Asset Movement Redesign — movement step likely eliminated or restructured, making this the right time to revisit | The `steps/` sub-package is Shape 1 code (closed two-value enum switch) housed in a Shape 2 structure (sibling handler package). Symptom: `steps.Collector` is a structural duplicate of `ability.Collector`, required only to avoid the `ability` ↔ `steps` import cycle. The step functions likely belong directly in `ability/` with no sub-package. Broader question: once movement is decoupled from the ability engine (see Asset Movement Redesign), what does ability resolution actually look like? Discovery should decide the right shape before implementing. |

---
<br />
<br />

# Initiative Write-Ups
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
- **Blocked by: Asset Movement Redesign** — Phase 3 Commit 3 (MaxHex enforcement on `UseAssetAbility` movement), the MaxHex portion of Phase 3 Commit 4 tests, and Phase 4 (Change Homeworld distance from spatial path). These will be re-planned as part of the movement redesign rather than executed against the current `spatial-model-effort-2-plan.md` — a new implementation plan covering them alongside the redesign will be written when the movement initiative reaches its plan session

<br/>
<br/>

## Asset Movement Redesign

### Problem

Movement is coupled to the faction's action slot — moving an asset consumes the entire turn's action via `UseAssetAbility`. Asset movement is also single-turn-only: an asset moves up to `AbilityStep.MaxHex` hexes per use, with no notion of multi-turn travel. The one multi-turn pattern, `Change Homeworld`, is modeled as a goal and applies only to a faction's homeworld — there is no general way for individual assets to set up longer journeys. `MaxHex` lives on the ability step rather than the asset definition, conflating "this asset's speed" with "this ability's range."

In play this means moving and acting in the same turn is impossible for any asset, and traveling further than `MaxHex` requires a full-turn action per hop in sequence.

### Approach

Decouple movement from the action phase. Promote per-turn speed to an asset-level field on `AssetDefinition`. Generalize the multi-turn pattern: any asset whose chosen destination exceeds per-turn speed sets up a `MovementOrder` (destination + remaining hexes) that ticks down each turn. The controller (GM, AI, test) can issue, revise, or cancel orders each turn. New mutations cover the order lifecycle; `AssetMoved` still fires on completion. The spatial layer's `Distance` primitive supplies hex counts.

Each Faction gets a "movement phase" before the action phase, where they can choose to put in movement orders for any valid assets. The engine processes all movement orders in sequence, ticking them down and moving assets accordingly. This allows for multi-turn movement, and for movement to happen alongside actions in the same turn.

Existing ability-step movement (e.g. `Strike Fleet`, `Capital Fleet`, `Integral Protocols`) needs to be reconciled with the new model — either as bonus hops on top of base speed or folded entirely into the new system. To be decided in discovery.

### Unlocks

- Decouples movement from the action economy — assets can move and act in the same turn
- Multi-turn moves for any asset, not just the homeworld
- Single coherent movement model (collapses ability-step movement and goal-based movement)
- Subsumes Spatial Phase 3 Commit 3 (MaxHex enforcement) and Phase 4 (Change Homeworld distance) into one coherent design

### Trigger

Active — current movement model felt limiting in play; design overlap with Spatial Phase 3 Commit 3 and Phase 4 makes this the right time to pause spatial work and redesign. Discovery is the next session.

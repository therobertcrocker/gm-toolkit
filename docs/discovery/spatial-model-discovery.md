# Spatial Model and World Graph — Discovery

The faction engine currently treats world locations as bare strings. There is no registry of what worlds exist, no spatial structure, and no way to compute distances. All spatial queries are O(factions × assets) scans. This discovery doc defines the spatial model that replaces them: a hybrid hex-grid / relational graph structure shared between the faction tool and the planned Codex.

<br/>
<br/>

## Spatial Hierarchy

The campaign world is called the **Astral Sea**. It is structured as a collection of Real-Space regions connected through Drift-Space.

<br/>

### Region

A Region is a named polygonal hex grid — a cluster of worlds in Real-Space. Each Region contains zero or more Fragments at specific hex coordinates, and zero or more boundary hexes that connect to Drift-Space.

**Fields:** `id`, `name`

A Region's boundary connections are defined inline — a list of hex positions within this Region that connect to hex positions in other Regions via Drift-Space.

<br/>

### Fragment

A Fragment is a world — a location in the Astral Sea where factions can place assets and bases. Every Fragment exists within a Region at a specific hex coordinate.

**Fields:**

| Field | Type | Description |
|---|---|---|
| `id` | string | Stable key; used as the reference in all other data |
| `name` | string | Display name |
| `tech_level` | int (0–4) | Minimum tech level required to purchase certain assets here |
| `population` | int | Total population; gates starship-type asset purchases |
| `region` | string | ID of the containing Region |
| `hex_q` | int | Axial hex coordinate Q within the Region |
| `hex_r` | int | Axial hex coordinate R within the Region |

Tech level tiers are campaign-configurable via a TOML file. For the current campaign:

| Level | Name |
|---|---|
| 0 | Primitive |
| 1 | Arcane |
| 2 | Arcano-Tech |
| 3 | Astral-Aware |
| 4 | Aether-Tech |

A Fragment may or may not exist at any given hex position, including boundary hexes.

<br/>

### Boundary Connections

A Region's boundary hexes are grid positions at the edge of Real-Space that border Drift-Space. Each boundary hex connects to one or more boundary hexes in other Regions. Boundary connections are defined as part of the Region.

**Fields per connection:** `from_q`, `from_r` (hex position within this Region), `to_region` (Region ID), `to_q`, `to_r` (hex position in the destination Region).

A single boundary hex may have multiple outgoing connections — to different hexes, in the same or different Regions.

<br/>
<br/>

## Distance Algorithm

Distance is always expressed as an integer number of hex-equivalents. It is the input to faction mechanics that depend on spatial position (Change Homeworld turn duration, asset movement range).

<br/>

### Intra-Region (Hex Distance)

When Fragment A and Fragment B are in the same Region, distance is standard axial hex distance:

```
distance = max(|q_a - q_b|, |r_a - r_b|, |(q_a + r_a) - (q_b + r_b)|)
```

<br/>

### Cross-Region (Drift Crossing)

When Fragment A and Fragment B are in different Regions, travel must pass through at least one boundary connection. The path is:

```
hex(Fragment A → boundary hex in Region A)
  + drift_cost(drift_rating)
  + hex(boundary hex in Region B → Fragment B)
```

The total distance is the minimum over all viable paths. This is a shortest-path query over a hybrid graph: implicit hex edges within Regions, and explicit boundary-connection edges between Regions weighted by the drift cost.

**Drift-Space crossing distance is always 1 in the graph.** The drift cost function converts this to a hex-equivalent based on the traversing asset's drift rating (see below). Drift rating is an asset property — it is passed as a parameter to the distance query and does not live in the spatial model.

<br/>
<br/>

## Drift Rating Cost Table

Drift rating is a property of starship-type assets, rated 1–5. Higher rating represents the ability to navigate Drift-Space more directly, approaching the equivalent of a 1-hex crossing.

The cost table is campaign-configurable via a TOML file in the faction data directory:

| Drift Rating | Hex-Equivalent Crossing Cost |
|---|---|
| 1 | 5 |
| 2 | 4 |
| 3 | 3 |
| 4 | 2 |
| 5 | 1 |

The table is loaded at startup alongside other rulebook data. Drift rating and its cost formula are faction-tool-specific; they do not belong to the shared `internal/spatial` package.

<br/>
<br/>

## Spatial Index

The spatial index is a derived, in-memory data structure rebuilt at the start of each turn and treated as read-only during resolution. It replaces the O(factions × assets) scans that currently appear in the engine.

**Contents:**
- `fragmentsByID` — `map[string]*Fragment`; direct lookup by Fragment ID
- `assetsByFragment` — `map[string][]*Asset`; all live assets per Fragment ID, across all factions
- `basesByFragment` — `map[string][]*Base`; all live bases per Fragment ID, across all factions

The index is built by scanning `FactionState` once at turn start. Code that currently iterates all factions and assets to find "who is on this world" is replaced by a single map lookup.

Distance queries are not cached in the index — they are computed on demand using the Region and Fragment data loaded from the shared TOML files.

<br/>
<br/>

## `internal/spatial` Package

A new shared package at `internal/spatial` defines the canonical Go types for all spatial entities:

- `Region` — id, name, boundary connections
- `BoundaryConnection` — from hex, to region, to hex
- `Fragment` — all fields above
- `SpatialData` — the loaded collection of all Regions and Fragments for a campaign
- `Distance(from, to FragmentID, driftRating int, data SpatialData) (int, error)` — the distance query

Both `internal/faction` and the planned `internal/codex` import this package. The Go types define the schema; the TOML files are the integration point between tools.

<br/>
<br/>

## Codex Compatibility

The Codex is the source of truth for spatial data. Both the faction tool and the Codex read from the same TOML files. The faction tool is configured with a `SPATIAL_DATA_DIR` environment variable pointing to the directory containing `regions.toml` and `fragments.toml`. When the Codex is built, this variable is updated to point at the Codex's data directory — no migration or sync step is required.

The faction tool's `Fragment` type is a proper subset of the Codex's eventual `Fragment` type. Because Go's TOML unmarshaling ignores unknown fields, the faction tool can read Codex-owned fragment files that contain richer data (descriptions, lore, NPC references) without any changes to the faction tool's loader.

<br/>
<br/>

## Integration Points

The following engine locations currently do O(factions × assets) world scans and are replaced by spatial index lookups:

| File | Function | Current behavior | Replaced by |
|---|---|---|---|
| `engine/action/actions/attack.go` | `eligibleDefenders`, `liveDefenders` | Scans all factions for assets on target world | `assetsByFragment[targetID]` |
| `engine/action/actions/expand_influence.go` | `rivalsOnWorld` | Scans all factions for assets on target world | `assetsByFragment[targetID]` |
| `engine/goal/progress.go` | `worldHasRivalPresence`, `rivalHasPlanetaryGovernmentOnWorld` | Scans all factions | `assetsByFragment`, `basesByFragment` |
| `engine/action/actions/seize_planet.go` | defender scan | Scans all factions | `assetsByFragment[targetID]` |

Additionally, `asset.Location` changes from a bare string to a Fragment ID. All code that reads or writes `asset.Location` must reference a Fragment known to the spatial model.

The `Change Homeworld` goal currently accepts a manually-entered turn count. With the spatial model, turn count is computed automatically: `1 + Distance(current homeworld, destination, driftRating)`.

The `MaxHex` field on `AssetDefinition` is currently parsed but never enforced. With the spatial model, `UseAssetAbility` movement validates against `Distance(from, to, driftRating) ≤ MaxHex`.

<br/>
<br/>

## TOML Schema

**`$SPATIAL_DATA_DIR/regions.toml`** — owned by Codex (or faction tool before Codex exists):

```toml
[[region]]
id = "corona-reach"
name = "The Corona Reach"

  [[region.boundary]]
  from_q = 3
  from_r = 7
  to_region = "void-expanse"
  to_q = 0
  to_r = 2
```

**`$SPATIAL_DATA_DIR/fragments.toml`** — owned by Codex:

```toml
[[fragment]]
id = "tartarus"
name = "Tartarus"
tech_level = 2
population = 500000
region = "corona-reach"
hex_q = 3
hex_r = 4
```

**`$FACTION_DATA_DIR/drift_costs.toml`** — faction-tool-specific, campaign-configurable:

```toml
# drift_costs[i] = hex-equivalent cost for drift rating (i+1)
drift_costs = [5, 4, 3, 2, 1]
```

**`$FACTION_DATA_DIR/tech_levels.toml`** — faction-tool-specific, campaign-configurable:

```toml
[[tech_level]]
level = 0
name = "Primitive"

[[tech_level]]
level = 1
name = "Arcane"

[[tech_level]]
level = 2
name = "Arcano-Tech"

[[tech_level]]
level = 3
name = "Astral-Aware"

[[tech_level]]
level = 4
name = "Aether-Tech"
```

<br/>
<br/>

## Faction Implementation Phases

The spatial model itself (types, distance algorithm, index) is one deliverable. Full faction-tool functionality requires additional work that depends on the spatial model being in place. These are tracked as implementation phases within this initiative, not deferred.

- **Drift rating on assets** — `AssetDefinition` gains a `drift_rating` field; loaded from the campaign's asset TOML files; used as the parameter to cross-region distance queries
- **Tech level enforcement** — `BuyAsset.Validate` and `ExpandInfluence.Validate` filter purchasable assets and destinations by Fragment tech level; requires Fragment lookup via the spatial index
- **P-flag enforcement** — `BuyAsset.Validate` checks the `P` flag against which faction holds Planetary Government on the target Fragment; requires `basesByFragment` from the spatial index
- **`MaxHex` enforcement** — `UseAssetAbility` movement validates `Distance(from, to, driftRating) ≤ MaxHex`; previously enforced only by GM adjudication
- **`Change Homeworld` distance computation** — turn count computed automatically from `Distance(current, destination, driftRating)` rather than entered manually by the GM
- **`faction create` wizard** — homeworld selection step against the Fragment registry; replaces the current free-text world entry

**Tracked separately:** The Pirates tag mechanic (movement cost per hop) depends on drift rating and this spatial model, but belongs to the programmatic tag handling initiative already tracked in planned-work.

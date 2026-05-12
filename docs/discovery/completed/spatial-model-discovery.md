# Spatial Model and World Graph — Discovery

The faction engine currently treats world locations as bare strings. There is no registry of what worlds exist, no spatial structure, and no way to compute distances. All spatial queries are O(factions × assets) scans. This discovery doc defines the spatial model that replaces them: a typed, map-aware system built around a `SpatialMap` interface with three concrete implementations, one of which — `HybridMap` — models the Astral Sea campaign world.

<br/>
<br/>

## Spatial Model Types

`internal/spatial` defines a `SpatialMap` interface implemented by three concrete map types. Each type represents a different way of organizing and traversing a game world.

```
SpatialMap (interface)
├── HexMap    — a single hex grid; intra-grid BFS; no cross-region travel
├── GraphMap  — nodes with explicit weighted edges; shortest-path over graph
└── HybridMap — hex-grid Regions connected by a relational boundary graph
```

The `Location` interface is the abstract type for any place where factions can have presence. Each map type defines its own concrete location type.

```go
type SpatialMap interface {
    Location(id string) (Location, bool)
    Distance(fromID, toID string, crossingCost int) (int, error)
}

type Location interface {
    ID() string
    Name() string
    TechLevel() int
    Population() int
}
```

`HexMap` and `GraphMap` are stubs in the initial implementation. `HybridMap` is fully implemented for the Astral Sea campaign.

<br/>
<br/>

## HybridMap — The Astral Sea

The Astral Sea is modeled as a `HybridMap`: a set of named hex-grid Regions connected through Drift-Space by a relational boundary graph. Intra-region travel is computed via BFS over the Region's valid hex set. Cross-region travel is computed via explicit boundary connections.

<br/>

### Region

A Region is a named, irregular hex grid — a specific set of valid hex positions in Real-Space. Each Region contains zero or more Fragments at specific hex coordinates, and zero or more boundary connections to other Regions.

**Fields:** `id`, `name`, `hexes` (explicit set of valid hex positions), `boundaries`

The hex set defines the shape of the Region. Only hexes listed in the set are traversable — routing must pass through valid hexes and cannot cut through missing positions (holes) or outside the Region's boundary. A hex position may or may not have a Fragment on it; empty hexes are traversable but carry no game significance.

<br/>

### Fragment

A Fragment is a world — a location in the Astral Sea where factions can place assets and bases. Every Fragment exists within a Region at a specific hex coordinate, which must be a member of the Region's hex set.

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

`Fragment` implements the `Location` interface. Its `Region` and `Hex` fields are also directly accessible by the faction engine.

Tech level tiers are campaign-configurable via a TOML file. For the current campaign:

| Level | Name |
|---|---|
| 0 | Primitive |
| 1 | Arcane |
| 2 | Arcano-Tech |
| 3 | Astral-Aware |
| 4 | Aether-Tech |

<br/>

### Boundary Connections

A boundary connection links a specific hex position in one Region to a specific hex position in another Region, representing a Drift-Space crossing. Either end of the connection may be an empty hex (no Fragment present).

**Fields per connection:** `from_q`, `from_r` (hex position in this Region), `to_region` (Region ID), `to_q`, `to_r` (hex position in the destination Region).

Boundary connections are bidirectional. A single hex may have multiple outgoing connections — to different hexes in the same or different Regions.

<br/>
<br/>

## Distance Algorithm

Distance is always expressed as an integer number of hex-equivalents. It is the input to faction mechanics that depend on spatial position (Change Homeworld turn duration, asset movement range).

<br/>

### Algorithm

`Distance(fromID, toID string, crossingCost int) (int, error)`

`crossingCost` is the caller-resolved hex-equivalent cost for one boundary crossing. Drift rating is a faction-engine concept — `internal/spatial` has no knowledge of it. The faction engine looks up the cost from its drift cost table and passes it in.

**Dijkstra over a unified graph:**

- **Nodes:** every valid hex position in every Region, identified by `(regionID, q, r)`
- **Intra-region edges:** each hex to each of its six axial neighbors, if the neighbor is in the Region's valid hex set; cost 1
- **Cross-region edges:** each boundary connection in both directions; cost `crossingCost`
- **Source:** hex of the `fromID` Fragment
- **Target:** hex of the `toID` Fragment

Returns the total cost of the shortest path. Returns an error if either Fragment ID is unknown or no path exists.

<br/>

### Intra-Region Example

When Fragment A and Fragment B are in the same Region, routing is BFS through the Region's valid hex set. Straight-line axial distance is not used — the path must pass through valid hexes, routing around any holes or boundary edges in the Region's shape.

<br/>

### Cross-Region Example

When Fragment A and Fragment B are in different Regions, the path is:

```
BFS(Fragment A → boundary hex in Region A)
  + crossingCost
  + BFS(landing hex in Region B → Fragment B)
```

The total distance is the minimum over all viable paths. Multiple boundary connections may exist; Dijkstra considers all of them.

<br/>
<br/>

## Drift Rating Cost Table

Drift rating is a property of starship-type assets, rated 1–5. Higher rating represents more direct Drift-Space navigation, reducing the hex-equivalent crossing cost.

The cost table is campaign-configurable via `$FACTION_DATA_DIR/drift_costs.toml`:

| Drift Rating | Hex-Equivalent Crossing Cost |
|---|---|
| 1 | 5 |
| 2 | 4 |
| 3 | 3 |
| 4 | 2 |
| 5 | 1 |

Drift rating and its cost table are faction-tool-specific; they do not belong to `internal/spatial`. The faction engine resolves `crossingCost = driftCosts[asset.DriftRating - 1]` and passes the result to `Distance()`.

<br/>
<br/>

## Spatial Index

The spatial index is a derived, in-memory data structure rebuilt at the start of each turn and treated as read-only during resolution. It replaces the O(factions × assets) scans that currently appear in the engine.

The index lives in the faction engine layer — not in `internal/spatial` — because it holds faction domain types (`*domain.Asset`, `*domain.Base`) that would couple the shared spatial package to faction-specific code.

**Contents:**
- `AssetsByFragment` — `map[string][]*domain.Asset`; all live assets per Fragment ID, across all factions
- `BasesByFragment` — `map[string][]*domain.Base`; all live bases per Fragment ID, across all factions

The index is built by scanning `FactionState` once at turn start. Code that currently iterates all factions and assets to find "who is on this world" is replaced by a single map lookup.

Distance queries are not cached in the index — they are computed on demand via `SpatialMap.Distance()`.

<br/>
<br/>

## `internal/spatial` Package

A new shared package at `internal/spatial` defines the canonical Go types and interfaces:

- `SpatialMap` — interface: `Location()`, `Distance()`
- `Location` — interface: `ID()`, `Name()`, `TechLevel()`, `Population()`
- `HybridMap` — full implementation; owns `Region`, `Fragment`, `HexCoord`, `BoundaryConnection`
- `HexMap` — stub
- `GraphMap` — stub
- `LoadHybrid(dataDir string) (*HybridMap, error)` — reads `regions.toml` and `fragments.toml`

Both `internal/faction` and the planned `internal/codex` import this package. The TOML files are the integration point between tools.

<br/>
<br/>

## Codex Compatibility

The Codex is the eventual source of truth for spatial data. Both the faction tool and the Codex read from the same TOML files. The faction tool is configured with a `SPATIAL_DATA_DIR` environment variable pointing to the directory containing `regions.toml` and `fragments.toml`, separate from `FACTION_DATA_DIR`. When the Codex is built, this variable is updated to point at the Codex's data directory — no migration or sync step is required.

The faction tool's `Fragment` type is a proper subset of the Codex's eventual `Fragment` type. Because Go's TOML unmarshaling ignores unknown fields, the faction tool can read Codex-owned fragment files that contain richer data (descriptions, lore, NPC references) without any changes to the faction tool's loader.

<br/>
<br/>

## Integration Points

The following engine locations currently do O(factions × assets) world scans and are replaced by spatial index lookups:

| File | Function | Current behavior | Replaced by |
|---|---|---|---|
| `engine/action/actions/eligibility.go` | `eligibleDefenders`, `liveDefenders` | Scans all factions for assets on target world | `AssetsByFragment[targetID]` |
| `engine/action/actions/expand_influence.go` | `rivalsOnWorld` | Scans all factions for assets on target world | `AssetsByFragment[targetID]` |
| `engine/goal/progress.go` | `worldHasRivalPresence`, `rivalHasPlanetaryGovernmentOnWorld` | Scans all factions | `AssetsByFragment`, `BasesByFragment` |
| `engine/action/actions/seize_planet.go` | defender scan | Scans all factions | `AssetsByFragment[targetID]`, `BasesByFragment[targetID]` |

Additionally, `asset.Location` changes from a bare display string to a Fragment ID. All code that reads or writes `asset.Location` must reference a Fragment known to the spatial model.

The `Change Homeworld` goal currently accepts a manually-entered turn count. With the spatial model, turn count is computed automatically: `1 + Distance(current homeworld, destination, crossingCost)`.

The `MaxHex` field on `AbilityStep` is currently parsed but never enforced. With the spatial model, `UseAssetAbility` movement validates `Distance(from, to, crossingCost) ≤ step.MaxHex`.

<br/>
<br/>

## TOML Schema

**`$SPATIAL_DATA_DIR/regions.toml`** — owned by Codex (or faction tool before Codex exists):

```toml
[[region]]
id    = "corona-reach"
name  = "The Corona Reach"
hexes = [[0,0],[1,0],[2,0],[0,1],[1,1],[0,2],[1,2],[2,2]]

  [[region.boundary]]
  from_q    = 0
  from_r    = 2
  to_region = "void-expanse"
  to_q      = 3
  to_r      = 1
```

**`$SPATIAL_DATA_DIR/fragments.toml`** — owned by Codex:

```toml
[[fragment]]
id         = "tartarus"
name       = "Tartarus"
tech_level = 2
population = 500000
region     = "corona-reach"
hex_q      = 3
hex_r      = 4
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
name  = "Primitive"

[[tech_level]]
level = 1
name  = "Arcane"

[[tech_level]]
level = 2
name  = "Arcano-Tech"

[[tech_level]]
level = 3
name  = "Astral-Aware"

[[tech_level]]
level = 4
name  = "Aether-Tech"
```

<br/>
<br/>

## Faction Implementation Phases

The spatial model itself (types, distance algorithm, index) is one deliverable. Full faction-tool functionality requires additional work that depends on the spatial model being in place. These are tracked as implementation phases within this initiative, not deferred.

- **Drift rating on assets** — `AssetDefinition` gains a `drift_rating` field; loaded from the campaign's asset TOML files; passed as the basis for `crossingCost` in distance queries
- **Tech level enforcement** — `BuyAsset.Validate` and `ExpandInfluence.Validate` filter purchasable assets and destinations by Fragment tech level; requires Fragment lookup via the spatial map
- **P-flag enforcement** — `BuyAsset.Validate` checks the `P` flag against which faction holds Planetary Government on the target Fragment; requires `BasesByFragment` from the spatial index
- **`MaxHex` enforcement** — `UseAssetAbility` movement validates `Distance(from, to, crossingCost) ≤ step.MaxHex`; previously enforced only by GM adjudication
- **`Change Homeworld` distance computation** — turn count computed automatically from `Distance(current, destination, crossingCost)` rather than entered manually by the GM
- **`faction create` wizard** — homeworld selection step against the Fragment registry; replaces the current free-text world entry

**Tracked separately:** The Pirates tag mechanic (movement cost per hop) depends on drift rating and this spatial model, but belongs to the programmatic tag handling initiative already tracked in planned-work.

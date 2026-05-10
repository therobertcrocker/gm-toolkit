# Spatial Model — Effort 1: `internal/spatial` Package

> Overview: [spatial-model-plan.md](spatial-model-plan.md)

One phase. Deliverable: a standalone, tested spatial library. No faction engine changes in this effort.

<br/>
<br/>

# Phase 1 — Spatial Types, Loader, and Distance

<br/>

## Commit 1 — `feat: add internal/spatial interfaces and HybridMap types`

### Task 1 — `internal/spatial/spatial.go`

Define the two package-level interfaces:

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

### Task 2 — `internal/spatial/hybrid_map.go` — types only

Define all HybridMap data types. No methods yet.

```go
type HexCoord struct{ Q, R int }

type BoundaryConnection struct {
    From     HexCoord
    ToRegion string
    To       HexCoord
}

type Region struct {
    ID         string
    Name       string
    Hexes      map[HexCoord]bool
    Boundaries []BoundaryConnection
}

type Fragment struct {
    FragmentID   string
    FragmentName string
    FragTechLevel  int
    FragPopulation int
    Region       string
    Hex          HexCoord
}

func (f *Fragment) ID() string       { return f.FragmentID }
func (f *Fragment) Name() string     { return f.FragmentName }
func (f *Fragment) TechLevel() int   { return f.FragTechLevel }
func (f *Fragment) Population() int  { return f.FragPopulation }

type HybridMap struct {
    regions   map[string]*Region
    fragments map[string]*Fragment
}
```

`Fragment` implements `Location`. `Region` and `Hex` fields are exported so the faction engine can read them directly.

### Task 3 — `internal/spatial/hex_map.go` and `internal/spatial/graph_map.go` — stubs

```go
type HexMap struct{}

func (h *HexMap) Location(_ string) (Location, bool)              { return nil, false }
func (h *HexMap) Distance(_, _ string, _ int) (int, error)        { return 0, errors.New("not implemented") }
```

Same shape for `GraphMap`. These exist to satisfy the interface contract and signal that the types are intended for future implementation.

<br/>

## Commit 2 — `feat: implement HybridMap TOML loader`

### Task 4 — TOML intermediate types in `hybrid_map.go`

Define unexported structs used only during decoding:

```go
type tomlRegion struct {
    ID         string      `toml:"id"`
    Name       string      `toml:"name"`
    Hexes      [][2]int    `toml:"hexes"`
    Boundaries []tomlBoundary `toml:"boundary"`
}

type tomlBoundary struct {
    FromQ    int    `toml:"from_q"`
    FromR    int    `toml:"from_r"`
    ToRegion string `toml:"to_region"`
    ToQ      int    `toml:"to_q"`
    ToR      int    `toml:"to_r"`
}

type tomlFragment struct {
    ID         string `toml:"id"`
    Name       string `toml:"name"`
    TechLevel  int    `toml:"tech_level"`
    Population int    `toml:"population"`
    Region     string `toml:"region"`
    HexQ       int    `toml:"hex_q"`
    HexR       int    `toml:"hex_r"`
}
```

### Task 5 — `LoadHybrid(dataDir string) (*HybridMap, error)`

```go
func LoadHybrid(dataDir string) (*HybridMap, error)
```

Steps:
1. Decode `regions.toml` into `[]tomlRegion`; convert each to `*Region` with `Hexes` as `map[HexCoord]bool`
2. Decode `fragments.toml` into `[]tomlFragment`; convert each to `*Fragment`
3. Validate: for each Fragment, check its Region exists and its `HexCoord` is in `region.Hexes`; return a descriptive error if not
4. Return `*HybridMap` with both maps populated

### Task 6 — `HybridMap.Location()`

```go
func (m *HybridMap) Location(id string) (Location, bool) {
    f, ok := m.fragments[id]
    return f, ok
}
```

<br/>

## Commit 3 — `feat: implement HybridMap Distance via Dijkstra`

### Task 7 — graph node type (unexported, in `hybrid_map.go`)

```go
type hexNode struct {
    regionID string
    coord    HexCoord
}
```

### Task 8 — six axial neighbor directions (package-level var)

```go
var hexNeighbors = [6]HexCoord{
    {1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, -1}, {-1, 1},
}
```

### Task 9 — `HybridMap.Distance()`

```go
func (m *HybridMap) Distance(fromID, toID string, crossingCost int) (int, error)
```

Algorithm:
1. Look up both Fragments; return error if either is unknown
2. If `fromID == toID`, return 0
3. Build source node `{fragment.Region, fragment.Hex}` and target node
4. Run Dijkstra using a min-heap priority queue:
   - **Intra-region relaxation:** for each of 6 axial neighbors, if the neighbor coord exists in `region.Hexes`, add edge with cost 1
   - **Cross-region relaxation:** for each `BoundaryConnection` whose `From` matches the current hex, add edge to `{toRegion, toCoord}` with cost `crossingCost`; boundary connections are bidirectional — also add the reverse edge during graph traversal
5. Return the settled cost at the target node, or an error if the target is unreachable

Use `container/heap` from the standard library for the priority queue.

<br/>

## Commit 4 — `test: add HybridMap distance unit tests`

### Task 10 — `internal/spatial/hybrid_map_test.go`

Build a small test `HybridMap` inline (do not load from TOML files) for deterministic test control. Cover all cases:

| Test | Setup | Expected |
|---|---|---|
| Same fragment | `Distance("a", "a", _)` | 0 |
| Adjacent intra-region | Two fragments on neighboring hexes in one Region | 1 |
| Non-adjacent intra-region | Two fragments 3 hex-steps apart, clear path | 3 |
| Hole routing | Grid with a missing hex forcing a detour | actual BFS cost > straight-line |
| Cross-region (single boundary) | Two Regions connected by one boundary | intra-A + crossingCost + intra-B |
| Cross-region (two boundaries) | Three Regions chained | sum of all segments |
| Unknown fragment | One or both IDs not in map | error |
| No path | Two disconnected Regions with no boundary | error |

<br/>
<br/>

## TOML Schema Reference

**`$SPATIAL_DATA_DIR/regions.toml`**

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

**`$SPATIAL_DATA_DIR/fragments.toml`**

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

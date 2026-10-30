# Spatial Warp Redesign — Implementation Plan

## Context / Goal

Two coupled changes, shipped as two commits (refactor first, feature second):

1. **Adjacency is hex geometry, not stored data (Model A).** Per the rules,
   crossing a shared border between adjacent regions is an ordinary hex step (cost
   1) — identical to an intra-region step. Today `neighbors` charges every
   inter-region edge the caller's `crossingCost` (the spike-drive `DriftCost`),
   which is a live bug, *and* stores derived adjacency edges redundant with hex
   adjacency. We drop stored adjacency: movement steps to any adjacent **occupied**
   hex at cost 1 via a load-time `map[HexCoord]→region` index; **warps are the only
   stored edges**, at `crossingCost`.
2. **Derived warps.** Replace hard-coded warp entries with a rule: any boundary hex
   marked with a `*` suffix in `layout.txt` warps to every marked hex in another
   region with clear line-of-sight to it.

Commit 1 lands the model/cost change while keeping warps authored explicitly (the
existing `[[warps]]` in `data.toml`). Commit 2 swaps explicit authoring for
`*`-mark derivation. This isolates the risky runtime model swap (reviewable as a
behavior-preserving refactor + bugfix) from the feature.

Model A removes `deriveBoundaries`, the `[[region.boundary]]` entries and their
load checks, the O(edges) reverse-scan in `neighbors`, the `AdjacencyCount` metric,
and the `BoundaryConnection` type. The reverse-scan goes because adjacency is now
derived symmetrically from `hexIndex` (no owned edges to scan back through) and
warps are expanded bidirectionally into `warps map[HexCoord][]warpLink` once at
load. Cost moves from per-query to build-once.

Equivalence: `deriveBoundaries` only ever created a boundary where `offsetNeighbors`
found a different-region neighbor (already 6-adjacent), so hex-neighbor enumeration
yields the same crossings. Warps span gaps (not hex-adjacent) and are the only
edges needing storage. `BoundaryConnection`/`.Boundaries` are used only in
`region_map.go` and `emit.go`; nothing else consumes adjacency data.

No discovery doc — design settled in conversation.

## Decisions Ratified in Planning

1. **Mark warp hexes with `*` in `layout.txt`** (region or world glyph: `A*`, `2*`).
   The `[[warps]]` table is removed (Commit 2).
2. **Line-of-sight is permissive (cube linedraw)** — lerp + `cube_round` with an
   epsilon nudge; corner grazes pass.
3. **Occupied region hexes block; only the two endpoints are excepted.** Empty/void
   transparent. A mark's own region body can block it.
4. **Markable hex = boundary hex** (≥1 neighbor not same-region; empty/OOB count).
   Interior mark is a fatal error. World hexes markable when on the boundary.
5. **A mark with no different-region LoS partner is a warning, not an error**
   (exit 0; stderr).
6. **Model A: warps are the only stored edges; adjacency is hex geometry.** Load
   builds `hexIndex` (hex→region) and `warps` (bidirectional). `neighbors` =
   adjacent occupied hexes at cost 1 + warp links at `crossingCost`. No stored
   adjacency, no `AdjacencyCount`, no reverse-scan.
7. **On disk: a top-level `[[warp]]` table; regions carry only `hexes`.** Stored in
   canonical (alpha-first `from`) orientation; load expands to both directions.
8. **A warp between two hex-adjacent marks is suppressed** (Commit 2): they already
   connect via a hex step — counted as connected, no warp edge.
9. **Validated against the reference example** — marks `A*=(1,3)`, `D*=(2,12)`,
   `B*=(3,6)`, `C*=(6,9)` give A\*↔D\*, B\*↔D\*, B\*↔C\*, C\*↔D\*; A\*↔B\*, A\*↔C\*
   blocked.
10. **Migration is not connection-preserving** — fixtures regenerate.

## Out of Scope

- Any edge cost/kind beyond "warp vs hex step."
- Strict-TOML rejection of leftover `[[warps]]`/`[[region.boundary]]` (non-strict
  decode ignores them; the migration rewrites all affected files).

---

## Work Breakdown

Two commits. Each compiles and tests green at its end. WIP commits within are fine.

---

### Commit 1 — `refactor(spatial): model adjacency as hex geometry, fix crossing cost`

Drops stored adjacency and the reverse-scan, fixes the border cost, changes the
file format to top-level `[[warp]]`. Warps stay authored via `data.toml`
`[[warps]]`, now converted into the new warp list. No `*` marks, no LoS yet.

#### Task 1.1 — `internal/spatial/region_map.go`

**1.1a.** Replace `BoundaryConnection` + `Region` with warp link/record types and a
hexes-only region. **Find:**
```go
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
```
**Replace with:**
```go
type Region struct {
	ID    string
	Name  string
	Hexes map[HexCoord]bool
}

// warpLink is one direction of a warp, keyed by its origin hex in RegionMap.warps.
type warpLink struct {
	toRegion string
	to       HexCoord
}

// warpRecord is a warp in canonical (alpha-first) orientation, before expansion
// into the bidirectional warps index.
type warpRecord struct {
	fromRegion string
	from       HexCoord
	toRegion   string
	to         HexCoord
}
```

**1.1b.** Add indexes to `RegionMap`. **Find:**
```go
type RegionMap struct {
	regions map[string]*Region
	worlds  map[string]*World
}
```
**Replace with:**
```go
type RegionMap struct {
	regions  map[string]*Region
	worlds   map[string]*World
	hexIndex map[HexCoord]string
	warps    map[HexCoord][]warpLink
}
```

**1.1c.** Replace TOML decode types. **Find:**
```go
type tomlRegion struct {
	ID         string         `toml:"id"`
	Name       string         `toml:"name"`
	Hexes      [][2]int       `toml:"hexes"`
	Boundaries []tomlBoundary `toml:"boundary"`
}

type tomlBoundary struct {
	FromQ    int    `toml:"from_q"`
	FromR    int    `toml:"from_r"`
	ToRegion string `toml:"to_region"`
	ToQ      int    `toml:"to_q"`
	ToR      int    `toml:"to_r"`
}
```
**Replace with:**
```go
type tomlRegion struct {
	ID    string   `toml:"id"`
	Name  string   `toml:"name"`
	Hexes [][2]int `toml:"hexes"`
}

type tomlWarp struct {
	FromRegion string `toml:"from_region"`
	FromQ      int    `toml:"from_q"`
	FromR      int    `toml:"from_r"`
	ToRegion   string `toml:"to_region"`
	ToQ        int    `toml:"to_q"`
	ToR        int    `toml:"to_r"`
}
```

**1.1d.** Add warps to `regionsFile`. **Find:**
```go
type regionsFile struct {
	Region []tomlRegion `toml:"region"`
}
```
**Replace with:**
```go
type regionsFile struct {
	Region []tomlRegion `toml:"region"`
	Warp   []tomlWarp   `toml:"warp"`
}
```

**1.1e.** Replace the region/boundary build + validation block in `LoadRegionMap`
(`region_map.go:91`–`region_map.go:126`). **Find:**
```go
	regions := make(map[string]*Region, len(regionsDoc.Region))
	for _, region := range regionsDoc.Region {
		hexes := make(map[HexCoord]bool, len(region.Hexes))
		for _, hex := range region.Hexes {
			hexes[HexCoord{Q: hex[0], R: hex[1]}] = true
		}
		boundaries := make([]BoundaryConnection, 0, len(region.Boundaries))
		for _, boundary := range region.Boundaries {
			boundaries = append(boundaries, BoundaryConnection{
				From:     HexCoord{Q: boundary.FromQ, R: boundary.FromR},
				ToRegion: boundary.ToRegion,
				To:       HexCoord{Q: boundary.ToQ, R: boundary.ToR},
			})
		}
		regions[region.ID] = &Region{
			ID:         region.ID,
			Name:       region.Name,
			Hexes:      hexes,
			Boundaries: boundaries,
		}
	}

	for _, region := range regions {
		for _, boundary := range region.Boundaries {
			if !region.Hexes[boundary.From] {
				return nil, fmt.Errorf("region %q boundary From=(%d,%d) is not in region's hexes", region.ID, boundary.From.Q, boundary.From.R)
			}
			target, ok := regions[boundary.ToRegion]
			if !ok {
				return nil, fmt.Errorf("region %q boundary references unknown region %q", region.ID, boundary.ToRegion)
			}
			if !target.Hexes[boundary.To] {
				return nil, fmt.Errorf("region %q boundary To=(%d,%d) is not in region %q's hexes", region.ID, boundary.To.Q, boundary.To.R, boundary.ToRegion)
			}
		}
	}
```
**Replace with:**
```go
	regions := make(map[string]*Region, len(regionsDoc.Region))
	for _, region := range regionsDoc.Region {
		hexes := make(map[HexCoord]bool, len(region.Hexes))
		for _, hex := range region.Hexes {
			hexes[HexCoord{Q: hex[0], R: hex[1]}] = true
		}
		regions[region.ID] = &Region{ID: region.ID, Name: region.Name, Hexes: hexes}
	}

	warpRecords := make([]warpRecord, 0, len(regionsDoc.Warp))
	for _, warp := range regionsDoc.Warp {
		warpRecords = append(warpRecords, warpRecord{
			fromRegion: warp.FromRegion,
			from:       HexCoord{Q: warp.FromQ, R: warp.FromR},
			toRegion:   warp.ToRegion,
			to:         HexCoord{Q: warp.ToQ, R: warp.ToR},
		})
	}

	regionMap, err := newRegionMap(regions, warpRecords)
	if err != nil {
		return nil, err
	}
```

**1.1f.** The worlds section then assigns into `regionMap`. **Find:**
```go
	worlds := make(map[string]*World, len(worldsDoc.World))
	for _, world := range worldsDoc.World {
		region, ok := regions[world.Region]
		if !ok {
			return nil, fmt.Errorf("world %q references unknown region %q", world.ID, world.Region)
		}
		hex := HexCoord{Q: world.HexQ, R: world.HexR}
		if !region.Hexes[hex] {
			return nil, fmt.Errorf("world %q at hex (%d,%d) is not within region %q", world.ID, hex.Q, hex.R, world.Region)
		}
		worlds[world.ID] = &World{
			id:         world.ID,
			name:       world.Name,
			techLevel:  world.TechLevel,
			population: world.Population,
			loc:        RegionHex{RegionID: world.Region, Coord: hex},
		}
	}

	return &RegionMap{regions: regions, worlds: worlds}, nil
}
```
**Replace with:**
```go
	worlds := make(map[string]*World, len(worldsDoc.World))
	for _, world := range worldsDoc.World {
		region, ok := regions[world.Region]
		if !ok {
			return nil, fmt.Errorf("world %q references unknown region %q", world.ID, world.Region)
		}
		hex := HexCoord{Q: world.HexQ, R: world.HexR}
		if !region.Hexes[hex] {
			return nil, fmt.Errorf("world %q at hex (%d,%d) is not within region %q", world.ID, hex.Q, hex.R, world.Region)
		}
		worlds[world.ID] = &World{
			id:         world.ID,
			name:       world.Name,
			techLevel:  world.TechLevel,
			population: world.Population,
			loc:        RegionHex{RegionID: world.Region, Coord: hex},
		}
	}

	regionMap.worlds = worlds
	return regionMap, nil
}

// newRegionMap builds the hex→region index and bidirectional warp index,
// validating one-region-per-hex and that each warp endpoint resolves to its
// declared region.
func newRegionMap(regions map[string]*Region, warpRecords []warpRecord) (*RegionMap, error) {
	hexIndex := make(map[HexCoord]string)
	for id, region := range regions {
		for hex := range region.Hexes {
			if other, dup := hexIndex[hex]; dup {
				return nil, fmt.Errorf("hex (%d,%d) is claimed by both region %q and region %q", hex.Q, hex.R, other, id)
			}
			hexIndex[hex] = id
		}
	}

	warps := make(map[HexCoord][]warpLink)
	for _, warp := range warpRecords {
		if hexIndex[warp.from] != warp.fromRegion {
			return nil, fmt.Errorf("warp From=(%d,%d) is not in region %q's hexes", warp.from.Q, warp.from.R, warp.fromRegion)
		}
		if _, ok := regions[warp.toRegion]; !ok {
			return nil, fmt.Errorf("warp references unknown region %q", warp.toRegion)
		}
		if hexIndex[warp.to] != warp.toRegion {
			return nil, fmt.Errorf("warp To=(%d,%d) is not in region %q's hexes", warp.to.Q, warp.to.R, warp.toRegion)
		}
		warps[warp.from] = append(warps[warp.from], warpLink{toRegion: warp.toRegion, to: warp.to})
		warps[warp.to] = append(warps[warp.to], warpLink{toRegion: warp.fromRegion, to: warp.from})
	}

	return &RegionMap{regions: regions, hexIndex: hexIndex, warps: warps}, nil
}
```
(`err` is already declared via `:=` here; if the compiler reports it unused before
this point, the existing `LoadRegionMap` already uses `err` for the decode calls —
keep those `:=`/`=` consistent.)

**1.1g.** Replace `RegionOfHex`. **Find:**
```go
func (regionMap *RegionMap) RegionOfHex(hex HexCoord) (string, bool) {
	for _, region := range regionMap.regions {
		if region.Hexes[hex] {
			return region.ID, true
		}
	}
	return "", false
}
```
**Replace with:**
```go
func (regionMap *RegionMap) RegionOfHex(hex HexCoord) (string, bool) {
	id, ok := regionMap.hexIndex[hex]
	return id, ok
}
```

**1.1h.** Replace the body of `neighbors`. **Find:**
```go
func (regionMap *RegionMap) neighbors(node hexNode, crossingCost int) []edge {
	var edges []edge

	region := regionMap.regions[node.regionID]

	for _, delta := range hexNeighbors {
		candidate := HexCoord{Q: node.coord.Q + delta.Q, R: node.coord.R + delta.R}
		if region.Hexes[candidate] {
			edges = append(edges, edge{to: hexNode{regionID: node.regionID, coord: candidate}, cost: 1})
		}
	}

	for _, boundary := range region.Boundaries {
		if boundary.From == node.coord {
			edges = append(edges, edge{to: hexNode{regionID: boundary.ToRegion, coord: boundary.To}, cost: crossingCost})
		}
	}

	for _, otherRegion := range regionMap.regions {
		if otherRegion.ID == node.regionID {
			continue
		}
		for _, boundary := range otherRegion.Boundaries {
			if boundary.ToRegion == node.regionID && boundary.To == node.coord {
				edges = append(edges, edge{to: hexNode{regionID: otherRegion.ID, coord: boundary.From}, cost: crossingCost})
			}
		}
	}

	return edges
}
```
**Replace with:**
```go
func (regionMap *RegionMap) neighbors(node hexNode, crossingCost int) []edge {
	var edges []edge

	for _, delta := range hexNeighbors {
		candidate := HexCoord{Q: node.coord.Q + delta.Q, R: node.coord.R + delta.R}
		if regionID, ok := regionMap.hexIndex[candidate]; ok {
			edges = append(edges, edge{to: hexNode{regionID: regionID, coord: candidate}, cost: 1})
		}
	}

	for _, link := range regionMap.warps[node.coord] {
		edges = append(edges, edge{to: hexNode{regionID: link.toRegion, coord: link.to}, cost: crossingCost})
	}

	return edges
}
```

#### Task 1.2 — `internal/spatial/builder/emit.go`

Regions hexes-only; warps as a top-level table. **Find:**
```go
type emitRegion struct {
	ID         string         `toml:"id"`
	Name       string         `toml:"name"`
	Hexes      [][2]int       `toml:"hexes"`
	Boundaries []emitBoundary `toml:"boundary"`
}

type emitBoundary struct {
	FromQ    int    `toml:"from_q"`
	FromR    int    `toml:"from_r"`
	ToRegion string `toml:"to_region"`
	ToQ      int    `toml:"to_q"`
	ToR      int    `toml:"to_r"`
}
```
**Replace with:**
```go
type emitRegion struct {
	ID    string   `toml:"id"`
	Name  string   `toml:"name"`
	Hexes [][2]int `toml:"hexes"`
}

type emitWarp struct {
	FromRegion string `toml:"from_region"`
	FromQ      int    `toml:"from_q"`
	FromR      int    `toml:"from_r"`
	ToRegion   string `toml:"to_region"`
	ToQ        int    `toml:"to_q"`
	ToR        int    `toml:"to_r"`
}
```

**Find:**
```go
func writeRegions(path string, derived *derivedMap) error {
	out := struct {
		Region []emitRegion `toml:"region"`
	}{}
	for _, r := range derived.Regions {
		em := emitRegion{ID: r.ID, Name: r.Name}
		for _, h := range r.Hexes {
			em.Hexes = append(em.Hexes, [2]int{h.Q, h.R})
		}
		for _, b := range r.Boundaries {
			em.Boundaries = append(em.Boundaries, emitBoundary{
				FromQ: b.From.Q, FromR: b.From.R,
				ToRegion: b.ToRegion,
				ToQ:      b.To.Q, ToR: b.To.R,
			})
		}
		out.Region = append(out.Region, em)
	}
	return writeTOML(path, out)
}
```
**Replace with:**
```go
func writeRegions(path string, derived *derivedMap) error {
	out := struct {
		Region []emitRegion `toml:"region"`
		Warp   []emitWarp   `toml:"warp"`
	}{}
	for _, r := range derived.Regions {
		em := emitRegion{ID: r.ID, Name: r.Name}
		for _, h := range r.Hexes {
			em.Hexes = append(em.Hexes, [2]int{h.Q, h.R})
		}
		out.Region = append(out.Region, em)
	}
	for _, w := range derived.Warps {
		out.Warp = append(out.Warp, emitWarp{
			FromRegion: w.FromRegion, FromQ: w.From.Q, FromR: w.From.R,
			ToRegion: w.ToRegion, ToQ: w.To.Q, ToR: w.To.R,
		})
	}
	return writeTOML(path, out)
}
```

#### Task 1.3 — `internal/spatial/builder/derive.go`

Remove adjacency derivation; warps become a flat canonical list (still sourced from
`data.Warps`). This is Commit 1's end-state for the file's warp/region types and
the `applyWarps` path.

**1.3a.** Types. **Find:**
```go
type derivedMap struct {
	Regions        []derivedRegion
	Worlds         []derivedWorld
	AdjacencyCount int
	WarpCount      int
}

type derivedRegion struct {
	ID         string
	Name       string
	Hexes      []spatial.HexCoord
	Boundaries []derivedBoundary
}

type derivedBoundary struct {
	From     spatial.HexCoord
	ToRegion string
	To       spatial.HexCoord
}
```
**Replace with:**
```go
type derivedMap struct {
	Regions   []derivedRegion
	Worlds    []derivedWorld
	Warps     []derivedWarp
	WarpCount int
}

type derivedRegion struct {
	ID    string
	Name  string
	Hexes []spatial.HexCoord
}

type derivedWarp struct {
	FromRegion string
	From       spatial.HexCoord
	ToRegion   string
	To         spatial.HexCoord
}
```

**1.3b.** Trim the `derivation` struct (drop `boundariesByRegion`, `seen`,
`adjacencyCount`). **Find:**
```go
type derivation struct {
	regions            map[string]*regionAcc
	cellRegion         map[gridCoord]string
	boundariesByRegion map[string][]derivedBoundary
	seen               map[edgeKey]bool
	adjacencyCount     int
}

func newDerivation(data *dataFile) *derivation {
	regions := make(map[string]*regionAcc, len(data.Regions))
	for id, entry := range data.Regions {
		regions[id] = &regionAcc{name: entry.Name, hexes: make(map[spatial.HexCoord]bool)}
	}
	return &derivation{
		regions:            regions,
		cellRegion:         map[gridCoord]string{},
		boundariesByRegion: map[string][]derivedBoundary{},
		seen:               map[edgeKey]bool{},
	}
}
```
**Replace with:**
```go
type derivation struct {
	regions    map[string]*regionAcc
	cellRegion map[gridCoord]string
}

func newDerivation(data *dataFile) *derivation {
	regions := make(map[string]*regionAcc, len(data.Regions))
	for id, entry := range data.Regions {
		regions[id] = &regionAcc{name: entry.Name, hexes: make(map[spatial.HexCoord]bool)}
	}
	return &derivation{
		regions:    regions,
		cellRegion: map[gridCoord]string{},
	}
}
```

**1.3c.** Rewire `derive`. **Find:**
```go
	state.deriveBoundaries(layout)

	if err := state.applyWarps(data.Warps, layout); err != nil {
		return nil, err
	}

	return state.assemble(worlds, len(data.Warps)), nil
}
```
**Replace with:**
```go
	warps, err := state.applyWarps(data.Warps, layout)
	if err != nil {
		return nil, err
	}

	return state.assemble(worlds, warps), nil
}
```

**1.3d.** Delete `deriveBoundaries` (`derive.go:239`–`derive.go:264`). **Replace
with:** (nothing — remove it).

**1.3e.** Rewrite `applyWarps` to return a canonical `[]derivedWarp` (keep
`validateWarpEndpoint` as-is, just below). **Find** the `applyWarps` function body
(`derive.go:266`–`derive.go:291`):
```go
func (state *derivation) applyWarps(warps []warpEntry, layout *layoutFile) error {
	for i, warp := range warps {
		context := fmt.Sprintf("[[warps]] #%d", i+1)
		if err := state.validateWarpEndpoint(layout, warp.From, context, "from"); err != nil {
			return err
		}
		if err := state.validateWarpEndpoint(layout, warp.To, context, "to"); err != nil {
			return err
		}
		if warp.From.Region == warp.To.Region {
			return dataErrorf(context, "both endpoints are in region %q; warps must cross region boundaries", warp.From.Region)
		}
		fromAxial := axialOf(gridCoord{Row: warp.From.Row, Col: warp.From.Col})
		toAxial := axialOf(gridCoord{Row: warp.To.Row, Col: warp.To.Col})
		ownerID, ownerCoord, neighborID, neighborCoord := canonicalEdge(warp.From.Region, fromAxial, warp.To.Region, toAxial)
		key := edgeKey{fromRegion: ownerID, from: ownerCoord, toRegion: neighborID, to: neighborCoord}
		if state.seen[key] {
			return dataErrorf(context, "duplicate warp: hex pair already connected by a boundary or warp")
		}
		state.seen[key] = true
		state.boundariesByRegion[ownerID] = append(state.boundariesByRegion[ownerID], derivedBoundary{
			From: ownerCoord, ToRegion: neighborID, To: neighborCoord,
		})
	}
	return nil
}
```
**Replace with:**
```go
func (state *derivation) applyWarps(warps []warpEntry, layout *layoutFile) ([]derivedWarp, error) {
	seen := map[edgeKey]bool{}
	out := make([]derivedWarp, 0, len(warps))
	for i, warp := range warps {
		context := fmt.Sprintf("[[warps]] #%d", i+1)
		if err := state.validateWarpEndpoint(layout, warp.From, context, "from"); err != nil {
			return nil, err
		}
		if err := state.validateWarpEndpoint(layout, warp.To, context, "to"); err != nil {
			return nil, err
		}
		if warp.From.Region == warp.To.Region {
			return nil, dataErrorf(context, "both endpoints are in region %q; warps must cross region boundaries", warp.From.Region)
		}
		fromAxial := axialOf(gridCoord{Row: warp.From.Row, Col: warp.From.Col})
		toAxial := axialOf(gridCoord{Row: warp.To.Row, Col: warp.To.Col})
		fromRegion, fromCoord, toRegion, toCoord := canonicalEdge(warp.From.Region, fromAxial, warp.To.Region, toAxial)
		key := edgeKey{fromRegion: fromRegion, from: fromCoord, toRegion: toRegion, to: toCoord}
		if seen[key] {
			return nil, dataErrorf(context, "duplicate warp: hex pair already connected")
		}
		seen[key] = true
		out = append(out, derivedWarp{FromRegion: fromRegion, From: fromCoord, ToRegion: toRegion, To: toCoord})
	}
	sort.Slice(out, func(i, j int) bool { return warpLess(out[i], out[j]) })
	return out, nil
}
```

**1.3f.** Replace `assemble` and `boundaryLess` (`derive.go:313`–`derive.go:359`).
**Find:**
```go
func (state *derivation) assemble(worlds []derivedWorld, warpCount int) *derivedMap {
	out := &derivedMap{
		AdjacencyCount: state.adjacencyCount,
		WarpCount:      warpCount,
	}
	ids := make([]string, 0, len(state.regions))
	for id := range state.regions {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		acc := state.regions[id]
		hexes := make([]spatial.HexCoord, 0, len(acc.hexes))
		for h := range acc.hexes {
			hexes = append(hexes, h)
		}
		sort.Slice(hexes, func(i, j int) bool {
			if hexes[i].Q != hexes[j].Q {
				return hexes[i].Q < hexes[j].Q
			}
			return hexes[i].R < hexes[j].R
		})
		boundaries := state.boundariesByRegion[id]
		sort.Slice(boundaries, func(i, j int) bool { return boundaryLess(boundaries[i], boundaries[j]) })
		out.Regions = append(out.Regions, derivedRegion{
			ID: id, Name: acc.name, Hexes: hexes, Boundaries: boundaries,
		})
	}
	sort.Slice(worlds, func(i, j int) bool { return worlds[i].ID < worlds[j].ID })
	out.Worlds = worlds
	return out
}

func boundaryLess(a, b derivedBoundary) bool {
	switch {
	case a.ToRegion != b.ToRegion:
		return a.ToRegion < b.ToRegion
	case a.From.Q != b.From.Q:
		return a.From.Q < b.From.Q
	case a.From.R != b.From.R:
		return a.From.R < b.From.R
	case a.To.Q != b.To.Q:
		return a.To.Q < b.To.Q
	default:
		return a.To.R < b.To.R
	}
}
```
**Replace with:**
```go
func (state *derivation) assemble(worlds []derivedWorld, warps []derivedWarp) *derivedMap {
	out := &derivedMap{Warps: warps, WarpCount: len(warps)}
	ids := make([]string, 0, len(state.regions))
	for id := range state.regions {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		acc := state.regions[id]
		hexes := make([]spatial.HexCoord, 0, len(acc.hexes))
		for h := range acc.hexes {
			hexes = append(hexes, h)
		}
		sort.Slice(hexes, func(i, j int) bool {
			if hexes[i].Q != hexes[j].Q {
				return hexes[i].Q < hexes[j].Q
			}
			return hexes[i].R < hexes[j].R
		})
		out.Regions = append(out.Regions, derivedRegion{ID: id, Name: acc.name, Hexes: hexes})
	}
	sort.Slice(worlds, func(i, j int) bool { return worlds[i].ID < worlds[j].ID })
	out.Worlds = worlds
	return out
}

func warpLess(a, b derivedWarp) bool {
	switch {
	case a.FromRegion != b.FromRegion:
		return a.FromRegion < b.FromRegion
	case a.From.Q != b.From.Q:
		return a.From.Q < b.From.Q
	case a.From.R != b.From.R:
		return a.From.R < b.From.R
	case a.ToRegion != b.ToRegion:
		return a.ToRegion < b.ToRegion
	case a.To.Q != b.To.Q:
		return a.To.Q < b.To.Q
	default:
		return a.To.R < b.To.R
	}
}
```

(`edgeKey` and `canonicalEdge` stay — `applyWarps` still uses both. `regionAcc`
unchanged.)

#### Task 1.4 — `internal/spatial/builder/builder.go`

Drop `AdjacencyCount`. **Find:**
```go
type Summary struct {
	Regions        []string
	WorldCount     int
	AdjacencyCount int
	WarpCount      int
}
```
**Replace with:**
```go
type Summary struct {
	Regions    []string
	WorldCount int
	WarpCount  int
}
```
**Find:**
```go
	s := Summary{
		WorldCount:     len(derived.Worlds),
		WarpCount:      derived.WarpCount,
		AdjacencyCount: derived.AdjacencyCount,
	}
```
**Replace with:**
```go
	s := Summary{
		WorldCount: len(derived.Worlds),
		WarpCount:  derived.WarpCount,
	}
```

#### Task 1.5 — `cmd/gm-toolkit/spatial/build.go`

Drop the adjacency line. **Find:**
```go
	fmt.Fprintf(out, "  %d worlds\n", s.WorldCount)
	fmt.Fprintf(out, "  %d adjacency boundaries\n", s.AdjacencyCount)
	fmt.Fprintf(out, "  %d warps\n", s.WarpCount)
```
**Replace with:**
```go
	fmt.Fprintf(out, "  %d worlds\n", s.WorldCount)
	fmt.Fprintf(out, "  %d warps\n", s.WarpCount)
```

#### Task 1.6 — Reformat fixture goldens (no source change)

`internal/spatial/builder/testdata/valid/helions_reach/golden/regions.toml` —
regenerate in the new format: each `[[region]]` keeps `id`/`name`/`hexes`, the
`[[region.boundary]]` blocks are gone, and the existing explicit warp becomes one
top-level `[[warp]]`:
```toml
[[warp]]
  from_region = "A"
  from_q = 1
  from_r = 1
  to_region = "D"
  to_q = 13
  to_r = 1
```
`layout.txt`, `source/data.toml`, and `golden/worlds.toml` are unchanged.

`campaigns/test-camp/spatial/regions.toml` — regenerate via
`go run ./cmd/gm-toolkit spatial build --campaign test-camp --replace`; its two
explicit warps become two `[[warp]]` entries, boundaries gone.

(The `warp_wrong_region` error fixture still validates — `validateWarpEndpoint` and
`[[warps]]` remain in Commit 1.)

#### Task 1.7 — Update unit tests for the new model

`internal/spatial/region_map_test.go` — build maps via `newRegionMap(regions,
warpRecords)` (set `.worlds` after) instead of `&RegionMap{...}` literals so
`hexIndex`/`warps` populate:
- Intra-region cases unchanged in expectation.
- "cross-region single boundary" → model two regions whose hexes are 6-adjacent
  across the border (no warp); cost is now `3` (1+1+1) at any `crossingCost`. This
  is the cost-bug regression.
- Add a **warp** case: two regions joined only by a `warpRecord` over a gap; cost
  includes `crossingCost` for the hop; add a reverse-direction case.
- "bidirectional traversal", "boundary shortcut", "no path", "negative cost":
  re-express with `warpRecord`s / adjacency.
- `TestLoadRegionMap`: replace `[[region.boundary]]` fixtures with `[[warp]]`;
  update error cases to the new `newRegionMap` messages (`warp From=(…) is not in
  region …`, `warp references unknown region …`, `warp To=(…) is not in region …`);
  add a happy-path assertion that a loaded warp appears in both directions of
  `regionMap.warps`; assert no `Boundaries` field exists (compile-time once removed).

`internal/spatial/builder/emit_test.go` — the round-trip fixture's `derivedMap`
uses `Warps []derivedWarp` and regions without boundaries. Assert one `[[region]]`
per region (hexes only) and that the `[[warp]]` table round-trips.

`internal/spatial/builder/derive_test.go` — `applyWarps` now returns
`([]derivedWarp, error)`; `assemble` takes `(worlds, warps)`. Update call sites and
assertions (assert on `derivedWarp`s; remove `derivedBoundary`/`AdjacencyCount`).
`mkData` keeps its `warps []warpEntry` parameter.

##### Commit message
```
refactor(spatial): model adjacency as hex geometry, fix crossing cost

- stop storing adjacency edges; a region crossing is an ordinary hex step
  (cost 1) resolved via a load-time hex→region index
- warps are the only stored edges, at the caller's crossing cost; this fixes
  borders costing a full drift jump
- expand warps bidirectionally at load, removing the per-query reverse scan
- regions.toml: regions carry hexes only; warps are a top-level [[warp]] table
- remove deriveBoundaries, BoundaryConnection, and AdjacencyCount
```

---

### Commit 2 — `feat(spatial): derive warps from line-of-sight between marked hexes`

Swaps explicit `[[warps]]` authoring for `*`-mark derivation. Anchors below match
Commit 1's end-state.

#### Task 2.1 — `internal/spatial/builder/layout.go`

Add a `Warp` flag to `cell`. **Find:**
```go
type cell struct {
	Glyph    glyph
	Row      int
	Col      int
	FileLine int
	FileCol  int
}
```
**Replace with:**
```go
type cell struct {
	Glyph    glyph
	Row      int
	Col      int
	Warp     bool
	FileLine int
	FileCol  int
}
```

In `parseLayoutRow`, consume an optional `*` after a region/world glyph. **Find:**
```go
		c := cell{Glyph: g, Row: row, Col: cellCol, FileLine: fileLine, FileCol: fileCol}
		cells = append(cells, c)

		if g.isWorld() {
			if prior, seen := markers[g]; seen {
				return nil, layoutErrorf(fileLine, fileCol, "world marker %q appears twice; first at line=%d col=%d", ch, prior.FileLine, prior.FileCol)
			}
			markers[g] = c
		}
		cellCol++
		i++
		if i >= len(line) {
			break
		}
		if line[i] != ' ' {
			return nil, layoutErrorf(fileLine, i+indentOffset+1, "expected single-space separator, got %q", line[i])
		}
		i++
```
**Replace with:**
```go
		c := cell{Glyph: g, Row: row, Col: cellCol, FileLine: fileLine, FileCol: fileCol}
		i++
		if (g.isRegion() || g.isWorld()) && i < len(line) && line[i] == '*' {
			c.Warp = true
			i++
		}
		cells = append(cells, c)

		if g.isWorld() {
			if prior, seen := markers[g]; seen {
				return nil, layoutErrorf(fileLine, fileCol, "world marker %q appears twice; first at line=%d col=%d", ch, prior.FileLine, prior.FileCol)
			}
			markers[g] = c
		}
		cellCol++
		if i >= len(line) {
			break
		}
		if line[i] != ' ' {
			return nil, layoutErrorf(fileLine, i+indentOffset+1, "expected single-space separator, got %q", line[i])
		}
		i++
```

The glyph's `i++` now runs before the marker peek; the trailing `i++` is removed.
A `*` after `.` or standalone falls through to the existing separator/glyph errors.

#### Task 2.2 — `internal/spatial/builder/geometry.go` (new) — full contents:

```go
package builder

import (
	"math"

	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

// Cube-coordinate hex line drawing, used to test warp line-of-sight. See
// https://www.redblobgames.com/grids/hexagons/#line-drawing.

type cubeCoord struct{ x, y, z int }

type fcube struct{ x, y, z float64 }

func axialToCube(h spatial.HexCoord) cubeCoord {
	return cubeCoord{x: h.Q, y: -h.Q - h.R, z: h.R}
}

func cubeToAxial(c cubeCoord) spatial.HexCoord {
	return spatial.HexCoord{Q: c.x, R: c.z}
}

func cubeDistance(a, b cubeCoord) int {
	return (absInt(a.x-b.x) + absInt(a.y-b.y) + absInt(a.z-b.z)) / 2
}

func hexDistance(a, b spatial.HexCoord) int {
	return cubeDistance(axialToCube(a), axialToCube(b))
}

func cubeRound(c fcube) cubeCoord {
	rx := math.Round(c.x)
	ry := math.Round(c.y)
	rz := math.Round(c.z)
	dx := math.Abs(rx - c.x)
	dy := math.Abs(ry - c.y)
	dz := math.Abs(rz - c.z)
	switch {
	case dx > dy && dx > dz:
		rx = -ry - rz
	case dy > dz:
		ry = -rx - rz
	default:
		rz = -rx - ry
	}
	return cubeCoord{x: int(rx), y: int(ry), z: int(rz)}
}

// hexLine returns the hexes the straight line between a and b passes through,
// inclusive of both endpoints. Endpoints are nudged by a sum-zero epsilon so a
// line crossing exactly on a hex border resolves deterministically.
func hexLine(a, b spatial.HexCoord) []spatial.HexCoord {
	ac := axialToCube(a)
	bc := axialToCube(b)
	n := cubeDistance(ac, bc)
	if n == 0 {
		return []spatial.HexCoord{a}
	}
	aN := fcube{float64(ac.x) + 1e-6, float64(ac.y) + 2e-6, float64(ac.z) - 3e-6}
	bN := fcube{float64(bc.x) + 1e-6, float64(bc.y) + 2e-6, float64(bc.z) - 3e-6}
	out := make([]spatial.HexCoord, 0, n+1)
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		out = append(out, cubeToAxial(cubeRound(fcube{
			x: lerp(aN.x, bN.x, t),
			y: lerp(aN.y, bN.y, t),
			z: lerp(aN.z, bN.z, t),
		})))
	}
	return out
}

func lerp(a, b, t float64) float64 { return a + (b-a)*t }

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
```

#### Task 2.3 — `internal/spatial/builder/derive.go`

**2.3a.** Add `Warnings` to `derivedMap`. **Find:**
```go
type derivedMap struct {
	Regions   []derivedRegion
	Worlds    []derivedWorld
	Warps     []derivedWarp
	WarpCount int
}
```
**Replace with:**
```go
type derivedMap struct {
	Regions   []derivedRegion
	Worlds    []derivedWorld
	Warps     []derivedWarp
	WarpCount int
	Warnings  []string
}
```

**2.3b.** Rewire `derive` to call `deriveWarps` (Commit-1 anchor). **Find:**
```go
	warps, err := state.applyWarps(data.Warps, layout)
	if err != nil {
		return nil, err
	}

	return state.assemble(worlds, warps), nil
}
```
**Replace with:**
```go
	warps, warnings, err := state.deriveWarps(layout)
	if err != nil {
		return nil, err
	}

	return state.assemble(worlds, warps, warnings), nil
}
```

**2.3c.** Replace `applyWarps` + `validateWarpEndpoint` (delete both) with
`deriveWarps` and helpers. **Find** (Commit 1's `applyWarps` + the unchanged
`validateWarpEndpoint`):
```go
func (state *derivation) applyWarps(warps []warpEntry, layout *layoutFile) ([]derivedWarp, error) {
	seen := map[edgeKey]bool{}
	out := make([]derivedWarp, 0, len(warps))
	for i, warp := range warps {
		context := fmt.Sprintf("[[warps]] #%d", i+1)
		if err := state.validateWarpEndpoint(layout, warp.From, context, "from"); err != nil {
			return nil, err
		}
		if err := state.validateWarpEndpoint(layout, warp.To, context, "to"); err != nil {
			return nil, err
		}
		if warp.From.Region == warp.To.Region {
			return nil, dataErrorf(context, "both endpoints are in region %q; warps must cross region boundaries", warp.From.Region)
		}
		fromAxial := axialOf(gridCoord{Row: warp.From.Row, Col: warp.From.Col})
		toAxial := axialOf(gridCoord{Row: warp.To.Row, Col: warp.To.Col})
		fromRegion, fromCoord, toRegion, toCoord := canonicalEdge(warp.From.Region, fromAxial, warp.To.Region, toAxial)
		key := edgeKey{fromRegion: fromRegion, from: fromCoord, toRegion: toRegion, to: toCoord}
		if seen[key] {
			return nil, dataErrorf(context, "duplicate warp: hex pair already connected")
		}
		seen[key] = true
		out = append(out, derivedWarp{FromRegion: fromRegion, From: fromCoord, ToRegion: toRegion, To: toCoord})
	}
	sort.Slice(out, func(i, j int) bool { return warpLess(out[i], out[j]) })
	return out, nil
}

func (state *derivation) validateWarpEndpoint(layout *layoutFile, addr hexAddr, context, side string) error {
	if addr.Row < 0 || addr.Row >= len(layout.cells) {
		return dataErrorf(context, "%s.row=%d is out of bounds (grid has %d rows)", side, addr.Row, len(layout.cells))
	}
	if addr.Col < 0 || addr.Col >= len(layout.cells[addr.Row]) {
		return dataErrorf(context, "%s.col=%d is out of bounds for row %d (has %d cells)", side, addr.Col, addr.Row, len(layout.cells[addr.Row]))
	}
	if _, ok := state.regions[addr.Region]; !ok {
		return dataErrorf(context, "%s.region=%q is not a declared region", side, addr.Region)
	}
	actual, ok := state.cellRegion[gridCoord{Row: addr.Row, Col: addr.Col}]
	if !ok {
		return dataErrorf(context, "%s cell at (row=%d, col=%d) is empty, not a region member", side, addr.Row, addr.Col)
	}
	if actual != addr.Region {
		return dataErrorf(context, "%s cell at (row=%d, col=%d) belongs to region %q, not %q", side, addr.Row, addr.Col, actual, addr.Region)
	}
	return nil
}
```
**Replace with:**
```go
type markedHex struct {
	region string
	coord  spatial.HexCoord
	grid   gridCoord
}

// deriveWarps connects every pair of marked hexes in different regions with clear
// line-of-sight that are more than one hex apart (adjacent marks already connect
// via a hex step). A marked hex must lie on a region boundary; a mark that reaches
// no other region is reported as a non-fatal warning.
func (state *derivation) deriveWarps(layout *layoutFile) ([]derivedWarp, []string, error) {
	marks := state.collectMarks(layout)

	for _, mark := range marks {
		if !state.isBoundaryHex(mark.grid, mark.region, layout) {
			return nil, nil, dataErrorf("",
				"hex at (row=%d, col=%d) in region %q is marked for warp but is not on a region boundary",
				mark.grid.Row, mark.grid.Col, mark.region)
		}
	}

	occupied := state.occupiedHexes()
	connected := make([]bool, len(marks))
	var warps []derivedWarp

	for i := 0; i < len(marks); i++ {
		for j := i + 1; j < len(marks); j++ {
			from, to := marks[i], marks[j]
			if from.region == to.region {
				continue
			}
			if !lineClear(from.coord, to.coord, occupied) {
				continue
			}
			connected[i] = true
			connected[j] = true
			if hexDistance(from.coord, to.coord) == 1 {
				continue // adjacent: an ordinary hex step, no warp edge needed
			}
			fromRegion, fromCoord, toRegion, toCoord := canonicalEdge(from.region, from.coord, to.region, to.coord)
			warps = append(warps, derivedWarp{
				FromRegion: fromRegion, From: fromCoord, ToRegion: toRegion, To: toCoord,
			})
		}
	}

	sort.Slice(warps, func(i, j int) bool { return warpLess(warps[i], warps[j]) })

	var warnings []string
	for i, mark := range marks {
		if !connected[i] {
			warnings = append(warnings, fmt.Sprintf(
				"hex %s(%d,%d) is marked for warp but reaches no other region",
				mark.region, mark.coord.Q, mark.coord.R))
		}
	}
	return warps, warnings, nil
}

func (state *derivation) collectMarks(layout *layoutFile) []markedHex {
	var marks []markedHex
	for row, rowCells := range layout.cells {
		for col, c := range rowCells {
			if !c.Warp {
				continue
			}
			grid := gridCoord{Row: row, Col: col}
			marks = append(marks, markedHex{
				region: state.cellRegion[grid],
				coord:  axialOf(grid),
				grid:   grid,
			})
		}
	}
	sort.Slice(marks, func(i, j int) bool {
		if marks[i].region != marks[j].region {
			return marks[i].region < marks[j].region
		}
		if marks[i].coord.Q != marks[j].coord.Q {
			return marks[i].coord.Q < marks[j].coord.Q
		}
		return marks[i].coord.R < marks[j].coord.R
	})
	return marks
}

func (state *derivation) isBoundaryHex(grid gridCoord, region string, layout *layoutFile) bool {
	for _, neighbor := range offsetNeighbors(grid) {
		if neighbor.Row < 0 || neighbor.Row >= len(layout.cells) {
			return true
		}
		if neighbor.Col < 0 || neighbor.Col >= len(layout.cells[neighbor.Row]) {
			return true
		}
		if state.cellRegion[neighbor] != region {
			return true
		}
	}
	return false
}

func (state *derivation) occupiedHexes() map[spatial.HexCoord]bool {
	occupied := map[spatial.HexCoord]bool{}
	for _, acc := range state.regions {
		for hex := range acc.hexes {
			occupied[hex] = true
		}
	}
	return occupied
}

func lineClear(from, to spatial.HexCoord, occupied map[spatial.HexCoord]bool) bool {
	line := hexLine(from, to)
	for _, hex := range line[1 : len(line)-1] {
		if occupied[hex] {
			return false
		}
	}
	return true
}
```

**2.3d.** Extend `assemble` to take warnings. **Find:**
```go
func (state *derivation) assemble(worlds []derivedWorld, warps []derivedWarp) *derivedMap {
	out := &derivedMap{Warps: warps, WarpCount: len(warps)}
```
**Replace with:**
```go
func (state *derivation) assemble(worlds []derivedWorld, warps []derivedWarp, warnings []string) *derivedMap {
	out := &derivedMap{Warps: warps, WarpCount: len(warps), Warnings: warnings}
```

**2.3e.** Delete the now-unused `edgeKey` type (nothing references it after
`applyWarps` is gone; `canonicalEdge` stays).

#### Task 2.4 — `internal/spatial/builder/data.go`

Remove the `Warps` field. **Find:**
```go
type dataFile struct {
	Regions map[string]regionEntry `toml:"regions"`
	Worlds  []worldEntry           `toml:"worlds"`
	Warps   []warpEntry            `toml:"warps"`
}
```
**Replace with:**
```go
type dataFile struct {
	Regions map[string]regionEntry `toml:"regions"`
	Worlds  []worldEntry           `toml:"worlds"`
}
```

Delete the now-unused `warpEntry` and `hexAddr` types. **Find:**
```go
type warpEntry struct {
	From hexAddr `toml:"from"`
	To   hexAddr `toml:"to"`
}

type hexAddr struct {
	Region string `toml:"region"`
	Row    int    `toml:"row"`
	Col    int    `toml:"col"`
}

```
**Replace with:** (nothing — remove it)

#### Task 2.5 — `internal/spatial/builder/builder.go`

Add `Warnings` to `Summary` and copy it in `summarize`. **Find:**
```go
type Summary struct {
	Regions    []string
	WorldCount int
	WarpCount  int
}
```
**Replace with:**
```go
type Summary struct {
	Regions    []string
	WorldCount int
	WarpCount  int
	Warnings   []string
}
```
**Find:**
```go
	s := Summary{
		WorldCount: len(derived.Worlds),
		WarpCount:  derived.WarpCount,
	}
```
**Replace with:**
```go
	s := Summary{
		WorldCount: len(derived.Worlds),
		WarpCount:  derived.WarpCount,
		Warnings:   derived.Warnings,
	}
```

#### Task 2.6 — `cmd/gm-toolkit/spatial/build.go`

Print warnings to stderr after the `Wrote …/worlds.toml` line:
```go
	for _, warning := range s.Warnings {
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %s\n", warning)
	}
```

#### Task 2.7 — Migrate `helions_reach` fixture to marks

`source/layout.txt` — add `A*` (row1/col3), `D*` (row2/col12), `B*` (row3/col6),
`C*` (row6/col9). `source/data.toml` — delete the `[[warps]]` block. `golden/
regions.toml` — regenerate; the single explicit warp is replaced by the four
derived warps (sorted by `warpLess`):

| from_region | from_q,from_r | to_region | to_q,to_r |
|-------------|---------------|-----------|-----------|
| A | 3,1 | D | 11,2 |
| B | 6,3 | C | 6,6 |
| B | 6,3 | D | 11,2 |
| C | 6,6 | D | 11,2 |

#### Task 2.8 — Replace `warp_wrong_region` with `interior_mark`

The `warp_wrong_region` fixture tested `validateWarpEndpoint`, which is gone.
Remove it and add `interior_mark` under
`internal/spatial/builder/testdata/errors/`:

`source/layout.txt` (a marked hex whose six neighbors are all same-region):
```
. . . . .
 . A A A .
. A A* A .
 . A A A .
. . . . .
```
`source/data.toml`:
```
[regions.A]
name = "Region A"
```
`err.txt`:
```
is not on a region boundary
```
Verify the marked cell is fully interior under odd-r geometry; widen the `A` blob
if not. The assertion is the err.txt substring. If the error-fixture test
discovers directories by name, drop `warp_wrong_region` and add the new fixture.

#### Task 2.9 — Update unit tests for derivation

`derive_test.go` — drop `mkData`'s `warps` parameter and `Warps:` field; add `Warp`
setting to `mkLayout`; add cases: the reference example (four `derivedWarp`s + two
exclusions), interior-mark error, dead-mark warning, adjacent-marks (distance 1 →
no warp, no warning). `layout_test.go` — `A*`/`2*` set `Warp`; stray `*` errors.

(The Commit 1 `region_map_test.go` and `emit_test.go` are unaffected — the runtime
and file format do not change in Commit 2.)

#### Task 2.10 — Migrate `test-camp` to marks

`source/layout.txt` — apply the four marks. `source/data.toml` — delete the two
`[[warps]]` blocks. Regenerate `regions.toml`/`worlds.toml` via
`spatial build --campaign test-camp --replace`.

##### Commit message
```
feat(spatial): derive warps from line-of-sight between marked hexes

- mark warp hexes with a `*` suffix in layout.txt (region or world glyph)
- derive warps between different-region marks with clear cube-linedraw LoS,
  blocked only by occupied region hexes (endpoints excepted), skipping
  hex-adjacent pairs
- require marked hexes to lie on a region boundary; warn (non-fatal, exit 0)
  on a mark that reaches no other region
- remove the [[warps]] table and from/to endpoint validation
- migrate helions_reach fixture and test-camp campaign to marked layouts
```

---

## Pre-Merge Checklist (reminder)

- Senior-engineer code review (Opus).
- Update `docs/dev_journals/faction-manager/decisions-log.md` with Decisions 1–10
  (supersedes the warp/adjacency cost-unification in log entries 270–271 and the
  boundary derivation in 272: adjacency is no longer stored).
- Update `docs/dev_journals/faction-manager/planned-work.md` (remove the warp
  redesign; note the load-time index and disjoint-region issues are resolved here).
- Version bump assessment after merge (`feat` + bug fix → likely minor).

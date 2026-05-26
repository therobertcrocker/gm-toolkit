package builder

import (
	"fmt"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

// Offset-coord neighbor deltas, format {colDelta, rowDelta}. Pointy-top hexes, odd-r staggered.
var (
	evenRowOffsets = [6][2]int{{-1, 0}, {1, 0}, {-1, -1}, {0, -1}, {-1, 1}, {0, 1}}
	oddRowOffsets  = [6][2]int{{-1, 0}, {1, 0}, {0, -1}, {1, -1}, {0, 1}, {1, 1}}
)

type gridCoord struct {
	Row, Col int
}

func axialOf(coord gridCoord) spatial.HexCoord {
	return spatial.HexCoord{Q: coord.Col - (coord.Row-(coord.Row&1))/2, R: coord.Row}
}

func offsetNeighbors(coord gridCoord) []gridCoord {
	deltas := evenRowOffsets
	if coord.Row&1 == 1 {
		deltas = oddRowOffsets
	}
	out := make([]gridCoord, 0, 6)
	for _, d := range deltas {
		out = append(out, gridCoord{Row: coord.Row + d[1], Col: coord.Col + d[0]})
	}
	return out
}

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

type derivedWorld struct {
	ID         string
	Name       string
	TechLevel  int
	Population int
	Region     string
	Coord      spatial.HexCoord
}

type regionAcc struct {
	name  string
	hexes map[spatial.HexCoord]bool
}

type edgeKey struct {
	fromRegion string
	from       spatial.HexCoord
	toRegion   string
	to         spatial.HexCoord
}

// canonicalEdge orders an inter-region edge so the alphabetically-first region
// owns it. Adjacency derivation and warp processing both depend on this rule
// agreeing across phases — keep them sharing this single helper.
func canonicalEdge(aReg string, a spatial.HexCoord, bReg string, b spatial.HexCoord) (ownerReg string, owner spatial.HexCoord, neighborReg string, neighbor spatial.HexCoord) {
	if bReg < aReg {
		return bReg, b, aReg, a
	}
	return aReg, a, bReg, b
}

// derivation holds the in-flight maps that the phase methods read and mutate.
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

func derive(layout *layoutFile, data *dataFile) (*derivedMap, error) {
	if err := crossCheckRegions(layout, data); err != nil {
		return nil, err
	}
	if err := crossCheckWorlds(layout, data); err != nil {
		return nil, err
	}

	state := newDerivation(data)
	state.assignRegionHexes(layout)

	worlds, err := state.placeWorlds(layout, data)
	if err != nil {
		return nil, err
	}

	state.deriveBoundaries(layout)

	if err := state.applyWarps(data.Warps, layout); err != nil {
		return nil, err
	}

	return state.assemble(worlds, len(data.Warps)), nil
}

func crossCheckRegions(layout *layoutFile, data *dataFile) error {
	for g := range layout.regionLetters {
		if _, ok := data.Regions[string(g)]; !ok {
			return dataErrorf(fmt.Sprintf("[regions.%s]", string(g)), "missing entry; region %q is used in layout", string(g))
		}
	}
	for id := range data.Regions {
		if len(id) != 1 || !glyph(id[0]).isRegion() {
			return dataErrorf(fmt.Sprintf("[regions.%s]", id), "region id must be a single uppercase letter 'A'-'Z'")
		}
		if !layout.regionLetters[glyph(id[0])] {
			return dataErrorf(fmt.Sprintf("[regions.%s]", id), "region %q has no cells in layout", id)
		}
	}
	return nil
}

func crossCheckWorlds(layout *layoutFile, data *dataFile) error {
	worldsByGlyph := make(map[glyph]*worldEntry, len(data.Worlds))
	for i := range data.Worlds {
		world := &data.Worlds[i]
		atG := glyph(world.At)
		if _, dup := worldsByGlyph[atG]; dup {
			return dataErrorf(fmt.Sprintf("[[worlds]] id=%q", world.ID), "duplicate `at` glyph %q", string(rune(atG)))
		}
		worldsByGlyph[atG] = world
		if _, ok := layout.markers[atG]; !ok {
			return dataErrorf(fmt.Sprintf("[[worlds]] id=%q", world.ID), "references marker %q but no such marker in layout.txt", string(rune(atG)))
		}
	}
	for g, c := range layout.markers {
		if _, ok := worldsByGlyph[g]; !ok {
			return layoutErrorf(c.FileLine, c.FileCol, "world marker %q has no [[worlds]] entry in data.toml", string(rune(g)))
		}
	}
	return nil
}

func (state *derivation) assignRegionHexes(layout *layoutFile) {
	for row, rowCells := range layout.cells {
		for col, c := range rowCells {
			if !c.Glyph.isRegion() {
				continue
			}
			coord := gridCoord{Row: row, Col: col}
			state.regions[string(c.Glyph)].hexes[axialOf(coord)] = true
			state.cellRegion[coord] = string(c.Glyph)
		}
	}
}

func (state *derivation) placeWorlds(layout *layoutFile, data *dataFile) ([]derivedWorld, error) {
	out := make([]derivedWorld, 0, len(data.Worlds))
	for _, world := range data.Worlds {
		marker := layout.markers[glyph(world.At)]
		regionID, err := state.inferWorldRegion(layout, world, marker)
		if err != nil {
			return nil, err
		}
		coord := gridCoord{Row: marker.Row, Col: marker.Col}
		axial := axialOf(coord)
		state.regions[regionID].hexes[axial] = true
		state.cellRegion[coord] = regionID
		out = append(out, derivedWorld{
			ID: world.ID, Name: world.Name, TechLevel: world.TechLevel, Population: world.Population,
			Region: regionID, Coord: axial,
		})
	}
	return out, nil
}

func (state *derivation) inferWorldRegion(layout *layoutFile, world worldEntry, marker cell) (string, error) {
	if world.Region != "" {
		if _, ok := state.regions[world.Region]; !ok {
			return "", dataErrorf(fmt.Sprintf("[[worlds]] id=%q", world.ID), "region override %q is not a declared region", world.Region)
		}
		return world.Region, nil
	}
	candidates := map[string]bool{}
	for _, neighbor := range offsetNeighbors(gridCoord{Row: marker.Row, Col: marker.Col}) {
		if neighbor.Row < 0 || neighbor.Row >= len(layout.cells) {
			continue
		}
		if neighbor.Col < 0 || neighbor.Col >= len(layout.cells[neighbor.Row]) {
			continue
		}
		if nc := layout.cells[neighbor.Row][neighbor.Col]; nc.Glyph.isRegion() {
			candidates[string(nc.Glyph)] = true
		}
	}
	switch len(candidates) {
	case 0:
		return "", dataErrorf(fmt.Sprintf("[[worlds]] id=%q", world.ID),
			"marker %q at (row=%d, col=%d) has no region-letter neighbors; add `region = \"X\"` to its entry",
			string(rune(world.At)), marker.Row, marker.Col)
	case 1:
		for id := range candidates {
			return id, nil
		}
	}
	return "", dataErrorf(fmt.Sprintf("[[worlds]] id=%q", world.ID),
		"marker %q at (row=%d, col=%d) has multiple region candidates %v; add `region = \"X\"` to disambiguate",
		string(rune(world.At)), marker.Row, marker.Col, sortedKeys(candidates))
}

func (state *derivation) deriveBoundaries(layout *layoutFile) {
	for coord, regionID := range state.cellRegion {
		for _, neighbor := range offsetNeighbors(coord) {
			if neighbor.Row < 0 || neighbor.Row >= len(layout.cells) {
				continue
			}
			if neighbor.Col < 0 || neighbor.Col >= len(layout.cells[neighbor.Row]) {
				continue
			}
			otherID, ok := state.cellRegion[neighbor]
			if !ok || otherID == regionID {
				continue
			}
			ownerID, ownerCoord, neighborID, neighborCoord := canonicalEdge(regionID, axialOf(coord), otherID, axialOf(neighbor))
			key := edgeKey{fromRegion: ownerID, from: ownerCoord, toRegion: neighborID, to: neighborCoord}
			if state.seen[key] {
				continue
			}
			state.seen[key] = true
			state.boundariesByRegion[ownerID] = append(state.boundariesByRegion[ownerID], derivedBoundary{
				From: ownerCoord, ToRegion: neighborID, To: neighborCoord,
			})
			state.adjacencyCount++
		}
	}
}

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

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

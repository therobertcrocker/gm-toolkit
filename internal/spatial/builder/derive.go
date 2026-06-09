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
	Regions   []derivedRegion
	Worlds    []derivedWorld
	Warps     []derivedWarp
	WarpCount int
	Warnings  []string
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

	warps, warnings, err := state.deriveWarps(layout)
	if err != nil {
		return nil, err
	}

	return state.assemble(worlds, warps, warnings), nil
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

func (state *derivation) assemble(worlds []derivedWorld, warps []derivedWarp, warnings []string) *derivedMap {
	out := &derivedMap{Warps: warps, WarpCount: len(warps), Warnings: warnings}
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

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

package builder

import (
	"fmt"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

// Offset-coord neighbor deltas, format {colDelta, rowDelta}. Pointy-top hexes, odd-r staggered.
var evenRowOffsets = [6][2]int{{-1, 0}, {1, 0}, {-1, -1}, {0, -1}, {-1, 1}, {0, 1}}
var oddRowOffsets = [6][2]int{{-1, 0}, {1, 0}, {0, -1}, {1, -1}, {0, 1}, {1, 1}}

func axialOf(row, col int) spatial.HexCoord {
	return spatial.HexCoord{Q: col - (row-(row&1))/2, R: row}
}

func offsetNeighbors(row, col int) [][2]int {
	deltas := evenRowOffsets
	if row&1 == 1 {
		deltas = oddRowOffsets
	}
	out := make([][2]int, 0, 6)
	for _, d := range deltas {
		out = append(out, [2]int{row + d[1], col + d[0]})
	}
	return out
}

type derivedMap struct {
	Regions        []derivedRegion
	Worlds         []derivedWorld
	AdjacencyCount int
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

// regionAcc is the in-flight region accumulator used during derivation.
type regionAcc struct {
	name  string
	hexes map[spatial.HexCoord]bool
}

func derive(layout *layoutFile, data *dataFile) (*derivedMap, error) {
	// 1. Cross-check regions: every layout letter has an entry; no extras; ids are 'A'-'Z'.
	for g := range layout.regionLetters {
		if _, ok := data.Regions[string(g)]; !ok {
			return nil, dataErrorf(fmt.Sprintf("[regions.%s]", string(g)), "missing entry; region %q is used in layout", string(g))
		}
	}
	for id := range data.Regions {
		if len(id) != 1 || !glyph(id[0]).isRegion() {
			return nil, dataErrorf(fmt.Sprintf("[regions.%s]", id), "region id must be a single uppercase letter 'A'-'Z'")
		}
		if !layout.regionLetters[glyph(id[0])] {
			return nil, dataErrorf(fmt.Sprintf("[regions.%s]", id), "region %q has no cells in layout", id)
		}
	}

	// 2. Cross-check worlds: bijection between layout markers and data.Worlds[*].At.Glyph.
	worldsByGlyph := make(map[glyph]*worldEntry, len(data.Worlds))
	for i := range data.Worlds {
		w := &data.Worlds[i]
		if _, dup := worldsByGlyph[w.At.Glyph]; dup {
			return nil, dataErrorf(fmt.Sprintf("[[worlds]] id=%q", w.ID), "duplicate `at` glyph %q", string(rune(w.At.Glyph)))
		}
		worldsByGlyph[w.At.Glyph] = w
		if _, ok := layout.markers[w.At.Glyph]; !ok {
			return nil, dataErrorf(fmt.Sprintf("[[worlds]] id=%q", w.ID), "references marker %q but no such marker in layout.txt", string(rune(w.At.Glyph)))
		}
	}
	for g, c := range layout.markers {
		if _, ok := worldsByGlyph[g]; !ok {
			return nil, layoutErrorf(c.FileLine, c.FileCol, "world marker %q has no [[worlds]] entry in data.toml", string(rune(g)))
		}
	}

	// 3. Build region hex sets from region-letter cells.
	regions := make(map[string]*regionAcc, len(data.Regions))
	for id, entry := range data.Regions {
		regions[id] = &regionAcc{name: entry.Name, hexes: make(map[spatial.HexCoord]bool)}
	}
	cellRegion := make(map[[2]int]string) // (row,col) -> region id
	for row, rowCells := range layout.cells {
		for col, c := range rowCells {
			if c.Glyph.isRegion() {
				regions[string(c.Glyph)].hexes[axialOf(row, col)] = true
				cellRegion[[2]int{row, col}] = string(c.Glyph)
			}
		}
	}

	// 4. Place worlds: infer region from non-empty region-letter neighbors, or honor explicit override.
	derivedWorlds := make([]derivedWorld, 0, len(data.Worlds))
	for _, w := range data.Worlds {
		c := layout.markers[w.At.Glyph]
		candidates := map[string]bool{}
		for _, n := range offsetNeighbors(c.Row, c.Col) {
			nrow, ncol := n[0], n[1]
			if nrow < 0 || nrow >= len(layout.cells) {
				continue
			}
			if ncol < 0 || ncol >= len(layout.cells[nrow]) {
				continue
			}
			nc := layout.cells[nrow][ncol]
			if nc.Glyph.isRegion() {
				candidates[string(nc.Glyph)] = true
			}
		}

		var regionID string
		switch {
		case w.Region != "":
			if _, ok := regions[w.Region]; !ok {
				return nil, dataErrorf(fmt.Sprintf("[[worlds]] id=%q", w.ID), "region override %q is not a declared region", w.Region)
			}
			regionID = w.Region
		case len(candidates) == 1:
			for id := range candidates {
				regionID = id
			}
		case len(candidates) == 0:
			return nil, dataErrorf(fmt.Sprintf("[[worlds]] id=%q", w.ID),
				"marker %q at (row=%d, col=%d) has no region-letter neighbors; add `region = \"X\"` to its entry",
				string(rune(w.At.Glyph)), c.Row, c.Col)
		default:
			return nil, dataErrorf(fmt.Sprintf("[[worlds]] id=%q", w.ID),
				"marker %q at (row=%d, col=%d) has multiple region candidates %v; add `region = \"X\"` to disambiguate",
				string(rune(w.At.Glyph)), c.Row, c.Col, sortedKeys(candidates))
		}

		coord := axialOf(c.Row, c.Col)
		regions[regionID].hexes[coord] = true
		cellRegion[[2]int{c.Row, c.Col}] = regionID
		derivedWorlds = append(derivedWorlds, derivedWorld{
			ID: w.ID, Name: w.Name, TechLevel: w.TechLevel, Population: w.Population,
			Region: regionID, Coord: coord,
		})
	}

	// 5. Derive adjacency boundaries, canonicalized to the alphabetically-first region's side.
	type edgeKey struct {
		fromRegion string
		from       spatial.HexCoord
		toRegion   string
		to         spatial.HexCoord
	}
	seen := map[edgeKey]bool{}
	boundariesByRegion := map[string][]derivedBoundary{}
	for ck, regionID := range cellRegion {
		row, col := ck[0], ck[1]
		for _, n := range offsetNeighbors(row, col) {
			nrow, ncol := n[0], n[1]
			if nrow < 0 || nrow >= len(layout.cells) {
				continue
			}
			if ncol < 0 || ncol >= len(layout.cells[nrow]) {
				continue
			}
			otherID, ok := cellRegion[[2]int{nrow, ncol}]
			if !ok || otherID == regionID {
				continue
			}
			ownerID, ownerCoord, neighborID, neighborCoord := regionID, axialOf(row, col), otherID, axialOf(nrow, ncol)
			if neighborID < ownerID {
				ownerID, neighborID = neighborID, ownerID
				ownerCoord, neighborCoord = neighborCoord, ownerCoord
			}
			key := edgeKey{fromRegion: ownerID, from: ownerCoord, toRegion: neighborID, to: neighborCoord}
			if seen[key] {
				continue
			}
			seen[key] = true
			boundariesByRegion[ownerID] = append(boundariesByRegion[ownerID], derivedBoundary{
				From: ownerCoord, ToRegion: neighborID, To: neighborCoord,
			})
		}
	}
	adjacencyCount := 0
	for _, bs := range boundariesByRegion {
		adjacencyCount += len(bs)
	}

	// 6. Process warps. Validate endpoints; emit on alphabetically-first region's side.
	for i, w := range data.Warps {
		ctx := fmt.Sprintf("[[warps]] #%d", i+1)
		if err := validateWarpEndpoint(layout, regions, cellRegion, w.From, ctx, "from"); err != nil {
			return nil, err
		}
		if err := validateWarpEndpoint(layout, regions, cellRegion, w.To, ctx, "to"); err != nil {
			return nil, err
		}
		if w.From.Region == w.To.Region {
			return nil, dataErrorf(ctx, "both endpoints are in region %q; warps must cross region boundaries", w.From.Region)
		}
		ownerID, ownerCoord := w.From.Region, axialOf(w.From.Row, w.From.Col)
		neighborID, neighborCoord := w.To.Region, axialOf(w.To.Row, w.To.Col)
		if neighborID < ownerID {
			ownerID, neighborID = neighborID, ownerID
			ownerCoord, neighborCoord = neighborCoord, ownerCoord
		}
		warpKey := edgeKey{fromRegion: ownerID, from: ownerCoord, toRegion: neighborID, to: neighborCoord}
		if seen[warpKey] {
			return nil, dataErrorf(ctx, "duplicate warp: hex pair already connected by a boundary or warp")
		}
		seen[warpKey] = true
		boundariesByRegion[ownerID] = append(boundariesByRegion[ownerID], derivedBoundary{
			From: ownerCoord, ToRegion: neighborID, To: neighborCoord,
		})
	}

	// 7. Assemble and sort.
	out := &derivedMap{AdjacencyCount: adjacencyCount}
	for _, id := range sortedKeys(mapKeysAsBoolSet(regions)) {
		acc := regions[id]
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
		bs := boundariesByRegion[id]
		sort.Slice(bs, func(i, j int) bool {
			switch {
			case bs[i].ToRegion != bs[j].ToRegion:
				return bs[i].ToRegion < bs[j].ToRegion
			case bs[i].From.Q != bs[j].From.Q:
				return bs[i].From.Q < bs[j].From.Q
			case bs[i].From.R != bs[j].From.R:
				return bs[i].From.R < bs[j].From.R
			case bs[i].To.Q != bs[j].To.Q:
				return bs[i].To.Q < bs[j].To.Q
			default:
				return bs[i].To.R < bs[j].To.R
			}
		})
		out.Regions = append(out.Regions, derivedRegion{
			ID: id, Name: acc.name, Hexes: hexes, Boundaries: bs,
		})
	}
	sort.Slice(derivedWorlds, func(i, j int) bool { return derivedWorlds[i].ID < derivedWorlds[j].ID })
	out.Worlds = derivedWorlds
	return out, nil
}

func validateWarpEndpoint(layout *layoutFile, regions map[string]*regionAcc, cellRegion map[[2]int]string, addr hexAddr, ctx, side string) error {
	if addr.Row < 0 || addr.Row >= len(layout.cells) {
		return dataErrorf(ctx, "%s.row=%d is out of bounds (grid has %d rows)", side, addr.Row, len(layout.cells))
	}
	if addr.Col < 0 || addr.Col >= len(layout.cells[addr.Row]) {
		return dataErrorf(ctx, "%s.col=%d is out of bounds for row %d (has %d cells)", side, addr.Col, addr.Row, len(layout.cells[addr.Row]))
	}
	if _, ok := regions[addr.Region]; !ok {
		return dataErrorf(ctx, "%s.region=%q is not a declared region", side, addr.Region)
	}
	actual, ok := cellRegion[[2]int{addr.Row, addr.Col}]
	if !ok {
		return dataErrorf(ctx, "%s cell at (row=%d, col=%d) is empty, not a region member", side, addr.Row, addr.Col)
	}
	if actual != addr.Region {
		return dataErrorf(ctx, "%s cell at (row=%d, col=%d) belongs to region %q, not %q", side, addr.Row, addr.Col, actual, addr.Region)
	}
	return nil
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// mapKeysAsBoolSet converts map[string]*regionAcc to map[string]bool for sortedKeys reuse.
func mapKeysAsBoolSet(m map[string]*regionAcc) map[string]bool {
	out := make(map[string]bool, len(m))
	for k := range m {
		out[k] = true
	}
	return out
}

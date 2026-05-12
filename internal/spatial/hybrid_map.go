package spatial

import (
	"container/heap"
	"fmt"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

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

type World struct {
	id         string
	name       string
	techLevel  int
	population int
	Region     string
	Hex        HexCoord
}

func (w *World) ID() string         { return w.id }
func (w *World) Name() string       { return w.name }
func (w *World) TechLevel() int     { return w.techLevel }
func (w *World) Population() int    { return w.population }
func (w *World) Coords() (q, r int) { return w.Hex.Q, w.Hex.R }
func (w *World) RegionID() string   { return w.Region }

type HybridMap struct {
	regions map[string]*Region
	worlds  map[string]*World
}

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

type tomlWorld struct {
	ID         string `toml:"id"`
	Name       string `toml:"name"`
	TechLevel  int    `toml:"tech_level"`
	Population int    `toml:"population"`
	Region     string `toml:"region"`
	HexQ       int    `toml:"hex_q"`
	HexR       int    `toml:"hex_r"`
}

type regionsFile struct {
	Region []tomlRegion `toml:"region"`
}

type worldsFile struct {
	World []tomlWorld `toml:"world"`
}

func LoadHybrid(dataDir string) (*HybridMap, error) {
	var regionsDoc regionsFile
	if _, err := toml.DecodeFile(filepath.Join(dataDir, "regions.toml"), &regionsDoc); err != nil {
		return nil, fmt.Errorf("loading regions.toml: %w", err)
	}

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

	var worldsDoc worldsFile
	if _, err := toml.DecodeFile(filepath.Join(dataDir, "worlds.toml"), &worldsDoc); err != nil {
		return nil, fmt.Errorf("loading worlds.toml: %w", err)
	}

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
			Region:     world.Region,
			Hex:        hex,
		}
	}

	return &HybridMap{regions: regions, worlds: worlds}, nil
}

func (hybridMap *HybridMap) Location(id string) (Location, bool) {
	world, ok := hybridMap.worlds[id]
	if !ok {
		return nil, false
	}
	return world, true
}

type hexNode struct {
	regionID string
	coord    HexCoord
}

type edge struct {
	to   hexNode
	cost int
}

var hexNeighbors = [6]HexCoord{
	{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, -1}, {-1, 1},
}

type pqItem struct {
	node hexNode
	cost int
}

type priorityQueue []*pqItem

func (priorityQ priorityQueue) Len() int           { return len(priorityQ) }
func (priorityQ priorityQueue) Less(i, j int) bool { return priorityQ[i].cost < priorityQ[j].cost }
func (priorityQ priorityQueue) Swap(i, j int) {
	priorityQ[i], priorityQ[j] = priorityQ[j], priorityQ[i]
}

func (priorityQ *priorityQueue) Push(item any) {
	*priorityQ = append(*priorityQ, item.(*pqItem))
}

func (priorityQ *priorityQueue) Pop() any {
	old := *priorityQ
	n := len(old)
	entry := old[n-1]
	old[n-1] = nil
	*priorityQ = old[:n-1]
	return entry
}

func (hybridMap *HybridMap) neighbors(node hexNode, crossingCost int) []edge {
	var edges []edge

	// (1) Look up the Region this node lives in.
	//     region := hybridMap.regions[node.regionID]

	region := hybridMap.regions[node.regionID]

	// (2) Intra-region: for each of the 6 axial deltas in hexNeighbors,
	//     compute candidate := HexCoord{node.coord.Q + delta.Q, node.coord.R + delta.R}.
	//     If region.Hexes[candidate] is true, append:
	//         edge{to: hexNode{node.regionID, candidate}, cost: 1}

	for _, delta := range hexNeighbors {
		candidate := HexCoord{Q: node.coord.Q + delta.Q, R: node.coord.R + delta.R}
		if region.Hexes[candidate] {
			edges = append(edges, edge{to: hexNode{regionID: node.regionID, coord: candidate}, cost: 1})
		}
	}

	// (3) Outgoing boundaries: walk region.Boundaries.
	//     For each boundary whose .From == node.coord, append:
	//         edge{to: hexNode{boundary.ToRegion, boundary.To}, cost: crossingCost}

	for _, boundary := range region.Boundaries {
		if boundary.From == node.coord {
			edges = append(edges, edge{to: hexNode{regionID: boundary.ToRegion, coord: boundary.To}, cost: crossingCost})
		}
	}

	// (4) Incoming boundaries (the "bidirectional" piece): walk every OTHER region in
	//     hybridMap.regions, then walk that region's Boundaries.
	//     If a boundary has .ToRegion == node.regionID AND .To == node.coord,
	//     it's a forward edge from THAT region pointing at us — so the reverse
	//     edge goes from us back to its source. Append:
	//         edge{to: hexNode{otherRegion.ID, boundary.From}, cost: crossingCost}

	for _, otherRegion := range hybridMap.regions {
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

func (hybridMap *HybridMap) Distance(fromID, toID string, crossingCost int) (int, error) {
	if crossingCost < 0 {
		return 0, fmt.Errorf("%w: crossingCost=%d (must be non-negative)", ErrInvalidCost, crossingCost)
	}
	fromWorld, ok := hybridMap.worlds[fromID]
	if !ok {
		return 0, fmt.Errorf("%w: %q", ErrUnknownWorld, fromID)
	}
	toWorld, ok := hybridMap.worlds[toID]
	if !ok {
		return 0, fmt.Errorf("%w: %q", ErrUnknownWorld, toID)
	}
	if fromID == toID {
		return 0, nil
	}

	source := hexNode{regionID: fromWorld.Region, coord: fromWorld.Hex}
	target := hexNode{regionID: toWorld.Region, coord: toWorld.Hex}

	dist := map[hexNode]int{source: 0}
	queue := &priorityQueue{}
	heap.Init(queue)
	heap.Push(queue, &pqItem{node: source, cost: 0})

	for queue.Len() > 0 {
		current := heap.Pop(queue).(*pqItem)
		if current.cost > dist[current.node] {
			continue
		}
		if current.node == target {
			return current.cost, nil
		}
		for _, next := range hybridMap.neighbors(current.node, crossingCost) {
			candidate := current.cost + next.cost
			if known, seen := dist[next.to]; !seen || candidate < known {
				dist[next.to] = candidate
				heap.Push(queue, &pqItem{node: next.to, cost: candidate})
			}
		}
	}

	return 0, fmt.Errorf("%w: from %q to %q", ErrNoPath, fromID, toID)
}

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
	loc        RegionHex
}

func (w *World) ID() string           { return w.id }
func (w *World) Name() string         { return w.name }
func (w *World) TechLevel() int       { return w.techLevel }
func (w *World) Population() int      { return w.population }
func (w *World) Coords() (q, r int)   { return w.loc.Coord.Q, w.loc.Coord.R }
func (w *World) RegionID() string     { return w.loc.RegionID }
func (w *World) RegionHex() RegionHex { return w.loc }

type RegionHex struct {
	RegionID string
	Coord    HexCoord
}

type RegionMap struct {
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

func LoadRegionMap(dataDir string) (*RegionMap, error) {
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
			loc:        RegionHex{RegionID: world.Region, Coord: hex},
		}
	}

	return &RegionMap{regions: regions, worlds: worlds}, nil
}

func (regionMap *RegionMap) Location(id string) (Location, bool) {
	world, ok := regionMap.worlds[id]
	if !ok {
		return nil, false
	}
	return world, true
}

func (regionMap *RegionMap) RegionOfHex(hex HexCoord) (string, bool) {
	for _, region := range regionMap.regions {
		if region.Hexes[hex] {
			return region.ID, true
		}
	}
	return "", false
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

func (regionMap *RegionMap) Distance(from, to RegionHex, crossingCost int) (int, error) {
	if crossingCost < 0 {
		return 0, fmt.Errorf("%w: crossingCost=%d (must be non-negative)", ErrInvalidCost, crossingCost)
	}
	if from == to {
		return 0, nil
	}

	source := hexNode{regionID: from.RegionID, coord: from.Coord}
	target := hexNode{regionID: to.RegionID, coord: to.Coord}

	dist, _ := dijkstra(regionMap, source, target, crossingCost)
	cost, found := dist[target]
	if !found {
		return 0, fmt.Errorf("%w: from %s(%d,%d) to %s(%d,%d)", ErrNoPath, from.RegionID, from.Coord.Q, from.Coord.R, to.RegionID, to.Coord.Q, to.Coord.R)
	}
	return cost, nil
}

func (regionMap *RegionMap) Path(from, to RegionHex, crossingCost int) ([]RegionHex, int, error) {
	if crossingCost < 0 {
		return nil, 0, fmt.Errorf("%w: crossingCost=%d (must be non-negative)", ErrInvalidCost, crossingCost)
	}
	if from == to {
		return []RegionHex{from}, 0, nil
	}

	source := hexNode{regionID: from.RegionID, coord: from.Coord}
	target := hexNode{regionID: to.RegionID, coord: to.Coord}

	return regionMap.pathBetween(source, target, crossingCost)
}

func (regionMap *RegionMap) pathBetween(source, target hexNode, crossingCost int) ([]RegionHex, int, error) {
	dist, prev := dijkstra(regionMap, source, target, crossingCost)
	cost, found := dist[target]
	if !found {
		return nil, 0, fmt.Errorf("%w: from %s(%d,%d) to %s(%d,%d)", ErrNoPath, source.regionID, source.coord.Q, source.coord.R, target.regionID, target.coord.Q, target.coord.R)
	}

	var path []RegionHex
	for at := target; at != source; at = prev[at] {
		path = append(path, RegionHex{RegionID: at.regionID, Coord: at.coord})
	}
	path = append(path, RegionHex{RegionID: source.regionID, Coord: source.coord})

	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path, cost, nil
}

func dijkstra(regionMap *RegionMap, source hexNode, target hexNode, crossingCost int) (map[hexNode]int, map[hexNode]hexNode) {
	dist := map[hexNode]int{source: 0}
	prev := map[hexNode]hexNode{}
	queue := &priorityQueue{}
	heap.Init(queue)
	heap.Push(queue, &pqItem{node: source, cost: 0})

	for queue.Len() > 0 {
		current := heap.Pop(queue).(*pqItem)
		if current.cost > dist[current.node] {
			continue
		}
		if current.node == target {
			return dist, prev
		}
		for _, next := range regionMap.neighbors(current.node, crossingCost) {
			candidate := current.cost + next.cost
			if known, seen := dist[next.to]; !seen || candidate < known {
				dist[next.to] = candidate
				prev[next.to] = current.node
				heap.Push(queue, &pqItem{node: next.to, cost: candidate})
			}
		}
	}
	return dist, prev
}

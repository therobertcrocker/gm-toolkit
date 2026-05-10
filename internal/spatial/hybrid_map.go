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

type Fragment struct {
	FragmentID     string
	FragmentName   string
	FragTechLevel  int
	FragPopulation int
	Region         string
	Hex            HexCoord
}

func (fragment *Fragment) ID() string      { return fragment.FragmentID }
func (fragment *Fragment) Name() string    { return fragment.FragmentName }
func (fragment *Fragment) TechLevel() int  { return fragment.FragTechLevel }
func (fragment *Fragment) Population() int { return fragment.FragPopulation }

type HybridMap struct {
	regions   map[string]*Region
	fragments map[string]*Fragment
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

type tomlFragment struct {
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

type fragmentsFile struct {
	Fragment []tomlFragment `toml:"fragment"`
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

	var fragmentsDoc fragmentsFile
	if _, err := toml.DecodeFile(filepath.Join(dataDir, "fragments.toml"), &fragmentsDoc); err != nil {
		return nil, fmt.Errorf("loading fragments.toml: %w", err)
	}

	fragments := make(map[string]*Fragment, len(fragmentsDoc.Fragment))
	for _, fragment := range fragmentsDoc.Fragment {
		region, ok := regions[fragment.Region]
		if !ok {
			return nil, fmt.Errorf("fragment %q references unknown region %q", fragment.ID, fragment.Region)
		}
		hex := HexCoord{Q: fragment.HexQ, R: fragment.HexR}
		if !region.Hexes[hex] {
			return nil, fmt.Errorf("fragment %q at hex (%d,%d) is not within region %q", fragment.ID, hex.Q, hex.R, fragment.Region)
		}
		fragments[fragment.ID] = &Fragment{
			FragmentID:     fragment.ID,
			FragmentName:   fragment.Name,
			FragTechLevel:  fragment.TechLevel,
			FragPopulation: fragment.Population,
			Region:         fragment.Region,
			Hex:            hex,
		}
	}

	return &HybridMap{regions: regions, fragments: fragments}, nil
}

func (hybridMap *HybridMap) Location(id string) (Location, bool) {
	fragment, ok := hybridMap.fragments[id]
	return fragment, ok
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
	idx  int
}

type priorityQueue []*pqItem

func (priorityQ priorityQueue) Len() int           { return len(priorityQ) }
func (priorityQ priorityQueue) Less(i, j int) bool { return priorityQ[i].cost < priorityQ[j].cost }
func (priorityQ priorityQueue) Swap(i, j int) {
	priorityQ[i], priorityQ[j] = priorityQ[j], priorityQ[i]
	priorityQ[i].idx = i
	priorityQ[j].idx = j
}

func (priorityQ *priorityQueue) Push(item any) {
	entry := item.(*pqItem)
	entry.idx = len(*priorityQ)
	*priorityQ = append(*priorityQ, entry)
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
	fromFragment, ok := hybridMap.fragments[fromID]
	if !ok {
		return 0, fmt.Errorf("unknown fragment %q", fromID)
	}
	toFragment, ok := hybridMap.fragments[toID]
	if !ok {
		return 0, fmt.Errorf("unknown fragment %q", toID)
	}
	if fromID == toID {
		return 0, nil
	}

	source := hexNode{regionID: fromFragment.Region, coord: fromFragment.Hex}
	target := hexNode{regionID: toFragment.Region, coord: toFragment.Hex}

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

	return 0, fmt.Errorf("no path from %q to %q", fromID, toID)
}

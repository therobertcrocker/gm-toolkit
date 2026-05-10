package spatial

import (
	"strings"
	"testing"
)

func newTestMap(regions map[string]*Region, fragments map[string]*Fragment) *HybridMap {
	return &HybridMap{regions: regions, fragments: fragments}
}

func makeHexes(coords ...HexCoord) map[HexCoord]bool {
	hexes := make(map[HexCoord]bool, len(coords))
	for _, coord := range coords {
		hexes[coord] = true
	}
	return hexes
}

func TestHybridMap_Distance_SameFragment(t *testing.T) {
	regions := map[string]*Region{
		"r1": {ID: "r1", Hexes: makeHexes(HexCoord{0, 0})},
	}
	fragments := map[string]*Fragment{
		"a": {FragmentID: "a", Region: "r1", Hex: HexCoord{0, 0}},
	}
	hybridMap := newTestMap(regions, fragments)

	got, err := hybridMap.Distance("a", "a", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 0 {
		t.Errorf("Distance(a,a) = %d, want 0", got)
	}
}

func TestHybridMap_Distance_AdjacentIntraRegion(t *testing.T) {
	regions := map[string]*Region{
		"r1": {ID: "r1", Hexes: makeHexes(HexCoord{0, 0}, HexCoord{1, 0})},
	}
	fragments := map[string]*Fragment{
		"a": {FragmentID: "a", Region: "r1", Hex: HexCoord{0, 0}},
		"b": {FragmentID: "b", Region: "r1", Hex: HexCoord{1, 0}},
	}
	hybridMap := newTestMap(regions, fragments)

	got, err := hybridMap.Distance("a", "b", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 1 {
		t.Errorf("Distance(a,b) = %d, want 1", got)
	}
}

func TestHybridMap_Distance_NonAdjacentIntraRegion(t *testing.T) {
	regions := map[string]*Region{
		"r1": {ID: "r1", Hexes: makeHexes(
			HexCoord{0, 0}, HexCoord{1, 0}, HexCoord{2, 0}, HexCoord{3, 0},
		)},
	}
	fragments := map[string]*Fragment{
		"a": {FragmentID: "a", Region: "r1", Hex: HexCoord{0, 0}},
		"b": {FragmentID: "b", Region: "r1", Hex: HexCoord{3, 0}},
	}
	hybridMap := newTestMap(regions, fragments)

	got, err := hybridMap.Distance("a", "b", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 3 {
		t.Errorf("Distance(a,b) = %d, want 3", got)
	}
}

func TestHybridMap_Distance_HoleRouting(t *testing.T) {
	regions := map[string]*Region{
		"r1": {ID: "r1", Hexes: makeHexes(
			HexCoord{0, 0}, HexCoord{0, 1}, HexCoord{1, 1}, HexCoord{2, 1}, HexCoord{2, 0},
		)},
	}
	fragments := map[string]*Fragment{
		"a": {FragmentID: "a", Region: "r1", Hex: HexCoord{0, 0}},
		"b": {FragmentID: "b", Region: "r1", Hex: HexCoord{2, 0}},
	}
	hybridMap := newTestMap(regions, fragments)

	got, err := hybridMap.Distance("a", "b", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 3 {
		t.Errorf("Distance(a,b) = %d, want 3 (detour around missing (1,0))", got)
	}
}

func TestHybridMap_Distance_CrossRegionSingleBoundary(t *testing.T) {
	regions := map[string]*Region{
		"a": {
			ID:    "a",
			Hexes: makeHexes(HexCoord{0, 0}, HexCoord{1, 0}),
			Boundaries: []BoundaryConnection{
				{From: HexCoord{1, 0}, ToRegion: "b", To: HexCoord{0, 0}},
			},
		},
		"b": {
			ID:    "b",
			Hexes: makeHexes(HexCoord{0, 0}, HexCoord{1, 0}),
		},
	}
	fragments := map[string]*Fragment{
		"p": {FragmentID: "p", Region: "a", Hex: HexCoord{0, 0}},
		"q": {FragmentID: "q", Region: "b", Hex: HexCoord{1, 0}},
	}
	hybridMap := newTestMap(regions, fragments)

	got, err := hybridMap.Distance("p", "q", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 7 {
		t.Errorf("Distance(p,q) = %d, want 7 (1 + 5 crossing + 1)", got)
	}
}

func TestHybridMap_Distance_CrossRegionTwoBoundaries(t *testing.T) {
	regions := map[string]*Region{
		"a": {
			ID:    "a",
			Hexes: makeHexes(HexCoord{0, 0}),
			Boundaries: []BoundaryConnection{
				{From: HexCoord{0, 0}, ToRegion: "b", To: HexCoord{0, 0}},
			},
		},
		"b": {
			ID:    "b",
			Hexes: makeHexes(HexCoord{0, 0}),
			Boundaries: []BoundaryConnection{
				{From: HexCoord{0, 0}, ToRegion: "c", To: HexCoord{0, 0}},
			},
		},
		"c": {
			ID:    "c",
			Hexes: makeHexes(HexCoord{0, 0}),
		},
	}
	fragments := map[string]*Fragment{
		"x": {FragmentID: "x", Region: "a", Hex: HexCoord{0, 0}},
		"y": {FragmentID: "y", Region: "c", Hex: HexCoord{0, 0}},
	}
	hybridMap := newTestMap(regions, fragments)

	got, err := hybridMap.Distance("x", "y", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 6 {
		t.Errorf("Distance(x,y) = %d, want 6 (3 + 3)", got)
	}
}

func TestHybridMap_Distance_UnknownFragment(t *testing.T) {
	regions := map[string]*Region{
		"r1": {ID: "r1", Hexes: makeHexes(HexCoord{0, 0})},
	}
	fragments := map[string]*Fragment{
		"a": {FragmentID: "a", Region: "r1", Hex: HexCoord{0, 0}},
	}
	hybridMap := newTestMap(regions, fragments)

	if _, err := hybridMap.Distance("nope", "a", 1); err == nil {
		t.Error("Distance with unknown source: want error, got nil")
	}
	if _, err := hybridMap.Distance("a", "nope", 1); err == nil {
		t.Error("Distance with unknown target: want error, got nil")
	}
}

func TestHybridMap_Distance_NoPath(t *testing.T) {
	regions := map[string]*Region{
		"a": {ID: "a", Hexes: makeHexes(HexCoord{0, 0})},
		"b": {ID: "b", Hexes: makeHexes(HexCoord{0, 0})},
	}
	fragments := map[string]*Fragment{
		"x": {FragmentID: "x", Region: "a", Hex: HexCoord{0, 0}},
		"y": {FragmentID: "y", Region: "b", Hex: HexCoord{0, 0}},
	}
	hybridMap := newTestMap(regions, fragments)

	_, err := hybridMap.Distance("x", "y", 1)
	if err == nil {
		t.Fatal("Distance across disconnected regions: want error, got nil")
	}
	if !strings.Contains(err.Error(), "no path") {
		t.Errorf("error = %q, want it to mention \"no path\"", err.Error())
	}
}

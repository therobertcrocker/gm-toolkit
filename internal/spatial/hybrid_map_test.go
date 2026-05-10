package spatial

import (
	"strings"
	"testing"
)

func makeHexes(coords ...HexCoord) map[HexCoord]bool {
	hexes := make(map[HexCoord]bool, len(coords))
	for _, coord := range coords {
		hexes[coord] = true
	}
	return hexes
}

func TestHybridMap_Distance(t *testing.T) {
	cases := []struct {
		name          string
		regions       map[string]*Region
		fragments     map[string]*Fragment
		fromID        string
		toID          string
		crossingCost  int
		wantDist      int
		wantErr       bool
		wantErrSubstr string
	}{
		{
			name: "same fragment returns 0",
			regions: map[string]*Region{
				"r1": {ID: "r1", Hexes: makeHexes(HexCoord{0, 0})},
			},
			fragments: map[string]*Fragment{
				"a": {FragmentID: "a", Region: "r1", Hex: HexCoord{0, 0}},
			},
			fromID: "a", toID: "a", crossingCost: 5,
			wantDist: 0,
		},
		{
			name: "adjacent intra-region",
			regions: map[string]*Region{
				"r1": {ID: "r1", Hexes: makeHexes(HexCoord{0, 0}, HexCoord{1, 0})},
			},
			fragments: map[string]*Fragment{
				"a": {FragmentID: "a", Region: "r1", Hex: HexCoord{0, 0}},
				"b": {FragmentID: "b", Region: "r1", Hex: HexCoord{1, 0}},
			},
			fromID: "a", toID: "b", crossingCost: 5,
			wantDist: 1,
		},
		{
			name: "non-adjacent intra-region",
			regions: map[string]*Region{
				"r1": {ID: "r1", Hexes: makeHexes(
					HexCoord{0, 0}, HexCoord{1, 0}, HexCoord{2, 0}, HexCoord{3, 0},
				)},
			},
			fragments: map[string]*Fragment{
				"a": {FragmentID: "a", Region: "r1", Hex: HexCoord{0, 0}},
				"b": {FragmentID: "b", Region: "r1", Hex: HexCoord{3, 0}},
			},
			fromID: "a", toID: "b", crossingCost: 5,
			wantDist: 3,
		},
		{
			name: "hole routing forces detour",
			regions: map[string]*Region{
				"r1": {ID: "r1", Hexes: makeHexes(
					HexCoord{0, 0}, HexCoord{0, 1}, HexCoord{1, 1}, HexCoord{2, 1}, HexCoord{2, 0},
				)},
			},
			fragments: map[string]*Fragment{
				"a": {FragmentID: "a", Region: "r1", Hex: HexCoord{0, 0}},
				"b": {FragmentID: "b", Region: "r1", Hex: HexCoord{2, 0}},
			},
			fromID: "a", toID: "b", crossingCost: 5,
			wantDist: 3,
		},
		{
			name: "cross-region single boundary",
			regions: map[string]*Region{
				"a": {
					ID:    "a",
					Hexes: makeHexes(HexCoord{0, 0}, HexCoord{1, 0}),
					Boundaries: []BoundaryConnection{
						{From: HexCoord{1, 0}, ToRegion: "b", To: HexCoord{0, 0}},
					},
				},
				"b": {ID: "b", Hexes: makeHexes(HexCoord{0, 0}, HexCoord{1, 0})},
			},
			fragments: map[string]*Fragment{
				"p": {FragmentID: "p", Region: "a", Hex: HexCoord{0, 0}},
				"q": {FragmentID: "q", Region: "b", Hex: HexCoord{1, 0}},
			},
			fromID: "p", toID: "q", crossingCost: 5,
			wantDist: 7,
		},
		{
			name: "cross-region two boundaries chained",
			regions: map[string]*Region{
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
				"c": {ID: "c", Hexes: makeHexes(HexCoord{0, 0})},
			},
			fragments: map[string]*Fragment{
				"x": {FragmentID: "x", Region: "a", Hex: HexCoord{0, 0}},
				"y": {FragmentID: "y", Region: "c", Hex: HexCoord{0, 0}},
			},
			fromID: "x", toID: "y", crossingCost: 3,
			wantDist: 6,
		},
		{
			name: "unknown source fragment",
			regions: map[string]*Region{
				"r1": {ID: "r1", Hexes: makeHexes(HexCoord{0, 0})},
			},
			fragments: map[string]*Fragment{
				"a": {FragmentID: "a", Region: "r1", Hex: HexCoord{0, 0}},
			},
			fromID: "nope", toID: "a", crossingCost: 1,
			wantErr: true, wantErrSubstr: "nope",
		},
		{
			name: "unknown target fragment",
			regions: map[string]*Region{
				"r1": {ID: "r1", Hexes: makeHexes(HexCoord{0, 0})},
			},
			fragments: map[string]*Fragment{
				"a": {FragmentID: "a", Region: "r1", Hex: HexCoord{0, 0}},
			},
			fromID: "a", toID: "nope", crossingCost: 1,
			wantErr: true, wantErrSubstr: "nope",
		},
		{
			name: "no path between disconnected regions",
			regions: map[string]*Region{
				"a": {ID: "a", Hexes: makeHexes(HexCoord{0, 0})},
				"b": {ID: "b", Hexes: makeHexes(HexCoord{0, 0})},
			},
			fragments: map[string]*Fragment{
				"x": {FragmentID: "x", Region: "a", Hex: HexCoord{0, 0}},
				"y": {FragmentID: "y", Region: "b", Hex: HexCoord{0, 0}},
			},
			fromID: "x", toID: "y", crossingCost: 1,
			wantErr: true, wantErrSubstr: "no path",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			hybridMap := &HybridMap{regions: testCase.regions, fragments: testCase.fragments}

			got, err := hybridMap.Distance(testCase.fromID, testCase.toID, testCase.crossingCost)

			if testCase.wantErr {
				if err == nil {
					t.Fatalf("Distance(%q,%q) = %d, want error", testCase.fromID, testCase.toID, got)
				}
				if testCase.wantErrSubstr != "" && !strings.Contains(err.Error(), testCase.wantErrSubstr) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), testCase.wantErrSubstr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != testCase.wantDist {
				t.Errorf("Distance(%q,%q) = %d, want %d", testCase.fromID, testCase.toID, got, testCase.wantDist)
			}
		})
	}
}

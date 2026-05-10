package spatial

import (
	"errors"
	"os"
	"path/filepath"
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
		name         string
		regions      map[string]*Region
		fragments    map[string]*Fragment
		fromID       string
		toID         string
		crossingCost int
		wantDist     int
		wantErr      error
	}{
		{
			name: "same fragment returns 0",
			regions: map[string]*Region{
				"r1": {ID: "r1", Hexes: makeHexes(HexCoord{0, 0})},
			},
			fragments: map[string]*Fragment{
				"a": {id: "a", Region: "r1", Hex: HexCoord{0, 0}},
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
				"a": {id: "a", Region: "r1", Hex: HexCoord{0, 0}},
				"b": {id: "b", Region: "r1", Hex: HexCoord{1, 0}},
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
				"a": {id: "a", Region: "r1", Hex: HexCoord{0, 0}},
				"b": {id: "b", Region: "r1", Hex: HexCoord{3, 0}},
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
				"a": {id: "a", Region: "r1", Hex: HexCoord{0, 0}},
				"b": {id: "b", Region: "r1", Hex: HexCoord{2, 0}},
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
				"p": {id: "p", Region: "a", Hex: HexCoord{0, 0}},
				"q": {id: "q", Region: "b", Hex: HexCoord{1, 0}},
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
				"x": {id: "x", Region: "a", Hex: HexCoord{0, 0}},
				"y": {id: "y", Region: "c", Hex: HexCoord{0, 0}},
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
				"a": {id: "a", Region: "r1", Hex: HexCoord{0, 0}},
			},
			fromID: "nope", toID: "a", crossingCost: 1,
			wantErr: ErrUnknownFragment,
		},
		{
			name: "unknown target fragment",
			regions: map[string]*Region{
				"r1": {ID: "r1", Hexes: makeHexes(HexCoord{0, 0})},
			},
			fragments: map[string]*Fragment{
				"a": {id: "a", Region: "r1", Hex: HexCoord{0, 0}},
			},
			fromID: "a", toID: "nope", crossingCost: 1,
			wantErr: ErrUnknownFragment,
		},
		{
			name: "bidirectional traversal against declared boundary direction",
			regions: map[string]*Region{
				"a": {
					ID:    "a",
					Hexes: makeHexes(HexCoord{0, 0}),
					Boundaries: []BoundaryConnection{
						{From: HexCoord{0, 0}, ToRegion: "b", To: HexCoord{0, 0}},
					},
				},
				"b": {ID: "b", Hexes: makeHexes(HexCoord{0, 0})},
			},
			fragments: map[string]*Fragment{
				"p": {id: "p", Region: "b", Hex: HexCoord{0, 0}},
				"q": {id: "q", Region: "a", Hex: HexCoord{0, 0}},
			},
			fromID: "p", toID: "q", crossingCost: 4,
			wantDist: 4,
		},
		{
			name: "boundary shortcut beats long intra-region path",
			regions: map[string]*Region{
				"a": {
					ID: "a",
					Hexes: makeHexes(
						HexCoord{0, 0}, HexCoord{1, 0}, HexCoord{2, 0},
						HexCoord{3, 0}, HexCoord{4, 0}, HexCoord{5, 0},
					),
					Boundaries: []BoundaryConnection{
						{From: HexCoord{0, 0}, ToRegion: "b", To: HexCoord{0, 0}},
						{From: HexCoord{5, 0}, ToRegion: "b", To: HexCoord{0, 0}},
					},
				},
				"b": {ID: "b", Hexes: makeHexes(HexCoord{0, 0})},
			},
			fragments: map[string]*Fragment{
				"x": {id: "x", Region: "a", Hex: HexCoord{0, 0}},
				"y": {id: "y", Region: "a", Hex: HexCoord{5, 0}},
			},
			fromID: "x", toID: "y", crossingCost: 1,
			wantDist: 2,
		},
		{
			name: "no path between disconnected regions",
			regions: map[string]*Region{
				"a": {ID: "a", Hexes: makeHexes(HexCoord{0, 0})},
				"b": {ID: "b", Hexes: makeHexes(HexCoord{0, 0})},
			},
			fragments: map[string]*Fragment{
				"x": {id: "x", Region: "a", Hex: HexCoord{0, 0}},
				"y": {id: "y", Region: "b", Hex: HexCoord{0, 0}},
			},
			fromID: "x", toID: "y", crossingCost: 1,
			wantErr: ErrNoPath,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			hybridMap := &HybridMap{regions: testCase.regions, fragments: testCase.fragments}

			got, err := hybridMap.Distance(testCase.fromID, testCase.toID, testCase.crossingCost)

			if testCase.wantErr != nil {
				if err == nil {
					t.Fatalf("Distance(%q,%q) = %d, want error %v", testCase.fromID, testCase.toID, got, testCase.wantErr)
				}
				if !errors.Is(err, testCase.wantErr) {
					t.Errorf("error = %v, want errors.Is(_, %v)", err, testCase.wantErr)
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

func TestHybridMap_Location(t *testing.T) {
	fragment := &Fragment{
		id:         "tartarus",
		name:       "Tartarus",
		techLevel:  3,
		population: 500000,
		Region:     "corona-reach",
		Hex:        HexCoord{1, 1},
	}
	hybridMap := &HybridMap{
		regions:   map[string]*Region{"corona-reach": {ID: "corona-reach"}},
		fragments: map[string]*Fragment{"tartarus": fragment},
	}

	t.Run("known id returns Location and true", func(t *testing.T) {
		got, ok := hybridMap.Location("tartarus")
		if !ok {
			t.Fatal("Location(tartarus): ok=false, want true")
		}
		if got.ID() != "tartarus" {
			t.Errorf("ID() = %q, want %q", got.ID(), "tartarus")
		}
		if got.Name() != "Tartarus" {
			t.Errorf("Name() = %q, want %q", got.Name(), "Tartarus")
		}
		if got.TechLevel() != 3 {
			t.Errorf("TechLevel() = %d, want 3", got.TechLevel())
		}
		if got.Population() != 500000 {
			t.Errorf("Population() = %d, want 500000", got.Population())
		}
	})

	t.Run("unknown id returns nil and false", func(t *testing.T) {
		got, ok := hybridMap.Location("does-not-exist")
		if ok {
			t.Error("Location(does-not-exist): ok=true, want false")
		}
		if got != nil {
			t.Errorf("Location(does-not-exist): got %v, want nil", got)
		}
	})
}

func TestLoadHybrid(t *testing.T) {
	const validRegions = `
[[region]]
id    = "alpha"
name  = "Alpha"
hexes = [[0,0],[1,0]]

  [[region.boundary]]
  from_q    = 1
  from_r    = 0
  to_region = "beta"
  to_q      = 0
  to_r      = 0

[[region]]
id    = "beta"
name  = "Beta"
hexes = [[0,0]]
`

	const validFragments = `
[[fragment]]
id         = "f1"
name       = "Frag One"
tech_level = 4
population = 100
region     = "alpha"
hex_q      = 0
hex_r      = 0

[[fragment]]
id         = "f2"
name       = "Frag Two"
tech_level = 2
population = 50
region     = "beta"
hex_q      = 0
hex_r      = 0
`

	cases := []struct {
		name          string
		regionsTOML   string
		fragmentsTOML string
		writeRegions  bool
		writeFragments bool
		wantErrSubstr string
		verify        func(t *testing.T, hybridMap *HybridMap)
	}{
		{
			name:          "happy path loads regions and fragments",
			regionsTOML:   validRegions,
			fragmentsTOML: validFragments,
			writeRegions:  true,
			writeFragments: true,
			verify: func(t *testing.T, hybridMap *HybridMap) {
				if loc, ok := hybridMap.Location("f1"); !ok || loc.Name() != "Frag One" {
					t.Errorf("Location(f1) = %v, %v; want Frag One", loc, ok)
				}
				if loc, ok := hybridMap.Location("f2"); !ok || loc.TechLevel() != 2 {
					t.Errorf("Location(f2) tech_level = %d, want 2", loc.TechLevel())
				}
			},
		},
		{
			name:           "missing regions.toml errors",
			fragmentsTOML:  validFragments,
			writeRegions:   false,
			writeFragments: true,
			wantErrSubstr:  "regions.toml",
		},
		{
			name:           "missing fragments.toml errors",
			regionsTOML:    validRegions,
			writeRegions:   true,
			writeFragments: false,
			wantErrSubstr:  "fragments.toml",
		},
		{
			name:           "malformed TOML errors",
			regionsTOML:    "this is = not [valid toml",
			fragmentsTOML:  validFragments,
			writeRegions:   true,
			writeFragments: true,
			wantErrSubstr:  "regions.toml",
		},
		{
			name: "fragment references unknown region",
			regionsTOML: `[[region]]
id = "alpha"
hexes = [[0,0]]
`,
			fragmentsTOML: `[[fragment]]
id = "f1"
region = "ghost"
hex_q = 0
hex_r = 0
`,
			writeRegions:   true,
			writeFragments: true,
			wantErrSubstr:  "unknown region",
		},
		{
			name: "fragment hex outside its region",
			regionsTOML: `[[region]]
id = "alpha"
hexes = [[0,0]]
`,
			fragmentsTOML: `[[fragment]]
id = "f1"
region = "alpha"
hex_q = 5
hex_r = 5
`,
			writeRegions:   true,
			writeFragments: true,
			wantErrSubstr:  "not within region",
		},
		{
			name: "boundary From not in declaring region",
			regionsTOML: `[[region]]
id = "alpha"
hexes = [[0,0]]

  [[region.boundary]]
  from_q = 9
  from_r = 9
  to_region = "beta"
  to_q = 0
  to_r = 0

[[region]]
id = "beta"
hexes = [[0,0]]
`,
			fragmentsTOML:  "",
			writeRegions:   true,
			writeFragments: true,
			wantErrSubstr:  "From=(9,9)",
		},
		{
			name: "boundary references unknown to_region",
			regionsTOML: `[[region]]
id = "alpha"
hexes = [[0,0]]

  [[region.boundary]]
  from_q = 0
  from_r = 0
  to_region = "ghost"
  to_q = 0
  to_r = 0
`,
			fragmentsTOML:  "",
			writeRegions:   true,
			writeFragments: true,
			wantErrSubstr:  `unknown region "ghost"`,
		},
		{
			name: "boundary To not in target region",
			regionsTOML: `[[region]]
id = "alpha"
hexes = [[0,0]]

  [[region.boundary]]
  from_q = 0
  from_r = 0
  to_region = "beta"
  to_q = 9
  to_r = 9

[[region]]
id = "beta"
hexes = [[0,0]]
`,
			fragmentsTOML:  "",
			writeRegions:   true,
			writeFragments: true,
			wantErrSubstr:  "To=(9,9)",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			dir := t.TempDir()
			if testCase.writeRegions {
				if err := os.WriteFile(filepath.Join(dir, "regions.toml"), []byte(testCase.regionsTOML), 0o644); err != nil {
					t.Fatalf("setup regions.toml: %v", err)
				}
			}
			if testCase.writeFragments {
				if err := os.WriteFile(filepath.Join(dir, "fragments.toml"), []byte(testCase.fragmentsTOML), 0o644); err != nil {
					t.Fatalf("setup fragments.toml: %v", err)
				}
			}

			hybridMap, err := LoadHybrid(dir)

			if testCase.wantErrSubstr != "" {
				if err == nil {
					t.Fatalf("LoadHybrid: want error containing %q, got nil", testCase.wantErrSubstr)
				}
				if !strings.Contains(err.Error(), testCase.wantErrSubstr) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), testCase.wantErrSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadHybrid: unexpected error: %v", err)
			}
			if testCase.verify != nil {
				testCase.verify(t, hybridMap)
			}
		})
	}
}

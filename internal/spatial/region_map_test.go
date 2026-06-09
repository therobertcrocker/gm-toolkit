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

func TestRegionMap_Distance(t *testing.T) {
	cases := []struct {
		name         string
		regions      map[string]*Region
		warps        []warpRecord
		from         RegionHex
		to           RegionHex
		crossingCost int
		wantDist     int
		wantErr      error
	}{
		{
			name: "same regionhex returns 0",
			regions: map[string]*Region{
				"r1": {ID: "r1", Hexes: makeHexes(HexCoord{0, 0})},
			},
			from:         RegionHex{RegionID: "r1", Coord: HexCoord{0, 0}},
			to:           RegionHex{RegionID: "r1", Coord: HexCoord{0, 0}},
			crossingCost: 5,
			wantDist:     0,
		},
		{
			name: "adjacent intra-region",
			regions: map[string]*Region{
				"r1": {ID: "r1", Hexes: makeHexes(HexCoord{0, 0}, HexCoord{1, 0})},
			},
			from:         RegionHex{RegionID: "r1", Coord: HexCoord{0, 0}},
			to:           RegionHex{RegionID: "r1", Coord: HexCoord{1, 0}},
			crossingCost: 5,
			wantDist:     1,
		},
		{
			name: "non-adjacent intra-region",
			regions: map[string]*Region{
				"r1": {ID: "r1", Hexes: makeHexes(
					HexCoord{0, 0}, HexCoord{1, 0}, HexCoord{2, 0}, HexCoord{3, 0},
				)},
			},
			from:         RegionHex{RegionID: "r1", Coord: HexCoord{0, 0}},
			to:           RegionHex{RegionID: "r1", Coord: HexCoord{3, 0}},
			crossingCost: 5,
			wantDist:     3,
		},
		{
			name: "hole routing forces detour",
			regions: map[string]*Region{
				"r1": {ID: "r1", Hexes: makeHexes(
					HexCoord{0, 0}, HexCoord{0, 1}, HexCoord{1, 1}, HexCoord{2, 1}, HexCoord{2, 0},
				)},
			},
			from:         RegionHex{RegionID: "r1", Coord: HexCoord{0, 0}},
			to:           RegionHex{RegionID: "r1", Coord: HexCoord{2, 0}},
			crossingCost: 5,
			wantDist:     3,
		},
		{
			name: "cross-region adjacent steps cost 1 not crossingCost",
			regions: map[string]*Region{
				"a": {ID: "a", Hexes: makeHexes(HexCoord{0, 0}, HexCoord{1, 0})},
				"b": {ID: "b", Hexes: makeHexes(HexCoord{2, 0}, HexCoord{3, 0})},
			},
			from:         RegionHex{RegionID: "a", Coord: HexCoord{0, 0}},
			to:           RegionHex{RegionID: "b", Coord: HexCoord{3, 0}},
			crossingCost: 5,
			wantDist:     3,
		},
		{
			name: "cross-region warp charges crossingCost",
			regions: map[string]*Region{
				"a": {ID: "a", Hexes: makeHexes(HexCoord{0, 0})},
				"b": {ID: "b", Hexes: makeHexes(HexCoord{50, 0})},
			},
			warps: []warpRecord{
				{fromRegion: "a", from: HexCoord{0, 0}, toRegion: "b", to: HexCoord{50, 0}},
			},
			from:         RegionHex{RegionID: "a", Coord: HexCoord{0, 0}},
			to:           RegionHex{RegionID: "b", Coord: HexCoord{50, 0}},
			crossingCost: 3,
			wantDist:     3,
		},
		{
			name: "cross-region warp traversal reverse direction",
			regions: map[string]*Region{
				"a": {ID: "a", Hexes: makeHexes(HexCoord{0, 0})},
				"b": {ID: "b", Hexes: makeHexes(HexCoord{50, 0})},
			},
			warps: []warpRecord{
				{fromRegion: "a", from: HexCoord{0, 0}, toRegion: "b", to: HexCoord{50, 0}},
			},
			from:         RegionHex{RegionID: "b", Coord: HexCoord{50, 0}},
			to:           RegionHex{RegionID: "a", Coord: HexCoord{0, 0}},
			crossingCost: 3,
			wantDist:     3,
		},
		{
			name: "two warps chained",
			regions: map[string]*Region{
				"a": {ID: "a", Hexes: makeHexes(HexCoord{0, 0})},
				"b": {ID: "b", Hexes: makeHexes(HexCoord{10, 0})},
				"c": {ID: "c", Hexes: makeHexes(HexCoord{20, 0})},
			},
			warps: []warpRecord{
				{fromRegion: "a", from: HexCoord{0, 0}, toRegion: "b", to: HexCoord{10, 0}},
				{fromRegion: "b", from: HexCoord{10, 0}, toRegion: "c", to: HexCoord{20, 0}},
			},
			from:         RegionHex{RegionID: "a", Coord: HexCoord{0, 0}},
			to:           RegionHex{RegionID: "c", Coord: HexCoord{20, 0}},
			crossingCost: 3,
			wantDist:     6,
		},
		{
			name: "warp shortcut beats long intra-region path",
			regions: map[string]*Region{
				"a": {ID: "a", Hexes: makeHexes(
					HexCoord{0, 0}, HexCoord{1, 0}, HexCoord{2, 0},
					HexCoord{3, 0}, HexCoord{4, 0}, HexCoord{5, 0},
				)},
				"b": {ID: "b", Hexes: makeHexes(HexCoord{100, 0})},
			},
			warps: []warpRecord{
				{fromRegion: "a", from: HexCoord{0, 0}, toRegion: "b", to: HexCoord{100, 0}},
				{fromRegion: "a", from: HexCoord{5, 0}, toRegion: "b", to: HexCoord{100, 0}},
			},
			from:         RegionHex{RegionID: "a", Coord: HexCoord{0, 0}},
			to:           RegionHex{RegionID: "a", Coord: HexCoord{5, 0}},
			crossingCost: 1,
			wantDist:     2,
		},
		{
			name: "negative cost is invalid",
			regions: map[string]*Region{
				"r1": {ID: "r1", Hexes: makeHexes(HexCoord{0, 0})},
			},
			from:         RegionHex{RegionID: "r1", Coord: HexCoord{0, 0}},
			to:           RegionHex{RegionID: "r1", Coord: HexCoord{0, 0}},
			crossingCost: -1,
			wantErr:      ErrInvalidCost,
		},
		{
			name: "no path between disconnected regions",
			regions: map[string]*Region{
				"a": {ID: "a", Hexes: makeHexes(HexCoord{0, 0})},
				"b": {ID: "b", Hexes: makeHexes(HexCoord{100, 0})},
			},
			from:         RegionHex{RegionID: "a", Coord: HexCoord{0, 0}},
			to:           RegionHex{RegionID: "b", Coord: HexCoord{100, 0}},
			crossingCost: 1,
			wantErr:      ErrNoPath,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			regionMap, err := newRegionMap(testCase.regions, testCase.warps)
			if err != nil {
				t.Fatalf("setup: %v", err)
			}
			regionMap.worlds = map[string]*World{}

			got, err := regionMap.Distance(testCase.from, testCase.to, testCase.crossingCost)

			if testCase.wantErr != nil {
				if err == nil {
					t.Fatalf("Distance(%v,%v) = %d, want error %v", testCase.from, testCase.to, got, testCase.wantErr)
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
				t.Errorf("Distance(%v,%v) = %d, want %d", testCase.from, testCase.to, got, testCase.wantDist)
			}
		})
	}
}

func TestRegionMap_Path(t *testing.T) {
	cases := []struct {
		name         string
		regions      map[string]*Region
		warps        []warpRecord
		from         RegionHex
		to           RegionHex
		crossingCost int
		wantPath     []RegionHex
		wantCost     int
		wantErr      error
	}{
		{
			name: "same regionhex returns single-step path with 0 cost",
			regions: map[string]*Region{
				"r1": {ID: "r1", Hexes: makeHexes(HexCoord{0, 0})},
			},
			from:         RegionHex{RegionID: "r1", Coord: HexCoord{0, 0}},
			to:           RegionHex{RegionID: "r1", Coord: HexCoord{0, 0}},
			crossingCost: 5,
			wantPath:     []RegionHex{{RegionID: "r1", Coord: HexCoord{0, 0}}},
			wantCost:     0,
		},
		{
			name: "adjacent intra-region",
			regions: map[string]*Region{
				"r1": {ID: "r1", Hexes: makeHexes(HexCoord{0, 0}, HexCoord{1, 0})},
			},
			from:         RegionHex{RegionID: "r1", Coord: HexCoord{0, 0}},
			to:           RegionHex{RegionID: "r1", Coord: HexCoord{1, 0}},
			crossingCost: 5,
			wantPath: []RegionHex{
				{RegionID: "r1", Coord: HexCoord{0, 0}},
				{RegionID: "r1", Coord: HexCoord{1, 0}},
			},
			wantCost: 1,
		},
		{
			name: "cross-region adjacent hop costs 1 not crossingCost",
			regions: map[string]*Region{
				"a": {ID: "a", Hexes: makeHexes(HexCoord{0, 0}, HexCoord{1, 0})},
				"b": {ID: "b", Hexes: makeHexes(HexCoord{2, 0}, HexCoord{3, 0})},
			},
			from:         RegionHex{RegionID: "a", Coord: HexCoord{0, 0}},
			to:           RegionHex{RegionID: "b", Coord: HexCoord{3, 0}},
			crossingCost: 5,
			wantPath: []RegionHex{
				{RegionID: "a", Coord: HexCoord{0, 0}},
				{RegionID: "a", Coord: HexCoord{1, 0}},
				{RegionID: "b", Coord: HexCoord{2, 0}},
				{RegionID: "b", Coord: HexCoord{3, 0}},
			},
			wantCost: 3,
		},
		{
			name: "negative cost is invalid",
			regions: map[string]*Region{
				"r1": {ID: "r1", Hexes: makeHexes(HexCoord{0, 0})},
			},
			from:         RegionHex{RegionID: "r1", Coord: HexCoord{0, 0}},
			to:           RegionHex{RegionID: "r1", Coord: HexCoord{0, 0}},
			crossingCost: -1,
			wantErr:      ErrInvalidCost,
		},
		{
			name: "no path between disconnected regions",
			regions: map[string]*Region{
				"a": {ID: "a", Hexes: makeHexes(HexCoord{0, 0})},
				"b": {ID: "b", Hexes: makeHexes(HexCoord{100, 0})},
			},
			from:         RegionHex{RegionID: "a", Coord: HexCoord{0, 0}},
			to:           RegionHex{RegionID: "b", Coord: HexCoord{100, 0}},
			crossingCost: 1,
			wantErr:      ErrNoPath,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			regionMap, err := newRegionMap(testCase.regions, testCase.warps)
			if err != nil {
				t.Fatalf("setup: %v", err)
			}
			regionMap.worlds = map[string]*World{}
			gotPath, gotCost, err := regionMap.Path(testCase.from, testCase.to, testCase.crossingCost)
			if testCase.wantErr != nil {
				if err == nil {
					t.Fatalf("Path(%v,%v) = %v, %d, want error %v", testCase.from, testCase.to, gotPath, gotCost, testCase.wantErr)
				}
				if !errors.Is(err, testCase.wantErr) {
					t.Errorf("error = %v, want errors.Is(_, %v)", err, testCase.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotCost != testCase.wantCost {
				t.Errorf("Path(%v,%v) cost = %d, want %d", testCase.from, testCase.to, gotCost, testCase.wantCost)
			}
			if len(gotPath) != len(testCase.wantPath) {
				t.Fatalf("Path(%v,%v) path length = %d, want %d\n  got:  %v\n  want: %v",
					testCase.from, testCase.to, len(gotPath), len(testCase.wantPath), gotPath, testCase.wantPath)
			}
			for i := range gotPath {
				if gotPath[i] != testCase.wantPath[i] {
					t.Errorf("Path(%v,%v) path[%d] = %v, want %v", testCase.from, testCase.to, i, gotPath[i], testCase.wantPath[i])
				}
			}
		})
	}
}

func TestRegionMap_Location(t *testing.T) {
	world := &World{
		id:         "tartarus",
		name:       "Tartarus",
		techLevel:  3,
		population: 500000,
		loc:        RegionHex{RegionID: "corona-reach", Coord: HexCoord{1, 1}},
	}
	regionMap := &RegionMap{
		regions: map[string]*Region{"corona-reach": {ID: "corona-reach"}},
		worlds:  map[string]*World{"tartarus": world},
	}

	t.Run("known id returns Location and true", func(t *testing.T) {
		got, ok := regionMap.Location("tartarus")
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
		got, ok := regionMap.Location("does-not-exist")
		if ok {
			t.Error("Location(does-not-exist): ok=true, want false")
		}
		if got != nil {
			t.Errorf("Location(does-not-exist): got %v, want nil", got)
		}
	})
}

func TestLoadRegionMap(t *testing.T) {
	const validRegions = `
[[region]]
id    = "alpha"
name  = "Alpha"
hexes = [[0,0],[1,0]]

[[region]]
id    = "beta"
name  = "Beta"
hexes = [[10,0]]

[[warp]]
from_region = "alpha"
from_q = 1
from_r = 0
to_region = "beta"
to_q = 10
to_r = 0
`

	const validWorlds = `
[[world]]
id         = "f1"
name       = "Frag One"
tech_level = 4
population = 100
region     = "alpha"
hex_q      = 0
hex_r      = 0

[[world]]
id         = "f2"
name       = "Frag Two"
tech_level = 2
population = 50
region     = "beta"
hex_q      = 10
hex_r      = 0
`

	cases := []struct {
		name          string
		regionsTOML   string
		worldsTOML    string
		writeRegions  bool
		writeWorlds   bool
		wantErrSubstr string
		verify        func(t *testing.T, regionMap *RegionMap)
	}{
		{
			name:         "happy path loads regions worlds and bidirectional warps",
			regionsTOML:  validRegions,
			worldsTOML:   validWorlds,
			writeRegions: true,
			writeWorlds:  true,
			verify: func(t *testing.T, regionMap *RegionMap) {
				if loc, ok := regionMap.Location("f1"); !ok || loc.Name() != "Frag One" {
					t.Errorf("Location(f1) = %v, %v; want Frag One", loc, ok)
				}
				if loc, ok := regionMap.Location("f2"); !ok || loc.TechLevel() != 2 {
					t.Errorf("Location(f2) tech_level = %d, want 2", loc.TechLevel())
				}
				forward := regionMap.warps[HexCoord{1, 0}]
				if len(forward) != 1 || forward[0].toRegion != "beta" || forward[0].to != (HexCoord{10, 0}) {
					t.Errorf("warps[(1,0)] = %v, want [{toRegion:beta to:(10,0)}]", forward)
				}
				reverse := regionMap.warps[HexCoord{10, 0}]
				if len(reverse) != 1 || reverse[0].toRegion != "alpha" || reverse[0].to != (HexCoord{1, 0}) {
					t.Errorf("warps[(10,0)] = %v, want [{toRegion:alpha to:(1,0)}]", reverse)
				}
			},
		},
		{
			name:          "missing regions.toml errors",
			worldsTOML:    validWorlds,
			writeRegions:  false,
			writeWorlds:   true,
			wantErrSubstr: "regions.toml",
		},
		{
			name:          "missing worlds.toml errors",
			regionsTOML:   validRegions,
			writeRegions:  true,
			writeWorlds:   false,
			wantErrSubstr: "worlds.toml",
		},
		{
			name:          "malformed TOML errors",
			regionsTOML:   "this is = not [valid toml",
			worldsTOML:    validWorlds,
			writeRegions:  true,
			writeWorlds:   true,
			wantErrSubstr: "regions.toml",
		},
		{
			name: "world references unknown region",
			regionsTOML: `[[region]]
id = "alpha"
hexes = [[0,0]]
`,
			worldsTOML: `[[world]]
id = "f1"
region = "ghost"
hex_q = 0
hex_r = 0
`,
			writeRegions:  true,
			writeWorlds:   true,
			wantErrSubstr: "unknown region",
		},
		{
			name: "world hex outside its region",
			regionsTOML: `[[region]]
id = "alpha"
hexes = [[0,0]]
`,
			worldsTOML: `[[world]]
id = "f1"
region = "alpha"
hex_q = 5
hex_r = 5
`,
			writeRegions:  true,
			writeWorlds:   true,
			wantErrSubstr: "not within region",
		},
		{
			name: "warp From not in fromRegion",
			regionsTOML: `[[region]]
id = "alpha"
hexes = [[0,0]]

[[region]]
id = "beta"
hexes = [[10,0]]

[[warp]]
from_region = "alpha"
from_q = 9
from_r = 9
to_region = "beta"
to_q = 10
to_r = 0
`,
			worldsTOML:    "",
			writeRegions:  true,
			writeWorlds:   true,
			wantErrSubstr: "From=(9,9)",
		},
		{
			name: "warp references unknown toRegion",
			regionsTOML: `[[region]]
id = "alpha"
hexes = [[0,0]]

[[warp]]
from_region = "alpha"
from_q = 0
from_r = 0
to_region = "ghost"
to_q = 0
to_r = 0
`,
			worldsTOML:    "",
			writeRegions:  true,
			writeWorlds:   true,
			wantErrSubstr: `unknown region "ghost"`,
		},
		{
			name: "warp To not in toRegion",
			regionsTOML: `[[region]]
id = "alpha"
hexes = [[0,0]]

[[region]]
id = "beta"
hexes = [[10,0]]

[[warp]]
from_region = "alpha"
from_q = 0
from_r = 0
to_region = "beta"
to_q = 9
to_r = 9
`,
			worldsTOML:    "",
			writeRegions:  true,
			writeWorlds:   true,
			wantErrSubstr: "To=(9,9)",
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
			if testCase.writeWorlds {
				if err := os.WriteFile(filepath.Join(dir, "worlds.toml"), []byte(testCase.worldsTOML), 0o644); err != nil {
					t.Fatalf("setup worlds.toml: %v", err)
				}
			}

			regionMap, err := LoadRegionMap(dir)

			if testCase.wantErrSubstr != "" {
				if err == nil {
					t.Fatalf("LoadRegionMap: want error containing %q, got nil", testCase.wantErrSubstr)
				}
				if !strings.Contains(err.Error(), testCase.wantErrSubstr) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), testCase.wantErrSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadRegionMap: unexpected error: %v", err)
			}
			if testCase.verify != nil {
				testCase.verify(t, regionMap)
			}
		})
	}
}

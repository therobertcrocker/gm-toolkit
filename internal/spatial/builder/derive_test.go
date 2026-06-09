package builder

import (
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

// mkLayout builds a layoutFile from a glyph grid. Optional warpAt positions set
// Warp=true on the named cells after the grid is assembled.
func mkLayout(grid [][]glyph, warpAt ...gridCoord) *layoutFile {
	layout := &layoutFile{
		regionLetters: map[glyph]bool{},
		markers:       map[glyph]cell{},
	}
	for row, rowGlyphs := range grid {
		var rowCells []cell
		for col, g := range rowGlyphs {
			c := cell{Glyph: g, Row: row, Col: col, FileLine: row + 1, FileCol: col + 1}
			rowCells = append(rowCells, c)
			if g.isRegion() {
				layout.regionLetters[g] = true
			}
			if g.isWorld() {
				layout.markers[g] = c
			}
		}
		layout.cells = append(layout.cells, rowCells)
	}
	for _, pos := range warpAt {
		layout.cells[pos.Row][pos.Col].Warp = true
	}
	return layout
}

func mkData(regionIDs []string, worlds []worldEntry) *dataFile {
	out := &dataFile{
		Regions: map[string]regionEntry{},
		Worlds:  worlds,
	}
	for _, id := range regionIDs {
		out.Regions[id] = regionEntry{Name: "Region " + id}
	}
	return out
}

func findWorld(derived *derivedMap, id string) *derivedWorld {
	for i := range derived.Worlds {
		if derived.Worlds[i].ID == id {
			return &derived.Worlds[i]
		}
	}
	return nil
}

func TestDerive_RegionInference(t *testing.T) {
	cases := []struct {
		name    string
		layout  *layoutFile
		data    *dataFile
		wantErr string
		check   func(*testing.T, *derivedMap)
	}{
		{
			name:   "single-candidate neighbor auto-infers region",
			layout: mkLayout([][]glyph{{'A', '1', 'A'}}),
			data: mkData(
				[]string{"A"},
				[]worldEntry{{ID: "x", Name: "X", At: atGlyph('1')}},
			),
			check: func(t *testing.T, derived *derivedMap) {
				w := findWorld(derived, "x")
				if w == nil {
					t.Fatal("world 'x' missing from derived")
				}
				if w.Region != "A" {
					t.Errorf("world.Region = %q, want %q", w.Region, "A")
				}
			},
		},
		{
			name:   "multi-candidate neighbors error without override",
			layout: mkLayout([][]glyph{{'A', '1', 'B'}}),
			data: mkData(
				[]string{"A", "B"},
				[]worldEntry{{ID: "x", Name: "X", At: atGlyph('1')}},
			),
			wantErr: "multiple region candidates",
		},
		{
			name: "zero-candidate (no region-letter neighbors) error",
			layout: mkLayout([][]glyph{
				{'A', '.', '.', '.', '.', '1', '.', '.'},
			}),
			data: mkData(
				[]string{"A"},
				[]worldEntry{{ID: "x", Name: "X", At: atGlyph('1')}},
			),
			wantErr: "no region-letter neighbors",
		},
		{
			name:   "explicit override honored on multi-candidate marker",
			layout: mkLayout([][]glyph{{'A', '1', 'B'}}),
			data: mkData(
				[]string{"A", "B"},
				[]worldEntry{{ID: "x", Name: "X", At: atGlyph('1'), Region: "A"}},
			),
			check: func(t *testing.T, derived *derivedMap) {
				w := findWorld(derived, "x")
				if w == nil {
					t.Fatal("world 'x' missing from derived")
				}
				if w.Region != "A" {
					t.Errorf("world.Region = %q, want %q (override should win)", w.Region, "A")
				}
			},
		},
		{
			name:   "override naming undeclared region errors",
			layout: mkLayout([][]glyph{{'A', '1', 'A'}}),
			data: mkData(
				[]string{"A"},
				[]worldEntry{{ID: "x", Name: "X", At: atGlyph('1'), Region: "C"}},
			),
			wantErr: `region override "C" is not a declared region`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			derived, err := derive(tc.layout, tc.data)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("got nil error, want substring %q", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %q, want substring %q", err.Error(), tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.check != nil {
				tc.check(t, derived)
			}
		})
	}
}

func TestDeriveWarps(t *testing.T) {
	// h is a shorthand for spatial.HexCoord used in expected warp lists.
	h := func(q, r int) spatial.HexCoord { return spatial.HexCoord{Q: q, R: r} }

	cases := []struct {
		name      string
		layout    *layoutFile
		data      *dataFile
		wantErr   string
		wantWarps []derivedWarp
		wantWarns []string
	}{
		{
			// Two isolated regions separated by empty space; clear LoS → one warp.
			name: "two marks clear LoS",
			layout: mkLayout([][]glyph{
				{'A', '.', '.', 'D'},
			}, gridCoord{0, 0}, gridCoord{0, 3}),
			data: mkData([]string{"A", "D"}, nil),
			wantWarps: []derivedWarp{
				{FromRegion: "A", From: h(0, 0), ToRegion: "D", To: h(3, 0)},
			},
		},
		{
			// Four marks in four regions; A↔B and C↔D are hex-adjacent (no warp
			// edge needed); A↔C, A↔D, B↔C, B↔D all have clear LoS → four warps.
			name: "four marks two adjacent exclusions",
			layout: mkLayout([][]glyph{
				{'A', '.', '.', '.', 'D'},
				{'B', '.', '.', 'C', '.'},
			}, gridCoord{0, 0}, gridCoord{0, 4}, gridCoord{1, 0}, gridCoord{1, 3}),
			data: mkData([]string{"A", "B", "C", "D"}, nil),
			wantWarps: []derivedWarp{
				{FromRegion: "A", From: h(0, 0), ToRegion: "C", To: h(3, 1)},
				{FromRegion: "A", From: h(0, 0), ToRegion: "D", To: h(4, 0)},
				{FromRegion: "B", From: h(0, 1), ToRegion: "C", To: h(3, 1)},
				{FromRegion: "B", From: h(0, 1), ToRegion: "D", To: h(4, 0)},
			},
		},
		{
			// LoS between A* and B* is blocked by an occupied A hex; D* connects
			// only to B* → warps: B↔D only; A* is isolated → warning.
			name: "mark blocked by occupied hex",
			layout: mkLayout([][]glyph{
				{'A', 'A', 'B', 'B', '.', 'D'},
			}, gridCoord{0, 0}, gridCoord{0, 5}),
			data: mkData([]string{"A", "B", "D"}, nil),
			wantWarns: []string{
				"hex A(0,0) is marked for warp but reaches no other region",
				"hex D(5,0) is marked for warp but reaches no other region",
			},
		},
		{
			// A mark completely surrounded by same-region hexes → fatal error.
			name: "interior mark error",
			layout: mkLayout([][]glyph{
				{'.', 'A', 'A', 'A', '.'},
				{'.', 'A', 'A', 'A', '.'},
				{'.', 'A', 'A', 'A', '.'},
			}, gridCoord{1, 2}),
			data:    mkData([]string{"A"}, nil),
			wantErr: "is not on a region boundary",
		},
		{
			// A mark on a boundary hex with no reachable different-region mark →
			// warning (non-fatal).
			name: "dead mark produces warning",
			layout: mkLayout([][]glyph{
				{'A', 'A', 'A'},
			}, gridCoord{0, 0}),
			data: mkData([]string{"A"}, nil),
			wantWarns: []string{
				"hex A(0,0) is marked for warp but reaches no other region",
			},
		},
		{
			// Two adjacent marks in different regions (hexDistance == 1) already
			// connect via a hex step; no warp edge, no warning.
			name: "adjacent marks no warp edge",
			layout: mkLayout([][]glyph{
				{'A', 'B'},
			}, gridCoord{0, 0}, gridCoord{0, 1}),
			data:      mkData([]string{"A", "B"}, nil),
			wantWarps: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			derived, err := derive(tc.layout, tc.data)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("got nil error, want substring %q", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %q, want substring %q", err.Error(), tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(derived.Warps) != len(tc.wantWarps) {
				t.Errorf("len(Warps) = %d, want %d\ngot:  %v\nwant: %v",
					len(derived.Warps), len(tc.wantWarps), derived.Warps, tc.wantWarps)
			} else {
				for i, w := range derived.Warps {
					want := tc.wantWarps[i]
					if w != want {
						t.Errorf("Warps[%d] = %v, want %v", i, w, want)
					}
				}
			}

			if len(derived.Warnings) != len(tc.wantWarns) {
				t.Errorf("len(Warnings) = %d, want %d\ngot:  %v\nwant: %v",
					len(derived.Warnings), len(tc.wantWarns), derived.Warnings, tc.wantWarns)
			} else {
				for i, w := range derived.Warnings {
					if !strings.Contains(w, tc.wantWarns[i]) {
						t.Errorf("Warnings[%d] = %q, want substring %q", i, w, tc.wantWarns[i])
					}
				}
			}
		})
	}
}

package builder

import (
	"strings"
	"testing"
)

func mkLayout(grid [][]glyph) *layoutFile {
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
	return layout
}

func mkData(regionIDs []string, worlds []worldEntry, warps []warpEntry) *dataFile {
	out := &dataFile{
		Regions: map[string]regionEntry{},
		Worlds:  worlds,
		Warps:   warps,
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

func findRegion(derived *derivedMap, id string) *derivedRegion {
	for i := range derived.Regions {
		if derived.Regions[i].ID == id {
			return &derived.Regions[i]
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
				[]worldEntry{{ID: "x", Name: "X", At: atGlyph{Glyph: '1'}}},
				nil,
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
				[]worldEntry{{ID: "x", Name: "X", At: atGlyph{Glyph: '1'}}},
				nil,
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
				[]worldEntry{{ID: "x", Name: "X", At: atGlyph{Glyph: '1'}}},
				nil,
			),
			wantErr: "no region-letter neighbors",
		},
		{
			name:   "explicit override honored on multi-candidate marker",
			layout: mkLayout([][]glyph{{'A', '1', 'B'}}),
			data: mkData(
				[]string{"A", "B"},
				[]worldEntry{{ID: "x", Name: "X", At: atGlyph{Glyph: '1'}, Region: "A"}},
				nil,
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
				[]worldEntry{{ID: "x", Name: "X", At: atGlyph{Glyph: '1'}, Region: "C"}},
				nil,
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

func TestDerive_AdjacencyCanonicalization(t *testing.T) {
	// A and B share an edge on a single row. The derive step should emit the
	// shared edge exactly once, on the alphabetically-first region's side (A),
	// leaving B's boundary list empty.
	layout := mkLayout([][]glyph{{'A', 'B'}})
	data := mkData([]string{"A", "B"}, nil, nil)

	derived, err := derive(layout, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	regionA := findRegion(derived, "A")
	regionB := findRegion(derived, "B")
	if regionA == nil || regionB == nil {
		t.Fatalf("missing regions: A=%v B=%v", regionA, regionB)
	}

	if got := len(regionA.Boundaries); got != 1 {
		t.Errorf("len(A.Boundaries) = %d, want 1", got)
	} else {
		b := regionA.Boundaries[0]
		if b.ToRegion != "B" {
			t.Errorf("A.Boundaries[0].ToRegion = %q, want %q", b.ToRegion, "B")
		}
	}

	if got := len(regionB.Boundaries); got != 0 {
		t.Errorf("len(B.Boundaries) = %d, want 0 (boundary should canonicalize to A's side)", got)
	}

	if got := derived.AdjacencyCount; got != 1 {
		t.Errorf("AdjacencyCount = %d, want 1", got)
	}
}

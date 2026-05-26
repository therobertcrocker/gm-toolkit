package builder

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeLayoutFile(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "layout.txt"), []byte(contents), 0o644); err != nil {
		t.Fatalf("write layout.txt: %v", err)
	}
	return dir
}

func TestParseLayout(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr string
		check   func(*testing.T, *layoutFile)
	}{
		{
			name: "comments and blank lines skipped",
			input: `# header comment
   # indented comment

A . A
 A . A
`,
			check: func(t *testing.T, layout *layoutFile) {
				if got := len(layout.cells); got != 2 {
					t.Fatalf("len(cells) = %d, want 2", got)
				}
				if !layout.regionLetters['A'] {
					t.Errorf("regionLetters missing 'A': %v", layout.regionLetters)
				}
			},
		},
		{
			name: "odd-row indent stripped preserves file column attribution",
			input: `A B C
 D E F
`,
			check: func(t *testing.T, layout *layoutFile) {
				row0Col0 := layout.cells[0][0]
				if row0Col0.Glyph != 'A' || row0Col0.FileLine != 1 || row0Col0.FileCol != 1 {
					t.Errorf("row 0 col 0 = {glyph=%q, line=%d, col=%d}, want {'A', 1, 1}",
						rune(row0Col0.Glyph), row0Col0.FileLine, row0Col0.FileCol)
				}
				row1Col0 := layout.cells[1][0]
				if row1Col0.Glyph != 'D' || row1Col0.FileLine != 2 || row1Col0.FileCol != 2 {
					t.Errorf("row 1 col 0 = {glyph=%q, line=%d, col=%d}, want {'D', 2, 2}",
						rune(row1Col0.Glyph), row1Col0.FileLine, row1Col0.FileCol)
				}
			},
		},
		{
			name:  "trailing whitespace tolerated",
			input: "A B C   \nD E F\n",
			check: func(t *testing.T, layout *layoutFile) {
				if len(layout.cells) != 2 {
					t.Fatalf("len(cells) = %d, want 2", len(layout.cells))
				}
				if len(layout.cells[0]) != 3 || len(layout.cells[1]) != 3 {
					t.Errorf("row widths = [%d, %d], want [3, 3]",
						len(layout.cells[0]), len(layout.cells[1]))
				}
			},
		},
		{
			name: "duplicate world marker",
			input: `A 1 A
 A 1 A
`,
			wantErr: "world marker '1' appears twice",
		},
		{
			name:    "invalid glyph",
			input:   "A ? A\n",
			wantErr: "invalid glyph '?'",
		},
		{
			name:    "missing single-space separator",
			input:   "AB\n",
			wantErr: "expected single-space separator",
		},
		{
			name: "empty grid (only comments and blanks)",
			input: `# only comments

# another comment
`,
			wantErr: "empty grid",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := writeLayoutFile(t, tc.input)
			layout, err := parseLayout(dir)

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
				tc.check(t, layout)
			}
		})
	}
}

func TestParseLayout_MissingFile(t *testing.T) {
	dir := t.TempDir()
	if _, err := parseLayout(dir); !errors.Is(err, ErrMissingSource) {
		t.Errorf("err = %v, want errors.Is(err, ErrMissingSource)", err)
	}
}

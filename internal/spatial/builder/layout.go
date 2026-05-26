package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type glyph rune

func (g glyph) isRegion() bool { return g >= 'A' && g <= 'Z' }
func (g glyph) isWorld() bool  { return (g >= '1' && g <= '9') || (g >= 'a' && g <= 'z') }
func (g glyph) isEmpty() bool  { return g == '.' }

type cell struct {
	Glyph    glyph
	Row      int
	Col      int
	FileLine int
	FileCol  int
}

type layoutFile struct {
	cells         [][]cell
	regionLetters map[glyph]bool
	markers       map[glyph]cell
}

func parseLayout(srcDir string) (*layoutFile, error) {
	path := filepath.Join(srcDir, "layout.txt")
	contents, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrMissingSource, path)
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	layout := &layoutFile{
		regionLetters: make(map[glyph]bool),
		markers:       make(map[glyph]cell),
	}

	for fileLine0, rawLine := range strings.Split(string(contents), "\n") {
		fileLine := fileLine0 + 1
		line := strings.TrimRight(rawLine, " \t\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(strings.TrimLeft(line, " \t"), "#") {
			continue
		}

		row := len(layout.cells)
		cells, err := parseLayoutRow(line, row, fileLine, layout.markers)
		if err != nil {
			return nil, err
		}
		for _, c := range cells {
			if c.Glyph.isRegion() {
				layout.regionLetters[c.Glyph] = true
			}
		}
		layout.cells = append(layout.cells, cells)
	}

	if len(layout.cells) == 0 {
		return nil, fmt.Errorf("layout.txt: empty grid (no non-comment, non-blank rows)")
	}
	return layout, nil
}

// parseLayoutRow scans a single grid row. Odd rows have a one-space indent that
// is rendering, not a cell; it is stripped before scanning. Each emitted cell
// records its 1-indexed FileLine/FileCol in the original (unstripped) file line.
func parseLayoutRow(line string, row, fileLine int, markers map[glyph]cell) ([]cell, error) {
	indentOffset := 0
	if row&1 == 1 && strings.HasPrefix(line, " ") {
		line = line[1:]
		indentOffset = 1
	}

	var cells []cell
	cellCol := 0
	for i := 0; i < len(line); {
		ch := line[i]
		fileCol := i + indentOffset + 1
		g := glyph(ch)
		if !g.isRegion() && !g.isWorld() && !g.isEmpty() {
			return nil, layoutErrorf(fileLine, fileCol, "invalid glyph %q (expected '.', 'A'-'Z', '1'-'9', or 'a'-'z')", ch)
		}
		c := cell{Glyph: g, Row: row, Col: cellCol, FileLine: fileLine, FileCol: fileCol}
		cells = append(cells, c)

		if g.isWorld() {
			if prior, seen := markers[g]; seen {
				return nil, layoutErrorf(fileLine, fileCol, "world marker %q appears twice; first at line=%d col=%d", ch, prior.FileLine, prior.FileCol)
			}
			markers[g] = c
		}
		cellCol++
		i++
		if i >= len(line) {
			break
		}
		if line[i] != ' ' {
			return nil, layoutErrorf(fileLine, i+indentOffset+1, "expected single-space separator, got %q", line[i])
		}
		i++
	}
	if len(cells) == 0 {
		return nil, layoutErrorf(fileLine, 1, "empty row")
	}
	return cells, nil
}

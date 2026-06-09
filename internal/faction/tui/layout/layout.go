package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// Region identifies one of the columns in the working area. To move a component
// to a different column, assign its View to a different Region key — no other
// change is needed; widths and separators follow.
type Region int

const (
	Left Region = iota
	Center
	Right
)

// regionOrder is the left-to-right render order used by Compose.
var regionOrder = []Region{Left, Center, Right}

// Panel is a region's content plus the base style that governs its alignment
// and color. Compose applies the region's width and height on top.
type Panel struct {
	Content string
	Style   lipgloss.Style
}

// RegionWidths splits the working area into a wide center flanked by equal-width
// left and right columns, accounting for the two single-column separators
// between them. Remainder goes to the center.
func RegionWidths(termWidth int) map[Region]int {
	totalW := max(termWidth-2, 4) // two 1-col separators
	left := totalW / 4
	right := totalW / 4
	return map[Region]int{
		Left:   left,
		Center: totalW - left - right,
		Right:  right,
	}
}

// Compose renders each region's panel into its width at the given height and
// joins them left-to-right with vertical separators between.
func Compose(panels map[Region]Panel, widths map[Region]int, height int) string {
	sep := styles.Rule.Render(
		strings.Repeat("│\n", height-1) + "│",
	)
	blocks := make([]string, 0, len(regionOrder)*2-1)
	for i, region := range regionOrder {
		if i > 0 {
			blocks = append(blocks, sep)
		}
		p := panels[region]
		blocks = append(blocks, p.Style.Width(widths[region]).Height(height).Render(p.Content))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, blocks...)
}

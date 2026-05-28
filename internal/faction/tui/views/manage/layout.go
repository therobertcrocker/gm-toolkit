package manage

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Region identifies one of the columns in the manage working area. To move a
// component to a different column, assign its View to a different Region key in
// manage.View — no other change is needed; widths and separators follow.
type Region int

const (
	Left Region = iota
	Center
	Right
)

// regionOrder is the left-to-right render order used by compose.
var regionOrder = []Region{Left, Center, Right}

// panel is a region's content plus the base style that governs its alignment
// and color. compose applies the region's width and height on top.
type panel struct {
	content string
	style   lipgloss.Style
}

// regionWidths splits the working area into a wide center flanked by
// equal-width left and right columns, accounting for the two single-column
// separators between them. Remainder goes to the center.
func regionWidths(termWidth int) map[Region]int {
	totalW := max(termWidth-2, 4) // two 1-col separators
	left := totalW / 4
	right := totalW / 4
	return map[Region]int{
		Left:   left,
		Center: totalW - left - right,
		Right:  right,
	}
}

// compose renders each region's panel into its width at the given height and
// joins them left-to-right with vertical separators between.
func compose(panels map[Region]panel, widths map[Region]int, height int) string {
	sep := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(
		strings.Repeat("│\n", height-1) + "│",
	)
	blocks := make([]string, 0, len(regionOrder)*2-1)
	for i, region := range regionOrder {
		if i > 0 {
			blocks = append(blocks, sep)
		}
		p := panels[region]
		blocks = append(blocks, p.style.Width(widths[region]).Height(height).Render(p.content))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, blocks...)
}

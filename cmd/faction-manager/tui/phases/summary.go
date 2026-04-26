package phases

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/tui/style"
)

type SummaryDoneMsg struct{}

type FactionSummaryRow struct {
	Name      string
	StartHP   int
	EndHP     int
	StartCoin int
	EndCoin   int
	Action    string
}

type CycleSummaryModel struct {
	rows  []FactionSummaryRow
	cycle int
}

func NewCycleSummaryModel(cycle int, rows []FactionSummaryRow) CycleSummaryModel {
	return CycleSummaryModel{rows: rows, cycle: cycle}
}

func (m CycleSummaryModel) Init() tea.Cmd { return nil }

func (m CycleSummaryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		return m, func() tea.Msg { return SummaryDoneMsg{} }
	}
	return m, nil
}

func (m CycleSummaryModel) View() string {
	var sb strings.Builder
	sb.WriteString(style.Header.Render(fmt.Sprintf("=== Cycle %d Complete ===", m.cycle)))
	sb.WriteString("\n\n")
	nameCol := lipgloss.NewStyle().Width(22)
	deltaCol := lipgloss.NewStyle().Width(12)

	fmt.Fprintf(&sb, "%s  %s  %s  %s\n",
		style.Muted.Inherit(nameCol).Render("Faction"),
		style.Muted.Inherit(deltaCol).Render("HP"),
		style.Muted.Inherit(deltaCol).Render("Coin"),
		style.Muted.Render("Action"),
	)
	sb.WriteString(style.Muted.Render(strings.Repeat("─", 64)))
	sb.WriteString("\n")
	for _, row := range m.rows {
		action := row.Action
		if action == "" {
			action = style.Muted.Render("—")
		}
		fmt.Fprintf(&sb, "%s  %s  %s  %s\n",
			nameCol.Render(summaryTruncate(row.Name, 22)),
			summaryDelta(row.StartHP, row.EndHP, deltaCol, true),
			summaryDelta(row.StartCoin, row.EndCoin, deltaCol, false),
			action,
		)
	}
	sb.WriteString("\n")
	sb.WriteString(style.Muted.Render("Press any key to exit"))
	return sb.String()
}

func summaryDelta(start, end int, col lipgloss.Style, isHP bool) string {
	text := fmt.Sprintf("%d→%d", start, end)
	if end > start {
		return style.HP.Inherit(col).Render(text)
	}
	if end < start {
		if isHP {
			return style.LowHP.Inherit(col).Render(text)
		}
		return style.Coin.Inherit(col).Render(text)
	}
	return style.Muted.Inherit(col).Render(text)
}

func summaryTruncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

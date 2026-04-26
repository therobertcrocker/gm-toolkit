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
	Name          string
	StartHP       int
	EndHP         int
	StartCoin     int
	EndCoin       int
	Action        string
	ResultSummary string
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
	actionCol := lipgloss.NewStyle().Width(18)

	fmt.Fprintf(&sb, "%s  %s  %s  %s  %s\n",
		style.Muted.Inherit(nameCol).Render("Faction"),
		style.Muted.Inherit(deltaCol).Render("HP"),
		style.Muted.Inherit(deltaCol).Render("Coin"),
		style.Muted.Inherit(actionCol).Render("Action"),
		style.Muted.Render("Result"),
	)
	sb.WriteString(style.Muted.Render(strings.Repeat("─", 80)))
	sb.WriteString("\n")
	for _, row := range m.rows {
		var actionRendered string
		if row.Action == "" {
			actionRendered = actionCol.Render(style.Muted.Render("—"))
		} else {
			actionRendered = actionCol.Render(summaryTruncate(row.Action, 18))
		}
		result := row.ResultSummary
		if result == "" {
			result = style.Muted.Render("—")
		}
		fmt.Fprintf(&sb, "%s  %s  %s  %s  %s\n",
			nameCol.Render(summaryTruncate(row.Name, 22)),
			summaryDelta(row.StartHP, row.EndHP, deltaCol, true),
			summaryDelta(row.StartCoin, row.EndCoin, deltaCol, false),
			actionRendered,
			result,
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

package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/tui/style"
)

const leftPanelWidth = 30

func renderSplitPanel(left, right string, totalWidth int) string {
	rightWidth := max(totalWidth-leftPanelWidth-2, 10)

	leftPanel := style.PanelBorder.Width(leftPanelWidth).Render(left)
	rightPanel := style.PanelBorder.Width(rightWidth).Render(right)

	leftLines := lipgloss.Height(leftPanel)
	rightLines := lipgloss.Height(rightPanel)

	if leftLines > rightLines {
		rightPanel = style.PanelBorder.Width(rightWidth).Height(leftLines - 2).Render(right)
	} else if rightLines > leftLines {
		leftPanel = style.PanelBorder.Width(leftPanelWidth).Height(rightLines - 2).Render(left)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}

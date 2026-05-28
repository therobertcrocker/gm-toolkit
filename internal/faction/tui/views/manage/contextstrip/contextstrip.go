package contextstrip

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type Model struct {
	campaignID   string
	cycleCurrent int
	cycleNext    int
	factionCount int
}

func New(factionState *state.FactionState) Model {
	return Model{
		campaignID:   factionState.CampaignID,
		cycleCurrent: factionState.CycleNumber,
		cycleNext:    factionState.CycleNumber + 1,
		factionCount: len(factionState.Factions),
	}
}

func (m Model) View() string {
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	value := lipgloss.NewStyle().Bold(true)

	lines := []string{
		label.Render("Campaign  ") + value.Render(m.campaignID),
		label.Render("Cycle     ") + value.Render(fmt.Sprintf("%d → %d", m.cycleCurrent, m.cycleNext)),
		label.Render("Factions  ") + value.Render(fmt.Sprintf("%d", m.factionCount)),
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6C7086")).
		Padding(1, 2).
		Width(28).
		Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

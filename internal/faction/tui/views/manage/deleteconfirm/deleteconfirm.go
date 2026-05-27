package deleteconfirm

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

type Model struct {
	faction *domain.Faction
}

func New(faction *domain.Faction) Model { return Model{faction: faction} }
func (m Model) Init() tea.Cmd           { return nil }
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}
func (m Model) View() string {
	return styles.Placeholder.Render("DeleteConfirm — Commit 3")
}

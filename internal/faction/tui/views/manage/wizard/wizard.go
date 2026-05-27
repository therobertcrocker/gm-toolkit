package wizard

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type Model struct {
	rulebook     *rulebook.Rulebook
	spatialMap   *spatial.RegionMap
	factionState *state.FactionState
}

func New(rb *rulebook.Rulebook, spatialMap *spatial.RegionMap, factionState *state.FactionState) Model {
	return Model{rulebook: rb, spatialMap: spatialMap, factionState: factionState}
}
func (m Model) Init() tea.Cmd { return nil }
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}
func (m Model) View() string {
	return styles.Placeholder.Render("Wizard — Commit 4")
}

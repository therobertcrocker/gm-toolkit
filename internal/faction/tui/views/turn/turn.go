package turn

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

type Model struct{}

func New() Model                                      { return Model{} }
func (m Model) Init() tea.Cmd                         { return nil }
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m Model) View() string {
	return styles.Placeholder.Render("Turn — interactive cycle execution ships in Initiative 3 (tui-turn)")
}

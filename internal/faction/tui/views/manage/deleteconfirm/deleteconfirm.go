package deleteconfirm

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage/msgs"
)

type keyMap struct {
	Confirm key.Binding
	Cancel  key.Binding
}

type Model struct {
	faction *domain.Faction
	keys    keyMap
}

func New(faction *domain.Faction) Model {
	return Model{
		faction: faction,
		keys: keyMap{
			Confirm: key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "yes")),
			Cancel:  key.NewBinding(key.WithKeys("n", "esc"), key.WithHelp("n/esc", "no")),
		},
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(km, m.keys.Confirm):
			return m, func() tea.Msg { return msgs.DeletedMsg{ID: m.faction.ID} }
		case key.Matches(km, m.keys.Cancel):
			return m, func() tea.Msg { return msgs.CancelMsg{} }
		}
	}
	return m, nil
}

func (m Model) View() string {
	prompt := fmt.Sprintf("Delete faction %q (%s)? [y/N]", m.faction.Name, m.faction.ID)
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9")).Padding(1, 2).Render(prompt)
}

func (m Model) Help() help.KeyMap { return helpKeys{m.keys} }

type helpKeys struct{ keys keyMap }

func (h helpKeys) ShortHelp() []key.Binding {
	return []key.Binding{h.keys.Confirm, h.keys.Cancel}
}
func (h helpKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{{h.keys.Confirm, h.keys.Cancel}}
}

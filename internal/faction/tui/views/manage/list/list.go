package list

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage/msgs"
)

type item struct {
	id    string
	name  string
	scale string
	hp    string
	coin  string
}

func (i item) Title() string       { return i.name }
func (i item) Description() string { return fmt.Sprintf("%s · HP %s · Coin %s", i.scale, i.hp, i.coin) }
func (i item) FilterValue() string { return i.name }

type keyMap struct {
	New    key.Binding
	Select key.Binding
}

type Model struct {
	list  list.Model
	keys  keyMap
	state *state.FactionState
}

func New(factionState *state.FactionState) Model {
	items := buildItems(factionState)
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Factions"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	return Model{
		list:  l,
		state: factionState,
		keys: keyMap{
			New:    key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
			Select: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		},
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.New):
			return m, func() tea.Msg { return msgs.RequestCreateMsg{} }
		case key.Matches(msg, m.keys.Select):
			if it, ok := m.list.SelectedItem().(item); ok {
				return m, func() tea.Msg { return msgs.RequestDetailMsg{FactionID: it.id} }
			}
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if len(m.state.Factions) == 0 {
		return emptyState()
	}
	return m.list.View()
}

func (m Model) Help() help.KeyMap { return helpKeys{m.keys} }

type helpKeys struct{ keys keyMap }

func (h helpKeys) ShortHelp() []key.Binding { return []key.Binding{h.keys.New, h.keys.Select} }
func (h helpKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{{h.keys.New, h.keys.Select}}
}

func buildItems(factionState *state.FactionState) []list.Item {
	ids := make([]string, 0, len(factionState.Factions))
	for id := range factionState.Factions {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	items := make([]list.Item, 0, len(ids))
	for _, id := range ids {
		f := factionState.Factions[id]
		items = append(items, item{
			id:    id,
			name:  f.Name,
			scale: string(f.Scale),
			hp:    fmt.Sprintf("%d/%d", f.CurrentHP, f.MaxHP),
			coin:  fmt.Sprintf("%d", f.Coin),
		})
	}
	return items
}

func emptyState() string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.NewStyle().Render("No factions yet."),
		styles.Dim.Render("Press n to create one"),
	)
}

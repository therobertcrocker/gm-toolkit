package phases

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
)

type ActionSelectedMsg struct{ Index int }

type actionItem struct {
	name string
	idx  int
}

func (i actionItem) Title() string       { return i.name }
func (i actionItem) Description() string { return "" }
func (i actionItem) FilterValue() string { return i.name }

type ActionSelectModel struct {
	list    list.Model
	actions []engine.Action
}

func NewActionSelectModel(actions []engine.Action) ActionSelectModel {
	items := make([]list.Item, 0, len(actions)+1)
	for i, action := range actions {
		items = append(items, actionItem{name: action.Name(), idx: i})
	}
	items = append(items, actionItem{name: "No Action", idx: -1})

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Select Action"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()

	return ActionSelectModel{list: l, actions: actions}
}

func (m ActionSelectModel) Init() tea.Cmd { return nil }

func (m ActionSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-2)
		return m, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			if item, ok := m.list.SelectedItem().(actionItem); ok {
				return m, func() tea.Msg { return ActionSelectedMsg{Index: item.idx} }
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ActionSelectModel) View() string {
	return m.list.View()
}

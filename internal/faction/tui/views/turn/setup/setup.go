package setup

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/layout"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/turn/msgs"
)

type item struct {
	name, scale, hp, coin string
}

func (i item) Title() string { return i.name }
func (i item) Description() string {
	return fmt.Sprintf("%s · HP %s · Coin %s", i.scale, i.hp, i.coin)
}
func (i item) FilterValue() string { return i.name }

type keyMap struct {
	Start key.Binding
}

type Model struct {
	factionState *state.FactionState
	list         list.Model
	keys         keyMap
	termWidth    int
	termHeight   int
}

func New(factionState *state.FactionState) Model {
	l := list.New(buildItems(factionState), list.NewDefaultDelegate(), 0, 0)
	l.Title = "Factions"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	return Model{
		factionState: factionState,
		list:         l,
		keys: keyMap{
			Start: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "start cycle")),
		},
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Start) && len(m.factionState.Factions) > 0 {
			return m, func() tea.Msg { return msgs.StartCycleMsg{} }
		}
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		m.list.SetSize(layout.RegionWidths(msg.Width)[layout.Left], msg.Height)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	contentH := max(m.termHeight, 1)
	placeholder := styles.Dim.Align(lipgloss.Center).AlignVertical(lipgloss.Center)

	leftContent := m.list.View()
	if len(m.factionState.Factions) == 0 {
		leftContent = emptyState()
	}
	panels := map[layout.Region]layout.Panel{
		layout.Left:   {Content: leftContent, Style: lipgloss.NewStyle()},
		layout.Center: {Content: "—", Style: placeholder},
		layout.Right:  {Content: "—", Style: placeholder},
	}
	return layout.Compose(panels, layout.RegionWidths(m.termWidth), contentH)
}

func (m Model) Help() help.KeyMap { return helpKeys{m.keys} }

type helpKeys struct{ keys keyMap }

func (h helpKeys) ShortHelp() []key.Binding  { return []key.Binding{h.keys.Start} }
func (h helpKeys) FullHelp() [][]key.Binding { return [][]key.Binding{{h.keys.Start}} }

// buildItems projects each faction in factionState into a list.Item, sorted
// alphabetically by faction ID.
func buildItems(factionState *state.FactionState) []list.Item {
	factionIDs := make([]string, 0, len(factionState.Factions))
	for id := range factionState.Factions {
		factionIDs = append(factionIDs, id)
	}
	sort.Strings(factionIDs)

	items := make([]list.Item, len(factionIDs))
	for i, id := range factionIDs {
		faction := factionState.Factions[id]
		items[i] = item{
			name:  faction.Name,
			scale: string(faction.Scale),
			hp:    fmt.Sprintf("%d/%d", faction.CurrentHP, faction.MaxHP),
			coin:  fmt.Sprintf("%d", faction.Coin),
		}
	}
	return items
}

func emptyState() string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.NewStyle().Render("No factions yet."),
		styles.Dim.Render("Create factions in Manage (tab), then return here to run a cycle."),
	)
}

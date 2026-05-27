package detail

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage/msgs"
)

type keyMap struct {
	Delete key.Binding
	Back   key.Binding
}

type Model struct {
	faction *domain.Faction
	keys    keyMap
}

func New(faction *domain.Faction) Model {
	return Model{
		faction: faction,
		keys: keyMap{
			Delete: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
			Back:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		},
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(km, m.keys.Delete):
			return m, func() tea.Msg { return msgs.RequestDeleteMsg{Faction: m.faction} }
		case key.Matches(km, m.keys.Back):
			return m, func() tea.Msg { return msgs.CancelMsg{} }
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.faction == nil {
		return "(no faction)"
	}
	var b strings.Builder
	f := m.faction

	fmt.Fprintf(&b, "%s  (%s)\n", f.Name, f.ID)
	fmt.Fprintf(&b, "Scale: %s\n", f.Scale)
	fmt.Fprintf(&b, "Force: %d  Cunning: %d  Wealth: %d\n", f.Force, f.Cunning, f.Wealth)
	fmt.Fprintf(&b, "HP: %d/%d   Coin: %d   XP: %d\n", f.CurrentHP, f.MaxHP, f.Coin, f.XP)
	fmt.Fprintf(&b, "Homeworld: %s\n", f.Homeworld.WorldID)

	fmt.Fprintf(&b, "\nTags:\n")
	for _, tag := range f.Tags {
		fmt.Fprintf(&b, "  - %s\n", tag.Name)
	}

	if f.ActiveGoal != nil {
		fmt.Fprintf(&b, "\nActive Goal: %s (progress %d)\n", f.ActiveGoal.GoalID, f.ActiveGoal.Progress)
	}

	fmt.Fprintf(&b, "\nAssets:\n")
	for _, asset := range domain.SortedAssets(f) {
		fmt.Fprintf(&b, "  - %s\n", asset.ID)
	}

	fmt.Fprintf(&b, "\nBases:\n")
	for _, base := range f.Bases {
		fmt.Fprintf(&b, "  - %s\n", base.Location.WorldID)
	}

	return b.String()
}

func (m Model) Help() help.KeyMap { return helpKeys{m.keys} }

type helpKeys struct{ keys keyMap }

func (h helpKeys) ShortHelp() []key.Binding {
	return []key.Binding{h.keys.Delete, h.keys.Back}
}
func (h helpKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{{h.keys.Delete, h.keys.Back}}
}

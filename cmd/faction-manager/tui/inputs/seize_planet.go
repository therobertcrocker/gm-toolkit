package inputs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/tui/style"
)

type SeizePlanetTargetSelectedMsg struct{ World string }

type SeizePlanetModel struct {
	worlds      []string
	worldCursor int
}

func NewSeizePlanetModel(worlds []string) SeizePlanetModel {
	return SeizePlanetModel{worlds: worlds}
}

func (m SeizePlanetModel) Init() tea.Cmd { return nil }

func (m SeizePlanetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "up", "k":
		if m.worldCursor > 0 {
			m.worldCursor--
		}
	case "down", "j":
		if m.worldCursor < len(m.worlds)-1 {
			m.worldCursor++
		}
	case "enter":
		if len(m.worlds) == 0 {
			break
		}
		world := m.worlds[m.worldCursor]
		return m, func() tea.Msg { return SeizePlanetTargetSelectedMsg{World: world} }
	}
	return m, nil
}

func (m SeizePlanetModel) View() string {
	var sb strings.Builder
	sb.WriteString(style.SectionTitle.Render("Seize Planet — Select Target World"))
	sb.WriteString("\n\n")
	for i, world := range m.worlds {
		cursor := "  "
		if i == m.worldCursor {
			cursor = "> "
		}
		fmt.Fprintf(&sb, "%s%s\n", cursor, world)
	}
	sb.WriteString("\n")
	sb.WriteString(style.Muted.Render("↑/↓ navigate  Enter select"))
	return sb.String()
}

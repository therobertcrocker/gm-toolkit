package inputs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/tui/style"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
)

type AbilityMoveDestinationSelectedMsg struct {
	Destination string
}

type MoveDestinationModel struct {
	assetName string
	worlds    []string
	cursor    int
}

func NewMoveDestinationModel(asset *domain.Asset, worlds []string, rulebook *loader.Rulebook) MoveDestinationModel {
	name := asset.DefinitionID
	if def, ok := rulebook.Assets[asset.DefinitionID]; ok {
		name = def.Name
	}
	return MoveDestinationModel{assetName: name, worlds: worlds}
}

func (m MoveDestinationModel) Init() tea.Cmd { return nil }

func (m MoveDestinationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.worlds)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.worlds) == 0 {
			return m, nil
		}
		dest := m.worlds[m.cursor]
		return m, func() tea.Msg { return AbilityMoveDestinationSelectedMsg{Destination: dest} }
	}
	return m, nil
}

func (m MoveDestinationModel) View() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n", style.SectionTitle.Render("Move Asset — Select Destination"))
	fmt.Fprintf(&sb, "%s\n\n", style.Muted.Render(m.assetName))
	for i, world := range m.worlds {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		fmt.Fprintf(&sb, "%s%s\n", cursor, world)
	}
	sb.WriteString("\n")
	sb.WriteString(style.Muted.Render("↑/↓ navigate  Enter confirm"))
	return sb.String()
}

package phases

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/tui/style"
)

type SkipChoiceMsg struct{ Skip bool }

type SkipTurnModel struct {
	factionName string
}

func NewSkipTurnModel(factionName string) SkipTurnModel {
	return SkipTurnModel{factionName: factionName}
}

func (m SkipTurnModel) Init() tea.Cmd { return nil }

func (m SkipTurnModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "s", "S":
		return m, func() tea.Msg { return SkipChoiceMsg{Skip: true} }
	case "enter", "n", "N":
		return m, func() tea.Msg { return SkipChoiceMsg{Skip: false} }
	}
	return m, nil
}

func (m SkipTurnModel) View() string {
	return fmt.Sprintf(
		"%s\n\n  [Enter] Start turn   [s] Skip",
		style.Header.Render(m.factionName+"'s Turn"),
	)
}

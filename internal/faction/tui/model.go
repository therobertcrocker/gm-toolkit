package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/modebar"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/turn"
)

type Model struct {
	bar     modebar.Model
	subs    map[modebar.Mode]tea.Model
	adapter *adapter.Adapter

	confirmExit  bool
	priorMode    modebar.Mode
	showHelpStub bool
}

func NewModel(adp *adapter.Adapter) Model {
	bar := modebar.New()
	return Model{
		bar:     bar,
		adapter: adp,
		subs: map[modebar.Mode]tea.Model{
			modebar.ModeManage: manage.New(),
			modebar.ModeTurn:   turn.New(),
		},
	}
}

func (m Model) Init() tea.Cmd { return nil }

package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/modebar"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/turn"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type Model struct {
	bar     modebar.Model
	subs    map[modebar.Mode]tea.Model
	adapter *adapter.Adapter

	confirmExit  bool
	priorMode    modebar.Mode
	showHelpStub bool
}

func NewModel(
	adp *adapter.Adapter,
	factionState *state.FactionState,
	paths *campaigns.Paths,
	rb *rulebook.Rulebook,
	spatialMap *spatial.RegionMap,
) Model {
	return Model{
		bar:     modebar.New(),
		adapter: adp,
		subs: map[modebar.Mode]tea.Model{
			modebar.ModeManage: manage.New(factionState, paths, rb, spatialMap),
			modebar.ModeTurn:   turn.New(),
		},
	}
}

func (m Model) Init() tea.Cmd { return nil }

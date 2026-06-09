package tui

import (
	"log/slog"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/modebar"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/turn"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type Model struct {
	bar          modebar.Model
	subs         map[modebar.Mode]tea.Model
	help         help.Model
	factionState *state.FactionState
	campaignID   string
	width        int
	height       int

	confirmExit bool
}

func NewModel(
	eng *engine.Engine,
	factionState *state.FactionState,
	paths *campaigns.Paths,
	rb *rulebook.Rulebook,
	spatialMap *spatial.RegionMap,
	log *slog.Logger,
) Model {
	h := help.New()
	h.ShowAll = false
	return Model{
		bar:          modebar.New(),
		help:         h,
		factionState: factionState,
		campaignID:   factionState.CampaignID,
		subs: map[modebar.Mode]tea.Model{
			modebar.ModeManage: manage.New(factionState, paths, rb, spatialMap),
			modebar.ModeTurn:   turn.New(factionState, paths, eng, rb, spatialMap, log),
		},
	}
}

func (m Model) Init() tea.Cmd { return nil }

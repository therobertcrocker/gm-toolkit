package tui

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

func Run(
	eng *engine.Engine,
	factionState *state.FactionState,
	paths *campaigns.Paths,
	rb *rulebook.Rulebook,
	spatialMap *spatial.RegionMap,
	log *slog.Logger,
) error {
	adp := adapter.New(eng, factionState, paths, log)
	defer adp.Stop()

	program := tea.NewProgram(NewModel(adp, factionState, paths, rb, spatialMap), tea.WithAltScreen())
	_, err := program.Run()
	return err
}

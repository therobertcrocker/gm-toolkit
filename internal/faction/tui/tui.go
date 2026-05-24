package tui

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/config"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
)

func Run(eng *engine.Engine, factionState *state.FactionState, cfg *config.Config, log *slog.Logger) error {
	adp := adapter.New(eng, factionState, cfg, log)
	defer adp.Stop()

	program := tea.NewProgram(NewModel(adp), tea.WithAltScreen())
	_, err := program.Run()
	return err
}

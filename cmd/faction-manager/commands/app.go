package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/commands/faction"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
)

type App struct {
	Engine      *engine.Engine
	dataDir     string
}

func NewApp(dataDir string) *App {
	return &App{dataDir: dataDir}
}

func (a *App) Execute() error {
	e, err := engine.New(a.dataDir)
	if err != nil {
		return fmt.Errorf("initializing engine: %w", err)
	}
	a.Engine = e

	root := &cobra.Command{
		Use:   "faction-manager",
		Short: "A faction tracker inspired by the mechanics of Stars Without Number",
	}

	root.AddCommand(a.reviewCmd())
	root.AddCommand(a.turnCmd())
	root.AddCommand(faction.NewCmd(a.Engine.Rulebook))
	root.AddCommand(a.narrateCmd())

	return root.Execute()
}

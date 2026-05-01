package commands

import (
	"github.com/spf13/cobra"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/commands/narrate"
)

func (a *App) narrateCmd() *cobra.Command {
	return narrate.NewCmd(a.Engine.Rulebook)
}

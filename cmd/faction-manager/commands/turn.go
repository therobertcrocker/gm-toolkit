package commands

import (
	"github.com/spf13/cobra"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/commands/turn"
)

func (a *App) turnCmd() *cobra.Command {
	return turn.NewCmd(a.Engine.Turn)
}

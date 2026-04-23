package commands

import (
	"github.com/spf13/cobra"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/commands/turn"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/forms"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
)

func (a *App) turnCmd() *cobra.Command {
	collector := forms.NewGMCollector()
	a.Engine.Action.Register(func() engine.Action { return engine.NewSellAsset(collector) })

	return turn.NewCmd(a.Engine)
}

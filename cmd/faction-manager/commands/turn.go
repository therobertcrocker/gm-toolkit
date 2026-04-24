package commands

import (
	"github.com/spf13/cobra"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/commands/turn"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/forms"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/actions"
)

func (a *App) turnCmd() *cobra.Command {
	collector := forms.NewGMCollector()
	a.Engine.Action.Register(func() engine.Action { return actions.NewSellAsset(collector) })
	a.Engine.Action.Register(func() engine.Action { return actions.NewRepairFaction() })
	a.Engine.Action.Register(func() engine.Action { return actions.NewRepairAsset(collector) })
	a.Engine.Action.Register(func() engine.Action { return actions.NewBuyAsset(collector) })
	a.Engine.Action.Register(func() engine.Action { return actions.NewRefitAsset(collector) })

	return turn.NewCmd(a.Engine)
}

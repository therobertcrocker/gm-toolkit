package commands

import (
	"github.com/spf13/cobra"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/commands/turn"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/actions"
)

func (a *App) turnCmd() *cobra.Command {
	a.Engine.Action.Register(func() engine.Action { return actions.NewSellAsset(nil) })
	a.Engine.Action.Register(func() engine.Action { return actions.NewRepairFaction() })
	a.Engine.Action.Register(func() engine.Action { return actions.NewRepairAsset(nil) })
	a.Engine.Action.Register(func() engine.Action { return actions.NewBuyAsset(nil) })
	a.Engine.Action.Register(func() engine.Action { return actions.NewRefitAsset(nil) })
	a.Engine.Action.Register(func() engine.Action { return actions.NewAttack(nil, nil) })
	a.Engine.Action.Register(func() engine.Action { return actions.NewExpandInfluence(nil, nil) })
	a.Engine.Action.Register(func() engine.Action { return actions.NewUseAssetAbility(nil, nil, nil) })

	return turn.NewCmd(a.Engine)
}

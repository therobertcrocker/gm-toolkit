package actions

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
)

// RegisterDefaultActions wires the standard SWN faction actions into the engine.
func RegisterDefaultActions(e *engine.Engine) {
	e.Action.Register(func(c action.Collector) action.Action { return NewSellAsset(c) })
	e.Action.Register(func(_ action.Collector) action.Action { return NewRepairFaction() })
	e.Action.Register(func(c action.Collector) action.Action { return NewRepairAsset(c) })
	e.Action.Register(func(c action.Collector) action.Action { return NewBuyAsset(c) })
	e.Action.Register(func(c action.Collector) action.Action { return NewRefitAsset(c) })
	e.Action.Register(func(c action.Collector) action.Action { return NewAttack(c, e.Rand) })
	e.Action.Register(func(c action.Collector) action.Action { return NewExpandInfluence(c, e.Rand) })
	e.Action.Register(func(c action.Collector) action.Action { return NewBribe(c) })
	e.Action.Register(func(c action.Collector) action.Action {
		return NewUseAssetAbility(c.(engine.InputCollector), e.Rand, e.Ability)
	})
	e.Action.Register(func(_ action.Collector) action.Action { return NewAbandonGoal() })
	e.Action.Register(func(c action.Collector) action.Action { return NewSeizePlanet(c) })
}

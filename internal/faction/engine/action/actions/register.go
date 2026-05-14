package actions

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
)

// RegisterDefaultActions wires the standard SWN faction actions into the engine.
func RegisterDefaultActions(e *engine.Engine) {
	e.Action.Register(func(c action.Collector) action.Action { return NewSellAsset(c) })
	e.Action.Register(func(_ action.Collector) action.Action { return NewRepairFaction() })
	e.Action.Register(func(c action.Collector) action.Action { return NewRepairAsset(c) })
	e.Action.Register(func(c action.Collector) action.Action { return NewBuyAsset(c, e.Hooks, e.World) })
	e.Action.Register(func(c action.Collector) action.Action { return NewRefitAsset(c) })
	e.Action.Register(func(c action.Collector) action.Action { return NewAttack(c, e.Rand, e.Hooks, worldIndex(e)) })
	e.Action.Register(func(c action.Collector) action.Action { return NewExpandInfluence(c, e.Rand, worldIndex(e), e.World) })
	e.Action.Register(func(c action.Collector) action.Action { return NewBribe(c) })
	e.Action.Register(func(c action.Collector) action.Action {
		return NewUseAssetAbility(c, e.Rand, e.Ability)
	})
	e.Action.Register(func(_ action.Collector) action.Action { return NewAbandonGoal() })
	e.Action.Register(func(c action.Collector) action.Action { return NewSeizePlanet(c, worldIndex(e)) })
}

func worldIndex(e *engine.Engine) *world.Index {
	if e.World == nil {
		return nil
	}
	return e.World.Index
}

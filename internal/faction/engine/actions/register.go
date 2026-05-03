package actions

import "github.com/therobertcrocker/gm-toolkit/internal/faction/engine"

// RegisterDefaultActions wires the standard SWN faction actions into the
// engine's ActionEngine. Each factory closes over engine collaborators
// (Rand, AbilityEngine) so callers only supply an InputCollector at
// AvailableActions time.
func RegisterDefaultActions(e *engine.Engine) {
	e.Action.Register(func(c engine.InputCollector) engine.Action { return NewSellAsset(c) })
	e.Action.Register(func(_ engine.InputCollector) engine.Action { return NewRepairFaction() })
	e.Action.Register(func(c engine.InputCollector) engine.Action { return NewRepairAsset(c) })
	e.Action.Register(func(c engine.InputCollector) engine.Action { return NewBuyAsset(c) })
	e.Action.Register(func(c engine.InputCollector) engine.Action { return NewRefitAsset(c) })
	e.Action.Register(func(c engine.InputCollector) engine.Action { return NewAttack(c, e.Rand) })
	e.Action.Register(func(c engine.InputCollector) engine.Action { return NewExpandInfluence(c, e.Rand) })
	e.Action.Register(func(c engine.InputCollector) engine.Action { return NewBribe(c) })
	e.Action.Register(func(c engine.InputCollector) engine.Action { return NewUseAssetAbility(c, e.Rand, e.AbilityEngine) })
	e.Action.Register(func(_ engine.InputCollector) engine.Action { return NewAbandonGoal() })
	e.Action.Register(func(c engine.InputCollector) engine.Action { return NewSeizePlanet(c) })
}

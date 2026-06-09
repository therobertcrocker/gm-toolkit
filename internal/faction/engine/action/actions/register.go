package actions

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

// RegisterDefaultActions wires the standard SWN faction actions into the action
// engine. Dependencies are passed explicitly so this package does not import the
// parent engine package — that keeps engine free to own registration without an
// import cycle.
// resolveRoller is read at factory-invocation (turn) time rather than captured
// by value, because the roller is a swappable seam — tests replace Engine.Rand
// with a deterministic FixedRoller after construction.
func RegisterDefaultActions(
	ae *action.ActionEngine,
	resolveRoller func() domain.Roller,
	hookRegistry *hooks.Registry,
	worldEngine *world.WorldEngine,
	rulebook *rulebook.Rulebook,
) {
	ae.Register(func(c action.Collector) action.Action { return NewSellAsset(c) })
	ae.Register(func(_ action.Collector) action.Action { return NewRepairFaction() })
	ae.Register(func(c action.Collector) action.Action { return NewRepairAsset(c) })
	ae.Register(func(c action.Collector) action.Action { return NewBuyAsset(c, hookRegistry, worldEngine) })
	ae.Register(func(c action.Collector) action.Action { return NewRefitAsset(c) })
	ae.Register(func(c action.Collector) action.Action {
		return NewAttack(c, resolveRoller(), hookRegistry, worldIndex(worldEngine))
	})
	ae.Register(func(c action.Collector) action.Action {
		return NewExpandInfluence(c, resolveRoller(), worldIndex(worldEngine), worldEngine)
	})
	ae.Register(func(c action.Collector) action.Action { return NewBribe(c) })
	ae.Register(func(c action.Collector) action.Action {
		return NewUseAssetAbility(c, resolveRoller())
	})
	ae.Register(func(_ action.Collector) action.Action { return NewAbandonGoal() })
	ae.Register(func(c action.Collector) action.Action { return NewSeizePlanet(c, worldIndex(worldEngine)) })
	ae.Register(func(c action.Collector) action.Action { return NewChangeHomeworld(c, worldEngine, rulebook) })
}

// worldIndex returns the spatial index the world-aware actions need, tolerating
// a World-less engine (dryrun smoke worlds, non-spatial actions).
func worldIndex(worldEngine *world.WorldEngine) *world.Index {
	if worldEngine == nil {
		return nil
	}
	return worldEngine.Index
}

package engine

import (
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func (combatant *Combatant) OpenActions() int {
	return combatant.ActionsLeft - combatant.StrikeOnlyActions
}

func (fight *Fight) regainActions(actor *Combatant) {
	actor.extraReactions = maps.Clone(actor.Creature.ExtraReactions)
	actions := actionsPerTurn
	quickened := actor.Condition("quickened") > 0
	if quickened {
		actions++
	}
	stunned := min(actor.Condition("stunned"), actions)
	// deliberate: actions lost to stunned count toward slowed, not on top of it, because the stunned condition says they do.
	lost := min(max(stunned, actor.Condition("slowed")), actions)
	if actor.stunnedForDuration() {
		lost = actions
	}
	if stunned > 0 {
		fight.reduceCondition(actor, "stunned", stunned)
	}
	actor.ActionsLeft = actions - lost
	actor.StrikeOnlyActions = 0
	// deliberate: a creature that loses an action loses the quickened one first, not a free one, because it picks which action it gives up.
	if quickened && lost == 0 {
		actor.StrikeOnlyActions = 1
	}
	if quickened || lost > 0 {
		fight.observer.Observe(ActionsRegained{Combatant: actor.ID, Actions: actor.ActionsLeft, Quickened: quickened, Lost: lost})
	}
}

func (combatant *Combatant) stunnedForDuration() bool {
	return slices.ContainsFunc(combatant.conditions, func(instance *conditionInstance) bool {
		return instance.name == "stunned" && instance.value == 0
	})
}

func (fight *Fight) loseActionsToStunned(bearer *Combatant) {
	stunned := min(bearer.Condition("stunned"), bearer.ActionsLeft)
	lost := stunned
	if bearer.stunnedForDuration() {
		lost = bearer.ActionsLeft
	}
	if lost == 0 {
		return
	}
	fight.reduceCondition(bearer, "stunned", stunned)
	bearer.ActionsLeft -= lost
	bearer.StrikeOnlyActions = max(bearer.StrikeOnlyActions-lost, 0)
	fight.observer.Observe(ActionsLost{Combatant: bearer.ID, Condition: "stunned", Lost: lost, Left: bearer.ActionsLeft})
}

func (fight *Fight) stand(actor *Combatant) {
	fight.changeConditions(actor, func() {
		actor.conditions = slices.DeleteFunc(actor.conditions, func(instance *conditionInstance) bool {
			return instance.name == "prone"
		})
	})
	fight.observer.Observe(StoodUp{Combatant: actor.ID})
	if actor.downed {
		actor.downed = false
		fight.fire(actor, TriggerEvent{Kind: creature.TriggerStoodUp})
	}
}

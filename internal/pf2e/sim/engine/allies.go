package engine

import (
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

type recoveryCap struct {
	dc    int
	clock timing
}

type sustainedEffect struct {
	target *Combatant
	done   bool
}

func (fight *Fight) Allies(actor *Combatant) []*Combatant {
	var allies []*Combatant
	for _, combatant := range fight.combatants {
		if combatant.Side == actor.Side && !combatant.Dead {
			allies = append(allies, combatant)
		}
	}
	return allies
}

// deliberate: an area catches the whole side it is for, not a share of it, because every creature counts as engaged with every other until A.004.2's position (E3 Decision 1, Decision 13).
func (fight *Fight) areaTargets(actor *Combatant, ability creature.Ability) []*Combatant {
	switch ability.Targets {
	case creature.TargetsAlly:
		return fight.Allies(actor)
	case creature.TargetsOthers:
		return slices.DeleteFunc(fight.Allies(actor), func(ally *Combatant) bool { return ally == actor })
	default:
		return fight.StandingFoes(actor)
	}
}

func targetID(target *Combatant) string {
	if target == nil {
		return ""
	}
	return target.ID
}

func (fight *Fight) checkTarget(actor *Combatant, ability creature.Ability, target *Combatant) bool {
	switch {
	case ability.Targets == creature.TargetsSelf, ability.Area != nil && ability.Targets != "":
		return true
	case ability.Targets == creature.TargetsAlly:
		return target != nil && !target.Dead && target.Side == actor.Side
	default:
		return target != nil && target.Standing()
	}
}

func (fight *Fight) easeRecovery(source effectSource, target *Combatant, step creature.Step) {
	if target.Dead {
		return
	}
	target.recoveryCaps = append(target.recoveryCaps, &recoveryCap{dc: step.RecoveryDC, clock: fight.newTiming(source.actor, target, step.Duration)})
	fight.observer.Observe(RecoveryEased{Combatant: target.ID, Ability: source.ability, DC: step.RecoveryDC})
}

func (combatant *Combatant) Sustaining() []string {
	var waiting []string
	for _, id := range slices.Sorted(maps.Keys(combatant.sustained)) {
		if !combatant.sustained[id].done {
			waiting = append(waiting, id)
		}
	}
	return waiting
}

// deliberate: an effect ends at the end of a turn with no Sustain, not when the next turn starts, because "sustained up to" lasts to the end of the caster's next turn only if it is sustained there.
func (fight *Fight) endSustains(actor *Combatant) {
	for _, id := range slices.Sorted(maps.Keys(actor.sustained)) {
		effect := actor.sustained[id]
		if effect.done {
			effect.done = false
			continue
		}
		delete(actor.sustained, id)
		fight.observer.Observe(SustainEnded{Combatant: actor.ID, Ability: id})
	}
}

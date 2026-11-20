package engine

import (
	"fmt"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

const maxTriggerDepth = 16

type TriggerEvent struct {
	Kind      string
	By        *Combatant
	Ability   string
	Strike    string
	Traits    []string
	Degree    Degree
	IsStrike  bool
	Melee     bool
	Ranged    bool
	Damaging  bool
	Condition string
}

func matches(owner *Combatant, trigger creature.Trigger, event TriggerEvent) bool {
	if trigger.On != event.Kind {
		return false
	}
	switch event.Kind {
	case creature.TriggerTargeted, creature.TriggerAttacked:
		if trigger.Attack != "" && !event.IsStrike {
			return false
		}
		if trigger.Attack == creature.AttackMeleeStrike && !event.Melee {
			return false
		}
		return degreeListed(trigger.Degrees, event.Degree)
	case creature.TriggerSaved:
		return (!trigger.Damaging || event.Damaging) && degreeListed(trigger.Degrees, event.Degree)
	case creature.TriggerDamaged:
		return !trigger.Melee || event.Melee
	case creature.TriggerTurnStart:
		return trigger.HPAtMost == 0 || owner.HP <= trigger.HPAtMost
	case creature.TriggerFoeActs:
		return (trigger.Ranged && event.Ranged) || overlaps(trigger.Traits, event.Traits)
	case creature.TriggerAllyGains:
		return slices.Contains(trigger.Conditions, event.Condition)
	}
	return true
}

func (fight *Fight) setFault(err error) {
	if fight.fault == nil {
		fight.fault = err
	}
}

func degreeListed(words []string, degree Degree) bool {
	return len(words) == 0 || slices.Contains(words, creature.DegreeWords[degree])
}

func (combatant *Combatant) reactionFor(abilityID string) bool {
	return combatant.extraReactions[abilityID] > 0 || !combatant.ReactionUsed
}

// deliberate: an ability's own extra reaction is spent before the shared one, not after, because Tactical Reflexes exists to keep the shared reaction free.
func (combatant *Combatant) spendReaction(abilityID string) {
	if combatant.extraReactions[abilityID] > 0 {
		combatant.extraReactions[abilityID]--
		return
	}
	combatant.ReactionUsed = true
}

func (fight *Fight) fire(owner *Combatant, event TriggerEvent) {
	if fight.triggerDepth >= maxTriggerDepth {
		fight.setFault(fmt.Errorf("%w: %s: %s nested more than %d deep", ErrTriggerLoop, owner.ID, event.Kind, maxTriggerDepth))
		return
	}
	fight.triggerDepth++
	defer func() { fight.triggerDepth-- }()
	if event.By != nil && !event.By.Standing() {
		return
	}
	var offered []string
	for _, id := range owner.Creature.AbilityIDs() {
		if !owner.Standing() {
			return
		}
		ability := owner.Creature.Abilities[id]
		if ability.Ignore || ability.Trigger == nil || !matches(owner, *ability.Trigger, event) || owner.Blocked(ability) != "" {
			continue
		}
		switch {
		case ability.Trigger.Cost == "":
			fight.runTriggered(owner, ability, event)
		case owner.Condition("stunned") > 0:
		case ability.Trigger.Cost == creature.CostFree || owner.reactionFor(id):
			offered = append(offered, id)
		}
	}
	if len(offered) == 0 || !owner.Standing() {
		return
	}
	picked, reacting := fight.decider.React(fight, owner, event, offered)
	if !reacting {
		return
	}
	if !slices.Contains(offered, picked) {
		fight.setFault(fmt.Errorf("%w: %s reacted with %q, which was not offered", ErrBadChoice, owner.ID, picked))
		return
	}
	ability := owner.Creature.Abilities[picked]
	if ability.Trigger.Cost == creature.CostReaction {
		owner.spendReaction(picked)
	}
	fight.runTriggered(owner, ability, event)
}

func (fight *Fight) runTriggered(owner *Combatant, ability creature.Ability, event TriggerEvent) {
	target, by := event.By, ""
	if target != nil {
		by = target.ID
	} else {
		next, found := fight.decider.NextTarget(fight, owner)
		if !found {
			return
		}
		target = next
	}
	fight.observer.Observe(Triggered{
		Combatant: owner.ID,
		Ability:   ability.ID,
		On:        event.Kind,
		Cost:      ability.Trigger.Cost,
		By:        by,
		ByAbility: event.Ability,
		ByStrike:  event.Strike,
	})
	fight.spend(owner, ability)
	fight.useAbility(abilityUse{actor: owner, ability: ability, target: target, fixed: event.By != nil, triggeredBy: by})
}

type pendingAction struct {
	actor     *Combatant
	action    string
	traits    []string
	disrupted bool
}

// deliberate: every standing foe may answer, not only those in reach, because every creature counts as engaged with every foe until A.004.2's position (Decision 12).
func (fight *Fight) provoke(actor *Combatant, action string, event TriggerEvent) bool {
	outer := fight.pending
	pending := &pendingAction{actor: actor, action: action, traits: event.Traits}
	fight.pending = pending
	event.Kind, event.By = creature.TriggerFoeActs, actor
	for _, foe := range fight.StandingFoes(actor) {
		fight.fire(foe, event)
	}
	fight.pending = outer
	return !pending.disrupted && actor.Standing()
}

func (fight *Fight) disrupt(attacker *Combatant, ability string, traits []string) {
	pending := fight.pending
	if pending == nil || pending.disrupted || !overlaps(traits, pending.traits) {
		return
	}
	pending.disrupted = true
	fight.observer.Observe(ActionDisrupted{Combatant: pending.actor.ID, Action: pending.action, By: attacker.ID, Ability: ability})
}

func (fight *Fight) gained(bearer *Combatant, condition string) {
	for _, ally := range fight.combatants {
		if ally != bearer && ally.Side == bearer.Side {
			fight.fire(ally, TriggerEvent{Kind: creature.TriggerAllyGains, By: bearer, Condition: condition})
		}
	}
}

func (fight *Fight) fastHeal(actor *Combatant) {
	amount := actor.Creature.FastHealing
	if amount > 0 && !actor.Dead && actor.HP < actor.MaxHP() {
		fight.heal(effectSource{actor: actor, ability: "fast-healing"}, actor, amount)
	}
}

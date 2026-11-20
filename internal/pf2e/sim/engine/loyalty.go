package engine

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

type grantedTrait struct {
	trait   string
	against *Combatant
	uses    int
	clock   timing
}

func answersLoyalty(foe *Combatant) bool {
	return slices.ContainsFunc(foe.Creature.AbilityIDs(), func(id string) bool {
		ability := foe.Creature.Abilities[id]
		return !ability.Ignore && ability.Trigger != nil && ability.Trigger.On == creature.TriggerLoyalty
	})
}

func (fight *Fight) loyalty(actor *Combatant, reason string) {
	// deliberate: once per creature per initiative round, not per turn of that creature, because Robert ruled the round (Decision 7). The mark is round+1 so the zero value means "never".
	if !actor.Standing() || actor.loyalRound == fight.round+1 {
		return
	}
	var answering []*Combatant
	for _, foe := range fight.StandingFoes(actor) {
		if answersLoyalty(foe) {
			answering = append(answering, foe)
		}
	}
	if len(answering) == 0 {
		return
	}
	actor.loyalRound = fight.round + 1
	fight.observer.Observe(LoyaltyAct{Combatant: actor.ID, Reason: reason})
	for _, foe := range answering {
		fight.fire(foe, TriggerEvent{Kind: creature.TriggerLoyalty, By: actor, Ability: reason})
	}
}

func (fight *Fight) grantTrait(source effectSource, target *Combatant, step creature.Step) {
	if target.Dead {
		return
	}
	target.grants = append(target.grants, &grantedTrait{
		trait:   step.Grant,
		against: source.actor,
		uses:    step.Uses,
		clock:   fight.newTiming(source.actor, target, step.Duration),
	})
	fight.observer.Observe(TraitGranted{Combatant: target.ID, Trait: step.Grant, Against: source.actor.ID, Ability: source.ability})
}

// deliberate: a granted trait is spent by the attack that carries it, hit or miss, not by the first hit, because Robert ruled the text over the rule elements (Decision 2).
func (fight *Fight) takeGrants(attacker, target *Combatant) []string {
	var traits []string
	attacker.grants = slices.DeleteFunc(attacker.grants, func(grant *grantedTrait) bool {
		if grant.against != target {
			return false
		}
		traits = append(traits, grant.trait)
		fight.observer.Observe(TraitSpent{Combatant: attacker.ID, Trait: grant.trait, Against: target.ID})
		grant.uses--
		return grant.uses < 1
	})
	return traits
}

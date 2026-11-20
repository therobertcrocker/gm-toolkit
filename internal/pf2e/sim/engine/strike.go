package engine

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func attackPenalty(attacksMade int, agile bool) int {
	switch {
	case attacksMade == 0:
		return 0
	case attacksMade == 1 && agile:
		return -4
	case attacksMade == 1:
		return -5
	case agile:
		return -8
	default:
		return -10
	}
}

type strikeOptions struct {
	ability     string
	extraDice   int
	extraDamage []creature.DamagePart
	counts      int
	penalty     string
	disrupts    []string
	lists       creature.Outcomes
	triggeredBy string
}

func (fight *Fight) strikePenalty(attacker *Combatant, strike creature.Strike, options strikeOptions) int {
	// deliberate: the multiple attack penalty applies on the attacker's own turn only, not to a reaction Strike, because the rules count attacks per turn (Decision 24).
	if attacker != fight.current {
		return 0
	}
	if options.penalty == creature.PenaltyShared {
		return attacker.lastPenalty
	}
	penalty := attackPenalty(attacker.AttacksThisTurn, strike.Agile())
	attacker.AttacksThisTurn += max(options.counts, 1)
	attacker.lastPenalty = penalty
	return penalty
}

func strikeDamage(parts []creature.DamagePart, options strikeOptions) []creature.DamagePart {
	if options.extraDice == 0 && len(options.extraDamage) == 0 {
		return parts
	}
	combined := slices.Clone(parts)
	for index := range combined {
		// deliberate: extra dice go on the first rolled part with no category, not on a flat number, because "an extra weapon damage die" needs a die size.
		if combined[index].Category == "" && combined[index].Roll.Sides > 0 {
			combined[index].Roll.Count += options.extraDice
			break
		}
	}
	return append(combined, options.extraDamage...)
}

func (fight *Fight) strike(attacker, target *Combatant, strike creature.Strike, options strikeOptions) (Degree, bool) {
	if !strike.Melee() && !fight.provoke(attacker, strike.ID, TriggerEvent{Ability: options.ability, Strike: strike.ID, Traits: strike.Traits, Ranged: true}) {
		return Failure, false
	}
	penalty := fight.strikePenalty(attacker, strike, options)
	if granted := fight.takeGrants(attacker, target); len(granted) > 0 {
		strike.Traits = append(slices.Clone(strike.Traits), granted...)
	}
	if !fight.passesConcealment(attacker, target) {
		return Failure, false
	}
	// deliberate: the target is told after the concealment check, not before, because a reaction to an attack that never rolls would be spent on nothing.
	attack := TriggerEvent{By: attacker, Ability: options.ability, Strike: strike.ID, Traits: strike.Traits, IsStrike: true, Melee: strike.Melee()}
	attack.Kind = creature.TriggerTargeted
	fight.fire(target, attack)
	if !attacker.Standing() {
		return Failure, false
	}
	attackSelectors, damageSelectors := []string{"attack", "ranged-attack"}, []string{"damage"}
	if strike.Melee() {
		attackSelectors, damageSelectors = []string{"attack", "melee-attack"}, []string{"damage", "melee-damage"}
	}
	die := fight.d20(attacker, creature.RollAttack)
	bonus := strike.Bonus + fight.modifierPool(attacker, attackSelectors, nil, target).Total()
	total := die + bonus + penalty
	armorClass := fight.armorClass(target, strike.Traits)
	target.spendBonuses("ac")
	degree := DegreeOf(die, total, armorClass)
	fight.observer.Observe(StrikeRolled{
		Attacker: attacker.ID,
		Target:   target.ID,
		Strike:   strike.ID,
		Die:      die,
		Bonus:    bonus,
		Penalty:  penalty,
		Total:    total,
		AC:       armorClass,
		Degree:   degree,
	})
	source := effectSource{actor: attacker, ability: options.ability, strike: strike.ID, traits: strike.Traits, melee: strike.Melee(), triggeredBy: options.triggeredBy}
	critical := degree == CriticalSuccess
	if critical {
		fight.disrupt(attacker, options.ability, options.disrupts)
	}
	parts := strikeDamage(strike.Damage, options)
	switch {
	case degree >= Success:
		damageBonus := fight.modifierPool(attacker, damageSelectors, nil, target).Total()
		fight.applyDamage(source, target, fight.rollDamage(parts, critical, damageBonus), critical)
	case degree == Failure:
		// deliberate: splash lands on the primary target only, not on its neighbours, because position comes with A.004.2 (E3 Decision 11).
		if splash := splashParts(parts); len(splash) > 0 {
			fight.applyDamage(source, target, fight.rollDamage(splash, false, 0), false)
		}
	}
	// deliberate: the lists run even when the target died, not only on a living target, because a list can hold a step for the attacker (a recharge); each step kind skips a dead target itself.
	fight.runSteps(source, strike.For(int(degree), true), target)
	fight.runSteps(source, options.lists.For(int(degree), true), target)
	attack.Kind, attack.Degree = creature.TriggerAttacked, degree
	fight.fire(target, attack)
	return degree, true
}

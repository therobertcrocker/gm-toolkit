package engine

import "github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"

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

func (fight *Fight) strike(attacker, target *Combatant, strike creature.Strike, ability string) {
	penalty := attackPenalty(attacker.AttacksThisTurn, strike.Agile())
	attacker.AttacksThisTurn++
	if !fight.passesConcealment(attacker, target) {
		return
	}
	attackSelectors, damageSelectors := []string{"attack", "ranged-attack"}, []string{"damage"}
	if strike.Melee() {
		attackSelectors, damageSelectors = []string{"attack", "melee-attack"}, []string{"damage", "melee-damage"}
	}
	die := fight.roller.Die(20)
	bonus := strike.Bonus + fight.modifierPool(attacker, attackSelectors, nil, target).Total()
	total := die + bonus + penalty
	armorClass := fight.armorClass(target, strike.Traits)
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
	source := effectSource{actor: attacker, ability: ability, strike: strike.ID, traits: strike.Traits}
	critical := degree == CriticalSuccess
	switch {
	case degree >= Success:
		damageBonus := fight.modifierPool(attacker, damageSelectors, nil, target).Total()
		fight.applyDamage(source, target, fight.rollDamage(strike.Damage, critical, damageBonus), critical)
	case degree == Failure:
		// deliberate: splash lands on the primary target only, not on its neighbours, because position comes with Effort 7 (Decision 11).
		if splash := splashParts(strike.Damage); len(splash) > 0 {
			fight.applyDamage(source, target, fight.rollDamage(splash, false, 0), false)
		}
	}
	if !target.Dead {
		fight.runSteps(source, strike.For(int(degree), true), target)
	}
}

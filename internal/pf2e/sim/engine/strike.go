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

func (fight *Fight) strike(attacker, target *Combatant, strike creature.Strike) {
	penalty := attackPenalty(attacker.AttacksThisTurn, strike.Agile())
	attacker.AttacksThisTurn++
	die := fight.roller.Die(20)
	total := die + strike.Bonus + penalty
	armorClass := target.Creature.Defense.AC
	degree := DegreeOf(die, total, armorClass)
	fight.observer.Observe(StrikeRolled{
		Attacker: attacker.ID,
		Target:   target.ID,
		Strike:   strike.ID,
		Die:      die,
		Bonus:    strike.Bonus,
		Penalty:  penalty,
		Total:    total,
		AC:       armorClass,
		Degree:   degree,
	})
	if degree < Success {
		return
	}
	fight.dealDamage(attacker, target, strike, degree == CriticalSuccess)
}

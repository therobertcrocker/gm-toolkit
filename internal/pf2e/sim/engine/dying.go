package engine

import "github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"

const deathThreshold = 4

func (fight *Fight) knockDown(target, attacker *Combatant, critical bool) {
	if target.Creature.Kind == creature.KindNPC {
		fight.kill(target)
		return
	}
	fight.moveBefore(target, attacker)
	fight.setDying(target, dyingOnKnockDown(target, critical))
}

func (fight *Fight) hitWhileDown(target *Combatant, critical bool) {
	if target.Dead {
		return
	}
	if target.Dying == 0 {
		fight.setDying(target, dyingOnKnockDown(target, critical))
		return
	}
	increase := 1
	if critical {
		increase = 2
	}
	fight.setDying(target, target.Dying+increase)
}

func dyingOnKnockDown(target *Combatant, critical bool) int {
	value := 1 + target.Wounded
	if critical {
		value++
	}
	return value
}

func (fight *Fight) setDying(target *Combatant, value int) {
	from := target.Dying
	target.Dying = min(value, deathThreshold)
	fight.observer.Observe(DyingChanged{Combatant: target.ID, From: from, To: target.Dying})
	if target.Dying >= deathThreshold {
		fight.kill(target)
	}
}

func (fight *Fight) kill(target *Combatant) {
	target.Dead = true
	fight.observer.Observe(Died{Combatant: target.ID})
}

func (fight *Fight) recoveryCheck(target *Combatant) {
	dc := 10 + target.Dying
	die := fight.roller.Die(20)
	degree := DegreeOf(die, die, dc)
	fight.observer.Observe(RecoveryChecked{Combatant: target.ID, Die: die, DC: dc, Degree: degree})

	change := 0
	switch degree {
	case CriticalSuccess:
		change = -2
	case Success:
		change = -1
	case Failure:
		change = 1
	case CriticalFailure:
		change = 2
	}
	next := max(target.Dying+change, 0)
	if next > 0 {
		fight.setDying(target, next)
		return
	}
	fight.setDying(target, 0)
	target.Wounded++
	fight.observer.Observe(Stabilized{Combatant: target.ID, Wounded: target.Wounded})
}

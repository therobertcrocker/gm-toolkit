package engine

import "github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"

const (
	defaultDyingMax   = 4
	defaultRecoveryDC = 10
)

func (combatant *Combatant) deathValue() int {
	dyingMax := combatant.Creature.DyingMax
	if dyingMax == 0 {
		dyingMax = defaultDyingMax
	}
	return dyingMax - combatant.Condition("doomed")
}

func (combatant *Combatant) MaxHP() int {
	return combatant.Creature.Defense.HP - combatant.Condition("drained")*max(combatant.Creature.Level, 1)
}

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
	limit := target.deathValue()
	// deliberate: hero points are spent when the PC would die, not at the first rise in dying, because the script plays it that way and Robert kept it (Decision 20b).
	if value >= limit && target.HeroPoints > 0 {
		fight.heroicRecovery(target)
		return
	}
	from := target.Dying
	target.Dying = min(value, max(limit, 1))
	fight.observer.Observe(DyingChanged{Combatant: target.ID, From: from, To: target.Dying})
	if value >= limit {
		fight.kill(target)
	}
}

func (fight *Fight) heroicRecovery(target *Combatant) {
	from := target.Dying
	target.HeroPoints = 0
	target.Dying = 0
	fight.observer.Observe(HeroicRecovery{Combatant: target.ID})
	fight.observer.Observe(DyingChanged{Combatant: target.ID, From: from, To: 0})
}

func (fight *Fight) drain(source, target *Combatant, before int) {
	gained := target.Condition("drained") - before
	if gained <= 0 || target.Dead {
		return
	}
	loss := gained * max(target.Creature.Level, 1)
	wasStanding := target.Standing()
	target.HP = max(target.HP-loss, 0)
	fight.observer.Observe(HPLost{Combatant: target.ID, Reason: "drained", Amount: loss, HPAfter: target.HP})
	if wasStanding && target.HP == 0 {
		fight.knockDown(target, source, false)
	}
}

func (fight *Fight) kill(target *Combatant) {
	target.Dead = true
	fight.observer.Observe(Died{Combatant: target.ID})
}

func (fight *Fight) recoveryCheck(target *Combatant) {
	base := target.Creature.RecoveryDC
	if base == 0 {
		base = defaultRecoveryDC
	}
	dc := base + target.Dying
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

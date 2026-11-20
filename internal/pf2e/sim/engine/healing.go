package engine

func (fight *Fight) heal(source effectSource, target *Combatant, amount int) {
	if target.Dead {
		return
	}
	hpBefore := target.HP
	wasDown := hpBefore == 0
	target.HP = min(target.HP+max(amount, 0), target.MaxHP())
	fight.observer.Observe(Healed{Source: source.actor.ID, Target: target.ID, Ability: source.ability, Amount: target.HP - hpBefore, HPAfter: target.HP, TriggeredBy: source.triggeredBy})
	if !wasDown || target.HP == 0 {
		return
	}
	if target.Dying > 0 {
		from := target.Dying
		target.Dying = 0
		target.Wounded++
		fight.observer.Observe(Revived{Combatant: target.ID, DyingFrom: from, Wounded: target.Wounded})
	}
	if source.actor != target && source.actor.Side == target.Side {
		fight.loyalty(source.actor, "revive")
	}
}

func (fight *Fight) gainTempHP(target *Combatant, amount int) {
	if target.Dead || amount <= target.TempHP {
		return
	}
	target.TempHP = amount
	fight.observer.Observe(TempHPGained{Combatant: target.ID, Amount: amount})
}

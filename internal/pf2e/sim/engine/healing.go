package engine

func (fight *Fight) heal(source effectSource, target *Combatant, amount int) {
	if target.Dead {
		return
	}
	wasDown := target.HP == 0
	target.HP = min(target.HP+max(amount, 0), target.MaxHP())
	fight.observer.Observe(Healed{Source: source.actor.ID, Target: target.ID, Ability: source.ability, Amount: amount, HPAfter: target.HP})
	if wasDown && target.HP > 0 && target.Dying > 0 {
		from := target.Dying
		target.Dying = 0
		target.Wounded++
		fight.observer.Observe(Revived{Combatant: target.ID, DyingFrom: from, Wounded: target.Wounded})
	}
}

func (fight *Fight) gainTempHP(target *Combatant, amount int) {
	if target.Dead || amount <= target.TempHP {
		return
	}
	target.TempHP = amount
	fight.observer.Observe(TempHPGained{Combatant: target.ID, Amount: amount})
}

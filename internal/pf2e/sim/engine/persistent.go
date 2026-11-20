package engine

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

const (
	persistentRecoveryDC = 15
	persistentAbility    = "persistent"
)

type persistentScale int

const (
	persistentWhole persistentScale = iota
	persistentDoubled
	persistentHalved
)

func (scale persistentScale) apply(amount int) int {
	switch scale {
	case persistentDoubled:
		return amount * 2
	case persistentHalved:
		return amount / 2
	}
	return amount
}

type persistentInstance struct {
	source effectSource
	part   creature.DamagePart
	scale  persistentScale
}

func (instance *persistentInstance) fourTimesAverage() int {
	roll := instance.part.Roll
	return instance.scale.apply(2 * (roll.Count*(roll.Sides+1) + 2*roll.Modifier))
}

func (instance *persistentInstance) dice() string {
	switch instance.scale {
	case persistentDoubled:
		return instance.part.Roll.String() + " x2"
	case persistentHalved:
		return instance.part.Roll.String() + " /2"
	}
	return instance.part.Roll.String()
}

func (fight *Fight) addPersistent(source effectSource, target *Combatant, part creature.DamagePart, scale persistentScale) {
	if slices.Contains(target.Creature.Defense.Immune, part.Type) {
		return
	}
	added := &persistentInstance{source: source, part: part, scale: scale}
	for index, existing := range target.persistent {
		if existing.part.Type != part.Type {
			continue
		}
		if added.fourTimesAverage() <= existing.fourTimesAverage() {
			return
		}
		target.persistent[index] = added
		fight.observer.Observe(PersistentChanged{Combatant: target.ID, Type: part.Type, Dice: added.dice()})
		return
	}
	target.persistent = append(target.persistent, added)
	fight.observer.Observe(PersistentChanged{Combatant: target.ID, Type: part.Type, Dice: added.dice()})
}

func (fight *Fight) takePersistentDamage(actor *Combatant) {
	var kept []*persistentInstance
	for _, instance := range actor.persistent {
		if actor.Dead {
			break
		}
		amount := instance.scale.apply(max(instance.part.Roll.Roll(fight.roller), 0))
		source := instance.source
		source.ability, source.strike, source.lethal, source.melee = persistentAbility, "", false, false
		fight.applyDamage(source, actor, damageRoll{parts: []rolledPart{{damageType: instance.part.Type, amount: amount}}}, false)
		if actor.Dead {
			break
		}
		die := fight.roller.Die(20)
		fight.observer.Observe(FlatChecked{Combatant: actor.ID, Reason: "persistent " + instance.part.Type, Die: die, DC: persistentRecoveryDC})
		if die >= persistentRecoveryDC {
			fight.observer.Observe(PersistentEnded{Combatant: actor.ID, Type: instance.part.Type})
			continue
		}
		kept = append(kept, instance)
	}
	if actor.Dead {
		kept = nil
	}
	actor.persistent = kept
}

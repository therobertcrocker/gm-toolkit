package engine

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

const persistentRecoveryDC = 15

type persistentInstance struct {
	source  effectSource
	part    creature.DamagePart
	doubled bool
}

func (instance *persistentInstance) twiceAverage() int {
	roll := instance.part.Roll
	value := roll.Count*(roll.Sides+1) + 2*roll.Modifier
	if instance.doubled {
		value *= 2
	}
	return value
}

func (instance *persistentInstance) dice() string {
	if instance.doubled {
		return instance.part.Roll.String() + " x2"
	}
	return instance.part.Roll.String()
}

func (fight *Fight) addPersistent(source effectSource, target *Combatant, part creature.DamagePart, doubled bool) {
	if slices.Contains(target.Creature.Defense.Immune, part.Type) {
		return
	}
	added := &persistentInstance{source: source, part: part, doubled: doubled}
	for index, existing := range target.persistent {
		if existing.part.Type != part.Type {
			continue
		}
		if added.twiceAverage() <= existing.twiceAverage() {
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
		amount := max(instance.part.Roll.Roll(fight.roller), 0)
		if instance.doubled {
			amount *= 2
		}
		source := instance.source
		source.ability, source.strike, source.lethal = "persistent", "", false
		fight.applyDamage(source, actor, damageRoll{parts: []rolledPart{{damageType: instance.part.Type, amount: amount}}}, false)
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

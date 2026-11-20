package engine

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

type rolledPart struct {
	damageType string
	amount     int
}

type damageRoll struct {
	parts      []rolledPart
	persistent []creature.DamagePart
	scale      persistentScale
}

func (fight *Fight) rollDamage(parts []creature.DamagePart, critical bool, bonus int) damageRoll {
	roll := damageRoll{}
	if critical {
		roll.scale = persistentDoubled
	}
	bonusAdded := false
	for _, part := range parts {
		if part.Category == "persistent" {
			roll.persistent = append(roll.persistent, part)
			continue
		}
		amount := part.Roll.Roll(fight.roller)
		if part.Category == "" && !bonusAdded {
			amount += bonus
			bonusAdded = true
		}
		amount = max(amount, 0)
		if critical && part.Category == "" {
			amount *= 2
		}
		roll.parts = addRolled(roll.parts, part.Type, amount)
	}
	// deliberate: a roll that comes to less than 1 deals 1, not 0, because penalties never take damage below 1; a save's halving comes after.
	if len(roll.parts) > 0 && roll.total() < 1 {
		roll.parts[0].amount = 1
	}
	return roll
}

func (roll damageRoll) total() int {
	total := 0
	for _, part := range roll.parts {
		total += part.amount
	}
	return total
}

func splashParts(parts []creature.DamagePart) []creature.DamagePart {
	var splash []creature.DamagePart
	for _, part := range parts {
		if part.Category == "splash" {
			splash = append(splash, part)
		}
	}
	return splash
}

func addRolled(parts []rolledPart, damageType string, amount int) []rolledPart {
	for index := range parts {
		if parts[index].damageType == damageType {
			parts[index].amount += amount
			return parts
		}
	}
	return append(parts, rolledPart{damageType: damageType, amount: amount})
}

func (roll damageRoll) basic(degree Degree) damageRoll {
	scaled := damageRoll{parts: make([]rolledPart, len(roll.parts)), persistent: roll.persistent, scale: roll.scale}
	switch degree {
	case Success:
		scaled.scale = persistentHalved
	case CriticalFailure:
		scaled.scale = persistentDoubled
	}
	total, scaledTotal := 0, 0
	for index, part := range roll.parts {
		amount := part.amount
		switch degree {
		case Success:
			amount /= 2
		case CriticalFailure:
			amount *= 2
		}
		scaled.parts[index] = rolledPart{damageType: part.damageType, amount: amount}
		total += part.amount
		scaledTotal += amount
	}
	if degree == Success && len(scaled.parts) > 0 {
		scaled.parts[0].amount += total/2 - scaledTotal
	}
	return scaled
}

func (fight *Fight) applyDamage(source effectSource, target *Combatant, roll damageRoll, critical bool) {
	parts := make([]DamageAmount, 0, len(roll.parts))
	total := 0
	for _, part := range roll.parts {
		taken := target.adjustedDamage(part.damageType, source.traits, part.amount)
		parts = append(parts, DamageAmount{Type: part.damageType, Rolled: part.amount, Taken: taken})
		total += taken
	}
	traitChange := 0
	if weakness, found := traitAdjustment(target.Creature.Defense.Weakness, source.traits); found && total > 0 {
		traitChange += weakness
	}
	if resistance, found := traitAdjustment(target.Creature.Defense.Resist, source.traits); found {
		traitChange -= min(resistance, total+traitChange)
	}
	total += traitChange

	absorbed := min(target.TempHP, total)
	target.TempHP -= absorbed
	wasStanding := target.Standing()
	target.HP = max(target.HP-(total-absorbed), 0)
	fight.observer.Observe(DamageDealt{
		Source:      source.actor.ID,
		Target:      target.ID,
		Strike:      source.strike,
		Ability:     source.ability,
		Critical:    critical,
		Parts:       parts,
		Total:       total,
		Absorbed:    absorbed,
		HPAfter:     target.HP,
		Trait:       traitChange,
		TriggeredBy: source.triggeredBy,
	})
	fight.adapt(target, parts)
	if total > absorbed {
		switch {
		case wasStanding && target.HP == 0 && source.lethal:
			fight.kill(target)
		case wasStanding && target.HP == 0:
			fight.knockDown(target, source.actor, critical)
		case !wasStanding:
			fight.hitWhileDown(target, critical)
		}
	}
	if !target.Dead {
		for _, part := range roll.persistent {
			fight.addPersistent(source, target, part, roll.scale)
		}
	}
	if total > 0 && source.actor.Side != target.Side && source.ability != persistentAbility {
		fight.fire(target, TriggerEvent{
			Kind:    creature.TriggerDamaged,
			By:      source.actor,
			Ability: source.ability,
			Strike:  source.strike,
			Traits:  source.traits,
			Melee:   source.melee,
		})
	}
}

func (fight *Fight) adapt(target *Combatant, parts []DamageAmount) {
	adaptive := target.Creature.Defense.Adaptive
	if adaptive == nil || target.Dead {
		return
	}
	for _, part := range parts {
		if part.Taken == 0 || !slices.Contains(adaptive.Types, part.Type) {
			continue
		}
		if part.Type != target.adaptiveType {
			target.adaptiveType = part.Type
			fight.observer.Observe(ResistanceAdapted{Combatant: target.ID, Type: part.Type, Value: adaptive.Value})
		}
		return
	}
}

func (target *Combatant) adjustedDamage(damageType string, traits []string, amount int) int {
	defense := target.Creature.Defense
	immune := slices.ContainsFunc(defense.Immune, func(key string) bool { return adjustmentCovers(key, damageType) })
	if immune {
		return 0
	}
	weakness, hasWeakness := strongestAdjustment(defense.Weakness, damageType, traits)
	for _, applied := range target.defenseChanges {
		if applied.change.Weakness > weakness && slices.Contains(applied.change.Types, damageType) {
			weakness, hasWeakness = applied.change.Weakness, true
		}
	}
	if hasWeakness && amount > 0 {
		amount += weakness
	}
	resistance, hasResistance := strongestAdjustment(defense.Resist, damageType, traits)
	if adaptive := defense.Adaptive; adaptive != nil && target.adaptiveType == damageType && adaptive.Value > resistance {
		resistance, hasResistance = adaptive.Value, true
	}
	if hasResistance {
		lowered := 0
		for _, applied := range target.defenseChanges {
			if slices.Contains(applied.change.Types, damageType) {
				lowered = max(lowered, applied.change.Resistance)
			}
		}
		amount -= max(resistance-lowered, 0)
	}
	return max(amount, 0)
}

func traitAdjustment(adjustments map[string]creature.Adjustment, traits []string) (int, bool) {
	strongest, found := 0, false
	for key, adjustment := range adjustments {
		if !creature.IsDefenseTrait(key) || !slices.Contains(traits, key) {
			continue
		}
		if !found || adjustment.Value > strongest {
			strongest, found = adjustment.Value, true
		}
	}
	return strongest, found
}

func strongestAdjustment(adjustments map[string]creature.Adjustment, damageType string, traits []string) (int, bool) {
	strongest, found := 0, false
	for key, adjustment := range adjustments {
		if creature.IsDefenseTrait(key) || !adjustmentCovers(key, damageType) {
			continue
		}
		excepted := slices.Contains(adjustment.Except, damageType) ||
			slices.ContainsFunc(traits, func(trait string) bool { return slices.Contains(adjustment.Except, trait) })
		if excepted {
			continue
		}
		if !found || adjustment.Value > strongest {
			strongest, found = adjustment.Value, true
		}
	}
	return strongest, found
}

func adjustmentCovers(key, damageType string) bool {
	switch key {
	case "all":
		return true
	case "physical":
		return creature.IsPhysical(damageType)
	default:
		return key == damageType
	}
}

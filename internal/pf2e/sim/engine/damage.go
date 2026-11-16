package engine

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func (fight *Fight) dealDamage(attacker, target *Combatant, strike creature.Strike, critical bool) {
	rolledByType := map[string]int{}
	var types []string
	for _, part := range strike.Damage {
		amount := max(part.Roll.Roll(fight.roller), 0)
		if critical {
			amount *= 2
		}
		if _, seen := rolledByType[part.Type]; !seen {
			types = append(types, part.Type)
		}
		rolledByType[part.Type] += amount
	}

	parts := make([]DamageAmount, 0, len(types))
	total := 0
	for _, damageType := range types {
		rolled := rolledByType[damageType]
		taken := adjustedDamage(target.Creature.Defense, damageType, strike.Traits, rolled)
		parts = append(parts, DamageAmount{Type: damageType, Rolled: rolled, Taken: taken})
		total += taken
	}

	wasStanding := target.Standing()
	target.HP = max(target.HP-total, 0)
	fight.observer.Observe(DamageDealt{
		Source:   attacker.ID,
		Target:   target.ID,
		Strike:   strike.ID,
		Critical: critical,
		Parts:    parts,
		Total:    total,
		HPAfter:  target.HP,
	})
	if total == 0 {
		return
	}
	switch {
	case wasStanding && target.HP == 0:
		fight.knockDown(target, attacker, critical)
	case !wasStanding:
		fight.hitWhileDown(target, critical)
	}
}

func adjustedDamage(defense creature.Defense, damageType string, traits []string, amount int) int {
	immune := slices.ContainsFunc(defense.Immune, func(key string) bool { return adjustmentCovers(key, damageType) })
	if immune {
		return 0
	}
	if weakness, found := strongestAdjustment(defense.Weakness, damageType, traits); found {
		amount += weakness
	}
	if resistance, found := strongestAdjustment(defense.Resist, damageType, traits); found {
		amount -= resistance
	}
	return max(amount, 0)
}

func strongestAdjustment(adjustments map[string]creature.Adjustment, damageType string, traits []string) (int, bool) {
	strongest, found := 0, false
	for key, adjustment := range adjustments {
		if !adjustmentCovers(key, damageType) {
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

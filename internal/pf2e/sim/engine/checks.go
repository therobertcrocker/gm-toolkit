package engine

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

var saveSelectors = map[string][]string{
	"fortitude": {"fortitude", "saving-throw"},
	"reflex":    {"reflex", "saving-throw"},
	"will":      {"will", "saving-throw"},
}

func (fight *Fight) modifierTotal(combatant *Combatant, selectors, traits []string) int {
	return fight.modifierPool(combatant, selectors, traits, nil).Total()
}

func (fight *Fight) modifierPool(combatant *Combatant, selectors, traits []string, target *Combatant) *creature.BonusPool {
	pool := creature.NewBonusPool()
	for _, modifier := range combatant.Creature.Modifiers {
		if !modifier.Applies(selectors, traits) {
			continue
		}
		value := modifier.Value
		if modifier.PerCondition != "" {
			if target == nil {
				continue
			}
			value *= target.Condition(modifier.PerCondition)
		}
		if value != 0 {
			pool.Add(modifier.Type, value)
		}
	}
	fight.conditionPenalties(combatant, selectors, pool)
	return pool
}

func (fight *Fight) armorClass(target *Combatant, traits []string) int {
	return target.Creature.Defense.AC + fight.modifierTotal(target, []string{"ac"}, traits)
}

func saveBonus(defense creature.Defense, save string) int {
	switch save {
	case "fortitude":
		return defense.Fortitude
	case "reflex":
		return defense.Reflex
	default:
		return defense.Will
	}
}

func (fight *Fight) savingThrow(target *Combatant, save string, dc, shift int, source effectSource) Degree {
	selectors := saveSelectors[save]
	bonus := saveBonus(target.Creature.Defense, save) + fight.modifierTotal(target, selectors, source.traits)
	die := fight.roller.Die(20)
	total := die + bonus
	degree := shiftDegree(DegreeOf(die, total, dc), shift)
	degree = applyDegreeShifts(target, selectors, source.traits, degree)
	fight.observer.Observe(SaveRolled{
		Combatant: target.ID,
		Source:    source.actor.ID,
		Ability:   source.ability,
		Save:      save,
		Die:       die,
		Bonus:     bonus,
		Total:     total,
		DC:        dc,
		Degree:    degree,
	})
	return degree
}

func shiftDegree(degree Degree, steps int) Degree {
	return min(max(degree+Degree(steps), CriticalFailure), CriticalSuccess)
}

// deliberate: a step's own shift applies before the saver's degree shifts, not after, because Robert chose that order at the E3 plan sign-off (Decision 20a).
func applyDegreeShifts(combatant *Combatant, selectors, traits []string, degree Degree) Degree {
	for _, id := range combatant.Creature.DegreeShiftIDs() {
		shift := combatant.Creature.DegreeShifts[id]
		if shift.Applies(selectors, traits) && Degree(slices.Index(creature.DegreeWords, shift.From)) == degree {
			return Degree(slices.Index(creature.DegreeWords, shift.To))
		}
	}
	return degree
}

package creature

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

func (creature *Creature) prepare() error {
	var problems []string
	report := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf(format, args...))
	}

	if creature.Name == "" {
		report("name is empty")
	}
	if creature.Kind != KindPC && creature.Kind != KindNPC {
		report("kind %q: want pc or npc", creature.Kind)
	}
	if creature.Defense.HP < 1 {
		report("defense.hp must be at least 1")
	}
	if creature.Defense.AC < 1 {
		report("defense.ac must be at least 1")
	}
	if creature.HeroPoints < 0 || creature.RecoveryDC < 0 {
		report("hero_points and recovery_dc must not be negative")
	}
	if creature.DyingMax != 0 && creature.DyingMax < 2 {
		report("dying_max %d: want at least 2, or leave it unset for 4", creature.DyingMax)
	}
	for _, table := range []struct {
		name        string
		adjustments map[string]Adjustment
	}{
		{"resist", creature.Defense.Resist},
		{"weakness", creature.Defense.Weakness},
	} {
		for _, key := range slices.Sorted(maps.Keys(table.adjustments)) {
			if key != "all" && key != "physical" && !IsDamageType(key) && !IsDefenseTrait(key) {
				report("defense.%s.%s: not a damage type, physical, all or one of %s", table.name, key, strings.Join(DefenseTraits, ", "))
			}
			if table.adjustments[key].Value < 1 {
				report("defense.%s.%s: value must be at least 1", table.name, key)
			}
		}
	}

	if adaptive := creature.Defense.Adaptive; adaptive != nil {
		if adaptive.Value < 1 || len(adaptive.Types) == 0 || !slices.ContainsFunc(adaptive.Types, IsDamageType) {
			report("defense.adaptive: want a value of at least 1 and damage types")
		}
		for _, damageType := range adaptive.Types {
			if !IsDamageType(damageType) {
				report("defense.adaptive: %q is not a damage type", damageType)
			}
		}
	}
	creature.checkFightRules(report)

	for _, id := range creature.StrikeIDs() {
		strike := creature.Strikes[id]
		strike.ID = id
		if strike.Name == "" {
			report("strike.%s: name is empty", id)
		}
		if strike.Range < 0 {
			report("strike.%s: range must not be negative", id)
		}
		creature.checkOutcomes("strike."+id, strike.Outcomes, true, report)
		if len(strike.Damage) == 0 {
			report("strike.%s: no damage", id)
		}
		prepareDamage("strike."+id, strike.Damage, report)
		if len(strike.Damage) > 0 && !slices.ContainsFunc(strike.Damage, func(part DamagePart) bool { return part.Category == "" }) {
			report("strike.%s: needs a damage part with no category", id)
		}
		creature.Strikes[id] = strike
	}

	for _, id := range creature.AbilityIDs() {
		ability := creature.Abilities[id]
		ability.ID = id
		creature.Abilities[id] = ability
		if ability.Ignore {
			continue
		}
		if ability.Name == "" {
			report("ability.%s: name is empty", id)
		}
		if ability.Actions < 1 || ability.Actions > 3 {
			report("ability.%s: actions %d: want 1 to 3", id, ability.Actions)
		}
		if len(ability.Steps) == 0 {
			report("ability.%s: no steps", id)
		}
		checkArea("ability."+id, ability.Area, report)
		for index := range ability.Steps {
			creature.checkStep(fmt.Sprintf("ability.%s.steps[%d]", id, index), &ability.Steps[index], report)
		}
	}

	if len(problems) > 0 {
		return fmt.Errorf("%w: %s", ErrInvalid, strings.Join(problems, "; "))
	}
	return nil
}

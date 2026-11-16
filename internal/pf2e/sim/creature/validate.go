package creature

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
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
	for _, table := range []struct {
		name        string
		adjustments map[string]Adjustment
	}{
		{"resist", creature.Defense.Resist},
		{"weakness", creature.Defense.Weakness},
	} {
		for _, key := range slices.Sorted(maps.Keys(table.adjustments)) {
			if key != "all" && key != "physical" && !IsDamageType(key) {
				report("defense.%s.%s: not a damage type, physical or all", table.name, key)
			}
		}
	}

	for _, id := range creature.StrikeIDs() {
		strike := creature.Strikes[id]
		strike.ID = id
		if strike.Name == "" {
			report("strike.%s: name is empty", id)
		}
		if len(strike.Damage) == 0 {
			report("strike.%s: no damage", id)
		}
		for index := range strike.Damage {
			part := &strike.Damage[index]
			expression, err := dice.Parse(part.Dice)
			if err != nil {
				report("strike.%s.damage[%d]: %v", id, index, err)
			}
			part.Roll = expression
			if !IsDamageType(part.Type) {
				report("strike.%s.damage[%d]: unknown damage type %q", id, index, part.Type)
			}
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
		for index, step := range ability.Steps {
			if _, found := creature.Strikes[step.Strike]; !found {
				report("ability.%s.steps[%d]: unknown strike %q", id, index, step.Strike)
			}
		}
	}

	if len(problems) > 0 {
		return fmt.Errorf("%w: %s", ErrInvalid, strings.Join(problems, "; "))
	}
	return nil
}

package engine

import (
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

type scriptedRoller struct {
	t      *testing.T
	values []int
}

func script(t *testing.T, values ...int) *scriptedRoller {
	return &scriptedRoller{t: t, values: values}
}

func (roller *scriptedRoller) Die(sides int) int {
	roller.t.Helper()
	if len(roller.values) == 0 {
		roller.t.Fatalf("scripted roller ran out on a d%d", sides)
	}
	value := roller.values[0]
	roller.values = roller.values[1:]
	if value < 1 || value > sides {
		roller.t.Fatalf("scripted value %d is not a d%d result", value, sides)
	}
	return value
}

type recorder struct {
	events []Event
}

func (recorder *recorder) Observe(event Event) {
	recorder.events = append(recorder.events, event)
}

func fist() creature.Strike {
	return creature.Strike{
		ID:    "fist",
		Name:  "Fist",
		Bonus: 10,
		Damage: []creature.DamagePart{
			{Dice: "1d6+2", Type: "bludgeoning", Roll: dice.Expression{Count: 1, Sides: 6, Modifier: 2}},
		},
	}
}

func testCombatant(id string, side Side, kind creature.Kind, hp, armorClass int) *Combatant {
	source := &creature.Creature{
		ID:         id,
		Name:       id,
		Kind:       kind,
		Level:      1,
		Perception: 5,
		Defense:    creature.Defense{HP: hp, AC: armorClass},
		Strikes:    map[string]creature.Strike{"fist": fist()},
	}
	return NewCombatant(side, source)
}

func testFight(roller dice.Roller, combatants ...*Combatant) (*Fight, *recorder) {
	events := &recorder{}
	fight := &Fight{
		roller:     roller,
		observer:   events,
		decider:    FirstFitting{},
		combatants: combatants,
		order:      slices.Clone(combatants),
	}
	return fight, events
}

func damagePart(t *testing.T, text, damageType string) creature.DamagePart {
	t.Helper()
	expression, err := dice.Parse(text)
	if err != nil {
		t.Fatalf("damagePart(%q): %v", text, err)
	}
	return creature.DamagePart{Dice: text, Type: damageType, Roll: expression}
}

func giveAbility(combatant *Combatant, ability creature.Ability) {
	if combatant.Creature.Abilities == nil {
		combatant.Creature.Abilities = map[string]creature.Ability{}
	}
	combatant.Creature.Abilities[ability.ID] = ability
}

func orderIDs(fight *Fight) []string {
	ids := make([]string, len(fight.order))
	for index, combatant := range fight.order {
		ids[index] = combatant.ID
	}
	return ids
}

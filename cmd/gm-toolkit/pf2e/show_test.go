package pf2e

import (
	"bytes"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

func TestPrintCreature(t *testing.T) {
	shown := &creature.Creature{
		Name:       "Abyssal Luna",
		Kind:       creature.KindNPC,
		Level:      16,
		Perception: 29,
		Speed:      creature.Speed{Land: 60, Fly: 180},
		Defense: creature.Defense{
			HP: 300, AC: 36, Fortitude: 27, Reflex: 28, Will: 29,
			Resist: map[string]creature.Adjustment{"all": {Value: 10, Except: []string{"holy"}}},
		},
		Strikes: map[string]creature.Strike{
			"slash": {Name: "Slash", Bonus: 30, Traits: []string{"agile"}, Damage: []creature.DamagePart{
				{Type: "slashing", Roll: dice.Expression{Count: 3, Sides: 6, Modifier: 14}},
				{Type: "mental", Roll: dice.Expression{Count: 1, Sides: 8}},
			}},
		},
		Abilities: map[string]creature.Ability{
			"abyssal-frenzy":  {Name: "Abyssal Frenzy", Actions: 2, Steps: []creature.Step{{Strike: "slash"}, {Strike: "slash"}}},
			"prophetic-wings": {Ignore: true},
		},
		Layers: []string{"import.toml", "hand.toml"},
	}
	var out bytes.Buffer
	printCreature(&out, shown, false)
	if strings.Contains(out.String(), "\033") {
		t.Errorf("unstyled output holds an escape code:\n%q", out.String())
	}
	for _, want := range []string{
		"Abyssal Luna  npc 16",
		"HP 300  AC 36  Fort +27  Ref +28  Will +29",
		"Resist all 10 (except holy)",
		"Speed 60 ft, fly 180 ft",
		"Slash +30 (agile)  3d6+14 slashing plus 1d8 mental",
		"Abyssal Frenzy [2]  strike slash, strike slash",
		"prophetic-wings (ignored)",
		"Layers import.toml + hand.toml",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out.String())
		}
	}
}

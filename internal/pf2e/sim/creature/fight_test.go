package creature_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

const fightRules = `
[modifier.occult-saves]
selector = "saving-throw"
type = "status"
value = 2
against = ["occult"]

[degree_shift.juggernaut]
selector = "fortitude"
from = "success"
to = "critical-success"
`

func TestLoadReadsFightRules(t *testing.T) {
	dir := writeLayers(t, map[string]string{"hand.toml": minimalHand + fightRules})
	loaded, err := creature.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	modifier, found := loaded.Modifiers["occult-saves"]
	if !found {
		t.Fatalf("modifiers = %v, want occult-saves", loaded.Modifiers)
	}
	wantModifier := creature.CheckModifier{Selector: "saving-throw", Type: "status", Value: 2, Against: []string{"occult"}}
	if modifier.Selector != wantModifier.Selector || modifier.Type != wantModifier.Type || modifier.Value != wantModifier.Value || !slices.Equal(modifier.Against, wantModifier.Against) {
		t.Errorf("modifier = %+v, want %+v", modifier, wantModifier)
	}
	if !modifier.Applies([]string{"will", "saving-throw"}, []string{"occult"}) {
		t.Error("modifier does not apply to a Will save against an occult effect")
	}
	if modifier.Applies([]string{"will", "saving-throw"}, []string{"arcane"}) {
		t.Error("modifier applies to a Will save against an arcane effect")
	}

	shift, found := loaded.DegreeShifts["juggernaut"]
	if !found {
		t.Fatalf("degree shifts = %v, want juggernaut", loaded.DegreeShifts)
	}
	if shift.Selector != "fortitude" || shift.From != "success" || shift.To != "critical-success" || len(shift.Against) != 0 {
		t.Errorf("degree shift = %+v, want fortitude success -> critical-success", shift)
	}
}

func TestLoadRejectsBadFightRules(t *testing.T) {
	cases := []struct {
		name     string
		rules    string
		wantText string
	}{
		{name: "unknown selector", rules: "[modifier.x]\nselector = \"speed\"\nvalue = 1\n", wantText: `selector "speed"`},
		{name: "unknown type", rules: "[modifier.x]\nselector = \"ac\"\ntype = \"luck\"\nvalue = 1\n", wantText: `type "luck"`},
		{name: "value zero", rules: "[modifier.x]\nselector = \"ac\"\nvalue = 0\n", wantText: "value must not be 0"},
		{name: "against on attack", rules: "[modifier.x]\nselector = \"attack\"\nvalue = 1\nagainst = [\"occult\"]\n", wantText: "against needs selector ac or a save"},
		{name: "degree shift on ac", rules: "[degree_shift.x]\nselector = \"ac\"\nfrom = \"success\"\nto = \"critical-success\"\n", wantText: `degree_shift.x: selector "ac"`},
		{name: "bad degree word", rules: "[degree_shift.x]\nselector = \"will\"\nfrom = \"great\"\nto = \"success\"\n", wantText: `from "great"`},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			dir := writeLayers(t, map[string]string{"hand.toml": minimalHand + "\n" + testCase.rules})
			_, err := creature.Load(dir)
			if !errors.Is(err, creature.ErrInvalid) {
				t.Fatalf("Load error = %v, want ErrInvalid", err)
			}
			if !strings.Contains(err.Error(), testCase.wantText) {
				t.Errorf("Load error = %v, want text %q", err, testCase.wantText)
			}
		})
	}
}

const lethalAbility = `
[ability.disintegrate]
name = "Disintegrate"
actions = 2
steps = [{ damage = [{ dice = "12d10", type = "untyped" }], lethal = true }]
`

func TestLoadReadsDownFields(t *testing.T) {
	hand := strings.Replace(minimalHand, "perception = 3", "perception = 3\nhero_points = 1\ndying_max = 6\nrecovery_dc = 6", 1) + lethalAbility
	loaded, err := creature.Load(writeLayers(t, map[string]string{"hand.toml": hand}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.HeroPoints != 1 || loaded.DyingMax != 6 || loaded.RecoveryDC != 6 {
		t.Errorf("hero points, dying max, recovery DC = %d, %d, %d, want 1, 6, 6", loaded.HeroPoints, loaded.DyingMax, loaded.RecoveryDC)
	}
	step := loaded.Abilities["disintegrate"].Steps[0]
	if !step.Lethal || step.Kind() != creature.StepDamage || step.Damage[0].Type != "untyped" {
		t.Errorf("disintegrate step = %+v, want a lethal untyped damage step", step)
	}

	cases := []struct {
		name     string
		hand     string
		wantText string
	}{
		{"dying max 1", strings.Replace(minimalHand, "perception = 3", "perception = 3\ndying_max = 1", 1), "dying_max 1"},
		{"negative hero points", strings.Replace(minimalHand, "perception = 3", "perception = 3\nhero_points = -1", 1), "hero_points, recovery_dc and fast_healing must not be negative"},
		{"negative fast healing", strings.Replace(minimalHand, "perception = 3", "perception = 3\nfast_healing = -1", 1), "hero_points, recovery_dc and fast_healing must not be negative"},
		{"lethal on a condition", minimalHand + "\n[ability.bad]\nname = \"Bad\"\nactions = 1\nsteps = [{ condition = \"off-guard\", lethal = true }]\n", "lethal needs damage"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := creature.Load(writeLayers(t, map[string]string{"hand.toml": testCase.hand}))
			if !errors.Is(err, creature.ErrInvalid) {
				t.Fatalf("Load error = %v, want ErrInvalid", err)
			}
			if !strings.Contains(err.Error(), testCase.wantText) {
				t.Errorf("Load error = %v, want text %q", err, testCase.wantText)
			}
		})
	}
}

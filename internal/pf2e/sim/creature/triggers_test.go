package creature_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

const triggeredAbilities = `
[extra_reaction]
on-attacked = 2

[ability.on-targeted]
name    = "On Targeted"
trigger = { on = "targeted", cost = "reaction", attack = "melee-strike" }
steps   = [{ strike = "fist" }]

[ability.on-attacked]
name    = "On Attacked"
trigger = { on = "attacked", cost = "reaction", attack = "strike", degrees = ["critical-success"] }
steps   = [{ strike = "fist" }]

[ability.on-saved]
name    = "On Saved"
trigger = { on = "saved", cost = "free", degrees = ["critical-failure", "failure"], damaging = true }
steps   = [{ strike = "fist" }]

[ability.on-damaged]
name    = "On Damaged"
trigger = { on = "damaged", melee = true }
steps   = [{ strike = "fist" }]

[ability.on-turn-start]
name    = "On Turn Start"
trigger = { on = "turn-start", cost = "free", hp_at_most = 75 }
limit   = { per_fight = 1 }
steps   = [{ strike = "fist" }]

[ability.on-initiative]
name    = "On Initiative"
trigger = { on = "initiative" }
steps   = [{ strike = "fist" }]

[ability.on-foe-acts]
name    = "On Foe Acts"
trigger = { on = "foe-acts", cost = "reaction", traits = ["manipulate", "move"], ranged = true }
steps   = [{ strike = "fist" }]

[ability.on-ally-gains]
name    = "On Ally Gains"
trigger = { on = "ally-gains", cost = "reaction", conditions = ["frightened", "stupefied"] }
steps   = [{ strike = "fist" }]

[ability.on-stood-up]
name    = "On Stood Up"
trigger = { on = "stood-up", cost = "reaction" }
steps   = [{ strike = "fist" }]

[ability.on-loyalty]
name    = "On Loyalty"
trigger = { on = "loyalty" }
steps   = [{ strike = "fist" }]
`

func TestLoadReadsTriggers(t *testing.T) {
	loaded, err := creature.Load(writeLayers(t, map[string]string{"hand.toml": minimalHand + triggeredAbilities}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := map[string]creature.Trigger{
		"on-targeted":   {On: creature.TriggerTargeted, Cost: creature.CostReaction, Attack: creature.AttackMeleeStrike},
		"on-attacked":   {On: creature.TriggerAttacked, Cost: creature.CostReaction, Attack: creature.AttackStrike, Degrees: []string{"critical-success"}},
		"on-saved":      {On: creature.TriggerSaved, Cost: creature.CostFree, Degrees: []string{"critical-failure", "failure"}, Damaging: true},
		"on-damaged":    {On: creature.TriggerDamaged, Melee: true},
		"on-turn-start": {On: creature.TriggerTurnStart, Cost: creature.CostFree, HPAtMost: 75},
		"on-initiative": {On: creature.TriggerInitiative},
		"on-foe-acts":   {On: creature.TriggerFoeActs, Cost: creature.CostReaction, Traits: []string{"manipulate", "move"}, Ranged: true},
		"on-ally-gains": {On: creature.TriggerAllyGains, Cost: creature.CostReaction, Conditions: []string{"frightened", "stupefied"}},
		"on-stood-up":   {On: creature.TriggerStoodUp, Cost: creature.CostReaction},
		"on-loyalty":    {On: creature.TriggerLoyalty},
	}
	for id, wantTrigger := range want {
		ability := loaded.Abilities[id]
		if ability.Trigger == nil {
			t.Errorf("%s trigger = nil, want %+v", id, wantTrigger)
			continue
		}
		if !reflect.DeepEqual(*ability.Trigger, wantTrigger) {
			t.Errorf("%s trigger = %+v, want %+v", id, *ability.Trigger, wantTrigger)
		}
		if low, high := ability.CostRange(); low != 0 || high != 0 {
			t.Errorf("%s CostRange() = %d, %d, want 0, 0", id, low, high)
		}
	}
	if !reflect.DeepEqual(loaded.ExtraReactions, map[string]int{"on-attacked": 2}) {
		t.Errorf("ExtraReactions = %v, want on-attacked 2", loaded.ExtraReactions)
	}
}

func TestLoadRejectsBadTriggers(t *testing.T) {
	cases := []struct {
		name     string
		keys     string
		extra    string
		wantText string
	}{
		{name: "unknown on", keys: `trigger = { on = "sneezed" }`, wantText: `ability.bad.trigger: on "sneezed": want targeted`},
		{name: "unknown cost", keys: `trigger = { on = "damaged", cost = "action" }`, wantText: `ability.bad.trigger: cost "action": want reaction, free, or leave it unset`},
		{name: "actions beside a trigger", keys: "actions = 1\n" + `trigger = { on = "damaged" }`, wantText: "ability.bad: a triggered ability sets no actions, max_actions, free or turn_ends"},
		{name: "attack on damaged", keys: `trigger = { on = "damaged", attack = "strike" }`, wantText: `ability.bad.trigger: attack "strike": want strike or melee-strike, on targeted or attacked`},
		{name: "unknown degree word", keys: `trigger = { on = "attacked", degrees = ["great"] }`, wantText: `ability.bad.trigger: degree "great": want critical-failure`},
		{name: "damaging on attacked", keys: `trigger = { on = "attacked", damaging = true }`, wantText: "ability.bad.trigger: damaging belongs on saved"},
		{name: "melee on saved", keys: `trigger = { on = "saved", melee = true }`, wantText: "ability.bad.trigger: melee belongs on damaged"},
		{name: "foe-acts with neither traits nor ranged", keys: `trigger = { on = "foe-acts", cost = "reaction" }`, wantText: "ability.bad.trigger: foe-acts needs traits or ranged, and no other kind takes them"},
		{name: "traits on turn-start", keys: `trigger = { on = "turn-start", traits = ["move"] }`, wantText: "ability.bad.trigger: foe-acts needs traits or ranged, and no other kind takes them"},
		{name: "ally-gains with no conditions", keys: `trigger = { on = "ally-gains", cost = "reaction" }`, wantText: "ability.bad.trigger: ally-gains needs conditions, and no other kind takes them"},
		{name: "hp_at_most on initiative", keys: `trigger = { on = "initiative", hp_at_most = 75 }`, wantText: "ability.bad.trigger: hp_at_most must be positive, on turn-start"},
		{name: "extra reaction for a missing ability", keys: `trigger = { on = "damaged", cost = "reaction" }`, extra: "[extra_reaction]\nmissing = 1\n", wantText: "extra_reaction.missing: want an ability whose trigger costs a reaction"},
		{name: "extra reaction for a passive ability", keys: `trigger = { on = "damaged" }`, extra: "[extra_reaction]\nbad = 1\n", wantText: "extra_reaction.bad: want an ability whose trigger costs a reaction"},
		{name: "extra reaction count of 0", keys: `trigger = { on = "damaged", cost = "reaction" }`, extra: "[extra_reaction]\nbad = 0\n", wantText: "extra_reaction.bad: must be at least 1"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			err := loadAbility(t, testCase.keys, testCase.extra)
			if !errors.Is(err, creature.ErrInvalid) {
				t.Fatalf("Load error = %v, want ErrInvalid", err)
			}
			if !strings.Contains(err.Error(), testCase.wantText) {
				t.Errorf("Load error = %v, want text %q", err, testCase.wantText)
			}
		})
	}
}

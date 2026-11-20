package creature_test

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

const stepAbilities = `
[spellcasting]
attack = 20
dc = 30

[ability.spores]
name = "Destiny Spores"
actions = 2
traits = ["mental", "occult"]
area = { shape = "cone", size = 40 }

[[ability.spores.steps]]
save = "will"
dc = 39
basic = true
damage = [{ dice = "15d6", type = "mental" }]
failure = [{ damage = [{ dice = "1d6", type = "mental" }] }]

[ability.bolt]
name = "Bolt"
actions = 2

[[ability.bolt.steps]]
spell_attack = true
damage = [{ dice = "2d6", type = "fire" }]
success = [{ damage = [{ dice = "1d4", type = "fire" }] }]
`

func TestLoadReadsSteps(t *testing.T) {
	hand := strings.Replace(minimalHand, "perception = 3", "perception = 3\nclass_dc = 25", 1) + stepAbilities
	loaded, err := creature.Load(writeLayers(t, map[string]string{"hand.toml": hand}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	spores := loaded.Abilities["spores"]
	if spores.Area == nil || spores.Area.Shape != "cone" || spores.Area.Size != 40 {
		t.Errorf("spores area = %+v, want cone 40", spores.Area)
	}
	if len(spores.Steps) != 1 {
		t.Fatalf("spores steps = %d, want 1", len(spores.Steps))
	}
	save := spores.Steps[0]
	if save.Kind() != creature.StepSave {
		t.Errorf("Kind() = %q, want save", save.Kind())
	}
	if want := (dice.Expression{Count: 15, Sides: 6}); save.Damage[0].Roll != want {
		t.Errorf("damage roll = %+v, want %+v", save.Damage[0].Roll, want)
	}
	onFailure := save.For(1, false)
	if len(onFailure) != 1 || onFailure[0].Kind() != creature.StepDamage {
		t.Errorf("For(failure) = %+v, want one damage step", onFailure)
	}

	bolt := loaded.Abilities["bolt"].Steps[0]
	if bolt.Kind() != creature.StepSpellAttack {
		t.Errorf("bolt Kind() = %q, want spell_attack", bolt.Kind())
	}
	onCritical := bolt.For(3, true)
	if len(onCritical) != 1 || onCritical[0].Damage[0].Dice != "1d4" {
		t.Errorf("For(critical success, attack) = %+v, want the success list", onCritical)
	}
}

const conditionAbility = `
[ability.terror]
name = "Terror"
actions = 2

[[ability.terror.steps]]
save = "will"
dc = 20
failure = [{ condition = "frightened", value = 1 }]
critical_failure = [{ condition = "off-guard", duration = { until = "source-turn-end" } }]

[condition_adjust.frightened]
gain = -1
`

func TestLoadReadsConditionSteps(t *testing.T) {
	loaded, err := creature.Load(writeLayers(t, map[string]string{"hand.toml": minimalHand + conditionAbility}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	save := loaded.Abilities["terror"].Steps[0]
	onFailure := save.For(1, false)
	if len(onFailure) != 1 || onFailure[0].Kind() != creature.StepCondition || onFailure[0].Condition != "frightened" || onFailure[0].Value != 1 {
		t.Errorf("For(failure) = %+v, want one frightened 1 step", onFailure)
	}
	onCriticalFailure := save.For(0, false)
	if len(onCriticalFailure) != 1 || onCriticalFailure[0].Condition != "off-guard" {
		t.Fatalf("For(critical failure) = %+v, want one off-guard step", onCriticalFailure)
	}
	if duration := onCriticalFailure[0].Duration; duration == nil || duration.Until != "source-turn-end" {
		t.Errorf("off-guard duration = %+v, want until source-turn-end", duration)
	}
	if adjustment := loaded.ConditionAdjustments["frightened"]; adjustment.Gain != -1 {
		t.Errorf("frightened adjustment = %+v, want gain -1", adjustment)
	}

	scout, err := creature.Load(writeLayers(t, map[string]string{"import.toml": scoutBuild}))
	if err != nil {
		t.Fatalf("Load scout: %v", err)
	}
	if bomb := scout.Strikes["fire-bomb"]; bomb.Range != 20 || bomb.Melee() {
		t.Errorf("fire-bomb range = %d, Melee() = %v, want 20 and false", bomb.Range, bomb.Melee())
	}
	if !scout.Strikes["needle"].Melee() {
		t.Error("needle Melee() = false, want true")
	}
}

func TestLoadReadsConditionValueDice(t *testing.T) {
	ability := "\n[ability.drain]\nname = \"Drain\"\nactions = 1\nsteps = [{ condition = \"drained\", value_dice = \"1d4\", stack = true }]\n"
	loaded, err := creature.Load(writeLayers(t, map[string]string{"hand.toml": minimalHand + ability}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	step := loaded.Abilities["drain"].Steps[0]
	if want := (dice.Expression{Count: 1, Sides: 4}); step.ValueRoll != want || !step.Stack {
		t.Errorf("drain step ValueRoll, Stack = %+v, %v, want %+v, true", step.ValueRoll, step.Stack, want)
	}
}

func TestLoadReadsStunnedForDuration(t *testing.T) {
	ability := "\n[ability.daze]\nname = \"Daze\"\nactions = 1\nsteps = [{ condition = \"stunned\", duration = { rounds = 1 } }]\n"
	loaded, err := creature.Load(writeLayers(t, map[string]string{"hand.toml": minimalHand + ability}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	step := loaded.Abilities["daze"].Steps[0]
	if step.Condition != "stunned" || step.Value != 0 || step.Duration == nil || step.Duration.Rounds != 1 {
		t.Errorf("daze step = %+v, want stunned with no value for 1 round", step)
	}
}

func TestLoadReadsReduceDisruptsAndFastHealing(t *testing.T) {
	hand := strings.Replace(minimalHand, "perception = 3", "perception = 3\nfast_healing = 10", 1) +
		"\n[ability.calm]\nname = \"Calm\"\nactions = 1\nsteps = [{ condition = \"frightened\", reduce = 1 }, { condition = \"stunned\", reduce = 2 }, { strike = \"fist\", disrupts = [\"manipulate\"] }]\n"
	loaded, err := creature.Load(writeLayers(t, map[string]string{"hand.toml": hand}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	steps := loaded.Abilities["calm"].Steps
	if steps[0].Reduce != 1 || steps[1].Condition != "stunned" || steps[1].Reduce != 2 || !slices.Equal(steps[2].Disrupts, []string{"manipulate"}) {
		t.Errorf("calm steps = %+v, want reduce 1, stunned reduce 2, disrupts manipulate", steps)
	}
	if loaded.FastHealing != 10 {
		t.Errorf("FastHealing = %d, want 10", loaded.FastHealing)
	}
}

func TestLoadRejectsBadConditions(t *testing.T) {
	cases := []struct {
		name     string
		step     string
		extra    string
		wantText string
	}{
		{name: "unknown condition", step: `condition = "sad"`, wantText: `unknown condition "sad"`},
		{name: "valued condition with no value", step: `condition = "frightened"`, wantText: "frightened needs a value of at least 1"},
		{name: "unvalued condition with a value", step: `condition = "off-guard", value = 1`, wantText: "off-guard takes no value"},
		{name: "rounds and until", step: `condition = "off-guard", duration = { rounds = 1, until = "source-turn-end" }`, wantText: "set rounds or until, not both"},
		{name: "bad boundary", step: `condition = "off-guard", duration = { until = "tomorrow" }`, wantText: `until "tomorrow"`},
		{name: "value on a save", step: `save = "will", dc = 20, value = 1`, wantText: "value belongs on a condition"},
		{name: "value and value_dice", step: `condition = "drained", value = 1, value_dice = "1d4"`, wantText: "set value or value_dice, not both"},
		{name: "stack on an unvalued condition", step: `condition = "off-guard", stack = true`, wantText: "off-guard takes no value"},
		{name: "value_dice on a heal", step: `heal = "1d4", value_dice = "1d4"`, wantText: "value belongs on a condition"},
		{name: "stunned with neither value nor duration", step: `condition = "stunned"`, wantText: "stunned needs a value or a duration"},
		{name: "slowed with a duration and no value", step: `condition = "slowed", duration = { rounds = 1 }`, wantText: "slowed needs a value of at least 1"},
		{name: "stack on stunned with no value", step: `condition = "stunned", duration = { rounds = 1 }, stack = true`, wantText: "stack needs a value"},
		{name: "unless names no condition", step: `strike = "fist", unless = "sad"`, wantText: `unless "sad": not a condition`},
		{name: "adjustment for no condition", step: `strike = "fist"`, extra: "[condition_adjust.sad]\ngain = 1\n", wantText: "condition_adjust.sad: not a condition"},
		{name: "adjustment with neither key", step: `strike = "fist"`, extra: "[condition_adjust.frightened]\n", wantText: "sets neither gain nor decay"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			ability := fmt.Sprintf("\n[ability.bad]\nname = \"Bad\"\nactions = 1\nsteps = [{ %s }]\n\n%s", testCase.step, testCase.extra)
			_, err := creature.Load(writeLayers(t, map[string]string{"hand.toml": minimalHand + ability}))
			if !errors.Is(err, creature.ErrInvalid) {
				t.Fatalf("Load error = %v, want ErrInvalid", err)
			}
			if !strings.Contains(err.Error(), testCase.wantText) {
				t.Errorf("Load error = %v, want text %q", err, testCase.wantText)
			}
		})
	}
}

func TestLoadRejectsBadSteps(t *testing.T) {
	const fireDamage = `damage = [{ dice = "1d6", type = "fire" }]`
	cases := []struct {
		name     string
		extra    string
		step     string
		wantText string
	}{
		{name: "two heads", step: `strike = "fist", save = "will", dc = 20`, wantText: "want exactly one of"},
		{name: "save with no dc", step: `save = "will"`, wantText: "a save needs dc, or dc_from class or spell"},
		{name: "dc and dc_from", step: `save = "will", dc = 20, dc_from = "class"`, wantText: "set dc or dc_from, not both"},
		{name: "dc_from class with no class dc", step: `save = "will", dc_from = "class"`, wantText: "dc_from class needs a class DC"},
		{name: "damage on a save that is not basic", step: `save = "will", dc = 20, ` + fireDamage, wantText: "damage on a save needs basic = true"},
		{name: "basic with no damage", step: `save = "will", dc = 20, basic = true`, wantText: "basic needs damage"},
		{name: "degree list on a damage step", step: fireDamage + `, failure = [{ strike = "fist" }]`, wantText: "failure needs a strike, save, spell_attack or check"},
		{name: "unknown strike in a degree list", step: `strike = "fist", success = [{ strike = "kick" }]`, wantText: `steps[0].success[0]: unknown strike "kick"`},
		{name: "bad area shape", extra: `area = { shape = "square", size = 10 }`, step: `strike = "fist"`, wantText: `shape "square"`},
		{name: "bad target", step: `strike = "fist", target = "ally"`, wantText: `target "ally"`},
		{name: "shift on a strike", step: `strike = "fist", shift = 1`, wantText: "belong on a save"},
		{name: "disrupts on a save", step: `save = "will", dc = 20, disrupts = ["manipulate"]`, wantText: "extra_dice, extra_damage, counts, penalty and disrupts belong on a strike"},
		{name: "reduce on a heal", step: `heal = "10", reduce = 1`, wantText: "reduce belongs on a condition and must be positive"},
		{name: "negative reduce", step: `condition = "frightened", reduce = -1`, wantText: "reduce belongs on a condition and must be positive"},
		{name: "reduce with a value", step: `condition = "frightened", reduce = 1, value = 1`, wantText: "reduce needs a valued condition and no value, stack or duration"},
		{name: "reduce with stack", step: `condition = "frightened", reduce = 1, stack = true`, wantText: "reduce needs a valued condition and no value, stack or duration"},
		{name: "reduce with a duration", step: `condition = "frightened", reduce = 1, duration = { rounds = 1 }`, wantText: "reduce needs a valued condition and no value, stack or duration"},
		{name: "reduce on off-guard", step: `condition = "off-guard", reduce = 1`, wantText: "reduce needs a valued condition and no value, stack or duration"},
		{name: "reroll twice", step: `reroll = "twice", rolls = ["attack"], uses = 1`, wantText: `reroll "twice": want worse or better`},
		{name: "reroll with no rolls", step: `reroll = "worse", uses = 1`, wantText: "rolls: want one or more of attack, save, skill"},
		{name: "reroll of an unknown roll", step: `reroll = "worse", rolls = ["attack", "damage"], uses = 1`, wantText: "rolls: want one or more of attack, save, skill"},
		{name: "reroll with neither uses nor a duration", step: `reroll = "better", rolls = ["save"]`, wantText: "reroll needs uses or a duration"},
		{name: "uses on a bonus to a save", step: `bonus = { selector = "will", type = "status", value = 1 }, uses = 1`, wantText: "uses on a bonus needs selector ac"},
		{name: "bonus with value 0", step: `bonus = { selector = "ac", type = "circumstance", value = 0 }, uses = 1`, wantText: "bonus: want a known selector and type, a value other than 0, and no per_condition"},
		{name: "uses on a damage step", step: fireDamage + `, uses = 1`, wantText: "uses belongs on one of reroll, bonus, grant and must not be negative"},
		{name: "grant fire", step: `grant = "fire", uses = 1`, wantText: `grant "fire": want one of adamantine, cold-iron, holy, silver, unholy, with uses of at least 1`},
		{name: "grant with no uses", step: `grant = "holy", duration = { until = "target-turn-end" }`, wantText: `grant "holy": want one of adamantine, cold-iron, holy, silver, unholy, with uses of at least 1`},
		{name: "rolls on a heal", step: `heal = "10", rolls = ["attack"]`, wantText: "rolls belongs on a reroll"},
		{name: "recovery_dc with no duration", step: `recovery_dc = 10`, wantText: "recovery_dc must be at least 1 and needs a duration"},
		{name: "negative recovery_dc", step: `recovery_dc = -1, duration = { rounds = 1 }`, wantText: "recovery_dc must be at least 1 and needs a duration"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			ability := fmt.Sprintf("\n[ability.bad]\nname = \"Bad\"\nactions = 1\n%s\nsteps = [{ %s }]\n", testCase.extra, testCase.step)
			_, err := creature.Load(writeLayers(t, map[string]string{"hand.toml": minimalHand + ability}))
			if !errors.Is(err, creature.ErrInvalid) {
				t.Fatalf("Load error = %v, want ErrInvalid", err)
			}
			if !strings.Contains(err.Error(), testCase.wantText) {
				t.Errorf("Load error = %v, want text %q", err, testCase.wantText)
			}
		})
	}
}

func TestLoadRejectsBadChecks(t *testing.T) {
	const fireDamage = `damage = [{ dice = "1d6", type = "fire" }]`
	const wantDefense = "a check needs one of dc and against (fortitude, reflex, will, perception)"
	cases := []struct {
		name     string
		skills   string
		step     string
		wantText string
	}{
		{"neither dc nor against", "[skill]\nintimidation = 10\n", `check = "intimidation"`, wantDefense},
		{"both dc and against", "[skill]\nintimidation = 10\n", `check = "intimidation", dc = 20, against = "will"`, wantDefense},
		{"against ac", "[skill]\nintimidation = 10\n", `check = "intimidation", against = "ac"`, wantDefense},
		{"a skill with no number", "[skill]\nintimidation = 10\n", `check = "athletics", against = "will"`, `check "athletics": want a skill the creature has a number for`},
		{"not a skill", "[skill]\nintimidation = 10\n", `check = "piloting", dc = 15`, `check "piloting": want a skill the creature has a number for`},
		{"a pc with no build and no skill table", "", `check = "intimidation", against = "will"`, `check "intimidation": want a skill the creature has a number for`},
		{"against on a save", "", `save = "will", dc = 20, against = "will"`, "against belongs on a check"},
		{"dc on a damage step", "", fireDamage + `, dc = 20`, "dc belongs on a save or a check"},
		{"dc_from on a check", "[skill]\nintimidation = 10\n", `check = "intimidation", dc = 20, dc_from = "class"`, "dc_from, basic and shift belong on a save"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			ability := fmt.Sprintf("\n%s\n[ability.bad]\nname = \"Bad\"\nactions = 1\nsteps = [{ %s }]\n", testCase.skills, testCase.step)
			_, err := creature.Load(writeLayers(t, map[string]string{"hand.toml": minimalHand + ability}))
			if !errors.Is(err, creature.ErrInvalid) {
				t.Fatalf("Load error = %v, want ErrInvalid", err)
			}
			if !strings.Contains(err.Error(), testCase.wantText) {
				t.Errorf("Load error = %v, want text %q", err, testCase.wantText)
			}
		})
	}
}

func TestLoadReadsCheckSteps(t *testing.T) {
	hand := "schema = 1\n\n[ability.demoralize]\nname = \"Demoralize\"\nactions = 1\nsteps = [{ check = \"intimidation\", against = \"will\", success = [{ condition = \"frightened\", value = 1 }] }, { check = \"athletics\", dc = 25 }]\n"
	loaded, err := creature.Load(writeLayers(t, map[string]string{"import.toml": fighterBuild, "hand.toml": hand}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	steps := loaded.Abilities["demoralize"].Steps
	if steps[0].Kind() != creature.StepCheck || steps[0].Against != "will" || len(steps[0].Success) != 1 || steps[1].DC != 25 {
		t.Errorf("steps = %+v, want a check against will with a success list, and a check at DC 25", steps)
	}
}

const splashBomb = `
[build.weapon.splash-bomb]
name       = "Splash Bomb"
category   = "martial"
base       = "alchemical-bomb"
dice       = 2
die        = 8
type       = "fire"
range      = 20
traits     = ["alchemical", "bomb", "splash"]
persistent = { dice = "3", type = "fire" }
splash     = 3

[build.modifier.splashy]
selector      = "splash"
value         = 1
weapon_traits = ["splash"]
`

func TestLoadReadsDamageCategories(t *testing.T) {
	loaded, err := creature.Load(writeLayers(t, map[string]string{"import.toml": scoutBuild + splashBomb}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	damage := loaded.Strikes["splash-bomb"].Damage
	if len(damage) != 3 {
		t.Fatalf("damage = %+v, want three parts", damage)
	}
	if damage[0].Type != "fire" || damage[0].Category != "" {
		t.Errorf("damage[0] = %s %s, want fire with no category", damage[0].Type, damage[0].Category)
	}
	for index, want := range []string{"3 fire persistent", "4 fire splash"} {
		part := damage[index+1]
		if got := part.Roll.String() + " " + part.Type + " " + part.Category; got != want {
			t.Errorf("damage[%d] = %q, want %q", index+1, got, want)
		}
	}
}

func TestLoadRejectsBadCategories(t *testing.T) {
	cases := []struct {
		name     string
		file     string
		layer    string
		wantText string
	}{
		{
			name:     "unknown category",
			file:     "hand.toml",
			layer:    minimalHand + "\n[ability.bad]\nname = \"Bad\"\nactions = 1\nsteps = [{ damage = [{ dice = \"1d6\", type = \"fire\", category = \"precision\" }] }]\n",
			wantText: `category "precision"`,
		},
		{
			name:     "strike with only splash",
			file:     "hand.toml",
			layer:    minimalHand + "\n[strike.spray]\nname = \"Spray\"\nbonus = 5\ndamage = [{ dice = \"3\", type = \"acid\", category = \"splash\" }]\n",
			wantText: "needs a damage part with no category",
		},
		{
			name:     "heal that does not parse",
			file:     "hand.toml",
			layer:    minimalHand + "\n[ability.bad]\nname = \"Bad\"\nactions = 1\nsteps = [{ heal = \"xd6\" }]\n",
			wantText: ".heal:",
		},
		{
			name:     "weapon persistent with an unknown type",
			file:     "import.toml",
			layer:    scoutBuild + "\n[build.weapon.sad-bomb]\nname = \"Sad Bomb\"\ncategory = \"martial\"\ndice = 1\ndie = 6\ntype = \"fire\"\npersistent = { dice = \"1d6\", type = \"sad\" }\n",
			wantText: "persistent: want parsable dice",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := creature.Load(writeLayers(t, map[string]string{testCase.file: testCase.layer}))
			if !errors.Is(err, creature.ErrInvalid) {
				t.Fatalf("Load error = %v, want ErrInvalid", err)
			}
			if !strings.Contains(err.Error(), testCase.wantText) {
				t.Errorf("Load error = %v, want text %q", err, testCase.wantText)
			}
		})
	}
}

const defensesAndRiders = `
[defense.weakness.holy]
value = 15

[defense.resist.all]
value  = 12
except = ["holy", "cold-iron"]

[defense.adaptive]
value = 25
types = ["acid", "electricity", "sonic"]

[strike.touch]
name    = "Terrifying Touch"
bonus   = 10
damage  = [{ dice = "1d6", type = "void" }]
success = [{ save = "will", dc = 30, unless = "frightened", failure = [{ condition = "frightened", value = 2 }] }]

[ability.flask]
name    = "Inflammation Flask"
actions = 1
steps   = [{ weaken = { weakness = 3, types = ["acid", "fire", "slashing"] }, duration = { rounds = 3 } }]

[modifier.brute]
selector      = "damage"
type          = "circumstance"
value         = 3
per_condition = "frightened"
`

func TestLoadReadsDefensesAndRiders(t *testing.T) {
	loaded, err := creature.Load(writeLayers(t, map[string]string{"hand.toml": minimalHand + defensesAndRiders}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	defense := loaded.Defense
	if defense.Weakness["holy"].Value != 15 {
		t.Errorf("weakness holy = %d, want 15", defense.Weakness["holy"].Value)
	}
	if all := defense.Resist["all"]; all.Value != 12 || len(all.Except) != 2 {
		t.Errorf("resist all = %+v, want 12 except holy and cold-iron", all)
	}
	if adaptive := defense.Adaptive; adaptive == nil || adaptive.Value != 25 || len(adaptive.Types) != 3 {
		t.Errorf("adaptive = %+v, want 25 over three types", adaptive)
	}

	riders := loaded.Strikes["touch"].Success
	if len(riders) != 1 || riders[0].Kind() != creature.StepSave || riders[0].Unless != "frightened" {
		t.Fatalf("touch success = %+v, want one Will save unless frightened", riders)
	}
	if failure := riders[0].Failure; len(failure) != 1 || failure[0].Condition != "frightened" || failure[0].Value != 2 {
		t.Errorf("touch save failure = %+v, want frightened 2", failure)
	}

	step := loaded.Abilities["flask"].Steps[0]
	if step.Kind() != creature.StepWeaken || step.Weaken.Weakness != 3 || len(step.Weaken.Types) != 3 || step.Duration.Rounds != 3 {
		t.Errorf("flask step = %+v, want weaken 3 over three types for 3 rounds", step)
	}
	if brute := loaded.Modifiers["brute"]; brute.PerCondition != "frightened" {
		t.Errorf("brute per_condition = %q, want frightened", brute.PerCondition)
	}
}

func TestLoadRejectsBadDefenses(t *testing.T) {
	cases := []struct {
		name     string
		extra    string
		wantText string
	}{
		{name: "weakness to no trait", extra: "\n[defense.weakness.chaos]\nvalue = 5\n", wantText: "one of adamantine, cold-iron"},
		{name: "adaptive to no damage type", extra: "\n[defense.adaptive]\nvalue = 25\ntypes = [\"acid\", \"sad\"]\n", wantText: `"sad" is not a damage type`},
		{name: "weaken with neither number", extra: "\n[ability.bad]\nname = \"Bad\"\nactions = 1\nsteps = [{ weaken = { types = [\"fire\"] } }]\n", wantText: "want a weakness or a resistance above 0"},
		{name: "duration on a damage step", extra: "\n[ability.bad]\nname = \"Bad\"\nactions = 1\nsteps = [{ damage = [{ dice = \"1d6\", type = \"fire\" }], duration = { rounds = 1 } }]\n", wantText: "duration belongs on one of condition, weaken, reroll, bonus"},
		{name: "per_condition on ac", extra: "\n[modifier.bad]\nselector = \"ac\"\nvalue = 1\nper_condition = \"frightened\"\n", wantText: "per_condition needs a condition"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := creature.Load(writeLayers(t, map[string]string{"hand.toml": minimalHand + testCase.extra}))
			if !errors.Is(err, creature.ErrInvalid) {
				t.Fatalf("Load error = %v, want ErrInvalid", err)
			}
			if !strings.Contains(err.Error(), testCase.wantText) {
				t.Errorf("Load error = %v, want text %q", err, testCase.wantText)
			}
		})
	}
}

const criticalWeapons = `
[build.weapon.crusher]
name             = "Crusher"
category         = "martial"
group            = "hammer"
dice             = 1
die              = 12
type             = "bludgeoning"
critical_success = [{ condition = "frightened", value = 2 }]

[build.weapon.training-hammer]
name     = "Training Hammer"
category = "advanced"
group    = "hammer"
dice     = 1
die      = 8
type     = "bludgeoning"

[build.weapon.dagger]
name     = "Dagger"
category = "simple"
group    = "knife"
dice     = 1
die      = 4
type     = "piercing"
`

func TestCriticalSpecializationFromBuild(t *testing.T) {
	layer := strings.Replace(fighterBuild, "[build]\n", "[build]\ncritical_specialization = \"master\"\n", 1) + criticalWeapons
	loaded, err := creature.Load(writeLayers(t, map[string]string{"import.toml": layer}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	riders := loaded.Strikes["crusher"].CriticalSuccess
	if len(riders) != 2 {
		t.Fatalf("crusher critical riders = %+v, want two", riders)
	}
	if riders[0].Condition != "frightened" {
		t.Errorf("first rider = %+v, want the weapon's own frightened", riders[0])
	}
	if riders[1].Save != "fortitude" || riders[1].DCFrom != "class" {
		t.Errorf("second rider = %+v, want a Fortitude save against the class DC", riders[1])
	}
	for _, id := range []string{"training-hammer", "dagger"} {
		if riders := loaded.Strikes[id].CriticalSuccess; len(riders) != 0 {
			t.Errorf("%s critical riders = %+v, want none (untrained hammer; knife has no row)", id, riders)
		}
	}
}

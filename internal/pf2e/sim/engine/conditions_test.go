package engine

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

func giveCondition(combatant *Combatant, name string, value int) {
	combatant.conditions = append(combatant.conditions, &conditionInstance{name: name, value: value})
}

func conditionStep(name string, value int, duration *creature.Duration) creature.Step {
	return creature.Step{Condition: name, Value: value, Duration: duration}
}

var (
	meleeAttack  = []string{"attack", "melee-attack"}
	rangedAttack = []string{"attack", "ranged-attack"}
	meleeDamage  = []string{"damage", "melee-damage"}
)

func TestConditionPenalties(t *testing.T) {
	type check struct {
		label     string
		selectors []string
		want      int
	}
	cases := []struct {
		condition string
		value     int
		wantAC    int
		checks    []check
	}{
		{"frightened", 2, -2, []check{{"will", []string{"will", "saving-throw"}, -2}, {"melee attack", meleeAttack, -2}, {"spell DC", []string{"spell-dc"}, -2}}},
		{"clumsy", 1, -1, []check{{"reflex", []string{"reflex", "saving-throw"}, -1}, {"ranged attack", rangedAttack, -1}, {"melee attack", meleeAttack, 0}}},
		{"enfeebled", 2, 0, []check{{"melee attack", meleeAttack, -2}, {"melee damage", meleeDamage, -2}, {"ranged attack", rangedAttack, 0}}},
		{"stupefied", 1, 0, []check{{"will", []string{"will", "saving-throw"}, -1}, {"spell attack", []string{"spell-attack"}, -1}, {"fortitude", []string{"fortitude", "saving-throw"}, 0}}},
		{"drained", 1, 0, []check{{"fortitude", []string{"fortitude", "saving-throw"}, -1}}},
		{"off-guard", 0, -2, nil},
		{"prone", 0, -2, []check{{"melee attack", meleeAttack, -2}, {"spell attack", []string{"spell-attack"}, -2}}},
		{"encumbered", 0, -1, nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.condition, func(t *testing.T) {
			bearer := testCombatant("bearer", SideParty, creature.KindPC, 50, 20)
			giveCondition(bearer, testCase.condition, testCase.value)
			fight, _ := testFight(script(t), bearer)

			if got := fight.armorClass(bearer, nil) - 20; got != testCase.wantAC {
				t.Errorf("AC change = %d, want %d", got, testCase.wantAC)
			}
			for _, check := range testCase.checks {
				if got := fight.modifierTotal(bearer, check.selectors, nil); got != check.want {
					t.Errorf("%s = %d, want %d", check.label, got, check.want)
				}
			}
		})
	}
}

func TestConditionsStackByType(t *testing.T) {
	cases := []struct {
		name       string
		conditions map[string]int
		wantAC     int
	}{
		{"two status penalties keep the worst", map[string]int{"frightened": 1, "sickened": 2}, 18},
		{"status and circumstance both count", map[string]int{"off-guard": 0, "frightened": 1}, 17},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			bearer := testCombatant("bearer", SideParty, creature.KindPC, 50, 20)
			for name, value := range testCase.conditions {
				giveCondition(bearer, name, value)
			}
			fight, _ := testFight(script(t), bearer)

			if got := fight.armorClass(bearer, nil); got != testCase.wantAC {
				t.Errorf("AC = %d, want %d", got, testCase.wantAC)
			}
		})
	}
}

func conditionChanges(events []Event) []ConditionChanged {
	return eventsOf[ConditionChanged](events)
}

func TestFrightenedDecays(t *testing.T) {
	t.Run("one per turn", func(t *testing.T) {
		source := testCombatant("source", SideEnemies, creature.KindNPC, 50, 20)
		bearer := testCombatant("bearer", SideParty, creature.KindPC, 50, 20)
		fight, events := testFight(script(t), source, bearer)

		fight.applyCondition(source, bearer, conditionStep("frightened", 2, nil))
		fight.endTurn(bearer)
		if got := bearer.Condition("frightened"); got != 1 {
			t.Errorf("frightened after one end of turn = %d, want 1", got)
		}
		fight.endTurn(bearer)
		if got := bearer.Condition("frightened"); got != 0 {
			t.Errorf("frightened after two ends of turn = %d, want 0", got)
		}

		want := []ConditionChanged{
			{Combatant: "bearer", Condition: "frightened", From: 0, To: 2},
			{Combatant: "bearer", Condition: "frightened", From: 2, To: 1},
			{Combatant: "bearer", Condition: "frightened", From: 1, To: 0},
		}
		if got := conditionChanges(events.events); !slices.Equal(got, want) {
			t.Errorf("changes = %+v, want %+v", got, want)
		}
	})
	t.Run("decay 2", func(t *testing.T) {
		source := testCombatant("source", SideEnemies, creature.KindNPC, 50, 20)
		bearer := testCombatant("bearer", SideParty, creature.KindPC, 50, 20)
		bearer.Creature.ConditionAdjustments = map[string]creature.ConditionAdjustment{"frightened": {Decay: 2}}
		fight, _ := testFight(script(t), source, bearer)

		fight.applyCondition(source, bearer, conditionStep("frightened", 2, nil))
		fight.endTurn(bearer)
		if got := bearer.Condition("frightened"); got != 0 {
			t.Errorf("frightened after one end of turn = %d, want 0", got)
		}
	})
}

func TestConditionGain(t *testing.T) {
	cases := []struct {
		name  string
		value int
		want  int
	}{
		{"frightened 1 becomes nothing", 1, 0},
		{"frightened 2 becomes 1", 2, 1},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			source := testCombatant("source", SideEnemies, creature.KindNPC, 50, 20)
			bearer := testCombatant("bearer", SideParty, creature.KindPC, 50, 20)
			bearer.Creature.ConditionAdjustments = map[string]creature.ConditionAdjustment{"frightened": {Gain: -1}}
			fight, events := testFight(script(t), source, bearer)

			fight.applyCondition(source, bearer, conditionStep("frightened", testCase.value, nil))

			if got := bearer.Condition("frightened"); got != testCase.want {
				t.Errorf("frightened = %d, want %d", got, testCase.want)
			}
			if testCase.want == 0 && len(conditionChanges(events.events)) != 0 {
				t.Errorf("changes = %+v, want none", conditionChanges(events.events))
			}
		})
	}
}

func TestDurationInRounds(t *testing.T) {
	source := testCombatant("source", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 50, 20)
	fight, _ := testFight(script(t), source, target)
	fight.current = source

	fight.applyCondition(source, target, conditionStep("clumsy", 1, &creature.Duration{Rounds: 1}))
	fight.endTurn(source)
	fight.startTurn(target)
	fight.endTurn(target)
	if got := target.Condition("clumsy"); got != 1 {
		t.Fatalf("clumsy after the target's turn = %d, want 1", got)
	}
	fight.startTurn(source)
	if got := target.Condition("clumsy"); got != 0 {
		t.Errorf("clumsy at the source's next start of turn = %d, want 0", got)
	}
}

func TestDurationUntilSourceTurnEnd(t *testing.T) {
	until := &creature.Duration{Until: "source-turn-end"}
	t.Run("applied during the source's turn", func(t *testing.T) {
		source := testCombatant("source", SideEnemies, creature.KindNPC, 50, 20)
		target := testCombatant("target", SideParty, creature.KindPC, 50, 20)
		fight, _ := testFight(script(t), source, target)
		fight.current = source

		fight.applyCondition(source, target, conditionStep("off-guard", 0, until))
		fight.endTurn(source)
		fight.current = target
		fight.startTurn(target)
		fight.endTurn(target)
		if got := target.Condition("off-guard"); got != 1 {
			t.Fatalf("off-guard before the source's next end of turn = %d, want 1", got)
		}
		fight.current = source
		fight.startTurn(source)
		fight.endTurn(source)
		if got := target.Condition("off-guard"); got != 0 {
			t.Errorf("off-guard after the source's next end of turn = %d, want 0", got)
		}
	})
	t.Run("applied during another creature's turn", func(t *testing.T) {
		source := testCombatant("source", SideEnemies, creature.KindNPC, 50, 20)
		target := testCombatant("target", SideParty, creature.KindPC, 50, 20)
		fight, _ := testFight(script(t), source, target)
		fight.current = target

		fight.applyCondition(source, target, conditionStep("off-guard", 0, until))
		fight.endTurn(target)
		if got := target.Condition("off-guard"); got != 1 {
			t.Fatalf("off-guard after the target's end of turn = %d, want 1", got)
		}
		fight.current = source
		fight.startTurn(source)
		fight.endTurn(source)
		if got := target.Condition("off-guard"); got != 0 {
			t.Errorf("off-guard after the source's end of turn = %d, want 0", got)
		}
	})
}

func TestDurationUntilTargetTurnStart(t *testing.T) {
	source := testCombatant("source", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 50, 20)
	fight, _ := testFight(script(t), source, target)
	fight.current = source

	fight.applyCondition(source, target, conditionStep("off-guard", 0, &creature.Duration{Until: "target-turn-start"}))
	fight.endTurn(source)
	fight.startTurn(source)
	if got := target.Condition("off-guard"); got != 1 {
		t.Fatalf("off-guard at the source's start of turn = %d, want 1", got)
	}
	fight.startTurn(target)
	if got := target.Condition("off-guard"); got != 0 {
		t.Errorf("off-guard at the target's start of turn = %d, want 0", got)
	}
}

func TestConditionImmunity(t *testing.T) {
	source := testCombatant("source", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 50, 20)
	target.Creature.Defense.Immune = []string{"frightened"}
	fight, events := testFight(script(t), source, target)

	fight.applyCondition(source, target, conditionStep("frightened", 2, nil))

	if got := target.Condition("frightened"); got != 0 {
		t.Errorf("frightened = %d, want 0", got)
	}
	if changes := conditionChanges(events.events); len(changes) != 0 {
		t.Errorf("changes = %+v, want none", changes)
	}
}

func TestEffectTraitImmunity(t *testing.T) {
	caster := testCombatant("caster", SideParty, creature.KindPC, 50, 20)
	avatar := testCombatant("avatar", SideEnemies, creature.KindNPC, 50, 20)
	avatar.Creature.Defense.Immune = []string{"death-effects"}
	fight, events := testFight(script(t), caster, avatar)

	fight.useAbility(abilityUse{actor: caster, ability: saveAbility("fortitude", 20, "death", "void"), target: avatar})

	want := []ImmuneToEffect{{Combatant: "avatar", Source: "caster", Ability: "test-save"}}
	if got := eventsOf[ImmuneToEffect](events.events); !slices.Equal(got, want) {
		t.Errorf("immunity events = %+v, want %+v", got, want)
	}
	if saves := eventsOf[SaveRolled](events.events); len(saves) != 0 {
		t.Errorf("saves = %+v, want none", saves)
	}
}

func TestUnlessSkipsTarget(t *testing.T) {
	caster := testCombatant("caster", SideEnemies, creature.KindNPC, 50, 20)
	scared := testCombatant("scared", SideParty, creature.KindPC, 50, 20)
	giveCondition(scared, "frightened", 1)
	calm := testCombatant("calm", SideParty, creature.KindPC, 50, 20)
	ability := creature.Ability{ID: "touch", Name: "Terrifying Touch", Actions: 2, Area: &creature.Area{Shape: "burst", Size: 10}, Steps: []creature.Step{
		{Save: "will", DC: 20, Unless: "frightened"},
	}}
	fight, events := testFight(script(t, 10), caster, scared, calm)

	fight.useAbility(abilityUse{actor: caster, ability: ability, target: scared})

	saves := eventsOf[SaveRolled](events.events)
	if len(saves) != 1 || saves[0].Combatant != "calm" {
		t.Errorf("saves = %+v, want one, from calm", saves)
	}
}

func TestConcealmentFlatCheck(t *testing.T) {
	cases := []struct {
		name              string
		attackerCondition string
		targetCondition   string
		dice              []int
		wantDC            int
		wantStrike        bool
	}{
		{"concealed target, die 4 misses", "", "concealed", []int{4}, 5, false},
		{"concealed target, die 5 rolls", "", "concealed", []int{5, 1}, 5, true},
		{"dazzled attacker", "dazzled", "", []int{5, 1}, 5, true},
		{"blinded attacker, die 10 misses", "blinded", "", []int{10}, 11, false},
		{"blinded attacker, die 11 rolls", "blinded", "", []int{11, 1}, 11, true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			attacker := testCombatant("attacker", SideParty, creature.KindPC, 50, 20)
			target := testCombatant("target", SideEnemies, creature.KindNPC, 50, 20)
			if testCase.attackerCondition != "" {
				giveCondition(attacker, testCase.attackerCondition, 0)
			}
			if testCase.targetCondition != "" {
				giveCondition(target, testCase.targetCondition, 0)
			}
			fight, events := testFight(script(t, testCase.dice...), attacker, target)
			fight.current = attacker

			fight.strike(attacker, target, fist(), strikeOptions{})

			checks := eventsOf[FlatChecked](events.events)
			if len(checks) != 1 || checks[0].DC != testCase.wantDC || checks[0].Reason != "concealment" {
				t.Errorf("flat checks = %+v, want one concealment check at DC %d", checks, testCase.wantDC)
			}
			if got := len(eventsOf[StrikeRolled](events.events)) == 1; got != testCase.wantStrike {
				t.Errorf("strike rolled = %v, want %v", got, testCase.wantStrike)
			}
			if attacker.AttacksThisTurn != 1 {
				t.Errorf("AttacksThisTurn = %d, want 1", attacker.AttacksThisTurn)
			}
		})
	}
}

const frighteningCaster = `schema = 1
name = "Frightening Caster"
kind = "npc"
level = 14
perception = 20

[defense]
hp = 500
ac = 36
fortitude = 25
reflex = 25
will = 25

[strike.claw]
name = "Claw"
bonus = 25
damage = [{ dice = "2d6", type = "slashing" }]

[ability.dread]
name = "Dread Spores"
actions = 3
traits = ["emotion", "fear", "mental"]
area = { shape = "cone", size = 40 }

[[ability.dread.steps]]
save = "will"
dc = 40
basic = true
damage = [{ dice = "2d6", type = "mental" }]
failure = [{ condition = "frightened", value = 1 }]
`

func TestConditionStepsFromALoadedLayer(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "frightening-caster")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "hand.toml"), []byte(frighteningCaster), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	run := func() []Event {
		caster, err := creature.Load(directory)
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		events := &recorder{}
		_, err = Run(Setup{
			Party:     loadTestCreatures(t, "dummy-fighter", "dummy-alchemist"),
			Enemies:   []*creature.Creature{caster},
			Roller:    dice.NewSeeded(7),
			Observer:  events,
			Decider:   FirstFitting{},
			MaxRounds: 2,
		})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		return events.events
	}

	first := run()
	if len(conditionChanges(first)) == 0 {
		t.Error("no ConditionChanged events, want at least one")
	}
	if second := run(); !reflect.DeepEqual(first, second) {
		t.Error("replaying the seed gave different events")
	}
}

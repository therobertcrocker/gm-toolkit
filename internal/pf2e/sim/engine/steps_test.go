package engine

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

func eventsOf[E Event](events []Event) []E {
	var found []E
	for _, event := range events {
		if typed, matches := event.(E); matches {
			found = append(found, typed)
		}
	}
	return found
}

func saveAbility(save string, dc int, traits ...string) creature.Ability {
	return creature.Ability{
		ID:      "test-save",
		Name:    "Test Save",
		Actions: 2,
		Traits:  traits,
		Steps:   []creature.Step{{Save: save, DC: dc}},
	}
}

func TestBasicSaveScalesDamage(t *testing.T) {
	cases := []struct {
		name         string
		die          int
		wantDamage   bool
		wantTotal    int
		wantCritical bool
	}{
		{"critical success", 20, false, 0, false},
		{"success", 15, true, 3, false},
		{"failure", 5, true, 7, false},
		{"natural 1 drops to critical failure", 1, true, 14, true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			caster := testCombatant("caster", SideEnemies, creature.KindNPC, 50, 20)
			target := testCombatant("target", SideParty, creature.KindPC, 100, 20)
			target.Creature.Defense.Will = 10
			ability := creature.Ability{ID: "mind-blast", Name: "Mind Blast", Actions: 2, Steps: []creature.Step{
				{Save: "will", DC: 20, Basic: true, Damage: []creature.DamagePart{damagePart(t, "2d6", "mental")}},
			}}
			fight, events := testFight(script(t, 3, 4, testCase.die), caster, target)

			fight.useAbility(caster, ability, target)

			dealt := eventsOf[DamageDealt](events.events)
			if !testCase.wantDamage {
				if len(dealt) != 0 {
					t.Errorf("damage = %+v, want none", dealt)
				}
				return
			}
			if len(dealt) != 1 || dealt[0].Total != testCase.wantTotal || dealt[0].Critical != testCase.wantCritical {
				t.Errorf("damage = %+v, want total %d, critical %v", dealt, testCase.wantTotal, testCase.wantCritical)
			}
		})
	}
}

func TestAreaCatchesEveryStandingFoe(t *testing.T) {
	caster := testCombatant("caster", SideEnemies, creature.KindNPC, 50, 20)
	first := testCombatant("first", SideParty, creature.KindPC, 100, 20)
	downed := testCombatant("downed", SideParty, creature.KindPC, 100, 20)
	downed.HP, downed.Dying = 0, 1
	third := testCombatant("third", SideParty, creature.KindPC, 100, 20)
	ability := creature.Ability{ID: "cone", Name: "Cone", Actions: 2, Area: &creature.Area{Shape: "cone", Size: 40}, Steps: []creature.Step{
		{Save: "reflex", DC: 20, Basic: true, Damage: []creature.DamagePart{damagePart(t, "1d6", "fire")}},
	}}
	roller := script(t, 4, 10, 10)
	fight, events := testFight(roller, caster, first, downed, third)

	fight.useAbility(caster, ability, first)

	saves := eventsOf[SaveRolled](events.events)
	var savers []string
	for _, save := range saves {
		savers = append(savers, save.Combatant)
	}
	if want := []string{"first", "third"}; !slices.Equal(savers, want) {
		t.Errorf("savers = %v, want %v", savers, want)
	}
	if len(roller.values) != 0 {
		t.Errorf("roller has %d values left, want 0", len(roller.values))
	}
}

func TestSaveDCFromClassAndSpell(t *testing.T) {
	cases := []struct {
		name          string
		classModifier int
		wantClassDC   int
	}{
		{"no modifier", 0, 25},
		{"class-dc modifier +1", 1, 26},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			caster := testCombatant("caster", SideEnemies, creature.KindNPC, 50, 20)
			caster.Creature.ClassDC = 25
			caster.Creature.Spellcasting = &creature.Spellcasting{Attack: 15, DC: 30}
			if testCase.classModifier != 0 {
				caster.Creature.Modifiers = map[string]creature.CheckModifier{"focus": {Selector: "class-dc", Value: testCase.classModifier}}
			}
			target := testCombatant("target", SideParty, creature.KindPC, 100, 20)
			ability := creature.Ability{ID: "two-saves", Name: "Two Saves", Actions: 2, Steps: []creature.Step{
				{Save: "will", DCFrom: "class"},
				{Save: "will", DCFrom: "spell"},
			}}
			fight, events := testFight(script(t, 10, 10), caster, target)

			fight.useAbility(caster, ability, target)

			saves := eventsOf[SaveRolled](events.events)
			if len(saves) != 2 || saves[0].DC != testCase.wantClassDC || saves[1].DC != 30 {
				t.Errorf("saves = %+v, want DCs %d and 30", saves, testCase.wantClassDC)
			}
		})
	}
}

func TestSaveModifierAgainstTraits(t *testing.T) {
	cases := []struct {
		name      string
		traits    []string
		wantBonus int
	}{
		{"occult effect", []string{"occult", "mental"}, 12},
		{"arcane effect", []string{"arcane"}, 10},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			caster := testCombatant("caster", SideEnemies, creature.KindNPC, 50, 20)
			target := testCombatant("target", SideParty, creature.KindPC, 100, 20)
			target.Creature.Defense.Will = 10
			target.Creature.Modifiers = map[string]creature.CheckModifier{
				"occult-saves": {Selector: "saving-throw", Type: "status", Value: 2, Against: []string{"occult"}},
			}
			fight, events := testFight(script(t, 10), caster, target)

			fight.useAbility(caster, saveAbility("will", 20, testCase.traits...), target)

			saves := eventsOf[SaveRolled](events.events)
			if len(saves) != 1 || saves[0].Bonus != testCase.wantBonus {
				t.Errorf("saves = %+v, want bonus %d", saves, testCase.wantBonus)
			}
		})
	}
}

func TestDegreeShiftUpgradesSave(t *testing.T) {
	cases := []struct {
		name string
		die  int
		want Degree
	}{
		{"success becomes critical success", 12, CriticalSuccess},
		{"failure stays failure", 5, Failure},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			caster := testCombatant("caster", SideEnemies, creature.KindNPC, 50, 20)
			target := testCombatant("target", SideParty, creature.KindPC, 100, 20)
			target.Creature.Defense.Fortitude = 10
			target.Creature.DegreeShifts = map[string]creature.DegreeShift{
				"juggernaut": {Selector: "fortitude", From: "success", To: "critical-success"},
			}
			fight, events := testFight(script(t, testCase.die), caster, target)

			fight.useAbility(caster, saveAbility("fortitude", 20), target)

			saves := eventsOf[SaveRolled](events.events)
			if len(saves) != 1 || saves[0].Degree != testCase.want {
				t.Errorf("saves = %+v, want degree %v", saves, testCase.want)
			}
		})
	}
}

func TestStepShiftWorsensSave(t *testing.T) {
	caster := testCombatant("caster", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 100, 20)
	target.Creature.Defense.Fortitude = 10
	target.Creature.DegreeShifts = map[string]creature.DegreeShift{
		"juggernaut": {Selector: "fortitude", From: "success", To: "critical-success"},
	}
	ability := saveAbility("fortitude", 20)
	ability.Steps[0].Shift = -1
	fight, events := testFight(script(t, 12), caster, target)

	fight.useAbility(caster, ability, target)

	saves := eventsOf[SaveRolled](events.events)
	if len(saves) != 1 || saves[0].Degree != Failure {
		t.Errorf("saves = %+v, want a failure: the step's shift lands first and the saver's shift no longer matches", saves)
	}
}

func TestSpellAttackUsesPenaltyAndOutcomes(t *testing.T) {
	cases := []struct {
		name             string
		dice             []int
		wantTotal        int
		wantDegree       Degree
		wantStepDamage   int
		wantStepCritical bool
	}{
		{"success", []int{12, 3, 2}, 22, Success, 3, false},
		{"natural 20 runs the success list", []int{20, 3, 2}, 30, CriticalSuccess, 6, true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			caster := testCombatant("caster", SideEnemies, creature.KindNPC, 50, 20)
			caster.Creature.Spellcasting = &creature.Spellcasting{Attack: 15, DC: 30}
			caster.AttacksThisTurn = 1
			target := testCombatant("target", SideParty, creature.KindPC, 100, 20)
			ability := creature.Ability{ID: "bolt", Name: "Bolt", Actions: 2, Steps: []creature.Step{{
				SpellAttack: true,
				Damage:      []creature.DamagePart{damagePart(t, "1d6", "fire")},
				Outcomes:    creature.Outcomes{Success: []creature.Step{{Damage: []creature.DamagePart{damagePart(t, "1d4", "fire")}}}},
			}}}
			fight, events := testFight(script(t, testCase.dice...), caster, target)

			fight.useAbility(caster, ability, target)

			attacks := eventsOf[SpellAttackRolled](events.events)
			if len(attacks) != 1 || attacks[0].Penalty != -5 || attacks[0].Total != testCase.wantTotal || attacks[0].Degree != testCase.wantDegree {
				t.Errorf("attacks = %+v, want penalty -5, total %d, degree %v", attacks, testCase.wantTotal, testCase.wantDegree)
			}
			dealt := eventsOf[DamageDealt](events.events)
			if len(dealt) != 2 {
				t.Fatalf("damage = %+v, want the step's damage and the success list's", dealt)
			}
			if dealt[0].Total != testCase.wantStepDamage || dealt[0].Critical != testCase.wantStepCritical {
				t.Errorf("step damage = %+v, want total %d, critical %v", dealt[0], testCase.wantStepDamage, testCase.wantStepCritical)
			}
			if dealt[1].Total != 2 || dealt[1].Ability != "bolt" {
				t.Errorf("success list damage = %+v, want 2 from bolt", dealt[1])
			}
		})
	}
}

func TestAbilityRetargetsWhenTargetFalls(t *testing.T) {
	caster := testCombatant("caster", SideParty, creature.KindPC, 50, 20)
	weak := testCombatant("weak", SideEnemies, creature.KindNPC, 5, 20)
	strong := testCombatant("strong", SideEnemies, creature.KindNPC, 50, 20)
	flurry := creature.Ability{ID: "flurry", Name: "Flurry", Actions: 3, Steps: []creature.Step{{Strike: "fist"}, {Strike: "fist"}, {Strike: "fist"}}}
	fight, events := testFight(script(t, 15, 6, 15, 1, 15), caster, weak, strong)

	fight.useAbility(caster, flurry, weak)

	var targets []string
	for _, rolled := range eventsOf[StrikeRolled](events.events) {
		targets = append(targets, rolled.Target)
	}
	if want := []string{"weak", "strong", "strong"}; !slices.Equal(targets, want) {
		t.Errorf("strike targets = %v, want %v", targets, want)
	}
	if !weak.Dead {
		t.Error("weak is not dead")
	}
}

const sporeCaster = `schema = 1
name = "Spore Caster"
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

[ability.spores]
name = "Destiny Spores"
actions = 3
traits = ["mental", "occult"]
area = { shape = "cone", size = 40 }

[[ability.spores.steps]]
save = "will"
dc = 34
basic = true
damage = [{ dice = "4d6", type = "mental" }]
`

func TestStepsFromALoadedLayer(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "spore-caster")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "hand.toml"), []byte(sporeCaster), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
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
		MaxRounds: 1,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	savesBy := map[string]int{}
	for _, save := range eventsOf[SaveRolled](events.events) {
		savesBy[save.Combatant]++
	}
	if savesBy["dummy-fighter"] != 1 || savesBy["dummy-alchemist"] != 1 || len(savesBy) != 2 {
		t.Errorf("saves by combatant = %v, want one each for dummy-fighter and dummy-alchemist", savesBy)
	}
}

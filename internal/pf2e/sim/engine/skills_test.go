package engine

import (
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func skillFight(t *testing.T, values ...int) (*Fight, *recorder, *Combatant, *Combatant) {
	t.Helper()
	user := testCombatant("kas", SideParty, creature.KindPC, 100, 20)
	user.Creature.Skills = map[string]int{"intimidation": 10, "athletics": 12, "stealth": 8}
	target := testCombatant("foe", SideEnemies, creature.KindNPC, 100, 20)
	target.Creature.Defense.Fortitude = 8
	target.Creature.Defense.Will = 5
	target.Creature.Perception = 7
	fight, events := testFight(script(t, values...), user, target)
	fight.current = user
	user.ActionsLeft = 3
	return fight, events, user, target
}

func checkAbility(id, skill string, step creature.Step, traits ...string) creature.Ability {
	step.Check = skill
	return creature.Ability{ID: id, Name: id, Actions: 1, Traits: traits, Steps: []creature.Step{step}}
}

func allDegrees() creature.Outcomes {
	return creature.Outcomes{
		CriticalSuccess: []creature.Step{{Condition: "frightened", Value: 2}},
		Success:         []creature.Step{{Condition: "frightened", Value: 1}},
		Failure:         []creature.Step{{Condition: "sickened", Value: 1}},
		CriticalFailure: []creature.Step{{Condition: "clumsy", Value: 1}},
	}
}

func checkedEvents(t *testing.T, fight *Fight, user, target *Combatant, ability creature.Ability, events *recorder) []SkillChecked {
	t.Helper()
	giveAbility(user, ability)
	if err := fight.perform(user, Choice{Kind: ChooseAbility, ID: ability.ID, Target: target}); err != nil {
		t.Fatalf("perform: %v", err)
	}
	return eventsOf[SkillChecked](events.events)
}

func TestSkillCheckAgainstADefense(t *testing.T) {
	for _, test := range []struct {
		name      string
		die       int
		degree    Degree
		condition string
		value     int
	}{
		{"critical success", 15, CriticalSuccess, "frightened", 2},
		{"success", 5, Success, "frightened", 1},
		{"failure", 4, Failure, "sickened", 1},
		{"critical failure", 1, CriticalFailure, "clumsy", 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			fight, events, user, target := skillFight(t, test.die)
			got := checkedEvents(t, fight, user, target, checkAbility("demoralize", "intimidation", creature.Step{Against: "will", Outcomes: allDegrees()}, "fear"), events)
			want := []SkillChecked{{Combatant: "kas", Target: "foe", Ability: "demoralize", Skill: "intimidation", Die: test.die, Bonus: 10, Total: test.die + 10, DC: 15, Degree: test.degree}}
			if !slices.Equal(got, want) {
				t.Errorf("SkillChecked = %+v, want %+v", got, want)
			}
			for _, name := range []string{"frightened", "sickened", "clumsy"} {
				wantValue := 0
				if name == test.condition {
					wantValue = test.value
				}
				if target.Condition(name) != wantValue {
					t.Errorf("target %s = %d, want %d", name, target.Condition(name), wantValue)
				}
			}
		})
	}

	t.Run("a critical success with no list runs nothing", func(t *testing.T) {
		fight, events, user, target := skillFight(t, 15)
		step := creature.Step{Against: "will", Outcomes: creature.Outcomes{Success: []creature.Step{{Condition: "frightened", Value: 1}}}}
		got := checkedEvents(t, fight, user, target, checkAbility("demoralize", "intimidation", step), events)
		if len(got) != 1 || got[0].Degree != CriticalSuccess {
			t.Fatalf("SkillChecked = %+v, want one critical success", got)
		}
		if target.Condition("frightened") != 0 {
			t.Errorf("target frightened = %d, want 0", target.Condition("frightened"))
		}
	})

	for _, test := range []struct {
		against string
		dc      int
	}{
		{"fortitude", 18},
		{"perception", 17},
	} {
		t.Run("against "+test.against, func(t *testing.T) {
			fight, events, user, target := skillFight(t, 10)
			got := checkedEvents(t, fight, user, target, checkAbility("shove", "athletics", creature.Step{Against: test.against}), events)
			if len(got) != 1 || got[0].DC != test.dc || got[0].Bonus != 12 {
				t.Errorf("SkillChecked = %+v, want DC %d and bonus 12", got, test.dc)
			}
		})
	}
}

func TestSkillCheckFixedDC(t *testing.T) {
	fight, events, user, target := skillFight(t, 10)
	target.Creature.Defense.Will = 30
	got := checkedEvents(t, fight, user, target, checkAbility("bluster", "intimidation", creature.Step{DC: 20, Outcomes: allDegrees()}), events)
	want := []SkillChecked{{Combatant: "kas", Target: "foe", Ability: "bluster", Skill: "intimidation", Die: 10, Bonus: 10, Total: 20, DC: 20, Degree: Success}}
	if !slices.Equal(got, want) {
		t.Errorf("SkillChecked = %+v, want %+v", got, want)
	}
	if target.Condition("frightened") != 1 {
		t.Errorf("target frightened = %d, want 1", target.Condition("frightened"))
	}
}

func TestSkillCheckModifiers(t *testing.T) {
	for _, test := range []struct {
		name  string
		skill string
		setup func(user, target *Combatant)
		bonus int
		dc    int
	}{
		{"circumstance bonus on the skill", "intimidation", func(user, _ *Combatant) {
			user.Creature.Modifiers = map[string]creature.CheckModifier{"prowess": {Selector: "intimidation", Type: "circumstance", Value: 2}}
		}, 12, 15},
		{"skill-check bonus", "athletics", func(user, _ *Combatant) {
			user.Creature.Modifiers = map[string]creature.CheckModifier{"focus": {Selector: "skill-check", Type: "status", Value: 1}}
		}, 13, 15},
		{"frightened 2 on the user", "intimidation", func(user, _ *Combatant) { giveCondition(user, "frightened", 2) }, 8, 15},
		{"sickened 1 on the user", "athletics", func(user, _ *Combatant) { giveCondition(user, "sickened", 1) }, 11, 15},
		{"frightened 1 on the target", "intimidation", func(_, target *Combatant) { giveCondition(target, "frightened", 1) }, 10, 14},
		{"enfeebled lowers athletics", "athletics", func(user, _ *Combatant) { giveCondition(user, "enfeebled", 2) }, 10, 15},
		{"enfeebled leaves intimidation", "intimidation", func(user, _ *Combatant) { giveCondition(user, "enfeebled", 2) }, 10, 15},
		{"stupefied lowers intimidation", "intimidation", func(user, _ *Combatant) { giveCondition(user, "stupefied", 2) }, 8, 15},
		{"stupefied leaves athletics", "athletics", func(user, _ *Combatant) { giveCondition(user, "stupefied", 2) }, 12, 15},
		{"clumsy lowers stealth", "stealth", func(user, _ *Combatant) { giveCondition(user, "clumsy", 1) }, 7, 15},
		{"clumsy leaves athletics", "athletics", func(user, _ *Combatant) { giveCondition(user, "clumsy", 1) }, 12, 15},
		{"drained leaves every skill", "athletics", func(user, _ *Combatant) { giveCondition(user, "drained", 1) }, 12, 15},
	} {
		t.Run(test.name, func(t *testing.T) {
			fight, events, user, target := skillFight(t, 10)
			test.setup(user, target)
			got := checkedEvents(t, fight, user, target, checkAbility("try", test.skill, creature.Step{Against: "will"}), events)
			if len(got) != 1 || got[0].Bonus != test.bonus || got[0].DC != test.dc {
				t.Errorf("SkillChecked = %+v, want bonus %d and DC %d", got, test.bonus, test.dc)
			}
		})
	}
}

func TestSkillCheckTakesMisfortune(t *testing.T) {
	t.Run("a skill reroll applies", func(t *testing.T) {
		fight, events, user, target := skillFight(t, 15, 3)
		user.rerolls = append(user.rerolls, &rerollEffect{keep: creature.RerollWorse, rolls: []string{creature.RollSkill}, uses: 1, ability: "ill-omen"})
		got := checkedEvents(t, fight, user, target, checkAbility("demoralize", "intimidation", creature.Step{Against: "will"}), events)
		if len(got) != 1 || got[0].Die != 3 || got[0].Degree != Failure {
			t.Errorf("SkillChecked = %+v, want die 3, a failure", got)
		}
		want := []Rerolled{{Combatant: "kas", Roll: "skill", Keep: "worse", First: 15, Second: 3, Kept: 3, Ability: "ill-omen"}}
		if rerolled := eventsOf[Rerolled](events.events); !slices.Equal(rerolled, want) {
			t.Errorf("Rerolled = %+v, want %+v", rerolled, want)
		}
	})

	t.Run("an attack reroll does not", func(t *testing.T) {
		fight, events, user, target := skillFight(t, 15)
		user.rerolls = append(user.rerolls, &rerollEffect{keep: creature.RerollWorse, rolls: []string{creature.RollAttack}, uses: 1})
		got := checkedEvents(t, fight, user, target, checkAbility("demoralize", "intimidation", creature.Step{Against: "will"}), events)
		if len(got) != 1 || got[0].Die != 15 {
			t.Errorf("SkillChecked = %+v, want die 15", got)
		}
		if len(user.rerolls) != 1 || len(eventsOf[Rerolled](events.events)) != 0 {
			t.Errorf("rerolls = %d, Rerolled = %+v, want the attack reroll kept and unused", len(user.rerolls), eventsOf[Rerolled](events.events))
		}
	})
}

func TestSkillCheckSkipsImmune(t *testing.T) {
	fight, events, user, target := skillFight(t)
	target.Creature.Defense.Immune = []string{"fear-effects"}
	got := checkedEvents(t, fight, user, target, checkAbility("demoralize", "intimidation", creature.Step{Against: "will", Outcomes: allDegrees()}, "emotion", "fear", "mental"), events)
	if len(got) != 0 {
		t.Errorf("SkillChecked = %+v, want no roll", got)
	}
	if immune := eventsOf[ImmuneToEffect](events.events); !slices.Equal(immune, []ImmuneToEffect{{Combatant: "foe", Source: "kas", Ability: "demoralize"}}) {
		t.Errorf("ImmuneToEffect = %+v, want foe against demoralize", immune)
	}
	if target.Condition("frightened") != 0 {
		t.Errorf("target frightened = %d, want 0", target.Condition("frightened"))
	}
}

func TestTargetModifierAgainstTraits(t *testing.T) {
	for _, test := range []struct {
		name   string
		traits []string
		dc     int
	}{
		{"a fear ability", []string{"emotion", "fear", "mental"}, 18},
		{"an ability without fear", []string{"mental"}, 15},
	} {
		t.Run(test.name, func(t *testing.T) {
			fight, events, user, target := skillFight(t, 10)
			target.Creature.Modifiers = map[string]creature.CheckModifier{"bravery": {Selector: "will", Type: "status", Value: 3, Against: []string{"fear"}}}
			got := checkedEvents(t, fight, user, target, checkAbility("demoralize", "intimidation", creature.Step{Against: "will"}, test.traits...), events)
			if len(got) != 1 || got[0].DC != test.dc {
				t.Errorf("SkillChecked = %+v, want DC %d", got, test.dc)
			}
		})
	}

	t.Run("a granted AC bonus with uses is neither read nor spent", func(t *testing.T) {
		fight, events, user, target := skillFight(t, 10)
		target.bonuses = append(target.bonuses, &grantedBonus{modifier: creature.CheckModifier{Selector: "ac", Type: "circumstance", Value: 2}, uses: 1})
		got := checkedEvents(t, fight, user, target, checkAbility("demoralize", "intimidation", creature.Step{Against: "will"}, "fear"), events)
		if len(got) != 1 || got[0].DC != 15 {
			t.Errorf("SkillChecked = %+v, want DC 15", got)
		}
		if len(target.bonuses) != 1 || target.bonuses[0].uses != 1 {
			t.Errorf("target bonuses = %d, want the AC bonus kept with its use", len(target.bonuses))
		}
	})
}

package engine

import (
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func rerollStep(keep string, rolls []string, uses int, duration *creature.Duration) creature.Step {
	return creature.Step{Reroll: keep, Rolls: rolls, Uses: uses, Duration: duration}
}

func rollsFight(t *testing.T, values ...int) (*Fight, *recorder, *Combatant, *Combatant, *scriptedRoller) {
	attacker := testCombatant("attacker", SideParty, creature.KindPC, 50, 20)
	target := testCombatant("target", SideEnemies, creature.KindNPC, 500, 25)
	roller := script(t, values...)
	fight, events := testFight(roller, attacker, target)
	return fight, events, attacker, target, roller
}

func strikeRolls(events *recorder) []StrikeRolled {
	return eventsOf[StrikeRolled](events.events)
}

func TestMisfortuneKeepsTheWorse(t *testing.T) {
	cases := []struct {
		name       string
		values     []int
		wantKept   int
		wantDegree Degree
	}{
		{"the lower die", []int{18, 12}, 12, Failure},
		{"a natural 1 second", []int{20, 1}, 1, CriticalFailure},
		{"a natural 1 first", []int{1, 20}, 1, CriticalFailure},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			fight, events, attacker, target, roller := rollsFight(t, testCase.values...)
			fight.addReroll(effectSource{actor: target, ability: "omen"}, attacker, rerollStep(creature.RerollWorse, []string{creature.RollAttack}, 1, nil))

			degree, _ := fight.strike(attacker, target, fist(), strikeOptions{})

			want := []Rerolled{{Combatant: "attacker", Roll: "attack", Keep: "worse", First: testCase.values[0], Second: testCase.values[1], Kept: testCase.wantKept, Ability: "omen"}}
			if got := eventsOf[Rerolled](events.events); !slices.Equal(got, want) {
				t.Errorf("Rerolled = %+v, want %+v", got, want)
			}
			if rolled := strikeRolls(events); len(rolled) != 1 || rolled[0].Die != testCase.wantKept || rolled[0].Degree != testCase.wantDegree || degree != testCase.wantDegree {
				t.Errorf("StrikeRolled = %+v, degree %v, want die %d and %v", rolled, degree, testCase.wantKept, testCase.wantDegree)
			}
			if len(roller.values) != 0 {
				t.Errorf("unrolled = %v, want none", roller.values)
			}
		})
	}
}

func TestFortuneKeepsTheBetter(t *testing.T) {
	cases := []struct {
		name       string
		values     []int
		wantKept   int
		wantDegree Degree
		wantDamage []int
	}{
		{"the higher die", []int{3, 16, 4}, 16, Success, []int{6}},
		{"a natural 20 second", []int{1, 20, 4}, 20, CriticalSuccess, []int{12}},
		{"two natural 1s", []int{1, 1}, 1, CriticalFailure, nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			fight, events, attacker, target, roller := rollsFight(t, testCase.values...)
			fight.addReroll(effectSource{actor: attacker, ability: "luck"}, attacker, rerollStep(creature.RerollBetter, []string{creature.RollAttack}, 1, nil))

			fight.strike(attacker, target, fist(), strikeOptions{})

			want := []Rerolled{{Combatant: "attacker", Roll: "attack", Keep: "better", First: testCase.values[0], Second: testCase.values[1], Kept: testCase.wantKept, Ability: "luck"}}
			if got := eventsOf[Rerolled](events.events); !slices.Equal(got, want) {
				t.Errorf("Rerolled = %+v, want %+v", got, want)
			}
			if rolled := strikeRolls(events); len(rolled) != 1 || rolled[0].Die != testCase.wantKept || rolled[0].Degree != testCase.wantDegree {
				t.Errorf("StrikeRolled = %+v, want die %d and %v", rolled, testCase.wantKept, testCase.wantDegree)
			}
			if got := damageTo(events, "target"); !slices.Equal(got, testCase.wantDamage) {
				t.Errorf("damage = %v, want %v", got, testCase.wantDamage)
			}
			if len(roller.values) != 0 {
				t.Errorf("unrolled = %v, want none", roller.values)
			}
		})
	}
}

func TestFortuneAndMisfortuneCancel(t *testing.T) {
	t.Run("one die, each loses a use", func(t *testing.T) {
		fight, events, attacker, target, roller := rollsFight(t, 15, 3)
		attack := []string{creature.RollAttack}
		fight.addReroll(effectSource{actor: target, ability: "omen"}, attacker, rerollStep(creature.RerollWorse, attack, 2, nil))
		fight.addReroll(effectSource{actor: attacker, ability: "luck"}, attacker, rerollStep(creature.RerollBetter, attack, 1, nil))

		fight.strike(attacker, target, fist(), strikeOptions{})

		if got := eventsOf[RerollsCancelled](events.events); !slices.Equal(got, []RerollsCancelled{{Combatant: "attacker", Roll: "attack"}}) {
			t.Errorf("RerollsCancelled = %+v, want one for attacker's attack", got)
		}
		if got := eventsOf[Rerolled](events.events); len(got) != 0 {
			t.Errorf("Rerolled = %+v, want none", got)
		}
		if rolled := strikeRolls(events); len(rolled) != 1 || rolled[0].Die != 15 || rolled[0].Degree != Success {
			t.Errorf("StrikeRolled = %+v, want one die of 15, a success", rolled)
		}
		if len(attacker.rerolls) != 1 || attacker.rerolls[0].ability != "omen" || attacker.rerolls[0].uses != 1 {
			t.Errorf("rerolls left = %+v, want omen with 1 use; luck spent", attacker.rerolls)
		}
		if len(roller.values) != 0 {
			t.Errorf("unrolled = %v, want none", roller.values)
		}
	})
	t.Run("two misfortunes, one applies", func(t *testing.T) {
		fight, events, attacker, target, _ := rollsFight(t, 15, 12)
		attack := []string{creature.RollAttack}
		fight.addReroll(effectSource{actor: target, ability: "omen"}, attacker, rerollStep(creature.RerollWorse, attack, 1, nil))
		fight.addReroll(effectSource{actor: target, ability: "hex"}, attacker, rerollStep(creature.RerollWorse, attack, 1, nil))

		fight.strike(attacker, target, fist(), strikeOptions{})

		want := []Rerolled{{Combatant: "attacker", Roll: "attack", Keep: "worse", First: 15, Second: 12, Kept: 12, Ability: "omen"}}
		if got := eventsOf[Rerolled](events.events); !slices.Equal(got, want) {
			t.Errorf("Rerolled = %+v, want %+v", got, want)
		}
		if len(attacker.rerolls) != 1 || attacker.rerolls[0].ability != "hex" {
			t.Errorf("rerolls left = %+v, want hex unspent", attacker.rerolls)
		}
	})
}

func TestRerollUses(t *testing.T) {
	attack := []string{creature.RollAttack}
	t.Run("uses 1 is gone after one roll", func(t *testing.T) {
		fight, events, attacker, target, roller := rollsFight(t, 5, 6, 5)
		fight.addReroll(effectSource{actor: target, ability: "omen"}, attacker, rerollStep(creature.RerollWorse, attack, 1, nil))

		fight.strike(attacker, target, fist(), strikeOptions{})
		fight.strike(attacker, target, fist(), strikeOptions{})

		if got := eventsOf[Rerolled](events.events); len(got) != 1 {
			t.Errorf("Rerolled = %+v, want one", got)
		}
		if len(attacker.rerolls) != 0 || len(roller.values) != 0 {
			t.Errorf("rerolls, unrolled = %+v, %v, want none, none", attacker.rerolls, roller.values)
		}
	})
	t.Run("a duration covers every roll until the source's next turn", func(t *testing.T) {
		fight, events, attacker, target, roller := rollsFight(t, 5, 6, 7, 8, 5, 9)
		fight.addReroll(effectSource{actor: target, ability: "omen"}, attacker, rerollStep(creature.RerollWorse, attack, 0, &creature.Duration{Rounds: 1}))

		fight.strike(attacker, target, fist(), strikeOptions{})
		fight.startTurn(attacker)
		fight.strike(attacker, target, fist(), strikeOptions{})
		fight.startTurn(target)
		fight.strike(attacker, target, fist(), strikeOptions{})

		want := []Rerolled{
			{Combatant: "attacker", Roll: "attack", Keep: "worse", First: 5, Second: 6, Kept: 5, Ability: "omen"},
			{Combatant: "attacker", Roll: "attack", Keep: "worse", First: 7, Second: 8, Kept: 7, Ability: "omen"},
		}
		if got := eventsOf[Rerolled](events.events); !slices.Equal(got, want) {
			t.Errorf("Rerolled = %+v, want %+v", got, want)
		}
		if rolled := strikeRolls(events); len(rolled) != 3 || rolled[2].Die != 5 {
			t.Errorf("StrikeRolled = %+v, want a third Strike on one die of 5", rolled)
		}
		if len(attacker.rerolls) != 0 || len(roller.values) != 1 {
			t.Errorf("rerolls, unrolled = %+v, %v, want none, [9]", attacker.rerolls, roller.values)
		}
	})
	t.Run("uses 1 with a duration ends unused", func(t *testing.T) {
		fight, events, attacker, target, roller := rollsFight(t, 5)
		fight.addReroll(effectSource{actor: target, ability: "omen"}, attacker, rerollStep(creature.RerollWorse, attack, 1, &creature.Duration{Rounds: 1}))

		fight.startTurn(attacker)
		if len(attacker.rerolls) != 1 {
			t.Fatalf("after the bearer's turn start: rerolls = %+v, want one", attacker.rerolls)
		}
		fight.startTurn(target)
		fight.strike(attacker, target, fist(), strikeOptions{})

		if got := eventsOf[Rerolled](events.events); len(got) != 0 || len(attacker.rerolls) != 0 || len(roller.values) != 0 {
			t.Errorf("Rerolled, rerolls, unrolled = %+v, %+v, %v, want none", got, attacker.rerolls, roller.values)
		}
	})
}

func TestRerollOnlyItsRolls(t *testing.T) {
	t.Run("an attack reroll leaves a save, a flat check and a recovery check alone", func(t *testing.T) {
		fight, events, attacker, target, roller := rollsFight(t, 12, 4, 9)
		fight.addReroll(effectSource{actor: target, ability: "omen"}, attacker, rerollStep(creature.RerollWorse, []string{creature.RollAttack}, 1, nil))
		giveCondition(target, "concealed", 0)

		fight.savingThrow(attacker, "will", 20, 0, effectSource{actor: target, ability: "test-save"})
		fight.passesConcealment(attacker, target)
		attacker.HP, attacker.Dying = 0, 1
		fight.recoveryCheck(attacker)

		if got := eventsOf[Rerolled](events.events); len(got) != 0 {
			t.Errorf("Rerolled = %+v, want none", got)
		}
		if saves := eventsOf[SaveRolled](events.events); len(saves) != 1 || saves[0].Die != 12 {
			t.Errorf("SaveRolled = %+v, want one die of 12", saves)
		}
		if flat := eventsOf[FlatChecked](events.events); len(flat) != 1 || flat[0].Die != 4 {
			t.Errorf("FlatChecked = %+v, want one die of 4", flat)
		}
		if recovery := eventsOf[RecoveryChecked](events.events); len(recovery) != 1 || recovery[0].Die != 9 {
			t.Errorf("RecoveryChecked = %+v, want one die of 9", recovery)
		}
		if len(attacker.rerolls) != 1 || attacker.rerolls[0].uses != 1 || len(roller.values) != 0 {
			t.Errorf("rerolls, unrolled = %+v, %v, want omen with its use, none", attacker.rerolls, roller.values)
		}
	})
	t.Run("a save reroll leaves an attack alone and sets the save's degree", func(t *testing.T) {
		fight, events, attacker, target, _ := rollsFight(t, 8, 15, 4)
		fight.addReroll(effectSource{actor: target, ability: "luck"}, target, rerollStep(creature.RerollBetter, []string{creature.RollSave}, 1, nil))

		fight.strike(target, attacker, fist(), strikeOptions{})
		degree := fight.savingThrow(target, "will", 20, 0, effectSource{actor: attacker, ability: "test-save"})

		if rolled := strikeRolls(events); len(rolled) != 1 || rolled[0].Die != 8 {
			t.Errorf("StrikeRolled = %+v, want one die of 8", rolled)
		}
		want := []Rerolled{{Combatant: "target", Roll: "save", Keep: "better", First: 15, Second: 4, Kept: 15, Ability: "luck"}}
		if got := eventsOf[Rerolled](events.events); !slices.Equal(got, want) {
			t.Errorf("Rerolled = %+v, want %+v", got, want)
		}
		if degree != Failure {
			t.Errorf("degree = %v, want Failure (15 against DC 20)", degree)
		}
	})
}

func acBonusStep(kind string, value, uses int, duration *creature.Duration) creature.Step {
	return creature.Step{Bonus: &creature.CheckModifier{Selector: "ac", Type: kind, Value: value}, Uses: uses, Duration: duration}
}

func TestGrantedBonus(t *testing.T) {
	t.Run("an AC bonus with uses 1 lasts one attack", func(t *testing.T) {
		fight, events, attacker, target, _ := rollsFight(t, 15, 15, 1)
		fight.addBonus(effectSource{actor: target, ability: "flip"}, target, acBonusStep("circumstance", 2, 1, nil))

		fight.strike(attacker, target, fist(), strikeOptions{})
		fight.strike(attacker, target, fist(), strikeOptions{})

		want := []BonusGained{{Combatant: "target", Source: "target", Ability: "flip", Selector: "ac", Value: 2}}
		if got := eventsOf[BonusGained](events.events); !slices.Equal(got, want) {
			t.Errorf("BonusGained = %+v, want %+v", got, want)
		}
		rolled := strikeRolls(events)
		if len(rolled) != 2 || rolled[0].AC != 27 || rolled[0].Degree != Failure || rolled[1].AC != 25 || rolled[1].Degree != Success {
			t.Errorf("StrikeRolled = %+v, want a miss against AC 27, then a hit against 25", rolled)
		}
		if len(target.bonuses) != 0 {
			t.Errorf("bonuses = %+v, want none", target.bonuses)
		}
	})
	t.Run("a status bonus stacks by type and ends at its boundary", func(t *testing.T) {
		fight, _, attacker, target, _ := rollsFight(t)
		target.Creature.Modifiers = map[string]creature.CheckModifier{"own": {Selector: "ac", Type: "status", Value: 1}}
		fight.addBonus(effectSource{actor: attacker, ability: "shield"}, target, acBonusStep("status", 3, 0, &creature.Duration{Rounds: 1}))

		if got := fight.armorClass(target, nil); got != 28 {
			t.Errorf("AC = %d, want 28: the larger status bonus only", got)
		}
		fight.addBonus(effectSource{actor: attacker, ability: "cover"}, target, acBonusStep("circumstance", 2, 0, &creature.Duration{Until: "source-turn-end"}))
		if got := fight.armorClass(target, nil); got != 30 {
			t.Errorf("AC with circumstance = %d, want 30", got)
		}
		fight.startTurn(target)
		if got := fight.armorClass(target, nil); got != 30 {
			t.Errorf("AC after the bearer's turn start = %d, want 30", got)
		}
		fight.startTurn(attacker)
		if got := fight.armorClass(target, nil); got != 28 {
			t.Errorf("AC after the source's turn start = %d, want 28: the granted status bonus ended", got)
		}
		fight.endTurn(attacker)
		if got := fight.armorClass(target, nil); got != 26 || len(target.bonuses) != 0 {
			t.Errorf("AC, bonuses after the source's turn end = %d, %+v, want 26, none", got, target.bonuses)
		}
	})
}

func TestBonusIsSpentByAMiss(t *testing.T) {
	t.Run("a Strike", func(t *testing.T) {
		fight, events, attacker, target, _ := rollsFight(t, 1)
		fight.addBonus(effectSource{actor: target, ability: "flip"}, target, acBonusStep("circumstance", 2, 1, nil))

		fight.strike(attacker, target, fist(), strikeOptions{})

		if rolled := strikeRolls(events); len(rolled) != 1 || rolled[0].Degree != CriticalFailure || rolled[0].AC != 27 {
			t.Errorf("StrikeRolled = %+v, want a critical miss against AC 27", rolled)
		}
		if len(target.bonuses) != 0 {
			t.Errorf("bonuses = %+v, want none", target.bonuses)
		}
	})
	t.Run("a spell attack", func(t *testing.T) {
		fight, events, attacker, target, _ := rollsFight(t, 2)
		attacker.Creature.Spellcasting = &creature.Spellcasting{Attack: 10}
		fight.addBonus(effectSource{actor: target, ability: "flip"}, target, acBonusStep("circumstance", 2, 1, nil))

		fight.spellAttack(effectSource{actor: attacker, ability: "bolt"}, creature.Step{SpellAttack: true}, target)

		if rolled := eventsOf[SpellAttackRolled](events.events); len(rolled) != 1 || rolled[0].AC != 27 || rolled[0].Degree != CriticalFailure {
			t.Errorf("SpellAttackRolled = %+v, want a critical miss against AC 27", rolled)
		}
		if len(target.bonuses) != 0 {
			t.Errorf("bonuses = %+v, want none", target.bonuses)
		}
	})
}

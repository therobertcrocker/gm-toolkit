package engine

import (
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func TestDownedOnOwnTurnKeepsInitiative(t *testing.T) {
	attacker := testCombatant("attacker", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 7, 20)
	other := testCombatant("other", SideParty, creature.KindPC, 50, 20)
	strike := fist()
	strike.Damage = append(strike.Damage, categorizedPart(t, "3", "fire", "persistent"))
	fight, _ := testFight(script(t, 15, 4, 14), attacker, target, other)
	before := orderIDs(fight)

	fight.strike(attacker, target, strike, strikeOptions{})
	fight.current = target
	fight.endTurn(target)

	if target.Dying != 1 {
		t.Fatalf("Dying = %d, want 1", target.Dying)
	}
	if got := orderIDs(fight); !slices.Equal(got, before) {
		t.Errorf("order = %v, want %v", got, before)
	}
}

func TestNoFlatCheckAfterPersistentKills(t *testing.T) {
	attacker := testCombatant("attacker", SideEnemies, creature.KindNPC, 50, 20)
	target := downedPC(30, 3)
	target.persistent = []*persistentInstance{{source: effectSource{actor: attacker}, part: categorizedPart(t, "3", "fire", "persistent")}}
	fight, events := testFight(script(t), attacker, target)

	fight.endTurn(target)

	if !target.Dead {
		t.Fatal("target alive, want dead")
	}
	if checks := eventsOf[FlatChecked](events.events); len(checks) != 0 {
		t.Errorf("flat checks = %+v, want none", checks)
	}
}

func TestDamageImmunityIsNotEffectImmunity(t *testing.T) {
	cases := []struct {
		name       string
		immune     string
		trait      string
		wantImmune bool
	}{
		{"damage type", "fire", "fire", false},
		{"trait-style immunity", "poison", "poison", true},
		{"effects key", "death-effects", "death", true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			caster := testCombatant("caster", SideParty, creature.KindPC, 50, 20)
			target := testCombatant("target", SideEnemies, creature.KindNPC, 50, 20)
			target.Creature.Defense.Immune = []string{testCase.immune}
			rolls := []int{10}
			if testCase.wantImmune {
				rolls = nil
			}
			fight, events := testFight(script(t, rolls...), caster, target)

			fight.useAbility(abilityUse{actor: caster, ability: saveAbility("reflex", 20, testCase.trait), target: target})

			immune := len(eventsOf[ImmuneToEffect](events.events)) == 1
			saved := len(eventsOf[SaveRolled](events.events)) == 1
			if immune != testCase.wantImmune || saved == testCase.wantImmune {
				t.Errorf("immune, saved = %v, %v, want %v, %v", immune, saved, testCase.wantImmune, !testCase.wantImmune)
			}
		})
	}
}

func TestDoomedWhileDyingKills(t *testing.T) {
	target := downedPC(30, 2)
	fight, _ := testFight(script(t), target)

	fight.applyCondition(target, target, conditionStep("doomed", 2, nil))

	if !target.Dead {
		t.Errorf("dying 2 with doomed 2: alive, want dead")
	}
}

func TestStabilizingAtDeathValueZeroDoesNotKill(t *testing.T) {
	target := downedPC(30, 1)
	target.HeroPoints = 1
	giveCondition(target, "doomed", 4)
	fight, events := testFight(script(t), target)

	fight.setDying(target, 0)

	if target.Dead || target.HeroPoints != 1 || len(eventsOf[HeroicRecovery](events.events)) != 0 {
		t.Errorf("Dead, HeroPoints = %v, %d, want false, 1 and no heroic recovery", target.Dead, target.HeroPoints)
	}
}

func TestBasicSuccessHalvesTheTotal(t *testing.T) {
	roll := damageRoll{parts: []rolledPart{{damageType: "fire", amount: 5}, {damageType: "cold", amount: 5}}}

	total := 0
	for _, part := range roll.basic(Success).parts {
		total += part.amount
	}

	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
}

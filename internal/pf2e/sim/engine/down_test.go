package engine

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func downedPC(hp, dying int) *Combatant {
	target := testCombatant("target", SideParty, creature.KindPC, hp, 20)
	target.HP, target.Dying = 0, dying
	return target
}

func TestDoomedLowersTheDeathValue(t *testing.T) {
	target := downedPC(30, 2)
	giveCondition(target, "doomed", 1)
	fight, _ := testFight(script(t), target)

	fight.hitWhileDown(target, false)

	if target.Dying != 3 || !target.Dead {
		t.Errorf("Dying, Dead = %d, %v, want 3, true", target.Dying, target.Dead)
	}
}

func TestDyingMax(t *testing.T) {
	target := downedPC(30, 4)
	target.Creature.DyingMax = 6
	fight, _ := testFight(script(t), target)

	fight.hitWhileDown(target, false)
	if target.Dying != 5 || target.Dead {
		t.Fatalf("after one hit: Dying, Dead = %d, %v, want 5, false", target.Dying, target.Dead)
	}
	fight.hitWhileDown(target, false)
	if target.Dying != 6 || !target.Dead {
		t.Errorf("after two hits: Dying, Dead = %d, %v, want 6, true", target.Dying, target.Dead)
	}
}

func TestRecoveryDC(t *testing.T) {
	target := downedPC(30, 1)
	target.Creature.RecoveryDC = 6
	fight, events := testFight(script(t, 7), target)

	fight.recoveryCheck(target)

	checks := eventsOf[RecoveryChecked](events.events)
	if len(checks) != 1 || checks[0].DC != 7 || checks[0].Degree != Success {
		t.Errorf("recovery checks = %+v, want one success at DC 7", checks)
	}
	if target.Dying != 0 {
		t.Errorf("Dying = %d, want 0", target.Dying)
	}
}

func TestHeroicRecovery(t *testing.T) {
	target := downedPC(30, 3)
	target.Wounded = 1
	target.HeroPoints = 1
	fight, events := testFight(script(t), target)

	fight.hitWhileDown(target, false)

	if recoveries := eventsOf[HeroicRecovery](events.events); len(recoveries) != 1 {
		t.Errorf("heroic recoveries = %+v, want one", recoveries)
	}
	if target.Dead || target.Dying != 0 || target.HP != 0 || target.Wounded != 1 || target.HeroPoints != 0 {
		t.Fatalf("Dead, Dying, HP, Wounded, HeroPoints = %v, %d, %d, %d, %d, want false, 0, 0, 1, 0",
			target.Dead, target.Dying, target.HP, target.Wounded, target.HeroPoints)
	}

	fight.hitWhileDown(target, true)
	if target.Dying != 3 || target.Dead {
		t.Fatalf("knocked down again: Dying, Dead = %d, %v, want 3, false", target.Dying, target.Dead)
	}
	fight.hitWhileDown(target, false)
	if !target.Dead {
		t.Errorf("Dead = false at dying 4 with no hero points, want true")
	}
}

func TestDrainedLosesHP(t *testing.T) {
	source := testCombatant("source", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 40, 20)
	target.Creature.Level = 5
	fight, events := testFight(script(t), source, target)

	fight.applyCondition(source, target, conditionStep("drained", 2, nil))

	lost := eventsOf[HPLost](events.events)
	if len(lost) != 1 || lost[0].Amount != 10 || lost[0].Reason != "drained" || lost[0].HPAfter != 30 {
		t.Errorf("HP lost = %+v, want one drained loss of 10 leaving 30", lost)
	}
	if target.HP != 30 || target.MaxHP() != 30 {
		t.Errorf("HP, MaxHP() = %d, %d, want 30, 30", target.HP, target.MaxHP())
	}

	fight.applyCondition(source, target, conditionStep("drained", 3, nil))

	if target.HP != 25 || target.MaxHP() != 25 {
		t.Errorf("after drained 3: HP, MaxHP() = %d, %d, want 25, 25", target.HP, target.MaxHP())
	}
}

func TestLethalDamageKills(t *testing.T) {
	caster := testCombatant("caster", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 5, 20)
	ability := creature.Ability{ID: "disintegrate", Name: "Disintegrate", Actions: 2, Steps: []creature.Step{
		{Damage: []creature.DamagePart{damagePart(t, "1d6+10", "untyped")}, Lethal: true},
	}}
	fight, events := testFight(script(t, 1), caster, target)

	fight.useAbility(abilityUse{actor: caster, ability: ability, target: target})

	if died := eventsOf[Died](events.events); len(died) != 1 || !target.Dead {
		t.Errorf("died = %+v, Dead = %v, want one Died and dead", died, target.Dead)
	}
	if dying := eventsOf[DyingChanged](events.events); len(dying) != 0 {
		t.Errorf("dying changes = %+v, want none", dying)
	}
}

func TestUntypedDamageTakesResistAll(t *testing.T) {
	target := NewCombatant(SideEnemies, &creature.Creature{Defense: creature.Defense{Resist: map[string]creature.Adjustment{"all": {Value: 5}}}})
	if got := target.adjustedDamage("untyped", nil, 20); got != 15 {
		t.Errorf("adjustedDamage = %d, want 15", got)
	}
}

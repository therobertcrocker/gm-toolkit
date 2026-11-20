package engine

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func categorizedPart(t *testing.T, text, damageType, category string) creature.DamagePart {
	t.Helper()
	part := damagePart(t, text, damageType)
	part.Category = category
	return part
}

func bombStrike(t *testing.T) creature.Strike {
	return creature.Strike{
		ID:     "bomb",
		Name:   "Bomb",
		Bonus:  10,
		Range:  20,
		Traits: []string{"bomb", "splash"},
		Damage: []creature.DamagePart{damagePart(t, "3d8", "fire"), categorizedPart(t, "3", "fire", "splash")},
	}
}

func lastDamage(t *testing.T, events *recorder) DamageDealt {
	t.Helper()
	dealt := eventsOf[DamageDealt](events.events)
	if len(dealt) == 0 {
		t.Fatal("no DamageDealt event")
	}
	return dealt[len(dealt)-1]
}

func TestPersistentDamageOnAHit(t *testing.T) {
	attacker := testCombatant("attacker", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 50, 20)
	strike := fist()
	strike.Damage = append(strike.Damage, categorizedPart(t, "3", "fire", "persistent"))
	fight, events := testFight(script(t, 15, 4, 14, 15), attacker, target)

	fight.strike(attacker, target, strike, strikeOptions{})
	if changed := eventsOf[PersistentChanged](events.events); len(changed) != 1 || changed[0].Type != "fire" || changed[0].Dice != "3" {
		t.Fatalf("PersistentChanged = %+v, want one 3 fire", changed)
	}

	fight.endTurn(target)
	if dealt := lastDamage(t, events); dealt.Ability != "persistent" || dealt.Total != 3 {
		t.Errorf("first end of turn: damage = %+v, want 3 persistent", dealt)
	}
	if len(target.persistent) != 1 {
		t.Fatalf("after flat check 14: %d persistent instances, want 1", len(target.persistent))
	}

	fight.endTurn(target)
	if ended := eventsOf[PersistentEnded](events.events); len(ended) != 1 || ended[0].Type != "fire" {
		t.Errorf("PersistentEnded = %+v, want one fire", ended)
	}
	if len(target.persistent) != 0 {
		t.Errorf("after flat check 15: %d persistent instances, want 0", len(target.persistent))
	}
	if target.HP != 38 {
		t.Errorf("HP = %d, want 38 (50, less 1d6+2 rolled 6, 3 and 3)", target.HP)
	}
}

func TestPersistentKeepsTheLarger(t *testing.T) {
	attacker := testCombatant("attacker", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 50, 20)
	fight, events := testFight(script(t), attacker, target)
	source := effectSource{actor: attacker}

	fight.addPersistent(source, target, categorizedPart(t, "1d6", "fire", "persistent"), persistentWhole)
	fight.addPersistent(source, target, categorizedPart(t, "2d6", "fire", "persistent"), persistentWhole)
	fight.addPersistent(source, target, categorizedPart(t, "1", "fire", "persistent"), persistentWhole)

	if len(target.persistent) != 1 || target.persistent[0].part.Roll.String() != "2d6" {
		t.Errorf("persistent = %+v, want one 2d6 fire", target.persistent)
	}
	if changed := eventsOf[PersistentChanged](events.events); len(changed) != 2 {
		t.Errorf("PersistentChanged = %+v, want two (1d6, then 2d6)", changed)
	}
}

func TestPersistentDoublesOnACriticalHit(t *testing.T) {
	attacker := testCombatant("attacker", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 50, 20)
	strike := fist()
	strike.Damage = append(strike.Damage, categorizedPart(t, "1d4", "fire", "persistent"))
	fight, events := testFight(script(t, 20, 4, 3, 1), attacker, target)

	fight.strike(attacker, target, strike, strikeOptions{})
	if len(target.persistent) != 1 || target.persistent[0].scale != persistentDoubled {
		t.Fatalf("persistent = %+v, want one doubled instance", target.persistent)
	}

	fight.endTurn(target)
	if dealt := lastDamage(t, events); dealt.Total != 6 {
		t.Errorf("persistent damage = %d, want 6 (rolled 3, doubled)", dealt.Total)
	}
}

func TestPersistentImmunity(t *testing.T) {
	attacker := testCombatant("attacker", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 50, 20)
	target.Creature.Defense.Immune = []string{"bleed"}
	fight, _ := testFight(script(t), attacker, target)

	fight.addPersistent(effectSource{actor: attacker}, target, categorizedPart(t, "1d6", "bleed", "persistent"), persistentWhole)

	if len(target.persistent) != 0 {
		t.Errorf("persistent = %+v, want none", target.persistent)
	}
}

func TestSplash(t *testing.T) {
	cases := []struct {
		name      string
		rolls     []int
		wantTotal int
		wantDealt bool
	}{
		{name: "hit", rolls: []int{15, 1, 2, 3}, wantTotal: 9, wantDealt: true},
		{name: "critical hit", rolls: []int{20, 1, 2, 3}, wantTotal: 15, wantDealt: true},
		{name: "failure", rolls: []int{5}, wantTotal: 3, wantDealt: true},
		{name: "critical failure", rolls: []int{1}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			attacker := testCombatant("attacker", SideParty, creature.KindPC, 50, 20)
			target := testCombatant("target", SideEnemies, creature.KindNPC, 100, 20)
			fight, events := testFight(script(t, testCase.rolls...), attacker, target)

			fight.strike(attacker, target, bombStrike(t), strikeOptions{})

			dealt := eventsOf[DamageDealt](events.events)
			if !testCase.wantDealt {
				if len(dealt) != 0 {
					t.Errorf("DamageDealt = %+v, want none", dealt)
				}
				return
			}
			if len(dealt) != 1 || dealt[0].Total != testCase.wantTotal || len(dealt[0].Parts) != 1 || dealt[0].Parts[0].Type != "fire" {
				t.Errorf("DamageDealt = %+v, want one fire amount of %d", dealt, testCase.wantTotal)
			}
		})
	}
}

func TestDamageMinimumOfOne(t *testing.T) {
	attacker := testCombatant("attacker", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 50, 20)
	strike := fist()
	strike.Damage = []creature.DamagePart{damagePart(t, "1d4-5", "bludgeoning")}
	fight, events := testFight(script(t, 15, 1), attacker, target)

	fight.strike(attacker, target, strike, strikeOptions{})

	if dealt := lastDamage(t, events); dealt.Total != 1 {
		t.Errorf("Total = %d, want 1", dealt.Total)
	}
}

func TestWeaknessSkipsAZeroPart(t *testing.T) {
	cases := []struct {
		name      string
		damage    func(t *testing.T) []creature.DamagePart
		rolls     []int
		wantTaken map[string]int
	}{
		{
			name: "a lone part raised to 1 takes the weakness",
			damage: func(t *testing.T) []creature.DamagePart {
				return []creature.DamagePart{damagePart(t, "1d4-5", "slashing")}
			},
			rolls:     []int{15, 1},
			wantTaken: map[string]int{"slashing": 6},
		},
		{
			name: "a part that rolled 0 beside another takes none",
			damage: func(t *testing.T) []creature.DamagePart {
				return []creature.DamagePart{damagePart(t, "1d4-5", "slashing"), damagePart(t, "1d6", "fire")}
			},
			rolls:     []int{15, 1, 3},
			wantTaken: map[string]int{"slashing": 0, "fire": 3},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			attacker := testCombatant("attacker", SideEnemies, creature.KindNPC, 50, 20)
			target := testCombatant("target", SideParty, creature.KindPC, 50, 20)
			target.Creature.Defense.Weakness = map[string]creature.Adjustment{"slashing": {Value: 5}}
			strike := fist()
			strike.Damage = testCase.damage(t)
			fight, events := testFight(script(t, testCase.rolls...), attacker, target)

			fight.strike(attacker, target, strike, strikeOptions{})

			dealt := lastDamage(t, events)
			if len(dealt.Parts) != len(testCase.wantTaken) {
				t.Fatalf("Parts = %+v, want %d parts", dealt.Parts, len(testCase.wantTaken))
			}
			for _, part := range dealt.Parts {
				if part.Taken != testCase.wantTaken[part.Type] {
					t.Errorf("%s taken = %d, want %d", part.Type, part.Taken, testCase.wantTaken[part.Type])
				}
			}
		})
	}
}

func TestTempHPAbsorbsFirst(t *testing.T) {
	attacker := testCombatant("attacker", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 100, 20)
	target.TempHP = 10
	fight, events := testFight(script(t), attacker, target)

	fight.applyDamage(effectSource{actor: attacker}, target, damageRoll{parts: []rolledPart{{damageType: "bludgeoning", amount: 15}}}, false)

	if dealt := lastDamage(t, events); dealt.Absorbed != 10 || dealt.Total != 15 {
		t.Errorf("Absorbed, Total = %d, %d, want 10, 15", dealt.Absorbed, dealt.Total)
	}
	if target.HP != 95 || target.TempHP != 0 {
		t.Errorf("HP, TempHP = %d, %d, want 95, 0", target.HP, target.TempHP)
	}

	target.TempHP = 8
	fight.gainTempHP(target, 5)
	if target.TempHP != 8 {
		t.Errorf("TempHP = %d, want 8 (5 does not replace 8)", target.TempHP)
	}
	if gained := eventsOf[TempHPGained](events.events); len(gained) != 0 {
		t.Errorf("TempHPGained = %+v, want none", gained)
	}
}

func TestHealingCapsAtMaxHP(t *testing.T) {
	healer := testCombatant("healer", SideParty, creature.KindPC, 30, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 30, 20)
	target.HP = 20
	fight, events := testFight(script(t), healer, target)

	fight.heal(effectSource{actor: healer}, target, 50)

	if target.HP != 30 {
		t.Errorf("HP = %d, want 30", target.HP)
	}
	if healed := eventsOf[Healed](events.events); len(healed) != 1 || healed[0].HPAfter != 30 {
		t.Errorf("Healed = %+v, want one at 30 HP", healed)
	}
}

func TestHealingRevivesADyingPC(t *testing.T) {
	healer := testCombatant("healer", SideParty, creature.KindPC, 30, 20)
	target := downedPC(30, 2)
	fight, events := testFight(script(t), healer, target)

	fight.heal(effectSource{actor: healer}, target, 5)

	if target.HP != 5 || target.Dying != 0 || target.Wounded != 1 {
		t.Errorf("HP, Dying, Wounded = %d, %d, %d, want 5, 0, 1", target.HP, target.Dying, target.Wounded)
	}
	if revived := eventsOf[Revived](events.events); len(revived) != 1 || revived[0].DyingFrom != 2 || revived[0].Wounded != 1 {
		t.Errorf("Revived = %+v, want one from dying 2 at wounded 1", revived)
	}
}

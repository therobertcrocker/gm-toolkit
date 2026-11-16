package engine

import (
	"reflect"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func TestDegreeOf(t *testing.T) {
	cases := []struct {
		name           string
		die, total, dc int
		want           Degree
	}{
		{"beat by 10", 12, 30, 20, CriticalSuccess},
		{"meet the DC", 10, 20, 20, Success},
		{"miss by 1", 10, 19, 20, Failure},
		{"miss by 10", 5, 10, 20, CriticalFailure},
		{"natural 20 lifts a failure", 20, 25, 30, Success},
		{"natural 20 stops at critical success", 20, 40, 20, CriticalSuccess},
		{"natural 1 drops a success", 1, 21, 20, Failure},
		{"natural 1 stops at critical failure", 1, 5, 20, CriticalFailure},
	}
	for _, testCase := range cases {
		if got := DegreeOf(testCase.die, testCase.total, testCase.dc); got != testCase.want {
			t.Errorf("%s: DegreeOf(%d, %d, %d) = %v, want %v", testCase.name, testCase.die, testCase.total, testCase.dc, got, testCase.want)
		}
	}
}

func TestAttackPenalty(t *testing.T) {
	cases := []struct {
		attacksMade int
		agile       bool
		want        int
	}{
		{0, false, 0}, {1, false, -5}, {2, false, -10}, {3, false, -10},
		{0, true, 0}, {1, true, -4}, {2, true, -8},
	}
	for _, testCase := range cases {
		if got := attackPenalty(testCase.attacksMade, testCase.agile); got != testCase.want {
			t.Errorf("attackPenalty(%d, %v) = %d, want %d", testCase.attacksMade, testCase.agile, got, testCase.want)
		}
	}
}

func TestStrikeMissDealsNoDamage(t *testing.T) {
	attacker := testCombatant("attacker", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 30, 22)
	fight, events := testFight(script(t, 5), attacker, target)

	fight.strike(attacker, target, fist())

	want := []Event{StrikeRolled{Attacker: "attacker", Target: "target", Strike: "fist", Die: 5, Bonus: 10, Total: 15, AC: 22, Degree: Failure}}
	if !reflect.DeepEqual(events.events, want) {
		t.Errorf("events = %+v, want %+v", events.events, want)
	}
	if target.HP != 30 {
		t.Errorf("HP = %d, want 30", target.HP)
	}
}

func TestStrikeCriticalDoublesDamage(t *testing.T) {
	attacker := testCombatant("attacker", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 30, 15)
	fight, events := testFight(script(t, 15, 4), attacker, target)

	fight.strike(attacker, target, fist())

	want := []Event{
		StrikeRolled{Attacker: "attacker", Target: "target", Strike: "fist", Die: 15, Bonus: 10, Total: 25, AC: 15, Degree: CriticalSuccess},
		DamageDealt{Source: "attacker", Target: "target", Strike: "fist", Critical: true, Parts: []DamageAmount{{Type: "bludgeoning", Rolled: 12, Taken: 12}}, Total: 12, HPAfter: 18},
	}
	if !reflect.DeepEqual(events.events, want) {
		t.Errorf("events = %+v, want %+v", events.events, want)
	}
}

func TestSecondAttackTakesPenalty(t *testing.T) {
	attacker := testCombatant("attacker", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 30, 40)
	fight, events := testFight(script(t, 10, 10), attacker, target)

	fight.strike(attacker, target, fist())
	fight.strike(attacker, target, fist())

	second := events.events[1].(StrikeRolled)
	if second.Penalty != -5 || second.Total != 15 {
		t.Errorf("second strike penalty, total = %d, %d, want -5, 15", second.Penalty, second.Total)
	}
	if attacker.AttacksThisTurn != 2 {
		t.Errorf("AttacksThisTurn = %d, want 2", attacker.AttacksThisTurn)
	}
}

func TestAdjustedDamage(t *testing.T) {
	defense := creature.Defense{
		Immune: []string{"poison"},
		Resist: map[string]creature.Adjustment{
			"physical": {Value: 5},
			"slashing": {Value: 10},
			"all":      {Value: 2, Except: []string{"holy", "fire"}},
		},
		Weakness: map[string]creature.Adjustment{"cold": {Value: 7}},
	}
	cases := []struct {
		name       string
		damageType string
		traits     []string
		amount     int
		want       int
	}{
		{"immunity", "poison", nil, 20, 0},
		{"only the highest resistance", "slashing", nil, 20, 10},
		{"physical covers piercing", "piercing", nil, 20, 15},
		{"weakness then resistance", "cold", nil, 20, 25},
		{"exception by damage type", "fire", nil, 20, 20},
		{"exception by strike trait", "mental", []string{"holy"}, 20, 20},
		{"resistance stops at zero", "mental", nil, 1, 0},
	}
	for _, testCase := range cases {
		if got := adjustedDamage(defense, testCase.damageType, testCase.traits, testCase.amount); got != testCase.want {
			t.Errorf("%s: adjustedDamage = %d, want %d", testCase.name, got, testCase.want)
		}
	}
}

func TestAdjustedDamageImmunityGroups(t *testing.T) {
	cases := []struct {
		name       string
		immune     []string
		damageType string
		want       int
	}{
		{"physical covers slashing", []string{"physical"}, "slashing", 0},
		{"physical does not cover fire", []string{"physical"}, "fire", 20},
		{"all covers fire", []string{"all"}, "fire", 0},
		{"all covers piercing", []string{"all"}, "piercing", 0},
	}
	for _, testCase := range cases {
		defense := creature.Defense{Immune: testCase.immune}
		if got := adjustedDamage(defense, testCase.damageType, nil, 20); got != testCase.want {
			t.Errorf("%s: adjustedDamage = %d, want %d", testCase.name, got, testCase.want)
		}
	}
}

func TestNPCDiesAtZero(t *testing.T) {
	attacker := testCombatant("attacker", SideParty, creature.KindPC, 50, 20)
	target := testCombatant("target", SideEnemies, creature.KindNPC, 5, 20)
	fight, events := testFight(script(t, 15, 6), attacker, target)

	fight.strike(attacker, target, fist())

	if !target.Dead || target.Dying != 0 {
		t.Errorf("Dead, Dying = %v, %d, want true, 0", target.Dead, target.Dying)
	}
	if last := events.events[len(events.events)-1]; last != (Died{Combatant: "target"}) {
		t.Errorf("last event = %+v, want Died", last)
	}
}

func TestPCKnockedDownGoesDying(t *testing.T) {
	cases := []struct {
		name      string
		wounded   int
		die       int
		wantDying int
	}{
		{"hit", 0, 12, 1},
		{"critical hit", 0, 20, 2},
		{"hit while wounded 2", 2, 12, 3},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			target := testCombatant("target", SideParty, creature.KindPC, 5, 20)
			target.Wounded = testCase.wounded
			other := testCombatant("other", SideParty, creature.KindPC, 30, 20)
			attacker := testCombatant("attacker", SideEnemies, creature.KindNPC, 50, 20)
			fight, events := testFight(script(t, testCase.die, 6), target, other, attacker)

			fight.strike(attacker, target, fist())

			if target.Dying != testCase.wantDying || target.Dead {
				t.Errorf("Dying, Dead = %d, %v, want %d, false", target.Dying, target.Dead, testCase.wantDying)
			}
			if want := []string{"other", "target", "attacker"}; !slices.Equal(orderIDs(fight), want) {
				t.Errorf("order = %v, want %v", orderIDs(fight), want)
			}
			if !slices.Contains(events.events, Event(InitiativeMoved{Combatant: "target", Before: "attacker"})) {
				t.Errorf("events = %+v, want an InitiativeMoved", events.events)
			}
		})
	}
}

func TestDamageWhileDown(t *testing.T) {
	cases := []struct {
		name      string
		dying     int
		wounded   int
		critical  bool
		wantDying int
		wantDead  bool
	}{
		{"hit while dying 1", 1, 0, false, 2, false},
		{"critical while dying 1", 1, 0, true, 3, false},
		{"hit while dying 3", 3, 0, false, 4, true},
		{"hit while stable and wounded 1", 0, 1, false, 2, false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			target := testCombatant("target", SideParty, creature.KindPC, 30, 20)
			target.HP, target.Dying, target.Wounded = 0, testCase.dying, testCase.wounded
			fight, _ := testFight(script(t), target)

			fight.hitWhileDown(target, testCase.critical)

			if target.Dying != testCase.wantDying || target.Dead != testCase.wantDead {
				t.Errorf("Dying, Dead = %d, %v, want %d, %v", target.Dying, target.Dead, testCase.wantDying, testCase.wantDead)
			}
		})
	}
}

func TestRecoveryCheck(t *testing.T) {
	cases := []struct {
		name        string
		dying       int
		die         int
		wantDying   int
		wantWounded int
		wantDead    bool
	}{
		{"natural 20 from dying 1", 1, 20, 0, 1, false},
		{"success from dying 1", 1, 12, 0, 1, false},
		{"success from dying 2", 2, 12, 1, 0, false},
		{"failure from dying 1", 1, 5, 2, 0, false},
		{"natural 1 from dying 1", 1, 1, 3, 0, false},
		{"natural 1 from dying 3", 3, 1, 4, 0, true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			target := testCombatant("target", SideParty, creature.KindPC, 30, 20)
			target.HP, target.Dying = 0, testCase.dying
			fight, _ := testFight(script(t, testCase.die), target)

			fight.recoveryCheck(target)

			if target.Dying != testCase.wantDying || target.Wounded != testCase.wantWounded || target.Dead != testCase.wantDead {
				t.Errorf("Dying, Wounded, Dead = %d, %d, %v, want %d, %d, %v",
					target.Dying, target.Wounded, target.Dead, testCase.wantDying, testCase.wantWounded, testCase.wantDead)
			}
		})
	}
}

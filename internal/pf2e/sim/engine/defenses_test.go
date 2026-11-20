package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func flatStrike(traits []string, parts ...creature.DamagePart) creature.Strike {
	return creature.Strike{ID: "blade", Name: "Blade", Bonus: 10, Traits: traits, Damage: parts}
}

func rolled(damageType string, amount int) damageRoll {
	return damageRoll{parts: []rolledPart{{damageType: damageType, amount: amount}}}
}

func TestTraitWeaknessAppliesOnce(t *testing.T) {
	attacker := testCombatant("attacker", SideParty, creature.KindPC, 50, 20)
	target := testCombatant("target", SideEnemies, creature.KindNPC, 200, 20)
	target.Creature.Defense.Weakness = map[string]creature.Adjustment{"holy": {Value: 15}}
	strike := flatStrike([]string{"holy"}, damagePart(t, "1d6", "slashing"), damagePart(t, "1d6", "spirit"))
	fight, events := testFight(script(t, 15, 3, 4), attacker, target)

	fight.strike(attacker, target, strike, strikeOptions{})

	dealt := lastDamage(t, events)
	if dealt.Trait != 15 || dealt.Total != 22 {
		t.Errorf("Trait, Total = %d, %d, want 15, 22 (3 + 4, then holy 15 once)", dealt.Trait, dealt.Total)
	}
}

func TestResistAllExceptTraits(t *testing.T) {
	cases := []struct {
		name      string
		traits    []string
		wantTotal int
	}{
		{name: "cold iron", traits: []string{"cold-iron"}, wantTotal: 40},
		{name: "plain", wantTotal: 16},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			attacker := testCombatant("attacker", SideParty, creature.KindPC, 50, 20)
			target := testCombatant("target", SideEnemies, creature.KindNPC, 200, 20)
			target.Creature.Defense.Resist = map[string]creature.Adjustment{"all": {Value: 12, Except: []string{"cold-iron"}}}
			strike := flatStrike(testCase.traits, damagePart(t, "20", "slashing"), damagePart(t, "20", "fire"))
			fight, events := testFight(script(t, 15), attacker, target)

			fight.strike(attacker, target, strike, strikeOptions{})

			if dealt := lastDamage(t, events); dealt.Total != testCase.wantTotal {
				t.Errorf("Total = %d, want %d", dealt.Total, testCase.wantTotal)
			}
		})
	}
}

func TestAdaptiveResistance(t *testing.T) {
	attacker := testCombatant("attacker", SideParty, creature.KindPC, 50, 20)
	target := testCombatant("target", SideEnemies, creature.KindNPC, 500, 20)
	target.Creature.Defense.Adaptive = &creature.Adaptive{Value: 25, Types: []string{"acid", "electricity", "sonic"}}
	fight, events := testFight(script(t), attacker, target)
	source := effectSource{actor: attacker}

	for _, hit := range []struct {
		damageType string
		wantTotal  int
		wantType   string
	}{
		{"acid", 30, "acid"},
		{"acid", 5, "acid"},
		{"electricity", 30, "electricity"},
		{"acid", 30, "acid"},
	} {
		fight.applyDamage(source, target, rolled(hit.damageType, 30), false)
		if dealt := lastDamage(t, events); dealt.Total != hit.wantTotal || target.adaptiveType != hit.wantType {
			t.Errorf("%s hit: Total, adapted = %d, %q, want %d, %q", hit.damageType, dealt.Total, target.adaptiveType, hit.wantTotal, hit.wantType)
		}
	}
	if adapted := eventsOf[ResistanceAdapted](events.events); len(adapted) != 3 {
		t.Errorf("ResistanceAdapted = %+v, want three (acid, electricity, acid)", adapted)
	}

	target.Creature.Defense.Resist = map[string]creature.Adjustment{"all": {Value: 12}}
	fight.applyDamage(source, target, rolled("acid", 30), false)
	if dealt := lastDamage(t, events); dealt.Total != 5 {
		t.Errorf("acid against all 12 and adaptive 25: Total = %d, want 5", dealt.Total)
	}
	fight.applyDamage(source, target, rolled("fire", 30), false)
	if dealt := lastDamage(t, events); dealt.Total != 18 {
		t.Errorf("fire against all 12: Total = %d, want 18", dealt.Total)
	}
}

func weakenStep(weakness, resistance, rounds int, types ...string) creature.Step {
	return creature.Step{
		Weaken:   &creature.DefenseChange{Weakness: weakness, Resistance: resistance, Types: types},
		Duration: &creature.Duration{Rounds: rounds},
	}
}

func TestAppliedWeakness(t *testing.T) {
	source := testCombatant("source", SideParty, creature.KindPC, 50, 20)
	target := testCombatant("target", SideEnemies, creature.KindNPC, 200, 20)
	fight, events := testFight(script(t), source, target)

	fight.weaken(source, target, weakenStep(3, 0, 3, "fire"))
	fight.applyDamage(effectSource{actor: source}, target, rolled("fire", 10), false)
	if dealt := lastDamage(t, events); dealt.Total != 13 {
		t.Errorf("fire with weakness 3: Total = %d, want 13", dealt.Total)
	}

	fight.startTurn(source)
	fight.startTurn(source)
	if len(target.defenseChanges) != 1 {
		t.Fatalf("after two of the source's turns: %d changes, want 1", len(target.defenseChanges))
	}
	fight.startTurn(source)
	if len(target.defenseChanges) != 0 {
		t.Errorf("after three of the source's turns: %d changes, want 0", len(target.defenseChanges))
	}
	if restored := eventsOf[DefenseRestored](events.events); len(restored) != 1 {
		t.Errorf("DefenseRestored = %+v, want one", restored)
	}
}

func TestLoweredResistance(t *testing.T) {
	source := testCombatant("source", SideParty, creature.KindPC, 50, 20)
	target := testCombatant("target", SideEnemies, creature.KindNPC, 200, 20)
	target.Creature.Defense.Resist = map[string]creature.Adjustment{"all": {Value: 12}}
	fight, events := testFight(script(t), source, target)

	fight.weaken(source, target, weakenStep(0, 3, 1, "bludgeoning", "piercing", "slashing"))

	fight.applyDamage(effectSource{actor: source}, target, rolled("slashing", 20), false)
	if dealt := lastDamage(t, events); dealt.Total != 11 {
		t.Errorf("slashing against all 12 lowered by 3: Total = %d, want 11", dealt.Total)
	}
	fight.applyDamage(effectSource{actor: source}, target, rolled("fire", 20), false)
	if dealt := lastDamage(t, events); dealt.Total != 8 {
		t.Errorf("fire against all 12: Total = %d, want 8", dealt.Total)
	}
}

func TestStrikeRiders(t *testing.T) {
	successRider := []creature.Step{{Condition: "frightened", Value: 1}}
	criticalRider := []creature.Step{{Condition: "sickened", Value: 1}}
	cases := []struct {
		name           string
		critical       bool
		rolls          []int
		wantFrightened bool
		wantSickened   bool
	}{
		{name: "hit", critical: true, rolls: []int{15, 3}, wantFrightened: true},
		{name: "critical hit with a critical list", critical: true, rolls: []int{20, 3}, wantSickened: true},
		{name: "critical hit with no critical list", rolls: []int{20, 3}, wantFrightened: true},
		{name: "miss", critical: true, rolls: []int{5}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			attacker := testCombatant("attacker", SideParty, creature.KindPC, 50, 20)
			target := testCombatant("target", SideEnemies, creature.KindNPC, 200, 20)
			strike := fist()
			strike.Success = successRider
			if testCase.critical {
				strike.CriticalSuccess = criticalRider
			}
			fight, _ := testFight(script(t, testCase.rolls...), attacker, target)

			fight.strike(attacker, target, strike, strikeOptions{})

			if got := target.Condition("frightened") > 0; got != testCase.wantFrightened {
				t.Errorf("frightened = %v, want %v", got, testCase.wantFrightened)
			}
			if got := target.Condition("sickened") > 0; got != testCase.wantSickened {
				t.Errorf("sickened = %v, want %v", got, testCase.wantSickened)
			}
		})
	}
}

func TestRiderUnlessFrightened(t *testing.T) {
	attacker := testCombatant("attacker", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 200, 20)
	giveCondition(target, "frightened", 1)
	strike := fist()
	strike.Success = []creature.Step{{
		Save:     "will",
		DC:       30,
		Unless:   "frightened",
		Outcomes: creature.Outcomes{Failure: []creature.Step{{Condition: "frightened", Value: 2}}},
	}}
	fight, events := testFight(script(t, 15, 3), attacker, target)

	fight.strike(attacker, target, strike, strikeOptions{})

	if saves := eventsOf[SaveRolled](events.events); len(saves) != 0 {
		t.Errorf("SaveRolled = %+v, want none against a frightened target", saves)
	}
}

func TestPerConditionDamage(t *testing.T) {
	cases := []struct {
		name       string
		frightened int
		wantTotal  int
	}{
		{name: "frightened 2", frightened: 2, wantTotal: 11},
		{name: "not frightened", wantTotal: 5},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			attacker := testCombatant("attacker", SideParty, creature.KindPC, 50, 20)
			attacker.Creature.Modifiers = map[string]creature.CheckModifier{
				"fearsome-brute": {Selector: "damage", Type: "circumstance", Value: 3, PerCondition: "frightened"},
			}
			target := testCombatant("target", SideEnemies, creature.KindNPC, 200, 20)
			if testCase.frightened > 0 {
				giveCondition(target, "frightened", testCase.frightened)
			}
			fight, events := testFight(script(t, 15, 3), attacker, target)

			fight.strike(attacker, target, fist(), strikeOptions{})

			if dealt := lastDamage(t, events); dealt.Total != testCase.wantTotal {
				t.Errorf("Total = %d, want %d (1d6+2 rolled 3, plus 3 per frightened)", dealt.Total, testCase.wantTotal)
			}
		})
	}
}

const runedHammer = `schema = 1
name  = "Kas"
kind  = "pc"
level = 14

[build]
ancestry_hp = 6
class_hp    = 10
key         = "str"
speed       = 25
critical_specialization = "master"

[build.boosts]
ancestry = ["str"]
class    = ["str"]

[build.proficiency]
class-dc            = "expert"
martial             = "master"
weapon-group-hammer = "legendary"

[build.weapon.protectors-fury]
name     = "Protector's Fury"
category = "martial"
group    = "hammer"
dice     = 1
die      = 12
type     = "bludgeoning"
critical_success = [
  { condition = "clumsy", value = 2, duration = { until = "source-turn-end" } },
  { condition = "enfeebled", value = 2, duration = { until = "source-turn-end" } },
  { condition = "frightened", value = 2 },
]
`

func TestRunesFromALoadedLayer(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "kas")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "hand.toml"), []byte(runedHammer), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	loaded, err := creature.Load(directory)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	kas := NewCombatant(SideParty, loaded)
	target := testCombatant("target", SideEnemies, creature.KindNPC, 500, 20)
	fight, events := testFight(script(t, 20, 6, 1), kas, target)

	fight.strike(kas, target, loaded.Strikes["protectors-fury"], strikeOptions{})

	gained := map[string]int{}
	for _, changed := range eventsOf[ConditionChanged](events.events) {
		gained[changed.Condition] = changed.To
	}
	for _, name := range []string{"clumsy", "enfeebled", "frightened"} {
		if gained[name] != 2 {
			t.Errorf("%s = %d, want 2", name, gained[name])
		}
	}
	if saves := eventsOf[SaveRolled](events.events); len(saves) != 1 || saves[0].Save != "fortitude" {
		t.Errorf("SaveRolled = %+v, want one Fortitude save", saves)
	}
}

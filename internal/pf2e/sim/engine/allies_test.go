package engine

import (
	"reflect"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

func healStep(t *testing.T, text string) creature.Step {
	t.Helper()
	expression, err := dice.Parse(text)
	if err != nil {
		t.Fatalf("healStep(%q): %v", text, err)
	}
	return creature.Step{Heal: text, HealRoll: expression}
}

func healAbility(t *testing.T, id, targets string, area bool, heal string) creature.Ability {
	ability := creature.Ability{ID: id, Name: id, Actions: 1, Targets: targets, Steps: []creature.Step{healStep(t, heal)}}
	if area {
		ability.Area = &creature.Area{Shape: "emanation", Size: 30}
	}
	return ability
}

func healed(events *recorder) map[string]int {
	amounts := map[string]int{}
	for _, event := range eventsOf[Healed](events.events) {
		amounts[event.Target] += event.Amount
	}
	return amounts
}

func TestSelfAbilityNeedsNoTarget(t *testing.T) {
	healer := testCombatant("healer", SideParty, creature.KindPC, 50, 20)
	foe := testCombatant("foe", SideEnemies, creature.KindNPC, 50, 20)
	healer.HP = 20
	giveAbility(healer, healAbility(t, "drink", creature.TargetsSelf, false, "10"))
	fight, events := testFight(script(t), healer, foe)
	fight.current, healer.ActionsLeft = healer, 3

	if err := fight.perform(healer, Choice{Kind: ChooseAbility, ID: "drink"}); err != nil {
		t.Fatalf("perform: %v", err)
	}

	want := []Event{
		ActionUsed{Combatant: "healer", Action: "drink", Actions: 1},
		Healed{Source: "healer", Target: "healer", Ability: "drink", Amount: 10, HPAfter: 30},
	}
	if !reflect.DeepEqual(events.events, want) {
		t.Errorf("events = %+v, want %+v", events.events, want)
	}
}

func TestAllyAbility(t *testing.T) {
	setup := func(t *testing.T) (*Fight, *recorder, *recordingDecider, *Combatant, *Combatant, *Combatant) {
		healer := testCombatant("healer", SideParty, creature.KindPC, 50, 20)
		ally := downedPC(50, 2)
		ally.ID = "ally"
		foe := testCombatant("foe", SideEnemies, creature.KindNPC, 50, 20)
		giveAbility(healer, healAbility(t, "feed", creature.TargetsAlly, false, "10"))
		fight, events := testFight(script(t), healer, ally, foe)
		decider := &recordingDecider{scriptedDecider: &scriptedDecider{}}
		fight.decider = decider
		fight.current, healer.ActionsLeft = healer, 3
		return fight, events, decider, healer, ally, foe
	}
	t.Run("a dying ally is lifted", func(t *testing.T) {
		fight, events, decider, healer, ally, _ := setup(t)

		if err := fight.perform(healer, Choice{Kind: ChooseAbility, ID: "feed", Target: ally}); err != nil {
			t.Fatalf("perform: %v", err)
		}

		if got := eventsOf[Revived](events.events); !slices.Equal(got, []Revived{{Combatant: "ally", DyingFrom: 2, Wounded: 1}}) {
			t.Errorf("Revived = %+v, want ally from dying 2", got)
		}
		if ally.HP != 10 || ally.Dying != 0 || decider.nextTargets != 0 {
			t.Errorf("ally HP, dying, retarget asks = %d, %d, %d, want 10, 0, 0", ally.HP, ally.Dying, decider.nextTargets)
		}
	})
	t.Run("a foe, a dead ally or no target is refused", func(t *testing.T) {
		fight, events, _, healer, ally, foe := setup(t)
		ally.Dead = true
		for _, target := range []*Combatant{foe, ally, nil} {
			wantRefused(t, fight.perform(healer, Choice{Kind: ChooseAbility, ID: "feed", Target: target}), "no fitting target")
		}
		if healer.ActionsLeft != 3 || len(events.events) != 0 {
			t.Errorf("ActionsLeft, events = %d, %+v, want 3, none", healer.ActionsLeft, events.events)
		}
	})
	t.Run("a dead ally is not swapped for another target", func(t *testing.T) {
		fight, events, decider, healer, ally, _ := setup(t)
		ally.Dead = true

		fight.useAbility(abilityUse{actor: healer, ability: healer.Creature.Abilities["feed"], target: ally, actions: 1})

		if got := eventsOf[Healed](events.events); len(got) != 0 || decider.nextTargets != 0 {
			t.Errorf("Healed, retarget asks = %+v, %d, want none, 0", got, decider.nextTargets)
		}
	})
}

func TestAllyArea(t *testing.T) {
	cases := []struct {
		targets    string
		wantHealed map[string]int
	}{
		{creature.TargetsAlly, map[string]int{"healer": 4, "ally": 4, "dying": 4}},
		{creature.TargetsOthers, map[string]int{"ally": 4, "dying": 4}},
	}
	for _, testCase := range cases {
		t.Run(testCase.targets, func(t *testing.T) {
			healer := testCombatant("healer", SideParty, creature.KindPC, 50, 20)
			ally := testCombatant("ally", SideParty, creature.KindPC, 50, 20)
			dying := downedPC(50, 1)
			dying.ID = "dying"
			dead := testCombatant("dead", SideParty, creature.KindPC, 50, 20)
			dead.HP, dead.Dead = 0, true
			foe := testCombatant("foe", SideEnemies, creature.KindNPC, 50, 20)
			healer.HP, ally.HP, foe.HP = 30, 30, 30
			giveAbility(healer, healAbility(t, "wave", testCase.targets, true, "1d6"))
			roller := script(t, 4)
			fight, events := testFight(roller, healer, ally, dying, dead, foe)
			fight.current, healer.ActionsLeft = healer, 3

			if err := fight.perform(healer, Choice{Kind: ChooseAbility, ID: "wave"}); err != nil {
				t.Fatalf("perform: %v", err)
			}

			if got := healed(events); !reflect.DeepEqual(got, testCase.wantHealed) {
				t.Errorf("healed = %v, want %v", got, testCase.wantHealed)
			}
			if len(roller.values) != 0 || dying.HP != 4 || dying.Dying != 0 || foe.HP != 30 {
				t.Errorf("unrolled, dying HP, dying, foe HP = %v, %d, %d, %d, want none, 4, 0, 30", roller.values, dying.HP, dying.Dying, foe.HP)
			}
		})
	}
}

func TestAreaStillCatchesFoes(t *testing.T) {
	actor := testCombatant("actor", SideParty, creature.KindPC, 50, 20)
	ally := testCombatant("ally", SideParty, creature.KindPC, 50, 20)
	first := testCombatant("first", SideEnemies, creature.KindNPC, 50, 20)
	second := testCombatant("second", SideEnemies, creature.KindNPC, 50, 20)
	dead := testCombatant("dead", SideEnemies, creature.KindNPC, 50, 20)
	dead.HP, dead.Dead = 0, true
	giveAbility(actor, creature.Ability{ID: "burst", Name: "Burst", Actions: 2, Area: &creature.Area{Shape: "burst", Size: 10}, Steps: []creature.Step{{Damage: []creature.DamagePart{damagePart(t, "5", "fire")}}}})
	fight, events := testFight(script(t), actor, ally, first, second, dead)
	fight.current, actor.ActionsLeft = actor, 3

	if err := fight.perform(actor, Choice{Kind: ChooseAbility, ID: "burst", Target: first}); err != nil {
		t.Fatalf("perform: %v", err)
	}

	var hit []string
	for _, dealt := range eventsOf[DamageDealt](events.events) {
		hit = append(hit, dealt.Target)
	}
	if !slices.Equal(hit, []string{"first", "second"}) {
		t.Errorf("hit = %v, want [first second]: every standing foe and no ally", hit)
	}
}

func TestFirstFittingSkipsFriendlyAbilities(t *testing.T) {
	actor := testCombatant("actor", SideParty, creature.KindPC, 50, 20)
	foe := testCombatant("foe", SideEnemies, creature.KindNPC, 50, 20)
	giveAbility(actor, healAbility(t, "drink", creature.TargetsSelf, false, "10"))
	giveAbility(actor, healAbility(t, "feed", creature.TargetsAlly, false, "10"))
	fight, _ := testFight(script(t), actor, foe)
	actor.ActionsLeft = 3

	choice, acting := FirstFitting{}.Decide(fight, actor)

	if want := (Choice{Kind: ChooseStrike, ID: "fist", Target: foe}); !acting || choice != want {
		t.Errorf("Decide = %+v, %v, want %+v", choice, acting, want)
	}
}

func TestRecoveryEased(t *testing.T) {
	cases := []struct {
		name        string
		recoveryDC  int
		wantDC      int
		wantDCAfter int
	}{
		{"the default DC is capped", 0, 10, 13},
		{"a lower own DC stays", 4, 7, 7},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			bloomer := testCombatant("bloomer", SideParty, creature.KindPC, 50, 20)
			target := downedPC(50, 3)
			target.Creature.RecoveryDC = testCase.recoveryDC
			target.Creature.DyingMax = 10
			fight, events := testFight(script(t, 20, 20), bloomer, target)

			fight.easeRecovery(effectSource{actor: bloomer, ability: "bloom"}, target, creature.Step{RecoveryDC: 10, Duration: &creature.Duration{Rounds: 1}})
			fight.recoveryCheck(target)
			fight.startTurn(bloomer)
			target.Dying = 3
			fight.recoveryCheck(target)

			if got := eventsOf[RecoveryEased](events.events); !slices.Equal(got, []RecoveryEased{{Combatant: "target", Ability: "bloom", DC: 10}}) {
				t.Errorf("RecoveryEased = %+v, want one at 10", got)
			}
			checks := eventsOf[RecoveryChecked](events.events)
			if len(checks) != 2 || checks[0].DC != testCase.wantDC || checks[1].DC != testCase.wantDCAfter {
				t.Errorf("RecoveryChecked = %+v, want DC %d, then %d after the duration", checks, testCase.wantDC, testCase.wantDCAfter)
			}
		})
	}
}

func sustainFight(t *testing.T, values ...int) (*Fight, *recorder, *Combatant, *Combatant) {
	actor := testCombatant("actor", SideParty, creature.KindPC, 50, 20)
	foe := testCombatant("foe", SideEnemies, creature.KindNPC, 100, 20)
	giveAbility(actor, creature.Ability{
		ID:      "song",
		Name:    "Song",
		Actions: 2,
		Steps:   []creature.Step{{Condition: "off-guard"}},
		Sustain: []creature.Step{{Damage: []creature.DamagePart{damagePart(t, "3", "sonic")}}},
	})
	fight, events := testFight(script(t, values...), actor, foe)
	fight.current, actor.ActionsLeft = actor, 3
	if err := fight.perform(actor, Choice{Kind: ChooseAbility, ID: "song", Target: foe}); err != nil {
		t.Fatalf("perform song: %v", err)
	}
	fight.endTurn(actor)
	events.events = nil
	actor.ActionsLeft = 3
	return fight, events, actor, foe
}

func TestSustain(t *testing.T) {
	sustain := Choice{Kind: ChooseSustain, ID: "song"}
	t.Run("FirstFitting Sustains first, once a turn", func(t *testing.T) {
		fight, events, actor, _ := sustainFight(t)

		choice, acting := FirstFitting{}.Decide(fight, actor)
		if !acting || choice != sustain {
			t.Fatalf("Decide = %+v, %v, want %+v", choice, acting, sustain)
		}
		if err := fight.perform(actor, choice); err != nil {
			t.Fatalf("perform: %v", err)
		}
		wantRefused(t, fight.perform(actor, sustain), "cannot Sustain")

		head := events.events[:2]
		want := []Event{ActionUsed{Combatant: "actor", Action: "song", Actions: 1, Target: "foe"}, Sustained{Combatant: "actor", Ability: "song"}}
		if !reflect.DeepEqual(head, want) {
			t.Errorf("first events = %+v, want %+v", head, want)
		}
		if got := damageTo(events, "foe"); !slices.Equal(got, []int{3}) || actor.ActionsLeft != 2 {
			t.Errorf("foe damage, ActionsLeft = %v, %d, want [3], 2", got, actor.ActionsLeft)
		}
		fight.endTurn(actor)
		if got := eventsOf[SustainEnded](events.events); len(got) != 0 {
			t.Errorf("SustainEnded = %+v, want none after a sustained turn", got)
		}
	})
	t.Run("a turn with no Sustain ends it", func(t *testing.T) {
		fight, events, actor, _ := sustainFight(t)

		fight.endTurn(actor)

		if got := eventsOf[SustainEnded](events.events); !slices.Equal(got, []SustainEnded{{Combatant: "actor", Ability: "song"}}) {
			t.Errorf("SustainEnded = %+v, want song", got)
		}
		wantRefused(t, fight.perform(actor, sustain), "cannot Sustain")
		if got := actor.Sustaining(); len(got) != 0 {
			t.Errorf("Sustaining = %v, want none", got)
		}
	})
	t.Run("an actor dying at the end of its turn keeps it to the end of its next turn", func(t *testing.T) {
		fight, events, actor, _ := sustainFight(t, 5)
		if err := fight.perform(actor, sustain); err != nil {
			t.Fatalf("perform: %v", err)
		}
		actor.HP, actor.Dying = 0, 1

		fight.endTurn(actor)
		if got := eventsOf[SustainEnded](events.events); len(got) != 0 {
			t.Fatalf("after the sustained turn's end: SustainEnded = %+v, want none", got)
		}
		if got := actor.Sustaining(); !slices.Equal(got, []string{"song"}) {
			t.Errorf("Sustaining = %v, want [song]", got)
		}
		if err := fight.takeTurn(actor); err != nil {
			t.Fatalf("next turn: %v", err)
		}

		if got := eventsOf[Sustained](events.events); len(got) != 1 {
			t.Errorf("Sustained = %+v, want only the first turn's", got)
		}
		if got := eventsOf[RecoveryChecked](events.events); len(got) != 1 || actor.Dying != 2 {
			t.Errorf("RecoveryChecked, dying = %+v, %d, want one failed check, 2", got, actor.Dying)
		}
		if got := eventsOf[SustainEnded](events.events); !slices.Equal(got, []SustainEnded{{Combatant: "actor", Ability: "song"}}) {
			t.Errorf("after the next turn's end, spent down: SustainEnded = %+v, want song", got)
		}
	})
	t.Run("a Sustain whose target fell runs no step", func(t *testing.T) {
		fight, events, actor, foe := sustainFight(t)
		foe.HP, foe.Dead = 0, true

		if err := fight.perform(actor, sustain); err != nil {
			t.Fatalf("perform: %v", err)
		}

		if got := eventsOf[Sustained](events.events); len(got) != 1 {
			t.Errorf("Sustained = %+v, want one", got)
		}
		if got := eventsOf[DamageDealt](events.events); len(got) != 0 {
			t.Errorf("DamageDealt = %+v, want none", got)
		}
	})
}

func TestSustainNeedsAnOpenAction(t *testing.T) {
	fight, _, actor, _ := sustainFight(t)
	actor.ActionsLeft, actor.StrikeOnlyActions = 1, 1

	wantRefused(t, fight.perform(actor, Choice{Kind: ChooseSustain, ID: "song"}), "cannot Sustain")
	if choice, _ := (FirstFitting{}).Decide(fight, actor); choice.Kind != ChooseStrike {
		t.Errorf("Decide = %+v, want a Strike with only the Strike-only action left", choice)
	}
	if actor.ActionsLeft != 1 {
		t.Errorf("ActionsLeft = %d, want 1", actor.ActionsLeft)
	}
}

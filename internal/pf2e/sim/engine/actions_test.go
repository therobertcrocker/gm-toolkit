package engine

import (
	"errors"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

func TestRegainActions(t *testing.T) {
	type given struct {
		name  string
		value int
	}
	cases := []struct {
		name           string
		conditions     []given
		wantActions    int
		wantStrikeOnly int
		wantStunned    int
		wantEvent      bool
		wantLost       int
	}{
		{name: "none", wantActions: 3},
		{name: "quickened", conditions: []given{{"quickened", 0}}, wantActions: 4, wantStrikeOnly: 1, wantEvent: true},
		{name: "slowed 1", conditions: []given{{"slowed", 1}}, wantActions: 2, wantEvent: true, wantLost: 1},
		{name: "slowed 1 with quickened loses the quickened action", conditions: []given{{"slowed", 1}, {"quickened", 0}}, wantActions: 3, wantEvent: true, wantLost: 1},
		{name: "slowed 2 with quickened", conditions: []given{{"slowed", 2}, {"quickened", 0}}, wantActions: 2, wantEvent: true, wantLost: 2},
		{name: "stunned 4", conditions: []given{{"stunned", 4}}, wantActions: 0, wantStunned: 1, wantEvent: true, wantLost: 3},
		{name: "stunned 4 with quickened", conditions: []given{{"stunned", 4}, {"quickened", 0}}, wantActions: 0, wantStunned: 0, wantEvent: true, wantLost: 4},
		{name: "stunned 1 with slowed 2 overlaps", conditions: []given{{"stunned", 1}, {"slowed", 2}}, wantActions: 1, wantEvent: true, wantLost: 2},
		{name: "stunned 2 with slowed 1", conditions: []given{{"stunned", 2}, {"slowed", 1}}, wantActions: 1, wantEvent: true, wantLost: 2},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			actor := testCombatant("actor", SideParty, creature.KindPC, 50, 20)
			for _, condition := range testCase.conditions {
				giveCondition(actor, condition.name, condition.value)
			}
			fight, events := testFight(script(t), actor)

			fight.regainActions(actor)

			if actor.ActionsLeft != testCase.wantActions || actor.StrikeOnlyActions != testCase.wantStrikeOnly {
				t.Errorf("ActionsLeft, StrikeOnlyActions = %d, %d, want %d, %d", actor.ActionsLeft, actor.StrikeOnlyActions, testCase.wantActions, testCase.wantStrikeOnly)
			}
			if got := actor.Condition("stunned"); got != testCase.wantStunned {
				t.Errorf("stunned = %d, want %d", got, testCase.wantStunned)
			}
			regained := eventsOf[ActionsRegained](events.events)
			if !testCase.wantEvent {
				if len(regained) != 0 {
					t.Errorf("ActionsRegained = %+v, want none", regained)
				}
				return
			}
			want := []ActionsRegained{{Combatant: "actor", Actions: testCase.wantActions, Quickened: actor.Condition("quickened") > 0, Lost: testCase.wantLost}}
			if !slices.Equal(regained, want) {
				t.Errorf("ActionsRegained = %+v, want %+v", regained, want)
			}
		})
	}

	t.Run("stunned 4 over two turns", func(t *testing.T) {
		actor := testCombatant("actor", SideParty, creature.KindPC, 50, 20)
		giveCondition(actor, "stunned", 4)
		fight, _ := testFight(script(t), actor)

		fight.regainActions(actor)
		if actor.ActionsLeft != 0 || actor.Condition("stunned") != 1 {
			t.Fatalf("first turn: ActionsLeft, stunned = %d, %d, want 0, 1", actor.ActionsLeft, actor.Condition("stunned"))
		}
		fight.regainActions(actor)
		if actor.ActionsLeft != 2 || actor.Condition("stunned") != 0 {
			t.Errorf("second turn: ActionsLeft, stunned = %d, %d, want 2, 0", actor.ActionsLeft, actor.Condition("stunned"))
		}
	})
}

func TestStrikeOnlyAction(t *testing.T) {
	actor := testCombatant("actor", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 50, 40)
	giveCondition(actor, "quickened", 0)
	giveAbility(actor, creature.Ability{ID: "heavy", Name: "Heavy", Actions: 3, Steps: []creature.Step{{Strike: "fist"}}})
	giveAbility(actor, creature.Ability{ID: "jab", Name: "Jab", Actions: 1, Steps: []creature.Step{{Strike: "fist"}}})
	fight, _ := testFight(script(t, 2, 2), actor, target)
	fight.regainActions(actor)

	if err := fight.perform(actor, Choice{Kind: ChooseAbility, ID: "heavy", Target: target}); err != nil {
		t.Fatalf("perform heavy: %v", err)
	}
	if actor.ActionsLeft != 1 || actor.OpenActions() != 0 {
		t.Fatalf("after heavy: ActionsLeft, OpenActions = %d, %d, want 1, 0", actor.ActionsLeft, actor.OpenActions())
	}
	if err := fight.perform(actor, Choice{Kind: ChooseAbility, ID: "jab", Target: target}); !errors.Is(err, ErrBadChoice) {
		t.Errorf("perform jab error = %v, want ErrBadChoice", err)
	}
	if err := fight.perform(actor, Choice{Kind: ChooseStrike, ID: "fist", Target: target}); err != nil {
		t.Fatalf("perform fist: %v", err)
	}
	if actor.ActionsLeft != 0 || actor.StrikeOnlyActions != 0 {
		t.Errorf("after fist: ActionsLeft, StrikeOnlyActions = %d, %d, want 0, 0", actor.ActionsLeft, actor.StrikeOnlyActions)
	}
}

func TestKnockedOutPCIsProne(t *testing.T) {
	attacker := testCombatant("attacker", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 5, 15)
	fight, events := testFight(script(t, 10, 3, 1, 1), attacker, target)

	fight.strike(attacker, target, attacker.Creature.Strikes["fist"], strikeOptions{})

	if target.HP != 0 || target.Dying != 1 {
		t.Fatalf("HP, Dying = %d, %d, want 0, 1", target.HP, target.Dying)
	}
	if target.Condition("prone") != 1 || target.Condition("off-guard") != 1 {
		t.Errorf("prone, off-guard = %d, %d, want 1, 1", target.Condition("prone"), target.Condition("off-guard"))
	}
	var afterDamage []Event
	for _, event := range events.events {
		switch event.(type) {
		case ConditionChanged, DyingChanged:
			afterDamage = append(afterDamage, event)
		}
	}
	wantOrder := []Event{
		ConditionChanged{Combatant: "target", Condition: "off-guard", From: 0, To: 1},
		ConditionChanged{Combatant: "target", Condition: "prone", From: 0, To: 1},
		DyingChanged{Combatant: "target", From: 0, To: 1},
	}
	if !slices.Equal(afterDamage, wantOrder) {
		t.Errorf("condition and dying events = %+v, want %+v", afterDamage, wantOrder)
	}

	fight.heal(effectSource{actor: attacker}, target, 10)
	if target.HP != 5 || target.Condition("prone") != 1 {
		t.Fatalf("healed: HP, prone = %d, %d, want 5, 1", target.HP, target.Condition("prone"))
	}

	events.events = nil
	if err := fight.takeTurn(target); err != nil {
		t.Fatalf("takeTurn: %v", err)
	}
	var turn []Event
	for _, event := range events.events {
		switch event.(type) {
		case ActionUsed, StoodUp:
			turn = append(turn, event)
		}
	}
	wantTurn := []Event{
		ActionUsed{Combatant: "target", Action: "stand", Actions: 1},
		StoodUp{Combatant: "target"},
		ActionUsed{Combatant: "target", Action: "fist", Actions: 1, Target: "attacker"},
		ActionUsed{Combatant: "target", Action: "fist", Actions: 1, Target: "attacker"},
	}
	if !slices.Equal(turn, wantTurn) {
		t.Errorf("turn = %+v, want %+v", turn, wantTurn)
	}
	if target.Condition("prone") != 0 || target.Condition("off-guard") != 0 {
		t.Errorf("after Stand: prone, off-guard = %d, %d, want 0, 0", target.Condition("prone"), target.Condition("off-guard"))
	}
}

func TestStandNeedsProne(t *testing.T) {
	t.Run("not prone", func(t *testing.T) {
		actor := testCombatant("actor", SideParty, creature.KindPC, 50, 20)
		fight, _ := testFight(script(t), actor)
		fight.regainActions(actor)

		if err := fight.perform(actor, Choice{Kind: ChooseStand}); !errors.Is(err, ErrBadChoice) {
			t.Errorf("perform error = %v, want ErrBadChoice", err)
		}
		if actor.ActionsLeft != 3 {
			t.Errorf("ActionsLeft = %d, want 3", actor.ActionsLeft)
		}
	})
	t.Run("prone with only the Strike-only action", func(t *testing.T) {
		actor := testCombatant("actor", SideParty, creature.KindPC, 50, 20)
		giveCondition(actor, "prone", 0)
		actor.ActionsLeft, actor.StrikeOnlyActions = 1, 1
		fight, _ := testFight(script(t), actor)

		if err := fight.perform(actor, Choice{Kind: ChooseStand}); !errors.Is(err, ErrBadChoice) {
			t.Errorf("perform error = %v, want ErrBadChoice", err)
		}
		if actor.ActionsLeft != 1 || actor.Condition("prone") != 1 {
			t.Errorf("ActionsLeft, prone = %d, %d, want 1, 1", actor.ActionsLeft, actor.Condition("prone"))
		}
	})
}

func TestSlowedForOneRound(t *testing.T) {
	source := testCombatant("source", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 50, 20)
	giveAbility(source, creature.Ability{ID: "slow", Name: "Slow", Actions: 3, Steps: []creature.Step{{
		Save:     "fortitude",
		DC:       20,
		Outcomes: creature.Outcomes{Failure: []creature.Step{conditionStep("slowed", 1, &creature.Duration{Rounds: 1})}},
	}}})
	events := &recorder{}
	_, err := Run(Setup{
		Party:     []*creature.Creature{target.Creature},
		Enemies:   []*creature.Creature{source.Creature},
		Roller:    script(t, 1, 20, 12, 1, 1, 20, 1, 1, 1),
		Observer:  events,
		Decider:   FirstFitting{},
		MaxRounds: 2,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	var actionsPerTargetTurn []int
	for _, event := range events.events {
		switch typed := event.(type) {
		case TurnStarted:
			if typed.Combatant == "target" {
				actionsPerTargetTurn = append(actionsPerTargetTurn, 0)
			}
		case ActionUsed:
			if typed.Combatant == "target" {
				actionsPerTargetTurn[len(actionsPerTargetTurn)-1] += typed.Actions
			}
		}
	}
	if want := []int{2, 3}; !slices.Equal(actionsPerTargetTurn, want) {
		t.Errorf("target actions per turn = %v, want %v", actionsPerTargetTurn, want)
	}
	want := []ActionsRegained{{Combatant: "target", Actions: 2, Lost: 1}}
	if got := eventsOf[ActionsRegained](events.events); !slices.Equal(got, want) {
		t.Errorf("ActionsRegained = %+v, want %+v", got, want)
	}
}

func TestConditionValueDice(t *testing.T) {
	t.Run("rolled once per target", func(t *testing.T) {
		source := testCombatant("source", SideEnemies, creature.KindNPC, 50, 20)
		first := testCombatant("first", SideParty, creature.KindPC, 50, 20)
		immune := testCombatant("immune", SideParty, creature.KindPC, 50, 20)
		immune.Creature.Defense.Immune = []string{"drained"}
		second := testCombatant("second", SideParty, creature.KindPC, 50, 20)
		roller := script(t, 2, 3)
		fight, _ := testFight(roller, source, first, immune, second)
		step := creature.Step{Condition: "drained", ValueDice: "1d4", ValueRoll: dice.Expression{Count: 1, Sides: 4}}

		fight.runStep(effectSource{actor: source}, step, []*Combatant{first, immune, second})

		if first.Condition("drained") != 2 || immune.Condition("drained") != 0 || second.Condition("drained") != 3 {
			t.Errorf("drained first, immune, second = %d, %d, %d, want 2, 0, 3", first.Condition("drained"), immune.Condition("drained"), second.Condition("drained"))
		}
		if first.HP != 48 || second.HP != 47 {
			t.Errorf("HP first, second = %d, %d, want 48, 47", first.HP, second.HP)
		}
		if len(roller.values) != 0 {
			t.Errorf("unused scripted rolls = %v, want none", roller.values)
		}
	})
	t.Run("stack raises doomed", func(t *testing.T) {
		source := testCombatant("source", SideEnemies, creature.KindNPC, 50, 20)
		target := testCombatant("target", SideParty, creature.KindPC, 50, 20)
		giveCondition(target, "doomed", 1)
		fight, _ := testFight(script(t), source, target)

		fight.applyCondition(source, target, creature.Step{Condition: "doomed", Value: 1, Stack: true})

		if got := target.Condition("doomed"); got != 2 {
			t.Errorf("doomed = %d, want 2", got)
		}
	})
	t.Run("stack on a dying PC kills", func(t *testing.T) {
		source := testCombatant("source", SideEnemies, creature.KindNPC, 50, 20)
		target := downedPC(50, 2)
		giveCondition(target, "doomed", 1)
		fight, events := testFight(script(t), source, target)

		fight.applyCondition(source, target, creature.Step{Condition: "doomed", Value: 1, Stack: true})

		if target.Condition("doomed") != 2 || !target.Dead {
			t.Errorf("doomed, Dead = %d, %v, want 2, true", target.Condition("doomed"), target.Dead)
		}
		if died := eventsOf[Died](events.events); len(died) != 1 || died[0].Combatant != "target" {
			t.Errorf("Died = %+v, want one for target", died)
		}
	})
}

func TestStunnedCountsDownWhileDown(t *testing.T) {
	t.Run("dying", func(t *testing.T) {
		target := downedPC(30, 1)
		giveCondition(target, "stunned", 4)
		fight, events := testFight(script(t, 5, 5), target)

		if err := fight.takeTurn(target); err != nil {
			t.Fatalf("first takeTurn: %v", err)
		}
		if target.Dead || target.Dying != 2 || target.Condition("stunned") != 1 {
			t.Fatalf("after one turn: Dead, Dying, stunned = %v, %d, %d, want false, 2, 1", target.Dead, target.Dying, target.Condition("stunned"))
		}
		if err := fight.takeTurn(target); err != nil {
			t.Fatalf("second takeTurn: %v", err)
		}
		if target.Dead || target.Dying != 3 || target.Condition("stunned") != 0 {
			t.Errorf("after two turns: Dead, Dying, stunned = %v, %d, %d, want false, 3, 0", target.Dead, target.Dying, target.Condition("stunned"))
		}
		if used := eventsOf[ActionUsed](events.events); len(used) != 0 {
			t.Errorf("ActionUsed = %+v, want none", used)
		}
		want := []ActionsRegained{
			{Combatant: "target", Actions: 0, Lost: 3},
			{Combatant: "target", Actions: 2, Lost: 1},
		}
		if got := eventsOf[ActionsRegained](events.events); !slices.Equal(got, want) {
			t.Errorf("ActionsRegained = %+v, want %+v", got, want)
		}
	})
	t.Run("stable at 0 HP", func(t *testing.T) {
		target := downedPC(30, 0)
		giveCondition(target, "stunned", 2)
		fight, events := testFight(script(t), target)

		if err := fight.takeTurn(target); err != nil {
			t.Fatalf("takeTurn: %v", err)
		}
		if got := target.Condition("stunned"); got != 0 {
			t.Errorf("stunned = %d, want 0", got)
		}
		if used := eventsOf[ActionUsed](events.events); len(used) != 0 {
			t.Errorf("ActionUsed = %+v, want none", used)
		}
	})
	t.Run("dies on the recovery check", func(t *testing.T) {
		target := downedPC(30, 3)
		giveCondition(target, "stunned", 4)
		fight, events := testFight(script(t, 1), target)

		if err := fight.takeTurn(target); err != nil {
			t.Fatalf("takeTurn: %v", err)
		}
		if !target.Dead || target.Condition("stunned") != 4 {
			t.Errorf("Dead, stunned = %v, %d, want true, 4", target.Dead, target.Condition("stunned"))
		}
		if regained := eventsOf[ActionsRegained](events.events); len(regained) != 0 {
			t.Errorf("ActionsRegained = %+v, want none", regained)
		}
	})
}

func TestStunnedOnOwnTurn(t *testing.T) {
	cases := []struct {
		name           string
		current        bool
		immune         bool
		quickened      bool
		actionsLeft    int
		strikeOnly     int
		stunned        int
		wantLeft       int
		wantStrikeOnly int
		wantStunned    int
		wantLost       []ActionsLost
	}{
		{name: "stunned 4 with 2 left", current: true, actionsLeft: 2, stunned: 4, wantLeft: 0, wantStunned: 2, wantLost: []ActionsLost{{Combatant: "actor", Condition: "stunned", Lost: 2, Left: 0}}},
		{name: "stunned 1 with 3 left", current: true, actionsLeft: 3, stunned: 1, wantLeft: 2, wantStunned: 0, wantLost: []ActionsLost{{Combatant: "actor", Condition: "stunned", Lost: 1, Left: 2}}},
		{name: "not the current actor", actionsLeft: 3, stunned: 4, wantLeft: 3, wantStunned: 4},
		{name: "no actions left", current: true, actionsLeft: 0, stunned: 4, wantLeft: 0, wantStunned: 4},
		{name: "quickened loses the Strike-only action first", current: true, quickened: true, actionsLeft: 2, strikeOnly: 1, stunned: 1, wantLeft: 1, wantStrikeOnly: 0, wantStunned: 0, wantLost: []ActionsLost{{Combatant: "actor", Condition: "stunned", Lost: 1, Left: 1}}},
		{name: "immune", current: true, immune: true, actionsLeft: 3, stunned: 4, wantLeft: 3, wantStunned: 0},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			source := testCombatant("source", SideEnemies, creature.KindNPC, 50, 20)
			actor := testCombatant("actor", SideParty, creature.KindPC, 50, 20)
			if testCase.immune {
				actor.Creature.Defense.Immune = []string{"stunned"}
			}
			if testCase.quickened {
				giveCondition(actor, "quickened", 0)
			}
			actor.ActionsLeft, actor.StrikeOnlyActions = testCase.actionsLeft, testCase.strikeOnly
			fight, events := testFight(script(t), source, actor)
			if testCase.current {
				fight.current = actor
			}

			fight.applyCondition(source, actor, conditionStep("stunned", testCase.stunned, nil))

			if actor.ActionsLeft != testCase.wantLeft || actor.StrikeOnlyActions != testCase.wantStrikeOnly {
				t.Errorf("ActionsLeft, StrikeOnlyActions = %d, %d, want %d, %d", actor.ActionsLeft, actor.StrikeOnlyActions, testCase.wantLeft, testCase.wantStrikeOnly)
			}
			if got := actor.Condition("stunned"); got != testCase.wantStunned {
				t.Errorf("stunned = %d, want %d", got, testCase.wantStunned)
			}
			if got := eventsOf[ActionsLost](events.events); !slices.Equal(got, testCase.wantLost) {
				t.Errorf("ActionsLost = %+v, want %+v", got, testCase.wantLost)
			}
		})
	}

	t.Run("the rest counts down next turn", func(t *testing.T) {
		source := testCombatant("source", SideEnemies, creature.KindNPC, 50, 20)
		actor := testCombatant("actor", SideParty, creature.KindPC, 50, 20)
		actor.ActionsLeft = 2
		fight, _ := testFight(script(t), source, actor)
		fight.current = actor

		fight.applyCondition(source, actor, conditionStep("stunned", 4, nil))
		fight.regainActions(actor)

		if actor.ActionsLeft != 1 || actor.Condition("stunned") != 0 {
			t.Errorf("ActionsLeft, stunned = %d, %d, want 1, 0", actor.ActionsLeft, actor.Condition("stunned"))
		}
	})
	t.Run("the ability in progress completes", func(t *testing.T) {
		actor := testCombatant("actor", SideEnemies, creature.KindNPC, 50, 20)
		target := testCombatant("target", SideParty, creature.KindPC, 50, 40)
		giveAbility(actor, creature.Ability{ID: "reckless", Name: "Reckless", Actions: 2, Steps: []creature.Step{
			{Condition: "stunned", Value: 1, Target: "self"},
			{Strike: "fist"},
		}})
		fight, events := testFight(script(t, 2), actor, target)
		fight.current = actor
		fight.regainActions(actor)

		if err := fight.perform(actor, Choice{Kind: ChooseAbility, ID: "reckless", Target: target}); err != nil {
			t.Fatalf("perform: %v", err)
		}
		if strikes := eventsOf[StrikeRolled](events.events); len(strikes) != 1 {
			t.Errorf("StrikeRolled = %+v, want one", strikes)
		}
		if actor.ActionsLeft != 0 || actor.Condition("stunned") != 0 {
			t.Errorf("ActionsLeft, stunned = %d, %d, want 0, 0", actor.ActionsLeft, actor.Condition("stunned"))
		}
		want := []ActionsLost{{Combatant: "actor", Condition: "stunned", Lost: 1, Left: 0}}
		if got := eventsOf[ActionsLost](events.events); !slices.Equal(got, want) {
			t.Errorf("ActionsLost = %+v, want %+v", got, want)
		}
	})
}

func TestStunnedForDuration(t *testing.T) {
	t.Run("through Run", func(t *testing.T) {
		source := testCombatant("source", SideEnemies, creature.KindNPC, 50, 20)
		target := testCombatant("target", SideParty, creature.KindPC, 50, 20)
		giveAbility(source, creature.Ability{ID: "daze", Name: "Daze", Actions: 3, Steps: []creature.Step{{
			Save:     "fortitude",
			DC:       20,
			Outcomes: creature.Outcomes{Failure: []creature.Step{conditionStep("stunned", 0, &creature.Duration{Rounds: 1})}},
		}}})
		events := &recorder{}
		_, err := Run(Setup{
			Party:     []*creature.Creature{target.Creature},
			Enemies:   []*creature.Creature{source.Creature},
			Roller:    script(t, 1, 20, 12, 20, 1, 1, 1),
			Observer:  events,
			Decider:   FirstFitting{},
			MaxRounds: 2,
		})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}

		var actionsPerTargetTurn []int
		for _, event := range events.events {
			switch typed := event.(type) {
			case TurnStarted:
				if typed.Combatant == "target" {
					actionsPerTargetTurn = append(actionsPerTargetTurn, 0)
				}
			case ActionUsed:
				if typed.Combatant == "target" {
					actionsPerTargetTurn[len(actionsPerTargetTurn)-1] += typed.Actions
				}
			}
		}
		if want := []int{0, 3}; !slices.Equal(actionsPerTargetTurn, want) {
			t.Errorf("target actions per turn = %v, want %v", actionsPerTargetTurn, want)
		}
		want := []ActionsRegained{{Combatant: "target", Actions: 0, Lost: 3}}
		if got := eventsOf[ActionsRegained](events.events); !slices.Equal(got, want) {
			t.Errorf("ActionsRegained = %+v, want %+v", got, want)
		}
	})

	type given struct {
		name  string
		value int
	}
	cases := []struct {
		name          string
		conditions    []given
		wantRegained  ActionsRegained
		wantStunned   int
		wantValueless bool
	}{
		{name: "with quickened", conditions: []given{{"stunned", 0}, {"quickened", 0}}, wantRegained: ActionsRegained{Combatant: "actor", Actions: 0, Quickened: true, Lost: 4}, wantStunned: 1, wantValueless: true},
		{name: "beside stunned 2", conditions: []given{{"stunned", 0}, {"stunned", 2}}, wantRegained: ActionsRegained{Combatant: "actor", Actions: 0, Lost: 3}, wantStunned: 1, wantValueless: true},
		{name: "beside slowed 1", conditions: []given{{"stunned", 0}, {"slowed", 1}}, wantRegained: ActionsRegained{Combatant: "actor", Actions: 0, Lost: 3}, wantStunned: 1, wantValueless: true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			actor := testCombatant("actor", SideParty, creature.KindPC, 50, 20)
			for _, condition := range testCase.conditions {
				giveCondition(actor, condition.name, condition.value)
			}
			fight, events := testFight(script(t), actor)

			fight.regainActions(actor)

			if actor.ActionsLeft != 0 || actor.StrikeOnlyActions != 0 {
				t.Errorf("ActionsLeft, StrikeOnlyActions = %d, %d, want 0, 0", actor.ActionsLeft, actor.StrikeOnlyActions)
			}
			if got := eventsOf[ActionsRegained](events.events); !slices.Equal(got, []ActionsRegained{testCase.wantRegained}) {
				t.Errorf("ActionsRegained = %+v, want %+v", got, testCase.wantRegained)
			}
			if got := actor.Condition("stunned"); got != testCase.wantStunned {
				t.Errorf("stunned = %d, want %d", got, testCase.wantStunned)
			}
			if got := actor.stunnedForDuration(); got != testCase.wantValueless {
				t.Errorf("stunnedForDuration() = %v, want %v", got, testCase.wantValueless)
			}
			var stunnedValues []int
			for _, instance := range actor.conditions {
				if instance.name == "stunned" {
					stunnedValues = append(stunnedValues, instance.value)
				}
			}
			if want := []int{0}; !slices.Equal(stunnedValues, want) {
				t.Errorf("stunned instance values = %v, want %v", stunnedValues, want)
			}
		})
	}

	t.Run("gained on the bearer's own turn", func(t *testing.T) {
		source := testCombatant("source", SideEnemies, creature.KindNPC, 50, 20)
		actor := testCombatant("actor", SideParty, creature.KindPC, 50, 20)
		actor.ActionsLeft = 2
		fight, events := testFight(script(t), source, actor)
		fight.current = actor

		fight.applyCondition(source, actor, conditionStep("stunned", 0, &creature.Duration{Rounds: 1}))

		want := []ActionsLost{{Combatant: "actor", Condition: "stunned", Lost: 2, Left: 0}}
		if got := eventsOf[ActionsLost](events.events); !slices.Equal(got, want) {
			t.Errorf("ActionsLost = %+v, want %+v", got, want)
		}
		if actor.ActionsLeft != 0 || actor.Condition("stunned") != 1 || !actor.stunnedForDuration() {
			t.Errorf("ActionsLeft, stunned, stunnedForDuration() = %d, %d, %v, want 0, 1, true", actor.ActionsLeft, actor.Condition("stunned"), actor.stunnedForDuration())
		}
	})
}

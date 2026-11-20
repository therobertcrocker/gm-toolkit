package engine

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

func limitedAbility(id string, actions int, limit *creature.Limit, traits ...string) creature.Ability {
	return creature.Ability{ID: id, Name: id, Actions: actions, Traits: traits, Limit: limit, Steps: []creature.Step{{Condition: "off-guard"}}}
}

func rechargeLimit() *creature.Limit {
	return &creature.Limit{Recharge: "1d4", RechargeRoll: dice.Expression{Count: 1, Sides: 4}}
}

func limitsFight(t *testing.T, values ...int) (*Fight, *recorder, *Combatant, *Combatant) {
	actor := testCombatant("actor", SideParty, creature.KindPC, 50, 20)
	target := testCombatant("target", SideEnemies, creature.KindNPC, 500, 20)
	fight, events := testFight(script(t, values...), actor, target)
	fight.current = actor
	actor.ActionsLeft = 3
	return fight, events, actor, target
}

func wantRefused(t *testing.T, err error, reason string) {
	t.Helper()
	if !errors.Is(err, ErrBadChoice) || !strings.Contains(err.Error(), reason) {
		t.Errorf("error = %v, want ErrBadChoice with %q", err, reason)
	}
}

func TestPerFightLimit(t *testing.T) {
	fight, events, actor, target := limitsFight(t)
	once := limitedAbility("once", 1, &creature.Limit{PerFight: 1})
	giveAbility(actor, once)
	choice := Choice{Kind: ChooseAbility, ID: "once", Target: target}

	if got := actor.UsesLeft(once); got != 1 {
		t.Errorf("UsesLeft before = %d, want 1", got)
	}
	if err := fight.perform(actor, choice); err != nil {
		t.Fatalf("first perform: %v", err)
	}
	if got := actor.UsesLeft(once); got != 0 {
		t.Errorf("UsesLeft after = %d, want 0", got)
	}
	want := []UseSpent{{Combatant: "actor", Ability: "once", Left: 0}}
	if got := eventsOf[UseSpent](events.events); !slices.Equal(got, want) {
		t.Errorf("UseSpent = %+v, want %+v", got, want)
	}
	wantRefused(t, fight.perform(actor, choice), "no uses left")
	if got := actor.UsesLeft(limitedAbility("free-for-all", 1, nil)); got != -1 {
		t.Errorf("UsesLeft of an unlimited ability = %d, want -1", got)
	}
}

func TestPoolIsShared(t *testing.T) {
	base := testCombatant("actor", SideParty, creature.KindPC, 50, 20)
	base.Creature.Pools = map[string]int{"potions": 2}
	actor := NewCombatant(SideParty, base.Creature)
	target := testCombatant("target", SideEnemies, creature.KindNPC, 500, 20)
	fight, events := testFight(script(t), actor, target)
	fight.current = actor
	actor.ActionsLeft = 3
	drink := limitedAbility("drink", 1, &creature.Limit{Pool: "potions"})
	feed := limitedAbility("feed", 1, &creature.Limit{Pool: "potions"})
	giveAbility(actor, drink)
	giveAbility(actor, feed)

	for _, id := range []string{"drink", "feed"} {
		if err := fight.perform(actor, Choice{Kind: ChooseAbility, ID: id, Target: target}); err != nil {
			t.Fatalf("perform %s: %v", id, err)
		}
	}

	want := []PoolSpent{
		{Combatant: "actor", Pool: "potions", Ability: "drink", Left: 1},
		{Combatant: "actor", Pool: "potions", Ability: "feed", Left: 0},
	}
	if got := eventsOf[PoolSpent](events.events); !slices.Equal(got, want) {
		t.Errorf("PoolSpent = %+v, want %+v", got, want)
	}
	for _, ability := range []creature.Ability{drink, feed} {
		if got := actor.Blocked(ability); got != "pool potions is empty" {
			t.Errorf("Blocked(%s) = %q, want pool potions is empty", ability.ID, got)
		}
	}
	if actor.Pool("potions") != 0 || actor.Creature.Pools["potions"] != 2 {
		t.Errorf("Pool, creature pool = %d, %d, want 0, 2", actor.Pool("potions"), actor.Creature.Pools["potions"])
	}
}

func TestPerRoundLimit(t *testing.T) {
	fight, _, actor, target := limitsFight(t)
	bomb := limitedAbility("bomb", 1, &creature.Limit{PerRound: 1})
	giveAbility(actor, bomb)
	choice := Choice{Kind: ChooseAbility, ID: "bomb", Target: target}

	if err := fight.perform(actor, choice); err != nil {
		t.Fatalf("first perform: %v", err)
	}
	if got := actor.Blocked(bomb); got != "used this round" {
		t.Errorf("Blocked = %q, want used this round", got)
	}
	wantRefused(t, fight.perform(actor, choice), "used this round")

	fight.tickLimits(actor)

	if got := actor.Blocked(bomb); got != "" {
		t.Errorf("Blocked after tickLimits = %q, want none", got)
	}
}

func TestRecharge(t *testing.T) {
	cases := []struct {
		name         string
		rolled       int
		blockedTurns int
	}{
		{"a roll of 1 is ready next turn", 1, 0},
		{"a roll of 4 waits three turns", 4, 3},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			fight, events, actor, target := limitsFight(t, testCase.rolled)
			fight.decider = &scriptedDecider{}
			breath := limitedAbility("breath", 1, rechargeLimit())
			giveAbility(actor, breath)

			if err := fight.perform(actor, Choice{Kind: ChooseAbility, ID: "breath", Target: target}); err != nil {
				t.Fatalf("perform: %v", err)
			}
			if got := eventsOf[RechargeRolled](events.events); !slices.Equal(got, []RechargeRolled{{Combatant: "actor", Ability: "breath", Rounds: testCase.rolled}}) {
				t.Errorf("RechargeRolled = %+v, want rounds %d", got, testCase.rolled)
			}
			if got := actor.Blocked(breath); got != "recharging" {
				t.Errorf("Blocked this turn = %q, want recharging", got)
			}
			for turn := 1; turn <= testCase.blockedTurns; turn++ {
				if err := fight.act(actor); err != nil {
					t.Fatalf("act: %v", err)
				}
				if got := actor.Blocked(breath); got != "recharging" {
					t.Errorf("Blocked on turn %d = %q, want recharging", turn, got)
				}
				if got := eventsOf[Recharged](events.events); len(got) != 0 {
					t.Errorf("Recharged on turn %d = %+v, want none", turn, got)
				}
			}
			if err := fight.act(actor); err != nil {
				t.Fatalf("act: %v", err)
			}
			if got := eventsOf[Recharged](events.events); !slices.Equal(got, []Recharged{{Combatant: "actor", Ability: "breath"}}) {
				t.Errorf("Recharged = %+v, want one at the start of the turn", got)
			}
			if got := actor.Blocked(breath); got != "" || actor.RechargeLeft("breath") != 0 {
				t.Errorf("Blocked, RechargeLeft = %q, %d, want none, 0", got, actor.RechargeLeft("breath"))
			}
		})
	}

	t.Run("a recharge step", func(t *testing.T) {
		fight, events, actor, _ := limitsFight(t)
		giveAbility(actor, limitedAbility("breath", 1, rechargeLimit()))
		source := effectSource{actor: actor, ability: "refresh"}
		step := creature.Step{Recharge: "breath"}

		fight.runStep(source, step, nil)
		if len(events.events) != 0 {
			t.Errorf("events for a ready ability = %+v, want none", events.events)
		}

		actor.recharging["breath"] = 3
		fight.runStep(source, step, nil)
		if got := eventsOf[Recharged](events.events); !slices.Equal(got, []Recharged{{Combatant: "actor", Ability: "breath", By: "refresh"}}) {
			t.Errorf("Recharged = %+v, want one by refresh", got)
		}
		if actor.RechargeLeft("breath") != 0 {
			t.Errorf("RechargeLeft = %d, want 0", actor.RechargeLeft("breath"))
		}
	})
}

func TestLimitsTickWhileDying(t *testing.T) {
	fight, events, actor, target := limitsFight(t, 2, 15, 5)
	fight.decider = &scriptedDecider{}
	breath := limitedAbility("breath", 1, rechargeLimit())
	giveAbility(actor, breath)
	if err := fight.perform(actor, Choice{Kind: ChooseAbility, ID: "breath", Target: target}); err != nil {
		t.Fatalf("perform: %v", err)
	}
	actor.HP, actor.Dying = 0, 2

	for turn := 1; turn <= 2; turn++ {
		if err := fight.act(actor); err != nil {
			t.Fatalf("act: %v", err)
		}
		if actor.Dying == 0 || actor.Dead {
			t.Fatalf("turn %d: Dying, Dead = %d, %v, want still dying", turn, actor.Dying, actor.Dead)
		}
	}
	fight.heal(effectSource{actor: target}, actor, 10)

	if actor.HP != 10 || actor.Dying != 0 {
		t.Fatalf("HP, Dying = %d, %d, want 10, 0", actor.HP, actor.Dying)
	}
	if got := actor.Blocked(breath); got != "" {
		t.Errorf("Blocked after revive = %q, want none", got)
	}
	if got := eventsOf[Recharged](events.events); !slices.Equal(got, []Recharged{{Combatant: "actor", Ability: "breath"}}) {
		t.Errorf("Recharged = %+v, want one while dying", got)
	}
}

func TestFlourishAndPress(t *testing.T) {
	t.Run("flourish", func(t *testing.T) {
		fight, _, actor, target := limitsFight(t)
		first := limitedAbility("first", 1, nil, "flourish")
		second := limitedAbility("second", 1, nil, "flourish")
		giveAbility(actor, first)
		giveAbility(actor, second)

		if err := fight.perform(actor, Choice{Kind: ChooseAbility, ID: "first", Target: target}); err != nil {
			t.Fatalf("perform first: %v", err)
		}
		if got := actor.Blocked(second); got != "flourish already used this round" {
			t.Errorf("Blocked = %q, want flourish already used this round", got)
		}
		wantRefused(t, fight.perform(actor, Choice{Kind: ChooseAbility, ID: "second", Target: target}), "flourish")

		fight.tickLimits(actor)

		if got := actor.Blocked(second); got != "" {
			t.Errorf("Blocked at the start of the next turn = %q, want none", got)
		}
	})
	t.Run("press", func(t *testing.T) {
		fight, _, actor, target := limitsFight(t, 1)
		follow := limitedAbility("follow", 1, nil, "press")
		giveAbility(actor, follow)

		if got := actor.Blocked(follow); got != "press needs an earlier attack this turn" {
			t.Errorf("Blocked before an attack = %q, want press needs an earlier attack this turn", got)
		}
		wantRefused(t, fight.perform(actor, Choice{Kind: ChooseAbility, ID: "follow", Target: target}), "press")
		if err := fight.perform(actor, Choice{Kind: ChooseStrike, ID: "fist", Target: target}); err != nil {
			t.Fatalf("perform fist: %v", err)
		}
		if got := actor.Blocked(follow); got != "" {
			t.Errorf("Blocked after an attack = %q, want none", got)
		}
	})
}

func TestFirstFittingSkipsBlocked(t *testing.T) {
	fight, _, actor, target := limitsFight(t)
	giveAbility(actor, limitedAbility("once", 1, &creature.Limit{PerFight: 1}))

	if choice, _ := (FirstFitting{}).Decide(fight, actor); choice.Kind != ChooseAbility || choice.ID != "once" {
		t.Fatalf("choice before = %+v, want the ability", choice)
	}
	actor.uses["once"] = 1

	want := Choice{Kind: ChooseStrike, ID: "fist", Target: target}
	if choice, acting := (FirstFitting{}).Decide(fight, actor); !acting || choice != want {
		t.Errorf("choice = %+v, %v, want %+v", choice, acting, want)
	}
}

func TestFreeAbilityDoesNotSpinTheTurn(t *testing.T) {
	hero := testCombatant("hero", SideParty, creature.KindPC, 500, 20)
	giveAbility(hero, creature.Ability{ID: "focus", Name: "Focus", Free: true, Limit: &creature.Limit{PerRound: 1}, Steps: []creature.Step{{Condition: "off-guard", Target: "self"}}})
	villain := testCombatant("villain", SideEnemies, creature.KindNPC, 500, 20)
	events := &recorder{}

	_, err := Run(Setup{
		Party:     []*creature.Creature{hero.Creature},
		Enemies:   []*creature.Creature{villain.Creature},
		Roller:    dice.NewSeeded(3),
		Observer:  events,
		Decider:   FirstFitting{},
		MaxRounds: 3,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	focusPerTurn := []int{}
	for _, event := range events.events {
		switch typed := event.(type) {
		case TurnStarted:
			if typed.Combatant == "hero" {
				focusPerTurn = append(focusPerTurn, 0)
			}
		case ActionUsed:
			if typed.Combatant == "hero" && typed.Action == "focus" {
				focusPerTurn[len(focusPerTurn)-1]++
			}
		}
	}
	if want := []int{1, 1, 1}; !slices.Equal(focusPerTurn, want) {
		t.Errorf("focus uses per hero turn = %v, want %v", focusPerTurn, want)
	}
}

func TestVariableCost(t *testing.T) {
	ram := creature.Ability{ID: "ram", Name: "Ram", Actions: 1, MaxActions: 3, Steps: []creature.Step{{Damage: []creature.DamagePart{
		{Dice: "1d6+1", Type: "force", PerAction: true, Roll: dice.Expression{Count: 1, Sides: 6, Modifier: 1}},
	}}}}
	heavy := creature.Ability{ID: "heavy", Name: "Heavy", Actions: 2, MaxActions: 3, Steps: []creature.Step{{Condition: "off-guard"}}}

	t.Run("three actions", func(t *testing.T) {
		fight, events, actor, target := limitsFight(t, 1, 2, 3)
		giveAbility(actor, ram)

		if err := fight.perform(actor, Choice{Kind: ChooseAbility, ID: "ram", Target: target, Actions: 3}); err != nil {
			t.Fatalf("perform: %v", err)
		}

		if dealt := lastDamage(t, events); dealt.Total != 9 {
			t.Errorf("damage = %d, want 9 (3d6+3 rolled as 1, 2, 3)", dealt.Total)
		}
		if used := eventsOf[ActionUsed](events.events); len(used) != 1 || used[0].Actions != 3 || actor.ActionsLeft != 0 {
			t.Errorf("ActionUsed, ActionsLeft = %+v, %d, want 3 actions spent, 0 left", used, actor.ActionsLeft)
		}
	})
	t.Run("four actions", func(t *testing.T) {
		fight, _, actor, target := limitsFight(t)
		giveAbility(actor, ram)
		actor.ActionsLeft = 4

		wantRefused(t, fight.perform(actor, Choice{Kind: ChooseAbility, ID: "ram", Target: target, Actions: 4}), "takes 1 to 3")
	})
	t.Run("below the minimum", func(t *testing.T) {
		fight, _, actor, target := limitsFight(t)
		giveAbility(actor, heavy)

		wantRefused(t, fight.perform(actor, Choice{Kind: ChooseAbility, ID: "heavy", Target: target, Actions: 1}), "takes 2 to 3")
		if actor.ActionsLeft != 3 {
			t.Errorf("ActionsLeft = %d, want 3", actor.ActionsLeft)
		}
	})
	t.Run("unset spends the minimum", func(t *testing.T) {
		fight, events, actor, target := limitsFight(t)
		giveAbility(actor, heavy)

		if err := fight.perform(actor, Choice{Kind: ChooseAbility, ID: "heavy", Target: target}); err != nil {
			t.Fatalf("perform: %v", err)
		}
		if used := eventsOf[ActionUsed](events.events); len(used) != 1 || used[0].Actions != 2 || actor.ActionsLeft != 1 {
			t.Errorf("ActionUsed, ActionsLeft = %+v, %d, want 2 actions spent, 1 left", used, actor.ActionsLeft)
		}
	})
}

func TestStrikeTweaks(t *testing.T) {
	maul := creature.Strike{ID: "maul", Name: "Maul", Bonus: 10, Damage: []creature.DamagePart{
		{Dice: "3d12", Type: "bludgeoning", Roll: dice.Expression{Count: 3, Sides: 12}},
	}}
	t.Run("extra dice on a hit and a critical hit", func(t *testing.T) {
		for _, testCase := range []struct {
			die       int
			wantTotal int
		}{
			{12, 15},
			{20, 30},
		} {
			attacker := testCombatant("attacker", SideParty, creature.KindPC, 50, 20)
			target := testCombatant("target", SideEnemies, creature.KindNPC, 500, 20)
			roller := script(t, testCase.die, 1, 2, 3, 4, 5)
			fight, events := testFight(roller, attacker, target)

			fight.strike(attacker, target, maul, strikeOptions{extraDice: 2})

			if dealt := lastDamage(t, events); dealt.Total != testCase.wantTotal {
				t.Errorf("die %d: damage = %d, want %d (5d12 rolled as 1 to 5)", testCase.die, dealt.Total, testCase.wantTotal)
			}
			if len(roller.values) != 0 {
				t.Errorf("die %d: unrolled values %v, want 5d12 to use them all", testCase.die, roller.values)
			}
		}
	})
	t.Run("extra splash damage lands on a miss", func(t *testing.T) {
		attacker := testCombatant("attacker", SideParty, creature.KindPC, 50, 20)
		target := testCombatant("target", SideEnemies, creature.KindNPC, 500, 20)
		fight, events := testFight(script(t, 5), attacker, target)

		fight.strike(attacker, target, fist(), strikeOptions{extraDamage: []creature.DamagePart{
			damagePart(t, "2d6", "fire"),
			categorizedPart(t, "3", "fire", "splash"),
		}})

		want := []DamageDealt{{Source: "attacker", Target: "target", Strike: "fist", Parts: []DamageAmount{{Type: "fire", Rolled: 3, Taken: 3}}, Total: 3, HPAfter: 497}}
		if got := eventsOf[DamageDealt](events.events); !reflect.DeepEqual(got, want) {
			t.Errorf("DamageDealt = %+v, want %+v", got, want)
		}
	})
	t.Run("counts as two attacks", func(t *testing.T) {
		attacker := testCombatant("attacker", SideParty, creature.KindPC, 50, 20)
		target := testCombatant("target", SideEnemies, creature.KindNPC, 500, 20)
		fight, events := testFight(script(t, 1, 1), attacker, target)
		fight.current = attacker

		fight.strike(attacker, target, fist(), strikeOptions{counts: 2})
		fight.strike(attacker, target, fist(), strikeOptions{})

		if penalties := strikePenalties(events); !slices.Equal(penalties, []int{0, -10}) || attacker.AttacksThisTurn != 3 {
			t.Errorf("penalties, AttacksThisTurn = %v, %d, want [0 -10], 3", penalties, attacker.AttacksThisTurn)
		}
	})
	t.Run("a shared penalty", func(t *testing.T) {
		attacker := testCombatant("attacker", SideParty, creature.KindPC, 50, 20)
		target := testCombatant("target", SideEnemies, creature.KindNPC, 500, 20)
		fight, events := testFight(script(t, 1, 1, 1), attacker, target)
		fight.current = attacker

		fight.strike(attacker, target, fist(), strikeOptions{})
		fight.strike(attacker, target, fist(), strikeOptions{penalty: creature.PenaltyShared})
		if attacker.AttacksThisTurn != 1 {
			t.Errorf("AttacksThisTurn after the shared Strike = %d, want 1", attacker.AttacksThisTurn)
		}
		fight.strike(attacker, target, fist(), strikeOptions{})

		if penalties := strikePenalties(events); !slices.Equal(penalties, []int{0, 0, -5}) {
			t.Errorf("penalties = %v, want [0 0 -5]", penalties)
		}
	})
	t.Run("not the attacker's turn", func(t *testing.T) {
		attacker := testCombatant("attacker", SideParty, creature.KindPC, 50, 20)
		target := testCombatant("target", SideEnemies, creature.KindNPC, 500, 20)
		fight, events := testFight(script(t, 1), attacker, target)
		fight.current = target
		attacker.AttacksThisTurn = 2

		fight.strike(attacker, target, fist(), strikeOptions{})

		if penalties := strikePenalties(events); !slices.Equal(penalties, []int{0}) || attacker.AttacksThisTurn != 2 {
			t.Errorf("penalties, AttacksThisTurn = %v, %d, want [0], 2", penalties, attacker.AttacksThisTurn)
		}
	})
}

func strikePenalties(events *recorder) []int {
	var penalties []int
	for _, rolled := range eventsOf[StrikeRolled](events.events) {
		penalties = append(penalties, rolled.Penalty)
	}
	return penalties
}

func TestSharedPenaltyStartsFreshEachTurn(t *testing.T) {
	actor := testCombatant("actor", SideParty, creature.KindPC, 50, 20)
	target := testCombatant("target", SideEnemies, creature.KindNPC, 500, 20)
	giveAbility(actor, creature.Ability{ID: "echo", Name: "Echo", Actions: 1, Steps: []creature.Step{{Strike: "fist", Penalty: creature.PenaltyShared}}})
	fight, events := testFight(script(t, 1, 1, 1, 1), actor, target)
	strike := Choice{Kind: ChooseStrike, ID: "fist", Target: target}
	fight.decider = &scriptedDecider{choices: []Choice{strike, strike, strike}}
	fight.current = actor
	if err := fight.act(actor); err != nil {
		t.Fatalf("first turn: %v", err)
	}
	fight.decider = &scriptedDecider{choices: []Choice{{Kind: ChooseAbility, ID: "echo", Target: target}}}

	if err := fight.act(actor); err != nil {
		t.Fatalf("second turn: %v", err)
	}

	if penalties := strikePenalties(events); !slices.Equal(penalties, []int{0, -5, -10, 0}) {
		t.Errorf("penalties = %v, want [0 -5 -10 0]", penalties)
	}
}

func TestSharedPenaltyAfterASpellAttack(t *testing.T) {
	actor := testCombatant("actor", SideParty, creature.KindPC, 50, 20)
	actor.Creature.Spellcasting = &creature.Spellcasting{Attack: 10, DC: 20}
	giveAbility(actor, creature.Ability{ID: "blast", Name: "Blast", Actions: 1, Steps: []creature.Step{
		{SpellAttack: true},
		{Strike: "fist", Penalty: creature.PenaltyShared},
	}})
	target := testCombatant("target", SideEnemies, creature.KindNPC, 500, 20)
	fight, events := testFight(script(t, 1, 1, 1), actor, target)
	fight.decider = &scriptedDecider{choices: []Choice{
		{Kind: ChooseStrike, ID: "fist", Target: target},
		{Kind: ChooseAbility, ID: "blast", Target: target},
	}}
	fight.current = actor

	if err := fight.act(actor); err != nil {
		t.Fatalf("act: %v", err)
	}

	spellAttacks := eventsOf[SpellAttackRolled](events.events)
	if len(spellAttacks) != 1 || spellAttacks[0].Penalty != -5 {
		t.Fatalf("SpellAttackRolled = %+v, want one at -5", spellAttacks)
	}
	if penalties := strikePenalties(events); !slices.Equal(penalties, []int{0, -5}) {
		t.Errorf("strike penalties = %v, want [0 -5]: the shared Strike takes the spell attack's penalty", penalties)
	}
}

func TestStrikeStepListsRun(t *testing.T) {
	finish := creature.Ability{ID: "finish", Name: "Finish", Actions: 1, Steps: []creature.Step{{
		Strike:   "fist",
		Outcomes: creature.Outcomes{Failure: []creature.Step{{Damage: []creature.DamagePart{damagePart(t, "1d4", "bludgeoning")}}}},
	}}}
	t.Run("the failure list on a miss", func(t *testing.T) {
		attacker := testCombatant("attacker", SideParty, creature.KindPC, 50, 20)
		target := testCombatant("target", SideEnemies, creature.KindNPC, 500, 20)
		fight, events := testFight(script(t, 5, 3), attacker, target)

		fight.useAbility(abilityUse{actor: attacker, ability: finish, target: target})

		dealt := eventsOf[DamageDealt](events.events)
		if len(dealt) != 1 || dealt[0].Total != 3 || dealt[0].Ability != "finish" {
			t.Errorf("DamageDealt = %+v, want 3 from finish", dealt)
		}
	})
	t.Run("not on a hit", func(t *testing.T) {
		attacker := testCombatant("attacker", SideParty, creature.KindPC, 50, 20)
		target := testCombatant("target", SideEnemies, creature.KindNPC, 500, 20)
		fight, events := testFight(script(t, 12, 4), attacker, target)

		fight.useAbility(abilityUse{actor: attacker, ability: finish, target: target})

		dealt := eventsOf[DamageDealt](events.events)
		if len(dealt) != 1 || dealt[0].Total != 6 {
			t.Errorf("DamageDealt = %+v, want only the fist's 6", dealt)
		}
	})
	t.Run("a recharge rider after a killing blow", func(t *testing.T) {
		attacker := testCombatant("attacker", SideEnemies, creature.KindNPC, 50, 20)
		giveAbility(attacker, limitedAbility("spores", 2, rechargeLimit()))
		attacker.recharging["spores"] = 2
		target := testCombatant("target", SideParty, creature.KindNPC, 5, 20)
		strike := fist()
		strike.CriticalSuccess = []creature.Step{{Recharge: "spores"}}
		fight, events := testFight(script(t, 20, 4), attacker, target)

		fight.strike(attacker, target, strike, strikeOptions{})

		if !target.Dead {
			t.Fatal("target is not dead")
		}
		if got := eventsOf[Recharged](events.events); !slices.Equal(got, []Recharged{{Combatant: "attacker", Ability: "spores", By: "fist"}}) {
			t.Errorf("Recharged = %+v, want spores by fist", got)
		}
		if attacker.RechargeLeft("spores") != 0 {
			t.Errorf("RechargeLeft = %d, want 0", attacker.RechargeLeft("spores"))
		}
	})
}

func TestExtraDiceSkipsFlatParts(t *testing.T) {
	t.Run("a flat part alone", func(t *testing.T) {
		attacker := testCombatant("attacker", SideParty, creature.KindPC, 50, 20)
		target := testCombatant("target", SideEnemies, creature.KindNPC, 500, 20)
		roller := script(t, 12)
		fight, events := testFight(roller, attacker, target)

		fight.strike(attacker, target, flatStrike(nil, damagePart(t, "5", "bludgeoning")), strikeOptions{extraDice: 2})

		if dealt := lastDamage(t, events); dealt.Total != 5 || len(roller.values) != 0 {
			t.Errorf("damage, unrolled = %d, %v, want 5 and no dice", dealt.Total, roller.values)
		}
	})
	t.Run("the die goes past the flat part", func(t *testing.T) {
		attacker := testCombatant("attacker", SideParty, creature.KindPC, 50, 20)
		target := testCombatant("target", SideEnemies, creature.KindNPC, 500, 20)
		roller := script(t, 12, 2, 3)
		fight, events := testFight(roller, attacker, target)

		fight.strike(attacker, target, flatStrike(nil, damagePart(t, "5", "bludgeoning"), damagePart(t, "1d6", "fire")), strikeOptions{extraDice: 1})

		if dealt := lastDamage(t, events); dealt.Total != 10 || len(roller.values) != 0 {
			t.Errorf("damage, unrolled = %d, %v, want 10 (5 + 2d6 rolled as 2, 3)", dealt.Total, roller.values)
		}
	})
}

func TestTurnEnds(t *testing.T) {
	fight, _, actor, target := limitsFight(t)
	giveAbility(actor, creature.Ability{ID: "last", Name: "Last", Actions: 1, TurnEnds: true, Steps: []creature.Step{{Condition: "off-guard"}}})

	if err := fight.perform(actor, Choice{Kind: ChooseAbility, ID: "last", Target: target}); err != nil {
		t.Fatalf("perform: %v", err)
	}

	if actor.ActionsLeft != 0 {
		t.Errorf("ActionsLeft = %d, want 0", actor.ActionsLeft)
	}

	t.Run("quickened", func(t *testing.T) {
		fight, _, actor, target := limitsFight(t)
		giveAbility(actor, creature.Ability{ID: "last", Name: "Last", Actions: 1, TurnEnds: true, Steps: []creature.Step{{Condition: "off-guard"}}})
		giveCondition(actor, "quickened", 0)
		fight.regainActions(actor)

		if err := fight.perform(actor, Choice{Kind: ChooseAbility, ID: "last", Target: target}); err != nil {
			t.Fatalf("perform: %v", err)
		}

		if actor.ActionsLeft != 0 || actor.OpenActions() != 0 {
			t.Errorf("ActionsLeft, OpenActions = %d, %d, want 0, 0", actor.ActionsLeft, actor.OpenActions())
		}
	})
}

func TestOtherFoe(t *testing.T) {
	twice := creature.Ability{ID: "twice", Name: "Twice", Actions: 2, Steps: []creature.Step{{Strike: "fist"}, {Strike: "fist", Target: creature.TargetOther}}}
	cases := []struct {
		name        string
		foes        []string
		wantTargets []string
	}{
		{"two foes", []string{"first", "second"}, []string{"first", "second"}},
		{"one foe", []string{"first"}, []string{"first", "first"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			actor := testCombatant("actor", SideParty, creature.KindPC, 50, 20)
			combatants := []*Combatant{actor}
			for _, id := range testCase.foes {
				combatants = append(combatants, testCombatant(id, SideEnemies, creature.KindNPC, 500, 20))
			}
			fight, events := testFight(script(t, 1, 1), combatants...)

			fight.useAbility(abilityUse{actor: actor, ability: twice, target: combatants[1]})

			var targets []string
			for _, rolled := range eventsOf[StrikeRolled](events.events) {
				targets = append(targets, rolled.Target)
			}
			if !slices.Equal(targets, testCase.wantTargets) {
				t.Errorf("targets = %v, want %v", targets, testCase.wantTargets)
			}
		})
	}
}

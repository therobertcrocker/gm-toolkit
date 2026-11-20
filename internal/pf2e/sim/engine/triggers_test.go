package engine

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

type recordingDecider struct {
	*scriptedDecider
	offered     [][]string
	nextTargets int
}

func (decider *recordingDecider) React(fight *Fight, actor *Combatant, event TriggerEvent, offered []string) (string, bool) {
	decider.offered = append(decider.offered, slices.Clone(offered))
	return decider.scriptedDecider.React(fight, actor, event, offered)
}

func (decider *recordingDecider) NextTarget(fight *Fight, actor *Combatant) (*Combatant, bool) {
	decider.nextTargets++
	return decider.scriptedDecider.NextTarget(fight, actor)
}

func triggeredAbility(id, on, cost string, steps ...creature.Step) creature.Ability {
	return creature.Ability{ID: id, Name: id, Trigger: &creature.Trigger{On: on, Cost: cost}, Steps: steps}
}

func offGuardStep() creature.Step {
	return creature.Step{Condition: "off-guard"}
}

func triggerFight(t *testing.T, reactions []string, values ...int) (*Fight, *recorder, *recordingDecider, *Combatant, *Combatant) {
	owner := testCombatant("owner", SideEnemies, creature.KindNPC, 100, 20)
	foe := testCombatant("foe", SideParty, creature.KindPC, 100, 20)
	fight, events := testFight(script(t, values...), owner, foe)
	decider := &recordingDecider{scriptedDecider: &scriptedDecider{reactions: reactions}}
	fight.decider = decider
	return fight, events, decider, owner, foe
}

func damagedBy(foe *Combatant) TriggerEvent {
	return TriggerEvent{Kind: creature.TriggerDamaged, By: foe}
}

func TestPassiveTriggerAlwaysFires(t *testing.T) {
	fight, events, decider, owner, foe := triggerFight(t, []string{"unused"})
	giveAbility(owner, triggeredAbility("spite", creature.TriggerDamaged, "", offGuardStep()))

	fight.fire(owner, damagedBy(foe))
	fight.fire(owner, damagedBy(foe))

	fired := Triggered{Combatant: "owner", Ability: "spite", On: "damaged", Cost: "", By: "foe"}
	if got := eventsOf[Triggered](events.events); !slices.Equal(got, []Triggered{fired, fired}) {
		t.Errorf("Triggered = %+v, want two %+v", got, fired)
	}
	if len(decider.offered) != 0 || len(decider.reactions) != 1 {
		t.Errorf("decider offered %v, reactions left %v, want never asked", decider.offered, decider.reactions)
	}
	if foe.Condition("off-guard") != 1 {
		t.Errorf("foe off-guard = %d, want 1", foe.Condition("off-guard"))
	}
}

func TestReactionAsksTheDecider(t *testing.T) {
	setup := func(t *testing.T, reactions []string) (*Fight, *recorder, *recordingDecider, *Combatant, *Combatant) {
		fight, events, decider, owner, foe := triggerFight(t, reactions)
		giveAbility(owner, triggeredAbility("b-react", creature.TriggerDamaged, creature.CostReaction, offGuardStep()))
		giveAbility(owner, triggeredAbility("a-react", creature.TriggerDamaged, creature.CostReaction, offGuardStep()))
		giveAbility(owner, triggeredAbility("c-free", creature.TriggerDamaged, creature.CostFree, offGuardStep()))
		giveAbility(owner, triggeredAbility("d-other-kind", creature.TriggerAttacked, creature.CostReaction, offGuardStep()))
		spent := triggeredAbility("e-spent", creature.TriggerDamaged, creature.CostReaction, offGuardStep())
		spent.Limit = &creature.Limit{PerFight: 1}
		giveAbility(owner, spent)
		owner.uses["e-spent"] = 1
		return fight, events, decider, owner, foe
	}

	t.Run("decline", func(t *testing.T) {
		fight, events, decider, owner, foe := setup(t, []string{""})
		fight.fire(owner, damagedBy(foe))
		if want := [][]string{{"a-react", "b-react", "c-free"}}; !reflect.DeepEqual(decider.offered, want) {
			t.Errorf("offered = %v, want %v", decider.offered, want)
		}
		if len(events.events) != 0 || owner.ReactionUsed || foe.Condition("off-guard") != 0 {
			t.Errorf("events, ReactionUsed, foe off-guard = %+v, %v, %d, want none, false, 0", events.events, owner.ReactionUsed, foe.Condition("off-guard"))
		}
	})
	t.Run("pick", func(t *testing.T) {
		fight, events, _, owner, foe := setup(t, []string{"b-react"})
		fight.fire(owner, damagedBy(foe))
		want := []Triggered{{Combatant: "owner", Ability: "b-react", On: "damaged", Cost: "reaction", By: "foe"}}
		if got := eventsOf[Triggered](events.events); !slices.Equal(got, want) {
			t.Errorf("Triggered = %+v, want %+v", got, want)
		}
		if !owner.ReactionUsed || foe.Condition("off-guard") != 1 {
			t.Errorf("ReactionUsed, foe off-guard = %v, %d, want true, 1", owner.ReactionUsed, foe.Condition("off-guard"))
		}
	})
}

func TestOneReactionPerRound(t *testing.T) {
	fight, events, decider, owner, foe := triggerFight(t, []string{"react", "react"})
	giveAbility(owner, triggeredAbility("react", creature.TriggerDamaged, creature.CostReaction, offGuardStep()))

	fight.fire(owner, damagedBy(foe))
	fight.fire(owner, damagedBy(foe))
	if len(decider.offered) != 1 || len(eventsOf[Triggered](events.events)) != 1 {
		t.Fatalf("offered %v, Triggered %d, want one offer and one reaction before the owner's turn", decider.offered, len(eventsOf[Triggered](events.events)))
	}

	if err := fight.act(owner); err != nil {
		t.Fatalf("act: %v", err)
	}
	fight.fire(owner, damagedBy(foe))
	if want := [][]string{{"react"}, {"react"}}; !reflect.DeepEqual(decider.offered, want) || len(eventsOf[Triggered](events.events)) != 2 {
		t.Errorf("offered %v, Triggered %d, want %v and 2 after the owner's turn", decider.offered, len(eventsOf[Triggered](events.events)), want)
	}
}

func TestExtraReaction(t *testing.T) {
	setup := func(t *testing.T, reactions []string, values ...int) (*Fight, *recordingDecider, *Combatant, *Combatant) {
		fight, _, decider, owner, foe := triggerFight(t, reactions, values...)
		owner.Creature.ExtraReactions = map[string]int{"aoo": 1}
		giveAbility(owner, triggeredAbility("aoo", creature.TriggerDamaged, creature.CostReaction, offGuardStep()))
		giveAbility(owner, triggeredAbility("other", creature.TriggerDamaged, creature.CostReaction, offGuardStep()))
		return fight, decider, owner, foe
	}

	t.Run("not there before the owner's first turn", func(t *testing.T) {
		fight, decider, owner, foe := setup(t, []string{"aoo", "aoo"})
		fight.fire(owner, damagedBy(foe))
		if !owner.ReactionUsed {
			t.Errorf("ReactionUsed = false, want the shared reaction spent")
		}
		fight.fire(owner, damagedBy(foe))
		if want := [][]string{{"aoo", "other"}}; !reflect.DeepEqual(decider.offered, want) {
			t.Errorf("offered = %v, want %v and nothing after", decider.offered, want)
		}
	})
	t.Run("the ability answers twice", func(t *testing.T) {
		fight, decider, owner, foe := setup(t, []string{"aoo", "aoo", "aoo"})
		if err := fight.act(owner); err != nil {
			t.Fatalf("act: %v", err)
		}
		fight.fire(owner, damagedBy(foe))
		if owner.ReactionUsed || owner.extraReactions["aoo"] != 0 {
			t.Errorf("after one: ReactionUsed, extra = %v, %d, want false, 0", owner.ReactionUsed, owner.extraReactions["aoo"])
		}
		fight.fire(owner, damagedBy(foe))
		if !owner.ReactionUsed {
			t.Errorf("after two: ReactionUsed = false, want true")
		}
		fight.fire(owner, damagedBy(foe))
		if want := [][]string{{"aoo", "other"}, {"aoo", "other"}}; !reflect.DeepEqual(decider.offered, want) {
			t.Errorf("offered = %v, want %v and nothing after", decider.offered, want)
		}
	})
	t.Run("another ability gets only the shared one", func(t *testing.T) {
		fight, decider, owner, foe := setup(t, []string{"other", "aoo", "aoo"})
		if err := fight.act(owner); err != nil {
			t.Fatalf("act: %v", err)
		}
		fight.fire(owner, damagedBy(foe))
		fight.fire(owner, damagedBy(foe))
		fight.fire(owner, damagedBy(foe))
		if want := [][]string{{"aoo", "other"}, {"aoo"}}; !reflect.DeepEqual(decider.offered, want) {
			t.Errorf("offered = %v, want %v", decider.offered, want)
		}
		if !owner.ReactionUsed || owner.extraReactions["aoo"] != 0 {
			t.Errorf("ReactionUsed, extra = %v, %d, want true, 0", owner.ReactionUsed, owner.extraReactions["aoo"])
		}
	})
	for _, dying := range []int{0, 1} {
		t.Run("a downed owner regains it at its turn", func(t *testing.T) {
			fight, _, owner, _ := setup(t, nil, 15)
			owner.Creature.Kind = creature.KindPC
			owner.HP, owner.Dying = 0, dying
			if err := fight.act(owner); err != nil {
				t.Fatalf("act: %v", err)
			}
			if owner.extraReactions["aoo"] != 1 {
				t.Errorf("dying %d: extra = %d, want 1", dying, owner.extraReactions["aoo"])
			}
		})
	}
}

func TestStunnedCannotReact(t *testing.T) {
	for _, stunned := range []int{1, 0} {
		fight, events, decider, owner, foe := triggerFight(t, []string{"react"})
		giveAbility(owner, triggeredAbility("react", creature.TriggerDamaged, creature.CostReaction, offGuardStep()))
		giveAbility(owner, triggeredAbility("passive", creature.TriggerDamaged, "", offGuardStep()))
		giveCondition(owner, "stunned", stunned)

		fight.fire(owner, damagedBy(foe))

		want := []Triggered{{Combatant: "owner", Ability: "passive", On: "damaged", By: "foe"}}
		if got := eventsOf[Triggered](events.events); !slices.Equal(got, want) {
			t.Errorf("stunned %d: Triggered = %+v, want %+v", stunned, got, want)
		}
		if len(decider.offered) != 0 || owner.ReactionUsed {
			t.Errorf("stunned %d: offered %v, ReactionUsed %v, want nothing offered", stunned, decider.offered, owner.ReactionUsed)
		}
	}
}

func TestDownedOwnerDoesNotFire(t *testing.T) {
	for _, state := range []string{"dying", "dead"} {
		t.Run(state, func(t *testing.T) {
			fight, events, decider, owner, foe := triggerFight(t, []string{"react"})
			giveAbility(owner, triggeredAbility("passive", creature.TriggerDamaged, "", offGuardStep()))
			giveAbility(owner, triggeredAbility("react", creature.TriggerDamaged, creature.CostReaction, offGuardStep()))
			owner.HP = 0
			if state == "dying" {
				owner.Dying = 1
			} else {
				owner.Dead = true
			}
			fight.fire(owner, damagedBy(foe))
			if len(events.events) != 0 || len(decider.offered) != 0 {
				t.Errorf("events %+v, offered %v, want nothing", events.events, decider.offered)
			}
		})
	}
	t.Run("own passive drops the owner", func(t *testing.T) {
		fight, events, decider, owner, foe := triggerFight(t, []string{"b-react"})
		hurt := creature.Step{Damage: []creature.DamagePart{damagePart(t, "200", "untyped")}, Target: creature.TargetSelf}
		giveAbility(owner, triggeredAbility("a-hurt", creature.TriggerDamaged, "", hurt))
		giveAbility(owner, triggeredAbility("b-react", creature.TriggerDamaged, creature.CostReaction, offGuardStep()))

		fight.fire(owner, damagedBy(foe))

		if !owner.Dead || len(decider.offered) != 0 {
			t.Errorf("Dead, offered = %v, %v, want true and nothing offered", owner.Dead, decider.offered)
		}
		if got := eventsOf[Triggered](events.events); len(got) != 1 || got[0].Ability != "a-hurt" {
			t.Errorf("Triggered = %+v, want a-hurt only", got)
		}
	})
}

func TestBlockedAbilityIsNotOffered(t *testing.T) {
	for _, cost := range []string{"", creature.CostReaction} {
		fight, events, decider, owner, foe := triggerFight(t, []string{"once", "once"})
		once := triggeredAbility("once", creature.TriggerDamaged, cost, offGuardStep())
		once.Limit = &creature.Limit{PerFight: 1}
		giveAbility(owner, once)

		fight.fire(owner, damagedBy(foe))
		owner.ReactionUsed = false
		fight.fire(owner, damagedBy(foe))

		if got := eventsOf[Triggered](events.events); len(got) != 1 {
			t.Errorf("cost %q: Triggered = %+v, want one", cost, got)
		}
		if got := eventsOf[UseSpent](events.events); !slices.Equal(got, []UseSpent{{Combatant: "owner", Ability: "once", Left: 0}}) {
			t.Errorf("cost %q: UseSpent = %+v, want one with none left", cost, got)
		}
		if cost == creature.CostReaction && len(decider.offered) != 1 {
			t.Errorf("offered = %v, want one offer", decider.offered)
		}
	}
}

func TestFreeTriggerNeedsNoReaction(t *testing.T) {
	fight, events, decider, owner, foe := triggerFight(t, []string{"free"})
	giveAbility(owner, triggeredAbility("free", creature.TriggerDamaged, creature.CostFree, offGuardStep()))
	owner.ReactionUsed = true

	fight.fire(owner, damagedBy(foe))

	if want := [][]string{{"free"}}; !reflect.DeepEqual(decider.offered, want) {
		t.Errorf("offered = %v, want %v", decider.offered, want)
	}
	want := []Triggered{{Combatant: "owner", Ability: "free", On: "damaged", Cost: "free", By: "foe"}}
	if got := eventsOf[Triggered](events.events); !slices.Equal(got, want) {
		t.Errorf("Triggered = %+v, want %+v", got, want)
	}
}

type unofferedReactor struct {
	FirstFitting
}

func (unofferedReactor) React(fight *Fight, actor *Combatant, event TriggerEvent, offered []string) (string, bool) {
	return "not-offered", true
}

func TestReactWithUnofferedAbilityFaults(t *testing.T) {
	owner := &creature.Creature{ID: "owner", Name: "owner", Kind: creature.KindNPC, Level: 1, Defense: creature.Defense{HP: 50, AC: 20}}
	owner.Abilities = map[string]creature.Ability{"react": triggeredAbility("react", creature.TriggerDamaged, creature.CostReaction, offGuardStep())}
	foe := &creature.Creature{ID: "foe", Name: "foe", Kind: creature.KindPC, Level: 1, Defense: creature.Defense{HP: 50, AC: 20}}
	fight, err := newFight(Setup{Party: []*creature.Creature{foe}, Enemies: []*creature.Creature{owner}, Roller: script(t, 10, 5), Observer: &recorder{}, Decider: unofferedReactor{}, MaxRounds: 3})
	if err != nil {
		t.Fatalf("newFight: %v", err)
	}
	fight.fire(fight.combatants[1], damagedBy(fight.combatants[0]))

	outcome, err := fight.run()

	if !errors.Is(err, ErrBadChoice) || !strings.Contains(err.Error(), `owner reacted with "not-offered", which was not offered`) || outcome != NoOutcome {
		t.Errorf("run = %v, %v, want NoOutcome and ErrBadChoice naming the ability", outcome, err)
	}
}

func TestTriggeredAbilityKeepsItsTarget(t *testing.T) {
	fight, events, decider, owner, foe := triggerFight(t, []string{"answer"})
	foe.Creature.Kind = creature.KindNPC
	second := testCombatant("second", SideParty, creature.KindPC, 100, 20)
	fight.combatants = append(fight.combatants, second)
	fight.order = append(fight.order, second)
	kill := creature.Step{Damage: []creature.DamagePart{damagePart(t, "200", "untyped")}}
	giveAbility(owner, triggeredAbility("answer", creature.TriggerDamaged, creature.CostReaction, kill, offGuardStep()))

	fight.fire(owner, damagedBy(foe))

	if !foe.Dead {
		t.Fatalf("foe Dead = false, want the first step to kill it")
	}
	if decider.nextTargets != 0 || second.Condition("off-guard") != 0 || len(eventsOf[ConditionChanged](events.events)) != 0 {
		t.Errorf("NextTarget calls, second off-guard, ConditionChanged = %d, %d, %+v, want 0, 0, none", decider.nextTargets, second.Condition("off-guard"), eventsOf[ConditionChanged](events.events))
	}
}

func TestTriggerWithNoCauseAimsAtAFoe(t *testing.T) {
	fight, events, _, owner, foe := triggerFight(t, nil)
	foe.Dead = true
	second := testCombatant("second", SideParty, creature.KindPC, 100, 20)
	fight.combatants = append(fight.combatants, second)
	hit := creature.Step{Damage: []creature.DamagePart{damagePart(t, "5", "untyped")}}
	self := creature.Step{Condition: "quickened", Target: creature.TargetSelf}
	giveAbility(owner, triggeredAbility("dawn", creature.TriggerTurnStart, "", hit, self))

	fight.fire(owner, TriggerEvent{Kind: creature.TriggerTurnStart})

	want := []Triggered{{Combatant: "owner", Ability: "dawn", On: "turn-start"}}
	if got := eventsOf[Triggered](events.events); !slices.Equal(got, want) {
		t.Errorf("Triggered = %+v, want %+v", got, want)
	}
	if got := damageTo(events, "second"); !slices.Equal(got, []int{5}) {
		t.Errorf("damage to second = %v, want [5]", got)
	}
	if owner.Condition("quickened") != 1 || second.Condition("quickened") != 0 {
		t.Errorf("quickened owner, second = %d, %d, want 1, 0", owner.Condition("quickened"), second.Condition("quickened"))
	}
}

func TestTriggeredDamageNamesItsCause(t *testing.T) {
	t.Run("damage step", func(t *testing.T) {
		fight, events, _, owner, foe := triggerFight(t, []string{"answer"})
		giveAbility(owner, triggeredAbility("answer", creature.TriggerDamaged, creature.CostReaction, creature.Step{Damage: []creature.DamagePart{damagePart(t, "5", "untyped")}}))

		fight.fire(owner, damagedBy(foe))

		want := []DamageDealt{{Source: "owner", Target: "foe", Ability: "answer", Parts: []DamageAmount{{Type: "untyped", Rolled: 5, Taken: 5}}, Total: 5, HPAfter: 95, TriggeredBy: "foe"}}
		if got := eventsOf[DamageDealt](events.events); !reflect.DeepEqual(got, want) {
			t.Errorf("DamageDealt = %+v, want %+v", got, want)
		}
	})
	t.Run("strike step", func(t *testing.T) {
		fight, events, _, owner, foe := triggerFight(t, []string{"answer"}, 15, 3)
		giveAbility(owner, triggeredAbility("answer", creature.TriggerDamaged, creature.CostReaction, creature.Step{Strike: "fist"}))

		fight.fire(owner, damagedBy(foe))

		want := []DamageDealt{{Source: "owner", Target: "foe", Strike: "fist", Ability: "answer", Parts: []DamageAmount{{Type: "bludgeoning", Rolled: 5, Taken: 5}}, Total: 5, HPAfter: 95, TriggeredBy: "foe"}}
		if got := eventsOf[DamageDealt](events.events); !reflect.DeepEqual(got, want) {
			t.Errorf("DamageDealt = %+v, want %+v", got, want)
		}
	})
	t.Run("heal step", func(t *testing.T) {
		fight, events, _, owner, foe := triggerFight(t, []string{"mend"})
		owner.HP = 50
		mend := creature.Step{Heal: "10", Target: creature.TargetSelf, HealRoll: damagePart(t, "10", "untyped").Roll}
		giveAbility(owner, triggeredAbility("mend", creature.TriggerDamaged, creature.CostReaction, mend))

		fight.fire(owner, damagedBy(foe))

		want := []Healed{{Source: "owner", Target: "owner", Ability: "mend", Amount: 10, HPAfter: 60, TriggeredBy: "foe"}}
		if got := eventsOf[Healed](events.events); !slices.Equal(got, want) {
			t.Errorf("Healed = %+v, want %+v", got, want)
		}
	})
}

func TestMatches(t *testing.T) {
	cases := []struct {
		name    string
		trigger creature.Trigger
		event   TriggerEvent
		hp      int
		want    bool
	}{
		{"other kind", creature.Trigger{On: "damaged"}, TriggerEvent{Kind: "saved"}, 100, false},
		{"targeted, any attack, by a spell", creature.Trigger{On: "targeted"}, TriggerEvent{Kind: "targeted"}, 100, true},
		{"targeted, strike, by a spell", creature.Trigger{On: "targeted", Attack: "strike"}, TriggerEvent{Kind: "targeted"}, 100, false},
		{"targeted, strike, by a ranged strike", creature.Trigger{On: "targeted", Attack: "strike"}, TriggerEvent{Kind: "targeted", IsStrike: true}, 100, true},
		{"targeted, melee strike, by a ranged strike", creature.Trigger{On: "targeted", Attack: "melee-strike"}, TriggerEvent{Kind: "targeted", IsStrike: true}, 100, false},
		{"targeted, melee strike, by a melee strike", creature.Trigger{On: "targeted", Attack: "melee-strike"}, TriggerEvent{Kind: "targeted", IsStrike: true, Melee: true}, 100, true},
		{"attacked, critical, on a critical", creature.Trigger{On: "attacked", Degrees: []string{"critical-success"}}, TriggerEvent{Kind: "attacked", Degree: CriticalSuccess}, 100, true},
		{"attacked, critical, on a hit", creature.Trigger{On: "attacked", Degrees: []string{"critical-success"}}, TriggerEvent{Kind: "attacked", Degree: Success}, 100, false},
		{"attacked, no degrees, on a miss", creature.Trigger{On: "attacked"}, TriggerEvent{Kind: "attacked", Degree: Failure}, 100, true},
		{"saved, damaging, no damage", creature.Trigger{On: "saved", Damaging: true}, TriggerEvent{Kind: "saved"}, 100, false},
		{"saved, damaging, damage", creature.Trigger{On: "saved", Damaging: true}, TriggerEvent{Kind: "saved", Damaging: true}, 100, true},
		{"saved, critical failure, on a failure", creature.Trigger{On: "saved", Degrees: []string{"critical-failure"}}, TriggerEvent{Kind: "saved", Degree: Failure}, 100, false},
		{"saved, critical failure, on a critical failure", creature.Trigger{On: "saved", Degrees: []string{"critical-failure"}}, TriggerEvent{Kind: "saved", Degree: CriticalFailure}, 100, true},
		{"damaged, melee, by a ranged hit", creature.Trigger{On: "damaged", Melee: true}, TriggerEvent{Kind: "damaged"}, 100, false},
		{"damaged, melee, by a melee hit", creature.Trigger{On: "damaged", Melee: true}, TriggerEvent{Kind: "damaged", Melee: true}, 100, true},
		{"damaged, any", creature.Trigger{On: "damaged"}, TriggerEvent{Kind: "damaged"}, 100, true},
		{"turn-start, threshold, below", creature.Trigger{On: "turn-start", HPAtMost: 75}, TriggerEvent{Kind: "turn-start"}, 74, true},
		{"turn-start, threshold, at", creature.Trigger{On: "turn-start", HPAtMost: 75}, TriggerEvent{Kind: "turn-start"}, 75, true},
		{"turn-start, threshold, above", creature.Trigger{On: "turn-start", HPAtMost: 75}, TriggerEvent{Kind: "turn-start"}, 76, false},
		{"turn-start, no threshold", creature.Trigger{On: "turn-start"}, TriggerEvent{Kind: "turn-start"}, 400, true},
		{"foe-acts, trait listed", creature.Trigger{On: "foe-acts", Traits: []string{"manipulate"}}, TriggerEvent{Kind: "foe-acts", Traits: []string{"attack", "manipulate"}}, 100, true},
		{"foe-acts, trait not listed", creature.Trigger{On: "foe-acts", Traits: []string{"manipulate"}}, TriggerEvent{Kind: "foe-acts", Traits: []string{"move"}}, 100, false},
		{"foe-acts, ranged, ranged strike", creature.Trigger{On: "foe-acts", Ranged: true}, TriggerEvent{Kind: "foe-acts", Ranged: true}, 100, true},
		{"foe-acts, ranged, other action", creature.Trigger{On: "foe-acts", Ranged: true}, TriggerEvent{Kind: "foe-acts", Traits: []string{"manipulate"}}, 100, false},
		{"foe-acts, traits only, ranged strike", creature.Trigger{On: "foe-acts", Traits: []string{"manipulate"}}, TriggerEvent{Kind: "foe-acts", Ranged: true}, 100, false},
		{"ally-gains, listed", creature.Trigger{On: "ally-gains", Conditions: []string{"frightened"}}, TriggerEvent{Kind: "ally-gains", Condition: "frightened"}, 100, true},
		{"ally-gains, not listed", creature.Trigger{On: "ally-gains", Conditions: []string{"frightened"}}, TriggerEvent{Kind: "ally-gains", Condition: "stupefied"}, 100, false},
		{"initiative", creature.Trigger{On: "initiative"}, TriggerEvent{Kind: "initiative"}, 100, true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			owner := testCombatant("owner", SideEnemies, creature.KindNPC, 400, 20)
			owner.HP = testCase.hp
			if got := matches(owner, testCase.trigger, testCase.event); got != testCase.want {
				t.Errorf("matches = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestFirstFittingSkipsTriggered(t *testing.T) {
	fight, _, _, owner, foe := triggerFight(t, nil)
	fight.current = owner
	owner.ActionsLeft = 3
	giveAbility(owner, triggeredAbility("a-triggered", creature.TriggerDamaged, "", offGuardStep()))

	choice, acting := FirstFitting{}.Decide(fight, owner)
	if !acting || choice.Kind != ChooseStrike || choice.ID != "fist" {
		t.Errorf("Decide = %+v, %v, want the fist Strike", choice, acting)
	}
	wantRefused(t, fight.perform(owner, Choice{Kind: ChooseAbility, ID: "a-triggered", Target: foe}), "no usable ability")
}

func firstIndex[E Event](events []Event) int {
	return slices.IndexFunc(events, func(event Event) bool {
		_, matches := event.(E)
		return matches
	})
}

func bow() creature.Strike {
	strike := fist()
	strike.ID, strike.Name, strike.Range = "bow", "Bow", 60
	return strike
}

func selfStep() creature.Step {
	return creature.Step{Condition: "quickened", Target: creature.TargetSelf}
}

func TestTargetedFiresBeforeTheRoll(t *testing.T) {
	type attackKind struct {
		name  string
		melee bool
		spell bool
	}
	melee, ranged, spell := attackKind{"melee strike", true, false}, attackKind{"ranged strike", false, false}, attackKind{"spell attack", false, true}
	cases := []struct {
		attack string
		kind   attackKind
		fires  bool
	}{
		{"", melee, true},
		{"", ranged, true},
		{"", spell, true},
		{creature.AttackMeleeStrike, melee, true},
		{creature.AttackMeleeStrike, ranged, false},
		{creature.AttackMeleeStrike, spell, false},
	}
	for _, testCase := range cases {
		t.Run(testCase.attack+" "+testCase.kind.name, func(t *testing.T) {
			fight, events, _, owner, foe := triggerFight(t, nil, 2)
			watch := triggeredAbility("watch", creature.TriggerTargeted, "", selfStep())
			watch.Trigger.Attack = testCase.attack
			giveAbility(owner, watch)
			foe.Creature.Spellcasting = &creature.Spellcasting{Attack: 10}

			rolledIndex := 0
			switch {
			case testCase.kind.spell:
				fight.spellAttack(effectSource{actor: foe, ability: "zap"}, creature.Step{SpellAttack: true}, owner)
				rolledIndex = firstIndex[SpellAttackRolled](events.events)
			case testCase.kind.melee:
				fight.strike(foe, owner, fist(), strikeOptions{})
				rolledIndex = firstIndex[StrikeRolled](events.events)
			default:
				fight.strike(foe, owner, bow(), strikeOptions{})
				rolledIndex = firstIndex[StrikeRolled](events.events)
			}

			triggered := eventsOf[Triggered](events.events)
			if !testCase.fires {
				if len(triggered) != 0 {
					t.Errorf("Triggered = %+v, want none", triggered)
				}
				return
			}
			want := Triggered{Combatant: "owner", Ability: "watch", On: "targeted", By: "foe"}
			if testCase.kind.spell {
				want.ByAbility = "zap"
			} else {
				want.ByStrike = map[bool]string{true: "fist", false: "bow"}[testCase.kind.melee]
			}
			if !slices.Equal(triggered, []Triggered{want}) {
				t.Errorf("Triggered = %+v, want %+v", triggered, want)
			}
			if index := firstIndex[Triggered](events.events); rolledIndex < 0 || index > rolledIndex {
				t.Errorf("Triggered at %d, roll at %d, want Triggered first", index, rolledIndex)
			}
		})
	}
}

func TestTargetedWaitsForConcealment(t *testing.T) {
	fight, events, _, owner, foe := triggerFight(t, nil, 4)
	giveAbility(owner, triggeredAbility("watch", creature.TriggerTargeted, "", selfStep()))
	giveCondition(owner, "concealed", 1)

	if _, rolled := fight.strike(foe, owner, fist(), strikeOptions{}); rolled {
		t.Fatalf("strike rolled, want the flat check to stop it")
	}
	if got := eventsOf[Triggered](events.events); len(got) != 0 {
		t.Errorf("Triggered = %+v, want none", got)
	}
	if want := []Event{FlatChecked{Combatant: "foe", Reason: "concealment", Die: 4, DC: 5}}; !reflect.DeepEqual(events.events, want) {
		t.Errorf("events = %+v, want %+v", events.events, want)
	}
}

func TestAttackedFiresAfterDamage(t *testing.T) {
	setup := func(t *testing.T, values ...int) (*Fight, *recorder, *recordingDecider, *Combatant, *Combatant) {
		fight, events, decider, owner, foe := triggerFight(t, []string{"spite"}, values...)
		spite := triggeredAbility("spite", creature.TriggerAttacked, creature.CostReaction, offGuardStep())
		spite.Trigger.Degrees = []string{"critical-success"}
		giveAbility(owner, spite)
		return fight, events, decider, owner, foe
	}
	rider := strikeOptions{lists: creature.Outcomes{CriticalSuccess: []creature.Step{{Condition: "sickened", Value: 1}}}}

	t.Run("critical hit", func(t *testing.T) {
		fight, events, _, owner, foe := setup(t, 20, 3)
		fight.strike(foe, owner, fist(), rider)

		want := []Event{
			StrikeRolled{Attacker: "foe", Target: "owner", Strike: "fist", Die: 20, Bonus: 10, Total: 30, AC: 20, Degree: CriticalSuccess},
			DamageDealt{Source: "foe", Target: "owner", Strike: "fist", Critical: true, Parts: []DamageAmount{{Type: "bludgeoning", Rolled: 10, Taken: 10}}, Total: 10, HPAfter: 90},
			ConditionChanged{Combatant: "owner", Condition: "sickened", From: 0, To: 1},
			Triggered{Combatant: "owner", Ability: "spite", On: "attacked", Cost: "reaction", By: "foe", ByStrike: "fist"},
			ConditionChanged{Combatant: "foe", Condition: "off-guard", From: 0, To: 1},
		}
		if !reflect.DeepEqual(events.events, want) {
			t.Errorf("events = %+v, want %+v", events.events, want)
		}
	})
	t.Run("hit", func(t *testing.T) {
		fight, events, decider, owner, foe := setup(t, 15, 3)
		fight.strike(foe, owner, fist(), rider)
		if got := eventsOf[Triggered](events.events); len(got) != 0 || len(decider.offered) != 0 {
			t.Errorf("Triggered %+v, offered %v, want none", got, decider.offered)
		}
	})
}

func TestSavedFiresForTheSaver(t *testing.T) {
	failureDamage := creature.Step{Save: "will", DC: 15, Outcomes: creature.Outcomes{Failure: []creature.Step{{Damage: []creature.DamagePart{damagePart(t, "5", "untyped")}}}}}
	cases := []struct {
		name     string
		degrees  []string
		damaging bool
		step     creature.Step
		die      int
		ally     bool
		fires    bool
	}{
		{name: "failure listed, saver fails", degrees: []string{"failure"}, step: creature.Step{Save: "will", DC: 15}, die: 10, fires: true},
		{name: "failure listed, saver succeeds", degrees: []string{"failure"}, step: creature.Step{Save: "will", DC: 15}, die: 16, fires: false},
		{name: "damaging, no damage anywhere", damaging: true, step: creature.Step{Save: "will", DC: 15}, die: 10, fires: false},
		{name: "damaging, damage in a degree list", damaging: true, step: failureDamage, die: 10, fires: true},
		{name: "against an ally's effect", step: creature.Step{Save: "will", DC: 15}, die: 10, ally: true, fires: false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			fight, events, _, owner, foe := triggerFight(t, nil, testCase.die)
			after := triggeredAbility("after", creature.TriggerSaved, "", selfStep())
			after.Trigger.Degrees, after.Trigger.Damaging = testCase.degrees, testCase.damaging
			giveAbility(owner, after)
			source := foe
			if testCase.ally {
				source = testCombatant("ally", SideEnemies, creature.KindNPC, 100, 20)
				fight.combatants = append(fight.combatants, source)
			}

			fight.saveStep(effectSource{actor: source, ability: "hex"}, testCase.step, []*Combatant{owner})

			var want []Triggered
			if testCase.fires {
				want = []Triggered{{Combatant: "owner", Ability: "after", On: "saved", By: "foe", ByAbility: "hex"}}
			}
			if got := eventsOf[Triggered](events.events); !slices.Equal(got, want) {
				t.Errorf("Triggered = %+v, want %+v", got, want)
			}
		})
	}
}

func TestDamagedFires(t *testing.T) {
	setup := func(t *testing.T, melee bool, values ...int) (*Fight, *recorder, *Combatant, *Combatant) {
		fight, events, _, owner, foe := triggerFight(t, nil, values...)
		hurt := triggeredAbility("hurt", creature.TriggerDamaged, "", selfStep())
		hurt.Trigger.Melee = melee
		giveAbility(owner, hurt)
		return fight, events, owner, foe
	}
	fired := func(events *recorder) int { return len(eventsOf[Triggered](events.events)) }

	t.Run("damage from a foe", func(t *testing.T) {
		fight, events, owner, foe := setup(t, false)
		fight.applyDamage(effectSource{actor: foe, ability: "zap"}, owner, rolled("untyped", 5), false)
		want := []Triggered{{Combatant: "owner", Ability: "hurt", On: "damaged", By: "foe", ByAbility: "zap"}}
		if got := eventsOf[Triggered](events.events); !slices.Equal(got, want) {
			t.Errorf("Triggered = %+v, want %+v", got, want)
		}
	})
	t.Run("damage from an ally", func(t *testing.T) {
		fight, events, owner, _ := setup(t, false)
		ally := testCombatant("ally", SideEnemies, creature.KindNPC, 100, 20)
		fight.applyDamage(effectSource{actor: ally, ability: "zap"}, owner, rolled("untyped", 5), false)
		if fired(events) != 0 {
			t.Errorf("Triggered = %d, want 0", fired(events))
		}
	})
	t.Run("0 damage after resistance", func(t *testing.T) {
		fight, events, owner, foe := setup(t, false)
		owner.Creature.Defense.Resist = map[string]creature.Adjustment{"all": {Value: 10}}
		fight.applyDamage(effectSource{actor: foe, ability: "zap"}, owner, rolled("untyped", 5), false)
		if got := lastDamage(t, events); got.Total != 0 || fired(events) != 0 {
			t.Errorf("Total, Triggered = %d, %d, want 0, 0", got.Total, fired(events))
		}
	})
	for _, melee := range []bool{false, true} {
		t.Run("a persistent tick set by a melee Strike", func(t *testing.T) {
			fight, events, owner, foe := setup(t, melee, 3, 2)
			owner.persistent = []*persistentInstance{{source: effectSource{actor: foe, strike: "fist", melee: true}, part: damagePart(t, "1d6", "fire")}}
			fight.takePersistentDamage(owner)
			if got := lastDamage(t, events); got.Total != 3 || got.Ability != "persistent" || fired(events) != 0 {
				t.Errorf("melee %v: Total, Ability, Triggered = %d, %q, %d, want 3, persistent, 0", melee, got.Total, got.Ability, fired(events))
			}
		})
	}
	t.Run("melee, a melee hit", func(t *testing.T) {
		fight, events, owner, foe := setup(t, true, 15, 3)
		fight.strike(foe, owner, fist(), strikeOptions{})
		if fired(events) != 1 {
			t.Errorf("Triggered = %d, want 1", fired(events))
		}
	})
	t.Run("melee, a ranged hit", func(t *testing.T) {
		fight, events, owner, foe := setup(t, true, 15, 3)
		fight.strike(foe, owner, bow(), strikeOptions{})
		if got := lastDamage(t, events); got.Total != 5 || fired(events) != 0 {
			t.Errorf("Total, Triggered = %d, %d, want 5, 0", got.Total, fired(events))
		}
	})
	t.Run("melee, a save's damage", func(t *testing.T) {
		fight, events, owner, foe := setup(t, true, 3, 10)
		basic := creature.Step{Save: "will", DC: 15, Basic: true, Damage: []creature.DamagePart{damagePart(t, "1d6", "fire")}}
		fight.saveStep(effectSource{actor: foe, ability: "blast"}, basic, []*Combatant{owner})
		if got := lastDamage(t, events); got.Total != 3 || fired(events) != 0 {
			t.Errorf("Total, Triggered = %d, %d, want 3, 0", got.Total, fired(events))
		}
	})
}

func TestReactionCanDropTheAttacker(t *testing.T) {
	fight, events, decider, owner, foe := triggerFight(t, []string{"punish"}, 2)
	foe.Creature.Kind = creature.KindNPC
	fight.current = foe
	flurry := creature.Ability{ID: "flurry", Name: "Flurry", Actions: 2, Steps: []creature.Step{{Strike: "fist"}, {Strike: "fist"}, {Strike: "fist"}}}
	giveAbility(foe, flurry)
	decider.choices = []Choice{{Kind: ChooseAbility, ID: "flurry", Target: owner}, {Kind: ChooseStrike, ID: "fist", Target: owner}}
	giveAbility(owner, triggeredAbility("punish", creature.TriggerAttacked, creature.CostReaction, creature.Step{Damage: []creature.DamagePart{damagePart(t, "200", "untyped")}}))

	if err := fight.act(foe); err != nil {
		t.Fatalf("act: %v", err)
	}

	if !foe.Dead {
		t.Fatalf("foe Dead = false, want the reaction to kill it")
	}
	if got := eventsOf[StrikeRolled](events.events); len(got) != 1 {
		t.Errorf("StrikeRolled = %+v, want one", got)
	}
	if len(decider.choices) != 1 || foe.ActionsLeft != 1 {
		t.Errorf("choices left, ActionsLeft = %d, %d, want 1, 1: the turn's loop ends", len(decider.choices), foe.ActionsLeft)
	}
}

func TestReactionsDoNotLoop(t *testing.T) {
	fight, events, decider, owner, foe := triggerFight(t, []string{"answer", "answer", "answer"})
	answer := triggeredAbility("answer", creature.TriggerDamaged, creature.CostReaction, creature.Step{Damage: []creature.DamagePart{damagePart(t, "5", "untyped")}})
	giveAbility(owner, answer)
	foe.Creature.Abilities = map[string]creature.Ability{"answer": answer}

	fight.applyDamage(effectSource{actor: foe, ability: "opening"}, owner, rolled("untyped", 5), false)

	if got := damageTo(events, "owner"); !slices.Equal(got, []int{5, 5}) {
		t.Errorf("owner damage = %v, want [5 5]", got)
	}
	if got := damageTo(events, "foe"); !slices.Equal(got, []int{5}) {
		t.Errorf("foe damage = %v, want [5]", got)
	}
	want := []Triggered{
		{Combatant: "owner", Ability: "answer", On: "damaged", Cost: "reaction", By: "foe", ByAbility: "opening"},
		{Combatant: "foe", Ability: "answer", On: "damaged", Cost: "reaction", By: "owner", ByAbility: "answer"},
	}
	if got := eventsOf[Triggered](events.events); !slices.Equal(got, want) {
		t.Errorf("Triggered = %+v, want %+v", got, want)
	}
	if len(decider.offered) != 2 || !owner.ReactionUsed || !foe.ReactionUsed {
		t.Errorf("offered, reactions used = %v, %v, %v, want two offers and both spent", decider.offered, owner.ReactionUsed, foe.ReactionUsed)
	}
}

func TestKnockedOutTargetCannotReact(t *testing.T) {
	fight, events, decider, owner, foe := triggerFight(t, []string{"spite"}, 15, 3)
	owner.Creature.Kind, owner.Side, foe.Side = creature.KindPC, SideParty, SideEnemies
	owner.HP = 5
	giveAbility(owner, triggeredAbility("spite", creature.TriggerAttacked, creature.CostReaction, offGuardStep()))
	giveAbility(owner, triggeredAbility("sting", creature.TriggerDamaged, creature.CostReaction, offGuardStep()))

	fight.strike(foe, owner, fist(), strikeOptions{})

	if owner.HP != 0 || owner.Dying != 1 {
		t.Fatalf("HP, Dying = %d, %d, want 0, 1", owner.HP, owner.Dying)
	}
	if got := eventsOf[Triggered](events.events); len(got) != 0 || len(decider.offered) != 0 {
		t.Errorf("Triggered %+v, offered %v, want none", got, decider.offered)
	}
}

func TestTurnStartTrigger(t *testing.T) {
	cases := []struct {
		hp          int
		wantActions int
	}{
		{75, 4},
		{74, 4},
		{76, 3},
	}
	for _, testCase := range cases {
		fight, events, _, owner, _ := triggerFight(t, nil)
		surge := triggeredAbility("surge", creature.TriggerTurnStart, "", creature.Step{GainActions: 1})
		surge.Trigger.HPAtMost = 75
		giveAbility(owner, surge)
		owner.HP = testCase.hp

		if err := fight.act(owner); err != nil {
			t.Fatalf("act: %v", err)
		}

		if owner.ActionsLeft != testCase.wantActions {
			t.Errorf("HP %d: ActionsLeft = %d, want %d", testCase.hp, owner.ActionsLeft, testCase.wantActions)
		}
		var want []Event
		if testCase.wantActions == 4 {
			want = []Event{
				TurnStarted{Combatant: "owner"},
				Triggered{Combatant: "owner", Ability: "surge", On: "turn-start"},
				ActionsGained{Combatant: "owner", Ability: "surge", Amount: 1},
			}
		} else {
			want = []Event{TurnStarted{Combatant: "owner"}}
		}
		if !reflect.DeepEqual(events.events, want) {
			t.Errorf("HP %d: events = %+v, want %+v", testCase.hp, events.events, want)
		}
	}
}

func TestFastHealing(t *testing.T) {
	cases := []struct {
		name    string
		kind    creature.Kind
		hp      int
		dying   int
		dead    bool
		values  []int
		want    []Healed
		wantEnd int
	}{
		{name: "below maximum", kind: creature.KindNPC, hp: 50, want: []Healed{{Source: "owner", Target: "owner", Ability: "fast-healing", Amount: 10, HPAfter: 60}}, wantEnd: 60},
		{name: "near maximum", kind: creature.KindNPC, hp: 95, want: []Healed{{Source: "owner", Target: "owner", Ability: "fast-healing", Amount: 5, HPAfter: 100}}, wantEnd: 100},
		{name: "at maximum", kind: creature.KindNPC, hp: 100, wantEnd: 100},
		{name: "stable at 0 HP", kind: creature.KindPC, hp: 0, want: []Healed{{Source: "owner", Target: "owner", Ability: "fast-healing", Amount: 10, HPAfter: 10}}, wantEnd: 10},
		{name: "dying at 0 HP", kind: creature.KindPC, hp: 0, dying: 1, want: []Healed{{Source: "owner", Target: "owner", Ability: "fast-healing", Amount: 10, HPAfter: 10}}, wantEnd: 10},
		{name: "dead", kind: creature.KindPC, hp: 0, dead: true, wantEnd: 0},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			fight, events, _, owner, _ := triggerFight(t, nil, testCase.values...)
			owner.Creature.Kind = testCase.kind
			owner.Creature.FastHealing = 10
			owner.HP, owner.Dying, owner.Dead = testCase.hp, testCase.dying, testCase.dead

			if err := fight.act(owner); err != nil {
				t.Fatalf("act: %v", err)
			}

			if got := eventsOf[Healed](events.events); !slices.Equal(got, testCase.want) || owner.HP != testCase.wantEnd {
				t.Errorf("Healed, HP = %+v, %d, want %+v, %d", got, owner.HP, testCase.want, testCase.wantEnd)
			}
		})
	}
}

func TestTurnStartBeforeFastHealing(t *testing.T) {
	fight, events, _, owner, _ := triggerFight(t, nil)
	owner.Creature.FastHealing = 10
	owner.HP = 70
	stand := triggeredAbility("last-stand", creature.TriggerTurnStart, "", selfStep())
	stand.Trigger.HPAtMost = 75
	giveAbility(owner, stand)

	if err := fight.act(owner); err != nil {
		t.Fatalf("act: %v", err)
	}

	want := []Event{
		TurnStarted{Combatant: "owner"},
		Triggered{Combatant: "owner", Ability: "last-stand", On: "turn-start"},
		ConditionChanged{Combatant: "owner", Condition: "quickened", From: 0, To: 1},
		Healed{Source: "owner", Target: "owner", Ability: "fast-healing", Amount: 10, HPAfter: 80},
	}
	if !reflect.DeepEqual(events.events, want) {
		t.Errorf("events = %+v, want %+v", events.events, want)
	}
}

func TestInitiativeTrigger(t *testing.T) {
	offGuardUntilMyTurnEnds := creature.Step{Condition: "off-guard", Duration: &creature.Duration{Until: "source-turn-end"}}
	first := &creature.Creature{ID: "first", Name: "first", Kind: creature.KindPC, Level: 1, Defense: creature.Defense{HP: 50, AC: 20}}
	first.Abilities = map[string]creature.Ability{"dawn": triggeredAbility("dawn", creature.TriggerInitiative, "", offGuardUntilMyTurnEnds)}
	second := &creature.Creature{ID: "second", Name: "second", Kind: creature.KindNPC, Level: 1, Defense: creature.Defense{HP: 50, AC: 20}}
	second.Abilities = map[string]creature.Ability{"dusk": triggeredAbility("dusk", creature.TriggerInitiative, "", offGuardUntilMyTurnEnds)}
	events := &recorder{}

	outcome, err := Run(Setup{Party: []*creature.Creature{first}, Enemies: []*creature.Creature{second}, Roller: script(t, 2, 20), Observer: events, Decider: FirstFitting{}, MaxRounds: 1})

	if err != nil || outcome != Stalemate {
		t.Fatalf("Run = %v, %v, want a stalemate", outcome, err)
	}
	want := []Event{
		InitiativeRolled{Combatant: "first", Die: 2, Total: 2},
		InitiativeRolled{Combatant: "second", Die: 20, Total: 20},
		Triggered{Combatant: "second", Ability: "dusk", On: "initiative"},
		ConditionChanged{Combatant: "first", Condition: "off-guard", From: 0, To: 1},
		Triggered{Combatant: "first", Ability: "dawn", On: "initiative"},
		ConditionChanged{Combatant: "second", Condition: "off-guard", From: 0, To: 1},
		RoundStarted{Round: 1},
		TurnStarted{Round: 1, Combatant: "second"},
		ConditionChanged{Combatant: "first", Condition: "off-guard", From: 1, To: 0},
		TurnStarted{Round: 1, Combatant: "first"},
		ConditionChanged{Combatant: "second", Condition: "off-guard", From: 1, To: 0},
		FightEnded{Rounds: 1, Outcome: Stalemate},
	}
	if !reflect.DeepEqual(events.events, want) {
		t.Errorf("events = %+v\nwant %+v", events.events, want)
	}
}

func foeActsFight(t *testing.T, trigger creature.Trigger, values ...int) (*Fight, *recorder, *Combatant, *Combatant) {
	fight, events, _, owner, foe := triggerFight(t, nil, values...)
	watch := triggeredAbility("watch", creature.TriggerFoeActs, "", selfStep())
	watch.Trigger.Traits, watch.Trigger.Ranged = trigger.Traits, trigger.Ranged
	giveAbility(owner, watch)
	fight.current = foe
	foe.ActionsLeft = 3
	return fight, events, owner, foe
}

func TestFoeActsOnTraits(t *testing.T) {
	manipulate := creature.Trigger{Traits: []string{"manipulate"}}
	t.Run("a manipulate ability", func(t *testing.T) {
		fight, events, owner, foe := foeActsFight(t, manipulate)
		giveAbility(foe, creature.Ability{ID: "brew", Name: "Brew", Actions: 1, Traits: []string{"alchemical", "manipulate"}, Steps: []creature.Step{offGuardStep()}})

		useScope(t, fight, foe, Choice{Kind: ChooseAbility, ID: "brew", Target: owner})

		want := []Event{
			ActionUsed{Combatant: "foe", Action: "brew", Actions: 1, Target: "owner"},
			Triggered{Combatant: "owner", Ability: "watch", On: "foe-acts", By: "foe", ByAbility: "brew"},
			ConditionChanged{Combatant: "owner", Condition: "quickened", From: 0, To: 1},
			ConditionChanged{Combatant: "owner", Condition: "off-guard", From: 0, To: 1},
		}
		if !reflect.DeepEqual(events.events, want) {
			t.Errorf("events = %+v, want %+v", events.events, want)
		}
	})
	t.Run("an ability with no listed trait", func(t *testing.T) {
		fight, events, owner, foe := foeActsFight(t, manipulate)
		giveAbility(foe, creature.Ability{ID: "focus", Name: "Focus", Actions: 1, Traits: []string{"concentrate"}, Steps: []creature.Step{offGuardStep()}})
		useScope(t, fight, foe, Choice{Kind: ChooseAbility, ID: "focus", Target: owner})
		if got := eventsOf[Triggered](events.events); len(got) != 0 {
			t.Errorf("Triggered = %+v, want none", got)
		}
	})
	t.Run("a ranged Strike", func(t *testing.T) {
		fight, events, owner, foe := foeActsFight(t, creature.Trigger{Ranged: true}, 2)
		foe.Creature.Strikes["bow"] = bow()
		useScope(t, fight, foe, Choice{Kind: ChooseStrike, ID: "bow", Target: owner})
		want := []Triggered{{Combatant: "owner", Ability: "watch", On: "foe-acts", By: "foe", ByStrike: "bow"}}
		if got := eventsOf[Triggered](events.events); !slices.Equal(got, want) {
			t.Errorf("Triggered = %+v, want %+v", got, want)
		}
		if index := firstIndex[Triggered](events.events); index > firstIndex[StrikeRolled](events.events) {
			t.Errorf("Triggered at %d, after StrikeRolled, want before", index)
		}
	})
	t.Run("a melee Strike", func(t *testing.T) {
		fight, events, owner, foe := foeActsFight(t, creature.Trigger{Ranged: true}, 2)
		useScope(t, fight, foe, Choice{Kind: ChooseStrike, ID: "fist", Target: owner})
		if got := eventsOf[Triggered](events.events); len(got) != 0 {
			t.Errorf("Triggered = %+v, want none", got)
		}
	})
	t.Run("a Stand", func(t *testing.T) {
		fight, events, _, foe := foeActsFight(t, creature.Trigger{Traits: []string{"move"}})
		giveCondition(foe, "prone", 0)
		useScope(t, fight, foe, Choice{Kind: ChooseStand})
		want := []Triggered{{Combatant: "owner", Ability: "watch", On: "foe-acts", By: "foe"}}
		if got := eventsOf[Triggered](events.events); !slices.Equal(got, want) {
			t.Errorf("Triggered = %+v, want %+v", got, want)
		}
		if foe.Condition("prone") != 0 || len(eventsOf[StoodUp](events.events)) != 1 {
			t.Errorf("prone, StoodUp = %d, %d, want 0, 1", foe.Condition("prone"), len(eventsOf[StoodUp](events.events)))
		}
	})
}

func opportunity(strikeStep creature.Step) creature.Ability {
	ability := triggeredAbility("aoo", creature.TriggerFoeActs, creature.CostReaction, strikeStep)
	ability.Trigger.Traits, ability.Trigger.Ranged = []string{"manipulate", "move"}, true
	return ability
}

func disruptFight(t *testing.T, values ...int) (*Fight, *recorder, *Combatant, *Combatant) {
	fight, events, _, owner, foe := triggerFight(t, []string{"aoo"}, values...)
	giveAbility(owner, opportunity(creature.Step{Strike: "fist", Disrupts: []string{"manipulate"}}))
	limit := rechargeLimit()
	limit.PerFight = 1
	giveAbility(foe, creature.Ability{ID: "brew", Name: "Brew", Actions: 1, Traits: []string{"manipulate"}, Limit: limit, Steps: []creature.Step{offGuardStep()}})
	fight.current = foe
	foe.ActionsLeft = 3
	return fight, events, owner, foe
}

func TestDisrupt(t *testing.T) {
	t.Run("a critical hit disrupts", func(t *testing.T) {
		fight, events, owner, foe := disruptFight(t, 2, 20, 3)

		useScope(t, fight, foe, Choice{Kind: ChooseAbility, ID: "brew", Target: owner})

		want := []Event{
			ActionUsed{Combatant: "foe", Action: "brew", Actions: 1, Target: "owner"},
			UseSpent{Combatant: "foe", Ability: "brew", Left: 0},
			RechargeRolled{Combatant: "foe", Ability: "brew", Rounds: 2},
			Triggered{Combatant: "owner", Ability: "aoo", On: "foe-acts", Cost: "reaction", By: "foe", ByAbility: "brew"},
			StrikeRolled{Attacker: "owner", Target: "foe", Strike: "fist", Die: 20, Bonus: 10, Total: 30, AC: 20, Degree: CriticalSuccess},
			ActionDisrupted{Combatant: "foe", Action: "brew", By: "owner", Ability: "aoo"},
			DamageDealt{Source: "owner", Target: "foe", Strike: "fist", Ability: "aoo", Critical: true, Parts: []DamageAmount{{Type: "bludgeoning", Rolled: 10, Taken: 10}}, Total: 10, HPAfter: 90, TriggeredBy: "foe"},
		}
		if !reflect.DeepEqual(events.events, want) {
			t.Errorf("events = %+v\nwant %+v", events.events, want)
		}
		if foe.ActionsLeft != 2 || owner.Condition("off-guard") != 0 {
			t.Errorf("ActionsLeft, owner off-guard = %d, %d, want 2, 0", foe.ActionsLeft, owner.Condition("off-guard"))
		}
	})
	t.Run("a hit does not", func(t *testing.T) {
		fight, events, owner, foe := disruptFight(t, 2, 15, 3)
		useScope(t, fight, foe, Choice{Kind: ChooseAbility, ID: "brew", Target: owner})
		if got := eventsOf[ActionDisrupted](events.events); len(got) != 0 || owner.Condition("off-guard") != 1 {
			t.Errorf("ActionDisrupted, owner off-guard = %+v, %d, want none, 1", got, owner.Condition("off-guard"))
		}
	})
	t.Run("a critical hit against a Stand does not", func(t *testing.T) {
		fight, events, _, foe := disruptFight(t, 20, 3)
		giveCondition(foe, "prone", 0)
		useScope(t, fight, foe, Choice{Kind: ChooseStand})
		if got := eventsOf[ActionDisrupted](events.events); len(got) != 0 || foe.Condition("prone") != 0 {
			t.Errorf("ActionDisrupted, prone = %+v, %d, want none, 0", got, foe.Condition("prone"))
		}
		if got := eventsOf[StrikeRolled](events.events); len(got) != 1 || got[0].Degree != CriticalSuccess {
			t.Errorf("StrikeRolled = %+v, want one critical hit", got)
		}
	})
}

func TestProvokeCanDropTheActor(t *testing.T) {
	fight, events, decider, owner, foe := triggerFight(t, []string{"aoo"})
	foe.Creature.Kind = creature.KindNPC
	fight.current = foe
	giveAbility(owner, opportunity(creature.Step{Damage: []creature.DamagePart{damagePart(t, "200", "untyped")}}))
	giveAbility(foe, creature.Ability{ID: "brew", Name: "Brew", Actions: 1, Traits: []string{"manipulate"}, Steps: []creature.Step{offGuardStep()}})
	decider.choices = []Choice{{Kind: ChooseAbility, ID: "brew", Target: owner}, {Kind: ChooseStrike, ID: "fist", Target: owner}}

	if err := fight.act(foe); err != nil {
		t.Fatalf("act: %v", err)
	}

	if !foe.Dead || owner.Condition("off-guard") != 0 {
		t.Errorf("Dead, owner off-guard = %v, %d, want true, 0: the ability does not run", foe.Dead, owner.Condition("off-guard"))
	}
	if len(decider.choices) != 1 || foe.ActionsLeft != 2 || len(eventsOf[StrikeRolled](events.events)) != 0 {
		t.Errorf("choices left, ActionsLeft, StrikeRolled = %d, %d, %d, want 1, 2, 0: the turn ends", len(decider.choices), foe.ActionsLeft, len(eventsOf[StrikeRolled](events.events)))
	}
}

func TestReactionStrikeHasNoPenalty(t *testing.T) {
	fight, events, _, owner, foe := triggerFight(t, []string{"aoo"}, 2, 2, 2, 2)
	giveAbility(owner, opportunity(creature.Step{Strike: "fist"}))
	giveAbility(foe, creature.Ability{ID: "brew", Name: "Brew", Actions: 1, Traits: []string{"manipulate"}, Steps: []creature.Step{offGuardStep()}})
	fight.current = owner
	for range 3 {
		fight.strike(owner, foe, fist(), strikeOptions{})
	}
	fight.current = foe
	foe.ActionsLeft = 3

	useScope(t, fight, foe, Choice{Kind: ChooseAbility, ID: "brew", Target: owner})

	if penalties := strikePenalties(events); !slices.Equal(penalties, []int{0, -5, -10, 0}) {
		t.Errorf("penalties = %v, want [0 -5 -10 0]", penalties)
	}
	if owner.AttacksThisTurn != 3 {
		t.Errorf("AttacksThisTurn = %d, want 3", owner.AttacksThisTurn)
	}
}

func allyGainsFight(t *testing.T) (*Fight, *recorder, *Combatant, *Combatant, *Combatant) {
	fight, events, _, owner, foe := triggerFight(t, []string{"comfort"})
	ally := testCombatant("ally", SideEnemies, creature.KindNPC, 100, 20)
	fight.combatants = append(fight.combatants, ally)
	comfort := triggeredAbility("comfort", creature.TriggerAllyGains, creature.CostReaction,
		creature.Step{Condition: "frightened", Reduce: 1},
		creature.Step{Condition: "frightened", Value: 1, Stack: true, Target: creature.TargetSelf})
	comfort.Trigger.Conditions = []string{"frightened"}
	giveAbility(owner, comfort)
	return fight, events, owner, foe, ally
}

func TestAllyGainsTrigger(t *testing.T) {
	frightened := func(value int) creature.Step { return creature.Step{Condition: "frightened", Value: value} }
	t.Run("an ally's value rises", func(t *testing.T) {
		for _, testCase := range []struct{ value, wantAlly int }{{2, 1}, {1, 0}} {
			fight, events, owner, foe, ally := allyGainsFight(t)
			fight.applyCondition(foe, ally, frightened(testCase.value))
			want := []Triggered{{Combatant: "owner", Ability: "comfort", On: "ally-gains", Cost: "reaction", By: "ally"}}
			if got := eventsOf[Triggered](events.events); !slices.Equal(got, want) {
				t.Errorf("frightened %d: Triggered = %+v, want %+v", testCase.value, got, want)
			}
			if ally.Condition("frightened") != testCase.wantAlly || owner.Condition("frightened") != 1 {
				t.Errorf("frightened %d: ally, owner = %d, %d, want %d, 1", testCase.value, ally.Condition("frightened"), owner.Condition("frightened"), testCase.wantAlly)
			}
			if testCase.wantAlly == 0 && len(ally.conditions) != 0 {
				t.Errorf("ally conditions = %+v, want frightened removed", ally.conditions)
			}
		}
	})
	t.Run("the same value again", func(t *testing.T) {
		fight, events, _, foe, ally := allyGainsFight(t)
		giveCondition(ally, "frightened", 2)
		fight.applyCondition(foe, ally, frightened(2))
		if got := eventsOf[Triggered](events.events); len(got) != 0 {
			t.Errorf("Triggered = %+v, want none", got)
		}
	})
	t.Run("the bearer itself", func(t *testing.T) {
		fight, events, owner, foe, _ := allyGainsFight(t)
		fight.applyCondition(foe, owner, frightened(2))
		if got := eventsOf[Triggered](events.events); len(got) != 0 || owner.Condition("frightened") != 2 {
			t.Errorf("Triggered, owner frightened = %+v, %d, want none, 2", got, owner.Condition("frightened"))
		}
	})
	t.Run("a foe", func(t *testing.T) {
		fight, events, owner, foe, _ := allyGainsFight(t)
		fight.applyCondition(owner, foe, frightened(2))
		if got := eventsOf[Triggered](events.events); len(got) != 0 || foe.Condition("frightened") != 2 {
			t.Errorf("Triggered, foe frightened = %+v, %d, want none, 2", got, foe.Condition("frightened"))
		}
	})
}

func TestStoodUpTrigger(t *testing.T) {
	fight, events, _, owner, foe := triggerFight(t, nil)
	owner.Creature.Kind, owner.Side, foe.Side = creature.KindPC, SideParty, SideEnemies
	giveAbility(owner, triggeredAbility("rise", creature.TriggerStoodUp, "", selfStep()))
	stand := func() {
		t.Helper()
		owner.ActionsLeft = 3
		useScope(t, fight, owner, Choice{Kind: ChooseStand})
	}

	fight.applyCondition(foe, owner, creature.Step{Condition: "prone"})
	stand()
	if got := eventsOf[Triggered](events.events); len(got) != 0 {
		t.Fatalf("ordinary prone: Triggered = %+v, want none", got)
	}

	fight.applyDamage(effectSource{actor: foe, ability: "maul"}, owner, rolled("untyped", 200), false)
	fight.heal(effectSource{actor: foe, ability: "mend"}, owner, 10)
	if owner.Condition("prone") != 1 || !owner.Standing() {
		t.Fatalf("prone, standing = %d, %v, want 1, true after the knock-out and heal", owner.Condition("prone"), owner.Standing())
	}
	stand()
	want := []Triggered{{Combatant: "owner", Ability: "rise", On: "stood-up"}}
	if got := eventsOf[Triggered](events.events); !slices.Equal(got, want) {
		t.Errorf("after the knock-out: Triggered = %+v, want %+v", got, want)
	}

	fight.applyCondition(foe, owner, creature.Step{Condition: "prone"})
	stand()
	if got := eventsOf[Triggered](events.events); !slices.Equal(got, want) {
		t.Errorf("second Stand: Triggered = %+v, want still %+v", got, want)
	}
}

func downedHealerFight(t *testing.T, dying int) (*Fight, *recorder, *Combatant, *scriptedRoller) {
	roller := script(t, 2, 2)
	owner := testCombatant("owner", SideParty, creature.KindPC, 100, 20)
	foe := testCombatant("foe", SideEnemies, creature.KindNPC, 100, 20)
	fight, events := testFight(roller, owner, foe)
	fight.current = owner
	owner.Creature.FastHealing = 10
	owner.HP, owner.Dying = 0, dying
	giveCondition(owner, "prone", 0)
	return fight, events, owner, roller
}

func TestFastHealingRevivesTheDowned(t *testing.T) {
	t.Run("dying", func(t *testing.T) {
		fight, events, owner, roller := downedHealerFight(t, 2)

		if err := fight.act(owner); err != nil {
			t.Fatalf("act: %v", err)
		}

		want := []Event{
			TurnStarted{Combatant: "owner"},
			Healed{Source: "owner", Target: "owner", Ability: "fast-healing", Amount: 10, HPAfter: 10},
			Revived{Combatant: "owner", DyingFrom: 2, Wounded: 1},
			ActionUsed{Combatant: "owner", Action: "stand", Actions: 1},
			ConditionChanged{Combatant: "owner", Condition: "off-guard", From: 1, To: 0},
			ConditionChanged{Combatant: "owner", Condition: "prone", From: 1, To: 0},
			StoodUp{Combatant: "owner"},
			ActionUsed{Combatant: "owner", Action: "fist", Actions: 1, Target: "foe"},
			StrikeRolled{Attacker: "owner", Target: "foe", Strike: "fist", Die: 2, Bonus: 10, Total: 12, AC: 20, Degree: Failure},
			ActionUsed{Combatant: "owner", Action: "fist", Actions: 1, Target: "foe"},
			StrikeRolled{Attacker: "owner", Target: "foe", Strike: "fist", Die: 2, Bonus: 10, Penalty: -5, Total: 7, AC: 20, Degree: CriticalFailure},
		}
		if !reflect.DeepEqual(events.events, want) {
			t.Errorf("events = %+v\nwant %+v", events.events, want)
		}
		if len(roller.values) != 0 || owner.HP != 10 || owner.Dying != 0 || owner.Wounded != 1 {
			t.Errorf("unrolled, HP, Dying, Wounded = %v, %d, %d, %d, want none, 10, 0, 1", roller.values, owner.HP, owner.Dying, owner.Wounded)
		}
	})
	t.Run("stable", func(t *testing.T) {
		fight, events, owner, roller := downedHealerFight(t, 0)

		if err := fight.act(owner); err != nil {
			t.Fatalf("act: %v", err)
		}

		if got := eventsOf[Healed](events.events); !slices.Equal(got, []Healed{{Source: "owner", Target: "owner", Ability: "fast-healing", Amount: 10, HPAfter: 10}}) {
			t.Errorf("Healed = %+v, want one heal to 10", got)
		}
		if len(eventsOf[Revived](events.events)) != 0 || len(eventsOf[StoodUp](events.events)) != 1 || len(eventsOf[StrikeRolled](events.events)) != 2 {
			t.Errorf("Revived, StoodUp, StrikeRolled = %d, %d, %d, want 0, 1, 2", len(eventsOf[Revived](events.events)), len(eventsOf[StoodUp](events.events)), len(eventsOf[StrikeRolled](events.events)))
		}
		if len(roller.values) != 0 || owner.ActionsLeft != 0 {
			t.Errorf("unrolled, ActionsLeft = %v, %d, want none, 0", roller.values, owner.ActionsLeft)
		}
	})
	t.Run("dead", func(t *testing.T) {
		fight, events, owner, _ := downedHealerFight(t, 0)
		owner.Dead = true
		if err := fight.act(owner); err != nil {
			t.Fatalf("act: %v", err)
		}
		if len(events.events) != 0 || owner.HP != 0 {
			t.Errorf("events, HP = %+v, %d, want none, 0", events.events, owner.HP)
		}
	})
}

func loopCreature(id string, kind creature.Kind, abilities ...creature.Ability) *creature.Creature {
	source := &creature.Creature{ID: id, Name: id, Kind: kind, Level: 1, Defense: creature.Defense{HP: 100, AC: 20}, Abilities: map[string]creature.Ability{}}
	for _, ability := range abilities {
		source.Abilities[ability.ID] = ability
	}
	return source
}

func TestTriggerLoopEndsTheFight(t *testing.T) {
	t.Run("ally-gains stacking on the triggering ally", func(t *testing.T) {
		soothe := triggeredAbility("soothe", creature.TriggerAllyGains, "", creature.Step{Condition: "frightened", Value: 1, Stack: true})
		soothe.Trigger.Conditions = []string{"frightened"}
		scare := creature.Ability{ID: "scare", Name: "Scare", Actions: 1, Steps: []creature.Step{{Condition: "frightened", Value: 1}}}
		party := []*creature.Creature{loopCreature("a", creature.KindPC), loopCreature("b", creature.KindPC, soothe)}
		enemies := []*creature.Creature{loopCreature("e", creature.KindNPC, scare)}

		outcome, err := Run(Setup{Party: party, Enemies: enemies, Roller: script(t, 1, 1, 20), Observer: &recorder{}, Decider: FirstFitting{}, MaxRounds: 1})

		if !errors.Is(err, ErrTriggerLoop) || err.Error() != "trigger loop: b: ally-gains nested more than 16 deep" || outcome != NoOutcome {
			t.Errorf("Run = %v, %v, want NoOutcome and %q", outcome, err, "trigger loop: b: ally-gains nested more than 16 deep")
		}
	})
	t.Run("passive damage back and forth", func(t *testing.T) {
		spite := triggeredAbility("spite", creature.TriggerDamaged, "", creature.Step{Damage: []creature.DamagePart{damagePart(t, "1", "untyped")}})
		poke := creature.Ability{ID: "poke", Name: "Poke", Actions: 1, Steps: []creature.Step{{Damage: []creature.DamagePart{damagePart(t, "1", "untyped")}}}}
		party := []*creature.Creature{loopCreature("a", creature.KindPC, spite)}
		enemies := []*creature.Creature{loopCreature("e", creature.KindNPC, spite, poke)}

		outcome, err := Run(Setup{Party: party, Enemies: enemies, Roller: script(t, 1, 20), Observer: &recorder{}, Decider: FirstFitting{}, MaxRounds: 1})

		if !errors.Is(err, ErrTriggerLoop) || err.Error() != "trigger loop: a: damaged nested more than 16 deep" || outcome != NoOutcome {
			t.Errorf("Run = %v, %v, want NoOutcome and %q", outcome, err, "trigger loop: a: damaged nested more than 16 deep")
		}
	})
}

func TestFirstFaultIsKept(t *testing.T) {
	owner := loopCreature("owner", creature.KindNPC, triggeredAbility("react", creature.TriggerDamaged, creature.CostReaction, offGuardStep()))
	foe := loopCreature("foe", creature.KindPC)
	fight, err := newFight(Setup{Party: []*creature.Creature{foe}, Enemies: []*creature.Creature{owner}, Roller: script(t, 10, 5), Observer: &recorder{}, Decider: &scriptedDecider{reactions: []string{"first-bad", "second-bad"}}, MaxRounds: 3})
	if err != nil {
		t.Fatalf("newFight: %v", err)
	}
	fight.fire(fight.combatants[1], damagedBy(fight.combatants[0]))
	fight.fire(fight.combatants[1], damagedBy(fight.combatants[0]))

	outcome, err := fight.run()

	if !errors.Is(err, ErrBadChoice) || !strings.Contains(err.Error(), `"first-bad"`) || outcome != NoOutcome {
		t.Errorf("run = %v, %v, want NoOutcome and ErrBadChoice naming first-bad", outcome, err)
	}
}

package engine

import (
	"reflect"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func sinAbility(t *testing.T, damage string) creature.Ability {
	return creature.Ability{
		ID:      "sin",
		Name:    "Sin",
		Trigger: &creature.Trigger{On: creature.TriggerLoyalty},
		Steps: []creature.Step{
			{Damage: []creature.DamagePart{damagePart(t, damage, "mental")}, Target: creature.TargetSelf},
			{Grant: "holy", Uses: 1, Duration: &creature.Duration{Until: "target-turn-end"}},
		},
	}
}

func loyaltyFight(t *testing.T, values ...int) (*Fight, *recorder, *Combatant, *Combatant, *Combatant) {
	hero := testCombatant("hero", SideParty, creature.KindPC, 50, 20)
	friend := testCombatant("friend", SideParty, creature.KindPC, 50, 20)
	boss := testCombatant("boss", SideEnemies, creature.KindNPC, 100, 20)
	giveAbility(boss, sinAbility(t, "5"))
	fight, events := testFight(script(t, values...), hero, friend, boss)
	fight.decider = &scriptedDecider{}
	fight.round = 1
	return fight, events, hero, friend, boss
}

func TestLoyaltyNeedsAnAnswer(t *testing.T) {
	hero := testCombatant("hero", SideParty, creature.KindPC, 50, 20)
	plain := testCombatant("plain", SideEnemies, creature.KindNPC, 100, 20)
	fight, events := testFight(script(t), hero, plain)
	fight.round = 1

	fight.loyalty(hero, "oath")
	if len(events.events) != 0 || hero.loyalRound != 0 {
		t.Fatalf("events, mark = %+v, %d, want none, 0", events.events, hero.loyalRound)
	}

	giveAbility(plain, sinAbility(t, "5"))
	fight.loyalty(hero, "oath")

	if got := eventsOf[LoyaltyAct](events.events); !slices.Equal(got, []LoyaltyAct{{Combatant: "hero", Reason: "oath"}}) {
		t.Errorf("LoyaltyAct = %+v, want one for hero", got)
	}
	if got := damageTo(events, "plain"); !slices.Equal(got, []int{5}) {
		t.Errorf("plain damage = %v, want [5]", got)
	}
}

func TestLoyaltyOncePerRound(t *testing.T) {
	fight, events, hero, friend, boss := loyaltyFight(t)

	fight.loyalty(hero, "oath")
	fight.loyalty(hero, "oath")
	fight.loyalty(friend, "oath")
	fight.moveBefore(hero, boss)
	fight.round = 2
	fight.loyalty(hero, "oath")
	fight.loyalty(hero, "oath")

	want := []LoyaltyAct{{Combatant: "hero", Reason: "oath"}, {Combatant: "friend", Reason: "oath"}, {Combatant: "hero", Reason: "oath"}}
	if got := eventsOf[LoyaltyAct](events.events); !slices.Equal(got, want) {
		t.Errorf("LoyaltyAct = %+v, want %+v", got, want)
	}
	if got := damageTo(events, "boss"); !slices.Equal(got, []int{5, 5, 5}) {
		t.Errorf("boss damage = %v, want [5 5 5]", got)
	}
}

func TestReviveIsAnActOfLoyalty(t *testing.T) {
	t.Run("a dying ally", func(t *testing.T) {
		fight, events, hero, friend, _ := loyaltyFight(t)
		friend.HP, friend.Dying = 0, 2

		fight.heal(effectSource{actor: hero, ability: "mend"}, friend, 10)

		if got := eventsOf[LoyaltyAct](events.events); !slices.Equal(got, []LoyaltyAct{{Combatant: "hero", Reason: "revive"}}) {
			t.Errorf("LoyaltyAct = %+v, want hero for revive", got)
		}
	})
	t.Run("a stable ally at 0 HP", func(t *testing.T) {
		fight, events, hero, friend, _ := loyaltyFight(t)
		friend.HP = 0

		fight.heal(effectSource{actor: hero, ability: "mend"}, friend, 10)

		if got := eventsOf[LoyaltyAct](events.events); !slices.Equal(got, []LoyaltyAct{{Combatant: "hero", Reason: "revive"}}) {
			t.Errorf("LoyaltyAct = %+v, want hero for revive", got)
		}
		if got := eventsOf[Revived](events.events); len(got) != 0 {
			t.Errorf("Revived = %+v, want none: a stable ally has no dying to lose", got)
		}
	})
	t.Run("a standing ally", func(t *testing.T) {
		fight, events, hero, friend, _ := loyaltyFight(t)
		friend.HP = 20

		fight.heal(effectSource{actor: hero, ability: "mend"}, friend, 10)

		if got := eventsOf[LoyaltyAct](events.events); len(got) != 0 {
			t.Errorf("LoyaltyAct = %+v, want none", got)
		}
	})
	t.Run("fast healing on oneself", func(t *testing.T) {
		fight, events, hero, _, _ := loyaltyFight(t)
		hero.Creature.FastHealing = 5
		hero.HP, hero.Dying = 0, 1

		fight.fastHeal(hero)

		if got := eventsOf[Revived](events.events); len(got) != 1 {
			t.Errorf("Revived = %+v, want one", got)
		}
		if got := eventsOf[LoyaltyAct](events.events); len(got) != 0 {
			t.Errorf("LoyaltyAct = %+v, want none", got)
		}
	})
	t.Run("a loyalty ability that revives fires once", func(t *testing.T) {
		fight, events, hero, friend, _ := loyaltyFight(t)
		friend.HP, friend.Dying = 0, 1
		feed := healAbility(t, "feed", creature.TargetsAlly, false, "10")
		feed.Tags = []string{creature.TagLoyalty}
		giveAbility(hero, feed)
		fight.current, hero.ActionsLeft = hero, 3

		if err := fight.perform(hero, Choice{Kind: ChooseAbility, ID: "feed", Target: friend}); err != nil {
			t.Fatalf("perform: %v", err)
		}

		if got := eventsOf[LoyaltyAct](events.events); !slices.Equal(got, []LoyaltyAct{{Combatant: "hero", Reason: "revive"}}) {
			t.Errorf("LoyaltyAct = %+v, want one, for the revive", got)
		}
		if got := eventsOf[Triggered](events.events); len(got) != 1 {
			t.Errorf("Triggered = %+v, want one", got)
		}
	})
}

func TestPassedSaveIsAnActOfLoyalty(t *testing.T) {
	cases := []struct {
		name     string
		die      int
		tagged   bool
		wantActs []LoyaltyAct
	}{
		{"success", 15, true, []LoyaltyAct{{Combatant: "hero", Reason: "song"}}},
		{"critical success", 20, true, []LoyaltyAct{{Combatant: "hero", Reason: "song"}}},
		{"failure", 10, true, nil},
		{"an untagged ability", 15, false, nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			fight, events, hero, _, boss := loyaltyFight(t, testCase.die)
			song := saveAbility("will", 15)
			song.ID = "song"
			if testCase.tagged {
				song.Tags = []string{creature.TagTestsLoyalty}
			}
			fight.current = boss

			fight.useAbility(abilityUse{actor: boss, ability: song, target: hero, actions: 2})

			if got := eventsOf[LoyaltyAct](events.events); !slices.Equal(got, testCase.wantActs) {
				t.Errorf("LoyaltyAct = %+v, want %+v", got, testCase.wantActs)
			}
		})
	}
}

func TestDisruptedAbilityIsNoAct(t *testing.T) {
	setup := func(t *testing.T, ability creature.Ability) (*Fight, *recorder, *Combatant, *Combatant) {
		fight, events, hero, friend, boss := loyaltyFight(t, 20, 1)
		guard := testCombatant("guard", SideEnemies, creature.KindNPC, 100, 20)
		giveAbility(guard, opportunity(creature.Step{Strike: "fist", Disrupts: []string{"manipulate"}}))
		fight.combatants = []*Combatant{hero, friend, guard, boss}
		fight.decider = &scriptedDecider{reactions: []string{"aoo"}}
		giveAbility(hero, ability)
		fight.current, hero.ActionsLeft = hero, 3
		return fight, events, hero, friend
	}
	t.Run("a loyalty ability", func(t *testing.T) {
		oath := creature.Ability{ID: "oath", Name: "Oath", Actions: 1, Traits: []string{"manipulate"}, Tags: []string{creature.TagLoyalty}, Targets: creature.TargetsSelf}
		fight, events, hero, _ := setup(t, oath)

		if err := fight.perform(hero, Choice{Kind: ChooseAbility, ID: "oath"}); err != nil {
			t.Fatalf("perform: %v", err)
		}

		if got := eventsOf[ActionDisrupted](events.events); len(got) != 1 {
			t.Errorf("ActionDisrupted = %+v, want one", got)
		}
		if got := eventsOf[LoyaltyAct](events.events); len(got) != 0 {
			t.Errorf("LoyaltyAct = %+v, want none", got)
		}
	})
	t.Run("a fed potion spends its pool and heals nothing", func(t *testing.T) {
		feed := healAbility(t, "feed", creature.TargetsAlly, false, "10")
		feed.Traits, feed.Tags, feed.Limit = []string{"manipulate"}, []string{creature.TagLoyalty}, &creature.Limit{Pool: "potions"}
		fight, events, hero, friend := setup(t, feed)
		hero.pools["potions"] = 1
		friend.HP, friend.Dying = 0, 1

		if err := fight.perform(hero, Choice{Kind: ChooseAbility, ID: "feed", Target: friend}); err != nil {
			t.Fatalf("perform: %v", err)
		}

		if got := eventsOf[PoolSpent](events.events); !slices.Equal(got, []PoolSpent{{Combatant: "hero", Pool: "potions", Ability: "feed", Left: 0}}) {
			t.Errorf("PoolSpent = %+v, want potions down to 0", got)
		}
		if got := eventsOf[ActionDisrupted](events.events); len(got) != 1 {
			t.Errorf("ActionDisrupted = %+v, want one", got)
		}
		if healedEvents, acts := eventsOf[Healed](events.events), eventsOf[LoyaltyAct](events.events); len(healedEvents) != 0 || len(acts) != 0 {
			t.Errorf("Healed, LoyaltyAct = %+v, %+v, want none, none", healedEvents, acts)
		}
		if friend.HP != 0 || friend.Dying != 1 {
			t.Errorf("friend HP, dying = %d, %d, want 0, 1", friend.HP, friend.Dying)
		}
	})
}

func grantedFight(t *testing.T, values ...int) (*Fight, *recorder, *Combatant, *Combatant, *Combatant) {
	hero := testCombatant("hero", SideParty, creature.KindPC, 50, 20)
	boss := testCombatant("boss", SideEnemies, creature.KindNPC, 100, 20)
	other := testCombatant("other", SideEnemies, creature.KindNPC, 100, 20)
	boss.Creature.Defense.Resist = map[string]creature.Adjustment{"all": {Value: 12, Except: []string{"holy", "cold-iron"}}}
	boss.Creature.Defense.Weakness = map[string]creature.Adjustment{"holy": {Value: 15}}
	fight, events := testFight(script(t, values...), hero, boss, other)
	fight.grantTrait(effectSource{actor: boss, ability: "sin"}, hero, creature.Step{Grant: "holy", Uses: 1, Duration: &creature.Duration{Until: "target-turn-end"}})
	events.events = nil
	return fight, events, hero, boss, other
}

func TestGrantedTrait(t *testing.T) {
	spent := []TraitSpent{{Combatant: "hero", Trait: "holy", Against: "boss"}}
	t.Run("a Strike against the granting creature carries it", func(t *testing.T) {
		fight, events, hero, boss, other := grantedFight(t, 15, 3, 15, 3, 15, 3)

		fight.strike(hero, other, fist(), strikeOptions{})
		if got := eventsOf[TraitSpent](events.events); len(got) != 0 || len(hero.grants) != 1 {
			t.Fatalf("after a Strike on another: TraitSpent, grants = %+v, %d, want none, 1", got, len(hero.grants))
		}
		fight.strike(hero, boss, fist(), strikeOptions{})
		fight.strike(hero, boss, fist(), strikeOptions{})

		dealt := eventsOf[DamageDealt](events.events)
		want := []DamageDealt{
			{Source: "hero", Target: "other", Strike: "fist", Parts: []DamageAmount{{Type: "bludgeoning", Rolled: 5, Taken: 5}}, Total: 5, HPAfter: 95},
			{Source: "hero", Target: "boss", Strike: "fist", Parts: []DamageAmount{{Type: "bludgeoning", Rolled: 5, Taken: 5}}, Total: 20, HPAfter: 80, Trait: 15},
			{Source: "hero", Target: "boss", Strike: "fist", Parts: []DamageAmount{{Type: "bludgeoning", Rolled: 5, Taken: 0}}, Total: 0, HPAfter: 80},
		}
		if !reflect.DeepEqual(dealt, want) {
			t.Errorf("DamageDealt = %+v\nwant %+v", dealt, want)
		}
		if got := eventsOf[TraitSpent](events.events); !slices.Equal(got, spent) || len(hero.grants) != 0 {
			t.Errorf("TraitSpent, grants = %+v, %d, want one, 0", got, len(hero.grants))
		}
	})
	t.Run("a miss spends it", func(t *testing.T) {
		fight, events, hero, boss, _ := grantedFight(t, 2)

		fight.strike(hero, boss, fist(), strikeOptions{})

		if got := eventsOf[TraitSpent](events.events); !slices.Equal(got, spent) || len(hero.grants) != 0 {
			t.Errorf("TraitSpent, grants = %+v, %d, want one, 0", got, len(hero.grants))
		}
	})
	t.Run("a failed concealment check spends it", func(t *testing.T) {
		fight, events, hero, boss, _ := grantedFight(t, 1)
		giveCondition(hero, "dazzled", 0)

		fight.strike(hero, boss, fist(), strikeOptions{})

		if got := eventsOf[FlatChecked](events.events); len(got) != 1 || got[0].Die != 1 {
			t.Errorf("FlatChecked = %+v, want one failed", got)
		}
		if got := eventsOf[TraitSpent](events.events); !slices.Equal(got, spent) || len(hero.grants) != 0 {
			t.Errorf("TraitSpent, grants = %+v, %d, want one, 0", got, len(hero.grants))
		}
	})
}

func TestGrantedTraitExpires(t *testing.T) {
	grant := creature.Step{Grant: "holy", Uses: 1, Duration: &creature.Duration{Until: "target-turn-end"}}
	t.Run("granted on the holder's own turn", func(t *testing.T) {
		hero := testCombatant("hero", SideParty, creature.KindPC, 50, 20)
		boss := testCombatant("boss", SideEnemies, creature.KindNPC, 100, 20)
		fight, events := testFight(script(t), hero, boss)
		fight.current = hero

		fight.grantTrait(effectSource{actor: boss, ability: "sin"}, hero, grant)
		fight.endTurn(hero)
		if len(hero.grants) != 1 {
			t.Fatalf("after this turn's end: grants = %d, want 1", len(hero.grants))
		}
		fight.endTurn(boss)
		fight.endTurn(hero)

		if len(hero.grants) != 0 {
			t.Errorf("after the next turn's end: grants = %d, want 0", len(hero.grants))
		}
		if got := eventsOf[TraitGranted](events.events); !slices.Equal(got, []TraitGranted{{Combatant: "hero", Trait: "holy", Against: "boss", Ability: "sin"}}) {
			t.Errorf("TraitGranted = %+v, want holy against boss", got)
		}
	})
	t.Run("granted on another creature's turn", func(t *testing.T) {
		hero := testCombatant("hero", SideParty, creature.KindPC, 50, 20)
		boss := testCombatant("boss", SideEnemies, creature.KindNPC, 100, 20)
		fight, _ := testFight(script(t), hero, boss)
		fight.current = boss

		fight.grantTrait(effectSource{actor: boss, ability: "sin"}, hero, grant)
		fight.endTurn(boss)
		if len(hero.grants) != 1 {
			t.Fatalf("after the granter's turn end: grants = %d, want 1", len(hero.grants))
		}
		fight.endTurn(hero)

		if len(hero.grants) != 0 {
			t.Errorf("after the holder's next turn end: grants = %d, want 0", len(hero.grants))
		}
	})
}

func TestLoyaltyDamageCanKill(t *testing.T) {
	hero := testCombatant("hero", SideParty, creature.KindPC, 50, 20)
	boss := testCombatant("boss", SideEnemies, creature.KindNPC, 100, 20)
	boss.HP = 5
	giveAbility(boss, sinAbility(t, "10"))
	giveAbility(hero, creature.Ability{ID: "oath", Name: "Oath", Actions: 1, Tags: []string{creature.TagLoyalty}, Targets: creature.TargetsSelf})
	fight, events := testFight(script(t), hero, boss)
	decider := &scriptedDecider{choices: []Choice{{Kind: ChooseAbility, ID: "oath"}, {Kind: ChooseStrike, ID: "fist", Target: boss}}}
	fight.decider = decider
	fight.round = 1

	if err := fight.takeTurn(hero); err != nil {
		t.Fatalf("takeTurn: %v", err)
	}

	if got := eventsOf[Died](events.events); !slices.Equal(got, []Died{{Combatant: "boss"}}) {
		t.Errorf("Died = %+v, want boss", got)
	}
	if got := eventsOf[TraitGranted](events.events); len(got) != 0 || len(hero.grants) != 0 {
		t.Errorf("TraitGranted, grants = %+v, %d, want none, 0", got, len(hero.grants))
	}
	if outcome, over := fight.outcome(); !over || outcome != PartyWins || len(decider.choices) != 1 {
		t.Errorf("outcome, over, choices left = %v, %v, %d, want PartyWins, true, 1", outcome, over, len(decider.choices))
	}
}

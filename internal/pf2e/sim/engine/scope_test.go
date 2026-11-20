package engine

import (
	"reflect"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func scopeFight(t *testing.T, values []int, combatants ...*Combatant) (*Fight, *recorder) {
	t.Helper()
	fight, events := testFight(script(t, values...), combatants...)
	fight.current = combatants[0]
	fight.decider = &scriptedDecider{}
	combatants[0].ActionsLeft = 3
	return fight, events
}

func useScope(t *testing.T, fight *Fight, actor *Combatant, choice Choice) {
	t.Helper()
	if err := fight.perform(actor, choice); err != nil {
		t.Fatalf("perform %s: %v", choice.ID, err)
	}
}

func damageTo(events *recorder, target string) []int {
	var totals []int
	for _, dealt := range eventsOf[DamageDealt](events.events) {
		if dealt.Target == target {
			totals = append(totals, dealt.Total)
		}
	}
	return totals
}

func TestScopeRows(t *testing.T) {
	t.Run("Vicious Swing, with Furious Focus", func(t *testing.T) {
		kas := scopeCombatant(t, "kas", SideParty)
		luna := scopeCombatant(t, "abyssal-luna", SideEnemies)
		fight, events := scopeFight(t, []int{10, 1, 2, 3, 4, 5, 1}, kas, luna)
		kas.ActionsLeft = 5

		useScope(t, fight, kas, Choice{Kind: ChooseAbility, ID: "vicious-swing", Target: luna})
		if got := damageTo(events, "abyssal-luna"); !slices.Equal(got, []int{26}) {
			t.Errorf("damage = %v, want [26] (5d12+11 rolled as 1 to 5)", got)
		}
		useScope(t, fight, kas, Choice{Kind: ChooseStrike, ID: "protectors-fury", Target: luna})
		if penalties := strikePenalties(events); !slices.Equal(penalties, []int{0, -5}) {
			t.Errorf("penalties = %v, want [0 -5]", penalties)
		}
		wantRefused(t, fight.perform(kas, Choice{Kind: ChooseAbility, ID: "vicious-swing", Target: luna}), "flourish")
	})

	t.Run("Brutal Finish", func(t *testing.T) {
		cases := []struct {
			name       string
			values     []int
			wantDamage []int
		}{
			{"hit", []int{1, 15, 1, 2, 3, 4}, []int{21}},
			{"miss", []int{1, 5, 7}, []int{7}},
			{"critical miss", []int{1, 1}, nil},
		}
		for _, testCase := range cases {
			t.Run(testCase.name, func(t *testing.T) {
				kas := scopeCombatant(t, "kas", SideParty)
				luna := scopeCombatant(t, "abyssal-luna", SideEnemies)
				fight, events := scopeFight(t, testCase.values, kas, luna)
				finish := Choice{Kind: ChooseAbility, ID: "brutal-finish", Target: luna}

				wantRefused(t, fight.perform(kas, finish), "press")
				useScope(t, fight, kas, Choice{Kind: ChooseStrike, ID: "protectors-fury", Target: luna})
				useScope(t, fight, kas, finish)

				if got := damageTo(events, "abyssal-luna"); !slices.Equal(got, testCase.wantDamage) {
					t.Errorf("damage = %v, want %v", got, testCase.wantDamage)
				}
				if penalties := strikePenalties(events); !slices.Equal(penalties, []int{0, -5}) {
					t.Errorf("penalties = %v, want [0 -5]", penalties)
				}
				if kas.ActionsLeft != 0 {
					t.Errorf("ActionsLeft = %d, want 0: the turn ends", kas.ActionsLeft)
				}
			})
		}
	})

	t.Run("Destiny Spores, Impending Fate, Fey Momentum", func(t *testing.T) {
		luna := scopeCombatant(t, "abyssal-luna", SideEnemies)
		kas := scopeCombatant(t, "kas", SideParty)
		tavish := scopeCombatant(t, "tavish", SideParty)
		arabesque := scopeCombatant(t, "arabesque", SideParty)
		giveCondition(arabesque, "doomed", 1)
		values := []int{3}
		for range 15 {
			values = append(values, 1)
		}
		values = append(values, 15, 10, 2, 20, 1, 1, 1, 1)
		fight, events := scopeFight(t, values, luna, kas, tavish, arabesque)

		useScope(t, fight, luna, Choice{Kind: ChooseAbility, ID: "destiny-spores", Target: kas})

		if got := eventsOf[RechargeRolled](events.events); !slices.Equal(got, []RechargeRolled{{Combatant: "abyssal-luna", Ability: "destiny-spores", Rounds: 3}}) {
			t.Errorf("RechargeRolled = %+v, want 3 rounds", got)
		}
		for _, check := range []struct {
			combatant  *Combatant
			wantDamage []int
			slowed     int
			doomed     int
		}{
			{kas, []int{7}, 0, 0},
			{tavish, []int{15}, 1, 0},
			{arabesque, []int{30}, 2, 2},
		} {
			if got := damageTo(events, check.combatant.ID); !slices.Equal(got, check.wantDamage) {
				t.Errorf("%s damage = %v, want %v", check.combatant.ID, got, check.wantDamage)
			}
			if check.combatant.Condition("slowed") != check.slowed || check.combatant.Condition("doomed") != check.doomed {
				t.Errorf("%s slowed, doomed = %d, %d, want %d, %d", check.combatant.ID, check.combatant.Condition("slowed"), check.combatant.Condition("doomed"), check.slowed, check.doomed)
			}
		}
		if got := luna.Blocked(luna.Creature.Abilities["destiny-spores"]); got != "recharging" {
			t.Errorf("Blocked = %q, want recharging", got)
		}

		useScope(t, fight, luna, Choice{Kind: ChooseStrike, ID: "slash", Target: kas})

		if got := damageTo(events, "kas"); !slices.Equal(got, []int{7, 36}) {
			t.Errorf("kas damage = %v, want [7 36] (the critical Slash doubles 17 + 1)", got)
		}
		if kas.Condition("doomed") != 1 {
			t.Errorf("kas doomed = %d, want 1", kas.Condition("doomed"))
		}
		if got := eventsOf[Recharged](events.events); !slices.Equal(got, []Recharged{{Combatant: "abyssal-luna", Ability: "destiny-spores", By: "slash"}}) {
			t.Errorf("Recharged = %+v, want destiny-spores by slash", got)
		}
		if got := luna.Blocked(luna.Creature.Abilities["destiny-spores"]); got != "" || luna.Condition("doomed") != 0 {
			t.Errorf("Blocked, luna doomed = %q, %d, want none, 0", got, luna.Condition("doomed"))
		}
	})

	t.Run("Walk the Timelines", func(t *testing.T) {
		luna := scopeCombatant(t, "abyssal-luna", SideEnemies)
		kas := scopeCombatant(t, "kas", SideParty)
		tavish := scopeCombatant(t, "tavish", SideParty)
		fight, events := scopeFight(t, []int{1, 1, 1}, luna, kas, tavish)
		luna.ActionsLeft = 5
		walk := Choice{Kind: ChooseAbility, ID: "walk-the-timelines", Target: kas}

		useScope(t, fight, luna, walk)
		useScope(t, fight, luna, Choice{Kind: ChooseStrike, ID: "slash", Target: kas})

		var targets []string
		for _, rolled := range eventsOf[StrikeRolled](events.events) {
			targets = append(targets, rolled.Target)
		}
		if !slices.Equal(targets, []string{"kas", "tavish", "kas"}) {
			t.Errorf("targets = %v, want [kas tavish kas]", targets)
		}
		if penalties := strikePenalties(events); !slices.Equal(penalties, []int{0, 0, -4}) {
			t.Errorf("penalties = %v, want [0 0 -4]", penalties)
		}
		if got := eventsOf[UseSpent](events.events); !slices.Equal(got, []UseSpent{{Combatant: "abyssal-luna", Ability: "walk-the-timelines", Left: 0}}) {
			t.Errorf("UseSpent = %+v, want none left", got)
		}
		wantRefused(t, fight.perform(luna, walk), "no uses left")
	})

	t.Run("Wail", func(t *testing.T) {
		eightFives := []int{5, 5, 5, 5, 5, 5, 5, 5}
		t.Run("success, critical success and failure", func(t *testing.T) {
			avatar := scopeCombatant(t, "avatar-of-lilith", SideEnemies)
			kas := scopeCombatant(t, "kas", SideParty)
			tavish := scopeCombatant(t, "tavish", SideParty)
			arabesque := scopeCombatant(t, "arabesque", SideParty)
			values := append(append([]int{2}, eightFives...), 10, 20, 8, 3)
			fight, events := scopeFight(t, values, avatar, kas, tavish, arabesque)

			useScope(t, fight, avatar, Choice{Kind: ChooseAbility, ID: "wail", Target: kas})

			degrees := map[string]Degree{}
			for _, save := range eventsOf[SaveRolled](events.events) {
				degrees[save.Combatant] = save.Degree
			}
			if want := (map[string]Degree{"kas": Success, "tavish": CriticalSuccess, "arabesque": Failure}); !reflect.DeepEqual(degrees, want) {
				t.Fatalf("degrees = %v, want %v", degrees, want)
			}
			for _, check := range []struct {
				id         string
				wantDamage []int
			}{
				{"kas", []int{40}},
				{"tavish", nil},
				{"arabesque", []int{40}},
			} {
				if got := damageTo(events, check.id); !slices.Equal(got, check.wantDamage) {
					t.Errorf("%s damage = %v, want %v", check.id, got, check.wantDamage)
				}
			}
			if got := eventsOf[HPLost](events.events); !slices.Equal(got, []HPLost{{Combatant: "arabesque", Reason: "drained", Amount: 42, HPAfter: 68}}) {
				t.Errorf("HPLost = %+v, want arabesque drained 3 for 42", got)
			}
			if arabesque.Condition("drained") != 3 {
				t.Errorf("arabesque drained = %d, want 3", arabesque.Condition("drained"))
			}
			if got := eventsOf[RechargeRolled](events.events); len(got) != 1 || got[0].Rounds != 2 {
				t.Errorf("RechargeRolled = %+v, want 2 rounds", got)
			}
		})
		t.Run("critical failure", func(t *testing.T) {
			avatar := scopeCombatant(t, "avatar-of-lilith", SideEnemies)
			kas := scopeCombatant(t, "kas", SideParty)
			values := append(append([]int{1}, eightFives...), 1)
			fight, events := scopeFight(t, values, avatar, kas)

			useScope(t, fight, avatar, Choice{Kind: ChooseAbility, ID: "wail", Target: kas})

			if got := damageTo(events, "kas"); !slices.Equal(got, []int{80}) {
				t.Errorf("damage = %v, want [80]", got)
			}
			if kas.Condition("drained") != 4 || kas.HP != 96 {
				t.Errorf("drained, HP = %d, %d, want 4, 96", kas.Condition("drained"), kas.HP)
			}
		})
	})

	t.Run("Ring of the Ram", func(t *testing.T) {
		kas := scopeCombatant(t, "kas", SideParty)
		luna := scopeCombatant(t, "abyssal-luna", SideEnemies)
		roller := script(t, 1, 2, 3, 4, 5, 6, 1, 2, 3, 5)
		fight, events := testFight(roller, kas, luna)
		fight.current = kas
		kas.ActionsLeft = 3
		ram := Choice{Kind: ChooseAbility, ID: "ring-of-the-ram", Target: luna, Actions: 3}

		useScope(t, fight, kas, ram)

		if got := damageTo(events, "abyssal-luna"); !slices.Equal(got, []int{27}) {
			t.Errorf("damage = %v, want [27] (9d6 rolled as 1 to 6, 1 to 3, on a failed save)", got)
		}
		if saves := eventsOf[SaveRolled](events.events); len(saves) != 1 || saves[0].Degree != Failure || saves[0].DC != 32 {
			t.Errorf("SaveRolled = %+v, want one failure against DC 32", saves)
		}
		if len(roller.values) != 0 || kas.ActionsLeft != 0 {
			t.Errorf("unrolled, ActionsLeft = %v, %d, want none, 0", roller.values, kas.ActionsLeft)
		}
		kas.ActionsLeft = 3
		wantRefused(t, fight.perform(kas, ram), "no uses left")
	})

	t.Run("Relic: Flare Bolt, Tidal Crash, Searing Wave", func(t *testing.T) {
		t.Run("Flare Bolt", func(t *testing.T) {
			cases := []struct {
				name           string
				values         []int
				wantDamage     int
				wantPersistent []string
			}{
				{"hit", []int{10, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, 18, []string{"3"}},
				{"critical hit", []int{20, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, 28, []string{"3 x2", "14"}},
			}
			for _, testCase := range cases {
				t.Run(testCase.name, func(t *testing.T) {
					tavish := scopeCombatant(t, "tavish", SideParty)
					luna := scopeCombatant(t, "abyssal-luna", SideEnemies)
					fight, events := scopeFight(t, testCase.values, tavish, luna)

					useScope(t, fight, tavish, Choice{Kind: ChooseAbility, ID: "relic-flare-bolt", Target: luna})

					dealt := eventsOf[DamageDealt](events.events)
					if len(dealt) == 0 || dealt[0].Total != testCase.wantDamage || !reflect.DeepEqual(dealt[0].Parts, []DamageAmount{{Type: "fire", Rolled: testCase.wantDamage, Taken: testCase.wantDamage}}) {
						t.Errorf("first DamageDealt = %+v, want %d fire in one event", dealt, testCase.wantDamage)
					}
					var persistent []string
					for _, changed := range eventsOf[PersistentChanged](events.events) {
						persistent = append(persistent, changed.Dice)
					}
					if !slices.Equal(persistent, testCase.wantPersistent) {
						t.Errorf("persistent = %v, want %v", persistent, testCase.wantPersistent)
					}
				})
			}
		})
		t.Run("Tidal Crash", func(t *testing.T) {
			tavish := scopeCombatant(t, "tavish", SideParty)
			luna := scopeCombatant(t, "abyssal-luna", SideEnemies)
			values := []int{10, 1, 1, 1}
			for range 8 {
				values = append(values, 1)
			}
			fight, events := scopeFight(t, values, tavish, luna)
			crash := Choice{Kind: ChooseAbility, ID: "relic-tidal-crash", Target: luna}

			useScope(t, fight, tavish, crash)

			want := []DamageAmount{{Type: "fire", Rolled: 11, Taken: 11}, {Type: "bludgeoning", Rolled: 16, Taken: 16}}
			if dealt := eventsOf[DamageDealt](events.events); len(dealt) != 1 || dealt[0].Total != 27 || !reflect.DeepEqual(dealt[0].Parts, want) {
				t.Errorf("DamageDealt = %+v, want 27 in one event: fire 11, bludgeoning 16", dealt)
			}
			tavish.ActionsLeft = 3
			wantRefused(t, fight.perform(tavish, crash), "no uses left")
		})
		t.Run("Searing Wave", func(t *testing.T) {
			sevenOnes := []int{1, 1, 1, 1, 1, 1, 1}
			cases := []struct {
				name       string
				values     []int
				wantDamage []int
				wantSaves  int
			}{
				{"hit", append(append([]int{4, 10, 1, 1, 1}, sevenOnes...), 2), []int{11, 7}, 1},
				{"miss", append(append([]int{4, 5}, sevenOnes...), 2), []int{8, 7}, 1},
				{"critical miss", []int{4, 1}, nil, 0},
			}
			for _, testCase := range cases {
				t.Run(testCase.name, func(t *testing.T) {
					tavish := scopeCombatant(t, "tavish", SideParty)
					luna := scopeCombatant(t, "abyssal-luna", SideEnemies)
					fight, events := scopeFight(t, testCase.values, tavish, luna)

					useScope(t, fight, tavish, Choice{Kind: ChooseAbility, ID: "relic-searing-wave", Target: luna})

					if got := damageTo(events, "abyssal-luna"); !slices.Equal(got, testCase.wantDamage) {
						t.Errorf("damage = %v, want %v", got, testCase.wantDamage)
					}
					saves := eventsOf[SaveRolled](events.events)
					if len(saves) != testCase.wantSaves {
						t.Errorf("SaveRolled = %+v, want %d", saves, testCase.wantSaves)
					}
					for _, save := range saves {
						if save.Save != "reflex" || save.DC != 35 {
							t.Errorf("save = %+v, want reflex against DC 35", save)
						}
					}
					if got := eventsOf[RechargeRolled](events.events); !slices.Equal(got, []RechargeRolled{{Combatant: "tavish", Ability: "relic-searing-wave", Rounds: 4}}) {
						t.Errorf("RechargeRolled = %+v, want 4 rounds", got)
					}
				})
			}
		})
	})

	t.Run("Quick Alchemy group", func(t *testing.T) {
		t.Run("six quick bombs empty the pool", func(t *testing.T) {
			tavish := scopeCombatant(t, "tavish", SideParty)
			luna := scopeCombatant(t, "abyssal-luna", SideEnemies)
			fight, events := scopeFight(t, []int{1, 1, 1, 1, 1, 1}, tavish, luna)
			bomb := Choice{Kind: ChooseAbility, ID: "quick-bomb", Target: luna}

			for range 6 {
				tavish.ActionsLeft = 2
				useScope(t, fight, tavish, bomb)
			}

			var left []int
			for _, spent := range eventsOf[PoolSpent](events.events) {
				left = append(left, spent.Left)
			}
			if !slices.Equal(left, []int{5, 4, 3, 2, 1, 0}) {
				t.Errorf("pool left = %v, want 5 down to 0", left)
			}
			tavish.ActionsLeft = 2
			wantRefused(t, fight.perform(tavish, bomb), "pool quick-bombs is empty")
		})
		t.Run("Sticky and Debilitating", func(t *testing.T) {
			tavish := scopeCombatant(t, "tavish", SideParty)
			luna := scopeCombatant(t, "abyssal-luna", SideEnemies)
			fight, events := scopeFight(t, []int{10, 1, 1, 1, 15, 1, 1, 1, 5}, tavish, luna)
			tavish.ActionsLeft = 4
			sticky := Choice{Kind: ChooseAbility, ID: "quick-bomb-sticky", Target: luna}
			debilitating := Choice{Kind: ChooseAbility, ID: "quick-bomb-debilitating", Target: luna}

			useScope(t, fight, tavish, sticky)
			useScope(t, fight, tavish, debilitating)

			var persistent []string
			for _, changed := range eventsOf[PersistentChanged](events.events) {
				persistent = append(persistent, changed.Dice)
			}
			if !slices.Equal(persistent, []string{"3", "8"}) {
				t.Errorf("persistent = %v, want [3 8]: Sticky adds persistent fire 8", persistent)
			}
			if saves := eventsOf[SaveRolled](events.events); len(saves) != 1 || saves[0].Save != "fortitude" || saves[0].Degree != Failure {
				t.Errorf("SaveRolled = %+v, want one failed Fortitude save", saves)
			}
			if luna.Condition("dazzled") != 1 {
				t.Errorf("luna dazzled = %d, want 1", luna.Condition("dazzled"))
			}
			tavish.ActionsLeft = 2
			wantRefused(t, fight.perform(tavish, sticky), "used this round")
			wantRefused(t, fight.perform(tavish, debilitating), "used this round")
			if tavish.Pool("quick-bombs") != 4 {
				t.Errorf("quick bombs = %d, want 4", tavish.Pool("quick-bombs"))
			}
		})
	})

	t.Run("Vengeful Spite", func(t *testing.T) {
		fourOnes := []int{1, 1, 1, 1}
		t.Run("critical hit, failed save", func(t *testing.T) {
			for _, frightenedBefore := range []int{0, 1} {
				kas := scopeCombatant(t, "kas", SideParty)
				avatar := scopeCombatant(t, "avatar-of-lilith", SideEnemies)
				values := append(append([]int{20, 1, 1, 1}, fourOnes...), 10)
				fight, events := scopeFight(t, values, kas, avatar)
				fight.decider.(*scriptedDecider).reactions = []string{"vengeful-spite"}
				if frightenedBefore > 0 {
					giveCondition(kas, "frightened", frightenedBefore)
				}

				useScope(t, fight, kas, Choice{Kind: ChooseStrike, ID: "protectors-fury", Target: avatar})

				want := []Triggered{{Combatant: "avatar-of-lilith", Ability: "vengeful-spite", On: "attacked", Cost: "reaction", By: "kas", ByStrike: "protectors-fury"}}
				if got := eventsOf[Triggered](events.events); !slices.Equal(got, want) {
					t.Errorf("frightened %d before: Triggered = %+v, want %+v", frightenedBefore, got, want)
				}
				if saves := eventsOf[SaveRolled](events.events); len(saves) != 1 || saves[0].DC != 38 || saves[0].Degree != Failure {
					t.Errorf("SaveRolled = %+v, want one failure against DC 38", saves)
				}
				if got := damageTo(events, "kas"); !slices.Equal(got, []int{18}) {
					t.Errorf("kas damage = %v, want [18] (4d10+14 rolled as 4)", got)
				}
				if kas.Condition("frightened") != 2 || kas.Condition("stunned") != 0 || !avatar.ReactionUsed {
					t.Errorf("frightened %d before: frightened, stunned, reaction used = %d, %d, %v, want 2, 0, true", frightenedBefore, kas.Condition("frightened"), kas.Condition("stunned"), avatar.ReactionUsed)
				}
			}
		})
		t.Run("critical hit, critically failed save", func(t *testing.T) {
			kas := scopeCombatant(t, "kas", SideParty)
			avatar := scopeCombatant(t, "avatar-of-lilith", SideEnemies)
			values := append(append([]int{20, 1, 1, 1, 1, 1}, fourOnes...), 1)
			fight, events := scopeFight(t, values, kas, avatar)
			fight.decider.(*scriptedDecider).reactions = []string{"vengeful-spite"}

			useScope(t, fight, kas, Choice{Kind: ChooseAbility, ID: "vicious-swing", Target: avatar})

			if got := damageTo(events, "avatar-of-lilith"); !slices.Equal(got, []int{20}) {
				t.Errorf("avatar damage = %v, want [20] ((5 + 11) x2, less resistance 12)", got)
			}
			if got := damageTo(events, "kas"); !slices.Equal(got, []int{36}) {
				t.Errorf("kas damage = %v, want [36] (18 doubled)", got)
			}
			if got := eventsOf[ActionsLost](events.events); !slices.Equal(got, []ActionsLost{{Combatant: "kas", Condition: "stunned", Lost: 1, Left: 0}}) {
				t.Errorf("ActionsLost = %+v, want the 1 action left", got)
			}
			if kas.Condition("frightened") != 2 || kas.Condition("stunned") != 3 || kas.ActionsLeft != 0 {
				t.Errorf("frightened, stunned, ActionsLeft = %d, %d, %d, want 2, 3, 0", kas.Condition("frightened"), kas.Condition("stunned"), kas.ActionsLeft)
			}

			if err := fight.act(kas); err != nil {
				t.Fatalf("next turn: %v", err)
			}
			if kas.ActionsLeft != 0 || kas.Condition("stunned") != 0 {
				t.Errorf("next turn: ActionsLeft, stunned = %d, %d, want 0, 0", kas.ActionsLeft, kas.Condition("stunned"))
			}
			if err := fight.act(kas); err != nil {
				t.Fatalf("turn after: %v", err)
			}
			if kas.ActionsLeft != 3 {
				t.Errorf("turn after: ActionsLeft = %d, want 3", kas.ActionsLeft)
			}
			if got := eventsOf[ActionsRegained](events.events); !slices.Equal(got, []ActionsRegained{{Combatant: "kas", Actions: 0, Lost: 3}}) {
				t.Errorf("ActionsRegained = %+v, want one: 0 actions, 3 lost", got)
			}
		})
		t.Run("critically failed damaging save", func(t *testing.T) {
			kas := scopeCombatant(t, "kas", SideParty)
			avatar := scopeCombatant(t, "avatar-of-lilith", SideEnemies)
			values := append(append([]int{1, 1, 1, 1}, fourOnes...), 10)
			fight, events := scopeFight(t, values, kas, avatar)
			fight.decider.(*scriptedDecider).reactions = []string{"vengeful-spite-save"}

			useScope(t, fight, kas, Choice{Kind: ChooseAbility, ID: "ring-of-the-ram", Target: avatar})

			if saves := eventsOf[SaveRolled](events.events); len(saves) != 2 || saves[0].Combatant != "avatar-of-lilith" || saves[0].Degree != CriticalFailure {
				t.Errorf("SaveRolled = %+v, want the Avatar's critical failure first", saves)
			}
			want := []Triggered{{Combatant: "avatar-of-lilith", Ability: "vengeful-spite-save", On: "saved", Cost: "reaction", By: "kas", ByAbility: "ring-of-the-ram"}}
			if got := eventsOf[Triggered](events.events); !slices.Equal(got, want) {
				t.Errorf("Triggered = %+v, want %+v", got, want)
			}
			if got := damageTo(events, "kas"); !slices.Equal(got, []int{18}) || kas.Condition("frightened") != 2 {
				t.Errorf("kas damage, frightened = %v, %d, want [18], 2", got, kas.Condition("frightened"))
			}
		})
		t.Run("reaction spent", func(t *testing.T) {
			kas := scopeCombatant(t, "kas", SideParty)
			avatar := scopeCombatant(t, "avatar-of-lilith", SideEnemies)
			roller := script(t, 20, 1, 1, 1, 1, 1, 1, 1)
			fight, events := testFight(roller, kas, avatar)
			fight.current, kas.ActionsLeft = kas, 3
			fight.decider = &scriptedDecider{reactions: []string{"vengeful-spite", "vengeful-spite-save"}}
			avatar.ReactionUsed = true

			useScope(t, fight, kas, Choice{Kind: ChooseStrike, ID: "protectors-fury", Target: avatar})
			useScope(t, fight, kas, Choice{Kind: ChooseAbility, ID: "ring-of-the-ram", Target: avatar})

			if got := eventsOf[Triggered](events.events); len(got) != 0 {
				t.Errorf("Triggered = %+v, want none", got)
			}
			if saves := eventsOf[SaveRolled](events.events); len(saves) != 1 || saves[0].Degree != CriticalFailure {
				t.Errorf("SaveRolled = %+v, want only the Avatar's critical failure", saves)
			}
			if len(roller.values) != 0 {
				t.Errorf("unrolled = %v, want none", roller.values)
			}
		})
	})

	t.Run("Lash Out", func(t *testing.T) {
		cases := []struct {
			name      string
			hp        int
			ranged    bool
			values    []int
			wantFired bool
		}{
			{"melee hit", 150, false, []int{10, 1, 1, 1, 6, 6}, true},
			{"ranged hit", 150, true, []int{10, 1, 1, 1}, false},
			{"hit that drops Arabesque", 10, false, []int{10, 1, 1, 1}, false},
		}
		for _, testCase := range cases {
			t.Run(testCase.name, func(t *testing.T) {
				avatar := scopeCombatant(t, "avatar-of-lilith", SideEnemies)
				arabesque := scopeCombatant(t, "arabesque", SideParty)
				arabesque.HP = testCase.hp
				roller := script(t, testCase.values...)
				fight, events := testFight(roller, avatar, arabesque)
				fight.current, avatar.ActionsLeft = avatar, 3
				fight.decider = &scriptedDecider{reactions: []string{"lash-out"}}
				strikeID := "slash"
				if testCase.ranged {
					bolt := avatar.Creature.Strikes["slash"]
					bolt.ID, bolt.Range = "bolt", 60
					avatar.Creature.Strikes["bolt"] = bolt
					strikeID = "bolt"
				}

				useScope(t, fight, avatar, Choice{Kind: ChooseStrike, ID: strikeID, Target: arabesque})

				if got := damageTo(events, "arabesque"); !slices.Equal(got, []int{18}) {
					t.Errorf("arabesque damage = %v, want [18]", got)
				}
				var want []Triggered
				var wantBack []DamageDealt
				if testCase.wantFired {
					want = []Triggered{{Combatant: "arabesque", Ability: "lash-out", On: "damaged", Cost: "reaction", By: "avatar-of-lilith", ByStrike: "slash"}}
					wantBack = []DamageDealt{{Source: "arabesque", Target: "avatar-of-lilith", Ability: "lash-out", Parts: []DamageAmount{{Type: "void", Rolled: 12, Taken: 0}}, Total: 0, HPAfter: 390, TriggeredBy: "avatar-of-lilith"}}
				}
				if got := eventsOf[Triggered](events.events); !slices.Equal(got, want) {
					t.Errorf("Triggered = %+v, want %+v", got, want)
				}
				var back []DamageDealt
				for _, dealt := range eventsOf[DamageDealt](events.events) {
					if dealt.Target == "avatar-of-lilith" {
						back = append(back, dealt)
					}
				}
				if !reflect.DeepEqual(back, wantBack) {
					t.Errorf("damage back = %+v, want %+v", back, wantBack)
				}
				if len(roller.values) != 0 {
					t.Errorf("unrolled = %v, want none", roller.values)
				}
			})
		}
	})

	t.Run("Fast Healing 10 and Last Stand", func(t *testing.T) {
		avatar := scopeCombatant(t, "avatar-of-lilith", SideEnemies)
		kas := scopeCombatant(t, "kas", SideParty)
		fight, events := testFight(script(t), avatar, kas)
		decider := &scriptedDecider{reactions: []string{"last-stand", "last-stand"}}
		fight.decider = decider
		avatar.HP = 70
		avatar.recharging["wail"] = 3

		if err := fight.act(avatar); err != nil {
			t.Fatalf("act: %v", err)
		}

		want := []Event{
			TurnStarted{Combatant: "avatar-of-lilith"},
			Triggered{Combatant: "avatar-of-lilith", Ability: "last-stand", On: "turn-start", Cost: "free"},
			UseSpent{Combatant: "avatar-of-lilith", Ability: "last-stand", Left: 0},
			Healed{Source: "avatar-of-lilith", Target: "avatar-of-lilith", Ability: "last-stand", Amount: 40, HPAfter: 110},
			Recharged{Combatant: "avatar-of-lilith", Ability: "wail", By: "last-stand"},
			ActionsGained{Combatant: "avatar-of-lilith", Ability: "last-stand", Amount: 1},
			Healed{Source: "avatar-of-lilith", Target: "avatar-of-lilith", Ability: "fast-healing", Amount: 10, HPAfter: 120},
		}
		if !reflect.DeepEqual(events.events, want) {
			t.Errorf("events = %+v\nwant %+v", events.events, want)
		}
		if avatar.ActionsLeft != 4 || avatar.HP != 120 || avatar.Blocked(avatar.Creature.Abilities["wail"]) != "" {
			t.Errorf("ActionsLeft, HP, wail blocked = %d, %d, %q, want 4, 120, none", avatar.ActionsLeft, avatar.HP, avatar.Blocked(avatar.Creature.Abilities["wail"]))
		}

		events.events = nil
		avatar.HP = 70
		if err := fight.act(avatar); err != nil {
			t.Fatalf("second act: %v", err)
		}
		if got := eventsOf[Triggered](events.events); len(got) != 0 || avatar.HP != 80 || avatar.ActionsLeft != 3 || len(decider.reactions) != 1 {
			t.Errorf("second turn: Triggered, HP, ActionsLeft, reactions left = %+v, %d, %d, %d, want none, 80, 3, 1", got, avatar.HP, avatar.ActionsLeft, len(decider.reactions))
		}

		fresh := scopeCombatant(t, "avatar-of-lilith", SideEnemies)
		freshFight, freshEvents := testFight(script(t), fresh, kas)
		freshFight.decider = &scriptedDecider{reactions: []string{"last-stand"}}
		fresh.HP = 76
		if err := freshFight.act(fresh); err != nil {
			t.Fatalf("act at 76: %v", err)
		}
		wantAt76 := []Event{
			TurnStarted{Combatant: "avatar-of-lilith"},
			Healed{Source: "avatar-of-lilith", Target: "avatar-of-lilith", Ability: "fast-healing", Amount: 10, HPAfter: 86},
		}
		if !reflect.DeepEqual(freshEvents.events, wantAt76) {
			t.Errorf("at 76 HP: events = %+v, want %+v", freshEvents.events, wantAt76)
		}
	})

	t.Run("Attack of Opportunity, Tactical Reflexes", func(t *testing.T) {
		hex := creature.Ability{ID: "hex", Name: "Hex", Actions: 1, Traits: []string{"manipulate", "occult"}, Steps: []creature.Step{{Condition: "off-guard"}}}
		setup := func(t *testing.T, values []int, reactions ...string) (*Fight, *recorder, *Combatant, *Combatant, *scriptedRoller) {
			avatar := scopeCombatant(t, "avatar-of-lilith", SideEnemies)
			kas := scopeCombatant(t, "kas", SideParty)
			giveAbility(avatar, hex)
			roller := script(t, values...)
			fight, events := testFight(roller, avatar, kas)
			fight.decider = &scriptedDecider{reactions: reactions}
			if err := fight.act(kas); err != nil {
				t.Fatalf("kas's turn: %v", err)
			}
			events.events = nil
			fight.current, avatar.ActionsLeft = avatar, 3
			return fight, events, avatar, kas, roller
		}

		t.Run("twice in a round, not a third time", func(t *testing.T) {
			fight, events, avatar, kas, roller := setup(t, []int{10, 1, 1, 1, 10, 1, 1, 1}, "attack-of-opportunity", "attack-of-opportunity", "attack-of-opportunity")
			for range 3 {
				useScope(t, fight, avatar, Choice{Kind: ChooseAbility, ID: "hex", Target: kas})
			}
			if penalties := strikePenalties(events); !slices.Equal(penalties, []int{0, 0}) {
				t.Errorf("penalties = %v, want [0 0]", penalties)
			}
			if got := damageTo(events, "avatar-of-lilith"); !slices.Equal(got, []int{2, 2}) {
				t.Errorf("avatar damage = %v, want [2 2] (3d12+11 rolled as 3, less resistance 12)", got)
			}
			if got := eventsOf[Triggered](events.events); len(got) != 2 || !kas.ReactionUsed || kas.extraReactions["attack-of-opportunity"] != 0 {
				t.Errorf("Triggered, reaction used, extra = %d, %v, %d, want 2, true, 0", len(got), kas.ReactionUsed, kas.extraReactions["attack-of-opportunity"])
			}
			if kas.Condition("off-guard") != 1 || len(roller.values) != 0 {
				t.Errorf("kas off-guard, unrolled = %d, %v, want 1, none", kas.Condition("off-guard"), roller.values)
			}
		})
		t.Run("a critical hit disrupts", func(t *testing.T) {
			fight, events, avatar, kas, _ := setup(t, []int{20, 1, 1, 1}, "attack-of-opportunity", "")
			useScope(t, fight, avatar, Choice{Kind: ChooseAbility, ID: "hex", Target: kas})
			want := []ActionDisrupted{{Combatant: "avatar-of-lilith", Action: "hex", By: "kas", Ability: "attack-of-opportunity"}}
			if got := eventsOf[ActionDisrupted](events.events); !slices.Equal(got, want) {
				t.Errorf("ActionDisrupted = %+v, want %+v", got, want)
			}
			if got := damageTo(events, "avatar-of-lilith"); !slices.Equal(got, []int{16}) || kas.Condition("off-guard") != 0 || avatar.ActionsLeft != 2 {
				t.Errorf("avatar damage, kas off-guard, ActionsLeft = %v, %d, %d, want [16], 0, 2", got, kas.Condition("off-guard"), avatar.ActionsLeft)
			}
		})
		t.Run("a Stand from prone", func(t *testing.T) {
			fight, events, avatar, _, _ := setup(t, []int{10, 1, 1, 1}, "attack-of-opportunity")
			giveCondition(avatar, "prone", 0)
			useScope(t, fight, avatar, Choice{Kind: ChooseStand})
			want := []Triggered{{Combatant: "kas", Ability: "attack-of-opportunity", On: "foe-acts", Cost: "reaction", By: "avatar-of-lilith"}}
			if got := eventsOf[Triggered](events.events); !slices.Equal(got, want) {
				t.Errorf("Triggered = %+v, want %+v", got, want)
			}
			if got := damageTo(events, "avatar-of-lilith"); !slices.Equal(got, []int{2}) || avatar.Condition("prone") != 0 {
				t.Errorf("avatar damage, prone = %v, %d, want [2], 0", got, avatar.Condition("prone"))
			}
		})
	})

	t.Run("Comforting Presence", func(t *testing.T) {
		for _, condition := range []string{"frightened", "stupefied"} {
			t.Run(condition, func(t *testing.T) {
				avatar := scopeCombatant(t, "avatar-of-lilith", SideEnemies)
				kas := scopeCombatant(t, "kas", SideParty)
				arabesque := scopeCombatant(t, "arabesque", SideParty)
				fight, events := testFight(script(t), avatar, kas, arabesque)
				fight.decider = &scriptedDecider{reactions: []string{"comforting-presence-" + condition}}

				fight.applyCondition(avatar, kas, creature.Step{Condition: condition, Value: 2})

				want := []Triggered{{Combatant: "arabesque", Ability: "comforting-presence-" + condition, On: "ally-gains", Cost: "reaction", By: "kas"}}
				if got := eventsOf[Triggered](events.events); !slices.Equal(got, want) {
					t.Errorf("Triggered = %+v, want %+v", got, want)
				}
				if kas.Condition(condition) != 1 || arabesque.Condition(condition) != 1 {
					t.Errorf("kas, arabesque %s = %d, %d, want 1, 1", condition, kas.Condition(condition), arabesque.Condition(condition))
				}
			})
		}
		t.Run("reaction spent on Lash Out", func(t *testing.T) {
			avatar := scopeCombatant(t, "avatar-of-lilith", SideEnemies)
			kas := scopeCombatant(t, "kas", SideParty)
			arabesque := scopeCombatant(t, "arabesque", SideParty)
			fight, events := testFight(script(t, 10, 1, 1, 1, 1, 1), avatar, kas, arabesque)
			decider := &scriptedDecider{reactions: []string{"lash-out", "comforting-presence-frightened"}}
			fight.decider = decider

			fight.strike(avatar, arabesque, avatar.Creature.Strikes["slash"], strikeOptions{})
			fight.applyCondition(avatar, kas, creature.Step{Condition: "frightened", Value: 2})

			want := []Triggered{{Combatant: "arabesque", Ability: "lash-out", On: "damaged", Cost: "reaction", By: "avatar-of-lilith", ByStrike: "slash"}}
			if got := eventsOf[Triggered](events.events); !slices.Equal(got, want) {
				t.Errorf("Triggered = %+v, want %+v", got, want)
			}
			if kas.Condition("frightened") != 2 || arabesque.Condition("frightened") != 0 || len(decider.reactions) != 1 {
				t.Errorf("kas, arabesque frightened, reactions left = %d, %d, %d, want 2, 0, 1", kas.Condition("frightened"), arabesque.Condition("frightened"), len(decider.reactions))
			}
		})
	})

	t.Run("Challenge Fate", func(t *testing.T) {
		t.Run("a Strike, then one reaction a round", func(t *testing.T) {
			kas := scopeCombatant(t, "kas", SideParty)
			luna := scopeCombatant(t, "abyssal-luna", SideEnemies)
			fight, events := scopeFight(t, []int{15, 4, 10}, kas, luna)
			decider := fight.decider.(*scriptedDecider)
			decider.reactions = []string{"challenge-fate", "challenge-fate"}

			useScope(t, fight, kas, Choice{Kind: ChooseStrike, ID: "protectors-fury", Target: luna})
			useScope(t, fight, kas, Choice{Kind: ChooseStrike, ID: "protectors-fury", Target: luna})

			wantTriggered := []Triggered{{Combatant: "abyssal-luna", Ability: "challenge-fate", On: "targeted", Cost: "reaction", By: "kas", ByStrike: "protectors-fury"}}
			if got := eventsOf[Triggered](events.events); !slices.Equal(got, wantTriggered) {
				t.Errorf("Triggered = %+v, want %+v", got, wantTriggered)
			}
			if got := eventsOf[RerollGained](events.events); !slices.Equal(got, []RerollGained{{Combatant: "kas", Source: "abyssal-luna", Ability: "challenge-fate", Keep: "worse"}}) {
				t.Errorf("RerollGained = %+v, want one on kas", got)
			}
			wantRerolled := []Rerolled{{Combatant: "kas", Roll: "attack", Keep: "worse", First: 15, Second: 4, Kept: 4, Ability: "challenge-fate"}}
			if got := eventsOf[Rerolled](events.events); !slices.Equal(got, wantRerolled) {
				t.Errorf("Rerolled = %+v, want %+v", got, wantRerolled)
			}
			rolled := strikeRolls(events)
			if len(rolled) != 2 || rolled[0].Die != 4 || rolled[0].Degree != Failure || rolled[1].Die != 10 {
				t.Errorf("StrikeRolled = %+v, want the kept 4 missing, then one die of 10", rolled)
			}
			if !luna.ReactionUsed || len(decider.reactions) != 1 || len(kas.rerolls) != 0 {
				t.Errorf("reaction used, reactions left, kas rerolls = %v, %d, %+v, want true, 1, none", luna.ReactionUsed, len(decider.reactions), kas.rerolls)
			}
		})
		t.Run("a spell attack", func(t *testing.T) {
			arabesque := scopeCombatant(t, "arabesque", SideParty)
			luna := scopeCombatant(t, "abyssal-luna", SideEnemies)
			giveAbility(arabesque, creature.Ability{ID: "bolt", Name: "Bolt", Actions: 2, Traits: []string{"concentrate"}, Steps: []creature.Step{{SpellAttack: true}}})
			fight, events := scopeFight(t, []int{18, 5}, arabesque, luna)
			fight.decider.(*scriptedDecider).reactions = []string{"challenge-fate"}

			useScope(t, fight, arabesque, Choice{Kind: ChooseAbility, ID: "bolt", Target: luna})

			wantTriggered := []Triggered{{Combatant: "abyssal-luna", Ability: "challenge-fate", On: "targeted", Cost: "reaction", By: "arabesque", ByAbility: "bolt"}}
			if got := eventsOf[Triggered](events.events); !slices.Equal(got, wantTriggered) {
				t.Errorf("Triggered = %+v, want %+v", got, wantTriggered)
			}
			wantRerolled := []Rerolled{{Combatant: "arabesque", Roll: "attack", Keep: "worse", First: 18, Second: 5, Kept: 5, Ability: "challenge-fate"}}
			if got := eventsOf[Rerolled](events.events); !slices.Equal(got, wantRerolled) {
				t.Errorf("Rerolled = %+v, want %+v", got, wantRerolled)
			}
			if rolled := eventsOf[SpellAttackRolled](events.events); len(rolled) != 1 || rolled[0].Die != 5 || rolled[0].Degree != Failure {
				t.Errorf("SpellAttackRolled = %+v, want the kept 5 missing", rolled)
			}
		})
		t.Run("a failed concealment check leaves the reaction", func(t *testing.T) {
			kas := scopeCombatant(t, "kas", SideParty)
			luna := scopeCombatant(t, "abyssal-luna", SideEnemies)
			giveCondition(kas, "dazzled", 0)
			fight, events := scopeFight(t, []int{3}, kas, luna)
			decider := fight.decider.(*scriptedDecider)
			decider.reactions = []string{"challenge-fate"}

			useScope(t, fight, kas, Choice{Kind: ChooseStrike, ID: "protectors-fury", Target: luna})

			if got := eventsOf[Triggered](events.events); len(got) != 0 || luna.ReactionUsed || len(decider.reactions) != 1 {
				t.Errorf("Triggered, reaction used, reactions left = %+v, %v, %d, want none, false, 1", got, luna.ReactionUsed, len(decider.reactions))
			}
		})
	})

	t.Run("Ill Omen", func(t *testing.T) {
		cases := []struct {
			name         string
			values       []int
			reactions    []string
			strikes      int
			wantRerolled []Rerolled
			wantDice     []int
		}{
			{
				name:      "failure: the next attack only, after the Attack of Opportunity",
				values:    []int{5, 10, 12, 3, 4},
				reactions: []string{"attack-of-opportunity"},
				strikes:   2,
				wantRerolled: []Rerolled{
					{Combatant: "kas", Roll: "attack", Keep: "worse", First: 12, Second: 3, Kept: 3, Ability: "ill-omen"},
				},
				wantDice: []int{5, 3, 4},
			},
			{
				name:    "critical failure: every attack until Luna's next turn",
				values:  []int{1, 12, 3, 14, 6, 2},
				strikes: 2,
				wantRerolled: []Rerolled{
					{Combatant: "kas", Roll: "attack", Keep: "worse", First: 12, Second: 3, Kept: 3, Ability: "ill-omen"},
					{Combatant: "kas", Roll: "attack", Keep: "worse", First: 14, Second: 6, Kept: 6, Ability: "ill-omen"},
				},
				wantDice: []int{3, 6, 2},
			},
			{
				name:     "success: nothing",
				values:   []int{15, 3},
				strikes:  1,
				wantDice: []int{3},
			},
		}
		for _, testCase := range cases {
			t.Run(testCase.name, func(t *testing.T) {
				luna := scopeCombatant(t, "abyssal-luna", SideEnemies)
				kas := scopeCombatant(t, "kas", SideParty)
				roller := script(t, testCase.values...)
				fight, events := testFight(roller, luna, kas)
				fight.decider = &scriptedDecider{reactions: testCase.reactions}
				fight.current, luna.ActionsLeft = luna, 3

				useScope(t, fight, luna, Choice{Kind: ChooseAbility, ID: "ill-omen", Target: kas})
				fight.current, kas.ActionsLeft = kas, 3
				for range testCase.strikes {
					useScope(t, fight, kas, Choice{Kind: ChooseStrike, ID: "protectors-fury", Target: luna})
				}
				fight.startTurn(luna)
				if len(roller.values) > 0 {
					useScope(t, fight, kas, Choice{Kind: ChooseStrike, ID: "protectors-fury", Target: luna})
				}

				if got := eventsOf[Rerolled](events.events); !slices.Equal(got, testCase.wantRerolled) {
					t.Errorf("Rerolled = %+v, want %+v", got, testCase.wantRerolled)
				}
				var dice []int
				for _, rolled := range strikeRolls(events) {
					dice = append(dice, rolled.Die)
				}
				if !slices.Equal(dice, testCase.wantDice) {
					t.Errorf("Strike dice = %v, want %v", dice, testCase.wantDice)
				}
				var wantTriggered []Triggered
				if testCase.reactions != nil {
					wantTriggered = []Triggered{{Combatant: "kas", Ability: "attack-of-opportunity", On: "foe-acts", Cost: "reaction", By: "abyssal-luna", ByAbility: "ill-omen"}}
				}
				if got := eventsOf[Triggered](events.events); !slices.Equal(got, wantTriggered) {
					t.Errorf("Triggered = %+v, want %+v", got, wantTriggered)
				}
				if saves := eventsOf[SaveRolled](events.events); len(saves) != 1 || saves[0].DC != 39 || saves[0].Save != "will" {
					t.Errorf("SaveRolled = %+v, want one Will save against DC 39", saves)
				}
				if len(kas.rerolls) != 0 || len(roller.values) != 0 {
					t.Errorf("kas rerolls, unrolled = %+v, %v, want none, none", kas.rerolls, roller.values)
				}
			})
		}
	})

	t.Run("Farabellus Flip", func(t *testing.T) {
		t.Run("a Slash that would hit by 1 misses; the second sees his normal AC", func(t *testing.T) {
			luna := scopeCombatant(t, "abyssal-luna", SideEnemies)
			kas := scopeCombatant(t, "kas", SideParty)
			fight, events := scopeFight(t, []int{7, 11, 1, 1, 1, 1}, luna, kas)
			decider := fight.decider.(*scriptedDecider)
			decider.reactions = []string{"farabellus-flip", "farabellus-flip"}

			useScope(t, fight, luna, Choice{Kind: ChooseStrike, ID: "slash", Target: kas})
			useScope(t, fight, luna, Choice{Kind: ChooseStrike, ID: "slash", Target: kas})

			wantTriggered := []Triggered{{Combatant: "kas", Ability: "farabellus-flip", On: "targeted", Cost: "reaction", By: "abyssal-luna", ByStrike: "slash"}}
			if got := eventsOf[Triggered](events.events); !slices.Equal(got, wantTriggered) {
				t.Errorf("Triggered = %+v, want %+v", got, wantTriggered)
			}
			if got := eventsOf[BonusGained](events.events); !slices.Equal(got, []BonusGained{{Combatant: "kas", Source: "kas", Ability: "farabellus-flip", Selector: "ac", Value: 2}}) {
				t.Errorf("BonusGained = %+v, want +2 AC on kas", got)
			}
			rolled := strikeRolls(events)
			if len(rolled) != 2 || rolled[0].Total != 37 || rolled[0].AC != 38 || rolled[0].Degree != Failure || rolled[1].AC != 36 || rolled[1].Degree != Success {
				t.Errorf("StrikeRolled = %+v, want 37 missing AC 38, then a hit against AC 36", rolled)
			}
			if got := damageTo(events, "kas"); !slices.Equal(got, []int{18}) {
				t.Errorf("kas damage = %v, want [18]", got)
			}
			if len(kas.bonuses) != 0 || len(decider.reactions) != 1 {
				t.Errorf("kas bonuses, reactions left = %+v, %d, want none, 1", kas.bonuses, len(decider.reactions))
			}
		})
		t.Run("a spell attack offers no reaction", func(t *testing.T) {
			luna := scopeCombatant(t, "abyssal-luna", SideEnemies)
			kas := scopeCombatant(t, "kas", SideParty)
			giveAbility(luna, creature.Ability{ID: "bolt", Name: "Bolt", Actions: 2, Traits: []string{"concentrate"}, Steps: []creature.Step{{SpellAttack: true}}})
			fight, events := scopeFight(t, []int{5}, luna, kas)
			decider := fight.decider.(*scriptedDecider)
			decider.reactions = []string{"farabellus-flip"}

			useScope(t, fight, luna, Choice{Kind: ChooseAbility, ID: "bolt", Target: kas})

			if got := eventsOf[Triggered](events.events); len(got) != 0 || kas.ReactionUsed || len(decider.reactions) != 1 {
				t.Errorf("Triggered, reaction used, reactions left = %+v, %v, %d, want none, false, 1", got, kas.ReactionUsed, len(decider.reactions))
			}
			if rolled := eventsOf[SpellAttackRolled](events.events); len(rolled) != 1 || rolled[0].AC != 36 {
				t.Errorf("SpellAttackRolled = %+v, want AC 36", rolled)
			}
		})
	})

	t.Run("Healing potions", func(t *testing.T) {
		kas := scopeCombatant(t, "kas", SideParty)
		tavish := scopeCombatant(t, "tavish", SideParty)
		kas.HP = 100
		tavish.HP, tavish.Dying = 0, 2
		fight, events := scopeFight(t, []int{1, 1, 1, 2, 2, 2, 1, 1, 1}, kas, tavish)
		drink := Choice{Kind: ChooseAbility, ID: "drink-healing-potion-moderate"}
		feed := Choice{Kind: ChooseAbility, ID: "feed-healing-potion-moderate", Target: tavish}

		useScope(t, fight, kas, drink)
		if kas.ActionsLeft != 2 || kas.HP != 113 {
			t.Errorf("after drinking: ActionsLeft, HP = %d, %d, want 2, 113", kas.ActionsLeft, kas.HP)
		}
		useScope(t, fight, kas, feed)
		kas.ActionsLeft = 3
		useScope(t, fight, kas, drink)

		wantSpent := []PoolSpent{
			{Combatant: "kas", Pool: "healing-potion-moderate", Ability: "drink-healing-potion-moderate", Left: 2},
			{Combatant: "kas", Pool: "healing-potion-moderate", Ability: "feed-healing-potion-moderate", Left: 1},
			{Combatant: "kas", Pool: "healing-potion-moderate", Ability: "drink-healing-potion-moderate", Left: 0},
		}
		if got := eventsOf[PoolSpent](events.events); !slices.Equal(got, wantSpent) {
			t.Errorf("PoolSpent = %+v, want %+v", got, wantSpent)
		}
		if got := eventsOf[Revived](events.events); !slices.Equal(got, []Revived{{Combatant: "tavish", DyingFrom: 2, Wounded: 1}}) {
			t.Errorf("Revived = %+v, want tavish from dying 2", got)
		}
		if kas.HP != 126 || tavish.HP != 16 {
			t.Errorf("kas, tavish HP = %d, %d, want 126, 16", kas.HP, tavish.HP)
		}
		wantRefused(t, fight.perform(kas, drink), "pool healing-potion-moderate is empty")
		wantRefused(t, fight.perform(kas, feed), "pool healing-potion-moderate is empty")
		if kas.Pool("healing-potion-greater") != 1 {
			t.Errorf("greater potions = %d, want 1", kas.Pool("healing-potion-greater"))
		}
	})

	t.Run("Elixir of Life, Healing Potion", func(t *testing.T) {
		tavish := scopeCombatant(t, "tavish", SideParty)
		kas := scopeCombatant(t, "kas", SideParty)
		tavish.HP, kas.HP = 100, 100
		fight, events := scopeFight(t, []int{1, 1, 1, 1, 1, 1, 1, 1}, tavish, kas)

		useScope(t, fight, tavish, Choice{Kind: ChooseAbility, ID: "drink-elixir-of-life"})
		useScope(t, fight, tavish, Choice{Kind: ChooseAbility, ID: "feed-healing-potion-moderate", Target: kas})

		if got := healed(events); !reflect.DeepEqual(got, map[string]int{"tavish": 17, "kas": 13}) {
			t.Errorf("healed = %v, want tavish 17 (5d6+12), kas 13 (3d8+10)", got)
		}
		tavish.ActionsLeft = 3
		for _, choice := range []Choice{
			{Kind: ChooseAbility, ID: "drink-elixir-of-life"},
			{Kind: ChooseAbility, ID: "feed-elixir-of-life", Target: kas},
			{Kind: ChooseAbility, ID: "drink-healing-potion-moderate"},
			{Kind: ChooseAbility, ID: "feed-healing-potion-moderate", Target: kas},
		} {
			wantRefused(t, fight.perform(tavish, choice), "is empty")
		}
		if tavish.Pool("elixir-of-life-moderate") != 0 || tavish.Pool("healing-potion-moderate") != 0 {
			t.Errorf("pools = %d, %d, want 0, 0", tavish.Pool("elixir-of-life-moderate"), tavish.Pool("healing-potion-moderate"))
		}
	})

	t.Run("Healing Wave", func(t *testing.T) {
		kas := scopeCombatant(t, "kas", SideParty)
		tavish := scopeCombatant(t, "tavish", SideParty)
		arabesque := scopeCombatant(t, "arabesque", SideParty)
		kas.HP, tavish.HP = 100, 100
		arabesque.HP, arabesque.Dying = 0, 1
		values := make([]int, 14)
		for index := range values {
			values[index] = 1
		}
		fight, events := scopeFight(t, values, kas, tavish, arabesque)
		wave := Choice{Kind: ChooseAbility, ID: "healing-wave"}

		useScope(t, fight, kas, wave)

		if got := healed(events); !reflect.DeepEqual(got, map[string]int{"kas": 14, "tavish": 14, "arabesque": 14}) {
			t.Errorf("healed = %v, want 14 each from one 14d4 roll", got)
		}
		if got := eventsOf[Revived](events.events); !slices.Equal(got, []Revived{{Combatant: "arabesque", DyingFrom: 1, Wounded: 1}}) {
			t.Errorf("Revived = %+v, want arabesque", got)
		}
		kas.ActionsLeft = 3
		wantRefused(t, fight.perform(kas, wave), "no uses left")
	})

	t.Run("Therapeutic Snap Peas", func(t *testing.T) {
		arabesque := scopeCombatant(t, "arabesque", SideParty)
		kas := scopeCombatant(t, "kas", SideParty)
		tavish := scopeCombatant(t, "tavish", SideParty)
		arabesque.HP, kas.HP, tavish.HP = 50, 50, 50
		values := make([]int, 20)
		for index := range values {
			values[index] = 1
		}
		fight, events := scopeFight(t, values, arabesque, kas, tavish)
		peas := Choice{Kind: ChooseAbility, ID: "therapeutic-snap-peas"}

		for range 5 {
			arabesque.ActionsLeft = 1
			useScope(t, fight, arabesque, peas)
		}

		if got := healed(events); !reflect.DeepEqual(got, map[string]int{"arabesque": 70, "kas": 70, "tavish": 70}) {
			t.Errorf("healed = %v, want 14 each, five times", got)
		}
		var left []int
		for _, spent := range eventsOf[PoolSpent](events.events) {
			left = append(left, spent.Left)
		}
		if !slices.Equal(left, []int{4, 3, 2, 1, 0}) {
			t.Errorf("pool left = %v, want 4 down to 0", left)
		}
		arabesque.ActionsLeft = 1
		wantRefused(t, fight.perform(arabesque, peas), "pool therapeutic-snap-peas is empty")
	})

	t.Run("Haste", func(t *testing.T) {
		arabesque := scopeCombatant(t, "arabesque", SideParty)
		kas := scopeCombatant(t, "kas", SideParty)
		tavish := scopeCombatant(t, "tavish", SideParty)
		luna := scopeCombatant(t, "abyssal-luna", SideEnemies)
		values := make([]int, 10)
		for index := range values {
			values[index] = 1
		}
		fight, events := scopeFight(t, values, arabesque, kas, tavish, luna)

		useScope(t, fight, arabesque, Choice{Kind: ChooseAbility, ID: "haste"})

		want := []ConditionChanged{
			{Combatant: "arabesque", Condition: "quickened", From: 0, To: 1},
			{Combatant: "kas", Condition: "quickened", From: 0, To: 1},
			{Combatant: "tavish", Condition: "quickened", From: 0, To: 1},
		}
		if got := conditionChanges(events.events); !slices.Equal(got, want) {
			t.Errorf("ConditionChanged = %+v, want %+v", got, want)
		}
		for _, hasted := range []*Combatant{arabesque, kas, tavish} {
			if len(hasted.conditions) != 1 || hasted.conditions[0].clock.rounds != 10 || hasted.conditions[0].clock.source != arabesque {
				t.Errorf("%s conditions = %+v, want quickened for 10 rounds from arabesque", hasted.ID, hasted.conditions)
			}
		}

		fight.current = kas
		fight.regainActions(kas)
		if kas.ActionsLeft != 4 || kas.StrikeOnlyActions != 1 {
			t.Fatalf("ActionsLeft, StrikeOnlyActions = %d, %d, want 4, 1", kas.ActionsLeft, kas.StrikeOnlyActions)
		}
		drink := Choice{Kind: ChooseAbility, ID: "drink-healing-potion-moderate"}
		useScope(t, fight, kas, drink)
		useScope(t, fight, kas, drink)
		wantRefused(t, fight.perform(kas, Choice{Kind: ChooseAbility, ID: "vicious-swing", Target: luna}), "(2 actions) with 1 open")
		useScope(t, fight, kas, drink)
		wantRefused(t, fight.perform(kas, Choice{Kind: ChooseAbility, ID: "drink-healing-potion-greater"}), "(1 actions) with 0 open")
		useScope(t, fight, kas, Choice{Kind: ChooseStrike, ID: "protectors-fury", Target: luna})

		if kas.ActionsLeft != 0 || kas.StrikeOnlyActions != 0 || len(strikeRolls(events)) != 1 {
			t.Errorf("ActionsLeft, StrikeOnlyActions, Strikes = %d, %d, %d, want 0, 0, 1", kas.ActionsLeft, kas.StrikeOnlyActions, len(strikeRolls(events)))
		}
	})

	t.Run("Bewitching Bloom (Cherry Blossom)", func(t *testing.T) {
		tavish := scopeCombatant(t, "tavish", SideParty)
		arabesque := scopeCombatant(t, "arabesque", SideParty)
		arabesque.HP, arabesque.Dying = 0, 2
		fight, events := scopeFight(t, []int{10, 10, 12}, tavish, arabesque)
		bloom := Choice{Kind: ChooseAbility, ID: "bewitching-bloom-cherry-blossom", Target: arabesque}

		useScope(t, fight, tavish, bloom)
		fight.recoveryCheck(arabesque)
		for range 9 {
			fight.startTurn(tavish)
		}
		arabesque.Dying = 2
		fight.recoveryCheck(arabesque)
		fight.startTurn(tavish)
		arabesque.Dying = 2
		fight.recoveryCheck(arabesque)

		if got := eventsOf[RecoveryEased](events.events); !slices.Equal(got, []RecoveryEased{{Combatant: "arabesque", Ability: "bewitching-bloom-cherry-blossom", DC: 10}}) {
			t.Errorf("RecoveryEased = %+v, want arabesque at 10", got)
		}
		var dcs []int
		for _, check := range eventsOf[RecoveryChecked](events.events) {
			dcs = append(dcs, check.DC)
		}
		if !slices.Equal(dcs, []int{10, 10, 12}) {
			t.Errorf("recovery DCs = %v, want [10 10 12]: capped for 10 rounds, then 10 + dying 2", dcs)
		}
		tavish.ActionsLeft = 3
		wantRefused(t, fight.perform(tavish, bloom), "no uses left")
	})

	t.Run("Swear Fidelity, Sin Vulnerability, Faithful Strike", func(t *testing.T) {
		kas := scopeCombatant(t, "kas", SideParty)
		avatar := scopeCombatant(t, "avatar-of-lilith", SideEnemies)
		tavish := scopeCombatant(t, "tavish", SideParty)
		values := []int{1, 1, 1, 1, 10, 1, 1, 1, 15, 1, 1, 1, 1, 1, 1, 1}
		fight, events := scopeFight(t, values, kas, avatar, tavish)
		fight.round = 1
		swear := Choice{Kind: ChooseAbility, ID: "swear-fidelity"}

		useScope(t, fight, kas, swear)

		wantSin := DamageDealt{Source: "avatar-of-lilith", Target: "avatar-of-lilith", Ability: "sin-vulnerability", Parts: []DamageAmount{{Type: "mental", Rolled: 16, Taken: 4}}, Total: 4, HPAfter: 386, TriggeredBy: "kas"}
		if dealt := eventsOf[DamageDealt](events.events); len(dealt) != 1 || !reflect.DeepEqual(dealt[0], wantSin) {
			t.Errorf("DamageDealt = %+v, want %+v", dealt, wantSin)
		}
		if got := eventsOf[TraitGranted](events.events); !slices.Equal(got, []TraitGranted{{Combatant: "kas", Trait: "holy", Against: "avatar-of-lilith", Ability: "sin-vulnerability"}}) {
			t.Errorf("TraitGranted = %+v, want holy on kas against the Avatar", got)
		}

		useScope(t, fight, kas, Choice{Kind: ChooseStrike, ID: "protectors-fury", Target: avatar})
		useScope(t, fight, kas, Choice{Kind: ChooseStrike, ID: "protectors-fury", Target: avatar})

		strikes := eventsOf[DamageDealt](events.events)[1:]
		wantStrikes := []DamageDealt{
			{Source: "kas", Target: "avatar-of-lilith", Strike: "protectors-fury", Parts: []DamageAmount{{Type: "bludgeoning", Rolled: 14, Taken: 14}}, Total: 29, HPAfter: 357, Trait: 15},
			{Source: "kas", Target: "avatar-of-lilith", Strike: "protectors-fury", Parts: []DamageAmount{{Type: "bludgeoning", Rolled: 14, Taken: 2}}, Total: 2, HPAfter: 355},
		}
		if !reflect.DeepEqual(strikes, wantStrikes) {
			t.Errorf("Strike damage = %+v\nwant %+v", strikes, wantStrikes)
		}
		if got := eventsOf[TraitSpent](events.events); !slices.Equal(got, []TraitSpent{{Combatant: "kas", Trait: "holy", Against: "avatar-of-lilith"}}) {
			t.Errorf("TraitSpent = %+v, want one", got)
		}

		kas.ActionsLeft = 1
		useScope(t, fight, kas, swear)
		fight.current, tavish.ActionsLeft = tavish, 1
		useScope(t, fight, tavish, swear)

		want := []LoyaltyAct{{Combatant: "kas", Reason: "swear-fidelity"}, {Combatant: "tavish", Reason: "swear-fidelity"}}
		if got := eventsOf[LoyaltyAct](events.events); !slices.Equal(got, want) {
			t.Errorf("LoyaltyAct = %+v, want %+v", got, want)
		}
		if got := damageTo(events, "avatar-of-lilith"); !slices.Equal(got, []int{4, 29, 2, 4}) {
			t.Errorf("avatar damage = %v, want [4 29 2 4]", got)
		}
	})

	t.Run("Feeding a potion", func(t *testing.T) {
		tavish := scopeCombatant(t, "tavish", SideParty)
		kas := scopeCombatant(t, "kas", SideParty)
		avatar := scopeCombatant(t, "avatar-of-lilith", SideEnemies)
		kas.HP, kas.Dying = 0, 2
		fight, events := scopeFight(t, []int{1, 1, 1, 1, 1, 1, 1, 1, 1}, tavish, kas, avatar)
		fight.round = 1

		useScope(t, fight, tavish, Choice{Kind: ChooseAbility, ID: "feed-elixir-of-life", Target: kas})

		if got := eventsOf[LoyaltyAct](events.events); !slices.Equal(got, []LoyaltyAct{{Combatant: "tavish", Reason: "revive"}}) {
			t.Errorf("LoyaltyAct = %+v, want one, for tavish's revive", got)
		}
		if got := eventsOf[Triggered](events.events); len(got) != 1 || got[0].Ability != "sin-vulnerability" || got[0].By != "tavish" {
			t.Errorf("Triggered = %+v, want one Sin Vulnerability by tavish", got)
		}
		if got := eventsOf[Revived](events.events); !slices.Equal(got, []Revived{{Combatant: "kas", DyingFrom: 2, Wounded: 1}}) || kas.HP != 17 {
			t.Errorf("Revived, kas HP = %+v, %d, want kas from dying 2, 17", got, kas.HP)
		}
		if got := damageTo(events, "avatar-of-lilith"); !slices.Equal(got, []int{4}) {
			t.Errorf("avatar damage = %v, want [4]", got)
		}
	})

	t.Run("Haste is an act of loyalty", func(t *testing.T) {
		arabesque := scopeCombatant(t, "arabesque", SideParty)
		kas := scopeCombatant(t, "kas", SideParty)
		tavish := scopeCombatant(t, "tavish", SideParty)
		avatar := scopeCombatant(t, "avatar-of-lilith", SideEnemies)
		fight, events := scopeFight(t, []int{1, 1, 1, 1}, arabesque, kas, tavish, avatar)
		fight.round = 1

		useScope(t, fight, arabesque, Choice{Kind: ChooseAbility, ID: "haste"})

		if got := eventsOf[LoyaltyAct](events.events); !slices.Equal(got, []LoyaltyAct{{Combatant: "arabesque", Reason: "haste"}}) {
			t.Errorf("LoyaltyAct = %+v, want arabesque for haste", got)
		}
		if got := eventsOf[TraitGranted](events.events); !slices.Equal(got, []TraitGranted{{Combatant: "arabesque", Trait: "holy", Against: "avatar-of-lilith", Ability: "sin-vulnerability"}}) {
			t.Errorf("TraitGranted = %+v, want holy on arabesque against the Avatar", got)
		}
		if arabesque.Condition("quickened") != 1 {
			t.Errorf("arabesque quickened = %d, want 1: Haste reaches the caster", arabesque.Condition("quickened"))
		}
	})

	t.Run("Demoralize, with Intimidating Glare and Intimidating Prowess", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			kas := scopeCombatant(t, "kas", SideParty)
			avatar := scopeCombatant(t, "avatar-of-lilith", SideEnemies)
			fight, events := scopeFight(t, []int{14}, kas, avatar)

			useScope(t, fight, kas, Choice{Kind: ChooseAbility, ID: "demoralize", Target: avatar})

			want := []SkillChecked{{Combatant: "kas", Target: "avatar-of-lilith", Ability: "demoralize", Skill: "intimidation", Die: 14, Bonus: 28, Total: 42, DC: 42, Degree: Success}}
			if got := eventsOf[SkillChecked](events.events); !slices.Equal(got, want) {
				t.Errorf("SkillChecked = %+v, want %+v", got, want)
			}
			if avatar.Condition("frightened") != 1 {
				t.Errorf("avatar frightened = %d, want 1", avatar.Condition("frightened"))
			}
		})

		t.Run("critical success, then a Strike against her lowered AC", func(t *testing.T) {
			kas := scopeCombatant(t, "kas", SideParty)
			avatar := scopeCombatant(t, "avatar-of-lilith", SideEnemies)
			fight, events := scopeFight(t, []int{20, 10, 1, 1, 1}, kas, avatar)

			useScope(t, fight, kas, Choice{Kind: ChooseAbility, ID: "demoralize", Target: avatar})
			if got := eventsOf[SkillChecked](events.events); len(got) != 1 || got[0].Total != 48 || got[0].Degree != CriticalSuccess {
				t.Errorf("SkillChecked = %+v, want a natural 20 for 48, a critical success", got)
			}
			if avatar.Condition("frightened") != 2 {
				t.Errorf("avatar frightened = %d, want 2", avatar.Condition("frightened"))
			}
			useScope(t, fight, kas, Choice{Kind: ChooseStrike, ID: "protectors-fury", Target: avatar})
			want := []StrikeRolled{{Attacker: "kas", Target: "avatar-of-lilith", Strike: "protectors-fury", Die: 10, Bonus: 30, Total: 40, AC: 37, Degree: Success}}
			if got := eventsOf[StrikeRolled](events.events); !slices.Equal(got, want) {
				t.Errorf("StrikeRolled = %+v, want %+v (AC 39 less frightened 2)", got, want)
			}
		})

		t.Run("a target immune to visual effects is skipped", func(t *testing.T) {
			kas := scopeCombatant(t, "kas", SideParty)
			avatar := scopeCombatant(t, "avatar-of-lilith", SideEnemies)
			avatar.Creature.Defense.Immune = append(avatar.Creature.Defense.Immune, "visual")
			fight, events := scopeFight(t, nil, kas, avatar)

			useScope(t, fight, kas, Choice{Kind: ChooseAbility, ID: "demoralize", Target: avatar})

			if got := eventsOf[ImmuneToEffect](events.events); !slices.Equal(got, []ImmuneToEffect{{Combatant: "avatar-of-lilith", Source: "kas", Ability: "demoralize"}}) {
				t.Errorf("ImmuneToEffect = %+v, want the Avatar against demoralize", got)
			}
			if got := eventsOf[SkillChecked](events.events); len(got) != 0 || avatar.Condition("frightened") != 0 {
				t.Errorf("SkillChecked = %+v, frightened = %d, want no roll and no condition", got, avatar.Condition("frightened"))
			}
		})
	})

	t.Run("Battle Cry", func(t *testing.T) {
		kas := scopeCombatant(t, "kas", SideParty).Creature
		avatar := scopeCombatant(t, "avatar-of-lilith", SideEnemies).Creature
		watch := &roundWatch{}
		fight, err := newFight(Setup{
			Party:     []*creature.Creature{kas},
			Enemies:   []*creature.Creature{avatar},
			Roller:    script(t, 10, 10, 14),
			Observer:  watch,
			Decider:   &scriptedDecider{reactions: []string{"battle-cry"}},
			MaxRounds: 1,
		})
		if err != nil {
			t.Fatalf("newFight: %v", err)
		}
		watch.combatant = fight.combatants[0]
		if outcome, err := fight.run(); err != nil || outcome != Stalemate {
			t.Fatalf("run = %v, %v, want a stalemate", outcome, err)
		}

		lastInitiative, checked, roundOne := -1, -1, -1
		for index, event := range watch.events {
			switch event.(type) {
			case InitiativeRolled:
				lastInitiative = index
			case SkillChecked:
				checked = index
			case RoundStarted:
				if roundOne < 0 {
					roundOne = index
				}
			}
		}
		if !(lastInitiative < checked && checked < roundOne) {
			t.Errorf("events at %d (last initiative), %d (check), %d (round 1), want the check between", lastInitiative, checked, roundOne)
		}
		wantTriggered := []Triggered{{Combatant: "kas", Ability: "battle-cry", On: "initiative", Cost: "free"}}
		if got := eventsOf[Triggered](watch.events); !slices.Equal(got, wantTriggered) {
			t.Errorf("Triggered = %+v, want %+v", got, wantTriggered)
		}
		wantChecked := []SkillChecked{{Combatant: "kas", Target: "avatar-of-lilith", Ability: "battle-cry", Skill: "intimidation", Die: 14, Bonus: 28, Total: 42, DC: 42, Degree: Success}}
		if got := eventsOf[SkillChecked](watch.events); !slices.Equal(got, wantChecked) {
			t.Errorf("SkillChecked = %+v, want %+v", got, wantChecked)
		}
		if !slices.Contains(watch.events, Event(ConditionChanged{Combatant: "avatar-of-lilith", Condition: "frightened", From: 0, To: 1})) {
			t.Error("the Avatar was not frightened 1")
		}
		if watch.reactionAtRound {
			t.Error("kas ReactionUsed = true at round 1, want his reaction unspent: Battle Cry is free")
		}
		if got := eventsOf[ActionUsed](watch.events); len(got) != 0 {
			t.Errorf("ActionUsed = %+v, want none", got)
		}
	})

	t.Run("Too Angry to Die", func(t *testing.T) {
		t.Run("dropped, healed by fast healing, Stands on his turn", func(t *testing.T) {
			avatar := scopeCombatant(t, "avatar-of-lilith", SideEnemies)
			kas := scopeCombatant(t, "kas", SideParty)
			kas.Creature.FastHealing = 10
			fight, events := scopeFight(t, []int{14}, avatar, kas)
			decider := &scriptedDecider{choices: []Choice{{Kind: ChooseStand}}, reactions: []string{"too-angry-to-die"}}
			fight.decider = decider
			kas.HP = 0
			fight.knockDown(kas, avatar, false)

			if err := fight.takeTurn(kas); err != nil {
				t.Fatalf("takeTurn: %v", err)
			}

			if got := eventsOf[Revived](events.events); !slices.Equal(got, []Revived{{Combatant: "kas", DyingFrom: 1, Wounded: 1}}) {
				t.Errorf("Revived = %+v, want kas from dying 1", got)
			}
			wantTriggered := []Triggered{{Combatant: "kas", Ability: "too-angry-to-die", On: "stood-up", Cost: "reaction"}}
			if got := eventsOf[Triggered](events.events); !slices.Equal(got, wantTriggered) {
				t.Errorf("Triggered = %+v, want %+v", got, wantTriggered)
			}
			wantChecked := []SkillChecked{{Combatant: "kas", Target: "avatar-of-lilith", Ability: "too-angry-to-die", Skill: "intimidation", Die: 14, Bonus: 28, Total: 42, DC: 42, Degree: Success}}
			if got := eventsOf[SkillChecked](events.events); !slices.Equal(got, wantChecked) {
				t.Errorf("SkillChecked = %+v, want %+v", got, wantChecked)
			}
			if avatar.Condition("frightened") != 1 || !kas.ReactionUsed {
				t.Errorf("avatar frightened, kas ReactionUsed = %d, %v, want 1, true", avatar.Condition("frightened"), kas.ReactionUsed)
			}
		})

		t.Run("a Stand from an ordinary prone offers nothing", func(t *testing.T) {
			kas := scopeCombatant(t, "kas", SideParty)
			avatar := scopeCombatant(t, "avatar-of-lilith", SideEnemies)
			fight, events := scopeFight(t, nil, kas, avatar)
			decider := &scriptedDecider{reactions: []string{"too-angry-to-die"}}
			fight.decider = decider
			giveCondition(kas, "prone", 0)

			useScope(t, fight, kas, Choice{Kind: ChooseStand})

			if got := eventsOf[StoodUp](events.events); len(got) != 1 {
				t.Errorf("StoodUp = %+v, want one", got)
			}
			if got := eventsOf[Triggered](events.events); len(got) != 0 || len(decider.reactions) != 1 {
				t.Errorf("Triggered = %+v, reactions left %v, want nothing offered", got, decider.reactions)
			}
		})

		t.Run("reaction spent on Farabellus Flip", func(t *testing.T) {
			kas := scopeCombatant(t, "kas", SideParty)
			avatar := scopeCombatant(t, "avatar-of-lilith", SideEnemies)
			fight, events := scopeFight(t, []int{1}, kas, avatar)
			decider := &scriptedDecider{reactions: []string{"farabellus-flip", "too-angry-to-die"}}
			fight.decider = decider
			kas.downed = true
			giveCondition(kas, "prone", 0)

			fight.strike(avatar, kas, avatar.Creature.Strikes["slash"], strikeOptions{})
			useScope(t, fight, kas, Choice{Kind: ChooseStand})

			wantTriggered := []Triggered{{Combatant: "kas", Ability: "farabellus-flip", On: "targeted", Cost: "reaction", By: "avatar-of-lilith", ByStrike: "slash"}}
			if got := eventsOf[Triggered](events.events); !slices.Equal(got, wantTriggered) {
				t.Errorf("Triggered = %+v, want %+v", got, wantTriggered)
			}
			if got := eventsOf[StoodUp](events.events); len(got) != 1 || len(decider.reactions) != 1 {
				t.Errorf("StoodUp = %+v, reactions left %v, want one Stand and Too Angry to Die not offered", got, decider.reactions)
			}
		})
	})
}

type roundWatch struct {
	recorder
	combatant       *Combatant
	reactionAtRound bool
	sawRound        bool
}

func (watch *roundWatch) Observe(event Event) {
	if _, isRound := event.(RoundStarted); isRound && !watch.sawRound {
		watch.sawRound = true
		watch.reactionAtRound = watch.combatant.ReactionUsed
	}
	watch.recorder.Observe(event)
}

package goal

import (
	"reflect"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// goalFixture is one shipped goal held by f1 (Force 3, Cunning 3, Wealth 2), with
// rival f2 (Force 3, Cunning 4, Wealth 4, HP 5) and third party f3.
type goalFixture struct {
	t            *testing.T
	compiled     *compiledGoal
	rulebook     *rulebook.Rulebook
	holder       *domain.Faction
	rival        *domain.Faction
	third        *domain.Faction
	factionState *state.FactionState
}

func newGoalFixture(t *testing.T, goalName string, progress int) *goalFixture {
	t.Helper()
	loaded := testRulebook(t)
	compiled := shippedGoal(t, loaded, goalName)
	holder := &domain.Faction{
		ID: "f1", Force: 3, Cunning: 3, Wealth: 2, CurrentHP: 10,
		Homeworld:  domain.Location{WorldID: "Tartarus"},
		ActiveGoal: &domain.ActiveGoal{GoalID: compiled.goal.ID, Progress: progress},
		Assets:     map[string]*domain.Asset{},
	}
	rival := &domain.Faction{
		ID: "f2", Force: 3, Cunning: 4, Wealth: 4, CurrentHP: 5,
		Homeworld: domain.Location{WorldID: "Krylos"},
		Assets:    map[string]*domain.Asset{},
	}
	third := &domain.Faction{
		ID: "f3", Force: 1, Cunning: 1, Wealth: 1, CurrentHP: 2,
		Homeworld: domain.Location{WorldID: "Vesta"},
		Assets:    map[string]*domain.Asset{},
	}
	return &goalFixture{
		t:        t,
		compiled: compiled,
		rulebook: loaded,
		holder:   holder,
		rival:    rival,
		third:    third,
		factionState: &state.FactionState{
			Factions:    map[string]*domain.Faction{"f1": holder, "f2": rival, "f3": third},
			Governments: map[string]string{"Tartarus": "f1", "Krylos": "f2"},
		},
	}
}

// testRulebook is the shipped rulebook with four test asset definitions added.
func testRulebook(t *testing.T) *rulebook.Rulebook {
	t.Helper()
	loaded := shippedRulebook(t)
	for _, definition := range []*domain.AssetDefinition{
		{ID: "force-def", Category: domain.StatForce, MinRating: 1},
		{ID: "force-high", Category: domain.StatForce, MinRating: 5},
		{ID: "wealth-def", Category: domain.StatWealth, MinRating: 1},
		{ID: "cunning-def", Category: domain.StatCunning, MinRating: 1},
	} {
		loaded.Assets[definition.ID] = definition
	}
	return loaded
}

func (fixture *goalFixture) give(owner *domain.Faction, assetID, definitionID, worldID string) *domain.Asset {
	asset := &domain.Asset{ID: assetID, DefinitionID: definitionID, OwnerID: owner.ID, Location: domain.Location{WorldID: worldID}}
	owner.Assets[assetID] = asset
	return asset
}

func (fixture *goalFixture) evaluate(batch ...domain.Mutation) []domain.Mutation {
	return fixture.compiled.evaluate(evalContext{
		faction:      fixture.holder,
		batch:        batch,
		factionState: fixture.factionState,
		rulebook:     fixture.rulebook,
	})
}

func (fixture *goalFixture) progressed(delta int, cause string) domain.GoalProgressed {
	return domain.GoalProgressed{FactionID: "f1", GoalID: fixture.compiled.goal.ID, Delta: delta, Cause: cause}
}

func (fixture *goalFixture) completed(xp int) []domain.Mutation {
	return []domain.Mutation{
		domain.GoalCompleted{FactionID: "f1", GoalID: fixture.compiled.goal.ID, XPAwarded: xp, Cause: causeCompleted},
		domain.XPAwarded{FactionID: "f1", Amount: xp, Cause: causeCompleted},
	}
}

func (fixture *goalFixture) want(got []domain.Mutation, want ...domain.Mutation) {
	fixture.t.Helper()
	if len(got) == 0 && len(want) == 0 {
		return
	}
	if !reflect.DeepEqual(got, want) {
		fixture.t.Errorf("mutations =\n  %+v\nwant\n  %+v", got, want)
	}
}

func kill(ownerID, assetID, cause, causedByID string) domain.AssetRemoved {
	return domain.AssetRemoved{FactionID: ownerID, AssetID: assetID, Cause: cause, CausedByFactionID: causedByID}
}

// --- Kills (CORE-G-001 to CORE-G-003, CORE-G-010), D-9 and D-14 ---

func TestEvaluate_MilitaryConquest_KillProgresses(t *testing.T) {
	fixture := newGoalFixture(t, "Military Conquest", 0)
	fixture.give(fixture.rival, "a1", "force-def", "Krylos")

	got := fixture.evaluate(kill("f2", "a1", "attack", "f1"))

	fixture.want(got, fixture.progressed(1, causeProgress))
}

func TestEvaluate_MilitaryConquest_AnyCauseCounts(t *testing.T) {
	for _, cause := range []string{"ability", "counterstrike"} {
		t.Run(cause, func(t *testing.T) {
			fixture := newGoalFixture(t, "Military Conquest", 0)
			fixture.give(fixture.rival, "a1", "force-def", "Krylos")

			got := fixture.evaluate(kill("f2", "a1", cause, "f1"))

			fixture.want(got, fixture.progressed(1, causeProgress))
		})
	}
}

func TestEvaluate_MilitaryConquest_KillsThatDoNotCount(t *testing.T) {
	cases := []struct {
		name  string
		setup func(fixture *goalFixture) domain.Mutation
	}{
		{"a Wealth kill", func(fixture *goalFixture) domain.Mutation {
			fixture.give(fixture.rival, "a1", "wealth-def", "Krylos")
			return kill("f2", "a1", "attack", "f1")
		}},
		{"a kill caused by f3", func(fixture *goalFixture) domain.Mutation {
			fixture.give(fixture.rival, "a1", "force-def", "Krylos")
			return kill("f2", "a1", "attack", "f3")
		}},
		{"a kill of the holder's own asset", func(fixture *goalFixture) domain.Mutation {
			fixture.give(fixture.holder, "a1", "force-def", "Tartarus")
			return kill("f1", "a1", "attack", "f1")
		}},
		{"a removal with no cause faction", func(fixture *goalFixture) domain.Mutation {
			fixture.give(fixture.rival, "a1", "force-def", "Krylos")
			return kill("f2", "a1", "bookkeeping", "")
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newGoalFixture(t, "Military Conquest", 0)

			got := fixture.evaluate(testCase.setup(fixture))

			fixture.want(got)
		})
	}
}

func TestEvaluate_MilitaryConquest_CompletesWithHalfProgressRoundedUp(t *testing.T) {
	fixture := newGoalFixture(t, "Military Conquest", 2)
	fixture.give(fixture.rival, "a1", "force-def", "Krylos")

	got := fixture.evaluate(kill("f2", "a1", "attack", "f1"))

	fixture.want(got, append([]domain.Mutation{fixture.progressed(1, causeProgress)}, fixture.completed(2)...)...)
}

func TestEvaluate_MilitaryConquest_TwoKillsInOneBatch(t *testing.T) {
	fixture := newGoalFixture(t, "Military Conquest", 2)
	fixture.give(fixture.rival, "a1", "force-def", "Krylos")
	fixture.give(fixture.rival, "a2", "force-def", "Krylos")

	got := fixture.evaluate(kill("f2", "a1", "attack", "f1"), kill("f2", "a2", "attack", "f1"))

	fixture.want(got, append([]domain.Mutation{fixture.progressed(2, causeProgress)}, fixture.completed(2)...)...)
}

func TestEvaluate_KillGoalsCountOnlyTheirCategory(t *testing.T) {
	cases := []struct {
		goalName   string
		counted    string
		notCounted string
	}{
		{"Commercial Expansion", "wealth-def", "force-def"},
		{"Intelligence Coup", "cunning-def", "wealth-def"},
	}
	for _, testCase := range cases {
		t.Run(testCase.goalName, func(t *testing.T) {
			fixture := newGoalFixture(t, testCase.goalName, 0)
			fixture.give(fixture.rival, "counted", testCase.counted, "Krylos")
			fixture.give(fixture.rival, "other", testCase.notCounted, "Krylos")

			got := fixture.evaluate(kill("f2", "counted", "attack", "f1"), kill("f2", "other", "attack", "f1"))

			fixture.want(got, fixture.progressed(1, causeProgress))
		})
	}
}

func TestEvaluate_InvincibleValor(t *testing.T) {
	t.Run("a kill rated above Force completes", func(t *testing.T) {
		fixture := newGoalFixture(t, "Invincible Valor", 0)
		fixture.give(fixture.rival, "a1", "force-high", "Krylos")

		got := fixture.evaluate(kill("f2", "a1", "attack", "f1"))

		fixture.want(got, fixture.completed(2)...)
	})
	t.Run("a kill rated below Force gives nothing", func(t *testing.T) {
		fixture := newGoalFixture(t, "Invincible Valor", 0)
		fixture.give(fixture.rival, "a1", "force-def", "Krylos")

		got := fixture.evaluate(kill("f2", "a1", "attack", "f1"))

		fixture.want(got)
	})
	t.Run("a kill rated equal to Force gives nothing", func(t *testing.T) {
		fixture := newGoalFixture(t, "Invincible Valor", 0)
		fixture.rulebook.Assets["force-equal"] = &domain.AssetDefinition{ID: "force-equal", Category: domain.StatForce, MinRating: 3}
		fixture.give(fixture.rival, "a1", "force-equal", "Krylos")

		got := fixture.evaluate(kill("f2", "a1", "attack", "f1"))

		fixture.want(got)
	})
}

// --- Damage (CORE-G-006), Decision 1 ---

func TestEvaluate_BloodTheEnemy_AssetDamage(t *testing.T) {
	fixture := newGoalFixture(t, "Blood the Enemy", 0)

	got := fixture.evaluate(domain.AssetHPDelta{FactionID: "f2", AssetID: "a1", Delta: -3, CausedByFactionID: "f1"})

	fixture.want(got, fixture.progressed(3, causeProgress))
}

func TestEvaluate_BloodTheEnemy_BaseHitCountsOnce(t *testing.T) {
	fixture := newGoalFixture(t, "Blood the Enemy", 0)

	got := fixture.evaluate(
		domain.BaseHPDelta{FactionID: "f2", BaseID: "b1", Delta: -4, Cause: "attack", CausedByFactionID: "f1"},
		domain.FactionHPDelta{FactionID: "f2", Delta: -4, Cause: "attack", CausedByFactionID: "f1"},
	)

	fixture.want(got, fixture.progressed(4, causeProgress))
}

func TestEvaluate_BloodTheEnemy_DamageThatDoesNotCount(t *testing.T) {
	cases := map[string]domain.Mutation{
		"a positive delta":             domain.AssetHPDelta{FactionID: "f2", AssetID: "a1", Delta: 2, CausedByFactionID: "f1"},
		"a zero delta":                 domain.AssetHPDelta{FactionID: "f2", AssetID: "a1", Delta: 0, CausedByFactionID: "f1"},
		"damage to the holder's own":   domain.AssetHPDelta{FactionID: "f1", AssetID: "a1", Delta: -2, CausedByFactionID: "f1"},
		"damage caused by a third":     domain.AssetHPDelta{FactionID: "f2", AssetID: "a1", Delta: -2, CausedByFactionID: "f3"},
		"a faction hit with no Base":   domain.FactionHPDelta{FactionID: "f2", Delta: -2, CausedByFactionID: "f1"},
		"a Base hit caused by a third": domain.BaseHPDelta{FactionID: "f2", BaseID: "b1", Delta: -2, CausedByFactionID: "f3"},
	}
	for name, mutation := range cases {
		t.Run(name, func(t *testing.T) {
			fixture := newGoalFixture(t, "Blood the Enemy", 0)

			got := fixture.evaluate(mutation)

			fixture.want(got)
		})
	}
}

func TestEvaluate_BloodTheEnemy_OverkillDoesNotCount(t *testing.T) {
	t.Run("a hit counts only the HP the asset had left", func(t *testing.T) {
		fixture := newGoalFixture(t, "Blood the Enemy", 0)
		fixture.give(fixture.rival, "a1", "force-def", "Krylos").CurrentHP = 1

		got := fixture.evaluate(domain.AssetHPDelta{FactionID: "f2", AssetID: "a1", Delta: -10, CausedByFactionID: "f1"})

		fixture.want(got, fixture.progressed(1, causeProgress))
	})
	t.Run("hits in one batch share the HP left", func(t *testing.T) {
		fixture := newGoalFixture(t, "Blood the Enemy", 0)
		fixture.give(fixture.rival, "a1", "force-def", "Krylos").CurrentHP = 3

		got := fixture.evaluate(
			domain.AssetHPDelta{FactionID: "f2", AssetID: "a1", Delta: -2, CausedByFactionID: "f1"},
			domain.AssetHPDelta{FactionID: "f2", AssetID: "a1", Delta: -2, CausedByFactionID: "f1"},
			domain.AssetHPDelta{FactionID: "f2", AssetID: "a1", Delta: -2, CausedByFactionID: "f1"},
		)

		fixture.want(got, fixture.progressed(3, causeProgress))
	})
	t.Run("a Base hit counts only the HP the Base had left", func(t *testing.T) {
		fixture := newGoalFixture(t, "Blood the Enemy", 0)
		fixture.rival.Bases = []*domain.Base{{ID: "b2", OwnerID: "f2", CurrentHP: 2, Location: domain.Location{WorldID: "Krylos"}}}

		got := fixture.evaluate(
			domain.BaseHPDelta{FactionID: "f2", BaseID: "b2", Delta: -5, CausedByFactionID: "f1"},
			domain.FactionHPDelta{FactionID: "f2", Delta: -5, CausedByFactionID: "f1"},
		)

		fixture.want(got, fixture.progressed(2, causeProgress))
	})
	t.Run("a third party's hit uses up the HP first", func(t *testing.T) {
		fixture := newGoalFixture(t, "Blood the Enemy", 0)
		fixture.give(fixture.rival, "a1", "force-def", "Krylos").CurrentHP = 3

		got := fixture.evaluate(
			domain.AssetHPDelta{FactionID: "f2", AssetID: "a1", Delta: -2, CausedByFactionID: "f3"},
			domain.AssetHPDelta{FactionID: "f2", AssetID: "a1", Delta: -4, CausedByFactionID: "f1"},
		)

		fixture.want(got, fixture.progressed(1, causeProgress))
	})
}

func TestEvaluate_BloodTheEnemy_Completes(t *testing.T) {
	fixture := newGoalFixture(t, "Blood the Enemy", 6)

	got := fixture.evaluate(domain.AssetHPDelta{FactionID: "f2", AssetID: "a1", Delta: -2, CausedByFactionID: "f1"})

	fixture.want(got, append([]domain.Mutation{fixture.progressed(2, causeProgress)}, fixture.completed(2)...)...)
}

// --- Base planted (CORE-G-005), D-22 ---

func basePlanted(causedByID, worldID string) domain.BaseAdded {
	return domain.BaseAdded{
		FactionID:         "f1",
		Base:              domain.Base{ID: "new-base", OwnerID: "f1", Location: domain.Location{WorldID: worldID}},
		Cause:             "expand_influence",
		CausedByFactionID: causedByID,
	}
}

func TestEvaluate_ExpandInfluence(t *testing.T) {
	t.Run("a new world with no rival pays 1", func(t *testing.T) {
		fixture := newGoalFixture(t, "Expand Influence", 0)

		got := fixture.evaluate(basePlanted("f1", "Hadrian"))

		fixture.want(got, fixture.completed(1)...)
	})
	t.Run("a rival asset on the world pays 2", func(t *testing.T) {
		fixture := newGoalFixture(t, "Expand Influence", 0)
		fixture.give(fixture.rival, "a1", "force-def", "Hadrian")

		got := fixture.evaluate(basePlanted("f1", "Hadrian"))

		fixture.want(got, fixture.completed(2)...)
	})
	t.Run("a rival Base alone on the world pays 2", func(t *testing.T) {
		fixture := newGoalFixture(t, "Expand Influence", 0)
		fixture.rival.Bases = []*domain.Base{{ID: "b2", OwnerID: "f2", Location: domain.Location{WorldID: "Hadrian"}}}

		got := fixture.evaluate(basePlanted("f1", "Hadrian"))

		fixture.want(got, fixture.completed(2)...)
	})
	t.Run("a destroyed rival's leftovers on the world pay 1", func(t *testing.T) {
		fixture := newGoalFixture(t, "Expand Influence", 0)
		fixture.rival.CurrentHP = 0
		fixture.give(fixture.rival, "a1", "force-def", "Hadrian")
		fixture.rival.Bases = []*domain.Base{{ID: "b2", OwnerID: "f2", Location: domain.Location{WorldID: "Hadrian"}}}

		got := fixture.evaluate(basePlanted("f1", "Hadrian"))

		fixture.want(got, fixture.completed(1)...)
	})
	t.Run("a world the holder already holds gives nothing", func(t *testing.T) {
		fixture := newGoalFixture(t, "Expand Influence", 0)
		fixture.holder.Bases = []*domain.Base{{ID: "b1", OwnerID: "f1", Location: domain.Location{WorldID: "Hadrian"}}}

		got := fixture.evaluate(basePlanted("f1", "Hadrian"))

		fixture.want(got)
	})
	t.Run("a Base caused by a rival gives nothing", func(t *testing.T) {
		fixture := newGoalFixture(t, "Expand Influence", 0)

		got := fixture.evaluate(basePlanted("f2", "Hadrian"))

		fixture.want(got)
	})
}

// --- Influence (CORE-G-011) ---

func TestEvaluate_WealthOfWorlds(t *testing.T) {
	t.Run("bought influence counts its delta", func(t *testing.T) {
		fixture := newGoalFixture(t, "Wealth of Worlds", 0)

		got := fixture.evaluate(domain.InfluenceDelta{FactionID: "f1", BaseID: "b1", Delta: 3, CausedByFactionID: "f1"})

		fixture.want(got, fixture.progressed(3, causeProgress))
	})
	t.Run("a lost influence gives nothing", func(t *testing.T) {
		fixture := newGoalFixture(t, "Wealth of Worlds", 0)

		got := fixture.evaluate(domain.InfluenceDelta{FactionID: "f1", BaseID: "b1", Delta: -2, CausedByFactionID: "f1"})

		fixture.want(got)
	})
	t.Run("influence caused by a rival gives nothing", func(t *testing.T) {
		fixture := newGoalFixture(t, "Wealth of Worlds", 0)

		got := fixture.evaluate(domain.InfluenceDelta{FactionID: "f2", BaseID: "b2", Delta: 3, CausedByFactionID: "f2"})

		fixture.want(got)
	})
	t.Run("four times Wealth completes", func(t *testing.T) {
		fixture := newGoalFixture(t, "Wealth of Worlds", 6)

		got := fixture.evaluate(domain.InfluenceDelta{FactionID: "f1", BaseID: "b1", Delta: 2, CausedByFactionID: "f1"})

		fixture.want(got, append([]domain.Mutation{fixture.progressed(2, causeProgress)}, fixture.completed(2)...)...)
	})
}

// --- Action taken (CORE-G-007), D-16 and B.013 ---

func TestEvaluate_PeaceableKingdom(t *testing.T) {
	t.Run("a peaceful action counts", func(t *testing.T) {
		fixture := newGoalFixture(t, "Peaceable Kingdom", 0)

		got := fixture.evaluate(domain.ActionTaken{FactionID: "f1", Action: "Buy Asset"})

		fixture.want(got, fixture.progressed(1, causeProgress))
	})
	t.Run("an Attack resets progress", func(t *testing.T) {
		fixture := newGoalFixture(t, "Peaceable Kingdom", 2)

		got := fixture.evaluate(domain.ActionTaken{FactionID: "f1", Action: "Attack"})

		fixture.want(got, fixture.progressed(-2, causeReset))
	})
	t.Run("an Attack at no progress gives nothing", func(t *testing.T) {
		fixture := newGoalFixture(t, "Peaceable Kingdom", 0)

		got := fixture.evaluate(domain.ActionTaken{FactionID: "f1", Action: "Attack"})

		fixture.want(got)
	})
	t.Run("a clean miss still resets", func(t *testing.T) {
		fixture := newGoalFixture(t, "Peaceable Kingdom", 3)

		got := fixture.evaluate(domain.ActionTaken{FactionID: "f1", Action: "Attack"})

		fixture.want(got, fixture.progressed(-3, causeReset))
	})
	t.Run("a rival's action gives nothing", func(t *testing.T) {
		fixture := newGoalFixture(t, "Peaceable Kingdom", 0)

		got := fixture.evaluate(domain.ActionTaken{FactionID: "f2", Action: "Buy Asset"})

		fixture.want(got)
	})
	t.Run("an empty batch gives nothing", func(t *testing.T) {
		fixture := newGoalFixture(t, "Peaceable Kingdom", 2)

		got := fixture.evaluate()

		fixture.want(got)
	})
	t.Run("the fourth peaceful turn completes", func(t *testing.T) {
		fixture := newGoalFixture(t, "Peaceable Kingdom", 3)

		got := fixture.evaluate(domain.ActionTaken{FactionID: "f1", Action: "Buy Asset"})

		fixture.want(got, append([]domain.Mutation{fixture.progressed(1, causeProgress)}, fixture.completed(1)...)...)
	})
	t.Run("abandoning the goal neither moves nor completes it", func(t *testing.T) {
		fixture := newGoalFixture(t, "Peaceable Kingdom", 3)

		got := fixture.evaluate(
			domain.ActionTaken{FactionID: "f1", Action: "Abandon Goal"},
			domain.GoalAbandoned{FactionID: "f1", GoalID: fixture.compiled.goal.ID, Cause: "player_choice", CausedByFactionID: "f1"},
		)

		fixture.want(got)
	})
	t.Run("a rival abandoning its goal does not stop the holder's", func(t *testing.T) {
		fixture := newGoalFixture(t, "Peaceable Kingdom", 0)

		got := fixture.evaluate(
			domain.ActionTaken{FactionID: "f1", Action: "Buy Asset"},
			domain.GoalAbandoned{FactionID: "f2", GoalID: "CORE-G-001", Cause: "player_choice", CausedByFactionID: "f2"},
		)

		fixture.want(got, fixture.progressed(1, causeProgress))
	})
}

// --- Government taken (CORE-G-004), D-8 and D-9 ---

func TestEvaluate_WorldSeizure(t *testing.T) {
	seized := func(fromID, toID string, resisted bool) domain.WorldGovernmentChanged {
		return domain.WorldGovernmentChanged{WorldID: "Krylos", FromFactionID: fromID, ToFactionID: toID, Resisted: resisted}
	}
	t.Run("a resisting ruler pays a sixth of its ratings, rounded up", func(t *testing.T) {
		fixture := newGoalFixture(t, "World Seizure", 0)

		got := fixture.evaluate(seized("f2", "f1", true))

		fixture.want(got, fixture.completed(2)...)
	})
	t.Run("an unresisted seizure pays 1", func(t *testing.T) {
		fixture := newGoalFixture(t, "World Seizure", 0)

		got := fixture.evaluate(seized("f2", "f1", false))

		fixture.want(got, fixture.completed(1)...)
	})
	t.Run("an independent world pays 1", func(t *testing.T) {
		fixture := newGoalFixture(t, "World Seizure", 0)

		got := fixture.evaluate(seized("", "f1", false))

		fixture.want(got, fixture.completed(1)...)
	})
	t.Run("a weak resisting ruler pays 1", func(t *testing.T) {
		fixture := newGoalFixture(t, "World Seizure", 0)
		fixture.rival.Force, fixture.rival.Cunning, fixture.rival.Wealth = 1, 1, 1

		got := fixture.evaluate(seized("f2", "f1", true))

		fixture.want(got, fixture.completed(1)...)
	})
	t.Run("a world taken by a rival gives nothing", func(t *testing.T) {
		fixture := newGoalFixture(t, "World Seizure", 0)

		got := fixture.evaluate(seized("f3", "f2", true))

		fixture.want(got)
	})
}

// --- Faction destroyed (CORE-G-008), D-9 and D-15 ---

func hpLoss(factionID string, delta int, causedByID string) domain.FactionHPDelta {
	return domain.FactionHPDelta{FactionID: factionID, Delta: delta, Cause: "attack", CausedByFactionID: causedByID}
}

func TestEvaluate_DestroyTheFoe(t *testing.T) {
	cases := []struct {
		name       string
		targetID   string
		batch      []domain.Mutation
		wantXP     int
		wantFailed bool
	}{
		{"the holder's blow destroys the target", "f2", []domain.Mutation{hpLoss("f2", -5, "f1")}, 5, false},
		{"the holder's last blow after a third party's", "f2", []domain.Mutation{hpLoss("f2", -3, "f3"), hpLoss("f2", -2, "f1")}, 5, false},
		{"a third party's last blow fails the goal", "f2", []domain.Mutation{hpLoss("f2", -3, "f1"), hpLoss("f2", -2, "f3")}, 0, true},
		{"a third party's only blow fails the goal", "f2", []domain.Mutation{hpLoss("f2", -5, "f3")}, 0, true},
		{"a third party's loss that leaves the target at 1 HP", "f2", []domain.Mutation{hpLoss("f2", -4, "f3")}, 0, false},
		{"the destruction of a faction not targeted", "f2", []domain.Mutation{hpLoss("f3", -2, "f1")}, 0, false},
		{"no target set", "", []domain.Mutation{hpLoss("f2", -5, "f1")}, 0, false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newGoalFixture(t, "Destroy the Foe", 0)
			fixture.holder.ActiveGoal.TargetFactionID = testCase.targetID

			got := fixture.evaluate(testCase.batch...)

			switch {
			case testCase.wantFailed:
				fixture.want(got, domain.GoalFailed{FactionID: "f1", GoalID: fixture.compiled.goal.ID, Cause: domain.GoalFailedTargetLost})
			case testCase.wantXP > 0:
				fixture.want(got, fixture.completed(testCase.wantXP)...)
			default:
				fixture.want(got)
			}
		})
	}
}

func TestEvaluate_DestroyTheFoe_TargetAlreadyLost(t *testing.T) {
	cases := []struct {
		name  string
		setup func(fixture *goalFixture)
	}{
		{"a target destroyed before the batch", func(fixture *goalFixture) { fixture.rival.CurrentHP = 0 }},
		{"a target gone from state", func(fixture *goalFixture) { delete(fixture.factionState.Factions, "f2") }},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newGoalFixture(t, "Destroy the Foe", 0)
			fixture.holder.ActiveGoal.TargetFactionID = "f2"
			testCase.setup(fixture)

			got := fixture.evaluate()

			fixture.want(got, domain.GoalFailed{FactionID: "f1", GoalID: fixture.compiled.goal.ID, Cause: domain.GoalFailedTargetLost})
		})
	}
}

// targetedGoal is a shipped goal given a target_faction input, held by f1 with
// f2 as its target: a GM-authored goal shape no shipped goal has.
func targetedGoal(t *testing.T, goalName string, progress int) *goalFixture {
	t.Helper()
	fixture := newGoalFixture(t, goalName, progress)
	goal := *fixture.compiled.goal
	start := domain.GoalStart{}
	if goal.Start != nil {
		start = *goal.Start
	}
	start.Inputs = append(slices.Clone(start.Inputs), inputTargetFaction)
	goal.Start = &start
	compiled, err := compile(&goal, testActionNames)
	if err != nil {
		t.Fatalf("compile targeted %s: %v", goalName, err)
	}
	fixture.compiled = compiled
	fixture.holder.ActiveGoal.TargetFactionID = "f2"
	return fixture
}

func TestEvaluate_LostTarget_ReplacesTheBatchProgress(t *testing.T) {
	failed := func(fixture *goalFixture) domain.GoalFailed {
		return domain.GoalFailed{FactionID: "f1", GoalID: fixture.compiled.goal.ID, Cause: domain.GoalFailedTargetLost}
	}
	t.Run("a batch that also moves progress", func(t *testing.T) {
		fixture := targetedGoal(t, "Peaceable Kingdom", 0)

		got := fixture.evaluate(domain.ActionTaken{FactionID: "f1", Action: "Buy Asset"}, hpLoss("f2", -5, "f3"))

		fixture.want(got, failed(fixture))
	})
	t.Run("a batch that also resets progress", func(t *testing.T) {
		fixture := targetedGoal(t, "Peaceable Kingdom", 2)

		got := fixture.evaluate(domain.ActionTaken{FactionID: "f1", Action: "Attack"}, hpLoss("f2", -5, "f3"))

		fixture.want(got, failed(fixture))
	})
	t.Run("the holder's own blow fails a goal it does not complete", func(t *testing.T) {
		fixture := targetedGoal(t, "Peaceable Kingdom", 0)

		got := fixture.evaluate(hpLoss("f2", -5, "f1"))

		fixture.want(got, failed(fixture))
	})
}

func TestEvaluate_LostTargetNeedsATargetInput(t *testing.T) {
	fixture := newGoalFixture(t, "Military Conquest", 0)
	fixture.holder.ActiveGoal.TargetFactionID = "f2"

	got := fixture.evaluate(hpLoss("f2", -5, "f3"))

	fixture.want(got)
}

// --- Stealthed assets (CORE-G-009), D-10 and Decision 2 ---

func (fixture *goalFixture) stealthed(assetID, worldID string) *domain.Asset {
	asset := fixture.give(fixture.holder, assetID, "cunning-def", worldID)
	asset.Stealthy = true
	return asset
}

func TestEvaluate_InsideEnemyTerritory_StealthInTheBatchCompletes(t *testing.T) {
	fixture := newGoalFixture(t, "Inside Enemy Territory", 2)
	fixture.stealthed("s1", "Krylos")
	fixture.stealthed("s2", "Krylos")
	fixture.give(fixture.holder, "s3", "cunning-def", "Krylos")

	got := fixture.evaluate(domain.AssetStealthApplied{FactionID: "f1", AssetID: "s3"})

	fixture.want(got, append([]domain.Mutation{fixture.progressed(1, causeProgress)}, fixture.completed(2)...)...)
}

func TestEvaluate_InsideEnemyTerritory_CountFollowsTheState(t *testing.T) {
	t.Run("stealth cleared in the batch", func(t *testing.T) {
		fixture := newGoalFixture(t, "Inside Enemy Territory", 2)
		fixture.stealthed("s1", "Krylos")
		fixture.stealthed("s2", "Krylos")

		got := fixture.evaluate(domain.AssetStealthCleared{FactionID: "f1", AssetID: "s2"})

		fixture.want(got, fixture.progressed(-1, causeProgress))
	})
	t.Run("an asset moved off rival ground in the batch", func(t *testing.T) {
		fixture := newGoalFixture(t, "Inside Enemy Territory", 2)
		fixture.stealthed("s1", "Krylos")
		fixture.stealthed("s2", "Krylos")

		got := fixture.evaluate(domain.AssetMoved{
			FactionID:    "f1",
			AssetID:      "s2",
			FromLocation: domain.Location{WorldID: "Krylos"},
			ToLocation:   domain.Location{WorldID: "Tartarus"},
		})

		fixture.want(got, fixture.progressed(-1, causeProgress))
	})
}

func TestEvaluate_InsideEnemyTerritory_CountsRivalGround(t *testing.T) {
	assetWorlds := map[string]string{
		"governed":         "Tartarus",
		"rival-governed":   "Krylos",
		"independent":      "Hadrian",
		"rival-homeworld":  "Anchorage",
		"seized-homeworld": "Vesta",
		"lost-homeworld":   "Brennus",
		"in-transit":       "",
	}
	fixture := newGoalFixture(t, "Inside Enemy Territory", 0)
	fixture.holder.Cunning = 5
	fixture.holder.Homeworld = domain.Location{WorldID: "Brennus"}
	fixture.rival.Homeworld = domain.Location{WorldID: "Anchorage"}
	fixture.third.Homeworld = domain.Location{WorldID: "Vesta"}
	fixture.factionState.Governments = map[string]string{"Tartarus": "f1", "Krylos": "f2", "Brennus": "f2", "Vesta": "f1"}
	var batch []domain.Mutation
	for assetID, worldID := range assetWorlds {
		fixture.give(fixture.holder, assetID, "cunning-def", worldID)
		batch = append(batch, domain.AssetStealthApplied{FactionID: "f1", AssetID: assetID})
	}

	got := fixture.evaluate(batch...)

	// Krylos, Anchorage and Brennus count; at Cunning 5 the goal does not complete.
	fixture.want(got, fixture.progressed(3, causeProgress))
}

func TestEvaluate_InsideEnemyTerritory_DestroyedRivalIsNoRivalGround(t *testing.T) {
	fixture := newGoalFixture(t, "Inside Enemy Territory", 0)
	fixture.third.CurrentHP = 0
	fixture.factionState.Governments["Ossuary"] = "f3"
	var batch []domain.Mutation
	for assetID, worldID := range map[string]string{"dead-governed": "Ossuary", "dead-homeworld": "Vesta", "live-governed": "Krylos"} {
		fixture.give(fixture.holder, assetID, "cunning-def", worldID)
		batch = append(batch, domain.AssetStealthApplied{FactionID: "f1", AssetID: assetID})
	}

	got := fixture.evaluate(batch...)

	// Only Krylos counts: f3 is destroyed, so neither its world nor its homeworld is rival ground.
	fixture.want(got, fixture.progressed(1, causeProgress))
}

func TestEvaluate_InsideEnemyTerritory_HeldAtStartNeverCounts(t *testing.T) {
	fixture := newGoalFixture(t, "Inside Enemy Territory", 0)
	fixture.stealthed("held", "Krylos")
	fixture.holder.ActiveGoal.StealthedAtStart = []string{"held"}
	fixture.give(fixture.holder, "new", "cunning-def", "Krylos")

	got := fixture.evaluate(domain.AssetStealthApplied{FactionID: "f1", AssetID: "new"})

	fixture.want(got, fixture.progressed(1, causeProgress))
}

func TestEvaluate_InsideEnemyTerritory_LeavesLiveStateUntouched(t *testing.T) {
	fixture := newGoalFixture(t, "Inside Enemy Territory", 0)
	live := fixture.give(fixture.holder, "s1", "cunning-def", "Krylos")

	fixture.evaluate(domain.AssetStealthApplied{FactionID: "f1", AssetID: "s1"})

	if live.Stealthy {
		t.Error("the live asset is stealthed; the measure must apply the batch to a copy")
	}
}

// --- Order of output ---

func TestEvaluate_CompletionOrder(t *testing.T) {
	fixture := newGoalFixture(t, "Blood the Enemy", 7)

	got := fixture.evaluate(domain.AssetHPDelta{FactionID: "f2", AssetID: "a1", Delta: -5, CausedByFactionID: "f1"})

	if len(got) != 3 {
		t.Fatalf("mutations = %+v, want three", got)
	}
	if _, ok := got[0].(domain.GoalProgressed); !ok {
		t.Errorf("first = %T, want GoalProgressed", got[0])
	}
	completed, ok := got[1].(domain.GoalCompleted)
	if !ok {
		t.Fatalf("second = %T, want GoalCompleted", got[1])
	}
	awarded, ok := got[2].(domain.XPAwarded)
	if !ok {
		t.Fatalf("third = %T, want XPAwarded", got[2])
	}
	if completed.XPAwarded != awarded.Amount {
		t.Errorf("GoalCompleted.XPAwarded = %d, XPAwarded.Amount = %d, want equal", completed.XPAwarded, awarded.Amount)
	}
}

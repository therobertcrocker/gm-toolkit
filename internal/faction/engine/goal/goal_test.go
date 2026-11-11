package goal

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func discardLog() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func newEngine(t *testing.T, loaded *rulebook.Rulebook) *GoalEngine {
	t.Helper()
	goalEngine, err := New(loaded, testActionNames, discardLog())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return goalEngine
}

func TestNew_CompilesTheShippedGoals(t *testing.T) {
	goalEngine := newEngine(t, shippedRulebook(t))

	if len(goalEngine.goals) != 11 {
		t.Errorf("compiled goals = %d, want 11", len(goalEngine.goals))
	}
}

func TestNew_RefusesABadGoal(t *testing.T) {
	badGoal := validThresholdGoal()
	badGoal.ID = "X-1"
	badGoal.XP = nil
	loaded := &rulebook.Rulebook{Goals: map[string]*domain.Goal{"X-1": badGoal}}

	_, err := New(loaded, testActionNames, discardLog())

	if err == nil {
		t.Fatal("New returned no error for a goal with no xp part")
	}
	for _, want := range []string{`goal "X-1"`, "needs an xp part"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err, want)
		}
	}
}

func TestUpdateProgress_NoActiveGoal(t *testing.T) {
	loaded := shippedRulebook(t)
	goalEngine := newEngine(t, loaded)
	faction := &domain.Faction{ID: "f1"}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	mutations := goalEngine.UpdateProgress(nil, factionState, discardLog())

	if mutations != nil {
		t.Errorf("mutations = %v, want nil", mutations)
	}
}

func TestUpdateProgress_UnknownGoalIsSkipped(t *testing.T) {
	loaded := shippedRulebook(t)
	goalEngine := newEngine(t, loaded)
	faction := &domain.Faction{ID: "f1", ActiveGoal: &domain.ActiveGoal{GoalID: "CORE-G-012"}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	mutations := goalEngine.UpdateProgress(nil, factionState, discardLog())

	if mutations != nil {
		t.Errorf("mutations = %v, want nil", mutations)
	}
}

func TestRuns(t *testing.T) {
	goalEngine := newEngine(t, shippedRulebook(t))
	cases := map[string]struct {
		heldGoal *domain.ActiveGoal
		want     bool
	}{
		"a known goal with no input":        {&domain.ActiveGoal{GoalID: "CORE-G-001"}, true},
		"a targeted goal with a target":     {&domain.ActiveGoal{GoalID: "CORE-G-008", TargetFactionID: "f2"}, true},
		"a targeted goal with no target":    {&domain.ActiveGoal{GoalID: "CORE-G-008"}, false},
		"a goal the rulebook does not know": {&domain.ActiveGoal{GoalID: "G-004"}, false},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			if got := goalEngine.Runs(testCase.heldGoal); got != testCase.want {
				t.Errorf("Runs(%+v) = %v, want %v", testCase.heldGoal, got, testCase.want)
			}
		})
	}
}

func TestUpdateProgress_RunsTheHeldGoal(t *testing.T) {
	fixture := newGoalFixture(t, "Military Conquest", 0)
	fixture.give(fixture.rival, "a1", "force-def", "Krylos")
	goalEngine := newEngine(t, fixture.rulebook)

	got := goalEngine.UpdateProgress([]domain.Mutation{kill("f2", "a1", "attack", "f1")}, fixture.factionState, discardLog())

	fixture.want(got, fixture.progressed(1, causeProgress))
}

func TestUpdateProgress_SkipsADestroyedHolder(t *testing.T) {
	fixture := newGoalFixture(t, "Military Conquest", 0)
	fixture.holder.CurrentHP = 0
	fixture.give(fixture.rival, "a1", "force-def", "Krylos")
	goalEngine := newEngine(t, fixture.rulebook)

	got := goalEngine.UpdateProgress([]domain.Mutation{kill("f2", "a1", "counterstrike", "f1")}, fixture.factionState, discardLog())

	fixture.want(got)
}

func TestUpdateProgress_ReadsEveryHolder(t *testing.T) {
	fixture := newGoalFixture(t, "Military Conquest", 0)
	bloodTheEnemy := shippedGoal(t, fixture.rulebook, "Blood the Enemy").goal.ID
	fixture.rival.ActiveGoal = &domain.ActiveGoal{GoalID: bloodTheEnemy}
	fixture.give(fixture.rival, "a1", "force-def", "Krylos")
	fixture.give(fixture.holder, "b1", "force-def", "Krylos").CurrentHP = 5
	goalEngine := newEngine(t, fixture.rulebook)
	batch := []domain.Mutation{
		domain.AssetHPDelta{FactionID: "f1", AssetID: "b1", Delta: -2, Cause: "counterstrike", CausedByFactionID: "f2"},
		kill("f2", "a1", "attack", "f1"),
	}

	got := goalEngine.UpdateProgress(batch, fixture.factionState, discardLog())

	fixture.want(got,
		fixture.progressed(1, causeProgress),
		domain.GoalProgressed{FactionID: "f2", GoalID: bloodTheEnemy, Delta: 2, Cause: causeProgress},
	)
}

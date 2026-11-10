package goal

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

func validThresholdGoal() *domain.Goal {
	return &domain.Goal{
		ID:        "G-TEST",
		Name:      "Test Goal",
		Progress:  &domain.GoalProgress{Event: "kill", Amount: amountOne},
		Threshold: &domain.GoalThreshold{Constant: 2},
		XP:        &domain.GoalXP{Formula: formulaConstant, Amount: 1},
	}
}

// completeOn turns the valid threshold goal into a complete goal on event.
func completeOn(goal *domain.Goal, event domain.GoalEvent) {
	goal.Progress = nil
	goal.Threshold = nil
	goal.Complete = &event
}

func TestCompile_Refuses(t *testing.T) {
	cases := []struct {
		name    string
		change  func(goal *domain.Goal)
		wantErr string
	}{
		{"neither end", func(goal *domain.Goal) { goal.Threshold = nil }, "exactly one of threshold or complete"},
		{"both ends", func(goal *domain.Goal) { goal.Complete = &domain.GoalEvent{Event: "kill"} }, "exactly one of threshold or complete"},
		{"threshold with no progress", func(goal *domain.Goal) { goal.Progress = nil }, "threshold needs a progress part"},
		{"reset with no progress", func(goal *domain.Goal) {
			completeOn(goal, domain.GoalEvent{Event: "kill"})
			goal.Reset = &domain.GoalEvent{Event: "action_taken"}
		}, "reset needs a progress part"},
		{"event and measure", func(goal *domain.Goal) { goal.Progress.Measure = "stealthed_assets" }, "exactly one of event or measure"},
		{"neither event nor measure", func(goal *domain.Goal) { goal.Progress = &domain.GoalProgress{} }, "exactly one of event or measure"},
		{"unknown progress event", func(goal *domain.Goal) { goal.Progress.Event = "kil" }, `progress: unknown event "kil"`},
		{"unknown reset event", func(goal *domain.Goal) { goal.Reset = &domain.GoalEvent{Event: "nap"} }, "reset: unknown event"},
		{"unknown complete event", func(goal *domain.Goal) { completeOn(goal, domain.GoalEvent{Event: "nap"}) }, "complete: unknown event"},
		{"wrong amount", func(goal *domain.Goal) { goal.Progress.Amount = amountDamage }, `does not accept amount "damage"`},
		{"missing amount", func(goal *domain.Goal) { goal.Progress.Amount = "" }, `does not accept amount ""`},
		{"unknown measure", func(goal *domain.Goal) { goal.Progress = &domain.GoalProgress{Measure: "assets"} }, `unknown measure "assets"`},
		{"measure with amount", func(goal *domain.Goal) {
			goal.Progress = &domain.GoalProgress{Measure: "stealthed_assets", Amount: amountOne}
		}, "takes no amount"},
		{"filter not accepted by event", func(goal *domain.Goal) {
			goal.Progress = &domain.GoalProgress{Event: "damage", Amount: amountDamage, Filters: domain.GoalFilters{Category: []domain.FactionStat{domain.StatForce}}}
		}, `filter "category" is not accepted here`},
		{"filter not accepted by measure", func(goal *domain.Goal) {
			goal.Progress = &domain.GoalProgress{Measure: "stealthed_assets", Filters: domain.GoalFilters{Owner: sideOwn}}
		}, `filter "owner" is not accepted here`},
		{"bad owner value", func(goal *domain.Goal) { goal.Progress.Filters.Owner = "enemy" }, `filter owner = "enemy"`},
		{"bad caused_by value", func(goal *domain.Goal) { goal.Progress.Filters.CausedBy = "enemy" }, "filter caused_by ="},
		{"bad faction value", func(goal *domain.Goal) {
			goal.Progress = &domain.GoalProgress{Event: "action_taken", Amount: amountOne, Filters: domain.GoalFilters{Faction: sideRival}}
		}, `filter faction = "rival"`},
		{"bad to value", func(goal *domain.Goal) {
			completeOn(goal, domain.GoalEvent{Event: eventGovernmentTaken, Filters: domain.GoalFilters{To: sideRival}})
		}, "filter to ="},
		{"bad target value", func(goal *domain.Goal) {
			completeOn(goal, domain.GoalEvent{Event: "faction_destroyed", Filters: domain.GoalFilters{Target: "enemy"}})
		}, "filter target ="},
		{"bad ground value", func(goal *domain.Goal) {
			goal.Progress = &domain.GoalProgress{Measure: "stealthed_assets", Filters: domain.GoalFilters{Ground: sideOwn}}
		}, `filter ground = "own"`},
		{"target with no start input", func(goal *domain.Goal) {
			completeOn(goal, domain.GoalEvent{Event: "faction_destroyed", Filters: domain.GoalFilters{Target: targetStartInput}})
		}, `needs start.inputs to name "target_faction"`},
		{"unknown start input", func(goal *domain.Goal) { goal.Start = &domain.GoalStart{Inputs: []string{"target_world"}} }, `start: unknown input "target_world"`},
		{"threshold with both", func(goal *domain.Goal) {
			goal.Threshold = &domain.GoalThreshold{Ratings: []domain.FactionStat{domain.StatForce}, Constant: 2}
		}, "exactly one of ratings or constant"},
		{"threshold with neither", func(goal *domain.Goal) { goal.Threshold = &domain.GoalThreshold{} }, "exactly one of ratings or constant"},
		{"negative constant", func(goal *domain.Goal) { goal.Threshold = &domain.GoalThreshold{Constant: -1} }, "must be >= 0"},
		{"multiplier with constant", func(goal *domain.Goal) { goal.Threshold = &domain.GoalThreshold{Constant: 2, Multiplier: 3} }, "multiplier needs ratings"},
		{"no xp", func(goal *domain.Goal) { goal.XP = nil }, "needs an xp part"},
		{"unknown formula", func(goal *domain.Goal) { goal.XP = &domain.GoalXP{Formula: "double"} }, `xp: unknown formula "double"`},
		{"constant with no amount", func(goal *domain.Goal) { goal.XP = &domain.GoalXP{Formula: formulaConstant} }, "constant needs amount > 0"},
		{"amount on another formula", func(goal *domain.Goal) {
			goal.XP = &domain.GoalXP{Formula: formulaHalfProgress, Amount: 1}
		}, "takes no amount or contested_bonus"},
		{"contested_bonus on a threshold goal", func(goal *domain.Goal) {
			goal.XP = &domain.GoalXP{Formula: formulaConstant, Amount: 1, ContestedBonus: 1}
		}, "contested_bonus needs a complete event that names a world"},
		{"contested_bonus on an event with no world", func(goal *domain.Goal) {
			completeOn(goal, domain.GoalEvent{Event: "faction_destroyed"})
			goal.XP = &domain.GoalXP{Formula: formulaConstant, Amount: 1, ContestedBonus: 1}
		}, "contested_bonus needs a complete event that names a world"},
		{"half_progress with no threshold", func(goal *domain.Goal) {
			completeOn(goal, domain.GoalEvent{Event: "kill"})
			goal.XP = &domain.GoalXP{Formula: formulaHalfProgress}
		}, "half_progress needs a threshold"},
		{"target_average with no input", func(goal *domain.Goal) { goal.XP = &domain.GoalXP{Formula: formulaTargetAverage} }, "target_average needs start.inputs"},
		{"ruler_average on another event", func(goal *domain.Goal) {
			completeOn(goal, domain.GoalEvent{Event: "kill"})
			goal.XP = &domain.GoalXP{Formula: formulaRulerAverage}
		}, "ruler_average needs complete.event"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			goal := validThresholdGoal()
			testCase.change(goal)

			_, err := compile(goal)

			if err == nil || !strings.Contains(err.Error(), testCase.wantErr) {
				t.Errorf("compile error = %v, want it to contain %q", err, testCase.wantErr)
			}
		})
	}
}

func TestCompile_AcceptsValidGoal(t *testing.T) {
	if _, err := compile(validThresholdGoal()); err != nil {
		t.Fatalf("compile: %v", err)
	}
}

func TestCompile_AcceptsShippedGoals(t *testing.T) {
	loaded := shippedRulebook(t)
	if len(loaded.Goals) == 0 {
		t.Fatal("the shipped rulebook has no goals")
	}
	for id, goal := range loaded.Goals {
		if _, err := compile(goal); err != nil {
			t.Errorf("compile %s (%s): %v", id, goal.Name, err)
		}
	}
}

func shippedRulebook(t *testing.T) *rulebook.Rulebook {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(file), "..", "..", "..", "..")
	camp := &campaigns.Campaign{Root: t.TempDir()}
	if _, err := campaigns.ComposeRulebook(camp, filepath.Join(repoRoot, "rulebooks", "swn"), false); err != nil {
		t.Fatalf("compose rulebook: %v", err)
	}
	loaded, err := rulebook.Load(camp.Paths().FactionDataDir)
	if err != nil {
		t.Fatalf("rulebook.Load: %v", err)
	}
	return loaded
}

func shippedGoal(t *testing.T, loaded *rulebook.Rulebook, name string) *compiledGoal {
	t.Helper()
	for _, goal := range loaded.Goals {
		if goal.Name == name {
			compiled, err := compile(goal)
			if err != nil {
				t.Fatalf("compile %s: %v", name, err)
			}
			return compiled
		}
	}
	t.Fatalf("no shipped goal named %q", name)
	return nil
}

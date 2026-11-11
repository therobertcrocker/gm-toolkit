package rulebook

import (
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func TestLoad_ShippedGoals(t *testing.T) {
	rulebook := loadShippedRulebook(t)
	if len(rulebook.Goals) != 11 {
		t.Fatalf("goals = %d, want 11", len(rulebook.Goals))
	}

	cases := []struct {
		goalID string
		check  func(t *testing.T, goal *domain.Goal)
	}{
		{"CORE-G-004", func(t *testing.T, goal *domain.Goal) {
			if goal.Start == nil || !slices.Equal(goal.Start.NeedsActions, []string{"Seize Planet"}) {
				t.Errorf("start = %+v, want needs_actions [Seize Planet]", goal.Start)
			}
			if goal.Complete == nil || goal.Complete.Event != "government_taken" || goal.Complete.Filters.To != "own" {
				t.Errorf("complete = %+v, want government_taken with to = own", goal.Complete)
			}
			if goal.Progress != nil || goal.Threshold != nil {
				t.Errorf("progress, threshold = %+v, %+v, want both nil", goal.Progress, goal.Threshold)
			}
		}},
		{"CORE-G-007", func(t *testing.T, goal *domain.Goal) {
			if goal.Reset == nil || goal.Reset.Event != "action_taken" || !slices.Equal(goal.Reset.Filters.Action, []string{"Attack"}) {
				t.Errorf("reset = %+v, want action_taken with action [Attack]", goal.Reset)
			}
			if goal.Threshold == nil || goal.Threshold.Constant != 4 {
				t.Errorf("threshold = %+v, want constant 4", goal.Threshold)
			}
		}},
		{"CORE-G-008", func(t *testing.T, goal *domain.Goal) {
			if goal.Start == nil || !slices.Equal(goal.Start.Inputs, []string{"target_faction"}) {
				t.Errorf("start = %+v, want inputs [target_faction]", goal.Start)
			}
		}},
		{"CORE-G-009", func(t *testing.T, goal *domain.Goal) {
			if goal.Progress == nil || goal.Progress.Measure != "stealthed_assets" {
				t.Fatalf("progress = %+v, want measure stealthed_assets", goal.Progress)
			}
			if goal.Progress.Filters.Ground != "rival" || !goal.Progress.Filters.ExcludeHeldAtStart {
				t.Errorf("progress filters = %+v, want ground = rival and exclude_held_at_start", goal.Progress.Filters)
			}
		}},
		{"CORE-G-010", func(t *testing.T, goal *domain.Goal) {
			if goal.Complete == nil || goal.Complete.Filters.RatedAbove != domain.StatForce {
				t.Errorf("complete = %+v, want rated_above Force", goal.Complete)
			}
		}},
		{"CORE-G-011", func(t *testing.T, goal *domain.Goal) {
			if goal.Start == nil || goal.Start.Gate == nil || goal.Start.Gate.RatingAboveLastCompletion != domain.StatWealth {
				t.Errorf("start = %+v, want gate rating_above_last_completion Wealth", goal.Start)
			}
			if goal.Threshold == nil || goal.Threshold.Multiplier != 4 {
				t.Errorf("threshold = %+v, want multiplier 4", goal.Threshold)
			}
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.goalID, func(t *testing.T) {
			goal, ok := rulebook.Goals[testCase.goalID]
			if !ok {
				t.Fatalf("goal %s not loaded", testCase.goalID)
			}
			testCase.check(t, goal)
		})
	}
}

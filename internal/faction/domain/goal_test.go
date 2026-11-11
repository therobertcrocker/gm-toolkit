package domain_test

import (
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func TestGoalFilters_Used(t *testing.T) {
	every := domain.GoalFilters{
		Owner:              "rival",
		CausedBy:           "own",
		Faction:            "own",
		Category:           []domain.FactionStat{domain.StatForce},
		RatedAbove:         domain.StatForce,
		NewWorld:           true,
		To:                 "own",
		Target:             "start_input",
		Action:             []string{"Attack"},
		ActionNot:          []string{"Attack"},
		Ground:             "rival",
		ExcludeHeldAtStart: true,
	}
	wantEvery := []string{
		domain.GoalFilterOwner,
		domain.GoalFilterCausedBy,
		domain.GoalFilterFaction,
		domain.GoalFilterCategory,
		domain.GoalFilterRatedAbove,
		domain.GoalFilterNewWorld,
		domain.GoalFilterTo,
		domain.GoalFilterTarget,
		domain.GoalFilterAction,
		domain.GoalFilterActionNot,
		domain.GoalFilterGround,
		domain.GoalFilterExcludeHeldAtStart,
	}
	if got := every.Used(); !slices.Equal(got, wantEvery) {
		t.Errorf("every filter set: Used() = %v, want %v", got, wantEvery)
	}

	if got := (domain.GoalFilters{}).Used(); len(got) != 0 {
		t.Errorf("zero value: Used() = %v, want none", got)
	}

	falseBools := domain.GoalFilters{NewWorld: false, ExcludeHeldAtStart: false}
	if got := falseBools.Used(); len(got) != 0 {
		t.Errorf("false bools: Used() = %v, want none", got)
	}
}

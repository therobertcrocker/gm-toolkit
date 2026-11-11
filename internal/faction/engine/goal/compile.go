package goal

import (
	"fmt"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

type compiledGoal struct {
	goal     *domain.Goal
	progress *eventSpec
	measure  *measureSpec
	reset    *eventSpec
	complete *eventSpec
}

// compile checks every name and combination in goal's parts and binds them.
// actionNames are the registered actions an action filter may name. Its errors do
// not name the goal; the caller does.
func compile(goal *domain.Goal, actionNames []string) (*compiledGoal, error) {
	compiled := &compiledGoal{goal: goal}
	if err := checkStart(goal.Start); err != nil {
		return nil, err
	}
	if (goal.Threshold == nil) == (goal.Complete == nil) {
		return nil, fmt.Errorf("a goal needs exactly one of threshold or complete")
	}
	if goal.Threshold != nil {
		if goal.Progress == nil {
			return nil, fmt.Errorf("threshold needs a progress part")
		}
		if err := checkThreshold(goal.Threshold); err != nil {
			return nil, err
		}
	}
	if goal.Reset != nil && goal.Progress == nil {
		return nil, fmt.Errorf("reset needs a progress part")
	}
	if progress := goal.Progress; progress != nil {
		if (progress.Event == "") == (progress.Measure == "") {
			return nil, fmt.Errorf("progress needs exactly one of event or measure")
		}
		if progress.Event != "" {
			spec, err := eventNamed("progress", progress.Event, progress.Filters, goal, actionNames)
			if err != nil {
				return nil, err
			}
			if !slices.Contains(spec.amounts, progress.Amount) {
				return nil, fmt.Errorf("progress: event %q does not accept amount %q (want one of %v)", progress.Event, progress.Amount, spec.amounts)
			}
			compiled.progress = spec
		} else {
			spec, ok := measures[progress.Measure]
			if !ok {
				return nil, fmt.Errorf("progress: unknown measure %q", progress.Measure)
			}
			if progress.Amount != "" {
				return nil, fmt.Errorf("progress: measure %q takes no amount", progress.Measure)
			}
			if err := checkFilters("progress", spec.filters, progress.Filters, goal, actionNames); err != nil {
				return nil, err
			}
			compiled.measure = &spec
		}
	}
	if goal.Reset != nil {
		spec, err := eventNamed("reset", goal.Reset.Event, goal.Reset.Filters, goal, actionNames)
		if err != nil {
			return nil, err
		}
		compiled.reset = spec
	}
	if goal.Complete != nil {
		spec, err := eventNamed("complete", goal.Complete.Event, goal.Complete.Filters, goal, actionNames)
		if err != nil {
			return nil, err
		}
		compiled.complete = spec
	}
	if err := checkXP(goal); err != nil {
		return nil, err
	}
	return compiled, nil
}

func eventNamed(part, name string, filters domain.GoalFilters, goal *domain.Goal, actionNames []string) (*eventSpec, error) {
	spec, ok := events[name]
	if !ok {
		return nil, fmt.Errorf("%s: unknown event %q", part, name)
	}
	if err := checkFilters(part, spec.filters, filters, goal, actionNames); err != nil {
		return nil, err
	}
	return &spec, nil
}

func checkFilters(part string, accepted []string, filters domain.GoalFilters, goal *domain.Goal, actionNames []string) error {
	for _, name := range filters.Used() {
		if !slices.Contains(accepted, name) {
			return fmt.Errorf("%s: filter %q is not accepted here (want one of %v)", part, name, accepted)
		}
	}
	values := []struct {
		name    string
		value   string
		allowed []string
	}{
		{domain.GoalFilterOwner, filters.Owner, []string{sideOwn, sideRival}},
		{domain.GoalFilterCausedBy, filters.CausedBy, []string{sideOwn, sideRival}},
		{domain.GoalFilterFaction, filters.Faction, []string{sideOwn}},
		{domain.GoalFilterTo, filters.To, []string{sideOwn}},
		{domain.GoalFilterTarget, filters.Target, []string{targetStartInput}},
		{domain.GoalFilterGround, filters.Ground, []string{groundRival}},
	}
	for _, check := range values {
		if check.value != "" && !slices.Contains(check.allowed, check.value) {
			return fmt.Errorf("%s: filter %s = %q (want one of %v)", part, check.name, check.value, check.allowed)
		}
	}
	if filters.Target == targetStartInput && !hasInput(goal, inputTargetFaction) {
		return fmt.Errorf("%s: filter target = %q needs start.inputs to name %q", part, targetStartInput, inputTargetFaction)
	}
	for _, action := range slices.Concat(filters.Action, filters.ActionNot) {
		if !slices.Contains(actionNames, action) {
			return fmt.Errorf("%s: unknown action %q (want one of %v)", part, action, actionNames)
		}
	}
	return nil
}

func checkStart(start *domain.GoalStart) error {
	if start == nil {
		return nil
	}
	for _, input := range start.Inputs {
		if input != inputTargetFaction {
			return fmt.Errorf("start: unknown input %q (want %q)", input, inputTargetFaction)
		}
	}
	return nil
}

func hasInput(goal *domain.Goal, input string) bool {
	return goal.Start != nil && slices.Contains(goal.Start.Inputs, input)
}

func checkThreshold(threshold *domain.GoalThreshold) error {
	if threshold.Constant < 0 || threshold.Multiplier < 0 {
		return fmt.Errorf("threshold: constant and multiplier must be >= 0, got %d and %d", threshold.Constant, threshold.Multiplier)
	}
	if (len(threshold.Ratings) > 0) == (threshold.Constant > 0) {
		return fmt.Errorf("threshold needs exactly one of ratings or constant")
	}
	if threshold.Multiplier > 0 && len(threshold.Ratings) == 0 {
		return fmt.Errorf("threshold: multiplier needs ratings")
	}
	return nil
}

func checkXP(goal *domain.Goal) error {
	xp := goal.XP
	if xp == nil {
		return fmt.Errorf("a goal needs an xp part")
	}
	switch xp.Formula {
	case formulaConstant:
		if xp.Amount <= 0 || xp.ContestedBonus < 0 {
			return fmt.Errorf("xp: constant needs amount > 0 and contested_bonus >= 0, got %d and %d", xp.Amount, xp.ContestedBonus)
		}
		if xp.ContestedBonus > 0 && (goal.Complete == nil || !events[goal.Complete.Event].hasWorld) {
			return fmt.Errorf("xp: contested_bonus needs a complete event that names a world")
		}
	case formulaHalfProgress:
		if goal.Threshold == nil {
			return fmt.Errorf("xp: %s needs a threshold", formulaHalfProgress)
		}
	case formulaTargetAverage:
		if !hasInput(goal, inputTargetFaction) {
			return fmt.Errorf("xp: %s needs start.inputs to name %q", formulaTargetAverage, inputTargetFaction)
		}
	case formulaRulerAverage:
		if goal.Complete == nil || goal.Complete.Event != eventGovernmentTaken {
			return fmt.Errorf("xp: %s needs complete.event = %q", formulaRulerAverage, eventGovernmentTaken)
		}
	default:
		return fmt.Errorf("xp: unknown formula %q", xp.Formula)
	}
	if xp.Formula != formulaConstant && (xp.Amount != 0 || xp.ContestedBonus != 0) {
		return fmt.Errorf("xp: formula %q takes no amount or contested_bonus", xp.Formula)
	}
	return nil
}

package goal

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

const (
	causeProgress  = "goal_progress"
	causeReset     = "goal_reset"
	causeCompleted = "goal_completed"
)

// evaluate reads one batch for the holder's active goal and returns the goal
// mutations it causes. It must run before the batch applies: the matchers read
// about-to-be-removed assets from live state, and XP is worked out from the state
// before the loss (frozen log D119).
func (compiled *compiledGoal) evaluate(ctx evalContext) []domain.Mutation {
	goal := compiled.goal
	progress := ctx.faction.ActiveGoal.Progress
	if compiled.reset != nil && len(compiled.reset.match(ctx, goal.Reset.Filters)) > 0 {
		// deliberate: a reset ends the batch for this goal, not only zeroes it, because progress counted in the same batch would outlive the reset (Decision 8).
		if progress == 0 {
			return nil
		}
		return []domain.Mutation{compiled.progressed(ctx, -progress, causeReset)}
	}
	var mutations []domain.Mutation
	delta := compiled.progressDelta(ctx, progress)
	if delta != 0 {
		mutations = append(mutations, compiled.progressed(ctx, delta, causeProgress))
		progress += delta
	}
	completing, completed := compiled.completion(ctx, progress, delta != 0)
	if !completed {
		return mutations
	}
	xp := compiled.xp(ctx, progress, completing)
	return append(mutations,
		domain.GoalCompleted{FactionID: ctx.faction.ID, GoalID: goal.ID, XPAwarded: xp, Cause: causeCompleted},
		domain.XPAwarded{FactionID: ctx.faction.ID, Amount: xp, Cause: causeCompleted},
	)
}

func (compiled *compiledGoal) progressed(ctx evalContext, delta int, cause string) domain.GoalProgressed {
	return domain.GoalProgressed{FactionID: ctx.faction.ID, GoalID: compiled.goal.ID, Delta: delta, Cause: cause}
}

// progressDelta is how far this batch moves progress: the sum of an event's
// matches, or the gap between a measure's count and the progress held.
func (compiled *compiledGoal) progressDelta(ctx evalContext, progress int) int {
	part := compiled.goal.Progress
	switch {
	case compiled.measure != nil:
		return compiled.measure.count(ctx, part.Filters) - progress
	case compiled.progress != nil:
		matches := compiled.progress.match(ctx, part.Filters)
		if part.Amount == amountOne {
			return len(matches)
		}
		total := 0
		for _, matched := range matches {
			total += matched.amount
		}
		return total
	default:
		return 0
	}
}

// completion reports whether the goal completes on this batch, and the match
// that completed it when a complete event did. A threshold completes only in a
// batch that moved progress (Decision 13).
func (compiled *compiledGoal) completion(ctx evalContext, progress int, moved bool) (match, bool) {
	if compiled.complete != nil {
		matches := compiled.complete.match(ctx, compiled.goal.Complete.Filters)
		if len(matches) == 0 {
			return match{}, false
		}
		return matches[0], true
	}
	return match{}, moved && progress >= thresholdOf(compiled.goal.Threshold, ctx.faction)
}

func thresholdOf(threshold *domain.GoalThreshold, faction *domain.Faction) int {
	if threshold.Constant > 0 {
		return threshold.Constant
	}
	total := 0
	for _, stat := range threshold.Ratings {
		total += faction.StatScore(stat)
	}
	if threshold.Multiplier > 0 {
		total *= threshold.Multiplier
	}
	return total
}

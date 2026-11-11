package goal

import (
	"fmt"
	"log/slog"
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// GoalEngine advances faction goals from the parts each goal states in the
// rulebook. It holds no goal ID of its own.
type GoalEngine struct {
	goals       map[string]*compiledGoal
	actionNames []string
	rulebook    *rulebook.Rulebook
	log         *slog.Logger
}

// New compiles every goal in the rulebook. A goal whose parts name something the
// engine cannot run is an error, not a goal that does nothing. actionNames are the
// registered actions: a goal's action filters may name only these, and its
// needs_actions gate is open only when it names these.
func New(rulebook *rulebook.Rulebook, actionNames []string, log *slog.Logger) (*GoalEngine, error) {
	goalEngine := &GoalEngine{
		goals:       make(map[string]*compiledGoal, len(rulebook.Goals)),
		actionNames: actionNames,
		rulebook:    rulebook,
		log:         log,
	}
	for _, goalID := range slices.Sorted(maps.Keys(rulebook.Goals)) {
		compiled, err := compile(rulebook.Goals[goalID], actionNames)
		if err != nil {
			return nil, fmt.Errorf("goal %q: %w", goalID, err)
		}
		goalEngine.goals[goalID] = compiled
	}
	goalEngine.log.Debug("goals compiled", "count", len(goalEngine.goals))
	return goalEngine, nil
}

// UpdateProgress reads one batch for the active goal of every faction not
// destroyed, in faction ID order, and returns the goal mutations it causes. Every goal reads the state
// before the batch, so it must be called before MutationEngine.Apply.
func (goalEngine *GoalEngine) UpdateProgress(
	mutations []domain.Mutation,
	factionState *state.FactionState,
	log *slog.Logger,
) []domain.Mutation {
	var goalMutations []domain.Mutation
	for _, factionID := range slices.Sorted(maps.Keys(factionState.Factions)) {
		faction := factionState.Factions[factionID]
		// deliberate: a destroyed faction's goal is skipped, not read, because the engine keeps destroyed factions in state and they must not earn goal XP.
		if faction.ActiveGoal == nil || faction.CurrentHP <= 0 {
			continue
		}
		compiled, ok := goalEngine.goals[faction.ActiveGoal.GoalID]
		if !ok {
			// deliberate: a held goal the rulebook does not know is skipped, not an error, because an older save can hold one (Decision 10).
			log.Warn("faction holds an unknown goal", "faction", factionID, "goal", faction.ActiveGoal.GoalID)
			continue
		}
		goalMutations = append(goalMutations, compiled.evaluate(evalContext{
			faction:      faction,
			batch:        mutations,
			factionState: factionState,
			rulebook:     goalEngine.rulebook,
		})...)
	}
	return goalMutations
}

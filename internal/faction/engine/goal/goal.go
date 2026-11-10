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
	goals map[string]*compiledGoal
	log   *slog.Logger
}

// New compiles every goal in the rulebook. A goal whose parts name something the
// engine cannot run is an error, not a goal that does nothing.
func New(rulebook *rulebook.Rulebook, log *slog.Logger) (*GoalEngine, error) {
	goalEngine := &GoalEngine{goals: make(map[string]*compiledGoal, len(rulebook.Goals)), log: log}
	for _, goalID := range slices.Sorted(maps.Keys(rulebook.Goals)) {
		compiled, err := compile(rulebook.Goals[goalID])
		if err != nil {
			return nil, fmt.Errorf("goal %q: %w", goalID, err)
		}
		goalEngine.goals[goalID] = compiled
	}
	return goalEngine, nil
}

// UpdateProgress reads a batch for the acting faction's active goal and returns
// the goal mutations it causes. It must be called before MutationEngine.Apply.
func (goalEngine *GoalEngine) UpdateProgress(
	actingFactionID string,
	mutations []domain.Mutation,
	factionState *state.FactionState,
	rulebook *rulebook.Rulebook,
	log *slog.Logger,
) []domain.Mutation {
	actingFaction, ok := factionState.Factions[actingFactionID]
	if !ok || actingFaction.ActiveGoal == nil {
		return nil
	}
	compiled, ok := goalEngine.goals[actingFaction.ActiveGoal.GoalID]
	if !ok {
		// deliberate: a held goal the rulebook does not know is skipped, not an error, because an older save can hold one (Decision 10).
		log.Warn("faction holds an unknown goal", "faction", actingFactionID, "goal", actingFaction.ActiveGoal.GoalID)
		return nil
	}
	return compiled.evaluate(evalContext{
		faction:      actingFaction,
		batch:        mutations,
		factionState: factionState,
		rulebook:     rulebook,
	})
}

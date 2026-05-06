package engine

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/eventhooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

const maxHookDepth = 5

// dispatchMutationReactors runs all MutationReactors registered for faction
// (faction-scoped and global) in registration order against combined. Returned
// mutations are appended and the process recurses; bounded at maxHookDepth
// levels. On cap trip the observer is notified and recursion stops — all
// already-collected mutations still apply.
func dispatchMutationReactors(
	registry *eventhooks.Registry,
	faction *domain.Faction,
	combined []domain.Mutation,
	factionState *state.FactionState,
	rb *rulebook.Rulebook,
	observer TurnObserver,
) []domain.Mutation {
	return reactDispatch(registry, faction, combined, factionState, rb, observer, 0)
}

func reactDispatch(
	registry *eventhooks.Registry,
	faction *domain.Faction,
	combined []domain.Mutation,
	factionState *state.FactionState,
	rb *rulebook.Rulebook,
	observer TurnObserver,
	depth int,
) []domain.Mutation {
	if depth >= maxHookDepth {
		observer.OnError(faction, fmt.Errorf("hook recursion depth exceeded for faction %s", faction.ID))
		return combined
	}

	reactors := registry.MutationReactorsFor(faction.ID, "")
	var newMutations []domain.Mutation
	for _, registered := range reactors {
		extra := registered.Hook.OnMutations(combined, factionState, rb)
		newMutations = append(newMutations, extra...)
	}

	if len(newMutations) == 0 {
		return combined
	}

	return reactDispatch(registry, faction, append(combined, newMutations...), factionState, rb, observer, depth+1)
}

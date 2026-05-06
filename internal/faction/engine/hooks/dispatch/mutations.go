package dispatch

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

const maxHookDepth = 5

// MutationReactors runs all MutationReactors registered for faction
// (faction-scoped and global) in registration order against combined.
// Returned mutations are appended and the process recurses until no
// new mutations are produced. Returns an error if the recursion depth
// exceeds maxHookDepth; already-collected mutations are still returned.
func MutationReactors(
	registry *hooks.Registry,
	faction *domain.Faction,
	combined []domain.Mutation,
	factionState *state.FactionState,
	rb *rulebook.Rulebook,
) ([]domain.Mutation, error) {
	return reactDispatch(registry, faction, combined, factionState, rb, 0)
}

func reactDispatch(
	registry *hooks.Registry,
	faction *domain.Faction,
	combined []domain.Mutation,
	factionState *state.FactionState,
	rb *rulebook.Rulebook,
	depth int,
) ([]domain.Mutation, error) {
	if depth >= maxHookDepth {
		return combined, fmt.Errorf("hook recursion depth exceeded for faction %s", faction.ID)
	}

	reactors := registry.MutationReactorsFor(faction.ID, "")
	var newMutations []domain.Mutation
	for _, registered := range reactors {
		extra := registered.Hook.OnMutations(combined, factionState, rb)
		newMutations = append(newMutations, extra...)
	}

	if len(newMutations) == 0 {
		return combined, nil
	}

	return reactDispatch(registry, faction, append(combined, newMutations...), factionState, rb, depth+1)
}

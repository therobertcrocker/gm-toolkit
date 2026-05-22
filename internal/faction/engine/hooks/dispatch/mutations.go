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
// exceeds maxHookDepth to prevent infinite loops.
func MutationReactors(
	registry *hooks.Registry,
	faction *domain.Faction,
	combined []domain.Mutation,
	factionState *state.FactionState,
	rulebook *rulebook.Rulebook,
) ([]domain.Mutation, error) {
	return reactDispatch(registry, faction, combined, factionState, rulebook, 0)
}

func reactDispatch(
	registry *hooks.Registry,
	faction *domain.Faction,
	combined []domain.Mutation,
	factionState *state.FactionState,
	rulebook *rulebook.Rulebook,
	depth int,
) ([]domain.Mutation, error) {
	if depth >= maxHookDepth {
		return nil, fmt.Errorf("hook recursion depth exceeded for faction %s", faction.ID)
	}

	reactors := registry.MutationReactorsFor(faction.ID, "")
	var newMutations []domain.Mutation
	for _, registered := range reactors {
		extra := registered.Hook.OnMutations(combined, factionState, rulebook)
		newMutations = append(newMutations, extra...)
	}

	if len(newMutations) == 0 {
		return combined, nil
	}

	// Recurse with only the newly emitted mutations as input so reactors do not
	// re-fire on the original triggers. The accumulated result is built up on
	// the way back out.
	deeper, err := reactDispatch(registry, faction, newMutations, factionState, rulebook, depth+1)
	return append(combined, deeper...), err
}

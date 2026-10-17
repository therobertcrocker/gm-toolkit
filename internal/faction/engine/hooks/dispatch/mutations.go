package dispatch

import (
	"fmt"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

const maxHookDepth = 5

// MutationReactors runs every registered MutationReactor against combined —
// global scope, then each faction's scope, then each asset's scope, in sorted
// ID order so the emitted slice is deterministic. Reactors belonging to a
// faction other than the acting one do run: a defender reacts to what was done
// to it. Returned mutations are appended and the process recurses until no new
// mutations are produced. Returns an error if the recursion depth exceeds
// maxHookDepth to prevent infinite loops.
func MutationReactors(
	registry *hooks.Registry,
	faction *domain.Faction,
	combined []domain.Mutation,
	factionState *state.FactionState,
	rulebook *rulebook.Rulebook,
) ([]domain.Mutation, error) {
	return reactDispatch(registry, faction, combined, factionState, rulebook, 0)
}

// allReactors collects every registered Cat 3 hook in deterministic order:
// global, then per faction (sorted by ID), then per asset within that faction
// (sorted by ID).
func allReactors(registry *hooks.Registry, factionState *state.FactionState) []hooks.RegisteredMutationReactor {
	reactors := append([]hooks.RegisteredMutationReactor(nil), registry.MutationReactorsGlobal()...)

	factionIDs := make([]string, 0, len(factionState.Factions))
	for id := range factionState.Factions {
		factionIDs = append(factionIDs, id)
	}
	sort.Strings(factionIDs)

	for _, factionID := range factionIDs {
		reactors = append(reactors, registry.MutationReactorsForFaction(factionID)...)
		for _, asset := range domain.SortedAssets(factionState.Factions[factionID]) {
			reactors = append(reactors, registry.MutationReactorsForAsset(factionID, asset.ID)...)
		}
	}
	return reactors
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

	reactors := allReactors(registry, factionState)
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

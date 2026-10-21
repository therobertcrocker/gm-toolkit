package turn

import (
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// ExpireConditions clears every condition, on any faction's asset, that ends at
// the start of factionID's turn. Factions in ID order, then assets in ID order.
func ExpireConditions(factionState *state.FactionState, factionID string) []domain.Mutation {
	var mutations []domain.Mutation
	for _, holderID := range slices.Sorted(maps.Keys(factionState.Factions)) {
		for _, asset := range domain.SortedAssets(factionState.Factions[holderID]) {
			for _, condition := range asset.Conditions {
				if condition.ExpiresAtTurnOf != factionID {
					continue
				}
				mutations = append(mutations, domain.AssetConditionCleared{
					FactionID:         holderID,
					AssetID:           asset.ID,
					Condition:         condition,
					Cause:             "condition_expired",
					CausedByFactionID: factionID,
				})
			}
		}
	}
	return mutations
}

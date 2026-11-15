package turn

import (
	"fmt"
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
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

// OfferConditionLifts asks faction, for each of its disabled assets in ID order,
// whether to pay the asset's half price to lift the condition. A price it cannot
// pay, after earlier lifts, is not offered.
func OfferConditionLifts(faction *domain.Faction, rulebook *rulebook.Rulebook, collector hooks.PaymentCollector) ([]domain.Mutation, error) {
	coin := faction.Coin
	var mutations []domain.Mutation
	for _, asset := range domain.SortedAssets(faction) {
		for _, condition := range asset.Conditions {
			if condition.Kind != domain.ConditionDisabled {
				continue
			}
			def, ok := rulebook.Assets[asset.DefinitionID]
			if !ok {
				return nil, fmt.Errorf("offer condition lifts: definition not found: %s", asset.DefinitionID)
			}
			price := def.HalfPrice()
			if price > coin {
				continue
			}
			paid, err := collector.ConfirmPayment(hooks.PaymentOffer{Payer: faction, Asset: asset, Amount: price, Reason: hooks.PaymentLiftDisable})
			if err != nil {
				return nil, fmt.Errorf("offer condition lifts: %w", err)
			}
			if !paid {
				continue
			}
			coin -= price
			mutations = append(mutations,
				domain.CoinDelta{FactionID: faction.ID, Delta: -price, Cause: "condition_lifted", CausedByFactionID: faction.ID},
				domain.AssetConditionCleared{FactionID: faction.ID, AssetID: asset.ID, Condition: condition, Cause: "condition_lifted", CausedByFactionID: faction.ID},
			)
		}
	}
	return mutations, nil
}

// OfferShelterReleases asks faction, for each of its sheltered assets in ID order
// that may leave, whether to bring it out, at no cost (Effort 5 Decision 7).
func OfferShelterReleases(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook, collector hooks.PaymentCollector) ([]domain.Mutation, error) {
	var mutations []domain.Mutation
	for _, asset := range domain.SortedAssets(faction) {
		for _, condition := range asset.Conditions {
			if condition.Kind != domain.ConditionSheltered || !domain.InDoorShelter(factionState.Factions, rulebook.Assets, condition) {
				continue
			}
			release, err := collector.ConfirmPayment(hooks.PaymentOffer{Payer: faction, Asset: asset, Amount: 0, Reason: hooks.PaymentReleaseShelter})
			if err != nil {
				return nil, fmt.Errorf("offer shelter releases: %w", err)
			}
			if !release {
				continue
			}
			mutations = append(mutations, domain.AssetConditionCleared{FactionID: faction.ID, AssetID: asset.ID, Condition: condition, Cause: "shelter_released", CausedByFactionID: faction.ID})
		}
	}
	return mutations, nil
}

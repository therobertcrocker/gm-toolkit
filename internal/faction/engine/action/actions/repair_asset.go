package actions

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// RepairAsset restores HP to one or more of the faction's damaged assets.
// Cost escalates per heal on the same asset; resets to 1 Coin for each new asset.
// A RepairCostModifier can lower each asset's price (Medical Center).
type RepairAsset struct {
	collector    action.Collector
	registry     *hooks.Registry
	factionID    string
	repairOrders []action.RepairOrder
	mutations    []domain.Mutation
}

func NewRepairAsset(collector action.Collector, registry *hooks.Registry) *RepairAsset {
	return &RepairAsset{collector: collector, registry: registry}
}

func (ra *RepairAsset) Name() string { return "Repair Asset" }

func (ra *RepairAsset) Validate(faction *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) bool {
	if faction.Coin < 1 {
		return false
	}
	for _, asset := range faction.Assets {
		def, ok := rulebook.Assets[asset.DefinitionID]
		if ok && asset.CurrentHP < def.HP {
			return true
		}
	}
	return false
}

func (ra *RepairAsset) Inputs(faction *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) error {
	var damaged []*domain.Asset
	for _, asset := range faction.Assets {
		def, ok := rulebook.Assets[asset.DefinitionID]
		if ok && asset.CurrentHP < def.HP {
			damaged = append(damaged, asset)
		}
	}

	orders, err := ra.collector.SelectRepairOrders(faction, damaged, rulebook)
	if err != nil {
		return fmt.Errorf("repair asset: %w", err)
	}
	ra.factionID = faction.ID
	ra.repairOrders = orders
	return nil
}

func (ra *RepairAsset) Resolve(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	totalCost := 0
	var mutations []domain.Mutation

	for _, order := range ra.repairOrders {
		def, ok := rulebook.Assets[order.Asset.DefinitionID]
		if !ok {
			return fmt.Errorf("asset definition not found: %s", order.Asset.DefinitionID)
		}

		attributeScore := faction.StatScore(def.Category)
		missing := def.HP - order.Asset.CurrentHP
		totalHeal := 0

		assetCost := 0
		for healIndex := range order.HealCount {
			healAmount := min(attributeScore, missing-totalHeal)
			if healAmount <= 0 {
				break
			}
			assetCost += healIndex + 1 // 1-indexed escalating cost
			totalHeal += healAmount
		}

		if totalHeal > 0 {
			totalCost += dispatch.ResolveRepairCost(ra.registry, faction, order.Asset, def, assetCost)
			mutations = append(mutations, domain.AssetHPDelta{
				FactionID:         ra.factionID,
				AssetID:           order.Asset.ID,
				Delta:             totalHeal,
				Cause:             "repair",
				CausedByFactionID: ra.factionID,
			})
			if order.Asset.CurrentHP+totalHeal >= def.HP {
				mutations = append(mutations, shelterCleared(ra.factionID, order.Asset, factionState, rulebook)...)
			}
		}
	}

	if totalCost > faction.Coin {
		return fmt.Errorf("insufficient Coin: need %d, have %d", totalCost, faction.Coin)
	}
	if totalCost > 0 {
		mutations = append(mutations, domain.CoinDelta{FactionID: ra.factionID, Delta: -totalCost, Cause: "repair", CausedByFactionID: ra.factionID})
	}

	ra.mutations = mutations
	return nil
}

func (ra *RepairAsset) Output() ([]domain.Mutation, error) {
	return ra.mutations, nil
}

// shelterCleared clears the sheltered conditions a repair to full HP ends
// (Effort 2 Decision 12): every one but a standing door shelter's, which the
// asset leaves only by its owner's release (A.001.5 Effort 5 Decision 8).
func shelterCleared(factionID string, asset *domain.Asset, factionState *state.FactionState, rulebook *rulebook.Rulebook) []domain.Mutation {
	var cleared []domain.Mutation
	for _, condition := range asset.Conditions {
		if condition.Kind == domain.ConditionSheltered && !domain.InDoorShelter(factionState.Factions, rulebook.Assets, condition) {
			cleared = append(cleared, domain.AssetConditionCleared{
				FactionID:         factionID,
				AssetID:           asset.ID,
				Condition:         condition,
				Cause:             "repair",
				CausedByFactionID: factionID,
			})
		}
	}
	return cleared
}

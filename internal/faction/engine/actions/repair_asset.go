package actions

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// RepairAsset restores HP to one or more of the faction's damaged assets.
// Cost escalates per heal on the same asset; resets to 1 Coin for each new asset.
type RepairAsset struct {
	collector    engine.InputCollector
	factionID    string
	repairOrders []engine.RepairOrder
	mutations    []domain.Mutation
}

func NewRepairAsset(collector engine.InputCollector) *RepairAsset {
	return &RepairAsset{collector: collector}
}

func (ra *RepairAsset) Name() string { return "Repair Asset" }

func (ra *RepairAsset) Validate(faction *domain.Faction, _ *state.FactionState, rulebook *loader.Rulebook) bool {
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

func (ra *RepairAsset) Inputs(faction *domain.Faction, _ *state.FactionState, rulebook *loader.Rulebook) error {
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

func (ra *RepairAsset) Resolve(faction *domain.Faction, _ *state.FactionState, rulebook *loader.Rulebook) error {
	totalCost := 0
	var mutations []domain.Mutation

	for _, order := range ra.repairOrders {
		def, ok := rulebook.Assets[order.Asset.DefinitionID]
		if !ok {
			return fmt.Errorf("asset definition not found: %s", order.Asset.DefinitionID)
		}

		attributeScore := statScore(faction, def.Category)
		missing := def.HP - order.Asset.CurrentHP
		totalHeal := 0

		for healIndex := range order.HealCount {
			healAmount := min(attributeScore, missing-totalHeal)
			if healAmount <= 0 {
				break
			}
			totalCost += healIndex + 1 // 1-indexed escalating cost
			totalHeal += healAmount
		}

		if totalHeal > 0 {
			mutations = append(mutations, domain.AssetHPDelta{
				FactionID:         ra.factionID,
				AssetID:           order.Asset.ID,
				Delta:             totalHeal,
				Cause:             "repair",
				CausedByFactionID: ra.factionID,
			})
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

// statScore returns the faction's attribute value for the given asset category.
func statScore(faction *domain.Faction, category domain.FactionStat) int {
	switch category {
	case domain.StatForce:
		return faction.Force
	case domain.StatCunning:
		return faction.Cunning
	case domain.StatWealth:
		return faction.Wealth
	default:
		return 0
	}
}

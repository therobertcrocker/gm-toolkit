package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
)

// InputCollector abstracts input collection for action resolution. The GM
// implementation uses interactive prompts; an AI implementation uses
// goal-driven selection logic. Actions call only the methods they need.
type InputCollector interface {
	SelectAsset(assets []*domain.Asset, rulebook *loader.Rulebook) (*domain.Asset, error)
	SelectRepairOrders(faction *domain.Faction, damagedAssets []*domain.Asset, rulebook *loader.Rulebook) ([]RepairOrder, error)
	SelectBuyOrder(worlds []string, purchasable []*domain.AssetDefinition) (BuyOrder, error)
	SelectRefitOrder(options []RefitOption, rulebook *loader.Rulebook) (RefitOrder, error)
}

// RepairOrder describes a single asset repair instruction: which asset and how
// many heals to apply. Cost escalates per heal on the same asset (1 Coin for
// the first, +1 for each subsequent).
type RepairOrder struct {
	Asset     *domain.Asset
	HealCount int
}

// BuyOrder describes a Buy Asset instruction: the target world and the
// definition of the asset to purchase.
type BuyOrder struct {
	World      string
	Definition *domain.AssetDefinition
}

// RefitOption bundles a refittable asset with the list of valid replacement
// definitions for that specific asset.
type RefitOption struct {
	Asset        *domain.Asset
	Replacements []*domain.AssetDefinition
}

// RefitOrder describes a Refit Asset instruction: which asset to refit and
// which definition to replace it with.
type RefitOrder struct {
	OldAsset      *domain.Asset
	NewDefinition *domain.AssetDefinition
}

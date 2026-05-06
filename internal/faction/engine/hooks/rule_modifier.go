package hooks

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

// MovementAbility describes an extra movement capability granted by a tag or
// asset effect.
type MovementAbility struct {
	Name        string
	Description string
}

// AssetCostModifier adjusts the Coin cost of an asset purchase (Cat 4).
type AssetCostModifier interface {
	ModifyAssetCost(buyer *domain.Faction, def *domain.AssetDefinition, world string, baseCost int) int
}

// MaintenanceCostModifier adjusts the per-turn maintenance cost of an asset
// (Cat 4).
type MaintenanceCostModifier interface {
	ModifyMaintenanceCost(owner *domain.Faction, asset *domain.Asset, baseCost int) int
}

// WorldTechLevelModifier adjusts the effective tech level of a world for a
// specific faction, used by TL-gated asset purchase logic (Cat 4).
type WorldTechLevelModifier interface {
	ModifyWorldTechLevel(faction *domain.Faction, world string, baseTL int) int
}

// AssetMovementGranter reports additional movement abilities conferred on an
// asset by a tag or special effect (Cat 4).
type AssetMovementGranter interface {
	GrantMovementAbilities(asset *domain.Asset) []MovementAbility
}

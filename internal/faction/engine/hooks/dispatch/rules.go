package dispatch

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// ResolveAssetCost runs all registered AssetCostModifiers for buyer in
// registration order, chaining outputs so each modifier sees the previous
// one's result. Returns baseCost when registry is nil or no modifiers match.
func ResolveAssetCost(registry *hooks.Registry, buyer *domain.Faction, def *domain.AssetDefinition, world string, baseCost int) int {
	if registry == nil {
		return baseCost
	}
	cost := baseCost
	for _, registered := range registry.AssetCostModifiersFor(buyer.ID, "") {
		cost = registered.Hook.ModifyAssetCost(buyer, def, world, cost)
	}
	return cost
}

// ResolveMaintenanceCost runs all registered MaintenanceCostModifiers for
// owner and asset in registration order, chaining outputs. Returns baseCost
// when registry is nil or no modifiers match.
func ResolveMaintenanceCost(registry *hooks.Registry, owner *domain.Faction, asset *domain.Asset, baseCost int) int {
	if registry == nil {
		return baseCost
	}
	cost := baseCost
	for _, registered := range registry.MaintenanceCostModifiersFor(owner.ID, asset.ID) {
		cost = registered.Hook.ModifyMaintenanceCost(owner, asset, cost)
	}
	return cost
}

// ResolveWorldTechLevel runs all registered WorldTechLevelModifiers for
// faction in registration order, chaining outputs. Returns baseTL when
// registry is nil or no modifiers match.
func ResolveWorldTechLevel(registry *hooks.Registry, faction *domain.Faction, world string, baseTL int) int {
	if registry == nil {
		return baseTL
	}
	tl := baseTL
	for _, registered := range registry.WorldTechLevelModifiersFor(faction.ID, "") {
		tl = registered.Hook.ModifyWorldTechLevel(faction, world, tl)
	}
	return tl
}

// GrantedMovementAbilities collects extra movement abilities granted to asset
// by all registered AssetMovementGranters (global + faction-scoped +
// asset-scoped). Returns nil when registry is nil or no granters match.
func GrantedMovementAbilities(registry *hooks.Registry, asset *domain.Asset) []hooks.MovementAbility {
	if registry == nil {
		return nil
	}
	var abilities []hooks.MovementAbility
	for _, registered := range registry.AssetMovementGrantersFor(asset.OwnerID, asset.ID) {
		abilities = append(abilities, registered.Hook.GrantMovementAbilities(asset)...)
	}
	return abilities
}

// ResolveTie consults registered TieResolvers for the attacker, then the
// defender; first registration wins. Returns TieStandard when no resolvers
// match.
func ResolveTie(registry *hooks.Registry, ctx hooks.RollContext, factionState *state.FactionState) hooks.TieOutcome {
	if registry == nil {
		return hooks.TieStandard
	}
	resolvers := registry.TieResolversFor(ctx.Actor.ID, "")
	if len(resolvers) == 0 && ctx.Opponent != nil {
		resolvers = registry.TieResolversFor(ctx.Opponent.ID, "")
	}
	if len(resolvers) == 0 {
		return hooks.TieStandard
	}
	return resolvers[0].Hook.ResolveTie(ctx, factionState)
}

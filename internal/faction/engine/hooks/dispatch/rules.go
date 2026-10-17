package dispatch

import (
	"slices"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// ResolveAssetCost runs the buyer's own AssetCostModifiers, then those registered
// against every asset standing on world — including a rival's, because a cost
// hook may belong to a faction other than the buyer (Local Investments). Hooks
// run in registration order, chaining outputs so each sees the previous result.
// onWorld is the world index's asset slice for world; a nil slice adds nothing.
func ResolveAssetCost(registry *hooks.Registry, buyer *domain.Faction, def *domain.AssetDefinition, world string, baseCost int, onWorld []*domain.Asset) hooks.AssetCostQuery {
	query := &hooks.AssetCostQuery{Cost: baseCost, Applied: map[string]bool{}}
	if registry == nil {
		return *query
	}
	for _, registered := range registry.AssetCostModifiersFor(buyer.ID, "") {
		registered.Hook.ModifyAssetCost(buyer, def, world, query)
	}
	for _, asset := range sortedByID(onWorld) {
		if asset.OwnerID == buyer.ID {
			continue
		}
		for _, registered := range registry.AssetCostModifiersForAsset(asset.OwnerID, asset.ID) {
			registered.Hook.ModifyAssetCost(buyer, def, world, query)
		}
	}
	return *query
}

// DiscountedCost applies a pending purchase discount to a resolved cost. The
// floor is half the base price, never above the pre-discount price, and the
// surcharge rides on top so no discount removes it.
func DiscountedCost(quote hooks.AssetCostQuery, baseCost, discount int) int {
	beforeSurcharge := quote.Cost - quote.Surcharge
	floor := min(baseCost/2, beforeSurcharge)
	return max(beforeSurcharge-discount, floor) + quote.Surcharge
}

// sortedByID returns assets in ID order so a world-scoped hook chain is
// deterministic regardless of index insertion order.
func sortedByID(assets []*domain.Asset) []*domain.Asset {
	sorted := slices.Clone(assets)
	slices.SortFunc(sorted, func(first, second *domain.Asset) int {
		return strings.Compare(first.ID, second.ID)
	})
	return sorted
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

// ResolveWorldTechLevel runs faction's own WorldTechLevelModifiers, then those
// registered against each asset faction owns, in asset ID order. A tech-level
// floor serves its owner only, so a rival's assets are never asked; each
// asset-scoped hook decides whether its reach covers world and def.
func ResolveWorldTechLevel(registry *hooks.Registry, faction *domain.Faction, def *domain.AssetDefinition, world string, baseTL int) int {
	if registry == nil {
		return baseTL
	}
	tl := baseTL
	for _, registered := range registry.WorldTechLevelModifiersFor(faction.ID, "") {
		tl = registered.Hook.ModifyWorldTechLevel(faction, def, world, tl)
	}
	for _, asset := range domain.SortedAssets(faction) {
		for _, registered := range registry.WorldTechLevelModifiersForAsset(faction.ID, asset.ID) {
			tl = registered.Hook.ModifyWorldTechLevel(faction, def, world, tl)
		}
	}
	return tl
}

// ResolveAttackEligibility reports whether query.Attacker may deal damage to
// query.Defender. The striker's hooks (global, faction and asset) answer as
// query.Side. On an attack, the target's asset hooks then answer as defense; a
// counterattack is not a defense, so the target is not asked. Any hook that
// refuses ends the fight.
func ResolveAttackEligibility(registry *hooks.Registry, query hooks.AttackEligibilityQuery) bool {
	if registry == nil {
		return true
	}
	for _, registered := range registry.AttackEligibilityRulesFor(query.Attacker.OwnerID, query.Attacker.ID) {
		if !registered.Hook.MayFight(query) {
			return false
		}
	}
	if query.Side != hooks.SideAttack {
		return true
	}
	defense := query
	defense.Side = hooks.SideDefense
	for _, registered := range registry.AttackEligibilityRulesForAsset(query.Defender.OwnerID, query.Defender.ID) {
		if !registered.Hook.MayFight(defense) {
			return false
		}
	}
	return true
}

// GrantedMovementAbilities collects the free movement offered to faction: its
// global and faction granters, then those registered against each asset it owns,
// in asset ID order. A grant serves its owner only, so a rival's assets are never
// asked. Returns nil when registry is nil or no granter offers anything.
func GrantedMovementAbilities(registry *hooks.Registry, faction *domain.Faction) []hooks.MovementAbility {
	if registry == nil {
		return nil
	}
	var abilities []hooks.MovementAbility
	for _, registered := range registry.AssetMovementGrantersFor(faction.ID, "") {
		abilities = append(abilities, registered.Hook.GrantMovementAbilities(faction)...)
	}
	for _, asset := range domain.SortedAssets(faction) {
		for _, registered := range registry.AssetMovementGrantersForAsset(faction.ID, asset.ID) {
			abilities = append(abilities, registered.Hook.GrantMovementAbilities(faction)...)
		}
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

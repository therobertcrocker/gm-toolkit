package dispatch

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// ResolveAssetCost runs the buyer's own AssetCostModifiers, then those registered
// against every asset standing on world — including a rival's, because a cost
// hook may belong to a faction other than the buyer (Local Investments). Hooks
// run in registration order, chaining outputs so each sees the previous result.
// onWorld is the assets on world in ID order (world.AssetsOnWorld), so the chain
// is the same every run; a nil slice adds nothing.
func ResolveAssetCost(registry *hooks.Registry, buyer *domain.Faction, def *domain.AssetDefinition, world string, baseCost int, onWorld []*domain.Asset) hooks.AssetCostQuery {
	query := &hooks.AssetCostQuery{Cost: baseCost, Applied: map[string]bool{}}
	if registry == nil {
		return *query
	}
	for _, registered := range registry.AssetCostModifiersFor(buyer.ID, "") {
		registered.Hook.ModifyAssetCost(buyer, def, world, query)
	}
	for _, asset := range onWorld {
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
	surcharge := quote.SurchargeTotal()
	beforeSurcharge := quote.Cost - surcharge
	floor := min(baseCost/2, beforeSurcharge)
	return max(beforeSurcharge-discount, floor) + surcharge
}

// PurchaseCharges splits a purchase's cost into the buyer's own charge and one
// charge per surcharge, caused by the faction that imposed it, so a reactor such
// as Bank can tell a cost a rival imposed from the price itself.
func PurchaseCharges(buyerID string, cost int, quote hooks.AssetCostQuery, cause string) []domain.Mutation {
	charges := []domain.Mutation{
		domain.CoinDelta{FactionID: buyerID, Delta: -(cost - quote.SurchargeTotal()), Cause: cause, CausedByFactionID: buyerID},
	}
	for _, surcharge := range quote.Surcharges {
		charges = append(charges, domain.CoinDelta{FactionID: buyerID, Delta: -surcharge.Amount, Cause: string(domain.SpecialCostSurcharge), CausedByFactionID: surcharge.FactionID})
	}
	return charges
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

// PermissionGranted reports whether one of query.Requester's own grant hooks grants
// the request: its global and faction granters, then those registered against each
// asset it holds, in asset ID order. A grant serves its owner only, so a rival's
// assets are never asked; the first yes ends the walk. Returns false when registry
// is nil.
func PermissionGranted(registry *hooks.Registry, query hooks.PermissionQuery) bool {
	if registry == nil {
		return false
	}
	for _, registered := range registry.PermissionGrantersFor(query.Requester.ID, "") {
		if registered.Hook.GrantsPermission(query) {
			return true
		}
	}
	for _, asset := range domain.SortedAssets(query.Requester) {
		for _, registered := range registry.PermissionGrantersForAsset(query.Requester.ID, asset.ID) {
			if registered.Hook.GrantsPermission(query) {
				return true
			}
		}
	}
	return false
}

// ResolveRepairCost runs the repairer's own RepairCostModifiers, then those
// registered against each asset it holds, in asset ID order, chaining the query.
// A repair discount serves its owner only, so a rival's assets are never asked.
// The cost never drops below 0.
func ResolveRepairCost(registry *hooks.Registry, repairer *domain.Faction, asset *domain.Asset, def *domain.AssetDefinition, baseCost int) int {
	if registry == nil {
		return baseCost
	}
	query := &hooks.RepairCostQuery{Cost: baseCost, Applied: map[string]bool{}}
	for _, registered := range registry.RepairCostModifiersFor(repairer.ID, "") {
		registered.Hook.ModifyRepairCost(repairer, asset, def, query)
	}
	for _, held := range domain.SortedAssets(repairer) {
		for _, registered := range registry.RepairCostModifiersForAsset(repairer.ID, held.ID) {
			registered.Hook.ModifyRepairCost(repairer, asset, def, query)
		}
	}
	return max(0, query.Cost)
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

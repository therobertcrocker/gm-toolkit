package dispatch

import (
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// ResolveAssetCost runs the buyer's global and faction AssetCostModifiers, then
// those of each of the buyer's own assets in ID order, then those registered
// against every rival asset standing on world, because a cost hook may belong to
// another faction (Local Investments). Hooks chain, each seeing the previous
// result. onWorld is the assets on world in ID order (world.AssetsOnWorld), so the
// chain is the same every run; a nil slice adds nothing.
func ResolveAssetCost(registry *hooks.Registry, buyer *domain.Faction, def *domain.AssetDefinition, world string, baseCost int, onWorld []*domain.Asset) hooks.AssetCostQuery {
	query := &hooks.AssetCostQuery{Cost: baseCost, Applied: map[string]bool{}}
	if registry == nil {
		return *query
	}
	for _, registered := range registry.AssetCostModifiersFor(buyer.ID, "") {
		registered.Hook.ModifyAssetCost(buyer, def, world, query)
	}
	for _, asset := range domain.SortedAssets(buyer) {
		for _, registered := range registry.AssetCostModifiersForAsset(buyer.ID, asset.ID) {
			registered.Hook.ModifyAssetCost(buyer, def, world, query)
		}
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

// AssetSpeedResolver returns the speed of an asset faction owns, by definition:
// the definition's own, run through faction's AssetSpeedModifiers, then those
// registered against each asset faction owns, in asset ID order. A speed floor
// serves its owner only, so a rival's assets are never asked. The modifiers are
// gathered once, so one resolver serves a whole pass over faction's assets. With
// a nil registry it returns the definition's speed.
func AssetSpeedResolver(registry *hooks.Registry, faction *domain.Faction) func(*domain.AssetDefinition) int {
	if registry == nil {
		return func(definition *domain.AssetDefinition) int { return definition.Speed }
	}
	modifiers := registry.AssetSpeedModifiersFor(faction.ID)
	for _, asset := range domain.SortedAssets(faction) {
		modifiers = append(modifiers, registry.AssetSpeedModifiersForAsset(faction.ID, asset.ID)...)
	}
	return func(definition *domain.AssetDefinition) int {
		speed := definition.Speed
		for _, registered := range modifiers {
			speed = registered.Hook.ModifyAssetSpeed(faction, definition, speed)
		}
		return speed
	}
}

// ArrivalTolls returns the tolls one of mover's assets owes on landing on world:
// one per rule that charges, each paid to the faction that registered it. It asks
// every faction but mover, in ID order, its faction bucket and then its own
// assets' buckets in ID order, because a toll is a rival's effect (Effort 4
// Decision 8). Who skips a toll is the caller's to decide.
func ArrivalTolls(registry *hooks.Registry, factions map[string]*domain.Faction, mover *domain.Faction, world string) []hooks.ArrivalToll {
	if registry == nil || mover == nil {
		return nil
	}
	var tolls []hooks.ArrivalToll
	for _, factionID := range slices.Sorted(maps.Keys(factions)) {
		if factionID == mover.ID {
			continue
		}
		rules := slices.Clone(registry.ArrivalTollRulesAt(hooks.FactionScope(factionID)))
		for _, asset := range domain.SortedAssets(factions[factionID]) {
			rules = append(rules, registry.ArrivalTollRulesAt(hooks.AssetScope(factionID, asset.ID))...)
		}
		for _, registered := range rules {
			if amount := registered.Hook.ArrivalToll(mover, world); amount > 0 {
				tolls = append(tolls, hooks.ArrivalToll{FactionID: factionID, Amount: amount})
			}
		}
	}
	return tolls
}

// AssetArrivalTolls returns the tolls asset owes on landing on world, paid by its
// current owner (Effort 4 Decision 9).
func AssetArrivalTolls(registry *hooks.Registry, factions map[string]*domain.Faction, asset *domain.Asset, world string) []hooks.ArrivalToll {
	// deliberate: a stealthed asset pays no toll, though it still pays a lockdown, because Robert reads the toll as a levy on what the holder can see (Effort 4 Decision 9).
	if asset.Stealthy {
		return nil
	}
	return ArrivalTolls(registry, factions, factions[asset.OwnerID], world)
}

// TollTotal is the Coin tolls charge in all.
func TollTotal(tolls []hooks.ArrivalToll) int {
	total := 0
	for _, toll := range tolls {
		total += toll.Amount
	}
	return total
}

// TollCharges pays each toll from payerID to the faction that charges it: a loss
// and a gain, both caused by that faction, as a steal is, so a Bank absorbs the
// loss and the holder keeps the gain (Effort 4 Decision 8).
func TollCharges(payerID string, tolls []hooks.ArrivalToll) []domain.Mutation {
	var charges []domain.Mutation
	for _, toll := range tolls {
		charges = append(charges,
			domain.CoinDelta{FactionID: payerID, Delta: -toll.Amount, Cause: string(domain.SpecialArrivalToll), CausedByFactionID: toll.FactionID},
			domain.CoinDelta{FactionID: toll.FactionID, Delta: toll.Amount, Cause: string(domain.SpecialArrivalToll), CausedByFactionID: toll.FactionID},
		)
	}
	return charges
}

// ResolvePopulationFloor returns the least world population on which faction may
// buy def: the setting's floor for def's type, run through faction's
// PopulationFloorModifiers, then those of each asset it owns in asset ID order. A
// floor serves its owner only, so a rival's assets are never asked. A waived
// faction, or a type with no floor in floors, needs none (Effort 4 Decision 4).
func ResolvePopulationFloor(registry *hooks.Registry, faction *domain.Faction, def *domain.AssetDefinition, floors map[domain.AssetType]int) int {
	baseFloor := floors[def.Type]
	if faction.PopulationWaiver || baseFloor == 0 {
		return 0
	}
	if registry == nil {
		return baseFloor
	}
	floor := baseFloor
	for _, registered := range registry.PopulationFloorModifiersFor(faction.ID) {
		floor = registered.Hook.ModifyPopulationFloor(faction, def, baseFloor, floor)
	}
	for _, asset := range domain.SortedAssets(faction) {
		for _, registered := range registry.PopulationFloorModifiersForAsset(faction.ID, asset.ID) {
			floor = registered.Hook.ModifyPopulationFloor(faction, def, baseFloor, floor)
		}
	}
	return floor
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

// ResolveTie asks the global hooks, then the attacker's side, then the defender's:
// on each side its faction's hooks, then each of its assets' in ID order. The first
// answer that is not TieStandard wins, because with asset buckets in the walk a
// resolver that does not match must not end the search (Effort 1 Decision 4).
// TieStandard when none answers.
func ResolveTie(registry *hooks.Registry, ctx hooks.RollContext, factionState *state.FactionState) hooks.TieOutcome {
	if registry == nil {
		return hooks.TieStandard
	}
	scopes := []hooks.Scope{hooks.GlobalScope()}
	for _, side := range []*domain.Faction{ctx.Actor, ctx.Opponent} {
		if side == nil {
			continue
		}
		scopes = append(scopes, hooks.FactionScope(side.ID))
		for _, asset := range domain.SortedAssets(side) {
			scopes = append(scopes, hooks.AssetScope(side.ID, asset.ID))
		}
	}
	for _, scope := range scopes {
		for _, registered := range registry.TieResolversAt(scope) {
			if outcome := registered.Hook.ResolveTie(ctx, factionState); outcome != hooks.TieStandard {
				return outcome
			}
		}
	}
	return hooks.TieStandard
}

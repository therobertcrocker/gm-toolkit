package actions

import (
	"fmt"
	"math"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// BuyAsset purchases a new asset and places it on a target world.
// The asset is flagged inactive (Ready: false) until the start of the next turn.
type BuyAsset struct {
	collector        action.Collector
	registry         *hooks.Registry
	worldEngine      *world.WorldEngine
	factionID        string
	buyOrder         action.BuyOrder
	newAsset         domain.Asset
	cost             int
	discountConsumed bool
	stealthTarget    string // asset ID to stealth when buying SWN-C3-002; empty if no eligible target
}

func NewBuyAsset(collector action.Collector, registry *hooks.Registry, worldEngine *world.WorldEngine) *BuyAsset {
	return &BuyAsset{collector: collector, registry: registry, worldEngine: worldEngine}
}

func (ba *BuyAsset) Name() string { return "Buy Asset" }

func (ba *BuyAsset) Validate(faction *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) bool {
	for _, worldID := range availableWorlds(faction) {
		if len(purchasableDefinitions(faction, rulebook, worldID, ba.worldEngine, ba.registry)) > 0 {
			return true
		}
	}
	return false
}

func (ba *BuyAsset) Inputs(faction *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) error {
	purchasableByWorld := make(map[string][]*domain.AssetDefinition)
	for _, worldID := range availableWorlds(faction) {
		defs := purchasableDefinitions(faction, rulebook, worldID, ba.worldEngine, ba.registry)
		if len(defs) > 0 {
			purchasableByWorld[worldID] = defs
		}
	}

	order, err := ba.collector.SelectBuyOrder(purchasableByWorld)
	if err != nil {
		return fmt.Errorf("buy asset: %w", err)
	}
	ba.factionID = faction.ID
	ba.buyOrder = order

	if order.Definition.ID == "SWN-C3-002" {
		targets := eligibleStealthTargets(faction, order.World, rulebook)
		switch len(targets) {
		case 0:
			// no eligible SF assets on the world — stealth quality purchased with no immediate target
		case 1:
			ba.stealthTarget = targets[0].ID
		default:
			target, err := ba.collector.SelectAsset(targets, rulebook)
			if err != nil {
				return fmt.Errorf("buy asset: select stealth target: %w", err)
			}
			ba.stealthTarget = target.ID
		}
	}
	return nil
}

func (ba *BuyAsset) Resolve(faction *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) error {
	def := ba.buyOrder.Definition
	quote := dispatch.ResolveAssetCost(ba.registry, faction, def, ba.buyOrder.World, def.Cost, assetsOnWorld(ba.worldEngine, ba.buyOrder.World))
	ba.cost = quote.Cost
	if faction.PendingPurchaseDiscount > 0 {
		ba.cost = dispatch.DiscountedCost(quote, def.Cost, faction.PendingPurchaseDiscount)
		ba.discountConsumed = true
	}
	if faction.Coin < ba.cost {
		return fmt.Errorf("insufficient Coin: need %d, have %d", ba.cost, faction.Coin)
	}
	ba.newAsset = domain.NewPurchasedAsset(faction, def, ba.buyOrder.World)
	return nil
}

func (ba *BuyAsset) Output() ([]domain.Mutation, error) {
	mutations := []domain.Mutation{
		domain.CoinDelta{FactionID: ba.factionID, Delta: -ba.cost, Cause: "buy", CausedByFactionID: ba.factionID},
	}
	if ba.buyOrder.Definition.ID != "SWN-C3-002" {
		mutations = append([]domain.Mutation{
			domain.AssetAdded{FactionID: ba.factionID, Asset: ba.newAsset, Cause: "buy", CausedByFactionID: ba.factionID},
		}, mutations...)
	}
	if ba.stealthTarget != "" {
		mutations = append(mutations, domain.AssetStealthApplied{
			FactionID:         ba.factionID,
			AssetID:           ba.stealthTarget,
			Cause:             "buy",
			CausedByFactionID: ba.factionID,
		})
	}
	if ba.discountConsumed {
		mutations = append(mutations, domain.PurchaseDiscountConsumed{
			FactionID:         ba.factionID,
			Cause:             "buy",
			CausedByFactionID: ba.factionID,
		})
	}
	return mutations, nil
}

// eligibleStealthTargets returns non-stealthy Special Forces assets owned by
// the faction on the given world — valid targets when buying SWN-C3-002 Stealth.
func eligibleStealthTargets(faction *domain.Faction, world string, rulebook *rulebook.Rulebook) []*domain.Asset {
	var result []*domain.Asset
	for _, asset := range faction.Assets {
		if asset.Location.WorldID != world || asset.Stealthy {
			continue
		}
		def, ok := rulebook.Assets[asset.DefinitionID]
		if ok && def.Type == domain.TypeSpecialForces {
			result = append(result, asset)
		}
	}
	return result
}

// availableWorlds returns the faction's homeworld plus the unique set of
// worlds where the faction already has assets.
func availableWorlds(faction *domain.Faction) []string {
	seen := map[string]struct{}{faction.Homeworld.WorldID: {}}
	for _, asset := range faction.Assets {
		seen[asset.Location.WorldID] = struct{}{}
	}
	worlds := make([]string, 0, len(seen))
	for world := range seen {
		worlds = append(worlds, world)
	}
	sort.Strings(worlds)
	return worlds
}

// assetsOnWorld returns the world index's assets on worldID, or nil when no index exists.
func assetsOnWorld(worldEngine *world.WorldEngine, worldID string) []*domain.Asset {
	if worldEngine == nil || worldEngine.Index == nil {
		return nil
	}
	return worldEngine.Index.AssetsByLocation[worldID]
}

// purchasableDefinitions returns the definitions the faction may buy on fragmentID via
// domain.CanPurchase: affordable at the cost Resolve will charge, meeting the minimum attribute rating, within
// the world's tech level, and holding Planetary Government for P-flagged defs. An unknown
// world imposes no tech ceiling (math.MaxInt) but cannot confer government, so P-flagged
// defs are rejected — matching the purchase_enabler path.
func purchasableDefinitions(faction *domain.Faction, rulebook *rulebook.Rulebook, fragmentID string, worldEngine *world.WorldEngine, registry *hooks.Registry) []*domain.AssetDefinition {
	worldTechLevel := math.MaxInt
	holdsGovernment := false
	if worldEngine != nil {
		loc, ok := worldEngine.Location(fragmentID)
		if !ok {
			return nil
		}
		worldTechLevel = loc.TechLevel()
		if worldEngine.Index != nil {
			holdsGovernment = domain.HasPlanetaryGovernment(faction.ID, worldEngine.Index.BasesByLocation[fragmentID])
		}
	}

	onWorld := assetsOnWorld(worldEngine, fragmentID)
	var result []*domain.AssetDefinition
	for _, def := range rulebook.Assets {
		techCap := worldTechLevel
		if worldEngine != nil {
			techCap = dispatch.ResolveWorldTechLevel(registry, faction, def, fragmentID, worldTechLevel)
		}
		cost := dispatch.ResolveAssetCost(registry, faction, def, fragmentID, def.Cost, onWorld).Cost
		if domain.CanPurchase(faction, def, domain.PurchaseTerms{Cost: cost, TechCap: techCap, HoldsGovernment: holdsGovernment}) {
			result = append(result, def)
		}
	}
	return result
}

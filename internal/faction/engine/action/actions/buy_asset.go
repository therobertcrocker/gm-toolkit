package actions

import (
	"fmt"
	"math"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// BuyAsset purchases a new asset and places it on a target world.
// The asset is flagged inactive (Ready: false) until the start of the next turn.
// A quality (Stealth) adds no asset: its effect acts on one the faction holds there.
type BuyAsset struct {
	collector        action.Collector
	registry         *hooks.Registry
	worldEngine      *world.WorldEngine
	effectsEngine    *effect.EffectsEngine
	factionID        string
	buyOrder         action.BuyOrder
	newAsset         domain.Asset
	qualityTarget    *domain.Asset
	qualityMutations []domain.Mutation
	refusal          *domain.PermissionRefused
	quote            hooks.AssetCostQuery
	cost             int
	discountConsumed bool
}

func NewBuyAsset(collector action.Collector, registry *hooks.Registry, worldEngine *world.WorldEngine, effectsEngine *effect.EffectsEngine) *BuyAsset {
	return &BuyAsset{collector: collector, registry: registry, worldEngine: worldEngine, effectsEngine: effectsEngine}
}

func (ba *BuyAsset) Name() string { return "Buy Asset" }

func (ba *BuyAsset) Validate(faction *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) bool {
	for _, worldID := range availableWorlds(faction) {
		if len(purchasableDefinitions(faction, rulebook, worldID, ba.worldEngine, ba.registry, ba.effectsEngine)) > 0 {
			return true
		}
	}
	return false
}

func (ba *BuyAsset) Inputs(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	purchasableByWorld := make(map[string][]*domain.AssetDefinition)
	for _, worldID := range availableWorlds(faction) {
		defs := purchasableDefinitions(faction, rulebook, worldID, ba.worldEngine, ba.registry, ba.effectsEngine)
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

	if order.Definition.Quality {
		target, err := ba.selectQualityTarget(faction, rulebook)
		if err != nil {
			return fmt.Errorf("buy asset: %w", err)
		}
		ba.qualityTarget = target
	}
	refusal, err := dispatch.RequestPermission(ba.registry, ba.collector, factionState, faction, order.Definition, order.World, "buy")
	if err != nil {
		return fmt.Errorf("buy asset: %w", err)
	}
	ba.refusal = refusal
	return nil
}

func (ba *BuyAsset) Resolve(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	if ba.refusal != nil {
		return nil
	}
	def := ba.buyOrder.Definition
	ba.quote = dispatch.ResolveAssetCost(ba.registry, faction, def, ba.buyOrder.World, def.Cost, world.AssetsOnWorld(ba.worldEngine, ba.buyOrder.World))
	ba.cost = ba.quote.Cost
	if faction.PendingPurchaseDiscount > 0 {
		ba.cost = dispatch.DiscountedCost(ba.quote, def.Cost, faction.PendingPurchaseDiscount)
		ba.discountConsumed = true
	}
	if faction.Coin < ba.cost {
		return fmt.Errorf("insufficient Coin: need %d, have %d", ba.cost, faction.Coin)
	}
	if def.Quality {
		ba.qualityMutations = ba.effectsEngine.FireQuality(faction, def, ba.qualityTarget, factionState)
		return nil
	}
	ba.newAsset = domain.NewPurchasedAsset(factionState.Factions, faction, def, ba.buyOrder.World)
	return nil
}

func (ba *BuyAsset) Output() ([]domain.Mutation, error) {
	if ba.refusal != nil {
		return []domain.Mutation{*ba.refusal}, nil
	}
	mutations := dispatch.PurchaseCharges(ba.factionID, ba.cost, ba.quote, "buy")
	if ba.buyOrder.Definition.Quality {
		mutations = append(mutations, ba.qualityMutations...)
	} else {
		mutations = append([]domain.Mutation{
			domain.AssetAdded{FactionID: ba.factionID, Asset: ba.newAsset, Cause: "buy", CausedByFactionID: ba.factionID},
		}, mutations...)
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

// selectQualityTarget picks the asset a quality acts on: the one candidate without
// a prompt, else the faction's choice. The menu offers a quality only where it has
// a candidate (D-22).
func (ba *BuyAsset) selectQualityTarget(faction *domain.Faction, rulebook *rulebook.Rulebook) (*domain.Asset, error) {
	def := ba.buyOrder.Definition
	targets := ba.effectsEngine.QualityTargets(def, faction, ba.buyOrder.World, rulebook)
	switch len(targets) {
	case 0:
		return nil, fmt.Errorf("%s has no target on %s", def.Name, ba.buyOrder.World)
	case 1:
		return targets[0], nil
	}
	target, err := ba.collector.SelectAsset(targets, rulebook)
	if err != nil {
		return nil, fmt.Errorf("select %s target: %w", def.Name, err)
	}
	if target == nil {
		return nil, fmt.Errorf("select %s target: %w", def.Name, action.ErrNoSelection)
	}
	return target, nil
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

// purchasableDefinitions returns the definitions the faction may buy on fragmentID via
// domain.CanPurchase: affordable at the cost Resolve will charge, meeting the minimum
// attribute rating, and within the world's tech level. A quality also needs a target on
// the world (D-22). A P-flagged def is offered everywhere; Inputs asks its permission
// after the choice (D-35). An unknown world imposes no tech ceiling (math.MaxInt).
func purchasableDefinitions(faction *domain.Faction, rulebook *rulebook.Rulebook, fragmentID string, worldEngine *world.WorldEngine, registry *hooks.Registry, effectsEngine *effect.EffectsEngine) []*domain.AssetDefinition {
	worldTechLevel := math.MaxInt
	if worldEngine != nil {
		loc, ok := worldEngine.Location(fragmentID)
		if !ok {
			return nil
		}
		worldTechLevel = loc.TechLevel()
	}

	onWorld := world.AssetsOnWorld(worldEngine, fragmentID)
	var result []*domain.AssetDefinition
	for _, def := range rulebook.Assets {
		if def.Quality && (effectsEngine == nil || len(effectsEngine.QualityTargets(def, faction, fragmentID, rulebook)) == 0) {
			continue
		}
		techCap := worldTechLevel
		if worldEngine != nil {
			techCap = dispatch.ResolveWorldTechLevel(registry, faction, def, fragmentID, worldTechLevel)
		}
		cost := dispatch.ResolveAssetCost(registry, faction, def, fragmentID, def.Cost, onWorld).Cost
		if domain.CanPurchase(faction, def, domain.PurchaseTerms{Cost: cost, TechCap: techCap}) {
			result = append(result, def)
		}
	}
	return result
}

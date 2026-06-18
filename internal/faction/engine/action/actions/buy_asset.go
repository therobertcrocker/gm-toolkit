package actions

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

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
	collector     action.Collector
	registry      *hooks.Registry
	worldEngine   *world.WorldEngine
	factionID     string
	buyOrder      action.BuyOrder
	newAsset      domain.Asset
	cost          int
	stealthTarget string // asset ID to stealth when buying SWN-C3-002; empty if no eligible target
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
	ba.cost = dispatch.ResolveAssetCost(ba.registry, faction, def, ba.buyOrder.World, def.Cost)
	if faction.Coin < ba.cost {
		return fmt.Errorf("insufficient Coin: need %d, have %d", ba.cost, faction.Coin)
	}
	ba.newAsset = domain.Asset{
		ID:           fmt.Sprintf("%s-%s-%d", ba.factionID, def.ID, nextAssetSuffix(faction, def)),
		DefinitionID: def.ID,
		OwnerID:      ba.factionID,
		Location:     domain.Location{WorldID: ba.buyOrder.World},
		CurrentHP:    def.HP,
		Ready:        false,
		Maintained:   true,
	}
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

// nextAssetSuffix returns the next unused numeric suffix for an asset of this
// definition owned by this faction. Monotonic: scans for the highest existing
// suffix matching `<factionID>-<defID>-N` and returns N+1, so IDs are never
// reused even after a sell+buy of the same definition.
func nextAssetSuffix(faction *domain.Faction, def *domain.AssetDefinition) int {
	prefix := fmt.Sprintf("%s-%s-", faction.ID, def.ID)
	highest := 0
	for _, asset := range faction.Assets {
		if !strings.HasPrefix(asset.ID, prefix) {
			continue
		}
		n, err := strconv.Atoi(strings.TrimPrefix(asset.ID, prefix))
		if err != nil {
			continue
		}
		if n > highest {
			highest = n
		}
	}
	return highest + 1
}

// purchasableDefinitions returns definitions the faction can afford, meets the
// minimum attribute rating for, passes tech level on fragmentID, and satisfies
// the P-flag requirement (Planetary Government on the target world).
func purchasableDefinitions(faction *domain.Faction, rulebook *rulebook.Rulebook, fragmentID string, worldEngine *world.WorldEngine, registry *hooks.Registry) []*domain.AssetDefinition {
	var worldTL int
	var worldKnown bool
	if worldEngine != nil {
		loc, ok := worldEngine.Location(fragmentID)
		if !ok {
			return nil
		}
		worldTL = dispatch.ResolveWorldTechLevel(registry, faction, fragmentID, loc.TechLevel())
		worldKnown = true
	}

	var result []*domain.AssetDefinition
	for _, def := range rulebook.Assets {
		if faction.Coin < def.Cost || statScore(faction, def.Category) < def.MinRating {
			continue
		}
		if worldKnown && def.TechLevel > worldTL {
			continue
		}
		if def.HasFlag(domain.FlagPermission) && worldEngine != nil && worldEngine.Index != nil {
			if !factionHasPlanetaryGovernment(faction.ID, worldEngine.Index.BasesByLocation[fragmentID]) {
				continue
			}
		}
		result = append(result, def)
	}
	return result
}

func factionHasPlanetaryGovernment(factionID string, bases []*domain.Base) bool {
	for _, base := range bases {
		if base.OwnerID == factionID {
			return true
		}
	}
	return false
}

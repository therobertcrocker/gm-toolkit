package actions

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// BuyAsset purchases a new asset and places it on a target world.
// The asset is flagged inactive (Ready: false) until the start of the next turn.
// Tech-level filtering and P-flag (government permission) checks are deferred.
type BuyAsset struct {
	collector     action.Collector
	factionID     string
	buyOrder      action.BuyOrder
	newAsset      domain.Asset
	cost          int
	stealthTarget string // asset ID to stealth when buying C3-002; empty if no eligible target
}

func NewBuyAsset(collector action.Collector) *BuyAsset {
	return &BuyAsset{collector: collector}
}

func (ba *BuyAsset) Name() string { return "Buy Asset" }

func (ba *BuyAsset) Validate(faction *domain.Faction, _ *state.FactionState, rulebook *loader.Rulebook) bool {
	for _, def := range rulebook.Assets {
		if faction.Coin >= def.Cost && statScore(faction, def.Category) >= def.MinRating {
			return true
		}
	}
	return false
}

func (ba *BuyAsset) Inputs(faction *domain.Faction, _ *state.FactionState, rulebook *loader.Rulebook) error {
	worlds := availableWorlds(faction)
	purchasable := purchasableDefinitions(faction, rulebook)

	order, err := ba.collector.SelectBuyOrder(worlds, purchasable)
	if err != nil {
		return fmt.Errorf("buy asset: %w", err)
	}
	ba.factionID = faction.ID
	ba.buyOrder = order

	if order.Definition.ID == "C3-002" {
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

func (ba *BuyAsset) Resolve(faction *domain.Faction, _ *state.FactionState, rulebook *loader.Rulebook) error {
	def := ba.buyOrder.Definition
	if faction.Coin < def.Cost {
		return fmt.Errorf("insufficient Coin: need %d, have %d", def.Cost, faction.Coin)
	}
	ba.cost = def.Cost
	ba.newAsset = domain.Asset{
		ID:           fmt.Sprintf("%s-%s-%d", ba.factionID, def.ID, nextAssetSuffix(faction, def)),
		DefinitionID: def.ID,
		OwnerID:      ba.factionID,
		Location:     ba.buyOrder.World,
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
	if ba.buyOrder.Definition.ID != "C3-002" {
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
// the faction on the given world — valid targets when buying C3-002 Stealth.
func eligibleStealthTargets(faction *domain.Faction, world string, rulebook *loader.Rulebook) []*domain.Asset {
	var result []*domain.Asset
	for _, asset := range faction.Assets {
		if asset.Location != world || asset.Stealthy {
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
	seen := map[string]struct{}{faction.Homeworld: {}}
	for _, asset := range faction.Assets {
		seen[asset.Location] = struct{}{}
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

// purchasableDefinitions returns definitions the faction can afford and
// meets the minimum attribute rating for. Tech-level and P-flag checks deferred.
func purchasableDefinitions(faction *domain.Faction, rulebook *loader.Rulebook) []*domain.AssetDefinition {
	var result []*domain.AssetDefinition
	for _, def := range rulebook.Assets {
		if faction.Coin >= def.Cost && statScore(faction, def.Category) >= def.MinRating {
			result = append(result, def)
		}
	}
	return result
}

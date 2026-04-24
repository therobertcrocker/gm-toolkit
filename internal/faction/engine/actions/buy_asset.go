package actions

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// BuyAsset purchases a new asset and places it on a target world.
// The asset is flagged inactive (Ready: false) until the start of the next turn.
// Tech-level filtering and P-flag (government permission) checks are deferred.
type BuyAsset struct {
	collector engine.InputCollector
	factionID string
	buyOrder  engine.BuyOrder
	newAsset  domain.Asset
	cost      int
}

func NewBuyAsset(collector engine.InputCollector) *BuyAsset {
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
	return nil
}

func (ba *BuyAsset) Resolve(faction *domain.Faction, _ *state.FactionState, _ *loader.Rulebook) error {
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
	return []domain.Mutation{
		domain.AssetAdded{FactionID: ba.factionID, Asset: ba.newAsset},
		domain.CoinDelta{FactionID: ba.factionID, Delta: -ba.cost},
	}, nil
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

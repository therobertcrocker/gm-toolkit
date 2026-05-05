package actions

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// SellAsset removes an asset from the faction and returns half its cost in Coin.
type SellAsset struct {
	collector     action.Collector
	selectedAsset *domain.Asset
	factionID     string
	saleValue     int
}

func NewSellAsset(collector action.Collector) *SellAsset {
	return &SellAsset{collector: collector}
}

func (sa *SellAsset) Name() string { return "Sell Asset" }

// Validate confirms the faction has at least one sellable asset.
// Bases of Influence are excluded once they are modelled.
func (sa *SellAsset) Validate(faction *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) bool {
	return len(faction.Assets) > 0
}

func (sa *SellAsset) Inputs(faction *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) error {
	selected, err := sa.collector.SelectAsset(faction.Assets, rulebook)
	if err != nil {
		return fmt.Errorf("sell asset: %w", err)
	}
	sa.selectedAsset = selected
	sa.factionID = faction.ID
	return nil
}

func (sa *SellAsset) Resolve(_ *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) error {
	def, ok := rulebook.Assets[sa.selectedAsset.DefinitionID]
	if !ok {
		return fmt.Errorf("asset definition not found: %s", sa.selectedAsset.DefinitionID)
	}
	sa.saleValue = def.Cost / 2
	return nil
}

func (sa *SellAsset) Output() ([]domain.Mutation, error) {
	return []domain.Mutation{
		domain.AssetRemoved{FactionID: sa.factionID, AssetID: sa.selectedAsset.ID, Cause: "sell", CausedByFactionID: sa.factionID},
		domain.CoinDelta{FactionID: sa.factionID, Delta: sa.saleValue, Cause: "sell", CausedByFactionID: sa.factionID},
	}, nil
}

package wizard

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func SelectStartingAssets(faction *domain.Faction, homeworld string, primary domain.FactionStat, otherStats []domain.FactionStat, ratings map[domain.FactionStat]int, scale domain.FactionScale, assets map[string]*domain.AssetDefinition) (map[string]*domain.Asset, error) {
	primaryCount, otherCount := domain.AssetCountsFromScale(scale)
	selected := make(map[string]*domain.Asset)

	// Primary attribute picks
	primaryDefs := assetsForStat(primary, ratings[primary], assets)
	if len(primaryDefs) == 0 {
		return nil, fmt.Errorf("no eligible %s assets for rating %d", primary, ratings[primary])
	}

	for i := range primaryCount {
		var selectedID string
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title(fmt.Sprintf("Primary Asset %d of %d (%s)", i+1, primaryCount, primary)).
					Options(assetOptions(primaryDefs)...).
					Value(&selectedID),
			),
		).Run(); err != nil {
			return nil, fmt.Errorf("wizard cancelled: %w", err)
		}
		def, ok := assets[selectedID]
		if !ok {
			return nil, fmt.Errorf("selected asset definition not found: %s", selectedID)
		}
		asset := newAsset(faction, selectedID, homeworld, def.HP)
		selected[asset.ID] = asset
		// we add the asset to the faction temporarily to ensure unique IDs for subsequent assets
		// this gets cleaned up when the faction is saved at the end of the wizard
		faction.Assets[asset.ID] = asset

	}

	// Other attribute picks
	otherStatOptions := make([]huh.Option[domain.FactionStat], len(otherStats))
	for i, stat := range otherStats {
		otherStatOptions[i] = huh.NewOption(string(stat), stat)
	}

	for i := range otherCount {
		var chosenStat domain.FactionStat
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[domain.FactionStat]().
					Title(fmt.Sprintf("Other Asset %d of %d — Choose Attribute", i+1, otherCount)).
					Options(otherStatOptions...).
					Value(&chosenStat),
			),
		).Run(); err != nil {
			return nil, fmt.Errorf("wizard cancelled: %w", err)
		}

		otherDefs := assetsForStat(chosenStat, ratings[chosenStat], assets)
		if len(otherDefs) == 0 {
			return nil, fmt.Errorf("no eligible %s assets for rating %d", chosenStat, ratings[chosenStat])
		}

		var selectedID string
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title(fmt.Sprintf("Other Asset %d of %d (%s)", i+1, otherCount, chosenStat)).
					Options(assetOptions(otherDefs)...).
					Value(&selectedID),
			),
		).Run(); err != nil {
			return nil, fmt.Errorf("wizard cancelled: %w", err)
		}

		def, ok := assets[selectedID]
		if !ok {
			return nil, fmt.Errorf("selected asset definition not found: %s", selectedID)
		}
		asset := newAsset(faction, selectedID, homeworld, def.HP)
		selected[asset.ID] = asset
		// we add the asset to the faction temporarily to ensure unique IDs for subsequent assets
		// this gets cleaned up when the faction is saved at the end of the wizard
		faction.Assets[asset.ID] = asset
	}

	return selected, nil
}

func assetsForStat(stat domain.FactionStat, rating int, assets map[string]*domain.AssetDefinition) []*domain.AssetDefinition {
	var result []*domain.AssetDefinition
	for _, def := range assets {
		if def.Category == stat && def.MinRating <= rating {
			result = append(result, def)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].MinRating != result[j].MinRating {
			return result[i].MinRating < result[j].MinRating
		}
		return result[i].Name < result[j].Name
	})
	return result
}

func assetOptions(defs []*domain.AssetDefinition) []huh.Option[string] {
	opts := make([]huh.Option[string], len(defs))
	for i, def := range defs {
		opts[i] = huh.NewOption(fmt.Sprintf("[%s] %s — %s", def.ID, def.Name, def.Description), def.ID)
	}
	return opts
}

func newAsset(faction *domain.Faction, definitionID, location string, hp int) *domain.Asset {
	return &domain.Asset{
		ID:           domain.NextAssetID(faction, definitionID),
		DefinitionID: definitionID,
		OwnerID:      faction.ID,
		Location:     location,
		CurrentHP:    hp,
		Ready:        true,
		Maintained:   true,
	}
}

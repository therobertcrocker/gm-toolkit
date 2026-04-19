package wizard

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func SelectStartingAssets(factionID, homeworld, primary string, otherStats []string, ratings map[string]int, scale domain.FactionScale, assets map[string]*domain.AssetDefinition) ([]*domain.Asset, error) {
	primaryCount, otherCount := domain.AssetCountsFromScale(scale)
	var selected []*domain.Asset
	assetIndex := 0

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
		selected = append(selected, newAsset(factionID, selectedID, homeworld, assetIndex, assets[selectedID].HP))
		assetIndex++
	}

	// Other attribute picks
	otherStatOptions := make([]huh.Option[string], len(otherStats))
	for i, stat := range otherStats {
		otherStatOptions[i] = huh.NewOption(stat, stat)
	}

	for i := range otherCount {
		var chosenStat string
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
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

		selected = append(selected, newAsset(factionID, selectedID, homeworld, assetIndex, assets[selectedID].HP))
		assetIndex++
	}

	return selected, nil
}

func assetsForStat(stat string, rating int, assets map[string]*domain.AssetDefinition) []*domain.AssetDefinition {
	var result []*domain.AssetDefinition
	for _, def := range assets {
		if string(def.Category) == stat && def.MinRating <= rating {
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

func newAsset(factionID, definitionID, location string, index, hp int) *domain.Asset {
	return &domain.Asset{
		ID:           fmt.Sprintf("%s-asset-%d", factionID, index),
		DefinitionID: definitionID,
		OwnerID:      factionID,
		Location:     location,
		CurrentHP:    hp,
		Ready:        true,
		Maintained:   true,
	}
}

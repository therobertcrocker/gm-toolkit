package forms

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
)

// GMCollector implements engine.InputCollector using interactive huh prompts.
type GMCollector struct{}

func NewGMCollector() *GMCollector {
	return &GMCollector{}
}

func (c *GMCollector) SelectAsset(assets []*domain.Asset, rulebook *loader.Rulebook) (*domain.Asset, error) {
	options := make([]huh.Option[*domain.Asset], 0, len(assets))
	for _, asset := range assets {
		label := asset.DefinitionID
		if def, ok := rulebook.Assets[asset.DefinitionID]; ok {
			label = fmt.Sprintf("%s (%s)", def.Name, asset.Location)
		}
		options = append(options, huh.NewOption(label, asset))
	}

	var selected *domain.Asset
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[*domain.Asset]().
				Title("Select an asset").
				Options(options...).
				Value(&selected),
		),
	).Run(); err != nil {
		return nil, fmt.Errorf("selection cancelled: %w", err)
	}
	return selected, nil
}

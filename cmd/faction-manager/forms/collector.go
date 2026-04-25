package forms

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
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

func (c *GMCollector) SelectRepairOrders(faction *domain.Faction, damagedAssets []*domain.Asset, rulebook *loader.Rulebook) ([]engine.RepairOrder, error) {
	assetOptions := make([]huh.Option[*domain.Asset], 0, len(damagedAssets))
	for _, asset := range damagedAssets {
		label := asset.DefinitionID
		if def, ok := rulebook.Assets[asset.DefinitionID]; ok {
			label = fmt.Sprintf("%s (%s) — HP: %d/%d", def.Name, asset.Location, asset.CurrentHP, def.HP)
		}
		assetOptions = append(assetOptions, huh.NewOption(label, asset))
	}

	var selected []*domain.Asset
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[*domain.Asset]().
				Title("Select assets to repair").
				Options(assetOptions...).
				Value(&selected),
		),
	).Run(); err != nil {
		return nil, fmt.Errorf("selection cancelled: %w", err)
	}
	if len(selected) == 0 {
		return nil, nil
	}

	var orders []engine.RepairOrder
	remainingCoin := faction.Coin

	for _, asset := range selected {
		maxHeals := maxAffordableHeals(remainingCoin)
		if maxHeals == 0 {
			break
		}

		healOptions := make([]huh.Option[int], maxHeals)
		for i := range maxHeals {
			cost := (i + 1) * (i + 2) / 2
			healOptions[i] = huh.NewOption(fmt.Sprintf("%d heal(s) — %d Coin", i+1, cost), i+1)
		}

		assetName := asset.DefinitionID
		if def, ok := rulebook.Assets[asset.DefinitionID]; ok {
			assetName = def.Name
		}

		var healCount int
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[int]().
					Title(fmt.Sprintf("Heals for %s", assetName)).
					Options(healOptions...).
					Value(&healCount),
			),
		).Run(); err != nil {
			return nil, fmt.Errorf("selection cancelled: %w", err)
		}

		orders = append(orders, engine.RepairOrder{Asset: asset, HealCount: healCount})
		remainingCoin -= healCount * (healCount + 1) / 2
	}

	return orders, nil
}

func (c *GMCollector) SelectBuyOrder(worlds []string, purchasable []*domain.AssetDefinition) (engine.BuyOrder, error) {
	worldOptions := make([]huh.Option[string], len(worlds))
	for i, world := range worlds {
		worldOptions[i] = huh.NewOption(world, world)
	}

	var selectedWorld string
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a world").
				Options(worldOptions...).
				Value(&selectedWorld),
		),
	).Run(); err != nil {
		return engine.BuyOrder{}, fmt.Errorf("selection cancelled: %w", err)
	}

	defOptions := make([]huh.Option[*domain.AssetDefinition], 0, len(purchasable))
	for _, def := range purchasable {
		label := fmt.Sprintf("%s [%s] — %d Coin, HP: %d", def.Name, def.Category, def.Cost, def.HP)
		defOptions = append(defOptions, huh.NewOption(label, def))
	}

	var selectedDef *domain.AssetDefinition
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[*domain.AssetDefinition]().
				Title("Select an asset to purchase").
				Options(defOptions...).
				Value(&selectedDef),
		),
	).Run(); err != nil {
		return engine.BuyOrder{}, fmt.Errorf("selection cancelled: %w", err)
	}

	return engine.BuyOrder{World: selectedWorld, Definition: selectedDef}, nil
}

func (c *GMCollector) SelectRefitOrder(options []engine.RefitOption, rulebook *loader.Rulebook) (engine.RefitOrder, error) {
	assetOptions := make([]huh.Option[*engine.RefitOption], 0, len(options))
	for index := range options {
		option := &options[index]
		label := option.Asset.DefinitionID
		if def, ok := rulebook.Assets[option.Asset.DefinitionID]; ok {
			label = fmt.Sprintf("%s (%s) — %d Coin", def.Name, option.Asset.Location, def.Cost)
		}
		assetOptions = append(assetOptions, huh.NewOption(label, option))
	}

	var selectedOption *engine.RefitOption
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[*engine.RefitOption]().
				Title("Select an asset to refit").
				Options(assetOptions...).
				Value(&selectedOption),
		),
	).Run(); err != nil {
		return engine.RefitOrder{}, fmt.Errorf("selection cancelled: %w", err)
	}

	oldDef := rulebook.Assets[selectedOption.Asset.DefinitionID]
	replacementOptions := make([]huh.Option[*domain.AssetDefinition], 0, len(selectedOption.Replacements))
	for _, def := range selectedOption.Replacements {
		delta := max(0, def.Cost-oldDef.Cost)
		label := fmt.Sprintf("%s — %d Coin (delta: +%d)", def.Name, def.Cost, delta)
		replacementOptions = append(replacementOptions, huh.NewOption(label, def))
	}

	var selectedDef *domain.AssetDefinition
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[*domain.AssetDefinition]().
				Title("Select replacement asset").
				Options(replacementOptions...).
				Value(&selectedDef),
		),
	).Run(); err != nil {
		return engine.RefitOrder{}, fmt.Errorf("selection cancelled: %w", err)
	}

	return engine.RefitOrder{OldAsset: selectedOption.Asset, NewDefinition: selectedDef}, nil
}

func (c *GMCollector) SelectAttackers(eligible []*domain.Asset, rulebook *loader.Rulebook) ([]*domain.Asset, error) {
	options := make([]huh.Option[*domain.Asset], 0, len(eligible))
	for _, asset := range eligible {
		label := asset.DefinitionID
		if def, ok := rulebook.Assets[asset.DefinitionID]; ok {
			atkLabel := ""
			if def.Attack != nil {
				atkLabel = fmt.Sprintf(", %s vs %s", def.Attack.AttackerStat, def.Attack.DefenderStat)
			}
			label = fmt.Sprintf("%s (%s) — HP: %d/%d%s", def.Name, asset.Location, asset.CurrentHP, def.HP, atkLabel)
		}
		options = append(options, huh.NewOption(label, asset))
	}

	var selected []*domain.Asset
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[*domain.Asset]().
				Title("Select attacking assets").
				Description("All selected assets commit now; sequence runs to completion.").
				Options(options...).
				Value(&selected),
		),
	).Run(); err != nil {
		return nil, fmt.Errorf("selection cancelled: %w", err)
	}
	return selected, nil
}

func (c *GMCollector) SelectDefender(attacker *domain.Asset, eligible []*domain.Asset, rulebook *loader.Rulebook) (*domain.Asset, error) {
	attackerName := attacker.DefinitionID
	if def, ok := rulebook.Assets[attacker.DefinitionID]; ok {
		attackerName = def.Name
	}

	options := make([]huh.Option[*domain.Asset], 0, len(eligible))
	for _, asset := range eligible {
		label := fmt.Sprintf("[%s] %s", asset.OwnerID, asset.DefinitionID)
		if def, ok := rulebook.Assets[asset.DefinitionID]; ok {
			counterLabel := "no counter"
			if def.Counter != nil {
				counterLabel = fmt.Sprintf("counter: %dd%d", def.Counter.NumDice, def.Counter.Sides)
			}
			label = fmt.Sprintf("[%s] %s (%s) — HP: %d/%d, %s",
				asset.OwnerID, def.Name, asset.Location, asset.CurrentHP, def.HP, counterLabel)
		}
		options = append(options, huh.NewOption(label, asset))
	}

	var selected *domain.Asset
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[*domain.Asset]().
				Title(fmt.Sprintf("Defender's choice — select target for %s", attackerName)).
				Options(options...).
				Value(&selected),
		),
	).Run(); err != nil {
		return nil, fmt.Errorf("selection cancelled: %w", err)
	}
	return selected, nil
}

func (c *GMCollector) ConfirmRedirectToBase(defenderFaction *domain.Faction, base *domain.Base, damage int) (bool, error) {
	var redirect bool
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Redirect %d damage to %s's Base on %s? (HP: %d)",
					damage, defenderFaction.Name, base.Location, base.CurrentHP)).
				Description("Defender's choice — damage to the Base is also dealt to faction HP.").
				Affirmative("Redirect to Base").
				Negative("Asset takes damage").
				Value(&redirect),
		),
	).Run(); err != nil {
		return false, fmt.Errorf("prompt cancelled: %w", err)
	}
	return redirect, nil
}

// maxAffordableHeals returns the most heals purchasable with the given Coin,
// where n heals costs n*(n+1)/2 total (1 + 2 + ... + n).
func maxAffordableHeals(coin int) int {
	n := 0
	for (n+1)*(n+2)/2 <= coin {
		n++
	}
	return n
}

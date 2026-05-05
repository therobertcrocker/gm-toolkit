package actions

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/ability"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// UseAssetAbility activates the special ability of one or more A-flagged assets.
// Assets are committed up front in resolution order; each is resolved in sequence.
type UseAssetAbility struct {
	collector      engine.InputCollector
	roller         domain.Roller
	abilityEngine  *ability.AbilityEngine
	selectedAssets []*domain.Asset
	mutations      []domain.Mutation
}

func NewUseAssetAbility(collector engine.InputCollector, roller domain.Roller, abilityEngine *ability.AbilityEngine) *UseAssetAbility {
	return &UseAssetAbility{
		collector:     collector,
		roller:        roller,
		abilityEngine: abilityEngine,
	}
}

func (u *UseAssetAbility) Name() string { return "Use Asset Ability" }

func (u *UseAssetAbility) Validate(faction *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) bool {
	for _, asset := range faction.Assets {
		def, ok := rulebook.Assets[asset.DefinitionID]
		if !ok {
			continue
		}
		if hasActionFlag(def) && asset.Ready && asset.Maintained {
			return true
		}
	}
	return false
}

func (u *UseAssetAbility) Inputs(faction *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) error {
	var candidates []*domain.Asset
	for _, asset := range faction.Assets {
		def, ok := rulebook.Assets[asset.DefinitionID]
		if !ok {
			continue
		}
		if hasActionFlag(def) && asset.Ready && asset.Maintained {
			candidates = append(candidates, asset)
		}
	}
	selected, err := u.collector.SelectAbilityAssets(faction, candidates, rulebook)
	if err != nil {
		return fmt.Errorf("use asset ability: %w", err)
	}
	u.selectedAssets = selected
	return nil
}

func (u *UseAssetAbility) Resolve(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	for _, asset := range u.selectedAssets {
		def, ok := rulebook.Assets[asset.DefinitionID]
		if !ok {
			return fmt.Errorf("use asset ability: definition not found: %s", asset.DefinitionID)
		}
		if def.Ability == nil {
			if _, err := u.collector.ConfirmAbilityApplied(asset, def); err != nil {
				return fmt.Errorf("use asset ability: confirm: %w", err)
			}
			continue
		}
		mutations, err := u.abilityEngine.Run(faction, asset, def, u.collector, u.roller, factionState, rulebook)
		if err != nil {
			return fmt.Errorf("use asset ability: %w", err)
		}
		u.mutations = append(u.mutations, mutations...)
	}
	return nil
}

func (u *UseAssetAbility) Output() ([]domain.Mutation, error) {
	return u.mutations, nil
}

func hasActionFlag(def *domain.AssetDefinition) bool {
	for _, flag := range def.Flags {
		if flag == domain.FlagAction {
			return true
		}
	}
	return false
}

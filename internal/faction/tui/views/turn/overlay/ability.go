package overlay

import (
	"fmt"

	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

// NewSelectAbilityAssets builds the multi-select of A-flagged assets whose
// abilities the faction activates (committed up front, resolved in order).
// Labels carry the ability effect. Answer: []*domain.Asset (uncapped). Esc cancels.
func NewSelectAbilityAssets(candidates []*domain.Asset, rb *rulebook.Rulebook) MultiSelectOverlay[*domain.Asset] {
	opts := make([]huh.Option[*domain.Asset], 0, len(candidates))
	for _, asset := range candidates {
		opts = append(opts, huh.NewOption(abilityLabel(asset, rb), asset))
	}
	return NewMultiSelect[*domain.Asset]("Activate which abilities?", "", opts, 0)
}

// NewConfirmAbilityApplied is the per-asset acknowledgment. The engine discards
// the returned bool (ability/dispatch.go:63), so "Skip" is inert — the modal only
// gates and paces resolution (OQ). Answer: bool. Esc cancels the action.
func NewConfirmAbilityApplied(asset *domain.Asset, def *domain.AssetDefinition) ConfirmOverlay {
	return NewConfirm(fmt.Sprintf("Apply %s ability?", def.Name), def.Description, "Apply", "Skip")
}

// NewSelectFactionTestTarget picks the opposed-test target for an ability (e.g.
// Informers' reveal-stealth). The header names the effect. Answer: *domain.Faction.
// Esc cancels.
func NewSelectFactionTestTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Faction) SelectOverlay[*domain.Faction] {
	opts := make([]huh.Option[*domain.Faction], 0, len(candidates))
	for _, faction := range candidates {
		opts = append(opts, huh.NewOption(faction.Name, faction))
	}
	return NewSelect[*domain.Faction]("Target which faction?", fmt.Sprintf("Ability effect: %s", effect), opts)
}

// abilityLabel renders "<name> · <effect>" (or just the name when the definition
// carries no ability effect — e.g. the structural-stub abilities).
func abilityLabel(asset *domain.Asset, rb *rulebook.Rulebook) string {
	name := assetName(asset, rb)
	if def, ok := rb.Assets[asset.DefinitionID]; ok && def.Ability != nil && def.Ability.Effect != "" {
		return fmt.Sprintf("%s · %s", name, def.Ability.Effect)
	}
	return name
}

package overlay

import (
	"fmt"

	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

// NewSelectAbilityHolders builds the multi-select of the holders whose abilities
// the faction activates (committed up front, resolved in order). Labels carry the
// ability effect. Answer: []domain.AbilityHolder (uncapped). Esc cancels.
func NewSelectAbilityHolders(candidates []domain.AbilityHolder, rb *rulebook.Rulebook) MultiSelectOverlay[domain.AbilityHolder] {
	opts := make([]huh.Option[domain.AbilityHolder], 0, len(candidates))
	for _, holder := range candidates {
		opts = append(opts, huh.NewOption(abilityLabel(holder, rb), holder))
	}
	return NewMultiSelect[domain.AbilityHolder]("Activate which abilities?", "", opts, 0)
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
// carries no ability effect — e.g. the structural-stub abilities). A tag's name
// is marked "(tag)".
func abilityLabel(holder domain.AbilityHolder, rb *rulebook.Rulebook) string {
	if holder.Tag != nil {
		return fmt.Sprintf("%s (tag) · %s", holder.Tag.Name, holder.Tag.Ability.Steps[0].Effect)
	}
	asset := holder.Asset
	name := assetName(asset, rb)
	if def, ok := rb.Assets[asset.DefinitionID]; ok && def.Ability != nil && len(def.Ability.Steps) > 0 {
		return fmt.Sprintf("%s · %s", name, def.Ability.Steps[0].Effect)
	}
	return name
}

// NewSelectAbilityAssetTarget picks the rival asset an ability acts on (Marketers,
// Seditionists). Labels carry the owner. Answer: *domain.Asset. Esc cancels.
func NewSelectAbilityAssetTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Asset, ownerNames map[string]string, rb *rulebook.Rulebook) SelectOverlay[*domain.Asset] {
	opts := make([]huh.Option[*domain.Asset], 0, len(candidates))
	for _, candidate := range candidates {
		owner := ownerNames[candidate.OwnerID]
		if owner == "" {
			owner = candidate.OwnerID
		}
		opts = append(opts, huh.NewOption(fmt.Sprintf("%s · %s", assetLabel(candidate, rb), owner), candidate))
	}
	return NewSelect[*domain.Asset]("Target which asset?", fmt.Sprintf("%s · %s", assetName(asset, rb), effect), opts)
}

// NewConfirmPayment asks the paying faction whether to pay. Its header names the
// faction asked, because it is often not the acting one. Esc does nothing, and a
// cancel would also drop every lift still to offer. Answer: bool.
func NewConfirmPayment(offer hooks.PaymentOffer, rb *rulebook.Rulebook) NoCancelConfirm {
	name := assetName(offer.Asset, rb)
	header := fmt.Sprintf("Asked of %s", offer.Payer.Name)
	switch offer.Reason {
	case hooks.PaymentLiftDisable:
		return NoCancelConfirm{NewConfirm(fmt.Sprintf("Pay %d Coin to bring %s back into use?", offer.Amount, name), header, "Pay", "Leave it disabled")}
	case hooks.PaymentRestore:
		return NoCancelConfirm{NewConfirm(fmt.Sprintf("Pay %d Coin to restore %s at 1 HP?", offer.Amount, name), header, "Pay", "Let it be destroyed")}
	}
	return NoCancelConfirm{NewConfirm(fmt.Sprintf("Pay %d Coin, or %s is disabled?", offer.Amount, name), header, "Pay", "Take the disable")}
}

// holderName is an ability holder's display name: its tag's, or its asset's.
func holderName(holder domain.AbilityHolder, rb *rulebook.Rulebook) string {
	if holder.Tag != nil {
		return holder.Tag.Name
	}
	return assetName(holder.Asset, rb)
}

// NewSelectAbilityWorldTarget picks the world an ability acts on (Teach Planetary
// Population). The header names the holder and the effect. Answer: string (world
// ID). Esc cancels.
func NewSelectAbilityWorldTarget(holder domain.AbilityHolder, effect domain.AbilityEffectType, worlds []string, worldNames map[string]string, rb *rulebook.Rulebook) SelectOverlay[string] {
	opts := make([]huh.Option[string], 0, len(worlds))
	for _, worldID := range worlds {
		name := worldNames[worldID]
		if name == "" {
			name = worldID
		}
		opts = append(opts, huh.NewOption(name, worldID))
	}
	return NewSelect[string]("Target which world?", fmt.Sprintf("%s · %s", holderName(holder, rb), effect), opts)
}

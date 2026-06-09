package overlay

import (
	"fmt"

	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

// NewSelectAttackers builds the multi-select of committable attackers. Each label
// carries the attacker's attack stat and damage so the GM can weigh the
// commitment. Answer: []*domain.Asset (uncapped). Esc cancels.
func NewSelectAttackers(eligible []*domain.Asset, rb *rulebook.Rulebook) MultiSelectOverlay[*domain.Asset] {
	opts := make([]huh.Option[*domain.Asset], 0, len(eligible))
	for _, asset := range eligible {
		opts = append(opts, huh.NewOption(attackerLabel(asset, rb), asset))
	}
	return NewMultiSelect[*domain.Asset]("Commit which attackers?", "", opts, 0)
}

// NewSelectDefender builds the per-attacker defender pick. The header names the
// attacker and its attack matchup; defender labels carry HP plus the owning
// faction (defenders belong to rivals). Answer: *domain.Asset. Esc cancels the
// whole attack action (the engine wraps the returned error → recoverable).
func NewSelectDefender(attacker *domain.Asset, eligible []*domain.Asset, ownerNames map[string]string, rb *rulebook.Rulebook) SelectOverlay[*domain.Asset] {
	header := fmt.Sprintf("%s attacks", assetName(attacker, rb))
	if def, ok := rb.Assets[attacker.DefinitionID]; ok && def.Attack != nil {
		header = fmt.Sprintf("%s attacks (%s vs %s)", assetName(attacker, rb), def.Attack.AttackerStat, def.Attack.DefenderStat)
	}
	opts := make([]huh.Option[*domain.Asset], 0, len(eligible))
	for _, defender := range eligible {
		owner := ownerNames[defender.OwnerID]
		if owner == "" {
			owner = defender.OwnerID
		}
		opts = append(opts, huh.NewOption(fmt.Sprintf("%s · %s", assetLabel(defender, rb), owner), defender))
	}
	return NewSelect[*domain.Asset]("Target which defender?", header, opts)
}

// NewConfirmRedirectToBase asks whether a winning hit lands on the defender's
// Base (which also damages faction HP) instead of the asset. Answer: bool. Esc
// cancels the attack action.
func NewConfirmRedirectToBase(defenderFaction *domain.Faction, base *domain.Base, damage int) ConfirmOverlay {
	header := fmt.Sprintf("Base HP %d/%d — redirecting also deals %d to %s's faction HP", base.CurrentHP, base.MaxHP, damage, defenderFaction.Name)
	title := fmt.Sprintf("Redirect %d damage to %s's base @ %s?", damage, defenderFaction.Name, base.Location.WorldID)
	return NewConfirm(title, header, "Redirect to base", "Hit the asset")
}

// attackerLabel renders "<name> · <stat> · dmg NdM(±K)" for a committable attacker.
func attackerLabel(asset *domain.Asset, rb *rulebook.Rulebook) string {
	name := assetName(asset, rb)
	def, ok := rb.Assets[asset.DefinitionID]
	if !ok || def.Attack == nil {
		return name
	}
	return fmt.Sprintf("%s · %s · dmg %s", name, def.Attack.AttackerStat, diceLabel(def.Attack.Damage))
}

func diceLabel(d domain.DiceRoll) string {
	if d.Modifier != 0 {
		return fmt.Sprintf("%dd%d%+d", d.NumDice, d.Sides, d.Modifier)
	}
	return fmt.Sprintf("%dd%d", d.NumDice, d.Sides)
}

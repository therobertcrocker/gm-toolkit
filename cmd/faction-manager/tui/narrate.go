package tui

import (
	"fmt"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/tui/style"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func narrateAction(action engine.Action, mutations []domain.Mutation, faction *domain.Faction, rulebook *loader.Rulebook) []string {
	switch action.Name() {
	case "Sell Asset":
		return narrateSell(mutations, faction, rulebook)
	case "Buy Asset":
		return narrateBuy(mutations, rulebook)
	case "Refit Asset":
		return narrateRefit(mutations, faction, rulebook)
	case "Repair Asset":
		return narrateRepairAsset(mutations, faction, rulebook)
	case "Repair Faction":
		return narrateRepairFaction(mutations)
	default:
		return nil
	}
}

func narrateExpandInfluence(mutations []domain.Mutation, faction *domain.Faction) []string {
	// Build a location index that covers both existing bases and newly added ones
	// (BaseAdded mutations haven't been applied to state yet at narration time).
	baseLocations := make(map[string]string)
	for _, base := range faction.Bases {
		baseLocations[base.ID] = base.Location
	}
	for _, m := range mutations {
		if mut, ok := m.(domain.BaseAdded); ok {
			baseLocations[mut.Base.ID] = mut.Base.Location
		}
	}

	var lines []string
	for _, m := range mutations {
		switch mut := m.(type) {
		case domain.BaseAdded:
			lines = append(lines, fmt.Sprintf("New Base placed on %s (%d HP)", mut.Base.Location, mut.Base.MaxHP))
		case domain.BaseHealed:
			lines = append(lines, fmt.Sprintf("Base on %s: +%d HP healed", baseLocations[mut.BaseID], mut.Delta))
		case domain.BaseExpanded:
			lines = append(lines, fmt.Sprintf("Base on %s: max HP +%d", baseLocations[mut.BaseID], mut.Delta))
		case domain.CoinDelta:
			if mut.Delta < 0 {
				lines = append(lines, fmt.Sprintf("Cost: %d Coin", -mut.Delta))
			}
		case domain.BaseHPDelta:
			if mut.Delta < 0 {
				lines = append(lines, fmt.Sprintf("Rival attack: base at %s took %d damage", baseLocations[mut.BaseID], -mut.Delta))
			}
		case domain.BaseDestroyed:
			lines = append(lines, fmt.Sprintf("Base at %s destroyed", baseLocations[mut.BaseID]))
		case domain.FactionHPDelta:
			if mut.Delta < 0 && mut.FactionID == faction.ID {
				lines = append(lines, fmt.Sprintf("Faction HP: %d damage from base attack", -mut.Delta))
			}
		}
	}
	return lines
}

func narrateSell(mutations []domain.Mutation, faction *domain.Faction, rulebook *loader.Rulebook) []string {
	name := "unknown"
	coin := 0
	for _, m := range mutations {
		switch mut := m.(type) {
		case domain.AssetRemoved:
			if asset := findAssetByID(mut.AssetID, faction); asset != nil {
				name = assetDisplayName(asset, rulebook)
			}
		case domain.CoinDelta:
			coin = mut.Delta
		}
	}
	return []string{fmt.Sprintf("Sold: %s (+%d Coin)", name, coin)}
}

func narrateBuy(mutations []domain.Mutation, rulebook *loader.Rulebook) []string {
	for _, m := range mutations {
		if mut, ok := m.(domain.AssetAdded); ok {
			name := mut.Asset.DefinitionID
			if def, ok := rulebook.Assets[mut.Asset.DefinitionID]; ok {
				name = def.Name
			}
			return []string{fmt.Sprintf("Bought: %s on %s", name, mut.Asset.Location)}
		}
	}
	return nil
}

func narrateRefit(mutations []domain.Mutation, faction *domain.Faction, rulebook *loader.Rulebook) []string {
	oldName := "unknown"
	newName := "unknown"
	for _, m := range mutations {
		switch mut := m.(type) {
		case domain.AssetRemoved:
			if asset := findAssetByID(mut.AssetID, faction); asset != nil {
				oldName = assetDisplayName(asset, rulebook)
			}
		case domain.AssetAdded:
			if def, ok := rulebook.Assets[mut.Asset.DefinitionID]; ok {
				newName = def.Name
			}
		}
	}
	return []string{fmt.Sprintf("Refitted: %s → %s", oldName, newName)}
}

func narrateRepairAsset(mutations []domain.Mutation, faction *domain.Faction, rulebook *loader.Rulebook) []string {
	var lines []string
	totalCost := 0
	for _, m := range mutations {
		switch mut := m.(type) {
		case domain.AssetHPDelta:
			if asset := findAssetByID(mut.AssetID, faction); asset != nil {
				lines = append(lines, fmt.Sprintf("%s: +%d HP", assetDisplayName(asset, rulebook), mut.Delta))
			}
		case domain.CoinDelta:
			totalCost = -mut.Delta
		}
	}
	if totalCost > 0 {
		lines = append(lines, fmt.Sprintf("Cost: %d Coin", totalCost))
	}
	return lines
}

func narrateRepairFaction(mutations []domain.Mutation) []string {
	for _, m := range mutations {
		if mut, ok := m.(domain.FactionHPDelta); ok {
			return []string{fmt.Sprintf("Faction HP restored: +%d", mut.Delta)}
		}
	}
	return nil
}

func narrateUseAssetAbility(mutations []domain.Mutation, faction *domain.Faction, factionState *state.FactionState, rulebook *loader.Rulebook) []string {
	if len(mutations) == 0 {
		return []string{"Ability applied (GM adjudicated)"}
	}
	var lines []string
	for _, m := range mutations {
		switch mut := m.(type) {
		case domain.AssetMoved:
			name := assetNameFromState(mut.AssetID, factionState, rulebook)
			lines = append(lines, fmt.Sprintf("%s relocated to %s", name, mut.ToLocation))
		case domain.CoinDelta:
			if mut.FactionID == faction.ID {
				if mut.Delta < 0 {
					lines = append(lines, fmt.Sprintf("Cost: %d Coin", -mut.Delta))
				} else if mut.Delta > 0 {
					lines = append(lines, fmt.Sprintf("%s gained %d Coin", faction.Name, mut.Delta))
				}
			} else if mut.Delta < 0 {
				targetName := mut.FactionID
				if f, ok := factionState.Factions[mut.FactionID]; ok {
					targetName = f.Name
				}
				lines = append(lines, fmt.Sprintf("%s lost %d Coin", targetName, -mut.Delta))
			}
		case domain.AssetStealthCleared:
			name := assetNameFromState(mut.AssetID, factionState, rulebook)
			lines = append(lines, fmt.Sprintf("%s revealed", name))
		}
	}
	return lines
}

func narrateAttack(collector *TUICollector, mutations []domain.Mutation, factionState *state.FactionState, rulebook *loader.Rulebook) []string {
	assetDeltas := make(map[string]int)
	assetDestroyedMap := make(map[string]bool)
	baseDeltas := make(map[string]int)
	baseDestroyedMap := make(map[string]bool)
	factionBaseRedirects := make(map[string]bool)

	for _, m := range mutations {
		switch mut := m.(type) {
		case domain.AssetHPDelta:
			assetDeltas[mut.AssetID] += mut.Delta
		case domain.AssetRemoved:
			assetDestroyedMap[mut.AssetID] = true
		case domain.BaseHPDelta:
			baseDeltas[mut.BaseID] += mut.Delta
			factionBaseRedirects[mut.FactionID] = true
		case domain.BaseDestroyed:
			baseDestroyedMap[mut.BaseID] = true
		}
	}

	var lines []string

	for _, attacker := range collector.attackers {
		defender := collector.defenders[attacker.ID]
		if defender == nil {
			continue
		}

		attackerName := assetNameFromState(attacker.ID, factionState, rulebook)
		defenderName := assetNameFromState(defender.ID, factionState, rulebook)
		defenderFaction := ownerFactionOf(defender.ID, factionState)
		attackerFaction := ownerFactionOf(attacker.ID, factionState)

		defDelta := assetDeltas[defender.ID]
		attackDelta := assetDeltas[attacker.ID]
		defHadRedirect := defenderFaction != nil && factionBaseRedirects[defenderFaction.ID]

		var outcome string
		switch {
		case defDelta < 0 && attackDelta >= 0:
			outcome = fmt.Sprintf("%s wins! %d damage", attackerFaction.Name, -defDelta)
		case defHadRedirect:
			outcome = fmt.Sprintf("%s wins! %s redirects damage to base", attackerFaction.Name, defenderFaction.Name)
		case defDelta < 0 && attackDelta < 0:
			outcome = fmt.Sprintf("Tie! %d damage from attack", -defDelta)
		default:
			outcome = fmt.Sprintf("%s holds", defenderName)
		}

		if defenderFaction != nil {
			lines = append(lines, fmt.Sprintf("%s → %s (%s): %s", attackerName, defenderName, defenderFaction.Name, outcome))
		} else {
			lines = append(lines, fmt.Sprintf("%s → %s: %s", attackerName, defenderName, outcome))
		}

		if assetDestroyedMap[defender.ID] {
			lines = append(lines, fmt.Sprintf("  %s (%s) destroyed", defenderName, defenderFaction.Name))
		}
		if attackDelta < 0 {
			lines = append(lines, fmt.Sprintf("  Counter: %d damage to %s", -attackDelta, attackerName))
			if assetDestroyedMap[attacker.ID] {
				lines = append(lines, fmt.Sprintf("  %s (%s) destroyed", attackerName, attackerFaction.Name))
			}
		}
	}

	for baseID, delta := range baseDeltas {
		if delta >= 0 {
			continue
		}
		base := findBaseByID(baseID, factionState)
		location := baseID
		if base != nil {
			location = base.Location
		}
		lines = append(lines, fmt.Sprintf("Base at %s: %d damage received", location, -delta))
		if baseDestroyedMap[baseID] {
			lines = append(lines, fmt.Sprintf("  Base at %s destroyed", location))
		}
	}

	if len(lines) == 0 {
		return []string{"Attack: no engagements resolved"}
	}
	return lines
}

func narrateGoalEvents(mutations []domain.Mutation, rulebook *loader.Rulebook) []string {
	var lines []string
	for _, m := range mutations {
		switch mut := m.(type) {
		case domain.GoalCompleted:
			name := mut.GoalID
			if goal, ok := rulebook.Goals[mut.GoalID]; ok {
				name = goal.Name
			}
			if mut.XPAwarded > 0 {
				lines = append(lines, fmt.Sprintf("Goal completed: %s (+%d XP)", name, mut.XPAwarded))
			} else {
				lines = append(lines, fmt.Sprintf("Goal completed: %s", name))
			}
		case domain.GoalAbandoned:
			name := mut.GoalID
			if goal, ok := rulebook.Goals[mut.GoalID]; ok {
				name = goal.Name
			}
			lines = append(lines, fmt.Sprintf("Goal abandoned: %s (income forfeited)", name))
		case domain.HomeworldChanged:
			lines = append(lines, fmt.Sprintf("Homeworld: %s → %s", mut.FromWorld, mut.ToWorld))
		case domain.TagAdded:
			lines = append(lines, fmt.Sprintf("Tag gained: %s", mut.Tag.Name))
		}
	}
	return lines
}

func renderLogSection(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(style.Muted.Render("─── Play-by-play ───"))
	sb.WriteString("\n")
	for _, line := range lines {
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return sb.String()
}

func logSummary(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return lines[0]
}

func findAssetByID(assetID string, faction *domain.Faction) *domain.Asset {
	for _, asset := range faction.Assets {
		if asset.ID == assetID {
			return asset
		}
	}
	return nil
}

func assetNameFromState(assetID string, factionState *state.FactionState, rulebook *loader.Rulebook) string {
	for _, faction := range factionState.Factions {
		if asset := findAssetByID(assetID, faction); asset != nil {
			return assetDisplayName(asset, rulebook)
		}
	}
	return assetID
}

func assetDisplayName(asset *domain.Asset, rulebook *loader.Rulebook) string {
	if def, ok := rulebook.Assets[asset.DefinitionID]; ok {
		return def.Name
	}
	return asset.DefinitionID
}

func ownerFactionOf(assetID string, factionState *state.FactionState) *domain.Faction {
	for _, faction := range factionState.Factions {
		for _, asset := range faction.Assets {
			if asset.ID == assetID {
				return faction
			}
		}
	}
	return nil
}

func findBaseByID(baseID string, factionState *state.FactionState) *domain.Base {
	for _, faction := range factionState.Factions {
		for _, base := range faction.Bases {
			if base.ID == baseID {
				return base
			}
		}
	}
	return nil
}

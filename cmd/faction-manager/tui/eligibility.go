package tui

import (
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func tuiStatScore(faction *domain.Faction, category domain.FactionStat) int {
	switch category {
	case domain.StatForce:
		return faction.Force
	case domain.StatCunning:
		return faction.Cunning
	case domain.StatWealth:
		return faction.Wealth
	default:
		return 0
	}
}

func tuiAvailableWorlds(faction *domain.Faction) []string {
	seen := map[string]struct{}{faction.Homeworld: {}}
	for _, asset := range faction.Assets {
		seen[asset.Location] = struct{}{}
	}
	worlds := make([]string, 0, len(seen))
	for world := range seen {
		worlds = append(worlds, world)
	}
	sort.Strings(worlds)
	return worlds
}

func tuiPurchasableDefinitions(faction *domain.Faction, rulebook *loader.Rulebook) []*domain.AssetDefinition {
	var result []*domain.AssetDefinition
	for _, def := range rulebook.Assets {
		if faction.Coin >= def.Cost && tuiStatScore(faction, def.Category) >= def.MinRating {
			result = append(result, def)
		}
	}
	return result
}

func tuiRefitOptions(faction *domain.Faction, rulebook *loader.Rulebook) []engine.RefitOption {
	var options []engine.RefitOption
	for _, asset := range faction.Assets {
		replacements := tuiValidRefitReplacements(faction, asset, rulebook)
		if len(replacements) == 0 {
			continue
		}
		options = append(options, engine.RefitOption{Asset: asset, Replacements: replacements})
	}
	return options
}

func tuiEligibleAttackers(faction *domain.Faction, factionState *state.FactionState, rulebook *loader.Rulebook) []*domain.Asset {
	var result []*domain.Asset
	for _, asset := range faction.Assets {
		if !asset.Ready || asset.CurrentHP <= 0 || !asset.Maintained {
			continue
		}
		def, ok := rulebook.Assets[asset.DefinitionID]
		if !ok || def.Attack == nil {
			continue
		}
		if tuiHasEligibleTarget(asset, faction.ID, factionState) {
			result = append(result, asset)
		}
	}
	return result
}

func tuiHasEligibleTarget(attacker *domain.Asset, attackerFactionID string, factionState *state.FactionState) bool {
	for factionID, faction := range factionState.Factions {
		if factionID == attackerFactionID {
			continue
		}
		for _, asset := range faction.Assets {
			if asset.Location == attacker.Location && !asset.Stealthy && asset.Ready && asset.CurrentHP > 0 && asset.Maintained {
				return true
			}
		}
	}
	return false
}

func tuiValidRefitReplacements(faction *domain.Faction, oldAsset *domain.Asset, rulebook *loader.Rulebook) []*domain.AssetDefinition {
	oldDef, ok := rulebook.Assets[oldAsset.DefinitionID]
	if !ok {
		return nil
	}
	var result []*domain.AssetDefinition
	for _, def := range rulebook.Assets {
		if def.ID == oldDef.ID {
			continue
		}
		if def.Category != oldDef.Category {
			continue
		}
		if tuiStatScore(faction, def.Category) < def.MinRating {
			continue
		}
		delta := max(0, def.Cost-oldDef.Cost)
		if faction.Coin < delta {
			continue
		}
		result = append(result, def)
	}
	return result
}

package goals

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func completeGoal(faction *domain.Faction, xp int) []domain.Mutation {
	goalID := faction.ActiveGoal.GoalID
	return []domain.Mutation{
		domain.GoalCompleted{
			FactionID: faction.ID,
			GoalID:    goalID,
			XPAwarded: xp,
			Cause:     "goal_completed",
		},
		domain.XPAwarded{
			FactionID: faction.ID,
			Amount:    xp,
			Cause:     "goal_completed",
		},
	}
}

func countAssetKillsByCategory(actingFactionID string, category domain.FactionStat, mutations []domain.Mutation, factionState *state.FactionState, rulebook *rulebook.Rulebook) int {
	count := 0
	for _, mutation := range mutations {
		v, ok := mutation.(domain.AssetRemoved)
		if !ok || v.Cause != "attack" || v.CausedByFactionID != actingFactionID || v.FactionID == actingFactionID {
			continue
		}
		rivalFaction, ok := factionState.Factions[v.FactionID]
		if !ok {
			continue
		}
		asset := findAsset(rivalFaction, v.AssetID)
		if asset == nil {
			continue
		}
		def, ok := rulebook.Assets[asset.DefinitionID]
		if !ok {
			continue
		}
		if def.Category == category {
			count++
		}
	}
	return count
}

func findAsset(faction *domain.Faction, assetID string) *domain.Asset {
	return faction.Assets[assetID]
}

func factionHasBaseOn(faction *domain.Faction, world string) bool {
	for _, base := range faction.Bases {
		if base.Location == world {
			return true
		}
	}
	return false
}

func worldHasRivalPresence(locationID, actingFactionID string, index *world.Index) bool {
	for _, asset := range index.AssetsByLocation[locationID] {
		if asset.OwnerID != actingFactionID {
			return true
		}
	}
	for _, base := range index.BasesByLocation[locationID] {
		if base.OwnerID != actingFactionID {
			return true
		}
	}
	return false
}

func rivalHasPlanetaryGovernmentOnWorld(actingFactionID, locationID string, factionState *state.FactionState, index *world.Index) bool {
	for _, base := range index.BasesByLocation[locationID] {
		if base.OwnerID == actingFactionID {
			continue
		}
		rival, ok := factionState.Factions[base.OwnerID]
		if !ok {
			continue
		}
		if factionHasPlanetaryGovernmentTag(rival) {
			return true
		}
	}
	return false
}

func factionHasPlanetaryGovernmentTag(faction *domain.Faction) bool {
	for _, tag := range faction.Tags {
		if tag.ID == "T-011" {
			return true
		}
	}
	return false
}

func factionHasUnstealthedAssetOn(faction *domain.Faction, world string) bool {
	for _, asset := range faction.Assets {
		if asset.Location == world && !asset.Stealthy {
			return true
		}
	}
	return false
}

func calcPlanetarySeizureXP(targetFactionID string, factionState *state.FactionState) int {
	if targetFactionID == "" {
		return 1
	}
	targetFaction, ok := factionState.Factions[targetFactionID]
	if !ok {
		return 1
	}
	avg := (targetFaction.Force + targetFaction.Cunning + targetFaction.Wealth) / 3
	xp := avg / 2
	if xp < 1 {
		return 1
	}
	return xp
}

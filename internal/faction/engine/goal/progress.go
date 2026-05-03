package goal

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func progressMilitaryConquest(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, rulebook *loader.Rulebook) []domain.Mutation {
	kills := countAssetKillsByCategory(actingFaction.ID, domain.StatForce, mutations, factionState, rulebook)
	if kills == 0 {
		return nil
	}
	newProgress := actingFaction.ActiveGoal.Progress + kills
	progressed := domain.GoalProgressed{
		FactionID: actingFaction.ID,
		GoalID:    actingFaction.ActiveGoal.GoalID,
		Delta:     kills,
		Cause:     "military_conquest",
	}
	if newProgress < actingFaction.Force {
		return []domain.Mutation{progressed}
	}
	return append([]domain.Mutation{progressed}, completeGoal(actingFaction, newProgress/2)...)
}

func progressCommercialExpansion(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, rulebook *loader.Rulebook) []domain.Mutation {
	kills := countAssetKillsByCategory(actingFaction.ID, domain.StatWealth, mutations, factionState, rulebook)
	if kills == 0 {
		return nil
	}
	newProgress := actingFaction.ActiveGoal.Progress + kills
	progressed := domain.GoalProgressed{
		FactionID: actingFaction.ID,
		GoalID:    actingFaction.ActiveGoal.GoalID,
		Delta:     kills,
		Cause:     "commercial_expansion",
	}
	if newProgress < actingFaction.Wealth {
		return []domain.Mutation{progressed}
	}
	return append([]domain.Mutation{progressed}, completeGoal(actingFaction, newProgress/2)...)
}

func progressIntelligenceCoup(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, rulebook *loader.Rulebook) []domain.Mutation {
	kills := countAssetKillsByCategory(actingFaction.ID, domain.StatCunning, mutations, factionState, rulebook)
	if kills == 0 {
		return nil
	}
	newProgress := actingFaction.ActiveGoal.Progress + kills
	progressed := domain.GoalProgressed{
		FactionID: actingFaction.ID,
		GoalID:    actingFaction.ActiveGoal.GoalID,
		Delta:     kills,
		Cause:     "intelligence_coup",
	}
	if newProgress < actingFaction.Cunning {
		return []domain.Mutation{progressed}
	}
	return append([]domain.Mutation{progressed}, completeGoal(actingFaction, newProgress/2)...)
}

func progressPlanetarySeizure(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState) []domain.Mutation {
	goal := actingFaction.ActiveGoal
	if goal.ProcessPhase != 1 {
		return nil
	}
	removedIDs := make(map[string]bool)
	for _, mutation := range mutations {
		if v, ok := mutation.(domain.AssetRemoved); ok {
			removedIDs[v.AssetID] = true
		}
	}
	for factionID, faction := range factionState.Factions {
		if factionID == actingFaction.ID {
			continue
		}
		for _, asset := range faction.Assets {
			if asset.Location == goal.TargetWorld && !asset.Stealthy && !removedIDs[asset.ID] {
				return nil
			}
		}
	}
	return []domain.Mutation{
		domain.GoalPhaseAdvanced{
			FactionID:      actingFaction.ID,
			GoalID:         goal.GoalID,
			ProcessPhase:   2,
			TurnsRemaining: 3,
			Cause:          "planetary_seizure_phase_advance",
		},
	}
}

func progressExpandInfluence(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState) []domain.Mutation {
	for _, mutation := range mutations {
		v, ok := mutation.(domain.BaseAdded)
		if !ok || v.CausedByFactionID != actingFaction.ID {
			continue
		}
		if factionHasBaseOn(actingFaction, v.Base.Location) {
			continue
		}
		xp := 1
		if worldHasRivalPresence(v.Base.Location, actingFaction.ID, factionState) {
			xp = 2
		}
		return completeGoal(actingFaction, xp)
	}
	return nil
}

func progressBloodTheEnemy(actingFaction *domain.Faction, mutations []domain.Mutation) []domain.Mutation {
	damage := 0
	for _, mutation := range mutations {
		switch v := mutation.(type) {
		case domain.AssetHPDelta:
			if v.CausedByFactionID == actingFaction.ID && v.FactionID != actingFaction.ID && v.Delta < 0 {
				damage += -v.Delta
			}
		case domain.BaseHPDelta:
			if v.CausedByFactionID == actingFaction.ID && v.FactionID != actingFaction.ID && v.Delta < 0 {
				damage += -v.Delta
			}
		}
	}
	if damage == 0 {
		return nil
	}
	newProgress := actingFaction.ActiveGoal.Progress + damage
	threshold := actingFaction.Force + actingFaction.Cunning + actingFaction.Wealth
	progressed := domain.GoalProgressed{
		FactionID: actingFaction.ID,
		GoalID:    actingFaction.ActiveGoal.GoalID,
		Delta:     damage,
		Cause:     "blood_the_enemy",
	}
	if newProgress < threshold {
		return []domain.Mutation{progressed}
	}
	return append([]domain.Mutation{progressed}, completeGoal(actingFaction, 2)...)
}

func progressPeaceableKingdom(actingFaction *domain.Faction, mutations []domain.Mutation) []domain.Mutation {
	attacked := false
	for _, mutation := range mutations {
		switch v := mutation.(type) {
		case domain.AssetHPDelta:
			if v.Cause == "attack" && v.CausedByFactionID == actingFaction.ID {
				attacked = true
			}
		case domain.BaseHPDelta:
			if v.Cause == "attack" && v.CausedByFactionID == actingFaction.ID {
				attacked = true
			}
		case domain.AssetStealthCleared:
			if v.Cause == "attack" && v.CausedByFactionID == actingFaction.ID && v.FactionID == actingFaction.ID {
				attacked = true
			}
		}
	}
	if attacked {
		if actingFaction.ActiveGoal.Progress == 0 {
			return nil
		}
		return []domain.Mutation{domain.GoalProgressed{
			FactionID: actingFaction.ID,
			GoalID:    actingFaction.ActiveGoal.GoalID,
			Delta:     -actingFaction.ActiveGoal.Progress,
			Cause:     "peaceable_kingdom_reset",
		}}
	}
	newProgress := actingFaction.ActiveGoal.Progress + 1
	progressed := domain.GoalProgressed{
		FactionID: actingFaction.ID,
		GoalID:    actingFaction.ActiveGoal.GoalID,
		Delta:     1,
		Cause:     "peaceable_kingdom",
	}
	if newProgress < 4 {
		return []domain.Mutation{progressed}
	}
	return append([]domain.Mutation{progressed}, completeGoal(actingFaction, 1)...)
}

func progressDestroyTheFoe(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState) []domain.Mutation {
	targetFaction, ok := factionState.Factions[actingFaction.ActiveGoal.TargetFactionID]
	if !ok {
		return nil
	}
	effectiveHP := targetFaction.CurrentHP
	for _, mutation := range mutations {
		if v, ok := mutation.(domain.FactionHPDelta); ok && v.FactionID == targetFaction.ID {
			effectiveHP += v.Delta
		}
	}
	if effectiveHP > 0 {
		return nil
	}
	avg := (targetFaction.Force + targetFaction.Cunning + targetFaction.Wealth) / 3
	return completeGoal(actingFaction, 1+avg)
}

func progressInsideEnemyTerritory(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState) []domain.Mutation {
	gained := 0
	for _, mutation := range mutations {
		v, ok := mutation.(domain.AssetStealthApplied)
		if !ok || v.FactionID != actingFaction.ID {
			continue
		}
		asset := findAsset(actingFaction, v.AssetID)
		if asset == nil {
			continue
		}
		if !rivalHasPlanetaryGovernmentOnWorld(actingFaction.ID, asset.Location, factionState) {
			continue
		}
		gained++
	}
	if gained == 0 {
		return nil
	}
	newProgress := actingFaction.ActiveGoal.Progress + gained
	progressed := domain.GoalProgressed{
		FactionID: actingFaction.ID,
		GoalID:    actingFaction.ActiveGoal.GoalID,
		Delta:     gained,
		Cause:     "inside_enemy_territory",
	}
	if newProgress < actingFaction.Cunning {
		return []domain.Mutation{progressed}
	}
	return append([]domain.Mutation{progressed}, completeGoal(actingFaction, 2)...)
}

func progressInvincibleValor(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, rulebook *loader.Rulebook) []domain.Mutation {
	for _, mutation := range mutations {
		v, ok := mutation.(domain.AssetRemoved)
		if !ok || v.Cause != "attack" || v.CausedByFactionID != actingFaction.ID || v.FactionID == actingFaction.ID {
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
		if def.Category == domain.StatForce && def.MinRating > actingFaction.Force {
			return completeGoal(actingFaction, 2)
		}
	}
	return nil
}

func progressWealthOfWorlds(actingFaction *domain.Faction, mutations []domain.Mutation) []domain.Mutation {
	spent := 0
	for _, mutation := range mutations {
		if v, ok := mutation.(domain.InfluenceDelta); ok && v.CausedByFactionID == actingFaction.ID && v.Delta > 0 {
			spent += v.Delta
		}
	}
	if spent == 0 {
		return nil
	}
	newProgress := actingFaction.ActiveGoal.Progress + spent
	progressed := domain.GoalProgressed{
		FactionID: actingFaction.ID,
		GoalID:    actingFaction.ActiveGoal.GoalID,
		Delta:     spent,
		Cause:     "wealth_of_worlds",
	}
	if newProgress < 4*actingFaction.Wealth {
		return []domain.Mutation{progressed}
	}
	return append([]domain.Mutation{progressed}, completeGoal(actingFaction, 2)...)
}

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

func countAssetKillsByCategory(actingFactionID string, category domain.FactionStat, mutations []domain.Mutation, factionState *state.FactionState, rulebook *loader.Rulebook) int {
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
	for _, asset := range faction.Assets {
		if asset.ID == assetID {
			return asset
		}
	}
	return nil
}

func factionHasBaseOn(faction *domain.Faction, world string) bool {
	for _, base := range faction.Bases {
		if base.Location == world {
			return true
		}
	}
	return false
}

func worldHasRivalPresence(world, actingFactionID string, factionState *state.FactionState) bool {
	for factionID, faction := range factionState.Factions {
		if factionID == actingFactionID {
			continue
		}
		for _, asset := range faction.Assets {
			if asset.Location == world {
				return true
			}
		}
		for _, base := range faction.Bases {
			if base.Location == world {
				return true
			}
		}
	}
	return false
}

func rivalHasPlanetaryGovernmentOnWorld(actingFactionID, world string, factionState *state.FactionState) bool {
	for factionID, faction := range factionState.Factions {
		if factionID == actingFactionID {
			continue
		}
		if !factionHasPlanetaryGovernmentTag(faction) {
			continue
		}
		for _, base := range faction.Bases {
			if base.Location == world {
				return true
			}
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

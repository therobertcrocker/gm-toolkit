package goals

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type InsideEnemyTerritory struct{}

func (InsideEnemyTerritory) GoalID() string { return "G-009" }

func (InsideEnemyTerritory) CheckLock(_ *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) (locks.GoalLock, []domain.Mutation) {
	return locks.GoalLock{Type: locks.LockNone}, nil
}

// UpdateProgress counts each asset stealthed this action on a rival's ground
// (Effort 4 Decision 6). An asset in transit has no world and does not count.
func (InsideEnemyTerritory) UpdateProgress(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook, _ *world.Index) []domain.Mutation {
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
		if asset.Location.WorldID == "" || !isEnemyTerritory(actingFaction.ID, asset.Location.WorldID, factionState) {
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

// isEnemyTerritory reports whether worldID is a rival's ground: a rival governs it,
// or it is independent and a rival's homeworld. A world the acting faction governs
// never is, a seized rival homeworld included.
func isEnemyTerritory(actingFactionID, worldID string, factionState *state.FactionState) bool {
	if !factionState.IsIndependent(worldID) {
		return factionState.Ruler(worldID) != actingFactionID
	}
	for factionID, faction := range factionState.Factions {
		if factionID != actingFactionID && faction.Homeworld.WorldID == worldID {
			return true
		}
	}
	return false
}

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

func (InsideEnemyTerritory) UpdateProgress(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook, index *world.Index) []domain.Mutation {
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
		if !rivalHasPlanetaryGovernmentOnWorld(actingFaction.ID, asset.Location, factionState, index) {
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

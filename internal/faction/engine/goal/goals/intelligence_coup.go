package goals

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type IntelligenceCoup struct{}

func (IntelligenceCoup) GoalID() string { return "G-003" }

func (IntelligenceCoup) CheckLock(_ *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) (locks.GoalLock, []domain.Mutation) {
	return locks.GoalLock{Type: locks.LockNone}, nil
}

func (IntelligenceCoup) UpdateProgress(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, rulebook *rulebook.Rulebook, _ *world.Index) []domain.Mutation {
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

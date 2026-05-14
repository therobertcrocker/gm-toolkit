package goals

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type WealthOfWorlds struct{}

func (WealthOfWorlds) GoalID() string { return "G-011" }

func (WealthOfWorlds) CheckLock(_ *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) (locks.GoalLock, []domain.Mutation) {
	return locks.GoalLock{Type: locks.LockNone}, nil
}

func (WealthOfWorlds) UpdateProgress(actingFaction *domain.Faction, mutations []domain.Mutation, _ *state.FactionState, _ *rulebook.Rulebook, _ *world.Index) []domain.Mutation {
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

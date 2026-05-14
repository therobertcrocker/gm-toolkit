package goals

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type ChangeHomeworld struct{}

func (ChangeHomeworld) GoalID() string { return "G-012" }

func (ChangeHomeworld) CheckLock(faction *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) (locks.GoalLock, []domain.Mutation) {
	newTurns := faction.ActiveGoal.TurnsRemaining - 1
	tick := domain.GoalTurnsTick{
		FactionID: faction.ID,
		GoalID:    faction.ActiveGoal.GoalID,
		Cause:     "change_homeworld_transit",
	}
	if newTurns == 0 {
		return locks.GoalLock{Type: locks.LockSkip}, []domain.Mutation{
			tick,
			domain.HomeworldChanged{
				FactionID: faction.ID,
				FromWorld: faction.Homeworld,
				ToWorld:   faction.ActiveGoal.TargetWorld,
				Cause:     "goal_completed",
			},
			domain.GoalCompleted{
				FactionID: faction.ID,
				GoalID:    faction.ActiveGoal.GoalID,
				XPAwarded: 0,
			},
		}
	}
	return locks.GoalLock{Type: locks.LockSkip}, []domain.Mutation{tick}
}

func (ChangeHomeworld) UpdateProgress(_ *domain.Faction, _ []domain.Mutation, _ *state.FactionState, _ *rulebook.Rulebook, _ *world.Index) []domain.Mutation {
	return nil
}

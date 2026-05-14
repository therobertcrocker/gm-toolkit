package goals

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type DestroyTheFoe struct{}

func (DestroyTheFoe) GoalID() string { return "G-008" }

func (DestroyTheFoe) CheckLock(_ *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) (locks.GoalLock, []domain.Mutation) {
	return locks.GoalLock{Type: locks.LockNone}, nil
}

func (DestroyTheFoe) UpdateProgress(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook, _ *world.Index) []domain.Mutation {
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

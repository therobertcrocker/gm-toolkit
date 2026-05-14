package goals

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type BloodTheEnemy struct{}

func (BloodTheEnemy) GoalID() string { return "G-006" }

func (BloodTheEnemy) CheckLock(_ *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) (locks.GoalLock, []domain.Mutation) {
	return locks.GoalLock{Type: locks.LockNone}, nil
}

func (BloodTheEnemy) UpdateProgress(actingFaction *domain.Faction, mutations []domain.Mutation, _ *state.FactionState, _ *rulebook.Rulebook, _ *world.Index) []domain.Mutation {
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

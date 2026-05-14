package goals

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type PeaceableKingdom struct{}

func (PeaceableKingdom) GoalID() string { return "G-007" }

func (PeaceableKingdom) CheckLock(_ *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) (locks.GoalLock, []domain.Mutation) {
	return locks.GoalLock{Type: locks.LockNone}, nil
}

func (PeaceableKingdom) UpdateProgress(actingFaction *domain.Faction, mutations []domain.Mutation, _ *state.FactionState, _ *rulebook.Rulebook, _ *world.Index) []domain.Mutation {
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

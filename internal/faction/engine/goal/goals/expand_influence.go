package goals

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type ExpandInfluence struct{}

func (ExpandInfluence) GoalID() string { return "G-005" }

func (ExpandInfluence) CheckLock(_ *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) (locks.GoalLock, []domain.Mutation) {
	return locks.GoalLock{Type: locks.LockNone}, nil
}

func (ExpandInfluence) UpdateProgress(actingFaction *domain.Faction, mutations []domain.Mutation, _ *state.FactionState, _ *rulebook.Rulebook, index *world.Index) []domain.Mutation {
	for _, mutation := range mutations {
		v, ok := mutation.(domain.BaseAdded)
		if !ok || v.CausedByFactionID != actingFaction.ID {
			continue
		}
		if factionHasBaseOn(actingFaction, v.Base.Location.WorldID) {
			continue
		}
		xp := 1
		if worldHasRivalPresence(v.Base.Location.WorldID, actingFaction.ID, index) {
			xp = 2
		}
		return completeGoal(actingFaction, xp)
	}
	return nil
}

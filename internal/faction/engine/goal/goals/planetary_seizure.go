package goals

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type PlanetarySeizure struct{}

func (PlanetarySeizure) GoalID() string { return "G-004" }

func (PlanetarySeizure) CheckLock(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) (locks.GoalLock, []domain.Mutation) {
	goal := faction.ActiveGoal
	if goal.ProcessPhase == 0 {
		return locks.GoalLock{Type: locks.LockNone}, nil
	}
	if goal.ProcessPhase == 1 {
		return locks.GoalLock{Type: locks.LockRestrictActions, AllowedActions: []string{"Attack"}}, nil
	}
	// Phase 2: occupation.
	if !factionHasUnstealthedAssetOn(faction, goal.TargetWorld) {
		return locks.GoalLock{Type: locks.LockNone}, []domain.Mutation{
			domain.GoalAbandoned{
				FactionID: faction.ID,
				GoalID:    goal.GoalID,
				Cause:     "occupation_failed",
			},
		}
	}
	tick := domain.GoalTurnsTick{
		FactionID: faction.ID,
		GoalID:    goal.GoalID,
		Cause:     "planetary_seizure_occupation",
	}
	newTurns := goal.TurnsRemaining - 1
	if newTurns == 0 {
		xp := calcPlanetarySeizureXP(goal.TargetFactionID, factionState)
		var pgTag domain.Tag
		if t, ok := rulebook.Tags["T-011"]; ok {
			pgTag = *t
		}
		return locks.GoalLock{Type: locks.LockNone}, []domain.Mutation{
			tick,
			domain.TagAdded{FactionID: faction.ID, Tag: pgTag, Cause: "goal_completed"},
			domain.GoalCompleted{FactionID: faction.ID, GoalID: goal.GoalID, XPAwarded: xp},
			domain.XPAwarded{FactionID: faction.ID, Amount: xp, Cause: "goal_completed"},
		}
	}
	return locks.GoalLock{Type: locks.LockNone}, []domain.Mutation{tick}
}

func (PlanetarySeizure) UpdateProgress(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook, _ *world.Index) []domain.Mutation {
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

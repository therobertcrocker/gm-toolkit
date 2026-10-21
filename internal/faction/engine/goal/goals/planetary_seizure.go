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
		// deliberate: re-check presence here, not only in UpdateProgress, because a rival can leave, or be taken by a reactor such as capture_target, where UpdateProgress cannot see it, and the Attack-only lock would then hold forever.
		if !rivalPresentOn(goal.TargetWorld.WorldID, faction.ID, factionState, nil) {
			return locks.GoalLock{Type: locks.LockNone}, []domain.Mutation{phaseOneCleared(faction.ID, goal.GoalID)}
		}
		return locks.GoalLock{Type: locks.LockRestrictActions, AllowedActions: []string{"Attack"}}, nil
	}
	// Phase 2: occupation.
	if !factionHasUnstealthedAssetOn(faction, goal.TargetWorld.WorldID) {
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
	cleared := make(map[string]bool)
	for _, mutation := range mutations {
		switch m := mutation.(type) {
		case domain.AssetRemoved:
			cleared[m.AssetID] = true
		case domain.AssetOwnerChanged:
			if m.ToFactionID == actingFaction.ID {
				cleared[m.AssetID] = true
			}
		}
	}
	if rivalPresentOn(goal.TargetWorld.WorldID, actingFaction.ID, factionState, cleared) {
		return nil
	}
	return []domain.Mutation{phaseOneCleared(actingFaction.ID, goal.GoalID)}
}

// rivalPresentOn reports whether a faction other than seizerID holds an
// unstealthed asset on world that cleared does not name.
func rivalPresentOn(world, seizerID string, factionState *state.FactionState, cleared map[string]bool) bool {
	for factionID, faction := range factionState.Factions {
		if factionID == seizerID {
			continue
		}
		for _, asset := range faction.Assets {
			if asset.Location.WorldID == world && !asset.Stealthy && !cleared[asset.ID] {
				return true
			}
		}
	}
	return false
}

func phaseOneCleared(factionID, goalID string) domain.GoalPhaseAdvanced {
	return domain.GoalPhaseAdvanced{
		FactionID:      factionID,
		GoalID:         goalID,
		ProcessPhase:   2,
		TurnsRemaining: 3,
		Cause:          "planetary_seizure_phase_advance",
	}
}

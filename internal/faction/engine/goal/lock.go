package goal

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// LockType classifies the constraint that a faction's active goal imposes on its turn.
type LockType int

const (
	LockNone            LockType = iota // normal flow
	LockSkip                            // skip faction entirely (Change Homeworld in-progress)
	LockRestrictActions                 // limit available actions (Seize Planet combat phase)
)

// GoalLock describes how the active goal constrains this faction's turn.
type GoalLock struct {
	Type           LockType
	AllowedActions []string // non-nil only when Type == LockRestrictActions
}

func factionHasUnstealthedAssetOn(faction *domain.Faction, world string) bool {
	for _, asset := range faction.Assets {
		if asset.Location == world && !asset.Stealthy {
			return true
		}
	}
	return false
}

func calcPlanetarySeizureXP(targetFactionID string, factionState *state.FactionState) int {
	if targetFactionID == "" {
		return 1
	}
	targetFaction, ok := factionState.Factions[targetFactionID]
	if !ok {
		return 1
	}
	avg := (targetFaction.Force + targetFaction.Cunning + targetFaction.Wealth) / 3
	xp := avg / 2
	if xp < 1 {
		return 1
	}
	return xp
}

func checkLockChangeHomeworld(faction *domain.Faction) (GoalLock, []domain.Mutation) {
	newTurns := faction.ActiveGoal.TurnsRemaining - 1
	tick := domain.GoalTurnsTick{
		FactionID: faction.ID,
		GoalID:    faction.ActiveGoal.GoalID,
		Cause:     "change_homeworld_transit",
	}
	if newTurns == 0 {
		return GoalLock{Type: LockSkip}, []domain.Mutation{
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
	return GoalLock{Type: LockSkip}, []domain.Mutation{tick}
}

func checkLockPlanetarySeizure(faction *domain.Faction, factionState *state.FactionState, rulebook *loader.Rulebook) (GoalLock, []domain.Mutation) {
	goal := faction.ActiveGoal
	if goal.ProcessPhase == 0 {
		return GoalLock{Type: LockNone}, nil
	}
	if goal.ProcessPhase == 1 {
		return GoalLock{Type: LockRestrictActions, AllowedActions: []string{"Attack"}}, nil
	}
	// Phase 2: occupation.
	if !factionHasUnstealthedAssetOn(faction, goal.TargetWorld) {
		return GoalLock{Type: LockNone}, []domain.Mutation{
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
		return GoalLock{Type: LockNone}, []domain.Mutation{
			tick,
			domain.TagAdded{FactionID: faction.ID, Tag: pgTag, Cause: "goal_completed"},
			domain.GoalCompleted{FactionID: faction.ID, GoalID: goal.GoalID, XPAwarded: xp},
			domain.XPAwarded{FactionID: faction.ID, Amount: xp, Cause: "goal_completed"},
		}
	}
	return GoalLock{Type: LockNone}, []domain.Mutation{tick}
}

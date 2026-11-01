package goals

import (
	"slices"

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
		// deliberate: re-check presence here, not only in UpdateProgress, because a rival can leave on its own turn, where UpdateProgress cannot see it, and the Attack-only lock would then hold forever.
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
		worldID := goal.TargetWorld.WorldID
		previousRuler := factionState.Ruler(worldID)
		xp := calcPlanetarySeizureXP(previousRuler, factionState)
		mutations := []domain.Mutation{tick}
		mutations = append(mutations, takeGovernment(faction, worldID, previousRuler, factionState, rulebook)...)
		mutations = append(mutations,
			domain.GoalCompleted{FactionID: faction.ID, GoalID: goal.GoalID, XPAwarded: xp},
			domain.XPAwarded{FactionID: faction.ID, Amount: xp, Cause: "goal_completed"},
		)
		return locks.GoalLock{Type: locks.LockNone}, mutations
	}
	return locks.GoalLock{Type: locks.LockNone}, []domain.Mutation{tick}
}

// takeGovernment makes seizer the ruler of worldID (overview Decision 10): the
// record changes hands, the seizer gains the government tag unless it holds it,
// and a previous ruler that governs no other world loses it. A seizer that already
// rules the world changes nothing. A rulebook with no government tag changes the
// record only.
func takeGovernment(seizer *domain.Faction, worldID, previousRuler string, factionState *state.FactionState, rulebook *rulebook.Rulebook) []domain.Mutation {
	if previousRuler == seizer.ID {
		return nil
	}
	mutations := []domain.Mutation{domain.WorldGovernmentChanged{
		WorldID:           worldID,
		FromFactionID:     previousRuler,
		ToFactionID:       seizer.ID,
		Cause:             "goal_completed",
		CausedByFactionID: seizer.ID,
	}}
	governmentTagID := rulebook.GovernmentTagID
	if governmentTagID == "" {
		return mutations
	}
	if !seizer.HasTag(governmentTagID) {
		mutations = append(mutations, domain.TagAdded{FactionID: seizer.ID, TagID: governmentTagID, Cause: "goal_completed", CausedByFactionID: seizer.ID})
	}
	previous := factionState.Factions[previousRuler]
	if previous == nil || !previous.HasTag(governmentTagID) || !slices.Equal(factionState.WorldsRuledBy(previousRuler), []string{worldID}) {
		return mutations
	}
	return append(mutations, domain.TagRemoved{FactionID: previousRuler, TagID: governmentTagID, Cause: "goal_completed", CausedByFactionID: seizer.ID})
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
		case domain.AssetConditionApplied:
			if m.Condition.Kind == domain.ConditionSheltered {
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
// unstealthed, unsheltered asset on world that cleared does not name. A sheltered
// asset hides and cannot be attacked, so the Attack-only lock would hold forever
// (Effort 2 Decision 11).
func rivalPresentOn(world, seizerID string, factionState *state.FactionState, cleared map[string]bool) bool {
	for factionID, faction := range factionState.Factions {
		if factionID == seizerID {
			continue
		}
		for _, asset := range faction.Assets {
			if asset.Location.WorldID == world && !asset.Hidden() && !cleared[asset.ID] {
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

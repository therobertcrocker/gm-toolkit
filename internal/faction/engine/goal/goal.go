package goal

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// GoalEngine evaluates and advances faction goal state.
type GoalEngine struct{}

func New() *GoalEngine { return &GoalEngine{} }

// CheckLock evaluates the faction's active goal and returns the appropriate lock
// state. Must be called before bookkeeping and action selection each turn.
func (ge *GoalEngine) CheckLock(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) (GoalLock, []domain.Mutation) {
	if faction.ActiveGoal == nil {
		return GoalLock{Type: LockNone}, nil
	}
	switch faction.ActiveGoal.GoalID {
	case "G-012":
		return checkLockChangeHomeworld(faction)
	case "G-004":
		return checkLockPlanetarySeizure(faction, factionState, rulebook)
	}
	return GoalLock{Type: LockNone}, nil
}

// UpdateProgress inspects the acting faction's mutation list for goal-relevant
// events and returns supplemental mutations if the goal advances or completes.
// Must be called before MutationEngine.Apply so destructive mutations have not
// yet fired.
func (ge *GoalEngine) UpdateProgress(
	actingFactionID string,
	mutations []domain.Mutation,
	factionState *state.FactionState,
	rulebook *rulebook.Rulebook,
) []domain.Mutation {
	actingFaction, ok := factionState.Factions[actingFactionID]
	if !ok || actingFaction.ActiveGoal == nil {
		return nil
	}
	switch actingFaction.ActiveGoal.GoalID {
	case "G-001":
		return progressMilitaryConquest(actingFaction, mutations, factionState, rulebook)
	case "G-002":
		return progressCommercialExpansion(actingFaction, mutations, factionState, rulebook)
	case "G-003":
		return progressIntelligenceCoup(actingFaction, mutations, factionState, rulebook)
	case "G-004":
		return progressPlanetarySeizure(actingFaction, mutations, factionState)
	case "G-005":
		return progressExpandInfluence(actingFaction, mutations, factionState)
	case "G-006":
		return progressBloodTheEnemy(actingFaction, mutations)
	case "G-007":
		return progressPeaceableKingdom(actingFaction, mutations)
	case "G-008":
		return progressDestroyTheFoe(actingFaction, mutations, factionState)
	case "G-009":
		return progressInsideEnemyTerritory(actingFaction, mutations, factionState)
	case "G-010":
		return progressInvincibleValor(actingFaction, mutations, factionState, rulebook)
	case "G-011":
		return progressWealthOfWorlds(actingFaction, mutations)
	}
	return nil
}

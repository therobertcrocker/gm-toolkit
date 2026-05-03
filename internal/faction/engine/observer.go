package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// TurnObserver is the engine's output channel. Methods are fire-and-forget —
// the engine does not wait for the observer and observers cannot affect game
// state. Callers wire one or more observers into the orchestrator to react to
// pipeline progress (display, logs, narration).
type TurnObserver interface {
	OnFactionTurnStarted(faction *domain.Faction)
	OnFactionSkipped(faction *domain.Faction)
	OnGoalLockApplied(faction *domain.Faction, lock GoalLock, mutations []domain.Mutation)
	OnBookkeepingApplied(faction *domain.Faction, result BookkeepingResult, mutations []domain.Mutation)
	OnActionSelected(faction *domain.Faction, action Action)
	OnActionResolved(faction *domain.Faction, action Action, mutations []domain.Mutation)
	OnFactionTurnCompleted(faction *domain.Faction)
	OnCycleCompleted(cycleNumber int, factionState *state.FactionState)
	OnError(faction *domain.Faction, err error)
}

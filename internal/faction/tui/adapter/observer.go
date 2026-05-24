package adapter

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/turn"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type observer struct{ adapter *Adapter }

func (a *Adapter) Observer() engine.TurnObserver { return &observer{adapter: a} }

func (o *observer) emit(kind EventKind, payload any) {
	select {
	case o.adapter.eventCh <- ObserverEventMsg{Kind: kind, Payload: payload}:
	default:
		o.adapter.log.Warn("observer event dropped — eventCh full", "kind", kind)
	}
}

func (o *observer) OnFactionTurnStarted(faction *domain.Faction)   { o.emit(EvtFactionTurnStarted, faction) }
func (o *observer) OnFactionSkipped(faction *domain.Faction)       { o.emit(EvtFactionSkipped, faction) }
func (o *observer) OnFactionTurnCompleted(faction *domain.Faction) { o.emit(EvtFactionTurnCompleted, faction) }
func (o *observer) OnStatRaiseSkipped(faction *domain.Faction)     { o.emit(EvtStatRaiseSkipped, faction) }

func (o *observer) OnGoalLockApplied(faction *domain.Faction, lock locks.GoalLock, muts []domain.Mutation) {
	o.emit(EvtGoalLockApplied, GoalLockAppliedPayload{Lock: lock, Mutations: muts})
}
func (o *observer) OnBookkeepingApplied(faction *domain.Faction, result turn.BookkeepingResult, muts []domain.Mutation) {
	o.emit(EvtBookkeepingApplied, BookkeepingAppliedPayload{Result: result, Mutations: muts})
}
func (o *observer) OnActionSelected(faction *domain.Faction, selected action.Action) {
	o.emit(EvtActionSelected, ActionSelectedPayload{Selected: selected})
}
func (o *observer) OnActionResolved(faction *domain.Faction, selected action.Action, muts []domain.Mutation) {
	o.emit(EvtActionResolved, ActionResolvedPayload{Selected: selected, Mutations: muts})
}
func (o *observer) OnCycleCompleted(cycleNumber int, factionState *state.FactionState) {
	o.emit(EvtCycleCompleted, CycleCompletedPayload{CycleNumber: cycleNumber, FactionState: factionState})
}
func (o *observer) OnError(faction *domain.Faction, err error) {
	o.emit(EvtError, ErrorPayload{Err: err})
}
func (o *observer) OnStatRaiseApplied(faction *domain.Faction, raised *domain.FactionStat, muts []domain.Mutation) {
	o.emit(EvtStatRaiseApplied, StatRaiseAppliedPayload{Raised: raised, Mutations: muts})
}
func (o *observer) OnMovementTicked(faction *domain.Faction, muts []domain.Mutation) {
	o.emit(EvtMovementTicked, MovementTickedPayload{Mutations: muts})
}
func (o *observer) OnMovementResolved(faction *domain.Faction, muts []domain.Mutation) {
	o.emit(EvtMovementResolved, MovementResolvedPayload{Mutations: muts})
}
func (o *observer) OnIndexSkipped(skipped []string) {
	o.emit(EvtIndexSkipped, IndexSkippedPayload{Skipped: skipped})
}

var _ engine.TurnObserver = (*observer)(nil)

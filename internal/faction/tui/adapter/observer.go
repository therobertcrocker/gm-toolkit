package adapter

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/turn"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type observer struct{ adapter *Adapter }

func (a *Adapter) Observer() engine.TurnObserver { return &observer{adapter: a} }

func (o *observer) OnFactionTurnStarted(faction *domain.Faction) {
	o.adapter.emit(EvtFactionTurnStarted, faction)
}
func (o *observer) OnFactionSkipped(faction *domain.Faction) {
	o.adapter.emit(EvtFactionSkipped, faction)
}
func (o *observer) OnFactionTurnCompleted(faction *domain.Faction) {
	o.adapter.emit(EvtFactionTurnCompleted, faction)
}
func (o *observer) OnStatRaiseSkipped(faction *domain.Faction) {
	o.adapter.emit(EvtStatRaiseSkipped, faction)
}

func (o *observer) OnBookkeepingApplied(faction *domain.Faction, result turn.BookkeepingResult, muts []domain.Mutation) {
	o.adapter.emit(EvtBookkeepingApplied, BookkeepingAppliedPayload{Result: result, Mutations: muts})
}
func (o *observer) OnActionSelected(faction *domain.Faction, selected action.Action) {
	o.adapter.emit(EvtActionSelected, ActionSelectedPayload{Selected: selected})
}
func (o *observer) OnActionResolved(faction *domain.Faction, selected action.Action, muts []domain.Mutation) {
	o.adapter.emit(EvtActionResolved, ActionResolvedPayload{Selected: selected, Mutations: muts})
}
func (o *observer) OnCycleCompleted(cycleNumber int, factionState *state.FactionState) {
	o.adapter.emit(EvtCycleCompleted, CycleCompletedPayload{CycleNumber: cycleNumber, FactionState: factionState})
}
func (o *observer) OnError(faction *domain.Faction, err error) {
	o.adapter.emit(EvtError, ErrorPayload{Err: err})
}
func (o *observer) OnStatRaiseApplied(faction *domain.Faction, raised *domain.FactionStat, muts []domain.Mutation) {
	o.adapter.emit(EvtStatRaiseApplied, StatRaiseAppliedPayload{Raised: raised, Mutations: muts})
}
func (o *observer) OnMovementTicked(faction *domain.Faction, muts []domain.Mutation) {
	o.adapter.emit(EvtMovementTicked, MovementTickedPayload{Mutations: muts})
}
func (o *observer) OnMovementResolved(faction *domain.Faction, muts []domain.Mutation) {
	o.adapter.emit(EvtMovementResolved, MovementResolvedPayload{Mutations: muts})
}
func (o *observer) OnIndexSkipped(skipped []string) {
	o.adapter.emit(EvtIndexSkipped, IndexSkippedPayload{Skipped: skipped})
}

var _ engine.TurnObserver = (*observer)(nil)

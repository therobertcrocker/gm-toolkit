// Package testharness provides headless implementations of InputCollector
// and TurnObserver for orchestrator tests, AI batch runs, and any future
// non-interactive engine consumer.
package testharness

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/turn"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// ObservedEvent captures a single TurnObserver callback for later assertion.
// Payload carries the call's auxiliary arguments (mutations, action, lock,
// etc.); Faction is the affected faction (nil for cycle-level events).
type ObservedEvent struct {
	Kind    string
	Faction *domain.Faction
	Payload any
}

// RecordingObserver appends every observer callback to Events in call order.
// Use Kinds() / FactionIDs() in assertions for compact comparisons.
type RecordingObserver struct {
	Events []ObservedEvent
}

func (r *RecordingObserver) OnFactionTurnStarted(faction *domain.Faction) {
	r.Events = append(r.Events, ObservedEvent{Kind: "TurnStarted", Faction: faction})
}

func (r *RecordingObserver) OnFactionSkipped(faction *domain.Faction) {
	r.Events = append(r.Events, ObservedEvent{Kind: "FactionSkipped", Faction: faction})
}

func (r *RecordingObserver) OnGoalLockApplied(faction *domain.Faction, lock goal.GoalLock, mutations []domain.Mutation) {
	r.Events = append(r.Events, ObservedEvent{
		Kind:    "GoalLockApplied",
		Faction: faction,
		Payload: GoalLockPayload{Lock: lock, Mutations: mutations},
	})
}

func (r *RecordingObserver) OnBookkeepingApplied(faction *domain.Faction, result turn.BookkeepingResult, mutations []domain.Mutation) {
	r.Events = append(r.Events, ObservedEvent{
		Kind:    "BookkeepingApplied",
		Faction: faction,
		Payload: BookkeepingPayload{Result: result, Mutations: mutations},
	})
}

func (r *RecordingObserver) OnActionSelected(faction *domain.Faction, selectedAction action.Action) {
	r.Events = append(r.Events, ObservedEvent{Kind: "ActionSelected", Faction: faction, Payload: selectedAction})
}

func (r *RecordingObserver) OnActionResolved(faction *domain.Faction, selectedAction action.Action, mutations []domain.Mutation) {
	r.Events = append(r.Events, ObservedEvent{
		Kind:    "ActionResolved",
		Faction: faction,
		Payload: ActionResolvedPayload{Action: selectedAction, Mutations: mutations},
	})
}

func (r *RecordingObserver) OnFactionTurnCompleted(faction *domain.Faction) {
	r.Events = append(r.Events, ObservedEvent{Kind: "TurnCompleted", Faction: faction})
}

func (r *RecordingObserver) OnCycleCompleted(cycleNumber int, factionState *state.FactionState) {
	r.Events = append(r.Events, ObservedEvent{
		Kind:    "CycleCompleted",
		Payload: CycleCompletedPayload{CycleNumber: cycleNumber, FactionState: factionState},
	})
}

func (r *RecordingObserver) OnError(faction *domain.Faction, err error) {
	r.Events = append(r.Events, ObservedEvent{Kind: "Error", Faction: faction, Payload: err})
}

// Kinds returns the recorded event kinds in order — convenient for sequence
// assertions that don't care about payload contents.
func (r *RecordingObserver) Kinds() []string {
	kinds := make([]string, len(r.Events))
	for i, event := range r.Events {
		kinds[i] = event.Kind
	}
	return kinds
}

// GoalLockPayload is the Payload type for GoalLockApplied events.
type GoalLockPayload struct {
	Lock      goal.GoalLock
	Mutations []domain.Mutation
}

// BookkeepingPayload is the Payload type for BookkeepingApplied events.
type BookkeepingPayload struct {
	Result    turn.BookkeepingResult
	Mutations []domain.Mutation
}

// ActionResolvedPayload is the Payload type for ActionResolved events.
type ActionResolvedPayload struct {
	Action    action.Action
	Mutations []domain.Mutation
}

// CycleCompletedPayload is the Payload type for CycleCompleted events.
type CycleCompletedPayload struct {
	CycleNumber  int
	FactionState *state.FactionState
}

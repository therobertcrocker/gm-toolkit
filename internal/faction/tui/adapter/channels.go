package adapter

import (
	"errors"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/turn"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type AskKind int

const (
	AskAwaitCheckpoint AskKind = iota
	AskSelectAction
	AskSelectStatRaise
	AskSelectMovementDecisions
	AskSelectTransportCargo
)

type CollectorAskMsg struct {
	Kind    AskKind
	Faction *domain.Faction
	Payload any
	Reply   chan<- any
}

type EventKind int

const (
	EvtFactionTurnStarted EventKind = iota
	EvtFactionSkipped
	EvtGoalLockApplied
	EvtBookkeepingApplied
	EvtActionSelected
	EvtActionResolved
	EvtFactionTurnCompleted
	EvtCycleCompleted
	EvtError
	EvtStatRaiseApplied
	EvtStatRaiseSkipped
	EvtMovementTicked
	EvtMovementResolved
	EvtIndexSkipped
)

type ObserverEventMsg struct {
	Kind    EventKind
	Payload any
}

type EngineDoneMsg struct {
	Err error
}

// Per-kind payload types for CollectorAskMsg.

type AwaitCheckpointPayload struct{ Phase string }
type SelectActionPayload struct{ Available []action.Action }
type SelectStatRaisePayload struct{ Eligible []domain.FactionStat }
type SelectMovementDecisionsPayload struct{ Eligible []*domain.Asset }
type SelectTransportCargoPayload struct {
	Transport     *domain.Asset
	EligibleCargo []*domain.Asset
	Profile       *domain.TransportProfile
}

// Per-kind payload types for ObserverEventMsg.

type GoalLockAppliedPayload struct {
	Lock      locks.GoalLock
	Mutations []domain.Mutation
}
type BookkeepingAppliedPayload struct {
	Result    turn.BookkeepingResult
	Mutations []domain.Mutation
}
type ActionSelectedPayload struct{ Selected action.Action }
type ActionResolvedPayload struct {
	Selected  action.Action
	Mutations []domain.Mutation
}
type CycleCompletedPayload struct {
	CycleNumber  int
	FactionState *state.FactionState
}
type ErrorPayload struct{ Err error }
type StatRaiseAppliedPayload struct {
	Raised    *domain.FactionStat
	Mutations []domain.Mutation
}
type MovementTickedPayload struct{ Mutations []domain.Mutation }
type MovementResolvedPayload struct{ Mutations []domain.Mutation }
type IndexSkippedPayload struct{ Skipped []string }

// FactionTurnStarted, FactionSkipped, FactionTurnCompleted, and StatRaiseSkipped
// carry only a *domain.Faction — stored directly in ObserverEventMsg.Payload.

// ErrTurnCanceled is the cancel sentinel returned when the user dismisses an
// ask prompt mid-turn. Routing logic for this lives in the action-selection initiative.
var ErrTurnCanceled = errors.New("turn canceled by user")

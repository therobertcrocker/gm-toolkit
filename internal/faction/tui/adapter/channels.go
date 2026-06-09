package adapter

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
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
	// Action-surface kinds (Effort 3). Appended only — never reorder.
	AskSelectAsset
	AskSelectModifiers
	AskConfirmReroll
	AskSelectBuyOrder
	AskSelectRefitOrder
	AskSelectRepairOrders
	AskSelectBribeTarget
	AskSelectSeizeTarget
	AskSelectChangeHomeworldTarget
	AskSelectAttackers
	AskSelectDefender
	AskConfirmRedirectToBase
	AskSelectExpandInfluenceOrder
	AskConfirmRivalFreeAttack
	AskSelectBaseAttackers
	AskSelectAbilityAssets
	AskConfirmAbilityApplied
	AskSelectFactionTestTarget
)

type CollectorAskMsg struct {
	Kind    AskKind
	Faction *domain.Faction
	Payload any
	Reply   chan<- any
}

// StreamClosedMsg is emitted by ObserverPump once eventCh is drained and closed
// — i.e. RunCycle has returned and every buffered event has been delivered. The
// router uses this (not EngineDoneMsg) as the teardown signal, so the final
// cycle-completed events always render before the view resets.
type StreamClosedMsg struct{}

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
	EvtCycleStarted
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

// Action-surface payloads (Effort 3).
type SelectAssetPayload struct{ Assets []*domain.Asset }
type SelectModifiersPayload struct{ Offers []hooks.ModifierOffer }
type ConfirmRerollPayload struct{ Directive hooks.RerollDirective }
type SelectBuyOrderPayload struct {
	PurchasableByWorld map[string][]*domain.AssetDefinition
	WorldNames         map[string]string // worldID -> display name (adapter-built)
	Coin               int               // for cost-vs-budget labels
}
type SelectRefitOrderPayload struct{ Options []action.RefitOption }

// RepairTarget is a per-asset, adapter-pre-derived repair cap. HealCount in the
// returned action.RepairOrder is a number of escalating-cost heal steps, each
// restoring up to HealHP (the faction's attribute score for the asset category);
// MaxSteps = ceil(Missing/HealHP) is the useful ceiling (repair_asset.go:72-78).
type RepairTarget struct {
	Asset    *domain.Asset
	HealHP   int
	MaxSteps int
	Missing  int
}
type SelectRepairOrdersPayload struct{ Targets []RepairTarget }

type SelectBribeTargetPayload struct {
	Bases      []*domain.Base
	OwnerNames map[string]string // ownerID -> faction name
	Coin       int
}

// BribeReply pairs the two return values of SelectBribeTarget into a single
// OverlayDoneMsg.Answer the adapter type-asserts (Decision 9: the overlay emits
// this so the adapter need not import overlay).
type BribeReply struct {
	Base   *domain.Base
	Amount int
}
type SelectSeizeTargetPayload struct {
	Worlds     []string
	WorldNames map[string]string // worldID -> display name (adapter-built)
}
type SelectChangeHomeworldTargetPayload struct {
	Worlds     []string
	WorldNames map[string]string
}
type SelectAttackersPayload struct{ Eligible []*domain.Asset }
type SelectDefenderPayload struct {
	Attacker   *domain.Asset
	Eligible   []*domain.Asset
	OwnerNames map[string]string // ownerID -> faction name (defenders may be rivals')
}
type ConfirmRedirectToBasePayload struct {
	DefenderFaction *domain.Faction
	Base            *domain.Base
	Damage          int
}

// ReinforceTarget is a per-base, adapter-pre-derived reinforce cap. CanHeal/CanMax
// are the two independent engine filters (damaged vs growable); HealCap and MaxCap
// are the Coin/HP headroom for each submode (expand_influence.go:121-150).
type ReinforceTarget struct {
	Base    *domain.Base
	CanHeal bool
	HealCap int // EffectiveMaxHP - CurrentHP
	CanMax  bool
	MaxCap  int // faction.MaxHP - base.MaxHP
}
type SelectExpandInfluenceOrderPayload struct {
	NewBaseWorlds  []string          // from the engine (eligibleNewBaseWorlds)
	WorldNames     map[string]string // worldID -> display name
	ReinforceBases []ReinforceTarget // adapter-derived
	Coin           int
}
type ConfirmRivalFreeAttackPayload struct {
	Rival       *domain.Faction
	RivalRoll   int
	FactionRoll int
}
type SelectBaseAttackersPayload struct {
	Rival    *domain.Faction
	Eligible []*domain.Asset
}
type SelectAbilityAssetsPayload struct{ Candidates []*domain.Asset }
type ConfirmAbilityAppliedPayload struct {
	Asset *domain.Asset
	Def   *domain.AssetDefinition
}
type SelectFactionTestTargetPayload struct {
	Asset      *domain.Asset
	Effect     domain.AbilityEffectType
	Candidates []*domain.Faction
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

type RailEntry struct{ ID, Name string }
type CycleStartedPayload struct {
	CycleNumber int
	Order       []RailEntry
}

// FactionTurnStarted, FactionSkipped, FactionTurnCompleted, and StatRaiseSkipped
// carry only a *domain.Faction — stored directly in ObserverEventMsg.Payload.

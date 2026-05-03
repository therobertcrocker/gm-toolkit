package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	engineaction "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/ability"
)

// InputCollector abstracts input collection for action and ability resolution.
// The GM implementation uses interactive prompts; AI and test implementations
// use scripted or goal-driven logic.
type InputCollector interface {
	engineaction.Collector
	ability.Collector
	SelectAction(faction *domain.Faction, available []Action) (Action, error)
	AwaitCheckpoint(phase string) error
}

// Type aliases so existing code using engine.RepairOrder etc. continues to
// compile during the refactor. Canonical home is engine/action.
type (
	RepairOrder          = engineaction.RepairOrder
	BuyOrder             = engineaction.BuyOrder
	RefitOption          = engineaction.RefitOption
	RefitOrder           = engineaction.RefitOrder
	ExpandMode           = engineaction.ExpandMode
	ReinforceMode        = engineaction.ReinforceMode
	ExpandInfluenceOrder = engineaction.ExpandInfluenceOrder
)

// ExpandMode constants — aliased from engine/action.
const (
	ExpandModeNew       = engineaction.ExpandModeNew
	ExpandModeReinforce = engineaction.ExpandModeReinforce
)

// ReinforceMode constants — aliased from engine/action.
const (
	ReinforceHeal = engineaction.ReinforceHeal
	ReinforceMax  = engineaction.ReinforceMax
)

// Phase checkpoint constants used by the orchestrator pipeline.
const (
	PhaseBookkeeping  = "bookkeeping"
	PhaseActionResult = "action_result"
	PhaseGoalLocked   = "goal_locked"
	PhaseCycleSummary = "cycle_summary"
)

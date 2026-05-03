package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/ability"
	engineaction "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
)

// InputCollector abstracts input collection for action and ability resolution.
// The GM implementation uses interactive prompts; AI and test implementations
// use scripted or goal-driven logic.
type InputCollector interface {
	engineaction.Collector
	ability.Collector
	SelectAction(faction *domain.Faction, available []engineaction.Action) (engineaction.Action, error)
	AwaitCheckpoint(phase string) error
}

// Checkpoint constants name the pipeline phases where the orchestrator pauses.
const (
	CheckpointBookkeeping  = "bookkeeping"
	CheckpointActionResult = "action_result"
	CheckpointGoalLocked   = "goal_locked"
	CheckpointCycleSummary = "cycle_summary"
)

package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/ability"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
)

// InputCollector abstracts input collection for action and ability resolution.
// The GM implementation uses interactive prompts; AI and test implementations
// use scripted or goal-driven logic.
type InputCollector interface {
	action.Collector
	ability.Collector
	SelectAction(faction *domain.Faction, available []action.Action) (action.Action, error)
	AwaitCheckpoint(phase string) error
}

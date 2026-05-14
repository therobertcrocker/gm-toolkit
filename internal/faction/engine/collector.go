package engine

//go:generate mockgen -destination=action/actions/mocks/mock_collector.go -package=mocks github.com/therobertcrocker/gm-toolkit/internal/faction/engine InputCollector

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// InputCollector abstracts input collection for action, ability, and hook resolution.
// The GM implementation uses interactive prompts; AI and test implementations
// use scripted or goal-driven logic.
type InputCollector interface {
	action.Collector
	hooks.Collector
	SelectAction(faction *domain.Faction, available []action.Action) (action.Action, error)
	AwaitCheckpoint(phase string) error
	SelectStatRaise(faction *domain.Faction, eligible []domain.FactionStat) (*domain.FactionStat, error)
}

package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
)

// PhaseCollector abstracts the four orchestrator-owned input prompts.
// The GM implementation uses interactive prompts; AI and test implementations
// use scripted or goal-driven logic.
type PhaseCollector interface {
	AwaitCheckpoint(phase string) error
	SelectAction(faction *domain.Faction, available []action.Action) (action.Action, error)
	SelectStatRaise(faction *domain.Faction, eligible []domain.FactionStat) (*domain.FactionStat, error)
	SelectMovementDecisions(faction *domain.Faction, eligible []*domain.Asset) ([]world.MovementDecision, error)
	SelectTransportCargo(transport *domain.Asset, eligibleCargo []*domain.Asset, profile *domain.TransportProfile) ([]*domain.Asset, error)
}

// Collectors bundles the two input contracts the orchestrator threads through
// the pipeline. Phase handles orchestrator-level prompts; Action is passed to
// sub-engines and satisfies hooks.Collector + ability.Collector transitively.
type Collectors struct {
	Phase  PhaseCollector
	Action action.Collector
}

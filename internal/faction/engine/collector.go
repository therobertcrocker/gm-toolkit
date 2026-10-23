package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
)

// PhaseCollector abstracts the orchestrator-owned input prompts: the phase
// checkpoints, action choice, stat raise, movement decisions, transport cargo and
// the relocate orders of a free movement grant, the pay-to-lift question at turn
// start, and an arriving asset's permission request.
// The GM implementation uses interactive prompts; AI and test implementations
// use scripted or goal-driven logic.
type PhaseCollector interface {
	hooks.PaymentCollector
	hooks.PermissionCollector
	AwaitCheckpoint(phase string) error
	SelectAction(faction *domain.Faction, available []action.Action) (action.Action, error)
	SelectStatRaise(faction *domain.Faction, eligible []domain.FactionStat) (*domain.FactionStat, error)
	SelectMovementDecisions(faction *domain.Faction, eligible []*domain.Asset) ([]world.MovementDecision, error)
	SelectTransportCargo(transport *domain.Asset, eligibleCargo []*domain.Asset, profile *domain.TransportProfile) ([]*domain.Asset, error)
	SelectRelocateOrders(acting *domain.Asset, candidates []*domain.Asset, candidateWorlds []string, maxAssets int) ([]action.RelocateOrder, error)
}

// Collectors bundles the two input contracts the orchestrator threads through
// the pipeline. Phase handles orchestrator-level prompts; Action is passed to
// sub-engines and satisfies hooks.Collector + ability.Collector transitively.
type Collectors struct {
	Phase  PhaseCollector
	Action action.Collector
}

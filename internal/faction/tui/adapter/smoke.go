package adapter

import (
	"log/slog"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/turn"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

// SmokePhaseCollector is a minimal PhaseCollector for the dry-run cycle.
// SelectAction picks the first available action; everything else is a no-op.
type SmokePhaseCollector struct {
	log *slog.Logger
}

func NewSmokePhaseCollector(log *slog.Logger) *SmokePhaseCollector {
	return &SmokePhaseCollector{log: log}
}

func (s *SmokePhaseCollector) AwaitCheckpoint(phase string) error {
	s.log.Info("smoke: checkpoint", "phase", phase)
	return nil
}

func (s *SmokePhaseCollector) SelectAction(faction *domain.Faction, available []action.Action) (action.Action, error) {
	if len(available) == 0 {
		s.log.Info("smoke: no actions available, skipping")
		return nil, nil
	}
	s.log.Info("smoke: selecting action", "action", available[0].Name())
	return available[0], nil
}

func (s *SmokePhaseCollector) SelectStatRaise(_ *domain.Faction, _ []domain.FactionStat) (*domain.FactionStat, error) {
	return nil, nil
}

func (s *SmokePhaseCollector) SelectMovementDecisions(_ *domain.Faction, _ []*domain.Asset) ([]world.MovementDecision, error) {
	return nil, nil
}

func (s *SmokePhaseCollector) SelectTransportCargo(_ *domain.Asset, _ []*domain.Asset, _ *domain.TransportProfile) ([]*domain.Asset, error) {
	return nil, nil
}

var _ engine.PhaseCollector = (*SmokePhaseCollector)(nil)

// SmokeObserver logs every engine event to slog at Info level.
type SmokeObserver struct {
	log *slog.Logger
}

func NewSmokeObserver(log *slog.Logger) *SmokeObserver {
	return &SmokeObserver{log: log}
}

func (o *SmokeObserver) OnFactionTurnStarted(faction *domain.Faction) {
	o.log.Info("EVT FactionTurnStarted", "faction", faction.ID)
}
func (o *SmokeObserver) OnFactionSkipped(faction *domain.Faction) {
	o.log.Info("EVT FactionSkipped", "faction", faction.ID)
}
func (o *SmokeObserver) OnGoalLockApplied(faction *domain.Faction, lock locks.GoalLock, mutations []domain.Mutation) {
	o.log.Info("EVT GoalLockApplied", "faction", faction.ID, "lock", lock.Type, "mutations", len(mutations))
}
func (o *SmokeObserver) OnBookkeepingApplied(faction *domain.Faction, result turn.BookkeepingResult, mutations []domain.Mutation) {
	o.log.Info("EVT BookkeepingApplied", "faction", faction.ID, "income_wealth", result.WealthIncome, "mutations", len(mutations))
}
func (o *SmokeObserver) OnActionSelected(faction *domain.Faction, selected action.Action) {
	o.log.Info("EVT ActionSelected", "faction", faction.ID, "action", selected.Name())
}
func (o *SmokeObserver) OnActionResolved(faction *domain.Faction, selected action.Action, mutations []domain.Mutation) {
	o.log.Info("EVT ActionResolved", "faction", faction.ID, "action", selected.Name(), "mutations", len(mutations))
}
func (o *SmokeObserver) OnFactionTurnCompleted(faction *domain.Faction) {
	o.log.Info("EVT FactionTurnCompleted", "faction", faction.ID)
}
func (o *SmokeObserver) OnCycleCompleted(cycleNumber int, _ *state.FactionState) {
	o.log.Info("EVT CycleCompleted", "cycle", cycleNumber)
}
func (o *SmokeObserver) OnError(faction *domain.Faction, err error) {
	factionID := "<nil>"
	if faction != nil {
		factionID = faction.ID
	}
	o.log.Error("EVT Error", "faction", factionID, "err", err)
}
func (o *SmokeObserver) OnStatRaiseApplied(faction *domain.Faction, raised *domain.FactionStat, mutations []domain.Mutation) {
	o.log.Info("EVT StatRaiseApplied", "faction", faction.ID, "stat", *raised, "mutations", len(mutations))
}
func (o *SmokeObserver) OnStatRaiseSkipped(faction *domain.Faction) {
	o.log.Info("EVT StatRaiseSkipped", "faction", faction.ID)
}
func (o *SmokeObserver) OnMovementTicked(faction *domain.Faction, mutations []domain.Mutation) {
	o.log.Info("EVT MovementTicked", "faction", faction.ID, "mutations", len(mutations))
}
func (o *SmokeObserver) OnMovementResolved(faction *domain.Faction, mutations []domain.Mutation) {
	o.log.Info("EVT MovementResolved", "faction", faction.ID, "mutations", len(mutations))
}
func (o *SmokeObserver) OnIndexSkipped(skipped []string) {
	o.log.Warn("EVT IndexSkipped", "worlds", skipped)
}

var _ engine.TurnObserver = (*SmokeObserver)(nil)

// smokeHexRouter satisfies world.HexRouter with no-op implementations.
// Safe only with zero-asset, zero-base faction states — none of its methods
// are reachable when the index loop has nothing to iterate.
type smokeHexRouter struct{}

func (smokeHexRouter) Location(_ string) (spatial.Location, bool) { return nil, false }

func (smokeHexRouter) Distance(_, _ spatial.RegionHex, _ int) (int, error) {
	return 0, spatial.ErrNotImplemented
}

func (smokeHexRouter) Path(_, _ spatial.RegionHex, _ int) ([]spatial.RegionHex, int, error) {
	return nil, 0, spatial.ErrNotImplemented
}

var _ world.HexRouter = smokeHexRouter{}

// NewSmokeWorldEngine returns a WorldEngine backed by smokeHexRouter.
// Only valid for zero-asset, zero-base states where spatial lookups never fire.
func NewSmokeWorldEngine(log *slog.Logger) *world.WorldEngine {
	return world.NewWithMap(smokeHexRouter{}, log)
}

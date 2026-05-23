package engine

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/config"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/turn"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/logging"
)

// Checkpoint constants name the pipeline phases where the orchestrator pauses
// for caller acknowledgement via PhaseCollector.AwaitCheckpoint.
const (
	CheckpointBookkeeping  = "bookkeeping"
	CheckpointMovement     = "movement"
	CheckpointActionResult = "action_result"
	CheckpointGoalLocked   = "goal_locked"
	CheckpointCycleSummary = "cycle_summary"
)

// RunCycle calls RunFactionTurn until a cycle completes. The caller must have
// already called Turn.Start (or be resuming a turn that's still InProgress).
func (e *Engine) RunCycle(
	factionState *state.FactionState,
	cfg *config.Config,
	collectors Collectors,
	observer TurnObserver,
) error {
	logging.RunHeader(e.log, "faction-manager", "dev", factionState.CampaignID, len(factionState.Factions))

	e.Tag.ApplyAll(factionState, e.Hooks, e.log.With("engine", "tag"))
	e.Effect.ApplyAll(factionState, e.Rulebook, e.Hooks, e.log.With("engine", "effect"))

	for {
		done, err := e.RunFactionTurn(factionState, cfg, collectors, observer)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}
}

// RunFactionTurn drives one faction's turn from goal-lock check through state
// save. Errors are reported to the observer via OnError and returned to the
// caller; partial mutations already applied stay applied.
//
// The boolean return reports whether this faction's turn closed out the
// current cycle (true when the turn cursor advanced past the last faction).
func (e *Engine) RunFactionTurn(
	factionState *state.FactionState,
	cfg *config.Config,
	collectors Collectors,
	observer TurnObserver,
) (bool, error) {
	turnLog := e.log.With("turn", factionState.CycleNumber)
	faction, err := e.setupFactionTurn(factionState, observer, turnLog)
	if err != nil {
		return false, err
	}
	turnLog = turnLog.With("faction", faction.ID)
	logging.TurnStart(turnLog, factionState.CycleNumber, faction.ID)

	lock, err := e.runGoalLockPhase(faction, factionState, cfg, collectors, observer, turnLog)
	if err != nil {
		return false, err
	}

	if err := e.runStatRaisePhase(faction, factionState, cfg, collectors, observer, turnLog); err != nil {
		return false, err
	}

	if err := e.runBookkeepingPhase(faction, factionState, cfg, collectors, observer, turnLog); err != nil {
		return false, err
	}
	if err := state.Save(cfg.StatePath, factionState); err != nil {
		observer.OnError(faction, err)
		return false, fmt.Errorf("saving state after bookkeeping: %w", err)
	}

	if err := e.runMovementPhase(faction, factionState, cfg, collectors, observer, turnLog); err != nil {
		return false, err
	}

	if lock.Type == locks.LockSkip {
		return e.finishFactionTurn(factionState, faction, cfg, collectors, observer, turnLog)
	}

	if err := e.runActionPhase(faction, factionState, cfg, collectors, observer, lock, turnLog); err != nil {
		return false, err
	}
	if err := state.Save(cfg.StatePath, factionState); err != nil {
		observer.OnError(faction, err)
		return false, fmt.Errorf("saving state after action resolution: %w", err)
	}

	return e.finishFactionTurn(factionState, faction, cfg, collectors, observer, turnLog)
}

func (e *Engine) setupFactionTurn(factionState *state.FactionState, observer TurnObserver, turnLog *slog.Logger) (*domain.Faction, error) {
	turnLog.Info("turn setup")
	if e.World != nil {
		skipped, err := e.World.RebuildIndex(factionState, turnLog.With("engine", "world"))
		if err != nil {
			observer.OnError(nil, err)
			return nil, err
		}
		if len(skipped) > 0 {
			observer.OnIndexSkipped(skipped)
		}
	}
	faction, err := e.Turn.CurrentFaction(factionState, turnLog.With("engine", "turn"))
	if err != nil {
		observer.OnError(nil, err)
		return nil, err
	}

	observer.OnFactionTurnStarted(faction)
	return faction, nil
}

func (e *Engine) runGoalLockPhase(
	faction *domain.Faction,
	factionState *state.FactionState,
	cfg *config.Config,
	collectors Collectors,
	observer TurnObserver,
	turnLog *slog.Logger,
) (locks.GoalLock, error) {
	phaseLog := turnLog.With("phase", "goal_lock")
	logging.PhaseStart(phaseLog, "goal_lock")
	phaseLog.Info("phase begin")

	lock, lockMutations := e.Goal.CheckLock(faction, factionState, e.Rulebook, phaseLog.With("engine", "goal"))
	observer.OnGoalLockApplied(faction, lock, lockMutations)

	if len(lockMutations) > 0 {
		if err := e.applyAndRecord(factionState, faction, lockMutations, cfg, phaseLog); err != nil {
			observer.OnError(faction, err)
			return lock, err
		}
	}

	if lock.Type == locks.LockSkip {
		if err := collectors.Phase.AwaitCheckpoint(CheckpointGoalLocked); err != nil {
			observer.OnError(faction, err)
			return lock, err
		}
	}

	phaseLog.Info("phase end", "lock", lock.Type)
	return lock, nil
}

func (e *Engine) runBookkeepingPhase(
	faction *domain.Faction,
	factionState *state.FactionState,
	cfg *config.Config,
	collectors Collectors,
	observer TurnObserver,
	turnLog *slog.Logger,
) error {
	phaseLog := turnLog.With("phase", "bookkeeping")
	logging.PhaseStart(phaseLog, "bookkeeping")
	phaseLog.Info("phase begin")

	bookResult, bookMutations, err := e.Turn.ApplyBookkeeping(factionState, e.Hooks, phaseLog.With("engine", "turn"))
	if err != nil {
		observer.OnError(faction, err)
		return err
	}
	if len(bookMutations) > 0 {
		if err := e.applyAndRecord(factionState, faction, bookMutations, cfg, phaseLog); err != nil {
			observer.OnError(faction, err)
			return err
		}
	}
	observer.OnBookkeepingApplied(faction, bookResult, bookMutations)
	if err := collectors.Phase.AwaitCheckpoint(CheckpointBookkeeping); err != nil {
		observer.OnError(faction, err)
		return err
	}
	phaseLog.Info("phase end", "mutations", len(bookMutations))
	return nil
}

func (e *Engine) runStatRaisePhase(
	faction *domain.Faction,
	factionState *state.FactionState,
	cfg *config.Config,
	collectors Collectors,
	observer TurnObserver,
	turnLog *slog.Logger,
) error {
	phaseLog := turnLog.With("phase", "stat_raise")
	logging.PhaseStart(phaseLog, "stat_raise")
	phaseLog.Info("phase begin")

	statToRaise, raiseMutations, err := prepareStatRaise(faction, collectors.Phase)
	if err != nil {
		observer.OnError(faction, err)
		return err
	}
	if statToRaise == nil {
		observer.OnStatRaiseSkipped(faction)
		phaseLog.Info("phase end", "outcome", "skipped")
		return nil
	}
	if len(raiseMutations) > 0 {
		if err := e.applyAndRecord(factionState, faction, raiseMutations, cfg, phaseLog); err != nil {
			observer.OnError(faction, err)
			return err
		}
		observer.OnStatRaiseApplied(faction, statToRaise, raiseMutations)
	}
	phaseLog.Info("phase end", "outcome", "raised", "stat", *statToRaise)
	return nil
}

func (e *Engine) runMovementPhase(
	faction *domain.Faction,
	factionState *state.FactionState,
	cfg *config.Config,
	collectors Collectors,
	observer TurnObserver,
	turnLog *slog.Logger,
) error {
	phaseLog := turnLog.With("phase", "movement")
	logging.PhaseStart(phaseLog, "movement")
	phaseLog.Info("phase begin")

	if e.World == nil {
		return fmt.Errorf("world engine not found")
	}

	worldLog := phaseLog.With("engine", "world")
	tickMutations, err := e.World.TickMovementOrders(faction, e.Rulebook, worldLog)
	if err != nil {
		observer.OnError(faction, err)
		return err
	}
	if len(tickMutations) > 0 {
		tickMutations, err = dispatch.MutationReactors(e.Hooks, faction, tickMutations, factionState, e.Rulebook)
		if err != nil {
			observer.OnError(faction, err)
			return err
		}
		if err := e.applyAndRecord(factionState, faction, tickMutations, cfg, phaseLog); err != nil {
			observer.OnError(faction, err)
			return err
		}
		observer.OnMovementTicked(faction, tickMutations)
	}

	decisionMutations, err := prepareMovementDecisions(faction, collectors.Phase, e.World, e.Rulebook, worldLog)
	if err != nil {
		observer.OnError(faction, err)
		return err
	}
	if len(decisionMutations) > 0 {
		decisionMutations, err = dispatch.MutationReactors(e.Hooks, faction, decisionMutations, factionState, e.Rulebook)
		if err != nil {
			observer.OnError(faction, err)
			return err
		}
		if err := e.applyAndRecord(factionState, faction, decisionMutations, cfg, phaseLog); err != nil {
			observer.OnError(faction, err)
			return err
		}
	}

	observer.OnMovementResolved(faction, decisionMutations)
	if err := collectors.Phase.AwaitCheckpoint(CheckpointMovement); err != nil {
		observer.OnError(faction, err)
		return err
	}
	phaseLog.Info("phase end", "tick_mutations", len(tickMutations), "decision_mutations", len(decisionMutations))
	return nil
}

func (e *Engine) runActionPhase(
	faction *domain.Faction,
	factionState *state.FactionState,
	cfg *config.Config,
	collectors Collectors,
	observer TurnObserver,
	lock locks.GoalLock,
	turnLog *slog.Logger,
) error {
	phaseLog := turnLog.With("phase", "action")
	logging.PhaseStart(phaseLog, "action")
	phaseLog.Info("phase begin")

	actionLog := phaseLog.With("engine", "action")
	available := e.Action.AvailableActions(faction, factionState, e.Rulebook, collectors.Action, actionLog)
	if lock.Type == locks.LockRestrictActions {
		available = filterAllowedActions(available, lock.AllowedActions)
	}

	selectedAction, err := collectors.Phase.SelectAction(faction, available)
	if err != nil {
		observer.OnError(faction, err)
		return err
	}
	if selectedAction == nil {
		observer.OnFactionSkipped(faction)
		phaseLog.Info("phase end", "outcome", "skipped")
		return nil
	}

	observer.OnActionSelected(faction, selectedAction)

	actionMutations, err := e.Action.Run(selectedAction, faction, factionState, e.Rulebook, actionLog)
	if err != nil {
		observer.OnError(faction, err)
		return err
	}

	var worldIndex *world.Index
	if e.World == nil {
		return fmt.Errorf("world engine not found")
	}
	worldIndex = e.World.Index
	goalMutations := e.Goal.UpdateProgress(faction.ID, actionMutations, factionState, e.Rulebook, worldIndex, phaseLog.With("engine", "goal"))
	combined := append(actionMutations, goalMutations...)

	combined, err = dispatch.MutationReactors(e.Hooks, faction, combined, factionState, e.Rulebook)
	if err != nil {
		observer.OnError(faction, err)
		return err
	}

	if err := e.applyAndRecord(factionState, faction, combined, cfg, phaseLog); err != nil {
		observer.OnError(faction, err)
		return err
	}
	observer.OnActionResolved(faction, selectedAction, combined)
	if err := collectors.Phase.AwaitCheckpoint(CheckpointActionResult); err != nil {
		observer.OnError(faction, err)
		return err
	}

	phaseLog.Info("phase end", "action", selectedAction.Name(), "mutations", len(combined))
	return nil
}

// finishFactionTurn observes turn completion, advances the turn cursor,
// persists state, and — if the cycle just closed out — fires the cycle
// summary observer + checkpoint.
func (e *Engine) finishFactionTurn(
	factionState *state.FactionState,
	faction *domain.Faction,
	cfg *config.Config,
	collectors Collectors,
	observer TurnObserver,
	turnLog *slog.Logger,
) (bool, error) {
	observer.OnFactionTurnCompleted(faction)

	cycleDone, err := e.Turn.Advance(factionState, turnLog.With("engine", "turn"))
	if err != nil {
		observer.OnError(faction, err)
		return false, err
	}
	if err := state.Save(cfg.StatePath, factionState); err != nil {
		observer.OnError(faction, err)
		return false, fmt.Errorf("saving state: %w", err)
	}

	if cycleDone {
		observer.OnCycleCompleted(factionState.CycleNumber, factionState)
		if err := collectors.Phase.AwaitCheckpoint(CheckpointCycleSummary); err != nil {
			observer.OnError(faction, err)
			return cycleDone, err
		}
	}
	turnLog.Info("turn finished", "cycle_done", cycleDone)
	return cycleDone, nil
}

// applyAndRecord applies mutations to in-memory state and appends a single
// EventRecord to the history file. It does not persist state to disk — callers
// are responsible for calling state.Save at phase-gate checkpoints.
func (e *Engine) applyAndRecord(
	factionState *state.FactionState,
	faction *domain.Faction,
	mutations []domain.Mutation,
	cfg *config.Config,
	phaseLog *slog.Logger,
) error {
	if len(mutations) == 0 {
		return nil
	}
	if err := e.Mutation.Apply(factionState, mutations, phaseLog.With("engine", "mutation")); err != nil {
		return fmt.Errorf("applying mutations: %w", err)
	}
	if err := turn.RecordHistory(cfg.HistoryPath, factionState, faction, mutations, phaseLog.With("engine", "turn")); err != nil {
		return fmt.Errorf("recording history: %w", err)
	}
	return nil
}

// filterAllowedActions narrows a list of available actions to those whose
// Name() appears in allowed. Used when a goal lock restricts the action set
// (e.g. Planetary Seizure Phase 1 → Attack only).
func filterAllowedActions(available []action.Action, allowed []string) []action.Action {
	if len(allowed) == 0 {
		return nil
	}
	allowedSet := make(map[string]bool, len(allowed))
	for _, name := range allowed {
		allowedSet[name] = true
	}
	filtered := make([]action.Action, 0, len(available))
	for _, a := range available {
		if allowedSet[a.Name()] {
			filtered = append(filtered, a)
		}
	}
	return filtered
}

func eligibleStatRaises(faction *domain.Faction) []domain.FactionStat {
	type entry struct {
		stat   domain.FactionStat
		rating int
	}
	all := []entry{
		{stat: domain.StatForce, rating: faction.Force},
		{stat: domain.StatCunning, rating: faction.Cunning},
		{stat: domain.StatWealth, rating: faction.Wealth},
	}
	eligible := make([]domain.FactionStat, 0, 3)
	for _, e := range all {
		if e.rating < 8 && faction.XP >= domain.HPValueForRating(e.rating+1) {
			eligible = append(eligible, e.stat)
		}
	}
	return eligible
}

func buildStatRaiseMutations(faction *domain.Faction, stat domain.FactionStat) []domain.Mutation {
	var oldRating int
	switch stat {
	case domain.StatForce:
		oldRating = faction.Force
	case domain.StatCunning:
		oldRating = faction.Cunning
	case domain.StatWealth:
		oldRating = faction.Wealth
	}
	cost := domain.HPValueForRating(oldRating + 1)
	return []domain.Mutation{
		domain.XPSpent{FactionID: faction.ID, Amount: cost, Cause: "stat_raise"},
		domain.StatRaised{FactionID: faction.ID, Stat: stat, OldRating: oldRating, NewRating: oldRating + 1, Cause: "stat_raise"},
	}
}

func prepareStatRaise(faction *domain.Faction, collector PhaseCollector) (*domain.FactionStat, []domain.Mutation, error) {
	eligible := eligibleStatRaises(faction)
	if len(eligible) == 0 {
		return nil, nil, nil // skip: don't call the collector
	}
	stat, err := collector.SelectStatRaise(faction, eligible)
	if err != nil {
		return nil, nil, fmt.Errorf("selecting stat raise: %w", err)
	}
	if stat == nil {
		return nil, nil, nil // player declined
	}
	return stat, buildStatRaiseMutations(faction, *stat), nil
}

func prepareMovementDecisions(
	faction *domain.Faction,
	collector PhaseCollector,
	worldEngine *world.WorldEngine,
	rulebook *rulebook.Rulebook,
	worldLog *slog.Logger,
) ([]domain.Mutation, error) {
	eligible := eligibleMovableAssets(faction, rulebook)
	if len(eligible) == 0 {
		return nil, nil
	}
	decisions, err := collector.SelectMovementDecisions(faction, eligible)
	if err != nil {
		return nil, fmt.Errorf("selecting movement decisions: %w", err)
	}
	if len(decisions) == 0 {
		return nil, nil
	}

	for i, decision := range decisions {
		if decision.Kind != world.MovementDecisionIssue {
			continue
		}
		asset, ok := faction.Assets[decision.AssetID]
		if !ok {
			return nil, fmt.Errorf("movement decision for unknown asset ID %q", decision.AssetID)
		}
		def := rulebook.Assets[asset.DefinitionID]
		if def == nil || def.Transport == nil {
			continue
		}
		cargoEligible := eligibleCargoForTransport(faction, asset, def.Transport, rulebook)
		cargoSelected, err := collector.SelectTransportCargo(asset, cargoEligible, def.Transport)
		if err != nil {
			return nil, fmt.Errorf("selecting transport cargo: %w", err)
		}
		if len(cargoSelected) > def.Transport.MaxCargo {
			return nil, fmt.Errorf("selected %d cargo assets, exceeding transport max of %d", len(cargoSelected), def.Transport.MaxCargo)
		}
		eligibleIDs := make(map[string]bool, len(cargoEligible))
		for _, c := range cargoEligible {
			eligibleIDs[c.ID] = true
		}
		cargoIDs := make([]string, len(cargoSelected))
		for j, c := range cargoSelected {
			if !eligibleIDs[c.ID] {
				return nil, fmt.Errorf("selected cargo asset %q is not eligible for transport", c.ID)
			}
			cargoIDs[j] = c.ID
		}
		decisions[i].CargoAssetIDs = cargoIDs
	}

	return worldEngine.BuildMovementMutations(decisions, faction, rulebook, worldLog)
}

func eligibleMovableAssets(faction *domain.Faction, rulebook *rulebook.Rulebook) []*domain.Asset {
	var eligible []*domain.Asset
	for _, asset := range faction.Assets {
		def := rulebook.Assets[asset.DefinitionID]
		if def != nil && def.Speed > 0 {
			eligible = append(eligible, asset)
		}
	}
	return eligible
}

func eligibleCargoForTransport(faction *domain.Faction, transport *domain.Asset, profile *domain.TransportProfile, rulebook *rulebook.Rulebook) []*domain.Asset {
	var eligible []*domain.Asset
	for _, asset := range faction.Assets {
		if asset.ID == transport.ID {
			continue
		}
		def := rulebook.Assets[asset.DefinitionID]
		if def == nil || asset.CurrentOrder != nil || def.Speed > 0 {
			continue
		}
		if slices.Contains(profile.CargoTypes, def.Type) &&
			!slices.Contains(profile.ExcludeCategories, def.Category) &&
			isCoLocated(asset.Location, transport.Location) {
			eligible = append(eligible, asset)
		}
	}
	return eligible
}

func isCoLocated(a, b domain.Location) bool {
	return a.WorldID == b.WorldID && a.RegionHex == b.RegionHex
}

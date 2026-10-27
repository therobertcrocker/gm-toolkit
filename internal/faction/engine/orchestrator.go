package engine

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/relocate"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/turn"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	factionErrors "github.com/therobertcrocker/gm-toolkit/internal/faction/errors"
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
	paths *campaigns.Paths,
	collectors Collectors,
	observer TurnObserver,
) error {
	logging.RunHeader(e.log, "faction-manager", "dev", factionState.CampaignID, len(factionState.Factions))

	// deliberate: reset before registering, because the TUI runs every cycle on one engine and each ApplyAll would add another copy of every hook.
	e.Hooks.Reset()
	e.Effect.ApplyAll(factionState, e.Rulebook, e.Hooks, e.Rand, e.log.With("engine", "effect"))

	for {
		done, err := e.RunFactionTurn(factionState, paths, collectors, observer)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}
}

// registerCapturedAssets moves each asset that changed owner in mutations to its
// new owner's hook bucket.
func (e *Engine) registerCapturedAssets(factionState *state.FactionState, mutations []domain.Mutation) {
	for _, mutation := range mutations {
		changed, ok := mutation.(domain.AssetOwnerChanged)
		if !ok {
			continue
		}
		// deliberate: clear the old bucket, not leave it, because a recapture in the same cycle would register the asset there a second time and its hooks would fire twice (Decision Record 13).
		e.Hooks.ClearAsset(changed.FactionID, changed.AssetID)
		owner := factionState.Factions[changed.ToFactionID]
		if owner == nil || owner.Assets[changed.AssetID] == nil {
			continue
		}
		e.Effect.ApplyAsset(owner, owner.Assets[changed.AssetID], e.Rulebook, e.Hooks, e.Rand)
	}
}

// RunFactionTurn drives one faction's turn from setup through state save:
// stat-raise, bookkeeping, movement, goal-lock (gating the action), action.
// Errors are reported to the observer via OnError and returned to the
// caller; partial mutations already applied stay applied.
//
// The boolean return reports whether this faction's turn closed out the
// current cycle (true when the turn cursor advanced past the last faction).
func (e *Engine) RunFactionTurn(
	factionState *state.FactionState,
	paths *campaigns.Paths,
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

	if err := e.runStatRaisePhase(faction, factionState, paths, collectors, observer, turnLog); err != nil {
		return false, err
	}

	if err := e.runBookkeepingPhase(faction, factionState, paths, collectors, observer, turnLog); err != nil {
		return false, err
	}
	if err := state.Save(paths.StatePath, factionState); err != nil {
		if factionErrors.IsRecoverable(err) {
			turnLog.Warn("recoverable failure", "err", err)
			observer.OnError(faction, err)
			return false, nil
		}
		turnLog.Error("failure", "err", err)
		observer.OnError(faction, err)
		return false, fmt.Errorf("saving state after bookkeeping: %w", err)
	}

	if err := e.runMovementPhase(faction, factionState, paths, collectors, observer, turnLog); err != nil {
		return false, err
	}

	lock, err := e.runGoalLockPhase(faction, factionState, paths, collectors, observer, turnLog)
	if err != nil {
		return false, err
	}

	if lock.Type == locks.LockSkip {
		return e.finishFactionTurn(factionState, faction, paths, collectors, observer, turnLog)
	}

	if err := e.runActionPhase(faction, factionState, paths, collectors, observer, lock, turnLog); err != nil {
		return false, err
	}
	if err := state.Save(paths.StatePath, factionState); err != nil {
		if factionErrors.IsRecoverable(err) {
			turnLog.Warn("recoverable failure", "err", err)
			observer.OnError(faction, err)
			return false, nil
		}
		turnLog.Error("failure", "err", err)
		observer.OnError(faction, err)
		return false, fmt.Errorf("saving state after action resolution: %w", err)
	}

	return e.finishFactionTurn(factionState, faction, paths, collectors, observer, turnLog)
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
	paths *campaigns.Paths,
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
		if err := e.applyAndRecord(factionState, faction, lockMutations, paths, phaseLog); err != nil {
			if factionErrors.IsRecoverable(err) {
				phaseLog.Warn("recoverable failure", "err", err)
				observer.OnError(faction, err)
				return lock, nil
			}
			phaseLog.Error("failure", "err", err)
			observer.OnError(faction, err)
			return lock, err
		}
	}

	if lock.Type == locks.LockSkip {
		if err := collectors.Phase.AwaitCheckpoint(CheckpointGoalLocked); err != nil {
			if factionErrors.IsRecoverable(err) {
				phaseLog.Warn("recoverable failure", "err", err)
				observer.OnError(faction, err)
				return lock, nil
			}
			phaseLog.Error("failure", "err", err)
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
	paths *campaigns.Paths,
	collectors Collectors,
	observer TurnObserver,
	turnLog *slog.Logger,
) error {
	phaseLog := turnLog.With("phase", "bookkeeping")
	logging.PhaseStart(phaseLog, "bookkeeping")
	phaseLog.Info("phase begin")

	var liftMutations, startMutations []domain.Mutation
	// deliberate: fire here, gated on BookkeepingApplied, not in turn setup, because setup re-runs on every resume and would pay a turn-start effect twice.
	if e.Turn.InProgress(factionState) && !factionState.CurrentTurn.BookkeepingApplied {
		var err error
		// deliberate: lifts apply before the turn-start effects, not with them, because a lifted asset is active again and its own turn_start entries fire this turn (Effort 3 Decision 8).
		if liftMutations, err = e.runConditionLifts(faction, factionState, paths, collectors.Phase, observer, phaseLog); err != nil {
			return err
		}
		startMutations = append(turn.ExpireConditions(factionState, faction.ID), e.Effect.FireTurnStart(faction, factionState, e.Rulebook, e.Rand)...)
		if err := e.applyAndRecord(factionState, faction, startMutations, paths, phaseLog); err != nil {
			if !factionErrors.IsRecoverable(err) {
				phaseLog.Error("failure", "err", err)
				observer.OnError(faction, err)
				return err
			}
			// deliberate: fall through to bookkeeping, not return, because returning here would skip this turn's income and maintenance.
			phaseLog.Warn("recoverable failure", "err", err)
			observer.OnError(faction, err)
			startMutations = nil
		}
	}

	bookResult, bookMutations, err := e.Turn.ApplyBookkeeping(factionState, e.Hooks, phaseLog.With("engine", "turn"))
	if err != nil {
		if factionErrors.IsRecoverable(err) {
			phaseLog.Warn("recoverable failure", "err", err)
			observer.OnError(faction, err)
			return nil // continue past this sub-engine call
		}
		phaseLog.Error("failure", "err", err)
		observer.OnError(faction, err)
		return err
	}
	if len(bookMutations) > 0 {
		if err := e.applyAndRecord(factionState, faction, bookMutations, paths, phaseLog); err != nil {
			if factionErrors.IsRecoverable(err) {
				phaseLog.Warn("recoverable failure", "err", err)
				observer.OnError(faction, err)
				return nil // continue past this sub-engine call
			}
			phaseLog.Error("failure", "err", err)
			observer.OnError(faction, err)
			return err
		}
	}
	observer.OnBookkeepingApplied(faction, bookResult, slices.Concat(liftMutations, startMutations, bookMutations))
	if err := collectors.Phase.AwaitCheckpoint(CheckpointBookkeeping); err != nil {
		if factionErrors.IsRecoverable(err) {
			phaseLog.Warn("recoverable failure", "err", err)
			observer.OnError(faction, err)
			return nil // continue past this sub-engine call
		}
		phaseLog.Error("failure", "err", err)
		observer.OnError(faction, err)
		return err
	}
	phaseLog.Info("phase end", "mutations", len(bookMutations))
	return nil
}

// runConditionLifts offers the faction's disabled assets their price and applies
// the lifts it pays for. A recoverable failure is reported and lifts nothing.
func (e *Engine) runConditionLifts(faction *domain.Faction, factionState *state.FactionState, paths *campaigns.Paths, collector PhaseCollector, observer TurnObserver, phaseLog *slog.Logger) ([]domain.Mutation, error) {
	mutations, err := turn.OfferConditionLifts(faction, e.Rulebook, collector)
	if err == nil && len(mutations) > 0 {
		err = e.applyAndRecord(factionState, faction, mutations, paths, phaseLog)
	}
	if err == nil {
		return mutations, nil
	}
	if !factionErrors.IsRecoverable(err) {
		phaseLog.Error("failure", "err", err)
		observer.OnError(faction, err)
		return nil, err
	}
	phaseLog.Warn("recoverable failure", "err", err)
	observer.OnError(faction, err)
	return nil, nil
}

func (e *Engine) runStatRaisePhase(
	faction *domain.Faction,
	factionState *state.FactionState,
	paths *campaigns.Paths,
	collectors Collectors,
	observer TurnObserver,
	turnLog *slog.Logger,
) error {
	phaseLog := turnLog.With("phase", "stat_raise")
	logging.PhaseStart(phaseLog, "stat_raise")
	phaseLog.Info("phase begin")

	statToRaise, raiseMutations, err := prepareStatRaise(faction, collectors.Phase)
	if err != nil {
		if factionErrors.IsRecoverable(err) {
			phaseLog.Warn("recoverable failure", "err", err)
			observer.OnError(faction, err)
			return nil // continue past this sub-engine call
		}
		phaseLog.Error("failure", "err", err)
		observer.OnError(faction, err)
		return err
	}
	if statToRaise == nil {
		observer.OnStatRaiseSkipped(faction)
		phaseLog.Info("phase end", "outcome", "skipped")
		return nil
	}
	if len(raiseMutations) > 0 {
		if err := e.applyAndRecord(factionState, faction, raiseMutations, paths, phaseLog); err != nil {
			if factionErrors.IsRecoverable(err) {
				phaseLog.Warn("recoverable failure", "err", err)
				observer.OnError(faction, err)
				return nil // continue past this sub-engine call
			}
			phaseLog.Error("failure", "err", err)
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
	paths *campaigns.Paths,
	collectors Collectors,
	observer TurnObserver,
	turnLog *slog.Logger,
) error {
	phaseLog := turnLog.With("phase", "movement")
	logging.PhaseStart(phaseLog, "movement")
	phaseLog.Info("phase begin")

	if e.World == nil {
		return ErrWorldEngineUnavailable
	}

	worldLog := phaseLog.With("engine", "world")
	tickMutations, err := e.World.TickMovementOrders(faction, e.Rulebook, worldLog)
	if err != nil {
		if factionErrors.IsRecoverable(err) {
			phaseLog.Warn("recoverable failure", "err", err)
			observer.OnError(faction, err)
			return nil // continue past this sub-engine call
		}
		phaseLog.Error("failure", "err", err)
		observer.OnError(faction, err)
		return err
	}
	if len(tickMutations) > 0 {
		tickMutations, err = e.checkArrivals(faction, factionState, tickMutations, collectors.Phase)
		if err != nil {
			if factionErrors.IsRecoverable(err) {
				phaseLog.Warn("recoverable failure", "err", err)
				observer.OnError(faction, err)
				return nil // continue past this sub-engine call
			}
			phaseLog.Error("failure", "err", err)
			observer.OnError(faction, err)
			return err
		}
		tickMutations, err = dispatch.MutationReactors(e.Hooks, faction, tickMutations, factionState, e.Rulebook)
		if err != nil {
			if factionErrors.IsRecoverable(err) {
				phaseLog.Warn("recoverable failure", "err", err)
				observer.OnError(faction, err)
				return nil // continue past this sub-engine call
			}
			phaseLog.Error("failure", "err", err)
			observer.OnError(faction, err)
			return err
		}
		if err := e.applyAndRecord(factionState, faction, tickMutations, paths, phaseLog); err != nil {
			if factionErrors.IsRecoverable(err) {
				phaseLog.Warn("recoverable failure", "err", err)
				observer.OnError(faction, err)
				return nil // continue past this sub-engine call
			}
			phaseLog.Error("failure", "err", err)
			observer.OnError(faction, err)
			return err
		}
		observer.OnMovementTicked(faction, tickMutations)
	}

	decisionMutations, err := prepareMovementDecisions(faction, collectors.Phase, e.World, e.Rulebook, worldLog)
	if err != nil {
		if factionErrors.IsRecoverable(err) {
			phaseLog.Warn("recoverable failure", "err", err)
			observer.OnError(faction, err)
			return nil // continue past this sub-engine call
		}
		phaseLog.Error("failure", "err", err)
		observer.OnError(faction, err)
		return err
	}
	if len(decisionMutations) > 0 {
		decisionMutations, err = dispatch.MutationReactors(e.Hooks, faction, decisionMutations, factionState, e.Rulebook)
		if err != nil {
			if factionErrors.IsRecoverable(err) {
				phaseLog.Warn("recoverable failure", "err", err)
				observer.OnError(faction, err)
				return nil // continue past this sub-engine call
			}
			phaseLog.Error("failure", "err", err)
			observer.OnError(faction, err)
			return err
		}
		if err := e.applyAndRecord(factionState, faction, decisionMutations, paths, phaseLog); err != nil {
			if factionErrors.IsRecoverable(err) {
				phaseLog.Warn("recoverable failure", "err", err)
				observer.OnError(faction, err)
				return nil // continue past this sub-engine call
			}
			phaseLog.Error("failure", "err", err)
			observer.OnError(faction, err)
			return err
		}
	}

	grantedMutations, err := e.runGrantedMovement(faction, factionState, paths, collectors.Phase, observer, phaseLog)
	if err != nil {
		if factionErrors.IsRecoverable(err) {
			phaseLog.Warn("recoverable failure", "err", err)
			observer.OnError(faction, err)
			return nil // continue past this sub-engine call
		}
		phaseLog.Error("failure", "err", err)
		observer.OnError(faction, err)
		return err
	}

	observer.OnMovementResolved(faction, append(decisionMutations, grantedMutations...))
	if err := collectors.Phase.AwaitCheckpoint(CheckpointMovement); err != nil {
		if factionErrors.IsRecoverable(err) {
			phaseLog.Warn("recoverable failure", "err", err)
			observer.OnError(faction, err)
			return nil // continue past this sub-engine call
		}
		phaseLog.Error("failure", "err", err)
		observer.OnError(faction, err)
		return err
	}
	phaseLog.Info("phase end", "tick_mutations", len(tickMutations), "decision_mutations", len(decisionMutations), "granted_mutations", len(grantedMutations))
	return nil
}

func (e *Engine) runActionPhase(
	faction *domain.Faction,
	factionState *state.FactionState,
	paths *campaigns.Paths,
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
		if factionErrors.IsRecoverable(err) {
			phaseLog.Warn("recoverable failure", "err", err)
			observer.OnError(faction, err)
			return nil // continue past this sub-engine call
		}
		phaseLog.Error("failure", "err", err)
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
		if factionErrors.IsRecoverable(err) {
			phaseLog.Warn("recoverable failure", "err", err)
			observer.OnError(faction, err)
			return nil // continue past this sub-engine call
		}
		phaseLog.Error("failure", "err", err)
		observer.OnError(faction, err)
		return err
	}

	var worldIndex *world.Index
	if e.World == nil {
		return ErrWorldEngineUnavailable
	}
	worldIndex = e.World.Index
	actionMutations = append(actionMutations, e.firePurchaseEffects(faction, actionMutations, factionState)...)

	combined, err := dispatch.MutationReactors(e.Hooks, faction, actionMutations, factionState, e.Rulebook)
	if err != nil {
		if factionErrors.IsRecoverable(err) {
			phaseLog.Warn("recoverable failure", "err", err)
			observer.OnError(faction, err)
			return nil // continue past this sub-engine call
		}
		phaseLog.Error("failure", "err", err)
		observer.OnError(faction, err)
		return err
	}
	// deliberate: goals read the batch after the reactors, not before, because a reactor's capture or removal is progress too (Effort 1 Decision Record 12).
	combined = append(combined, e.Goal.UpdateProgress(faction.ID, combined, factionState, e.Rulebook, worldIndex, phaseLog.With("engine", "goal"))...)

	if err := e.applyAndRecord(factionState, faction, combined, paths, phaseLog); err != nil {
		if factionErrors.IsRecoverable(err) {
			phaseLog.Warn("recoverable failure", "err", err)
			observer.OnError(faction, err)
			return nil // continue past this sub-engine call
		}
		phaseLog.Error("failure", "err", err)
		observer.OnError(faction, err)
		return err
	}
	// deliberate: register the captured asset now, not at the next cycle, and only that asset, not every hook, because its hooks are keyed under its old owner, and a full re-prime would also wake hooks bought this cycle (Decision 9).
	e.registerCapturedAssets(factionState, combined)
	observer.OnActionResolved(faction, selectedAction, combined)
	if err := collectors.Phase.AwaitCheckpoint(CheckpointActionResult); err != nil {
		if factionErrors.IsRecoverable(err) {
			phaseLog.Warn("recoverable failure", "err", err)
			observer.OnError(faction, err)
			return nil // continue past this sub-engine call
		}
		phaseLog.Error("failure", "err", err)
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
	paths *campaigns.Paths,
	collectors Collectors,
	observer TurnObserver,
	turnLog *slog.Logger,
) (bool, error) {
	observer.OnFactionTurnCompleted(faction)

	cycleDone, err := e.Turn.Advance(factionState, turnLog.With("engine", "turn"))
	if err != nil {
		if factionErrors.IsRecoverable(err) {
			turnLog.Warn("recoverable failure", "err", err)
			observer.OnError(faction, err)
			return false, nil
		}
		turnLog.Error("failure", "err", err)
		observer.OnError(faction, err)
		return false, err
	}
	if err := state.Save(paths.StatePath, factionState); err != nil {
		if factionErrors.IsRecoverable(err) {
			turnLog.Warn("recoverable failure", "err", err)
			observer.OnError(faction, err)
			return false, nil
		}
		turnLog.Error("failure", "err", err)
		observer.OnError(faction, err)
		return false, fmt.Errorf("saving state: %w", err)
	}

	if cycleDone {
		observer.OnCycleCompleted(factionState.CycleNumber, factionState)
		if err := collectors.Phase.AwaitCheckpoint(CheckpointCycleSummary); err != nil {
			if factionErrors.IsRecoverable(err) {
				turnLog.Warn("recoverable failure", "err", err)
				observer.OnError(faction, err)
				return cycleDone, nil
			}
			turnLog.Error("failure", "err", err)
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
	paths *campaigns.Paths,
	phaseLog *slog.Logger,
) error {
	if len(mutations) == 0 {
		return nil
	}
	if err := e.Mutation.Apply(factionState, mutations, phaseLog.With("engine", "mutation")); err != nil {
		for _, miss := range err.Misses {
			phaseLog.Error("mutation miss",
				"mutation_type", miss.MutationType,
				"faction_id", miss.FactionID,
				"entity_id", miss.EntityID,
			)
		}
		return fmt.Errorf("applying mutations: %w", err)
	}
	if err := turn.RecordHistory(paths.HistoryPath, factionState, faction, mutations, phaseLog.With("engine", "turn")); err != nil {
		return fmt.Errorf("recording history: %w", err)
	}
	return nil
}

// filterAllowedActions narrows a list of available actions to those whose
// Name() appears in allowed. Used when a goal lock restricts the action set
// (e.g. Planetary Seizure Phase 1 → Attack only).
// firePurchaseEffects runs the on_purchase effects of every asset the action
// bought, in batch order. The asset is the copy the AssetAdded carries, because
// it is not in state until the batch applies.
func (e *Engine) firePurchaseEffects(faction *domain.Faction, actionMutations []domain.Mutation, factionState *state.FactionState) []domain.Mutation {
	var mutations []domain.Mutation
	for _, mutation := range actionMutations {
		added, ok := mutation.(domain.AssetAdded)
		// deliberate: whitelist the purchase causes, not exclude refit, so a new AssetAdded emitter fires nothing until it is judged a purchase.
		if !ok || (added.Cause != "buy" && added.Cause != "ability") {
			continue
		}
		mutations = append(mutations, e.Effect.FirePurchase(faction, &added.Asset, factionState, e.Rulebook, e.Rand)...)
	}
	return mutations
}

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
		// deliberate: a sheltered asset stays put, not moves, because Boltholes make it unusable until repaired (Effort 2 Decision 11).
		if def != nil && def.Speed > 0 && !asset.HasCondition(domain.ConditionSheltered) {
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
		if def == nil || asset.CurrentOrder != nil || def.Speed > 0 || asset.HasCondition(domain.ConditionSheltered) {
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

// runGrantedMovement offers each free movement grant once, after the drift
// decisions apply: cargo loaded this phase is already in transit, so no grant can
// move it a second time. Each grant is its own batch, so the next grant sees the
// Coin and locations the last one changed. A grant with nothing to move, that the
// faction cannot pay for, or that the GM declines is skipped.
func (e *Engine) runGrantedMovement(
	faction *domain.Faction,
	factionState *state.FactionState,
	paths *campaigns.Paths,
	collector PhaseCollector,
	observer TurnObserver,
	phaseLog *slog.Logger,
) ([]domain.Mutation, error) {
	var resolved []domain.Mutation
	for _, grant := range dispatch.GrantedMovementAbilities(e.Hooks, faction) {
		params := grant.Relocate
		mutations, err := relocate.Run(relocate.Request{
			Faction:      faction,
			Acting:       grant.Source,
			Params:       &params,
			World:        e.World,
			Rulebook:     e.Rulebook,
			Registry:     e.Hooks,
			FactionState: factionState,
			Collector:    collector,
		})
		if errors.Is(err, relocate.ErrInsufficientCoin) {
			phaseLog.Warn("granted movement skipped", "source", grant.Source.ID, "err", err)
			observer.OnError(faction, err)
			continue
		}
		// deliberate: Esc declines this grant, not the turn, because the decision batch has already applied and must still resolve.
		if errors.Is(err, relocate.ErrNoCandidates) || errors.Is(err, action.ErrNoSelection) || errors.Is(err, action.ErrTurnCanceled) {
			continue
		}
		if err != nil {
			return resolved, fmt.Errorf("granted movement from %q: %w", grant.Source.ID, err)
		}
		mutations, err = dispatch.MutationReactors(e.Hooks, faction, mutations, factionState, e.Rulebook)
		if err != nil {
			return resolved, err
		}
		if err := e.applyAndRecord(factionState, faction, mutations, paths, phaseLog); err != nil {
			return resolved, err
		}
		resolved = append(resolved, mutations...)
	}
	return resolved, nil
}

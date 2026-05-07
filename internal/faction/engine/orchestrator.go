package engine

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/config"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// Checkpoint constants name the pipeline phases where the orchestrator pauses
// for caller acknowledgement via InputCollector.AwaitCheckpoint.
const (
	CheckpointBookkeeping  = "bookkeeping"
	CheckpointActionResult = "action_result"
	CheckpointGoalLocked   = "goal_locked"
	CheckpointCycleSummary = "cycle_summary"
)

// RunFactionTurn drives one faction's turn from goal-lock check through state
// save. Errors are reported to the observer via OnError and returned to the
// caller; partial mutations already applied stay applied.
//
// The boolean return reports whether this faction's turn closed out the
// current cycle (true when the turn cursor advanced past the last faction).
func (e *Engine) RunFactionTurn(
	factionState *state.FactionState,
	cfg *config.Config,
	collector InputCollector,
	observer TurnObserver,
) (bool, error) {
	faction, err := e.Turn.CurrentFaction(factionState)
	if err != nil {
		observer.OnError(nil, err)
		return false, err
	}
	observer.OnFactionTurnStarted(faction)

	// Phase 1: Goal Lock Check. The engine queries the Goal subsystem for locks
	// applying to this faction at the start of its turn, and any mutations
	// resulting from those locks are applied immediately. The lock type (if
	// any) determines whether the faction is allowed to select an action this
	// turn or is forced to skip.

	lock, lockMutations := e.Goal.CheckLock(faction, factionState, e.Rulebook)
	observer.OnGoalLockApplied(faction, lock, lockMutations)

	if lock.Type == goal.LockSkip {
		if len(lockMutations) > 0 {
			if err := e.applyAndRecord(factionState, faction, lockMutations, cfg); err != nil {
				observer.OnError(faction, err)
				return false, err
			}
		}
		if err := collector.AwaitCheckpoint(CheckpointGoalLocked); err != nil {
			observer.OnError(faction, err)
			return false, err
		}
		return e.finishFactionTurn(factionState, faction, cfg, collector, observer)
	}

	if len(lockMutations) > 0 {
		if err := e.applyAndRecord(factionState, faction, lockMutations, cfg); err != nil {
			observer.OnError(faction, err)
			return false, err
		}
	}

	// Phase 2: Bookkeeping. The engine applies any bookkeeping mutations before action selection,
	// so that they can affect available actions and be observed by the caller.

	bookResult, bookMutations, err := e.Turn.ApplyBookkeeping(factionState, e.Hooks)
	if err != nil {
		observer.OnError(faction, err)
		return false, err
	}
	if len(bookMutations) > 0 {
		if err := e.applyAndRecord(factionState, faction, bookMutations, cfg); err != nil {
			observer.OnError(faction, err)
			return false, err
		}
	}
	observer.OnBookkeepingApplied(faction, bookResult, bookMutations)
	if err := collector.AwaitCheckpoint(CheckpointBookkeeping); err != nil {
		observer.OnError(faction, err)
		return false, err
	}

	// Phase 3: Action Selection and Resolution. The engine queries the Action
	// subsystem for available actions, passing along the lock type and allowed
	// actions if relevant.

	available := e.Action.AvailableActions(faction, factionState, e.Rulebook, collector)
	if lock.Type == goal.LockRestrictActions {
		available = filterAllowedActions(available, lock.AllowedActions)
	}

	selectedAction, err := collector.SelectAction(faction, available)
	if err != nil {
		observer.OnError(faction, err)
		return false, err
	}
	if selectedAction == nil {
		observer.OnFactionSkipped(faction)
		return e.finishFactionTurn(factionState, faction, cfg, collector, observer)
	}

	observer.OnActionSelected(faction, selectedAction)

	actionMutations, err := e.Action.Run(selectedAction, faction, factionState, e.Rulebook)
	if err != nil {
		observer.OnError(faction, err)
		return false, err
	}
	goalMutations := e.Goal.UpdateProgress(faction.ID, actionMutations, factionState, e.Rulebook)
	combined := append(actionMutations, goalMutations...)

	// Phase 4: MutationReactor dispatch. Registered hooks (Cat 3) fire in
	// registration order; returned mutations are appended and the loop
	// recurses until no new mutations are produced.
	var dispatchErr error
	combined, dispatchErr = dispatch.MutationReactors(e.Hooks, faction, combined, factionState, e.Rulebook)
	if dispatchErr != nil {
		observer.OnError(faction, dispatchErr)
		return false, dispatchErr
	}

	// Phase 5: Apply mutations and persist state. The engine applies all mutations
	// in a single batch to preserve order, then records a single EventRecord in the
	// history file with the full mutation set. The state is also saved after mutation
	// application. If any of these steps fail, the error is reported and returned;
	// successfully applied mutations are not rolled back.

	if err := e.applyAndRecord(factionState, faction, combined, cfg); err != nil {
		observer.OnError(faction, err)
		return false, err
	}
	observer.OnActionResolved(faction, selectedAction, combined)
	if err := collector.AwaitCheckpoint(CheckpointActionResult); err != nil {
		observer.OnError(faction, err)
		return false, err
	}

	return e.finishFactionTurn(factionState, faction, cfg, collector, observer)
}

// RunCycle calls RunFactionTurn until a cycle completes. The caller must have
// already called Turn.Start (or be resuming a turn that's still InProgress).
func (e *Engine) RunCycle(
	factionState *state.FactionState,
	cfg *config.Config,
	collector InputCollector,
	observer TurnObserver,
) error {
	for {
		done, err := e.RunFactionTurn(factionState, cfg, collector, observer)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}
}

// finishFactionTurn observes turn completion, advances the turn cursor,
// persists state, and — if the cycle just closed out — fires the cycle
// summary observer + checkpoint.
func (e *Engine) finishFactionTurn(
	factionState *state.FactionState,
	faction *domain.Faction,
	cfg *config.Config,
	collector InputCollector,
	observer TurnObserver,
) (bool, error) {
	observer.OnFactionTurnCompleted(faction)

	cycleDone, err := e.Turn.Advance(factionState)
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
		if err := collector.AwaitCheckpoint(CheckpointCycleSummary); err != nil {
			observer.OnError(faction, err)
			return cycleDone, err
		}
	}
	return cycleDone, nil
}

// applyAndRecord is the canonical write path: apply mutations to in-memory
// state, append a single EventRecord to the history file, and persist state
// to disk. Mutation order is preserved; the orchestrator is responsible for
// composing the final ordered slice before invoking.
func (e *Engine) applyAndRecord(
	factionState *state.FactionState,
	faction *domain.Faction,
	mutations []domain.Mutation,
	cfg *config.Config,
) error {
	if len(mutations) == 0 {
		return nil
	}
	e.Mutation.Apply(factionState, mutations)
	if err := e.History.Record(cfg.HistoryPath, factionState, faction, mutations); err != nil {
		return fmt.Errorf("recording history: %w", err)
	}
	if err := state.Save(cfg.StatePath, factionState); err != nil {
		return fmt.Errorf("saving state: %w", err)
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

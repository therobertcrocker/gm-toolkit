package engine

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/config"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
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

	lock, lockMutations := e.Goal.CheckLock(faction, factionState, e.Rulebook)
	observer.OnGoalLockApplied(faction, lock, lockMutations)

	if lock.Type == LockSkip {
		if len(lockMutations) > 0 {
			if err := e.applyAndRecord(factionState, faction, lockMutations, cfg); err != nil {
				observer.OnError(faction, err)
				return false, err
			}
		}
		if err := collector.AwaitCheckpoint(PhaseGoalLocked); err != nil {
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

	bookResult, bookMutations, err := e.Turn.ApplyBookkeeping(factionState)
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
	if err := collector.AwaitCheckpoint(PhaseBookkeeping); err != nil {
		observer.OnError(faction, err)
		return false, err
	}

	available := e.Action.AvailableActions(faction, factionState, e.Rulebook, collector)
	if lock.Type == LockRestrictActions {
		available = filterAllowedActions(available, lock.AllowedActions)
	}

	action, err := collector.SelectAction(faction, available)
	if err != nil {
		observer.OnError(faction, err)
		return false, err
	}
	if action == nil {
		observer.OnFactionSkipped(faction)
		return e.finishFactionTurn(factionState, faction, cfg, collector, observer)
	}

	observer.OnActionSelected(faction, action)

	actionMutations, err := e.Action.Run(action, faction, factionState, e.Rulebook)
	if err != nil {
		observer.OnError(faction, err)
		return false, err
	}
	goalMutations := e.Goal.UpdateProgress(faction.ID, actionMutations, factionState, e.Rulebook)
	combined := append(actionMutations, goalMutations...)

	// === EventHook dispatch site (deferred per decision #5) ===
	// When the Tag Engine lands, registered hooks fire here in registration
	// order. Returned mutations are appended to `combined` and the loop
	// recurses with depth bound 5; on cap trip the engine logs and stops.

	if err := e.applyAndRecord(factionState, faction, combined, cfg); err != nil {
		observer.OnError(faction, err)
		return false, err
	}
	observer.OnActionResolved(faction, action, combined)
	if err := collector.AwaitCheckpoint(PhaseActionResult); err != nil {
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
		if err := collector.AwaitCheckpoint(PhaseCycleSummary); err != nil {
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
	record, err := buildEventRecord(factionState, faction, mutations)
	if err != nil {
		return fmt.Errorf("building event record: %w", err)
	}
	if err := e.History.Record(cfg.HistoryPath, record); err != nil {
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
func filterAllowedActions(available []Action, allowed []string) []Action {
	if len(allowed) == 0 {
		return nil
	}
	allowedSet := make(map[string]bool, len(allowed))
	for _, name := range allowed {
		allowedSet[name] = true
	}
	filtered := make([]Action, 0, len(available))
	for _, action := range available {
		if allowedSet[action.Name()] {
			filtered = append(filtered, action)
		}
	}
	return filtered
}

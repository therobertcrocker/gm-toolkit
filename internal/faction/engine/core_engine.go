package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/ability"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/history"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/mutation"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/turn"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

// Package engine is the turn pipeline for the faction system.
//
// Five files form the conceptual core:
//
//   - core_engine.go        — composition root; owns the Rulebook and all sub-engines
//   - core_orchestrator.go  — drives one faction's turn through the pipeline stages
//   - core_observer.go      — TurnObserver interface; fire-and-forget output channel
//   - core_input_collector.go — InputCollector interface; abstracts GM vs. AI vs. test input
//   - core_event_hook.go    — EventHook interface; mutation side-effects (Tag Engine hook point)
//
// Supporting files (roller.go) and sub-packages (action/, ability/, goal/,
// history/, mutation/, turn/, testharness/) provide the mechanics the
// orchestrator delegates to.

// Engine is the composition root. It owns the Rulebook and all sub-engines.
type Engine struct {
	Rulebook *rulebook.Rulebook
	Rand     domain.Roller
	Turn     *turn.TurnEngine
	Mutation *mutation.MutationEngine
	Action   *action.ActionEngine
	Ability  *ability.AbilityEngine
	History  *history.HistoryEngine
	Goal     *goal.GoalEngine
}

func New(dataDir string) (*Engine, error) {
	rb, err := rulebook.Load(dataDir)
	if err != nil {
		return nil, err
	}
	return NewWithRulebook(rb), nil
}

func NewWithRulebook(rb *rulebook.Rulebook) *Engine {
	e := &Engine{Rulebook: rb, Rand: NewRandRoller()}
	e.Turn = turn.New(e.Rand)
	e.Mutation = mutation.New()
	e.Action = action.New()
	e.Ability = ability.New()
	e.History = history.New()
	e.Goal = goal.New()
	return e
}

package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/ability"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/history"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/mutation"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/turn"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

// Package engine is the turn pipeline for the faction system.
//
//   - engine.go      — composition root; owns the Rulebook and all sub-engines
//   - orchestrator.go — drives one faction's turn through the pipeline stages
//   - observer.go    — TurnObserver interface; fire-and-forget output channel
//   - collector.go   — InputCollector interface; abstracts GM vs. AI vs. test input
//   - dispatch/      — category-based hook dispatch (Cat 3 mutations, Cat 1+2 rolls, Cat 4+5 rules)
//
// Supporting files (roller.go) and sub-packages (action/, ability/, goal/,
// history/, mutation/, turn/, testharness/) provide the mechanics the
// orchestrator delegates to.

// Engine is the composition root. It owns the Rulebook and all sub-engines.
type Engine struct {
	Rulebook *rulebook.Rulebook
	Rand     domain.Roller
	Hooks    *hooks.Registry
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
	e := &Engine{Rulebook: rb, Rand: NewRandRoller(), Hooks: hooks.NewRegistry()}
	e.Turn = turn.New(e.Rand)
	e.Mutation = mutation.New()
	e.Action = action.New()
	e.Ability = ability.New()
	e.History = history.New()
	e.Goal = goal.New()
	return e
}

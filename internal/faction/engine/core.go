package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/config"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/ability"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/history"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/mutation"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/turn"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

// Package engine is the turn pipeline for the faction system.
//
//   - core.go      — composition root; owns the Rulebook and all sub-engines
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
	World    *world.Engine
}

func New(cfg *config.Config) (*Engine, error) {
	rb, err := rulebook.Load(cfg.FactionDataDir)
	if err != nil {
		return nil, err
	}
	eng := NewWithRulebook(rb)
	if cfg.SpatialDataDir != "" {
		worldEngine, err := world.New(cfg.SpatialDataDir)
		if err != nil {
			return nil, err
		}
		eng.World = worldEngine
	}
	return eng, nil
}

func NewWithRulebook(rulebook *rulebook.Rulebook) *Engine {
	e := &Engine{Rulebook: rulebook, Rand: NewRandRoller(), Hooks: hooks.NewRegistry()}
	e.Turn = turn.New(e.Rand, e.Rulebook)
	e.Mutation = mutation.New()
	e.Action = action.New()
	e.Ability = ability.New()
	e.History = history.New()
	e.Goal = goal.New()
	return e
}

package engine

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/mutation"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/tag"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/turn"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

var ErrWorldEngineUnavailable = errors.New("engine: world engine unavailable")

// Package engine is the turn pipeline for the faction system.
//
//   - core.go      — composition root; owns the Rulebook and all sub-engines
//   - orchestrator.go — drives one faction's turn through the pipeline stages
//   - observer.go    — TurnObserver interface; fire-and-forget output channel
//   - collector.go   — PhaseCollector interface + Collectors struct; abstracts GM vs. AI vs. test input
//   - dispatch/      — category-based hook dispatch (Cat 3 mutations, Cat 1+2 rolls, Cat 4+5 rules)
//
// Supporting files (roller.go) and sub-packages (action/, goal/,
// mutation/, turn/, testharness/) provide the mechanics the orchestrator
// delegates to.

// Engine is the composition root. It owns the Rulebook and all sub-engines.
type Engine struct {
	Rulebook *rulebook.Rulebook
	Rand     domain.Roller
	Hooks    *hooks.Registry
	Turn     *turn.TurnEngine
	Tag      *tag.TagEngine
	Effect   *effect.EffectsEngine
	Mutation *mutation.MutationEngine
	Action   *action.ActionEngine
	Goal     *goal.GoalEngine
	World    *world.WorldEngine
	log      *slog.Logger
}

func NewWithRulebook(rulebook *rulebook.Rulebook, worldEngine *world.WorldEngine, log *slog.Logger) (*Engine, error) {
	e := &Engine{Rulebook: rulebook, Rand: NewRandRoller(log), Hooks: hooks.NewRegistry(), World: worldEngine, log: log}
	e.Turn = turn.New(e.Rand, e.Rulebook, log)
	e.Mutation = mutation.New(log)
	e.Action = action.New(log)
	e.Goal = goal.New(log)
	e.Tag = tag.New(log)
	e.Effect = effect.New(effects.Registrars(), log)
	for _, defID := range slices.Sorted(maps.Keys(rulebook.Assets)) {
		def := rulebook.Assets[defID]
		if def.Transport != nil {
			if err := e.Effect.Register(effects.NewTransportHandler(def)); err != nil {
				return nil, fmt.Errorf("engine: %w", err)
			}
		}
		if len(def.Effects) > 0 {
			if err := e.Effect.RegisterEffectsList(def); err != nil {
				return nil, fmt.Errorf("engine: %w", err)
			}
		}
	}
	actions.RegisterDefaultActions(e.Action, func() domain.Roller { return e.Rand }, e.Hooks, e.World, e.Rulebook, e.Effect)
	return e, nil
}

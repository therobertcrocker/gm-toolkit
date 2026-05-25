package engine

import (
	"errors"
	"log/slog"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
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

var (
	ErrSpatialDataDirRequired = errors.New("spatial data dir is required for world engine")
	ErrWorldEngineUnavailable = errors.New("engine: world engine unavailable")
)

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

func New(paths *campaigns.Paths, log *slog.Logger) (*Engine, error) {
	rb, err := rulebook.Load(paths.FactionDataDir)
	if err != nil {
		return nil, err
	}
	eng := NewWithRulebook(rb, log)
	if paths.SpatialDataDir == "" {
		return nil, ErrSpatialDataDirRequired
	}
	worldEngine, err := world.New(paths.SpatialDataDir, log)
	if err != nil {
		return nil, err
	}
	eng.World = worldEngine

	return eng, nil
}

func NewWithRulebook(rulebook *rulebook.Rulebook, log *slog.Logger) *Engine {
	e := &Engine{Rulebook: rulebook, Rand: NewRandRoller(log), Hooks: hooks.NewRegistry(), log: log}
	e.Turn = turn.New(e.Rand, e.Rulebook, log)
	e.Mutation = mutation.New(log)
	e.Action = action.New(log)
	e.Goal = goal.New(log)
	e.Tag = tag.New(log)
	e.Effect = effect.New(log)
	for _, def := range rulebook.Assets {
		if def.Transport != nil {
			e.Effect.Register(effects.NewTransportHandler(def))
		}
	}
	return e
}

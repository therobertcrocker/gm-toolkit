package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/ability"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/history"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/mutation"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/turn"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
)

// Engine is the composition root. It owns the Rulebook and all sub-engines.
type Engine struct {
	Rulebook *loader.Rulebook
	Rand     domain.Roller
	Turn     *turn.TurnEngine
	Mutation *mutation.MutationEngine
	Action   *action.ActionEngine
	Ability  *ability.AbilityEngine
	History  *history.HistoryEngine
	Goal     *goal.GoalEngine
}

func New(dataDir string) (*Engine, error) {
	rb, err := loader.Load(dataDir)
	if err != nil {
		return nil, err
	}
	return NewWithRulebook(rb), nil
}

func NewWithRulebook(rb *loader.Rulebook) *Engine {
	e := &Engine{Rulebook: rb, Rand: NewRandRoller()}
	e.Turn = turn.New(e.Rand)
	e.Mutation = mutation.New()
	e.Action = action.New()
	e.Ability = ability.New()
	e.History = history.New()
	e.Goal = goal.New()
	return e
}

package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
)

// Engine is the core orchestrator. It owns the Rulebook and composes all
// sub-engines. Sub-engines that are not yet implemented are nil.
type Engine struct {
	Rulebook      *loader.Rulebook
	Turn          *TurnEngine
	Mutation      *MutationEngine
	Action        *ActionEngine
	AbilityEngine *AbilityEngine
	History       *HistoryEngine
	// Goal *GoalEngine — future
	// Tag  *TagEngine  — future
}

func New(dataDir string) (*Engine, error) {
	rb, err := loader.Load(dataDir)
	if err != nil {
		return nil, err
	}
	e := &Engine{Rulebook: rb}
	e.Mutation = newMutationEngine()
	e.Turn = newTurnEngine(e.Mutation)
	e.Action = newActionEngine()
	e.AbilityEngine = NewAbilityEngine()
	e.History = newHistoryEngine()
	return e, nil
}

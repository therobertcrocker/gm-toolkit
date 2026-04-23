package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// Action is the contract all faction actions implement. The Action Engine calls
// Validate, Inputs, Resolve, and Output in order. Each action is a stateful
// struct — Inputs stores collected data, Resolve reads it, Output produces the
// mutation list.
type Action interface {
	// Name returns the display name shown in the action selection menu.
	Name() string
	// Validate confirms preconditions are met; gates action selection.
	Validate(faction *domain.Faction, factionState *state.FactionState, rulebook *loader.Rulebook) bool
	// Inputs collects what is needed before resolution — GM prompts or AI logic.
	Inputs(faction *domain.Faction, factionState *state.FactionState, rulebook *loader.Rulebook) error
	// Resolve executes the action logic using data collected by Inputs.
	Resolve(faction *domain.Faction, factionState *state.FactionState, rulebook *loader.Rulebook) error
	// Output produces the mutation list from resolved state.
	Output() ([]domain.Mutation, error)
}

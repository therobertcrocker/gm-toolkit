package ability

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/ability/steps"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// CustomAbilityHandler is a full override for a specific asset definition ID.
type CustomAbilityHandler func(
	faction *domain.Faction,
	asset *domain.Asset,
	collector Collector,
	roller domain.Roller,
	factionState *state.FactionState,
	rulebook *rulebook.Rulebook,
) ([]domain.Mutation, error)

type AbilityEngine struct {
	customHandlers map[string]CustomAbilityHandler
}

func New() *AbilityEngine {
	return &AbilityEngine{
		customHandlers: make(map[string]CustomAbilityHandler),
	}
}

// RegisterCustomHandler registers a bespoke handler for a specific asset definition ID.
func (ae *AbilityEngine) RegisterCustomHandler(defID string, handler CustomAbilityHandler) {
	ae.customHandlers[defID] = handler
}

// Run resolves an asset's ability. Returns nil, nil when the definition has no Ability.
func (ae *AbilityEngine) Run(
	faction *domain.Faction,
	asset *domain.Asset,
	def *domain.AssetDefinition,
	collector Collector,
	roller domain.Roller,
	factionState *state.FactionState,
	rulebook *rulebook.Rulebook,
) ([]domain.Mutation, error) {
	if handler, ok := ae.customHandlers[def.ID]; ok {
		return handler(faction, asset, collector, roller, factionState, rulebook)
	}
	if def.Ability == nil {
		return nil, nil
	}
	var mutations []domain.Mutation
	for _, step := range def.Ability.Steps {
		stepMutations, err := runStep(faction, asset, step, collector, roller, factionState, rulebook)
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, stepMutations...)
	}
	return mutations, nil
}

func runStep(
	faction *domain.Faction,
	asset *domain.Asset,
	step domain.AbilityStep,
	collector Collector,
	roller domain.Roller,
	factionState *state.FactionState,
	rulebook *rulebook.Rulebook,
) ([]domain.Mutation, error) {
	switch step.Type {
	case domain.AbilityStepMovement:
		return steps.MovementStepHandler(faction, asset, step, collector, roller, factionState, rulebook)
	case domain.AbilityStepFactionTest:
		return steps.FactionTestStepHandler(faction, asset, step, collector, roller, factionState, rulebook)
	default:
		return nil, fmt.Errorf("no handler for ability step type %q", step.Type)
	}
}

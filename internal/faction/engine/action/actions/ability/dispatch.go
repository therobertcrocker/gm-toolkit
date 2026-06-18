package ability

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// AbilityHandler resolves a single asset's special ability.
type AbilityHandler func(
	faction *domain.Faction,
	asset *domain.Asset,
	def *domain.AssetDefinition,
	collector action.Collector,
	roller domain.Roller,
	factionState *state.FactionState,
	rulebook *rulebook.Rulebook,
) ([]domain.Mutation, error)

var handlers = map[domain.AbilityEffectType]AbilityHandler{
	domain.EffectRevealStealth: revealStealth,
}

// Dispatch routes an asset's ability to the handler registered for its effect.
// Assets with no ability block, or an effect with no registered handler, fall
// through to confirmApplied (the not-yet-built effects).
func Dispatch(
	faction *domain.Faction,
	asset *domain.Asset,
	def *domain.AssetDefinition,
	collector action.Collector,
	roller domain.Roller,
	factionState *state.FactionState,
	rulebook *rulebook.Rulebook,
) ([]domain.Mutation, error) {
	if def.Ability == nil {
		return confirmApplied(faction, asset, def, collector, roller, factionState, rulebook)
	}
	handler, ok := handlers[def.Ability.Effect]
	if !ok {
		return confirmApplied(faction, asset, def, collector, roller, factionState, rulebook)
	}
	return handler(faction, asset, def, collector, roller, factionState, rulebook)
}

func confirmApplied(
	_ *domain.Faction,
	asset *domain.Asset,
	def *domain.AssetDefinition,
	collector action.Collector,
	_ domain.Roller,
	_ *state.FactionState,
	_ *rulebook.Rulebook,
) ([]domain.Mutation, error) {
	if _, err := collector.ConfirmAbilityApplied(asset, def); err != nil {
		return nil, err
	}
	return nil, nil
}

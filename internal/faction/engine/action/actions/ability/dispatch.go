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

var handlers = map[string]AbilityHandler{
	"C1-002": informers, // Informers

	// Structural stubs — real implementations land in F-014.
	"W1-002": confirmApplied, // Harvesters
	"W3-001": confirmApplied, // Postech Industry
	"W7-001": confirmApplied, // Pretech Manufactory
	"F5-002": confirmApplied, // Pretech Logistics
	"W6-001": confirmApplied, // Venture Capital
	"W6-003": confirmApplied, // Commodities Broker
	"C4-004": confirmApplied, // Seditionists
	"W5-001": confirmApplied, // Marketers
	"W4-002": confirmApplied, // Monopoly
}

// Dispatch routes an asset's ability to its registered handler.
// Falls through to confirmApplied for unregistered IDs (defensive; should not occur post-flag-corrections).
func Dispatch(
	faction *domain.Faction,
	asset *domain.Asset,
	def *domain.AssetDefinition,
	collector action.Collector,
	roller domain.Roller,
	factionState *state.FactionState,
	rulebook *rulebook.Rulebook,
) ([]domain.Mutation, error) {
	handler, ok := handlers[def.ID]
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

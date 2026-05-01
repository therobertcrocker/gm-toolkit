package digest

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func resolveFactionName(factionID string, factionState *state.FactionState) string {
	if factionState == nil {
		return factionID
	}
	if f, ok := factionState.Factions[factionID]; ok {
		return f.Name
	}
	return factionID
}

// resolveAssetName looks up the asset's display name via the rulebook definition.
// Falls back to assetID if the asset or definition is absent (e.g. already destroyed).
func resolveAssetName(factionID, assetID string, factionState *state.FactionState, rulebook *loader.Rulebook) string {
	if factionState == nil {
		return assetID
	}
	f, ok := factionState.Factions[factionID]
	if !ok {
		return assetID
	}
	for _, asset := range f.Assets {
		if asset.ID == assetID {
			if rulebook != nil {
				if def, ok := rulebook.Assets[asset.DefinitionID]; ok {
					return def.Name
				}
			}
			return assetID
		}
	}
	return assetID
}

func resolveGoalName(goalID string, rulebook *loader.Rulebook) string {
	if rulebook == nil {
		return goalID
	}
	if g, ok := rulebook.Goals[goalID]; ok {
		return g.Name
	}
	return goalID
}

func resolveBaseLocation(factionID, baseID string, factionState *state.FactionState) string {
	if factionState == nil {
		return ""
	}
	f, ok := factionState.Factions[factionID]
	if !ok {
		return ""
	}
	for _, base := range f.Bases {
		if base.ID == baseID {
			return base.Location
		}
	}
	return ""
}

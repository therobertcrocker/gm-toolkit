package actions

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

func indexFromState(factionState *state.FactionState) *world.Index {
	idx := &world.Index{
		AssetsByLocation: make(map[string][]*domain.Asset),
		BasesByLocation:  make(map[string][]*domain.Base),
		AssetsByHex:      make(map[spatial.RegionHex][]*domain.Asset),
	}
	for _, faction := range factionState.Factions {
		for _, asset := range faction.Assets {
			if !domain.IsInFlight(asset.Location) {
				idx.AssetsByLocation[asset.Location.WorldID] = append(idx.AssetsByLocation[asset.Location.WorldID], asset)
			}
			idx.AssetsByHex[asset.Location.RegionHex] = append(idx.AssetsByHex[asset.Location.RegionHex], asset)
		}
		for _, base := range faction.Bases {
			idx.BasesByLocation[base.Location.WorldID] = append(idx.BasesByLocation[base.Location.WorldID], base)
		}
	}
	return idx
}

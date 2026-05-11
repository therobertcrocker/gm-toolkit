package actions

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func indexFromState(factionState *state.FactionState) *world.Index {
	idx := &world.Index{
		AssetsByLocation: make(map[string][]*domain.Asset),
		BasesByLocation:  make(map[string][]*domain.Base),
	}
	for _, faction := range factionState.Factions {
		for _, asset := range faction.Assets {
			idx.AssetsByLocation[asset.Location] = append(idx.AssetsByLocation[asset.Location], asset)
		}
		for _, base := range faction.Bases {
			idx.BasesByLocation[base.Location] = append(idx.BasesByLocation[base.Location], base)
		}
	}
	return idx
}

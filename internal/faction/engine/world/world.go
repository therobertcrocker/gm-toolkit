package world

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type Index struct {
	AssetsByLocation map[string][]*domain.Asset
	BasesByLocation  map[string][]*domain.Base
}

type Engine struct {
	spatialMap spatial.SpatialMap
	Index      *Index
}

func New(dataDir string) (*Engine, error) {
	spatialMap, err := spatial.LoadHybrid(dataDir)
	if err != nil {
		return nil, err
	}
	return NewWithMap(spatialMap), nil
}

func NewWithMap(spatialMap spatial.SpatialMap) *Engine {
	return &Engine{spatialMap: spatialMap}
}

func (engine *Engine) RebuildIndex(factionState *state.FactionState) {
	index := &Index{
		AssetsByLocation: make(map[string][]*domain.Asset),
		BasesByLocation:  make(map[string][]*domain.Base),
	}
	for _, faction := range factionState.Factions {
		for _, asset := range faction.Assets {
			if _, ok := engine.spatialMap.Location(asset.Location); !ok {
				// TODO: surface skipped locations once a logging layer exists
				continue
			}
			index.AssetsByLocation[asset.Location] = append(index.AssetsByLocation[asset.Location], asset)
		}
		for _, base := range faction.Bases {
			if _, ok := engine.spatialMap.Location(base.Location); !ok {
				// TODO: surface skipped locations once a logging layer exists
				continue
			}
			index.BasesByLocation[base.Location] = append(index.BasesByLocation[base.Location], base)
		}
	}
	engine.Index = index
}

func (engine *Engine) Location(id string) (spatial.Location, bool) {
	// type assert to HexLocation
	if loc, ok := engine.spatialMap.Location(id); ok {
		if hexLoc, ok := loc.(spatial.HexLocation); ok {
			return hexLoc, true
		}
	}
	return nil, false
}

func (engine *Engine) Distance(fromID, toID string, crossingCost int) (int, error) {
	return engine.spatialMap.Distance(fromID, toID, crossingCost)
}

package world

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type HexRouter interface {
	spatial.SpatialMap
	Distance(from, to spatial.RegionHex, crossingCost int) (int, error)
	Path(from, to spatial.RegionHex, crossingCost int) ([]spatial.RegionHex, int, error)
}

var _ HexRouter = (*spatial.RegionMap)(nil)

type Index struct {
	AssetsByLocation map[string][]*domain.Asset
	BasesByLocation  map[string][]*domain.Base
	AssetsByHex      map[spatial.RegionHex][]*domain.Asset
}

type WorldEngine struct {
	spatialMap HexRouter
	Index      *Index
}

func New(dataDir string) (*WorldEngine, error) {
	spatialMap, err := spatial.LoadRegionMap(dataDir)
	if err != nil {
		return nil, err
	}
	return NewWithMap(spatialMap), nil
}

func NewWithMap(spatialMap HexRouter) *WorldEngine {
	return &WorldEngine{spatialMap: spatialMap}
}

func (engine *WorldEngine) RebuildIndex(factionState *state.FactionState) ([]string, error) {
	index := &Index{
		AssetsByLocation: make(map[string][]*domain.Asset),
		BasesByLocation:  make(map[string][]*domain.Base),
		AssetsByHex:      make(map[spatial.RegionHex][]*domain.Asset),
	}
	var skipped []string
	for _, faction := range factionState.Factions {
		for _, asset := range faction.Assets {
			index.AssetsByHex[asset.Location.RegionHex] = append(index.AssetsByHex[asset.Location.RegionHex], asset)

			if !domain.IsInFlight(asset.Location) {
				if _, ok := engine.spatialMap.Location(asset.Location.WorldID); !ok {
					skipped = append(skipped, asset.Location.WorldID)
					continue
				}
				index.AssetsByLocation[asset.Location.WorldID] = append(index.AssetsByLocation[asset.Location.WorldID], asset)
			}
		}
		for _, base := range faction.Bases {
			if _, ok := engine.spatialMap.Location(base.Location.WorldID); !ok {
				skipped = append(skipped, base.Location.WorldID)
				continue
			}
			index.BasesByLocation[base.Location.WorldID] = append(index.BasesByLocation[base.Location.WorldID], base)
		}
	}
	engine.Index = index
	return skipped, nil
}

func (engine *WorldEngine) Location(id string) (spatial.RegionLocation, bool) {
	loc, ok := engine.spatialMap.Location(id)
	if !ok {
		return nil, false
	}
	hexLoc, ok := loc.(spatial.RegionLocation)
	return hexLoc, ok
}

func (engine *WorldEngine) Distance(from, to spatial.RegionHex, crossingCost int) (int, error) {
	return engine.spatialMap.Distance(from, to, crossingCost)
}

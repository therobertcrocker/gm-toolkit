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

var _ world.HexRouter = (*testHexRouter)(nil)

type testHexRouter struct {
	worlds map[string]spatial.RegionHex
}

func (r *testHexRouter) Location(id string) (spatial.Location, bool) {
	hex, ok := r.worlds[id]
	if !ok {
		return nil, false
	}
	return &testWorldLoc{id: id, hex: hex}, true
}

func (r *testHexRouter) Distance(_, _ spatial.RegionHex, _ int) (int, error) {
	return 1, nil
}

func (r *testHexRouter) DistancesFrom(_ spatial.RegionHex, _ int) (map[spatial.RegionHex]int, error) {
	distances := make(map[spatial.RegionHex]int, len(r.worlds))
	for _, hex := range r.worlds {
		distances[hex] = 1
	}
	return distances, nil
}

func (r *testHexRouter) Path(from, to spatial.RegionHex, _ int) ([]spatial.RegionHex, int, error) {
	return []spatial.RegionHex{from, to}, 1, nil
}

func (r *testHexRouter) AllWorlds() []spatial.Location {
	worlds := make([]spatial.Location, 0, len(r.worlds))
	for id, hex := range r.worlds {
		worlds = append(worlds, &testWorldLoc{id: id, hex: hex})
	}
	return worlds
}

var _ spatial.RegionLocation = (*testWorldLoc)(nil)

type testWorldLoc struct {
	id  string
	hex spatial.RegionHex
}

func (l *testWorldLoc) ID() string                   { return l.id }
func (l *testWorldLoc) Name() string                 { return l.id }
func (l *testWorldLoc) TechLevel() int               { return 5 }
func (l *testWorldLoc) Population() int              { return 0 }
func (l *testWorldLoc) Coords() (q, r int)           { return l.hex.Coord.Q, l.hex.Coord.R }
func (l *testWorldLoc) RegionID() string             { return l.hex.RegionID }
func (l *testWorldLoc) RegionHex() spatial.RegionHex { return l.hex }

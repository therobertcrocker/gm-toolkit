package world

import (
	"fmt"
	"log/slog"
	"slices"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type HexRouter interface {
	spatial.SpatialMap
	Distance(from, to spatial.RegionHex, crossingCost int) (int, error)
	DistancesFrom(from spatial.RegionHex, crossingCost int) (map[spatial.RegionHex]int, error)
	Path(from, to spatial.RegionHex, crossingCost int) ([]spatial.RegionHex, int, error)
	AllWorlds() []spatial.Location
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
	log        *slog.Logger
}

func New(dataDir string, log *slog.Logger) (*WorldEngine, error) {
	spatialMap, err := spatial.LoadRegionMap(dataDir)
	if err != nil {
		return nil, err
	}
	return NewWithMap(spatialMap, log), nil
}

func NewWithMap(spatialMap HexRouter, log *slog.Logger) *WorldEngine {
	return &WorldEngine{spatialMap: spatialMap, log: log}
}

func (engine *WorldEngine) RebuildIndex(factionState *state.FactionState, log *slog.Logger) ([]string, error) {
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

// WorldsWithin returns the IDs of every world within maxHex of the given hex,
// name-sorted. The origin hex itself is included only when includeOrigin is set
// (other-scope relocate may target the home world; self-scope must move off it).
// Distance is raw hex count.
func (engine *WorldEngine) WorldsWithin(from spatial.RegionHex, maxHex int, includeOrigin bool) ([]string, error) {
	// deliberate: crossingCost 0 — "within N hexes" is raw hex distance, not the
	// drift-priced journey ChangeHomeworld uses (Effort 3 Plan, Decision 5).
	distances, err := engine.spatialMap.DistancesFrom(from, 0)
	if err != nil {
		return nil, err
	}
	return engine.worldsMatching(func(hexLoc spatial.RegionLocation) bool {
		dist, reachable := distances[hexLoc.RegionHex()]
		return reachable && dist <= maxHex && (includeOrigin || dist > 0)
	}), nil
}

// WorldsInRegion returns the IDs of every world sharing the origin hex's region,
// name-sorted. The origin hex itself is included only when includeOrigin is set.
// Region membership is a pure equality check — no distance math, since every hex
// already carries its RegionID.
func (engine *WorldEngine) WorldsInRegion(from spatial.RegionHex, includeOrigin bool) ([]string, error) {
	return engine.worldsMatching(func(hexLoc spatial.RegionLocation) bool {
		hex := hexLoc.RegionHex()
		return hex.RegionID == from.RegionID && (includeOrigin || hex != from)
	}), nil
}

// worldsMatching returns the name-sorted IDs of every world whose RegionLocation
// satisfies pred. WorldsWithin and WorldsInRegion share this scaffolding; only the
// per-candidate predicate (hex-distance vs region equality) differs.
func (engine *WorldEngine) worldsMatching(pred func(spatial.RegionLocation) bool) []string {
	var within []string
	for _, loc := range engine.spatialMap.AllWorlds() {
		hexLoc, ok := loc.(spatial.RegionLocation)
		if !ok {
			continue
		}
		if pred(hexLoc) {
			within = append(within, hexLoc.ID())
		}
	}
	sort.Strings(within)
	return within
}

// BuildRelocateMutations validates that destination is one of the caller-supplied
// legalWorlds — the range-filtered candidates from WorldsWithin (hex radius) or
// WorldsInRegion (same region) — and emits the AssetMoved for the moved asset, plus a
// CoinDelta when coinCost > 0. Instant, distinct from the drift-order movement phase.
// Range geometry lives in the enumerators; this primitive is range-mode-agnostic.
func (engine *WorldEngine) BuildRelocateMutations(
	faction *domain.Faction,
	moved *domain.Asset,
	destination domain.Location,
	legalWorlds []string,
	coinCost int,
) ([]domain.Mutation, error) {
	if !slices.Contains(legalWorlds, destination.WorldID) {
		return nil, fmt.Errorf("relocate: destination %q is not a legal target", destination.WorldID)
	}
	mutations := []domain.Mutation{domain.AssetMoved{
		FactionID:         faction.ID,
		AssetID:           moved.ID,
		FromLocation:      moved.Location,
		ToLocation:        destination,
		Cause:             "ability",
		CausedByFactionID: faction.ID,
	}}
	if coinCost > 0 {
		mutations = append(mutations, domain.CoinDelta{
			FactionID:         faction.ID,
			Delta:             -coinCost,
			Cause:             "ability",
			CausedByFactionID: faction.ID,
		})
	}
	return mutations, nil
}

package ability

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
	"go.uber.org/mock/gomock"
)

// relocRouter places worlds on a Q-line; hex distance = |ΔQ|. The full relocate
// suite lives in engine/relocate; this copy serves the wiring and dispatch tests.
type relocRouter struct{ worlds map[string]spatial.RegionHex }

func (r *relocRouter) Location(id string) (spatial.Location, bool) {
	hex, ok := r.worlds[id]
	if !ok {
		return nil, false
	}
	return &relocLoc{id: id, hex: hex}, true
}

func (r *relocRouter) Distance(from, to spatial.RegionHex, _ int) (int, error) {
	dist := from.Coord.Q - to.Coord.Q
	if dist < 0 {
		dist = -dist
	}
	return dist, nil
}

func (r *relocRouter) DistancesFrom(from spatial.RegionHex, crossingCost int) (map[spatial.RegionHex]int, error) {
	distances := make(map[spatial.RegionHex]int, len(r.worlds))
	for _, hex := range r.worlds {
		dist, err := r.Distance(from, hex, crossingCost)
		if err != nil {
			return nil, err
		}
		distances[hex] = dist
	}
	return distances, nil
}

func (r *relocRouter) Path(from, to spatial.RegionHex, _ int) ([]spatial.RegionHex, int, error) {
	return []spatial.RegionHex{from, to}, 1, nil
}

func (r *relocRouter) AllWorlds() []spatial.Location {
	worlds := make([]spatial.Location, 0, len(r.worlds))
	for id, hex := range r.worlds {
		worlds = append(worlds, &relocLoc{id: id, hex: hex})
	}
	return worlds
}

var _ world.HexRouter = (*relocRouter)(nil)

type relocLoc struct {
	id  string
	hex spatial.RegionHex
}

func (l *relocLoc) ID() string                   { return l.id }
func (l *relocLoc) Name() string                 { return l.id }
func (l *relocLoc) TechLevel() int               { return 5 }
func (l *relocLoc) Population() int              { return 0 }
func (l *relocLoc) Coords() (int, int)           { return l.hex.Coord.Q, l.hex.Coord.R }
func (l *relocLoc) RegionID() string             { return l.hex.RegionID }
func (l *relocLoc) RegionHex() spatial.RegionHex { return l.hex }

func relocHex(region string, q int) spatial.RegionHex {
	return spatial.RegionHex{RegionID: region, Coord: spatial.HexCoord{Q: q, R: 0}}
}

// relocEngine seeds region r1 with origin/near/mid/far (Q 0/1/2/5) and r2 with
// offregion (Q 10), parked far out so it never falls in hex range.
func relocEngine() *world.WorldEngine {
	return world.NewWithMap(&relocRouter{worlds: map[string]spatial.RegionHex{
		"origin":    relocHex("r1", 0),
		"near":      relocHex("r1", 1),
		"mid":       relocHex("r1", 2),
		"far":       relocHex("r1", 5),
		"offregion": relocHex("r2", 10),
	}}, slog.New(slog.DiscardHandler))
}

func relocRulebook() *rulebook.Rulebook {
	return &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"SF": {ID: "SF", Type: domain.TypeSpecialForces},
		"MU": {ID: "MU", Type: domain.TypeMilitaryUnit},
	}}
}

func TestRelocateAssets_Self_Hex(t *testing.T) {
	ctrl := gomock.NewController(t)
	acting := &domain.Asset{ID: "seductress", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "origin", RegionHex: relocHex("r1", 0)}}
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"seductress": acting}}

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]action.RelocateOrder{{Asset: acting, DestinationWorld: "near"}}, nil)

	mutations, err := relocateAssets(StepContext{
		Faction:     faction,
		Asset:       acting,
		Step:        domain.AbilityStep{Effect: domain.EffectRelocate, Relocate: &domain.RelocateParams{Scope: domain.RelocateSelf, Range: domain.RangeHexes, MaxHex: 1}},
		Collector:   collector,
		WorldEngine: relocEngine(),
		Rulebook:    relocRulebook(),
	})
	if err != nil {
		t.Fatalf("relocateAssets: %v", err)
	}
	if len(mutations) != 1 {
		t.Fatalf("len(mutations) = %d, want 1 (AssetMoved, no CoinDelta)", len(mutations))
	}
	moved, ok := mutations[0].(domain.AssetMoved)
	if !ok {
		t.Fatalf("mutations[0] type = %T, want AssetMoved", mutations[0])
	}
	if moved.AssetID != "seductress" || moved.ToLocation.WorldID != "near" {
		t.Errorf("AssetMoved = %+v, want AssetID=seductress ToLocation=near", moved)
	}
}

func TestRelocateAssets_NilParams(t *testing.T) {
	acting := &domain.Asset{ID: "seductress", OwnerID: "f1"}
	_, err := relocateAssets(StepContext{
		Faction: &domain.Faction{ID: "f1"},
		Asset:   acting,
		Step:    domain.AbilityStep{Effect: domain.EffectRelocate},
	})
	if err == nil || !strings.Contains(err.Error(), "no relocate params") {
		t.Fatalf("err = %v, want the no-relocate-params error", err)
	}
}

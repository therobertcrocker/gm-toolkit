package world_test

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

// relocateRouter is a HexRouter whose worlds sit on a line: a world's hex distance
// from another equals the difference of their Coord.Q, ignoring region. Region
// membership is carried separately so the two enumerators can be tested independently.
type relocateRouter struct {
	worlds map[string]spatial.RegionHex
}

func (r *relocateRouter) Location(id string) (spatial.Location, bool) {
	hex, ok := r.worlds[id]
	if !ok {
		return nil, false
	}
	return &relocateLoc{id: id, hex: hex}, true
}

func (r *relocateRouter) Distance(from, to spatial.RegionHex, _ int) (int, error) {
	dist := from.Coord.Q - to.Coord.Q
	if dist < 0 {
		dist = -dist
	}
	return dist, nil
}

func (r *relocateRouter) DistancesFrom(from spatial.RegionHex, crossingCost int) (map[spatial.RegionHex]int, error) {
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

func (r *relocateRouter) Path(from, to spatial.RegionHex, _ int) ([]spatial.RegionHex, int, error) {
	return []spatial.RegionHex{from, to}, 1, nil
}

func (r *relocateRouter) AllWorlds() []spatial.Location {
	worlds := make([]spatial.Location, 0, len(r.worlds))
	for id, hex := range r.worlds {
		worlds = append(worlds, &relocateLoc{id: id, hex: hex})
	}
	return worlds
}

var _ world.HexRouter = (*relocateRouter)(nil)

type relocateLoc struct {
	id  string
	hex spatial.RegionHex
}

func (l *relocateLoc) ID() string                   { return l.id }
func (l *relocateLoc) Name() string                 { return l.id }
func (l *relocateLoc) TechLevel() int               { return 5 }
func (l *relocateLoc) Population() int              { return 0 }
func (l *relocateLoc) Coords() (int, int)           { return l.hex.Coord.Q, l.hex.Coord.R }
func (l *relocateLoc) RegionID() string             { return l.hex.RegionID }
func (l *relocateLoc) RegionHex() spatial.RegionHex { return l.hex }

func hexAt(region string, q int) spatial.RegionHex {
	return spatial.RegionHex{RegionID: region, Coord: spatial.HexCoord{Q: q, R: 0}}
}

// relocateTestRouter places four worlds in region r1 (origin=0, near=1, mid=2, far=5)
// and one in r2 (offregion) parked far out in Q so it never falls in hex range.
func relocateTestRouter() *relocateRouter {
	return &relocateRouter{worlds: map[string]spatial.RegionHex{
		"origin":    hexAt("r1", 0),
		"near":      hexAt("r1", 1),
		"mid":       hexAt("r1", 2),
		"far":       hexAt("r1", 5),
		"offregion": hexAt("r2", 10),
	}}
}

func newRelocateEngine() *world.WorldEngine {
	return world.NewWithMap(relocateTestRouter(), slog.New(slog.DiscardHandler))
}

func TestBuildRelocateMutations_LegalNoCoin(t *testing.T) {
	engine := newRelocateEngine()
	faction := &domain.Faction{ID: "f1"}
	moved := &domain.Asset{ID: "asset-1", Location: domain.Location{WorldID: "origin", RegionHex: hexAt("r1", 0)}}
	dest := domain.Location{WorldID: "mid", RegionHex: hexAt("r1", 2)}

	mutations, err := engine.BuildRelocateMutations(faction, moved, dest, []string{"mid", "near"}, 0)
	if err != nil {
		t.Fatalf("BuildRelocateMutations: %v", err)
	}
	if len(mutations) != 1 {
		t.Fatalf("len(mutations) = %d, want 1 (no CoinDelta when coinCost == 0)", len(mutations))
	}
	moveMut, ok := mutations[0].(domain.AssetMoved)
	if !ok {
		t.Fatalf("mutations[0] type = %T, want AssetMoved", mutations[0])
	}
	if moveMut.AssetID != "asset-1" || moveMut.ToLocation.WorldID != "mid" {
		t.Errorf("AssetMoved = %+v, want AssetID=asset-1 ToLocation=mid", moveMut)
	}
}

func TestBuildRelocateMutations_LegalWithCoin(t *testing.T) {
	engine := newRelocateEngine()
	faction := &domain.Faction{ID: "f1"}
	moved := &domain.Asset{ID: "asset-1", Location: domain.Location{WorldID: "origin", RegionHex: hexAt("r1", 0)}}
	dest := domain.Location{WorldID: "near", RegionHex: hexAt("r1", 1)}

	mutations, err := engine.BuildRelocateMutations(faction, moved, dest, []string{"near"}, 1)
	if err != nil {
		t.Fatalf("BuildRelocateMutations: %v", err)
	}
	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2 (AssetMoved + CoinDelta)", len(mutations))
	}
	coinMut, ok := mutations[1].(domain.CoinDelta)
	if !ok {
		t.Fatalf("mutations[1] type = %T, want CoinDelta", mutations[1])
	}
	if coinMut.Delta != -1 {
		t.Errorf("CoinDelta.Delta = %d, want -1", coinMut.Delta)
	}
}

func TestBuildRelocateMutations_DestinationNotLegal(t *testing.T) {
	engine := newRelocateEngine()
	faction := &domain.Faction{ID: "f1"}
	moved := &domain.Asset{ID: "asset-1", Location: domain.Location{WorldID: "origin", RegionHex: hexAt("r1", 0)}}
	dest := domain.Location{WorldID: "far", RegionHex: hexAt("r1", 5)}

	// "far" is omitted from the legal set — the collector cannot inject it.
	mutations, err := engine.BuildRelocateMutations(faction, moved, dest, []string{"mid", "near"}, 0)
	if err == nil {
		t.Error("BuildRelocateMutations: expected error for destination outside legal set, got nil")
	}
	if mutations != nil {
		t.Errorf("mutations = %v, want nil on error", mutations)
	}
}

func TestWorldsWithin(t *testing.T) {
	engine := newRelocateEngine()

	// origin at r1 hex 0; in-region worlds: near=1, mid=2, far=5; offregion at r2 hex 10.
	// maxHex=3 keeps near and mid; origin (dist 0), far (dist 5), and offregion (dist 10)
	// are all excluded. Distance ignores region, so offregion's exclusion is purely range.
	within, err := engine.WorldsWithin(hexAt("r1", 0), 3, false)
	if err != nil {
		t.Fatalf("WorldsWithin: %v", err)
	}
	want := []string{"mid", "near"}
	if !reflect.DeepEqual(within, want) {
		t.Errorf("WorldsWithin = %v, want %v (name-sorted, origin and out-of-range excluded)", within, want)
	}
}

func TestWorldsWithin_IncludeOrigin(t *testing.T) {
	engine := newRelocateEngine()

	// Same fixture/range as TestWorldsWithin, but includeOrigin=true (other-scope): the
	// origin world joins near and mid; far and offregion are still out of range.
	within, err := engine.WorldsWithin(hexAt("r1", 0), 3, true)
	if err != nil {
		t.Fatalf("WorldsWithin: %v", err)
	}
	want := []string{"mid", "near", "origin"}
	if !reflect.DeepEqual(within, want) {
		t.Errorf("WorldsWithin = %v, want %v (name-sorted, origin included, out-of-range excluded)", within, want)
	}
}

func TestWorldsInRegion(t *testing.T) {
	engine := newRelocateEngine()

	// origin at r1 hex 0; same-region worlds near/mid/far are all included regardless of
	// hex distance; origin itself is excluded; offregion (r2) is filtered out by region.
	within, err := engine.WorldsInRegion(hexAt("r1", 0), false)
	if err != nil {
		t.Fatalf("WorldsInRegion: %v", err)
	}
	want := []string{"far", "mid", "near"}
	if !reflect.DeepEqual(within, want) {
		t.Errorf("WorldsInRegion = %v, want %v (name-sorted, origin and other regions excluded)", within, want)
	}
}

func TestWorldsInRegion_IncludeOrigin(t *testing.T) {
	engine := newRelocateEngine()

	// includeOrigin=true (other-scope): every r1 world including origin is returned;
	// offregion (r2) is still filtered out by region.
	within, err := engine.WorldsInRegion(hexAt("r1", 0), true)
	if err != nil {
		t.Fatalf("WorldsInRegion: %v", err)
	}
	want := []string{"far", "mid", "near", "origin"}
	if !reflect.DeepEqual(within, want) {
		t.Errorf("WorldsInRegion = %v, want %v (name-sorted, origin included, other regions excluded)", within, want)
	}
}

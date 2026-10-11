package ability

import (
	"log/slog"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
	"go.uber.org/mock/gomock"
)

// relocRouter places worlds on a Q-line; hex distance = |ΔQ| ignoring region, and
// each world carries a RegionID so the hex and region enumerators are exercised
// independently.
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

// relocRulebook types each definition ID for eligibleRelocateAssets' AssetType filter.
func relocRulebook() *rulebook.Rulebook {
	return &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"SF": {ID: "SF", Type: domain.TypeSpecialForces},
		"MU": {ID: "MU", Type: domain.TypeMilitaryUnit},
	}}
}

func TestRelocate_Self_Hex(t *testing.T) {
	ctrl := gomock.NewController(t)
	acting := &domain.Asset{ID: "seductress", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "origin", RegionHex: relocHex("r1", 0)}}
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"seductress": acting}}

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]action.RelocateOrder{{Asset: acting, DestinationWorld: "near"}}, nil)

	mutations, err := relocate(StepContext{
		Faction:     faction,
		Asset:       acting,
		Step:        domain.AbilityStep{Effect: domain.EffectRelocate, Relocate: &domain.RelocateParams{Scope: domain.RelocateSelf, Range: domain.RangeHexes, MaxHex: 1}},
		Collector:   collector,
		WorldEngine: relocEngine(),
		Rulebook:    relocRulebook(),
	})
	if err != nil {
		t.Fatalf("relocate: %v", err)
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

func TestRelocate_OtherSingle_Hex(t *testing.T) {
	ctrl := gomock.NewController(t)
	shipping := &domain.Asset{ID: "cs", DefinitionID: "MU", OwnerID: "f1", Location: domain.Location{WorldID: "origin", RegionHex: relocHex("r1", 0)}}
	sfInRange := &domain.Asset{ID: "sf-near", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "near", RegionHex: relocHex("r1", 1)}}
	wrongType := &domain.Asset{ID: "mu-near", DefinitionID: "MU", OwnerID: "f1", Location: domain.Location{WorldID: "near", RegionHex: relocHex("r1", 1)}}
	sfOutOfRange := &domain.Asset{ID: "sf-far", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "far", RegionHex: relocHex("r1", 5)}}
	faction := &domain.Faction{ID: "f1", Coin: 5, Assets: map[string]*domain.Asset{
		"cs": shipping, "sf-near": sfInRange, "mu-near": wrongType, "sf-far": sfOutOfRange,
	}}
	step := domain.AbilityStep{Effect: domain.EffectRelocate, Relocate: &domain.RelocateParams{
		Scope: domain.RelocateOther, Range: domain.RangeHexes, MaxHex: 3, CoinCost: 1, AssetType: domain.TypeSpecialForces, MaxAssets: 1,
	}}

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]action.RelocateOrder{{Asset: sfInRange, DestinationWorld: "mid"}}, nil)

	mutations, err := relocate(StepContext{
		Faction: faction, Asset: shipping, Step: step,
		Collector: collector, WorldEngine: relocEngine(), Rulebook: relocRulebook(),
	})
	if err != nil {
		t.Fatalf("relocate: %v", err)
	}
	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2 (AssetMoved + CoinDelta)", len(mutations))
	}
	moved, ok := mutations[0].(domain.AssetMoved)
	if !ok || moved.AssetID != "sf-near" || moved.ToLocation.WorldID != "mid" {
		t.Fatalf("mutations[0] = %+v, want AssetMoved sf-near -> mid", mutations[0])
	}
	coin, ok := mutations[1].(domain.CoinDelta)
	if !ok || coin.Delta != -1 {
		t.Errorf("mutations[1] = %+v, want CoinDelta -1", mutations[1])
	}
}

func TestRelocate_OtherSingle_InsufficientCoin(t *testing.T) {
	ctrl := gomock.NewController(t)
	shipping := &domain.Asset{ID: "cs", DefinitionID: "MU", OwnerID: "f1", Location: domain.Location{WorldID: "origin", RegionHex: relocHex("r1", 0)}}
	sfInRange := &domain.Asset{ID: "sf-near", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "near", RegionHex: relocHex("r1", 1)}}
	faction := &domain.Faction{ID: "f1", Coin: 0, Assets: map[string]*domain.Asset{"cs": shipping, "sf-near": sfInRange}}
	step := domain.AbilityStep{Effect: domain.EffectRelocate, Relocate: &domain.RelocateParams{
		Scope: domain.RelocateOther, Range: domain.RangeHexes, MaxHex: 3, CoinCost: 1, AssetType: domain.TypeSpecialForces, MaxAssets: 1,
	}}

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]action.RelocateOrder{{Asset: sfInRange, DestinationWorld: "mid"}}, nil)

	mutations, err := relocate(StepContext{
		Faction: faction, Asset: shipping, Step: step,
		Collector: collector, WorldEngine: relocEngine(), Rulebook: relocRulebook(),
	})
	if err == nil {
		t.Fatal("relocate: expected insufficient-Coin error, got nil")
	}
	if mutations != nil {
		t.Errorf("mutations = %v, want nil on affordability failure", mutations)
	}
}

func TestRelocate_ExceedsMaxAssets(t *testing.T) {
	ctrl := gomock.NewController(t)
	shipping := &domain.Asset{ID: "cs", DefinitionID: "MU", OwnerID: "f1", Location: domain.Location{WorldID: "origin", RegionHex: relocHex("r1", 0)}}
	sf1 := &domain.Asset{ID: "sf-1", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "near", RegionHex: relocHex("r1", 1)}}
	sf2 := &domain.Asset{ID: "sf-2", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "mid", RegionHex: relocHex("r1", 2)}}
	faction := &domain.Faction{ID: "f1", Coin: 5, Assets: map[string]*domain.Asset{"cs": shipping, "sf-1": sf1, "sf-2": sf2}}
	step := domain.AbilityStep{Effect: domain.EffectRelocate, Relocate: &domain.RelocateParams{
		Scope: domain.RelocateOther, Range: domain.RangeHexes, MaxHex: 3, CoinCost: 1, AssetType: domain.TypeSpecialForces, MaxAssets: 1,
	}}

	// Collector over-selects (two orders) against a MaxAssets=1 cap; the handler guard rejects it.
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]action.RelocateOrder{{Asset: sf1, DestinationWorld: "mid"}, {Asset: sf2, DestinationWorld: "near"}}, nil)

	mutations, err := relocate(StepContext{
		Faction: faction, Asset: shipping, Step: step,
		Collector: collector, WorldEngine: relocEngine(), Rulebook: relocRulebook(),
	})
	if err == nil {
		t.Fatal("relocate: expected MaxAssets-exceeded error, got nil")
	}
	if mutations != nil {
		t.Errorf("mutations = %v, want nil when the cap is exceeded", mutations)
	}
}

func TestRelocate_OtherMulti_Region(t *testing.T) {
	ctrl := gomock.NewController(t)
	net := &domain.Asset{ID: "net", DefinitionID: "MU", OwnerID: "f1", Location: domain.Location{WorldID: "origin", RegionHex: relocHex("r1", 0)}}
	sf1 := &domain.Asset{ID: "sf-1", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "near", RegionHex: relocHex("r1", 1)}}
	sf2 := &domain.Asset{ID: "sf-2", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "far", RegionHex: relocHex("r1", 5)}}
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"net": net, "sf-1": sf1, "sf-2": sf2}}
	step := domain.AbilityStep{Effect: domain.EffectRelocate, Relocate: &domain.RelocateParams{
		Scope: domain.RelocateOther, Range: domain.RangeRegion, AssetType: domain.TypeSpecialForces, MaxAssets: 0,
	}}

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]action.RelocateOrder{{Asset: sf1, DestinationWorld: "mid"}, {Asset: sf2, DestinationWorld: "near"}}, nil)

	mutations, err := relocate(StepContext{
		Faction: faction, Asset: net, Step: step,
		Collector: collector, WorldEngine: relocEngine(), Rulebook: relocRulebook(),
	})
	if err != nil {
		t.Fatalf("relocate: %v", err)
	}
	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2 (two AssetMoved, no CoinDelta)", len(mutations))
	}
	for i, m := range mutations {
		if _, ok := m.(domain.AssetMoved); !ok {
			t.Errorf("mutations[%d] type = %T, want AssetMoved", i, m)
		}
	}
}

// TestEligibleRelocateAssets_Filtering proves the three exclusions: wrong AssetType,
// a world outside the range-filtered candidate set (hex out-of-range or out-of-region
// both manifest as "not in candidateWorlds"), and the acting asset itself.
func TestEligibleRelocateAssets_Filtering(t *testing.T) {
	acting := &domain.Asset{ID: "cs", DefinitionID: "MU", OwnerID: "f1", Location: domain.Location{WorldID: "origin"}}
	sfNear := &domain.Asset{ID: "sf-near", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "near"}}
	muNear := &domain.Asset{ID: "mu-near", DefinitionID: "MU", OwnerID: "f1", Location: domain.Location{WorldID: "near"}}
	sfFar := &domain.Asset{ID: "sf-far", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "far"}}
	sfOff := &domain.Asset{ID: "sf-off", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "offregion"}}
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{
		"cs": acting, "sf-near": sfNear, "mu-near": muNear, "sf-far": sfFar, "sf-off": sfOff,
	}}
	ctx := StepContext{Faction: faction, Asset: acting, Rulebook: relocRulebook()}
	params := &domain.RelocateParams{Scope: domain.RelocateOther, AssetType: domain.TypeSpecialForces}

	eligible := eligibleRelocateAssets(ctx, params, []string{"mid", "near"})

	if len(eligible) != 1 {
		t.Fatalf("eligible = %v, want [sf-near] (wrong-type, out-of-range, off-region, and self all excluded)", assetIDs(eligible))
	}
	if eligible[0].ID != "sf-near" {
		t.Errorf("eligible[0].ID = %q, want sf-near", eligible[0].ID)
	}
}

func assetIDs(assets []*domain.Asset) []string {
	ids := make([]string, len(assets))
	for i, asset := range assets {
		ids[i] = asset.ID
	}
	return ids
}

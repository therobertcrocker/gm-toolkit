package world_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type stubMap struct{}

func (s *stubMap) Location(_ string) (spatial.Location, bool)          { return nil, false }
func (s *stubMap) Distance(_, _ spatial.RegionHex, _ int) (int, error) { return 1, nil }
func (s *stubMap) Path(_, _ spatial.RegionHex, _ int) ([]spatial.RegionHex, int, error) {
	return nil, 0, nil
}

type stubRegionLoc struct{ rh spatial.RegionHex }

func (s stubRegionLoc) ID() string                   { return "dest" }
func (s stubRegionLoc) Name() string                 { return "dest" }
func (s stubRegionLoc) TechLevel() int               { return 0 }
func (s stubRegionLoc) Population() int              { return 0 }
func (s stubRegionLoc) Coords() (int, int)           { return 0, 0 }
func (s stubRegionLoc) RegionID() string             { return s.rh.RegionID }
func (s stubRegionLoc) RegionHex() spatial.RegionHex { return s.rh }

type stubMapWithDest struct{ stubMap }

func (s *stubMapWithDest) Location(id string) (spatial.Location, bool) {
	if id == "dest" {
		return stubRegionLoc{rh: spatial.RegionHex{RegionID: "r1", Coord: spatial.HexCoord{Q: 3, R: 0}}}, true
	}
	return nil, false
}

func testRulebook(speed int) *rulebook.Rulebook {
	return &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			"test-def": {ID: "test-def", Speed: speed},
		},
		DriftCosts: []int{3, 2, 1},
	}
}

func makePath(n int) []spatial.RegionHex {
	path := make([]spatial.RegionHex, n)
	for i := range path {
		path[i] = spatial.RegionHex{RegionID: "stub", Coord: spatial.HexCoord{Q: i, R: 0}}
	}
	return path
}

func factionWithOrder(stepIdx int, path []spatial.RegionHex, dest domain.Location) *domain.Faction {
	return &domain.Faction{
		ID: "alpha",
		Assets: map[string]*domain.Asset{
			"asset-1": {
				ID:           "asset-1",
				DefinitionID: "test-def",
				OwnerID:      "alpha",
				CurrentOrder: &domain.MovementOrder{
					AssetID:     "asset-1",
					StepIdx:     stepIdx,
					Path:        path,
					Destination: dest,
				},
			},
		},
	}
}

type stubSingleWorld struct {
	stubMap
	id  string
	hex spatial.RegionHex
}

func (s *stubSingleWorld) Location(id string) (spatial.Location, bool) {
	if id == s.id {
		return stubRegionLoc{rh: s.hex}, true
	}
	return nil, false
}

func containsAsset(assets []*domain.Asset, id string) bool {
	for _, a := range assets {
		if a.ID == id {
			return true
		}
	}
	return false
}

func assetIDs(assets []*domain.Asset) []string {
	ids := make([]string, len(assets))
	for i, a := range assets {
		ids[i] = a.ID
	}
	return ids
}

func TestRebuildIndex_HexIndex(t *testing.T) {
	worldAHex := spatial.RegionHex{RegionID: "r1", Coord: spatial.HexCoord{Q: 2, R: 1}}
	midFlightHex := spatial.RegionHex{RegionID: "r1", Coord: spatial.HexCoord{Q: 5, R: 3}}

	we := world.NewWithMap(&stubSingleWorld{id: "a", hex: worldAHex})

	assetX := &domain.Asset{ID: "x", Location: domain.Location{WorldID: "a", RegionHex: worldAHex}}
	assetY := &domain.Asset{ID: "y", Location: domain.Location{WorldID: "", RegionHex: midFlightHex}}
	assetZ := &domain.Asset{ID: "z", Location: domain.Location{WorldID: "a", RegionHex: worldAHex}}

	factionState := &state.FactionState{
		Factions: map[string]*domain.Faction{
			"alpha": {
				ID:     "alpha",
				Assets: map[string]*domain.Asset{"x": assetX, "y": assetY, "z": assetZ},
			},
		},
	}

	skipped, err := we.RebuildIndex(factionState)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(skipped) != 0 {
		t.Errorf("expected no skipped worlds, got %v", skipped)
	}

	idx := we.Index

	locA := idx.AssetsByLocation["a"]
	if !containsAsset(locA, "x") || !containsAsset(locA, "z") {
		t.Errorf("AssetsByLocation[a] missing x or z: %v", assetIDs(locA))
	}
	if containsAsset(locA, "y") {
		t.Errorf("AssetsByLocation[a] should not contain mid-flight asset y")
	}

	hexA := idx.AssetsByHex[worldAHex]
	if !containsAsset(hexA, "x") || !containsAsset(hexA, "z") {
		t.Errorf("AssetsByHex[worldAHex] missing x or z: %v", assetIDs(hexA))
	}

	hexMid := idx.AssetsByHex[midFlightHex]
	if !containsAsset(hexMid, "y") {
		t.Errorf("AssetsByHex[midFlightHex] missing y: %v", assetIDs(hexMid))
	}
}

func TestTickMovementOrders_NoOrders(t *testing.T) {
	we := world.NewWithMap(&stubMap{})
	faction := &domain.Faction{
		ID: "alpha",
		Assets: map[string]*domain.Asset{
			"asset-1": {
				ID:           "asset-1",
				DefinitionID: "test-def",
				OwnerID:      "alpha",
			},
		},
	}

	mutations, err := we.TickMovementOrders(faction, testRulebook(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mutations) != 0 {
		t.Errorf("expected no mutations, got %d", len(mutations))
	}
}

func TestTickMovementOrders_MidFlight(t *testing.T) {
	we := world.NewWithMap(&stubMap{})
	dest := domain.Location{WorldID: "beta"}
	faction := factionWithOrder(0, makePath(5), dest)

	mutations, err := we.TickMovementOrders(faction, testRulebook(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mutations) != 1 {
		t.Fatalf("expected 1 mutation, got %d", len(mutations))
	}
	progressed, ok := mutations[0].(domain.MovementOrderProgressed)
	if !ok {
		t.Fatalf("expected MovementOrderProgressed, got %T", mutations[0])
	}
	if progressed.NewStepIdx != 1 {
		t.Errorf("expected NewStepIdx=1, got %d", progressed.NewStepIdx)
	}
}

func TestTickMovementOrders_Completion(t *testing.T) {
	we := world.NewWithMap(&stubMap{})
	dest := domain.Location{WorldID: "beta"}
	faction := factionWithOrder(3, makePath(5), dest)

	mutations, err := we.TickMovementOrders(faction, testRulebook(2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mutations) != 1 {
		t.Fatalf("expected 1 mutation, got %d", len(mutations))
	}
	completed, ok := mutations[0].(domain.MovementOrderCompleted)
	if !ok {
		t.Fatalf("expected MovementOrderCompleted, got %T", mutations[0])
	}
	if completed.FinalLocation != dest {
		t.Errorf("expected FinalLocation=%v, got %v", dest, completed.FinalLocation)
	}
}

func TestBuildMovementMutations_CargoManifestPassthrough(t *testing.T) {
	we := world.NewWithMap(&stubMapWithDest{})
	rb := testRulebook(1)
	faction := &domain.Faction{
		ID: "alpha",
		Assets: map[string]*domain.Asset{
			"transport-1": {
				ID:           "transport-1",
				DefinitionID: "test-def",
				OwnerID:      "alpha",
				Location: domain.Location{
					WorldID:   "origin",
					RegionHex: spatial.RegionHex{RegionID: "r0", Coord: spatial.HexCoord{Q: 0, R: 0}},
				},
			},
		},
	}
	decisions := []world.MovementDecision{
		{
			AssetID:       "transport-1",
			Kind:          world.MovementDecisionIssue,
			Destination:   &domain.Location{WorldID: "dest"},
			CargoAssetIDs: []string{"cargo-1"},
		},
	}

	mutations, err := we.BuildMovementMutations(decisions, faction, rb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mutations) != 1 {
		t.Fatalf("expected 1 mutation, got %d", len(mutations))
	}
	issued, ok := mutations[0].(domain.MovementOrderIssued)
	if !ok {
		t.Fatalf("expected MovementOrderIssued, got %T", mutations[0])
	}
	if len(issued.Order.CargoAssetIDs) != 1 || issued.Order.CargoAssetIDs[0] != "cargo-1" {
		t.Errorf("expected CargoAssetIDs=[cargo-1], got %v", issued.Order.CargoAssetIDs)
	}
}

func TestTickMovementOrders_MultiHexPerTick(t *testing.T) {
	we := world.NewWithMap(&stubMap{})
	dest := domain.Location{WorldID: "beta"}
	faction := factionWithOrder(0, makePath(10), dest)

	mutations, err := we.TickMovementOrders(faction, testRulebook(3))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mutations) != 1 {
		t.Fatalf("expected 1 mutation, got %d", len(mutations))
	}
	progressed, ok := mutations[0].(domain.MovementOrderProgressed)
	if !ok {
		t.Fatalf("expected MovementOrderProgressed, got %T", mutations[0])
	}
	if progressed.NewStepIdx != 3 {
		t.Errorf("expected NewStepIdx=3, got %d", progressed.NewStepIdx)
	}
}

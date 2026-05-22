package effects_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

// testPath returns n evenly spaced hexes along the Q axis in region "void".
func testPath(n int) []spatial.RegionHex {
	path := make([]spatial.RegionHex, n)
	for i := range path {
		path[i] = spatial.RegionHex{RegionID: "void", Coord: spatial.HexCoord{Q: i, R: 0}}
	}
	return path
}

// stateWithTransport builds a minimal FactionState containing one faction with
// the given transport (and optional cargo) already placed in the asset map.
func stateWithTransport(transport *domain.Asset, cargo *domain.Asset) *state.FactionState {
	assets := map[string]*domain.Asset{transport.ID: transport}
	if cargo != nil {
		assets[cargo.ID] = cargo
	}
	return &state.FactionState{
		Factions: map[string]*domain.Faction{
			"alpha": {ID: "alpha", Assets: assets},
		},
	}
}

func newReactor() *effects.TransportReactor {
	return &effects.TransportReactor{
		FactionID: "alpha",
		AssetID:   "transport-1",
		Profile:   &domain.TransportProfile{CoinCost: 2},
	}
}

// --- MovementOrderIssued ---

func TestTransportReactor_IssuedWithCargo(t *testing.T) {
	path := testPath(5)
	transport := &domain.Asset{
		ID:      "transport-1",
		OwnerID: "alpha",
		CurrentOrder: &domain.MovementOrder{
			AssetID:       "transport-1",
			Path:          path,
			CargoAssetIDs: []string{"cargo-1"},
		},
	}
	cargo := &domain.Asset{
		ID:       "cargo-1",
		OwnerID:  "alpha",
		Location: domain.Location{WorldID: "world-a"},
	}
	factionState := stateWithTransport(transport, cargo)

	mutations := []domain.Mutation{
		domain.MovementOrderIssued{
			FactionID: "alpha",
			AssetID:   "transport-1",
			Order: domain.MovementOrder{
				AssetID:       "transport-1",
				Path:          path,
				CargoAssetIDs: []string{"cargo-1"},
			},
		},
	}
	result := newReactor().OnMutations(mutations, factionState, nil)

	if len(result) != 2 {
		t.Fatalf("expected 2 mutations (CoinDelta + AssetMoved), got %d", len(result))
	}
	coin, ok := result[0].(domain.CoinDelta)
	if !ok {
		t.Fatalf("result[0] = %T, want CoinDelta", result[0])
	}
	if coin.Delta != -2 {
		t.Errorf("CoinDelta.Delta = %d, want -2 (transport_cost)", coin.Delta)
	}
	moved, ok := result[1].(domain.AssetMoved)
	if !ok {
		t.Fatalf("result[1] = %T, want AssetMoved", result[1])
	}
	if moved.AssetID != "cargo-1" {
		t.Errorf("AssetMoved.AssetID = %q, want cargo-1", moved.AssetID)
	}
	if !domain.IsInFlight(moved.ToLocation) {
		t.Errorf("ToLocation.WorldID = %q, want \"\" (mid-flight)", moved.ToLocation.WorldID)
	}
	if got, want := moved.ToLocation.RegionHex.Coord, path[0].Coord; got != want {
		t.Errorf("ToLocation.HexCoords = %v, want %v (Path[0])", got, want)
	}
}

func TestTransportReactor_IssuedEmptyCargo(t *testing.T) {
	path := testPath(3)
	transport := &domain.Asset{ID: "transport-1", OwnerID: "alpha"}
	factionState := stateWithTransport(transport, nil)

	mutations := []domain.Mutation{
		domain.MovementOrderIssued{
			FactionID: "alpha",
			AssetID:   "transport-1",
			Order: domain.MovementOrder{
				AssetID:       "transport-1",
				Path:          path,
				CargoAssetIDs: nil,
			},
		},
	}
	result := newReactor().OnMutations(mutations, factionState, nil)
	if len(result) != 0 {
		t.Errorf("expected no reactor output for empty cargo manifest, got %d mutations", len(result))
	}
}

func TestTransportReactor_IssuedUnrelatedAsset(t *testing.T) {
	path := testPath(3)
	transport := &domain.Asset{ID: "transport-1", OwnerID: "alpha"}
	factionState := stateWithTransport(transport, nil)

	mutations := []domain.Mutation{
		domain.MovementOrderIssued{
			FactionID: "alpha",
			AssetID:   "other-transport", // different asset — reactor should ignore
			Order: domain.MovementOrder{
				AssetID:       "other-transport",
				Path:          path,
				CargoAssetIDs: []string{"cargo-1"},
			},
		},
	}
	result := newReactor().OnMutations(mutations, factionState, nil)
	if len(result) != 0 {
		t.Errorf("expected no reactor output for unrelated asset, got %d mutations", len(result))
	}
}

// --- MovementOrderProgressed ---

func TestTransportReactor_Progressed(t *testing.T) {
	path := testPath(5)
	newHex := path[2]
	transport := &domain.Asset{
		ID:      "transport-1",
		OwnerID: "alpha",
		CurrentOrder: &domain.MovementOrder{
			AssetID:       "transport-1",
			Path:          path,
			CargoAssetIDs: []string{"cargo-1"},
		},
	}
	cargo := &domain.Asset{
		ID:       "cargo-1",
		OwnerID:  "alpha",
		Location: domain.Location{RegionHex: path[0]},
	}
	factionState := stateWithTransport(transport, cargo)

	mutations := []domain.Mutation{
		domain.MovementOrderProgressed{
			FactionID:  "alpha",
			AssetID:    "transport-1",
			NewStepIdx: 2,
			RegionHex:  newHex,
		},
	}
	result := newReactor().OnMutations(mutations, factionState, nil)

	if len(result) != 1 {
		t.Fatalf("expected 1 AssetMoved for cargo follow, got %d", len(result))
	}
	moved, ok := result[0].(domain.AssetMoved)
	if !ok {
		t.Fatalf("result[0] = %T, want AssetMoved", result[0])
	}
	if moved.AssetID != "cargo-1" {
		t.Errorf("AssetMoved.AssetID = %q, want cargo-1", moved.AssetID)
	}
	if got, want := moved.ToLocation.RegionHex.Coord, newHex.Coord; got != want {
		t.Errorf("ToLocation.HexCoords = %v, want %v (new transport hex)", got, want)
	}
}

// --- MovementOrderCompleted ---

func TestTransportReactor_Completed(t *testing.T) {
	path := testPath(5)
	finalLoc := domain.Location{
		WorldID:   "world-c",
		RegionHex: spatial.RegionHex{RegionID: "void", Coord: spatial.HexCoord{Q: 4, R: 0}},
	}
	transport := &domain.Asset{
		ID:      "transport-1",
		OwnerID: "alpha",
		CurrentOrder: &domain.MovementOrder{
			AssetID:       "transport-1",
			Path:          path,
			CargoAssetIDs: []string{"cargo-1"},
		},
	}
	cargo := &domain.Asset{
		ID:       "cargo-1",
		OwnerID:  "alpha",
		Location: domain.Location{RegionHex: path[2]},
	}
	factionState := stateWithTransport(transport, cargo)

	mutations := []domain.Mutation{
		domain.MovementOrderCompleted{
			FactionID:     "alpha",
			AssetID:       "transport-1",
			FinalLocation: finalLoc,
		},
	}
	result := newReactor().OnMutations(mutations, factionState, nil)

	if len(result) != 1 {
		t.Fatalf("expected 1 AssetMoved for cargo arrival, got %d", len(result))
	}
	moved, ok := result[0].(domain.AssetMoved)
	if !ok {
		t.Fatalf("result[0] = %T, want AssetMoved", result[0])
	}
	if moved.ToLocation != finalLoc {
		t.Errorf("ToLocation = %v, want %v (final destination)", moved.ToLocation, finalLoc)
	}
}

// --- MovementOrderCancelled ---

func TestTransportReactor_Cancelled(t *testing.T) {
	path := testPath(5)
	strandedAt := domain.Location{RegionHex: path[2]}
	transport := &domain.Asset{
		ID:      "transport-1",
		OwnerID: "alpha",
		CurrentOrder: &domain.MovementOrder{
			AssetID:       "transport-1",
			Path:          path,
			CargoAssetIDs: []string{"cargo-1"},
		},
	}
	cargo := &domain.Asset{
		ID:       "cargo-1",
		OwnerID:  "alpha",
		Location: domain.Location{RegionHex: path[2]},
	}
	factionState := stateWithTransport(transport, cargo)

	mutations := []domain.Mutation{
		domain.MovementOrderCancelled{
			FactionID:  "alpha",
			AssetID:    "transport-1",
			StrandedAt: strandedAt,
		},
	}
	result := newReactor().OnMutations(mutations, factionState, nil)

	if len(result) != 1 {
		t.Fatalf("expected 1 AssetMoved stranding cargo, got %d", len(result))
	}
	moved, ok := result[0].(domain.AssetMoved)
	if !ok {
		t.Fatalf("result[0] = %T, want AssetMoved", result[0])
	}
	if moved.ToLocation != strandedAt {
		t.Errorf("ToLocation = %v, want %v (stranded hex)", moved.ToLocation, strandedAt)
	}
}

// --- MovementOrderRevised ---

func TestTransportReactor_Revised(t *testing.T) {
	path := testPath(3)
	transport := &domain.Asset{ID: "transport-1", OwnerID: "alpha"}
	factionState := stateWithTransport(transport, nil)

	mutations := []domain.Mutation{
		domain.MovementOrderRevised{
			FactionID: "alpha",
			AssetID:   "transport-1",
			NewOrder: domain.MovementOrder{
				AssetID:       "transport-1",
				Path:          path,
				CargoAssetIDs: []string{"cargo-1"},
			},
		},
	}
	result := newReactor().OnMutations(mutations, factionState, nil)
	if len(result) != 0 {
		t.Errorf("expected no reactor output on Revised (cargo manifest carried forward by BuildMovementMutations), got %d", len(result))
	}
}

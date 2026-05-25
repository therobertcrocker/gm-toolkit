package scenarios

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

// Transport lifecycle path: 5 hexes, Speed=2 → completes in 2 ticks after issuance.
var transportPath = []spatial.RegionHex{
	{RegionID: "void", Coord: spatial.HexCoord{Q: 0, R: 0}},
	{RegionID: "void", Coord: spatial.HexCoord{Q: 1, R: 0}},
	{RegionID: "void", Coord: spatial.HexCoord{Q: 2, R: 0}},
	{RegionID: "void", Coord: spatial.HexCoord{Q: 3, R: 0}},
	{RegionID: "void", Coord: spatial.HexCoord{Q: 4, R: 0}},
}

// setupTransportHarness builds a harness with a transport-capable asset ("transport-1",
// Speed=2, CoinCost=1, MaxCargo=1, CargoTypes=[SpecialForces]) and one cargo asset
// ("cargo-1", Type=SpecialForces) co-located at "world-a". The EffectsEngine
// is wired and hooks are registered before returning.
func setupTransportHarness(t *testing.T) *testharness.Harness {
	t.Helper()
	h := testharness.NewHarness(t)

	transportProfile := &domain.TransportProfile{
		MaxHex:     4,
		CoinCost:   1,
		CargoTypes: []domain.AssetType{domain.TypeSpecialForces},
		MaxCargo:   1,
	}
	transportDef := &domain.AssetDefinition{
		ID:          "test-transport",
		Speed:       2,
		DriftRating: 1,
		Transport:   transportProfile,
	}
	cargoDef := &domain.AssetDefinition{
		ID:   "test-cargo",
		Type: domain.TypeSpecialForces,
	}
	h.Engine.Rulebook.Assets["test-transport"] = transportDef
	h.Engine.Rulebook.Assets["test-cargo"] = cargoDef
	h.Engine.Rulebook.DriftCosts = []int{5}

	h.Engine.Effect.Register(effects.NewTransportHandler(transportDef))

	h.AddFaction("alpha", "world-a", 4, 3, 2)
	alpha := h.FactionState.Factions["alpha"]
	alpha.Coin = 10

	transport := &domain.Asset{
		ID:           "transport-1",
		DefinitionID: "test-transport",
		OwnerID:      "alpha",
		Location:     domain.Location{WorldID: "world-a"},
		CurrentHP:    3,
		Ready:        true,
		Maintained:   true,
	}
	cargo := &domain.Asset{
		ID:           "cargo-1",
		DefinitionID: "test-cargo",
		OwnerID:      "alpha",
		Location:     domain.Location{WorldID: "world-a"},
		CurrentHP:    3,
		Ready:        true,
		Maintained:   true,
	}
	alpha.Assets[transport.ID] = transport
	alpha.Assets[cargo.ID] = cargo

	h.SpatialMap.PathFn = func(_, _ spatial.RegionHex, _ int) ([]spatial.RegionHex, int, error) {
		return transportPath, len(transportPath) - 1, nil
	}
	h.Collector.SelectActionFn = func(*domain.Faction, []action.Action) (action.Action, error) {
		return nil, nil
	}
	// Default cargo selector: always pick cargo-1 when it appears in eligible.
	h.Collector.SelectTransportCargoFn = func(_ *domain.Asset, eligible []*domain.Asset, _ *domain.TransportProfile) ([]*domain.Asset, error) {
		for _, a := range eligible {
			if a.ID == "cargo-1" {
				return []*domain.Asset{a}, nil
			}
		}
		return nil, nil
	}

	return h
}

// transportMutationsFor collects all movement-phase mutations for a faction
// across both MovementTicked and MovementResolved observer events.
func transportMutationsFor(t *testing.T, observer *testharness.RecordingObserver, factionID string) []domain.Mutation {
	t.Helper()
	return movementMutationsFor(t, observer, factionID)
}

// Task 3 — full lifecycle: issue → tick → complete with cargo co-movement.
func TestTransport_LifecycleIssueTickComplete(t *testing.T) {
	h := setupTransportHarness(t)

	h.Collector.SelectMovementDecisionsFn = func(_ *domain.Faction, eligible []*domain.Asset) ([]world.MovementDecision, error) {
		for _, asset := range eligible {
			if asset.ID == "transport-1" && asset.CurrentOrder == nil && asset.Location.WorldID != "world-c" {
				return []world.MovementDecision{{
					AssetID:     asset.ID,
					Kind:        world.MovementDecisionIssue,
					Destination: &domain.Location{WorldID: "world-c"},
				}}, nil
			}
		}
		return nil, nil
	}

	// Turn 1 — issuance.
	runCycle(t, h)
	mutations := transportMutationsFor(t, h.Observer, "alpha")

	assertOneMutationOfKind[domain.MovementOrderIssued](t, mutations, "Turn 1 issuance")

	var sawCoinDelta bool
	for _, m := range mutations {
		if cd, ok := m.(domain.CoinDelta); ok && cd.Cause == "transport_cost" {
			sawCoinDelta = true
		}
	}
	if !sawCoinDelta {
		t.Error("Turn 1: no CoinDelta{Cause: transport_cost} — transport Coin cost not charged")
	}

	transport := h.FactionState.Factions["alpha"].Assets["transport-1"]
	cargo := h.FactionState.Factions["alpha"].Assets["cargo-1"]

	if transport.CurrentOrder == nil {
		t.Fatal("Turn 1: transport CurrentOrder is nil after issuance")
	}
	if !domain.IsInFlight(transport.Location) {
		t.Errorf("Turn 1: transport.Location.WorldID = %q, want \"\" (mid-flight)", transport.Location.WorldID)
	}
	if got, want := transport.Location.RegionHex, transportPath[0]; got != want {
		t.Errorf("Turn 1: transport.Location.RegionHex = %+v, want %+v (Path[0])", got, want)
	}
	if got, want := cargo.Location.RegionHex, transportPath[0]; got != want {
		t.Errorf("Turn 1: cargo.Location.RegionHex = %+v, want %+v (follows transport to Path[0])", got, want)
	}

	// Turn 2 — tick (StepIdx 0→2, mid-flight).
	runCycle(t, h)
	mutations = transportMutationsFor(t, h.Observer, "alpha")

	progressed := assertOneMutationOfKind[domain.MovementOrderProgressed](t, mutations, "Turn 2 tick")
	if got, want := progressed.NewStepIdx, 2; got != want {
		t.Errorf("Turn 2: Progressed.NewStepIdx = %d, want %d", got, want)
	}

	transport = h.FactionState.Factions["alpha"].Assets["transport-1"]
	cargo = h.FactionState.Factions["alpha"].Assets["cargo-1"]

	if got, want := transport.Location.RegionHex, transportPath[2]; got != want {
		t.Errorf("Turn 2: transport.Location.RegionHex = %+v, want %+v", got, want)
	}
	if got, want := cargo.Location.RegionHex, transportPath[2]; got != want {
		t.Errorf("Turn 2: cargo.Location.RegionHex = %+v, want %+v (follows transport)", got, want)
	}

	// Turn 3 — completion (StepIdx 2→4 = len-1).
	runCycle(t, h)
	mutations = transportMutationsFor(t, h.Observer, "alpha")

	completed := assertOneMutationOfKind[domain.MovementOrderCompleted](t, mutations, "Turn 3 completion")
	if got, want := completed.FinalLocation.WorldID, "world-c"; got != want {
		t.Errorf("Turn 3: Completed.FinalLocation.WorldID = %q, want %q", got, want)
	}

	transport = h.FactionState.Factions["alpha"].Assets["transport-1"]
	cargo = h.FactionState.Factions["alpha"].Assets["cargo-1"]

	if transport.CurrentOrder != nil {
		t.Errorf("Turn 3: transport.CurrentOrder = %+v, want nil after completion", transport.CurrentOrder)
	}
	if got, want := transport.Location.WorldID, "world-c"; got != want {
		t.Errorf("Turn 3: transport.Location.WorldID = %q, want %q", got, want)
	}
	if got, want := cargo.Location.WorldID, "world-c"; got != want {
		t.Errorf("Turn 3: cargo.Location.WorldID = %q, want %q (cargo arrives with transport)", got, want)
	}
}

// Task 4 — cancel mid-flight: cargo strands at the transport's current hex.
func TestTransport_CancelMidFlight(t *testing.T) {
	h := setupTransportHarness(t)

	h.Collector.SelectMovementDecisionsFn = func(_ *domain.Faction, eligible []*domain.Asset) ([]world.MovementDecision, error) {
		for _, asset := range eligible {
			if asset.ID != "transport-1" {
				continue
			}
			if asset.CurrentOrder == nil {
				return []world.MovementDecision{{
					AssetID:     asset.ID,
					Kind:        world.MovementDecisionIssue,
					Destination: &domain.Location{WorldID: "world-c"},
				}}, nil
			}
			return []world.MovementDecision{{
				AssetID: asset.ID,
				Kind:    world.MovementDecisionCancel,
			}}, nil
		}
		return nil, nil
	}

	// Turn 1 — issuance.
	runCycle(t, h)
	// Turn 2 — tick (0→2) concurrent with cancellation.
	runCycle(t, h)

	mutations := transportMutationsFor(t, h.Observer, "alpha")
	cancelled := assertOneMutationOfKind[domain.MovementOrderCancelled](t, mutations, "Turn 2 cancel")

	if cancelled.AssetID != "transport-1" {
		t.Errorf("Cancelled.AssetID = %q, want transport-1", cancelled.AssetID)
	}

	transport := h.FactionState.Factions["alpha"].Assets["transport-1"]
	cargo := h.FactionState.Factions["alpha"].Assets["cargo-1"]

	if transport.CurrentOrder != nil {
		t.Errorf("transport.CurrentOrder = %+v, want nil after cancel", transport.CurrentOrder)
	}
	if !domain.IsInFlight(transport.Location) {
		t.Errorf("transport.Location.WorldID = %q, want \"\" (stranded mid-flight)", transport.Location.WorldID)
	}
	if got, want := transport.Location.RegionHex, transportPath[2]; got != want {
		t.Errorf("transport.Location.RegionHex = %+v, want %+v (post-tick hex)", got, want)
	}
	if !domain.IsInFlight(cargo.Location) {
		t.Errorf("cargo.Location.WorldID = %q, want \"\" (stranded at mid-flight hex)", cargo.Location.WorldID)
	}
	if got, want := cargo.Location.RegionHex, transportPath[2]; got != want {
		t.Errorf("cargo.Location.RegionHex = %+v, want %+v (stranded with transport)", got, want)
	}
}

// Task 5 — revise mid-flight: cargo manifest carries forward; cargo completes at new dest.
func TestTransport_ReviseMidFlight(t *testing.T) {
	h := setupTransportHarness(t)

	// Path to new destination; 3 hexes with Speed=2 completes in one tick.
	pathToD := []spatial.RegionHex{
		transportPath[2],
		{RegionID: "void", Coord: spatial.HexCoord{Q: 5, R: 0}},
		{RegionID: "void", Coord: spatial.HexCoord{Q: 6, R: 0}},
	}

	callCount := 0
	h.SpatialMap.PathFn = func(_, _ spatial.RegionHex, _ int) ([]spatial.RegionHex, int, error) {
		callCount++
		if callCount == 1 {
			return transportPath, len(transportPath) - 1, nil
		}
		return pathToD, len(pathToD) - 1, nil
	}

	h.Collector.SelectMovementDecisionsFn = func(_ *domain.Faction, eligible []*domain.Asset) ([]world.MovementDecision, error) {
		for _, asset := range eligible {
			if asset.ID != "transport-1" {
				continue
			}
			// Only issue from the original starting world; guard against re-issuing after arrival.
			if asset.CurrentOrder == nil && asset.Location.WorldID == "world-a" {
				return []world.MovementDecision{{
					AssetID:     asset.ID,
					Kind:        world.MovementDecisionIssue,
					Destination: &domain.Location{WorldID: "world-c"},
				}}, nil
			}
			if asset.CurrentOrder != nil && asset.CurrentOrder.Destination.WorldID == "world-c" {
				return []world.MovementDecision{{
					AssetID:     asset.ID,
					Kind:        world.MovementDecisionRevise,
					Destination: &domain.Location{WorldID: "world-d"},
				}}, nil
			}
		}
		return nil, nil
	}

	// Turn 1 — issuance to world-c.
	runCycle(t, h)
	// Turn 2 — tick (0→2) concurrent with revision to world-d.
	runCycle(t, h)

	mutations := transportMutationsFor(t, h.Observer, "alpha")
	revised := assertOneMutationOfKind[domain.MovementOrderRevised](t, mutations, "Turn 2 revise")

	if got, want := revised.NewOrder.Destination.WorldID, "world-d"; got != want {
		t.Errorf("Revised.NewOrder.Destination.WorldID = %q, want %q", got, want)
	}
	if got, want := revised.NewOrder.CargoAssetIDs, []string{"cargo-1"}; len(got) != 1 || got[0] != want[0] {
		t.Errorf("Revised.NewOrder.CargoAssetIDs = %v, want [cargo-1] (carried forward)", got)
	}

	var sawRevisionCost bool
	for _, m := range mutations {
		if cd, ok := m.(domain.CoinDelta); ok && cd.Cause == domain.CauseMovementRevision {
			sawRevisionCost = true
			if cd.Delta != -1 {
				t.Errorf("revision CoinDelta.Delta = %d, want -1", cd.Delta)
			}
		}
	}
	if !sawRevisionCost {
		t.Errorf("Turn 2: no CoinDelta with cause %q", domain.CauseMovementRevision)
	}
	var sawNewTransportCost bool
	for _, m := range mutations {
		if cd, ok := m.(domain.CoinDelta); ok && cd.Cause == "transport_cost" {
			sawNewTransportCost = true
		}
	}
	if sawNewTransportCost {
		t.Error("Turn 2: unexpected transport_cost CoinDelta on revise — transport cost is only charged at issuance")
	}

	// Turn 3 — tick on new path (len=3, StepIdx=0+2=2=len-1 → completion).
	runCycle(t, h)
	mutations = transportMutationsFor(t, h.Observer, "alpha")

	completed := assertOneMutationOfKind[domain.MovementOrderCompleted](t, mutations, "Turn 3 completion")
	if got, want := completed.FinalLocation.WorldID, "world-d"; got != want {
		t.Errorf("Turn 3: Completed.FinalLocation.WorldID = %q, want %q (revised dest)", got, want)
	}

	transport := h.FactionState.Factions["alpha"].Assets["transport-1"]
	cargo := h.FactionState.Factions["alpha"].Assets["cargo-1"]

	if transport.CurrentOrder != nil {
		t.Errorf("Turn 3: transport.CurrentOrder should be nil after completion")
	}
	if got, want := transport.Location.WorldID, "world-d"; got != want {
		t.Errorf("Turn 3: transport.Location.WorldID = %q, want %q", got, want)
	}
	if got, want := cargo.Location.WorldID, "world-d"; got != want {
		t.Errorf("Turn 3: cargo.Location.WorldID = %q, want %q (cargo arrives at revised dest)", got, want)
	}
}

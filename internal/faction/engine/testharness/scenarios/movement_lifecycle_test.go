package scenarios

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

// Movement-phase integration tests covering the full order lifecycle:
// issue → tick → completion, cancellation mid-flight, and revision mid-flight.
// All three use a single transport-less asset with Speed=2 walking a scripted
// 5-hex path (Path[0]..Path[4]) from "Tartarus" to "Krylos".

var lifecyclePath = []spatial.RegionHex{
	{RegionID: "void", Coord: spatial.HexCoord{Q: 0, R: 0}},
	{RegionID: "void", Coord: spatial.HexCoord{Q: 1, R: 0}},
	{RegionID: "void", Coord: spatial.HexCoord{Q: 2, R: 0}},
	{RegionID: "void", Coord: spatial.HexCoord{Q: 3, R: 0}},
	{RegionID: "void", Coord: spatial.HexCoord{Q: 4, R: 0}},
}

func setupMovementHarness(t *testing.T) *testharness.Harness {
	t.Helper()
	h := testharness.NewHarness(t, testDataDir)
	h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	h.Engine.Rulebook.Assets[testharness.DefSecurityPersonnel].Speed = 2
	h.SpatialMap.PathFn = func(from, _ spatial.RegionHex, _ int) ([]spatial.RegionHex, int, error) {
		return lifecyclePath, len(lifecyclePath) - 1, nil
	}
	h.Collector.SelectActionFn = func(*domain.Faction, []action.Action) (action.Action, error) {
		return nil, nil
	}
	return h
}

func runCycle(t *testing.T, h *testharness.Harness) {
	t.Helper()
	h.Observer.Events = nil
	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Cfg, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}
}

func assertOneMutationOfKind[T domain.Mutation](t *testing.T, mutations []domain.Mutation, label string) T {
	t.Helper()
	var found T
	count := 0
	for _, m := range mutations {
		if cast, ok := m.(T); ok {
			found = cast
			count++
		}
	}
	if count != 1 {
		t.Fatalf("%s: got %d %T mutations in %v, want exactly 1", label, count, found, mutations)
	}
	return found
}

// Task 13: faction A issues an order on its asset; over three turns the order
// is issued, ticks once, then completes at the destination.
func TestMovement_LifecycleIssueTickComplete(t *testing.T) {
	h := setupMovementHarness(t)

	h.Collector.SelectMovementDecisionsFn = func(faction *domain.Faction, eligible []*domain.Asset) ([]world.MovementDecision, error) {
		asset := eligible[0]
		if asset.CurrentOrder != nil {
			return nil, nil
		}
		if asset.Location.WorldID == "Krylos" {
			return nil, nil
		}
		return []world.MovementDecision{{
			AssetID:     asset.ID,
			Kind:        world.MovementDecisionIssue,
			Destination: &domain.Location{WorldID: "Krylos"},
		}}, nil
	}

	// Turn 1 — issuance.
	runCycle(t, h)
	asset := h.FactionState.Factions["alpha"].Assets["alpha-asset-1"]

	issuance := assertOneMutationOfKind[domain.MovementOrderIssued](t, movementMutationsFor(t, h.Observer, "alpha"), "Turn 1")
	if got, want := len(issuance.Order.Path), len(lifecyclePath); got != want {
		t.Errorf("issued Path length: got %d, want %d", got, want)
	}
	if asset.CurrentOrder == nil {
		t.Fatalf("Turn 1: asset CurrentOrder is nil, want issued order")
	}
	if !domain.IsInFlight(asset.Location) {
		t.Errorf("Turn 1: asset.Location.WorldID: got %q, want \"\" (asset moves on issue)", asset.Location.WorldID)
	}
	if got, want := asset.Location.RegionHex, lifecyclePath[0]; got != want {
		t.Errorf("Turn 1: asset.Location.RegionHex: got %+v, want %+v", got, want)
	}

	// Turn 2 — tick advances StepIdx by Speed (2).
	runCycle(t, h)
	asset = h.FactionState.Factions["alpha"].Assets["alpha-asset-1"]

	progressed := assertOneMutationOfKind[domain.MovementOrderProgressed](t, movementMutationsFor(t, h.Observer, "alpha"), "Turn 2")
	if got, want := progressed.NewStepIdx, 2; got != want {
		t.Errorf("Turn 2: Progressed.NewStepIdx: got %d, want %d", got, want)
	}
	if asset.CurrentOrder == nil {
		t.Fatalf("Turn 2: asset CurrentOrder cleared prematurely")
	}
	if got, want := asset.CurrentOrder.StepIdx, 2; got != want {
		t.Errorf("Turn 2: CurrentOrder.StepIdx: got %d, want %d", got, want)
	}
	if got, want := asset.Location.RegionHex, lifecyclePath[2]; got != want {
		t.Errorf("Turn 2: asset.Location.RegionHex: got %+v, want %+v", got, want)
	}

	// Turn 3 — tick reaches end of path; completion.
	runCycle(t, h)
	asset = h.FactionState.Factions["alpha"].Assets["alpha-asset-1"]

	completed := assertOneMutationOfKind[domain.MovementOrderCompleted](t, movementMutationsFor(t, h.Observer, "alpha"), "Turn 3")
	if got, want := completed.FinalLocation.WorldID, "Krylos"; got != want {
		t.Errorf("Turn 3: Completed.FinalLocation.WorldID: got %q, want %q", got, want)
	}
	if asset.CurrentOrder != nil {
		t.Errorf("Turn 3: asset.CurrentOrder: got %+v, want nil after completion", asset.CurrentOrder)
	}
	if got, want := asset.Location.WorldID, "Krylos"; got != want {
		t.Errorf("Turn 3: asset.Location.WorldID: got %q, want %q", got, want)
	}
}

// Task 14: faction A issues an order, then cancels it mid-flight. The tick on
// the cancellation turn runs concurrent with the cancel decision, so the asset
// ends up stranded at the post-tick hex with WorldID cleared.
func TestMovement_CancelMidFlight(t *testing.T) {
	h := setupMovementHarness(t)

	h.Collector.SelectMovementDecisionsFn = func(faction *domain.Faction, eligible []*domain.Asset) ([]world.MovementDecision, error) {
		asset := eligible[0]
		if asset.CurrentOrder == nil {
			return []world.MovementDecision{{
				AssetID:     asset.ID,
				Kind:        world.MovementDecisionIssue,
				Destination: &domain.Location{WorldID: "Krylos"},
			}}, nil
		}
		return []world.MovementDecision{{
			AssetID: asset.ID,
			Kind:    world.MovementDecisionCancel,
		}}, nil
	}

	// Turn 1 — issuance.
	runCycle(t, h)
	// Turn 2 — tick (Path[0] → Path[2]) concurrent with cancellation.
	runCycle(t, h)

	asset := h.FactionState.Factions["alpha"].Assets["alpha-asset-1"]
	mutations := movementMutationsFor(t, h.Observer, "alpha")
	cancelled := assertOneMutationOfKind[domain.MovementOrderCancelled](t, mutations, "Turn 2")

	if cancelled.AssetID != asset.ID {
		t.Errorf("Cancelled.AssetID: got %q, want %q", cancelled.AssetID, asset.ID)
	}
	if asset.CurrentOrder != nil {
		t.Errorf("asset.CurrentOrder: got %+v, want nil after cancel", asset.CurrentOrder)
	}
	if !domain.IsInFlight(asset.Location) {
		t.Errorf("asset.Location.WorldID: got %q, want \"\" (stranded mid-flight)", asset.Location.WorldID)
	}
	if got, want := asset.Location.RegionHex, lifecyclePath[2]; got != want {
		t.Errorf("asset.Location.RegionHex: got %+v, want %+v (post-tick hex)", got, want)
	}
}

// Task 15: faction A issues an order to Krylos, then revises mid-flight to a
// new destination (Volari). The revision costs 1 Coin and the order completes
// at the new destination via subsequent ticks.
func TestMovement_ReviseMidFlight(t *testing.T) {
	h := setupMovementHarness(t)

	pathToVolari := []spatial.RegionHex{
		{RegionID: "void", Coord: spatial.HexCoord{Q: 10, R: 0}},
		{RegionID: "void", Coord: spatial.HexCoord{Q: 11, R: 0}},
		{RegionID: "void", Coord: spatial.HexCoord{Q: 12, R: 0}},
	}
	callCount := 0
	h.SpatialMap.PathFn = func(_, _ spatial.RegionHex, _ int) ([]spatial.RegionHex, int, error) {
		callCount++
		if callCount == 1 {
			return lifecyclePath, len(lifecyclePath) - 1, nil
		}
		return pathToVolari, len(pathToVolari) - 1, nil
	}

	h.Collector.SelectMovementDecisionsFn = func(faction *domain.Faction, eligible []*domain.Asset) ([]world.MovementDecision, error) {
		asset := eligible[0]
		if asset.CurrentOrder == nil {
			if asset.Location.WorldID == "Volari" {
				return nil, nil
			}
			return []world.MovementDecision{{
				AssetID:     asset.ID,
				Kind:        world.MovementDecisionIssue,
				Destination: &domain.Location{WorldID: "Krylos"},
			}}, nil
		}
		if asset.CurrentOrder.Destination.WorldID == "Krylos" {
			return []world.MovementDecision{{
				AssetID:     asset.ID,
				Kind:        world.MovementDecisionRevise,
				Destination: &domain.Location{WorldID: "Volari"},
			}}, nil
		}
		return nil, nil
	}

	// Turn 1 — issuance toward Krylos.
	runCycle(t, h)
	// Turn 2 — revision toward Volari (concurrent with tick).
	runCycle(t, h)

	mutations := movementMutationsFor(t, h.Observer, "alpha")
	revised := assertOneMutationOfKind[domain.MovementOrderRevised](t, mutations, "Turn 2")
	if got, want := revised.NewOrder.Destination.WorldID, "Volari"; got != want {
		t.Errorf("Revised.NewOrder.Destination.WorldID: got %q, want %q", got, want)
	}

	var coinDelta domain.CoinDelta
	var sawCoinDelta bool
	for _, m := range mutations {
		if d, ok := m.(domain.CoinDelta); ok && d.Cause == domain.CauseMovementRevision {
			coinDelta = d
			sawCoinDelta = true
		}
	}
	if !sawCoinDelta {
		t.Fatalf("Turn 2: no CoinDelta with cause %q in %v", domain.CauseMovementRevision, mutations)
	}
	if coinDelta.Delta != -1 {
		t.Errorf("Turn 2: CoinDelta.Delta: got %d, want -1", coinDelta.Delta)
	}

	// Turn 3 — tick on the new (Volari) path; this trip is length 3 and Speed
	// is 2, so the order completes at Volari.
	runCycle(t, h)

	asset := h.FactionState.Factions["alpha"].Assets["alpha-asset-1"]
	completed := assertOneMutationOfKind[domain.MovementOrderCompleted](t, movementMutationsFor(t, h.Observer, "alpha"), "Turn 3")
	if got, want := completed.FinalLocation.WorldID, "Volari"; got != want {
		t.Errorf("Turn 3: Completed.FinalLocation.WorldID: got %q, want %q (revised destination)", got, want)
	}
	if asset.CurrentOrder != nil {
		t.Errorf("Turn 3: asset.CurrentOrder: got %+v, want nil after completion", asset.CurrentOrder)
	}
	if got, want := asset.Location.WorldID, "Volari"; got != want {
		t.Errorf("Turn 3: asset.Location.WorldID: got %q, want %q", got, want)
	}
}

// TestMovement_RevisePathOriginsFromPostTickHex pins the tick-before-decisions
// ordering: a revise built on the same turn as a tick must originate from the
// asset's POST-tick hex. Under the old (collected-then-applied) ordering the
// revise was built from pre-tick state, so the new order's Path[0] disagreed
// with asset.Location and the asset "jumped" on the next tick.
func TestMovement_RevisePathOriginsFromPostTickHex(t *testing.T) {
	h := setupMovementHarness(t)
	longPath := []spatial.RegionHex{
		{RegionID: "void", Coord: spatial.HexCoord{Q: 0, R: 0}},
		{RegionID: "void", Coord: spatial.HexCoord{Q: 1, R: 0}},
		{RegionID: "void", Coord: spatial.HexCoord{Q: 2, R: 0}},
		{RegionID: "void", Coord: spatial.HexCoord{Q: 3, R: 0}},
		{RegionID: "void", Coord: spatial.HexCoord{Q: 4, R: 0}},
		{RegionID: "void", Coord: spatial.HexCoord{Q: 5, R: 0}},
		{RegionID: "void", Coord: spatial.HexCoord{Q: 6, R: 0}},
	}
	var reviseFromHex spatial.RegionHex
	callCount := 0
	h.SpatialMap.PathFn = func(from, _ spatial.RegionHex, _ int) ([]spatial.RegionHex, int, error) {
		callCount++
		if callCount == 1 {
			return longPath, len(longPath) - 1, nil
		}
		reviseFromHex = from
		return []spatial.RegionHex{from, {RegionID: "void", Coord: spatial.HexCoord{Q: 10, R: 0}}}, 1, nil
	}

	h.Collector.SelectMovementDecisionsFn = func(_ *domain.Faction, eligible []*domain.Asset) ([]world.MovementDecision, error) {
		asset := eligible[0]
		if asset.CurrentOrder == nil {
			if asset.Location.WorldID == "Volari" {
				return nil, nil
			}
			return []world.MovementDecision{{
				AssetID:     asset.ID,
				Kind:        world.MovementDecisionIssue,
				Destination: &domain.Location{WorldID: "Krylos"},
			}}, nil
		}
		if asset.CurrentOrder.Destination.WorldID == "Krylos" {
			return []world.MovementDecision{{
				AssetID:     asset.ID,
				Kind:        world.MovementDecisionRevise,
				Destination: &domain.Location{WorldID: "Volari"},
			}}, nil
		}
		return nil, nil
	}

	// Turn 1 — issuance toward Krylos along longPath.
	runCycle(t, h)
	// Turn 2 — tick (StepIdx 0 → 2, asset at longPath[2]) then revise. Revise
	// must originate from longPath[2], not longPath[0].
	runCycle(t, h)

	mutations := movementMutationsFor(t, h.Observer, "alpha")
	revised := assertOneMutationOfKind[domain.MovementOrderRevised](t, mutations, "Turn 2")

	if got, want := reviseFromHex, longPath[2]; got != want {
		t.Errorf("pathfinder invoked with from = %+v, want %+v (post-tick hex; old ordering would have used %+v)", got, want, longPath[0])
	}
	if got, want := revised.NewOrder.Origin.RegionHex, longPath[2]; got != want {
		t.Errorf("Revised.NewOrder.Origin.RegionHex: got %+v, want %+v (post-tick hex)", got, want)
	}
	if got, want := revised.NewOrder.Path[0], longPath[2]; got != want {
		t.Errorf("Revised.NewOrder.Path[0]: got %+v, want %+v (post-tick hex)", got, want)
	}

	asset := h.FactionState.Factions["alpha"].Assets["alpha-asset-1"]
	if got, want := asset.Location.RegionHex, asset.CurrentOrder.Path[0]; got != want {
		t.Errorf("asset.Location.RegionHex (%+v) disagrees with CurrentOrder.Path[0] (%+v); asset would jump on next tick", got, want)
	}
}

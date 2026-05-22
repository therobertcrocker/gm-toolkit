package scenarios

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

// LockSkip + movement: the Movement Phase runs unconditionally; only the
// Action Phase is gated. These scenarios cover the three movement-phase
// inputs (no orders, in-flight tick, new issue) while a faction sits under
// LockSkip via an active Change Homeworld goal.

func newLockSkipGoal() *domain.ActiveGoal {
	return &domain.ActiveGoal{
		GoalID:         "G-012",
		TargetWorld:    domain.Location{WorldID: "Krylos"},
		TurnsRemaining: 2,
	}
}

func assertActionPhaseSkipped(t *testing.T, observer *testharness.RecordingObserver, factionID string) {
	t.Helper()
	for _, ev := range observer.Events {
		if ev.Kind == "ActionSelected" && ev.Faction != nil && ev.Faction.ID == factionID {
			t.Errorf("ActionSelected fired for LockSkip faction %q — action phase was not skipped", factionID)
		}
	}
}

func movementMutationsFor(t *testing.T, observer *testharness.RecordingObserver, factionID string) []domain.Mutation {
	t.Helper()
	var mutations []domain.Mutation
	var sawResolved bool
	for _, ev := range observer.Events {
		if ev.Faction == nil || ev.Faction.ID != factionID {
			continue
		}
		if ev.Kind != "MovementTicked" && ev.Kind != "MovementResolved" {
			continue
		}
		batch, ok := ev.Payload.([]domain.Mutation)
		if !ok {
			t.Fatalf("%s payload type: got %T, want []domain.Mutation", ev.Kind, ev.Payload)
		}
		mutations = append(mutations, batch...)
		if ev.Kind == "MovementResolved" {
			sawResolved = true
		}
	}
	if !sawResolved {
		t.Fatalf("no MovementResolved event for faction %q", factionID)
	}
	return mutations
}

// Scenario 1: LockSkip faction with no movable assets. Movement Phase still
// runs (MovementResolved fires with no mutations) and the action phase is
// skipped.
func TestLockSkip_MovementRuns_NoOrders(t *testing.T) {
	h := testharness.NewHarness(t, testDataDir)
	locked := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	locked.ActiveGoal = newLockSkipGoal()

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Cfg, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	mutations := movementMutationsFor(t, h.Observer, "alpha")
	if len(mutations) != 0 {
		t.Errorf("MovementResolved mutations: got %d, want 0 (no movable assets)", len(mutations))
	}
	assertActionPhaseSkipped(t, h.Observer, "alpha")
}

// Scenario 2: LockSkip faction with a pre-seeded in-flight order. The tick
// advances StepIdx by Speed and a MovementOrderProgressed mutation is
// emitted; the action phase is still skipped.
func TestLockSkip_InFlightOrderTicks(t *testing.T) {
	h := testharness.NewHarness(t, testDataDir)
	locked := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	locked.ActiveGoal = newLockSkipGoal()

	// Make the default def movement-eligible (Effort 3 Phase 5 will set this
	// in TOML; for now the test injects Speed > 0 directly).
	h.Engine.Rulebook.Assets[testharness.DefSecurityPersonnel].Speed = 2

	asset := locked.Assets["alpha-asset-1"]
	path := []spatial.RegionHex{
		{RegionID: "void", Coord: spatial.HexCoord{Q: 0, R: 0}},
		{RegionID: "void", Coord: spatial.HexCoord{Q: 1, R: 0}},
		{RegionID: "void", Coord: spatial.HexCoord{Q: 2, R: 0}},
		{RegionID: "void", Coord: spatial.HexCoord{Q: 3, R: 0}},
		{RegionID: "void", Coord: spatial.HexCoord{Q: 4, R: 0}},
	}
	asset.CurrentOrder = &domain.MovementOrder{
		AssetID:     asset.ID,
		Origin:      domain.Location{RegionHex: path[0]},
		Destination: domain.Location{WorldID: "Krylos", RegionHex: path[4]},
		Path:        path,
		StepIdx:     1,
		DriftRating: 0,
	}
	asset.Location = domain.Location{RegionHex: path[1]}

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Cfg, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	tickedAsset := h.FactionState.Factions["alpha"].Assets["alpha-asset-1"]
	if tickedAsset.CurrentOrder == nil {
		t.Fatalf("alpha-asset-1 CurrentOrder cleared unexpectedly; order should still be in flight")
	}
	if got, want := tickedAsset.CurrentOrder.StepIdx, 3; got != want {
		t.Errorf("alpha-asset-1 StepIdx after tick: got %d, want %d", got, want)
	}
	if got, want := tickedAsset.Location.RegionHex.Coord, (spatial.HexCoord{Q: 3, R: 0}); got != want {
		t.Errorf("alpha-asset-1 hex coord after tick: got %+v, want %+v", got, want)
	}

	records := testharness.ReadHistory(t, h.Cfg.HistoryPath)
	if _, ok := testharness.FindMutationByTypeAndCause(records, "movement_order_progressed", domain.CauseMovementTick); !ok {
		t.Fatalf("history missing movement_order_progressed with cause %q", domain.CauseMovementTick)
	}

	assertActionPhaseSkipped(t, h.Observer, "alpha")
}

// Scenario 3: LockSkip faction issuing a new order on a fresh asset. The
// Movement Phase emits MovementOrderIssued; no CoinDelta is emitted (issuance
// is free — only revisions cost). The action phase is still skipped.
func TestLockSkip_NewIssueIsFree(t *testing.T) {
	h := testharness.NewHarness(t, testDataDir)
	locked := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	locked.ActiveGoal = newLockSkipGoal()

	h.Engine.Rulebook.Assets[testharness.DefSecurityPersonnel].Speed = 2

	h.Collector.SelectMovementDecisionsFn = func(faction *domain.Faction, eligible []*domain.Asset) ([]world.MovementDecision, error) {
		if faction.ID != "alpha" || len(eligible) == 0 {
			return nil, nil
		}
		return []world.MovementDecision{{
			AssetID:     eligible[0].ID,
			Kind:        world.MovementDecisionIssue,
			Destination: &domain.Location{WorldID: "Krylos"},
		}}, nil
	}

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Cfg, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	mutations := movementMutationsFor(t, h.Observer, "alpha")
	var sawIssue, sawCoinDelta bool
	for _, mutation := range mutations {
		switch mutation.(type) {
		case domain.MovementOrderIssued:
			sawIssue = true
		case domain.CoinDelta:
			sawCoinDelta = true
		}
	}
	if !sawIssue {
		t.Errorf("expected MovementOrderIssued in movement-phase mutations, got %d mutations of other types", len(mutations))
	}
	if sawCoinDelta {
		t.Errorf("issue should be free — CoinDelta emitted during movement phase")
	}
	assertActionPhaseSkipped(t, h.Observer, "alpha")
}

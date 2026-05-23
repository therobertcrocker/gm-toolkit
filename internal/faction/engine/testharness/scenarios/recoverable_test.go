package scenarios

import (
	"errors"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

// TestRunFactionTurn_RecoverableErrNoSelection drives the orchestrator's
// recoverable path: Action.Run returns ErrNoSelection, the turn must continue
// (nil return) and fire TurnCompleted without firing ActionResolved.
func TestRunFactionTurn_RecoverableErrNoSelection(t *testing.T) {
	h := testharness.NewHarness(t, testDataDir)
	h.AddFaction("alpha", "Tartarus", 4, 3, 2)

	h.Collector.SelectActionFn = func(_ *domain.Faction, available []action.Action) (action.Action, error) {
		for _, a := range available {
			if a.Name() == "Sell Asset" {
				return a, nil
			}
		}
		t.Fatal("Sell Asset not in available actions")
		return nil, nil
	}
	h.Collector.SelectAssetFn = func(_ []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
		return nil, nil // triggers ErrNoSelection in SellAsset.Inputs
	}

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	_, err := h.Engine.RunFactionTurn(h.FactionState, h.Cfg, h.Collectors, h.Observer)
	if err != nil {
		t.Fatalf("RunFactionTurn returned error, want nil: %v", err)
	}

	hasErrorEvent := false
	for _, e := range h.Observer.Events {
		if e.Kind == "Error" {
			if payload, ok := e.Payload.(error); ok && errors.Is(payload, action.ErrNoSelection) {
				hasErrorEvent = true
				break
			}
		}
	}
	if !hasErrorEvent {
		t.Error("expected Error event wrapping ErrNoSelection, not found")
	}

	for _, e := range h.Observer.Events {
		if e.Kind == "ActionResolved" {
			t.Error("unexpected ActionResolved event: action should have been skipped")
		}
	}

	hasTurnCompleted := false
	for _, e := range h.Observer.Events {
		if e.Kind == "TurnCompleted" {
			hasTurnCompleted = true
			break
		}
	}
	if !hasTurnCompleted {
		t.Error("expected TurnCompleted event: finishFactionTurn should have run")
	}
}

package mutation

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func newMutationTestState() *state.FactionState {
	return &state.FactionState{
		Factions: map[string]*domain.Faction{
			"f1": {
				ID:   "f1",
				Name: "Faction One",
				Coin: 10,
				Assets: map[string]*domain.Asset{
					"a1": {ID: "a1", DefinitionID: "infantry", Location: domain.Location{WorldID: "Anchorage"}, Maintained: true},
					"a2": {ID: "a2", DefinitionID: "spy_net", Location: domain.Location{WorldID: "Anchorage"}, Maintained: false},
				},
			},
		},
	}
}

func TestMutationEngine_CoinDelta(t *testing.T) {
	s := newMutationTestState()
	me := New()

	if err := me.Apply(s, []domain.Mutation{
		domain.CoinDelta{FactionID: "f1", Delta: 5},
	}); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if got := s.Factions["f1"].Coin; got != 15 {
		t.Errorf("Coin = %d, want 15", got)
	}

	if err := me.Apply(s, []domain.Mutation{
		domain.CoinDelta{FactionID: "f1", Delta: -3},
	}); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if got := s.Factions["f1"].Coin; got != 12 {
		t.Errorf("Coin = %d, want 12", got)
	}
}

func TestMutationEngine_AssetRemoved(t *testing.T) {
	s := newMutationTestState()
	me := New()

	if err := me.Apply(s, []domain.Mutation{
		domain.AssetRemoved{FactionID: "f1", AssetID: "a1"},
	}); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}

	assets := s.Factions["f1"].Assets
	if len(assets) != 1 {
		t.Fatalf("len(Assets) = %d, want 1", len(assets))
	}
	if assets["a2"].ID != "a2" {
		t.Errorf("remaining asset ID = %q, want %q", assets["a2"].ID, "a2")
	}
}

func TestMutationEngine_AssetMaintainedFlag(t *testing.T) {
	s := newMutationTestState()
	me := New()

	if err := me.Apply(s, []domain.Mutation{
		domain.AssetMaintainedFlag{FactionID: "f1", AssetID: "a1", Maintained: false},
	}); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if s.Factions["f1"].Assets["a1"].Maintained {
		t.Error("expected a1 to be unmaintained")
	}

	if err := me.Apply(s, []domain.Mutation{
		domain.AssetMaintainedFlag{FactionID: "f1", AssetID: "a2", Maintained: true},
	}); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if !s.Factions["f1"].Assets["a2"].Maintained {
		t.Error("expected a2 to be maintained")
	}
}

func TestMutationEngine_UnknownFactionErrors(t *testing.T) {
	s := newMutationTestState()
	me := New()

	err := me.Apply(s, []domain.Mutation{
		domain.CoinDelta{FactionID: "missing", Delta: 100},
	})
	if err == nil {
		t.Fatal("Apply() expected error for unknown faction, got nil")
	}
	if len(err.Misses) != 1 {
		t.Fatalf("len(Misses) = %d, want 1", len(err.Misses))
	}
	if got := s.Factions["f1"].Coin; got != 10 {
		t.Errorf("Coin = %d, want 10 (unrelated faction must be unmodified)", got)
	}
}

func TestMutationEngine_AssetStealthCleared(t *testing.T) {
	s := newMutationTestState()
	s.Factions["f1"].Assets["a1"].Stealthy = true
	me := New()

	if err := me.Apply(s, []domain.Mutation{
		domain.AssetStealthCleared{FactionID: "f1", AssetID: "a1"},
	}); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if s.Factions["f1"].Assets["a1"].Stealthy {
		t.Error("expected a1 Stealthy to be false after AssetStealthCleared")
	}
}

func newMutationTestStateWithBase() *state.FactionState {
	s := newMutationTestState()
	s.Factions["f1"].Bases = []*domain.Base{
		{ID: "b1", OwnerID: "f1", Location: domain.Location{WorldID: "Anchorage"}, CurrentHP: 10, MaxHP: 10},
	}
	return s
}

func TestMutationEngine_BaseHPDelta(t *testing.T) {
	s := newMutationTestStateWithBase()
	me := New()

	if err := me.Apply(s, []domain.Mutation{
		domain.BaseHPDelta{FactionID: "f1", BaseID: "b1", Delta: -4},
	}); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if got := s.Factions["f1"].Bases[0].CurrentHP; got != 6 {
		t.Errorf("Base CurrentHP = %d, want 6", got)
	}
}

func TestMutationEngine_BaseDestroyed(t *testing.T) {
	s := newMutationTestStateWithBase()
	me := New()

	if err := me.Apply(s, []domain.Mutation{
		domain.BaseDestroyed{FactionID: "f1", BaseID: "b1"},
	}); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if got := len(s.Factions["f1"].Bases); got != 0 {
		t.Errorf("len(Bases) = %d, want 0 after BaseDestroyed", got)
	}
}

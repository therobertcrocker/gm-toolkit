package engine

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func newMutationTestState() *state.FactionState {
	return &state.FactionState{
		Factions: []*domain.Faction{
			{
				ID:   "f1",
				Name: "Faction One",
				Coin: 10,
				Assets: []*domain.Asset{
					{ID: "a1", DefinitionID: "infantry", Location: "Anchorage", Maintained: true},
					{ID: "a2", DefinitionID: "spy_net", Location: "Anchorage", Maintained: false},
				},
			},
		},
	}
}

func TestMutationEngine_CoinDelta(t *testing.T) {
	s := newMutationTestState()
	me := newMutationEngine()

	me.Apply(s, []domain.Mutation{
		domain.CoinDelta{FactionID: "f1", Delta: 5},
	})
	if got := s.Factions[0].Coin; got != 15 {
		t.Errorf("Coin = %d, want 15", got)
	}

	me.Apply(s, []domain.Mutation{
		domain.CoinDelta{FactionID: "f1", Delta: -3},
	})
	if got := s.Factions[0].Coin; got != 12 {
		t.Errorf("Coin = %d, want 12", got)
	}
}

func TestMutationEngine_AssetRemoved(t *testing.T) {
	s := newMutationTestState()
	me := newMutationEngine()

	me.Apply(s, []domain.Mutation{
		domain.AssetRemoved{FactionID: "f1", AssetID: "a1"},
	})

	assets := s.Factions[0].Assets
	if len(assets) != 1 {
		t.Fatalf("len(Assets) = %d, want 1", len(assets))
	}
	if assets[0].ID != "a2" {
		t.Errorf("remaining asset ID = %q, want %q", assets[0].ID, "a2")
	}
}

func TestMutationEngine_AssetMaintainedFlag(t *testing.T) {
	s := newMutationTestState()
	me := newMutationEngine()

	// mark the maintained asset as unmaintained
	me.Apply(s, []domain.Mutation{
		domain.AssetMaintainedFlag{FactionID: "f1", AssetID: "a1", Maintained: false},
	})
	if s.Factions[0].Assets[0].Maintained {
		t.Error("expected a1 to be unmaintained")
	}

	// mark the unmaintained asset as maintained
	me.Apply(s, []domain.Mutation{
		domain.AssetMaintainedFlag{FactionID: "f1", AssetID: "a2", Maintained: true},
	})
	if !s.Factions[0].Assets[1].Maintained {
		t.Error("expected a2 to be maintained")
	}
}

func TestMutationEngine_UnknownFactionIsNoop(t *testing.T) {
	s := newMutationTestState()
	me := newMutationEngine()

	me.Apply(s, []domain.Mutation{
		domain.CoinDelta{FactionID: "missing", Delta: 100},
	})
	if got := s.Factions[0].Coin; got != 10 {
		t.Errorf("Coin = %d, want 10 (unknown faction should be no-op)", got)
	}
}

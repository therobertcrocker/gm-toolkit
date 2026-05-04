package actions

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"go.uber.org/mock/gomock"
)

func TestSeizePlanet_Validate(t *testing.T) {
	t.Run("no active goal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		faction := &domain.Faction{ID: "f1"}
		factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}
		if NewSeizePlanet(mocks.NewMockCollector(ctrl)).Validate(faction, factionState, nil) {
			t.Error("expected false when no active goal")
		}
	})

	t.Run("wrong goal type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		asset := &domain.Asset{ID: "a1", OwnerID: "f1", Location: "Krylos"}
		faction := &domain.Faction{
			ID:         "f1",
			ActiveGoal: &domain.ActiveGoal{GoalID: "G-001", ProcessPhase: 0},
			Assets:     []*domain.Asset{asset},
		}
		factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}
		if NewSeizePlanet(mocks.NewMockCollector(ctrl)).Validate(faction, factionState, nil) {
			t.Error("expected false when goal ID is not G-004")
		}
	})

	t.Run("no contested world", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		asset := &domain.Asset{ID: "a1", OwnerID: "f1", Location: "Krylos"}
		faction := &domain.Faction{
			ID:         "f1",
			ActiveGoal: &domain.ActiveGoal{GoalID: "G-004", ProcessPhase: 0},
			Assets:     []*domain.Asset{asset},
		}
		factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}
		if NewSeizePlanet(mocks.NewMockCollector(ctrl)).Validate(faction, factionState, nil) {
			t.Error("expected false when no rival assets share a world with the faction")
		}
	})

	t.Run("valid contested world", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		f1Asset := &domain.Asset{ID: "a1", OwnerID: "f1", Location: "Krylos"}
		f2Asset := &domain.Asset{ID: "a2", OwnerID: "f2", Location: "Krylos"}
		faction := &domain.Faction{
			ID:         "f1",
			ActiveGoal: &domain.ActiveGoal{GoalID: "G-004", ProcessPhase: 0},
			Assets:     []*domain.Asset{f1Asset},
		}
		rival := &domain.Faction{ID: "f2", Assets: []*domain.Asset{f2Asset}}
		factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction, "f2": rival}}
		if !NewSeizePlanet(mocks.NewMockCollector(ctrl)).Validate(faction, factionState, nil) {
			t.Error("expected true when faction and rival share a world")
		}
	})
}

// TestSeizePlanet_Output: emits GoalInitiated{G-004, targetWorld, ProcessPhase=1}.
func TestSeizePlanet_Output(t *testing.T) {
	ctrl := gomock.NewController(t)
	f1Asset := &domain.Asset{ID: "a1", OwnerID: "f1", Location: "Krylos"}
	f2Asset := &domain.Asset{ID: "a2", OwnerID: "f2", Location: "Krylos"}
	faction := &domain.Faction{
		ID:         "f1",
		ActiveGoal: &domain.ActiveGoal{GoalID: "G-004", ProcessPhase: 0},
		Assets:     []*domain.Asset{f1Asset},
	}
	rival := &domain.Faction{ID: "f2", Assets: []*domain.Asset{f2Asset}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction, "f2": rival}}

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectSeizeTarget(gomock.Any(), gomock.Any()).Return("Krylos", nil)

	act := NewSeizePlanet(collector)
	if err := act.Inputs(faction, factionState, nil); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, factionState, nil); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}

	if len(mutations) != 1 {
		t.Fatalf("len(mutations) = %d, want 1; got %v", len(mutations), mutations)
	}
	initiated, ok := mutations[0].(domain.GoalInitiated)
	if !ok {
		t.Fatalf("mutations[0] type = %T, want GoalInitiated", mutations[0])
	}
	if initiated.GoalID != "G-004" || initiated.TargetWorld != "Krylos" || initiated.ProcessPhase != 1 {
		t.Errorf("GoalInitiated = %+v, want {GoalID:G-004, TargetWorld:Krylos, ProcessPhase:1}", initiated)
	}
}

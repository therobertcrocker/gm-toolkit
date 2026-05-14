package goals

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func makeLockRulebook() *rulebook.Rulebook {
	return &rulebook.Rulebook{
		Tags: map[string]*domain.Tag{
			"T-011": {ID: "T-011", Name: "Planetary Government"},
		},
	}
}

func TestChangeHomeworld_InTransit(t *testing.T) {
	faction := &domain.Faction{
		ID:        "f1",
		Homeworld: "Tartarus",
		ActiveGoal: &domain.ActiveGoal{
			GoalID:         "G-012",
			TargetWorld:    "Krylos",
			TurnsRemaining: 2,
		},
	}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	lock, mutations := ChangeHomeworld{}.CheckLock(faction, factionState, makeLockRulebook())

	if lock.Type != locks.LockSkip {
		t.Errorf("LockType = %v, want LockSkip", lock.Type)
	}
	if len(mutations) != 1 {
		t.Fatalf("len(mutations) = %d, want 1; got %v", len(mutations), mutations)
	}
	tick, ok := mutations[0].(domain.GoalTurnsTick)
	if !ok {
		t.Fatalf("mutations[0] type = %T, want GoalTurnsTick", mutations[0])
	}
	if tick.FactionID != "f1" || tick.GoalID != "G-012" {
		t.Errorf("GoalTurnsTick = %+v, want {FactionID: f1, GoalID: G-012}", tick)
	}
}

func TestChangeHomeworld_Completing(t *testing.T) {
	faction := &domain.Faction{
		ID:        "f1",
		Homeworld: "Tartarus",
		ActiveGoal: &domain.ActiveGoal{
			GoalID:         "G-012",
			TargetWorld:    "Krylos",
			TurnsRemaining: 1,
		},
	}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	lock, mutations := ChangeHomeworld{}.CheckLock(faction, factionState, makeLockRulebook())

	if lock.Type != locks.LockSkip {
		t.Errorf("LockType = %v, want LockSkip", lock.Type)
	}
	// GoalTurnsTick, HomeworldChanged, GoalCompleted
	if len(mutations) != 3 {
		t.Fatalf("len(mutations) = %d, want 3; got %v", len(mutations), mutations)
	}
	if _, ok := mutations[0].(domain.GoalTurnsTick); !ok {
		t.Errorf("mutations[0] type = %T, want GoalTurnsTick", mutations[0])
	}
	hw, ok := mutations[1].(domain.HomeworldChanged)
	if !ok {
		t.Fatalf("mutations[1] type = %T, want HomeworldChanged", mutations[1])
	}
	if hw.FromWorld != "Tartarus" || hw.ToWorld != "Krylos" {
		t.Errorf("HomeworldChanged = {%s → %s}, want {Tartarus → Krylos}", hw.FromWorld, hw.ToWorld)
	}
	if _, ok := mutations[2].(domain.GoalCompleted); !ok {
		t.Errorf("mutations[2] type = %T, want GoalCompleted", mutations[2])
	}
}

func TestPlanetarySeizure_Phase1(t *testing.T) {
	faction := &domain.Faction{
		ID: "f1",
		ActiveGoal: &domain.ActiveGoal{
			GoalID:       "G-004",
			TargetWorld:  "Krylos",
			ProcessPhase: 1,
		},
	}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	lock, mutations := PlanetarySeizure{}.CheckLock(faction, factionState, makeLockRulebook())

	if lock.Type != locks.LockRestrictActions {
		t.Errorf("LockType = %v, want LockRestrictActions", lock.Type)
	}
	if len(lock.AllowedActions) != 1 || lock.AllowedActions[0] != "Attack" {
		t.Errorf("AllowedActions = %v, want [Attack]", lock.AllowedActions)
	}
	if len(mutations) != 0 {
		t.Errorf("mutations = %v, want nil", mutations)
	}
}

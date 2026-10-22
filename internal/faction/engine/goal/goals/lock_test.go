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
		Homeworld: domain.Location{WorldID: "Tartarus"},
		ActiveGoal: &domain.ActiveGoal{
			GoalID:         "G-012",
			TargetWorld:    domain.Location{WorldID: "Krylos"},
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
		Homeworld: domain.Location{WorldID: "Tartarus"},
		ActiveGoal: &domain.ActiveGoal{
			GoalID:         "G-012",
			TargetWorld:    domain.Location{WorldID: "Krylos"},
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
	if hw.FromWorld.WorldID != "Tartarus" || hw.ToWorld.WorldID != "Krylos" {
		t.Errorf("HomeworldChanged = {%s → %s}, want {Tartarus → Krylos}", hw.FromWorld.WorldID, hw.ToWorld.WorldID)
	}
	if _, ok := mutations[2].(domain.GoalCompleted); !ok {
		t.Errorf("mutations[2] type = %T, want GoalCompleted", mutations[2])
	}
}

// makeSeizureState builds a phase-1 Planetary Seizure on Krylos for f1, with
// rival f2 holding the given assets.
func makeSeizureState(rivalAssets map[string]*domain.Asset) (*state.FactionState, *domain.Faction) {
	faction := &domain.Faction{
		ID: "f1",
		ActiveGoal: &domain.ActiveGoal{
			GoalID:       "G-004",
			TargetWorld:  domain.Location{WorldID: "Krylos"},
			ProcessPhase: 1,
		},
	}
	rival := &domain.Faction{ID: "f2", Assets: rivalAssets}
	return &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction, "f2": rival}}, faction
}

func TestPlanetarySeizure_Phase1(t *testing.T) {
	factionState, faction := makeSeizureState(map[string]*domain.Asset{
		"r1": {ID: "r1", OwnerID: "f2", Location: domain.Location{WorldID: "Krylos"}},
	})

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

func TestPlanetarySeizure_Phase1_NoRivalAdvances(t *testing.T) {
	cases := []struct {
		name        string
		rivalAssets map[string]*domain.Asset
	}{
		{"empty world", nil},
		{"stealthed rival", map[string]*domain.Asset{
			"r1": {ID: "r1", OwnerID: "f2", Location: domain.Location{WorldID: "Krylos"}, Stealthy: true},
		}},
		{"sheltered rival", map[string]*domain.Asset{
			"r1": {ID: "r1", OwnerID: "f2", Location: domain.Location{WorldID: "Krylos"},
				Conditions: []domain.Condition{{Kind: domain.ConditionSheltered, SourceAssetID: "bolt"}}},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState, faction := makeSeizureState(tc.rivalAssets)

			lock, mutations := PlanetarySeizure{}.CheckLock(faction, factionState, makeLockRulebook())

			if lock.Type != locks.LockNone {
				t.Errorf("LockType = %v, want LockNone", lock.Type)
			}
			if len(mutations) != 1 {
				t.Fatalf("len(mutations) = %d, want 1; got %v", len(mutations), mutations)
			}
			advanced, ok := mutations[0].(domain.GoalPhaseAdvanced)
			if !ok {
				t.Fatalf("mutations[0] type = %T, want GoalPhaseAdvanced", mutations[0])
			}
			if advanced.ProcessPhase != 2 || advanced.TurnsRemaining != 3 {
				t.Errorf("GoalPhaseAdvanced = {phase %d, turns %d}, want {2, 3}", advanced.ProcessPhase, advanced.TurnsRemaining)
			}
		})
	}
}

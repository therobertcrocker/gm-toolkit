package goal

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func TestCheckLock_NoActiveGoal(t *testing.T) {
	ge := New()
	faction := &domain.Faction{ID: "f1"}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	lock, mutations := ge.CheckLock(faction, factionState, &rulebook.Rulebook{})

	if lock.Type != locks.LockNone {
		t.Errorf("LockType = %v, want LockNone", lock.Type)
	}
	if len(mutations) != 0 {
		t.Errorf("mutations = %v, want nil", mutations)
	}
}

// TestCheckLock_DataOnlyGoal_NoHandler exercises the Shape 2 data-only
// fall-through: a faction's ActiveGoal references a goal ID that exists in
// the rulebook's goals.toml but has no registered Go handler. The contract
// is that CheckLock returns LockNone with no mutations (data-only goals
// no-op rather than panic, mirroring how tags handle unregistered IDs).
func TestCheckLock_DataOnlyGoal_NoHandler(t *testing.T) {
	ge := New()
	faction := &domain.Faction{ID: "f1", ActiveGoal: &domain.ActiveGoal{GoalID: "g1"}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	lock, mutations := ge.CheckLock(faction, factionState, &rulebook.Rulebook{})

	if lock.Type != locks.LockNone {
		t.Errorf("LockType = %v, want LockNone", lock.Type)
	}
	if len(mutations) != 0 {
		t.Errorf("mutations = %v, want nil", mutations)
	}
}

// TestUpdateProgress_DataOnlyGoal_NoHandler is the UpdateProgress counterpart
// to TestCheckLock_DataOnlyGoal_NoHandler: when a faction's ActiveGoal.GoalID
// has no registered handler, UpdateProgress must return nil rather than
// panicking on the missing key.
func TestUpdateProgress_DataOnlyGoal_NoHandler(t *testing.T) {
	ge := New()
	faction := &domain.Faction{ID: "f1", ActiveGoal: &domain.ActiveGoal{GoalID: "g1"}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	mutations := ge.UpdateProgress("f1", nil, factionState, &rulebook.Rulebook{}, nil)

	if mutations != nil {
		t.Errorf("mutations = %v, want nil", mutations)
	}
}

package goal

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func makeProgressRulebook() *rulebook.Rulebook {
	return &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			"force-asset-def": {
				ID:       "force-asset-def",
				Category: domain.StatForce,
			},
			"wealth-asset-def": {
				ID:       "wealth-asset-def",
				Category: domain.StatWealth,
			},
			"cunning-asset-def": {
				ID:       "cunning-asset-def",
				Category: domain.StatCunning,
			},
		},
	}
}

// killMutation builds an AssetRemoved that looks like an attack kill of assetID on rivalFactionID.
func killMutation(actingFactionID, rivalFactionID, assetID string) domain.AssetRemoved {
	return domain.AssetRemoved{
		FactionID:         rivalFactionID,
		AssetID:           assetID,
		Cause:             "attack",
		CausedByFactionID: actingFactionID,
	}
}

// makeMilitaryConquestState builds a minimal FactionState for G-001 tests.
// f1 is the acting faction with the given Force rating and goal progress.
// f2 is the rival with one Force asset in its Assets slice.
func makeMilitaryConquestState(f1Force, goalProgress int) (*state.FactionState, *domain.Faction) {
	rivalAsset := &domain.Asset{ID: "d1", DefinitionID: "force-asset-def", OwnerID: "f2"}
	acting := &domain.Faction{
		ID:    "f1",
		Force: f1Force,
		ActiveGoal: &domain.ActiveGoal{
			GoalID:   "G-001",
			Progress: goalProgress,
		},
	}
	rival := &domain.Faction{
		ID:     "f2",
		Assets: []*domain.Asset{rivalAsset},
	}
	factionState := &state.FactionState{
		Factions: map[string]*domain.Faction{"f1": acting, "f2": rival},
	}
	return factionState, acting
}

func TestProgressMilitaryConquest_NoKills(t *testing.T) {
	factionState, acting := makeMilitaryConquestState(3, 0)
	rb := makeProgressRulebook()
	mutations := progressMilitaryConquest(acting, nil, factionState, rb)
	if len(mutations) != 0 {
		t.Errorf("expected no mutations for no kills, got %v", mutations)
	}
}

func TestProgressMilitaryConquest_KillBelowThreshold(t *testing.T) {
	factionState, acting := makeMilitaryConquestState(3, 0)
	rb := makeProgressRulebook()

	input := []domain.Mutation{killMutation("f1", "f2", "d1")}
	mutations := progressMilitaryConquest(acting, input, factionState, rb)

	if len(mutations) != 1 {
		t.Fatalf("len(mutations) = %d, want 1; got %v", len(mutations), mutations)
	}
	prog, ok := mutations[0].(domain.GoalProgressed)
	if !ok {
		t.Fatalf("mutations[0] type = %T, want GoalProgressed", mutations[0])
	}
	if prog.Delta != 1 {
		t.Errorf("GoalProgressed.Delta = %d, want 1", prog.Delta)
	}
}

func TestProgressMilitaryConquest_KillCompletesGoal(t *testing.T) {
	// Force=1 means a single kill (progress 0+1 >= 1) triggers completion.
	factionState, acting := makeMilitaryConquestState(1, 0)
	rb := makeProgressRulebook()

	input := []domain.Mutation{killMutation("f1", "f2", "d1")}
	mutations := progressMilitaryConquest(acting, input, factionState, rb)

	// GoalProgressed + GoalCompleted + XPAwarded
	if len(mutations) != 3 {
		t.Fatalf("len(mutations) = %d, want 3; got %v", len(mutations), mutations)
	}
	if _, ok := mutations[0].(domain.GoalProgressed); !ok {
		t.Errorf("mutations[0] type = %T, want GoalProgressed", mutations[0])
	}
	if _, ok := mutations[1].(domain.GoalCompleted); !ok {
		t.Errorf("mutations[1] type = %T, want GoalCompleted", mutations[1])
	}
	if _, ok := mutations[2].(domain.XPAwarded); !ok {
		t.Errorf("mutations[2] type = %T, want XPAwarded", mutations[2])
	}
}

func TestProgressCommercialExpansion_NoKills(t *testing.T) {
	rivalAsset := &domain.Asset{ID: "d1", DefinitionID: "wealth-asset-def", OwnerID: "f2"}
	acting := &domain.Faction{
		ID:     "f1",
		Wealth: 2,
		ActiveGoal: &domain.ActiveGoal{GoalID: "G-002", Progress: 0},
	}
	rival := &domain.Faction{ID: "f2", Assets: []*domain.Asset{rivalAsset}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": acting, "f2": rival}}

	mutations := progressCommercialExpansion(acting, nil, factionState, makeProgressRulebook())
	if len(mutations) != 0 {
		t.Errorf("expected no mutations for no kills, got %v", mutations)
	}
}

func TestProgressCommercialExpansion_KillCompletesGoal(t *testing.T) {
	// Wealth=1 means one Wealth kill completes the goal.
	rivalAsset := &domain.Asset{ID: "d1", DefinitionID: "wealth-asset-def", OwnerID: "f2"}
	acting := &domain.Faction{
		ID:     "f1",
		Wealth: 1,
		ActiveGoal: &domain.ActiveGoal{GoalID: "G-002", Progress: 0},
	}
	rival := &domain.Faction{ID: "f2", Assets: []*domain.Asset{rivalAsset}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": acting, "f2": rival}}

	input := []domain.Mutation{killMutation("f1", "f2", "d1")}
	mutations := progressCommercialExpansion(acting, input, factionState, makeProgressRulebook())

	if len(mutations) != 3 {
		t.Fatalf("len(mutations) = %d, want 3; got %v", len(mutations), mutations)
	}
	if _, ok := mutations[1].(domain.GoalCompleted); !ok {
		t.Errorf("mutations[1] type = %T, want GoalCompleted", mutations[1])
	}
}

func TestProgressIntelligenceCoup_NoKills(t *testing.T) {
	rivalAsset := &domain.Asset{ID: "d1", DefinitionID: "cunning-asset-def", OwnerID: "f2"}
	acting := &domain.Faction{
		ID:      "f1",
		Cunning: 2,
		ActiveGoal: &domain.ActiveGoal{GoalID: "G-003", Progress: 0},
	}
	rival := &domain.Faction{ID: "f2", Assets: []*domain.Asset{rivalAsset}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": acting, "f2": rival}}

	mutations := progressIntelligenceCoup(acting, nil, factionState, makeProgressRulebook())
	if len(mutations) != 0 {
		t.Errorf("expected no mutations for no kills, got %v", mutations)
	}
}

func TestProgressIntelligenceCoup_KillBelowThreshold(t *testing.T) {
	rivalAsset := &domain.Asset{ID: "d1", DefinitionID: "cunning-asset-def", OwnerID: "f2"}
	acting := &domain.Faction{
		ID:      "f1",
		Cunning: 3,
		ActiveGoal: &domain.ActiveGoal{GoalID: "G-003", Progress: 0},
	}
	rival := &domain.Faction{ID: "f2", Assets: []*domain.Asset{rivalAsset}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": acting, "f2": rival}}

	input := []domain.Mutation{killMutation("f1", "f2", "d1")}
	mutations := progressIntelligenceCoup(acting, input, factionState, makeProgressRulebook())

	if len(mutations) != 1 {
		t.Fatalf("len(mutations) = %d, want 1; got %v", len(mutations), mutations)
	}
	prog, ok := mutations[0].(domain.GoalProgressed)
	if !ok {
		t.Fatalf("mutations[0] type = %T, want GoalProgressed", mutations[0])
	}
	if prog.Delta != 1 {
		t.Errorf("GoalProgressed.Delta = %d, want 1", prog.Delta)
	}
}

func TestProgressIntelligenceCoup_KillCompletesGoal(t *testing.T) {
	rivalAsset := &domain.Asset{ID: "d1", DefinitionID: "cunning-asset-def", OwnerID: "f2"}
	acting := &domain.Faction{
		ID:      "f1",
		Cunning: 1,
		ActiveGoal: &domain.ActiveGoal{GoalID: "G-003", Progress: 0},
	}
	rival := &domain.Faction{ID: "f2", Assets: []*domain.Asset{rivalAsset}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": acting, "f2": rival}}

	input := []domain.Mutation{killMutation("f1", "f2", "d1")}
	mutations := progressIntelligenceCoup(acting, input, factionState, makeProgressRulebook())

	if len(mutations) != 3 {
		t.Fatalf("len(mutations) = %d, want 3; got %v", len(mutations), mutations)
	}
	if _, ok := mutations[1].(domain.GoalCompleted); !ok {
		t.Errorf("mutations[1] type = %T, want GoalCompleted", mutations[1])
	}
}

func TestProgressMilitaryConquest_WrongCategory_NoProgress(t *testing.T) {
	// Rival has a Wealth asset, not Force — should not count toward MilitaryConquest.
	rivalAsset := &domain.Asset{ID: "d1", DefinitionID: "wealth-asset-def", OwnerID: "f2"}
	acting := &domain.Faction{
		ID:    "f1",
		Force: 3,
		ActiveGoal: &domain.ActiveGoal{GoalID: "G-001", Progress: 0},
	}
	rival := &domain.Faction{ID: "f2", Assets: []*domain.Asset{rivalAsset}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": acting, "f2": rival}}

	input := []domain.Mutation{killMutation("f1", "f2", "d1")}
	mutations := progressMilitaryConquest(acting, input, factionState, makeProgressRulebook())
	if len(mutations) != 0 {
		t.Errorf("expected no mutations for wrong asset category, got %v", mutations)
	}
}

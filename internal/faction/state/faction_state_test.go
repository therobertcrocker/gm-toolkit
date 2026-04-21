package state

import (
	"path/filepath"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func TestFactionStateRoundTrip(t *testing.T) {
	original := &FactionState{
		CampaignID: "test-campaign",
		TurnNumber: 3,
		Factions: []*domain.Faction{
			{
				ID:        "iron-collective",
				Name:      "Iron Collective",
				Scale:     domain.ScaleMajor,
				Force:     6,
				Cunning:   5,
				Wealth:    3,
				CurrentHP: 28,
				MaxHP:     28,
				Coin:      5,
				XP:        2,
				Homeworld: "Tartarus",
				Tags: []*domain.Tag{
					{ID: "T-001", Name: "Colonists", Description: "Settler faction", Effect: "+1 Wealth"},
				},
				Goal: &domain.Goal{
					ID:         "G-001",
					Name:       "Military Conquest",
					Description: "Destroy half the enemy's assets",
					Difficulty: "half_assets_destroyed",
				},
				Assets: []*domain.Asset{
					{
						ID:           "iron-collective-asset-0",
						DefinitionID: "F1-001",
						OwnerID:      "iron-collective",
						Location:     "Tartarus",
						CurrentHP:    3,
						Stealthy:     false,
						Ready:        true,
						Maintained:   true,
					},
				},
			},
		},
	}

	path := filepath.Join(t.TempDir(), "faction_state.toml")

	if err := Save(path, original); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.CampaignID != original.CampaignID {
		t.Errorf("CampaignID: got %q, want %q", loaded.CampaignID, original.CampaignID)
	}
	if loaded.TurnNumber != original.TurnNumber {
		t.Errorf("TurnNumber: got %d, want %d", loaded.TurnNumber, original.TurnNumber)
	}
	if len(loaded.Factions) != 1 {
		t.Fatalf("Factions: got %d, want 1", len(loaded.Factions))
	}

	got := loaded.Factions[0]
	want := original.Factions[0]

	if got.ID != want.ID {
		t.Errorf("Faction.ID: got %q, want %q", got.ID, want.ID)
	}
	if got.Name != want.Name {
		t.Errorf("Faction.Name: got %q, want %q", got.Name, want.Name)
	}
	if got.Scale != want.Scale {
		t.Errorf("Faction.Scale: got %q, want %q", got.Scale, want.Scale)
	}
	if got.Force != want.Force {
		t.Errorf("Faction.Force: got %d, want %d", got.Force, want.Force)
	}
	if got.Cunning != want.Cunning {
		t.Errorf("Faction.Cunning: got %d, want %d", got.Cunning, want.Cunning)
	}
	if got.Wealth != want.Wealth {
		t.Errorf("Faction.Wealth: got %d, want %d", got.Wealth, want.Wealth)
	}
	if got.CurrentHP != want.CurrentHP {
		t.Errorf("Faction.CurrentHP: got %d, want %d", got.CurrentHP, want.CurrentHP)
	}
	if got.MaxHP != want.MaxHP {
		t.Errorf("Faction.MaxHP: got %d, want %d", got.MaxHP, want.MaxHP)
	}
	if got.Coin != want.Coin {
		t.Errorf("Faction.Coin: got %d, want %d", got.Coin, want.Coin)
	}
	if got.XP != want.XP {
		t.Errorf("Faction.XP: got %d, want %d", got.XP, want.XP)
	}
	if got.Homeworld != want.Homeworld {
		t.Errorf("Faction.Homeworld: got %q, want %q", got.Homeworld, want.Homeworld)
	}

	if len(got.Tags) != 1 {
		t.Fatalf("Faction.Tags: got %d, want 1", len(got.Tags))
	}
	if got.Tags[0].ID != want.Tags[0].ID || got.Tags[0].Name != want.Tags[0].Name {
		t.Errorf("Faction.Tags[0]: got %+v, want %+v", got.Tags[0], want.Tags[0])
	}

	if got.Goal == nil {
		t.Fatal("Faction.Goal: got nil, want non-nil")
	}
	if got.Goal.ID != want.Goal.ID || got.Goal.Name != want.Goal.Name || got.Goal.Difficulty != want.Goal.Difficulty {
		t.Errorf("Faction.Goal: got %+v, want %+v", got.Goal, want.Goal)
	}

	if len(got.Assets) != 1 {
		t.Fatalf("Faction.Assets: got %d, want 1", len(got.Assets))
	}
	a := got.Assets[0]
	wa := want.Assets[0]
	if a.ID != wa.ID || a.DefinitionID != wa.DefinitionID || a.OwnerID != wa.OwnerID {
		t.Errorf("Asset identity: got %+v, want %+v", a, wa)
	}
	if a.Location != wa.Location || a.CurrentHP != wa.CurrentHP || a.Ready != wa.Ready || a.Maintained != wa.Maintained {
		t.Errorf("Asset state: got %+v, want %+v", a, wa)
	}
}

func TestTurnStateRoundTrip(t *testing.T) {
	original := &FactionState{
		CampaignID: "test-campaign",
		TurnNumber: 2,
		CurrentTurn: &domain.TurnState{
			InProgress:   true,
			TurnNumber:   2,
			FactionOrder: []string{"faction-b", "faction-a"},
			CurrentIndex: 1,
			Phase:        domain.PhaseAction,
		},
	}

	path := filepath.Join(t.TempDir(), "faction_state.toml")

	if err := Save(path, original); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.CurrentTurn == nil {
		t.Fatal("CurrentTurn: got nil, want non-nil")
	}
	ct := loaded.CurrentTurn
	want := original.CurrentTurn
	if ct.InProgress != want.InProgress {
		t.Errorf("InProgress: got %v, want %v", ct.InProgress, want.InProgress)
	}
	if ct.TurnNumber != want.TurnNumber {
		t.Errorf("TurnNumber: got %d, want %d", ct.TurnNumber, want.TurnNumber)
	}
	if ct.CurrentIndex != want.CurrentIndex {
		t.Errorf("CurrentIndex: got %d, want %d", ct.CurrentIndex, want.CurrentIndex)
	}
	if ct.Phase != want.Phase {
		t.Errorf("Phase: got %v, want %v", ct.Phase, want.Phase)
	}
	if len(ct.FactionOrder) != len(want.FactionOrder) {
		t.Fatalf("FactionOrder length: got %d, want %d", len(ct.FactionOrder), len(want.FactionOrder))
	}
	for i := range want.FactionOrder {
		if ct.FactionOrder[i] != want.FactionOrder[i] {
			t.Errorf("FactionOrder[%d]: got %q, want %q", i, ct.FactionOrder[i], want.FactionOrder[i])
		}
	}
}

func TestTurnStateOmittedWhenNil(t *testing.T) {
	s := &FactionState{CampaignID: "test", TurnNumber: 1}
	path := filepath.Join(t.TempDir(), "faction_state.toml")

	if err := Save(path, s); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.CurrentTurn != nil {
		t.Errorf("CurrentTurn: got non-nil, want nil when no turn in progress")
	}
}

func TestLoadMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.toml")
	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load() on missing file should not error, got: %v", err)
	}
	if s == nil {
		t.Fatal("Load() on missing file should return empty state, got nil")
	}
}

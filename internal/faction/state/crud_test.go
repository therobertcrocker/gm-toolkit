package state

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func minimalFaction(id string) *domain.Faction {
	return &domain.Faction{
		ID:    id,
		Name:  "Test Faction",
		Scale: domain.ScaleMinor,
	}
}

func newTestState() *FactionState {
	return &FactionState{
		CampaignID:  "test-campaign",
		CycleNumber: 1,
		Factions:    map[string]*domain.Faction{},
	}
}

func TestCreateFaction(t *testing.T) {
	tests := []struct {
		name    string
		seedID  string // pre-existing faction ID, empty means no seed
		id      string // faction ID to create
		wantErr error
	}{
		{name: "success", id: "new-faction", wantErr: nil},
		{name: "duplicate", seedID: "existing", id: "existing", wantErr: ErrFactionAlreadyExists},
		{name: "empty id", id: "", wantErr: ErrInvalidFactionID},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "faction_state.toml")
			fs := newTestState()
			if tc.seedID != "" {
				fs.Factions[tc.seedID] = minimalFaction(tc.seedID)
			}

			err := CreateFaction(path, fs, minimalFaction(tc.id))

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("CreateFaction() error = %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr == nil {
				if _, ok := fs.Factions[tc.id]; !ok {
					t.Errorf("Factions[%q] not present after create", tc.id)
				}
			}
		})
	}
}

func TestDeleteFaction(t *testing.T) {
	tests := []struct {
		name     string
		seedID   string
		deleteID string
		wantErr  error
	}{
		{name: "success", seedID: "existing", deleteID: "existing", wantErr: nil},
		{name: "not found", seedID: "", deleteID: "missing", wantErr: ErrFactionNotFound},
		{name: "empty id", seedID: "", deleteID: "", wantErr: ErrInvalidFactionID},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "faction_state.toml")
			fs := newTestState()
			if tc.seedID != "" {
				fs.Factions[tc.seedID] = minimalFaction(tc.seedID)
			}

			err := DeleteFaction(path, fs, tc.deleteID)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("DeleteFaction() error = %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr == nil {
				if _, ok := fs.Factions[tc.deleteID]; ok {
					t.Errorf("Factions[%q] still present after delete", tc.deleteID)
				}
			}
		})
	}
}

func TestCreateFactionSaveFailureRollsBack(t *testing.T) {
	// A directory path makes os.Create fail, so Save errors after the
	// in-memory insert; the insert must be rolled back.
	dirPath := t.TempDir()
	fs := newTestState()

	err := CreateFaction(dirPath, fs, minimalFaction("doomed"))
	if err == nil {
		t.Fatal("CreateFaction() expected save error, got nil")
	}
	if _, ok := fs.Factions["doomed"]; ok {
		t.Error("Factions[\"doomed\"] present after failed create; insert not rolled back")
	}
}

func TestDeleteFactionSaveFailureRollsBack(t *testing.T) {
	dirPath := t.TempDir()
	fs := newTestState()
	fs.Factions["keep"] = minimalFaction("keep")

	err := DeleteFaction(dirPath, fs, "keep")
	if err == nil {
		t.Fatal("DeleteFaction() expected save error, got nil")
	}
	if _, ok := fs.Factions["keep"]; !ok {
		t.Error("Factions[\"keep\"] missing after failed delete; deletion not rolled back")
	}
}

func TestCreateDeleteRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faction_state.toml")
	fs := newTestState()
	faction := minimalFaction("round-trip-faction")

	if err := CreateFaction(path, fs, faction); err != nil {
		t.Fatalf("CreateFaction() error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() after create error: %v", err)
	}
	if _, ok := loaded.Factions[faction.ID]; !ok {
		t.Fatalf("faction %q not found after create + load", faction.ID)
	}

	if err := DeleteFaction(path, loaded, faction.ID); err != nil {
		t.Fatalf("DeleteFaction() error: %v", err)
	}

	reloaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() after delete error: %v", err)
	}
	if _, ok := reloaded.Factions[faction.ID]; ok {
		t.Errorf("faction %q still present after delete + load", faction.ID)
	}
}

package state

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

func TestFactionStateRoundTrip(t *testing.T) {
	original := &FactionState{
		CampaignID:  "test-campaign",
		CycleNumber: 3,
		Factions: map[string]*domain.Faction{
			"iron-collective": {
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
				Homeworld: domain.Location{WorldID: "Tartarus"},
				Tags:      []string{"SWN-T-001"},
				ActiveGoal: &domain.ActiveGoal{
					GoalID: "CORE-G-001",
				},
				Assets: map[string]*domain.Asset{
					"iron-collective-asset-0": {
						ID:           "iron-collective-asset-0",
						DefinitionID: "SWN-F1-001",
						OwnerID:      "iron-collective",
						Location:     domain.Location{WorldID: "Tartarus"},
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
	if loaded.CycleNumber != original.CycleNumber {
		t.Errorf("CycleNumber: got %d, want %d", loaded.CycleNumber, original.CycleNumber)
	}
	if len(loaded.Factions) != 1 {
		t.Fatalf("Factions: got %d, want 1", len(loaded.Factions))
	}

	got := loaded.Factions["iron-collective"]
	want := original.Factions["iron-collective"]

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
	if got.Homeworld.WorldID != want.Homeworld.WorldID {
		t.Errorf("Faction.Homeworld: got %q, want %q", got.Homeworld.WorldID, want.Homeworld.WorldID)
	}

	if !reflect.DeepEqual(got.Tags, want.Tags) {
		t.Errorf("Faction.Tags: got %v, want %v", got.Tags, want.Tags)
	}

	if got.ActiveGoal == nil {
		t.Fatal("Faction.ActiveGoal: got nil, want non-nil")
	}
	if got.ActiveGoal.GoalID != want.ActiveGoal.GoalID {
		t.Errorf("Faction.ActiveGoal.GoalID: got %q, want %q", got.ActiveGoal.GoalID, want.ActiveGoal.GoalID)
	}

	if len(got.Assets) != 1 {
		t.Fatalf("Faction.Assets: got %d, want 1", len(got.Assets))
	}
	a := got.Assets["iron-collective-asset-0"]
	wa := want.Assets["iron-collective-asset-0"]
	if a.ID != wa.ID || a.DefinitionID != wa.DefinitionID || a.OwnerID != wa.OwnerID {
		t.Errorf("Asset identity: got %+v, want %+v", a, wa)
	}
	if a.Location != wa.Location || a.CurrentHP != wa.CurrentHP || a.Ready != wa.Ready || a.Maintained != wa.Maintained {
		t.Errorf("Asset state: got %+v, want %+v", a, wa)
	}
}

func TestSaveWritesLocationsInline(t *testing.T) {
	mireholt := domain.Location{WorldID: "mireholt", RegionHex: spatial.RegionHex{RegionID: "F", Coord: spatial.HexCoord{Q: 7, R: -2}}}
	awkward := domain.Location{WorldID: `quote"back\slash`, RegionHex: spatial.RegionHex{RegionID: "A", Coord: spatial.HexCoord{Q: 1, R: 2}}}
	inFlight := domain.Location{RegionHex: spatial.RegionHex{RegionID: "B", Coord: spatial.HexCoord{Q: 3, R: 4}}}
	original := &FactionState{
		CampaignID:  "test-campaign",
		CycleNumber: 1,
		Factions: map[string]*domain.Faction{
			"ashfall": {
				ID:         "ashfall",
				Name:       "Ashfall",
				Homeworld:  mireholt,
				Tags:       []string{"SWN-T-001"},
				ActiveGoal: &domain.ActiveGoal{GoalID: "CORE-G-004"},
				Assets: map[string]*domain.Asset{
					"ashfall-SWN-F1-001-1": {
						ID:           "ashfall-SWN-F1-001-1",
						DefinitionID: "SWN-F1-001",
						OwnerID:      "ashfall",
						Location:     inFlight,
						CurrentHP:    3,
						CurrentOrder: &domain.MovementOrder{
							AssetID:     "ashfall-SWN-F1-001-1",
							Origin:      mireholt,
							Destination: awkward,
							Path:        []spatial.RegionHex{mireholt.RegionHex, inFlight.RegionHex, awkward.RegionHex},
							StepIdx:     1,
						},
					},
				},
				Bases: []*domain.Base{{ID: "ashfall-homeworld", OwnerID: "ashfall", Location: mireholt, CurrentHP: 20, MaxHP: 20, IsHomeworld: true}},
			},
		},
	}

	path := filepath.Join(t.TempDir(), "state.toml")
	if err := Save(path, original); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading state.toml: %v", err)
	}
	for _, header := range []string{"homeworld]", "location]", "region_hex]", "Coord]", "target_world]", "origin]", "destination]", "path]]"} {
		if strings.Contains(string(written), header) {
			t.Errorf("state.toml has a %q header, want every location inline:\n%s", header, written)
		}
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if !reflect.DeepEqual(loaded.Factions, original.Factions) {
		t.Errorf("Factions after round trip:\n got %+v\nwant %+v", loaded.Factions["ashfall"], original.Factions["ashfall"])
	}
}

func TestTurnStateRoundTrip(t *testing.T) {
	original := &FactionState{
		CampaignID:  "test-campaign",
		CycleNumber: 2,
		CurrentTurn: &domain.TurnState{
			InProgress:         true,
			CycleNumber:        2,
			FactionOrder:       []string{"faction-b", "faction-a"},
			CurrentIndex:       1,
			BookkeepingApplied: true,
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
	if ct.CycleNumber != want.CycleNumber {
		t.Errorf("CycleNumber: got %d, want %d", ct.CycleNumber, want.CycleNumber)
	}
	if ct.CurrentIndex != want.CurrentIndex {
		t.Errorf("CurrentIndex: got %d, want %d", ct.CurrentIndex, want.CurrentIndex)
	}
	if ct.BookkeepingApplied != want.BookkeepingApplied {
		t.Errorf("BookkeepingApplied: got %v, want %v", ct.BookkeepingApplied, want.BookkeepingApplied)
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
	s := &FactionState{CampaignID: "test", CycleNumber: 1}
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

func TestLoad_RefusesTagCopies(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.toml")
	body := "[factions.alpha]\nid = \"alpha\"\ntags = [{ id = \"T-011\", name = \"Planetary Government\" }]\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "tags") {
		t.Fatalf("Load = %v, want an error naming the tags key", err)
	}
}

func TestGovernmentsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.toml")
	original := &FactionState{Factions: map[string]*domain.Faction{}, Governments: map[string]string{"Krylos": "f1"}}
	if err := Save(path, original); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reflect.DeepEqual(loaded.Governments, original.Governments) {
		t.Errorf("Governments = %v, want %v", loaded.Governments, original.Governments)
	}
}

func TestGovernmentsOmittedWhenEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.toml")
	if err := Save(path, &FactionState{Factions: map[string]*domain.Faction{}}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if strings.Contains(string(body), "governments") {
		t.Errorf("state.toml = %q, want no governments table", body)
	}
}

func TestRuler(t *testing.T) {
	var nilState *FactionState
	if got := nilState.Ruler("Krylos"); got != "" {
		t.Errorf("nil state Ruler = %q, want none", got)
	}
	if got := (&FactionState{}).Ruler("Krylos"); got != "" {
		t.Errorf("no record Ruler = %q, want none (an independent world)", got)
	}
	if got := (&FactionState{Governments: map[string]string{"Krylos": "f1"}}).Ruler("Krylos"); got != "f1" {
		t.Errorf("Ruler = %q, want f1", got)
	}
}

func TestWorldsRuledBy(t *testing.T) {
	var nilState *FactionState
	if got := nilState.WorldsRuledBy("f1"); got != nil {
		t.Errorf("nil state WorldsRuledBy = %v, want none", got)
	}
	governed := &FactionState{Governments: map[string]string{"Krylos": "f1", "Hadrian": "f2", "Anchorage": "f1"}}
	if got := governed.WorldsRuledBy("f1"); !reflect.DeepEqual(got, []string{"Anchorage", "Krylos"}) {
		t.Errorf("WorldsRuledBy(f1) = %v, want [Anchorage Krylos]", got)
	}
}

func TestIsIndependent(t *testing.T) {
	var nilState *FactionState
	if !nilState.IsIndependent("Krylos") {
		t.Error("nil state IsIndependent = false, want true")
	}
	governed := &FactionState{Governments: map[string]string{"Krylos": "f1"}}
	if governed.IsIndependent("Krylos") {
		t.Error("a governed world IsIndependent = true, want false")
	}
	if !governed.IsIndependent("Hadrian") {
		t.Error("a world with no entry IsIndependent = false, want true")
	}
}

func TestClone_LeavesTheOriginalUntouched(t *testing.T) {
	original := &FactionState{
		Factions:    map[string]*domain.Faction{"f1": {ID: "f1", Coin: 3}},
		Governments: map[string]string{"Tartarus": "f1"},
		CurrentTurn: &domain.TurnState{FactionOrder: []string{"f1"}},
	}

	clone := original.Clone()
	clone.Factions["f1"].Coin = 9
	clone.Factions["f2"] = &domain.Faction{ID: "f2"}
	clone.Governments["Tartarus"] = "f2"
	clone.CurrentTurn.FactionOrder[0] = "f2"
	clone.CurrentTurn.CurrentIndex = 1

	if original.Factions["f1"].Coin != 3 || len(original.Factions) != 1 {
		t.Errorf("original factions changed: %+v", original.Factions)
	}
	if original.Governments["Tartarus"] != "f1" {
		t.Errorf("original governments changed: %v", original.Governments)
	}
	if original.CurrentTurn.FactionOrder[0] != "f1" || original.CurrentTurn.CurrentIndex != 0 {
		t.Errorf("original turn changed: %+v", original.CurrentTurn)
	}
}

package overlay

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

// testRegionMap builds a one-region, two-world map via the TOML loader (the only
// public constructor). world-b sits at hex (1,0) in region "void".
func testRegionMap(t *testing.T) *spatial.RegionMap {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"regions.toml": "[[region]]\nid = \"void\"\nname = \"The Void\"\nhexes = [[0,0],[1,0]]\n",
		"worlds.toml": "[[world]]\nid = \"world-a\"\nname = \"Alpha\"\ntech_level = 5\nregion = \"void\"\nhex_q = 0\nhex_r = 0\n\n" +
			"[[world]]\nid = \"world-b\"\nname = \"Beta\"\ntech_level = 5\nregion = \"void\"\nhex_q = 1\nhex_r = 0\n",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	rm, err := spatial.LoadRegionMap(dir)
	if err != nil {
		t.Fatalf("LoadRegionMap: %v", err)
	}
	return rm
}

// TestMovementDecisions pins the OQ-9 invariant: Issue/Revise emit a Destination
// carrying BOTH WorldID and the map-resolved RegionHex; Cancel emits a nil
// Destination; Hold emits nothing.
func TestMovementDecisions(t *testing.T) {
	spatialMap := testRegionMap(t)
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"scout": {ID: "scout", Name: "Scout"},
	}}
	eligible := []*domain.Asset{
		{ID: "a-hold", DefinitionID: "scout", Location: domain.Location{WorldID: "world-a"}},
		{ID: "a-issue", DefinitionID: "scout", Location: domain.Location{WorldID: "world-a"}},
		{ID: "a-revise", DefinitionID: "scout", Location: domain.Location{WorldID: "world-a"}},
		{ID: "a-cancel", DefinitionID: "scout", Location: domain.Location{WorldID: "world-a"}},
	}

	o := NewMovement(eligible, rb, spatialMap)
	o.data.rows[0].kind = orderNone
	o.data.rows[1].kind, o.data.rows[1].worldID = orderIssue, "world-b"
	o.data.rows[2].kind, o.data.rows[2].worldID = orderRevise, "world-b"
	o.data.rows[3].kind = orderCancel

	got := o.decisions()

	byID := make(map[string]world.MovementDecision, len(got))
	for _, d := range got {
		byID[d.AssetID] = d
	}

	if len(got) != 3 {
		t.Fatalf("decisions() len = %d, want 3 (hold omitted); got %+v", len(got), got)
	}
	if _, ok := byID["a-hold"]; ok {
		t.Error("hold asset should not emit a decision")
	}

	wantHex := spatial.RegionHex{RegionID: "void", Coord: spatial.HexCoord{Q: 1, R: 0}}
	for _, id := range []string{"a-issue", "a-revise"} {
		d := byID[id]
		wantKind := world.MovementDecisionIssue
		if id == "a-revise" {
			wantKind = world.MovementDecisionRevise
		}
		if d.Kind != wantKind {
			t.Errorf("%s: Kind = %v, want %v", id, d.Kind, wantKind)
		}
		if d.Destination == nil {
			t.Fatalf("%s: Destination is nil, want both WorldID and RegionHex set", id)
		}
		if d.Destination.WorldID != "world-b" {
			t.Errorf("%s: Destination.WorldID = %q, want %q", id, d.Destination.WorldID, "world-b")
		}
		if d.Destination.RegionHex != wantHex {
			t.Errorf("%s: Destination.RegionHex = %+v, want %+v", id, d.Destination.RegionHex, wantHex)
		}
	}

	cancel := byID["a-cancel"]
	if cancel.Kind != world.MovementDecisionCancel {
		t.Errorf("cancel: Kind = %v, want Cancel", cancel.Kind)
	}
	if cancel.Destination != nil {
		t.Errorf("cancel: Destination = %+v, want nil", cancel.Destination)
	}
}

// escDoneAnswer sends Esc to an overlay and returns the OverlayDoneMsg answer.
// The phase collector type-asserts the answer, so the decline must be the typed
// empty slice — an untyped nil would fail the assertion.
func escDoneAnswer(t *testing.T, o Overlay) any {
	t.Helper()
	_, cmd := o.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("Esc produced no command, want OverlayDoneMsg")
	}
	done, ok := cmd().(OverlayDoneMsg)
	if !ok {
		t.Fatal("Esc command did not emit OverlayDoneMsg")
	}
	return done.Answer
}

func TestMovementEscDeclines(t *testing.T) {
	spatialMap := testRegionMap(t)
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"scout": {ID: "scout", Name: "Scout"},
	}}
	eligible := []*domain.Asset{
		{ID: "a-1", DefinitionID: "scout", Location: domain.Location{WorldID: "world-a"}},
	}

	answer := escDoneAnswer(t, NewMovement(eligible, rb, spatialMap))
	decisions, ok := answer.([]world.MovementDecision)
	if !ok {
		t.Fatalf("answer = %T, want []world.MovementDecision", answer)
	}
	if len(decisions) != 0 {
		t.Errorf("decisions = %+v, want empty", decisions)
	}
}

func TestTransportCargoEscDeclines(t *testing.T) {
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"freighter": {ID: "freighter", Name: "Freighter"},
		"scout":     {ID: "scout", Name: "Scout"},
	}}
	transport := &domain.Asset{ID: "t-1", DefinitionID: "freighter"}
	eligible := []*domain.Asset{{ID: "c-1", DefinitionID: "scout"}}

	answer := escDoneAnswer(t, NewTransportCargo(transport, eligible, &domain.TransportProfile{MaxCargo: 2}, rb))
	cargo, ok := answer.([]*domain.Asset)
	if !ok {
		t.Fatalf("answer = %T, want []*domain.Asset", answer)
	}
	if len(cargo) != 0 {
		t.Errorf("cargo = %+v, want empty", cargo)
	}
}

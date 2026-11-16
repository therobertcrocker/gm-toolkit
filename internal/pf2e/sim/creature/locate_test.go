package creature_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func TestLocate(t *testing.T) {
	root := t.TempDir()
	partyDir := filepath.Join(root, "pf2e", "party")
	encountersDir := filepath.Join(root, "encounters")
	for _, dir := range []string{
		filepath.Join(partyDir, "kas"),
		filepath.Join(encountersDir, "fight-a", "luna"),
		filepath.Join(encountersDir, "fight-a", "lilith"),
		filepath.Join(encountersDir, "fight-b", "lilith"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	cases := []struct {
		id      string
		want    string
		wantErr error
	}{
		{id: "kas", want: filepath.Join(partyDir, "kas")},
		{id: "luna", want: filepath.Join(encountersDir, "fight-a", "luna")},
		{id: "lilith", wantErr: creature.ErrCreatureAmbiguous},
		{id: "nobody", wantErr: creature.ErrCreatureNotFound},
		{id: "../kas", wantErr: creature.ErrBadCreatureID},
	}
	for _, testCase := range cases {
		got, err := creature.Locate(partyDir, encountersDir, testCase.id)
		if testCase.wantErr != nil {
			if !errors.Is(err, testCase.wantErr) {
				t.Errorf("Locate(%q) err = %v, want %v", testCase.id, err, testCase.wantErr)
			}
			continue
		}
		if err != nil || got != testCase.want {
			t.Errorf("Locate(%q) = %q, %v, want %q", testCase.id, got, err, testCase.want)
		}
	}
}

func TestLocateWithoutEncountersFolder(t *testing.T) {
	root := t.TempDir()
	partyDir := filepath.Join(root, "party")
	if err := os.MkdirAll(filepath.Join(partyDir, "kas"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, err := creature.Locate(partyDir, filepath.Join(root, "missing"), "kas"); err != nil {
		t.Errorf("Locate: %v", err)
	}
}

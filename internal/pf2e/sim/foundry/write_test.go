package foundry

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func copyFile(t *testing.T, from, to string) {
	t.Helper()
	content, err := os.ReadFile(from)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(to, content, 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
}

func TestImportFightWritesNPCLayers(t *testing.T) {
	fight := filepath.Join(t.TempDir(), "fall-of-luna")
	party := filepath.Join(t.TempDir(), "party")
	copyFile(t, filepath.Join("testdata", "actors", "abyssal-luna.json"), filepath.Join(fight, "source", "actors", "abyssal-luna.json"))
	copyFile(t, filepath.Join("testdata", "items", "swear-fidelity.json"), filepath.Join(fight, "source", "items", "swear-fidelity.json"))

	reports, err := ImportFight(fight, party)
	if err != nil {
		t.Fatalf("ImportFight: %v", err)
	}
	if len(reports) != 2 {
		t.Fatalf("reports = %d, want 2", len(reports))
	}
	luna := reports[0]
	if luna.Source != "actors/abyssal-luna.json" || luna.Creature != "abyssal-luna" || luna.Folder != filepath.Join(fight, "abyssal-luna") {
		t.Errorf("luna report = %+v", luna)
	}
	content, err := os.ReadFile(filepath.Join(fight, "abyssal-luna", "import.toml"))
	if err != nil {
		t.Fatalf("read layer: %v", err)
	}
	if !strings.HasPrefix(string(content), "# Written by gm-toolkit pf2e sim import from actors/abyssal-luna.json.") {
		t.Errorf("layer header = %q", strings.SplitN(string(content), "\n", 2)[0])
	}
	loaded, err := creature.Load(luna.Folder)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Defense.HP != 300 {
		t.Errorf("HP = %d, want 300", loaded.Defense.HP)
	}
	loose := reports[1]
	if loose.Source != "items/swear-fidelity.json" || loose.Folder != "" || !hasSkip(loose, "Swear Fidelity", ReasonNotPlaced) {
		t.Errorf("loose item report = %+v", loose)
	}
}

func TestImportFightNeedsExports(t *testing.T) {
	_, err := ImportFight(t.TempDir(), t.TempDir())
	if !errors.Is(err, ErrNoExports) {
		t.Errorf("err = %v, want ErrNoExports", err)
	}
}

func TestImportFightRejectsFileNamesThatAreNotIDs(t *testing.T) {
	for _, name := range []string{".json", "Abyssal Luna.json", "source.json"} {
		fight := filepath.Join(t.TempDir(), "fall-of-luna")
		copyFile(t, filepath.Join("testdata", "actors", "abyssal-luna.json"), filepath.Join(fight, "source", "actors", name))
		if _, err := ImportFight(fight, t.TempDir()); !errors.Is(err, errBadFileName) {
			t.Errorf("%q: err = %v, want errBadFileName", name, err)
		}
	}
}

func TestImportFightPlacesPCsInParty(t *testing.T) {
	fight := filepath.Join(t.TempDir(), "fall-of-luna")
	party := filepath.Join(t.TempDir(), "party")
	copyFile(t, filepath.Join("testdata", "actors", "kas.json"), filepath.Join(fight, "source", "actors", "kas.json"))

	reports, err := ImportFight(fight, party)
	if err != nil {
		t.Fatalf("ImportFight: %v", err)
	}
	if len(reports) != 1 || reports[0].Kind != creature.KindPC || reports[0].Folder != filepath.Join(party, "kas") {
		t.Fatalf("reports = %+v, want one pc report for party/kas", reports)
	}
	loaded, err := creature.Load(reports[0].Folder)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Build == nil || loaded.Strikes["protectors-fury"].Name != "Protector's Fury" {
		t.Errorf("loaded = %+v, want Kas with a build and Protector's Fury", loaded)
	}
}

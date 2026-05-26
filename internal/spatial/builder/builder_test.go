package builder

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

var update = flag.Bool("update", false, "regenerate testdata/valid/*/golden files")

func TestBuild_Valid(t *testing.T) {
	entries, err := os.ReadDir("testdata/valid")
	if err != nil {
		t.Fatalf("read testdata/valid: %v", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		t.Run(name, func(t *testing.T) {
			src := filepath.Join("testdata", "valid", name, "source")
			goldenDir := filepath.Join("testdata", "valid", name, "golden")
			dst := t.TempDir()

			if _, err := Build(Options{SourceDir: src, DestDir: dst}); err != nil {
				t.Fatalf("Build: %v", err)
			}

			for _, fname := range []string{"regions.toml", "worlds.toml"} {
				gotBytes, err := os.ReadFile(filepath.Join(dst, fname))
				if err != nil {
					t.Fatalf("read emitted %s: %v", fname, err)
				}
				goldenPath := filepath.Join(goldenDir, fname)

				if *update {
					if err := os.MkdirAll(goldenDir, 0o755); err != nil {
						t.Fatalf("mkdir golden dir: %v", err)
					}
					if err := os.WriteFile(goldenPath, gotBytes, 0o644); err != nil {
						t.Fatalf("write golden %s: %v", fname, err)
					}
					continue
				}

				wantBytes, err := os.ReadFile(goldenPath)
				if err != nil {
					t.Fatalf("read golden %s: %v (regenerate with -update)", fname, err)
				}
				if string(gotBytes) != string(wantBytes) {
					t.Errorf("%s mismatch (regenerate with -update):\n--- got ---\n%s\n--- want ---\n%s",
						fname, gotBytes, wantBytes)
				}
			}
		})
	}
}

func TestGoldens_LoadableByEngine(t *testing.T) {
	regionMap, err := spatial.LoadRegionMap("testdata/valid/helions_reach/golden")
	if err != nil {
		t.Fatalf("spatial.LoadRegionMap on goldens: %v", err)
	}
	for _, worldID := range []string{"helion", "anvil", "bastion", "veilport"} {
		loc, ok := regionMap.Location(worldID)
		if !ok {
			t.Errorf("world %q missing from loaded golden", worldID)
			continue
		}
		world := loc.(*spatial.World)
		if world.RegionID() == "" {
			t.Errorf("world %q loaded with empty region", worldID)
		}
	}
	if regionID, ok := regionMap.RegionOfHex(spatial.HexCoord{Q: 13, R: 1}); !ok || regionID != "D" {
		t.Errorf("hex (13,1) = (region=%q, ok=%v), want (\"D\", true)", regionID, ok)
	}
}

func TestBuild_SelfCheckRoundTrip(t *testing.T) {
	src := "testdata/valid/helions_reach/source"
	dst := t.TempDir()
	if _, err := Build(Options{SourceDir: src, DestDir: dst}); err != nil {
		t.Fatalf("Build: %v", err)
	}
	regionMap, err := spatial.LoadRegionMap(dst)
	if err != nil {
		t.Fatalf("LoadRegionMap: %v", err)
	}
	for _, worldID := range []string{"helion", "anvil", "bastion", "veilport"} {
		if _, ok := regionMap.Location(worldID); !ok {
			t.Errorf("world %q missing from loaded RegionMap", worldID)
		}
	}
}

func TestBuild_MissingSource(t *testing.T) {
	dst := t.TempDir()
	_, err := Build(Options{SourceDir: t.TempDir(), DestDir: dst})
	if !errors.Is(err, ErrMissingSource) {
		t.Errorf("err = %v, want errors.Is(err, ErrMissingSource)", err)
	}
}

func TestBuild_OutputExists(t *testing.T) {
	src := "testdata/valid/helions_reach/source"
	dst := t.TempDir()

	if _, err := Build(Options{SourceDir: src, DestDir: dst}); err != nil {
		t.Fatalf("initial Build: %v", err)
	}

	_, err := Build(Options{SourceDir: src, DestDir: dst})
	if !errors.Is(err, ErrOutputExists) {
		t.Errorf("second Build without Replace: err = %v, want errors.Is(err, ErrOutputExists)", err)
	}

	if _, err := Build(Options{SourceDir: src, DestDir: dst, Replace: true}); err != nil {
		t.Errorf("Build with Replace: unexpected error %v", err)
	}
}

func TestBuild_Errors(t *testing.T) {
	entries, err := os.ReadDir("testdata/errors")
	if err != nil {
		t.Fatalf("read testdata/errors: %v", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		t.Run(name, func(t *testing.T) {
			src := filepath.Join("testdata", "errors", name, "source")
			wantBytes, err := os.ReadFile(filepath.Join("testdata", "errors", name, "err.txt"))
			if err != nil {
				t.Fatalf("read err.txt: %v", err)
			}
			want := strings.TrimSpace(string(wantBytes))

			dst := t.TempDir()
			_, err = Build(Options{SourceDir: src, DestDir: dst})
			if err == nil {
				t.Fatalf("got nil error, want substring %q", want)
			}
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("err = %q, want substring %q", err.Error(), want)
			}
		})
	}
}

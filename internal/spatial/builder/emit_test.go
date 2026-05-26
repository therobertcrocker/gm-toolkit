package builder

import (
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"

	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

func TestEmit_RoundTrip(t *testing.T) {
	derived := &derivedMap{
		Regions: []derivedRegion{
			{
				ID:   "A",
				Name: "Region A",
				Hexes: []spatial.HexCoord{
					{Q: 0, R: 0},
					{Q: 1, R: 0},
				},
				Boundaries: []derivedBoundary{
					{
						From:     spatial.HexCoord{Q: 1, R: 0},
						ToRegion: "B",
						To:       spatial.HexCoord{Q: 2, R: 0},
					},
				},
			},
			{
				ID:    "B",
				Name:  "Region B",
				Hexes: []spatial.HexCoord{{Q: 2, R: 0}},
			},
		},
		Worlds: []derivedWorld{
			{
				ID:         "x",
				Name:       "X",
				TechLevel:  7,
				Population: 1000,
				Region:     "A",
				Coord:      spatial.HexCoord{Q: 0, R: 0},
			},
		},
	}

	dir := t.TempDir()
	if err := emit(dir, derived); err != nil {
		t.Fatalf("emit: %v", err)
	}

	var regionsFile struct {
		Region []emitRegion `toml:"region"`
	}
	if _, err := toml.DecodeFile(filepath.Join(dir, "regions.toml"), &regionsFile); err != nil {
		t.Fatalf("decode regions.toml: %v", err)
	}
	if got := len(regionsFile.Region); got != 2 {
		t.Fatalf("len(regions) = %d, want 2", got)
	}

	regionA := regionsFile.Region[0]
	if regionA.ID != "A" || regionA.Name != "Region A" {
		t.Errorf("region A meta = {%q, %q}, want {\"A\", \"Region A\"}", regionA.ID, regionA.Name)
	}
	if len(regionA.Hexes) != 2 {
		t.Errorf("region A hex count = %d, want 2", len(regionA.Hexes))
	} else {
		if regionA.Hexes[0] != [2]int{0, 0} || regionA.Hexes[1] != [2]int{1, 0} {
			t.Errorf("region A hexes = %v, want [[0 0] [1 0]]", regionA.Hexes)
		}
	}
	if len(regionA.Boundaries) != 1 {
		t.Fatalf("region A boundary count = %d, want 1", len(regionA.Boundaries))
	}
	boundary := regionA.Boundaries[0]
	if boundary.FromQ != 1 || boundary.FromR != 0 || boundary.ToRegion != "B" || boundary.ToQ != 2 || boundary.ToR != 0 {
		t.Errorf("boundary = {fromQ=%d fromR=%d toRegion=%q toQ=%d toR=%d}, want {1, 0, \"B\", 2, 0}",
			boundary.FromQ, boundary.FromR, boundary.ToRegion, boundary.ToQ, boundary.ToR)
	}

	regionB := regionsFile.Region[1]
	if regionB.ID != "B" || len(regionB.Hexes) != 1 || len(regionB.Boundaries) != 0 {
		t.Errorf("region B = %+v, want ID=\"B\", 1 hex, 0 boundaries", regionB)
	}

	var worldsFile struct {
		World []emitWorld `toml:"world"`
	}
	if _, err := toml.DecodeFile(filepath.Join(dir, "worlds.toml"), &worldsFile); err != nil {
		t.Fatalf("decode worlds.toml: %v", err)
	}
	if len(worldsFile.World) != 1 {
		t.Fatalf("len(worlds) = %d, want 1", len(worldsFile.World))
	}
	world := worldsFile.World[0]
	want := emitWorld{
		ID: "x", Name: "X", TechLevel: 7, Population: 1000,
		Region: "A", HexQ: 0, HexR: 0,
	}
	if world != want {
		t.Errorf("world = %+v, want %+v", world, want)
	}
}

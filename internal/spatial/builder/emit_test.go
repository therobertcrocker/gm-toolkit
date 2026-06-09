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
			},
			{
				ID:    "B",
				Name:  "Region B",
				Hexes: []spatial.HexCoord{{Q: 2, R: 0}},
			},
		},
		Warps: []derivedWarp{
			{FromRegion: "A", From: spatial.HexCoord{Q: 1, R: 0}, ToRegion: "B", To: spatial.HexCoord{Q: 2, R: 0}},
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

	var doc struct {
		Region []emitRegion `toml:"region"`
		Warp   []emitWarp   `toml:"warp"`
	}
	if _, err := toml.DecodeFile(filepath.Join(dir, "regions.toml"), &doc); err != nil {
		t.Fatalf("decode regions.toml: %v", err)
	}
	if got := len(doc.Region); got != 2 {
		t.Fatalf("len(regions) = %d, want 2", got)
	}

	regionA := doc.Region[0]
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

	regionB := doc.Region[1]
	if regionB.ID != "B" || len(regionB.Hexes) != 1 {
		t.Errorf("region B = %+v, want ID=\"B\", 1 hex", regionB)
	}

	if got := len(doc.Warp); got != 1 {
		t.Fatalf("warp count = %d, want 1", got)
	}
	warp := doc.Warp[0]
	if warp.FromRegion != "A" || warp.FromQ != 1 || warp.FromR != 0 ||
		warp.ToRegion != "B" || warp.ToQ != 2 || warp.ToR != 0 {
		t.Errorf("warp = %+v, want {FromRegion:A FromQ:1 FromR:0 ToRegion:B ToQ:2 ToR:0}", warp)
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

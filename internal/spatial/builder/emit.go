package builder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type emitRegion struct {
	ID    string   `toml:"id"`
	Name  string   `toml:"name"`
	Hexes [][2]int `toml:"hexes"`
}

type emitWarp struct {
	FromRegion string `toml:"from_region"`
	FromQ      int    `toml:"from_q"`
	FromR      int    `toml:"from_r"`
	ToRegion   string `toml:"to_region"`
	ToQ        int    `toml:"to_q"`
	ToR        int    `toml:"to_r"`
}

type emitWorld struct {
	ID         string `toml:"id"`
	Name       string `toml:"name"`
	TechLevel  int    `toml:"tech_level"`
	Population int    `toml:"population"`
	Region     string `toml:"region"`
	HexQ       int    `toml:"hex_q"`
	HexR       int    `toml:"hex_r"`
}

func emit(dstDir string, derived *derivedMap) error {
	if err := writeRegions(filepath.Join(dstDir, "regions.toml"), derived); err != nil {
		return err
	}
	if err := writeWorlds(filepath.Join(dstDir, "worlds.toml"), derived); err != nil {
		return err
	}
	return nil
}

func writeRegions(path string, derived *derivedMap) error {
	out := struct {
		Region []emitRegion `toml:"region"`
		Warp   []emitWarp   `toml:"warp"`
	}{}
	for _, r := range derived.Regions {
		em := emitRegion{ID: r.ID, Name: r.Name}
		for _, h := range r.Hexes {
			em.Hexes = append(em.Hexes, [2]int{h.Q, h.R})
		}
		out.Region = append(out.Region, em)
	}
	for _, w := range derived.Warps {
		out.Warp = append(out.Warp, emitWarp{
			FromRegion: w.FromRegion, FromQ: w.From.Q, FromR: w.From.R,
			ToRegion: w.ToRegion, ToQ: w.To.Q, ToR: w.To.R,
		})
	}
	return writeTOML(path, out)
}

func writeWorlds(path string, derived *derivedMap) error {
	out := struct {
		World []emitWorld `toml:"world"`
	}{}
	for _, w := range derived.Worlds {
		out.World = append(out.World, emitWorld{
			ID: w.ID, Name: w.Name,
			TechLevel: w.TechLevel, Population: w.Population,
			Region: w.Region,
			HexQ:   w.Coord.Q, HexR: w.Coord.R,
		})
	}
	return writeTOML(path, out)
}

func writeTOML(path string, v any) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(v); err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	return nil
}

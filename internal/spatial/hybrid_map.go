package spatial

import (
	"fmt"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type HexCoord struct{ Q, R int }

type BoundaryConnection struct {
	From     HexCoord
	ToRegion string
	To       HexCoord
}

type Region struct {
	ID         string
	Name       string
	Hexes      map[HexCoord]bool
	Boundaries []BoundaryConnection
}

type Fragment struct {
	FragmentID     string
	FragmentName   string
	FragTechLevel  int
	FragPopulation int
	Region         string
	Hex            HexCoord
}

func (fragment *Fragment) ID() string      { return fragment.FragmentID }
func (fragment *Fragment) Name() string    { return fragment.FragmentName }
func (fragment *Fragment) TechLevel() int  { return fragment.FragTechLevel }
func (fragment *Fragment) Population() int { return fragment.FragPopulation }

type HybridMap struct {
	regions   map[string]*Region
	fragments map[string]*Fragment
}

type tomlRegion struct {
	ID         string         `toml:"id"`
	Name       string         `toml:"name"`
	Hexes      [][2]int       `toml:"hexes"`
	Boundaries []tomlBoundary `toml:"boundary"`
}

type tomlBoundary struct {
	FromQ    int    `toml:"from_q"`
	FromR    int    `toml:"from_r"`
	ToRegion string `toml:"to_region"`
	ToQ      int    `toml:"to_q"`
	ToR      int    `toml:"to_r"`
}

type tomlFragment struct {
	ID         string `toml:"id"`
	Name       string `toml:"name"`
	TechLevel  int    `toml:"tech_level"`
	Population int    `toml:"population"`
	Region     string `toml:"region"`
	HexQ       int    `toml:"hex_q"`
	HexR       int    `toml:"hex_r"`
}

type regionsFile struct {
	Region []tomlRegion `toml:"region"`
}

type fragmentsFile struct {
	Fragment []tomlFragment `toml:"fragment"`
}

func LoadHybrid(dataDir string) (*HybridMap, error) {
	var regionsDoc regionsFile
	if _, err := toml.DecodeFile(filepath.Join(dataDir, "regions.toml"), &regionsDoc); err != nil {
		return nil, fmt.Errorf("loading regions.toml: %w", err)
	}

	regions := make(map[string]*Region, len(regionsDoc.Region))
	for _, region := range regionsDoc.Region {
		hexes := make(map[HexCoord]bool, len(region.Hexes))
		for _, hex := range region.Hexes {
			hexes[HexCoord{Q: hex[0], R: hex[1]}] = true
		}
		boundaries := make([]BoundaryConnection, 0, len(region.Boundaries))
		for _, boundary := range region.Boundaries {
			boundaries = append(boundaries, BoundaryConnection{
				From:     HexCoord{Q: boundary.FromQ, R: boundary.FromR},
				ToRegion: boundary.ToRegion,
				To:       HexCoord{Q: boundary.ToQ, R: boundary.ToR},
			})
		}
		regions[region.ID] = &Region{
			ID:         region.ID,
			Name:       region.Name,
			Hexes:      hexes,
			Boundaries: boundaries,
		}
	}

	var fragmentsDoc fragmentsFile
	if _, err := toml.DecodeFile(filepath.Join(dataDir, "fragments.toml"), &fragmentsDoc); err != nil {
		return nil, fmt.Errorf("loading fragments.toml: %w", err)
	}

	fragments := make(map[string]*Fragment, len(fragmentsDoc.Fragment))
	for _, fragment := range fragmentsDoc.Fragment {
		region, ok := regions[fragment.Region]
		if !ok {
			return nil, fmt.Errorf("fragment %q references unknown region %q", fragment.ID, fragment.Region)
		}
		hex := HexCoord{Q: fragment.HexQ, R: fragment.HexR}
		if !region.Hexes[hex] {
			return nil, fmt.Errorf("fragment %q at hex (%d,%d) is not within region %q", fragment.ID, hex.Q, hex.R, fragment.Region)
		}
		fragments[fragment.ID] = &Fragment{
			FragmentID:     fragment.ID,
			FragmentName:   fragment.Name,
			FragTechLevel:  fragment.TechLevel,
			FragPopulation: fragment.Population,
			Region:         fragment.Region,
			Hex:            hex,
		}
	}

	return &HybridMap{regions: regions, fragments: fragments}, nil
}

func (hybridMap *HybridMap) Location(id string) (Location, bool) {
	fragment, ok := hybridMap.fragments[id]
	return fragment, ok
}

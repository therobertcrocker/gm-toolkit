package builder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type dataFile struct {
	Regions map[string]regionEntry `toml:"regions"`
	Worlds  []worldEntry           `toml:"worlds"`
	Warps   []warpEntry            `toml:"warps"`
}

type regionEntry struct {
	Name string `toml:"name"`
}

type worldEntry struct {
	ID         string  `toml:"id"`
	Name       string  `toml:"name"`
	At         atGlyph `toml:"at"`
	TechLevel  int     `toml:"tech_level"`
	Population int     `toml:"population"`
	Region     string  `toml:"region"`
}

type atGlyph struct{ Glyph glyph }

func (a *atGlyph) UnmarshalTOML(v interface{}) error {
	switch x := v.(type) {
	case int64:
		if x < 1 || x > 9 {
			return fmt.Errorf("'at' integer must be in [1,9], got %d", x)
		}
		a.Glyph = glyph('0' + byte(x))
		return nil
	case string:
		if len(x) != 1 {
			return fmt.Errorf("'at' string must be a single character, got %q", x)
		}
		r := rune(x[0])
		if r < 'a' || r > 'z' {
			return fmt.Errorf("'at' string must be 'a'-'z', got %q", x)
		}
		a.Glyph = glyph(r)
		return nil
	default:
		return fmt.Errorf("'at' must be integer 1-9 or string 'a'-'z', got %T", v)
	}
}

type warpEntry struct {
	From hexAddr `toml:"from"`
	To   hexAddr `toml:"to"`
}

type hexAddr struct {
	Region string `toml:"region"`
	Row    int    `toml:"row"`
	Col    int    `toml:"col"`
}

func parseData(srcDir string) (*dataFile, error) {
	path := filepath.Join(srcDir, "data.toml")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrMissingSource, path)
		}
		return nil, fmt.Errorf("stat %s: %w", path, err)
	}
	var data dataFile
	if _, err := toml.DecodeFile(path, &data); err != nil {
		return nil, dataErrorf("", "decode: %v", err)
	}
	return &data, nil
}

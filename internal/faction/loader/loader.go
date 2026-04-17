package loader

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// DataLoader is implemented by each static data type.
type DataLoader interface {
	Load(path string) error
}

// StaticData holds all static game data, populated by loaders at startup.
type StaticData struct {
	Assets map[string]*domain.AssetDefinition
}

// LoadStaticData loads all static data files from the given directory.
func LoadStaticData(dataDir string) (*StaticData, error) {
	slog.Info("loading static data", "dir", dataDir)

	data := &StaticData{}

	loaders := []struct {
		loader DataLoader
		file   string
	}{
		{&AssetLoader{data: data}, "assets.toml"},
	}

	for _, l := range loaders {
		path := filepath.Join(dataDir, l.file)
		if err := l.loader.Load(path); err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", l.file, err)
		}
	}

	slog.Info("static data loaded", "assets", len(data.Assets))

	return data, nil
}

// decodeTOML opens a TOML file and decodes it into target.
func decodeTOML(path string, target any) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open %s: %w", path, err)
	}
	defer f.Close()

	if _, err := toml.NewDecoder(f).Decode(target); err != nil {
		return fmt.Errorf("could not decode %s: %w", path, err)
	}

	return nil
}

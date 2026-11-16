package creature

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	importLayer = "import.toml"
	handLayer   = "hand.toml"
)

var (
	ErrNoLayers = errors.New("creature has neither an import nor a hand layer")
	ErrInvalid  = errors.New("invalid creature")
)

func Load(dir string, variants ...string) (*Creature, error) {
	var names []string
	for _, name := range []string{importLayer, handLayer} {
		_, err := os.Stat(filepath.Join(dir, name))
		if err == nil {
			names = append(names, name)
			continue
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("load creature %s: %w", dir, err)
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("load creature %s: %w", dir, ErrNoLayers)
	}
	for _, variant := range variants {
		names = append(names, variant+".toml")
	}

	merged := map[string]any{}
	for _, name := range names {
		layer, err := readLayer(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("load creature %s: %w", dir, err)
		}
		mergeInto(merged, layer)
	}

	var buffer bytes.Buffer
	if err := toml.NewEncoder(&buffer).Encode(merged); err != nil {
		return nil, fmt.Errorf("load creature %s: encode merged layers: %w", dir, err)
	}
	var creature Creature
	if _, err := toml.Decode(buffer.String(), &creature); err != nil {
		return nil, fmt.Errorf("load creature %s: decode merged layers: %w", dir, err)
	}
	creature.ID = filepath.Base(dir)
	creature.Layers = names
	if err := creature.prepare(); err != nil {
		return nil, fmt.Errorf("load creature %s: %w", dir, err)
	}
	return &creature, nil
}

func readLayer(path string) (map[string]any, error) {
	var typed Creature
	metadata, err := toml.DecodeFile(path, &typed)
	if err != nil {
		return nil, fmt.Errorf("layer %s: %w", path, err)
	}
	if undecoded := metadata.Undecoded(); len(undecoded) > 0 {
		keys := make([]string, len(undecoded))
		for index, key := range undecoded {
			keys[index] = key.String()
		}
		return nil, fmt.Errorf("layer %s: unknown keys: %s", path, strings.Join(keys, ", "))
	}
	if typed.Schema != SchemaVersion {
		return nil, fmt.Errorf("layer %s: schema %d, want %d", path, typed.Schema, SchemaVersion)
	}
	var layer map[string]any
	if _, err := toml.DecodeFile(path, &layer); err != nil {
		return nil, fmt.Errorf("layer %s: %w", path, err)
	}
	return layer, nil
}

func mergeInto(base, overlay map[string]any) {
	for key, value := range overlay {
		overlayTable, overlayIsTable := value.(map[string]any)
		baseTable, baseIsTable := base[key].(map[string]any)
		if overlayIsTable && baseIsTable {
			mergeInto(baseTable, overlayTable)
			continue
		}
		base[key] = value
	}
}

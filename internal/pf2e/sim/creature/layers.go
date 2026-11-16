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
	ErrNoLayers   = errors.New("creature has neither an import nor a hand layer")
	ErrInvalid    = errors.New("invalid creature")
	ErrBadVariant = errors.New("bad variant name")
)

func Load(directory string, variants ...string) (*Creature, error) {
	var names []string
	for _, name := range []string{importLayer, handLayer} {
		_, err := os.Stat(filepath.Join(directory, name))
		if err == nil {
			names = append(names, name)
			continue
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("load creature %s: %w", directory, err)
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("load creature %s: %w", directory, ErrNoLayers)
	}
	for _, variant := range variants {
		name := variant + ".toml"
		if !isFolderName(variant) || name == importLayer || name == handLayer {
			return nil, fmt.Errorf("load creature %s: %w: %q", directory, ErrBadVariant, variant)
		}
		names = append(names, name)
	}

	merged := map[string]any{}
	for _, name := range names {
		layer, err := readLayer(filepath.Join(directory, name))
		if err != nil {
			return nil, fmt.Errorf("load creature %s: %w", directory, err)
		}
		mergeInto(merged, layer)
	}

	var buffer bytes.Buffer
	if err := toml.NewEncoder(&buffer).Encode(merged); err != nil {
		return nil, fmt.Errorf("load creature %s: encode merged layers: %w", directory, err)
	}
	var creature Creature
	if _, err := toml.Decode(buffer.String(), &creature); err != nil {
		return nil, fmt.Errorf("load creature %s: decode merged layers: %w", directory, err)
	}
	creature.ID = filepath.Base(directory)
	creature.Layers = names
	if err := creature.prepare(); err != nil {
		return nil, fmt.Errorf("load creature %s: %w", directory, err)
	}
	return &creature, nil
}

func isFolderName(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.ContainsAny(name, `/\`)
}

func readLayer(path string) (map[string]any, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("layer %s: %w", path, err)
	}
	var typed Creature
	metadata, err := toml.Decode(string(content), &typed)
	if err != nil {
		return nil, fmt.Errorf("layer %s: %w", path, err)
	}
	if undecoded := metadata.Undecoded(); len(undecoded) > 0 {
		return nil, fmt.Errorf("layer %s: unknown keys: %s", path, joinKeys(undecoded))
	}
	// deliberate: a key with a capital letter is an error, not accepted, because the decoder matches fields without case but the merge does not, so `HP` and `hp` would both survive the merge and either could win.
	var capitalised []toml.Key
	for _, key := range metadata.Keys() {
		if text := key.String(); text != strings.ToLower(text) {
			capitalised = append(capitalised, key)
		}
	}
	if len(capitalised) > 0 {
		return nil, fmt.Errorf("layer %s: keys must be lower case: %s", path, joinKeys(capitalised))
	}
	if typed.Schema != SchemaVersion {
		return nil, fmt.Errorf("layer %s: schema %d, want %d", path, typed.Schema, SchemaVersion)
	}
	var layer map[string]any
	if _, err := toml.Decode(string(content), &layer); err != nil {
		return nil, fmt.Errorf("layer %s: %w", path, err)
	}
	return layer, nil
}

func joinKeys(keys []toml.Key) string {
	texts := make([]string, len(keys))
	for index, key := range keys {
		texts[index] = key.String()
	}
	return strings.Join(texts, ", ")
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

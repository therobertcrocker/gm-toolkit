package foundry

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

var (
	ErrNoExports   = errors.New("no actor exports in source/actors")
	errBadFileName = errors.New("export file name is not a creature id")
)

const layerHeader = "# Written by gm-toolkit pf2e sim import from %s.\n# Import rewrites this file. Put changes in hand.toml.\n\n"

func ImportFight(fightDirectory, partyDirectory string) ([]*Report, error) {
	sourceDirectory := filepath.Join(fightDirectory, "source")
	actorPaths, err := jsonFiles(filepath.Join(sourceDirectory, "actors"))
	if err != nil {
		return nil, fmt.Errorf("import %s: %w", fightDirectory, err)
	}
	if len(actorPaths) == 0 {
		return nil, fmt.Errorf("import %s: %w", fightDirectory, ErrNoExports)
	}
	var reports []*Report
	for _, path := range actorPaths {
		report, err := importActorFile(path, sourceDirectory, fightDirectory, partyDirectory)
		if err != nil {
			return reports, fmt.Errorf("import %s: %w", fightDirectory, err)
		}
		reports = append(reports, report)
	}
	itemPaths, err := jsonFiles(filepath.Join(sourceDirectory, "items"))
	if err != nil {
		return reports, fmt.Errorf("import %s: %w", fightDirectory, err)
	}
	for _, path := range itemPaths {
		loose, err := readItem(path)
		if err != nil {
			return reports, fmt.Errorf("import %s: %w", fightDirectory, err)
		}
		report := &Report{Source: relativeSource(sourceDirectory, path)}
		report.skip(loose.Name, loose.Type, ReasonNotPlaced)
		reports = append(reports, report)
	}
	return reports, nil
}

// deliberate: a directory listing, not filepath.Glob, because a campaign path holding [, * or ? would be read as a pattern.
func jsonFiles(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			paths = append(paths, filepath.Join(directory, entry.Name()))
		}
	}
	return paths, nil
}

func importActorFile(path, sourceDirectory, fightDirectory, partyDirectory string) (*Report, error) {
	source, err := readActor(path)
	if err != nil {
		return nil, err
	}
	id := strings.TrimSuffix(filepath.Base(path), ".json")
	if id == "" || id != slugify(id) {
		return nil, fmt.Errorf("%w: %s: use lowercase letters, digits and dashes", errBadFileName, path)
	}
	importer, directory := importNPC, filepath.Join(fightDirectory, id)
	if source.Type == "character" {
		importer, directory = importPC, filepath.Join(partyDirectory, id)
	} else if id == "source" {
		return nil, fmt.Errorf("%w: %s: an NPC folder named source would hold the fight's exports", errBadFileName, path)
	}
	layer, report, err := importer(source)
	if err != nil {
		return nil, err
	}
	report.Source = relativeSource(sourceDirectory, path)
	report.Creature = id
	report.Folder = directory
	if err := writeLayer(directory, report.Source, layer); err != nil {
		return nil, err
	}
	return report, nil
}

func writeLayer(directory, source string, layer *creature.Creature) error {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("write layer %s: %w", directory, err)
	}
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, layerHeader, source)
	if err := toml.NewEncoder(&buffer).Encode(layer); err != nil {
		return fmt.Errorf("write layer %s: %w", directory, err)
	}
	path := filepath.Join(directory, creature.ImportLayer)
	if err := os.WriteFile(path, buffer.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write layer %s: %w", path, err)
	}
	return nil
}

func relativeSource(sourceDirectory, path string) string {
	relative, err := filepath.Rel(sourceDirectory, path)
	if err != nil {
		return filepath.Base(path)
	}
	return filepath.ToSlash(relative)
}

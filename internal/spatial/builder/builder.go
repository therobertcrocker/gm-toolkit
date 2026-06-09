package builder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type Options struct {
	SourceDir string
	DestDir   string
	Replace   bool
}

type Summary struct {
	Regions    []string
	WorldCount int
	WarpCount  int
	Warnings   []string
}

func Build(opts Options) (Summary, error) {
	if err := checkOutputs(opts); err != nil {
		return Summary{}, err
	}
	layout, err := parseLayout(opts.SourceDir)
	if err != nil {
		return Summary{}, err
	}
	data, err := parseData(opts.SourceDir)
	if err != nil {
		return Summary{}, err
	}
	derived, err := derive(layout, data)
	if err != nil {
		return Summary{}, err
	}
	if err := emit(opts.DestDir, derived); err != nil {
		return Summary{}, err
	}
	if err := selfCheck(opts.DestDir); err != nil {
		return Summary{}, err
	}
	return summarize(derived), nil
}

func checkOutputs(opts Options) error {
	if opts.Replace {
		return nil
	}
	for _, name := range []string{"regions.toml", "worlds.toml"} {
		path := filepath.Join(opts.DestDir, name)
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%w: %s", ErrOutputExists, path)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("stat %s: %w", path, err)
		}
	}
	return nil
}

func selfCheck(dstDir string) error {
	if _, err := spatial.LoadRegionMap(dstDir); err != nil {
		return fmt.Errorf("internal: canonical output failed self-check (%w); please file a bug", err)
	}
	return nil
}

func summarize(derived *derivedMap) Summary {
	s := Summary{
		WorldCount: len(derived.Worlds),
		WarpCount:  derived.WarpCount,
		Warnings:   derived.Warnings,
	}
	for _, r := range derived.Regions {
		s.Regions = append(s.Regions, r.ID)
	}
	return s
}

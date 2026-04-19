package engine

import (
	"os"
	"path/filepath"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
)

type Engine struct {
	Rulebook *loader.Rulebook
}

func New() (*Engine, error) {
	dataDir, err := resolveDataDir()
	if err != nil {
		return nil, err
	}
	rb, err := loader.Load(dataDir)
	if err != nil {
		return nil, err
	}
	return &Engine{Rulebook: rb}, nil
}

func resolveDataDir() (string, error) {
	if dir := os.Getenv("FACTION_DATA_DIR"); dir != "" {
		return dir, nil
	}

	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exe), "data"), nil
}

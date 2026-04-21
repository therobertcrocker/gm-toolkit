package engine

import (
	"os"
	"path/filepath"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
)

// Engine is the core orchestrator. It owns the Rulebook and composes all
// sub-engines. Sub-engines that are not yet implemented are nil.
type Engine struct {
	Rulebook *loader.Rulebook
	Turn     *TurnEngine
	// Action *ActionEngine — future
	// Goal   *GoalEngine   — future
	// Tag    *TagEngine     — future
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
	e := &Engine{Rulebook: rb}
	e.Turn = newTurnEngine()
	return e, nil
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

package tui

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
)

// RunDryRun exercises one full RunCycle against a minimal smoke faction state
// and exits. No TUI is launched; all output is structured slog lines.
func RunDryRun(log *slog.Logger) error {
	log.Info("dryrun: starting RunCycle")

	rb := &rulebook.Rulebook{
		Assets:     map[string]*domain.AssetDefinition{},
		Tags:       map[string]*domain.Tag{},
		Goals:      map[string]*domain.Goal{},
		DriftCosts: []int{1},
	}

	eng := engine.NewWithRulebook(rb, adapter.NewSmokeWorldEngine(log), log)

	tmpDir, err := os.MkdirTemp("", "gm-toolkit-dryrun-*")
	if err != nil {
		return fmt.Errorf("dryrun: creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	paths := &campaigns.Paths{
		StatePath:   filepath.Join(tmpDir, "state.toml"),
		HistoryPath: filepath.Join(tmpDir, "history.jsonl"),
	}

	factionState := newSmokeFactionState()

	if err := eng.Turn.Start(factionState); err != nil {
		return fmt.Errorf("dryrun: starting turn: %w", err)
	}

	collectors := engine.Collectors{
		Phase:  adapter.NewSmokePhaseCollector(log),
		Action: adapter.NewSmokeActionCollector(),
	}
	observer := adapter.NewSmokeObserver(log)

	if err := eng.RunCycle(factionState, paths, collectors, observer); err != nil {
		return fmt.Errorf("dryrun: RunCycle: %w", err)
	}

	log.Info("dryrun: RunCycle completed cleanly")
	return nil
}

func newSmokeFactionState() *state.FactionState {
	f := &domain.Faction{
		ID:        "smoke-faction",
		Name:      "Smoke Faction",
		Scale:     domain.ScaleMinor,
		Force:     4,
		Cunning:   3,
		Wealth:    1,
		CurrentHP: 1,
		MaxHP:     10,
		Coin:      0,
		XP:        0,
		Assets:    make(map[string]*domain.Asset),
	}
	return &state.FactionState{
		CampaignID: "smoke",
		Factions:   map[string]*domain.Faction{f.ID: f},
	}
}

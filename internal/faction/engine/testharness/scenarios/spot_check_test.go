package scenarios

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/logging"
)

func TestSpotCheck_RenderedOutput(t *testing.T) {
	for _, debug := range []bool{false, true} {
		logsDir := t.TempDir()
		log, err := logging.New(logsDir, debug)
		if err != nil {
			t.Fatalf("logging.New: %v", err)
		}

		dir := t.TempDir()
		paths := &campaigns.Paths{
			FactionDataDir: testDataDir,
			StatePath:      filepath.Join(dir, "state.toml"),
			HistoryPath:    filepath.Join(dir, "history.jsonl"),
		}
		rb, err := rulebook.Load(paths.FactionDataDir)
		if err != nil {
			t.Fatalf("rulebook.Load: %v", err)
		}

		eng := engine.NewWithRulebook(rb, world.NewWithMap(&testharness.StubSpatialMap{}, log), log)

		factionState := &state.FactionState{CampaignID: "spot-check", Factions: make(map[string]*domain.Faction)}
		for _, id := range []string{"alpha", "beta"} {
			factionState.Factions[id] = &domain.Faction{
				ID: id, Name: id, Scale: domain.ScaleMinor,
				Force: 4, Cunning: 3, Wealth: 2,
				Homeworld: domain.Location{WorldID: "Tartarus"},
				MaxHP:     20, CurrentHP: 20,
				Assets: map[string]*domain.Asset{
					id + "-a1": {
						ID: id + "-a1", DefinitionID: testharness.DefSecurityPersonnel,
						OwnerID: id, Location: domain.Location{WorldID: "Tartarus"},
						CurrentHP: 3, Ready: true, Maintained: true,
					},
				},
			}
		}

		collector := &testharness.ScriptedCollector{}
		collector.SelectActionFn = func(_ *domain.Faction, available []action.Action) (action.Action, error) {
			for _, a := range available {
				if a.Name() == "Sell Asset" {
					return a, nil
				}
			}
			if len(available) == 0 {
				return nil, nil
			}
			return available[0], nil
		}
		collector.SelectAssetFn = func(assets []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
			if len(assets) == 0 {
				return nil, nil
			}
			return assets[0], nil
		}

		if err := eng.Turn.Start(factionState); err != nil {
			t.Fatalf("Turn.Start: %v", err)
		}
		obs := &testharness.RecordingObserver{}
		if err := eng.RunCycle(factionState, paths, engine.Collectors{Phase: collector, Action: collector}, obs); err != nil {
			t.Fatalf("RunCycle: %v", err)
		}

		symlink := filepath.Join(logsDir, "latest.log")
		target, err := os.Readlink(symlink)
		if err != nil {
			t.Fatalf("readlink latest.log: %v", err)
		}
		contents, err := os.ReadFile(filepath.Join(logsDir, target))
		if err != nil {
			t.Fatalf("reading log file: %v", err)
		}
		t.Logf("\n=== debug=%v ===\n%s", debug, contents)
	}
}

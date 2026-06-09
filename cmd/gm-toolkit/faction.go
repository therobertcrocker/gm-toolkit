package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui"
	"github.com/therobertcrocker/gm-toolkit/internal/logging"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

func newFactionCmd() *cobra.Command {
	var dryRun bool
	var debug bool
	var campaignOverride string

	cmd := &cobra.Command{
		Use:   "faction",
		Short: "Open the faction TUI",
		Long:  "Launches the faction-manager TUI against the active campaign. Use --campaign <id> to override the active pointer for this invocation. Use --dryrun for a one-faction smoke cycle.",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

			if dryRun {
				return tui.RunDryRun(log)
			}

			return factionRun(debug, campaignOverride)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dryrun", false, "run a one-faction smoke cycle and exit")
	cmd.Flags().BoolVar(&debug, "debug", false, "write debug-level logs to the campaign log file")
	cmd.Flags().StringVar(&campaignOverride, "campaign", "", "campaign id to use instead of the active one")
	return cmd
}

func factionRun(debug bool, campaignOverride string) error {
	reg, err := campaigns.LoadRegistry()
	if err != nil {
		return fmt.Errorf("faction: %w", err)
	}
	camp, err := campaigns.ResolveActive(reg, campaignOverride)
	if err != nil {
		return formatActiveResolveError(err)
	}
	paths := camp.Paths()

	log, err := logging.New(paths.LogsDir, debug)
	if err != nil {
		return fmt.Errorf("faction: init logging in %s: %w", paths.LogsDir, err)
	}

	rb, err := rulebook.Load(paths.FactionDataDir)
	if err != nil {
		return fmt.Errorf("faction: load rulebook from %s: %w\n  hint: seed it with 'gm-toolkit campaign add-rules --rules <path>'", paths.FactionDataDir, err)
	}

	spatialMap, err := spatial.LoadRegionMap(paths.SpatialDataDir)
	if err != nil {
		return fmt.Errorf("faction: load spatial from %s: %w\n  hint: (spatial seed command TBD — see F-012)", paths.SpatialDataDir, err)
	}

	factionState, err := state.Load(paths.StatePath)
	if err != nil {
		return fmt.Errorf("faction: load state from %s: %w", paths.StatePath, err)
	}
	if factionState.CampaignID == "" {
		factionState.CampaignID = camp.ID()
	}

	eng := engine.NewWithRulebook(rb, world.NewWithMap(spatialMap, log), log)

	return tui.Run(eng, factionState, &paths, rb, spatialMap, log)
}

func formatActiveResolveError(err error) error {
	switch {
	case errors.Is(err, campaigns.ErrNoActiveCampaign):
		return fmt.Errorf("faction: no active campaign\n  hint: gm-toolkit campaign create <id> --path <dir>\n  or:   gm-toolkit campaign set-active --campaign <id>")
	case errors.Is(err, campaigns.ErrCampaignNotRegistered):
		return fmt.Errorf("faction: %w\n  hint: gm-toolkit campaign register <path>", err)
	case errors.Is(err, campaigns.ErrCampaignRootMissing):
		return fmt.Errorf("faction: %w\n  hint: re-register at its new location or restore the directory", err)
	default:
		return fmt.Errorf("faction: %w", err)
	}
}

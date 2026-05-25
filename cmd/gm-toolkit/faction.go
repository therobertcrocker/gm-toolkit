package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui"
)

func newFactionCmd() *cobra.Command {
	var dryRun bool
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

			return factionRun(log, campaignOverride)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dryrun", false, "run a one-faction smoke cycle and exit")
	cmd.Flags().StringVar(&campaignOverride, "campaign", "", "campaign id to use instead of the active one")
	return cmd
}

func factionRun(log *slog.Logger, campaignOverride string) error {
	reg, err := campaigns.LoadRegistry()
	if err != nil {
		return fmt.Errorf("faction: %w", err)
	}
	camp, err := campaigns.ResolveActive(reg, campaignOverride)
	if err != nil {
		return formatActiveResolveError(err)
	}
	paths := camp.Paths()

	rb, err := rulebook.Load(paths.FactionDataDir)
	if err != nil {
		return fmt.Errorf("faction: load rulebook from %s: %w\n  hint: seed it with 'gm-toolkit campaign add-rules --rules <path>'", paths.FactionDataDir, err)
	}

	factionState, err := state.Load(paths.StatePath)
	if err != nil {
		return fmt.Errorf("faction: load state from %s: %w", paths.StatePath, err)
	}
	if factionState.CampaignID == "" {
		factionState.CampaignID = camp.ID()
	}

	eng := engine.NewWithRulebook(rb, log)

	return tui.Run(eng, factionState, &paths, log)
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

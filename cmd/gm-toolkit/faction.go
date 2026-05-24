package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui"
)

var (
	factionDryRun bool
)

var factionCmd = &cobra.Command{
	Use:   "faction",
	Short: "Open the faction TUI",
	Long:  "Launches the faction-manager TUI. Use --dryrun to run a one-faction smoke cycle through the engine and exit (developer-only).",
	RunE:  factionRun,
}

func init() {
	factionCmd.Flags().BoolVar(&factionDryRun, "dryrun", false, "run a one-faction smoke cycle and exit")
	rootCmd.AddCommand(factionCmd)
}

func factionRun(cmd *cobra.Command, args []string) error {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	if factionDryRun {
		return tui.RunDryRun(log)
	}

	return tui.Run(nil, nil, nil, log)
}

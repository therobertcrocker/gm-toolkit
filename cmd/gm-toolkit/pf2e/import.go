package pf2e

import (
	"fmt"
	"io"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/foundry"
)

func newImportCmd(campaignOverride *string) *cobra.Command {
	return &cobra.Command{
		Use:   "import <fight>",
		Short: "Import a fight's Foundry exports into creature files",
		Long:  "Reads encounters/<fight>/source/actors/*.json and writes each creature's import.toml: PCs under pf2e/party/<pc>/, NPCs under encounters/<fight>/<npc>/. For each creature it prints the merged stat block and everything it did not import, with the reason.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fight := args[0]
			if !creature.IsFolderName(fight) {
				return fmt.Errorf("pf2e sim import: bad fight name %q", fight)
			}
			campaign, err := resolveCampaign(*campaignOverride)
			if err != nil {
				return fmt.Errorf("pf2e sim import: %w", err)
			}
			paths := campaign.Paths()
			reports, importErr := foundry.ImportFight(filepath.Join(paths.EncountersDir, fight), paths.PartyDir)
			out := cmd.OutOrStdout()
			styled := isTerminal(out)
			failed := 0
			for _, report := range reports {
				var shown *creature.Creature
				var loadErr error
				if report.Folder != "" {
					shown, loadErr = creature.Load(report.Folder)
					if loadErr != nil {
						failed++
					}
				}
				printReport(out, report, shown, loadErr, styled)
			}
			if importErr != nil {
				return fmt.Errorf("pf2e sim import: %w", importErr)
			}
			if failed > 0 {
				return fmt.Errorf("pf2e sim import: %d creatures failed to load", failed)
			}
			return nil
		},
	}
}

func printReport(out io.Writer, report *foundry.Report, shown *creature.Creature, loadErr error, styled bool) {
	bold, reset := "", ""
	if styled {
		bold, reset = ansiBold, ansiReset
	}
	switch {
	case report.Folder == "":
		fmt.Fprintf(out, "%s%s%s  not imported\n", bold, report.Source, reset)
	case loadErr != nil:
		fmt.Fprintf(out, "%s%s%s  wrote %s\n", bold, report.Source, reset, filepath.Join(report.Folder, creature.ImportLayer))
		fmt.Fprintf(out, "Load failed: %v\n", loadErr)
	default:
		fmt.Fprintf(out, "%s%s%s  wrote %s\n", bold, report.Source, reset, filepath.Join(report.Folder, creature.ImportLayer))
		printCreature(out, shown, styled)
	}

	byReason := map[string][]foundry.Skip{}
	for _, skip := range report.Skipped {
		byReason[skip.Reason] = append(byReason[skip.Reason], skip)
	}
	for _, reason := range slices.Sorted(maps.Keys(byReason)) {
		fmt.Fprintf(out, "Skipped, %s:\n", reason)
		for _, skip := range byReason[reason] {
			fmt.Fprintf(out, "  %s: %s\n", skip.Item, skip.What)
		}
	}
	if len(report.Unused) > 0 {
		counts := make([]string, 0, len(report.Unused))
		for _, kind := range slices.Sorted(maps.Keys(report.Unused)) {
			counts = append(counts, fmt.Sprintf("%s ×%d", kind, report.Unused[kind]))
		}
		fmt.Fprintf(out, "Not used by the sim: %s\n", strings.Join(counts, ", "))
	}
	fmt.Fprintln(out)
}

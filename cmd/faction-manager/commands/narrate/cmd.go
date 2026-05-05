package narrate

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/paths"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/narrative"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/narrative/digest"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func NewCmd(rulebook *rulebook.Rulebook) *cobra.Command {
	var campaignID string
	var outPath string
	var seed int64
	var toStdout bool

	cmd := &cobra.Command{
		Use:   "narrate <cycle>",
		Short: "Generate a narrative summary of a faction cycle",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cycleNumber, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid cycle number %q: %w", args[0], err)
			}

			if !cmd.Flags().Changed("seed") {
				seed = time.Now().UnixNano()
			}

			p := paths.New(campaignID)

			factionState, err := state.Load(p.State)
			if err != nil {
				return fmt.Errorf("loading state: %w", err)
			}

			records, err := narrative.LoadCycle(p.History, cycleNumber)
			if err != nil {
				return fmt.Errorf("loading history: %w", err)
			}

			cycleDigest, err := digest.Build(records, cycleNumber, factionState, rulebook)
			if err != nil {
				return fmt.Errorf("building digest: %w", err)
			}

			output, err := narrative.NewWireRenderer().Render(cycleDigest, seed)
			if err != nil {
				return fmt.Errorf("rendering narrative: %w", err)
			}

			if toStdout {
				fmt.Print(output)
				fmt.Fprintf(os.Stderr, "seed: %d\n", seed)
				return nil
			}

			var dest string
			if cmd.Flags().Changed("out") {
				dest = outPath
			} else {
				dest, err = resolveOutputPath(p.Narratives, cycleNumber)
				if err != nil {
					return fmt.Errorf("resolving output path: %w", err)
				}
			}

			if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
				return fmt.Errorf("creating output directory: %w", err)
			}
			if err := os.WriteFile(dest, []byte(output), 0644); err != nil {
				return fmt.Errorf("writing narrative: %w", err)
			}

			fmt.Printf("wrote %s (seed: %d)\n", dest, seed)
			return nil
		},
	}

	cmd.Flags().StringVar(&campaignID, "campaign", "", "campaign ID (required)")
	cmd.Flags().StringVar(&outPath, "out", "", "output path (bypasses increment, overwrites)")
	cmd.Flags().Int64Var(&seed, "seed", 0, "RNG seed for variant selection")
	cmd.Flags().BoolVar(&toStdout, "stdout", false, "print to stdout instead of writing a file")
	cmd.MarkFlagRequired("campaign")

	return cmd
}

func resolveOutputPath(narrativesDir string, cycleNumber int) (string, error) {
	base := narrativesDir
	primary := filepath.Join(base, fmt.Sprintf("cycle-%03d.md", cycleNumber))
	if _, err := os.Stat(primary); os.IsNotExist(err) {
		return primary, nil
	}
	for i := 2; i <= 999; i++ {
		candidate := filepath.Join(base, fmt.Sprintf("cycle-%03d-%03d.md", cycleNumber, i))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("all output slots for cycle %03d are taken", cycleNumber)
}

package spatial

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial/builder"
)

func newBuildCmd() *cobra.Command {
	var (
		campaignID string
		replace    bool
	)
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build canonical regions.toml + worlds.toml from spatial/source/ inputs",
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := campaigns.LoadRegistry()
			if err != nil {
				return fmt.Errorf("spatial build: %w", err)
			}
			camp, err := campaigns.ResolveActive(reg, campaignID)
			if err != nil {
				if errors.Is(err, campaigns.ErrNoActiveCampaign) {
					return fmt.Errorf("spatial build: no active campaign and no --campaign specified\n  hint: gm-toolkit campaign set-active --campaign <id>")
				}
				return fmt.Errorf("spatial build: %w", err)
			}
			opts := builder.Options{
				SourceDir: filepath.Join(camp.Root, "spatial", "source"),
				DestDir:   filepath.Join(camp.Root, "spatial"),
				Replace:   replace,
			}
			summary, err := builder.Build(opts)
			if err != nil {
				switch {
				case errors.Is(err, builder.ErrMissingSource):
					return fmt.Errorf("spatial build: %w\n  hint: create %s/layout.txt and data.toml", err, opts.SourceDir)
				case errors.Is(err, builder.ErrOutputExists):
					return fmt.Errorf("spatial build: %w\n  hint: pass --replace to overwrite", err)
				default:
					return fmt.Errorf("spatial build: %w", err)
				}
			}
			printSummary(cmd, camp.ID(), summary, opts.DestDir)
			return nil
		},
	}
	cmd.Flags().StringVar(&campaignID, "campaign", "", "target campaign id (defaults to active)")
	cmd.Flags().BoolVar(&replace, "replace", false, "overwrite existing regions.toml / worlds.toml")
	return cmd
}

func printSummary(cmd *cobra.Command, campaignID string, s builder.Summary, dstDir string) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Built spatial map for campaign %s:\n", campaignID)
	fmt.Fprintf(out, "  %d regions (%s)\n", len(s.Regions), strings.Join(s.Regions, ", "))
	fmt.Fprintf(out, "  %d worlds\n", s.WorldCount)
	fmt.Fprintf(out, "  %d adjacency boundaries\n", s.AdjacencyCount)
	fmt.Fprintf(out, "  %d warps\n", s.WarpCount)
	fmt.Fprintf(out, "Wrote %s/regions.toml\n", dstDir)
	fmt.Fprintf(out, "Wrote %s/worlds.toml\n", dstDir)
}

package campaign

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
)

func newAddRulesCmd() *cobra.Command {
	var (
		source   string
		campaign string
		replace  bool
	)

	cmd := &cobra.Command{
		Use:   "add-rules",
		Short: "Copy rulebook TOML into a campaign's rulebook/",
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := campaigns.LoadRegistry()
			if err != nil {
				return fmt.Errorf("add-rules: %w", err)
			}
			camp, err := campaigns.ResolveActive(reg, campaign)
			if err != nil {
				if errors.Is(err, campaigns.ErrNoActiveCampaign) {
					return fmt.Errorf("add-rules: no active campaign and no --campaign specified\n  hint: gm-toolkit campaign set-active --campaign <id>")
				}
				return fmt.Errorf("add-rules: %w", err)
			}
			copied, err := campaigns.CopyRulebook(camp, source, replace)
			if err != nil {
				if errors.Is(err, campaigns.ErrRulebookConflict) {
					return fmt.Errorf("add-rules: %w\n  hint: pass --replace to overwrite", err)
				}
				return fmt.Errorf("add-rules: %w", err)
			}
			fmt.Printf("Copied %d file(s) into %s/rulebook/\n", copied, camp.Root)
			return nil
		},
	}

	cmd.Flags().StringVar(&source, "rules", "", "source directory of TOML files (required)")
	cmd.Flags().StringVar(&campaign, "campaign", "", "target campaign id (defaults to active)")
	cmd.Flags().BoolVar(&replace, "replace", false, "overwrite conflicting files in destination")
	_ = cmd.MarkFlagRequired("rules")
	return cmd
}

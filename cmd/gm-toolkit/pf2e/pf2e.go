package pf2e

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
)

func NewCmd() *cobra.Command {
	var campaignOverride string
	cmd := &cobra.Command{
		Use:   "pf2e",
		Short: "Pathfinder 2e tools",
		Long:  "Pathfinder 2e tools for the active campaign. Use --campaign <id> to override the active pointer for this invocation.",
	}
	cmd.PersistentFlags().StringVar(&campaignOverride, "campaign", "", "campaign id to use instead of the active one")
	cmd.AddCommand(newSimCmd(&campaignOverride))
	return cmd
}

func resolveCampaign(override string) (*campaigns.Campaign, error) {
	registry, err := campaigns.LoadRegistry()
	if err != nil {
		return nil, err
	}
	campaign, err := campaigns.ResolveActive(registry, override)
	switch {
	case err == nil:
		return campaign, nil
	case errors.Is(err, campaigns.ErrNoActiveCampaign):
		return nil, fmt.Errorf("no active campaign and no --campaign given\n  hint: gm-toolkit campaign set-active --campaign <id>")
	case errors.Is(err, campaigns.ErrCampaignNotRegistered):
		return nil, fmt.Errorf("%w\n  hint: gm-toolkit campaign register <path>", err)
	case errors.Is(err, campaigns.ErrCampaignRootMissing):
		return nil, fmt.Errorf("%w\n  hint: re-register at its new location or restore the directory", err)
	default:
		return nil, err
	}
}

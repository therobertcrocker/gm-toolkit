package pf2e

import "github.com/spf13/cobra"

func newSimCmd(campaignOverride *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sim",
		Short: "Simulate PF2e fights",
		Long:  "Simulate Pathfinder 2e fights from the campaign's creature files.",
	}
	cmd.AddCommand(newShowCmd(campaignOverride))
	return cmd
}

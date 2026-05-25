package campaign

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "campaign",
		Short: "Manage campaigns",
		Long:  "Create, register, switch between, and seed gm-toolkit campaigns.",
	}
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newRegisterCmd())
	cmd.AddCommand(newSetActiveCmd())
	cmd.AddCommand(newAddRulesCmd())
	return cmd
}

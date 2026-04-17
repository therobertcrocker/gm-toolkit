package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var factionCmd = &cobra.Command{
	Use:   "faction",
	Short: "Manage factions",
}

var factionCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new faction",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Create faction - not yet implemented")
		return nil
	},
}

var factionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all factions",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("List factions - not yet implemented")
		return nil
	},
}

func init() {
	factionCmd.AddCommand(factionCreateCmd)
	factionCmd.AddCommand(factionListCmd)
}

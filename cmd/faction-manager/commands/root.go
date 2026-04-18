package commands

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "faction-manager",
	Short: "A faction tracker inspired by the mechanics of Stars Without Number",
}

func init() {
	RootCmd.AddCommand(reviewCmd)
	RootCmd.AddCommand(turnCmd)
	RootCmd.AddCommand(factionCmd)
}

package commands

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "faction-manager",
	Short: "A faction tracker for Stars Without Number campaigns",
}

func init() {
	RootCmd.AddCommand(reviewCmd)
	RootCmd.AddCommand(turnCmd)
	RootCmd.AddCommand(factionCmd)
}

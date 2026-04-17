package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var turnCmd = &cobra.Command{
	Use:   "turn",
	Short: "Execute a faction turn",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Turn mode - not yet implemented")
		return nil
	},
}

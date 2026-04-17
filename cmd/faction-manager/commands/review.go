package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Review current faction state and turn history",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Review mode - not yet implemented")
		return nil
	},
}

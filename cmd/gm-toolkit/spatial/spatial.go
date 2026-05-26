package spatial

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spatial",
		Short: "Author and build spatial maps",
		Long:  "Build the canonical regions.toml / worlds.toml from GM-authored layout.txt + data.toml sources.",
	}
	cmd.AddCommand(newBuildCmd())
	return cmd
}

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gm-toolkit",
	Short: "GM Toolkit — tools for running Stars Without Number factions and worlds",
	Long:  "GM Toolkit bundles utilities for the GM running Stars Without Number campaigns. Run `gm-toolkit <subcommand> --help` for per-subcommand details.",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

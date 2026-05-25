package main

import (
	"github.com/spf13/cobra"

	"github.com/therobertcrocker/gm-toolkit/cmd/gm-toolkit/campaign"
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "gm-toolkit",
		Short: "GM Toolkit — tools for running Stars Without Number factions and worlds",
		Long:  "GM Toolkit bundles utilities for the GM running Stars Without Number campaigns. Run `gm-toolkit <subcommand> --help` for per-subcommand details.",
	}
	root.AddCommand(campaign.NewCmd())
	root.AddCommand(newFactionCmd())
	return root
}

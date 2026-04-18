package main

import (
	"fmt"
	"os"

	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/commands"
)

func main() {
	if err := commands.RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

package main

import (
	"fmt"
	"os"

	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/commands"
)

func main() {
	app := commands.NewApp()
	if err := app.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

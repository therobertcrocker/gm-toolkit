package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/commands"
)

func main() {
	godotenv.Load() // no-op if .env not present

	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	app := commands.NewApp(cfg.FactionDataDir)
	if err := app.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

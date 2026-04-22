package main

import (
	"fmt"
	"os"
)

// Config holds all environment-driven configuration for the faction manager.
type Config struct {
	FactionDataDir string
}

// LoadConfig reads required environment variables and returns a populated Config.
// Call godotenv.Load() before LoadConfig() so .env values are present.
func LoadConfig() (*Config, error) {
	dataDir := os.Getenv("FACTION_DATA_DIR")
	if dataDir == "" {
		return nil, fmt.Errorf("FACTION_DATA_DIR is not set (check your .env file)")
	}
	return &Config{FactionDataDir: dataDir}, nil
}

package state

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

type FactionState struct {
	CampaignID  string                     `toml:"campaign_id"`
	CycleNumber int                        `toml:"cycle_number"`
	Factions    map[string]*domain.Faction `toml:"factions"`
	CurrentTurn *domain.TurnState          `toml:"current_turn,omitempty"`
}

// Load reads campaign state from path. Returns an empty State if the file does not exist.
func Load(path string) (*FactionState, error) {
	var s FactionState
	if _, err := toml.DecodeFile(path, &s); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &FactionState{Factions: make(map[string]*domain.Faction)}, nil
		}
		return nil, err
	}
	if s.Factions == nil {
		s.Factions = make(map[string]*domain.Faction)
	}
	for _, faction := range s.Factions {
		if faction.Assets == nil {
			faction.Assets = make(map[string]*domain.Asset)
		}
	}
	return &s, nil
}

// Save writes campaign state to path, creating parent directories as needed.
// On the first save it also writes a seed file alongside state.toml so that
// scripts/reset-campaign.sh can restore the campaign to its initial state.
func Save(path string, s *FactionState) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(s); err != nil {
		return err
	}
	return ensureSeed(path, s)
}

// ensureSeed writes state.seed.toml next to state.toml if one does not already
// exist. It is a no-op on every subsequent save, preserving the original state.
func ensureSeed(statePath string, s *FactionState) error {
	seedPath := filepath.Join(filepath.Dir(statePath), "state.seed.toml")
	if _, err := os.Stat(seedPath); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	f, err := os.Create(seedPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(s)
}

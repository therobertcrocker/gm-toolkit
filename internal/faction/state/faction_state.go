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
func Save(path string, s *FactionState) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(s)
}

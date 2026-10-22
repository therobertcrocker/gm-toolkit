package state

import (
	"errors"
	"os"
	"path/filepath"
	"slices"

	"github.com/BurntSushi/toml"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

type FactionState struct {
	CampaignID  string                     `toml:"campaign_id"`
	CycleNumber int                        `toml:"cycle_number"`
	Factions    map[string]*domain.Faction `toml:"factions"`
	// Governments maps a world ID to the faction that rules it. A world with no
	// entry is independent: no tracked faction governs it (A.001.6 overview Decision 8).
	Governments map[string]string `toml:"governments,omitempty"`
	CurrentTurn *domain.TurnState `toml:"current_turn,omitempty"`
}

// Ruler returns the ID of the faction that governs worldID, or "" for an
// independent world. A nil state has no governments.
func (factionState *FactionState) Ruler(worldID string) string {
	if factionState == nil {
		return ""
	}
	return factionState.Governments[worldID]
}

// IsIndependent reports whether no tracked faction governs worldID.
func (factionState *FactionState) IsIndependent(worldID string) bool {
	return factionState.Ruler(worldID) == ""
}

// WorldsRuledBy returns the worlds factionID governs, in ID order.
func (factionState *FactionState) WorldsRuledBy(factionID string) []string {
	if factionState == nil {
		return nil
	}
	var worlds []string
	for worldID, ruler := range factionState.Governments {
		if ruler == factionID {
			worlds = append(worlds, worldID)
		}
	}
	slices.Sort(worlds)
	return worlds
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

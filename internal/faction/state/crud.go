package state

import (
	"errors"
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

var (
	ErrFactionAlreadyExists = errors.New("state: faction already exists")
	ErrFactionNotFound      = errors.New("state: faction not found")
	ErrInvalidFactionID     = errors.New("state: invalid faction id (empty)")
)

// CreateFaction validates uniqueness, applies the mutation to fs.Factions,
// records governedWorldID as a world the faction rules (none when it is empty),
// then persists to path. The mutation is atomic: if Save fails the in-memory
// insert is rolled back, so fs.Factions always matches what is on disk and the
// caller may safely retry.
func CreateFaction(path string, fs *FactionState, faction *domain.Faction, governedWorldID string) error {
	if faction.ID == "" {
		return ErrInvalidFactionID
	}
	if _, exists := fs.Factions[faction.ID]; exists {
		return fmt.Errorf("state: create %q: %w", faction.ID, ErrFactionAlreadyExists)
	}
	previousRuler, governed := fs.Governments[governedWorldID]
	fs.Factions[faction.ID] = faction
	if governedWorldID != "" {
		if fs.Governments == nil {
			fs.Governments = make(map[string]string)
		}
		fs.Governments[governedWorldID] = faction.ID
	}
	if err := Save(path, fs); err != nil {
		delete(fs.Factions, faction.ID)
		if governedWorldID != "" {
			restoreGovernment(fs, governedWorldID, previousRuler, governed)
		}
		return fmt.Errorf("state: create %q: save: %w", faction.ID, err)
	}
	return nil
}

// restoreGovernment puts worldID's entry back as it was before a failed save.
func restoreGovernment(fs *FactionState, worldID, previousRuler string, governed bool) {
	if governed {
		fs.Governments[worldID] = previousRuler
		return
	}
	delete(fs.Governments, worldID)
}

// DeleteFaction validates that id exists in fs.Factions, removes the entry and
// every world it governs from the record, then persists. Like CreateFaction the mutation is atomic: if Save fails the
// removed entry is restored. Returns ErrFactionNotFound if absent.
func DeleteFaction(path string, fs *FactionState, id string) error {
	if id == "" {
		return ErrInvalidFactionID
	}
	removed, exists := fs.Factions[id]
	if !exists {
		return fmt.Errorf("state: delete %q: %w", id, ErrFactionNotFound)
	}
	ruled := worldsRuledBy(fs, id)
	delete(fs.Factions, id)
	for _, worldID := range ruled {
		delete(fs.Governments, worldID)
	}
	if err := Save(path, fs); err != nil {
		fs.Factions[id] = removed
		for _, worldID := range ruled {
			fs.Governments[worldID] = id
		}
		return fmt.Errorf("state: delete %q: save: %w", id, err)
	}
	return nil
}

func worldsRuledBy(fs *FactionState, factionID string) []string {
	var worlds []string
	for worldID, ruler := range fs.Governments {
		if ruler == factionID {
			worlds = append(worlds, worldID)
		}
	}
	return worlds
}

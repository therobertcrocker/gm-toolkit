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
// then persists to path. The mutation is atomic: if Save fails the in-memory
// insert is rolled back, so fs.Factions always matches what is on disk and the
// caller may safely retry.
func CreateFaction(path string, fs *FactionState, faction *domain.Faction) error {
	if faction.ID == "" {
		return ErrInvalidFactionID
	}
	if _, exists := fs.Factions[faction.ID]; exists {
		return fmt.Errorf("state: create %q: %w", faction.ID, ErrFactionAlreadyExists)
	}
	fs.Factions[faction.ID] = faction
	if err := Save(path, fs); err != nil {
		delete(fs.Factions, faction.ID)
		return fmt.Errorf("state: create %q: save: %w", faction.ID, err)
	}
	return nil
}

// DeleteFaction validates that id exists in fs.Factions, removes the entry,
// then persists. Like CreateFaction the mutation is atomic: if Save fails the
// removed entry is restored. Returns ErrFactionNotFound if absent.
func DeleteFaction(path string, fs *FactionState, id string) error {
	if id == "" {
		return ErrInvalidFactionID
	}
	removed, exists := fs.Factions[id]
	if !exists {
		return fmt.Errorf("state: delete %q: %w", id, ErrFactionNotFound)
	}
	delete(fs.Factions, id)
	if err := Save(path, fs); err != nil {
		fs.Factions[id] = removed
		return fmt.Errorf("state: delete %q: save: %w", id, err)
	}
	return nil
}

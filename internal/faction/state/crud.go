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
// then persists to path. On Save failure, the in-memory mutation is NOT rolled
// back — caller surfaces the error and retries.
func CreateFaction(path string, fs *FactionState, faction *domain.Faction) error {
	if faction.ID == "" {
		return ErrInvalidFactionID
	}
	if _, exists := fs.Factions[faction.ID]; exists {
		return fmt.Errorf("state: create %q: %w", faction.ID, ErrFactionAlreadyExists)
	}
	fs.Factions[faction.ID] = faction
	if err := Save(path, fs); err != nil {
		return fmt.Errorf("state: create %q: save: %w", faction.ID, err)
	}
	return nil
}

// DeleteFaction validates that id exists in fs.Factions, removes the entry,
// then persists. Returns ErrFactionNotFound if absent.
func DeleteFaction(path string, fs *FactionState, id string) error {
	if id == "" {
		return ErrInvalidFactionID
	}
	if _, exists := fs.Factions[id]; !exists {
		return fmt.Errorf("state: delete %q: %w", id, ErrFactionNotFound)
	}
	delete(fs.Factions, id)
	if err := Save(path, fs); err != nil {
		return fmt.Errorf("state: delete %q: save: %w", id, err)
	}
	return nil
}

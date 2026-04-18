package commands

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func runDeleteFactionWizard(campaignID string) error {
	statePath := filepath.Join(".", "campaigns", campaignID, "faction_state.toml")

	s, err := state.Load(statePath)
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	if len(s.Factions) == 0 {
		fmt.Println("No factions to delete.")
		return nil
	}

	options := make([]huh.Option[string], len(s.Factions))
	for i, f := range s.Factions {
		options[i] = huh.NewOption(f.Name, f.ID)
	}

	var targetID string

	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Delete Faction").
				Description("Select a faction to delete").
				Options(options...).
				Value(&targetID),
		),
	).Run(); err != nil {
		return fmt.Errorf("wizard cancelled: %w", err)
	}

	var target *domain.Faction
	for _, f := range s.Factions {
		if f.ID == targetID {
			target = f
			break
		}
	}

	summary := fmt.Sprintf(
		"Name: %s\nHomeworld: %s\nForce: %d  Cunning: %d  Wealth: %d\nHP: %d / %d",
		target.Name, target.Homeworld,
		target.Force, target.Cunning, target.Wealth,
		target.CurrentHP, target.MaxHP,
	)

	var confirmed bool

	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Delete %q?", target.Name)).
				Description(summary+"\n\nThis cannot be undone.").
				Value(&confirmed),
		),
	).Run(); err != nil {
		return fmt.Errorf("wizard cancelled: %w", err)
	}

	if !confirmed {
		fmt.Println("Delete cancelled.")
		return nil
	}

	filtered := s.Factions[:0]
	for _, f := range s.Factions {
		if f.ID != targetID {
			filtered = append(filtered, f)
		}
	}
	s.Factions = filtered

	if err := state.Save(statePath, s); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	fmt.Printf("Faction %q deleted.\n", target.Name)
	return nil
}
